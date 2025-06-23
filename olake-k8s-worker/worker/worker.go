package worker

import (
	"fmt"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"olake-k8s-worker/activities"
	"olake-k8s-worker/config"
	"olake-k8s-worker/logger"
	"olake-k8s-worker/shared"
	"olake-k8s-worker/utils"
	"olake-k8s-worker/workflows"
)

type K8sWorker struct {
	temporalClient client.Client
	worker         worker.Worker
	config         *config.Config
}

func NewK8sWorker() (*K8sWorker, error) {
	// Get Temporal address from environment variable
	temporalAddr := utils.GetEnv("TEMPORAL_ADDRESS", shared.DefaultTemporalAddress)

	logger.Infof("Connecting to Temporal at: %s", temporalAddr)

	// Connect to Temporal
	c, err := client.Dial(client.Options{
		HostPort: temporalAddr,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %v", err)
	}

	// Create worker - HARDCODED to K8s task queue only
	w := worker.New(c, shared.TaskQueue, worker.Options{})

	logger.Infof("Registering workflows and activities for task queue: %s", shared.TaskQueue)

	// Register workflows
	w.RegisterWorkflow(workflows.DiscoverCatalogWorkflow)
	w.RegisterWorkflow(workflows.TestConnectionWorkflow)
	w.RegisterWorkflow(workflows.RunSyncWorkflow)

	// Register activities
	w.RegisterActivity(activities.DiscoverCatalogActivity)
	w.RegisterActivity(activities.TestConnectionActivity)
	w.RegisterActivity(activities.SyncActivity)

	logger.Info("Successfully registered all workflows and activities")

	return &K8sWorker{
		temporalClient: c,
		worker:         w,
	}, nil
}

func NewK8sWorkerWithConfig(cfg *config.Config) (*K8sWorker, error) {
	logger.Infof("Connecting to Temporal at: %s", cfg.Temporal.Address)

	// Connect to Temporal
	c, err := client.Dial(client.Options{
		HostPort: cfg.Temporal.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %v", err)
	}

	// Create worker with config
	w := worker.New(c, cfg.Temporal.TaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize:     cfg.Worker.MaxConcurrentActivities,
		MaxConcurrentWorkflowTaskExecutionSize: cfg.Worker.MaxConcurrentWorkflows,
		Identity:                               cfg.Worker.WorkerIdentity,
	})

	// Register workflows and activities (same as before)
	w.RegisterWorkflow(workflows.DiscoverCatalogWorkflow)
	w.RegisterWorkflow(workflows.TestConnectionWorkflow)
	w.RegisterWorkflow(workflows.RunSyncWorkflow)

	w.RegisterActivity(activities.DiscoverCatalogActivity)
	w.RegisterActivity(activities.TestConnectionActivity)
	w.RegisterActivity(activities.SyncActivity)

	return &K8sWorker{
		temporalClient: c,
		worker:         w,
		config:         cfg, // Store config in struct
	}, nil
}

func (w *K8sWorker) Start() error {
	logger.Info("Starting Temporal worker...")
	return w.worker.Start()
}

func (w *K8sWorker) Stop() {
	logger.Info("Stopping Temporal worker...")
	w.worker.Stop()
	w.temporalClient.Close()
	logger.Info("Temporal worker stopped")
}
