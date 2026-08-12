#!/usr/bin/env python3
"""Deterministic, stdlib-only inventory for the Hub test rebase.

The scanner is deliberately allow-listed.  It never walks the repository root;
every input is below one of the constants in SOURCE_ROOTS or TARGET_ROOTS.
"""

from __future__ import annotations

import argparse
import ctypes
import errno
import fnmatch
import hashlib
import json
import os
import re
import secrets
import selectors
import stat
import subprocess
import sys
import tempfile
import time
from copy import deepcopy
from dataclasses import dataclass
from pathlib import Path, PurePosixPath
from typing import Any, Callable, Iterable


TOOL_VERSION = "2.2.1"
_SEALED_REPO_ROOT = os.environ.get("HUB_PHASE128_REPO_ROOT")
REPO_ROOT = (
    Path(_SEALED_REPO_ROOT)
    if _SEALED_REPO_ROOT is not None
    else Path(__file__).resolve().parents[3]
)
if not REPO_ROOT.is_absolute():
    raise RuntimeError("HUB_PHASE128_REPO_ROOT must be absolute")
OUTPUT_PATH = REPO_ROOT / ".planning/audits/test-rebase/inventory.json"
SCHEMA_PATH = REPO_ROOT / ".planning/audits/test-rebase/inventory.schema.json"

# These limits are authority boundaries, not tuning hints.  The canonical
# snapshot is comfortably below them; an unexpectedly large input/tree must be
# reviewed instead of consuming unbounded host resources.
MAX_JSON_BYTES = 8 * 1024 * 1024
MAX_SOURCE_FILE_BYTES = 4 * 1024 * 1024
MAX_SOURCE_SNAPSHOT_BYTES = 512 * 1024 * 1024
MAX_WALK_DIRECTORIES = 4_096
MAX_WALK_FILES = 20_000
MAX_PATH_COMPONENTS = 64
MAX_COMPONENT_UTF8_BYTES = 255
MAX_PATH_UTF8_BYTES = 4096
MAX_INVENTORY_ROWS = 2_000
MAX_GIT_TREE_OUTPUT_BYTES = 128 * 1024 * 1024
MAX_GIT_STDERR_BYTES = 64 * 1024
GIT_COMMAND_TIMEOUT_SECONDS = 60

SOURCE_ROOTS = {
    "go": Path("backend/testframework"),
    "go_server": Path("backend/cmd/server"),
    "vitest": Path("testframework/stages/unit/frontend-test"),
    "playwright": Path("testframework/stages/e2e/domains"),
    "playwright_support": Path("testframework/stages/e2e"),
    "hooks": Path("backend/internal"),
}
TARGET_ROOTS = {
    "go": "backend/tests/",
    "vitest": "frontend/tests/",
    "api": "testframework/api/",
    "integration": "testframework/integration/",
    "contract": "testframework/contract/",
    "frontend_contract": "testframework/contract/frontend/",
    "playwright": "testframework/e2e/",
    "playwright_support": "testframework/e2e/support/",
    "python": "testframework/",
    "load": "testframework/load/",
}
PYTHON_COMPONENT_CLASSES = frozenset(
    {"python-api", "python-integration", "python-contract"}
)
COMPONENT_CLASS_TARGET_KEYS = {
    "python-api": "api",
    "python-integration": "integration",
    "python-contract": "contract",
}
PROTECTED_PREFIX = "pro-qa-ai/"
FILE_ATTRIBUTE_REPARSE_POINT = getattr(stat, "FILE_ATTRIBUTE_REPARSE_POINT", 0x400)

# These are the acceptance facts captured in evidence.md.  Source-tree drift is
# a deliberate RED result until the ledger and evidence are reviewed together.
EXPECTED = {
    "go_files": 832,
    "go_legacy_files": 831,
    "go_server_files": 1,
    "go_declared_tests": 3674,
    "go_legacy_declared_tests": 3672,
    "go_server_declared_tests": 2,
    "go_unit_files": 595,
    "go_component_files": 237,
    "vitest_files": 180,
    "frontend_unit_files": 156,
    "frontend_static_contract_files": 24,
    "frontend_support_files": 3,
    "playwright_files": 20,
    "playwright_cases": 53,
    "playwright_support_files": 14,
    "canonical_hook_files": 39,
    "target_python_tests": 0,
    "unmapped_files": 0,
}

GO_TEST_RE = re.compile(r"(?m)^func\s+((?:Test|Benchmark|Example)[A-Za-z0-9_]*)\s*\(")
TS_CASE_RE = re.compile(
    r"\b(?:it|test)(?:\.(?:only|skip|todo|fixme))?\s*\(\s*"
    r'(?:"((?:\\.|[^"\\])*)"|\'((?:\\.|[^\'\\])*)\'|`((?:\\.|[^`\\])*)`)'
)
TRACE_ID_RE = re.compile(r"\b(?:TC|FR|GZ|ECC)-[A-Za-z0-9._-]+\b")
GO_SKIP_RE = re.compile(r"\b(?:t|b|suite)\.(Skip|Skipf|SkipNow)\s*\(")
TS_SKIP_RE = re.compile(r"\b(?:it|test|describe)\.(skip|todo|skipIf|runIf)\s*\(")
HOOK_DECL_RE = re.compile(r"(?m)^(?:func|var|const|type)\s+(Export[A-Za-z0-9_]*)\b")
GO_IMPORT_RE = re.compile(
    r'(?m)^\s*(?:(?P<alias>[A-Za-z_.][A-Za-z0-9_.]*)\s+)?"(?P<path>[^"\r\n]+)"'
)
GO_SINGLE_IMPORT_RE = re.compile(
    r'(?m)^\s*import\s+(?:(?P<alias>[A-Za-z_.][A-Za-z0-9_.]*)\s+)?"(?P<path>[^"\r\n]+)"'
)
TS_IMPORT_RE = re.compile(
    r"(?:\bfrom\s+|\brequire\s*\(\s*|\bimport\s*)[\"']([^\"']+)[\"']"
)

SERVICE_PATTERNS = {
    "postgres": ("POSTGRES_TEST_URL", "postgresql://", 'enttest.Open(t, "postgres"'),
    "sqlite": ("enttest.Open", "sqlite3", "modernc.org/sqlite"),
    "mongodb": ("MONGODB_TEST_URL", "mongodb://", "mongo.Connect"),
    "redis": ("REDIS_TEST_URL", "redis.NewClient", "miniredis"),
    "nats": ("NATS_URL", "nats-server", "natsclient"),
    "object-storage": ("MINIO", "minio.New", "s3.New"),
    "http-server": ("httptest.NewServer", "net.Listen("),
    "external-process": ("exec.Command", "os/exec"),
    "browser": ("@playwright/test",),
    "dom": ("@testing-library/", "jsdom"),
}

# A Go file is component-level only when its content proves that it crosses a
# concrete process, network, or persistence boundary.  Directory names and
# historical filenames are intentionally absent from this decision.
GO_COMPONENT_MARKERS = (
    "enttest.Open",
    "POSTGRES_TEST_URL",
    "MONGODB_TEST_URL",
    "REDIS_TEST_URL",
    "testcontainers",
    "dockertest",
    "httptest.NewServer",
    "net.Listen(",
    "exec.Command",
    "nats-server",
)
GO_COMPONENT_SUPPORT_IMPORTS = (
    "/backend/testframework/env",
    "/backend/testframework/mongotest",
)
FRONTEND_STATIC_CONTRACT_MARKERS = (
    "node:fs",
    "node:path",
    "node:child_process",
)

GO_OWNER_WAVES = {
    "auth": "P133/TASK-210",
    "rbac": "P133/TASK-210",
    "audit": "P133/TASK-210",
    "ws": "P133/TASK-210",
    "calls": "P133/TASK-210",
    "platform": "P133/TASK-211",
    "apigateway": "P133/TASK-211",
    "dev": "P133/TASK-211",
    "ent": "P133/TASK-211",
    "env": "P133/TASK-211",
    "status": "P133/TASK-211",
    "tracker": "P134/TASK-212",
    "project": "P134/TASK-212",
    "sprint": "P134/TASK-212",
    "tms": "P134/TASK-212",
    "messenger": "P134/TASK-213",
    "notification": "P134/TASK-213",
    "accounting": "P135/TASK-214",
    "calendar": "P135/TASK-214",
    "intranet": "P135/TASK-214",
    "crm": "P135/TASK-214",
    "hr": "P135/TASK-214",
    "bases": "P135/TASK-214",
    "dashboard": "P135/TASK-214",
    "ai": "P135/TASK-215",
    "search": "P135/TASK-215",
    "storage": "P135/TASK-215",
    "backup": "P135/TASK-215",
    "imex": "P135/TASK-215",
    "github": "P135/TASK-215",
    "gitlab": "P135/TASK-215",
    "onec": "P135/TASK-215",
    "marketplace": "P135/TASK-215",
    "mail": "P135/TASK-215",
}
COMPONENT_OWNER_WAVES = {
    **{
        domain: "P133/TASK-227"
        for domain in (
            "auth",
            "rbac",
            "audit",
            "ws",
            "calls",
            "platform",
            "apigateway",
            "dev",
            "ent",
            "env",
            "status",
        )
    },
    **{
        domain: "P134/TASK-228"
        for domain in ("tracker", "project", "sprint", "tms", "messenger", "notification")
    },
    **{
        domain: "P135/TASK-229"
        for domain in (
            "accounting",
            "calendar",
            "intranet",
            "crm",
            "hr",
            "bases",
            "dashboard",
            "ai",
            "search",
            "storage",
            "backup",
            "imex",
            "github",
            "gitlab",
            "onec",
            "marketplace",
            "mail",
        )
    },
}
FRONTEND_OWNER_WAVES = {
    "auth": "P130/TASK-206",
    "platform": "P130/TASK-206",
    "tracker": "P131/TASK-207",
    "messenger": "P131/TASK-207",
    "tms": "P131/TASK-207",
    "calls": "P131/TASK-207",
    "admin": "P131/TASK-208",
    "ai": "P131/TASK-208",
    "dashboard": "P131/TASK-208",
    "hr": "P131/TASK-208",
    "notifications": "P131/TASK-208",
    "search": "P131/TASK-208",
    "v4": "P131/TASK-208",
    "support": "P130/TASK-206",
}
COMPLEX_E2E_SOURCES = {
    "testframework/stages/e2e/domains/ai/ai-hub.spec.ts",
    "testframework/stages/e2e/domains/messenger/messenger.spec.ts",
    "testframework/stages/e2e/domains/platform/a11y.spec.ts",
    "testframework/stages/e2e/domains/tms/tms.spec.ts",
    "testframework/stages/e2e/domains/tracker/project-lifecycle.spec.ts",
}
OWNER_WAVES = set(GO_OWNER_WAVES.values()) | set(FRONTEND_OWNER_WAVES.values()) | {
    *COMPONENT_OWNER_WAVES.values(),
    "P131/TASK-230",
    "P138/TASK-195",
    "P138/TASK-196",
    "P137/TASK-194",
}


def posix(path: Path) -> str:
    return path.as_posix()


def rel(path: Path) -> str:
    absolute = Path(os.path.abspath(path))
    if (
        len(absolute.parts) - 1 > MAX_PATH_COMPONENTS
        or len(str(absolute).encode("utf-8")) > MAX_PATH_UTF8_BYTES
        or any(len(part.encode("utf-8")) > MAX_COMPONENT_UTF8_BYTES for part in absolute.parts[1:])
    ):
        raise ValueError("repository root exceeds component/path budget")
    root = Path(os.path.abspath(REPO_ROOT))
    return posix(absolute.relative_to(root))


def require_relative_source_path(value: str) -> None:
    candidate = PurePosixPath(value)
    if (
        not value
        or "\\" in value
        or value.startswith("/")
        or candidate.as_posix() != value
        or any(part in {"", ".", ".."} for part in candidate.parts)
        or any(ord(character) < 32 or ord(character) == 127 for character in value)
        or len(candidate.parts) > MAX_PATH_COMPONENTS
        or len(value.encode("utf-8")) > MAX_PATH_UTF8_BYTES
        or any(len(part.encode("utf-8")) > MAX_COMPONENT_UTF8_BYTES for part in candidate.parts)
    ):
        raise ValueError(f"non-canonical repository source path: {value!r}")
    if value == PROTECTED_PREFIX.rstrip("/") or value.startswith(PROTECTED_PREFIX):
        raise ValueError(f"protected path rejected: {value}")


