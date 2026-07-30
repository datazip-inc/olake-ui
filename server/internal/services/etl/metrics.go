package etl

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	enumspb "go.temporal.io/api/enums/v1"
)

const (
	// metricsRefreshTTL caps the DB + Temporal reads regardless of scrape interval.
	metricsRefreshTTL = 10 * time.Second
	// statsRefreshTTL re-reads stats.json between snapshots at the CLI's rewrite cadence.
	statsRefreshTTL = 2 * time.Second
	// metricsRefreshTimeout bails a hung rebuild while staying under Prometheus' scrape timeout.
	metricsRefreshTimeout = 5 * time.Second
)

var metricsLabelNames = []string{"job_id", "job_name", "source_name", "destination_name"}

// metricsCollector rebuilds the olake_* series per scrape from three sources —
// jobs table (labels), Temporal visibility (status/start/duration) and stats.json
// (records/bytes/cpu/memory) — in two tiers: a full snapshot at most every 10s and
// a stats.json-only re-read at most every 2s.
type metricsCollector struct {
	mu            sync.Mutex
	lastRefresh   time.Time
	lastStatsRead time.Time

	db           *database.Database
	tempClient   *temporal.Temporal
	statsBaseDir string

	handler http.Handler

	status    *prometheus.GaugeVec
	startTime *prometheus.GaugeVec
	duration  *prometheus.GaugeVec
	records   *prometheus.GaugeVec
	bytes     *prometheus.GaugeVec
	cpu       *prometheus.GaugeVec
	memory    *prometheus.GaugeVec

	scrapeErrors prometheus.Counter

	// statsTargets maps job_id → its latest run's stats.json location and labels.
	statsTargets map[string]statsTarget
}

type statsTarget struct {
	workflowID string
	labels     prometheus.Labels
}

func newMetricsCollector(db *database.Database, tempClient *temporal.Temporal, statsBaseDir string) *metricsCollector {
	newVec := func(name, help string) *prometheus.GaugeVec {
		return prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: name, Help: help}, metricsLabelNames)
	}

	m := &metricsCollector{
		db:           db,
		tempClient:   tempClient,
		statsBaseDir: statsBaseDir,
		status:       newVec("olake_sync_status", "Status of the most recent sync run: 0=running, 1=succeeded, 2=failed."),
		// start_time/duration come from Temporal, not stats.json's process clock —
		// that clock resets on retry and freezes if the process dies, so stuck syncs
		// would never trip a duration alert.
		startTime: newVec("olake_sync_start_time_seconds", "Unix timestamp (seconds) when the most recent sync run started."),
		duration:  newVec("olake_sync_duration_seconds", "Duration of the most recent sync run in seconds. Live while running; frozen at completion."),
		records:   newVec("olake_sync_records_ingested", "Records ingested in the most recent sync run (rollback-corrected; committed-only at completion)."),
		bytes:     newVec("olake_sync_bytes_read", "Bytes read from the source in the most recent sync run."),
		cpu:       newVec("olake_process_cpu_usage_ratio", "CPU utilization ratio (0-1) self-reported by the sync process via stats.json."),
		memory:    newVec("olake_process_memory_usage_bytes", "Memory in bytes self-reported by the sync process via stats.json."),
		scrapeErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "olake_metrics_scrape_errors_total",
			Help: "Total number of metrics snapshot rebuild failures; the previous snapshot is served instead.",
		}),
		statsTargets: map[string]statsTarget{},
	}

	// Own registry (not the global default) so the exposition holds only these series.
	// scrapeErrors counts snapshot rebuild failures: on a Postgres or Temporal read error
	// the gauges below keep their previous values and the scrape still returns 200.
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		m.status, m.startTime, m.duration, m.records, m.bytes, m.cpu, m.memory,
		m.scrapeErrors,
	)
	m.handler = promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	return m
}

