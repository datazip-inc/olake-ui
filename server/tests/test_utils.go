package tests

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var dindContainer testcontainers.Container

const (
	StartupComposeCmd = `
		apk add --no-cache docker-compose curl && 
		echo "Docker-compose installed." && 
		cd /mnt && 
		echo "Starting docker-compose..." && 
		docker-compose up -d && 
		echo "Services started. Waiting for containers..." && 
		sleep 5 && 
		docker-compose ps
	`
	icebergDB           = "postgres_postgres_public"
	icebergCatalog      = "olake_iceberg"
	currentTestTable    = "postgres_test_table_olake"
	sparkConnectAddress = "sc://localhost:15002"
)

func DinDTestContainer(t *testing.T) error {
	ctx := context.Background()
	projectRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		return fmt.Errorf("could not determine project root: %w", err)
	}
	t.Logf("Project root identified at: %s", projectRoot)

	// query the postgres source
	ExecuteQuery(context.Background(), t, "create")
	ExecuteQuery(context.Background(), t, "clean")
	ExecuteQuery(context.Background(), t, "add")

	req := testcontainers.ContainerRequest{
		Image:        "docker:25.0-dind",
		ExposedPorts: []string{"8000:8000/tcp", "2375/tcp"},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Privileged = true
			hc.Binds = []string{
				fmt.Sprintf("%s:/mnt:rw", projectRoot),
			}
			hc.ExtraHosts = append(hc.ExtraHosts, "host.docker.internal:host-gateway")
		},
		Cmd: []string{"dockerd", "--host=unix:///var/run/docker.sock", "--host=tcp://0.0.0.0:2375"},
		ConfigModifier: func(config *container.Config) {
			config.WorkingDir = "/mnt"
		},
		Env: map[string]string{
			"TELEMETRY_DISABLED": "true",
			"DOCKER_TLS_CERTDIR": "", // No need for TLS in tests
		},
		WaitingFor: wait.ForLog("API listen on").WithStartupTimeout(60 * time.Second),
	}

	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return fmt.Errorf("failed to start DinD container: %w", err)
	}
	// persist container for later cleanup
	dindContainer = ctr

	t.Log("Waiting for Docker daemon to be ready...")
	time.Sleep(3 * time.Second)

	// Verify Docker is working
	if code, out, err := ExecCommand(ctx, ctr, "docker info"); err != nil || code != 0 {
		return fmt.Errorf("docker daemon not ready (%d): %s\n%s", code, err, out)
	}
	t.Log("Docker daemon is ready")

	// Start docker-compose
	t.Log("Starting docker-compose services...")
	if code, out, err := ExecCommand(ctx, ctr, StartupComposeCmd); err != nil || code != 0 {
		return fmt.Errorf("docker compose up failed (%d): %s\n%s", code, err, out)
	}
	t.Logf("OLake UI is ready and accessible at: http://localhost:8000")
	return nil
}

func ExecuteQuery(ctx context.Context, t *testing.T, operation string) {
	t.Helper()
	connStr := "postgres://postgres@localhost:5433/postgres?sslmode=disable"
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
		query = fmt.Sprintf("DELETE FROM %s", integrationTestTable)

	case "add":
		insertTestData(t, ctx, db, integrationTestTable)
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
func insertTestData(t *testing.T, ctx context.Context, db *sqlx.DB, tableName string) {
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
