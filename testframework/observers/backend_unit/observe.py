"""Fail-closed host observer for the one frozen ``backend-unit`` run.

The observer is deliberately not a launcher.  It acquires an immutable source
lease, attaches to one already-started local Docker container, and publishes a
receipt only after every host, source, image, and artifact binding remains
authoritative.
"""

from __future__ import annotations

import hashlib
import json
import os
import re
import secrets
import sys
import time
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Mapping, Sequence


# The required production invocation is a script below ``testframework``.  It
# has no third-party dependency: put that one package root on sys.path only for
# the direct-script form, then use the same package imports as an imported test.
if __package__ in (None, ""):
    _TESTFRAMEWORK_ROOT = Path(__file__).resolve().parents[2]
    if str(_TESTFRAMEWORK_ROOT) not in sys.path:
        sys.path.insert(0, str(_TESTFRAMEWORK_ROOT))
    from observers.backend_unit import (  # type: ignore[import-not-found]
        contracts,
        publication,
        resources,
        root_handoff,
        safe_paths,
    )
    from observers.backend_unit.docker_adapter import (  # type: ignore[import-not-found]
        DockerReadAdapter,
        DockerReadError,
        InvalidEvidence,
    )
    from observers.backend_unit.lease import (  # type: ignore[import-not-found]
        SourceLease,
        SourceLeaseError,
        SourceMutationError,
    )
else:
    from . import contracts, publication, resources, root_handoff, safe_paths
    from .docker_adapter import DockerReadAdapter, DockerReadError, InvalidEvidence
    from .lease import SourceLease, SourceLeaseError, SourceMutationError


__all__ = ("BootstrapError", "InvalidEvidence", "bind_unique_candidate", "main")


class BootstrapError(RuntimeError):
    """The exact observer command cannot establish its fixed host authority."""


_PROJECT = "worktree-master"
_SERVICE = "backend-unit"
_SOURCE_LEASE = ".backend-unit-source-mutation.lease"
_COMPLETE_MARKER = "CONTAINER_COMPLETE"
_ABORTED_MARKER = "CONTAINER_ABORTED"
_OWNER = "owner.json"
_CONTAINER_RECEIPT = "container-receipt.json"
_MAX_OWNER_BYTES = 1024
_MAX_RECEIPT_BYTES = 64 * 1024 * 1024
_CONCURRENCY_SCOPE = "docker-ps-visible-containers-only"
_ENGINE_ID_RE = re.compile(r"[0-9a-f]{64}\Z")
_HOSTNAME_RE = re.compile(r"[0-9a-f]{12,64}\Z")
_NONCE_RE = re.compile(r"[0-9a-f]{32}\Z")
_RUN_ID_RE = re.compile(r"backend-unit-[0-9]{8}T[0-9]{6}Z-[0-9a-f]{12,64}-[0-9a-f]{8}\Z")
_SHA256_RE = re.compile(r"sha256:[0-9a-f]{64}\Z")
_RFC3339_NANO_RE = re.compile(r"[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{9}Z\Z")
_FACT_KEYS = (
    "source_mutated",
    "invalid_provenance",
    "oom",
    "resource",
    "budget",
    "tmpfs",
    "timeout",
    "discovery_failure",
    "stage_failure",
)
_POLICY_FACT_KEYS = frozenset(
    (*_FACT_KEYS, "explicit_stops", "prohibited_backend_unit_overlap",
     "prohibited_mail_calendar_qa_test_build_overlap", "buildkit_idle",
     "whole_run_deadline_expired")
)
_STOP_PREDICATES = {
    "interrupt": "interrupt",
    "qa": "qa_overlap",
    "disk": "disk_pressure",
    "ram": "host_ram_floor",
    "deadline": "whole_run_timeout",
}
_PS_FIELDS = frozenset({"Command", "CreatedAt", "ID", "Image", "Labels", "LocalVolumes", "Mounts", "Names", "Networks", "Ports", "RunningFor", "Size", "State", "Status"})
_NON_QA_COMPOSE_SERVICES = frozenset({"postgres", "mongo", "redis", "nats", "seaweedfs", "livekit", "meilisearch", "hub-mail", "backend", "frontend", "qa-lock", "qa-postgres", "qa-mongo", "qa-redis", "qa-nats", "qa-seaweedfs"})
_QA_COMPOSE_SERVICES = frozenset({"backend-unit", "frontend-unit", "python-contract", "python-api", "python-integration", "python-e2e-smoke", "python-e2e-full", "python-load", "mail-calendar-qa-test-build"})
_BUILDKIT_IMAGE = "moby/buildkit:buildx-stable-1"
_CONTAINER_ARTIFACTS = (
    "CONTAINER_COMPLETE", "build.stderr.log", "build.stdout.log", "owner.json",
    "package-plan.build.json", "package-plan.race.json", "package-plan.unit.json", "package-plan.vet.json",
    "package-results.build.jsonl", "package-results.race.jsonl", "package-results.unit.jsonl", "package-results.vet.jsonl",
    "race.stderr.log", "race.stdout.log", "stage-build.json", "stage-race.json", "stage-unit.json", "stage-vet.json",
    "unit.stderr.log", "unit.stdout.log", "vet.stderr.log", "vet.stdout.log",
)
_HOST_MOUNTS = (
    ("bind", "./backend/cmd", "/opt/hub/source/backend/cmd", True, None),
    ("bind", "./backend/internal", "/opt/hub/source/backend/internal", True, None),
    ("bind", "./backend/tests", "/opt/hub/source/backend/tests", True, None),
    ("bind", "./backend/ent/schema", "/opt/hub/source/backend/ent/schema", True, None),
    ("bind", "./backend/migrations", "/opt/hub/source/backend/migrations", True, None),
    ("bind", "./backend/testframework", "/opt/hub/source/backend/testframework", True, None),
    ("bind", "./testframework/reports", "/reports", False, None),
    ("tmpfs", None, "/tmp", False, 67108864),
    ("tmpfs", None, "/var/cache/hub-go", False, 805306368),
    ("tmpfs", None, "/var/tmp/hub-go", False, 402653184),
)
_HOST_CONFIG = {
    "memory": 2147483648,
    "memory_swap": 2147483648,
    "nano_cpus": 6000000000,
    "read_only_rootfs": True,
    "security_opt": ["no-new-privileges:true"],
    "cap_drop": ["ALL"],
    "pids_limit": 256,
    "init": True,
    "network_mode": "worktree-master_qa-internal",
}


def _sha256(payload: bytes) -> str:
    return "sha256:" + hashlib.sha256(payload).hexdigest()


def _utc_timestamp() -> str:
    now = datetime.now(timezone.utc)
    return now.strftime("%Y-%m-%dT%H:%M:%S") + f".{now.microsecond:06d}Z"


def _rfc3339_nanos(value: object) -> int:
    if type(value) is not str or _RFC3339_NANO_RE.fullmatch(value) is None:
        raise InvalidEvidence("INVALID_EVIDENCE: timestamp is not canonical RFC3339Nano")
    try:
        seconds = int(datetime.strptime(value[:19], "%Y-%m-%dT%H:%M:%S").replace(tzinfo=timezone.utc).timestamp())
    except ValueError as error:
        raise InvalidEvidence("INVALID_EVIDENCE: timestamp has an invalid calendar value") from error
    return seconds * 1_000_000_000 + int(value[20:29])


def _strict_object(pairs: list[tuple[str, Any]]) -> dict[str, Any]:
    result: dict[str, Any] = {}
    for key, value in pairs:
        if type(key) is not str or key in result:
            raise InvalidEvidence("INVALID_EVIDENCE: JSON contains duplicate keys")
        result[key] = value
    return result


def _reject_constant(_value: str) -> None:
    raise InvalidEvidence("INVALID_EVIDENCE: JSON contains a non-finite value")


def _strict_json(payload: bytes, label: str) -> dict[str, Any]:
    if type(payload) is not bytes:
        raise InvalidEvidence(f"INVALID_EVIDENCE: {label} is not bytes")
    try:
        value = json.loads(
            payload.decode("utf-8", "strict"),
            object_pairs_hook=_strict_object,
            parse_constant=_reject_constant,
        )
    except InvalidEvidence:
        raise
    except (UnicodeError, TypeError, ValueError) as error:
        raise InvalidEvidence(f"INVALID_EVIDENCE: {label} is malformed JSON") from error
    if type(value) is not dict:
        raise InvalidEvidence(f"INVALID_EVIDENCE: {label} is not an object")
    return value


def _canonical_json(payload: Mapping[str, object]) -> bytes:
    try:
        return json.dumps(
            payload,
            ensure_ascii=False,
            allow_nan=False,
            separators=(",", ":"),
            sort_keys=True,
        ).encode("utf-8", "strict") + b"\n"
    except (TypeError, ValueError, UnicodeError) as error:
        raise InvalidEvidence("INVALID_EVIDENCE: receipt cannot be canonicalized") from error


def _labels(value: object) -> dict[str, str]:
    if isinstance(value, Mapping):
        if not all(type(key) is str and type(item) is str for key, item in value.items()):
            raise InvalidEvidence("INVALID_EVIDENCE: Docker labels are malformed")
        return dict(value)
    if type(value) is not str:
        raise InvalidEvidence("INVALID_EVIDENCE: Docker labels are malformed")
    result: dict[str, str] = {}
    if not value:
        return result
    for entry in value.split(","):
        if entry.count("=") != 1:
            raise InvalidEvidence("INVALID_EVIDENCE: Docker labels are malformed")
        key, item = entry.split("=", 1)
        if not key or not item or key in result:
            raise InvalidEvidence("INVALID_EVIDENCE: Docker labels are malformed")
        result[key] = item
    return result


