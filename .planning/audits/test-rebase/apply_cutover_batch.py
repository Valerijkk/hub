#!/usr/bin/env python3
"""Check or transactionally apply one literal Phase 129 cutover batch.

This Phase 128 tool never executes manifest strings.  Application is opt-in,
bounded to 56 literal file operations, uses descriptor-relative no-follow
traversal, and rolls back every completed rename if any later step fails.
Delete before-images remain in a task-owned transaction directory until the
atomic Git commit has been independently verified and exact cleanup is safe.
"""

from __future__ import annotations

import argparse
import ctypes
import errno
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import re
import runpy
import secrets
import stat
import sys
import tempfile
from typing import Any
import unicodedata

_SEALED_REPO_ROOT = os.environ.get("HUB_PHASE128_REPO_ROOT")
REPO_ROOT = (
    Path(_SEALED_REPO_ROOT)
    if _SEALED_REPO_ROOT is not None
    else Path(__file__).resolve().parents[3]
)
if not REPO_ROOT.is_absolute():
    raise RuntimeError("HUB_PHASE128_REPO_ROOT must be absolute")
DEFAULT_MANIFEST = ".planning/audits/test-rebase/cutover-manifest.json"
TRANSACTION_PARENT = ".planning/audits/test-rebase"
DIRECTORY_OWNER_MARKER = ".hub-cutover-owner"
MAX_MANIFEST_BYTES = 32 * 1024 * 1024
MAX_SOURCE_BYTES = 64 * 1024 * 1024
MAX_BATCH_OPERATIONS = 56
MAX_PATH_COMPONENTS = 64
MAX_COMPONENT_UTF8_BYTES = 255
MAX_PATH_UTF8_BYTES = 4096
PROTECTED_PREFIX = "pro-qa-ai/"
SHA256_RE = re.compile(r"^[0-9a-f]{64}$")
GIT_SHA_RE = re.compile(r"^[0-9a-f]{40}(?:[0-9a-f]{24})?$")
SLUG_RE = re.compile(r"^[a-z0-9]+(?:-[a-z0-9]+)*$")
REPARSE = getattr(stat, "FILE_ATTRIBUTE_REPARSE_POINT", 0x400)
_INDEPENDENT_VALIDATOR: Any = None


class CutoverError(ValueError):
    pass


class MissingPathError(CutoverError):
    """A proven ENOENT; other path failures remain fatal."""


class SimulatedProcessCrash(BaseException):
    """Self-test only: bypass in-process recovery to exercise receipt recovery."""


def validate_manifest_data(*args: Any, **kwargs: Any) -> list[str]:
    """Load the sibling independent validator by exact source path, not sys.path."""

    global _INDEPENDENT_VALIDATOR
    if _INDEPENDENT_VALIDATOR is None:
        namespace = runpy.run_path(
            str(Path(__file__).with_name("validate_cutover_manifest.py")),
            run_name="phase128_independent_cutover_validator",
        )
        candidate = namespace.get("validate_manifest_data")
        if not callable(candidate):
            raise CutoverError("independent cutover validator entrypoint is missing")
        _INDEPENDENT_VALIDATOR = candidate
    return _INDEPENDENT_VALIDATOR(*args, **kwargs)


def rename_noreplace(source_fd: int, source: str, target_fd: int, target: str) -> None:
    """Linux descriptor-relative atomic rename that can never replace target."""

    renameat2 = getattr(ctypes.CDLL(None, use_errno=True), "renameat2", None)
    if renameat2 is None:
        raise CutoverError("renameat2(RENAME_NOREPLACE) is unavailable")
    renameat2.argtypes = [ctypes.c_int, ctypes.c_char_p, ctypes.c_int, ctypes.c_char_p, ctypes.c_uint]
    renameat2.restype = ctypes.c_int
    if renameat2(source_fd, os.fsencode(source), target_fd, os.fsencode(target), 1) == 0:
        return
    number = ctypes.get_errno()
    if number in {errno.ENOSYS, errno.EINVAL, errno.ENOTSUP}:
        raise CutoverError("renameat2(RENAME_NOREPLACE) is unsupported")
    if number in {errno.EEXIST, errno.ENOTEMPTY}:
        raise CutoverError(f"atomic no-replace destination exists: {target}")
    raise OSError(number, os.strerror(number), target)


def canonical_path(value: Any, label: str) -> str:
    if not isinstance(value, str) or not value:
        raise CutoverError(f"{label} must be a non-empty string")
    if "\\" in value or value.startswith(("/", "//")) or re.match(r"^[A-Za-z]:", value):
        raise CutoverError(f"{label} is not repository-relative POSIX")
    if unicodedata.normalize("NFC", value) != value or any(ord(char) < 32 or ord(char) == 127 for char in value):
        raise CutoverError(f"{label} is not canonical Unicode")
    parts = value.split("/")
    if any(part in {"", ".", ".."} or part.rstrip(" .") != part or ":" in part for part in parts):
        raise CutoverError(f"{label} contains a traversal/alias segment")
    if len(parts) > MAX_PATH_COMPONENTS:
        raise CutoverError(f"{label} exceeds {MAX_PATH_COMPONENTS} component budget")
    if len(value.encode("utf-8")) > MAX_PATH_UTF8_BYTES:
        raise CutoverError(f"{label} exceeds {MAX_PATH_UTF8_BYTES} UTF-8 byte budget")
    if any(len(part.encode("utf-8")) > MAX_COMPONENT_UTF8_BYTES for part in parts):
        raise CutoverError(f"{label} component exceeds {MAX_COMPONENT_UTF8_BYTES} UTF-8 byte budget")
    if PurePosixPath(value).as_posix() != value:
        raise CutoverError(f"{label} is not canonical POSIX-relative")
    if value == "pro-qa-ai" or value.startswith(PROTECTED_PREFIX):
        raise CutoverError(f"{label} enters protected pro-qa-ai")
    return value


def signature(meta: os.stat_result, *, content: bool) -> tuple[int, ...]:
    value = (int(meta.st_mode), int(meta.st_dev), int(meta.st_ino), int(meta.st_nlink), int(getattr(meta, "st_file_attributes", 0)))
    if content:
        value += (
            int(meta.st_size), int(getattr(meta, "st_mtime_ns", 0)),
            int(getattr(meta, "st_ctime_ns", 0)),
        )
    return value


def is_link(meta: os.stat_result) -> bool:
    return stat.S_ISLNK(meta.st_mode) or bool(int(getattr(meta, "st_file_attributes", 0)) & REPARSE)


def canonical_json(value: Any) -> bytes:
    return (json.dumps(value, ensure_ascii=False, indent=2, sort_keys=True) + "\n").encode("utf-8")


def open_physical_root(path: Path, flags: int) -> tuple[int, tuple[int, ...]]:
    """Open every absolute repo-root component no-follow and reject protected roots."""

    absolute = Path(os.path.abspath(path))
    if (
        len(absolute.parts) - 1 > MAX_PATH_COMPONENTS
        or len(str(absolute).encode("utf-8")) > MAX_PATH_UTF8_BYTES
        or any(len(part.encode("utf-8")) > MAX_COMPONENT_UTF8_BYTES for part in absolute.parts[1:])
    ):
        raise CutoverError("repository root exceeds component/path budget")
    protected = os.path.normcase(os.path.abspath(REPO_ROOT / PROTECTED_PREFIX.rstrip("/")))
    candidate = os.path.normcase(str(absolute))
    try:
        inside_protected = os.path.commonpath([candidate, protected]) == protected
    except ValueError:
        inside_protected = False
    if inside_protected:
        raise CutoverError("repository root enters protected pro-qa-ai")
    anchor = Path(absolute.anchor)
    descriptor = os.open(anchor, flags)
    try:
        for part in absolute.parts[1:]:
            before = os.stat(part, dir_fd=descriptor, follow_symlinks=False)
            if is_link(before) or not stat.S_ISDIR(before.st_mode):
                raise CutoverError("repository root ancestor is not physical")
            next_fd = os.open(part, flags, dir_fd=descriptor)
            opened = os.fstat(next_fd)
            if signature(before, content=False) != signature(opened, content=False):
                os.close(next_fd)
                raise CutoverError("repository root ancestor changed while opening")
            os.close(descriptor)
            descriptor = next_fd
        return descriptor, signature(os.fstat(descriptor), content=False)
    except Exception:
        os.close(descriptor)
        raise


