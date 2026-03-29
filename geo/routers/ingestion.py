from __future__ import annotations

import logging

import httpx
from fastapi import APIRouter, HTTPException, status

from models import ReplayArtifact, ReplayRequest, ReplayResponse, TriggerResponse
from scheduler import datagouv_job, openagenda_job, vacuum_job
from services.debug import list_debug_artifacts
from services.ingestion import ingest_datagouv, ingest_openagenda
from services.replay import replay_artifact

logger = logging.getLogger(__name__)
router = APIRouter(prefix="/ingestion", tags=["ingestion"])


@router.post("/openagenda", response_model=TriggerResponse)
async def trigger_openagenda() -> TriggerResponse:
    try:
        result = await ingest_openagenda()
    except httpx.HTTPStatusError as exc:
        logger.warning(
            "OpenAgenda ingestion failed with upstream status %s",
            exc.response.status_code,
        )
        raise HTTPException(
            status_code=status.HTTP_502_BAD_GATEWAY,
            detail=f"OpenAgenda upstream error: {exc.response.status_code}",
        ) from exc

    return TriggerResponse(job="openagenda", result=result)


@router.post("/datagouv", response_model=TriggerResponse)
async def trigger_datagouv() -> TriggerResponse:
    try:
        result = await ingest_datagouv()
    except httpx.HTTPStatusError as exc:
        logger.warning(
            "data.gouv ingestion failed with upstream status %s",
            exc.response.status_code,
        )
        raise HTTPException(
            status_code=status.HTTP_502_BAD_GATEWAY,
            detail=f"data.gouv upstream error: {exc.response.status_code}",
        ) from exc

    return TriggerResponse(job="datagouv", result=result)


@router.post("/jobs/openagenda", response_model=dict[str, str])
async def trigger_openagenda_job() -> dict[str, str]:
    await openagenda_job()
    return {"job": "openagenda_job", "status": "completed"}


@router.post("/jobs/datagouv", response_model=dict[str, str])
async def trigger_datagouv_job() -> dict[str, str]:
    await datagouv_job()
    return {"job": "datagouv_job", "status": "completed"}


@router.post("/vacuum", response_model=dict[str, str])
async def trigger_vacuum() -> dict[str, str]:
    await vacuum_job()
    return {"job": "vacuum_job", "status": "completed"}


@router.post("/replay", response_model=ReplayResponse)
async def trigger_replay(request: ReplayRequest) -> ReplayResponse:
    try:
        return await replay_artifact(request.artifact_id)
    except FileNotFoundError as exc:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=str(exc),
        ) from exc


@router.get("/artifacts", response_model=list[ReplayArtifact])
async def get_artifacts(limit: int = 20) -> list[ReplayArtifact]:
    return list_debug_artifacts(limit=max(1, min(limit, 100)))
