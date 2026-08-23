from __future__ import annotations

from typing import NoReturn, TypeGuard

import pytest

from .authz import AuthzHarness, AuthzHarnessDelegate, AuthzHarnessFactory


AUTHZ_HARNESS_FACTORY_KEY: pytest.StashKey[AuthzHarnessFactory] = pytest.StashKey()
_MISSING = object()
_INVALID = object()


def provide_authz_harness_factory(config: pytest.Config, factory: AuthzHarnessFactory) -> None:
    result = _provide_factory_safe(config, factory)
    del config, factory
    if result is True:
        return
    del result
    _raise_invalid()


def require_authz_harness_factory(config: pytest.Config) -> AuthzHarnessFactory:
    result = _factory_from_config_safe(config)
    del config
    if _factory_is_valid_safe(result):
        return result
    del result
    _raise_unavailable()


def create_authz_harness_from_config(config: pytest.Config) -> AuthzHarness:
    result = _harness_from_config_safe(config)
    del config
    if type(result) is AuthzHarness:
        return result
    del result
    _raise_unavailable()


def cleanup_authz_harness(harness: AuthzHarness) -> None:
    result = _cleanup_harness_safe(harness)
    del harness
    if result is True:
        return
    del result
    _raise_cleanup_failed()


def _provide_factory_safe(config: object, factory: object) -> bool:
    try:
        if not _factory_is_valid_safe(factory):
            return False
        stash = getattr(config, "stash", None)
        if stash is None or stash.get(AUTHZ_HARNESS_FACTORY_KEY, _MISSING) is not _MISSING:
            return False
        stash[AUTHZ_HARNESS_FACTORY_KEY] = factory
        return True
    except Exception:
        return False


def _factory_from_config_safe(config: object) -> AuthzHarnessFactory | object:
    try:
        stash = getattr(config, "stash", None)
        if stash is None:
            return _INVALID
        factory = stash.get(AUTHZ_HARNESS_FACTORY_KEY, _MISSING)
        if not _factory_is_valid_safe(factory):
            return _INVALID
        return factory
    except Exception:
        return _INVALID


def _harness_from_config_safe(config: object) -> AuthzHarness | object:
    try:
        factory = _factory_from_config_safe(config)
        if not _factory_is_valid_safe(factory):
            return _INVALID
        delegate = factory()
        if not _delegate_is_valid_safe(delegate):
            return _INVALID
        result = AuthzHarness(delegate)
        del factory, delegate
        return result
    except Exception:
        return _INVALID


def _cleanup_harness_safe(harness: object) -> bool:
    try:
        if type(harness) is not AuthzHarness:
            return False
        harness.cleanup()
        return True
    except Exception:
        return False


def _factory_is_valid_safe(value: object) -> TypeGuard[AuthzHarnessFactory]:
    try:
        return isinstance(value, AuthzHarnessFactory) and callable(value)
    except Exception:
        return False


def _delegate_is_valid_safe(value: object) -> TypeGuard[AuthzHarnessDelegate]:
    try:
        return isinstance(value, AuthzHarnessDelegate)
    except Exception:
        return False


def _raise_invalid() -> NoReturn:
    raise pytest.UsageError("authorization harness is invalid") from None


def _raise_unavailable() -> NoReturn:
    raise pytest.UsageError("authorization harness is unavailable") from None


def _raise_cleanup_failed() -> NoReturn:
    raise pytest.UsageError("authorization harness cleanup failed") from None
