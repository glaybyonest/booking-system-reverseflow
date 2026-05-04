package transport

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"reserveflow/backend/internal/infrastructure/middleware"
	"reserveflow/backend/internal/modules/payments/application"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.With(authMiddleware).Post("/payments", h.create)
	r.With(authMiddleware).Get("/payments/{paymentId}", h.get)
}

type createPaymentRequest struct {
	BookingID      string `json:"bookingId"`
	IdempotencyKey string `json:"idempotencyKey"`
	ForceStatus    string `json:"forceStatus"`
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req createPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.Error(w, err)
		return
	}
	if req.ForceStatus == "" {
		req.ForceStatus = r.URL.Query().Get("forceStatus")
	}
	payment, err := h.service.Process(r.Context(), middleware.UserIDFromContext(r.Context()), req.BookingID, req.IdempotencyKey, req.ForceStatus)
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, payment)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	payment, err := h.service.GetPayment(r.Context(), middleware.UserIDFromContext(r.Context()), chi.URLParam(r, "paymentId"))
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, payment)
}
