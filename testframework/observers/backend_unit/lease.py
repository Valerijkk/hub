from __future__ import annotations

import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import stat
import subprocess
from typing import Any, Mapping, Protocol


_FIXED_LEASE_PATH = PurePosixPath("/reports/.backend-unit-source-mutation.lease")
_SOURCE_ROOTS = (
    "backend/cmd",
    "backend/internal",
    "backend/tests",
    "backend/ent/schema",
    "backend/migrations",
    "backend/testframework",
)
_CANONICAL_BUILD_INPUT_PATHS = (
    "docker-compose.local.yml",
    "testframework/docker/Dockerfile.qa",
    "testframework/docker/entrypoints/backend-unit.sh",
    "testframework/docker/backend_unit/runtime.sh",
    "testframework/docker/backend_unit/plan.sh",
    "testframework/docker/backend_unit/execute.sh",
    "testframework/docker/backend_unit/receipt.sh",
    "testframework/observers/**",
    "testframework/schemas/backend-unit-package-plan.schema.v1.json",
    "testframework/schemas/backend-unit-container-receipt.schema.v1.json",
    "testframework/schemas/backend-unit-observer-receipt.schema.v1.json",
    "testframework/policies/backend-unit-observer.v1.json",
    "testframework/policies/backend-unit-phase150-08c2-subject.v1.json",
    "testframework/policies/root-compose-testframework.v1.json",
    "testframework/selftests/architecture/docker/**",
    "testframework/selftests/architecture/observers/**",
)
_MANIFEST_PATHS = (
    "docker-compose.local.yml",
    "testframework/docker/Dockerfile.qa",
    "testframework/docker/entrypoints/backend-unit.sh",
    "testframework/docker/backend_unit/runtime.sh",
    "testframework/docker/backend_unit/plan.sh",
    "testframework/docker/backend_unit/execute.sh",
    "testframework/docker/backend_unit/receipt.sh",
    "testframework/observers/__init__.py",
    "testframework/observers/backend_unit/__init__.py",
    "testframework/observers/backend_unit/contracts.py",
    "testframework/observers/backend_unit/safe_paths.py",
    "testframework/observers/backend_unit/lease.py",
    "testframework/observers/backend_unit/resources.py",
    "testframework/observers/backend_unit/docker_adapter.py",
    "testframework/observers/backend_unit/publication.py",
    "testframework/observers/backend_unit/observe.py",
    "testframework/schemas/backend-unit-package-plan.schema.v1.json",
    "testframework/schemas/backend-unit-container-receipt.schema.v1.json",
    "testframework/schemas/backend-unit-observer-receipt.schema.v1.json",
    "testframework/policies/backend-unit-observer.v1.json",
    "testframework/policies/backend-unit-phase150-08c2-subject.v1.json",
    "testframework/policies/root-compose-testframework.v1.json",
    "testframework/selftests/architecture/docker/test_root_compose_contract.py",
    "testframework/selftests/architecture/observers/test_backend_unit_receipts.py",
    "testframework/selftests/architecture/observers/test_backend_unit_receipts_review_contract.py",
    "testframework/selftests/architecture/observers/test_backend_unit_observer.py",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/valid/container-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/valid/final-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/valid/observer-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/valid/package-plan.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/valid/SEAL",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/bound-backend-image-observer-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/build-input-hash-observer-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/build-input-manifest-observer-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/count-package-plan.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/discovery-argv-package-plan.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/duplicate-import-path-package-plan.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/engine-id-observer-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/fixed-image-id-observer-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/foreign-run-observer-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/inventory-container-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/inventory-hash-container-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/non-bijection-package-plan.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/nonce-observer-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/row-set-hash-package-plan.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/SEAL.extra-field",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/SEAL.missing-field",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/SEAL.reordered",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/timeout-outcome-final-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/unknown-key-package-plan.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/uppercase-digest-package-plan.json",
)
_FROZEN_SUBJECTS = {
    "backend/internal/calls/call_durable_credential_allocation.go": (
        "sha256:be6ff058468c4e288e0b0d7060fd17cf3118229d009f28dc056d89f281e99d4d"
    ),
    "backend/tests/calls/call_durable_credential_allocation_test.go": (
        "sha256:0ae525c86a75369bdb5790f26b1308f42574fb545883328da24011d4150b03a9"
    ),
}
_FIXED_IMAGE_INSPECT_ARGVS = {
    "python_contract": (
        "docker",
        "image",
        "inspect",
        "--format",
        "{{json .}}",
        "worktree-master-python-contract",
    ),
    "backend_unit": (
        "docker",
        "image",
        "inspect",
        "--format",
        "{{json .}}",
        "worktree-master-backend-unit",
    ),
}
_GIT_HEAD_ARGV = ("git", "rev-parse", "HEAD")
_GIT_STATUS_ARGV = ("git", "status", "--porcelain=v1", "--untracked-files=all")
_MAX_FILE_BYTES = 64 * 1024 * 1024
_SHA256_PREFIX = "sha256:"