def bind_unique_candidate(
    candidates: Sequence[Mapping[str, object]],
    since: str,
    window_seconds: int,
    project: str,
    service: str,
    fixed_backend_image_id: str,
) -> dict[str, Any]:
    """Validate one already-observed candidate for the private test seam.

    Production binding stays in :class:`DockerReadAdapter`; this small pure
    validator has no Docker authority and cannot launch or stop anything.
    """

    if type(candidates) is not tuple or len(candidates) != 1:
        raise InvalidEvidence("INVALID_EVIDENCE: exactly one candidate is required")
    if type(window_seconds) is not int or window_seconds != 120:
        raise InvalidEvidence("INVALID_EVIDENCE: candidate window is not frozen")
    if project != _PROJECT or service != _SERVICE or _SHA256_RE.fullmatch(fixed_backend_image_id) is None:
        raise InvalidEvidence("INVALID_EVIDENCE: candidate scope is not frozen")
    candidate = candidates[0]
    container_id = candidate.get("container_id")
    started_at = candidate.get("started_at")
    image_id = candidate.get("image_id")
    container_image = candidate.get("container_image")
    marker = candidate.get("owner_marker")
    if (
        type(container_id) is not str
        or _ENGINE_ID_RE.fullmatch(container_id) is None
        or image_id != fixed_backend_image_id
        or container_image != fixed_backend_image_id
        or not isinstance(marker, Mapping)
    ):
        raise InvalidEvidence("INVALID_EVIDENCE: candidate identity is malformed")
    labels = _labels(candidate.get("labels"))
    if labels.get("com.docker.compose.project") != project or labels.get("com.docker.compose.service") != service:
        raise InvalidEvidence("INVALID_EVIDENCE: candidate is foreign")
    hostname = marker.get("container_hostname")
    nonce = marker.get("owner_nonce")
    run_id = marker.get("run_id")
    if (
        type(hostname) is not str
        or _HOSTNAME_RE.fullmatch(hostname) is None
        or not container_id.startswith(hostname)
        or type(nonce) is not str
        or _NONCE_RE.fullmatch(nonce) is None
        or type(run_id) is not str
        or _RUN_ID_RE.fullmatch(run_id) is None
        or not run_id.endswith(f"-{hostname}-{nonce[:8]}")
    ):
        raise InvalidEvidence("INVALID_EVIDENCE: candidate owner marker is foreign")
    start = _rfc3339_nanos(started_at)
    lower = _rfc3339_nanos(since)
    if not lower <= start < lower + window_seconds * 1_000_000_000:
        raise InvalidEvidence("INVALID_EVIDENCE: candidate start lies outside its window")
    return {
        "container_id": container_id,
        "container_hostname": hostname,
        "run_id": run_id,
        "owner_nonce": nonce,
        "backend_image_id": fixed_backend_image_id,
        "container_image": fixed_backend_image_id,
        "started_at": started_at,
    }


def _required_candidate(candidate: Mapping[str, object], fixed_images: Mapping[str, str]) -> dict[str, Any]:
    """Normalize the one bound candidate returned by either private runtime."""

    image = fixed_images.get("backend_unit")
    if type(image) is not str or _SHA256_RE.fullmatch(image) is None:
        raise InvalidEvidence("INVALID_EVIDENCE: fixed backend image is malformed")
    fields = ("container_id", "container_hostname", "run_id", "owner_nonce", "container_image")
    if not isinstance(candidate, Mapping) or any(field not in candidate for field in fields):
        raise InvalidEvidence("INVALID_EVIDENCE: bound candidate is incomplete")
    container_id = candidate["container_id"]
    hostname = candidate["container_hostname"]
    run_id = candidate["run_id"]
    nonce = candidate["owner_nonce"]
    container_image = candidate["container_image"]
    if (
        type(container_id) is not str or _ENGINE_ID_RE.fullmatch(container_id) is None
        or type(hostname) is not str or _HOSTNAME_RE.fullmatch(hostname) is None
        or not container_id.startswith(hostname)
        or type(nonce) is not str or _NONCE_RE.fullmatch(nonce) is None
        or type(run_id) is not str or _RUN_ID_RE.fullmatch(run_id) is None
        or not run_id.endswith(f"-{hostname}-{nonce[:8]}")
        or container_image != image
        or candidate.get("backend_image_id", image) != image
    ):
        raise InvalidEvidence("INVALID_EVIDENCE: bound candidate does not retain exact identity")
    return dict(candidate)


def _sample_record(sample: Mapping[str, object], phase: str) -> dict[str, object]:
    required = ("free_ram_bytes", "free_c_bytes", "PercentDiskTime", "CurrentDiskQueueLength")
    if not isinstance(sample, Mapping) or any(key not in sample for key in required):
        raise InvalidEvidence("INVALID_EVIDENCE: host resource sample is incomplete")
    timestamp = sample.get("timestamp", _utc_timestamp())
    if type(timestamp) is not str:
        raise InvalidEvidence("INVALID_EVIDENCE: host resource timestamp is malformed")
    result = {
        "phase": phase,
        "timestamp": timestamp,
        "free_ram_bytes": sample["free_ram_bytes"],
        "free_c_bytes": sample["free_c_bytes"],
        "percent_disk_time": sample["PercentDiskTime"],
        "disk_queue_length": sample["CurrentDiskQueueLength"],
    }
    if (
        type(result["free_ram_bytes"]) is not int
        or type(result["free_c_bytes"]) is not int
        or type(result["percent_disk_time"]) not in (int, float)
        or type(result["disk_queue_length"]) not in (int, float)
    ):
        raise InvalidEvidence("INVALID_EVIDENCE: host resource values are malformed")
    return result


def _policy(raw: Mapping[str, object]) -> dict[str, object]:
    facts: dict[str, object] = {}
    for key in _FACT_KEYS:
        value = raw.get(key, False)
        if type(value) is not bool:
            raise InvalidEvidence("INVALID_EVIDENCE: classification fact is malformed")
        facts[key] = value
    for key in (
        "prohibited_backend_unit_overlap",
        "prohibited_mail_calendar_qa_test_build_overlap",
        "buildkit_idle",
        "whole_run_deadline_expired",
    ):
        default = True if key == "buildkit_idle" else False
        value = raw.get(key, default)
        if type(value) is not bool:
            raise InvalidEvidence("INVALID_EVIDENCE: classification predicate is malformed")
        facts[key] = value
    stops = raw.get("explicit_stops", ())
    if type(stops) is not tuple:
        raise InvalidEvidence("INVALID_EVIDENCE: explicit stops are malformed")
    facts["explicit_stops"] = stops
    return {"sealed_classifications": _sealed_classifications(), "classification_facts": facts}


def _sealed_classifications() -> tuple[str, ...]:
    # `resources` deliberately holds this closed language; exporting a local
    # copy would create another authority.  It is stable internal policy data.
    value = getattr(resources, "_CLASSIFICATIONS", None)
    if type(value) is not tuple or not all(type(item) is str for item in value):
        raise InvalidEvidence("INVALID_EVIDENCE: sealed classification policy is unavailable")
    return value


def _settled_record(samples: tuple[dict[str, object], ...], raw: Mapping[str, object]) -> dict[str, object]:
    policy = _policy(raw)
    try:
        record = resources.classification_record(samples, policy)  # type: ignore[arg-type]
    except Exception as error:
        raise InvalidEvidence("INVALID_EVIDENCE: classification evidence is invalid") from error
    selected = record.get("selected_stop")
    explicit = record.get("explicit_stops")
    if selected is not None and isinstance(explicit, tuple) and selected not in explicit:
        settled = dict(policy["classification_facts"])  # type: ignore[arg-type]
        settled["explicit_stops"] = (*explicit, selected)
        try:
            record = resources.classification_record(
                samples, {"sealed_classifications": _sealed_classifications(), "classification_facts": settled}  # type: ignore[arg-type]
            )
        except Exception as error:
            raise InvalidEvidence("INVALID_EVIDENCE: derived stop could not be sealed") from error
    return record


