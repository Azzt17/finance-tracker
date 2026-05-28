package handler

import (
	"context"
	"net/http"

	"github.com/Azzt17/finance-tracker/internal/model"
)

type SyncRepository interface {
	SyncTransactions(ctx context.Context, transactions []model.TransactionInput) ([]model.Transaction, error)
}

type SyncHandler struct {
	repository SyncRepository
}

func NewSyncHandler(repository SyncRepository) *SyncHandler {
	return &SyncHandler{repository: repository}
}

func (h *SyncHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/sync/transactions", h.syncTransactions)
}

func (h *SyncHandler) syncTransactions(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Transactions []model.TransactionInput `json:"transactions"`
	}

	if err := decodeJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	if len(payload.Transactions) == 0 {
		writeError(w, http.StatusBadRequest, "transactions list cannot be empty", "VALIDATION_ERROR")
		return
	}

	// For background sync, the client usually sends offline transactions.
	// We pass them to repository to perform batch insert/update (upsert by ClientTransactionID).
	synced, err := h.repository.SyncTransactions(r.Context(), payload.Transactions)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	if synced == nil {
		synced = []model.Transaction{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":      "sync successful",
		"synced_count": len(synced),
		"data":         synced,
	})
}
