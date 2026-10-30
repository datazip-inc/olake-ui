package reconciler

// Shared setup for the reconcile tests.
//
// A reconcile test builds a testEnv, puts manifest objects in it, runs a reconciler on one, and
// then checks the outcome: the status annotations on the objects, the rows in the fake OLake, and
// what the reconcilers reported. Three stand-ins make that possible without a cluster, a database
// or Temporal:
//   - objects: the stored manifests, as a GitOps mode holds them (ConfigMaps and Secrets in
//     Kubernetes mode). How a mode loads and watches them is tested in that mode's package;
//   - fakeETL (fake_etl_test.go): an in-memory OLake ETL service;
//   - statusRecorder (below): the StatusSink. It remembers the status and indicator actions the
//     reconcilers report, and writes the status onto the stored object.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/gitops/secretref"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
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

// testEnv is everything one test runs against: the stored manifest objects, a fake OLake (the ETL
// service), and a recorder of the status the reconcilers report.
type testEnv struct {
	t       *testing.T
	objects map[string]*ResourceData // objectKey -> the stored object
	etl     *fakeETL                 // the fake ETL service
	status  *statusRecorder          // what the reconcilers reported
	secrets *SecretRefs              // no secret managers configured
}

// newTestEnv starts a test with the given objects already stored.
func newTestEnv(t *testing.T, objs ...*ResourceData) *testEnv {
	t.Helper()
	// retryMap is package state: a retry window left by one test must not leak into the next
	t.Cleanup(func() { retryMap.Clear() })
	e := &testEnv{t: t, objects: map[string]*ResourceData{}, etl: newFakeETL(), secrets: NewSecretRefs(&appconfig.Config{}, secretref.K8sConfig{})}
	e.status = &statusRecorder{objects: e.objects}
	for _, obj := range objs {
		e.createObject(obj)
	}
	return e
}

func objectKey(namespace, name string) string { return namespace + "/" + name }

// resetRecordings forgets the recorded ETL calls, status updates and indicator actions, so the
// next assertions see only what the next reconcile does. Stored data (objects, OLake rows) stays.
func (e *testEnv) resetRecordings() {
	e.etl.calls = nil
	e.status.statusUpdates, e.status.indicatorActions = nil, nil
}

// reconcileSource, reconcileDestination and reconcileJob run the reconciler once for the named
// object, as a GitOps mode does after a change, and return its result (for example a requeue).
func (e *testEnv) reconcileSource(name string) ctrl.Result {
	e.t.Helper()
	source := &Source{ETL: e.etl, Sink: e.status, Secrets: e.secrets}
	return e.checkResult(source.Reconcile(context.Background(), e.load(name)))
}

func (e *testEnv) reconcileDestination(name string) ctrl.Result {
	e.t.Helper()
	destination := &Destination{ETL: e.etl, Sink: e.status, Secrets: e.secrets}
	return e.checkResult(destination.Reconcile(context.Background(), e.load(name)))
}

// reconcileJob hands the job its Streams object as a mode finds it: the stored object
// StreamsBelongTo matches, or none.
func (e *testEnv) reconcileJob(name string) ctrl.Result {
	e.t.Helper()
	job := e.load(name)
	var streams *ResourceData
	for _, obj := range e.objects {
		if candidate := cloneObject(obj); StreamsBelongTo(candidate, job) {
			streams = candidate
		}
	}
	jobReconciler := &Job{ETL: e.etl, Sink: e.status}
	return e.checkResult(jobReconciler.Reconcile(context.Background(), job, streams))
}

// checkResult fails the test if a reconcile returned an error: the reconcilers report failures as
// status, so an error here is a bug in the test or code.
func (e *testEnv) checkResult(result ctrl.Result, err error) ctrl.Result {
	e.t.Helper()
	require.NoError(e.t, err)
	return result
}

