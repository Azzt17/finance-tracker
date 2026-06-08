package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Azzt17/finance-tracker/internal/middleware"
	"github.com/Azzt17/finance-tracker/internal/model"
)

type TransactionRepository interface {
	List(ctx context.Context, userID int64, yearMonth string, categoryID *int64, limit, offset int) ([]model.Transaction, error)
	Create(ctx context.Context, userID int64, input model.TransactionInput) (model.Transaction, error)
	Get(ctx context.Context, userID int64, id int64) (model.Transaction, error)
	Update(ctx context.Context, userID int64, id int64, input model.TransactionInput) (model.Transaction, error)
	Delete(ctx context.Context, userID int64, id int64) error
}

type TransactionHandler struct {
	repository TransactionRepository
}

func NewTransactionHandler(repository TransactionRepository) *TransactionHandler {
	return &TransactionHandler{repository: repository}
}

func (h *TransactionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/transactions", h.list)
	mux.HandleFunc("POST /api/v1/transactions", h.create)
	mux.HandleFunc("PATCH /api/v1/transactions/{id}", h.update)
	mux.HandleFunc("DELETE /api/v1/transactions/{id}", h.delete)
}

func (h *TransactionHandler) list(w http.ResponseWriter, r *http.Request) {
	yearMonth := r.URL.Query().Get("year_month")
	if yearMonth == "" {
		writeError(w, http.StatusBadRequest, "year_month is required (format YYYY-MM)", "VALIDATION_ERROR")
		return
	}

	var categoryID *int64
	if cidStr := r.URL.Query().Get("category_id"); cidStr != "" {
		cid, err := strconv.ParseInt(cidStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid category_id", "VALIDATION_ERROR")
			return
		}
		categoryID = &cid
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	userID := middleware.UserFromContext(r.Context()).ID
	transactions, err := h.repository.List(r.Context(), userID, yearMonth, categoryID, limit, offset)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	if transactions == nil {
		transactions = []model.Transaction{}
	}
	writeJSON(w, http.StatusOK, transactions)
}

func (h *TransactionHandler) create(w http.ResponseWriter, r *http.Request) {
	var input model.TransactionInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	if input.ClientTransactionID == "" {
		writeError(w, http.StatusBadRequest, "client_transaction_id is required", "VALIDATION_ERROR")
		return
	}
	if input.Amount == 0 {
		writeError(w, http.StatusBadRequest, "amount is required", "VALIDATION_ERROR")
		return
	}

	userID := middleware.UserFromContext(r.Context()).ID
	transaction, err := h.repository.Create(r.Context(), userID, input)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, transaction)
}

func (h *TransactionHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id", "VALIDATION_ERROR")
		return
	}

	var input model.TransactionInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	userID := middleware.UserFromContext(r.Context()).ID
	transaction, err := h.repository.Update(r.Context(), userID, id, input)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, transaction)
}

func (h *TransactionHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id", "VALIDATION_ERROR")
		return
	}
	userID := middleware.UserFromContext(r.Context()).ID
	if err := h.repository.Delete(r.Context(), userID, id); err != nil {
		writeRepositoryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted successfully"})
}
