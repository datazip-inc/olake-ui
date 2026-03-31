package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

// after test destination suceeds, we create Opt Catalog first, then ETL Destination
func (s *AppService) CreateDestinationWithCatalog(ctx context.Context, projectID string, req *dto.CreateDestinationRequest, userID *int) error {
	var catalogName string
	var err error

	if s.opt != nil {
		catalogName, err = s.opt.CreateCatalog(ctx, extractWriterConfig(req.Config), true)
		if err != nil {
			return fmt.Errorf("failed to create optimization catalog: %s", err)
		}
	}

	if err := s.etl.CreateDestination(ctx, req, projectID, userID); err != nil {
		return fmt.Errorf("failed to create destination: %s", err)
	}

	logger.Infof("Destination[%s] and Catalog[%s] created successfully", req.Name, catalogName)

	return nil
}

// after test destination suceeds, first update the Opt Catalog, then update the ETL destination
func (s *AppService) UpdateDestinationWithCatalog(ctx context.Context, id int, projectID string, req *dto.UpdateDestinationRequest, userID *int) error {
	var catalogName string
	var err error

	if s.opt != nil {
		catalogName, err = s.opt.UpdateCatalog(ctx, extractWriterConfig(req.Config))
		if err != nil {
			return fmt.Errorf("failed to update optimization catalog: %s", err)
		}
	}

	if err := s.etl.UpdateDestination(ctx, id, projectID, req, userID); err != nil {
		return fmt.Errorf("failed to update destination: %s", err)
	}

	logger.Infof("Destination[%s] and Catalog[%s] updated successfully", req.Name, catalogName)

	return nil
}

// first delete the catalog from Optimization, then delete the destination from ETL
func (s *AppService) DeleteDestinationWithCatalog(ctx context.Context, id int) (*dto.DeleteDestinationResponse, error) {
	var catalogName string
	var err error

	if s.opt != nil {
		destination, err := s.etl.GetDestinationByID(id)
		if err != nil {
			return nil, fmt.Errorf("failed to find destination id: %s", err)
		}

		// Extract catalog name from config
		catalogReq, err := s.opt.CreateOptConfig(extractWriterConfig(destination.Config), false)
		if err != nil {
			return nil, fmt.Errorf("failed to parse catalog config: %s", err)
		}

		catalogName, err = s.opt.DeleteCatalogInOpt(ctx, catalogReq.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to delete catalog in optimization: %s", err)
		}
	}

	resp, err := s.etl.DeleteDestination(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete destination: %s", err)
	}

	logger.Infof("Destination[%s] and Catalog[%s] deleted successfully", resp.Name, catalogName)

	return resp, nil
}

// extractWriterConfig unwraps the ETL destination config format {"type":"...","writer":{...}}
// returning just the inner writer JSON. If the config is already flat, it is returned unchanged.
func extractWriterConfig(configJSON string) string {
	var wrapped struct {
		Writer json.RawMessage `json:"writer"`
	}
	if err := json.Unmarshal([]byte(configJSON), &wrapped); err == nil && len(wrapped.Writer) > 0 {
		return string(wrapped.Writer)
	}
	return configJSON
}

func (s *AppService) SyncCatalogs(ctx context.Context) {
	destinations, err := s.db.ListDestinations()
	if err != nil {
		logger.Errorf("Failed to list destinations for optimization sync: %s", err)
		return
	}

	// Sort destinations by "CreatedAt" descending (newest first)
	sort.Slice(destinations, func(i, j int) bool {
		return destinations[i].CreatedAt.After(destinations[j].CreatedAt)
	})

	for _, dest := range destinations {
		if !strings.EqualFold(dest.DestType, "iceberg") {
			logger.Debugf("Skipping catalog creation for destination [%s] as the type is not iceberg", dest.Name)
			continue
		}

		catalogName, err := s.opt.CreateCatalog(ctx, extractWriterConfig(dest.Config), true)
		if err != nil {
			logger.Warnf("Failed to create catalog[%s] from destination[%s]: %s", catalogName, dest.Name, err)
		} else {
			logger.Infof("Catalog[%s] created successfully from destination[%s]", catalogName, dest.Name)
		}
	}
}
