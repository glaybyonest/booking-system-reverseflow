package application

import (
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog"

	rediscache "reserveflow/backend/internal/infrastructure/redis"
	seatsdomain "reserveflow/backend/internal/modules/seats/domain"
)

type Repository interface {
	GetSessionLayoutState(ctx context.Context, sessionID string) (*SessionLayoutState, error)
	UpsertSessionLayout(ctx context.Context, sessionID string, layout seatsdomain.StoredSeatLayout, now time.Time) (*LayoutMutationResult, error)
	DeleteSessionLayout(ctx context.Context, sessionID string, now time.Time) (*LayoutMutationResult, error)
	GetHallLayoutState(ctx context.Context, hallID string) (*HallLayoutState, error)
	UpsertHallLayout(ctx context.Context, hallID string, layout seatsdomain.StoredSeatLayout, now time.Time) (*LayoutMutationResult, error)
}

type Cache interface {
	Del(ctx context.Context, keys ...string) error
}

type Service struct {
	repo  Repository
	log   zerolog.Logger
	cache Cache
}

func NewService(repo Repository, log zerolog.Logger) *Service {
	return &Service{repo: repo, log: log}
}

func (s *Service) WithCache(cache Cache) *Service {
	s.cache = cache
	return s
}

func (s *Service) invalidateSeatmaps(ctx context.Context, sessionIDs ...string) {
	if s.cache == nil || len(sessionIDs) == 0 {
		return
	}

	seen := make(map[string]struct{}, len(sessionIDs))
	keys := make([]string, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		sessionID = strings.TrimSpace(sessionID)
		if sessionID == "" {
			continue
		}
		if _, exists := seen[sessionID]; exists {
			continue
		}
		seen[sessionID] = struct{}{}
		keys = append(keys, rediscache.SeatmapKey(sessionID))
	}
	if len(keys) == 0 {
		return
	}
	if err := s.cache.Del(ctx, keys...); err != nil {
		s.log.Warn().Err(err).Strs("seatmap_keys", keys).Msg("failed to invalidate seat map cache after layout change")
	}
}
