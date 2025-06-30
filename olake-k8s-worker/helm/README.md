# OLake Helm Chart

This Helm chart deploys the complete OLake data pipeline ecosystem on Kubernetes.

## Components

- **PostgreSQL** - Database for Temporal and OLake
- **Elasticsearch** - Search engine for Temporal
- **Temporal** - Workflow engine (server + UI)
- **OLake UI** - Main application (backend + frontend)
- **OLake K8s Worker** - Kubernetes-based data pipeline worker
- **Storage** - Shared persistent storage for jobs

## Quick Start

### Prerequisites

- Kubernetes cluster (minikube, kind, or cloud provider)
- Helm 3.x installed
- Docker images built (for local development)

### Build Images (Local Development)

```bash
# From the project root directory
docker build -t olake-ui:local .
cd olake-k8s-worker
docker build -t olake-k8s-worker:local .
cd ..
```

### Install with Default Values

```bash
# Deploy to default namespace with development settings
helm install olake ./helm/olake
```

### Install for Different Environments

```bash
# Development environment (local images, NodePort services, minimal resources)
helm install olake ./helm/olake -f ./helm/olake/values/development.yaml

# Staging environment (registry images, LoadBalancer services, moderate resources)
helm install olake ./helm/olake -f ./helm/olake/values/staging.yaml

# Production environment (specific image tags, high availability, production resources)
helm install olake ./helm/olake -f ./helm/olake/values/production.yaml
```

### Custom Configuration

```bash
# Override specific values
helm install olake ./helm/olake \
  --set olakeWorker.replicaCount=3 \
  --set postgresql.auth.password=mypassword \
  --set storage.persistentVolumeClaim.size=50Gi
```

## Configuration

### Key Configuration Options

| Parameter                                           | Description               | Default       |
|-----------------------------------------------------|---------------------------|---------------|
| `namespace.name`                                    | Kubernetes namespace      | `olake`       |
| `postgresql.auth.password`                          | PostgreSQL password       | `temporal123` |
| `olakeWorker.replicaCount`                          | Number of worker replicas | `1`           |
| `storage.persistentVolumeClaim.size`                | Storage size for jobs     | `10Gi`        |
| `olakeWorker.config.worker.maxConcurrentActivities` | Max concurrent activities | `10`          |

### Environment-Specific Values

#### Development (`values/development.yaml`)
- Local Docker images
- NodePort services for easy access
- Minimal resource requirements
- Shorter timeouts for faster feedback

#### Staging (`values/staging.yaml`)
- Registry images with staging tags
- LoadBalancer services
- Moderate resource allocation
- Production-like timeouts

#### Production (`values/production.yaml`)
- Specific versioned images
- High availability (multiple replicas)
- Production-grade resources
- Security hardening
- Large storage allocations

## Accessing Services

After deployment, services are available at:

### Development (NodePort)
- **OLake UI Frontend**: `http://minikube-ip:30082`
- **OLake UI Backend**: `http://minikube-ip:30081`
- **Temporal UI**: `http://minikube-ip:30080`

### Staging/Production (LoadBalancer/Ingress)
Check service external IPs:
```bash
kubectl get services -n olake
```

## Management Commands

### Upgrade

```bash
# Upgrade with new values
helm upgrade olake ./helm/olake -f ./helm/olake/values/development.yaml
```

### Uninstall

```bash
# Remove the deployment
helm uninstall olake

# Optional: Delete namespace and PVCs
kubectl delete namespace olake
kubectl delete pvc --all -n olake
```

### Status and Logs

```bash
# Check deployment status
helm status olake
kubectl get pods -n olake

# View logs
kubectl logs -f deployment/olake-k8s-worker -n olake
kubectl logs -f deployment/olake-ui -n olake
kubectl logs -f deployment/temporal -n olake
```

## Troubleshooting

### Common Issues

1. **Images not found**
   - For local development, ensure images are built and available in the cluster
   - For minikube: `eval $(minikube docker-env)` before building

2. **Storage issues**
   - Check if storage class exists: `kubectl get storageclass`
   - Verify PVC status: `kubectl get pvc -n olake`

3. **Service connectivity**
   - Check service endpoints: `kubectl get endpoints -n olake`
   - Verify pod readiness: `kubectl get pods -n olake`

### Debug Commands

```bash
# Check all resources
kubectl get all -n olake

# Describe problematic pods
kubectl describe pods -n olake

# Check recent events
kubectl get events -n olake --sort-by=.metadata.creationTimestamp

# Check logs for specific components
kubectl logs deployment/postgresql -n olake
kubectl logs deployment/elasticsearch -n olake
kubectl logs deployment/temporal -n olake
```

## Customization

### Adding Custom Configurations

Create your own values file:

```yaml
# custom-values.yaml
olakeWorker:
  config:
    worker:
      maxConcurrentActivities: 20
    timeouts:
      workflow:
        sync: 1440h  # 60 days

postgresql:
  resources:
    requests:
      memory: "2Gi"
```

```bash
helm install olake ./helm/olake -f custom-values.yaml
```

### Using External Databases

To use external PostgreSQL or Elasticsearch:

```yaml
# external-db-values.yaml
postgresql:
  enabled: false

olakeWorker:
  config:
    database:
      host: "external-postgres.example.com"
      password: "external-password"

temporal:
  server:
    env:
      - name: POSTGRES_SEEDS
        value: "external-postgres.example.com"
```

## Development

### Helm Chart Structure

```
helm/olake/
├── Chart.yaml                 # Chart metadata
├── values.yaml                # Default values
├── values/                    # Environment-specific values
│   ├── development.yaml
│   ├── staging.yaml
│   └── production.yaml
└── templates/                 # Kubernetes manifests
    ├── namespace.yaml
    ├── storage/
    ├── postgresql/
    ├── elasticsearch/
    ├── temporal/
    ├── olake-ui/
    └── olake-worker/
```

### Testing Changes

```bash
# Dry run to check generated manifests
helm install olake ./helm/olake --dry-run --debug

# Template specific values
helm template olake ./helm/olake -f values/development.yaml
```

## Security Considerations

### Production Checklist

- [ ] Change default passwords
- [ ] Use specific image tags (not `latest`)
- [ ] Enable RBAC with minimal permissions
- [ ] Use non-root security contexts
- [ ] Configure network policies
- [ ] Set up TLS for external services
- [ ] Regular security updates

### Secrets Management

For production, consider using external secret management:

```yaml
# Use external secrets operator or similar
olakeWorker:
  config:
    database:
      password: "{{ .Values.externalSecrets.dbPassword }}"
```

## Contributing

1. Make changes to templates or values
2. Test with `helm template` and `helm install --dry-run`
3. Verify deployment in development environment
4. Update documentation
5. Submit pull request