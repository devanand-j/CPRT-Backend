# CPRT LIS Flow Diagrams

This file is the quick visual companion to the main KT document. It shows how the application is structured, how requests move through the code, and how the core lab workflow moves from patient registration to final report generation.

## 1. Application Architecture

```mermaid
flowchart LR
    A[Client / Postman / Frontend] --> B[Echo Router]
    B --> C[Middleware<br/>Logger, Recover, CORS]
    C --> D[Auth Middleware<br/>JWTAuth]
    D --> E[Role Checks<br/>RolePolicyGuard / RequireRole]
    E --> F[Handlers]
    F --> G[Services]
    G --> H[Postgres Repositories]
    H --> I[(PostgreSQL)]
```

The code follows a clean layered structure. Handlers deal with HTTP, services coordinate business operations, repositories hold raw SQL, and PostgreSQL is the single source of truth.

## 2. Request Flow Inside the Code

```mermaid
sequenceDiagram
    participant U as User/Client
    participant R as Router
    participant M as Middleware
    participant H as Handler
    participant S as Service
    participant P as Postgres Repository
    participant D as Database

    U->>R: HTTP Request
    R->>M: Apply middleware
    M->>M: Validate JWT / role
    M->>H: Forward request
    H->>H: Bind and validate payload
    H->>S: Call service method
    S->>P: Call repository method
    P->>D: Execute SQL
    D-->>P: Rows / result
    P-->>S: Domain model / response data
    S-->>H: Business result
    H-->>U: JSON response
```

Most business rules are implemented close to the handler and repository layers. Services are intentionally thin, so most debugging usually starts in the handler for input issues or in the repository for SQL and data issues.

## 3. Authentication Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant A as Auth Handler
    participant S as Auth Service
    participant U as User Repository
    participant DB as PostgreSQL

    C->>A: POST /api/auth/login
    A->>S: Login(login_id/username, password)
    S->>U: GetByUsername()
    U->>DB: SELECT user + account_group
    DB-->>U: User row
    U-->>S: User model
    S->>S: Compare bcrypt password
    S->>S: Build JWT claims
    S-->>A: token + user
    A-->>C: JSON response with token
```

Login supports `login_id`, `username`, or email lookup in the repository. After successful bcrypt verification, the service issues an HS256 JWT containing the user UUID, role, and group ID.

## 4. Patient to Report Workflow

```mermaid
flowchart TD
    A[Register Patient] --> B[Generate Bill]
    B --> C[Create or Reuse Lab Order]
    C --> D[Collect Sample]
    D --> E[Verify Results]
    E --> F[Certify Results]
    F --> G[Fetch Final Report]
```

This is the main business flow of the project. Almost every operational use case falls somewhere in this chain, so this is the best mental model for onboarding someone new.

## 5. Billing Flow

```mermaid
flowchart TD
    A[Receive patient UUID and selected services] --> B[Resolve patient UUID to patient_id]
    B --> C[Calculate total amount]
    C --> D[Apply discount]
    D --> E[Calculate tax amount]
    E --> F[Compute net amount]
    F --> G[Compute received and balance]
    G --> H[Set payment status]
    H --> I[Insert lab_bills record]
    I --> J[Insert lab_bill_items records]
```

The `GenerateBill` path is the main billing implementation. It runs in a transaction and calculates totals before storing both the bill header and bill items.

## 6. Lab Sample Collection Flow

```mermaid
flowchart TD
    A[Sample collection request] --> B[Find bill by integer bill id]
    B --> C[Find existing lab order]
    C -->|Not found| D[Create new lab order]
    C -->|Found| E[Reuse existing order]
    D --> F[Upsert sample by barcode]
    E --> F
    F --> G[Update bill report_status to Collected]
    G --> H[Return sample response]
```

Sample collection is resilient to missing orders because it auto-creates a `lab_orders` row when needed. The sample is upserted using the barcode, which means repeating the same sample number updates the existing row instead of duplicating it.

## 7. Result Verification and Certification Flow

```mermaid
flowchart TD
    A[Verify results request] --> B[Loop through params]
    B --> C[Upsert into lab_results by bill_id + param_id]
    C --> D[Mark bill report_status as Verified]
    D --> E[Doctor certifies result]
    E --> F[Update lab_bills]
    F --> G[Set report_status = Certified]
    G --> H[Set dispatch_ready = true]
```

Verification stores or updates individual result parameters, while certification is the doctor sign-off step that marks the bill as ready for dispatch. The unique index on `lab_results(bill_id, param_id)` is what makes repeated verification idempotent per parameter.

## 8. Database Relationship View

```mermaid
erDiagram
    ACCOUNT_GROUPS ||--o{ USERS : has
    PATIENTS ||--o{ LAB_BILLS : owns
    PATIENTS ||--o{ LAB_ORDERS : has
    LAB_BILLS ||--o{ LAB_BILL_ITEMS : contains
    LAB_BILLS ||--o{ LAB_RESULTS : produces
    LAB_BILLS ||--o{ LAB_ORDERS : generates
    LAB_ORDERS ||--o{ SAMPLES : produces
    LAB_SERVICES ||--o{ LAB_BILL_ITEMS : referenced_by
```

The operational center of the schema is `lab_bills`. It connects patient registration, billing, ordering, sample handling, and result reporting into one traceable workflow.
