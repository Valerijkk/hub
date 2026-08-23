from __future__ import annotations

from typing import NoReturn, Protocol, TypeGuard, runtime_checkable

import pytest

from .source_bundle import ContractSourceBundle, SourceSnapshot


@runtime_checkable
class ContractSourceBundleDelegate(Protocol):
    """Launcher-owned delegate for one claimed immutable source snapshot."""

    def materialize(self) -> SourceSnapshot:
        """Return the single claimed source snapshot for this test."""
        ...


@runtime_checkable
class ContractSourceBundleFactory(Protocol):
    """Future launcher creates one source delegate per test invocation."""

    def __call__(self) -> ContractSourceBundleDelegate:
        """Create one delegate without exposing loader or filesystem authority."""
        ...


CONTRACT_SOURCE_BUNDLE_FACTORY_KEY: pytest.StashKey[ContractSourceBundleFactory] = (
    pytest.StashKey()
)
_MISSING = object()
_INVALID = object()


def provide_contract_source_bundle_factory(
    config: pytest.Config,
    factory: ContractSourceBundleFactory,
) -> None:
    result = _provide_factory_safe(config, factory)
    del config, factory
    if result is True:
        return
    del result
    _raise_invalid()


def require_contract_source_bundle_factory(
    config: pytest.Config,
) -> ContractSourceBundleFactory:
    factory = _factory_from_config_safe(config)
    del config
    if _factory_valid(factory):
        return factory
    del factory
    _raise_unavailable()


def create_contract_source_bundle_from_config(config: pytest.Config) -> ContractSourceBundle:
    bundle = _bundle_from_config_safe(config)
    del config
    if type(bundle) is ContractSourceBundle:
        return bundle
    del bundle
    _raise_unavailable()


@pytest.fixture
def contract_source_bundle(request: pytest.FixtureRequest) -> ContractSourceBundle:
    config = request.config
    result = create_contract_source_bundle_from_config(config)
    del request, config
    return result


def _provide_factory_safe(config: object, factory: object) -> bool:
    try:
        if not _factory_valid(factory):
            return False
        stash = getattr(config, "stash", None)
        if stash is None or stash.get(CONTRACT_SOURCE_BUNDLE_FACTORY_KEY, _MISSING) is not _MISSING:
            return False
        stash[CONTRACT_SOURCE_BUNDLE_FACTORY_KEY] = factory
        return True
    except Exception:
        return False


def _factory_from_config_safe(config: object) -> ContractSourceBundleFactory | object:
    try:
        stash = getattr(config, "stash", None)
        if stash is None:
            return _INVALID
        factory = stash.get(CONTRACT_SOURCE_BUNDLE_FACTORY_KEY, _MISSING)
        if not _factory_valid(factory):
            return _INVALID
        return factory
    except Exception:
        return _INVALID


def _bundle_from_config_safe(config: object) -> ContractSourceBundle | object:
    try:
        factory = _factory_from_config_safe(config)
        if not _factory_valid(factory):
            return _INVALID
        delegate = factory()
        if not _delegate_valid(delegate):
            return _INVALID
        snapshot = delegate.materialize()
        if type(snapshot) is not SourceSnapshot:
            return _INVALID
        bundle = ContractSourceBundle(snapshot)
        del factory, delegate, snapshot
        return bundle
    except Exception:
        return _INVALID


def _factory_valid(value: object) -> TypeGuard[ContractSourceBundleFactory]:
    try:
        return isinstance(value, ContractSourceBundleFactory) and callable(value)
    except Exception:
        return False


def _delegate_valid(value: object) -> TypeGuard[ContractSourceBundleDelegate]:
    try:
        return isinstance(value, ContractSourceBundleDelegate)
    except Exception:
        return False


def _raise_invalid() -> NoReturn:
    raise pytest.UsageError("contract source bundle factory is invalid") from None


def _raise_unavailable() -> NoReturn:
    raise pytest.UsageError("contract source bundle factory is unavailable") from None
