"""Fail-closed publication of a completed backend-unit observer packet."""

from __future__ import annotations

import ctypes
import hashlib
import json
import os
from collections.abc import Callable, Mapping
from pathlib import Path
from typing import Any

from . import safe_paths
from .contracts import (
    ReceiptContractError,
    parse_seal,
    validate_container_receipt,
    validate_final_receipt,
    validate_observer_receipt,
)


class PublicationError(RuntimeError):
    """Raised when an observer packet cannot be safely made authoritative."""


class PublicationRenameError(PublicationError):
    """A rename failure with the only accepted proof status for the final path."""

    def __init__(self, message: str, *, final_owned: bool) -> None:
        if type(final_owned) is not bool:
            raise TypeError("final_owned must be a bool")
        super().__init__(message)
        self.final_owned = final_owned


_PARTIAL_SUFFIX = ".partial"
_COMPLETE_MARKER = "CONTAINER_COMPLETE"
_ABORTED_MARKER = "CONTAINER_ABORTED"
_TERMINAL_MARKERS = (_COMPLETE_MARKER, _ABORTED_MARKER)
_CONTAINER_RECEIPT = "container-receipt.json"
_FINAL_RECEIPT = "final-receipt.json"
_OBSERVER_RECEIPT = "observer-receipt.json"
_SEAL = "SEAL"
_MAX_RECEIPT_BYTES = 64 * 1024 * 1024


def _canonical_json_bytes(document: Mapping[str, object]) -> bytes:
    try:
        rendered = json.dumps(
            document,
            allow_nan=False,
            ensure_ascii=False,
            separators=(",", ":"),
            sort_keys=True,
        )
        return rendered.encode("utf-8") + b"\n"
    except Exception as exc:
        raise PublicationError("receipt is not canonical JSON serializable") from exc


def _parse_canonical_json(payload: bytes, label: str) -> dict[str, object]:
    if type(payload) is not bytes or not payload or b"\r" in payload or not payload.endswith(b"\n"):
        raise PublicationError(f"{label} must be nonempty canonical LF JSON")
    try:
        document = json.loads(payload.decode("utf-8"))
    except Exception as exc:
        raise PublicationError(f"{label} is not UTF-8 JSON") from exc
    if type(document) is not dict or _canonical_json_bytes(document) != payload:
        raise PublicationError(f"{label} is not canonical JSON")
    return document


def _digest(payload: bytes) -> str:
    return "sha256:" + hashlib.sha256(payload).hexdigest()


def _final_directory(run_dir: Path) -> Path:
    if not isinstance(run_dir, Path) or not run_dir.is_absolute():
        raise PublicationError("publication directory must be an absolute pathlib.Path")
    name = run_dir.name
    if not name.endswith(_PARTIAL_SUFFIX) or name == _PARTIAL_SUFFIX:
        raise PublicationError("publication directory must end with .partial")
    final = run_dir.with_name(name[: -len(_PARTIAL_SUFFIX)])
    if final.parent != run_dir.parent or final == run_dir:
        raise PublicationError("publication must stay in one directory")
    try:
        safe_paths._output_name(run_dir, os.name == "nt")
        safe_paths._output_name(final, os.name == "nt")
    except Exception as exc:
        raise PublicationError("publication directory name is unsafe") from exc
    return final


def _close(descriptor: int) -> None:
    safe_paths._close(None, descriptor)


def _read_windows_handle(handle: int, size: int) -> bytes:
    try:
        read = ctypes.windll.kernel32.ReadFile
        read.argtypes = (ctypes.c_void_p, ctypes.c_void_p, ctypes.c_uint32, ctypes.POINTER(ctypes.c_uint32), ctypes.c_void_p)
        read.restype = ctypes.c_int
        pieces: list[bytes] = []
        remaining = size
        while remaining:
            length = min(remaining, 1024 * 1024)
            buffer = ctypes.create_string_buffer(length)
            consumed = ctypes.c_uint32()
            if not read(ctypes.c_void_p(handle), buffer, length, ctypes.byref(consumed), None):
                raise OSError(ctypes.get_last_error(), "ReadFile failed")
            if consumed.value == 0:
                raise OSError("ReadFile ended before the stable size")
            pieces.append(buffer.raw[: consumed.value])
            remaining -= int(consumed.value)
        return b"".join(pieces)
    except Exception as exc:
        raise PublicationError("descriptor-anchored Windows receipt read failed") from exc


