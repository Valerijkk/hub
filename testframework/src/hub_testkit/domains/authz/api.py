from __future__ import annotations

from collections.abc import Callable
from dataclasses import dataclass, field
from typing import NoReturn, Protocol

from hub_testkit.http import Actor, HttpDiagnostic, HttpExchange, HttpRequest, HttpResponse

from .builders import (
    approve_pending_request,
    deactivate_request,
    me_request,
    pending_principals_request,
    refresh_request,
    registration_mode_request,
    reject_pending_request,
)


class AuthzApiError(ValueError):
    pass


class AuthzSender(Protocol):
    def send(self, request: HttpRequest) -> HttpExchange:
        """Send one validated authorization request through the injected boundary."""
        ...


_INVALID = object()
_RequestBuilder = Callable[..., HttpRequest]


@dataclass(frozen=True, slots=True)
class AuthzApi:
    sender: AuthzSender = field(repr=False)

    def __post_init__(self) -> None:
        valid = _sender_is_valid_safe(self.sender)
        if valid is True:
            return
        del valid
        _raise_sender_invalid()

    def registration_mode(self, actor: Actor | None, correlation_id: str) -> HttpExchange:
        result = _call_safe(self, registration_mode_request, (actor, correlation_id))
        del actor, correlation_id
        return _exchange_or_raise(result)

    def list_pending(self, actor: Actor, correlation_id: str) -> HttpExchange:
        result = _call_safe(self, pending_principals_request, (actor, correlation_id))
        del actor, correlation_id
        return _exchange_or_raise(result)

    def approve_pending(self, actor: Actor, user_id: str, correlation_id: str) -> HttpExchange:
        result = _call_safe(self, approve_pending_request, (actor, user_id, correlation_id))
        del actor, user_id, correlation_id
        return _exchange_or_raise(result)

    def reject_pending(self, actor: Actor, user_id: str, correlation_id: str) -> HttpExchange:
        result = _call_safe(self, reject_pending_request, (actor, user_id, correlation_id))
        del actor, user_id, correlation_id
        return _exchange_or_raise(result)

    def deactivate(self, actor: Actor, user_id: str, correlation_id: str) -> HttpExchange:
        result = _call_safe(self, deactivate_request, (actor, user_id, correlation_id))
        del actor, user_id, correlation_id
        return _exchange_or_raise(result)

    def me(self, actor: Actor, correlation_id: str) -> HttpExchange:
        result = _call_safe(self, me_request, (actor, correlation_id))
        del actor, correlation_id
        return _exchange_or_raise(result)

    def refresh(self, refresh_token: str, correlation_id: str) -> HttpExchange:
        result = _call_safe(self, refresh_request, (refresh_token, correlation_id))
        del refresh_token, correlation_id
        return _exchange_or_raise(result)

    def _send(self, request: HttpRequest) -> HttpExchange:
        result = _send_safe(self.sender, request)
        del request
        return _exchange_or_raise(result)


def _sender_is_valid_safe(sender: object) -> bool:
    try:
        return sender is not None and callable(getattr(sender, "send", None))
    except Exception:
        return False


def _call_safe(
    api: AuthzApi,
    builder: _RequestBuilder,
    arguments: tuple[object, ...],
) -> HttpExchange | object:
    try:
        request = builder(*arguments)
        result = api._send(request)
        del request
        return result
    except Exception:
        return _INVALID


def _send_safe(sender: object, request: object) -> HttpExchange | object:
    try:
        if type(request) is not HttpRequest:
            return _INVALID
        send = getattr(sender, "send", None)
        if not callable(send):
            return _INVALID
        exchange = send(request)
        if type(exchange) is not HttpExchange:
            return _INVALID
        diagnostic = exchange.diagnostic
        path = _public_diagnostic_path_safe(request.path)
        if type(path) is not str:
            return _INVALID
        result = HttpExchange(
            response=HttpResponse(
                status_code=exchange.response.status_code,
                body=exchange.response.body,
            ),
            diagnostic=HttpDiagnostic(
                method=diagnostic.method,
                path=path,
                status_code=diagnostic.status_code,
                attempts=diagnostic.attempts,
                retry_class=diagnostic.retry_class,
                correlation_id=diagnostic.correlation_id,
            ),
        )
        del exchange, diagnostic, path
        return result
    except Exception:
        return _INVALID


def _public_diagnostic_path_safe(path: object) -> str | object:
    try:
        if type(path) is not str:
            return _INVALID
        prefix = "/api/admin/users/"
        for action in ("approve", "reject", "deactivate"):
            suffix = f"/{action}"
            if path.startswith(prefix) and path.endswith(suffix):
                return f"/api/admin/users/{{user_id}}/{action}"
        return path
    except Exception:
        return _INVALID


def _exchange_or_raise(result: object) -> HttpExchange:
    if type(result) is HttpExchange:
        return result
    del result
    _raise_exchange_invalid()


def _raise_sender_invalid() -> NoReturn:
    raise AuthzApiError("authorization sender is invalid") from None


def _raise_exchange_invalid() -> NoReturn:
    raise AuthzApiError("authorization exchange is invalid") from None
