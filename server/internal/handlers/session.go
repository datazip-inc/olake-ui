package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/utils"
)

type sessionPayload struct {
	UserID int `json:"user_id"`
}

type sessionStore struct {
	db      *sql.DB
	enabled bool
	secure  bool
}

const (
	sessionCookieName = "olake-session"
	sessionMaxAgeDays = 30
)

func newSessionStore(cfg appconfig.Config) (*sessionStore, error) {
	store := &sessionStore{
		enabled: cfg.SessionOn,
		secure:  cfg.RunMode != "localdev",
	}
	if !store.enabled {
		return store, nil
	}

	dsn, err := database.BuildPostgresURIFromConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build postgres URI for session store: %w", err)
	}

	dbConn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect postgres session store: %w", err)
	}
	if err := dbConn.Ping(); err != nil {
		_ = dbConn.Close()
		return nil, fmt.Errorf("failed to ping postgres session store: %w", err)
	}

	store.db = dbConn
	return store, nil
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

	_, err = s.db.Exec(`
		INSERT INTO session (session_key, session_data, session_expiry)
		VALUES ($1, $2, $3)
		ON CONFLICT (session_key)
		DO UPDATE SET session_data = EXCLUDED.session_data, session_expiry = EXCLUDED.session_expiry
	`, sessionID, payload, expiresAt)
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

	var rawPayload []byte
	err = s.db.QueryRow(`
		SELECT session_data
		FROM session
		WHERE session_key = $1 AND session_expiry > NOW()
	`, sessionID).Scan(&rawPayload)
	if err != nil {
		return 0, false
	}

	var payload sessionPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil || payload.UserID == 0 {
		return 0, false
	}

	return payload.UserID, true
}
