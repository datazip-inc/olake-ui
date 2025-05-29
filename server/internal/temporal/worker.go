package temporal

import (
	"fmt"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// Worker handles Temporal worker functionality
type Worker struct {
	temporalClient client.Client
	worker         worker.Worker
}

// NewWorker creates a new Temporal worker
func NewWorker(address string) (*Worker, error) {
	if address == "" {
		address = "localhost:7233" // Default Temporal address
	}

	c, err := client.Dial(client.Options{
		HostPort: address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %v", err)
	}

	// Create a worker
	w := worker.New(c, TaskQueue, worker.Options{})

	// Register workflows
	//w.RegisterWorkflow(DockerRunnerWorkflow)
	w.RegisterWorkflow(DiscoverCatalogWorkflow)
	//w.RegisterWorkflow(GetSpecWorkflow)
	w.RegisterWorkflow(TestConnectionWorkflow)
	w.RegisterWorkflow(RunSyncWorkflow)

	// Register activities
	//w.RegisterActivity(ExecuteDockerCommandActivity)
	w.RegisterActivity(DiscoverCatalogActivity)
	//w.RegisterActivity(GetSpecActivity)
	w.RegisterActivity(TestConnectionActivity)
	w.RegisterActivity(SyncActivity)

	return &Worker{
		temporalClient: c,
		worker:         w,
	}, nil
}

// Start starts the worker
func (w *Worker) Start() error {
	return w.worker.Start()
}

// Stop stops the worker
func (w *Worker) Stop() {
	w.worker.Stop()
	w.temporalClient.Close()
}
