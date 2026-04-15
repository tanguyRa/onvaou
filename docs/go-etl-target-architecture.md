# Go ETL Target Architecture

## Purpose

Define the target architecture and functional requirements for moving the event ingestion pipeline from the Python `geo` service to Go, while keeping the Python `geo` service for geospatial search and Python-native geo/data processing.

This document is intended to be implementation-ready for a developer.

## Problem Statement

The current Python `geo` service owns two different responsibilities:

- Public geo/event APIs used by the frontend
- Private ETL ingestion jobs that fetch, crawl, normalize, deduplicate, and persist event data

Those concerns have different runtime characteristics:

- Search APIs are request-driven and benefit from Python geo libraries already used by `geo`
- ETL is mostly network IO, scheduling, JSON/CSV parsing, crawling orchestration, and database writes

The ETL part is a better fit for Go and should move into the Go codebase. The Python `geo` service should remain focused on event search, city resolution, and geospatial processing.

## Goals

- Move ingestion orchestration from Python to Go
- Keep the Python `geo` service in place for search and geo-specific processing
- Preserve the existing database schema and frontend behavior during migration
- Make Go the primary owner of writing event-source data into Postgres
- Keep replay/debugability as a first-class part of ingestion
- Support API-based sources and web crawlers under a common ingestion framework

## Non-Goals

- Do not remove the Python `geo` service
- Do not migrate public search endpoints as part of this work
- Do not redesign the event schema unless required for ingestion correctness
- Do not change the frontend contract for event search in this phase
- Do not require all crawlers to be rewritten immediately if a Python implementation is still temporarily needed

## Current State

### Python `geo` service currently owns

- Search endpoints: `/events`, `/events/{event_id}`, `/cities/search`
- Ingestion triggers: `/ingestion/openagenda`, `/ingestion/datagouv`, replay/debug routes
- Scheduled jobs via APScheduler
- Source harvesters for OpenAgenda and data.gouv
- Deduplication and event upsert logic
- BAN geocoding for missing coordinates

### Go backend currently owns

- Main backend API
- Auth/session/payment flows
- Shared Postgres access patterns via `pgx` and `sqlc`

## Target State

### Service boundaries

#### Go backend

Owns:

- Event ingestion orchestration
- Source fetchers and crawlers
- Event normalization
- Deduplication
- Event upsert and source hash persistence
- Ingestion scheduling
- Ingestion admin endpoints or internal triggers
- Ingestion debug artifacts and replay support

Does not own:

- Public event search API
- BAN-backed city autocomplete endpoint for frontend use
- Python-native geospatial transformations that materially benefit from Python libraries

#### Python `geo` service

Owns:

- Public geo/event search endpoints
- BAN-based city lookup for UI search
- Geospatial processing that depends on Python geo libraries
- Optional downstream enrichment jobs that are truly geo-library-heavy

Does not own:

- Scheduled source ingestion from event providers
- Event write-path orchestration into the DB

## Proposed Repository Layout

Add Go ingestion code under `back/`:

```text
back/
  cmd/
    server/                 # existing HTTP API
    ingest/                 # new worker entrypoint
  internal/
    ingestion/
      sources/
        openagenda/
        datagouv/
        mairie/
      normalize/
      dedup/
      store/
      scheduler/
      replay/
      model/
      admin/
```

### Package responsibilities

- `internal/ingestion/model`
  - Canonical ingestion models independent from HTTP/database rows
- `internal/ingestion/sources/openagenda`
  - Agenda discovery
  - Event pagination
  - Response parsing
- `internal/ingestion/sources/datagouv`
  - Dataset search
  - Resource parsing for CSV/JSON/GeoJSON
- `internal/ingestion/sources/mairie`
  - Web crawling/scraping source integrations
- `internal/ingestion/normalize`
  - Event normalization rules
  - Address assembly
  - Source field mapping
- `internal/ingestion/dedup`
  - Exact hash duplicate detection
  - Near-duplicate detection policy
- `internal/ingestion/store`
  - Postgres persistence
  - Event upsert
  - Source-hash writes
  - Agenda metadata persistence if retained
- `internal/ingestion/scheduler`
  - Cron registration and execution
- `internal/ingestion/replay`
  - Raw payload artifact persistence
  - Replay command path
- `internal/ingestion/admin`
  - Optional internal/admin HTTP endpoints to trigger jobs or inspect status

## Runtime Architecture

### Ingestion worker

ETL should run in a dedicated Go process, not inside the request-serving API path by default.

Recommended runtime:

- `back/cmd/server`: regular API server
- `back/cmd/ingest`: ingestion worker process

The worker process should:

