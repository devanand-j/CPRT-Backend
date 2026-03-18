# CPRT LIS — Knowledge Transfer Document

**Project:** CPRT Laboratory Information System (LIS)  
**Module:** cprt-lis (Go backend)  
**Language / Runtime:** Go 1.22  
**Database:** PostgreSQL 13+  
**Date Prepared:** March 2026

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Tech Stack & Dependencies](#2-tech-stack--dependencies)
3. [Repository Structure](#3-repository-structure)
4. [Environment Setup](#4-environment-setup)
5. [Architecture Overview](#5-architecture-overview)
6. [Database Schema](#6-database-schema)
   - 6.1 Core Tables
   - 6.2 Schema Evolution (migrations)
7. [Domain Models (Go Structs)](#7-domain-models-go-structs)
8. [Authentication & Authorization](#8-authentication--authorization)
9. [API Endpoints Reference](#9-api-endpoints-reference)
10. [Module Deep-Dives](#10-module-deep-dives)
    - 10.1 Auth Module
    - 10.2 Users Module
    - 10.3 Patients Module
    - 10.4 Billing Module
    - 10.5 Orders Module
    - 10.6 Lab Workflow Module
11. [Lab Workflow — End-to-End Business Flow](#11-lab-workflow--end-to-end-business-flow)
12. [Code Layering Pattern](#12-code-layering-pattern)
13. [Seeded Test Data](#13-seeded-test-data)
14. [Known Quirks & Important Notes](#14-known-quirks--important-notes)

---

## 1. Project Overview

CPRT LIS is a **Laboratory Information System** backend built entirely in Go. It manages the full lifecycle of a patient's visit to a diagnostic lab:

```
Patient Registration → Bill Generation → Lab Order → Sample Collection
        → Result Entry (Verify) → Result Sign-off (Certify) → Report Fetch
```

All data is stored in a single PostgreSQL database. The HTTP layer is thin and stateless — every request is authenticated with a short-lived JWT.

---

## 2. Tech Stack & Dependencies

| Concern | Library | Version |
|---|---|---|
| HTTP framework | `github.com/labstack/echo/v4` | v4.12.0 |
| Database driver | `github.com/jackc/pgx/v5` (pgxpool) | v5.5.5 |
| JWT generation & parsing | `github.com/golang-jwt/jwt/v5` | v5.2.1 |
| Password hashing | `golang.org/x/crypto/bcrypt` | v0.22.0 |
| `.env` loading | `github.com/joho/godotenv` | v1.5.1 |
| ORM (declared but not used in queries) | `gorm.io/gorm` | v1.25.11 |

> **Note:** GORM is in `go.mod` but the actual repository code uses **raw SQL through pgx/pgxpool** — not GORM. All database interaction is hand-written SQL.

---

## 3. Repository Structure

```
d:\CPRT\
├── go.mod                        # Module name: cprt-lis
├── README.md
├── seed_users.sql                # Seed script for initial users + account groups
├── postman_collection.json       # Postman collection for all endpoints
│
├── cmd/
│   └── api/
│       └── main.go               # Entry point — wires everything together + HTTP server
│
├── internal/
│   ├── config/
│   │   └── config.go             # Reads .env / environment variables
│   │
│   ├── db/
│   │   └── postgres.go           # Creates pgxpool connection (max 20, min 2 conns)
│   │
│   ├── domain/
│   │   └── models.go             # Pure Go structs — no DB or HTTP tags for most
│   │
│   ├── http/
│   │   ├── router.go             # Echo router setup + route registration
│   │   ├── handlers/
│   │   │   ├── contracts.go      # Service interfaces consumed by handlers
│   │   │   ├── auth.go           # POST /api/auth/login
│   │   │   ├── users.go          # GET /api/users, PATCH /api/users/:userId
│   │   │   ├── patients.go       # Patient CRUD + search + history
│   │   │   ├── billing.go        # Bill generation, services, payment
│   │   │   ├── orders.go         # Lab order create + status update
│   │   │   ├── lab.go            # Sample → Verify → Certify → Report
│   │   │   └── health.go         # GET /health (no auth)
│   │   └── middleware/
│   │       └── jwt.go            # JWT validation + role guards
│   │
│   ├── repository/
│   │   ├── interfaces.go         # Repository interfaces (port definitions)
│   │   └── postgres/
│   │       ├── users.go
│   │       ├── patients.go
│   │       ├── billing.go
│   │       ├── orders.go
│   │       └── lab.go
│   │
│   └── service/
│       ├── auth.go               # Login + JWT signing
│       ├── users.go
│       ├── patients.go
│       ├── billing.go
│       ├── orders.go
│       └── lab.go
│
└── migrations/
    ├── 001_core_lis_schema.sql   # All base tables
    ├── 002_patient_master_extensions.sql
    ├── 003_billing_financial_extensions.sql
    ├── 004_identity_auth_extensions.sql
    ├── 005_patient_billing_combined_seed.sql
    ├── 006_lab_workflow_extensions.sql
    └── 007_lab_results_conflict_fix.sql
```

---

## 4. Environment Setup

### 4.1 Required `.env` file (place at project root)

```env
PORT=8080
DATABASE_URL=postgres://USER:PASSWORD@HOST:5432/DBNAME?sslmode=disable
JWT_SECRET=your-super-secret-key
JWT_ISSUER=cprt-lis
JWT_TTL_MINUTES=60
```

The config loader (`internal/config/config.go`) walks up the directory tree to find a `.env` file, so it works regardless of which subdirectory you run from.

### 4.2 Database setup (run in order)

```sql
-- 1. Base schema
\i migrations/001_core_lis_schema.sql

-- 2. Patient extensions
\i migrations/002_patient_master_extensions.sql

-- 3. Billing extensions
\i migrations/003_billing_financial_extensions.sql

-- 4. Identity / auth extensions
\i migrations/004_identity_auth_extensions.sql

-- 5. Combined seed data
\i migrations/005_patient_billing_combined_seed.sql

-- 6. Lab workflow extensions
\i migrations/006_lab_workflow_extensions.sql

-- 7. Lab results unique-index fix
\i migrations/007_lab_results_conflict_fix.sql

-- 8. Seed users
\i seed_users.sql
```

> Migrations 002 and 003 are **redundant** with migration 005 (005 runs both ALTERs again inside a transaction, safely using `IF NOT EXISTS`). Running them is safe but not strictly necessary if 005 is run.

### 4.3 Running the server

```bash
go run ./cmd/api
# or build + run:
go build -o cprt-lis ./cmd/api && ./cprt-lis
```

---

## 5. Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP Client / Postman                 │
└────────────────────────┬────────────────────────────────┘
                         │  HTTP (JSON)
┌────────────────────────▼────────────────────────────────┐
│                   Echo HTTP Server                       │
│  ┌──────────────────────────────────────────────────┐   │
│  │  Middleware: Logger, Recover, CORS               │   │
│  │  Middleware: JWTAuth (validates Bearer token)    │   │
│  │  Middleware: RolePolicyGuard (path-based RBAC)   │   │
│  │  Middleware: RequireRole (per-route RBAC)        │   │
│  └──────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────┐   │
│  │  Handlers (internal/http/handlers/)              │   │
│  │  - Bind & validate request                       │   │
│  │  - Call service method                           │   │
│  │  - Return JSON response                          │   │
│  └──────────────────────────────────────────────────┘   │
└────────────────────────┬────────────────────────────────┘
                         │  Go interfaces
┌────────────────────────▼────────────────────────────────┐
│                  Service Layer                           │
│  (internal/service/) — thin orchestration only          │
│  No business-logic validation beyond what handlers do   │
└────────────────────────┬────────────────────────────────┘
                         │  Repository interfaces
┌────────────────────────▼────────────────────────────────┐
│              Repository Layer (postgres/)                │
│  Raw SQL via pgx/pgxpool — all queries live here        │
└────────────────────────┬────────────────────────────────┘
                         │  pgxpool
┌────────────────────────▼────────────────────────────────┐
│                   PostgreSQL Database                    │
└─────────────────────────────────────────────────────────┘
```

The design is a **clean 3-layer architecture**:
- **Handler** — HTTP concerns only (parse, validate input, format output)
- **Service** — thin pass-through; intended for future business logic
- **Repository** — all SQL, no HTTP knowledge

Dependency direction: `Handler → Service interface → Repository interface → Concrete postgres implementation`

Interfaces are defined in two places:
- `internal/http/handlers/contracts.go` — service interfaces used by handlers
- `internal/repository/interfaces.go` — repository interfaces used by services

---

## 6. Database Schema

### 6.1 Core Tables (from `001_core_lis_schema.sql`)

#### `roles`
| Column | Type | Notes |
|---|---|---|
| id | BIGSERIAL PK | |
| role_code | TEXT UNIQUE | e.g. `ADMIN` |
| role_name | TEXT | |

#### `account_groups`
| Column | Type | Notes |
|---|---|---|
| id | BIGSERIAL PK | |
| group_code | TEXT UNIQUE | `SUPER_ADMIN`, `ADMIN`, `DOCTOR`, `RECEPTIONIST`, `TECHNICIAN` |
| group_name | TEXT | |
| status | TEXT | Default `ACTIVE` |

#### `users`
| Column | Type | Notes |
|---|---|---|
| id | BIGSERIAL PK | |
| user_uuid | UUID UNIQUE | |
| username | TEXT UNIQUE | Used for login |
| email | TEXT | |
| phone | TEXT | |
| password_hash | TEXT | bcrypt hash |
| status | TEXT | `ACTIVE` / `Inactive` |
| created_at | TIMESTAMPTZ | |
| group_id | BIGINT → account_groups | Added in migration 004 |
| login_id | TEXT UNIQUE | Added in 004, initially copied from username |
| user_name | TEXT | Display name, added in 004 |
| last_login | TIMESTAMPTZ | Added in 004 |
| updated_by | TEXT | Added in 004 |
| updated_at | TIMESTAMPTZ | Added in 004 |

#### `user_account_groups`
Junction table — links `users` ↔ `account_groups` (many-to-many, legacy). After migration 004, `users.group_id` is the primary group reference.

#### `patients`
| Column | Type | Notes |
|---|---|---|
| id | BIGSERIAL PK | |
| patient_uuid | UUID UNIQUE | External identifier used in all API calls |
| patient_no | BIGSERIAL UNIQUE | Auto-incrementing display number |
| prefix | TEXT | `Mr.`, `Ms.`, etc. |
| first_name | TEXT NOT NULL | Full name stored in one field |
| gender | TEXT | |
| age | INT | |
| age_unit | TEXT | `Yrs`, `Months`, `Days` |
| phone | TEXT | |
| op_ip_no | TEXT | Outpatient/Inpatient number |
| created_at | TIMESTAMPTZ | |
| patient_type | TEXT | Added in 002; `Out Patients`, `In Patients` |
| created_by | TEXT | Added in 002 |
| status | TEXT | Added in 002; default `Active` |

#### `lab_services`
| Column | Type | Notes |
|---|---|---|
| id | BIGSERIAL PK | |
| service_code | TEXT UNIQUE | |
| service_name | TEXT | |
| specimen_type | TEXT | e.g. `Blood` |
| department | TEXT | |
| base_price | NUMERIC(12,2) | |
| tat_minutes | INT | Turnaround time |
| status | TEXT | Default `ACTIVE` |

#### `lab_bills`
| Column | Type | Notes |
|---|---|---|
| id | BIGSERIAL PK | Used internally in all lab operations |
| bill_uuid | UUID UNIQUE | External identifier returned to clients |
| bill_no | BIGSERIAL UNIQUE | Human-readable bill number (added migration 003) |
| patient_id | BIGINT → patients | |
| visit_id | BIGINT → patient_visits | Optional |
| doctor_id | BIGINT → users | Optional |
| referred_by | TEXT | Added in 003 |
| hospital_name | TEXT | Added in 003 |
| total_amount | NUMERIC(12,2) | Sum of all service rates |
| discount_amount | NUMERIC(12,2) | |
| tax_amount | NUMERIC(12,2) | Calculated: `(total - discount) * tax% / 100` |
| net_amount | NUMERIC(12,2) | `total - discount + tax` |
| received_amount | NUMERIC(12,2) | Added in 003; tracks payments |
| balance_amount | NUMERIC(12,2) | Added in 003; `net - received` |
| payment_status | TEXT | Added in 003; `Paid`, `Pending`, `Overpaid` |
| payment_mode | TEXT | `CASH`, `CARD`, etc. |
| status | TEXT | Default `DRAFT`; set to `FINALIZED` on GenerateBill |
| report_status | TEXT | Added in 006; `Pending` → `Collected` → `Verified` → `Certified` |
| certified_by_user | TEXT | Added in 006 |
| certification_remarks | TEXT | Added in 006 |
| certified_at | TIMESTAMPTZ | Added in 006 |
| dispatch_ready | BOOLEAN | Added in 006; set true on certify |
| created_at | TIMESTAMPTZ | |

#### `lab_bill_items`
| Column | Type | Notes |
|---|---|---|
| id | BIGSERIAL PK | |
| bill_id | BIGINT → lab_bills | |
| service_id | BIGINT → lab_services | Nullable (GenerateBill passes NULL for ad-hoc items) |
| qty | INT | Default 1 |
| unit_price | NUMERIC(12,2) | |
| discount | NUMERIC(12,2) | Default 0 |
| tax | NUMERIC(12,2) | Default 0 |
| line_total | NUMERIC(12,2) | `qty * unit_price` |
| status | TEXT | Default `PENDING` |

#### `lab_orders`
| Column | Type | Notes |
|---|---|---|
| id | BIGSERIAL PK | |
| order_uuid | UUID UNIQUE | |
| bill_id | BIGINT → lab_bills | |
| patient_id | BIGINT → patients | |
| visit_id | BIGINT → patient_visits | Optional |
| order_status | TEXT | `PENDING`, `IN_PROGRESS`, `COMPLETED` |
| created_at | TIMESTAMPTZ | |

#### `samples`
| Column | Type | Notes |
|---|---|---|
| id | BIGSERIAL PK | |
| sample_uuid | UUID UNIQUE | |
| order_id | BIGINT → lab_orders | |
| specimen_type | TEXT | Hardcoded to `Blood` currently |
| barcode | TEXT UNIQUE | `sample_no` from API request |
| collected_by | BIGINT → users | Note: stored as user_id (integer) in schema; API sends text name |
| collected_at | TIMESTAMPTZ | |
| status | TEXT | `PENDING` / `Collected` |

#### `worksheets`
| Column | Type | Notes |
|---|---|---|
| id | BIGSERIAL PK | |
| worksheet_uuid | UUID UNIQUE | |
| created_by | BIGINT → users | |
| created_at | TIMESTAMPTZ | |
| status | TEXT | Default `OPEN` |

> **Note:** Worksheet numbers are generated as `WS-{MAX(id)+1}` in code — they are not stored back in the worksheets table by the sample collection flow. The worksheets table is populated elsewhere or manually.

#### `lab_results`
| Column | Type | Notes |
|---|---|---|
| id | BIGSERIAL PK | |
| order_item_id | BIGINT → order_items | Original FK (may be NULL when using bill-based flow) |
| bill_id | BIGINT → lab_bills | Added in 006; used by Result Verify/Certify flow |
| param_id | TEXT | Added in 006; e.g. `HB`, `WBC` |
| param_name | TEXT | Added in 006; human-readable |
| result_value | TEXT | |
| unit | TEXT | |
| reference_range | TEXT | |
| abnormal_flag | TEXT | `Y` / `N` |
| result_status | TEXT | `ENTERED` → `Verified` |
| verified_by | BIGINT → users | Original column; references user ID |
| verified_by_user | TEXT | Added in 006; stores username string |
| verified_at | TIMESTAMPTZ | |

**Unique constraint:** `uq_lab_results_bill_param_full ON lab_results(bill_id, param_id)` (migration 007 recreated this without the partial-index WHERE clause from 006, making ALL `(bill_id, param_id)` pairs unique).

#### `audit_logs`
Declared in 001 but truncated in the migration file visible here — columns are `id`, `entity_type`, `entity_id`, `action`, `actor_id`. Not used by any current API endpoint.

#### Other tables (declared, not actively used by current API)
- `patient_demographics` — extended patient info
- `patient_visits` — visit tracking
- `hospitals`, `wards` — facility data
- `price_profiles`, `service_prices` — tiered pricing
- `order_items`, `worksheet_samples` — granular order/worksheet tracking

---

### 6.2 Schema Evolution Summary

| Migration | What Changed |
|---|---|
| 001 | All base tables created |
| 002 | Added `patient_type`, `created_by`, `status` to `patients` |
| 003 | Added `bill_no`, `referred_by`, `hospital_name`, `received_amount`, `balance_amount`, `payment_status` to `lab_bills` |
| 004 | Added `group_id`, `login_id`, `user_name`, `last_login`, `updated_by`, `updated_at` to `users`; back-filled `group_id` from `user_account_groups`; added unique index on `login_id` |
| 005 | Combined 002+003 ALTERerations + seeded 2 patients (idempotent) |
| 006 | Added `report_status`, `certified_by_user`, `certification_remarks`, `certified_at`, `dispatch_ready` to `lab_bills`; added `bill_id`, `param_id`, `param_name`, `verified_by_user` to `lab_results`; unique partial index on `lab_results(bill_id, param_id)` |
| 007 | Dropped partial index from 006 and replaced with full unique index on `lab_results(bill_id, param_id)` |

---

## 7. Domain Models (Go Structs)

All domain types live in `internal/domain/models.go`. These are plain Go structs — no ORM annotations.

### `User`
Maps to the `users` table joined with `account_groups`. Key fields: `UserUUID` (external ID), `GroupCode` (used as the JWT `role` claim), `PasswordHash` (bcrypt).

### `Patient`
Maps to the `patients` table. `PatientUUID` is the external identifier. `PatientNo` is the auto-incrementing display number.

### `LabBill`
Maps to `lab_bills`. `BillUUID` is returned to the client; internal queries use `ID` (bigint). `PaymentStatus` is calculated: `Paid` / `Pending` / `Overpaid`.

### `LabOrder`
Maps to `lab_orders`. Linked to a bill and patient.

### `LabService`
Maps to `lab_services`. Retrieved by GET /billing/services.

### Response-only structs (no direct table mapping)
- `PatientSearchResult` — flattened patient search result
- `PatientHistoryItem` — bill + service line for patient history
- `SampleCollectionResponse` — returned after sample collection
- `ResultVerificationResponse` — returned after result verification
- `ResultCertificationResponse` — returned after certification
- `LabReportResponse` + `LabReportResult` — full report with results array

---

## 8. Authentication & Authorization

### 8.1 Login Flow

```
POST /api/auth/login
Body: { "login_id": "admin", "password": "password123" }
       OR { "username": "admin", "password": "password123" }
```

1. Handler binds request, prefers `login_id`, falls back to `username`.
2. Calls `AuthService.Login()`.
3. Service fetches user from DB by `login_id OR username OR email`.
4. `bcrypt.CompareHashAndPassword` validates the password.
5. If user is `inactive`, login is rejected.
6. JWT is signed with `HS256` containing:
   - `sub` = `user_uuid`
   - `role` = `group_code` (or `role` field fallback)
   - `group_id` = numeric group ID
   - `iss`, `iat`, `exp`
7. Response includes the token, user info, and a permissions list.

### 8.2 JWT Validation (all secured routes)

`middleware.JWTAuth(secret)` — reads `Authorization: Bearer <token>`, parses and validates, then stores claims in Echo context:
- `c.Get("user_uuid")` — the user's UUID
- `c.Get("user_role")` — the role string
- `c.Get("user_group_id")` — the group ID

### 8.3 Role-Based Access Control

Two middleware layers work together:

**`RolePolicyGuard()`** — path-based rules applied to ALL secured routes:
- Any path ending in `/users` or containing `/users/` → requires `SUPER_ADMIN` or `ADMIN`
- Any path containing `/certify` → requires `DOCTOR`

**`RequireRole(...roles)`** — per-route inline role check:
- `GET /api/users` and `PATCH /api/users/:userId` → `SUPER_ADMIN` or `ADMIN` only

### 8.4 Roles and Permissions (as returned by Login)

| Role | Permissions |
|---|---|
| `SUPER_ADMIN` | PATIENT_READ, PATIENT_WRITE, BILLING_READ, BILLING_WRITE, LAB_VERIFY, LAB_CERTIFY, USER_MANAGE |
| `ADMIN` | PATIENT_READ, BILLING_WRITE, LAB_VERIFY, LAB_CERTIFY, USER_MANAGE |
| `DOCTOR` | PATIENT_READ, LAB_VERIFY, LAB_CERTIFY |
| `TECHNICIAN` | PATIENT_READ, LAB_VERIFY |
| *(any other)* | PATIENT_READ |

> The permissions array in the login response is informational only — the backend enforces access via middleware, not this list.

---

## 9. API Endpoints Reference

### Public Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/health` | None | Returns `{"status":"ok"}` |
| POST | `/api/auth/login` | None | Returns JWT token |

### Secured Endpoints (require `Authorization: Bearer <token>`)

#### Users

| Method | Path | Roles | Description |
|---|---|---|---|
| GET | `/api/users` | SUPER_ADMIN, ADMIN | List all users |
| PATCH | `/api/users/:userId` | SUPER_ADMIN, ADMIN | Update user (status, group, password) |

**PATCH /api/users/:userId body:**
```json
{
  "status": "Active",
  "account_group_id": 2,
  "password": "newplainpassword",
  "updated_by": "admin"
}
```

#### Patients

| Method | Path | Roles | Description |
|---|---|---|---|
| POST | `/api/patients/register` | Any authenticated | Register new patient |
| POST | `/api/patients` | Any authenticated | Alias for register |
| GET | `/api/patients` | Any authenticated | Search patients |
| GET | `/api/patients/search` | Any authenticated | Search patients (alias) |
| GET | `/api/patients/:id` | Any authenticated | Get patient history by UUID |
| GET | `/api/patients/:patientId/history` | Any authenticated | Get patient billing history |
| PATCH | `/api/patients/:patientId` | Any authenticated | Update patient profile |

**POST /api/patients/register body:**
```json
{
  "prefix": "Mr.",
  "first_name": "John Doe",
  "gender": "Male",
  "age": 25,
  "age_unit": "Yrs",
  "phone_no": "9876543210",
  "op_ip_no": "OP-1005",
  "patient_type": "Out Patients",
  "created_by": "USR-admin",
  "status": "Active"
}
```

**GET /api/patients?query=John** — searches by name, UUID, OP/IP number, phone, patient_no. Returns up to 100 results ordered by `created_at DESC`.

**Age unit normalization:** The handler normalizes `age_unit` input — `Y`/`year`/`years`/`Yrs` → `"Yrs"`, `M`/`month(s)` → `"Months"`, `D`/`day(s)` → `"Days"`.

#### Billing

| Method | Path | Roles | Description |
|---|---|---|---|
| POST | `/api/billing/new` | Any authenticated | Generate complete bill with services |
| POST | `/api/billing/bills` | Any authenticated | Create bare bill header only |
| POST | `/api/billing/bills/:id/items` | Any authenticated | Add a single item to a bill |
| GET | `/api/billing/services` | Any authenticated | List all available lab services |
| PATCH | `/api/billing/:billId/payment` | Any authenticated | Record payment against a bill |

**POST /api/billing/new body (main billing endpoint):**
```json
{
  "patient_id": "uuid-of-patient",
  "referred_by": "Dr. Smith",
  "hospital_name": "CPRT Hospital",
  "services": [
    { "service_id": "1", "service_name": "CBC", "rate": 350.00 },
    { "service_id": "2", "service_name": "LFT", "rate": 500.00 }
  ],
  "discount_amt": 50.00,
  "tax_percent": 0,
  "received_amt": 800.00
}
```

> `patient_id` here is the **patient UUID string**, not the integer ID.

**PATCH /api/billing/:billId/payment** — `:billId` is the bill's **integer ID** (not UUID):
```json
{ "received_amt": 200.00, "payment_mode": "CASH" }
```

**Payment status logic:**
- `balance <= 0` → `"Paid"`
- `balance > 0` → `"Pending"`
- `balance < 0` (overpaid by combined payments) → `"Overpaid"`

#### Orders

| Method | Path | Roles | Description |
|---|---|---|---|
| POST | `/api/orders` | Any authenticated | Create a lab order linked to a bill |
| PATCH | `/api/orders/:id/status` | Any authenticated | Update order status |

**POST /api/orders body:**
```json
{
  "bill_id": 5,
  "patient_id": 3,
  "status": "PENDING"
}
```

#### Lab Workflow

| Method | Path | Roles | Description |
|---|---|---|---|
| PATCH | `/api/lab/sample-collection/:billId` | Any authenticated | Mark sample collected for a bill |
| POST | `/api/lab/results/verify` | Any authenticated | Enter and verify test results |
| PATCH | `/api/lab/results/certify/:billId` | **DOCTOR only** | Certify (sign-off) results |
| GET | `/api/lab/reports/:billId` | Any authenticated | Fetch the final lab report |

**PATCH /api/lab/sample-collection/:billId body:**
```json
{
  "sample_no": "SAMP-001",
  "collected_by": "tech1",
  "worksheet_no": "WS-5"
}
```

**POST /api/lab/results/verify body:**
```json
{
  "bill_id": 5,
  "params": [
    { "param_id": "HB", "param_name": "Haemoglobin", "result_value": "13.5", "is_abnormal": false },
    { "param_id": "WBC", "param_name": "White Blood Cells", "result_value": "11.2", "is_abnormal": true }
  ],
  "verified_by": "tech1"
}
```

**PATCH /api/lab/results/certify/:billId body:**
```json
{
  "certified_by": "doctor1",
  "remarks": "Results reviewed and approved"
}
```

---

## 10. Module Deep-Dives

### 10.1 Auth Module

**Files:** `internal/service/auth.go`, `internal/http/handlers/auth.go`

- Login accepts `login_id` or `username` field in body (both acceptable).
- `bcrypt.CompareHashAndPassword` is used — never store or compare raw passwords.
- JWT claims include `role` = `group_code` from `account_groups`. This is the field that all RBAC checks use.
- The user `status` check uses `strings.EqualFold` — so `INACTIVE`, `Inactive`, `inactive` all block login.
- Response wraps everything in a single-element array (legacy API contract).

### 10.2 Users Module

**Files:** `internal/repository/postgres/users.go`, `internal/http/handlers/users.go`

- `GetAll` joins `users` with `account_groups` via `group_id`.
- `Update` supports partial updates — only non-nil fields are applied. Passing `password` (plaintext) in the PATCH request auto-hashes it via bcrypt before saving.
- The `userId` path parameter is a **string representation of the bigint ID** (`u.id::bigint`), not the UUID.
- User IDs in responses are prefixed with `USR-` (e.g., `USR-5`).

### 10.3 Patients Module

**Files:** `internal/repository/postgres/patients.go`, `internal/http/handlers/patients.go`

- `GET /api/patients/:id` actually returns **billing history**, not patient details — it calls `GetHistory`. This is a naming quirk.
- `GET /api/patients/:patientId/history` does the exact same thing (both exist).
- Search is full-text ILIKE on UUID, name, OP/IP number, phone, and patient number. Empty query returns all patients (up to 100).
- UpdateProfile uses COALESCE pattern — only fields with non-empty values are updated.

### 10.4 Billing Module

**Files:** `internal/repository/postgres/billing.go`, `internal/http/handlers/billing.go`

Two bill creation paths exist:

**Path 1 — `POST /api/billing/bills` then `POST /api/billing/bills/:id/items`:**
- Creates a bare bill header first, then adds items one-by-one. This is the two-step atomic approach.

**Path 2 — `POST /api/billing/new` (GenerateBill):**
- The main flow. Resolves `patient_uuid` → `patient_id`, calculates all totals, inserts bill header and all items in a single **transaction**. This is what should be used for production billing.

**Tax calculation formula:**
```
tax_amount = (total_amount - discount_amount) × tax_percent ÷ 100
net_amount = total_amount - discount_amount + tax_amount
balance_amount = net_amount - received_amount
```

**Payment Update (`PATCH /billing/:billId/payment`):**
- Adds to existing `received_amount` (cumulative, not replacing).
- Recalculates `balance_amount` and `payment_status`.

### 10.5 Orders Module

**Files:** `internal/repository/postgres/orders.go`, `internal/http/handlers/orders.go`

Simple CRUD. Orders link a `bill_id` to a `patient_id`. Order status values used in practice: `PENDING`, `IN_PROGRESS`, `COMPLETED`.

> Note: Sample collection in the Lab module auto-creates an order if one doesn't exist for the bill. So separate order creation is optional.

### 10.6 Lab Workflow Module

**Files:** `internal/repository/postgres/lab.go`, `internal/http/handlers/lab.go`

This is the most complex module. See Section 11 for the full flow.

**Key implementation details:**
- `MarkSampleCollected` — finds (or creates) an order for the bill, then upserts a sample record using `barcode` as the unique key. Sets `lab_bills.report_status = 'Collected'`.
- `VerifyResults` — loops through params and upserts each row in `lab_results` with `ON CONFLICT(bill_id, param_id) DO UPDATE`. Sets `lab_bills.report_status = 'Verified'`.
- `CertifyResults` — updates `lab_bills` directly: sets `report_status = 'Certified'`, `dispatch_ready = TRUE`, stores certifier and remarks.
- `GetReport` — JOINs `lab_bills` + `patients` + `lab_results` to produce the final report. Abnormal flag `Y` → displayed as `H` (High) in report output.

---

## 11. Lab Workflow — End-to-End Business Flow

```
Step 1: Register Patient
  POST /api/patients/register
  → Returns patient_uuid (e.g. "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")
  → Save this UUID

Step 2: Generate Bill
  POST /api/billing/new
  Body: { patient_id: <patient_uuid>, services: [...], received_amt: ... }
  → Returns bill_id (UUID string) and bill_no (integer)
  → Save the INTEGER bill_no (used as :billId in all lab routes)
  → lab_bills.report_status = 'Pending'

Step 3: Sample Collection
  PATCH /api/lab/sample-collection/:billId
  Body: { sample_no: "SAMP-001", collected_by: "tech1" }
  → lab_orders record created/found
  → samples record upserted
  → lab_bills.report_status = 'Collected'

Step 4: Result Verification
  POST /api/lab/results/verify
  Body: { bill_id: <integer>, params: [...], verified_by: "tech1" }
  → lab_results rows upserted (one per param)
  → lab_bills.report_status = 'Verified'

Step 5: Result Certification  ← DOCTOR role required
  PATCH /api/lab/results/certify/:billId
  Body: { certified_by: "doctor1", remarks: "..." }
  → lab_bills.report_status = 'Certified'
  → lab_bills.dispatch_ready = TRUE

Step 6: Get Report
  GET /api/lab/reports/:billId
  → Returns full report: patient info, verifier, certifier, all params with values and flags
```

> **Important:** All `/:billId` route parameters in the Lab and Billing sections use the **integer `lab_bills.id`** (the `bill_no` returned in the GenerateBill response), **not** the UUID string.

---

## 12. Code Layering Pattern

Understanding how a new feature should be added:

### Adding a new endpoint (example: GET patient demographics)

1. **Domain:** Add struct to `internal/domain/models.go` if new data shape is needed.
2. **Repository Interface:** Add method signature to `internal/repository/interfaces.go` → `PatientRepository`.
3. **Repository Implementation:** Implement SQL in `internal/repository/postgres/patients.go`.
4. **Service:** Add pass-through method in `internal/service/patients.go`.
5. **Handler Contract:** Add method to `PatientService` interface in `internal/http/handlers/contracts.go`.
6. **Handler:** Add handler method in `internal/http/handlers/patients.go`.
7. **Router:** Register route in `registerAPIRoutes()` in `internal/http/router.go`.

### Connection pool configuration

Set in `internal/db/postgres.go`:
- `MaxConns = 20`
- `MinConns = 2`
- `MaxConnLifetime = 1 hour`
- Connection timeout = 10 seconds on startup

---

## 13. Seeded Test Data

After running `seed_users.sql`, the following accounts are available. All passwords are `password123`.

| Login ID | Role | Group Code |
|---|---|---|
| `superadmin` | Super Admin | `SUPER_ADMIN` |
| `admin` | Admin | `ADMIN` |
| `doctor1` | Doctor | `DOCTOR` |
| `receptionist1` | Receptionist | `RECEPTIONIST` |
| `tech1` | Technician | `TECHNICIAN` |

Two seed patients are inserted (idempotent, checked by `op_ip_no`):
- **John Doe** — `OP-1002`, Male, 25 Yrs, phone: 9876543210
- **Priya Sharma** — `OP-1003`, Female, 31 Yrs, phone: 9123456780

---

## 14. Known Quirks & Important Notes

### ⚠️ Bill ID vs Bill UUID
The `:billId` parameter in routes like `/lab/sample-collection/:billId`, `/billing/:billId/payment`, `/lab/results/certify/:billId`, and `/lab/reports/:billId` is the **integer `lab_bills.id`** (bigint), NOT the UUID string returned in `bill_id` field of the GenerateBill response. The UUID is in `bill_uuid` / `bill_id` (string) in responses; the integer is in `bill_no`.

### ⚠️ `GET /api/patients/:id` Returns History, Not Profile
This route calls `GetHistory`, returning billing history items. If you want patient profile/demographics, use the search endpoint.

### ⚠️ Payment Update is Additive
`PATCH /billing/:billId/payment` **adds** the `received_amt` to the existing received amount — it does not replace it. Sending the same payment twice will double-count it.

### ⚠️ `collected_by` in Sample Table
The `samples.collected_by` column is a `BIGINT` referencing `users.id`, but the API accepts a plain text string (username). The repository inserts the `sample_no` as the barcode, not `collectedBy`. The `collectedBy` value is only echoed back in the response but not stored to the DB. This is a gap.

### ⚠️ Worksheet Generation
Worksheet numbers (`WS-{n}`) are generated as `MAX(id)+1` from the worksheets table but are not written back to the worksheets table by the sample collection flow. The worksheet number is returned in the API response and can be optionally overridden by passing `worksheet_no` in the request body.

### ⚠️ GORM in go.mod but Not Used
`gorm.io/gorm` is listed as a dependency but no application code uses GORM. All DB access is raw SQL via `pgx/pgxpool`. It can be removed from `go.mod` safely.

### ⚠️ Tax Field Dual Meaning
In `domain.LabBill`, the `Tax` field is used as **tax percentage** when passed into `GenerateBill` (input), but stored as the **calculated tax amount** in the DB and returned as the amount in responses. This is resolved inside `GenerateBill` in the billing repository.

### ⚠️ No Pagination on Most List Endpoints
Patient search is hard-limited to 100 rows. Users list, services list, and history have no pagination. Large datasets will need pagination added.

### ⚠️ Migrations Are Not Automated
There is no migration tool (like `golang-migrate`). Migrations must be run manually in order via psql or pgAdmin. Running 005 after 002 and 003 is safe (all ALTERs use `IF NOT EXISTS`), but do not skip 006 or 007.

---

*End of Knowledge Transfer Document*
