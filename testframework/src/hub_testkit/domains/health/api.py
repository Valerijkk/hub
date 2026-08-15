from __future__ import annotations

import json

from hub_testkit.http.policy import HttpPolicy
from hub_testkit.http.transport import HttpClientProtocol, HttpRequest, HttpxTransport
from hub_testkit.runtime.redaction import Redactor

from .models import HealthResponse, HealthTransport, LiveHealth
from .oracles import parse_live_health


class HealthTransportError(ValueError):
    pass


def get_live_health(transport: HealthTransport) -> LiveHealth:
    try:
        response = transport.request("GET", "/api/live", headers={})
    except Exception:
        raise HealthTransportError("live health transport failed") from None
    return parse_live_health(response)


def get_live_health_via_http(
    client: HttpClientProtocol,
    *,
    correlation_id: str,
    redactor: Redactor,
    policy: HttpPolicy = HttpPolicy(),
) -> LiveHealth:
    try:
        exchange = HttpxTransport(
            client=client,
            policy=policy,
            redactor=redactor,
        ).send(
            HttpRequest(
                method="GET",
                path="/api/live",
                actor=None,
                correlation_id=correlation_id,
                idempotency_key=None,
                body=None,
            )
        )
        body = exchange.response.body
        if body.startswith(b"\xef\xbb\xbf"):
            raise ValueError("health payload is invalid")
        document = json.loads(body.decode("utf-8", errors="strict"))
        if type(document) is not dict:
            raise ValueError("health payload is invalid")
    except Exception:
        raise HealthTransportError("live health response is invalid") from None
    return parse_live_health(
        HealthResponse(status_code=exchange.response.status_code, body=document)
    )
