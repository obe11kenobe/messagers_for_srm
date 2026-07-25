package messager

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userIDKey contextKey = "user_id"

func RequireAuth(secret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		token := strings.TrimPrefix(header, "Bearer ")
		if token == header { // префикса "Bearer " не было
			http.Error(w, "Нужен заголовок Authorization: Bearer <token>", http.StatusUnauthorized)
			return
		}

		userID, err := ParseUserID(token, secret)
		if err != nil {
			http.Error(w, "Некорректный токен", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}

// UserIDFromContext достаёт user_id, положенный туда RequireAuth.
func UserIDFromContext(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(userIDKey).(int)
	return userID, ok
}
