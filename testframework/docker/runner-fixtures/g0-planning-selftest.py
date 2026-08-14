#!/usr/bin/env python3
"""Package-C-only image fixture for the shared A0 dependency-admission code."""

from __future__ import annotations

import hashlib
import json
import pathlib
import shutil
import sys
from typing import Callable, Final

ADMISSION_ROOT: Final[pathlib.Path] = pathlib.Path("/opt/hub/admission")
if str(ADMISSION_ROOT) not in sys.path:
    sys.path.insert(0, str(ADMISSION_ROOT))

from verify_local_dependency_bundle import verify_bundle, BundleVerificationError, SOURCE_COMMIT
from admit_local_dependency_assets import admit_bundle, AdmissionError


REPORT_ROOT: Final[pathlib.Path] = pathlib.Path("/reports/g0-planning-selftest")
FIXTURE_ROOT: Final[pathlib.Path] = REPORT_ROOT / "shared-admission-fixtures"
MANIFEST_NAME: Final[str] = "dependency-bundle.manifest.v1.json"
BUNDLED_SCHEMA_PATH: Final[pathlib.Path] = pathlib.Path(
    "/opt/hub/testframework/docker/dependency-bundle.schema.v1.json"
)
FIXTURE_CASES: Final[tuple[str, ...]] = (
    "valid-manifest",
    "duplicate-asset-id",
    "asset-hash-mismatch",
    "asset-path-escape",
    "missing-asset",
    "unbound-manifest",
)
ASSET_TARGETS: Final[tuple[tuple[str, tuple[str, ...]], ...]] = (
    ("go-module-cache", ("backend-unit", "backend-openapi")),
    ("go-race-toolchain", ("backend-unit",)),
    ("frontend-node-modules", ("frontend-unit",)),
    ("swag-binary", ("backend-openapi",)),
)


def load_baked_schema() -> dict[str, object]:
    raw = BUNDLED_SCHEMA_PATH.read_bytes()
    schema = json.loads(raw.decode("utf-8"))
    if not isinstance(schema, dict):
        raise RuntimeError("baked dependency schema is not an object")
    if schema.get("$id") != "hub.qa.dependency-bundle/v1":
        raise RuntimeError("baked dependency schema id is not closed")
    required = schema.get("required")
    if required != ["schema", "bundle_id", "source_commit", "platform", "assets", "bases", "bundle_sha256"]:
        raise RuntimeError("baked dependency schema root keys are not closed")
    definitions = schema.get("$defs")
    if not isinstance(definitions, dict):
        raise RuntimeError("baked dependency schema definitions are unavailable")
    asset = definitions.get("asset")
    base = definitions.get("base")
    if not isinstance(asset, dict) or not isinstance(base, dict):
        raise RuntimeError("baked dependency schema asset/base definitions are unavailable")
    if asset.get("required") != [
        "asset_id", "targets", "relative_path", "sha256", "bytes", "license_spdx",
        "license_evidence_sha256", "provenance_sha256",
    ] or base.get("required") != [
        "base_id", "consumers", "local_image_id", "manifest_digest", "license_spdx",
        "license_evidence_sha256", "provenance_sha256",
    ]:
        raise RuntimeError("baked dependency schema nested keys are not closed")
    return schema


BAKED_SCHEMA: Final[dict[str, object]] = load_baked_schema()


def assert_manifest_against_baked_schema(manifest: dict[str, object]) -> None:
    required = BAKED_SCHEMA["required"]
    definitions = BAKED_SCHEMA["$defs"]
    if list(manifest.keys()) != required:
        raise RuntimeError("fixture manifest does not match baked schema root order")
    if manifest["schema"] != BAKED_SCHEMA["properties"]["schema"]["const"]:
        raise RuntimeError("fixture manifest schema identity does not match baked schema")
    if not isinstance(manifest["assets"], list) or not isinstance(manifest["bases"], list):
        raise RuntimeError("fixture manifest arrays do not match baked schema")
    asset_required = definitions["asset"]["required"]
    base_required = definitions["base"]["required"]
    if any(not isinstance(item, dict) or list(item.keys()) != asset_required for item in manifest["assets"]):
        raise RuntimeError("fixture asset records do not match baked schema")
    if any(not isinstance(item, dict) or list(item.keys()) != base_required for item in manifest["bases"]):
        raise RuntimeError("fixture base records do not match baked schema")


