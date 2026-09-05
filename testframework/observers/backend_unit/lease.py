from __future__ import annotations

import ctypes
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import subprocess
from typing import Any, Mapping, Protocol

from . import safe_paths


_FIXED_LEASE_RELATIVE = PurePosixPath(
    "testframework/reports/.backend-unit-source-mutation.lease"
)
_FIXED_LEASE_NAME = _FIXED_LEASE_RELATIVE.name
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
            root: _sha256_prefixed(_canonical_tree_bytes(self._repo_root, root))
            for root in _SOURCE_ROOTS
        }

    def canonical_build_input_manifest(self) -> bytes:
        expanded: set[str] = set()
        for pattern in _CANONICAL_BUILD_INPUT_PATHS:
            regular_paths = _expand_canonical_pattern(self._repo_root, pattern)
            for relative in regular_paths:
                if relative in expanded:
                    raise SourceLeaseError("Canonical build inputs overlap")
                expanded.add(relative)
        if expanded != set(_MANIFEST_PATHS):
            raise SourceLeaseError(
                "Canonical build-input expansion differs from the frozen file set"
            )
        rows = [
            relative.encode("utf-8")
            + b"\t"
            + _sha256_prefixed(
                _read_regular_bytes(self._repo_root, relative)
            ).encode("ascii")
            + b"\n"
            for relative in _MANIFEST_PATHS
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
            relative: _sha256_prefixed(
                _read_regular_bytes(self._repo_root, relative)
            )
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
        parent_descriptor: int,
        probe: SourceProbe,
        *,
        require_source_roots: bool,
    ) -> None:
        self.path = path
        self._descriptor = descriptor
        self._parent_descriptor = parent_descriptor
        self._lease_name = path.name
        self.probe = probe
        self._require_source_roots = require_source_roots
        self._released = False
        self._parent_identity = _lease_parent_identity(parent_descriptor)
        self._identity = _lease_file_identity(
            descriptor, self._parent_identity[0]
        )

    @classmethod
    def acquire(
        cls,
        path: Path,
        *,
        probe: SourceProbe | None = None,
        repo_root: Path | None = None,
        trusted_reports_root: Path | None = None,
    ) -> SourceLease:
        if probe is None:
            if trusted_reports_root is not None:
                raise SourceLeaseError(
                    "Production source lease cannot use an injected reports root"
                )
            root = _validated_repository_root(repo_root)
            _validate_repository_lease_path(path, root)
            probe = _SystemSourceProbe(root)
            require_source_roots = True
        elif repo_root is not None:
            if trusted_reports_root is not None:
                raise SourceLeaseError("Injected probe lease roots are ambiguous")
            root = _validated_repository_root(repo_root)
            _validate_repository_lease_path(path, root)
            require_source_roots = callable(
                getattr(probe, "source_root_digests", None)
            )
        else:
            _validate_injected_probe_lease_path(path, trusted_reports_root)
            require_source_roots = callable(
                getattr(probe, "source_root_digests", None)
            )
        parent_descriptor = -1
        descriptor = -1
        parent_identity: tuple[int, int, int | None, int | None] | None = None
        lease_identity: tuple[int, int, int] | None = None
        try:
            parent_descriptor, parent_identity = _open_lease_parent(path.parent)
            descriptor = _create_lease_relative(parent_descriptor, path.name)
            lease_identity = _lease_file_identity(descriptor, parent_identity[0])
            _assert_lease_parent_identity(parent_descriptor, parent_identity)
            return cls(
                path,
                descriptor,
                parent_descriptor,
                probe,
                require_source_roots=require_source_roots,
            )
        except Exception:
            if descriptor >= 0:
                if (
                    lease_identity is not None
                    and parent_identity is not None
                    and parent_descriptor >= 0
                ):
                    _discard_created_lease(
                        parent_descriptor,
                        path.name,
                        descriptor,
                        lease_identity,
                        parent_identity,
                    )
                _close_lease_descriptor_quietly(descriptor)
            if parent_descriptor >= 0:
                _close_lease_descriptor_quietly(parent_descriptor)
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
        failure: Exception | None = None
        try:
            self._assert_path_identity()
            _delete_owned_lease(
                self._parent_descriptor,
                self._lease_name,
                self._descriptor,
                self._identity,
                self._parent_identity,
            )
        except SourceLeaseError as error:
            failure = error
        except Exception as error:
            failure = SourceLeaseError("Source lease could not be released")
            failure.__cause__ = error
        finally:
            close_failures: list[Exception] = []
            for descriptor in (self._descriptor, self._parent_descriptor):
                try:
                    _close_lease_descriptor(descriptor)
                except Exception as error:
                    close_failures.append(error)
            self._released = True
        if failure is not None:
            raise failure
        if close_failures:
            raise SourceLeaseError("Source lease descriptor could not be closed") from close_failures[0]

    def _assert_active(self) -> None:
        if self._released:
            raise SourceLeaseError("Source lease has already been released")

    def _assert_path_identity(self) -> None:
        _assert_lease_parent_identity(
            self._parent_descriptor, self._parent_identity
        )
        if _lease_file_identity(
            self._descriptor, self._parent_identity[0]
        ) != self._identity:
            raise SourceLeaseError("Source lease path no longer names the acquired lease")
        current_descriptor = -1
        try:
            current_descriptor = _open_existing_lease_relative(
                self._parent_descriptor, self._lease_name
            )
            if _lease_file_identity(
                current_descriptor, self._parent_identity[0]
            ) != self._identity:
                raise SourceLeaseError(
                    "Source lease path no longer names the acquired lease"
                )
        except (OSError, safe_paths.SafePathError) as error:
            raise SourceLeaseError("Source lease path is unavailable") from error
        finally:
            if current_descriptor >= 0:
                _close_lease_descriptor_quietly(current_descriptor)


