package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"

	apperrors "reserveflow/backend/internal/infrastructure/errors"
)

type RateLimiter interface {
	Incr(ctx context.Context, key string, ttl time.Duration) (int64, error)
}

func RateLimit(limiter RateLimiter, limit int64, window time.Duration, log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if limiter == nil {
				next.ServeHTTP(w, r)
				return
			}
			key := rateLimitKey(r)
			count, err := limiter.Incr(r.Context(), key, window)
			if err != nil {
				log.Warn().Err(err).Str("key", key).Msg("rate limit check failed")
				next.ServeHTTP(w, r)
				return
			}
			if count > limit {
				Error(w, apperrors.New("RATE_LIMITED", "Too many requests", http.StatusTooManyRequests))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func rateLimitKey(r *http.Request) string {
	if userID := UserIDFromContext(r.Context()); userID != "" {
		return "ratelimit:user:" + userID
	}
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		ip = strings.TrimSpace(strings.Split(ip, ",")[0])
	} else {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			ip = host
		} else {
			ip = r.RemoteAddr
		}
	}
	return "ratelimit:ip:" + ip
}
