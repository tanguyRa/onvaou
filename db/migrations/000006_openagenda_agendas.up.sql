BEGIN;

CREATE TABLE "openagenda_agenda_fetches" (
    fetch_id BIGSERIAL PRIMARY KEY,
    batch_hash TEXT NOT NULL UNIQUE,
    agenda_count INTEGER NOT NULL,
    raw_payload JSONB NOT NULL,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "openagenda_agendas" (
    uid BIGINT PRIMARY KEY,
    slug TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    official BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ,
    upcoming_events INTEGER NOT NULL DEFAULT 0,
    recently_added_events INTEGER NOT NULL DEFAULT 0,
    source_url TEXT NOT NULL DEFAULT '',
    last_payload_hash TEXT NOT NULL DEFAULT '',
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_fetch_attempt_at TIMESTAMPTZ,
    last_fetch_success_at TIMESTAMPTZ,
    last_event_sync_at TIMESTAMPTZ,
    last_event_batch_hash TEXT NOT NULL DEFAULT '',
    last_fetch_error TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_openagenda_agendas_relevant
ON "openagenda_agendas" (official, upcoming_events, updated_at DESC);

CREATE TABLE "openagenda_agenda_payloads" (
    uid BIGINT NOT NULL REFERENCES "openagenda_agendas" (uid) ON DELETE CASCADE,
    payload_hash TEXT NOT NULL,
    raw_payload JSONB NOT NULL,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (uid, payload_hash)
);

CREATE TABLE "openagenda_event_fetches" (
    agenda_uid BIGINT NOT NULL REFERENCES "openagenda_agendas" (uid) ON DELETE CASCADE,
    batch_hash TEXT NOT NULL,
    event_count INTEGER NOT NULL,
    raw_payload JSONB NOT NULL,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (agenda_uid, batch_hash)
);

COMMIT;
