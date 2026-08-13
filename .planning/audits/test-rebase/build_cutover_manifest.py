#!/usr/bin/env python3
"""Build the sealed Phase 129 cutover manifest without executing it.

The builder is deliberately stdlib-only.  Repository JSON is treated as data:
all three authority files are byte-sealed here before any selector, path or
command-shaped value is interpreted.  The default operation writes canonical
JSON to stdout.  ``--write`` accepts only the canonical manifest path and uses
a sibling O_EXCL staging file followed by atomic Linux no-replace rename.
"""

from __future__ import annotations

import argparse
import copy
import ctypes
import errno
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import re
import secrets
import stat
import sys
import tempfile
import unicodedata
from typing import Any, Iterable


_SEALED_REPO_ROOT = os.environ.get("HUB_PHASE128_REPO_ROOT")
REPO_ROOT = (
    Path(_SEALED_REPO_ROOT)
    if _SEALED_REPO_ROOT is not None
    else Path(__file__).resolve().parents[3]
)
if not REPO_ROOT.is_absolute():
    raise RuntimeError("HUB_PHASE128_REPO_ROOT must be absolute")
OUTPUT_PATH = ".planning/audits/test-rebase/cutover-manifest.json"
MAX_BATCH_OPERATIONS = 56
EXPECTED_BATCH_COUNT = 173
EXPECTED_OPERATION_COUNT = 1120
MAX_FILE_BYTES = 64 * 1024 * 1024
MAX_PATH_COMPONENTS = 64
MAX_COMPONENT_UTF8_BYTES = 255
MAX_PATH_UTF8_BYTES = 4096
PROTECTED_PREFIX = "pro-qa-ai/"
SHA256_RE = re.compile(r"^[0-9a-f]{64}$")
GIT_SHA_RE = re.compile(r"^[0-9a-f]{40}(?:[0-9a-f]{24})?$")
ASCII_SLUG_RE = re.compile(r"^[a-z0-9]+(?:-[a-z0-9]+)*$")
FILE_ATTRIBUTE_REPARSE_POINT = getattr(stat, "FILE_ATTRIBUTE_REPARSE_POINT", 0x400)
WINDOWS_RESERVED_NAMES = {
    "CON",
    "PRN",
    "AUX",
    "NUL",
    *(f"COM{number}" for number in range(1, 10)),
    *(f"LPT{number}" for number in range(1, 10)),
    *(f"COM{number}" for number in ("¹", "²", "³")),
    *(f"LPT{number}" for number in ("¹", "²", "³")),
}

AUTHORITY_FILES = {
    "inventory": {
        "path": ".planning/audits/test-rebase/inventory.json",
        "sha256": "83b6a31149807d21441a2ede91dec0b36b4a4396c4c0ceb58308b58878f5ad89",
    },
    "founder_manifest": {
        "path": ".planning/audits/test-rebase/founder-greenfield-cutover-manifest.json",
        "sha256": "292b40dbddcc5e5709eef1504a1d6331e5cc45ddeb6e1534642d3ca3dc4288b7",
    },
    "architecture_contract": {
        "path": ".planning/audits/test-rebase/test-architecture-contract.json",
        "sha256": "8e9f4819f4f9813c2b9c92ec7cae8dba47782c61a2414dda649aa7ac0f3d315e",
    },
}
LEDGER_REVISION = "b9f690da5cdcf322fc5b28a8d96c71a422407154"
NON_LEDGER_PATHS_SHA256 = "9b06611c2e5f9342ae081b6917e4033808c642e0b65216d7e60a0b2f89a62d49"
PHASE129_BATCH_ORDER_SHA256 = "faebb25e84d4efb01fc02fb515118ed28d5224be50dc02ec45da3baea42bb9f5"
PHASE129_PLAN_ORDER = ("129-02", "129-03", "129-04", "129-05", "129-06", "129-07")
PHASE129_PLAN_KINDS = {
    "129-02": "go-hook-free",
    "129-03": "frontend-unit",
    "129-04": "go-component-delete",
    "129-06": "go-hook-blocked",
    "129-07": "zero-consumer",
}
PHASE129_TAIL_DOMAINS = {
    ("129-03", "frontend-unit"): ("support",),
    ("129-04", "go-component-delete"): ("backend-support",),
}
PHASE129_05_FINAL_SUPPORT_DOMAINS = (
    "support",
    "frontend",
    "root-scripts",
    "pipeline",
    "legacy-stages",
    "legacy-docs",
)
PHASE129_PLAN_TASKS = {
    "129-02": ("TASK-197",),
    "129-03": ("TASK-198",),
    "129-04": ("TASK-219",),
    "129-05": ("TASK-219",),
    "129-06": ("TASK-234",),
    "129-07": ("TASK-219", "TASK-234"),
}
PHASE129_BATCH_TASK_OVERRIDES = {
    "129-05-support-root-scripts": ("TASK-219", "TASK-201"),
    "129-05-support-pipeline": ("TASK-219", "TASK-201"),
    "129-05-support-legacy-stages": ("TASK-219", "TASK-201"),
    "129-05-support-legacy-docs": ("TASK-219", "TASK-201"),
    "129-07-zero-consumer-ws": ("TASK-219", "TASK-234", "TASK-201"),
}

