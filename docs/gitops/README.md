# GitOps for OLake

Manage OLake sources, destinations, jobs, and stream selections as Kubernetes custom resources in git. Argo CD applies the manifests; an **embedded reconciler** inside `olake-ui` syncs each CR into OLake by calling the ETL service layer directly (same logic as the HTTP API).

## Prerequisites

- Kubernetes cluster with OLake UI deployed **inside** the cluster (in-cluster kubeconfig).
- CRDs installed via the [olake-helm](https://github.com/datazip-inc/olake-helm) chart (`helm/olake/crds/`).
- `GITOPS_ENABLED=true` on the `olake-ui` Deployment (set by `gitops.enabled: true` in Helm).
- Argo CD (optional but recommended) with custom health Lua from [olake-helm `argocd-health.yaml`](https://github.com/datazip-inc/olake-helm/blob/master/helm/olake/docs/argocd-health.yaml).

## Quick start

### 1. Install CRDs

CRDs are installed automatically when deploying via the [olake-helm](https://github.com/datazip-inc/olake-helm) chart (`helm/olake/crds/`). CRD schema YAML is maintained in **olake-helm only** — keep it in sync with the reconciler API types in `server/internal/gitops/api/v1/` when fields change.

### 2. Enable GitOps on olake-ui

**Kubernetes (recommended):** use the olake-helm chart:

```yaml
gitops:
  enabled: true
olakeUI:
  replicaCount: 1
```

See [olake-helm/docs/gitops.md](https://github.com/datazip-inc/olake-helm/blob/master/helm/olake/docs/gitops.md).

**Docker Compose:** set `GITOPS_ENABLED=true` only when running inside a cluster with in-cluster config (disabled by default in `docker-compose-v1.yml`).

**POC note:** With olake-helm (`gitops.enabled: true`, `gitops.rbac.create: true`), a ClusterRole for `olake.io` resources is created automatically so the reconciler can watch all namespaces. Run a **single** `olake-ui` replica (no leader election in POC).

### 3. Configure Argo CD health (once per cluster)

Merge the Lua snippets from [olake-helm `docs/argocd-health.yaml`](https://github.com/datazip-inc/olake-helm/blob/master/helm/olake/docs/argocd-health.yaml) into the `argocd-cm` ConfigMap.

| Phase | Argo CD health |
|-------|----------------|
| `Ready` | Healthy |
| `Failed` | Degraded (message from `.status.message`) |
| `Pending` / empty | Progressing |

Sync success in Argo CD only means YAML was applied. **Degraded** means the operator failed to sync into OLake.

### 4. Apply sample manifests

Replace `spec.projectId` with your OLake project ID. Examples: [olake-helm/examples/gitops/](https://github.com/datazip-inc/olake-helm/tree/master/helm/olake/examples/gitops).

## Git repository layout

Do **not** use Argo CD sync waves. The Job reconciler waits (Pending + requeue) if Source, Destination, or Streams are missing, then continues when they exist.

Waves would wait for the previous wave to be Healthy before applying the next. Streams has no independent reconciler — the Job controller sets Streams `Ready` only after the Job syncs — so a Streams-before-Job wave deadlocks.

```
repo/olake/
  source-smoke.yaml
  destination-smoke.yaml
  streams-smoke.yaml
  job-smoke.yaml
```

## CRD reference

API group: `olake.io/v1`

| Kind | Spec | Purpose |
|------|------|---------|
| `Source` | `projectId`, `userId`, `config` | Source connector JSON (same as POST `/sources`) |
| `Destination` | `projectId`, `userId`, `config` | Destination connector JSON |
| `Job` | `projectId`, `userId`, `config` | Job metadata; `config` uses `source` / `destination` **by name or numeric ID** |
| `Streams` | `project_id`, `job`, `config` | Large streams catalog JSON; `job` is Job CR name or numeric ID |

### Job `spec.config` shape (git-facing)

```json
{
  "name": "smoke-job",
  "source": "smoke-source",
  "destination": "smoke-dest",
  "frequency": "0 */6 * * *",
  "activate": true
}
```

`source` and `destination` accept either the OLake entity **name** (same as the Source/Destination CR name or `config.name`) or the numeric **entity ID** (same as `status.entityId` after sync).

```json
{
  "name": "smoke-job",
  "source": "42",
  "destination": "7",
  "frequency": "0 */6 * * *",
  "activate": true
}
```

The operator resolves names or IDs to database IDs before calling `CreateJob` / `UpdateJob`.

### Streams `spec` shape

```yaml
spec:
  project_id: "123"
  job: smoke-job   # Job CR metadata.name, or OLake job ID string after sync
  config: |
    { "selected_streams": { ... } }
```

### Status

All kinds share:

```yaml
status:
  phase: Pending | Ready | Failed
  message: human-readable detail
  entityId: OLake entity ID after sync
  observedGeneration: ...
```

## Reconciliation behavior

- **Create/update only** — removing a CR from git does **not** delete the OLake entity (v1). Disable Argo CD prune for `olake.io` kinds or clean up manually.
- **Validation errors** (bad JSON, missing fields) → `phase=Failed`, no requeue. Fix git and re-sync.
- **Missing dependencies** (job before source exists) → `phase=Pending`, requeue every 10s.
- **Transient DB/service errors** → requeue with backoff, status may stay Pending or Failed with message.
- **Stream updates** — Job controller watches `Streams` CRs; on change it runs stream-difference then `UpdateJob` (same as the UI API).

Audit fields (`created_by` / `updated_by`) come from `spec.userId` (OLake user ID). Must match an existing user (e.g. `1` for the admin user from signup-init). Missing or invalid `userId` → `phase=Failed`.

## Troubleshooting

| Symptom | Likely cause | Action |
|---------|--------------|--------|
| Argo CD Sync OK, Health Progressing forever | Operator not running | Check `GITOPS_ENABLED=true`, pod logs, in-cluster config |
| `phase=Pending`, message "waiting for source …" | Source not in OLake yet | Wait for Source CR to reach Ready; Job requeues automatically |
| `phase=Pending`, "waiting for Streams CR …" | Missing or wrong `job` ref | Add Streams CR with matching `spec.job` and `project_id` |
| `phase=Failed`, validation message | Bad `spec.config` | Fix JSON to match API DTOs; check connector types |
| `phase=Failed` after DB error | Postgres/Temporal down | Fix infrastructure; operator will requeue |
| Health Degraded, Sync Synced | Expected on operator failure | Read `.status.message` and pod logs |
| Secrets in git | Passwords in CR YAML | Use SealedSecrets / ExternalSecrets post-POC |

Debug commands:

```bash
kubectl describe source smoke-source
kubectl get events --field-selector involvedObject.kind=Source
kubectl logs -l app=olake-ui | grep -i gitops
```

## Security notes

- GitOps RBAC is limited to `olake.io` resources (plus Events for controller diagnostics). It does **not** grant access to Secrets, Pods, or other cluster resources.
- Cluster-wide list/watch means the olake-ui ServiceAccount can read every `olake.io` CR in every namespace, including connector credentials stored in `spec.config`. Treat CR YAML like sensitive config.
- One olake-ui instance reconciles all matching CRs into its OLake database (keyed by `spec.projectId`). On shared clusters, isolate by project ID and namespace conventions, or run separate OLake installs per tenant.
- Post-v0: leader election for multiple replicas, delete finalizers, sealed secrets.

## Related files

- GitOps API types (reconciler): `server/internal/gitops/api/v1/` — run `make generate-deepcopy` after type changes to refresh `zz_generated.deepcopy.go`
- Reconcilers: `server/internal/gitops/`
- CRD YAML (install): [olake-helm `helm/olake/crds/`](https://github.com/datazip-inc/olake-helm/tree/master/helm/olake/crds)
- **Helm deploy (CRDs, RBAC, GITOPS_ENABLED):** [olake-helm](https://github.com/datazip-inc/olake-helm/tree/master/helm/olake)
- Argo CD health: [olake-helm/docs/argocd-health.yaml](https://github.com/datazip-inc/olake-helm/blob/master/helm/olake/docs/argocd-health.yaml)
- Examples: [olake-helm/examples/gitops/](https://github.com/datazip-inc/olake-helm/tree/master/helm/olake/examples/gitops)
