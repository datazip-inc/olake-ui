package gitops

import (
	"context"
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/services/etl"
)

func testSourceConnection(ctx context.Context, etlSvc *etl.Service, sourceType, version string, config dto.JSONConfig) error {
	_, _, err := etlSvc.TestSourceConnection(ctx, &dto.SourceTestConnectionRequest{
		Type:    sourceType,
		Version: version,
		Config:  config,
	})
	if err != nil {
		return NonRetryableError(fmt.Errorf("source connection test failed: %w", err))
	}
	return nil
}

func testDestinationConnection(ctx context.Context, etlSvc *etl.Service, destType, version string, config dto.JSONConfig, sourceType, sourceVersion string) error {
	_, _, err := etlSvc.TestDestinationConnection(ctx, &dto.DestinationTestConnectionRequest{
		Type:          destType,
		Version:       version,
		Config:        config,
		SourceType:    sourceType,
		SourceVersion: sourceVersion,
	})
	if err != nil {
		return NonRetryableError(fmt.Errorf("destination connection test failed: %w", err))
	}
	return nil
}