def _descriptor_identity(snapshot: Any) -> tuple[int, int, int, int, int, int, int]:
    """Return the complete stable identity from one retained regular descriptor."""

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
    except PublicationError:
        raise
    except Exception as exc:
        raise PublicationError("descriptor identity metadata is unavailable") from exc


def _read_regular_relative(
    run_dir: Path, relative: str, *, allow_empty: bool = False
) -> tuple[bytes, tuple[int, int, int, int, int, int, int]]:
    descriptor = -1
    try:
        descriptor = safe_paths.open_regular_relative(run_dir, relative)
        before_snapshot = safe_paths._fstat(None, descriptor)
        before = safe_paths._require_regular(before_snapshot)
        size = int(before["size"])
        if (
            type(allow_empty) is not bool
            or size < 0
            or size > _MAX_RECEIPT_BYTES
            or (size == 0 and not allow_empty)
        ):
            raise PublicationError(f"{relative} has an unsafe receipt size")
        if os.name == "nt":
            payload = _read_windows_handle(descriptor, size)
        else:
            pieces: list[bytes] = []
            remaining = size
            while remaining:
                piece = os.read(descriptor, min(remaining, 1024 * 1024))
                if not piece:
                    raise PublicationError(f"{relative} ended before its stable size")
                pieces.append(piece)
                remaining -= len(piece)
            payload = b"".join(pieces)
        after_snapshot = safe_paths._fstat(None, descriptor)
        safe_paths._require_regular(after_snapshot)
        safe_paths._require_stable(before_snapshot, after_snapshot)
        before_identity = _descriptor_identity(before_snapshot)
        after_identity = _descriptor_identity(after_snapshot)
        if before_identity != after_identity:
            raise PublicationError(
                f"{relative} descriptor generation changed while being read"
            )
        if len(payload) != size:
            raise PublicationError(f"{relative} changed during the descriptor-anchored read")
        return payload, after_identity
    except PublicationError:
        raise
    except Exception as exc:
        raise PublicationError(f"{relative} is missing or unstable") from exc
    finally:
        if descriptor >= 0:
            _close(descriptor)


def _assert_empty_regular_marker(run_dir: Path) -> None:
    descriptor = -1
    try:
        descriptor = safe_paths.open_regular_relative(run_dir, _COMPLETE_MARKER)
        values = safe_paths._require_regular(safe_paths._fstat(None, descriptor))
        if int(values["size"]) != 0:
            raise PublicationError("CONTAINER_COMPLETE must be an empty terminal marker")
    except PublicationError:
        raise
    except Exception as exc:
        raise PublicationError("CONTAINER_COMPLETE is missing or unstable") from exc
    finally:
        if descriptor >= 0:
            _close(descriptor)


