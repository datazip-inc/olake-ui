package tests

import (
	"testing"

	_ "github.com/lib/pq"
)

func TestDinDIntegration(t *testing.T) {
	err := DinDTestContainer(t)
	if err != nil {
		t.Errorf("Error in Docker in Docker container start up: %s", err)
	}
}
