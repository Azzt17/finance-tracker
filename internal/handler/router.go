package handler

import (
	"database/sql"
	"io/fs"
	"net/http"

	"github.com/Azzt17/finance-tracker/internal/middleware"
	"github.com/Azzt17/finance-tracker/internal/repository"
)

type Config struct {
	StaticFS fs.FS
	DB       *sql.DB
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

	// Stack middlewares: CORS -> Recoverer -> Logger
	return middleware.CORS(middleware.Recoverer(middleware.Logger(mux)))
}

func healthCheck(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
