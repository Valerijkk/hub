from __future__ import annotations

import json
from uuid import UUID

from hub_testkit.http import Actor, HttpRequest

from .models import IdentityModelError, SessionCredential


class IdentityBuilderError(ValueError):
    pass


def actor_from_session(user_id: str, credential: SessionCredential) -> Actor:
    try:
        if type(credential) is not SessionCredential or str(UUID(user_id)) != user_id:
            raise IdentityBuilderError("identity request is invalid")
        return Actor(subject=user_id, token=credential.access_token)
    except Exception:
        raise IdentityBuilderError("identity request is invalid") from None


def refresh_request(refresh_token: str, correlation_id: str) -> HttpRequest:
    try:
        _require_token(refresh_token)
        body = json.dumps(
            {"refresh_token": refresh_token},
            ensure_ascii=True,
            separators=(",", ":"),
        ).encode("utf-8")
        return HttpRequest(
            method="POST",
            path="/api/auth/refresh",
            actor=None,
            correlation_id=correlation_id,
            idempotency_key=None,
            body=body,
            content_type="application/json",
        )
    except Exception:
        raise IdentityBuilderError("identity request is invalid") from None


def me_request(actor: Actor | None, correlation_id: str) -> HttpRequest:
    return _request("GET", "/api/auth/me", actor, correlation_id)


def logout_request(actor: Actor, correlation_id: str) -> HttpRequest:
    return _request("POST", "/api/auth/logout", actor, correlation_id)


def sessions_request(actor: Actor, correlation_id: str) -> HttpRequest:
    return _request("GET", "/api/auth/sessions", actor, correlation_id)


def revoke_session_request(actor: Actor, session_id: str, correlation_id: str) -> HttpRequest:
    try:
        if str(UUID(session_id)) != session_id:
            raise IdentityBuilderError("identity request is invalid")
        return _request(
            "DELETE",
            f"/api/auth/sessions/{session_id}",
            actor,
            correlation_id,
        )
    except Exception:
        raise IdentityBuilderError("identity request is invalid") from None


def _request(
    method: str,
    path: str,
    actor: Actor | None,
    correlation_id: str,
) -> HttpRequest:
    try:
        return HttpRequest(
            method=method,
            path=path,
            actor=actor,
            correlation_id=correlation_id,
            idempotency_key=None,
            body=None,
            content_type=None,
        )
    except Exception:
        raise IdentityBuilderError("identity request is invalid") from None


def _require_token(value: object) -> str:
    try:
        return SessionCredential(
            session_id="00000000-0000-4000-8000-000000000000",
            access_token="placeholder-access-token",
            refresh_token=value,
        ).refresh_token
    except Exception:
        raise IdentityModelError("identity model is invalid") from None