def _validated_repository_root(repo_root: Path | None) -> Path:
    root = repo_root or Path(__file__).resolve().parents[3]
    if not isinstance(root, Path) or not root.is_absolute():
        raise SourceLeaseError("Repository root is not an absolute anchored path")
    return root


def _validate_repository_lease_path(path: Path, repo_root: Path) -> None:
    expected = repo_root.joinpath(*_FIXED_LEASE_RELATIVE.parts)
    if not isinstance(path, Path) or path != expected:
        raise SourceLeaseError("Source lease path is not the fixed observer lease")


def _validate_injected_probe_lease_path(
    path: Path, trusted_reports_root: Path | None
) -> None:
    if (
        not isinstance(path, Path)
        or not isinstance(trusted_reports_root, Path)
        or not trusted_reports_root.is_absolute()
        or trusted_reports_root.name != "reports"
        or path != trusted_reports_root / _FIXED_LEASE_NAME
    ):
        raise SourceLeaseError("Injected probe lease path is not a trusted reports root")


def _open_lease_parent(parent: Path) -> tuple[int, tuple[int, int, int | None, int | None]]:
    if not isinstance(parent, Path) or not parent.is_absolute():
        raise SourceLeaseError("Source lease parent is not an absolute path")
    descriptor = -1
    ancestors: tuple[int, ...] = ()
    retained = False
    try:
        if os.name == "nt":
            descriptor, ancestors, expected_device = safe_paths._nt_open_absolute_directory(
                parent
            )
            for ancestor in reversed(ancestors[:-1]):
                _close_lease_descriptor_quietly(ancestor)
            ancestors = ()
        else:
            descriptor, expected_device = safe_paths._posix_open_directory(parent)
        identity = _lease_parent_identity(descriptor)
        if identity[0] != expected_device:
            raise SourceLeaseError("Source lease parent crossed a device boundary")
        retained = True
        return descriptor, identity
    except (OSError, safe_paths.SafePathError) as error:
        raise SourceLeaseError("Source lease parent is unavailable or unsafe") from error
    except Exception:
        raise
    finally:
        for ancestor in reversed(ancestors):
            _close_lease_descriptor_quietly(ancestor)
        if descriptor >= 0 and not retained:
            _close_lease_descriptor_quietly(descriptor)


def _lease_parent_identity(
    descriptor: int,
) -> tuple[int, int, int | None, int | None]:
    before = safe_paths._fstat(None, descriptor)
    before_values = safe_paths._require_directory(before)
    after = safe_paths._fstat(None, descriptor)
    after_values = safe_paths._require_directory(after)
    if (
        before_values["disk"] is not True
        or after_values["disk"] is not True
        or any(
            before_values[key] != after_values[key] for key in ("device", "inode")
        )
    ):
        raise SourceLeaseError("Source lease parent changed while being opened")
    owner = getattr(before, "st_uid", None)
    mode = getattr(before, "st_mode", None)
    if owner is not None and type(owner) is not int:
        raise SourceLeaseError("Source lease parent owner is malformed")
    if mode is not None and type(mode) is not int:
        raise SourceLeaseError("Source lease parent mode is malformed")
    return (
        int(before_values["device"]),
        int(before_values["inode"]),
        owner,
        mode,
    )


