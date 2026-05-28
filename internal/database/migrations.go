package database

import (
	"context"
	"database/sql"
	"embed"
)

//go:embed migrations/*.up.sql
var migrationFiles embed.FS

func Migrate(ctx context.Context, db *sql.DB) error {
	migrations := []string{
		"migrations/000001_initial_schema.up.sql",
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	for _, migration := range migrations {
		statement, err := migrationFiles.ReadFile(migration)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, string(statement)); err != nil {
			return err
		}
	}

	return tx.Commit()
}
