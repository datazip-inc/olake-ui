package reconciler

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
	"github.com/datazip-inc/olake-ui/server/internal/types"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
)

const (
	jobConfigMap     = "orders-cm"
	streamsConfigMap = "orders-streams"
	ordersSelection  = `{"selected_streams":{"public":[{"stream_name":"orders","partition_regex":"","normalization":false}]}}`
	// fakeETL's stream difference output, as stored in UpdateJobRequest.DifferenceStreams
	streamDiffJSON = `{"selected_streams":{"public":["users"]}}`
)

var errTemporalDown = fmt.Errorf("%w: dial tcp temporal:7233: connection refused", temporal.ErrWorkflowStart)

// newJobTestEnv is a source "pg", a destination "lake", the Job ConfigMap for job "orders" and its Streams ConfigMap.
func newJobTestEnv(t *testing.T, jobFields map[string]interface{}) (*testEnv, *models.Source, *models.Destination) {
	t.Helper()
	e := newTestEnv(t,
		newObject(KindJob, jobConfigMap, manifestData(jobConfigJSON(jobFields))),
		newObject(KindStreams, streamsConfigMap, streamsManifestData(jobConfigMap, userSelectedStreams)),
	)
	return e, e.etl.addSource("pg", "v1.0.0"), e.etl.addDestination("lake")
}

// newSyncedJobTestEnv is newJobTestEnv after a first reconcile created the job; recorded calls are reset.
func newSyncedJobTestEnv(t *testing.T, jobFields map[string]interface{}) (*testEnv, *models.Job) {
	t.Helper()
	e, _, _ := newJobTestEnv(t, jobFields)
	e.reconcileJob(jobConfigMap)
	require.Equal(t, PhaseReady, e.annotationsOf(jobConfigMap)[AnnotationPhase], e.annotationsOf(jobConfigMap)[AnnotationMessage])
	e.resetRecordings()
	return e, e.etl.jobs["orders"]
}

// Typed views of fakeETL's recorded calls, in call order.
func discoverStoredCatalogs(fakeService *fakeETL) []types.StreamsCatalog {
	return callArgs[types.StreamsCatalog](fakeService, "DiscoverWithCatalog")
}
func discoverRequests(fakeService *fakeETL) []*dto.StreamsRequest {
	return callArgs[*dto.StreamsRequest](fakeService, "DiscoverWithCatalog")
}
func streamDiffRequests(fakeService *fakeETL) []dto.StreamDifferenceRequest {
	return callArgs[dto.StreamDifferenceRequest](fakeService, "GetStreamDifference")
}
func createJobRequests(fakeService *fakeETL) []*dto.CreateJobRequest {
	return callArgs[*dto.CreateJobRequest](fakeService, "CreateJob")
}
func updateJobRequests(fakeService *fakeETL) []*dto.UpdateJobRequest {
	return callArgs[*dto.UpdateJobRequest](fakeService, "UpdateJob")
}
func activateJobRequests(fakeService *fakeETL) []dto.JobStatusRequest {
	return callArgs[dto.JobStatusRequest](fakeService, "ActivateJob")
}

// ---
// UNIT TESTS
// ---

// TestAdvancedSettingsEqual checks how stored advanced settings compare with the ConfigMap's:
// nil and empty are equal, removing or adding settings is a change, and JSON key order is not.
func TestAdvancedSettingsEqual(t *testing.T) {
	eight, nine := 8, 9
	settings := &dto.AdvancedSettings{MaxDiscoverThreads: &eight, TargetQueryEngines: []string{"spark"}}
	strPtr := func(s string) *string { return &s }

	assert.True(t, advancedSettingsEqual(nil, nil))
	assert.True(t, advancedSettingsEqual(strPtr(""), nil))
	assert.False(t, advancedSettingsEqual(strPtr(`{"max_discover_threads":8}`), nil), "removing settings from the ConfigMap is a change")
	assert.False(t, advancedSettingsEqual(nil, settings))
	assert.True(t, advancedSettingsEqual(strPtr(`{"target_query_engines":["spark"], "max_discover_threads":8}`), settings), "key order is not a change")
	assert.False(t, advancedSettingsEqual(strPtr(`{"max_discover_threads":8}`), settings))
	assert.False(t, advancedSettingsEqual(strPtr(`{"max_discover_threads":8,"target_query_engines":["spark"]}`), &dto.AdvancedSettings{MaxDiscoverThreads: &nine, TargetQueryEngines: []string{"spark"}}))
}

// ---
// RECONCILE TESTS
// ---

