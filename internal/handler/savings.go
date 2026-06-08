package handler

import (
	"context"
	"net/http"

	"github.com/Azzt17/finance-tracker/internal/middleware"
	"github.com/Azzt17/finance-tracker/internal/model"
)

type SavingsRepository interface {
	List(ctx context.Context, userID int64, yearMonth string) ([]model.SavingsGoal, error)
	Create(ctx context.Context, userID int64, input model.SavingsGoalInput) (model.SavingsGoal, error)
	Update(ctx context.Context, userID int64, id int64, input model.SavingsGoalInput) (model.SavingsGoal, error)
	Delete(ctx context.Context, userID int64, id int64) error
}

type SavingsGoalHandler struct {
	repository SavingsRepository
}

func NewSavingsGoalHandler(repository SavingsRepository) *SavingsGoalHandler {
	return &SavingsGoalHandler{repository: repository}
}

func (h *SavingsGoalHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/savings", h.list)
	mux.HandleFunc("POST /api/v1/savings", h.create)
	mux.HandleFunc("PATCH /api/v1/savings/{id}", h.update)
	mux.HandleFunc("DELETE /api/v1/savings/{id}", h.delete)
}

func (h *SavingsGoalHandler) list(w http.ResponseWriter, r *http.Request) {
	yearMonth := r.URL.Query().Get("year_month")
	if yearMonth == "" {
		writeError(w, http.StatusBadRequest, "year_month is required", "VALIDATION_ERROR")
		return
	}

	userID := middleware.UserFromContext(r.Context()).ID
	goals, err := h.repository.List(r.Context(), userID, yearMonth)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	if goals == nil {
		goals = []model.SavingsGoal{}
	}
	writeJSON(w, http.StatusOK, goals)
}

func (h *SavingsGoalHandler) create(w http.ResponseWriter, r *http.Request) {
	var input model.SavingsGoalInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	if input.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required", "VALIDATION_ERROR")
		return
	}
	if input.TargetAmount == 0 {
		writeError(w, http.StatusBadRequest, "target_amount is required", "VALIDATION_ERROR")
		return
	}
	if input.YearMonth == "" {
		writeError(w, http.StatusBadRequest, "year_month is required", "VALIDATION_ERROR")
		return
	}

	userID := middleware.UserFromContext(r.Context()).ID
	goal, err := h.repository.Create(r.Context(), userID, input)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, goal)
}

func (h *SavingsGoalHandler) update(w http.ResponseWriter, r *http.Request) {
	var input model.SavingsGoalInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid savings id", "VALIDATION_ERROR")
		return
	}

	userID := middleware.UserFromContext(r.Context()).ID
	goal, err := h.repository.Update(r.Context(), userID, id, input)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, goal)
}

func (h *SavingsGoalHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid savings id", "VALIDATION_ERROR")
		return
	}
	userID := middleware.UserFromContext(r.Context()).ID
	if err := h.repository.Delete(r.Context(), userID, id); err != nil {
		writeRepositoryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted successfully"})
}
