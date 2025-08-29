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
func NewWorker() (*Worker, error) {
	c, err := client.Dial(client.Options{
		HostPort: TemporalAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %v", err)
	}

	// Create a worker
	w := worker.New(c, TaskQueue, worker.Options{})

	// Register workflows
	w.RegisterWorkflow(DiscoverCatalogWorkflow)
	w.RegisterWorkflow(TestConnectionWorkflow)
	w.RegisterWorkflow(RunSyncWorkflow)
	w.RegisterWorkflow(FetchSpecWorkflow)

	// Register activities
	w.RegisterActivity(DiscoverCatalogActivity)
	w.RegisterActivity(TestConnectionActivity)
	w.RegisterActivity(SyncActivity)
	w.RegisterActivity(FetchSpecActivity)

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
