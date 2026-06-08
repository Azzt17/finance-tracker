package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Azzt17/finance-tracker/internal/middleware"
	"github.com/Azzt17/finance-tracker/internal/repository"
)

type AnalyticsHandler struct {
	repo *repository.AnalyticsRepository
}

func NewAnalyticsHandler(repo *repository.AnalyticsRepository) *AnalyticsHandler {
	return &AnalyticsHandler{repo: repo}
}

func (h *AnalyticsHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/analytics/spending-by-category", h.GetSpendingByCategory)
	mux.HandleFunc("GET /api/v1/analytics/monthly-trend", h.GetMonthlyTrend)
	mux.HandleFunc("GET /api/v1/analytics/daily-spending", h.GetDailySpending)
}

func (h *AnalyticsHandler) GetSpendingByCategory(w http.ResponseWriter, r *http.Request) {
	yearMonth := r.URL.Query().Get("year_month")
	if yearMonth == "" || len(strings.Split(yearMonth, "-")) != 2 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "year_month is required in YYYY-MM format"})
		return
	}

	userID := middleware.UserFromContext(r.Context()).ID
	data, err := h.repo.GetSpendingByCategory(r.Context(), userID, yearMonth)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, data)
}

func (h *AnalyticsHandler) GetMonthlyTrend(w http.ResponseWriter, r *http.Request) {
	monthsStr := r.URL.Query().Get("months")
	months := 6
	if monthsStr != "" {
		if m, err := strconv.Atoi(monthsStr); err == nil {
			months = m
		}
	}

	userID := middleware.UserFromContext(r.Context()).ID
	data, err := h.repo.GetMonthlyTrend(r.Context(), userID, months)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, data)
}

func (h *AnalyticsHandler) GetDailySpending(w http.ResponseWriter, r *http.Request) {
	yearMonth := r.URL.Query().Get("year_month")
	if yearMonth == "" || len(strings.Split(yearMonth, "-")) != 2 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "year_month is required in YYYY-MM format"})
		return
	}

	userID := middleware.UserFromContext(r.Context()).ID
	data, err := h.repo.GetDailySpending(r.Context(), userID, yearMonth)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, data)
}
