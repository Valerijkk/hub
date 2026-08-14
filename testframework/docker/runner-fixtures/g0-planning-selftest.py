#!/usr/bin/env python3
"""Package-C-only negative-red/positive-green fixture for archive admission."""

from __future__ import annotations

import gzip
import hashlib
import json
import os
import pathlib
import shutil
import sys
from typing import Any, Callable, Final

ADMISSION_ROOT: Final[pathlib.Path] = pathlib.Path("/opt/hub/admission")
if str(ADMISSION_ROOT) not in sys.path:
    sys.path.insert(0, str(ADMISSION_ROOT))

import verify_local_dependency_bundle as verifier
import admit_local_dependency_assets as admission
from admit_local_dependency_assets import admit_bundle, AdmissionError
from verify_local_dependency_bundle import verify_bundle, BundleVerificationError, SOURCE_COMMIT

REPORT_ROOT: Final[pathlib.Path] = pathlib.Path("/reports/g0-planning-selftest")
FIXTURE_ROOT: Final[pathlib.Path] = REPORT_ROOT / "shared-admission-fixtures"
MANIFEST_NAME: Final[str] = "dependency-bundle.manifest.v1.json"
ADMISSION_RECEIPT_NAME: Final[str] = "g0-admission-helper.admission.v1.json"
BUNDLED_SCHEMA_PATH: Final[pathlib.Path] = pathlib.Path("/opt/hub/testframework/docker/dependency-bundle.schema.v1.json")
FIXTURE_CASES: Final[tuple[str, ...]] = (
    "valid-manifest", "duplicate-asset-id", "asset-hash-mismatch", "asset-path-escape", "missing-asset", "unbound-manifest",
    "archive-lstat-regular-only", "archive-safe-handle-required", "archive-replaced-between-passes", "archive-replaced-before-staging",
    "archive-stream-cap-before-tar", "archive-compressed", "archive-truncated-tail", "archive-terminal-missing-zero", "archive-terminal-eof", "archive-terminal-trailing-byte", "archive-checksum-octal", "archive-posix-octal-drift",
    "archive-forbidden-member", "archive-noncanonical-order", "archive-pax-closed", "archive-pax-empty", "archive-pax-invalid-utf8", "archive-pax-prefix-cap", "archive-pax-duplicate-key", "archive-pax-nonpath", "archive-raw-tarfile-crosscheck",
    "archive-marker-absence", "archive-layout-bounds",
    "staging-target-order", "staging-target-extra", "staging-target-missing", "staging-target-collision", "staging-source-identity-mismatch", "staging-rehash-growth", "archive-observation-phase-hooks", "receipt-ancestor-swap",
)
ASSET_SPECS: Final[tuple[tuple[str, tuple[str, ...], str], ...]] = (
    ("go-module-cache", ("backend-unit", "backend-openapi"), "assets/go-module-cache.tar"),
    ("go-race-toolchain", ("backend-unit",), "assets/go-race-toolchain.tar"),
    ("frontend-node-modules", ("frontend-unit",), "assets/frontend-node-modules.tar"),
    ("swag-binary", ("backend-openapi",), "assets/swag-binary.tar"),
)


def sha256_prefixed(raw: bytes) -> str:
    return "sha256:" + hashlib.sha256(raw).hexdigest()


def write_json(path: pathlib.Path, value: object) -> bytes:
    raw = (json.dumps(value, separators=(",", ":"), ensure_ascii=False) + "\n").encode("utf-8")
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_bytes(raw)
    return raw


def _octal(value: int, width: int) -> bytes:
    encoded = f"{value:0{width - 1}o}".encode("ascii") + b"\0"
    if len(encoded) != width:
        raise RuntimeError("fixture octal field overflow")
    return encoded


def _text(value: str, width: int) -> bytes:
    encoded = value.encode("utf-8")
    if len(encoded) > width:
        raise RuntimeError("fixture tar field overflow")
    return encoded + b"\0" * (width - len(encoded))


def tar_header(name: str, kind: str, mode: int, size: int) -> bytes:
    header = bytearray(512)
    header[0:100] = _text(name, 100)
    header[100:108] = _octal(mode, 8)
    header[108:116] = _octal(0, 8)
    header[116:124] = _octal(0, 8)
    header[124:136] = _octal(size, 12)
    header[136:148] = _octal(0, 12)
    header[148:156] = b" " * 8
    header[156:157] = kind.encode("ascii")
    header[257:263] = b"ustar\0"
    header[263:265] = b"00"
    header[329:337] = _octal(0, 8)
    header[337:345] = _octal(0, 8)
    checksum = sum(header)
    header[148:156] = f"{checksum:06o}\0 ".encode("ascii")
    return bytes(header)


def tar_entry(name: str, kind: str, mode: int, payload: bytes = b"") -> bytes:
    size = len(payload) if kind != "5" else 0
    return tar_header(name, kind, mode, size) + payload + b"\0" * ((-size) % 512)


def tar_archive(entries: list[tuple[str, str, int, bytes]]) -> bytes:
    return b"".join(tar_entry(*entry) for entry in entries) + b"\0" * 1024


def pax_payload(key: str, value: str) -> bytes:
    tail = f" {key}={value}\n".encode("utf-8")
    length = len(tail) + 1
    while True:
        candidate = str(length).encode("ascii") + tail
        if len(candidate) == length:
            return candidate
        length = len(candidate)


