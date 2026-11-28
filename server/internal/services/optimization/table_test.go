package optimization

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
)

// processListServer fakes Fusion's optimizing-processes endpoint. Each entry
// maps a status filter ("" = unfiltered) to the newest run of that kind.
type processListServer struct {
	mu       sync.Mutex
	runs     map[string][]map[string]interface{}
	statuses []string
}

func (p *processListServer) handler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		p.mu.Lock()
		p.statuses = append(p.statuses, status)
		p.mu.Unlock()

		list, ok := p.runs[status]
		if !ok {
			t.Errorf("unexpected status filter %q", status)
		}
		resp := map[string]interface{}{
			"code":    200,
			"message": "success",
			"result":  map[string]interface{}{"list": list},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func (p *processListServer) filteredStatuses() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]string, 0, len(p.statuses))
	for _, s := range p.statuses {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func run(processID, status string, finishTime int64) map[string]interface{} {
	return map[string]interface{}{
		"processId":  processID,
		"status":     status,
		"finishTime": float64(finishTime),
	}
}

func newProcessListService(t *testing.T, runs map[string][]map[string]interface{}) (*Service, *processListServer) {
	t.Helper()
	fake := &processListServer{runs: runs}
	srv := httptest.NewServer(fake.handler(t))
	t.Cleanup(srv.Close)
	return &Service{baseURL: srv.URL, client: srv.Client()}, fake
}

func TestFetchLatestProcessInfoReturnsLatestRealRun(t *testing.T) {
	tests := []struct {
		name           string
		runs           map[string][]map[string]interface{}
		wantStatus     string
		wantFinishTime int64
		wantSkipped    bool
		wantFilters    []string
	}{
		{
			name: "latest run is a real run, no extra requests",
			runs: map[string][]map[string]interface{}{
				"":       {run("p1", constants.OptProcessStatusSuccess, 300)},
				"SUCCESS": {},
				"FAILED":  {},
			},
			wantStatus:     constants.OptProcessStatusSuccess,
			wantFinishTime: 300,
			wantSkipped:    false,
			wantFilters:    nil,
		},
		{
			name: "skipped runs hide a success, success is reported with the skip kept",
			runs: map[string][]map[string]interface{}{
				"":       {run("p3", constants.OptProcessStatusSkipped, 300)},
				"SUCCESS": {run("p1", constants.OptProcessStatusSuccess, 100)},
				"FAILED":  {},
			},
			wantStatus:     constants.OptProcessStatusSuccess,
			wantFinishTime: 100,
			wantSkipped:    true,
			wantFilters:    []string{"SUCCESS", "FAILED"},
		},
		{
			name: "failed run newer than the last success wins",
			runs: map[string][]map[string]interface{}{
				"":       {run("p4", constants.OptProcessStatusSkipped, 400)},
				"SUCCESS": {run("p1", constants.OptProcessStatusSuccess, 100)},
				"FAILED":  {run("p2", constants.OptProcessStatusFailed, 200)},
			},
			wantStatus:     constants.OptProcessStatusFailed,
			wantFinishTime: 200,
			wantSkipped:    true,
			wantFilters:    []string{"SUCCESS", "FAILED"},
		},
		{
			name: "every run skipped stays skipped",
			runs: map[string][]map[string]interface{}{
				"":       {run("p2", constants.OptProcessStatusSkipped, 200)},
				"SUCCESS": {},
				"FAILED":  {},
			},
			wantStatus:     constants.OptProcessStatusSkipped,
			wantFinishTime: 200,
			wantSkipped:    true,
			wantFilters:    []string{"SUCCESS", "FAILED"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, fake := newProcessListService(t, tt.runs)

			info, err := svc.fetchLatestProcessInfo(context.Background(), "cat", "db", "tbl", constants.OptProcessTypeFull)
			if err != nil {
				t.Fatalf("fetchLatestProcessInfo: %v", err)
			}
			if info.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", info.Status, tt.wantStatus)
			}
			if info.FinishTime != tt.wantFinishTime {
				t.Errorf("finishTime = %d, want %d", info.FinishTime, tt.wantFinishTime)
			}
			if (info.SkippedRun != nil) != tt.wantSkipped {
				t.Errorf("skippedRun present = %v, want %v", info.SkippedRun != nil, tt.wantSkipped)
			}
			if tt.wantSkipped && info.SkippedRun.Status != constants.OptProcessStatusSkipped {
				t.Errorf("skippedRun.status = %q, want SKIPPED", info.SkippedRun.Status)
			}

			got := fake.filteredStatuses()
			if fmt.Sprint(got) != fmt.Sprint(tt.wantFilters) {
				t.Errorf("status-filtered requests = %v, want %v", got, tt.wantFilters)
			}
		})
	}
}
