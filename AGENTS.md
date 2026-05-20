# ReserveFlow Codex Instructions

- Project includes a Go backend, a Next.js frontend, Docker Compose, Kubernetes manifests, and docs.
- Main backend architecture is a modular monolith.
- Do not split modules into microservices unless the user explicitly asks for an architecture change.
- Keep domain logic out of HTTP handlers.
- PostgreSQL is source of truth for booking consistency.
- Redis is only for temporary state, cache, idempotency, rate limit.
- Use SELECT FOR UPDATE for hold flow.
- Protect against double booking.
- Keep APIs under `/api/v1`.
- Frontend auth goes through Next.js route handlers with HTTP-only cookies; browser code should not store tokens in localStorage.
- Preserve the current UI direction: Inter, `#F8F9FA`, white rounded cards, minimal gray/black palette, and the existing seat map style.
- Every new feature must include tests or documented reason why not.
- Update docs when API/domain changes.
- Run relevant backend and frontend checks before final response when possible.
