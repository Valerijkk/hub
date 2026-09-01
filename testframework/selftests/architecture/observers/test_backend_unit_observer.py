from __future__ import annotations

import hashlib
import io
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
    changed_identity: bool = False
    existing_output: bool = False
    cross_directory_rename: bool = False
    windows: bool = False
    openat_flags: list[int] = field(default_factory=list)
    create_file_flags: list[int] = field(default_factory=list)
    writes: list[tuple[str, bytes]] = field(default_factory=list)
    renames: list[tuple[str, str]] = field(default_factory=list)
    closes: list[int] = field(default_factory=list)
    _fstat_calls: int = 0

    def openat(self, _directory_fd: int, _name: str, flags: int) -> int:
        self.openat_flags.append(flags)
        if self.fail_open:
            raise OSError("unsafe path")
        return 17

    def create_file_w(self, _path: str, flags: int) -> int:
        self.create_file_flags.append(flags)
        if self.fail_open:
            raise OSError("unsafe reparse point")
        return 18

    def fstat(self, _descriptor: int) -> dict[str, int | bool]:
        self._fstat_calls += 1
        changed = self.changed_identity and self._fstat_calls > 1
        return {
            "device": 7,
            "inode": 12 if changed else 11,
            "size": 1 if changed else 0,
            "regular": not self.nonregular,
            "nlink": 2 if self.hard_link else 1,
            "reparse": False,
        }

    def destination_exists(self, _path: Path) -> bool:
        return self.existing_output

    def write_new(self, path: Path, payload: bytes) -> Path:
        self.writes.append((path.as_posix(), payload))
        return Path(path.as_posix() + ".tmp")

    def rename_same_directory(self, source: Path, destination: Path) -> None:
        self.renames.append((source.as_posix(), destination.as_posix()))
        if self.cross_directory_rename:
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
class _ObserverRuntime:
    classification: str = "SEALED_EXECUTABLE_PASS"
    unsealable: bool = False
    interrupted_before_bind: bool = False
    python_version: tuple[int, int, int] = (3, 12, 3)
    stdlib_only: bool = True
    source_drift_after_lease: bool = False
    candidate: dict[str, Any] | None = None
    events: list[str] = field(default_factory=list)
    writes: list[str] = field(default_factory=list)
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
        if self.interrupted_before_bind:
            raise KeyboardInterrupt
        return _candidate() if self.candidate is None else self.candidate

    def monitor(self, _candidate: dict[str, Any]) -> list[dict[str, int]]:
        self.events.append("monitor")
        return [_safe_sample(), _safe_sample()]

    def terminal_verify(self, _candidate: dict[str, Any]) -> None:
        self.events.append("terminal_verify")
        if self.unsealable:
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
        (
            "docker",
            "inspect",
            "--type",
            "container",
            "--format",
            "{{json .}}",
            _CONTAINER_ID,
        ),
        ("docker", "ps", "--no-trunc", "--format", "{{json .}}"),
        (
            "docker",
            "stats",
            "--no-stream",
            "--format",
            "{{json .}}",
            _CONTAINER_ID,
        ),
        (
            "docker",
            "image",
            "inspect",
            "--format",
            "{{json .}}",
            "worktree-master-python-contract",
        ),
        (
            "docker",
            "image",
            "inspect",
            "--format",
            "{{json .}}",
            "worktree-master-backend-unit",
        ),
    )
    assert all(call.shell is False for call in runner.calls)
    assert all(call.timeout_seconds == 10 for call in runner.calls)
    assert all(call.output_cap_bytes == 1024 * 1024 for call in runner.calls)
    assert runner.calls[0].stream is True
    assert all(not call.stream for call in runner.calls[1:])
    forbidden = {"compose", "run", "create", "start", "build", "pull", "prune", "rm"}
    assert not any(
        token in forbidden for call in runner.calls for token in call.argv
    )


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

    for forbidden_project in (
        "WORKTREE-master",
        "worktree-master;docker",
        "../worktree-master",
    ):
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
def test_event_stream_caps_terminate_reap_and_refuse_to_seal(
    stream: _EventStream,
) -> None:
    runner = _DockerRunner(_docker_responses(), stream)
    with pytest.raises(docker_adapter.DockerReadError):
        docker_adapter.DockerReadAdapter(_PROJECT, runner=runner).events(_START)
    assert stream.terminated is True
    assert stream.reaped is True
    assert runner.calls == [
        _Invocation(_events_argv(), False, 10, 1024 * 1024, True)
    ]


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
        _Invocation(
            ("docker", "stop", "--timeout", "5", _CONTAINER_ID),
            False,
            10,
            1024 * 1024,
            False,
        )
    ]


