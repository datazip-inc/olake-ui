package telemetry

import (
	"context"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

// waitForTelemetryInstance waits briefly for InitTelemetry's goroutine to set instance.
func waitForTelemetryInstance() bool {
	for i := 0; i < 50; i++ {
		if instance != nil {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return instance != nil
}

// TrackInstalledFusion fires when optimization/Fusion is available at startup.
func TrackInstalledFusion() {
	go func() {
		if !waitForTelemetryInstance() {
			return
		}

		if err := TrackEvent(context.Background(), EventInstalledFusion, nil); err != nil {
			logger.Debug("Failed to track installed fusion event: %s", err)
		}
	}()
}

// TrackInstallationMode reports whether the app is running via helm (k8s) or docker.
func TrackInstallationMode(mode string) {
	go func() {
		if !waitForTelemetryInstance() {
			return
		}

		properties := map[string]interface{}{
			"mode": mode,
		}
		if err := TrackEvent(context.Background(), EventInstallationMode, properties); err != nil {
			logger.Debug("Failed to track installation mode event: %s", err)
		}
	}()
}
