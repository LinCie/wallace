# Wallace — Project Guide

## What this project is

Wallace is an ERP (Enterprise Resource Planning) and POS (Point of Sale) backend. It is a Go HTTP API backed by PostgreSQL, organized into domain modules with handler, service, and repository layers.

The current implementation provides authentication, users, organizations and membership, and organization-scoped product management. Other ERP/POS capabilities are project goals, not implemented modules yet. There is no frontend application in this repository; the HTML under `internal/docs` is an API reference page.

Core conventions:

- Use **UUID v7** for entity IDs, generated through `internal/database.GenerateID()`.
- Treat the **organization as the tenant boundary** for business data.
- Users are global identities and can belong to multiple organizations through `organization_members`.
- Carry validated user and organization IDs through request context, then pass organization IDs explicitly into business-data operations.
- Keep HTTP handling, business rules, and SQL in their respective layers.

## Stack

- Go; `go.mod` is the source of truth for the required Go version and dependencies.
- `net/http` and Chi for routing and middleware.
- PostgreSQL with `sqlx` and the `pgx` driver; SQL is written in repositories, not an ORM.
- Goose for SQL migrations.
- `go-playground/validator` for DTO validation.
- HS256 JWT access tokens, Argon2id password hashing, and rotating opaque refresh tokens.
- Air for local live reload; Docker Compose for local PostgreSQL.

## Folder structure

```text
cmd/
  server/main.go          HTTP server entry point
  migrate/main.go         Database migration CLI
internal/
  config/                 Environment-based configuration
  database/               Connection setup, UUID generation, PostgreSQL error codes
  docs/                   Embedded OpenAPI specification and Scalar API reference
  httpx/                  Shared HTTP binding, validation, and response helpers
  jwt/                    Access-token generation and validation
  middleware/             Authentication, organization access, and CORS
  modules/
    auth/                 Registration, login, refresh, logout, password/token helpers
    organizations/        Organizations and membership-based access
    products/             Organization-scoped product CRUD
    users/                Global user persistence and service used by auth
  router/                 Routes, middleware ordering, and module composition
migrations/               Timestamped Goose SQL schema migrations
tmp/                      Local Air build output (not application source)
```

## Root files

| File | Responsibility |
| --- | --- |
| `AGENTS.md` | Project overview, architecture, and contributor guidance; keep aligned with the implementation. |
| `go.mod`, `go.sum` | Go module definition, dependency versions, and checksums. |
| `.env.example` | Configuration template; document environment variables without real secrets. |
| `.env` | Local environment values; never commit or copy secrets into documentation. |
| `.air.toml` | Live-reload configuration; builds `cmd/server` into `tmp/main` and loads `.env`. |
| `compose.yaml` | Local PostgreSQL service (`database`) and persistent volume. |
| `Makefile` | Development and migration shortcuts; check targets against actual entry points before using. |
| `.gitignore` | Exclusions for local configuration, build output, and other untracked artifacts. |

## Entry points and shared files

| File | What belongs here |
| --- | --- |
| `cmd/server/main.go` | Load configuration, open/close the database, construct the router, and start the HTTP server. No domain logic. |
| `cmd/migrate/main.go` | Goose CLI setup and dispatch for `up`, `down`, `status`, `reset`, and `version`. |
| `internal/config/config.go` | Configuration type, environment parsing, defaults, and required-value checks. |
| `internal/database/database.go` | PostgreSQL connection creation and connectivity check. |
| `internal/database/id.go` | Shared UUID v7 generation. |
| `internal/database/errors.go` | Shared PostgreSQL error-code constants. |
| `internal/router/router.go` | Construct modules, register routes, and apply middleware at the appropriate scope. |
| `internal/middleware/auth.go` | Bearer authentication, organization membership checks, and typed context accessors. |
| `internal/middleware/cors.go` | Allowed origin, credentials, headers, methods, and preflight handling. |
| `internal/jwt/jwt.go` | JWT claims, signing, expiration, and validation; user ID is the token subject. |
| `internal/httpx/body.go` | Bind a single JSON body and validate DTO tags. |
| `internal/httpx/query.go` | Bind query parameters and validate DTO tags. |
| `internal/httpx/response.go` | Consistent JSON success and error responses. |
| `internal/httpx/constructor.go` | Shared HTTP middleware function type. |
| `internal/docs/docs.go` | Embed and serve the API reference and OpenAPI document. |
| `internal/docs/index.html` | Scalar API reference UI pointing to `/openapi.json`. |
| `internal/docs/openapi.json` | API paths, schemas, security requirements, and response examples; update with API changes. |