def archive_entries(asset_id: str, *, omit_marker: bool = False, noncanonical: bool = False) -> list[tuple[str, str, int, bytes]]:
    if asset_id == "go-module-cache":
        entries = [
            ("pkg", "5", 0o755, b""), ("pkg/mod", "5", 0o755, b""), ("pkg/mod/cache", "5", 0o755, b""),
            ("pkg/mod/cache/download", "5", 0o755, b""), ("pkg/mod/cache/download/fixture", "0", 0o644, b"cache\n"),
        ]
    elif asset_id == "go-race-toolchain":
        entries = [("bin", "5", 0o755, b""), ("bin/cc", "0", 0o755, b"cc\n"), ("bin/go", "0", 0o755, b"go\n")]
    elif asset_id == "frontend-node-modules":
        entries = [
            ("node_modules", "5", 0o755, b""), ("node_modules/@biomejs", "5", 0o755, b""),
            ("node_modules/@biomejs/biome", "5", 0o755, b""), ("node_modules/@biomejs/biome/bin", "5", 0o755, b""),
            ("node_modules/@biomejs/biome/bin/biome", "0", 0o755, b"biome\n"),
            ("node_modules/typescript", "5", 0o755, b""), ("node_modules/typescript/bin", "5", 0o755, b""),
            ("node_modules/typescript/bin/tsc", "0", 0o644, b"tsc\n"), ("node_modules/vite", "5", 0o755, b""),
            ("node_modules/vite/bin", "5", 0o755, b""), ("node_modules/vite/bin/vite.js", "0", 0o644, b"vite\n"),
            ("node_modules/vitest", "5", 0o755, b""), ("node_modules/vitest/vitest.mjs", "0", 0o644, b"vitest\n"),
        ]
    elif asset_id == "swag-binary":
        entries = [("swag", "0", 0o755, b"swag\n")]
    else:
        raise RuntimeError("unknown fixture asset")
    if omit_marker:
        entries = [entry for entry in entries if entry[0] not in {"bin/cc", "node_modules", "pkg/mod/cache/download", "swag"}]
    if noncanonical:
        entries[1], entries[2] = entries[2], entries[1]
    return entries


def archive_layout(entries: list[tuple[str, str, int, bytes]]) -> tuple[int, int, str, dict[str, str]]:
    projected: list[tuple[bytes, bytes]] = []
    digests: dict[str, str] = {}
    expanded = 0
    for path, kind, mode, payload in entries:
        normalised = path[:-1] if kind == "5" and path.endswith("/") else path
        if kind == "5":
            projected.append((normalised.encode("utf-8"), f"D\t{normalised}\t0755\t0\t-\n".encode("ascii")))
        else:
            digest = sha256_prefixed(payload)
            digests[normalised] = digest
            expanded += len(payload)
            projected.append((normalised.encode("utf-8"), f"F\t{normalised}\t{mode:04o}\t{len(payload)}\t{digest}\n".encode("ascii")))
    return len(entries), expanded, sha256_prefixed(b"".join(line for _, line in sorted(projected))), digests


def canonical_bundle_sha256(manifest: dict[str, object]) -> str:
    projection = {key: value for key, value in manifest.items() if key != "bundle_sha256"}
    return sha256_prefixed(json.dumps(projection, sort_keys=False, separators=(",", ":"), ensure_ascii=False).encode("utf-8"))


def load_baked_schema() -> dict[str, object]:
    raw = BUNDLED_SCHEMA_PATH.read_bytes()
    schema = json.loads(raw.decode("utf-8"))
    asset = schema.get("$defs", {}).get("asset") if isinstance(schema, dict) else None
    expected = ["asset_id", "targets", "relative_path", "sha256", "bytes", "archive_format", "member_count", "expanded_bytes", "layout_sha256", "license_spdx", "license_evidence_sha256", "provenance_sha256"]
    if not isinstance(asset, dict) or asset.get("required") != expected:
        raise RuntimeError("baked dependency schema does not close A0.1 archive fields")
    return schema


BAKED_SCHEMA: Final[dict[str, object]] = load_baked_schema()


def assert_shared_source_contract() -> None:
    source = (ADMISSION_ROOT / "verify_local_dependency_bundle.py").read_text(encoding="utf-8")
    writer = (ADMISSION_ROOT / "admit_local_dependency_assets.py").read_text(encoding="utf-8")
    required = ("archive_format", "member_count", "expanded_bytes", "layout_sha256", "O_NOFOLLOW", "os.fstat", "tarfile.open(fileobj=", "mode=\"r:\"", "E_ASSET_ARCHIVE_INVALID", "go-race-toolchain.json", "hub.qa.go-race-toolchain-provenance/v1", "cc_target", "cgo_enabled", "capture_observation", "receipt_observations", "post_raw_scan", "post_tarfile", "def _parse_pax(payload: bytes) -> str:", "if not payload:", "if path is None or not path:", "member_path = pending_pax")
    forbidden = ("Path.read_bytes()", "extractall(", "tarfile.open(name=", "tarfile.open(path=")
    if any(token not in source for token in required) or any(token in source for token in forbidden):
        raise RuntimeError("shared archive source contract drifted")
    if "_receipt_asset(archive, staged_targets)" not in writer or "source_observations=observations" in writer or "\"source_observations\": observations" in writer:
        raise RuntimeError("admission writer synthesized source observations")


def write_task_owner(bootstrap_root: pathlib.Path) -> bytes:
    return write_json(bootstrap_root / "owner-marker.v1.json", {"schema": "hub.qa.bootstrap-owner/v1", "task": "g0-assets-runner", "run": "g0-planning-fixture", "owner_nonce": "1" * 32})


def write_handoff_owner(ingress_root: pathlib.Path, manifest_raw: bytes, handoff_id: str = "g0-fixture-handoff") -> bytes:
    return write_json(ingress_root / "handoff-owner.v1.json", {"schema": "hub.qa.g0-offline-ingress-owner/v1", "handoff_id": handoff_id, "source_commit": SOURCE_COMMIT, "dependency_bundle_manifest_sha256": sha256_prefixed(manifest_raw), "owner": "g0-planning-selftest", "released_at_utc": "2026-08-14T00:00:00Z"})


def write_bound_manifest(manifest_path: pathlib.Path, ingress_root: pathlib.Path, manifest: dict[str, object]) -> str:
    required = BAKED_SCHEMA["$defs"]["asset"]["required"]
    if any(list(item.keys()) != required for item in manifest["assets"]):
        raise RuntimeError("fixture asset ordering does not match schema")
    manifest_raw = write_json(manifest_path, manifest)
    (ingress_root / MANIFEST_NAME).parent.mkdir(parents=True, exist_ok=True)
    (ingress_root / MANIFEST_NAME).write_bytes(manifest_raw)
    return sha256_prefixed(write_handoff_owner(ingress_root, manifest_raw))


