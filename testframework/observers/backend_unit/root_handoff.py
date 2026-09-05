"""Root-only provenance handoff for the backend-unit observer lifecycle.

This module is deliberately a control-plane utility.  It does not start,
build, test, stop, or otherwise manage containers.
"""

from __future__ import annotations

import ctypes
import json
import os
import sys
from pathlib import Path
from typing import Any, Sequence

from . import safe_paths
from .lease import (
    SourceLease,
    _sha256_prefixed,
    _validate_canonical_build_input_manifest_bytes,
    _validate_digest_mapping,
    _validate_snapshot,
)


__all__ = (
    "RootHandoffError",
    "canonicalize_root_handoff",
    "parse_root_handoff",
    "validate_root_handoff",
    "read_root_handoff",
    "read_root_handoff_bytes",
    "write_prebuild_handoff",
    "finalize_postgreen_handoff",
    "main",
)


_SCHEMA = "hub.qa.backend-unit-root-handoff/v1"
_TARGET = "backend-unit"
_MAX_HANDOFF_BYTES = 64 * 1024
_REPORTS_RELATIVE = "testframework/reports"
_SOURCE_LEASE_RELATIVE = (
    "testframework/reports/.backend-unit-source-mutation.lease"
)
_PROVISIONAL_RELATIVE = (
    "testframework/reports/.backend-unit-root-handoff.prebuild.v1.json"
)
_HANDOFF_RELATIVE = "testframework/reports/.backend-unit-root-handoff.v1.json"
_HANDOFF_KEYS = (
    "schema",
    "target",
    "prebuild",
    "postgreen",
    "canonical_build_input_manifest_bytes",
    "canonical_build_input_manifest_sha256",
    "fixed_image_ids",
)
_PROVISIONAL_KEYS = (
    "prebuild",
    "canonical_build_input_manifest_bytes",
    "canonical_build_input_manifest_sha256",
)
_EXPECTED_PACKAGE = "observers.backend_unit"
_EXPECTED_SPEC_NAME = "observers.backend_unit.root_handoff"


class RootHandoffError(RuntimeError):
    """The root control-plane handoff is absent, malformed, or unsafe."""


def _validated_repository_root(repo_root: Path) -> Path:
    if not isinstance(repo_root, Path) or not repo_root.is_absolute():
        raise RootHandoffError("Repository root is not an absolute path")
    try:
        canonical = repo_root.resolve(strict=True)
    except OSError as error:
        raise RootHandoffError("Repository root is unavailable") from error
    if canonical != repo_root:
        raise RootHandoffError("Repository root is not canonical")
    return canonical


def _module_repository_root() -> Path:
    try:
        root = Path(__file__).resolve(strict=True).parents[3]
    except OSError as error:
        raise RootHandoffError("Module repository root is unavailable") from error
    return _validated_repository_root(root)


def _fixed_path(repo_root: Path, relative: str) -> Path:
    parts = tuple(relative.split("/"))
    if not parts or any(not part or part in (".", "..") for part in parts):
        raise RootHandoffError("Internal handoff path is malformed")
    return repo_root.joinpath(*parts)


def _validate_source_snapshot(snapshot: object) -> dict[str, Any]:
    try:
        _validate_snapshot(snapshot)  # type: ignore[arg-type]
    except Exception as error:
        raise RootHandoffError("Source snapshot is malformed") from error
    return snapshot  # type: ignore[return-value]


def _manifest_from_text(value: object) -> bytes:
    if type(value) is not str:
        raise RootHandoffError("Canonical build manifest is not a UTF-8 string")
    try:
        manifest = value.encode("utf-8", "strict")
    except UnicodeError as error:
        raise RootHandoffError("Canonical build manifest is not UTF-8") from error
    try:
        return _validate_canonical_build_input_manifest_bytes(manifest)
    except Exception as error:
        raise RootHandoffError("Canonical build manifest is malformed") from error


def _manifest_from_bytes(value: object) -> bytes:
    try:
        return _validate_canonical_build_input_manifest_bytes(value)
    except Exception as error:
        raise RootHandoffError("Canonical build manifest is malformed") from error


def _validate_manifest_binding(
    snapshot: dict[str, Any], manifest: bytes, declared_digest: object
) -> None:
    digest = _sha256_prefixed(manifest)
    if (
        type(declared_digest) is not str
        or declared_digest != digest
        or snapshot["canonical_build_input_manifest_sha256"] != digest
    ):
        raise RootHandoffError("Canonical build manifest hash does not bind source snapshot")