// TestJobReconcileCreate covers the happy path for a new job.
// It checks that both connectors are tested, discover runs twice (fresh, then merged with the ConfigMap selection),
// CreateJob gets the right request and user, both ConfigMaps end Ready with the new entity ID,
// and a second reconcile of unchanged objects does nothing.
func TestJobReconcileCreate(t *testing.T) {
	e, src, dest := newJobTestEnv(t, map[string]interface{}{"advanced_settings": map[string]interface{}{"max_discover_threads": 8}})
	e.reconcileJob(jobConfigMap)

	// both connectors are tested, and the destination test carries the job's source
	assert.Equal(t, 1, e.etl.callCount("TestSourceConnection"))
	assert.Equal(t, []*dto.DestinationTestConnectionRequest{{
		Type: "iceberg", Version: "v1.0.0", Config: `{"catalog_type":"glue"}`, SourceType: "postgres", SourceVersion: "v1.0.0",
	}}, callArgs[*dto.DestinationTestConnectionRequest](e.etl, "TestDestinationConnection"))

	// a new job discovers the source, then merges the ConfigMap's selection with it
	assert.Equal(t, []types.StreamsCatalog{{}, {Available: e.etl.discoveredStreams, Selected: userSelectedStreams}}, discoverStoredCatalogs(e.etl))
	for _, req := range discoverRequests(e.etl) {
		assert.Equal(t, -1, req.JobID)
		assert.Equal(t, "orders", req.JobName)
		assert.Equal(t, 8, *req.MaxDiscoverThreads)
	}

	assert.Equal(t, []string{"CreateJob"}, e.etl.writeCalls(), "a new active job needs no ActivateJob")
	maxDiscoverThreads := 8
	assert.Equal(t, []*dto.CreateJobRequest{{
		JobMetadata: dto.JobMetadata{
			Name: "orders", Frequency: "0 * * * *", Activate: true,
			AdvancedSettings: &dto.AdvancedSettings{MaxDiscoverThreads: &maxDiscoverThreads},
		},
		Source:                 &dto.DriverConfig{ID: &src.ID, Version: "v1.0.0"},
		Destination:            &dto.DriverConfig{ID: &dest.ID},
		AvailableStreamsConfig: e.etl.discoveredStreams,
		SelectedStreamsConfig:  userSelectedStreams,
	}}, createJobRequests(e.etl))
	userID := 1
	assert.Equal(t, []*int{&userID}, callArgs[*int](e.etl, "CreateJob"), "the user is data.userId")

	job := e.etl.jobs["orders"]
	for _, name := range []string{jobConfigMap, streamsConfigMap} {
		assert.Equal(t, PhaseReady, e.phaseOf(name))
		assert.Equal(t, idString(job.ID), e.annotationsOf(name)[AnnotationEntityID])
		assert.Equal(t, []string{indicatorDelete}, e.status.indicatorActionsFor(name))
	}

	// the same objects again change nothing in OLake
	e.resetRecordings()
	e.reconcileJob(jobConfigMap)
	assert.Empty(t, e.etl.writeCalls())
	assert.Zero(t, e.etl.callCount("DiscoverWithCatalog"))
	assert.Zero(t, e.etl.callCount("TestSourceConnection"))
	assert.Empty(t, e.status.statusUpdates)
}

// TestJobReconcileCreateInactive checks that a new job with activate=false is created inactive by CreateJob alone.
func TestJobReconcileCreateInactive(t *testing.T) {
	e, _, _ := newJobTestEnv(t, map[string]interface{}{"activate": false})
	e.reconcileJob(jobConfigMap)
	require.Equal(t, []string{"CreateJob"}, e.etl.writeCalls(), "CreateJob applies the activation itself")
	assert.False(t, createJobRequests(e.etl)[0].Activate)
}

// TestJobReconcileWaitsForDependencies covers a job whose source, destination or Streams ConfigMap is missing
// (or sits in another project or namespace).
// The job stays Pending with a "waiting for" message, requeues, writes nothing to OLake and spawns no failure indicator.
// Once the dependency appears, the next reconcile creates the job.
func TestJobReconcileWaitsForDependencies(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(e *testEnv)
		wantMessage string
		resolve     func(e *testEnv)
	}{
		{
			name:        "source",
			setup:       func(e *testEnv) { delete(e.etl.sources, "pg") },
			wantMessage: `waiting for source "pg"`,
			resolve:     func(e *testEnv) { e.etl.addSource("pg", "v1.0.0") },
		},
		{
			name:        "destination",
			setup:       func(e *testEnv) { delete(e.etl.destinations, "lake") },
			wantMessage: `waiting for destination "lake"`,
			resolve:     func(e *testEnv) { e.etl.addDestination("lake") },
		},
		{
			name: "streams",
			setup: func(e *testEnv) {
				e.deleteObject(streamsConfigMap)
			},
			wantMessage: `waiting for Streams referencing job "orders-cm"`,
			resolve: func(e *testEnv) {
				e.createObject(newObject(KindStreams, streamsConfigMap, streamsManifestData(jobConfigMap, userSelectedStreams)))
			},
		},
		{
			name: "streams in another project",
			setup: func(e *testEnv) {
				data := streamsManifestData(jobConfigMap, userSelectedStreams)
				data[DataProjectID] = "999"
				e.setData(streamsConfigMap, data)
			},
			wantMessage: `waiting for Streams referencing job "orders-cm"`,
			resolve: func(e *testEnv) {
				e.setData(streamsConfigMap, streamsManifestData(jobConfigMap, userSelectedStreams))
			},
		},
		{
			name: "streams in another namespace",
			setup: func(e *testEnv) {
				e.deleteObject(streamsConfigMap)
				other := newObject(KindStreams, streamsConfigMap, streamsManifestData(jobConfigMap, userSelectedStreams))
				other.Namespace = "other"
				e.createObject(other)
			},
			wantMessage: `waiting for Streams referencing job "orders-cm"`,
			resolve: func(e *testEnv) {
				e.createObject(newObject(KindStreams, streamsConfigMap, streamsManifestData(jobConfigMap, userSelectedStreams)))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, _, _ := newJobTestEnv(t, nil)
			tt.setup(e)
			result := e.reconcileJob(jobConfigMap)
			assert.Equal(t, syncRequeueAfter, result.RequeueAfter)
			assert.Equal(t, PhasePending, e.phaseOf(jobConfigMap))
			assert.Equal(t, tt.wantMessage, e.messageOf(jobConfigMap))
			assert.Empty(t, e.status.indicatorActions, "waiting for a dependency is not a failure")
			assert.Empty(t, e.etl.writeCalls())

			tt.resolve(e)
			e.reconcileJob(jobConfigMap)
			assert.Equal(t, PhaseReady, e.phaseOf(jobConfigMap), e.messageOf(jobConfigMap))
			assert.Equal(t, []string{"CreateJob"}, e.etl.writeCalls())
		})
	}
}