GROUP_SPECS = [
    {
        "id": "MOVE-GO-UNIT-HOOK-FREE",
        "founder_action": "move",
        "operation_action": "move",
        "selector": {"runner": "go-test", "target_class": "go-unit", "hook_consumers_count": 0},
        "count": 350,
        "tests": 1264,
        "paths_sha256": "af256cc0c5ad59fe6671641cea5ac138b2e6a4796c23b64aa038ec216ca40ff7",
        "plan": "129-02",
        "wave": 2,
        "task": ["TASK-197"],
        "id_kind": "go-hook-free",
        "owner": "be1",
        "verifiers": ["be-ver"],
        "gates": ["architecture-contract", "backend-unit"],
    },
    {
        "id": "MOVE-FRONTEND-UNIT",
        "founder_action": "move",
        "operation_action": "move",
        "selector": {"runner": "vitest", "target_class": "frontend-unit"},
        "count": 156,
        "tests": 409,
        "paths_sha256": "96fd1cacc09c39a06e894600de9168c5d9a3cf1971ec63f5bfe9c01d978b2b24",
        "plan": "129-03",
        "wave": 3,
        "task": ["TASK-198"],
        "id_kind": "frontend-unit",
        "owner": "fe1",
        "verifiers": ["fe-ver"],
        "gates": ["architecture-contract", "frontend-unit"],
    },
    {
        "id": "MOVE-FRONTEND-RENDER-SUPPORT",
        "founder_action": "move",
        "operation_action": "move",
        "selector": {
            "source_path": "testframework/stages/unit/frontend-test/render.tsx",
            "runner": "vitest-support",
            "target_class": "frontend-support",
        },
        "count": 1,
        "tests": 0,
        "paths_sha256": "a49f50bbf0b6667a10b0392169fb760311defe06288f9a3775f9d23ebdfe5a23",
        "plan": "129-03",
        "wave": 3,
        "task": ["TASK-198"],
        "id_kind": "frontend-unit",
        "owner": "fe1",
        "verifiers": ["fe-ver"],
        "gates": ["architecture-contract", "frontend-unit"],
    },
    {
        "id": "DELETE-GO-COMPONENT",
        "founder_action": "delete",
        "operation_action": "delete",
        "selector": {"runner": "go-test", "target_class_in": ["python-api", "python-integration"]},
        "count": 237,
        "tests": 982,
        "paths_sha256": "45bc86bd5be985e48d9b599200aa29641e02eb4947c97cb5ff0eb3f49d941e18",
        "plan": "129-04",
        "wave": 4,
        "task": ["TASK-219"],
        "id_kind": "go-component-delete",
        "owner": "devops-engineer",
        "verifiers": ["code-reviewer", "qa-auto-ver"],
        "gates": ["architecture-contract", "backend-unit"],
    },
    {
        "id": "DELETE-FRONTEND-STATIC-CONTRACT",
        "founder_action": "delete",
        "operation_action": "delete",
        "selector": {"runner": "vitest", "target_class": "python-static-contract"},
        "count": 24,
        "tests": 110,
        "paths_sha256": "4952f7970784c22570a263c31836a2e08a31a12592cccbb78e2b696409ae92b9",
        "plan": "129-05",
        "wave": 5,
        "task": ["TASK-219"],
        "id_kind": "static",
        "owner": "devops-engineer",
        "verifiers": ["code-reviewer", "qa-auto-ver"],
        "gates": ["architecture-contract", "frontend-unit", "python-scaffold"],
    },
    {
        "id": "DELETE-FRONTEND-COMPILE-FIXTURES",
        "founder_action": "delete",
        "operation_action": "delete",
        "selector": {
            "source_path_in": [
                "testframework/stages/unit/frontend-test/domains/platform/updateBase-version-required.fixture.ts",
                "testframework/stages/unit/frontend-test/domains/tracker/projects-sprint-version-required.fixture.ts",
            ],
            "runner": "vitest-support",
        },
        "count": 2,
        "tests": 0,
        "paths_sha256": "dd0084eb74d5e65fc574849c051c30bedb717379cb215b6ec8c10674f28ea956",
        "plan": "129-05",
        "wave": 5,
        "task": ["TASK-219"],
        "id_kind": "support",
        "owner": "devops-engineer",
        "verifiers": ["code-reviewer", "qa-auto-ver"],
        "gates": ["architecture-contract", "frontend-unit", "python-scaffold"],
    },
    {
        "id": "DELETE-PLAYWRIGHT-TS-SPECS",
        "founder_action": "delete",
        "operation_action": "delete",
        "selector": {"runner": "playwright-ts", "target_class": "python-e2e"},
        "count": 20,
        "tests": 53,
        "paths_sha256": "aa521c92eca8744fe41e22d2b278d12d0f9562e08936938a34a6f54c70c0a2fb",
        "plan": "129-05",
        "wave": 5,
        "task": ["TASK-219"],
        "id_kind": "e2e",
        "owner": "devops-engineer",
        "verifiers": ["code-reviewer", "qa-auto-ver"],
        "gates": ["architecture-contract", "frontend-unit", "python-scaffold"],
    },
    {
        "id": "DELETE-PLAYWRIGHT-TS-SUPPORT",
        "founder_action": "delete",
        "operation_action": "delete",
        "selector": {"runner": "playwright-support-ts", "target_class": "python-e2e-support"},
        "count": 14,
        "tests": 0,
        "paths_sha256": "ec40b12fce301f951d299a1e32a651eab73cd4b3cac621d2a1c65aaa6b4e4eef",
        "plan": "129-05",
        "wave": 5,
        "task": ["TASK-219"],
        "id_kind": "support",
        "owner": "devops-engineer",
        "verifiers": ["code-reviewer", "qa-auto-ver"],
        "gates": ["architecture-contract", "frontend-unit", "python-scaffold"],
    },
    {
        "id": "KEEP-GO-UNIT-HOOK-DEPENDENT",
        "founder_action": "keep",
        "operation_action": "keep-blocked",
        "selector": {"runner": "go-test", "target_class": "go-unit", "hook_consumers_count": ">0"},
        "count": 245,
        "tests": 1428,
        "paths_sha256": "86e55748e6c5986ec4eef370b51fb3f02c7fca7ec24fa8eb1e85d1913ad33924",
        "plan": "129-06",
        "wave": 6,
        "task": ["TASK-234"],
        "id_kind": "go-hook-blocked",
        "owner": "be1",
        "verifiers": ["be-ver"],
        "gates": ["architecture-contract", "backend-unit"],
    },
]

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

GATE_REGISTRY = [
    {
        "id": "architecture-contract",
        "image": "python:3.12-slim",
        "expected_image_digest": None,
        "cpus": "1.0",
        "memory": "1g",
        "memory_swap": "1g",
        "pids_limit": 128,
        "wall_timeout_seconds": 300,
        "kill_grace_seconds": 10,
        "network": "none",
        "read_only_repo": True,
        "remove_container": True,
        "task_name_template": "hub-129-architecture-contract-{task_run}",
        "required_labels": {"hub.task-run": "{task_run}", "hub.owner": "phase129"},
        "outer_cleanup": "finally-exact-name-and-label",
        "entrypoints": [
            "python3 .planning/audits/test-rebase/validate_cutover_manifest.py --check .planning/audits/test-rebase/cutover-manifest.json",
            "python3 .planning/audits/test-rebase/validate_cutover_manifest.py --stdin --self-test",
        ],
    },
    {
        "id": "backend-unit",
        "image": "golang:1.25-bookworm",
        "expected_image_digest": None,
        "cpus": "2.0",
        "memory": "4g",
        "memory_swap": "4g",
        "pids_limit": 256,
        "wall_timeout_seconds": 720,
        "kill_grace_seconds": 10,
        "network": "none",
        "read_only_repo": True,
        "remove_container": True,
        "task_name_template": "hub-129-backend-unit-{task_run}",
        "required_labels": {"hub.task-run": "{task_run}", "hub.owner": "phase129"},
        "outer_cleanup": "finally-exact-name-and-label",
        "entrypoints": [
            "go build -tags release ./...",
            "go vet ./...",
            "go test -race -count=1 -timeout=600s ./tests/...",
        ],
    },
    {
        "id": "frontend-unit",
        "image": "mcr.microsoft.com/playwright:v1.61.0-noble",
        "expected_image_digest": None,
        "cpus": "2.0",
        "memory": "4g",
        "memory_swap": "4g",
        "pids_limit": 256,
        "wall_timeout_seconds": 1200,
        "kill_grace_seconds": 10,
        "network": "none",
        "read_only_repo": True,
        "remove_container": True,
        "task_name_template": "hub-129-frontend-unit-{task_run}",
        "required_labels": {"hub.task-run": "{task_run}", "hub.owner": "phase129"},
        "outer_cleanup": "finally-exact-name-and-label",
        "entrypoints": [
            "npx tsc --noEmit",
            "npm run lint",
            "npx vitest run tests",
            "npm run build",
        ],
    },
    {
        "id": "python-scaffold",
        "image": "hub-testframework:phase129",
        "expected_image_digest": None,
        "cpus": "2.0",
        "memory": "4g",
        "memory_swap": "4g",
        "pids_limit": 256,
        "wall_timeout_seconds": 600,
        "kill_grace_seconds": 10,
        "network": "none",
        "read_only_repo": True,
        "remove_container": True,
        "task_name_template": "hub-129-python-scaffold-{task_run}",
        "required_labels": {"hub.task-run": "{task_run}", "hub.owner": "phase129"},
        "outer_cleanup": "finally-exact-name-and-label",
        "entrypoints": (
            [f"test -f {path}" for path in PYTHON_SCAFFOLD_REQUIRED_FILES]
            + [f"test ! -e {path}" for path in PYTHON_SCAFFOLD_FORBIDDEN_PATHS]
            + [
                "python3 -m pytest --collect-only -q testframework/selftests testframework/tests/api",
                "python3 -m pytest -q -m scaffold testframework/selftests",
            ]
        ),
    },
]


class ManifestError(ValueError):
    """A concise, non-secret validation failure."""


class MissingPathError(ManifestError):
    """A proven ENOENT; only this error permits Git packed-ref fallback."""


def rename_noreplace(source_fd: int, source: str, target_fd: int, target: str) -> None:
    """Linux descriptor-relative atomic rename that refuses replacement."""

    renameat2 = getattr(ctypes.CDLL(None, use_errno=True), "renameat2", None)
    if renameat2 is None:
        raise ManifestError("renameat2(RENAME_NOREPLACE) is unavailable")
    renameat2.argtypes = [ctypes.c_int, ctypes.c_char_p, ctypes.c_int, ctypes.c_char_p, ctypes.c_uint]
    renameat2.restype = ctypes.c_int
    if renameat2(source_fd, os.fsencode(source), target_fd, os.fsencode(target), 1) == 0:
        return
    number = ctypes.get_errno()
    if number in {errno.ENOSYS, errno.EINVAL, errno.ENOTSUP}:
        raise ManifestError("renameat2(RENAME_NOREPLACE) is unsupported")
    if number in {errno.EEXIST, errno.ENOTEMPTY}:
        raise ManifestError(f"atomic no-replace destination exists: {target}")
    raise OSError(number, os.strerror(number), target)


