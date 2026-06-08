package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/Azzt17/finance-tracker/internal/model"
	"github.com/Azzt17/finance-tracker/internal/repository"
)

func TestSavingsRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := repository.NewSavingsRepository(db)
	ctx := context.Background()

	yearMonth := time.Now().Format("2006-01")
	input := model.SavingsGoalInput{
		Name:         "Test Savings",
		TargetAmount: 1000000,
		CurrentSaved: 200000,
		YearMonth:    yearMonth,
		IsAchieved:   false,
	}

	// Create
	created, err := repo.Create(ctx, 1, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created.Name != input.Name {
		t.Errorf("expected name %s, got %s", input.Name, created.Name)
	}

	// List
	goals, err := repo.List(ctx, 1, yearMonth)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(goals) != 1 {
		t.Errorf("expected 1 goal, got %d", len(goals))
	}

	// Update
	input.CurrentSaved = 500000
	updated, err := repo.Update(ctx, 1, created.ID, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.CurrentSaved != 500000 {
		t.Errorf("expected amount 500000, got %d", updated.CurrentSaved)
	}

	// Delete
	err = repo.Delete(ctx, 1, created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	goals, _ = repo.List(ctx, 1, yearMonth)
	if len(goals) != 0 {
		t.Errorf("expected 0 goal after delete, got %d", len(goals))
	}
}

func TestSavingsRepository_UpdateNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := repository.NewSavingsRepository(db)
	ctx := context.Background()

	_, err := repo.Update(ctx, 1, 999, model.SavingsGoalInput{
		Name:         "Missing",
		TargetAmount: 1000000,
		CurrentSaved: 100000,
		YearMonth:    "2026-05",
	})
	if err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
