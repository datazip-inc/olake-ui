package tests

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/apache/spark-connect-go/v35/spark/sql"
	"github.com/docker/docker/api/types/container"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// Install required tools during initial setup
	setupToolsCmd = `
        apt-get update && 
        apt-get install -y curl postgresql-client dnsutils iputils-ping ncurses-bin ca-certificates gnupg lsb-release netcat-openbsd && 
        curl -fsSL https://deb.nodesource.com/setup_20.x | bash - &&
        apt-get install -y nodejs &&
        npm install -g pnpm &&
        curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg &&
        echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null &&
        apt-get update &&
        apt-get install -y docker-compose-plugin &&
        update-ca-certificates
    `

	// Download destination docker-compose
	// TODO: Either move destination and source config into the same Docker Compose setup or download both from the same location for consistency.
	downloadDestinationComposeCmd = `
        cd /mnt &&
        curl -fsSL -o docker-compose.destination.yml \
            https://raw.githubusercontent.com/datazip-inc/olake/master/destination/iceberg/local-test/docker-compose.yml
    `

	// Start postgres test infrastructure
	startPostgresCmd = `
        cd /mnt/server/tests &&
        docker compose up -d &&
        for i in $(seq 1 30); do
            if docker exec olake_postgres-test psql -h localhost -U postgres -d postgres -c "SELECT 1" 2>/dev/null; then
                echo "PostgreSQL ready."
                break
            fi
            sleep 2
        done &&
        docker exec olake_postgres-test psql -U postgres -d postgres -c \
            "SELECT slot_name, plugin, slot_type, active FROM pg_replication_slots WHERE slot_name = 'olake_slot';"
    `

	// Start destination services (iceberg stack)
	startDestinationCmd = `
        cd /mnt &&
        docker compose -f docker-compose.destination.yml up -d minio mc postgres spark-iceberg &&
        sleep 5 &&
        docker compose -f docker-compose.destination.yml ps
    `

	// Start OLake application
	startOLakeCmd = `
        cd /mnt && 
        mkdir -p /mnt/olake-data && 
        docker compose up -d && 
        for i in $(seq 1 60); do
            if curl -f http://localhost:8000/health 2>/dev/null || curl -f http://localhost:8000 2>/dev/null; then
                echo "OLake UI ready."
                break
            fi
            sleep 2
        done
    `

	// Network setup
	networkSetupCmd = `
        docker network create olake-network || true &&
        docker network connect olake-network olake-ui || true &&
        docker network connect olake-network postgres || true &&
        docker network connect olake-network olake_postgres-test || true
    `

	// Install Playwright and dependencies
	installPlaywrightCmd = `
        cd /mnt/ui &&
        pnpm add -D @playwright/test &&
        pnpm exec playwright install --with-deps chromium
    `

	// Run Playwright tests
	runPlaywrightCmd = `
        cd /mnt/ui &&
        PLAYWRIGHT_TEST_BASE_URL=http://localhost:8000 DEBUG=pw:api npx playwright test tests/flows/job-end-to-end.spec.ts
    `

	icebergDB        = "postgres_iceberg_jdbc_job_postgres_public"
	icebergCatalog   = "olake_iceberg"
	currentTestTable = "postgres_test_table_olake"
)