class _SystemRuntime:
    """The only production authority; all paths derive from this module."""

    def __init__(self) -> None:
        self._repo_root = Path(__file__).resolve().parents[3]
        self._reports = self._repo_root / "testframework" / "reports"
        self._lease: SourceLease | None = None
        self._adapter: DockerReadAdapter | None = None
        self._since: str | None = None
        self._preflight: tuple[dict[str, object], ...] = ()
        self._runtime_samples: tuple[dict[str, object], ...] = ()
        self._concurrency: tuple[dict[str, object], ...] = ()
        self._events: tuple[dict[str, object], ...] = ()
        self._raw: dict[str, object] = {}
        self._handoff: dict[str, Any] | None = None
        self._handoff_digest: str | None = None
        self._manifest: bytes | None = None
        self._fixed_images: dict[str, str] | None = None
        self._terminal: dict[str, Any] | None = None
        self._stop_evidence: dict[str, object] | None = None
        self._terminal_die_time_nano: int | None = None
        self._terminal_die_session: Any | None = None
        self._original_event_session: Any | None = None
        self._guarded_stop_count = 0
        self._stats: tuple[dict[str, object], ...] = ()
        self._event_session: Any | None = None

    def bootstrap(self, argv: tuple[str, ...]) -> None:
        expected_script = Path(__file__).resolve()
        if sys.version_info[:2] != (3, 12):
            raise BootstrapError("observer requires Python 3.12")
        if Path.cwd().resolve() != self._repo_root:
            raise BootstrapError("observer must run from the canonical repository root")
        if __name__ == "__main__":
            if __package__ not in (None, "") or Path(sys.argv[0]).resolve() != expected_script:
                raise BootstrapError("observer requires the exact script invocation")
        if argv != ("--compose-project", _PROJECT):
            raise BootstrapError("observer arguments are not frozen")
        self._adapter = DockerReadAdapter(_PROJECT)

    def preflight(self) -> tuple[dict[str, object], ...]:
        samples: list[dict[str, object]] = []
        for index in range(3):
            try:
                sample = dict(resources.collect_host_sample())
            except Exception as error:
                raise InvalidEvidence("INVALID_EVIDENCE: host preflight evidence is unavailable") from error
            sample["phase"] = "preflight"
            samples.append(sample)
            if index != 2:
                time.sleep(2)
        try:
            resources.preflight(tuple(samples))  # type: ignore[arg-type]
        except Exception as error:
            raise InvalidEvidence("INVALID_EVIDENCE: host preflight contract failed") from error
        self._preflight = tuple(samples)
        return self._preflight

    def acquire_lease(self) -> None:
        if self._lease is not None:
            raise InvalidEvidence("INVALID_EVIDENCE: source lease was already acquired")
        try:
            self._lease = SourceLease.acquire(self._reports / _SOURCE_LEASE, repo_root=self._repo_root)
        except Exception as error:
            raise InvalidEvidence("INVALID_EVIDENCE: source lease cannot be acquired") from error

    def _active_lease(self) -> SourceLease:
        if self._lease is None:
            raise InvalidEvidence("INVALID_EVIDENCE: source lease is unavailable")
        return self._lease

    def snapshot(self, _phase: str) -> dict[str, Any]:
        try:
            return self._active_lease().snapshot()
        except Exception as error:
            raise InvalidEvidence("INVALID_EVIDENCE: source snapshot is unavailable") from error

    def fixed_images(self) -> dict[str, str]:
        try:
            images = self._active_lease().fixed_image_ids()
        except Exception as error:
            raise InvalidEvidence("INVALID_EVIDENCE: fixed image IDs are unavailable") from error
        self._fixed_images = images
        return dict(images)

    def canonical_manifest(self) -> tuple[bytes, str]:
        try:
            manifest = self._active_lease().canonical_build_input_manifest_bytes()
        except Exception as error:
            raise InvalidEvidence("INVALID_EVIDENCE: canonical build manifest is unavailable") from error
        self._manifest = manifest
        return manifest, _sha256(manifest)

    def validate_handoff(self, snapshot: Mapping[str, object], manifest: bytes, images: Mapping[str, str]) -> str:
        try:
            payload = root_handoff.read_root_handoff_bytes()
            handoff = root_handoff.parse_root_handoff(payload)
        except Exception as error:
            raise InvalidEvidence("INVALID_EVIDENCE: final root handoff is unavailable") from error
        if (
            handoff["postgreen"] != snapshot
            or handoff["canonical_build_input_manifest_bytes"].encode("utf-8", "strict") != manifest
            or handoff["canonical_build_input_manifest_sha256"] != _sha256(manifest)
            or handoff["fixed_image_ids"] != dict(images)
        ):
            raise InvalidEvidence("INVALID_EVIDENCE: root handoff does not bind current observer authority")
        self._handoff = handoff
        self._handoff_digest = _sha256(payload)
        return self._handoff_digest

    def ready(self, _session: str) -> None:
        now = datetime.now(timezone.utc)
        self._since = now.strftime("%Y-%m-%dT%H:%M:%S.") + f"{now.microsecond * 1000:09d}Z"

    def _adapter_or_fail(self) -> DockerReadAdapter:
        if self._adapter is None:
            raise InvalidEvidence("INVALID_EVIDENCE: Docker adapter is unavailable")
        return self._adapter

    def recoverable_bound_container_id(self) -> str | None:
        """Return only an adapter-retained exact binding, without Docker I/O."""

        if self._stop_evidence is not None:
            self._validated_stop_evidence()
            raise InvalidEvidence(
                "INVALID_EVIDENCE: an existing guarded-stop attempt must not be retried"
            )
        if self._adapter is None:
            return None
        try:
            container_id = self._adapter.recoverable_bound_container_id()
        except Exception as error:
            raise InvalidEvidence(
                "INVALID_EVIDENCE: recoverable bound container authority is unavailable"
            ) from error
        if container_id is not None and (
            type(container_id) is not str
            or _ENGINE_ID_RE.fullmatch(container_id) is None
        ):
            raise InvalidEvidence(
                "INVALID_EVIDENCE: recoverable bound container identity is malformed"
            )
        return container_id

    def bind_candidate(self, images: dict[str, str]) -> dict[str, Any]:
        since = self._since
        if since is None or images != self._fixed_images:
            raise InvalidEvidence("INVALID_EVIDENCE: bind authority is not initialized")
        adapter = self._adapter_or_fail()
        try:
            container_id = adapter.bind_unique_candidate_since(since, images["backend_unit"])
            inspected = adapter.inspect(container_id)
        except Exception as error:
            raise InvalidEvidence("INVALID_EVIDENCE: unique candidate cannot be bound") from error
        config = inspected.get("Config")
        hostname = config.get("Hostname") if isinstance(config, Mapping) else None
        if (
            type(container_id) is not str
            or _ENGINE_ID_RE.fullmatch(container_id) is None
            or type(hostname) is not str
            or _HOSTNAME_RE.fullmatch(hostname) is None
            or not container_id.startswith(hostname)
            or inspected.get("Image") != images.get("backend_unit")
        ):
            raise InvalidEvidence("INVALID_EVIDENCE: bound candidate identity is incomplete")
        host_config = self._host_config(inspected)
        # This negative fact is meaningful only after the full frozen tmpfs
        # topology and its size options have been validated from inspect.
        self._raw["tmpfs"] = False
        run_dir = self._reports / f"backend-unit-{hostname}.partial"
        owner_bytes = self._read_run(run_dir, _OWNER, _MAX_OWNER_BYTES)
        owner = _strict_json(owner_bytes, "owner marker")
        if tuple(owner) != ("run_id", "owner_nonce", "container_hostname"):
            raise InvalidEvidence("INVALID_EVIDENCE: owner marker keys are not canonical")
        run_id = owner.get("run_id")
        nonce = owner.get("owner_nonce")
        if (
            type(run_id) is not str
            or _RUN_ID_RE.fullmatch(run_id) is None
            or type(nonce) is not str
            or _NONCE_RE.fullmatch(nonce) is None
            or owner.get("container_hostname") != hostname
            or not run_id.endswith(f"-{hostname}-{nonce[:8]}")
        ):
            raise InvalidEvidence("INVALID_EVIDENCE: owner marker does not bind candidate identity")
        expected_owner = {"run_id": run_id, "owner_nonce": nonce, "container_hostname": hostname}
        if owner_bytes != json.dumps(expected_owner, separators=(",", ":"), ensure_ascii=True).encode("ascii"):
            raise InvalidEvidence("INVALID_EVIDENCE: owner marker bytes are not exact")
        return {
            "container_id": container_id, "container_hostname": hostname,
            "run_id": run_id, "owner_nonce": nonce,
            "container_image": images["backend_unit"], "backend_image_id": images["backend_unit"],
            "run_dir": run_dir, "owner_bytes": owner_bytes, "inspect": inspected,
            "host_config": host_config,
        }

    def _read_run(self, run_dir: Path, relative: str, maximum: int) -> bytes:
        try:
            return safe_paths.read_regular_relative(run_dir, relative, max_bytes=maximum)
        except Exception as error:
            raise InvalidEvidence(f"INVALID_EVIDENCE: run artifact {relative} is unavailable") from error

    def _assert_terminal_marker_set(self, run_dir: Path, expected: str) -> None:
        """Fail closed unless the run directory has one exact terminal marker."""

        terminal_names = (_COMPLETE_MARKER, _ABORTED_MARKER)
        if expected not in terminal_names:
            raise InvalidEvidence("INVALID_EVIDENCE: terminal marker expectation is invalid")
        try:
            entries = safe_paths.list_directory_relative(run_dir.parent, run_dir.name)
        except Exception as error:
            raise InvalidEvidence(
                "INVALID_EVIDENCE: terminal marker directory state is unavailable"
            ) from error
        observed = tuple(name for name in entries if name in terminal_names)
        extra = tuple(
            name for name in entries
            if name.startswith("CONTAINER_") and name not in terminal_names
        )
        if extra or observed != (expected,):
            raise InvalidEvidence(
                "INVALID_EVIDENCE: terminal marker set does not select one exact state"
            )

    def _host_config(self, inspected: Mapping[str, object]) -> dict[str, object]:
        host = inspected.get("HostConfig")
        mounts = inspected.get("Mounts")
        if not isinstance(host, Mapping) or not isinstance(mounts, Sequence) or isinstance(mounts, (str, bytes)):
            raise InvalidEvidence("INVALID_EVIDENCE: Docker inspect host configuration is incomplete")
        actual = {
            "memory": host.get("Memory"), "memory_swap": host.get("MemorySwap"),
            "nano_cpus": host.get("NanoCpus"), "read_only_rootfs": host.get("ReadonlyRootfs"),
            "security_opt": host.get("SecurityOpt"), "cap_drop": host.get("CapDrop"),
            "pids_limit": host.get("PidsLimit"), "init": host.get("Init"), "network_mode": host.get("NetworkMode"),
        }
        if actual != _HOST_CONFIG:
            raise InvalidEvidence("INVALID_EVIDENCE: Docker host configuration differs from frozen policy")
        ordered: list[dict[str, object]] = []
        by_destination: dict[str, Mapping[str, object]] = {}
        for raw in mounts:
            if not isinstance(raw, Mapping) or type(raw.get("Destination")) is not str:
                raise InvalidEvidence("INVALID_EVIDENCE: Docker mount evidence is malformed")
            destination = raw["Destination"]
            if destination in by_destination:
                raise InvalidEvidence("INVALID_EVIDENCE: Docker mount evidence is ambiguous")
            by_destination[destination] = raw
        if len(by_destination) != len(_HOST_MOUNTS):
            raise InvalidEvidence("INVALID_EVIDENCE: Docker mount count differs from frozen policy")
        for mount_type, source, destination, read_only, size in _HOST_MOUNTS:
            raw = by_destination.get(destination)
            if raw is None or raw.get("Type") != mount_type or raw.get("RW") is not (not read_only):
                raise InvalidEvidence("INVALID_EVIDENCE: Docker mount differs from frozen policy")
            if mount_type == "bind":
                raw_source = raw.get("Source")
                expected = self._repo_root / str(source).removeprefix("./")
                if type(raw_source) is not str or Path(raw_source).resolve() != expected.resolve():
                    raise InvalidEvidence("INVALID_EVIDENCE: Docker bind source differs from frozen policy")
            else:
                raw_source = raw.get("Source")
                if raw_source not in (None, ""):
                    raise InvalidEvidence("INVALID_EVIDENCE: tmpfs mount has a source")
                options = raw.get("TmpfsOptions")
                observed_size = options.get("SizeBytes") if isinstance(options, Mapping) else None
                if observed_size != size:
                    raise InvalidEvidence("INVALID_EVIDENCE: tmpfs size differs from frozen policy")
            ordered.append({"type": mount_type, "source": source, "destination": destination, "read_only": read_only, "size_bytes": size})
        return {**_HOST_CONFIG, "mounts": ordered}

    def monitor(self, candidate: dict[str, Any]) -> tuple[dict[str, object], ...]:
        adapter = self._adapter_or_fail()
        since = self._since
        if since is None:
            raise InvalidEvidence("INVALID_EVIDENCE: event start is unavailable")
        timing = resources.ObserverTiming()
        deadline = time.monotonic() + timing.whole_run_deadline_seconds
        next_sample = time.monotonic()
        samples: list[dict[str, object]] = []
        concurrency: list[dict[str, object]] = []
        events: list[dict[str, object]] = []
        stats_samples: list[dict[str, object]] = []
        terminal_inspect: dict[str, object] | None = None
        try:
            session = adapter.open_bound_events(since)
        except Exception as error:
            raise InvalidEvidence("INVALID_EVIDENCE: bound event stream is unavailable") from error
        self._event_session = session
        self._original_event_session = session
        interrupted = False
        try:
            while True:
                delay = next_sample - time.monotonic()
                if delay > 0:
                    time.sleep(delay)
                now = time.monotonic()
                if now >= deadline:
                    self._raw["whole_run_deadline_expired"] = True
                try:
                    sample = dict(resources.collect_host_sample())
                    stats = adapter.stats(candidate["container_id"])
                    ps_records = adapter.ps()
                except Exception as error:
                    raise InvalidEvidence("INVALID_EVIDENCE: runtime observation is unavailable") from error
                sample["phase"] = "runtime"
                if type(sample.get("timestamp")) is not str:
                    raise InvalidEvidence("INVALID_EVIDENCE: runtime sample timestamp is malformed")
                samples.append(sample)
                stats_samples.append(self._stats_sample(stats, candidate["container_id"]))
                concurrency.append(self._concurrency_sample(ps_records, sample["timestamp"], candidate["container_id"]))
                self._runtime_samples = tuple(samples)
                self._concurrency = tuple(concurrency)
                self._stats = tuple(stats_samples)
                record = _settled_record(tuple((*self._preflight, *samples)), self._raw)
                self._retain_classification_record(record)
                self._issue_selected_stop(candidate, record)
                event = session.poll()
                if event is not None:
                    status = event.get("status")
                    timestamp = event.get("timeNano")
                    if (
                        status not in ("create", "start", "die", "destroy")
                        or type(timestamp) is not int
                        or event.get("id") != candidate["container_id"]
                    ):
                        raise InvalidEvidence("INVALID_EVIDENCE: lifecycle event is malformed")
                    after_guarded_stop = self._after_guarded_stop(timestamp)
                    events.append({"event": status, "timestamp": _event_timestamp(timestamp)})
                    self._events = tuple(events)
                    if status == "die":
                        if self._stop_evidence is not None and not after_guarded_stop:
                            raise InvalidEvidence("INVALID_EVIDENCE: terminal die preceded the guarded stop")
                        terminal_inspect = adapter.inspect(candidate["container_id"])
                        self._terminal_die_time_nano = timestamp
                        self._terminal_die_session = session
                        break
                    if status == "destroy":
                        raise InvalidEvidence("INVALID_EVIDENCE: container destroyed before terminal inspection")
                next_sample += timing.runtime_sample_interval_seconds
                if next_sample <= time.monotonic():
                    next_sample = time.monotonic() + timing.runtime_sample_interval_seconds
        except KeyboardInterrupt:
            interrupted = True
            raise
        finally:
            if not interrupted:
                self._close_event_session(session)
        if not samples or not concurrency or terminal_inspect is None:
            raise InvalidEvidence("INVALID_EVIDENCE: terminal runtime evidence is incomplete")
        self._runtime_samples = tuple(samples)
        self._concurrency = tuple(concurrency)
        self._events = tuple(events)
        self._stats = tuple(stats_samples)
        self._terminal = dict(terminal_inspect)
        return self._runtime_samples

    def _stats_sample(self, value: object, bound_id: object) -> dict[str, object]:
        if type(value) is not dict or type(bound_id) is not str:
            raise InvalidEvidence("INVALID_EVIDENCE: Docker stats evidence is malformed")
        required = {"BlockIO", "CPUPerc", "Container", "ID", "MemPerc", "MemUsage", "Name", "NetIO", "PIDs"}
        if set(value) != required or not all(type(item) is str for item in value.values()):
            raise InvalidEvidence("INVALID_EVIDENCE: Docker stats fields are not frozen")
        if value["ID"] not in (bound_id, bound_id[:12]):
            raise InvalidEvidence("INVALID_EVIDENCE: Docker stats belong to a foreign container")
        return dict(value)

    def _retain_classification_record(self, record: Mapping[str, object]) -> None:
        explicit = record.get("explicit_stops")
        if type(explicit) is not tuple:
            raise InvalidEvidence("INVALID_EVIDENCE: classification stop evidence is malformed")
        self._raw["explicit_stops"] = explicit
        predicates = record.get("predicates")
        if not isinstance(predicates, Mapping):
            raise InvalidEvidence("INVALID_EVIDENCE: classification predicates are malformed")
        if predicates.get("buildkit_idle") is False:
            self._raw["resource"] = True

    def _issue_selected_stop(self, candidate: Mapping[str, object], record: Mapping[str, object]) -> None:
        selected = record.get("selected_stop")
        if selected is None:
            if self._stop_evidence is not None:
                self._validated_stop_evidence()
                raise InvalidEvidence("INVALID_EVIDENCE: stop evidence exists without a selected predicate")
            return
        if type(selected) is not tuple or len(selected) != 2 or type(selected[0]) is not str or type(selected[1]) is not int:
            raise InvalidEvidence("INVALID_EVIDENCE: selected stop is malformed")
        if self._stop_evidence is not None:
            self._validated_stop_evidence(selected=selected, candidate=candidate)
            return
        if self._terminal_die_time_nano is not None or self._terminal is not None:
            raise InvalidEvidence("INVALID_EVIDENCE: guarded stop cannot target an already-terminal container")
        container_id = candidate.get("container_id")
        if type(container_id) is not str or _ENGINE_ID_RE.fullmatch(container_id) is None:
            raise InvalidEvidence("INVALID_EVIDENCE: guarded-stop container identity is malformed")
        stop_start_time_nano = time.time_ns()
        if type(stop_start_time_nano) is not int or stop_start_time_nano < 0:
            raise InvalidEvidence("INVALID_EVIDENCE: guarded-stop start boundary is malformed")
        actual_argv = ("docker", "stop", "--timeout", "5", container_id)
        self._stop_evidence = {
            "state": "in_flight",
            "selected": selected,
            "argv": actual_argv,
            "start_time_nano": stop_start_time_nano,
        }
        self._guarded_stop_count += 1
        result = self.stop_bound_once(container_id)
        self._stop_evidence = {
            "state": "confirmed",
            "selected": selected,
            "argv": actual_argv,
            "result": result,
            "start_time_nano": stop_start_time_nano,
        }

    def _validated_stop_evidence(
        self,
        *,
        selected: tuple[str, int] | None = None,
        candidate: Mapping[str, object] | None = None,
    ) -> Mapping[str, object]:
        evidence = self._stop_evidence
        if type(evidence) is not dict:
            raise InvalidEvidence("INVALID_EVIDENCE: guarded-stop evidence is malformed")
        state = evidence.get("state")
        observed_selected = evidence.get("selected")
        argv = evidence.get("argv")
        start_time_nano = evidence.get("start_time_nano")
        if (
            type(observed_selected) is not tuple
            or len(observed_selected) != 2
            or type(observed_selected[0]) is not str
            or type(observed_selected[1]) is not int
            or observed_selected[0] not in _STOP_PREDICATES
            or observed_selected[1] < 0
            or type(argv) is not tuple
            or len(argv) != 5
            or argv[:4] != ("docker", "stop", "--timeout", "5")
            or type(argv[4]) is not str
            or _ENGINE_ID_RE.fullmatch(argv[4]) is None
        ):
            raise InvalidEvidence("INVALID_EVIDENCE: guarded-stop evidence is malformed")
        if type(start_time_nano) is not int or start_time_nano < 0:
            raise InvalidEvidence("INVALID_EVIDENCE: guarded-stop start boundary is malformed")
        if selected is not None and observed_selected != selected:
            raise InvalidEvidence("INVALID_EVIDENCE: selected stop changed after the guarded stop")
        if candidate is not None and argv[4] != candidate.get("container_id"):
            raise InvalidEvidence("INVALID_EVIDENCE: guarded-stop container identity changed")
        if state == "in_flight":
            if set(evidence) != {"state", "selected", "argv", "start_time_nano"}:
                raise InvalidEvidence("INVALID_EVIDENCE: in-flight guarded-stop evidence is malformed")
            raise InvalidEvidence("INVALID_EVIDENCE: guarded-stop attempt did not complete")
        if state != "confirmed" or set(evidence) != {"state", "selected", "argv", "result", "start_time_nano"}:
            raise InvalidEvidence("INVALID_EVIDENCE: guarded-stop evidence is malformed")
        result = evidence.get("result")
        if type(result) is not dict or set(result) != {"result"} or result.get("result") != "stopped":
            raise InvalidEvidence("INVALID_EVIDENCE: guarded-stop confirmation is malformed")
        return evidence

    def _validated_stop_start_time_nano(self) -> int:
        evidence = self._validated_stop_evidence()
        start_time_nano = evidence["start_time_nano"]
        if type(start_time_nano) is not int:
            raise InvalidEvidence("INVALID_EVIDENCE: guarded-stop start boundary is malformed")
        return start_time_nano

    def _after_guarded_stop(self, event_time_nano: object) -> bool:
        if type(event_time_nano) is not int or event_time_nano < 0:
            raise InvalidEvidence("INVALID_EVIDENCE: lifecycle event timestamp is malformed")
        if self._stop_evidence is None:
            return False
        return event_time_nano > self._validated_stop_start_time_nano()

    def _selected_stop_for_terminal(
        self, settled_record: Mapping[str, object]
    ) -> tuple[str, int] | None:
        """Accept only the frozen, settled selection passed from ``main``."""

        selected = settled_record.get("selected_stop")
        explicit = settled_record.get("explicit_stops")
        if type(explicit) is not tuple:
            raise InvalidEvidence("INVALID_EVIDENCE: terminal selected-stop evidence is malformed")
        if selected is None:
            return None
        if (
            type(selected) is not tuple
            or len(selected) != 2
            or type(selected[0]) is not str
            or selected[0] not in _STOP_PREDICATES
            or type(selected[1]) is not int
            or selected[1] < 0
            or selected not in explicit
        ):
            raise InvalidEvidence("INVALID_EVIDENCE: terminal selected-stop evidence is malformed")
        return selected

    def _validated_aborted_terminal_die(
        self, candidate: Mapping[str, object], selected: tuple[str, int]
    ) -> Mapping[str, object]:
        """Bind one confirmed guarded stop to the die from its original session."""

        evidence = self._validated_stop_evidence(selected=selected, candidate=candidate)
        if self._guarded_stop_count != 1:
            raise InvalidEvidence("INVALID_EVIDENCE: aborted receipt requires exactly one guarded stop")
        die_time_nano = self._terminal_die_time_nano
        session = self._original_event_session
        if (
            type(die_time_nano) is not int
            or die_time_nano < 0
            or session is None
            or self._terminal_die_session is not session
            or not self._after_guarded_stop(die_time_nano)
        ):
            raise InvalidEvidence("INVALID_EVIDENCE: aborted receipt lacks a guarded terminal die")
        expected_event = {"event": "die", "timestamp": _event_timestamp(die_time_nano)}
        if sum(1 for event in self._events if event == expected_event) != 1:
            raise InvalidEvidence("INVALID_EVIDENCE: terminal die event is not uniquely retained")
        return evidence

    def _record_interrupt_and_stop(self, candidate: Mapping[str, object], samples: Sequence[dict[str, object]]) -> None:
        earlier = _settled_record(tuple((*self._preflight, *samples)), self._raw).get("selected_stop")
        existing = self._raw.get("explicit_stops", ())
        if type(existing) is not tuple:
            raise InvalidEvidence("INVALID_EVIDENCE: interrupt stop evidence is malformed")
        self._raw["explicit_stops"] = (*existing, ("interrupt", time.time_ns()))
        record = _settled_record(tuple((*self._preflight, *samples)), self._raw)
        if earlier is not None and record.get("selected_stop") != earlier:
            raise InvalidEvidence("INVALID_EVIDENCE: interrupt replaced an earlier selected stop")
        self._retain_classification_record(record)
        self._issue_selected_stop(candidate, record)

    def recover_interrupt(self, candidate: Mapping[str, object]) -> tuple[dict[str, object], ...]:
        """Drain the original bound event session through the guarded terminal die."""

        session = self._event_session
        if session is None:
            raise InvalidEvidence("INVALID_EVIDENCE: original interrupted event session is unavailable")
        try:
            poll = getattr(session, "poll", None)
            close = getattr(session, "close", None)
            if not callable(poll) or not callable(close):
                raise InvalidEvidence("INVALID_EVIDENCE: original interrupted event session is malformed")
            container_id = candidate.get("container_id")
            if type(container_id) is not str or _ENGINE_ID_RE.fullmatch(container_id) is None:
                raise InvalidEvidence("INVALID_EVIDENCE: interrupted container identity is malformed")
            if self._terminal_die_time_nano is not None or self._terminal is not None:
                raise InvalidEvidence("INVALID_EVIDENCE: interrupt arrived after terminal container evidence")
            self._record_interrupt_and_stop(candidate, self._runtime_samples)
            adapter = self._adapter_or_fail()
            deadline = time.monotonic() + resources.ObserverTiming().destroy_grace_seconds
            events = list(self._events)
            while True:
                event = poll()
                if event is None:
                    remaining = deadline - time.monotonic()
                    if remaining <= 0:
                        raise InvalidEvidence("INVALID_EVIDENCE: interrupted container has no terminal die evidence")
                    time.sleep(min(0.01, remaining))
                    continue
                status = event.get("status")
                timestamp = event.get("timeNano")
                if (
                    status not in ("create", "start", "die", "destroy")
                    or type(timestamp) is not int
                    or event.get("id") != container_id
                ):
                    raise InvalidEvidence("INVALID_EVIDENCE: interrupted lifecycle event is malformed")
                after_guarded_stop = self._after_guarded_stop(timestamp)
                events.append({"event": status, "timestamp": _event_timestamp(timestamp)})
                self._events = tuple(events)
                if status != "die":
                    continue
                if not after_guarded_stop:
                    raise InvalidEvidence("INVALID_EVIDENCE: interrupted terminal die preceded the guarded stop")
                self._terminal = dict(adapter.inspect(container_id))
                self._terminal_die_time_nano = timestamp
                self._terminal_die_session = session
                return self._runtime_samples
        finally:
            self._close_event_session(session)

    def _close_event_session(self, session: object) -> None:
        try:
            close = getattr(session, "close")
            if not callable(close):
                raise TypeError("event session has no close method")
            close()
        except Exception as error:
            raise InvalidEvidence("INVALID_EVIDENCE: event stream cleanup is uncertain") from error
        finally:
            if self._event_session is session:
                self._event_session = None

    def _concurrency_sample(
        self, records: Sequence[Mapping[str, object]], timestamp: object, bound_id: object
    ) -> dict[str, object]:
        if type(timestamp) is not str or type(bound_id) is not str or _ENGINE_ID_RE.fullmatch(bound_id) is None:
            raise InvalidEvidence("INVALID_EVIDENCE: concurrency identity is malformed")
        backend_count = 0
        mail_calendar_count = 0
        buildkit_busy = False
        observed_bound = False
        for record in records:
            if type(record) is not dict or set(record) != _PS_FIELDS:
                raise InvalidEvidence("INVALID_EVIDENCE: Docker ps fields are not frozen")
            if not all(type(record[name]) is str for name in _PS_FIELDS):
                raise InvalidEvidence("INVALID_EVIDENCE: Docker ps value is malformed")
            record_id = record["ID"]
            if _ENGINE_ID_RE.fullmatch(record_id) is None:
                raise InvalidEvidence("INVALID_EVIDENCE: Docker ps ID is not full and canonical")
            labels = _labels(record["Labels"])
            project = labels.get("com.docker.compose.project")
            service = labels.get("com.docker.compose.service")
            if (project is None) != (service is None):
                raise InvalidEvidence("INVALID_EVIDENCE: Docker compose labels are incomplete")
            if project is None:
                if (
                    record["Image"] != _BUILDKIT_IMAGE
                    or not record["Names"].startswith("buildx_buildkit_")
                    or labels
                ):
                    raise InvalidEvidence("INVALID_EVIDENCE: Docker ps foreign or unknown evidence")
                buildkit_busy = True
                continue
            if project != _PROJECT or service not in (_NON_QA_COMPOSE_SERVICES | _QA_COMPOSE_SERVICES):
                raise InvalidEvidence("INVALID_EVIDENCE: Docker compose tuple is not allowed")
            if service == _SERVICE:
                backend_count += 1
                if record_id == bound_id:
                    observed_bound = True
            if service == "mail-calendar-qa-test-build":
                mail_calendar_count += 1
        if backend_count < 1 or not observed_bound:
            raise InvalidEvidence("INVALID_EVIDENCE: Docker ps does not contain the bound backend-unit")
        self._raw["prohibited_backend_unit_overlap"] = backend_count > 1
        self._raw["prohibited_mail_calendar_qa_test_build_overlap"] = mail_calendar_count > 0
        self._raw["buildkit_idle"] = not buildkit_busy
        if buildkit_busy:
            self._raw["resource"] = True
        return {
            "phase": "runtime", "timestamp": timestamp,
            "backend_unit_count": backend_count,
            "mail_calendar_qa_test_build_count": mail_calendar_count,
            "buildkit_state": "busy" if buildkit_busy else "idle",
        }

    def raw_observations_for(self, _candidate: dict[str, Any], _samples: tuple[dict[str, object], ...]) -> dict[str, object]:
        return dict(self._raw)

    def stop_bound_once(self, container_id: str) -> dict[str, object]:
        try:
            result = self._adapter_or_fail().stop_bound_once(container_id)
        except Exception as error:
            raise InvalidEvidence("INVALID_EVIDENCE: guarded Docker stop failed") from error
        if result != {"result": "stopped"}:
            raise InvalidEvidence("INVALID_EVIDENCE: guarded Docker stop confirmation is malformed")
        return result

    def terminal_verify(
        self, candidate: dict[str, Any], settled_record: Mapping[str, object]
    ) -> None:
        """Verify the receipt branch selected by the settled observer outcome."""

        terminal = self._terminal
        if terminal is None:
            raise InvalidEvidence("INVALID_EVIDENCE: terminal inspect was not captured")
        if terminal.get("Id") != candidate["container_id"] or terminal.get("Image") != candidate["container_image"]:
            raise InvalidEvidence("INVALID_EVIDENCE: terminal container identity drifted")
        state = terminal.get("State")
        if not isinstance(state, Mapping) or state.get("Status") != "exited" or type(state.get("ExitCode")) is not int:
            raise InvalidEvidence("INVALID_EVIDENCE: terminal container state is malformed")
        run_dir = candidate.get("run_dir")
        owner_bytes = candidate.get("owner_bytes")
        if not isinstance(run_dir, Path) or type(owner_bytes) is not bytes:
            raise InvalidEvidence("INVALID_EVIDENCE: terminal run directory is malformed")
        if self._read_run(run_dir, _OWNER, _MAX_OWNER_BYTES) != owner_bytes:
            raise InvalidEvidence("INVALID_EVIDENCE: owner marker changed")
        receipt_bytes = self._read_run(run_dir, _CONTAINER_RECEIPT, _MAX_RECEIPT_BYTES)
        receipt = _strict_json(receipt_bytes, "container receipt")
        if _canonical_json(receipt) != receipt_bytes:
            raise InvalidEvidence("INVALID_EVIDENCE: container receipt is not canonical")
        try:
            contracts.validate_container_receipt(receipt)
        except Exception as error:
            raise InvalidEvidence("INVALID_EVIDENCE: container receipt violates its contract") from error
        selected = self._selected_stop_for_terminal(settled_record)
        receipt_state = receipt.get("state")
        if receipt_state == "CONTAINER_COMPLETE":
            if selected is not None or self._stop_evidence is not None:
                raise InvalidEvidence("INVALID_EVIDENCE: complete receipt cannot bind a guarded stop")
            marker = _COMPLETE_MARKER
        elif receipt_state == "CONTAINER_ABORTED":
            if selected is None:
                raise InvalidEvidence("INVALID_EVIDENCE: aborted receipt lacks a selected stop")
            marker = _ABORTED_MARKER
        else:
            raise InvalidEvidence("INVALID_EVIDENCE: container receipt state is not terminal")
        self._assert_terminal_marker_set(run_dir, marker)
        if self._read_run(run_dir, marker, 0) != b"":
            raise InvalidEvidence("INVALID_EVIDENCE: terminal marker is not empty")
        confirmed_stop = self._validate_container_bindings(
            receipt, run_dir, candidate, state, selected
        )
        # ``_validate_container_bindings`` has now bound either the complete
        # plan/artifact set or the contract-valid aborted discovery state.
        self._raw["discovery_failure"] = False
        budget = receipt.get("report_budget")
        terminal_receipt = receipt.get("terminal")
        if not isinstance(budget, Mapping) or not isinstance(terminal_receipt, Mapping):
            raise InvalidEvidence("INVALID_EVIDENCE: terminal receipt evidence is malformed")
        truncated_due_to_budget = budget.get("truncated_due_to_budget")
        if type(truncated_due_to_budget) is not bool:
            raise InvalidEvidence("INVALID_EVIDENCE: report budget evidence is malformed")
        self._raw["budget"] = truncated_due_to_budget
        self._raw["timeout"] = bool(
            confirmed_stop is not None and confirmed_stop["selected"][0] == "deadline"
        )
        if receipt_state == "CONTAINER_COMPLETE":
            self._raw["stage_failure"] = terminal_receipt["exit_code"] != 0
        candidate["container_receipt_bytes"] = receipt_bytes
        candidate["container_receipt"] = receipt
        candidate["cleanup"] = {"container_removed": False}
        try:
            adapter_images = self._adapter_or_fail().inspect_fixed_images()
            lease_images = self._active_lease().fixed_image_ids()
            manifest = self._active_lease().canonical_build_input_manifest_bytes()
        except Exception as error:
            raise InvalidEvidence("INVALID_EVIDENCE: terminal image or manifest evidence is unavailable") from error
        if adapter_images != self._fixed_images or lease_images != self._fixed_images or manifest != self._manifest:
            raise InvalidEvidence("INVALID_EVIDENCE: fixed image or manifest evidence drifted")

    def _validate_container_bindings(
        self,
        receipt: Mapping[str, object],
        run_dir: Path,
        candidate: Mapping[str, object],
        state: Mapping[str, object],
        selected: tuple[str, int] | None,
    ) -> Mapping[str, object] | None:
        terminal = receipt.get("terminal")
        if (
            receipt.get("run_id") != candidate["run_id"]
            or receipt.get("owner_nonce") != candidate["owner_nonce"]
            or receipt.get("container_hostname") != candidate["container_hostname"]
            or not isinstance(terminal, Mapping)
        ):
            raise InvalidEvidence("INVALID_EVIDENCE: container receipt does not bind terminal identity")
        receipt_state = receipt.get("state")
        confirmed_stop: Mapping[str, object] | None = None
        if receipt_state == "CONTAINER_COMPLETE":
            if (
                selected is not None
                or self._stop_evidence is not None
                or terminal.get("signal") is not None
                or terminal.get("exit_code") != state.get("ExitCode")
                or state.get("Error") not in (None, "")
            ):
                raise InvalidEvidence("INVALID_EVIDENCE: complete receipt does not bind terminal identity")
        elif receipt_state == "CONTAINER_ABORTED":
            if selected is None:
                raise InvalidEvidence("INVALID_EVIDENCE: aborted receipt lacks a selected stop")
            confirmed_stop = self._validated_aborted_terminal_die(candidate, selected)
            if (
                terminal.get("signal") != "TERM"
                or terminal.get("exit_code") != 143
                or state.get("Status") != "exited"
                or state.get("ExitCode") != 143
                or state.get("OOMKilled") is not False
                or state.get("Error") not in (None, "")
            ):
                raise InvalidEvidence("INVALID_EVIDENCE: aborted receipt does not bind guarded termination")
        else:
            raise InvalidEvidence("INVALID_EVIDENCE: container receipt state is not terminal")
        before = receipt.get("source_manifest_before_sha256")
        after = receipt.get("source_manifest_after_sha256")
        if type(before) is not str or _SHA256_RE.fullmatch(before) is None or before != after:
            raise InvalidEvidence("INVALID_EVIDENCE: container source manifest changed")
        cgroup = receipt.get("cgroup")
        if not isinstance(cgroup, Mapping) or cgroup.get("baseline") != cgroup.get("final"):
            raise InvalidEvidence("INVALID_EVIDENCE: cgroup evidence changed")
        snapshot = cgroup.get("final")
        if not isinstance(snapshot, Mapping):
            raise InvalidEvidence("INVALID_EVIDENCE: cgroup snapshot is malformed")
        if cgroup.get("family") == "v2":
            events = snapshot.get("memory.events")
            if snapshot.get("memory.max") != "2147483648" or snapshot.get("memory.swap.max") != "0" or not isinstance(events, Mapping):
                raise InvalidEvidence("INVALID_EVIDENCE: v2 cgroup contract differs from runner")
            oom = events.get("oom", 0)
            oom_kill = events.get("oom_kill", 0)
        elif cgroup.get("family") == "v1":
            events = snapshot.get("memory.oom_control")
            if snapshot.get("memory.limit_in_bytes") != "2147483648" or snapshot.get("memory.memsw.limit_in_bytes") != "2147483648" or not isinstance(events, Mapping):
                raise InvalidEvidence("INVALID_EVIDENCE: v1 cgroup contract differs from runner")
            oom = events.get("under_oom", 0)
            oom_kill = events.get("oom_kill", 0)
        else:
            raise InvalidEvidence("INVALID_EVIDENCE: cgroup family is not frozen")
        if type(oom) is not int or type(oom_kill) is not int or oom < 0 or oom_kill < 0:
            raise InvalidEvidence("INVALID_EVIDENCE: cgroup OOM evidence is malformed")
        self._raw["oom"] = bool(state.get("OOMKilled")) or oom > 0 or oom_kill > 0
        inventory = receipt.get("artifact_inventory")
        if not isinstance(inventory, Sequence) or isinstance(inventory, (str, bytes)):
            raise InvalidEvidence("INVALID_EVIDENCE: container artifact inventory is malformed")
        try:
            paths = tuple(item["path"] for item in inventory if isinstance(item, Mapping))
        except (KeyError, TypeError) as error:
            raise InvalidEvidence("INVALID_EVIDENCE: container artifact inventory is malformed") from error
        if receipt_state == "CONTAINER_COMPLETE":
            expected_paths = _CONTAINER_ARTIFACTS
        else:
            expected_paths = self._aborted_artifact_paths(receipt)
        if len(paths) != len(inventory) or paths != expected_paths:
            raise InvalidEvidence("INVALID_EVIDENCE: container artifact inventory paths are not frozen")
        inventory_lines: list[bytes] = []
        for item in inventory:
            if not isinstance(item, Mapping):
                raise InvalidEvidence("INVALID_EVIDENCE: container artifact row is malformed")
            path, digest, size = item.get("path"), item.get("sha256"), item.get("size")
            if type(path) is not str or type(digest) is not str or type(size) is not int or size < 0:
                raise InvalidEvidence("INVALID_EVIDENCE: container artifact row is malformed")
            payload = self._read_run(run_dir, path, _MAX_RECEIPT_BYTES)
            if len(payload) != size or _sha256(payload) != digest:
                raise InvalidEvidence("INVALID_EVIDENCE: container artifact bytes changed")
            inventory_lines.append(f"{path}\t{digest}\t{size}\n".encode("ascii"))
        if receipt.get("artifact_inventory_sha256") != _sha256(b"".join(inventory_lines)):
            raise InvalidEvidence("INVALID_EVIDENCE: container artifact inventory hash drifted")
        if receipt_state == "CONTAINER_ABORTED":
            return confirmed_stop
        plans = receipt.get("package_plan_sha256")
        stages = receipt.get("stage_receipt_sha256")
        if not isinstance(plans, Mapping) or not isinstance(stages, Mapping) or set(plans) != {"build", "vet", "unit", "race"} or set(stages) != set(plans):
            raise InvalidEvidence("INVALID_EVIDENCE: container stage digest maps are malformed")
        for stage in ("build", "vet", "unit", "race"):
            plan_payload = self._read_run(run_dir, f"package-plan.{stage}.json", _MAX_RECEIPT_BYTES)
            stage_payload = self._read_run(run_dir, f"stage-{stage}.json", _MAX_RECEIPT_BYTES)
            results_payload = self._read_run(run_dir, f"package-results.{stage}.jsonl", _MAX_RECEIPT_BYTES)
            if _sha256(plan_payload) != plans[stage] or _sha256(stage_payload) != stages[stage]:
                raise InvalidEvidence("INVALID_EVIDENCE: container stage artifact hash drifted")
            stage_document = _strict_json(stage_payload, f"stage {stage} receipt")
            if (
                stage_document.get("run_id") != candidate["run_id"]
                or stage_document.get("owner_nonce") != candidate["owner_nonce"]
                or stage_document.get("stage") != stage
                or stage_document.get("package_plan_sha256") != plans[stage]
                or stage_document.get("package_result_sha256") != _sha256(results_payload)
            ):
                raise InvalidEvidence("INVALID_EVIDENCE: container stage receipt does not bind artifacts")
        return confirmed_stop

    def _aborted_artifact_paths(self, receipt: Mapping[str, object]) -> tuple[str, ...]:
        """Derive the ABORTED inventory without touching COMPLETE-only digest maps."""

        abort = receipt.get("abort")
        if not isinstance(abort, Mapping):
            raise InvalidEvidence("INVALID_EVIDENCE: aborted receipt has no abort evidence")
        planned = abort.get("planned_stages")
        completed = abort.get("completed_stages")
        if (
            not isinstance(planned, Sequence)
            or isinstance(planned, (str, bytes))
            or not isinstance(completed, Sequence)
            or isinstance(completed, (str, bytes))
            or any(type(stage) is not str for stage in planned)
            or any(type(stage) is not str for stage in completed)
        ):
            raise InvalidEvidence("INVALID_EVIDENCE: aborted receipt stage evidence is malformed")
        paths = {_ABORTED_MARKER, _OWNER}
        for stage in planned:
            paths.add(f"package-plan.{stage}.json")
        for stage in completed:
            paths.update((
                f"{stage}.stderr.log", f"{stage}.stdout.log",
                f"package-results.{stage}.jsonl", f"stage-{stage}.json",
            ))
        return tuple(sorted(paths))

    def verify_chain(self, before: Mapping[str, object], after: Mapping[str, object]) -> None:
        handoff = self._handoff
        if handoff is None or handoff["prebuild"] != before or handoff["postgreen"] != before:
            raise InvalidEvidence("INVALID_EVIDENCE: root handoff source chain drifted")
        try:
            self._active_lease().verify_snapshot_chain(handoff["prebuild"], handoff["postgreen"], after)
        except Exception as error:
            raise InvalidEvidence("INVALID_EVIDENCE: observer source chain drifted") from error

    def build_receipts(
        self,
        candidate: Mapping[str, object],
        before: Mapping[str, object],
        after: Mapping[str, object],
        record: Mapping[str, object],
    ) -> tuple[dict[str, object], dict[str, object]]:
        if self._handoff is None or self._handoff_digest is None or self._manifest is None or self._fixed_images is None:
            raise InvalidEvidence("INVALID_EVIDENCE: receipt authority is incomplete")
        receipt_bytes = candidate.get("container_receipt_bytes")
        host_config = candidate.get("host_config")
        cleanup = candidate.get("cleanup")
        if type(receipt_bytes) is not bytes or not isinstance(host_config, Mapping) or not isinstance(cleanup, Mapping):
            raise InvalidEvidence("INVALID_EVIDENCE: terminal receipt evidence is incomplete")
        selected = record.get("selected_stop")
        explicit = record.get("explicit_stops")
        if type(explicit) is not tuple or any(type(item) is not tuple or len(item) != 2 for item in explicit):
            raise InvalidEvidence("INVALID_EVIDENCE: explicit stop receipt evidence is malformed")
        if selected is not None and (not isinstance(selected, tuple) or len(selected) != 2 or selected not in explicit):
            raise InvalidEvidence("INVALID_EVIDENCE: selected stop is malformed")
        for key in ("tmpfs", "discovery_failure"):
            if key not in self._raw or type(self._raw[key]) is not bool:
                raise InvalidEvidence(
                    f"INVALID_EVIDENCE: measured classification fact {key} is unavailable"
                )
        facts = {
            key: self._raw[key] if key in ("tmpfs", "discovery_failure") else self._raw.get(key, False)
            for key in _FACT_KEYS
        }
        facts.update({"predicates": record["predicates"], "explicit_stops": [list(item) for item in explicit], "selected_stop": list(selected) if selected else None, "classification": record["classification"], "immutable_outcome": record["immutable_outcome"]})
        stop_argv: list[str] | None = None
        guarded: dict[str, object] = {"predicate": None, "argv": None, "result": None}
        evidence = self._stop_evidence
        if selected is None:
            if evidence is not None:
                self._validated_stop_evidence(candidate=candidate)
                raise InvalidEvidence("INVALID_EVIDENCE: an unselected stop cannot be sealed")
        else:
            if evidence is None:
                raise InvalidEvidence("INVALID_EVIDENCE: selected stop lacks actual guarded-stop evidence")
            evidence = self._validated_stop_evidence(selected=selected, candidate=candidate)
            argv = evidence["argv"]
            result = evidence["result"]
            expected_argv = ("docker", "stop", "--timeout", "5", candidate["container_id"])
            if tuple(argv) != expected_argv or result != {"result": "stopped"}:
                raise InvalidEvidence("INVALID_EVIDENCE: guarded-stop evidence is not exact")
            stop_argv = list(expected_argv)
            guarded = {"predicate": _STOP_PREDICATES[selected[0]], "argv": stop_argv, "result": result["result"]}
        seal_timestamp = _utc_timestamp()
        final = {
            "schema": "hub.qa.backend-unit-final-receipt/v1", "target": "backend-unit",
            "run_id": candidate["run_id"], "owner_nonce": candidate["owner_nonce"],
            "container_receipt_sha256": _sha256(receipt_bytes), "container_id": candidate["container_id"],
            "backend_image_id": self._fixed_images["backend_unit"],
            "canonical_build_input_manifest_sha256": _sha256(self._manifest),
            "classification": record["classification"], "immutable_outcome": record["immutable_outcome"],
            "seal_timestamp": seal_timestamp,
        }
        final_bytes = _canonical_json(final)
        inventory = [
            {"path": "container-receipt.json", "sha256": _sha256(receipt_bytes), "size": len(receipt_bytes)},
            {"path": "final-receipt.json", "sha256": _sha256(final_bytes), "size": len(final_bytes)},
        ]
        inventory_bytes = b"".join(f"{item['path']}\t{item['sha256']}\t{item['size']}\n".encode("ascii") for item in inventory)
        observer = {
            "schema": "hub.qa.backend-unit-observer-receipt/v1", "target": "backend-unit",
            "run_id": candidate["run_id"], "owner_nonce": candidate["owner_nonce"],
            "container_id": candidate["container_id"], "container_image": candidate["container_image"],
            "container_hostname": candidate["container_hostname"], "container_receipt_sha256": _sha256(receipt_bytes),
            "compose_sha256": _manifest_compose_digest(self._manifest),
            "git": {"head_before": before["head"], "head_after": after["head"], "status_before": before["status"], "status_after": after["status"]},
            "frozen_subjects": after["frozen_subjects"],
            "source_snapshots": {"prebuild": self._handoff["prebuild"], "postgreen": self._handoff["postgreen"], "observer_snapshot": dict(after)},
            "root_handoff_sha256": self._handoff_digest,
            "concurrency_observation_scope": _CONCURRENCY_SCOPE,
            "canonical_build_input_manifest_bytes": self._manifest.decode("utf-8", "strict"),
            "canonical_build_input_manifest_sha256": _sha256(self._manifest),
            "fixed_image_ids": dict(self._fixed_images), "bound_backend_image_id": self._fixed_images["backend_unit"],
            "host_config": dict(host_config),
            "resource_samples": [_sample_record(sample, "preflight") for sample in self._preflight] + [_sample_record(sample, "runtime") for sample in self._runtime_samples],
            "concurrency_samples": list(self._concurrency), "lifecycle_events": list(self._events),
            "classification_facts": facts, "classification": record["classification"], "immutable_outcome": record["immutable_outcome"],
            "selected_stop": list(selected) if selected else None, "stop_argv": stop_argv, "guarded_stop": guarded,
            "artifact_inventory": inventory, "artifact_inventory_sha256": _sha256(inventory_bytes),
            "cleanup": dict(cleanup), "final_receipt": final, "seal_timestamp": seal_timestamp,
        }
        try:
            contracts.validate_final_receipt(final)
            contracts.validate_observer_receipt(observer)
        except Exception as error:
            raise InvalidEvidence("INVALID_EVIDENCE: observer receipt is not schema-valid") from error
        return final, observer

    def publish(self, final: dict[str, object], observer: dict[str, object]) -> None:
        run_dir = self._reports / f"backend-unit-{observer['container_hostname']}.partial"
        try:
            contracts.validate_final_receipt(final)
            contracts.validate_observer_receipt(observer)
            publication.publish(run_dir, final, observer)
        except Exception as error:
            raise InvalidEvidence("INVALID_EVIDENCE: receipt publication is not authoritative") from error

    def release_lease(self) -> None:
        if self._lease is not None:
            self._lease.release()
            self._lease = None


