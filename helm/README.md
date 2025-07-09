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

## Overview

Olake-UI is built on top of Olake CLI to execute commands via UI.

- [UI Readme](/olake_frontend/README.md)
- [Server Readme](/server/README.md)
- [API Contracts](/api-contract.md)
- [Contributor Guidlines](/CONTRIBUTING.md)

## Components

-   **OLake UI**: The main application providing user interface and backend API
-   **OLake Worker**: Kubernetes-native worker executing data synchronization, discovery, and testing tasks
-   **Temporal**: Workflow orchestration engine managing data ingestion pipeline lifecycle
-   **PostgreSQL**: Primary data store for both OLake application and Temporal
-   **Elasticsearch**: Advanced visibility and search capabilities for Temporal
-   **NFS Server**(Optional): Self-managed in-cluster NFS server with dynamic provisioning for shared storage

## TL;DR

```bash
# Clone the repository
git clone https://github.com/datazip-inc/olake-ui.git
cd olake-ui

# Install with default values
helm install olake ./helm/olake
```

## Prerequisites

-   Kubernetes 1.19+
-   Helm 3.2.0+
-   A [Default StorageClass](https://kubernetes.io/docs/tasks/administer-cluster/change-default-storage-class/) defined

## Installation

```bash
# Install on a specific namespace
helm install olake ./helm/olake --namespace <namespace> --create-namespace

#Install with specific values file
helm install olake ./helm/olake -f </path/to/values/file> --namespace <namespace> --create-namespace
```

## Features

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

## Accessing Services

```bash
# Get service information
kubectl get svc

# Access UI locally
kubectl port-forward svc/olake-ui 8000:8000 8080:8080
```

### Common Issues

1. **Storage provisioning failures**
   ```bash
   kubectl get pv,pvc
   kubectl describe pvc shared-storage
   ```

2. **NFS server not ready**
   ```bash
   kubectl get pods -l app.kubernetes.io/name=olake-nfs-server
   kubectl logs -l app.kubernetes.io/name=olake-nfs-server
   ```

3. **Network connectivity**
   ```bash
   kubectl exec -it <pod> -- nslookup olake-nfs-server.olake.svc.cluster.local
   ```

## Configuration Reference

### Key Configuration Options

| Parameter | Description | Default |
|-----------|-------------|---------|
| `nfsServer.enabled` | Enable dynamic NFS provisioning | `true` |
| `nfsServer.image.repository` | NFS server image repository | `devxygmbh/nfs-server-provisioner` |
| `nfsServer.persistence.size` | NFS server storage size | `20Gi` |
| `nfsServer.storageClass.name` | Dynamic StorageClass name | `nfs-server` |
| `global.job.sync.antiAffinity.enabled` | Enable sync job anti-affinity | `true` |

For complete configuration options, see [values.yaml](./olake/values.yaml).

## Upgrading

```bash
# Upgrade to latest version
helm upgrade olake ./helm/olake -f ./helm/olake/values.yaml

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