class Repo:
    """Held-descriptor repository boundary for reads and literal renames."""

    def __init__(self, root: Path):
        required = ("O_CLOEXEC", "O_DIRECTORY", "O_NOFOLLOW", "O_PATH")
        missing = [name for name in required if not hasattr(os, name)]
        if os.name != "posix" or missing or os.open not in getattr(os, "supports_dir_fd", set()):
            raise CutoverError(f"descriptor-relative no-follow unavailable: os={os.name} missing={missing}")
        self.root = Path(os.path.abspath(root))
        self.dir_flags = os.O_RDONLY | os.O_CLOEXEC | os.O_DIRECTORY | os.O_NOFOLLOW
        self.file_flags = os.O_RDONLY | os.O_CLOEXEC | os.O_NOFOLLOW | getattr(os, "O_NONBLOCK", 0)
        self.root_fd, root_signature = open_physical_root(self.root, self.dir_flags)
        opened = os.fstat(self.root_fd)
        if root_signature != signature(opened, content=False):
            os.close(self.root_fd)
            raise CutoverError("repository root link/identity rejected")
        self.root_signature = root_signature

    def close(self) -> None:
        if self.root_fd >= 0:
            os.close(self.root_fd)
            self.root_fd = -1

    def ensure_identity(self) -> None:
        check_fd, current = open_physical_root(self.root, self.dir_flags)
        os.close(check_fd)
        if current != self.root_signature or signature(os.fstat(self.root_fd), content=False) != self.root_signature:
            raise CutoverError("repository root identity changed")

    def open_directory(self, relative: str, *, create: bool = False, created: list[str] | None = None) -> int:
        canonical_path(relative, "repository directory")
        current_fd = os.dup(self.root_fd)
        prefix: list[str] = []
        try:
            for part in relative.split("/"):
                prefix.append(part)
                try:
                    before = os.stat(part, dir_fd=current_fd, follow_symlinks=False)
                except FileNotFoundError:
                    if not create:
                        raise MissingPathError(f"repository directory is absent: {relative}") from None
                    os.mkdir(part, 0o755, dir_fd=current_fd)
                    os.fsync(current_fd)
                    if created is not None:
                        created.append("/".join(prefix))
                    before = os.stat(part, dir_fd=current_fd, follow_symlinks=False)
                if is_link(before) or not stat.S_ISDIR(before.st_mode):
                    raise CutoverError(f"physical directory required: {relative}")
                next_fd = os.open(part, self.dir_flags, dir_fd=current_fd)
                opened = os.fstat(next_fd)
                if signature(before, content=False) != signature(opened, content=False):
                    os.close(next_fd)
                    raise CutoverError(f"directory identity changed: {relative}")
                os.close(current_fd)
                current_fd = next_fd
            return current_fd
        except Exception:
            os.close(current_fd)
            raise

    def open_parent(self, relative: str, *, create: bool = False, created: list[str] | None = None) -> tuple[int, str]:
        path = canonical_path(relative, "operation path")
        pure = PurePosixPath(path)
        parent = pure.parent.as_posix()
        parent_fd = os.dup(self.root_fd) if parent == "." else self.open_directory(parent, create=create, created=created)
        return parent_fd, pure.name

    def read(self, relative: str, limit: int) -> bytes:
        parent_fd, leaf = self.open_parent(relative)
        descriptor = -1
        try:
            try:
                before = os.stat(leaf, dir_fd=parent_fd, follow_symlinks=False)
            except FileNotFoundError:
                raise MissingPathError(f"required file is absent: {relative}") from None
            if is_link(before):
                raise CutoverError(f"link/reparse source rejected: {relative}")
            descriptor = os.open(leaf, self.file_flags, dir_fd=parent_fd)
            opened = os.fstat(descriptor)
            if signature(before, content=True) != signature(opened, content=True):
                raise CutoverError(f"source identity changed while opening: {relative}")
            if not stat.S_ISREG(opened.st_mode) or int(opened.st_nlink) != 1 or int(opened.st_size) > limit:
                raise CutoverError(f"bounded single-link regular file required: {relative}")
            chunks: list[bytes] = []
            size = 0
            while True:
                chunk = os.read(descriptor, min(1024 * 1024, limit + 1 - size))
                if not chunk:
                    break
                chunks.append(chunk)
                size += len(chunk)
                if size > limit:
                    raise CutoverError(f"read budget exceeded: {relative}")
            if signature(opened, content=True) != signature(os.fstat(descriptor), content=True):
                raise CutoverError(f"source changed during read: {relative}")
            return b"".join(chunks)
        finally:
            if descriptor >= 0:
                os.close(descriptor)
            os.close(parent_fd)

    def head(self) -> str:
        text = self.read(".git/HEAD", 4096).decode("ascii").strip()
        if text.startswith("ref: "):
            reference = canonical_path(f".git/{text[5:]}", "Git HEAD reference")
            try:
                text = self.read(reference, 4096).decode("ascii").strip()
            except MissingPathError:
                packed = self.read(".git/packed-refs", 8 * 1024 * 1024).decode("ascii")
                matches = [line.split(" ", 1)[0] for line in packed.splitlines() if line.endswith(f" {reference[5:]}")]
                if len(matches) != 1:
                    raise
                text = matches[0]
        if GIT_SHA_RE.fullmatch(text) is None:
            raise CutoverError("repository HEAD is not canonical")
        return text


def require_literal_operation_budget(value: Any, label: str) -> list[Any]:
    if not isinstance(value, list) or not value or len(value) > MAX_BATCH_OPERATIONS:
        raise CutoverError(
            f"{label} must contain 1..{MAX_BATCH_OPERATIONS} literal operations"
        )
    return value


def load_batch(repo: Repo, manifest_path: str, batch_id: str) -> tuple[dict[str, Any], bytes]:
    if manifest_path != DEFAULT_MANIFEST:
        raise CutoverError(f"only canonical manifest path is accepted: {DEFAULT_MANIFEST}")
    payload = repo.read(canonical_path(manifest_path, "manifest"), MAX_MANIFEST_BYTES)
    manifest = json.loads(payload.decode("utf-8"))
    if not isinstance(manifest, dict) or canonical_json(manifest) != payload:
        raise CutoverError("manifest must be a canonical JSON object")
    if manifest.get("status") not in {"planned", "in-progress"}:
        raise CutoverError("manifest status is not applicable")
    source = manifest.get("source")
    if not isinstance(source, dict) or source.get("repository_head") != repo.head():
        raise CutoverError("manifest repository head does not match live HEAD")
    validation_errors = validate_manifest_data(manifest, repo.root, check_filesystem=True)
    if validation_errors:
        raise CutoverError(f"independent manifest validation failed: {validation_errors[0]}")
    batches = manifest.get("batches")
    matches = [item for item in batches if isinstance(item, dict) and item.get("id") == batch_id] if isinstance(batches, list) else []
    if len(matches) != 1:
        raise CutoverError("batch ID must identify exactly one manifest batch")
    batch = matches[0]
    if batch.get("state") != "planned":
        raise CutoverError("only a planned batch may be applied")
    operations = require_literal_operation_budget(batch.get("operations"), "batch")
    if batch.get("preconditions", {}).get("operations_sha256") != hashlib.sha256(
        json.dumps(operations, ensure_ascii=False, separators=(",", ":"), sort_keys=True).encode("utf-8")
    ).hexdigest():
        raise CutoverError("operation digest mismatch")
    return batch, payload


