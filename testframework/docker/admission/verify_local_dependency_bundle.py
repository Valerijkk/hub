"""Closed, offline verifier for the G0 dependency bundle.

Archives are accepted only through one private no-follow descriptor. The public
result is declarative; `open_verified_bundle_session` retains descriptors for a
same-process helper staging transaction and never serialises their paths.
"""

from __future__ import annotations

import hashlib
import json
import os
import re
import stat
import tarfile
import unicodedata
from dataclasses import dataclass
from pathlib import Path
from typing import Any, BinaryIO, Final

SOURCE_COMMIT: Final = "61b38573682d7c01e6d9381babc5085537d7990f"
SCHEMA: Final = "hub.qa.dependency-bundle/v1"
BASE_RECEIPT_SCHEMA: Final = "hub.qa.local-base-inspect/v1"
SHA256_RE: Final = re.compile(r"^sha256:[0-9a-f]{64}$")
RUN_RE: Final = re.compile(r"^[a-z0-9][a-z0-9-]{7,63}$")
NONCE_RE: Final = re.compile(r"^[0-9a-f]{32}$")
MAX_STREAM_CHUNK: Final = 1_048_576
MAX_METADATA_BYTES: Final = 16_777_216
MAX_RAW_ARCHIVE_BYTES: Final = 2_147_483_648
MAX_ARCHIVE_MEMBERS: Final = 200_000
MAX_REGULAR_MEMBER_BYTES: Final = 268_435_456
MAX_EXPANDED_BYTES: Final = 4_294_967_296
MAX_PAX_BYTES: Final = 65_536
_OBSERVATION_PHASES: Final = (
    "opened", "post_hash", "post_raw_scan", "post_tarfile", "pre_stage", "post_stage",
)

REASON_CODES: Final = frozenset({
    "E_MANIFEST_UTF8", "E_MANIFEST_SCHEMA", "E_DUPLICATE_ASSET_ID",
    "E_ASSET_PATH_ESCAPE", "E_ASSET_MISSING", "E_ASSET_BYTES_MISMATCH",
    "E_ASSET_SHA256_MISMATCH", "E_ASSET_ARCHIVE_INVALID",
    "E_BUNDLE_SHA256_MISMATCH", "E_LICENSE_SPDX_INVALID",
    "E_LICENSE_EVIDENCE_MISSING", "E_PROVENANCE_EVIDENCE_MISSING",
    "E_BASE_MISSING", "E_BASE_ID_MISMATCH", "E_SOURCE_COMMIT_MISMATCH",
    "E_RECEIPT_BINDING_INVALID",
})
SPDX: Final = frozenset({
    "0BSD", "Apache-2.0", "Artistic-2.0", "BSD-2-Clause", "BSD-3-Clause",
    "CC0-1.0", "EPL-2.0", "ISC", "MIT", "MPL-2.0", "PSF-2.0",
    "Python-2.0", "Unicode-DFS-2016", "Zlib",
})
ASSET_LAYOUT: Final = (
    ("go-module-cache", ("backend-unit", "backend-openapi"), "assets/go-module-cache.tar"),
    ("go-race-toolchain", ("backend-unit",), "assets/go-race-toolchain.tar"),
    ("frontend-node-modules", ("frontend-unit",), "assets/frontend-node-modules.tar"),
    ("swag-binary", ("backend-openapi",), "assets/swag-binary.tar"),
)
BASE_LAYOUT: Final = (
    ("go-backend-unit", ("backend-unit",)),
    ("go-backend-openapi", ("backend-openapi",)),
    ("frontend-unit", ("frontend-unit",)),
    ("bootstrap-helper-python", ("g0-admission-helper", "g0-planning-selftest")),
)
_ROOT_KEYS: Final = ("schema", "bundle_id", "source_commit", "platform", "assets", "bases", "bundle_sha256")
_ASSET_KEYS: Final = (
    "asset_id", "targets", "relative_path", "sha256", "bytes", "archive_format",
    "member_count", "expanded_bytes", "layout_sha256", "license_spdx",
    "license_evidence_sha256", "provenance_sha256",
)
_BASE_KEYS: Final = (
    "base_id", "consumers", "local_image_id", "manifest_digest", "license_spdx",
    "license_evidence_sha256", "provenance_sha256",
)
_RECEIPT_KEYS: Final = ("schema", "task", "run", "owner_marker_sha256", "receipt_nonce", "bases")
_RECEIPT_BASE_KEYS: Final = ("base_id", "local_image_id", "manifest_digest")
_OWNER_KEYS: Final = ("schema", "handoff_id", "source_commit", "dependency_bundle_manifest_sha256", "owner", "released_at_utc")
_RACE_PROVENANCE_KEYS: Final = (
    "schema", "asset_id", "go_base_id", "go_base_manifest_digest", "go_base_local_image_id",
    "cc_path", "cc_sha256", "cc_target", "cgo_enabled",
)
_EXECUTABLE_MARKERS: Final = frozenset({
    "bin/cc", "bin/go", "node_modules/@biomejs/biome/bin/biome", "swag",
})


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


