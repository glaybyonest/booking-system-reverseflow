# Kafka Events

ReserveFlow writes domain events to PostgreSQL first and publishes them asynchronously from `outbox_events`.

## Topics

- `seat.held`
- `booking.created`
- `payment.succeeded`
- `payment.failed`
- `booking.confirmed`
- `booking.expired`
- `booking.cancelled`
- `notification.created`

## Envelope

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

## Notification Consumer

The worker listens to:

- `booking.confirmed`
- `booking.expired`
- `payment.failed`
- `booking.cancelled`

Consumer idempotency is stored in `processed_events`. If the same `eventId` is received twice, only one notification is created.

## Outbox Pattern

Business transactions insert into `outbox_events` before commit. The worker:

1. Selects pending rows with `FOR UPDATE SKIP LOCKED`.
2. Publishes to Kafka topic named by `event_type`.
3. Marks the row `published`.
4. Retries later if Kafka publish fails.
