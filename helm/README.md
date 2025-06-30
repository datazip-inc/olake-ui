# OLake: Open-Source Data Lakehouse Platform (Full Stack Helm Chart)

OLake is a cloud-native, open-source platform for database replication and data lakehouse management.  
This Helm chart deploys the **entire OLake stack**—including UI, API, worker, Temporal, PostgreSQL, and Elasticsearch—on any Kubernetes cluster.

---

## 🚀 Quick Start

### Prerequisites
- Kubernetes cluster (minikube, kind, EKS, GKE, AKS, etc.)
- [Helm 3.x](https://helm.sh/docs/intro/install/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

### Install OLake Stack with Helm
```bash
git clone https://github.com/datazip-inc/olake-ui.git
cd olake-ui
helm install olake ./helm/olake
```

#### Accessing the Services
- **OLake UI**: `kubectl get svc -n olake` (see NodePort/LoadBalancer IPs)
- **Temporal UI**: `kubectl get svc -n olake`
- **API/Worker**: Exposed as ClusterIP by default (see Helm values for customization)

---

## ⚙️ Configuration

OLake is highly configurable via Helm values and environment variables.

- **Namespace**: Default is `olake`
- **Temporal Address**: `temporal.olake.svc.cluster.local:7233`
- **Database**: `postgres.olake.svc.cluster.local`
- **PVC Storage**: `olake-jobs-pvc` for job configs and state

**To customize:**
```bash
# Edit values.yaml or provide your own
helm install olake ./helm/olake -f my-values.yaml
```
See [values.yaml](./olake/values.yaml) for all options.

---

## 🛠️ Advanced: Self-Management

### Custom Configuration
- All major settings (Temporal, DB, worker, timeouts, resources) are configurable.
- See [values.yaml](./olake/values.yaml) for all options.

### Monitoring & Debugging
```bash
kubectl logs -l app.kubernetes.io/name=olake-k8s-worker -n olake -f
kubectl get pods -n olake
kubectl describe pod <pod-name> -n olake
```

---

## 📝 Example: Minimal Custom Values

```yaml
olakeWorker:
  image:
    tag: "v1.0.0"
  config:
    temporal:
      address: "temporal.olake.svc.cluster.local:7233"
    database:
      host: "postgres.olake.svc.cluster.local"
      user: "olake"
      password: "olake"
```

---

## 🆘 Troubleshooting

- **Pods not starting?**  
  `kubectl get events -n olake`
- **Database issues?**  
  `kubectl get pods -n olake | grep postgres`
- **Logs:**  
  `kubectl logs deployment/olake-k8s-worker -n olake`

---

## 🧩 Contributing

- Fork, branch, and PR as usual.
- Please add tests and update docs for new features.
- For local development, see [CONTRIBUTING.md](../olake-k8s-worker/CONTRIBUTING.md).

---

## 🏁 Next Steps

- [ ] Add production monitoring/alerting
- [ ] Support for external databases and storage

---

## 📚 More Information

- [OLake Documentation](https://github.com/datazip-inc/olake-ui)
- [Helm Chart Reference](./olake/values.yaml)