from __future__ import annotations

import csv
import io
import json
import logging
from collections.abc import Iterable
from typing import Any

import httpx

from config import get_settings
from harvesters.common import build_address, first_string, parse_datetime
from models import Event
from services.debug import persist_debug_artifact
from services.ban import resolve_address

logger = logging.getLogger(__name__)

DATASET_KEYWORDS = (
    "agenda",
    "evenement",
    "événement",
    "manifestation",
    "culture",
    "festival",
    "mairie",
)

SEARCH_QUERIES = (
    "agenda",
    "evenement",
    "culture",
    "festival",
    "mairie",
)


def _matches_keywords(*values: str) -> bool:
    haystack = " ".join(values).lower()
    return any(keyword in haystack for keyword in DATASET_KEYWORDS)


def _pick_field(row: dict[str, object], *names: str) -> str:
    lowered = {key.lower(): key for key in row}
    for candidate in names:
        if candidate in lowered:
            return first_string(row[lowered[candidate]])
    return ""


def _iter_rows_from_json(payload: object) -> Iterable[dict[str, object]]:
    if isinstance(payload, list):
        for item in payload:
            if isinstance(item, dict):
                yield item
        return
    if isinstance(payload, dict):
        for key in ("data", "results", "records", "items"):
            value = payload.get(key)
            if isinstance(value, list):
                for item in value:
                    if isinstance(item, dict):
                        yield item


def _record_row_failure(
    *,
    row: dict[str, object],
    source_uid: str,
    source_url: str,
    exc: Exception,
) -> str:
    return persist_debug_artifact(
        source_tag="datagouv",
        artifact_type="datagouv_raw",
        stage="normalize",
        identifier=source_uid,
        payload=row,
        metadata={
            "error": str(exc),
            "source_uid": source_uid,
            "source_url": source_url,
        },
    )


async def normalize_datagouv_row(
    row: dict[str, object],
    source_uid: str,
    source_url: str,
) -> Event | None:
    title = _pick_field(row, "title", "titre", "name", "nom", "event", "summary")
    start_dt = parse_datetime(
        _pick_field(
            row,
            "start_dt",
            "start_date",
            "date_start",
            "date_debut",
            "date",
            "datetime",
        )
    )
    address = build_address(
        [
            _pick_field(row, "address", "adresse", "lieu", "location"),
            _pick_field(row, "postal_code", "code_postal", "postcode"),
            _pick_field(row, "city", "ville", "commune"),
        ]
    )

    if not title or start_dt is None or not address:
        return None

    lon_text = _pick_field(row, "longitude", "lon", "lng", "x")
    lat_text = _pick_field(row, "latitude", "lat", "y")
    if lon_text and lat_text:
        coordinates = (float(lon_text.replace(",", ".")), float(lat_text.replace(",", ".")))
    else:
        coordinates = await resolve_address(address)

    return Event(
        source_uid=source_uid,
        title=title,
        description=_pick_field(row, "description", "details", "resume", "summary"),
        start_dt=start_dt,
        end_dt=parse_datetime(
            _pick_field(row, "end_dt", "end_date", "date_fin", "end")
        ),
        location_name=_pick_field(row, "location_name", "venue", "place", "lieu"),
        address=address,
        longitude=coordinates[0],
        latitude=coordinates[1],
        source_tag="datagouv",
        source_url=source_url,
    )


