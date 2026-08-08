#!/usr/bin/env python3
"""Fail-closed validator for the exact Phase 129 cutover manifest.

The validator never executes a command found in JSON and never imports the
builder module.  It independently reconciles the sealed ledgers, literal operations,
resource gates and filesystem state allowed by each batch state.
"""

from __future__ import annotations

import argparse
import copy
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import re
import stat
import sys
import tempfile
from typing import Any
import unicodedata
import zlib


_SEALED_REPO_ROOT = os.environ.get("HUB_PHASE128_REPO_ROOT")
REPO_ROOT = (
    Path(_SEALED_REPO_ROOT)
    if _SEALED_REPO_ROOT is not None
    else Path(__file__).resolve().parents[3]
)
if not REPO_ROOT.is_absolute():
    raise RuntimeError("HUB_PHASE128_REPO_ROOT must be absolute")
MAX_BATCH_OPERATIONS = 64
MAX_FILE_BYTES = 64 * 1024 * 1024
MAX_STDIN_BYTES = 32 * 1024 * 1024
MAX_PATH_COMPONENTS = 64
MAX_COMPONENT_UTF8_BYTES = 255
MAX_PATH_UTF8_BYTES = 4096
PROTECTED_PREFIX = "pro-qa-ai/"
SHA256_RE = re.compile(r"^[0-9a-f]{64}$")
ASCII_SLUG_RE = re.compile(r"^[a-z0-9]+(?:-[a-z0-9]+)*$")
FILE_ATTRIBUTE_REPARSE_POINT = getattr(stat, "FILE_ATTRIBUTE_REPARSE_POINT", 0x400)
HEX_SHA_RE = re.compile(r"^[0-9a-f]{40}(?:[0-9a-f]{24})?$")
TIMESTAMP_RE = re.compile(r"^20[0-9]{2}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$")
DYNAMIC_BATCH_KEYS = {
    "state",
    "state_transition",
    "consumer_transitions",
    "commit",
    "verifier_record",
    "evidence",
    "rollback",
}

AUTHORITY_FILES = {
    "inventory": (".planning/audits/test-rebase/inventory.json", "72de6a81b2bd5da8873a98e784a5069a661e9caea083f10722aa85e9f91b4511"),
    "founder_manifest": (".planning/audits/test-rebase/founder-greenfield-cutover-manifest.json", "df503a7f136aed70f0a9aaed2a576350f692574b0fd9f5506bd4b7f5fed0376a"),
    "architecture_contract": (".planning/audits/test-rebase/test-architecture-contract.json", "62eb27ecd93164cf404639e6c8b85d14843613da2d34ea4da08101a673bf1335"),
}
LEDGER_REVISION = "5708671f00411c7c9e9dbee88b725de2092cbacd"
PYTHON_SCAFFOLD_REQUIRED_FILES = [
    "testframework/pyproject.toml", "testframework/uv.lock",
    "testframework/.python-version", "testframework/README.md",
    "testframework/src/hub_testkit/__init__.py",
    "testframework/src/hub_testkit/config/__init__.py",
    "testframework/src/hub_testkit/config/models.py",
    "testframework/src/hub_testkit/config/load.py",
    "testframework/src/hub_testkit/runtime/__init__.py",
    "testframework/src/hub_testkit/runtime/ids.py",
    "testframework/src/hub_testkit/runtime/ownership.py",
    "testframework/src/hub_testkit/runtime/redaction.py",
    "testframework/src/hub_testkit/traceability/__init__.py",
    "testframework/src/hub_testkit/traceability/model.py",
    "testframework/src/hub_testkit/traceability/loader.py",
    "testframework/src/hub_testkit/traceability/index.py",
    "testframework/src/hub_testkit/domains/health/__init__.py",
    "testframework/src/hub_testkit/domains/health/api.py",
    "testframework/src/hub_testkit/domains/health/models.py",
    "testframework/src/hub_testkit/domains/health/oracles.py",
    "testframework/selftests/runner/test_launcher_contract.py",
    "testframework/selftests/runner/test_launcher_cleanup.py",
    "testframework/selftests/runtime/test_ids.py",
    "testframework/selftests/runtime/test_ownership.py",
    "testframework/selftests/runtime/test_redaction.py",
    "testframework/selftests/traceability/test_loader.py",
    "testframework/selftests/traceability/test_index.py",
    "testframework/tests/conftest.py",
    "testframework/tests/api/health/test_health.py",
    "testframework/cases/health/TC-HEALTH-0001.md",
    "testframework/schemas/test-case.schema.json",
    "testframework/schemas/run.schema.json",
    "testframework/policies/markers.toml",
    "testframework/policies/environments.toml",
    "testframework/policies/redaction.toml",
    "testframework/docker/Dockerfile.core",
    "testframework/docker/run-task.ps1",
    "testframework/docker/images.lock.json",
    "testframework/runner/entrypoint.py",
    "testframework/reports/.gitignore",
]
PYTHON_SCAFFOLD_FORBIDDEN_PATHS = [
    "hub_testkit", "tests", "cases", "selftests", "docker", "load",
    "testframework/python", "testframework/core", "testframework/api",
    "testframework/integration", "testframework/contract", "testframework/e2e",
    "testframework/conftest.py", "testframework/requirements.lock",
    "testframework/pytest.ini", "testframework/docker/Dockerfile",
    "testframework/docker/compose.yaml", "testframework/tests/e2e/support",
    "testframework/tests/e2e/pages", "testframework/tests/e2e/steps",
]


class ManifestError(ValueError):
    pass


class MissingPathError(ManifestError):
    """A proven ENOENT; other filesystem failures must never trigger fallback."""


def canonical_json_bytes(value: Any, *, pretty: bool = False) -> bytes:
    if pretty:
        return (json.dumps(value, ensure_ascii=False, indent=2, sort_keys=True) + "\n").encode("utf-8")
    return json.dumps(value, ensure_ascii=False, separators=(",", ":"), sort_keys=True).encode("utf-8")


def object_digest(value: Any) -> str:
    return hashlib.sha256(canonical_json_bytes(value)).hexdigest()


def path_list_digest(paths: list[str]) -> str:
    return hashlib.sha256(("\n".join(sorted(paths)) + "\n").encode("utf-8")).hexdigest()


def is_symlink_or_reparse(metadata: os.stat_result) -> bool:
    return stat.S_ISLNK(metadata.st_mode) or bool(
        int(getattr(metadata, "st_file_attributes", 0)) & FILE_ATTRIBUTE_REPARSE_POINT
    )


def metadata_signature(metadata: os.stat_result, *, content: bool) -> tuple[int, ...]:
    result = (
        int(metadata.st_mode), int(metadata.st_dev), int(metadata.st_ino),
        int(metadata.st_nlink), int(getattr(metadata, "st_file_attributes", 0)),
    )
    if content:
        result += (
            int(metadata.st_size),
            int(getattr(metadata, "st_mtime_ns", int(metadata.st_mtime * 1_000_000_000))),
            int(getattr(metadata, "st_ctime_ns", int(metadata.st_ctime * 1_000_000_000))),
        )
    return result


def open_physical_root(path: Path, flags: int) -> tuple[int, tuple[int, ...]]:
    """Independently open every repo-root component no-follow."""

    absolute = Path(os.path.abspath(path))
    if (
        len(absolute.parts) - 1 > MAX_PATH_COMPONENTS
        or len(str(absolute).encode("utf-8")) > MAX_PATH_UTF8_BYTES
        or any(len(part.encode("utf-8")) > MAX_COMPONENT_UTF8_BYTES for part in absolute.parts[1:])
    ):
        raise ManifestError("validator repository root exceeds component/path budget")
    protected = os.path.normcase(os.path.abspath(REPO_ROOT / PROTECTED_PREFIX.rstrip("/")))
    candidate = os.path.normcase(str(absolute))
    try:
        inside_protected = os.path.commonpath([candidate, protected]) == protected
    except ValueError:
        inside_protected = False
    if inside_protected:
        raise ManifestError("validator repository root enters protected pro-qa-ai")
    anchor = Path(absolute.anchor)
    descriptor = os.open(anchor, flags)
    try:
        for part in absolute.parts[1:]:
            before = os.stat(part, dir_fd=descriptor, follow_symlinks=False)
            if is_symlink_or_reparse(before) or not stat.S_ISDIR(before.st_mode):
                raise ManifestError("validator repository ancestor is not physical")
            next_fd = os.open(part, flags, dir_fd=descriptor)
            opened = os.fstat(next_fd)
            if metadata_signature(before, content=False) != metadata_signature(opened, content=False):
                os.close(next_fd)
                raise ManifestError("validator repository ancestor changed while opening")
            os.close(descriptor)
            descriptor = next_fd
        return descriptor, metadata_signature(os.fstat(descriptor), content=False)
    except Exception:
        os.close(descriptor)
        raise


def canonical_path_errors(value: Any, label: str) -> list[str]:
    if not isinstance(value, str) or not value:
        return [f"{label} must be a non-empty string"]
    if "\\" in value or value.startswith(("/", "//")) or re.match(r"^[A-Za-z]:", value):
        return [f"{label} is not repository-relative POSIX"]
    if unicodedata.normalize("NFC", value) != value or any(ord(ch) < 32 or ord(ch) == 127 for ch in value):
        return [f"{label} is not canonical Unicode"]
    parts = value.split("/")
    if any(part in {"", ".", ".."} or part.rstrip(" .") != part or ":" in part for part in parts):
        return [f"{label} contains traversal/alias segment"]
    if len(parts) > MAX_PATH_COMPONENTS:
        return [f"{label} exceeds {MAX_PATH_COMPONENTS} component budget"]
    if len(value.encode("utf-8")) > MAX_PATH_UTF8_BYTES:
        return [f"{label} exceeds {MAX_PATH_UTF8_BYTES} UTF-8 byte budget"]
    if any(len(part.encode("utf-8")) > MAX_COMPONENT_UTF8_BYTES for part in parts):
        return [f"{label} component exceeds {MAX_COMPONENT_UTF8_BYTES} UTF-8 byte budget"]
    if PurePosixPath(value).as_posix() != value:
        return [f"{label} is not canonical POSIX-relative"]
    if value == "pro-qa-ai" or value.startswith(PROTECTED_PREFIX):
        return [f"{label} enters protected pro-qa-ai"]
    return []


def require_canonical_path(value: Any, label: str) -> str:
    errors = canonical_path_errors(value, label)
    if errors:
        raise ManifestError(errors[0])
    return str(value)


def reject_portable_aliases(paths: list[str], label: str) -> None:
    seen: dict[str, str] = {}
    for path in paths:
        require_canonical_path(path, label)
        key = "/".join(unicodedata.normalize("NFC", part).casefold().rstrip(" .") for part in path.split("/"))
        previous = seen.setdefault(key, path)
        if previous != path:
            raise ManifestError(f"{label} portable alias/case collision")


