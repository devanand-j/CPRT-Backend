# CPRT LIS Backend (Go + PostgreSQL)

Starter backend for a Lab Information System. Built with Go, Echo, and PostgreSQL.

## Requirements
- Go 1.22+
- PostgreSQL 13+

## Setup
1. Copy .env.example to .env and update values.
2. Apply migrations from migrations/001_core_lis_schema.sql.
3. Start the API:

```
go run ./cmd/api
```

## API Endpoints (sample)
- POST /api/auth/login
- GET /api/users
- PATCH /api/users/:userId
- POST /api/patients
- GET /api/patients/:id
- GET /api/patients
- POST /api/billing/bills
- POST /api/billing/bills/:id/items
- POST /api/orders
- PATCH /api/orders/:id/status

## Notes
- Create a user row with a bcrypt password hash to log in.
- All endpoints except /health and /auth/login require a Bearer token.