func DinDTestContainer(t *testing.T) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	projectRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		return fmt.Errorf("could not determine project root: %w", err)
	}
	t.Logf("Project root identified at: %s", projectRoot)

	req := testcontainers.ContainerRequest{
		Image: "ubuntu:22.04",
		Env: map[string]string{
			"DOCKER_TLS_CERTDIR":           "",
			"TELEMETRY_DISABLED":           "true",
			"TESTCONTAINERS_RYUK_DISABLED": "true",
			"DEBIAN_FRONTEND":              "noninteractive",
		},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Privileged = true
			hc.Binds = []string{
				fmt.Sprintf("%s:/mnt:rw", projectRoot),
			}
			// Tmpfs mounts create temporary in-memory filesystems inside the container.
			// These directories behave like RAM disks they exist only in memory (not on disk) and are automatically cleaned up when the container stops.
			// This is useful for high-performance temporary storage.
			// 70GB for docker and 15GB for shared memory space in Linux
			hc.Tmpfs = map[string]string{
				"/var/lib/docker": "size=70G",
				"/dev/shm":        "size=10G",
			}
			hc.Resources.Memory = 22 * 1024 * 1024 * 1024 // 22GB
			hc.ExtraHosts = append(hc.ExtraHosts, "host.docker.internal:host-gateway")
		},
		ConfigModifier: func(config *container.Config) {
			config.WorkingDir = "/mnt"
		},
		ExposedPorts: []string{"8000/tcp", "2375/tcp", "5433/tcp", "15002/tcp"},
		Cmd: []string{
			"/bin/sh", "-c",
			`set -e
			apt-get update -y &&
			apt-get install -y --no-install-recommends ca-certificates curl gnupg lsb-release iproute2 procps &&
			apt-get install -y --no-install-recommends docker.io &&
			mkdir -p /var/lib/docker /var/run/docker &&
			exec dockerd --host=tcp://0.0.0.0:2375 --host=unix:///var/run/docker.sock
			`,
		},
		WaitingFor: wait.ForExec([]string{"docker", "-H", "tcp://127.0.0.1:2375", "info"}).WithStartupTimeout(60 * time.Second).WithPollInterval(1 * time.Second),
	}

	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return fmt.Errorf("failed to start DinD container: %w", err)
	}

	// container host
	host, err := ctr.Host(ctx)
	if err != nil {
		return fmt.Errorf("failed to get host: %w", err)
	}

	// postgres port
	postgresPort, err := ctr.MappedPort(ctx, "5433/tcp")
	if err != nil {
		return fmt.Errorf("failed to get postgres port: %w", err)
	}

	// iceberg-spark port
	sparkPort, err := ctr.MappedPort(ctx, "15002/tcp")
	if err != nil {
		return fmt.Errorf("failed to get spark port: %w", err)
	}

	t.Log("Docker daemon is ready")
	// Step 1: Install tools
	t.Log("Installing required tools...")
	if code, out, err := ExecCommand(ctx, ctr, setupToolsCmd); err != nil || code != 0 {
		return fmt.Errorf("tools installation failed (%d): %s\n%s", code, err, out)
	}

	// Step 2: Download destination docker-compose
	t.Log("Downloading destination docker-compose...")
	if code, out, err := ExecCommand(ctx, ctr, downloadDestinationComposeCmd); err != nil || code != 0 {
		return fmt.Errorf("destination docker-compose download failed (%d): %s\n%s", code, err, out)
	}

	// Step 3: Start PostgreSQL test infrastructure
	t.Log("Starting PostgreSQL test infrastructure...")
	if code, out, err := ExecCommand(ctx, ctr, startPostgresCmd); err != nil || code != 0 {
		return fmt.Errorf("postgres startup failed (%d): %s\n%s", code, err, out)
	}

	// Step 4: Start destination services (Iceberg stack)
	t.Log("Starting destination services...")
	if code, out, err := ExecCommand(ctx, ctr, startDestinationCmd); err != nil || code != 0 {
		return fmt.Errorf("destination services startup failed (%d): %s\n%s", code, err, out)
	}

	// Step 5: Patch docker-compose for local images
	t.Log("Patching docker-compose to build local images...")
	if err := PatchDockerCompose(ctx, t, ctr); err != nil {
		return err
	}

	// Step 6: Start OLake application
	t.Log("Starting OLake docker-compose services...")
	if code, out, err := ExecCommand(ctx, ctr, startOLakeCmd); err != nil || code != 0 {
		return fmt.Errorf("OLake startup failed (%d): %s\n%s", code, err, out)
	}

	// Step 7: Setup networks
	t.Log("Setting up Docker networks...")
	if code, out, err := ExecCommand(ctx, ctr, networkSetupCmd); err != nil || code != 0 {
		t.Logf("Warning: Network setup failed (%d): %s\n%s", code, err, out)
	}

	// Step 8: Query the postgres source
	ExecuteQuery(ctx, t, "create", host, postgresPort.Port())
	ExecuteQuery(ctx, t, "clean", host, postgresPort.Port())
	ExecuteQuery(ctx, t, "add", host, postgresPort.Port())

	t.Logf("OLake UI is ready and accessible at: http://localhost:8000")

	// Step 9: Install Playwright
	t.Log("Installing Playwright and dependencies...")
	if code, out, err := ExecCommand(ctx, ctr, installPlaywrightCmd); err != nil || code != 0 {
		return fmt.Errorf("playwright installation failed (%d): %s\n%s", code, err, out)
	}

	// Step 10: Run Playwright tests
	t.Log("Executing Playwright tests...")
	if code, out, err := ExecCommandWithStreaming(ctx, t, ctr, runPlaywrightCmd); err != nil || code != 0 {
		return fmt.Errorf("playwright tests failed (%d): %s\n%s", code, err, out)
	}
	t.Log("Playwright tests passed successfully.")

	// wait before verifying iceberg data
	t.Log("Waiting for 60 seconds before verifying iceberg data...")
	time.Sleep(60 * time.Second)

	// Step 11: Verify in iceberg
	t.Logf("Starting Iceberg data verification...")
	VerifyIcebergTest(ctx, t, ctr, host, sparkPort.Port())
	return nil
}