def _read_metadata_bytes(path: Path, failure_code: str) -> bytes:
    """Bounded metadata reader. Asset archives never reach this helper."""
    chunks: list[bytes] = []
    total = 0
    try:
        with path.open("rb") as handle:
            while True:
                chunk = handle.read(MAX_STREAM_CHUNK)
                if not chunk:
                    break
                total += len(chunk)
                if total > MAX_METADATA_BYTES:
                    _reject(failure_code)
                chunks.append(chunk)
    except OSError:
        _reject(failure_code)
    return b"".join(chunks)


def read_strict_json(path: Path, *, failure_code: str) -> tuple[bytes, Any]:
    if failure_code not in REASON_CODES:
        raise ValueError("unclosed dependency JSON failure code")
    raw = _read_metadata_bytes(path, failure_code)
    if raw.startswith(b"\xef\xbb\xbf"):
        _reject("E_MANIFEST_UTF8")
    try:
        value = json.loads(raw.decode("utf-8"), object_pairs_hook=_no_duplicate_object, parse_constant=_no_constant)
    except UnicodeDecodeError:
        _reject("E_MANIFEST_UTF8")
    except (json.JSONDecodeError, BundleVerificationError):
        _reject(failure_code)
    return raw, value


def sha256_prefixed(raw: bytes) -> str:
    return "sha256:" + hashlib.sha256(raw).hexdigest()


def canonical_bundle_sha256(bundle: dict[str, Any]) -> str:
    if not isinstance(bundle, dict) or tuple(bundle.keys()) != _ROOT_KEYS:
        _reject("E_BUNDLE_SHA256_MISMATCH")
    projection = {key: value for key, value in bundle.items() if key != "bundle_sha256"}
    try:
        raw = json.dumps(projection, ensure_ascii=False, sort_keys=False, separators=(",", ":"), allow_nan=False).encode("utf-8")
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


def _regular_relative(root: Path, relative: str, code: str, *, asset_only: bool = False, nonregular_code: str | None = None) -> Path:
    if not isinstance(relative, str) or relative.startswith("/") or "\\" in relative or (asset_only and not relative.startswith("assets/")):
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
            mode = parent.lstat().st_mode
        except OSError:
            _reject(code)
        if stat.S_ISLNK(mode) or not stat.S_ISDIR(mode):
            _reject(code)
        parent = parent.parent
    try:
        mode = candidate.lstat().st_mode
    except OSError:
        _reject(code)
    if stat.S_ISLNK(mode) or not stat.S_ISREG(mode):
        _reject(nonregular_code or code)
    return candidate


def _sha256_file(path: Path, code: str) -> str:
    try:
        with path.open("rb") as handle:
            digest = hashlib.sha256()
            while True:
                chunk = handle.read(MAX_STREAM_CHUNK)
                if not chunk:
                    break
                digest.update(chunk)
    except OSError:
        _reject(code)
    return "sha256:" + digest.hexdigest()


def _verify_evidence(root: Path, relative: str, expected_digest: str, code: str) -> None:
    if _sha256_file(_regular_relative(root, relative, code), code) != expected_digest:
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
    if _exact_keys(value, ("os", "architecture", "variant")) != {"os": "linux", "architecture": "amd64", "variant": None}:
        _reject("E_MANIFEST_SCHEMA")


def _validate_expected_owner_bindings(task_owner_marker_sha256: object, handoff_owner_marker_sha256: object) -> tuple[str, str]:
    task = _digest(task_owner_marker_sha256, "E_RECEIPT_BINDING_INVALID")
    handoff = _digest(handoff_owner_marker_sha256, "E_RECEIPT_BINDING_INVALID")
    if task == handoff:
        _reject("E_RECEIPT_BINDING_INVALID")
    return task, handoff


def _verify_owner(root: Path, manifest_sha: str, expected_handoff_owner_marker_sha256: str) -> str:
    owner_raw, owner = read_strict_json(root / "handoff-owner.v1.json", failure_code="E_RECEIPT_BINDING_INVALID")
    owner = _exact_keys(owner, _OWNER_KEYS, "E_RECEIPT_BINDING_INVALID")
    owner_sha = sha256_prefixed(owner_raw)
    if (
        owner["schema"] != "hub.qa.g0-offline-ingress-owner/v1"
        or not isinstance(owner["handoff_id"], str) or RUN_RE.fullmatch(owner["handoff_id"]) is None
        or _digest(owner["dependency_bundle_manifest_sha256"], "E_RECEIPT_BINDING_INVALID") != manifest_sha
        or not isinstance(owner["owner"], str) or not isinstance(owner["released_at_utc"], str)
    ):
        _reject("E_RECEIPT_BINDING_INVALID")
    if owner["source_commit"] != SOURCE_COMMIT:
        _reject("E_SOURCE_COMMIT_MISMATCH")
    if expected_handoff_owner_marker_sha256 != owner_sha:
        _reject("E_RECEIPT_BINDING_INVALID")
    return owner_sha


@dataclass(frozen=True)
class FileIdentity:
    device: int
    inode: int
    mode: int
    size: int
    mtime_ns: int