def sha256_prefixed(raw: bytes) -> str:
    return "sha256:" + hashlib.sha256(raw).hexdigest()


def write_json(path: pathlib.Path, value: object) -> bytes:
    raw = (json.dumps(value, separators=(",", ":"), ensure_ascii=False) + "\n").encode("utf-8")
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_bytes(raw)
    return raw


def canonical_bundle_sha256(manifest: dict[str, object]) -> str:
    projection = {key: value for key, value in manifest.items() if key != "bundle_sha256"}
    return sha256_prefixed(json.dumps(projection, sort_keys=False, separators=(",", ":"), ensure_ascii=False).encode("utf-8"))


def write_task_owner(bootstrap_root: pathlib.Path) -> bytes:
    return write_json(
        bootstrap_root / "owner-marker.v1.json",
        {
            "schema": "hub.qa.bootstrap-owner/v1",
            "task": "g0-assets-runner",
            "run": "g0-planning-fixture",
            "owner_nonce": "1" * 32,
        },
    )


def write_handoff_owner(ingress_root: pathlib.Path, manifest_raw: bytes) -> bytes:
    owner = {
        "schema": "hub.qa.g0-offline-ingress-owner/v1",
        "handoff_id": "g0-fixture-handoff",
        "source_commit": SOURCE_COMMIT,
        "dependency_bundle_manifest_sha256": sha256_prefixed(manifest_raw),
        "owner": "g0-planning-selftest",
        "released_at_utc": "2026-08-14T00:00:00Z",
    }
    return write_json(ingress_root / "handoff-owner.v1.json", owner)


def write_valid_mismatched_handoff_owner(ingress_root: pathlib.Path, manifest_raw: bytes) -> str:
    owner = {
        "schema": "hub.qa.g0-offline-ingress-owner/v1",
        "handoff_id": "g0-fixture-handoff-other",
        "source_commit": SOURCE_COMMIT,
        "dependency_bundle_manifest_sha256": sha256_prefixed(manifest_raw),
        "owner": "g0-planning-selftest",
        "released_at_utc": "2026-08-14T00:00:00Z",
    }
    return sha256_prefixed(write_json(ingress_root / "handoff-owner.v1.json", owner))


def write_bound_manifest(
    manifest_path: pathlib.Path, ingress_root: pathlib.Path, manifest: dict[str, object]
) -> str:
    assert_manifest_against_baked_schema(manifest)
    manifest_raw = write_json(manifest_path, manifest)
    ingress_manifest_path = ingress_root / MANIFEST_NAME
    ingress_manifest_path.parent.mkdir(parents=True, exist_ok=True)
    ingress_manifest_path.write_bytes(manifest_raw)
    return sha256_prefixed(write_handoff_owner(ingress_root, manifest_raw))