def _unlink_regular(path: Path) -> bool:
    outer_fd = -1
    parent_fd = -1
    ancestors: tuple[int, ...] = ()
    descriptor = -1
    probe = -1
    try:
        if os.name == "nt":
            outer_fd, ancestors, outer_device = safe_paths._nt_open_absolute_directory(path.parent.parent)
            parent_fd = safe_paths._nt_open_relative(outer_fd, path.parent.name, True, tolerate_missing=True)
            if parent_fd < 0:
                return True
            parent_values = safe_paths._require_directory(safe_paths._fstat(None, parent_fd))
            if int(parent_values["device"]) != outer_device:
                raise PublicationError("SEAL parent crossed a device boundary")
            descriptor = safe_paths._nt_open_relative(parent_fd, path.name, False, write=True, tolerate_missing=True)
            if descriptor < 0:
                return True
            safe_paths._require_regular(safe_paths._fstat(None, descriptor))
            safe_paths._windows_delete_handle(descriptor)
            _close(descriptor)
            descriptor = -1
            probe = safe_paths._nt_open_relative(parent_fd, path.name, False, tolerate_missing=True)
            if probe >= 0:
                raise PublicationError("SEAL removal was not confirmed")
            return True

        outer_fd, outer_device = safe_paths._posix_open_directory(path.parent.parent)
        if not safe_paths._native_destination_exists(outer_fd, path.parent.name):
            return True
        directory_flags = os.O_RDONLY | safe_paths._O_DIRECTORY | safe_paths.O_NOFOLLOW | safe_paths._O_CLOEXEC
        parent_fd = safe_paths._posix_openat(outer_fd, path.parent.name, directory_flags)
        parent_values = safe_paths._require_directory(safe_paths._fstat(None, parent_fd))
        if int(parent_values["device"]) != outer_device:
            raise PublicationError("SEAL parent crossed a device boundary")
        if not safe_paths._native_destination_exists(parent_fd, path.name):
            return True
        flags = os.O_RDONLY | safe_paths.O_NOFOLLOW | safe_paths._O_CLOEXEC | safe_paths._O_NONBLOCK
        descriptor = safe_paths._posix_openat(parent_fd, path.name, flags)
        before = safe_paths._require_regular(safe_paths._fstat(None, descriptor))
        safe_paths._require_stable(before, safe_paths._fstat(None, descriptor))
        library = ctypes.CDLL(None, use_errno=True)
        unlink = library.unlinkat
        unlink.argtypes = (ctypes.c_int, ctypes.c_char_p, ctypes.c_int)
        unlink.restype = ctypes.c_int
        if unlink(parent_fd, os.fsencode(path.name), 0) != 0:
            raise OSError(ctypes.get_errno(), "unlinkat failed", path.name)
        if safe_paths._native_destination_exists(parent_fd, path.name):
            raise PublicationError("SEAL removal was not confirmed")
        return True
    except PublicationError:
        raise
    except Exception as exc:
        raise PublicationError("SEAL removal is uncertain") from exc
    finally:
        if descriptor >= 0:
            _close(descriptor)
        if probe >= 0:
            _close(probe)
        if os.name == "nt":
            if parent_fd >= 0:
                _close(parent_fd)
            for ancestor in reversed(ancestors):
                _close(ancestor)
        else:
            if parent_fd >= 0:
                _close(parent_fd)
            if outer_fd >= 0:
                _close(outer_fd)


