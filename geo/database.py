from __future__ import annotations

import asyncio
from typing import Any

import asyncpg

from config import get_settings


class Database:
    def __init__(self) -> None:
        self._pool: asyncpg.Pool | None = None
        self._postgis_version: str | None = None

    async def connect(self) -> None:
        if self._pool is not None:
            return

        settings = get_settings()
        last_error: Exception | None = None

        for attempt in range(1, settings.database_connect_retries + 1):
            try:
                self._pool = await asyncpg.create_pool(
                    dsn=settings.database_url,
                    min_size=1,
                    max_size=10,
                )

                async with self._pool.acquire() as connection:
                    self._postgis_version = await connection.fetchval(
                        "SELECT PostGIS_Version()"
                    )

                return
            except (asyncpg.PostgresError, OSError) as exc:
                last_error = exc
                if self._pool is not None:
                    await self._pool.close()
                    self._pool = None

                if attempt == settings.database_connect_retries:
                    break

                await asyncio.sleep(settings.database_connect_retry_delay_seconds)

        raise RuntimeError("Failed to connect to PostGIS database") from last_error

    async def close(self) -> None:
        if self._pool is None:
            return

        await self._pool.close()
        self._pool = None
        self._postgis_version = None

    @property
    def pool(self) -> asyncpg.Pool:
        if self._pool is None:
            raise RuntimeError("Database pool is not initialized")
        return self._pool

    @property
    def postgis_version(self) -> str:
        if self._postgis_version is None:
            raise RuntimeError("PostGIS version is not available")
        return self._postgis_version


database = Database()


async def get_pool() -> asyncpg.Pool:
    return database.pool


async def check_database_connection(pool: asyncpg.Pool) -> dict[str, Any]:
    async with pool.acquire() as connection:
        return await connection.fetchrow("SELECT 1 AS ok")
