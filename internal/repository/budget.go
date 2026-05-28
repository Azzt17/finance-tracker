package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Azzt17/finance-tracker/internal/model"
)

type BudgetRepository struct {
	db *sql.DB
}

func NewBudgetRepository(db *sql.DB) *BudgetRepository {
	return &BudgetRepository{db: db}
}

func (r *BudgetRepository) GetAggregation(ctx context.Context, yearMonth string) (model.BudgetAggregation, error) {
	var totalBudget int64
	err := r.db.QueryRowContext(ctx, "SELECT total_budget FROM budget_allocation WHERE year_month = ?", yearMonth).Scan(&totalBudget)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return model.BudgetAggregation{}, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.name, SUM(t.amount) as total
		FROM categories c
		JOIN transactions t ON c.id = t.category_id
		WHERE strftime('%Y-%m', t.transacted_at) = ?
		GROUP BY c.id, c.name
	`, yearMonth)
	if err != nil {
		return model.BudgetAggregation{}, err
	}
	defer rows.Close()

	var totalSpent int64
	var spendingByCategory []model.CategorySpending

	for rows.Next() {
		var cat model.CategorySpending
		if err := rows.Scan(&cat.CategoryID, &cat.CategoryName, &cat.Total); err != nil {
			return model.BudgetAggregation{}, err
		}
		totalSpent += cat.Total
		spendingByCategory = append(spendingByCategory, cat)
	}
	if err := rows.Err(); err != nil {
		return model.BudgetAggregation{}, err
	}

	if spendingByCategory == nil {
		spendingByCategory = []model.CategorySpending{}
	}

	return model.BudgetAggregation{
		YearMonth:          yearMonth,
		TotalBudget:        totalBudget,
		TotalSpent:         totalSpent,
		RemainingBalance:   totalBudget - totalSpent,
		SpendingByCategory: spendingByCategory,
	}, nil
}

func (r *BudgetRepository) SetTotalBudget(ctx context.Context, yearMonth string, totalBudget int64) error {
	var id int64
	err := r.db.QueryRowContext(ctx, "SELECT id FROM budget_allocation WHERE year_month = ?", yearMonth).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		_, err = r.db.ExecContext(ctx, "INSERT INTO budget_allocation (year_month, total_budget) VALUES (?, ?)", yearMonth, totalBudget)
		return err
	} else if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, "UPDATE budget_allocation SET total_budget = ?, updated_at = datetime('now') WHERE id = ?", totalBudget, id)
	return err
}
