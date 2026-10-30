package file

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/datazip-inc/olake-ui/server/internal/gitops/reconciler"
	"github.com/datazip-inc/olake-ui/server/internal/services"
	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

const (
	fileDebounce = 5 * time.Second
	// stateFileName is the default state file, kept in OLAKE_GITOPS_FILE_DIR. It is hidden, so the
	// watcher does not read it as a manifest.
	stateFileName = ".olake-gitops-state.json"
)

// resourceState is the file-mode counterpart of the olake.io/* status annotations.
type resourceState struct {
	Hash     string `json:"observedHash"`
	Phase    string `json:"phase"`
	Message  string `json:"message,omitempty"`
	EntityID string `json:"entityId,omitempty"`
}

// Watcher runs GitOps in file mode: it watches a directory of JSON manifests (manifest.go), hands
// them to the reconcilers, and is their StatusSink, keeping status in a state file.
type Watcher struct {
	Dir            string
	stateFile      string
	temporalClient *temporal.Temporal

	// the reconcilers, which report back through this Watcher's SetPhase and Indicator
	reconcileSource      func(ctx context.Context, res *reconciler.ResourceData) (ctrl.Result, error)
	reconcileDestination func(ctx context.Context, res *reconciler.ResourceData) (ctrl.Result, error)
	reconcileJob         func(ctx context.Context, job, streams *reconciler.ResourceData) (ctrl.Result, error)

	mu        sync.Mutex
	resources map[string]reconciler.ResourceData // resource key -> last parsed manifest
	paths     map[string]string                  // resource key -> file it was read from
	state     map[string]resourceState           // resource key -> status, mirrored to stateFile
	debounce  map[string]*time.Timer
	retries   map[string]*time.Timer
	reconcile chan string
	stop      chan struct{}
	watcher   *fsnotify.Watcher
}

// Setup reads the state file and starts watching dir, so a bad directory or an unwritable state file
// fails startup. stateFile defaults to dir/.olake-gitops-state.json.
func Setup(app *services.AppService, dir, stateFile string) (*Watcher, error) {
	if dir == "" {
		return nil, fmt.Errorf("OLAKE_GITOPS_FILE_DIR is empty")
	}
	etlSvc := app.ETL()
	fw := newWatcher(dir, stateFile, etlSvc.Temporal())
	fw.reconcileSource = (&reconciler.Source{ETL: etlSvc, Sink: fw}).Reconcile
	fw.reconcileDestination = (&reconciler.Destination{ETL: etlSvc, Sink: fw}).Reconcile
	fw.reconcileJob = (&reconciler.Job{ETL: etlSvc, Sink: fw}).Reconcile
	if err := fw.setup(); err != nil {
		return nil, err
	}
	logger.Infof("GitOps file watcher watching %s", dir)
	return fw, nil
}

func newWatcher(dir, stateFile string, t *temporal.Temporal) *Watcher {
	if stateFile == "" {
		stateFile = filepath.Join(dir, stateFileName)
	}
	fw := &Watcher{
		Dir:            dir,
		stateFile:      stateFile,
		temporalClient: t,
		resources:      map[string]reconciler.ResourceData{},
		paths:          map[string]string{},
		state:          map[string]resourceState{},
		debounce:       map[string]*time.Timer{},
		retries:        map[string]*time.Timer{},
		reconcile:      make(chan string, 64),
		stop:           make(chan struct{}),
	}
	return fw
}

// resourceKey identifies a resource in the index and in the state file: "<kind>/<namespace>/<name>",
// with an empty namespace in file mode. Changing it orphans the status in existing state files.
func resourceKey(r *reconciler.ResourceData) string {
	return r.Kind + "/" + r.Namespace + "/" + r.Name
}

func (fw *Watcher) setup() error {
	if err := fw.readStateFile(); err != nil {
		return err
	}
	// write once now so a read-only or missing location fails at startup, not after the first sync
	if err := fw.writeStateFile(); err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("fsnotify: %w", err)
	}

	if err = watcher.Add(fw.Dir); err != nil {
		watcher.Close()
		return fmt.Errorf("watch %s: %w", fw.Dir, err)
	}
	fw.watcher = watcher
	return nil
}

// Run reconciles every manifest, then watches the directory until ctx is cancelled.
func (fw *Watcher) Run(ctx context.Context) error {
	defer fw.watcher.Close()
	defer close(fw.stop)
	if err := fw.walk(fw.Dir, 0); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return nil
			}
			fw.handleEvent(event)
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return nil
			}
			logger.Errorf("gitops file watcher: %s", err)
		case path := <-fw.reconcile:
			fw.process(ctx, path)
		}
	}
}