- Load the shared app config
- Open DB connectivity
- Register scheduled jobs
- Run manual one-shot jobs from CLI flags
- Optionally expose a small private/admin HTTP surface for health and trigger endpoints

### Why separate process

- Isolates ingestion failures from user-facing API traffic
- Allows different scaling/restart policies
- Makes cron execution explicit
- Avoids coupling long-running jobs to request-serving lifecycle

## Functional Requirements

## FR-1 Source ingestion ownership

The Go worker must become the primary runtime for all event-source ingestion into the shared Postgres database.

This includes:

- OpenAgenda
- data.gouv
- Mairie/web crawlers added in future phases
- Additional API/crawl sources later

## FR-2 Source execution model

Each source integration must implement a common contract:

- `Discover` if the source has catalogs or agendas
- `Fetch` raw source payloads
- `Normalize` raw payloads into canonical event models
- `Persist` normalized events
- `Report` structured result metrics

Minimum result metrics:

- `fetched`
- `inserted`
- `updated`
- `duplicates`
- `failed`
- source-specific details

## FR-3 OpenAgenda ingestion

Go must support the equivalent of the current Python behavior:

- Fetch agenda catalog
- Detect changed agendas
- Persist agenda metadata if that schema remains in use
- Fetch paginated agenda events
- Skip unchanged event batches using batch hashes
- Normalize events
- Upsert events into `events`
- Mark agenda sync success/failure

## FR-4 data.gouv ingestion

Go must support:

- Dataset search via configured queries
- Filtering relevant datasets/resources
- Parsing CSV, JSON, and GeoJSON resources
- Per-record normalization into canonical event models
- Skipping malformed records while preserving job continuity

## FR-5 Crawling support

The Go ETL architecture must support web crawling as a first-class source type.

Requirements:

- A source can be API-driven or crawler-driven
- Crawlers must support pagination and backoff
- Per-page and per-item failures must be isolated
- Raw fetched pages or parsed payloads must be capturable as artifacts

Implementation note:

- If a specific crawler truly depends on Python or Playwright in a way that is not worth porting immediately, the Go worker may temporarily call an external helper or preserve a narrow bridge. That bridge should be explicit and temporary.

## FR-6 Canonical event normalization

The Go worker must normalize all sources into a canonical event shape equivalent to what the current DB expects:

- `source_uid`
- `title`
- `description`
- `start_dt`
- `end_dt`
- `location_name`
- `address`
- `longitude`
- `latitude`
- `source_tag`
- `source_url`

Normalization requirements:

- Reject records missing minimum required fields
- Parse dates robustly across known formats
- Build a normalized address from source fragments
- Use source-provided coordinates when available
- Resolve missing coordinates via BAN or equivalent geocoder

## FR-7 Deduplication

The Go worker must preserve current dedup semantics:

- Exact duplicates detected via source/content hash
- Near duplicates detected within a bounded time window using title/address similarity
- Duplicate decisions must happen before final upsert

Behavioral requirement:

- Ported dedup logic should produce materially equivalent results to the current Python behavior unless a deliberate change is documented

## FR-8 Persistence

The Go worker must write to the existing shared Postgres/PostGIS schema.

Requirements:

- Upsert into `events`
- Maintain `source_hashes`
- Preserve current geometry write behavior using WGS84 coordinates
- Run writes transactionally per event or per safe batch
- Continue processing after isolated record-level failures

## FR-9 Debug artifacts and replay

The Go worker must retain the current ability to debug ingestion failures.

Required capabilities:

- Persist raw payload artifacts for failed or sampled records
- Persist metadata about source, stage, identifier, and error
- Replay a saved artifact through normalization/persistence
- List recent artifacts for inspection

This is mandatory. Migration is incomplete without equivalent replay/debug support.

## FR-10 Scheduling

Go must own the ingestion schedule now handled by APScheduler.

Minimum jobs:

- OpenAgenda daily
- data.gouv daily
- vacuum/cleanup job if still required

Scheduling requirements:

- Cron expressions configurable by env/config
- Timezone configurable
- Jobs must not overlap for the same source unless explicitly allowed
- Last run result must be logged in a structured way

## FR-11 Triggering model

The system must support both:

- Scheduled execution
- Manual execution

Manual execution can be exposed through:

- CLI flags on `back/cmd/ingest`
- Private/admin HTTP endpoints

Admin endpoints must not be exposed publicly without authentication/authorization.

## FR-12 Observability

The Go worker must provide:

- Structured logs with source/job identifiers
- Job-level summary metrics
- Record-level error counts
- Clear distinction between upstream fetch failure, normalization failure, and DB persistence failure

Preferred additions:

- Prometheus-compatible metrics if observability stack exists later

## FR-13 Idempotency

Ingestion jobs must be safe to re-run.

Requirements:

