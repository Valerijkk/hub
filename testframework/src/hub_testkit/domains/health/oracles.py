from __future__ import annotations

from .models import HealthResponse, LiveHealth


class HealthOracleError(ValueError):
    pass


def parse_live_health(response: object) -> LiveHealth:
    try:
        if type(response) is not HealthResponse:
            raise HealthOracleError("live health response is invalid")
        if type(response.status_code) is not int or response.status_code != 200:
            raise HealthOracleError("live health response is invalid")
        body = response.body
        if type(body) is not dict or len(body) != 1:
            raise HealthOracleError("live health response is invalid")
        key, status = next(iter(body.items()))
        if type(key) is not str or key != "status":
            raise HealthOracleError("live health response is invalid")
        if type(status) is not str or status != "ok":
            raise HealthOracleError("live health response is invalid")
        return LiveHealth(status="ok")
    except Exception:
        raise HealthOracleError("live health response is invalid") from None
