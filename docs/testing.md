# Testing

## Unit-тесты

Unit-тесты покрывают:

- переходы статусов брони
- `CanConfirm`
- `CanCancel`
- `CanExpire`
- расчёт времени удержания
- доменные ошибки
- валидацию forced-статуса платежа

Запуск:

```sh
cd reserveflow
make test
```

## Интеграционные тесты

Интеграционные тесты запускаются под build tag `integration` и используют testcontainers.

Покрытие:

- register/login/me
- список событий
- сеансы события
- детали сеанса
- карта мест
- удержание места
- успешная оплата
- идемпотентный replay платежа
- конфликт идемпотентности
- защита по владельцу платежа
- неуспешная оплата и освобождение места
- освобождение места по expiration job
- критичный конкурентный hold

Запуск:

```sh
cd reserveflow
make test-integration
```

## Критичный тест конкуренции

`backend/tests/booking_concurrency_integration_test.go` поднимает PostgreSQL, применяет миграции и seed, затем запускает 20 параллельных попыток hold для одного `session_seat`.

Ожидаемый результат:

- 1 успешное удержание
- 19 конфликтов
- ровно одна `pending` бронь
- `session_seat.status = held`

Этот тест подтверждает, что row lock в PostgreSQL защищает от double-booking.
