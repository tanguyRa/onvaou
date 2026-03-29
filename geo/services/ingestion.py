from __future__ import annotations

import logging

from models import IngestionResult
from services.store import (
    has_openagenda_event_batch_changed,
    list_relevant_openagenda_agenda_ids,
    list_changed_openagenda_agendas,
    mark_openagenda_agenda_sync_result,
    mark_openagenda_agenda_synced,
    store_openagenda_agenda_fetch,
    store_openagenda_agenda_payloads,
    store_openagenda_event_fetch,
    upsert_event_batch,
    upsert_openagenda_agendas,
)
from harvesters.datagouv import fetch_datagouv_events
from harvesters.openagenda import (
    batch_hash_for_agendas,
    batch_hash_for_events,
    fetch_openagenda_agendas_raw,
    fetch_openagenda_events_raw,
    normalize_openagenda_agenda,
    normalize_openagenda_events,
)

logger = logging.getLogger(__name__)


async def sync_openagenda_agenda_catalog() -> dict[str, int]:
    raw_agendas = await fetch_openagenda_agendas_raw()
    raw_batch_hash = batch_hash_for_agendas(raw_agendas)
    is_new_fetch = await store_openagenda_agenda_fetch(
        batch_hash=raw_batch_hash,
        raw_agendas=raw_agendas,
    )
    agendas = [
        agenda
        for agenda in (normalize_openagenda_agenda(item) for item in raw_agendas)
        if agenda is not None
    ]
    changed_agendas = await list_changed_openagenda_agendas(agendas)
    stored_payloads = await store_openagenda_agenda_payloads(changed_agendas)
    agenda_store_result = await upsert_openagenda_agendas(changed_agendas)
    return {
        "fetched": len(raw_agendas),
        "batch_new": int(is_new_fetch),
        "changed": len(changed_agendas),
        "payloads_stored": stored_payloads,
        "inserted": agenda_store_result["inserted"],
        "updated": agenda_store_result["updated"],
    }


async def sync_openagenda_agenda(agenda_id: int) -> IngestionResult:
    try:
        raw_events = await fetch_openagenda_events_raw(agenda_id)
        event_batch_hash = batch_hash_for_events(raw_events)
        await store_openagenda_event_fetch(
            agenda_uid=agenda_id,
            batch_hash=event_batch_hash,
            raw_events=raw_events,
        )

        if not await has_openagenda_event_batch_changed(
            agenda_uid=agenda_id,
            batch_hash=event_batch_hash,
        ):
            result = IngestionResult(
                source_tag="openagenda",
                fetched=len(raw_events),
                details=[f"agenda {agenda_id}: unchanged batch"],
            )
            await mark_openagenda_agenda_sync_result(
                agenda_uid=agenda_id,
                batch_hash=event_batch_hash,
                error="",
            )
            await mark_openagenda_agenda_synced(agenda_id)
            return result

        events = await normalize_openagenda_events(raw_events, agenda_uid=agenda_id)
        result = await upsert_event_batch("openagenda", events)
        await mark_openagenda_agenda_sync_result(
            agenda_uid=agenda_id,
            batch_hash=event_batch_hash,
            error="",
        )
        await mark_openagenda_agenda_synced(agenda_id)
        result.details.append(f"agenda {agenda_id}: processed")
        return result
    except Exception as exc:
        await mark_openagenda_agenda_sync_result(
            agenda_uid=agenda_id,
            error=str(exc),
        )
        raise


async def ingest_openagenda() -> IngestionResult:
    catalog_result = await sync_openagenda_agenda_catalog()
    agenda_ids = await list_relevant_openagenda_agenda_ids()

    result = IngestionResult(source_tag="openagenda")
    result.details.append(
        "agenda catalog: "
        f"fetched={catalog_result['fetched']} "
        f"changed={catalog_result['changed']} "
        f"inserted={catalog_result['inserted']} "
        f"updated={catalog_result['updated']}"
    )

    for agenda_id in agenda_ids:
        try:
            agenda_result = await sync_openagenda_agenda(agenda_id)
        except Exception as exc:
            result.failed += 1
            result.details.append(f"agenda {agenda_id}: {exc}")
            continue

        result.fetched += agenda_result.fetched
        result.inserted += agenda_result.inserted
        result.updated += agenda_result.updated
        result.duplicates += agenda_result.duplicates
        result.failed += agenda_result.failed
        result.details.extend(agenda_result.details)

    logger.info("OpenAgenda ingestion completed: %s", result.model_dump())
    return result


async def ingest_datagouv() -> IngestionResult:
    events = await fetch_datagouv_events()
    result = await upsert_event_batch("datagouv", events)
    logger.info("data.gouv ingestion completed: %s", result.model_dump())
    return result
