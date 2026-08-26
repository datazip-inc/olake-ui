package temporal

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
)

const IndicatorWorkflow = "IndicatorWorkflow"

// IndicatorRequest is shared with the worker IndicatorWorkflow.
type IndicatorRequest struct {
	Action    string `json:"action"`    // spawn | delete
	Name      string `json:"name"`      // final pod/container name (DNS-1123)
	Namespace string `json:"namespace"` // ignored in Docker mode
	Kind      string `json:"kind"`      // source | destination | job | streams
	CRName    string `json:"cr_name"`   // originating ConfigMap name
	Message   string `json:"message"`   // error text for spawn
}

// StartIndicator submits IndicatorWorkflow fire-and-forget (no wait for result).
func (t *Temporal) StartIndicator(ctx context.Context, req IndicatorRequest) error {
	workflowID := fmt.Sprintf("indicator-%s-%s-%s-%d", req.Namespace, req.Name, req.Action, time.Now().Unix())
	_, err := t.Client.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: t.taskQueue,
	}, IndicatorWorkflow, req)
	return err
}
