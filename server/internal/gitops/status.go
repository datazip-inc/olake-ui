package gitops

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	syncRequeueAfter = 30 * time.Second
	syncRetryTimeout = 5 * time.Minute
)

// retryStarted tracks when a transient retry window began (process-local).
var retryStarted sync.Map // ResourceData.key() -> time.Time

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

func observedHashForResource(r ResourceData, override string) string {
	if override != "" {
		return override
	}
	return ContentHash(r.Data)
}

func failResource(ctx context.Context, sink StatusSink, r ResourceData, err error, observedHash string) (ctrl.Result, error) {
	retryStarted.Delete(r.key())
	msg := err.Error()
	hash := observedHashForResource(r, observedHash)
	_ = sink.SetPhase(ctx, r, PhaseFailed, msg, "", hash)
	_ = sink.SpawnIndicator(ctx, r, msg)
	return ctrl.Result{}, nil
}

func waitResource(ctx context.Context, sink StatusSink, r ResourceData, msg, observedHash string) (ctrl.Result, error) {
	retryStarted.Delete(r.key())
	hash := observedHashForResource(r, observedHash)
	_ = sink.SetPhase(ctx, r, PhasePending, msg, "", hash)
	return ctrl.Result{RequeueAfter: syncRequeueAfter}, nil
}

// requeueTransient retries a transient error as Pending. After syncRetryTimeout
// it fails with an indicator. Dependency waits use waitResource (no timeout).
func requeueTransient(ctx context.Context, sink StatusSink, r ResourceData, err error, observedHash string) (ctrl.Result, error) {
	now := time.Now()
	started := now
	if v, ok := retryStarted.Load(r.key()); ok {
		started = v.(time.Time)
	} else {
		retryStarted.Store(r.key(), now)
	}
	if now.Sub(started) >= syncRetryTimeout {
		return failResource(ctx, sink, r, Permanent(fmt.Errorf("sync retry timed out after %s: %w", syncRetryTimeout, err)), observedHash)
	}
	hash := observedHashForResource(r, observedHash)
	_ = sink.SetPhase(ctx, r, PhasePending, fmt.Sprintf("retrying: %s", err), "", hash)
	return ctrl.Result{RequeueAfter: syncRequeueAfter}, nil
}

func successResource(ctx context.Context, sink StatusSink, r ResourceData, entityID int, observedHash string) error {
	retryStarted.Delete(r.key())
	_ = sink.DeleteIndicator(ctx, r)
	hash := observedHashForResource(r, observedHash)
	return sink.SetPhase(ctx, r, PhaseReady, "", strconvEntityID(entityID), hash)
}

func strconvEntityID(id int) string {
	if id <= 0 {
		return ""
	}
	return strconv.Itoa(id)
}
