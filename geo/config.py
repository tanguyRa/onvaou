from functools import lru_cache

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )

    app_name: str = "onvaou-geo"
    database_url: str = Field(..., alias="DATABASE_URL")
    ban_api_url: str = Field(
        default="https://api-adresse.data.gouv.fr",
        alias="BAN_API_URL",
    )
    datagouv_api_url: str = Field(
        default="https://www.data.gouv.fr/api/1",
        alias="DATAGOUV_API_URL",
    )
    openagenda_api_url: str = Field(
        default="https://api.openagenda.com/v2",
        alias="OPENAGENDA_API_URL",
    )
    openagenda_api_key: str = Field(default="", alias="OPENAGENDA_API_KEY")
    openagenda_official_only: bool = Field(
        default=True,
        alias="OPENAGENDA_OFFICIAL_ONLY",
    )
    openagenda_agenda_updated_within_days: int = Field(
        default=30,
        alias="OPENAGENDA_AGENDA_UPDATED_WITHIN_DAYS",
    )
    openagenda_max_agendas: int = Field(default=0, alias="OPENAGENDA_MAX_AGENDAS")
    http_timeout_seconds: float = Field(default=20.0, alias="HTTP_TIMEOUT_SECONDS")
    ban_retry_attempts: int = Field(default=3, alias="BAN_RETRY_ATTEMPTS")
    scheduler_timezone: str = Field(
        default="Europe/Paris",
        alias="SCHEDULER_TIMEZONE",
    )
    log_level: str = Field(default="INFO", alias="LOG_LEVEL")
    ingestion_debug_dir: str = Field(
        default="data/ingestion-debug",
        alias="INGESTION_DEBUG_DIR",
    )
    cors_allowed_origins: str = Field(default="", alias="CORS_ALLOWED_ORIGINS")
    database_connect_retries: int = Field(
        default=15,
        alias="DATABASE_CONNECT_RETRIES",
    )
    database_connect_retry_delay_seconds: float = Field(
        default=1.0,
        alias="DATABASE_CONNECT_RETRY_DELAY_SECONDS",
    )

    @property
    def cors_origins(self) -> list[str]:
        return [
            origin.strip()
            for origin in self.cors_allowed_origins.split(",")
            if origin.strip()
        ]


@lru_cache
def get_settings() -> Settings:
    return Settings()