def validate_operations(repo: Repo, batch: dict[str, Any]) -> list[dict[str, Any]]:
    normalized: list[dict[str, Any]] = []
    sources: set[str] = set()
    targets: set[str] = set()
    identities: dict[tuple[int, int], str] = {}
    for index, raw in enumerate(batch["operations"]):
        if not isinstance(raw, dict) or raw.get("action") not in {"move", "delete"}:
            raise CutoverError(f"operation {index} is not an executable move/delete")
        source = canonical_path(raw.get("source"), f"source {index}")
        expected_sha = raw.get("source_sha256")
        if SHA256_RE.fullmatch(str(expected_sha)) is None:
            raise CutoverError(f"source digest invalid: {source}")
        if source in sources:
            raise CutoverError(f"duplicate source: {source}")
        payload = repo.read(source, MAX_SOURCE_BYTES)
        if hashlib.sha256(payload).hexdigest() != expected_sha:
            raise CutoverError(f"source digest mismatch: {source}")
        parent_fd, leaf = repo.open_parent(source)
        try:
            meta = os.stat(leaf, dir_fd=parent_fd, follow_symlinks=False)
        finally:
            os.close(parent_fd)
        identity = (int(meta.st_dev), int(meta.st_ino))
        previous = identities.setdefault(identity, source)
        if previous != source:
            raise CutoverError(f"hardlink/identity collision: {previous} and {source}")
        confirmation = repo.read(source, MAX_SOURCE_BYTES)
        confirmation_parent, confirmation_leaf = repo.open_parent(source)
        try:
            confirmed_meta = os.stat(confirmation_leaf, dir_fd=confirmation_parent, follow_symlinks=False)
        finally:
            os.close(confirmation_parent)
        if payload != confirmation or signature(meta, content=True) != signature(confirmed_meta, content=True):
            raise CutoverError(f"source changed during independent recheck: {source}")
        operation = {
            "action": raw["action"],
            "source": source,
            "source_sha256": expected_sha,
            "_source_signature": signature(confirmed_meta, content=True),
        }
        sources.add(source)
        if raw["action"] == "move":
            target = canonical_path(raw.get("target"), f"target {index}")
            if target in targets or target in sources or target == source:
                raise CutoverError(f"move target collision: {target}")
            # Missing ancestors are allowed in check mode, but an existing
            # component may never be a link and an existing target is fatal.
            try:
                target_parent, target_leaf = repo.open_parent(target)
            except MissingPathError:
                pass
            else:
                try:
                    try:
                        target_meta = os.stat(target_leaf, dir_fd=target_parent, follow_symlinks=False)
                    except FileNotFoundError:
                        target_meta = None
                    if target_meta is not None:
                        raise CutoverError(f"existing move destination rejected: {target}")
                finally:
                    os.close(target_parent)
            operation["target"] = target
            targets.add(target)
        normalized.append(operation)
    if sources.intersection(targets):
        raise CutoverError("source/target collision across batch")
    aliases: dict[str, str] = {}
    for path in sorted(sources.union(targets)):
        key = "/".join(unicodedata.normalize("NFC", part).casefold().rstrip(" .") for part in path.split("/"))
        previous = aliases.setdefault(key, path)
        if previous != path:
            raise CutoverError(f"portable path alias/case collision: {previous} and {path}")
    return normalized


def _read_matching_record(directory_fd: int, name: str, payload: bytes) -> bool:
    """Return true only for an existing, single-link, byte-identical record."""

    try:
        descriptor = os.open(
            name,
            os.O_RDONLY | os.O_CLOEXEC | os.O_NOFOLLOW,
            dir_fd=directory_fd,
        )
    except FileNotFoundError:
        return False
    try:
        meta = os.fstat(descriptor)
        if not stat.S_ISREG(meta.st_mode) or int(meta.st_nlink) != 1 or int(meta.st_size) != len(payload):
            raise CutoverError(f"existing transaction record identity rejected: {name}")
        existing = bytearray()
        while len(existing) <= len(payload):
            chunk = os.read(descriptor, len(payload) + 1 - len(existing))
            if not chunk:
                break
            existing.extend(chunk)
        if bytes(existing) != payload:
            raise CutoverError(f"existing transaction record differs: {name}")
    finally:
        os.close(descriptor)
    return True


def write_record(
    directory_fd: int,
    name: str,
    value: dict[str, Any],
    *,
    failure_mode: str | None = None,
) -> str:
    """Durably publish one immutable record without exposing partial final bytes."""

    if re.fullmatch(r"[a-z0-9][a-z0-9.-]{0,95}\.json", name) is None:
        raise CutoverError(f"transaction record name rejected: {name}")
    payload = canonical_json(value)
    digest = hashlib.sha256(payload).hexdigest()
    if _read_matching_record(directory_fd, name, payload):
        return digest

    staging_name = f".{name}.pending"
    try:
        descriptor = os.open(
            staging_name,
            os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_CLOEXEC | os.O_NOFOLLOW,
            0o600,
            dir_fd=directory_fd,
        )
    except FileExistsError:
        try:
            if _read_matching_record(directory_fd, staging_name, payload):
                descriptor = -1
            else:  # pragma: no cover - helper either returns true or raises
                raise CutoverError(f"pending transaction record vanished: {name}")
        except CutoverError:
            # A process crash can leave only a partial task-owned staging leaf.
            # It was never authority. Remove exactly that no-follow leaf and
            # recreate it; the canonical final name remains untouched.
            try:
                os.unlink(staging_name, dir_fd=directory_fd)
                os.fsync(directory_fd)
            except OSError as problem:
                raise CutoverError(f"partial transaction staging cleanup failed: {name}: {problem}") from problem
            descriptor = os.open(
                staging_name,
                os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_CLOEXEC | os.O_NOFOLLOW,
                0o600,
                dir_fd=directory_fd,
            )
    preserve_staging = False
    try:
        if descriptor >= 0:
            written = 0
            while written < len(payload):
                count = os.write(descriptor, payload[written:])
                if count <= 0:
                    raise CutoverError(f"short transaction record write: {name}")
                written += count
            os.fsync(descriptor)
            meta = os.fstat(descriptor)
            if not stat.S_ISREG(meta.st_mode) or int(meta.st_nlink) != 1 or int(meta.st_size) != len(payload):
                raise CutoverError(f"transaction staging identity rejected: {name}")
    finally:
        if descriptor >= 0:
            os.close(descriptor)

    if failure_mode == "crash-after-record-stage-fsync":
        preserve_staging = True
        raise SimulatedProcessCrash(staging_name)
    if failure_mode == "record-destination-race":
        racer = os.open(
            name,
            os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_CLOEXEC | os.O_NOFOLLOW,
            0o600,
            dir_fd=directory_fd,
        )
        try:
            os.write(racer, b"racer")
            os.fsync(racer)
        finally:
            os.close(racer)
        os.fsync(directory_fd)
    try:
        rename_noreplace(directory_fd, staging_name, directory_fd, name)
    except CutoverError as problem:
        if "atomic no-replace destination exists" not in str(problem):
            raise
        try:
            identical_final = _read_matching_record(directory_fd, name, payload)
        except CutoverError as mismatch:
            raise CutoverError(f"existing transaction record differs: {name}") from mismatch
        if not identical_final:
            raise CutoverError(f"existing transaction record differs: {name}") from problem
    finally:
        if not preserve_staging:
            try:
                os.unlink(staging_name, dir_fd=directory_fd)
            except FileNotFoundError:
                pass
    os.fsync(directory_fd)
    return digest


def read_optional(repo: Repo, relative: str) -> bytes | None:
    try:
        return repo.read(relative, MAX_SOURCE_BYTES)
    except MissingPathError:
        return None


def missing_target_directories(repo: Repo, operations: list[dict[str, Any]]) -> list[str]:
    """Plan every missing target directory without mutating the tree."""

    missing: set[str] = set()
    for operation in operations:
        if operation["action"] != "move":
            continue
        parent = PurePosixPath(operation["target"]).parent
        parts = list(parent.parts) if parent.as_posix() != "." else []
        first_missing = False
        for index in range(len(parts)):
            relative = "/".join(parts[: index + 1])
            if first_missing:
                missing.add(relative)
                continue
            try:
                descriptor = repo.open_directory(relative)
            except MissingPathError:
                first_missing = True
                missing.add(relative)
            else:
                os.close(descriptor)
    return sorted(missing, key=lambda value: (value.count("/"), value))


def stage_planned_directory(
    transaction_fd: int,
    transaction_relative: str,
    index: int,
    relative: str,
) -> dict[str, Any]:
    """Create an owned directory before recording intent or touching target."""

    stage = f"mkdir-stage-{index:02d}"
    token = secrets.token_hex(16)
    os.mkdir(stage, 0o700, dir_fd=transaction_fd)
    stage_fd = os.open(stage, os.O_RDONLY | os.O_CLOEXEC | os.O_DIRECTORY | os.O_NOFOLLOW, dir_fd=transaction_fd)
    try:
        meta = os.fstat(stage_fd)
        if is_link(meta) or not stat.S_ISDIR(meta.st_mode):
            raise CutoverError(f"staged directory identity rejected: {relative}")
        marker = {
            "owner_token": token,
            "path": relative,
            "transaction": transaction_relative,
        }
        marker_payload = canonical_json(marker)
        marker_fd = os.open(
            DIRECTORY_OWNER_MARKER,
            os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_CLOEXEC | os.O_NOFOLLOW,
            0o600,
            dir_fd=stage_fd,
        )
        try:
            written = 0
            while written < len(marker_payload):
                count = os.write(marker_fd, marker_payload[written:])
                if count <= 0:
                    raise CutoverError("short directory owner-marker write")
                written += count
            os.fsync(marker_fd)
        finally:
            os.close(marker_fd)
        os.fsync(stage_fd)
        identity = [int(meta.st_dev), int(meta.st_ino)]
    finally:
        os.close(stage_fd)
    os.fsync(transaction_fd)
    return {
        "identity": identity,
        "owner_token": token,
        "path": relative,
        "stage": stage,
        "state": "mkdir-intent",
    }


