package gitops

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestSetupRequiresNamespace(t *testing.T) {
	_, err := Setup(nil, &rest.Config{}, "")
	assert.ErrorContains(t, err, "POD_NAMESPACE is required")
}

func TestKindPredicate(t *testing.T) {
	unmanaged := managedConfigMap(KindSource, "pg", nil)
	delete(unmanaged.Labels, LabelManaged)
	disabled := managedConfigMap(KindSource, "pg", nil)
	disabled.Labels[LabelManaged] = "false"

	tests := []struct {
		name string
		obj  client.Object
		want bool
	}{
		{name: "managed source ConfigMap", obj: managedConfigMap(KindSource, "pg", nil), want: true},
		{name: "managed source Secret", obj: managedSecret(KindSource, "pg", nil), want: true},
		{name: "managed object of another kind", obj: managedConfigMap(KindJob, "orders", nil), want: false},
		{name: "kind label without the managed label", obj: unmanaged, want: false},
		{name: "managed label not true", obj: disabled, want: false},
	}
	managedSource := kindPredicate(KindSource)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, managedSource.Create(event.CreateEvent{Object: tt.obj}))
		})
	}
}

// A Secret alone and a missing object are covered by TestSourceReconcileFromSecret and TestSourceReconcileDeleted.
func TestFetchManagedPrefersConfigMap(t *testing.T) {
	e := newTestEnv(t, managedSecret(KindSource, "pg", manifestData("secret")), managedConfigMap(KindSource, "pg", manifestData("cm")))
	res, err := fetchManaged(context.Background(), e.client, client.ObjectKey{Namespace: testNamespace, Name: "pg"})
	require.NoError(t, err)
	assert.Equal(t, ObjectKindConfigMap, res.ObjectKind)
	assert.Equal(t, "cm", res.Config())
}

func TestEnqueueDependentJobsForConnector(t *testing.T) {
	jobWith := func(name, source, destination string) *corev1.ConfigMap {
		return managedConfigMap(KindJob, name, manifestData(jobConfigJSON(map[string]interface{}{"source": source, "destination": destination})))
	}
	otherNamespace := jobWith("other-namespace", "pg", "lake")
	otherNamespace.Namespace = "other"
	e := newTestEnv(t,
		jobWith("by-name", "pg", "lake"),
		jobWith("by-id", "12", "13"),
		jobWith("by-object-name", "pg-cm", "lake-cm"),
		jobWith("other", "mysql", "s3"),
		otherNamespace,
		managedConfigMap(KindJob, "invalid", manifestData(`{"name":"broken"}`)),
	)
	dependentJobs := e.jobController().enqueueDependentJobs
	ctx := context.Background()

	// a job names the OLake connector from the object's config, or its entity id; never the object name
	srcCM := managedConfigMap(KindSource, "pg-cm", manifestData(`{"name":"pg","type":"postgres","version":"v1.0.0","config":{}}`))
	srcCM.Annotations = map[string]string{AnnotationEntityID: "12"}
	assert.ElementsMatch(t, []reconcile.Request{requestFor("by-name"), requestFor("by-id")}, dependentJobs(ctx, srcCM))

	destCM := managedConfigMap(KindDestination, "lake-cm", manifestData(`{"name":"lake","type":"iceberg","version":"v1.0.0","config":{}}`))
	destCM.Annotations = map[string]string{AnnotationEntityID: "13"}
	assert.ElementsMatch(t, []reconcile.Request{requestFor("by-name"), requestFor("by-id")}, dependentJobs(ctx, destCM))

	// a Secret that has not synced yet has no entity id, so only name refs match
	srcSecret := managedSecret(KindSource, "pg-secret", manifestData(`{"name":"pg"}`))
	assert.ElementsMatch(t, []reconcile.Request{requestFor("by-name")}, dependentJobs(ctx, srcSecret))
}

func TestEnqueueDependentJobsForStreams(t *testing.T) {
	synced := managedConfigMap(KindJob, "synced", manifestData(jobConfigJSON(nil)))
	synced.Annotations = map[string]string{AnnotationEntityID: "42"}
	otherProject := managedConfigMap(KindJob, "other-project", map[string]string{DataProjectID: "999", DataConfig: jobConfigJSON(nil)})
	e := newTestEnv(t, managedConfigMap(KindJob, "orders-cm", manifestData(jobConfigJSON(nil))), synced, otherProject)
	dependentJobs := e.jobController().enqueueDependentJobs
	ctx := context.Background()

	assert.Equal(t, []reconcile.Request{requestFor("orders-cm")}, dependentJobs(ctx, managedConfigMap(KindStreams, "s", streamsManifestData("orders-cm", ""))))
	assert.Equal(t, []reconcile.Request{requestFor("synced")}, dependentJobs(ctx, managedConfigMap(KindStreams, "s", streamsManifestData("42", ""))))
	assert.Empty(t, dependentJobs(ctx, managedConfigMap(KindStreams, "s", streamsManifestData("missing", ""))))
	assert.Empty(t, dependentJobs(ctx, managedConfigMap(KindStreams, "s", streamsManifestData("other-project", ""))), "streams of another project")
	assert.Empty(t, dependentJobs(ctx, &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "s", Namespace: testNamespace}}), "not a managed type")
}
