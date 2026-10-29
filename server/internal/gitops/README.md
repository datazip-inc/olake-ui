# GitOps for OLake

Manage OLake sources, destinations, jobs, and stream selections as labelled **ConfigMaps** (or, for Source/Destination, **Secrets**) in git. Argo CD or Flux applies the manifests; an **embedded reconciler** inside `olake-ui` syncs each object into OLake.

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

<!-- ### 2. Apply sample manifests

Replace `projectId` with your OLake project ID. Examples: [olake-helm/examples/gitops/](https://github.com/datazip-inc/olake-helm/tree/master/helm/olake/examples/gitops). -->

### 2. Argo CD / Flux health

Status is written to **annotations** on the managed object (ConfigMap or Secret; not a status subresource):

| Annotation | Meaning |
|------------|---------|
| `olake.io/phase` | `Pending`, `Ready`, or `Failed` |
| `olake.io/message` | Human-readable detail |
| `olake.io/entity-id` | OLake entity ID after sync |
| `olake.io/observed-hash` | Content hash for skip logic |

Kubernetes Events: reason `Synced` (normal) or `SyncFailed` (warning).

**TODO:** Tool-specific health (Argo CD Lua, Flux) keyed off `olake.io/phase` — see olake-helm docs.

## ConfigMap / Secret reference

Label every managed object:

```yaml
metadata:
  labels:
    olake.io/managed: "true" # controller only watches objects with this label
    olake.io/kind: source   # source | destination | job | streams
```

| Kind | Object type | `data` keys | Purpose |
|------|-------------|-------------|---------|
| source | ConfigMap **or** Secret | `projectId`, `userId`, `config` | Source connector JSON (POST `/sources` shape) |
| destination | ConfigMap **or** Secret | `projectId`, `userId`, `config` | Destination connector JSON |
| job | ConfigMap only | `projectId`, `userId`, `config` | Job metadata; `config.source` / `destination` by name or numeric ID; `config.activate` defaults to `true` (set `false` to keep the job's schedule paused) |
| streams | ConfigMap only | `projectId`, `job`, `config` | Streams selection; `job` matches Job ConfigMap name or job entity ID |

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

### Source or Destination as a Secret

A Source or Destination can be defined as a **Secret** instead of a ConfigMap. The Secret carries the same `data` keys (`projectId`, `userId`, `config`) and the same `olake.io/managed` / `olake.io/kind` labels. Use one object type per Source/Destination name; if both exist for the same name, the ConfigMap wins.

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
      "version": "v1.0.0",
      "config": { "host": "...", "username": "...", "password": "..." }
    }
```

Editing the Secret's `data.config` re-triggers reconcile the same way editing a ConfigMap does. To keep the config in git and only the credentials out of it, use [secret references](#secret-references) instead.

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
- **Credentials:** use [secret references](#secret-references) (`${env:...}` or a cloud secret manager) instead of writing credentials in the file.

### Status

There are no annotations, so status (phase, message, entity ID, observed hash) is kept in `.olake-gitops-state.json` in the directory and written to the olake-ui log on each change. The file plays the role of the annotations: it keeps a restart from re-applying every job, which would cancel running syncs.

- Keep it on persistent storage, and add it to `.gitignore` if the directory is a git checkout.
- If the directory is read-only (for example a git-sync mount), set `OLAKE_GITOPS_STATE_FILE` to a writable path. olake-ui fails at startup if it cannot write the file.
- Deleting the state file, or a manifest, makes the next start reconcile those entities from scratch.

`OLAKE_GITOPS_FILE_DIR` changes the directory, for example when running the server outside Docker.
File mode also works inside Kubernetes, for example with manifests mounted from a volume: set `GITOPS_MODE=file` instead of `kubernetes`.

## Secret references

Keep credentials out of git: set a string field in a source or destination `config` to a reference instead of the value. olake-ui fetches the value when it reconciles, and stores the resolved config encrypted in its database. The value is never written back to the ConfigMap or file.

```json
"config": {
  "host": "pg.prod.internal",
  "username": "${aws-sm:olake/prod/pg#username}",
  "password": "${aws-sm:olake/prod/pg#password}"
}
```

Format: `${<provider>:<path>[#<key>]}`

- The reference must be the whole value. `"x${...}"` is a plain string.
- `#key` reads one key of a secret that holds a JSON object, such as `{"username":"...","password":"..."}`. The value keeps its JSON type: with `{"port":5432}`, `"port": "${...#port}"` becomes the number `5432`, while `{"port":"5432"}` gives the string `"5432"`. Without `#key`, the whole secret value is used as a string.
- References work in nested objects too, such as `ssh_config.password` or `writer.aws_secret_key`.

| Provider | Reference | Works in |
|----------|-----------|----------|
| AWS Secrets Manager | `${aws-sm:<name or ARN>[#key]}` | Kubernetes, Docker |
| GCP Secret Manager | `${gcp-sm:<secret-id>[#key]}` or `${gcp-sm:projects/<project>/secrets/<id>[/versions/<v>][#key]}` | Kubernetes, Docker |
| Azure Key Vault | `${azure-kv:<vault-name>/<secret-name>[/<version>][#key]}` | Kubernetes, Docker |
| Kubernetes Secret | `${secrets:<secret-name>#<key>}` | Kubernetes mode only |
| Environment variable | `${env:GITOPS_SECRET_<NAME>}` | Kubernetes, Docker |

**Requirements and behaviour**

- **`OLAKE_SECRET_KEY`** must be set on olake-ui and the worker, because resolved values are stored in the database. Without it, a resource with references fails. On AWS, prefer a KMS key ARN as its value.
- **Rotation:** a `Ready` resource with references is re-checked every `GITOPS_SECRETS_REFRESH_INTERVAL` (a duration with a unit, such as `15m`; default `1h`, minimum `1m`, `0` turns it off). A changed secret is connection-tested and applied without any change in git. To re-check now, edit the object, for example `kubectl annotate cm <name> olake.io/refresh="$(date +%s)" --overwrite`.
- **Errors:** a missing secret or key, or access denied, makes the resource `Failed` with the reason. Network errors and throttling are retried like other transient errors.
- **Settings** go on olake-ui as environment variables. In Helm use `olakeUI.env`; in Docker Compose use the olake-ui service's `environment`.

### AWS Secrets Manager

Authentication, preferred first:

1. **EKS Pod Identity.** Associate the role with olake-ui's ServiceAccount:
   ```bash
   kubectl -n <namespace> get deploy olake-ui -o jsonpath='{.spec.template.spec.serviceAccountName}'
   aws eks create-pod-identity-association --cluster-name <cluster> \
     --namespace <namespace> --service-account <service-account> --role-arn <role-arn>
   ```
2. **IRSA.** Annotate olake-ui's ServiceAccount with `eks.amazonaws.com/role-arn: <role-arn>`. By default olake-ui runs as the job ServiceAccount, so set it in `global.jobServiceAccount.annotations`. A ServiceAccount has one role: if S3 log storage also uses IRSA, give one role both permissions.
3. **EC2 instance profile or ECS task role** (Docker on AWS).
4. **Access keys** (testing only): `GITOPS_SECRETS_AWS_ACCESS_KEY_ID` and `GITOPS_SECRETS_AWS_SECRET_ACCESS_KEY`.

Role permission:

```json
{
  "Effect": "Allow",
  "Action": "secretsmanager:GetSecretValue",
  "Resource": "arn:aws:secretsmanager:<region>:<account>:secret:olake/*"
}
```

Add `kms:Decrypt` on the key if the secret uses a customer-managed KMS key.

Settings: `GITOPS_SECRETS_AWS_REGION` is needed when references use secret names. An ARN carries its own region.

### GCP Secret Manager

Authentication, preferred first:

1. **GKE Workload Identity.** Grant the role on the secret to olake-ui's ServiceAccount:
   ```bash
   gcloud secrets add-iam-policy-binding <secret-id> --role=roles/secretmanager.secretAccessor \
     --member="principal://iam.googleapis.com/projects/<project-number>/locations/global/workloadIdentityPools/<project-id>.svc.id.goog/subject/ns/<namespace>/sa/<service-account>"
   ```
2. **Attached service account** (Docker on a GCE VM).
3. **Workload Identity Federation** from a non-GKE cluster (EKS, AKS): mount the credential configuration file and set `GOOGLE_APPLICATION_CREDENTIALS` to its path.
4. **Service account JSON key** (testing only): mount the key file and set `GOOGLE_APPLICATION_CREDENTIALS` to its path. Many organisations block key creation.

Permission: `roles/secretmanager.secretAccessor` on each secret.

Settings: `GITOPS_SECRETS_GCP_PROJECT` is needed for short secret ids outside GKE or GCE. On GKE and GCE the project comes from the metadata server.

The latest version is read unless the reference pins one. To rotate, add a new version.

### Azure Key Vault

Authentication, preferred first:

1. **AKS Workload Identity.** Create a user-assigned managed identity and a federated credential for `system:serviceaccount:<namespace>:<service-account>`. Then annotate olake-ui's ServiceAccount with `azure.workload.identity/client-id: <client-id>`, and add the pod label `azure.workload.identity/use: "true"` with `olakeUI.podLabels`.
2. **Managed identity** (Docker on an Azure VM).
3. **Service principal client secret** (testing only): `GITOPS_SECRETS_AZURE_TENANT_ID`, `GITOPS_SECRETS_AZURE_CLIENT_ID` and `GITOPS_SECRETS_AZURE_CLIENT_SECRET`.

Permission: the **Key Vault Secrets User** role on the vault, which must use the Azure RBAC permission model.

The vault URL is `https://<vault-name>.vault.azure.net`. The latest version is read unless the reference pins one. To rotate, add a new version.

### Kubernetes Secret

Kubernetes mode only. The Secret must be in olake-ui's namespace and labelled `olake.io/managed=true`. Other Secrets cannot be read.

```bash
kubectl -n <namespace> create secret generic pg-creds \
  --from-literal=username=olake --from-literal=password='<password>'
kubectl -n <namespace> label secret pg-creds olake.io/managed=true
```

```json
"password": "${secrets:pg-creds#password}"
```

The Helm chart's GitOps Role lets olake-ui read Secrets in its namespace. The Secret can come from External Secrets Operator, Sealed Secrets or the Vault Secrets Operator; add the label to the Secret they create.

### Environment variable

Only variables whose name starts with `GITOPS_SECRET_` can be read, so a config cannot read olake-ui's own settings such as `OLAKE_SECRET_KEY`. Docker Compose, with the values in an uncommitted `.env` file:

```yaml
olake-ui:
  environment:
    GITOPS_SECRET_PG_PASSWORD: ${GITOPS_SECRET_PG_PASSWORD}
```

```json
"password": "${env:GITOPS_SECRET_PG_PASSWORD}"
```

A changed value takes effect after olake-ui restarts.

## Failure indicators

On permanent reconcile failure, olake-ui starts Temporal **`IndicatorWorkflow`**. **olake-worker** creates a short-lived busybox Pod (K8s) or container (Docker) labelled `olake.io/indicator=true`. The container writes the error to `/dev/termination-log` and exits 1.

Configure alerts on failed Pods with that label. olake-ui never creates Pods directly.

## Reconciliation behavior

- **Create/update only** — deleting a ConfigMap or Secret from git does **not** delete the OLake entity (v1).
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
| Argo synced but phase empty | Operator may be disabled; check `GITOPS_MODE=kubernetes` |

```bash
kubectl describe configmap my-postgres
kubectl get events --field-selector involvedObject.name=my-postgres
kubectl logs -l app.kubernetes.io/name=olake-ui | grep -i gitops
```

## Security notes

- olake-ui GitOps RBAC: ConfigMaps and Secrets (read/write status annotations on both), Events in the release namespace. **No pods create/delete.**
- Secret `update`/`patch` (needed for status annotations) applies to all Secrets in the namespace, not only GitOps-managed ones — use a dedicated namespace for OLake GitOps.
- Secret references: olake-ui reads only Secrets labelled `olake.io/managed=true` and environment variables starting with `GITOPS_SECRET_`. For cloud secret managers, scope olake-ui's IAM permission to the secrets it needs, because anyone who can write a config can reference any secret that identity can read.
- Failure indicators use worker SA (already has `pods` create for connector jobs).
