# ReserveFlow Codex Instructions

- Project is backend-only for now.
- Do not implement frontend or design.
- Main backend architecture is modular monolith.
- Keep domain logic out of HTTP handlers.
- PostgreSQL is source of truth for booking consistency.
- Redis is only for temporary state, cache, idempotency, rate limit.
- Use SELECT FOR UPDATE for hold flow.
- Protect against double booking.
- Keep APIs under `/api/v1`.
- Every new feature must include tests or documented reason why not.
- Update docs when API/domain changes.
- Run `go test ./...` before final response when possible.