// syncStatusValue maps a Temporal status to the gauge: open runs (including retry
// backoff and paused) → 0 so status never flaps; completed → 1; everything else
// (failed/timed-out/terminated/canceled) → 2.
func syncStatusValue(status enumspb.WorkflowExecutionStatus) float64 {
	switch status {
	case enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING,
		enumspb.WORKFLOW_EXECUTION_STATUS_PAUSED:
		return 0
	case enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		return 1
	default:
		return 2
	}
}

// metricsRow stages one job's computed values so the gauges are mutated only after
// every backend read has succeeded.
type metricsRow struct {
	labels     prometheus.Labels
	workflowID string
	status     float64
	start      float64
	duration   float64
	stats      *syncStats // nil when stats.json is unreadable this snapshot
}

// refresh runs on the scrape path: a full snapshot at most every metricsRefreshTTL,
// a stats.json-only re-read every statsRefreshTTL in between. The mutex serialises any
// overlapping scrapes; on error the previous snapshot is kept and the error counted.
func (m *metricsCollector) refresh(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if time.Since(m.lastRefresh) >= metricsRefreshTTL {
		rows, err := m.gather(ctx)

		// Stamp lastRefresh even on failure: a sick backend is retried once per TTL,
		// not once per scrape. lastStatsRead is left so the 2s stats tier keeps running.
		m.lastRefresh = time.Now()
		if err != nil {
			m.scrapeErrors.Inc()
			return err
		}

		m.commit(rows)
		m.lastStatsRead = time.Now()
		return nil
	}

	if time.Since(m.lastStatsRead) >= statsRefreshTTL {
		m.refreshStats()
		m.lastStatsRead = time.Now()
	}
	return nil
}

func (m *metricsCollector) gather(ctx context.Context) ([]metricsRow, error) {
	projectIDs, err := m.db.ListDistinctProjectIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list project ids: %s", err)
	}

	rows := make([]metricsRow, 0)
	for _, projectID := range projectIDs {
		jobs, err := m.db.ListJobsByProjectID(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("failed to list jobs project_id[%s]: %s", projectID, err)
		}

		// Sync runs only — a clear-destination run must never surface as a job's latest sync.
		lastRuns, err := fetchLatestJobRunsByJobIDs(ctx, m.tempClient, projectID, jobs, temporal.Sync)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch latest sync runs project_id[%s]: %s", projectID, err)
		}

		for _, job := range jobs {
			run, ok := lastRuns[job.ID]
			if !ok {
				continue // no run within Temporal retention → no series (absent = "no recent run")
			}

			var sourceName, destinationName string
			if job.Source != nil {
				sourceName = job.Source.Name
			}
			if job.Destination != nil {
				destinationName = job.Destination.Name
			}

			duration := time.Since(run.StartTime).Seconds()
			if run.CloseTime != nil {
				duration = run.CloseTime.Sub(run.StartTime).Seconds()
			}

			stats, err := readSyncStats(m.statsBaseDir, run.WorkflowID)
			if err != nil {
				// File not written yet or truncate-rewrite race: keep this job's previous
				// stats series; it self-heals on the next read.
				logger.Debugf("metrics: stats.json unavailable for job %d workflow %s: %s", job.ID, run.WorkflowID, err)
				stats = nil
			}

			rows = append(rows, metricsRow{
				labels: prometheus.Labels{
					"job_id":           strconv.Itoa(job.ID),
					"job_name":         job.Name,
					"source_name":      sourceName,
					"destination_name": destinationName,
				},
				workflowID: run.WorkflowID,
				status:     syncStatusValue(run.Status),
				start:      float64(run.StartTime.Unix()),
				duration:   duration,
				stats:      stats,
			})
		}
	}

	return rows, nil
}

