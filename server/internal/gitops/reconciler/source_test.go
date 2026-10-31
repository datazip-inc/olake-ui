package reconciler

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/datazip-inc/olake-ui/server/internal/gitops/secretref"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
)

const sourceConfigJSON = `{"name":"pg","type":"postgres","version":"v1.0.0","config":{"host":"pg","port":5432}}`

func TestSourceReconcileCreate(t *testing.T) {
	e := newTestEnv(t, newObject(KindSource, "pg-cm", manifestData(sourceConfigJSON)))
	e.reconcileSource("pg-cm")

	assert.Equal(t, []string{"CreateSource"}, e.etl.writeCalls())
	assert.Equal(t, []*dto.CreateSourceRequest{{Name: "pg", Type: "postgres", Version: "v1.0.0", Config: `{"host":"pg","port":5432}`}},
		callArgs[*dto.CreateSourceRequest](e.etl, "CreateSource"))
	userID := 1
	assert.Equal(t, []*int{&userID}, callArgs[*int](e.etl, "CreateSource"), "the user is data.userId")
	assert.Equal(t, 1, e.etl.callCount("TestSourceConnection"), "a new source is tested before it is created")

	annotations := e.annotationsOf("pg-cm")
	assert.Equal(t, PhaseReady, annotations[AnnotationPhase])
	assert.Equal(t, idString(e.etl.sources["pg"].ID), annotations[AnnotationEntityID])
	assert.Equal(t, []string{indicatorDelete}, e.status.indicatorActionsFor("pg-cm"))

	// the same data again is skipped without touching OLake
	e.resetRecordings()
	e.reconcileSource("pg-cm")
	assert.Empty(t, e.etl.calls)
}

func TestSourceReconcileExisting(t *testing.T) {
	t.Run("equal config is adopted without a test or an update", func(t *testing.T) {
		e := newTestEnv(t, newObject(KindSource, "pg-cm", manifestData(`{"name":"pg","type":"postgres","version":"v1.0.0","config":{"port":5432,"host":"pg"}}`)))
		src := e.etl.addSource("pg", "v1.0.0")
		e.reconcileSource("pg-cm")
		assert.Empty(t, e.etl.writeCalls())
		assert.Zero(t, e.etl.callCount("TestSourceConnection"))
		assert.Equal(t, idString(src.ID), e.annotationsOf("pg-cm")[AnnotationEntityID])
	})

	t.Run("changed config is tested then updated", func(t *testing.T) {
		e := newTestEnv(t, newObject(KindSource, "pg-cm", manifestData(`{"name":"pg","type":"postgres","version":"v1.1.0","config":{"host":"pg","port":5432}}`)))
		src := e.etl.addSource("pg", "v1.0.0")
		e.reconcileSource("pg-cm")
		assert.Equal(t, []string{"UpdateSource"}, e.etl.writeCalls())
		assert.Equal(t, []int{src.ID}, callArgs[int](e.etl, "UpdateSource"))
		assert.Equal(t, []*dto.UpdateSourceRequest{{Name: "pg", Type: "postgres", Version: "v1.1.0", Config: `{"host":"pg","port":5432}`}},
			callArgs[*dto.UpdateSourceRequest](e.etl, "UpdateSource"))
		assert.Equal(t, 1, e.etl.callCount("TestSourceConnection"))
		assert.Equal(t, PhaseReady, e.phaseOf("pg-cm"))
	})

	t.Run("rejected update fails and keeps the stored source", func(t *testing.T) {
		e := newTestEnv(t, newObject(KindSource, "pg-cm", manifestData(`{"name":"pg","type":"postgres","version":"v1.1.0","config":{"host":"pg","port":5432}}`)))
		src := e.etl.addSource("pg", "v1.0.0")
		e.etl.failingMethods["UpdateSource"] = errors.New("version not found")
		e.reconcileSource("pg-cm")
		assert.Equal(t, PhaseFailed, e.phaseOf("pg-cm"))
		assert.Contains(t, e.messageOf("pg-cm"), "version not found")
		assert.Equal(t, idString(src.ID), e.annotationsOf("pg-cm")[AnnotationEntityID], "the object manages the found source, even though the update failed")
		assert.Equal(t, []string{indicatorSpawn}, e.status.indicatorActionsFor("pg-cm"))
	})

	t.Run("failed test keeps the stored source", func(t *testing.T) {
		e := newTestEnv(t, newObject(KindSource, "pg-cm", manifestData(`{"name":"pg","type":"postgres","version":"v1.0.0","config":{"host":"pg","password":"wrong"}}`)))
		e.etl.addSource("pg", "v1.0.0")
		e.etl.operationResults[testSourceOperationID] = operationResult{output: connectionCheckOutput("FAILED", "password authentication failed")}
		e.reconcileSource("pg-cm")
		assert.Empty(t, e.etl.writeCalls())
		assert.Equal(t, PhaseFailed, e.phaseOf("pg-cm"))
		assert.Contains(t, e.messageOf("pg-cm"), "password authentication failed")
		assert.Equal(t, []string{indicatorSpawn}, e.status.indicatorActionsFor("pg-cm"))
	})
}

