BEGIN;

CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE "events" (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    start_dt TIMESTAMPTZ NOT NULL,
    end_dt TIMESTAMPTZ,
    location_name TEXT NOT NULL,
    address TEXT NOT NULL,
    geom GEOGRAPHY(Point, 4326) NOT NULL,
    source_tag TEXT NOT NULL,
    source_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_events_geom ON "events" USING GIST (geom);

CREATE INDEX idx_events_start_dt ON "events" (start_dt);

CREATE TABLE "source_hashes" (
    source_tag TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    event_id UUID NOT NULL REFERENCES "events" (event_id) ON DELETE CASCADE,
    PRIMARY KEY (source_tag, content_hash)
);

COMMIT;
