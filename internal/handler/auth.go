package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/Azzt17/finance-tracker/internal/middleware"
	"github.com/Azzt17/finance-tracker/internal/model"
	"github.com/Azzt17/finance-tracker/internal/repository"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	envUsername string
	envPassword string
	oauthConf   *oauth2.Config
}

func NewAuthHandler(userRepo repository.UserRepository, sessionRepo repository.SessionRepository, envUsername, envPassword string, oauthConf *oauth2.Config) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		envUsername: envUsername,
		envPassword: envPassword,
		oauthConf:   oauthConf,
	}
}

func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.HandleFunc("POST /api/v1/auth/register", h.Register)
	mux.HandleFunc("POST /api/v1/auth/logout", h.Logout)
	mux.HandleFunc("GET /api/v1/auth/me", h.Me)
	mux.HandleFunc("GET /api/v1/auth/google", h.GoogleLogin)
	mux.HandleFunc("GET /api/v1/auth/google/callback", h.GoogleCallback)
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	user, err := h.userRepo.GetByUsername(r.Context(), req.Username)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid username or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		// Fallback for initial admin setup using .env credentials
		if req.Username == h.envUsername && req.Password == h.envPassword && h.envUsername != "" && h.envPassword != "" {
			// Update the database with the new bcrypt hash
			hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err == nil {
				_ = h.userRepo.UpdatePassword(r.Context(), user.ID, string(hash))
			}
		} else {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid username or password"})
			return
		}
	}

	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(24 * 7 * time.Hour) // 7 days

	session := &model.Session{
		ID:        sessionID,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	}

	if err := h.sessionRepo.Create(r.Context(), session); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create session"})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if len(req.Username) < 3 || len(req.Password) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username must be at least 3 chars and password at least 6 chars"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
		return
	}

	user, err := h.userRepo.Create(r.Context(), req.Username, string(hash))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username already taken"})
		return
	}

	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(24 * 7 * time.Hour) // 7 days

	session := &model.Session{
		ID:        sessionID,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	}

	if err := h.sessionRepo.Create(r.Context(), session); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create session"})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil && cookie.Value != "" {
		_ = h.sessionRepo.Delete(r.Context(), cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		// If authenticated via BasicAuth, UserFromContext is nil, but request is authorized.
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "authenticated_via_basic_auth",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":         user.ID,
		"username":   user.Username,
		"created_at": user.CreatedAt,
	})
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if h.oauthConf == nil {
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "Google OAuth is not configured"})
		return
	}
	url := h.oauthConf.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if h.oauthConf == nil {
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "Google OAuth is not configured"})
		return
	}

	state := r.FormValue("state")
	if state != "state-token" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid oauth state"})
		return
	}

	code := r.FormValue("code")
	token, err := h.oauthConf.Exchange(r.Context(), code)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to exchange token"})
		return
	}

	client := h.oauthConf.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to parse user info"})
		return
	}

	user, err := h.userRepo.GetByGoogleID(r.Context(), userInfo.ID)
	if err != nil && err != repository.ErrNotFound {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	if user == nil {
		// Create new OAuth user
		// Note: we use email as username or prefix it if needed.
		// For simplicity, let's use the email as username. If username is taken, it will fail.
		user, err = h.userRepo.CreateOAuthUser(r.Context(), userInfo.Email, userInfo.Email, userInfo.ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
			return
		}
	}

	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(24 * 7 * time.Hour) // 7 days

	session := &model.Session{
		ID:        sessionID,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	}

	if err := h.sessionRepo.Create(r.Context(), session); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create session"})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