def _identity_from_stat(value: os.stat_result) -> FileIdentity:
    if not stat.S_ISREG(value.st_mode):
        _reject("E_ASSET_ARCHIVE_INVALID")
    return FileIdentity(value.st_dev, value.st_ino, value.st_mode, value.st_size, value.st_mtime_ns)


@dataclass(frozen=True)
class ArchiveMember:
    path: str
    kind: str
    mode: int
    size: int
    sha256: str | None


class HeldArchive:
    """Private same-process capability retaining one initial no-follow descriptor."""

    def __init__(self, path: Path, descriptor: int, identity: FileIdentity) -> None:
        self.path = path
        self.descriptor = descriptor
        self.identity = identity
        self._stream: BinaryIO | None = None
        self._closed = False
        self._observations: list[tuple[str, FileIdentity, FileIdentity]] = []

    def stream(self) -> BinaryIO:
        if self._closed:
            _reject("E_ASSET_ARCHIVE_INVALID")
        if self._stream is None:
            self._stream = os.fdopen(self.descriptor, "rb", closefd=False)
        return self._stream

    def _current_identities(self) -> tuple[FileIdentity, FileIdentity]:
        if self._closed:
            _reject("E_ASSET_ARCHIVE_INVALID")
        try:
            held = _identity_from_stat(os.fstat(self.descriptor))
            named = _identity_from_stat(self.path.lstat())
        except OSError:
            _reject("E_ASSET_ARCHIVE_INVALID")
        if held != self.identity or named != self.identity:
            _reject("E_ASSET_ARCHIVE_INVALID")
        return held, named

    def recheck(self) -> None:
        self._current_identities()

    def capture_observation(self, phase: str) -> None:
        """Record exactly one actual descriptor/path identity at a closed phase."""
        if not isinstance(phase, str) or len(self._observations) >= len(_OBSERVATION_PHASES) or phase != _OBSERVATION_PHASES[len(self._observations)]:
            _reject("E_ASSET_ARCHIVE_INVALID")
        descriptor, pathname = self._current_identities()
        self._observations.append((phase, descriptor, pathname))

    def receipt_observations(self) -> dict[str, dict[str, dict[str, int]]]:
        """Return only the fully captured, ordered observation transcript."""
        if tuple(phase for phase, _, _ in self._observations) != _OBSERVATION_PHASES:
            _reject("E_ASSET_ARCHIVE_INVALID")
        observations: dict[str, dict[str, dict[str, int]]] = {}
        for phase, descriptor, pathname in self._observations:
            observations[phase] = {
                "descriptor": {
                    "device": descriptor.device,
                    "inode": descriptor.inode,
                    "mode": descriptor.mode,
                    "size": descriptor.size,
                    "mtime_ns": descriptor.mtime_ns,
                },
                "pathname": {
                    "device": pathname.device,
                    "inode": pathname.inode,
                    "mode": pathname.mode,
                    "size": pathname.size,
                    "mtime_ns": pathname.mtime_ns,
                },
            }
        return observations

    def seek_start(self) -> BinaryIO:
        self.recheck()
        try:
            handle = self.stream()
            handle.seek(0)
        except OSError:
            _reject("E_ASSET_ARCHIVE_INVALID")
        return handle

    def close(self) -> None:
        if self._closed:
            return
        self._closed = True
        try:
            if self._stream is not None:
                self._stream.close()
        except OSError:
            pass
        try:
            os.close(self.descriptor)
        except OSError:
            pass


def _open_held_archive(path: Path) -> HeldArchive:
    nofollow = getattr(os, "O_NOFOLLOW", None)
    if nofollow is None:
        _reject("E_ASSET_ARCHIVE_INVALID")
    descriptor: int | None = None
    held: HeldArchive | None = None
    try:
        before = _identity_from_stat(path.lstat())
        descriptor = os.open(path, os.O_RDONLY | nofollow | getattr(os, "O_CLOEXEC", 0))
        held = HeldArchive(path, descriptor, _identity_from_stat(os.fstat(descriptor)))
        if before != held.identity:
            held.close()
            _reject("E_ASSET_ARCHIVE_INVALID")
        held.recheck()
        held.capture_observation("opened")
        return held
    except BundleVerificationError:
        if held is not None:
            held.close()
        elif descriptor is not None:
            try:
                os.close(descriptor)
            except OSError:
                pass
        raise
    except OSError:
        if held is not None:
            held.close()
        elif descriptor is not None:
            try:
                os.close(descriptor)
            except OSError:
                pass
        _reject("E_ASSET_ARCHIVE_INVALID")


def _read_exact(handle: BinaryIO, length: int) -> bytes:
    chunks: list[bytes] = []
    remaining = length
    while remaining:
        try:
            chunk = handle.read(min(MAX_STREAM_CHUNK, remaining))
        except OSError:
            _reject("E_ASSET_ARCHIVE_INVALID")
        if not chunk:
            _reject("E_ASSET_ARCHIVE_INVALID")
        chunks.append(chunk)
        remaining -= len(chunk)
    return b"".join(chunks)


