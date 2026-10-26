# UI integration tests

Playwright tests that drive the OLake UI against a running stack. They are not run on their own:
the Go integration test in `server/tests` starts the stack (source, destination, OLake), then
runs one tagged test of a spec at a time, with its inputs in `E2E_*` variables, and checks the
destination data between the steps.

## Layout

```
tests/
├── integration/
│   ├── postgres-iceberg.spec.ts   Postgres → Iceberg (JDBC) workflow, one test per step (@create-job, @first-sync, ...)
│   └── helpers.ts                 UI actions the workflow uses (forms, stream settings, history, logs)
├── setup/auth.setup.ts            logs in once and saves the session to tests/.auth/user.json
├── fixtures/                      page-object fixtures; testAuthenticated reuses the saved session
├── pages/                         page objects for login, sources, destinations, jobs and the job wizard
├── utils/                         connector configs, test-connection modal checks, select helpers
├── enums/                         connector and catalog names as the UI shows them
└── types/                         form config types
```

## Inputs

| Variable | Set by | Meaning |
|---|---|---|
| `E2E_JOB_NAME` | Go test | Name of the job; the source, destination and database names derive from it. Without it every test skips. |
| `E2E_SERVICE_HOST` | Go test | Address the sync driver uses to reach the source and destination ports (default `172.17.0.2`). |
| `E2E_WEBHOOK_URL` | Go test | Webhook receiver URL for the `@alerts` step. |

## Running

Run the whole workflow through Go (see `server/tests`):

```bash
cd server
go test -v ./tests -timeout 0 -run '^TestPostgres$'                    # DinD, e.g. on the test server
OLAKE_TEST_ENV=runner OLAKE_TEST_RUNNER_ANY_HOST=true \
  go test -v ./tests -timeout 0 -run '^TestPostgres$'                  # on this machine's Docker
```

One step against a stack that is already up, after the steps before it:

```bash
cd ui
E2E_JOB_NAME=e2e_postgres_jdbc E2E_SERVICE_HOST=172.17.0.1 \
  npx playwright test tests/integration/postgres-iceberg.spec.ts --grep '@create-job'
```

Failure screenshots go to `ui/test-results/`.
