package kubernetes

import (
	"context"
	"testing"

	"github.com/datazip-inc/olake-ui/server/internal/gitops/reconciler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestSetupRequiresNamespace(t *testing.T) {
	_, err := Setup(nil, &rest.Config{}, "")
	assert.ErrorContains(t, err, "POD_NAMESPACE is required")
}

func TestKindPredicate(t *testing.T) {
	unmanaged := managedConfigMap(reconciler.KindSource, "pg", nil)
	delete(unmanaged.Labels, LabelManaged)
	disabled := managedConfigMap(reconciler.KindSource, "pg", nil)
	disabled.Labels[LabelManaged] = "false"

	tests := []struct {
		name string
		obj  client.Object
		want bool
	}{
		{name: "managed source ConfigMap", obj: managedConfigMap(reconciler.KindSource, "pg", nil), want: true},
		{name: "managed source Secret", obj: managedSecret(reconciler.KindSource, "pg", nil), want: true},
		{name: "managed object of another kind", obj: managedConfigMap(reconciler.KindJob, "orders", nil), want: false},
		{name: "kind label without the managed label", obj: unmanaged, want: false},
		{name: "managed label not true", obj: disabled, want: false},
	}
	managedSource := kindPredicate(reconciler.KindSource)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, managedSource.Create(event.CreateEvent{Object: tt.obj}))
		})
	}
}

func TestFetchManagedPrefersConfigMap(t *testing.T) {
	k8sClient := newFakeClient(managedSecret(reconciler.KindSource, "pg", manifestData("secret")), managedConfigMap(reconciler.KindSource, "pg", manifestData("cm")))
	res, err := fetchManaged(context.Background(), k8sClient, client.ObjectKey{Namespace: testNamespace, Name: "pg"})
	require.NoError(t, err)
	assert.Equal(t, ObjectKindConfigMap, res.ObjectKind)
	assert.Equal(t, "cm", res.Config())
}

func TestEnqueueDependentJobsForConnector(t *testing.T) {
	jobWith := func(name, source, destination string) *corev1.ConfigMap {
		return managedConfigMap(reconciler.KindJob, name, manifestData(jobConfigJSON(map[string]interface{}{"source": source, "destination": destination})))
	}
	otherNamespace := jobWith("other-namespace", "pg", "lake")
	otherNamespace.Namespace = "other"
	k8sClient := newFakeClient(
		jobWith("by-name", "pg", "lake"),
		jobWith("by-id", "12", "13"),
		jobWith("by-object-name", "pg-cm", "lake-cm"),
		jobWith("other", "mysql", "s3"),
		otherNamespace,
		managedConfigMap(reconciler.KindJob, "invalid", manifestData(`{"name":"broken"}`)),
	)
	dependentJobs := (&jobController{client: k8sClient}).enqueueDependentJobs
	ctx := context.Background()

	// a job names the OLake connector from the object's config, or its entity id; never the object name
	srcCM := managedConfigMap(reconciler.KindSource, "pg-cm", manifestData(`{"name":"pg","type":"postgres","version":"v1.0.0","config":{}}`))
	srcCM.Annotations = map[string]string{reconciler.AnnotationEntityID: "12"}
	assert.ElementsMatch(t, []reconcile.Request{requestFor("by-name"), requestFor("by-id")}, dependentJobs(ctx, srcCM))

	destCM := managedConfigMap(reconciler.KindDestination, "lake-cm", manifestData(`{"name":"lake","type":"iceberg","version":"v1.0.0","config":{}}`))
	destCM.Annotations = map[string]string{reconciler.AnnotationEntityID: "13"}
	assert.ElementsMatch(t, []reconcile.Request{requestFor("by-name"), requestFor("by-id")}, dependentJobs(ctx, destCM))

	// a Secret that has not synced yet has no entity id, so only name refs match
	srcSecret := managedSecret(reconciler.KindSource, "pg-secret", manifestData(`{"name":"pg"}`))
	assert.ElementsMatch(t, []reconcile.Request{requestFor("by-name")}, dependentJobs(ctx, srcSecret))
}

func TestEnqueueDependentJobsForStreams(t *testing.T) {
	synced := managedConfigMap(reconciler.KindJob, "synced", manifestData(jobConfigJSON(nil)))
	synced.Annotations = map[string]string{reconciler.AnnotationEntityID: "42"}
	otherProject := managedConfigMap(reconciler.KindJob, "other-project", map[string]string{reconciler.DataProjectID: "999", reconciler.DataConfig: jobConfigJSON(nil)})
	k8sClient := newFakeClient(managedConfigMap(reconciler.KindJob, "orders-cm", manifestData(jobConfigJSON(nil))), synced, otherProject)
	dependentJobs := (&jobController{client: k8sClient}).enqueueDependentJobs
	ctx := context.Background()

	assert.Equal(t, []reconcile.Request{requestFor("orders-cm")}, dependentJobs(ctx, managedConfigMap(reconciler.KindStreams, "s", streamsManifestData("orders-cm", ""))))
	assert.Equal(t, []reconcile.Request{requestFor("synced")}, dependentJobs(ctx, managedConfigMap(reconciler.KindStreams, "s", streamsManifestData("42", ""))))
	assert.Empty(t, dependentJobs(ctx, managedConfigMap(reconciler.KindStreams, "s", streamsManifestData("missing", ""))))
	assert.Empty(t, dependentJobs(ctx, managedConfigMap(reconciler.KindStreams, "s", streamsManifestData("other-project", ""))), "streams of another project")
	assert.Empty(t, dependentJobs(ctx, &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "s", Namespace: testNamespace}}), "not a managed type")
}

