package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/Hiroki111/go-backend-example/internal/auth"
)

type AuthContextKey string

const (
	UserIDKey AuthContextKey = "userID"
	RoleKey   AuthContextKey = "role"
)

func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userId, role, err := auth.ParseJWTToken(parts[1])
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userId)
		ctx = context.WithValue(ctx, RoleKey, role)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (h *Handler) RequireRole(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return h.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(RoleKey).(string)
		if !ok {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if role != requiredRole {
			http.Error(w, "invalid role", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