def _read_digest(handle: BinaryIO, length: int) -> str:
    digest = hashlib.sha256()
    remaining = length
    while remaining:
        try:
            chunk = handle.read(min(MAX_STREAM_CHUNK, remaining))
        except OSError:
            _reject("E_ASSET_ARCHIVE_INVALID")
        if not chunk:
            _reject("E_ASSET_ARCHIVE_INVALID")
        digest.update(chunk)
        remaining -= len(chunk)
    return "sha256:" + digest.hexdigest()


def _consume_padding(handle: BinaryIO, length: int) -> None:
    padding = (-length) % 512
    if padding and any(byte != 0 for byte in _read_exact(handle, padding)):
        _reject("E_ASSET_ARCHIVE_INVALID")


def _stream_raw_integrity(held: HeldArchive, expected_bytes: int, expected_sha256: str) -> None:
    handle = held.seek_start()
    digest = hashlib.sha256()
    total = 0
    while True:
        try:
            chunk = handle.read(MAX_STREAM_CHUNK)
        except OSError:
            _reject("E_ASSET_ARCHIVE_INVALID")
        if not chunk:
            break
        total += len(chunk)
        if total > MAX_RAW_ARCHIVE_BYTES:
            _reject("E_ASSET_ARCHIVE_INVALID")
        digest.update(chunk)
    held.recheck()
    if total != expected_bytes:
        _reject("E_ASSET_BYTES_MISMATCH")
    if "sha256:" + digest.hexdigest() != expected_sha256:
        _reject("E_ASSET_SHA256_MISMATCH")
    held.capture_observation("post_hash")


def _tar_number(field: bytes) -> int:
    if not field or field[0] & 0x80:
        _reject("E_ASSET_ARCHIVE_INVALID")
    end = len(field)
    while end and field[end - 1] in {0, 32}:
        end -= 1
    content = field[:end]
    if any(value not in b"01234567" for value in content) or any(value not in {0, 32} for value in field[end:]):
        _reject("E_ASSET_ARCHIVE_INVALID")
    return int(content or b"0", 8)


def _tar_text(field: bytes) -> str:
    if b"\0" in field:
        value, padding = field.split(b"\0", 1)
        if any(padding):
            _reject("E_ASSET_ARCHIVE_INVALID")
    else:
        value = field
    try:
        return value.decode("utf-8", "strict")
    except UnicodeDecodeError:
        _reject("E_ASSET_ARCHIVE_INVALID")


def _normalised_path(value: str, *, is_directory: bool) -> str:
    if is_directory and value.endswith("/"):
        value = value[:-1]
    if not value or value.startswith("/") or "\\" in value or any(part in {"", ".", ".."} for part in value.split("/")):
        _reject("E_ASSET_ARCHIVE_INVALID")
    return value


def _parse_header(block: bytes) -> tuple[str, int, int, str]:
    if len(block) != 512 or block[257:263] != b"ustar\0" or block[263:265] != b"00":
        _reject("E_ASSET_ARCHIVE_INVALID")
    if _tar_number(block[148:156]) != sum(block[:148] + b" " * 8 + block[156:]):
        _reject("E_ASSET_ARCHIVE_INVALID")
    mode = _tar_number(block[100:108])
    uid = _tar_number(block[108:116])
    gid = _tar_number(block[116:124])
    size = _tar_number(block[124:136])
    mtime = _tar_number(block[136:148])
    if uid != 0 or gid != 0 or mtime != 0 or _tar_number(block[329:337]) != 0 or _tar_number(block[337:345]) != 0:
        _reject("E_ASSET_ARCHIVE_INVALID")
    if _tar_text(block[157:257]) or _tar_text(block[265:297]) or _tar_text(block[297:329]):
        _reject("E_ASSET_ARCHIVE_INVALID")
    name = _tar_text(block[:100])
    prefix = _tar_text(block[345:500])
    type_flag = block[156:157]
    if type_flag not in {b"\0", b"0", b"5", b"x"}:
        _reject("E_ASSET_ARCHIVE_INVALID")
    return type_flag.decode("ascii"), mode, size, f"{prefix}/{name}" if prefix else name


def _parse_pax(payload: bytes) -> str:
    if not payload:
        _reject("E_ASSET_ARCHIVE_INVALID")
    offset = 0
    values: dict[str, str] = {}
    while offset < len(payload):
        space = payload.find(b" ", offset)
        if space <= offset:
            _reject("E_ASSET_ARCHIVE_INVALID")
        digits = payload[offset:space]
        if not digits.isascii() or not digits.isdigit() or len(digits) > 5 or (len(digits) > 1 and digits.startswith(b"0")):
            _reject("E_ASSET_ARCHIVE_INVALID")
        length = int(digits)
        if length <= 0 or offset + length > len(payload):
            _reject("E_ASSET_ARCHIVE_INVALID")
        record = payload[offset:offset + length]
        if record[-1:] != b"\n" or record.find(b" ") != len(digits):
            _reject("E_ASSET_ARCHIVE_INVALID")
        try:
            key_value = record[space + 1:-1].decode("utf-8", "strict")
        except UnicodeDecodeError:
            _reject("E_ASSET_ARCHIVE_INVALID")
        if "=" not in key_value:
            _reject("E_ASSET_ARCHIVE_INVALID")
        key, value = key_value.split("=", 1)
        if key in values or key != "path":
            _reject("E_ASSET_ARCHIVE_INVALID")
        values[key] = value
        offset += length
    path = values.get("path")
    if path is None or not path:
        _reject("E_ASSET_ARCHIVE_INVALID")
    return path


