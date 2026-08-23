from __future__ import annotations

import traceback
from collections.abc import Mapping

import pytest

from hub_testkit.domains.health.api import HealthTransportError, get_live_health
from hub_testkit.domains.health.models import HealthResponse, LiveHealth
from hub_testkit.domains.health.oracles import HealthOracleError


class _RecordingTransport:
    def __init__(self, response: HealthResponse) -> None:
        self.calls: list[tuple[str, str, dict[str, str]]] = []
        self.response = response

    def request(
        self,
        method: str,
        path: str,
        *,
        headers: Mapping[str, str],
    ) -> HealthResponse:
        self.calls.append((method, path, dict(headers)))
        return self.response


class _ExplodingTransport:
    def __init__(self, canary: str) -> None:
        self.canary = canary

    def request(
        self,
        method: str,
        path: str,
        *,
        headers: Mapping[str, str],
    ) -> HealthResponse:
        raise RuntimeError(self.canary)


class _HostileValue:
    def __init__(self, canary: str) -> None:
        self.canary = canary

    def __repr__(self) -> str:
        return self.canary


@pytest.mark.case_id("TC-HEALTH-0001")
@pytest.mark.api
def test_live_health_uses_fixed_anonymous_probe_and_returns_parsed_result() -> None:
    transport = _RecordingTransport(HealthResponse(status_code=200, body={"status": "ok"}))

    result = get_live_health(transport)

    assert result == LiveHealth(status="ok")
    assert transport.calls == [("GET", "/api/live", {})]
    assert "Authorization" not in transport.calls[0][2]

    for response in (
        HealthResponse(status_code=201, body={"status": "ok"}),
        HealthResponse(status_code="200", body={"status": "ok"}),
        HealthResponse(status_code=True, body={"status": "ok"}),
        HealthResponse(status_code=200, body=["status", "ok"]),
        HealthResponse(status_code=200, body={}),
        HealthResponse(status_code=200, body={"status": "degraded"}),
        HealthResponse(status_code=200, body={"status": False}),
        HealthResponse(status_code=200, body={"status": "ok", "version": "1.0"}),
        HealthResponse(status_code=200, body={"status": "ok", "readiness": "ready"}),
    ):
        _assert_oracle_rejection(response)

    for body in (
        _HostileValue("HEALTH-BODY-CANARY"),
        {"status": _HostileValue("HEALTH-STATUS-CANARY")},
    ):
        _assert_hostile_body_is_hidden(body)

    _assert_transport_failure_is_hidden()


def _assert_oracle_rejection(response: HealthResponse) -> None:
    with pytest.raises(HealthOracleError) as raised:
        get_live_health(_RecordingTransport(response))

    assert str(raised.value) == "live health response is invalid"
    assert raised.value.__cause__ is None
    assert raised.value.__suppress_context__ is True


def _assert_hostile_body_is_hidden(body: object) -> None:
    canary = _canary_from_body(body)

    with pytest.raises(HealthOracleError) as raised:
        get_live_health(_RecordingTransport(HealthResponse(status_code=200, body=body)))

    assert str(raised.value) == "live health response is invalid"
    assert raised.value.__cause__ is None
    assert raised.value.__suppress_context__ is True
    assert canary not in "".join(traceback.format_exception(raised.value))


def _assert_transport_failure_is_hidden() -> None:
    canary = "HEALTH-TRANSPORT-CANARY"

    with pytest.raises(HealthTransportError) as raised:
        get_live_health(_ExplodingTransport(canary))

    assert str(raised.value) == "live health transport failed"
    assert raised.value.__cause__ is None
    assert raised.value.__suppress_context__ is True
    assert canary not in "".join(traceback.format_exception(raised.value))


def _canary_from_body(body: object) -> str:
    if isinstance(body, _HostileValue):
        return body.canary
    if type(body) is not dict:
        raise AssertionError("hostile test body is invalid")
    value = body["status"]
    if not isinstance(value, _HostileValue):
        raise AssertionError("hostile test status is invalid")
    return value.canary