def canonical_path_errors(value: Any, label: str) -> list[str]:
    if not isinstance(value, str):
        return [f"{label} must be a string"]
    if not value:
        return [f"{label} is empty"]
    if "\\" in value:
        return [f"{label} contains a backslash"]
    if value.startswith("//") or value.startswith("/") or re.match(r"^[A-Za-z]:", value):
        return [f"{label} is absolute/UNC/drive-qualified"]
    if unicodedata.normalize("NFC", value) != value:
        return [f"{label} is not Unicode-NFC"]
    if any(ord(character) < 32 or ord(character) == 127 for character in value):
        return [f"{label} contains a control character"]
    parts = value.split("/")
    if any(part in {"", ".", ".."} for part in parts):
        return [f"{label} contains an empty/dot/traversal segment"]
    if len(parts) > MAX_PATH_COMPONENTS:
        return [f"{label} exceeds {MAX_PATH_COMPONENTS} component budget"]
    if len(value.encode("utf-8")) > MAX_PATH_UTF8_BYTES:
        return [f"{label} exceeds {MAX_PATH_UTF8_BYTES} UTF-8 byte budget"]
    for part in parts:
        if len(part.encode("utf-8")) > MAX_COMPONENT_UTF8_BYTES:
            return [f"{label} component exceeds {MAX_COMPONENT_UTF8_BYTES} UTF-8 byte budget"]
        if part.rstrip(" .") != part:
            return [f"{label} contains a trailing dot/space alias"]
        if ":" in part:
            return [f"{label} contains a colon/ADS segment"]
        if part.split(".", 1)[0].upper() in WINDOWS_RESERVED_NAMES:
            return [f"{label} contains a reserved Windows name"]
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


def portable_key(value: str) -> str:
    return "/".join(unicodedata.normalize("NFC", part).casefold().rstrip(" .") for part in value.split("/"))


def reject_portable_aliases(paths: Iterable[str], label: str) -> None:
    seen: dict[str, str] = {}
    for value in paths:
        require_canonical_path(value, label)
        key = portable_key(value)
        previous = seen.setdefault(key, value)
        if previous != value:
            raise ManifestError(f"{label} portable alias/case collision")


def is_symlink_or_reparse(metadata: os.stat_result) -> bool:
    return stat.S_ISLNK(metadata.st_mode) or bool(
        int(getattr(metadata, "st_file_attributes", 0)) & FILE_ATTRIBUTE_REPARSE_POINT
    )


def metadata_signature(metadata: os.stat_result, *, content: bool) -> tuple[int, ...]:
    result = (
        int(metadata.st_mode),
        int(metadata.st_dev),
        int(metadata.st_ino),
        int(getattr(metadata, "st_file_attributes", 0)),
    )
    if content:
        result += (
            int(metadata.st_size),
            int(getattr(metadata, "st_mtime_ns", int(metadata.st_mtime * 1_000_000_000))),
            int(getattr(metadata, "st_ctime_ns", int(metadata.st_ctime * 1_000_000_000))),
        )
    return result


def open_physical_root(path: Path, flags: int) -> tuple[int, tuple[int, ...]]:
    """Open an absolute repo root one component at a time without following aliases."""

    absolute = Path(os.path.abspath(path))
    if (
        len(absolute.parts) - 1 > MAX_PATH_COMPONENTS
        or len(str(absolute).encode("utf-8")) > MAX_PATH_UTF8_BYTES
        or any(len(part.encode("utf-8")) > MAX_COMPONENT_UTF8_BYTES for part in absolute.parts[1:])
    ):
        raise ManifestError("repository root exceeds component/path budget")
    protected = os.path.normcase(os.path.abspath(REPO_ROOT / PROTECTED_PREFIX.rstrip("/")))
    candidate = os.path.normcase(str(absolute))
    try:
        inside_protected = os.path.commonpath([candidate, protected]) == protected
    except ValueError:
        inside_protected = False
    if inside_protected:
        raise ManifestError("repository root enters protected pro-qa-ai")
    anchor = Path(absolute.anchor)
    descriptor = os.open(anchor, flags)
    try:
        for part in absolute.parts[1:]:
            before = os.stat(part, dir_fd=descriptor, follow_symlinks=False)
            if is_symlink_or_reparse(before) or not stat.S_ISDIR(before.st_mode):
                raise ManifestError("repository root ancestor is not a physical directory")
            next_fd = os.open(part, flags, dir_fd=descriptor)
            opened = os.fstat(next_fd)
            if metadata_signature(before, content=False) != metadata_signature(opened, content=False):
                os.close(next_fd)
                raise ManifestError("repository root ancestor changed while opening")
            os.close(descriptor)
            descriptor = next_fd
        return descriptor, metadata_signature(os.fstat(descriptor), content=False)
    except Exception:
        os.close(descriptor)
        raise