def _validate_member(path: str, kind: str, mode: int, size: int, seen: set[str], seen_folded: set[str]) -> str:
    path = _normalised_path(path, is_directory=kind == "D")
    if kind == "D":
        if mode != 0o755 or size != 0:
            _reject("E_ASSET_ARCHIVE_INVALID")
    elif mode != 0o644 and not (mode == 0o755 and path in _EXECUTABLE_MARKERS):
        _reject("E_ASSET_ARCHIVE_INVALID")
    if mode & 0o7000:
        _reject("E_ASSET_ARCHIVE_INVALID")
    folded = unicodedata.normalize("NFC", path).casefold()
    if path in seen or folded in seen_folded:
        _reject("E_ASSET_ARCHIVE_INVALID")
    seen.add(path)
    seen_folded.add(folded)
    return path


def _scan_posix_pax_tar(held: HeldArchive) -> tuple[ArchiveMember, ...]:
    handle = held.seek_start()
    entries: list[ArchiveMember] = []
    seen: set[str] = set()
    seen_folded: set[str] = set()
    prior_path: bytes | None = None
    expanded = 0
    pending_pax: str | None = None
    has_pending_pax = False
    while True:
        header = _read_exact(handle, 512)
        if header == b"\0" * 512:
            if _read_exact(handle, 512) != b"\0" * 512:
                _reject("E_ASSET_ARCHIVE_INVALID")
            try:
                trailing = handle.read(1)
            except OSError:
                _reject("E_ASSET_ARCHIVE_INVALID")
            if trailing:
                _reject("E_ASSET_ARCHIVE_INVALID")
            break
        type_flag, mode, size, raw_path = _parse_header(header)
        if type_flag == "x":
            if has_pending_pax or size > MAX_PAX_BYTES:
                _reject("E_ASSET_ARCHIVE_INVALID")
            pending_pax = _parse_pax(_read_exact(handle, size))
            has_pending_pax = True
            _consume_padding(handle, size)
            continue
        if type_flag not in {"\0", "0", "5"}:
            _reject("E_ASSET_ARCHIVE_INVALID")
        kind = "D" if type_flag == "5" else "F"
        if has_pending_pax:
            if pending_pax is None:
                _reject("E_ASSET_ARCHIVE_INVALID")
            member_path = pending_pax
        else:
            member_path = raw_path
        path = _validate_member(member_path, kind, mode, size, seen, seen_folded)
        pending_pax = None
        has_pending_pax = False
        encoded_path = path.encode("utf-8")
        if prior_path is not None and encoded_path <= prior_path:
            _reject("E_ASSET_ARCHIVE_INVALID")
        prior_path = encoded_path
        if len(entries) >= MAX_ARCHIVE_MEMBERS:
            _reject("E_ASSET_ARCHIVE_INVALID")
        if kind == "F":
            if size > MAX_REGULAR_MEMBER_BYTES or expanded + size > MAX_EXPANDED_BYTES:
                _reject("E_ASSET_ARCHIVE_INVALID")
            digest = _read_digest(handle, size)
            expanded += size
        else:
            digest = None
        _consume_padding(handle, size)
        entries.append(ArchiveMember(path, kind, mode, size, digest))
    if has_pending_pax or not entries or expanded < 1:
        _reject("E_ASSET_ARCHIVE_INVALID")
    held.recheck()
    held.capture_observation("post_raw_scan")
    return tuple(entries)


def _tarfile_view(held: HeldArchive) -> tuple[ArchiveMember, ...]:
    handle = held.seek_start()
    entries: list[ArchiveMember] = []
    try:
        with tarfile.open(fileobj=handle, mode="r:") as archive:
            for member in archive:
                if member.isdir():
                    kind = "D"
                elif member.isreg():
                    kind = "F"
                else:
                    _reject("E_ASSET_ARCHIVE_INVALID")
                if member.uid != 0 or member.gid != 0 or member.mtime != 0 or member.uname or member.gname or set(member.pax_headers) - {"path"}:
                    _reject("E_ASSET_ARCHIVE_INVALID")
                path = _normalised_path(member.name, is_directory=kind == "D")
                mode = member.mode & 0o7777
                if kind == "F":
                    stream = archive.extractfile(member)
                    if stream is None:
                        _reject("E_ASSET_ARCHIVE_INVALID")
                    with stream:
                        digest = _read_digest(stream, member.size)
                else:
                    digest = None
                entries.append(ArchiveMember(path, kind, mode, member.size, digest))
    except (OSError, tarfile.TarError):
        _reject("E_ASSET_ARCHIVE_INVALID")
    held.recheck()
    held.capture_observation("post_tarfile")
    return tuple(entries)


