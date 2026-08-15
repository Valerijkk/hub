from __future__ import annotations

import re
from dataclasses import dataclass, field
from datetime import datetime
from typing import Literal
from uuid import UUID


class AuthzModelError(ValueError):
    pass


_LABEL_RE = re.compile(r"^[a-z][a-z0-9-]{2,31}$")
_NONCE_RE = re.compile(r"^[0-9a-f]{32}$")
_KINDS = frozenset({"founder", "admin", "member"})
_STATES = frozenset({"active", "pending"})


@dataclass(frozen=True, slots=True)
class PrincipalSeed:
    label: str
    kind: Literal["founder", "admin", "member"]
    state: Literal["active", "pending"]
    credential_count: int

    def __post_init__(self) -> None:
        try:
            if type(self.label) is not str or _LABEL_RE.fullmatch(self.label) is None:
                raise AuthzModelError("authorization seed is invalid")
            _require_kind(self.kind)
            _require_state(self.state)
            if type(self.credential_count) is not int or self.credential_count not in {0, 1}:
                raise AuthzModelError("authorization seed is invalid")
            if self.state == "active" and self.credential_count != 1:
                raise AuthzModelError("authorization seed is invalid")
            if self.state == "pending" and self.credential_count != 0:
                raise AuthzModelError("authorization seed is invalid")
        except Exception:
            raise AuthzModelError("authorization seed is invalid") from None


@dataclass(frozen=True, slots=True)
class PrincipalCredential:
    access_token: str = field(repr=False)
    refresh_token: str = field(repr=False)

    def __post_init__(self) -> None:
        try:
            _require_text(self.access_token, 4096)
            _require_text(self.refresh_token, 4096)
        except Exception:
            raise AuthzModelError("authorization credential is invalid") from None


@dataclass(frozen=True, slots=True)
class ProvisionedPrincipal:
    user_id: str = field(repr=False)
    email: str = field(repr=False)
    kind: Literal["founder", "admin", "member"]
    state: Literal["active", "pending"]
    credentials: tuple[PrincipalCredential, ...] = field(repr=False)
    nonce: str = field(repr=False)

    def __post_init__(self) -> None:
        try:
            _require_uuid(self.user_id)
            _require_text(self.email, 320)
            _require_kind(self.kind)
            _require_state(self.state)
            _require_nonce(self.nonce)
            if type(self.credentials) is not tuple or len(self.credentials) > 1:
                raise AuthzModelError("provisioned principal is invalid")
            credentials = tuple(
                PrincipalCredential(
                    access_token=credential.access_token,
                    refresh_token=credential.refresh_token,
                )
                for credential in self.credentials
                if type(credential) is PrincipalCredential
            )
            if len(credentials) != len(self.credentials):
                raise AuthzModelError("provisioned principal is invalid")
            if self.state == "active" and len(credentials) != 1:
                raise AuthzModelError("provisioned principal is invalid")
            if self.state == "pending" and credentials != ():
                raise AuthzModelError("provisioned principal is invalid")
            object.__setattr__(self, "credentials", credentials)
        except Exception:
            raise AuthzModelError("provisioned principal is invalid") from None


@dataclass(frozen=True, slots=True)
class RegistrationMode:
    mode: Literal["open", "invite_only"]

    def __post_init__(self) -> None:
        try:
            if type(self.mode) is not str or self.mode not in {"open", "invite_only"}:
                raise AuthzModelError("registration mode is invalid")
        except Exception:
            raise AuthzModelError("registration mode is invalid") from None


@dataclass(frozen=True, slots=True)
class PendingPrincipal:
    user_id: str = field(repr=False)
    email: str = field(repr=False)
    pending_since: str

    def __post_init__(self) -> None:
        try:
            _require_uuid(self.user_id)
            _require_text(self.email, 320)
            _require_timestamp(self.pending_since)
        except Exception:
            raise AuthzModelError("pending principal is invalid") from None


@dataclass(frozen=True, slots=True)
class ActionMessage:
    message: Literal["user approved", "user rejected"]

    def __post_init__(self) -> None:
        try:
            if type(self.message) is not str or self.message not in {"user approved", "user rejected"}:
                raise AuthzModelError("authorization action is invalid")
        except Exception:
            raise AuthzModelError("authorization action is invalid") from None


@dataclass(frozen=True, slots=True)
class ManagementSuccess:
    success: bool

    def __post_init__(self) -> None:
        try:
            if type(self.success) is not bool:
                raise AuthzModelError("authorization action is invalid")
        except Exception:
            raise AuthzModelError("authorization action is invalid") from None


@dataclass(frozen=True, slots=True)
class ErrorShape:
    status_code: int
    canonical_message: str = field(repr=False)

    def __post_init__(self) -> None:
        try:
            if type(self.status_code) is not int or not 400 <= self.status_code <= 599:
                raise AuthzModelError("authorization error shape is invalid")
            _require_text(self.canonical_message, 1024)
        except Exception:
            raise AuthzModelError("authorization error shape is invalid") from None


def _require_uuid(value: object) -> str:
    if type(value) is not str:
        raise AuthzModelError("authorization model is invalid")
    try:
        if str(UUID(value)) != value:
            raise AuthzModelError("authorization model is invalid")
    except Exception:
        raise AuthzModelError("authorization model is invalid") from None
    return value


def _require_nonce(value: object) -> str:
    if type(value) is not str or _NONCE_RE.fullmatch(value) is None:
        raise AuthzModelError("authorization model is invalid")
    return value


def _require_kind(value: object) -> str:
    if type(value) is not str or value not in _KINDS:
        raise AuthzModelError("authorization model is invalid")
    return value


def _require_state(value: object) -> str:
    if type(value) is not str or value not in _STATES:
        raise AuthzModelError("authorization model is invalid")
    return value


def _require_text(value: object, maximum: int) -> str:
    if type(value) is not str or not value.strip() or len(value) > maximum:
        raise AuthzModelError("authorization model is invalid")
    if any(ord(character) < 32 or ord(character) == 127 for character in value):
        raise AuthzModelError("authorization model is invalid")
    return value


def _require_timestamp(value: object) -> str:
    text = _require_text(value, 128)
    try:
        parsed = datetime.fromisoformat(text.replace("Z", "+00:00"))
    except Exception:
        raise AuthzModelError("authorization model is invalid") from None
    if parsed.tzinfo is None:
        raise AuthzModelError("authorization model is invalid")
    return text
