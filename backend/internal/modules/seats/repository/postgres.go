package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"reserveflow/backend/internal/modules/seats/domain"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetSeatMap(ctx context.Context, sessionID string) (*domain.SeatMap, error) {
	var seatMap domain.SeatMap
	err := r.db.QueryRow(ctx, `
		SELECT s.id, e.id, e.title, h.id, h.name
		FROM sessions s
		JOIN events e ON e.id = s.event_id
		JOIN halls h ON h.id = s.hall_id
		WHERE s.id = $1
	`, sessionID).Scan(&seatMap.SessionID, &seatMap.Event.ID, &seatMap.Event.Title, &seatMap.Hall.ID, &seatMap.Hall.Name)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT seat.id, seat.row_label, seat.seat_number, ss.status, seat.base_price
		FROM session_seats ss
		JOIN seats seat ON seat.id = ss.seat_id
		WHERE ss.session_id = $1
		ORDER BY seat.row_label, seat.seat_number
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var seat domain.Seat
		if err := rows.Scan(&seat.SeatID, &seat.Row, &seat.Number, &seat.Status, &seat.Price); err != nil {
			return nil, err
		}
		seatMap.Seats = append(seatMap.Seats, seat)
	}
	return &seatMap, rows.Err()
}
