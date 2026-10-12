package gitops

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/yaml"

	"github.com/datazip-inc/olake-ui/server/internal/services/etl"
	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

const fileDebounce = 5 * time.Second

type resourceState struct {
	hash     string
	phase    string
	message  string
	entityID string
}

// FileWatcher watches a directory of ConfigMap YAML files and runs the same
// reconcile path as the Kubernetes controllers.
type FileWatcher struct {
	Dir            string
	source         *SourceReconciler
	dest           *DestinationReconciler
	job            *JobReconciler
	temporalClient *temporal.Temporal

	mu        sync.Mutex
	resources map[string]ResourceData // resource key -> last parsed manifest
	paths     map[string]string       // resource key -> file it was read from
	state     map[string]resourceState
	debounce  map[string]*time.Timer
	retries   map[string]*time.Timer
	reconcile chan string
	stop      chan struct{}
	watcher   *fsnotify.Watcher
}

func newFileWatcher(dir string, etlSvc *etl.Service, t *temporal.Temporal) *FileWatcher {
	fw := &FileWatcher{
		Dir:            dir,
		temporalClient: t,
		resources:      map[string]ResourceData{},
		paths:          map[string]string{},
		state:          map[string]resourceState{},
		debounce:       map[string]*time.Timer{},
		retries:        map[string]*time.Timer{},
		reconcile:      make(chan string, 64),
		stop:           make(chan struct{}),
	}
	sink := StatusSink(fw)
	fw.source = &SourceReconciler{ETL: etlSvc, Sink: sink}
	fw.dest = &DestinationReconciler{ETL: etlSvc, Sink: sink}
	fw.job = &JobReconciler{ETL: etlSvc, Sink: sink, findStreams: fw.findStreams}
	return fw
}

func (fw *FileWatcher) setup() error {
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

func (fw *FileWatcher) run(ctx context.Context) error {
	defer fw.watcher.Close()
	defer close(fw.stop)
	entries, err := os.ReadDir(fw.Dir)
	if err != nil {
		return fmt.Errorf("read %s: %w", fw.Dir, err)
	}
	for _, e := range entries {
		if e.IsDir() || !isYAML(e.Name()) {
			continue
		}
		fw.enqueue(filepath.Join(fw.Dir, e.Name()), 0)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return nil
			}
			// remove/rename is processed too: the file is gone, so process drops it from the index
			// (the OLake entity is kept, as in-cluster). Editors that save by rename are a remove
			// followed by a create, which the debounce folds into one reconcile.
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 && isYAML(event.Name) {
				fw.enqueue(event.Name, fileDebounce)
			}
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

func (fw *FileWatcher) enqueue(path string, delay time.Duration) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	if t, ok := fw.debounce[path]; ok {
		t.Stop()
	}
	fw.debounce[path] = time.AfterFunc(delay, func() {
		// block instead of dropping: a reconcile can run for minutes (discover), and a dropped
		// path would not be retried until its file changes again
		select {
		case fw.reconcile <- path:
		case <-fw.stop:
		}
	})
}

func (fw *FileWatcher) process(ctx context.Context, path string) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			fw.forget(path)
			return
		}
		logger.Errorf("read %s: %s", path, err)
		return
	}
	res, err := ParseManagedConfigMap(raw)
	if err != nil {
		logger.Errorf("parse %s: %s", path, err)
		return
	}

	fw.mu.Lock()
	// the file now holds another resource (metadata.name or kind changed): drop the old one
	for key, p := range fw.paths {
		if p == path && key != res.key() {
			fw.dropLocked(key)
		}
	}
	res.Annotations = fw.annotationsFor(res)
	fw.resources[res.key()] = *res
	fw.paths[res.key()] = path
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

// forget drops the resource read from path, as a deleted ConfigMap drops out of the in-cluster
// lists: jobs stop seeing it. The OLake entity is not deleted.
func (fw *FileWatcher) forget(path string) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	for key, p := range fw.paths {
		if p == path {
			fw.dropLocked(key)
		}
	}
}

// dropLocked removes a resource from the index. Caller must hold fw.mu.
func (fw *FileWatcher) dropLocked(key string) {
	delete(fw.resources, key)
	delete(fw.paths, key)
	delete(fw.state, key)
}

