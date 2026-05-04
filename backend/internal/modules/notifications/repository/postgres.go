package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"reserveflow/backend/internal/modules/notifications/domain"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) List(ctx context.Context, userID string) ([]domain.Notification, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, type, title, message, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.Notification, 0)
	for rows.Next() {
		var notification domain.Notification
		if err := rows.Scan(&notification.ID, &notification.UserID, &notification.Type, &notification.Title, &notification.Message, &notification.IsRead, &notification.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, notification)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) MarkRead(ctx context.Context, userID, id string) error {
	cmd, err := r.db.Exec(ctx, `
		UPDATE notifications SET is_read = TRUE WHERE id = $1 AND user_id = $2
	`, id, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PostgresRepository) CreateFromEvent(ctx context.Context, eventID, eventType, userID string, template domain.Template) (bool, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
		INSERT INTO processed_events (id, event_id, event_type, processed_at)
		VALUES ($1, $2, $3, $4)
	`, uuid.NewString(), eventID, eventType, time.Now().UTC())
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return false, nil
		}
		return false, err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO notifications (id, user_id, type, title, message, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, FALSE, $6)
	`, uuid.NewString(), userID, template.Type, template.Title, template.Message, time.Now().UTC())
	if err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}
