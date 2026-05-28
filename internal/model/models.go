package model

import "time"

type Transaction struct {
	ID                  int64     `json:"id"`
	ClientTransactionID string    `json:"client_transaction_id"`
	Amount              int64     `json:"amount"`
	CategoryID          *int64    `json:"category_id,omitempty"`
	Note                string    `json:"note,omitempty"`
	TransactedAt        time.Time `json:"transacted_at"`
	CreatedAt           time.Time `json:"created_at"`
	IsSynced            bool      `json:"is_synced"`
}

type TransactionInput struct {
	ClientTransactionID string    `json:"client_transaction_id"`
	Amount              int64     `json:"amount"`
	CategoryID          *int64    `json:"category_id,omitempty"`
	Note                string    `json:"note,omitempty"`
	TransactedAt        time.Time `json:"transacted_at"`
	IsSynced            *bool     `json:"is_synced,omitempty"`
	IsDeleted           *bool     `json:"is_deleted,omitempty"`
}

type BudgetAllocation struct {
	ID          int64     `json:"id"`
	YearMonth   string    `json:"year_month"`
	TotalBudget int64     `json:"total_budget"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BudgetAllocationInput struct {
	YearMonth   string `json:"year_month"`
	TotalBudget int64  `json:"total_budget"`
}

type CategorySpending struct {
	CategoryID   int64  `json:"category_id"`
	CategoryName string `json:"category_name"`
	Total        int64  `json:"total"`
}

type BudgetAggregation struct {
	YearMonth          string             `json:"year_month"`
	TotalBudget        int64              `json:"total_budget"`
	TotalSpent         int64              `json:"total_spent"`
	RemainingBalance   int64              `json:"remaining_balance"`
	SpendingByCategory []CategorySpending `json:"spending_by_category"`
}

type Category struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	IconEmoji  string    `json:"icon_emoji,omitempty"`
	IsQuickAdd bool      `json:"is_quick_add"`
	SortOrder  int64     `json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
}

type CategoryInput struct {
	Name       string `json:"name"`
	IconEmoji  string `json:"icon_emoji,omitempty"`
	IsQuickAdd bool   `json:"is_quick_add"`
	SortOrder  int64  `json:"sort_order"`
}

type SavingsGoal struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	TargetAmount int64     `json:"target_amount"`
	CurrentSaved int64     `json:"current_saved"`
	YearMonth    string    `json:"year_month"`
	IsAchieved   bool      `json:"is_achieved"`
	CreatedAt    time.Time `json:"created_at"`
}

type SavingsGoalInput struct {
	Name         string `json:"name"`
	TargetAmount int64  `json:"target_amount"`
	CurrentSaved int64  `json:"current_saved"`
	YearMonth    string `json:"year_month"`
	IsAchieved   bool   `json:"is_achieved"`
}
