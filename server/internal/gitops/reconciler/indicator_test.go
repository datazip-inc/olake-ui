package reconciler

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/client"

	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
)

// fakeTemporalClient records workflow starts. It implements only ExecuteWorkflow; any other client
// method panics on the nil embedded interface, so a test fails loudly if the sink starts using one.
type fakeTemporalClient struct {
	client.Client
	err            error
	workflowStarts []workflowStart
}

type workflowStart struct {
	workflow interface{}
	args     []interface{}
}

func (f *fakeTemporalClient) ExecuteWorkflow(_ context.Context, _ client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) { //nolint:gocritic // signature fixed by client.Client
	f.workflowStarts = append(f.workflowStarts, workflowStart{workflow: workflow, args: args})
	return nil, f.err
}

// indicatorStart is the workflow start the sink should make for req.
func indicatorStart(req *temporal.IndicatorRequest) workflowStart {
	return workflowStart{workflow: temporal.IndicatorWorkflow, args: []interface{}{req}}
}

func TestStartIndicator(t *testing.T) {
	ctx := context.Background()
	res := &ResourceData{Kind: KindJob, Name: "Orders_CM", Namespace: testNamespace}
	start := func(temporalClient *fakeTemporalClient, action, errMsg string) {
		StartIndicator(ctx, &temporal.Temporal{Client: temporalClient}, res, action, errMsg)
	}
	// the pod name is the kind plus the object name made DNS-1123 safe
	pod := temporal.IndicatorRequest{Name: "olake-failed-job-orders-cm", Namespace: testNamespace, Kind: KindJob, ResourceName: "Orders_CM"}

	t.Run("spawn sends the pod and the error for its termination log", func(t *testing.T) {
		temporalClient := &fakeTemporalClient{}
		start(temporalClient, indicatorSpawn, "source connection test failed")
		want := pod
		want.Action, want.Message = indicatorSpawn, "source connection test failed"
		assert.Equal(t, []workflowStart{indicatorStart(&want)}, temporalClient.workflowStarts)
	})

	t.Run("delete names the same pod", func(t *testing.T) {
		temporalClient := &fakeTemporalClient{}
		start(temporalClient, indicatorDelete, "")
		want := pod
		want.Action = indicatorDelete
		assert.Equal(t, []workflowStart{indicatorStart(&want)}, temporalClient.workflowStarts)
	})

	t.Run("a failed start does not fail the reconcile", func(t *testing.T) {
		// the indicator is best effort: the status already carries the error
		temporalClient := &fakeTemporalClient{err: errors.New("temporal unavailable")}
		start(temporalClient, indicatorSpawn, "boom")
		assert.Len(t, temporalClient.workflowStarts, 1)
	})
}

func TestIndicatorName(t *testing.T) {
	tests := []struct {
		name string
		kind string
		want string
	}{
		{name: "orders", kind: KindJob, want: "olake-failed-job-orders"},
		{name: "Prod_Orders.v2", kind: KindJob, want: "olake-failed-job-prod-orders-v2"},
		{name: "a--b__c", kind: KindSource, want: "olake-failed-source-a-b-c"},
		{name: "-edge-", kind: KindSource, want: "olake-failed-source-edge"},
		{name: "Café", kind: KindSource, want: "olake-failed-source-caf"},
		{name: "___", kind: KindDestination, want: "olake-failed-destination-unnamed"},
		{name: "", kind: KindStreams, want: "olake-failed-streams-unnamed"},
		// cut to 63 chars without leaving a trailing dash
		{name: strings.Repeat("a", 46) + "-bbbbbbbbb", kind: KindJob, want: "olake-failed-job-" + strings.Repeat("a", 46)},
		{name: strings.Repeat("x", 80), kind: KindJob, want: "olake-failed-job-" + strings.Repeat("x", 46)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, indicatorName(tt.name, tt.kind))
		})
	}
}
