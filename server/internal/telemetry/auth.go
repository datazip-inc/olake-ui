package telemetry

import (
	"context"

	"github.com/beego/beego/v2/core/logs"
	"github.com/datazip/olake-ui/server/internal/models"
)

// TrackUserLogin tracks when a user logs in to  olake-ui
func TrackUserLogin(ctx context.Context, user *models.User) {
	go func() {
		if instance == nil || user == nil {
			return
		}

		instance.username = user.Username
		properties := map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
		}

		err := TrackEvent(ctx, EventUserLogin, properties)
		if err != nil {
			logs.Debug("Failed to track user login event: %s", err)
		}
	}()
}
