# Booking Flow

## Сценарий удержания места (Hold)

`POST /api/v1/bookings/hold`

1. API получает `userId` из JWT.
2. Приложение валидирует `sessionId` и `seatId`.
3. Открывается транзакция PostgreSQL.
4. Репозиторий выполняет блокирующий запрос:

```sql
SELECT *
FROM session_seats
WHERE session_id = $1 AND seat_id = $2
FOR UPDATE;
```

5. Если место не `available`, возвращается `409 SEAT_NOT_AVAILABLE`.
6. Рассчитывается срок удержания: `now + HOLD_TTL`.
7. В `session_seats` статус меняется на `held`.
8. Создаётся `pending` бронь и `held` booking item.
9. В `outbox_events` записываются `seat.held` и `booking.created`.
10. Транзакция фиксируется.
11. В Redis сохраняется hold key с TTL.
12. Кэш карты мест инвалидируется.

Из-за row lock только одна параллельная транзакция может удержать конкретное место; остальные получат конфликт.

## Успешная оплата

`POST /api/v1/payments`

Use case оплаты блокирует бронь, проверяет владельца, статус `pending`, срок `expires_at` и затем создаёт mock-платёж.

При успехе:

- `payment` -> `succeeded`
- `booking` -> `confirmed`
- `booking_item` -> `booked`
- `session_seat` -> `booked`
- Redis hold key удаляется
- кэш карты мест инвалидируется
- пишутся события `payment.succeeded` и `booking.confirmed`

## Неуспешная оплата

При forced/mock failure:

- `payment` -> `failed`
- `booking` -> `payment_failed`
- `booking_item` -> `payment_failed`
- `session_seat` -> `available`
- Redis hold key удаляется
- кэш карты мест инвалидируется
- пишется событие `payment.failed`

## Истечение удержания (Expiration)

Worker выполняется каждые 15 секунд:

```sql
SELECT id
FROM bookings
WHERE status = 'pending' AND expires_at < now()
ORDER BY expires_at
LIMIT 100
FOR UPDATE SKIP LOCKED;
```

Для каждой истёкшей брони:

- `booking` и `booking_item` переводятся в `expired`
- место в `session_seats` освобождается
- пишется событие `booking.expired`
- удаляется Redis hold key
- инвалидируется кэш карты мест

Задача идемпотентна, потому что выбирает только `pending` брони с истёкшим временем.

Сначала worker собирает заблокированные `booking.id`, закрывает result set и только после этого выполняет follow-up запросы в той же транзакции. Это предотвращает `pgx connection busy`, сохраняя семантику `FOR UPDATE SKIP LOCKED`.
