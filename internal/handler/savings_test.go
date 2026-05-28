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

type mockSavingsRepo struct {
	goals []model.SavingsGoal
}

func (m *mockSavingsRepo) List(ctx context.Context, yearMonth string) ([]model.SavingsGoal, error) {
	return m.goals, nil
}

func (m *mockSavingsRepo) Create(ctx context.Context, input model.SavingsGoalInput) (model.SavingsGoal, error) {
	g := model.SavingsGoal{
		ID:           1,
		Name:         input.Name,
		TargetAmount: input.TargetAmount,
		CurrentSaved: input.CurrentSaved,
		YearMonth:    input.YearMonth,
		IsAchieved:   input.IsAchieved,
		CreatedAt:    time.Now(),
	}
	m.goals = append(m.goals, g)
	return g, nil
}

func (m *mockSavingsRepo) Update(ctx context.Context, id int64, input model.SavingsGoalInput) (model.SavingsGoal, error) {
	if len(m.goals) > 0 {
		m.goals[0].CurrentSaved = input.CurrentSaved
		return m.goals[0], nil
	}
	return model.SavingsGoal{}, nil
}

func (m *mockSavingsRepo) Delete(ctx context.Context, id int64) error {
	m.goals = []model.SavingsGoal{}
	return nil
}

func TestSavingsHandler_List(t *testing.T) {
	repo := &mockSavingsRepo{
		goals: []model.SavingsGoal{
			{ID: 1, Name: "Beli HP"},
		},
	}
	h := handler.NewSavingsGoalHandler(repo)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/savings?year_month=2026-05", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestSavingsHandler_Create(t *testing.T) {
	repo := &mockSavingsRepo{}
	h := handler.NewSavingsGoalHandler(repo)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"name":"Liburan","target_amount":5000000,"year_month":"2026-05"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/savings", strings.NewReader(body))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestSavingsHandler_Update(t *testing.T) {
	repo := &mockSavingsRepo{
		goals: []model.SavingsGoal{{ID: 1, Name: "Beli HP", CurrentSaved: 0}},
	}
	h := handler.NewSavingsGoalHandler(repo)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"name":"Beli HP","target_amount":5000000,"current_saved":1000000}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/savings/1", strings.NewReader(body))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestSavingsHandler_Delete(t *testing.T) {
	repo := &mockSavingsRepo{
		goals: []model.SavingsGoal{{ID: 1, Name: "Beli HP"}},
	}
	h := handler.NewSavingsGoalHandler(repo)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/savings/1", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
