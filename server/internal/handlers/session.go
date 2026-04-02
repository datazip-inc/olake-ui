package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
)

type sessionPayload struct {
	UserID int `json:"user_id"`
}

type sessionStore struct {
	db      *database.Database
	enabled bool
	secure  bool
}

const (
	sessionCookieName = "olake-session"
	sessionMaxAgeDays = 30
)

func newSessionStore(cfg *appconfig.Config, db *database.Database) *sessionStore {
	return &sessionStore{
		db:      db,
		enabled: cfg.SessionOn,
		secure:  cfg.RunMode != "localdev",
	}
}

func (s *sessionStore) SetUserSession(c *gin.Context, userID int) error {
	if !s.enabled {
		return nil
	}

	sessionID := utils.ULID()
	expiresAt := time.Now().Add(sessionMaxAgeDays * 24 * time.Hour)

	payload, err := json.Marshal(sessionPayload{UserID: userID})
	if err != nil {
		return fmt.Errorf("failed to marshal session payload: %w", err)
	}

	err = s.db.UpsertSession(sessionID, payload, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to persist session: %w", err)
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookieName, sessionID, sessionMaxAgeDays*24*60*60, "/", "", s.secure, true)
	return nil
}

func (s *sessionStore) GetUserID(c *gin.Context) (int, bool) {
	if !s.enabled {
		return 0, false
	}

	sessionID, err := c.Cookie(sessionCookieName)
	if err != nil || sessionID == "" {
		return 0, false
	}

	rawPayload, err := s.db.GetActiveSessionData(sessionID)
	if err != nil {
		return 0, false
	}

	var payload sessionPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil || payload.UserID == 0 {
		return 0, false
	}

	return payload.UserID, true
}
