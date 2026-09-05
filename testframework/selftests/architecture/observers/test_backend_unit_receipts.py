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


def _artifact(path: str) -> dict[str, object]:
    return {
        "path": path,
        "sha256": "sha256:" + hashlib.sha256(path.encode("utf-8")).hexdigest(),
        "size": len(path),
    }


def _rebind_abort_inventory(document: dict[str, object]) -> None:
    inventory = document["artifact_inventory"]
    assert type(inventory) is list
    inventory.sort(key=lambda record: str(record["path"]))
    payload = b"".join(
        (
            str(record["path"]) + "\t" + str(record["sha256"])
            + "\t" + str(record["size"]) + "\n"
        ).encode("ascii")
        for record in inventory
    )
    document["artifact_inventory_sha256"] = "sha256:" + hashlib.sha256(payload).hexdigest()


_ABORT_STREAM_CAPS = {
    "build": (2097152, 2097152),
    "vet": (2097152, 2097152),
    "unit": (16777216, 4194304),
    "race": (25165824, 4194304),
}


def _canonical_receipt_size(document: Mapping[str, object]) -> int:
    return len(
        json.dumps(
            document,
            allow_nan=False,
            ensure_ascii=False,
            separators=(",", ":"),
            sort_keys=True,
        ).encode("utf-8") + b"\n"
    )


def _bind_abort_budget(document: dict[str, object]) -> None:
    inventory = document["artifact_inventory"]
    abort = document["abort"]
    assert type(inventory) is list and type(abort) is dict
    sizes = {record["path"]: record["size"] for record in inventory}
    assert all(type(path) is str and type(size) is int for path, size in sizes.items())
    completed = abort["completed_stages"]
    assert type(completed) is list and all(type(stage) is str for stage in completed)
    streams: dict[str, object] = {}
    for stage, (stdout_cap, stderr_cap) in _ABORT_STREAM_CAPS.items():
        streams[stage] = {
            "stdout": {
                "cap_bytes": stdout_cap,
                "used_bytes": sizes[f"{stage}.stdout.log"] if stage in completed else 0,
                "truncated": False,
            },
            "stderr": {
                "cap_bytes": stderr_cap,
                "used_bytes": sizes[f"{stage}.stderr.log"] if stage in completed else 0,
                "truncated": False,
            },
        }
    document["report_budget"] = {
        "total_bytes": 67108864,
        "reserved_metadata_bytes": 8388608,
        "metadata_used_bytes": 0,
        "streams": streams,
        "total_used_bytes": 0,
        "truncated_due_to_budget": False,
        "raw_output_partial": False,
    }
    metadata_inventory = sum(
        size for path, size in sizes.items()
        if not (path.endswith(".stdout.log") or path.endswith(".stderr.log"))
    )
    inventory_total = sum(sizes.values())
    budget = document["report_budget"]
    assert type(budget) is dict
    for _ in range(16):
        receipt_size = _canonical_receipt_size(document)
        expected_metadata = metadata_inventory + receipt_size
        expected_total = inventory_total + receipt_size
        if (
            budget["metadata_used_bytes"] == expected_metadata
            and budget["total_used_bytes"] == expected_total
        ):
            return
        budget["metadata_used_bytes"] = expected_metadata
        budget["total_used_bytes"] = expected_total
    raise AssertionError("abort report budget fixed point did not converge")


def _aborted_container_document(
    phase: str, planned: tuple[str, ...], completed: tuple[str, ...], active_stage: str | None
) -> dict[str, object]:
    complete = json.loads(json.dumps(_fixture_json("valid/container-receipt.json")))
    prefix = (
        "schema", "target", "run_id", "owner_nonce", "container_hostname",
        "launcher_sha256", "command_profile_sha256",
        "source_manifest_before_sha256", "source_manifest_after_sha256",
        "cgroup", "go_environment",
    )
    document = {key: complete[key] for key in prefix}
    document["abort"] = {
        "signal": "TERM",
        "phase": phase,
        "active_stage": active_stage,
        "planned_stages": list(planned),
        "completed_stages": list(completed),
        "child_group_reaped": True,
        "stdout_reader_reaped": True,
        "stderr_reader_reaped": True,
        "unpublished_stage_rolled_back": True,
    }
    paths = {"CONTAINER_ABORTED", "owner.json"}
    paths.update(f"package-plan.{stage}.json" for stage in planned)
    for stage in completed:
        paths.update({
            f"{stage}.stderr.log", f"{stage}.stdout.log",
            f"package-results.{stage}.jsonl", f"stage-{stage}.json",
        })
    document["artifact_inventory"] = [_artifact(path) for path in sorted(paths)]
    document["artifact_inventory_sha256"] = "sha256:" + "0" * 64
    document["report_budget"] = complete["report_budget"]
    document["terminal"] = {"signal": "TERM", "exit_code": 143}
    document["state"] = "CONTAINER_ABORTED"
    _rebind_abort_inventory(document)
    _bind_abort_budget(document)
    return document


