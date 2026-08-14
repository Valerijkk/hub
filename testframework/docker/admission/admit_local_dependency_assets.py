"""Admit one closed dependency bundle and stage it from held descriptors only.

Unlike `_verify_bundle_or_raise`, this writer retains the private verified
session so no staged archive can be copied from a later pathname reopen.
"""

from __future__ import annotations

import hashlib
import json
import os
import stat
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Final

from verify_local_dependency_bundle import (
    MAX_RAW_ARCHIVE_BYTES,
    MAX_STREAM_CHUNK,
    AdmittedArchive,
    BundleVerificationError,
    FileIdentity,
    open_verified_bundle_session,
)

ADMISSION_RECEIPT_NAME: Final = "g0-admission-helper.admission.v1.json"
MAX_ADMISSION_RECEIPT_BYTES: Final = 4 * 1024 * 1024


class AdmissionError(BundleVerificationError):
    """Admission failure that remains within the closed plan reason-code set."""


@dataclass(frozen=True)
class _CreatedTarget:
    parent_descriptor: int
    name: str
    identity: FileIdentity


@dataclass(frozen=True)
class _ReceiptDestination:
    report_root: Path
    report_descriptor: int
    report_identity: tuple[int, int, int]
    bootstrap_descriptor: int
    bootstrap_identity: tuple[int, int, int]
    receipt_name: str


def _archive_error() -> None:
    raise AdmissionError("E_ASSET_ARCHIVE_INVALID")


def _regular_identity(value: os.stat_result) -> FileIdentity:
    if not stat.S_ISREG(value.st_mode):
        _archive_error()
    return FileIdentity(value.st_dev, value.st_ino, value.st_mode, value.st_size, value.st_mtime_ns)


def _directory_identity(value: os.stat_result) -> tuple[int, int, int]:
    if stat.S_ISLNK(value.st_mode) or not stat.S_ISDIR(value.st_mode):
        _archive_error()
    return value.st_dev, value.st_ino, value.st_mode


def _write_all(descriptor: int, raw: bytes) -> None:
    offset = 0
    while offset < len(raw):
        try:
            written = os.write(descriptor, raw[offset:])
        except OSError:
            _archive_error()
        if written < 1:
            _archive_error()
        offset += written


def _read_digest(descriptor: int, expected_bytes: int, hard_cap: int) -> tuple[int, str]:
    if (
        not isinstance(expected_bytes, int)
        or isinstance(expected_bytes, bool)
        or not isinstance(hard_cap, int)
        or isinstance(hard_cap, bool)
        or expected_bytes < 0
        or hard_cap < 0
        or expected_bytes > hard_cap
    ):
        _archive_error()
    digest = hashlib.sha256()
    total = 0
    while True:
        remaining = expected_bytes - total
        # Read only one byte beyond the closed receipt bound, so a concurrent
        # growth cannot make this same-FD rehash allocate or consume an
        # unbounded additional chunk before it fails closed.
        read_size = min(MAX_STREAM_CHUNK, remaining + 1)
        try:
            raw = os.read(descriptor, read_size)
        except OSError:
            _archive_error()
        if not raw:
            return total, "sha256:" + digest.hexdigest()
        if len(raw) > remaining:
            _archive_error()
        total += len(raw)
        digest.update(raw)


def _source_stream_copy(archive: AdmittedArchive, target_descriptor: int, expected_bytes: int, expected_sha256: str) -> None:
    try:
        source = archive.held.seek_start()
    except BundleVerificationError:
        _archive_error()
    digest = hashlib.sha256()
    total = 0
    while True:
        try:
            raw = source.read(MAX_STREAM_CHUNK)
        except OSError:
            _archive_error()
        if not raw:
            break
        total += len(raw)
        if total > expected_bytes:
            _archive_error()
        digest.update(raw)
        _write_all(target_descriptor, raw)
    if total != expected_bytes or "sha256:" + digest.hexdigest() != expected_sha256:
        _archive_error()
    try:
        archive.held.recheck()
    except BundleVerificationError:
        _archive_error()


def _directory_flags() -> int:
    nofollow = getattr(os, "O_NOFOLLOW", None)
    directory = getattr(os, "O_DIRECTORY", None)
    if nofollow is None or directory is None:
        _archive_error()
    return os.O_RDONLY | nofollow | directory | getattr(os, "O_CLOEXEC", 0)


def _open_report_root(report_root: Path) -> tuple[int, tuple[int, int, int]]:
    descriptor: int | None = None
    try:
        before = _directory_identity(report_root.lstat())
        descriptor = os.open(report_root, _directory_flags())
        after = _directory_identity(os.fstat(descriptor))
    except (AdmissionError, OSError):
        if descriptor is not None:
            try:
                os.close(descriptor)
            except OSError:
                pass
        _archive_error()
    if before != after:
        try:
            os.close(descriptor)
        except OSError:
            pass
        _archive_error()
    return descriptor, before


