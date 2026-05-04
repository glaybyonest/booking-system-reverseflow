package application

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"

	apperrors "reserveflow/backend/internal/infrastructure/errors"
	"reserveflow/backend/internal/modules/events/domain"
)

type Repository interface {
	ListEvents(ctx context.Context) ([]domain.Event, error)
	GetEvent(ctx context.Context, id string) (*domain.Event, error)
	GetEventSessions(ctx context.Context, eventID string) ([]domain.SessionSummary, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListEvents(ctx context.Context) ([]domain.Event, error) {
	events, err := s.repo.ListEvents(ctx)
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
	return event, nil
}

func (s *Service) GetEventSessions(ctx context.Context, eventID string) ([]domain.SessionSummary, error) {
	if _, err := s.GetEvent(ctx, eventID); err != nil {
		return nil, err
	}
	sessions, err := s.repo.GetEventSessions(ctx, eventID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	return sessions, nil
}
