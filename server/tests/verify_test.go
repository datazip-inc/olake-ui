package tests

import (
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func TestDinDIntegration(t *testing.T) {
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	err := DinDTestContainer(t)
	if err != nil {
		t.Errorf("Error in Docker in Docker container start up: %s", err)
	}
}
