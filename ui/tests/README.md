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

| Variable           | Set by  | Meaning                                                                                                                    |
| ------------------ | ------- | -------------------------------------------------------------------------------------------------------------------------- |
| `E2E_JOB_NAME`     | Go test | Name of the job; the source and destination names derive from it. Without it every test skips.                             |
| `E2E_DATABASE`     | Go test | Destination database the job writes to and the Go checks read (Dynamic: `<job>_postgres_public`). Needed by `@create-job`. |
| `E2E_SERVICE_HOST` | Go test | Address the sync driver uses to reach the source and destination ports (default `172.17.0.1`).                             |
| `E2E_WEBHOOK_URL`  | Go test | Webhook receiver URL for the `@alerts` step.                                                                               |

## Running

Run the whole workflow through Go (see `server/tests`), on this machine's Docker. It runs only on
a GitHub-hosted runner unless `OLAKE_TEST_RUNNER_ANY_HOST=true`, since the stack publishes its
ports on all interfaces:

```bash
cd server
OLAKE_TEST_RUNNER_ANY_HOST=true go test -v ./tests -timeout 0 -run '^TestPostgres$'
```

One step against a stack that is already up, after the steps before it:

```bash
cd ui
E2E_JOB_NAME=e2e_postgres_jdbc E2E_DATABASE=e2e_postgres_jdbc_postgres_public E2E_SERVICE_HOST=172.17.0.1 \
  npx playwright test tests/integration/postgres-iceberg.spec.ts --grep '@create-job'
```

Failure screenshots go to `ui/test-results/`. To log every browser action (Playwright's
`DEBUG=pw:api`), set `OLAKE_TEST_PW_DEBUG=true` for the Go test, or `DEBUG=pw:api` for a direct
Playwright run.
