package telemetry

import (
	"context"

	"github.com/beego/beego/v2/core/logs"
	"github.com/datazip/olake-frontend/server/internal/constants"
)

// TrackUserLogin tracks when a user logs in to  olake-ui
func TrackUserLogin(ctx context.Context, userID int, email, username string) error {
	if instance != nil {
		instance.username = username
	}

	properties := map[string]interface{}{
		"user_id": userID,
		"email":   email,
	}

	if err := TrackEvent(ctx, constants.EventUserLogin, properties); err != nil {
		logs.Error("Failed to track user login event: %v", err)
		return err
	}

	return nil
}