def build_valid_case(case_name: str) -> tuple[pathlib.Path, pathlib.Path, pathlib.Path, str, str, str, str, pathlib.Path]:
    case_root = FIXTURE_ROOT / case_name
    if case_root.exists():
        raise RuntimeError(f"fixture root already exists: {case_name}")
    ingress_root = case_root / "ingress" / "dependency-bundle"
    controls_root = case_root / "controls"
    bootstrap_root = case_root / "bootstrap"
    assets: list[dict[str, object]] = []
    race_cc_digest = ""
    for asset_id, targets, relative_path in ASSET_SPECS:
        entries = archive_entries(asset_id)
        raw = tar_archive(entries)
        count, expanded, layout, member_digests = archive_layout(entries)
        (ingress_root / relative_path).parent.mkdir(parents=True, exist_ok=True)
        (ingress_root / relative_path).write_bytes(raw)
        license_raw = f"fixture license evidence for {asset_id}\n".encode("ascii")
        (ingress_root / "evidence" / "licenses" / "assets" / f"{asset_id}.txt").parent.mkdir(parents=True, exist_ok=True)
        (ingress_root / "evidence" / "licenses" / "assets" / f"{asset_id}.txt").write_bytes(license_raw)
        if asset_id == "go-race-toolchain":
            race_cc_digest = member_digests["bin/cc"]
        assets.append({"asset_id": asset_id, "targets": list(targets), "relative_path": relative_path, "sha256": sha256_prefixed(raw), "bytes": len(raw), "archive_format": "posix-pax-tar", "member_count": count, "expanded_bytes": expanded, "layout_sha256": layout, "license_spdx": "MIT", "license_evidence_sha256": sha256_prefixed(license_raw), "provenance_sha256": ""})
    base_specs = (("go-backend-unit", ("backend-unit",), "1", "2"), ("go-backend-openapi", ("backend-openapi",), "3", "4"), ("frontend-unit", ("frontend-unit",), "5", "6"), ("bootstrap-helper-python", ("g0-admission-helper", "g0-planning-selftest"), "7", "8"))
    bases: list[dict[str, object]] = []
    for base_id, consumers, image_marker, digest_marker in base_specs:
        license_raw = f"fixture base license evidence for {base_id}\n".encode("ascii")
        provenance_raw = (json.dumps({"base_id": base_id}, separators=(",", ":")) + "\n").encode("utf-8")
        (ingress_root / "evidence" / "licenses" / "bases" / f"{base_id}.txt").parent.mkdir(parents=True, exist_ok=True)
        (ingress_root / "evidence" / "licenses" / "bases" / f"{base_id}.txt").write_bytes(license_raw)
        (ingress_root / "evidence" / "provenance" / "bases" / f"{base_id}.json").parent.mkdir(parents=True, exist_ok=True)
        (ingress_root / "evidence" / "provenance" / "bases" / f"{base_id}.json").write_bytes(provenance_raw)
        bases.append({"base_id": base_id, "consumers": list(consumers), "local_image_id": "sha256:" + image_marker * 64, "manifest_digest": "sha256:" + digest_marker * 64, "license_spdx": "MIT", "license_evidence_sha256": sha256_prefixed(license_raw), "provenance_sha256": sha256_prefixed(provenance_raw)})
    for asset in assets:
        asset_id = asset["asset_id"]
        if asset_id == "go-race-toolchain":
            provenance = {"schema": "hub.qa.go-race-toolchain-provenance/v1", "asset_id": "go-race-toolchain", "go_base_id": "go-backend-unit", "go_base_manifest_digest": bases[0]["manifest_digest"], "go_base_local_image_id": bases[0]["local_image_id"], "cc_path": "bin/cc", "cc_sha256": race_cc_digest, "cc_target": "linux/amd64", "cgo_enabled": True}
        else:
            provenance = {"asset_id": asset_id}
        provenance_raw = (json.dumps(provenance, separators=(",", ":")) + "\n").encode("utf-8")
        path = ingress_root / "evidence" / "provenance" / "assets" / f"{asset_id}.json"
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(provenance_raw)
        asset["provenance_sha256"] = sha256_prefixed(provenance_raw)
    manifest: dict[str, object] = {"schema": "hub.qa.dependency-bundle/v1", "bundle_id": "g0-fixture-valid", "source_commit": SOURCE_COMMIT, "platform": {"os": "linux", "architecture": "amd64", "variant": None}, "assets": assets, "bases": bases}
    manifest["bundle_sha256"] = canonical_bundle_sha256(manifest)
    manifest_path = controls_root / MANIFEST_NAME
    task_owner_raw = write_task_owner(bootstrap_root)
    task_owner_sha = sha256_prefixed(task_owner_raw)
    handoff_owner_sha = write_bound_manifest(manifest_path, ingress_root, manifest)
    receipt_bases = [{"base_id": base["base_id"], "local_image_id": base["local_image_id"], "manifest_digest": base["manifest_digest"]} for base in bases]
    receipt_raw = write_json(bootstrap_root / "local-base-inspect.v1.json", {"schema": "hub.qa.local-base-inspect/v1", "task": "g0-assets-runner", "run": "g0-planning-fixture", "owner_marker_sha256": task_owner_sha, "receipt_nonce": "3" * 32, "bases": receipt_bases})
    return manifest_path, ingress_root, bootstrap_root / "local-base-inspect.v1.json", sha256_prefixed(receipt_raw), "3" * 32, task_owner_sha, handoff_owner_sha, case_root


def refresh_bindings(manifest_path: pathlib.Path, ingress_root: pathlib.Path, receipt_path: pathlib.Path, case_root: pathlib.Path) -> tuple[str, str, str]:
    return sha256_prefixed(receipt_path.read_bytes()), sha256_prefixed((case_root / "bootstrap" / "owner-marker.v1.json").read_bytes()), sha256_prefixed((ingress_root / "handoff-owner.v1.json").read_bytes())


def admission_receipt_path(case_root: pathlib.Path) -> pathlib.Path:
    return case_root / "bootstrap" / ADMISSION_RECEIPT_NAME


def staged_tar_path(case_root: pathlib.Path, target: str, relative_path: str) -> pathlib.Path:
    return case_root / "bootstrap" / "build-context" / target / "bindings" / pathlib.PurePosixPath(relative_path)


def staged_tar_files(case_root: pathlib.Path) -> set[str]:
    root = case_root / "bootstrap" / "build-context"
    if not root.exists():
        return set()
    return {path.relative_to(case_root).as_posix() for path in root.rglob("*") if path.is_file()}


def assert_no_admission_publication(case_root: pathlib.Path, *, preserved_staged: set[str] | None = None) -> None:
    if admission_receipt_path(case_root).exists():
        raise RuntimeError("rejected admission wrote success receipt")
    actual = staged_tar_files(case_root)
    if actual != (preserved_staged or set()):
        raise RuntimeError("rejected admission wrote staged tar")


