package gitops

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
)

// File mode tests run the real FileWatcher on a temp directory against fakeETL. They call process
// directly instead of going through fsnotify and the debounce timer, which e2e covers.

const (
	fileSource  = `{"kind":"source","projectId":"123","userId":1,"name":"pg","type":"postgres","version":"v1.0.0","config":{"host":"pg","port":5432}}`
	fileJob     = `{"kind":"job","projectId":"123","userId":1,"name":"orders","source":"pg","destination":"lake","frequency":"0 * * * *"}`
	fileStreams = `{"kind":"streams","projectId":"123","job":"orders","streams":` + userSelectedStreams + `}`
)

type fileTestEnv struct {
	t   *testing.T
	dir string
	etl *fakeETL
	fw  *FileWatcher
}

// newFileTestEnv is an empty manifest directory with its default state file, and an empty OLake.
func newFileTestEnv(t *testing.T) *fileTestEnv {
	t.Cleanup(func() { retryMap.Clear() })
	e := &fileTestEnv{t: t, dir: t.TempDir(), etl: newFakeETL()}
	e.restart()
	return e
}

// restart starts a new FileWatcher on the same directory, state file and OLake, as olake-ui does
// when it restarts.
func (e *fileTestEnv) restart() {
	e.t.Helper()
	fw := newFileWatcher(e.dir, "", e.etl, &temporal.Temporal{Client: &fakeTemporalClient{}})
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

// stateOf is the status of a resource, read back from the state file as a restart would.
func (e *fileTestEnv) stateOf(kind, name string) (resourceState, bool) {
	e.t.Helper()
	reader := newFileWatcher(e.dir, "", nil, nil)
	require.NoError(e.t, reader.readStateFile())
	st, ok := reader.state[(&ResourceData{Kind: kind, Name: name}).key()]
	return st, ok
}

// ---
// UNIT TESTS
// ---

func TestFileWatcherStateFile(t *testing.T) {
	t.Run("a corrupt state file stops startup instead of replaying every job", func(t *testing.T) {
		e := newFileTestEnv(t)
		e.writeFile(stateFileName, `{"job//orders":`)
		err := newFileWatcher(e.dir, "", nil, nil).readStateFile()
		assert.ErrorContains(t, err, "delete it to reconcile every resource from scratch")
	})

	t.Run("an unwritable state file stops startup", func(t *testing.T) {
		dir := t.TempDir()
		err := newFileWatcher(dir, filepath.Join(dir, "missing", "state.json"), nil, nil).setup()
		assert.ErrorContains(t, err, "write gitops state")
	})

	t.Run("an unchanged status is not written again", func(t *testing.T) {
		e := newFileTestEnv(t)
		res := &ResourceData{Kind: KindSource, Name: "pg", Data: manifestData(sourceConfigJSON)}
		require.NoError(t, e.fw.SetPhase(context.Background(), res, PhasePending, "retrying: db down", "", ""))
		stateFile := filepath.Join(e.dir, stateFileName)
		require.NoError(t, os.Remove(stateFile))

		// a transient error requeues every 30s with the same status
		require.NoError(t, e.fw.SetPhase(context.Background(), res, PhasePending, "retrying: db down", "", ""))
		assert.NoFileExists(t, stateFile)

		require.NoError(t, e.fw.SetPhase(context.Background(), res, PhaseReady, "", "7", ""))
		st, _ := e.stateOf(KindSource, "pg")
		assert.Equal(t, resourceState{Hash: ContentHash(res.Data), Phase: PhaseReady, EntityID: "7"}, st)
	})

	t.Run("a failed write leaves the status unchanged so the next one retries", func(t *testing.T) {
		dir := t.TempDir()
		fw := newFileWatcher(dir, filepath.Join(dir, "missing", "state.json"), nil, nil)
		res := &ResourceData{Kind: KindSource, Name: "pg"}
		err := fw.SetPhase(context.Background(), res, PhaseReady, "", "7", "h")
		require.Error(t, err)
		assert.Empty(t, fw.state)
	})
}

// ---
// RECONCILE TESTS
// ---

func TestFileWatcherSyncsSource(t *testing.T) {
	e := newFileTestEnv(t)
	e.process(e.writeFile("pg.json", fileSource))

	assert.Equal(t, []string{"CreateSource"}, e.etl.writeCalls())
	st, ok := e.stateOf(KindSource, "pg")
	require.True(t, ok)
	assert.Equal(t, PhaseReady, st.Phase, st.Message)
	assert.Equal(t, idString(e.etl.sources["pg"].ID), st.EntityID)
}

func TestFileWatcherJobWaitsForStreamsFile(t *testing.T) {
	e := newFileTestEnv(t)
	e.etl.addSource("pg", "v1.0.0")
	e.etl.addDestination("lake")
	jobPath := e.writeFile("orders.json", fileJob)
	e.process(jobPath)

	st, _ := e.stateOf(KindJob, "orders")
	assert.Equal(t, PhasePending, st.Phase)
	assert.Equal(t, `waiting for Streams referencing job "orders"`, st.Message)
	assert.Empty(t, e.etl.writeCalls())

	// the streams file queues its job, which then syncs
	e.process(e.writeFile("orders-streams.json", fileStreams))
	require.Equal(t, jobPath, e.nextEnqueued())
	e.process(jobPath)

	assert.Equal(t, []string{"CreateJob"}, e.etl.writeCalls())
	for _, kind := range []string{KindJob, KindStreams} {
		st, _ := e.stateOf(kind, "orders")
		assert.Equal(t, PhaseReady, st.Phase, "%s: %s", kind, st.Message)
		assert.Equal(t, idString(e.etl.jobs["orders"].ID), st.EntityID)
	}
}

func TestFileWatcherRestartDoesNotReapply(t *testing.T) {
	// re-applying a job on restart would cancel its running sync
	e := newFileTestEnv(t)
	e.etl.addDestination("lake")
	// streams before the job, so the job finds them on its first reconcile
	paths := []string{e.writeFile("pg.json", fileSource), e.writeFile("orders-streams.json", fileStreams), e.writeFile("orders.json", fileJob)}
	for _, p := range paths {
		e.process(p)
	}
	st, _ := e.stateOf(KindJob, "orders")
	require.Equal(t, PhaseReady, st.Phase, st.Message)

	e.restart()
	e.etl.calls = nil
	for _, p := range paths {
		e.process(p)
	}
	assert.Empty(t, e.etl.writeCalls())
	assert.Zero(t, e.etl.callCount("TestSourceConnection"))
	assert.Zero(t, e.etl.callCount("DiscoverWithCatalog"))
}

func TestFileWatcherDeletedFile(t *testing.T) {
	e := newFileTestEnv(t)
	path := e.writeFile("pg.json", fileSource)
	e.process(path)

	require.NoError(t, os.Remove(path))
	e.process(path)

	_, ok := e.stateOf(KindSource, "pg")
	assert.False(t, ok, "the status goes with the file, so a file that comes back is reconciled from scratch")
	assert.Contains(t, e.etl.sources, "pg", "the OLake source is kept")
}

func TestFileWatcherFileRenamesEntity(t *testing.T) {
	e := newFileTestEnv(t)
	path := e.writeFile("pg.json", fileSource)
	e.process(path)

	e.writeFile("pg.json", `{"kind":"source","projectId":"123","userId":1,"name":"pg2","type":"postgres","version":"v1.0.0","config":{"host":"pg","port":5432}}`)
	e.process(path)

	_, ok := e.stateOf(KindSource, "pg")
	assert.False(t, ok, "the old entity is dropped, so jobs stop seeing it")
	st, _ := e.stateOf(KindSource, "pg2")
	assert.Equal(t, PhaseReady, st.Phase)
}

func TestFileWatcherSkipsUnparseableFile(t *testing.T) {
	e := newFileTestEnv(t)
	path := e.writeFile("pg.json", fileSource)
	e.process(path)

	// a half-saved file is logged and ignored; the last good version and its status stay
	e.writeFile("pg.json", `{"kind":"source","name":`)
	e.process(path)
	st, _ := e.stateOf(KindSource, "pg")
	assert.Equal(t, PhaseReady, st.Phase)
	assert.Equal(t, []string{"CreateSource"}, e.etl.writeCalls())
}
