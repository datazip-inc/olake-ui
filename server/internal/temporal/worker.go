package temporal

import (
	"go.temporal.io/sdk/worker"
)

// Worker handles Temporal worker functionality
type Worker struct {
	worker   worker.Worker
	temporal *Temporal
}

// NewWorker creates a new Temporal worker with the provided client
func NewWorker(temporal *Temporal) *Worker {
	w := worker.New(temporal.Client, TaskQueue, worker.Options{})

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
	w.RegisterActivity(SyncCleanupActivity)

	return &Worker{
		worker:   w,
		temporal: temporal,
	}
}

// Start starts the worker
func (w *Worker) Start() error {
	return w.worker.Start()
}

// Stop stops the worker and closes the client
func (w *Worker) Stop() {
	w.worker.Stop()
	w.temporal.Close()
}
