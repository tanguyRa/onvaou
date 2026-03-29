from __future__ import annotations

import json
import uuid
from datetime import UTC, datetime
from typing import Any

from asyncpg import Connection

from database import database
from models import Event, IngestionResult, OpenAgendaAgenda
from services.debug import persist_debug_artifact
from services.dedup import check_duplicate


def _json_payload(value: Any) -> str:
    return json.dumps(value, ensure_ascii=True, sort_keys=True, default=str)


def _merge_event(existing_id: str, event: Event) -> Event:
    return event.model_copy(update={"event_id": uuid.UUID(existing_id)})


async def _upsert_event(connection: Connection, event: Event) -> str:
    event = event.with_event_id()
    decision = await check_duplicate(connection, event)

    if decision.action == "exact":
        await connection.execute(
            """
            INSERT INTO source_hashes (source_tag, content_hash, event_id)
            VALUES ($1, $2, $3)
            ON CONFLICT (source_tag, content_hash) DO NOTHING
            """,
            event.source_tag,
            event.content_hash,
            uuid.UUID(decision.event_id),
        )
        return "duplicate"

    if decision.action == "near" and decision.event_id is not None:
        event = _merge_event(decision.event_id, event)

    await connection.execute(
        """
        INSERT INTO events (
            event_id,
            title,
            description,
            start_dt,
            end_dt,
            location_name,
            address,
            geom,
            source_tag,
            source_url
        )
        VALUES (
            $1,
            $2,
            $3,
            $4,
            $5,
            $6,
            $7,
            ST_SetSRID(ST_MakePoint($8, $9), 4326)::geography,
            $10,
            $11
        )
        ON CONFLICT (event_id) DO UPDATE SET
            title = EXCLUDED.title,
            description = CASE
                WHEN EXCLUDED.description = '' THEN events.description
                ELSE EXCLUDED.description
            END,
            start_dt = EXCLUDED.start_dt,
            end_dt = COALESCE(EXCLUDED.end_dt, events.end_dt),
            location_name = CASE
                WHEN EXCLUDED.location_name = '' THEN events.location_name
                ELSE EXCLUDED.location_name
            END,
            address = EXCLUDED.address,
            geom = EXCLUDED.geom,
            source_tag = events.source_tag,
            source_url = CASE
                WHEN EXCLUDED.source_url = '' THEN events.source_url
                ELSE EXCLUDED.source_url
            END
        """,
        event.event_id,
        event.title,
        event.description,
        event.start_dt,
        event.end_dt,
        event.location_name,
        event.address,
        event.longitude,
        event.latitude,
        event.source_tag,
        event.source_url,
    )

    await connection.execute(
        """
        INSERT INTO source_hashes (source_tag, content_hash, event_id)
        VALUES ($1, $2, $3)
        ON CONFLICT (source_tag, content_hash) DO UPDATE SET
            event_id = EXCLUDED.event_id
        """,
        event.source_tag,
        event.content_hash,
        event.event_id,
    )

    return "updated" if decision.action == "near" else "inserted"


async def upsert_event_batch(source_tag: str, events: list[Event]) -> IngestionResult:
    result = IngestionResult(source_tag=source_tag, fetched=len(events))
    normalized_events = [event.with_event_id() for event in events]

    async with database.pool.acquire() as connection:
        for event in normalized_events:
            try:
                async with connection.transaction():
                    status = await _upsert_event(connection, event)
            except Exception as exc:
                artifact_id = persist_debug_artifact(
                    source_tag=event.source_tag,
                    artifact_type="normalized_event",
                    stage="upsert",
                    identifier=event.source_uid,
                    payload=event,
                    metadata={"error": str(exc)},
                )
                result.failed += 1
                result.details.append(
                    f"{event.source_uid}: {exc} [artifact_id={artifact_id}]"
                )
                continue

            if status == "duplicate":
                result.duplicates += 1
            elif status == "updated":
                result.updated += 1
            else:
                result.inserted += 1

    return result


async def upsert_openagenda_agendas(agendas: list[OpenAgendaAgenda]) -> dict[str, int]:
    inserted = 0
    updated = 0

    async with database.pool.acquire() as connection:
        for agenda in agendas:
            status = await connection.fetchval(
                """
                INSERT INTO openagenda_agendas (
                    uid,
                    slug,
                    title,
                    description,
                    official,
                    updated_at,
                    upcoming_events,
                    recently_added_events,
                    source_url,
                    last_payload_hash,
                    last_seen_at,
                    last_fetch_success_at,
                    last_fetch_error
                )
                VALUES (
                    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
                    COALESCE($11, CURRENT_TIMESTAMP), CURRENT_TIMESTAMP, ''
                )
                ON CONFLICT (uid) DO UPDATE SET
                    slug = EXCLUDED.slug,
                    title = EXCLUDED.title,
                    description = EXCLUDED.description,
                    official = EXCLUDED.official,
                    updated_at = EXCLUDED.updated_at,
                    upcoming_events = EXCLUDED.upcoming_events,
                    recently_added_events = EXCLUDED.recently_added_events,
                    source_url = EXCLUDED.source_url,
                    last_payload_hash = EXCLUDED.last_payload_hash,
                    last_seen_at = EXCLUDED.last_seen_at,
                    last_fetch_success_at = CURRENT_TIMESTAMP,
                    last_fetch_error = ''
                RETURNING (xmax = 0) AS inserted
                """,
                agenda.uid,
                agenda.slug,
                agenda.title,
                agenda.description,
                agenda.official,
                agenda.updated_at,
                agenda.upcoming_events,
                agenda.recently_added_events,
                agenda.source_url,
                agenda.last_payload_hash,
                agenda.last_seen_at,
            )
            if status:
                inserted += 1
            else:
                updated += 1

    return {"inserted": inserted, "updated": updated}


