from __future__ import annotations

import hashlib
import uuid
from datetime import UTC, datetime

from pydantic import BaseModel, Field, field_validator, model_validator


class Event(BaseModel):
    event_id: uuid.UUID | None = None
    source_uid: str
    title: str
    description: str = ""
    start_dt: datetime
    end_dt: datetime | None = None
    location_name: str = ""
    address: str
    longitude: float
    latitude: float
    source_tag: str
    source_url: str

    @field_validator("title", "description", "location_name", "address", "source_uid")
    @classmethod
    def normalize_text(cls, value: str) -> str:
        return " ".join(value.split()).strip()

    @field_validator("source_tag")
    @classmethod
    def normalize_source_tag(cls, value: str) -> str:
        return value.strip().lower()

    @field_validator("start_dt", "end_dt")
    @classmethod
    def normalize_datetime(cls, value: datetime | None) -> datetime | None:
        if value is None:
            return None
        if value.tzinfo is None:
            return value.replace(tzinfo=UTC)
        return value.astimezone(UTC)

    def with_event_id(self) -> "Event":
        if self.event_id is not None:
            return self

        basis = f"{self.source_tag}:{self.source_uid}:{self.start_dt.isoformat()}"
        return self.model_copy(
            update={"event_id": uuid.uuid5(uuid.NAMESPACE_URL, basis)}
        )

    @property
    def content_hash(self) -> str:
        payload = f"{self.title.lower()}|{self.start_dt.isoformat()}|{self.source_tag}"
        return hashlib.sha256(payload.encode("utf-8")).hexdigest()

    @property
    def dedup_text(self) -> str:
        parts = [self.title.lower(), self.address.lower()]
        return " ".join(part for part in parts if part).strip()


class EventSearchParams(BaseModel):
    city: str | None = None
    lat: float | None = None
    lon: float | None = None
    radius_km: int = Field(default=5, ge=1, le=50)
    days: int = Field(default=28, ge=1, le=28)
    page: int = Field(default=1, ge=1)
    limit: int = Field(default=50, ge=1, le=200)

    @field_validator("city")
    @classmethod
    def normalize_city(cls, value: str | None) -> str | None:
        if value is None:
            return None

        normalized = " ".join(value.split()).strip()
        return normalized or None

    @model_validator(mode="after")
    def validate_location_inputs(self) -> "EventSearchParams":
        has_city = self.city is not None
        has_coordinates = self.lat is not None or self.lon is not None

        if has_city and has_coordinates:
            raise ValueError("city is mutually exclusive with lat/lon")

        if not has_city and self.lat is None and self.lon is None:
            raise ValueError("either city or lat/lon must be provided")

        if (self.lat is None) != (self.lon is None):
            raise ValueError("lat and lon must be provided together")

        return self


class EventSummary(BaseModel):
    event_id: uuid.UUID
    title: str
    start_dt: datetime
    address: str
    lat: float
    lon: float
    source_tag: str
    source_url: str


class EventDetail(EventSummary):
    description: str
    end_dt: datetime | None = None
    location_name: str = ""


class EventSearchResponse(BaseModel):
    total: int
    page: int
    results: list[EventSummary]


class CitySuggestion(BaseModel):
    name: str
    city: str
    postcode: str
    lat: float
    lon: float


class IngestionResult(BaseModel):
    source_tag: str
    fetched: int = 0
    inserted: int = 0
    updated: int = 0
    duplicates: int = 0
    failed: int = 0
    details: list[str] = Field(default_factory=list)


class OpenAgendaAgenda(BaseModel):
    uid: int
    slug: str = ""
    title: str
    description: str = ""
    official: bool = False
    updated_at: datetime | None = None
    upcoming_events: int = 0
    recently_added_events: int = 0
    source_url: str = ""
    last_payload_hash: str = ""
    last_seen_at: datetime | None = None
    last_fetch_attempt_at: datetime | None = None
    last_fetch_success_at: datetime | None = None
    last_event_sync_at: datetime | None = None
    last_event_batch_hash: str = ""
    last_fetch_error: str = ""

    @field_validator(
        "slug",
        "title",
        "description",
        "source_url",
        "last_payload_hash",
        "last_event_batch_hash",
        "last_fetch_error",
    )
    @classmethod
    def normalize_agenda_text(cls, value: str) -> str:
        return " ".join(value.split()).strip()

    @field_validator(
        "updated_at",
        "last_seen_at",
        "last_fetch_attempt_at",
        "last_fetch_success_at",
        "last_event_sync_at",
    )
    @classmethod
    def normalize_agenda_datetime(cls, value: datetime | None) -> datetime | None:
        if value is None:
            return None
        if value.tzinfo is None:
            return value.replace(tzinfo=UTC)
        return value.astimezone(UTC)
