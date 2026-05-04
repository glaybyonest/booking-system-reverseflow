package middleware

import (
	"net/http"
	"strings"

	infraauth "reserveflow/backend/internal/infrastructure/auth"
	apperrors "reserveflow/backend/internal/infrastructure/errors"
)

func Auth(jwt *infraauth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				Error(w, apperrors.Unauthorized("Missing bearer token"))
				return
			}
			claims, err := jwt.ParseAccessToken(strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				Error(w, apperrors.Unauthorized("Invalid bearer token"))
				return
			}
			next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), claims.UserID, claims.Role)))
		})
	}
}
