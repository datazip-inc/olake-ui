package gitops

import (
	"context"
	"strconv"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

const syncRequeueAfter = 30 * time.Second

func skipReconcile(annotations map[string]string, dataHash string) bool {
	if annotations == nil {
		return false
	}
	if annotations[AnnotationObservedHash] != dataHash {
		return false
	}
	phase := annotations[AnnotationPhase]
	return phase == PhaseReady || phase == PhaseFailed
}

func failResource(ctx context.Context, sink StatusSink, r ResourceData, err error) (ctrl.Result, error) {
	msg := err.Error()
	_ = sink.SetPhase(ctx, r, PhaseFailed, msg, "")
	_ = sink.SpawnIndicator(ctx, r, msg)
	return ctrl.Result{}, nil
}

func waitResource(ctx context.Context, sink StatusSink, r ResourceData, msg string) (ctrl.Result, error) {
	_ = sink.SetPhase(ctx, r, PhasePending, msg, "")
	return ctrl.Result{RequeueAfter: syncRequeueAfter}, nil
}

func successResource(ctx context.Context, sink StatusSink, r ResourceData, entityID int) error {
	_ = sink.DeleteIndicator(ctx, r)
	return sink.SetPhase(ctx, r, PhaseReady, "", strconvEntityID(entityID))
}

func strconvEntityID(id int) string {
	if id <= 0 {
		return ""
	}
	return strconv.Itoa(id)
}
