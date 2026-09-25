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
	"go.temporal.io/sdk/converter"
)

// OperationKind identifies what a long-running operation does. It is also the prefix of
// the operation ID, which is how a poll request is routed to the right result decoder.
type OperationKind string

const (
	OperationSpec             OperationKind = "fetch-spec"
	OperationTestConnection   OperationKind = "test-connection"
	OperationDiscoverCatalog  OperationKind = "discover-catalog"
	OperationStreamDifference OperationKind = "difference"
)

// pollableKinds is the allow-list for the operations endpoints. Only IDs carrying one of
// these prefixes can be polled, so sync and clear-destination workflows — which have
// their own endpoints and their own authorization — cannot be probed through this API.
var pollableKinds = []OperationKind{
	OperationSpec,
	OperationTestConnection,
	OperationDiscoverCatalog,
	OperationStreamDifference,
}

// OperationStatus is the API-facing lifecycle state of an operation.
type OperationStatus string

const (
	OperationRunning   OperationStatus = "running"
	OperationSucceeded OperationStatus = "succeeded"
	OperationFailed    OperationStatus = "failed"
	OperationCanceled  OperationStatus = "canceled"
	OperationTimedOut  OperationStatus = "timed_out"
)

// Terminal reports whether no further status change is possible.
func (s OperationStatus) Terminal() bool { return s != OperationRunning }

// operationIDSuffixLen is the hex width of the random suffix in an operation ID.
const operationIDSuffixLen = 32

// Memo keys. A spec response echoes back the connector type and version that were asked
// for, but the result endpoint never saw the original request. Rather than trusting the
// client to resend them or adding a table, they ride along on the workflow's Temporal
// memo — keeping operation state in Temporal, where the rest of it already lives.
const (
	MemoSpecType    = "spec_type"
	MemoSpecVersion = "spec_version"
)

// OperationState is what a status poll reports back.
type OperationState struct {
	ID         string
	Kind       OperationKind
	Status     OperationStatus
	StartedAt  *time.Time
	FinishedAt *time.Time
	// Memo carries values recorded when the operation was started, such as the connector
	// type and version a spec fetch was asked for.
	Memo map[string]string
}

// NewOperationID builds "<kind>-<projectID>-<random hex>".
// The random suffix replaces the unix-second suffix these workflow IDs used to carry.
// That scheme was guessable, which would let one project poll another project's
// operation now that polling exists, and it collided: two spec fetches for the same
// connector type within the same second produced an identical ID. The project ID is
// embedded so the poll handlers can authorize without a database lookup.
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

