package temporal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	workflowservice "go.temporal.io/api/workflowservice/v1"
)

// OperationKind is also the operation ID's prefix.
type OperationKind string

const (
	OperationSpec             OperationKind = "fetch-spec"
	OperationTestConnection   OperationKind = "test-connection"
	OperationDiscoverCatalog  OperationKind = "discover-catalog"
	OperationStreamDifference OperationKind = "difference"
)

// pollableKinds is an allow-list: sync and clear-destination workflows have their own
// endpoints and authorization, and must not be reachable through the operations API.
var pollableKinds = []OperationKind{
	OperationSpec,
	OperationTestConnection,
	OperationDiscoverCatalog,
	OperationStreamDifference,
}

type OperationStatus string

const (
	OperationRunning   OperationStatus = "running"
	OperationSucceeded OperationStatus = "succeeded"
	OperationFailed    OperationStatus = "failed"
)

const operationIDSuffixLen = 32

// NewOperationID builds "<kind>-<projectID>-<random hex>". The random suffix makes IDs
// unguessable across projects and collision-free; the project ID lets pollers be
// authorized without a database lookup.
func NewOperationID(kind OperationKind, projectID string) (string, error) {
	if strings.TrimSpace(projectID) == "" {
		return "", fmt.Errorf("project id is required to start an operation")
	}
	buf := make([]byte, operationIDSuffixLen/2)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to generate operation id: %s", err)
	}
	return fmt.Sprintf("%s-%s-%s", kind, projectID, hex.EncodeToString(buf)), nil
}

// ParseOperationID returns the kind and the project the ID was issued for.
func ParseOperationID(operationID string) (OperationKind, string, error) {
	for _, kind := range pollableKinds {
		prefix := string(kind) + "-"
		if !strings.HasPrefix(operationID, prefix) {
			continue
		}
		rest := operationID[len(prefix):]
		// The suffix is fixed-width hex with no '-', so the last '-' always separates it
		// from the project ID, even when the project ID itself contains hyphens.
		sep := strings.LastIndexByte(rest, '-')
		if sep <= 0 {
			return "", "", fmt.Errorf("malformed operation id")
		}
		projectID, suffix := rest[:sep], rest[sep+1:]
		if len(suffix) != operationIDSuffixLen {
			return "", "", fmt.Errorf("malformed operation id")
		}
		if _, err := hex.DecodeString(suffix); err != nil {
			return "", "", fmt.Errorf("malformed operation id")
		}
		return kind, projectID, nil
	}
	return "", "", fmt.Errorf("unrecognised operation id")
}

func operationStatusFor(status enumspb.WorkflowExecutionStatus) OperationStatus {
	switch status {
	case enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING, enumspb.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW:
		return OperationRunning
	case enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		return OperationSucceeded
	default:
		return OperationFailed
	}
}

func IsOperationNotFound(err error) bool {
	var notFound *serviceerror.NotFound
	return errors.As(err, &notFound)
}

// DescribeOperation returns the operation's state. When wait is positive it first
// long-polls Temporal for up to wait, returning as soon as the workflow closes.
func (t *Temporal) DescribeOperation(ctx context.Context, operationID string, wait time.Duration) (OperationStatus, error) {
	if wait > 0 {
		t.awaitClose(ctx, operationID, wait)
	}

	desc, err := t.Client.DescribeWorkflowExecution(ctx, operationID, "")
	if err != nil {
		return "", err
	}

	return operationStatusFor(desc.GetWorkflowExecutionInfo().GetStatus()), nil
}

func (t *Temporal) awaitClose(ctx context.Context, operationID string, wait time.Duration) {
	timedCtx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()

	_, err := t.Client.WorkflowService().GetWorkflowExecutionHistory(
		timedCtx,
		&workflowservice.GetWorkflowExecutionHistoryRequest{
			Namespace:              temporalNamespace(),
			Execution:              &commonpb.WorkflowExecution{WorkflowId: operationID},
			WaitNewEvent:           true,
			HistoryEventFilterType: enumspb.HISTORY_EVENT_FILTER_TYPE_CLOSE_EVENT,
		},
	)
	if err == nil || timedCtx.Err() != nil || IsOperationNotFound(err) {
		return
	}

	// The long-poll failed early, so the workflow will most likely still read as running.
	// Hold until the window ends anyway: answering instantly would let clients poll in a
	// tight loop for as long as the error persists.
	logger.Debugf("long-poll for operation %s ended early: %s", operationID, err)
	<-timedCtx.Done()
}

// FetchOperationOutput must only be called once the operation has closed; run.Get then
// returns immediately instead of blocking.
func (t *Temporal) FetchOperationOutput(ctx context.Context, operationID string) (map[string]interface{}, error) {
	return ExtractWorkflowResponse(ctx, t.Client.GetWorkflow(ctx, operationID, ""))
}

// OperationFailureMessage recovers why a closed operation failed. DescribeWorkflowExecution
// reports the status but not the cause, which only surfaces through the workflow result.
func (t *Temporal) OperationFailureMessage(ctx context.Context, operationID string) string {
	result := make(map[string]interface{})
	if err := t.Client.GetWorkflow(ctx, operationID, "").Get(ctx, &result); err != nil {
		return err.Error()
	}
	return ""
}

// temporalNamespace is needed for raw WorkflowService calls, which, unlike the SDK
// client, don't carry the namespace implicitly.
func temporalNamespace() string {
	if ns := strings.TrimSpace(appconfig.Load().TemporalNamespace); ns != "" {
		return ns
	}
	return "default"
}