def assert_admission_receipt(case_root: pathlib.Path, admitted: dict[str, Any]) -> None:
    top_level = (
        "schema", "bundle_id", "source_commit", "manifest_sha256", "assets", "bases",
        "local_base_receipt_sha256", "receipt_nonce", "verdict",
    )
    public_asset = (
        "asset_id", "targets", "relative_path", "sha256", "bytes", "archive_format",
        "member_count", "expanded_bytes", "layout_sha256",
    )
    observation_keys = ("opened", "post_hash", "post_raw_scan", "post_tarfile", "pre_stage", "post_stage")
    identity_keys = ("device", "inode", "mode", "size", "mtime_ns")
    stage_keys = ("target", "relative_tar_path", "bytes", "sha256", "create_new", "fsync", "safe_reopen", "rehash_sha256")
    if tuple(admitted) != top_level:
        raise RuntimeError("admission receipt top-level compatibility contract drifted")
    if admitted["schema"] != "hub.qa.g0-dependency-admission-receipt/v1" or admitted["verdict"] != "ACCEPT":
        raise RuntimeError("admission receipt verdict contract drifted")
    if not admission_receipt_path(case_root).is_file():
        raise RuntimeError("valid admission did not publish its fixed receipt")
    for asset in admitted["assets"]:
        if tuple(asset) != public_asset + ("descriptor_contract", "source_observations", "staged_targets"):
            raise RuntimeError("receipt asset closure drifted")
        if asset["descriptor_contract"] != "linux-held-nofollow-fd/v1":
            raise RuntimeError("receipt descriptor contract drifted")
        observations = asset["source_observations"]
        if tuple(observations) != observation_keys:
            raise RuntimeError("receipt source observations are not closed")
        opened = observations["opened"]
        if tuple(opened) != ("descriptor", "pathname") or tuple(opened["descriptor"]) != identity_keys or tuple(opened["pathname"]) != identity_keys:
            raise RuntimeError("receipt opened identity shape drifted")
        for phase in observation_keys:
            observation = observations[phase]
            if tuple(observation) != ("descriptor", "pathname"):
                raise RuntimeError("receipt source observation phase shape drifted")
            if tuple(observation["descriptor"]) != identity_keys or tuple(observation["pathname"]) != identity_keys:
                raise RuntimeError("receipt source observation identity shape drifted")
            if observation["descriptor"] != opened["descriptor"] or observation["pathname"] != opened["descriptor"]:
                raise RuntimeError("receipt source observation identity drifted")
            if observation["descriptor"]["size"] != asset["bytes"]:
                raise RuntimeError("receipt source observation size drifted")
        expected_targets = list(asset["targets"])
        staged = asset["staged_targets"]
        if [child.get("target") for child in staged] != expected_targets:
            raise RuntimeError("receipt staged target order drifted")
        for child, target in zip(staged, expected_targets, strict=True):
            if tuple(child) != stage_keys:
                raise RuntimeError("receipt staged child closure drifted")
            relative_tar_path = f"bootstrap/build-context/{target}/bindings/{asset['relative_path']}"
            if child != {
                "target": target,
                "relative_tar_path": relative_tar_path,
                "bytes": asset["bytes"],
                "sha256": asset["sha256"],
                "create_new": True,
                "fsync": True,
                "safe_reopen": True,
                "rehash_sha256": asset["sha256"],
            }:
                raise RuntimeError("receipt staged child facts drifted")
            staged_path = staged_tar_path(case_root, target, asset["relative_path"])
            if not staged_path.is_file() or staged_path.stat().st_size != asset["bytes"]:
                raise RuntimeError("receipt staged tar is absent or has wrong length")
            if sha256_prefixed(staged_path.read_bytes()) != asset["sha256"]:
                raise RuntimeError("receipt staged tar has wrong digest")


def rewrite_manifest(manifest_path: pathlib.Path, ingress_root: pathlib.Path, mutate: Callable[[dict[str, Any]], None]) -> None:
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    mutate(manifest)
    manifest["bundle_sha256"] = canonical_bundle_sha256(manifest)
    write_bound_manifest(manifest_path, ingress_root, manifest)


def replace_archive(manifest_path: pathlib.Path, ingress_root: pathlib.Path, asset_index: int, raw: bytes) -> None:
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    relative_path = manifest["assets"][asset_index]["relative_path"]
    (ingress_root / relative_path).write_bytes(raw)
    manifest["assets"][asset_index]["sha256"] = sha256_prefixed(raw)
    manifest["assets"][asset_index]["bytes"] = len(raw)
    manifest["bundle_sha256"] = canonical_bundle_sha256(manifest)
    write_bound_manifest(manifest_path, ingress_root, manifest)


def expect_rejection(case_name: str, expected_code: str, mutate: Callable[[pathlib.Path, pathlib.Path, pathlib.Path, pathlib.Path], Callable[[], None] | None]) -> None:
    manifest_path, ingress_root, receipt_path, receipt_sha, nonce, task_owner_sha, handoff_owner_sha, case_root = build_valid_case(case_name)
    restore = mutate(manifest_path, ingress_root, receipt_path, case_root)
    try:
        receipt_sha, task_owner_sha, handoff_owner_sha = refresh_bindings(manifest_path, ingress_root, receipt_path, case_root)
        for operation in (
            lambda: verify_bundle(manifest_path, ingress_root, receipt_path, receipt_sha, nonce, expected_task_owner_marker_sha256=task_owner_sha, expected_handoff_owner_marker_sha256=handoff_owner_sha),
            lambda: admit_bundle(manifest_path, ingress_root, receipt_path, receipt_sha, nonce, admission_receipt_path(case_root), expected_task_owner_marker_sha256=task_owner_sha, expected_handoff_owner_marker_sha256=handoff_owner_sha),
        ):
            try:
                operation()
            except (BundleVerificationError, AdmissionError) as error:
                if error.code != expected_code:
                    raise RuntimeError(f"{case_name} returned {error.code} instead of {expected_code}") from error
            else:
                raise RuntimeError(f"{case_name} unexpectedly accepted")
        assert_no_admission_publication(case_root)
    finally:
        if restore is not None:
            restore()


def no_change(_: pathlib.Path, __: pathlib.Path, ___: pathlib.Path, ____: pathlib.Path) -> None:
    return None