- Re-running the same source batch must not create duplicate rows
- Partial job failure must not require manual cleanup before retry
- Batch hashes or equivalent source-state markers should be used where useful

## FR-14 Configuration

The Go worker must support configuration for:

- Source credentials
- Source base URLs
- Timeouts
- Retry counts
- Scheduler cron expressions
- Geocoder settings
- Debug artifact storage path
- Source enable/disable flags

## Non-Functional Requirements

## NFR-1 Reliability

- A single bad record must not fail the full source job
- A single source failing must not prevent other sources from running
- Upstream 5xx handling should distinguish retryable and non-retryable cases

## NFR-2 Performance

- Network fetches should use bounded concurrency
- Parsing and persistence should avoid unbounded memory growth
- Large source datasets should be processable in chunks

## NFR-3 Maintainability

- Each source must be independently testable
- Parsing/normalization logic should be separated from transport code
- Store logic should be reusable across sources

## NFR-4 Compatibility

- The migration should not require the frontend to change
- The Python `geo` service should continue reading the same event data

## NFR-5 Security

- Secrets must come from environment/config, not source code
- Admin/trigger endpoints must be protected
- Crawlers must respect configured rate limits and domain-specific politeness constraints

## Data Ownership and DB Contract

The DB remains the shared contract between Go ETL and Python `geo`.

### Go ETL writes

- `events`
- `source_hashes`
- source-specific metadata tables already introduced for ingestion, if retained

### Python `geo` reads

- `events`
- any existing derived/source tables needed for search or geo processing

Important rule:

- The migration should favor preserving the current schema first
- Any schema changes required by the Go ETL migration must be backward-compatible with the Python `geo` read path

## Interfaces

## CLI interface

The worker should support commands such as:

```bash
go run ./cmd/ingest --source=openagenda --once
go run ./cmd/ingest --source=datagouv --once
go run ./cmd/ingest --all --once
go run ./cmd/ingest --replay-artifact=<artifact-id>
go run ./cmd/ingest --scheduler
```

Exact flags may vary, but these capabilities are required.

## Optional admin HTTP interface

Suggested private routes:

- `POST /internal/ingestion/openagenda`
- `POST /internal/ingestion/datagouv`
- `POST /internal/ingestion/replay`
- `GET /internal/ingestion/artifacts`
- `GET /internal/ingestion/health`

These should be internal-only or admin-protected.

## Migration Plan

## Phase 1 Foundation

- Create `back/cmd/ingest`
- Add shared ingestion config
- Add package scaffolding under `back/internal/ingestion`
- Add structured result models
- Add artifact persistence/replay primitives

## Phase 2 OpenAgenda migration

- Port agenda catalog fetch
- Port agenda event fetch
- Port batch hashing and unchanged detection
- Port normalization and persistence
- Validate parity against Python outputs

## Phase 3 data.gouv migration

- Port dataset discovery
- Port resource parsers
- Port per-record normalization
- Validate parity against Python outputs

## Phase 4 Scheduler and triggers

- Move cron execution to Go worker
- Add manual CLI and optional admin endpoints
- Disable Python scheduler for migrated sources

## Phase 5 Python ETL decommission

- Remove ETL routes from `geo` for migrated sources
- Remove scheduler ownership from `geo`
- Keep `geo` search APIs intact

## Acceptance Criteria

- A developer can run the Go ingestion worker locally and trigger each supported source manually
- Scheduled ingestion for migrated sources runs in Go, not Python
- The Python `geo` service remains functional for search endpoints
- Existing event search results remain backed by the same DB tables
- Re-running a migrated source does not create duplicate events
- Failed records produce inspectable debug artifacts
- Replay of a saved artifact is possible from Go
- Job summaries provide fetched/inserted/updated/duplicate/failed counts
- Source-specific migration from Python can be completed independently without blocking unrelated sources

## Risks

- Dedup behavior may drift if similarity logic is ported loosely
- Crawlers can become the hardest part if browser automation is required
- Replay/debug support is easy to underbuild and should not be deferred
- If ingestion is embedded only in the main Go server, operational coupling will remain too high

## Open Decisions

- Whether `back/cmd/ingest` should expose an HTTP admin surface or remain CLI-only
- Whether some crawler sources should temporarily remain Python-backed behind an explicit bridge
- Whether agenda/source metadata tables remain as-is or are simplified later
- Whether BAN geocoding for ingestion stays in Go or is routed through the Python service

## Recommendation

Implement the migration as a dedicated Go ingestion worker with source-by-source parity, while keeping the Python `geo` service focused on search and geospatial processing.

That gives a clean service boundary:

- Go owns data ingestion and write-path orchestration
- Python owns search and geo-heavy read-path behavior