def _open_or_create_directory(parent_descriptor: int, name: str) -> int:
    if not name or "/" in name or "\\" in name or name in {".", ".."}:
        _archive_error()
    descriptor: int | None = None
    try:
        try:
            os.mkdir(name, 0o700, dir_fd=parent_descriptor)
        except FileExistsError:
            pass
        named = os.stat(name, dir_fd=parent_descriptor, follow_symlinks=False)
        if stat.S_ISLNK(named.st_mode) or not stat.S_ISDIR(named.st_mode):
            _archive_error()
        descriptor = os.open(name, _directory_flags(), dir_fd=parent_descriptor)
        opened = os.fstat(descriptor)
    except (AdmissionError, OSError):
        if descriptor is not None:
            try:
                os.close(descriptor)
            except OSError:
                pass
        _archive_error()
    if (named.st_dev, named.st_ino, named.st_mode) != (opened.st_dev, opened.st_ino, opened.st_mode):
        try:
            os.close(descriptor)
        except OSError:
            pass
        _archive_error()
    return descriptor


def _destination_parent(bootstrap_descriptor: int, target: str, relative_path: str) -> tuple[int, str]:
    parts = relative_path.split("/")
    if not target or any(not part or part in {".", ".."} for part in parts) or "\\" in relative_path:
        _archive_error()
    parent_parts = ("build-context", target, "bindings", *parts[:-1])
    descriptor: int | None = None
    try:
        descriptor = _open_or_create_directory(bootstrap_descriptor, parent_parts[0])
        for part in parent_parts[1:]:
            child = _open_or_create_directory(descriptor, part)
            os.close(descriptor)
            descriptor = child
    except BaseException:
        try:
            os.close(descriptor)
        except OSError:
            pass
        raise
    return descriptor, parts[-1]


def _safe_unlink_child(parent_descriptor: int, name: str, expected: FileIdentity, *, allow_same_file_content_mutation: bool = False) -> None:
    try:
        current = _regular_identity(os.stat(name, dir_fd=parent_descriptor, follow_symlinks=False))
        if current != expected and (
            not allow_same_file_content_mutation
            or (current.device, current.inode, current.mode) != (expected.device, expected.inode, expected.mode)
        ):
            return
        os.unlink(name, dir_fd=parent_descriptor)
    except (AdmissionError, FileNotFoundError, OSError):
        return


def _stage_archive_for_target(archive: AdmittedArchive, target: str, bootstrap_descriptor: int) -> _CreatedTarget:
    expected_bytes = archive.public["bytes"]
    expected_sha256 = archive.public["sha256"]
    relative_path = archive.public["relative_path"]
    if not isinstance(expected_bytes, int) or isinstance(expected_bytes, bool) or not isinstance(expected_sha256, str) or not isinstance(relative_path, str):
        _archive_error()
    parent_descriptor, name = _destination_parent(bootstrap_descriptor, target, relative_path)
    target_descriptor: int | None = None
    created_identity: FileIdentity | None = None
    keep_parent_descriptor = False
    try:
        nofollow = getattr(os, "O_NOFOLLOW", None)
        if nofollow is None:
            _archive_error()
        target_descriptor = os.open(
            name,
            os.O_WRONLY | os.O_CREAT | os.O_EXCL | nofollow | getattr(os, "O_CLOEXEC", 0),
            0o600,
            dir_fd=parent_descriptor,
        )
        created_identity = _regular_identity(os.fstat(target_descriptor))
        _source_stream_copy(archive, target_descriptor, expected_bytes, expected_sha256)
        os.fsync(target_descriptor)
        staged_identity = _regular_identity(os.fstat(target_descriptor))
        created_identity = staged_identity
        os.close(target_descriptor)
        target_descriptor = None
        named = _regular_identity(os.stat(name, dir_fd=parent_descriptor, follow_symlinks=False))
        if named != staged_identity:
            _archive_error()
        reopened = os.open(name, os.O_RDONLY | nofollow | getattr(os, "O_CLOEXEC", 0), dir_fd=parent_descriptor)
        try:
            reopened_identity = _regular_identity(os.fstat(reopened))
            if reopened_identity != staged_identity:
                _archive_error()
            length, digest = _read_digest(reopened, expected_bytes, MAX_RAW_ARCHIVE_BYTES)
        finally:
            try:
                os.close(reopened)
            except OSError:
                pass
        final_named = _regular_identity(os.stat(name, dir_fd=parent_descriptor, follow_symlinks=False))
        if final_named != staged_identity or length != expected_bytes or digest != expected_sha256:
            _archive_error()
        keep_parent_descriptor = True
        return _CreatedTarget(parent_descriptor, name, staged_identity)
    except BaseException as error:
        if target_descriptor is not None:
            try:
                created_identity = _regular_identity(os.fstat(target_descriptor))
            except (AdmissionError, OSError):
                pass
            try:
                os.close(target_descriptor)
            except OSError:
                pass
        if created_identity is not None:
            _safe_unlink_child(parent_descriptor, name, created_identity, allow_same_file_content_mutation=True)
        if isinstance(error, OSError):
            _archive_error()
        raise
    finally:
        if not keep_parent_descriptor:
            try:
                os.close(parent_descriptor)
            except OSError:
                pass


