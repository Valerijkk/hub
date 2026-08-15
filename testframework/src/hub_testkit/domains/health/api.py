from __future__ import annotations

from .models import HealthTransport, LiveHealth
from .oracles import parse_live_health


class HealthTransportError(ValueError):
    pass


def get_live_health(transport: HealthTransport) -> LiveHealth:
    try:
        response = transport.request("GET", "/api/live", headers={})
    except Exception:
        raise HealthTransportError("live health transport failed") from None
    return parse_live_health(response)