// walk watches dir and the directories under it, and enqueues the manifests in them. Manifests sit at
// the root or in a directory per job
func (fw *Watcher) walk(dir string, delay time.Duration) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return fw.watcher.Add(path)
		}
		if isManifest(path) {
			fw.enqueue(path, delay)
		}
		return nil
	})
}

// handleEvent enqueues a changed manifest. A remove/rename is processed too: the file is gone, so
// process drops it from the index (the OLake entity is kept, as in-cluster). Editors that save by
// rename are a remove followed by a create, which the debounce folds into one reconcile. A new
// directory is watched and its manifests read.
func (fw *Watcher) handleEvent(event fsnotify.Event) {
	if event.Op&fsnotify.Create != 0 {
		if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
			if err := fw.walk(event.Name, fileDebounce); err != nil {
				logger.Errorf("gitops file watcher: %s", err)
			}
			return
		}
	}
	if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 && isManifest(event.Name) {
		fw.enqueue(event.Name, fileDebounce)
	}
}

func (fw *Watcher) enqueue(path string, delay time.Duration) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	if t, ok := fw.debounce[path]; ok {
		t.Stop()
	}
	fw.debounce[path] = time.AfterFunc(delay, func() {
		// block rather than drop: discover can take minutes; a dropped path won't retry until file changes
		select {
		case fw.reconcile <- path:
		case <-fw.stop:
		}
	})
}

func (fw *Watcher) process(ctx context.Context, path string) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			fw.dropFile(path)
			return
		}
		logger.Errorf("read %s: %s", path, err)
		return
	}
	res, err := parseFileManifest(raw)
	if err != nil {
		logger.Errorf("parse %s: %s", path, err)
		return
	}

	fw.mu.Lock()
	// file now holds a different resource (name or kind changed): drop the old one to avoid ghosts
	for key, p := range fw.paths {
		if p == path && key != resourceKey(res) {
			fw.dropLocked(key)
		}
	}
	// restore persisted status so reconcilers see the same annotations as before restart
	res.Annotations = fw.annotationsFor(res)
	fw.resources[resourceKey(res)] = *res
	fw.paths[resourceKey(res)] = path
	fw.mu.Unlock()

	result, syncErr := fw.sync(ctx, res)
	if syncErr != nil {
		logger.Errorf("reconcile %s: %s", path, syncErr)
	}
	if result.RequeueAfter > 0 {
		fw.scheduleRetry(path, result.RequeueAfter)
	}

	fw.enqueueDependentJobs(res)
}

// dropFile drops the resource read from path, as a deleted ConfigMap drops out of the in-cluster
// lists: jobs stop seeing it. The OLake entity is not deleted.
func (fw *Watcher) dropFile(path string) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	for key, p := range fw.paths {
		if p == path {
			fw.dropLocked(key)
		}
	}
}

// dropLocked removes a resource from the index and its status, as a deleted ConfigMap loses its
// annotations: if the file comes back it is reconciled from scratch. Caller must hold fw.mu.
func (fw *Watcher) dropLocked(key string) {
	delete(fw.resources, key)
	delete(fw.paths, key)
	if _, ok := fw.state[key]; !ok {
		// no state to persist: skip disk write
		return
	}
	delete(fw.state, key)
	if err := fw.writeStateFile(); err != nil {
		logger.Errorf("gitops: %s", err)
	}
}

func (fw *Watcher) scheduleRetry(path string, after time.Duration) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	if t, ok := fw.retries[path]; ok {
		t.Stop()
	}
	fw.retries[path] = time.AfterFunc(after, func() {
		fw.enqueue(path, 0)
	})
}

// enqueueDependentJobs re-runs the jobs a Source, Destination or Streams change can unblock or
// retry, as the in-cluster job controller's watches do. Going through the queue keeps their
// Pending/transient requeues.
func (fw *Watcher) enqueueDependentJobs(changed *reconciler.ResourceData) {
	fw.mu.Lock()
	dep := *changed
	dep.Annotations = fw.annotationsFor(&dep) // entity ID may be set by this reconcile

	var paths []string
	for key, r := range fw.resources {
		if r.Kind != reconciler.KindJob {
			continue
		}
		job := r
		job.Annotations = fw.annotationsFor(&job)
		if reconciler.JobDependsOn(&job, &dep) {
			paths = append(paths, fw.paths[key])
		}
	}
	fw.mu.Unlock()

	for _, p := range paths {
		fw.enqueue(p, 0)
	}
}

