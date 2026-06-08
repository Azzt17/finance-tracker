package handler

import (
	"context"
	"net/http"

	"github.com/Azzt17/finance-tracker/internal/middleware"
	"github.com/Azzt17/finance-tracker/internal/model"
)

type BudgetRepository interface {
	GetAggregation(ctx context.Context, userID int64, yearMonth string) (model.BudgetAggregation, error)
	SetTotalBudget(ctx context.Context, userID int64, yearMonth string, totalBudget int64) error
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

	userID := middleware.UserFromContext(r.Context()).ID
	agg, err := h.repository.GetAggregation(r.Context(), userID, yearMonth)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	if agg.SpendingByCategory == nil {
		agg.SpendingByCategory = []model.CategorySpending{}
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

	userID := middleware.UserFromContext(r.Context()).ID
	if err := h.repository.SetTotalBudget(r.Context(), userID, yearMonth, input.TotalBudget); err != nil {
		writeRepositoryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "budget updated successfully"})
}
