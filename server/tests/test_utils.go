package tests

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/apache/spark-connect-go/v35/spark/sql"
	"github.com/stretchr/testify/require"

	"github.com/datazip-inc/olake-ui/server/tests/destinations"
	"github.com/datazip-inc/olake-ui/server/tests/sources"
)

// olakeComposeFiles starts OLake from the root docker-compose-v1.yml, with olake-ui taken from the
// olake-ui:ci image built from this checkout (see compose.olake-ui.yml). The Fusion services in
// that file are in the "fusion" profile, which the test does not enable. The project has its own
// name, so taking it down never touches an OLake a developer runs from the same checkout, whose
// project is named after the directory.
const olakeComposeFiles = "-p olake-app-test -f docker-compose-v1.yml -f server/tests/compose.olake-ui.yml"

// olakeHome, under the checkout, stands in for PWD in the root docker-compose-v1.yml, whose job
// state lives in ${PWD}/olake-data. The test's state then stays out of the olake-data folder of
// an OLake a developer runs from the same checkout. It is gitignored like any olake-data folder.
const olakeHome = "server/tests/.olake-home"

// clearDirCmd empties dir, creating it if needed. It runs rm in a container because files the
// stack's containers wrote can belong to root.
func clearDirCmd(dir string) string {
	return fmt.Sprintf("mkdir -p %[1]s && docker run --rm -v %[1]s:/clear alpine:3.20 find /clear -mindepth 1 -delete", shellQuote(dir))
}

// olakeCompose is the docker compose command for the test's OLake project, run from root.
func olakeCompose(root string) string {
	return fmt.Sprintf("PWD=%s docker compose %s", shellQuote(root+"/"+olakeHome), olakeComposeFiles)
}

// shellQuote quotes s as one word for /bin/sh. The checkout path in runner mode can hold spaces
// or shell characters, so every interpolated value goes through it.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// Build the olake-ui:ci image from the checkout
func buildOLakeUICmd(root string) string {
	return fmt.Sprintf("cd %s && %s build olake-ui", shellQuote(root), olakeCompose(root))
}

// Start OLake application
func startOLakeCmd(root string) string {
	return fmt.Sprintf(`
        cd %[1]s &&
        mkdir -p %[2]s/olake-data &&
        %[3]s up -d &&
        for i in $(seq 1 60); do
            if curl -f http://localhost:8000/health 2>/dev/null || curl -f http://localhost:8000 2>/dev/null; then
                echo "OLake UI ready."
                break
            fi
            sleep 2
        done
    `, shellQuote(root), shellQuote(olakeHome), olakeCompose(root))
}

