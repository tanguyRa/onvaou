from __future__ import annotations

from dataclasses import dataclass
from datetime import timedelta

from asyncpg import Connection
from fuzzywuzzy import fuzz

from models import Event


@dataclass(slots=True)
class DedupDecision:
    event_id: str | None
    action: str


async def check_duplicate(connection: Connection, event: Event) -> DedupDecision:
    existing_event_id = await connection.fetchval(
        """
        SELECT event_id
        FROM source_hashes
        WHERE source_tag = $1 AND content_hash = $2
        """,
        event.source_tag,
        event.content_hash,
    )
    if existing_event_id:
        return DedupDecision(event_id=str(existing_event_id), action="exact")

    candidate_rows = await connection.fetch(
        """
        SELECT event_id, title, address
        FROM events
        WHERE start_dt BETWEEN $1 AND $2
        LIMIT 50
        """,
        event.start_dt - timedelta(hours=12),
        event.start_dt + timedelta(hours=12),
    )

    best_match_id: str | None = None
    best_score = 0

    for candidate in candidate_rows:
        haystack = " ".join(
            [
                (candidate.get("title") or "").lower().strip(),
                (candidate.get("address") or "").lower().strip(),
            ]
        ).strip()
        if not haystack:
            continue

        score = fuzz.token_sort_ratio(event.dedup_text, haystack)
        if score >= 85 and score > best_score:
            best_score = score
            best_match_id = str(candidate["event_id"])

    if best_match_id:
        return DedupDecision(event_id=best_match_id, action="near")

    return DedupDecision(event_id=None, action="new")
