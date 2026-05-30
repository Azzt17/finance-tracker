package service

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// extractDBPath parses the database URL to get the file path
func extractDBPath(databaseURL string) string {
	path := databaseURL
	path = strings.TrimPrefix(path, "file:")
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}
	return path
}

// BackupDatabase runs VACUUM INTO to create a safe backup of the database
func BackupDatabase(db *sql.DB, databaseURL string) (string, error) {
	dbPath := extractDBPath(databaseURL)

	// Create backup directory next to the database file
	backupDir := filepath.Join(filepath.Dir(dbPath), "backup")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	filename := fmt.Sprintf("finance-tracker_%s.db", time.Now().Format("20060102_150405"))
	destPath := filepath.Join(backupDir, filename)

	slog.Info("Starting database backup", "dest", destPath)

	query := fmt.Sprintf("VACUUM INTO '%s'", destPath)
	if _, err := db.Exec(query); err != nil {
		return "", fmt.Errorf("VACUUM INTO failed: %w", err)
	}

	slog.Info("Database backup completed successfully", "dest", destPath)
	return destPath, nil
}

// StartBackupCron starts a goroutine that backs up the database every day at 00:00
func StartBackupCron(db *sql.DB, databaseURL string) {
	go func() {
		for {
			now := time.Now()
			// Calculate time until next midnight
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			duration := next.Sub(now)

			slog.Info("Scheduled next database backup", "next_backup", next)
			time.Sleep(duration)

			if _, err := BackupDatabase(db, databaseURL); err != nil {
				slog.Error("Automated database backup failed", "error", err)
			}
		}
	}()
}
