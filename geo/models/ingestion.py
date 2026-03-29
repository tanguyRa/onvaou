from __future__ import annotations

from datetime import datetime
from typing import Any

from pydantic import BaseModel, Field

from models.event import IngestionResult


class TriggerResponse(BaseModel):
    job: str
    result: IngestionResult


class ReplayRequest(BaseModel):
    artifact_id: str = Field(min_length=1)


class ReplayArtifact(BaseModel):
    artifact_id: str
    source_tag: str
    artifact_type: str
    stage: str
    identifier: str
    saved_at: datetime
    metadata: dict[str, Any] = Field(default_factory=dict)


class ReplayResponse(BaseModel):
    artifact: ReplayArtifact
    status: str
    message: str
    result: IngestionResult | None = None
