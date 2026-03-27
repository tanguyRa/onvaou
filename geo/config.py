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
    openagenda_api_key: str = Field(default="", alias="OPENAGENDA_API_KEY")
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
