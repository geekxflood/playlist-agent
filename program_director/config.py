"""Configuration management for the Playlist Agent."""

import os
from pathlib import Path
from typing import Any

import yaml
from pydantic import BaseModel, Field


class ThemeConfig(BaseModel):
    """Theme configuration for playlist generation."""

    name: str
    description: str
    duration: int = Field(default=180, description="Duration in minutes")
    keywords: list[str] = Field(default_factory=list, description="Genre/theme keywords")


class OllamaConfig(BaseModel):
    """Ollama LLM configuration."""

    url: str = Field(default="http://localhost:11434")
    model: str = Field(default="llama3:8b")


class ErsatzTVConfig(BaseModel):
    """ErsatzTV API configuration."""

    url: str = Field(default="http://localhost:8409")


class RadarrConfig(BaseModel):
    """Radarr API configuration."""

    url: str = Field(default="http://localhost:7878")
    api_key: str = Field(default="")


class SonarrConfig(BaseModel):
    """Sonarr API configuration."""

    url: str = Field(default="http://localhost:8989")
    api_key: str = Field(default="")


class AgentConfig(BaseModel):
    """Main agent configuration."""

    ollama: OllamaConfig = Field(default_factory=OllamaConfig)
    ersatztv: ErsatzTVConfig = Field(default_factory=ErsatzTVConfig)
    radarr: RadarrConfig = Field(default_factory=RadarrConfig)
    sonarr: SonarrConfig = Field(default_factory=SonarrConfig)
    themes: list[ThemeConfig] = Field(default_factory=list)


def load_config(config_path: Path | None = None) -> AgentConfig:
    """Load configuration from file or environment variables."""
    config_data: dict[str, Any] = {}

    # Try loading from config file
    if config_path and config_path.exists():
        with open(config_path) as f:
            config_data = yaml.safe_load(f) or {}

    # Override with environment variables
    if ollama_url := os.getenv("OLLAMA_URL"):
        config_data.setdefault("ollama", {})["url"] = ollama_url
    if ollama_model := os.getenv("OLLAMA_MODEL"):
        config_data.setdefault("ollama", {})["model"] = ollama_model
    if ersatztv_url := os.getenv("ERSATZTV_URL"):
        config_data.setdefault("ersatztv", {})["url"] = ersatztv_url
    if radarr_url := os.getenv("RADARR_URL"):
        config_data.setdefault("radarr", {})["url"] = radarr_url
    if radarr_api_key := os.getenv("RADARR_API_KEY"):
        config_data.setdefault("radarr", {})["api_key"] = radarr_api_key
    if sonarr_url := os.getenv("SONARR_URL"):
        config_data.setdefault("sonarr", {})["url"] = sonarr_url
    if sonarr_api_key := os.getenv("SONARR_API_KEY"):
        config_data.setdefault("sonarr", {})["api_key"] = sonarr_api_key

    return AgentConfig(**config_data)
