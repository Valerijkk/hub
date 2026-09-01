from __future__ import annotations

import hashlib
import io
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

import pytest

from observers.backend_unit import lease, observe, publication, resources, safe_paths


_PROJECT = "worktree-master"
_CONTAINER_ID = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
_OWNER_NONCE = "a1b2c3d4e5f60718293a4b5c6d7e8f90"
_RUN_ID = "backend-unit-20260901T120000Z-a1b2c3d4e5f6-a1b2c3d4"
_PYTHON_IMAGE_ID = "sha256:" + "c" * 64
_BACKEND_IMAGE_ID = "sha256:" + "d" * 64
_MANIFEST_BYTES = b"docker-compose.local.yml\tsha256:" + b"1" * 64 + b"\n"
_MANIFEST_SHA256 = "sha256:" + hashlib.sha256(_MANIFEST_BYTES).hexdigest()
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


def _safe_sample() -> dict[str, int]:
    return {
        "free_ram_bytes": 2 * 1024**3,
        "free_c_bytes": 250 * 1024**3,
        "PercentDiskTime": 1,
        "CurrentDiskQueueLength": 0,
    }


@dataclass
class _AdversarialPathOps:
    windows: bool = False
    hard_link: bool = False
    nonregular: bool = False
    parent_swap_after_open: bool = False
    rename_race: bool = False
    openat_flags: list[int] = field(default_factory=list)
    create_file_flags: list[int] = field(default_factory=list)
    closes: list[int] = field(default_factory=list)
    writes: list[tuple[str, bytes]] = field(default_factory=list)
    renames: list[tuple[str, str]] = field(default_factory=list)
    _fstat_calls: int = 0

    def openat(self, _directory_fd: int, _name: str, flags: int) -> int:
        self.openat_flags.append(flags)
        return 17

    def create_file_w(self, _path: str, flags: int) -> int:
        self.create_file_flags.append(flags)
        return 18

    def fstat(self, _descriptor: int) -> dict[str, int | bool]:
        self._fstat_calls += 1
        changed_identity = self.parent_swap_after_open and self._fstat_calls > 1
        return {
            "device": 7,
            "inode": 12 if changed_identity else 11,
            "size": 0,
            "regular": not self.nonregular,
            "nlink": 2 if self.hard_link else 1,
            "reparse": False,
        }

    def destination_exists(self, _path: Path) -> bool:
        return False

    def write_new(self, path: Path, payload: bytes) -> Path:
        self.writes.append((path.as_posix(), payload))
        return Path(path.as_posix() + ".tmp")

    def rename_same_directory(self, source: Path, destination: Path) -> None:
        self.renames.append((source.as_posix(), destination.as_posix()))
        if self.rename_race:
            raise OSError("parent directory identity changed before rename")

    def close(self, descriptor: int) -> None:
        self.closes.append(descriptor)


@dataclass
class _SnapshotProbe:
    head: str = "a" * 40
    status: str = ""
    manifest: bytes = _MANIFEST_BYTES
    python_image_id: str = _PYTHON_IMAGE_ID
    backend_image_id: str = _BACKEND_IMAGE_ID
    subject_digest: str = "sha256:" + "f" * 64
    commands: list[tuple[str, ...]] = field(default_factory=list)
    acl_changes: list[str] = field(default_factory=list)

    def git(self, argv: tuple[str, ...]) -> str:
        self.commands.append(argv)
        if argv == ("git", "rev-parse", "HEAD"):
            return self.head + "\n"
        if argv == ("git", "status", "--porcelain=v1", "--untracked-files=all"):
            return self.status
        raise AssertionError(f"unexpected Git command: {argv!r}")

    def canonical_build_input_manifest(self) -> bytes:
        return self.manifest

    def fixed_image_ids(self) -> dict[str, str]:
        return {
            "python_contract": self.python_image_id,
            "backend_unit": self.backend_image_id,
        }

    def frozen_subject_digests(self) -> dict[str, str]:
        return {
            "backend/internal/calls/call_durable_credential_allocation.go": self.subject_digest,
            "backend/tests/calls/call_durable_credential_allocation_test.go": (
                "sha256:0ae525c86a75369bdb5790f26b1308f42574fb545883328da24011d4150b03a9"
            ),
        }


