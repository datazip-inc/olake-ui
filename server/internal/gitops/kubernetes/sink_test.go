package kubernetes

import (
	"context"
	"testing"

	"github.com/datazip-inc/olake-ui/server/internal/gitops/reconciler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestSinkSetPhase(t *testing.T) {
	ctx := context.Background()
	key := client.ObjectKey{Namespace: testNamespace, Name: "pg-cm"}
	data := manifestData(sourceConfigJSON)

	cm := managedConfigMap(reconciler.KindSource, "pg-cm", data)
	cm.Annotations = map[string]string{"argocd.argoproj.io/tracking-id": "app:/ConfigMap:olake/pg-cm"}
	k8sClient := newFakeClient(cm)
	recorder := record.NewFakeRecorder(10)
	sink := NewSink(k8sClient, recorder, nil)
	res := resourceFromCM(cm)

	require.NoError(t, sink.SetPhase(ctx, &res, &reconciler.Status{Phase: reconciler.PhaseReady, EntityID: "12", ObservedHash: "h1"}))
	var got corev1.ConfigMap
	require.NoError(t, k8sClient.Get(ctx, key, &got))
	assert.Equal(t, map[string]string{
		"argocd.argoproj.io/tracking-id":      "app:/ConfigMap:olake/pg-cm",
		reconciler.AnnotationPhase:            reconciler.PhaseReady,
		reconciler.AnnotationMessage:          "",
		reconciler.AnnotationEntityID:         "12",
		reconciler.AnnotationObservedHash:     "h1",
		reconciler.AnnotationReconcileHandled: "",
	}, got.Annotations, "other annotations are kept")
	assert.Equal(t, data, got.Data)
	assert.Equal(t, "Normal Synced resource synced to OLake", <-recorder.Events)

	// an identical status is not patched again, so it does not trigger another reconcile
	version := got.ResourceVersion
	require.NoError(t, sink.SetPhase(ctx, &res, &reconciler.Status{Phase: reconciler.PhaseReady, EntityID: "12", ObservedHash: "h1"}))
	require.NoError(t, k8sClient.Get(ctx, key, &got))
	assert.Equal(t, version, got.ResourceVersion)
	<-recorder.Events

	require.NoError(t, sink.SetPhase(ctx, &res, &reconciler.Status{Phase: reconciler.PhaseFailed, Message: "bad password", EntityID: "12", ObservedHash: "h2"}))
	require.NoError(t, k8sClient.Get(ctx, key, &got))
	assert.Equal(t, "h2", got.Annotations[reconciler.AnnotationObservedHash])
	assert.Equal(t, "bad password", got.Annotations[reconciler.AnnotationMessage])
	assert.Equal(t, "Warning SyncFailed bad password", <-recorder.Events)

	require.NoError(t, sink.SetPhase(ctx, &res, &reconciler.Status{Phase: reconciler.PhasePending, Message: "waiting for source", EntityID: "12", ObservedHash: "h2"}))
	assert.Equal(t, "Normal Pending waiting for source", <-recorder.Events)
}

func TestSinkSetPhaseReconcileHandled(t *testing.T) {
	ctx := context.Background()
	key := client.ObjectKey{Namespace: testNamespace, Name: "pg-cm"}
	cm := managedConfigMap(reconciler.KindSource, "pg-cm", manifestData(sourceConfigJSON))
	cm.Annotations = map[string]string{reconciler.AnnotationReconcileRequest: "2"}
	k8sClient := newFakeClient(cm)
	sink := NewSink(k8sClient, record.NewFakeRecorder(10), nil)
	res := resourceFromCM(cm)
	get := func() *corev1.ConfigMap {
		var got corev1.ConfigMap
		require.NoError(t, k8sClient.Get(ctx, key, &got))
		return &got
	}

	ready := reconciler.Status{Phase: reconciler.PhaseReady, EntityID: "12", ObservedHash: "h1"}
	require.NoError(t, sink.SetPhase(ctx, &res, &ready))

	// answering a request is a change on its own, even when the rest of the status is the same
	answered := ready
	answered.ReconcileHandled = "2"
	require.NoError(t, sink.SetPhase(ctx, &res, &answered))
	got := get()
	assert.Equal(t, "2", got.Annotations[reconciler.AnnotationReconcileHandled])
	assert.Equal(t, "2", got.Annotations[reconciler.AnnotationReconcileRequest], "the user's request is left as it is")

	version := got.ResourceVersion
	require.NoError(t, sink.SetPhase(ctx, &res, &answered))
	assert.Equal(t, version, get().ResourceVersion, "an answered request is not patched again")
}
