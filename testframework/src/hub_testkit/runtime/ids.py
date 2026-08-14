from __future__ import annotations

import re
from collections.abc import Mapping
from dataclasses import dataclass
from hashlib import sha256


class RuntimeIdentityError(ValueError):
    pass


_TASK_PATTERN = re.compile(r"^[a-z][a-z0-9-]{2,31}$")
_RUN_PATTERN = re.compile(r"^[a-z0-9][a-z0-9-]{7,63}$")
_WORKER_PATTERN = re.compile(r"^[a-z0-9][a-z0-9-]{0,31}$")
_NONCE_PATTERN = re.compile(r"^[0-9a-f]{32}$")
_KIND_PATTERN = re.compile(r"^[a-z](?:[a-z0-9-]{0,30}[a-z0-9])?$")
_RUNTIME_ID_KEYS = frozenset({"task_id", "run_id", "worker_id", "nonce"})


@dataclass(frozen=True, slots=True)
class RuntimeIds:
    task_id: str
    run_id: str
    worker_id: str
    nonce: str

    def __post_init__(self) -> None:
        _validate_identifier("task ID", self.task_id, _TASK_PATTERN)
        _validate_identifier("run ID", self.run_id, _RUN_PATTERN)
        _validate_identifier("worker ID", self.worker_id, _WORKER_PATTERN)
        _validate_identifier("nonce", self.nonce, _NONCE_PATTERN)


def parse_runtime_ids(raw: Mapping[str, object]) -> RuntimeIds:
    if not isinstance(raw, Mapping):
        raise RuntimeIdentityError("runtime IDs must use exactly four supported fields")
    try:
        if set(raw) != _RUNTIME_ID_KEYS:
            raise RuntimeIdentityError("runtime IDs must use exactly four supported fields")
        return RuntimeIds(
            task_id=_require_string(raw["task_id"]),
            run_id=_require_string(raw["run_id"]),
            worker_id=_require_string(raw["worker_id"]),
            nonce=_require_string(raw["nonce"]),
        )
    except KeyError:
        raise RuntimeIdentityError("runtime IDs must use exactly four supported fields") from None


def resource_name(kind: str, ids: RuntimeIds, max_length: int) -> str:
    _validate_identifier("resource kind", kind, _KIND_PATTERN)
    if not isinstance(ids, RuntimeIds):
        raise RuntimeIdentityError("runtime IDs are invalid")
    if type(max_length) is not int or max_length < 32:
        raise RuntimeIdentityError("resource name bound is invalid")

    full_name = "-".join(("hubqa", kind, ids.task_id, ids.run_id, ids.worker_id, ids.nonce))
    if len(full_name) <= max_length:
        return full_name

    digest = sha256(full_name.encode("ascii")).hexdigest()[:16]
    prefix_length = max_length - len(digest) - 1
    prefix = full_name[:prefix_length].rstrip("-")
    return f"{prefix}-{digest}"


def _require_string(value: object) -> str:
    if type(value) is not str:
        raise RuntimeIdentityError("runtime ID values must be strings")
    return value


def _validate_identifier(label: str, value: object, pattern: re.Pattern[str]) -> None:
    if type(value) is not str or pattern.fullmatch(value) is None:
        raise RuntimeIdentityError(f"{label} is invalid")
