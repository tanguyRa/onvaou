from __future__ import annotations

import asyncio

import httpx

from config import get_settings
from models import CitySuggestion


async def _search_ban(query: str, *, limit: int, ban_type: str | None = None) -> list[dict]:
    settings = get_settings()
    query = " ".join(query.split()).strip()
    if not query:
        raise ValueError("query is required")

    last_error: Exception | None = None

    async with httpx.AsyncClient(
        base_url=settings.ban_api_url,
        timeout=settings.http_timeout_seconds,
    ) as client:
        for attempt in range(1, settings.ban_retry_attempts + 1):
            try:
                response = await client.get(
                    "/search/",
                    params={
                        "q": query,
                        "limit": limit,
                        "autocomplete": 1,
                        **({"type": ban_type} if ban_type else {}),
                    },
                )
                response.raise_for_status()
                payload = response.json()
                features = payload.get("features") or []
                if not features:
                    raise LookupError(f"BAN returned no result for query: {query}")

                return features
            except (httpx.HTTPError, LookupError, ValueError) as exc:
                last_error = exc
                if attempt == settings.ban_retry_attempts:
                    break
                await asyncio.sleep(2 ** (attempt - 1))

    raise RuntimeError(f"failed to search BAN for query: {query}") from last_error


def _parse_coordinates(feature: dict, query: str) -> tuple[float, float]:
    coordinates = feature.get("geometry", {}).get("coordinates") or []
    if len(coordinates) < 2:
        raise LookupError(f"BAN returned invalid coordinates for: {query}")

    return float(coordinates[0]), float(coordinates[1])


async def resolve_address(address: str) -> tuple[float, float]:
    features = await _search_ban(address, limit=1)
    return _parse_coordinates(features[0], address)


async def resolve_city(city: str) -> tuple[float, float]:
    features = await _search_ban(city, limit=1, ban_type="municipality")
    return _parse_coordinates(features[0], city)


async def search_cities(query: str, *, limit: int = 5) -> list[CitySuggestion]:
    features = await _search_ban(query, limit=limit, ban_type="municipality")
    suggestions: list[CitySuggestion] = []

    for feature in features:
        lon, lat = _parse_coordinates(feature, query)
        properties = feature.get("properties", {})
        suggestions.append(
            CitySuggestion(
                name=properties.get("label", ""),
                city=properties.get("city", ""),
                postcode=properties.get("postcode", ""),
                lat=lat,
                lon=lon,
            )
        )

    return suggestions
