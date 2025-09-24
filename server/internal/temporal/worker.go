package temporal

import (
	"go.temporal.io/sdk/worker"
)

// Worker handles Temporal worker functionality
type Worker struct {
	worker worker.Worker
	client *Client
}

// NewWorker creates a new Temporal worker with the provided client
func NewWorker(c *Client) *Worker {
	w := worker.New(c.GetClient(), TaskQueue, worker.Options{})

	// Register workflows
	// w.RegisterWorkflow(DiscoverCatalogWorkflow)
	// w.RegisterWorkflow(TestConnectionWorkflow)
	// w.RegisterWorkflow(RunSyncWorkflow)
	w.RegisterWorkflow("ExecuteWorkflow")

	// Register activities
	// w.RegisterActivity(DiscoverCatalogActivity)
	// w.RegisterActivity(TestConnectionActivity)
	// w.RegisterActivity(SyncActivity)
	w.RegisterActivity("ExecuteActivity")

	return &Worker{
		worker: w,
		client: c,
	}
}

// Start starts the worker
func (w *Worker) Start() error {
	return w.worker.Start()
}

// Stop stops the worker and closes the client
func (w *Worker) Stop() {
	w.worker.Stop()
	w.client.Close()
}
