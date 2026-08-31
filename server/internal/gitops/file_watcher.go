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
	resources map[string]ResourceData
	state     map[string]resourceState
	debounce  map[string]*time.Timer
	retries   map[string]*time.Timer
	reconcile chan string
}

func newFileWatcher(dir string, etlSvc *etl.Service, t *temporal.Temporal) *FileWatcher {
	fw := &FileWatcher{
		Dir:            dir,
		temporalClient: t,
		resources:      map[string]ResourceData{},
		state:          map[string]resourceState{},
		debounce:       map[string]*time.Timer{},
		retries:        map[string]*time.Timer{},
		reconcile:      make(chan string, 64),
	}
	sink := StatusSink(fw)
	fw.source = &SourceReconciler{ETL: etlSvc, Sink: sink}
	fw.dest = &DestinationReconciler{ETL: etlSvc, Sink: sink}
	fw.job = &JobReconciler{ETL: etlSvc, Sink: sink, findStreams: fw.findStreams}
	return fw
}

func (fw *FileWatcher) Start(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("fsnotify: %w", err)
	}
	defer watcher.Close()

	if err := watcher.Add(fw.Dir); err != nil {
		return fmt.Errorf("watch %s: %w", fw.Dir, err)
	}

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
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&(fsnotify.Create|fsnotify.Write) != 0 && isYAML(event.Name) {
				fw.enqueue(event.Name, fileDebounce)
			}
			// v0: delete is no-op
		case err, ok := <-watcher.Errors:
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
		select {
		case fw.reconcile <- path:
		default:
			logger.Warnf("gitops file watcher queue full, dropping %s", path)
		}
	})
}

func (fw *FileWatcher) process(ctx context.Context, path string) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
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
	fw.resources[res.key()] = *res
	if st, ok := fw.state[res.key()]; ok {
		res.Annotations = map[string]string{
			AnnotationObservedHash: st.hash,
			AnnotationPhase:        st.phase,
			AnnotationMessage:      st.message,
			AnnotationEntityID:     st.entityID,
		}
	}
	fw.mu.Unlock()

	result, syncErr := fw.sync(ctx, res)
	if syncErr != nil {
		logger.Errorf("reconcile %s: %s", path, syncErr)
	}
	if result.RequeueAfter > 0 {
		fw.scheduleRetry(path, result.RequeueAfter)
	}

	switch res.Kind {
	case KindSource, KindDestination, KindStreams:
		fw.reconcileJobs(ctx)
	}
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

func (fw *FileWatcher) reconcileJobs(ctx context.Context) {
	fw.mu.Lock()
	var jobs []ResourceData
	for _, r := range fw.resources {
		if r.Kind == KindJob {
			jobs = append(jobs, r)
		}
	}
	fw.mu.Unlock()
	for i := range jobs {
		jobs[i].Annotations = fw.annotationsFor(&jobs[i])
		_, _ = fw.job.sync(ctx, &jobs[i])
	}
}

func (fw *FileWatcher) sync(ctx context.Context, res *ResourceData) (ctrl.Result, error) {
	switch res.Kind {
	case KindSource:
		return fw.source.sync(ctx, res)
	case KindDestination:
		return fw.dest.sync(ctx, res)
	case KindJob:
		return fw.job.sync(ctx, res)
	case KindStreams:
		return ctrl.Result{}, nil
	default:
		return ctrl.Result{}, fmt.Errorf("unsupported olake.io/kind %q", res.Kind)
	}
}

func (fw *FileWatcher) findStreams(_ context.Context, job *ResourceData, projectID string, entityID int) (*ResourceData, error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	for _, r := range fw.resources {
		if r.Kind != KindStreams {
			continue
		}
		if r.ProjectID() != projectID {
			continue
		}
		if matchesNameOrID(r.JobRef(), job.Name, entityID) {
			cp := r
			return &cp, nil
		}
	}
	return nil, ErrStreamsNotFound
}

func (fw *FileWatcher) annotationsFor(r *ResourceData) map[string]string {
	fw.mu.Lock()
	defer fw.mu.Unlock()
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