class _SafePublicationWriter:
    """Production seam restricted to descriptor-anchored safe-path operations."""

    def list_terminal_markers(self, run_dir: Path) -> tuple[str, ...]:
        try:
            names = safe_paths.list_directory_relative(run_dir.parent, run_dir.name)
            return tuple(marker for marker in _TERMINAL_MARKERS if marker in names)
        except Exception as exc:
            raise PublicationError("terminal marker directory state is unavailable") from exc

    def assert_terminal_and_stable(self, run_dir: Path) -> None:
        _read_regular_relative(run_dir, _CONTAINER_RECEIPT)

    def assert_absent(self, final_dir: Path) -> None:
        parent_fd = -1
        ancestors: tuple[int, ...] = ()
        existing = -1
        try:
            if os.name == "nt":
                parent_fd, ancestors, _ = safe_paths._nt_open_absolute_directory(final_dir.parent)
                existing = safe_paths._nt_open_relative(parent_fd, final_dir.name, True, tolerate_missing=True)
                if existing >= 0:
                    raise PublicationError("final publication directory already exists")
                return
            parent_fd, _ = safe_paths._posix_open_directory(final_dir.parent)
            if safe_paths._native_destination_exists(parent_fd, final_dir.name):
                raise PublicationError("final publication directory already exists")
        except PublicationError:
            raise
        except Exception as exc:
            raise PublicationError("cannot establish final publication absence") from exc
        finally:
            if existing >= 0:
                _close(existing)
            if os.name == "nt":
                for ancestor in reversed(ancestors):
                    _close(ancestor)
            elif parent_fd >= 0:
                _close(parent_fd)

    def read(
        self, run_dir: Path, relative: str
    ) -> tuple[bytes, tuple[int, int, int, int, int, int, int]]:
        return _read_regular_relative(
            run_dir,
            relative,
            allow_empty=relative in _TERMINAL_MARKERS,
        )

    def write(self, path: Path, payload: bytes) -> None:
        try:
            safe_paths.atomic_write_same_dir(path, payload)
        except Exception as exc:
            raise PublicationError(f"safe write failed for {path.name}") from exc

    def remove(self, path: Path) -> bool:
        return _unlink_regular(path)

    def rename_same_directory(self, partial: Path, final: Path) -> None:
        if partial.parent != final.parent:
            raise PublicationError("cross-directory publication is denied")
        parent_fd = -1
        ancestors: tuple[int, ...] = ()
        source = -1
        existing = -1
        renamed = -1
        committed = False
        try:
            if os.name == "nt":
                parent_fd, ancestors, _ = safe_paths._nt_open_absolute_directory(partial.parent)
                existing = safe_paths._nt_open_relative(parent_fd, final.name, True, tolerate_missing=True)
                if existing >= 0:
                    raise PublicationError("final publication directory already exists")
                source = safe_paths._nt_open_relative(parent_fd, partial.name, True, write=True)
                before = safe_paths._require_directory(safe_paths._fstat(None, source))
                safe_paths._windows_rename_relative(source, parent_fd, final.name)
                renamed = safe_paths._nt_open_relative(parent_fd, final.name, True)
                after = safe_paths._require_directory(safe_paths._fstat(None, renamed))
                if before["device"] != after["device"] or before["inode"] != after["inode"]:
                    raise PublicationError("final publication directory identity changed")
                committed = True
                return

            parent_fd, parent_device = safe_paths._posix_open_directory(partial.parent)
            if safe_paths._native_destination_exists(parent_fd, final.name):
                raise PublicationError("final publication directory already exists")
            flags = os.O_RDONLY | safe_paths._O_DIRECTORY | safe_paths.O_NOFOLLOW | safe_paths._O_CLOEXEC
            source = safe_paths._posix_openat(parent_fd, partial.name, flags)
            before = safe_paths._require_directory(safe_paths._fstat(None, source))
            if int(before["device"]) != parent_device:
                raise PublicationError("partial publication directory crossed a device boundary")
            safe_paths._native_renameat(parent_fd, partial.name, parent_fd, final.name)
            renamed = safe_paths._posix_openat(parent_fd, final.name, flags)
            after = safe_paths._require_directory(safe_paths._fstat(None, renamed))
            if int(after["device"]) != parent_device or before["inode"] != after["inode"]:
                raise PublicationError("partial publication directory identity changed")
            committed = True
        except PublicationRenameError:
            raise
        except Exception as exc:
            raise PublicationRenameError(
                "safe final publication rename failed",
                final_owned=committed,
            ) from exc
        finally:
            if source >= 0:
                _close(source)
            if renamed >= 0:
                _close(renamed)
            if existing >= 0:
                _close(existing)
            if os.name == "nt":
                for ancestor in reversed(ancestors):
                    _close(ancestor)
            elif parent_fd >= 0:
                _close(parent_fd)


def _validate_semantics(final: Mapping[str, object], observer: Mapping[str, object]) -> None:
    try:
        validate_final_receipt(final)
        validate_observer_receipt(observer)
        if observer.get("final_receipt") != final:
            raise PublicationError("observer receipt must embed the exact final receipt")
        if observer.get("bound_backend_image_id") != final.get("backend_image_id"):
            raise PublicationError("final receipt backend image is not bound to the observer")
        fixed_images = observer.get("fixed_image_ids")
        if not isinstance(fixed_images, Mapping) or fixed_images.get("backend_unit") != final.get("backend_image_id"):
            raise PublicationError("final receipt backend image is not the fixed backend image")
    except PublicationError:
        raise
    except Exception as exc:
        raise PublicationError("receipt semantics are not sealable") from exc


def _inventory_entry(observer: Mapping[str, object], path: str, payload: bytes) -> None:
    records = observer.get("artifact_inventory")
    if type(records) is not list:
        raise PublicationError("observer artifact inventory is malformed")
    matches = [record for record in records if isinstance(record, Mapping) and record.get("path") == path]
    if len(matches) != 1:
        raise PublicationError(f"observer artifact inventory lacks one {path} row")
    record = matches[0]
    if record.get("sha256") != _digest(payload) or type(record.get("size")) is not int or record["size"] != len(payload):
        raise PublicationError(f"observer artifact inventory does not bind exact {path} bytes")


