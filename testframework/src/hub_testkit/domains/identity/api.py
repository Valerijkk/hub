from __future__ import annotations

from dataclasses import dataclass, field
from typing import Protocol

from hub_testkit.http import Actor, HttpExchange, HttpRequest

from .builders import (
    logout_request,
    me_request,
    refresh_request,
    revoke_session_request,
    sessions_request,
)


class IdentityApiError(ValueError):
    pass


class IdentitySender(Protocol):
    def send(self, request: HttpRequest) -> HttpExchange:
        """Send one already validated request through the injected boundary."""
        ...


@dataclass(frozen=True, slots=True)
class IdentityApi:
    sender: IdentitySender = field(repr=False)

    def __post_init__(self) -> None:
        try:
            if self.sender is None or not callable(getattr(self.sender, "send", None)):
                raise IdentityApiError("identity sender is invalid")
        except Exception:
            raise IdentityApiError("identity sender is invalid") from None

    def refresh(self, refresh_token: str, correlation_id: str) -> HttpExchange:
        return self._send(refresh_request(refresh_token, correlation_id))

    def me(self, actor: Actor | None, correlation_id: str) -> HttpExchange:
        return self._send(me_request(actor, correlation_id))

    def logout(self, actor: Actor, correlation_id: str) -> HttpExchange:
        return self._send(logout_request(actor, correlation_id))

    def list_sessions(self, actor: Actor, correlation_id: str) -> HttpExchange:
        return self._send(sessions_request(actor, correlation_id))

    def revoke_session(
        self,
        actor: Actor,
        session_id: str,
        correlation_id: str,
    ) -> HttpExchange:
        return self._send(revoke_session_request(actor, session_id, correlation_id))

    def _send(self, request: HttpRequest) -> HttpExchange:
        try:
            exchange = self.sender.send(request)
            if type(exchange) is not HttpExchange:
                raise IdentityApiError("identity exchange is invalid")
            return exchange
        except Exception:
            raise IdentityApiError("identity exchange is invalid") from None
