package handler

import (
	"context"
	"net/http"
)

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

type BudgetRepository interface {
	GetAggregation(ctx context.Context, yearMonth string) (BudgetAggregation, error)
	SetTotalBudget(ctx context.Context, yearMonth string, totalBudget int64) error
}

type BudgetHandler struct {
	repository BudgetRepository
}

func NewBudgetHandler(repository BudgetRepository) *BudgetHandler {
	return &BudgetHandler{repository: repository}
}

func (h *BudgetHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/budget/{year_month}", h.get)
	mux.HandleFunc("PUT /api/v1/budget/{year_month}", h.update)
}

func (h *BudgetHandler) get(w http.ResponseWriter, r *http.Request) {
	yearMonth := r.PathValue("year_month")
	if yearMonth == "" {
		writeError(w, http.StatusBadRequest, "year_month is required", "VALIDATION_ERROR")
		return
	}

	agg, err := h.repository.GetAggregation(r.Context(), yearMonth)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	if agg.SpendingByCategory == nil {
		agg.SpendingByCategory = []CategorySpending{}
	}
	writeJSON(w, http.StatusOK, agg)
}

func (h *BudgetHandler) update(w http.ResponseWriter, r *http.Request) {
	yearMonth := r.PathValue("year_month")
	if yearMonth == "" {
		writeError(w, http.StatusBadRequest, "year_month is required", "VALIDATION_ERROR")
		return
	}

	var input struct {
		TotalBudget int64 `json:"total_budget"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	if err := h.repository.SetTotalBudget(r.Context(), yearMonth, input.TotalBudget); err != nil {
		writeRepositoryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "budget updated successfully"})
}
