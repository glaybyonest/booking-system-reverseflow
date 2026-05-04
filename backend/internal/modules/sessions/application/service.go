package application

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"

	apperrors "reserveflow/backend/internal/infrastructure/errors"
	"reserveflow/backend/internal/modules/sessions/domain"
)

type Repository interface {
	GetSession(ctx context.Context, id string) (*domain.Session, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetSession(ctx context.Context, id string) (*domain.Session, error) {
	session, err := s.repo.GetSession(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.New(apperrors.CodeNotFound, "Session not found", http.StatusNotFound)
		}
		return nil, apperrors.Internal(err)
	}
	return session, nil
}
