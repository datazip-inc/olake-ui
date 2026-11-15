package reconciler

import (
	"context"
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
)

// connectionSucceeded is the status the CLI check command reports for a working connection.
const connectionSucceeded = "SUCCEEDED"

// connectionService is the part of the ETL service that runs connection tests.
type connectionService interface {
	TestSourceConnection(ctx context.Context, projectID string, req *dto.SourceTestConnectionRequest) (string, error)
	TestDestinationConnection(ctx context.Context, projectID string, req *dto.DestinationTestConnectionRequest) (string, error)
	AwaitOperationOutput(ctx context.Context, operationID string) (map[string]interface{}, error)
}

func testSourceConnection(ctx context.Context, svc connectionService, projectID, sourceType, version string, config dto.JSONConfig) error {
	operationID, err := svc.TestSourceConnection(ctx, projectID, &dto.SourceTestConnectionRequest{
		Type:    sourceType,
		Version: version,
		Config:  config,
	})
	if err == nil {
		err = awaitConnection(ctx, svc, operationID)
	}
	if err != nil {
		return workflowError(fmt.Errorf("source connection test failed: %w", err))
	}
	return nil
}

func testDestinationConnection(ctx context.Context, svc connectionService, projectID, destType, version string, config dto.JSONConfig, sourceType, sourceVersion string) error {
	operationID, err := svc.TestDestinationConnection(ctx, projectID, &dto.DestinationTestConnectionRequest{
		Type:          destType,
		Version:       version,
		Config:        config,
		SourceType:    sourceType,
		SourceVersion: sourceVersion,
	})
	if err == nil {
		err = awaitConnection(ctx, svc, operationID)
	}
	if err != nil {
		return workflowError(fmt.Errorf("destination connection test failed: %w", err))
	}
	return nil
}

// awaitConnection waits for a started connection test. It returns an error when the test did not
// run, or when it ran and reported the connection as failed: the check returns no error in that
// case, only a FAILED status.
func awaitConnection(ctx context.Context, svc connectionService, operationID string) error {
	output, err := svc.AwaitOperationOutput(ctx, operationID)
	if err != nil {
		return err
	}
	result, err := temporal.DecodeConnectionStatus(output)
	if err != nil {
		return err
	}
	if status, _ := result["status"].(string); status != connectionSucceeded {
		message, _ := result["message"].(string)
		return fmt.Errorf("status %q: %s", status, message)
	}
	return nil
}
