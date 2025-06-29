package worker

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"olake-k8s-worker/activities"
	"olake-k8s-worker/config"
	"olake-k8s-worker/database/service"
	"olake-k8s-worker/logger"
	"olake-k8s-worker/pods"
	"olake-k8s-worker/workflows"
)

type K8sWorker struct {
	temporalClient client.Client
	worker         worker.Worker
	config         *config.Config
	healthServer   *HealthServer
	jobService     service.JobDataService
	podManager     *pods.K8sPodManager
	startTime      time.Time
}

// NewK8sWorkerWithConfig creates a new K8s worker with full configuration
func NewK8sWorkerWithConfig(cfg *config.Config) (*K8sWorker, error) {
	logger.Infof("Connecting to Temporal at: %s", cfg.Temporal.Address)

	// Set global config for workflows to use
	workflows.SetConfig(cfg)
	logger.Info("Set global configuration for workflows")

	// Create database service
	jobService, err := service.NewPostgresJobService()
	if err != nil {
		return nil, fmt.Errorf("failed to create job service: %v", err)
	}
	logger.Info("Created database job service")

	// Create pod manager
	podManager, err := pods.NewK8sPodManager(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create pod manager: %v", err)
	}
	logger.Info("Created K8s pod manager")

	// Connect to Temporal
	c, err := client.Dial(client.Options{
		HostPort: cfg.Temporal.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %v", err)
	}

	// Create worker with config - HARDCODED to K8s task queue only
	w := worker.New(c, cfg.Temporal.TaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize:     cfg.Worker.MaxConcurrentActivities,
		MaxConcurrentWorkflowTaskExecutionSize: cfg.Worker.MaxConcurrentWorkflows,
		Identity:                               cfg.Worker.WorkerIdentity,
	})

	logger.Infof("Registering workflows and activities for task queue: %s", cfg.Temporal.TaskQueue)

	// Log timeout configuration
	logger.Infof("Activity timeouts configured - Discover: %v, Test: %v, Sync: %v",
		cfg.GetActivityTimeout("discover"),
		cfg.GetActivityTimeout("test"),
		cfg.GetActivityTimeout("sync"))

	// Register workflows - these will set WorkflowID from Temporal execution context
	w.RegisterWorkflow(workflows.DiscoverCatalogWorkflow)
	w.RegisterWorkflow(workflows.TestConnectionWorkflow)
	w.RegisterWorkflow(workflows.RunSyncWorkflow)

	// Create activities with injected dependencies
	activitiesInstance := activities.NewActivities(jobService, podManager)

	// Register activities - these receive WorkflowID from workflows
	w.RegisterActivity(activitiesInstance.DiscoverCatalogActivity)
	w.RegisterActivity(activitiesInstance.TestConnectionActivity)
	w.RegisterActivity(activitiesInstance.SyncActivity)

	logger.Info("Successfully registered all workflows and activities")

	k8sWorker := &K8sWorker{
		temporalClient: c,
		worker:         w,
		config:         cfg,
		jobService:     jobService,
		podManager:     podManager,
		startTime:      time.Now(),
	}

	// Create health server
	k8sWorker.healthServer = NewHealthServer(k8sWorker, 8090)

	return k8sWorker, nil
}

// NewK8sWorker creates a new K8s worker with default configuration (for backward compatibility)
func NewK8sWorker() (*K8sWorker, error) {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %v", err)
	}

	// Use the full configuration constructor
	return NewK8sWorkerWithConfig(cfg)
}

func (w *K8sWorker) Start() error {
	logger.Info("Starting Temporal worker...")

	// Start health server in background
	go func() {
		logger.Info("Starting health check server on :8090")
		if err := w.healthServer.Start(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("Health server failed: %v", err)
		}
	}()

	// Start temporal worker
	if err := w.worker.Start(); err != nil {
		return fmt.Errorf("failed to start Temporal worker: %v", err)
	}

	logger.Info("K8s Worker started successfully")
	return nil
}

func (w *K8sWorker) Stop() {
	logger.Info("Stopping Temporal worker...")

	// Stop health server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := w.healthServer.Stop(ctx); err != nil {
		logger.Errorf("Failed to stop health server: %v", err)
	}

	// Close job service
	if err := w.jobService.Close(); err != nil {
		logger.Errorf("Failed to close job service: %v", err)
	}

	// Stop temporal worker
	w.worker.Stop()
	w.temporalClient.Close()
	logger.Info("Temporal worker stopped")
}

// GetConfig returns the worker configuration
func (w *K8sWorker) GetConfig() *config.Config {
	return w.config
}

// GetUptime returns how long the worker has been running
func (w *K8sWorker) GetUptime() time.Duration {
	return time.Since(w.startTime)
}

// IsHealthy returns true if the worker is healthy
func (w *K8sWorker) IsHealthy() bool {
	// Could add more sophisticated health checks here
	return w.temporalClient != nil && w.worker != nil
}