func TestSourceReconcileFailures(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]string
		setup       func(fakeService *fakeETL)
		wantPhase   string
		wantMessage string
		wantWrites  []string
	}{
		{name: "invalid spec", data: manifestData(`{"name":"pg","type":"nope","version":"v1.0.0","config":{}}`), wantPhase: PhaseFailed, wantMessage: "invalid source type"},
		{name: "bad userId", data: map[string]string{DataProjectID: testProjectID, DataUserID: "x", DataConfig: sourceConfigJSON}, wantPhase: PhaseFailed, wantMessage: "data.userId must be a positive integer"},
		{
			name: "database unreachable on lookup", data: manifestData(sourceConfigJSON),
			setup: func(fakeService *fakeETL) {
				fakeService.failingMethods["GetSourceByName"] = errors.New("connection refused")
			},
			wantPhase: PhasePending, wantMessage: "retrying: connection refused",
		},
		{
			name: "temporal unreachable on test", data: manifestData(sourceConfigJSON),
			setup: func(fakeService *fakeETL) {
				fakeService.failingMethods["TestSourceConnection"] = fmt.Errorf("%w: refused", temporal.ErrWorkflowStart)
			},
			wantPhase: PhasePending, wantMessage: "retrying:",
		},
		{
			name: "create rejected", data: manifestData(sourceConfigJSON),
			setup:     func(fakeService *fakeETL) { fakeService.failingMethods["CreateSource"] = errors.New("duplicate name") },
			wantPhase: PhaseFailed, wantMessage: "duplicate name", wantWrites: []string{"CreateSource"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEnv(t, newObject(KindSource, "pg-cm", tt.data))
			if tt.setup != nil {
				tt.setup(e.etl)
			}
			e.reconcileSource("pg-cm")
			annotations := e.annotationsOf("pg-cm")
			assert.Equal(t, tt.wantPhase, annotations[AnnotationPhase])
			assert.Contains(t, annotations[AnnotationMessage], tt.wantMessage)
			assert.Equal(t, tt.wantWrites, e.etl.writeCalls())
			if tt.wantPhase == PhaseFailed {
				assert.Equal(t, []string{indicatorSpawn}, e.status.indicatorActionsFor("pg-cm"))
			} else {
				assert.Empty(t, e.status.indicatorActions, "a retry is not a failure")
			}
		})
	}
}

func TestSourceReconcileFailedRetriesOnlyOnChange(t *testing.T) {
	e := newTestEnv(t, newObject(KindSource, "pg-cm", manifestData(`{"name":"pg","type":"nope","version":"v1.0.0","config":{}}`)))
	e.reconcileSource("pg-cm")
	require.Equal(t, PhaseFailed, e.phaseOf("pg-cm"))

	e.resetRecordings()
	e.reconcileSource("pg-cm")
	assert.Empty(t, e.status.statusUpdates, "a failure for unchanged data is not retried")

	e.setData("pg-cm", manifestData(sourceConfigJSON))
	e.reconcileSource("pg-cm")
	assert.Equal(t, PhaseReady, e.phaseOf("pg-cm"))
	assert.Equal(t, []string{indicatorDelete}, e.status.indicatorActionsFor("pg-cm"), "fixing the data removes the indicator")
}

