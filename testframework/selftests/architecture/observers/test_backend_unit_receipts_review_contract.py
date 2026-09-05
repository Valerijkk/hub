from __future__ import annotations

import hashlib
import json
from collections.abc import Callable, Mapping
from pathlib import Path

import pytest
from jsonschema import Draft202012Validator, ValidationError

from observers.backend_unit.contracts import (
    ReceiptContractError,
    parse_seal,
    validate_container_receipt,
    validate_final_receipt,
    validate_observer_receipt,
    validate_package_plan,
)


_REPO_ROOT = Path(__file__).resolve().parents[4]
_FIXTURE_ROOT = Path(__file__).parent / "fixtures" / "backend_unit"
_SCHEMA_PATHS = {
    "package_plan": _REPO_ROOT
    / "testframework/schemas/backend-unit-package-plan.schema.v1.json",
    "container_receipt": _REPO_ROOT
    / "testframework/schemas/backend-unit-container-receipt.schema.v1.json",
    "observer_receipt": _REPO_ROOT
    / "testframework/schemas/backend-unit-observer-receipt.schema.v1.json",
}
_INVALID_SEAL_FIXTURES = (
    "invalid/SEAL.missing-field",
    "invalid/SEAL.extra-field",
    "invalid/SEAL.reordered",
)
_SEMANTIC_INVALID_FIXTURES: tuple[
    tuple[str, str, Callable[[Mapping[str, object]], None]], ...
] = (
    (
        "package_plan",
        "invalid/discovery-argv-package-plan.json",
        validate_package_plan,
    ),
    (
        "package_plan",
        "invalid/duplicate-import-path-package-plan.json",
        validate_package_plan,
    ),
    ("package_plan", "invalid/count-package-plan.json", validate_package_plan),
    (
        "package_plan",
        "invalid/row-set-hash-package-plan.json",
        validate_package_plan,
    ),
    (
        "package_plan",
        "invalid/non-bijection-package-plan.json",
        validate_package_plan,
    ),
    (
        "container_receipt",
        "invalid/inventory-container-receipt.json",
        validate_container_receipt,
    ),
    (
        "container_receipt",
        "invalid/inventory-hash-container-receipt.json",
        validate_container_receipt,
    ),
    (
        "observer_receipt",
        "invalid/foreign-run-observer-receipt.json",
        validate_observer_receipt,
    ),
    (
        "observer_receipt",
        "invalid/engine-id-observer-receipt.json",
        validate_observer_receipt,
    ),
    (
        "observer_receipt",
        "invalid/build-input-manifest-observer-receipt.json",
        validate_observer_receipt,
    ),
    (
        "observer_receipt",
        "invalid/build-input-hash-observer-receipt.json",
        validate_observer_receipt,
    ),
    (
        "observer_receipt",
        "invalid/fixed-image-id-observer-receipt.json",
        validate_observer_receipt,
    ),
    (
        "observer_receipt",
        "invalid/bound-backend-image-observer-receipt.json",
        validate_observer_receipt,
    ),
    (
        "final_receipt",
        "invalid/timeout-outcome-final-receipt.json",
        validate_final_receipt,
    ),
)
_OBSERVER_SEMANTIC_FAILURES = (
    (
        "invalid/foreign-run-observer-receipt.json",
        "final receipt run_id does not exactly bind observer receipt",
    ),
    (
        "invalid/engine-id-observer-receipt.json",
        "final receipt container_id does not exactly bind observer receipt",
    ),
    (
        "invalid/fixed-image-id-observer-receipt.json",
        "container image and bound backend image must equal the fixed backend image ID",
    ),
    (
        "invalid/bound-backend-image-observer-receipt.json",
        "container image and bound backend image must equal the fixed backend image ID",
    ),
    (
        "invalid/build-input-manifest-observer-receipt.json",
        "canonical build-input manifest paths or order differ from frozen policy",
    ),
    (
        "invalid/build-input-hash-observer-receipt.json",
        "canonical build-input manifest hash does not bind its exact bytes",
    ),
)


def _fixture_bytes(relative_path: str) -> bytes:
    payload = (_FIXTURE_ROOT / relative_path).read_bytes()
    assert payload.endswith(b"\n")
    return payload


def _fixture_json(relative_path: str) -> Mapping[str, object]:
    document = json.loads(_fixture_bytes(relative_path))
    assert type(document) is dict
    return document


def _schema(family: str) -> Mapping[str, object]:
    document = json.loads(_SCHEMA_PATHS[family].read_bytes())
    assert type(document) is dict
    return document