// TestJobReconcileFailures covers non-retryable failures on create.
// Each case ends Failed with no requeue, spawns a failure indicator, and shows the cause in the status message.
// The Streams ConfigMap shares the failure unless it was never read.
// wantWrites checks that a failure before CreateJob writes nothing to OLake.
func TestJobReconcileFailures(t *testing.T) {
	tests := []struct {
		name        string
		jobFields   map[string]interface{}
		setup       func(e *testEnv)
		wantMessage string
		// streamsFailed is false when the failure is found before the Streams ConfigMap is read
		streamsFailed bool
		wantWrites    []string
	}{
		{
			name:        "invalid job config",
			jobFields:   map[string]interface{}{"frequency": nil},
			wantMessage: "validation failed",
		},
		{
			name: "missing userId",
			setup: func(e *testEnv) {
				data := manifestData(jobConfigJSON(nil))
				delete(data, DataUserID)
				e.setData(jobConfigMap, data)
			},
			wantMessage: "data.userId must be a positive integer",
		},
		{
			name: "legacy streams ConfigMap",
			setup: func(e *testEnv) {
				e.setData(streamsConfigMap, streamsManifestData(jobConfigMap, `{"streams":[{"stream":{"name":"users"}}],"selected_streams":{}}`))
			},
			wantMessage:   "legacy streams config (streams[]) is not supported",
			streamsFailed: true,
		},
		{
			name:          "source below the streams v2 version",
			setup:         func(e *testEnv) { e.etl.sources["pg"].Version = "v0.9.0" },
			wantMessage:   `GitOps requires source version v1.0.0 or later (source "pg" is on v0.9.0)`,
			streamsFailed: true,
		},
		{
			name: "source connection fails",
			setup: func(e *testEnv) {
				e.etl.operationResults[testSourceOperationID] = operationResult{output: connectionCheckOutput("FAILED", "password authentication failed")}
			},
			wantMessage:   "password authentication failed",
			streamsFailed: true,
		},
		{
			name: "destination connection fails",
			setup: func(e *testEnv) {
				e.etl.operationResults[testDestinationOperationID] = operationResult{output: connectionCheckOutput("FAILED", "bucket not found")}
			},
			wantMessage:   "bucket not found",
			streamsFailed: true,
		},
		{
			name:          "discover fails in the CLI",
			setup:         func(e *testEnv) { e.etl.failingMethods["DiscoverWithCatalog"] = errors.New(`schema "nope" not found`) },
			wantMessage:   `schema "nope" not found`,
			streamsFailed: true,
		},
		{
			name:          "create rejected",
			setup:         func(e *testEnv) { e.etl.failingMethods["CreateJob"] = errors.New("invalid cron expression") },
			wantMessage:   "invalid cron expression",
			streamsFailed: true,
			wantWrites:    []string{"CreateJob"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, _, _ := newJobTestEnv(t, tt.jobFields)
			if tt.setup != nil {
				tt.setup(e)
			}
			result := e.reconcileJob(jobConfigMap)
			assert.Zero(t, result.RequeueAfter, "a non-retryable failure is not requeued")
			assert.Equal(t, PhaseFailed, e.phaseOf(jobConfigMap))
			assert.Contains(t, e.messageOf(jobConfigMap), tt.wantMessage)
			assert.Equal(t, []string{indicatorSpawn}, e.status.indicatorActionsFor(jobConfigMap))
			assert.Equal(t, tt.wantWrites, e.etl.writeCalls())
			if tt.streamsFailed {
				assert.Equal(t, PhaseFailed, e.phaseOf(streamsConfigMap))
				assert.Contains(t, e.messageOf(streamsConfigMap), tt.wantMessage)
				assert.Equal(t, []string{indicatorSpawn}, e.status.indicatorActionsFor(streamsConfigMap))
			} else {
				assert.Empty(t, e.phaseOf(streamsConfigMap))
			}
		})
	}
}

