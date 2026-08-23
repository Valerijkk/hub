from __future__ import annotations

from dataclasses import dataclass, field
from typing import NoReturn, Protocol, TypeGuard, runtime_checkable

from hub_testkit.domains.authz.models import PrincipalSeed, ProvisionedPrincipal
from hub_testkit.http import HttpExchange, HttpRequest


class AuthzHarnessError(ValueError):
    pass


@runtime_checkable
class AuthzHarnessDelegate(Protocol):
    """Launcher-owned delegate behind the opaque authorization test facade."""

    def send(self, request: HttpRequest) -> HttpExchange:
        """Send one already validated request through the accepted runtime boundary."""
        ...

    def provision_principal(self, seed: PrincipalSeed) -> ProvisionedPrincipal:
        """Provision one principal and only its exact credentials into this ledger."""
        ...

    def cleanup(self) -> None:
        """Remove exactly the nonce-bound resources recorded by this delegate."""
        ...


@dataclass(frozen=True, slots=True, repr=False, eq=False)
class AuthzHarness:
    """Opaque per-test facade that never exposes its launcher-owned delegate."""

    _delegate: AuthzHarnessDelegate = field(repr=False)

    def __post_init__(self) -> None:
        valid = _delegate_is_valid_safe(self._delegate)
        if valid is True:
            return
        del valid
        _raise_invalid()

    def send(self, request: HttpRequest) -> HttpExchange:
        result = _delegate_send_safe(self._delegate, request)
        del request
        if type(result) is HttpExchange:
            return result
        del result
        _raise_invalid()

    def provision_principal(self, seed: PrincipalSeed) -> ProvisionedPrincipal:
        result = _delegate_provision_safe(self._delegate, seed)
        del seed
        if type(result) is ProvisionedPrincipal:
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
class AuthzHarnessFactory(Protocol):
    """Future launcher supplies a fresh authorization delegate for every test."""

    def __call__(self) -> AuthzHarnessDelegate:
        """Create one delegate with one empty nonce-bound cleanup ledger."""
        ...


_INVALID = object()


def _delegate_is_valid_safe(delegate: object) -> TypeGuard[AuthzHarnessDelegate]:
    try:
        return isinstance(delegate, AuthzHarnessDelegate)
    except Exception:
        return False


def _delegate_send_safe(delegate: object, request: object) -> HttpExchange | object:
    try:
        if type(request) is not HttpRequest or not _delegate_is_valid_safe(delegate):
            return _INVALID
        result = delegate.send(request)
        if type(result) is not HttpExchange:
            return _INVALID
        return result
    except Exception:
        return _INVALID


def _delegate_provision_safe(
    delegate: object,
    seed: object,
) -> ProvisionedPrincipal | object:
    try:
        if type(seed) is not PrincipalSeed or not _delegate_is_valid_safe(delegate):
            return _INVALID
        result = delegate.provision_principal(seed)
        if type(result) is not ProvisionedPrincipal:
            return _INVALID
        return result
    except Exception:
        return _INVALID


def _delegate_cleanup_safe(delegate: object) -> bool:
    try:
        if not _delegate_is_valid_safe(delegate):
            return False
        result: object = delegate.cleanup()
        if result is not None:
            return False
        return True
    except Exception:
        return False


def _raise_invalid() -> NoReturn:
    raise AuthzHarnessError("authorization harness is invalid") from None
