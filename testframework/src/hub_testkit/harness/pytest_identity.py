from __future__ import annotations

import pytest

from .identity import IdentityHarnessFactory


IDENTITY_HARNESS_FACTORY_KEY: pytest.StashKey[IdentityHarnessFactory] = pytest.StashKey()
_MISSING = object()


def provide_identity_harness_factory(
    config: pytest.Config,
    factory: IdentityHarnessFactory,
) -> None:
    try:
        if not isinstance(factory, IdentityHarnessFactory) or not callable(factory):
            raise ValueError("identity harness factory is invalid")
        if config.stash.get(IDENTITY_HARNESS_FACTORY_KEY, _MISSING) is not _MISSING:
            raise ValueError("identity harness is duplicated")
        config.stash[IDENTITY_HARNESS_FACTORY_KEY] = factory
    except Exception:
        raise pytest.UsageError("identity harness is invalid") from None


def require_identity_harness_factory(config: pytest.Config) -> IdentityHarnessFactory:
    try:
        factory = config.stash.get(IDENTITY_HARNESS_FACTORY_KEY, _MISSING)
        if (
            factory is _MISSING
            or not isinstance(factory, IdentityHarnessFactory)
            or not callable(factory)
        ):
            raise ValueError("identity harness factory is unavailable")
        return factory
    except Exception:
        raise pytest.UsageError("identity harness is unavailable") from None