def assert_valid() -> None:
    manifest_path, ingress_root, receipt_path, receipt_sha, nonce, task_owner_sha, handoff_owner_sha, case_root = build_valid_case("valid-manifest")
    verified = verify_bundle(manifest_path, ingress_root, receipt_path, receipt_sha, nonce, expected_task_owner_marker_sha256=task_owner_sha, expected_handoff_owner_marker_sha256=handoff_owner_sha)
    admitted = admit_bundle(manifest_path, ingress_root, receipt_path, receipt_sha, nonce, admission_receipt_path(case_root), expected_task_owner_marker_sha256=task_owner_sha, expected_handoff_owner_marker_sha256=handoff_owner_sha)
    if tuple(verified) != ("verdict", "manifest_sha256", "assets", "bases", "reason_code") or admitted["verdict"] != "ACCEPT":
        raise RuntimeError("valid archive fixture did not retain public contract")
    assert_admission_receipt(case_root, admitted)


def assert_owner_binding_rejections() -> None:
    manifest_path, ingress_root, receipt_path, receipt_sha, nonce, task_owner_sha, handoff_owner_sha, case_root = build_valid_case("owner-bindings")
    for expected_task, expected_handoff in ((None, handoff_owner_sha), (task_owner_sha, None), (handoff_owner_sha, handoff_owner_sha), ("sha256:" + "0" * 64, handoff_owner_sha)):
        try:
            verify_bundle(manifest_path, ingress_root, receipt_path, receipt_sha, nonce, expected_task_owner_marker_sha256=expected_task, expected_handoff_owner_marker_sha256=expected_handoff)
        except BundleVerificationError as error:
            if error.code != "E_RECEIPT_BINDING_INVALID":
                raise RuntimeError("owner binding rejection drifted") from error
        else:
            raise RuntimeError("missing or substituted owner binding accepted")
        success = admission_receipt_path(case_root)
        try:
            admit_bundle(manifest_path, ingress_root, receipt_path, receipt_sha, nonce, success, expected_task_owner_marker_sha256=expected_task, expected_handoff_owner_marker_sha256=expected_handoff)
        except (BundleVerificationError, AdmissionError) as error:
            if error.code != "E_RECEIPT_BINDING_INVALID":
                raise RuntimeError("owner binding admission rejection drifted") from error
        else:
            raise RuntimeError("admission accepted substituted owner binding")
        assert_no_admission_publication(case_root)


def assert_receipt_ancestor_swap_rejection() -> None:
    manifest_path, ingress_root, receipt_path, receipt_sha, nonce, task_owner_sha, handoff_owner_sha, case_root = build_valid_case("receipt-ancestor-swap")
    original = admission._write_create_new
    displaced_bootstrap = case_root / "bootstrap-displaced"

    def replace_ancestor_before_publication(*args: Any, **kwargs: Any) -> Any:
        original_bootstrap = case_root / "bootstrap"
        replacement_bootstrap = case_root / "bootstrap-replacement"
        replacement_bootstrap.mkdir()
        os.replace(original_bootstrap, displaced_bootstrap)
        os.replace(replacement_bootstrap, original_bootstrap)
        return original(*args, **kwargs)

    admission._write_create_new = replace_ancestor_before_publication
    try:
        try:
            admit_bundle(manifest_path, ingress_root, receipt_path, receipt_sha, nonce, admission_receipt_path(case_root), expected_task_owner_marker_sha256=task_owner_sha, expected_handoff_owner_marker_sha256=handoff_owner_sha)
        except (BundleVerificationError, AdmissionError) as error:
            if error.code != "E_RECEIPT_BINDING_INVALID":
                raise RuntimeError("receipt ancestor replacement rejection drifted") from error
        else:
            raise RuntimeError("receipt ancestor replacement unexpectedly accepted")
        if (displaced_bootstrap / ADMISSION_RECEIPT_NAME).exists():
            raise RuntimeError("receipt ancestor replacement left a redirected success receipt")
        displaced_context = displaced_bootstrap / "build-context"
        if displaced_context.exists() and any(path.is_file() for path in displaced_context.rglob("*")):
            raise RuntimeError("receipt ancestor replacement left redirected staged tar")
        assert_no_admission_publication(case_root)
    finally:
        admission._write_create_new = original


def assert_staging_collision_rejection() -> None:
    manifest_path, ingress_root, receipt_path, receipt_sha, nonce, task_owner_sha, handoff_owner_sha, case_root = build_valid_case("staging-target-collision")
    collision = staged_tar_path(case_root, "backend-unit", "assets/go-module-cache.tar")
    collision.parent.mkdir(parents=True, exist_ok=True)
    collision.write_bytes(b"foreign-stage-collision\n")
    preserved = {collision.relative_to(case_root).as_posix()}
    try:
        admit_bundle(manifest_path, ingress_root, receipt_path, receipt_sha, nonce, admission_receipt_path(case_root), expected_task_owner_marker_sha256=task_owner_sha, expected_handoff_owner_marker_sha256=handoff_owner_sha)
    except AdmissionError as error:
        if error.code != "E_ASSET_ARCHIVE_INVALID":
            raise RuntimeError("staging collision rejection drifted") from error
    else:
        raise RuntimeError("staging collision unexpectedly accepted")
    if collision.read_bytes() != b"foreign-stage-collision\n":
        raise RuntimeError("staging collision was modified")
    assert_no_admission_publication(case_root, preserved_staged=preserved)


def assert_staging_source_identity_rejection() -> None:
    manifest_path, ingress_root, receipt_path, receipt_sha, nonce, task_owner_sha, handoff_owner_sha, case_root = build_valid_case("staging-source-identity-mismatch")
    original = verifier.HeldArchive.capture_observation

    def replace_before_stage(held: Any, phase: str) -> None:
        original(held, phase)
        if phase == "pre_stage":
            replacement = held.path.with_name(held.path.name + ".replacement")
            replacement.write_bytes(held.path.read_bytes())
            os.replace(replacement, held.path)

    verifier.HeldArchive.capture_observation = replace_before_stage
    try:
        try:
            admit_bundle(manifest_path, ingress_root, receipt_path, receipt_sha, nonce, admission_receipt_path(case_root), expected_task_owner_marker_sha256=task_owner_sha, expected_handoff_owner_marker_sha256=handoff_owner_sha)
        except (BundleVerificationError, AdmissionError) as error:
            if error.code != "E_ASSET_ARCHIVE_INVALID":
                raise RuntimeError("source identity rejection drifted") from error
        else:
            raise RuntimeError("source replacement before staging unexpectedly accepted")
        assert_no_admission_publication(case_root)
    finally:
        verifier.HeldArchive.capture_observation = original


