package etl

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	enumspb "go.temporal.io/api/enums/v1"
)

const (
	// metricsRefreshTTL caps DB/Temporal reads regardless of scrape interval.
	metricsRefreshTTL = 10 * time.Second
	// statsRefreshTTL matches the sync CLI's stats.json rewrite cadence: scrapes
	// between full snapshots re-read the file once it is 2s stale.
	statsRefreshTTL = 2 * time.Second
	// metricsRefreshTimeout bounds a single snapshot rebuild.
	metricsRefreshTimeout = 15 * time.Second
)

var metricsLabelNames = []string{"job_id", "job_name", "source_name", "destination_name"}

// metricsCollector rebuilds every olake_* series on demand from three durable
// sources: the jobs table (labels), Temporal visibility (status/start/duration)
// and stats.json on the shared volume (records/bytes/cpu/memory). All work runs
// on the scrape path in two freshness tiers: a full snapshot (DB + Temporal) at
// most every 10s, and a filesystem-only stats.json re-read at most every 2s —
// the file's own rewrite cadence.
type metricsCollector struct {
	mu            sync.Mutex
	lastRefresh   time.Time
	lastStatsRead time.Time

	// Data-source seams, wired to the real DB/Temporal helpers in
	// newMetricsCollector and to fakes in tests.
	listProjectIDs func() ([]string, error)
	listJobs       func(projectID string) ([]*models.Job, error)
	fetchLastRuns  func(ctx context.Context, projectID string, jobs []*models.Job) (map[int]JobLastRunInfo, error)
	statsBaseDir   string
	now            func() time.Time

	registry *prometheus.Registry
	handler  http.Handler

	status    *prometheus.GaugeVec
	startTime *prometheus.GaugeVec
	duration  *prometheus.GaugeVec
	records   *prometheus.GaugeVec
	bytes     *prometheus.GaugeVec
	cpu       *prometheus.GaugeVec
	memory    *prometheus.GaugeVec

	scrapeErrors prometheus.Counter

	// statsTargets maps job_id → latest run's stats.json location and labels.
	statsTargets map[string]statsTarget
}

// statsTarget is one job's stats.json polling target from the last snapshot.
type statsTarget struct {
	workflowID string
	labels     prometheus.Labels
}

func newMetricsCollector(
	listProjectIDs func() ([]string, error),
	listJobs func(projectID string) ([]*models.Job, error),
	fetchLastRuns func(ctx context.Context, projectID string, jobs []*models.Job) (map[int]JobLastRunInfo, error),
	statsBaseDir string,
) *metricsCollector {
	newVec := func(name, help string) *prometheus.GaugeVec {
		return prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: name, Help: help}, metricsLabelNames)
	}

	m := &metricsCollector{
		listProjectIDs: listProjectIDs,
		listJobs:       listJobs,
		fetchLastRuns:  fetchLastRuns,
		statsBaseDir:   statsBaseDir,
		now:            time.Now,
		registry:       prometheus.NewRegistry(),
		status:         newVec("olake_sync_status", "Status of the most recent sync run: 0=running, 1=succeeded, 2=failed."),
		// Start time and duration come from Temporal (StartTime/CloseTime), not
		// stats.json: the file's "Seconds Elapsed" is the CLI process clock — it
		// resets on every retry attempt and freezes if the process dies, so stuck
		// syncs would never trip duration alerts. Temporal times are exact,
		// workflow-level, and come free with the status query.
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

	m.registry.MustRegister(
		m.status, m.startTime, m.duration, m.records, m.bytes, m.cpu, m.memory,
		m.scrapeErrors,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	m.handler = promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})

	return m
}

// syncStatusValue maps a Temporal execution status to the olake_sync_status gauge.
// Open workflows (Running, ContinuedAsNew, Paused) report 0 — a workflow in retry
// backoff between attempts stays 0, so no attempt-level flapping; a paused run
// eventually trips duration-based stuck alerts. Cancel counts as failed.
func syncStatusValue(status enumspb.WorkflowExecutionStatus) float64 {
	switch status {
	case enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING,
		enumspb.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW,
		enumspb.WORKFLOW_EXECUTION_STATUS_PAUSED:
		return 0
	case enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		return 1
	default: // Failed, TimedOut, Terminated, Canceled, Unspecified
		return 2
	}
}

// metricsRow is one job's fully computed snapshot values, staged so the gauge
// vecs are only mutated after every backend read has succeeded.
type metricsRow struct {
	labels     prometheus.Labels
	workflowID string
	status     float64
	start      float64
	duration   float64
	stats      *syncStats // nil when stats.json is unreadable this snapshot
}

// refresh keeps the series fresh in two TTL tiers, both on the scrape path:
// a full snapshot (DB + Temporal + stats.json) at most every metricsRefreshTTL,
// and a stats.json-only re-read at most every statsRefreshTTL in between. The
// mutex doubles as a singleflight guard for concurrent scrapes. On error the
// previous snapshot is served and the scrape-error counter increments.
func (m *metricsCollector) refresh(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.now().Sub(m.lastRefresh) >= metricsRefreshTTL {
		rows, err := m.gather(ctx)
		if err != nil {
			m.scrapeErrors.Inc()
			return err
		}
		m.commit(rows)
		now := m.now()
		m.lastRefresh = now
		m.lastStatsRead = now
		return nil
	}

	if m.now().Sub(m.lastStatsRead) >= statsRefreshTTL {
		m.refreshStats()
		m.lastStatsRead = m.now()
	}
	return nil
}