def _validate_container_bytes(
    final: Mapping[str, object], observer: Mapping[str, object], payload: bytes
) -> dict[str, object]:
    document = _parse_canonical_json(payload, _CONTAINER_RECEIPT)
    try:
        validate_container_receipt(document)
    except Exception as exc:
        raise PublicationError("container receipt semantics are not sealable") from exc
    for field in ("run_id", "owner_nonce"):
        if document.get(field) != final.get(field) or document.get(field) != observer.get(field):
            raise PublicationError(f"container receipt {field} does not bind final and observer identity")
    hostname = document.get("container_hostname")
    container_id = final.get("container_id")
    if (
        type(hostname) is not str
        or hostname != observer.get("container_hostname")
        or container_id != observer.get("container_id")
        or type(container_id) is not str
        or not container_id.startswith(hostname)
    ):
        raise PublicationError("container receipt hostname does not bind the full observer container identity")
    digest = _digest(payload)
    if final.get("container_receipt_sha256") != digest or observer.get("container_receipt_sha256") != digest:
        raise PublicationError("container receipt bytes do not bind final and observer receipts")
    _inventory_entry(observer, _CONTAINER_RECEIPT, payload)
    return document


def _seal_bytes(final: Mapping[str, object], observer_payload: bytes) -> bytes:
    try:
        run_id = final["run_id"]
        owner_nonce = final["owner_nonce"]
        if type(run_id) is not str or type(owner_nonce) is not str:
            raise TypeError("run identity is not a string")
        payload = (
            b"hub.qa.backend-unit-seal/v1\n"
            + b"run_id="
            + run_id.encode("ascii")
            + b"\nowner_nonce="
            + owner_nonce.encode("ascii")
            + b"\nobserver_receipt_sha256=sha256:"
            + hashlib.sha256(observer_payload).hexdigest().encode("ascii")
            + b"\n"
        )
        parsed = parse_seal(payload)
        if parsed["run_id"] != run_id or parsed["owner_nonce"] != owner_nonce:
            raise PublicationError("SEAL identity does not match final receipt")
        return payload
    except PublicationError:
        raise
    except Exception as exc:
        raise PublicationError("SEAL cannot be formed from receipt identity") from exc


def _read_payload(
    reader: Callable[
        [Path, str], tuple[bytes, tuple[int, int, int, int, int, int, int]]
    ],
    run_dir: Path,
    relative: str,
    *,
    expected_identity: tuple[int, int, int, int, int, int, int] | None = None,
) -> tuple[bytes, tuple[int, int, int, int, int, int, int]]:
    try:
        result = reader(run_dir, relative)
    except Exception as exc:
        raise PublicationError(f"cannot read {relative} through the writer seam") from exc
    if type(result) is not tuple or len(result) != 2:
        raise PublicationError(f"{relative} reader did not return payload and identity")
    payload, identity = result
    if type(payload) is not bytes:
        raise PublicationError(f"{relative} reader did not return exact bytes")
    if (
        type(identity) is not tuple
        or len(identity) != 7
        or any(type(value) is not int for value in identity)
    ):
        raise PublicationError(
            f"{relative} reader did not return a stable descriptor identity"
        )
    if expected_identity is not None and identity != expected_identity:
        raise PublicationError(f"{relative} identity changed across publication reads")
    return payload, identity


def _read_exact(
    reader: Callable[
        [Path, str], tuple[bytes, tuple[int, int, int, int, int, int, int]]
    ],
    run_dir: Path,
    relative: str,
    expected: bytes,
    *,
    expected_identity: tuple[int, int, int, int, int, int, int] | None = None,
) -> tuple[bytes, tuple[int, int, int, int, int, int, int]]:
    actual, identity = _read_payload(
        reader,
        run_dir,
        relative,
        expected_identity=expected_identity,
    )
    if actual != expected:
        raise PublicationError(f"{relative} was not durably written as exact bytes")
    return actual, identity