def sha256_bytes(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def is_symlink_or_reparse(metadata: os.stat_result) -> bool:
    attributes = int(getattr(metadata, "st_file_attributes", 0))
    return stat.S_ISLNK(metadata.st_mode) or bool(
        attributes & FILE_ATTRIBUTE_REPARSE_POINT
    )


def metadata_identity(metadata: os.stat_result, *, include_content: bool) -> tuple[int, ...]:
    identity = (
        int(metadata.st_mode),
        int(metadata.st_dev),
        int(metadata.st_ino),
        int(metadata.st_nlink),
        int(getattr(metadata, "st_file_attributes", 0)),
    )
    if include_content:
        identity += (
            int(metadata.st_size),
            int(getattr(metadata, "st_mtime_ns", int(metadata.st_mtime * 1_000_000_000))),
            int(getattr(metadata, "st_ctime_ns", int(metadata.st_ctime * 1_000_000_000))),
        )
    return identity


@dataclass(frozen=True)
class SourceSnapshotEntry:
    """One immutable blob from the pinned Git tree used for classification."""

    allowed_root: Path
    blob_oid: str
    mode: str
    data: bytes
    sha256: str


def require_descriptor_traversal() -> None:
    """Fail closed unless openat-style no-follow traversal is available."""

    required_flags = ("O_CLOEXEC", "O_DIRECTORY", "O_NOFOLLOW")
    missing = [name for name in required_flags if not hasattr(os, name)]
    supports_dir_fd = os.open in getattr(os, "supports_dir_fd", set())
    if os.name != "posix" or missing or not supports_dir_fd:
        details = f"os={os.name} missing={missing} dir_fd={supports_dir_fd}"
        raise ValueError(
            "descriptor-relative O_NOFOLLOW traversal is unavailable; "
            f"filesystem authority fails closed ({details})"
        )


def open_physical_directory(path: Path) -> tuple[int, os.stat_result]:
    """Open every absolute directory component no-follow; reject protected root aliases."""

    require_descriptor_traversal()
    absolute = Path(os.path.abspath(path))
    if (
        len(absolute.parts) - 1 > MAX_PATH_COMPONENTS
        or len(str(absolute).encode("utf-8")) > MAX_PATH_UTF8_BYTES
        or any(len(part.encode("utf-8")) > MAX_COMPONENT_UTF8_BYTES for part in absolute.parts[1:])
    ):
        raise ValueError("repository root exceeds component/path budget")
    protected = os.path.normcase(os.path.abspath(REPO_ROOT / PROTECTED_PREFIX.rstrip("/")))
    candidate = os.path.normcase(str(absolute))
    try:
        inside_protected = os.path.commonpath([candidate, protected]) == protected
    except ValueError:
        inside_protected = False
    if inside_protected:
        raise ValueError(f"repository root enters protected pro-qa-ai: {path}")
    flags = os.O_RDONLY | os.O_CLOEXEC | os.O_DIRECTORY | os.O_NOFOLLOW
    anchor = Path(absolute.anchor)
    descriptor = os.open(anchor, flags)
    current = anchor
    try:
        for part in absolute.parts[1:]:
            current /= part
            before = os.stat(part, dir_fd=descriptor, follow_symlinks=False)
            if is_symlink_or_reparse(before) or not stat.S_ISDIR(before.st_mode):
                raise ValueError(f"physical repository ancestor required: {current}")
            next_fd = os.open(part, flags, dir_fd=descriptor)
            opened = os.fstat(next_fd)
            if metadata_identity(before, include_content=False) != metadata_identity(
                opened, include_content=False
            ):
                os.close(next_fd)
                raise ValueError(f"repository ancestor changed while opening: {current}")
            os.close(descriptor)
            descriptor = next_fd
        return descriptor, os.fstat(descriptor)
    except Exception:
        os.close(descriptor)
        raise


def open_existing_path(
    path: Path,
    allowed_root: Path,
    *,
    repo_root: Path = REPO_ROOT,
    require_directory: bool = False,
) -> tuple[int, os.stat_result, list[tuple[Path, tuple[int, ...]]]]:
    """Open a leaf through held directory descriptors without following links."""

    require_descriptor_traversal()
    repository = Path(os.path.abspath(repo_root))
    allowed = Path(os.path.abspath(repository / allowed_root))
    candidate = Path(os.path.abspath(path))
    try:
        allowed.relative_to(repository)
        relative = candidate.relative_to(repository)
        candidate.relative_to(allowed)
    except ValueError as error:
        raise ValueError(
            f"path escaped allow-listed root {allowed_root}: {path}"
        ) from error
    if len(relative.parts) > MAX_PATH_COMPONENTS or len(relative.as_posix().encode("utf-8")) > MAX_PATH_UTF8_BYTES:
        raise ValueError(f"path component-depth/byte budget exceeded: {path}")
    if any(len(part.encode("utf-8")) > MAX_COMPONENT_UTF8_BYTES for part in relative.parts):
        raise ValueError(f"path component byte budget exceeded: {path}")

    directory_flags = os.O_RDONLY | os.O_CLOEXEC | os.O_DIRECTORY | os.O_NOFOLLOW
    file_flags = os.O_RDONLY | os.O_CLOEXEC | os.O_NOFOLLOW | getattr(os, "O_NONBLOCK", 0)
    descriptors: list[int] = []
    snapshots: list[tuple[Path, tuple[int, ...]]] = []
    current_path = repository
    try:
        root_descriptor, root_opened = open_physical_directory(repository)
        descriptors.append(root_descriptor)
        snapshots.append(
            (repository, metadata_identity(root_opened, include_content=False))
        )

        for index, part in enumerate(relative.parts):
            is_leaf = index == len(relative.parts) - 1
            current_path = current_path / part
            try:
                before = os.stat(part, dir_fd=descriptors[-1], follow_symlinks=False)
            except OSError as error:
                raise ValueError(f"metadata unavailable for {current_path}: {error}") from error
            if is_symlink_or_reparse(before):
                raise ValueError(f"symlink/reparse component rejected: {current_path}")
            flags = directory_flags if not is_leaf or require_directory else file_flags
            try:
                descriptor = os.open(part, flags, dir_fd=descriptors[-1])
            except OSError as error:
                raise ValueError(
                    f"descriptor-relative open rejected {current_path}: {error}"
                ) from error
            descriptors.append(descriptor)
            opened = os.fstat(descriptor)
            include_content = is_leaf and not require_directory
            if metadata_identity(before, include_content=include_content) != metadata_identity(
                opened, include_content=include_content
            ):
                raise ValueError(f"metadata changed while opening {current_path}")
            if not is_leaf and not stat.S_ISDIR(opened.st_mode):
                raise ValueError(f"non-directory source ancestor rejected: {current_path}")
            if is_leaf:
                expected = stat.S_ISDIR if require_directory else stat.S_ISREG
                if not expected(opened.st_mode):
                    kind = "directory" if require_directory else "regular file"
                    raise ValueError(f"source leaf is not a {kind}: {current_path}")
                if not require_directory and int(opened.st_nlink) != 1:
                    raise ValueError(
                        f"source regular file link count must be one: {current_path} "
                        f"(st_nlink={opened.st_nlink})"
                    )
            snapshots.append(
                (current_path, metadata_identity(opened, include_content=include_content))
            )

        leaf_descriptor = descriptors.pop()
        leaf_metadata = os.fstat(leaf_descriptor)
        return leaf_descriptor, leaf_metadata, snapshots
    except Exception:
        for descriptor in reversed(descriptors):
            try:
                os.close(descriptor)
            except OSError:
                pass
        raise
    finally:
        for descriptor in reversed(descriptors):
            try:
                os.close(descriptor)
            except OSError:
                pass


def inspect_existing_path(
    path: Path,
    allowed_root: Path,
    *,
    repo_root: Path = REPO_ROOT,
    require_directory: bool = False,
) -> os.stat_result:
    """Reject escape/link races with two independent descriptor traversals."""

    descriptor, before, snapshots_before = open_existing_path(
        path,
        allowed_root,
        repo_root=repo_root,
        require_directory=require_directory,
    )
    os.close(descriptor)
    descriptor, after, snapshots_after = open_existing_path(
        path,
        allowed_root,
        repo_root=repo_root,
        require_directory=require_directory,
    )
    os.close(descriptor)
    if snapshots_before != snapshots_after:
        raise ValueError(f"metadata changed during descriptor recheck (TOCTOU): {path}")
    return after


def read_bounded_allowed_with_identity(
    path: Path,
    root: Path,
    *,
    max_bytes: int,
    repo_root: Path = REPO_ROOT,
) -> tuple[bytes, tuple[int, ...]]:
    source_path = Path(os.path.abspath(path)).relative_to(
        Path(os.path.abspath(repo_root))
    ).as_posix()
    if source_path.startswith(PROTECTED_PREFIX):
        raise ValueError(f"protected path rejected: {source_path}")
    descriptor, before, snapshots_before = open_existing_path(
        path, root, repo_root=repo_root
    )
    try:
        if int(before.st_size) > max_bytes:
            raise ValueError(
                f"input file exceeds {max_bytes} byte budget: {path}"
            )
        with os.fdopen(descriptor, "rb", closefd=False) as stream:
            data = stream.read(max_bytes + 1)
        if len(data) > max_bytes:
            raise ValueError(
                f"input file exceeded {max_bytes} byte budget while reading: {path}"
            )
        during = os.fstat(descriptor)
        if metadata_identity(before, include_content=True) != metadata_identity(
            during, include_content=True
        ):
            raise ValueError(f"source changed during descriptor read (TOCTOU): {path}")
    finally:
        os.close(descriptor)
    descriptor, after, snapshots_after = open_existing_path(
        path, root, repo_root=repo_root
    )
    os.close(descriptor)
    if metadata_identity(before, include_content=True) != metadata_identity(
        after, include_content=True
    ):
        raise ValueError(f"source changed during read (TOCTOU): {path}")
    if snapshots_before != snapshots_after:
        raise ValueError(f"source ancestor changed during read (TOCTOU): {path}")
    return data, metadata_identity(after, include_content=True)


def read_bounded_allowed(
    path: Path,
    root: Path,
    *,
    max_bytes: int,
    repo_root: Path = REPO_ROOT,
) -> bytes:
    data, _ = read_bounded_allowed_with_identity(
        path,
        root,
        max_bytes=max_bytes,
        repo_root=repo_root,
    )
    return data


def read_allowed(
    path: Path,
    root: Path,
    *,
    snapshot: dict[str, SourceSnapshotEntry] | None = None,
) -> tuple[bytes, str]:
    if snapshot is None:
        data = read_bounded_allowed(path, root, max_bytes=MAX_SOURCE_FILE_BYTES)
    else:
        source_path = rel(path)
        entry = snapshot.get(source_path)
        if entry is None or entry.allowed_root != root:
            raise ValueError(f"source is absent from pinned Git snapshot: {source_path}")
        data = entry.data
    return data, data.decode("utf-8", errors="replace")


def sorted_files(
    root: Path,
    pattern: str,
    *,
    repo_root: Path = REPO_ROOT,
    max_directories: int = MAX_WALK_DIRECTORIES,
    max_files: int = MAX_WALK_FILES,
) -> list[Path]:
    absolute = repo_root / root
    excluded = {"node_modules", ".git", "coverage", "dist"}

    def walk_once() -> tuple[list[Path], list[tuple[str, tuple[int, ...]]]]:
        root_fd, root_meta, _ = open_existing_path(
            absolute, root, repo_root=repo_root, require_directory=True
        )
        files: list[Path] = []
        snapshots: list[tuple[str, tuple[int, ...]]] = [
            (Path(root).as_posix(), metadata_identity(root_meta, include_content=False))
        ]
        visited_directories = 0
        visited_files = 0
        directory_flags = os.O_RDONLY | os.O_CLOEXEC | os.O_DIRECTORY | os.O_NOFOLLOW

        def visit(directory_fd: int, relative: Path, depth: int) -> None:
            nonlocal visited_directories, visited_files
            visited_directories += 1
            if visited_directories > max_directories:
                raise ValueError(f"source traversal exceeded {max_directories} directory budget: {root}")
            if depth > MAX_PATH_COMPONENTS:
                raise ValueError(f"source traversal exceeded {MAX_PATH_COMPONENTS} component budget: {relative}")
            try:
                names = sorted(os.listdir(directory_fd))
            except OSError as error:
                raise ValueError(f"descriptor-relative source enumeration failed: {relative}: {error}") from error
            visited_files += len(names)
            if visited_files > max_files:
                raise ValueError(f"source traversal exceeded {max_files} entry budget: {root}")
            for name in names:
                if (
                    name in {"", ".", ".."}
                    or "/" in name
                    or "\\" in name
                    or len(name.encode("utf-8")) > MAX_COMPONENT_UTF8_BYTES
                ):
                    raise ValueError(f"source traversal component rejected: {relative / name}")
                child_relative = relative / name
                if depth + 1 > MAX_PATH_COMPONENTS:
                    raise ValueError(
                        f"source traversal exceeded {MAX_PATH_COMPONENTS} component budget: {child_relative}"
                    )
                if len(child_relative.as_posix().encode("utf-8")) > MAX_PATH_UTF8_BYTES:
                    raise ValueError(f"source traversal path byte budget exceeded: {child_relative}")
                try:
                    before = os.stat(name, dir_fd=directory_fd, follow_symlinks=False)
                except OSError as error:
                    raise ValueError(f"source traversal metadata inaccessible: {child_relative}: {error}") from error
                if is_symlink_or_reparse(before):
                    raise ValueError(f"source traversal link/reparse rejected: {child_relative}")
                if stat.S_ISDIR(before.st_mode):
                    snapshots.append((child_relative.as_posix(), metadata_identity(before, include_content=False)))
                    if name in excluded:
                        continue
                    child_fd = os.open(name, directory_flags, dir_fd=directory_fd)
                    try:
                        opened = os.fstat(child_fd)
                        if metadata_identity(before, include_content=False) != metadata_identity(opened, include_content=False):
                            raise ValueError(f"source directory changed while opening: {child_relative}")
                        visit(child_fd, child_relative, depth + 1)
                        if metadata_identity(opened, include_content=False) != metadata_identity(os.fstat(child_fd), include_content=False):
                            raise ValueError(f"source directory identity changed during enumeration: {child_relative}")
                    finally:
                        os.close(child_fd)
                elif stat.S_ISREG(before.st_mode):
                    if int(before.st_nlink) != 1:
                        raise ValueError(f"source enumeration regular file link count must be one: {child_relative}")
                    snapshots.append((child_relative.as_posix(), metadata_identity(before, include_content=True)))
                    if fnmatch.fnmatch(name, pattern):
                        files.append(Path(os.path.abspath(repo_root)) / child_relative)
                else:
                    raise ValueError(f"source traversal special file rejected: {child_relative}")
        try:
            visit(root_fd, Path(root), len(Path(root).parts))
        finally:
            os.close(root_fd)
        return files, snapshots

    first_files, first_snapshot = walk_once()
    second_files, second_snapshot = walk_once()
    first_rel = [path.relative_to(Path(os.path.abspath(repo_root))).as_posix() for path in first_files]
    second_rel = [path.relative_to(Path(os.path.abspath(repo_root))).as_posix() for path in second_files]
    if first_snapshot != second_snapshot or first_rel != second_rel:
        raise ValueError(f"source enumeration changed during descriptor recheck: {root}")
    return first_files


def git_read_environment() -> dict[str, str]:
    environment = os.environ.copy()
    environment.update(
        {
            "GIT_NO_REPLACE_OBJECTS": "1",
            "GIT_OPTIONAL_LOCKS": "0",
            "GIT_PAGER": "cat",
            "GIT_CONFIG_GLOBAL": os.devnull,
            "GIT_CONFIG_NOSYSTEM": "1",
        }
    )
    return environment


def run_git_bounded(
    arguments: list[str],
    *,
    max_stdout: int,
    timeout_seconds: int = GIT_COMMAND_TIMEOUT_SECONDS,
) -> bytes:
    """Run one fixed Git read with bounded stdout/stderr and wall time."""

    process = subprocess.Popen(
        ["git", *arguments],
        cwd=REPO_ROOT,
        env=git_read_environment(),
        stdin=subprocess.DEVNULL,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        bufsize=0,
    )
    if process.stdout is None or process.stderr is None:
        raise ValueError("bounded Git reader did not expose output pipes")
    selector = selectors.DefaultSelector()
    selector.register(process.stdout, selectors.EVENT_READ, "stdout")
    selector.register(process.stderr, selectors.EVENT_READ, "stderr")
    output = {"stdout": bytearray(), "stderr": bytearray()}
    limits = {"stdout": max_stdout, "stderr": MAX_GIT_STDERR_BYTES}
    deadline = time.monotonic() + timeout_seconds
    try:
        while selector.get_map():
            remaining = deadline - time.monotonic()
            if remaining <= 0:
                raise ValueError(f"Git read exceeded {timeout_seconds}s wall timeout")
            events = selector.select(remaining)
            if not events:
                raise ValueError(f"Git read exceeded {timeout_seconds}s wall timeout")
            for key, _ in events:
                chunk = os.read(key.fileobj.fileno(), 64 * 1024)
                if not chunk:
                    selector.unregister(key.fileobj)
                    continue
                bucket = output[str(key.data)]
                bucket.extend(chunk)
                if len(bucket) > limits[str(key.data)]:
                    raise ValueError(
                        f"Git {key.data} exceeded {limits[str(key.data)]} byte budget"
                    )
        remaining = max(0.1, deadline - time.monotonic())
        return_code = process.wait(timeout=remaining)
        if return_code != 0:
            detail = bytes(output["stderr"]).decode("utf-8", errors="replace").strip()
            raise ValueError(f"Git read failed ({return_code}): {detail[:512]}")
        return bytes(output["stdout"])
    except Exception:
        if process.poll() is None:
            process.kill()
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired as error:
                raise ValueError("Git reader could not be terminated within 5s") from error
        raise
    finally:
        selector.close()
        process.stdout.close()
        process.stderr.close()


def git_tree_blobs(head: str) -> dict[str, tuple[str, str, int]]:
    """Return bounded regular-blob metadata from the exact pinned commit tree."""

    if re.fullmatch(r"[0-9a-f]{40}", head) is None:
        raise ValueError(f"invalid pinned Git HEAD: {head!r}")
    roots = sorted(
        {
            SOURCE_ROOTS["go"].as_posix(),
            SOURCE_ROOTS["go_server"].as_posix(),
            SOURCE_ROOTS["hooks"].as_posix(),
            TARGET_ROOTS["python"].rstrip("/"),
        }
    )
    payload = run_git_bounded(
        ["ls-tree", "-r", "-l", "-z", "--full-tree", head, "--", *roots],
        max_stdout=MAX_GIT_TREE_OUTPUT_BYTES,
    )
    result: dict[str, tuple[str, str, int]] = {}
    for record in payload.split(b"\0"):
        if not record:
            continue
        try:
            metadata, raw_path = record.split(b"\t", 1)
            fields = metadata.split()
            if len(fields) != 4:
                raise ValueError("unexpected ls-tree metadata field count")
            mode, kind, raw_oid, raw_size = fields
            path = raw_path.decode("utf-8", errors="strict")
            oid = raw_oid.decode("ascii", errors="strict")
            size_text = raw_size.decode("ascii", errors="strict")
        except (ValueError, UnicodeError) as error:
            raise ValueError("malformed bounded Git tree record") from error
        if kind != b"blob":
            continue
        if mode not in {b"100644", b"100755", b"120000"}:
            raise ValueError(f"unsupported Git blob mode for {path}: {mode!r}")
        if re.fullmatch(r"[0-9a-f]{40}", oid) is None or not size_text.isascii() or not size_text.isdigit():
            raise ValueError(f"invalid Git blob metadata for {path}")
        require_relative_source_path(path)
        if path in result:
            raise ValueError(f"duplicate path in pinned Git tree: {path}")
        if len(result) >= MAX_WALK_FILES:
            raise ValueError(f"pinned Git tree exceeded {MAX_WALK_FILES} entry budget")
        result[path] = (oid, mode.decode("ascii"), int(size_text))
    return result


def read_git_blobs(
    requested: list[tuple[str, int]],
    *,
    timeout_seconds: int = GIT_COMMAND_TIMEOUT_SECONDS,
) -> dict[str, bytes]:
    """Read exact blob IDs through one bounded git-cat-file batch protocol."""

    process = subprocess.Popen(
        ["git", "cat-file", "--batch"],
        cwd=REPO_ROOT,
        env=git_read_environment(),
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.DEVNULL,
        bufsize=0,
    )
    if process.stdin is None or process.stdout is None:
        raise ValueError("Git batch reader did not expose protocol pipes")
    selector = selectors.DefaultSelector()
    selector.register(process.stdout, selectors.EVENT_READ)
    buffer = bytearray()
    deadline = time.monotonic() + timeout_seconds

    def fill() -> None:
        remaining = deadline - time.monotonic()
        if remaining <= 0 or not selector.select(remaining):
            raise ValueError(f"Git blob read exceeded {timeout_seconds}s wall timeout")
        chunk = os.read(process.stdout.fileno(), 64 * 1024)
        if not chunk:
            raise ValueError("Git blob reader ended before the declared payload")
        buffer.extend(chunk)
        if len(buffer) > MAX_SOURCE_FILE_BYTES + 1024:
            raise ValueError("Git blob protocol buffer exceeded bounded file budget")

    def take_line() -> bytes:
        while b"\n" not in buffer:
            if len(buffer) > 512:
                raise ValueError("Git blob protocol header exceeded 512 bytes")
            fill()
        index = buffer.index(b"\n")
        line = bytes(buffer[:index])
        del buffer[: index + 1]
        return line

    def take_exact(length: int) -> bytes:
        while len(buffer) < length:
            fill()
        value = bytes(buffer[:length])
        del buffer[:length]
        return value

    result: dict[str, bytes] = {}
    try:
        for oid, expected_size in requested:
            process.stdin.write(f"{oid}\n".encode("ascii"))
            process.stdin.flush()
            header = take_line().split()
            if len(header) != 3 or header[0] != oid.encode("ascii") or header[1] != b"blob":
                raise ValueError(f"Git blob protocol rejected object {oid}")
            try:
                actual_size = int(header[2])
            except ValueError as error:
                raise ValueError(f"Git blob protocol size invalid for {oid}") from error
            if actual_size != expected_size or actual_size > MAX_SOURCE_FILE_BYTES:
                raise ValueError(f"Git blob size changed/exceeded budget for {oid}")
            data = take_exact(actual_size)
            if take_exact(1) != b"\n":
                raise ValueError(f"Git blob protocol terminator invalid for {oid}")
            authenticated_oid = hashlib.sha1(
                f"blob {actual_size}\0".encode("ascii") + data,
                usedforsecurity=False,
            ).hexdigest()
            if authenticated_oid != oid:
                raise ValueError(f"Git blob payload failed object-ID authentication: {oid}")
            result[oid] = data
        if buffer:
            raise ValueError("Git blob reader emitted undeclared protocol bytes")
        process.stdin.close()
        remaining = max(0.1, deadline - time.monotonic())
        return_code = process.wait(timeout=remaining)
        if return_code != 0:
            raise ValueError(f"Git blob reader failed ({return_code})")
        return result
    except Exception:
        if process.poll() is None:
            process.kill()
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired as error:
                raise ValueError("Git blob reader could not be terminated within 5s") from error
        raise
    finally:
        selector.close()
        if not process.stdin.closed:
            process.stdin.close()
        process.stdout.close()


def revision() -> str:
    value = run_git_bounded(
        ["rev-parse", "--verify", "HEAD^{commit}"], max_stdout=4096, timeout_seconds=10
    ).decode("ascii", errors="strict").strip()
    if not re.fullmatch(r"[0-9a-f]{40}", value):
        raise ValueError(f"unexpected git revision: {value!r}")
    return value


def revision_is_ancestor(value: str, descendant: str) -> bool:
    if not re.fullmatch(r"[0-9a-f]{40}", value) or not re.fullmatch(
        r"[0-9a-f]{40}", descendant
    ):
        return False
    result = subprocess.run(
        ["git", "merge-base", "--is-ancestor", value, descendant],
        cwd=REPO_ROOT,
        env=git_read_environment(),
        check=False,
        capture_output=True,
        text=True,
        timeout=10,
    )
    return result.returncode == 0


def line_sites(text: str, pattern: re.Pattern[str]) -> list[str]:
    return [f"L{text.count(chr(10), 0, match.start()) + 1}:{match.group(1)}" for match in pattern.finditer(text)]


def ts_cases(text: str) -> list[str]:
    cases = []
    for match in TS_CASE_RE.finditer(text):
        title = next(group for group in match.groups() if group is not None)
        cases.append(title.replace("\\`", "`").replace("\\\"", '"').replace("\\'", "'"))
    return cases


def go_import_pairs(text: str) -> list[tuple[str | None, str]]:
    pairs = set()
    for pattern in (GO_IMPORT_RE, GO_SINGLE_IMPORT_RE):
        for match in pattern.finditer(text):
            alias = match.group("alias")
            # GO_IMPORT_RE sees the keyword in `import "path"` as an alias;
            # GO_SINGLE_IMPORT_RE supplies the canonical pair for that form.
            if alias == "import":
                continue
            pairs.add((alias, match.group("path")))
    return sorted(pairs, key=lambda item: ((item[0] or ""), item[1]))


def external_go_imports(text: str) -> list[str]:
    imports = []
    for _, value in go_import_pairs(text):
        if "." in value.split("/", 1)[0]:
            imports.append(value)
    return sorted(set(imports))


def external_ts_imports(text: str) -> list[str]:
    imports = []
    for value in TS_IMPORT_RE.findall(text):
        if value.startswith((".", "/", "@/")):
            continue
        if value.startswith("@"):
            parts = value.split("/")
            imports.append("/".join(parts[:2]))
        else:
            imports.append(value.split("/", 1)[0])
    return sorted(set(imports))


def service_hints(text: str) -> list[str]:
    return sorted(
        service
        for service, markers in SERVICE_PATTERNS.items()
        if any(marker in text for marker in markers)
    )


def domain_for(relative: Path, runner: str) -> str:
    parts = relative.parts
    if runner in {"vitest", "vitest-support"}:
        if parts and parts[0] == "domains":
            return parts[1] if len(parts) > 1 else "unknown"
        return "support"
    if runner == "playwright" and parts:
        return parts[1] if parts[0] == "domains" and len(parts) > 1 else parts[0]
    if runner == "go" and parts:
        return parts[0]
    return "unknown"


def go_target(relative: Path) -> str:
    return posix(Path(TARGET_ROOTS["go"]) / relative)


def is_go_component(text: str, dependencies: list[str]) -> bool:
    return any(marker in text for marker in GO_COMPONENT_MARKERS) or any(
        marker in dependency
        for dependency in dependencies
        for marker in GO_COMPONENT_SUPPORT_IMPORTS
    )


def go_component_target_class(services: list[str], dependencies: list[str]) -> str:
    if "http-server" in services:
        return "python-api"
    integration_services = {
        "postgres",
        "sqlite",
        "mongodb",
        "redis",
        "nats",
        "object-storage",
        "external-process",
    }
    integration_dependencies = ("testcontainers", "dockertest") + GO_COMPONENT_SUPPORT_IMPORTS
    if integration_services.intersection(services) or any(
        marker in dependency
        for dependency in dependencies
        for marker in integration_dependencies
    ):
        return "python-integration"
    # A component marker without a resolved runtime dependency still needs an
    # explicit Python parity destination; it is not silently treated as Go unit.
    return "python-contract"


def go_component_target(relative: Path, target_class: str) -> str:
    target_key = COMPONENT_CLASS_TARGET_KEYS.get(target_class)
    if target_key is None:
        raise ValueError(f"unsupported Go component target class: {target_class}")
    parts = list(relative.parts)
    filename = parts[-1]
    stem = filename[: -len("_test.go")] if filename.endswith("_test.go") else Path(filename).stem
    parts[-1] = f"test_{stem.replace('-', '_')}.py"
    return posix(Path(TARGET_ROOTS[target_key]).joinpath(*parts))


def vitest_target(relative: Path) -> str:
    parts = list(relative.parts)
    if parts and parts[0] == "domains":
        parts = parts[1:]
    return posix(Path(TARGET_ROOTS["vitest"]).joinpath(*parts))


def frontend_support_target(relative: Path) -> str:
    parts = list(relative.parts)
    if parts and parts[0] == "domains":
        parts = parts[1:]
    return posix(Path(TARGET_ROOTS["vitest"]).joinpath("support", *parts))


def python_test_filename(filename: str) -> str:
    suffixes = (".test.tsx", ".spec.tsx", ".test.ts", ".spec.ts", "_test.go")
    stem = next(
        (filename[: -len(suffix)] for suffix in suffixes if filename.endswith(suffix)),
        Path(filename).stem,
    )
    normalized = re.sub(r"[^A-Za-z0-9_]+", "_", stem).strip("_")
    return f"test_{normalized}.py"


def frontend_static_contract_target(relative: Path) -> str:
    parts = list(relative.parts)
    if parts and parts[0] == "domains":
        parts = parts[1:]
    parts[-1] = python_test_filename(parts[-1])
    return posix(Path(TARGET_ROOTS["frontend_contract"]).joinpath(*parts))


def e2e_target(relative: Path) -> str:
    parts = list(relative.parts)
    if parts and parts[0] == "domains":
        parts = parts[1:]
    parts[-1] = python_test_filename(parts[-1])
    return posix(Path(TARGET_ROOTS["playwright"]).joinpath(*parts))


def e2e_support_target(relative: Path) -> str:
    parts = list(relative.parts)
    filename = parts[-1]
    stem = filename[: -len(".tsx")] if filename.endswith(".tsx") else filename[: -len(".ts")]
    parts[-1] = f"{re.sub(r'[^A-Za-z0-9_]+', '_', stem).strip('_')}.py"
    return posix(Path(TARGET_ROOTS["playwright_support"]).joinpath(*parts))


def replacement_status(target_path: str) -> str:
    if (
        target_path.startswith(("/", "//"))
        or "\\" in target_path
        or target_path == "pro-qa-ai"
        or target_path.startswith(PROTECTED_PREFIX)
    ):
        raise ValueError(f"replacement target is outside the repository boundary: {target_path}")
    parts = target_path.split("/")
    if (
        any(part in {"", ".", ".."} for part in parts)
        or len(parts) > MAX_PATH_COMPONENTS
        or len(target_path.encode("utf-8")) > MAX_PATH_UTF8_BYTES
        or any(len(part.encode("utf-8")) > MAX_COMPONENT_UTF8_BYTES for part in parts)
    ):
        raise ValueError(f"replacement target exceeds canonical path budget: {target_path}")

    def inspect() -> tuple[bool, tuple[tuple[int, ...], ...]]:
        root_fd, root_meta = open_physical_directory(REPO_ROOT)
        descriptors = [root_fd]
        snapshots = [metadata_identity(root_meta, include_content=False)]
        directory_flags = os.O_RDONLY | os.O_CLOEXEC | os.O_DIRECTORY | os.O_NOFOLLOW
        file_flags = os.O_RDONLY | os.O_CLOEXEC | os.O_NOFOLLOW | getattr(os, "O_NONBLOCK", 0)
        try:
            for index, part in enumerate(parts):
                leaf = index == len(parts) - 1
                try:
                    before = os.stat(part, dir_fd=descriptors[-1], follow_symlinks=False)
                except FileNotFoundError:
                    return False, tuple(snapshots)
                if is_symlink_or_reparse(before):
                    raise ValueError(f"replacement target link/reparse rejected: {target_path}")
                descriptor = os.open(part, file_flags if leaf else directory_flags, dir_fd=descriptors[-1])
                descriptors.append(descriptor)
                opened = os.fstat(descriptor)
                if metadata_identity(before, include_content=leaf) != metadata_identity(opened, include_content=leaf):
                    raise ValueError(f"replacement target changed while opening: {target_path}")
                if leaf:
                    if not stat.S_ISREG(opened.st_mode) or int(opened.st_nlink) != 1:
                        raise ValueError(f"replacement target must be a single-link regular file: {target_path}")
                elif not stat.S_ISDIR(opened.st_mode):
                    raise ValueError(f"replacement target ancestor is not a directory: {target_path}")
                snapshots.append(metadata_identity(opened, include_content=leaf))
            return True, tuple(snapshots)
        finally:
            for descriptor in reversed(descriptors):
                os.close(descriptor)

    before = inspect()
    after = inspect()
    if before != after:
        raise ValueError(f"replacement target changed during descriptor recheck: {target_path}")
    return "candidate-present" if before[0] else "not-started"


def canonical_owner_wave(
    source_path: str, runner: str, domain: str, target_class: str
) -> str:
    if runner == "go-test":
        wave = (
            COMPONENT_OWNER_WAVES.get(domain)
            if target_class in PYTHON_COMPONENT_CLASSES
            else GO_OWNER_WAVES.get(domain)
        )
    elif runner in {"vitest", "vitest-support"}:
        wave = (
            "P131/TASK-230"
            if target_class == "python-static-contract"
            else FRONTEND_OWNER_WAVES.get(domain)
        )
    elif runner == "playwright-ts":
        wave = "P138/TASK-196" if source_path in COMPLEX_E2E_SOURCES else "P138/TASK-195"
    elif runner == "playwright-support-ts":
        wave = "P137/TASK-194"
    else:
        wave = None
    if wave is None:
        raise ValueError(
            f"no owner wave for {runner}/{domain}/{target_class}: {source_path}"
        )
    return wave


def base_row(
    *,
    source_path: str,
    source_hash: str,
    runner: str,
    domain: str,
    test_ids: list[str],
    skip_sites: list[str],
    dependencies: list[str],
    services: list[str],
    target_class: str,
    target_path: str,
    coverage_bearing: bool,
) -> dict[str, Any]:
    return {
        "source_path": source_path,
        "sha256": source_hash,
        "runner": runner,
        "domain": domain,
        "owner_wave": canonical_owner_wave(source_path, runner, domain, target_class),
        "test_ids": test_ids,
        "test_count": len(test_ids),
        "skips": {"count": len(skip_sites), "sites": skip_sites},
        "service_hints": services,
        "dependency_hints": dependencies,
        "target_class": target_class,
        "target_path": target_path,
        "coverage_bearing": coverage_bearing,
        "hook_consumers": [],
        "baseline_status": "not-executed-in-task-202",
        "replacement_status": replacement_status(target_path),
        "parity_status": "not-assessed",
        "disposition": "retain-until-parity",
        "evidence": [f"sha256:{source_hash}", f"declared-tests:{len(test_ids)}"],
        "trace_ids": sorted(set(TRACE_ID_RE.findall("\n".join(test_ids)))),
    }


def go_imports_by_module(text: str, package_names: dict[str, str]) -> dict[str, str]:
    result = {}
    marker = "/internal/"
    for alias, import_path in go_import_pairs(text):
        if marker not in import_path:
            continue
        module = import_path.split(marker, 1)[1]
        if module not in package_names:
            continue
        if alias == "_":
            continue
        result[module] = alias if alias and alias != "." else ("." if alias == "." else package_names[module])
    return result


def hook_inventory(
    go_sources: list[tuple[Path, str]],
    *,
    hook_paths: list[Path] | None = None,
    snapshot: dict[str, SourceSnapshotEntry] | None = None,
) -> tuple[list[dict[str, Any]], dict[str, list[dict[str, str]]]]:
    hook_paths = (
        sorted_files(SOURCE_ROOTS["hooks"], "export_hooks*.go")
        if hook_paths is None
        else hook_paths
    )
    raw_hooks = []
    package_names: dict[str, str] = {}
    for path in hook_paths:
        data, text = read_allowed(path, SOURCE_ROOTS["hooks"], snapshot=snapshot)
        relative = path.relative_to((REPO_ROOT / SOURCE_ROOTS["hooks"]).resolve())
        module = posix(relative.parent)
        package_match = re.search(r"(?m)^package\s+([A-Za-z0-9_]+)\s*$", text)
        package_name = package_match.group(1) if package_match else relative.parent.name
        package_names[module] = package_name
        raw_hooks.append((path, data, text, module, package_name))

    consumers_by_test: dict[str, list[dict[str, str]]] = {rel(path): [] for path, _ in go_sources}
    records = []
    records_by_module: dict[str, list[dict[str, Any]]] = {}
    for path, data, text, module, package_name in raw_hooks:
        symbols = sorted(set(HOOK_DECL_RE.findall(text)))
        record = {
            "source_path": rel(path),
            "sha256": sha256_bytes(data),
            "module": module,
            "package": package_name,
            "canonical": path.name == "export_hooks.go",
            "symbols": [{"name": symbol, "consumer_test_files": []} for symbol in symbols],
            "deletion_decision": "none-in-task-202",
        }
        records.append(record)
        records_by_module.setdefault(module, []).append(record)

    # Resolve only modules actually imported by each test.  This is both more
    # precise and dramatically cheaper than comparing every exported symbol to
    # every test file.
    for test_path, test_text in go_sources:
        test_rel = rel(test_path)
        for module, alias in go_imports_by_module(test_text, package_names).items():
            for record in records_by_module.get(module, []):
                for symbol_record in record["symbols"]:
                    symbol = symbol_record["name"]
                    needle = symbol if alias == "." else f"{alias}.{symbol}"
                    if needle not in test_text:
                        continue
                    usage = rf"(?<![A-Za-z0-9_]){re.escape(symbol)}\b" if alias == "." else rf"\b{re.escape(alias)}\.{re.escape(symbol)}\b"
                    if re.search(usage, test_text):
                        symbol_record["consumer_test_files"].append(test_rel)
                        consumers_by_test[test_rel].append(
                            {"hook_source": record["source_path"], "symbol": symbol}
                        )
    for record in records:
        for symbol_record in record["symbols"]:
            symbol_record["consumer_test_files"].sort()
    for links in consumers_by_test.values():
        links.sort(key=lambda item: (item["hook_source"], item["symbol"]))
    return records, consumers_by_test


def target_python_test_count(paths: list[Path] | None = None) -> int:
    candidates = (
        sorted_files(Path(TARGET_ROOTS["python"]), "*.py")
        if paths is None
        else paths
    )
    return sum(
        candidate.name.startswith("test_") or candidate.name.endswith("_test.py")
        for candidate in candidates
    )


def discover_frontend_support_paths() -> list[Path]:
    return sorted(
        {
            path
            for pattern in ("render.tsx", "*.fixture.ts", "*.fixture.tsx")
            for path in sorted_files(SOURCE_ROOTS["vitest"], pattern)
        },
        key=rel,
    )


def discover_e2e_support_paths() -> list[Path]:
    e2e_domain_root = (REPO_ROOT / SOURCE_ROOTS["playwright"]).resolve()
    return sorted(
        {
            path
            for pattern in ("*.ts", "*.tsx")
            for path in sorted_files(SOURCE_ROOTS["playwright_support"], pattern)
            if not path.is_relative_to(e2e_domain_root)
        },
        key=rel,
    )


def discover_inventory_path_plan() -> dict[str, list[Path]]:
    return {
        "go": sorted_files(SOURCE_ROOTS["go"], "*_test.go"),
        "go_server": sorted_files(SOURCE_ROOTS["go_server"], "*_test.go"),
        "vitest": sorted(
            {
                path
                for pattern in ("*.test.ts", "*.test.tsx", "*.spec.ts", "*.spec.tsx")
                for path in sorted_files(SOURCE_ROOTS["vitest"], pattern)
            },
            key=rel,
        ),
        "frontend_support": discover_frontend_support_paths(),
        "e2e": sorted_files(SOURCE_ROOTS["playwright"], "*.spec.ts"),
        "e2e_support": discover_e2e_support_paths(),
        "hooks": sorted_files(SOURCE_ROOTS["hooks"], "export_hooks*.go"),
        "target_python": sorted_files(Path(TARGET_ROOTS["python"]), "*.py"),
    }


def git_inventory_path_plan(tree_paths: Iterable[str]) -> dict[str, list[Path]]:
    paths = sorted(set(tree_paths))
    excluded = {"node_modules", ".git", "coverage", "dist"}

    def below(path: str, root: Path) -> bool:
        prefix = root.as_posix().rstrip("/") + "/"
        return path.startswith(prefix)

    def selected(root: Path, patterns: tuple[str, ...]) -> list[Path]:
        return [
            REPO_ROOT / path
            for path in paths
            if below(path, root)
            and not any(part in excluded for part in PurePosixPath(path).parts)
            and any(fnmatch.fnmatch(PurePosixPath(path).name, item) for item in patterns)
        ]

    e2e_root = SOURCE_ROOTS["playwright_support"]
    e2e_domain_prefix = SOURCE_ROOTS["playwright"].as_posix().rstrip("/") + "/"
    return {
        "go": selected(SOURCE_ROOTS["go"], ("*_test.go",)),
        "go_server": selected(SOURCE_ROOTS["go_server"], ("*_test.go",)),
        "vitest": selected(
            SOURCE_ROOTS["vitest"],
            ("*.test.ts", "*.test.tsx", "*.spec.ts", "*.spec.tsx"),
        ),
        "frontend_support": selected(
            SOURCE_ROOTS["vitest"], ("render.tsx", "*.fixture.ts", "*.fixture.tsx")
        ),
        "e2e": selected(SOURCE_ROOTS["playwright"], ("*.spec.ts",)),
        "e2e_support": [
            REPO_ROOT / path
            for path in paths
            if below(path, e2e_root)
            and not any(part in excluded for part in PurePosixPath(path).parts)
            and not path.startswith(e2e_domain_prefix)
            and PurePosixPath(path).suffix in {".ts", ".tsx"}
        ],
        "hooks": selected(SOURCE_ROOTS["hooks"], ("export_hooks*.go",)),
        "target_python": selected(Path(TARGET_ROOTS["python"]), ("*.py",)),
    }


def path_plan_signature(plan: dict[str, list[Path]]) -> dict[str, list[str]]:
    return {key: [rel(path) for path in values] for key, values in sorted(plan.items())}


def path_plan_specs(plan: dict[str, list[Path]]) -> dict[str, tuple[Path, Path]]:
    roots = {
        "go": SOURCE_ROOTS["go"],
        "go_server": SOURCE_ROOTS["go_server"],
        "vitest": SOURCE_ROOTS["vitest"],
        "frontend_support": SOURCE_ROOTS["vitest"],
        "e2e": SOURCE_ROOTS["playwright"],
        "e2e_support": SOURCE_ROOTS["playwright_support"],
        "hooks": SOURCE_ROOTS["hooks"],
        "target_python": Path(TARGET_ROOTS["python"]),
    }
    result: dict[str, tuple[Path, Path]] = {}
    for group, paths in plan.items():
        for path in paths:
            source_path = rel(path)
            candidate = (path, roots[group])
            previous = result.get(source_path)
            if previous is not None and previous != candidate:
                raise ValueError(f"inventory source belongs to conflicting roots: {source_path}")
            result[source_path] = candidate
    return result


def build_git_source_snapshot(
    plan: dict[str, list[Path]], tree: dict[str, tuple[str, str, int]]
) -> dict[str, SourceSnapshotEntry]:
    specs = path_plan_specs(plan)
    requested_by_oid: dict[str, int] = {}
    total = 0
    for source_path in sorted(specs):
        metadata = tree.get(source_path)
        if metadata is None:
            raise ValueError(f"worktree-selected inventory source is untracked at pinned HEAD: {source_path}")
        oid, mode, size = metadata
        if mode not in {"100644", "100755"}:
            raise ValueError(f"inventory source is not a regular Git blob: {source_path}")
        if size > MAX_SOURCE_FILE_BYTES:
            raise ValueError(f"pinned Git source exceeds {MAX_SOURCE_FILE_BYTES} byte budget: {source_path}")
        total += size
        if total > MAX_SOURCE_SNAPSHOT_BYTES:
            raise ValueError(
                f"pinned Git source snapshot exceeds {MAX_SOURCE_SNAPSHOT_BYTES} byte budget"
            )
        previous_size = requested_by_oid.setdefault(oid, size)
        if previous_size != size:
            raise ValueError(f"Git blob size is inconsistent for {oid}")
    blobs = read_git_blobs(sorted(requested_by_oid.items()))
    snapshot: dict[str, SourceSnapshotEntry] = {}
    for source_path, (_, allowed_root) in specs.items():
        oid, mode, _ = tree[source_path]
        data = blobs[oid]
        snapshot[source_path] = SourceSnapshotEntry(
            allowed_root=allowed_root,
            blob_oid=oid,
            mode=mode,
            data=data,
            sha256=sha256_bytes(data),
        )
    return snapshot


def verify_worktree_against_snapshot(
    snapshot: dict[str, SourceSnapshotEntry],
    *,
    repo_root: Path = REPO_ROOT,
    after_read: Callable[[str, int], None] | None = None,
) -> None:
    total = 0
    for index, source_path in enumerate(sorted(snapshot)):
        entry = snapshot[source_path]
        data, _ = read_bounded_allowed_with_identity(
            repo_root / source_path,
            entry.allowed_root,
            max_bytes=MAX_SOURCE_FILE_BYTES,
            repo_root=repo_root,
        )
        total += len(data)
        if total > MAX_SOURCE_SNAPSHOT_BYTES:
            raise ValueError("worktree verification exceeded coherent snapshot byte budget")
        if sha256_bytes(data) != entry.sha256 or data != entry.data:
            raise ValueError(f"worktree differs from pinned Git source snapshot: {source_path}")
        if after_read is not None:
            after_read(source_path, index)


def assert_head_unchanged(pinned_head: str, head_reader: Callable[[], str] = revision) -> None:
    final_head = head_reader()
    if final_head != pinned_head:
        raise ValueError(f"repository HEAD changed during coherent inventory build: {pinned_head} -> {final_head}")


def _build_inventory_from_snapshot(
    plan: dict[str, list[Path]],
    snapshot: dict[str, SourceSnapshotEntry],
    pinned_head: str,
) -> dict[str, Any]:
    go_paths = plan["go"]
    go_server_paths = plan["go_server"]
    vitest_paths = plan["vitest"]
    frontend_support_paths = plan["frontend_support"]
    e2e_paths = plan["e2e"]
    e2e_support_paths = plan["e2e_support"]
    hook_paths = plan["hooks"]
    target_python_paths = plan["target_python"]

    go_sources = []
    for path in go_paths:
        _, text = read_allowed(path, SOURCE_ROOTS["go"], snapshot=snapshot)
        relative = path.relative_to((REPO_ROOT / SOURCE_ROOTS["go"]).resolve())
        go_sources.append(
            (path, text, SOURCE_ROOTS["go"], relative, domain_for(relative, "go"), relative)
        )
    for path in go_server_paths:
        _, text = read_allowed(path, SOURCE_ROOTS["go_server"], snapshot=snapshot)
        relative = path.relative_to((REPO_ROOT / SOURCE_ROOTS["go_server"]).resolve())
        go_sources.append(
            (
                path,
                text,
                SOURCE_ROOTS["go_server"],
                relative,
                "platform",
                Path("server") / relative,
            )
        )
    go_sources.sort(key=lambda item: rel(item[0]))
    hooks, consumers_by_test = hook_inventory(
        [(path, text) for path, text, _, _, _, _ in go_sources],
        hook_paths=hook_paths,
        snapshot=snapshot,
    )

    rows = []
    for path, text, source_root, relative, domain, target_relative in go_sources:
        data, _ = read_allowed(path, source_root, snapshot=snapshot)
        dependencies = external_go_imports(text)
        component = is_go_component(text, dependencies)
        services = service_hints(text)
        if component:
            target_class = go_component_target_class(services, dependencies)
            target_path = go_component_target(target_relative, target_class)
        else:
            target_class = "go-unit"
            target_path = go_target(target_relative)
        row = base_row(
            source_path=rel(path),
            source_hash=sha256_bytes(data),
            runner="go-test",
            domain=domain,
            test_ids=GO_TEST_RE.findall(text),
            skip_sites=line_sites(text, GO_SKIP_RE),
            dependencies=dependencies,
            services=services,
            target_class=target_class,
            target_path=target_path,
            coverage_bearing=True,
        )
        if component:
            mapping_hints = services or dependencies or ["component-marker-fallback"]
            row["evidence"].append(
                f"conversion-target:{target_class};hints:{','.join(mapping_hints)}"
            )
        else:
            row["evidence"].append("target:go-unit")
        row["hook_consumers"] = consumers_by_test[row["source_path"]]
        row["trace_ids"] = sorted(set(TRACE_ID_RE.findall(text)))
        rows.append(row)

    for path in vitest_paths:
        data, text = read_allowed(path, SOURCE_ROOTS["vitest"], snapshot=snapshot)
        relative = path.relative_to((REPO_ROOT / SOURCE_ROOTS["vitest"]).resolve())
        static_contract = any(marker in text for marker in FRONTEND_STATIC_CONTRACT_MARKERS)
        target_class = "python-static-contract" if static_contract else "frontend-unit"
        target_path = (
            frontend_static_contract_target(relative)
            if static_contract
            else vitest_target(relative)
        )
        row = base_row(
            source_path=rel(path),
            source_hash=sha256_bytes(data),
            runner="vitest",
            domain=domain_for(relative, "vitest"),
            test_ids=ts_cases(text),
            skip_sites=line_sites(text, TS_SKIP_RE),
            dependencies=external_ts_imports(text),
            services=service_hints(text),
            target_class=target_class,
            target_path=target_path,
            coverage_bearing=not static_contract,
        )
        row["evidence"].append(
            "conversion-target:python-static-contract;hints:node-fs-path-process"
            if static_contract
            else "target:frontend-unit"
        )
        row["trace_ids"] = sorted(set(TRACE_ID_RE.findall(text)))
        rows.append(row)

    for path in frontend_support_paths:
        data, text = read_allowed(path, SOURCE_ROOTS["vitest"], snapshot=snapshot)
        relative = path.relative_to((REPO_ROOT / SOURCE_ROOTS["vitest"]).resolve())
        row = base_row(
            source_path=rel(path),
            source_hash=sha256_bytes(data),
            runner="vitest-support",
            domain=domain_for(relative, "vitest-support"),
            test_ids=[],
            skip_sites=[],
            dependencies=external_ts_imports(text),
            services=service_hints(text),
            target_class="frontend-support",
            target_path=frontend_support_target(relative),
            coverage_bearing=False,
        )
        row["evidence"].append("target:frontend-support")
        row["trace_ids"] = sorted(set(TRACE_ID_RE.findall(text)))
        rows.append(row)

    for path in e2e_paths:
        data, text = read_allowed(path, SOURCE_ROOTS["playwright"], snapshot=snapshot)
        relative = path.relative_to((REPO_ROOT / SOURCE_ROOTS["playwright"]).resolve())
        row = base_row(
            source_path=rel(path),
            source_hash=sha256_bytes(data),
            runner="playwright-ts",
            domain=domain_for(relative, "playwright"),
            test_ids=ts_cases(text),
            skip_sites=line_sites(text, TS_SKIP_RE),
            dependencies=external_ts_imports(text),
            services=service_hints(text),
            target_class="python-e2e",
            target_path=e2e_target(relative),
            coverage_bearing=False,
        )
        row["trace_ids"] = sorted(set(TRACE_ID_RE.findall(text)))
        rows.append(row)

    for path in e2e_support_paths:
        data, text = read_allowed(path, SOURCE_ROOTS["playwright_support"], snapshot=snapshot)
        relative = path.relative_to(
            (REPO_ROOT / SOURCE_ROOTS["playwright_support"]).resolve()
        )
        row = base_row(
            source_path=rel(path),
            source_hash=sha256_bytes(data),
            runner="playwright-support-ts",
            domain="support",
            test_ids=[],
            skip_sites=[],
            dependencies=external_ts_imports(text),
            services=service_hints(text),
            target_class="python-e2e-support",
            target_path=e2e_support_target(relative),
            coverage_bearing=False,
        )
        row["evidence"].append("conversion-target:python-e2e-support")
        row["trace_ids"] = sorted(set(TRACE_ID_RE.findall(text)))
        rows.append(row)

    rows.sort(key=lambda item: item["source_path"])
    counts = {
        "go_files": len(go_paths) + len(go_server_paths),
        "go_legacy_files": len(go_paths),
        "go_server_files": len(go_server_paths),
        "go_declared_tests": sum(row["test_count"] for row in rows if row["runner"] == "go-test"),
        "go_legacy_declared_tests": sum(
            row["test_count"]
            for row in rows
            if str(row["source_path"]).startswith("backend/testframework/")
        ),
        "go_server_declared_tests": sum(
            row["test_count"]
            for row in rows
            if str(row["source_path"]).startswith("backend/cmd/server/")
        ),
        "go_unit_files": sum(row["target_class"] == "go-unit" for row in rows),
        "go_component_files": sum(row["target_class"] in PYTHON_COMPONENT_CLASSES for row in rows),
        "go_component_api_files": sum(row["target_class"] == "python-api" for row in rows),
        "go_component_integration_files": sum(
            row["target_class"] == "python-integration" for row in rows
        ),
        "go_component_contract_files": sum(
            row["target_class"] == "python-contract" for row in rows
        ),
        "vitest_files": len(vitest_paths),
        "vitest_declared_tests": sum(row["test_count"] for row in rows if row["runner"] == "vitest"),
        "frontend_unit_files": sum(row["target_class"] == "frontend-unit" for row in rows),
        "frontend_static_contract_files": sum(
            row["target_class"] == "python-static-contract" for row in rows
        ),
        "frontend_support_files": len(frontend_support_paths),
        "playwright_files": len(e2e_paths),
        "playwright_cases": sum(row["test_count"] for row in rows if row["runner"] == "playwright-ts"),
        "playwright_skip_sites": sum(row["skips"]["count"] for row in rows if row["runner"] == "playwright-ts"),
        "playwright_support_files": len(e2e_support_paths),
        "canonical_hook_files": sum(item["canonical"] for item in hooks),
        "hook_fragment_files": sum(not item["canonical"] for item in hooks),
        "hook_files_total": len(hooks),
        "hook_symbols": sum(len(item["symbols"]) for item in hooks),
        "hook_symbols_with_consumers": sum(
            bool(symbol["consumer_test_files"]) for item in hooks for symbol in item["symbols"]
        ),
        "target_python_tests": target_python_test_count(target_python_paths),
        "unmapped_files": sum(not row.get("target_path") for row in rows),
    }
    return {
        "schema_version": 2,
        "tool_version": TOOL_VERSION,
        "revision": pinned_head,
        "roots": {
            "sources": {key: posix(value) for key, value in SOURCE_ROOTS.items()},
            "targets": TARGET_ROOTS,
            "protected_exclusion": PROTECTED_PREFIX,
        },
        "expected": EXPECTED,
        "counts": counts,
        "rows": rows,
        "hooks": hooks,
    }


def live_test_paths_from_plan(plan: dict[str, list[Path]]) -> set[str]:
    return {
        rel(path)
        for group in (
            "go",
            "go_server",
            "vitest",
            "frontend_support",
            "e2e",
            "e2e_support",
        )
        for path in plan[group]
    }


def build_inventory_coherent(
    *,
    head_reader: Callable[[], str] = revision,
    after_verify_read: Callable[[str, int], None] | None = None,
) -> tuple[dict[str, Any], str, set[str]]:
    """Classify one immutable Git tree and prove the worktree exactly matches it."""

    pinned_head = head_reader()
    worktree_plan = discover_inventory_path_plan()
    tree = git_tree_blobs(pinned_head)
    git_plan = git_inventory_path_plan(tree)
    if path_plan_signature(worktree_plan) != path_plan_signature(git_plan):
        raise ValueError(
            "worktree inventory source set differs from pinned Git tree "
            "(untracked/deleted/renamed source rejected)"
        )
    snapshot = build_git_source_snapshot(git_plan, tree)
    data = _build_inventory_from_snapshot(git_plan, snapshot, pinned_head)

    before_verify = discover_inventory_path_plan()
    if path_plan_signature(before_verify) != path_plan_signature(git_plan):
        raise ValueError("inventory source set changed before full worktree verification")
    verify_worktree_against_snapshot(snapshot, after_read=after_verify_read)
    after_verify = discover_inventory_path_plan()
    if path_plan_signature(after_verify) != path_plan_signature(git_plan):
        raise ValueError("inventory source set changed during full worktree verification")
    assert_head_unchanged(pinned_head, head_reader)
    return data, pinned_head, live_test_paths_from_plan(git_plan)


def build_inventory() -> dict[str, Any]:
    return build_inventory_coherent()[0]


def strings(value: Any) -> Iterable[str]:
    if isinstance(value, str):
        yield value
    elif isinstance(value, dict):
        for item in value.values():
            yield from strings(item)
    elif isinstance(value, list):
        for item in value:
            yield from strings(item)


SCHEMA_KEYWORDS = {
    "$schema",
    "$id",
    "$defs",
    "$ref",
    "title",
    "type",
    "required",
    "properties",
    "additionalProperties",
    "items",
    "enum",
    "const",
    "pattern",
    "minimum",
}


def resolve_local_ref(root_schema: dict[str, Any], reference: str) -> dict[str, Any]:
    if not reference.startswith("#/"):
        raise ValueError(f"only local schema refs are supported: {reference}")
    current: Any = root_schema
    for raw_part in reference[2:].split("/"):
        part = raw_part.replace("~1", "/").replace("~0", "~")
        if not isinstance(current, dict) or part not in current:
            raise ValueError(f"unresolved schema ref: {reference}")
        current = current[part]
    if not isinstance(current, dict):
        raise ValueError(f"schema ref does not resolve to an object: {reference}")
    return current


def schema_contract_errors(
    value: Any,
    schema: dict[str, Any],
    *,
    root_schema: dict[str, Any],
    path: str = "$",
) -> list[str]:
    """Validate the explicit JSON-Schema subset used by inventory.schema.json.

    This is intentionally not advertised as a general JSON Schema engine.  It
    rejects unsupported assertion keywords so the check cannot silently pretend
    to enforce a contract it does not understand.
    """

    errors = []
    unsupported = sorted(set(schema) - SCHEMA_KEYWORDS)
    if unsupported:
        errors.append(f"{path}: unsupported schema keywords {unsupported}")
        return errors
    if "$ref" in schema:
        try:
            referenced = resolve_local_ref(root_schema, schema["$ref"])
        except ValueError as error:
            return [f"{path}: {error}"]
        return schema_contract_errors(value, referenced, root_schema=root_schema, path=path)

    expected_type = schema.get("type")
    type_checks = {
        "object": lambda item: isinstance(item, dict),
        "array": lambda item: isinstance(item, list),
        "string": lambda item: isinstance(item, str),
        "integer": lambda item: isinstance(item, int) and not isinstance(item, bool),
        "boolean": lambda item: isinstance(item, bool),
    }
    if expected_type is not None:
        checker = type_checks.get(expected_type)
        if checker is None:
            errors.append(f"{path}: unsupported schema type {expected_type!r}")
            return errors
        if not checker(value):
            errors.append(f"{path}: expected {expected_type}, got {type(value).__name__}")
            return errors

    if "const" in schema and value != schema["const"]:
        errors.append(f"{path}: expected const {schema['const']!r}, got {value!r}")
    if "enum" in schema and value not in schema["enum"]:
        errors.append(f"{path}: value {value!r} is not in enum {schema['enum']!r}")
    if "pattern" in schema and isinstance(value, str) and re.search(schema["pattern"], value) is None:
        errors.append(f"{path}: string does not match pattern {schema['pattern']!r}")
    if "minimum" in schema and isinstance(value, int) and value < schema["minimum"]:
        errors.append(f"{path}: value {value} is below minimum {schema['minimum']}")

    if isinstance(value, dict):
        required = schema.get("required", [])
        for key in required:
            if key not in value:
                errors.append(f"{path}: missing required property {key!r}")
        properties = schema.get("properties", {})
        for key, child_schema in properties.items():
            if key in value:
                errors.extend(
                    schema_contract_errors(
                        value[key], child_schema, root_schema=root_schema, path=f"{path}.{key}"
                    )
                )
        if schema.get("additionalProperties") is False:
            for key in sorted(set(value) - set(properties)):
                errors.append(f"{path}: additional property {key!r} is forbidden")
    if isinstance(value, list) and "items" in schema:
        for index, item in enumerate(value):
            errors.extend(
                schema_contract_errors(
                    item, schema["items"], root_schema=root_schema, path=f"{path}[{index}]"
                )
            )
    return errors


def validate_inventory(
    data: dict[str, Any],
    *,
    live_source_paths: set[str] | None = None,
    enforce_expected: bool = True,
    current_revision: str | None = None,
) -> list[str]:
    errors = []
    rows = data.get("rows", [])
    if not isinstance(rows, list):
        return ["inventory rows must be a list"]
    if len(rows) > MAX_INVENTORY_ROWS:
        return [
            f"inventory rows exceed {MAX_INVENTORY_ROWS} row budget: {len(rows)}"
        ]
    if any(not isinstance(row, dict) for row in rows):
        return ["inventory rows must contain only objects"]
    source_paths = [row.get("source_path") for row in rows]
    unique_paths = set(source_paths)
    if len(unique_paths) != len(source_paths):
        duplicates = sorted(path for path in unique_paths if source_paths.count(path) > 1)
        errors.append(f"duplicate source rows: {duplicates}")
    if live_source_paths is not None:
        missing = sorted(live_source_paths - unique_paths)
        extra = sorted(unique_paths - live_source_paths)
        if missing:
            errors.append(f"missing source rows: {missing}")
        if extra:
            errors.append(f"unknown source rows: {extra}")
    target_paths = [row.get("target_path") for row in rows]
    duplicate_targets = sorted(path for path in set(target_paths) if target_paths.count(path) > 1)
    if duplicate_targets:
        errors.append(f"duplicate target paths: {duplicate_targets}")

    allowed = {
        "go-test": {
            "go-unit": TARGET_ROOTS["go"],
            "python-api": TARGET_ROOTS["api"],
            "python-integration": TARGET_ROOTS["integration"],
            "python-contract": TARGET_ROOTS["contract"],
        },
        "vitest": {
            "frontend-unit": TARGET_ROOTS["vitest"],
            "python-static-contract": TARGET_ROOTS["frontend_contract"],
        },
        "vitest-support": {
            "frontend-support": f"{TARGET_ROOTS['vitest']}support/",
        },
        "playwright-ts": {"python-e2e": TARGET_ROOTS["playwright"]},
        "playwright-support-ts": {
            "python-e2e-support": TARGET_ROOTS["playwright_support"],
        },
    }
    for row in rows:
        runner = row.get("runner")
        if runner not in allowed:
            errors.append(f"bad runner for {row.get('source_path')}: {runner}")
            continue
        classes = allowed[runner]
        target_class = row.get("target_class")
        if target_class not in classes:
            errors.append(f"bad target class for {row.get('source_path')}: {row.get('target_class')}")
            prefix = ""
        else:
            prefix = classes[target_class]
        try:
            expected_owner = canonical_owner_wave(
                str(row.get("source_path")),
                str(runner),
                str(row.get("domain")),
                str(target_class),
            )
        except ValueError as error:
            errors.append(str(error))
        else:
            if row.get("owner_wave") != expected_owner or row.get("owner_wave") not in OWNER_WAVES:
                errors.append(
                    f"bad owner wave for {row.get('source_path')}: "
                    f"expected={expected_owner} actual={row.get('owner_wave')}"
                )
        if not prefix or not str(row.get("target_path", "")).startswith(prefix):
            errors.append(f"bad target root for {row.get('source_path')}: {row.get('target_path')}")
        target_path = str(row.get("target_path", ""))
        python_classes = PYTHON_COMPONENT_CLASSES | {
            "python-static-contract",
            "python-e2e",
            "python-e2e-support",
        }
        if target_class in python_classes and not target_path.endswith(".py"):
            errors.append(f"non-Python target for {row.get('source_path')}: {row.get('target_path')}")
        if target_class == "go-unit" and not target_path.endswith("_test.go"):
            errors.append(f"non-Go unit target for {row.get('source_path')}: {row.get('target_path')}")
        if target_class in {"frontend-unit", "frontend-support"} and not target_path.endswith(
            (".ts", ".tsx")
        ):
            errors.append(f"non-TypeScript frontend target for {row.get('source_path')}: {target_path}")
        expected_coverage = runner == "go-test" or target_class == "frontend-unit"
        if row.get("coverage_bearing") is not expected_coverage:
            errors.append(f"bad coverage-bearing flag for {row.get('source_path')}")
        if row.get("test_count") != len(row.get("test_ids", [])):
            errors.append(f"test count mismatch for {row.get('source_path')}")
        skips = row.get("skips", {})
        if skips.get("count") != len(skips.get("sites", [])):
            errors.append(f"skip count mismatch for {row.get('source_path')}")
        if row.get("disposition") != "retain-until-parity":
            errors.append(f"unsafe disposition for {row.get('source_path')}: {row.get('disposition')}")
        if row.get("parity_status") != "not-assessed":
            errors.append(f"unproven parity status for {row.get('source_path')}: {row.get('parity_status')}")
        if row.get("replacement_status") not in {"not-started", "candidate-present"}:
            errors.append(
                f"bad replacement status for {row.get('source_path')}: {row.get('replacement_status')}"
            )
        if not row.get("evidence"):
            errors.append(f"missing evidence for {row.get('source_path')}")

    counts = data.get("counts", {})
    calculated = {
        "go_files": sum(row.get("runner") == "go-test" for row in rows),
        "go_legacy_files": sum(
            str(row.get("source_path", "")).startswith("backend/testframework/")
            for row in rows
        ),
        "go_server_files": sum(
            str(row.get("source_path", "")).startswith("backend/cmd/server/")
            for row in rows
        ),
        "go_declared_tests": sum(
            row.get("test_count", 0) for row in rows if row.get("runner") == "go-test"
        ),
        "go_legacy_declared_tests": sum(
            row.get("test_count", 0)
            for row in rows
            if str(row.get("source_path", "")).startswith("backend/testframework/")
        ),
        "go_server_declared_tests": sum(
            row.get("test_count", 0)
            for row in rows
            if str(row.get("source_path", "")).startswith("backend/cmd/server/")
        ),
        "go_unit_files": sum(row.get("target_class") == "go-unit" for row in rows),
        "go_component_files": sum(
            row.get("target_class") in PYTHON_COMPONENT_CLASSES for row in rows
        ),
        "go_component_api_files": sum(row.get("target_class") == "python-api" for row in rows),
        "go_component_integration_files": sum(
            row.get("target_class") == "python-integration" for row in rows
        ),
        "go_component_contract_files": sum(
            row.get("target_class") == "python-contract" for row in rows
        ),
        "vitest_files": sum(row.get("runner") == "vitest" for row in rows),
        "frontend_unit_files": sum(row.get("target_class") == "frontend-unit" for row in rows),
        "frontend_static_contract_files": sum(
            row.get("target_class") == "python-static-contract" for row in rows
        ),
        "frontend_support_files": sum(row.get("runner") == "vitest-support" for row in rows),
        "playwright_files": sum(row.get("runner") == "playwright-ts" for row in rows),
        "playwright_cases": sum(row.get("test_count", 0) for row in rows if row.get("runner") == "playwright-ts"),
        "playwright_support_files": sum(
            row.get("runner") == "playwright-support-ts" for row in rows
        ),
        "canonical_hook_files": sum(hook.get("canonical") is True for hook in data.get("hooks", [])),
        "unmapped_files": sum(not row.get("target_path") for row in rows),
    }
    for key, value in calculated.items():
        if counts.get(key) != value:
            errors.append(f"count reconciliation failed for {key}: metadata={counts.get(key)} rows={value}")
    if enforce_expected and (counts.get("playwright_files") != 20 or counts.get("playwright_cases") != 53):
        errors.append(
            "Playwright reconciliation failed: "
            f"files={counts.get('playwright_files')} declared_cases={counts.get('playwright_cases')}"
        )
    if enforce_expected:
        if data.get("tool_version") != TOOL_VERSION:
            errors.append(f"tool version mismatch: ledger={data.get('tool_version')} current={TOOL_VERSION}")
        ledger_revision = str(data.get("revision", ""))
        checked_revision = current_revision if current_revision is not None else revision()
        if not revision_is_ancestor(ledger_revision, checked_revision):
            errors.append(
                "ledger revision is not an ancestor of pinned HEAD: "
                f"ledger={ledger_revision} current={checked_revision}"
            )
        for key, value in EXPECTED.items():
            if counts.get(key) != value:
                errors.append(f"expected baseline drift for {key}: expected={value} actual={counts.get(key)}")

    row_lookup = {row.get("source_path"): row for row in rows}
    hooks = data.get("hooks", [])
    hook_sources = [hook.get("source_path") for hook in hooks]
    duplicate_hooks = sorted(path for path in set(hook_sources) if hook_sources.count(path) > 1)
    if duplicate_hooks:
        errors.append(f"duplicate hook files: {duplicate_hooks}")
    forward_links = set()
    for hook in hooks:
        hook_source = hook.get("source_path")
        if not str(hook_source).startswith("backend/internal/") or not str(hook_source).split("/")[-1].startswith("export_hooks"):
            errors.append(f"bad hook source path: {hook_source}")
        for symbol in hook.get("symbols", []):
            name = symbol.get("name")
            for consumer in symbol.get("consumer_test_files", []):
                forward_links.add((hook_source, name, consumer))
                row = row_lookup.get(consumer)
                if row is None or row.get("runner") != "go-test":
                    errors.append(f"bad hook consumer link {hook_source}:{name} -> {consumer}")
                    continue
                expected_link = {"hook_source": hook_source, "symbol": name}
                if expected_link not in row.get("hook_consumers", []):
                    errors.append(f"missing reverse hook link {hook_source}:{name} -> {consumer}")
    for consumer, row in row_lookup.items():
        for link in row.get("hook_consumers", []):
            candidate = (link.get("hook_source"), link.get("symbol"), consumer)
            if candidate not in forward_links:
                errors.append(
                    "orphan reverse hook link "
                    f"{link.get('hook_source')}:{link.get('symbol')} <- {consumer}"
                )

    for value in strings(data):
        normalized = value.replace("\\", "/").lstrip("./")
        if normalized.startswith(PROTECTED_PREFIX) and value != PROTECTED_PREFIX:
            errors.append(f"protected path leaked into ledger: {value}")
    return errors


def render(data: dict[str, Any]) -> bytes:
    return (json.dumps(data, ensure_ascii=False, indent=2, sort_keys=True) + "\n").encode("utf-8")


def rename_noreplace(source_fd: int, source: str, target_fd: int, target: str) -> None:
    """Publish through Linux renameat2 without ever replacing authority data."""

    renameat2 = getattr(ctypes.CDLL(None, use_errno=True), "renameat2", None)
    if renameat2 is None:
        raise ValueError("renameat2(RENAME_NOREPLACE) is unavailable")
    renameat2.argtypes = [ctypes.c_int, ctypes.c_char_p, ctypes.c_int, ctypes.c_char_p, ctypes.c_uint]
    renameat2.restype = ctypes.c_int
    if renameat2(source_fd, os.fsencode(source), target_fd, os.fsencode(target), 1) == 0:
        return
    number = ctypes.get_errno()
    if number in {errno.ENOSYS, errno.EINVAL, errno.ENOTSUP}:
        raise ValueError("renameat2(RENAME_NOREPLACE) is unsupported")
    if number in {errno.EEXIST, errno.ENOTEMPTY}:
        raise FileExistsError(number, "inventory target already exists", target)
    raise OSError(number, os.strerror(number), target)


def read_inventory_target(parent_fd: int, target_name: str, expected: os.stat_result) -> bytes:
    descriptor = os.open(
        target_name,
        os.O_RDONLY | os.O_CLOEXEC | os.O_NOFOLLOW,
        dir_fd=parent_fd,
    )
    try:
        opened = os.fstat(descriptor)
        if (
            is_symlink_or_reparse(opened)
            or not stat.S_ISREG(opened.st_mode)
            or int(opened.st_nlink) != 1
            or metadata_identity(opened, include_content=True)
            != metadata_identity(expected, include_content=True)
        ):
            raise ValueError("canonical inventory target changed before descriptor read")
        if int(opened.st_size) > MAX_JSON_BYTES:
            raise ValueError("canonical inventory target exceeds JSON byte budget")
        result = bytearray()
        while len(result) <= MAX_JSON_BYTES:
            chunk = os.read(descriptor, min(1024 * 1024, MAX_JSON_BYTES + 1 - len(result)))
            if not chunk:
                break
            result.extend(chunk)
        if len(result) > MAX_JSON_BYTES:
            raise ValueError("canonical inventory target exceeds JSON byte budget")
        if metadata_identity(os.fstat(descriptor), include_content=True) != metadata_identity(
            opened, include_content=True
        ):
            raise ValueError("canonical inventory target changed during descriptor read")
        return bytes(result)
    finally:
        os.close(descriptor)


class SimulatedInventoryPublicationCrash(BaseException):
    """Self-test only: model process loss after staging fsync, before publish."""


def write_inventory_atomically(
    payload: bytes,
    *,
    repo_root: Path = REPO_ROOT,
    output_path: Path | None = None,
    failure_mode: str | None = None,
) -> None:
    """Idempotently publish a missing ledger; never replace existing authority."""

    require_descriptor_traversal()
    audit_root = Path(".planning/audits/test-rebase")
    repository = Path(os.path.abspath(repo_root))
    canonical_output = repository / audit_root / "inventory.json"
    output = canonical_output if output_path is None else Path(os.path.abspath(output_path))
    if output != canonical_output:
        raise ValueError("inventory publication target must be canonical for the selected repository root")
    if failure_mode not in {None, "destination-race", "crash-after-stage-fsync"}:
        raise ValueError(f"unknown inventory publication failure mode: {failure_mode}")
    parent_fd, parent_meta, _ = open_existing_path(
        repository / audit_root,
        audit_root,
        repo_root=repository,
        require_directory=True,
    )
    target_name = output.name
    staging_name = f".inventory.{os.getpid()}.{secrets.token_hex(8)}.tmp"
    descriptor = -1
    preserve_staging = False
    try:
        try:
            existing = os.stat(target_name, dir_fd=parent_fd, follow_symlinks=False)
        except FileNotFoundError:
            existing = None
        if existing is not None:
            if is_symlink_or_reparse(existing) or not stat.S_ISREG(existing.st_mode):
                raise ValueError("canonical inventory target must be a physical regular file")
            if int(existing.st_nlink) != 1:
                raise ValueError("canonical inventory target link count must be one")
            if read_inventory_target(parent_fd, target_name, existing) != payload:
                raise ValueError(
                    "canonical inventory differs; no-overwrite publication requires explicit authority rotation"
                )
            if metadata_identity(parent_meta, include_content=False) != metadata_identity(
                os.fstat(parent_fd), include_content=False
            ):
                raise ValueError("inventory parent changed during idempotent verification")
            return

        flags = os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_CLOEXEC | os.O_NOFOLLOW
        descriptor = os.open(staging_name, flags, 0o600, dir_fd=parent_fd)
        staging_meta = os.fstat(descriptor)
        if not stat.S_ISREG(staging_meta.st_mode) or int(staging_meta.st_nlink) != 1:
            raise ValueError("inventory staging file is not an exclusive regular file")
        written = 0
        while written < len(payload):
            count = os.write(descriptor, payload[written:])
            if count <= 0:
                raise OSError("short write while staging inventory")
            written += count
        os.fsync(descriptor)
        # The identity snapshot must be taken after the payload lands: the
        # post-close comparison includes size/mtime/ctime, which change on write.
        staging_meta = os.fstat(descriptor)
        os.close(descriptor)
        descriptor = -1

        staged = os.stat(staging_name, dir_fd=parent_fd, follow_symlinks=False)
        if (
            is_symlink_or_reparse(staged)
            or not stat.S_ISREG(staged.st_mode)
            or int(staged.st_nlink) != 1
            or metadata_identity(staged, include_content=True)
            != metadata_identity(staging_meta, include_content=True)
        ):
            raise ValueError("inventory staging identity changed")
        verify_fd = os.open(
            staging_name,
            os.O_RDONLY | os.O_CLOEXEC | os.O_NOFOLLOW,
            dir_fd=parent_fd,
        )
        try:
            verified = bytearray()
            while len(verified) <= len(payload):
                chunk = os.read(verify_fd, min(1024 * 1024, len(payload) + 1 - len(verified)))
                if not chunk:
                    break
                verified.extend(chunk)
            if bytes(verified) != payload:
                raise ValueError("inventory staging payload verification failed")
            if metadata_identity(os.fstat(verify_fd), include_content=True) != metadata_identity(
                staged, include_content=True
            ):
                raise ValueError("inventory staging changed during verification")
        finally:
            os.close(verify_fd)
        if metadata_identity(parent_meta, include_content=False) != metadata_identity(
            os.fstat(parent_fd), include_content=False
        ):
            raise ValueError("inventory parent changed during generation (TOCTOU)")
        if failure_mode == "crash-after-stage-fsync":
            preserve_staging = True
            raise SimulatedInventoryPublicationCrash(staging_name)
        if failure_mode == "destination-race":
            racer_fd = os.open(
                target_name,
                os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_CLOEXEC | os.O_NOFOLLOW,
                0o600,
                dir_fd=parent_fd,
            )
            try:
                os.write(racer_fd, b"racer")
                os.fsync(racer_fd)
            finally:
                os.close(racer_fd)
            os.fsync(parent_fd)
        try:
            rename_noreplace(parent_fd, staging_name, parent_fd, target_name)
        except FileExistsError as error:
            raise ValueError("inventory target appeared during atomic publication") from error
        os.fsync(parent_fd)
    finally:
        if descriptor >= 0:
            os.close(descriptor)
        if not preserve_staging:
            try:
                os.unlink(staging_name, dir_fd=parent_fd)
            except FileNotFoundError:
                pass
        os.close(parent_fd)


def live_test_paths() -> set[str]:
    return live_test_paths_from_plan(discover_inventory_path_plan())


def run_check() -> int:
    audit_root = Path(".planning/audits/test-rebase")
    try:
        inventory_bytes = read_bounded_allowed(
            OUTPUT_PATH, audit_root, max_bytes=MAX_JSON_BYTES
        )
        schema_bytes = read_bounded_allowed(
            SCHEMA_PATH, audit_root, max_bytes=MAX_JSON_BYTES
        )
    except (OSError, ValueError) as problem:
        print(f"RED: ledger/schema descriptor read failed: {problem}", file=sys.stderr)
        return 1
    on_disk = json.loads(inventory_bytes)
    schema = json.loads(schema_bytes)
    expected, pinned_head, current_paths = build_inventory_coherent()
    errors = [
        f"schema contract: {error}"
        for error in schema_contract_errors(on_disk, schema, root_schema=schema)
    ]
    errors.extend(
        validate_inventory(
            on_disk,
            live_source_paths=current_paths,
            current_revision=pinned_head,
        )
    )
    # `revision` is provenance for the source snapshot, not a cachebuster for
    # every later planning-only commit. validate_inventory already proves that
    # it is an ancestor; compare all live inputs against that recorded value.
    expected["revision"] = on_disk.get("revision")
    expected_bytes = render(expected)
    actual_bytes = inventory_bytes
    if actual_bytes != expected_bytes:
        errors.append("generated ledger differs from current allow-listed inputs")
    if errors:
        for error in errors:
            print(f"RED: {error}", file=sys.stderr)
        return 1
    assert_head_unchanged(pinned_head)
    print(
        "GREEN: ledger matches allow-listed inputs; "
        f"sha256={sha256_bytes(actual_bytes)} rows={len(on_disk['rows'])} "
        f"LEDGER_HEAD={str(on_disk['revision'])[:8]}"
    )
    return 0


def sentinel_row(source_path: str) -> dict[str, Any]:
    return {
        "source_path": source_path,
        "sha256": "0" * 64,
        "runner": "go-test",
        "domain": "auth",
        "owner_wave": "P133/TASK-210",
        "test_ids": ["TestSentinel"],
        "test_count": 1,
        "skips": {"count": 0, "sites": []},
        "service_hints": [],
        "dependency_hints": [],
        "target_class": "go-unit",
        "target_path": "backend/tests/sentinel/sentinel_test.go",
        "coverage_bearing": True,
        "hook_consumers": [],
        "baseline_status": "sentinel",
        "replacement_status": "not-started",
        "parity_status": "not-assessed",
        "disposition": "retain-until-parity",
        "evidence": ["sentinel"],
        "trace_ids": [],
    }


def run_self_test() -> int:
    # TemporaryDirectory documents and enforces that the sentinel never writes
    # into the real audit directory or mutates inventory.json.
    with tempfile.TemporaryDirectory(prefix="hub-test-inventory-sentinel-") as temp_dir:
        sentinel_root = Path(temp_dir)
        publication_payload = b'{"sentinel":"inventory-publication"}\n'

        idempotent_root = sentinel_root / "publication-idempotent"
        idempotent_audit = idempotent_root / ".planning/audits/test-rebase"
        idempotent_audit.mkdir(parents=True)
        idempotent_output = idempotent_audit / "inventory.json"
        write_inventory_atomically(
            publication_payload,
            repo_root=idempotent_root,
            output_path=idempotent_output,
        )
        write_inventory_atomically(
            publication_payload,
            repo_root=idempotent_root,
            output_path=idempotent_output,
        )
        if idempotent_output.read_bytes() != publication_payload:
            print("RED: inventory idempotent publication changed canonical bytes", file=sys.stderr)
            return 1
        try:
            write_inventory_atomically(
                b'{"sentinel":"different"}\n',
                repo_root=idempotent_root,
                output_path=idempotent_output,
            )
        except ValueError as error:
            if "canonical inventory differs" not in str(error):
                print(f"RED: inventory conflict sentinel failed for wrong reason: {error}", file=sys.stderr)
                return 1
        else:
            print("RED: inventory conflict sentinel replaced existing authority", file=sys.stderr)
            return 1

        race_root = sentinel_root / "publication-race"
        race_audit = race_root / ".planning/audits/test-rebase"
        race_audit.mkdir(parents=True)
        race_output = race_audit / "inventory.json"
        try:
            write_inventory_atomically(
                publication_payload,
                repo_root=race_root,
                output_path=race_output,
                failure_mode="destination-race",
            )
        except ValueError as error:
            if "target appeared during atomic publication" not in str(error):
                print(f"RED: inventory publication-race failed for wrong reason: {error}", file=sys.stderr)
                return 1
        else:
            print("RED: inventory publication-race overwrote an appearing target", file=sys.stderr)
            return 1
        if race_output.read_bytes() != b"racer":
            print("RED: inventory publication-race did not preserve racer bytes", file=sys.stderr)
            return 1

        crash_root = sentinel_root / "publication-crash"
        crash_audit = crash_root / ".planning/audits/test-rebase"
        crash_audit.mkdir(parents=True)
        crash_output = crash_audit / "inventory.json"
        try:
            write_inventory_atomically(
                publication_payload,
                repo_root=crash_root,
                output_path=crash_output,
                failure_mode="crash-after-stage-fsync",
            )
        except SimulatedInventoryPublicationCrash:
            pass
        else:
            print("RED: inventory publication-crash sentinel did not interrupt", file=sys.stderr)
            return 1
        staged = list(crash_audit.glob(".inventory.*.tmp"))
        if crash_output.exists() or len(staged) != 1 or staged[0].read_bytes() != publication_payload:
            print("RED: inventory publication-crash exposed partial canonical bytes", file=sys.stderr)
            return 1

        safe_root = sentinel_root / "safe"
        safe_root.mkdir()
        safe_source = safe_root / "source_test.go"
        safe_source.write_text("package safe\n", encoding="utf-8")
        try:
            inspect_existing_path(
                safe_source,
                Path("safe"),
                repo_root=sentinel_root,
            )
        except ValueError as error:
            print(f"RED: clean physical source sentinel failed: {error}", file=sys.stderr)
            return 1

        coherent_root = sentinel_root / "coherent"
        coherent_root.mkdir()
        coherent_a = coherent_root / "a.txt"
        coherent_b = coherent_root / "b.txt"
        coherent_a.write_bytes(b"generation-zero-a")
        coherent_b.write_bytes(b"generation-zero-b")
        coherent_snapshot = {
            "coherent/a.txt": SourceSnapshotEntry(
                allowed_root=Path("coherent"),
                blob_oid="1" * 40,
                mode="100644",
                data=b"generation-zero-a",
                sha256=sha256_bytes(b"generation-zero-a"),
            ),
            "coherent/b.txt": SourceSnapshotEntry(
                allowed_root=Path("coherent"),
                blob_oid="2" * 40,
                mode="100644",
                data=b"generation-zero-b",
                sha256=sha256_bytes(b"generation-zero-b"),
            ),
        }

        def cross_file_mutation(_: str, index: int) -> None:
            if index == 0:
                coherent_a.write_bytes(b"generation-one-a")
                coherent_b.write_bytes(b"generation-one-b")

        try:
            verify_worktree_against_snapshot(
                coherent_snapshot,
                repo_root=sentinel_root,
                after_read=cross_file_mutation,
            )
        except ValueError as error:
            if "worktree differs from pinned Git source snapshot" not in str(error):
                print(f"RED: cross-file snapshot sentinel failed for wrong reason: {error}", file=sys.stderr)
                return 1
        else:
            print("RED: cross-file mutation produced a hybrid inventory snapshot", file=sys.stderr)
            return 1

        try:
            assert_head_unchanged("1" * 40, lambda: "2" * 40)
        except ValueError as error:
            if "HEAD changed during coherent inventory build" not in str(error):
                print(f"RED: HEAD-drift sentinel failed for wrong reason: {error}", file=sys.stderr)
                return 1
        else:
            print("RED: inventory HEAD drift was not rejected", file=sys.stderr)
            return 1

        hardlink_source = safe_root / "hardlink_test.go"
        hardlink_source.write_text("package safe\n", encoding="utf-8")
        os.link(hardlink_source, sentinel_root / "unlisted-external-alias.go")
        try:
            inspect_existing_path(
                hardlink_source,
                Path("safe"),
                repo_root=sentinel_root,
            )
        except ValueError as error:
            if "link count must be one" not in str(error):
                print(
                    f"RED: hardlink sentinel failed for wrong reason: {error}",
                    file=sys.stderr,
                )
                return 1
        else:
            print("RED: unlisted external hardlink alias was not rejected", file=sys.stderr)
            return 1

        try:
            read_bounded_allowed(
                safe_source,
                Path("safe"),
                max_bytes=4,
                repo_root=sentinel_root,
            )
        except ValueError as error:
            if "byte budget" not in str(error):
                print(f"RED: file-budget sentinel failed for wrong reason: {error}", file=sys.stderr)
                return 1
        else:
            print("RED: source file byte budget was not enforced", file=sys.stderr)
            return 1

        directory_budget_root = sentinel_root / "directory-budget"
        (directory_budget_root / "nested").mkdir(parents=True)
        try:
            sorted_files(
                Path("directory-budget"),
                "*",
                repo_root=sentinel_root,
                max_directories=1,
            )
        except ValueError as error:
            if "directory budget" not in str(error):
                print(
                    f"RED: directory-budget sentinel failed for wrong reason: {error}",
                    file=sys.stderr,
                )
                return 1
        else:
            print("RED: traversal directory budget was not enforced", file=sys.stderr)
            return 1

        file_budget_root = sentinel_root / "file-budget"
        file_budget_root.mkdir()
        (file_budget_root / "one.txt").write_text("one\n", encoding="utf-8")
        (file_budget_root / "two.txt").write_text("two\n", encoding="utf-8")
        try:
            sorted_files(
                Path("file-budget"),
                "*",
                repo_root=sentinel_root,
                max_files=1,
            )
        except ValueError as error:
            if "file budget" not in str(error):
                print(f"RED: file-budget sentinel failed for wrong reason: {error}", file=sys.stderr)
                return 1
        else:
            print("RED: traversal file budget was not enforced", file=sys.stderr)
            return 1

        linked_root = sentinel_root / "linked"
        try:
            linked_root.symlink_to(safe_root, target_is_directory=True)
            inspect_existing_path(
                linked_root / "source_test.go",
                Path("linked"),
                repo_root=sentinel_root,
            )
        except ValueError as error:
            if "symlink/reparse component rejected" not in str(error):
                print(
                    f"RED: physical source sentinel failed for wrong reason: {error}",
                    file=sys.stderr,
                )
                return 1
        except OSError as error:
            print(f"RED: symlink sentinel setup failed closed: {error}", file=sys.stderr)
            return 1
        else:
            print("RED: symlink/reparse source sentinel was not rejected", file=sys.stderr)
            return 1

        fake_reparse = type(
            "FakeReparseMetadata",
            (),
            {"st_mode": stat.S_IFREG, "st_file_attributes": FILE_ATTRIBUTE_REPARSE_POINT},
        )()
        if not is_symlink_or_reparse(fake_reparse):
            print("RED: Windows reparse metadata sentinel was not recognized", file=sys.stderr)
            return 1

        expected_e2e_target = "testframework/e2e/platform/test_critical_gate.py"
        actual_e2e_target = e2e_target(Path("platform/critical-gate.spec.ts"))
        if TARGET_ROOTS["python"] != "testframework/" or actual_e2e_target != expected_e2e_target:
            print(
                "RED: root Python/E2E target mapping drifted: "
                f"python={TARGET_ROOTS['python']} e2e={actual_e2e_target}",
                file=sys.stderr,
            )
            return 1
        mapping_checks = (
            (go_component_target_class(["http-server"], []), "python-api"),
            (go_component_target_class(["postgres"], []), "python-integration"),
            (
                go_component_target_class([], ["github.com/ory/dockertest/v3"]),
                "python-integration",
            ),
            (
                is_go_component(
                    "", ["github.com/valerii/hub/backend/testframework/env"]
                ),
                True,
            ),
            (
                go_component_target_class(
                    [], ["github.com/valerii/hub/backend/testframework/mongotest"]
                ),
                "python-integration",
            ),
            (go_component_target_class([], []), "python-contract"),
            (
                go_component_target(
                    Path("audit/coverage_gate_task225_test.go"), "python-integration"
                ),
                "testframework/integration/audit/test_coverage_gate_task225.py",
            ),
            (
                vitest_target(Path("domains/tracker/issue-card.test.tsx")),
                "frontend/tests/tracker/issue-card.test.tsx",
            ),
            (
                frontend_static_contract_target(
                    Path("domains/platform/docs-drift-task123.test.ts")
                ),
                "testframework/contract/frontend/platform/test_docs_drift_task123.py",
            ),
            (
                frontend_support_target(
                    Path("domains/tracker/projects-sprint-version-required.fixture.ts")
                ),
                "frontend/tests/support/tracker/projects-sprint-version-required.fixture.ts",
            ),
            (
                e2e_support_target(Path("shared/pages/base-page.ts")),
                "testframework/e2e/support/shared/pages/base_page.py",
            ),
            (
                go_target(Path("server/wiring_fiber_recover_test.go")),
                "backend/tests/server/wiring_fiber_recover_test.go",
            ),
            (
                canonical_owner_wave("sentinel", "go-test", "auth", "go-unit"),
                "P133/TASK-210",
            ),
            (
                canonical_owner_wave("sentinel", "go-test", "auth", "python-api"),
                "P133/TASK-227",
            ),
            (
                canonical_owner_wave(
                    "sentinel", "go-test", "tracker", "python-integration"
                ),
                "P134/TASK-228",
            ),
            (
                canonical_owner_wave("sentinel", "go-test", "ai", "python-contract"),
                "P135/TASK-229",
            ),
            (
                canonical_owner_wave(
                    "sentinel", "vitest", "platform", "python-static-contract"
                ),
                "P131/TASK-230",
            ),
        )
        if any(actual != expected for actual, expected in mapping_checks):
            print(f"RED: target/owner mapping drifted: {mapping_checks}", file=sys.stderr)
            return 1
        sentinel_path = "backend/testframework/sentinel/sentinel_test.go"
        clean = {
            "counts": {
                "go_files": 1,
                "go_legacy_files": 1,
                "go_server_files": 0,
                "go_declared_tests": 1,
                "go_legacy_declared_tests": 1,
                "go_server_declared_tests": 0,
                "go_unit_files": 1,
                "go_component_files": 0,
                "go_component_api_files": 0,
                "go_component_integration_files": 0,
                "go_component_contract_files": 0,
                "vitest_files": 0,
                "frontend_unit_files": 0,
                "frontend_static_contract_files": 0,
                "frontend_support_files": 0,
                "playwright_files": 0,
                "playwright_cases": 0,
                "playwright_support_files": 0,
                "canonical_hook_files": 0,
                "unmapped_files": 0,
            },
            "rows": [sentinel_row(sentinel_path)],
            "hooks": [],
            "sentinel_temp_dir": temp_dir,
        }
        clean_errors = validate_inventory(clean, live_source_paths={sentinel_path}, enforce_expected=False)
        duplicate = deepcopy(clean)
        duplicate["rows"].append(deepcopy(duplicate["rows"][0]))
        duplicate_errors = validate_inventory(duplicate, live_source_paths={sentinel_path}, enforce_expected=False)
        removed = deepcopy(clean)
        removed["rows"] = []
        removed_errors = validate_inventory(removed, live_source_paths={sentinel_path}, enforce_expected=False)
        misrouted = deepcopy(clean)
        misrouted["rows"][0]["target_class"] = "python-integration"
        misrouted_errors = validate_inventory(
            misrouted, live_source_paths={sentinel_path}, enforce_expected=False
        )
        oversized_rows_errors = validate_inventory(
            {"rows": [{} for _ in range(MAX_INVENTORY_ROWS + 1)]},
            enforce_expected=False,
        )
        if clean_errors:
            print(f"RED: clean sentinel unexpectedly failed: {clean_errors}", file=sys.stderr)
            return 1
        if not any("duplicate source rows" in error for error in duplicate_errors):
            print("RED: duplicate sentinel was not detected", file=sys.stderr)
            return 1
        if not any("missing source rows" in error for error in removed_errors):
            print("RED: removal sentinel was not detected", file=sys.stderr)
            return 1
        if not any("bad target root" in error for error in misrouted_errors):
            print("RED: component-to-backend misroute was not detected", file=sys.stderr)
            return 1
        if not any("row budget" in error for error in oversized_rows_errors):
            print("RED: inventory row budget was not enforced", file=sys.stderr)
            return 1
        schema = json.loads(
            read_bounded_allowed(
                SCHEMA_PATH,
                Path(".planning/audits/test-rebase"),
                max_bytes=MAX_JSON_BYTES,
            )
        )
        row_schema = {"$ref": "#/$defs/testRow"}
        clean_schema_errors = schema_contract_errors(
            clean["rows"][0], row_schema, root_schema=schema, path="$.sentinel"
        )
        missing_owner = deepcopy(clean["rows"][0])
        del missing_owner["owner_wave"]
        missing_owner_errors = schema_contract_errors(
            missing_owner, row_schema, root_schema=schema, path="$.sentinel"
        )
        if clean_schema_errors:
            print(f"RED: clean schema sentinel unexpectedly failed: {clean_schema_errors}", file=sys.stderr)
            return 1
        if not any("missing required property 'owner_wave'" in error for error in missing_owner_errors):
            print("RED: schema sentinel did not reject missing owner_wave", file=sys.stderr)
            return 1
    print(
        "GREEN: root Python/frontend/E2E/component mappings and clean inventory/schema/"
        "physical-path sentinels passed; symlink/reparse, hardlink, file/tree/row-budget, "
        "duplicate, removal, misroute, missing-owner, coherent Git/worktree cross-file/HEAD-drift, "
        "and atomic publication race/crash/idempotence sentinels produced RED"
    )
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    actions = parser.add_mutually_exclusive_group(required=True)
    actions.add_argument("--generate", action="store_true", help="write deterministic inventory.json")
    actions.add_argument("--check", action="store_true", help="validate inventory.json against current inputs")
    actions.add_argument("--self-test", action="store_true", help="run safe in-memory/temp sentinel validation")
    args = parser.parse_args()
    if args.self_test:
        return run_self_test()
    if args.check:
        return run_check()
    data, pinned_head, current_paths = build_inventory_coherent()
    errors = validate_inventory(
        data,
        live_source_paths=current_paths,
        current_revision=pinned_head,
    )
    if errors:
        for error in errors:
            print(f"RED: {error}", file=sys.stderr)
        return 1
    payload = render(data)
    assert_head_unchanged(pinned_head)
    write_inventory_atomically(payload)
    assert_head_unchanged(pinned_head)
    print(f"WROTE {rel(OUTPUT_PATH)} sha256={sha256_bytes(payload)} rows={len(data['rows'])}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
