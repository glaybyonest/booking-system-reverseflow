package transport

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"reserveflow/backend/internal/infrastructure/middleware"
	"reserveflow/backend/internal/modules/bookings/application"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.With(authMiddleware).Post("/bookings/hold", h.hold)
	r.With(authMiddleware).Get("/bookings/me", h.mine)
	r.With(authMiddleware).Get("/bookings/{bookingId}", h.get)
	r.With(authMiddleware).Post("/bookings/{bookingId}/cancel", h.cancel)
}

type holdRequest struct {
	SessionID string `json:"sessionId"`
	SeatID    string `json:"seatId"`
}

func (h *Handler) hold(w http.ResponseWriter, r *http.Request) {
	var req holdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.Error(w, err)
		return
	}
	result, err := h.service.HoldSeat(r.Context(), middleware.UserIDFromContext(r.Context()), req.SessionID, req.SeatID)
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, result)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	booking, err := h.service.GetBooking(r.Context(), middleware.UserIDFromContext(r.Context()), chi.URLParam(r, "bookingId"))
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, booking)
}

func (h *Handler) mine(w http.ResponseWriter, r *http.Request) {
	bookings, err := h.service.GetMyBookings(r.Context(), middleware.UserIDFromContext(r.Context()))
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"items": bookings})
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	if err := h.service.CancelBooking(r.Context(), middleware.UserIDFromContext(r.Context()), chi.URLParam(r, "bookingId")); err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}
