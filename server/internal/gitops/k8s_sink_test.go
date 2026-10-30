package gitops

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestK8sSinkSetPhase(t *testing.T) {
	ctx := context.Background()
	key := client.ObjectKey{Namespace: testNamespace, Name: "pg-cm"}
	data := manifestData(sourceConfigJSON)

	cm := managedConfigMap(KindSource, "pg-cm", data)
	cm.Annotations = map[string]string{"argocd.argoproj.io/tracking-id": "app:/ConfigMap:olake/pg-cm"}
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cm).Build()
	recorder := record.NewFakeRecorder(10)
	sink := NewK8sSink(k8sClient, recorder, nil)
	res := resourceFromCM(cm)

	require.NoError(t, sink.SetPhase(ctx, &res, PhaseReady, "", "12", ""))
	var got corev1.ConfigMap
	require.NoError(t, k8sClient.Get(ctx, key, &got))
	assert.Equal(t, map[string]string{
		"argocd.argoproj.io/tracking-id": "app:/ConfigMap:olake/pg-cm",
		AnnotationPhase:                  PhaseReady,
		AnnotationMessage:                "",
		AnnotationEntityID:               "12",
		AnnotationObservedHash:           ContentHash(data),
	}, got.Annotations, "an empty observed hash defaults to the data hash; other annotations are kept")
	assert.Equal(t, data, got.Data)
	assert.Equal(t, "Normal Synced resource synced to OLake", <-recorder.Events)

	// an identical status is not patched again, so it does not trigger another reconcile
	version := got.ResourceVersion
	require.NoError(t, sink.SetPhase(ctx, &res, PhaseReady, "", "12", ""))
	require.NoError(t, k8sClient.Get(ctx, key, &got))
	assert.Equal(t, version, got.ResourceVersion)
	<-recorder.Events

	require.NoError(t, sink.SetPhase(ctx, &res, PhaseFailed, "bad password", "12", "custom-hash"))
	require.NoError(t, k8sClient.Get(ctx, key, &got))
	assert.Equal(t, "custom-hash", got.Annotations[AnnotationObservedHash])
	assert.Equal(t, "bad password", got.Annotations[AnnotationMessage])
	assert.Equal(t, "Warning SyncFailed bad password", <-recorder.Events)

	require.NoError(t, sink.SetPhase(ctx, &res, PhasePending, "waiting for source", "12", ""))
	assert.Equal(t, "Normal Pending waiting for source", <-recorder.Events)
}
