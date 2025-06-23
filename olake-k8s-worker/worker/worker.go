package worker

import (
	"fmt"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"olake-k8s-worker/activities"
	"olake-k8s-worker/logger"
	"olake-k8s-worker/shared"
	"olake-k8s-worker/workflows"
)

type K8sWorker struct {
	temporalClient client.Client
	worker         worker.Worker
}

func NewK8sWorker() (*K8sWorker, error) {
	// Get Temporal address from environment variable
	temporalAddr := os.Getenv("TEMPORAL_ADDRESS")
	if temporalAddr == "" {
		temporalAddr = shared.DefaultTemporalAddress
	}

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
