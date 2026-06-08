package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Azzt17/finance-tracker/internal/handler"
	"github.com/Azzt17/finance-tracker/internal/middleware"
	"github.com/Azzt17/finance-tracker/internal/model"
)

type mockSyncRepo struct {
	synced []model.Transaction
}

func (m *mockSyncRepo) SyncTransactions(ctx context.Context, userID int64, inputs []model.TransactionInput) ([]model.Transaction, error) {
	for _, in := range inputs {
		m.synced = append(m.synced, model.Transaction{
			ClientTransactionID: in.ClientTransactionID,
			Amount:              in.Amount,
			Note:                in.Note,
			IsSynced:            true,
		})
	}
	return m.synced, nil
}

func TestSyncHandler_SyncTransactions(t *testing.T) {
	repo := &mockSyncRepo{}
	h := handler.NewSyncHandler(repo)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"transactions":[{"client_transaction_id":"abc","amount":10000}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync/transactions", strings.NewReader(body))
	req = req.WithContext(middleware.ContextWithUser(req.Context(), &model.User{ID: 1}))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