def _validate_fixed_image_ids(value: object) -> dict[str, str]:
    try:
        _validate_digest_mapping(
            value,
            ("python_contract", "backend_unit"),
            "Fixed image IDs",
        )
    except Exception as error:
        raise RootHandoffError("Fixed image IDs are malformed") from error
    return value  # type: ignore[return-value]


def validate_root_handoff(document: object) -> dict[str, Any]:
    """Validate the exact ordered root-handoff object without normalizing it."""

    if type(document) is not dict or tuple(document) != _HANDOFF_KEYS:
        raise RootHandoffError("Root handoff keys are not closed and canonical")
    if document["schema"] != _SCHEMA or document["target"] != _TARGET:
        raise RootHandoffError("Root handoff schema or target is invalid")
    prebuild = _validate_source_snapshot(document["prebuild"])
    postgreen = _validate_source_snapshot(document["postgreen"])
    if prebuild["status"] != "" or postgreen["status"] != "" or prebuild != postgreen:
        raise RootHandoffError("Root handoff source snapshots are not clean and equal")
    manifest = _manifest_from_text(document["canonical_build_input_manifest_bytes"])
    _validate_manifest_binding(
        prebuild,
        manifest,
        document["canonical_build_input_manifest_sha256"],
    )
    _validate_manifest_binding(
        postgreen,
        manifest,
        document["canonical_build_input_manifest_sha256"],
    )
    _validate_fixed_image_ids(document["fixed_image_ids"])
    return document


def _validate_provisional(document: object) -> dict[str, Any]:
    if type(document) is not dict or tuple(document) != _PROVISIONAL_KEYS:
        raise RootHandoffError("Prebuild handoff keys are not closed and canonical")
    prebuild = _validate_source_snapshot(document["prebuild"])
    if prebuild["status"] != "":
        raise RootHandoffError("Prebuild source snapshot is dirty")
    manifest = _manifest_from_text(document["canonical_build_input_manifest_bytes"])
    _validate_manifest_binding(
        prebuild,
        manifest,
        document["canonical_build_input_manifest_sha256"],
    )
    return document


def _canonical_json_bytes(document: dict[str, Any]) -> bytes:
    try:
        payload = (
            json.dumps(
                document,
                ensure_ascii=False,
                allow_nan=False,
                separators=(",", ":"),
                sort_keys=False,
            ).encode("utf-8", "strict")
            + b"\n"
        )
    except (TypeError, ValueError, UnicodeError) as error:
        raise RootHandoffError("Handoff cannot be canonically encoded") from error
    if not payload or len(payload) > _MAX_HANDOFF_BYTES or b"\r" in payload:
        raise RootHandoffError("Canonical handoff exceeds its byte contract")
    return payload


def canonicalize_root_handoff(document: object) -> bytes:
    """Return the exact compact UTF-8/LF serialization for a final handoff."""

    return _canonical_json_bytes(validate_root_handoff(document))


def _canonicalize_provisional(document: object) -> bytes:
    return _canonical_json_bytes(_validate_provisional(document))


def _reject_json_constant(value: str) -> None:
    raise RootHandoffError("Handoff JSON contains a non-finite value")


def _no_duplicate_object(pairs: list[tuple[str, Any]]) -> dict[str, Any]:
    document: dict[str, Any] = {}
    for key, value in pairs:
        if key in document:
            raise RootHandoffError("Handoff JSON contains duplicate keys")
        document[key] = value
    return document


def _parse_json(payload: object) -> dict[str, Any]:
    if (
        type(payload) is not bytes
        or not payload
        or len(payload) > _MAX_HANDOFF_BYTES
        or b"\r" in payload
        or not payload.endswith(b"\n")
    ):
        raise RootHandoffError("Handoff bytes are outside the canonical byte contract")
    try:
        document = json.loads(
            payload.decode("utf-8", "strict"),
            object_pairs_hook=_no_duplicate_object,
            parse_constant=_reject_json_constant,
        )
    except (UnicodeError, ValueError, TypeError) as error:
        if isinstance(error, RootHandoffError):
            raise
        raise RootHandoffError("Handoff JSON is malformed") from error
    if type(document) is not dict:
        raise RootHandoffError("Handoff JSON root is not an object")
    return document


def parse_root_handoff(payload: object) -> dict[str, Any]:
    """Parse only an already-canonical final root handoff."""

    document = validate_root_handoff(_parse_json(payload))
    if canonicalize_root_handoff(document) != payload:
        raise RootHandoffError("Root handoff bytes are not canonical")
    return document