class SafeRepo:
    """Bounded no-follow reads plus final ancestor/absence TOCTOU checks."""

    def __init__(self, root: Path):
        self.root = Path(os.path.abspath(root))
        self._directories: dict[Path, tuple[int, ...]] = {}
        self._absent_relatives: set[str] = set()
        self._identities: dict[tuple[int, int], str] = {}
        self._cache: dict[str, bytes] = {}
        self._file_signatures: dict[str, tuple[tuple[int, ...], ...]] = {}
        directory_flags, _ = self._descriptor_flags()
        root_fd, root_signature = open_physical_root(self.root, directory_flags)
        os.close(root_fd)
        self._directories[self.root] = root_signature

    @staticmethod
    def _descriptor_flags() -> tuple[int, int]:
        required = ("O_CLOEXEC", "O_DIRECTORY", "O_NOFOLLOW", "O_PATH")
        missing = [name for name in required if not hasattr(os, name)]
        if os.name != "posix" or missing or os.open not in getattr(os, "supports_dir_fd", set()):
            raise ManifestError(
                "descriptor-relative O_NOFOLLOW traversal unavailable; "
                f"filesystem authority fails closed (os={os.name} missing={missing})"
            )
        directory = os.O_PATH | os.O_CLOEXEC | os.O_DIRECTORY | os.O_NOFOLLOW
        leaf = os.O_RDONLY | os.O_CLOEXEC | os.O_NOFOLLOW | getattr(os, "O_NONBLOCK", 0)
        return directory, leaf

    def _path(self, relative: str) -> Path:
        require_canonical_path(relative, "repository path")
        return self.root.joinpath(*relative.split("/"))

    def _read_once(self, relative: str, max_bytes: int) -> tuple[bytes, tuple[tuple[int, ...], ...]]:
        require_canonical_path(relative, "repository path")
        directory_flags, leaf_flags = self._descriptor_flags()
        descriptors: list[int] = []
        signatures: list[tuple[int, ...]] = []
        try:
            root_fd, root_signature = open_physical_root(self.root, directory_flags)
            descriptors.append(root_fd)
            root_opened = os.fstat(root_fd)
            if root_signature != self._directories[self.root]:
                raise ManifestError("repository root identity changed")
            if root_signature != metadata_signature(root_opened, content=False):
                raise ManifestError("repository root changed while opening")
            signatures.append(metadata_signature(root_opened, content=False))
            for index, part in enumerate(relative.split("/")):
                leaf = index == len(relative.split("/")) - 1
                try:
                    before = os.stat(part, dir_fd=descriptors[-1], follow_symlinks=False)
                except FileNotFoundError:
                    raise MissingPathError(f"required path is missing: {relative}") from None
                except OSError as error:
                    raise ManifestError(f"metadata unavailable for {relative}: {error.__class__.__name__}") from None
                if is_symlink_or_reparse(before):
                    raise ManifestError(f"symlink/reparse component rejected: {relative}")
                try:
                    descriptor = os.open(part, leaf_flags if leaf else directory_flags, dir_fd=descriptors[-1])
                except OSError as error:
                    raise ManifestError(f"safe descriptor open failed for {relative}: {error.__class__.__name__}") from None
                descriptors.append(descriptor)
                opened = os.fstat(descriptor)
                include_content = leaf
                if metadata_signature(before, content=include_content) != metadata_signature(opened, content=include_content):
                    raise ManifestError(f"file identity changed before read (TOCTOU): {relative}")
                if leaf:
                    if not stat.S_ISREG(opened.st_mode) or int(opened.st_nlink) != 1:
                        raise ManifestError(f"single-link regular file required: {relative}")
                    if int(opened.st_size) > max_bytes:
                        raise ManifestError(f"file exceeds bounded read limit: {relative}")
                elif not stat.S_ISDIR(opened.st_mode):
                    raise ManifestError(f"non-directory ancestor rejected: {relative}")
                signatures.append(metadata_signature(opened, content=include_content))
            chunks: list[bytes] = []
            size = 0
            while True:
                chunk = os.read(descriptors[-1], min(1024 * 1024, max_bytes + 1 - size))
                if not chunk:
                    break
                chunks.append(chunk)
                size += len(chunk)
                if size > max_bytes:
                    raise ManifestError(f"file grew beyond bounded read limit: {relative}")
            opened_after = os.fstat(descriptors[-1])
            if signatures[-1] != metadata_signature(opened_after, content=True):
                raise ManifestError(f"file changed during read (TOCTOU): {relative}")
            return b"".join(chunks), tuple(signatures)
        finally:
            for descriptor in reversed(descriptors):
                try:
                    os.close(descriptor)
                except OSError:
                    pass

    def read(self, relative: str, *, max_bytes: int = MAX_FILE_BYTES) -> bytes:
        if relative in self._cache:
            return self._cache[relative]
        payload, first = self._read_once(relative, max_bytes)
        confirmation, second = self._read_once(relative, max_bytes)
        if first != second or payload != confirmation:
            raise ManifestError(f"file/ancestor changed during independent recheck: {relative}")
        leaf = second[-1]
        identity = (leaf[1], leaf[2])
        previous = self._identities.setdefault(identity, relative)
        if previous != relative:
            raise ManifestError(f"filesystem alias/hardlink collision: {previous} and {relative}")
        self._cache[relative] = payload
        self._file_signatures[relative] = second
        return payload

    def sha256(self, relative: str) -> str:
        return hashlib.sha256(self.read(relative)).hexdigest()

    def assert_absent(self, relative: str) -> None:
        require_canonical_path(relative, "repository path")
        directory_flags, _ = self._descriptor_flags()

        def inspect() -> tuple[str | None, tuple[tuple[int, ...], ...]]:
            descriptors: list[int] = []
            signatures: list[tuple[int, ...]] = []
            current = self.root
            try:
                root_fd, root_signature = open_physical_root(self.root, directory_flags)
                descriptors.append(root_fd)
                root_meta = os.fstat(root_fd)
                if root_signature != self._directories[self.root] or metadata_signature(root_meta, content=False) != root_signature:
                    raise ManifestError("repository root identity changed during destination check")
                signatures.append(metadata_signature(root_meta, content=False))
                for part in relative.split("/"):
                    current = current / part
                    try:
                        before = os.stat(part, dir_fd=descriptors[-1], follow_symlinks=False)
                    except FileNotFoundError:
                        return str(current), tuple(signatures)
                    if is_symlink_or_reparse(before):
                        raise ManifestError(f"symlink/reparse destination component rejected: {relative}")
                    if not stat.S_ISDIR(before.st_mode):
                        return None, tuple(signatures + [metadata_signature(before, content=True)])
                    descriptor = os.open(part, directory_flags, dir_fd=descriptors[-1])
                    descriptors.append(descriptor)
                    opened = os.fstat(descriptor)
                    if metadata_signature(before, content=False) != metadata_signature(opened, content=False):
                        raise ManifestError(f"destination ancestor changed (TOCTOU): {relative}")
                    signatures.append(metadata_signature(opened, content=False))
                return None, tuple(signatures)
            finally:
                for descriptor in reversed(descriptors):
                    try:
                        os.close(descriptor)
                    except OSError:
                        pass

        missing_before, signature_before = inspect()
        missing_after, signature_after = inspect()
        if missing_before is None or missing_after is None:
            raise ManifestError(f"existing move destination rejected: {relative}")
        if missing_before != missing_after or signature_before != signature_after:
            raise ManifestError(f"destination changed during absence check (TOCTOU): {relative}")
        self._absent_relatives.add(relative)

    def open_directory(self, relative: str) -> tuple[int, tuple[int, ...]]:
        """Return a held no-follow descriptor for an existing repository directory."""

        require_canonical_path(relative, "repository directory")
        self._descriptor_flags()
        directory_flags = os.O_RDONLY | os.O_CLOEXEC | os.O_DIRECTORY | os.O_NOFOLLOW
        descriptors: list[int] = []
        try:
            root_fd, root_signature = open_physical_root(self.root, directory_flags)
            descriptors.append(root_fd)
            if root_signature != self._directories[self.root] or metadata_signature(os.fstat(root_fd), content=False) != root_signature:
                raise ManifestError("repository root identity changed while opening directory")
            for part in relative.split("/"):
                before = os.stat(part, dir_fd=descriptors[-1], follow_symlinks=False)
                if is_symlink_or_reparse(before) or not stat.S_ISDIR(before.st_mode):
                    raise ManifestError(f"physical directory component required: {relative}")
                descriptor = os.open(part, directory_flags, dir_fd=descriptors[-1])
                descriptors.append(descriptor)
                opened = os.fstat(descriptor)
                if metadata_signature(before, content=False) != metadata_signature(opened, content=False):
                    raise ManifestError(f"directory changed while opening (TOCTOU): {relative}")
            result = descriptors.pop()
            meta = os.fstat(result)
            return result, metadata_signature(meta, content=False)
        finally:
            for descriptor in reversed(descriptors):
                try:
                    os.close(descriptor)
                except OSError:
                    pass

    def finalize(self) -> None:
        for relative, before in self._file_signatures.items():
            _, after = self._read_once(relative, MAX_FILE_BYTES)
            if before != after:
                raise ManifestError(f"file identity changed during validation (TOCTOU): {relative}")
        directory_flags, _ = self._descriptor_flags()
        for directory, before in self._directories.items():
            try:
                descriptor, after = open_physical_root(directory, directory_flags)
                os.close(descriptor)
            except OSError as error:
                raise ManifestError(f"directory changed during validation: {error.__class__.__name__}") from None
            if before != after:
                raise ManifestError("directory identity changed during validation (TOCTOU)")
        for relative in tuple(self._absent_relatives):
            self.assert_absent(relative)


def reject_duplicate_json_object(pairs: list[tuple[str, Any]]) -> dict[str, Any]:
    result: dict[str, Any] = {}
    for key, value in pairs:
        if key in result:
            raise ManifestError(f"duplicate JSON key rejected: {key!r}")
        result[key] = value
    return result


def json_load(payload: bytes, label: str) -> dict[str, Any]:
    try:
        value = json.loads(
            payload.decode("utf-8"), object_pairs_hook=reject_duplicate_json_object
        )
    except (UnicodeDecodeError, json.JSONDecodeError) as error:
        raise ManifestError(f"invalid UTF-8 JSON authority: {label}") from error
    if not isinstance(value, dict):
        raise ManifestError(f"JSON authority must be an object: {label}")
    return value


def canonical_json_bytes(value: Any, *, pretty: bool = False) -> bytes:
    if pretty:
        text = json.dumps(value, ensure_ascii=False, indent=2, sort_keys=True) + "\n"
    else:
        text = json.dumps(value, ensure_ascii=False, separators=(",", ":"), sort_keys=True)
    return text.encode("utf-8")


def object_digest(value: Any) -> str:
    return hashlib.sha256(canonical_json_bytes(value)).hexdigest()


def path_list_digest(paths: Iterable[str]) -> str:
    return hashlib.sha256(("\n".join(sorted(paths)) + "\n").encode("utf-8")).hexdigest()


def row_matches(row: dict[str, Any], selector: dict[str, Any]) -> bool:
    for key, expected in selector.items():
        if key.endswith("_in"):
            if row.get(key[:-3]) not in expected:
                return False
        elif key == "hook_consumers_count":
            consumers = row.get("hook_consumers")
            if not isinstance(consumers, list):
                return False
            count = len(consumers)
            if expected == ">0":
                if count <= 0:
                    return False
            elif count != expected:
                return False
        elif row.get(key) != expected:
            return False
    return True


