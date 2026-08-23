from __future__ import annotations

from dataclasses import dataclass, field
from typing import NoReturn, Protocol

from hub_testkit.http import HttpDiagnostic, HttpExchange, HttpRequest, HttpResponse

from .builders import verify_magic_link_request
from .models import OwnedVerifyScenario, PlatformAbuseExchange


class PlatformAbuseApiError(ValueError):
    pass


class PlatformAbuseSender(Protocol):
    def send(self, request: HttpRequest) -> PlatformAbuseExchange:
        """Send one public verification request through the injected runtime boundary."""
        ...


_INVALID = object()


@dataclass(frozen=True, slots=True)
class PlatformAbuseApi:
    sender: PlatformAbuseSender = field(repr=False)

    def __post_init__(self) -> None:
        valid = _sender_is_valid_safe(self.sender)
        if valid is True:
            return
        del valid
        _raise_sender_invalid()

    def verify(self, scenario: OwnedVerifyScenario, correlation_id: str) -> PlatformAbuseExchange:
        result = _verify_safe(self, scenario, correlation_id)
        del scenario, correlation_id
        if type(result) is PlatformAbuseExchange:
            return result
        del result
        _raise_exchange_invalid()

    def _send(self, request: HttpRequest) -> PlatformAbuseExchange:
        result = _send_safe(self.sender, request)
        del request
        if type(result) is PlatformAbuseExchange:
            return result
        del result
        _raise_exchange_invalid()


def _sender_is_valid_safe(sender: object) -> bool:
    try:
        return sender is not None and callable(getattr(sender, "send", None))
    except Exception:
        return False


def _verify_safe(
    api: PlatformAbuseApi,
    scenario: object,
    correlation_id: object,
) -> PlatformAbuseExchange | object:
    try:
        if type(scenario) is not OwnedVerifyScenario or type(correlation_id) is not str:
            return _INVALID
        request = verify_magic_link_request(scenario, correlation_id)
        result = api._send(request)
        del request
        return result
    except Exception:
        return _INVALID


def _send_safe(sender: object, request: object) -> PlatformAbuseExchange | object:
    try:
        if type(request) is not HttpRequest:
            return _INVALID
        send = getattr(sender, "send", None)
        if not callable(send):
            return _INVALID
        exchange = send(request)
        if type(exchange) is not PlatformAbuseExchange:
            return _INVALID
        response = exchange.exchange.response
        diagnostic = exchange.exchange.diagnostic
        normalized = PlatformAbuseExchange(
            exchange=HttpExchange(
                response=HttpResponse(status_code=response.status_code, body=response.body),
                diagnostic=HttpDiagnostic(
                    method=diagnostic.method,
                    path=diagnostic.path,
                    status_code=diagnostic.status_code,
                    attempts=diagnostic.attempts,
                    retry_class=diagnostic.retry_class,
                    correlation_id=diagnostic.correlation_id,
                ),
            ),
            retry_after=exchange.retry_after,
        )
        del exchange, response, diagnostic
        return normalized
    except Exception:
        return _INVALID


def _raise_sender_invalid() -> NoReturn:
    raise PlatformAbuseApiError("platform abuse sender is invalid") from None


def _raise_exchange_invalid() -> NoReturn:
    raise PlatformAbuseApiError("platform abuse exchange is invalid") from None
