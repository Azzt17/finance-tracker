package database

import (
	"context"
	"database/sql"
	"testing"
)

func TestMigrateCreatesInitialSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := Open(ctx, "file:migrations-test?mode=memory&cache=shared&_foreign_keys=on")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("migrate database twice: %v", err)
	}

	for _, table := range []string{"schema_migrations", "categories", "budgets", "transactions"} {
		assertTableExists(t, db, table)
	}

	var foreignKeysEnabled int
	if err := db.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&foreignKeysEnabled); err != nil {
		t.Fatalf("read foreign key pragma: %v", err)
	}
	if foreignKeysEnabled != 1 {
		t.Fatalf("expected foreign keys to be enabled, got %d", foreignKeysEnabled)
	}
}

func assertTableExists(t *testing.T, db *sql.DB, table string) {
	t.Helper()

	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`,
		table,
	).Scan(&count)
	if err != nil {
		t.Fatalf("query table %q: %v", table, err)
	}
	if count != 1 {
		t.Fatalf("expected table %q to exist", table)
	}
}
