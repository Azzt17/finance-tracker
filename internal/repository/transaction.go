package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Azzt17/finance-tracker/internal/model"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) List(ctx context.Context, userID int64, yearMonth string, categoryID *int64, limit, offset int) (transactions []model.Transaction, err error) {
	query := `
		SELECT id,
			client_transaction_id,
			amount,
			category_id,
			note,
			transacted_at,
			created_at,
			is_synced
		FROM transactions
		WHERE strftime('%Y-%m', transacted_at) = ? AND user_id = ?
	`
	args := []any{yearMonth, userID}

	if categoryID != nil {
		query += " AND category_id = ?"
		args = append(args, *categoryID)
	}

	query += " ORDER BY transacted_at DESC, id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	transactions = []model.Transaction{}
	for rows.Next() {
		transaction, err := scanTransaction(rows)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *TransactionRepository) Create(ctx context.Context, userID int64, input model.TransactionInput) (model.Transaction, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO transactions (
			user_id,
			client_transaction_id,
			amount,
			category_id,
			note,
			transacted_at,
			is_synced
		)
		VALUES (?, ?, ?, ?, ?, COALESCE(?, datetime('now')), COALESCE(?, 1))
	`, userID, input.ClientTransactionID, input.Amount, nullableInt64(input.CategoryID), nullableString(input.Note), dbTime(input.TransactedAt), nullableBoolInt(input.IsSynced))
	if err != nil {
		return model.Transaction{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Transaction{}, err
	}

	return r.Get(ctx, userID, id)
}

func (r *TransactionRepository) Get(ctx context.Context, userID int64, id int64) (model.Transaction, error) {
	transaction, err := scanTransaction(r.db.QueryRowContext(ctx, `
		SELECT id,
			client_transaction_id,
			amount,
			category_id,
			note,
			transacted_at,
			created_at,
			is_synced
		FROM transactions
		WHERE id = ? AND user_id = ?
	`, id, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return model.Transaction{}, ErrNotFound
	}
	if err != nil {
		return model.Transaction{}, err
	}

	return transaction, nil
}

func (r *TransactionRepository) Update(ctx context.Context, userID int64, id int64, input model.TransactionInput) (model.Transaction, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE transactions
		SET client_transaction_id = ?,
			amount = ?,
			category_id = ?,
			note = ?,
			transacted_at = COALESCE(?, transacted_at),
			is_synced = COALESCE(?, is_synced)
		WHERE id = ? AND user_id = ?
	`, input.ClientTransactionID, input.Amount, nullableInt64(input.CategoryID), nullableString(input.Note), dbTime(input.TransactedAt), nullableBoolInt(input.IsSynced), id, userID)
	if err != nil {
		return model.Transaction{}, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return model.Transaction{}, err
	}
	if affected == 0 {
		return model.Transaction{}, ErrNotFound
	}

	return r.Get(ctx, userID, id)
}

func (r *TransactionRepository) Delete(ctx context.Context, userID int64, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM transactions WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

type transactionScanner interface {
	Scan(dest ...any) error
}

func scanTransaction(scanner transactionScanner) (model.Transaction, error) {
	var (
		transaction  model.Transaction
		categoryID   sql.NullInt64
		note         sql.NullString
		transactedAt string
		createdAt    string
		isSynced     int
	)
	if err := scanner.Scan(
		&transaction.ID,
		&transaction.ClientTransactionID,
		&transaction.Amount,
		&categoryID,
		&note,
		&transactedAt,
		&createdAt,
		&isSynced,
	); err != nil {
		return model.Transaction{}, err
	}

	parsedTransactedAt, err := parseDBTime(transactedAt)
	if err != nil {
		return model.Transaction{}, err
	}
	parsedCreatedAt, err := parseDBTime(createdAt)
	if err != nil {
		return model.Transaction{}, err
	}

	transaction.CategoryID = int64FromNull(categoryID)
	transaction.Note = stringFromNull(note)
	transaction.TransactedAt = parsedTransactedAt
	transaction.CreatedAt = parsedCreatedAt
	transaction.IsSynced = intToBool(isSynced)

	return transaction, nil
}
