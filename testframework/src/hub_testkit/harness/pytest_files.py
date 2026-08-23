from __future__ import annotations

from typing import NoReturn, TypeGuard

import pytest

from .files import (
    FilesHarnessDelegate,
    FilesHarnessFactory,
    FilesRuntime,
)


FILES_HARNESS_FACTORY_KEY: pytest.StashKey[FilesHarnessFactory] = pytest.StashKey()
_MISSING = object()
_INVALID = object()


def provide_files_harness_factory(
    config: pytest.Config,
    factory: FilesHarnessFactory,
) -> None:
    result = _provide_factory_safe(config, factory)
    del config, factory
    if result is True:
        return
    del result
    _raise_invalid()


def require_files_harness_factory(config: pytest.Config) -> FilesHarnessFactory:
    result = _factory_from_config_safe(config)
    del config
    if _factory_is_valid_safe(result):
        return result
    del result
    _raise_unavailable()


def create_files_runtime_from_config(config: pytest.Config) -> FilesRuntime:
    result = _runtime_from_config_safe(config)
    del config
    if type(result) is FilesRuntime:
        return result
    del result
    _raise_unavailable()


def cleanup_files_runtime(runtime: FilesRuntime) -> None:
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
        if stash is None or stash.get(FILES_HARNESS_FACTORY_KEY, _MISSING) is not _MISSING:
            return False
        stash[FILES_HARNESS_FACTORY_KEY] = factory
        return True
    except Exception:
        return False


def _factory_from_config_safe(config: object) -> FilesHarnessFactory | object:
    try:
        stash = getattr(config, "stash", None)
        if stash is None:
            return _INVALID
        factory = stash.get(FILES_HARNESS_FACTORY_KEY, _MISSING)
        if not _factory_is_valid_safe(factory):
            return _INVALID
        return factory
    except Exception:
        return _INVALID


def _runtime_from_config_safe(config: object) -> FilesRuntime | object:
    try:
        factory = _factory_from_config_safe(config)
        if not _factory_is_valid_safe(factory):
            return _INVALID
        delegate = factory()
        if not _delegate_is_valid_safe(delegate):
            return _INVALID
        runtime = FilesRuntime._create(delegate)
        del factory, delegate
        return runtime
    except Exception:
        return _INVALID


def _cleanup_runtime_safe(runtime: object) -> bool:
    try:
        if type(runtime) is not FilesRuntime:
            return False
        runtime._cleanup()
        return True
    except Exception:
        return False


def _factory_is_valid_safe(value: object) -> TypeGuard[FilesHarnessFactory]:
    try:
        return isinstance(value, FilesHarnessFactory) and callable(value)
    except Exception:
        return False


def _delegate_is_valid_safe(value: object) -> TypeGuard[FilesHarnessDelegate]:
    try:
        return isinstance(value, FilesHarnessDelegate)
    except Exception:
        return False


def _raise_invalid() -> NoReturn:
    raise pytest.UsageError("files harness is invalid") from None


def _raise_unavailable() -> NoReturn:
    raise pytest.UsageError("files harness is unavailable") from None


def _raise_cleanup_failed() -> NoReturn:
    raise pytest.UsageError("files harness cleanup failed") from None