class IndependentRepo:
    """Validator-owned descriptor-relative reader; no builder code is imported."""

    def __init__(self, root: Path):
        self.root = Path(os.path.abspath(root))
        self._identities: dict[tuple[int, int], str] = {}
        self._observations: dict[str, tuple[str, tuple[tuple[int, ...], ...], int]] = {}
        self._absent_paths: set[str] = set()
        directory_flags, _ = self.flags()
        root_fd, self._root_signature = open_physical_root(self.root, directory_flags)
        os.close(root_fd)

    @staticmethod
    def flags() -> tuple[int, int]:
        required = ("O_CLOEXEC", "O_DIRECTORY", "O_NOFOLLOW", "O_PATH")
        missing = [name for name in required if not hasattr(os, name)]
        if os.name != "posix" or missing or os.open not in getattr(os, "supports_dir_fd", set()):
            raise ManifestError(f"descriptor-relative no-follow unavailable: os={os.name} missing={missing}")
        return (
            os.O_PATH | os.O_CLOEXEC | os.O_DIRECTORY | os.O_NOFOLLOW,
            os.O_RDONLY | os.O_CLOEXEC | os.O_NOFOLLOW | getattr(os, "O_NONBLOCK", 0),
        )

    def _read_once(self, relative: str, max_bytes: int) -> tuple[bytes, tuple[tuple[int, ...], ...]]:
        require_canonical_path(relative, "validator repository path")
        directory_flags, file_flags = self.flags()
        descriptors: list[int] = []
        signatures: list[tuple[int, ...]] = []
        try:
            root_fd, root_signature = open_physical_root(self.root, directory_flags)
            descriptors.append(root_fd)
            root_opened = os.fstat(root_fd)
            if (
                root_signature != self._root_signature
                or root_signature != metadata_signature(root_opened, content=False)
            ):
                raise ManifestError("validator repository root link/identity rejected")
            signatures.append(metadata_signature(root_opened, content=False))
            parts = relative.split("/")
            for index, part in enumerate(parts):
                leaf = index == len(parts) - 1
                try:
                    before = os.stat(part, dir_fd=descriptors[-1], follow_symlinks=False)
                except FileNotFoundError:
                    raise MissingPathError(f"validator required path is missing: {relative}") from None
                if is_symlink_or_reparse(before):
                    raise ManifestError(f"validator symlink/reparse component rejected: {relative}")
                descriptor = os.open(part, file_flags if leaf else directory_flags, dir_fd=descriptors[-1])
                descriptors.append(descriptor)
                opened = os.fstat(descriptor)
                if metadata_signature(before, content=leaf) != metadata_signature(opened, content=leaf):
                    raise ManifestError(f"validator identity changed while opening: {relative}")
                if leaf:
                    if not stat.S_ISREG(opened.st_mode) or int(opened.st_nlink) != 1:
                        raise ManifestError(f"validator single-link regular file required: {relative}")
                    if int(opened.st_size) > max_bytes:
                        raise ManifestError(f"validator bounded read exceeded: {relative}")
                elif not stat.S_ISDIR(opened.st_mode):
                    raise ManifestError(f"validator non-directory ancestor rejected: {relative}")
                signatures.append(metadata_signature(opened, content=leaf))
            chunks: list[bytes] = []
            size = 0
            while True:
                chunk = os.read(descriptors[-1], min(1024 * 1024, max_bytes + 1 - size))
                if not chunk:
                    break
                chunks.append(chunk)
                size += len(chunk)
                if size > max_bytes:
                    raise ManifestError(f"validator bounded read exceeded: {relative}")
            if signatures[-1] != metadata_signature(os.fstat(descriptors[-1]), content=True):
                raise ManifestError(f"validator source changed during read: {relative}")
            return b"".join(chunks), tuple(signatures)
        finally:
            for descriptor in reversed(descriptors):
                try:
                    os.close(descriptor)
                except OSError:
                    pass

    def read(self, relative: str, max_bytes: int = MAX_FILE_BYTES) -> bytes:
        payload, before = self._read_once(relative, max_bytes)
        confirmation, after = self._read_once(relative, max_bytes)
        if payload != confirmation or before != after:
            raise ManifestError(f"validator file changed during independent recheck: {relative}")
        identity = (after[-1][1], after[-1][2])
        previous = self._identities.setdefault(identity, relative)
        if previous != relative:
            raise ManifestError(f"validator hardlink/alias collision: {previous} and {relative}")
        observation = (hashlib.sha256(payload).hexdigest(), after, max_bytes)
        prior_observation = self._observations.setdefault(relative, observation)
        if prior_observation[:2] != observation[:2]:
            raise ManifestError(f"validator file identity changed across reads: {relative}")
        return payload

    def sha256(self, relative: str) -> str:
        return hashlib.sha256(self.read(relative)).hexdigest()

    def assert_absent(self, relative: str, *, record: bool = True) -> None:
        require_canonical_path(relative, "validator destination")
        directory_flags, _ = self.flags()
        for _ in range(2):
            descriptors: list[int] = []
            try:
                root_fd, root_signature = open_physical_root(self.root, directory_flags)
                descriptors.append(root_fd)
                if root_signature != self._root_signature or metadata_signature(os.fstat(root_fd), content=False) != root_signature:
                    raise ManifestError("validator repository root identity changed")
                missing = False
                for part in relative.split("/"):
                    try:
                        meta = os.stat(part, dir_fd=descriptors[-1], follow_symlinks=False)
                    except FileNotFoundError:
                        missing = True
                        break
                    if is_symlink_or_reparse(meta):
                        raise ManifestError(f"validator destination link rejected: {relative}")
                    if not stat.S_ISDIR(meta.st_mode):
                        break
                    descriptor = os.open(part, directory_flags, dir_fd=descriptors[-1])
                    opened = os.fstat(descriptor)
                    if metadata_signature(meta, content=False) != metadata_signature(opened, content=False):
                        os.close(descriptor)
                        raise ManifestError(f"validator destination ancestor changed: {relative}")
                    descriptors.append(descriptor)
                if not missing:
                    raise ManifestError(f"existing move destination rejected: {relative}")
            finally:
                for descriptor in reversed(descriptors):
                    os.close(descriptor)
        if record:
            self._absent_paths.add(relative)

    def finalize(self) -> None:
        directory_flags, _ = self.flags()
        root_fd, current_root = open_physical_root(self.root, directory_flags)
        os.close(root_fd)
        if current_root != self._root_signature:
            raise ManifestError("validator repository root changed before finalization")
        for relative, (digest, signature, max_bytes) in self._observations.items():
            payload, after = self._read_once(relative, max_bytes)
            if after != signature or hashlib.sha256(payload).hexdigest() != digest:
                raise ManifestError(f"validator file changed before finalization: {relative}")
        for relative in tuple(self._absent_paths):
            self.assert_absent(relative, record=False)


def json_load(payload: bytes, label: str) -> dict[str, Any]:
    value = json.loads(payload.decode("utf-8"))
    if not isinstance(value, dict):
        raise ManifestError(f"{label} must be a JSON object")
    return value


def read_git_head(safe: IndependentRepo) -> str:
    head = safe.read(".git/HEAD", 4096).decode("ascii").strip()
    if head.startswith("ref: "):
        reference = require_canonical_path(f".git/{head[5:]}", "git reference")
        try:
            value = safe.read(reference, 4096).decode("ascii").strip()
        except MissingPathError:
            packed = safe.read(".git/packed-refs", 8 * 1024 * 1024).decode("ascii")
            matches = [line.split(" ", 1)[0] for line in packed.splitlines() if line.endswith(f" {head[5:]}")]
            if len(matches) != 1:
                raise
            value = matches[0]
    else:
        value = head
    if HEX_SHA_RE.fullmatch(value) is None:
        raise ManifestError("validator repository HEAD is not canonical")
    return value


def read_loose_git_object(safe: IndependentRepo, object_id: str) -> tuple[str, bytes]:
    """Read and authenticate one bounded loose Git object."""

    if not is_git_sha(object_id):
        raise ManifestError("Git commit object ID is invalid")
    try:
        compressed = safe.read(f".git/objects/{object_id[:2]}/{object_id[2:]}", 16 * 1024 * 1024)
    except MissingPathError as problem:
        raise ManifestError(
            f"descriptor-safe loose Git commit unavailable (packed fallback is forbidden): {object_id}"
        ) from problem
    decompressor = zlib.decompressobj()
    raw = decompressor.decompress(compressed, 16 * 1024 * 1024 + 1)
    if len(raw) > 16 * 1024 * 1024 or decompressor.unused_data or decompressor.unconsumed_tail:
        raise ManifestError(f"Git commit object exceeds bounded decode: {object_id}")
    raw += decompressor.flush(16 * 1024 * 1024 + 1 - len(raw))
    if len(raw) > 16 * 1024 * 1024:
        raise ManifestError(f"Git commit object exceeds bounded decode: {object_id}")
    algorithm = hashlib.sha1 if len(object_id) == 40 else hashlib.sha256
    if algorithm(raw).hexdigest() != object_id:
        raise ManifestError(f"Git commit object digest mismatch: {object_id}")
    header, separator, body = raw.partition(b"\x00")
    if not separator or b" " not in header:
        raise ManifestError(f"Git object header invalid: {object_id}")
    kind_raw, declared_raw = header.split(b" ", 1)
    try:
        kind = kind_raw.decode("ascii")
        declared = int(declared_raw.decode("ascii"))
    except (UnicodeError, ValueError) as problem:
        raise ManifestError(f"Git object header invalid: {object_id}") from problem
    if declared != len(body):
        raise ManifestError(f"Git object length mismatch: {object_id}")
    if kind not in {"commit", "tree", "blob"}:
        raise ManifestError(f"Git object type rejected: {kind}")
    return kind, body


def read_loose_commit_details(safe: IndependentRepo, object_id: str) -> tuple[str, list[str]]:
    kind, body = read_loose_git_object(safe, object_id)
    if kind != "commit":
        raise ManifestError(f"Git object is not a commit: {object_id}")
    tree: str | None = None
    parents: list[str] = []
    for line in body.split(b"\n"):
        if not line:
            break
        if line.startswith(b"tree "):
            tree = line[5:].decode("ascii", "strict")
        if line.startswith(b"parent "):
            parent = line[7:].decode("ascii", "strict")
            if len(parent) != len(object_id) or not is_git_sha(parent):
                raise ManifestError(f"Git parent ID invalid: {object_id}")
            parents.append(parent)
    if tree is None or len(tree) != len(object_id) or not is_git_sha(tree):
        raise ManifestError(f"Git commit tree ID invalid: {object_id}")
    return tree, parents


def read_loose_commit(safe: IndependentRepo, object_id: str) -> list[str]:
    return read_loose_commit_details(safe, object_id)[1]


def read_git_tree_artifact(safe: IndependentRepo, commit_id: str, relative: str) -> bytes:
    """Resolve a regular blob through authenticated loose commit/tree objects."""

    errors = canonical_path_errors(relative, "historical manifest path")
    if errors:
        raise ManifestError(errors[0])
    tree_id, _ = read_loose_commit_details(safe, commit_id)
    oid_bytes = len(commit_id) // 2
    parts = relative.split("/")
    for depth, part in enumerate(parts):
        kind, body = read_loose_git_object(safe, tree_id)
        if kind != "tree":
            raise ManifestError(f"Git path ancestor is not a tree: {relative}")
        cursor = 0
        matched: tuple[str, str] | None = None
        while cursor < len(body):
            space = body.find(b" ", cursor)
            nul = body.find(b"\x00", space + 1)
            if space <= cursor or nul <= space or nul + 1 + oid_bytes > len(body):
                raise ManifestError(f"Git tree encoding invalid: {tree_id}")
            mode = body[cursor:space].decode("ascii", "strict")
            name = body[space + 1:nul].decode("utf-8", "strict")
            object_id = body[nul + 1:nul + 1 + oid_bytes].hex()
            cursor = nul + 1 + oid_bytes
            if name == part:
                if matched is not None:
                    raise ManifestError(f"duplicate Git tree entry: {relative}")
                matched = (mode, object_id)
        if matched is None:
            raise ManifestError(f"historical manifest absent from commit tree: {relative}")
        mode, tree_id = matched
        if depth < len(parts) - 1 and mode not in {"40000", "040000"}:
            raise ManifestError(f"historical manifest ancestor mode rejected: {relative}")
    if mode not in {"100644", "100755"}:
        raise ManifestError(f"historical manifest blob mode rejected: {relative}")
    kind, payload = read_loose_git_object(safe, tree_id)
    if kind != "blob":
        raise ManifestError(f"historical manifest is not a blob: {relative}")
    return payload


def require_git_ancestor(safe: IndependentRepo, ancestor: str, descendant: str) -> None:
    """Bounded descriptor-safe commit walk; fail closed on packed/missing objects."""

    pending = [descendant]
    seen: set[str] = set()
    while pending and len(seen) <= 4096:
        current = pending.pop()
        if current == ancestor:
            return
        if current in seen:
            continue
        seen.add(current)
        pending.extend(read_loose_commit(safe, current))
    if len(seen) > 4096:
        raise ManifestError("Git ancestry walk exceeds 4096-commit budget")
    raise ManifestError(f"Git commit is not an ancestor: {ancestor} -> {descendant}")


def error(message: str) -> list[str]:
    return [message]


def is_sha256(value: Any) -> bool:
    return isinstance(value, str) and SHA256_RE.fullmatch(value) is not None


def is_git_sha(value: Any) -> bool:
    return isinstance(value, str) and HEX_SHA_RE.fullmatch(value) is not None


def batch_projection_digest(manifest: dict[str, Any], batch_id: str) -> str:
    """Hash one batch plus immutable shared authority, never sibling state."""

    matches = [
        copy.deepcopy(batch)
        for batch in manifest.get("batches", [])
        if isinstance(batch, dict) and batch.get("id") == batch_id
    ]
    batch = matches[0] if len(matches) == 1 else None
    if isinstance(batch, dict):
        transition = batch.get("state_transition")
        if isinstance(transition, dict):
            transition["manifest_after_projection_sha256"] = None
            history = transition.get("history")
            if isinstance(history, list) and history and isinstance(history[-1], dict):
                history[-1]["after_projection_sha256"] = None
        verifier = batch.get("verifier_record")
        if isinstance(verifier, dict):
            if "receipt_path" in verifier:
                verifier["receipt_path"] = None
            if "receipt_sha256" in verifier:
                verifier["receipt_sha256"] = None
        consumers = batch.get("consumer_transitions")
        if isinstance(consumers, list):
            for item in consumers:
                if isinstance(item, dict) and item.get("consumer_id") == "consumer-manifest-self":
                    item["after_sha256"] = None
    projection = {
        "schema_version": manifest.get("schema_version"),
        "manifest_id": manifest.get("manifest_id"),
        "source": copy.deepcopy(manifest.get("source")),
        "invariants": copy.deepcopy(manifest.get("invariants")),
        "gate_registry": copy.deepcopy(manifest.get("gate_registry")),
        "consumer_registry": copy.deepcopy(manifest.get("consumer_registry")),
        "totals": copy.deepcopy(manifest.get("totals")),
        "batch_id": batch_id,
        "batch": batch,
    }
    return object_digest(projection)


def validate_top_level(actual: dict[str, Any]) -> list[str]:
    errors: list[str] = []
    required = {
        "schema_version",
        "manifest_id",
        "status",
        "source",
        "invariants",
        "gate_registry",
        "consumer_registry",
        "batches",
        "totals",
    }
    if set(actual) != required:
        errors.append("top-level schema keys changed")
    source = actual.get("source")
    if not isinstance(source, dict) or not is_git_sha(source.get("repository_head")):
        errors.append("wrong repository head: canonical Git SHA required")
    batches = actual.get("batches")
    if not isinstance(batches, list) or not batches:
        errors.append("batches must be a non-empty ordered array")
    registry = actual.get("consumer_registry")
    if not isinstance(registry, list) or not registry:
        errors.append("consumer registry must be a non-empty ordered array")
    gates = actual.get("gate_registry")
    if not isinstance(gates, list) or not gates:
        errors.append("gate registry must be a non-empty ordered array")
    return errors