def _assert_lease_parent_identity(
    descriptor: int, expected: tuple[int, int, int | None, int | None]
) -> None:
    if _lease_parent_identity(descriptor) != expected:
        raise SourceLeaseError("Source lease parent authority changed")


def _create_lease_relative(parent_descriptor: int, name: str) -> int:
    if name != _FIXED_LEASE_NAME:
        raise SourceLeaseError("Source lease name is not fixed")
    try:
        if os.name == "nt":
            return safe_paths._nt_open_relative(
                parent_descriptor,
                name,
                False,
                write=True,
                create_new=True,
            )
        if safe_paths.O_NOFOLLOW == 0:
            raise SourceLeaseError("Descriptor-relative lease create is unavailable")
        flags = (
            os.O_RDWR
            | os.O_CREAT
            | os.O_EXCL
            | safe_paths.O_NOFOLLOW
            | safe_paths._O_CLOEXEC
            | safe_paths._O_BINARY
        )
        return safe_paths._posix_openat(parent_descriptor, name, flags, 0o600)
    except (OSError, safe_paths.SafePathError) as error:
        raise SourceLeaseError("Source lease could not be acquired") from error


def _open_existing_lease_relative(parent_descriptor: int, name: str) -> int:
    if name != _FIXED_LEASE_NAME:
        raise SourceLeaseError("Source lease name is not fixed")
    try:
        if os.name == "nt":
            return safe_paths._nt_open_relative(parent_descriptor, name, False)
        if safe_paths.O_NOFOLLOW == 0:
            raise SourceLeaseError("Descriptor-relative lease inspection is unavailable")
        flags = (
            os.O_RDONLY
            | safe_paths.O_NOFOLLOW
            | safe_paths._O_CLOEXEC
            | safe_paths._O_BINARY
            | safe_paths._O_NONBLOCK
        )
        return safe_paths._posix_openat(parent_descriptor, name, flags)
    except (OSError, safe_paths.SafePathError) as error:
        raise SourceLeaseError("Source lease path is unavailable") from error


def _lease_file_identity(descriptor: int, parent_device: int) -> tuple[int, int, int]:
    try:
        before = safe_paths._fstat(None, descriptor)
        before_values = safe_paths._require_regular(before)
        if before_values["device"] != parent_device:
            raise SourceLeaseError("Source lease crossed a device boundary")
        after = safe_paths._fstat(None, descriptor)
        safe_paths._require_stable(before, after)
        return (
            int(before_values["device"]),
            int(before_values["inode"]),
            int(before_values["size"]),
        )
    except (OSError, safe_paths.SafePathError) as error:
        raise SourceLeaseError("Source lease descriptor is unsafe") from error


def _delete_owned_lease(
    parent_descriptor: int,
    name: str,
    descriptor: int,
    expected_identity: tuple[int, int, int],
    parent_identity: tuple[int, int, int | None, int | None],
) -> None:
    _assert_lease_parent_identity(parent_descriptor, parent_identity)
    if _lease_file_identity(descriptor, parent_identity[0]) != expected_identity:
        raise SourceLeaseError("Source lease descriptor changed before release")
    current_descriptor = -1
    try:
        current_descriptor = _open_existing_lease_relative(parent_descriptor, name)
        if (
            _lease_file_identity(current_descriptor, parent_identity[0])
            != expected_identity
        ):
            raise SourceLeaseError("Source lease path no longer names the acquired lease")
        if os.name == "nt":
            safe_paths._windows_delete_handle(descriptor)
        else:
            # POSIX has no portable unlink-by-descriptor primitive; cooperative
            # peers must not swap the verified fixed name before this unlinkat.
            _unlinkat_lease_name(parent_descriptor, name)
    except (OSError, safe_paths.SafePathError) as error:
        raise SourceLeaseError("Source lease could not be removed") from error
    finally:
        if current_descriptor >= 0:
            _close_lease_descriptor_quietly(current_descriptor)


def _discard_created_lease(
    parent_descriptor: int,
    name: str,
    descriptor: int,
    identity: tuple[int, int, int],
    parent_identity: tuple[int, int, int | None, int | None],
) -> None:
    try:
        _delete_owned_lease(
            parent_descriptor,
            name,
            descriptor,
            identity,
            parent_identity,
        )
    except Exception:
        pass


