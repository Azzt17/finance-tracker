package handler_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/Azzt17/finance-tracker/internal/database"
	"github.com/Azzt17/finance-tracker/internal/handler"
	"github.com/Azzt17/finance-tracker/internal/repository"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := database.Open(context.Background(), "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	return db
}

func TestAuthHandler_Login(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	_, _ = userRepo.Create(context.Background(), "alice", string(hashed))

	h := handler.NewAuthHandler(userRepo, sessionRepo, "", "", nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Valid login
	body, _ := json.Marshal(handler.LoginRequest{Username: "alice", Password: "password123"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}
	cookies := w.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != "session_id" {
		t.Errorf("expected session_id cookie")
	}

	// Invalid password
	body, _ = json.Marshal(handler.LoginRequest{Username: "alice", Password: "wrong"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	h := handler.NewAuthHandler(userRepo, sessionRepo, "", "", nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}
	cookies := w.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Value != "" {
		t.Errorf("expected empty session_id cookie")
	}
}

func TestAuthHandler_Register(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	h := handler.NewAuthHandler(userRepo, sessionRepo, "", "", nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Valid registration
	body, _ := json.Marshal(handler.LoginRequest{Username: "bob", Password: "password123"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}
	cookies := w.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != "session_id" {
		t.Errorf("expected session_id cookie")
	}

	// Invalid registration (short pass)
	body, _ = json.Marshal(handler.LoginRequest{Username: "bob2", Password: "12"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", w.Code)
	}

	// Duplicate username
	body, _ = json.Marshal(handler.LoginRequest{Username: "bob", Password: "password123"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for duplicate, got %d", w.Code)
	}
}

func TestAuthHandler_Google(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	h := handler.NewAuthHandler(userRepo, sessionRepo, "", "", nil) // no config
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Google login (not implemented)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected 501 Not Implemented, got %d", w.Code)
	}

	// Google callback (not implemented)
	req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/google/callback", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected 501 Not Implemented, got %d", w.Code)
	}
}
