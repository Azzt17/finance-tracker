package database

import (
	"context"
	"database/sql"
	"embed"
)

//go:embed migrations/*.up.sql
var migrationFiles embed.FS

type migration struct {
	version int
	name    string
	path    string
}

func Migrate(ctx context.Context, db *sql.DB) error {
	migrations := []migration{
		{
			version: 1,
			name:    "initial_schema",
			path:    "migrations/000001_initial_schema.up.sql",
		},
	}

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TEXT NOT NULL DEFAULT (datetime('now'))
		)
	`); err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	for _, migration := range migrations {
		applied, err := migrationApplied(ctx, tx, migration.version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		statement, err := migrationFiles.ReadFile(migration.path)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, string(statement)); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT OR IGNORE INTO schema_migrations (version, name)
			VALUES (?, ?)
		`, migration.version, migration.name); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func migrationApplied(ctx context.Context, tx *sql.Tx, version int) (bool, error) {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)
	`, version).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
