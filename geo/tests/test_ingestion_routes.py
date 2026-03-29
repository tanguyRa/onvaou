from __future__ import annotations

from datetime import UTC, datetime
from pathlib import Path

from fastapi.testclient import TestClient

from config import get_settings
from main import app
from models import IngestionResult, ReplayResponse


def test_trigger_openagenda_route(monkeypatch):
    async def fake_ingest_openagenda():
        return IngestionResult(source_tag="openagenda", fetched=3, inserted=2, duplicates=1)

    monkeypatch.setattr("routers.ingestion.ingest_openagenda", fake_ingest_openagenda)

    client = TestClient(app)
    response = client.post("/ingestion/openagenda")

    assert response.status_code == 200
    assert response.json()["result"]["inserted"] == 2


def test_trigger_datagouv_route(monkeypatch):
    async def fake_ingest_datagouv():
        return IngestionResult(source_tag="datagouv", fetched=4, updated=1)

    monkeypatch.setattr("routers.ingestion.ingest_datagouv", fake_ingest_datagouv)

    client = TestClient(app)
    response = client.post("/ingestion/datagouv")

    assert response.status_code == 200
    assert response.json()["job"] == "datagouv"


def test_replay_route(monkeypatch):
    async def fake_replay_artifact(artifact_id: str):
        return ReplayResponse(
            artifact={
                "artifact_id": artifact_id,
                "source_tag": "openagenda",
                "artifact_type": "openagenda_raw",
                "stage": "normalize",
                "identifier": "oa-1",
                "saved_at": datetime.now(UTC),
                "metadata": {},
            },
            status="replayed",
            message="Replay completed",
            result=IngestionResult(source_tag="openagenda", fetched=1, inserted=1),
        )

    monkeypatch.setattr("routers.ingestion.replay_artifact", fake_replay_artifact)

    client = TestClient(app)
    response = client.post("/ingestion/replay", json={"artifact_id": "artifact-123"})

    assert response.status_code == 200
    assert response.json()["status"] == "replayed"


def test_debug_artifact_written(monkeypatch, tmp_path: Path):
    monkeypatch.setenv("INGESTION_DEBUG_DIR", str(tmp_path))
    get_settings.cache_clear()

    from services.debug import persist_debug_artifact, load_debug_artifact

    artifact_id = persist_debug_artifact(
        source_tag="openagenda",
        artifact_type="openagenda_raw",
        stage="normalize",
        identifier="oa-1",
        payload={"id": "oa-1"},
        metadata={"error": "boom"},
    )

    payload = load_debug_artifact(artifact_id)

    assert payload["artifact_id"] == artifact_id
    assert payload["metadata"]["error"] == "boom"

    get_settings.cache_clear()


def test_artifact_list_route(monkeypatch):
    monkeypatch.setattr(
        "routers.ingestion.list_debug_artifacts",
        lambda limit: [
            {
                "artifact_id": "artifact-1",
                "source_tag": "openagenda",
                "artifact_type": "openagenda_raw",
                "stage": "normalize",
                "identifier": "oa-1",
                "saved_at": datetime.now(UTC),
                "metadata": {},
            }
        ],
    )

    client = TestClient(app)
    response = client.get("/ingestion/artifacts?limit=1")

    assert response.status_code == 200
    assert response.json()[0]["artifact_id"] == "artifact-1"
