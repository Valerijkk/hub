"""Closed, offline verifier for the G0 dependency bundle.

This module is intentionally stdlib-only. It is shared by the ephemeral helper
and the image-baked planning fixture; it has no network, Docker, or host
launcher capability.
"""

from __future__ import annotations

import hashlib
import json
import os
import re
import stat
from pathlib import Path
from typing import Any, Final

SOURCE_COMMIT: Final = "7b5d5501cefab34f9aa5661e4f8c38e71474f9ce"
SCHEMA: Final = "hub.qa.dependency-bundle/v1"
BASE_RECEIPT_SCHEMA: Final = "hub.qa.local-base-inspect/v1"
SHA256_RE: Final = re.compile(r"^sha256:[0-9a-f]{64}$")
RUN_RE: Final = re.compile(r"^[a-z0-9][a-z0-9-]{7,63}$")
NONCE_RE: Final = re.compile(r"^[0-9a-f]{32}$")
REASON_CODES: Final = frozenset({
    "E_MANIFEST_UTF8",
    "E_MANIFEST_SCHEMA",
    "E_DUPLICATE_ASSET_ID",
    "E_ASSET_PATH_ESCAPE",
    "E_ASSET_MISSING",
    "E_ASSET_BYTES_MISMATCH",
    "E_ASSET_SHA256_MISMATCH",
    "E_BUNDLE_SHA256_MISMATCH",
    "E_LICENSE_SPDX_INVALID",
    "E_LICENSE_EVIDENCE_MISSING",
    "E_PROVENANCE_EVIDENCE_MISSING",
    "E_BASE_MISSING",
    "E_BASE_ID_MISMATCH",
    "E_SOURCE_COMMIT_MISMATCH",
    "E_RECEIPT_BINDING_INVALID",
})
SPDX: Final = frozenset({
    "0BSD", "Apache-2.0", "Artistic-2.0", "BSD-2-Clause", "BSD-3-Clause",
    "CC0-1.0", "EPL-2.0", "ISC", "MIT", "MPL-2.0", "PSF-2.0",
    "Python-2.0", "Unicode-DFS-2016", "Zlib",
})
ASSET_LAYOUT: Final = (
    ("go-module-cache", ("backend-unit", "backend-openapi"), "assets/go-module-cache"),
    ("go-race-toolchain", ("backend-unit",), "assets/go-race-toolchain"),
    ("frontend-node-modules", ("frontend-unit",), "assets/frontend-node-modules"),
    ("swag-binary", ("backend-openapi",), "assets/swag-binary"),
)
BASE_LAYOUT: Final = (
    ("go-backend-unit", ("backend-unit",)),
    ("go-backend-openapi", ("backend-openapi",)),
    ("frontend-unit", ("frontend-unit",)),
    ("bootstrap-helper-python", ("g0-admission-helper", "g0-planning-selftest")),
)
_ROOT_KEYS: Final = ("schema", "bundle_id", "source_commit", "platform", "assets", "bases", "bundle_sha256")
_ASSET_KEYS: Final = (
    "asset_id", "targets", "relative_path", "sha256", "bytes", "license_spdx",
    "license_evidence_sha256", "provenance_sha256",
)
_BASE_KEYS: Final = (
    "base_id", "consumers", "local_image_id", "manifest_digest", "license_spdx",
    "license_evidence_sha256", "provenance_sha256",
)
_RECEIPT_KEYS: Final = ("schema", "task", "run", "owner_marker_sha256", "receipt_nonce", "bases")
_RECEIPT_BASE_KEYS: Final = ("base_id", "local_image_id", "manifest_digest")
_OWNER_KEYS: Final = (
    "schema", "handoff_id", "source_commit", "dependency_bundle_manifest_sha256", "owner", "released_at_utc",
)


class BundleVerificationError(ValueError):
    """A closed, plan-defined rejection code."""

    def __init__(self, code: str) -> None:
        if code not in REASON_CODES:
            raise ValueError("non-plan dependency verification reason code")
        super().__init__(code)
        self.code = code


def _reject(code: str) -> None:
    raise BundleVerificationError(code)


def _no_duplicate_object(pairs: list[tuple[str, Any]]) -> dict[str, Any]:
    result: dict[str, Any] = {}
    for key, value in pairs:
        if not isinstance(key, str) or key in result:
            _reject("E_MANIFEST_SCHEMA")
        result[key] = value
    return result


