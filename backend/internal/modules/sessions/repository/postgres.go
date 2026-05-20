package repository

import (
	"context"
	"database/sql"

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
	var hallID sql.NullString
	var hallName sql.NullString
	var venueName sql.NullString
	var startsAt sql.NullTime
	var endsAt sql.NullTime
	var externalSource sql.NullString
	var externalID sql.NullString
	var sourceURL sql.NullString
	err := r.db.QueryRow(ctx, `
		SELECT
			s.id,
			s.event_id,
			e.id,
			e.title,
			s.hall_id,
			h.name,
			v.name,
			s.starts_at,
			s.ends_at,
			s.status,
			s.is_bookable,
			s.external_source,
			s.external_id,
			s.source_url
		FROM sessions s
		JOIN events e ON e.id = s.event_id
		LEFT JOIN halls h ON h.id = s.hall_id
		LEFT JOIN venues v ON v.id = COALESCE(h.venue_id, e.venue_id)
		WHERE s.id = $1
	`, id).Scan(
		&session.ID,
		&session.EventID,
		&session.Event.ID,
		&session.Event.Title,
		&hallID,
		&hallName,
		&venueName,
		&startsAt,
		&endsAt,
		&session.Status,
		&session.IsBookable,
		&externalSource,
		&externalID,
		&sourceURL,
	)
	if err != nil {
		return nil, err
	}
	if hallID.Valid {
		session.HallID = &hallID.String
	}
	if hallName.Valid || venueName.Valid {
		session.Hall = &domain.HallRef{
			Name:  hallName.String,
			Venue: venueName.String,
		}
		if hallID.Valid {
			session.Hall.ID = &hallID.String
		}
	}
	if startsAt.Valid {
		start := startsAt.Time
		session.StartsAt = &start
	}
	if endsAt.Valid {
		end := endsAt.Time
		session.EndsAt = &end
	}
	if externalSource.Valid {
		session.ExternalSource = &externalSource.String
	}
	if externalID.Valid {
		session.ExternalID = &externalID.String
	}
	if sourceURL.Valid {
		session.SourceURL = &sourceURL.String
	}
	return &session, nil
}