class SourceLeaseError(RuntimeError):
    """The observer cannot establish an authoritative source lease."""


class SourceMutationError(SourceLeaseError):
    """A source snapshot is dirty, malformed, or changed during observation."""


class SourceProbe(Protocol):
    def git(self, argv: tuple[str, ...]) -> str: ...

    def canonical_build_input_manifest(self) -> bytes: ...

    def fixed_image_ids(self) -> dict[str, str]: ...

    def frozen_subject_digests(self) -> dict[str, str]: ...


class _SystemSourceProbe:
    """The production-only seam. It never changes ACLs or source contents."""

    def __init__(self, repo_root: Path) -> None:
        self._repo_root = repo_root

    def git(self, argv: tuple[str, ...]) -> str:
        try:
            completed = subprocess.run(
                argv,
                cwd=self._repo_root,
                check=False,
                capture_output=True,
                encoding="utf-8",
                errors="strict",
                shell=False,
                timeout=10,
            )
        except (OSError, subprocess.SubprocessError) as error:
            raise SourceLeaseError("Git snapshot command could not run") from error
        if completed.returncode != 0 or completed.stderr:
            raise SourceLeaseError("Git snapshot command was not authoritative")
        return completed.stdout

    def source_root_digests(self) -> dict[str, str]:
        return {
            root: _sha256_prefixed(_canonical_tree_bytes(self._repo_root / root))
            for root in _SOURCE_ROOTS
        }

    def canonical_build_input_manifest(self) -> bytes:
        expanded: dict[str, Path] = {}
        for pattern in _CANONICAL_BUILD_INPUT_PATHS:
            regular_paths = _expand_canonical_pattern(self._repo_root, pattern)
            for relative, candidate in regular_paths:
                if relative in expanded:
                    raise SourceLeaseError("Canonical build inputs overlap")
                expanded[relative] = candidate
        if set(expanded) != set(_MANIFEST_PATHS):
            raise SourceLeaseError(
                "Canonical build-input expansion differs from the frozen file set"
            )
        rows = [
            relative.encode("utf-8")
            + b"\t"
            + _sha256_prefixed(_read_regular_bytes(path)).encode("ascii")
            + b"\n"
            for relative in _MANIFEST_PATHS
            for path in (expanded[relative],)
        ]
        if not rows:
            raise SourceLeaseError("Canonical build inputs are empty")
        return b"".join(rows)

    def fixed_image_ids(self) -> dict[str, str]:
        image_ids: dict[str, str] = {}
        for name, argv in _FIXED_IMAGE_INSPECT_ARGVS.items():
            try:
                completed = subprocess.run(
                    argv,
                    cwd=self._repo_root,
                    check=False,
                    capture_output=True,
                    encoding="utf-8",
                    errors="strict",
                    shell=False,
                    timeout=10,
                )
            except (OSError, subprocess.SubprocessError) as error:
                raise SourceLeaseError("Fixed image inspect could not run") from error
            if completed.returncode != 0 or completed.stderr:
                raise SourceLeaseError("Fixed image inspect was not authoritative")
            try:
                document = json.loads(completed.stdout)
            except (TypeError, ValueError) as error:
                raise SourceLeaseError("Fixed image inspect was malformed") from error
            image_id = document.get("Id") if type(document) is dict else None
            if not _is_sha256(image_id):
                raise SourceLeaseError("Fixed image inspect did not return an image ID")
            image_ids[name] = image_id
        return image_ids

    def frozen_subject_digests(self) -> dict[str, str]:
        observed = {
            relative: _sha256_prefixed(_read_regular_bytes(self._repo_root / relative))
            for relative in _FROZEN_SUBJECTS
        }
        if observed != _FROZEN_SUBJECTS:
            raise SourceMutationError("A frozen subject does not match its pinned digest")
        return observed


