package repository

import (
	"context"
	"database/sql"

	"github.com/Azzt17/finance-tracker/internal/model"
)

type ExportRepository struct {
	db       *sql.DB
	budget   *BudgetRepository
	tx       *TransactionRepository
	category *CategoryRepository
	savings  *SavingsRepository
}

func NewExportRepository(db *sql.DB) *ExportRepository {
	return &ExportRepository{
		db:       db,
		budget:   NewBudgetRepository(db),
		tx:       NewTransactionRepository(db),
		category: NewCategoryRepository(db),
		savings:  NewSavingsRepository(db),
	}
}

func (r *ExportRepository) GetAggregation(ctx context.Context, userID int64, yearMonth string) (model.BudgetAggregation, error) {
	return r.budget.GetAggregation(ctx, userID, yearMonth)
}

func (r *ExportRepository) ListTransactions(ctx context.Context, userID int64, yearMonth string) ([]model.Transaction, error) {
	return r.tx.List(ctx, userID, yearMonth, nil, 10000, 0)
}

func (r *ExportRepository) ExportFullData(ctx context.Context, userID int64) (model.UserDataExport, error) {
	var data model.UserDataExport
	data.Categories = []model.Category{}
	data.Transactions = []model.Transaction{}
	data.Budgets = []model.BudgetAllocation{}
	data.SavingsGoals = []model.SavingsGoal{}

	// Categories
	catRows, err := r.db.QueryContext(ctx, "SELECT id, name, icon_emoji, is_quick_add, sort_order, created_at FROM categories WHERE user_id = ?", userID)
	if err == nil {
		defer catRows.Close()
		for catRows.Next() {
			var c model.Category
			var iconEmoji sql.NullString
			if err := catRows.Scan(&c.ID, &c.Name, &iconEmoji, &c.IsQuickAdd, &c.SortOrder, &c.CreatedAt); err == nil {
				c.IconEmoji = iconEmoji.String
				data.Categories = append(data.Categories, c)
			}
		}
	}

	// Transactions
	txRows, err := r.db.QueryContext(ctx, "SELECT id, client_transaction_id, amount, category_id, note, transacted_at, created_at, is_synced FROM transactions WHERE user_id = ?", userID)
	if err == nil {
		defer txRows.Close()
		for txRows.Next() {
			var t model.Transaction
			var catID sql.NullInt64
			var note sql.NullString
			if err := txRows.Scan(&t.ID, &t.ClientTransactionID, &t.Amount, &catID, &note, &t.TransactedAt, &t.CreatedAt, &t.IsSynced); err == nil {
				if catID.Valid {
					c := catID.Int64
					t.CategoryID = &c
				}
				t.Note = note.String
				data.Transactions = append(data.Transactions, t)
			}
		}
	}

	// Budgets
	bRows, err := r.db.QueryContext(ctx, "SELECT id, year_month, total_budget, created_at, updated_at FROM budget_allocation WHERE user_id = ?", userID)
	if err == nil {
		defer bRows.Close()
		for bRows.Next() {
			var b model.BudgetAllocation
			if err := bRows.Scan(&b.ID, &b.YearMonth, &b.TotalBudget, &b.CreatedAt, &b.UpdatedAt); err == nil {
				data.Budgets = append(data.Budgets, b)
			}
		}
	}

	// Savings Goals
	sRows, err := r.db.QueryContext(ctx, "SELECT id, name, target_amount, current_saved, year_month, is_achieved, created_at FROM savings_goals WHERE user_id = ?", userID)
	if err == nil {
		defer sRows.Close()
		for sRows.Next() {
			var s model.SavingsGoal
			if err := sRows.Scan(&s.ID, &s.Name, &s.TargetAmount, &s.CurrentSaved, &s.YearMonth, &s.IsAchieved, &s.CreatedAt); err == nil {
				data.SavingsGoals = append(data.SavingsGoals, s)
			}
		}
	}

	return data, nil
}

func (r *ExportRepository) ImportFullData(ctx context.Context, userID int64, data model.UserDataExport) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Categories
	categoryIDMap := make(map[int64]int64)
	for _, c := range data.Categories {
		var newID int64
		err := tx.QueryRowContext(ctx, "SELECT id FROM categories WHERE user_id = ? AND name = ?", userID, c.Name).Scan(&newID)
		if err == sql.ErrNoRows {
			res, err := tx.ExecContext(ctx, "INSERT INTO categories (user_id, name, icon_emoji, is_quick_add, sort_order) VALUES (?, ?, ?, ?, ?)",
				userID, c.Name, c.IconEmoji, c.IsQuickAdd, c.SortOrder)
			if err != nil {
				return err
			}
			newID, _ = res.LastInsertId()
		} else if err != nil {
			return err
		}
		categoryIDMap[c.ID] = newID
	}

	// 2. Transactions
	for _, t := range data.Transactions {
		var newCatID *int64
		if t.CategoryID != nil {
			if mapped, ok := categoryIDMap[*t.CategoryID]; ok {
				id := mapped
				newCatID = &id
			}
		}

		_, err := tx.ExecContext(ctx, `
			INSERT INTO transactions (user_id, client_transaction_id, amount, category_id, note, transacted_at, is_synced)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(user_id, client_transaction_id) DO UPDATE SET
				amount = excluded.amount,
				category_id = excluded.category_id,
				note = excluded.note,
				transacted_at = excluded.transacted_at
		`, userID, t.ClientTransactionID, t.Amount, newCatID, t.Note, t.TransactedAt, 1)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