def prepare_expected_for_snapshot(
    expected: dict[str, Any], actual: dict[str, Any], errors: list[str]
) -> dict[str, Any]:
    """Inject only non-authority before-image hashes into a derived skeleton.

    Ledger hashes and every path/action/order remain derived.  The founder
    authority seals the 28 non-ledger paths but not their bytes, and an applied
    tree no longer contains those before images; their builder-captured hashes
    are therefore syntax/digest-checked rather than reinvented from JSON.
    """

    adjusted = copy.deepcopy(expected)
    actual_registry = actual.get("consumer_registry")
    expected_registry = adjusted.get("consumer_registry")
    if not isinstance(actual_registry, list) or not isinstance(expected_registry, list):
        return adjusted
    if len(actual_registry) != len(expected_registry):
        errors.append("missing/extra consumer registry entry")
        return adjusted
    for expected_item, actual_item in zip(expected_registry, actual_registry):
        if not isinstance(actual_item, dict):
            errors.append("consumer registry entry is not an object")
            continue
        before = actual_item.get("before_sha256")
        if actual_item.get("mode") == "self":
            if before is not None:
                errors.append("manifest-self before hash must be null in planned authority")
        elif not is_sha256(before):
            errors.append(f"consumer before hash invalid: {actual_item.get('id')}")
        expected_item["before_sha256"] = before

    expected_batches = adjusted.get("batches", [])
    actual_batches = actual.get("batches", [])
    if not isinstance(actual_batches, list) or len(actual_batches) != len(expected_batches):
        return adjusted
    adjusted_registry_by_id = {item["id"]: item for item in expected_registry}
    for expected_batch, actual_batch in zip(expected_batches, actual_batches):
        if not isinstance(actual_batch, dict):
            errors.append("batch entry is not an object")
            continue
        actual_operations = actual_batch.get("operations")
        expected_operations = expected_batch.get("operations")
        if not isinstance(actual_operations, list) or not isinstance(expected_operations, list):
            continue
        if len(actual_operations) != len(expected_operations):
            continue
        for expected_op, actual_op in zip(expected_operations, actual_operations):
            if expected_op.get("authority_group") == "DELETE-NONLEDGER-CLOSURE":
                captured = actual_op.get("source_sha256") if isinstance(actual_op, dict) else None
                if not is_sha256(captured):
                    errors.append(f"non-ledger current file SHA invalid: {expected_op.get('source')}")
                expected_op["source_sha256"] = captured
        expected_batch["preconditions"]["operations_sha256"] = object_digest(expected_operations)
        consumer_ids = expected_batch["affected_root_consumers"]
        expected_batch["preconditions"]["consumer_before_snapshot_sha256"] = object_digest(
            [adjusted_registry_by_id[item] for item in consumer_ids]
        )
    return adjusted


def static_projection(manifest: dict[str, Any]) -> dict[str, Any]:
    projected = copy.deepcopy(manifest)
    projected["status"] = "<dynamic-status>"
    for batch in projected.get("batches", []):
        if not isinstance(batch, dict):
            continue
        for key in DYNAMIC_BATCH_KEYS:
            batch.pop(key, None)
    return projected


def global_state_order_errors(manifest: dict[str, Any]) -> list[str]:
    """Require verified-prefix, optional applied batch, then planned suffix."""

    errors: list[str] = []
    phase = "verified"
    for index, batch in enumerate(manifest.get("batches", [])):
        if not isinstance(batch, dict):
            continue
        state = batch.get("state")
        if state not in {"planned", "applied-unverified", "verified"}:
            continue
        if phase == "verified":
            if state == "applied-unverified":
                phase = "applied-unverified"
            elif state == "planned":
                phase = "planned"
        elif phase == "applied-unverified":
            if state == "planned":
                phase = "planned"
            else:
                errors.append(
                    f"global batch transition order is not monotone at index {index}: {state}"
                )
        elif state != "planned":
            errors.append(
                f"global batch transition order is not monotone at index {index}: {state}"
            )
    return errors


def planned_batch_state(
    batch: dict[str, Any], registry_by_id: dict[str, dict[str, Any]]
) -> dict[str, Any]:
    transitions = []
    for consumer_id in batch.get("affected_root_consumers", []):
        consumer = registry_by_id.get(str(consumer_id))
        if consumer is None or consumer.get("mode") == "delete":
            continue
        transitions.append(
            {
                "consumer_id": consumer_id,
                "before_sha256": consumer.get("before_sha256"),
                "after_sha256": None,
            }
        )
    return {
        "state": "planned",
        "state_transition": {
            "history": [],
            "manifest_before_sha256": None,
            "manifest_after_projection_sha256": None,
        },
        "consumer_transitions": transitions,
        "commit": {"applied_commit": None, "verifier_commit": None},
        "verifier_record": {"identity": None, "verdict": None, "verified_at": None},
        "evidence": {"docker": None, "evidence_uri": None, "cleanup": None},
        "rollback": {
            "method": "git-revert-exact-commit",
            "commit_to_revert": None,
            "revert_commit": None,
        },
    }


def expected_global_predecessor_manifest(
    manifest: dict[str, Any], batch_index: int
) -> dict[str, Any]:
    predecessor = copy.deepcopy(manifest)
    batches = predecessor.get("batches")
    registry = predecessor.get("consumer_registry")
    if (
        not isinstance(batches, list)
        or not 0 <= batch_index < len(batches)
        or not isinstance(batches[batch_index], dict)
        or not isinstance(registry, list)
    ):
        raise ManifestError("cannot derive immediate global predecessor from malformed manifest")
    registry_by_id = {
        str(item.get("id")): item
        for item in registry
        if isinstance(item, dict) and isinstance(item.get("id"), str)
    }
    batch = batches[batch_index]
    for key, value in planned_batch_state(batch, registry_by_id).items():
        batch[key] = value
    states = [item.get("state") for item in batches if isinstance(item, dict)]
    predecessor["status"] = (
        "planned"
        if states and all(state == "planned" for state in states)
        else "verified"
        if states and all(state == "verified" for state in states)
        else "in-progress"
    )
    return predecessor


def predecessor_snapshot_errors(
    manifest: dict[str, Any],
    historical: dict[str, Any],
    batch_index: int,
    historical_commit: str,
) -> list[str]:
    """Compare authenticated bytes with the exact preceding global state."""

    expected = expected_global_predecessor_manifest(manifest, batch_index)
    normalized = copy.deepcopy(historical)
    errors: list[str] = []
    if batch_index == 0:
        source = normalized.get("source")
        historical_source = source.get("repository_head") if isinstance(source, dict) else None
        if isinstance(source, dict):
            source["repository_head"] = historical_commit
        for batch in normalized.get("batches", []):
            preconditions = batch.get("preconditions") if isinstance(batch, dict) else None
            if isinstance(preconditions, dict):
                if preconditions.get("repository_head") != historical_source:
                    errors.append(
                        "bootstrap historical batch repository head is inconsistent"
                    )
                preconditions["repository_head"] = historical_commit
    if normalized != expected:
        errors.append(
            "authenticated historical manifest is not the exact immediate global predecessor"
        )
    return errors


def validate_static_structure(
    actual: dict[str, Any], expected: dict[str, Any] | None = None
) -> list[str]:
    errors: list[str] = []
    if expected is not None:
        adjusted = prepare_expected_for_snapshot(expected, actual, errors)
        if static_projection(actual) != static_projection(adjusted):
            errors.append("independent planned manifest mismatch: missing/extra/reordered operation, batch, consumer or gate")

    registry = actual.get("consumer_registry", [])
    registry_ids: list[str] = []
    registry_paths: list[str] = []
    for index, item in enumerate(registry if isinstance(registry, list) else []):
        if not isinstance(item, dict):
            errors.append(f"consumer registry item {index} is not an object")
            continue
        consumer_id = item.get("id")
        path = item.get("path")
        if not isinstance(consumer_id, str):
            errors.append(f"consumer registry item {index} has no stable ID")
        else:
            registry_ids.append(consumer_id)
        errors.extend(canonical_path_errors(path, f"consumer registry path {index}"))
        if isinstance(path, str):
            registry_paths.append(path)
    if len(registry_ids) != len(set(registry_ids)) or len(registry_paths) != len(set(registry_paths)):
        errors.append("consumer registry contains duplicate IDs/paths")
    try:
        reject_portable_aliases(registry_paths, "consumer registry")
    except ManifestError as problem:
        errors.append(str(problem))

    operation_sources: list[str] = []
    move_targets: list[str] = []
    batch_ids: list[str] = []
    batches = actual.get("batches", [])
    if not isinstance(batches, list):
        return errors
    for index, batch in enumerate(batches):
        if not isinstance(batch, dict):
            errors.append(f"batch {index} is not an object")
            continue
        batch_id = batch.get("id")
        if not isinstance(batch_id, str) or ASCII_SLUG_RE.fullmatch(batch_id) is None:
            errors.append(f"batch {index} ID is not lowercase ASCII slug")
        else:
            batch_ids.append(batch_id)
        if batch.get("order") != index + 1:
            errors.append(f"batch order mismatch at index {index}")
        operations = batch.get("operations")
        if not isinstance(operations, list) or not operations:
            errors.append(f"batch operations missing: {batch_id}")
            continue
        if len(operations) > MAX_BATCH_OPERATIONS:
            errors.append(f"batch exceeds 64 operations: {batch_id}")
        sources_here: list[str] = []
        for op_index, operation in enumerate(operations):
            if not isinstance(operation, dict):
                errors.append(f"operation is not an object: {batch_id}:{op_index}")
                continue
            source = operation.get("source")
            errors.extend(canonical_path_errors(source, f"operation source {batch_id}:{op_index}"))
            if isinstance(source, str):
                sources_here.append(source)
                operation_sources.append(source)
            if operation.get("action") == "move":
                target = operation.get("target")
                errors.extend(canonical_path_errors(target, f"move target {batch_id}:{op_index}"))
                if isinstance(target, str):
                    move_targets.append(target)
            if not is_sha256(operation.get("source_sha256")):
                errors.append(f"operation source digest invalid: {batch_id}:{op_index}")
        preconditions = batch.get("preconditions")
        if not isinstance(preconditions, dict):
            errors.append(f"preconditions missing: {batch_id}")
        else:
            if preconditions.get("operations_sha256") != object_digest(operations):
                errors.append(f"digest substitution detected: {batch_id}")
            if preconditions.get("source_paths_sha256") != path_list_digest(sources_here):
                errors.append(f"source path digest substitution detected: {batch_id}")
            if preconditions.get("max_operations") != MAX_BATCH_OPERATIONS:
                errors.append(f"unbounded batch declaration: {batch_id}")
        stale = batch.get("stale_reference_scan")
        if not isinstance(stale, dict):
            errors.append(f"stale-reference declaration missing: {batch_id}")
        else:
            patterns = stale.get("exact_patterns")
            roots = stale.get("roots")
            consumers = batch.get("affected_root_consumers")
            if patterns != sorted(set(sources_here)):
                errors.append(f"stale-reference declaration failure: {batch_id}")
            if not isinstance(consumers, list) or not consumers:
                errors.append(f"missing affected consumer declaration: {batch_id}")
            elif roots != [registry[registry_ids.index(item)]["path"] for item in consumers if item in registry_ids]:
                errors.append(f"stale-reference roots mismatch: {batch_id}")
            if stale.get("required_remaining_occurrences") != 0:
                errors.append(f"stale-reference zero policy changed: {batch_id}")
    if len(batch_ids) != len(set(batch_ids)):
        errors.append("duplicate batch IDs")
    if len(operation_sources) != len(set(operation_sources)):
        errors.append("duplicate/overlapping operation source")
    if len(move_targets) != len(set(move_targets)):
        errors.append("duplicate move target")
    try:
        reject_portable_aliases(operation_sources, "operation source")
        reject_portable_aliases(move_targets, "move target")
    except ManifestError as problem:
        errors.append(str(problem))
    return errors


