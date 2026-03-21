package aggregator

import (
	"context"
	"fmt"
	"sort"

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
		if catalogType == "custom" {
			catalogType = "JDBC"
		}

		catalogData := models.CatalogWithDatabases{
			Name:      catalogName,
			Type:      catalogType,
			Databases: make([]string, 0),
		}

		properties, ok := catalogMap["catalogProperties"].(map[string]interface{})
		if ok {
			if createdAt, ok := properties["created-at"].(string); ok {
				catalogData.CreatedAt = createdAt
			}
			if olakeManaged, ok := properties["olake-created"].(string); ok {
				catalogData.OLakeCreated = olakeManaged == "true"
			}
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

		sort.Strings(catalogData.Databases)

		response.Catalogs = append(response.Catalogs, catalogData)
	}

	sort.Slice(response.Catalogs, func(i, j int) bool {
		return response.Catalogs[i].Name < response.Catalogs[j].Name
	})

	return response, nil
}

// GetTablesWithDetails fetches all tables with full details for a specific catalog and database
func (s *Service) GetTablesWithDetails(ctx context.Context, catalog, databaseName string, db *database.Database) (*models.TablesResponse, error) {
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
		return nil, fmt.Errorf("unexpected tables result format for %s.%s: got %T", catalog, databaseName, tablesResult)
	}

	for _, item := range tablesList {
		tableMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid table item type: got %T", item)
		}

		tableName, ok := tableMap["name"].(string)
		if !ok || tableName == "" {
			return nil, fmt.Errorf("missing or invalid table name in %v", tableMap)
		}

		tableInfo := models.TableInfo{
			Name: tableName,
			Enabled: false,
			ByOLake: false,
		}

		tableDetails, err := s.table.GetTableDetails(ctx, catalog, databaseName, tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to get details for table %s.%s.%s: %v\n", catalog, databaseName, tableName, err)
		}

		detailsMap, ok := tableDetails.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid tableDetails type: expected map[string]interface{}, got %T", tableDetails)
		}

		baseMetrics, ok := detailsMap["baseMetrics"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("missing or invalid baseMetrics for table %s.%s.%s", catalog, databaseName, tableName)
		}

		totalSize, ok := baseMetrics["totalSize"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid totalSize in baseMetrics for table %s.%s.%s", catalog, databaseName, tableName)
		}

		tableInfo.TotalSize = totalSize

		properties, ok := detailsMap["properties"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("missing or invalid properties for table %s.%s.%s", catalog, databaseName, tableName)
		}

		if enabled, ok := properties["self-optimizing.enabled"]; ok {
			tableInfo.Enabled = enabled.(string) == "true"
		}

		if _, ok := properties["olake.2pc"]; ok {
			tableInfo.ByOLake = true
		}

		tableSummary, ok := detailsMap["tableSummary"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("missing or invalid tableSummary for table %s.%s.%s", catalog, databaseName, tableName)
		}

		healthScore, ok := tableSummary["healthScore"].(float64)
		if !ok {
			return nil, fmt.Errorf("missing or invalid healthScore in tableSummary for table %s.%s.%s", catalog, databaseName, tableName)
		}

		tableInfo.HealthScore = int(healthScore)

		// fetch latest optimizing processes for each type
		res, err := s.fetchLatestProcessInfo(ctx, catalog, databaseName, tableName, "MINOR")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch latest minor process info: %s", err)
		}
		tableInfo.Minor = res

		res, err = s.fetchLatestProcessInfo(ctx, catalog, databaseName, tableName, "MAJOR")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch latest major process info: %s", err)
		}
		tableInfo.Major = res

		res, err = s.fetchLatestProcessInfo(ctx, catalog, databaseName, tableName, "FULL")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch latest full process info: %s", err)
		}
		tableInfo.Full = res

		response.Tables = append(response.Tables, tableInfo)
	}

	return response, nil
}
