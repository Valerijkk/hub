from __future__ import annotations

import traceback
from collections.abc import Iterator, Mapping

import pytest

from hub_testkit.runtime import RuntimeIdentityError, RuntimeIds, parse_runtime_ids, resource_name


def _raw_ids(**overrides: object) -> dict[str, object]:
    raw: dict[str, object] = {
        "task_id": "framework-selftest",
        "run_id": "run-20260815",
        "worker_id": "gw0",
        "nonce": "0123456789abcdef0123456789abcdef",
    }
    raw.update(overrides)
    return raw


class _HostileRuntimeIds(Mapping[str, object]):
    canary = "hostile-runtime-id-canary"

    def __getitem__(self, key: str) -> object:
        raise KeyError(self.canary)

    def __iter__(self) -> Iterator[str]:
        raise KeyError(self.canary)

    def __len__(self) -> int:
        return 4


def test_runtime_ids_are_frozen_and_accept_the_full_boundary() -> None:
    ids = parse_runtime_ids(_raw_ids())

    assert ids == RuntimeIds(
        task_id="framework-selftest",
        run_id="run-20260815",
        worker_id="gw0",
        nonce="0123456789abcdef0123456789abcdef",
    )
    with pytest.raises(AttributeError):
        ids.worker_id = "gw1"


@pytest.mark.parametrize(
    "raw",
    [
        _raw_ids(task_id="ab"),
        _raw_ids(task_id="a" * 33),
        _raw_ids(run_id="short"),
        _raw_ids(run_id="run_20260815"),
        _raw_ids(worker_id="GW0"),
        _raw_ids(nonce="a" * 31),
        _raw_ids(nonce="A" * 32),
        _raw_ids(worker_id=True),
        _raw_ids(extra="forbidden"),
    ],
)
def test_runtime_ids_reject_malformed_or_open_input(raw: dict[str, object]) -> None:
    with pytest.raises(RuntimeIdentityError):
        parse_runtime_ids(raw)


def test_runtime_identity_error_does_not_echo_a_rejected_value() -> None:
    canary = "secret-runtime-canary"

    with pytest.raises(RuntimeIdentityError) as error:
        parse_runtime_ids(_raw_ids(task_id=canary))

    assert canary not in str(error.value)


def test_hostile_runtime_mapping_does_not_leak_its_key_error_context() -> None:
    with pytest.raises(RuntimeIdentityError) as error:
        parse_runtime_ids(_HostileRuntimeIds())

    formatted = "".join(traceback.format_exception(error.value))
    assert str(error.value) == "runtime IDs must use exactly four supported fields"
    assert error.value.__cause__ is None
    assert error.value.__suppress_context__ is True
    assert _HostileRuntimeIds.canary not in formatted


def test_resource_name_is_bounded_deterministic_and_nonce_bound() -> None:
    ids = parse_runtime_ids(_raw_ids())
    other_ids = parse_runtime_ids(_raw_ids(nonce="fedcba9876543210fedcba9876543210"))

    full_name = resource_name("postgres-schema", ids, max_length=200)
    short_name = resource_name("postgres-schema", ids, max_length=48)

    assert ids.nonce in full_name
    assert short_name == resource_name("postgres-schema", ids, max_length=48)
    assert len(short_name) <= 48
    assert short_name != resource_name("postgres-schema", other_ids, max_length=48)


@pytest.mark.parametrize(
    ("kind", "max_length"),
    (("DROP_TABLE", 64), ("postgres-", 64), ("db", 31), ("db", True)),
)
def test_resource_name_rejects_unsafe_kind_or_unusable_bound(
    kind: object,
    max_length: object,
) -> None:
    ids = parse_runtime_ids(_raw_ids())

    with pytest.raises(RuntimeIdentityError):
        resource_name(kind, ids, max_length=max_length)
