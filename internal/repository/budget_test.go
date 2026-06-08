package repository_test

import (
	"context"
	"testing"

	"github.com/Azzt17/finance-tracker/internal/repository"
)

func TestBudgetRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := repository.NewBudgetRepository(db)
	ctx := context.Background()

	yearMonth := "2026-05"

	// Set Budget
	err := repo.SetTotalBudget(ctx, 1, yearMonth, 5000000)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Update Budget
	err = repo.SetTotalBudget(ctx, 1, yearMonth, 6000000)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Get Aggregation
	agg, err := repo.GetAggregation(ctx, 1, yearMonth)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if agg.TotalBudget != 6000000 {
		t.Errorf("expected budget 6000000, got %d", agg.TotalBudget)
	}
}
