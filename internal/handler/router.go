package handler

import (
	"database/sql"
	"net/http"

	"github.com/Azzt17/finance-tracker/internal/middleware"
)

type Config struct {
	StaticDir string
	DB        *sql.DB
}

func NewRouter(config Config) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthCheck)

	// transactions := NewTransactionHandler(repository.NewTransactionRepository(config.DB))
	// transactions.RegisterRoutes(mux)

	// budgets := NewBudgetHandler(repository.NewBudgetAllocationRepository(config.DB))
	// budgets.RegisterRoutes(mux)

	// categories := NewCategoryHandler(repository.NewCategoryRepository(config.DB))
	// categories.RegisterRoutes(mux)

	// savingsGoals := NewSavingsGoalHandler(repository.NewSavingsGoalRepository(config.DB))
	// savingsGoals.RegisterRoutes(mux)

	staticDir := config.StaticDir
	if staticDir == "" {
		staticDir = "web"
	}
	mux.Handle("GET /", http.FileServer(http.Dir(staticDir)))

	return middleware.Recoverer(middleware.Logger(mux))
}

func healthCheck(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, envelope{
		"status": "ok",
	})
}
