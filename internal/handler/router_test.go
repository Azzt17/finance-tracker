package handler_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/Azzt17/finance-tracker/internal/handler"
	_ "github.com/mattn/go-sqlite3"
)

func TestNewRouter(t *testing.T) {
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("hello")},
	}

	// Create a mock db
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer func() { _ = db.Close() }()

	cfg := handler.Config{
		StaticFS: mockFS,
		DB:       db,
	}

	r := handler.NewRouter(cfg)

	// Test healthcheck
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
