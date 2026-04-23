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
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

type sessionPayload struct {
	UserID int `json:"user_id"`
}

type sessionStore struct {
	db      *database.Database
	enabled bool
}

const (
	sessionCookieName = "olake-session"
	sessionMaxAgeDays = 30
)

func newSessionStore(cfg *appconfig.Config, db *database.Database) *sessionStore {
	return &sessionStore{
		db:      db,
		enabled: cfg.SessionOn,
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

	err = s.db.CreateSession(sessionID, payload, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to persist session: %w", err)
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookieName, sessionID, sessionMaxAgeDays*24*60*60, "/", "", isSecureRequest(c), true)
	return nil
}

func isSecureRequest(c *gin.Context) bool {
	// Direct HTTPS connection to this server.
	if c.Request.TLS != nil {
		return true
	}
	// HTTPS terminated at a reverse proxy/load balancer.
	return c.GetHeader("X-Forwarded-Proto") == "https"
}

func (s *sessionStore) GetUserID(c *gin.Context) (int, bool) {
	if !s.enabled {
		return 0, false
	}

	sessionID, err := c.Cookie(sessionCookieName)
	if err != nil || sessionID == "" {
		logger.Errorf("failed to get session cookie: %s", utils.Ternary(err != nil, err, "no session cookie found"))
		return 0, false
	}

	rawPayload, err := s.db.GetActiveSessionData(sessionID)
	if err != nil {
		logger.Errorf("failed to get active session data: %s", err)
		return 0, false
	}

	var payload sessionPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil || payload.UserID == 0 {
		logger.Errorf("failed to unmarshal session payload: %s", err)
		return 0, false
	}

	return payload.UserID, true
}
