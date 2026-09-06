from __future__ import annotations

import ctypes
from copy import deepcopy
import hashlib
import io
import itertools
import json
import os
from dataclasses import dataclass, field
from pathlib import Path
from types import SimpleNamespace
from typing import Any

import pytest

from observers.backend_unit import (
    docker_adapter,
    lease,
    observe,
    publication,
    resources,
    root_handoff,
    safe_paths,
)


_PROJECT = "worktree-master"
_START = "2026-09-01T12:00:00.000000000Z"
_CONTAINER_ID = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
_HOSTNAME = _CONTAINER_ID[:12]
_OWNER_NONCE = "a1b2c3d4e5f60718293a4b5c6d7e8f90"
_RUN_ID = f"backend-unit-20260901T120000Z-{_HOSTNAME}-a1b2c3d4"
_PYTHON_IMAGE_ID = "sha256:" + "c" * 64
_BACKEND_IMAGE_ID = "sha256:" + "d" * 64
_FROZEN_SUBJECTS = {
    "backend/internal/calls/call_durable_credential_allocation.go": (
        "sha256:be6ff058468c4e288e0b0d7060fd17cf3118229d009f28dc056d89f281e99d4d"
    ),
    "backend/tests/calls/call_durable_credential_allocation_test.go": (
        "sha256:0ae525c86a75369bdb5790f26b1308f42574fb545883328da24011d4150b03a9"
    ),
}
_SOURCE_ROOT_DIGESTS = {
    "backend/cmd": "sha256:" + "1" * 64,
    "backend/internal": "sha256:" + "2" * 64,
    "backend/tests": "sha256:" + "3" * 64,
    "backend/ent/schema": "sha256:" + "4" * 64,
    "backend/migrations": "sha256:" + "5" * 64,
    "backend/testframework": "sha256:" + "6" * 64,
}
_FIXTURE_PATHS = (
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
    "testframework/observers/backend_unit/root_handoff.py",
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
    *_FIXTURE_PATHS,
)
_MANIFEST_CONTENTS = {
    path: f"frozen-content-{ordinal:03d}\n".encode()
    for ordinal, path in enumerate(_MANIFEST_PATHS, start=1)
}


def _manifest_bytes(contents: dict[str, bytes]) -> bytes:
    assert tuple(contents) == _MANIFEST_PATHS
    return b"".join(
        path.encode()
        + b"\tsha256:"
        + hashlib.sha256(contents[path]).hexdigest().encode()
        + b"\n"
        for path in _MANIFEST_PATHS
    )


_MANIFEST_BYTES = _manifest_bytes(_MANIFEST_CONTENTS)
_MANIFEST_SHA256 = "sha256:" + hashlib.sha256(_MANIFEST_BYTES).hexdigest()
_ROOT_HANDOFF_SHA256 = "sha256:" + hashlib.sha256(
    _MANIFEST_CONTENTS["testframework/observers/backend_unit/root_handoff.py"]
).hexdigest()
_SEALED_CLASSIFICATIONS = (
    "SEALED_EXECUTABLE_PASS",
    "SEALED_EXECUTABLE_FAIL_DISCOVERY",
    "SEALED_EXECUTABLE_FAIL_STAGE",
    "SEALED_TIMEOUT_UNTRIAGED",
    "SEALED_INFRA_ABORTED_OOM",
    "SEALED_INFRA_ABORTED_RESOURCE_CONTRACT",
    "SEALED_INFRA_ABORTED_REPORT_BUDGET",
    "SEALED_INFRA_ABORTED_TMPFS",
    "SEALED_INFRA_ABORTED_HOST_RAM_FLOOR",
    "SEALED_INFRA_ABORTED_DISK_PRESSURE",
    "SEALED_INFRA_ABORTED_WHOLE_RUN_TIMEOUT",
    "SEALED_INTERRUPTED",
    "SEALED_INVALID_QA_OVERLAP",
    "SEALED_INVALID_SOURCE_MUTATED",
    "SEALED_INVALID_PROVENANCE",
)


@dataclass(frozen=True)
class _Invocation:
    argv: tuple[str, ...]
    shell: bool
    timeout_seconds: int
    output_cap_bytes: int
    stream: bool


class _EventStream:
    def __init__(
        self, lines: tuple[bytes, ...], *, terminate_raises: bool = False
    ) -> None:
        self._lines = lines
        self._next_line = 0
        self._terminate_raises = terminate_raises
        self.terminated = False
        self.reaped = False

    def __iter__(self) -> Any:
        return iter(self._lines)

    def poll_line(self, timeout_seconds: int) -> bytes | None:
        if type(timeout_seconds) is not int or timeout_seconds != 5:
            raise AssertionError("event polling must use the frozen timeout")
        if self._next_line == len(self._lines):
            return None
        line = self._lines[self._next_line]
        self._next_line += 1
        return line

    def terminate(self) -> None:
        self.terminated = True
        if self._terminate_raises:
            raise OSError("event termination failed")

    def reap(self) -> None:
        self.reaped = True


@dataclass
class _DockerRunner:
    one_shot: dict[tuple[str, ...], bytes]
    stream: _EventStream = field(default_factory=lambda: _EventStream(()))
    calls: list[_Invocation] = field(default_factory=list)

    def __call__(
        self,
        argv: tuple[str, ...],
        *,
        shell: bool,
        timeout_seconds: int,
        output_cap_bytes: int,
        stream: bool,
    ) -> bytes | _EventStream:
        self.calls.append(
            _Invocation(argv, shell, timeout_seconds, output_cap_bytes, stream)
        )
        if stream:
            return self.stream
        return self.one_shot[argv]


class _BoundedStdout:
    def __init__(self, payload: bytes) -> None:
        self._payload = payload
        self._offset = 0
        self.read_sizes: list[int] = []

    def read(self, size: int) -> bytes:
        self.read_sizes.append(size)
        result = self._payload[self._offset : self._offset + size]
        self._offset += len(result)
        return result


class _BoundedProcess:
    def __init__(self, payload: bytes) -> None:
        self.stdout = _BoundedStdout(payload)
        self.running = True
        self.terminated = False
        self.killed = False
        self.wait_calls: list[float] = []

    def poll(self) -> None | int:
        return None if self.running else 0

    def terminate(self) -> None:
        self.terminated = True
        self.running = False

    def kill(self) -> None:
        self.killed = True
        self.running = False

    def wait(self, timeout: float) -> int:
        self.wait_calls.append(timeout)
        self.running = False
        return 0


@dataclass
class _PathOps:
    fail_open: bool = False
    hard_link: bool = False
    nonregular: bool = False
    reparse: bool = False
    root_reparse: bool = False
    intermediate_reparse: bool = False
    descendant_cross_device: bool = False
    size_changed: bool = False
    identity_changed: bool = False
    existing_output: bool = False
    cross_directory_rename: bool = False
    windows: bool = False
    swap_intermediate_name: bool = False
    root_open_calls: list[Path] = field(default_factory=list)
    nt_absolute_directory_calls: list[Path] = field(default_factory=list)
    openat_calls: list[tuple[int, str, int]] = field(default_factory=list)
    nt_relative_calls: list[tuple[int, str, bool]] = field(default_factory=list)
    parent_open_calls: list[Path] = field(default_factory=list)
    write_new_at_calls: list[tuple[int, str, bytes]] = field(default_factory=list)
    destination_checks: list[Path] = field(default_factory=list)
    renameat_calls: list[tuple[int, str, int, str]] = field(default_factory=list)
    closes: list[int] = field(default_factory=list)
    fstat_descriptors: list[int] = field(default_factory=list)
    descriptor_metadata: dict[int, dict[str, int | bool]] = field(default_factory=dict)
    _fstat_calls: dict[int, int] = field(default_factory=dict)
    _next_descriptor: int = 17

    def _new_descriptor(self, *, directory: bool, role: str) -> int:
        descriptor = self._next_descriptor
        self._next_descriptor += 1
        reparse = (
            self.root_reparse
            if role == "root"
            else self.intermediate_reparse
            if role == "intermediate"
            else self.reparse
            if role == "leaf"
            else False
        )
        self.descriptor_metadata[descriptor] = {
            "device": 8
            if self.descendant_cross_device and role in {"intermediate", "leaf"}
            else 7,
            "inode": descriptor,
            "size": 0,
            "regular": not directory and not self.nonregular,
            "nlink": 2 if not directory and self.hard_link else 1,
            "reparse": reparse,
            "directory": directory,
            "disk": True,
        }
        return descriptor

    def open_root(self, path: Path) -> int:
        self.root_open_calls.append(path)
        return self._new_descriptor(directory=True, role="root")

    def nt_open_absolute_directory(self, path: Path) -> int:
        self.nt_absolute_directory_calls.append(path)
        return self._new_descriptor(directory=True, role="root")

    def openat(self, directory_fd: int, name: str, flags: int) -> int:
        self.openat_calls.append((directory_fd, name, flags))
        if self.fail_open:
            raise OSError("unsafe path")
        if self.swap_intermediate_name and len(self.openat_calls) == 1:
            self.swap_intermediate_name = False
        directory = bool(flags & safe_paths._O_DIRECTORY)
        return self._new_descriptor(
            directory=directory,
            role="intermediate" if directory else "leaf",
        )

    def nt_open_relative(self, parent_fd: int, name: str, directory: bool) -> int:
        self.nt_relative_calls.append((parent_fd, name, directory))
        if self.fail_open:
            raise OSError("unsafe reparse point")
        return self._new_descriptor(
            directory=directory,
            role="intermediate" if directory else "leaf",
        )

    def fstat(self, descriptor: int) -> dict[str, int | bool]:
        self.fstat_descriptors.append(descriptor)
        self._fstat_calls[descriptor] = self._fstat_calls.get(descriptor, 0) + 1
        snapshot = dict(self.descriptor_metadata[descriptor])
        if not bool(snapshot["directory"]) and self._fstat_calls[descriptor] > 1:
            if self.identity_changed:
                snapshot["device"] = 8
            if self.size_changed and self._fstat_calls[descriptor] > 2:
                snapshot["size"] = 1
        return snapshot

    def open_parent(self, path: Path) -> int:
        self.parent_open_calls.append(path)
        return self._new_descriptor(directory=True, role="parent")

    def destination_exists(self, path: Path) -> bool:
        self.destination_checks.append(path)
        return self.existing_output

    def write_new_at(self, parent_fd: int, destination: str, payload: bytes) -> Path:
        self.write_new_at_calls.append((parent_fd, destination, payload))
        return Path(f".{destination}.test.tmp")

    def renameat(
        self, source_parent_fd: int, source: str, destination_parent_fd: int, destination: str
    ) -> None:
        self.renameat_calls.append(
            (source_parent_fd, source, destination_parent_fd, destination)
        )
        if self.cross_directory_rename or source_parent_fd != destination_parent_fd:
            raise OSError("cross-directory rename")

    def close(self, descriptor: int) -> None:
        self.closes.append(descriptor)


@dataclass
class _LeaseProbe:
    dirty: bool = False
    head: str = "b" * 40
    manifest_bytes: bytes = _MANIFEST_BYTES
    subject_digest: str = "sha256:" + "f" * 64
    source_root_digest_map: dict[str, str] = field(
        default_factory=lambda: dict(_SOURCE_ROOT_DIGESTS)
    )
    commands: list[tuple[str, ...]] = field(default_factory=list)
    acl_changes: list[str] = field(default_factory=list)

    def git(self, argv: tuple[str, ...]) -> str:
        self.commands.append(argv)
        if argv == ("git", "rev-parse", "HEAD"):
            return self.head + "\n"
        if argv == ("git", "status", "--porcelain=v1", "--untracked-files=all"):
            return " M hostile.txt\n" if self.dirty else ""
        raise AssertionError(f"unexpected Git argv: {argv!r}")

    def canonical_build_input_manifest(self) -> bytes:
        return self.manifest_bytes

    def frozen_subject_digests(self) -> dict[str, str]:
        return {
            "backend/internal/calls/call_durable_credential_allocation.go": self.subject_digest,
            "backend/tests/calls/call_durable_credential_allocation_test.go": (
                "sha256:0ae525c86a75369bdb5790f26b1308f42574fb545883328da24011d4150b03a9"
            ),
        }

    def source_root_digests(self) -> dict[str, str]:
        return dict(self.source_root_digest_map)


def _injected_lease_path(tmp_path: Path) -> tuple[Path, Path]:
    reports_root = tmp_path / "reports"
    reports_root.mkdir()
    return reports_root / ".backend-unit-source-mutation.lease", reports_root


class _NtDirectoryQuery:
    def __init__(self, responses: list[tuple[int, bytes, int | None]]) -> None:
        self.responses = responses
        self.descriptors: list[int] = []

    def __call__(self, *arguments: object) -> int:
        descriptor = arguments[0]
        io_status = arguments[4]
        buffer = arguments[5]
        maximum = arguments[6]
        assert isinstance(descriptor, ctypes.c_void_p)
        assert hasattr(io_status, "_obj")
        assert type(maximum) is int
        self.descriptors.append(int(descriptor.value))
        status, payload, information = self.responses.pop(0)
        io_status._obj.Information = len(payload) if information is None else information
        if payload:
            ctypes.memmove(buffer, payload, min(len(payload), maximum))
        return status


def _nt_directory_record(
    name: str,
    *,
    next_offset: int = 0,
    attributes: int = 0,
) -> bytes:
    record_type = type(
        "_TestFileIdBothDirectoryInformation",
        (ctypes.Structure,),
        {
            "_fields_": [
                ("NextEntryOffset", ctypes.c_uint32),
                ("FileIndex", ctypes.c_uint32),
                ("CreationTime", ctypes.c_int64),
                ("LastAccessTime", ctypes.c_int64),
                ("LastWriteTime", ctypes.c_int64),
                ("ChangeTime", ctypes.c_int64),
                ("EndOfFile", ctypes.c_int64),
                ("AllocationSize", ctypes.c_int64),
                ("FileAttributes", ctypes.c_uint32),
                ("FileNameLength", ctypes.c_uint32),
                ("EaSize", ctypes.c_uint32),
                ("ShortNameLength", ctypes.c_ubyte),
                ("ShortName", ctypes.c_uint16 * 12),
                ("FileId", ctypes.c_int64),
            ]
        },
    )
    encoded = name.encode("utf-16-le")
    record = record_type()
    record.NextEntryOffset = next_offset
    record.FileAttributes = attributes
    record.FileNameLength = len(encoded)
    return bytes(record) + encoded


@dataclass
class _PublicationWriter:
    events: list[tuple[str, str]] = field(default_factory=list)
    payloads: dict[str, bytes] = field(default_factory=dict)
    path_payloads: dict[Path, bytes] = field(default_factory=dict)
    opaque_file_identities: dict[
        Path, tuple[int, int, int, int, int, int, int]
    ] = field(default_factory=dict)
    read_counts: dict[Path, int] = field(default_factory=dict)
    existing_final: bool = False
    active_container_writer: bool = False
    cross_directory: bool = False
    read_swaps: dict[str, bytes] = field(default_factory=dict)
    identity_replacements: dict[
        str, tuple[int, int, int, int, int, int, int]
    ] = field(default_factory=dict)
    identity_swap_after_reads: dict[str, int] = field(default_factory=dict)
    remove_fails: bool = False
    identity_events: list[tuple[str, str]] = field(default_factory=list)
    identity_checks: list[
        tuple[str, str, tuple[int, int, int, int, int, int, int]]
    ] = field(default_factory=list)
    foreign_final: bytes = b"foreign-final-receipt"
    foreign_final_path: Path = Path("/reports/backend-unit-run/final-receipt.json")
    touched_paths: list[Path] = field(default_factory=list)
    terminal_markers: tuple[str, ...] = ('CONTAINER_COMPLETE',)
    terminal_marker_payloads: dict[str, bytes] = field(default_factory=dict)
    terminal_marker_run_dir: Path = Path('/reports/backend-unit-run.partial')
    terminal_marker_list_fails: bool = False
    terminal_marker_path_failures: frozenset[str] = frozenset()
    terminal_marker_mutation_after_lists: int | None = None
    terminal_marker_replacement: tuple[str, ...] = ()
    terminal_marker_list_count: int = 0
    terminal_marker_events: list[tuple[str, str]] = field(default_factory=list)
    tree_entries_fail: bool = False
    tree_mutation_after_lists: int | None = None
    tree_replacement: tuple[str, ...] | None = None
    tree_extra_entry: str | None = None
    tree_list_count: int = 0
    tree_entry_events: list[tuple[str, str]] = field(default_factory=list)
    payload_replacements: dict[str, bytes] = field(default_factory=dict)
    payload_swap_after_reads: dict[str, int] = field(default_factory=dict)
    _next_opaque_identity: int = 0

    def __post_init__(self) -> None:
        if self.existing_final:
            self.path_payloads[self.foreign_final_path] = self.foreign_final
            self._assign_opaque_identity(self.foreign_final_path)
        for marker in self.terminal_markers:
            path = self.terminal_marker_run_dir / marker
            if path not in self.path_payloads:
                self.path_payloads[path] = self.terminal_marker_payloads.get(marker, b'')
                self._assign_opaque_identity(path)
        self._materialize_canonical_container_inventory()

    def _materialize_canonical_container_inventory(self) -> None:
        """Expose every declared base artifact through the descriptor-like seam."""

        container_payload = self.payloads.get('container-receipt.json')
        if container_payload is None:
            return
        container_path = self.terminal_marker_run_dir / 'container-receipt.json'
        if container_path not in self.path_payloads:
            self.path_payloads[container_path] = container_payload
        if container_path not in self.opaque_file_identities:
            self._assign_opaque_identity(container_path)
        document = json.loads(container_payload)
        records = document.get('artifact_inventory')
        if type(records) is not list:
            raise AssertionError('canonical container fixture lacks artifact inventory')
        fixtures = {
            'CONTAINER_ABORTED': b'',
            'CONTAINER_COMPLETE': b'',
            'owner.json': b'{}',
            'stage-build.json': b'{}',
        }
        for record in records:
            if type(record) is not dict:
                raise AssertionError('canonical container fixture has malformed inventory')
            relative = record.get('path')
            if type(relative) is not str or relative not in fixtures:
                raise AssertionError(f'fixture bytes are missing for {relative!r}')
            payload = fixtures[relative]
            if record.get('sha256') != _sha256(payload) or record.get('size') != len(payload):
                raise AssertionError(f'fixture bytes do not match inventory for {relative}')
            path = self.terminal_marker_run_dir / relative
            if path not in self.path_payloads:
                self.path_payloads[path] = payload
                self._assign_opaque_identity(path)

    def _assign_opaque_identity(
        self, path: Path
    ) -> tuple[int, int, int, int, int, int, int]:
        self._next_opaque_identity += 1
        identity = (
            7,
            self._next_opaque_identity,
            0o100600,
            1,
            len(self.path_payloads[path]),
            1_000_000 + self._next_opaque_identity,
            2_000_000 + self._next_opaque_identity,
        )
        self.opaque_file_identities[path] = identity
        return identity

    def _payload_path(self, run_dir: Path, relative: str) -> Path:
        path = run_dir / relative
        if path not in self.path_payloads:
            try:
                self.path_payloads[path] = self.payloads[relative]
            except KeyError as error:
                raise FileNotFoundError(relative) from error
            self._assign_opaque_identity(path)
        return path

    def _maybe_replace_opaque_identity(self, path: Path, relative: str) -> None:
        read_count = self.read_counts.get(path, 0)
        replacement = self.identity_replacements.get(relative)
        after_reads = self.identity_swap_after_reads.get(relative)
        if replacement is not None and after_reads is not None and read_count >= after_reads:
            self.opaque_file_identities[path] = replacement
        payload_replacement = self.payload_replacements.get(relative)
        payload_after_reads = self.payload_swap_after_reads.get(relative)
        if (
            payload_replacement is not None
            and payload_after_reads is not None
            and read_count >= payload_after_reads
        ):
            self.path_payloads[path] = payload_replacement
        self.read_counts[path] = read_count + 1

    def assert_terminal_and_stable(self, run_dir: Path) -> None:
        self.events.append(("assert_terminal", run_dir.name))
        if self.active_container_writer:
            raise OSError("container writer is still active")

    def assert_absent(self, final_dir: Path) -> None:
        self.events.append(("assert_absent", final_dir.name))
        if self.existing_final or any(
            path.parent == final_dir for path in self.path_payloads
        ):
            raise FileExistsError("final directory already exists")

    def list_terminal_markers(self, run_dir: Path) -> tuple[str, ...]:
        self.terminal_marker_events.append(('list_terminal_markers', run_dir.name))
        if self.terminal_marker_list_fails:
            raise OSError('terminal marker listing failed')
        if (
            self.terminal_marker_mutation_after_lists is not None
            and self.terminal_marker_list_count
            >= self.terminal_marker_mutation_after_lists
        ):
            markers = self.terminal_marker_replacement
        else:
            markers = self.terminal_markers
        self.terminal_marker_list_count += 1
        return markers

    def list_entries(self, run_dir: Path) -> tuple[str, ...]:
        """Read-only anchored enumeration of the current partial run tree."""

        self.events.append(('list_entries', run_dir.name))
        self.tree_entry_events.append(('list_entries', run_dir.name))
        if self.tree_entries_fail:
            raise OSError('run tree listing failed')
        entries = tuple(
            sorted(path.name for path in self.path_payloads if path.parent == run_dir)
        )
        if (
            self.tree_mutation_after_lists is not None
            and self.tree_list_count >= self.tree_mutation_after_lists
        ):
            if self.tree_replacement is not None:
                entries = self.tree_replacement
            elif self.tree_extra_entry is not None:
                entries = (*entries, self.tree_extra_entry)
            else:
                raise AssertionError('tree mutation must provide its replacement')
        self.tree_list_count += 1
        return entries

    def read(
        self, run_dir: Path, relative: str
    ) -> tuple[bytes, tuple[int, int, int, int, int, int, int]]:
        if relative in self.terminal_marker_path_failures:
            raise OSError('terminal marker path is unavailable')
        if relative in ('CONTAINER_COMPLETE', 'CONTAINER_ABORTED'):
            path = self._payload_path(run_dir, relative)
            self.touched_paths.append(path)
            self._maybe_replace_opaque_identity(path, relative)
            identity = self.opaque_file_identities[path]
            self.identity_checks.append(('read', relative, identity))
            if relative in self.read_swaps:
                return self.read_swaps[relative], identity
            return self.path_payloads[path], identity
        path = self._payload_path(run_dir, relative)
        self.touched_paths.append(path)
        self.events.append(("read", relative))
        self.identity_events.append(("read", relative))
        self._maybe_replace_opaque_identity(path, relative)
        identity = self.opaque_file_identities[path]
        self.identity_checks.append(("read", relative, identity))
        if relative in self.read_swaps:
            return self.read_swaps[relative], identity
        return self.path_payloads[path], identity

    def write(self, path: Path, payload: bytes) -> None:
        self.touched_paths.append(path)
        if path in self.path_payloads:
            raise FileExistsError(path.name)
        self.path_payloads[path] = payload
        self._assign_opaque_identity(path)
        self.payloads[path.name] = payload
        self.events.append(("write", path.name))
        self.identity_events.append(("write", path.name))

    def remove(self, path: Path) -> bool:
        self.touched_paths.append(path)
        self.events.append(("remove", path.name))
        self.identity_events.append(("remove", path.name))
        if self.remove_fails:
            return False
        self.path_payloads.pop(path, None)
        self.opaque_file_identities.pop(path, None)
        self.read_counts.pop(path, None)
        self.payloads.pop(path.name, None)
        return True

    def rename_same_directory(self, partial: Path, final: Path) -> None:
        self.events.append(("rename", f"{partial.name}->{final.name}"))
        self.identity_events.append(("rename", f"{partial.name}->{final.name}"))
        if self.cross_directory:
            raise OSError("cross-directory publication")
        if any(path.parent == final for path in self.path_payloads):
            raise FileExistsError("final directory already exists")
        for source in tuple(
            path for path in self.path_payloads if path.parent == partial
        ):
            destination = final / source.name
            self.path_payloads[destination] = self.path_payloads.pop(source)
            self.opaque_file_identities[destination] = self.opaque_file_identities.pop(
                source
            )
            if source in self.read_counts:
                self.read_counts[destination] = self.read_counts.pop(source)


