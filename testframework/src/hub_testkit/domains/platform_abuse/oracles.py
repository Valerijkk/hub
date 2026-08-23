from __future__ import annotations

import json
from typing import NoReturn

from .models import OwnedVerifyScenario, PlatformAbuseExchange, TokenPair


class PlatformAbuseOracleError(ValueError):
    pass


_BODY_LIMIT_BYTES = 64 * 1024
_INVALID = object()


def parse_token_pair(exchange: PlatformAbuseExchange) -> TokenPair:
    result = _token_pair_safe(exchange)
    del exchange
    if type(result) is TokenPair:
        return result
    del result
    _raise_invalid()


def rate_limit_sequence_is_bounded(
    first: PlatformAbuseExchange,
    second: PlatformAbuseExchange,
    third: PlatformAbuseExchange,
    fourth: PlatformAbuseExchange,
    scenario: OwnedVerifyScenario,
) -> bool:
    result = _rate_limit_sequence_safe(first, second, third, fourth, scenario)
    del first, second, third, fourth, scenario
    if result is True:
        return True
    del result
    _raise_invalid()


def verify_replay_is_rejected(
    first: PlatformAbuseExchange,
    replay: PlatformAbuseExchange,
    scenario: OwnedVerifyScenario,
) -> bool:
    result = _replay_is_rejected_safe(first, replay, scenario)
    del first, replay, scenario
    if result is True:
        return True
    del result
    _raise_invalid()


def _token_pair_safe(exchange: object) -> TokenPair | object:
    try:
        document = _document_safe(exchange, 200)
        if type(document) is not dict or set(document) != {
            "access_token",
            "refresh_token",
            "expires_in",
        }:
            return _INVALID
        result = TokenPair(
            access_token=document["access_token"],
            refresh_token=document["refresh_token"],
            expires_in=document["expires_in"],
        )
        del document
        return result
    except Exception:
        return _INVALID


def _rate_limit_sequence_safe(
    first: object,
    second: object,
    third: object,
    fourth: object,
    scenario: object,
) -> bool:
    try:
        if not _scenario_is_rate_limited_safe(scenario):
            return False
        statuses = (
            _status_safe(first),
            _status_safe(second),
            _status_safe(third),
            _status_safe(fourth),
        )
        retry_after = (
            _retry_after_safe(first),
            _retry_after_safe(second),
            _retry_after_safe(third),
            _retry_after_safe(fourth),
        )
        if statuses != (401, 401, 401, 429) or retry_after != (None, None, None, "600"):
            return False
        return all(
            _error_body_hides_scenario_safe(exchange, scenario)
            and _diagnostic_hides_scenario_safe(exchange, scenario)
            for exchange in (first, second, third, fourth)
        )
    except Exception:
        return False


def _replay_is_rejected_safe(
    first: object,
    replay: object,
    scenario: object,
) -> bool:
    try:
        if not _scenario_is_replay_safe(scenario):
            return False
        pair = _token_pair_safe(first)
        if type(pair) is not TokenPair or not _token_pair_repr_is_safe(pair):
            return False
        if _status_safe(replay) != 401 or _retry_after_safe(replay) is not None:
            return False
        result = (
            _replay_error_body_hides_values_safe(replay, scenario, pair)
            and _diagnostic_hides_scenario_safe(first, scenario)
            and _diagnostic_hides_scenario_safe(replay, scenario)
        )
        del pair
        return result
    except Exception:
        return False


def _document_safe(exchange: object, expected_status: object) -> dict[str, object] | object:
    try:
        if type(exchange) is not PlatformAbuseExchange or type(expected_status) is not int:
            return _INVALID
        response = exchange.exchange.response
        if response.status_code != expected_status:
            return _INVALID
        body = response.body
        if len(body) > _BODY_LIMIT_BYTES or body.startswith(b"\xef\xbb\xbf"):
            return _INVALID
        document = json.loads(
            body.decode("utf-8", errors="strict"),
            object_pairs_hook=_unique_object_safe,
        )
        del body
        if type(document) is not dict:
            return _INVALID
        return document
    except Exception:
        return _INVALID


