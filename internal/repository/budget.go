package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Azzt17/finance-tracker/internal/model"
)

type BudgetAllocationRepository struct {
	db *sql.DB
}

func NewBudgetAllocationRepository(db *sql.DB) *BudgetAllocationRepository {
	return &BudgetAllocationRepository{db: db}
}

func (r *BudgetAllocationRepository) List(ctx context.Context) (budgets []model.BudgetAllocation, err error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, year_month, total_budget, created_at, updated_at
		FROM budget_allocation
		ORDER BY year_month DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	budgets = []model.BudgetAllocation{}
	for rows.Next() {
		budget, err := scanBudgetAllocation(rows)
		if err != nil {
			return nil, err
		}
		budgets = append(budgets, budget)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return budgets, nil
}

func (r *BudgetAllocationRepository) Create(ctx context.Context, input model.BudgetAllocationInput) (model.BudgetAllocation, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO budget_allocation (year_month, total_budget)
		VALUES (?, ?)
	`, input.YearMonth, input.TotalBudget)
	if err != nil {
		return model.BudgetAllocation{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.BudgetAllocation{}, err
	}

	return r.Get(ctx, id)
}

func (r *BudgetAllocationRepository) Get(ctx context.Context, id int64) (model.BudgetAllocation, error) {
	budget, err := scanBudgetAllocation(r.db.QueryRowContext(ctx, `
		SELECT id, year_month, total_budget, created_at, updated_at
		FROM budget_allocation
		WHERE id = ?
	`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return model.BudgetAllocation{}, ErrNotFound
	}
	if err != nil {
		return model.BudgetAllocation{}, err
	}

	return budget, nil
}

func (r *BudgetAllocationRepository) Update(ctx context.Context, id int64, input model.BudgetAllocationInput) (model.BudgetAllocation, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE budget_allocation
		SET year_month = ?, total_budget = ?, updated_at = datetime('now')
		WHERE id = ?
	`, input.YearMonth, input.TotalBudget, id)
	if err != nil {
		return model.BudgetAllocation{}, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return model.BudgetAllocation{}, err
	}
	if affected == 0 {
		return model.BudgetAllocation{}, ErrNotFound
	}

	return r.Get(ctx, id)
}

func (r *BudgetAllocationRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM budget_allocation WHERE id = ?`, id)
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

type budgetAllocationScanner interface {
	Scan(dest ...any) error
}

func scanBudgetAllocation(scanner budgetAllocationScanner) (model.BudgetAllocation, error) {
	var (
		budget    model.BudgetAllocation
		createdAt string
		updatedAt string
	)
	if err := scanner.Scan(
		&budget.ID,
		&budget.YearMonth,
		&budget.TotalBudget,
		&createdAt,
		&updatedAt,
	); err != nil {
		return model.BudgetAllocation{}, err
	}

	parsedCreatedAt, err := parseDBTime(createdAt)
	if err != nil {
		return model.BudgetAllocation{}, err
	}
	parsedUpdatedAt, err := parseDBTime(updatedAt)
	if err != nil {
		return model.BudgetAllocation{}, err
	}

	budget.CreatedAt = parsedCreatedAt
	budget.UpdatedAt = parsedUpdatedAt

	return budget, nil
}