def _final_receipt_schema() -> Mapping[str, object]:
    observer_schema = _schema("observer_receipt")
    definitions = observer_schema["$defs"]
    assert type(definitions) is dict
    final_receipt = definitions["finalReceipt"]
    assert type(final_receipt) is dict
    return {"$schema": observer_schema["$schema"], "$defs": definitions, **final_receipt}


def _schema_for(family: str) -> Mapping[str, object]:
    if family == "final_receipt":
        return _final_receipt_schema()
    return _schema(family)


def _bind_review_abort_budget(document: dict[str, object]) -> None:
    inventory = document["artifact_inventory"]
    abort = document["abort"]
    assert type(inventory) is list and type(abort) is dict
    sizes = {record["path"]: record["size"] for record in inventory}
    assert all(type(path) is str and type(size) is int for path, size in sizes.items())
    completed = abort["completed_stages"]
    assert type(completed) is list and all(type(stage) is str for stage in completed)
    caps = {
        "build": (2097152, 2097152), "vet": (2097152, 2097152),
        "unit": (16777216, 4194304), "race": (25165824, 4194304),
    }
    streams: dict[str, object] = {}
    for stage, (stdout_cap, stderr_cap) in caps.items():
        streams[stage] = {
            "stdout": {"cap_bytes": stdout_cap, "used_bytes": sizes[f"{stage}.stdout.log"] if stage in completed else 0, "truncated": False},
            "stderr": {"cap_bytes": stderr_cap, "used_bytes": sizes[f"{stage}.stderr.log"] if stage in completed else 0, "truncated": False},
        }
    document["report_budget"] = {
        "total_bytes": 67108864, "reserved_metadata_bytes": 8388608,
        "metadata_used_bytes": 0, "streams": streams, "total_used_bytes": 0,
        "truncated_due_to_budget": False, "raw_output_partial": False,
    }
    metadata_inventory = sum(
        size for path, size in sizes.items()
        if not (path.endswith(".stdout.log") or path.endswith(".stderr.log"))
    )
    inventory_total = sum(sizes.values())
    budget = document["report_budget"]
    assert type(budget) is dict
    for _ in range(16):
        receipt_size = len(json.dumps(
            document, allow_nan=False, ensure_ascii=False, separators=(",", ":"), sort_keys=True,
        ).encode("utf-8") + b"\n")
        expected_metadata = metadata_inventory + receipt_size
        expected_total = inventory_total + receipt_size
        if budget["metadata_used_bytes"] == expected_metadata and budget["total_used_bytes"] == expected_total:
            return
        budget["metadata_used_bytes"] = expected_metadata
        budget["total_used_bytes"] = expected_total
    raise AssertionError("abort report budget fixed point did not converge")


def _review_abort_execution_document() -> dict[str, object]:
    complete = json.loads(json.dumps(_fixture_json("valid/container-receipt.json")))
    prefix = (
        "schema", "target", "run_id", "owner_nonce", "container_hostname",
        "launcher_sha256", "command_profile_sha256",
        "source_manifest_before_sha256", "source_manifest_after_sha256",
        "cgroup", "go_environment",
    )
    document = {key: complete[key] for key in prefix}
    document["abort"] = {
        "signal": "TERM", "phase": "execution", "active_stage": "vet",
        "planned_stages": ["build", "vet", "unit", "race"],
        "completed_stages": ["build"],
        "child_group_reaped": True, "stdout_reader_reaped": True,
        "stderr_reader_reaped": True, "unpublished_stage_rolled_back": True,
    }
    paths = (
        "CONTAINER_ABORTED", "build.stderr.log", "build.stdout.log", "owner.json",
        "package-plan.build.json", "package-plan.race.json", "package-plan.unit.json", "package-plan.vet.json",
        "package-results.build.jsonl", "stage-build.json",
    )
    inventory = [
        {
            "path": path,
            "sha256": "sha256:" + hashlib.sha256(path.encode("utf-8")).hexdigest(),
            "size": len(path),
        }
        for path in paths
    ]
    payload = b"".join(
        (
            str(record["path"]) + "\t" + str(record["sha256"])
            + "\t" + str(record["size"]) + "\n"
        ).encode("ascii")
        for record in inventory
    )
    document["artifact_inventory"] = inventory
    document["artifact_inventory_sha256"] = "sha256:" + hashlib.sha256(payload).hexdigest()
    document["report_budget"] = complete["report_budget"]
    document["terminal"] = {"signal": "TERM", "exit_code": 143}
    document["state"] = "CONTAINER_ABORTED"
    _bind_review_abort_budget(document)
    return document