// TestJobReconcileTransientErrors covers outages (database or Temporal down) during create.
// The job stays Pending with a "retrying" message, requeues, spawns no indicator and writes nothing.
// After the outage ends, the next reconcile creates the job.
func TestJobReconcileTransientErrors(t *testing.T) {
	tests := []struct {
		name  string
		setup func(e *testEnv)
	}{
		{name: "source lookup", setup: func(e *testEnv) { e.etl.failingMethods["GetSourceByName"] = errors.New("db: connection refused") }},
		{name: "destination lookup", setup: func(e *testEnv) { e.etl.failingMethods["GetDestinationByName"] = errors.New("db: connection refused") }},
		{name: "job lookup", setup: func(e *testEnv) { e.etl.failingMethods["GetJobByName"] = errors.New("db: connection refused") }},
		{name: "connection test start", setup: func(e *testEnv) { e.etl.failingMethods["TestSourceConnection"] = errTemporalDown }},
		{name: "discover start", setup: func(e *testEnv) {
			e.etl.failingMethods["DiscoverWithCatalog"] = fmt.Errorf("discover: %w", errTemporalDown)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, _, _ := newJobTestEnv(t, nil)
			tt.setup(e)
			result := e.reconcileJob(jobConfigMap)
			assert.Equal(t, syncRequeueAfter, result.RequeueAfter)
			assert.Equal(t, PhasePending, e.phaseOf(jobConfigMap))
			assert.Contains(t, e.messageOf(jobConfigMap), "retrying: ")
			assert.Empty(t, e.status.indicatorActions)
			assert.Empty(t, e.etl.writeCalls())
			assert.Empty(t, e.phaseOf(streamsConfigMap))

			// the outage ends: the next requeue creates the job
			clear(e.etl.failingMethods)
			e.reconcileJob(jobConfigMap)
			assert.Equal(t, PhaseReady, e.phaseOf(jobConfigMap))
		})
	}
}

// TestJobReconcileRetriesWhenSourceIsFixed checks that a Failed job is not retried while its inputs are unchanged,
// but is retried when only the source changes (new updated_at), with no edit to the Job ConfigMap.
// A successful retry deletes the failure indicators of both ConfigMaps.
func TestJobReconcileRetriesWhenSourceIsFixed(t *testing.T) {
	e, _, _ := newJobTestEnv(t, nil)
	e.etl.failingMethods["DiscoverWithCatalog"] = errors.New(`schema "nope" not found`)
	e.reconcileJob(jobConfigMap)
	require.Equal(t, PhaseFailed, e.phaseOf(jobConfigMap))

	// unchanged inputs: the failure stands and discover does not run again
	e.resetRecordings()
	e.reconcileJob(jobConfigMap)
	assert.Zero(t, e.etl.callCount("DiscoverWithCatalog"))

	// fixing only the source (a new updated_at) retries the job without a Job ConfigMap edit
	delete(e.etl.failingMethods, "DiscoverWithCatalog")
	e.etl.sources["pg"].UpdatedAt = time.Unix(2000, 0)
	e.resetRecordings()
	e.reconcileJob(jobConfigMap)
	assert.Equal(t, PhaseReady, e.phaseOf(jobConfigMap))
	assert.Equal(t, PhaseReady, e.phaseOf(streamsConfigMap))
	assert.Equal(t, []string{indicatorDelete}, e.status.indicatorActionsFor(jobConfigMap))
	assert.Equal(t, []string{indicatorDelete}, e.status.indicatorActionsFor(streamsConfigMap))
}

// TestJobReconcileFailedStreamsWaitForAChange checks that a Failed Streams ConfigMap counts as settled for its data.
// Reconciling again writes no status, so status patches do not loop the job.
// Fixing the Streams data lets both ConfigMaps reach Ready.
func TestJobReconcileFailedStreamsWaitForAChange(t *testing.T) {
	e, _ := newSyncedJobTestEnv(t, nil)
	e.setData(streamsConfigMap, streamsManifestData(jobConfigMap, `{"selected_streams":{}}`))
	e.reconcileJob(jobConfigMap)
	require.Equal(t, PhaseFailed, e.phaseOf(streamsConfigMap))

	// a failed Streams ConfigMap is settled for its data, so status patches do not loop the job
	e.resetRecordings()
	e.reconcileJob(jobConfigMap)
	assert.Empty(t, e.status.statusUpdates)

	e.setData(streamsConfigMap, streamsManifestData(jobConfigMap, ordersSelection))
	e.reconcileJob(jobConfigMap)
	assert.Equal(t, PhaseReady, e.phaseOf(jobConfigMap))
	assert.Equal(t, PhaseReady, e.phaseOf(streamsConfigMap))
}

