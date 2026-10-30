package gitops

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

// The success and FAILED-status paths, and a test that cannot start, are covered by the reconcile
// tests; these are the outcomes only the connection helper sees.
func TestTestSourceConnection(t *testing.T) {
	tests := []struct {
		name    string
		result  operationResult
		wantErr string
	}{
		{name: "output without a connection status", result: operationResult{output: map[string]interface{}{}}, wantErr: "connection status not found"},
		{name: "workflow failed while running", result: operationResult{err: errors.New("activity timed out")}, wantErr: "activity timed out"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeService := newFakeETL()
			fakeService.operationResults[testSourceOperationID] = tt.result
			err := testSourceConnection(context.Background(), fakeService, testProjectID, "postgres", "v1.0.0", dto.JSONConfig(`{"host":"pg"}`))
			assert.ErrorContains(t, err, tt.wantErr)
			assert.ErrorIs(t, err, ErrNonRetryable, "the test ran, so retrying gives the same result")
		})
	}
}
