from __future__ import annotations

from dataclasses import dataclass, field
from typing import NoReturn, Protocol, TypeGuard, runtime_checkable

from hub_testkit.domains.platform_abuse.models import (
    OwnedVerifyScenario,
    PlatformAbuseExchange,
    PlatformAbuseSeed,
)
from hub_testkit.http import HttpRequest


class PlatformAbuseHarnessError(ValueError):
    pass


@runtime_checkable
class PlatformAbuseHarnessDelegate(Protocol):
    """Launcher-owned delegate for one fresh production-like abuse scenario."""

    def send(self, request: HttpRequest) -> PlatformAbuseExchange:
        """Send one admitted public verification request through the runtime boundary."""
        ...

    def provision_scenario(self, seed: PlatformAbuseSeed) -> OwnedVerifyScenario:
        """Bind DevMode=false, exclusive healthy dependencies, and four exact owned entries.

        The launcher supplies one stable server-observed client identity before any request.
        """
        ...

    def cleanup(self) -> None:
        """Remove only the exact four state entries recorded for this nonce-bound scenario."""
        ...


@dataclass(frozen=True, slots=True, repr=False, eq=False)
class PlatformAbuseHarness:
    """Opaque per-test facade that never exposes its launcher-owned delegate."""

    _delegate: PlatformAbuseHarnessDelegate = field(repr=False)

    def __post_init__(self) -> None:
        valid = _delegate_is_valid_safe(self._delegate)
        if valid is True:
            return
        del valid
        _raise_invalid()

    def send(self, request: HttpRequest) -> PlatformAbuseExchange:
        result = _delegate_send_safe(self._delegate, request)
        del request
        if type(result) is PlatformAbuseExchange:
            return result
        del result
        _raise_invalid()

    def provision_scenario(self, seed: PlatformAbuseSeed) -> OwnedVerifyScenario:
        result = _delegate_provision_safe(self._delegate, seed)
        del seed
        if type(result) is OwnedVerifyScenario:
            return result
        del result
        _raise_invalid()

    def cleanup(self) -> None:
        result = _delegate_cleanup_safe(self._delegate)
        if result is True:
            return
        del result
        _raise_invalid()


@runtime_checkable
class PlatformAbuseHarnessFactory(Protocol):
    """Future launcher supplies one fresh nonce-bound delegate for every test."""

    def __call__(self) -> PlatformAbuseHarnessDelegate:
        """Create one delegate with an empty exact four-entry ownership ledger."""
        ...


_INVALID = object()


def _delegate_is_valid_safe(delegate: object) -> TypeGuard[PlatformAbuseHarnessDelegate]:
    try:
        return isinstance(delegate, PlatformAbuseHarnessDelegate)
    except Exception:
        return False


def _delegate_send_safe(delegate: object, request: object) -> PlatformAbuseExchange | object:
    try:
        if type(request) is not HttpRequest or not _delegate_is_valid_safe(delegate):
            return _INVALID
        result = delegate.send(request)
        if type(result) is not PlatformAbuseExchange:
            return _INVALID
        return result
    except Exception:
        return _INVALID


def _delegate_provision_safe(
    delegate: object,
    seed: object,
) -> OwnedVerifyScenario | object:
    try:
        if type(seed) is not PlatformAbuseSeed or not _delegate_is_valid_safe(delegate):
            return _INVALID
        result = delegate.provision_scenario(seed)
        if type(result) is not OwnedVerifyScenario:
            return _INVALID
        return result
    except Exception:
        return _INVALID


def _delegate_cleanup_safe(delegate: object) -> bool:
    try:
        if not _delegate_is_valid_safe(delegate):
            return False
        result: object = delegate.cleanup()
        return result is None
    except Exception:
        return False


def _raise_invalid() -> NoReturn:
    raise PlatformAbuseHarnessError("platform abuse harness is invalid") from None
