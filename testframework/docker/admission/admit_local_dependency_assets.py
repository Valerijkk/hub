"""Write a receipt only after the shared closed bundle verifier accepts."""

from __future__ import annotations

import json
import os
import sys
from pathlib import Path
from typing import Any

from verify_local_dependency_bundle import BundleVerificationError, _verify_bundle_or_raise


class AdmissionError(BundleVerificationError):
    """Admission failure that remains within the closed plan reason-code set."""


def _write_create_new(path: Path, value: dict[str, Any]) -> None:
    raw = json.dumps(value, ensure_ascii=False, separators=(",", ":"), allow_nan=False).encode("utf-8")
    try:
        descriptor = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
    except (FileExistsError, OSError):
        raise AdmissionError("E_RECEIPT_BINDING_INVALID") from None
    try:
        with os.fdopen(descriptor, "wb", closefd=True) as handle:
            handle.write(raw)
            handle.flush()
            os.fsync(handle.fileno())
    except OSError:
        raise AdmissionError("E_RECEIPT_BINDING_INVALID") from None


def admit_bundle(
    manifest_path: Path,
    ingress_root: Path,
    base_receipt_path: Path,
    base_receipt_sha256: str,
    receipt_nonce: str,
    success_receipt_path: Path,
    *,
    expected_task_owner_marker_sha256: str,
    expected_handoff_owner_marker_sha256: str,
) -> dict[str, Any]:
    """Verify first, then publish one create-new immutable success receipt."""
    bundle, verified = _verify_bundle_or_raise(
        manifest_path,
        ingress_root,
        base_receipt_path,
        base_receipt_sha256,
        receipt_nonce,
        expected_task_owner_marker_sha256=expected_task_owner_marker_sha256,
        expected_handoff_owner_marker_sha256=expected_handoff_owner_marker_sha256,
    )
    receipt = {
        "schema": "hub.qa.g0-dependency-admission-receipt/v1",
        "bundle_id": bundle["bundle_id"],
        "source_commit": bundle["source_commit"],
        "manifest_sha256": verified["manifest_sha256"],
        "assets": verified["assets"],
        "bases": verified["bases"],
        "local_base_receipt_sha256": base_receipt_sha256,
        "receipt_nonce": receipt_nonce,
        "verdict": "ACCEPT",
    }
    _write_create_new(success_receipt_path, receipt)
    return receipt


def _required_environment(name: str) -> str:
    value = os.environ.get(name)
    if value is None or value == "":
        raise AdmissionError("E_RECEIPT_BINDING_INVALID")
    return value


def main() -> int:
    if len(sys.argv) != 1:
        raise AdmissionError("E_MANIFEST_SCHEMA")
    report_root = Path("/work")
    try:
        receipt = admit_bundle(
            Path("/opt/hub/controls/dependency-bundle.manifest.v1.json"),
            report_root / "ingress" / "dependency-bundle",
            report_root / "bootstrap" / "local-base-inspect.v1.json",
            _required_environment("HUB_QA_PRIVATE_LOCAL_BASE_RECEIPT_SHA256"),
            _required_environment("HUB_QA_PRIVATE_LOCAL_BASE_RECEIPT_NONCE"),
            report_root / "bootstrap" / "g0-admission-helper.admission.v1.json",
            expected_task_owner_marker_sha256=_required_environment("HUB_QA_PRIVATE_OWNER_MARKER_SHA256"),
            expected_handoff_owner_marker_sha256=_required_environment("HUB_QA_PRIVATE_OFFLINE_INGRESS_OWNER_MARKER_SHA256"),
        )
    except BundleVerificationError as error:
        sys.stderr.write(error.code + "\n")
        return 2
    sys.stdout.write(json.dumps(receipt, ensure_ascii=False, separators=(",", ":")) + "\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
