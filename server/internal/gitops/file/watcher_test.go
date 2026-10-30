package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/datazip-inc/olake-ui/server/internal/gitops/reconciler"
)

// File mode tests run the real Watcher on a temp directory with stand-in reconcilers. They cover
// what file mode owns: reading manifests, the state file, restarts, and queueing the jobs a change
// affects. The reconcile rules are tested in the reconciler package. Tests call process directly
// instead of going through fsnotify and the debounce timer, which e2e covers.

const (
	testSelection = `{"selected_streams":{"public":[{"stream_name":"users","partition_regex":"","normalization":true}]}}`
	fileSource    = `{"kind":"source","projectId":"123","userId":1,"name":"pg","type":"postgres","version":"v1.0.0","config":{"host":"pg","port":5432}}`
	fileJob       = `{"kind":"job","projectId":"123","userId":1,"name":"orders","source":"pg","destination":"lake","frequency":"0 * * * *"}`
	fileStreams   = `{"kind":"streams","projectId":"123","job":"orders","streams":` + testSelection + `}`

	// the entity ids the stand-in reconcilers report
	sourceEntityID = "7"
	jobEntityID    = "8"
)

type fileTestEnv struct {
	t   *testing.T
	dir string
	fw  *Watcher

	// reconciled is every resource handed to a reconciler, in order, as the reconciler got it
	// (with the status the watcher restored)
	reconciled []reconciler.ResourceData
	// jobStreams is the streams resource handed to the last job reconcile, nil when there was none
	jobStreams *reconciler.ResourceData
}

// newFileTestEnv is an empty manifest directory with its default state file.
func newFileTestEnv(t *testing.T) *fileTestEnv {
	e := &fileTestEnv{t: t, dir: t.TempDir()}
	e.restart()
	return e
}

// restart starts a new Watcher on the same directory and state file, as olake-ui does when it
// restarts.
func (e *fileTestEnv) restart() {
	e.t.Helper()
	fw := newWatcher(e.dir, "", nil)
	fw.reconcileSource = e.reconcileConnector
	fw.reconcileDestination = e.reconcileConnector
	fw.reconcileJob = e.reconcileJob
	require.NoError(e.t, fw.readStateFile())
	e.t.Cleanup(func() {
		fw.mu.Lock()
		for _, timer := range fw.debounce {
			timer.Stop()
		}
		for _, timer := range fw.retries {
			timer.Stop()
		}
		fw.mu.Unlock()
		close(fw.stop) // releases timers that already fired and wait to enqueue
	})
	e.fw = fw
}

// reconcileConnector stands in for the source and destination reconcilers: it records the resource
// and reports it Ready through the watcher, with the hash of its data as a reconciler does.
func (e *fileTestEnv) reconcileConnector(ctx context.Context, res *reconciler.ResourceData) (ctrl.Result, error) {
	e.reconciled = append(e.reconciled, *res)
	return ctrl.Result{}, e.fw.SetPhase(ctx, res, reconciler.PhaseReady, "", sourceEntityID, reconciler.ContentHash(res.Data))
}

// reconcileJob stands in for the job reconciler: Pending without streams, else both Ready.
func (e *fileTestEnv) reconcileJob(ctx context.Context, job, streams *reconciler.ResourceData) (ctrl.Result, error) {
	e.reconciled = append(e.reconciled, *job)
	e.jobStreams = streams
	if streams == nil {
		return ctrl.Result{}, e.fw.SetPhase(ctx, job, reconciler.PhasePending, "waiting for Streams", "", reconciler.ContentHash(job.Data))
	}
	if err := e.fw.SetPhase(ctx, job, reconciler.PhaseReady, "", jobEntityID, reconciler.ContentHash(job.Data)); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, e.fw.SetPhase(ctx, streams, reconciler.PhaseReady, "", jobEntityID, reconciler.ContentHash(streams.Data))
}

