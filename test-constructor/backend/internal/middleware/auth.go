package middleware

import (
	"context"
	"net/http"
	"strings"
	"test-constructor/internal/auth"
)

type contextKey string

const UserContextKey contextKey = "user"

func AuthMiddleware(jwtService auth.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
				return
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				http.Error(w, "Неверный формат токена", http.StatusUnauthorized)
				return
			}

			claims, err := jwtService.ValidateToken(tokenParts[1])
			if err != nil {
				http.Error(w, "Недействительный токен", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
