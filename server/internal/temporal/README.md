# Temporal-based Docker Runner

This package provides a Temporal-based implementation for running Docker commands. It offers improved reliability, observability, and error handling compared to the direct Docker command execution approach.

## Features

- Durable execution with automatic retries
- Detailed workflow history for debugging
- Heartbeats to track long-running operations
- Improved monitoring and visibility
- Better error handling and recovery

## Prerequisites

1. Install and run a local Temporal server:

```bash
# Using docker-compose
docker-compose up -d --build

# OR using Temporal CLI
temporal server start-dev
```

See [Temporal documentation](https://docs.temporal.io/clusters/quick-install) for more installation options.

## Usage

### Starting a Temporal Worker

You need to run at least one worker to process workflow and activity tasks:

```go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/datazip/olake-server/internal/temporal"
)

func main() {
    // Create and start a worker
    worker, err := temporal.NewWorker("")
    if err != nil {
        log.Fatalf("Failed to create worker: %v", err)
    }
    
    // Start the worker
    go func() {
        if err := worker.Start(); err != nil {
            log.Fatalf("Failed to start worker: %v", err)
        }
    }()
    
    // Handle graceful shutdown
    signalChan := make(chan os.Signal, 1)
    signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
    <-signalChan
    
    log.Println("Shutting down worker...")
    worker.Stop()
}
```

### Using the Temporal-based Docker Runner

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/datazip/olake-server/internal/docker"
)

func main() {
    // Create a Temporal-based runner
    runner, err := docker.NewTemporalRunner("", "")
    if err != nil {
        log.Fatalf("Failed to create Temporal runner: %v", err)
    }
    defer runner.Close()
    
    // Example: Get catalog from PostgreSQL source
    config := `{
        "host": "postgres",
        "port": 5432,
        "database": "example",
        "username": "postgres",
        "password": "postgres"
    }`
    
    result, err := runner.GetCatalog("postgres", "latest", config, 1)
    if err != nil {
        log.Fatalf("Failed to get catalog: %v", err)
    }
    
    fmt.Printf("Catalog result: %+v\n", result)
}
```

## Monitoring and Debugging

You can access the Temporal Web UI to monitor and debug workflow executions:

- Local development: http://localhost:8233
- With standard Temporal: http://localhost:8080

The Web UI provides:
- Workflow execution history
- Activity details and failures
- Workflow retry information
- Query and signal capabilities

## Advanced Usage

### Custom Workflow Configurations

You can customize workflow options like timeouts, retry policies, and task queues by modifying the Client implementation.

### Running with a Production Temporal Cluster

For production, configure your application to connect to your production Temporal cluster:

```go
// Connect to production Temporal cluster
runner, err := docker.NewTemporalRunner("", "temporal.example.com:7233")
```

## Troubleshooting

1. **Worker Not Processing Tasks:** Ensure the worker is running and registered to the same task queue.
2. **Connection Issues:** Verify Temporal server is running and accessible.
3. **Docker Execution Failures:** Check Docker is installed and available to the worker process.

## Additional Resources

- [Temporal Documentation](https://docs.temporal.io/)
- [Go SDK Documentation](https://pkg.go.dev/go.temporal.io/sdk)