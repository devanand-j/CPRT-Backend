# CPRT LIS — Backend API

Laboratory Information System (LIS) REST API built with **Go**, **Echo** and **PostgreSQL**.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.25+ |
| HTTP Framework | Echo v4 |
| Database | PostgreSQL 13+ |
| ORM / Query | pgx v5 (raw SQL) |
| Migrations | golang-migrate |
| API Docs | Swagger UI + ReDoc (swaggo) |
| Auth | JWT (Bearer token) |

---

## Prerequisites

- Go 1.25+
- PostgreSQL 13+
- `swag` CLI — install once:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

---

## Environment Variables

Create a `.env` file at the project root (copy the block below and fill in your values):

```env
PORT=8080
DATABASE_URL=host=<host> user=<user> password=<password> dbname=<dbname> port=5432 sslmode=require
JWT_SECRET=your-secret-key
JWT_ISSUER=cprt-lis
JWT_TTL_MINUTES=60
```

> `DATABASE_URL` accepts both key=value DSN format and `postgres://` URL format.

---

## Running Locally

```bash
# 1. Clone and enter the project
cd backend

# 2. Install dependencies
go mod download

# 3. Set up your .env file (see above)

# 4. Start the server
go run .
```

On startup the server will:
1. Connect to the database
2. **Auto-run any pending migrations** from the `migrations/` folder
3. Start listening on the configured port

---

## Database Migrations

Migrations live in `migrations/` as numbered `.up.sql` files:

**They run automatically on every server start.** Already-applied migrations are skipped — tracked via a `schema_migrations` table in your database.

To add a new migration, create the next numbered file:

```
migrations/008_your_change.up.sql
```

It will be applied automatically on the next startup.

---

## Seed Data

After running migrations, seed initial users:

```bash
psql "$DATABASE_URL" -f seed_users.sql
```

Default credentials for all seeded users: `password123`

| Login ID | Role |
|---|---|
| `superadmin` | SUPER_ADMIN |
| `admin` | ADMIN |
| `doctor1` | DOCTOR |
| `receptionist1` | RECEPTIONIST |
| `tech1` | TECHNICIAN |

---

## API Documentation

Once the server is running, two interactive documentation UIs are available:

| UI | URL |
|---|---|
| **Swagger UI** | http://localhost:8080/swagger/index.html |
| **ReDoc** | http://localhost:8080/redoc |

### Authentication in Swagger UI

1. Call `POST /api/auth/login` with your credentials
2. Copy the `token` value from the response
3. Click the **Authorize** button (top right)
4. Enter: `Bearer <your-token>`
5. All secured endpoints will now work

### Regenerating Docs

Run this whenever you change handler annotations:

```bash
swag init -g main.go -o ./docs --parseDependency --parseInternal
```

---

## Docker

```bash
# Build
docker build -t cprt-lis .

# Run
docker run -p 8080:8080 --env-file .env cprt-lis
```

The Docker build automatically regenerates Swagger docs and compiles the binary.

---

## Lab Workflow (in order)

```
1. POST   /api/auth/login                          → get JWT token
2. POST   /api/patients/register                   → register patient, get patient_id
3. GET    /api/billing/services                    → list available tests/services
4. POST   /api/billing/new                         → generate bill with services
5. POST   /api/orders                              → create lab order for the bill
6. PATCH  /api/lab/sample-collection/:billId       → mark sample collected
7. POST   /api/lab/results/verify                  → enter and verify test results
8. PATCH  /api/lab/results/certify/:billId         → doctor certifies the report
9. GET    /api/lab/reports/:billId                 → fetch final report
```

> Full request/response schema for every endpoint is documented in Swagger UI / ReDoc.