def validate_transition_history(
    batch: dict[str, Any],
    state: str,
    manifest: dict[str, Any],
    safe: IndependentRepo,
    batch_index: int,
) -> list[str]:
    errors: list[str] = []
    batch_id = str(batch.get("id"))
    transition = batch.get("state_transition")
    if not isinstance(transition, dict):
        return [f"state transition missing: {batch_id}"]
    history = transition.get("history")
    if not isinstance(history, list):
        return [f"state transition history missing: {batch_id}"]
    if state == "planned":
        expected = {
            "history": [],
            "manifest_before_sha256": None,
            "manifest_after_projection_sha256": None,
        }
        if transition != expected:
            errors.append(f"planned state has transition evidence: {batch_id}")
        return errors
    required_pairs = [("planned", "applied-unverified")]
    if state == "verified":
        required_pairs.append(("applied-unverified", "verified"))
    if len(history) != len(required_pairs):
        return [f"invalid transition count/skip: {batch_id}"]
    for index, (entry, pair) in enumerate(zip(history, required_pairs)):
        if not isinstance(entry, dict) or entry.get("from") != pair[0] or entry.get("to") != pair[1]:
            errors.append(f"invalid transition order/skip: {batch_id}:{index}")
            continue
        if set(entry) != {
            "after_projection_sha256", "before_manifest_commit",
            "before_manifest_path", "before_manifest_sha256", "from",
            "recorded_at", "to",
        }:
            errors.append(f"transition artifact schema mismatch: {batch_id}:{index}")
        if not is_sha256(entry.get("before_manifest_sha256")):
            errors.append(f"transition before manifest digest invalid: {batch_id}:{index}")
        if not is_sha256(entry.get("after_projection_sha256")):
            errors.append(f"transition after projection digest invalid: {batch_id}:{index}")
        if not isinstance(entry.get("recorded_at"), str) or TIMESTAMP_RE.fullmatch(entry["recorded_at"]) is None:
            errors.append(f"transition timestamp invalid: {batch_id}:{index}")
        if index > 0 and isinstance(history[index - 1], dict):
            if history[index - 1].get("after_projection_sha256") != entry.get("before_manifest_sha256"):
                errors.append(f"transition digest chain is discontinuous: {batch_id}:{index}")
        if index == 0:
            historical_path = entry.get("before_manifest_path")
            historical_commit = entry.get("before_manifest_commit")
            expected_prefix = f".planning/audits/test-rebase/transition-artifacts/{batch_id}/"
            if not isinstance(historical_path, str) or not historical_path.startswith(expected_prefix) or not historical_path.endswith(".json"):
                errors.append(f"first transition historical artifact path invalid: {batch_id}")
            elif not is_git_sha(historical_commit):
                errors.append(f"first transition historical commit invalid: {batch_id}")
            else:
                payload, artifact_errors = read_bound_artifact(
                    safe,
                    historical_path,
                    entry.get("before_manifest_sha256"),
                    f"historical manifest {batch_id}",
                    max_bytes=32 * 1024 * 1024,
                )
                errors.extend(artifact_errors)
                if payload is not None:
                    try:
                        committed_payload = read_git_tree_artifact(safe, historical_commit, historical_path)
                        historical = json_load(payload, "historical manifest")
                        if committed_payload != payload:
                            errors.append(f"historical manifest differs from authenticated commit tree: {batch_id}")
                        if canonical_json_bytes(historical, pretty=True) != payload:
                            errors.append(f"historical manifest is not canonical JSON: {batch_id}")

                        applied_commit = batch.get("commit", {}).get("applied_commit")
                        if not is_git_sha(applied_commit):
                            errors.append(f"applied commit missing for predecessor authentication: {batch_id}")
                        else:
                            applied_parents = read_loose_commit(safe, str(applied_commit))
                            if applied_parents != [str(historical_commit)]:
                                errors.append(
                                    f"historical commit is not the sole immediate parent of applied commit: {batch_id}"
                                )
                            require_git_ancestor(
                                safe, str(historical_commit), str(applied_commit)
                            )

                        if batch_index == 0:
                            source_commit = manifest.get("source", {}).get("repository_head")
                            if historical_commit != source_commit:
                                errors.append(f"first batch predecessor commit mismatch: {batch_id}")
                            historical_source = historical.get("source", {}).get("repository_head")
                            if (
                                not is_git_sha(historical_source)
                                or historical_source == historical_commit
                            ):
                                errors.append(
                                    f"historical manifest predecessor commit invalid: {batch_id}"
                                )
                            else:
                                read_loose_commit(safe, str(historical_source))
                                require_git_ancestor(
                                    safe, str(historical_source), str(historical_commit)
                                )
                        else:
                            batches = manifest.get("batches", [])
                            previous = batches[batch_index - 1]
                            previous_commit = (
                                previous.get("commit", {}).get("verifier_commit")
                                if isinstance(previous, dict)
                                else None
                            )
                            if (
                                not isinstance(previous, dict)
                                or previous.get("state") != "verified"
                                or not is_git_sha(previous_commit)
                            ):
                                errors.append(
                                    f"prior batch is not verified before transition: {batch_id}"
                                )
                            else:
                                read_loose_commit(safe, str(previous_commit))
                                require_git_ancestor(
                                    safe, str(previous_commit), str(historical_commit)
                                )
                        errors.extend(
                            f"{item}: {batch_id}"
                            for item in predecessor_snapshot_errors(
                                manifest,
                                historical,
                                batch_index,
                                str(historical_commit),
                            )
                        )
                    except (ManifestError, UnicodeError, json.JSONDecodeError, zlib.error) as problem:
                        errors.append(f"historical manifest authentication failed: {batch_id}: {problem}")
        elif entry.get("before_manifest_path") is not None or entry.get("before_manifest_commit") is not None:
            errors.append(f"only first transition may bind a historical artifact: {batch_id}:{index}")
    if history:
        last = history[-1]
        if transition.get("manifest_before_sha256") != last.get("before_manifest_sha256"):
            errors.append(f"transition before digest mismatch: {batch_id}")
        projection = batch_projection_digest(manifest, batch_id)
        if transition.get("manifest_after_projection_sha256") != projection:
            errors.append(f"transition current projection mismatch: {batch_id}")
        if last.get("after_projection_sha256") != projection:
            errors.append(f"transition history projection mismatch: {batch_id}")
    return errors


def validate_consumer_transitions(
    manifest: dict[str, Any],
    batch: dict[str, Any],
    registry_by_id: dict[str, dict[str, Any]],
    safe: IndependentRepo,
) -> list[str]:
    errors: list[str] = []
    batch_id = str(batch.get("id"))
    state = batch.get("state")
    transitions = batch.get("consumer_transitions")
    if not isinstance(transitions, list):
        return [f"consumer transitions missing: {batch_id}"]
    expected_ids = [
        item
        for item in batch.get("affected_root_consumers", [])
        if item in registry_by_id and registry_by_id[item].get("mode") != "delete"
    ]
    if [item.get("consumer_id") for item in transitions if isinstance(item, dict)] != expected_ids:
        return [f"consumer transition membership/order mismatch: {batch_id}"]
    for item in transitions:
        if not isinstance(item, dict):
            errors.append(f"consumer transition is not an object: {batch_id}")
            continue
        consumer_id = item.get("consumer_id")
        consumer = registry_by_id.get(str(consumer_id))
        if consumer is None:
            errors.append(f"unknown consumer transition: {batch_id}")
            continue
        if state == "planned":
            if item.get("before_sha256") != consumer.get("before_sha256") or item.get("after_sha256") is not None:
                errors.append(f"planned consumer transition changed: {batch_id}:{consumer_id}")
            continue
        if not is_sha256(item.get("before_sha256")):
            errors.append(f"consumer before hash missing: {batch_id}:{consumer_id}")
        if consumer["mode"] == "self":
            expected_after = batch_projection_digest(manifest, batch_id)
        else:
            try:
                expected_after = safe.sha256(consumer["path"])
            except ManifestError as problem:
                errors.append(f"consumer after file invalid: {batch_id}:{consumer_id}: {problem}")
                continue
        if item.get("after_sha256") != expected_after:
            errors.append(f"consumer after hash mismatch: {batch_id}:{consumer_id}")
    return errors


def read_bound_artifact(
    safe: IndependentRepo,
    path: Any,
    digest: Any,
    label: str,
    *,
    max_bytes: int = 16 * 1024 * 1024,
    commit: str | None = None,
) -> tuple[bytes | None, list[str]]:
    errors = canonical_path_errors(path, label)
    if not is_sha256(digest):
        errors.append(f"{label} SHA-256 is invalid")
    if errors:
        return None, errors
    try:
        payload = safe.read(str(path), max_bytes)
    except ManifestError as problem:
        return None, [f"{label} is unreadable: {problem}"]
    if hashlib.sha256(payload).hexdigest() != digest:
        errors.append(f"{label} SHA-256 mismatch")
    if commit is not None:
        if not is_git_sha(commit):
            errors.append(f"{label} authenticated commit is invalid")
        else:
            try:
                committed_payload = read_git_tree_artifact(safe, commit, str(path))
                if committed_payload != payload:
                    errors.append(f"{label} differs from authenticated commit tree")
            except (ManifestError, UnicodeError, zlib.error) as problem:
                errors.append(f"{label} commit-tree authentication failed: {problem}")
    return payload, errors


def validate_verified_receipt(
    manifest: dict[str, Any],
    batch: dict[str, Any],
    safe: IndependentRepo,
) -> list[str]:
    """Require an external canonical receipt binding verifier, commits and evidence."""

    errors: list[str] = []
    verifier = batch.get("verifier_record")
    commit = batch.get("commit")
    if not isinstance(verifier, dict) or not isinstance(commit, dict):
        return ["verified receipt commit/verifier objects missing"]
    expected_verifier_keys = {
        "identity", "verdict", "verified_at", "receipt_path", "receipt_sha256"
    }
    if set(verifier) != expected_verifier_keys:
        errors.append(f"verified receipt pointer schema mismatch: {batch.get('id')}")
    payload, artifact_errors = read_bound_artifact(
        safe,
        verifier.get("receipt_path"),
        verifier.get("receipt_sha256"),
        f"verified receipt {batch.get('id')}",
        commit=(
            str(commit.get("verifier_commit"))
            if is_git_sha(commit.get("verifier_commit"))
            else None
        ),
    )
    errors.extend(artifact_errors)
    if payload is None:
        return errors
    try:
        receipt = json_load(payload, "verified receipt")
    except (ManifestError, UnicodeError, json.JSONDecodeError) as problem:
        return errors + [f"verified receipt JSON invalid: {problem}"]
    if canonical_json_bytes(receipt, pretty=True) != payload:
        errors.append(f"verified receipt is not canonical JSON: {batch.get('id')}")
    receipt_keys = {
        "schema_version", "manifest_id", "batch_id", "applied_commit",
        "verifier_commit", "verifier_identity", "verified_at",
        "operations_sha256", "transition_chain_sha256", "evidence_sha256",
        "commit_chain_sha256", "verifier_binding_sha256", "receipt_sha256",
    }
    if set(receipt) != receipt_keys:
        errors.append(f"verified receipt schema mismatch: {batch.get('id')}")
        return errors
    receipt_body = {key: value for key, value in receipt.items() if key != "receipt_sha256"}
    exact = {
        "schema_version": 1,
        "manifest_id": manifest.get("manifest_id"),
        "batch_id": batch.get("id"),
        "applied_commit": commit.get("applied_commit"),
        "verifier_commit": commit.get("verifier_commit"),
        "verifier_identity": verifier.get("identity"),
        "verified_at": verifier.get("verified_at"),
        "operations_sha256": batch.get("preconditions", {}).get("operations_sha256"),
        "transition_chain_sha256": object_digest(batch.get("state_transition", {}).get("history")),
        "evidence_sha256": object_digest(batch.get("evidence")),
        "commit_chain_sha256": object_digest({
            "applied_commit": commit.get("applied_commit"),
            "verifier_commit": commit.get("verifier_commit"),
        }),
        "verifier_binding_sha256": object_digest({
            "identity": verifier.get("identity"),
            "verdict": verifier.get("verdict"),
            "verified_at": verifier.get("verified_at"),
        }),
    }
    if receipt_body != exact:
        errors.append(f"verified receipt does not bind manifest state: {batch.get('id')}")
    if receipt.get("receipt_sha256") != object_digest(receipt_body):
        errors.append(f"verified receipt self-digest mismatch: {batch.get('id')}")
    return errors


