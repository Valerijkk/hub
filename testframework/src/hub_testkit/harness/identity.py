from __future__ import annotations

from typing import Protocol, runtime_checkable

from hub_testkit.domains.identity.models import IdentitySeed, ProvisionedIdentity
from hub_testkit.http import HttpExchange, HttpRequest


@runtime_checkable
class IdentityHarness(Protocol):
    """Future identity harness owns direct nonce-bound provisioning and exact cleanup."""

    def send(self, request: HttpRequest) -> HttpExchange:
        """Send one validated request through the accepted runtime boundary."""
        ...

    def provision(self, seed: IdentitySeed) -> ProvisionedIdentity:
        """Create exactly the requested session count.

        Deactivated credentials are created then cache-invalidated.
        """
        ...

    def cleanup(self) -> None:
        """Remove exactly the identities and sessions owned by this harness invocation."""
        ...


@runtime_checkable
class IdentityHarnessFactory(Protocol):
    """Future runtime provides one fresh harness and ownership ledger per test."""

    def __call__(self) -> IdentityHarness:
        """Create a new nonce-bound harness with an empty exact cleanup ledger."""
        ...
