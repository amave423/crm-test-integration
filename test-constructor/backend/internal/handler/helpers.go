package handler

import (
	"encoding/json"
	"net/http"
	"test-constructor/internal/auth"
	"test-constructor/internal/dto"
)

type contextKey string

const UserContextKey contextKey = "user"

func GetUserFromContext(r *http.Request) (*auth.JWTClaims, bool) {
	claims, ok := r.Context().Value(UserContextKey).(*auth.JWTClaims)
	return claims, ok
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, dto.ErrorResponse{Error: message})
}