def validate_evidence(
    manifest: dict[str, Any],
    batch: dict[str, Any],
    safe: IndependentRepo,
) -> list[str]:
    errors: list[str] = []
    batch_id = str(batch.get("id"))
    evidence = batch.get("evidence")
    if not isinstance(evidence, dict):
        return [f"evidence record missing: {batch_id}"]
    expected_evidence_keys = {
        "docker", "evidence_uri", "evidence_artifact_sha256",
        "source_revision", "cleanup",
    }
    if set(evidence) != expected_evidence_keys:
        errors.append(f"verified evidence aggregate schema mismatch: {batch_id}")
    commit = batch.get("commit")
    if not isinstance(commit, dict) or evidence.get("source_revision") != commit.get("applied_commit"):
        errors.append(f"verified evidence source revision mismatch: {batch_id}")
    authenticated_commit = (
        str(commit.get("verifier_commit"))
        if isinstance(commit, dict) and is_git_sha(commit.get("verifier_commit"))
        else None
    )
    docker = evidence.get("docker")
    if not isinstance(docker, list):
        return [f"verified Docker evidence missing: {batch_id}"]
    required_gate_ids = batch.get("docker_gate_ids")
    if [item.get("gate_id") for item in docker if isinstance(item, dict)] != required_gate_ids:
        errors.append(f"Docker evidence gate membership/order mismatch: {batch_id}")
    gates = manifest.get("gate_registry")
    gate_by_id = {
        item.get("id"): item for item in gates if isinstance(item, dict)
    } if isinstance(gates, list) else {}
    for item in docker:
        if not isinstance(item, dict):
            errors.append(f"Docker evidence entry malformed: {batch_id}")
            continue
        evidence_keys = {
            "gate_id", "image_ref", "image_digest", "immutable_image",
            "image_inspected_at", "image_lock_path", "image_lock_sha256",
            "source_revision", "gate_contract_sha256", "command", "task_run",
            "owner_marker_path", "owner_marker_sha256", "stdout_path",
            "stdout_sha256", "stderr_path", "stderr_sha256", "duration_ms",
            "collected", "pass", "fail", "skip", "cpus", "memory",
            "memory_swap", "pids_limit", "network", "read_only", "remove",
            "wall_timeout_seconds", "kill_grace_seconds", "task_name", "labels",
            "outer_cleanup", "exit_code", "cleanup",
        }
        if set(item) != evidence_keys:
            errors.append(f"Docker evidence schema mismatch: {batch_id}:{item.get('gate_id')}")
        digest = item.get("image_digest")
        if not isinstance(digest, str) or re.fullmatch(r"sha256:[0-9a-f]{64}", digest) is None:
            errors.append(f"Docker image digest invalid: {batch_id}")
        gate = gate_by_id.get(item.get("gate_id"))
        if not isinstance(gate, dict):
            errors.append(f"Docker evidence references unknown gate: {batch_id}")
            continue
        expected_digest = gate.get("expected_image_digest")
        if not isinstance(expected_digest, str) or re.fullmatch(r"sha256:[0-9a-f]{64}", expected_digest) is None:
            errors.append(
                "verified transition blocked until gate image digest is independently "
                f"sealed: {batch_id}:{item.get('gate_id')}"
            )
        elif digest != expected_digest:
            errors.append(
                f"Docker inspected digest differs from sealed gate digest: {batch_id}:{item.get('gate_id')}"
            )
        task_run = item.get("task_run")
        if not isinstance(task_run, str) or re.fullmatch(r"[a-z0-9][a-z0-9-]{5,79}", task_run) is None:
            errors.append(f"Docker task-run identity invalid: {batch_id}")
            continue
        template = gate.get("task_name_template")
        expected_name = template.replace("{task_run}", task_run) if isinstance(template, str) else None
        expected_labels = {
            str(key): str(value).replace("{task_run}", task_run)
            for key, value in gate.get("required_labels", {}).items()
        } if isinstance(gate.get("required_labels"), dict) else None
        exact = {
            "image_ref": gate.get("image"),
            "immutable_image": f"{gate.get('image')}@{expected_digest}",
            "source_revision": commit.get("applied_commit") if isinstance(commit, dict) else None,
            "gate_contract_sha256": object_digest(gate),
            "command": gate.get("entrypoints"),
            "cpus": gate.get("cpus"),
            "memory": gate.get("memory"),
            "memory_swap": gate.get("memory_swap"),
            "pids_limit": gate.get("pids_limit"),
            "network": gate.get("network"),
            "read_only": gate.get("read_only_repo"),
            "remove": gate.get("remove_container"),
            "wall_timeout_seconds": gate.get("wall_timeout_seconds"),
            "kill_grace_seconds": gate.get("kill_grace_seconds"),
            "task_name": expected_name,
            "labels": expected_labels,
            "outer_cleanup": gate.get("outer_cleanup"),
        }
        if any(item.get(key) != value for key, value in exact.items()):
            errors.append(f"Docker exact budget/identity evidence mismatch: {batch_id}:{item.get('gate_id')}")
        if item.get("exit_code") != 0 or item.get("fail") != 0:
            errors.append(f"Docker green evidence invalid: {batch_id}:{item.get('gate_id')}")
        counts = tuple(item.get(key) for key in ("collected", "pass", "fail", "skip"))
        if not all(isinstance(value, int) and value >= 0 for value in counts):
            errors.append(f"Docker result counts invalid: {batch_id}:{item.get('gate_id')}")
        elif item.get("collected") != item.get("pass") + item.get("fail") + item.get("skip"):
            errors.append(f"Docker result counts do not reconcile: {batch_id}:{item.get('gate_id')}")
        duration = item.get("duration_ms")
        timeout = gate.get("wall_timeout_seconds")
        if not isinstance(duration, int) or duration < 0 or not isinstance(timeout, int) or duration > timeout * 1000:
            errors.append(f"Docker duration exceeds wall budget: {batch_id}:{item.get('gate_id')}")
        if not isinstance(item.get("image_inspected_at"), str) or TIMESTAMP_RE.fullmatch(item["image_inspected_at"]) is None:
            errors.append(f"Docker image inspection timestamp invalid: {batch_id}:{item.get('gate_id')}")
        report_root = f"testframework/reports/{task_run}"
        expected_paths = {
            "owner_marker_path": f"{report_root}/.owner.json",
            "stdout_path": f"{report_root}/{item.get('gate_id')}.stdout.log",
            "stderr_path": f"{report_root}/{item.get('gate_id')}.stderr.log",
            "image_lock_path": f"{report_root}/{item.get('gate_id')}.image-lock.json",
        }
        if any(item.get(key) != value for key, value in expected_paths.items()):
            errors.append(f"Docker task-owned artifact path mismatch: {batch_id}:{item.get('gate_id')}")
        owner_payload, owner_errors = read_bound_artifact(
            safe, item.get("owner_marker_path"), item.get("owner_marker_sha256"),
            f"owner marker {batch_id}:{item.get('gate_id')}", max_bytes=4096,
            commit=authenticated_commit,
        )
        errors.extend(owner_errors)
        if owner_payload is not None:
            try:
                owner = json_load(owner_payload, "owner marker")
                if canonical_json_bytes(owner, pretty=True) != owner_payload or owner != {
                    "owner": "phase129",
                    "source_revision": commit.get("applied_commit") if isinstance(commit, dict) else None,
                    "task_run": task_run,
                }:
                    errors.append(f"owner marker content mismatch: {batch_id}:{item.get('gate_id')}")
            except (ManifestError, UnicodeError, json.JSONDecodeError) as problem:
                errors.append(f"owner marker JSON invalid: {batch_id}:{problem}")
        lock_payload, lock_errors = read_bound_artifact(
            safe, item.get("image_lock_path"), item.get("image_lock_sha256"),
            f"image lock {batch_id}:{item.get('gate_id')}", max_bytes=4096,
            commit=authenticated_commit,
        )
        errors.extend(lock_errors)
        if lock_payload is not None:
            try:
                lock = json_load(lock_payload, "image lock")
                if canonical_json_bytes(lock, pretty=True) != lock_payload or lock != {
                    "actual_image_digest": digest,
                    "expected_image_digest": expected_digest,
                    "image_ref": gate.get("image"),
                    "inspected_at": item.get("image_inspected_at"),
                    "source_revision": commit.get("applied_commit") if isinstance(commit, dict) else None,
                }:
                    errors.append(f"image lock content mismatch: {batch_id}:{item.get('gate_id')}")
            except (ManifestError, UnicodeError, json.JSONDecodeError) as problem:
                errors.append(f"image lock JSON invalid: {batch_id}:{problem}")
        for stream in ("stdout", "stderr"):
            _, stream_errors = read_bound_artifact(
                safe, item.get(f"{stream}_path"), item.get(f"{stream}_sha256"),
                f"raw {stream} {batch_id}:{item.get('gate_id')}",
                commit=authenticated_commit,
            )
            errors.extend(stream_errors)
        cleanup = item.get("cleanup")
        if not isinstance(cleanup, dict) or set(cleanup) != {
            "container_absent", "labels_absent", "inspected_at",
            "proof_path", "proof_sha256",
        }:
            errors.append(f"Docker cleanup proof schema mismatch: {batch_id}:{item.get('gate_id')}")
        else:
            if cleanup.get("container_absent") is not True or cleanup.get("labels_absent") is not True:
                errors.append(f"Docker exact cleanup not proven: {batch_id}:{item.get('gate_id')}")
            if cleanup.get("proof_path") != f"{report_root}/{item.get('gate_id')}.cleanup.json":
                errors.append(f"Docker cleanup proof path mismatch: {batch_id}:{item.get('gate_id')}")
            if not isinstance(cleanup.get("inspected_at"), str) or TIMESTAMP_RE.fullmatch(cleanup["inspected_at"]) is None:
                errors.append(f"Docker cleanup timestamp invalid: {batch_id}:{item.get('gate_id')}")
            cleanup_payload, cleanup_errors = read_bound_artifact(
                safe, cleanup.get("proof_path"), cleanup.get("proof_sha256"),
                f"cleanup proof {batch_id}:{item.get('gate_id')}", max_bytes=4096,
                commit=authenticated_commit,
            )
            errors.extend(cleanup_errors)
            if cleanup_payload is not None:
                try:
                    cleanup_proof = json_load(cleanup_payload, "gate cleanup proof")
                    expected_cleanup_proof = {
                        "container_absent": True,
                        "exit_code": item.get("exit_code"),
                        "gate_id": item.get("gate_id"),
                        "image_digest": digest,
                        "inspected_at": cleanup.get("inspected_at"),
                        "labels": item.get("labels"),
                        "labels_absent": True,
                        "schema_version": 1,
                        "source_revision": commit.get("applied_commit") if isinstance(commit, dict) else None,
                        "task_name": item.get("task_name"),
                        "task_run": task_run,
                    }
                    if canonical_json_bytes(cleanup_proof, pretty=True) != cleanup_payload or cleanup_proof != expected_cleanup_proof:
                        errors.append(f"gate cleanup proof content mismatch: {batch_id}:{item.get('gate_id')}")
                except (ManifestError, UnicodeError, json.JSONDecodeError) as problem:
                    errors.append(f"gate cleanup proof JSON invalid: {batch_id}:{problem}")
    task_runs = {item.get("task_run") for item in docker if isinstance(item, dict)}
    if len(task_runs) != 1:
        errors.append(f"verified evidence must use one task-run: {batch_id}")
    task_run = next(iter(task_runs), None)
    if isinstance(task_run, str):
        expected_evidence_uri = f"testframework/reports/{task_run}/evidence.json"
        if evidence.get("evidence_uri") != expected_evidence_uri:
            errors.append(f"evidence URI is not task-owned: {batch_id}")
    evidence_payload, evidence_errors = read_bound_artifact(
        safe,
        evidence.get("evidence_uri"),
        evidence.get("evidence_artifact_sha256"),
        f"aggregate evidence {batch_id}",
        commit=authenticated_commit,
    )
    errors.extend(evidence_errors)
    if evidence_payload is not None and isinstance(task_run, str):
        try:
            evidence_artifact = json_load(evidence_payload, "aggregate evidence")
            expected_evidence_artifact = {
                "batch_id": batch.get("id"),
                "cleanup_proof_sha256": evidence.get("cleanup", {}).get("proof_sha256"),
                "gates": [
                    {
                        "cleanup_proof_sha256": item.get("cleanup", {}).get("proof_sha256"),
                        "duration_ms": item.get("duration_ms"),
                        "exit_code": item.get("exit_code"),
                        "gate_id": item.get("gate_id"),
                        "image_digest": item.get("image_digest"),
                        "stderr_sha256": item.get("stderr_sha256"),
                        "stdout_sha256": item.get("stdout_sha256"),
                    }
                    for item in docker if isinstance(item, dict)
                ],
                "manifest_id": manifest.get("manifest_id"),
                "result": "PASS",
                "schema_version": 1,
                "source_revision": commit.get("applied_commit") if isinstance(commit, dict) else None,
                "task_run": task_run,
            }
            if canonical_json_bytes(evidence_artifact, pretty=True) != evidence_payload or evidence_artifact != expected_evidence_artifact:
                errors.append(f"aggregate evidence content mismatch: {batch_id}")
        except (ManifestError, UnicodeError, json.JSONDecodeError) as problem:
            errors.append(f"aggregate evidence JSON invalid: {batch_id}:{problem}")
    cleanup = evidence.get("cleanup")
    if not isinstance(cleanup, dict) or set(cleanup) != {
        "all_resources_absent", "inspected_at", "proof_path", "proof_sha256"
    }:
        errors.append(f"aggregate cleanup proof schema mismatch: {batch_id}")
    else:
        if cleanup.get("all_resources_absent") is not True:
            errors.append(f"aggregate cleanup not proven: {batch_id}")
        if isinstance(task_run, str) and cleanup.get("proof_path") != f"testframework/reports/{task_run}/cleanup.json":
            errors.append(f"aggregate cleanup proof path mismatch: {batch_id}")
        if not isinstance(cleanup.get("inspected_at"), str) or TIMESTAMP_RE.fullmatch(cleanup["inspected_at"]) is None:
            errors.append(f"aggregate cleanup timestamp invalid: {batch_id}")
        aggregate_cleanup_payload, cleanup_errors = read_bound_artifact(
            safe, cleanup.get("proof_path"), cleanup.get("proof_sha256"),
            f"aggregate cleanup artifact {batch_id}", max_bytes=4096,
            commit=authenticated_commit,
        )
        errors.extend(cleanup_errors)
        if aggregate_cleanup_payload is not None and isinstance(task_run, str):
            try:
                aggregate_cleanup = json_load(aggregate_cleanup_payload, "aggregate cleanup proof")
                expected_aggregate_cleanup = {
                    "all_resources_absent": True,
                    "inspected_at": cleanup.get("inspected_at"),
                    "labels_absent": True,
                    "schema_version": 1,
                    "source_revision": commit.get("applied_commit") if isinstance(commit, dict) else None,
                    "task_run": task_run,
                }
                if canonical_json_bytes(aggregate_cleanup, pretty=True) != aggregate_cleanup_payload or aggregate_cleanup != expected_aggregate_cleanup:
                    errors.append(f"aggregate cleanup proof content mismatch: {batch_id}")
            except (ManifestError, UnicodeError, json.JSONDecodeError) as problem:
                errors.append(f"aggregate cleanup proof JSON invalid: {batch_id}:{problem}")
    return errors