def _receipt_binding_error() -> None:
    raise AdmissionError("E_RECEIPT_BINDING_INVALID")


def _open_receipt_destination(success_receipt_path: Path) -> _ReceiptDestination:
    if success_receipt_path.name != ADMISSION_RECEIPT_NAME or success_receipt_path.parent.name != "bootstrap":
        _receipt_binding_error()
    report_root = success_receipt_path.parent.parent
    report_descriptor: int | None = None
    bootstrap_descriptor: int | None = None
    try:
        report_descriptor, report_identity = _open_report_root(report_root)
        bootstrap_descriptor = _open_or_create_directory(report_descriptor, "bootstrap")
        bootstrap_identity = _directory_identity(os.fstat(bootstrap_descriptor))
    except (AdmissionError, OSError):
        if bootstrap_descriptor is not None:
            try:
                os.close(bootstrap_descriptor)
            except OSError:
                pass
        if report_descriptor is not None:
            try:
                os.close(report_descriptor)
            except OSError:
                pass
        _receipt_binding_error()
    destination = _ReceiptDestination(
        report_root=report_root,
        report_descriptor=report_descriptor,
        report_identity=report_identity,
        bootstrap_descriptor=bootstrap_descriptor,
        bootstrap_identity=bootstrap_identity,
        receipt_name=ADMISSION_RECEIPT_NAME,
    )
    try:
        _assert_receipt_destination(destination)
    except BaseException:
        _close_receipt_destination(destination)
        raise
    return destination


def _assert_receipt_destination(destination: _ReceiptDestination, expected_receipt: FileIdentity | None = None) -> None:
    try:
        if _directory_identity(destination.report_root.lstat()) != destination.report_identity:
            _receipt_binding_error()
        if _directory_identity(os.fstat(destination.report_descriptor)) != destination.report_identity:
            _receipt_binding_error()
        bootstrap_named = _directory_identity(os.stat("bootstrap", dir_fd=destination.report_descriptor, follow_symlinks=False))
        if bootstrap_named != destination.bootstrap_identity:
            _receipt_binding_error()
        if _directory_identity(os.fstat(destination.bootstrap_descriptor)) != destination.bootstrap_identity:
            _receipt_binding_error()
        if expected_receipt is None:
            try:
                os.stat(destination.receipt_name, dir_fd=destination.bootstrap_descriptor, follow_symlinks=False)
            except FileNotFoundError:
                return
            _receipt_binding_error()
        if _regular_identity(os.stat(destination.receipt_name, dir_fd=destination.bootstrap_descriptor, follow_symlinks=False)) != expected_receipt:
            _receipt_binding_error()
    except (AdmissionError, OSError):
        _receipt_binding_error()


def _close_receipt_destination(destination: _ReceiptDestination) -> None:
    for descriptor in (destination.bootstrap_descriptor, destination.report_descriptor):
        try:
            os.close(descriptor)
        except OSError:
            pass


