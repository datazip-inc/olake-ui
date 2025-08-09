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
-   Default StorageClass: This is required by the chart to dynamically provision persistent volumes for PostgreSQL and the shared storage volume.
    ```bash
    # Check if there is a default StorageClass
    kubectl get sc

    # The output should have one class with (default) next to its name.
    # If not, default storage class must be set up. 
    # https://kubernetes.io/docs/tasks/administer-cluster/change-default-storage-class/
    
    # On Minikube, storage class support needs to be added: 
    # minikube addons enable storage-provisioner
    ```

## Installation

### Step 1: Add the OLake Helm Repository
```bash
helm repo add olake https://datazip-inc.github.io/olake
helm repo update
```

### Step 2: Install the Chart
```bash
# Using default values.yaml file
helm install olake olake/olake
```

Just like any typical Helm chart, a values file could be crafted that would define/override any of the values exposed into the default [values.yaml](https://github.com/datazip-inc/olake/helm/olake/values.yaml).
```bash
# Using the crafted my-values.yaml file
helm install --values my-values.yaml olake olake/olake
```

### Step 3: Access the OLake UI
```bash
# The UI service port (8080) needs to be forwarded to the local machine
kubectl port-forward svc/olake-ui 8000:8000

# Now, open a browser and head over to http://localhost:8000
```
Perform the login with the default credentials: `admin` / `password`.

**Note:** If you installed OLake with Ingress enabled, port-forwarding is not necessary. Simply access the application using the configured Ingress hostname.

## Upgrading

```bash
# Upgrade to latest version
helm upgrade olake ./helm/olake

# Upgrade with new values
helm upgrade olake ./helm/olake -f new-values.yaml
```

## Configuring the Helm Chart

### Initial User Setup

For enhanced security, the default admin user credentials could be replaced by using a pre-existing Kubernetes secret.

#### 1. Create a Kubernetes secret
```bash
kubectl create secret generic olake-admin-credentials \
  --from-literal=username='superadmin' \
  --from-literal=password='a-very-secure-password' \
  --from-literal=email='admin@mycompany.com'
```

#### 2. Use the secret in `values.yaml`
```yaml
olakeUI:
  initUser:
    # Reference the newly created secret. This automaticaly overwrites the initial admin credentials
    existingSecret: "olake-admin-credentials"
    secretKeys:
      username: "username"
      password: "password"
      email: "email"
```

### Ingress Configuration

The OLake UI can be exposed using an Ingress. Any Ingress controller can be used; however, the following example assumes the use of the Nginx Ingress controller and includes annotations specific to it.

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

### JobID-Based Node Mapping
With this powerful feature, specific data jobs can be routed to specific Kubernetes nodes, by which performance and reliability can be optimized.

***Where is the JobID found?***
The JobID is an integer that is automatically assigned to each job created in OLake UI. The JobID can be found in the corresponding row for each job on the Jobs page.

```yaml
global:
  jobMapping:
    123:
      olake.io/workload-type: "heavy"
    456:
      node-type: "high-cpu"
    789:
      olake.io/workload-type: "small"
    999: {} # Empty mapping uses default scheduling
  
  # - JobID Format: Must be positive integers (e.g., 123, 456, 789)
  # - Label Keys: Must follow RFC 1123 DNS subdomain format (lowercase letters, numbers, hyphens, dots)
  # - Label Values: Must be valid Kubernetes label values (63 chars max, alphanumeric with hyphens)
```

**Note on Default Behavior:** For any JobID that is not specified in the jobMapping configuration, the corresponding job's pod will be scheduled by the standard Kubernetes scheduler, which places it on any available node in the cluster.

### Cloud IAM Integration

OLake's "activity pods" (the pods by which the actual data sync is performed) can be allowed to securely access cloud resources(AWS Glue or S3) using IAM roles.

**Note:** For detailed instructions on the creation of IAM roles and service accounts, the official documentation for [AWS IRSA](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html), [GCP Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity), or [Azure Workload Identity](https://learn.microsoft.com/en-us/azure/aks/workload-identity-deploy-cluster) should be referred to.

```yaml
global:
  jobServiceAccount:
    # Create a service account for job pods
    create: true
    name: "olake-job-sa"
    
    # Cloud provider IAM role associations
    annotations:
      # AWS IRSA
      eks.amazonaws.com/role-arn: "arn:aws:iam::123456789012:role/olake-job-role"
      
      # GCP Workload Identity
      iam.gke.io/gcp-service-account: "olake-job@project.iam.gserviceaccount.com"
      
      # Azure Workload Identity
      azure.workload.identity/client-id: "12345678-1234-1234-1234-123456789012"
```

### Shared Storage Configuration

The OLake application components (UI, Worker, and Activity Pods) require a shared ReadWriteMany (RWX) volume for **coordinating pipeline state and metadata**.

For production, a robust, highly-available RWX-capable storage solution such as [AWS EFS](https://github.com/kubernetes-sigs/aws-efs-csi-driver), [GKE Filestore](https://cloud.google.com/filestore/docs/csi-driver), or [Azure Files](https://docs.microsoft.com/en-us/azure/aks/azure-files-csi) must be used. This is achieved by disabling the built-in NFS server and providing an existing PersistentVolumeClaim (PVC) that is backed by a managed storage service. An example for using external PVC is given below:

```yaml
nfsServer:
  # 1. The development NFS server is disabled
  enabled: false
  
  # 2. An existing ReadWriteMany PersistentVolumeClaim is specified
  external:
    name: "my-rwx-pvc"
```

**Note:** For development and quick starts, a simple NFS server is included and enabled by default. This provides an out-of-the-box shared storage solution without any external dependencies. However, because this server runs as a single pod, it represents a single point of failure and is not recommended for production use.

## Monitoring and Troubleshooting

### View Logs

```bash
# OLake UI logs
kubectl logs -l app.kubernetes.io/name=olake-ui -f

# OLake Worker logs
kubectl logs -l app.kubernetes.io/name=olake-workers -f
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
   kubectl get pods -l app.kubernetes.io/name=olake-workers
   kubectl logs -l app.kubernetes.io/name=olake-workers -f
   
   # Verify worker configuration
   kubectl describe configmap olake-workers-config
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

## Uninstallation

```bash
# Uninstall the release
helm uninstall olake --namespace olake

# Optional: Delete namespace and all data
kubectl delete namespace olake
```

**Note:** Some resources are intentionally preserved after `helm uninstall` to prevent accidental data loss:

- **PersistentVolumeClaims (PVCs)**: `olake-shared-storage` and database PVCs are retained to preserve job data, configurations, and historical information
- **NFS Server Resources**: If installed using the built-in NFS server, the following resources persist:
  - `Service/olake-nfs-server`
  - `StatefulSet/olake-nfs-server` 
  - `ClusterRole/olake-nfs-server`
  - `ClusterRoleBinding/olake-nfs-server`
  - `StorageClass/nfs-server`
  - `ServiceAccount/olake-nfs-server`

## Contributing

We ❤️ contributions! Check our [Bounty Program](https://olake.io/docs/community/issues-and-prs#goodies).

- UI contributions: [CONTRIBUTING.md](../CONTRIBUTING.md)
- Core contributions: [OLake Main Repository](https://github.com/datazip-inc/olake)
- Documentation: [OLake Docs Repository](https://github.com/datazip-inc/olake-docs)

## Support

- [GitHub Issues](https://github.com/datazip-inc/olake-ui/issues)
- [Slack Community](https://join.slack.com/t/getolake/shared_invite/zt-2utw44do6-g4XuKKeqBghBMy2~LcJ4ag)
- [Documentation](https://olake.io/docs)