// ExecCommandWithStreaming executes a command and streams output in real-time
func ExecCommandWithStreaming(ctx context.Context, t *testing.T, ctr testcontainers.Container, cmd string) (int, string, error) {
	exitCode, reader, err := ctr.Exec(ctx, []string{"sh", "-c", cmd})
	if err != nil {
		return -1, "", err
	}

	var output strings.Builder
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		t.Log(line)
		output.WriteString(line + "\n")
	}

	if err := scanner.Err(); err != nil {
		return exitCode, output.String(), err
	}

	return exitCode, output.String(), nil
}

// PatchDockerCompose updates olake-ui to build from local code
// TODO: Remove patch command and find alternative to use local code
func PatchDockerCompose(ctx context.Context, t *testing.T, ctr testcontainers.Container) error {
	patchCmd := `
    set -e
    tmpfile=$(mktemp)
    awk '
    BEGIN{svc="";}
    /^  olake-ui:/{svc="olake-ui"; print; next}
    /^  temporal-worker:/{svc="temporal-worker"; print; next}
    /^  [A-Za-z0-9_-]+:/{ if (svc!="") svc=""; print; next}
    {
      if (svc=="olake-ui" && $0 ~ /^    image:/) {
        print "    build:";
        print "      context: .";
        print "      dockerfile: Dockerfile";
        next
      }
      print
    }
    ' /mnt/docker-compose.yml > "$tmpfile" && mv "$tmpfile" /mnt/docker-compose.yml
`

	code, out, err := ExecCommand(ctx, ctr, patchCmd)
	if err != nil || code != 0 {
		t.Logf("docker-compose patch output: %s", string(out))
		return fmt.Errorf("failed to patch docker-compose.yml (%d): %s\n%s", code, err, out)
	}
	t.Log("docker-compose.yml patched to build local images")
	t.Logf("Patched docker-compose.yml:\n%s", string(out))

	return nil
}

func VerifyIcebergTest(ctx context.Context, t *testing.T, ctr testcontainers.Container, host, port string) {
	sparkConnectAddress := fmt.Sprintf("sc://%s:%s", host, port)
	spark, err := sql.NewSessionBuilder().Remote(sparkConnectAddress).Build(ctx)
	require.NoError(t, err, "Failed to connect to Spark Connect server")
	defer func() {
		if stopErr := spark.Stop(); stopErr != nil {
			t.Errorf("Failed to stop Spark session: %v", stopErr)
		}
		if ctr != nil {
			t.Log("Running cleanup...")
			// Stop docker-compose services
			_, _, _ = ExecCommand(ctx, ctr, "cd /mnt && docker-compose down -v --remove-orphans")
			// Terminate the DinD container
			if err := ctr.Terminate(ctx); err != nil {
				t.Logf("Warning: failed to terminate container: %v", err)
			}
			t.Log("Cleanup complete")
		}
	}()
	countQuery := fmt.Sprintf(
		"SELECT COUNT(DISTINCT _olake_id) as unique_count FROM %s.%s.%s",
		icebergCatalog, icebergDB, currentTestTable,
	)
	t.Logf("Executing query: %s", countQuery)

	countQueryDf, err := spark.Sql(ctx, countQuery)
	require.NoError(t, err, "Failed to execute query on the table")

	rows, err := countQueryDf.Collect(ctx)
	require.NoError(t, err, "Failed to collect data rows from Iceberg")
	require.NotEmpty(t, rows, "No rows returned for _op_type = 'r'")

	// check count and verify
	countValue := rows[0].Value("unique_count").(int64)
	require.Equal(t, int64(5), countValue, "Expected count to be 5")
	t.Logf("✅ Test passed: count value %v matches expected value 5", countValue)
}