def slug(value: str) -> str:
    normalized = unicodedata.normalize("NFKD", value).encode("ascii", "ignore").decode("ascii").lower()
    result = re.sub(r"[^a-z0-9]+", "-", normalized).strip("-")
    if not result or ASCII_SLUG_RE.fullmatch(result) is None:
        raise ManifestError(f"domain cannot form a stable ASCII slug: {value!r}")
    return result


def read_git_head(safe: SafeRepo) -> str:
    head = safe.read(".git/HEAD", max_bytes=4096).decode("ascii", "strict").strip()
    if head.startswith("ref: "):
        reference = head[5:]
        require_canonical_path(f".git/{reference}", "git HEAD reference")
        try:
            value = safe.read(f".git/{reference}", max_bytes=4096).decode("ascii", "strict").strip()
        except MissingPathError as direct_error:
            packed = safe.read(".git/packed-refs", max_bytes=8 * 1024 * 1024).decode("ascii", "strict")
            matches = [line.split(" ", 1)[0] for line in packed.splitlines() if line.endswith(f" {reference}")]
            if len(matches) != 1:
                raise direct_error
            value = matches[0]
    else:
        value = head
    if GIT_SHA_RE.fullmatch(value) is None:
        raise ManifestError("repository HEAD is not a canonical Git object ID")
    return value


def load_authorities(safe: SafeRepo) -> tuple[dict[str, Any], dict[str, Any], dict[str, Any]]:
    loaded: dict[str, dict[str, Any]] = {}
    for name, spec in AUTHORITY_FILES.items():
        payload = safe.read(spec["path"], max_bytes=8 * 1024 * 1024)
        actual = hashlib.sha256(payload).hexdigest()
        if actual != spec["sha256"]:
            raise ManifestError(f"sealed authority digest mismatch: {name}")
        loaded[name] = json_load(payload, name)
    return loaded["inventory"], loaded["founder_manifest"], loaded["architecture_contract"]


def validate_authority_contracts(
    inventory: dict[str, Any], founder: dict[str, Any], contract: dict[str, Any]
) -> dict[str, list[dict[str, Any]]]:
    rows = inventory.get("rows")
    if inventory.get("revision") != LEDGER_REVISION or not isinstance(rows, list) or len(rows) != 1049:
        raise ManifestError("sealed inventory revision/row count changed")
    source_ledger = founder.get("source_ledger")
    if not isinstance(source_ledger, dict) or source_ledger.get("sha256") != AUTHORITY_FILES["inventory"]["sha256"]:
        raise ManifestError("founder manifest does not pin the sealed inventory")
    if founder.get("status") != "verified-hook-blocked":
        raise ManifestError("founder manifest status changed")
    decision = contract.get("decision")
    if not isinstance(decision, dict) or decision.get("migration_mode") != "greenfield-rebuild":
        raise ManifestError("architecture contract migration mode changed")
    if decision.get("row_parity_required") is not False or decision.get("dual_run_required") is not False:
        raise ManifestError("architecture contract reintroduced parity/dual-run")
    snapshot = contract.get("cutover_snapshot")
    if not isinstance(snapshot, dict) or snapshot.get("ledger_sha256") != AUTHORITY_FILES["inventory"]["sha256"]:
        raise ManifestError("architecture contract cutover snapshot changed")

    founder_partitions = founder.get("row_partitions")
    contract_groups = snapshot.get("groups")
    if not isinstance(founder_partitions, list) or not isinstance(contract_groups, list):
        raise ManifestError("authority group arrays are missing")
    founder_by_id = {item.get("id"): item for item in founder_partitions if isinstance(item, dict)}
    contract_by_id = {item.get("id"): item for item in contract_groups if isinstance(item, dict)}
    if set(founder_by_id) != {spec["id"] for spec in GROUP_SPECS} or set(contract_by_id) != set(founder_by_id):
        raise ManifestError("authority group IDs changed")

    membership: dict[str, list[dict[str, Any]]] = {spec["id"]: [] for spec in GROUP_SPECS}
    source_paths: list[str] = []
    total_tests = 0
    for index, row in enumerate(rows):
        if not isinstance(row, dict):
            raise ManifestError(f"inventory row {index} is not an object")
        source = require_canonical_path(row.get("source_path"), f"inventory row {index} source")
        if not SHA256_RE.fullmatch(str(row.get("sha256", ""))):
            raise ManifestError(f"inventory row source digest invalid: {source}")
        matches = [spec for spec in GROUP_SPECS if row_matches(row, spec["selector"])]
        if len(matches) != 1:
            raise ManifestError(f"inventory row must match exactly one group: {source}")
        membership[matches[0]["id"]].append(row)
        source_paths.append(source)
        total_tests += int(row.get("test_count", -1))
    if len(set(source_paths)) != 1049 or total_tests != 4246:
        raise ManifestError("inventory partition is not unique/exhaustive")
    reject_portable_aliases(source_paths, "ledger source")

    for spec in GROUP_SPECS:
        selected = membership[spec["id"]]
        paths = [str(row["source_path"]) for row in selected]
        tests = sum(int(row.get("test_count", -1)) for row in selected)
        if len(selected) != spec["count"] or tests != spec["tests"] or path_list_digest(paths) != spec["paths_sha256"]:
            raise ManifestError(f"sealed partition mismatch: {spec['id']}")
        founder_item = founder_by_id[spec["id"]]
        contract_item = contract_by_id[spec["id"]]
        if (
            founder_item.get("action") != spec["founder_action"]
            or founder_item.get("selector") != spec["selector"]
            or founder_item.get("source_rows") != spec["count"]
            or founder_item.get("declared_tests") != spec["tests"]
        ):
            raise ManifestError(f"founder partition declaration mismatch: {spec['id']}")
        if (
            contract_item.get("selector") != spec["selector"]
            or contract_item.get("expected_count") != spec["count"]
            or contract_item.get("expected_declared_tests") != spec["tests"]
            or contract_item.get("source_paths_sha256") != spec["paths_sha256"]
        ):
            raise ManifestError(f"contract partition declaration mismatch: {spec['id']}")

    blocked = membership["KEEP-GO-UNIT-HOOK-DEPENDENT"]
    links = [link for row in blocked for link in row.get("hook_consumers", [])]
    hook_files = {link.get("hook_source") for link in links if isinstance(link, dict)}
    hook_symbols = {
        f"{link.get('hook_source')}::{link.get('symbol')}" for link in links if isinstance(link, dict)
    }
    if (len(links), len(hook_files), len(hook_symbols)) != (653, 42, 434):
        raise ManifestError("hook-blocked link seal changed")

    nonledger = founder.get("non_ledger_delete_paths")
    if not isinstance(nonledger, list) or len(nonledger) != 28 or len(set(nonledger)) != 28:
        raise ManifestError("non-ledger closure count/uniqueness changed")
    for index, path in enumerate(nonledger):
        require_canonical_path(path, f"non-ledger path {index}")
    if path_list_digest(nonledger) != NON_LEDGER_PATHS_SHA256 or set(nonledger).intersection(source_paths):
        raise ManifestError("non-ledger closure digest/overlap changed")
    return membership


def operation_from_row(row: dict[str, Any], spec: dict[str, Any], safe: SafeRepo, *, verify_fs: bool) -> dict[str, Any]:
    source = str(row["source_path"])
    source_sha = str(row["sha256"])
    if verify_fs and safe.sha256(source) != source_sha:
        raise ManifestError(f"source content digest mismatch: {source}")
    operation: dict[str, Any] = {
        "action": spec["operation_action"],
        "authority_group": spec["id"],
        "ledger_source": True,
        "source": source,
        "source_sha256": source_sha,
    }
    if spec["operation_action"] == "move":
        target = require_canonical_path(row.get("target_path"), f"move target for {source}")
        operation["target"] = target
        if verify_fs:
            safe.assert_absent(target)
    elif spec["operation_action"] == "keep-blocked":
        links = row.get("hook_consumers")
        if not isinstance(links, list) or not links:
            raise ManifestError(f"hook-blocked row lost exact hook links: {source}")
        normalized_links = []
        for link in links:
            if not isinstance(link, dict):
                raise ManifestError(f"malformed hook link: {source}")
            hook_source = require_canonical_path(link.get("hook_source"), f"hook source for {source}")
            symbol = link.get("symbol")
            if not isinstance(symbol, str) or re.fullmatch(r"Export[A-Za-z0-9_]*", symbol) is None:
                raise ManifestError(f"malformed hook symbol: {source}")
            normalized_links.append({"hook_source": hook_source, "symbol": symbol})
        operation["hook_links"] = sorted(normalized_links, key=lambda item: (item["hook_source"], item["symbol"]))
        operation["release_condition"] = "regenerated-zero-link-authority-and-docker-green"
    return operation


