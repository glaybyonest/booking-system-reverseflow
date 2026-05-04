package application

import (
	"context"
	"net/http"

	apperrors "reserveflow/backend/internal/infrastructure/errors"
	"reserveflow/backend/internal/infrastructure/observability"
	"reserveflow/backend/internal/modules/notifications/domain"
)

type Repository interface {
	List(ctx context.Context, userID string) ([]domain.Notification, error)
	MarkRead(ctx context.Context, userID, id string) error
	CreateFromEvent(ctx context.Context, eventID, eventType, userID string, template domain.Template) (bool, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, userID string) ([]domain.Notification, error) {
	notifications, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	return notifications, nil
}

func (s *Service) MarkRead(ctx context.Context, userID, id string) error {
	if err := s.repo.MarkRead(ctx, userID, id); err != nil {
		return apperrors.New(apperrors.CodeNotFound, "Notification not found", http.StatusNotFound)
	}
	return nil
}

func (s *Service) HandleDomainEvent(ctx context.Context, eventID, eventType, userID string) error {
	template, ok := domain.TemplateForEvent(eventType)
	if !ok || userID == "" {
		return nil
	}
	created, err := s.repo.CreateFromEvent(ctx, eventID, eventType, userID, template)
	if err != nil {
		return err
	}
	if created {
		observability.NotificationCreatedTotal.Inc()
	}
	return nil
}