def _parse_provisional(payload: object) -> dict[str, Any]:
    document = _validate_provisional(_parse_json(payload))
    if _canonicalize_provisional(document) != payload:
        raise RootHandoffError("Prebuild handoff bytes are not canonical")
    return document


def _read_relative(repo_root: Path, relative: str) -> bytes:
    try:
        return safe_paths.read_regular_relative(
            repo_root,
            relative,
            max_bytes=_MAX_HANDOFF_BYTES,
        )
    except Exception as error:
        raise RootHandoffError("Handoff file is absent or unstable") from error


def read_root_handoff_bytes() -> bytes:
    """Descriptor-read and validate the exact final handoff bytes."""

    root = _module_repository_root()
    payload = _read_relative(root, _HANDOFF_RELATIVE)
    parse_root_handoff(payload)
    return payload


def read_root_handoff() -> dict[str, Any]:
    """Descriptor-read, parse, and validate the exact final handoff object."""

    return parse_root_handoff(read_root_handoff_bytes())


def _acquire_lease(repo_root: Path) -> SourceLease:
    try:
        return SourceLease.acquire(
            _fixed_path(repo_root, _SOURCE_LEASE_RELATIVE),
            repo_root=repo_root,
        )
    except Exception as error:
        raise RootHandoffError("Source lease could not be acquired") from error


def _release_lease(lease: SourceLease) -> None:
    try:
        lease.release()
    except Exception as error:
        raise RootHandoffError("Source lease could not be released") from error


def _capture_clean_source(lease: SourceLease) -> tuple[dict[str, Any], bytes, str]:
    try:
        snapshot = _validate_source_snapshot(lease.snapshot())
        manifest = _manifest_from_bytes(lease.canonical_build_input_manifest_bytes())
    except RootHandoffError:
        raise
    except Exception as error:
        raise RootHandoffError("Source provenance could not be captured") from error
    digest = _sha256_prefixed(manifest)
    if snapshot["status"] != "" or snapshot["canonical_build_input_manifest_sha256"] != digest:
        raise RootHandoffError("Source provenance is dirty or changed during capture")
    return snapshot, manifest, digest


def _write_new(path: Path, payload: bytes) -> None:
    if not payload or len(payload) > _MAX_HANDOFF_BYTES:
        raise RootHandoffError("Handoff write payload exceeds its byte contract")
    try:
        safe_paths.atomic_write_same_dir(path, payload)
    except Exception as error:
        raise RootHandoffError("Handoff destination already exists or is unsafe") from error


def write_prebuild_handoff() -> dict[str, Any]:
    """Capture and publish the fixed O_EXCL provisional record under a lease."""

    root = _module_repository_root()
    lease = _acquire_lease(root)
    try:
        prebuild, manifest, digest = _capture_clean_source(lease)
        document: dict[str, Any] = {
            "prebuild": prebuild,
            "canonical_build_input_manifest_bytes": manifest.decode("utf-8", "strict"),
            "canonical_build_input_manifest_sha256": digest,
        }
        payload = _canonicalize_provisional(document)
        _write_new(_fixed_path(root, _PROVISIONAL_RELATIVE), payload)
        if _read_relative(root, _PROVISIONAL_RELATIVE) != payload:
            raise RootHandoffError("Prebuild handoff reread did not match written bytes")
        return document
    finally:
        _release_lease(lease)


def _descriptor_identity(snapshot: Any) -> tuple[int, int, int, int, int, int, int]:
    """Return the full opaque identity of one stable regular descriptor."""

    values = safe_paths._require_regular(snapshot)
    try:
        if isinstance(snapshot, dict):
            mode = snapshot["mode"]
            mtime_ns = snapshot["mtime_ns"]
            ctime_ns = snapshot["ctime_ns"]
        else:
            mode = snapshot.st_mode
            mtime_ns = snapshot.st_mtime_ns
            ctime_ns = snapshot.st_ctime_ns
        if any(type(value) is not int for value in (mode, mtime_ns, ctime_ns)):
            raise TypeError("descriptor generation metadata is malformed")
        return (
            int(values["device"]),
            int(values["inode"]),
            mode,
            int(values["nlink"]),
            int(values["size"]),
            mtime_ns,
            ctime_ns,
        )
    except Exception as error:
        raise RootHandoffError("Provisional handoff identity is malformed") from error