def test_review_contract_requires_exact_inline_aborted_execution_inventory() -> None:
    document = _review_abort_execution_document()

    Draft202012Validator(_schema("container_receipt")).validate(document)
    validate_container_receipt(document)
    document["artifact_inventory"].append({
        "path": "temp-output",
        "sha256": "sha256:" + "a" * 64,
        "size": 0,
    })
    document["artifact_inventory"].sort(key=lambda record: str(record["path"]))
    payload = b"".join(
        (
            str(record["path"]) + "\t" + str(record["sha256"])
            + "\t" + str(record["size"]) + "\n"
        ).encode("ascii")
        for record in document["artifact_inventory"]
    )
    document["artifact_inventory_sha256"] = "sha256:" + hashlib.sha256(payload).hexdigest()

    Draft202012Validator(_schema("container_receipt")).validate(document)
    with pytest.raises(ReceiptContractError, match="exactly derive"):
        validate_container_receipt(document)


def test_complete_valid_seal_rejects_only_the_crlf_mutation() -> None:
    valid_seal = _fixture_bytes("valid/SEAL")

    assert valid_seal.count(b"\n") == 4
    assert b"\r" not in valid_seal
    assert set(parse_seal(valid_seal)) == {
        "run_id",
        "owner_nonce",
        "observer_receipt_sha256",
    }
    assert parse_seal(valid_seal)["observer_receipt_sha256"] == (
        "sha256:"
        + hashlib.sha256(_fixture_bytes("valid/observer-receipt.json")).hexdigest()
    )

    crlf_only = valid_seal.replace(b"\n", b"\r\n")

    assert crlf_only.replace(b"\r", b"") == valid_seal
    with pytest.raises(ReceiptContractError):
        parse_seal(crlf_only)


@pytest.mark.parametrize("fixture", _INVALID_SEAL_FIXTURES)
def test_committed_invalid_seal_fixtures_are_rejected(fixture: str) -> None:
    with pytest.raises(ReceiptContractError):
        parse_seal(_fixture_bytes(fixture))


@pytest.mark.parametrize(
    ("family", "fixture", "validator"), _SEMANTIC_INVALID_FIXTURES
)
def test_semantic_invalid_fixtures_pass_shape_then_fail_matching_semantics(
    family: str,
    fixture: str,
    validator: Callable[[Mapping[str, object]], None],
) -> None:
    document = _fixture_json(fixture)

    Draft202012Validator(_schema_for(family)).validate(document)
    with pytest.raises(ReceiptContractError):
        validator(document)


@pytest.mark.parametrize(("fixture", "expected_message"), _OBSERVER_SEMANTIC_FAILURES)
def test_observer_semantic_invalid_fixtures_fail_for_their_named_rule(
    fixture: str, expected_message: str
) -> None:
    document = _fixture_json(fixture)

    Draft202012Validator(_schema("observer_receipt")).validate(document)
    with pytest.raises(ReceiptContractError) as raised:
        validate_observer_receipt(document)

    assert str(raised.value) == expected_message


@pytest.mark.parametrize("field", ("root_handoff_sha256", "concurrency_observation_scope"))
def test_observer_schema_rejects_missing_root_handoff_bindings(field: str) -> None:
    document = dict(_fixture_json("valid/observer-receipt.json"))
    document.pop(field)

    with pytest.raises(ValidationError):
        Draft202012Validator(_schema("observer_receipt")).validate(document)


def test_observer_schema_rejects_image_binding_inside_source_snapshot() -> None:
    document = json.loads(json.dumps(_fixture_json("valid/observer-receipt.json")))
    document["source_snapshots"]["observer_snapshot"]["fixed_image_ids"] = {
        "python_contract": "sha256:" + "c" * 64,
        "backend_unit": "sha256:" + "d" * 64,
    }

    with pytest.raises(ValidationError):
        Draft202012Validator(_schema("observer_receipt")).validate(document)


def test_valid_final_receipt_preserves_sealed_timeout_outcome_pair() -> None:
    document = _fixture_json("valid/final-receipt.json")

    Draft202012Validator(_final_receipt_schema()).validate(document)
    assert document["classification"] == "SEALED_TIMEOUT_UNTRIAGED"
    assert document["immutable_outcome"] == "TIMEOUT_UNTRIAGED"
    validate_final_receipt(document)
