# OLake Kubernetes Worker

A cloud-native Temporal worker that executes OLake data integration activities as Kubernetes Pods. This worker is part of the modular `olake-workers` architecture, specifically designed for Kubernetes environments.

## 🚀 Quickstart

### Prerequisites
- Go 1.23+
- Docker
- Access to a Kubernetes cluster with kubectl configured
- Temporal server running and accessible
- Olake UI running and accessible

### Build and Run Locally

1. **Clone and navigate to worker:**
   ```bash
   git clone <repo>
   cd olake-ui/olake-workers/k8s
   ```

2. **Build the worker:**
   ```bash
   go mod tidy
   go build -o olake-workers main.go
   ```

3. **Configure environment:**
   ```bash
   export TEMPORAL_HOST_PORT="localhost:7233"
   export TEMPORAL_NAMESPACE="default"
   export TEMPORAL_TASK_QUEUE="OLAKE_K8S_TASK_QUEUE"
   export DB_HOST="localhost"
   export DB_PORT="5432"
   export DB_NAME="temporal"
   export DB_USER="temporal"
   export DB_PASSWORD="temporal"
   ```

4. **Run the worker:**
   ```bash
   ./olake-workers
   ```

The worker will connect to Temporal and start listening for activities on the configured task queue.

---

## 🏗️ How It Works

The OLake Kubernetes Worker executes data integration activities as isolated Kubernetes Pods:

### Activity Types
- **Discover**: Analyzes source systems to catalog available tables/schemas
- **Test**: Validates connectivity to destination systems  
- **Sync**: Performs data replication between source and destination

### Execution Model
1. **Receives activity** from Temporal server
2. **Creates Kubernetes Pod** with appropriate container image for the task
3. **Mounts shared storage** (NFS) for data exchange between activity pods
4. **Monitors pod execution** and collects results
5. **Reports results** back to Temporal workflow

### Storage Architecture
- **Shared NFS volume** mounted at `/mnt/shared` in all activity pods
- **Activity isolation** - each activity runs in a separate Kubernetes Pod
- **Result persistence** - outputs stored in shared storage for retrieval

---

## 🛠️ Development

### Building Docker Image

```bash
# Build locally
docker build -t olakego/olake-workers:local .

# For minikube development
minikube image load olakego/olake-workers:local
```

### Project Structure

```
olake-workers/k8s/
├── main.go              # Entry point
├── activities/          # Temporal activity implementations
├── workflows/           # Temporal workflow definitions
├── worker/              # Worker setup and health endpoints
├── pods/                # Kubernetes Pod management
├── config/              # Configuration loading and validation
├── database/            # PostgreSQL integration
├── shared/              # Shared types and constants
└── utils/               # Utility packages
```

### Local Development Workflow

1. **Make code changes**
2. **Build and test:**
   ```bash
   go build -o olake-workers main.go
   go test ./...
   ```
3. **Build Docker image:**
   ```bash
   docker build -t olakego/olake-workers:local .
   ```
4. **Deploy to cluster** (see Helm chart documentation)

---

## ⚙️ Configuration

The worker is configured via environment variables:

### Required Configuration

| Variable              | Description               | Example                                 |
|-----------------------|---------------------------|-----------------------------------------|
| `TEMPORAL_HOST_PORT`  | Temporal server address   | `temporal.olake.svc.cluster.local:7233` |
| `TEMPORAL_NAMESPACE`  | Temporal namespace        | `default`                               |
| `TEMPORAL_TASK_QUEUE` | Task queue to listen on   | `olake-workers`                      |
| `DB_HOST`             | PostgreSQL host           | `postgresql.olake.svc.cluster.local`    |
| `DB_PORT`             | PostgreSQL port           | `5432`                                  |
| `DB_NAME`             | Database name             | `temporal`                              |
| `DB_USER`             | Database user             | `temporal`                              |
| `DB_PASSWORD`         | Database password         | `temporal`                              |

### Optional Configuration

| Variable                    | Description                              | Default |
|-----------------------------|------------------------------------------|---------|
| `LOG_LEVEL`                 | Logging level (debug, info, warn, error) | `info`  |
| `HEALTH_PORT`               | Health check server port                 | `8090`  |
| `MAX_CONCURRENT_ACTIVITIES` | Max parallel activities                  | `15`    |
| `MAX_CONCURRENT_WORKFLOWS`  | Max parallel workflows                   | `10`    |

### Worker-Specific Timeouts

Configure activity and workflow timeouts:

| Variable                    | Description               | Default |
|-----------------------------|---------------------------|---------|
| `TIMEOUT_WORKFLOW_DISCOVER` | Discover workflow timeout | `3h`    |
| `TIMEOUT_WORKFLOW_TEST`     | Test workflow timeout     | `3h`    |
| `TIMEOUT_WORKFLOW_SYNC`     | Sync workflow timeout     | `720h`  |
| `TIMEOUT_ACTIVITY_DISCOVER` | Discover activity timeout | `2h`    |
| `TIMEOUT_ACTIVITY_TEST`     | Test activity timeout     | `2h`    |
| `TIMEOUT_ACTIVITY_SYNC`     | Sync activity timeout     | `700h`  |

---

## 🔍 Monitoring

### Health Endpoints

The worker exposes health endpoints on port 8090:

```bash
# Liveness probe
curl http://localhost:8090/health/live

# Readiness probe  
curl http://localhost:8090/health/ready
```

### Logging

Structured JSON logging with configurable levels:

```bash
# Set debug logging
export LOG_LEVEL=debug

# View worker logs (in Kubernetes)
kubectl logs -l app.kubernetes.io/name=olake-workers -n olake
```

---

## 🐛 Troubleshooting

### Worker Won't Start

**Check Temporal connectivity:**
```bash
# Test connection to Temporal
telnet <temporal-host> 7233

# Check worker logs
kubectl logs -l app.kubernetes.io/name=olake-workers -n olake
```

**Verify database access:**
```bash
# Test PostgreSQL connection
psql -h <db-host> -p <db-port> -U <db-user> -d <db-name>
```

### Pods Failing to Execute

**Check RBAC permissions:**
```bash
# Verify service account has pod creation permissions
kubectl auth can-i create pods --as=system:serviceaccount:olake:olake-workers-sa -n olake
```

**Storage issues:**
```bash
# Check PVC status
kubectl get pvc -n olake

# Test NFS mounting
kubectl exec deployment/olake-workers -n olake -- mount | grep nfs
```

### Activity Timeouts

**Check resource constraints:**
```bash
# View pod resource usage
kubectl top pods -n olake

# Check node resources
kubectl describe nodes
```

**Adjust timeouts if needed:**
```bash
# Increase activity timeout
export TIMEOUT_ACTIVITY_SYNC=800h
```

### Common Issues

1. **Image pull errors**: Ensure `olakego/olake-workers` image is accessible
2. **Permission denied**: Check Kubernetes RBAC configuration
3. **Storage mounting failures**: Verify NFS server is running and accessible
4. **Database connection timeouts**: Check network policies and firewall rules

## 📚 Related Documentation
- **Helm Chart**: See `helm/README.md` for deployment instructions
- **OLake UI**: Main application repository
- **Temporal**: [Official Temporal Go SDK docs](https://docs.temporal.io/dev-guide/go)