def build_valid_case(case_name: str) -> tuple[pathlib.Path, pathlib.Path, pathlib.Path, str, str, str, str, pathlib.Path]:
    case_root = FIXTURE_ROOT / case_name
    if case_root.exists():
        raise RuntimeError(f"fixture root already exists: {case_name}")
    ingress_root = case_root / "ingress" / "dependency-bundle"
    controls_root = case_root / "controls"
    bootstrap_root = case_root / "bootstrap"
    asset_records: list[dict[str, object]] = []
    for index, (asset_id, targets) in enumerate(ASSET_TARGETS):
        relative_path = f"assets/{asset_id}"
        payload = f"g0-planning-fixture:{asset_id}:{index}\n".encode("ascii")
        asset_path = ingress_root / relative_path
        asset_path.parent.mkdir(parents=True, exist_ok=True)
        asset_path.write_bytes(payload)
        license_evidence = f"fixture license evidence for {asset_id}\n".encode("ascii")
        provenance_evidence = (json.dumps({"asset_id": asset_id}, separators=(",", ":")) + "\n").encode("utf-8")
        license_path = ingress_root / "evidence" / "licenses" / "assets" / f"{asset_id}.txt"
        provenance_path = ingress_root / "evidence" / "provenance" / "assets" / f"{asset_id}.json"
        license_path.parent.mkdir(parents=True, exist_ok=True)
        provenance_path.parent.mkdir(parents=True, exist_ok=True)
        license_path.write_bytes(license_evidence)
        provenance_path.write_bytes(provenance_evidence)
        asset_records.append(
            {
                "asset_id": asset_id,
                "targets": list(targets),
                "relative_path": relative_path,
                "sha256": sha256_prefixed(payload),
                "bytes": len(payload),
                "license_spdx": "MIT",
                "license_evidence_sha256": sha256_prefixed(license_evidence),
                "provenance_sha256": sha256_prefixed(provenance_evidence),
            }
        )

    base_specs = (
        ("go-backend-unit", ("backend-unit",), "1", "2"),
        ("go-backend-openapi", ("backend-openapi",), "3", "4"),
        ("frontend-unit", ("frontend-unit",), "5", "6"),
        ("bootstrap-helper-python", ("g0-admission-helper", "g0-planning-selftest"), "7", "8"),
    )
    receipt_nonce = "3" * 32
    bases: list[dict[str, object]] = []
    receipt_bases: list[dict[str, str]] = []
    for base_id, consumers, image_marker, digest_marker in base_specs:
        base_license = f"fixture base license evidence for {base_id}\n".encode("ascii")
        base_provenance = (json.dumps({"base_id": base_id}, separators=(",", ":")) + "\n").encode("utf-8")
        base_license_path = ingress_root / "evidence" / "licenses" / "bases" / f"{base_id}.txt"
        base_provenance_path = ingress_root / "evidence" / "provenance" / "bases" / f"{base_id}.json"
        base_license_path.parent.mkdir(parents=True, exist_ok=True)
        base_provenance_path.parent.mkdir(parents=True, exist_ok=True)
        base_license_path.write_bytes(base_license)
        base_provenance_path.write_bytes(base_provenance)
        base_image_id = "sha256:" + image_marker * 64
        base_manifest_digest = "sha256:" + digest_marker * 64
        bases.append(
            {
                "base_id": base_id,
                "consumers": list(consumers),
                "local_image_id": base_image_id,
                "manifest_digest": base_manifest_digest,
                "license_spdx": "MIT",
                "license_evidence_sha256": sha256_prefixed(base_license),
                "provenance_sha256": sha256_prefixed(base_provenance),
            }
        )
        receipt_bases.append(
            {
                "base_id": base_id,
                "local_image_id": base_image_id,
                "manifest_digest": base_manifest_digest,
            }
        )
    base_receipt_path = bootstrap_root / "local-base-inspect.v1.json"
    manifest: dict[str, object] = {
        "schema": "hub.qa.dependency-bundle/v1",
        "bundle_id": "g0-fixture-valid",
        "source_commit": SOURCE_COMMIT,
        "platform": {"os": "linux", "architecture": "amd64", "variant": None},
        "assets": asset_records,
        "bases": bases,
    }
    manifest["bundle_sha256"] = canonical_bundle_sha256(manifest)
    manifest_path = controls_root / MANIFEST_NAME
    task_owner_raw = write_task_owner(bootstrap_root)
    task_owner_marker_sha256 = sha256_prefixed(task_owner_raw)
    handoff_owner_marker_sha256 = write_bound_manifest(manifest_path, ingress_root, manifest)
    handoff_owner_raw = (ingress_root / "handoff-owner.v1.json").read_bytes()
    if task_owner_raw == handoff_owner_raw or task_owner_marker_sha256 == handoff_owner_marker_sha256:
        raise RuntimeError("fixture task and ingress handoff owners must remain distinct")
    base_receipt = {
        "schema": "hub.qa.local-base-inspect/v1",
        "task": "g0-assets-runner",
        "run": "g0-planning-fixture",
        "owner_marker_sha256": task_owner_marker_sha256,
        "receipt_nonce": receipt_nonce,
        "bases": receipt_bases,
    }
    base_receipt_raw = write_json(base_receipt_path, base_receipt)
    return (
        manifest_path,
        ingress_root,
        base_receipt_path,
        sha256_prefixed(base_receipt_raw),
        receipt_nonce,
        task_owner_marker_sha256,
        handoff_owner_marker_sha256,
        case_root,
    )


def rewrite_manifest(
    manifest_path: pathlib.Path,
    ingress_root: pathlib.Path,
    base_receipt_path: pathlib.Path,
    mutate: Callable[[dict[str, object]], None],
) -> None:
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    mutate(manifest)
    manifest["bundle_sha256"] = canonical_bundle_sha256(manifest)
    write_bound_manifest(manifest_path, ingress_root, manifest)


