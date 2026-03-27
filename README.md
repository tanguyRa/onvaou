# saas_seed

A full-stack SaaS starter kit with:
- SvelteKit frontend (`front`)
- Go HTTP API (`back`)
- Postgres schema + SQLC-generated repository layer (`db`)
- Better Auth (email/password + JWT)
- Polar billing integration (products, checkout, portal, webhooks)
- Optional LLM provider abstraction (Gemini/Anthropic)

## What It Does Today

At the moment, the project ships a solid foundation for:
- marketing page + pricing page
- registration/login flows
- protected app shell with profile + billing settings
- checkout creation and customer portal redirects through Polar
- webhook ingestion from Polar into local `subscription`/`events` tables
- authenticated proxying from SvelteKit to Go API with JWT forwarding

The Go API surface is intentionally small right now:
- `GET /api/ping`
- `GET /api/secured/ping` (reads authenticated user from JWT)
- `POST /webhooks/polar`

## Repository Layout

```text
.
├── back/          # Go API (handlers, middleware, config, llm, storage)
├── front/         # SvelteKit app (routes, auth client, payment APIs, UI)
├── db/            # SQL migrations, SQL queries, sqlc config
├── compose.yml    # local docker stack
└── Makefile       # common developer commands
```

## Architecture

### Request Flow (Protected API call)
1. Browser calls SvelteKit endpoint under `front/src/routes/api/...`.
2. SvelteKit gets a JWT from Better Auth cookies (`auth.api.getToken`).
3. SvelteKit proxies to Go API (`GO_API_URL`) with `Authorization: Bearer <jwt>`.
4. Go middleware parses JWT and injects `session.UserInfo` in context.
5. Go middleware acquires DB connection and sets Postgres session var `app.user_id`.
6. SQL queries execute under Postgres RLS policies.

### Billing Flow
1. Frontend fetches products from `/api/payments/products`.
2. Frontend posts selected slug to `/api/payments/checkout`.
3. Server creates Polar checkout and returns redirect URL.
4. Polar webhook hits `POST /webhooks/polar`.
5. Go handler verifies signature and upserts local subscription/event state.

## Tech Stack

- Frontend: SvelteKit 2, Svelte 5, TypeScript, Bun adapter
- Auth: Better Auth + JWT plugin + SvelteKit handler
- Payments: Polar SDK + `@polar-sh/better-auth` + `@polar-sh/sveltekit`
- Backend: Go 1.24+, `net/http`, `pgx/v5`, `sqlc`
- DB: PostgreSQL 16, migrations + Row Level Security
- Storage: MinIO adapter available (not yet wired into routes)

## Local Development

### Prerequisites
- Docker + Docker Compose
- `make`
- 1Password CLI (`op`) for the default secure compose workflow

### Environment
1. Start from `.env.example` and fill `.env.dev`.
2. Required minimum for end-to-end:
   - Postgres credentials
   - `BETTER_AUTH_SECRET`
   - `DATABASE_URL` (resolved via compose env in this setup)
3. For billing:
   - `PAYMENT_PROVIDER=polar`
   - `POLAR_ACCESS_TOKEN`
   - `POLAR_WEBHOOK_SECRET`

### Run

```bash
make start
```

This builds and runs:
- frontend on `127.0.0.1:3000`
- backend on `127.0.0.1:8080`
- postgres on `127.0.0.1:5432`

And applies migrations automatically.

### Useful Commands

```bash
make logs
make stop
make clean
make migrate-up
make migrate-down N=1
make sqlc
```

## Database Model (Current)

### Auth tables
- `user`
- `session`
- `account`
- `verification`
- `jwks`

### Payments tables
- `subscription` (1 row per user, tracks plan/status/period end)
- `events` (stores webhook/event payloads)

### Security model
- Row Level Security is enabled across auth and payment tables.
- Policies use `current_setting('app.user_id', true)` and optional internal bypass flag.
- Go middleware sets `app.user_id` per request from authenticated JWT subject.

## Frontend Routes (Current)

Public:
- `/`
- `/pricing`
- `/login`
- `/register`

Protected:
- `/app`
- `/settings/profile`
- `/settings/billing`
- `/settings/portal` (server route redirecting to Polar portal)

API routes in frontend:
- `/api/auth/[...all]` (Better Auth handler)
- `/api/[...path]` (proxy to Go API + JWT forwarding)
- `/api/payments/products`
- `/api/payments/checkout`

## Developer Philosophy

The codebase follows a pragmatic “foundation first” approach:
- Keep runtime surfaces small and explicit.
- Favor composition over framework magic (plain `net/http`, explicit middleware).
- Keep DB access typed and generated (`sqlc`) instead of ad-hoc SQL in handlers.
- Enforce tenant boundaries in the DB (RLS), not only in application code.
- Build provider abstractions early (LLM/payment/storage) so integrations can swap.
- Make behavior testable at package level (middleware, handlers, llm providers).

## Testing

Backend:

```bash
cd back
go test ./...
```

Frontend static checks:

```bash
cd front
npm run check
```

## Contributing Tips

- For new backend endpoints:
  1. Add route in `back/internal/server/routes.go`
  2. Add middleware chain as needed
  3. Implement handler in `back/internal/handlers`
  4. Add SQL in `db/queries/*.sql`, regenerate with `make sqlc`
  5. Add/adjust migrations in `db/migrations` when schema changes
- Preserve RLS assumptions whenever adding new tables and queries.
- Keep auth context and DB session wiring intact for multi-tenant safety.

