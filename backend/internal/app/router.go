package app

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"reserveflow/backend/internal/infrastructure/middleware"
	"reserveflow/backend/internal/infrastructure/observability"
)

func NewRouter(deps *Dependencies) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recovery(deps.Log))
	r.Use(middleware.RateLimit(deps.Redis, 300, time.Minute, deps.Log))
	r.Use(middleware.Logging(deps.Log))
	r.Use(observability.MetricsMiddleware)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		middleware.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := deps.DB.Ping(ctx); err != nil {
			middleware.JSON(w, http.StatusServiceUnavailable, map[string]string{"status": "db_unavailable"})
			return
		}
		if err := deps.Redis.Ping(ctx); err != nil {
			middleware.JSON(w, http.StatusServiceUnavailable, map[string]string{"status": "redis_unavailable"})
			return
		}
		middleware.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	r.Handle("/metrics", observability.Handler())

	authMiddleware := middleware.Auth(deps.JWT)
	r.Route("/api/v1", func(api chi.Router) {
		deps.AuthHandler.Routes(api, authMiddleware)
		deps.EventsHandler.Routes(api)
		deps.SessionsHandler.Routes(api)
		deps.SeatsHandler.Routes(api)
		deps.BookingsHandler.Routes(api, authMiddleware)
		deps.PaymentsHandler.Routes(api, authMiddleware)
		deps.NotificationsHandler.Routes(api, authMiddleware)
		api.Route("/admin", func(admin chi.Router) {
			admin.Use(authMiddleware)
			admin.Use(middleware.RequireAdmin)
			deps.IntegrationsHandler.Routes(admin)
		})
	})
	return r
}