def assert_staging_rehash_growth_rejection() -> None:
    manifest_path, ingress_root, receipt_path, receipt_sha, nonce, task_owner_sha, handoff_owner_sha, case_root = build_valid_case("staging-rehash-growth")
    target = staged_tar_path(case_root, "backend-unit", "assets/go-module-cache.tar")
    original_read = admission.os.read
    read_sizes: list[int] = []
    grown = False

    def grow_after_safe_reopen(descriptor: int, size: int) -> bytes:
        nonlocal grown
        if not grown and target.exists() and os.fstat(descriptor).st_ino == target.stat().st_ino:
            expected_bytes = target.stat().st_size
            if size != min(verifier.MAX_STREAM_CHUNK, expected_bytes + 1):
                raise RuntimeError("staged rehash did not request the closed bound plus one byte")
            append_descriptor = os.open(target, os.O_WRONLY | os.O_APPEND | os.O_NOFOLLOW)
            try:
                os.write(append_descriptor, b"x")
            finally:
                os.close(append_descriptor)
            grown = True
        read_sizes.append(size)
        return original_read(descriptor, size)

    admission.os.read = grow_after_safe_reopen
    try:
        try:
            admit_bundle(manifest_path, ingress_root, receipt_path, receipt_sha, nonce, admission_receipt_path(case_root), expected_task_owner_marker_sha256=task_owner_sha, expected_handoff_owner_marker_sha256=handoff_owner_sha)
        except AdmissionError as error:
            if error.code != "E_ASSET_ARCHIVE_INVALID":
                raise RuntimeError("staged rehash growth rejection drifted") from error
        else:
            raise RuntimeError("staged rehash growth unexpectedly accepted")
        if not grown or len(read_sizes) != 1:
            raise RuntimeError("staged rehash growth was not rejected on its first bounded read")
        assert_no_admission_publication(case_root)
    finally:
        admission.os.read = original_read


def _expect_observation_rejection(case_name: str, replacement: Callable[[Callable[[Any, str], None]], Callable[[Any, str], None]]) -> None:
    manifest_path, ingress_root, receipt_path, receipt_sha, nonce, task_owner_sha, handoff_owner_sha, case_root = build_valid_case(case_name)
    original = verifier.HeldArchive.capture_observation
    verifier.HeldArchive.capture_observation = replacement(original)
    try:
        try:
            admit_bundle(manifest_path, ingress_root, receipt_path, receipt_sha, nonce, admission_receipt_path(case_root), expected_task_owner_marker_sha256=task_owner_sha, expected_handoff_owner_marker_sha256=handoff_owner_sha)
        except (BundleVerificationError, AdmissionError) as error:
            if error.code != "E_ASSET_ARCHIVE_INVALID":
                raise RuntimeError(f"{case_name} returned {error.code}") from error
        else:
            raise RuntimeError(f"{case_name} unexpectedly accepted")
        assert_no_admission_publication(case_root)
    finally:
        verifier.HeldArchive.capture_observation = original


def assert_observation_phase_hooks() -> None:
    expected = ("opened", "post_hash", "post_raw_scan", "post_tarfile", "pre_stage", "post_stage")
    manifest_path, ingress_root, receipt_path, receipt_sha, nonce, task_owner_sha, handoff_owner_sha, case_root = build_valid_case("archive-observation-phase-hooks")
    original = verifier.HeldArchive.capture_observation
    captured: dict[int, list[str]] = {}

    def record(held: Any, phase: str) -> None:
        original(held, phase)
        captured.setdefault(id(held), []).append(phase)

    verifier.HeldArchive.capture_observation = record
    try:
        admitted = admit_bundle(manifest_path, ingress_root, receipt_path, receipt_sha, nonce, admission_receipt_path(case_root), expected_task_owner_marker_sha256=task_owner_sha, expected_handoff_owner_marker_sha256=handoff_owner_sha)
    finally:
        verifier.HeldArchive.capture_observation = original
    if len(captured) != len(ASSET_SPECS) or any(tuple(phases) != expected for phases in captured.values()):
        raise RuntimeError("held-descriptor observation phases were not captured exactly once in order")
    assert_admission_receipt(case_root, admitted)

    def missing(capture: Callable[[Any, str], None]) -> Callable[[Any, str], None]:
        def wrapped(held: Any, phase: str) -> None:
            if phase != "post_raw_scan":
                capture(held, phase)
        return wrapped

    def duplicate(capture: Callable[[Any, str], None]) -> Callable[[Any, str], None]:
        def wrapped(held: Any, phase: str) -> None:
            capture(held, phase)
            if phase == "post_hash":
                capture(held, phase)
        return wrapped

    def reordered(capture: Callable[[Any, str], None]) -> Callable[[Any, str], None]:
        def wrapped(held: Any, phase: str) -> None:
            capture(held, "post_raw_scan" if phase == "post_hash" else phase)
        return wrapped

    _expect_observation_rejection("archive-observation-phase-hooks-missing", missing)
    _expect_observation_rejection("archive-observation-phase-hooks-duplicate", duplicate)
    _expect_observation_rejection("archive-observation-phase-hooks-reordered", reordered)


def _archive_mutation(asset_index: int, transformer: Callable[[bytes], bytes]) -> Callable[[pathlib.Path, pathlib.Path, pathlib.Path, pathlib.Path], None]:
    def mutate(manifest_path: pathlib.Path, ingress_root: pathlib.Path, _: pathlib.Path, __: pathlib.Path) -> None:
        manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
        raw = (ingress_root / manifest["assets"][asset_index]["relative_path"]).read_bytes()
        replace_archive(manifest_path, ingress_root, asset_index, transformer(raw))
    return mutate


