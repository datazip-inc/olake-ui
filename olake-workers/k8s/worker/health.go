package worker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"olake-ui/olake-workers/k8s/logger"
)

const healthPort = 8090

// HealthServer provides health check endpoints
type HealthServer struct {
	server *http.Server
	worker *K8sWorker
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// writeJSON writes a JSON response with status and common headers
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if status > 0 {
		w.WriteHeader(status)
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Errorf("Failed to encode response: %v", err)
	}
}

// NewHealthServer creates a new health check server
func NewHealthServer(worker *K8sWorker, port int) *HealthServer {
	mux := http.NewServeMux()

	hs := &HealthServer{
		worker: worker,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", healthPort),
			Handler: mux,
		},
	}

	// Register health check endpoints
	mux.HandleFunc("/health", hs.healthHandler)
	mux.HandleFunc("/ready", hs.readinessHandler)
	mux.HandleFunc("/metrics", hs.metricsHandler)

	return hs
}

// Start starts the health check server
func (hs *HealthServer) Start() error {
	logger.Infof("Starting health check server on port %d", healthPort)
	return hs.server.ListenAndServe()
}

// healthHandler handles liveness probe requests
func (hs *HealthServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Checks: map[string]string{
			"worker": "running",
		},
	}

	// Check both Temporal client and worker components for complete health validation:
	// - temporalClient: The client connection to Temporal server (required for communication)
	// - worker: The actual Temporal worker instance (required for activity/workflow execution)
	// Both must be operational for the pod to process work. If either fails, Kubernetes
	// should restart the pod via liveness probe to restore functionality.
	if hs.worker.temporalClient == nil {
		response.Status = "unhealthy"
		response.Checks["worker"] = "temporal_client_disconnected"
		writeJSON(w, http.StatusServiceUnavailable, response)
		return
	}

	if hs.worker.worker == nil {
		response.Status = "unhealthy"
		response.Checks["worker"] = "temporal_worker_failed"
		writeJSON(w, http.StatusServiceUnavailable, response)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

// readinessHandler handles readiness probe requests
func (hs *HealthServer) readinessHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "ready",
		Timestamp: time.Now(),
		Checks: map[string]string{
			"temporal": "unknown",
			"database": "unknown",
		},
	}

	// Check Temporal connection - verifies worker and client are both initialized.
	// Readiness requires both components to be available before accepting traffic:
	// - worker: Must be non-nil (initialization completed)
	// - temporalClient: Must be connected (can communicate with Temporal server)
	// This prevents routing requests to pods that can't process workflows/activities.
	if hs.worker != nil && hs.worker.temporalClient != nil {
		response.Checks["temporal"] = "connected"
	} else {
		response.Status = "not_ready"
		response.Checks["temporal"] = "disconnected"
	}

	// Check database connectivity - ensures job metadata can be read/written.
	// Database access is required for:
	// - Fetching job configurations and state
	// - Updating job progress and results
	// - Temporal workflow coordination
	// Without database access, workflows will fail during execution.
	if hs.worker.db.Ping() == nil {
		response.Checks["database"] = "connected"
	} else {
		response.Status = "not_ready"
		response.Checks["database"] = "disconnected"
	}

	// Set HTTP status code based on overall health
	if response.Status == "not_ready" {
		writeJSON(w, http.StatusServiceUnavailable, response)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

// metricsHandler provides basic metrics
func (hs *HealthServer) metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Basic metrics - can be enhanced with Prometheus later
	metrics := map[string]interface{}{
		"worker_status":  "running",
		"uptime_seconds": time.Since(hs.worker.startTime).Seconds(),
		"timestamp":      time.Now(),
	}

	writeJSON(w, http.StatusOK, metrics)
}
