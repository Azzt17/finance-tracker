package middleware

import (
	"context"
	"crypto/subtle"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Azzt17/finance-tracker/internal/model"
	"github.com/Azzt17/finance-tracker/internal/repository"
)

type contextKey string

const userContextKey contextKey = "user"

// UserFromContext retrieves the authenticated user from the request context.
// It may return nil if the request was authenticated via legacy Basic Auth fallback.
func UserFromContext(ctx context.Context) *model.User {
	u, _ := ctx.Value(userContextKey).(*model.User)
	return u
}

func matchPathAuth(pattern, path string) bool {
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(path, prefix+"/")
	}
	matched, _ := filepath.Match(pattern, path)
	return matched
}

// Auth is the new authentication middleware that supports both Session Cookies
// and legacy Basic Auth as a fallback.
func Auth(
	sessionRepo repository.SessionRepository,
	userRepo repository.UserRepository,
	basicUsername string,
	basicPassword string,
	excludedPaths ...string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, path := range excludedPaths {
				if matchPathAuth(path, r.URL.Path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			var user *model.User

			// 1. Try Cookie Auth
			cookie, err := r.Cookie("session_id")
			if err == nil && cookie.Value != "" {
				session, err := sessionRepo.GetByID(r.Context(), cookie.Value)
				if err == nil && session.ExpiresAt.After(time.Now()) {
					u, err := userRepo.GetByID(r.Context(), session.UserID)
					if err == nil {
						user = u
					}
				}
			}

			// If cookie auth successful
			if user != nil {
				ctx := context.WithValue(r.Context(), userContextKey, user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// 2. Try Basic Auth (Fallback)
			if basicUsername != "" && basicPassword != "" {
				username, pass, ok := r.BasicAuth()
				if ok && subtle.ConstantTimeCompare([]byte(username), []byte(basicUsername)) == 1 && subtle.ConstantTimeCompare([]byte(pass), []byte(basicPassword)) == 1 {
					// Legacy fallback sets user to the default admin (ID 1)
					u, err := userRepo.GetByID(r.Context(), 1)
					if err == nil {
						user = u
					} else {
						// Create an empty user with ID 1 so context won't be nil
						user = &model.User{ID: 1, Username: "admin"}
					}
					ctx := context.WithValue(r.Context(), userContextKey, user)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			w.WriteHeader(http.StatusUnauthorized)
		})
	}
}
