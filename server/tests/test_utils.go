package tests

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/apache/spark-connect-go/v35/spark/sql"
	"github.com/stretchr/testify/require"

	"github.com/datazip-inc/olake-ui/server/tests/destinations"
	"github.com/datazip-inc/olake-ui/server/tests/sources"
)

// olakeComposeFiles starts OLake from the root docker-compose.yml, with olake-ui taken from the
// olake-ui:ci image built from this checkout (see compose.olake-ui.yml). The project has its own
// name, so taking it down never touches an OLake a developer runs from the same checkout, whose
// project is named after the directory.
const olakeComposeFiles = "-p olake-app-test -f docker-compose.yml -f server/tests/compose.olake-ui.yml"

// shellQuote quotes s as one word for /bin/sh. The checkout path in runner mode can hold spaces
// or shell characters, so every interpolated value goes through it.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// Build the olake-ui:ci image from the checkout
func buildOLakeUICmd(root string) string {
	return fmt.Sprintf("cd %s && docker compose %s build olake-ui", shellQuote(root), olakeComposeFiles)
}

// Start OLake application
func startOLakeCmd(root string) string {
	return fmt.Sprintf(`
        cd %[1]s &&
        mkdir -p olake-data &&
        docker compose %[2]s up -d &&
        for i in $(seq 1 60); do
            if curl -f http://localhost:8000/health 2>/dev/null || curl -f http://localhost:8000 2>/dev/null; then
                echo "OLake UI ready."
                break
            fi
            sleep 2
        done
    `, shellQuote(root), olakeComposeFiles)
}

// Install Playwright and dependencies. --frozen-lockfile installs exactly what pnpm-lock.yaml
// pins, leaves package.json and the lockfile in the checkout unchanged, and fails if they disagree.
func installPlaywrightCmd(root string) string {
	return fmt.Sprintf(`
        cd %s &&
        pnpm install --frozen-lockfile &&
        pnpm exec playwright install --with-deps chromium
    `, shellQuote(root+"/ui"))
}

