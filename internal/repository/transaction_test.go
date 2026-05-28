package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/Azzt17/finance-tracker/internal/model"
	"github.com/Azzt17/finance-tracker/internal/repository"
)

func TestTransactionRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := repository.NewTransactionRepository(db)
	ctx := context.Background()

	input := model.TransactionInput{
		ClientTransactionID: "test-client-id-1",
		Amount:              50000,
		Note:                "Test Transaction",
		TransactedAt:        time.Now(),
	}

	// Create
	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created.Amount != input.Amount {
		t.Errorf("expected amount %d, got %d", input.Amount, created.Amount)
	}

	// Get
	fetched, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if fetched.ClientTransactionID != input.ClientTransactionID {
		t.Errorf("expected client id %s, got %s", input.ClientTransactionID, fetched.ClientTransactionID)
	}

	// List
	yearMonth := time.Now().Format("2006-01")
	txs, err := repo.List(ctx, yearMonth, nil, 100, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(txs) != 1 {
		t.Errorf("expected 1 tx, got %d", len(txs))
	}

	// Update
	input.Amount = 60000
	updated, err := repo.Update(ctx, created.ID, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Amount != 60000 {
		t.Errorf("expected amount 60000, got %d", updated.Amount)
	}

	// Delete
	err = repo.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	txs, _ = repo.List(ctx, yearMonth, nil, 100, 0)
	if len(txs) != 0 {
		t.Errorf("expected 0 tx after delete, got %d", len(txs))
	}
}
