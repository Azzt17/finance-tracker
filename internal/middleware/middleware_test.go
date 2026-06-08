package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Azzt17/finance-tracker/internal/middleware"
	"github.com/Azzt17/finance-tracker/internal/model"
)

func TestBasicAuth(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	mw := middleware.BasicAuth(handler, "user", "pass", "/healthz", "/static/icons/*")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("user", "pass")
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected exempt path status %d, got %d", http.StatusNoContent, w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/static/icons/icon-192x192.png", nil)
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected exempt prefix status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestLogger(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := middleware.Logger(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	mw.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRecoverer(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	mw := middleware.Recoverer(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	mw.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestRequireRole(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := middleware.RequireRole("admin")(handler)

	// Test 1: No user in context
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	w1 := httptest.NewRecorder()
	mw.ServeHTTP(w1, req1)
	if w1.Code != http.StatusForbidden {
		t.Errorf("expected forbidden for no user, got %d", w1.Code)
	}

	// Test 2: User in context but wrong role
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	user2 := &model.User{ID: 1, Username: "test", Role: "user"}
	req2 = req2.WithContext(middleware.ContextWithUser(req2.Context(), user2))
	w2 := httptest.NewRecorder()
	mw.ServeHTTP(w2, req2)
	if w2.Code != http.StatusForbidden {
		t.Errorf("expected forbidden for wrong role, got %d", w2.Code)
	}

	// Test 3: User in context with correct role
	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	user3 := &model.User{ID: 2, Username: "admin", Role: "admin"}
	req3 = req3.WithContext(middleware.ContextWithUser(req3.Context(), user3))
	w3 := httptest.NewRecorder()
	mw.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Errorf("expected OK for correct role, got %d", w3.Code)
	}
}
