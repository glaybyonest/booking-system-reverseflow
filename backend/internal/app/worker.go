package app

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/segmentio/kafka-go"

	infrakafka "reserveflow/backend/internal/infrastructure/kafka"
	"reserveflow/backend/internal/infrastructure/observability"
)

func RunWorker(ctx context.Context, deps *Dependencies) error {
	deps.Log.Info().Msg("backend-worker starting")
	go expirationLoop(ctx, deps)
	go outboxLoop(ctx, deps)
	for _, topic := range []string{"booking.confirmed", "booking.expired", "payment.failed", "booking.cancelled"} {
		go notificationConsumerLoop(ctx, deps, topic)
	}
	<-ctx.Done()
	deps.Log.Info().Msg("backend-worker stopped")
	return ctx.Err()
}

func expirationLoop(ctx context.Context, deps *Dependencies) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			changes, err := deps.BookingsService.ExpireBooking(ctx, 100)
			if err != nil {
				deps.Log.Error().Err(err).Msg("expiration job failed")
				continue
			}
			if len(changes) > 0 {
				deps.Log.Info().Int("expired", len(changes)).Msg("expired pending bookings")
			}
		}
	}
}

func outboxLoop(ctx context.Context, deps *Dependencies) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := publishOutboxBatch(ctx, deps, 100); err != nil {
				deps.Log.Error().Err(err).Msg("outbox publisher failed")
			}
		}
	}
}

type outboxRow struct {
	ID            string
	EventType     string
	AggregateType string
	AggregateID   string
	Payload       []byte
	CreatedAt     time.Time
}

func publishOutboxBatch(ctx context.Context, deps *Dependencies, limit int) error {
	tx, err := deps.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
		SELECT id, event_type, aggregate_type, aggregate_id, payload, created_at
		FROM outbox_events
		WHERE status = 'pending'
		ORDER BY created_at
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	items := make([]outboxRow, 0)
	for rows.Next() {
		var item outboxRow
		if err := rows.Scan(&item.ID, &item.EventType, &item.AggregateType, &item.AggregateID, &item.Payload, &item.CreatedAt); err != nil {
			return err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	rows.Close()

	published := 0
	for _, item := range items {
		var payload map[string]any
		if err := json.Unmarshal(item.Payload, &payload); err != nil {
			deps.Log.Error().Err(err).Str("event_id", item.ID).Msg("invalid outbox payload")
			continue
		}
		envelope := infrakafka.EventEnvelope{
			EventID:       item.ID,
			EventType:     item.EventType,
			AggregateType: item.AggregateType,
			AggregateID:   item.AggregateID,
			OccurredAt:    item.CreatedAt,
			Payload:       payload,
		}
		encoded, err := json.Marshal(envelope)
		if err != nil {
			return err
		}
		if err := deps.Kafka.Publish(ctx, item.EventType, item.AggregateID, encoded); err != nil {
			deps.Log.Warn().Err(err).Str("event_id", item.ID).Str("topic", item.EventType).Msg("kafka publish failed; will retry")
			continue
		}
		if _, err := tx.Exec(ctx, `
			UPDATE outbox_events SET status = 'published', published_at = $2 WHERE id = $1
		`, item.ID, time.Now().UTC()); err != nil {
			return err
		}
		published++
		observability.OutboxPublishedTotal.Inc()
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	if published > 0 {
		deps.Log.Info().Int("published", published).Msg("outbox events published")
	}
	return nil
}

func notificationConsumerLoop(ctx context.Context, deps *Dependencies, topic string) {
	reader := infrakafka.NewReader(deps.Config.KafkaBrokers, "reserveflow-notifications", topic)
	defer reader.Close()
	deps.Log.Info().Str("topic", topic).Msg("notification consumer started")
	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			deps.Log.Warn().Err(err).Str("topic", topic).Msg("kafka fetch failed")
			time.Sleep(time.Second)
			continue
		}
		if err := handleNotificationMessage(ctx, deps, msg); err != nil {
			deps.Log.Error().Err(err).Str("topic", topic).Msg("notification handling failed")
			continue
		}
		if err := reader.CommitMessages(ctx, msg); err != nil {
			deps.Log.Warn().Err(err).Str("topic", topic).Msg("kafka commit failed")
		}
	}
}

func handleNotificationMessage(ctx context.Context, deps *Dependencies, msg kafka.Message) error {
	var envelope infrakafka.EventEnvelope
	if err := json.Unmarshal(msg.Value, &envelope); err != nil {
		return err
	}
	userID, _ := envelope.Payload["userId"].(string)
	return deps.NotificationsService.HandleDomainEvent(ctx, envelope.EventID, envelope.EventType, userID)
}