@dataclass
class _RawObserverRuntime:
    raw_observations: dict[str, Any] = field(default_factory=dict)
    candidate: dict[str, Any] | None = None
    terminal_error: bool = False
    interrupt_before_bind: bool = False
    source_drift_after_lease: bool = False
    python_version: tuple[int, int, int] = (3, 12, 3)
    stdlib_only: bool = True
    events: list[str] = field(default_factory=list)
    writes: list[str] = field(default_factory=list)
    stop_calls: list[tuple[str, ...]] = field(default_factory=list)
    published: tuple[dict[str, Any], dict[str, Any]] | None = None

    def bootstrap(self, argv: tuple[str, ...]) -> None:
        self.events.append("bootstrap")
        assert argv == ("--compose-project", _PROJECT)
        assert self.python_version[:2] == (3, 12)
        assert self.stdlib_only is True

    def preflight(self) -> tuple[dict[str, int], dict[str, int], dict[str, int]]:
        self.events.append("preflight")
        return (_safe_sample(), _safe_sample(), _safe_sample())

    def acquire_lease(self) -> None:
        self.events.append("lease.acquire")

    def snapshot(self, phase: str) -> dict[str, Any]:
        self.events.append(f"snapshot.{phase}")
        snapshot = _snapshot()
        if phase == "after" and self.source_drift_after_lease:
            snapshot["head"] = "c" * 40
        return snapshot

    def fixed_images(self) -> dict[str, str]:
        self.events.append("images")
        return {"python_contract": _PYTHON_IMAGE_ID, "backend_unit": _BACKEND_IMAGE_ID}

    def canonical_manifest(self) -> tuple[bytes, str]:
        self.events.append("manifest")
        return _MANIFEST_BYTES, _MANIFEST_SHA256

    def ready(self, session: str) -> None:
        self.events.append("ready")
        assert len(session) == 32 and set(session) <= set("0123456789abcdef")

    def bind_candidate(self, _images: dict[str, str]) -> dict[str, Any]:
        self.events.append("bind")
        if self.interrupt_before_bind:
            raise KeyboardInterrupt
        return _candidate() if self.candidate is None else self.candidate

    def monitor(self, _candidate: dict[str, Any]) -> tuple[dict[str, int], ...]:
        self.events.append("monitor")
        return (_safe_sample(), _safe_sample())

    def raw_observations_for(
        self, _candidate: dict[str, Any], _samples: tuple[dict[str, int], ...]
    ) -> dict[str, Any]:
        self.events.append("raw")
        return dict(self.raw_observations)

    def stop_bound_once(self, container_id: str) -> dict[str, Any]:
        argv = ("docker", "stop", "--timeout", "5", container_id)
        self.events.append("stop")
        self.stop_calls.append(argv)
        if len(self.stop_calls) != 1:
            raise AssertionError("a guarded stop must be single-use")
        return {"result": "stopped", "argv": argv}

    def terminal_verify(self, _candidate: dict[str, Any]) -> None:
        self.events.append("terminal_verify")
        if self.terminal_error:
            raise OSError("unsealable evidence")

    def publish(self, final: dict[str, Any], observer: dict[str, Any]) -> None:
        self.events.append("publish")
        self.writes.extend(("final-receipt.json", "observer-receipt.json", "SEAL"))
        self.published = (final, observer)

    def release_lease(self) -> None:
        self.events.append("lease.release")


def _safe_sample() -> dict[str, int]:
    return {
        "free_ram_bytes": 2 * 1024**3,
        "free_c_bytes": 250 * 1024**3,
        "PercentDiskTime": 1,
        "CurrentDiskQueueLength": 0,
    }


def _snapshot() -> dict[str, Any]:
    return {
        "head": "b" * 40,
        "status": "",
        "canonical_build_input_manifest_bytes": _MANIFEST_BYTES,
        "canonical_build_input_manifest_sha256": _MANIFEST_SHA256,
        "fixed_image_ids": {
            "python_contract": _PYTHON_IMAGE_ID,
            "backend_unit": _BACKEND_IMAGE_ID,
        },
        "frozen_subjects": _LeaseProbe().frozen_subject_digests(),
    }


def _candidate(**changes: object) -> dict[str, Any]:
    candidate: dict[str, Any] = {
        "container_id": _CONTAINER_ID,
        "started_at": _START,
        "image_id": _BACKEND_IMAGE_ID,
        "container_image": _BACKEND_IMAGE_ID,
        "labels": {
            "com.docker.compose.project": _PROJECT,
            "com.docker.compose.service": "backend-unit",
        },
        "owner_marker": {
            "run_id": _RUN_ID,
            "owner_nonce": _OWNER_NONCE,
            "container_hostname": _HOSTNAME,
        },
    }
    candidate.update(changes)
    return candidate


def _docker_responses() -> dict[tuple[str, ...], bytes]:
    return {
        (
            "docker",
            "inspect",
            "--type",
            "container",
            "--format",
            "{{json .}}",
            _CONTAINER_ID,
        ): json.dumps(
            {
                "Id": _CONTAINER_ID,
                "Image": _BACKEND_IMAGE_ID,
                "Config": {"Hostname": _HOSTNAME},
                "ConfigLabels": {
                    "com.docker.compose.project": _PROJECT,
                    "com.docker.compose.service": "backend-unit",
                },
            }
        ).encode(),
        ("docker", "ps", "--no-trunc", "--format", "{{json .}}"): b"[]\n",
        (
            "docker",
            "stats",
            "--no-stream",
            "--format",
            "{{json .}}",
            _CONTAINER_ID,
        ): b"{}\n",
        (
            "docker",
            "image",
            "inspect",
            "--format",
            "{{json .}}",
            "worktree-master-python-contract",
        ): json.dumps({"Id": _PYTHON_IMAGE_ID}).encode(),
        (
            "docker",
            "image",
            "inspect",
            "--format",
            "{{json .}}",
            "worktree-master-backend-unit",
        ): json.dumps({"Id": _BACKEND_IMAGE_ID}).encode(),
        ("docker", "stop", "--timeout", "5", _CONTAINER_ID): b"backend-unit\n",
    }


def _events_argv() -> tuple[str, ...]:
    return (
        "docker",
        "events",
        "--since",
        _START,
        "--filter",
        "type=container",
        "--filter",
        "event=create",
        "--filter",
        "event=start",
        "--filter",
        "event=die",
        "--filter",
        "event=destroy",
        "--format",
        "{{json .}}",
    )


def _canonical_receipt_bytes(document: dict[str, Any]) -> bytes:
    return (
        json.dumps(
            document,
            allow_nan=False,
            ensure_ascii=False,
            separators=(",", ":"),
            sort_keys=True,
        ).encode("utf-8")
        + b"\n"
    )


def _sha256(payload: bytes) -> str:
    return "sha256:" + hashlib.sha256(payload).hexdigest()


def _inventory_rows(payloads: dict[str, bytes]) -> tuple[list[dict[str, Any]], str]:
    rows = [
        {"path": path, "sha256": _sha256(payload), "size": len(payload)}
        for path, payload in sorted(payloads.items())
    ]
    inventory_bytes = b"".join(
        f"{row['path']}\t{row['sha256']}\t{row['size']}\n".encode("ascii")
        for row in rows
    )
    return rows, _sha256(inventory_bytes)


def _source_snapshot() -> dict[str, Any]:
    return {
        "head": "b" * 40,
        "status": "",
        "source_root_sha256": dict(_SOURCE_ROOT_DIGESTS),
        "frozen_subjects": dict(_FROZEN_SUBJECTS),
        "canonical_build_input_manifest_sha256": _MANIFEST_SHA256,
    }


def _root_handoff_snapshot(
    *,
    status: str = "",
    manifest_sha256: str = _MANIFEST_SHA256,
) -> dict[str, Any]:
    snapshot = _source_snapshot()
    snapshot["status"] = status
    snapshot["canonical_build_input_manifest_sha256"] = manifest_sha256
    return snapshot


def _root_handoff_document(
    *,
    prebuild: dict[str, Any] | None = None,
    postgreen: dict[str, Any] | None = None,
    manifest: bytes = _MANIFEST_BYTES,
    manifest_sha256: str = _MANIFEST_SHA256,
    fixed_image_ids: dict[str, str] | None = None,
) -> dict[str, Any]:
    prebuild_snapshot = (
        _root_handoff_snapshot() if prebuild is None else deepcopy(prebuild)
    )
    postgreen_snapshot = (
        deepcopy(prebuild_snapshot) if postgreen is None else deepcopy(postgreen)
    )
    return {
        "schema": "hub.qa.backend-unit-root-handoff/v1",
        "target": "backend-unit",
        "prebuild": prebuild_snapshot,
        "postgreen": postgreen_snapshot,
        "canonical_build_input_manifest_bytes": manifest.decode("utf-8"),
        "canonical_build_input_manifest_sha256": manifest_sha256,
        "fixed_image_ids": (
            {
                "python_contract": _PYTHON_IMAGE_ID,
                "backend_unit": _BACKEND_IMAGE_ID,
            }
            if fixed_image_ids is None
            else dict(fixed_image_ids)
        ),
    }


def _root_handoff_provisional(
    *,
    prebuild: dict[str, Any] | None = None,
    manifest: bytes = _MANIFEST_BYTES,
    manifest_sha256: str = _MANIFEST_SHA256,
) -> dict[str, Any]:
    return {
        "prebuild": _root_handoff_snapshot() if prebuild is None else deepcopy(prebuild),
        "canonical_build_input_manifest_bytes": manifest.decode("utf-8"),
        "canonical_build_input_manifest_sha256": manifest_sha256,
    }


@dataclass
class _RootHandoffLease:
    snapshots: tuple[dict[str, Any], ...]
    manifest: bytes = _MANIFEST_BYTES
    fixed_images: dict[str, str] = field(
        default_factory=lambda: {
            "python_contract": _PYTHON_IMAGE_ID,
            "backend_unit": _BACKEND_IMAGE_ID,
        }
    )
    events: list[str] = field(default_factory=list)
    snapshot_index: int = 0

    def snapshot(self) -> dict[str, Any]:
        self.events.append("snapshot")
        index = min(self.snapshot_index, len(self.snapshots) - 1)
        self.snapshot_index += 1
        return deepcopy(self.snapshots[index])

    def canonical_build_input_manifest_bytes(self) -> bytes:
        self.events.append("manifest")
        return self.manifest

    def fixed_image_ids(self) -> dict[str, str]:
        self.events.append("images")
        return dict(self.fixed_images)

    def verify_unchanged(
        self, prebuild: dict[str, Any], postgreen: dict[str, Any]
    ) -> None:
        self.events.append("verify")
        if prebuild != postgreen:
            raise RuntimeError("source changed")

    def release(self) -> None:
        self.events.append("release")


def _publication_payloads() -> tuple[bytes, dict[str, Any], dict[str, Any]]:
    stage_digest = {
        "build": "sha256:" + "7" * 64,
        "vet": "sha256:" + "8" * 64,
        "unit": "sha256:" + "9" * 64,
        "race": "sha256:" + "a" * 64,
    }
    streams = {
        stage: {
            "stdout": {"cap_bytes": stdout_cap, "used_bytes": 0, "truncated": False},
            "stderr": {"cap_bytes": stderr_cap, "used_bytes": 0, "truncated": False},
        }
        for stage, (stdout_cap, stderr_cap) in {
            "build": (2 * 1024 * 1024, 2 * 1024 * 1024),
            "vet": (2 * 1024 * 1024, 2 * 1024 * 1024),
            "unit": (16 * 1024 * 1024, 4 * 1024 * 1024),
            "race": (24 * 1024 * 1024, 4 * 1024 * 1024),
        }.items()
    }
    container_inventory, container_inventory_sha256 = _inventory_rows(
        {
            "CONTAINER_COMPLETE": b"",
            "owner.json": b"{}",
            "stage-build.json": b"{}",
        }
    )
    cgroup_snapshot = {
        "memory.max": "2147483648",
        "memory.swap.max": "0",
        "memory.events": {"oom": 0, "oom_kill": 0},
    }
    container = {
        "schema": "hub.qa.backend-unit-container-receipt/v1",
        "target": "backend-unit",
        "run_id": _RUN_ID,
        "owner_nonce": _OWNER_NONCE,
        "container_hostname": _HOSTNAME,
        "launcher_sha256": "sha256:" + "0" * 64,
        "command_profile_sha256": "sha256:" + "1" * 64,
        "source_manifest_before_sha256": "sha256:" + "2" * 64,
        "source_manifest_after_sha256": "sha256:" + "2" * 64,
        "cgroup": {
            "family": "v2",
            "baseline": cgroup_snapshot,
            "final": deepcopy(cgroup_snapshot),
        },
        "go_environment": {
            "go_version": "go version go1.25.0 linux/amd64",
            "GOFLAGS": "-p=1",
            "GOENV": "off",
            "GOWORK": "off",
            "GOTOOLCHAIN": "local",
            "GOPROXY": "off",
            "GOMAXPROCS": None,
            "CGO_ENABLED": "0",
            "GOOS": "linux",
            "GOARCH": "amd64",
            "GOCACHE": "/var/cache/hub-go",
            "GOTMPDIR": "/var/tmp/hub-go",
            "runtime_affecting_environment": [
                {"name": name, "present": False, "value": None}
                for name in ("GODEBUG", "GOEXPERIMENT", "GOMEMLIMIT")
            ],
        },
        "package_plan_sha256": stage_digest,
        "stage_receipt_sha256": dict(stage_digest),
        "artifact_inventory": container_inventory,
        "artifact_inventory_sha256": container_inventory_sha256,
        "report_budget": {
            "total_bytes": 64 * 1024 * 1024,
            "reserved_metadata_bytes": 8 * 1024 * 1024,
            "metadata_used_bytes": 0,
            "streams": streams,
            "total_used_bytes": 0,
            "truncated_due_to_budget": False,
            "raw_output_partial": False,
        },
        "terminal": {"signal": None, "exit_code": 0},
        "state": "CONTAINER_COMPLETE",
    }
    container_payload = _canonical_receipt_bytes(container)
    final = {
        "schema": "hub.qa.backend-unit-final-receipt/v1",
        "target": "backend-unit",
        "run_id": _RUN_ID,
        "owner_nonce": _OWNER_NONCE,
        "container_receipt_sha256": _sha256(container_payload),
        "container_id": _CONTAINER_ID,
        "backend_image_id": _BACKEND_IMAGE_ID,
        "canonical_build_input_manifest_sha256": _MANIFEST_SHA256,
        "classification": "SEALED_EXECUTABLE_PASS",
        "immutable_outcome": None,
        "seal_timestamp": "2026-09-01T18:00:00Z",
    }
    final_payload = _canonical_receipt_bytes(final)
    artifact_inventory, artifact_inventory_sha256 = _inventory_rows(
        {
            "container-receipt.json": container_payload,
            "final-receipt.json": final_payload,
        }
    )
    snapshot = _source_snapshot()
    observer = {
        "schema": "hub.qa.backend-unit-observer-receipt/v1",
        "target": "backend-unit",
        "run_id": _RUN_ID,
        "owner_nonce": _OWNER_NONCE,
        "container_id": _CONTAINER_ID,
        "container_image": _BACKEND_IMAGE_ID,
        "container_hostname": _HOSTNAME,
        "container_receipt_sha256": _sha256(container_payload),
        "compose_sha256": _sha256(_MANIFEST_CONTENTS["docker-compose.local.yml"]),
        "git": {
            "head_before": "b" * 40,
            "head_after": "b" * 40,
            "status_before": "",
            "status_after": "",
        },
        "frozen_subjects": dict(_FROZEN_SUBJECTS),
        "source_snapshots": {
            "prebuild": deepcopy(snapshot),
            "postgreen": deepcopy(snapshot),
            "observer_snapshot": deepcopy(snapshot),
        },
        "root_handoff_sha256": _ROOT_HANDOFF_SHA256,
        "concurrency_observation_scope": "docker-ps-visible-containers-only",
        "canonical_build_input_manifest_bytes": _MANIFEST_BYTES.decode("utf-8"),
        "canonical_build_input_manifest_sha256": _MANIFEST_SHA256,
        "fixed_image_ids": {
            "python_contract": _PYTHON_IMAGE_ID,
            "backend_unit": _BACKEND_IMAGE_ID,
        },
        "bound_backend_image_id": _BACKEND_IMAGE_ID,
        "host_config": {
            "memory": 2 * 1024 * 1024 * 1024,
            "memory_swap": 2 * 1024 * 1024 * 1024,
            "nano_cpus": 6_000_000_000,
            "read_only_rootfs": True,
            "security_opt": ["no-new-privileges:true"],
            "cap_drop": ["ALL"],
            "pids_limit": 256,
            "init": True,
            "network_mode": "worktree-master_qa-internal",
            "mounts": [
                {
                    "type": mount_type,
                    "source": source,
                    "destination": destination,
                    "read_only": read_only,
                    "size_bytes": size_bytes,
                }
                for mount_type, source, destination, read_only, size_bytes in (
                    ("bind", "./backend/cmd", "/opt/hub/source/backend/cmd", True, None),
                    ("bind", "./backend/internal", "/opt/hub/source/backend/internal", True, None),
                    ("bind", "./backend/tests", "/opt/hub/source/backend/tests", True, None),
                    ("bind", "./backend/ent/schema", "/opt/hub/source/backend/ent/schema", True, None),
                    ("bind", "./backend/migrations", "/opt/hub/source/backend/migrations", True, None),
                    ("bind", "./backend/testframework", "/opt/hub/source/backend/testframework", True, None),
                    ("bind", "./testframework/reports", "/reports", False, None),
                    ("tmpfs", None, "/tmp", False, 64 * 1024 * 1024),
                    ("tmpfs", None, "/var/cache/hub-go", False, 768 * 1024 * 1024),
                    ("tmpfs", None, "/var/tmp/hub-go", False, 384 * 1024 * 1024),
                )
            ],
        },
        "resource_samples": [
            {
                "phase": "preflight",
                "timestamp": _START,
                "free_ram_bytes": 8 * 1024 * 1024 * 1024,
                "free_c_bytes": 250 * 1024 * 1024 * 1024,
                "percent_disk_time": 0,
                "disk_queue_length": 0,
            }
        ],
        "concurrency_samples": [
            {
                "phase": "preflight",
                "timestamp": _START,
                "backend_unit_count": 0,
                "mail_calendar_qa_test_build_count": 0,
                "buildkit_state": "idle",
            }
        ],
        "lifecycle_events": [],
        "classification_facts": {
            "source_mutated": False,
            "invalid_provenance": False,
            "oom": False,
            "resource": False,
            "budget": False,
            "tmpfs": False,
            "timeout": False,
            "discovery_failure": False,
            "stage_failure": False,
            "predicates": {
                "preflight_ram_floor_met": True,
                "preflight_c_free_floor_met": True,
                "preflight_disk_pressure_2_of_3": False,
                "two_consecutive_ram_below_2_gib": False,
                "rolling_disk_pressure_2_of_3": False,
                "prohibited_backend_unit_overlap": False,
                "prohibited_mail_calendar_qa_test_build_overlap": False,
                "buildkit_idle": True,
                "whole_run_deadline_expired": False,
            },
            "explicit_stops": [],
            "selected_stop": None,
            "classification": "SEALED_EXECUTABLE_PASS",
            "immutable_outcome": None,
        },
        "classification": "SEALED_EXECUTABLE_PASS",
        "immutable_outcome": None,
        "selected_stop": None,
        "stop_argv": None,
        "guarded_stop": {"predicate": None, "argv": None, "result": None},
        "artifact_inventory": artifact_inventory,
        "artifact_inventory_sha256": artifact_inventory_sha256,
        "cleanup": {"container_removed": True},
        "final_receipt": final,
        "seal_timestamp": "2026-09-01T18:00:00Z",
    }
    return container_payload, final, observer


def _aborted_publication_payloads() -> tuple[bytes, dict[str, Any], dict[str, Any]]:
    complete_payload, final, observer = _publication_payloads()
    complete = json.loads(complete_payload)
    assert type(complete) is dict
    common_fields = (
        'schema', 'target', 'run_id', 'owner_nonce', 'container_hostname',
        'launcher_sha256', 'command_profile_sha256',
        'source_manifest_before_sha256', 'source_manifest_after_sha256',
        'cgroup', 'go_environment',
    )
    container = {field: deepcopy(complete[field]) for field in common_fields}
    container['abort'] = {
        'signal': 'TERM',
        'phase': 'bootstrap',
        'active_stage': None,
        'planned_stages': [],
        'completed_stages': [],
        'child_group_reaped': True,
        'stdout_reader_reaped': True,
        'stderr_reader_reaped': True,
        'unpublished_stage_rolled_back': True,
    }
    inventory, inventory_sha256 = _inventory_rows(
        {'CONTAINER_ABORTED': b'', 'owner.json': b'{}'}
    )
    container['artifact_inventory'] = inventory
    container['artifact_inventory_sha256'] = inventory_sha256
    container['report_budget'] = deepcopy(complete['report_budget'])
    container['terminal'] = {'signal': 'TERM', 'exit_code': 143}
    container['state'] = 'CONTAINER_ABORTED'
    budget = container['report_budget']
    assert type(budget) is dict
    for _ in range(16):
        container_payload = _canonical_receipt_bytes(container)
        metadata_bytes = sum(record['size'] for record in inventory) + len(
            container_payload
        )
        if (
            budget['metadata_used_bytes'] == metadata_bytes
            and budget['total_used_bytes'] == metadata_bytes
        ):
            break
        budget['metadata_used_bytes'] = metadata_bytes
        budget['total_used_bytes'] = metadata_bytes
    else:
        raise AssertionError('aborted receipt budget fixed point did not converge')

    final['container_receipt_sha256'] = _sha256(container_payload)
    observer['container_receipt_sha256'] = _sha256(container_payload)
    observer['final_receipt'] = deepcopy(final)
    final_payload = _canonical_receipt_bytes(final)
    observer['artifact_inventory'], observer['artifact_inventory_sha256'] = _inventory_rows(
        {'container-receipt.json': container_payload, 'final-receipt.json': final_payload}
    )
    return container_payload, final, observer


