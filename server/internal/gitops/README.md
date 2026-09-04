# GitOps for OLake

Manage OLake sources, destinations, jobs, and stream selections as labelled **ConfigMaps** (or, for Source/Destination, **Secrets**) in git. Argo CD or Flux applies the manifests; an **embedded reconciler** inside `olake-ui` syncs each object into OLake by calling the ETL service layer (same logic as the HTTP API).

## Prerequisites

- Kubernetes: OLake UI deployed **inside** the cluster (in-cluster kubeconfig).
- `GITOPS_ENABLED=true` on the `olake-ui` Deployment (Helm `gitops.enabled: true`).
- GitOps RBAC: namespace `Role` on `configmaps` and `secrets` (get/list/watch/update/patch each), and `events` (create/patch). **No `pods` permission on olake-ui.**

**Secrets RBAC note:** granting `update`/`patch` on Secrets (needed because the operator writes `olake.io/*` status annotations back onto whichever object type a Source/Destination uses) applies to **every** Secret in the release namespace, not only ones managed by GitOps. Deploy OLake GitOps in a dedicated namespace, or accept that trust boundary.

## Quick start

### 1. Enable GitOps

**Kubernetes (Helm):**

```yaml
gitops:
  enabled: true
```

See [olake-helm/docs/gitops.md](https://github.com/datazip-inc/olake-helm/blob/master/helm/olake/docs/gitops.md).

### 2. Apply sample manifests

Replace `projectId` with your OLake project ID. Examples: [olake-helm/examples/gitops/](https://github.com/datazip-inc/olake-helm/tree/master/helm/olake/examples/gitops).

### 3. Argo CD / Flux health

Status is written to **annotations** on the managed object (ConfigMap or Secret; not a status subresource):

| Annotation | Meaning |
|------------|---------|
| `olake.io/phase` | `Pending`, `Ready`, or `Failed` |
| `olake.io/message` | Human-readable detail |
| `olake.io/entity-id` | OLake entity ID after sync |
| `olake.io/observed-hash` | Content hash for skip logic |

Kubernetes Events: reason `Synced` (normal) or `SyncFailed` (warning).

**TODO:** Tool-specific health (Argo CD Lua, Flux) keyed off `olake.io/phase` — see olake-helm docs.

## Git repository layout

```
repo/olake/
  source-smoke.yaml
  destination-smoke.yaml
  streams-smoke.yaml
  job-smoke.yaml
```

## ConfigMap / Secret reference

Label every managed object:

```yaml
metadata:
  labels:
    olake.io/managed: "true"
    olake.io/kind: source   # source | destination | job | streams
```

| Kind | Object type | `data` keys | Purpose |
|------|-------------|-------------|---------|
| source | ConfigMap **or** Secret | `projectId`, `userId`, `config` | Source connector JSON (POST `/sources` shape) |
| destination | ConfigMap **or** Secret | `projectId`, `userId`, `config` | Destination connector JSON |
| job | ConfigMap only | `projectId`, `userId`, `config` | Job metadata; `config.source` / `destination` by name or numeric ID |
| streams | ConfigMap only | `projectId`, `job`, `config` | Streams catalog; `job` matches Job ConfigMap name or job entity ID |

### Streams catalog formats

Two formats are accepted in the Streams ConfigMap `data.config` JSON. **Pick one format at job creation and keep it** — switching between legacy and split after the job exists is unsupported and may leave `selected_streams_config` inconsistent (the operator does not enforce this; it is your responsibility).

**Legacy (default)** — full catalog in git:

```json
{
  "streams": [ ... stream metadata ... ],
  "selected_streams": { "public": [ { "stream_name": "users", "sync_mode": "cdc", ... } ] }
}
```

Stored as `streams_config` in the DB; `selected_streams_config` is empty. No discover step at reconcile time.

**Split (GitOps-optimized)** — user choices only in git:

```json
{
  "selected_streams": {
    "public": [
      {
        "stream_name": "users",
        "sync_mode": "cdc",
        "normalization": true,
        "selected_columns": { "columns": [], "sync_new_columns": true }
      }
    ]
  }
}
```

- `streams[]` must **not** be present in the CM.
- `selected_streams` must be non-empty (deselect-all fails reconcile).
- On reconcile the operator runs discover against the source and stores discovered `streams[]` metadata as `streams_config` in the DB (and `selected_streams` as `selected_streams_config` when split). The CM `selected_streams` are the merge input; after discover, `mergeCatalogs` (CLI) may update `selected_columns` when `sync_new_columns` is true and refreshes `streams[]` when the source schema changes. On create the CM is stored as-is; on update the merged discover output is persisted for both configs.
- Sync/clear mount both files at runtime (requires source connector version that supports `--selected-streams`; version gating is handled by temporal/worker, not the gitops reconciler).

### Credentials via Secrets (Source and Destination only)

A Source or Destination can be defined directly as a **Secret** instead of a ConfigMap when the connector config holds credentials that shouldn't be plaintext in git. There is no separate reference field — the Secret itself carries the same `data` shape (`projectId`, `userId`, `config`) and the same `olake.io/managed`/`olake.io/kind` labels as the ConfigMap form. Pick exactly one object type per Source/Destination name; if both exist for the same name, the ConfigMap wins.

Example: `source-secret-smoke.yaml` in olake-helm.

Example source as a Secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-postgres
  labels:
    olake.io/managed: "true"
    olake.io/kind: source
type: Opaque
stringData:
  projectId: "123"
  userId: "1"
  config: |
    {
      "name": "my-postgres",
      "type": "postgres",
      "version": "v0.2.7",
      "config": "{\"host\":\"...\",\"password\":\"...\"}"
    }
```

Rotating the Secret's `data.config` re-triggers reconcile the same way editing the ConfigMap would (both are watched).

Do **not** put `olake.io/phase` or other status annotations in git — the operator writes them.

## Failure indicators

On permanent reconcile failure, olake-ui starts Temporal **`IndicatorWorkflow`** (fire-and-forget). **olake-worker** creates a short-lived busybox Pod (K8s) or container (Docker) labelled `olake.io/indicator=true`. The container writes the error to `/dev/termination-log` and exits 1.

Configure alerts on failed Pods with that label. olake-ui never creates Pods directly.

## Reconciliation behavior

- **Create/update only** — deleting a ConfigMap or Secret from git does **not** delete the OLake entity (v1).
- **Validation errors** → `phase=Failed`, indicator spawned, no requeue until content changes.
- **Failed job + source/destination fix** → retrying a Failed job does **not** require editing the job file. Changing the referenced source or destination (ConfigMap or Secret data) invalidates the skip hash and re-runs the job. Streams are not part of that hash (catalogs are large); a Failed job does not auto-retry on a streams-only edit.
- **Missing dependencies** → `phase=Pending`, requeue every 30s, no indicator.
- **Transient errors** → requeue, stay non-terminal.
- **Streams** — no standalone reconciler; Job controller owns streams status and drift detection.

`userId` must be an existing OLake user ID (string in `data.userId`).

## Troubleshooting

| Symptom | Action |
|---------|--------|
| Pending, waiting for source | Ensure source ConfigMap reaches Ready first |
| Failed, validation | Fix `data.config` JSON |
| Failed, no indicator Pod | Check worker logs and Temporal; worker needs `pods` create |
| Argo synced but phase empty | Operator may be disabled; check `GITOPS_ENABLED` |

```bash
kubectl describe configmap my-postgres
kubectl get events --field-selector involvedObject.name=my-postgres
kubectl logs -l app.kubernetes.io/name=olake-ui | grep -i gitops
```

## Security notes

- olake-ui GitOps RBAC: ConfigMaps and Secrets (read/write status annotations on both), Events in the release namespace. **No pods create/delete.**
- Secret `update`/`patch` (needed for status annotations) applies to all Secrets in the namespace, not only GitOps-managed ones — use a dedicated namespace for GitOps-managed OLake when possible.
- Failure indicators use worker SA (already has `pods` create for connector jobs).