def apply_staged_directory(repo: Repo, transaction_fd: int, record: dict[str, Any]) -> None:
    parent_fd, leaf = repo.open_parent(record["path"])
    try:
        rename_noreplace(transaction_fd, record["stage"], parent_fd, leaf)
        os.fsync(transaction_fd)
        os.fsync(parent_fd)
    finally:
        os.close(parent_fd)


def expected_owner_marker(record: dict[str, Any], transaction_relative: str) -> bytes:
    return canonical_json({
        "owner_token": record["owner_token"],
        "path": record["path"],
        "transaction": transaction_relative,
    })


def remove_owner_marker(repo: Repo, directory: str, record: dict[str, Any], transaction_relative: str) -> None:
    marker_path = f"{directory}/{DIRECTORY_OWNER_MARKER}"
    payload = read_optional(repo, marker_path)
    if payload is None:
        return
    if payload != expected_owner_marker(record, transaction_relative):
        raise CutoverError(f"directory owner marker mismatch: {directory}")
    directory_fd = repo.open_directory(directory)
    try:
        os.unlink(DIRECTORY_OWNER_MARKER, dir_fd=directory_fd)
        os.fsync(directory_fd)
    finally:
        os.close(directory_fd)


def rollback_one(
    repo: Repo,
    transaction_relative: str,
    operation: dict[str, Any],
    *,
    inject_source_race: bool = False,
) -> bool:
    """Reverse one possibly-applied intent without ever replacing a source."""

    source = operation["source"]
    other = operation.get("target") if operation["action"] == "move" else f"{transaction_relative}/{operation['backup']}"
    source_payload = read_optional(repo, source)
    other_payload = read_optional(repo, str(other))
    expected_sha = operation["source_sha256"]
    if source_payload is not None:
        if hashlib.sha256(source_payload).hexdigest() != expected_sha:
            raise CutoverError(f"rollback source collision/content mismatch: {source}")
        if other_payload is not None:
            raise CutoverError(f"rollback refuses to overwrite reappeared source: {source}")
        return False
    if other_payload is None or hashlib.sha256(other_payload).hexdigest() != expected_sha:
        raise CutoverError(f"rollback before-image missing/content mismatch: {source}")
    source_fd, source_leaf = repo.open_parent(source)
    other_fd, other_leaf = repo.open_parent(str(other))
    try:
        try:
            os.stat(source_leaf, dir_fd=source_fd, follow_symlinks=False)
        except FileNotFoundError:
            pass
        else:
            raise CutoverError(f"rollback source reappeared before rename: {source}")
        if inject_source_race:
            collision = os.open(
                source_leaf,
                os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_CLOEXEC | os.O_NOFOLLOW,
                0o600,
                dir_fd=source_fd,
            )
            os.close(collision)
            os.fsync(source_fd)
        rename_noreplace(other_fd, other_leaf, source_fd, source_leaf)
        os.fsync(other_fd)
        os.fsync(source_fd)
    finally:
        os.close(other_fd)
        os.close(source_fd)
    return True


def cleanup_created_directories(
    repo: Repo,
    transaction_relative: str,
    transaction_fd: int,
    directories: list[dict[str, Any]],
) -> None:
    failures: list[str] = []
    for record in sorted(
        directories,
        key=lambda value: (str(value.get("path", "")).count("/"), str(value.get("path", ""))),
        reverse=True,
    ):
        if record.get("_cleaned") is True:
            continue
        relative = canonical_path(record.get("path"), "rollback created directory")
        index = record.get("index")
        try:
            write_record(transaction_fd, f"rmdir-{index:02d}-intent.json", {
                "identity": record["identity"],
                "path": relative,
                "state": "rmdir-intent",
            })
            parent_fd, leaf = repo.open_parent(relative)
            try:
                meta = os.stat(leaf, dir_fd=parent_fd, follow_symlinks=False)
                if is_link(meta) or not stat.S_ISDIR(meta.st_mode) or [int(meta.st_dev), int(meta.st_ino)] != record.get("identity"):
                    raise CutoverError(f"rollback directory identity mismatch: {relative}")
                remove_owner_marker(repo, relative, record, transaction_relative)
                os.rmdir(leaf, dir_fd=parent_fd)
                os.fsync(parent_fd)
            finally:
                os.close(parent_fd)
            write_record(transaction_fd, f"rmdir-{index:02d}-done.json", {
                "identity": record["identity"],
                "path": relative,
                "state": "rmdir-done",
            })
        except Exception as problem:
            failures.append(f"{relative}:{problem.__class__.__name__}")
    if failures:
        raise CutoverError(f"rollback directory cleanup incomplete: {failures}")


def read_canonical_record(repo: Repo, relative: str) -> dict[str, Any] | None:
    payload = read_optional(repo, relative)
    if payload is None:
        return None
    value = json.loads(payload.decode("utf-8"))
    if not isinstance(value, dict) or canonical_json(value) != payload:
        raise CutoverError(f"transaction record is not canonical: {relative}")
    return value


