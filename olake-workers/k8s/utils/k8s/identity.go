package k8s

import (
	"fmt"
	"os"
)

// GenerateWorkerIdentity creates a unique worker identity based on hostname
func GenerateWorkerIdentity() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return fmt.Sprintf("olake-ui/olake-workers/k8s-%s", hostname)
}