def _policy(raw_observations: dict[str, Any]) -> dict[str, Any]:
    return {
        "sealed_classifications": _SEALED_CLASSIFICATIONS,
        "classification_facts": raw_observations,
    }


def test_canonical_manifest_is_exact_and_excludes_second_observer_review_file() -> None:
    manifest_paths = tuple(
        line.split(b"\t", 1)[0].decode() for line in _MANIFEST_BYTES.splitlines()
    )

    assert manifest_paths == _MANIFEST_PATHS
    assert len(manifest_paths) == len(set(manifest_paths))
    assert "testframework/selftests/architecture/observers/test_backend_unit_observer.py" in manifest_paths
    assert (
        "testframework/selftests/architecture/observers/"
        "test_backend_unit_observer_review_contract.py"
    ) not in manifest_paths
    assert set(_FIXTURE_PATHS) <= set(manifest_paths)
    for line, path in zip(_MANIFEST_BYTES.splitlines(), _MANIFEST_PATHS, strict=True):
        _, digest = line.split(b"\tsha256:", 1)
        assert digest.decode() == hashlib.sha256(_MANIFEST_CONTENTS[path]).hexdigest()
        assert digest.decode() != hashlib.sha256(path.encode()).hexdigest()

    changed_contents = dict(_MANIFEST_CONTENTS)
    changed_contents[_MANIFEST_PATHS[0]] += b"changed"
    assert _manifest_bytes(changed_contents) != _MANIFEST_BYTES
    assert hashlib.sha256(_MANIFEST_BYTES).hexdigest() == _MANIFEST_SHA256.removeprefix(
        "sha256:"
    )


def test_docker_adapter_has_only_literal_read_argvs_and_fixed_images() -> None:
    event = json.dumps(
        {
            "id": _CONTAINER_ID,
            "status": "create",
            "timeNano": 1_788_264_000_000_000_000,
            "Actor": {
                "Attributes": {
                    "com.docker.compose.project": _PROJECT,
                    "com.docker.compose.service": "backend-unit",
                }
            },
        }
    ).encode() + b"\n"
    stream = _EventStream((event,))
    adapter, runner = _recoverable_bound_adapter(stream)

    assert adapter.bind_unique_candidate_since(_START, _BACKEND_IMAGE_ID) == _CONTAINER_ID
    session = adapter.open_bound_events(_START)
    assert session.poll() == json.loads(event)
    session.close()
    assert stream.terminated is True
    assert stream.reaped is True
    with pytest.raises(docker_adapter.DockerReadError):
        adapter.open_bound_events(_START)
    assert adapter.inspect(_CONTAINER_ID)["Id"] == _CONTAINER_ID
    assert adapter.stats(_CONTAINER_ID) == {}
    assert adapter.inspect_fixed_images() == {
        "python_contract": _PYTHON_IMAGE_ID,
        "backend_unit": _BACKEND_IMAGE_ID,
    }
    assert tuple(call.argv for call in runner.calls) == (
        ("docker", "ps", "--no-trunc", "--format", "{{json .}}"),
        (
            "docker",
            "inspect",
            "--type",
            "container",
            "--format",
            "{{json .}}",
            _CONTAINER_ID,
        ),
        _events_argv(),
        ("docker", "inspect", "--type", "container", "--format", "{{json .}}", _CONTAINER_ID),
        ("docker", "stats", "--no-stream", "--format", "{{json .}}", _CONTAINER_ID),
        ("docker", "image", "inspect", "--format", "{{json .}}", "worktree-master-python-contract"),
        ("docker", "image", "inspect", "--format", "{{json .}}", "worktree-master-backend-unit"),
    )
    assert all(call.shell is False for call in runner.calls)
    assert all(call.timeout_seconds == 10 for call in runner.calls)
    assert all(call.output_cap_bytes == 1024 * 1024 for call in runner.calls)
    assert [call.stream for call in runner.calls] == [False, False, True, False, False, False, False]
    forbidden = {"compose", "run", "create", "start", "build", "pull", "prune", "rm"}
    assert not any(token in forbidden for call in runner.calls for token in call.argv)


def test_docker_adapter_rejects_invalid_since_dynamic_images_and_extra_argv() -> None:
    for invalid_since in ("2026-09-01T12:00:00Z", "not-a-timestamp", 1, None):
        runner = _DockerRunner(_docker_responses())
        with pytest.raises(docker_adapter.DockerReadError):
            docker_adapter.DockerReadAdapter(
                _PROJECT, runner=runner
            ).open_bound_events(  # type: ignore[arg-type]
                invalid_since
            )
        assert runner.calls == []

    for invalid_image in ("registry.invalid/backend-unit", "sha256:" + "D" * 64):
        runner = _DockerRunner(_docker_responses())
        with pytest.raises(docker_adapter.DockerReadError):
            docker_adapter.DockerReadAdapter(_PROJECT, runner=runner).bind_unique_candidate_since(
                _START, invalid_image
            )
        assert runner.calls == []

    runner = _DockerRunner(_docker_responses())
    adapter = docker_adapter.DockerReadAdapter(_PROJECT, runner=runner)
    with pytest.raises(TypeError):
        adapter.inspect_fixed_images("registry.invalid/backend-unit")  # type: ignore[call-arg]
    with pytest.raises(TypeError):
        adapter.inspect(_CONTAINER_ID, "--all")  # type: ignore[call-arg]
    assert runner.calls == []

    for forbidden_project in ("WORKTREE-master", "worktree-master;docker", "../worktree-master"):
        with pytest.raises((ValueError, docker_adapter.DockerReadError)):
            docker_adapter.DockerReadAdapter(
                forbidden_project, runner=_DockerRunner(_docker_responses())
            )


def test_docker_adapter_refuses_unbound_and_differently_bound_valid_container_reads() -> None:
    adapter, runner = _recoverable_bound_adapter()
    for method in (adapter.inspect, adapter.stats):
        with pytest.raises(docker_adapter.DockerReadError):
            method(_CONTAINER_ID)
    assert runner.calls == []

    assert adapter.bind_unique_candidate_since(_START, _BACKEND_IMAGE_ID) == _CONTAINER_ID
    calls_after_valid_bind = tuple(runner.calls)
    foreign_valid_id = "f" * 64
    for method in (adapter.inspect, adapter.stats):
        with pytest.raises(docker_adapter.DockerReadError):
            method(foreign_valid_id)
    assert tuple(runner.calls) == calls_after_valid_bind


@pytest.mark.parametrize(
    "payload",
    (
        b'{"duplicate":1,"duplicate":2}\n',
        b'{"number":NaN}\n',
        b'{"record":1} {"trailing":2}\n',
    ),
)
@pytest.mark.parametrize("path", ("events", "ps", "inspect", "stats", "fixed-image"))
def test_docker_adapter_rejects_duplicate_nonfinite_and_trailing_json_every_read_path(
    path: str,
    payload: bytes,
) -> None:
    responses = _docker_responses()
    if path == "events":
        stream = _EventStream((payload,))
        adapter, runner = _recoverable_bound_adapter(stream)
        assert adapter.bind_unique_candidate_since(_START, _BACKEND_IMAGE_ID) == _CONTAINER_ID
        session = adapter.open_bound_events(_START)
        with pytest.raises(docker_adapter.InvalidEvidence):
            session.poll()
        assert stream.terminated is True
        assert stream.reaped is True
        for unavailable in (
            adapter.recoverable_bound_container_id,
            lambda: adapter.stop_bound_once(_CONTAINER_ID),
            lambda: adapter.open_bound_events(_START),
        ):
            with pytest.raises(docker_adapter.DockerReadError):
                unavailable()
        assert tuple(call.argv for call in runner.calls) == (
            ("docker", "ps", "--no-trunc", "--format", "{{json .}}"),
            (
                "docker",
                "inspect",
                "--type",
                "container",
                "--format",
                "{{json .}}",
                _CONTAINER_ID,
            ),
            _events_argv(),
        )
        return
    elif path == "ps":
        responses[("docker", "ps", "--no-trunc", "--format", "{{json .}}")] = payload
        runner = _DockerRunner(responses)
        adapter = docker_adapter.DockerReadAdapter(_PROJECT, runner=runner)
        invoke = lambda: adapter.bind_unique_candidate_since(_START, _BACKEND_IMAGE_ID)
        expected_calls = 1
    elif path == "inspect":
        adapter, runner = _recoverable_bound_adapter()
        assert adapter.bind_unique_candidate_since(_START, _BACKEND_IMAGE_ID) == _CONTAINER_ID
        runner.one_shot[
            ("docker", "inspect", "--type", "container", "--format", "{{json .}}", _CONTAINER_ID)
        ] = payload
        invoke = lambda: adapter.inspect(_CONTAINER_ID)
        expected_calls = 3
    elif path == "stats":
        adapter, runner = _recoverable_bound_adapter()
        assert adapter.bind_unique_candidate_since(_START, _BACKEND_IMAGE_ID) == _CONTAINER_ID
        runner.one_shot[
            ("docker", "stats", "--no-stream", "--format", "{{json .}}", _CONTAINER_ID)
        ] = payload
        invoke = lambda: adapter.stats(_CONTAINER_ID)
        expected_calls = 3
    else:
        responses[("docker", "image", "inspect", "--format", "{{json .}}", "worktree-master-python-contract")] = payload
        runner = _DockerRunner(responses)
        adapter = docker_adapter.DockerReadAdapter(_PROJECT, runner=runner)
        invoke = adapter.inspect_fixed_images
        expected_calls = 1

    with pytest.raises(docker_adapter.DockerReadError):
        invoke()
    assert len(runner.calls) == expected_calls


def test_event_cleanup_reaps_when_terminate_raises() -> None:
    event = json.dumps(
        {
            "id": _CONTAINER_ID,
            "status": "create",
            "timeNano": 1_788_264_000_000_000_000,
            "Actor": {
                "Attributes": {
                    "com.docker.compose.project": _PROJECT,
                    "com.docker.compose.service": "backend-unit",
                }
            },
        }
    ).encode() + b"\n"
    stream = _EventStream((event,), terminate_raises=True)
    adapter, runner = _recoverable_bound_adapter(stream)
    assert adapter.bind_unique_candidate_since(_START, _BACKEND_IMAGE_ID) == _CONTAINER_ID
    session = adapter.open_bound_events(_START)
    assert session.poll() == json.loads(event)
    with pytest.raises(docker_adapter.InvalidEvidence):
        session.close()
    assert stream.terminated is True
    assert stream.reaped is True
    assert [call.argv for call in runner.calls] == [
        ("docker", "ps", "--no-trunc", "--format", "{{json .}}"),
        ("docker", "inspect", "--type", "container", "--format", "{{json .}}", _CONTAINER_ID),
        _events_argv(),
    ]
    for unavailable in (adapter.recoverable_bound_container_id, lambda: adapter.stop_bound_once(_CONTAINER_ID)):
        with pytest.raises(docker_adapter.DockerReadError):
            unavailable()


