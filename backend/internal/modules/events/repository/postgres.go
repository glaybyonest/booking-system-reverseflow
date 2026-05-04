package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"reserveflow/backend/internal/modules/events/domain"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) ListEvents(ctx context.Context) ([]domain.Event, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, title, description, category, poster_url, status, created_at, updated_at
		FROM events
		WHERE status = 'published'
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []domain.Event
	for rows.Next() {
		var event domain.Event
		if err := rows.Scan(&event.ID, &event.Title, &event.Description, &event.Category, &event.PosterURL, &event.Status, &event.CreatedAt, &event.UpdatedAt); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (r *PostgresRepository) GetEvent(ctx context.Context, id string) (*domain.Event, error) {
	var event domain.Event
	err := r.db.QueryRow(ctx, `
		SELECT id, title, description, category, poster_url, status, created_at, updated_at
		FROM events
		WHERE id = $1
	`, id).Scan(&event.ID, &event.Title, &event.Description, &event.Category, &event.PosterURL, &event.Status, &event.CreatedAt, &event.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *PostgresRepository) GetEventSessions(ctx context.Context, eventID string) ([]domain.SessionSummary, error) {
	rows, err := r.db.Query(ctx, `
		SELECT s.id, s.event_id, s.hall_id, h.name, s.starts_at, s.ends_at, s.status
		FROM sessions s
		JOIN halls h ON h.id = s.hall_id
		WHERE s.event_id = $1
		ORDER BY s.starts_at
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []domain.SessionSummary
	for rows.Next() {
		var session domain.SessionSummary
		if err := rows.Scan(&session.ID, &session.EventID, &session.HallID, &session.HallName, &session.StartsAt, &session.EndsAt, &session.Status); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}
