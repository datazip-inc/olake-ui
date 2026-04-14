package database

import (
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/models"
)

// CreateSession persists a session row by session key.
func (db *Database) CreateSession(sessionKey string, sessionData []byte, sessionExpiry time.Time) error {
	return db.conn.Create(&models.Session{
		SessionKey:    sessionKey,
		SessionData:   sessionData,
		SessionExpiry: sessionExpiry,
	}).Error
}

// GetActiveSessionData fetches session payload bytes for a non-expired session key.
func (db *Database) GetActiveSessionData(sessionKey string) ([]byte, error) {
	var session models.Session
	if err := db.conn.
		Where("session_key = ? AND session_expiry > NOW()", sessionKey).
		Take(&session).Error; err != nil {
		return nil, err
	}
	return session.SessionData, nil
}
