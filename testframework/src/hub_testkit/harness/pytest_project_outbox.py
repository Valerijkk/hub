from __future__ import annotations

from typing import NoReturn, TypeGuard

import pytest

from .project_outbox import (
    ProjectOutboxHarnessDelegate,
    ProjectOutboxHarnessFactory,
    ProjectOutboxRuntime,
)


PROJECT_OUTBOX_HARNESS_FACTORY_KEY: pytest.StashKey[ProjectOutboxHarnessFactory] = (
    pytest.StashKey()
)
_MISSING = object()
_INVALID = object()


def provide_project_outbox_harness_factory(
    config: pytest.Config,
    factory: ProjectOutboxHarnessFactory,
) -> None:
    result = _provide_factory_safe(config, factory)
    del config, factory
    if result is True:
        return
    del result
    _raise_invalid()


def require_project_outbox_harness_factory(
    config: pytest.Config,
) -> ProjectOutboxHarnessFactory:
    result = _factory_from_config_safe(config)
    del config
    if _factory_is_valid_safe(result):
        return result
    del result
    _raise_unavailable()


def create_project_outbox_runtime_from_config(
    config: pytest.Config,
) -> ProjectOutboxRuntime:
    result = _runtime_from_config_safe(config)
    del config
    if type(result) is ProjectOutboxRuntime:
        return result
    del result
    _raise_unavailable()


def cleanup_project_outbox_runtime(runtime: ProjectOutboxRuntime) -> None:
    result = _cleanup_runtime_safe(runtime)
    del runtime
    if result is True:
        return
    del result
    _raise_cleanup_failed()


def _provide_factory_safe(config: object, factory: object) -> bool:
    try:
        if not _factory_is_valid_safe(factory):
            return False
        stash = getattr(config, "stash", None)
        if (
            stash is None
            or stash.get(PROJECT_OUTBOX_HARNESS_FACTORY_KEY, _MISSING) is not _MISSING
        ):
            return False
        stash[PROJECT_OUTBOX_HARNESS_FACTORY_KEY] = factory
        return True
    except Exception:
        return False


def _factory_from_config_safe(config: object) -> ProjectOutboxHarnessFactory | object:
    try:
        stash = getattr(config, "stash", None)
        if stash is None:
            return _INVALID
        factory = stash.get(PROJECT_OUTBOX_HARNESS_FACTORY_KEY, _MISSING)
        if not _factory_is_valid_safe(factory):
            return _INVALID
        return factory
    except Exception:
        return _INVALID


def _runtime_from_config_safe(config: object) -> ProjectOutboxRuntime | object:
    try:
        factory = _factory_from_config_safe(config)
        if not _factory_is_valid_safe(factory):
            return _INVALID
        delegate = factory()
        if not _delegate_is_valid_safe(delegate):
            return _INVALID
        runtime = ProjectOutboxRuntime._create(delegate)
        del factory, delegate
        return runtime
    except Exception:
        return _INVALID


def _cleanup_runtime_safe(runtime: object) -> bool:
    try:
        if type(runtime) is not ProjectOutboxRuntime:
            return False
        runtime._cleanup()
        return True
    except Exception:
        return False


def _factory_is_valid_safe(value: object) -> TypeGuard[ProjectOutboxHarnessFactory]:
    try:
        return isinstance(value, ProjectOutboxHarnessFactory) and callable(value)
    except Exception:
        return False


def _delegate_is_valid_safe(value: object) -> TypeGuard[ProjectOutboxHarnessDelegate]:
    try:
        return isinstance(value, ProjectOutboxHarnessDelegate)
    except Exception:
        return False


def _raise_invalid() -> NoReturn:
    raise pytest.UsageError("project outbox harness is invalid") from None


def _raise_unavailable() -> NoReturn:
    raise pytest.UsageError("project outbox harness is unavailable") from None


def _raise_cleanup_failed() -> NoReturn:
    raise pytest.UsageError("project outbox harness cleanup failed") from None
