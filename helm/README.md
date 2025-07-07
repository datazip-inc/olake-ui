# OLake Helm Chart

<h1 align="center" style="border-bottom: none">
    <a href="https://datazip.io/olake" target="_blank">
        <img alt="olake" src="https://github.com/user-attachments/assets/d204f25f-5289-423c-b3f2-44b2194bdeaf" width="100" height="100"/>
    </a>
    <br>OLake
</h1>

<p align="center">This <b>Helm Chart</b> deploys <b>OLake</b>, fastest open-source tool for replicating Databases to Apache Iceberg or Data Lakehouse. ⚡ Efficient, quick and scalable data ingestion for real-time analytics.</p>

## Components

-   **OLake UI**: The main application, providing the user interface and backend API.
-   **OLake Worker**: A Kubernetes-native worker that executes data synchronization, discovery, and testing tasks.
-   **Temporal**: A powerful workflow orchestration engine used to manage and monitor the lifecycle of data ingestion pipelines.
-   **PostgreSQL**: The primary data store for both the OLake application and Temporal.
-   **NFS Server**: An in-cluster NFS server for shared storage between OLake components. **Alternatively**, use an external NFS provider like AWS EFS or Azure Files.
-   **Elasticsearch**: Used by Temporal for advanced visibility and search capabilities.

## TL;DR
```bash
# Clone the repository
git clone https://github.com/datazip-inc/olake-ui.git
cd olake-ui

# https://kubernetes.io/docs/concepts/services-networking/cluster-ip-allocation/#why-do-you-need-to-reserve-service-cluster-ips
helm install olake ./helm/olake --set nfsServer.clusterIP=<STATIC_CLUSTER_IP> -f ./helm/olake/values.yaml
```

## Prerequisites

-   Kubernetes 1.19+
-   Helm 3.2.0+
-   Use helper script to find an available IP in cluster's service CIDR range:
    ```bash
    # Clone the repository
    git clone https://github.com/datazip-inc/olake-ui.git
    cd olake-ui

    # Run the IP discovery script
    chmod +x ./helm/find-service-ip.sh
    ./helm/find-service-ip.sh
    ```
-   Alternatively, use a `ReadWriteMany` (RWX) storage class if not using the built-in NFS server. Check **Storage Configuration** section below for details.

## Installation
```bash
# Install with custom values and static NFS IP
helm install olake ./helm/olake --set nfsServer.clusterIP=<STATIC_CLUSTER_IP> -f ./helm/olake/values.yaml
```

### Job Scheduling

The data processing pods (for sync, discover, and test jobs) can be scheduled on specific WorkerNodes. This is useful for isolating workloads or ensuring they run on nodes with specific capabilities.

**Example: Using node selectors and tolerations**

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
    discover:
      antiAffinity:
        enabled: false
        strategy: "soft"
        weight: 10
    test:
      antiAffinity:
        enabled: false
        strategy: "soft"
        weight: 50
```

### Storage Configuration

The chart deploys an NFS server, or an existing `ReadWriteMany` (RWX) storage solution can also be used.

**Option 1: Use the built-in NFS server (default)**

This is the easiest way to get started. The chart will create an NFS server with specific IP address.

```yaml
nfsServer:
  enabled: true
  clusterIP: "10.0.255.50" # Example static IP address
  persistence:
    size: 50Gi
    storageClass: ""
```

**Option 2: Use an external RWX storage provider**

For production environments, it is recommended to use a managed RWX storage solution like AWS EFS, Azure Files, or GCP Filestore.

```yaml
nfsServer:
  enabled: false
  external:
    name: "my-rwx-pvc" # Name of the ReadWriteMany PVC
```

## Accessing Services

After installation, the services can be accessed via:

```bash
# Get service information
kubectl get svc -n olake

# Port forward to access UI locally
kubectl port-forward -n olake svc/olake-ui 8000:8000 8080:8080

# Access Temporal UI (if enabled)
kubectl port-forward -n olake svc/temporal-ui 8088:8088
```

## Monitoring and Troubleshooting

### View Logs

```bash
# OLake Worker logs
kubectl logs -n olake -l app.kubernetes.io/name=olake-worker -f

# OLake UI logs
kubectl logs -n olake -l app.kubernetes.io/name=olake-ui -f

# Temporal logs
kubectl logs -n olake -l app.kubernetes.io/name=temporal -f
```

### Common Issues

1. **Pods stuck in Pending state**
   ```bash
   kubectl describe pod <pod-name> -n olake
   kubectl get events -n olake --sort-by='.lastTimestamp'
   ```

2. **Storage issues**
   ```bash
   kubectl get pv,pvc -n olake
   kubectl describe pvc olake-shared-storage -n olake
   ```

3. **Network connectivity**
   ```bash
   kubectl exec -it <worker-pod> -n olake -- nslookup temporal.olake.svc.cluster.local
   ```

## Configuration Reference

### Key Configuration Options

| Parameter | Description | Default |
|-----------|-------------|---------|
| `global.job.sync.antiAffinity.enabled` | Enable anti-affinity for sync jobs | `true` |
| `global.job.sync.antiAffinity.strategy` | Anti-affinity strategy (hard/soft) | `hard` |
| `nfsServer.enabled` | Enable built-in NFS server | `true` |
| `nfsServer.external.name` | External RWX PVC name | `""` |
| `nfsServer.clusterIP` | Static IP address for NFS Server | `""` |

For a complete list of configuration options, see [values.yaml](./olake/values.yaml).

## Upgrading

```bash
# Upgrade to latest version
helm upgrade olake ./helm/olake --set nfsServer.clusterIP=<STATIC_CLUSTER_IP>

# Upgrade with new values
helm upgrade olake ./helm/olake --set nfsServer.clusterIP=<STATIC_CLUSTER_IP> -f new-values.yaml
```

## Uninstallation

```bash
# Uninstall the release
helm uninstall olake --namespace olake

# Optional: Delete the namespace (this will delete all data)
kubectl delete namespace olake
```

## Support

For issues and support:
- [GitHub Issues](https://github.com/datazip-inc/olake-ui/issues)
- [Documentation](https://github.com/datazip-inc/olake-ui)