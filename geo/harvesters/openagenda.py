from __future__ import annotations

import hashlib
import json
import logging
from datetime import UTC, datetime, timedelta
from typing import Any

import httpx

from config import get_settings
from harvesters.common import build_address, first_string, parse_datetime
from models import Event, OpenAgendaAgenda
from services.ban import resolve_address
from services.debug import persist_debug_artifact

logger = logging.getLogger(__name__)


def _openagenda_headers() -> dict[str, str]:
    return {"key": get_settings().openagenda_api_key}


def payload_hash(payload: Any) -> str:
    encoded = json.dumps(
        payload,
        ensure_ascii=True,
        sort_keys=True,
        separators=(",", ":"),
        default=str,
    ).encode("utf-8")
    return hashlib.sha256(encoded).hexdigest()


def _extract_coordinates(payload: dict[str, Any]) -> tuple[float, float] | None:
    geom = payload.get("geom")
    if isinstance(geom, dict):
        coordinates = geom.get("coordinates")
        if isinstance(coordinates, list) and len(coordinates) >= 2:
            return float(coordinates[0]), float(coordinates[1])

    location = payload.get("location")
    if isinstance(location, dict):
        coordinates = location.get("coordinates")
        if isinstance(coordinates, list) and len(coordinates) >= 2:
            return float(coordinates[0]), float(coordinates[1])

    lon = payload.get("longitude") or payload.get("lon")
    lat = payload.get("latitude") or payload.get("lat")
    if lon is not None and lat is not None:
        return float(lon), float(lat)

    return None


def _extract_timing(payload: dict[str, Any]) -> tuple[object, object]:
    for key in ("timings", "timing", "dates"):
        raw_value = payload.get(key)
        if isinstance(raw_value, list) and raw_value:
            first_item = raw_value[0]
            if isinstance(first_item, dict):
                start = (
                    first_item.get("begin")
                    or first_item.get("start")
                    or first_item.get("startTime")
                )
                end = (
                    first_item.get("end")
                    or first_item.get("finish")
                    or first_item.get("endTime")
                )
                return start, end

    return payload.get("firstDate"), payload.get("lastDate")


async def fetch_openagenda_agendas_raw(
    *,
    updated_since: datetime | None = None,
    official_only: bool | None = None,
    max_pages: int | None = None,
) -> list[dict[str, Any]]:
    settings = get_settings()
    if not settings.openagenda_api_key:
        logger.warning("OPENAGENDA_API_KEY is not configured; skipping agenda discovery")
        return []

    if official_only is None:
        official_only = settings.openagenda_official_only
    if updated_since is None and settings.openagenda_agenda_updated_within_days > 0:
        updated_since = datetime.now(UTC) - timedelta(
            days=settings.openagenda_agenda_updated_within_days
        )

    agendas: list[dict[str, Any]] = []
    params: list[tuple[str, str]] = [
        ("size", "100"),
        ("sort", "recentlyAddedEvents.desc"),
        ("includeFields[]", "description"),
        ("includeFields[]", "slug"),
        ("includeFields[]", "summary"),
        ("includeFields[]", "title"),
        ("includeFields[]", "uid"),
        ("includeFields[]", "updatedAt"),
        ("includeFields[]", "official"),
    ]
    if official_only:
        params.append(("official", "1"))
    if updated_since is not None:
        params.append(("updatedAt.gte", updated_since.astimezone(UTC).isoformat()))

    async with httpx.AsyncClient(
        base_url=settings.openagenda_api_url,
        headers=_openagenda_headers(),
        timeout=settings.http_timeout_seconds,
    ) as client:
        after: list[Any] | None = None
        page = 0
        while True:
            page += 1
            request_params = list(params)
            if after:
                request_params.extend([("after[]", str(item)) for item in after])

            try:
                response = await client.get("/agendas", params=request_params)
                response.raise_for_status()
            except httpx.HTTPStatusError as exc:
                if exc.response.status_code >= 500 and agendas:
                    logger.warning(
                        "OpenAgenda agenda pagination stopped after page %s due to upstream %s; "
                        "returning %s agendas collected so far",
                        page,
                        exc.response.status_code,
                        len(agendas),
                    )
                    break
                raise

            payload = response.json()
            items = payload.get("agendas") or []
            agendas.extend(item for item in items if isinstance(item, dict))

            after = payload.get("after")
            if not after or (max_pages is not None and page >= max_pages):
                break

    return agendas


def normalize_openagenda_agenda(raw_agenda: dict[str, Any]) -> OpenAgendaAgenda | None:
    uid = raw_agenda.get("uid")
    if uid is None:
        return None

    summary = raw_agenda.get("summary") or {}
    published_events = summary.get("publishedEvents") or {}
    recently_added = summary.get("recentlyAddedEvents") or {}

    return OpenAgendaAgenda(
        uid=int(uid),
        slug=first_string(raw_agenda.get("slug")),
        title=first_string(raw_agenda.get("title")) or str(uid),
        description=first_string(raw_agenda.get("description")),
        official=bool(raw_agenda.get("official")),
        updated_at=parse_datetime(raw_agenda.get("updatedAt")),
        upcoming_events=int(published_events.get("current") or 0)
        + int(published_events.get("upcoming") or 0),
        recently_added_events=int(sum(recently_added.values()))
        if isinstance(recently_added, dict)
        else 0,
        source_url=f"https://openagenda.com/agendas/{uid}",
        last_payload_hash=payload_hash(raw_agenda),
        last_seen_at=datetime.now(UTC),
    )


