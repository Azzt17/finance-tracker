package model

import "time"

type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"
)

type Transaction struct {
	ID          string          `json:"id"`
	CategoryID  string          `json:"category_id"`
	Type        TransactionType `json:"type"`
	Amount      int64           `json:"amount"`
	Description string          `json:"description,omitempty"`
	OccurredAt  time.Time       `json:"occurred_at"`
	CreatedAt   time.Time       `json:"created_at"`
}

type TransactionInput struct {
	CategoryID  string          `json:"category_id"`
	Type        TransactionType `json:"type"`
	Amount      int64           `json:"amount"`
	Description string          `json:"description,omitempty"`
	OccurredAt  time.Time       `json:"occurred_at"`
}

type Budget struct {
	ID         string    `json:"id"`
	CategoryID string    `json:"category_id"`
	Limit      int64     `json:"limit"`
	Period     string    `json:"period"`
	CreatedAt  time.Time `json:"created_at"`
}

type BudgetInput struct {
	CategoryID string `json:"category_id"`
	Limit      int64  `json:"limit"`
	Period     string `json:"period"`
}

type Category struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Type      TransactionType `json:"type"`
	CreatedAt time.Time       `json:"created_at"`
}

type CategoryInput struct {
	Name string          `json:"name"`
	Type TransactionType `json:"type"`
}