def recover_created_directories(
    repo: Repo,
    transaction_relative: str,
    transaction_fd: int,
    directories: list[str],
) -> list[dict[str, Any]]:
    """Converge staged/applied/cleanup receipt states without guessing ownership."""

    recovered: list[dict[str, Any]] = []
    for index, relative in enumerate(directories):
        canonical = canonical_path(relative, f"journal directory {index}")
        intent = read_canonical_record(repo, f"{transaction_relative}/mkdir-{index:02d}-intent.json")
        applied = read_canonical_record(repo, f"{transaction_relative}/mkdir-{index:02d}-applied.json")
        if intent is None:
            if applied is not None:
                raise CutoverError(f"mkdir applied receipt lacks durable intent: {canonical}")
            continue
        expected_intent_keys = {"identity", "owner_token", "path", "stage", "state"}
        if (
            set(intent) != expected_intent_keys
            or intent.get("path") != canonical
            or intent.get("stage") != f"mkdir-stage-{index:02d}"
            or intent.get("state") != "mkdir-intent"
            or not isinstance(intent.get("owner_token"), str)
            or re.fullmatch(r"[0-9a-f]{32}", intent["owner_token"]) is None
            or not isinstance(intent.get("identity"), list)
            or len(intent["identity"]) != 2
            or not all(isinstance(value, int) and value >= 0 for value in intent["identity"])
        ):
            raise CutoverError(f"mkdir intent receipt mismatch: {canonical}")
        expected_applied = {
            "identity": intent["identity"],
            "owner_token": intent["owner_token"],
            "path": canonical,
            "stage": intent["stage"],
            "state": "mkdir-applied",
        }
        if applied is not None and applied != expected_applied:
            raise CutoverError(f"mkdir applied receipt mismatch: {canonical}")
        expected_stage_cleanup_intent = {
            "identity": intent["identity"], "path": canonical, "state": "stage-cleanup-intent",
        }
        expected_stage_cleanup_done = {
            "identity": intent["identity"], "path": canonical, "state": "stage-cleanup-done",
        }
        stage_cleanup_intent = read_canonical_record(
            repo, f"{transaction_relative}/stage-{index:02d}-cleanup-intent.json"
        )
        stage_cleanup_done = read_canonical_record(
            repo, f"{transaction_relative}/stage-{index:02d}-cleanup-done.json"
        )
        if stage_cleanup_intent is not None and stage_cleanup_intent != expected_stage_cleanup_intent:
            raise CutoverError(f"mkdir stage cleanup intent mismatch: {canonical}")
        if stage_cleanup_done is not None and stage_cleanup_done != expected_stage_cleanup_done:
            raise CutoverError(f"mkdir stage cleanup receipt mismatch: {canonical}")
        stage_relative = f"{transaction_relative}/{intent['stage']}"
        stage_fd = -1
        target_fd = -1
        try:
            try:
                stage_fd = repo.open_directory(stage_relative)
            except MissingPathError:
                pass
            try:
                target_fd = repo.open_directory(canonical)
            except MissingPathError:
                pass
            if stage_fd >= 0 and target_fd >= 0:
                raise CutoverError(f"mkdir stage and target both exist: {canonical}")
            if stage_fd >= 0:
                if applied is not None:
                    raise CutoverError(f"mkdir applied receipt still names staging directory: {canonical}")
                if stage_cleanup_done is not None:
                    raise CutoverError(f"mkdir stage exists after cleanup receipt: {canonical}")
                meta = os.fstat(stage_fd)
                if [int(meta.st_dev), int(meta.st_ino)] != intent["identity"]:
                    raise CutoverError(f"mkdir stage identity changed: {canonical}")
                marker = read_optional(repo, f"{stage_relative}/{DIRECTORY_OWNER_MARKER}")
                if marker is not None and marker != expected_owner_marker(intent, transaction_relative):
                    raise CutoverError(f"mkdir stage marker mismatch: {canonical}")
                if stage_cleanup_intent is None:
                    if marker is None:
                        raise CutoverError(f"mkdir stage lacks ownership marker before cleanup intent: {canonical}")
                    write_record(
                        transaction_fd,
                        f"stage-{index:02d}-cleanup-intent.json",
                        expected_stage_cleanup_intent,
                    )
                if marker is not None:
                    os.unlink(DIRECTORY_OWNER_MARKER, dir_fd=stage_fd)
                    os.fsync(stage_fd)
                os.close(stage_fd)
                stage_fd = -1
                os.rmdir(intent["stage"], dir_fd=transaction_fd)
                os.fsync(transaction_fd)
                write_record(
                    transaction_fd,
                    f"stage-{index:02d}-cleanup-done.json",
                    expected_stage_cleanup_done,
                )
                continue
            if target_fd >= 0:
                if stage_cleanup_intent is not None or stage_cleanup_done is not None:
                    raise CutoverError(f"mkdir target conflicts with stage cleanup receipts: {canonical}")
                meta = os.fstat(target_fd)
                if [int(meta.st_dev), int(meta.st_ino)] != intent["identity"]:
                    raise CutoverError(f"mkdir target identity changed: {canonical}")
                if applied is None:
                    marker = repo.read(f"{canonical}/{DIRECTORY_OWNER_MARKER}", 4096)
                    if marker != expected_owner_marker(intent, transaction_relative):
                        raise CutoverError(f"mkdir target lacks intent ownership proof: {canonical}")
                    applied = expected_applied
                    write_record(transaction_fd, f"mkdir-{index:02d}-applied.json", applied)
                recovered.append({**applied, "index": index})
                continue
            rmdir_intent = read_canonical_record(repo, f"{transaction_relative}/rmdir-{index:02d}-intent.json")
            rmdir_done = read_canonical_record(repo, f"{transaction_relative}/rmdir-{index:02d}-done.json")
            if applied is not None:
                expected_rmdir_intent = {"identity": intent["identity"], "path": canonical, "state": "rmdir-intent"}
                expected_rmdir_done = {"identity": intent["identity"], "path": canonical, "state": "rmdir-done"}
                if rmdir_intent != expected_rmdir_intent or (rmdir_done is not None and rmdir_done != expected_rmdir_done):
                    raise CutoverError(f"applied mkdir disappeared without cleanup intent: {canonical}")
                write_record(transaction_fd, f"rmdir-{index:02d}-done.json", expected_rmdir_done)
                recovered.append({**applied, "index": index, "_cleaned": True})
            elif rmdir_intent is not None or rmdir_done is not None:
                raise CutoverError(f"unapplied mkdir has cleanup receipt: {canonical}")
            elif stage_cleanup_intent is not None:
                if stage_cleanup_done is None:
                    write_record(
                        transaction_fd,
                        f"stage-{index:02d}-cleanup-done.json",
                        expected_stage_cleanup_done,
                    )
            elif stage_cleanup_done is not None:
                raise CutoverError(f"mkdir stage cleanup done lacks intent: {canonical}")
        finally:
            if stage_fd >= 0:
                os.close(stage_fd)
            if target_fd >= 0:
                os.close(target_fd)
    return recovered


def rollback_intents(
    repo: Repo,
    transaction_relative: str,
    transaction_fd: int,
    operations: list[dict[str, Any]],
    directories: list[dict[str, Any]],
    *,
    receipt: bool,
    inject_source_race: bool = False,
    failure_mode: str | None = None,
) -> None:
    write_record(transaction_fd, "rollback-intent.json", {
        "operations_sha256": hashlib.sha256(canonical_json(operations)).hexdigest(),
        "state": "rollback-intent",
    })
    for index, operation in reversed(list(enumerate(operations))):
        write_record(transaction_fd, f"rollback-{index:02d}-intent.json", {
            "operation": operation,
            "state": "rollback-intent",
        })
        rollback_one(
            repo,
            transaction_relative,
            operation,
            inject_source_race=inject_source_race and index == 0,
        )
        write_record(transaction_fd, f"rollback-{index:02d}-done.json", {
            "source": operation["source"],
            "state": "rollback-done",
        })
    cleanup_created_directories(repo, transaction_relative, transaction_fd, directories)
    if failure_mode == "crash-after-cleanup-before-receipt":
        raise SimulatedProcessCrash(transaction_relative)
    if receipt:
        stable_directories = [
            {key: value for key, value in record.items() if not key.startswith("_")}
            for record in directories
        ]
        write_record(transaction_fd, "rolled-back.json", {
            "directories_sha256": hashlib.sha256(canonical_json(stable_directories)).hexdigest(),
            "operations": len(operations),
            "operations_sha256": hashlib.sha256(canonical_json(operations)).hexdigest(),
            "state": "rolled-back",
        })


def remove_failed_transaction(audit_fd: int, transaction_fd: int, transaction_name: str) -> None:
    """Remove only a fully rolled-back task-owned journal; any error is fatal."""

    for name in os.listdir(transaction_fd):
        meta = os.stat(name, dir_fd=transaction_fd, follow_symlinks=False)
        if is_link(meta) or not stat.S_ISREG(meta.st_mode) or int(meta.st_nlink) != 1:
            raise CutoverError(f"unexpected transaction cleanup entry: {name}")
        os.unlink(name, dir_fd=transaction_fd)
    os.fsync(transaction_fd)
    os.close(transaction_fd)
    os.rmdir(transaction_name, dir_fd=audit_fd)
    os.fsync(audit_fd)


