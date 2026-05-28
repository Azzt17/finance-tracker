package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Azzt17/finance-tracker/internal/handler"
	"github.com/Azzt17/finance-tracker/internal/model"
)

type mockTransactionRepo struct {
	txs []model.Transaction
}

func (m *mockTransactionRepo) List(ctx context.Context, yearMonth string, categoryID *int64, limit, offset int) ([]model.Transaction, error) {
	return m.txs, nil
}

func (m *mockTransactionRepo) Create(ctx context.Context, input model.TransactionInput) (model.Transaction, error) {
	t := model.Transaction{
		ID:                  1,
		ClientTransactionID: input.ClientTransactionID,
		Amount:              input.Amount,
		Note:                input.Note,
		TransactedAt:        input.TransactedAt,
		CreatedAt:           time.Now(),
	}
	m.txs = append(m.txs, t)
	return t, nil
}

func (m *mockTransactionRepo) Update(ctx context.Context, id int64, input model.TransactionInput) (model.Transaction, error) {
	if len(m.txs) > 0 {
		m.txs[0].Amount = input.Amount
		return m.txs[0], nil
	}
	return model.Transaction{}, nil
}

func (m *mockTransactionRepo) Delete(ctx context.Context, id int64) error {
	m.txs = []model.Transaction{}
	return nil
}

func (m *mockTransactionRepo) Get(ctx context.Context, id int64) (model.Transaction, error) {
	if len(m.txs) > 0 {
		return m.txs[0], nil
	}
	return model.Transaction{}, nil
}

func TestTransactionHandler_List(t *testing.T) {
	repo := &mockTransactionRepo{
		txs: []model.Transaction{
			{ID: 1, Amount: 50000, Note: "Test"},
		},
	}
	h := handler.NewTransactionHandler(repo)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions?year_month=2026-05", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestTransactionHandler_Create(t *testing.T) {
	repo := &mockTransactionRepo{}
	h := handler.NewTransactionHandler(repo)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"client_transaction_id":"abc-123","amount":50000,"note":"Makan"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", strings.NewReader(body))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestTransactionHandler_Update(t *testing.T) {
	repo := &mockTransactionRepo{
		txs: []model.Transaction{{ID: 1, Amount: 50000}},
	}
	h := handler.NewTransactionHandler(repo)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"client_transaction_id":"abc-123","amount":60000}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/transactions/1", strings.NewReader(body))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestTransactionHandler_Delete(t *testing.T) {
	repo := &mockTransactionRepo{
		txs: []model.Transaction{{ID: 1, Amount: 50000}},
	}
	h := handler.NewTransactionHandler(repo)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/transactions/1", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
