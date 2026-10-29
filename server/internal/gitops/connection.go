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
	if err == nil {
		err = awaitConnection(ctx, etlSvc, operationID)
	}
	if err != nil {
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
	if err == nil {
		err = awaitConnection(ctx, etlSvc, operationID)
	}
	if err != nil {
		return workflowError(fmt.Errorf("destination connection test failed: %w", err))
	}
	return nil
}

// awaitConnection waits for a started connection test. It returns an error when the test did not
// run, or when it ran and reported the connection as failed: the check returns no error in that
// case, only a FAILED status.
func awaitConnection(ctx context.Context, etlSvc *etl.Service, operationID string) error {
	output, err := etlSvc.AwaitOperationOutput(ctx, operationID)
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
