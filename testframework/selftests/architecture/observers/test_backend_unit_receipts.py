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
_VALID_SCHEMA_FIXTURES = (
    ("package_plan", "valid/package-plan.json"),
    ("container_receipt", "valid/container-receipt.json"),
    ("observer_receipt", "valid/observer-receipt.json"),
)
_STRUCTURAL_INVALID_FIXTURES = (
    ("package_plan", "invalid/unknown-key-package-plan.json"),
    ("package_plan", "invalid/uppercase-digest-package-plan.json"),
    ("observer_receipt", "invalid/nonce-observer-receipt.json"),
)
_SEMANTIC_INVALID_FIXTURES: tuple[
    tuple[str, Callable[[Mapping[str, object]], None]], ...
] = (
    ("invalid/discovery-argv-package-plan.json", validate_package_plan),
    ("invalid/duplicate-import-path-package-plan.json", validate_package_plan),
    ("invalid/count-package-plan.json", validate_package_plan),
    ("invalid/row-set-hash-package-plan.json", validate_package_plan),
    ("invalid/non-bijection-package-plan.json", validate_package_plan),
    ("invalid/inventory-container-receipt.json", validate_container_receipt),
    ("invalid/inventory-hash-container-receipt.json", validate_container_receipt),
    ("invalid/foreign-run-observer-receipt.json", validate_observer_receipt),
    ("invalid/engine-id-observer-receipt.json", validate_observer_receipt),
    (
        "invalid/build-input-manifest-observer-receipt.json",
        validate_observer_receipt,
    ),
    ("invalid/build-input-hash-observer-receipt.json", validate_observer_receipt),
    ("invalid/fixed-image-id-observer-receipt.json", validate_observer_receipt),
    (
        "invalid/bound-backend-image-observer-receipt.json",
        validate_observer_receipt,
    ),
    ("invalid/timeout-outcome-final-receipt.json", validate_final_receipt),
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
_INVALID_LF_SEAL_BYTES = (
    (
        "missing required observer receipt hash",
        b"hub.qa.backend-unit-seal/v1\n"
        b"run_id=backend-unit-20260901T120000Z-a1b2c3d4e5f6-a1b2c3d4\n"
        b"owner_nonce=a1b2c3d4e5f60718293a4b5c6d7e8f90\n",
    ),
    (
        "extra field",
        b"hub.qa.backend-unit-seal/v1\n"
        b"run_id=backend-unit-20260901T120000Z-a1b2c3d4e5f6-a1b2c3d4\n"
        b"owner_nonce=a1b2c3d4e5f60718293a4b5c6d7e8f90\n"
        b"observer_receipt_sha256=sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee\n"
        b"extra=field\n",
    ),
    (
        "reordered fields",
        b"hub.qa.backend-unit-seal/v1\n"
        b"owner_nonce=a1b2c3d4e5f60718293a4b5c6d7e8f90\n"
        b"run_id=backend-unit-20260901T120000Z-a1b2c3d4e5f6-a1b2c3d4\n"
        b"observer_receipt_sha256=sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee\n",
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


def test_valid_fixture_bytes_cover_all_receipt_families_and_final_receipt() -> None:
    for family, fixture in _VALID_SCHEMA_FIXTURES:
        Draft202012Validator(_schema(family)).validate(_fixture_json(fixture))

    Draft202012Validator(_final_receipt_schema()).validate(
        _fixture_json("valid/final-receipt.json")
    )

    validate_package_plan(_fixture_json("valid/package-plan.json"))
    validate_container_receipt(_fixture_json("valid/container-receipt.json"))
    observer = _fixture_json("valid/observer-receipt.json")
    final = _fixture_json("valid/final-receipt.json")
    validate_observer_receipt(observer)
    validate_final_receipt(final)

    assert observer["final_receipt"] == final
    assert "observer_receipt_sha256" not in final
    assert observer["root_handoff_sha256"].startswith("sha256:")
    assert observer["concurrency_observation_scope"] == (
        "docker-ps-visible-containers-only"
    )
    for snapshot in observer["source_snapshots"].values():
        assert set(snapshot) == {
            "head",
            "status",
            "source_root_sha256",
            "frozen_subjects",
            "canonical_build_input_manifest_sha256",
        }
    final_artifact = next(
        artifact
        for artifact in observer["artifact_inventory"]
        if artifact["path"] == "final-receipt.json"
    )
    assert final_artifact == {
        "path": "final-receipt.json",
        "sha256": "sha256:"
        + hashlib.sha256(_fixture_bytes("valid/final-receipt.json")).hexdigest(),
        "size": len(_fixture_bytes("valid/final-receipt.json")),
    }


@pytest.mark.parametrize(("family", "fixture"), _STRUCTURAL_INVALID_FIXTURES)
def test_closed_schemas_reject_single_shape_mutations(
    family: str, fixture: str
) -> None:
    with pytest.raises(ValidationError):
        Draft202012Validator(_schema(family)).validate(_fixture_json(fixture))


@pytest.mark.parametrize(
    "field", ("root_handoff_sha256", "concurrency_observation_scope")
)
def test_observer_schema_requires_root_handoff_and_concurrency_scope(field: str) -> None:
    document = dict(_fixture_json("valid/observer-receipt.json"))
    document.pop(field)

    with pytest.raises(ValidationError):
        Draft202012Validator(_schema("observer_receipt")).validate(document)


def test_observer_schema_rejects_image_ids_nested_in_source_snapshots() -> None:
    document = json.loads(
        json.dumps(_fixture_json("valid/observer-receipt.json"))
    )
    document["source_snapshots"]["prebuild"]["fixed_image_ids"] = {
        "python_contract": "sha256:" + "c" * 64,
        "backend_unit": "sha256:" + "d" * 64,
    }

    with pytest.raises(ValidationError):
        Draft202012Validator(_schema("observer_receipt")).validate(document)


@pytest.mark.parametrize(("fixture", "validator"), _SEMANTIC_INVALID_FIXTURES)
def test_receipt_semantics_reject_cross_field_mutations(
    fixture: str, validator: Callable[[Mapping[str, object]], None]
) -> None:
    with pytest.raises(ReceiptContractError):
        validator(_fixture_json(fixture))


@pytest.mark.parametrize(("fixture", "expected_message"), _OBSERVER_SEMANTIC_FAILURES)
def test_observer_semantic_invalid_fixtures_fail_for_their_named_rule(
    fixture: str, expected_message: str
) -> None:
    with pytest.raises(ReceiptContractError) as raised:
        validate_observer_receipt(_fixture_json(fixture))

    assert str(raised.value) == expected_message


def test_timeout_outcome_is_accepted_only_with_its_frozen_sealed_classification() -> None:
    document = _fixture_json("valid/final-receipt.json")
    assert document["classification"] == "SEALED_TIMEOUT_UNTRIAGED"
    assert document["immutable_outcome"] == "TIMEOUT_UNTRIAGED"
    validate_final_receipt(document)


def test_parse_seal_accepts_only_canonical_lf_finalized_bytes() -> None:
    parsed = parse_seal(_fixture_bytes("valid/SEAL"))

    assert parsed == {
        "run_id": "backend-unit-20260901T120000Z-a1b2c3d4e5f6-a1b2c3d4",
        "owner_nonce": "a1b2c3d4e5f60718293a4b5c6d7e8f90",
        "observer_receipt_sha256": (
            "sha256:b77de46321d41915ed279dacfcb8acdf3d9002d4d3e418bf2629fbab0be416fe"
        ),
    }
    assert parsed["observer_receipt_sha256"] == (
        "sha256:"
        + hashlib.sha256(_fixture_bytes("valid/observer-receipt.json")).hexdigest()
    )

    with pytest.raises(ReceiptContractError):
        parse_seal(b"hub.qa.backend-unit-seal/v1\r\n")

    for _, payload in _INVALID_LF_SEAL_BYTES:
        with pytest.raises(ReceiptContractError):
            parse_seal(payload)