func TestControllersIgnoreDeletedObjects(t *testing.T) {
	// a delete event reconciles an object that is gone: nothing to load, so the reconciler never
	// runs and the OLake entity is kept
	k8sClient := newFakeClient()
	source := &resourceController{client: k8sClient, reconcile: func(context.Context, *reconciler.ResourceData) (ctrl.Result, error) {
		t.Fatal("reconcile called for a deleted object")
		return ctrl.Result{}, nil
	}}
	assert.Zero(t, reconcileOnce(t, source, "gone"))
	job := &jobController{client: k8sClient, reconcile: func(context.Context, *reconciler.ResourceData, *reconciler.ResourceData) (ctrl.Result, error) {
		t.Fatal("reconcile called for a deleted object")
		return ctrl.Result{}, nil
	}}
	assert.Zero(t, reconcileOnce(t, job, "gone"))
}

func TestSecretBackedSource(t *testing.T) {
	ctx := context.Background()
	k8sClient := newFakeClient(managedSecret(reconciler.KindSource, "pg-secret", manifestData(sourceConfigJSON)))
	sink := NewSink(k8sClient, record.NewFakeRecorder(10), nil)
	var loaded []reconciler.ResourceData
	controller := &resourceController{client: k8sClient, reconcile: func(ctx context.Context, res *reconciler.ResourceData) (ctrl.Result, error) {
		loaded = append(loaded, *res)
		// what a reconciler writes on success: the hash of the data it read
		return ctrl.Result{}, sink.SetPhase(ctx, res, reconciler.PhaseReady, "", "7", reconciler.ContentHash(res.Data))
	}}

	reconcileOnce(t, controller, "pg-secret")
	require.Len(t, loaded, 1)
	assert.Equal(t, ObjectKindSecret, loaded[0].ObjectKind)
	assert.Equal(t, sourceConfigJSON, loaded[0].Config(), "the reconciler reads the decoded Secret data")

	var secret corev1.Secret
	require.NoError(t, k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "pg-secret"}, &secret))
	assert.Equal(t, reconciler.PhaseReady, secret.Annotations[reconciler.AnnotationPhase], "status is written on the Secret itself")
	assert.Equal(t, sourceConfigJSON, string(secret.Data[reconciler.DataConfig]), "status writes never touch the data")

	// the next load carries an observed hash that matches its decoded data, so the reconciler skips it
	reconcileOnce(t, controller, "pg-secret")
	require.Len(t, loaded, 2)
	assert.Equal(t, reconciler.ContentHash(loaded[1].Data), loaded[1].Annotations[reconciler.AnnotationObservedHash])
}

func TestJobControllerFindsStreams(t *testing.T) {
	selection := `{"selected_streams":{"public":[{"stream_name":"users"}]}}`
	otherNamespace := managedConfigMap(reconciler.KindStreams, "orders-streams-elsewhere", streamsManifestData("orders-cm", selection))
	otherNamespace.Namespace = "other"
	k8sClient := newFakeClient(
		managedConfigMap(reconciler.KindJob, "orders-cm", manifestData(jobConfigJSON(nil))),
		managedConfigMap(reconciler.KindJob, "lonely-cm", manifestData(jobConfigJSON(map[string]interface{}{"name": "lonely"}))),
		managedConfigMap(reconciler.KindStreams, "orders-streams", streamsManifestData("orders-cm", selection)),
		managedConfigMap(reconciler.KindStreams, "billing-streams", streamsManifestData("billing-cm", selection)),
		otherNamespace,
	)
	var job, streams *reconciler.ResourceData
	controller := &jobController{client: k8sClient, reconcile: func(_ context.Context, j, s *reconciler.ResourceData) (ctrl.Result, error) {
		job, streams = j, s
		return ctrl.Result{}, nil
	}}

	reconcileOnce(t, controller, "orders-cm")
	assert.Equal(t, "orders-cm", job.Name)
	require.NotNil(t, streams)
	assert.Equal(t, "orders-streams", streams.Name, "the Streams ConfigMap in the job's namespace that names the job")
	assert.Equal(t, selection, streams.Config())

	reconcileOnce(t, controller, "lonely-cm")
	assert.Nil(t, streams, "a job without a Streams ConfigMap gets none")
}
