from __future__ import annotations

from pathlib import Path

import pytest
from pydantic import ValidationError

from hub_testkit.config import ConfigurationError, load_environment_policy, load_settings


PROJECT_ROOT = Path(__file__).resolve().parents[3]
POLICY_PATH = PROJECT_ROOT / "policies" / "environments.toml"
POLICY_HEADER = 'schema = "hub.testframework.environments/v1"'
LOCAL_ENVIRONMENT = "\n".join(
    (
        "[[environments]]",
        'id = "docker-local"',
        'base_urls = ["http://hub-api:8080/"]',
        'artifact_levels = ["failure"]',
        "allow_dev_endpoints = true",
    )
)


def _policy():
    return load_environment_policy(POLICY_PATH)


def _raw_settings(**overrides: object) -> dict[str, object]:
    raw: dict[str, object] = {
        "environment_id": "docker-local",
        "base_url": "http://hub-api:8080",
        "resource_prefix": "hubqa",
        "artifact_level": "debug",
    }
    raw.update(overrides)
    return raw


def test_environment_policy_has_only_deterministic_docker_admission() -> None:
    policy = _policy()

    assert policy.schema == "hub.testframework.environments/v1"
    assert [environment.id for environment in policy.environments] == [
        "docker-local",
        "docker-ci",
        "docker-integration",
    ]
    assert {
        environment.id: (
            tuple(str(url) for url in environment.base_urls),
            environment.artifact_levels,
            environment.allow_dev_endpoints,
        )
        for environment in policy.environments
    } == {
        "docker-local": (("http://hub-api:8080/",), ("failure", "trace", "debug"), True),
        "docker-ci": (("http://hub-api:8080/",), ("failure", "trace"), False),
        "docker-integration": (("http://hub-api:8080/",), ("failure", "trace"), False),
    }


def test_settings_are_normalized_immutable_and_derive_environment_capability() -> None:
    settings = load_settings(_raw_settings(), _policy())

    assert str(settings.base_url) == "http://hub-api:8080/"
    assert settings.resource_prefix == "hubqa"
    assert settings.allow_dev_endpoints is True
    with pytest.raises(ValidationError):
        settings.resource_prefix = "other"


@pytest.mark.parametrize(
    "raw",
    [
        _raw_settings(extra="forbidden"),
        _raw_settings(environment_id=True),
        _raw_settings(base_url=8080),
        _raw_settings(resource_prefix="HubQA"),
        _raw_settings(resource_prefix="other"),
        _raw_settings(artifact_level="verbose"),
    ],
)
def test_settings_reject_unknown_keys_and_non_string_input(raw: dict[str, object]) -> None:
    with pytest.raises(ConfigurationError):
        load_settings(raw, _policy())


@pytest.mark.parametrize(
    "raw",
    [
        _raw_settings(environment_id="production"),
        _raw_settings(base_url="https://hub.example.test"),
        _raw_settings(base_url="http://localhost:8080"),
        _raw_settings(environment_id="docker-ci", artifact_level="debug"),
    ],
)
def test_settings_reject_production_unknown_or_mismatched_admission(raw: dict[str, object]) -> None:
    with pytest.raises(ConfigurationError):
        load_settings(raw, _policy())


def test_settings_error_does_not_echo_a_rejected_url_canary() -> None:
    canary = "http://secret-url-canary@["

    with pytest.raises(ConfigurationError) as error:
        load_settings(_raw_settings(base_url=canary), _policy())

    assert canary not in str(error.value)
    assert error.value.__cause__ is None


@pytest.mark.parametrize(
    "contents",
    [
        "\n".join((POLICY_HEADER, "unexpected = true", "", LOCAL_ENVIRONMENT)),
        "\n".join((POLICY_HEADER, "", LOCAL_ENVIRONMENT, "unexpected = true")),
        "\n".join((POLICY_HEADER, "", LOCAL_ENVIRONMENT, "", LOCAL_ENVIRONMENT)),
    ],
)
def test_environment_policy_rejects_unknown_or_duplicate_structure(
    tmp_path: Path,
    contents: str,
) -> None:
    path = tmp_path / "environments.toml"
    path.write_text(contents, encoding="utf-8")

    with pytest.raises(ConfigurationError):
        load_environment_policy(path)


def test_environment_policy_rejects_invalid_utf8_without_echoing_bytes(tmp_path: Path) -> None:
    path = tmp_path / "environments.toml"
    canary = b"secret-policy-canary"
    path.write_bytes(b"schema = \"hub.testframework.environments/v1\"\n" + canary + b"\xff")

    with pytest.raises(ConfigurationError) as error:
        load_environment_policy(path)

    assert canary.decode("ascii") not in str(error.value)
    assert error.value.__cause__ is None