async def fetch_openagenda_events_raw(
    agenda_uid: int,
    *,
    include_current: bool = True,
    include_upcoming: bool = True,
    max_pages: int | None = None,
) -> list[dict[str, Any]]:
    settings = get_settings()
    if not settings.openagenda_api_key:
        logger.warning("OPENAGENDA_API_KEY is not configured; skipping event fetch")
        return []

    relative_values: list[str] = []
    if include_current:
        relative_values.append("current")
    if include_upcoming:
        relative_values.append("upcoming")

    events: list[dict[str, Any]] = []

    async with httpx.AsyncClient(
        base_url=settings.openagenda_api_url,
        headers=_openagenda_headers(),
        timeout=settings.http_timeout_seconds,
    ) as client:
        after: list[Any] | None = None
        page = 0
        while True:
            page += 1
            params: list[tuple[str, str]] = [
                ("size", "300"),
                ("monolingual", "fr"),
                ("detailed", "1"),
            ]
            params.extend([("relative[]", value) for value in relative_values])
            if after:
                params.extend([("after[]", str(item)) for item in after])

            response = await client.get(f"/agendas/{agenda_uid}/events", params=params)
            response.raise_for_status()
            payload = response.json()
            items = payload.get("events") or []
            events.extend(item for item in items if isinstance(item, dict))

            after = payload.get("after")
            if not after or (max_pages is not None and page >= max_pages):
                break

    return events


async def normalize_openagenda_event(
    raw_event: dict[str, Any],
    *,
    agenda_uid: int | None = None,
) -> Event | None:
    title = first_string(raw_event.get("title"))
    start_raw, end_raw = _extract_timing(raw_event)
    start_dt = parse_datetime(start_raw)

    location = (
        raw_event.get("location") if isinstance(raw_event.get("location"), dict) else {}
    )
    address = build_address(
        [
            first_string(raw_event.get("address")),
            first_string(location.get("address")),
            first_string(location.get("postalCode")),
            first_string(location.get("city")),
        ]
    )
    if not title or start_dt is None or not address:
        return None

    coordinates = _extract_coordinates(raw_event)
    if coordinates is None:
        coordinates = await resolve_address(address)

    raw_uid = str(
        raw_event.get("uid") or raw_event.get("id") or raw_event.get("slug") or title
    )
    source_uid = f"{agenda_uid}:{raw_uid}" if agenda_uid is not None else raw_uid
    source_url = (
        first_string(raw_event.get("canonicalUrl"))
        or first_string(raw_event.get("url"))
        or first_string(raw_event.get("html"))
        or (
            f"https://openagenda.com/agendas/{agenda_uid}/events/{raw_uid}"
            if agenda_uid is not None
            else f"https://openagenda.com/{raw_uid}"
        )
    )

    return Event(
        source_uid=source_uid,
        title=title,
        description=first_string(raw_event.get("longDescription"))
        or first_string(raw_event.get("description")),
        start_dt=start_dt,
        end_dt=parse_datetime(end_raw),
        location_name=first_string(location.get("name"))
        or first_string(raw_event.get("locationName")),
        address=address,
        longitude=coordinates[0],
        latitude=coordinates[1],
        source_tag="openagenda",
        source_url=source_url,
    )


async def normalize_openagenda_events(
    raw_events: list[dict[str, Any]],
    *,
    agenda_uid: int,
) -> list[Event]:
    events: list[Event] = []
    for raw_event in raw_events:
        try:
            event = await normalize_openagenda_event(raw_event, agenda_uid=agenda_uid)
        except Exception as exc:
            artifact_id = persist_debug_artifact(
                source_tag="openagenda",
                artifact_type="openagenda_raw",
                stage="normalize",
                identifier=f"{agenda_uid}:{raw_event.get('uid') or raw_event.get('id') or 'unknown'}",
                payload=raw_event,
                metadata={"error": str(exc), "agenda_uid": agenda_uid},
            )
            logger.warning(
                "Skipping OpenAgenda event %s from agenda %s: %s [artifact_id=%s]",
                raw_event.get("uid") or raw_event.get("id"),
                agenda_uid,
                exc,
                artifact_id,
            )
            continue
        if event is not None:
            events.append(event)
    return events


def batch_hash_for_agendas(raw_agendas: list[dict[str, Any]]) -> str:
    return payload_hash(raw_agendas)


def batch_hash_for_events(raw_events: list[dict[str, Any]]) -> str:
    return payload_hash(raw_events)


async def fetch_openagenda_events() -> list[Event]:
    raw_agendas = await fetch_openagenda_agendas_raw()
    normalized_agendas = [
        agenda
        for agenda in (
            normalize_openagenda_agenda(raw_agenda) for raw_agenda in raw_agendas
        )
        if agenda is not None and agenda.upcoming_events > 0
    ]

    events: list[Event] = []
    for agenda in normalized_agendas:
        raw_events = await fetch_openagenda_events_raw(agenda.uid)
        events.extend(await normalize_openagenda_events(raw_events, agenda_uid=agenda.uid))

    return events


async def replay_openagenda_payload(payload: dict[str, Any]) -> Event | None:
    if not isinstance(payload, dict):
        raise ValueError("OpenAgenda replay payload must be an object")
    return await normalize_openagenda_event(payload)