func (m *metricsCollector) gather(ctx context.Context) ([]metricsRow, error) {
	projectIDs, err := m.listProjectIDs()
	if err != nil {
		return nil, fmt.Errorf("failed to list project ids: %s", err)
	}

	rows := make([]metricsRow, 0)
	for _, projectID := range projectIDs {
		jobs, err := m.listJobs(projectID)
		if err != nil {
			return nil, fmt.Errorf("failed to list jobs project_id[%s]: %s", projectID, err)
		}

		lastRuns, err := m.fetchLastRuns(ctx, projectID, jobs)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch latest sync runs project_id[%s]: %s", projectID, err)
		}

		for _, job := range jobs {
			run, ok := lastRuns[job.ID]
			if !ok {
				// No sync run within Temporal retention → no series for this job;
				// dashboards treat absent series as "no recent run".
				continue
			}

			var sourceName, destinationName string
			if job.Source != nil {
				sourceName = job.Source.Name
			}
			if job.Destination != nil {
				destinationName = job.Destination.Name
			}

			duration := m.now().Sub(run.StartTime).Seconds()
			if run.CloseTime != nil {
				duration = run.CloseTime.Sub(run.StartTime).Seconds()
			}

			stats, err := readSyncStats(m.statsWorkDir(run.WorkflowID))
			if err != nil {
				// Missing file (run just started) or truncate-rewrite race: keep the
				// previously reported stats series for this job; self-heals next snapshot.
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

// commit replaces the emitted series with the staged rows. The status/start/duration
// vecs are fully reset so deleted jobs (and renamed labels) drop out automatically.
// The four stats vecs are updated per job instead: a job whose stats.json was
// unreadable this snapshot keeps its previous series.
func (m *metricsCollector) commit(rows []metricsRow) {
	m.status.Reset()
	m.startTime.Reset()
	m.duration.Reset()

	statsVecs := []*prometheus.GaugeVec{m.records, m.bytes, m.cpu, m.memory}

	live := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		m.status.With(row.labels).Set(row.status)
		m.startTime.With(row.labels).Set(row.start)
		m.duration.With(row.labels).Set(row.duration)

		jobID := row.labels["job_id"]
		live[jobID] = struct{}{}
		m.statsTargets[jobID] = statsTarget{workflowID: row.workflowID, labels: row.labels}
		if row.stats == nil {
			continue
		}
		for _, vec := range statsVecs {
			// Delete first so a renamed job/source/destination can't leave a stale series.
			vec.DeletePartialMatch(prometheus.Labels{"job_id": jobID})
		}
		m.setStats(row.labels, row.stats)
	}

	// Prune stats series of jobs that disappeared from the DB or fell out of
	// Temporal retention — the emitted set must mirror the current snapshot.
	for jobID := range m.statsTargets {
		if _, ok := live[jobID]; ok {
			continue
		}
		for _, vec := range statsVecs {
			vec.DeletePartialMatch(prometheus.Labels{"job_id": jobID})
		}
		delete(m.statsTargets, jobID)
	}
}

func (m *metricsCollector) setStats(labels prometheus.Labels, stats *syncStats) {
	m.records.With(labels).Set(stats.Records)
	m.bytes.With(labels).Set(stats.Bytes)
	m.cpu.With(labels).Set(stats.CPURatio)
	m.memory.With(labels).Set(stats.Memory)
}

// refreshStats re-reads stats.json for every target of the last snapshot and
// updates the four stats gauges. Filesystem-only — no DB or Temporal load.
// Caller must hold m.mu. Before the first snapshot there are no targets.
func (m *metricsCollector) refreshStats() {
	for _, target := range m.statsTargets {
		stats, err := readSyncStats(m.statsWorkDir(target.workflowID))
		if err != nil {
			continue // unreadable → keep previous values; self-heals next read
		}
		m.setStats(target.labels, stats)
	}
}

// statsWorkDir resolves a run's work directory on the shared volume using the
// exact workflow ID from visibility (same contract as telemetry's enrichWithSyncStats).
func (m *metricsCollector) statsWorkDir(workflowID string) string {
	return filepath.Join(m.statsBaseDir, fmt.Sprintf("%x", sha256.Sum256([]byte(workflowID))))
}

// MetricsHandler serves GET /metrics: a TTL-guarded snapshot refresh followed by
// Prometheus text exposition. A refresh failure serves the previous snapshot (and
// bumps olake_metrics_scrape_errors_total) instead of failing the scrape.
func (s Service) MetricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), metricsRefreshTimeout)
		defer cancel()

		if err := s.metrics.refresh(ctx); err != nil {
			logger.Errorf("metrics refresh failed, serving previous snapshot: %s", err)
		}
		s.metrics.handler.ServeHTTP(w, r)
	})
}
