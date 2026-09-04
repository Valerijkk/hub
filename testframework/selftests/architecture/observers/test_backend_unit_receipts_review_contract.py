from __future__ import annotations

import hashlib
import json
from collections.abc import Callable, Mapping
from pathlib import Path

import pytest
from jsonschema import Draft202012Validator

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


def test_valid_final_receipt_preserves_sealed_timeout_outcome_pair() -> None:
    document = _fixture_json("valid/final-receipt.json")

    Draft202012Validator(_final_receipt_schema()).validate(document)
    assert document["classification"] == "SEALED_TIMEOUT_UNTRIAGED"
    assert document["immutable_outcome"] == "TIMEOUT_UNTRIAGED"
    validate_final_receipt(document)