// commit swaps the exposed series to the staged rows. status/start/duration are
// Reset so deleted jobs drop out; the stats gauges are updated per job (not Reset)
// so a job with unreadable stats.json keeps its last values.
func (m *metricsCollector) commit(rows []metricsRow) {
	m.status.Reset()
	m.startTime.Reset()
	m.duration.Reset()

	live := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		m.status.With(row.labels).Set(row.status)
		m.startTime.With(row.labels).Set(row.start)
		m.duration.With(row.labels).Set(row.duration)

		jobID := row.labels["job_id"]
		live[jobID] = struct{}{}
		m.statsTargets[jobID] = statsTarget{workflowID: row.workflowID, labels: row.labels}

		if row.stats != nil {
			m.setStats(row.labels, row.stats)
		}
	}

	// Prune stats series of jobs gone from the DB or past Temporal retention.
	for jobID := range m.statsTargets {
		if _, ok := live[jobID]; ok {
			continue
		}
		m.deleteStats(jobID)
		delete(m.statsTargets, jobID)
	}
}

// deleteStats removes a job's four stats series when it leaves the snapshot.
func (m *metricsCollector) deleteStats(jobID string) {
	match := prometheus.Labels{"job_id": jobID}
	m.records.DeletePartialMatch(match)
	m.bytes.DeletePartialMatch(match)
	m.cpu.DeletePartialMatch(match)
	m.memory.DeletePartialMatch(match)
}

func (m *metricsCollector) setStats(labels prometheus.Labels, stats *syncStats) {
	m.records.With(labels).Set(stats.Records)
	m.bytes.With(labels).Set(stats.Bytes)
	m.cpu.With(labels).Set(stats.CPURatio)
	m.memory.With(labels).Set(stats.Memory)
}

// refreshStats re-reads stats.json for the last snapshot's jobs — filesystem only,
// no DB or Temporal load. Caller holds m.mu.
func (m *metricsCollector) refreshStats() {
	for _, target := range m.statsTargets {
		stats, err := readSyncStats(m.statsBaseDir, target.workflowID)
		if err != nil {
			continue // unreadable → keep previous values; self-heals next read
		}
		m.setStats(target.labels, stats)
	}
}

// syncStats holds the pass-through values read from a sync run's stats.json.
// Key contract with the sync CLI: "Synced Records" (number), "Bytes Read"
// (number), "CPU Utilization" (number 0-1), "Memory Usage Bytes" (number).
type syncStats struct {
	Records  float64
	Bytes    float64
	CPURatio float64
	Memory   float64
}

// readSyncStats parses a sync run's stats.json. Missing or unparsable keys yield
// zero values without error (older images ship fewer keys); a missing file or
// truncated JSON (the CLI truncate-rewrites every ~2s) errors so the caller keeps
// the job's previous values.
func readSyncStats(baseDir, workflowID string) (*syncStats, error) {
	statsPath := filepath.Join(baseDir, temporal.GetWorkflowDirectory(temporal.Sync, workflowID), "stats.json")
	raw, err := utils.ReadJSONFile(statsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %s", statsPath, err)
	}

	stats := &syncStats{}
	if v, ok := raw["Synced Records"].(float64); ok {
		stats.Records = v
	}
	if v, ok := raw["Bytes Read"].(float64); ok {
		stats.Bytes = v
	}
	if v, ok := raw["CPU Utilization"].(float64); ok {
		stats.CPURatio = v
	}
	if v, ok := raw["Memory Usage Bytes"].(float64); ok {
		stats.Memory = v
	}

	return stats, nil
}

// RefreshMetrics rebuilds the snapshot (TTL-guarded, bounded by the refresh timeout)
// and returns the Prometheus exposition handler for the caller to serve. On a backend
// failure it returns the previous snapshot's handler together with the error.
func (s Service) RefreshMetrics(ctx context.Context) (http.Handler, error) {
	ctx, cancel := context.WithTimeout(ctx, metricsRefreshTimeout)
	defer cancel()
	return s.metrics.handler, s.metrics.refresh(ctx)
}