def _read_open_descriptor(
    descriptor: int, max_bytes: int
) -> tuple[bytes, tuple[int, int, int, int, int, int, int]]:
    before = safe_paths._fstat(None, descriptor)
    values = safe_paths._require_regular(before)
    before_identity = _descriptor_identity(before)
    size = int(values["size"])
    if size > max_bytes:
        raise RootHandoffError("Provisional handoff exceeds its byte contract")
    try:
        payload = safe_paths._read_descriptor_exact(
            descriptor,
            size,
            ops=None,
            windows=os.name == "nt",
        )
        after = safe_paths._fstat(None, descriptor)
        safe_paths._require_stable(before, after)
        after_identity = _descriptor_identity(after)
    except RootHandoffError:
        raise
    except Exception as error:
        raise RootHandoffError("Provisional handoff changed while opened") from error
    if len(payload) != size or before_identity != after_identity:
        raise RootHandoffError("Provisional handoff changed while opened")
    return payload, after_identity


def _read_provisional(
    repo_root: Path,
) -> tuple[bytes, tuple[int, int, int, int, int, int, int]]:
    """Read the provisional handoff and retain its full descriptor identity."""

    reports = _fixed_path(repo_root, _REPORTS_RELATIVE)
    parent = -1
    descriptor = -1
    ancestors: tuple[int, ...] = ()
    try:
        if os.name == "nt":
            parent, ancestors, device = safe_paths._nt_open_absolute_directory(reports)
            descriptor = safe_paths._nt_open_relative(
                parent, Path(_PROVISIONAL_RELATIVE).name, False
            )
        else:
            if safe_paths.O_NOFOLLOW == 0:
                raise RootHandoffError("POSIX provisional read requires O_NOFOLLOW")
            parent, device = safe_paths._posix_open_directory(reports)
            flags = os.O_RDONLY | safe_paths.O_NOFOLLOW | safe_paths._O_CLOEXEC | safe_paths._O_NONBLOCK
            descriptor = safe_paths._posix_openat(
                parent, Path(_PROVISIONAL_RELATIVE).name, flags
            )
        payload, identity = _read_open_descriptor(descriptor, _MAX_HANDOFF_BYTES)
        if identity[0] != device:
            raise RootHandoffError("Provisional handoff crossed a device boundary")
        return payload, identity
    except RootHandoffError:
        raise
    except Exception as error:
        raise RootHandoffError("Provisional handoff is absent or unsafe") from error
    finally:
        if descriptor >= 0:
            safe_paths._close(None, descriptor)
        if parent >= 0:
            safe_paths._close(None, parent)
        for ancestor in reversed(ancestors):
            if ancestor != parent:
                safe_paths._close(None, ancestor)


def _remove_provisional(
    repo_root: Path,
    expected_payload: bytes,
    expected_identity: tuple[int, int, int, int, int, int, int],
) -> None:
    """Remove only the descriptor-verified provisional file after final reread."""

    reports = _fixed_path(repo_root, _REPORTS_RELATIVE)
    parent = -1
    descriptor = -1
    probe = -1
    ancestors: tuple[int, ...] = ()
    try:
        if os.name == "nt":
            parent, ancestors, device = safe_paths._nt_open_absolute_directory(reports)
            descriptor = safe_paths._nt_open_relative(
                parent,
                Path(_PROVISIONAL_RELATIVE).name,
                False,
                write=True,
            )
            payload, identity = _read_open_descriptor(descriptor, _MAX_HANDOFF_BYTES)
            if identity[0] != device or payload != expected_payload or identity != expected_identity:
                raise RootHandoffError("Provisional handoff identity does not match")
            safe_paths._windows_delete_handle(descriptor)
            safe_paths._close(None, descriptor)
            descriptor = -1
            probe = safe_paths._nt_open_relative(
                parent,
                Path(_PROVISIONAL_RELATIVE).name,
                False,
                tolerate_missing=True,
            )
            if probe >= 0:
                raise RootHandoffError("Provisional handoff removal was not confirmed")
            return

        if safe_paths.O_NOFOLLOW == 0:
            raise RootHandoffError("POSIX provisional removal requires O_NOFOLLOW")
        parent, device = safe_paths._posix_open_directory(reports)
        flags = os.O_RDONLY | safe_paths.O_NOFOLLOW | safe_paths._O_CLOEXEC | safe_paths._O_NONBLOCK
        descriptor = safe_paths._posix_openat(
            parent,
            Path(_PROVISIONAL_RELATIVE).name,
            flags,
        )
        payload, identity = _read_open_descriptor(descriptor, _MAX_HANDOFF_BYTES)
        if identity[0] != device or payload != expected_payload or identity != expected_identity:
            raise RootHandoffError("Provisional handoff identity does not match")
        library = ctypes.CDLL(None, use_errno=True)
        unlink = library.unlinkat
        unlink.argtypes = (ctypes.c_int, ctypes.c_char_p, ctypes.c_int)
        unlink.restype = ctypes.c_int
        name = Path(_PROVISIONAL_RELATIVE).name
        if unlink(parent, os.fsencode(name), 0) != 0:
            raise OSError(ctypes.get_errno(), "unlinkat failed", name)
        if safe_paths._native_destination_exists(parent, name):
            raise RootHandoffError("Provisional handoff removal was not confirmed")
    except RootHandoffError:
        raise
    except Exception as error:
        raise RootHandoffError("Provisional handoff removal is uncertain") from error
    finally:
        if descriptor >= 0:
            safe_paths._close(None, descriptor)
        if probe >= 0:
            safe_paths._close(None, probe)
        if parent >= 0:
            safe_paths._close(None, parent)
        for ancestor in reversed(ancestors):
            if ancestor != parent:
                safe_paths._close(None, ancestor)


