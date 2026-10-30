package gitops

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
)

const sourceConfigJSON = `{"name":"pg","type":"postgres","version":"v1.0.0","config":{"host":"pg","port":5432}}`

func TestSourceReconcileCreate(t *testing.T) {
	e := newTestEnv(t, managedConfigMap(KindSource, "pg-cm", manifestData(sourceConfigJSON)))
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
		e := newTestEnv(t, managedConfigMap(KindSource, "pg-cm", manifestData(`{"name":"pg","type":"postgres","version":"v1.0.0","config":{"port":5432,"host":"pg"}}`)))
		src := e.etl.addSource("pg", "v1.0.0")
		e.reconcileSource("pg-cm")
		assert.Empty(t, e.etl.writeCalls())
		assert.Zero(t, e.etl.callCount("TestSourceConnection"))
		assert.Equal(t, idString(src.ID), e.annotationsOf("pg-cm")[AnnotationEntityID])
	})

	t.Run("changed config is tested then updated", func(t *testing.T) {
		e := newTestEnv(t, managedConfigMap(KindSource, "pg-cm", manifestData(`{"name":"pg","type":"postgres","version":"v1.1.0","config":{"host":"pg","port":5432}}`)))
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
		e := newTestEnv(t, managedConfigMap(KindSource, "pg-cm", manifestData(`{"name":"pg","type":"postgres","version":"v1.1.0","config":{"host":"pg","port":5432}}`)))
		src := e.etl.addSource("pg", "v1.0.0")
		e.etl.failingMethods["UpdateSource"] = errors.New("version not found")
		e.reconcileSource("pg-cm")
		assert.Equal(t, PhaseFailed, e.phaseOf("pg-cm"))
		assert.Contains(t, e.messageOf("pg-cm"), "version not found")
		assert.Equal(t, idString(src.ID), e.annotationsOf("pg-cm")[AnnotationEntityID], "the object manages the found source, even though the update failed")
		assert.Equal(t, []string{indicatorSpawn}, e.status.indicatorActionsFor("pg-cm"))
	})

	t.Run("failed test keeps the stored source", func(t *testing.T) {
		e := newTestEnv(t, managedConfigMap(KindSource, "pg-cm", manifestData(`{"name":"pg","type":"postgres","version":"v1.0.0","config":{"host":"pg","password":"wrong"}}`)))
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
			e := newTestEnv(t, managedConfigMap(KindSource, "pg-cm", tt.data))
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
	e := newTestEnv(t, managedConfigMap(KindSource, "pg-cm", manifestData(`{"name":"pg","type":"nope","version":"v1.0.0","config":{}}`)))
	e.reconcileSource("pg-cm")
	require.Equal(t, PhaseFailed, e.phaseOf("pg-cm"))

	e.resetRecordings()
	e.reconcileSource("pg-cm")
	assert.Empty(t, e.status.statusUpdates, "a failure for unchanged data is not retried")

	e.setConfigMapData("pg-cm", manifestData(sourceConfigJSON))
	e.reconcileSource("pg-cm")
	assert.Equal(t, PhaseReady, e.phaseOf("pg-cm"))
	assert.Equal(t, []string{indicatorDelete}, e.status.indicatorActionsFor("pg-cm"), "fixing the data removes the indicator")
}

func TestSourceReconcileCreatedButReloadFailed(t *testing.T) {
	e := newTestEnv(t, managedConfigMap(KindSource, "pg-cm", manifestData(sourceConfigJSON)))
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
	e := newTestEnv(t, managedConfigMap(KindSource, "pg-cm", manifestData(sourceConfigJSON)))
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

func TestSourceReconcileFromSecret(t *testing.T) {
	e := newTestEnv(t, managedSecret(KindSource, "pg-secret", manifestData(sourceConfigJSON)))
	e.reconcileSource("pg-secret")
	assert.Equal(t, []string{"CreateSource"}, e.etl.writeCalls())

	var secret corev1.Secret
	require.NoError(t, e.client.Get(context.Background(), client.ObjectKey{Namespace: testNamespace, Name: "pg-secret"}, &secret))
	assert.Equal(t, PhaseReady, secret.Annotations[AnnotationPhase], "status is written on the Secret itself")
	assert.Equal(t, sourceConfigJSON, string(secret.Data[DataConfig]), "status writes never touch the data")

	// the observed hash on the Secret matches its decoded data, so the same data is skipped
	e.resetRecordings()
	e.reconcileSource("pg-secret")
	assert.Empty(t, e.etl.calls)
}

func TestSourceReconcileDeleted(t *testing.T) {
	e := newTestEnv(t)
	e.reconcileSource("gone")
	assert.Empty(t, e.etl.calls)
	assert.Empty(t, e.status.statusUpdates)
}