func TestSourceReconcileCreatedButReloadFailed(t *testing.T) {
	e := newTestEnv(t, newObject(KindSource, "pg-cm", manifestData(sourceConfigJSON)))
	// the DB drops right after the insert, so the reload of the new source fails
	e.etl.failCall = func(method string) error {
		if method == "GetSourceByName" && e.etl.callCount("CreateSource") == 1 && e.etl.callCount("GetSourceByName") == 2 {
			return errors.New("connection reset")
		}
		return nil
	}
	e.reconcileSource("pg-cm")
	assert.Equal(t, PhasePending, e.phaseOf("pg-cm"))
	assert.Contains(t, e.messageOf("pg-cm"), "connection reset")

	// the retry finds the created source and does not create it twice
	e.reconcileSource("pg-cm")
	assert.Equal(t, 1, e.etl.callCount("CreateSource"))
	assert.Equal(t, PhaseReady, e.phaseOf("pg-cm"))
	assert.Equal(t, idString(e.etl.sources["pg"].ID), e.annotationsOf("pg-cm")[AnnotationEntityID])
}

func TestSourceReconcileStatusWriteFails(t *testing.T) {
	e := newTestEnv(t, newObject(KindSource, "pg-cm", manifestData(sourceConfigJSON)))
	e.status.failStatusUpdates = errors.New("the object has been modified")
	result := e.reconcileSource("pg-cm")
	assert.Equal(t, syncRequeueAfter, result.RequeueAfter, "the source is retried until its status is written")
	require.Equal(t, 1, e.etl.callCount("CreateSource"))

	// the retry finds the created source and does not create it again
	e.status.failStatusUpdates = nil
	e.reconcileSource("pg-cm")
	assert.Equal(t, 1, e.etl.callCount("CreateSource"))
	assert.Equal(t, PhaseReady, e.phaseOf("pg-cm"))
}

func TestSourceReconcileSecretRefresh(t *testing.T) {
	const config = `{"name":"pg","type":"postgres","version":"v1.0.0","config":{"host":"pg","password":"${aws-sm:olake/pg}"}}`
	e := newTestEnv(t, newObject(KindSource, "pg-cm", manifestData(config)))
	provider := &fakeProvider{secret: "p1"}
	e.secrets = &SecretRefs{
		resolver: secretref.NewWithProviders(map[string]secretref.Provider{"aws-sm": provider}),
		refresh:  time.Hour,
	}

	result := e.reconcileSource("pg-cm")
	assert.Equal(t, []string{"CreateSource"}, e.etl.writeCalls())
	assert.Equal(t, PhaseReady, e.phaseOf("pg-cm"))
	assert.Equal(t, []string{indicatorDelete}, e.status.indicatorActionsFor("pg-cm"))
	assert.GreaterOrEqual(t, result.RequeueAfter, time.Hour, "a source with references is re-checked")

	// the re-check with the same secret writes nothing: no update, no status, no indicator
	e.resetRecordings()
	result = e.reconcileSource("pg-cm")
	assert.Empty(t, e.etl.writeCalls())
	assert.Zero(t, e.etl.callCount("TestSourceConnection"))
	assert.Empty(t, e.status.statusUpdates)
	assert.Empty(t, e.status.indicatorActions)
	assert.GreaterOrEqual(t, result.RequeueAfter, time.Hour)

	// a rotated secret is tested and updated in OLake; the status stays Ready, so it is not rewritten
	provider.secret = "p2"
	e.resetRecordings()
	result = e.reconcileSource("pg-cm")
	assert.Equal(t, []string{"UpdateSource"}, e.etl.writeCalls())
	assert.Equal(t, 1, e.etl.callCount("TestSourceConnection"))
	assert.Contains(t, callArgs[*dto.UpdateSourceRequest](e.etl, "UpdateSource")[0].Config, `"password":"p2"`)
	assert.Empty(t, e.status.statusUpdates)
	assert.Empty(t, e.status.indicatorActions)
	assert.GreaterOrEqual(t, result.RequeueAfter, time.Hour)

	// a Failed source is not re-checked: it waits for a change in git
	e.object("pg-cm").Annotations[AnnotationPhase] = PhaseFailed
	e.resetRecordings()
	result = e.reconcileSource("pg-cm")
	assert.Empty(t, e.etl.calls)
	assert.Zero(t, result.RequeueAfter)
}

