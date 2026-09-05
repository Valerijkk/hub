"""Fail-closed offline integrity checks for one backend-unit container packet."""

from __future__ import annotations

import hashlib
import json
import os
import re
import sys
from collections.abc import Mapping, Sequence
from pathlib import Path
from typing import Final

from jsonschema import Draft202012Validator, ValidationError

from .contracts import ReceiptContractError, validate_container_receipt
from .safe_paths import SafePathError, list_directory_relative, read_regular_relative


_READ_BUDGET_BYTES: Final = 64 * 1024 * 1024
_RUN_DIR_RE: Final = re.compile(r"backend-unit-([0-9a-f]{12,64})\.partial\Z")
_RECEIPT_NAME: Final = "container-receipt.json"
_TERMINAL_MARKERS: Final = {
    "CONTAINER_COMPLETE": "CONTAINER_COMPLETE",
    "CONTAINER_ABORTED": "CONTAINER_ABORTED",
}
_COMPLETE_STAGES: Final = ("build", "vet", "unit", "race")
_COMPLETE_INVENTORY_NAMES: Final = frozenset(
    {
        "CONTAINER_COMPLETE",
        "owner.json",
        *(
            name
            for stage in _COMPLETE_STAGES
            for name in (
                f"package-plan.{stage}.json",
                f"package-results.{stage}.jsonl",
                f"stage-{stage}.json",
                f"{stage}.stdout.log",
                f"{stage}.stderr.log",
            )
        ),
    }
)
_SCHEMA_PATH: Final = (
    Path(__file__).resolve().parents[3]
    / "testframework/schemas/backend-unit-container-receipt.schema.v1.json"
)


def _reject_constant(_: str) -> object:
    raise ValueError("non-finite JSON constants are denied")


def _unique_object(pairs: list[tuple[str, object]]) -> dict[str, object]:
    result: dict[str, object] = {}
    for key, value in pairs:
        if key in result:
            raise ValueError("duplicate JSON key")
        result[key] = value
    return result


def _parse_strict_json(payload: bytes) -> dict[str, object]:
    try:
        document = json.loads(
            payload.decode("utf-8", "strict"),
            object_pairs_hook=_unique_object,
            parse_constant=_reject_constant,
        )
    except (UnicodeDecodeError, ValueError, json.JSONDecodeError) as error:
        raise ReceiptContractError("container packet JSON is unusable") from error
    if type(document) is not dict:
        raise ReceiptContractError("container packet JSON root must be an object")
    return document


def _schema_validator() -> Draft202012Validator:
    try:
        schema = json.loads(_SCHEMA_PATH.read_text(encoding="utf-8"))
        if type(schema) is not dict:
            raise ValueError("schema root must be an object")
        return Draft202012Validator(schema)
    except Exception as error:
        raise ReceiptContractError("container packet schema is unavailable") from error


def _packet_inventory(document: Mapping[str, object]) -> dict[str, tuple[str, int]]:
    inventory = document.get("artifact_inventory")
    if type(inventory) is not list:
        raise ReceiptContractError("container packet inventory is unusable")

    records: dict[str, tuple[str, int]] = {}
    for record in inventory:
        if type(record) is not dict:
            raise ReceiptContractError("container packet inventory is unusable")
        path = record.get("path")
        digest = record.get("sha256")
        size = record.get("size")
        if (
            type(path) is not str
            or not path
            or "/" in path
            or "\\" in path
            or type(digest) is not str
            or type(size) is not int
            or size < 0
            or path in records
        ):
            raise ReceiptContractError("container packet inventory is not flat")
        records[path] = (digest, size)
    return records


def _read_with_budget(
    reports_root: Path, relative: str, consumed: int
) -> tuple[bytes, int]:
    remaining = _READ_BUDGET_BYTES - consumed
    if remaining < 0:
        raise ReceiptContractError("container packet read budget is exhausted")
    try:
        payload = read_regular_relative(reports_root, relative, max_bytes=remaining)
    except SafePathError as error:
        raise ReceiptContractError("container packet path is unsafe or unreadable") from error
    return payload, consumed + len(payload)


