package utils

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

// GetLogRetentionPeriod returns how many days of per-sync log folders to keep,
// from LOG_RETENTION_PERIOD if set and valid, else constants.DefaultLogRetentionPeriod.
func GetLogRetentionPeriod() int {
	if val := os.Getenv("LOG_RETENTION_PERIOD"); val != "" {
		if days, err := strconv.Atoi(val); err == nil && days > 0 {
			return days
		}
	}
	return constants.DefaultLogRetentionPeriod
}

// hasStaleLogFile reports whether dirPath contains a .log/.log.gz file last
// modified before cutoff.
func hasStaleLogFile(dirPath string, cutoff time.Time) bool {
	found := false
	_ = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		if (strings.HasSuffix(path, ".log") || strings.HasSuffix(path, ".log.gz")) &&
			info.ModTime().Before(cutoff) {
			found = true
			return filepath.SkipDir
		}
		return nil
	})
	return found
}

// CleanOldLogs removes per-run folders directly under logDir (the NFS/filesystem
// persistent config dir) that are either empty or hold a sync log older than
// retentionDays. S3 storage mode has no local directory to sweep, so this is
// filesystem-only.
func CleanOldLogs(logDir string, retentionDays int) {
	logger.Info("running log cleaner...")
	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	entries, err := os.ReadDir(logDir)
	if err != nil {
		logger.Errorf("log cleaner: failed to read log dir %s: %s", logDir, err)
		return
	}

	for _, entry := range entries {
		// "telemetry" holds the install-wide telemetry id/file, not sync logs — never sweep it.
		if !entry.IsDir() || entry.Name() == "telemetry" {
			continue
		}

		dirPath := filepath.Join(logDir, entry.Name())
		subEntries, err := os.ReadDir(dirPath)
		stale := err != nil || len(subEntries) == 0 || hasStaleLogFile(dirPath, cutoff)
		if !stale {
			continue
		}

		logger.Infof("log cleaner: deleting stale folder %s", dirPath)
		if err := os.RemoveAll(dirPath); err != nil {
			logger.Errorf("log cleaner: failed to delete %s: %s", dirPath, err)
		}
	}
}

// StartLogCleaner runs an immediate cleanup pass (catching up on any cycles
// missed while the server was down) and then schedules CleanOldLogs to run
// every midnight for as long as the process is alive.
func StartLogCleaner(logDir string, retentionDays int) {
	logger.Infof("log cleaner started (retention: %d days)", retentionDays)
	CleanOldLogs(logDir, retentionDays)

	c := cron.New()
	if err := c.AddFunc("@midnight", func() {
		CleanOldLogs(logDir, retentionDays)
	}); err != nil {
		logger.Errorf("log cleaner: failed to schedule midnight run: %s", err)
		return
	}
	c.Start()
}
