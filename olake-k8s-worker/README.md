# OLake Kubernetes Worker

A Kubernetes-native Temporal worker for the OLake data replication platform that executes data synchronization tasks using Kubernetes Jobs instead of Docker containers.

## Overview

This worker is designed to run alongside the existing Docker-based worker, providing a cloud-native execution environment. It connects to the same Temporal server but uses a different task queue (`OLAKE_K8S_TASK_QUEUE`) and executes operations as Kubernetes Jobs.

**Note:** This implementation is currently focused on local development and testing with minikube.

## Architecture

- **Activities**: Kubernetes Job-based execution for catalog discovery, connection testing, and data synchronization
- **Workflows**: Identical to server workflows for compatibility
- **Database**: PostgreSQL integration with environment-aware table naming
- **Configuration**: ConfigMap-based configuration management
- **Logging**: Structured logging with Zerolog
- **Health Checks**: HTTP endpoints for Kubernetes probes

## Local Development Setup

### Prerequisites

- **minikube** (for local Kubernetes cluster)
- **kubectl** (configured for minikube)
- **Go 1.21+** (for building from source)
- **Docker** (for building images)
- **Make** (for build automation)

### Quick Start

1. **Start minikube**:
```bash
minikube start
```

2. **Clone and build**:
```bash
git clone https://github.com/datazip-inc/olake-ui.git
cd olake-ui/olake-k8s-worker

# Build the worker binary locally
make build-local

# Or build Docker image for minikube
make build-minikube
```

3. **Deploy full stack to minikube**:
```bash
cd k8s/manifests/full-stack/scripts
./deploy-all.sh
```

4. **Access services**:
```bash
# Get minikube IP
minikube ip

# Access services:
# - OLake UI: http://<minikube-ip>:30082
# - OLake API: http://<minikube-ip>:30081  
# - Temporal UI: http://<minikube-ip>:30080
```

## Development Workflow

### Building

```bash
# Build Go binary locally
make build-local

# Build Docker image and load into minikube
make build-minikube

# Run tests
make test

# Clean up
make clean
```

### Testing Locally

```bash
# Option 1: Run with Go (requires kubectl access to minikube)
make run-local

# Option 2: Run in minikube cluster
make deploy-minikube
```

### Configuration for Local Development

The worker uses these default configurations for local development:

- **Temporal Address**: `temporal.olake.svc.cluster.local:7233`
- **Database**: PostgreSQL in minikube cluster
- **Health Port**: `8090` (to avoid conflict with OLake server on 8080)
- **Log Level**: `debug` for local development

### Makefile Targets

| Target | Description |
|--------|-------------|
| `build-local` | Build Go binary for local development |
| `build-minikube` | Build Docker image and load into minikube |
| `run-local` | Run worker locally with minikube cluster access |
| `deploy-minikube` | Deploy worker to minikube cluster |
| `test` | Run unit tests |
| `clean` | Clean up local artifacts |
| `logs` | Show worker logs from minikube |
| `status` | Check deployment status in minikube |

## Job Execution

The worker creates Kubernetes Jobs in minikube for each operation:

- **Discover Catalog**: Analyzes source schema and returns available tables/streams
- **Test Connection**: Validates source/destination connectivity  
- **Sync Data**: Performs actual data replication

## Monitoring in Minikube

```bash
# Check worker status
make status

# View worker logs
make logs

# List all jobs
kubectl get jobs -n olake

# Check specific job logs
kubectl logs job/<job-name> -n olake

# Access Temporal UI
minikube service temporal-ui -n olake
```

## Troubleshooting

### Common Local Development Issues

1. **minikube not accessible**:
```bash
minikube status
kubectl cluster-info
```

2. **Image not found**:
```bash
# Rebuild and load image
make build-minikube
```

3. **Worker not starting**:
```bash
# Check logs and events
make logs
kubectl get events -n olake --sort-by=.metadata.creationTimestamp
```

4. **Database connection issues**:
```bash
# Check if PostgreSQL is running
kubectl get pods -n olake | grep postgresql
```

### Debugging Commands

```bash
# Full deployment status
cd k8s/manifests/full-stack/scripts && ./status.sh

# Worker-specific debugging
kubectl describe deployment olake-k8s-worker -n olake
kubectl logs -l app.kubernetes.io/name=olake-k8s-worker -n olake -f

# Database connectivity test
kubectl exec -it deployment/postgresql -n olake -- psql -U olake -d olake -c "SELECT 1;"
```

## Local Development vs Production

This implementation is optimized for local development with minikube. For production deployment:

- Images should be pushed to a container registry
- Resource limits should be adjusted for production workloads
- Persistent storage should be configured for PostgreSQL
- TLS/SSL should be enabled for all connections
- RBAC should be more restrictive

## Contributing

1. Follow existing code patterns
2. Test changes in minikube environment
3. Add tests for new functionality
4. Update documentation
5. Ensure proper error handling and logging

## Next Steps

- [ ] Production-ready Helm charts
- [ ] Integration with external monitoring systems
- [ ] Performance optimization for large datasets
- [ ] Multi-cluster deployment support