package temporal

import (
	"go.temporal.io/sdk/worker"
)

// Worker handles Temporal worker functionality
type Worker struct {
	client *Client
	worker worker.Worker
}

// NewWorker creates a new Temporal worker
func NewWorker(c *Client) (*Worker, error) {
	w := worker.New(c.GetClient(), TaskQueue, worker.Options{})

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
		client: c,
		worker: w,
	}, nil
}

// Start starts the worker
func (w *Worker) Start() error {
	return w.worker.Start()
}

// Stop stops the worker
func (w *Worker) Stop() {
	w.worker.Stop()
	w.client.Close()
}