def _unlinkat_lease_name(parent_descriptor: int, name: str) -> None:
    if os.name == "nt" or name != _FIXED_LEASE_NAME:
        raise SourceLeaseError("Descriptor-relative lease removal is unavailable")
    try:
        library = ctypes.CDLL(None, use_errno=True)
        unlinkat = library.unlinkat
        unlinkat.argtypes = (ctypes.c_int, ctypes.c_char_p, ctypes.c_int)
        unlinkat.restype = ctypes.c_int
        if unlinkat(parent_descriptor, os.fsencode(name), 0) != 0:
            raise OSError(ctypes.get_errno(), "unlinkat failed", name)
    except OSError:
        raise
    except Exception as error:
        raise OSError("descriptor-relative lease removal failed") from error


def _close_lease_descriptor(descriptor: int) -> None:
    if descriptor < 0:
        return
    try:
        if os.name == "nt":
            close = ctypes.windll.kernel32.CloseHandle
            close.argtypes = (ctypes.c_void_p,)
            close.restype = ctypes.c_int
            if not close(ctypes.c_void_p(descriptor)):
                raise OSError(ctypes.get_last_error(), "CloseHandle failed")
        else:
            os.close(descriptor)
    except OSError:
        raise
    except Exception as error:
        raise OSError("Source lease descriptor close failed") from error


def _close_lease_descriptor_quietly(descriptor: int) -> None:
    try:
        _close_lease_descriptor(descriptor)
    except Exception:
        pass


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


def _canonical_tree_bytes(repo_root: Path, root: str) -> bytes:
    root_relative = _canonical_repo_relative(root, "Source root")
    rows: list[bytes] = []
    for kind, display_relative, full_relative in _walk_anchored_tree(
        repo_root, root_relative, "Source root"
    ):
        encoded_relative = display_relative.encode("utf-8")
        if kind == "D":
            rows.append(b"D\t" + encoded_relative + b"\n")
            continue
        rows.append(
            b"F\t"
            + encoded_relative
            + b"\t"
            + _sha256_prefixed(
                _read_regular_bytes(repo_root, full_relative)
            ).encode("ascii")
            + b"\n"
        )
    return b"".join(rows)


def _expand_canonical_pattern(repo_root: Path, pattern: str) -> tuple[str, ...]:
    if type(pattern) is not str or not pattern or pattern.startswith("/"):
        raise SourceLeaseError("Canonical build-input pattern is malformed")
    if "*" not in pattern:
        relative = _canonical_repo_relative(pattern, "Canonical build input")
        _assert_anchored_regular(repo_root, relative, "Canonical build input")
        return (relative,)
    if not pattern.endswith("/**") or pattern.count("*") != 2:
        raise SourceLeaseError("Canonical build-input wildcard is not closed")
    root_relative = _canonical_repo_relative(
        pattern.removesuffix("/**"), "Canonical build-input directory"
    )
    files = tuple(
        full_relative
        for kind, _display_relative, full_relative in _walk_anchored_tree(
            repo_root, root_relative, "Canonical build-input directory"
        )
        if kind == "F"
    )
    if not files:
        raise SourceLeaseError("Canonical build-input wildcard matched no files")
    return files


def _walk_anchored_tree(
    repo_root: Path, root_relative: str, label: str
) -> tuple[tuple[str, str, str], ...]:
    _assert_anchored_directory(repo_root, root_relative, label)
    entries: list[tuple[str, str, str]] = []

    def visit(parent_relative: str, display_parent: PurePosixPath) -> None:
        for name in _anchored_directory_entry_names(repo_root, parent_relative, label):
            child_relative = _canonical_repo_relative(
                f"{parent_relative}/{name}", label
            )
            display_relative = display_parent / name
            if _is_anchored_directory(repo_root, child_relative, label):
                entries.append(("D", display_relative.as_posix(), child_relative))
                visit(child_relative, display_relative)
                continue
            _assert_anchored_regular(repo_root, child_relative, label)
            entries.append(("F", display_relative.as_posix(), child_relative))

    visit(root_relative, PurePosixPath("."))
    return tuple(entries)