def _replace_after_raw_hash(manifest_path: pathlib.Path, ingress_root: pathlib.Path, _: pathlib.Path, __: pathlib.Path) -> Callable[[], None]:
    original = verifier._stream_raw_integrity
    def replace(held: Any, expected_bytes: int, expected_sha256: str) -> None:
        original(held, expected_bytes, expected_sha256)
        replacement = held.path.with_name(held.path.name + ".replacement")
        replacement.write_bytes(held.path.read_bytes())
        os.replace(replacement, held.path)
    verifier._stream_raw_integrity = replace
    return lambda: setattr(verifier, "_stream_raw_integrity", original)


def _replace_after_tarfile(manifest_path: pathlib.Path, ingress_root: pathlib.Path, _: pathlib.Path, __: pathlib.Path) -> Callable[[], None]:
    original = verifier._tarfile_view
    def replace(held: Any) -> Any:
        result = original(held)
        replacement = held.path.with_name(held.path.name + ".replacement")
        replacement.write_bytes(held.path.read_bytes())
        os.replace(replacement, held.path)
        return result
    verifier._tarfile_view = replace
    return lambda: setattr(verifier, "_tarfile_view", original)


def _without_nofollow(_: pathlib.Path, __: pathlib.Path, ___: pathlib.Path, ____: pathlib.Path) -> Callable[[], None]:
    original = getattr(verifier.os, "O_NOFOLLOW", None)
    if original is None:
        return lambda: None
    delattr(verifier.os, "O_NOFOLLOW")
    return lambda: setattr(verifier.os, "O_NOFOLLOW", original)


def _over_raw_cap(_: pathlib.Path, __: pathlib.Path, ___: pathlib.Path, ____: pathlib.Path) -> Callable[[], None]:
    original = verifier.MAX_RAW_ARCHIVE_BYTES
    verifier.MAX_RAW_ARCHIVE_BYTES = 1
    return lambda: setattr(verifier, "MAX_RAW_ARCHIVE_BYTES", original)


def _nonregular_asset(manifest_path: pathlib.Path, ingress_root: pathlib.Path, _: pathlib.Path, __: pathlib.Path) -> None:
    path = ingress_root / "assets" / "go-module-cache.tar"
    path.unlink()
    path.mkdir()
    return None


def _missing_asset(_: pathlib.Path, ingress_root: pathlib.Path, __: pathlib.Path, ___: pathlib.Path) -> None:
    (ingress_root / "assets" / "go-module-cache.tar").unlink()
    return None


def _unbound(manifest_path: pathlib.Path, _: pathlib.Path, __: pathlib.Path, ___: pathlib.Path) -> None:
    manifest_path.unlink()
    return None


def _duplicate(manifest_path: pathlib.Path, ingress_root: pathlib.Path, _: pathlib.Path, __: pathlib.Path) -> None:
    rewrite_manifest(manifest_path, ingress_root, lambda manifest: manifest["assets"].append(dict(manifest["assets"][0])))
    return None


def _staging_target_reordered(manifest_path: pathlib.Path, ingress_root: pathlib.Path, _: pathlib.Path, __: pathlib.Path) -> None:
    rewrite_manifest(manifest_path, ingress_root, lambda manifest: manifest["assets"][0].update({"targets": ["backend-openapi", "backend-unit"]}))
    return None


def _staging_target_extra(manifest_path: pathlib.Path, ingress_root: pathlib.Path, _: pathlib.Path, __: pathlib.Path) -> None:
    rewrite_manifest(manifest_path, ingress_root, lambda manifest: manifest["assets"][0].update({"targets": ["backend-unit", "backend-openapi", "frontend-unit"]}))
    return None


def _staging_target_missing(manifest_path: pathlib.Path, ingress_root: pathlib.Path, _: pathlib.Path, __: pathlib.Path) -> None:
    rewrite_manifest(manifest_path, ingress_root, lambda manifest: manifest["assets"][0].update({"targets": ["backend-unit"]}))
    return None


def _wrong_hash(manifest_path: pathlib.Path, ingress_root: pathlib.Path, _: pathlib.Path, __: pathlib.Path) -> None:
    rewrite_manifest(manifest_path, ingress_root, lambda manifest: manifest["assets"][0].update({"sha256": "sha256:" + "0" * 64}))
    return None


def _path_escape(manifest_path: pathlib.Path, ingress_root: pathlib.Path, _: pathlib.Path, __: pathlib.Path) -> None:
    rewrite_manifest(manifest_path, ingress_root, lambda manifest: manifest["assets"][0].update({"relative_path": "../escape.tar"}))
    return None


def _layout_bound(manifest_path: pathlib.Path, ingress_root: pathlib.Path, _: pathlib.Path, __: pathlib.Path) -> None:
    rewrite_manifest(manifest_path, ingress_root, lambda manifest: manifest["assets"][0].update({"member_count": manifest["assets"][0]["member_count"] + 1}))
    return None


def _marker_absence(manifest_path: pathlib.Path, ingress_root: pathlib.Path, _: pathlib.Path, __: pathlib.Path) -> None:
    entries = archive_entries("go-module-cache", omit_marker=True)
    raw = tar_archive(entries)
    count, expanded, layout, _ = archive_layout(entries)
    replace_archive(manifest_path, ingress_root, 0, raw)
    rewrite_manifest(manifest_path, ingress_root, lambda manifest: manifest["assets"][0].update({"member_count": count, "expanded_bytes": expanded, "layout_sha256": layout}))
    return None


def _noncanonical(manifest_path: pathlib.Path, ingress_root: pathlib.Path, _: pathlib.Path, __: pathlib.Path) -> None:
    replace_archive(manifest_path, ingress_root, 1, tar_archive(archive_entries("go-race-toolchain", noncanonical=True)))
    return None


def _forbidden_member(manifest_path: pathlib.Path, ingress_root: pathlib.Path, _: pathlib.Path, __: pathlib.Path) -> None:
    replace_archive(manifest_path, ingress_root, 3, tar_archive([("swag", "2", 0o755, b"")]))
    return None


def _bad_pax(manifest_path: pathlib.Path, ingress_root: pathlib.Path, _: pathlib.Path, __: pathlib.Path) -> None:
    raw = tar_entry("PaxHeader", "x", 0o644, b"10 path=swag\n") + tar_entry("swag", "0", 0o755, b"swag\n") + b"\0" * 1024
    replace_archive(manifest_path, ingress_root, 3, raw)
    return None


