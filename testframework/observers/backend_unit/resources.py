"""Fail-closed host-resource sampling and receipt classification.

The observer owns no test process here.  This module has only two jobs: read
one fixed, non-interactive host metric command and reduce already-observed
facts to the single sealed classification language.  An unreadable metric,
an unexpected JSON shape, or an ambiguous fact is a resource-policy failure;
callers must therefore leave the run unsealed.
"""

from __future__ import annotations

import json
import re
import subprocess
import threading
from calendar import timegm
from dataclasses import dataclass
from datetime import datetime, timezone
from typing import Final


class ResourcePolicyError(ValueError):
    """Raised when host-resource evidence is incomplete or unsafe."""


__all__ = (
    "ObserverTiming",
    "ResourcePolicyError",
    "collect_host_sample",
    "preflight",
    "classification_record",
    "classify",
)


_GIB: Final = 1024**3
_PREFLIGHT_RAM_BYTES_MIN: Final = 2 * _GIB
_PREFLIGHT_C_FREE_BYTES_MIN: Final = 250 * _GIB
_DISK_SAMPLE_COUNT: Final = 3
_DISK_DEFER_COUNT: Final = 2
_PERCENT_DISK_TIME_MIN: Final = 70
_DISK_QUEUE_LENGTH_MIN: Final = 2
_RUNTIME_RAM_CONSECUTIVE_SAMPLES: Final = 2
_WHOLE_RUN_DEADLINE_SECONDS: Final = 6 * 60 * 60
_METRIC_TIMEOUT_SECONDS: Final = 10
_METRIC_OUTPUT_BYTES_MAX: Final = 16 * 1024
_METRIC_STDERR_BYTES_MAX: Final = 4 * 1024
_PROCESS_REAP_SECONDS: Final = 1
_PREFLIGHT_SAMPLE_INTERVAL_NANOSECONDS: Final = 2 * 1_000_000_000
_UNTIMED_DERIVED_STOP_TIMESTAMP: Final = (2**63) - 1