async def _parse_resource(client: httpx.AsyncClient, resource: dict) -> list[Event]:
    url = resource.get("url")
    if not isinstance(url, str) or not url:
        return []

    response = await client.get(url)
    response.raise_for_status()

    format_name = first_string(resource.get("format")).lower()
    source_uid_prefix = str(resource.get("id") or url)
    events: list[Event] = []

    if format_name == "csv":
        content = response.text
        reader = csv.DictReader(io.StringIO(content))
        for index, row in enumerate(reader):
            source_uid = f"{source_uid_prefix}:{index}"
            try:
                event = await normalize_datagouv_row(row, source_uid, url)
            except Exception as exc:
                artifact_id = _record_row_failure(
                    row=row,
                    source_uid=source_uid,
                    source_url=url,
                    exc=exc,
                )
                logger.warning(
                    "Skipping data.gouv record %s: %s [artifact_id=%s]",
                    source_uid,
                    exc,
                    artifact_id,
                )
                continue
            if event is not None:
                events.append(event)
        return events

    if format_name in {"json", "geojson"}:
        payload = response.json()
        for index, row in enumerate(_iter_rows_from_json(payload)):
            source_uid = f"{source_uid_prefix}:{index}"
            try:
                event = await normalize_datagouv_row(row, source_uid, url)
            except Exception as exc:
                artifact_id = _record_row_failure(
                    row=row,
                    source_uid=source_uid,
                    source_url=url,
                    exc=exc,
                )
                logger.warning(
                    "Skipping data.gouv record %s: %s [artifact_id=%s]",
                    source_uid,
                    exc,
                    artifact_id,
                )
                continue
            if event is not None:
                events.append(event)
        return events

    content_type = response.headers.get("content-type", "")
    if "json" in content_type:
        payload = json.loads(response.text)
        for index, row in enumerate(_iter_rows_from_json(payload)):
            source_uid = f"{source_uid_prefix}:{index}"
            try:
                event = await normalize_datagouv_row(row, source_uid, url)
            except Exception as exc:
                artifact_id = _record_row_failure(
                    row=row,
                    source_uid=source_uid,
                    source_url=url,
                    exc=exc,
                )
                logger.warning(
                    "Skipping data.gouv record %s: %s [artifact_id=%s]",
                    source_uid,
                    exc,
                    artifact_id,
                )
                continue
            if event is not None:
                events.append(event)

    return events


async def replay_datagouv_payload(
    payload: dict[str, Any],
    *,
    source_uid: str,
    source_url: str,
) -> Event | None:
    if not isinstance(payload, dict):
        raise ValueError("data.gouv replay payload must be an object")
    return await normalize_datagouv_row(payload, source_uid, source_url)


async def fetch_datagouv_events() -> list[Event]:
    settings = get_settings()
    events: list[Event] = []
    seen_dataset_ids: set[str] = set()
    seen_resource_ids: set[str] = set()

    async with httpx.AsyncClient(
        base_url=settings.datagouv_api_url,
        timeout=settings.http_timeout_seconds,
        follow_redirects=True,
    ) as client:
        total_datasets = 0
        for query in SEARCH_QUERIES:
            response = await client.get("/datasets/", params={"page_size": 20, "q": query})
            response.raise_for_status()
            payload = response.json()
            datasets = payload.get("data") or payload.get("results") or []
            total_datasets += len(datasets)

            for dataset in datasets:
                if not isinstance(dataset, dict):
                    continue
                dataset_id = str(dataset.get("id") or dataset.get("slug") or "")
                if dataset_id and dataset_id in seen_dataset_ids:
                    continue
                if dataset_id:
                    seen_dataset_ids.add(dataset_id)

                title = first_string(dataset.get("title"))
                description = first_string(dataset.get("description"))
                if not _matches_keywords(title, description):
                    continue

                resources = dataset.get("resources") or []
                for resource in resources:
                    if not isinstance(resource, dict):
                        continue
                    resource_id = str(resource.get("id") or resource.get("url") or "")
                    if resource_id and resource_id in seen_resource_ids:
                        continue
                    if resource_id:
                        seen_resource_ids.add(resource_id)

                    format_name = first_string(resource.get("format")).lower()
                    if format_name not in {"csv", "json", "geojson"}:
                        continue
                    try:
                        events.extend(await _parse_resource(client, resource))
                    except Exception as exc:
                        logger.warning(
                            "Skipping data.gouv resource %s: %s",
                            resource.get("id") or resource.get("url"),
                            exc,
                        )

        if total_datasets == 0:
            logger.warning(
                "data.gouv search returned no datasets for queries=%s",
                ",".join(SEARCH_QUERIES),
            )

    return events
