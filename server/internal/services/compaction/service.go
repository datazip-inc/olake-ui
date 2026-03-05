package compaction

import (
	"os"

	"github.com/datazip-inc/olake-ui/server/internal/database"
)

// CompactionService is a unified service for compaction operations
type CompactionService struct {
	db     *database.Database
	client *Compaction
}

// InitCompactionService constructs a CompactionService with initialized client
func InitCompactionService(db *database.Database) (*CompactionService, error) {
	baseURL := os.Getenv("AMORO_BASE_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:1630"
	}

	apiKey := os.Getenv("AMORO_API_KEY")
	apiSecret := os.Getenv("AMORO_API_SECRET")

	client := NewClient(baseURL, apiKey, apiSecret)

	return &CompactionService{
		db:     db,
		client: client,
	}, nil
}

// GetClient returns the underlying Compaction client
func (s *CompactionService) GetClient() *Compaction {
	return s.client
}

// GetDB returns the database instance
func (s *CompactionService) GetDB() *database.Database {
	return s.db
}