class SourceLease:
    """An exclusive lease and immutable source-observation snapshot provider."""

    def __init__(
        self,
        path: Path,
        descriptor: int,
        probe: SourceProbe,
        *,
        require_source_roots: bool,
    ) -> None:
        self.path = path
        self._descriptor = descriptor
        self.probe = probe
        self._require_source_roots = require_source_roots
        self._released = False
        self._identity = _identity(os.fstat(descriptor))

    @classmethod
    def acquire(
        cls,
        path: Path,
        *,
        probe: SourceProbe | None = None,
        repo_root: Path | None = None,
    ) -> SourceLease:
        _validate_fixed_lease_path(path)
        _validate_lease_parent(path.parent)
        flags = os.O_RDWR | os.O_CREAT | os.O_EXCL
        flags |= getattr(os, "O_BINARY", 0)
        flags |= getattr(os, "O_NOFOLLOW", 0)
        descriptor = -1
        opened: os.stat_result | None = None
        created = False
        try:
            descriptor = os.open(path, flags, 0o600)
            created = True
            opened = os.fstat(descriptor)
            if not _admitted_regular(opened) or opened.st_nlink != 1:
                raise SourceLeaseError("Source lease is not an exclusive regular file")
            on_path = path.lstat()
            if _identity(opened) != _identity(on_path) or not _admitted_regular(on_path):
                raise SourceLeaseError("Source lease changed while being acquired")
            if probe is None:
                root = repo_root or Path(__file__).resolve().parents[3]
                probe = _SystemSourceProbe(root)
                require_source_roots = True
            else:
                require_source_roots = callable(getattr(probe, "source_root_digests", None))
            return cls(
                path,
                descriptor,
                probe,
                require_source_roots=require_source_roots,
            )
        except Exception:
            if descriptor >= 0:
                try:
                    os.close(descriptor)
                except OSError:
                    pass
            if created and opened is not None and (path.exists() or path.is_symlink()):
                try:
                    current = path.lstat()
                    if _identity(current) == _identity(opened):
                        path.unlink()
                except OSError:
                    pass
            raise

    def snapshot(self) -> dict[str, Any]:
        self._assert_active()
        self._assert_path_identity()
        try:
            head = _parse_git_head(self.probe.git(_GIT_HEAD_ARGV))
            status = _parse_git_status(self.probe.git(_GIT_STATUS_ARGV))
            manifest = self.probe.canonical_build_input_manifest()
            fixed_image_ids = self.probe.fixed_image_ids()
            frozen_subjects = self.probe.frozen_subject_digests()
        except SourceLeaseError:
            raise
        except Exception as error:
            raise SourceLeaseError("Source snapshot probe failed") from error
        if type(manifest) is not bytes or not manifest or b"\r" in manifest:
            raise SourceLeaseError("Canonical build input manifest is malformed")
        snapshot: dict[str, Any] = {
            "head": head,
            "status": status,
            "canonical_build_input_manifest_bytes": manifest,
            "canonical_build_input_manifest_sha256": _sha256_prefixed(manifest),
            "fixed_image_ids": fixed_image_ids,
            "frozen_subjects": frozen_subjects,
        }
        root_method = getattr(self.probe, "source_root_digests", None)
        if callable(root_method):
            try:
                snapshot["source_root_sha256"] = root_method()
            except Exception as error:
                raise SourceLeaseError("Source-root snapshot probe failed") from error
        _validate_snapshot(snapshot, require_source_roots=self._require_source_roots)
        return snapshot

    def verify_unchanged(self, before: Mapping[str, Any], after: Mapping[str, Any]) -> None:
        self._assert_active()
        self._assert_path_identity()
        _validate_snapshot(before, require_source_roots=self._require_source_roots)
        _validate_snapshot(after, require_source_roots=self._require_source_roots)
        if before["status"] or after["status"]:
            raise SourceMutationError("Git worktree is dirty")
        if dict(before) != dict(after):
            raise SourceMutationError("Source snapshot changed")

    def verify_snapshot_chain(
        self,
        prebuild: Mapping[str, Any],
        postgreen: Mapping[str, Any],
        observer_snapshot: Mapping[str, Any],
    ) -> None:
        self.verify_unchanged(prebuild, postgreen)
        self.verify_unchanged(postgreen, observer_snapshot)

    def release(self) -> None:
        if self._released:
            return
        try:
            current = self.path.lstat()
        except OSError as error:
            raise SourceLeaseError("Source lease could not be inspected before release") from error
        try:
            if (
                _identity(current) != self._identity
                or not _admitted_regular(current)
                or current.st_nlink != 1
            ):
                raise SourceLeaseError("Source lease changed before release")
        finally:
            try:
                os.close(self._descriptor)
            except OSError as error:
                raise SourceLeaseError("Source lease could not be closed") from error
            self._released = True
        try:
            self.path.unlink()
        except OSError as error:
            raise SourceLeaseError("Source lease could not be removed") from error

    def _assert_active(self) -> None:
        if self._released:
            raise SourceLeaseError("Source lease has already been released")

    def _assert_path_identity(self) -> None:
        try:
            current = self.path.lstat()
        except OSError as error:
            raise SourceLeaseError("Source lease path is unavailable") from error
        if (
            _identity(current) != self._identity
            or not _admitted_regular(current)
            or current.st_nlink != 1
        ):
            raise SourceLeaseError("Source lease path no longer names the acquired lease")


