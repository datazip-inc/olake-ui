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

const fileDebounce = 200 * time.Millisecond

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
	index     map[string]ResourceData
	state     map[string]resourceState
	debounce  map[string]*time.Timer
	retries   map[string]*time.Timer
	reconcile chan string
}

func newFileWatcher(dir string, etlSvc *etl.Service, t *temporal.Temporal) *FileWatcher {
	fw := &FileWatcher{
		Dir:            dir,
		temporalClient: t,
		index:          map[string]ResourceData{},
		state:     map[string]resourceState{},
		debounce:  map[string]*time.Timer{},
		retries:   map[string]*time.Timer{},
		reconcile: make(chan string, 64),
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
			// DELETE is a no-op (same as k8s v0: entity stays in DB).
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
	res, skip, err := ParseManagedConfigMap(raw, filepath.Base(path))
	if err != nil {
		logger.Errorf("parse %s: %s", path, err)
		return
	}
	if skip {
		return
	}

	fw.mu.Lock()
	fw.index[res.key()] = *res
	if st, ok := fw.state[res.key()]; ok {
		res.Annotations = map[string]string{
			AnnotationObservedHash: st.hash,
			AnnotationPhase:        st.phase,
			AnnotationMessage:      st.message,
			AnnotationEntityID:     st.entityID,
		}
	}
	fw.mu.Unlock()

	result, syncErr := fw.sync(ctx, *res)
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
	for _, r := range fw.index {
		if r.Kind == KindJob {
			jobs = append(jobs, r)
		}
	}
	fw.mu.Unlock()
	for _, job := range jobs {
		job.Annotations = fw.annotationsFor(job)
		_, _ = fw.job.sync(ctx, job)
	}
}

func (fw *FileWatcher) sync(ctx context.Context, res ResourceData) (ctrl.Result, error) {
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
		return ctrl.Result{}, nil
	}
}

func (fw *FileWatcher) findStreams(_ context.Context, job ResourceData, projectID string, entityID int) (*ResourceData, error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	for _, r := range fw.index {
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
	return nil, nil
}

func (fw *FileWatcher) annotationsFor(r ResourceData) map[string]string {
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

func (fw *FileWatcher) SpawnIndicator(ctx context.Context, r ResourceData, errMsg string) error {
	return spawnIndicatorViaTemporal(ctx, fw.temporalClient, r, errMsg)
}

func (fw *FileWatcher) DeleteIndicator(ctx context.Context, r ResourceData) error {
	return deleteIndicatorViaTemporal(ctx, fw.temporalClient, r)
}

func (fw *FileWatcher) SetPhase(_ context.Context, r ResourceData, phase, message, entityID string) error {
	fw.mu.Lock()
	fw.state[r.key()] = resourceState{
		hash:     ContentHash(r.Data),
		phase:    phase,
		message:  message,
		entityID: entityID,
	}
	fw.mu.Unlock()
	return nil
}

// ParseManagedConfigMap unmarshals a ConfigMap YAML document. skip is true when
// the document is not an OLake-managed ConfigMap.
func ParseManagedConfigMap(raw []byte, filename string) (*ResourceData, bool, error) {
	var cm corev1.ConfigMap
	if err := yaml.Unmarshal(raw, &cm); err != nil {
		return nil, false, fmt.Errorf("yaml: %w", err)
	}
	if cm.Kind != "" && cm.Kind != "ConfigMap" {
		return nil, true, nil
	}
	if cm.Labels[LabelManaged] != LabelManagedValue {
		return nil, true, nil
	}
	kind := cm.Labels[LabelKind]
	if kind == "" {
		return nil, true, nil
	}
	name := cm.Name
	if name == "" {
		name = strings.TrimSuffix(filename, filepath.Ext(filename))
	}
	res := resourceFromCM(&cm)
	res.Name = name
	res.Kind = kind
	return &res, false, nil
}

func isYAML(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".yaml" || ext == ".yml"
}
