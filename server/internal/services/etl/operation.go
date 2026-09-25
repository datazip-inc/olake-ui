package etl

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
)

// Operation errors. They live here rather than in the shared constants package because
// nothing outside the operations API raises or inspects them.
var (
	ErrOperationNotFound  = errors.New("operation not found")
	ErrOperationNotDone   = errors.New("operation has not finished")
	ErrOperationFailed    = errors.New("operation failed")
	ErrOperationForbidden = errors.New("operation does not belong to this project")
)

// operationMaxWait is the ceiling on how long a status poll may be held server-side.
//
// It must stay comfortably below any proxy or VPN idle timeout in front of the server:
// holding the request briefly is what saves the client from holding one open for the whole
// operation, but holding it too long recreates the very problem this exists to solve. Ten
// seconds leaves roughly 3x margin under an aggressive 30s proxy.
const operationMaxWait = 10 * time.Second

// resolveOperation validates the ID and confirms it belongs to the requesting project.
// The project is embedded in the ID, which is the only thing tying an operation to its
// owner, so this check is what stops one project polling another's discovery output.
func resolveOperation(projectID, operationID string) (temporal.OperationKind, error) {
	kind, owner, err := temporal.ParseOperationID(operationID)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrOperationNotFound, err)
	}
	if owner != projectID {
		return "", ErrOperationForbidden
	}
	return kind, nil
}

// GetOperationStatus reports whether an operation is still running, and if not, how it
// ended. When wait is positive the call blocks for up to that long waiting for the
// operation to close, returning as soon as it does.
func (s Service) GetOperationStatus(ctx context.Context, projectID, operationID string, wait time.Duration) (*dto.OperationStatusResponse, error) {
	if s.temporal == nil {
		return nil, fmt.Errorf("temporal client not available")
	}
	if _, err := resolveOperation(projectID, operationID); err != nil {
		return nil, err
	}

	if wait > 0 {
		if wait > operationMaxWait {
			wait = operationMaxWait
		}
		s.temporal.AwaitOperationClose(ctx, operationID, wait)
	}

	state, err := s.temporal.DescribeOperation(ctx, operationID)
	if err != nil {
		if temporal.IsOperationNotFound(err) {
			return nil, fmt.Errorf("%w: %s", ErrOperationNotFound, operationID)
		}
		return nil, fmt.Errorf("failed to describe operation: %s", err)
	}

	resp := &dto.OperationStatusResponse{
		OperationID: state.ID,
		Kind:        string(state.Kind),
		Status:      string(state.Status),
	}
	if state.StartedAt != nil {
		resp.StartedAt = state.StartedAt.UTC().Format(time.RFC3339)
	}
	if state.FinishedAt != nil {
		resp.FinishedAt = state.FinishedAt.UTC().Format(time.RFC3339)
	}
	// DescribeWorkflowExecution reports that an operation failed but not why, so the
	// reason is recovered from the workflow result. Only done for closed, unsuccessful
	// operations, where that read returns immediately.
	if state.Status.Terminal() && state.Status != temporal.OperationSucceeded {
		resp.Error = s.temporal.OperationFailureMessage(ctx, operationID)
	}

	return resp, nil
}

// GetOperationResult returns a finished operation's payload, shaped into the same DTO the
// endpoint that started it used to return synchronously.
func (s Service) GetOperationResult(ctx context.Context, projectID, operationID string) (interface{}, error) {
	if s.temporal == nil {
		return nil, fmt.Errorf("temporal client not available")
	}
	kind, err := resolveOperation(projectID, operationID)
	if err != nil {
		return nil, err
	}

	state, err := s.temporal.DescribeOperation(ctx, operationID)
	if err != nil {
		if temporal.IsOperationNotFound(err) {
			return nil, fmt.Errorf("%w: %s", ErrOperationNotFound, operationID)
		}
		return nil, fmt.Errorf("failed to describe operation: %s", err)
	}

	switch state.Status {
	case temporal.OperationRunning:
		return nil, ErrOperationNotDone
	case temporal.OperationSucceeded:
	default:
		// The workflow itself did not complete — a failed image pull, a timeout, a
		// cancellation. This is distinct from a connector that ran and reported it could
		// not reach the database, which is a successful operation carrying a FAILED
		// connection status.
		reason := s.temporal.OperationFailureMessage(ctx, operationID)
		if reason == "" {
			reason = string(state.Status)
		}
		return nil, fmt.Errorf("%w: %s", ErrOperationFailed, reason)
	}

	output, err := s.temporal.FetchOperationOutput(ctx, operationID)
	if err != nil {
		return nil, fmt.Errorf("failed to read operation result: %s", err)
	}

	return s.decodeOperationOutput(kind, operationID, state, output)
}

// decodeOperationOutput reshapes a workflow's raw output into the response body its
// originating endpoint has always returned. Keeping the shaping here — rather than in the
// handler — is what keeps the payloads identical to the previous blocking API.
func (s Service) decodeOperationOutput(
	kind temporal.OperationKind,
	operationID string,
	state *temporal.OperationState,
	output map[string]interface{},
) (interface{}, error) {
	switch kind {
	case temporal.OperationSpec:
		return dto.SpecResponse{
			Version: state.Memo[temporal.MemoSpecVersion],
			Type:    state.Memo[temporal.MemoSpecType],
			Spec:    output,
		}, nil

	case temporal.OperationTestConnection:
		result, err := temporal.DecodeConnectionStatus(output)
		if err != nil {
			return nil, err
		}
		// Tail the connector's own logs from the operation directory, which is named
		// after the operation ID.
		logs, err := utils.ReadLogs(filepath.Join(constants.DefaultConfigDir, operationID), -1, -1, "older")
		if err != nil {
			return nil, fmt.Errorf("failed to read logs for operation %s: %s", operationID, err)
		}
		return dto.TestConnectionResponse{ConnectionResult: result, Logs: logs.Logs}, nil

	case temporal.OperationDiscoverCatalog:
		return output, nil

	case temporal.OperationStreamDifference:
		return dto.StreamDifferenceResponse{DifferenceStreams: output}, nil

	default:
		return nil, fmt.Errorf("unsupported operation kind %q", kind)
	}
}

// CancelOperation stops a running operation and the connector container behind it.
// Only an explicit request cancels; closing a page does not.
func (s Service) CancelOperation(ctx context.Context, projectID, operationID string) error {
	if s.temporal == nil {
		return fmt.Errorf("temporal client not available")
	}
	if _, err := resolveOperation(projectID, operationID); err != nil {
		return err
	}
	if err := s.temporal.CancelOperation(ctx, operationID); err != nil {
		if temporal.IsOperationNotFound(err) {
			return fmt.Errorf("%w: %s", ErrOperationNotFound, operationID)
		}
		return err
	}
	return nil
}
