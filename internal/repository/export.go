package repository

import (
	"context"
	"database/sql"

	"github.com/Azzt17/finance-tracker/internal/model"
)

type ExportRepository struct {
	db     *sql.DB
	budget *BudgetRepository
	tx     *TransactionRepository
}

func NewExportRepository(db *sql.DB) *ExportRepository {
	return &ExportRepository{
		db:     db,
		budget: NewBudgetRepository(db),
		tx:     NewTransactionRepository(db),
	}
}

func (r *ExportRepository) GetAggregation(ctx context.Context, yearMonth string) (model.BudgetAggregation, error) {
	return r.budget.GetAggregation(ctx, yearMonth)
}

func (r *ExportRepository) ListTransactions(ctx context.Context, yearMonth string) ([]model.Transaction, error) {
	return r.tx.List(ctx, yearMonth, nil, 10000, 0)
}