// playwrightPatterns limit the values Run.Playwright puts in the Playwright shell command.
var (
	jobPattern    = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	specPattern   = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_./-]*\.spec\.ts$`) // no leading '-': Playwright would read an option
	tagPattern    = regexp.MustCompile(`^@[a-z0-9][a-z0-9-]*$`)
	envKeyPattern = regexp.MustCompile(`^E2E_[A-Z0-9_]+$`)
)

// playwrightCmd runs the tests of spec (relative to ui/) tagged tag, or all of them when tag is
// empty, with env set for the Playwright process.
func playwrightCmd(root, spec, tag string, env map[string]string) (string, error) {
	if !specPattern.MatchString(spec) {
		return "", fmt.Errorf("invalid spec path %q", spec)
	}
	if tag != "" && !tagPattern.MatchString(tag) {
		return "", fmt.Errorf("invalid tag %q", tag)
	}
	keys := make([]string, 0, len(env))
	for k := range env {
		if !envKeyPattern.MatchString(k) {
			return "", fmt.Errorf("invalid Playwright env name %q", k)
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var vars strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&vars, "%s=%s ", k, shellQuote(env[k]))
	}
	grep := ""
	if tag != "" {
		// A tag is matched as a whole word, so @sync does not also pick @sync-again.
		grep = " --grep " + shellQuote(tag+`(?![a-z0-9-])`)
	}
	return fmt.Sprintf(`
        cd %s &&
        %sPLAYWRIGHT_TEST_BASE_URL=http://localhost:8000 DEBUG=pw:api npx playwright test %s%s
    `, shellQuote(root+"/ui"), vars.String(), shellQuote(spec), grep), nil
}

// sparkConnectPort is the port the checks reach Spark Connect on. Every destination's compose
// file must run Spark Connect there; see destinations.Destination.
const sparkConnectPort = "15002/tcp"

// runTimeout bounds a whole workflow run. It is under both CI jobs' timeouts, with room for their
// setup steps and the cleanup.
const runTimeout = 90 * time.Minute

// Workflow is the source and destination of an integration run, and the name of the job the
// Playwright specs create.
type Workflow struct {
	Source      sources.Source
	Destination destinations.Destination
	Job         string // passed to the specs as E2E_JOB_NAME
}

// database is the destination database OLake writes the job to: <job>_<source namespace>, lowercased.
func (w *Workflow) database() string {
	return strings.ToLower(w.Job + "_" + w.Source.Namespace)
}

// composeStartCmd starts the compose file in dir (relative to server/tests under root) as project
// olake-<project>-test, runs ready in a subshell (skipped when empty), then lists the services.
// project is "source" or "destination", so the two never share a project whatever their names.
func composeStartCmd(root, project, dir string, services []string, ready string) string {
	name := fmt.Sprintf("olake-%s-test", project)
	cmd := fmt.Sprintf("cd %s && docker compose -p %s up -d %s", shellQuote(root+"/server/tests/"+dir), name, strings.Join(services, " "))
	if strings.TrimSpace(ready) != "" {
		cmd += " && (" + ready + ")"
	}
	return cmd + " && docker compose -p " + name + " ps"
}

// Run is a started integration stack: the Env with the source, the destination and OLake up,
// the source seeded and Playwright installed. Its methods fail the test on error.
type Run struct {
	t           *testing.T
	ctx         context.Context
	env         Env
	w           *Workflow
	start       time.Time
	ServiceHost string // what the driver, and the specs, use to reach the published ports

	sourceHost, sourcePort string
	sparkHost, sparkPort   string
}

// StartWorkflow starts the Env, the source, the destination and OLake, and seeds the source.
func StartWorkflow(t *testing.T, w *Workflow) *Run {
	t.Helper()
	if !jobPattern.MatchString(w.Job) {
		t.Fatalf("invalid job name %q", w.Job)
	}
	ctx, cancel := context.WithTimeout(context.Background(), runTimeout)
	r := &Run{t: t, ctx: ctx, w: w, start: time.Now()}

	ports := []string{sparkConnectPort}
	if w.Source.Port != "" {
		ports = append(ports, w.Source.Port)
	}
	env, err := newEnv(ctx, t, ports)
	t.Cleanup(cancel) // registered after the Env's own cleanup, so it runs before it
	r.require(err)
	r.env = env
	root := env.Root()

	// The address the driver uses for the source and destination; checked before anything starts.
	r.ServiceHost, err = env.ServiceHost(ctx)
	r.require(err)
	t.Logf("Service host: %s", r.ServiceHost)

	r.sparkHost, r.sparkPort, err = env.Addr(ctx, sparkConnectPort)
	r.require(err, "failed to get spark port")
	r.sourceHost = r.sparkHost
	if w.Source.Port != "" {
		r.sourceHost, r.sourcePort, err = env.Addr(ctx, w.Source.Port)
		r.require(err, "failed to get %s port", w.Source.Name)
	}

	r.Step("Starting %s source...", w.Source.Name)
	r.exec(composeStartCmd(root, "source", w.Source.ComposeDir, nil, w.Source.ReadyCmd), "%s startup failed", w.Source.Name)

	r.Step("Starting %s destination...", w.Destination.Name)
	r.exec(composeStartCmd(root, "destination", w.Destination.ComposeDir, w.Destination.Services, w.Destination.ReadyCmd), "destination services startup failed")

	// Build olake-ui from this checkout, unless the workflow already built olake-ui:ci
	if os.Getenv("OLAKE_TEST_SKIP_UI_BUILD") == "true" {
		r.Step("Using the prebuilt olake-ui image...")
		r.exec("docker image inspect --format '{{.Id}}' olake-ui:ci", "OLAKE_TEST_SKIP_UI_BUILD is set but olake-ui:ci is missing")
	} else {
		r.Step("Building the olake-ui image...")
		r.exec(buildOLakeUICmd(root), "olake-ui image build failed")
	}

	r.Step("Starting OLake docker-compose services...")
	r.exec(startOLakeCmd(root), "OLake startup failed")

	r.Step("Seeding the %s source...", w.Source.Name)
	w.Source.Seed(ctx, t, r.sourceHost, r.sourcePort)

	r.Step("Installing Playwright and dependencies...")
	r.require(env.InstallPlaywright(ctx, t))
	t.Logf("OLake UI is ready and accessible at: http://localhost:8000")
	return r
}

// Step logs a step with the time since the run started.
func (r *Run) Step(format string, args ...any) {
	r.t.Helper()
	r.t.Logf("[%6.1fs] "+format, append([]any{time.Since(r.start).Seconds()}, args...)...)
}

// require fails the test when err is set.
func (r *Run) require(err error, msgAndArgs ...any) {
	r.t.Helper()
	if err != nil {
		require.NoError(r.t, err, msgAndArgs...)
	}
}

// exec runs cmd in the Env and fails the test with its output when it does not exit 0.
func (r *Run) exec(cmd, format string, args ...any) []byte {
	r.t.Helper()
	code, out, err := r.env.Exec(r.ctx, cmd)
	if err != nil || code != 0 {
		r.t.Fatalf("%s (%d): %v\n%s", fmt.Sprintf(format, args...), code, err, out)
	}
	return out
}

// Playwright runs the tests of spec tagged tag. Every run gets E2E_JOB_NAME and
// E2E_SERVICE_HOST; env adds more E2E_* variables.
func (r *Run) Playwright(spec, tag string, env map[string]string) {
	r.t.Helper()
	vars := map[string]string{"E2E_JOB_NAME": r.w.Job, "E2E_SERVICE_HOST": r.ServiceHost}
	for k, v := range env {
		vars[k] = v
	}
	cmd, err := playwrightCmd(r.env.Root(), spec, tag, vars)
	r.require(err)
	r.Step("Playwright %s %s...", spec, tag)
	code, out, err := r.env.ExecStream(r.ctx, r.t, cmd)
	if err != nil || code != 0 {
		r.t.Fatalf("playwright %s %s failed (%d): %v\n%s", spec, tag, code, err, out)
	}
}

// SourceSQL runs statements on the source, in order.
func (r *Run) SourceSQL(statements ...string) {
	r.t.Helper()
	r.Step("Changing the %s source...", r.w.Source.Name)
	r.w.Source.Exec(r.ctx, r.t, r.sourceHost, r.sourcePort, statements...)
}

// Check runs checks on the destination as subtests of a subtest named phase, through one Spark
// Connect session. It fails the test if any check fails.
func (r *Run) Check(phase string, checks ...Check) {
	r.t.Helper()
	r.Step("Checking the destination: %s...", phase)
	ok := r.t.Run(phase, func(t *testing.T) {
		spark, err := sql.NewSessionBuilder().Remote(fmt.Sprintf("sc://%s:%s", r.sparkHost, r.sparkPort)).Build(r.ctx)
		require.NoError(t, err, "Failed to connect to Spark Connect server")
		defer func() {
			if stopErr := spark.Stop(); stopErr != nil {
				t.Errorf("Failed to stop Spark session: %v", stopErr)
			}
		}()
		table := func(name string) string { return r.w.Destination.Table(r.w.database(), name) }
		runChecks(r.ctx, t, spark, table, checks)
	})
	if !ok {
		r.t.FailNow()
	}
}

// webhookContainer receives OLake's alert webhook and prints each request body to its log.
const webhookContainer = "olake-webhook-test"

// webhookReceiver is a minimal HTTP server that prints every POST body.
const webhookReceiver = `
import http.server
class H(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        n = int(self.headers.get("Content-Length") or 0)
        print(self.rfile.read(n).decode(errors="replace"), flush=True)
        self.send_response(200)
        self.end_headers()
    def log_message(self, *args):
        pass
http.server.HTTPServer(("", 8099), H).serve_forever()
`

// StartWebhookReceiver starts the receiver, published on 8099, and returns the URL OLake posts
// alerts to.
func (r *Run) StartWebhookReceiver() string {
	r.t.Helper()
	r.Step("Starting the webhook receiver...")
	r.exec(fmt.Sprintf("docker rm -f %[1]s >/dev/null 2>&1; docker run -d --name %[1]s -p 8099:8099 python:3.12-alpine python -u -c %[2]s",
		webhookContainer, shellQuote(webhookReceiver)), "webhook receiver startup failed")
	return fmt.Sprintf("http://%s:8099/alert", r.ServiceHost)
}

// WaitForWebhook waits until the receiver has logged a request containing text.
func (r *Run) WaitForWebhook(text string, timeout time.Duration) {
	r.t.Helper()
	r.Step("Waiting for the webhook %q...", text)
	deadline := time.Now().Add(timeout)
	for {
		out := r.exec("docker logs "+webhookContainer+" 2>&1", "reading the webhook receiver log failed")
		if strings.Contains(string(out), text) {
			r.t.Logf("✅ webhook received %q", text)
			return
		}
		if time.Now().After(deadline) {
			r.t.Fatalf("no webhook with %q after %s; receiver log:\n%s", text, timeout, out)
		}
		time.Sleep(5 * time.Second)
	}
}

// StopSource stops the source container, to make the next sync fail.
func (r *Run) StopSource() {
	r.t.Helper()
	r.Step("Stopping the %s source...", r.w.Source.Name)
	r.exec("docker stop "+shellQuote(r.w.Source.Container), "stopping the source failed")
}