// playwrightPatterns limit the values Run.Playwright puts in the Playwright shell command.
var (
	jobPattern    = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	specPattern   = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_./-]*\.spec\.ts$`) // no leading '-': Playwright would read an option
	tagPattern    = regexp.MustCompile(`^@[a-z0-9][a-z0-9-]*$`)
	envKeyPattern = regexp.MustCompile(`^E2E_[A-Z0-9_]+$`)
)

// playwrightCmd runs the tests of spec (relative to ui/) tagged tag, or all of them when tag is
// empty, with env set for the Playwright process. debug logs every browser action (DEBUG=pw:api).
func playwrightCmd(root, spec, tag string, env map[string]string, debug bool) (string, error) {
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
	if debug {
		vars.WriteString("DEBUG=pw:api ")
	}
	grep := ""
	if tag != "" {
		// A tag is matched as a whole word, so @sync does not also pick @sync-again.
		grep = " --grep " + shellQuote(tag+`(?![a-z0-9-])`)
	}
	return fmt.Sprintf(`
        cd %s &&
        %sPLAYWRIGHT_TEST_BASE_URL=http://localhost:8000 npx playwright test %s%s
    `, shellQuote(root+"/ui"), vars.String(), shellQuote(spec), grep), nil
}

// sparkConnectPort is the port the checks reach Spark Connect on. Every destination's compose
// file must run Spark Connect there; see destinations.Destination.
const sparkConnectPort = "15002/tcp"

// runTimeout bounds a whole workflow run. It is under the CI job's timeout, with room for its
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
	start := fmt.Sprintf("docker compose -p %s up -d %s", name, strings.Join(services, " "))
	if strings.TrimSpace(ready) != "" {
		start += " && (" + ready + ")"
	}
	start += " && docker compose -p " + name + " ps"
	// On failure, print the containers' logs: the cleanup removes the containers, and with them
	// the only record of why one exited.
	return fmt.Sprintf("cd %s && { %s; } || { echo '--- %s container logs ---'; docker compose -p %s logs --no-color --tail 40; exit 1; }",
		shellQuote(root+"/server/tests/"+dir), start, name, name)
}

// Run is a started integration stack: the source, the destination and OLake up, the source
// seeded and Playwright installed. Its methods fail the test on error.
type Run struct {
	t           *testing.T
	ctx         context.Context
	env         *Env
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

	env, err := newEnv(t)
	t.Cleanup(cancel) // registered after the Env's own cleanup, so it runs before it
	r.require(err)
	r.env = env
	root := env.Root()

	// The address the driver uses for the source and destination; checked before anything starts.
	r.ServiceHost, err = env.ServiceHost(ctx)
	r.require(err)
	t.Logf("Service host: %s", r.ServiceHost)

	r.sparkHost, r.sparkPort = env.Addr(sparkConnectPort)
	r.sourceHost = r.sparkHost
	if w.Source.Port != "" {
		r.sourceHost, r.sourcePort = env.Addr(w.Source.Port)
	}

	// A checkout that ran the test before still holds that run's Iceberg tables and job state;
	// a new run must not read them.
	r.Step("Clearing data left by earlier runs...")
	r.exec(clearDirCmd(root+"/"+olakeHome), "clearing %s failed", olakeHome)
	if w.Destination.DataDir != "" {
		r.exec(clearDirCmd(root+"/server/tests/"+w.Destination.ComposeDir+"/"+w.Destination.DataDir), "clearing the %s data failed", w.Destination.Name)
	}

	// The source, the destination, OLake and Playwright do not depend on each other, so they start
	// at the same time. The first that fails cancels the others.
	startCtx, cancelStart := context.WithCancel(ctx)
	var tasks sync.WaitGroup
	// Also when Seed fails the test below: no task may outlive StartWorkflow and log after the
	// test has ended.
	defer func() {
		cancelStart()
		tasks.Wait()
	}()
	source := r.task(startCtx, cancelStart, &tasks, "source "+w.Source.Name, func(ctx context.Context) error {
		return r.run(ctx, composeStartCmd(root, "source", w.Source.ComposeDir, nil, w.Source.ReadyCmd), "%s startup failed", w.Source.Name)
	})
	destination := r.task(startCtx, cancelStart, &tasks, "destination "+w.Destination.Name, func(ctx context.Context) error {
		return r.run(ctx, composeStartCmd(root, "destination", w.Destination.ComposeDir, w.Destination.Services, w.Destination.ReadyCmd), "destination services startup failed")
	})
	olake := r.task(startCtx, cancelStart, &tasks, "OLake", func(ctx context.Context) error {
		// Build olake-ui from this checkout, unless the workflow already built olake-ui:ci
		if os.Getenv("OLAKE_TEST_SKIP_UI_BUILD") == "true" {
			if err := r.run(ctx, "docker image inspect --format '{{.Id}}' olake-ui:ci", "OLAKE_TEST_SKIP_UI_BUILD is set but olake-ui:ci is missing"); err != nil {
				return err
			}
		} else if err := r.run(ctx, buildOLakeUICmd(root), "olake-ui image build failed"); err != nil {
			return err
		}
		return r.run(ctx, startOLakeCmd(root), "OLake startup failed")
	})
	playwright := r.task(startCtx, cancelStart, &tasks, "Playwright", func(ctx context.Context) error {
		return env.CheckPlaywright(ctx, t)
	})

	// Seed as soon as the source is ready. Seed fails the test through t, which only the test's
	// own goroutine may do, so it runs here.
	var failures []string
	if err := <-source; err != nil {
		failures = append(failures, err.Error())
	} else {
		r.Step("Seeding the %s source...", w.Source.Name)
		w.Source.Seed(ctx, t, r.sourceHost, r.sourcePort)
	}
	for _, done := range []<-chan error{destination, olake, playwright} {
		if err := <-done; err != nil {
			failures = append(failures, err.Error())
		}
	}
	if len(failures) > 0 {
		t.Fatalf("startup failed:\n%s", strings.Join(failures, "\n"))
	}
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

// exec runs cmd and fails the test with its output when it does not exit 0.
func (r *Run) exec(cmd, format string, args ...any) []byte {
	r.t.Helper()
	code, out, err := r.env.Exec(r.ctx, cmd)
	if err != nil || code != 0 {
		r.t.Fatalf("%s (%d): %v\n%s", fmt.Sprintf(format, args...), code, err, out)
	}
	return out
}

// run is exec for a task goroutine: it returns the failure instead of failing the test, which
// only the test's own goroutine may do.
func (r *Run) run(ctx context.Context, cmd, format string, args ...any) error {
	code, out, err := r.env.Exec(ctx, cmd)
	if err != nil || code != 0 {
		return fmt.Errorf("%s (%d): %v\n%s", fmt.Sprintf(format, args...), code, err, out)
	}
	return nil
}

// task runs fn on its own goroutine, counted in wg, and logs when it starts and ends. The returned
// channel gets fn's error; an error also calls cancel, which stops the other tasks of the group.
func (r *Run) task(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup, name string, fn func(context.Context) error) <-chan error {
	done := make(chan error, 1)
	r.Step("Starting %s...", name)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := fn(ctx)
		if err != nil {
			cancel()
			r.Step("%s failed", name)
		} else {
			r.Step("%s is ready", name)
		}
		done <- err
	}()
	return done
}

// Playwright runs the tests of spec tagged tag. Every run gets E2E_JOB_NAME, E2E_DATABASE (the
// destination database the checks read, so the spec does not derive it again) and
// E2E_SERVICE_HOST; env adds more E2E_* variables.
func (r *Run) Playwright(spec, tag string, env map[string]string) {
	r.t.Helper()
	vars := map[string]string{
		"E2E_JOB_NAME":     r.w.Job,
		"E2E_DATABASE":     r.w.database(),
		"E2E_SERVICE_HOST": r.ServiceHost,
	}
	for k, v := range env {
		vars[k] = v
	}
	// OLAKE_TEST_PW_DEBUG=true logs every browser action; off, the log holds the test steps only.
	cmd, err := playwrightCmd(r.env.Root(), spec, tag, vars, os.Getenv("OLAKE_TEST_PW_DEBUG") == "true")
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
		// One probe first: when Spark Connect is down, fail now rather than retry every check.
		if err := waitForSpark(r.ctx, spark); err != nil {
			t.Fatalf("Spark Connect at sc://%s:%s does not answer: %v", r.sparkHost, r.sparkPort, err)
		}
		table := func(name string) string { return r.w.Destination.Table(r.w.database(), name) }
		runChecks(r.ctx, t, spark, table, checks)
	})
	if !ok {
		r.t.FailNow()
	}
}

// MatchesSource checks that the destination's current rows of table equal the source's rows,
// column by column. The current row of an id is its newest row by _olake_timestamp, and an id
// whose newest row is a delete tombstone (_op_type 'd') has none: that reads Append tables, which
// keep every change, and Upsert tables the same way. where, a code constant or empty, limits the
// source rows the way a Data Filter or CDC Only limits what reaches the destination.
func (r *Run) MatchesSource(table string, columns []string, where string) Check {
	if len(columns) == 0 || columns[0] != keyColumn {
		panic(fmt.Sprintf("MatchesSource(%s): the first column must be %s", table, keyColumn))
	}
	srcCols := make([]string, len(columns))
	dstCols := make([]string, len(columns))
	for i, c := range columns {
		srcCols[i] = fmt.Sprintf("CAST(%s AS TEXT)", ident(c))
		dstCols[i] = fmt.Sprintf("CAST(%s AS STRING)", ident(c))
	}
	srcQuery := fmt.Sprintf("SELECT %s FROM %s", strings.Join(srcCols, ", "), ident(table))
	name := fmt.Sprintf("%s matches the source (%s)", table, strings.Join(columns, ", "))
	if where != "" {
		srcQuery += " WHERE " + where
		name += " where " + where
	}
	srcQuery += " ORDER BY " + keyColumn
	return Check{
		Name: name,
		Run: func(ctx context.Context, spark sql.SparkSession, tableName func(string) string) error {
			source, err := r.w.Source.Query(ctx, r.sourceHost, r.sourcePort, srcQuery)
			if err != nil {
				return fmt.Errorf("reading the source: %w", err)
			}
			tbl := tableName(ident(table))
			destination, err := refreshed(ctx, spark, tbl, fmt.Sprintf(
				"SELECT %s FROM (SELECT *, ROW_NUMBER() OVER (PARTITION BY %s ORDER BY _olake_timestamp DESC) AS _current FROM %s) "+
					"WHERE _current = 1 AND _op_type != 'd' ORDER BY %s",
				strings.Join(dstCols, ", "), keyColumn, tbl, keyColumn))
			if err != nil {
				return fmt.Errorf("reading the destination: %w", err)
			}
			return diffRows(columns, source, destination)
		},
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
