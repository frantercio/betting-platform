package auth

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey string

const userKey ctxKey = "user_id"

// Middleware validates Bearer tokens and injects user_id into the context.
func (m *Manager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/healthz") {
			next.ServeHTTP(w, r)
			return
		}
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			http.Error(w, `{"code":"UNAUTHORIZED","message":"missing bearer token"}`, http.StatusUnauthorized)
			return
		}
		claims, err := m.Verify(strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			http.Error(w, `{"code":"UNAUTHORIZED","message":"invalid token"}`, http.StatusUnauthorized)
			return
		}
		sub, _ := claims["sub"].(string)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, sub)))
	})
}

// UserID reads the authenticated user from the request context.
func UserID(r *http.Request) string {
	if v, ok := r.Context().Value(userKey).(string); ok {
		return v
	}
	return ""
}

// WithUser injects a user for tests / internal calls.
func WithUser(r *http.Request, userID string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), userKey, userID))
}