@pytest.mark.parametrize(
    "changes",
    (
        {"container_id": _CONTAINER_ID[:-1]},
        {"started_at": "2026-09-01T11:57:59.999999999Z"},
        {"image_id": "sha256:" + "e" * 64},
        {"container_image": "sha256:" + "e" * 64},
        {"labels": {}},
        {
            "owner_marker": {
                "run_id": _RUN_ID,
                "owner_nonce": _OWNER_NONCE,
                "container_hostname": "bad-owner",
            }
        },
    ),
)
def test_candidate_binding_rejects_every_foreign_identity(
    changes: dict[str, object],
) -> None:
    with pytest.raises(observe.InvalidEvidence, match="INVALID_EVIDENCE"):
        observe.bind_unique_candidate(
            (_candidate(**changes),),
            _START,
            120,
            _PROJECT,
            "backend-unit",
            _BACKEND_IMAGE_ID,
        )


def test_candidate_binding_rejects_zero_or_two_valid_candidates_and_keeps_prefix() -> None:
    other_id = "f" * 64
    other = _candidate(
        container_id=other_id,
        owner_marker={
            "run_id": _RUN_ID,
            "owner_nonce": _OWNER_NONCE,
            "container_hostname": other_id[:12],
        },
    )
    for candidates in ((), (_candidate(), other)):
        with pytest.raises(observe.InvalidEvidence, match="INVALID_EVIDENCE"):
            observe.bind_unique_candidate(
                candidates, _START, 120, _PROJECT, "backend-unit", _BACKEND_IMAGE_ID
            )

    bound = observe.bind_unique_candidate(
        (_candidate(),), _START, 120, _PROJECT, "backend-unit", _BACKEND_IMAGE_ID
    )
    assert bound["container_id"] == _CONTAINER_ID
    assert bound["container_hostname"] == _CONTAINER_ID[: len(_HOSTNAME)]
    assert _CONTAINER_ID.startswith(bound["container_hostname"])
    assert bound["run_id"] == _RUN_ID
    assert bound["owner_nonce"] == _OWNER_NONCE
    assert bound["backend_image_id"] == _BACKEND_IMAGE_ID


@pytest.mark.parametrize(
    "relative",
    ("/absolute.json", "../parent.json", "nested/../../escape.json", "C:/escape.json"),
)
def test_safe_paths_reject_escape_before_open_or_write(relative: str) -> None:
    ops = _PathOps()
    with pytest.raises(safe_paths.SafePathError):
        safe_paths.open_regular_relative(
            Path("/reports/run.partial"), relative, ops=ops
        )
    assert ops.openat_flags == []
    assert ops.create_file_flags == []
    assert ops.writes == []


@pytest.mark.parametrize(
    "ops",
    (
        _PathOps(fail_open=True),
        _PathOps(hard_link=True),
        _PathOps(nonregular=True),
        _PathOps(changed_identity=True),
    ),
)
def test_safe_paths_fail_closed_for_links_types_and_identity_races(
    ops: _PathOps,
) -> None:
    with pytest.raises(safe_paths.SafePathError):
        safe_paths.open_regular_relative(
            Path("/reports/run.partial"), "owner.json", ops=ops
        )
    assert ops.writes == []
    assert ops.renames == []


def test_safe_paths_recheck_posix_windows_and_final_rename_boundary() -> None:
    posix = _PathOps()
    assert (
        safe_paths.open_regular_relative(
            Path("/reports/run.partial"), "owner.json", ops=posix
        )
        == 17
    )
    assert posix.openat_flags
    assert all(flags & safe_paths.O_NOFOLLOW for flags in posix.openat_flags)

    windows = _PathOps(windows=True)
    assert (
        safe_paths.open_regular_relative(
            Path("C:/reports/run.partial"), "owner.json", ops=windows
        )
        == 18
    )
    assert windows.create_file_flags
    assert all(
        flags & safe_paths.FILE_FLAG_OPEN_REPARSE_POINT
        for flags in windows.create_file_flags
    )

    for ops in (_PathOps(existing_output=True), _PathOps(cross_directory_rename=True)):
        with pytest.raises(safe_paths.SafePathError):
            safe_paths.atomic_write_same_dir(
                Path("/reports/run.partial/final-receipt.json"), b"{}\n", ops=ops
            )
        if ops.existing_output:
            assert ops.writes == []
        assert all(
            source.rsplit("/", 1)[0] == destination.rsplit("/", 1)[0]
            for source, destination in ops.renames
        )


