from __future__ import annotations

import hashlib
import io
import itertools
import json
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

import pytest

from observers.backend_unit import (
    docker_adapter,
    lease,
    observe,
    publication,
    resources,
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
_MANIFEST_BYTES = b"".join(
    path.encode() + b"\tsha256:" + hashlib.sha256(path.encode()).hexdigest().encode() + b"\n"
    for path in _MANIFEST_PATHS
)
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


@dataclass(frozen=True)
class _Invocation:
    argv: tuple[str, ...]
    shell: bool
    timeout_seconds: int
    output_cap_bytes: int
    stream: bool


class _EventStream:
    def __init__(self, lines: tuple[bytes, ...]) -> None:
        self._lines = lines
        self.terminated = False
        self.reaped = False

    def __iter__(self) -> Any:
        return iter(self._lines)

    def terminate(self) -> None:
        self.terminated = True

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


@dataclass
class _PathOps:
    fail_open: bool = False
    hard_link: bool = False
    nonregular: bool = False
    reparse: bool = False
    size_changed: bool = False
    identity_changed: bool = False
    existing_output: bool = False
    cross_directory_rename: bool = False
    windows: bool = False
    swap_intermediate_name: bool = False
    openat_calls: list[tuple[int, str, int]] = field(default_factory=list)
    create_file_flags: list[int] = field(default_factory=list)
    writes: list[tuple[str, bytes]] = field(default_factory=list)
    renameat_calls: list[tuple[int, str, int, str]] = field(default_factory=list)
    closes: list[int] = field(default_factory=list)
    _fstat_calls: int = 0
    fstat_descriptors: list[int] = field(default_factory=list)

    def openat(self, directory_fd: int, name: str, flags: int) -> int:
        self.openat_calls.append((directory_fd, name, flags))
        if self.fail_open:
            raise OSError("unsafe path")
        if self.swap_intermediate_name and len(self.openat_calls) == 1:
            self.swap_intermediate_name = False
        return 17 + len(self.openat_calls)

    def create_file_w(self, _path: str, flags: int) -> int:
        self.create_file_flags.append(flags)
        if self.fail_open:
            raise OSError("unsafe reparse point")
        return 18

    def fstat(self, descriptor: int) -> dict[str, int | bool]:
        self._fstat_calls += 1
        self.fstat_descriptors.append(descriptor)
        return {
            "device": 8 if self.identity_changed and self._fstat_calls > 1 else 7,
            "inode": 11,
            "size": 1 if self.size_changed and self._fstat_calls > 1 else 0,
            "regular": not self.nonregular,
            "nlink": 2 if self.hard_link else 1,
            "reparse": self.reparse,
        }

    def destination_exists(self, _path: Path) -> bool:
        return self.existing_output

    def write_new(self, path: Path, payload: bytes) -> Path:
        self.writes.append((path.as_posix(), payload))
        return Path(path.name + ".tmp")

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
            return " M hostile.txt\n" if self.dirty else ""
        raise AssertionError(f"unexpected Git argv: {argv!r}")

    def canonical_build_input_manifest(self) -> bytes:
        return self.manifest_bytes

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
class _PublicationWriter:
    events: list[tuple[str, str]] = field(default_factory=list)
    payloads: dict[str, bytes] = field(default_factory=dict)
    existing_final: bool = False
    active_container_writer: bool = False
    cross_directory: bool = False

    def assert_terminal_and_stable(self, _run_dir: Path) -> None:
        if self.active_container_writer:
            raise OSError("container writer is still active")

    def assert_absent(self, _final_dir: Path) -> None:
        if self.existing_final:
            raise FileExistsError("final directory already exists")

    def write(self, path: Path, payload: bytes) -> None:
        self.payloads[path.name] = payload
        self.events.append(("write", path.name))

    def rename_same_directory(self, partial: Path, final: Path) -> None:
        self.events.append(("rename", f"{partial.name}->{final.name}"))
        if self.cross_directory:
            raise OSError("cross-directory publication")


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


def _publication_payloads() -> tuple[dict[str, Any], dict[str, Any]]:
    return (
        {
            "run_id": _RUN_ID,
            "owner_nonce": _OWNER_NONCE,
            "classification": "SEALED_TIMEOUT_UNTRIAGED",
            "immutable_outcome": "TIMEOUT_UNTRIAGED",
            "container_id": _CONTAINER_ID,
            "bound_backend_image_id": _BACKEND_IMAGE_ID,
        },
        {
            "run_id": _RUN_ID,
            "owner_nonce": _OWNER_NONCE,
            "container_id": _CONTAINER_ID,
            "bound_backend_image_id": _BACKEND_IMAGE_ID,
            "container_image": _BACKEND_IMAGE_ID,
            "canonical_build_input_manifest_bytes": _MANIFEST_BYTES.decode(),
            "canonical_build_input_manifest_sha256": _MANIFEST_SHA256,
            "fixed_image_ids": {
                "python_contract": _PYTHON_IMAGE_ID,
                "backend_unit": _BACKEND_IMAGE_ID,
            },
        },
    )


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
    assert hashlib.sha256(_MANIFEST_BYTES).hexdigest() == _MANIFEST_SHA256.removeprefix(
        "sha256:"
    )


def test_docker_adapter_has_only_literal_read_argvs_and_fixed_images() -> None:
    event = json.dumps(
        {
            "id": _CONTAINER_ID,
            "status": "create",
            "timeNano": 1,
            "Actor": {
                "Attributes": {
                    "com.docker.compose.project": _PROJECT,
                    "com.docker.compose.service": "backend-unit",
                }
            },
        }
    ).encode() + b"\n"
    runner = _DockerRunner(_docker_responses(), _EventStream((event,)))
    adapter = docker_adapter.DockerReadAdapter(_PROJECT, runner=runner)

    assert adapter.events(_START) == (json.loads(event),)
    assert adapter.inspect(_CONTAINER_ID)["Id"] == _CONTAINER_ID
    assert adapter.ps() == ()
    assert adapter.stats(_CONTAINER_ID) == {}
    assert adapter.inspect_fixed_images() == {
        "python_contract": _PYTHON_IMAGE_ID,
        "backend_unit": _BACKEND_IMAGE_ID,
    }
    assert tuple(call.argv for call in runner.calls) == (
        _events_argv(),
        ("docker", "inspect", "--type", "container", "--format", "{{json .}}", _CONTAINER_ID),
        ("docker", "ps", "--no-trunc", "--format", "{{json .}}"),
        ("docker", "stats", "--no-stream", "--format", "{{json .}}", _CONTAINER_ID),
        ("docker", "image", "inspect", "--format", "{{json .}}", "worktree-master-python-contract"),
        ("docker", "image", "inspect", "--format", "{{json .}}", "worktree-master-backend-unit"),
    )
    assert all(call.shell is False for call in runner.calls)
    assert all(call.timeout_seconds == 10 for call in runner.calls)
    assert all(call.output_cap_bytes == 1024 * 1024 for call in runner.calls)
    assert runner.calls[0].stream is True
    assert all(not call.stream for call in runner.calls[1:])
    forbidden = {"compose", "run", "create", "start", "build", "pull", "prune", "rm"}
    assert not any(token in forbidden for call in runner.calls for token in call.argv)


def test_docker_adapter_rejects_invalid_since_dynamic_images_and_extra_argv() -> None:
    for invalid_since in ("2026-09-01T12:00:00Z", "not-a-timestamp", 1, None):
        runner = _DockerRunner(_docker_responses())
        with pytest.raises(docker_adapter.DockerReadError):
            docker_adapter.DockerReadAdapter(_PROJECT, runner=runner).events(  # type: ignore[arg-type]
                invalid_since
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


@pytest.mark.parametrize(
    "stream",
    (
        _EventStream(tuple(b"{}\n" for _ in range(17))),
        _EventStream((b"x" * (64 * 1024 + 1),)),
    ),
)
def test_event_stream_caps_terminate_reap_and_refuse_to_seal(stream: _EventStream) -> None:
    runner = _DockerRunner(_docker_responses(), stream)
    with pytest.raises(docker_adapter.DockerReadError):
        docker_adapter.DockerReadAdapter(_PROJECT, runner=runner).events(_START)
    assert stream.terminated is True
    assert stream.reaped is True
    assert runner.calls == [_Invocation(_events_argv(), False, 10, 1024 * 1024, True)]


def test_bound_ids_and_stop_are_full_canonical_single_use_authority() -> None:
    for invalid_id in ("", _CONTAINER_ID[:-1], _CONTAINER_ID.upper(), "f" * 63 + "g"):
        runner = _DockerRunner(_docker_responses())
        adapter = docker_adapter.DockerReadAdapter(_PROJECT, runner=runner)
        for method in (adapter.inspect, adapter.stats, adapter.stop_bound_once):
            with pytest.raises(docker_adapter.DockerReadError):
                method(invalid_id)
        assert runner.calls == []

    runner = _DockerRunner(_docker_responses())
    adapter = docker_adapter.DockerReadAdapter(_PROJECT, runner=runner)
    adapter.bind_candidate(_CONTAINER_ID)
    assert adapter.stop_bound_once(_CONTAINER_ID) == {"result": "stopped"}
    with pytest.raises(docker_adapter.DockerReadError):
        adapter.stop_bound_once(_CONTAINER_ID)
    assert runner.calls == [
        _Invocation(("docker", "stop", "--timeout", "5", _CONTAINER_ID), False, 10, 1024 * 1024, False)
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
    assert ops.openat_calls == []
    assert ops.create_file_flags == []
    assert ops.writes == []


@pytest.mark.parametrize(
    "ops",
    (
        _PathOps(fail_open=True),
        _PathOps(hard_link=True),
        _PathOps(nonregular=True),
        _PathOps(size_changed=True),
    ),
)
def test_safe_paths_fail_closed_for_links_types_and_size_only_toctou(
    ops: _PathOps,
) -> None:
    with pytest.raises(safe_paths.SafePathError):
        safe_paths.open_regular_relative(Path("/reports/run.partial"), "owner.json", ops=ops)
    assert ops.writes == []
    assert ops.renameat_calls == []
    assert ops.closes


def test_safe_paths_rejects_post_open_identity_toctou_on_same_opened_descriptor_without_publication() -> None:
    ops = _PathOps(identity_changed=True)
    with pytest.raises(safe_paths.SafePathError):
        safe_paths.open_regular_relative(Path("/reports/run.partial"), "owner.json", ops=ops)
    assert ops.fstat_descriptors == [18, 18]
    assert ops.closes == [18]
    assert ops.writes == []
    assert ops.renameat_calls == []


def test_safe_paths_windows_reparse_closes_handle_without_write() -> None:
    ops = _PathOps(windows=True, reparse=True)
    with pytest.raises(safe_paths.SafePathError):
        safe_paths.open_regular_relative(Path("C:/reports/run.partial"), "owner.json", ops=ops)
    assert ops.create_file_flags
    assert all(flags & safe_paths.FILE_FLAG_OPEN_REPARSE_POINT for flags in ops.create_file_flags)
    assert ops.closes == [18]
    assert ops.writes == []


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


def test_safe_paths_atomic_write_uses_renameat_with_one_parent_fd() -> None:
    ops = _PathOps()
    safe_paths.atomic_write_same_dir(
        Path("/reports/run.partial/final-receipt.json"), b"{}\n", ops=ops
    )
    assert ops.writes == [("/reports/run.partial/final-receipt.json", b"{}\n")]
    assert len(ops.renameat_calls) == 1
    source_parent_fd, source, destination_parent_fd, destination = ops.renameat_calls[0]
    assert source_parent_fd == destination_parent_fd
    assert source.endswith(".tmp")
    assert destination == "final-receipt.json"

    for failed_ops in (_PathOps(existing_output=True), _PathOps(cross_directory_rename=True)):
        with pytest.raises(safe_paths.SafePathError):
            safe_paths.atomic_write_same_dir(
                Path("/reports/run.partial/final-receipt.json"), b"{}\n", ops=failed_ops
            )
        if failed_ops.existing_output:
            assert failed_ops.writes == []


def test_lease_snapshots_bind_exact_manifest_images_and_no_acl_claim() -> None:
    probe = _LeaseProbe()
    source_lease = lease.SourceLease.acquire(Path("/reports/.backend-unit-source-mutation.lease"), probe=probe)
    prebuild = source_lease.snapshot()
    postgreen = source_lease.snapshot()
    observer_snapshot = source_lease.snapshot()
    source_lease.verify_snapshot_chain(prebuild, postgreen, observer_snapshot)
    source_lease.release()

    assert probe.commands == [
        ("git", "rev-parse", "HEAD"),
        ("git", "status", "--porcelain=v1", "--untracked-files=all"),
    ] * 3
    assert prebuild["canonical_build_input_manifest_bytes"] == _MANIFEST_BYTES
    assert prebuild["canonical_build_input_manifest_sha256"] == _MANIFEST_SHA256
    assert prebuild["fixed_image_ids"] == {"python_contract": _PYTHON_IMAGE_ID, "backend_unit": _BACKEND_IMAGE_ID}
    assert prebuild["frozen_subjects"] == probe.frozen_subject_digests()
    assert probe.acl_changes == []


@pytest.mark.parametrize(
    "changed_probe",
    (
        _LeaseProbe(head="c" * 40),
        _LeaseProbe(dirty=True),
        _LeaseProbe(manifest_bytes=_MANIFEST_BYTES + b"changed\n"),
        _LeaseProbe(python_image_id="sha256:" + "e" * 64),
        _LeaseProbe(backend_image_id="sha256:" + "e" * 64),
        _LeaseProbe(subject_digest="sha256:" + "e" * 64),
    ),
)
def test_lease_rejects_every_prebuild_postgreen_or_observer_drift(changed_probe: _LeaseProbe) -> None:
    original = _LeaseProbe()
    source_lease = lease.SourceLease.acquire(Path("/reports/.backend-unit-source-mutation.lease"), probe=original)
    prebuild = source_lease.snapshot()
    postgreen = source_lease.snapshot()
    source_lease.probe = changed_probe
    observer_snapshot = source_lease.snapshot()
    with pytest.raises(lease.SourceMutationError):
        source_lease.verify_snapshot_chain(prebuild, postgreen, observer_snapshot)
    source_lease.release()
    assert original.acl_changes == []
    assert changed_probe.acl_changes == []


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
    assert tuple(expected for _, expected in cases) == _SEALED_CLASSIFICATIONS
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

    runner = _DockerRunner(_docker_responses())
    adapter = docker_adapter.DockerReadAdapter(_PROJECT, runner=runner)
    adapter.bind_candidate(_CONTAINER_ID)
    assert adapter.stop_bound_once(_CONTAINER_ID) == {"result": "stopped"}
    with pytest.raises(docker_adapter.DockerReadError):
        adapter.stop_bound_once(_CONTAINER_ID)
    assert [call.argv for call in runner.calls] == [("docker", "stop", "--timeout", "5", _CONTAINER_ID)]


def test_publication_validates_before_write_then_hashes_seals_and_renames() -> None:
    final, observer = _publication_payloads()
    writer = _PublicationWriter()
    validation_events: list[str] = []
    final_dir = publication.publish(
        Path("/reports/backend-unit-run.partial"), final, observer,
        validate=lambda receipt, observed: validation_events.append(f"{receipt['run_id']}:{observed['container_id']}"),
        writer=writer,
    )

    assert validation_events == [f"{_RUN_ID}:{_CONTAINER_ID}"]
    assert writer.events == [
        ("write", "final-receipt.json"), ("write", "observer-receipt.json"),
        ("write", "SEAL"), ("rename", "backend-unit-run.partial->backend-unit-run"),
    ]
    assert final_dir == Path("/reports/backend-unit-run")
    assert writer.payloads["final-receipt.json"].endswith(b"\n")
    assert writer.payloads["observer-receipt.json"].endswith(b"\n")
    assert writer.payloads["SEAL"] == (
        b"hub.qa.backend-unit-seal/v1\n" + f"run_id={_RUN_ID}\n".encode()
        + f"owner_nonce={_OWNER_NONCE}\n".encode() + b"observer_receipt_sha256=sha256:"
        + hashlib.sha256(writer.payloads["observer-receipt.json"]).hexdigest().encode() + b"\n"
    )


@pytest.mark.parametrize("failure", ("semantic", "artifact-hash"))
def test_publication_failure_precedes_every_write(failure: str) -> None:
    final, observer = _publication_payloads()
    writer = _PublicationWriter()
    validation_events: list[str] = []

    def reject(_final: dict[str, Any], _observer: dict[str, Any]) -> None:
        validation_events.append(failure)
        raise publication.PublicationError(failure)

    with pytest.raises(publication.PublicationError, match=failure):
        publication.publish(Path("/reports/backend-unit-run.partial"), final, observer, validate=reject, writer=writer)
    assert validation_events == [failure]
    assert writer.events == []
    assert writer.payloads == {}


@pytest.mark.parametrize("writer", (_PublicationWriter(active_container_writer=True), _PublicationWriter(existing_final=True), _PublicationWriter(cross_directory=True)))
def test_publication_refuses_active_existing_or_cross_directory_output(writer: _PublicationWriter) -> None:
    final, observer = _publication_payloads()
    with pytest.raises(publication.PublicationError):
        publication.publish(
            Path("/reports/backend-unit-run.partial"), final, observer,
            validate=lambda _final, _observer: None, writer=writer,
        )
    if writer.active_container_writer or writer.existing_final:
        assert writer.events == []


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
        final["bound_backend_image_id"]
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