def _no_constant(_: str) -> None:
    _reject("E_MANIFEST_SCHEMA")


def read_strict_json(path: Path, *, failure_code: str) -> tuple[bytes, Any]:
    if failure_code not in REASON_CODES:
        raise ValueError("unclosed dependency JSON failure code")
    try:
        raw = path.read_bytes()
    except OSError:
        _reject(failure_code)
    if raw.startswith(b"\xef\xbb\xbf"):
        _reject("E_MANIFEST_UTF8")
    try:
        value = json.loads(
            raw.decode("utf-8"), object_pairs_hook=_no_duplicate_object,
            parse_constant=_no_constant,
        )
    except UnicodeDecodeError:
        _reject("E_MANIFEST_UTF8")
    except (json.JSONDecodeError, BundleVerificationError):
        _reject(failure_code)
    return raw, value


def sha256_prefixed(raw: bytes) -> str:
    return "sha256:" + hashlib.sha256(raw).hexdigest()


def canonical_bundle_sha256(bundle: dict[str, Any]) -> str:
    """Hash the non-self projection without reordering any received object."""
    if not isinstance(bundle, dict) or tuple(bundle.keys()) != _ROOT_KEYS:
        _reject("E_BUNDLE_SHA256_MISMATCH")
    projection = {key: value for key, value in bundle.items() if key != "bundle_sha256"}
    try:
        raw = json.dumps(
            projection, ensure_ascii=False, sort_keys=False, separators=(",", ":"),
            allow_nan=False,
        ).encode("utf-8")
    except (TypeError, ValueError):
        _reject("E_MANIFEST_SCHEMA")
    return sha256_prefixed(raw)


def _exact_keys(value: Any, expected: tuple[str, ...], code: str = "E_MANIFEST_SCHEMA") -> dict[str, Any]:
    if not isinstance(value, dict) or tuple(value.keys()) != expected:
        _reject(code)
    return value


def _digest(value: Any, code: str = "E_MANIFEST_SCHEMA") -> str:
    if not isinstance(value, str) or SHA256_RE.fullmatch(value) is None:
        _reject(code)
    return value


def _regular_relative(root: Path, relative: str, code: str, *, asset_only: bool = False) -> Path:
    if (
        not isinstance(relative, str)
        or relative.startswith("/")
        or "\\" in relative
        or (asset_only and not relative.startswith("assets/"))
    ):
        _reject(code)
    parts = relative.split("/")
    if any(not part or part in {".", ".."} for part in parts):
        _reject(code)
    root_abs = Path(os.path.abspath(root))
    candidate = Path(os.path.abspath(root_abs.joinpath(*parts)))
    if candidate == root_abs or root_abs not in candidate.parents:
        _reject(code)
    parent = candidate.parent
    while parent != root_abs:
        try:
            parent_mode = parent.lstat().st_mode
        except OSError:
            _reject(code)
        if stat.S_ISLNK(parent_mode) or not stat.S_ISDIR(parent_mode):
            _reject(code)
        parent = parent.parent
    try:
        mode = candidate.lstat().st_mode
    except OSError:
        _reject(code)
    if stat.S_ISLNK(mode) or not stat.S_ISREG(mode):
        _reject(code)
    return candidate


def _verify_evidence(root: Path, relative: str, expected_digest: str, code: str) -> None:
    path = _regular_relative(root, relative, code)
    try:
        actual = sha256_prefixed(path.read_bytes())
    except OSError:
        _reject(code)
    if actual != expected_digest:
        _reject(code)