def nonledger_domain(path: str) -> str:
    if path.startswith("backend/testframework/"):
        return "backend-support"
    if path.startswith("frontend/"):
        return "frontend"
    if path.startswith("scripts/"):
        return "root-scripts"
    if path.startswith("testframework/pipeline/"):
        return "pipeline"
    if path.startswith("testframework/stages/"):
        return "legacy-stages"
    if path.startswith("testframework/"):
        return "legacy-docs"
    raise ManifestError(f"non-ledger path has no stable domain: {path}")


def hook_guard_operations(inventory: dict[str, Any], safe: SafeRepo, *, verify_fs: bool) -> dict[str, list[dict[str, Any]]]:
    hooks = inventory.get("hooks")
    if not isinstance(hooks, list) or len(hooks) != 43:
        raise ManifestError("inventory hook file seal changed")
    by_domain: dict[str, list[dict[str, Any]]] = {}
    sources: list[str] = []
    for hook in hooks:
        if not isinstance(hook, dict):
            raise ManifestError("malformed hook inventory item")
        source = require_canonical_path(hook.get("source_path"), "hook guard source")
        source_sha = str(hook.get("sha256", ""))
        if SHA256_RE.fullmatch(source_sha) is None:
            raise ManifestError(f"hook source digest invalid: {source}")
        if verify_fs and safe.sha256(source) != source_sha:
            raise ManifestError(f"hook source content digest mismatch: {source}")
        links: list[dict[str, str]] = []
        symbols = hook.get("symbols")
        if not isinstance(symbols, list):
            raise ManifestError(f"hook symbol list malformed: {source}")
        for symbol in symbols:
            if not isinstance(symbol, dict) or not isinstance(symbol.get("name"), str):
                raise ManifestError(f"hook symbol malformed: {source}")
            consumers = symbol.get("consumer_test_files")
            if not isinstance(consumers, list):
                raise ManifestError(f"hook consumers malformed: {source}")
            for consumer in consumers:
                consumer_path = require_canonical_path(consumer, f"hook consumer for {source}")
                links.append({"consumer_source": consumer_path, "symbol": symbol["name"]})
        operation = {
            "action": "keep-blocked",
            "authority_group": "ZERO-CONSUMER-HOOK-GUARD",
            "ledger_source": False,
            "source": source,
            "source_sha256": source_sha,
            "consumer_links": sorted(links, key=lambda item: (item["consumer_source"], item["symbol"])),
            "release_condition": "regenerated-exact-zero-consumer-authority",
        }
        domain = slug(str(hook.get("module", "")))
        by_domain.setdefault(domain, []).append(operation)
        sources.append(source)
    if len(set(sources)) != 43:
        raise ManifestError("hook guard sources are not unique")
    return by_domain


def consumer_role(path: str, founder: dict[str, Any]) -> str:
    edit = founder["edit_in_place_paths"]
    if path in edit["runtime_and_ci"]:
        return "runtime-ci-config"
    if path in edit["root_testframework_docs_rewrite"]:
        return "testframework-doc"
    if path in edit["current_project_docs_sync"]:
        return "current-project-doc"
    return "delete-operation"


def build_consumer_registry(founder: dict[str, Any], safe: SafeRepo, *, verify_fs: bool) -> list[dict[str, Any]]:
    edit = founder.get("edit_in_place_paths")
    if not isinstance(edit, dict):
        raise ManifestError("founder edit-in-place registry missing")
    paths: list[str] = []
    for key in ("runtime_and_ci", "root_testframework_docs_rewrite", "current_project_docs_sync"):
        values = edit.get(key)
        if not isinstance(values, list):
            raise ManifestError(f"founder consumer list missing: {key}")
        paths.extend(str(value) for value in values)
    delete_paths = founder.get("non_ledger_delete_paths")
    assert isinstance(delete_paths, list)
    paths.extend(str(value) for value in delete_paths)
    if len(paths) != 65 or len(set(paths)) != 65:
        raise ManifestError("shared consumer/delete registry must contain 65 unique existing paths")
    reject_portable_aliases(paths + [OUTPUT_PATH], "consumer registry path")
    registry: list[dict[str, Any]] = []
    for index, path in enumerate(sorted(paths)):
        require_canonical_path(path, f"consumer registry path {index}")
        before_sha = safe.sha256(path) if verify_fs else None
        registry.append(
            {
                "id": f"consumer-{index + 1:03d}",
                "path": path,
                "role": consumer_role(path, founder),
                "mode": "delete" if path in delete_paths else "edit",
                "before_sha256": before_sha,
            }
        )
    registry.append(
        {
            "id": "consumer-manifest-self",
            "path": OUTPUT_PATH,
            "role": "cutover-manifest-self",
            "mode": "self",
            "before_sha256": None,
        }
    )
    return registry


def affected_consumers(
    registry: list[dict[str, Any]], *, plan: str, operation_sources: list[str]
) -> list[str]:
    result: list[str] = []
    for item in registry:
        role = item["role"]
        mode = item["mode"]
        path = item["path"]
        include = role == "cutover-manifest-self"
        if mode == "delete" and path in operation_sources:
            include = True
        if mode == "edit":
            if role in {"current-project-doc", "testframework-doc"}:
                include = True
            if role == "runtime-ci-config":
                if path in {".github/workflows/ci.yml", ".gitlab-ci.yml"}:
                    include = True
                elif plan in {"129-03", "129-05"}:
                    include = True
        if include:
            result.append(item["id"])
    return result