def _recursively_sorted(document: Mapping[str, object]) -> dict[str, object]:
    sorted_document = json.loads(json.dumps(
        document,
        allow_nan=False,
        ensure_ascii=False,
        separators=(",", ":"),
        sort_keys=True,
    ))
    assert type(sorted_document) is dict
    return sorted_document


@pytest.mark.parametrize(
    ("phase", "planned", "completed", "active_stage"),
    (
        ("bootstrap", (), (), None),
        ("planning", ("build",), (), None),
        ("execution", ("build", "vet", "unit", "race"), ("build",), "vet"),
    ),
)
def test_container_receipt_accepts_recursively_sort_keys_canonical_complete_and_abort(
    phase: str, planned: tuple[str, ...], completed: tuple[str, ...], active_stage: str | None
) -> None:
    complete = _recursively_sorted(_fixture_json("valid/container-receipt.json"))
    aborted = _recursively_sorted(_aborted_container_document(phase, planned, completed, active_stage))

    for document in (complete, aborted):
        Draft202012Validator(_schema("container_receipt")).validate(document)
        validate_container_receipt(document)


@pytest.mark.parametrize("mutation", ("missing", "extra"))
def test_container_receipt_recursively_sorted_keys_remain_closed(mutation: str) -> None:
    complete = _recursively_sorted(_fixture_json("valid/container-receipt.json"))
    if mutation == "missing":
        del complete["terminal"]
    else:
        complete["unexpected"] = None

    with pytest.raises(ReceiptContractError, match="complete container receipt keys are not closed and canonical"):
        validate_container_receipt(complete)


@pytest.mark.parametrize(
    ("phase", "planned", "completed", "active_stage"),
    (
        ("bootstrap", (), (), None),
        ("planning", (), (), None),
        ("planning", ("build",), (), None),
        ("execution", ("build", "vet", "unit", "race"), ("build",), "vet"),
        ("execution", ("build", "vet", "unit", "race"), ("build", "vet", "unit", "race"), None),
    ),
)
def test_container_receipt_accepts_legacy_complete_and_inline_sealed_abort_prefixes(
    phase: str, planned: tuple[str, ...], completed: tuple[str, ...], active_stage: str | None
) -> None:
    complete = _fixture_json("valid/container-receipt.json")
    aborted = _aborted_container_document(phase, planned, completed, active_stage)

    Draft202012Validator(_schema("container_receipt")).validate(complete)
    validate_container_receipt(complete)
    Draft202012Validator(_schema("container_receipt")).validate(aborted)
    validate_container_receipt(aborted)
    assert tuple(aborted) == (
        "schema", "target", "run_id", "owner_nonce", "container_hostname",
        "launcher_sha256", "command_profile_sha256",
        "source_manifest_before_sha256", "source_manifest_after_sha256",
        "cgroup", "go_environment", "abort", "artifact_inventory",
        "artifact_inventory_sha256", "report_budget", "terminal", "state",
    )


