package database

import (
	"context"
	"database/sql"
)

func Migrate(ctx context.Context, db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS categories (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL CHECK (type IN ('income', 'expense')),
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_name_type
			ON categories (name, type)`,
		`CREATE TABLE IF NOT EXISTS budgets (
			id TEXT PRIMARY KEY,
			category_id TEXT NOT NULL,
			amount_limit INTEGER NOT NULL CHECK (amount_limit >= 0),
			period TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (category_id) REFERENCES categories (id)
				ON UPDATE CASCADE
				ON DELETE RESTRICT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_budgets_category_id
			ON budgets (category_id)`,
		`CREATE INDEX IF NOT EXISTS idx_budgets_period
			ON budgets (period)`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id TEXT PRIMARY KEY,
			category_id TEXT NOT NULL,
			type TEXT NOT NULL CHECK (type IN ('income', 'expense')),
			amount INTEGER NOT NULL CHECK (amount > 0),
			description TEXT,
			occurred_at TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (category_id) REFERENCES categories (id)
				ON UPDATE CASCADE
				ON DELETE RESTRICT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_category_id
			ON transactions (category_id)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_occurred_at
			ON transactions (occurred_at)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_type
			ON transactions (type)`,
		`INSERT OR IGNORE INTO schema_migrations (version, name)
			VALUES (1, 'initial_finance_schema')`,
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return err
		}
	}

	return tx.Commit()
}
