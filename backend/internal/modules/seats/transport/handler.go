package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"reserveflow/backend/internal/infrastructure/middleware"
	"reserveflow/backend/internal/modules/seats/application"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/sessions/{sessionId}/seats", h.getSeatMap)
}

func (h *Handler) getSeatMap(w http.ResponseWriter, r *http.Request) {
	seatMap, err := h.service.GetSeatMap(r.Context(), chi.URLParam(r, "sessionId"))
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, seatMap)
}
