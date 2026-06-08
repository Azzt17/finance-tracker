package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/Azzt17/finance-tracker/internal/model"
	"github.com/Azzt17/finance-tracker/internal/repository"
)

func TestSyncRepository_SyncTransactions(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := repository.NewSyncRepository(db)
	ctx := context.Background()

	inputs := []model.TransactionInput{
		{
			ClientTransactionID: "sync-1",
			Amount:              10000,
			Note:                "Sync Test 1",
			TransactedAt:        time.Now(),
		},
		{
			ClientTransactionID: "sync-2",
			Amount:              20000,
			Note:                "Sync Test 2",
			TransactedAt:        time.Now(),
		},
	}

	_, err := repo.SyncTransactions(ctx, 1, inputs)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify they are inserted
	txRepo := repository.NewTransactionRepository(db)
	txs, err := txRepo.List(ctx, 1, time.Now().Format("2006-01"), nil, 100, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(txs) != 2 {
		t.Errorf("expected 2 synced tx, got %d", len(txs))
	}
}
