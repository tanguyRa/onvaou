from __future__ import annotations

from typing import Any

from models import Event, IngestionResult, ReplayResponse
from services.debug import artifact_summary, load_debug_artifact
from services.store import upsert_event_batch
from harvesters.datagouv import replay_datagouv_payload
from harvesters.openagenda import replay_openagenda_payload


async def replay_artifact(artifact_id: str) -> ReplayResponse:
    artifact = load_debug_artifact(artifact_id)
    summary = artifact_summary(artifact_id)
    payload = artifact.get("payload")
    metadata = artifact.get("metadata", {})

    event: Event | None = None
    if artifact["artifact_type"] == "openagenda_raw":
        event = await replay_openagenda_payload(payload)
    elif artifact["artifact_type"] == "datagouv_raw":
        event = await replay_datagouv_payload(
            payload,
            source_uid=str(metadata.get("source_uid") or artifact["identifier"]),
            source_url=str(metadata.get("source_url") or ""),
        )
    elif artifact["artifact_type"] == "normalized_event":
        event = Event.model_validate(payload)
    else:
        return ReplayResponse(
            artifact=summary,
            status="unsupported",
            message=f"Unsupported artifact type: {artifact['artifact_type']}",
        )

    if event is None:
        return ReplayResponse(
            artifact=summary,
            status="skipped",
            message="Replay payload still does not normalize into an event",
        )

    result = await upsert_event_batch(event.source_tag, [event])
    return ReplayResponse(
        artifact=summary,
        status="replayed",
        message="Replay completed",
        result=result,
    )
