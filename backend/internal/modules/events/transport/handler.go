package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"reserveflow/backend/internal/infrastructure/middleware"
	"reserveflow/backend/internal/modules/events/application"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/events", h.list)
	r.Get("/events/{eventId}", h.get)
	r.Get("/events/{eventId}/sessions", h.sessions)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	events, err := h.service.ListEvents(r.Context())
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"items": events})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	event, err := h.service.GetEvent(r.Context(), chi.URLParam(r, "eventId"))
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, event)
}

func (h *Handler) sessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.service.GetEventSessions(r.Context(), chi.URLParam(r, "eventId"))
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"items": sessions})
}
