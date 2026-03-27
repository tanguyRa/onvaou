from fastapi import APIRouter, Depends
from asyncpg import Pool

from database import database, get_pool
from models import HealthResponse


router = APIRouter(tags=["health"])


@router.get("/health", response_model=HealthResponse)
async def healthcheck(pool: Pool = Depends(get_pool)) -> HealthResponse:
    async with pool.acquire() as connection:
        await connection.execute("SELECT 1")

    return HealthResponse(
        status="ok",
        db="connected",
        postgis=database.postgis_version,
    )
