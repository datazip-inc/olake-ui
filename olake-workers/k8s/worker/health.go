package worker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"olake-ui/olake-workers/k8s/logger"
)

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

// NewHealthServer creates a new health check server
func NewHealthServer(worker *K8sWorker, port int) *HealthServer {
	mux := http.NewServeMux()

	hs := &HealthServer{
		worker: worker,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", 8090),
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
	logger.Infof("Starting health check server on port 8090")
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

	if hs.worker.temporalClient == nil || hs.worker.worker == nil {
		response.Status = "unhealthy"
		if hs.worker.temporalClient == nil {
			response.Checks["worker"] = "temporal_client_disconnected"
		} else {
			response.Checks["worker"] = "temporal_worker_failed"
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

	// Check Temporal connection
	if hs.worker != nil && hs.worker.temporalClient != nil {
		response.Checks["temporal"] = "connected"
	} else {
		response.Status = "not_ready"
		response.Checks["temporal"] = "disconnected"
	}

	// Check database connectivity
	if hs.worker.jobService.HealthCheck() == nil {
		response.Checks["database"] = "connected"
	} else {
		response.Status = "not_ready"
		response.Checks["database"] = "disconnected"
	}

	// Set HTTP status code based on overall health
	if response.Status == "not_ready" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// metricsHandler provides basic metrics
func (hs *HealthServer) metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Basic metrics - can be enhanced with Prometheus later
	metrics := map[string]interface{}{
		"worker_status":  "running",
		"uptime_seconds": time.Since(hs.worker.startTime).Seconds(),
		"timestamp":      time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