def _canonical_repo_relative(relative: str, label: str) -> str:
    if type(label) is not str or not label:
        raise SourceLeaseError("Source-path label is malformed")
    try:
        parts = safe_paths._relative_parts(relative, os.name == "nt")
    except safe_paths.SafePathError as error:
        raise SourceLeaseError(f"{label} is not a validated relative path") from error
    canonical = "/".join(parts)
    if canonical != relative or any(character in relative for character in "\x00\t\r\n"):
        raise SourceLeaseError(f"{label} is not a canonical relative path")
    return canonical


def _anchored_directory_entry_names(
    repo_root: Path, relative: str, label: str
) -> tuple[str, ...]:
    canonical = _canonical_repo_relative(relative, label)
    try:
        names = safe_paths.list_directory_relative(repo_root, canonical)
    except safe_paths.SafePathError as error:
        raise SourceLeaseError(f"{label} cannot be enumerated") from error
    for name in names:
        _canonical_repo_relative(f"{canonical}/{name}", label)
    return names


def _is_anchored_directory(repo_root: Path, relative: str, label: str) -> bool:
    try:
        _assert_anchored_directory(repo_root, relative, label)
    except SourceLeaseError:
        return False
    return True


def _assert_anchored_directory(repo_root: Path, relative: str, label: str) -> None:
    canonical = _canonical_repo_relative(relative, label)
    try:
        safe_paths.list_directory_relative(repo_root, canonical)
    except safe_paths.SafePathError as error:
        raise SourceLeaseError(f"{label} is not a safe directory") from error


def _assert_anchored_regular(repo_root: Path, relative: str, label: str) -> None:
    descriptor = -1
    try:
        descriptor = safe_paths.open_regular_relative(
            repo_root, _canonical_repo_relative(relative, label)
        )
        before = safe_paths._fstat(None, descriptor)
        safe_paths._require_regular(before)
        after = safe_paths._fstat(None, descriptor)
        safe_paths._require_stable(before, after)
    except (OSError, safe_paths.SafePathError) as error:
        raise SourceLeaseError(f"{label} is not a safe regular file") from error
    finally:
        if descriptor >= 0:
            safe_paths._close(None, descriptor)


def _read_regular_bytes(repo_root: Path, relative: str) -> bytes:
    descriptor = -1
    try:
        canonical = _canonical_repo_relative(relative, "Source file")
        descriptor = safe_paths.open_regular_relative(repo_root, canonical)
        before = safe_paths._fstat(None, descriptor)
        before_values = safe_paths._require_regular(before)
        size = int(before_values["size"])
        if not 1 <= size <= _MAX_FILE_BYTES:
            raise SourceLeaseError("Source file size is outside the bounded read limit")
        content = _read_descriptor_exact_bytes(descriptor, size)
        after = safe_paths._fstat(None, descriptor)
        safe_paths._require_stable(before, after)
        if len(content) != size:
            raise SourceLeaseError("Source file changed while being read")
        return content
    except (OSError, safe_paths.SafePathError) as error:
        raise SourceLeaseError("Source file could not be read") from error
    finally:
        if descriptor >= 0:
            safe_paths._close(None, descriptor)


def _read_descriptor_exact_bytes(descriptor: int, size: int) -> bytes:
    if os.name == "nt":
        return _read_windows_descriptor_bytes(descriptor, size)
    chunks: list[bytes] = []
    remaining = size
    while remaining:
        chunk = os.read(descriptor, min(65_536, remaining))
        if not chunk:
            raise SourceLeaseError("Source file ended before its verified size")
        chunks.append(chunk)
        remaining -= len(chunk)
    return b"".join(chunks)


def _read_windows_descriptor_bytes(descriptor: int, size: int) -> bytes:
    try:
        read = ctypes.windll.kernel32.ReadFile
        read.argtypes = (
            ctypes.c_void_p,
            ctypes.c_void_p,
            ctypes.c_uint32,
            ctypes.POINTER(ctypes.c_uint32),
            ctypes.c_void_p,
        )
        read.restype = ctypes.c_int
        chunks: list[bytes] = []
        remaining = size
        while remaining:
            limit = min(65_536, remaining)
            buffer = ctypes.create_string_buffer(limit)
            received = ctypes.c_uint32()
            if not read(
                ctypes.c_void_p(descriptor),
                buffer,
                limit,
                ctypes.byref(received),
                None,
            ) or received.value == 0:
                raise OSError(ctypes.get_last_error(), "ReadFile failed")
            chunks.append(buffer.raw[: int(received.value)])
            remaining -= int(received.value)
        return b"".join(chunks)
    except OSError:
        raise
    except Exception as error:
        raise OSError("Windows descriptor read failed") from error


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
