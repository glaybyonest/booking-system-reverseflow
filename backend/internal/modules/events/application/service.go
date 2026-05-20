package application

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	apperrors "reserveflow/backend/internal/infrastructure/errors"
	"reserveflow/backend/internal/modules/events/domain"
)

type Repository interface {
	ListEvents(ctx context.Context, query domain.ListQuery) ([]domain.Event, int, error)
	ListMapEvents(ctx context.Context, query domain.ListQuery) ([]domain.MapEvent, error)
	GetEvent(ctx context.Context, id string) (*domain.Event, error)
	GetEventExternalLinks(ctx context.Context, eventID string) ([]domain.ExternalLink, error)
	GetEventSessions(ctx context.Context, eventID string) ([]domain.SessionSummary, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListEvents(ctx context.Context, query domain.ListQuery) ([]domain.Event, int, error) {
	query = normalizeListQuery(query)
	events, total, err := s.repo.ListEvents(ctx, query)
	if err != nil {
		return nil, 0, apperrors.Internal(err)
	}
	return events, total, nil
}

func (s *Service) ListMapEvents(ctx context.Context, query domain.ListQuery) ([]domain.MapEvent, error) {
	query = normalizeListQuery(query)
	events, err := s.repo.ListMapEvents(ctx, query)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	return events, nil
}

func (s *Service) GetEvent(ctx context.Context, id string) (*domain.Event, error) {
	event, err := s.repo.GetEvent(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.New(apperrors.CodeNotFound, "Event not found", http.StatusNotFound)
		}
		return nil, apperrors.Internal(err)
	}
	externalLinks, err := s.repo.GetEventExternalLinks(ctx, id)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	sessions, err := s.repo.GetEventSessions(ctx, id)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	event.ExternalLinks = externalLinks
	event.Sessions = sessions
	return event, nil
}

func (s *Service) GetEventSessions(ctx context.Context, eventID string) ([]domain.SessionSummary, error) {
	if _, err := s.repo.GetEvent(ctx, eventID); err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.New(apperrors.CodeNotFound, "Event not found", http.StatusNotFound)
		}
		return nil, apperrors.Internal(err)
	}
	sessions, err := s.repo.GetEventSessions(ctx, eventID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	return sessions, nil
}

func normalizeListQuery(query domain.ListQuery) domain.ListQuery {
	if query.Limit <= 0 {
		query.Limit = 24
	}
	if query.Limit > 200 {
		query.Limit = 200
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	return query
}

func ParseDateRangeValue(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	layouts := []string{time.RFC3339, "2006-01-02"}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return &parsed, nil
		}
	}
	return nil, apperrors.Validation("invalid date format")
}
