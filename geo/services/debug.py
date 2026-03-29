from __future__ import annotations

import json
import logging
from datetime import UTC, datetime
from pathlib import Path
from typing import Any
from uuid import uuid4

from config import get_settings
from models import ReplayArtifact

logger = logging.getLogger(__name__)


def _debug_root() -> Path:
    path = Path(get_settings().ingestion_debug_dir)
    path.mkdir(parents=True, exist_ok=True)
    return path


def _serialize(value: Any) -> Any:
    if isinstance(value, dict):
        return {str(key): _serialize(item) for key, item in value.items()}
    if isinstance(value, list):
        return [_serialize(item) for item in value]
    if isinstance(value, tuple):
        return [_serialize(item) for item in value]
    if hasattr(value, "model_dump"):
        return value.model_dump(mode="json")
    if hasattr(value, "isoformat"):
        return value.isoformat()
    return value


def persist_debug_artifact(
    *,
    source_tag: str,
    artifact_type: str,
    stage: str,
    identifier: str,
    payload: Any,
    metadata: dict[str, Any] | None = None,
) -> str:
    timestamp = datetime.now(UTC)
    artifact_id = f"{timestamp.strftime('%Y%m%dT%H%M%S')}-{uuid4().hex}"
    artifact = {
        "artifact_id": artifact_id,
        "source_tag": source_tag,
        "artifact_type": artifact_type,
        "stage": stage,
        "identifier": identifier,
        "saved_at": timestamp.isoformat(),
        "metadata": _serialize(metadata or {}),
        "payload": _serialize(payload),
    }
    path = _debug_root() / f"{artifact_id}.json"
    path.write_text(json.dumps(artifact, ensure_ascii=True, indent=2), encoding="utf-8")
    logger.debug("Saved ingestion debug artifact", extra={"artifact_id": artifact_id})
    return artifact_id


def load_debug_artifact(artifact_id: str) -> dict[str, Any]:
    path = _debug_root() / f"{artifact_id}.json"
    if not path.exists():
        raise FileNotFoundError(f"unknown ingestion artifact: {artifact_id}")
    return json.loads(path.read_text(encoding="utf-8"))


def artifact_summary(artifact_id: str) -> ReplayArtifact:
    artifact = load_debug_artifact(artifact_id)
    return ReplayArtifact.model_validate(
        {
            "artifact_id": artifact["artifact_id"],
            "source_tag": artifact["source_tag"],
            "artifact_type": artifact["artifact_type"],
            "stage": artifact["stage"],
            "identifier": artifact["identifier"],
            "saved_at": artifact["saved_at"],
            "metadata": artifact.get("metadata", {}),
        }
    )


def list_debug_artifacts(limit: int = 50) -> list[ReplayArtifact]:
    artifacts: list[ReplayArtifact] = []
    files = sorted(_debug_root().glob("*.json"), reverse=True)
    for path in files[:limit]:
        payload = json.loads(path.read_text(encoding="utf-8"))
        artifacts.append(
            ReplayArtifact.model_validate(
                {
                    "artifact_id": payload["artifact_id"],
                    "source_tag": payload["source_tag"],
                    "artifact_type": payload["artifact_type"],
                    "stage": payload["stage"],
                    "identifier": payload["identifier"],
                    "saved_at": payload["saved_at"],
                    "metadata": payload.get("metadata", {}),
                }
            )
        )
    return artifacts
