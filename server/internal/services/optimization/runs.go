package optimization

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

// Fusion pages compaction and snapshot-expiry runs together and filters by one process type at
// a time, so "everything except expiry" has to be assembled here: pull the runs page by page,
// drop the expiry ones and paginate what is left.
const (
	runsFetchPageSize = 200
	// runsFetchMax bounds how many runs are pulled for the in-memory optimization filter.
	runsFetchMax = 5000
)

// GetTableRuns lists a table's optimizing processes (compaction and snapshot expiry), newest first.
// kind is one of constants.RunKindAll, RunKindExpiry or RunKindOptimization.
func (s *Service) GetTableRuns(ctx context.Context, catalog, database, table, kind, status string, page, pageSize int) (*dto.TableRunsResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	switch kind {
	case constants.RunKindAll, "":
		return s.fetchRunsPage(ctx, catalog, database, table, "", status, page, pageSize)
	case constants.RunKindExpiry:
		return s.fetchRunsPage(ctx, catalog, database, table, constants.OptProcessTypeExpiry, status, page, pageSize)
	case constants.RunKindOptimization:
		return s.fetchOptimizationRuns(ctx, catalog, database, table, status, page, pageSize)
	default:
		return nil, &HTTPError{StatusCode: http.StatusBadRequest, Body: []byte(fmt.Sprintf("unknown run type %q", kind))}
	}
}

// fetchRunsPage forwards one page to Fusion as-is.
func (s *Service) fetchRunsPage(ctx context.Context, catalog, database, table, processType, status string, page, pageSize int) (*dto.TableRunsResponse, error) {
	path := fmt.Sprintf(constants.OptPathTableOptimizingProcesses, catalog, database, table)
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	params.Set("pageSize", strconv.Itoa(pageSize))
	if processType != "" {
		params.Set("type", processType)
	}
	if status != "" {
		params.Set("status", status)
	}

	var result dto.TableRunsResponse
	if err := s.DoInto(ctx, http.MethodGet, path, params, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get runs for %s.%s.%s: %w", catalog, database, table, err)
	}
	return &result, nil
}

// fetchOptimizationRuns collects every non-expiry run and paginates in memory.
func (s *Service) fetchOptimizationRuns(ctx context.Context, catalog, database, table, status string, page, pageSize int) (*dto.TableRunsResponse, error) {
	kept := make([]json.RawMessage, 0, pageSize)
	for fusionPage, fetched := 1, 0; fetched < runsFetchMax; fusionPage++ {
		chunk, err := s.fetchRunsPage(ctx, catalog, database, table, "", status, fusionPage, runsFetchPageSize)
		if err != nil {
			return nil, err
		}
		for _, raw := range chunk.List {
			var run struct {
				OptimizingType string `json:"optimizingType"`
			}
			if err := json.Unmarshal(raw, &run); err != nil {
				return nil, fmt.Errorf("failed to parse run for %s.%s.%s: %s", catalog, database, table, err)
			}
			if run.OptimizingType != constants.OptProcessTypeExpiry {
				kept = append(kept, raw)
			}
		}
		fetched += len(chunk.List)
		if len(chunk.List) < runsFetchPageSize || fetched >= chunk.Total {
			break
		}
	}

	start := (page - 1) * pageSize
	if start > len(kept) {
		start = len(kept)
	}
	end := start + pageSize
	if end > len(kept) {
		end = len(kept)
	}
	return &dto.TableRunsResponse{List: kept[start:end], Total: len(kept)}, nil
}
