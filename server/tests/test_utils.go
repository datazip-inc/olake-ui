package tests

import (
	"context"
	"fmt"
	"os"
	"regexp"
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

// Run a Playwright spec, relative to ui/
func runPlaywrightCmd(root, job, serviceHost, spec string) string {
	return fmt.Sprintf(`
        cd %s &&
        E2E_JOB_NAME=%s E2E_SERVICE_HOST=%s PLAYWRIGHT_TEST_BASE_URL=http://localhost:8000 DEBUG=pw:api npx playwright test %s
    `, shellQuote(root+"/ui"), shellQuote(job), shellQuote(serviceHost), shellQuote(spec))
}

// jobPattern and specPattern limit the values RunWorkflow puts in the Playwright shell command.
var (
	jobPattern  = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	specPattern = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_./-]*\.spec\.ts$`) // no leading '-': Playwright would read an option
)

// sparkConnectPort is the port VerifyDestination reaches Spark Connect on. Every destination's
// compose file must run Spark Connect there; see destinations.Destination.
const sparkConnectPort = "15002/tcp"

// Workflow is one integration run: the source and destination to start, the Playwright spec
// that creates and syncs the job through the UI, and the checks run on the destination afterwards.
type Workflow struct {
	Source      sources.Source
	Destination destinations.Destination
	Spec        string  // Playwright spec relative to ui/
	Job         string  // name of the job the spec creates; passed to it as E2E_JOB_NAME
	Checks      []Check // run on the destination after the spec has passed
}

// database is the destination database OLake writes the job to: <job>_<source namespace>, lowercased.
func (w Workflow) database() string {
	return strings.ToLower(w.Job + "_" + w.Source.Namespace)
}

// Check verifies the destination through Spark. table turns a source table name into the
// fully qualified name Spark reads it from.
type Check func(ctx context.Context, t *testing.T, spark sql.SparkSession, table func(name string) string)

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

// RunWorkflow starts the Env, the source, the destination and OLake, seeds the source, runs the
// Playwright spec, and runs the workflow's checks on the destination.
func RunWorkflow(t *testing.T, w Workflow) error {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Minute) // under both CI jobs' timeouts (60 and 75 minutes), with room for setup and cleanup
	defer cancel()
	if !jobPattern.MatchString(w.Job) {
		return fmt.Errorf("invalid job name %q", w.Job)
	}
	if !specPattern.MatchString(w.Spec) {
		return fmt.Errorf("invalid spec path %q", w.Spec)
	}

	ports := []string{sparkConnectPort}
	if w.Source.Port != "" {
		ports = append(ports, w.Source.Port)
	}
	env, err := newEnv(ctx, t, ports)
	if err != nil {
		return err
	}
	root := env.Root()

	// The address the driver uses for the source and destination; checked before anything starts.
	serviceHost, err := env.ServiceHost(ctx)
	if err != nil {
		return err
	}
	t.Logf("Service host: %s", serviceHost)

	// Spark Connect address
	sparkHost, sparkPort, err := env.Addr(ctx, sparkConnectPort)
	if err != nil {
		return fmt.Errorf("failed to get spark port: %w", err)
	}

	// source address, if the source publishes a port
	sourceHost, sourcePort := sparkHost, ""
	if w.Source.Port != "" {
		if sourceHost, sourcePort, err = env.Addr(ctx, w.Source.Port); err != nil {
			return fmt.Errorf("failed to get %s port: %w", w.Source.Name, err)
		}
	}

	// Step 1: Start the source
	t.Logf("Starting %s source...", w.Source.Name)
	if code, out, err := env.Exec(ctx, composeStartCmd(root, "source", w.Source.ComposeDir, nil, w.Source.ReadyCmd)); err != nil || code != 0 {
		return fmt.Errorf("%s startup failed (%d): %s\n%s", w.Source.Name, code, err, out)
	}

	// Step 2: Start the destination
	t.Logf("Starting %s destination...", w.Destination.Name)
	if code, out, err := env.Exec(ctx, composeStartCmd(root, "destination", w.Destination.ComposeDir, w.Destination.Services, w.Destination.ReadyCmd)); err != nil || code != 0 {
		return fmt.Errorf("destination services startup failed (%d): %s\n%s", code, err, out)
	}

	// Step 3: Build olake-ui from this checkout, unless the workflow already built olake-ui:ci
	if os.Getenv("OLAKE_TEST_SKIP_UI_BUILD") == "true" {
		t.Log("Using the prebuilt olake-ui image...")
		if code, out, err := env.Exec(ctx, "docker image inspect --format '{{.Id}}' olake-ui:ci"); err != nil || code != 0 {
			return fmt.Errorf("OLAKE_TEST_SKIP_UI_BUILD is set but olake-ui:ci is missing (%d): %s\n%s", code, err, out)
		}
	} else {
		t.Log("Building the olake-ui image...")
		if code, out, err := env.Exec(ctx, buildOLakeUICmd(root)); err != nil || code != 0 {
			return fmt.Errorf("olake-ui image build failed (%d): %s\n%s", code, err, out)
		}
	}

	// Step 4: Start OLake application
	t.Log("Starting OLake docker-compose services...")
	if code, out, err := env.Exec(ctx, startOLakeCmd(root)); err != nil || code != 0 {
		return fmt.Errorf("OLake startup failed (%d): %s\n%s", code, err, out)
	}

	// Step 5: Seed the source
	w.Source.Seed(ctx, t, sourceHost, sourcePort)

	t.Logf("OLake UI is ready and accessible at: http://localhost:8000")

	// Step 6: Install Playwright
	t.Log("Installing Playwright and dependencies...")
	if err := env.InstallPlaywright(ctx, t); err != nil {
		return err
	}

	// Step 7: Run Playwright tests
	t.Log("Executing Playwright tests...")
	if code, out, err := env.ExecStream(ctx, t, runPlaywrightCmd(root, w.Job, serviceHost, w.Spec)); err != nil || code != 0 {
		return fmt.Errorf("playwright tests failed (%d): %s\n%s", code, err, out)
	}
	t.Log("Playwright tests passed successfully.")

	// wait before verifying destination data
	t.Log("Waiting for 60 seconds before verifying destination data...")
	time.Sleep(60 * time.Second)

	// Step 8: Verify the destination
	t.Logf("Starting %s data verification...", w.Destination.Name)
	VerifyDestination(ctx, t, sparkHost, sparkPort, w)
	return nil
}

// VerifyDestination connects to Spark and runs every check of the workflow.
func VerifyDestination(ctx context.Context, t *testing.T, host, port string, w Workflow) {
	sparkConnectAddress := fmt.Sprintf("sc://%s:%s", host, port)
	spark, err := sql.NewSessionBuilder().Remote(sparkConnectAddress).Build(ctx)
	require.NoError(t, err, "Failed to connect to Spark Connect server")
	defer func() {
		if stopErr := spark.Stop(); stopErr != nil {
			t.Errorf("Failed to stop Spark session: %v", stopErr)
		}
	}()
	table := func(name string) string { return w.Destination.Table(w.database(), name) }
	for _, check := range w.Checks {
		check(ctx, t, spark, table)
	}
}