// TestSourceReconcileRequest covers olake.io/reconcile-request: a new value re-runs a synced source
// in any phase, the reconcile echoes it into olake.io/reconcile-handled, and an answered request
// leaves the source skipped.
func TestSourceReconcileRequest(t *testing.T) {
	const config = `{"name":"pg","type":"postgres","version":"v1.0.0","config":{"host":"pg","password":"${aws-sm:olake/pg}"}}`
	newSecretSourceEnv := func(t *testing.T, provider *fakeProvider) *testEnv {
		e := newTestEnv(t, newObject(KindSource, "pg-cm", manifestData(config)))
		e.secrets = &SecretRefs{
			resolver: secretref.NewWithProviders(map[string]secretref.Provider{"aws-sm": provider}),
			refresh:  time.Hour,
		}
		return e
	}
	request := func(e *testEnv, value string) { e.object("pg-cm").Annotations[AnnotationReconcileRequest] = value }

	t.Run("a Failed source is retried once its secret is fixed", func(t *testing.T) {
		provider := &fakeProvider{err: errors.Join(secretref.ErrUnresolvable, errors.New("secret olake/pg not found"))}
		e := newSecretSourceEnv(t, provider)
		e.reconcileSource("pg-cm")
		require.Equal(t, PhaseFailed, e.phaseOf("pg-cm"))

		// fixing the secret alone changes nothing in git, so the Failed source stays skipped
		provider.err, provider.secret = nil, "p1"
		e.resetRecordings()
		e.reconcileSource("pg-cm")
		assert.Empty(t, e.etl.calls)

		request(e, "1")
		e.resetRecordings()
		e.reconcileSource("pg-cm")
		require.Equal(t, []string{"CreateSource"}, e.etl.writeCalls())
		assert.Contains(t, callArgs[*dto.CreateSourceRequest](e.etl, "CreateSource")[0].Config, `"password":"p1"`)
		assert.Equal(t, PhaseReady, e.phaseOf("pg-cm"))
		assert.Equal(t, "1", e.annotationsOf("pg-cm")[AnnotationReconcileHandled])
		assert.Equal(t, []string{indicatorDelete}, e.status.indicatorActionsFor("pg-cm"))

		// the answered request is not run again: no loop on the status write it caused
		e.resetRecordings()
		e.reconcileSource("pg-cm")
		assert.Empty(t, e.status.statusUpdates)
	})

	t.Run("a request that fails again is answered and not looped", func(t *testing.T) {
		provider := &fakeProvider{err: errors.Join(secretref.ErrUnresolvable, errors.New("access denied"))}
		e := newSecretSourceEnv(t, provider)
		e.reconcileSource("pg-cm")
		require.Equal(t, PhaseFailed, e.phaseOf("pg-cm"))

		request(e, "1")
		e.resetRecordings()
		e.reconcileSource("pg-cm")
		assert.Equal(t, PhaseFailed, e.phaseOf("pg-cm"))
		assert.Equal(t, "1", e.annotationsOf("pg-cm")[AnnotationReconcileHandled])

		e.resetRecordings()
		e.reconcileSource("pg-cm")
		assert.Empty(t, e.status.statusUpdates)
	})

	t.Run("a Ready source without references is re-checked", func(t *testing.T) {
		e := newTestEnv(t, newObject(KindSource, "pg-cm", manifestData(sourceConfigJSON)))
		e.reconcileSource("pg-cm")
		require.Equal(t, PhaseReady, e.phaseOf("pg-cm"))

		request(e, "2026-10-31T10:00:00Z")
		e.resetRecordings()
		e.reconcileSource("pg-cm")
		assert.Equal(t, 1, e.etl.callCount("GetSourceByName"), "OLake is read again")
		assert.Empty(t, e.etl.writeCalls(), "and matches, so nothing is written to it")
		assert.Equal(t, []statusUpdate{{objectName: "pg-cm", phase: PhaseReady, entityID: idString(e.etl.sources["pg"].ID)}}, e.status.statusUpdates)
		assert.Equal(t, "2026-10-31T10:00:00Z", e.annotationsOf("pg-cm")[AnnotationReconcileHandled])
	})

	t.Run("a request made during a reconcile runs next", func(t *testing.T) {
		e := newTestEnv(t, newObject(KindSource, "pg-cm", manifestData(sourceConfigJSON)))
		e.reconcileSource("pg-cm")
		request(e, "1")
		res := e.load("pg-cm")
		request(e, "2") // the user annotates again while the reconcile of "1" runs
		source := &Source{ETL: e.etl, Sink: e.status, Secrets: e.secrets}
		e.checkResult(source.Reconcile(context.Background(), res))
		assert.Equal(t, "1", e.annotationsOf("pg-cm")[AnnotationReconcileHandled], "the status answers the request it was read with")

		e.resetRecordings()
		e.reconcileSource("pg-cm")
		assert.Equal(t, 1, e.etl.callCount("GetSourceByName"))
		assert.Equal(t, "2", e.annotationsOf("pg-cm")[AnnotationReconcileHandled])
	})
}