def _status_safe(exchange: object) -> int | object:
    try:
        if type(exchange) is not PlatformAbuseExchange:
            return _INVALID
        return exchange.exchange.response.status_code
    except Exception:
        return _INVALID


def _retry_after_safe(exchange: object) -> str | None | object:
    try:
        if type(exchange) is not PlatformAbuseExchange:
            return _INVALID
        return exchange.retry_after
    except Exception:
        return _INVALID


def _scenario_is_rate_limited_safe(scenario: object) -> bool:
    try:
        return type(scenario) is OwnedVerifyScenario and scenario.flow == "unknown-code-rate-limit"
    except Exception:
        return False


def _scenario_is_replay_safe(scenario: object) -> bool:
    try:
        return type(scenario) is OwnedVerifyScenario and scenario.flow == "one-time-code-replay"
    except Exception:
        return False


def _token_pair_repr_is_safe(pair: object) -> bool:
    try:
        if type(pair) is not TokenPair:
            return False
        rendered = f"{pair!r}{pair!s}"
        return pair.access_token not in rendered and pair.refresh_token not in rendered
    except Exception:
        return False


def _diagnostic_hides_scenario_safe(exchange: object, scenario: object) -> bool:
    try:
        if (
            type(exchange) is not PlatformAbuseExchange
            or type(scenario) is not OwnedVerifyScenario
        ):
            return False
        diagnostic = exchange.exchange.diagnostic
        rendered = f"{diagnostic!r}{diagnostic!s}{exchange!r}{exchange!s}"
        raw_values = (
            scenario.email,
            scenario.code,
            scenario.nonce,
            scenario.server_observed_client,
            scenario.server_observed_client_hash,
        )
        return all(value not in rendered for value in raw_values)
    except Exception:
        return False


def _error_body_hides_scenario_safe(exchange: object, scenario: object) -> bool:
    try:
        if (
            type(exchange) is not PlatformAbuseExchange
            or type(scenario) is not OwnedVerifyScenario
        ):
            return False
        if exchange.exchange.response.status_code not in {401, 429}:
            return False
        body = exchange.exchange.response.body
        values = (
            scenario.email,
            scenario.code,
            scenario.nonce,
            scenario.server_observed_client,
            scenario.server_observed_client_hash,
        )
        result = all(value.encode("utf-8") not in body for value in values)
        del body, values
        return result
    except Exception:
        return False


def _replay_error_body_hides_values_safe(
    replay: object,
    scenario: object,
    pair: object,
) -> bool:
    try:
        if (
            type(replay) is not PlatformAbuseExchange
            or type(scenario) is not OwnedVerifyScenario
            or type(pair) is not TokenPair
            or replay.exchange.response.status_code != 401
        ):
            return False
        body = replay.exchange.response.body
        values = (
            scenario.email,
            scenario.code,
            scenario.nonce,
            scenario.server_observed_client,
            scenario.server_observed_client_hash,
            pair.access_token,
            pair.refresh_token,
        )
        result = all(value.encode("utf-8") not in body for value in values)
        del body, values
        return result
    except Exception:
        return False


def _unique_object_safe(pairs: object) -> dict[str, object] | object:
    try:
        if type(pairs) is not list:
            return _INVALID
        document: dict[str, object] = {}
        for pair in pairs:
            if type(pair) is not tuple or len(pair) != 2:
                return _INVALID
            key, value = pair
            if type(key) is not str or key in document:
                return _INVALID
            document[key] = value
        return document
    except Exception:
        return _INVALID


def _raise_invalid() -> NoReturn:
    raise PlatformAbuseOracleError("platform abuse response is invalid") from None
