# Declarative CLI Bundles

OLake UI can now reconcile jobs from a file bundle that matches the OLake CLI contract and export existing UI jobs back into the same bundle shape.

## Bundle layout

Required files:

- `source.json`
- `destination.json`
- `streams.json`

Optional files:

- `state.json`
- `olake-ui.json`

Example layout:

```text
mongodb-orders/
├── source.json
├── destination.json
├── streams.json
└── olake-ui.json
```

## `olake-ui.json`

`olake-ui.json` is the only server-side overlay. It carries the metadata that OLake UI needs but the raw CLI files do not include.

```json
{
  "apply_identity": "mongodb-orders",
  "job_name": "mongodb-orders-cdc",
  "source_name": "mongodb-orders-source",
  "source_type": "mongodb",
  "source_version": "v0.5.1",
  "destination_name": "parquet-minio-local",
  "destination_type": "parquet",
  "destination_version": "v0.5.1",
  "frequency": "0 * * * *",
  "activate": true
}
```

Notes:

- `source_type`, `source_version`, and `destination_version` are effectively required for server-side apply.
- `destination_type` can be omitted when `destination.json` has a top-level `type`.
- If names are omitted, the server infers deterministic defaults from the bundle name.
- If `state.json` is omitted, the server preserves existing UI state during apply.

## Apply API

Preview:

```bash
curl -sS -b cookies.txt \
  -F bundle=@mongodb-orders.zip \
  "http://localhost:8000/api/v1/project/123/jobs/apply-cli-bundle?dry_run=true" | jq
```

Apply:

```bash
curl -sS -b cookies.txt \
  -F bundle=@mongodb-orders.zip \
  "http://localhost:8000/api/v1/project/123/jobs/apply-cli-bundle" | jq
```

Apply is idempotent:

- same bundle + same current state = `unchanged`
- config drift = `updated`
- missing resource = `created`
- omitted `state.json` = existing state is `preserved`

## Export API

Export a UI job back into a CLI bundle:

```bash
curl -sS -b cookies.txt \
  "http://localhost:8000/api/v1/project/123/jobs/42/export-cli-bundle?format=zip" \
  -o mongodb-orders.zip
```

Include checkpoint state when you want a reconstruction bundle:

```bash
curl -sS -b cookies.txt \
  "http://localhost:8000/api/v1/project/123/jobs/42/export-cli-bundle?format=zip&include_state=true" \
  -o mongodb-orders-with-state.zip
```

## Example bundle

See [examples/cli-bundles/mongodb-parquet](/home/sabino/Downloads/olake-ui-fork/examples/cli-bundles/mongodb-parquet).
