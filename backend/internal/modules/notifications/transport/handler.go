package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"reserveflow/backend/internal/infrastructure/middleware"
	"reserveflow/backend/internal/modules/notifications/application"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.With(authMiddleware).Get("/notifications", h.list)
	r.With(authMiddleware).Post("/notifications/{id}/read", h.markRead)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	notifications, err := h.service.List(r.Context(), middleware.UserIDFromContext(r.Context()))
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"items": notifications})
}

func (h *Handler) markRead(w http.ResponseWriter, r *http.Request) {
	if err := h.service.MarkRead(r.Context(), middleware.UserIDFromContext(r.Context()), chi.URLParam(r, "id")); err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"status": "read"})
}
