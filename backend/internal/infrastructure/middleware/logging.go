package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type loggingRecorder struct {
	http.ResponseWriter
	status int
}

func (r *loggingRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func Logging(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &loggingRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			log.Info().
				Str("request_id", RequestIDFromContext(r.Context())).
				Str("user_id", UserIDFromContext(r.Context())).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", rec.status).
				Dur("latency", time.Since(start)).
				Msg("http request")
		})
	}
}