async def store_openagenda_agenda_fetch(
    *,
    batch_hash: str,
    raw_agendas: list[dict[str, Any]],
) -> bool:
    async with database.pool.acquire() as connection:
        inserted = await connection.fetchval(
            """
            INSERT INTO openagenda_agenda_fetches (batch_hash, agenda_count, raw_payload)
            VALUES ($1, $2, $3::jsonb)
            ON CONFLICT (batch_hash) DO NOTHING
            RETURNING 1
            """,
            batch_hash,
            len(raw_agendas),
            _json_payload(raw_agendas),
        )
    return bool(inserted)


async def list_changed_openagenda_agendas(
    agendas: list[OpenAgendaAgenda],
) -> list[OpenAgendaAgenda]:
    if not agendas:
        return []

    async with database.pool.acquire() as connection:
        rows = await connection.fetch(
            """
            SELECT uid, last_payload_hash
            FROM openagenda_agendas
            WHERE uid = ANY($1::bigint[])
            """,
            [agenda.uid for agenda in agendas],
        )

    known_hashes = {int(row["uid"]): str(row["last_payload_hash"] or "") for row in rows}
    return [
        agenda
        for agenda in agendas
        if known_hashes.get(agenda.uid, "") != agenda.last_payload_hash
    ]


async def store_openagenda_agenda_payloads(agendas: list[OpenAgendaAgenda]) -> int:
    if not agendas:
        return 0

    stored = 0
    async with database.pool.acquire() as connection:
        for agenda in agendas:
            await connection.execute(
                """
                INSERT INTO openagenda_agendas (uid, title, last_fetch_attempt_at)
                VALUES ($1, $2, CURRENT_TIMESTAMP)
                ON CONFLICT (uid) DO UPDATE SET
                    last_fetch_attempt_at = CURRENT_TIMESTAMP
                """,
                agenda.uid,
                agenda.title,
            )

            inserted = await connection.fetchval(
                """
                INSERT INTO openagenda_agenda_payloads (uid, payload_hash, raw_payload)
                VALUES ($1, $2, $3::jsonb)
                ON CONFLICT (uid, payload_hash) DO NOTHING
                RETURNING 1
                """,
                agenda.uid,
                agenda.last_payload_hash,
                _json_payload(agenda.model_dump(mode="json")),
            )
            if inserted:
                stored += 1

    return stored


async def list_relevant_openagenda_agenda_ids(limit: int | None = None) -> list[int]:
    query = """
        SELECT uid
        FROM openagenda_agendas
        WHERE upcoming_events > 0
        ORDER BY official DESC, upcoming_events DESC, updated_at DESC NULLS LAST
    """
    args: list[object] = []
    if limit is not None and limit > 0:
        query += " LIMIT $1"
        args.append(limit)

    async with database.pool.acquire() as connection:
        rows = await connection.fetch(query, *args)
    return [int(row["uid"]) for row in rows]


async def has_openagenda_event_batch_changed(
    *,
    agenda_uid: int,
    batch_hash: str,
) -> bool:
    async with database.pool.acquire() as connection:
        stored_hash = await connection.fetchval(
            """
            SELECT last_event_batch_hash
            FROM openagenda_agendas
            WHERE uid = $1
            """,
            agenda_uid,
        )
    return str(stored_hash or "") != batch_hash


async def store_openagenda_event_fetch(
    *,
    agenda_uid: int,
    batch_hash: str,
    raw_events: list[dict[str, Any]],
) -> bool:
    async with database.pool.acquire() as connection:
        inserted = await connection.fetchval(
            """
            INSERT INTO openagenda_event_fetches (agenda_uid, batch_hash, event_count, raw_payload)
            VALUES ($1, $2, $3, $4::jsonb)
            ON CONFLICT (agenda_uid, batch_hash) DO NOTHING
            RETURNING 1
            """,
            agenda_uid,
            batch_hash,
            len(raw_events),
            _json_payload(raw_events),
        )
        await connection.execute(
            """
            UPDATE openagenda_agendas
            SET last_fetch_attempt_at = CURRENT_TIMESTAMP
            WHERE uid = $1
            """,
            agenda_uid,
        )
    return bool(inserted)


async def mark_openagenda_agenda_synced(agenda_uid: int) -> None:
    async with database.pool.acquire() as connection:
        await connection.execute(
            """
            UPDATE openagenda_agendas
            SET last_event_sync_at = $2
            WHERE uid = $1
            """,
            agenda_uid,
            datetime.now(UTC),
        )


async def mark_openagenda_agenda_sync_result(
    *,
    agenda_uid: int,
    batch_hash: str | None = None,
    error: str = "",
) -> None:
    async with database.pool.acquire() as connection:
        await connection.execute(
            """
            UPDATE openagenda_agendas
            SET last_fetch_success_at = CASE
                    WHEN $2 = '' THEN CURRENT_TIMESTAMP
                    ELSE last_fetch_success_at
                END,
                last_event_batch_hash = COALESCE($3, last_event_batch_hash),
                last_fetch_error = $2
            WHERE uid = $1
            """,
            agenda_uid,
            error,
            batch_hash,
        )


async def vacuum_events() -> None:
    async with database.pool.acquire() as connection:
        await connection.execute("VACUUM ANALYZE events")
        await connection.execute("VACUUM ANALYZE source_hashes")
