from __future__ import annotations

from datetime import UTC, datetime
from uuid import uuid4

from fastapi.testclient import TestClient

from database import get_pool
from main import app


class FakeAcquire:
    def __init__(self, connection):
        self.connection = connection

    async def __aenter__(self):
        return self.connection

    async def __aexit__(self, exc_type, exc, tb):
        return False


class FakePool:
    def __init__(self, connection):
        self.connection = connection

    def acquire(self):
        return FakeAcquire(self.connection)


class FakeEventsConnection:
    def __init__(self, *, total=0, rows=None, row=None):
        self.total = total
        self.rows = rows or []
        self.row = row
        self.fetchval_calls = []
        self.fetch_calls = []
        self.fetchrow_calls = []

    async def fetchval(self, query, *args):
        self.fetchval_calls.append((query, args))
        return self.total

    async def fetch(self, query, *args):
        self.fetch_calls.append((query, args))
        return self.rows

    async def fetchrow(self, query, *args):
        self.fetchrow_calls.append((query, args))
        return self.row


def test_list_events_with_city(monkeypatch):
    connection = FakeEventsConnection(
        total=2,
        rows=[
            {
                "event_id": uuid4(),
                "title": "Marche locale",
                "start_dt": datetime(2025, 4, 1, 14, 0, tzinfo=UTC),
                "address": "Place Bellecour, Lyon",
                "lat": 45.75,
                "lon": 4.83,
                "source_tag": "openagenda",
                "source_url": "https://example.test/events/1",
            }
        ],
    )

    async def fake_get_pool():
        return FakePool(connection)

    async def fake_resolve_city(city: str):
        assert city == "Lyon"
        return 4.83, 45.75

    app.dependency_overrides[get_pool] = fake_get_pool
    monkeypatch.setattr("routers.events.resolve_city", fake_resolve_city)

    client = TestClient(app)
    response = client.get("/events", params={"city": "Lyon", "radius_km": 10})

    assert response.status_code == 200
    assert response.json()["total"] == 2
    assert response.json()["results"][0]["title"] == "Marche locale"
    assert connection.fetchval_calls[0][1] == (4.83, 45.75, 10000, 28)
    assert connection.fetch_calls[0][1] == (4.83, 45.75, 10000, 28, 50, 0)

    app.dependency_overrides.clear()


def test_list_events_rejects_invalid_location_mix():
    client = TestClient(app)
    response = client.get("/events", params={"city": "Lyon", "lat": 45.75, "lon": 4.83})

    assert response.status_code == 422


def test_list_events_rejects_invalid_radius():
    client = TestClient(app)
    response = client.get("/events", params={"city": "Lyon", "radius_km": 51})

    assert response.status_code == 422


def test_get_event_detail():
    event_id = uuid4()
    connection = FakeEventsConnection(
        row={
            "event_id": event_id,
            "title": "Concert",
            "description": "Soiree acoustique",
            "start_dt": datetime(2025, 4, 2, 18, 0, tzinfo=UTC),
            "end_dt": datetime(2025, 4, 2, 20, 0, tzinfo=UTC),
            "location_name": "Salle des fetes",
            "address": "1 rue de la Republique, Lyon",
            "lat": 45.76,
            "lon": 4.84,
            "source_tag": "openagenda",
            "source_url": "https://example.test/events/2",
        }
    )

    async def fake_get_pool():
        return FakePool(connection)

    app.dependency_overrides[get_pool] = fake_get_pool

    client = TestClient(app)
    response = client.get(f"/events/{event_id}")

    assert response.status_code == 200
    assert response.json()["event_id"] == str(event_id)
    assert connection.fetchrow_calls[0][1] == (event_id,)

    app.dependency_overrides.clear()


def test_get_event_detail_returns_404():
    connection = FakeEventsConnection(row=None)

    async def fake_get_pool():
        return FakePool(connection)

    app.dependency_overrides[get_pool] = fake_get_pool

    client = TestClient(app)
    response = client.get(f"/events/{uuid4()}")

    assert response.status_code == 404
    assert response.json()["detail"] == "Event not found"

    app.dependency_overrides.clear()


def test_city_search(monkeypatch):
    async def fake_search_cities(query: str, *, limit: int = 5):
        assert query == "Bord"
        assert limit == 5
        return [
            {
                "name": "Bordeaux",
                "city": "Bordeaux",
                "postcode": "33000",
                "lat": 44.84,
                "lon": -0.58,
            }
        ]

    monkeypatch.setattr("routers.events.search_cities", fake_search_cities)

    client = TestClient(app)
    response = client.get("/cities/search", params={"q": "Bord"})

    assert response.status_code == 200
    assert response.json()[0]["city"] == "Bordeaux"
