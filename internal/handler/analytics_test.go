package handler_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Azzt17/finance-tracker/internal/handler"
	"github.com/Azzt17/finance-tracker/internal/repository"
	_ "github.com/mattn/go-sqlite3"
)

func setupAnalyticsRouter() (http.Handler, *sql.DB) {
	db, _ := sql.Open("sqlite3", ":memory:")
	_, _ = db.Exec(`
		CREATE TABLE categories (id INTEGER PRIMARY KEY, name TEXT, icon_emoji TEXT);
		CREATE TABLE transactions (id INTEGER PRIMARY KEY, category_id INTEGER, amount INTEGER, transacted_at DATETIME);
		CREATE TABLE budget_allocation (id INTEGER PRIMARY KEY, year_month TEXT, total_budget INTEGER);
	`)
	repo := repository.NewAnalyticsRepository(db)
	analyticsHandler := handler.NewAnalyticsHandler(repo)

	mux := http.NewServeMux()
	analyticsHandler.RegisterRoutes(mux)

	return mux, db
}

func TestAnalyticsHandler_GetSpendingByCategory(t *testing.T) {
	r, db := setupAnalyticsRouter()
	defer func() { _ = db.Close() }()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/spending-by-category", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing year_month, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/analytics/spending-by-category?year_month=2026-05", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAnalyticsHandler_GetMonthlyTrend(t *testing.T) {
	r, db := setupAnalyticsRouter()
	defer func() { _ = db.Close() }()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/monthly-trend", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/analytics/monthly-trend?months=3", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAnalyticsHandler_GetDailySpending(t *testing.T) {
	r, db := setupAnalyticsRouter()
	defer func() { _ = db.Close() }()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/daily-spending", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing year_month, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/analytics/daily-spending?year_month=2026-05", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
