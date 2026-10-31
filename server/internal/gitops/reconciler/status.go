package reconciler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	syncRequeueAfter = 30 * time.Second
	// syncRetryTimeout bounds retries of a transient error (Temporal or the DB unreachable); after it the
	// resource is Failed and is not retried until its data changes. If the cause is fixed later, force a
	// retry with: `kubectl annotate cm|secret <name> olake.io/phase-`
	syncRetryTimeout = 30 * time.Minute
)

// retryMap tracks retries
var retryMap sync.Map // ResourceData.key() -> time.Time (when retry started)

// skipReconcile: true when this resource was already synced for the current fingerprint.
// How: observed-hash matches dataHash AND phase is Ready or Failed.
// Hash mismatch (content/connectors changed) or non-terminal phase → run reconcile.
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

func observedHashForResource(r *ResourceData, override string) string {
	if override != "" {
		return override
	}
	return ContentHash(r.Data)
}

func failResource(ctx context.Context, sink StatusSink, r *ResourceData, err error, observedHash string) (ctrl.Result, error) {
	retryMap.Delete(r.key())
	msg := err.Error()
	hash := observedHashForResource(r, observedHash)
	_ = sink.SetPhase(ctx, r, PhaseFailed, msg, r.Annotations[AnnotationEntityID], hash)
	sink.Indicator(ctx, r, indicatorSpawn, msg)
	return ctrl.Result{}, nil
}

func waitResource(ctx context.Context, sink StatusSink, r *ResourceData, msg, observedHash string) (ctrl.Result, error) {
	retryMap.Delete(r.key())
	hash := observedHashForResource(r, observedHash)
	_ = sink.SetPhase(ctx, r, PhasePending, msg, r.Annotations[AnnotationEntityID], hash)
	return ctrl.Result{RequeueAfter: syncRequeueAfter}, nil
}

// failOrRequeue fails r on a non-retryable error and retries any other error as Pending.
func failOrRequeue(ctx context.Context, sink StatusSink, r *ResourceData, err error, observedHash string) (ctrl.Result, error) {
	if errors.Is(err, ErrNonRetryable) {
		return failResource(ctx, sink, r, err, observedHash)
	}
	return requeueTransient(ctx, sink, r, err, observedHash)
}

// requeueTransient retries a transient error as Pending. After syncRetryTimeout
// it fails with an indicator. Dependency waits use waitResource (no timeout).
func requeueTransient(ctx context.Context, sink StatusSink, r *ResourceData, err error, observedHash string) (ctrl.Result, error) {
	now := time.Now()
	started := now
	if v, ok := retryMap.Load(r.key()); ok {
		started = v.(time.Time)
	} else {
		retryMap.Store(r.key(), now)
	}
	if now.Sub(started) >= syncRetryTimeout {
		return failResource(ctx, sink, r, NonRetryableError(fmt.Errorf("sync retry timed out after %s: %w", syncRetryTimeout, err)), observedHash)
	}
	hash := observedHashForResource(r, observedHash)
	_ = sink.SetPhase(ctx, r, PhasePending, fmt.Sprintf("retrying: %s", err), r.Annotations[AnnotationEntityID], hash)
	return ctrl.Result{RequeueAfter: syncRequeueAfter}, nil
}

// successResource marks r Ready. When r is already Ready with the same status, it writes nothing:
// a re-check, such as a secret refresh, would otherwise emit a Synced event and start an indicator
// delete workflow every time.
func successResource(ctx context.Context, sink StatusSink, r *ResourceData, entityID int, observedHash string) error {
	retryMap.Delete(r.key())
	hash := observedHashForResource(r, observedHash)
	id := strconvEntityID(entityID)
	if r.Annotations[AnnotationPhase] == PhaseReady &&
		r.Annotations[AnnotationMessage] == "" &&
		r.Annotations[AnnotationEntityID] == id &&
		r.Annotations[AnnotationObservedHash] == hash {
		return nil
	}
	sink.Indicator(ctx, r, indicatorDelete, "")
	return sink.SetPhase(ctx, r, PhaseReady, "", id, hash)
}

func strconvEntityID(id int) string {
	if id <= 0 {
		return ""
	}
	return strconv.Itoa(id)
}
