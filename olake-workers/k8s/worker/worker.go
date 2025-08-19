package worker

import (
	"fmt"
	"net/http"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"olake-ui/olake-workers/k8s/activities"
	"olake-ui/olake-workers/k8s/config"
	"olake-ui/olake-workers/k8s/database/service"
	"olake-ui/olake-workers/k8s/logger"
	"olake-ui/olake-workers/k8s/pods"
	"olake-ui/olake-workers/k8s/utils/k8s"
	"olake-ui/olake-workers/k8s/workflows"
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
	// Set computed values that depend on runtime environment
	cfg.Kubernetes.JobMapping = k8s.LoadJobMappingFromEnv()
	cfg.Worker.WorkerIdentity = k8s.GenerateWorkerIdentity()

	logger.Infof("Connecting to Temporal at: %s", cfg.Temporal.Address)

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

	// Connect to Temporal with custom logger
	c, err := client.Dial(client.Options{
		HostPort: cfg.Temporal.Address,
		Logger:   logger.NewTemporalLogger(),
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

	// Register workflows - these will set WorkflowID from Temporal execution context
	w.RegisterWorkflow(workflows.DiscoverCatalogWorkflow)
	w.RegisterWorkflow(workflows.TestConnectionWorkflow)
	w.RegisterWorkflow(workflows.RunSyncWorkflow)

	// Create activities with injected dependencies
	activitiesInstance := activities.NewActivities(jobService, podManager, cfg)

	// Register activities - these receive WorkflowID from workflows
	w.RegisterActivity(activitiesInstance.DiscoverCatalogActivity)
	w.RegisterActivity(activitiesInstance.TestConnectionActivity)
	w.RegisterActivity(activitiesInstance.SyncActivity)

	logger.Info("Successfully registered all workflows and activities")
	logger.Infof("Worker Identity: %s", cfg.Worker.WorkerIdentity)

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

func (w *K8sWorker) Start() error {
	logger.Info("Starting Temporal worker...")

	// Start health server in background
	go func() {
		if err := w.healthServer.Start(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("Health server failed: %v", err)
		}
	}()

	// Start temporal worker using blocking Run method
	// This will block until the worker is stopped
	err := w.worker.Run(worker.InterruptCh())
	if err != nil {
		return fmt.Errorf("temporal worker failed: %w", err)
	}

	logger.Info("Temporal worker stopped")
	return nil
}