def _list_packet(reports_root: Path, run_dir: str) -> tuple[str, ...]:
    try:
        return list_directory_relative(reports_root, run_dir)
    except SafePathError as error:
        raise ReceiptContractError("container packet directory is unsafe or unreadable") from error


def validate_container_packet(reports_root: Path, run_dir: str) -> None:
    """Validate an offline artifact packet; this never asserts a test outcome."""

    if not isinstance(reports_root, Path) or not reports_root.is_absolute():
        raise ReceiptContractError("reports root must be an absolute Path")
    if type(run_dir) is not str:
        raise ReceiptContractError("container packet run directory is invalid")
    matched = _RUN_DIR_RE.fullmatch(run_dir)
    if matched is None:
        raise ReceiptContractError("container packet run directory is invalid")
    expected_hostname = matched.group(1)

    receipt_relative = f"{run_dir}/{_RECEIPT_NAME}"
    receipt_payload, consumed = _read_with_budget(reports_root, receipt_relative, 0)
    document = _parse_strict_json(receipt_payload)
    try:
        _schema_validator().validate(document)
        validate_container_receipt(document)
    except (ReceiptContractError, ValidationError) as error:
        raise ReceiptContractError("container packet receipt is invalid") from error

    hostname = document.get("container_hostname")
    if hostname != expected_hostname:
        raise ReceiptContractError("container packet hostname does not bind run directory")
    state = document.get("state")
    marker = _TERMINAL_MARKERS.get(state)
    if marker is None:
        raise ReceiptContractError("container packet terminal state is invalid")

    inventory = _packet_inventory(document)
    if state == "CONTAINER_COMPLETE" and set(inventory) != _COMPLETE_INVENTORY_NAMES:
        raise ReceiptContractError("complete container packet inventory is not canonical")
    expected_names = set(inventory)
    expected_names.add(_RECEIPT_NAME)
    initial_names = _list_packet(reports_root, run_dir)
    if tuple(sorted(expected_names)) != initial_names:
        raise ReceiptContractError("container packet tree does not exactly match inventory")
    if marker not in inventory:
        raise ReceiptContractError("container packet terminal marker is not inventoried")

    for path, (digest, size) in inventory.items():
        payload, consumed = _read_with_budget(reports_root, f"{run_dir}/{path}", consumed)
        if len(payload) != size or "sha256:" + hashlib.sha256(payload).hexdigest() != digest:
            raise ReceiptContractError("container packet artifact does not match inventory")
        if path == marker and payload != b"":
            raise ReceiptContractError("container packet terminal marker must be empty")

    try:
        final_receipt = read_regular_relative(
            reports_root,
            receipt_relative,
            max_bytes=len(receipt_payload),
        )
    except SafePathError as error:
        raise ReceiptContractError("container packet receipt changed while being validated") from error
    final_names = _list_packet(reports_root, run_dir)
    if final_receipt != receipt_payload or final_names != initial_names:
        raise ReceiptContractError("container packet changed while being validated")


def main(argv: Sequence[str] | None = None) -> int:
    """Check only the fixed reports packet selected by the fixed environment key."""

    arguments = sys.argv[1:] if argv is None else argv
    if tuple(arguments) != ():
        print("DOCKER_ARTIFACT_PACKET_INVALID", file=sys.stderr)
        return 1
    run_dir = os.environ.get("HUB_BACKEND_UNIT_RUN_DIR")
    try:
        validate_container_packet(Path("/reports"), run_dir)
    except ReceiptContractError:
        print("DOCKER_ARTIFACT_PACKET_INVALID", file=sys.stderr)
        return 1
    print("DOCKER_ARTIFACT_PACKET_PASS artifact_integrity_only_not_test_verdict")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