def apply_transaction(
    repo: Repo,
    batch: dict[str, Any],
    operations: list[dict[str, Any]],
    manifest_payload: bytes,
    *,
    failure_mode: str | None = None,
) -> str:
    manifest_snapshot = json.loads(manifest_payload.decode("utf-8"))
    if not isinstance(manifest_snapshot, dict) or canonical_json(manifest_snapshot) != manifest_payload:
        raise CutoverError("transaction manifest snapshot must be a canonical JSON object")
    manifest_sha256 = hashlib.sha256(manifest_payload).hexdigest()
    directories = missing_target_directories(repo, operations)
    journal_operations = []
    for index, operation in enumerate(operations):
        item = {key: value for key, value in operation.items() if not key.startswith("_")}
        if item["action"] == "delete":
            item["backup"] = f"delete-{index:02d}-{hashlib.sha256(item['source'].encode()).hexdigest()[:16]}"
        journal_operations.append(item)
    transaction_name = f".cutover-txn-{batch['id']}-{secrets.token_hex(8)}"
    transaction_relative = f"{TRANSACTION_PARENT}/{transaction_name}"
    audit_fd = repo.open_directory(TRANSACTION_PARENT)
    transaction_fd = -1
    intended: list[dict[str, Any]] = []
    created: list[dict[str, Any]] = []
    try:
        os.mkdir(transaction_name, 0o700, dir_fd=audit_fd)
        os.fsync(audit_fd)
        transaction_fd = os.open(transaction_name, repo.dir_flags, dir_fd=audit_fd)
        sealed_manifest_sha = write_record(transaction_fd, "manifest.json", manifest_snapshot)
        if sealed_manifest_sha != manifest_sha256:
            raise CutoverError("durable transaction manifest digest mismatch")
        journal = {
            "batch_id": batch["id"],
            "directories": directories,
            "manifest_sha256": manifest_sha256,
            "operations": journal_operations,
            "repository_head": repo.head(),
            "schema_version": 1,
            "state": "prepared",
        }
        journal_sha = write_record(transaction_fd, "journal.json", journal)
        for index, relative in enumerate(directories):
            created_record = stage_planned_directory(
                transaction_fd,
                transaction_relative,
                index,
                relative,
            )
            write_record(transaction_fd, f"mkdir-{index:02d}-intent.json", created_record)
            if failure_mode == "crash-after-mkdir-intent" and index == 0:
                raise SimulatedProcessCrash(transaction_relative)
            apply_staged_directory(repo, transaction_fd, created_record)
            if failure_mode == "crash-after-mkdir-before-applied" and index == 0:
                raise SimulatedProcessCrash(transaction_relative)
            applied_record = {**created_record, "state": "mkdir-applied", "index": index}
            created.append(applied_record)
            write_record(
                transaction_fd,
                f"mkdir-{index:02d}-applied.json",
                {key: value for key, value in applied_record.items() if key != "index"},
            )
            remove_owner_marker(repo, relative, applied_record, transaction_relative)
            if failure_mode == "crash-after-mkdir-applied" and index == 0:
                raise SimulatedProcessCrash(transaction_relative)
        for index, operation in enumerate(journal_operations):
            write_record(transaction_fd, f"operation-{index:02d}-intent.json", {
                "operation": operation,
                "state": "rename-intent",
            })
            intended.append(operation)
            source_fd, source_leaf = repo.open_parent(operation["source"])
            try:
                current = os.stat(source_leaf, dir_fd=source_fd, follow_symlinks=False)
                original = operations[index]
                if (
                    is_link(current)
                    or not stat.S_ISREG(current.st_mode)
                    or int(current.st_nlink) != 1
                    or signature(current, content=True) != tuple(original["_source_signature"])
                ):
                    raise CutoverError(f"source changed before rename: {operation['source']}")
                other = operation.get("target") if operation["action"] == "move" else f"{transaction_relative}/{operation['backup']}"
                other_fd, other_leaf = repo.open_parent(str(other))
                try:
                    try:
                        os.stat(other_leaf, dir_fd=other_fd, follow_symlinks=False)
                    except FileNotFoundError:
                        pass
                    else:
                        raise CutoverError(f"operation destination appeared: {other}")
                    if failure_mode == "destination-race" and index == 0:
                        racer = os.open(
                            other_leaf,
                            os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_CLOEXEC | os.O_NOFOLLOW,
                            0o600,
                            dir_fd=other_fd,
                        )
                        os.write(racer, b"racer")
                        os.fsync(racer)
                        os.close(racer)
                        os.fsync(other_fd)
                    rename_noreplace(source_fd, source_leaf, other_fd, other_leaf)
                    os.fsync(source_fd)
                    os.fsync(other_fd)
                finally:
                    os.close(other_fd)
            finally:
                os.close(source_fd)
            if failure_mode == "after-rename-before-done" and index == 0:
                raise CutoverError("forced crash-window sentinel")
            if failure_mode == "rollback-source-race" and index == 0:
                raise CutoverError("forced atomic rollback-source-race sentinel")
            if failure_mode == "source-collision" and index == 0:
                source_fd, source_leaf = repo.open_parent(operation["source"])
                try:
                    collision_fd = os.open(
                        source_leaf,
                        os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_CLOEXEC | os.O_NOFOLLOW,
                        0o600,
                        dir_fd=source_fd,
                    )
                    os.close(collision_fd)
                    os.fsync(source_fd)
                finally:
                    os.close(source_fd)
                raise CutoverError("forced source-collision sentinel")
            if failure_mode == "cleanup-failure" and index == 0 and created:
                nuisance_fd, nuisance_leaf = repo.open_parent(f"{created[-1]['path']}/cleanup-sentinel")
                try:
                    marker_fd = os.open(
                        nuisance_leaf,
                        os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_CLOEXEC | os.O_NOFOLLOW,
                        0o600,
                        dir_fd=nuisance_fd,
                    )
                    os.close(marker_fd)
                    os.fsync(nuisance_fd)
                finally:
                    os.close(nuisance_fd)
                raise CutoverError("forced cleanup-failure sentinel")
            write_record(transaction_fd, f"operation-{index:02d}-done.json", {
                "source": operation["source"],
                "state": "rename-done",
            })
        write_record(transaction_fd, "applied.json", {
            "batch_id": batch["id"],
            "directories": directories,
            "journal_sha256": journal_sha,
            "manifest_sha256": manifest_sha256,
            "operations_sha256": hashlib.sha256(canonical_json(journal_operations)).hexdigest(),
            "repository_head": journal["repository_head"],
            "state": "applied-awaiting-atomic-commit",
        })
        os.fsync(audit_fd)
        return transaction_relative
    except Exception as original:
        if transaction_fd >= 0:
            try:
                recovered_directories = recover_created_directories(
                    repo,
                    transaction_relative,
                    transaction_fd,
                    directories,
                )
                rollback_intents(
                    repo,
                    transaction_relative,
                    transaction_fd,
                    intended,
                    recovered_directories,
                    receipt=False,
                    inject_source_race=failure_mode == "rollback-source-race",
                )
                remove_failed_transaction(audit_fd, transaction_fd, transaction_name)
                transaction_fd = -1
            except Exception as rollback_problem:
                raise CutoverError(
                    f"rollback incomplete after {original.__class__.__name__}: "
                    f"{rollback_problem.__class__.__name__}: {rollback_problem}"
                ) from original
        raise
    finally:
        if transaction_fd >= 0:
            os.close(transaction_fd)
        os.close(audit_fd)


def validate_rolled_back_state(
    repo: Repo,
    transaction_relative: str,
    journal: dict[str, Any],
    receipt: dict[str, Any],
) -> None:
    operations = journal["operations"]
    restored_directories: list[dict[str, Any]] = []
    for operation in operations:
        source = read_optional(repo, operation["source"])
        other = operation.get("target") if operation["action"] == "move" else f"{transaction_relative}/{operation['backup']}"
        if source is None or hashlib.sha256(source).hexdigest() != operation["source_sha256"] or read_optional(repo, str(other)) is not None:
            raise CutoverError(f"rolled-back receipt filesystem state mismatch: {operation['source']}")
    for index, relative in enumerate(journal["directories"]):
        intent = read_canonical_record(repo, f"{transaction_relative}/mkdir-{index:02d}-intent.json")
        applied = read_canonical_record(repo, f"{transaction_relative}/mkdir-{index:02d}-applied.json")
        if intent is None:
            continue
        for candidate in (relative, f"{transaction_relative}/{intent.get('stage')}"):
            try:
                descriptor = repo.open_directory(str(candidate))
            except MissingPathError:
                pass
            else:
                os.close(descriptor)
                raise CutoverError(f"rolled-back receipt left directory state: {candidate}")
        if applied is not None:
            record = {**applied, "index": index}
            restored_directories.append(record)
            done = read_canonical_record(repo, f"{transaction_relative}/rmdir-{index:02d}-done.json")
            expected_done = {"identity": applied.get("identity"), "path": relative, "state": "rmdir-done"}
            if done != expected_done:
                raise CutoverError(f"rolled-back receipt lacks directory cleanup proof: {relative}")
    expected = {
        "directories_sha256": hashlib.sha256(canonical_json(restored_directories)).hexdigest(),
        "operations": len(operations),
        "operations_sha256": hashlib.sha256(canonical_json(operations)).hexdigest(),
        "state": "rolled-back",
    }
    if receipt != expected:
        raise CutoverError("rolled-back receipt binding mismatch")


