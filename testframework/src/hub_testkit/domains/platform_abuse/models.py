from __future__ import annotations

import hashlib
import re
from dataclasses import dataclass, field
from typing import Literal, NoReturn

from hub_testkit.http import HttpExchange


class PlatformAbuseModelError(ValueError):
    pass


_LABEL_RE = re.compile(r"^[a-z][a-z0-9-]{2,31}$")
_NONCE_RE = re.compile(r"^[0-9a-f]{32}$")
_HASH_RE = re.compile(r"^[0-9a-f]{16}$")
_CODE_RE = re.compile(r"^[A-Za-z0-9]{6}$")
_FLOWS = frozenset({"unknown-code-rate-limit", "one-time-code-replay"})


@dataclass(frozen=True, slots=True)
class PlatformAbuseSeed:
    label: str
    flow: Literal["unknown-code-rate-limit", "one-time-code-replay"]

    def __post_init__(self) -> None:
        valid = _seed_is_valid_safe(self.label, self.flow)
        if valid is True:
            return
        del valid
        _raise_invalid()


@dataclass(frozen=True, slots=True)
class OwnedVerifyScenario:
    flow: Literal["unknown-code-rate-limit", "one-time-code-replay"]
    email: str = field(repr=False)
    code: str = field(repr=False)
    nonce: str = field(repr=False)
    server_observed_client: str = field(repr=False)
    server_observed_client_hash: str = field(repr=False)
    identity_active: bool
    delivery_attempted: bool
    development_login_used: bool
    dev_mode: bool
    redis_healthy: bool
    database_healthy: bool
    redis_exclusive: bool
    database_exclusive: bool
    owned_state_entries: tuple[str, str, str, str] = field(repr=False)

    def __post_init__(self) -> None:
        valid = _scenario_is_valid_safe(
            self.flow,
            self.email,
            self.code,
            self.nonce,
            self.server_observed_client,
            self.server_observed_client_hash,
            self.identity_active,
            self.delivery_attempted,
            self.development_login_used,
            self.dev_mode,
            self.redis_healthy,
            self.database_healthy,
            self.redis_exclusive,
            self.database_exclusive,
            self.owned_state_entries,
        )
        if valid is True:
            return
        del valid
        _raise_invalid()


@dataclass(frozen=True, slots=True)
class PlatformAbuseExchange:
    exchange: HttpExchange = field(repr=False)
    retry_after: str | None = field(repr=False)

    def __post_init__(self) -> None:
        valid = _exchange_is_valid_safe(self.exchange, self.retry_after)
        if valid is True:
            return
        del valid
        _raise_invalid()


@dataclass(frozen=True, slots=True)
class TokenPair:
    access_token: str = field(repr=False)
    refresh_token: str = field(repr=False)
    expires_in: int

    def __post_init__(self) -> None:
        valid = _token_pair_is_valid_safe(
            self.access_token,
            self.refresh_token,
            self.expires_in,
        )
        if valid is True:
            return
        del valid
        _raise_invalid()


def scenario_has_canonical_client_hash(scenario: OwnedVerifyScenario) -> bool:
    result = _scenario_hash_is_canonical_safe(scenario)
    del scenario
    if result is True:
        return True
    del result
    _raise_invalid()


def _seed_is_valid_safe(label: object, flow: object) -> bool:
    try:
        return (
            type(label) is str
            and _LABEL_RE.fullmatch(label) is not None
            and _flow_is_valid(flow)
        )
    except Exception:
        return False


def _scenario_is_valid_safe(
    flow: object,
    email: object,
    code: object,
    nonce: object,
    server_observed_client: object,
    server_observed_client_hash: object,
    identity_active: object,
    delivery_attempted: object,
    development_login_used: object,
    dev_mode: object,
    redis_healthy: object,
    database_healthy: object,
    redis_exclusive: object,
    database_exclusive: object,
    owned_state_entries: object,
) -> bool:
    try:
        if not _flow_is_valid(flow) or not _email_is_valid_safe(email):
            return False
        if type(code) is not str or _CODE_RE.fullmatch(code) is None:
            return False
        if type(nonce) is not str or _NONCE_RE.fullmatch(nonce) is None:
            return False
        if not _text_is_valid_safe(server_observed_client, 256):
            return False
        if type(server_observed_client_hash) is not str or _HASH_RE.fullmatch(
            server_observed_client_hash
        ) is None:
            return False
        expected_hash = hashlib.sha256(server_observed_client.encode("utf-8")).hexdigest()[:16]
        if expected_hash != server_observed_client_hash:
            return False
        if (
            type(identity_active) is not bool
            or type(delivery_attempted) is not bool
            or delivery_attempted is not False
            or type(development_login_used) is not bool
            or development_login_used is not False
        ):
            return False
        if flow == "one-time-code-replay" and identity_active is not True:
            return False
        if (
            type(dev_mode) is not bool
            or dev_mode is not False
            or type(redis_healthy) is not bool
            or redis_healthy is not True
            or type(database_healthy) is not bool
            or database_healthy is not True
            or type(redis_exclusive) is not bool
            or redis_exclusive is not True
            or type(database_exclusive) is not bool
            or database_exclusive is not True
        ):
            return False
        return _owned_entries_are_exact_safe(
            owned_state_entries,
            email,
            server_observed_client_hash,
        )
    except Exception:
        return False


def _exchange_is_valid_safe(exchange: object, retry_after: object) -> bool:
    try:
        if type(exchange) is not HttpExchange:
            return False
        if retry_after is None:
            return True
        return type(retry_after) is str and retry_after == "600"
    except Exception:
        return False


def _token_pair_is_valid_safe(
    access_token: object,
    refresh_token: object,
    expires_in: object,
) -> bool:
    try:
        return (
            _text_is_valid_safe(access_token, 4096)
            and _text_is_valid_safe(refresh_token, 4096)
            and type(expires_in) is int
            and expires_in > 0
        )
    except Exception:
        return False


def _scenario_hash_is_canonical_safe(scenario: object) -> bool:
    try:
        if type(scenario) is not OwnedVerifyScenario:
            return False
        expected = hashlib.sha256(
            scenario.server_observed_client.encode("utf-8")
        ).hexdigest()[:16]
        return scenario.server_observed_client_hash == expected
    except Exception:
        return False


def _flow_is_valid(value: object) -> bool:
    return type(value) is str and value in _FLOWS


def _email_is_valid_safe(value: object) -> bool:
    try:
        return (
            _text_is_valid_safe(value, 320)
            and type(value) is str
            and value == value.lower()
            and value.count("@") == 1
            and not value.startswith("@")
            and not value.endswith("@")
        )
    except Exception:
        return False


def _text_is_valid_safe(value: object, maximum: int) -> bool:
    try:
        return (
            type(value) is str
            and bool(value.strip())
            and len(value) <= maximum
            and all(32 <= ord(character) != 127 for character in value)
        )
    except Exception:
        return False


def _owned_entries_are_exact_safe(
    entries: object,
    email: object,
    client_hash: object,
) -> bool:
    try:
        if type(entries) is not tuple or len(entries) != 4:
            return False
        if type(email) is not str or type(client_hash) is not str:
            return False
        expected = (
            f"rl:auth:ip:{client_hash}",
            f"rl:auth-verify-email:email:{email}",
            f"magiclink:fails:{email}",
            f"magiclink:{email}",
        )
        return entries == expected and all(type(entry) is str for entry in entries)
    except Exception:
        return False


def _raise_invalid() -> NoReturn:
    raise PlatformAbuseModelError("platform abuse model is invalid") from None
