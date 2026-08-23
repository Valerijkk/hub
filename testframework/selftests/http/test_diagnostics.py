from __future__ import annotations

import dataclasses
import traceback
from pathlib import Path

import pytest

from hub_testkit.http.diagnostics import diagnostic_fields
from hub_testkit.http.errors import HttpDiagnosticError, HttpTransportError
from hub_testkit.http.transport import HttpDiagnostic, HttpExchange, HttpResponse
from hub_testkit.runtime.redaction import Redactor


REDACTION_POLICY_PATH = Path(__file__).resolve().parents[2] / "policies" / "redaction.toml"


def _diagnostic() -> HttpDiagnostic:
    return HttpDiagnostic(
        method="GET",
        path="/api/resource",
        status_code=200,
        attempts=1,
        retry_class="none",
        correlation_id="trace-001",
    )


def test_diagnostic_fields_are_immutable_redacted_and_closed() -> None:
    diagnostic = _diagnostic()
    redactor = Redactor.from_policy(REDACTION_POLICY_PATH)
    fields = diagnostic_fields(diagnostic, redactor)

    assert dict(fields) == {
        "method": "GET",
        "path": "/api/resource",
        "status_code": 200,
        "attempts": 1,
        "retry_class": "none",
        "correlation_id": "trace-001",
    }
    with pytest.raises(TypeError):
        fields["actor"] = "forbidden"  # type: ignore[index]
    with pytest.raises(dataclasses.FrozenInstanceError):
        diagnostic.attempts = 2
    exchange = HttpExchange(
        response=HttpResponse(status_code=200, body=b"{}"),
        diagnostic=diagnostic,
    )
    with pytest.raises(dataclasses.FrozenInstanceError):
        exchange.response = HttpResponse(status_code=201, body=b"{}")
    with pytest.raises(HttpTransportError):
        HttpExchange(response=object(), diagnostic=diagnostic)
    redacted = diagnostic_fields(
        HttpDiagnostic(
            method="GET",
            path="/api/Bearer DIAGNOSTIC-TOKEN-CANARY",
            status_code=200,
            attempts=1,
            retry_class="none",
            correlation_id="trace-001",
        ),
        redactor,
    )
    assert redacted["path"] == "/api/[REDACTED]"
    assert "DIAGNOSTIC-TOKEN-CANARY" not in repr(redacted)


def test_diagnostic_rejects_bypass_and_redaction_canaries_without_leaking() -> None:
    canary = "DIAGNOSTIC-SECRET-CANARY"
    bypass = object.__new__(HttpDiagnostic)
    object.__setattr__(bypass, "method", "GET")
    object.__setattr__(bypass, "path", "/api/resource")
    object.__setattr__(bypass, "status_code", 200)
    object.__setattr__(bypass, "attempts", 1)
    object.__setattr__(bypass, "retry_class", "none")
    object.__setattr__(bypass, "correlation_id", f"trace\n{canary}")

    with pytest.raises(HttpDiagnosticError) as raised:
        diagnostic_fields(bypass, Redactor.from_policy(REDACTION_POLICY_PATH))

    assert str(raised.value) == "HTTP diagnostic is invalid"
    assert raised.value.__cause__ is None
    assert raised.value.__suppress_context__ is True
    assert canary not in "".join(traceback.format_exception(raised.value))

    for factory in (
        lambda: HttpDiagnostic(
            method="GET",
            path="/api/resource",
            status_code=200,
            attempts=2,
            retry_class="none",
            correlation_id="trace-001",
        ),
        lambda: HttpDiagnostic(
            method="GET",
            path="/api/resource",
            status_code=200,
            attempts=True,
            retry_class="none",
            correlation_id="trace-001",
        ),
    ):
        with pytest.raises(HttpDiagnosticError):
            factory()
