from __future__ import annotations

import json
from typing import NoReturn

from hub_testkit.http import HttpRequest

from .models import OwnedVerifyScenario, scenario_has_canonical_client_hash


class PlatformAbuseBuilderError(ValueError):
    pass


_INVALID = object()


def verify_magic_link_request(
    scenario: OwnedVerifyScenario,
    correlation_id: str,
) -> HttpRequest:
    result = _verify_magic_link_request_safe(scenario, correlation_id)
    del scenario, correlation_id
    if type(result) is HttpRequest:
        return result
    del result
    _raise_invalid()


def _verify_magic_link_request_safe(
    scenario: object,
    correlation_id: object,
) -> HttpRequest | object:
    try:
        if type(scenario) is not OwnedVerifyScenario or not _scenario_is_admitted_safe(scenario):
            return _INVALID
        if type(correlation_id) is not str:
            return _INVALID
        body = json.dumps(
            {"email": scenario.email, "code": scenario.code},
            ensure_ascii=True,
            separators=(",", ":"),
        ).encode("utf-8")
        result = HttpRequest(
            method="POST",
            path="/api/auth/magic-link/verify",
            actor=None,
            correlation_id=correlation_id,
            idempotency_key=None,
            body=body,
            content_type="application/json",
        )
        del body
        return result
    except Exception:
        return _INVALID


def _scenario_is_admitted_safe(scenario: object) -> bool:
    try:
        return (
            type(scenario) is OwnedVerifyScenario
            and (
                scenario.flow != "one-time-code-replay"
                or scenario.identity_active is True
            )
            and scenario.delivery_attempted is False
            and scenario.development_login_used is False
            and scenario.dev_mode is False
            and scenario.redis_healthy is True
            and scenario.database_healthy is True
            and scenario.redis_exclusive is True
            and scenario.database_exclusive is True
            and scenario.flow in {"unknown-code-rate-limit", "one-time-code-replay"}
            and scenario_has_canonical_client_hash(scenario) is True
            and _owned_entries_match_safe(scenario)
        )
    except Exception:
        return False


def _raise_invalid() -> NoReturn:
    raise PlatformAbuseBuilderError("platform abuse request is invalid") from None


def _owned_entries_match_safe(scenario: object) -> bool:
    try:
        if type(scenario) is not OwnedVerifyScenario:
            return False
        expected = (
            f"rl:auth:ip:{scenario.server_observed_client_hash}",
            f"rl:auth-verify-email:email:{scenario.email}",
            f"magiclink:fails:{scenario.email}",
            f"magiclink:{scenario.email}",
        )
        return scenario.owned_state_entries == expected
    except Exception:
        return False
