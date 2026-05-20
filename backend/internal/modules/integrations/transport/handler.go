package transport

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"reserveflow/backend/internal/infrastructure/middleware"
	integrationsapp "reserveflow/backend/internal/modules/integrations/application"
	seatsdomain "reserveflow/backend/internal/modules/seats/domain"
)

type Handler struct {
	service *integrationsapp.Service
}

func NewHandler(service *integrationsapp.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/sessions/{sessionId}/layout", h.getSessionLayout)
	r.Put("/sessions/{sessionId}/layout", h.putSessionLayout)
	r.Delete("/sessions/{sessionId}/layout", h.deleteSessionLayout)
	r.Get("/halls/{hallId}/layout", h.getHallLayout)
	r.Put("/halls/{hallId}/layout", h.putHallLayout)
}

type layoutRequest struct {
	Layout seatsdomain.StoredSeatLayout `json:"layout"`
}

func (h *Handler) getSessionLayout(w http.ResponseWriter, r *http.Request) {
	state, err := h.service.GetSessionLayoutState(r.Context(), chi.URLParam(r, "sessionId"))
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, state)
}

func (h *Handler) putSessionLayout(w http.ResponseWriter, r *http.Request) {
	var req layoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		middleware.Error(w, err)
		return
	}
	result, err := h.service.UpsertSessionLayout(r.Context(), chi.URLParam(r, "sessionId"), req.Layout)
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, result)
}

func (h *Handler) deleteSessionLayout(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.DeleteSessionLayout(r.Context(), chi.URLParam(r, "sessionId"))
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, result)
}

func (h *Handler) getHallLayout(w http.ResponseWriter, r *http.Request) {
	state, err := h.service.GetHallLayoutState(r.Context(), chi.URLParam(r, "hallId"))
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, state)
}

func (h *Handler) putHallLayout(w http.ResponseWriter, r *http.Request) {
	var req layoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		middleware.Error(w, err)
		return
	}
	result, err := h.service.UpsertHallLayout(r.Context(), chi.URLParam(r, "hallId"), req.Layout)
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, result)
}
