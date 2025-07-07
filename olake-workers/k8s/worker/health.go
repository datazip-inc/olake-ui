package worker

import (
	"context"
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
	logger.Infof("Starting health check server on %s", hs.server.Addr)
	return hs.server.ListenAndServe()
}

// Stop stops the health check server
func (hs *HealthServer) Stop(ctx context.Context) error {
	logger.Info("Stopping health check server")
	return hs.server.Shutdown(ctx)
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

	// Check if worker is still running
	if hs.worker == nil {
		response.Status = "unhealthy"
		response.Checks["worker"] = "not_initialized"
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
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	// TODO: Add database connectivity check if needed

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
