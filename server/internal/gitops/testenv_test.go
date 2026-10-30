package gitops

// Shared setup for the reconcile tests.
//
// A reconcile test builds a testEnv, puts ConfigMaps or Secrets in it, runs the real controller's
// Reconcile, and then checks the outcome: the status annotations on the objects, the rows in the
// fake OLake, and what the reconcilers reported. Three stand-ins make that possible without a
// cluster, a database or Temporal:
//   - a fake cluster: controller-runtime's in-memory client, holding the ConfigMaps and Secrets;
//   - fakeETL (fake_etl_test.go): an in-memory OLake ETL service;
//   - statusRecorder (below): remembers the status and indicator actions the reconcilers report.

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	testNamespace = "olake"
	testProjectID = "123"
	testUserID    = "1"

	userSelectedStreams = `{"selected_streams":{"public":[{"stream_name":"users","partition_regex":"","normalization":true}]}}`
)

func TestMain(m *testing.M) {
	// the reconcilers log through controller-runtime; without a logger it warns after 30s
	ctrllog.SetLogger(logr.Discard())
	os.Exit(m.Run())
}

// testEnv is everything one test runs against: a fake cluster holding the ConfigMaps and Secrets,
// a fake OLake (the ETL service), and a recorder of the status the reconcilers report.
type testEnv struct {
	t      *testing.T
	client client.Client   // the fake cluster
	etl    *fakeETL        // the fake ETL service
	status *statusRecorder // what the reconcilers reported
}

// newTestEnv starts a test with the given ConfigMaps and Secrets already in the fake cluster.
func newTestEnv(t *testing.T, objs ...client.Object) *testEnv {
	t.Helper()
	// retryMap is package state: a retry window left by one test must not leak into the next
	t.Cleanup(func() { retryMap.Clear() })
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
	return &testEnv{
		t:      t,
		client: k8sClient,
		etl:    newFakeETL(),
		status: &statusRecorder{realSink: NewK8sSink(k8sClient, record.NewFakeRecorder(100), nil)},
	}
}

// resetRecordings forgets the recorded ETL calls, status updates and indicator actions, so the
// next assertions see only what the next reconcile does. Stored data (objects, OLake rows) stays.
func (e *testEnv) resetRecordings() {
	e.etl.calls = nil
	e.status.statusUpdates, e.status.indicatorActions = nil, nil
}

// reconcileSource, reconcileDestination and reconcileJob run the real controller once for the named
// object, as Kubernetes would after a change, and return its result (for example a requeue).
func (e *testEnv) reconcileSource(name string) ctrl.Result {
	source := &SourceReconciler{ETL: e.etl, Sink: e.status}
	return e.runReconcile(&resourceController{client: e.client, reconcile: source.reconcile}, name)
}

func (e *testEnv) reconcileDestination(name string) ctrl.Result {
	destination := &DestinationReconciler{ETL: e.etl, Sink: e.status}
	return e.runReconcile(&resourceController{client: e.client, reconcile: destination.reconcile}, name)
}

func (e *testEnv) reconcileJob(name string) ctrl.Result {
	return e.runReconcile(e.jobController(), name)
}

// jobController is the real job controller wired to this test's fakes.
func (e *testEnv) jobController() *jobController {
	return &jobController{client: e.client, job: &JobReconciler{ETL: e.etl, Sink: e.status}}
}

// runReconcile calls controller.Reconcile for the named object and fails the test if it returns an
// error: the reconcilers report failures as status, so an error here is a bug in the test or code.
func (e *testEnv) runReconcile(controller reconcile.Reconciler, name string) ctrl.Result {
	e.t.Helper()
	result, err := controller.Reconcile(context.Background(), requestFor(name))
	require.NoError(e.t, err)
	return result
}

// createObject adds a ConfigMap or Secret to the fake cluster, as kubectl create does.
func (e *testEnv) createObject(obj client.Object) {
	e.t.Helper()
	require.NoError(e.t, e.client.Create(context.Background(), obj))
}

// setConfigMapData replaces a ConfigMap's data and keeps its annotations, as kubectl apply does.
func (e *testEnv) setConfigMapData(name string, data map[string]string) {
	e.t.Helper()
	cm := e.getConfigMap(name)
	cm.Data = data
	require.NoError(e.t, e.client.Update(context.Background(), cm))
}

// clearConfigMapAnnotations drops a ConfigMap's annotations, as kubectl replace or a forced sync does.
func (e *testEnv) clearConfigMapAnnotations(name string) {
	e.t.Helper()
	cm := e.getConfigMap(name)
	cm.Annotations = nil
	require.NoError(e.t, e.client.Update(context.Background(), cm))
}

// getConfigMap reads the named ConfigMap from the fake cluster.
func (e *testEnv) getConfigMap(name string) *corev1.ConfigMap {
	e.t.Helper()
	var cm corev1.ConfigMap
	require.NoError(e.t, e.client.Get(context.Background(), client.ObjectKey{Namespace: testNamespace, Name: name}, &cm))
	return &cm
}

