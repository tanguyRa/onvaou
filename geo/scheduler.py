from __future__ import annotations

import logging
from zoneinfo import ZoneInfo

from apscheduler.schedulers.asyncio import AsyncIOScheduler
from apscheduler.triggers.cron import CronTrigger

from config import get_settings
from services.ingestion import ingest_datagouv, ingest_openagenda
from services.store import vacuum_events

logger = logging.getLogger(__name__)

settings = get_settings()
scheduler = AsyncIOScheduler(timezone=ZoneInfo(settings.scheduler_timezone))


async def openagenda_job() -> None:
    result = await ingest_openagenda()
    logger.info("openagenda_job result: %s", result.model_dump())


async def datagouv_job() -> None:
    result = await ingest_datagouv()
    logger.info("datagouv_job result: %s", result.model_dump())


async def vacuum_job() -> None:
    await vacuum_events()
    logger.info("vacuum_job completed")


def start_scheduler() -> None:
    if scheduler.running:
        return

    scheduler.add_job(
        openagenda_job,
        trigger=CronTrigger(hour=3, minute=0),
        id="openagenda_job",
        replace_existing=True,
    )
    scheduler.add_job(
        datagouv_job,
        trigger=CronTrigger(hour=3, minute=30),
        id="datagouv_job",
        replace_existing=True,
    )
    scheduler.add_job(
        vacuum_job,
        trigger=CronTrigger(day_of_week="sun", hour=4, minute=0),
        id="vacuum_job",
        replace_existing=True,
    )
    scheduler.start()


async def stop_scheduler() -> None:
    if scheduler.running:
        scheduler.shutdown(wait=False)
