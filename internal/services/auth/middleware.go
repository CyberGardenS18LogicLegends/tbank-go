package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"tbank-go/internal/utils"
)

func AuthMiddleware(jwtSecret string, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
				return
			}

			// Extract the JWT token from the header
			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Parse the JWT token to extract the UID
			uid, err := utils.ParseJWT(token, jwtSecret)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Log UID for debugging
			log.Info("user authenticated", slog.String("userUID", uid))

			// Attach the UID to the request context
			ctx := context.WithValue(r.Context(), "userUID", uid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