// load is a copy of the named object in testNamespace, as a mode reads it before a reconcile.
func (e *testEnv) load(name string) *ResourceData {
	e.t.Helper()
	obj, ok := e.objects[objectKey(testNamespace, name)]
	require.True(e.t, ok, "no object %s", name)
	return cloneObject(obj)
}

func cloneObject(obj *ResourceData) *ResourceData {
	out := *obj
	out.Data = map[string]string{}
	for k, v := range obj.Data {
		out.Data[k] = v
	}
	out.Annotations = map[string]string{}
	for k, v := range obj.Annotations {
		out.Annotations[k] = v
	}
	return &out
}

// createObject stores an object, as kubectl create does.
func (e *testEnv) createObject(obj *ResourceData) {
	e.t.Helper()
	key := objectKey(obj.Namespace, obj.Name)
	require.NotContains(e.t, e.objects, key)
	e.objects[key] = cloneObject(obj)
}

// deleteObject removes the named object.
func (e *testEnv) deleteObject(name string) {
	e.t.Helper()
	e.object(name)
	delete(e.objects, objectKey(testNamespace, name))
}

// setData replaces an object's data and keeps its annotations, as kubectl apply does.
func (e *testEnv) setData(name string, data map[string]string) {
	e.t.Helper()
	e.object(name).Data = data
}

// clearAnnotations drops an object's annotations, as kubectl replace or a forced sync does.
func (e *testEnv) clearAnnotations(name string) {
	e.t.Helper()
	e.object(name).Annotations = map[string]string{}
}

// object is the stored object with the given name in testNamespace.
func (e *testEnv) object(name string) *ResourceData {
	e.t.Helper()
	obj, ok := e.objects[objectKey(testNamespace, name)]
	require.True(e.t, ok, "no object %s", name)
	return obj
}

// annotationsOf, phaseOf and messageOf read the status a reconcile wrote on the named object.
func (e *testEnv) annotationsOf(name string) map[string]string {
	e.t.Helper()
	return e.object(name).Annotations
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
//     object;
//   - a failure indicator action: "spawn" a pod showing the error when a sync fails, "delete" it
//     when the sync succeeds again.
//
// statusRecorder remembers every status update and indicator action, so a test can assert on them.
// Status updates are also written onto the stored object, as Kubernetes mode's sink does, so the
// next reconcile in the same test sees them. Indicator actions are only remembered: starting the
// indicator workflow needs Temporal, and StartIndicator is tested in indicator_test.go.
type statusRecorder struct {
	// objects are the testEnv's stored objects, which status updates are written onto.
	objects map[string]*ResourceData

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
	objectName string
	phase      string // PhasePending, PhaseReady or PhaseFailed
	message    string
	entityID   string // the OLake entity id, "" when there is none
}

// indicatorAction is one failure indicator request for an object.
type indicatorAction struct {
	objectName string
	action     string // indicatorSpawn or indicatorDelete
}

// SetPhase remembers the status update, then writes it onto the stored object.
func (s *statusRecorder) SetPhase(_ context.Context, r *ResourceData, phase, message, entityID, observedHash string) error {
	s.statusUpdates = append(s.statusUpdates, statusUpdate{objectName: r.Name, phase: phase, message: message, entityID: entityID})
	if s.failStatusUpdates != nil {
		return s.failStatusUpdates
	}
	obj, ok := s.objects[objectKey(r.Namespace, r.Name)]
	if !ok {
		return fmt.Errorf("object %s/%s not found", r.Namespace, r.Name)
	}
	obj.Annotations[AnnotationPhase] = phase
	obj.Annotations[AnnotationMessage] = message
	obj.Annotations[AnnotationEntityID] = entityID
	obj.Annotations[AnnotationObservedHash] = observedHash
	return nil
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

// newObject builds a manifest object of the given kind in testNamespace.
func newObject(kind, name string, data map[string]string) *ResourceData {
	return &ResourceData{Kind: kind, Name: name, Namespace: testNamespace, Data: data, Annotations: map[string]string{}}
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
