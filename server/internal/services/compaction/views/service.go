package aggregator

import (
	"context"
	"fmt"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/resources/catalogs"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/client"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/models"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/resources/tables"
)

type Service struct {
	compaction *client.Compaction
}

func NewService(c *client.Compaction) *Service {
	return &Service{
		compaction: c,
	}
}

// GetCatalogsWithDatabases fetches all catalogs and their databases (without table details)
func (c *Service) GetCatalogsWithDatabases(ctx context.Context) (*models.CatalogsResponse, error) {
	// Step 1: Get all catalogs
	catalog := catalog.NewService(c.compaction)
	catalogsResult, err := catalog.GetCatalogs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalogs: %w", err)
	}

	// Parse catalogs result
	catalogsList, ok := catalogsResult.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected catalogs format, got type: %T", catalogsResult)
	}

	response := &models.CatalogsResponse{
		Catalogs: make([]models.CatalogWithDatabases, 0, len(catalogsList)),
	}

	// Step 2: For each catalog, get databases
	for _, catalogItem := range catalogsList {
		catalogMap, ok := catalogItem.(map[string]interface{})
		if !ok {
			continue
		}

		// Amoro uses "catalogName" and "catalogType" fields
		catalogName, ok := catalogMap["catalogName"].(string)
		if !ok {
			if nameVal := catalogMap["catalogName"]; nameVal != nil {
				catalogName = fmt.Sprintf("%v", nameVal)
			}
		}

		catalogType, ok := catalogMap["catalogType"].(string)
		if !ok {
			if typeVal := catalogMap["catalogType"]; typeVal != nil {
				catalogType = fmt.Sprintf("%v", typeVal)
			}
		}

		if catalogName == "" {
			continue
		}

		catalogData := models.CatalogWithDatabases{
			Name:      catalogName,
			Type:      catalogType,
			Databases: make([]string, 0),
		}

		// Get databases for this catalog
		tbl := table.NewService(c.compaction)
		databasesResult, err := tbl.GetDatabases(ctx, catalogName, "")
		if err != nil {
			fmt.Printf("Failed to get databases for catalog %s: %v\n", catalogName, err)
			response.Catalogs = append(response.Catalogs, catalogData)
			continue
		}

		databasesList, ok := databasesResult.([]interface{})
		if !ok {
			response.Catalogs = append(response.Catalogs, catalogData)
			continue
		}

		// Extract database names
		for _, dbItem := range databasesList {
			dbName, ok := dbItem.(string)
			if ok && dbName != "" {
				catalogData.Databases = append(catalogData.Databases, dbName)
			}
		}

		response.Catalogs = append(response.Catalogs, catalogData)
	}

	return response, nil
}

// GetTablesWithDetails fetches all tables with full details for a specific catalog and database
func (c *Service) GetTablesWithDetails(ctx context.Context, catalog string, databaseName string, db *database.Database) (*models.TablesResponse, error) {
	response := &models.TablesResponse{
		Catalog:  catalog,
		Database: databaseName,
		Tables:   make([]models.TableInfo, 0),
	}

	// Get tables for this database
	tbl := table.NewService(c.compaction)
	tablesResult, err := tbl.GetTables(ctx, catalog, databaseName, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get tables for catalog %s, database %s: %w", catalog, databaseName, err)
	}

	tablesList, ok := tablesResult.([]interface{})
	if !ok {
		return response, nil
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

			// Fetch table details to get totalSize
			tableDetails, err := tbl.GetTableDetails(ctx, catalog, databaseName, tableName)
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

			// Fetch optimizing processes to get MAJOR, MINOR, FULL optimization data
			optimizingProcesses, err := tbl.GetOptimizingProcesses(ctx, catalog, databaseName, tableName)
			if err != nil {
				fmt.Printf("Failed to get optimizing processes for table %s.%s.%s: %v\n", catalog, databaseName, tableName, err)
			} else {
				// Parse optimization processes
				if processesMap, ok := optimizingProcesses.(map[string]interface{}); ok {
					if processList, ok := processesMap["list"].([]interface{}); ok {
						// Track latest process for each type
						latestProcesses := make(map[string]map[string]interface{})

						// Iterate through all processes to find the latest for each type
						for _, processItem := range processList {
							if process, ok := processItem.(map[string]interface{}); ok {
								optimizingType, _ := process["optimizingType"].(string)
								finishTime, _ := process["finishTime"].(float64)

								if optimizingType != "" {
									// Check if this is the latest process for this type
									if existing, exists := latestProcesses[optimizingType]; !exists {
										latestProcesses[optimizingType] = process
									} else {
										existingFinishTime, _ := existing["finishTime"].(float64)
										if finishTime > existingFinishTime {
											latestProcesses[optimizingType] = process
										}
									}
								}
							}
						}

						// Extract data for MAJOR, MINOR, FULL
						for processType, process := range latestProcesses {
							finishTime, _ := process["finishTime"].(float64)
							status, _ := process["status"].(string)

							// Convert timestamp to relative time format
							var lastRun string
							if finishTime > 0 {
								timestamp := time.Unix(0, int64(finishTime)*int64(time.Millisecond))
								duration := time.Since(timestamp)

								if duration < time.Minute {
									seconds := int(duration.Seconds())
									if seconds == 1 {
										lastRun = "1 sec ago"
									} else {
										lastRun = fmt.Sprintf("%d secs ago", seconds)
									}
								} else if duration < time.Hour {
									minutes := int(duration.Minutes())
									if minutes == 1 {
										lastRun = "1 minute ago"
									} else {
										lastRun = fmt.Sprintf("%d minutes ago", minutes)
									}
								} else {
									hours := int(duration.Hours())
									if hours == 1 {
										lastRun = "1 hour ago"
									} else {
										lastRun = fmt.Sprintf("%d hours ago", hours)
									}
								}
							}

							optimizationInfo := &models.OptimizationInfo{
								LastRun: lastRun,
								Status:  status,
							}

							switch processType {
							case "MAJOR":
								tableInfo.Major = optimizationInfo
							case "MINOR":
								tableInfo.Minor = optimizationInfo
							case "FULL":
								tableInfo.Full = optimizationInfo
							}
						}
					}
				}
			}

			response.Tables = append(response.Tables, tableInfo)
		}
	}

	return response, nil
}
