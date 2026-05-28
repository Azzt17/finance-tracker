package handler

import (
	"net/http"

	"github.com/Azzt17/finance-tracker/internal/middleware"
)

type Config struct {
	StaticDir string
}

func NewRouter(config Config) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthCheck)

	transactions := NewTransactionHandler()
	transactions.RegisterRoutes(mux)

	budgets := NewBudgetHandler()
	budgets.RegisterRoutes(mux)

	categories := NewCategoryHandler()
	categories.RegisterRoutes(mux)

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
