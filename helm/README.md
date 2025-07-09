# OLake Helm Chart

<h1 align="center" style="border-bottom: none">
    <a href="https://datazip.io/olake" target="_blank">
        <img alt="olake" src="https://github.com/user-attachments/assets/d204f25f-5289-423c-b3f2-44b2194bdeaf" width="100" height="100"/>
    </a>
    <br>OLake
</h1>

<p align="center">Fastest open-source tool for replicating Databases to Apache Iceberg or Data Lakehouse. ⚡ Efficient, quick and scalable data ingestion for real-time analytics. Starting with MongoDB. Visit <a href="https://olake.io/" target="_blank">olake.io/docs</a> for the full documentation, and benchmarks</p>

<p align="center">
    <a href="https://github.com/datazip-inc/olake-ui/issues"><img alt="GitHub issues" src="https://img.shields.io/github/issues/datazip-inc/olake"/></a> <a href="https://olake.io/docs"><img alt="Documentation" height="22" src="https://img.shields.io/badge/view-Documentation-blue?style=for-the-badge"/></a>
    <a href="https://join.slack.com/t/getolake/shared_invite/zt-2utw44do6-g4XuKKeqBghBMy2~LcJ4ag"><img alt="slack" src="https://img.shields.io/badge/Join%20Our%20Community-Slack-blue"/></a>
</p>

## Components

-   **OLake UI**: The main application providing user interface and backend API
-   **OLake Worker**: Kubernetes-native worker executing data synchronization, discovery, and testing tasks
-   **Temporal**: Workflow orchestration engine managing data ingestion pipeline lifecycle
-   **PostgreSQL**: Primary data store for both OLake application and Temporal
-   **Elasticsearch**: Advanced visibility and search capabilities for Temporal
-   **NFS Server**(Optional): Self-managed in-cluster NFS server with dynamic provisioning for shared storage

## Prerequisites

-   Kubernetes 1.19+
-   Helm 3.2.0+
-   A [Default StorageClass](https://kubernetes.io/docs/tasks/administer-cluster/change-default-storage-class/) defined

## Installation

```bash
# Install on a default namespace
helm install olake ./helm/olake

#Install with specific values file
helm install olake ./helm/olake -f </path/to/values/file> --namespace <namespace> --create-namespace
```

## Accessing Services

```bash
# Get service information
kubectl get svc

# Access UI locally
kubectl port-forward svc/olake-ui 8000:8000 8080:8080
```

## Features

### Initial User Setup

Configure the default admin user credentials during installation:

```yaml
olakeUI:
  initUser:
    adminUser:
      username: "admin"
      password: "your-secure-password"
```

### Ingress Configuration

Enable external access to the OLake UI through Kubernetes Ingress:

```yaml
olakeUI:
  ingress:
    enabled: true
    className: "nginx"
    annotations:
      nginx.ingress.kubernetes.io/rewrite-target: /
    hosts:
      - host: olake.example.com
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: olake-tls
        hosts:
          - olake.example.com
```

### Job Scheduling

Data processing pods can be scheduled on specific nodes with custom constraints:

```yaml
global:
  job:
    sync:
      nodeSelector:
        nodegroup: large
      tolerations:
        - key: "workload"
          operator: "Equal"
          value: "true"
          effect: "NoSchedule"
      antiAffinity:
        enabled: true
        strategy: "hard"
        topologyKey: "kubernetes.io/hostname"
        weight: 100
```

### Storage Configuration

**Dynamic NFS Provisioning (default)**

The chart deploys a self-managed NFS server with dynamic provisioning:

```yaml
nfsServer:
  enabled: true
  persistence:
    size: 20Gi
  storageClass:
    name: "nfs-server"
```

**External Storage Provider**

For production environments, use external ReadWriteMany storage:

```yaml
nfsServer:
  enabled: false
  external:
    name: "my-rwx-pvc"
```


## Monitoring and Troubleshooting

### View Logs

```bash
# OLake UI logs
kubectl logs -l app.kubernetes.io/name=olake-ui -f

# OLake Worker logs
kubectl logs -l app.kubernetes.io/name=olake-worker -f

# Temporal logs
kubectl logs -l app.kubernetes.io/name=temporal -f

# PostgreSQL logs
kubectl logs -l app.kubernetes.io/name=postgresql -f

# NFS Server logs
kubectl logs -l app.kubernetes.io/name=olake-nfs-server -f
```

### Common Issues

1. **Pod deployment failures**
   ```bash
   # Check pod status across all components
   kubectl get pods -l app.kubernetes.io/instance=olake
   
   # Describe problematic pods for detailed error information
   kubectl describe pod <pod-name>
   
   # Check recent events for deployment issues
   kubectl get events --sort-by='.lastTimestamp' --field-selector type!=Normal
   ```

2. **OLake UI not starting**
   ```bash
   # Check UI pod status and logs
   kubectl get pods -l app.kubernetes.io/name=olake-ui
   kubectl logs -l app.kubernetes.io/name=olake-ui -f
   ```

3. **OLake Worker issues**
   ```bash
   # Check worker pod status and logs
   kubectl get pods -l app.kubernetes.io/name=olake-worker
   kubectl logs -l app.kubernetes.io/name=olake-worker -f
   
   # Verify worker configuration
   kubectl describe configmap olake-worker-config
   ```

4. **Storage provisioning failures**
   ```bash
   # Check PVC and PV status
   kubectl get pv,pvc
   kubectl describe pvc shared-storage
   
   # Check NFS server pod and StorageClass
   kubectl get pods -l app.kubernetes.io/name=olake-nfs-server
   kubectl describe storageclass nfs-server
   ```

5. **Network connectivity issues**
   ```bash
   # Test service discovery
   kubectl exec -it <pod> -- nslookup temporal
   kubectl exec -it <pod> -- nslookup postgresql
   kubectl exec -it <pod> -- nslookup olake-nfs-server
   
   # Check service endpoints
   kubectl get endpoints temporal postgresql olake-nfs-server
   ```

## Upgrading

```bash
# Upgrade to latest version
helm upgrade olake ./helm/olake

# Upgrade with new values
helm upgrade olake ./helm/olake -f new-values.yaml
```

## Uninstallation

```bash
# Uninstall the release
helm uninstall olake --namespace olake

# Optional: Delete namespace and all data
kubectl delete namespace olake
```

## Contributing

We ❤️ contributions! Check our [Bounty Program](https://olake.io/docs/community/issues-and-prs#goodies).

- UI contributions: [CONTRIBUTING.md](../CONTRIBUTING.md)
- Core contributions: [OLake Main Repository](https://github.com/datazip-inc/olake)
- Documentation: [OLake Docs Repository](https://github.com/datazip-inc/olake-docs)

## Support

- [GitHub Issues](https://github.com/datazip-inc/olake-ui/issues)
- [Slack Community](https://join.slack.com/t/getolake/shared_invite/zt-2utw44do6-g4XuKKeqBghBMy2~LcJ4ag)
- [Documentation](https://olake.io/docs)