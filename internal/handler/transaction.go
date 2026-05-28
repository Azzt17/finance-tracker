package handler

import (
	"net/http"
	"time"

	"github.com/Azzt17/finance-tracker/internal/model"
)

type TransactionHandler struct{}

func NewTransactionHandler() *TransactionHandler {
	return &TransactionHandler{}
}

func (h *TransactionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/transactions", h.list)
	mux.HandleFunc("POST /api/v1/transactions", h.create)
	mux.HandleFunc("GET /api/v1/transactions/{id}", h.get)
	mux.HandleFunc("PUT /api/v1/transactions/{id}", h.update)
	mux.HandleFunc("DELETE /api/v1/transactions/{id}", h.delete)
}

func (h *TransactionHandler) list(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, envelope{
		"data": []model.Transaction{},
	})
}

func (h *TransactionHandler) create(w http.ResponseWriter, r *http.Request) {
	var input model.TransactionInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	transaction := model.Transaction{
		ID:          "pending",
		CategoryID:  input.CategoryID,
		Type:        input.Type,
		Amount:      input.Amount,
		Description: input.Description,
		OccurredAt:  input.OccurredAt,
		CreatedAt:   time.Now().UTC(),
	}

	writeJSON(w, http.StatusCreated, envelope{
		"data": transaction,
	})
}

func (h *TransactionHandler) get(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, envelope{
		"data": envelope{
			"id": r.PathValue("id"),
		},
	})
}

func (h *TransactionHandler) update(w http.ResponseWriter, r *http.Request) {
	var input model.TransactionInput
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

func (h *TransactionHandler) delete(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNoContent, nil)
}