// writeFile writes a manifest and returns its path.
func (e *fileTestEnv) writeFile(name, content string) string {
	e.t.Helper()
	path := filepath.Join(e.dir, name)
	require.NoError(e.t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

// process reconciles the file at path, as the watcher does once its debounce ends.
func (e *fileTestEnv) process(path string) {
	e.fw.process(context.Background(), path)
}

// nextEnqueued is the next path the watcher queued for reconcile, for example a job that a streams
// change unblocked.
func (e *fileTestEnv) nextEnqueued() string {
	e.t.Helper()
	select {
	case path := <-e.fw.reconcile:
		return path
	case <-time.After(2 * time.Second):
		e.t.Fatal("nothing was queued for reconcile")
		return ""
	}
}

// lastReconciled is the resource handed to the last reconcile.
func (e *fileTestEnv) lastReconciled() reconciler.ResourceData {
	e.t.Helper()
	require.NotEmpty(e.t, e.reconciled, "nothing was reconciled")
	return e.reconciled[len(e.reconciled)-1]
}

// stateOf is the status of a resource, read back from the state file as a restart would.
func (e *fileTestEnv) stateOf(kind, name string) (resourceState, bool) {
	e.t.Helper()
	reader := newWatcher(e.dir, "", nil)
	require.NoError(e.t, reader.readStateFile())
	st, ok := reader.state[resourceKey(&reconciler.ResourceData{Kind: kind, Name: name})]
	return st, ok
}

// ---
// UNIT TESTS
// ---

func TestWatcherStateFile(t *testing.T) {
	t.Run("a corrupt state file stops startup instead of replaying every job", func(t *testing.T) {
		e := newFileTestEnv(t)
		e.writeFile(stateFileName, `{"job//orders":`)
		err := newWatcher(e.dir, "", nil).readStateFile()
		assert.ErrorContains(t, err, "delete it to reconcile every resource from scratch")
	})

	t.Run("an unwritable state file stops startup", func(t *testing.T) {
		dir := t.TempDir()
		err := newWatcher(dir, filepath.Join(dir, "missing", "state.json"), nil).setup()
		assert.ErrorContains(t, err, "write gitops state")
	})

	t.Run("an unchanged status is not written again", func(t *testing.T) {
		e := newFileTestEnv(t)
		res := &reconciler.ResourceData{Kind: reconciler.KindSource, Name: "pg"}
		require.NoError(t, e.fw.SetPhase(context.Background(), res, reconciler.PhasePending, "retrying: db down", "", "h1"))
		stateFile := filepath.Join(e.dir, stateFileName)
		require.NoError(t, os.Remove(stateFile))

		// a transient error requeues every 30s with the same status
		require.NoError(t, e.fw.SetPhase(context.Background(), res, reconciler.PhasePending, "retrying: db down", "", "h1"))
		assert.NoFileExists(t, stateFile)

		require.NoError(t, e.fw.SetPhase(context.Background(), res, reconciler.PhaseReady, "", "7", "h1"))
		st, _ := e.stateOf(reconciler.KindSource, "pg")
		assert.Equal(t, resourceState{Hash: "h1", Phase: reconciler.PhaseReady, EntityID: "7"}, st)
	})

	t.Run("a failed write leaves the status unchanged so the next one retries", func(t *testing.T) {
		dir := t.TempDir()
		fw := newWatcher(dir, filepath.Join(dir, "missing", "state.json"), nil)
		res := &reconciler.ResourceData{Kind: reconciler.KindSource, Name: "pg"}
		err := fw.SetPhase(context.Background(), res, reconciler.PhaseReady, "", "7", "h")
		require.Error(t, err)
		assert.Empty(t, fw.state)
	})
}

// ---
// RECONCILE TESTS
// ---

func TestWatcherReconcilesSourceFile(t *testing.T) {
	e := newFileTestEnv(t)
	e.process(e.writeFile("pg.json", fileSource))

	res := e.lastReconciled()
	assert.Equal(t, reconciler.KindSource, res.Kind)
	assert.Equal(t, "pg", res.Name)
	assert.JSONEq(t, `{"name":"pg","type":"postgres","version":"v1.0.0","config":{"host":"pg","port":5432}}`, res.Config())

	st, ok := e.stateOf(reconciler.KindSource, "pg")
	require.True(t, ok)
	assert.Equal(t, resourceState{Hash: reconciler.ContentHash(res.Data), Phase: reconciler.PhaseReady, EntityID: sourceEntityID}, st)
}

func TestWatcherJobWaitsForStreamsFile(t *testing.T) {
	e := newFileTestEnv(t)
	jobPath := e.writeFile("orders.json", fileJob)
	e.process(jobPath)
	assert.Nil(t, e.jobStreams)
	st, _ := e.stateOf(reconciler.KindJob, "orders")
	assert.Equal(t, reconciler.PhasePending, st.Phase)

	// the streams file queues its job, which then gets it
	e.process(e.writeFile("orders-streams.json", fileStreams))
	require.Equal(t, jobPath, e.nextEnqueued())
	e.process(jobPath)

	require.NotNil(t, e.jobStreams)
	assert.Equal(t, "orders", e.jobStreams.Name)
	assert.JSONEq(t, testSelection, e.jobStreams.Config())
	for _, kind := range []string{reconciler.KindJob, reconciler.KindStreams} {
		st, _ := e.stateOf(kind, "orders")
		assert.Equal(t, reconciler.PhaseReady, st.Phase, kind)
		assert.Equal(t, jobEntityID, st.EntityID, kind)
	}
}

func TestWatcherRestartRestoresStatus(t *testing.T) {
	// the reconcilers skip a resource whose status matches its data; without the restored status a
	// restart would re-apply every job and cancel its running sync
	e := newFileTestEnv(t)
	paths := []string{e.writeFile("pg.json", fileSource), e.writeFile("orders-streams.json", fileStreams), e.writeFile("orders.json", fileJob)}
	for _, p := range paths {
		e.process(p)
	}

	e.restart()
	e.reconciled = nil
	for _, p := range paths {
		e.process(p)
	}
	require.Len(t, e.reconciled, 2, "the source and the job; streams have no reconciler of their own")
	for _, res := range e.reconciled {
		assert.Equal(t, reconciler.PhaseReady, res.Annotations[reconciler.AnnotationPhase], res.Kind)
		assert.Equal(t, reconciler.ContentHash(res.Data), res.Annotations[reconciler.AnnotationObservedHash], res.Kind)
	}
	assert.Equal(t, jobEntityID, e.jobStreams.Annotations[reconciler.AnnotationEntityID], "the job's streams carry their status too")
}

func TestWatcherDeletedFile(t *testing.T) {
	e := newFileTestEnv(t)
	path := e.writeFile("pg.json", fileSource)
	e.process(path)

	require.NoError(t, os.Remove(path))
	e.process(path)

	assert.Len(t, e.reconciled, 1, "a deleted file is not reconciled, so the OLake source is kept")
	_, ok := e.stateOf(reconciler.KindSource, "pg")
	assert.False(t, ok, "the status goes with the file, so a file that comes back is reconciled from scratch")
}

func TestWatcherFileRenamesEntity(t *testing.T) {
	e := newFileTestEnv(t)
	path := e.writeFile("pg.json", fileSource)
	e.process(path)

	e.writeFile("pg.json", `{"kind":"source","projectId":"123","userId":1,"name":"pg2","type":"postgres","version":"v1.0.0","config":{"host":"pg","port":5432}}`)
	e.process(path)

	_, ok := e.stateOf(reconciler.KindSource, "pg")
	assert.False(t, ok, "the old entity is dropped, so jobs stop seeing it")
	st, _ := e.stateOf(reconciler.KindSource, "pg2")
	assert.Equal(t, reconciler.PhaseReady, st.Phase)
}

func TestWatcherSkipsUnparseableFile(t *testing.T) {
	e := newFileTestEnv(t)
	path := e.writeFile("pg.json", fileSource)
	e.process(path)

	// a half-saved file is logged and ignored; the last good version and its status stay
	e.writeFile("pg.json", `{"kind":"source","name":`)
	e.process(path)
	assert.Len(t, e.reconciled, 1)
	st, _ := e.stateOf(reconciler.KindSource, "pg")
	assert.Equal(t, reconciler.PhaseReady, st.Phase)
}