def rollback_transaction(
    repo: Repo,
    transaction_relative: str,
    *,
    self_test: bool = False,
    failure_mode: str | None = None,
) -> None:
    path = canonical_path(transaction_relative, "rollback transaction")
    prefix = f"{TRANSACTION_PARENT}/.cutover-txn-"
    if not path.startswith(prefix):
        raise CutoverError("rollback accepts only a printed Phase 128 transaction path")
    journal_payload = repo.read(f"{path}/journal.json", MAX_MANIFEST_BYTES)
    manifest_payload = repo.read(f"{path}/manifest.json", MAX_MANIFEST_BYTES)
    applied_payload = read_optional(repo, f"{path}/applied.json")
    rolled_back_payload = read_optional(repo, f"{path}/rolled-back.json")
    journal = json.loads(journal_payload.decode("utf-8"))
    manifest = json.loads(manifest_payload.decode("utf-8"))
    applied = json.loads(applied_payload.decode("utf-8")) if applied_payload is not None else None
    rolled_back = json.loads(rolled_back_payload.decode("utf-8")) if rolled_back_payload is not None else None
    if not isinstance(journal, dict) or not isinstance(manifest, dict) or (applied is not None and not isinstance(applied, dict)) or (rolled_back is not None and not isinstance(rolled_back, dict)):
        raise CutoverError("rollback receipt records must be JSON objects")
    if canonical_json(journal) != journal_payload or (
        applied is not None and canonical_json(applied) != applied_payload
    ) or canonical_json(manifest) != manifest_payload or (
        rolled_back is not None and canonical_json(rolled_back) != rolled_back_payload
    ):
        raise CutoverError("rollback receipt records are not canonical")
    operations = journal.get("operations")
    directories = journal.get("directories")
    if (
        journal.get("schema_version") != 1
        or not isinstance(directories, list)
    ):
        raise CutoverError("rollback receipt binding is invalid")
    try:
        operations = require_literal_operation_budget(operations, "rollback journal")
    except CutoverError as problem:
        raise CutoverError("rollback receipt binding is invalid") from problem
    if any(not isinstance(value, str) for value in directories) or len(directories) != len(set(directories)):
        raise CutoverError("rollback directory receipt contains duplicates/non-strings")
    move_targets = [item.get("target") for item in operations if isinstance(item, dict) and item.get("action") == "move"]
    for relative in directories:
        canonical = canonical_path(relative, "rollback directory receipt")
        if not any(isinstance(target, str) and target.startswith(f"{canonical}/") for target in move_targets):
            raise CutoverError(f"rollback directory is not an ancestor of a sealed move target: {canonical}")
    manifest_sha256 = hashlib.sha256(manifest_payload).hexdigest()
    operations_sha256 = hashlib.sha256(canonical_json(operations)).hexdigest()
    if journal.get("manifest_sha256") != manifest_sha256:
        raise CutoverError("rollback journal is not bound to its durable manifest snapshot")
    if applied is not None:
        expected_applied = {
            "batch_id": journal.get("batch_id"),
            "directories": directories,
            "journal_sha256": hashlib.sha256(journal_payload).hexdigest(),
            "manifest_sha256": manifest_sha256,
            "operations_sha256": operations_sha256,
            "repository_head": journal.get("repository_head"),
            "state": "applied-awaiting-atomic-commit",
        }
        if applied != expected_applied:
            raise CutoverError("successful rollback receipt is not exactly bound to journal and manifest")
    if not self_test:
        validation_errors = validate_manifest_data(manifest, repo.root, check_filesystem=False)
        if validation_errors:
            raise CutoverError(f"rollback manifest validation failed: {validation_errors[0]}")
        batches = manifest.get("batches")
        matches = [item for item in batches if isinstance(item, dict) and item.get("id") == journal.get("batch_id")] if isinstance(batches, list) else []
        if len(matches) != 1 or not isinstance(matches[0].get("operations"), list):
            raise CutoverError("rollback batch is not uniquely bound to the manifest")
        expected_operations: list[dict[str, Any]] = []
        for index, raw in enumerate(matches[0]["operations"]):
            item = {key: raw[key] for key in ("action", "source", "source_sha256")}
            if raw["action"] == "move":
                item["target"] = raw["target"]
            else:
                item["backup"] = f"delete-{index:02d}-{hashlib.sha256(raw['source'].encode()).hexdigest()[:16]}"
            expected_operations.append(item)
        if operations != expected_operations:
            raise CutoverError("rollback journal operations differ from the independently validated manifest")
    transaction_fd = repo.open_directory(path)
    try:
        if rolled_back is not None:
            validate_rolled_back_state(repo, path, journal, rolled_back)
            return
        created_directories = recover_created_directories(repo, path, transaction_fd, directories)
        rollback_intents(
            repo,
            path,
            transaction_fd,
            operations,
            created_directories,
            receipt=True,
            failure_mode=failure_mode,
        )
    finally:
        os.close(transaction_fd)


