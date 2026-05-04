package application

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	goredis "github.com/redis/go-redis/v9"

	apperrors "reserveflow/backend/internal/infrastructure/errors"
	rediscache "reserveflow/backend/internal/infrastructure/redis"
	"reserveflow/backend/internal/modules/seats/domain"
)

type Repository interface {
	GetSeatMap(ctx context.Context, sessionID string) (*domain.SeatMap, error)
}

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
}

type Service struct {
	repo     Repository
	cache    Cache
	cacheTTL time.Duration
}

func NewService(repo Repository, cache Cache, cacheTTL time.Duration) *Service {
	return &Service{repo: repo, cache: cache, cacheTTL: cacheTTL}
}

func (s *Service) GetSeatMap(ctx context.Context, sessionID string) (*domain.SeatMap, error) {
	key := rediscache.SeatmapKey(sessionID)
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, key); err == nil {
			var seatMap domain.SeatMap
			if json.Unmarshal([]byte(cached), &seatMap) == nil {
				return &seatMap, nil
			}
		} else if !errors.Is(err, goredis.Nil) {
			// Cache failures are intentionally non-fatal; PostgreSQL remains source of truth.
		}
	}
	seatMap, err := s.repo.GetSeatMap(ctx, sessionID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "Seat map not found", http.StatusNotFound)
	}
	if s.cache != nil {
		if encoded, err := json.Marshal(seatMap); err == nil {
			_ = s.cache.Set(ctx, key, encoded, s.cacheTTL)
		}
	}
	return seatMap, nil
}
