package aggregator

import (
	"context"
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/client"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/models"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/resources/catalog"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/resources/table"
)

type Service struct {
	compaction *client.Compaction
	catalog    *catalog.Service
	table      *table.Service
}

func NewService(c *client.Compaction, cat *catalog.Service, tbl *table.Service) *Service {
	return &Service{
		compaction: c,
		catalog:    cat,
		table:      tbl,
	}
}

// fetches all catalogs and their respective databases
func (s *Service) GetCatalogsWithDatabases(ctx context.Context) (*models.CatalogsResponse, error) {
	catalogsResult, err := s.catalog.GetCatalogs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalogs: %s", err)
	}

	catalogsList, ok := catalogsResult.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected catalogs format, got type: %T", catalogsResult)
	}

	response := &models.CatalogsResponse{
		Catalogs: make([]models.CatalogWithDatabases, 0, len(catalogsList)),
	}

	// for each catalog, get the databases
	for _, catalogItem := range catalogsList {
		catalogMap, _ := catalogItem.(map[string]interface{})
		catalogName, _ := catalogMap["catalogName"].(string)
		catalogType, _ := catalogMap["catalogType"].(string)

		catalogData := models.CatalogWithDatabases{
			Name:      catalogName,
			Type:      catalogType,
			Databases: make([]string, 0),
		}

		databasesResult, err := s.table.GetDatabases(ctx, catalogName, "")
		if err != nil {
			fmt.Printf("failed to get databases for catalog %s: %v\n", catalogName, err)
			response.Catalogs = append(response.Catalogs, catalogData)
			continue
		}

		// Extract database names
		databasesList, _ := databasesResult.([]interface{})
		for _, dbItem := range databasesList {
			dbName := dbItem.(string)
			catalogData.Databases = append(catalogData.Databases, dbName)
		}

		response.Catalogs = append(response.Catalogs, catalogData)
	}

	return response, nil
}

// GetTablesWithDetails fetches all tables with full details for a specific catalog and database
func (s *Service) GetTablesWithDetails(ctx context.Context, catalog string, databaseName string, db *database.Database) (*models.TablesResponse, error) {
	response := &models.TablesResponse{
		Catalog:  catalog,
		Database: databaseName,
		Tables:   make([]models.TableInfo, 0),
	}

	tablesResult, err := s.table.GetTables(ctx, catalog, databaseName, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get tables for catalog %s, database %s: %w", catalog, databaseName, err)
	}

	tablesList, ok := tablesResult.([]interface{})
	if !ok {
		return response, nil
	}

	// Fetch catalog metadata once to get table properties
	var catalogMeta *models.CatalogRequest
	catalogMeta, err = s.getCatalogMetadata(ctx, catalog)
	if err != nil {
		fmt.Printf("Failed to get catalog metadata for %s: %v\n", catalog, err)
	}

	// Parse tables and fetch details for each
	for _, tableItem := range tablesList {
		tableMap, ok := tableItem.(map[string]interface{})
		if !ok {
			continue
		}

		tableName, _ := tableMap["name"].(string)

		if tableName != "" {
			tableInfo := models.TableInfo{
				Name:    tableName,
				ByOLake: false,
				Enabled: false,
			}

			// Check if table is managed by OLake
			if db != nil {
				isManagedByOLake, err := db.CheckTableManagedByOLake(catalog, databaseName, tableName)
				if err != nil {
					fmt.Printf("Failed to check if table %s.%s.%s is managed by OLake: %v\n", catalog, databaseName, tableName, err)
				} else {
					tableInfo.ByOLake = isManagedByOLake
				}
			}

			// Get enabled/disabled status from catalog table properties
			if catalogMeta != nil {
				tableInfo.Enabled = s.getTableEnabledStatus(catalogMeta, databaseName, tableName)
			}

			// Fetch table details to get totalSize
			tableDetails, err := s.table.GetTableDetails(ctx, catalog, databaseName, tableName)
			if err != nil {
				fmt.Printf("Failed to get details for table %s.%s.%s: %v\n", catalog, databaseName, tableName, err)
			} else {
				// Extract totalSize from baseMetrics
				if detailsMap, ok := tableDetails.(map[string]interface{}); ok {
					if baseMetrics, ok := detailsMap["baseMetrics"].(map[string]interface{}); ok {
						if totalSize, ok := baseMetrics["totalSize"].(string); ok {
							tableInfo.TotalSize = totalSize
						}
					}
				}
			}

			// Fetch latest optimizing processes for each type using API parameters
			tableInfo.Minor = s.fetchLatestProcessInfo(ctx, catalog, databaseName, tableName, "MINOR")
			tableInfo.Major = s.fetchLatestProcessInfo(ctx, catalog, databaseName, tableName, "MAJOR")
			tableInfo.Full = s.fetchLatestProcessInfo(ctx, catalog, databaseName, tableName, "FULL")

			response.Tables = append(response.Tables, tableInfo)
		}
	}

	return response, nil
}