def _layout_sha256(entries: tuple[ArchiveMember, ...]) -> str:
    lines: list[bytes] = []
    for member in sorted(entries, key=lambda value: value.path.encode("utf-8")):
        if member.kind == "D":
            lines.append(f"D\t{member.path}\t0755\t0\t-\n".encode("utf-8"))
        elif member.sha256 is not None:
            lines.append(f"F\t{member.path}\t{member.mode:04o}\t{member.size}\t{member.sha256}\n".encode("utf-8"))
        else:
            _reject("E_ASSET_ARCHIVE_INVALID")
    return "sha256:" + hashlib.sha256(b"".join(lines)).hexdigest()


def _assert_markers(asset_id: str, entries: tuple[ArchiveMember, ...]) -> None:
    observed = {member.path: member for member in entries}
    if asset_id == "go-module-cache":
        required = {"pkg/mod/cache/download": ("D", 0o755)}
    elif asset_id == "go-race-toolchain":
        required = {"bin/cc": ("F", 0o755), "bin/go": ("F", 0o755)}
    elif asset_id == "frontend-node-modules":
        required = {
            "node_modules": ("D", 0o755),
            "node_modules/typescript/bin/tsc": ("F", 0o644),
            "node_modules/@biomejs/biome/bin/biome": ("F", 0o755),
            "node_modules/vite/bin/vite.js": ("F", 0o644),
            "node_modules/vitest/vitest.mjs": ("F", 0o644),
        }
    elif asset_id == "swag-binary":
        required = {"swag": ("F", 0o755)}
    else:
        _reject("E_ASSET_ARCHIVE_INVALID")
    for path, (kind, mode) in required.items():
        member = observed.get(path)
        if member is None or member.kind != kind or member.mode != mode:
            _reject("E_ASSET_ARCHIVE_INVALID")


@dataclass
class AdmittedArchive:
    asset_id: str
    held: HeldArchive
    members: tuple[ArchiveMember, ...]
    public: dict[str, Any]


class VerifiedBundleSession:
    """Private, helper-only source capability; it must be closed in-process."""

    def __init__(self, bundle: dict[str, Any], result: dict[str, Any], archives: tuple[AdmittedArchive, ...]) -> None:
        self.bundle = bundle
        self.result = result
        self.archives = archives
        self._closed = False

    def close(self) -> None:
        if self._closed:
            return
        self._closed = True
        for archive in self.archives:
            archive.held.close()

    def __enter__(self) -> "VerifiedBundleSession":
        return self

    def __exit__(self, exc_type: object, exc_value: object, traceback: object) -> None:
        self.close()


def _verify_assets(bundle: dict[str, Any], root: Path) -> tuple[set[str], list[dict[str, Any]], tuple[AdmittedArchive, ...]]:
    values = bundle["assets"]
    if not isinstance(values, list):
        _reject("E_MANIFEST_SCHEMA")
    seen_ids: set[str] = set()
    for value in values:
        if not isinstance(value, dict) or not isinstance(value.get("asset_id"), str):
            _reject("E_MANIFEST_SCHEMA")
        if value["asset_id"] in seen_ids:
            _reject("E_DUPLICATE_ASSET_ID")
        seen_ids.add(value["asset_id"])
    if len(values) != len(ASSET_LAYOUT):
        _reject("E_MANIFEST_SCHEMA")
    expected_files = {"handoff-owner.v1.json", "dependency-bundle.manifest.v1.json"}
    public_assets: list[dict[str, Any]] = []
    admitted: list[AdmittedArchive] = []
    try:
        for value, expected in zip(values, ASSET_LAYOUT, strict=True):
            asset = _exact_keys(value, _ASSET_KEYS)
            asset_id, targets, relative_path = expected
            supplied_path = asset["relative_path"]
            if not isinstance(supplied_path, str) or not supplied_path.startswith("assets/") or supplied_path.startswith("/") or "\\" in supplied_path or any(not part or part in {".", ".."} for part in supplied_path.split("/")):
                _reject("E_ASSET_PATH_ESCAPE")
            if (
                asset["asset_id"] != asset_id or asset["targets"] != list(targets) or asset["relative_path"] != relative_path
                or asset["archive_format"] != "posix-pax-tar"
                or not isinstance(asset["bytes"], int) or isinstance(asset["bytes"], bool) or asset["bytes"] < 1
                or not isinstance(asset["member_count"], int) or isinstance(asset["member_count"], bool) or asset["member_count"] < 1
                or not isinstance(asset["expanded_bytes"], int) or isinstance(asset["expanded_bytes"], bool) or asset["expanded_bytes"] < 1
            ):
                _reject("E_MANIFEST_SCHEMA")
            if not isinstance(asset["license_spdx"], str) or asset["license_spdx"] not in SPDX:
                _reject("E_LICENSE_SPDX_INVALID")
            digest = _digest(asset["sha256"])
            layout_digest = _digest(asset["layout_sha256"])
            license_digest = _digest(asset["license_evidence_sha256"])
            provenance_digest = _digest(asset["provenance_sha256"])
            held = _open_held_archive(_regular_relative(root, supplied_path, "E_ASSET_MISSING", asset_only=True, nonregular_code="E_ASSET_ARCHIVE_INVALID"))
            try:
                _stream_raw_integrity(held, asset["bytes"], digest)
                raw_members = _scan_posix_pax_tar(held)
                if raw_members != _tarfile_view(held):
                    _reject("E_ASSET_ARCHIVE_INVALID")
                if len(raw_members) != asset["member_count"] or sum(member.size for member in raw_members if member.kind == "F") != asset["expanded_bytes"] or _layout_sha256(raw_members) != layout_digest:
                    _reject("E_ASSET_ARCHIVE_INVALID")
                _assert_markers(asset_id, raw_members)
                held.recheck()
            except BaseException:
                held.close()
                raise
            license_path = f"evidence/licenses/assets/{asset_id}.txt"
            provenance_path = f"evidence/provenance/assets/{asset_id}.json"
            _verify_evidence(root, license_path, license_digest, "E_LICENSE_EVIDENCE_MISSING")
            _verify_evidence(root, provenance_path, provenance_digest, "E_PROVENANCE_EVIDENCE_MISSING")
            expected_files.update({relative_path, license_path, provenance_path})
            public = {
                "asset_id": asset_id, "targets": list(targets), "relative_path": relative_path,
                "sha256": digest, "bytes": asset["bytes"], "archive_format": "posix-pax-tar",
                "member_count": asset["member_count"], "expanded_bytes": asset["expanded_bytes"], "layout_sha256": layout_digest,
            }
            public_assets.append(public)
            admitted.append(AdmittedArchive(asset_id, held, raw_members, public))
    except BaseException:
        for archive in admitted:
            archive.held.close()
        raise
    return expected_files, public_assets, tuple(admitted)