## Domain-module file responsibilities

Use the existing module layout when adding or extending a domain. Not every module needs every file; do not create empty layers just for symmetry.

| File | What belongs here |
| --- | --- |
| `entity.go` | Domain/persistence structs, database tags, enums, and domain constants. Keep separate from transport DTOs. |
| `dto.go` | HTTP request, query, and response structs with `json`, `query`, and `validate` tags. |
| `error.go` | Module-specific sentinel errors shared by services, repositories, and handlers. |
| `handler.go` | Service interface consumed by HTTP handlers; parse/validate input, read context, map DTOs, and map errors to HTTP statuses. Do not put SQL here. |
| `service.go` | Repository interface consumed by the service, business rules, normalization, pagination logic, and orchestration. Keep independent of HTTP requests/responses. |
| `repository.go` | PostgreSQL implementation, parameterized SQL, transactions, organization filters, and database-to-domain error mapping. |
| `module.go` | Wire dependencies and expose the service and, where applicable, handler to the router. |

Current module-specific details:

- **`products`** has all seven standard files. It supports create, retrieve, list, update, and soft delete. Status values are `active` and `inactive`. List pagination defaults to page 1 / limit 20, with a maximum limit of 100.
- **`organizations`** has all seven standard files. Creation inserts the organization and the creator's `owner` membership in one transaction. Retrieve/list operations return only organizations the caller belongs to. Membership roles are `owner` and `member`; current product access checks membership, not role-specific permissions.
- **`users`** has `entity.go`, `error.go`, `handler.go`, `service.go`, `repository.go`, and `module.go`. Its `handler.go` currently only declares the service interface; there are no user HTTP endpoints or DTO file. Auth consumes its service for user creation and lookup by email.
- **`auth`** has `dto.go`, `entity.go`, `error.go`, `handler.go`, `service.go`, and `module.go`, but no SQL repository. `entity.go` aliases the users module's `User` type. Auth uses the users service and a refresh-session store.

Auth helper files:

| File | Responsibility |
| --- | --- |
| `auth/password.go` | Argon2id password hashing/verification and bounded hashing concurrency. |
| `auth/token.go` | Internal token result, random refresh-token generation, and SHA-256 refresh-token digests. |
| `auth/refresh_store.go` | Refresh-session type and storage interface. |
| `auth/refresh_memory.go` | Mutex-protected in-memory session storage, expiry handling, rotation, and deletion. |

Refresh sessions are currently process-local: restarting the server loses them, and separate server instances do not share them. Do not describe this store as persistent.

## Application and request flow

### Startup

```text
cmd/server
  → config.NewConfig()
  → database.NewDB()
  → router.NewRouter()
  → module constructors wire repositories → services → handlers
  → HTTP server listens
```

Global middleware strips trailing slashes, logs requests, recovers panics, and handles CORS. `/docs` and `/openapi.json` are public; application routes are under `/v1`.

### Authentication

1. `POST /v1/auth/register` validates input, hashes the password, and creates a global user. Registration does not log the user in.
2. `POST /v1/auth/login` verifies credentials and issues a short-lived access JWT plus an opaque refresh token.
3. The access token is returned in JSON; the refresh token is delivered only in an HttpOnly cookie scoped to `/v1/auth`. The store holds its digest, not the raw token.
4. `POST /v1/auth/refresh` reads the cookie and rotates the refresh session. `POST /v1/auth/logout` removes the session and clears the cookie.
5. Protected routes require `Authorization: Bearer <access-token>`. `WithAuth` validates the token and adds the user UUID to context.

### Organizations and tenant-scoped data

1. Authenticated callers create/list organizations at `/v1/organizations` or retrieve one at `/v1/organizations/{organizationID}`. These routes use the authenticated user ID rather than requiring an active-organization header.
2. Product requests to `/v1/products` require `X-Organization-ID: <organization-uuid>` as well as the bearer token.
3. `WithOrganizationAccess` parses the header, verifies the user's membership, and stores the organization UUID in context. Non-members receive a not-found response.
4. The handler reads the validated organization ID from context, binds input, and calls the service.
5. The service applies business rules and delegates to the repository.
6. The repository writes `organization_id` on creation and filters reads, updates, and deletes by both organization and resource ID where applicable.
7. The handler maps the result to a response DTO and uses `httpx` to write JSON.

