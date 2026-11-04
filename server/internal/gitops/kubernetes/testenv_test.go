package kubernetes

// Shared setup for the Kubernetes mode tests: a fake cluster (controller-runtime's in-memory client)
// and builders for labelled ConfigMaps and Secrets. The reconcile rules themselves are tested in the
// reconciler package; these tests cover loading, watching and status writing.

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/datazip-inc/olake-ui/server/internal/gitops/reconciler"
)

const (
	testNamespace = "olake"
	testProjectID = "123"
	testUserID    = "1"

	sourceConfigJSON = `{"name":"pg","type":"postgres","version":"v1.0.0","config":{"host":"pg","port":5432}}`
)

func TestMain(m *testing.M) {
	// the controllers log through controller-runtime; without a logger it warns after 30s
	ctrllog.SetLogger(logr.Discard())
	os.Exit(m.Run())
}

// newFakeClient is a fake cluster holding objs.
func newFakeClient(objs ...client.Object) client.Client {
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
}

// reconcileOnce runs controller.Reconcile for the named object in testNamespace and fails the test
// on an error.
func reconcileOnce(t *testing.T, controller interface {
	Reconcile(context.Context, ctrl.Request) (ctrl.Result, error)
}, name string) ctrl.Result {
	t.Helper()
	result, err := controller.Reconcile(context.Background(), requestFor(name))
	require.NoError(t, err)
	return result
}

// requestFor is the reconcile request Kubernetes sends for the named object in testNamespace.
func requestFor(name string) ctrl.Request {
	return ctrl.Request{NamespacedName: client.ObjectKey{Namespace: testNamespace, Name: name}}
}

// managedConfigMap and managedSecret build an object labelled for GitOps (olake.io/managed and
// olake.io/kind), so the controllers pick it up.
func managedConfigMap(kind, name string, data map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: testNamespace, Labels: managedLabels(kind)},
		Data:       data,
	}
}

func managedSecret(kind, name string, data map[string]string) *corev1.Secret {
	bytes := map[string][]byte{}
	for k, v := range data {
		bytes[k] = []byte(v)
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: testNamespace, Labels: managedLabels(kind)},
		Data:       bytes,
	}
}

// manifestData is the data of a Source, Destination or Job object: the test project and user, and
// config as data.config.
func manifestData(config string) map[string]string {
	return map[string]string{reconciler.DataProjectID: testProjectID, reconciler.DataUserID: testUserID, reconciler.DataConfig: config}
}

// streamsManifestData is the data of a Streams object: jobRef names the Job object it belongs to.
func streamsManifestData(jobRef, config string) map[string]string {
	return map[string]string{reconciler.DataProjectID: testProjectID, reconciler.DataJob: jobRef, reconciler.DataConfig: config}
}

// jobConfigJSON is job "orders" from source "pg" to destination "lake"; fields override it.
func jobConfigJSON(fields map[string]interface{}) string {
	cfg := map[string]interface{}{"name": "orders", "source": "pg", "destination": "lake", "frequency": "0 * * * *"}
	for k, v := range fields {
		cfg[k] = v
	}
	b, _ := json.Marshal(cfg)
	return string(b)
}
