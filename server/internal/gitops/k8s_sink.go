package gitops

import (
	"context"

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

func (s *K8sSink) SetPhase(ctx context.Context, r ResourceData, phase, message, entityID string) error {
	dataHash := ContentHash(r.Data)
	key := client.ObjectKey{Name: r.Name, Namespace: r.Namespace}

	// Patch annotations only; does not touch data. A concurrent Argo sync that
	// replaces the whole ConfigMap can still drop these keys before we read —
	// next reconcile will rewrite them, or skipReconcile may miss if hash is gone.
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		var cm corev1.ConfigMap
		if err := s.client.Get(ctx, key, &cm); err != nil {
			return err
		}
		if cm.Annotations == nil {
			cm.Annotations = map[string]string{}
		}
		if cm.Annotations[AnnotationPhase] == phase &&
			cm.Annotations[AnnotationMessage] == message &&
			cm.Annotations[AnnotationEntityID] == entityID &&
			cm.Annotations[AnnotationObservedHash] == dataHash {
			return nil
		}
		original := cm.DeepCopy()
		cm.Annotations[AnnotationPhase] = phase
		cm.Annotations[AnnotationMessage] = message
		cm.Annotations[AnnotationEntityID] = entityID
		cm.Annotations[AnnotationObservedHash] = dataHash
		return s.client.Patch(ctx, &cm, client.MergeFrom(original))
	})
	if err != nil {
		return err
	}

	if s.recorder != nil {
		var cm corev1.ConfigMap
		cm.Name = r.Name
		cm.Namespace = r.Namespace
		eventType := corev1.EventTypeNormal
		reason, eventMsg := "Pending", message
		switch phase {
		case PhaseReady:
			reason, eventMsg = "Synced", "resource synced to OLake"
		case PhaseFailed:
			eventType = corev1.EventTypeWarning
			reason, eventMsg = "SyncFailed", message
		}
		s.recorder.Event(&cm, eventType, reason, eventMsg)
	}
	return nil
}