// TestJobReconcileUpdate covers an existing job that drifts from its ConfigMaps.
// Each subtest changes one thing and checks the exact OLake calls:
// which of discover, connection tests, stream difference, UpdateJob and ActivateJob run, and which must not.
func TestJobReconcileUpdate(t *testing.T) {
	t.Run("job field change updates with the stored catalog", func(t *testing.T) {
		e, job := newSyncedJobTestEnv(t, nil)
		e.setData(jobConfigMap, manifestData(jobConfigJSON(map[string]interface{}{"frequency": "*/5 * * * *"})))
		e.reconcileJob(jobConfigMap)

		assert.Equal(t, []string{"UpdateJob"}, e.etl.writeCalls())
		assert.Equal(t, []int{job.ID}, callArgs[int](e.etl, "UpdateJob"))
		assert.Zero(t, e.etl.callCount("DiscoverWithCatalog"), "only streams or connector changes need a discover")
		assert.Zero(t, e.etl.callCount("TestSourceConnection"))
		assert.Zero(t, e.etl.callCount("GetStreamDifference"))
		// the stored catalog is sent back unchanged, with no stream difference
		assert.Equal(t, []*dto.UpdateJobRequest{{
			JobMetadata:            dto.JobMetadata{Name: "orders", Frequency: "*/5 * * * *", Activate: true},
			Source:                 &dto.DriverConfig{ID: &job.SourceID, Version: "v1.0.0"},
			Destination:            &dto.DriverConfig{ID: &job.DestID},
			AvailableStreamsConfig: *job.AvailableStreamsConfig,
			SelectedStreamsConfig:  *job.SelectedStreamsConfig,
		}}, updateJobRequests(e.etl))
		assert.Equal(t, PhaseReady, e.phaseOf(jobConfigMap))
	})

	t.Run("advanced settings change updates the job", func(t *testing.T) {
		e, _ := newSyncedJobTestEnv(t, map[string]interface{}{"advanced_settings": map[string]interface{}{"max_discover_threads": 8}})
		e.setData(jobConfigMap, manifestData(jobConfigJSON(nil)))
		e.reconcileJob(jobConfigMap)
		require.Equal(t, []string{"UpdateJob"}, e.etl.writeCalls(), "removing advanced_settings from the ConfigMap is a change")
		assert.Nil(t, updateJobRequests(e.etl)[0].AdvancedSettings)
	})

	t.Run("streams change runs discover and the stream difference", func(t *testing.T) {
		e, job := newSyncedJobTestEnv(t, nil)
		e.setData(streamsConfigMap, streamsManifestData(jobConfigMap, ordersSelection))
		e.reconcileJob(jobConfigMap)

		assert.Zero(t, e.etl.callCount("TestSourceConnection"), "connectors did not change")
		require.Equal(t, []types.StreamsCatalog{{Available: *job.AvailableStreamsConfig, Selected: ordersSelection}}, discoverStoredCatalogs(e.etl))
		assert.Equal(t, job.ID, discoverRequests(e.etl)[0].JobID)

		assert.Equal(t, []dto.StreamDifferenceRequest{{UpdatedAvailableStreamsConfig: e.etl.discoveredStreams, UpdatedSelectedStreamsConfig: ordersSelection}}, streamDiffRequests(e.etl))
		assert.Equal(t, []int{job.ID}, callArgs[int](e.etl, "GetStreamDifference"))

		require.Equal(t, []string{"UpdateJob"}, e.etl.writeCalls())
		req := updateJobRequests(e.etl)[0]
		assert.Equal(t, ordersSelection, req.SelectedStreamsConfig)
		assert.JSONEq(t, streamDiffJSON, req.DifferenceStreams)
		assert.Equal(t, PhaseReady, e.phaseOf(streamsConfigMap))

		// the applied selection is not diffed again
		e.resetRecordings()
		e.reconcileJob(jobConfigMap)
		assert.Zero(t, e.etl.callCount("GetStreamDifference"))
		assert.Empty(t, e.etl.writeCalls())
	})

	t.Run("destination change tests connectors and re-merges", func(t *testing.T) {
		e, job := newSyncedJobTestEnv(t, nil)
		otherDestination := e.etl.addDestination("otherDestination")
		e.setData(jobConfigMap, manifestData(jobConfigJSON(map[string]interface{}{"destination": "otherDestination"})))
		e.reconcileJob(jobConfigMap)

		assert.Equal(t, 1, e.etl.callCount("TestSourceConnection"))
		assert.Equal(t, 1, e.etl.callCount("TestDestinationConnection"))
		assert.Equal(t, []types.StreamsCatalog{job.StreamsCatalog()}, discoverStoredCatalogs(e.etl), "the stored catalog is merged with the source")
		assert.Zero(t, e.etl.callCount("GetStreamDifference"), "the selection did not change")
		require.Equal(t, []string{"UpdateJob"}, e.etl.writeCalls())
		assert.Equal(t, otherDestination.ID, *updateJobRequests(e.etl)[0].Destination.ID)
	})

	t.Run("activation change alone goes through ActivateJob", func(t *testing.T) {
		e, job := newSyncedJobTestEnv(t, nil)
		e.setData(jobConfigMap, manifestData(jobConfigJSON(map[string]interface{}{"activate": false})))
		e.reconcileJob(jobConfigMap)
		assert.Equal(t, []string{"ActivateJob"}, e.etl.writeCalls())
		assert.Equal(t, []dto.JobStatusRequest{{Activate: false}}, activateJobRequests(e.etl))
		assert.Equal(t, []int{job.ID}, callArgs[int](e.etl, "ActivateJob"))
		assert.Equal(t, PhaseReady, e.phaseOf(jobConfigMap))
	})

	t.Run("activation and field change keep UpdateJob on the current activation", func(t *testing.T) {
		e, _ := newSyncedJobTestEnv(t, nil)
		e.setData(jobConfigMap, manifestData(jobConfigJSON(map[string]interface{}{"activate": false, "frequency": "*/5 * * * *"})))
		e.reconcileJob(jobConfigMap)
		// UpdateJob stores activation without pausing the schedule, so it must not carry the change
		require.Equal(t, []string{"UpdateJob", "ActivateJob"}, e.etl.writeCalls())
		assert.True(t, updateJobRequests(e.etl)[0].Activate)
		assert.False(t, activateJobRequests(e.etl)[0].Activate)
	})
}