```text
Request → router → auth → organization access → handler → service → repository → PostgreSQL
Response ← HTTP/DTO mapping ← domain result / error ← repository
```

Success JSON is wrapped as `{ "data": ... }`; errors use `{ "code": "...", "message": "..." }`. No-content responses have no JSON body.

## ID and data-isolation rules

- Use `github.com/gofrs/uuid/v5` UUID values in Go and PostgreSQL `UUID` columns. Generate new entity IDs with `database.GenerateID()`; do not introduce integer IDs or a different UUID version.
- Product creation optionally accepts a client-provided ID, which must be UUID v7 with the RFC 4122 variant. When omitted, the repository generates it. This is not an indication that every endpoint accepts caller-supplied IDs.
- A UUID identifies a record; it does **not** authorize access. Always enforce membership and organization scope.
- For organization-owned business records, derive `organization_id` from validated context, not an untrusted request-body field. Pass it explicitly through service/repository boundaries and include it in SQL filters.
- Scope tenant-local uniqueness to the organization. Product slugs are unique per organization among non-deleted products; user emails are global among non-deleted users.
- Products and users have `deleted_at`; normal queries exclude deleted rows. Product deletion updates timestamps rather than removing rows.
- `organization_members` uses a composite `(organization_id, user_id)` primary key rather than a separate generated ID.
- Preserve transactional consistency for multi-write operations such as organization creation plus owner membership.

## Migrations

Files use timestamped names and Goose `-- +goose Up` / `-- +goose Down` sections. All migration creation and execution must use the defined commands in `Makefile`; do not invoke Goose or `cmd/migrate` directly.

- Create: `make migration name=<descriptive_name>`; then edit the generated SQL file.
- Apply: `make migrate-up`.
- Roll back the latest migration: `make migrate-down`.
- Inspect: `make migrate-status` or `make migrate-version`.
- Reset: `make migrate-reset` (destructive; use only when explicitly requested).

| File | Schema responsibility |
| --- | --- |
| `20261008092604_create_products_table.sql` | Initial products table, timestamps, soft deletion, and indexes. |
| `20261008092605_create_users_table.sql` | Users table, password hashes, timestamps, soft deletion, and email uniqueness. |
| `20261008092606_create_organizations_table.sql` | Organizations, memberships, product organization ownership, and per-organization slug uniqueness. |

Read migrations in order: the initial product schema alone does not represent the current schema. Add schema changes as new migrations and keep repository SQL, entities, DTOs, and API documentation consistent.

## Local development

- Copy `.env.example` to `.env` and configure values. `DATABASE_URL` and `JWT_SECRET` are required by the server; there is no fallback JWT secret in `config.NewConfig()`.
- `docker compose up -d database` starts the configured PostgreSQL service.
- Export the required environment variables before running `go run ./cmd/server`; the server entry point does not load `.env` itself.
- `make dev` runs Air, which loads `.env`. Migration targets in the Makefile also load `.env`.
- Use `make migrate-up` to apply migrations and `make migrate-status` to inspect their status; `MIGRATIONS_DIR` defaults to `migrations`.
- For plain-HTTP local development only, set `COOKIE_SECURE=false`. Keep secure cookies and a strong JWT secret outside local development.
- Some Makefile shortcuts do not match the current tree: build/start reference `cmd/web` / `bin/web`, seed references an absent `cmd/seed`, and PostgreSQL shortcuts name `postgres` instead of Compose's `database`. Use the actual entry points/service above rather than assuming these targets work.

## Guidance for changes

- Follow the nearest existing domain-module pattern and keep changes within the requested scope.
- Put shared HTTP, database, and authentication behavior in the existing shared packages; avoid duplicating it across handlers.
- Preserve organization isolation across every business-data operation, including lists, updates, deletes, and future ERP/POS modules.
- Register new endpoints in `internal/router/router.go`; update `internal/docs/openapi.json` when their contract changes.
- Never expose password hashes or raw refresh tokens in response DTOs, logs, or documentation.
- Do not add dependencies, layers, or speculative ERP/POS modules unless the requested feature needs them.
- Do not run tests unless explicitly requested or required by applicable instructions. For documentation-only changes, check paths and claims against source; no test run is needed.