def state_fields(registry_by_id: dict[str, dict[str, Any]], consumer_ids: list[str]) -> dict[str, Any]:
    transitions = []
    for consumer_id in consumer_ids:
        consumer = registry_by_id[consumer_id]
        if consumer["mode"] == "delete":
            continue
        transitions.append(
            {
                "consumer_id": consumer_id,
                "before_sha256": consumer["before_sha256"],
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


def chunked(values: list[dict[str, Any]]) -> list[list[dict[str, Any]]]:
    return [values[index : index + MAX_BATCH_OPERATIONS] for index in range(0, len(values), MAX_BATCH_OPERATIONS)]


def phase129_group_order(key: tuple[str, str, str]) -> tuple[int, int, str | int]:
    """Preserve the accepted numbered-plan cohort/domain order without a batch index."""

    plan, kind, domain = key
    if plan not in PHASE129_PLAN_ORDER:
        raise ManifestError(f"unexpected Phase 129 plan in batch grouping: {plan}")
    plan_order = PHASE129_PLAN_ORDER.index(plan)
    if plan == "129-05":
        if kind == "static":
            return plan_order, 0, domain
        if kind == "support" and domain in {"platform", "tracker"}:
            return plan_order, 1, domain
        if kind == "e2e":
            return plan_order, 2, domain
        if kind == "support" and domain in PHASE129_05_FINAL_SUPPORT_DOMAINS:
            return plan_order, 3, PHASE129_05_FINAL_SUPPORT_DOMAINS.index(domain)
        raise ManifestError(f"unexpected Phase 129-05 batch cohort: {kind}/{domain}")
    if PHASE129_PLAN_KINDS.get(plan) != kind:
        raise ManifestError(f"unexpected Phase 129 batch kind: {plan}/{kind}")
    tail = PHASE129_TAIL_DOMAINS.get((plan, kind), ())
    if domain in tail:
        return plan_order, 1, tail.index(domain)
    return plan_order, 0, domain


def phase129_batch_tasks(plan: str, batch_id: str) -> list[str]:
    try:
        tasks = PHASE129_BATCH_TASK_OVERRIDES.get(batch_id, PHASE129_PLAN_TASKS[plan])
    except KeyError as error:
        raise ManifestError(f"unexpected Phase 129 task plan: {plan}") from error
    return list(tasks)


def make_batch_id(plan: str, kind: str, domain: str, index: int, total: int) -> str:
    base = f"{plan}-{kind}-{slug(domain)}"
    return f"{base}-{index:02d}" if total > 1 else base


def build_batch(
    *,
    batch_id: str,
    order: int,
    plan: str,
    wave: int,
    tasks: list[str],
    owner: str,
    verifiers: list[str],
    domain: str,
    authority_groups: list[str],
    operations: list[dict[str, Any]],
    gate_ids: list[str],
    registry: list[dict[str, Any]],
    authority_digests: dict[str, str],
    source_head: str,
) -> dict[str, Any]:
    operation_sources = [item["source"] for item in operations]
    consumer_ids = affected_consumers(registry, plan=plan, operation_sources=operation_sources)
    registry_by_id = {item["id"]: item for item in registry}
    exact_patterns = sorted(set(operation_sources))
    batch = {
        "id": batch_id,
        "order": order,
        "phase": 129,
        "plan": plan,
        "wave": wave,
        "tasks": tasks,
        "owner": owner,
        "verifiers": verifiers,
        "domain": domain,
        "authority_groups": sorted(set(authority_groups)),
        "operations": operations,
        "affected_root_consumers": consumer_ids,
        "preconditions": {
            "repository_head": source_head,
            "authority_sha256": authority_digests,
            "operations_sha256": object_digest(operations),
            "source_paths_sha256": path_list_digest(operation_sources),
            "consumer_before_snapshot_sha256": object_digest(
                [registry_by_id[item] for item in consumer_ids]
            ),
            "max_operations": MAX_BATCH_OPERATIONS,
        },
        "docker_gate_ids": gate_ids,
        "stale_reference_scan": {
            "exact_patterns": exact_patterns,
            "roots": [registry_by_id[item]["path"] for item in consumer_ids],
            "required_remaining_occurrences": 0,
        },
    }
    batch.update(state_fields(registry_by_id, consumer_ids))
    return batch


def derive_planned_manifest(
    repo_root: Path = REPO_ROOT,
    *,
    verify_filesystem: bool = True,
    source_head_override: str | None = None,
) -> dict[str, Any]:
    safe = SafeRepo(repo_root)
    inventory, founder, contract = load_authorities(safe)
    membership = validate_authority_contracts(inventory, founder, contract)
    source_head = source_head_override or read_git_head(safe)
    if GIT_SHA_RE.fullmatch(source_head) is None:
        raise ManifestError("source head override is invalid")
    registry = build_consumer_registry(founder, safe, verify_fs=verify_filesystem)
    authority_digests = {name: spec["sha256"] for name, spec in AUTHORITY_FILES.items()}

    grouped: dict[tuple[str, str, str], list[dict[str, Any]]] = {}
    group_meta: dict[tuple[str, str, str], dict[str, Any]] = {}
    ledger_sources: list[str] = []
    move_targets: list[str] = []
    for spec in GROUP_SPECS:
        for row in membership[spec["id"]]:
            operation = operation_from_row(row, spec, safe, verify_fs=verify_filesystem)
            domain = slug(str(row.get("domain", "")))
            key = (spec["plan"], spec["id_kind"], domain)
            grouped.setdefault(key, []).append(operation)
            group_meta[key] = spec
            ledger_sources.append(operation["source"])
            if operation["action"] == "move":
                move_targets.append(operation["target"])

    nonledger = founder["non_ledger_delete_paths"]
    for path in nonledger:
        source = str(path)
        source_sha = safe.sha256(source) if verify_filesystem else None
        operation = {
            "action": "delete",
            "authority_group": "DELETE-NONLEDGER-CLOSURE",
            "ledger_source": False,
            "source": source,
            "source_sha256": source_sha,
        }
        domain = nonledger_domain(source)
        if source.startswith("backend/testframework/"):
            key = ("129-04", "go-component-delete", domain)
            meta = {
                "plan": "129-04",
                "wave": 4,
                "task": ["TASK-219"],
                "owner": "devops-engineer",
                "verifiers": ["code-reviewer", "qa-auto-ver"],
                "gates": ["architecture-contract", "backend-unit"],
            }
        else:
            key = ("129-05", "support", domain)
            meta = {
                "plan": "129-05",
                "wave": 5,
                "task": ["TASK-219", "TASK-201"] if domain in {
                    "root-scripts", "pipeline", "legacy-stages", "legacy-docs"
                } else ["TASK-219"],
                "owner": "devops-engineer",
                "verifiers": ["code-reviewer", "qa-auto-ver"],
                "gates": ["architecture-contract", "frontend-unit", "python-scaffold"],
            }
        grouped.setdefault(key, []).append(operation)
        group_meta[key] = meta

    hooks_by_domain = hook_guard_operations(inventory, safe, verify_fs=verify_filesystem)
    for domain, operations in hooks_by_domain.items():
        key = ("129-07", "zero-consumer", domain)
        grouped[key] = operations
        group_meta[key] = {
            "plan": "129-07",
            "wave": 7,
            "task": ["TASK-219", "TASK-234"] + (["TASK-201"] if domain == "ws" else []),
            "owner": "devops-engineer",
            "verifiers": ["code-reviewer", "qa-auto-ver"],
            "gates": ["architecture-contract", "backend-unit"],
        }

    reject_portable_aliases(ledger_sources + [str(item) for item in nonledger], "operation source")
    reject_portable_aliases(move_targets, "move target")
    if set(ledger_sources) != {str(row["source_path"]) for row in inventory["rows"]}:
        raise ManifestError("ledger rows are not represented exactly once")
    if len(move_targets) != 507 or len(set(move_targets)) != 507:
        raise ManifestError("move target set is not exact/unique")

    batches: list[dict[str, Any]] = []
    order = 0
    for key in sorted(grouped, key=phase129_group_order):
        plan, kind, domain = key
        operations = sorted(grouped[key], key=lambda item: item["source"])
        chunks = chunked(operations)
        meta = group_meta[key]
        for chunk_index, chunk in enumerate(chunks, 1):
            batch_id = make_batch_id(plan, kind, domain, chunk_index, len(chunks))
            tasks = phase129_batch_tasks(plan, batch_id)
            if tasks != list(meta["task"]):
                raise ManifestError(f"numbered-plan TASK authority drift: {batch_id}")
            order += 1
            batches.append(
                build_batch(
                    batch_id=batch_id,
                    order=order,
                    plan=plan,
                    wave=int(meta["wave"]),
                    tasks=tasks,
                    owner=str(meta["owner"]),
                    verifiers=list(meta["verifiers"]),
                    domain=domain,
                    authority_groups=[item["authority_group"] for item in chunk],
                    operations=chunk,
                    gate_ids=list(meta["gates"]),
                    registry=registry,
                    authority_digests=authority_digests,
                    source_head=source_head,
                )
            )

    ordered_batch_sha256 = hashlib.sha256(
        "\n".join(str(batch["id"]) for batch in batches).encode("utf-8")
    ).hexdigest()
    if ordered_batch_sha256 != PHASE129_BATCH_ORDER_SHA256:
        raise ManifestError("numbered-plan batch order digest changed")

    if verify_filesystem:
        safe.finalize()
    action_counts: dict[str, int] = {}
    for batch in batches:
        if len(batch["operations"]) > MAX_BATCH_OPERATIONS:
            raise ManifestError(f"batch exceeds {MAX_BATCH_OPERATIONS}: {batch['id']}")
        for operation in batch["operations"]:
            action_counts[operation["action"]] = action_counts.get(operation["action"], 0) + 1
    expected_counts = {"move": 507, "delete": 325, "keep-blocked": 288}
    if action_counts != expected_counts:
        raise ManifestError(f"operation totals changed: {action_counts!r}")
    operation_count = sum(len(batch["operations"]) for batch in batches)
    maximum_batch = max(len(batch["operations"]) for batch in batches)
    if (len(batches), operation_count, maximum_batch) != (
        EXPECTED_BATCH_COUNT,
        EXPECTED_OPERATION_COUNT,
        MAX_BATCH_OPERATIONS,
    ):
        raise ManifestError(
            "sealed batch arithmetic changed: expected 173 batches / 1,120 "
            f"operations / max 56, got {len(batches)} / {operation_count} / "
            f"{maximum_batch}"
        )

    return {
        "schema_version": 1,
        "manifest_id": "PHASE-129-FOUNDER-GREENFIELD-EXACT-CUTOVER",
        "status": "planned",
        "source": {
            "repository_head": source_head,
            "ledger_revision": LEDGER_REVISION,
            "authority_files": [
                {"id": name, "path": spec["path"], "sha256": spec["sha256"]}
                for name, spec in AUTHORITY_FILES.items()
            ],
        },
        "invariants": {
            "max_batch_operations": MAX_BATCH_OPERATIONS,
            "ledger_rows_exact_once": 1049,
            "ledger_moves": 507,
            "ledger_keep_blocked": 245,
            "ledger_deletes": 297,
            "nonledger_deletes": 28,
            "hook_files_keep_blocked": 43,
            "protected_prefix": PROTECTED_PREFIX,
            "no_command_execution_from_manifest": True,
            "state_machine": ["planned", "applied-unverified", "verified"],
        },
        "gate_registry": copy.deepcopy(GATE_REGISTRY),
        "consumer_registry": registry,
        "batches": batches,
        "totals": {
            "batches": len(batches),
            "operations": sum(len(batch["operations"]) for batch in batches),
            "actions": action_counts,
            "consumer_registry_entries": len(registry),
        },
    }


def write_transactionally(
    repo_root: Path,
    relative_target: str,
    payload: bytes,
    *,
    inject_destination_race: bool = False,
) -> Path:
    if relative_target != OUTPUT_PATH:
        raise ManifestError(f"unexpected write target; only {OUTPUT_PATH} is allowed")
    if len(payload) > 32 * 1024 * 1024:
        raise ManifestError("manifest payload exceeds 32 MiB write budget")
    require_canonical_path(relative_target, "write target")
    safe = SafeRepo(repo_root)
    parent_relative = str(PurePosixPath(relative_target).parent)
    parent_fd, parent_signature = safe.open_directory(parent_relative)
    target_name = PurePosixPath(relative_target).name
    staging_name = f".cutover-manifest.{os.getpid()}.{secrets.token_hex(8)}.tmp"
    try:
        existing = os.stat(target_name, dir_fd=parent_fd, follow_symlinks=False)
    except FileNotFoundError:
        existing = None
    except OSError:
        os.close(parent_fd)
        raise
    if existing is not None:
        os.close(parent_fd)
        raise ManifestError("existing manifest target rejected; overwrite is forbidden")

    descriptor = -1
    try:
        flags = os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_CLOEXEC | os.O_NOFOLLOW
        descriptor = os.open(staging_name, flags, 0o600, dir_fd=parent_fd)
        staging_meta = os.fstat(descriptor)
        if not stat.S_ISREG(staging_meta.st_mode) or int(getattr(staging_meta, "st_nlink", 1)) != 1:
            raise ManifestError("transactional staging file rejected")
        written = 0
        while written < len(payload):
            written += os.write(descriptor, payload[written:])
        os.fsync(descriptor)
        os.close(descriptor)
        descriptor = -1
        verify_fd = os.open(staging_name, os.O_RDONLY | os.O_CLOEXEC | os.O_NOFOLLOW, dir_fd=parent_fd)
        try:
            staged_meta = os.fstat(verify_fd)
            if (
                not stat.S_ISREG(staged_meta.st_mode)
                or int(staged_meta.st_nlink) != 1
                or metadata_signature(staged_meta, content=True)
                != metadata_signature(staging_meta, content=True)
            ):
                raise ManifestError("transactional staging identity changed")
            staged_payload = bytearray()
            while True:
                chunk = os.read(verify_fd, 1024 * 1024)
                if not chunk:
                    break
                staged_payload.extend(chunk)
                if len(staged_payload) > len(payload):
                    raise ManifestError("transactional staging grew beyond payload")
            if metadata_signature(os.fstat(verify_fd), content=True) != metadata_signature(
                staged_meta, content=True
            ):
                raise ManifestError("transactional staging changed during verification")
        finally:
            os.close(verify_fd)
        if bytes(staged_payload) != payload:
            raise ManifestError("transactional staging verification failed")
        if inject_destination_race:
            racer = os.open(
                target_name,
                os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_CLOEXEC | os.O_NOFOLLOW,
                0o600,
                dir_fd=parent_fd,
            )
            os.write(racer, b"racer")
            os.fsync(racer)
            os.close(racer)
            os.fsync(parent_fd)
        if parent_signature != metadata_signature(os.fstat(parent_fd), content=False):
            raise ManifestError("manifest parent changed during staging (TOCTOU)")
        rename_noreplace(parent_fd, staging_name, parent_fd, target_name)
        os.fsync(parent_fd)
        return Path(repo_root) / relative_target
    finally:
        if descriptor >= 0:
            os.close(descriptor)
        try:
            os.unlink(staging_name, dir_fd=parent_fd)
        except FileNotFoundError:
            pass
        os.close(parent_fd)


def run_write_self_test() -> list[str]:
    """Prove an appearing final destination cannot be overwritten."""

    failures: list[str] = []
    deep_errors = canonical_path_errors("/".join(["a"] * (MAX_PATH_COMPONENTS + 1)), "sentinel")
    if not any("component budget" in item for item in deep_errors):
        failures.append("builder component-depth sentinel was not rejected")
    chunks = chunked([{"index": index} for index in range(MAX_BATCH_OPERATIONS + 1)])
    if [len(chunk) for chunk in chunks] != [MAX_BATCH_OPERATIONS, 1]:
        failures.append("builder 56-operation chunk boundary sentinel failed")
    try:
        json_load(b'{"outer":{"value":1,"value":2}}', "duplicate-key sentinel")
    except ManifestError as problem:
        if "duplicate JSON key" not in str(problem):
            failures.append("builder nested duplicate-key sentinel returned wrong failure")
    else:
        failures.append("builder nested duplicate JSON key was accepted")
    for reserved in (
        "COM1.txt", "NUL.json", "COM¹", "COM².txt", "COM³.log",
        "LPT¹", "LPT².txt", "LPT³.log",
    ):
        if not any(
            "reserved Windows name" in item
            for item in canonical_path_errors(f"legacy/{reserved}", "sentinel")
        ):
            failures.append(f"builder Windows device-name sentinel was accepted: {reserved}")
    with tempfile.TemporaryDirectory(prefix="hub-builder-race-") as temporary:
        root = Path(temporary)
        (root / ".planning" / "audits" / "test-rebase").mkdir(parents=True)
        target = root / OUTPUT_PATH
        try:
            write_transactionally(
                root,
                OUTPUT_PATH,
                b"{}\n",
                inject_destination_race=True,
            )
        except ManifestError as problem:
            if "atomic no-replace destination exists" not in str(problem):
                failures.append("builder destination-race returned wrong failure")
        else:
            failures.append("builder destination-race did not fail")
        if not target.is_file() or target.read_bytes() != b"racer":
            failures.append("builder destination-race overwrote appearing target")
    return failures


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo-root", type=Path, default=REPO_ROOT)
    parser.add_argument("--dry-run", action="store_true", help="explicit alias for the default stdout mode")
    parser.add_argument("--write", metavar="PATH", help=f"transactionally write only {OUTPUT_PATH}")
    parser.add_argument("--self-test", action="store_true", help="run temp-only no-overwrite sentinel")
    args = parser.parse_args(argv)
    if sum(bool(value) for value in (args.write, args.dry_run, args.self_test)) > 1:
        parser.error("--write, --dry-run and --self-test are mutually exclusive")
    try:
        if args.self_test:
            failures = run_write_self_test()
            if failures:
                raise ManifestError(failures[0])
            print("PASS: builder atomic no-replace race sentinel")
            return 0
        manifest = derive_planned_manifest(args.repo_root)
        payload = canonical_json_bytes(manifest, pretty=True)
        if args.write:
            target = write_transactionally(args.repo_root, args.write, payload)
            print(f"PASS: wrote sealed planned manifest to {target}")
        else:
            sys.stdout.buffer.write(payload)
    except (ManifestError, OSError, UnicodeError) as error:
        print(f"FAIL: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
