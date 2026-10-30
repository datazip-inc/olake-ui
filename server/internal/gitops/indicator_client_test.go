package gitops

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/client"
	"k8s.io/client-go/tools/record"

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

func TestK8sSinkIndicator(t *testing.T) {
	ctx := context.Background()
	res := &ResourceData{Kind: KindJob, Name: "Orders_CM", Namespace: testNamespace}
	newSink := func(temporalClient *fakeTemporalClient) *K8sSink {
		return NewK8sSink(nil, record.NewFakeRecorder(1), &temporal.Temporal{Client: temporalClient})
	}
	// the pod name is the object name made DNS-1123 safe, plus the kind
	pod := temporal.IndicatorRequest{Name: "orders-cm-olake-job", Namespace: testNamespace, Kind: KindJob, ResourceName: "Orders_CM"}

	t.Run("spawn sends the pod and the error for its termination log", func(t *testing.T) {
		temporalClient := &fakeTemporalClient{}
		newSink(temporalClient).Indicator(ctx, res, indicatorSpawn, "source connection test failed")
		want := pod
		want.Action, want.Message = indicatorSpawn, "source connection test failed"
		assert.Equal(t, []workflowStart{indicatorStart(&want)}, temporalClient.workflowStarts)
	})

	t.Run("message is cut to the termination log limit", func(t *testing.T) {
		temporalClient := &fakeTemporalClient{}
		newSink(temporalClient).Indicator(ctx, res, indicatorSpawn, strings.Repeat("x", terminationLogMax+100))
		want := pod
		want.Action, want.Message = indicatorSpawn, strings.Repeat("x", terminationLogMax)
		assert.Equal(t, []workflowStart{indicatorStart(&want)}, temporalClient.workflowStarts)
	})

	t.Run("delete names the same pod", func(t *testing.T) {
		temporalClient := &fakeTemporalClient{}
		newSink(temporalClient).Indicator(ctx, res, indicatorDelete, "")
		want := pod
		want.Action = indicatorDelete
		assert.Equal(t, []workflowStart{indicatorStart(&want)}, temporalClient.workflowStarts)
	})

	t.Run("a failed start does not fail the reconcile", func(t *testing.T) {
		// the indicator is best effort: the phase annotation and the event already carry the error
		temporalClient := &fakeTemporalClient{err: errors.New("temporal unavailable")}
		newSink(temporalClient).Indicator(ctx, res, indicatorSpawn, "boom")
		assert.Len(t, temporalClient.workflowStarts, 1)
	})
}
