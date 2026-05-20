package application

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	apperrors "reserveflow/backend/internal/infrastructure/errors"
	seatsdomain "reserveflow/backend/internal/modules/seats/domain"
)

type LayoutScope string

const (
	LayoutScopeNone    LayoutScope = "none"
	LayoutScopeHall    LayoutScope = "hall"
	LayoutScopeSession LayoutScope = "session"
)

type LayoutVenueRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type LayoutHallRef struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Venue LayoutVenueRef `json:"venue"`
}

type SessionLayoutState struct {
	SessionID       string                        `json:"sessionId"`
	EventID         string                        `json:"eventId"`
	EventTitle      string                        `json:"eventTitle"`
	Source          string                        `json:"source"`
	BookingMode     string                        `json:"bookingMode"`
	IsBookable      bool                          `json:"isBookable"`
	Hall            *LayoutHallRef                `json:"hall,omitempty"`
	Layout          *seatsdomain.StoredSeatLayout `json:"layout,omitempty"`
	FallbackLayout  *seatsdomain.StoredSeatLayout `json:"fallbackLayout,omitempty"`
	EffectiveLayout *seatsdomain.StoredSeatLayout `json:"effectiveLayout,omitempty"`
	LayoutSource    LayoutScope                   `json:"layoutSource"`
}

type HallLayoutSessionSummary struct {
	ID         string     `json:"id"`
	EventID    string     `json:"eventId"`
	EventTitle string     `json:"eventTitle"`
	StartsAt   *time.Time `json:"startsAt,omitempty"`
	IsBookable bool       `json:"isBookable"`
}

type HallLayoutState struct {
	HallID   string                        `json:"hallId"`
	Name     string                        `json:"name"`
	Venue    LayoutVenueRef                `json:"venue"`
	Layout   *seatsdomain.StoredSeatLayout `json:"layout,omitempty"`
	Sessions []HallLayoutSessionSummary    `json:"sessions"`
}

type LayoutMutationResult struct {
	SessionID         *string `json:"sessionId,omitempty"`
	HallID            string  `json:"hallId"`
	EventID           string  `json:"eventId"`
	EventTitle        string  `json:"eventTitle"`
	BookingMode       string  `json:"bookingMode"`
	IsBookable        bool    `json:"isBookable"`
	VisibleInCatalog  bool    `json:"visibleInCatalog"`
	MaterializedSeats int     `json:"materializedSeats"`
}

func (s *Service) GetSessionLayoutState(ctx context.Context, sessionID string) (*SessionLayoutState, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, apperrors.Validation("sessionId is required")
	}
	state, err := s.repo.GetSessionLayoutState(ctx, sessionID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.New(apperrors.CodeNotFound, "Session not found", http.StatusNotFound)
		}
		return nil, apperrors.Internal(err)
	}
	return state, nil
}

func (s *Service) UpsertSessionLayout(ctx context.Context, sessionID string, layout seatsdomain.StoredSeatLayout) (*LayoutMutationResult, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, apperrors.Validation("sessionId is required")
	}
	if err := layout.Validate(); err != nil {
		return nil, apperrors.Validation(err.Error())
	}
	result, err := s.repo.UpsertSessionLayout(ctx, sessionID, layout, time.Now().UTC())
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.New(apperrors.CodeNotFound, "Session not found", http.StatusNotFound)
		}
		return nil, apperrors.Internal(err)
	}
	if result.SessionID != nil {
		s.invalidateSeatmaps(ctx, *result.SessionID)
	}
	return result, nil
}

func (s *Service) DeleteSessionLayout(ctx context.Context, sessionID string) (*LayoutMutationResult, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, apperrors.Validation("sessionId is required")
	}
	result, err := s.repo.DeleteSessionLayout(ctx, sessionID, time.Now().UTC())
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.New(apperrors.CodeNotFound, "Session not found", http.StatusNotFound)
		}
		return nil, apperrors.Internal(err)
	}
	if result.SessionID != nil {
		s.invalidateSeatmaps(ctx, *result.SessionID)
	}
	return result, nil
}

func (s *Service) GetHallLayoutState(ctx context.Context, hallID string) (*HallLayoutState, error) {
	if strings.TrimSpace(hallID) == "" {
		return nil, apperrors.Validation("hallId is required")
	}
	state, err := s.repo.GetHallLayoutState(ctx, hallID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.New(apperrors.CodeNotFound, "Hall not found", http.StatusNotFound)
		}
		return nil, apperrors.Internal(err)
	}
	return state, nil
}

func (s *Service) UpsertHallLayout(ctx context.Context, hallID string, layout seatsdomain.StoredSeatLayout) (*LayoutMutationResult, error) {
	if strings.TrimSpace(hallID) == "" {
		return nil, apperrors.Validation("hallId is required")
	}
	if err := layout.Validate(); err != nil {
		return nil, apperrors.Validation(err.Error())
	}
	result, err := s.repo.UpsertHallLayout(ctx, hallID, layout, time.Now().UTC())
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.New(apperrors.CodeNotFound, "Hall not found", http.StatusNotFound)
		}
		return nil, apperrors.Internal(err)
	}
	state, stateErr := s.repo.GetHallLayoutState(ctx, hallID)
	if stateErr != nil {
		if stateErr == pgx.ErrNoRows {
			return nil, apperrors.New(apperrors.CodeNotFound, "Hall not found", http.StatusNotFound)
		}
		return nil, apperrors.Internal(stateErr)
	}
	sessionIDs := make([]string, 0, len(state.Sessions))
	for _, item := range state.Sessions {
		sessionIDs = append(sessionIDs, item.ID)
	}
	s.invalidateSeatmaps(ctx, sessionIDs...)
	return result, nil
}
