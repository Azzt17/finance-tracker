package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Azzt17/finance-tracker/internal/database"
	"github.com/Azzt17/finance-tracker/internal/handler"
	"github.com/Azzt17/finance-tracker/internal/service"
	"github.com/Azzt17/finance-tracker/web"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var version = "dev"

func main() {
	addr := env("ADDR", ":8080")
	databaseURL := databaseURL()

	ctx := context.Background()
	db, err := database.Open(ctx, databaseURL)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("database close failed", "error", err)
		}
	}()

	if err := database.Migrate(ctx, db); err != nil {
		slog.Error("database migration failed", "error", err)
		os.Exit(1)
	}
	slog.Info("database migrated")

	if err := database.Seed(ctx, db); err != nil {
		slog.Error("database seeding failed", "error", err)
		os.Exit(1)
	}
	slog.Info("database seeded")

	// Start automated backup cron
	service.StartBackupCron(db, databaseURL)
	slog.Info("automated backup service started")

	var oauthConf *oauth2.Config
	if clientID := os.Getenv("GOOGLE_CLIENT_ID"); clientID != "" {
		oauthConf = &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_CALLBACK_URL"),
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		}
	}

	server := &http.Server{
		Addr: addr,
		Handler: handler.NewRouter(handler.Config{
			StaticFS:     web.StaticFS,
			DB:           db,
			AuthUsername: os.Getenv("APP_USERNAME"),
			AuthPassword: os.Getenv("APP_PASSWORD"),
			Version:      version,
			DatabaseURL:  databaseURL,
			OAuthConfig:  oauthConf,
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("starting http server", "addr", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-shutdownCtx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("http server shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("http server stopped")
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func databaseURL() string {
	if value := os.Getenv("DATABASE_URL"); value != "" {
		return value
	}
	if os.Getenv("FLY_APP_NAME") != "" {
		return "file:/data/finance-tracker.db?_foreign_keys=on&_busy_timeout=5000"
	}

	return "file:finance-tracker.db?_foreign_keys=on&_busy_timeout=5000"
}
