# GitOps for OLake

Manage OLake sources, destinations, jobs, and stream selections as labelled **ConfigMaps** (or, for Source/Destination, **Secrets**) in git. Argo CD or Flux applies the manifests; an **embedded reconciler** inside `olake-ui` syncs each object into OLake by calling the ETL service layer (same logic as the HTTP API).

## Prerequisites

- Kubernetes: OLake UI deployed **inside** the cluster (in-cluster kubeconfig), or Docker Compose with file-based manifests mounted at `/etc/olake/gitops`.
- Source drivers at **v1.0.0 or later**. A job whose source is older fails reconcile until the Source is upgraded.
- `GITOPS_MODE` on `olake-ui`: `kubernetes` for ConfigMaps/Secrets (Helm `gitops.enabled: true` sets it), or `file` for a directory of JSON manifests.
  Unset or empty disables GitOps.
  Any other value fails startup.
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

**Docker Compose:**

```yaml
environment:
  GITOPS_MODE: file
volumes:
  - ./olake-gitops:/etc/olake/gitops
```

olake-ui reads plain JSON files from `/etc/olake/gitops` — no ConfigMaps, labels or annotations. See [Docker Compose (file mode)](#docker-compose-file-mode).

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
| job | ConfigMap only | `projectId`, `userId`, `config` | Job metadata; `config.source` / `destination` by name or numeric ID; `config.activate` defaults to `true` (set `false` to keep the job's schedule paused) |
| streams | ConfigMap only | `projectId`, `job`, `config` | Streams catalog; `job` matches Job ConfigMap name or job entity ID |

### Streams catalog format

The Streams ConfigMap `data.config` JSON holds the user's choices only, in the `selected_streams.json` shape:

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

The job runs in the split format: the catalog is stored as `available_streams_config` + `selected_streams_config` (`streams_config` stays NULL) and passed with `--available-streams` / `--selected-streams`.

- A CM in the legacy `streams.json` shape (with `streams[]`) is rejected. Remove `streams[]` and keep `selected_streams`.
- On create the operator runs discover against the source for `available_streams_config`, then discover again against it and the CM's `selected_streams`, so the stored selection comes from the CLI's merge (same as on update).
- On update with a changed CM (or a changed source/destination), discover merges against the stored `available_streams_config` and the CM's `selected_streams`; the CLI's merge drops selections whose stream no longer exists, updates `selected_columns` when `sync_new_columns` is true, and refreshes `streams[]` when the source schema changes.
- An update that does not change the CM (a Job CM edit, or a changed source/destination) starts from the job's stored catalog, not the CM's. A changed source/destination runs discover against that stored catalog.
- `selected_streams` must be non-empty (deselect-all fails reconcile).
- A changed CM triggers a stream difference between the stored catalog and the newly merged one, and clear-destination for the streams that differ.

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

## Docker Compose (file mode)

With `GITOPS_MODE=file`, olake-ui watches a directory of JSON files instead of ConfigMaps. Each file describes one entity with the same fields the OLake API takes, plus `kind`, `projectId`, `userId` and, for streams, `job`. There are no Kubernetes fields (`apiVersion`, `metadata`, labels, annotations).

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
File mode also works inside Kubernetes, for example with manifests mounted from a volume: set `GITOPS_MODE=file` instead of `kubernetes`.

## Failure indicators

On permanent reconcile failure, olake-ui starts Temporal **`IndicatorWorkflow`** (fire-and-forget). **olake-worker** creates a short-lived busybox Pod (K8s) or container (Docker) labelled `olake.io/indicator=true`. The container writes the error to `/dev/termination-log` and exits 1.

Configure alerts on failed Pods with that label. olake-ui never creates Pods directly.

## Reconciliation behavior

- **Create/update only** — deleting a ConfigMap or Secret from git does **not** delete the OLake entity (v1).
- **Validation errors** → `phase=Failed`, indicator spawned, no requeue until content changes.
- **Failed job + source/destination fix** → retrying a Failed job does **not** require editing the job file. Changing the referenced source or destination (ConfigMap or Secret data) invalidates the skip hash and re-runs the job. Streams are not part of that hash (catalogs are large); the job checks the Streams CM's own hash instead, so editing its data also re-runs a Failed job. A failed discover or connection test is `Failed` too: it retries on one of these edits, not on a timer.
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
| Argo synced but phase empty | Operator may be disabled; check `GITOPS_MODE=kubernetes` |

```bash
kubectl describe configmap my-postgres
kubectl get events --field-selector involvedObject.name=my-postgres
kubectl logs -l app.kubernetes.io/name=olake-ui | grep -i gitops
```

## Security notes

- olake-ui GitOps RBAC: ConfigMaps and Secrets (read/write status annotations on both), Events in the release namespace. **No pods create/delete.**
- Secret `update`/`patch` (needed for status annotations) applies to all Secrets in the namespace, not only GitOps-managed ones — use a dedicated namespace for GitOps-managed OLake when possible.
- Failure indicators use worker SA (already has `pods` create for connector jobs).
