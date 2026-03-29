from __future__ import annotations

from typing import Annotated
from uuid import UUID

from asyncpg import Pool
from fastapi import APIRouter, Depends, HTTPException, Query, status

from database import get_pool
from models import (
    CitySuggestion,
    EventDetail,
    EventSearchParams,
    EventSearchResponse,
    EventSummary,
)
from services.ban import resolve_city, search_cities


router = APIRouter(tags=["events"])


async def get_event_search_params(
    city: Annotated[str | None, Query()] = None,
    lat: Annotated[float | None, Query()] = None,
    lon: Annotated[float | None, Query()] = None,
    radius_km: Annotated[int, Query(ge=1, le=50)] = 5,
    days: Annotated[int, Query(ge=1, le=28)] = 28,
    page: Annotated[int, Query(ge=1)] = 1,
    limit: Annotated[int, Query(ge=1, le=200)] = 50,
) -> EventSearchParams:
    return EventSearchParams(
        city=city,
        lat=lat,
        lon=lon,
        radius_km=radius_km,
        days=days,
        page=page,
        limit=limit,
    )


@router.get("/events", response_model=EventSearchResponse)
async def list_events(
    params: EventSearchParams = Depends(get_event_search_params),
    pool: Pool = Depends(get_pool),
) -> EventSearchResponse:
    if params.city is not None:
        lon, lat = await resolve_city(params.city)
    else:
        lat = float(params.lat)
        lon = float(params.lon)

    offset = (params.page - 1) * params.limit

    async with pool.acquire() as connection:
        total = await connection.fetchval(
            """
            SELECT COUNT(*)
            FROM events
            WHERE ST_DWithin(
                geom,
                ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
                $3
            )
            AND start_dt BETWEEN NOW() AND NOW() + ($4 * INTERVAL '1 day')
            """,
            lon,
            lat,
            params.radius_km * 1000,
            params.days,
        )
        rows = await connection.fetch(
            """
            SELECT
                event_id,
                title,
                start_dt,
                address,
                ST_Y(geom::geometry) AS lat,
                ST_X(geom::geometry) AS lon,
                source_tag,
                source_url
            FROM events
            WHERE ST_DWithin(
                geom,
                ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
                $3
            )
            AND start_dt BETWEEN NOW() AND NOW() + ($4 * INTERVAL '1 day')
            ORDER BY start_dt ASC
            LIMIT $5
            OFFSET $6
            """,
            lon,
            lat,
            params.radius_km * 1000,
            params.days,
            params.limit,
            offset,
        )

    return EventSearchResponse(
        total=int(total or 0),
        page=params.page,
        results=[EventSummary.model_validate(dict(row)) for row in rows],
    )


@router.get("/events/{event_id}", response_model=EventDetail)
async def get_event(event_id: UUID, pool: Pool = Depends(get_pool)) -> EventDetail:
    async with pool.acquire() as connection:
        row = await connection.fetchrow(
            """
            SELECT
                event_id,
                title,
                description,
                start_dt,
                end_dt,
                location_name,
                address,
                ST_Y(geom::geometry) AS lat,
                ST_X(geom::geometry) AS lon,
                source_tag,
                source_url
            FROM events
            WHERE event_id = $1
            """,
            event_id,
        )

    if row is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Event not found",
        )

    return EventDetail.model_validate(dict(row))


@router.get("/cities/search", response_model=list[CitySuggestion])
async def list_cities(
    q: Annotated[str, Query(min_length=1)],
) -> list[CitySuggestion]:
    return await search_cities(q)
