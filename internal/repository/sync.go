package repository

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/Azzt17/finance-tracker/internal/model"
)

type SyncRepository struct {
	db *sql.DB
}

func NewSyncRepository(db *sql.DB) *SyncRepository {
	return &SyncRepository{db: db}
}

func (r *SyncRepository) SyncTransactions(ctx context.Context, inputs []model.TransactionInput) ([]model.Transaction, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var synced []model.Transaction

	for _, input := range inputs {
		var exists bool
		var id int64
		err := tx.QueryRowContext(ctx, "SELECT id FROM transactions WHERE client_transaction_id = ?", input.ClientTransactionID).Scan(&id)
		if err == nil {
			exists = true
		} else if err != sql.ErrNoRows {
			return nil, err
		}

		if exists {
			_, err = tx.ExecContext(ctx, `
				UPDATE transactions
				SET amount = ?, category_id = ?, note = ?, transacted_at = COALESCE(?, transacted_at), is_synced = 1
				WHERE id = ?
			`, input.Amount, nullableInt64(input.CategoryID), input.Note, dbTime(input.TransactedAt), id)
			if err != nil {
				slog.Error("sync update failed", "error", err)
				continue
			}
		} else {
			res, err := tx.ExecContext(ctx, `
				INSERT INTO transactions (client_transaction_id, amount, category_id, note, transacted_at, is_synced)
				VALUES (?, ?, ?, ?, COALESCE(?, datetime('now')), 1)
			`, input.ClientTransactionID, input.Amount, nullableInt64(input.CategoryID), input.Note, dbTime(input.TransactedAt))
			if err != nil {
				slog.Error("sync insert failed", "error", err)
				continue
			}
			id, _ = res.LastInsertId()
		}

		synced = append(synced, model.Transaction{
			ID:                  id,
			ClientTransactionID: input.ClientTransactionID,
			Amount:              input.Amount,
			CategoryID:          input.CategoryID,
			Note:                input.Note,
			TransactedAt:        input.TransactedAt,
			IsSynced:            true,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	if synced == nil {
		synced = []model.Transaction{}
	}

	return synced, nil
}
