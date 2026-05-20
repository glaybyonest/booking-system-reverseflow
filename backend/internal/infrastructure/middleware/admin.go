package middleware

import (
	"net/http"

	apperrors "reserveflow/backend/internal/infrastructure/errors"
)

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if UserRoleFromContext(r.Context()) != "admin" {
			Error(w, apperrors.Forbidden("Admin access required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