def _verify_closed_tree(root: Path, expected_files: set[str]) -> None:
    root_abs = Path(os.path.abspath(root))
    try:
        root_mode = root_abs.lstat().st_mode
    except OSError:
        _reject("E_RECEIPT_BINDING_INVALID")
    if stat.S_ISLNK(root_mode) or not stat.S_ISDIR(root_mode):
        _reject("E_RECEIPT_BINDING_INVALID")

    found_files: set[str] = set()
    found_directories: set[str] = set()
    for current, directories, files in os.walk(root_abs, followlinks=False):
        current_path = Path(current)
        for name in directories:
            child = current_path / name
            try:
                mode = child.lstat().st_mode
            except OSError:
                _reject("E_RECEIPT_BINDING_INVALID")
            if stat.S_ISLNK(mode) or not stat.S_ISDIR(mode):
                _reject("E_RECEIPT_BINDING_INVALID")
            found_directories.add(child.relative_to(root_abs).as_posix())
        for name in files:
            child = current_path / name
            try:
                mode = child.lstat().st_mode
            except OSError:
                _reject("E_RECEIPT_BINDING_INVALID")
            if stat.S_ISLNK(mode) or not stat.S_ISREG(mode):
                _reject("E_RECEIPT_BINDING_INVALID")
            found_files.add(child.relative_to(root_abs).as_posix())
    expected_directories: set[str] = set()
    for relative in expected_files:
        parts = relative.split("/")[:-1]
        for index in range(1, len(parts) + 1):
            expected_directories.add("/".join(parts[:index]))
    if found_files != expected_files or found_directories != expected_directories:
        _reject("E_RECEIPT_BINDING_INVALID")


def _verify_platform(value: Any) -> None:
    platform = _exact_keys(value, ("os", "architecture", "variant"))
    if platform != {"os": "linux", "architecture": "amd64", "variant": None}:
        _reject("E_MANIFEST_SCHEMA")


def _validate_expected_owner_bindings(
    task_owner_marker_sha256: object,
    handoff_owner_marker_sha256: object,
) -> tuple[str, str]:
    """Close the distinct task and ingress owner bindings before file reads."""
    task_owner_marker_sha256 = _digest(task_owner_marker_sha256, "E_RECEIPT_BINDING_INVALID")
    handoff_owner_marker_sha256 = _digest(handoff_owner_marker_sha256, "E_RECEIPT_BINDING_INVALID")
    if task_owner_marker_sha256 == handoff_owner_marker_sha256:
        _reject("E_RECEIPT_BINDING_INVALID")
    return task_owner_marker_sha256, handoff_owner_marker_sha256


def _verify_owner(root: Path, manifest_sha: str, expected_handoff_owner_marker_sha256: str) -> str:
    owner_raw, owner = read_strict_json(root / "handoff-owner.v1.json", failure_code="E_RECEIPT_BINDING_INVALID")
    owner = _exact_keys(owner, _OWNER_KEYS, "E_RECEIPT_BINDING_INVALID")
    owner_sha = sha256_prefixed(owner_raw)
    if (
        owner["schema"] != "hub.qa.g0-offline-ingress-owner/v1"
        or not isinstance(owner["handoff_id"], str)
        or RUN_RE.fullmatch(owner["handoff_id"]) is None
        or _digest(owner["dependency_bundle_manifest_sha256"], "E_RECEIPT_BINDING_INVALID") != manifest_sha
        or not isinstance(owner["owner"], str)
        or not isinstance(owner["released_at_utc"], str)
    ):
        _reject("E_RECEIPT_BINDING_INVALID")
    if owner["source_commit"] != SOURCE_COMMIT:
        _reject("E_SOURCE_COMMIT_MISMATCH")
    if expected_handoff_owner_marker_sha256 != owner_sha:
        _reject("E_RECEIPT_BINDING_INVALID")
    return owner_sha


