package gitops

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	k8srecord "k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	olakev1 "github.com/datazip-inc/olake-ui/server/internal/gitops/api/v1"
)

const syncRequeueAfter = 30 * time.Second

func NewStatusWriter(c client.Client, recorder k8srecord.EventRecorder) StatusWriter {
	return StatusWriter{client: c, recorder: recorder}
}

type StatusWriter struct {
	client   client.Client
	recorder k8srecord.EventRecorder
}

// SetStatus writes CR status for the given phase.
func (sw StatusWriter) SetStatus(ctx context.Context, obj client.Object, phase string, generation int64, entityID int, message string) error {
	reason, eventMsg := "Pending", message
	switch phase {
	case olakev1.PhaseReady:
		reason, eventMsg = "Synced", "resource synced to OLake"
	case olakev1.PhaseFailed:
		reason, eventMsg = "SyncFailed", message
	}
	return sw.patchStatus(ctx, obj, phase, message, entityID, generation, reason, eventMsg)
}

// SetStatusIfChanged skips the API write when status already matches.
// Useful on RequeueAfter retries where Pending would otherwise be rewritten every cycle.
func (sw StatusWriter) SetStatusIfChanged(ctx context.Context, obj client.Object, phase string, generation int64, entityID int, message string, current olakev1.ResourceStatus) error {
	if current.Phase == phase && current.ObservedGeneration == generation && current.Message == message && current.EntityID == entityID {
		return nil
	}
	return sw.SetStatus(ctx, obj, phase, generation, entityID, message)
}

func (sw StatusWriter) patchStatus(
	ctx context.Context,
	obj client.Object,
	phase, message string,
	entityID int,
	generation int64,
	reason, eventMsg string,
) error {
	key := client.ObjectKeyFromObject(obj)
	status := olakev1.ResourceStatus{
		Phase:              phase,
		Message:            message,
		EntityID:           entityID,
		ObservedGeneration: generation,
	}

	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return sw.getAndSetStatus(ctx, key, obj, status)
	})
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	if sw.recorder != nil {
		eventType := corev1.EventTypeNormal
		if phase == olakev1.PhaseFailed {
			eventType = corev1.EventTypeWarning
		}
		sw.recorder.Event(obj, eventType, reason, eventMsg)
	}
	return nil
}

// getAndSetStatus fetches the current CR and writes status in one step, per type.
// obj only picks the case; fresh is what actually gets read/written (avoids stale writes).
func (sw StatusWriter) getAndSetStatus(ctx context.Context, key client.ObjectKey, obj client.Object, status olakev1.ResourceStatus) error {
	switch obj.(type) {
	case *olakev1.Source:
		fresh := &olakev1.Source{}
		if err := sw.client.Get(ctx, key, fresh); err != nil {
			return err
		}
		fresh.Status = status
		return sw.client.Status().Update(ctx, fresh)
	case *olakev1.Destination:
		fresh := &olakev1.Destination{}
		if err := sw.client.Get(ctx, key, fresh); err != nil {
			return err
		}
		fresh.Status = status
		return sw.client.Status().Update(ctx, fresh)
	case *olakev1.Job:
		fresh := &olakev1.Job{}
		if err := sw.client.Get(ctx, key, fresh); err != nil {
			return err
		}
		fresh.Status = status
		return sw.client.Status().Update(ctx, fresh)
	case *olakev1.Streams:
		fresh := &olakev1.Streams{}
		if err := sw.client.Get(ctx, key, fresh); err != nil {
			return err
		}
		fresh.Status = status
		return sw.client.Status().Update(ctx, fresh)
	default:
		return fmt.Errorf("unsupported status object type %T", obj)
	}
}

// skipReconcile reports whether this spec generation already reached a terminal
// phase (Ready or Failed). GenerationChangedPredicate only filters watch events,
// not RequeueAfter - this guard skips redundant work on those retries.
func skipReconcile(phase string, observedGeneration, generation int64) bool {
	// spec changed since last observed - so no skip
	if observedGeneration != generation {
		return false
	}
	return phase == olakev1.PhaseReady || phase == olakev1.PhaseFailed
}

func failCR(ctx context.Context, sw StatusWriter, obj client.Object, generation int64, err error) (ctrl.Result, error) {
	_ = sw.SetStatus(ctx, obj, olakev1.PhaseFailed, generation, 0, err.Error())
	return ctrl.Result{}, nil
}
