from __future__ import annotations

from typing import Literal

from pydantic import AnyHttpUrl, BaseModel, ConfigDict, Field, model_validator


ArtifactLevel = Literal["failure", "trace", "debug"]


class AllowedEnvironment(BaseModel):
    model_config = ConfigDict(extra="forbid", frozen=True, strict=True)

    id: str = Field(pattern=r"^[a-z][a-z0-9-]{2,31}$")
    base_urls: tuple[AnyHttpUrl, ...] = Field(min_length=1)
    artifact_levels: tuple[ArtifactLevel, ...] = Field(min_length=1)
    allow_dev_endpoints: bool

    @model_validator(mode="after")
    def _require_unique_values(self) -> AllowedEnvironment:
        if len({str(url) for url in self.base_urls}) != len(self.base_urls):
            raise ValueError("environment base URLs must be unique")
        if len(set(self.artifact_levels)) != len(self.artifact_levels):
            raise ValueError("environment artifact levels must be unique")
        return self


class EnvironmentPolicy(BaseModel):
    model_config = ConfigDict(extra="forbid", frozen=True, strict=True)

    schema: Literal["hub.testframework.environments/v1"]
    environments: tuple[AllowedEnvironment, ...] = Field(min_length=1)

    @model_validator(mode="after")
    def _require_unique_environment_ids(self) -> EnvironmentPolicy:
        if len({environment.id for environment in self.environments}) != len(self.environments):
            raise ValueError("environment IDs must be unique")
        return self


class HubSettings(BaseModel):
    model_config = ConfigDict(extra="forbid", frozen=True, strict=True)

    environment_id: str = Field(pattern=r"^[a-z][a-z0-9-]{2,31}$")
    base_url: AnyHttpUrl
    resource_prefix: Literal["hubqa"]
    artifact_level: ArtifactLevel
    allow_dev_endpoints: bool