def _verify_assets(bundle: dict[str, Any], root: Path) -> tuple[set[str], list[dict[str, Any]]]:
    assets = bundle["assets"]
    if not isinstance(assets, list):
        _reject("E_MANIFEST_SCHEMA")
    seen_ids: set[str] = set()
    for value in assets:
        if not isinstance(value, dict):
            _reject("E_MANIFEST_SCHEMA")
        supplied_id = value.get("asset_id")
        if not isinstance(supplied_id, str):
            _reject("E_MANIFEST_SCHEMA")
        if supplied_id in seen_ids:
            _reject("E_DUPLICATE_ASSET_ID")
        seen_ids.add(supplied_id)
    if len(assets) != len(ASSET_LAYOUT):
        _reject("E_MANIFEST_SCHEMA")

    expected_files = {"handoff-owner.v1.json", "dependency-bundle.manifest.v1.json"}
    output: list[dict[str, Any]] = []
    for value, expected in zip(assets, ASSET_LAYOUT, strict=True):
        asset = _exact_keys(value, _ASSET_KEYS)
        asset_id, targets, relative_path = expected
        supplied_path = asset["relative_path"]
        if (
            not isinstance(supplied_path, str)
            or not supplied_path.startswith("assets/")
            or supplied_path.startswith("/")
            or "\\" in supplied_path
            or any(not part or part in {".", ".."} for part in supplied_path.split("/"))
        ):
            _reject("E_ASSET_PATH_ESCAPE")
        asset_path = _regular_relative(root, supplied_path, "E_ASSET_MISSING", asset_only=True)
        if (
            asset["asset_id"] != asset_id
            or asset["targets"] != list(targets)
            or asset["relative_path"] != relative_path
            or not isinstance(asset["bytes"], int)
            or isinstance(asset["bytes"], bool)
            or asset["bytes"] < 1
        ):
            _reject("E_MANIFEST_SCHEMA")
        if not isinstance(asset["license_spdx"], str) or asset["license_spdx"] not in SPDX:
            _reject("E_LICENSE_SPDX_INVALID")
        digest = _digest(asset["sha256"])
        license_digest = _digest(asset["license_evidence_sha256"])
        provenance_digest = _digest(asset["provenance_sha256"])
        try:
            raw = asset_path.read_bytes()
        except OSError:
            _reject("E_ASSET_MISSING")
        if len(raw) != asset["bytes"]:
            _reject("E_ASSET_BYTES_MISMATCH")
        if sha256_prefixed(raw) != digest:
            _reject("E_ASSET_SHA256_MISMATCH")
        license_path = f"evidence/licenses/assets/{asset_id}.txt"
        provenance_path = f"evidence/provenance/assets/{asset_id}.json"
        _verify_evidence(root, license_path, license_digest, "E_LICENSE_EVIDENCE_MISSING")
        _verify_evidence(root, provenance_path, provenance_digest, "E_PROVENANCE_EVIDENCE_MISSING")
        expected_files.update({relative_path, license_path, provenance_path})
        output.append({
            "asset_id": asset_id,
            "targets": list(targets),
            "relative_path": relative_path,
            "sha256": digest,
        })
    return expected_files, output


def _verify_base_receipt(
    bundle: dict[str, Any],
    root: Path,
    receipt_path: Path,
    receipt_sha256: str,
    receipt_nonce: str,
    task_owner_marker_sha256: str,
) -> tuple[set[str], list[dict[str, Any]]]:
    if SHA256_RE.fullmatch(receipt_sha256) is None or NONCE_RE.fullmatch(receipt_nonce) is None:
        _reject("E_RECEIPT_BINDING_INVALID")
    raw, receipt = read_strict_json(receipt_path, failure_code="E_RECEIPT_BINDING_INVALID")
    if sha256_prefixed(raw) != receipt_sha256:
        _reject("E_RECEIPT_BINDING_INVALID")
    receipt = _exact_keys(receipt, _RECEIPT_KEYS, "E_RECEIPT_BINDING_INVALID")
    if (
        receipt["schema"] != BASE_RECEIPT_SCHEMA
        or receipt["task"] != "g0-assets-runner"
        or not isinstance(receipt["run"], str)
        or RUN_RE.fullmatch(receipt["run"]) is None
        or _digest(receipt["owner_marker_sha256"], "E_RECEIPT_BINDING_INVALID") != task_owner_marker_sha256
        or receipt["receipt_nonce"] != receipt_nonce
    ):
        _reject("E_RECEIPT_BINDING_INVALID")
    bases = bundle["bases"]
    receipt_bases = receipt["bases"]
    if not isinstance(bases, list) or not isinstance(receipt_bases, list):
        _reject("E_BASE_MISSING")
    if len(bases) != len(BASE_LAYOUT) or len(receipt_bases) != len(BASE_LAYOUT):
        _reject("E_BASE_MISSING")
    expected_files: set[str] = set()
    output: list[dict[str, Any]] = []
    for base_value, receipt_value, expected in zip(bases, receipt_bases, BASE_LAYOUT, strict=True):
        base = _exact_keys(base_value, _BASE_KEYS)
        receipt_base = _exact_keys(receipt_value, _RECEIPT_BASE_KEYS, "E_RECEIPT_BINDING_INVALID")
        base_id, consumers = expected
        if base["base_id"] != base_id or base["consumers"] != list(consumers) or receipt_base["base_id"] != base_id:
            _reject("E_BASE_ID_MISMATCH")
        if not isinstance(base["license_spdx"], str) or base["license_spdx"] not in SPDX:
            _reject("E_LICENSE_SPDX_INVALID")
        local_image_id = _digest(base["local_image_id"], "E_BASE_ID_MISMATCH")
        manifest_digest = _digest(base["manifest_digest"], "E_BASE_ID_MISMATCH")
        if receipt_base["local_image_id"] != local_image_id or receipt_base["manifest_digest"] != manifest_digest:
            _reject("E_BASE_ID_MISMATCH")
        license_digest = _digest(base["license_evidence_sha256"])
        provenance_digest = _digest(base["provenance_sha256"])
        license_path = f"evidence/licenses/bases/{base_id}.txt"
        provenance_path = f"evidence/provenance/bases/{base_id}.json"
        _verify_evidence(root, license_path, license_digest, "E_LICENSE_EVIDENCE_MISSING")
        _verify_evidence(root, provenance_path, provenance_digest, "E_PROVENANCE_EVIDENCE_MISSING")
        expected_files.update({license_path, provenance_path})
        output.append({
            "base_id": base_id,
            "consumers": list(consumers),
            "local_image_id": local_image_id,
            "manifest_digest": manifest_digest,
        })
    return expected_files, output


