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
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close database: %v", err)
		}
	})

	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("migrate database twice: %v", err)
	}

	for _, table := range []string{
		"schema_migrations",
		"categories",
		"budget_allocation",
		"savings_goals",
		"transactions",
	} {
		assertTableExists(t, db, table)
	}

	assertColumn(t, db, "categories", "id", "INTEGER", true, nil, true)
	assertColumn(t, db, "categories", "name", "TEXT", true, nil, false)
	assertColumn(t, db, "categories", "icon_emoji", "TEXT", false, nil, false)
	assertColumn(t, db, "categories", "is_quick_add", "INTEGER", true, strPtr("0"), false)
	assertColumn(t, db, "categories", "sort_order", "INTEGER", true, strPtr("0"), false)
	assertColumn(t, db, "categories", "created_at", "TEXT", true, strPtr("datetime('now')"), false)

	assertColumn(t, db, "budget_allocation", "id", "INTEGER", true, nil, true)
	assertColumn(t, db, "budget_allocation", "year_month", "TEXT", true, nil, false)
	assertColumn(t, db, "budget_allocation", "total_budget", "INTEGER", true, nil, false)
	assertColumn(t, db, "budget_allocation", "created_at", "TEXT", true, strPtr("datetime('now')"), false)
	assertColumn(t, db, "budget_allocation", "updated_at", "TEXT", true, strPtr("datetime('now')"), false)

	assertColumn(t, db, "savings_goals", "id", "INTEGER", true, nil, true)
	assertColumn(t, db, "savings_goals", "name", "TEXT", true, nil, false)
	assertColumn(t, db, "savings_goals", "target_amount", "INTEGER", true, nil, false)
	assertColumn(t, db, "savings_goals", "current_saved", "INTEGER", true, strPtr("0"), false)
	assertColumn(t, db, "savings_goals", "year_month", "TEXT", true, nil, false)
	assertColumn(t, db, "savings_goals", "is_achieved", "INTEGER", true, strPtr("0"), false)
	assertColumn(t, db, "savings_goals", "created_at", "TEXT", true, strPtr("datetime('now')"), false)

	assertColumn(t, db, "transactions", "id", "INTEGER", true, nil, true)
	assertColumn(t, db, "transactions", "client_transaction_id", "TEXT", true, nil, false)
	assertColumn(t, db, "transactions", "amount", "INTEGER", true, nil, false)
	assertColumn(t, db, "transactions", "category_id", "INTEGER", false, nil, false)
	assertColumn(t, db, "transactions", "note", "TEXT", false, nil, false)
	assertColumn(t, db, "transactions", "transacted_at", "TEXT", true, strPtr("datetime('now')"), false)
	assertColumn(t, db, "transactions", "created_at", "TEXT", true, strPtr("datetime('now')"), false)
	assertColumn(t, db, "transactions", "is_synced", "INTEGER", true, strPtr("1"), false)

	assertForeignKey(t, db, "transactions", "category_id", "categories", "id")

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

func assertColumn(
	t *testing.T,
	db *sql.DB,
	table string,
	name string,
	columnType string,
	notNull bool,
	defaultValue *string,
	primaryKey bool,
) {
	t.Helper()

	rows, err := db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		t.Fatalf("query columns for %q: %v", table, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.Fatalf("close columns for %q: %v", table, err)
		}
	}()

	for rows.Next() {
		var (
			cid           int
			gotName       string
			gotType       string
			gotNotNull    int
			gotDefault    sql.NullString
			gotPrimaryKey int
		)
		if err := rows.Scan(&cid, &gotName, &gotType, &gotNotNull, &gotDefault, &gotPrimaryKey); err != nil {
			t.Fatalf("scan column for %q: %v", table, err)
		}
		if gotName != name {
			continue
		}

		if gotType != columnType {
			t.Fatalf("expected %s.%s type %q, got %q", table, name, columnType, gotType)
		}
		if (gotNotNull == 1) != notNull && !primaryKey {
			t.Fatalf("expected %s.%s not_null=%t, got %d", table, name, notNull, gotNotNull)
		}
		if defaultValue == nil && gotDefault.Valid {
			t.Fatalf("expected %s.%s to have no default, got %q", table, name, gotDefault.String)
		}
		if defaultValue != nil {
			if !gotDefault.Valid {
				t.Fatalf("expected %s.%s default %q, got none", table, name, *defaultValue)
			}
			if gotDefault.String != *defaultValue {
				t.Fatalf("expected %s.%s default %q, got %q", table, name, *defaultValue, gotDefault.String)
			}
		}
		if (gotPrimaryKey == 1) != primaryKey {
			t.Fatalf("expected %s.%s primary_key=%t, got %d", table, name, primaryKey, gotPrimaryKey)
		}
		return
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate columns for %q: %v", table, err)
	}

	t.Fatalf("expected column %s.%s to exist", table, name)
}

func strPtr(value string) *string {
	return &value
}

func assertForeignKey(
	t *testing.T,
	db *sql.DB,
	table string,
	fromColumn string,
	toTable string,
	toColumn string,
) {
	t.Helper()

	rows, err := db.Query(`PRAGMA foreign_key_list(` + table + `)`)
	if err != nil {
		t.Fatalf("query foreign keys for %q: %v", table, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.Fatalf("close foreign keys for %q: %v", table, err)
		}
	}()

	for rows.Next() {
		var (
			id       int
			seq      int
			gotTable string
			gotFrom  string
			gotTo    string
			onUpdate string
			onDelete string
			match    string
		)
		if err := rows.Scan(&id, &seq, &gotTable, &gotFrom, &gotTo, &onUpdate, &onDelete, &match); err != nil {
			t.Fatalf("scan foreign key for %q: %v", table, err)
		}
		if gotFrom == fromColumn && gotTable == toTable && gotTo == toColumn {
			return
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate foreign keys for %q: %v", table, err)
	}

	t.Fatalf("expected foreign key %s.%s -> %s.%s", table, fromColumn, toTable, toColumn)
}