@dataclass
class _HashingWriter:
    events: list[tuple[str, str]] = field(default_factory=list)
    payloads: dict[str, bytes] = field(default_factory=dict)
    active_writer: bool = False
    final_exists: bool = False

    def assert_terminal_and_stable(self, _run_dir: Path) -> None:
        if self.active_writer:
            raise OSError("container remains a writer")

    def assert_absent(self, _final_dir: Path) -> None:
        if self.final_exists:
            raise FileExistsError("final directory exists")

    def write(self, path: Path, payload: bytes) -> None:
        self.payloads[path.name] = payload
        self.events.append(("write", path.name))

    def rename_same_directory(self, partial: Path, final: Path) -> None:
        self.events.append(("rename", f"{partial.name}->{final.name}"))


@dataclass
class _ObserverRuntime:
    classification: str = "SEALED_EXECUTABLE_PASS"
    unsealable: bool = False
    interrupt_before_bind: bool = False
    events: list[str] = field(default_factory=list)
    writes: list[str] = field(default_factory=list)
    published: tuple[dict[str, Any], dict[str, Any]] | None = None

    def bootstrap(self, argv: tuple[str, ...]) -> None:
        self.events.append("bootstrap")
        assert argv == ("--compose-project", _PROJECT)

    def preflight(self) -> tuple[dict[str, int], dict[str, int], dict[str, int]]:
        self.events.append("preflight")
        return (_safe_sample(), _safe_sample(), _safe_sample())

    def acquire_lease(self) -> None:
        self.events.append("lease.acquire")

    def snapshot(self, phase: str) -> dict[str, Any]:
        self.events.append(f"snapshot.{phase}")
        return {
            "head": "a" * 40,
            "status": "",
            "canonical_build_input_manifest_bytes": _MANIFEST_BYTES,
            "canonical_build_input_manifest_sha256": _MANIFEST_SHA256,
            "fixed_image_ids": {
                "python_contract": _PYTHON_IMAGE_ID,
                "backend_unit": _BACKEND_IMAGE_ID,
            },
            "frozen_subjects": {
                "backend/internal/calls/call_durable_credential_allocation.go": (
                    "sha256:" + "f" * 64
                )
            },
        }

    def fixed_images(self) -> dict[str, str]:
        self.events.append("images")
        return {
            "python_contract": _PYTHON_IMAGE_ID,
            "backend_unit": _BACKEND_IMAGE_ID,
        }

    def canonical_manifest(self) -> tuple[bytes, str]:
        self.events.append("manifest")
        return _MANIFEST_BYTES, _MANIFEST_SHA256

    def ready(self, _session: str) -> None:
        self.events.append("ready")

    def bind_candidate(self, _images: dict[str, str]) -> dict[str, Any]:
        self.events.append("bind")
        if self.interrupt_before_bind:
            raise KeyboardInterrupt
        return {
            "container_id": _CONTAINER_ID,
            "run_id": _RUN_ID,
            "owner_nonce": _OWNER_NONCE,
            "backend_image_id": _BACKEND_IMAGE_ID,
            "container_image": _BACKEND_IMAGE_ID,
        }

    def monitor(self, _candidate: dict[str, Any]) -> list[dict[str, int]]:
        self.events.append("monitor")
        return [_safe_sample(), _safe_sample()]

    def terminal_verify(self, _candidate: dict[str, Any]) -> None:
        self.events.append("terminal_verify")
        if self.unsealable:
            raise OSError("incomplete evidence")

    def publish(self, final: dict[str, Any], observer: dict[str, Any]) -> None:
        self.events.append("publish")
        self.writes.extend(("final-receipt.json", "observer-receipt.json", "SEAL"))
        self.published = (final, observer)

    def release_lease(self) -> None:
        self.events.append("lease.release")


def test_observer_timing_contract_is_literal_and_observer_owned() -> None:
    timing = resources.ObserverTiming()

    assert timing.candidate_window_seconds == 120
    assert timing.runtime_sample_interval_seconds == 5
    assert timing.ready_deadline_seconds == 30
    assert timing.destroy_grace_seconds == 30
    assert timing.whole_run_deadline_seconds == 6 * 60 * 60
    assert timing.deadline_owner == "observer"