// TestJobReconcileUpdateWhileClearDestinationRuns checks that an update rejected because a clear-destination is running
// is a wait, not a failure: Pending, requeued, no indicator, entity ID kept.
// When the clear ends, the next reconcile diffs again and applies the update.
func TestJobReconcileUpdateWhileClearDestinationRuns(t *testing.T) {
	e, job := newSyncedJobTestEnv(t, nil)
	e.setData(streamsConfigMap, streamsManifestData(jobConfigMap, ordersSelection))
	e.etl.failingMethods["UpdateJob"] = fmt.Errorf("%w: job %d", constants.ErrClearDestinationRunning, job.ID)

	result := e.reconcileJob(jobConfigMap)
	assert.Equal(t, syncRequeueAfter, result.RequeueAfter)
	assert.Equal(t, PhasePending, e.phaseOf(jobConfigMap), "a running clear-destination is a wait, never a failure")
	assert.Contains(t, e.messageOf(jobConfigMap), constants.ErrClearDestinationRunning.Error())
	assert.Equal(t, idString(job.ID), e.annotationsOf(jobConfigMap)[AnnotationEntityID])
	assert.Empty(t, e.status.indicatorActions)

	// the clear ends: the requeue diffs again and applies the edit
	delete(e.etl.failingMethods, "UpdateJob")
	e.resetRecordings()
	e.reconcileJob(jobConfigMap)
	assert.Equal(t, 1, e.etl.callCount("GetStreamDifference"))
	assert.Equal(t, []string{"UpdateJob"}, e.etl.writeCalls())
	assert.Equal(t, PhaseReady, e.phaseOf(jobConfigMap))
	assert.Equal(t, PhaseReady, e.phaseOf(streamsConfigMap))
}

// TestJobReconcileUpdateErrors covers failures on the update path:
// stream difference cannot start (Pending) or fails (Failed), UpdateJob rejected, and ActivateJob rejected (both Failed).
// wantWrites checks which OLake writes were attempted before the failure.
func TestJobReconcileUpdateErrors(t *testing.T) {
	tests := []struct {
		name       string
		jobFields  map[string]interface{}
		streams    string
		setup      func(e *testEnv)
		wantPhase  string
		wantWrites []string
	}{
		{
			name: "stream difference cannot start", streams: ordersSelection,
			setup:     func(e *testEnv) { e.etl.failingMethods["GetStreamDifference"] = errTemporalDown },
			wantPhase: PhasePending,
		},
		{
			name: "stream difference fails", streams: ordersSelection,
			setup: func(e *testEnv) {
				e.etl.operationResults[streamDiffOperationID] = operationResult{err: errors.New("difference failed")}
			},
			wantPhase: PhaseFailed,
		},
		{
			name: "update rejected", jobFields: map[string]interface{}{"frequency": "*/5 * * * *"},
			setup:     func(e *testEnv) { e.etl.failingMethods["UpdateJob"] = errors.New("invalid cron expression") },
			wantPhase: PhaseFailed, wantWrites: []string{"UpdateJob"},
		},
		{
			name: "activation rejected", jobFields: map[string]interface{}{"activate": false},
			setup:     func(e *testEnv) { e.etl.failingMethods["ActivateJob"] = errors.New("schedule not found") },
			wantPhase: PhaseFailed, wantWrites: []string{"ActivateJob"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, _ := newSyncedJobTestEnv(t, nil)
			if tt.jobFields != nil {
				e.setData(jobConfigMap, manifestData(jobConfigJSON(tt.jobFields)))
			}
			if tt.streams != "" {
				e.setData(streamsConfigMap, streamsManifestData(jobConfigMap, tt.streams))
			}
			tt.setup(e)
			e.reconcileJob(jobConfigMap)
			assert.Equal(t, tt.wantPhase, e.phaseOf(jobConfigMap), e.messageOf(jobConfigMap))
			assert.Equal(t, tt.wantWrites, e.etl.writeCalls())
		})
	}
}