def _empty_pax(manifest_path: pathlib.Path, ingress_root: pathlib.Path, _: pathlib.Path, __: pathlib.Path) -> None:
    raw = tar_entry("PaxHeader", "x", 0o644) + tar_entry("swag", "0", 0o755, b"swag\n") + b"\0" * 1024
    replace_archive(manifest_path, ingress_root, 3, raw)
    return None


def _pax_variant(payload: bytes) -> Callable[[pathlib.Path, pathlib.Path, pathlib.Path, pathlib.Path], None]:
    def mutate(manifest_path: pathlib.Path, ingress_root: pathlib.Path, _: pathlib.Path, __: pathlib.Path) -> None:
        raw = tar_entry("PaxHeader", "x", 0o644, payload) + tar_entry("swag", "0", 0o755, b"swag\n") + b"\0" * 1024
        replace_archive(manifest_path, ingress_root, 3, raw)
    return mutate


def _octal_drift(raw: bytes) -> bytes:
    changed = bytearray(raw)
    changed[100] = ord("9")
    changed[148:156] = b" " * 8
    changed[148:156] = f"{sum(changed[:512]):06o}\0 ".encode("ascii")
    return bytes(changed)


def _checksum_drift(raw: bytes) -> bytes:
    changed = bytearray(raw)
    changed[0] ^= 1
    return bytes(changed)


def _tarfile_drift(_: pathlib.Path, __: pathlib.Path, ___: pathlib.Path, ____: pathlib.Path) -> Callable[[], None]:
    original = verifier._tarfile_view
    verifier._tarfile_view = lambda held: tuple()
    return lambda: setattr(verifier, "_tarfile_view", original)


def main() -> int:
    if FIXTURE_ROOT.exists():
        raise RuntimeError("planning fixture root already exists")
    try:
        assert_shared_source_contract()
        assert_valid()
        assert_owner_binding_rejections()
        assert_receipt_ancestor_swap_rejection()
        expect_rejection("duplicate-asset-id", "E_DUPLICATE_ASSET_ID", _duplicate)
        expect_rejection("staging-target-order", "E_MANIFEST_SCHEMA", _staging_target_reordered)
        expect_rejection("staging-target-extra", "E_MANIFEST_SCHEMA", _staging_target_extra)
        expect_rejection("staging-target-missing", "E_MANIFEST_SCHEMA", _staging_target_missing)
        assert_staging_collision_rejection()
        assert_staging_source_identity_rejection()
        assert_staging_rehash_growth_rejection()
        expect_rejection("asset-hash-mismatch", "E_ASSET_SHA256_MISMATCH", _wrong_hash)
        expect_rejection("asset-path-escape", "E_ASSET_PATH_ESCAPE", _path_escape)
        expect_rejection("missing-asset", "E_ASSET_MISSING", _missing_asset)
        expect_rejection("unbound-manifest", "E_MANIFEST_SCHEMA", _unbound)
        expect_rejection("archive-lstat-regular-only", "E_ASSET_ARCHIVE_INVALID", _nonregular_asset)
        expect_rejection("archive-safe-handle-required", "E_ASSET_ARCHIVE_INVALID", _without_nofollow)
        expect_rejection("archive-replaced-between-passes", "E_ASSET_ARCHIVE_INVALID", _replace_after_raw_hash)
        expect_rejection("archive-replaced-before-staging", "E_ASSET_ARCHIVE_INVALID", _replace_after_tarfile)
        expect_rejection("archive-stream-cap-before-tar", "E_ASSET_ARCHIVE_INVALID", _over_raw_cap)
        expect_rejection("archive-compressed", "E_ASSET_ARCHIVE_INVALID", _archive_mutation(0, gzip.compress))
        expect_rejection("archive-truncated-tail", "E_ASSET_ARCHIVE_INVALID", _archive_mutation(0, lambda raw: raw[:-1]))
        expect_rejection("archive-terminal-missing-zero", "E_ASSET_ARCHIVE_INVALID", _archive_mutation(0, lambda raw: raw[:-512]))
        expect_rejection("archive-terminal-eof", "E_ASSET_ARCHIVE_INVALID", _archive_mutation(0, lambda raw: raw + b"\0" * 512))
        expect_rejection("archive-terminal-trailing-byte", "E_ASSET_ARCHIVE_INVALID", _archive_mutation(0, lambda raw: raw + b"x"))
        expect_rejection("archive-checksum-octal", "E_ASSET_ARCHIVE_INVALID", _archive_mutation(0, _checksum_drift))
        expect_rejection("archive-posix-octal-drift", "E_ASSET_ARCHIVE_INVALID", _archive_mutation(0, _octal_drift))
        expect_rejection("archive-forbidden-member", "E_ASSET_ARCHIVE_INVALID", _forbidden_member)
        expect_rejection("archive-noncanonical-order", "E_ASSET_ARCHIVE_INVALID", _noncanonical)
        expect_rejection("archive-pax-closed", "E_ASSET_ARCHIVE_INVALID", _bad_pax)
        expect_rejection("archive-pax-empty", "E_ASSET_ARCHIVE_INVALID", _empty_pax)
        expect_rejection("archive-pax-invalid-utf8", "E_ASSET_ARCHIVE_INVALID", _pax_variant(b"10 path=\xff\n"))
        expect_rejection("archive-pax-prefix-cap", "E_ASSET_ARCHIVE_INVALID", _pax_variant(b"100000 path=swag\n"))
        expect_rejection("archive-pax-duplicate-key", "E_ASSET_ARCHIVE_INVALID", _pax_variant(pax_payload("path", "swag") + pax_payload("path", "swag")))
        expect_rejection("archive-pax-nonpath", "E_ASSET_ARCHIVE_INVALID", _pax_variant(pax_payload("uid", "0")))
        expect_rejection("archive-raw-tarfile-crosscheck", "E_ASSET_ARCHIVE_INVALID", _tarfile_drift)
        expect_rejection("archive-marker-absence", "E_ASSET_ARCHIVE_INVALID", _marker_absence)
        expect_rejection("archive-layout-bounds", "E_ASSET_ARCHIVE_INVALID", _layout_bound)
        assert_observation_phase_hooks()
        print(json.dumps({"target": "g0-planning-selftest", "status": "PASS", "fixtures": list(FIXTURE_CASES)}, separators=(",", ":")))
        return 0
    finally:
        if FIXTURE_ROOT.exists():
            shutil.rmtree(FIXTURE_ROOT)


if __name__ == "__main__":
    raise SystemExit(main())
