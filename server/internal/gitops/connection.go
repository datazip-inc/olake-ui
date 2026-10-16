package gitops

import (
	"context"
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/services/etl"
	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
)

// connectionSucceeded is the status the CLI check command reports for a working connection.
const connectionSucceeded = "SUCCEEDED"

func testSourceConnection(ctx context.Context, etlSvc *etl.Service, projectID, sourceType, version string, config dto.JSONConfig) error {
	operationID, err := etlSvc.TestSourceConnection(ctx, projectID, &dto.SourceTestConnectionRequest{
		Type:    sourceType,
		Version: version,
		Config:  config,
	})
	result, err := awaitConnectionStatus(ctx, etlSvc, operationID, err)
	if err := connectionError(result, err); err != nil {
		return workflowError(fmt.Errorf("source connection test failed: %w", err))
	}
	return nil
}

func testDestinationConnection(ctx context.Context, etlSvc *etl.Service, projectID, destType, version string, config dto.JSONConfig, sourceType, sourceVersion string) error {
	operationID, err := etlSvc.TestDestinationConnection(ctx, projectID, &dto.DestinationTestConnectionRequest{
		Type:          destType,
		Version:       version,
		Config:        config,
		SourceType:    sourceType,
		SourceVersion: sourceVersion,
	})
	result, err := awaitConnectionStatus(ctx, etlSvc, operationID, err)
	if err := connectionError(result, err); err != nil {
		return workflowError(fmt.Errorf("destination connection test failed: %w", err))
	}
	return nil
}

// awaitConnectionStatus waits for a started connection test and returns its {message, status}.
func awaitConnectionStatus(ctx context.Context, etlSvc *etl.Service, operationID string, startErr error) (map[string]interface{}, error) {
	if startErr != nil {
		return nil, fmt.Errorf("connection test failed: %w", startErr)
	}
	output, err := etlSvc.AwaitOperationOutput(ctx, operationID)
	if err != nil {
		return nil, fmt.Errorf("connection test failed: %w", err)
	}
	return temporal.DecodeConnectionStatus(output)
}

// connectionError returns the test's error, or an error when the check ran but reported the
// connection as failed: the test returns no error in that case, only a FAILED status.
func connectionError(result map[string]interface{}, err error) error {
	if err != nil {
		return err
	}
	if status, _ := result["status"].(string); status != connectionSucceeded {
		message, _ := result["message"].(string)
		return fmt.Errorf("status %q: %s", status, message)
	}
	return nil
}
