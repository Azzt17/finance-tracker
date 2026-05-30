package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/Azzt17/finance-tracker/internal/repository"
	_ "github.com/mattn/go-sqlite3"
)

func TestAnalyticsRepository_GetSpendingByCategory(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE categories (id INTEGER PRIMARY KEY, name TEXT, icon_emoji TEXT);
		CREATE TABLE transactions (id INTEGER PRIMARY KEY, category_id INTEGER, amount INTEGER, transacted_at DATETIME);
	`)
	if err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	repo := repository.NewAnalyticsRepository(db)
	ctx := context.Background()

	res, err := repo.GetSpendingByCategory(ctx, "2026-05")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Total != 0 || len(res.Breakdown) != 0 {
		t.Errorf("expected empty result, got total %d, breakdown %d", res.Total, len(res.Breakdown))
	}

	db.Exec(`INSERT INTO categories (id, name, icon_emoji) VALUES (1, 'Food', '🍔')`)
	db.Exec(`INSERT INTO transactions (category_id, amount, transacted_at) VALUES (1, 100, '2026-05-01 10:00:00')`)
	db.Exec(`INSERT INTO transactions (category_id, amount, transacted_at) VALUES (NULL, 50, '2026-05-02 10:00:00')`)
	db.Exec(`INSERT INTO transactions (category_id, amount, transacted_at) VALUES (1, -20, '2026-05-03 10:00:00')`)

	res, err = repo.GetSpendingByCategory(ctx, "2026-05")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Total != 150 {
		t.Errorf("expected total 150, got %d", res.Total)
	}
	if len(res.Breakdown) != 2 {
		t.Errorf("expected 2 categories, got %d", len(res.Breakdown))
	}
}

func TestAnalyticsRepository_GetMonthlyTrend(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE transactions (id INTEGER PRIMARY KEY, amount INTEGER, transacted_at DATETIME);
		CREATE TABLE budget_allocation (id INTEGER PRIMARY KEY, year_month TEXT, total_budget INTEGER);
	`)
	if err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	repo := repository.NewAnalyticsRepository(db)
	ctx := context.Background()

	res, err := repo.GetMonthlyTrend(ctx, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Data) != 3 {
		t.Errorf("expected 3 months data, got %d", len(res.Data))
	}
}

func TestAnalyticsRepository_GetDailySpending(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE transactions (id INTEGER PRIMARY KEY, amount INTEGER, transacted_at DATETIME);
	`)
	if err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	repo := repository.NewAnalyticsRepository(db)
	ctx := context.Background()

	res, err := repo.GetDailySpending(ctx, "2026-05")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Data) != 31 {
		t.Errorf("expected 31 days, got %d", len(res.Data))
	}

	_, err = repo.GetDailySpending(ctx, "invalid")
	if err == nil {
		t.Errorf("expected error on invalid format")
	}
}