def _write_create_new(bootstrap_descriptor: int, name: str, value: dict[str, Any]) -> FileIdentity:
    raw = json.dumps(value, ensure_ascii=False, separators=(",", ":"), allow_nan=False).encode("utf-8")
    descriptor: int | None = None
    created_identity: FileIdentity | None = None
    try:
        nofollow = getattr(os, "O_NOFOLLOW", None)
        if nofollow is None or not name or "/" in name or "\\" in name or name in {".", ".."} or len(raw) > MAX_ADMISSION_RECEIPT_BYTES:
            _receipt_binding_error()
        descriptor = os.open(name, os.O_RDWR | os.O_CREAT | os.O_EXCL | nofollow | getattr(os, "O_CLOEXEC", 0), 0o600, dir_fd=bootstrap_descriptor)
        created_identity = _regular_identity(os.fstat(descriptor))
        _write_all(descriptor, raw)
        os.fsync(descriptor)
        if _regular_identity(os.fstat(descriptor)).st_size != len(raw):
            _receipt_binding_error()
        os.lseek(descriptor, 0, os.SEEK_SET)
        written_bytes, written_sha256 = _read_digest(descriptor, len(raw), MAX_ADMISSION_RECEIPT_BYTES)
        if written_bytes != len(raw) or written_sha256 != "sha256:" + hashlib.sha256(raw).hexdigest():
            _receipt_binding_error()
        created_identity = _regular_identity(os.fstat(descriptor))
        os.close(descriptor)
        descriptor = None
        if _regular_identity(os.stat(name, dir_fd=bootstrap_descriptor, follow_symlinks=False)) != created_identity:
            _receipt_binding_error()
        return created_identity
    except (AdmissionError, OSError):
        if descriptor is not None:
            try:
                created_identity = _regular_identity(os.fstat(descriptor))
            except (AdmissionError, OSError):
                pass
            try:
                os.close(descriptor)
            except OSError:
                pass
        if created_identity is not None:
            _safe_unlink_child(bootstrap_descriptor, name, created_identity)
        _receipt_binding_error()


def _receipt_asset(archive: AdmittedArchive, staged_targets: list[dict[str, Any]]) -> dict[str, Any]:
    public = archive.public
    return {
        "asset_id": public["asset_id"],
        "targets": public["targets"],
        "relative_path": public["relative_path"],
        "sha256": public["sha256"],
        "bytes": public["bytes"],
        "archive_format": public["archive_format"],
        "member_count": public["member_count"],
        "expanded_bytes": public["expanded_bytes"],
        "layout_sha256": public["layout_sha256"],
        "descriptor_contract": "linux-held-nofollow-fd/v1",
        "source_observations": archive.held.receipt_observations(),
        "staged_targets": staged_targets,
    }


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
    """Verify, safely stage each held source archive, then publish one receipt."""
    destination = _open_receipt_destination(success_receipt_path)
    created_targets: list[_CreatedTarget] = []
    created_receipt: FileIdentity | None = None
    try:
        with open_verified_bundle_session(
            manifest_path,
            ingress_root,
            base_receipt_path,
            base_receipt_sha256,
            receipt_nonce,
            expected_task_owner_marker_sha256=expected_task_owner_marker_sha256,
            expected_handoff_owner_marker_sha256=expected_handoff_owner_marker_sha256,
        ) as session:
            receipt_assets: list[dict[str, Any]] = []
            for archive in session.archives:
                expected_bytes = archive.public["bytes"]
                if not isinstance(expected_bytes, int) or isinstance(expected_bytes, bool):
                    _archive_error()
                targets = archive.public["targets"]
                if not isinstance(targets, list) or any(not isinstance(target, str) for target in targets):
                    _archive_error()
                staged_targets: list[dict[str, Any]] = []
                _assert_receipt_destination(destination)
                archive.held.capture_observation("pre_stage")
                for target in targets:
                    created = _stage_archive_for_target(archive, target, destination.bootstrap_descriptor)
                    created_targets.append(created)
                    staged_targets.append({
                        "target": target,
                        "relative_tar_path": f"bootstrap/build-context/{target}/bindings/{archive.public['relative_path']}",
                        "bytes": archive.public["bytes"],
                        "sha256": archive.public["sha256"],
                        "create_new": True,
                        "fsync": True,
                        "safe_reopen": True,
                        "rehash_sha256": archive.public["sha256"],
                    })
                archive.held.capture_observation("post_stage")
                receipt_assets.append(_receipt_asset(archive, staged_targets))
            receipt = {
                "schema": "hub.qa.g0-dependency-admission-receipt/v1",
                "bundle_id": session.bundle["bundle_id"],
                "source_commit": session.bundle["source_commit"],
                "manifest_sha256": session.result["manifest_sha256"],
                "assets": receipt_assets,
                "bases": session.result["bases"],
                "local_base_receipt_sha256": base_receipt_sha256,
                "receipt_nonce": receipt_nonce,
                "verdict": "ACCEPT",
            }
            _assert_receipt_destination(destination)
            created_receipt = _write_create_new(destination.bootstrap_descriptor, destination.receipt_name, receipt)
            _assert_receipt_destination(destination, created_receipt)
            return receipt
    except BaseException:
        if created_receipt is not None:
            _safe_unlink_child(destination.bootstrap_descriptor, destination.receipt_name, created_receipt)
        for created in reversed(created_targets):
            _safe_unlink_child(created.parent_descriptor, created.name, created.identity)
        raise
    finally:
        for created in created_targets:
            try:
                os.close(created.parent_descriptor)
            except OSError:
                pass
        _close_receipt_destination(destination)


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
            report_root / "bootstrap" / ADMISSION_RECEIPT_NAME,
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
