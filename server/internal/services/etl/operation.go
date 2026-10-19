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

var (
	ErrOperationNotFound  = errors.New("operation not found")
	ErrOperationNotDone   = errors.New("operation has not finished")
	ErrOperationFailed    = errors.New("operation failed")
	ErrOperationForbidden = errors.New("operation does not belong to this project")
)

// operationMaxWait caps how long a status poll is held server-side. It must stay well
// below proxy/VPN idle timeouts, or the long-poll recreates the problem it solves.
const operationMaxWait = 10 * time.Second

// resolveOperation checks that the ID belongs to the requesting project. The project is
// embedded in the ID, which is the only thing tying an operation to its owner.
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

func (s Service) describeOperation(ctx context.Context, operationID string, wait time.Duration) (temporal.OperationStatus, error) {
	status, err := s.temporal.DescribeOperation(ctx, operationID, wait)
	if err != nil {
		if temporal.IsOperationNotFound(err) {
			return "", fmt.Errorf("%w: %s", ErrOperationNotFound, operationID)
		}
		return "", fmt.Errorf("failed to describe operation: %s", err)
	}
	return status, nil
}

func (s Service) GetOperationStatus(ctx context.Context, projectID, operationID string, wait time.Duration) (*dto.OperationStatusResponse, error) {
	if s.temporal == nil {
		return nil, fmt.Errorf("temporal client not available")
	}
	if _, err := resolveOperation(projectID, operationID); err != nil {
		return nil, err
	}

	status, err := s.describeOperation(ctx, operationID, min(wait, operationMaxWait))
	if err != nil {
		return nil, err
	}

	resp := &dto.OperationStatusResponse{Status: string(status)}
	if status == temporal.OperationFailed {
		resp.Error = s.temporal.OperationFailureMessage(ctx, operationID)
	}
	return resp, nil
}

// GetOperationResult returns a finished operation's payload in the same shape the
// submitting endpoint used to return synchronously.
func (s Service) GetOperationResult(ctx context.Context, projectID, operationID string) (interface{}, error) {
	if s.temporal == nil {
		return nil, fmt.Errorf("temporal client not available")
	}
	kind, err := resolveOperation(projectID, operationID)
	if err != nil {
		return nil, err
	}

	status, err := s.describeOperation(ctx, operationID, 0)
	if err != nil {
		return nil, err
	}

	switch status {
	case temporal.OperationRunning:
		return nil, ErrOperationNotDone
	case temporal.OperationFailed:
		// The workflow itself did not complete (image pull, timeout). A connector that ran
		// and could not reach the database is instead a succeeded operation carrying a
		// FAILED connection status.
		return nil, fmt.Errorf("%w: %s", ErrOperationFailed, s.temporal.OperationFailureMessage(ctx, operationID))
	}

	// Discover writes several catalog files, so its result is read from them rather than from the
	// single file the workflow reports.
	if kind == temporal.OperationDiscoverCatalog {
		return discoverCatalogResult(ctx, operationID)
	}

	output, err := s.temporal.FetchOperationOutput(ctx, operationID)
	if err != nil {
		return nil, fmt.Errorf("failed to read operation result: %s", err)
	}

	return decodeOperationOutput(ctx, kind, operationID, output)
}

func decodeOperationOutput(ctx context.Context, kind temporal.OperationKind, operationID string, output map[string]interface{}) (interface{}, error) {
	switch kind {
	case temporal.OperationSpec:
		return dto.SpecResponse{Spec: output}, nil

	case temporal.OperationTestConnection:
		result, err := temporal.DecodeConnectionStatus(output)
		if err != nil {
			return nil, err
		}
		// The operation directory is named after the operation ID.
		logs, err := utils.ReadLogs(ctx, filepath.Join(constants.DefaultConfigDir, operationID), -1, -1, "older")
		if err != nil {
			return nil, fmt.Errorf("failed to read logs for operation %s: %s", operationID, err)
		}
		return dto.TestConnectionResponse{ConnectionResult: dto.NewConnectionResult(result), Logs: logs.Logs}, nil

	case temporal.OperationStreamDifference:
		return dto.StreamDifferenceResponse{DifferenceStreams: output}, nil

	default:
		return nil, fmt.Errorf("unsupported operation kind %q", kind)
	}
}