def finalize_postgreen_handoff() -> dict[str, Any]:
    """Bind a verified postgreen source capture and fixed images to final handoff."""

    root = _module_repository_root()
    lease = _acquire_lease(root)
    try:
        provisional_payload, provisional_identity = _read_provisional(root)
        provisional = _parse_provisional(provisional_payload)
        postgreen, manifest, digest = _capture_clean_source(lease)
        try:
            lease.verify_unchanged(provisional["prebuild"], postgreen)
        except Exception as error:
            raise RootHandoffError("Prebuild and postgreen source snapshots differ") from error
        if (
            manifest.decode("utf-8", "strict")
            != provisional["canonical_build_input_manifest_bytes"]
            or digest != provisional["canonical_build_input_manifest_sha256"]
        ):
            raise RootHandoffError("Prebuild and postgreen manifests differ")
        try:
            fixed_image_ids = _validate_fixed_image_ids(lease.fixed_image_ids())
        except RootHandoffError:
            raise
        except Exception as error:
            raise RootHandoffError("Fixed image IDs could not be captured") from error
        document: dict[str, Any] = {
            "schema": _SCHEMA,
            "target": _TARGET,
            "prebuild": provisional["prebuild"],
            "postgreen": postgreen,
            "canonical_build_input_manifest_bytes": manifest.decode("utf-8", "strict"),
            "canonical_build_input_manifest_sha256": digest,
            "fixed_image_ids": fixed_image_ids,
        }
        payload = canonicalize_root_handoff(document)
        _write_new(_fixed_path(root, _HANDOFF_RELATIVE), payload)
        if read_root_handoff_bytes() != payload:
            raise RootHandoffError("Final handoff reread did not match written bytes")
        _remove_provisional(root, provisional_payload, provisional_identity)
        return document
    finally:
        _release_lease(lease)


def _canonical_cli_root() -> Path:
    expected = _module_repository_root()
    cwd = Path.cwd()
    try:
        canonical_cwd = cwd.resolve(strict=True)
    except OSError as error:
        raise RootHandoffError("Current working directory is unavailable") from error
    if cwd != expected or canonical_cwd != expected:
        raise RootHandoffError("Current working directory is not the canonical repository root")
    return expected


def _assert_module_invocation() -> None:
    spec_name = getattr(__spec__, "name", None)
    if __package__ != _EXPECTED_PACKAGE or spec_name != _EXPECTED_SPEC_NAME:
        raise RootHandoffError("root handoff requires exact package module invocation")


def main(argv: Sequence[str] | None = None) -> int:
    """Run exactly one root-only provenance phase; never invoke QA operations."""

    _assert_module_invocation()
    arguments = tuple(sys.argv[1:] if argv is None else argv)
    if sys.version_info[:2] != (3, 12):
        raise RootHandoffError("root handoff requires Python 3.12")
    if len(arguments) != 2 or arguments[0] != "--phase" or arguments[1] not in {
        "prebuild",
        "postgreen",
    }:
        raise RootHandoffError("usage: root_handoff --phase prebuild|postgreen")
    _canonical_cli_root()
    if arguments[1] == "prebuild":
        write_prebuild_handoff()
    else:
        finalize_postgreen_handoff()
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except RootHandoffError as error:
        raise SystemExit(str(error)) from error
