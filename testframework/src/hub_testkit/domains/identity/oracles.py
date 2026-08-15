from __future__ import annotations

import json
from datetime import datetime

from hub_testkit.http import HttpExchange

from .models import MeIdentity, SessionIdentity, TokenPair


class IdentityOracleError(ValueError):
    pass


_USER_REQUIRED = frozenset(
    {"id", "email", "timezone", "language", "system_role", "status", "onboarding_completed"}
)
_USER_OPTIONAL_TEXT = frozenset(
    {
        "name",
        "display_name",
        "avatar_url",
        "telegram_id",
        "phone",
        "position",
        "department",
    }
)
_USER_ALLOWED = _USER_REQUIRED | _USER_OPTIONAL_TEXT | frozenset({"birthday", "custom_fields"})
_SESSION_REQUIRED = frozenset({"id", "device", "last_active_at", "created_at", "is_current"})
_SESSION_ALLOWED = _SESSION_REQUIRED | frozenset({"ip_address"})


def parse_token_pair(exchange: HttpExchange) -> TokenPair:
    try:
        document = _document(exchange, 200)
        if type(document) is not dict or set(document) != {
            "access_token",
            "refresh_token",
            "expires_in",
        }:
            raise IdentityOracleError("identity response is invalid")
        return TokenPair(
            access_token=document["access_token"],
            refresh_token=document["refresh_token"],
            expires_in=document["expires_in"],
        )
    except Exception:
        raise IdentityOracleError("identity response is invalid") from None


def parse_me_identity(exchange: HttpExchange) -> MeIdentity:
    try:
        document = _document(exchange, 200)
        if type(document) is not dict or not _USER_REQUIRED <= set(document):
            raise IdentityOracleError("identity response is invalid")
        if set(document) - _USER_ALLOWED:
            raise IdentityOracleError("identity response is invalid")
        for key in _USER_REQUIRED - {"onboarding_completed"}:
            _require_text(document[key])
        if type(document["onboarding_completed"]) is not bool:
            raise IdentityOracleError("identity response is invalid")
        for key in _USER_OPTIONAL_TEXT:
            if key in document:
                _require_text(document[key])
        if "birthday" in document and document["birthday"] is not None:
            _require_timestamp(document["birthday"])
        if "custom_fields" in document and type(document["custom_fields"]) is not dict:
            raise IdentityOracleError("identity response is invalid")
        return MeIdentity(
            id=document["id"],
            email=document["email"],
            status=document["status"],
        )
    except Exception:
        raise IdentityOracleError("identity response is invalid") from None


def parse_session_identities(exchange: HttpExchange) -> tuple[SessionIdentity, ...]:
    try:
        document = _document(exchange, 200)
        if type(document) is not list:
            raise IdentityOracleError("identity response is invalid")
        sessions: list[SessionIdentity] = []
        for item in document:
            if type(item) is not dict or set(item) - _SESSION_ALLOWED:
                raise IdentityOracleError("identity response is invalid")
            if set(item) != _SESSION_REQUIRED and set(item) != _SESSION_ALLOWED:
                raise IdentityOracleError("identity response is invalid")
            _require_session_device(item["device"])
            _require_timestamp(item["last_active_at"])
            _require_timestamp(item["created_at"])
            if "ip_address" in item:
                _require_text(item["ip_address"])
            sessions.append(
                SessionIdentity(
                    session_id=item["id"],
                    is_current=item["is_current"],
                )
            )
        if len({session.session_id for session in sessions}) != len(sessions):
            raise IdentityOracleError("identity response is invalid")
        return tuple(sessions)
    except Exception:
        raise IdentityOracleError("identity response is invalid") from None


def _document(exchange: object, expected_status: int) -> object:
    try:
        if type(exchange) is not HttpExchange or exchange.response.status_code != expected_status:
            raise IdentityOracleError("identity response is invalid")
        body = exchange.response.body
        if body.startswith(b"\xef\xbb\xbf"):
            raise IdentityOracleError("identity response is invalid")
        return json.loads(
            body.decode("utf-8", errors="strict"),
            object_pairs_hook=_unique_object,
        )
    except Exception:
        raise IdentityOracleError("identity response is invalid") from None


def _require_text(value: object) -> str:
    if type(value) is not str or not value.strip() or len(value) > 4096:
        raise IdentityOracleError("identity response is invalid")
    if any(ord(character) < 32 or ord(character) == 127 for character in value):
        raise IdentityOracleError("identity response is invalid")
    return value


def _require_timestamp(value: object) -> str:
    text = _require_text(value)
    try:
        parsed = datetime.fromisoformat(text.replace("Z", "+00:00"))
    except Exception:
        raise IdentityOracleError("identity response is invalid") from None
    if parsed.tzinfo is None:
        raise IdentityOracleError("identity response is invalid")
    return text


def _require_session_device(value: object) -> str:
    if type(value) is not str or len(value) > 4096:
        raise IdentityOracleError("identity response is invalid")
    if any(ord(character) < 32 or ord(character) == 127 for character in value):
        raise IdentityOracleError("identity response is invalid")
    return value


def _unique_object(pairs: list[tuple[str, object]]) -> dict[str, object]:
    document: dict[str, object] = {}
    for key, value in pairs:
        if type(key) is not str or key in document:
            raise IdentityOracleError("identity response is invalid")
        document[key] = value
    return document
