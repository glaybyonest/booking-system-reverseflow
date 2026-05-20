package transport

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	apperrors "reserveflow/backend/internal/infrastructure/errors"
	"reserveflow/backend/internal/infrastructure/middleware"
	"reserveflow/backend/internal/modules/events/application"
	"reserveflow/backend/internal/modules/events/domain"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/events", h.list)
	r.Get("/events/map", h.mapList)
	r.Get("/events/{eventId}", h.get)
	r.Get("/events/{eventId}/sessions", h.sessions)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	query, err := parseListQuery(r)
	if err != nil {
		middleware.Error(w, err)
		return
	}
	events, total, err := h.service.ListEvents(r.Context(), query)
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"items": events, "total": total})
}

func (h *Handler) mapList(w http.ResponseWriter, r *http.Request) {
	query, err := parseListQuery(r)
	if err != nil {
		middleware.Error(w, err)
		return
	}
	events, err := h.service.ListMapEvents(r.Context(), query)
	if err != nil {
		middleware.Error(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"events": events})
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

func parseListQuery(r *http.Request) (domain.ListQuery, error) {
	values := r.URL.Query()
	limit, err := parseInt(values.Get("limit"), 24)
	if err != nil {
		return domain.ListQuery{}, apperrors.Validation("limit must be an integer")
	}
	offset, err := parseInt(values.Get("offset"), 0)
	if err != nil {
		return domain.ListQuery{}, apperrors.Validation("offset must be an integer")
	}
	from, err := application.ParseDateRangeValue(values.Get("from"))
	if err != nil {
		return domain.ListQuery{}, err
	}
	to, err := application.ParseDateRangeValue(values.Get("to"))
	if err != nil {
		return domain.ListQuery{}, err
	}
	onlyActual := true
	if raw := values.Get("onlyActual"); raw != "" {
		parsed, parseErr := strconv.ParseBool(raw)
		if parseErr != nil {
			return domain.ListQuery{}, apperrors.Validation("onlyActual must be true or false")
		}
		onlyActual = parsed
	}

	return domain.ListQuery{
		City:        values.Get("city"),
		Source:      values.Get("source"),
		Category:    values.Get("category"),
		From:        from,
		To:          to,
		BookingMode: values.Get("bookingMode"),
		OnlyActual:  onlyActual,
		Limit:       limit,
		Offset:      offset,
	}, nil
}

func parseInt(raw string, fallback int) (int, error) {
	if raw == "" {
		return fallback, nil
	}
	return strconv.Atoi(raw)
}
