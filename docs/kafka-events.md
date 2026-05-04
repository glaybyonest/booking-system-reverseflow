# Kafka Events

В ReserveFlow доменные события сначала фиксируются в PostgreSQL, а затем асинхронно публикуются в Kafka из таблицы `outbox_events`.

## Топики

- `seat.held`
- `booking.created`
- `payment.succeeded`
- `payment.failed`
- `booking.confirmed`
- `booking.expired`
- `booking.cancelled`
- `notification.created`

## Формат события (envelope)

```json
{
  "eventId": "uuid",
  "eventType": "booking.confirmed",
  "aggregateType": "booking",
  "aggregateId": "uuid",
  "occurredAt": "2026-05-04T12:00:00Z",
  "payload": {
    "userId": "uuid",
    "bookingId": "uuid",
    "sessionId": "uuid",
    "seatId": "uuid"
  }
}
```

## Notification consumer

Worker подписан на:

- `booking.confirmed`
- `booking.expired`
- `payment.failed`
- `booking.cancelled`

Идемпотентность consumer-а хранится в `processed_events`. Если одинаковый `eventId` приходит повторно, уведомление создаётся только один раз.

## Outbox pattern

Бизнес-транзакции пишут события в `outbox_events` до commit. Далее worker:

1. Выбирает pending-записи с `FOR UPDATE SKIP LOCKED`.
2. Публикует сообщение в Kafka-топик с именем `event_type`.
3. Помечает запись как `published`.
4. При ошибке публикации повторяет отправку позже.