def validate_batch_state(
    manifest: dict[str, Any],
    batch: dict[str, Any],
    registry_by_id: dict[str, dict[str, Any]],
    safe: IndependentRepo,
    *,
    batch_index: int,
    check_filesystem: bool,
) -> list[str]:
    errors: list[str] = []
    batch_id = str(batch.get("id"))
    state = batch.get("state")
    if state not in {"planned", "applied-unverified", "verified"}:
        return [f"invalid state: {batch_id}"]
    errors.extend(
        validate_transition_history(batch, str(state), manifest, safe, batch_index)
    )
    errors.extend(validate_consumer_transitions(manifest, batch, registry_by_id, safe))
    commit = batch.get("commit")
    verifier = batch.get("verifier_record")
    rollback = batch.get("rollback")
    evidence = batch.get("evidence")
    if not isinstance(commit, dict) or not isinstance(verifier, dict) or not isinstance(rollback, dict) or not isinstance(evidence, dict):
        return errors + [f"state evidence objects missing: {batch_id}"]
    if state == "planned":
        expected_commit = {"applied_commit": None, "verifier_commit": None}
        expected_verifier = {"identity": None, "verdict": None, "verified_at": None}
        expected_evidence = {"docker": None, "evidence_uri": None, "cleanup": None}
        expected_rollback = {
            "method": "git-revert-exact-commit",
            "commit_to_revert": None,
            "revert_commit": None,
        }
        if commit != expected_commit or verifier != expected_verifier or evidence != expected_evidence or rollback != expected_rollback:
            errors.append(f"planned commit/verifier/evidence/rollback slots must be null: {batch_id}")
    else:
        if any(operation.get("action") == "keep-blocked" for operation in batch.get("operations", []) if isinstance(operation, dict)):
            errors.append(f"keep-blocked batch cannot transition without regenerated authority: {batch_id}")
        if not is_git_sha(commit.get("applied_commit")):
            errors.append(f"applied commit missing: {batch_id}")
        if rollback.get("method") != "git-revert-exact-commit" or rollback.get("commit_to_revert") != commit.get("applied_commit"):
            errors.append(f"rollback slot does not identify the applied commit: {batch_id}")
        if rollback.get("revert_commit") is not None and not is_git_sha(rollback.get("revert_commit")):
            errors.append(f"revert commit field invalid: {batch_id}")
        if state == "applied-unverified":
            if commit.get("verifier_commit") is not None or verifier != {"identity": None, "verdict": None, "verified_at": None}:
                errors.append(f"applied-unverified batch contains verifier claim: {batch_id}")
        else:
            if not is_git_sha(commit.get("verifier_commit")):
                errors.append(f"verifier commit missing: {batch_id}")
            if verifier.get("identity") not in batch.get("verifiers", []) or verifier.get("identity") == batch.get("owner"):
                errors.append(f"independent verifier identity invalid: {batch_id}")
            if verifier.get("verdict") != "PASS":
                errors.append(f"independent verifier verdict missing: {batch_id}")
            if not isinstance(verifier.get("verified_at"), str) or TIMESTAMP_RE.fullmatch(verifier["verified_at"]) is None:
                errors.append(f"verifier timestamp invalid: {batch_id}")
            errors.extend(validate_evidence(manifest, batch, safe))
            errors.extend(validate_verified_receipt(manifest, batch, safe))

        if check_filesystem and is_git_sha(commit.get("applied_commit")):
            try:
                applied_commit = str(commit["applied_commit"])
                read_loose_commit(safe, applied_commit)
                planned_source = manifest.get("source", {}).get("repository_head")
                if not is_git_sha(planned_source):
                    raise ManifestError("manifest planned source commit is invalid")
                if planned_source == applied_commit:
                    raise ManifestError("applied commit must advance beyond planned source commit")
                read_loose_commit(safe, str(planned_source))
                require_git_ancestor(safe, str(planned_source), applied_commit)
                live_head = read_git_head(safe)
                if state == "applied-unverified":
                    if live_head != applied_commit:
                        raise ManifestError("live HEAD is not the applied commit")
                else:
                    verifier_commit = commit.get("verifier_commit")
                    if not is_git_sha(verifier_commit) or verifier_commit == applied_commit:
                        raise ManifestError("verified commit must be distinct from applied commit")
                    read_loose_commit(safe, str(verifier_commit))
                    require_git_ancestor(safe, applied_commit, str(verifier_commit))
                    later_active = any(
                        isinstance(item, dict) and item.get("state") != "planned"
                        for item in manifest.get("batches", [])[batch_index + 1 :]
                    )
                    if later_active:
                        require_git_ancestor(safe, str(verifier_commit), live_head)
                    elif live_head != verifier_commit:
                        raise ManifestError("live HEAD is not the latest verifier commit")
            except (ManifestError, UnicodeError, zlib.error) as problem:
                errors.append(f"descriptor-safe Git commit-chain validation failed: {batch_id}: {problem}")

    if not check_filesystem:
        return errors
    patterns = batch.get("stale_reference_scan", {}).get("exact_patterns", [])
    for operation in batch.get("operations", []):
        if not isinstance(operation, dict):
            continue
        source = operation.get("source")
        source_sha = operation.get("source_sha256")
        target = operation.get("target")
        action = operation.get("action")
        try:
            if state == "planned":
                if safe.sha256(source) != source_sha:
                    errors.append(f"planned source hash mismatch: {batch_id}:{source}")
                if action == "move":
                    safe.assert_absent(target)
            elif action == "move":
                safe.assert_absent(source)
                if safe.sha256(target) != source_sha:
                    errors.append(f"applied move target hash mismatch: {batch_id}:{target}")
            elif action == "delete":
                safe.assert_absent(source)
        except ManifestError as problem:
            errors.append(f"filesystem state mismatch: {batch_id}: {problem}")
    if state != "planned":
        for consumer_id in batch.get("affected_root_consumers", []):
            consumer = registry_by_id.get(consumer_id)
            if consumer is None or consumer.get("mode") != "edit":
                continue
            try:
                payload = safe.read(consumer["path"])
            except ManifestError as problem:
                errors.append(f"stale-reference root unreadable: {batch_id}:{problem}")
                continue
            for pattern in patterns:
                if pattern.encode("utf-8") in payload:
                    errors.append(f"stale-reference declaration not cleared: {batch_id}:{consumer_id}")
                    break
    return errors


def _row_group(row: dict[str, Any]) -> tuple[str, str] | None:
    runner = row.get("runner")
    target = row.get("target_class")
    consumers = row.get("hook_consumers")
    if runner == "go-test" and target == "go-unit" and isinstance(consumers, list):
        return ("MOVE-GO-UNIT-HOOK-FREE", "move") if not consumers else ("KEEP-GO-UNIT-HOOK-DEPENDENT", "keep-blocked")
    mapping = {
        ("go-test", "python-api"): ("DELETE-GO-COMPONENT", "delete"),
        ("go-test", "python-integration"): ("DELETE-GO-COMPONENT", "delete"),
        ("vitest", "frontend-unit"): ("MOVE-FRONTEND-UNIT", "move"),
        ("vitest", "python-static-contract"): ("DELETE-FRONTEND-STATIC-CONTRACT", "delete"),
        ("playwright-ts", "python-e2e"): ("DELETE-PLAYWRIGHT-TS-SPECS", "delete"),
        ("playwright-support-ts", "python-e2e-support"): ("DELETE-PLAYWRIGHT-TS-SUPPORT", "delete"),
    }
    if (runner, target) in mapping:
        return mapping[(runner, target)]
    if runner == "vitest-support" and row.get("source_path") == "testframework/stages/unit/frontend-test/render.tsx":
        return "MOVE-FRONTEND-RENDER-SUPPORT", "move"
    if runner == "vitest-support" and row.get("source_path") in {
        "testframework/stages/unit/frontend-test/domains/platform/updateBase-version-required.fixture.ts",
        "testframework/stages/unit/frontend-test/domains/tracker/projects-sprint-version-required.fixture.ts",
    }:
        return "DELETE-FRONTEND-COMPILE-FIXTURES", "delete"
    return None


def _expected_gate_contract() -> dict[str, dict[str, Any]]:
    common = {
        "architecture-contract": (
            "python:3.12-slim", "1.0", "1g", "1g", 128, 300,
            "hub-129-architecture-contract-{task_run}",
            [
                "python3 .planning/audits/test-rebase/validate_cutover_manifest.py --check .planning/audits/test-rebase/cutover-manifest.json",
                "python3 .planning/audits/test-rebase/validate_cutover_manifest.py --stdin --self-test",
            ],
        ),
        "backend-unit": (
            "golang:1.25-bookworm", "2.0", "4g", "4g", 256, 720,
            "hub-129-backend-unit-{task_run}", [
                "go build -tags release ./...",
                "go vet ./...",
                "go test -race -count=1 -timeout=600s ./tests/...",
            ]
        ),
        "frontend-unit": (
            "mcr.microsoft.com/playwright:v1.61.0-noble", "2.0", "4g", "4g", 256, 1200,
            "hub-129-frontend-unit-{task_run}", [
                "npx tsc --noEmit", "npm run lint", "npx vitest run tests", "npm run build",
            ]
        ),
        "python-scaffold": (
            "hub-testframework:phase129", "2.0", "4g", "4g", 256, 600,
            "hub-129-python-scaffold-{task_run}",
            [f"test -f {path}" for path in PYTHON_SCAFFOLD_REQUIRED_FILES]
            + [f"test ! -e {path}" for path in PYTHON_SCAFFOLD_FORBIDDEN_PATHS]
            + [
                "python3 -m pytest --collect-only -q testframework/selftests testframework/tests/api",
                "python3 -m pytest -q -m scaffold testframework/selftests",
            ],
        ),
    }
    return {
        gate_id: {
            "image": image,
            "expected_image_digest": None,
            "cpus": cpu,
            "memory": memory,
            "memory_swap": swap,
            "pids_limit": pids,
            "wall_timeout_seconds": timeout,
            "kill_grace_seconds": 10,
            "network": "none",
            "read_only_repo": True,
            "remove_container": True,
            "task_name_template": template,
            "required_labels": {"hub.task-run": "{task_run}", "hub.owner": "phase129"},
            "outer_cleanup": "finally-exact-name-and-label",
            "entrypoints": entrypoints,
        }
        for gate_id, (image, cpu, memory, swap, pids, timeout, template, entrypoints) in common.items()
    }


def _independent_slug(value: str) -> str:
    normalized = unicodedata.normalize("NFKD", value).encode("ascii", "ignore").decode("ascii").lower()
    result = re.sub(r"[^a-z0-9]+", "-", normalized).strip("-")
    if not result or ASCII_SLUG_RE.fullmatch(result) is None:
        raise ManifestError(f"independent domain cannot form stable slug: {value!r}")
    return result


def _nonledger_domain(path: str) -> str:
    for prefix, domain in (
        ("backend/testframework/", "backend-support"),
        ("frontend/", "frontend"),
        ("scripts/", "root-scripts"),
        ("testframework/pipeline/", "pipeline"),
        ("testframework/stages/", "legacy-stages"),
        ("testframework/", "legacy-docs"),
    ):
        if path.startswith(prefix):
            return domain
    raise ManifestError(f"independent non-ledger domain missing: {path}")


