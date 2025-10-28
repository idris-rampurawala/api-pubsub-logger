package middleware

import (
	"net/http"

	"api-pubsub-logger/internal/utils"
)

// UserIDMiddleware extracts the user ID from headers and adds it to the context
func UserIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		ctx := utils.SetUserID(r.Context(), userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