@pytest.mark.parametrize("windows", (False, True))
@pytest.mark.parametrize("failure", ("hard_link", "nonregular", "parent_swap"))
def test_safe_paths_fail_closed_for_hard_links_nonregular_files_and_handle_races(
    windows: bool, failure: str
) -> None:
    ops = _AdversarialPathOps(
        windows=windows,
        hard_link=failure == "hard_link",
        nonregular=failure == "nonregular",
        parent_swap_after_open=failure == "parent_swap",
    )
    root = Path("C:/reports/run.partial") if windows else Path("/reports/run.partial")

    with pytest.raises(safe_paths.SafePathError):
        safe_paths.open_regular_relative(root, "nested/owner.json", ops=ops)

    assert ops.closes in ([], [17], [18])
    assert ops.writes == []
    if windows:
        assert ops.create_file_flags
        assert all(
            flags & safe_paths.FILE_FLAG_OPEN_REPARSE_POINT
            for flags in ops.create_file_flags
        )
    else:
        assert ops.openat_flags
        assert all(flags & safe_paths.O_NOFOLLOW for flags in ops.openat_flags)


def test_safe_paths_reject_intermediate_directory_swap_before_final_rename() -> None:
    ops = _AdversarialPathOps(rename_race=True)

    with pytest.raises(safe_paths.SafePathError):
        safe_paths.atomic_write_same_dir(
            Path("/reports/run.partial/final-receipt.json"), b"{}\n", ops=ops
        )

    assert ops.renames == [
        (
            "/reports/run.partial/final-receipt.json.tmp",
            "/reports/run.partial/final-receipt.json",
        )
    ]


@pytest.mark.parametrize(
    "changed_probe",
    (
        _SnapshotProbe(head="b" * 40),
        _SnapshotProbe(status=" M hostile.txt\n"),
        _SnapshotProbe(manifest=_MANIFEST_BYTES + b"changed\n"),
        _SnapshotProbe(subject_digest="sha256:" + "e" * 64),
        _SnapshotProbe(python_image_id="sha256:" + "e" * 64),
        _SnapshotProbe(backend_image_id="sha256:" + "e" * 64),
    ),
)
def test_root_prebuild_postgreen_and_observer_snapshot_chain_rejects_every_drift(
    changed_probe: _SnapshotProbe,
) -> None:
    before_probe = _SnapshotProbe()
    source_lease = lease.SourceLease.acquire(
        Path("/reports/.backend-unit-source-mutation.lease"), probe=before_probe
    )
    prebuild = source_lease.snapshot()
    postgreen = source_lease.snapshot()
    source_lease.probe = changed_probe
    observer = source_lease.snapshot()

    with pytest.raises(lease.SourceMutationError):
        source_lease.verify_snapshot_chain(prebuild, postgreen, observer)

    source_lease.release()
    assert before_probe.acl_changes == []
    assert changed_probe.acl_changes == []


def test_root_prebuild_postgreen_and_observer_snapshot_chain_requires_exact_equality() -> None:
    probe = _SnapshotProbe()
    source_lease = lease.SourceLease.acquire(
        Path("/reports/.backend-unit-source-mutation.lease"), probe=probe
    )
    prebuild = source_lease.snapshot()
    postgreen = source_lease.snapshot()
    observer = source_lease.snapshot()

    source_lease.verify_snapshot_chain(prebuild, postgreen, observer)
    source_lease.release()
    assert prebuild["head"] == postgreen["head"] == observer["head"]
    assert prebuild["status"] == postgreen["status"] == observer["status"]
    assert prebuild["canonical_build_input_manifest_bytes"] == _MANIFEST_BYTES
    assert prebuild["canonical_build_input_manifest_sha256"] == _MANIFEST_SHA256
    assert prebuild["fixed_image_ids"] == {
        "python_contract": _PYTHON_IMAGE_ID,
        "backend_unit": _BACKEND_IMAGE_ID,
    }
    assert prebuild["frozen_subjects"] == postgreen["frozen_subjects"] == observer[
        "frozen_subjects"
    ]