@pytest.mark.parametrize(
    ("mutation", "schema_valid", "expected_message"),
    (
        ("partial_stage_group", True, "aborted container inventory does not exactly derive from abort stages"),
        ("missing_marker", True, "aborted container inventory does not exactly derive from abort stages"),
        ("extra_artifact", True, "aborted container inventory does not exactly derive from abort stages"),
        ("nonprefix", False, "abort.planned_stages must be an ordered frozen stage prefix"),
        ("false_reap", False, "abort.child_group_reaped must prove reaping or rollback"),
        ("complete_crossover", False, "complete container receipt keys are not closed and canonical"),
        ("aborted_crossover", False, "aborted container receipt keys are not closed and canonical"),
        ("terminal_mismatch", False, "aborted container terminal must be TERM/143"),
        ("bootstrap_progress", False, "bootstrap abort must have no planned or completed stages and no active stage"),
        ("planning_full_plan", False, "planning abort must have an incomplete plan, no completed stages, and no active stage"),
        ("execution_short_plan", False, "execution abort must plan the complete frozen stage order"),
        ("execution_wrong_active", False, "execution abort active_stage must be the first unfinished stage or null"),
        ("execution_finished_active", False, "execution abort active_stage must be the first unfinished stage or null"),
        ("completed_stream_mismatch", True, "aborted report budget stream use does not bind its completed artifact size"),
        ("unfinished_stream_nonzero", True, "aborted report budget unfinished streams must be zero and untruncated"),
        ("completed_stream_truncated", True, "aborted report budget must not truncate completed stream artifacts"),
        ("fixed_point_metadata_mismatch", True, "aborted report budget metadata must include inventory metadata and canonical receipt bytes"),
    ),
)
def test_container_receipt_rejects_abort_branch_crossovers_and_incomplete_proof(
    mutation: str, schema_valid: bool, expected_message: str
) -> None:
    aborted = _aborted_container_document(
        "execution", ("build", "vet", "unit", "race"), ("build",), "vet"
    )
    document: dict[str, object]
    if mutation == "complete_crossover":
        document = json.loads(json.dumps(_fixture_json("valid/container-receipt.json")))
        document["abort"] = aborted["abort"]
    else:
        document = aborted
        if mutation == "partial_stage_group":
            document["artifact_inventory"] = [
                record for record in document["artifact_inventory"]
                if record["path"] != "build.stdout.log"
            ]
            _rebind_abort_inventory(document)
        elif mutation == "missing_marker":
            document["artifact_inventory"] = [
                record for record in document["artifact_inventory"]
                if record["path"] != "CONTAINER_ABORTED"
            ]
            _rebind_abort_inventory(document)
        elif mutation == "extra_artifact":
            document["artifact_inventory"].append(_artifact("scratch.tmp"))
            _rebind_abort_inventory(document)
        elif mutation == "nonprefix":
            document["abort"]["planned_stages"] = ["build", "unit"]
        elif mutation == "false_reap":
            document["abort"]["child_group_reaped"] = False
        elif mutation == "aborted_crossover":
            document["package_plan_sha256"] = {
                stage: "sha256:" + "a" * 64
                for stage in ("build", "vet", "unit", "race")
            }
        elif mutation == "terminal_mismatch":
            document["terminal"] = {"signal": "TERM", "exit_code": 0}
        elif mutation == "bootstrap_progress":
            document["abort"]["phase"] = "bootstrap"
            document["abort"]["planned_stages"] = ["build"]
            document["abort"]["completed_stages"] = []
            document["abort"]["active_stage"] = None
        elif mutation == "planning_full_plan":
            document["abort"]["phase"] = "planning"
            document["abort"]["completed_stages"] = []
            document["abort"]["active_stage"] = None
        elif mutation == "execution_short_plan":
            document["abort"]["planned_stages"] = ["build"]
            document["abort"]["completed_stages"] = []
            document["abort"]["active_stage"] = "build"
        elif mutation == "execution_wrong_active":
            document["abort"]["active_stage"] = "race"
        elif mutation == "execution_finished_active":
            document["abort"]["completed_stages"] = ["build", "vet", "unit", "race"]
            document["abort"]["active_stage"] = "race"
        elif mutation == "completed_stream_mismatch":
            budget = document["report_budget"]
            assert type(budget) is dict
            streams = budget["streams"]
            assert type(streams) is dict
            build = streams["build"]
            assert type(build) is dict
            stdout = build["stdout"]
            assert type(stdout) is dict
            previous_used = stdout["used_bytes"]
            assert type(previous_used) is int
            stdout["used_bytes"] = 0
            budget["total_used_bytes"] -= previous_used
        elif mutation == "unfinished_stream_nonzero":
            budget = document["report_budget"]
            assert type(budget) is dict
            streams = budget["streams"]
            assert type(streams) is dict
            vet = streams["vet"]
            assert type(vet) is dict
            stdout = vet["stdout"]
            assert type(stdout) is dict
            stdout["used_bytes"] = 1
            budget["total_used_bytes"] += 1
        elif mutation == "completed_stream_truncated":
            for record in document["artifact_inventory"]:
                if record["path"] == "build.stdout.log":
                    record["size"] = 2097152
                    break
            else:
                raise AssertionError("completed stdout artifact missing")
            _rebind_abort_inventory(document)
            _bind_abort_budget(document)
            budget = document["report_budget"]
            assert type(budget) is dict
            streams = budget["streams"]
            assert type(streams) is dict
            build = streams["build"]
            assert type(build) is dict
            stdout = build["stdout"]
            assert type(stdout) is dict
            stdout["truncated"] = True
            budget["truncated_due_to_budget"] = True
            budget["raw_output_partial"] = True
        elif mutation == "fixed_point_metadata_mismatch":
            budget = document["report_budget"]
            assert type(budget) is dict
            budget["metadata_used_bytes"] += 1
            budget["total_used_bytes"] += 1

    if schema_valid:
        Draft202012Validator(_schema("container_receipt")).validate(document)
    else:
        with pytest.raises(ValidationError):
            Draft202012Validator(_schema("container_receipt")).validate(document)
    with pytest.raises(ReceiptContractError) as raised:
        validate_container_receipt(document)
    assert str(raised.value) == expected_message


def test_complete_container_inventory_excludes_its_receipt_bytes() -> None:
    complete = json.loads(json.dumps(_fixture_json("valid/container-receipt.json")))
    inventory = complete["artifact_inventory"]
    assert type(inventory) is list
    inventory.append(_artifact("container-receipt.json"))
    _rebind_abort_inventory(complete)

    Draft202012Validator(_schema("container_receipt")).validate(complete)
    with pytest.raises(ReceiptContractError) as raised:
        validate_container_receipt(complete)
    assert str(raised.value) == "complete container inventory must exclude its receipt bytes"


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
