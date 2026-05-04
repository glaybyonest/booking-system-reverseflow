package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"reserveflow/backend/internal/infrastructure/middleware"
	"reserveflow/backend/internal/modules/sessions/application"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/sessions/{sessionId}", h.get)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	session, err := h.service.GetSession(r.Context(), chi.URLParam(r, "sessionId"))
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, session)
}