def _verify_bundle_or_raise(
    manifest_path: Path,
    ingress_root: Path,
    base_receipt_path: Path,
    base_receipt_sha256: str,
    receipt_nonce: str,
    *,
    expected_task_owner_marker_sha256: str,
    expected_handoff_owner_marker_sha256: str,
) -> tuple[dict[str, Any], dict[str, Any]]:
    task_owner_marker_sha256, handoff_owner_marker_sha256 = _validate_expected_owner_bindings(
        expected_task_owner_marker_sha256,
        expected_handoff_owner_marker_sha256,
    )
    manifest_raw, bundle = read_strict_json(manifest_path, failure_code="E_MANIFEST_SCHEMA")
    bundle = _exact_keys(bundle, _ROOT_KEYS)
    if bundle["schema"] != SCHEMA or not isinstance(bundle["bundle_id"], str) or RUN_RE.fullmatch(bundle["bundle_id"]) is None:
        _reject("E_MANIFEST_SCHEMA")
    if bundle["source_commit"] != SOURCE_COMMIT:
        _reject("E_SOURCE_COMMIT_MISMATCH")
    _verify_platform(bundle["platform"])
    if _digest(bundle["bundle_sha256"]) != canonical_bundle_sha256(bundle):
        _reject("E_BUNDLE_SHA256_MISMATCH")
    ingress_manifest = _regular_relative(ingress_root, "dependency-bundle.manifest.v1.json", "E_MANIFEST_SCHEMA")
    try:
        ingress_raw = ingress_manifest.read_bytes()
    except OSError:
        _reject("E_MANIFEST_SCHEMA")
    if ingress_raw != manifest_raw:
        _reject("E_MANIFEST_SCHEMA")
    manifest_sha = sha256_prefixed(manifest_raw)
    _verify_owner(ingress_root, manifest_sha, handoff_owner_marker_sha256)
    expected_files, assets = _verify_assets(bundle, ingress_root)
    base_files, bases = _verify_base_receipt(
        bundle,
        ingress_root,
        base_receipt_path,
        base_receipt_sha256,
        receipt_nonce,
        task_owner_marker_sha256,
    )
    expected_files.update(base_files)
    _verify_closed_tree(ingress_root, expected_files)
    return bundle, {
        "verdict": "ACCEPT",
        "manifest_sha256": manifest_sha,
        "assets": assets,
        "bases": bases,
        "reason_code": None,
    }


def verify_bundle(
    manifest_path: Path,
    ingress_root: Path,
    base_receipt_path: Path,
    base_receipt_sha256: str,
    receipt_nonce: str,
    *,
    expected_task_owner_marker_sha256: str,
    expected_handoff_owner_marker_sha256: str,
) -> dict[str, Any]:
    """Return only the closed public acceptance projection or raise a plan code."""
    _, result = _verify_bundle_or_raise(
        manifest_path,
        ingress_root,
        base_receipt_path,
        base_receipt_sha256,
        receipt_nonce,
        expected_task_owner_marker_sha256=expected_task_owner_marker_sha256,
        expected_handoff_owner_marker_sha256=expected_handoff_owner_marker_sha256,
    )
    return result
