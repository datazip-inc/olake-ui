package gitops

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8srecord "k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
)

// K8sSink patches ConfigMap annotations; failure indicators are dispatched via Temporal to the worker.
//
// Status lives in metadata.annotations because ConfigMap has no status
// subresource. Git applies labels+data; we apply olake.io/* annotations only
// via MergeFrom patch. Annotation loss is possible when GitOps uses replace
// semantics or overwrites metadata — then reconcilers re-run as if never
// synced (skipReconcile returns false). Not an issue for file mode (GITOPS_FILE_DIR).
type K8sSink struct {
	client   client.Client
	recorder k8srecord.EventRecorder
	temporal *temporal.Temporal
}

func NewK8sSink(c client.Client, recorder k8srecord.EventRecorder, t *temporal.Temporal) *K8sSink {
	return &K8sSink{client: c, recorder: recorder, temporal: t}
}

func (s *K8sSink) SpawnIndicator(ctx context.Context, r ResourceData, errMsg string) error {
	return spawnIndicatorViaTemporal(ctx, s.temporal, r, errMsg)
}

func (s *K8sSink) DeleteIndicator(ctx context.Context, r ResourceData) error {
	return deleteIndicatorViaTemporal(ctx, s.temporal, r)
}

func (s *K8sSink) SetPhase(ctx context.Context, r ResourceData, phase, message, entityID, observedHash string) error {
	dataHash := observedHashForResource(r, observedHash)
	key := client.ObjectKey{Name: r.Name, Namespace: r.Namespace}

	// Patch annotations only; does not touch data. A concurrent Argo sync that
	// replaces the whole object can still drop these keys before we read —
	// next reconcile will rewrite them, or skipReconcile may miss if hash is gone.
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		obj := newManagedObject(r.ObjectKind)
		if err := s.client.Get(ctx, key, obj); err != nil {
			return err
		}
		ann := obj.GetAnnotations()
		if ann == nil {
			ann = map[string]string{}
		}
		if ann[AnnotationPhase] == phase &&
			ann[AnnotationMessage] == message &&
			ann[AnnotationEntityID] == entityID &&
			ann[AnnotationObservedHash] == dataHash {
			return nil
		}
		original, ok := obj.DeepCopyObject().(client.Object)
		if !ok {
			return fmt.Errorf("deep copy %T: not a client.Object", obj)
		}
		ann[AnnotationPhase] = phase
		ann[AnnotationMessage] = message
		ann[AnnotationEntityID] = entityID
		ann[AnnotationObservedHash] = dataHash
		obj.SetAnnotations(ann)
		return s.client.Patch(ctx, obj, client.MergeFrom(original))
	})
	if err != nil {
		return err
	}

	if s.recorder != nil {
		evObj := newManagedObject(r.ObjectKind)
		evObj.SetName(r.Name)
		evObj.SetNamespace(r.Namespace)
		eventType := corev1.EventTypeNormal
		reason, eventMsg := "Pending", message
		switch phase {
		case PhaseReady:
			reason, eventMsg = "Synced", "resource synced to OLake"
		case PhaseFailed:
			eventType = corev1.EventTypeWarning
			reason, eventMsg = "SyncFailed", message
		}
		s.recorder.Event(evObj, eventType, reason, eventMsg)
	}
	return nil
}