def run_self_test() -> list[str]:
    failures: list[str] = []
    if MAX_BATCH_OPERATIONS != 56:
        failures.append("operational batch cap differs from canonical max 56")
    try:
        require_literal_operation_budget([None] * MAX_BATCH_OPERATIONS, "self-test batch")
    except CutoverError:
        failures.append("canonical 56-operation batch was rejected")
    try:
        require_literal_operation_budget(
            [None] * (MAX_BATCH_OPERATIONS + 1), "self-test oversized batch"
        )
    except CutoverError:
        pass
    else:
        failures.append("57-operation batch exceeded canonical cap without rejection")
    for value in (
        "../escape",
        "pro-qa-ai/secret",
        "C:/outside",
        "/".join(["a"] * (MAX_PATH_COMPONENTS + 1)),
    ):
        try:
            canonical_path(value, "sentinel")
        except CutoverError:
            pass
        else:
            failures.append(f"unsafe path sentinel accepted: {value}")
    with tempfile.TemporaryDirectory(prefix="hub-apply-cutover-") as temporary:
        publication = Path(temporary) / "record-publication"
        publication.mkdir()
        publication_fd = os.open(publication, os.O_RDONLY | os.O_DIRECTORY | os.O_CLOEXEC)
        try:
            record = {"schema_version": 1, "sentinel": "durable-publication"}
            expected_payload = canonical_json(record)
            first_digest = write_record(publication_fd, "idempotent.json", record)
            second_digest = write_record(publication_fd, "idempotent.json", record)
            if first_digest != second_digest or (publication / "idempotent.json").read_bytes() != expected_payload:
                failures.append("record publication idempotence sentinel changed canonical bytes")
            try:
                write_record(
                    publication_fd,
                    "destination-race.json",
                    record,
                    failure_mode="record-destination-race",
                )
            except CutoverError as problem:
                if "existing transaction record differs" not in str(problem):
                    failures.append("record destination-race sentinel returned wrong failure")
            else:
                failures.append("record destination-race sentinel overwrote an appearing final")
            if (publication / "destination-race.json").read_bytes() != b"racer":
                failures.append("record destination-race sentinel did not preserve racer bytes")
            try:
                write_record(
                    publication_fd,
                    "crash-publication.json",
                    record,
                    failure_mode="crash-after-record-stage-fsync",
                )
            except SimulatedProcessCrash:
                pass
            else:
                failures.append("record crash-publication sentinel did not interrupt")
            pending = publication / ".crash-publication.json.pending"
            if (publication / "crash-publication.json").exists() or pending.read_bytes() != expected_payload:
                failures.append("record crash-publication exposed partial final bytes")
            recovered_digest = write_record(publication_fd, "crash-publication.json", record)
            if recovered_digest != first_digest or pending.exists() or (publication / "crash-publication.json").read_bytes() != expected_payload:
                failures.append("record crash-publication did not recover idempotently")
        finally:
            os.close(publication_fd)

        root = Path(temporary) / "crash-window"
        (root / ".git" / "refs" / "heads").mkdir(parents=True)
        (root / ".planning" / "audits" / "test-rebase").mkdir(parents=True)
        (root / "src").mkdir()
        (root / ".git" / "HEAD").write_text("ref: refs/heads/main\n", encoding="ascii")
        (root / ".git" / "refs" / "heads" / "main").write_text("1" * 40 + "\n", encoding="ascii")
        (root / "src" / "one.txt").write_bytes(b"one")
        (root / "src" / "two.txt").write_bytes(b"two")
        repo = Repo(root)
        try:
            operations = [
                {"action": "move", "source": "src/one.txt", "source_sha256": hashlib.sha256(b"one").hexdigest(), "target": "dst/one.txt"},
                {"action": "delete", "source": "src/two.txt", "source_sha256": hashlib.sha256(b"two").hexdigest()},
            ]
            checked = validate_operations(repo, {"operations": operations})
            try:
                apply_transaction(
                    repo,
                    {"id": "self-test"},
                    checked,
                    canonical_json({"self_test": "crash-window"}),
                    failure_mode="after-rename-before-done",
                )
            except CutoverError as problem:
                if "forced crash-window" not in str(problem):
                    failures.append("crash-window rollback returned wrong failure")
            else:
                failures.append("crash-window rollback sentinel did not fail")
            if not (root / "src" / "one.txt").is_file() or (root / "dst" / "one.txt").exists():
                failures.append("crash-window rollback did not restore source/destination")
            os.link(root / "src" / "one.txt", root / "src" / "hardlink.txt")
            try:
                validate_operations(repo, {"operations": [operations[0]]})
            except CutoverError as problem:
                if "single-link" not in str(problem):
                    failures.append("hardlink sentinel returned wrong failure")
            else:
                failures.append("hardlink sentinel accepted")
        finally:
            repo.close()
        for mode, expected_fragment in (
            ("destination-race", "rollback incomplete"),
            ("rollback-source-race", "rollback incomplete"),
            ("source-collision", "rollback incomplete"),
            ("cleanup-failure", "directory cleanup incomplete"),
        ):
            fixture = Path(temporary) / mode
            (fixture / ".git" / "refs" / "heads").mkdir(parents=True)
            (fixture / ".planning" / "audits" / "test-rebase").mkdir(parents=True)
            (fixture / "src").mkdir()
            (fixture / ".git" / "HEAD").write_text("ref: refs/heads/main\n", encoding="ascii")
            (fixture / ".git" / "refs" / "heads" / "main").write_text("1" * 40 + "\n", encoding="ascii")
            (fixture / "src" / "one.txt").write_bytes(b"one")
            candidate = Repo(fixture)
            try:
                operation = [{
                    "action": "move",
                    "source": "src/one.txt",
                    "source_sha256": hashlib.sha256(b"one").hexdigest(),
                    "target": "dst/one.txt",
                }]
                checked = validate_operations(candidate, {"operations": operation})
                try:
                    apply_transaction(
                        candidate,
                        {"id": f"self-test-{mode}"},
                        checked,
                        canonical_json({"self_test": mode}),
                        failure_mode=mode,
                    )
                except CutoverError as problem:
                    if expected_fragment not in str(problem):
                        failures.append(f"{mode} returned wrong failure")
                else:
                    failures.append(f"{mode} sentinel did not fail")
            finally:
                candidate.close()
            if mode == "destination-race":
                if (fixture / "src" / "one.txt").read_bytes() != b"one" or (fixture / "dst" / "one.txt").read_bytes() != b"racer":
                    failures.append("atomic apply race overwrote source or appearing destination")
            if mode == "rollback-source-race":
                if (fixture / "src" / "one.txt").read_bytes() != b"" or (fixture / "dst" / "one.txt").read_bytes() != b"one":
                    failures.append("atomic rollback race overwrote appearing source or lost before-image")
        for crash_mode in (
            "crash-after-mkdir-intent",
            "crash-after-mkdir-before-applied",
            "crash-after-mkdir-applied",
        ):
            fixture = Path(temporary) / crash_mode
            (fixture / ".git" / "refs" / "heads").mkdir(parents=True)
            (fixture / ".planning" / "audits" / "test-rebase").mkdir(parents=True)
            (fixture / "src").mkdir()
            (fixture / ".git" / "HEAD").write_text("ref: refs/heads/main\n", encoding="ascii")
            (fixture / ".git" / "refs" / "heads" / "main").write_text("1" * 40 + "\n", encoding="ascii")
            (fixture / "src" / "one.txt").write_bytes(b"one")
            crashed_repo = Repo(fixture)
            transaction = None
            try:
                operation = [{
                    "action": "move",
                    "source": "src/one.txt",
                    "source_sha256": hashlib.sha256(b"one").hexdigest(),
                    "target": "dst/one.txt",
                }]
                checked = validate_operations(crashed_repo, {"operations": operation})
                try:
                    apply_transaction(
                        crashed_repo,
                        {"id": f"self-test-{crash_mode}"},
                        checked,
                        canonical_json({"self_test": crash_mode}),
                        failure_mode=crash_mode,
                    )
                except SimulatedProcessCrash as problem:
                    transaction = str(problem)
                else:
                    failures.append(f"{crash_mode} did not simulate process loss")
            finally:
                crashed_repo.close()
            if transaction is not None:
                recovery_repo = Repo(fixture)
                try:
                    rollback_transaction(recovery_repo, transaction, self_test=True)
                except CutoverError as problem:
                    failures.append(f"{crash_mode} receipt recovery failed: {problem}")
                finally:
                    recovery_repo.close()
            if not (fixture / "src" / "one.txt").is_file() or (fixture / "dst").exists():
                failures.append(f"{crash_mode} receipt recovery did not restore directory state")
        success = Path(temporary) / "successful-receipt"
        (success / ".git" / "refs" / "heads").mkdir(parents=True)
        (success / ".planning" / "audits" / "test-rebase").mkdir(parents=True)
        (success / "src").mkdir()
        (success / ".git" / "HEAD").write_text("ref: refs/heads/main\n", encoding="ascii")
        (success / ".git" / "refs" / "heads" / "main").write_text("1" * 40 + "\n", encoding="ascii")
        (success / "src" / "one.txt").write_bytes(b"one")
        success_repo = Repo(success)
        try:
            operation = [{
                "action": "move",
                "source": "src/one.txt",
                "source_sha256": hashlib.sha256(b"one").hexdigest(),
                "target": "dst/one.txt",
            }]
            checked = validate_operations(success_repo, {"operations": operation})
            transaction = apply_transaction(
                success_repo,
                {"id": "self-test-success"},
                checked,
                canonical_json({"self_test": "successful-receipt"}),
            )
            try:
                rollback_transaction(
                    success_repo,
                    transaction,
                    self_test=True,
                    failure_mode="crash-after-cleanup-before-receipt",
                )
            except SimulatedProcessCrash:
                pass
            else:
                failures.append("cleanup-before-receipt crash sentinel did not interrupt")
            rollback_transaction(success_repo, transaction, self_test=True)
            rollback_transaction(success_repo, transaction, self_test=True)
            if not (success / "src" / "one.txt").is_file() or (success / "dst" / "one.txt").exists():
                failures.append("successful receipt rollback did not restore exact state")
            if not (success / transaction / "rolled-back.json").is_file():
                failures.append("successful receipt rollback record is missing")
        finally:
            success_repo.close()
    return failures


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo-root", type=Path, default=REPO_ROOT)
    parser.add_argument("--manifest", default=DEFAULT_MANIFEST)
    parser.add_argument("--batch-id")
    actions = parser.add_mutually_exclusive_group(required=True)
    actions.add_argument("--check", action="store_true")
    actions.add_argument("--self-test", action="store_true")
    actions.add_argument("--apply", action="store_true")
    actions.add_argument("--rollback", metavar="TRANSACTION")
    args = parser.parse_args(argv)
    try:
        if args.self_test:
            failures = run_self_test()
            if failures:
                raise CutoverError(failures[0])
            print("PASS: apply guard temp-only no-replace/crash/link/collision/cleanup/receipt sentinels")
            return 0
        if args.rollback:
            expected = os.environ.get("HUB_CUTOVER_ROLLBACK_CONFIRM")
            if expected != args.rollback:
                raise CutoverError(
                    "rollback confirmation mismatch; HUB_CUTOVER_ROLLBACK_CONFIRM "
                    "must equal the exact transaction path"
                )
            repo = Repo(args.repo_root)
            try:
                repo.ensure_identity()
                rollback_transaction(repo, args.rollback)
                repo.ensure_identity()
            finally:
                repo.close()
            print(f"ROLLED_BACK: transaction={args.rollback}")
            return 0
        if not args.batch_id or SLUG_RE.fullmatch(args.batch_id) is None:
            raise CutoverError("--batch-id must be a canonical lowercase slug")
        repo = Repo(args.repo_root)
        try:
            repo.ensure_identity()
            batch, manifest_payload = load_batch(repo, args.manifest, args.batch_id)
            operations = validate_operations(repo, batch)
            repo.ensure_identity()
            if args.check:
                print(f"PASS: literal batch check only; id={args.batch_id} operations={len(operations)}")
                return 0
            if os.environ.get("HUB_CUTOVER_APPLY_CONFIRM") != args.batch_id:
                raise CutoverError("apply confirmation mismatch; HUB_CUTOVER_APPLY_CONFIRM must equal batch ID")
            repo.ensure_identity()
            transaction = apply_transaction(
                repo,
                batch,
                operations,
                manifest_payload,
            )
            print(f"APPLIED: id={args.batch_id} operations={len(operations)} rollback_transaction={transaction}")
            return 0
        finally:
            repo.close()
    except (CutoverError, OSError, UnicodeError, json.JSONDecodeError, KeyError, TypeError) as problem:
        print(f"FAIL: {problem}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