def _verify_base_receipt(bundle: dict[str, Any], root: Path, receipt_path: Path, receipt_sha256: str, receipt_nonce: str, task_owner_marker_sha256: str) -> tuple[set[str], list[dict[str, Any]]]:
    if SHA256_RE.fullmatch(receipt_sha256) is None or NONCE_RE.fullmatch(receipt_nonce) is None:
        _reject("E_RECEIPT_BINDING_INVALID")
    raw, receipt = read_strict_json(receipt_path, failure_code="E_RECEIPT_BINDING_INVALID")
    if sha256_prefixed(raw) != receipt_sha256:
        _reject("E_RECEIPT_BINDING_INVALID")
    receipt = _exact_keys(receipt, _RECEIPT_KEYS, "E_RECEIPT_BINDING_INVALID")
    if receipt["schema"] != BASE_RECEIPT_SCHEMA or receipt["task"] != "g0-assets-runner" or not isinstance(receipt["run"], str) or RUN_RE.fullmatch(receipt["run"]) is None or _digest(receipt["owner_marker_sha256"], "E_RECEIPT_BINDING_INVALID") != task_owner_marker_sha256 or receipt["receipt_nonce"] != receipt_nonce:
        _reject("E_RECEIPT_BINDING_INVALID")
    values = bundle["bases"]
    receipt_values = receipt["bases"]
    if not isinstance(values, list) or not isinstance(receipt_values, list) or len(values) != len(BASE_LAYOUT) or len(receipt_values) != len(BASE_LAYOUT):
        _reject("E_BASE_MISSING")
    expected_files: set[str] = set()
    public_bases: list[dict[str, Any]] = []
    for value, receipt_value, expected in zip(values, receipt_values, BASE_LAYOUT, strict=True):
        base = _exact_keys(value, _BASE_KEYS)
        receipt_base = _exact_keys(receipt_value, _RECEIPT_BASE_KEYS, "E_RECEIPT_BINDING_INVALID")
        base_id, consumers = expected
        if base["base_id"] != base_id or base["consumers"] != list(consumers) or receipt_base["base_id"] != base_id:
            _reject("E_BASE_ID_MISMATCH")
        if not isinstance(base["license_spdx"], str) or base["license_spdx"] not in SPDX:
            _reject("E_LICENSE_SPDX_INVALID")
        image_id = _digest(base["local_image_id"], "E_BASE_ID_MISMATCH")
        manifest_digest = _digest(base["manifest_digest"], "E_BASE_ID_MISMATCH")
        if receipt_base["local_image_id"] != image_id or receipt_base["manifest_digest"] != manifest_digest:
            _reject("E_BASE_ID_MISMATCH")
        license_path = f"evidence/licenses/bases/{base_id}.txt"
        provenance_path = f"evidence/provenance/bases/{base_id}.json"
        _verify_evidence(root, license_path, _digest(base["license_evidence_sha256"]), "E_LICENSE_EVIDENCE_MISSING")
        _verify_evidence(root, provenance_path, _digest(base["provenance_sha256"]), "E_PROVENANCE_EVIDENCE_MISSING")
        expected_files.update({license_path, provenance_path})
        public_bases.append({"base_id": base_id, "consumers": list(consumers), "local_image_id": image_id, "manifest_digest": manifest_digest})
    return expected_files, public_bases


