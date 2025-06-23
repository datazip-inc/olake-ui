package worker

import (
	"fmt"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"olake-k8s-worker/activities"
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

	fmt.Printf("Connecting to Temporal at: %s\n", temporalAddr)

	// Connect to Temporal
	c, err := client.Dial(client.Options{
		HostPort: temporalAddr,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %v", err)
	}

	// Create worker - HARDCODED to K8s task queue only
	w := worker.New(c, shared.TaskQueue, worker.Options{})

	// Register workflows
	w.RegisterWorkflow(workflows.DiscoverCatalogWorkflow)
	w.RegisterWorkflow(workflows.TestConnectionWorkflow)
	w.RegisterWorkflow(workflows.RunSyncWorkflow)

	// Register activities
	w.RegisterActivity(activities.DiscoverCatalogActivity)
	w.RegisterActivity(activities.TestConnectionActivity)
	w.RegisterActivity(activities.SyncActivity)

	return &K8sWorker{
		temporalClient: c,
		worker:         w,
	}, nil
}

func (w *K8sWorker) Start() error {
	return w.worker.Start()
}

func (w *K8sWorker) Stop() {
	w.worker.Stop()
	w.temporalClient.Close()
}