// annotationsOf, phaseOf and messageOf read the status a reconcile wrote on the named ConfigMap.
func (e *testEnv) annotationsOf(name string) map[string]string {
	e.t.Helper()
	return e.getConfigMap(name).Annotations
}

func (e *testEnv) phaseOf(name string) string {
	e.t.Helper()
	return e.annotationsOf(name)[AnnotationPhase]
}

func (e *testEnv) messageOf(name string) string {
	e.t.Helper()
	return e.annotationsOf(name)[AnnotationMessage]
}

// statusRecorder stands in for the reconcilers' StatusSink, the object they report results through.
// A reconcile reports two things:
//   - its status (phase, message, entity id), which ends up as olake.io/* annotations on the
//     ConfigMap or Secret;
//   - a failure indicator action: "spawn" a pod showing the error when a sync fails, "delete" it
//     when the sync succeeds again.
//
// statusRecorder remembers every status update and indicator action, so a test can assert on them.
// Status updates are also passed on to the real K8sSink, so the annotations are really written and
// the next reconcile in the same test sees them. Indicator actions are only remembered: starting
// the indicator workflow needs Temporal, and the indicator itself is tested in
// indicator_client_test.go.
type statusRecorder struct {
	// realSink writes status updates to the ConfigMap or Secret in the fake cluster.
	realSink *K8sSink

	// failStatusUpdates, when set, makes every status update fail with it, to test a reconcile
	// whose status write is rejected (for example a conflict with another writer).
	failStatusUpdates error

	// statusUpdates and indicatorActions are everything the reconcilers reported, in order.
	// testEnv.resetRecordings clears both.
	statusUpdates    []statusUpdate
	indicatorActions []indicatorAction
}

// statusUpdate is one status a reconcile wrote for an object.
type statusUpdate struct {
	objectName string // the ConfigMap or Secret name
	phase      string // PhasePending, PhaseReady or PhaseFailed
	message    string
	entityID   string // the OLake entity id, "" when there is none
}

// indicatorAction is one failure indicator request for an object.
type indicatorAction struct {
	objectName string
	action     string // indicatorSpawn or indicatorDelete
}

// SetPhase remembers the status update, then writes it through the real sink.
func (s *statusRecorder) SetPhase(ctx context.Context, r *ResourceData, phase, message, entityID, observedHash string) error {
	s.statusUpdates = append(s.statusUpdates, statusUpdate{objectName: r.Name, phase: phase, message: message, entityID: entityID})
	if s.failStatusUpdates != nil {
		return s.failStatusUpdates
	}
	return s.realSink.SetPhase(ctx, r, phase, message, entityID, observedHash)
}

// Indicator remembers the indicator action.
func (s *statusRecorder) Indicator(_ context.Context, r *ResourceData, action, _ string) {
	s.indicatorActions = append(s.indicatorActions, indicatorAction{objectName: r.Name, action: action})
}

// lastStatusUpdate is the last status written for the named object; the test fails if there is none.
func (s *statusRecorder) lastStatusUpdate(t *testing.T, objectName string) statusUpdate {
	t.Helper()
	for i := len(s.statusUpdates) - 1; i >= 0; i-- {
		if s.statusUpdates[i].objectName == objectName {
			return s.statusUpdates[i]
		}
	}
	t.Fatalf("no status written for %s", objectName)
	return statusUpdate{}
}

// indicatorActionsFor lists the indicator actions for the named object, in order, for example
// []string{"spawn"} after a failure or []string{"delete"} after a success.
func (s *statusRecorder) indicatorActionsFor(objectName string) []string {
	var actions []string
	for _, a := range s.indicatorActions {
		if a.objectName == objectName {
			actions = append(actions, a.action)
		}
	}
	return actions
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
	return map[string]string{DataProjectID: testProjectID, DataUserID: testUserID, DataConfig: config}
}

// streamsManifestData is the data of a Streams object: jobRef names the Job object it belongs to,
// and config is the stream selection.
func streamsManifestData(jobRef, config string) map[string]string {
	return map[string]string{DataProjectID: testProjectID, DataUserID: testUserID, DataJob: jobRef, DataConfig: config}
}

// jobConfigJSON is job "orders" from source "pg" to destination "lake"; fields override it, and a nil
// value removes the field.
func jobConfigJSON(fields map[string]interface{}) string {
	cfg := map[string]interface{}{"name": "orders", "source": "pg", "destination": "lake", "frequency": "0 * * * *"}
	for k, v := range fields {
		if v == nil {
			delete(cfg, k)
			continue
		}
		cfg[k] = v
	}
	b, _ := json.Marshal(cfg)
	return string(b)
}

// idString formats an entity id the way the olake.io/entity-id annotation stores it.
func idString(n int) string { return strconv.Itoa(n) }