// TestJobReconcileStatusWriteFails checks that a failed status write after CreateJob requeues the job.
// The retry finds the created job by name, does not create it again, and writes Ready on both ConfigMaps.
func TestJobReconcileStatusWriteFails(t *testing.T) {
	e, _, _ := newJobTestEnv(t, nil)
	e.status.failStatusUpdates = errors.New("the object has been modified")
	result := e.reconcileJob(jobConfigMap)
	assert.Equal(t, syncRequeueAfter, result.RequeueAfter, "the job is retried until its status is written")
	require.Equal(t, 1, e.etl.callCount("CreateJob"))

	// the retry finds the created job and does not create it again
	e.status.failStatusUpdates = nil
	e.reconcileJob(jobConfigMap)
	assert.Equal(t, 1, e.etl.callCount("CreateJob"))
	assert.Equal(t, PhaseReady, e.phaseOf(jobConfigMap))
	assert.Equal(t, PhaseReady, e.phaseOf(streamsConfigMap))
}

// TestJobReconcileCreatedButReloadFailed checks that when the read-back after CreateJob fails, the job goes Pending.
// The next reconcile finds the created job and does not create a second one.
func TestJobReconcileCreatedButReloadFailed(t *testing.T) {
	e, _, _ := newJobTestEnv(t, nil)
	e.etl.failCall = func(method string) error {
		if method == "GetJobByName" && e.etl.callCount("CreateJob") == 1 && e.etl.callCount("GetJobByName") == 2 {
			return errors.New("db: connection reset")
		}
		return nil
	}
	e.reconcileJob(jobConfigMap)
	assert.Equal(t, PhasePending, e.phaseOf(jobConfigMap))

	e.reconcileJob(jobConfigMap)
	assert.Equal(t, 1, e.etl.callCount("CreateJob"))
	assert.Equal(t, PhaseReady, e.phaseOf(jobConfigMap))
}

// TestJobReconcileAfterStatusLost checks recovery when the status annotations are wiped (kubectl replace, forced sync).
// The job is found by name, not recreated, and connectors are not retested.
// Streams count as unapplied, so one stream difference runs; it is empty because the selection is unchanged.
// Both ConfigMaps return to Ready and the entity ID annotation is restored.
func TestJobReconcileAfterStatusLost(t *testing.T) {
	// kubectl replace or a forced sync drops the status annotations of both objects
	e, job := newSyncedJobTestEnv(t, nil)
	e.clearAnnotations(jobConfigMap)
	e.clearAnnotations(streamsConfigMap)
	e.reconcileJob(jobConfigMap)

	assert.Zero(t, e.etl.callCount("CreateJob"), "the job is found by name, not recreated")
	assert.Zero(t, e.etl.callCount("TestSourceConnection"))
	// the selection is unchanged, so the CLI's difference is empty and nothing is cleared
	assert.Equal(t, []dto.StreamDifferenceRequest{{
		UpdatedAvailableStreamsConfig: *job.AvailableStreamsConfig,
		UpdatedSelectedStreamsConfig:  *job.SelectedStreamsConfig,
	}}, streamDiffRequests(e.etl))
	assert.Equal(t, PhaseReady, e.phaseOf(jobConfigMap))
	assert.Equal(t, idString(job.ID), e.annotationsOf(jobConfigMap)[AnnotationEntityID])
	assert.Equal(t, PhaseReady, e.phaseOf(streamsConfigMap))
}

// TestJobReconcileAdoptedJobFailureKeepsEntityID checks that when the first update of an adopted job is rejected,
// both ConfigMaps end Failed but still carry the adopted job's entity ID.
func TestJobReconcileAdoptedJobFailureKeepsEntityID(t *testing.T) {
	// a job created in the UI, adopted by a Job ConfigMap whose first update is rejected
	e, src, dest := newJobTestEnv(t, map[string]interface{}{"frequency": "*/5 * * * *"})
	e.etl.jobs["orders"] = &models.Job{
		ID: 77, Name: "orders", ProjectID: testProjectID, SourceID: src.ID, DestID: dest.ID, Active: true,
		Frequency: "0 * * * *", AvailableStreamsConfig: utils.StringPtr(e.etl.discoveredStreams), SelectedStreamsConfig: utils.StringPtr(userSelectedStreams),
	}
	e.etl.failingMethods["UpdateJob"] = errors.New("invalid cron expression")
	e.reconcileJob(jobConfigMap)

	for _, name := range []string{jobConfigMap, streamsConfigMap} {
		assert.Equal(t, PhaseFailed, e.phaseOf(name))
		assert.Equal(t, "77", e.annotationsOf(name)[AnnotationEntityID], "%s manages job 77, even though the update failed", name)
	}
}

