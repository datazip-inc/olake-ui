package telemetry

import (
	"context"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

// startupEventTimeout bounds how long a startup event waits for InitTelemetry.
const startupEventTimeout = 5 * time.Second

// trackStartupEvent fires the event once InitTelemetry finishes, dropping it
// if telemetry never became available.
func trackStartupEvent(event string, properties map[string]interface{}) {
	go func() {
		select {
		case <-ready:
		case <-time.After(startupEventTimeout):
			return
		}
		if instance == nil {
			return
		}
		if err := TrackEvent(context.Background(), event, properties); err != nil {
			logger.Debugf("Failed to track %s event: %s", event, err)
		}
	}()
}


func TrackInstalledFusion(KubernetesServiceHost string) {
	if KubernetesServiceHost != "" {
		mode = "helm"
	} else {
		mode = "docker"
	}
	trackStartupEvent(EventInstalledFusion, map[string]interface{}{"mode": mode})
}

func TrackInstallationMode(KubernetesServiceHost string) {
	if KubernetesServiceHost != "" {
		mode = "helm"
	} else {
		mode = "docker"
	}
	trackStartupEvent(EventInstallationMode, map[string]interface{}{"mode": mode})
}