func (fw *Watcher) sync(ctx context.Context, res *reconciler.ResourceData) (ctrl.Result, error) {
	switch res.Kind {
	case reconciler.KindSource:
		return fw.reconcileSource(ctx, res)
	case reconciler.KindDestination:
		return fw.reconcileDestination(ctx, res)
	case reconciler.KindJob:
		return fw.reconcileJob(ctx, res, fw.findStreams(res))
	case reconciler.KindStreams:
		// no standalone reconciler; the job owns streams status (enqueueDependentJobs)
		return ctrl.Result{}, nil
	default:
		return ctrl.Result{}, fmt.Errorf("unsupported olake.io/kind %q", res.Kind)
	}
}

// findStreams returns the streams manifest of job, with its status, or nil when there is none yet.
func (fw *Watcher) findStreams(job *reconciler.ResourceData) *reconciler.ResourceData {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	for _, r := range fw.resources {
		if reconciler.StreamsBelongTo(&r, job) {
			streams := r
			streams.Annotations = fw.annotationsFor(&streams)
			return &streams
		}
	}
	return nil
}

// annotationsFor reads fw.state. Caller must hold fw.mu.
func (fw *Watcher) annotationsFor(r *reconciler.ResourceData) map[string]string {
	st, ok := fw.state[resourceKey(r)]
	if !ok {
		return map[string]string{}
	}
	return map[string]string{
		reconciler.AnnotationObservedHash: st.Hash,
		reconciler.AnnotationPhase:        st.Phase,
		reconciler.AnnotationMessage:      st.Message,
		reconciler.AnnotationEntityID:     st.EntityID,
	}
}

func (fw *Watcher) Indicator(ctx context.Context, r *reconciler.ResourceData, action, errMsg string) {
	reconciler.StartIndicator(ctx, fw.temporalClient, r, action, errMsg)
}

// SetPhase records status in memory and in stateFile, and logs each change, standing in for the
// annotation patch and Event of kubernetes.Sink. Unchanged status writes nothing, so a resource requeued
// every syncRequeueAfter with the same message does not log again.
func (fw *Watcher) SetPhase(_ context.Context, r *reconciler.ResourceData, phase, message, entityID, observedHash string) error {
	key := resourceKey(r)
	next := resourceState{
		Hash:     observedHash,
		Phase:    phase,
		Message:  message,
		EntityID: entityID,
	}

	fw.mu.Lock()
	prev, existed := fw.state[key]
	if existed && prev == next {
		// no change: skip disk write and log (requeues hit this every 30s)
		fw.mu.Unlock()
		return nil
	}
	fw.state[key] = next
	if err := fw.writeStateFile(); err != nil {
		// roll back memory so it stays consistent with the file; next SetPhase call will retry
		if existed {
			fw.state[key] = prev
		} else {
			delete(fw.state, key)
		}
		fw.mu.Unlock()
		return err
	}
	fw.mu.Unlock()

	switch phase {
	case reconciler.PhaseReady:
		logger.Infof("gitops %s %q: synced to OLake (entity %s)", r.Kind, r.Name, entityID)
	case reconciler.PhaseFailed:
		logger.Errorf("gitops %s %q: failed: %s", r.Kind, r.Name, message)
	default:
		logger.Infof("gitops %s %q: %s: %s", r.Kind, r.Name, phase, message)
	}
	return nil
}

// readStateFile reads stateFile. A missing file is an empty state (first start). An unreadable one
// is an error rather than an empty state: starting empty replays every job.
func (fw *Watcher) readStateFile() error {
	raw, err := os.ReadFile(fw.stateFile)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read gitops state %s: %w", fw.stateFile, err)
	}
	state := map[string]resourceState{}
	if err := json.Unmarshal(raw, &state); err != nil {
		return fmt.Errorf("parse gitops state %s (delete it to reconcile every resource from scratch): %w", fw.stateFile, err)
	}
	fw.state = state
	return nil
}

// writeStateFile writes fw.state to stateFile. It writes a temp file and renames it over
// stateFile, so a crash mid-write cannot leave a truncated file that readStateFile rejects.
// Caller must hold fw.mu.
func (fw *Watcher) writeStateFile() error {
	raw, err := json.Marshal(fw.state)
	if err != nil {
		return fmt.Errorf("encode gitops state: %w", err)
	}
	tmp := fw.stateFile + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return fmt.Errorf("write gitops state %s: %w", fw.stateFile, err)
	}
	if err := os.Rename(tmp, fw.stateFile); err != nil {
		return fmt.Errorf("write gitops state %s: %w", fw.stateFile, err)
	}
	return nil
}