def _validate_fixed_lease_path(path: Path) -> None:
    if not isinstance(path, Path) or path.as_posix() != _FIXED_LEASE_PATH.as_posix():
        raise SourceLeaseError("Source lease path is not the fixed observer lease")


def _validate_lease_parent(parent: Path) -> None:
    try:
        value = parent.lstat()
    except OSError as error:
        raise SourceLeaseError("Source lease parent is unavailable") from error
    if not _admitted_directory(value):
        raise SourceLeaseError("Source lease parent is unsafe")


def _validate_snapshot(snapshot: Mapping[str, Any], *, require_source_roots: bool) -> None:
    if not isinstance(snapshot, Mapping):
        raise SourceMutationError("Source snapshot is not a mapping")
    required = {
        "head",
        "status",
        "canonical_build_input_manifest_bytes",
        "canonical_build_input_manifest_sha256",
        "fixed_image_ids",
        "frozen_subjects",
    }
    allowed = required | {"source_root_sha256"}
    keys = set(snapshot)
    if not required <= keys or not keys <= allowed:
        raise SourceMutationError("Source snapshot keys are not closed")
    if require_source_roots and "source_root_sha256" not in snapshot:
        raise SourceMutationError("Source-root snapshot is unavailable")
    head = snapshot["head"]
    status = snapshot["status"]
    manifest = snapshot["canonical_build_input_manifest_bytes"]
    manifest_digest = snapshot["canonical_build_input_manifest_sha256"]
    if not _is_git_sha(head) or type(status) is not str:
        raise SourceMutationError("Git snapshot is malformed")
    if (
        type(manifest) is not bytes
        or not manifest
        or b"\r" in manifest
        or manifest_digest != _sha256_prefixed(manifest)
    ):
        raise SourceMutationError("Canonical build input manifest changed or is malformed")
    _validate_digest_mapping(
        snapshot["fixed_image_ids"],
        ("python_contract", "backend_unit"),
        "Fixed image IDs",
    )
    _validate_digest_mapping(
        snapshot["frozen_subjects"],
        tuple(_FROZEN_SUBJECTS),
        "Frozen subject digests",
    )
    if "source_root_sha256" in snapshot:
        _validate_digest_mapping(
            snapshot["source_root_sha256"], _SOURCE_ROOTS, "Source-root digests"
        )


