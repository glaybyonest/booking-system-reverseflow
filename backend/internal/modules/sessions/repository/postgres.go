package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"reserveflow/backend/internal/modules/sessions/domain"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetSession(ctx context.Context, id string) (*domain.Session, error) {
	var session domain.Session
	err := r.db.QueryRow(ctx, `
		SELECT s.id, s.event_id, e.id, e.title, s.hall_id, h.id, h.name, v.name, s.starts_at, s.ends_at, s.status
		FROM sessions s
		JOIN events e ON e.id = s.event_id
		JOIN halls h ON h.id = s.hall_id
		JOIN venues v ON v.id = h.venue_id
		WHERE s.id = $1
	`, id).Scan(
		&session.ID,
		&session.EventID,
		&session.Event.ID,
		&session.Event.Title,
		&session.HallID,
		&session.Hall.ID,
		&session.Hall.Name,
		&session.Hall.Venue,
		&session.StartsAt,
		&session.EndsAt,
		&session.Status,
	)
	if err != nil {
		return nil, err
	}
	return &session, nil
}