def test_publication_validates_hashes_then_writes_final_observer_seal_and_renames() -> None:
    final = {
        "run_id": _RUN_ID,
        "owner_nonce": _OWNER_NONCE,
        "classification": "SEALED_EXECUTABLE_PASS",
        "immutable_outcome": None,
        "container_id": _CONTAINER_ID,
        "bound_backend_image_id": _BACKEND_IMAGE_ID,
    }
    observer = {
        "run_id": _RUN_ID,
        "owner_nonce": _OWNER_NONCE,
        "container_id": _CONTAINER_ID,
        "fixed_image_ids": {
            "python_contract": _PYTHON_IMAGE_ID,
            "backend_unit": _BACKEND_IMAGE_ID,
        },
        "bound_backend_image_id": _BACKEND_IMAGE_ID,
        "container_image": _BACKEND_IMAGE_ID,
        "canonical_build_input_manifest_bytes": _MANIFEST_BYTES.decode(),
        "canonical_build_input_manifest_sha256": _MANIFEST_SHA256,
    }
    writer = _HashingWriter()
    validation_events: list[str] = []

    final_dir = publication.publish(
        Path("/reports/backend-unit-run.partial"),
        final,
        observer,
        validate=lambda receipt, observed: validation_events.append(
            f"{receipt['run_id']}:{observed['container_id']}"
        ),
        writer=writer,
    )

    assert validation_events == [f"{_RUN_ID}:{_CONTAINER_ID}"]
    assert writer.events == [
        ("write", "final-receipt.json"),
        ("write", "observer-receipt.json"),
        ("write", "SEAL"),
        ("rename", "backend-unit-run.partial->backend-unit-run"),
    ]
    assert final_dir == Path("/reports/backend-unit-run")
    assert writer.payloads["final-receipt.json"].endswith(b"\n")
    assert writer.payloads["observer-receipt.json"].endswith(b"\n")
    assert writer.payloads["SEAL"] == (
        b"hub.qa.backend-unit-seal/v1\n"
        + f"run_id={_RUN_ID}\n".encode()
        + f"owner_nonce={_OWNER_NONCE}\n".encode()
        + b"observer_receipt_sha256=sha256:"
        + hashlib.sha256(writer.payloads["observer-receipt.json"]).hexdigest().encode()
        + b"\n"
    )


@pytest.mark.parametrize("writer", (_HashingWriter(active_writer=True), _HashingWriter(final_exists=True)))
def test_publication_never_writes_before_terminal_prevalidation(writer: _HashingWriter) -> None:
    with pytest.raises(publication.PublicationError):
        publication.publish(
            Path("/reports/backend-unit-run.partial"),
            {"run_id": _RUN_ID, "owner_nonce": _OWNER_NONCE},
            {"run_id": _RUN_ID, "owner_nonce": _OWNER_NONCE},
            validate=lambda _final, _observer: None,
            writer=writer,
        )

    assert writer.events == []
    assert writer.payloads == {}


@pytest.mark.parametrize("classification", _SEALED_CLASSIFICATIONS)
def test_every_sealed_classification_exits_zero_only_after_publication(
    classification: str,
) -> None:
    runtime = _ObserverRuntime(classification=classification)

    assert observe.main(("--compose-project", _PROJECT), runtime=runtime, stdout=io.StringIO()) == 0
    assert runtime.published is not None
    final, observer = runtime.published
    assert final["classification"] == classification
    assert final["run_id"] == observer["run_id"] == _RUN_ID
    assert final["owner_nonce"] == observer["owner_nonce"] == _OWNER_NONCE
    assert final["container_id"] == observer["container_id"] == _CONTAINER_ID
    assert observer["fixed_image_ids"] == {
        "python_contract": _PYTHON_IMAGE_ID,
        "backend_unit": _BACKEND_IMAGE_ID,
    }
    assert observer["bound_backend_image_id"] == observer["container_image"] == _BACKEND_IMAGE_ID
    assert runtime.writes == ["final-receipt.json", "observer-receipt.json", "SEAL"]


@pytest.mark.parametrize(
    ("runtime", "argv", "expected_exit"),
    (
        (_ObserverRuntime(), ("--unexpected", _PROJECT), 64),
        (_ObserverRuntime(unsealable=True), ("--compose-project", _PROJECT), 75),
        (_ObserverRuntime(interrupt_before_bind=True), ("--compose-project", _PROJECT), 130),
    ),
)
def test_usage_unsealable_and_unbound_interrupt_never_publish(
    runtime: _ObserverRuntime, argv: tuple[str, str], expected_exit: int
) -> None:
    assert observe.main(argv, runtime=runtime, stdout=io.StringIO()) == expected_exit
    assert runtime.writes == []
    assert "publish" not in runtime.events
