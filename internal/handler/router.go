package handler

import (
	"database/sql"
	"io/fs"
	"net/http"

	"github.com/Azzt17/finance-tracker/internal/middleware"
	"github.com/Azzt17/finance-tracker/internal/model"
	"github.com/Azzt17/finance-tracker/internal/repository"
)

type Config struct {
	StaticFS     fs.FS
	DB           *sql.DB
	AuthUsername string
	AuthPassword string
	Version      string
	DatabaseURL  string
}

func NewRouter(config Config) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"version": config.Version,
		})
	})

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

	analytics := NewAnalyticsHandler(repository.NewAnalyticsRepository(config.DB))
	analytics.RegisterRoutes(mux)

	auth := NewAuthHandler(repository.NewUserRepository(config.DB), repository.NewSessionRepository(config.DB), config.AuthUsername, config.AuthPassword)
	auth.RegisterRoutes(mux)

	system := &SystemHandler{Config: config}

	// Apply RequireRole middleware for admin endpoints
	requireAdmin := middleware.RequireRole(model.RoleAdmin)

	mux.Handle("GET /api/v1/system/backup", requireAdmin(http.HandlerFunc(system.DownloadBackup)))

	// Internal endpoint (protected by localhost check)
	mux.Handle("POST /internal/backup", requireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		system.TriggerBackup(w, r)
	})))

	if config.StaticFS != nil {
		mux.Handle("GET /", http.FileServer(http.FS(config.StaticFS)))
	}

	var app http.Handler = mux
	app = middleware.Auth(
		repository.NewSessionRepository(config.DB),
		repository.NewUserRepository(config.DB),
		config.AuthUsername,
		config.AuthPassword,
		"/",
		"/index.html",
		"/healthz",
		"/manifest.json",
		"/sw.js",
		"/static/icons/*",
		"/static/js/*",
		"/static/css/*",
		"/internal/*",
		"/api/v1/auth/login",
	)(app)

	return middleware.Recoverer(middleware.Logger(app))
}
