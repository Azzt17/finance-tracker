package handler

import (
	"database/sql"
	"io/fs"
	"net/http"

	"github.com/Azzt17/finance-tracker/internal/middleware"
	"github.com/Azzt17/finance-tracker/internal/repository"
)

type Config struct {
	StaticFS          fs.FS
	DB                *sql.DB
	AuthUsername      string
	AuthPassword      string
}

func NewRouter(config Config) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthCheck)

	transactions := NewTransactionHandler(repository.NewTransactionRepository(config.DB))
	transactions.RegisterRoutes(mux)

	budgets := NewBudgetHandler(repository.NewBudgetRepository(config.DB))
	budgets.RegisterRoutes(mux)

	categories := NewCategoryHandler(repository.NewCategoryRepository(config.DB))
	categories.RegisterRoutes(mux)

	savingsGoals := NewSavingsGoalHandler(repository.NewSavingsRepository(config.DB))
	savingsGoals.RegisterRoutes(mux)

	exports := NewExportHandler(repository.NewExportRepository(config.DB))
	exports.RegisterRoutes(mux)

	sync := NewSyncHandler(repository.NewSyncRepository(config.DB))
	sync.RegisterRoutes(mux)

	if config.StaticFS != nil {
		mux.Handle("GET /", http.FileServer(http.FS(config.StaticFS)))
	}

	var app http.Handler = mux
	if config.AuthUsername != "" || config.AuthPassword != "" {
		app = middleware.BasicAuth(
			app,
			config.AuthUsername,
			config.AuthPassword,
			"/healthz",
			"/manifest.json",
			"/sw.js",
			"/static/icons/*",
		)
	}

	return middleware.Recoverer(middleware.Logger(app))
}

func healthCheck(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