_METRIC_KEYS: Final = (
    "free_ram_bytes",
    "free_c_bytes",
    "PercentDiskTime",
    "CurrentDiskQueueLength",
)
_OPTIONAL_SAMPLE_KEYS: Final = frozenset({"timestamp", "phase"})
_CANONICAL_TIMESTAMP_RE: Final = re.compile(
    r"[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{6}Z\Z"
)
_STOP_KINDS: Final = ("interrupt", "qa", "disk", "ram", "deadline")
_STOP_TIE_ORDER: Final = {kind: index for index, kind in enumerate(_STOP_KINDS)}
_FACT_KEYS: Final = frozenset(
    {
        "source_mutated",
        "invalid_provenance",
        "oom",
        "resource",
        "budget",
        "tmpfs",
        "timeout",
        "discovery_failure",
        "stage_failure",
        "explicit_stops",
        "prohibited_backend_unit_overlap",
        "prohibited_mail_calendar_qa_test_build_overlap",
        "buildkit_idle",
        "whole_run_deadline_expired",
    }
)
_BOOLEAN_FACT_KEYS: Final = _FACT_KEYS - frozenset({"explicit_stops", "buildkit_idle"})
_CLASSIFICATIONS: Final = (
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

# This is deliberately one literal argv.  No caller-controlled fragment can
# reach PowerShell, and stdout is a single compressed JSON object.
_POWERSHELL_METRIC_ARGV: Final = (
    "powershell.exe",
    "-NoLogo",
    "-NoProfile",
    "-NonInteractive",
    "-ExecutionPolicy",
    "Bypass",
    "-Command",
    "$ErrorActionPreference='Stop';"
    "$os=@(Get-CimInstance Win32_OperatingSystem);"
    "$disk=@(Get-CimInstance Win32_PerfFormattedData_PerfDisk_PhysicalDisk|"
    "Where-Object {$_.Name -eq '_Total'});"
    "$c=@(Get-CimInstance Win32_LogicalDisk -Filter \"DeviceID='C:'\");"
    "if($os.Count -ne 1 -or $c.Count -ne 1 -or $disk.Count -ne 1){throw 'metric source cardinality'};"
    "if($null -eq $os[0].FreePhysicalMemory -or $null -eq $c[0].FreeSpace -or "
    "$null -eq $disk[0].PercentDiskTime -or $null -eq $disk[0].CurrentDiskQueueLength){throw 'metric counter missing'};"
    "[pscustomobject]@{"
    "free_ram_bytes=[int64]$os[0].FreePhysicalMemory*1KB;"
    "free_c_bytes=[int64]$c[0].FreeSpace;"
    "PercentDiskTime=[int]$disk[0].PercentDiskTime;"
    "CurrentDiskQueueLength=[int]$disk[0].CurrentDiskQueueLength"
    "}|ConvertTo-Json -Compress",
)


@dataclass(frozen=True)
class ObserverTiming:
    """The only observer timing values permitted by the frozen policy."""

    candidate_window_seconds: int = 120
    runtime_sample_interval_seconds: int = 5
    ready_deadline_seconds: int = 30
    destroy_grace_seconds: int = 30
    whole_run_deadline_seconds: int = _WHOLE_RUN_DEADLINE_SECONDS
    deadline_owner: str = "observer"


def _json_pairs_without_duplicates(pairs: list[tuple[str, object]]) -> dict[str, object]:
    result: dict[str, object] = {}
    for key, value in pairs:
        if type(key) is not str or key in result:
            raise ResourcePolicyError("metric JSON has duplicate or invalid keys")
        result[key] = value
    return result


def _bounded_stream_read(stream: object, maximum: int, sink: list[object]) -> None:
    """Retain at most ``maximum + 1`` bytes from one Popen-owned pipe."""

    try:
        if stream is None:
            raise ResourcePolicyError("metric process did not expose a pipe")
        payload = stream.read(maximum + 1)
        if type(payload) is not bytes:
            raise ResourcePolicyError("metric process emitted a non-byte stream")
        sink.append(payload)
    except Exception as error:
        sink.append(error)


def _terminate_kill_and_reap(process: subprocess.Popen[bytes]) -> None:
    """Leave no process behind after any timeout, overflow, or pipe uncertainty."""

    if process.poll() is not None:
        return
    try:
        process.terminate()
        process.wait(timeout=_PROCESS_REAP_SECONDS)
        return
    except (OSError, subprocess.TimeoutExpired):
        pass
    try:
        process.kill()
        process.wait(timeout=_PROCESS_REAP_SECONDS)
    except (OSError, subprocess.TimeoutExpired) as error:
        raise ResourcePolicyError("metric process could not be reaped") from error


def _join_reader(reader: threading.Thread) -> None:
    reader.join(timeout=_PROCESS_REAP_SECONDS)
    if reader.is_alive():
        raise ResourcePolicyError("metric pipe reader could not be reaped")


def _run_metric_command() -> bytes:
    """Execute the sole admitted host metric command with bounded pipe reads."""

    process: subprocess.Popen[bytes] | None = None
    stdout_reader: threading.Thread | None = None
    stderr_reader: threading.Thread | None = None
    stdout_sink: list[object] = []
    stderr_sink: list[object] = []
    try:
        process = subprocess.Popen(
            _POWERSHELL_METRIC_ARGV,
            shell=False,
            bufsize=0,
            stdin=subprocess.DEVNULL,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )
        stdout_reader = threading.Thread(
            target=_bounded_stream_read,
            args=(process.stdout, _METRIC_OUTPUT_BYTES_MAX, stdout_sink),
            daemon=False,
        )
        stderr_reader = threading.Thread(
            target=_bounded_stream_read,
            args=(process.stderr, _METRIC_STDERR_BYTES_MAX, stderr_sink),
            daemon=False,
        )
        stdout_reader.start()
        stderr_reader.start()
        try:
            process.wait(timeout=_METRIC_TIMEOUT_SECONDS)
        except subprocess.TimeoutExpired as error:
            _terminate_kill_and_reap(process)
            raise ResourcePolicyError("host metric command timed out") from error
        _join_reader(stdout_reader)
        _join_reader(stderr_reader)
        if type(process.returncode) is not int or process.returncode != 0:
            raise ResourcePolicyError("host metric command failed")
        if len(stdout_sink) != 1 or len(stderr_sink) != 1:
            raise ResourcePolicyError("host metric pipe result is ambiguous")
        stdout, stderr = stdout_sink[0], stderr_sink[0]
        if type(stdout) is not bytes or type(stderr) is not bytes:
            raise ResourcePolicyError("host metric command returned an invalid stream")
        if not stdout or len(stdout) > _METRIC_OUTPUT_BYTES_MAX:
            raise ResourcePolicyError("host metric stdout exceeded its bound")
        if stderr or len(stderr) > _METRIC_STDERR_BYTES_MAX:
            raise ResourcePolicyError("host metric stderr is unsafe")
        return stdout
    except (OSError, ResourcePolicyError) as error:
        raise ResourcePolicyError("host metric command was not authoritative") from error
    finally:
        if process is not None:
            _terminate_kill_and_reap(process)
            if process.stdout is not None:
                process.stdout.close()
            if process.stderr is not None:
                process.stderr.close()
        if stdout_reader is not None:
            _join_reader(stdout_reader)
        if stderr_reader is not None:
            _join_reader(stderr_reader)


def _validate_sample(sample: object) -> dict[str, object]:
    if type(sample) is not dict:
        raise ResourcePolicyError("resource sample must be a plain mapping")
    keys = frozenset(sample)
    required = frozenset(_METRIC_KEYS)
    if not required <= keys or keys - required - _OPTIONAL_SAMPLE_KEYS:
        raise ResourcePolicyError("resource sample keys are not closed")
    for key in _METRIC_KEYS:
        value = sample.get(key)
        if type(value) is not int or value < 0:
            raise ResourcePolicyError("resource metric is not a non-negative integer")
    if sample["PercentDiskTime"] > 100:
        raise ResourcePolicyError("disk time percentage is out of range")
    if "timestamp" in sample and (
        type(sample["timestamp"]) is not str
        or _CANONICAL_TIMESTAMP_RE.fullmatch(sample["timestamp"]) is None
    ):
        raise ResourcePolicyError("sample timestamp is malformed")
    if "phase" in sample and (
        type(sample["phase"]) is not str
        or sample["phase"] not in {"preflight", "runtime"}
    ):
        raise ResourcePolicyError("sample phase is malformed")
    return sample


def _canonical_utc_timestamp() -> str:
    return datetime.now(timezone.utc).isoformat(timespec="microseconds").replace(
        "+00:00", "Z"
    )


def _timestamp_nanoseconds(sample: dict[str, object]) -> int | None:
    value = sample.get("timestamp")
    if value is None:
        return None
    if type(value) is not str or _CANONICAL_TIMESTAMP_RE.fullmatch(value) is None:
        raise ResourcePolicyError("sample timestamp is malformed")
    try:
        parsed = datetime.strptime(value, "%Y-%m-%dT%H:%M:%S.%fZ")
    except ValueError as error:
        raise ResourcePolicyError("sample timestamp is not a real UTC instant") from error
    return timegm(parsed.timetuple()) * 1_000_000_000 + parsed.microsecond * 1_000


def collect_host_sample() -> dict[str, int | str]:
    """Read and validate one noninteractive Windows host-resource snapshot."""

    raw = _run_metric_command()
    try:
        decoded = json.loads(raw.decode("utf-8", errors="strict"), object_pairs_hook=_json_pairs_without_duplicates)
    except (UnicodeDecodeError, json.JSONDecodeError, ResourcePolicyError) as error:
        raise ResourcePolicyError("host metric output is not closed JSON") from error
    sample = _validate_sample(decoded)
    return {
        **{key: sample[key] for key in _METRIC_KEYS},
        "timestamp": _canonical_utc_timestamp(),
    }  # type: ignore[return-value]


def _disk_pressure(sample: dict[str, object]) -> bool:
    return (
        sample["PercentDiskTime"] >= _PERCENT_DISK_TIME_MIN
        or sample["CurrentDiskQueueLength"] >= _DISK_QUEUE_LENGTH_MIN
    )


def _preflight_samples(samples: object) -> tuple[dict[str, object], dict[str, object], dict[str, object]]:
    if type(samples) is not tuple or len(samples) != _DISK_SAMPLE_COUNT:
        raise ResourcePolicyError("preflight requires exactly three samples")
    first = _validate_sample(samples[0])
    second = _validate_sample(samples[1])
    third = _validate_sample(samples[2])
    return first, second, third


def preflight(samples: tuple[dict[str, int], dict[str, int], dict[str, int]]) -> None:
    """Reject a host that is below a fixed floor or under disk pressure."""

    validated = _preflight_samples(samples)
    if any(sample["free_ram_bytes"] < _PREFLIGHT_RAM_BYTES_MIN for sample in validated):
        raise ResourcePolicyError("free RAM is below the 2 GiB floor")
    if any(sample["free_c_bytes"] < _PREFLIGHT_C_FREE_BYTES_MIN for sample in validated):
        raise ResourcePolicyError("C: free space is below the 250 GiB floor")
    if sum(_disk_pressure(sample) for sample in validated) >= _DISK_DEFER_COUNT:
        raise ResourcePolicyError("disk pressure met the 2-of-3 defer rule")
    timestamps = tuple(_timestamp_nanoseconds(sample) for sample in validated)
    if all(timestamp is not None for timestamp in timestamps):
        first, second, third = timestamps
        if (
            second - first < _PREFLIGHT_SAMPLE_INTERVAL_NANOSECONDS
            or third - second < _PREFLIGHT_SAMPLE_INTERVAL_NANOSECONDS
        ):
            raise ResourcePolicyError("preflight samples were not spaced by two seconds")


def _runtime_samples(samples: object) -> tuple[dict[str, object], ...]:
    if type(samples) is not tuple:
        raise ResourcePolicyError("runtime samples must be an immutable tuple")
    if not samples:
        raise ResourcePolicyError("runtime requires at least one resource sample")
    return tuple(dict(_validate_sample(sample)) for sample in samples)


def _two_consecutive_ram_below_floor(samples: tuple[dict[str, object], ...]) -> tuple[bool, int | None]:
    consecutive = 0
    for index, sample in enumerate(samples):
        if sample["free_ram_bytes"] < _PREFLIGHT_RAM_BYTES_MIN:
            consecutive += 1
            if consecutive == _RUNTIME_RAM_CONSECUTIVE_SAMPLES:
                return True, index
        else:
            consecutive = 0
    return False, None


def _rolling_disk_pressure(samples: tuple[dict[str, object], ...]) -> tuple[bool, int | None]:
    for end in range(_DISK_SAMPLE_COUNT - 1, len(samples)):
        window = samples[end - (_DISK_SAMPLE_COUNT - 1) : end + 1]
        if sum(_disk_pressure(sample) for sample in window) >= _DISK_DEFER_COUNT:
            return True, end
    return False, None


def _facts(policy: object) -> dict[str, object]:
    if type(policy) is not dict:
        raise ResourcePolicyError("classification policy must be a plain mapping")
    permitted_policy_keys = {"classification_facts", "sealed_classifications"}
    if set(policy) - permitted_policy_keys or "classification_facts" not in policy:
        raise ResourcePolicyError("classification policy keys are not closed")
    classifications = policy.get("sealed_classifications", _CLASSIFICATIONS)
    if type(classifications) is not tuple or classifications != _CLASSIFICATIONS:
        raise ResourcePolicyError("classification language is not frozen")
    raw = policy["classification_facts"]
    if type(raw) is not dict or set(raw) - _FACT_KEYS:
        raise ResourcePolicyError("classification facts are not closed")
    for key in _BOOLEAN_FACT_KEYS:
        if key in raw and type(raw[key]) is not bool:
            raise ResourcePolicyError("classification fact must be boolean")
    if "buildkit_idle" in raw and type(raw["buildkit_idle"]) is not bool:
        raise ResourcePolicyError("BuildKit state must be boolean")
    return raw


def _explicit_stops(value: object) -> tuple[tuple[str, int], ...]:
    if value is None:
        return ()
    if type(value) is not tuple:
        raise ResourcePolicyError("explicit stops must be an immutable tuple")
    stops: list[tuple[str, int]] = []
    for stop in value:
        if type(stop) is not tuple or len(stop) != 2:
            raise ResourcePolicyError("explicit stop is malformed")
        kind, timestamp = stop
        if type(kind) is not str or kind not in _STOP_TIE_ORDER:
            raise ResourcePolicyError("explicit stop kind is invalid")
        if type(timestamp) is not int or timestamp < 0:
            raise ResourcePolicyError("explicit stop timestamp is invalid")
        stops.append((kind, timestamp))
    return tuple(stops)


def _select_stop(stops: tuple[tuple[str, int], ...]) -> tuple[str, int] | None:
    if not stops:
        return None
    return min(stops, key=lambda stop: (stop[1], _STOP_TIE_ORDER[stop[0]]))


def _classification_from(
    facts: dict[str, object],
    selected_stop: tuple[str, int] | None,
) -> str:
    if facts.get("source_mutated", False):
        return "SEALED_INVALID_SOURCE_MUTATED"
    if facts.get("invalid_provenance", False):
        return "SEALED_INVALID_PROVENANCE"
    if selected_stop is not None:
        return {
            "interrupt": "SEALED_INTERRUPTED",
            "qa": "SEALED_INVALID_QA_OVERLAP",
            "disk": "SEALED_INFRA_ABORTED_DISK_PRESSURE",
            "ram": "SEALED_INFRA_ABORTED_HOST_RAM_FLOOR",
            "deadline": "SEALED_INFRA_ABORTED_WHOLE_RUN_TIMEOUT",
        }[selected_stop[0]]
    if facts.get("oom", False):
        return "SEALED_INFRA_ABORTED_OOM"
    if facts.get("resource", False) or facts.get("buildkit_idle", True) is False:
        return "SEALED_INFRA_ABORTED_RESOURCE_CONTRACT"
    if facts.get("budget", False):
        return "SEALED_INFRA_ABORTED_REPORT_BUDGET"
    if facts.get("tmpfs", False):
        return "SEALED_INFRA_ABORTED_TMPFS"
    if facts.get("timeout", False):
        return "SEALED_TIMEOUT_UNTRIAGED"
    if facts.get("discovery_failure", False):
        return "SEALED_EXECUTABLE_FAIL_DISCOVERY"
    if facts.get("stage_failure", False):
        return "SEALED_EXECUTABLE_FAIL_STAGE"
    return "SEALED_EXECUTABLE_PASS"


def classification_record(samples: tuple[dict[str, int], ...], policy: dict[str, object]) -> dict[str, object]:
    """Return the sealed class plus every raw sample and predicate decision."""

    validated_samples = _runtime_samples(samples)
    facts = _facts(policy)
    ram_low, ram_index = _two_consecutive_ram_below_floor(validated_samples)
    disk_pressure, disk_index = _rolling_disk_pressure(validated_samples)
    preflight_pressure = (
        len(validated_samples) >= _DISK_SAMPLE_COUNT
        and sum(_disk_pressure(sample) for sample in validated_samples[:_DISK_SAMPLE_COUNT]) >= _DISK_DEFER_COUNT
    )
    predicates = {
        "preflight_ram_floor_met": all(sample["free_ram_bytes"] >= _PREFLIGHT_RAM_BYTES_MIN for sample in validated_samples),
        "preflight_c_free_floor_met": all(sample["free_c_bytes"] >= _PREFLIGHT_C_FREE_BYTES_MIN for sample in validated_samples),
        "preflight_disk_pressure_2_of_3": preflight_pressure,
        "two_consecutive_ram_below_2_gib": ram_low,
        "rolling_disk_pressure_2_of_3": disk_pressure,
        "prohibited_backend_unit_overlap": facts.get("prohibited_backend_unit_overlap", False),
        "prohibited_mail_calendar_qa_test_build_overlap": facts.get("prohibited_mail_calendar_qa_test_build_overlap", False),
        "buildkit_idle": facts.get("buildkit_idle", True),
        "whole_run_deadline_expired": facts.get("whole_run_deadline_expired", False),
    }
    raw_stops = _explicit_stops(facts.get("explicit_stops"))
    derived_stops: list[tuple[str, int]] = []
    if predicates["prohibited_backend_unit_overlap"] or predicates["prohibited_mail_calendar_qa_test_build_overlap"]:
        derived_stops.append(("qa", _UNTIMED_DERIVED_STOP_TIMESTAMP))
    if disk_pressure and disk_index is not None:
        disk_timestamp = _timestamp_nanoseconds(validated_samples[disk_index])
        derived_stops.append(
            (
                "disk",
                disk_timestamp
                if disk_timestamp is not None
                else _UNTIMED_DERIVED_STOP_TIMESTAMP,
            )
        )
    if ram_low and ram_index is not None:
        ram_timestamp = _timestamp_nanoseconds(validated_samples[ram_index])
        derived_stops.append(
            (
                "ram",
                ram_timestamp
                if ram_timestamp is not None
                else _UNTIMED_DERIVED_STOP_TIMESTAMP,
            )
        )
    if predicates["whole_run_deadline_expired"]:
        derived_stops.append(("deadline", _UNTIMED_DERIVED_STOP_TIMESTAMP))
    selected_stop = _select_stop((*raw_stops, *derived_stops))
    classification = _classification_from(facts, selected_stop)
    immutable_outcome = "TIMEOUT_UNTRIAGED" if classification == "SEALED_TIMEOUT_UNTRIAGED" else None
    return {
        "classification": classification,
        "immutable_outcome": immutable_outcome,
        "raw_samples": tuple(dict(sample) for sample in validated_samples),
        "predicates": predicates,
        "explicit_stops": raw_stops,
        "selected_stop": selected_stop,
    }


def classify(samples: tuple[dict[str, int], ...], policy: dict[str, object]) -> str:
    """Return the one closed classification selected from authoritative facts."""

    result = classification_record(samples, policy)["classification"]
    if type(result) is not str or result not in _CLASSIFICATIONS:
        raise ResourcePolicyError("classification result is not sealed")
    return result
