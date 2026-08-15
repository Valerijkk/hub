from __future__ import annotations

import re
from dataclasses import dataclass, field
from typing import Literal
from uuid import UUID


class IdentityModelError(ValueError):
    pass


_LABEL_RE = re.compile(r"^[a-z][a-z0-9-]{2,31}$")
_STATUSES = frozenset({"active", "deactivated"})


@dataclass(frozen=True, slots=True)
class IdentitySeed:
    label: str
    status: Literal["active", "deactivated"]
    session_count: int

    def __post_init__(self) -> None:
        try:
            if type(self.label) is not str or _LABEL_RE.fullmatch(self.label) is None:
                raise IdentityModelError("identity seed is invalid")
            _require_status(self.status)
            if type(self.session_count) is not int or not 1 <= self.session_count <= 4:
                raise IdentityModelError("identity seed is invalid")
        except Exception:
            raise IdentityModelError("identity seed is invalid") from None


@dataclass(frozen=True, slots=True)
class SessionCredential:
    session_id: str
    access_token: str = field(repr=False)
    refresh_token: str = field(repr=False)

    def __post_init__(self) -> None:
        try:
            _require_canonical_uuid(self.session_id)
            _require_text(self.access_token, 4096)
            _require_text(self.refresh_token, 4096)
        except Exception:
            raise IdentityModelError("session credential is invalid") from None


@dataclass(frozen=True, slots=True)
class ProvisionedIdentity:
    user_id: str
    email: str = field(repr=False)
    status: Literal["active", "deactivated"]
    sessions: tuple[SessionCredential, ...]

    def __post_init__(self) -> None:
        try:
            _require_canonical_uuid(self.user_id)
            _require_text(self.email, 320)
            _require_status(self.status)
            if type(self.sessions) is not tuple or not self.sessions:
                raise IdentityModelError("provisioned identity is invalid")
            sessions = tuple(
                SessionCredential(
                    session_id=session.session_id,
                    access_token=session.access_token,
                    refresh_token=session.refresh_token,
                )
                for session in self.sessions
                if type(session) is SessionCredential
            )
            if len(sessions) != len(self.sessions):
                raise IdentityModelError("provisioned identity is invalid")
            if len({session.session_id for session in sessions}) != len(sessions):
                raise IdentityModelError("provisioned identity is invalid")
            object.__setattr__(self, "sessions", sessions)
        except Exception:
            raise IdentityModelError("provisioned identity is invalid") from None


@dataclass(frozen=True, slots=True)
class TokenPair:
    access_token: str = field(repr=False)
    refresh_token: str = field(repr=False)
    expires_in: int

    def __post_init__(self) -> None:
        try:
            _require_text(self.access_token, 4096)
            _require_text(self.refresh_token, 4096)
            if type(self.expires_in) is not int or self.expires_in <= 0:
                raise IdentityModelError("token pair is invalid")
        except Exception:
            raise IdentityModelError("token pair is invalid") from None


@dataclass(frozen=True, slots=True)
class MeIdentity:
    id: str
    email: str = field(repr=False)
    status: Literal["active", "deactivated"]

    def __post_init__(self) -> None:
        try:
            _require_canonical_uuid(self.id)
            _require_text(self.email, 320)
            _require_status(self.status)
        except Exception:
            raise IdentityModelError("identity projection is invalid") from None


@dataclass(frozen=True, slots=True)
class SessionIdentity:
    session_id: str
    is_current: bool

    def __post_init__(self) -> None:
        try:
            _require_canonical_uuid(self.session_id)
            if type(self.is_current) is not bool:
                raise IdentityModelError("session projection is invalid")
        except Exception:
            raise IdentityModelError("session projection is invalid") from None


def _require_canonical_uuid(value: object) -> str:
    if type(value) is not str:
        raise IdentityModelError("identity model is invalid")
    try:
        if str(UUID(value)) != value:
            raise IdentityModelError("identity model is invalid")
    except Exception:
        raise IdentityModelError("identity model is invalid") from None
    return value


def _require_status(value: object) -> str:
    if type(value) is not str or value not in _STATUSES:
        raise IdentityModelError("identity model is invalid")
    return value


def _require_text(value: object, maximum: int) -> str:
    if type(value) is not str or not value.strip() or len(value) > maximum:
        raise IdentityModelError("identity model is invalid")
    if any(ord(character) < 32 or ord(character) == 127 for character in value):
        raise IdentityModelError("identity model is invalid")
    return value
