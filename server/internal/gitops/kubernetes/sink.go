package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8srecord "k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/datazip-inc/olake-ui/server/internal/gitops/reconciler"
	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
)

// Sink is the StatusSink of Kubernetes mode. It patches ConfigMap/Secret annotations; failure indicators are dispatched via Temporal to the worker.
//
// Status lives in metadata.annotations because ConfigMap has no status
// subresource. Git applies labels+data; we apply olake.io/* annotations only
// via MergeFrom patch. Annotation loss is possible when GitOps uses replace
// semantics or overwrites metadata — then reconcilers re-run as if never
// synced.
type Sink struct {
	client   client.Client
	recorder k8srecord.EventRecorder
	temporal *temporal.Temporal
}

func NewSink(c client.Client, recorder k8srecord.EventRecorder, t *temporal.Temporal) *Sink {
	return &Sink{client: c, recorder: recorder, temporal: t}
}

func (s *Sink) SetPhase(ctx context.Context, r *reconciler.ResourceData, phase, message, entityID, observedHash string) error {
	key := client.ObjectKey{Name: r.Name, Namespace: r.Namespace}

	// Patch annotations only; does not touch data. A concurrent Argo sync that
	// replaces the whole object can still drop these keys before we read —
	// next reconcile will rewrite them, or re-run the sync if the hash is gone.
	var obj client.Object
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		obj = newManagedObject(r.ObjectKind)
		if err := s.client.Get(ctx, key, obj); err != nil {
			return err
		}
		ann := obj.GetAnnotations()
		if ann == nil {
			ann = map[string]string{}
		}
		if ann[reconciler.AnnotationPhase] == phase &&
			ann[reconciler.AnnotationMessage] == message &&
			ann[reconciler.AnnotationEntityID] == entityID &&
			ann[reconciler.AnnotationObservedHash] == observedHash {
			return nil
		}
		original, ok := obj.DeepCopyObject().(client.Object)
		if !ok {
			return fmt.Errorf("deep copy %T: not a client.Object", obj)
		}
		ann[reconciler.AnnotationPhase] = phase
		ann[reconciler.AnnotationMessage] = message
		ann[reconciler.AnnotationEntityID] = entityID
		ann[reconciler.AnnotationObservedHash] = observedHash
		obj.SetAnnotations(ann)
		return s.client.Patch(ctx, obj, client.MergeFrom(original))
	})
	if err != nil {
		return err
	}

	eventType := corev1.EventTypeNormal
	reason, eventMsg := "Pending", message
	switch phase {
	case reconciler.PhaseReady:
		reason, eventMsg = "Synced", "resource synced to OLake"
	case reconciler.PhaseFailed:
		eventType = corev1.EventTypeWarning
		reason, eventMsg = "SyncFailed", message
	}
	s.recorder.Event(obj, eventType, reason, eventMsg)
	return nil
}

// Indicator asks Temporal to spawn or delete the failure indicator for r.
func (s *Sink) Indicator(ctx context.Context, r *reconciler.ResourceData, action, errMsg string) {
	reconciler.StartIndicator(ctx, s.temporal, r, action, errMsg)
}