func ExecuteQuery(ctx context.Context, t *testing.T, operation, host, port string) {
	t.Helper()
	connStr := fmt.Sprintf("postgres://postgres@%s:%s/postgres?sslmode=disable", host, port)
	db, ok := sqlx.ConnectContext(ctx, "postgres", connStr)
	require.NoError(t, ok, "failed to connect to postgres")
	defer func() {
		require.NoError(t, db.Close(), "failed to close postgres connection")
	}()

	// integration test uses only one stream for testing
	integrationTestTable := currentTestTable
	var query string

	switch operation {
	case "create":
		query = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				col_bigint BIGINT,
				col_bigserial BIGSERIAL PRIMARY KEY,
				col_bool BOOLEAN,
				col_char CHAR(1),
				col_character CHAR(10),
				col_character_varying VARCHAR(50),
				col_date DATE,
				col_decimal NUMERIC,
				col_double_precision DOUBLE PRECISION,
				col_float4 REAL,
				col_int INT,
				col_int2 SMALLINT,
				col_integer INTEGER,
				col_interval INTERVAL,
				col_json JSON,
				col_jsonb JSONB,
				col_name NAME,
				col_numeric NUMERIC,
				col_real REAL,
				col_text TEXT,
				col_timestamp TIMESTAMP,
				col_timestamptz TIMESTAMPTZ,
				col_uuid UUID,
				col_varbit VARBIT(20),
				col_xml XML,
				CONSTRAINT unique_custom_key UNIQUE (col_bigserial)
			)`, integrationTestTable)

	case "drop":
		query = fmt.Sprintf("DROP TABLE IF EXISTS %s", integrationTestTable)

	case "clean":
		query = fmt.Sprintf("TRUNCATE TABLE %s", integrationTestTable)

	case "add":
		insertTestData(ctx, t, db, integrationTestTable)
		return // Early return since we handle all inserts in the helper function

	case "insert":
		query = fmt.Sprintf(`
			INSERT INTO %s (
				col_bigint, col_bool, col_char, col_character,
				col_character_varying, col_date, col_decimal,
				col_double_precision, col_float4, col_int, col_int2,
				col_integer, col_interval, col_json, col_jsonb,
				col_name, col_numeric, col_real, col_text,
				col_timestamp, col_timestamptz, col_uuid, col_varbit, col_xml
			) VALUES (
				123456789012345, TRUE, 'c', 'char_val',
				'varchar_val', '2023-01-01', 123.45,
				123.456789, 123.45, 123, 123, 12345,
				'1 hour', '{"key": "value"}', '{"key": "value"}',
				'test_name', 123.45, 123.45, 'sample text',
				'2023-01-01 12:00:00', '2023-01-01 12:00:00+00',
				'123e4567-e89b-12d3-a456-426614174000', B'101010',
				'<tag>value</tag>'
			)`, integrationTestTable)

	case "update":
		query = fmt.Sprintf(`
			UPDATE %s SET
				col_bigint = 123456789012340,
				col_bool = FALSE,
				col_char = 'd',
				col_character = 'updated__',
				col_character_varying = 'updated val',
				col_date = '2024-07-01',
				col_decimal = 543.21,
				col_double_precision = 987.654321,
				col_float4 = 543.21,
				col_int = 321,
				col_int2 = 321,
				col_integer = 54321,
				col_interval = '2 hours',
				col_json = '{"new": "json"}',
				col_jsonb = '{"new": "jsonb"}',
				col_name = 'updated_name',
				col_numeric = 321.00,
				col_real = 321.00,
				col_text = 'updated text',
				col_timestamp = '2024-07-01 15:30:00',
				col_timestamptz = '2024-07-01 15:30:00+00',
				col_uuid = '00000000-0000-0000-0000-000000000000',
				col_varbit = B'111000',
				col_xml = '<updated>value</updated>'
			WHERE col_bigserial = 1`, integrationTestTable)

	case "delete":
		query = fmt.Sprintf("DELETE FROM %s WHERE col_bigserial = 1", integrationTestTable)

	default:
		t.Fatalf("Unsupported operation: %s", operation)
	}
	_, err := db.ExecContext(ctx, query)
	require.NoError(t, err, "Failed to execute %s operation", operation)
}

// insertTestData inserts test data into the specified table
func insertTestData(ctx context.Context, t *testing.T, db *sqlx.DB, tableName string) {
	t.Helper()

	for i := 1; i <= 5; i++ {
		query := fmt.Sprintf(`
		INSERT INTO %s (
			col_bigint, col_bigserial, col_bool, col_char, col_character,
			col_character_varying, col_date, col_decimal,
			col_double_precision, col_float4, col_int, col_int2, col_integer,
			col_interval, col_json, col_jsonb, col_name, col_numeric,
			col_real, col_text, col_timestamp, col_timestamptz,
			col_uuid, col_varbit, col_xml
		) VALUES (
			123456789012345, DEFAULT, TRUE, 'c', 'char_val',
			'varchar_val', '2023-01-01', 123.45,
			123.456789, 123.45, 123, 123, 12345, '1 hour', '{"key": "value"}',
			'{"key": "value"}', 'test_name', 123.45, 123.45,
			'sample text', '2023-01-01 12:00:00',
			'2023-01-01 12:00:00+00',
			'123e4567-e89b-12d3-a456-426614174000', B'101010',
			'<tag>value</tag>'
		)`, tableName)

		_, err := db.ExecContext(ctx, query)
		require.NoError(t, err, "Failed to insert test data")
	}
}

// Helper function to execute container commands
func ExecCommand(
	ctx context.Context,
	c testcontainers.Container,
	cmd string,
) (int, []byte, error) {
	code, reader, err := c.Exec(ctx, []string{"/bin/sh", "-c", cmd})
	if err != nil {
		return code, nil, err
	}
	output, _ := io.ReadAll(reader)
	return code, output, nil
}