def validate_independent_authority(
    actual: dict[str, Any],
    repo_root: Path,
    *,
    verify_live_sources: bool,
) -> list[str]:
    """Reconcile operations from sealed inputs without builder imports/derivation."""

    errors: list[str] = []
    safe = IndependentRepo(repo_root)
    provided_source_sha: dict[str, Any] = {}
    for batch in actual.get("batches", []) if isinstance(actual.get("batches"), list) else []:
        if not isinstance(batch, dict):
            continue
        for operation in batch.get("operations", []) if isinstance(batch.get("operations"), list) else []:
            if isinstance(operation, dict) and isinstance(operation.get("source"), str):
                provided_source_sha[operation["source"]] = operation.get("source_sha256")
    loaded: dict[str, dict[str, Any]] = {}
    for name, (path, digest) in AUTHORITY_FILES.items():
        try:
            payload = safe.read(path, 8 * 1024 * 1024)
            if hashlib.sha256(payload).hexdigest() != digest:
                errors.append(f"independent sealed authority mismatch: {name}")
            loaded[name] = json_load(payload, name)
        except (ManifestError, UnicodeError, json.JSONDecodeError) as problem:
            errors.append(f"independent authority read failed: {name}: {problem}")
            return errors
    inventory = loaded["inventory"]
    founder = loaded["founder_manifest"]
    if inventory.get("revision") != LEDGER_REVISION or not isinstance(inventory.get("rows"), list):
        return errors + ["independent inventory revision/rows changed"]
    rows = inventory["rows"]
    if len(rows) != 1049:
        errors.append("independent inventory row count changed")

    expected: dict[str, dict[str, Any]] = {}
    placement: dict[str, tuple[str, str, str, dict[str, Any]]] = {}
    group_meta: dict[str, tuple[str, str, int, list[str], str, list[str], list[str]]] = {
        "MOVE-GO-UNIT-HOOK-FREE": ("129-02", "go-hook-free", 2, ["TASK-197", "TASK-201"], "be1", ["be-ver"], ["architecture-contract", "backend-unit"]),
        "MOVE-FRONTEND-UNIT": ("129-03", "frontend-unit", 3, ["TASK-198", "TASK-201"], "fe1", ["fe-ver"], ["architecture-contract", "frontend-unit"]),
        "MOVE-FRONTEND-RENDER-SUPPORT": ("129-03", "frontend-unit", 3, ["TASK-198", "TASK-201"], "fe1", ["fe-ver"], ["architecture-contract", "frontend-unit"]),
        "DELETE-GO-COMPONENT": ("129-04", "go-component-delete", 4, ["TASK-219", "TASK-201"], "be1", ["code-reviewer", "qa-auto-ver"], ["architecture-contract", "backend-unit"]),
        "DELETE-FRONTEND-STATIC-CONTRACT": ("129-05", "static", 5, ["TASK-219", "TASK-201"], "fe1", ["qa-auto-ver"], ["architecture-contract", "frontend-unit", "python-scaffold"]),
        "DELETE-FRONTEND-COMPILE-FIXTURES": ("129-05", "support", 5, ["TASK-219", "TASK-201"], "fe1", ["qa-auto-ver"], ["architecture-contract", "frontend-unit", "python-scaffold"]),
        "DELETE-PLAYWRIGHT-TS-SPECS": ("129-05", "e2e", 5, ["TASK-219", "TASK-201"], "fe1", ["qa-auto-ver"], ["architecture-contract", "frontend-unit", "python-scaffold"]),
        "DELETE-PLAYWRIGHT-TS-SUPPORT": ("129-05", "support", 5, ["TASK-219", "TASK-201"], "fe1", ["qa-auto-ver"], ["architecture-contract", "frontend-unit", "python-scaffold"]),
        "KEEP-GO-UNIT-HOOK-DEPENDENT": ("129-06", "go-hook-blocked", 6, ["TASK-234", "TASK-201"], "be2", ["be-ver", "arch-ver"], ["architecture-contract", "backend-unit"]),
    }
    for index, row in enumerate(rows):
        if not isinstance(row, dict):
            errors.append(f"independent row {index} malformed")
            continue
        source = row.get("source_path")
        if canonical_path_errors(source, f"independent row {index}"):
            errors.extend(canonical_path_errors(source, f"independent row {index}"))
            continue
        classification = _row_group(row)
        if classification is None:
            errors.append(f"independent row has no unique group: {source}")
            continue
        group, action = classification
        operation: dict[str, Any] = {
            "action": action,
            "authority_group": group,
            "ledger_source": True,
            "source": source,
            "source_sha256": row.get("sha256"),
        }
        if action == "move":
            operation["target"] = row.get("target_path")
        elif action == "keep-blocked":
            links = row.get("hook_consumers", [])
            operation["hook_links"] = sorted(
                ({"hook_source": item.get("hook_source"), "symbol": item.get("symbol")} for item in links if isinstance(item, dict)),
                key=lambda item: (str(item["hook_source"]), str(item["symbol"])),
            )
            operation["release_condition"] = "regenerated-zero-link-authority-and-docker-green"
        expected[str(source)] = operation
        plan, kind, wave, tasks, owner, verifiers, gates = group_meta[group]
        placement[str(source)] = (
            plan, kind, _independent_slug(str(row.get("domain", ""))),
            {"wave": wave, "tasks": tasks, "owner": owner, "verifiers": verifiers, "gates": gates},
        )

    nonledger = founder.get("non_ledger_delete_paths")
    if not isinstance(nonledger, list) or len(nonledger) != 28:
        errors.append("independent non-ledger closure changed")
        nonledger = []
    for source in nonledger:
        if not isinstance(source, str):
            continue
        if verify_live_sources:
            try:
                source_sha = safe.sha256(source)
            except MissingPathError:
                source_sha = None
            except ManifestError as problem:
                errors.append(f"independent non-ledger source unsafe: {source}: {problem}")
                source_sha = None
        else:
            source_sha = provided_source_sha.get(source)
        expected[source] = {
            "action": "delete",
            "authority_group": "DELETE-NONLEDGER-CLOSURE",
            "ledger_source": False,
            "source": source,
            "source_sha256": source_sha,
        }
        if source.startswith("backend/testframework/"):
            placement[source] = (
                "129-04", "go-component-delete", _nonledger_domain(source),
                {"wave": 4, "tasks": ["TASK-219", "TASK-201"], "owner": "be1", "verifiers": ["code-reviewer", "qa-auto-ver"], "gates": ["architecture-contract", "backend-unit"]},
            )
        else:
            placement[source] = (
                "129-05", "support", _nonledger_domain(source),
                {"wave": 5, "tasks": ["TASK-219", "TASK-201"], "owner": "fe1", "verifiers": ["qa-auto-ver"], "gates": ["architecture-contract", "frontend-unit", "python-scaffold"]},
            )
    hooks = inventory.get("hooks")
    if not isinstance(hooks, list) or len(hooks) != 43:
        errors.append("independent hook guard changed")
        hooks = []
    for hook in hooks:
        if not isinstance(hook, dict) or not isinstance(hook.get("source_path"), str):
            continue
        source = hook["source_path"]
        links = []
        for symbol in hook.get("symbols", []):
            if not isinstance(symbol, dict):
                continue
            for consumer in symbol.get("consumer_test_files", []):
                links.append({"consumer_source": consumer, "symbol": symbol.get("name")})
        expected[source] = {
            "action": "keep-blocked",
            "authority_group": "ZERO-CONSUMER-HOOK-GUARD",
            "ledger_source": False,
            "source": source,
            "source_sha256": hook.get("sha256"),
            "consumer_links": sorted(links, key=lambda item: (str(item["consumer_source"]), str(item["symbol"]))),
            "release_condition": "regenerated-exact-zero-consumer-authority",
        }
        placement[source] = (
            "129-07", "zero-consumer", _independent_slug(str(hook.get("module", ""))),
            {"wave": 7, "tasks": ["TASK-219", "TASK-201"], "owner": "be2", "verifiers": ["be-ver", "arch-ver", "qa-auto-ver"], "gates": ["architecture-contract", "backend-unit"]},
        )

    observed: dict[str, dict[str, Any]] = {}
    for batch in actual.get("batches", []) if isinstance(actual.get("batches"), list) else []:
        if not isinstance(batch, dict):
            continue
        for operation in batch.get("operations", []) if isinstance(batch.get("operations"), list) else []:
            if isinstance(operation, dict) and isinstance(operation.get("source"), str):
                if operation["source"] in observed:
                    errors.append(f"independent duplicate operation: {operation['source']}")
                observed[operation["source"]] = operation
    if set(observed) != set(expected):
        errors.append("independent operation source union differs from sealed 1,120-operation authority")
    planned_only = all(
        isinstance(batch, dict) and batch.get("state") == "planned"
        for batch in actual.get("batches", []) if isinstance(actual.get("batches"), list)
    )
    for source in sorted(set(observed).intersection(expected)):
        wanted = expected[source]
        actual_op = observed[source]
        if not planned_only and wanted.get("authority_group") == "DELETE-NONLEDGER-CLOSURE":
            wanted = dict(wanted)
            wanted["source_sha256"] = actual_op.get("source_sha256")
            expected[source] = wanted
        if actual_op != wanted:
            errors.append(f"independent operation mismatch: {source}")

    grouped: dict[tuple[str, str, str], list[dict[str, Any]]] = {}
    metadata: dict[tuple[str, str, str], dict[str, Any]] = {}
    for source, operation in expected.items():
        plan, kind, domain, meta = placement[source]
        key = (plan, kind, domain)
        grouped.setdefault(key, []).append(operation)
        metadata[key] = meta
    plan_order = {"129-02": 2, "129-03": 3, "129-04": 4, "129-05": 5, "129-06": 6, "129-07": 7}
    expected_batches: list[dict[str, Any]] = []
    for key in sorted(grouped, key=lambda item: (plan_order[item[0]], item[1], item[2])):
        plan, kind, domain = key
        operations = sorted(grouped[key], key=lambda item: str(item["source"]))
        chunks = [operations[index:index + MAX_BATCH_OPERATIONS] for index in range(0, len(operations), MAX_BATCH_OPERATIONS)]
        for chunk_index, chunk in enumerate(chunks, 1):
            base = f"{plan}-{kind}-{domain}"
            expected_batches.append({
                "id": f"{base}-{chunk_index:02d}" if len(chunks) > 1 else base,
                "plan": plan,
                "wave": metadata[key]["wave"],
                "tasks": metadata[key]["tasks"],
                "owner": metadata[key]["owner"],
                "verifiers": metadata[key]["verifiers"],
                "domain": domain,
                "authority_groups": sorted({str(item["authority_group"]) for item in chunk}),
                "operations": chunk,
                "docker_gate_ids": metadata[key]["gates"],
            })
    actual_batches = actual.get("batches") if isinstance(actual.get("batches"), list) else []
    if len(actual_batches) != len(expected_batches):
        errors.append("independent batch count differs from literal 64-operation partition")
    for order, (batch, wanted) in enumerate(zip(actual_batches, expected_batches), 1):
        if not isinstance(batch, dict):
            continue
        projection = {key: batch.get(key) for key in wanted}
        if projection != wanted or batch.get("order") != order or batch.get("phase") != 129:
            errors.append(f"independent batch membership/metadata mismatch: order {order}")

    edit = founder.get("edit_in_place_paths")
    consumer_paths: list[str] = []
    roles: dict[str, str] = {}
    if isinstance(edit, dict):
        for key, role in (
            ("runtime_and_ci", "runtime-ci-config"),
            ("root_testframework_docs_rewrite", "testframework-doc"),
            ("current_project_docs_sync", "current-project-doc"),
        ):
            values = edit.get(key)
            if not isinstance(values, list):
                errors.append(f"independent founder consumer list missing: {key}")
                continue
            for value in values:
                if isinstance(value, str):
                    consumer_paths.append(value)
                    roles[value] = role
    for source in nonledger:
        if isinstance(source, str):
            consumer_paths.append(source)
            roles[source] = "delete-operation"
    expected_registry_shape = [
        {"id": f"consumer-{index + 1:03d}", "path": path, "role": roles[path], "mode": "delete" if path in nonledger else "edit"}
        for index, path in enumerate(sorted(consumer_paths))
    ] + [{"id": "consumer-manifest-self", "path": ".planning/audits/test-rebase/cutover-manifest.json", "role": "cutover-manifest-self", "mode": "self"}]
    registry = actual.get("consumer_registry") if isinstance(actual.get("consumer_registry"), list) else []
    if [{key: item.get(key) for key in ("id", "path", "role", "mode")} for item in registry if isinstance(item, dict)] != expected_registry_shape:
        errors.append("independent consumer registry membership/order mismatch")
    for item in registry:
        if not isinstance(item, dict):
            continue
        if item.get("mode") == "self":
            if item.get("before_sha256") is not None:
                errors.append("independent manifest-self before hash must be null")
        elif planned_only and verify_live_sources:
            try:
                if item.get("before_sha256") != safe.sha256(str(item.get("path"))):
                    errors.append(f"independent live consumer hash mismatch: {item.get('id')}")
            except ManifestError as problem:
                errors.append(f"independent consumer unreadable: {item.get('id')}: {problem}")
        elif not is_sha256(item.get("before_sha256")):
            errors.append(f"independent historical consumer hash invalid: {item.get('id')}")
    if len(consumer_paths) != 65 or len(set(consumer_paths)) != 65:
        errors.append("independent consumer registry closure changed")
    expected_registry_by_id = {item["id"]: item for item in expected_registry_shape}
    for batch, wanted in zip(actual_batches, expected_batches):
        if not isinstance(batch, dict):
            continue
        operation_sources = {str(item["source"]) for item in wanted["operations"]}
        expected_consumers: list[str] = []
        for item in expected_registry_shape:
            include = item["role"] == "cutover-manifest-self"
            if item["mode"] == "delete" and item["path"] in operation_sources:
                include = True
            if item["mode"] == "edit":
                if item["role"] in {"current-project-doc", "testframework-doc"}:
                    include = True
                elif item["role"] == "runtime-ci-config" and (
                    item["path"] in {".github/workflows/ci.yml", ".gitlab-ci.yml"}
                    or wanted["plan"] in {"129-03", "129-05"}
                ):
                    include = True
            if include:
                expected_consumers.append(item["id"])
        if batch.get("affected_root_consumers") != expected_consumers:
            errors.append(f"independent affected-consumer set mismatch: {wanted['id']}")
        preconditions = batch.get("preconditions")
        if not isinstance(preconditions, dict):
            continue
        if preconditions.get("repository_head") != actual.get("source", {}).get("repository_head"):
            errors.append(f"independent batch repository head mismatch: {wanted['id']}")
        expected_authority = {name: digest for name, (_, digest) in AUTHORITY_FILES.items()}
        if preconditions.get("authority_sha256") != expected_authority:
            errors.append(f"independent batch authority seal mismatch: {wanted['id']}")
        # Static validation recomputes this digest from the full live registry;
        # membership is independently fixed above so omissions cannot hide.
        if any(consumer_id not in expected_registry_by_id for consumer_id in expected_consumers):
            errors.append(f"independent unknown expected consumer: {wanted['id']}")

    gates = actual.get("gate_registry")
    gate_by_id = {item.get("id"): item for item in gates if isinstance(item, dict)} if isinstance(gates, list) else {}
    expected_gates = _expected_gate_contract()
    if set(gate_by_id) != set(expected_gates):
        errors.append("independent gate IDs changed")
    for gate_id, required in expected_gates.items():
        gate = gate_by_id.get(gate_id, {})
        if set(gate) != {"id"}.union(required):
            errors.append(f"independent gate schema mismatch: {gate_id}")
        for key, value in required.items():
            if gate.get(key) != value:
                errors.append(f"independent gate budget mismatch: {gate_id}:{key}")
    if actual.get("totals") != {
        "batches": len(expected_batches),
        "operations": len(expected),
        "actions": {"move": 507, "delete": 325, "keep-blocked": 288},
        "consumer_registry_entries": 66,
    }:
        errors.append("independent manifest totals changed")
    source = actual.get("source")
    expected_authority_files = [
        {"id": name, "path": path, "sha256": digest}
        for name, (path, digest) in AUTHORITY_FILES.items()
    ]
    if not isinstance(source, dict) or source.get("ledger_revision") != LEDGER_REVISION or source.get("authority_files") != expected_authority_files:
        errors.append("independent top-level source authority changed")
    if actual.get("invariants") != {
        "max_batch_operations": 64,
        "ledger_rows_exact_once": 1049,
        "ledger_moves": 507,
        "ledger_keep_blocked": 245,
        "ledger_deletes": 297,
        "nonledger_deletes": 28,
        "hook_files_keep_blocked": 43,
        "protected_prefix": PROTECTED_PREFIX,
        "no_command_execution_from_manifest": True,
        "state_machine": ["planned", "applied-unverified", "verified"],
    }:
        errors.append("independent manifest invariants changed")
    if actual.get("schema_version") != 1 or actual.get("manifest_id") != "PHASE-129-FOUNDER-GREENFIELD-EXACT-CUTOVER":
        errors.append("independent manifest identity changed")
    try:
        safe.finalize()
    except ManifestError as problem:
        errors.append(str(problem))
    return errors


def validate_manifest_data(
    actual: dict[str, Any],
    repo_root: Path,
    *,
    check_filesystem: bool = True,
    expected: dict[str, Any] | None = None,
) -> list[str]:
    errors = validate_top_level(actual)
    source = actual.get("source")
    source_head = source.get("repository_head") if isinstance(source, dict) else None
    if not is_git_sha(source_head):
        return errors
    errors.extend(
        validate_independent_authority(
            actual,
            repo_root,
            verify_live_sources=check_filesystem,
        )
    )
    errors.extend(validate_static_structure(actual, expected))
    registry = actual.get("consumer_registry")
    batches = actual.get("batches")
    if not isinstance(registry, list) or not isinstance(batches, list):
        return errors
    registry_by_id = {
        item["id"]: item
        for item in registry
        if isinstance(item, dict) and isinstance(item.get("id"), str)
    }
    safe = IndependentRepo(repo_root)
    states: list[str] = []
    errors.extend(global_state_order_errors(actual))
    for batch_index, batch in enumerate(batches):
        if not isinstance(batch, dict):
            continue
        states.append(str(batch.get("state")))
        errors.extend(
            validate_batch_state(
                actual,
                batch,
                registry_by_id,
                safe,
                batch_index=batch_index,
                check_filesystem=check_filesystem,
            )
        )
    expected_status = "planned" if all(state == "planned" for state in states) else (
        "verified" if states and all(state == "verified" for state in states) else "in-progress"
    )
    if actual.get("status") != expected_status:
        errors.append("top-level status does not match batch states")
    if check_filesystem and all(state == "planned" for state in states):
        try:
            live_head = read_git_head(safe)
            if live_head != source_head:
                errors.append("live repository HEAD differs from manifest source.repository_head")
        except ManifestError as problem:
            errors.append(f"repository head validation failed: {problem}")
    if check_filesystem:
        try:
            safe.finalize()
        except ManifestError as problem:
            errors.append(str(problem))
    return errors