// ParseOperationID validates an operation ID and recovers the kind and project it was
// issued for. Callers must compare the returned project against the request's project:
// the ID is the only thing tying an operation to its owner.
func ParseOperationID(operationID string) (OperationKind, string, error) {
	for _, kind := range pollableKinds {
		prefix := string(kind) + "-"
		if !strings.HasPrefix(operationID, prefix) {
			continue
		}
		rest := operationID[len(prefix):]
		// The random suffix is fixed-width hex containing no separator, so the final '-'
		// always divides it from the project ID — even when the project ID has hyphens.
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

// operationStatusFor maps Temporal's execution status onto the API vocabulary.
func operationStatusFor(status enumspb.WorkflowExecutionStatus) OperationStatus {
	switch status {
	case enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING, enumspb.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW:
		return OperationRunning
	case enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		return OperationSucceeded
	case enumspb.WORKFLOW_EXECUTION_STATUS_CANCELED:
		return OperationCanceled
	case enumspb.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
		return OperationTimedOut
	default: // FAILED, TERMINATED, UNSPECIFIED
		return OperationFailed
	}
}

// IsOperationNotFound reports whether Temporal has no record of the operation, which
// happens for an unknown ID or once namespace retention has expired.
func IsOperationNotFound(err error) bool {
	var notFound *serviceerror.NotFound
	return errors.As(err, &notFound)
}

// DescribeOperation reports an operation's current state. This is the authoritative
// source: the work runs in a Temporal workflow, so Temporal — not the shared config
// directory — knows whether it is running, finished or failed.
func (t *Temporal) DescribeOperation(ctx context.Context, operationID string) (*OperationState, error) {
	kind, _, err := ParseOperationID(operationID)
	if err != nil {
		return nil, err
	}

	desc, err := t.Client.DescribeWorkflowExecution(ctx, operationID, "")
	if err != nil {
		return nil, err
	}

	info := desc.GetWorkflowExecutionInfo()
	state := &OperationState{
		ID:     operationID,
		Kind:   kind,
		Status: operationStatusFor(info.GetStatus()),
	}
	if ts := info.GetStartTime(); ts != nil {
		started := ts.AsTime()
		state.StartedAt = &started
	}
	if ts := info.GetCloseTime(); ts != nil {
		finished := ts.AsTime()
		state.FinishedAt = &finished
	}
	state.Memo = decodeMemo(info.GetMemo())
	return state, nil
}

// decodeMemo unpacks the string-valued memo fields recorded at start time. The same
// payload-decoding approach is already used for search attributes in the job services.
func decodeMemo(memo *commonpb.Memo) map[string]string {
	if memo == nil || len(memo.GetFields()) == 0 {
		return nil
	}
	dc := converter.GetDefaultDataConverter()
	out := make(map[string]string, len(memo.GetFields()))
	for key, payload := range memo.GetFields() {
		var value string
		if err := dc.FromPayload(payload, &value); err == nil {
			out[key] = value
		}
	}
	return out
}

// AwaitOperationClose blocks until the operation closes or wait elapses, whichever comes
// first, then returns. It uses Temporal's history long-poll rather than a ticker, so it
// returns the moment the workflow finishes instead of on the next poll boundary.
//
// Holding a request for a bounded few seconds is the point: it saves the client from
// holding one open for the entire operation, which is what proxies and VPNs cut.
// Errors are deliberately not surfaced — the caller re-describes the operation either
// way, and a long-poll timeout simply means "still running".
func (t *Temporal) AwaitOperationClose(ctx context.Context, operationID string, wait time.Duration) {
	if wait <= 0 {
		return
	}

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
	if err != nil && timedCtx.Err() == nil {
		logger.Debugf("long-poll for operation %s ended early: %s", operationID, err)
	}
}

// FetchOperationOutput returns the workflow's decoded output. Call it only once the
// operation has closed, where the underlying run.Get returns immediately rather than
// blocking. The payload itself is read off the shared volume, not carried by Temporal.
func (t *Temporal) FetchOperationOutput(ctx context.Context, operationID string) (map[string]interface{}, error) {
	return ExtractWorkflowResponse(ctx, t.Client.GetWorkflow(ctx, operationID, ""))
}

// OperationFailureMessage recovers why a closed operation did not succeed.
// DescribeWorkflowExecution reports the status but not the cause, which only surfaces
// through the workflow result.
func (t *Temporal) OperationFailureMessage(ctx context.Context, operationID string) string {
	result := make(map[string]interface{})
	if err := t.Client.GetWorkflow(ctx, operationID, "").Get(ctx, &result); err != nil {
		return err.Error()
	}
	return ""
}

// CancelOperation stops a running operation. Without this, abandoning the UI leaves the
// connector container running to completion with nobody to read its output.
func (t *Temporal) CancelOperation(ctx context.Context, operationID string) error {
	if err := t.Client.CancelWorkflow(ctx, operationID, ""); err != nil {
		return fmt.Errorf("failed to cancel operation: %s", err)
	}
	return nil
}

// temporalNamespace resolves the configured namespace for raw WorkflowService calls,
// which — unlike the SDK client — do not carry it implicitly.
func temporalNamespace() string {
	if ns := strings.TrimSpace(appconfig.Load().TemporalNamespace); ns != "" {
		return ns
	}
	return "default"
}
