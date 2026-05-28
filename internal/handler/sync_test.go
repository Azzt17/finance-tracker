package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Azzt17/finance-tracker/internal/handler"
	"github.com/Azzt17/finance-tracker/internal/model"
)

type mockSyncRepo struct{}

func (m *mockSyncRepo) SyncTransactions(ctx context.Context, input []model.TransactionInput) ([]model.Transaction, error) {
	return []model.Transaction{}, nil
}

func TestSyncHandler_SyncTransactions(t *testing.T) {
	repo := &mockSyncRepo{}
	h := handler.NewSyncHandler(repo)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"transactions":[{"client_transaction_id":"abc","amount":10000}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync/transactions", strings.NewReader(body))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