// TestJobReconcileRefsByID checks that the Job ConfigMap can refer to its source and destination by entity ID,
// and the Streams ConfigMap can refer to the job by entity ID.
// ID refs must not trigger name lookups for the source or destination.
func TestJobReconcileRefsByID(t *testing.T) {
	e, src, dest := newJobTestEnv(t, nil)
	e.setData(jobConfigMap, manifestData(jobConfigJSON(map[string]interface{}{"source": idString(src.ID), "destination": idString(dest.ID)})))
	e.reconcileJob(jobConfigMap)
	require.Equal(t, PhaseReady, e.phaseOf(jobConfigMap), e.messageOf(jobConfigMap))
	assert.Zero(t, e.etl.callCount("GetSourceByName"))
	assert.Zero(t, e.etl.callCount("GetDestinationByName"))
	assert.Equal(t, src.ID, *createJobRequests(e.etl)[0].Source.ID)
	assert.Equal(t, dest.ID, *createJobRequests(e.etl)[0].Destination.ID)

	// the Streams ConfigMap may point at the synced job by its entity id
	job := e.etl.jobs["orders"]
	e.setData(streamsConfigMap, streamsManifestData(idString(job.ID), ordersSelection))
	e.reconcileJob(jobConfigMap)
	assert.Equal(t, PhaseReady, e.phaseOf(streamsConfigMap), e.messageOf(jobConfigMap))
	assert.Equal(t, ordersSelection, *e.etl.jobs["orders"].SelectedStreamsConfig)
}

// TestJobReconcileRequest covers olake.io/reconcile-request on the Job and Streams ConfigMaps: each
// re-runs a synced job, and each object echoes its own request.
func TestJobReconcileRequest(t *testing.T) {
	t.Run("a job Failed after a long outage is retried on request", func(t *testing.T) {
		e, _ := newSyncedJobTestEnv(t, nil)
		e.setData(jobConfigMap, manifestData(jobConfigJSON(map[string]interface{}{"frequency": "*/5 * * * *"})))
		e.etl.failingMethods["GetJobByName"] = errors.New("db: connection refused")
		retryMap.Store(e.load(jobConfigMap).key(), time.Now().Add(-syncRetryTimeout))
		e.reconcileJob(jobConfigMap)
		require.Equal(t, PhaseFailed, e.phaseOf(jobConfigMap))

		// the outage ends, but nothing in git changed, so the Failed job stays skipped
		clear(e.etl.failingMethods)
		e.resetRecordings()
		e.reconcileJob(jobConfigMap)
		assert.Zero(t, e.etl.callCount("GetJobByName"))
		assert.Empty(t, e.status.statusUpdates)

		e.object(jobConfigMap).Annotations[AnnotationReconcileRequest] = "1"
		e.reconcileJob(jobConfigMap)
		assert.Equal(t, []string{"UpdateJob"}, e.etl.writeCalls())
		assert.Equal(t, PhaseReady, e.phaseOf(jobConfigMap))
		assert.Equal(t, "1", e.annotationsOf(jobConfigMap)[AnnotationReconcileHandled])
		assert.Empty(t, e.annotationsOf(streamsConfigMap)[AnnotationReconcileHandled], "streams had no request")

		e.resetRecordings()
		e.reconcileJob(jobConfigMap)
		assert.Zero(t, e.etl.callCount("GetJobByName"), "the answered request is not run again")
		assert.Empty(t, e.status.statusUpdates)
	})

	t.Run("a request on a Ready job re-checks it without changing OLake", func(t *testing.T) {
		e, _ := newSyncedJobTestEnv(t, nil)
		e.object(jobConfigMap).Annotations[AnnotationReconcileRequest] = "1"
		e.reconcileJob(jobConfigMap)
		assert.Equal(t, 1, e.etl.callCount("GetJobByName"))
		assert.Empty(t, e.etl.writeCalls())
		assert.Equal(t, PhaseReady, e.phaseOf(jobConfigMap))
		assert.Equal(t, "1", e.annotationsOf(jobConfigMap)[AnnotationReconcileHandled])
	})

	t.Run("a request on Failed streams is answered and not looped", func(t *testing.T) {
		e, _ := newSyncedJobTestEnv(t, nil)
		e.setData(streamsConfigMap, streamsManifestData(jobConfigMap, `{"selected_streams":{}}`))
		e.reconcileJob(jobConfigMap)
		require.Equal(t, PhaseFailed, e.phaseOf(streamsConfigMap))

		e.object(streamsConfigMap).Annotations[AnnotationReconcileRequest] = "1"
		e.resetRecordings()
		e.reconcileJob(jobConfigMap)
		assert.NotEmpty(t, e.etl.calls, "the request re-runs the job")
		assert.Equal(t, PhaseFailed, e.phaseOf(streamsConfigMap), "the selection is still invalid")
		assert.Equal(t, "1", e.annotationsOf(streamsConfigMap)[AnnotationReconcileHandled])

		e.resetRecordings()
		e.reconcileJob(jobConfigMap)
		assert.Empty(t, e.status.statusUpdates)
	})
}