def reason_code(error: BaseException) -> str:
    code = getattr(error, "code", None)
    return code if isinstance(code, str) else str(error)


def verify_with_bindings(
    manifest_path: pathlib.Path,
    ingress_root: pathlib.Path,
    receipt_path: pathlib.Path,
    receipt_sha256: str,
    receipt_nonce: str,
    expected_task_owner_marker_sha256: str | None,
    expected_handoff_owner_marker_sha256: str | None,
) -> dict[str, object]:
    return verify_bundle(
        manifest_path=manifest_path,
        ingress_root=ingress_root,
        base_receipt_path=receipt_path,
        base_receipt_sha256=receipt_sha256,
        receipt_nonce=receipt_nonce,
        expected_task_owner_marker_sha256=expected_task_owner_marker_sha256,
        expected_handoff_owner_marker_sha256=expected_handoff_owner_marker_sha256,
    )


def admit_with_bindings(
    manifest_path: pathlib.Path,
    ingress_root: pathlib.Path,
    receipt_path: pathlib.Path,
    receipt_sha256: str,
    receipt_nonce: str,
    success_receipt_path: pathlib.Path,
    expected_task_owner_marker_sha256: str | None,
    expected_handoff_owner_marker_sha256: str | None,
) -> dict[str, object]:
    return admit_bundle(
        manifest_path=manifest_path,
        ingress_root=ingress_root,
        base_receipt_path=receipt_path,
        base_receipt_sha256=receipt_sha256,
        receipt_nonce=receipt_nonce,
        success_receipt_path=success_receipt_path,
        expected_task_owner_marker_sha256=expected_task_owner_marker_sha256,
        expected_handoff_owner_marker_sha256=expected_handoff_owner_marker_sha256,
    )


def assert_rejected(case_name: str, expected_code: str) -> None:
    (
        manifest_path,
        ingress_root,
        receipt_path,
        receipt_sha256,
        receipt_nonce,
        task_owner_marker_sha256,
        handoff_owner_marker_sha256,
        case_root,
    ) = build_valid_case(case_name)
    if case_name == "duplicate-asset-id":
        rewrite_manifest(manifest_path, ingress_root, receipt_path, lambda value: value["assets"].append(dict(value["assets"][0])))
    elif case_name == "asset-hash-mismatch":
        rewrite_manifest(manifest_path, ingress_root, receipt_path, lambda value: value["assets"][0].update({"sha256": "sha256:" + "0" * 64}))
    elif case_name == "asset-path-escape":
        rewrite_manifest(manifest_path, ingress_root, receipt_path, lambda value: value["assets"][0].update({"relative_path": "../escape.bin"}))
    elif case_name == "missing-asset":
        (ingress_root / "assets" / "go-module-cache").unlink()
    elif case_name == "unbound-manifest":
        manifest_path.unlink()
    else:
        raise RuntimeError(f"unknown rejected fixture: {case_name}")

    receipt_sha256 = sha256_prefixed(receipt_path.read_bytes())
    task_owner_marker_sha256 = sha256_prefixed((case_root / "bootstrap" / "owner-marker.v1.json").read_bytes())
    handoff_owner_marker_sha256 = sha256_prefixed((ingress_root / "handoff-owner.v1.json").read_bytes())

    try:
        verify_with_bindings(
            manifest_path,
            ingress_root,
            receipt_path,
            receipt_sha256,
            receipt_nonce,
            task_owner_marker_sha256,
            handoff_owner_marker_sha256,
        )
    except BundleVerificationError as error:
        if reason_code(error) != expected_code:
            raise RuntimeError(f"{case_name} returned {reason_code(error)} instead of {expected_code}") from error
    else:
        raise RuntimeError(f"{case_name} unexpectedly verified")

    success_receipt_path = case_root / "admission" / "success-receipt.json"
    try:
        admit_with_bindings(
            manifest_path,
            ingress_root,
            receipt_path,
            receipt_sha256,
            receipt_nonce,
            success_receipt_path,
            task_owner_marker_sha256,
            handoff_owner_marker_sha256,
        )
    except (BundleVerificationError, AdmissionError) as error:
        if reason_code(error) != expected_code:
            raise RuntimeError(f"{case_name} admission returned {reason_code(error)} instead of {expected_code}") from error
    else:
        raise RuntimeError(f"{case_name} admission unexpectedly succeeded")
    if success_receipt_path.exists():
        raise RuntimeError(f"{case_name} wrote an admission success receipt after rejection")