def _validate_digest_mapping(value: object, keys: tuple[str, ...], label: str) -> None:
    if (
        type(value) is not dict
        or tuple(value) != keys
        or not all(_is_sha256(value[key]) for key in keys)
    ):
        raise SourceMutationError(f"{label} are malformed")


def _parse_git_head(value: object) -> str:
    if type(value) is not str or not value.endswith("\n") or value.count("\n") != 1:
        raise SourceLeaseError("Git HEAD output is malformed")
    head = value[:-1]
    if not _is_git_sha(head):
        raise SourceLeaseError("Git HEAD is not a full lower-case SHA")
    return head


def _parse_git_status(value: object) -> str:
    if type(value) is not str or "\x00" in value:
        raise SourceLeaseError("Git status output is malformed")
    return value


def _canonical_tree_bytes(root: Path) -> bytes:
    try:
        root_stat = root.lstat()
    except OSError as error:
        raise SourceLeaseError("Source root is unavailable") from error
    if not _admitted_directory(root_stat):
        raise SourceLeaseError("Source root is unsafe")
    rows: list[bytes] = []

    def visit(directory: Path, relative: PurePosixPath) -> None:
        try:
            with os.scandir(directory) as scan:
                entries = sorted(scan, key=lambda entry: entry.name)
        except OSError as error:
            raise SourceLeaseError("Source root cannot be enumerated") from error
        for entry in entries:
            path = Path(entry.path)
            child_relative = relative / entry.name
            try:
                value = path.lstat()
            except OSError as error:
                raise SourceLeaseError("Source root changed while being enumerated") from error
            if _admitted_directory(value):
                rows.append(b"D\t" + child_relative.as_posix().encode("utf-8") + b"\n")
                visit(path, child_relative)
            elif _admitted_regular(value):
                rows.append(
                    b"F\t"
                    + child_relative.as_posix().encode("utf-8")
                    + b"\t"
                    + _sha256_prefixed(_read_regular_bytes(path)).encode("ascii")
                    + b"\n"
                )
            else:
                raise SourceLeaseError("Source root contains an unsafe entry")

    visit(root, PurePosixPath("."))
    return b"".join(rows)


def _expand_canonical_pattern(repo_root: Path, pattern: str) -> tuple[tuple[str, Path], ...]:
    if type(pattern) is not str or not pattern or pattern.startswith("/"):
        raise SourceLeaseError("Canonical build-input pattern is malformed")
    if "*" not in pattern:
        candidate = repo_root / pattern
        relative = _contained_regular_relative(repo_root, candidate)
        if relative is None:
            raise SourceLeaseError("A canonical build input is not a regular file")
        return ((relative, candidate),)
    if not pattern.endswith("/**") or pattern.count("*") != 2:
        raise SourceLeaseError("Canonical build-input wildcard is not closed")
    root = repo_root / pattern.removesuffix("/**")
    _contained_directory_relative(repo_root, root)
    files: list[tuple[str, Path]] = []

    def visit(directory: Path) -> None:
        try:
            with os.scandir(directory) as scan:
                entries = sorted(scan, key=lambda entry: entry.name)
        except OSError as error:
            raise SourceLeaseError("Canonical build-input directory cannot be enumerated") from error
        for entry in entries:
            candidate = Path(entry.path)
            try:
                value = candidate.lstat()
            except OSError as error:
                raise SourceLeaseError("Canonical build-input directory changed") from error
            if _admitted_directory(value):
                visit(candidate)
                continue
            if _admitted_regular(value):
                relative = _contained_regular_relative(repo_root, candidate)
                if relative is None:
                    raise SourceLeaseError("Canonical build-input file is not regular")
                files.append((relative, candidate))
                continue
            raise SourceLeaseError("Canonical build-input directory contains an unsafe entry")

    visit(root)
    if not files:
        raise SourceLeaseError("Canonical build-input wildcard matched no files")
    return tuple(files)


