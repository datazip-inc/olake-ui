package etl

import (
	"context"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
)

// AppService is a unified service exposing all domain operations backed by shared deps.
type Service struct {
	// single ORM facade using one Ormer
	db       *database.Database
	temporal *temporal.Temporal
	metrics  *metricsCollector
}

// InitAppService constructs a unified AppService with singletons.
func InitService(db *database.Database) (*Service, error) {
	client, err := temporal.NewClient()
	if err != nil {
		return nil, err
	}

	metrics := newMetricsCollector(
		db.ListDistinctProjectIDs,
		db.ListJobsByProjectID,
		func(ctx context.Context, projectID string, jobs []*models.Job) (map[int]JobLastRunInfo, error) {
			// Metrics consider sync runs only — a clear-destination run must not surface as a job's latest sync.
			return fetchLatestJobRunsByJobIDs(ctx, client, projectID, jobs, temporal.Sync)
		},
		constants.DefaultConfigDir,
	)

	return &Service{
		db:       db,
		temporal: client,
		metrics:  metrics,
	}, nil
}