def assert_valid() -> None:
    (
        manifest_path,
        ingress_root,
        receipt_path,
        receipt_sha256,
        receipt_nonce,
        task_owner_marker_sha256,
        handoff_owner_marker_sha256,
        case_root,
    ) = build_valid_case("valid-manifest")
    verified = verify_with_bindings(
        manifest_path,
        ingress_root,
        receipt_path,
        receipt_sha256,
        receipt_nonce,
        task_owner_marker_sha256,
        handoff_owner_marker_sha256,
    )
    if not isinstance(verified, dict):
        raise RuntimeError("valid-manifest verifier result is not an object")
    success_receipt_path = case_root / "admission" / "success-receipt.json"
    admitted = admit_with_bindings(
        manifest_path,
        ingress_root,
        receipt_path,
        receipt_sha256,
        receipt_nonce,
        success_receipt_path,
        task_owner_marker_sha256,
        handoff_owner_marker_sha256,
    )
    if not isinstance(admitted, dict) or not success_receipt_path.is_file():
        raise RuntimeError("valid-manifest did not create an admission success receipt")


def assert_binding_rejection(case_name: str, binding_kind: str) -> None:
    (
        manifest_path,
        ingress_root,
        receipt_path,
        receipt_sha256,
        receipt_nonce,
        task_owner_marker_sha256,
        handoff_owner_marker_sha256,
        case_root,
    ) = build_valid_case(case_name)
    if binding_kind == "null-task-owner":
        expected_task_owner_marker_sha256: str | None = None
        expected_handoff_owner_marker_sha256: str | None = handoff_owner_marker_sha256
    elif binding_kind == "null-handoff-owner":
        expected_task_owner_marker_sha256 = task_owner_marker_sha256
        expected_handoff_owner_marker_sha256 = None
    elif binding_kind == "equal-owners":
        expected_task_owner_marker_sha256 = handoff_owner_marker_sha256
        expected_handoff_owner_marker_sha256 = handoff_owner_marker_sha256
    elif binding_kind == "mismatch-task-owner":
        expected_task_owner_marker_sha256 = "sha256:" + "0" * 64
        expected_handoff_owner_marker_sha256 = handoff_owner_marker_sha256
    elif binding_kind == "mismatch-handoff-owner":
        wrong_ingress_root = case_root / "other-valid-ingress"
        shutil.copytree(ingress_root, wrong_ingress_root)
        wrong_handoff_owner_marker_sha256 = write_valid_mismatched_handoff_owner(
            wrong_ingress_root,
            (wrong_ingress_root / MANIFEST_NAME).read_bytes(),
        )
        if wrong_handoff_owner_marker_sha256 == handoff_owner_marker_sha256:
            raise RuntimeError("mismatch-handoff-owner fixture did not produce a distinct valid ingress owner")
        expected_task_owner_marker_sha256 = task_owner_marker_sha256
        expected_handoff_owner_marker_sha256 = wrong_handoff_owner_marker_sha256
    else:
        raise RuntimeError(f"unknown binding probe: {binding_kind}")
    try:
        verify_with_bindings(
            manifest_path,
            ingress_root,
            receipt_path,
            receipt_sha256,
            receipt_nonce,
            expected_task_owner_marker_sha256,
            expected_handoff_owner_marker_sha256,
        )
    except BundleVerificationError as error:
        if reason_code(error) != "E_RECEIPT_BINDING_INVALID":
            raise RuntimeError(f"{case_name} verifier returned {reason_code(error)}") from error
    else:
        raise RuntimeError(f"{case_name} verifier unexpectedly accepted owner bindings")

    success_receipt_path = case_root / "admission" / "success-receipt.json"
    try:
        admit_with_bindings(
            manifest_path,
            ingress_root,
            receipt_path,
            receipt_sha256,
            receipt_nonce,
            success_receipt_path,
            expected_task_owner_marker_sha256,
            expected_handoff_owner_marker_sha256,
        )
    except (BundleVerificationError, AdmissionError) as error:
        if reason_code(error) != "E_RECEIPT_BINDING_INVALID":
            raise RuntimeError(f"{case_name} admission returned {reason_code(error)}") from error
    else:
        raise RuntimeError(f"{case_name} admission unexpectedly accepted owner bindings")
    if success_receipt_path.exists():
        raise RuntimeError(f"{case_name} wrote an admission success receipt after binding rejection")


