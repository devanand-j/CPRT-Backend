# CPRT LIS Backend (Go + PostgreSQL)

Starter backend for a Lab Information System. Built with Go, Echo, and PostgreSQL.

## Requirements
- Go 1.22+
- PostgreSQL 13+

## Setup
1. Copy .env.example to .env and update values.
2. Apply migrations from migrations/001_init.sql.
3. Start the API:

```
go run ./cmd/api
```

## API Endpoints (sample)
- POST /api/v1/auth/login
- POST /api/v1/patients
- GET /api/v1/patients/:id
- GET /api/v1/patients?mrn=&phone=
- POST /api/v1/billing/bills
- POST /api/v1/billing/bills/:id/items
- POST /api/v1/orders
- PATCH /api/v1/orders/:id/status

## Notes
- Create a user row with a bcrypt password hash to log in.
- All endpoints except /health and /auth/login require a Bearer token.