def _expected_terminal_marker(container: Mapping[str, object]) -> str:
    state = container.get("state")
    if state == _COMPLETE_MARKER:
        return _COMPLETE_MARKER
    if state == _ABORTED_MARKER:
        return _ABORTED_MARKER
    raise PublicationError("container receipt state does not select a terminal marker")


def _assert_terminal_marker_set(
    lister: Callable[[Path], tuple[str, ...]], run_dir: Path, expected: str
) -> None:
    try:
        present = lister(run_dir)
    except Exception as exc:
        raise PublicationError("terminal marker directory state is unavailable") from exc
    if (
        type(present) is not tuple
        or any(type(marker) is not str for marker in present)
        or len(present) != len(set(present))
        or any(marker not in _TERMINAL_MARKERS for marker in present)
        or set(present) != {expected}
    ):
        raise PublicationError("container terminal markers do not select exactly one expected marker")


def _assert_empty_terminal_marker(
    reader: Callable[
        [Path, str], tuple[bytes, tuple[int, int, int, int, int, int, int]]
    ],
    lister: Callable[[Path], tuple[str, ...]],
    run_dir: Path,
    expected: str,
    *,
    expected_identity: tuple[int, int, int, int, int, int, int] | None = None,
) -> tuple[int, int, int, int, int, int, int]:
    _assert_terminal_marker_set(lister, run_dir, expected)
    payload, identity = _read_payload(
        reader,
        run_dir,
        expected,
        expected_identity=expected_identity,
    )
    if payload != b"":
        raise PublicationError(f"{expected} must be an empty terminal marker")
    _assert_terminal_marker_set(lister, run_dir, expected)
    return identity


def _normalize_prewrite(call: Callable[[], None]) -> None:
    try:
        call()
    except PublicationError:
        raise
    except Exception as exc:
        raise PublicationError("publication precondition failed") from exc


def _normalize_write(call: Callable[[], None]) -> None:
    try:
        call()
    except PublicationError:
        raise
    except Exception as exc:
        raise PublicationError("publication write or rename failed") from exc


def _rename_with_outcome(rename: Callable[[Path, Path], None], run_dir: Path, final_dir: Path) -> None:
    try:
        rename(run_dir, final_dir)
    except PublicationRenameError:
        raise
    except Exception as exc:
        raise PublicationRenameError("publication writer rename outcome is unproven", final_owned=False) from exc


def _invalidate_seal(
    remove: Callable[[Path], object],
    run_dir: Path,
    final_dir: Path,
    *,
    final_owned: bool,
) -> None:
    try:
        directories = (run_dir, final_dir) if final_owned else (run_dir,)
        for directory in directories:
            if remove(directory / _SEAL) is not True:
                raise PublicationError("SEAL cleanup did not confirm absence")
    except Exception as exc:
        raise PublicationError("SEAL cleanup is uncertain; partial packet is non-authoritative") from exc


