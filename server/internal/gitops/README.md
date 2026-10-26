# GitOps for OLake

Manage OLake sources, destinations, jobs, and stream selections as labelled **ConfigMaps** in git. Argo CD or Flux applies the manifests; an **embedded reconciler** inside `olake-ui` syncs each object into OLake.

## Prerequisites

- Kubernetes: OLake UI deployed **inside** the cluster (in-cluster kubeconfig), or Docker Compose with file-based manifests mounted at `/etc/olake/gitops`.
- Source drivers at **v1.0.0 or later**. A job whose source is older fails reconcile until the Source is upgraded.
- `GITOPS_ENABLED=true` on the `olake-ui` Deployment (Helm `gitops.enabled: true`).
- GitOps RBAC: namespace `Role` on `configmaps` (get/list/watch/update/patch) and `events` (create/patch). **No `pods` permission on olake-ui.**

## Quick start

### 1. Enable GitOps

**Kubernetes (Helm):**

```yaml
gitops:
  enabled: true
```

See [olake-helm/docs/gitops.md](https://github.com/datazip-inc/olake-helm/blob/master/helm/olake/docs/gitops.md).

**Docker Compose:**

```yaml
environment:
  GITOPS_ENABLED: "true"
volumes:
  - ./olake-gitops:/etc/olake/gitops
```

olake-ui reads plain JSON files from `/etc/olake/gitops` — no ConfigMaps, labels or annotations. See [Docker Compose (file mode)](#docker-compose-file-mode).

### 2. Apply sample manifests

<!-- ### 2. Apply sample manifests

Replace `projectId` with your OLake project ID. Examples: [olake-helm/examples/gitops/](https://github.com/datazip-inc/olake-helm/tree/master/helm/olake/examples/gitops). -->

### 2. Argo CD / Flux health

Status is written to **annotations** on the managed object (not a status subresource):

| Annotation | Meaning |
|------------|---------|
| `olake.io/phase` | `Pending`, `Ready`, or `Failed` |
| `olake.io/message` | Human-readable detail |
| `olake.io/entity-id` | OLake entity ID after sync |
| `olake.io/observed-hash` | Content hash for skip logic |

Kubernetes Events: reason `Synced` (normal) or `SyncFailed` (warning).

**TODO:** Tool-specific health (Argo CD Lua, Flux) keyed off `olake.io/phase` — see olake-helm docs.

## ConfigMap reference

Label every managed object:

```yaml
metadata:
  labels:
    olake.io/managed: "true" # controller only watches objects with this label
    olake.io/kind: source   # source | destination | job | streams
```

| Kind | `data` keys | Purpose |
|------|-------------|---------|
| source | `projectId`, `userId`, `config` | Source connector JSON (POST `/sources` shape) |
| destination | `projectId`, `userId`, `config` | Destination connector JSON |
| job | `projectId`, `userId`, `config` | Job metadata; `config.source` / `destination` by name or numeric ID; `config.activate` defaults to `true` (set `false` to keep the job's schedule paused) |
| streams | `projectId`, `job`, `config` | Streams selection; `job` matches Job ConfigMap name or job entity ID |

### Streams catalog format

The Streams ConfigMap `data.config` JSON holds your selection only, in the `selected_streams.json` shape. OLake discovers the available streams from the source and merges your selection into them.

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

- The source driver must be **v1.0.0 or later**. Older sources fail the job.
- A full legacy catalog (with a `streams[]` array) is rejected. Keep only `selected_streams`.
- `selected_streams` must be non-empty.
- When the selection changes, OLake clears the destination for the streams that differ.
- Selections for streams that no longer exist in the source are dropped.

## Docker Compose (file mode)

Outside Kubernetes, olake-ui watches a directory of JSON files instead of ConfigMaps. Each file describes one entity with the same fields the OLake API takes, plus `kind`, `projectId`, `userId` and, for streams, `job`. There are no Kubernetes fields (`apiVersion`, `metadata`, labels, annotations).

### Files

**Source**

```json
{
  "kind": "source",
  "projectId": "123",
  "userId": 1,
  "name": "pg_prod",
  "type": "postgres",
  "version": "v1.0.0",
  "config": { "host": "postgres", "port": 5432, "database": "app", "username": "olake", "password": "..." }
}
```

**Destination**

```json
{
  "kind": "destination",
  "projectId": "123",
  "userId": 1,
  "name": "lake",
  "type": "iceberg",
  "version": "v1.0.0",
  "config": { "type": "ICEBERG", "writer": { "type": "glue", "iceberg_s3_path": "s3://bucket/warehouse", "aws_region": "ap-south-1" } }
}
```

**Job**: `source` and `destination` are the OLake name or numeric id (as a string) of the source and destination. `activate` defaults to `true`; `advanced_settings` is optional.

```json
{
  "kind": "job",
  "projectId": "123",
  "userId": 1,
  "name": "orders_sync",
  "source": "pg_prod",
  "destination": "lake",
  "frequency": "0 */6 * * *",
  "activate": true
}
```

**Streams**: the stream selection of one job, in the `selected_streams.json` shape (see [Streams catalog format](#streams-catalog-format)). `job` is the job's OLake name or numeric id.

```json
{
  "kind": "streams",
  "projectId": "123",
  "job": "orders_sync",
  "selected_streams": {
    "public": [
      { "stream_name": "orders", "sync_mode": "cdc", "normalization": true }
    ]
  }
}
```

### Rules

- **Kind:** `kind` is `source`, `destination`, `job` or `streams`.
- **Required fields:** `projectId` is required on every file. `userId` is required on source, destination and job files. Both may be a string or a number.
- **Identity:** a source, destination or job is identified by `kind` + `name` (the OLake name), and streams by their `job`. The file name and directory do not matter.
- **Files read:** every `*.json` file, at the root of the directory or in a directory under it (for example one directory per job). The directory holds only these files and the state file.
- **Editing:** a file is reconciled 5 s after its last change.
- **Deleting:** deleting a file drops the entity from GitOps, but the OLake entity is kept, as in Kubernetes mode.
- **Credentials:** there is no Secret equivalent; credentials are written in the source/destination file. Keep the directory private.

### Status

There are no annotations, so status (phase, message, entity ID, observed hash) is kept in `.olake-gitops-state.json` in the directory and written to the olake-ui log on each change. The file plays the role of the annotations: it keeps a restart from re-applying every job, which would cancel running syncs.

- Keep it on persistent storage, and add it to `.gitignore` if the directory is a git checkout.
- If the directory is read-only (for example a git-sync mount), set `GITOPS_STATE_FILE` to a writable path. olake-ui fails at startup if it cannot write the file.
- Deleting the state file, or a manifest, makes the next start reconcile those entities from scratch.

`GITOPS_FILE_DIR` changes the directory, for example when running the server outside Docker.

## Failure indicators

On permanent reconcile failure, olake-ui starts Temporal **`IndicatorWorkflow`**. **olake-worker** creates a short-lived busybox Pod (K8s) or container (Docker) labelled `olake.io/indicator=true`. The container writes the error to `/dev/termination-log` and exits 1.

Configure alerts on failed Pods with that label. olake-ui never creates Pods directly.

## Reconciliation behavior

- **Create/update only** — deleting a ConfigMap from git does **not** delete the OLake entity (v1).
- **Validation errors** → `phase=Failed`, indicator spawned, no requeue until content changes.
- **Failed job + fix** → no need to edit the job file. Changing the referenced source, destination, or Streams ConfigMap data re-runs a Failed job. A failed discover or connection test is `Failed` too: it retries on one of these edits, not on a timer.
- **Missing dependencies** → the Job waits for its source, destination, and Streams ConfigMap to exist. Until they do: `phase=Pending`, requeue every 30s, no indicator. No timeout; it waits as long as needed.
- **Transient errors** (Temporal or DB unreachable) → `phase=Pending`, requeue every 30s. After 30 minutes it becomes `Failed` with an indicator. To retry after fixing the cause: `kubectl annotate cm <name> olake.io/phase-`.
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

- olake-ui GitOps RBAC: ConfigMaps (read/write status annotations) and Events in the release namespace. **No pods create/delete.**
- Failure indicators use worker SA (already has `pods` create for connector jobs).
