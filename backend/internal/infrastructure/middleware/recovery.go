package middleware

import (
	"net/http"

	"github.com/rs/zerolog"

	apperrors "reserveflow/backend/internal/infrastructure/errors"
)

func Recovery(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error().
						Str("request_id", RequestIDFromContext(r.Context())).
						Interface("panic", rec).
						Msg("request panic recovered")
					Error(w, apperrors.New(apperrors.CodeInternalError, "Internal server error", http.StatusInternalServerError))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