def assert_omitted_binding_rejections() -> None:
    (
        manifest_path,
        ingress_root,
        receipt_path,
        receipt_sha256,
        receipt_nonce,
        task_owner_marker_sha256,
        handoff_owner_marker_sha256,
        case_root,
    ) = build_valid_case("binding-omitted")
    omitted_calls: tuple[tuple[str, Callable[[], object]], ...] = (
        (
            "verifier-task-owner",
            lambda: verify_bundle(
                manifest_path=manifest_path,
                ingress_root=ingress_root,
                base_receipt_path=receipt_path,
                base_receipt_sha256=receipt_sha256,
                receipt_nonce=receipt_nonce,
                expected_handoff_owner_marker_sha256=handoff_owner_marker_sha256,
            ),
        ),
        (
            "verifier-handoff-owner",
            lambda: verify_bundle(
                manifest_path=manifest_path,
                ingress_root=ingress_root,
                base_receipt_path=receipt_path,
                base_receipt_sha256=receipt_sha256,
                receipt_nonce=receipt_nonce,
                expected_task_owner_marker_sha256=task_owner_marker_sha256,
            ),
        ),
        (
            "admission-task-owner",
            lambda: admit_bundle(
                manifest_path=manifest_path,
                ingress_root=ingress_root,
                base_receipt_path=receipt_path,
                base_receipt_sha256=receipt_sha256,
                receipt_nonce=receipt_nonce,
                success_receipt_path=case_root / "admission" / "omitted-task-owner.json",
                expected_handoff_owner_marker_sha256=handoff_owner_marker_sha256,
            ),
        ),
        (
            "admission-handoff-owner",
            lambda: admit_bundle(
                manifest_path=manifest_path,
                ingress_root=ingress_root,
                base_receipt_path=receipt_path,
                base_receipt_sha256=receipt_sha256,
                receipt_nonce=receipt_nonce,
                success_receipt_path=case_root / "admission" / "omitted-handoff-owner.json",
                expected_task_owner_marker_sha256=task_owner_marker_sha256,
            ),
        ),
    )
    for label, operation in omitted_calls:
        try:
            operation()
        except TypeError:
            continue
        except BundleVerificationError as error:
            raise RuntimeError(f"binding-omitted {label} returned {reason_code(error)} instead of TypeError") from error
        raise RuntimeError(f"binding-omitted {label} did not require both owner bindings")


def assert_owner_binding_probes() -> None:
    assert_omitted_binding_rejections()
    assert_binding_rejection("binding-null-task-owner", "null-task-owner")
    assert_binding_rejection("binding-null-handoff-owner", "null-handoff-owner")
    assert_binding_rejection("binding-equal-owners", "equal-owners")
    assert_binding_rejection("binding-mismatch-task-owner", "mismatch-task-owner")
    assert_binding_rejection("binding-mismatch-handoff-owner", "mismatch-handoff-owner")


def main() -> int:
    if FIXTURE_ROOT.exists():
        raise RuntimeError("planning fixture root already exists")
    try:
        assert_valid()
        assert_owner_binding_probes()
        assert_rejected("duplicate-asset-id", "E_DUPLICATE_ASSET_ID")
        assert_rejected("asset-hash-mismatch", "E_ASSET_SHA256_MISMATCH")
        assert_rejected("asset-path-escape", "E_ASSET_PATH_ESCAPE")
        assert_rejected("missing-asset", "E_ASSET_MISSING")
        assert_rejected("unbound-manifest", "E_MANIFEST_SCHEMA")
        print(json.dumps({"target": "g0-planning-selftest", "status": "PASS", "fixtures": list(FIXTURE_CASES)}, separators=(",", ":")))
        return 0
    finally:
        if FIXTURE_ROOT.exists():
            shutil.rmtree(FIXTURE_ROOT)


if __name__ == "__main__":
    raise SystemExit(main())
