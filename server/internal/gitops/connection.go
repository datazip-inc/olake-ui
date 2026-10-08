package gitops

import (
	"context"
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/services/etl"
)

// connectionSucceeded is the status the CLI check command reports for a working connection.
const connectionSucceeded = "SUCCEEDED"

func testSourceConnection(ctx context.Context, etlSvc *etl.Service, sourceType, version string, config dto.JSONConfig) error {
	result, _, err := etlSvc.TestSourceConnection(ctx, &dto.SourceTestConnectionRequest{
		Type:    sourceType,
		Version: version,
		Config:  config,
	})
	if err := connectionError(result, err); err != nil {
		return workflowError(fmt.Errorf("source connection test failed: %w", err))
	}
	return nil
}

func testDestinationConnection(ctx context.Context, etlSvc *etl.Service, destType, version string, config dto.JSONConfig, sourceType, sourceVersion string) error {
	result, _, err := etlSvc.TestDestinationConnection(ctx, &dto.DestinationTestConnectionRequest{
		Type:          destType,
		Version:       version,
		Config:        config,
		SourceType:    sourceType,
		SourceVersion: sourceVersion,
	})
	if err := connectionError(result, err); err != nil {
		return workflowError(fmt.Errorf("destination connection test failed: %w", err))
	}
	return nil
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
