package handler

import (
	"net/http"
	"time"

	"github.com/Azzt17/finance-tracker/internal/model"
)

type BudgetHandler struct{}

func NewBudgetHandler() *BudgetHandler {
	return &BudgetHandler{}
}

func (h *BudgetHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/budgets", h.list)
	mux.HandleFunc("POST /api/v1/budgets", h.create)
	mux.HandleFunc("GET /api/v1/budgets/{id}", h.get)
	mux.HandleFunc("PUT /api/v1/budgets/{id}", h.update)
	mux.HandleFunc("DELETE /api/v1/budgets/{id}", h.delete)
}

func (h *BudgetHandler) list(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, envelope{
		"data": []model.Budget{},
	})
}

func (h *BudgetHandler) create(w http.ResponseWriter, r *http.Request) {
	var input model.BudgetInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	budget := model.Budget{
		ID:         "pending",
		CategoryID: input.CategoryID,
		Limit:      input.Limit,
		Period:     input.Period,
		CreatedAt:  time.Now().UTC(),
	}

	writeJSON(w, http.StatusCreated, envelope{
		"data": budget,
	})
}

func (h *BudgetHandler) get(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, envelope{
		"data": envelope{
			"id": r.PathValue("id"),
		},
	})
}

func (h *BudgetHandler) update(w http.ResponseWriter, r *http.Request) {
	var input model.BudgetInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, envelope{
		"data": envelope{
			"id": r.PathValue("id"),
		},
	})
}

func (h *BudgetHandler) delete(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNoContent, nil)
}