func (fw *FileWatcher) scheduleRetry(path string, after time.Duration) {
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
func (fw *FileWatcher) enqueueDependentJobs(changed *ResourceData) {
	fw.mu.Lock()
	changedRes := *changed
	changedRes.Annotations = fw.annotationsFor(&changedRes) // entity ID may be set by this reconcile

	var match func(job *ResourceData) bool
	switch changed.Kind {
	case KindStreams:
		match = func(job *ResourceData) bool {
			return matchesNameOrID(changedRes.JobRef(), job.Name, job.EntityID())
		}
	case KindSource, KindDestination:
		// a job refers to the OLake entity name from the object's config, not the object's name
		name := configName(changedRes.Config())
		match = func(job *ResourceData) bool {
			cfg, err := ParseAndValidateJobConfig(job.Config())
			if err != nil {
				return false
			}
			ref := cfg.Source
			if changed.Kind == KindDestination {
				ref = cfg.Destination
			}
			return matchesNameOrID(ref, name, changedRes.EntityID())
		}
	default:
		fw.mu.Unlock()
		return
	}

	var paths []string
	for key, r := range fw.resources {
		if r.Kind != KindJob || r.Namespace != changed.Namespace {
			continue
		}
		job := r
		job.Annotations = fw.annotationsFor(&job)
		if match(&job) {
			paths = append(paths, fw.paths[key])
		}
	}
	fw.mu.Unlock()

	for _, p := range paths {
		fw.enqueue(p, 0)
	}
}

func (fw *FileWatcher) sync(ctx context.Context, res *ResourceData) (ctrl.Result, error) {
	switch res.Kind {
	case KindSource:
		return reconcileResource(ctx, fw, fw.source, res)
	case KindDestination:
		return reconcileResource(ctx, fw, fw.dest, res)
	case KindJob:
		return fw.job.reconcileJob(ctx, res)
	case KindStreams:
		// no standalone reconciler; the job owns streams status (enqueueDependentJobs)
		return ctrl.Result{}, nil
	default:
		return ctrl.Result{}, fmt.Errorf("unsupported olake.io/kind %q", res.Kind)
	}
}

func (fw *FileWatcher) findStreams(_ context.Context, job *ResourceData, projectID string, entityID int) (*ResourceData, error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	for _, r := range fw.resources {
		if r.Kind != KindStreams || r.Namespace != job.Namespace {
			continue
		}
		if r.ProjectID() != projectID {
			continue
		}
		if matchesNameOrID(r.JobRef(), job.Name, entityID) {
			cp := r
			cp.Annotations = fw.annotationsFor(&cp)
			return &cp, nil
		}
	}
	return nil, ErrStreamsNotFound
}

// annotationsFor reads fw.state. Caller must hold fw.mu.
func (fw *FileWatcher) annotationsFor(r *ResourceData) map[string]string {
	st, ok := fw.state[r.key()]
	if !ok {
		return map[string]string{}
	}
	return map[string]string{
		AnnotationObservedHash: st.hash,
		AnnotationPhase:        st.phase,
		AnnotationMessage:      st.message,
		AnnotationEntityID:     st.entityID,
	}
}

func (fw *FileWatcher) SpawnIndicator(ctx context.Context, r *ResourceData, errMsg string) error {
	return spawnIndicator(ctx, fw.temporalClient, r, errMsg)
}

func (fw *FileWatcher) DeleteIndicator(ctx context.Context, r *ResourceData) error {
	return deleteIndicator(ctx, fw.temporalClient, r)
}

func (fw *FileWatcher) SetPhase(_ context.Context, r *ResourceData, phase, message, entityID, observedHash string) error {
	hash := observedHashForResource(r, observedHash)
	fw.mu.Lock()
	fw.state[r.key()] = resourceState{
		hash:     hash,
		phase:    phase,
		message:  message,
		entityID: entityID,
	}
	fw.mu.Unlock()
	return nil
}

// ParseManagedConfigMap unmarshals and validates an OLake GitOps ConfigMap YAML
// document. File mode requires every YAML file in GITOPS_FILE_DIR to be a
// labelled ConfigMap (Secrets are not supported on the filesystem).
func ParseManagedConfigMap(raw []byte) (*ResourceData, error) {
	var cm corev1.ConfigMap
	if err := yaml.Unmarshal(raw, &cm); err != nil {
		return nil, fmt.Errorf("yaml: %w", err)
	}
	if cm.Kind != "ConfigMap" {
		if cm.Kind == "" {
			return nil, fmt.Errorf("missing kind: file mode requires kind: ConfigMap")
		}
		return nil, fmt.Errorf("unsupported kind %q: file mode only supports ConfigMap", cm.Kind)
	}
	if cm.Labels[LabelManaged] != LabelManagedValue {
		return nil, fmt.Errorf("missing or invalid label %s=%q (required %q)", LabelManaged, cm.Labels[LabelManaged], LabelManagedValue)
	}
	kind := cm.Labels[LabelKind]
	if kind == "" {
		return nil, fmt.Errorf("missing label %q", LabelKind)
	}
	if !isValidManagedKind(kind) {
		return nil, fmt.Errorf("invalid label %s=%q (want source, destination, job, or streams)", LabelKind, kind)
	}
	if cm.Name == "" {
		return nil, fmt.Errorf("metadata.name is required")
	}
	res := resourceFromCM(&cm)
	return &res, nil
}

func isValidManagedKind(kind string) bool {
	switch kind {
	case KindSource, KindDestination, KindJob, KindStreams:
		return true
	default:
		return false
	}
}

func isYAML(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".yaml" || ext == ".yml"
}
