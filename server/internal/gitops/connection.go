package gitops

import (
	"context"
	"fmt"
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/services/etl"
)

func testSourceConnection(ctx context.Context, etlSvc *etl.Service, sourceType, version string, config dto.JSONConfig) error {
	result, _, err := etlSvc.TestSourceConnection(ctx, &dto.SourceTestConnectionRequest{
		Type:    sourceType,
		Version: version,
		Config:  config,
	})
	if err := connectionTestError("source", result, err); err != nil {
		return NonRetryableError(err)
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
	if err := connectionTestError("destination", result, err); err != nil {
		return NonRetryableError(err)
	}
	return nil
}

// test connection returns nil error even if the connection test failed
// so check the status and message to determine if the connection test failed
func connectionTestError(kind string, result map[string]interface{}, err error) error {
	if err != nil {
		return fmt.Errorf("%s connection test failed: %w", kind, err)
	}
	status, _ := result["status"].(string)
	if strings.EqualFold(status, "SUCCEEDED") {
		return nil
	}
	msg, _ := result["message"].(string)
	if msg == "" {
		msg = "status=" + status
	}
	return fmt.Errorf("%s connection test failed: %s", kind, msg)
}