@pytest.mark.parametrize(("size", "overflows"), ((3, False), (4, False), (5, True)))
def test_bounded_subprocess_output_probes_only_cap_plus_one_without_real_subprocess(
    size: int,
    overflows: bool,
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    cap = 4
    process = _BoundedProcess(b"x" * size)

    def fake_popen(argv: tuple[str, ...], **kwargs: object) -> _BoundedProcess:
        assert argv == ("docker", "version")
        assert kwargs["shell"] is False
        return process

    def fake_read_pipe_chunk(
        owned_process: _BoundedProcess, timeout_seconds: float, max_bytes: int
    ) -> bytes:
        assert owned_process is process
        assert max_bytes > 0
        assert 0 < timeout_seconds <= 1
        chunk = owned_process.stdout.read(max_bytes)
        if chunk == b"":
            owned_process.running = False
        return chunk

    monkeypatch.setattr(docker_adapter.subprocess, "Popen", fake_popen)
    monkeypatch.setattr(docker_adapter, "_read_pipe_chunk", fake_read_pipe_chunk)
    if overflows:
        with pytest.raises(docker_adapter.DockerReadError):
            docker_adapter._bounded_subprocess_output(
                ("docker", "version"), timeout_seconds=1, output_cap_bytes=cap
            )
        assert process.terminated is True
    else:
        assert docker_adapter._bounded_subprocess_output(
            ("docker", "version"), timeout_seconds=1, output_cap_bytes=cap
        ) == b"x" * size
        assert process.terminated is False
    assert process.killed is False
    assert process.stdout.read_sizes[0] == cap + 1


@pytest.mark.parametrize("cap", ("records", "bytes"))
def test_event_stream_caps_terminate_reap_and_refuse_to_seal(cap: str) -> None:
    event: dict[str, object] = {
        "id": _CONTAINER_ID,
        "status": "create",
        "timeNano": 1_788_264_000_000_000_000,
        "Actor": {
            "Attributes": {
                "com.docker.compose.project": _PROJECT,
                "com.docker.compose.service": "backend-unit",
            }
        },
    }
    if cap == "records":
        stream = _EventStream(tuple(json.dumps(event).encode() + b"\n" for _ in range(17)))
        successful_polls = 16
    else:
        event["padding"] = "x" * (64 * 1024)
        stream = _EventStream((json.dumps(event).encode() + b"\n",))
        successful_polls = 0

    adapter, runner = _recoverable_bound_adapter(stream)
    assert adapter.bind_unique_candidate_since(_START, _BACKEND_IMAGE_ID) == _CONTAINER_ID
    session = adapter.open_bound_events(_START)
    for _ in range(successful_polls):
        assert session.poll() == event
    with pytest.raises(docker_adapter.InvalidEvidence):
        session.poll()
    assert stream.terminated is True
    assert stream.reaped is True
    assert [call.argv for call in runner.calls] == [
        ("docker", "ps", "--no-trunc", "--format", "{{json .}}"),
        ("docker", "inspect", "--type", "container", "--format", "{{json .}}", _CONTAINER_ID),
        _events_argv(),
    ]
    for unavailable in (
        adapter.recoverable_bound_container_id,
        lambda: adapter.stop_bound_once(_CONTAINER_ID),
        lambda: adapter.open_bound_events(_START),
    ):
        with pytest.raises(docker_adapter.DockerReadError):
            unavailable()


def test_bound_ids_and_stop_are_full_canonical_single_use_authority() -> None:
    adapter, runner = _recoverable_bound_adapter()
    assert adapter.bind_unique_candidate_since(_START, _BACKEND_IMAGE_ID) == _CONTAINER_ID
    calls_after_valid_bind = tuple(runner.calls)
    for invalid_id in ("", _CONTAINER_ID[:-1], _CONTAINER_ID.upper(), "f" * 63 + "g"):
        for method in (adapter.inspect, adapter.stats, adapter.stop_bound_once):
            with pytest.raises(docker_adapter.DockerReadError):
                method(invalid_id)
        assert tuple(runner.calls) == calls_after_valid_bind

    assert adapter.stop_bound_once(_CONTAINER_ID) == {"result": "stopped"}
    with pytest.raises(docker_adapter.DockerReadError):
        adapter.stop_bound_once(_CONTAINER_ID)
    assert [call.argv for call in runner.calls] == [
        ("docker", "ps", "--no-trunc", "--format", "{{json .}}"),
        ("docker", "inspect", "--type", "container", "--format", "{{json .}}", _CONTAINER_ID),
        ("docker", "stop", "--timeout", "5", _CONTAINER_ID),
    ]


def _recoverable_bound_adapter(
    stream: _EventStream | None = None,
) -> tuple[docker_adapter.DockerReadAdapter, _DockerRunner]:
    responses = _docker_responses()
    responses[("docker", "ps", "--no-trunc", "--format", "{{json .}}")] = (
        json.dumps(
            {
                "ID": _CONTAINER_ID,
                "Image": _BACKEND_IMAGE_ID,
                "Labels": {
                    "com.docker.compose.project": _PROJECT,
                    "com.docker.compose.service": "backend-unit",
                },
            }
        ).encode()
        + b"\n"
    )
    responses[
        (
            "docker",
            "inspect",
            "--type",
            "container",
            "--format",
            "{{json .}}",
            _CONTAINER_ID,
        )
    ] = json.dumps(
        {
            "Id": _CONTAINER_ID,
            "Image": _BACKEND_IMAGE_ID,
            "Config": {"Hostname": _HOSTNAME},
            "ConfigLabels": {
                "com.docker.compose.project": _PROJECT,
                "com.docker.compose.service": "backend-unit",
            },
            "State": {"StartedAt": "2026-09-01T12:00:01.000000000Z"},
        }
    ).encode()
    runner = _DockerRunner(responses, _EventStream(()) if stream is None else stream)
    return docker_adapter.DockerReadAdapter(_PROJECT, runner=runner), runner


def test_recoverable_bound_container_id_is_none_without_binding_or_runner_io() -> None:
    runner = _DockerRunner(_docker_responses())
    adapter = docker_adapter.DockerReadAdapter(_PROJECT, runner=runner)

    assert adapter.recoverable_bound_container_id() is None
    assert runner.calls == []


def test_recoverable_bound_container_id_returns_normal_unique_bind_without_extra_io(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    adapter, runner = _recoverable_bound_adapter()
    bind_time_calls = 0
    original_require_bind_time = adapter._require_bind_time

    def observe_bind_time(deadline: float) -> None:
        nonlocal bind_time_calls
        bind_time_calls += 1
        original_require_bind_time(deadline)

    monkeypatch.setattr(adapter, "_require_bind_time", observe_bind_time)

    assert adapter.bind_unique_candidate_since(_START, _BACKEND_IMAGE_ID) == _CONTAINER_ID
    assert bind_time_calls == 6
    calls_before_accessor = tuple(runner.calls)
    assert adapter.recoverable_bound_container_id() == _CONTAINER_ID
    assert tuple(runner.calls) == calls_before_accessor


def test_recoverable_bound_container_id_survives_final_bind_interrupt_and_stops_once(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    adapter, runner = _recoverable_bound_adapter()
    bind_time_calls = 0

    def interrupt_after_retained_bind(_deadline: float) -> None:
        nonlocal bind_time_calls
        bind_time_calls += 1
        if bind_time_calls == 6:
            assert adapter._bound_container_id == _CONTAINER_ID
            assert adapter._bound_backend_image_id == _BACKEND_IMAGE_ID
            raise KeyboardInterrupt

    monkeypatch.setattr(adapter, "_require_bind_time", interrupt_after_retained_bind)

    with pytest.raises(KeyboardInterrupt):
        adapter.bind_unique_candidate_since(_START, _BACKEND_IMAGE_ID)

    assert bind_time_calls == 6
    assert adapter.recoverable_bound_container_id() == _CONTAINER_ID
    assert adapter.stop_bound_once(_CONTAINER_ID) == {"result": "stopped"}
    assert [call.argv for call in runner.calls] == [
        ("docker", "ps", "--no-trunc", "--format", "{{json .}}"),
        ("docker", "inspect", "--type", "container", "--format", "{{json .}}", _CONTAINER_ID),
        ("docker", "stop", "--timeout", "5", _CONTAINER_ID),
    ]


def test_recoverable_bound_container_id_is_none_after_prebind_interrupt_without_stop(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    adapter, runner = _recoverable_bound_adapter()

    def interrupt_before_retained_bind(_deadline: float) -> None:
        raise KeyboardInterrupt

    monkeypatch.setattr(adapter, "_require_bind_time", interrupt_before_retained_bind)

    with pytest.raises(KeyboardInterrupt):
        adapter.bind_unique_candidate_since(_START, _BACKEND_IMAGE_ID)

    assert adapter.recoverable_bound_container_id() is None
    with pytest.raises(docker_adapter.DockerReadError):
        adapter.stop_bound_once(_CONTAINER_ID)
    assert runner.calls == []


def test_recoverable_bound_container_id_rejects_failed_and_malformed_retained_state() -> None:
    failed_responses = _docker_responses()
    failed_responses[("docker", "ps", "--no-trunc", "--format", "{{json .}}")] = (
        json.dumps(
            {
                "ID": _CONTAINER_ID[:-1],
                "Image": _BACKEND_IMAGE_ID,
                "Labels": {
                    "com.docker.compose.project": _PROJECT,
                    "com.docker.compose.service": "backend-unit",
                },
            }
        ).encode()
        + b"\n"
    )
    failed_runner = _DockerRunner(failed_responses)
    failed_adapter = docker_adapter.DockerReadAdapter(_PROJECT, runner=failed_runner)
    with pytest.raises(docker_adapter.InvalidEvidence):
        failed_adapter.bind_unique_candidate_since(_START, _BACKEND_IMAGE_ID)
    with pytest.raises(docker_adapter.DockerReadError):
        failed_adapter.recoverable_bound_container_id()
    assert [call.argv for call in failed_runner.calls] == [
        ("docker", "ps", "--no-trunc", "--format", "{{json .}}"),
    ]

    for container_id, image_id in (
        (None, _BACKEND_IMAGE_ID),
        (_CONTAINER_ID, None),
        (1, _BACKEND_IMAGE_ID),
        (_CONTAINER_ID[:-1], _BACKEND_IMAGE_ID),
        (_CONTAINER_ID, "sha256:" + "e" * 63),
    ):
        malformed_runner = _DockerRunner(_docker_responses())
        malformed_adapter = docker_adapter.DockerReadAdapter(_PROJECT, runner=malformed_runner)
        malformed_adapter._bound_container_id = container_id
        malformed_adapter._bound_backend_image_id = image_id
        with pytest.raises(docker_adapter.DockerReadError):
            malformed_adapter.recoverable_bound_container_id()
        assert malformed_runner.calls == []


def test_repeated_recoverable_accessor_does_not_consume_single_stop_authority() -> None:
    adapter, runner = _recoverable_bound_adapter()
    assert adapter.bind_unique_candidate_since(_START, _BACKEND_IMAGE_ID) == _CONTAINER_ID
    calls_before_accessors = tuple(runner.calls)

    assert adapter.recoverable_bound_container_id() == _CONTAINER_ID
    assert adapter.recoverable_bound_container_id() == _CONTAINER_ID
    assert tuple(runner.calls) == calls_before_accessors
    assert adapter.stop_bound_once(_CONTAINER_ID) == {"result": "stopped"}
    with pytest.raises(docker_adapter.DockerReadError):
        adapter.stop_bound_once(_CONTAINER_ID)
    assert [call.argv for call in runner.calls] == [
        ("docker", "ps", "--no-trunc", "--format", "{{json .}}"),
        ("docker", "inspect", "--type", "container", "--format", "{{json .}}", _CONTAINER_ID),
        ("docker", "stop", "--timeout", "5", _CONTAINER_ID),
    ]


@pytest.mark.parametrize(
    "changes",
    (
        {"container_id": _CONTAINER_ID[:-1]},
        {"started_at": "2026-09-01T11:59:59.999999999Z"},
        {"started_at": "2026-09-01T12:02:00.000000000Z"},
        {"started_at": "2026-09-01T12:02:00.000000001Z"},
        {"image_id": "sha256:" + "e" * 64},
        {"container_image": "sha256:" + "e" * 64},
        {"labels": {}},
        {"owner_marker": {"run_id": "foreign-run", "owner_nonce": _OWNER_NONCE, "container_hostname": _HOSTNAME}},
        {"owner_marker": {"run_id": _RUN_ID, "owner_nonce": "0" * 32, "container_hostname": _HOSTNAME}},
        {"owner_marker": {"run_id": _RUN_ID, "owner_nonce": _OWNER_NONCE, "container_hostname": "fedcba987654"}},
    ),
)
def test_candidate_binding_rejects_every_foreign_identity_and_window_boundary(
    changes: dict[str, object],
) -> None:
    with pytest.raises(observe.InvalidEvidence, match="INVALID_EVIDENCE"):
        observe.bind_unique_candidate(
            (_candidate(**changes),), _START, 120, _PROJECT, "backend-unit", _BACKEND_IMAGE_ID
        )


def test_candidate_binding_requires_hostname_exact_prefix_and_open_window() -> None:
    accepted = _candidate(started_at="2026-09-01T12:01:59.999999999Z")
    bound = observe.bind_unique_candidate(
        (accepted,), _START, 120, _PROJECT, "backend-unit", _BACKEND_IMAGE_ID
    )
    assert bound["container_id"] == _CONTAINER_ID
    assert bound["container_hostname"] == _CONTAINER_ID[:12]
    assert _CONTAINER_ID.startswith(bound["container_hostname"])
    assert bound["run_id"] == _RUN_ID
    assert bound["owner_nonce"] == _OWNER_NONCE
    assert bound["backend_image_id"] == _BACKEND_IMAGE_ID

    nonprefix_lower_hex = _candidate(
        owner_marker={
            "run_id": _RUN_ID,
            "owner_nonce": _OWNER_NONCE,
            "container_hostname": "fedcba987654",
        }
    )
    with pytest.raises(observe.InvalidEvidence, match="INVALID_EVIDENCE"):
        observe.bind_unique_candidate(
            (nonprefix_lower_hex,), _START, 120, _PROJECT, "backend-unit", _BACKEND_IMAGE_ID
        )


def test_candidate_binding_rejects_zero_or_two_individually_valid_candidates() -> None:
    other_id = "f" * 64
    other = _candidate(
        container_id=other_id,
        owner_marker={"run_id": _RUN_ID, "owner_nonce": _OWNER_NONCE, "container_hostname": other_id[:12]},
    )
    for candidates in ((), (_candidate(), other)):
        with pytest.raises(observe.InvalidEvidence, match="INVALID_EVIDENCE"):
            observe.bind_unique_candidate(
                candidates, _START, 120, _PROJECT, "backend-unit", _BACKEND_IMAGE_ID
            )


@pytest.mark.parametrize(
    "candidate",
    (
        _candidate(labels={}),
        _candidate(container_id=_CONTAINER_ID[:-1]),
        _candidate(owner_marker={"run_id": "foreign-run", "owner_nonce": _OWNER_NONCE, "container_hostname": _HOSTNAME}),
        _candidate(owner_marker={"run_id": _RUN_ID, "owner_nonce": "0" * 32, "container_hostname": _HOSTNAME}),
        _candidate(owner_marker={"run_id": _RUN_ID, "owner_nonce": _OWNER_NONCE, "container_hostname": "fedcba987654"}),
        _candidate(image_id="sha256:" + "e" * 64),
        _candidate(container_image="sha256:" + "e" * 64),
    ),
)
def test_main_never_monitors_stops_or_publishes_an_identity_mismatch(
    candidate: dict[str, Any], monkeypatch: pytest.MonkeyPatch
) -> None:
    runtime = _RawObserverRuntime(candidate=candidate)
    monkeypatch.setattr(observe, "_new_system_runtime", lambda: runtime)

    assert observe.main(("--compose-project", _PROJECT)) == 75
    assert "monitor" not in runtime.events
    assert "stop" not in runtime.events
    assert "publish" not in runtime.events
    assert runtime.stop_calls == []
    assert runtime.writes == []
    assert runtime.events[-1] == "lease.release"


@pytest.mark.parametrize("relative", ("/absolute.json", "../parent.json", "nested/../../escape.json", "C:/escape.json"))
def test_safe_paths_reject_escape_before_open_or_write(relative: str) -> None:
    ops = _PathOps()
    with pytest.raises(safe_paths.SafePathError):
        safe_paths.open_regular_relative(Path("/reports/run.partial"), relative, ops=ops)
    assert ops.root_open_calls == []
    assert ops.nt_absolute_directory_calls == []
    assert ops.openat_calls == []
    assert ops.parent_open_calls == []
    assert ops.write_new_at_calls == []


@pytest.mark.parametrize(
    "ops",
    (
        _PathOps(fail_open=True),
        _PathOps(hard_link=True),
        _PathOps(nonregular=True),
        _PathOps(reparse=True),
        _PathOps(size_changed=True),
    ),
)
def test_safe_paths_fail_closed_for_links_types_reparse_and_size_only_toctou(
    ops: _PathOps,
) -> None:
    with pytest.raises(safe_paths.SafePathError):
        safe_paths.open_regular_relative(Path("/reports/run.partial"), "owner.json", ops=ops)
    assert ops.write_new_at_calls == []
    assert ops.renameat_calls == []
    assert ops.closes


def test_safe_paths_rejects_post_open_identity_toctou_on_same_opened_descriptor_without_publication() -> None:
    ops = _PathOps(identity_changed=True)
    with pytest.raises(safe_paths.SafePathError):
        safe_paths.open_regular_relative(Path("/reports/run.partial"), "owner.json", ops=ops)
    assert ops.fstat_descriptors == [17, 18, 18]
    assert ops.closes == [17, 18]
    assert ops.write_new_at_calls == []
    assert ops.renameat_calls == []


def test_safe_paths_windows_reparse_closes_handle_without_write() -> None:
    ops = _PathOps(windows=True, reparse=True)
    with pytest.raises(safe_paths.SafePathError):
        safe_paths.open_regular_relative(Path("C:/reports/run.partial"), "owner.json", ops=ops)
    assert ops.nt_absolute_directory_calls == [Path("C:/reports/run.partial")]
    assert ops.nt_relative_calls == [(17, "owner.json", False)]
    assert ops.fstat_descriptors == [17, 18]
    assert ops.closes == [18, 17]
    assert ops.write_new_at_calls == []


@pytest.mark.parametrize(
    ("windows", "root"),
    (
        (False, Path("/reports/run.partial")),
        (True, Path("C:/reports/run.partial")),
    ),
)
def test_safe_paths_reject_cross_device_descendant_and_closes_every_rooted_handle(
    windows: bool,
    root: Path,
) -> None:
    ops = _PathOps(windows=windows, descendant_cross_device=True)
    with pytest.raises(safe_paths.SafePathError):
        safe_paths.open_regular_relative(root, "owner.json", ops=ops)

    assert ops.fstat_descriptors == [17, 18]
    assert ops.closes == [18, 17]
    assert ops.write_new_at_calls == []
    assert ops.renameat_calls == []
    assert ops.parent_open_calls == []
    if windows:
        assert ops.nt_absolute_directory_calls == [root]
        assert ops.nt_relative_calls == [(17, "owner.json", False)]
        assert ops.root_open_calls == []
        assert ops.openat_calls == []
    else:
        assert ops.root_open_calls == [root]
        assert [(parent, name) for parent, name, _flags in ops.openat_calls] == [
            (17, "owner.json")
        ]
        assert ops.nt_absolute_directory_calls == []
        assert ops.nt_relative_calls == []


@pytest.mark.parametrize(
    ("windows", "root"),
    (
        (False, Path("/reports/run.partial")),
        (True, Path("C:/reports/run.partial")),
    ),
)
def test_safe_paths_reject_reparse_root_without_opening_a_descendant(
    windows: bool,
    root: Path,
) -> None:
    ops = _PathOps(windows=windows, root_reparse=True)
    with pytest.raises(safe_paths.SafePathError):
        safe_paths.open_regular_relative(root, "owner.json", ops=ops)

    assert ops.fstat_descriptors == [17]
    assert ops.closes == [17]
    assert ops.openat_calls == []
    assert ops.nt_relative_calls == []
    assert ops.write_new_at_calls == []
    assert ops.renameat_calls == []
    assert ops.parent_open_calls == []
    if windows:
        assert ops.nt_absolute_directory_calls == [root]
        assert ops.root_open_calls == []
    else:
        assert ops.root_open_calls == [root]
        assert ops.nt_absolute_directory_calls == []


@pytest.mark.parametrize(
    ("windows", "root"),
    (
        (False, Path("/reports/run.partial")),
        (True, Path("C:/reports/run.partial")),
    ),
)
def test_safe_paths_reject_reparse_intermediate_and_closes_every_opened_handle(
    windows: bool,
    root: Path,
) -> None:
    ops = _PathOps(windows=windows, intermediate_reparse=True)
    with pytest.raises(safe_paths.SafePathError):
        safe_paths.open_regular_relative(root, "sealed/owner.json", ops=ops)

    assert ops.fstat_descriptors == [17, 18]
    assert ops.closes == [18, 17]
    assert ops.write_new_at_calls == []
    assert ops.renameat_calls == []
    assert ops.parent_open_calls == []
    if windows:
        assert ops.nt_absolute_directory_calls == [root]
        assert ops.nt_relative_calls == [(17, "sealed", True)]
        assert ops.root_open_calls == []
        assert ops.openat_calls == []
    else:
        assert ops.root_open_calls == [root]
        assert [(parent, name) for parent, name, _flags in ops.openat_calls] == [
            (17, "sealed")
        ]
        assert ops.nt_absolute_directory_calls == []
        assert ops.nt_relative_calls == []


def test_safe_paths_posix_anchors_every_component_and_swap_cannot_redirect_leaf() -> None:
    ops = _PathOps(swap_intermediate_name=True)
    descriptor = safe_paths.open_regular_relative(
        Path("/reports/run.partial"), "sealed/owner.json", ops=ops
    )
    assert descriptor == 19
    assert len(ops.openat_calls) == 2
    assert ops.openat_calls[0][1] == "sealed"
    assert ops.openat_calls[1][1] == "owner.json"
    assert ops.openat_calls[1][0] == 18
    assert all(flags & safe_paths.O_NOFOLLOW for _, _, flags in ops.openat_calls)
    assert ops.root_open_calls == [Path("/reports/run.partial")]
    assert all(
        {"directory", "device", "reparse", "regular", "nlink"} <= metadata.keys()
        for metadata in ops.descriptor_metadata.values()
    )
    assert ops.descriptor_metadata[17]["directory"] is True
    assert ops.descriptor_metadata[19]["regular"] is True
    assert ops.descriptor_metadata[19]["nlink"] == 1


def test_safe_paths_atomic_write_uses_renameat_with_one_parent_fd() -> None:
    ops = _PathOps()
    safe_paths.atomic_write_same_dir(
        Path("/reports/run.partial/final-receipt.json"), b"{}\n", ops=ops
    )
    assert ops.parent_open_calls == [Path("/reports/run.partial")]
    assert ops.destination_checks == [Path("/reports/run.partial/final-receipt.json")]
    assert ops.write_new_at_calls == [(17, "final-receipt.json", b"{}\n")]
    assert len(ops.renameat_calls) == 1
    source_parent_fd, source, destination_parent_fd, destination = ops.renameat_calls[0]
    assert source_parent_fd == destination_parent_fd
    assert source.endswith(".tmp")
    assert destination == "final-receipt.json"
    assert ops.closes == [17]

    for failed_ops in (_PathOps(existing_output=True), _PathOps(cross_directory_rename=True)):
        with pytest.raises(safe_paths.SafePathError):
            safe_paths.atomic_write_same_dir(
                Path("/reports/run.partial/final-receipt.json"), b"{}\n", ops=failed_ops
            )
        if failed_ops.existing_output:
            assert failed_ops.write_new_at_calls == []
        assert failed_ops.parent_open_calls == [Path("/reports/run.partial")]
        assert failed_ops.closes == [17]


def test_safe_paths_reject_windows_alternate_data_streams_before_rooted_open_or_write() -> None:
    read_ops = _PathOps(windows=True)
    with pytest.raises(safe_paths.SafePathError):
        safe_paths.open_regular_relative(
            Path("C:/reports/run.partial"), "owner.json:alternate", ops=read_ops
        )
    assert read_ops.nt_absolute_directory_calls == []
    assert read_ops.nt_relative_calls == []

    write_ops = _PathOps(windows=True)
    with pytest.raises(safe_paths.SafePathError):
        safe_paths.atomic_write_same_dir(
            Path("C:/reports/run.partial/final-receipt.json:alternate"),
            b"{}\n",
            ops=write_ops,
        )
    assert write_ops.parent_open_calls == []
    assert write_ops.write_new_at_calls == []


@pytest.mark.skipif(os.name == "nt", reason="POSIX descriptor enumeration only")
def test_safe_paths_posix_directory_enumeration_is_sorted_and_bounded(
    tmp_path: Path,
) -> None:
    root = tmp_path / "root"
    directory = root / "source"
    directory.mkdir(parents=True)
    for name in ("zeta", "alpha", ".hidden"):
        (directory / name).write_bytes(b"source\n")

    assert safe_paths.list_directory_relative(root, "source") == (
        ".hidden",
        "alpha",
        "zeta",
    )
    with pytest.raises(safe_paths.SafePathError):
        safe_paths.list_directory_relative(root, "source", max_entries=2)
    with pytest.raises(safe_paths.SafePathError):
        safe_paths.list_directory_relative(root, "source", max_name_bytes=3)


@pytest.mark.skipif(os.name == "nt", reason="POSIX descriptor enumeration only")
def test_safe_paths_posix_directory_timestamp_race_fails_closed(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
) -> None:
    root = tmp_path / "root"
    directory = root / "source"
    directory.mkdir(parents=True)
    (directory / "before").write_bytes(b"source\n")
    original = safe_paths._posix_directory_entry_names

    def mutate_after_enumeration(
        descriptor: int, max_entries: int, max_name_bytes: int
    ) -> tuple[str, ...]:
        names = original(descriptor, max_entries, max_name_bytes)
        before = directory.stat()
        (directory / "raced").write_bytes(b"source\n")
        changed_mtime_ns = before.st_mtime_ns + 1_000_000
        os.utime(directory, ns=(before.st_atime_ns, changed_mtime_ns))
        assert directory.stat().st_mtime_ns == changed_mtime_ns
        return names

    monkeypatch.setattr(
        safe_paths, "_posix_directory_entry_names", mutate_after_enumeration
    )
    with pytest.raises(safe_paths.SafePathError):
        safe_paths.list_directory_relative(root, "source")


@pytest.mark.parametrize(
    "responses",
    (
        ((0xC0000001, b"", 0),),
        ((0, b"", 0),),
        ((0, _nt_directory_record("offset", next_offset=1), None),),
        ((0, _nt_directory_record("reparse", attributes=0x00000400), None),),
        (
            (0, _nt_directory_record("duplicate"), None),
            (0, _nt_directory_record("duplicate"), None),
            (0x80000006, b"", 0),
        ),
    ),
    ids=("status", "empty-buffer", "offset", "reparse", "duplicate"),
)
def test_safe_paths_windows_handle_enumeration_rejects_malformed_records(
    responses: tuple[tuple[int, bytes, int | None], ...],
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    query = _NtDirectoryQuery(list(responses))
    monkeypatch.setattr(
        safe_paths.ctypes,
        "windll",
        SimpleNamespace(ntdll=SimpleNamespace(NtQueryDirectoryFile=query)),
        raising=False,
    )

    with pytest.raises(safe_paths.SafePathError):
        safe_paths._nt_directory_entry_names(73, 8, 255)
    assert query.descriptors and set(query.descriptors) == {73}


def test_safe_paths_windows_handle_enumeration_rejects_timestamp_change(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
) -> None:
    root = tmp_path / "root"
    root.mkdir()
    before = {
        "device": 7,
        "inode": 73,
        "size": 0,
        "regular": False,
        "nlink": 1,
        "reparse": False,
        "directory": True,
        "disk": True,
        "mode": 0o100600,
        "mtime_ns": 1_000_011,
        "ctime_ns": 2_000_011,
        "mode": 0o040755,
        "mtime_ns": 11,
        "ctime_ns": 13,
    }
    after = {**before, "mtime_ns": 17, "ctime_ns": 19}
    snapshots = iter((before, after))
    monkeypatch.setattr(safe_paths.os, "name", "nt")
    monkeypatch.setattr(
        safe_paths,
        "_open_windows_directory_relative",
        lambda _root, _parts: (73, 7),
    )
    monkeypatch.setattr(
        safe_paths,
        "_fstat",
        lambda _ops, _descriptor: next(snapshots),
    )
    monkeypatch.setattr(
        safe_paths,
        "_nt_directory_entry_names",
        lambda _descriptor, _max_entries, _max_name_bytes: ("entry",),
    )
    monkeypatch.setattr(safe_paths, "_close", lambda _ops, _descriptor: None)

    with pytest.raises(safe_paths.SafePathError):
        safe_paths.list_directory_relative(root, "source")


def test_lease_snapshots_bind_exact_source_snapshot_and_no_acl_claim(
    tmp_path: Path,
) -> None:
    probe = _LeaseProbe()
    path, reports_root = _injected_lease_path(tmp_path)
    source_lease = lease.SourceLease.acquire(
        path,
        probe=probe,
        trusted_reports_root=reports_root,
    )
    prebuild = source_lease.snapshot()
    postgreen = source_lease.snapshot()
    observer_snapshot = source_lease.snapshot()
    source_lease.verify_snapshot_chain(prebuild, postgreen, observer_snapshot)
    source_lease.release()

    assert probe.commands == [
        ("git", "rev-parse", "HEAD"),
        ("git", "status", "--porcelain=v1", "--untracked-files=all"),
    ] * 3
    assert tuple(prebuild) == (
        "head",
        "status",
        "source_root_sha256",
        "frozen_subjects",
        "canonical_build_input_manifest_sha256",
    )
    assert prebuild["canonical_build_input_manifest_sha256"] == _MANIFEST_SHA256
    assert prebuild["frozen_subjects"] == probe.frozen_subject_digests()
    assert tuple(prebuild["source_root_sha256"]) == tuple(_SOURCE_ROOT_DIGESTS)
    assert prebuild["source_root_sha256"] == _SOURCE_ROOT_DIGESTS
    assert probe.acl_changes == []


@pytest.mark.parametrize(
    "changed_probe",
    (
        _LeaseProbe(head="c" * 40),
        _LeaseProbe(dirty=True),
        _LeaseProbe(manifest_bytes=_manifest_bytes({
            **_MANIFEST_CONTENTS,
            _MANIFEST_PATHS[0]: _MANIFEST_CONTENTS[_MANIFEST_PATHS[0]] + b"changed",
        })),
        _LeaseProbe(subject_digest="sha256:" + "e" * 64),
        _LeaseProbe(
            source_root_digest_map={
                **_SOURCE_ROOT_DIGESTS,
                "backend/internal": "sha256:" + "e" * 64,
            }
        ),
    ),
)
def test_lease_rejects_every_prebuild_postgreen_or_observer_drift(
    changed_probe: _LeaseProbe, tmp_path: Path
) -> None:
    original = _LeaseProbe()
    path, reports_root = _injected_lease_path(tmp_path)
    source_lease = lease.SourceLease.acquire(
        path,
        probe=original,
        trusted_reports_root=reports_root,
    )
    prebuild = source_lease.snapshot()
    postgreen = source_lease.snapshot()
    source_lease.probe = changed_probe
    observer_snapshot = source_lease.snapshot()
    with pytest.raises(lease.SourceMutationError):
        source_lease.verify_snapshot_chain(prebuild, postgreen, observer_snapshot)
    source_lease.release()
    assert original.acl_changes == []
    assert changed_probe.acl_changes == []


def test_lease_injected_probe_requires_exact_explicit_reports_root(
    tmp_path: Path,
) -> None:
    path, reports_root = _injected_lease_path(tmp_path)
    probe = _LeaseProbe()
    for trusted_reports_root in (
        None,
        Path("reports"),
        tmp_path / "not-reports",
    ):
        with pytest.raises(lease.SourceLeaseError):
            lease.SourceLease.acquire(
                path,
                probe=probe,
                trusted_reports_root=trusted_reports_root,
            )
    missing_reports_root = tmp_path / "missing" / "reports"
    with pytest.raises(lease.SourceLeaseError):
        lease.SourceLease.acquire(
            missing_reports_root / ".backend-unit-source-mutation.lease",
            probe=probe,
            trusted_reports_root=missing_reports_root,
        )
    assert not path.exists()


def test_lease_system_probe_rejects_injected_reports_root(tmp_path: Path) -> None:
    path, reports_root = _injected_lease_path(tmp_path)
    with pytest.raises(lease.SourceLeaseError):
        lease.SourceLease.acquire(
            path,
            repo_root=tmp_path,
            trusted_reports_root=reports_root,
        )
    assert not path.exists()


def test_root_handoff_accepts_only_exact_canonical_final_and_provisional_bytes() -> None:
    document = _root_handoff_document()
    payload = root_handoff.canonicalize_root_handoff(document)
    provisional = _root_handoff_provisional()
    provisional_payload = root_handoff._canonicalize_provisional(provisional)

    assert root_handoff.parse_root_handoff(payload) == document
    assert root_handoff._parse_provisional(provisional_payload) == provisional
    assert payload.endswith(b"\n") and b"\r" not in payload
    assert payload == root_handoff.canonicalize_root_handoff(document)

    reordered = {"target": document["target"], **document}
    missing = deepcopy(document)
    missing.pop("fixed_image_ids")
    extra = deepcopy(document)
    extra["extra"] = True
    dirty = _root_handoff_document(
        prebuild=_root_handoff_snapshot(status=" M source.go\n")
    )
    unequal = _root_handoff_document(
        postgreen={
            **_root_handoff_snapshot(),
            "head": "c" * 40,
        }
    )
    changed_manifest = _MANIFEST_BYTES.replace(
        _MANIFEST_BYTES.splitlines()[0].split(b"\t", 1)[1],
        b"sha256:" + b"0" * 64,
        1,
    )
    manifest_hash_mismatch = _root_handoff_document(manifest=changed_manifest)
    snapshot_manifest_mismatch = _root_handoff_document(
        prebuild=_root_handoff_snapshot(manifest_sha256="sha256:" + "0" * 64)
    )
    reordered_fixed_images = _root_handoff_document(
        fixed_image_ids={
            "backend_unit": _BACKEND_IMAGE_ID,
            "python_contract": _PYTHON_IMAGE_ID,
        }
    )
    invalid_fixed_digest = _root_handoff_document(
        fixed_image_ids={
            "python_contract": _PYTHON_IMAGE_ID,
            "backend_unit": "sha256:" + "z" * 64,
        }
    )
    snapshot_with_image = _root_handoff_document()
    for phase in ("prebuild", "postgreen"):
        snapshot_with_image[phase]["fixed_image_ids"] = {
            "python_contract": _PYTHON_IMAGE_ID,
            "backend_unit": _BACKEND_IMAGE_ID,
        }

    for malformed in (
        reordered,
        missing,
        extra,
        dirty,
        unequal,
        manifest_hash_mismatch,
        snapshot_manifest_mismatch,
        reordered_fixed_images,
        invalid_fixed_digest,
        snapshot_with_image,
    ):
        with pytest.raises(root_handoff.RootHandoffError):
            root_handoff.validate_root_handoff(malformed)

    duplicate = payload.replace(
        b'{"schema":',
        b'{"schema":"duplicate","schema":',
        1,
    )
    extra_bytes = payload[:-2] + b',"extra":true}\n'
    whitespace = payload.replace(b'":"', b'": "', 1)
    for malformed_bytes in (
        duplicate,
        extra_bytes,
        whitespace,
        payload.rstrip(b"\n") + b"\r\n",
        payload[:-1],
    ):
        with pytest.raises(root_handoff.RootHandoffError):
            root_handoff.parse_root_handoff(malformed_bytes)


def test_root_handoff_prebuild_then_postgreen_uses_only_fixed_root_and_paths(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    root = Path("/root-handoff-repository")
    snapshot = _root_handoff_snapshot()
    source_lease = _RootHandoffLease((snapshot, snapshot))
    files: dict[str, bytes] = {}
    events: list[str] = []
    acquired_paths: list[Path] = []
    read_roots: list[Path] = []

    def acquire(path: Path) -> _RootHandoffLease:
        assert path == root
        acquired_paths.append(path)
        events.append("acquire")
        return source_lease

    def write(path: Path, payload: bytes) -> None:
        relative = path.relative_to(root).as_posix()
        assert relative in {
            "testframework/reports/.backend-unit-root-handoff.prebuild.v1.json",
            "testframework/reports/.backend-unit-root-handoff.v1.json",
        }
        if relative in files:
            raise FileExistsError(relative)
        files[relative] = payload
        events.append(f"write:{relative}")

    def read(read_root: Path, relative: str) -> bytes:
        assert read_root == root
        read_roots.append(read_root)
        events.append(f"read:{relative}")
        return files[relative]

    def read_provisional(read_root: Path) -> tuple[bytes, tuple[int, int, int, int, int, int, int]]:
        assert read_root == root
        read_roots.append(read_root)
        events.append("read:testframework/reports/.backend-unit-root-handoff.prebuild.v1.json")
        payload = files["testframework/reports/.backend-unit-root-handoff.prebuild.v1.json"]
        return payload, (7, 11, 0o100600, 1, len(payload), 1_000_011, 2_000_011)

    def remove(
        remove_root: Path,
        expected_payload: bytes,
        expected_identity: tuple[int, int, int, int, int, int, int],
    ) -> None:
        assert remove_root == root
        assert expected_payload == files[
            "testframework/reports/.backend-unit-root-handoff.prebuild.v1.json"
        ]
        assert expected_identity == (
            7,
            11,
            0o100600,
            1,
            len(expected_payload),
            1_000_011,
            2_000_011,
        )
        events.append("remove:provisional")
        files.pop("testframework/reports/.backend-unit-root-handoff.prebuild.v1.json")

    monkeypatch.setattr(root_handoff, "_module_repository_root", lambda: root)
    monkeypatch.setattr(root_handoff, "_acquire_lease", acquire)
    monkeypatch.setattr(root_handoff, "_write_new", write)
    monkeypatch.setattr(root_handoff, "_read_relative", read)
    monkeypatch.setattr(root_handoff, "_read_provisional", read_provisional)
    monkeypatch.setattr(root_handoff, "_remove_provisional", remove)

    prebuild = root_handoff.write_prebuild_handoff()
    final = root_handoff.finalize_postgreen_handoff()

    assert tuple(prebuild) == (
        "prebuild",
        "canonical_build_input_manifest_bytes",
        "canonical_build_input_manifest_sha256",
    )
    assert final["prebuild"] == final["postgreen"] == snapshot
    assert root_handoff.parse_root_handoff(
        files["testframework/reports/.backend-unit-root-handoff.v1.json"]
    ) == final
    assert acquired_paths == [root, root]
    assert read_roots == [root, root, root]
    assert events == [
        "acquire",
        "write:testframework/reports/.backend-unit-root-handoff.prebuild.v1.json",
        "read:testframework/reports/.backend-unit-root-handoff.prebuild.v1.json",
        "acquire",
        "read:testframework/reports/.backend-unit-root-handoff.prebuild.v1.json",
        "write:testframework/reports/.backend-unit-root-handoff.v1.json",
        "read:testframework/reports/.backend-unit-root-handoff.v1.json",
        "remove:provisional",
    ]
    assert source_lease.events == [
        "snapshot",
        "manifest",
        "release",
        "snapshot",
        "manifest",
        "verify",
        "images",
        "release",
    ]


@pytest.mark.parametrize(
    "failure",
    (
        FileExistsError("already exists"),
        safe_paths.SafePathError("symlink destination"),
        safe_paths.SafePathError("reparse destination"),
    ),
    ids=("nonreplace", "symlink", "reparse"),
)
def test_root_handoff_write_is_nonreplace_and_fails_closed_on_unsafe_destination(
    failure: Exception, monkeypatch: pytest.MonkeyPatch
) -> None:
    target = Path("/root-handoff-repository/testframework/reports/receipt.json")
    payload = b'{"sealed":true}\n'
    calls: list[tuple[Path, bytes]] = []

    def reject(path: Path, written: bytes) -> None:
        calls.append((path, written))
        raise failure

    monkeypatch.setattr(root_handoff.safe_paths, "atomic_write_same_dir", reject)
    with pytest.raises(root_handoff.RootHandoffError):
        root_handoff._write_new(target, payload)
    assert calls == [(target, payload)]


def test_root_handoff_descriptor_toctou_and_provisional_removal_are_exact(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    root = Path("/root-handoff-repository")
    stable = {
        "device": 7,
        "inode": 11,
        "size": 4,
        "regular": True,
        "nlink": 1,
        "reparse": False,
        "disk": True,
        "mode": 0o100600,
        "mtime_ns": 1_000_011,
        "ctime_ns": 2_000_011,
    }
    identity_changed = {**stable, "inode": 12}
    snapshots = iter((stable, identity_changed))
    monkeypatch.setattr(
        root_handoff.safe_paths,
        "_fstat",
        lambda _ops, _descriptor: next(snapshots),
    )
    monkeypatch.setattr(
        root_handoff.safe_paths,
        "_read_descriptor_exact",
        lambda _descriptor, _size, **_kwargs: b"same",
    )
    with pytest.raises(root_handoff.RootHandoffError):
        root_handoff._read_open_descriptor(71, 64)

    byte_snapshots = iter((stable, stable))
    monkeypatch.setattr(
        root_handoff.safe_paths,
        "_fstat",
        lambda _ops, _descriptor: next(byte_snapshots),
    )
    monkeypatch.setattr(
        root_handoff.safe_paths,
        "_read_descriptor_exact",
        lambda _descriptor, _size, **_kwargs: b"bad",
    )
    with pytest.raises(root_handoff.RootHandoffError):
        root_handoff._read_open_descriptor(71, 64)

    expected_payload = b"expected-provisional\n"
    expected_identity = (
        7,
        11,
        0o100600,
        1,
        len(expected_payload),
        1_000_011,
        2_000_011,
    )
    opened: list[tuple[int, str, int]] = []
    deleted: list[int] = []
    closes: list[int] = []
    monkeypatch.setattr(root_handoff.os, "name", "nt")
    monkeypatch.setattr(
        root_handoff.safe_paths,
        "_nt_open_absolute_directory",
        lambda reports: (17, (), 7),
    )

    def open_relative(
        parent: int,
        name: str,
        directory: bool,
        *,
        write: bool = False,
        tolerate_missing: bool = False,
    ) -> int:
        opened.append((parent, name, int(write)))
        assert directory is False
        if tolerate_missing:
            return -1
        return 18

    monkeypatch.setattr(root_handoff.safe_paths, "_nt_open_relative", open_relative)
    monkeypatch.setattr(
        root_handoff,
        "_read_open_descriptor",
        lambda _descriptor, _max_bytes: (expected_payload, expected_identity),
    )
    monkeypatch.setattr(
        root_handoff.safe_paths,
        "_windows_delete_handle",
        lambda descriptor: deleted.append(descriptor),
    )
    monkeypatch.setattr(
        root_handoff.safe_paths,
        "_close",
        lambda _ops, descriptor: closes.append(descriptor),
    )

    root_handoff._remove_provisional(
        root, expected_payload, expected_identity
    )
    assert opened == [
        (17, ".backend-unit-root-handoff.prebuild.v1.json", 1),
        (17, ".backend-unit-root-handoff.prebuild.v1.json", 0),
    ]
    assert deleted == [18]
    assert closes == [18, 17]


@pytest.mark.parametrize("platform", ("posix", "nt"))
def test_root_handoff_removal_preserves_same_byte_foreign_identity_across_reopen(
    platform: str, monkeypatch: pytest.MonkeyPatch
) -> None:
    root = Path("/root-handoff-repository")
    payload = b"same-provisional-bytes\n"
    expected_identity = (7, 11, 0o100600, 1, len(payload), 1_000_011, 2_000_011)
    foreign_identity = (7, 12, 0o100600, 1, len(payload), 1_000_012, 2_000_012)
    opened: list[tuple[int, str, bool]] = []
    closes: list[int] = []
    deletes: list[int] = []

    monkeypatch.setattr(root_handoff.os, "name", "nt" if platform == "nt" else "posix")
    monkeypatch.setattr(
        root_handoff,
        "_read_open_descriptor",
        lambda _descriptor, _max_bytes: (payload, foreign_identity),
    )
    monkeypatch.setattr(
        root_handoff.safe_paths,
        "_close",
        lambda _ops, descriptor: closes.append(descriptor),
    )

    if platform == "nt":
        monkeypatch.setattr(
            root_handoff.safe_paths,
            "_nt_open_absolute_directory",
            lambda _reports: (17, (), 7),
        )

        def open_relative(
            parent: int, name: str, directory: bool, *, write: bool = False, **_kwargs: object
        ) -> int:
            opened.append((parent, name, write))
            return 18

        monkeypatch.setattr(root_handoff.safe_paths, "_nt_open_relative", open_relative)
        monkeypatch.setattr(
            root_handoff.safe_paths,
            "_windows_delete_handle",
            lambda descriptor: deletes.append(descriptor),
        )
    else:
        monkeypatch.setattr(root_handoff.safe_paths, "O_NOFOLLOW", 1)
        monkeypatch.setattr(
            root_handoff.safe_paths,
            "_posix_open_directory",
            lambda _reports: (17, 7),
        )

        def openat(parent: int, name: str, _flags: int) -> int:
            opened.append((parent, name, False))
            return 18

        monkeypatch.setattr(root_handoff.safe_paths, "_posix_openat", openat)
        monkeypatch.setattr(
            root_handoff.ctypes,
            "CDLL",
            lambda *_args, **_kwargs: pytest.fail("foreign provisional must not reach unlinkat"),
        )

    with pytest.raises(root_handoff.RootHandoffError, match="identity does not match"):
        root_handoff._remove_provisional(root, payload, expected_identity)

    assert deletes == []
    assert opened == [(17, ".backend-unit-root-handoff.prebuild.v1.json", platform == "nt")]
    assert closes == [18, 17]


def test_root_handoff_cli_rejects_nonmodule_wrong_cwd_python_and_argv(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    for package, spec_name in (
        (None, None),
        ("observers.backend_unit", "root_handoff"),
        ("observers.other", "observers.backend_unit.root_handoff"),
    ):
        with monkeypatch.context() as scoped:
            scoped.setattr(root_handoff, "__package__", package)
            scoped.setattr(
                root_handoff,
                "__spec__",
                SimpleNamespace(name=spec_name),
            )
            with pytest.raises(root_handoff.RootHandoffError):
                root_handoff._assert_module_invocation()

    root = Path("/root-handoff-repository")
    with monkeypatch.context() as scoped:
        scoped.setattr(root_handoff, "_module_repository_root", lambda: root)
        scoped.setattr(
            root_handoff,
            "Path",
            SimpleNamespace(cwd=lambda: Path("/different-repository")),
        )
        with pytest.raises(root_handoff.RootHandoffError):
            root_handoff._canonical_cli_root()

    phase_calls: list[str] = []
    cwd_calls: list[str] = []

    def canonical_cli_root() -> Path:
        cwd_calls.append("cwd")
        return root

    monkeypatch.setattr(root_handoff, "_assert_module_invocation", lambda: None)
    monkeypatch.setattr(root_handoff, "_canonical_cli_root", canonical_cli_root)
    monkeypatch.setattr(
        root_handoff.sys,
        "version_info",
        (3, 12, 0, "final", 0),
    )
    monkeypatch.setattr(
        root_handoff,
        "write_prebuild_handoff",
        lambda: phase_calls.append("prebuild"),
    )
    monkeypatch.setattr(
        root_handoff,
        "finalize_postgreen_handoff",
        lambda: phase_calls.append("postgreen"),
    )
    assert root_handoff.main(("--phase", "prebuild")) == 0
    assert root_handoff.main(("--phase", "postgreen")) == 0
    assert phase_calls == ["prebuild", "postgreen"]
    assert cwd_calls == ["cwd", "cwd"]

    for argv in ((), ("--phase",), ("--phase", "other"), ("phase", "prebuild")):
        with pytest.raises(root_handoff.RootHandoffError):
            root_handoff.main(argv)
    assert phase_calls == ["prebuild", "postgreen"]
    assert cwd_calls == ["cwd", "cwd"]

    for version in ((3, 11, 9, "final", 0), (3, 13, 0, "final", 0)):
        monkeypatch.setattr(root_handoff.sys, "version_info", version)
        with pytest.raises(root_handoff.RootHandoffError, match="Python 3.12"):
            root_handoff.main(("--phase", "prebuild"))
        assert phase_calls == ["prebuild", "postgreen"]
        assert cwd_calls == ["cwd", "cwd"]


def test_observer_timing_preflight_and_closed_classification_priority() -> None:
    timing = resources.ObserverTiming()
    assert (
        timing.candidate_window_seconds,
        timing.runtime_sample_interval_seconds,
        timing.ready_deadline_seconds,
        timing.destroy_grace_seconds,
        timing.whole_run_deadline_seconds,
        timing.deadline_owner,
    ) == (120, 5, 30, 30, 6 * 60 * 60, "observer")

    resources.preflight((_safe_sample(), _safe_sample(), _safe_sample()))
    for unsafe in (
        {**_safe_sample(), "free_ram_bytes": 2 * 1024**3 - 1},
        {**_safe_sample(), "free_c_bytes": 250 * 1024**3 - 1},
        {**_safe_sample(), "PercentDiskTime": 70},
        {**_safe_sample(), "CurrentDiskQueueLength": 2},
    ):
        with pytest.raises(resources.ResourcePolicyError):
            resources.preflight((unsafe, unsafe, _safe_sample()))

    cases = (
        ({"source_mutated": True, "invalid_provenance": True}, "SEALED_INVALID_SOURCE_MUTATED"),
        ({"invalid_provenance": True, "explicit_stops": (("ram", 1),)}, "SEALED_INVALID_PROVENANCE"),
        ({"explicit_stops": (("interrupt", 1),)}, "SEALED_INTERRUPTED"),
        ({"explicit_stops": (("qa", 1),)}, "SEALED_INVALID_QA_OVERLAP"),
        ({"explicit_stops": (("disk", 1),)}, "SEALED_INFRA_ABORTED_DISK_PRESSURE"),
        ({"explicit_stops": (("ram", 1),)}, "SEALED_INFRA_ABORTED_HOST_RAM_FLOOR"),
        ({"explicit_stops": (("deadline", 1),)}, "SEALED_INFRA_ABORTED_WHOLE_RUN_TIMEOUT"),
        ({"oom": True, "resource": True}, "SEALED_INFRA_ABORTED_OOM"),
        ({"resource": True, "budget": True}, "SEALED_INFRA_ABORTED_RESOURCE_CONTRACT"),
        ({"budget": True, "tmpfs": True}, "SEALED_INFRA_ABORTED_REPORT_BUDGET"),
        ({"tmpfs": True, "timeout": True}, "SEALED_INFRA_ABORTED_TMPFS"),
        ({"timeout": True}, "SEALED_TIMEOUT_UNTRIAGED"),
        ({"discovery_failure": True}, "SEALED_EXECUTABLE_FAIL_DISCOVERY"),
        ({"stage_failure": True}, "SEALED_EXECUTABLE_FAIL_STAGE"),
        ({}, "SEALED_EXECUTABLE_PASS"),
    )
    assert sorted(expected for _, expected in cases) == sorted(_SEALED_CLASSIFICATIONS)
    for raw_observations, expected in cases:
        assert resources.classify((_safe_sample(),), _policy(raw_observations)) == expected


_SIMULTANEOUS_STOP_CASES = (
    (("interrupt",), "SEALED_INTERRUPTED"),
    (("qa",), "SEALED_INVALID_QA_OVERLAP"),
    (("disk",), "SEALED_INFRA_ABORTED_DISK_PRESSURE"),
    (("ram",), "SEALED_INFRA_ABORTED_HOST_RAM_FLOOR"),
    (("deadline",), "SEALED_INFRA_ABORTED_WHOLE_RUN_TIMEOUT"),
    (("interrupt", "qa"), "SEALED_INTERRUPTED"),
    (("interrupt", "disk"), "SEALED_INTERRUPTED"),
    (("interrupt", "ram"), "SEALED_INTERRUPTED"),
    (("interrupt", "deadline"), "SEALED_INTERRUPTED"),
    (("qa", "disk"), "SEALED_INVALID_QA_OVERLAP"),
    (("qa", "ram"), "SEALED_INVALID_QA_OVERLAP"),
    (("qa", "deadline"), "SEALED_INVALID_QA_OVERLAP"),
    (("disk", "ram"), "SEALED_INFRA_ABORTED_DISK_PRESSURE"),
    (("disk", "deadline"), "SEALED_INFRA_ABORTED_DISK_PRESSURE"),
    (("ram", "deadline"), "SEALED_INFRA_ABORTED_HOST_RAM_FLOOR"),
    (("interrupt", "qa", "disk"), "SEALED_INTERRUPTED"),
    (("interrupt", "qa", "ram"), "SEALED_INTERRUPTED"),
    (("interrupt", "qa", "deadline"), "SEALED_INTERRUPTED"),
    (("interrupt", "disk", "ram"), "SEALED_INTERRUPTED"),
    (("interrupt", "disk", "deadline"), "SEALED_INTERRUPTED"),
    (("interrupt", "ram", "deadline"), "SEALED_INTERRUPTED"),
    (("qa", "disk", "ram"), "SEALED_INVALID_QA_OVERLAP"),
    (("qa", "disk", "deadline"), "SEALED_INVALID_QA_OVERLAP"),
    (("qa", "ram", "deadline"), "SEALED_INVALID_QA_OVERLAP"),
    (("disk", "ram", "deadline"), "SEALED_INFRA_ABORTED_DISK_PRESSURE"),
    (("interrupt", "qa", "disk", "ram"), "SEALED_INTERRUPTED"),
    (("interrupt", "qa", "disk", "deadline"), "SEALED_INTERRUPTED"),
    (("interrupt", "qa", "ram", "deadline"), "SEALED_INTERRUPTED"),
    (("interrupt", "disk", "ram", "deadline"), "SEALED_INTERRUPTED"),
    (("qa", "disk", "ram", "deadline"), "SEALED_INVALID_QA_OVERLAP"),
    (("interrupt", "qa", "disk", "ram", "deadline"), "SEALED_INTERRUPTED"),
)


@pytest.mark.parametrize(("kinds", "expected"), _SIMULTANEOUS_STOP_CASES)
def test_classification_resolves_all_31_simultaneous_stop_subsets_by_tie_order(
    kinds: tuple[str, ...], expected: str
) -> None:
    explicit_stops = tuple((kind, 100) for kind in kinds)
    record = resources.classification_record(
        (_safe_sample(),), _policy({"explicit_stops": explicit_stops})
    )
    assert record["classification"] == expected
    assert record["explicit_stops"] == explicit_stops


@pytest.mark.parametrize(
    ("explicit_stops", "expected"),
    (
        ((("qa", 10), ("interrupt", 11)), "SEALED_INVALID_QA_OVERLAP"),
        ((("disk", 10), ("qa", 11)), "SEALED_INFRA_ABORTED_DISK_PRESSURE"),
        ((("ram", 10), ("disk", 11)), "SEALED_INFRA_ABORTED_HOST_RAM_FLOOR"),
        ((("deadline", 10), ("ram", 11)), "SEALED_INFRA_ABORTED_WHOLE_RUN_TIMEOUT"),
        ((("deadline", 100), ("interrupt", 101)), "SEALED_INFRA_ABORTED_WHOLE_RUN_TIMEOUT"),
    ),
)
def test_classification_prefers_earliest_stop_before_later_higher_priority(
    explicit_stops: tuple[tuple[str, int], ...], expected: str
) -> None:
    record = resources.classification_record(
        (_safe_sample(),), _policy({"explicit_stops": explicit_stops})
    )
    assert record["classification"] == expected
    assert record["selected_stop"] == explicit_stops[0]


def test_source_and_provenance_ceiling_preempt_every_explicit_stop() -> None:
    simultaneous = (("interrupt", 100), ("qa", 100), ("disk", 100), ("ram", 100), ("deadline", 100))
    assert resources.classify(
        (_safe_sample(),), _policy({"source_mutated": True, "invalid_provenance": True, "explicit_stops": simultaneous})
    ) == "SEALED_INVALID_SOURCE_MUTATED"
    assert resources.classify(
        (_safe_sample(),), _policy({"invalid_provenance": True, "explicit_stops": simultaneous})
    ) == "SEALED_INVALID_PROVENANCE"


def test_runtime_monitoring_retains_samples_and_stops_only_once_on_priority() -> None:
    one_of_three = (_safe_sample(), {**_safe_sample(), "PercentDiskTime": 70}, _safe_sample())
    two_of_three = (_safe_sample(), {**_safe_sample(), "PercentDiskTime": 70}, {**_safe_sample(), "CurrentDiskQueueLength": 2})
    policy = _policy({})
    assert resources.classification_record(one_of_three, policy)["predicates"]["rolling_disk_pressure_2_of_3"] is False
    disk_record = resources.classification_record(two_of_three, policy)
    assert disk_record["predicates"]["rolling_disk_pressure_2_of_3"] is True
    assert disk_record["classification"] == "SEALED_INFRA_ABORTED_DISK_PRESSURE"

    samples = ({**_safe_sample(), "free_ram_bytes": 2 * 1024**3 - 1},) * 2
    ram_record = resources.classification_record(samples, policy)
    assert ram_record["classification"] == "SEALED_INFRA_ABORTED_HOST_RAM_FLOOR"
    assert ram_record["raw_samples"] == samples
    assert ram_record["predicates"]["two_consecutive_ram_below_2_gib"] is True
    assert ram_record["immutable_outcome"] is None
    assert resources.classify((_safe_sample(),), _policy({"timeout": True})) == "SEALED_TIMEOUT_UNTRIAGED"

    adapter, runner = _recoverable_bound_adapter()
    assert adapter.bind_unique_candidate_since(_START, _BACKEND_IMAGE_ID) == _CONTAINER_ID
    assert adapter.stop_bound_once(_CONTAINER_ID) == {"result": "stopped"}
    with pytest.raises(docker_adapter.DockerReadError):
        adapter.stop_bound_once(_CONTAINER_ID)
    assert [call.argv for call in runner.calls] == [
        ("docker", "ps", "--no-trunc", "--format", "{{json .}}"),
        ("docker", "inspect", "--type", "container", "--format", "{{json .}}", _CONTAINER_ID),
        ("docker", "stop", "--timeout", "5", _CONTAINER_ID),
    ]


class _InterruptSession:
    def __init__(
        self,
        events: tuple[dict[str, object] | None, ...],
        trace: list[tuple[str, object]] | None = None,
    ) -> None:
        self._events = list(events)
        self._trace = [] if trace is None else trace
        self.poll_calls = 0
        self.close_calls = 0

    def poll(self) -> dict[str, object] | None:
        self.poll_calls += 1
        event = self._events.pop(0) if self._events else None
        self._trace.append(("poll", event))
        return event

    def close(self) -> None:
        self.close_calls += 1


@dataclass
class _ObserverTerminalAdapter:
    terminal: dict[str, object]
    trace: list[tuple[str, object]] = field(default_factory=list)
    stop_calls: list[str] = field(default_factory=list)
    inspect_calls: list[str] = field(default_factory=list)
    inspect_documents: list[dict[str, object]] = field(default_factory=list)

    def stop_bound_once(self, container_id: str) -> dict[str, str]:
        self.stop_calls.append(container_id)
        self.trace.append(("stop", container_id))
        return {"result": "stopped"}

    def inspect(self, container_id: str) -> dict[str, object]:
        assert container_id == _CONTAINER_ID
        document = deepcopy(self.terminal)
        document["inspect_only"] = {
            "nonce": "recovery-die-inspect-only-7b7f6e9a",
        }
        self.inspect_calls.append(container_id)
        self.inspect_documents.append(document)
        self.trace.append(("inspect", container_id))
        return document

    def inspect_fixed_images(self) -> dict[str, str]:
        return {
            "python_contract": _PYTHON_IMAGE_ID,
            "backend_unit": _BACKEND_IMAGE_ID,
        }


class _ObserverTerminalLease:
    def fixed_image_ids(self) -> dict[str, str]:
        return {
            "python_contract": _PYTHON_IMAGE_ID,
            "backend_unit": _BACKEND_IMAGE_ID,
        }

    def canonical_build_input_manifest_bytes(self) -> bytes:
        return _MANIFEST_BYTES


def _die_event(time_nano: int) -> dict[str, object]:
    return {"status": "die", "timeNano": time_nano, "id": _CONTAINER_ID}


def _interrupted_runtime(
    events: tuple[dict[str, object] | None, ...], *, raw: dict[str, object] | None = None
) -> tuple[observe._SystemRuntime, _InterruptSession, _ObserverTerminalAdapter, dict[str, object]]:
    runtime = observe._SystemRuntime()
    trace: list[tuple[str, object]] = []
    session = _InterruptSession(events, trace)
    adapter = _ObserverTerminalAdapter({"Id": _CONTAINER_ID}, trace)
    candidate = _candidate()
    runtime._adapter = adapter
    runtime._preflight = (_safe_sample(),)
    runtime._runtime_samples = (_safe_sample(),)
    runtime._raw = {} if raw is None else dict(raw)
    runtime._event_session = session
    runtime._original_event_session = session
    return runtime, session, adapter, candidate


def _freeze_interrupt_clock(monkeypatch: pytest.MonkeyPatch, sleeps: list[float]) -> None:
    monkeypatch.setattr(observe.time, "time_ns", lambda: 1_000)
    monkeypatch.setattr(observe.time, "monotonic", lambda: 1.0)
    monkeypatch.setattr(observe.time, "sleep", lambda seconds: sleeps.append(seconds))


def test_recover_interrupt_reuses_original_session_through_die_then_closes_it(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    runtime, session, adapter, candidate = _interrupted_runtime((None, _die_event(1_001)))
    _freeze_interrupt_clock(monkeypatch, [])

    assert runtime.recover_interrupt(candidate) == (_safe_sample(),)

    assert adapter.stop_calls == [_CONTAINER_ID]
    assert adapter.inspect_calls == [_CONTAINER_ID]
    assert [entry[0] for entry in adapter.trace] == ["stop", "poll", "poll", "inspect"]
    assert adapter.trace[1:] == [
        ("poll", None),
        ("poll", _die_event(1_001)),
        ("inspect", _CONTAINER_ID),
    ]
    expected_terminal = {
        "Id": _CONTAINER_ID,
        "inspect_only": {"nonce": "recovery-die-inspect-only-7b7f6e9a"},
    }
    assert "inspect_only" not in candidate
    assert adapter.terminal == {"Id": _CONTAINER_ID}
    assert runtime._terminal == expected_terminal
    assert adapter.inspect_documents == [expected_terminal]
    assert runtime._terminal is not adapter.inspect_documents[0]
    assert runtime._terminal["inspect_only"] is not adapter.inspect_documents[0]["inspect_only"]
    adapter.terminal["Id"] = "f" * 64
    inspect_only = adapter.inspect_documents[0]["inspect_only"]
    assert isinstance(inspect_only, dict)
    inspect_only["nonce"] = "mutated-after-die"
    assert runtime._terminal == expected_terminal
    assert runtime._terminal_die_session is session
    assert runtime._original_event_session is session
    assert runtime._event_session is None
    assert session.poll_calls == 2
    assert session.close_calls == 1


def test_monitor_stops_host_and_docker_sampling_after_terminal_die(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    """A normal terminal die closes the stream before another runtime sample."""

    runtime = observe._SystemRuntime()
    trace: list[str] = []
    terminal_seen = False
    terminal_labels = {"com.docker.compose.project": _PROJECT, "evidence": "initial"}
    terminal_config = {"Labels": terminal_labels}
    terminal_mount = {"Destination": "/workspace", "Type": "bind"}
    terminal_mounts = [terminal_mount]
    terminal = {
        "Id": _CONTAINER_ID,
        "Config": terminal_config,
        "Mounts": terminal_mounts,
    }
    expected_terminal = {
        "Id": _CONTAINER_ID,
        "Config": {"Labels": {"com.docker.compose.project": _PROJECT, "evidence": "initial"}},
        "Mounts": [{"Destination": "/workspace", "Type": "bind"}],
    }

    class OneDieSession:
        close_calls = 0

        def poll(self) -> dict[str, object]:
            nonlocal terminal_seen
            if terminal_seen:
                raise AssertionError("monitor polled after terminal die")
            terminal_seen = True
            trace.append("poll:die")
            return _die_event(1_001)

        def close(self) -> None:
            self.close_calls += 1
            trace.append("close")

    session = OneDieSession()
    stats = {
        "BlockIO": "0B / 0B",
        "CPUPerc": "0.00%",
        "Container": _HOSTNAME,
        "ID": _CONTAINER_ID,
        "MemPerc": "0.00%",
        "MemUsage": "0B / 2GiB",
        "Name": "backend-unit",
        "NetIO": "0B / 0B",
        "PIDs": "0",
    }
    ps_record = {field: "" for field in observe._PS_FIELDS}
    ps_record.update(
        {
            "ID": _CONTAINER_ID,
            "Image": _BACKEND_IMAGE_ID,
            "Labels": "com.docker.compose.project=worktree-master,com.docker.compose.service=backend-unit",
            "Names": "worktree-master-backend-unit-1",
        }
    )

    class PostDieAdapter:
        def open_bound_events(self, since: str) -> OneDieSession:
            assert since == _START
            trace.append("open")
            return session

        def stats(self, container_id: str) -> dict[str, str]:
            assert container_id == _CONTAINER_ID
            if terminal_seen:
                raise AssertionError("monitor sampled Docker stats after terminal die")
            trace.append("stats")
            return dict(stats)

        def ps(self) -> tuple[dict[str, str], ...]:
            if terminal_seen:
                raise AssertionError("monitor sampled Docker ps after terminal die")
            trace.append("ps")
            return (dict(ps_record),)

        def inspect(self, container_id: str) -> dict[str, object]:
            assert container_id == _CONTAINER_ID
            trace.append("inspect")
            return terminal

    def collect_once() -> dict[str, object]:
        if terminal_seen:
            raise AssertionError("monitor sampled host after terminal die")
        trace.append("host")
        return {**_safe_sample(), "timestamp": "2026-09-01T12:00:00.000000Z"}

    actual_concurrency_sample = runtime._concurrency_sample

    def concurrency_once(
        records: tuple[dict[str, object], ...], timestamp: object, bound_id: object
    ) -> dict[str, object]:
        if terminal_seen:
            raise AssertionError("monitor sampled concurrency after terminal die")
        trace.append("concurrency")
        return actual_concurrency_sample(records, timestamp, bound_id)

    runtime._adapter = PostDieAdapter()  # type: ignore[assignment]
    runtime._since = _START
    runtime._preflight = (_safe_sample(),)
    runtime._concurrency_sample = concurrency_once  # type: ignore[method-assign]
    monkeypatch.setattr(resources, "collect_host_sample", collect_once)

    observed = runtime.monitor(_candidate())

    assert observed == (
        {**_safe_sample(), "timestamp": "2026-09-01T12:00:00.000000Z", "phase": "runtime"},
    )
    assert runtime._terminal == expected_terminal
    assert runtime._terminal is not terminal
    assert runtime._terminal["Config"] is not terminal_config
    assert runtime._terminal["Config"]["Labels"] is not terminal_labels
    assert runtime._terminal["Mounts"] is not terminal_mounts
    assert runtime._terminal["Mounts"][0] is not terminal_mount
    terminal_labels["evidence"] = "mutated-after-die"
    terminal_mount["Destination"] = "/mutated-after-die"
    assert runtime._terminal == expected_terminal
    assert runtime._terminal_die_session is session
    assert runtime._event_session is None
    assert session.close_calls == 1
    assert trace == ["open", "host", "stats", "ps", "concurrency", "poll:die", "inspect", "close"]


def test_monitor_seals_backend_unit_overlap_and_never_stops_a_foreign_run_twice(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    """Only the bound backend-unit may be stopped when another run overlaps it."""

    foreign_id = "f" * 64

    def ps_record(container_id: str, name: str) -> dict[str, str]:
        record = {field: "" for field in observe._PS_FIELDS}
        record.update(
            {
                "ID": container_id,
                "Image": _BACKEND_IMAGE_ID,
                "Labels": "com.docker.compose.project=worktree-master,com.docker.compose.service=backend-unit",
                "Names": name,
            }
        )
        return record

    records = (
        ps_record(_CONTAINER_ID, "worktree-master-backend-unit-1"),
        ps_record(foreign_id, "worktree-master-backend-unit-2"),
    )
    session = _InterruptSession((_die_event(1_001),))
    stats = {
        "BlockIO": "0B / 0B",
        "CPUPerc": "0.00%",
        "Container": _HOSTNAME,
        "ID": _CONTAINER_ID,
        "MemPerc": "0.00%",
        "MemUsage": "0B / 2GiB",
        "Name": "backend-unit",
        "NetIO": "0B / 0B",
        "PIDs": "0",
    }

    class OverlapAdapter(_ObserverTerminalAdapter):
        def open_bound_events(self, since: str) -> _InterruptSession:
            assert since == _START
            return session

        def stats(self, container_id: str) -> dict[str, str]:
            assert container_id == _CONTAINER_ID
            return dict(stats)

        def ps(self) -> tuple[dict[str, str], ...]:
            return tuple(dict(record) for record in records)

    runtime = observe._SystemRuntime()
    adapter = OverlapAdapter({"Id": _CONTAINER_ID})
    candidate = _candidate()
    runtime._adapter = adapter
    runtime._since = _START
    runtime._preflight = (_safe_sample(),)
    _freeze_interrupt_clock(monkeypatch, [])
    monkeypatch.setattr(
        resources,
        "collect_host_sample",
        lambda: {**_safe_sample(), "timestamp": "2026-09-01T12:00:00.000000Z"},
    )

    assert runtime.monitor(candidate) == (
        {**_safe_sample(), "timestamp": "2026-09-01T12:00:00.000000Z", "phase": "runtime"},
    )
    assert runtime._concurrency == (
        {
            "phase": "runtime",
            "timestamp": "2026-09-01T12:00:00.000000Z",
            "backend_unit_count": 2,
            "mail_calendar_qa_test_build_count": 0,
            "buildkit_state": "idle",
        },
    )
    assert runtime._raw["prohibited_backend_unit_overlap"] is True
    settled = observe._settled_record(
        tuple((*runtime._preflight, *runtime._runtime_samples)),
        runtime.raw_observations_for(candidate, runtime._runtime_samples),
    )
    assert settled["classification"] == "SEALED_INVALID_QA_OVERLAP"
    assert settled["selected_stop"] == ("qa", (2**63) - 1)
    assert settled["explicit_stops"] == (("qa", (2**63) - 1),)
    assert runtime._stop_evidence is not None
    assert runtime._stop_evidence["selected"] == ("qa", (2**63) - 1)
    assert runtime._stop_evidence["argv"] == (
        "docker",
        "stop",
        "--timeout",
        "5",
        _CONTAINER_ID,
    )
    assert runtime._guarded_stop_count == 1
    assert adapter.stop_calls == [_CONTAINER_ID]

    repeated = runtime._concurrency_sample(
        records,
        "2026-09-01T12:00:00.000000Z",
        _CONTAINER_ID,
    )
    assert repeated["backend_unit_count"] == 2
    runtime._issue_selected_stop(candidate, settled)
    with pytest.raises(observe.InvalidEvidence, match="original interrupted event session is unavailable"):
        runtime.recover_interrupt(candidate)
    assert runtime._guarded_stop_count == 1
    assert adapter.stop_calls == [_CONTAINER_ID]

    with pytest.raises(observe.InvalidEvidence, match="does not contain the bound backend-unit"):
        observe._SystemRuntime()._concurrency_sample(
            (ps_record(foreign_id, "worktree-master-backend-unit-2"),),
            "2026-09-01T12:00:00.000000Z",
            _CONTAINER_ID,
        )


def test_monitor_interrupt_retains_its_original_session_for_recovery(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    runtime = observe._SystemRuntime()
    session = _InterruptSession((_die_event(1_001),))
    adapter = _ObserverTerminalAdapter({"Id": _CONTAINER_ID})
    adapter.open_bound_events = lambda since: session
    runtime._adapter = adapter
    runtime._since = _START
    runtime._preflight = (_safe_sample(),)

    def raise_interrupt() -> dict[str, object]:
        raise KeyboardInterrupt

    monkeypatch.setattr(resources, "collect_host_sample", raise_interrupt)
    with pytest.raises(KeyboardInterrupt):
        runtime.monitor(_candidate())

    assert runtime._event_session is session
    assert runtime._original_event_session is session
    assert session.close_calls == 0
    _freeze_interrupt_clock(monkeypatch, [])
    runtime.recover_interrupt(_candidate())
    assert session.close_calls == 1


@pytest.mark.parametrize("die_time_nano", (999, 1_000))
def test_recover_interrupt_rejects_die_before_or_at_guarded_stop_boundary(
    die_time_nano: int, monkeypatch: pytest.MonkeyPatch
) -> None:
    runtime, session, adapter, candidate = _interrupted_runtime((_die_event(die_time_nano),))
    _freeze_interrupt_clock(monkeypatch, [])

    with pytest.raises(observe.InvalidEvidence, match="terminal die preceded"):
        runtime.recover_interrupt(candidate)

    assert adapter.stop_calls == [_CONTAINER_ID]
    assert session.close_calls == 1


def test_recover_interrupt_never_retries_an_in_flight_guarded_stop(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    runtime, session, adapter, candidate = _interrupted_runtime((_die_event(1_001),))
    runtime._stop_evidence = {
        "state": "in_flight",
        "selected": ("interrupt", 1_000),
        "argv": ("docker", "stop", "--timeout", "5", _CONTAINER_ID),
        "start_time_nano": 1_000,
    }
    runtime._guarded_stop_count = 1
    _freeze_interrupt_clock(monkeypatch, [])

    with pytest.raises(observe.InvalidEvidence, match="attempt did not complete"):
        runtime.recover_interrupt(candidate)

    assert adapter.stop_calls == []
    assert session.close_calls == 1


def test_recover_interrupt_rejects_terminal_evidence_that_precedes_stop(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    runtime, session, adapter, candidate = _interrupted_runtime((_die_event(1_001),))
    runtime._terminal = {"Id": _CONTAINER_ID}
    _freeze_interrupt_clock(monkeypatch, [])

    with pytest.raises(observe.InvalidEvidence, match="after terminal"):
        runtime.recover_interrupt(candidate)

    assert adapter.stop_calls == []
    assert session.close_calls == 1


def test_recover_interrupt_preserves_an_earlier_qa_selection(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    runtime, _session, adapter, candidate = _interrupted_runtime(
        (_die_event(1_001),), raw={"explicit_stops": (("qa", 42),)}
    )
    _freeze_interrupt_clock(monkeypatch, [])

    runtime.recover_interrupt(candidate)

    assert runtime._stop_evidence is not None
    assert runtime._stop_evidence["selected"] == ("qa", 42)
    assert runtime._raw["explicit_stops"] == (("qa", 42), ("interrupt", 1_000))
    assert adapter.stop_calls == [_CONTAINER_ID]


def test_recover_interrupt_waits_briefly_for_a_stable_event_stream(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    runtime, session, _adapter, candidate = _interrupted_runtime((None, _die_event(1_001)))
    sleeps: list[float] = []
    _freeze_interrupt_clock(monkeypatch, sleeps)

    runtime.recover_interrupt(candidate)

    assert session.poll_calls == 2
    assert sleeps == [0.01]


def _terminal_candidate() -> dict[str, object]:
    return {
        "container_id": _CONTAINER_ID,
        "container_image": _BACKEND_IMAGE_ID,
        "container_hostname": _HOSTNAME,
        "run_id": _RUN_ID,
        "owner_nonce": _OWNER_NONCE,
        "run_dir": Path("/reports/backend-unit-run.partial"),
        "owner_bytes": b"{}",
    }


def _aborted_terminal_runtime(
    monkeypatch: pytest.MonkeyPatch,
    *,
    terminal: dict[str, object] | None = None,
    markers: tuple[str, ...] = ("CONTAINER_ABORTED",),
    marker_payload: bytes = b"",
    marker_read_error: bool = False,
    receipt: dict[str, object] | None = None,
) -> tuple[observe._SystemRuntime, dict[str, object], _ObserverTerminalAdapter]:
    payload, _final, _observer = _aborted_publication_payloads()
    document = json.loads(payload) if receipt is None else deepcopy(receipt)
    payload = _canonical_receipt_bytes(document)
    runtime = observe._SystemRuntime()
    candidate = _terminal_candidate()
    adapter = _ObserverTerminalAdapter(
        terminal
        if terminal is not None
        else {
            "Id": _CONTAINER_ID,
            "Image": _BACKEND_IMAGE_ID,
            "State": {
                "Status": "exited",
                "ExitCode": 143,
                "OOMKilled": False,
                "Error": "",
            },
        }
    )
    session = object()
    runtime._adapter = adapter
    runtime._lease = _ObserverTerminalLease()
    runtime._fixed_images = adapter.inspect_fixed_images()
    runtime._manifest = _MANIFEST_BYTES
    runtime._terminal = deepcopy(adapter.terminal)
    runtime._stop_evidence = {
        "state": "confirmed",
        "selected": ("qa", 42),
        "argv": ("docker", "stop", "--timeout", "5", _CONTAINER_ID),
        "result": {"result": "stopped"},
        "start_time_nano": 1_000,
    }
    runtime._guarded_stop_count = 1
    runtime._original_event_session = session
    runtime._terminal_die_session = session
    runtime._terminal_die_time_nano = 1_001
    runtime._events = ({"event": "die", "timestamp": observe._event_timestamp(1_001)},)

    def read_run(_run_dir: Path, relative: str, _maximum: int) -> bytes:
        if relative == "owner.json":
            return b"{}"
        if relative == "container-receipt.json":
            return payload
        if relative == "CONTAINER_ABORTED":
            if marker_read_error:
                raise observe.InvalidEvidence("INVALID_EVIDENCE: marker is nonregular")
            return marker_payload
        raise AssertionError(f"unexpected aborted artifact read: {relative}")

    runtime._read_run = read_run
    monkeypatch.setattr(
        safe_paths,
        "list_directory_relative",
        lambda _parent, _name: markers,
    )
    return runtime, candidate, adapter


def _aborted_record() -> dict[str, object]:
    return {"selected_stop": ("qa", 42), "explicit_stops": (("qa", 42),)}


def test_terminal_verify_accepts_aborted_receipt_only_with_exact_guarded_die(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    runtime, candidate, _adapter = _aborted_terminal_runtime(monkeypatch)

    runtime.terminal_verify(candidate, _aborted_record())

    assert candidate["container_receipt"]["state"] == "CONTAINER_ABORTED"
    assert candidate["cleanup"] == {"container_removed": False}


@pytest.mark.parametrize(
    "failure",
    (
        "missing_stop",
        "mismatched_stop",
        "missing_die",
        "foreign_die_session",
        "die_at_stop_boundary",
        "running_state",
        "wrong_exit_code",
        "oom_killed",
        "terminal_error",
        "receipt_terminal",
        "receipt_state",
    ),
)
def test_terminal_verify_rejects_incomplete_or_mismatched_aborted_evidence(
    failure: str, monkeypatch: pytest.MonkeyPatch
) -> None:
    receipt: dict[str, object] | None = None
    if failure in {"receipt_terminal", "receipt_state"}:
        payload, _final, _observer = _aborted_publication_payloads()
        receipt = json.loads(payload)
        if failure == "receipt_terminal":
            receipt["terminal"]["exit_code"] = 0
        else:
            receipt["state"] = "CONTAINER_COMPLETE"
    runtime, candidate, _adapter = _aborted_terminal_runtime(
        monkeypatch, receipt=receipt
    )
    if failure == "missing_stop":
        runtime._stop_evidence = None
    elif failure == "mismatched_stop":
        assert runtime._stop_evidence is not None
        runtime._stop_evidence["selected"] = ("ram", 42)
    elif failure == "missing_die":
        runtime._terminal_die_time_nano = None
    elif failure == "foreign_die_session":
        runtime._terminal_die_session = object()
    elif failure == "die_at_stop_boundary":
        runtime._terminal_die_time_nano = 1_000
        runtime._events = (
            {"event": "die", "timestamp": observe._event_timestamp(1_000)},
        )
    elif failure == "running_state":
        assert runtime._terminal is not None
        runtime._terminal["State"] = {
            "Status": "running", "ExitCode": 143, "OOMKilled": False, "Error": ""
        }
    elif failure == "wrong_exit_code":
        assert runtime._terminal is not None
        runtime._terminal["State"]["ExitCode"] = 0
    elif failure == "oom_killed":
        assert runtime._terminal is not None
        runtime._terminal["State"]["OOMKilled"] = True
    elif failure == "terminal_error":
        assert runtime._terminal is not None
        runtime._terminal["State"]["Error"] = "engine error"

    with pytest.raises(observe.InvalidEvidence):
        runtime.terminal_verify(candidate, _aborted_record())


@pytest.mark.parametrize(
    ("markers", "marker_payload", "marker_read_error"),
    (
        (("CONTAINER_ABORTED", "CONTAINER_COMPLETE"), b"", False),
        (("CONTAINER_COMPLETE",), b"", False),
        ((), b"", False),
        (("CONTAINER_ABORTED",), b"not-empty", False),
        (("CONTAINER_ABORTED",), b"", True),
    ),
)
def test_terminal_verify_rejects_ambiguous_or_unsafe_aborted_marker(
    markers: tuple[str, ...],
    marker_payload: bytes,
    marker_read_error: bool,
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    runtime, candidate, _adapter = _aborted_terminal_runtime(
        monkeypatch,
        markers=markers,
        marker_payload=marker_payload,
        marker_read_error=marker_read_error,
    )

    with pytest.raises(observe.InvalidEvidence):
        runtime.terminal_verify(candidate, _aborted_record())


def test_terminal_verify_rejects_complete_receipt_with_a_selected_stop(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    payload, _final, _observer = _publication_payloads()
    runtime = observe._SystemRuntime()
    candidate = _terminal_candidate()
    runtime._terminal = {
        "Id": _CONTAINER_ID,
        "Image": _BACKEND_IMAGE_ID,
        "State": {"Status": "exited", "ExitCode": 0, "OOMKilled": False, "Error": ""},
    }
    runtime._read_run = lambda _run_dir, relative, _maximum: (
        b"{}" if relative == "owner.json" else payload
    )
    monkeypatch.setattr(
        safe_paths,
        "list_directory_relative",
        lambda _parent, _name: ("CONTAINER_COMPLETE",),
    )

    with pytest.raises(observe.InvalidEvidence, match="complete receipt cannot bind"):
        runtime.terminal_verify(candidate, _aborted_record())

@pytest.mark.parametrize(
    ('aborted', 'marker'),
    ((False, 'CONTAINER_COMPLETE'), (True, 'CONTAINER_ABORTED')),
)
def test_publication_accepts_one_empty_state_selected_terminal_marker(
    aborted: bool, marker: str
) -> None:
    if aborted:
        container_payload, final, observer = _aborted_publication_payloads()
    else:
        container_payload, final, observer = _publication_payloads()
    partial = Path('/reports/backend-unit-run.partial')
    writer = _PublicationWriter(
        payloads={'container-receipt.json': container_payload},
        terminal_markers=(marker,),
    )

    assert publication.publish(partial, final, observer, writer=writer) == Path(
        '/reports/backend-unit-run'
    )
    final_dir = Path('/reports/backend-unit-run')
    container = json.loads(container_payload)
    inventory = container['artifact_inventory']
    assert type(inventory) is list
    declared_names = {record['path'] for record in inventory}
    assert set(path.name for path in writer.path_payloads if path.parent == final_dir) == (
        declared_names
        | {'container-receipt.json', 'final-receipt.json', 'observer-receipt.json', 'SEAL'}
    )
    for record in inventory:
        payload = writer.path_payloads[final_dir / record['path']]
        assert _sha256(payload) == record['sha256']
        assert len(payload) == record['size']
    assert writer.path_payloads[final_dir / marker] == b''
    assert writer.terminal_marker_events == [
        ('list_terminal_markers', 'backend-unit-run.partial')
    ] * 10
    assert writer.tree_entry_events == [
        ('list_entries', 'backend-unit-run.partial')
    ] * 6
    assert [name for event, name in writer.events if event == 'write'] == [
        'final-receipt.json', 'observer-receipt.json', 'SEAL'
    ]
    assert writer.events[-1] == (
        'rename', 'backend-unit-run.partial->backend-unit-run'
    )


@pytest.mark.parametrize(
    ('aborted', 'markers', 'marker_payloads'),
    (
        (False, ('CONTAINER_COMPLETE', 'CONTAINER_COMPLETE'), {}),
        (False, ('CONTAINER_ABORTED',), {}),
        (True, ('CONTAINER_COMPLETE',), {}),
        (False, ('CONTAINER_COMPLETE', 'CONTAINER_ABORTED'), {}),
        (False, (), {}),
        (False, ('CONTAINER_COMPLETE',), {'CONTAINER_COMPLETE': b'x'}),
    ),
)
def test_publication_rejects_wrong_opposite_both_missing_or_nonempty_terminal_marker_before_write(
    aborted: bool, markers: tuple[str, ...], marker_payloads: dict[str, bytes]
) -> None:
    if aborted:
        container_payload, final, observer = _aborted_publication_payloads()
    else:
        container_payload, final, observer = _publication_payloads()
    writer = _PublicationWriter(
        payloads={'container-receipt.json': container_payload},
        terminal_markers=markers,
        terminal_marker_payloads=marker_payloads,
    )

    with pytest.raises(publication.PublicationError):
        publication.publish(Path('/reports/backend-unit-run.partial'), final, observer, writer=writer)

    assert not any(event[0] in {'write', 'remove', 'rename'} for event in writer.events)
    assert 'SEAL' not in writer.payloads


@pytest.mark.parametrize('failure', ('listing', 'path'))
def test_publication_rejects_terminal_marker_listing_or_path_failure_before_write(
    failure: str,
) -> None:
    container_payload, final, observer = _publication_payloads()
    writer = _PublicationWriter(
        payloads={'container-receipt.json': container_payload},
        terminal_marker_list_fails=failure == 'listing',
        terminal_marker_path_failures=(
            frozenset({'CONTAINER_COMPLETE'}) if failure == 'path' else frozenset()
        ),
    )

    with pytest.raises(publication.PublicationError):
        publication.publish(Path('/reports/backend-unit-run.partial'), final, observer, writer=writer)

    assert not any(event[0] in {'write', 'remove', 'rename'} for event in writer.events)
    assert 'SEAL' not in writer.payloads


def test_publication_rejects_same_byte_terminal_marker_identity_swap_before_rename() -> None:
    container_payload, final, observer = _publication_payloads()
    writer = _PublicationWriter(
        payloads={'container-receipt.json': container_payload},
        identity_replacements={
            'CONTAINER_COMPLETE': (7, 99, 0o100600, 1, 0, 1_000_099, 2_000_099)
        },
        identity_swap_after_reads={'CONTAINER_COMPLETE': 6},
    )

    with pytest.raises(publication.PublicationError):
        publication.publish(Path('/reports/backend-unit-run.partial'), final, observer, writer=writer)

    assert ('write', 'SEAL') in writer.events
    assert writer.events[-1] == ('remove', 'SEAL')
    assert writer.path_payloads[
        Path('/reports/backend-unit-run.partial') / 'CONTAINER_COMPLETE'
    ] == b''
    assert 'SEAL' not in writer.payloads
    assert not any(event[0] == 'rename' for event in writer.events)


def test_publication_rejects_terminal_marker_set_mutation_after_first_check() -> None:
    container_payload, final, observer = _publication_payloads()
    writer = _PublicationWriter(
        payloads={'container-receipt.json': container_payload},
        terminal_marker_mutation_after_lists=1,
        terminal_marker_replacement=('CONTAINER_ABORTED',),
    )

    with pytest.raises(publication.PublicationError):
        publication.publish(Path('/reports/backend-unit-run.partial'), final, observer, writer=writer)

    assert writer.terminal_marker_events == [
        ('list_terminal_markers', 'backend-unit-run.partial')
    ] * 2
    assert not any(event[0] in {'write', 'remove', 'rename'} for event in writer.events)


@pytest.mark.parametrize('mutation', ('extra', 'missing'))
def test_publication_rejects_noncanonical_base_tree_before_first_write(
    mutation: str,
) -> None:
    container_payload, final, observer = _publication_payloads()
    partial = Path('/reports/backend-unit-run.partial')
    writer = _PublicationWriter(payloads={'container-receipt.json': container_payload})
    if mutation == 'extra':
        extra = partial / 'foreign-unbound.log'
        writer.path_payloads[extra] = b'foreign'
        writer._assign_opaque_identity(extra)
    else:
        missing = partial / 'stage-build.json'
        writer.path_payloads.pop(missing)
        writer.opaque_file_identities.pop(missing)
    inputs_before_publish = dict(writer.path_payloads)

    with pytest.raises(publication.PublicationError):
        publication.publish(partial, final, observer, writer=writer)

    assert writer.path_payloads == inputs_before_publish
    assert not any(event[0] in {'write', 'remove', 'rename'} for event in writer.events)
    assert 'SEAL' not in writer.payloads


def test_publication_rejects_unavailable_base_tree_listing_before_first_write() -> None:
    container_payload, final, observer = _publication_payloads()
    partial = Path('/reports/backend-unit-run.partial')
    writer = _PublicationWriter(
        payloads={'container-receipt.json': container_payload}, tree_entries_fail=True
    )
    inputs_before_publish = dict(writer.path_payloads)

    with pytest.raises(publication.PublicationError):
        publication.publish(partial, final, observer, writer=writer)

    assert writer.path_payloads == inputs_before_publish
    assert writer.tree_entry_events == [('list_entries', partial.name)]
    assert not any(event[0] in {'write', 'remove', 'rename'} for event in writer.events)


@pytest.mark.parametrize(
    'replacement', ('identity', 'nlink', 'content_hash', 'size')
)
def test_publication_rejects_base_artifact_replacement_after_initial_validation(
    replacement: str,
) -> None:
    container_payload, final, observer = _publication_payloads()
    partial = Path('/reports/backend-unit-run.partial')
    target = partial / 'stage-build.json'
    writer = _PublicationWriter(payloads={'container-receipt.json': container_payload})
    identity = writer.opaque_file_identities[target]
    if replacement == 'identity':
        writer.identity_replacements['stage-build.json'] = (
            identity[0], identity[1] + 100, identity[2], identity[3], identity[4],
            identity[5] + 100, identity[6] + 100,
        )
        writer.identity_swap_after_reads['stage-build.json'] = 1
    elif replacement == 'nlink':
        writer.identity_replacements['stage-build.json'] = (
            identity[0], identity[1], identity[2], identity[3] + 1, identity[4],
            identity[5], identity[6],
        )
        writer.identity_swap_after_reads['stage-build.json'] = 1
    elif replacement == 'content_hash':
        writer.payload_replacements['stage-build.json'] = b'[]'
        writer.payload_swap_after_reads['stage-build.json'] = 1
    else:
        writer.payload_replacements['stage-build.json'] = b'{"size":7}'
        writer.payload_swap_after_reads['stage-build.json'] = 1
    input_paths_before_publish = frozenset(writer.path_payloads)

    with pytest.raises(publication.PublicationError):
        publication.publish(partial, final, observer, writer=writer)

    assert frozenset(writer.path_payloads) == input_paths_before_publish
    assert target in writer.path_payloads
    assert not any(event[0] in {'write', 'remove', 'rename'} for event in writer.events)
    assert 'SEAL' not in writer.payloads


@pytest.mark.parametrize(
    ('mutation_after_lists', 'expected_tree_checks', 'writes_before_rejection'),
    ((2, 3, ()), (4, 5, ('final-receipt.json', 'observer-receipt.json', 'SEAL'))),
)
def test_publication_rejects_tree_mutation_before_first_write_or_pre_rename(
    mutation_after_lists: int,
    expected_tree_checks: int,
    writes_before_rejection: tuple[str, ...],
) -> None:
    container_payload, final, observer = _publication_payloads()
    partial = Path('/reports/backend-unit-run.partial')
    writer = _PublicationWriter(
        payloads={'container-receipt.json': container_payload},
        tree_mutation_after_lists=mutation_after_lists,
        tree_extra_entry='foreign-unbound.log',
    )
    inputs_before_publish = dict(writer.path_payloads)

    with pytest.raises(publication.PublicationError):
        publication.publish(partial, final, observer, writer=writer)

    assert writer.tree_entry_events == [('list_entries', partial.name)] * expected_tree_checks
    assert [name for event, name in writer.events if event == 'write'] == list(
        writes_before_rejection
    )
    assert not any(event == 'rename' for event, _name in writer.events)
    assert all(writer.path_payloads[path] == payload for path, payload in inputs_before_publish.items())
    assert not any(
        event == 'remove' and name in {path.name for path in inputs_before_publish}
        for event, name in writer.events
    )
    if writes_before_rejection:
        assert [name for event, name in writer.events if event == 'remove'] == ['SEAL']
        assert 'SEAL' not in writer.payloads
        assert set(writer.payloads) == {
            'container-receipt.json', 'final-receipt.json', 'observer-receipt.json'
        }
    else:
        assert not any(event == 'remove' for event, _name in writer.events)
        assert set(writer.payloads) == {'container-receipt.json'}


def test_publication_validates_canonical_container_before_write_then_rereads_seals_and_renames() -> None:
    container_payload, final, observer = _publication_payloads()
    writer = _PublicationWriter(payloads={"container-receipt.json": container_payload})
    validation_events: list[tuple[tuple[str, str], ...]] = []

    def validate(receipt: dict[str, Any], observed: dict[str, Any]) -> None:
        assert receipt["run_id"] == _RUN_ID
        assert observed["container_id"] == _CONTAINER_ID
        validation_events.append(tuple(writer.events))

    final_dir = publication.publish(
        Path("/reports/backend-unit-run.partial"),
        final,
        observer,
        validate=validate,
        writer=writer,
    )

    assert container_payload == _canonical_receipt_bytes(json.loads(container_payload))
    assert validation_events == [
        (
            ("assert_terminal", "backend-unit-run.partial"),
            ("assert_absent", "backend-unit-run"),
            ("read", "container-receipt.json"),
        )
    ]
    assert writer.events == [
        ("assert_terminal", "backend-unit-run.partial"),
        ("assert_absent", "backend-unit-run"),
        ("read", "container-receipt.json"),
        ("read", "container-receipt.json"),
        ("list_entries", "backend-unit-run.partial"),
        ("read", "container-receipt.json"),
        ("read", "owner.json"),
        ("read", "stage-build.json"),
        ("list_entries", "backend-unit-run.partial"),
        ("list_entries", "backend-unit-run.partial"),
        ("read", "container-receipt.json"),
        ("read", "owner.json"),
        ("read", "stage-build.json"),
        ("list_entries", "backend-unit-run.partial"),
        ("write", "final-receipt.json"),
        ("read", "final-receipt.json"),
        ("write", "observer-receipt.json"),
        ("read", "observer-receipt.json"),
        ("write", "SEAL"),
        ("read", "SEAL"),
        ("read", "final-receipt.json"),
        ("read", "observer-receipt.json"),
        ("read", "container-receipt.json"),
        ("read", "SEAL"),
        ("list_entries", "backend-unit-run.partial"),
        ("read", "SEAL"),
        ("read", "container-receipt.json"),
        ("read", "final-receipt.json"),
        ("read", "observer-receipt.json"),
        ("read", "owner.json"),
        ("read", "stage-build.json"),
        ("list_entries", "backend-unit-run.partial"),
        ("rename", "backend-unit-run.partial->backend-unit-run"),
    ]
    assert writer.identity_events == [
        event
        for event in writer.events
        if event[0] not in {'assert_terminal', 'assert_absent', 'list_entries'}
    ]
    assert final_dir == Path("/reports/backend-unit-run")
    assert writer.payloads["final-receipt.json"] == _canonical_receipt_bytes(final)
    assert writer.payloads["observer-receipt.json"] == _canonical_receipt_bytes(observer)
    assert writer.payloads["SEAL"] == (
        b"hub.qa.backend-unit-seal/v1\n"
        + f"run_id={_RUN_ID}\n".encode()
        + f"owner_nonce={_OWNER_NONCE}\n".encode()
        + b"observer_receipt_sha256="
        + _sha256(writer.payloads["observer-receipt.json"]).encode()
        + b"\n"
    )


def test_publication_rejects_after_container_stability_before_any_write() -> None:
    container_payload, final, observer = _publication_payloads()
    writer = _PublicationWriter(payloads={"container-receipt.json": container_payload})

    def reject(_final: dict[str, Any], _observer: dict[str, Any]) -> None:
        raise publication.PublicationError("instrumentation")

    with pytest.raises(publication.PublicationError, match="instrumentation"):
        publication.publish(
            Path("/reports/backend-unit-run.partial"),
            final,
            observer,
            validate=reject,
            writer=writer,
        )

    assert writer.events == [
        ("assert_terminal", "backend-unit-run.partial"),
        ("assert_absent", "backend-unit-run"),
        ("read", "container-receipt.json"),
    ]
    assert writer.payloads == {"container-receipt.json": container_payload}


@pytest.mark.parametrize(
    ("active_container_writer", "existing_final", "expected_events"),
    (
        (True, False, [("assert_terminal", "backend-unit-run.partial")]),
        (
            False,
            True,
            [
                ("assert_terminal", "backend-unit-run.partial"),
                ("assert_absent", "backend-unit-run"),
            ],
        ),
    ),
)
def test_publication_preserves_active_or_foreign_existing_final(
    active_container_writer: bool,
    existing_final: bool,
    expected_events: list[tuple[str, str]],
) -> None:
    container_payload, final, observer = _publication_payloads()
    writer = _PublicationWriter(
        payloads={"container-receipt.json": container_payload},
        active_container_writer=active_container_writer,
        existing_final=existing_final,
        foreign_final=b"foreign-final-sentinel",
    )
    foreign_final = b"foreign-final-sentinel"

    with pytest.raises(publication.PublicationError):
        publication.publish(
            Path("/reports/backend-unit-run.partial"),
            final,
            observer,
            validate=lambda _final, _observer: None,
            writer=writer,
        )

    assert writer.events == expected_events
    assert writer.payloads == {"container-receipt.json": container_payload}
    assert not writer.identity_events
    if existing_final:
        assert writer.path_payloads[writer.foreign_final_path] == foreign_final
        assert writer.foreign_final_path not in writer.touched_paths
        assert len(writer.opaque_file_identities[writer.foreign_final_path]) == 7


def test_publication_cross_directory_failure_cleans_only_owned_seal() -> None:
    container_payload, final, observer = _publication_payloads()
    writer = _PublicationWriter(
        payloads={"container-receipt.json": container_payload},
        cross_directory=True,
    )

    with pytest.raises(publication.PublicationError, match="publication writer rename outcome"):
        publication.publish(
            Path("/reports/backend-unit-run.partial"),
            final,
            observer,
            validate=lambda _final, _observer: None,
            writer=writer,
        )

    assert writer.events[-2:] == [
        ("rename", "backend-unit-run.partial->backend-unit-run"),
        ("remove", "SEAL"),
    ]
    assert writer.identity_events[-2:] == [
        ("rename", "backend-unit-run.partial->backend-unit-run"),
        ("remove", "SEAL"),
    ]
    assert "SEAL" not in writer.payloads
    assert set(writer.payloads) == {
        "container-receipt.json",
        "final-receipt.json",
        "observer-receipt.json",
    }


@pytest.mark.parametrize(
    "swapped_name",
    (
        "container-receipt.json",
        "final-receipt.json",
        "observer-receipt.json",
        "SEAL",
    ),
)
def test_publication_rejects_identity_or_byte_swap_without_authoritative_rename(
    swapped_name: str,
) -> None:
    container_payload, final, observer = _publication_payloads()
    writer = _PublicationWriter(
        payloads={"container-receipt.json": container_payload},
        read_swaps={swapped_name: b"{}\n"},
    )

    with pytest.raises(publication.PublicationError):
        publication.publish(
            Path("/reports/backend-unit-run.partial"),
            final,
            observer,
            validate=lambda _final, _observer: None,
            writer=writer,
        )

    assert not any(event[0] == "rename" for event in writer.events)
    if swapped_name == "SEAL":
        assert writer.events[-1] == ("remove", "SEAL")
        assert "SEAL" not in writer.payloads
    else:
        assert "SEAL" not in writer.payloads
        assert not any(event[0] == "remove" for event in writer.events)


def test_publication_rejects_same_bytes_different_opaque_file_identity_before_seal() -> None:
    container_payload, final, observer = _publication_payloads()
    partial = Path("/reports/backend-unit-run.partial")
    container_path = partial / "container-receipt.json"
    writer = _PublicationWriter(
        payloads={"container-receipt.json": container_payload},
        identity_replacements={
            "container-receipt.json": (
                7,
                99,
                0o100600,
                1,
                len(container_payload),
                1_000_099,
                2_000_099,
            )
        },
        identity_swap_after_reads={"container-receipt.json": 1},
    )

    with pytest.raises(publication.PublicationError):
        publication.publish(
            partial,
            final,
            observer,
            validate=lambda _final, _observer: None,
            writer=writer,
        )

    assert writer.path_payloads[container_path] == container_payload
    assert [name for _event, name, _identity in writer.identity_checks] == [
        "container-receipt.json",
        "CONTAINER_COMPLETE",
        "container-receipt.json",
    ]
    first_identity = writer.identity_checks[0][2]
    replaced_identity = writer.identity_checks[-1][2]
    assert first_identity[4] == len(container_payload)
    assert replaced_identity == (
        7,
        99,
        0o100600,
        1,
        len(container_payload),
        1_000_099,
        2_000_099,
    )
    assert first_identity != replaced_identity
    assert writer.events == [
        ("assert_terminal", "backend-unit-run.partial"),
        ("assert_absent", "backend-unit-run"),
        ("read", "container-receipt.json"),
        ("read", "container-receipt.json"),
    ]
    assert "SEAL" not in writer.payloads
    assert not any(event[0] in {"write", "remove", "rename"} for event in writer.events)


@pytest.mark.parametrize("binding", ("run", "nonce", "manifest", "image"))
def test_publication_rejects_run_nonce_manifest_and_image_binding_before_write(
    binding: str,
) -> None:
    container_payload, final, observer = _publication_payloads()
    if binding == "run":
        final["run_id"] = f"backend-unit-20260901T120001Z-{_HOSTNAME}-a1b2c3d4"
    elif binding == "nonce":
        final["owner_nonce"] = "f" * 32
    elif binding == "manifest":
        final["canonical_build_input_manifest_sha256"] = "sha256:" + "f" * 64
    else:
        final["backend_image_id"] = "sha256:" + "f" * 64

    writer = _PublicationWriter(payloads={"container-receipt.json": container_payload})
    with pytest.raises(publication.PublicationError):
        publication.publish(
            Path("/reports/backend-unit-run.partial"),
            final,
            observer,
            validate=lambda _final, _observer: None,
            writer=writer,
        )

    assert writer.events == [
        ("assert_terminal", "backend-unit-run.partial"),
        ("assert_absent", "backend-unit-run"),
    ]
    assert writer.payloads == {"container-receipt.json": container_payload}


@pytest.mark.parametrize(
    ("raw_observations", "expected", "expected_stop_calls"),
    (
        ({}, "SEALED_EXECUTABLE_PASS", 0),
        ({"discovery_failure": True}, "SEALED_EXECUTABLE_FAIL_DISCOVERY", 0),
        ({"stage_failure": True}, "SEALED_EXECUTABLE_FAIL_STAGE", 0),
        ({"timeout": True}, "SEALED_TIMEOUT_UNTRIAGED", 0),
        ({"oom": True}, "SEALED_INFRA_ABORTED_OOM", 0),
        ({"resource": True}, "SEALED_INFRA_ABORTED_RESOURCE_CONTRACT", 0),
        ({"budget": True}, "SEALED_INFRA_ABORTED_REPORT_BUDGET", 0),
        ({"tmpfs": True}, "SEALED_INFRA_ABORTED_TMPFS", 0),
        (
            {"explicit_stops": (("ram", 100),)},
            "SEALED_INFRA_ABORTED_HOST_RAM_FLOOR",
            1,
        ),
        (
            {"explicit_stops": (("disk", 100),)},
            "SEALED_INFRA_ABORTED_DISK_PRESSURE",
            1,
        ),
        (
            {"explicit_stops": (("deadline", 100),)},
            "SEALED_INFRA_ABORTED_WHOLE_RUN_TIMEOUT",
            1,
        ),
        ({"explicit_stops": (("interrupt", 100),)}, "SEALED_INTERRUPTED", 1),
        ({"explicit_stops": (("qa", 100),)}, "SEALED_INVALID_QA_OVERLAP", 1),
        ({"source_mutated": True}, "SEALED_INVALID_SOURCE_MUTATED", 0),
        ({"invalid_provenance": True}, "SEALED_INVALID_PROVENANCE", 0),
    ),
)
def test_main_derives_every_sealed_classification_from_raw_observations(
    raw_observations: dict[str, Any],
    expected: str,
    expected_stop_calls: int,
    monkeypatch: pytest.MonkeyPatch,
    capsys: pytest.CaptureFixture[str],
) -> None:
    runtime = _RawObserverRuntime(raw_observations=raw_observations)
    monkeypatch.setattr(observe, "_new_system_runtime", lambda: runtime)

    assert observe.main(("--compose-project", _PROJECT)) == 0
    ready = capsys.readouterr().out
    session = ready.removeprefix("READY observer_session=").strip()
    assert ready == f"READY observer_session={session}\n"
    assert len(session) == 32
    assert set(session) <= set("0123456789abcdef")
    expected_events = [
        "bootstrap",
        "preflight",
        "lease.acquire",
        "snapshot.before",
        "images",
        "manifest",
        "ready",
        "bind",
        "monitor",
        "raw",
    ]
    expected_events.extend(["stop"] * expected_stop_calls)
    expected_events.extend(
        ["terminal_verify", "snapshot.after", "publish", "lease.release"]
    )
    assert runtime.events == expected_events
    assert runtime.writes == ["final-receipt.json", "observer-receipt.json", "SEAL"]
    assert runtime.published is not None
    final, observer = runtime.published
    assert final["classification"] == expected
    assert final["run_id"] == observer["run_id"] == _RUN_ID
    assert final["owner_nonce"] == observer["owner_nonce"] == _OWNER_NONCE
    assert final["container_id"] == observer["container_id"] == _CONTAINER_ID
    assert (
        final["backend_image_id"]
        == observer["container_image"]
        == observer["fixed_image_ids"]["backend_unit"]
        == _BACKEND_IMAGE_ID
    )
    assert observer["fixed_image_ids"]["python_contract"] == _PYTHON_IMAGE_ID
    assert observer["canonical_build_input_manifest_bytes"] == _MANIFEST_BYTES.decode()
    assert observer["canonical_build_input_manifest_sha256"] == _MANIFEST_SHA256
    expected_stop_argv = ("docker", "stop", "--timeout", "5", _CONTAINER_ID)
    assert runtime.stop_calls == [expected_stop_argv] * expected_stop_calls
    assert runtime.events.count("stop") == expected_stop_calls


def test_main_coincident_stops_uses_selected_predicate_once_and_receipt_argv(
    monkeypatch: pytest.MonkeyPatch
) -> None:
    runtime = _RawObserverRuntime(
        raw_observations={
            "explicit_stops": (("deadline", 100), ("ram", 100), ("disk", 100), ("qa", 100), ("interrupt", 100)),
            "later_explicit_stops": (("interrupt", 101), ("qa", 102)),
        }
    )
    monkeypatch.setattr(observe, "_new_system_runtime", lambda: runtime)

    assert observe.main(("--compose-project", _PROJECT)) == 0
    assert runtime.stop_calls == [("docker", "stop", "--timeout", "5", _CONTAINER_ID)]
    assert runtime.events.count("stop") == 1
    assert runtime.published is not None
    final, observer = runtime.published
    assert final["classification"] == "SEALED_INTERRUPTED"
    assert observer["selected_stop"] == ("interrupt", 100)
    assert observer["stop_argv"] == ("docker", "stop", "--timeout", "5", _CONTAINER_ID)


@pytest.mark.parametrize(
    ("runtime", "argv", "expected_exit"),
    (
        (_RawObserverRuntime(), ("--wrong", _PROJECT), 64),
        (_RawObserverRuntime(terminal_error=True), ("--compose-project", _PROJECT), 75),
        (_RawObserverRuntime(interrupt_before_bind=True), ("--compose-project", _PROJECT), 130),
        (_RawObserverRuntime(source_drift_after_lease=True), ("--compose-project", _PROJECT), 75),
    ),
)
def test_main_never_seals_malformed_unsealable_or_prebind_interrupt(
    runtime: _RawObserverRuntime, argv: tuple[str, str], expected_exit: int, monkeypatch: pytest.MonkeyPatch
) -> None:
    monkeypatch.setattr(observe, "_new_system_runtime", lambda: runtime)

    assert observe.main(argv) == expected_exit
    assert runtime.writes == []
    assert "publish" not in runtime.events
    assert runtime.stop_calls == []
    assert runtime.events.count("lease.release") <= 1
    if expected_exit in (75, 130):
        assert runtime.events[-1] == "lease.release"


def test_main_rejects_non312_or_nonstdlib_bootstrap_before_ready_or_write(
    monkeypatch: pytest.MonkeyPatch
) -> None:
    for runtime in (_RawObserverRuntime(python_version=(3, 11, 9)), _RawObserverRuntime(stdlib_only=False)):
        monkeypatch.setattr(observe, "_new_system_runtime", lambda: runtime)
        with pytest.raises(observe.BootstrapError):
            observe.main(("--compose-project", _PROJECT))
        assert runtime.events == ["bootstrap"]
        assert runtime.writes == []


class _MainInterruptSystemRuntime(observe._SystemRuntime):
    """Controlled real-type runtime for the main interrupt branch only."""

    def __init__(
        self,
        *,
        earlier_stop: str | None = None,
        terminal_failure: str | None = None,
        in_flight_stop: bool = False,
    ) -> None:
        super().__init__()
        self.events: list[str] = []
        self.published: tuple[dict[str, object], dict[str, object]] | None = None
        self._terminal_failure = terminal_failure
        self._adapter = _ObserverTerminalAdapter(
            {
                "Id": _CONTAINER_ID,
                "Image": _BACKEND_IMAGE_ID,
                "State": {
                    "Status": "exited",
                    "ExitCode": 143,
                    "OOMKilled": False,
                    "Error": "",
                },
            }
        )
        self._lease = _ObserverTerminalLease()
        self._fixed_images = self._adapter.inspect_fixed_images()
        self._manifest = _MANIFEST_BYTES
        self._raw = (
            {} if earlier_stop is None else {"explicit_stops": ((earlier_stop, 42),)}
        )
        self._receipt_bytes, _final, _observer = _aborted_publication_payloads()
        if in_flight_stop:
            selected = earlier_stop or "qa"
            self._raw = {"explicit_stops": ((selected, 42),)}
            self._stop_evidence = {
                "state": "in_flight",
                "selected": (selected, 42),
                "argv": ("docker", "stop", "--timeout", "5", _CONTAINER_ID),
                "start_time_nano": 1_000,
            }
            self._guarded_stop_count = 1
            self._adapter.stop_calls.append(_CONTAINER_ID)

    def bootstrap(self, _argv: tuple[str, ...]) -> None:
        self.events.append("bootstrap")

    def preflight(self) -> tuple[dict[str, object], ...]:
        self.events.append("preflight")
        self._preflight = (_safe_sample(),)
        return self._preflight

    def acquire_lease(self) -> None:
        self.events.append("lease.acquire")

    def snapshot(self, phase: str) -> dict[str, object]:
        self.events.append(f"snapshot.{phase}")
        return {"head": "frozen"}

    def fixed_images(self) -> dict[str, str]:
        self.events.append("images")
        assert self._fixed_images is not None
        return dict(self._fixed_images)

    def canonical_manifest(self) -> tuple[bytes, str]:
        self.events.append("manifest")
        return _MANIFEST_BYTES, _MANIFEST_SHA256

    def validate_handoff(
        self,
        _snapshot: dict[str, object],
        _manifest: bytes,
        _images: dict[str, str],
    ) -> str:
        self.events.append("handoff")
        return "sha256:" + "e" * 64

    def ready(self, _session: str) -> None:
        self.events.append("ready")
        self._since = _START

    def bind_candidate(self, _images: dict[str, str]) -> dict[str, Any]:
        self.events.append("bind")
        return _terminal_candidate()

    def monitor(self, _candidate: dict[str, Any]) -> tuple[dict[str, object], ...]:
        self.events.append("monitor")
        self._runtime_samples = (_safe_sample(),)
        session = _InterruptSession((_die_event(1_001),))
        self._event_session = session
        self._original_event_session = session
        raise KeyboardInterrupt

    def raw_observations_for(
        self,
        _candidate: dict[str, Any],
        _samples: tuple[dict[str, object], ...],
    ) -> dict[str, object]:
        self.events.append("raw")
        return dict(self._raw)

    def stop_bound_once(self, container_id: str) -> dict[str, object]:
        self.events.append("stop")
        return super().stop_bound_once(container_id)

    def recover_interrupt(
        self, candidate: dict[str, object]
    ) -> tuple[dict[str, object], ...]:
        self.events.append("recover")
        samples = super().recover_interrupt(candidate)
        if self._terminal_failure == "missing_stop":
            self._stop_evidence = None
        elif self._terminal_failure == "invalid_stop":
            assert self._stop_evidence is not None
            self._stop_evidence["result"] = {"result": "not-stopped"}
        elif self._terminal_failure == "missing_die":
            self._terminal_die_time_nano = None
        elif self._terminal_failure == "foreign_die":
            self._terminal_die_session = object()
        return samples

    def _read_run(self, _run_dir: Path, relative: str, _maximum: int) -> bytes:
        if relative == "owner.json":
            return b"{}"
        if relative == "container-receipt.json":
            return self._receipt_bytes
        if relative == "CONTAINER_ABORTED":
            return b""
        raise AssertionError(f"unexpected main interrupt artifact read: {relative}")

    def terminal_verify(
        self, candidate: dict[str, Any], record: dict[str, object]
    ) -> None:
        self.events.append("terminal_verify")
        super().terminal_verify(candidate, record)

    def verify_chain(
        self, _before: dict[str, object], _after: dict[str, object]
    ) -> None:
        self.events.append("verify_chain")

    def build_receipts(
        self,
        _candidate: dict[str, object],
        _before: dict[str, object],
        _after: dict[str, object],
        record: dict[str, object],
    ) -> tuple[dict[str, object], dict[str, object]]:
        self.events.append("build")
        self.built_record = dict(record)
        return (
            {"classification": record["classification"]},
            {
                "classification": record["classification"],
                "selected_stop": record["selected_stop"],
            },
        )

    def publish(self, final: dict[str, object], observer: dict[str, object]) -> None:
        self.events.append("publish")
        self.published = (final, observer)

    def release_lease(self) -> None:
        self.events.append("lease.release")


@pytest.mark.parametrize(
    ("earlier_stop", "expected_classification"),
    (
        (None, "SEALED_INTERRUPTED"),
        ("qa", "SEALED_INVALID_QA_OVERLAP"),
        ("disk", "SEALED_INFRA_ABORTED_DISK_PRESSURE"),
        ("ram", "SEALED_INFRA_ABORTED_HOST_RAM_FLOOR"),
        ("deadline", "SEALED_INFRA_ABORTED_WHOLE_RUN_TIMEOUT"),
    ),
)
def test_main_system_runtime_interrupt_reuses_one_stop_and_preserves_earlier_selection(
    earlier_stop: str | None,
    expected_classification: str,
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    runtime = _MainInterruptSystemRuntime(earlier_stop=earlier_stop)
    monkeypatch.setattr(observe, "_new_system_runtime", lambda: runtime)
    monkeypatch.setattr(
        safe_paths,
        "list_directory_relative",
        lambda _parent, _name: ("CONTAINER_ABORTED",),
    )
    monkeypatch.setattr(observe.time, "time_ns", lambda: 1_000)

    assert isinstance(runtime, observe._SystemRuntime)
    assert observe.main(("--compose-project", _PROJECT)) == 0
    assert runtime.events == [
        "bootstrap",
        "preflight",
        "lease.acquire",
        "snapshot.before",
        "images",
        "manifest",
        "handoff",
        "ready",
        "bind",
        "monitor",
        "recover",
        "stop",
        "raw",
        "terminal_verify",
        "snapshot.after",
        "verify_chain",
        "raw",
        "build",
        "publish",
        "lease.release",
    ]
    assert runtime._adapter is not None
    assert runtime._adapter.stop_calls == [_CONTAINER_ID]
    assert runtime.published is not None
    final, observer = runtime.published
    assert final["classification"] == expected_classification
    assert observer["classification"] == expected_classification
    selected = ("interrupt", 1_000) if earlier_stop is None else (earlier_stop, 42)
    assert observer["selected_stop"] == selected
    assert runtime.built_record["selected_stop"] == selected
    assert runtime._raw["explicit_stops"][-1] == ("interrupt", 1_000)
    if earlier_stop is not None:
        assert expected_classification != "SEALED_INTERRUPTED"


@pytest.mark.parametrize(
    "terminal_failure", ("missing_stop", "invalid_stop", "missing_die", "foreign_die")
)
def test_main_system_runtime_interrupt_refuses_unconfirmed_stop_or_die_before_publish(
    terminal_failure: str, monkeypatch: pytest.MonkeyPatch
) -> None:
    runtime = _MainInterruptSystemRuntime(
        earlier_stop="qa", terminal_failure=terminal_failure
    )
    monkeypatch.setattr(observe, "_new_system_runtime", lambda: runtime)
    monkeypatch.setattr(
        safe_paths,
        "list_directory_relative",
        lambda _parent, _name: ("CONTAINER_ABORTED",),
    )
    monkeypatch.setattr(observe.time, "time_ns", lambda: 1_000)

    assert observe.main(("--compose-project", _PROJECT)) == 75
    assert runtime._adapter is not None
    assert runtime._adapter.stop_calls == [_CONTAINER_ID]
    assert "terminal_verify" in runtime.events
    assert "build" not in runtime.events
    assert "publish" not in runtime.events
    assert runtime.published is None
    assert runtime.events[-1] == "lease.release"


def test_main_system_runtime_interrupt_does_not_retry_an_in_flight_stop(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    runtime = _MainInterruptSystemRuntime(earlier_stop="qa", in_flight_stop=True)
    monkeypatch.setattr(observe, "_new_system_runtime", lambda: runtime)
    monkeypatch.setattr(observe.time, "time_ns", lambda: 1_000)

    assert observe.main(("--compose-project", _PROJECT)) == 75
    assert runtime._adapter is not None
    assert runtime._adapter.stop_calls == [_CONTAINER_ID]
    assert runtime.events.count("stop") == 0
    assert "terminal_verify" not in runtime.events
    assert "build" not in runtime.events
    assert "publish" not in runtime.events
    assert runtime.published is None


class _MainBindInterruptSystemRuntime(observe._SystemRuntime):
    def __init__(self, *, recoverable_id: str | None) -> None:
        super().__init__()
        self.events: list[str] = []
        self.recovery_trace: list[tuple[str, str | None]] = []
        self.recoverable_lookup_count = 0
        self.stop_arguments: list[str] = []
        self._recoverable_id = recoverable_id
        self._adapter = _ObserverTerminalAdapter({"Id": _CONTAINER_ID})

    def bootstrap(self, _argv: tuple[str, ...]) -> None:
        self.events.append("bootstrap")

    def preflight(self) -> tuple[dict[str, object], ...]:
        self.events.append("preflight")
        return (_safe_sample(),)

    def acquire_lease(self) -> None:
        self.events.append("lease.acquire")

    def snapshot(self, phase: str) -> dict[str, object]:
        self.events.append(f"snapshot.{phase}")
        return {"head": "frozen"}

    def fixed_images(self) -> dict[str, str]:
        self.events.append("images")
        return {
            "python_contract": _PYTHON_IMAGE_ID,
            "backend_unit": _BACKEND_IMAGE_ID,
        }

    def canonical_manifest(self) -> tuple[bytes, str]:
        self.events.append("manifest")
        return _MANIFEST_BYTES, _MANIFEST_SHA256

    def validate_handoff(
        self,
        _snapshot: dict[str, object],
        _manifest: bytes,
        _images: dict[str, str],
    ) -> str:
        self.events.append("handoff")
        return "sha256:" + "e" * 64

    def ready(self, _session: str) -> None:
        self.events.append("ready")

    def bind_candidate(self, _images: dict[str, str]) -> dict[str, Any]:
        self.events.append("bind")
        raise KeyboardInterrupt

    def recoverable_bound_container_id(self) -> str | None:
        self.events.append("recoverable")
        self.recoverable_lookup_count += 1
        self.recovery_trace.append(("recoverable", self._recoverable_id))
        return self._recoverable_id

    def stop_bound_once(self, container_id: str) -> dict[str, object]:
        self.events.append("stop")
        self.stop_arguments.append(container_id)
        self.recovery_trace.append(("stop", container_id))
        return super().stop_bound_once(container_id)

    def release_lease(self) -> None:
        self.events.append("lease.release")
        self.recovery_trace.append(("lease.release", None))


_RECOVERABLE_BOUND_CONTAINER_ID = "f" * 64


@pytest.mark.parametrize(
    ("recoverable_id", "expected_exit", "expected_stop_arguments"),
    ((None, 130, ()), (_RECOVERABLE_BOUND_CONTAINER_ID, 75, (_RECOVERABLE_BOUND_CONTAINER_ID,))),
)
def test_main_system_runtime_prebind_and_postbind_interrupts_never_publish(
    recoverable_id: str | None,
    expected_exit: int,
    expected_stop_arguments: tuple[str, ...],
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    runtime = _MainBindInterruptSystemRuntime(recoverable_id=recoverable_id)
    monkeypatch.setattr(observe, "_new_system_runtime", lambda: runtime)

    assert isinstance(runtime, observe._SystemRuntime)
    assert observe.main(("--compose-project", _PROJECT)) == expected_exit
    assert runtime._adapter is not None
    assert runtime.events == [
        "bootstrap",
        "preflight",
        "lease.acquire",
        "snapshot.before",
        "images",
        "manifest",
        "handoff",
        "ready",
        "bind",
        "recoverable",
        *(["stop"] if expected_stop_arguments else []),
        "lease.release",
    ]
    assert runtime.recoverable_lookup_count == 1
    assert runtime.stop_arguments == list(expected_stop_arguments)
    assert runtime._adapter.stop_calls == list(expected_stop_arguments)
    assert runtime.recovery_trace == [
        ("recoverable", recoverable_id),
        *(("stop", container_id) for container_id in expected_stop_arguments),
        ("lease.release", None),
    ]
    assert runtime.recovery_trace.index(("recoverable", recoverable_id)) < runtime.recovery_trace.index(
        ("lease.release", None)
    )
    if expected_stop_arguments:
        assert runtime.recovery_trace.index(("recoverable", recoverable_id)) < runtime.recovery_trace.index(
            ("stop", recoverable_id)
        )
    assert "monitor" not in runtime.events
    assert "publish" not in runtime.events
    assert runtime.events[-1] == "lease.release"
