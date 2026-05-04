package transport

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"reserveflow/backend/internal/infrastructure/middleware"
	"reserveflow/backend/internal/modules/auth/application"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Post("/auth/register", h.register)
	r.Post("/auth/login", h.login)
	r.Post("/auth/refresh", h.refresh)
	r.Post("/auth/logout", h.logout)
	r.With(authMiddleware).Get("/auth/me", h.me)
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.Error(w, err)
		return
	}
	result, err := h.service.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, result)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.Error(w, err)
		return
	}
	result, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, result)
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.Error(w, err)
		return
	}
	tokens, err := h.service.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, tokens)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.Error(w, err)
		return
	}
	if err := h.service.Logout(r.Context(), req.RefreshToken); err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	user, err := h.service.Me(r.Context(), middleware.UserIDFromContext(r.Context()))
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, user)
}
