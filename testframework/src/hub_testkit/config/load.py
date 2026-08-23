from __future__ import annotations

import tomllib
from collections.abc import Mapping
from pathlib import Path

from pydantic import AnyHttpUrl, TypeAdapter, ValidationError

from .models import AllowedEnvironment, EnvironmentPolicy, HubSettings


class ConfigurationError(ValueError):
    pass


_ENVIRONMENT_POLICY_KEYS = frozenset({"schema", "environments"})
_ENVIRONMENT_KEYS = frozenset(
    {"id", "base_urls", "artifact_levels", "allow_dev_endpoints"}
)
_SETTINGS_KEYS = frozenset(
    {"environment_id", "base_url", "resource_prefix", "artifact_level"}
)


def load_environment_policy(path: str | Path) -> EnvironmentPolicy:
    try:
        with Path(path).open("rb") as stream:
            document = tomllib.load(stream)
    except (OSError, UnicodeDecodeError, tomllib.TOMLDecodeError):
        raise ConfigurationError("environment policy could not be loaded") from None

    try:
        _require_exact_keys(document, _ENVIRONMENT_POLICY_KEYS)
        schema = _require_string(document["schema"])
        raw_environments = _require_list(document["environments"])
        if not raw_environments:
            raise ConfigurationError("environment policy must list environments")
        environments = tuple(_parse_environment(item) for item in raw_environments)
        return EnvironmentPolicy(schema=schema, environments=environments)
    except (KeyError, ValidationError):
        raise ConfigurationError("environment policy is invalid") from None


def load_settings(raw: Mapping[str, object], policy: EnvironmentPolicy) -> HubSettings:
    if not isinstance(raw, Mapping) or not isinstance(policy, EnvironmentPolicy):
        raise ConfigurationError("settings and policy must use the supported types")
    _require_exact_keys(raw, _SETTINGS_KEYS)
    try:
        environment_id = _require_string(raw["environment_id"])
        base_url = _parse_http_url(_require_string(raw["base_url"]))
        resource_prefix = _require_string(raw["resource_prefix"])
        artifact_level = _require_string(raw["artifact_level"])
    except KeyError:
        raise ConfigurationError("settings are incomplete") from None

    environment = next(
        (
            candidate
            for candidate in policy.environments
            if candidate.id == environment_id
        ),
        None,
    )
    if environment is None:
        raise ConfigurationError("environment is not admitted")
    if str(base_url) not in {str(candidate) for candidate in environment.base_urls}:
        raise ConfigurationError("base URL is not admitted")
    if artifact_level not in environment.artifact_levels:
        raise ConfigurationError("artifact level is not admitted")

    try:
        return HubSettings(
            environment_id=environment_id,
            base_url=base_url,
            resource_prefix=resource_prefix,
            artifact_level=artifact_level,
            allow_dev_endpoints=environment.allow_dev_endpoints,
        )
    except ValidationError:
        raise ConfigurationError("settings are invalid") from None


def _parse_environment(value: object) -> AllowedEnvironment:
    if not isinstance(value, Mapping):
        raise ConfigurationError("environment entry is invalid")
    _require_exact_keys(value, _ENVIRONMENT_KEYS)
    try:
        raw_urls = _require_list(value["base_urls"])
        raw_levels = _require_list(value["artifact_levels"])
        if not raw_urls or not raw_levels:
            raise ConfigurationError("environment entry must not be empty")
        return AllowedEnvironment(
            id=_require_string(value["id"]),
            base_urls=tuple(_parse_http_url(_require_string(url)) for url in raw_urls),
            artifact_levels=tuple(_require_string(level) for level in raw_levels),
            allow_dev_endpoints=_require_bool(value["allow_dev_endpoints"]),
        )
    except (KeyError, ValidationError):
        raise ConfigurationError("environment entry is invalid") from None


def _parse_http_url(value: str) -> AnyHttpUrl:
    try:
        return TypeAdapter(AnyHttpUrl).validate_python(value)
    except ValidationError:
        raise ConfigurationError("base URL is invalid") from None


def _require_exact_keys(value: Mapping[str, object], expected: frozenset[str]) -> None:
    if set(value) != expected:
        raise ConfigurationError("mapping has unsupported keys")


def _require_list(value: object) -> list[object]:
    if type(value) is not list:
        raise ConfigurationError("value must be a list")
    return value


def _require_string(value: object) -> str:
    if type(value) is not str or not value:
        raise ConfigurationError("value must be a non-empty string")
    return value


def _require_bool(value: object) -> bool:
    if type(value) is not bool:
        raise ConfigurationError("value must be a boolean")
    return value