def _verify_race_provenance(root: Path, bundle: dict[str, Any], archives: tuple[AdmittedArchive, ...]) -> None:
    race = next((archive for archive in archives if archive.asset_id == "go-race-toolchain"), None)
    cc = next((member for member in race.members if member.path == "bin/cc"), None) if race is not None else None
    if cc is None or cc.sha256 is None:
        _reject("E_ASSET_ARCHIVE_INVALID")
    raw, value = read_strict_json(root / "evidence/provenance/assets/go-race-toolchain.json", failure_code="E_PROVENANCE_EVIDENCE_MISSING")
    asset = bundle["assets"][1]
    if sha256_prefixed(raw) != asset["provenance_sha256"]:
        _reject("E_PROVENANCE_EVIDENCE_MISSING")
    value = _exact_keys(value, _RACE_PROVENANCE_KEYS, "E_PROVENANCE_EVIDENCE_MISSING")
    base = bundle["bases"][0]
    if value != {"schema": "hub.qa.go-race-toolchain-provenance/v1", "asset_id": "go-race-toolchain", "go_base_id": "go-backend-unit", "go_base_manifest_digest": base["manifest_digest"], "go_base_local_image_id": base["local_image_id"], "cc_path": "bin/cc", "cc_sha256": cc.sha256, "cc_target": "linux/amd64", "cgo_enabled": True}:
        _reject("E_PROVENANCE_EVIDENCE_MISSING")


def open_verified_bundle_session(manifest_path: Path, ingress_root: Path, base_receipt_path: Path, base_receipt_sha256: str, receipt_nonce: str, *, expected_task_owner_marker_sha256: str, expected_handoff_owner_marker_sha256: str) -> VerifiedBundleSession:
    """Return a private descriptor-owning session after full archive admission."""
    task_owner, handoff_owner = _validate_expected_owner_bindings(expected_task_owner_marker_sha256, expected_handoff_owner_marker_sha256)
    manifest_raw, bundle = read_strict_json(manifest_path, failure_code="E_MANIFEST_SCHEMA")
    bundle = _exact_keys(bundle, _ROOT_KEYS)
    if bundle["schema"] != SCHEMA or not isinstance(bundle["bundle_id"], str) or RUN_RE.fullmatch(bundle["bundle_id"]) is None:
        _reject("E_MANIFEST_SCHEMA")
    if bundle["source_commit"] != SOURCE_COMMIT:
        _reject("E_SOURCE_COMMIT_MISMATCH")
    _verify_platform(bundle["platform"])
    if _digest(bundle["bundle_sha256"]) != canonical_bundle_sha256(bundle):
        _reject("E_BUNDLE_SHA256_MISMATCH")
    ingress_raw = _read_metadata_bytes(_regular_relative(ingress_root, "dependency-bundle.manifest.v1.json", "E_MANIFEST_SCHEMA"), "E_MANIFEST_SCHEMA")
    if ingress_raw != manifest_raw:
        _reject("E_MANIFEST_SCHEMA")
    manifest_sha = sha256_prefixed(manifest_raw)
    _verify_owner(ingress_root, manifest_sha, handoff_owner)
    expected_files, assets, archives = _verify_assets(bundle, ingress_root)
    try:
        base_files, bases = _verify_base_receipt(bundle, ingress_root, base_receipt_path, base_receipt_sha256, receipt_nonce, task_owner)
        _verify_race_provenance(ingress_root, bundle, archives)
        expected_files.update(base_files)
        _verify_closed_tree(ingress_root, expected_files)
        return VerifiedBundleSession(bundle, {"verdict": "ACCEPT", "manifest_sha256": manifest_sha, "assets": assets, "bases": bases, "reason_code": None}, archives)
    except BaseException:
        for archive in archives:
            archive.held.close()
        raise


def _verify_bundle_or_raise(manifest_path: Path, ingress_root: Path, base_receipt_path: Path, base_receipt_sha256: str, receipt_nonce: str, *, expected_task_owner_marker_sha256: str, expected_handoff_owner_marker_sha256: str) -> tuple[dict[str, Any], dict[str, Any]]:
    with open_verified_bundle_session(manifest_path, ingress_root, base_receipt_path, base_receipt_sha256, receipt_nonce, expected_task_owner_marker_sha256=expected_task_owner_marker_sha256, expected_handoff_owner_marker_sha256=expected_handoff_owner_marker_sha256) as session:
        return session.bundle, session.result


def verify_bundle(manifest_path: Path, ingress_root: Path, base_receipt_path: Path, base_receipt_sha256: str, receipt_nonce: str, *, expected_task_owner_marker_sha256: str, expected_handoff_owner_marker_sha256: str) -> dict[str, Any]:
    """Return only the five-field public acceptance projection or raise."""
    _, result = _verify_bundle_or_raise(manifest_path, ingress_root, base_receipt_path, base_receipt_sha256, receipt_nonce, expected_task_owner_marker_sha256=expected_task_owner_marker_sha256, expected_handoff_owner_marker_sha256=expected_handoff_owner_marker_sha256)
    return result