def test_lease_snapshots_bind_exact_manifest_images_and_no_acl_claim() -> None:
    probe = _LeaseProbe()
    source_lease = lease.SourceLease.acquire(
        Path("/reports/.backend-unit-source-mutation.lease"), probe=probe
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
    assert prebuild["canonical_build_input_manifest_bytes"] == _MANIFEST_BYTES
    assert prebuild["canonical_build_input_manifest_sha256"] == _MANIFEST_SHA256
    assert prebuild["fixed_image_ids"] == {
        "python_contract": _PYTHON_IMAGE_ID,
        "backend_unit": _BACKEND_IMAGE_ID,
    }
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
def test_lease_rejects_every_prebuild_postgreen_or_observer_drift(
    changed_probe: _LeaseProbe,
) -> None:
    original = _LeaseProbe()
    source_lease = lease.SourceLease.acquire(
        Path("/reports/.backend-unit-source-mutation.lease"), probe=original
    )
    prebuild = source_lease.snapshot()
    postgreen = source_lease.snapshot()
    source_lease.probe = changed_probe
    observer_snapshot = source_lease.snapshot()
    with pytest.raises(lease.SourceMutationError):
        source_lease.verify_snapshot_chain(
            prebuild, postgreen, observer_snapshot
        )
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

    policy = {"sealed_classifications": _SEALED_CLASSIFICATIONS}
    cases = (
        ({"source_mutated": True, "invalid_provenance": True}, "SEALED_INVALID_SOURCE_MUTATED"),
        ({"invalid_provenance": True, "stop": "ram"}, "SEALED_INVALID_PROVENANCE"),
        ({"stop": "interrupt", "oom": True}, "SEALED_INTERRUPTED"),
        ({"stop": "qa", "oom": True}, "SEALED_INVALID_QA_OVERLAP"),
        ({"stop": "disk", "oom": True}, "SEALED_INFRA_ABORTED_DISK_PRESSURE"),
        ({"stop": "ram", "oom": True}, "SEALED_INFRA_ABORTED_HOST_RAM_FLOOR"),
        ({"stop": "deadline", "oom": True}, "SEALED_INFRA_ABORTED_WHOLE_RUN_TIMEOUT"),
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
    for facts, expected in cases:
        assert resources.classify(
            (_safe_sample(),), {**policy, "classification_facts": facts}
        ) == expected


def test_runtime_monitoring_retains_samples_and_stops_only_once_on_priority() -> None:
    one_of_three = (
        _safe_sample(),
        {**_safe_sample(), "PercentDiskTime": 70},
        _safe_sample(),
    )
    two_of_three = (
        _safe_sample(),
        {**_safe_sample(), "PercentDiskTime": 70},
        {**_safe_sample(), "CurrentDiskQueueLength": 2},
    )
    policy = {"classification_facts": {}, "sealed_classifications": _SEALED_CLASSIFICATIONS}
    assert resources.classification_record(one_of_three, policy)["predicates"][
        "rolling_disk_pressure_2_of_3"
    ] is False
    disk_record = resources.classification_record(two_of_three, policy)
    assert disk_record["predicates"]["rolling_disk_pressure_2_of_3"] is True
    assert disk_record["classification"] == "SEALED_INFRA_ABORTED_DISK_PRESSURE"

    samples = (
        {**_safe_sample(), "free_ram_bytes": 2 * 1024**3 - 1},
    ) * 2
    ram_record = resources.classification_record(samples, policy)
    assert ram_record["classification"] == "SEALED_INFRA_ABORTED_HOST_RAM_FLOOR"
    assert ram_record["raw_samples"] == samples
    assert ram_record["predicates"]["two_consecutive_ram_below_2_gib"] is True
    assert ram_record["immutable_outcome"] is None
    assert resources.classify(
        (_safe_sample(),),
        {**policy, "classification_facts": {"timeout": True}},
    ) == "SEALED_TIMEOUT_UNTRIAGED"

    runner = _DockerRunner(_docker_responses())
    adapter = docker_adapter.DockerReadAdapter(_PROJECT, runner=runner)
    adapter.bind_candidate(_CONTAINER_ID)
    assert adapter.stop_bound_once(_CONTAINER_ID) == {"result": "stopped"}
    with pytest.raises(docker_adapter.DockerReadError):
        adapter.stop_bound_once(_CONTAINER_ID)
    assert [call.argv for call in runner.calls] == [
        ("docker", "stop", "--timeout", "5", _CONTAINER_ID)
    ]


def test_publication_validates_before_write_then_hashes_seals_and_renames() -> None:
    final, observer = _publication_payloads()
    writer = _PublicationWriter()
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


@pytest.mark.parametrize("failure", ("semantic", "artifact-hash"))
def test_publication_failure_precedes_every_write(failure: str) -> None:
    final, observer = _publication_payloads()
    writer = _PublicationWriter()
    validation_events: list[str] = []

    def reject(_final: dict[str, Any], _observer: dict[str, Any]) -> None:
        validation_events.append(failure)
        raise publication.PublicationError(failure)

    with pytest.raises(publication.PublicationError, match=failure):
        publication.publish(
            Path("/reports/backend-unit-run.partial"),
            final,
            observer,
            validate=reject,
            writer=writer,
        )
    assert validation_events == [failure]
    assert writer.events == []
    assert writer.payloads == {}


@pytest.mark.parametrize(
    "writer",
    (
        _PublicationWriter(active_container_writer=True),
        _PublicationWriter(existing_final=True),
        _PublicationWriter(cross_directory=True),
    ),
)
def test_publication_refuses_active_existing_or_cross_directory_output(
    writer: _PublicationWriter,
) -> None:
    final, observer = _publication_payloads()
    with pytest.raises(publication.PublicationError):
        publication.publish(
            Path("/reports/backend-unit-run.partial"),
            final,
            observer,
            validate=lambda _final, _observer: None,
            writer=writer,
        )
    if writer.active_container_writer or writer.existing_final:
        assert writer.events == []


def test_observer_lifecycle_bootstrap_lease_ready_publish_and_identity_order() -> None:
    runtime = _ObserverRuntime()
    stdout = io.StringIO()
    assert observe.main(
        ("--compose-project", _PROJECT), runtime=runtime, stdout=stdout
    ) == 0
    session = stdout.getvalue().removeprefix("READY observer_session=").strip()
    assert stdout.getvalue() == f"READY observer_session={session}\n"
    assert len(session) == 32
    assert set(session) <= set("0123456789abcdef")
    assert runtime.events == [
        "bootstrap",
        "preflight",
        "lease.acquire",
        "snapshot.before",
        "images",
        "manifest",
        "ready",
        "bind",
        "monitor",
        "terminal_verify",
        "snapshot.after",
        "publish",
        "lease.release",
    ]
    assert runtime.writes == [
        "final-receipt.json",
        "observer-receipt.json",
        "SEAL",
    ]
    assert runtime.published is not None
    final, observer = runtime.published
    assert final["run_id"] == observer["run_id"] == _RUN_ID
    assert final["owner_nonce"] == observer["owner_nonce"] == _OWNER_NONCE
    assert final["container_id"] == observer["container_id"] == _CONTAINER_ID
    assert observer["canonical_build_input_manifest_bytes"] == _MANIFEST_BYTES.decode()
    assert observer["canonical_build_input_manifest_sha256"] == _MANIFEST_SHA256


@pytest.mark.parametrize(
    ("facts", "classification"),
    (
        ({}, "SEALED_EXECUTABLE_PASS"),
        ({"discovery_failure": True}, "SEALED_EXECUTABLE_FAIL_DISCOVERY"),
        ({"stage_failure": True}, "SEALED_EXECUTABLE_FAIL_STAGE"),
        ({"timeout": True}, "SEALED_TIMEOUT_UNTRIAGED"),
        ({"oom": True}, "SEALED_INFRA_ABORTED_OOM"),
        ({"resource": True}, "SEALED_INFRA_ABORTED_RESOURCE_CONTRACT"),
        ({"budget": True}, "SEALED_INFRA_ABORTED_REPORT_BUDGET"),
        ({"tmpfs": True}, "SEALED_INFRA_ABORTED_TMPFS"),
        ({"stop": "ram"}, "SEALED_INFRA_ABORTED_HOST_RAM_FLOOR"),
        ({"stop": "disk"}, "SEALED_INFRA_ABORTED_DISK_PRESSURE"),
        ({"stop": "deadline"}, "SEALED_INFRA_ABORTED_WHOLE_RUN_TIMEOUT"),
        ({"stop": "interrupt"}, "SEALED_INTERRUPTED"),
        ({"stop": "qa"}, "SEALED_INVALID_QA_OVERLAP"),
        ({"source_mutated": True}, "SEALED_INVALID_SOURCE_MUTATED"),
        ({"invalid_provenance": True}, "SEALED_INVALID_PROVENANCE"),
    ),
)
def test_every_sealed_classification_publishes_only_from_matching_facts_and_exits_zero(
    facts: dict[str, bool | str], classification: str
) -> None:
    def run_observer(
        expected_classification: str, classification_facts: dict[str, bool | str]
    ) -> tuple[int, _PublicationWriter]:
        writer = _PublicationWriter()
        observed = resources.classify(
            (_safe_sample(),),
            {
                "classification_facts": classification_facts,
                "sealed_classifications": _SEALED_CLASSIFICATIONS,
            },
        )
        if observed != expected_classification:
            return 75, writer

        final, observer = _publication_payloads()
        final["classification"] = observed
        if observed != "SEALED_TIMEOUT_UNTRIAGED":
            del final["immutable_outcome"]

        def validate(receipt: dict[str, Any], observed_receipt: dict[str, Any]) -> None:
            if resources.classify(
                (_safe_sample(),),
                {
                    "classification_facts": classification_facts,
                    "sealed_classifications": _SEALED_CLASSIFICATIONS,
                },
            ) != receipt["classification"]:
                raise publication.PublicationError("classification facts mismatch")
            if (
                receipt["run_id"] != observed_receipt["run_id"]
                or receipt["owner_nonce"] != observed_receipt["owner_nonce"]
                or receipt["container_id"] != observed_receipt["container_id"]
                or receipt["bound_backend_image_id"]
                != observed_receipt["bound_backend_image_id"]
            ):
                raise publication.PublicationError("invalid publication facts")
            if receipt["classification"] == "SEALED_TIMEOUT_UNTRIAGED":
                if receipt.get("immutable_outcome") != "TIMEOUT_UNTRIAGED":
                    raise publication.PublicationError("invalid timeout outcome")
            elif "immutable_outcome" in receipt:
                raise publication.PublicationError("unexpected immutable outcome")

        publication.publish(
            Path("/reports/backend-unit-run.partial"),
            final,
            observer,
            validate=validate,
            writer=writer,
        )
        return 0, writer

    assert classification in _SEALED_CLASSIFICATIONS
    exit_code, writer = run_observer(classification, facts)
    assert exit_code == 0
    assert writer.events == [
        ("write", "final-receipt.json"),
        ("write", "observer-receipt.json"),
        ("write", "SEAL"),
        ("rename", "backend-unit-run.partial->backend-unit-run"),
    ]

    mismatched = next(
        candidate for candidate in _SEALED_CLASSIFICATIONS if candidate != classification
    )
    rejected_exit, rejected_writer = run_observer(mismatched, facts)
    assert rejected_exit == 75
    assert rejected_writer.events == []
    assert rejected_writer.payloads == {}


@pytest.mark.parametrize(
    ("runtime", "argv", "expected_exit"),
    (
        (_ObserverRuntime(), ("--wrong", _PROJECT), 64),
        (_ObserverRuntime(unsealable=True), ("--compose-project", _PROJECT), 75),
        (
            _ObserverRuntime(interrupted_before_bind=True),
            ("--compose-project", _PROJECT),
            130,
        ),
        (
            _ObserverRuntime(source_drift_after_lease=True),
            ("--compose-project", _PROJECT),
            75,
        ),
    ),
)
def test_observer_exit_contract_never_seals_unbound_or_invalid_evidence(
    runtime: _ObserverRuntime, argv: tuple[str, str], expected_exit: int
) -> None:
    assert observe.main(argv, runtime=runtime, stdout=io.StringIO()) == expected_exit
    assert runtime.writes == []
    assert "publish" not in runtime.events
    assert runtime.events.count("lease.release") <= 1


def test_observer_rejects_non312_or_nonstdlib_bootstrap_before_ready_or_write() -> None:
    for runtime in (
        _ObserverRuntime(python_version=(3, 11, 9)),
        _ObserverRuntime(stdlib_only=False),
    ):
        with pytest.raises(observe.BootstrapError):
            observe.main(
                ("--compose-project", _PROJECT), runtime=runtime, stdout=io.StringIO()
            )
        assert runtime.events == ["bootstrap"]
        assert runtime.writes == []
