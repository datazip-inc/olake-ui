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
	runsDefaultPageSize = 20
	runsFetchPageSize   = 200
	// runsFetchMax bounds how many runs are pulled for the in-memory optimization filter.
	runsFetchMax = 5000
)

// GetTableRuns lists a table's optimizing processes (compaction and snapshot expiry), newest first.
func (s *Service) GetTableRuns(ctx context.Context, query dto.TableRunsQuery) (*dto.TableRunsResponse, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = runsDefaultPageSize
	}

	switch query.Kind {
	case constants.RunKindAll, "":
		return s.fetchRunsPage(ctx, query, "")
	case constants.RunKindExpiry:
		return s.fetchRunsPage(ctx, query, constants.OptProcessTypeExpiry)
	case constants.RunKindOptimization:
		return s.fetchOptimizationRuns(ctx, query)
	default:
		return nil, &HTTPError{StatusCode: http.StatusBadRequest, Body: []byte(fmt.Sprintf("unknown run type %q", query.Kind))}
	}
}

// fetchRunsPage forwards one page to Fusion as-is, narrowed to processType when it is set.
func (s *Service) fetchRunsPage(ctx context.Context, query dto.TableRunsQuery, processType string) (*dto.TableRunsResponse, error) {
	path := fmt.Sprintf(constants.OptPathTableOptimizingProcesses, query.Catalog, query.Database, query.Table)
	params := url.Values{}
	params.Set("page", strconv.Itoa(query.Page))
	params.Set("pageSize", strconv.Itoa(query.PageSize))
	if processType != "" {
		params.Set("type", processType)
	}
	if query.Status != "" {
		params.Set("status", query.Status)
	}

	var result dto.TableRunsResponse
	if err := s.DoInto(ctx, http.MethodGet, path, params, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get runs for %s.%s.%s: %w", query.Catalog, query.Database, query.Table, err)
	}
	return &result, nil
}

// fetchOptimizationRuns collects every non-expiry run and paginates in memory.
func (s *Service) fetchOptimizationRuns(ctx context.Context, query dto.TableRunsQuery) (*dto.TableRunsResponse, error) {
	kept := make([]json.RawMessage, 0, runsFetchPageSize)
	chunkQuery := query
	chunkQuery.PageSize = runsFetchPageSize
	for fusionPage, fetched := 1, 0; fetched < runsFetchMax; fusionPage++ {
		chunkQuery.Page = fusionPage
		chunk, err := s.fetchRunsPage(ctx, chunkQuery, "")
		if err != nil {
			return nil, err
		}
		for _, raw := range chunk.List {
			var run struct {
				OptimizingType string `json:"optimizingType"`
			}
			if err := json.Unmarshal(raw, &run); err != nil {
				return nil, fmt.Errorf("failed to parse run for %s.%s.%s: %s", query.Catalog, query.Database, query.Table, err)
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

	start := (query.Page - 1) * query.PageSize
	if start > len(kept) {
		start = len(kept)
	}
	end := start + query.PageSize
	if end > len(kept) {
		end = len(kept)
	}
	return &dto.TableRunsResponse{List: kept[start:end], Total: len(kept)}, nil
}