def publish(
    run_dir: Path,
    final: dict[str, object],
    observer: dict[str, object],
    *,
    validate: Callable[[dict[str, object], dict[str, object]], None] | None = None,
    writer: Any | None = None,
) -> Path:
    """Publish exact receipt bytes only after all semantic and artifact bindings hold."""

    final_dir = _final_directory(run_dir)
    if type(final) is not dict or type(observer) is not dict:
        raise PublicationError("publication receipts must be plain dictionaries")
    active_writer = _SafePublicationWriter() if writer is None else writer
    instrumentation = (lambda _final, _observer: None) if validate is None else validate
    if not callable(instrumentation):
        raise PublicationError("publication validator is not callable")
    try:
        assert_terminal = active_writer.assert_terminal_and_stable
        list_terminal_markers = active_writer.list_terminal_markers
        assert_absent = active_writer.assert_absent
        read = active_writer.read
        write = active_writer.write
        remove = active_writer.remove
        rename = active_writer.rename_same_directory
    except Exception as exc:
        raise PublicationError("publication writer seam is incomplete") from exc
    if not all(callable(call) for call in (assert_terminal, list_terminal_markers, assert_absent, read, write, remove, rename)):
        raise PublicationError("publication writer seam is incomplete")

    _normalize_prewrite(lambda: assert_terminal(run_dir))
    _normalize_prewrite(lambda: assert_absent(final_dir))
    _normalize_prewrite(lambda: _validate_semantics(final, observer))
    container_payload, container_identity = _read_payload(
        read, run_dir, _CONTAINER_RECEIPT
    )
    container = _validate_container_bytes(final, observer, container_payload)
    expected_marker = _expected_terminal_marker(container)
    marker_identity = _assert_empty_terminal_marker(
        read,
        list_terminal_markers,
        run_dir,
        expected_marker,
    )
    _normalize_prewrite(lambda: instrumentation(final, observer))
    _normalize_prewrite(lambda: _validate_semantics(final, observer))
    container_payload, _ = _read_payload(
        read,
        run_dir,
        _CONTAINER_RECEIPT,
        expected_identity=container_identity,
    )
    container = _validate_container_bytes(final, observer, container_payload)
    if _expected_terminal_marker(container) != expected_marker:
        raise PublicationError("container receipt terminal state changed before publication")
    marker_identity = _assert_empty_terminal_marker(
        read,
        list_terminal_markers,
        run_dir,
        expected_marker,
        expected_identity=marker_identity,
    )
    final_payload = _canonical_json_bytes(final)
    _inventory_entry(observer, _FINAL_RECEIPT, final_payload)
    observer_payload = _canonical_json_bytes(observer)
    seal_payload = _seal_bytes(final, observer_payload)

    seal_attempted = False
    final_owned = False
    try:
        _normalize_write(lambda: write(run_dir / _FINAL_RECEIPT, final_payload))
        _, final_identity = _read_exact(read, run_dir, _FINAL_RECEIPT, final_payload)
        _normalize_write(lambda: write(run_dir / _OBSERVER_RECEIPT, observer_payload))
        _, observer_identity = _read_exact(
            read, run_dir, _OBSERVER_RECEIPT, observer_payload
        )
        seal_attempted = True
        _normalize_write(lambda: write(run_dir / _SEAL, seal_payload))
        written_seal, seal_identity = _read_exact(read, run_dir, _SEAL, seal_payload)
        written_final_payload, _ = _read_exact(
            read,
            run_dir,
            _FINAL_RECEIPT,
            final_payload,
            expected_identity=final_identity,
        )
        written_final = _parse_canonical_json(written_final_payload, _FINAL_RECEIPT)
        written_observer_payload, _ = _read_exact(
            read,
            run_dir,
            _OBSERVER_RECEIPT,
            observer_payload,
            expected_identity=observer_identity,
        )
        written_observer = _parse_canonical_json(written_observer_payload, _OBSERVER_RECEIPT)
        _validate_semantics(written_final, written_observer)
        _validate_container_bytes(
            written_final,
            written_observer,
            _read_exact(
                read,
                run_dir,
                _CONTAINER_RECEIPT,
                container_payload,
                expected_identity=container_identity,
            )[0],
        )
        _inventory_entry(written_observer, _FINAL_RECEIPT, final_payload)
        written_seal, _ = _read_exact(
            read,
            run_dir,
            _SEAL,
            written_seal,
            expected_identity=seal_identity,
        )
        parsed_seal = parse_seal(written_seal)
        if (
            parsed_seal["run_id"] != written_final["run_id"]
            or parsed_seal["owner_nonce"] != written_final["owner_nonce"]
            or parsed_seal["observer_receipt_sha256"] != _digest(written_observer_payload)
        ):
            raise PublicationError("written SEAL does not bind exact written observer bytes")
        _assert_empty_terminal_marker(
            read,
            list_terminal_markers,
            run_dir,
            expected_marker,
            expected_identity=marker_identity,
        )
        try:
            _rename_with_outcome(rename, run_dir, final_dir)
        except PublicationRenameError as exc:
            final_owned = exc.final_owned
            raise
    except BaseException:
        if seal_attempted:
            _invalidate_seal(remove, run_dir, final_dir, final_owned=final_owned)
        raise
    return final_dir