def _contained_directory_relative(repo_root: Path, candidate: Path) -> str:
    try:
        root_stat = repo_root.lstat()
        if not _admitted_directory(root_stat):
            raise SourceLeaseError("Repository root is unsafe")
        relative = candidate.relative_to(repo_root)
        if not relative.parts or any(part in {"", ".", ".."} for part in relative.parts):
            raise SourceLeaseError("Canonical build-input directory escaped the repository")
        current = repo_root
        for component in relative.parts:
            current /= component
            if not _admitted_directory(current.lstat()):
                raise SourceLeaseError("Canonical build-input directory is unsafe")
        return relative.as_posix()
    except (OSError, ValueError) as error:
        raise SourceLeaseError("Canonical build-input directory cannot be resolved safely") from error


def _contained_regular_relative(repo_root: Path, candidate: Path) -> str | None:
    try:
        root_stat = repo_root.lstat()
        if not _admitted_directory(root_stat):
            raise SourceLeaseError("Repository root is unsafe")
        relative = candidate.relative_to(repo_root)
        if not relative.parts or any(part in {"", ".", ".."} for part in relative.parts):
            raise SourceLeaseError("Canonical build input escaped the repository")
        current = repo_root
        for index, component in enumerate(relative.parts):
            current /= component
            value = current.lstat()
            final = index == len(relative.parts) - 1
            if final and _admitted_regular(value):
                return relative.as_posix()
            if final and _admitted_directory(value):
                return None
            if not final and _admitted_directory(value):
                continue
            raise SourceLeaseError("Canonical build input is unsafe")
        raise SourceLeaseError("Canonical build input is unavailable")
    except (OSError, ValueError) as error:
        raise SourceLeaseError("Canonical build input cannot be resolved safely") from error


def _read_regular_bytes(path: Path) -> bytes:
    descriptor = -1
    try:
        before = path.lstat()
        if not _admitted_regular(before):
            raise SourceLeaseError("Source file is unsafe")
        flags = (
            os.O_RDONLY
            | getattr(os, "O_BINARY", 0)
            | getattr(os, "O_NOFOLLOW", 0)
        )
        descriptor = os.open(path, flags)
        opened = os.fstat(descriptor)
        if _identity(before) != _identity(opened) or not _admitted_regular(opened):
            raise SourceLeaseError("Source file changed while being opened")
        chunks: list[bytes] = []
        remaining = _MAX_FILE_BYTES + 1
        while remaining:
            chunk = os.read(descriptor, min(65_536, remaining))
            if not chunk:
                break
            chunks.append(chunk)
            remaining -= len(chunk)
        after = os.fstat(descriptor)
        content = b"".join(chunks)
        if _identity(opened) != _identity(after) or not 1 <= len(content) <= _MAX_FILE_BYTES:
            raise SourceLeaseError("Source file changed while being read")
        return content
    except OSError as error:
        raise SourceLeaseError("Source file could not be read") from error
    finally:
        if descriptor >= 0:
            try:
                os.close(descriptor)
            except OSError:
                pass


def _admitted_regular(value: os.stat_result) -> bool:
    return stat.S_ISREG(value.st_mode) and not _is_reparse_point(value)


def _admitted_directory(value: os.stat_result) -> bool:
    return stat.S_ISDIR(value.st_mode) and not _is_reparse_point(value)


def _is_reparse_point(value: os.stat_result) -> bool:
    reparse = getattr(stat, "FILE_ATTRIBUTE_REPARSE_POINT", 0)
    return bool(getattr(value, "st_file_attributes", 0) & reparse)


def _identity(value: os.stat_result) -> tuple[int, int, int, int, int]:
    return (
        value.st_dev,
        value.st_ino,
        value.st_size,
        value.st_mtime_ns,
        value.st_ctime_ns,
    )


def _is_sha256(value: object) -> bool:
    return (
        type(value) is str
        and len(value) == 71
        and value.startswith(_SHA256_PREFIX)
        and set(value[len(_SHA256_PREFIX):]) <= set("0123456789abcdef")
    )


def _is_git_sha(value: object) -> bool:
    return (
        type(value) is str
        and len(value) == 40
        and set(value) <= set("0123456789abcdef")
    )


def _sha256_prefixed(value: bytes) -> str:
    return _SHA256_PREFIX + hashlib.sha256(value).hexdigest()
