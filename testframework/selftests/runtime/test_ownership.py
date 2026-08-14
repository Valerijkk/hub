from __future__ import annotations

import traceback
from dataclasses import replace

import pytest

from hub_testkit.runtime import (
    OwnedResource,
    OwnershipError,
    OwnershipLedger,
    RuntimeIds,
    resource_name,
)


def _ids() -> RuntimeIds:
    return RuntimeIds(
        task_id="framework-selftest",
        run_id="run-20260815",
        worker_id="gw0",
        nonce="0123456789abcdef0123456789abcdef",
    )


def _resource(ids: RuntimeIds, *, resource_id: str = "container-1") -> OwnedResource:
    return OwnedResource(
        kind="postgres",
        resource_id=resource_id,
        name=resource_name("postgres", ids, max_length=96),
        nonce=ids.nonce,
        labels={
            "hub.qa.task": ids.task_id,
            "hub.qa.run": ids.run_id,
            "hub.qa.worker": ids.worker_id,
            "hub.qa.nonce": ids.nonce,
        },
    )


class _HostileCleanupScope:
    canary = "hostile-cleanup-scope-canary"

    def __iter__(self):
        raise TypeError(self.canary)


def test_ledger_registers_only_immutable_exactly_owned_resources() -> None:
    ids = _ids()
    resource = _resource(ids)
    ledger = OwnershipLedger(ids)

    ledger.register(resource)

    assert ledger.snapshot() == (resource,)
    with pytest.raises(TypeError):
        resource.labels["hub.qa.task"] = "other"


@pytest.mark.parametrize(
    "resource",
    [
        OwnedResource(
            kind="postgres",
            resource_id="container-2",
            name="foreign-nonce",
            nonce="fedcba9876543210fedcba9876543210",
            labels={
                "hub.qa.task": "framework-selftest",
                "hub.qa.run": "run-20260815",
                "hub.qa.worker": "gw0",
                "hub.qa.nonce": "fedcba9876543210fedcba9876543210",
            },
        ),
        OwnedResource(
            kind="postgres",
            resource_id="container-3",
            name="foreign-run",
            nonce="0123456789abcdef0123456789abcdef",
            labels={
                "hub.qa.task": "framework-selftest",
                "hub.qa.run": "run-foreign",
                "hub.qa.worker": "gw0",
                "hub.qa.nonce": "0123456789abcdef0123456789abcdef",
            },
        ),
        OwnedResource(
            kind="postgres",
            resource_id="container-4",
            name="extra-label",
            nonce="0123456789abcdef0123456789abcdef",
            labels={
                "hub.qa.task": "framework-selftest",
                "hub.qa.run": "run-20260815",
                "hub.qa.worker": "gw0",
                "hub.qa.nonce": "0123456789abcdef0123456789abcdef",
                "hub.qa.extra": "forbidden",
            },
        ),
    ],
)
def test_ledger_rejects_foreign_or_open_ownership(resource: OwnedResource) -> None:
    with pytest.raises(OwnershipError):
        OwnershipLedger(_ids()).register(resource)


def test_ledger_rejects_duplicate_identity_and_duplicate_name() -> None:
    ids = _ids()
    resource = _resource(ids)
    ledger = OwnershipLedger(ids)
    ledger.register(resource)

    with pytest.raises(OwnershipError):
        ledger.register(replace(resource, name="other-name"))
    with pytest.raises(OwnershipError):
        ledger.register(replace(resource, resource_id="container-2"))


def test_cleanup_requires_the_full_registered_identity_before_removal() -> None:
    ids = _ids()
    resource = _resource(ids)
    ledger = OwnershipLedger(ids)
    ledger.register(resource)

    tampered = replace(resource, labels={**resource.labels, "hub.qa.run": "run-foreign"})
    with pytest.raises(OwnershipError):
        ledger.assert_cleanup_scope((tampered,))
    with pytest.raises(OwnershipError):
        ledger.assert_cleanup_scope((resource.name,))

    ledger.mark_cleaned((resource,))
    assert ledger.snapshot() == ()


def test_hostile_cleanup_iterable_does_not_leak_its_type_error_context() -> None:
    ledger = OwnershipLedger(_ids())

    with pytest.raises(OwnershipError) as error:
        ledger.assert_cleanup_scope(_HostileCleanupScope())

    formatted = "".join(traceback.format_exception(error.value))
    assert str(error.value) == "cleanup scope is invalid"
    assert error.value.__cause__ is None
    assert error.value.__suppress_context__ is True
    assert _HostileCleanupScope.canary not in formatted