def read_manifest(repo_root: Path, relative_path: str) -> dict[str, Any]:
    require_canonical_path(relative_path, "manifest check path")
    safe = IndependentRepo(repo_root)
    payload = safe.read(relative_path, max_bytes=32 * 1024 * 1024)
    safe.finalize()
    value = json_load(payload, "cutover manifest")
    if canonical_json_bytes(value, pretty=True) != payload:
        raise ManifestError("manifest bytes are not deterministic canonical JSON")
    return value


def read_manifest_stdin() -> dict[str, Any]:
    payload = sys.stdin.buffer.read(MAX_STDIN_BYTES + 1)
    if len(payload) > MAX_STDIN_BYTES:
        raise ManifestError("stdin manifest exceeds 32 MiB budget")
    value = json_load(payload, "stdin cutover manifest")
    if canonical_json_bytes(value, pretty=True) != payload:
        raise ManifestError("stdin manifest bytes are not deterministic canonical JSON")
    return value


def sentinel_detected(
    name: str,
    mutated: dict[str, Any],
    expected: dict[str, Any],
    repo_root: Path,
    fragment: str,
) -> str | None:
    errors = validate_manifest_data(
        mutated,
        repo_root,
        check_filesystem=False,
        expected=expected,
    )
    if not any(fragment in item for item in errors):
        return f"{name} sentinel was not rejected with {fragment!r}"
    return None


def two_batch_projection_sentinel() -> str | None:
    """Prove two verified batches can bind receipts without cross-batch cycles."""

    def verified_batch(batch_id: str, digit: str) -> dict[str, Any]:
        return {
            "id": batch_id,
            "state": "verified",
            "operations": [],
            "preconditions": {"operations_sha256": object_digest([])},
            "evidence": {"sentinel": batch_id},
            "commit": {
                "applied_commit": digit * 40,
                "verifier_commit": str((int(digit) + 1) % 10) * 40,
            },
            "state_transition": {
                "manifest_after_projection_sha256": None,
                "history": [{"after_projection_sha256": None}],
            },
            "verifier_record": {
                "identity": f"{batch_id}-verifier",
                "verdict": "PASS",
                "verified_at": "2026-08-08T00:00:00Z",
                "receipt_path": None,
                "receipt_sha256": None,
            },
            "consumer_transitions": [
                {"consumer_id": "consumer-manifest-self", "after_sha256": None}
            ],
        }

    manifest = {
        "schema_version": 1,
        "manifest_id": "projection-sentinel",
        "status": "in-progress",
        "source": {"repository_head": "1" * 40},
        "invariants": {"max_batch_operations": 64},
        "gate_registry": [],
        "consumer_registry": [],
        "totals": {"batches": 2},
        "batches": [verified_batch("batch-a", "1"), verified_batch("batch-b", "3")],
    }

    def bind_receipt(batch: dict[str, Any]) -> dict[str, Any]:
        commit = batch["commit"]
        verifier = batch["verifier_record"]
        body = {
            "schema_version": 1,
            "manifest_id": manifest["manifest_id"],
            "batch_id": batch["id"],
            "applied_commit": commit["applied_commit"],
            "verifier_commit": commit["verifier_commit"],
            "verifier_identity": verifier["identity"],
            "verified_at": verifier["verified_at"],
            "operations_sha256": batch["preconditions"]["operations_sha256"],
            "transition_chain_sha256": object_digest(batch["state_transition"]["history"]),
            "evidence_sha256": object_digest(batch["evidence"]),
            "commit_chain_sha256": object_digest(commit),
            "verifier_binding_sha256": object_digest({
                "identity": verifier["identity"],
                "verdict": verifier["verdict"],
                "verified_at": verifier["verified_at"],
            }),
        }
        receipt = {**body, "receipt_sha256": object_digest(body)}
        verifier["receipt_path"] = f"receipts/{batch['id']}.json"
        verifier["receipt_sha256"] = hashlib.sha256(
            canonical_json_bytes(receipt, pretty=True)
        ).hexdigest()
        return receipt

    digest_a = batch_projection_digest(manifest, "batch-a")
    batch_a = manifest["batches"][0]
    batch_a["state_transition"]["manifest_after_projection_sha256"] = digest_a
    batch_a["state_transition"]["history"][-1]["after_projection_sha256"] = digest_a
    batch_a["consumer_transitions"][0]["after_sha256"] = digest_a
    receipt_a = bind_receipt(batch_a)
    if receipt_a["transition_chain_sha256"] != object_digest(batch_a["state_transition"]["history"]):
        return "first verified receipt did not bind its projection-bearing transition chain"
    if batch_projection_digest(manifest, "batch-a") != digest_a:
        return "first verified batch projection is self-referential"

    digest_b = batch_projection_digest(manifest, "batch-b")
    batch_b = manifest["batches"][1]
    batch_b["state_transition"]["manifest_after_projection_sha256"] = digest_b
    batch_b["state_transition"]["history"][-1]["after_projection_sha256"] = digest_b
    batch_b["consumer_transitions"][0]["after_sha256"] = digest_b
    receipt_b = bind_receipt(batch_b)
    if receipt_b["transition_chain_sha256"] != object_digest(batch_b["state_transition"]["history"]):
        return "second verified receipt did not bind its projection-bearing transition chain"
    if batch_projection_digest(manifest, "batch-b") != digest_b:
        return "second verified batch projection is self-referential"
    if batch_projection_digest(manifest, "batch-a") != digest_a:
        return "second verified batch made first batch projection stale"

    batch_b["evidence"] = {"sentinel": "future-batch-b-evidence"}
    batch_b["state_transition"]["history"].append(
        {"after_projection_sha256": "5" * 64}
    )
    if batch_projection_digest(manifest, "batch-a") != digest_a:
        return "sibling future/dynamic fields entered batch-local projection"

    exact_predecessor = expected_global_predecessor_manifest(manifest, 1)
    if predecessor_snapshot_errors(
        manifest, exact_predecessor, 1, batch_a["commit"]["verifier_commit"]
    ):
        return "exact immediate global predecessor was rejected"
    stale_original = copy.deepcopy(exact_predecessor)
    for key, value in planned_batch_state(stale_original["batches"][0], {}).items():
        stale_original["batches"][0][key] = value
    stale_original["status"] = "planned"
    if not predecessor_snapshot_errors(
        manifest, stale_original, 1, manifest["source"]["repository_head"]
    ):
        return "stale original planned snapshot was accepted for the second batch"

    out_of_order = copy.deepcopy(manifest)
    for key, value in planned_batch_state(out_of_order["batches"][0], {}).items():
        out_of_order["batches"][0][key] = value
    if not global_state_order_errors(out_of_order):
        return "non-monotone global batch transition order was accepted"
    return None


def run_self_test(repo_root: Path, clean: dict[str, Any]) -> list[str]:
    failures: list[str] = []
    projection_failure = two_batch_projection_sentinel()
    if projection_failure:
        failures.append(projection_failure)
    clean_errors = validate_manifest_data(clean, repo_root, check_filesystem=True, expected=clean)
    if clean_errors:
        failures.append(f"unchanged derived manifest is not GREEN: {clean_errors[0]}")
        return failures

    sentinels: list[tuple[str, dict[str, Any], str]] = []
    traversal = copy.deepcopy(clean)
    traversal["batches"][0]["operations"][0]["source"] = "../escape"
    sentinels.append(("traversal", traversal, "independent planned manifest mismatch"))

    deep_path = copy.deepcopy(clean)
    deep_path["batches"][0]["operations"][0]["source"] = "/".join(
        ["a"] * (MAX_PATH_COMPONENTS + 1)
    )
    sentinels.append(("component-depth", deep_path, "exceeds 64 component budget"))

    duplicate = copy.deepcopy(clean)
    duplicate["batches"][0]["operations"][1] = copy.deepcopy(duplicate["batches"][0]["operations"][0])
    sentinels.append(("duplicate-operation", duplicate, "independent planned manifest mismatch"))

    omitted = copy.deepcopy(clean)
    omitted["batches"][0]["operations"].pop()
    sentinels.append(("omitted-operation", omitted, "independent planned manifest mismatch"))

    oversized = copy.deepcopy(clean)
    source_batch = max(oversized["batches"], key=lambda item: len(item["operations"]))
    while len(source_batch["operations"]) <= MAX_BATCH_OPERATIONS:
        source_batch["operations"].append(copy.deepcopy(source_batch["operations"][0]))
    sentinels.append(("over-64", oversized, "batch exceeds 64"))

    digest = copy.deepcopy(clean)
    digest["batches"][0]["preconditions"]["operations_sha256"] = "0" * 64
    sentinels.append(("digest-substitution", digest, "digest substitution"))

    missing_consumer = copy.deepcopy(clean)
    missing_consumer["consumer_registry"].pop(0)
    sentinels.append(("missing-consumer", missing_consumer, "missing/extra consumer"))

    invalid_transition = copy.deepcopy(clean)
    invalid_transition["status"] = "in-progress"
    invalid_transition["batches"][0]["state"] = "verified"
    sentinels.append(("invalid-transition", invalid_transition, "invalid transition count/skip"))

    protected = copy.deepcopy(clean)
    protected["batches"][0]["operations"][0]["source"] = "pro-qa-ai/forbidden"
    sentinels.append(("protected-path", protected, "enters protected pro-qa-ai"))

    stale = copy.deepcopy(clean)
    stale["batches"][0]["stale_reference_scan"]["exact_patterns"].pop()
    sentinels.append(("stale-reference", stale, "stale-reference declaration failure"))

    wrong_head = copy.deepcopy(clean)
    wrong_head["source"]["repository_head"] = "0" * len(clean["source"]["repository_head"])
    wrong_head_errors = validate_manifest_data(
        wrong_head,
        repo_root,
        check_filesystem=True,
        expected=clean,
    )
    if not any("live repository HEAD differs" in item for item in wrong_head_errors):
        failures.append("wrong-head sentinel did not reach live filesystem/HEAD validation")

    for name, mutated, fragment in sentinels:
        failure = sentinel_detected(name, mutated, clean, repo_root, fragment)
        if failure:
            failures.append(failure)

    with tempfile.TemporaryDirectory(prefix="hub-cutover-selftest-") as temporary:
        root = Path(temporary)
        (root / "real").mkdir()
        (root / "real" / "source.txt").write_bytes(b"safe")
        os.symlink(root / "real", root / "linked")
        try:
            IndependentRepo(root).read("linked/source.txt")
        except ManifestError as problem:
            if "symlink/reparse" not in str(problem):
                failures.append("symlink/reparse sentinel returned the wrong failure")
        else:
            failures.append("symlink/reparse sentinel was not rejected")

        destination_root = root / "destination-fixture"
        destination_root.mkdir()
        (destination_root / "target.txt").write_bytes(b"occupied")
        try:
            IndependentRepo(destination_root).assert_absent("target.txt")
        except ManifestError as problem:
            if "existing move destination" not in str(problem):
                failures.append("existing-destination sentinel returned the wrong failure")
        else:
            failures.append("existing-destination sentinel was not rejected")

        fake_reparse = type(
            "FakeStat",
            (),
            {"st_mode": stat.S_IFREG, "st_file_attributes": FILE_ATTRIBUTE_REPARSE_POINT},
        )()
        if not is_symlink_or_reparse(fake_reparse):
            failures.append("Windows reparse metadata sentinel was not recognized")
    return failures


def print_errors(errors: list[str]) -> None:
    for item in errors[:30]:
        print(f"FAIL: {item}", file=sys.stderr)
    if len(errors) > 30:
        print(f"FAIL: {len(errors) - 30} additional errors suppressed", file=sys.stderr)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo-root", type=Path, default=REPO_ROOT)
    sources = parser.add_mutually_exclusive_group(required=True)
    sources.add_argument("--check", metavar="PATH")
    sources.add_argument("--stdin", action="store_true")
    parser.add_argument("--self-test", action="store_true")
    args = parser.parse_args(argv)
    try:
        manifest = read_manifest_stdin() if args.stdin else read_manifest(args.repo_root, args.check)
        if args.self_test:
            failures = run_self_test(args.repo_root, manifest)
            if failures:
                print_errors(failures)
                return 1
            print(
                "PASS: independent input manifest GREEN; temp-only negative sentinels RED, "
                "including two-batch stale-predecessor/global-order rejection"
            )
            return 0
        errors = validate_manifest_data(manifest, args.repo_root)
        if errors:
            print_errors(errors)
            return 1
        states = [batch["state"] for batch in manifest["batches"]]
        print(
            "PASS: cutover manifest exact; "
            f"batches={len(states)} planned={states.count('planned')} "
            f"applied-unverified={states.count('applied-unverified')} "
            f"verified={states.count('verified')}"
        )
        return 0
    except (ManifestError, OSError, UnicodeError, KeyError, TypeError, ValueError) as problem:
        print(f"FAIL: {problem}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
