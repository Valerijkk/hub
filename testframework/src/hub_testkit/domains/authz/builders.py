from __future__ import annotations

import json
from typing import NoReturn
from uuid import UUID

from hub_testkit.http import Actor, HttpRequest

from .models import PrincipalCredential, ProvisionedPrincipal


class AuthzBuilderError(ValueError):
    pass


_INVALID = object()


def actor_from_principal(principal: ProvisionedPrincipal) -> Actor:
    result = _actor_from_principal_safe(principal)
    del principal
    return _actor_or_raise(result)


def registration_mode_request(actor: Actor | None, correlation_id: str) -> HttpRequest:
    result = _request("GET", "/api/admin/registration-mode", actor, correlation_id)
    del actor, correlation_id
    return _request_or_raise(result)


def pending_principals_request(actor: Actor, correlation_id: str) -> HttpRequest:
    result = _request("GET", "/api/admin/users/pending", actor, correlation_id)
    del actor, correlation_id
    return _request_or_raise(result)


def approve_pending_request(actor: Actor, user_id: str, correlation_id: str) -> HttpRequest:
    result = _principal_action_request(actor, user_id, "approve", correlation_id)
    del actor, user_id, correlation_id
    return _request_or_raise(result)


def reject_pending_request(actor: Actor, user_id: str, correlation_id: str) -> HttpRequest:
    result = _principal_action_request(actor, user_id, "reject", correlation_id)
    del actor, user_id, correlation_id
    return _request_or_raise(result)


def deactivate_request(actor: Actor, user_id: str, correlation_id: str) -> HttpRequest:
    result = _principal_action_request(actor, user_id, "deactivate", correlation_id)
    del actor, user_id, correlation_id
    return _request_or_raise(result)


def me_request(actor: Actor, correlation_id: str) -> HttpRequest:
    result = _request("GET", "/api/auth/me", actor, correlation_id)
    del actor, correlation_id
    return _request_or_raise(result)


def refresh_request(refresh_token: str, correlation_id: str) -> HttpRequest:
    result = _refresh_request_safe(refresh_token, correlation_id)
    del refresh_token, correlation_id
    return _request_or_raise(result)


def _actor_from_principal_safe(principal: object) -> Actor | object:
    try:
        if type(principal) is not ProvisionedPrincipal or principal.state != "active":
            return _INVALID
        if len(principal.credentials) != 1:
            return _INVALID
        credential = principal.credentials[0]
        result = Actor(subject=principal.user_id, token=credential.access_token)
        del credential
        return result
    except Exception:
        return _INVALID


def _refresh_request_safe(refresh_token: object, correlation_id: object) -> HttpRequest | object:
    try:
        if type(correlation_id) is not str:
            return _INVALID
        token = _require_token(refresh_token)
        if type(token) is not str:
            return _INVALID
        body = json.dumps(
            {"refresh_token": token},
            ensure_ascii=True,
            separators=(",", ":"),
        ).encode("utf-8")
        result = HttpRequest(
            method="POST",
            path="/api/auth/refresh",
            actor=None,
            correlation_id=correlation_id,
            idempotency_key=None,
            body=body,
            content_type="application/json",
        )
        del token, body
        return result
    except Exception:
        return _INVALID


def _principal_action_request(
    actor: object,
    user_id: object,
    action: object,
    correlation_id: object,
) -> HttpRequest | object:
    try:
        if type(action) is not str or action not in {"approve", "reject", "deactivate"}:
            return _INVALID
        if type(user_id) is not str or str(UUID(user_id)) != user_id:
            return _INVALID
        path = f"/api/admin/users/{user_id}/{action}"
        result = _request("POST", path, actor, correlation_id)
        del path
        return result
    except Exception:
        return _INVALID


def _request(
    method: object,
    path: object,
    actor: object,
    correlation_id: object,
) -> HttpRequest | object:
    try:
        if type(method) is not str or type(path) is not str or type(correlation_id) is not str:
            return _INVALID
        if actor is not None and type(actor) is not Actor:
            return _INVALID
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
        return _INVALID


def _require_token(value: object) -> str | object:
    try:
        if type(value) is not str:
            return _INVALID
        token: str = value
        result = PrincipalCredential(
            access_token="placeholder-access-token",
            refresh_token=token,
        ).refresh_token
        del token
        return result
    except Exception:
        return _INVALID


def _actor_or_raise(result: object) -> Actor:
    if type(result) is Actor:
        return result
    del result
    _raise_invalid()


def _request_or_raise(result: object) -> HttpRequest:
    if type(result) is HttpRequest:
        return result
    del result
    _raise_invalid()


def _raise_invalid() -> NoReturn:
    raise AuthzBuilderError("authorization request is invalid") from None
