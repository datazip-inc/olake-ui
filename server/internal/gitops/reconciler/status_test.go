package reconciler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamsCMApplied(t *testing.T) {
	data := map[string]string{DataConfig: userSelectedStreams}
	hash := ContentHash(data)

	assert.False(t, streamsCMApplied(nil, data))
	assert.True(t, streamsCMApplied(map[string]string{AnnotationPhase: PhaseReady, AnnotationObservedHash: hash}, data))
	// unlike skipReconcile, a failed selection was never applied, so it is still streams drift
	assert.False(t, streamsCMApplied(map[string]string{AnnotationPhase: PhaseFailed, AnnotationObservedHash: hash}, data))
	assert.False(t, streamsCMApplied(map[string]string{AnnotationPhase: PhaseReady, AnnotationObservedHash: "old"}, data))
}

func TestContentHash(t *testing.T) {
	a := ContentHash(map[string]string{"projectId": "1", "config": "{}"})
	assert.Equal(t, a, ContentHash(map[string]string{"config": "{}", "projectId": "1"}), "key order must not matter")
	assert.NotEqual(t, a, ContentHash(map[string]string{"projectId": "1", "config": "{ }"}), "any data byte change must change the hash")
	// values that would collide under naive concatenation
	assert.NotEqual(t,
		ContentHash(map[string]string{"a": "b", "c": "d"}),
		ContentHash(map[string]string{"a": "b\",\"c\":\"d"}))
}

func TestRequeueTransient(t *testing.T) {
	ctx := context.Background()
	res := &ResourceData{Kind: KindSource, Name: "pg", Namespace: testNamespace, Annotations: map[string]string{AnnotationEntityID: "5"}}
	cause := errors.New("db unreachable")

	t.Run("within the retry window it is pending and requeued", func(t *testing.T) {
		e := newTestEnv(t, newObject(KindSource, "pg", nil))
		result, err := requeueTransient(ctx, e.status, res, cause, "h")
		require.NoError(t, err)
		assert.Equal(t, syncRequeueAfter, result.RequeueAfter)
		assert.Equal(t, statusUpdate{objectName: "pg", phase: PhasePending, message: "retrying: db unreachable", entityID: "5"}, e.status.lastStatusUpdate(t, "pg"))
		assert.Empty(t, e.status.indicatorActions)

		// the retry window starts at the first failure and is kept across retries
		first, ok := retryMap.Load(res.key())
		require.True(t, ok)
		_, _ = requeueTransient(ctx, e.status, res, cause, "h")
		again, _ := retryMap.Load(res.key())
		assert.Equal(t, first, again)
	})

	t.Run("after the retry window it fails with an indicator", func(t *testing.T) {
		e := newTestEnv(t, newObject(KindSource, "pg", nil))
		retryMap.Store(res.key(), time.Now().Add(-syncRetryTimeout))
		result, err := requeueTransient(ctx, e.status, res, cause, "h")
		require.NoError(t, err)
		assert.Zero(t, result.RequeueAfter)
		last := e.status.lastStatusUpdate(t, "pg")
		assert.Equal(t, PhaseFailed, last.phase)
		assert.Contains(t, last.message, "sync retry timed out after 30m0s: db unreachable")
		assert.Equal(t, []string{indicatorSpawn}, e.status.indicatorActionsFor("pg"))
		_, ok := retryMap.Load(res.key())
		assert.False(t, ok, "a terminal phase must reset the retry window")
	})

	t.Run("a success or a wait resets the retry window", func(t *testing.T) {
		e := newTestEnv(t, newObject(KindSource, "pg", nil))
		retryMap.Store(res.key(), time.Now().Add(-time.Hour))
		require.NoError(t, successResource(ctx, e.status, res, 5, "h"))
		_, ok := retryMap.Load(res.key())
		assert.False(t, ok)

		retryMap.Store(res.key(), time.Now().Add(-time.Hour))
		_, _ = waitResource(ctx, e.status, res, "waiting", "h")
		_, ok = retryMap.Load(res.key())
		assert.False(t, ok)
	})
}
