package database

import (
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/models"
	"gorm.io/gorm/clause"
)

// UpsertSession persists or updates a session row by session key.
func (db *Database) UpsertSession(sessionKey string, sessionData []byte, sessionExpiry time.Time) error {
	return db.conn.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "session_key"}},
			DoUpdates: clause.AssignmentColumns([]string{"session_data", "session_expiry"}),
		}).
		Create(&models.Session{
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