def _event_timestamp(value: int) -> str:
    if type(value) is not int or value < 0:
        raise InvalidEvidence("INVALID_EVIDENCE: event timestamp is malformed")
    seconds, nanos = divmod(value, 1_000_000_000)
    rendered = datetime.fromtimestamp(seconds, timezone.utc).strftime("%Y-%m-%dT%H:%M:%S")
    return f"{rendered}.{nanos:09d}Z"


def _manifest_compose_digest(manifest: bytes) -> str:
    if not manifest.endswith(b"\n") or b"\r" in manifest:
        raise InvalidEvidence("INVALID_EVIDENCE: canonical manifest is malformed")
    first = manifest.split(b"\n", 1)[0].split(b"\t")
    if len(first) != 2 or first[0] != b"docker-compose.local.yml":
        raise InvalidEvidence("INVALID_EVIDENCE: canonical manifest compose row is malformed")
    try:
        digest = first[1].decode("ascii")
    except UnicodeDecodeError as error:
        raise InvalidEvidence("INVALID_EVIDENCE: canonical manifest compose digest is malformed") from error
    if _SHA256_RE.fullmatch(digest) is None:
        raise InvalidEvidence("INVALID_EVIDENCE: canonical manifest compose digest is malformed")
    return digest


def _legacy_receipts(
    candidate: Mapping[str, object],
    images: Mapping[str, str],
    manifest: bytes,
    manifest_digest: str,
    record: Mapping[str, object],
) -> tuple[dict[str, object], dict[str, object]]:
    """Private compatibility payload for an injected selftest runtime only."""

    selected = record["selected_stop"]
    stop_argv = ("docker", "stop", "--timeout", "5", candidate["container_id"]) if selected else None
    final = {
        "schema": "hub.qa.backend-unit-final-receipt/v1", "target": "backend-unit",
        "run_id": candidate["run_id"], "owner_nonce": candidate["owner_nonce"],
        "container_id": candidate["container_id"], "container_receipt_sha256": "sha256:" + "0" * 64,
        "backend_image_id": images["backend_unit"], "canonical_build_input_manifest_sha256": manifest_digest,
        "classification": record["classification"], "immutable_outcome": record["immutable_outcome"], "seal_timestamp": _utc_timestamp(),
    }
    observer = {
        "run_id": candidate["run_id"], "owner_nonce": candidate["owner_nonce"], "container_id": candidate["container_id"],
        "container_hostname": candidate["container_hostname"], "container_image": images["backend_unit"],
        "fixed_image_ids": dict(images), "canonical_build_input_manifest_bytes": manifest.decode("utf-8"),
        "canonical_build_input_manifest_sha256": manifest_digest, "classification": record["classification"],
        "immutable_outcome": record["immutable_outcome"], "selected_stop": selected, "stop_argv": stop_argv,
    }
    return final, observer


def _new_system_runtime() -> _SystemRuntime:
    return _SystemRuntime()


def _exact_argv(argv: object) -> tuple[str, str] | None:
    if type(argv) is not tuple or argv != ("--compose-project", _PROJECT):
        return None
    return argv


def _runtime_candidate(value: object, images: Mapping[str, str]) -> dict[str, Any]:
    """Keep the injected selftest runtime private; production is already normalized."""

    if not isinstance(value, Mapping):
        raise InvalidEvidence("INVALID_EVIDENCE: runtime candidate is malformed")
    required = {"container_id", "container_hostname", "run_id", "owner_nonce", "container_image"}
    if required <= set(value):
        return _required_candidate(value, images)
    started_at = value.get("started_at")
    if type(started_at) is not str:
        raise InvalidEvidence("INVALID_EVIDENCE: runtime candidate has no start timestamp")
    return bind_unique_candidate((value,), started_at, 120, _PROJECT, _SERVICE, images["backend_unit"])


def main(argv: Sequence[str] | None = None) -> int:
    """Run the observer; 64 is usage, 75 is unsealed evidence, 130 pre-bind interrupt."""

    supplied: object = tuple(sys.argv[1:]) if argv is None else argv
    exact = _exact_argv(supplied)
    if exact is None:
        return 64
    runtime = _new_system_runtime()
    leased = False
    result = 75
    try:
        try:
            runtime.bootstrap(exact)
        except AssertionError as error:
            raise BootstrapError("observer bootstrap contract failed") from error
        runtime.preflight()
        runtime.acquire_lease()
        leased = True
        before = runtime.snapshot("before")
        images = runtime.fixed_images()
        manifest, manifest_digest = runtime.canonical_manifest()
        validate_handoff = getattr(runtime, "validate_handoff", None)
        if callable(validate_handoff):
            validate_handoff(before, manifest, images)
        session = secrets.token_hex(16)
        runtime.ready(session)
        sys.stdout.write(f"READY observer_session={session}\n")
        sys.stdout.flush()
        candidate = _runtime_candidate(runtime.bind_candidate(images), images)
        try:
            monitored = runtime.monitor(candidate)
            raw = runtime.raw_observations_for(candidate, monitored)
            record = _settled_record(monitored, raw)
            if isinstance(runtime, _SystemRuntime) and (raw.get("source_mutated") or raw.get("invalid_provenance")):
                raise InvalidEvidence("INVALID_EVIDENCE: source or provenance authority was lost")
            if not isinstance(runtime, _SystemRuntime) and record["selected_stop"] is not None and record["classification"] not in ("SEALED_INVALID_SOURCE_MUTATED", "SEALED_INVALID_PROVENANCE"):
                runtime.stop_bound_once(candidate["container_id"])
            if isinstance(runtime, _SystemRuntime):
                runtime.terminal_verify(candidate, record)
            else:
                runtime.terminal_verify(candidate)
            after = runtime.snapshot("after")
            if before != after:
                raise InvalidEvidence("INVALID_EVIDENCE: source changed during observation")
            verify_chain = getattr(runtime, "verify_chain", None)
            if callable(verify_chain):
                verify_chain(before, after)
            if isinstance(runtime, _SystemRuntime):
                terminal_facts = runtime.raw_observations_for(candidate, monitored)
                record = _settled_record(tuple((*runtime._preflight, *monitored)), terminal_facts)
                final, observer = runtime.build_receipts(candidate, before, after, record)
            else:
                final, observer = _legacy_receipts(candidate, images, manifest, manifest_digest, record)
            runtime.publish(final, observer)
            result = 0
        except KeyboardInterrupt:
            if not isinstance(runtime, _SystemRuntime):
                raise
            monitored = runtime.recover_interrupt(candidate)
            raw = runtime.raw_observations_for(candidate, monitored)
            if raw.get("source_mutated") or raw.get("invalid_provenance"):
                raise InvalidEvidence("INVALID_EVIDENCE: source or provenance authority was lost")
            record = _settled_record(tuple((*runtime._preflight, *monitored)), raw)
            runtime.terminal_verify(candidate, record)
            after = runtime.snapshot("after")
            if before != after:
                raise InvalidEvidence("INVALID_EVIDENCE: source changed during interrupted observation")
            runtime.verify_chain(before, after)
            terminal_facts = runtime.raw_observations_for(candidate, monitored)
            record = _settled_record(tuple((*runtime._preflight, *monitored)), terminal_facts)
            final, observer = runtime.build_receipts(candidate, before, after, record)
            runtime.publish(final, observer)
            result = 0
    except BootstrapError:
        raise
    except KeyboardInterrupt:
        if isinstance(runtime, _SystemRuntime):
            try:
                recoverable_id = runtime.recoverable_bound_container_id()
                if recoverable_id is None:
                    result = 130
                else:
                    runtime.stop_bound_once(recoverable_id)
                    result = 75
            except (KeyboardInterrupt, InvalidEvidence, DockerReadError, OSError, ValueError, TypeError):
                result = 75
        else:
            result = 130
    except (InvalidEvidence, DockerReadError, SourceLeaseError, SourceMutationError, contracts.ReceiptContractError, publication.PublicationError, root_handoff.RootHandoffError, resources.ResourcePolicyError, OSError, ValueError, KeyError, TypeError, UnicodeError):
        result = 75
    finally:
        if leased:
            try:
                runtime.release_lease()
            except Exception:
                result = 75
    return result


if __name__ == "__main__":
    raise SystemExit(main())
