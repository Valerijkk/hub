"""Fail-closed Docker evidence adapter for one bound backend-unit container.

This is observer infrastructure, never a Docker/Compose launcher.  It binds a
single candidate before exposing either inspection, statistics, or event
evidence, and invalid evidence burns that authority.
"""

from __future__ import annotations

import calendar
import datetime as _datetime
import json
import os
import re
import select
import subprocess
import time
from collections.abc import Callable, Mapping
from typing import Final, Protocol, runtime_checkable


class DockerReadError(RuntimeError):
    """Docker evidence was unavailable or did not meet its read contract."""


class InvalidEvidence(DockerReadError):
    """Fail-closed observer evidence failure; callers map it to exit status 75."""

    exit_status: Final[int] = 75
    exit_code: Final[int] = 75


__all__ = (
    "BoundEventSession",
    "DockerReadAdapter",
    "DockerReadError",
    "InvalidEvidence",
)


_PROJECT_RE: Final = re.compile(r"[a-z0-9][a-z0-9_-]{0,62}\Z")
_CONTAINER_ID_RE: Final = re.compile(r"[0-9a-f]{64}\Z")
_HOSTNAME_RE: Final = re.compile(r"[0-9a-f]{12,64}\Z")
_IMAGE_ID_RE: Final = re.compile(r"sha256:[0-9a-f]{64}\Z")
_RFC3339_NANO_RE: Final = re.compile(
    r"[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{9}Z\Z"
)

_ONE_SHOT_TIMEOUT_SECONDS: Final = 10
_ONE_SHOT_OUTPUT_CAP_BYTES: Final = 1024 * 1024
_BIND_POLL_SECONDS: Final = 5
_BIND_WINDOW_SECONDS: Final = 120
_EVENT_POLL_SECONDS: Final = 5
_DIE_GRACE_SECONDS: Final = 30
_EVENT_RECORDS_MAX: Final = 16
_EVENT_BYTES_MAX: Final = 64 * 1024
_EVENT_ACTIONS: Final = frozenset({"create", "start", "die", "destroy"})
_SERVICE: Final = "backend-unit"
_SESSION_FACTORY_TOKEN: Final = object()
_FIXED_IMAGES: Final = (
    ("python_contract", "worktree-master-python-contract"),
    ("backend_unit", "worktree-master-backend-unit"),
)
_PS_ARGV: Final = ("docker", "ps", "--no-trunc", "--format", "{{json .}}")


class _StrictJsonError(ValueError):
    """Internal sentinel for duplicate, trailing, or non-finite JSON."""


@runtime_checkable
class _EventProcess(Protocol):
    def poll_line(self, timeout_seconds: int) -> bytes | None: ...

    def terminate(self) -> None: ...

    def reap(self) -> None: ...


Runner = Callable[..., object]


class _SubprocessEventProcess:
    """Incremental, no-thread event-stream wrapper with a bounded line buffer."""

    def __init__(self, process: subprocess.Popen[bytes]) -> None:
        self._process = process
        self._pending = bytearray()

    def poll_line(self, timeout_seconds: int) -> bytes | None:
        if type(timeout_seconds) is not int or timeout_seconds != _EVENT_POLL_SECONDS:
            raise DockerReadError("Docker event poll timeout is not frozen")
        if self._process.stdout is None:
            raise DockerReadError("Docker event stream has no stdout")
        deadline = time.monotonic() + timeout_seconds
        while True:
            newline = self._pending.find(b"\n")
            if newline >= 0:
                line = bytes(self._pending[: newline + 1])
                del self._pending[: newline + 1]
                return line
            if len(self._pending) > _EVENT_BYTES_MAX:
                raise DockerReadError("Docker event stream exceeded its byte cap")
            remaining = deadline - time.monotonic()
            if remaining <= 0:
                return None
            chunk = _read_pipe_chunk(
                self._process,
                min(remaining, _EVENT_POLL_SECONDS),
                _EVENT_BYTES_MAX - len(self._pending) + 1,
            )
            if chunk is None:
                return None
            if chunk == b"":
                raise DockerReadError("Docker event stream ended before closure")
            self._pending.extend(chunk)

    def terminate(self) -> None:
        if self._process.poll() is None:
            self._process.terminate()

    def reap(self) -> None:
        try:
            self._process.wait(timeout=_ONE_SHOT_TIMEOUT_SECONDS)
        except subprocess.TimeoutExpired:
            self.terminate()
            self._process.wait(timeout=_ONE_SHOT_TIMEOUT_SECONDS)


def _read_pipe_chunk(
    process: subprocess.Popen[bytes], timeout_seconds: float, max_bytes: int
) -> bytes | None:
    """Read one available pipe chunk without a reader thread or unbounded buffer."""

    if max_bytes < 1 or process.stdout is None:
        raise DockerReadError("Docker process pipe is unavailable")
    descriptor = process.stdout.fileno()
    if os.name != "nt":
        try:
            readable, _, _ = select.select((descriptor,), (), (), timeout_seconds)
        except Exception as error:
            raise DockerReadError("Docker process pipe could not be polled") from error
        if not readable:
            return None
        try:
            return os.read(descriptor, min(65_536, max_bytes))
        except Exception as error:
            raise DockerReadError("Docker process pipe could not be read") from error

    try:
        import msvcrt
        import ctypes

        handle = msvcrt.get_osfhandle(descriptor)
        deadline = time.monotonic() + timeout_seconds
        while True:
            available = ctypes.c_uint32()
            peek = ctypes.windll.kernel32.PeekNamedPipe
            peek.argtypes = (
                ctypes.c_void_p,
                ctypes.c_void_p,
                ctypes.c_uint32,
                ctypes.c_void_p,
                ctypes.POINTER(ctypes.c_uint32),
                ctypes.c_void_p,
            )
            peek.restype = ctypes.c_int
            if not peek(
                ctypes.c_void_p(handle),
                None,
                0,
                None,
                ctypes.byref(available),
                None,
            ):
                if process.poll() is not None:
                    return b""
                raise OSError(ctypes.get_last_error(), "PeekNamedPipe failed")
            if available.value:
                return os.read(descriptor, min(int(available.value), 65_536, max_bytes))
            remaining = deadline - time.monotonic()
            if remaining <= 0:
                return None
            if process.poll() is not None:
                return b""
            time.sleep(min(0.05, remaining))
    except DockerReadError:
        raise
    except Exception as error:
        raise DockerReadError("Docker process pipe could not be polled") from error


def _terminate_and_wait(process: subprocess.Popen[bytes]) -> None:
    if process.poll() is None:
        process.terminate()
    try:
        process.wait(timeout=_ONE_SHOT_TIMEOUT_SECONDS)
    except subprocess.TimeoutExpired:
        process.kill()
        process.wait(timeout=_ONE_SHOT_TIMEOUT_SECONDS)


def _bounded_subprocess_output(
    argv: tuple[str, ...], *, timeout_seconds: int, output_cap_bytes: int
) -> bytes:
    """Collect a one-shot result under fixed time and byte caps without threads."""

    try:
        process = subprocess.Popen(
            argv,
            shell=False,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            bufsize=0,
        )
    except Exception as error:
        raise DockerReadError("Docker command could not be started") from error
    payload = bytearray()
    deadline = time.monotonic() + timeout_seconds
    try:
        while True:
            remaining = deadline - time.monotonic()
            if remaining <= 0:
                raise DockerReadError("Docker command exceeded the fixed timeout")
            chunk = _read_pipe_chunk(
                process,
                remaining,
                output_cap_bytes - len(payload) + 1,
            )
            if chunk is None:
                if process.poll() is None:
                    raise DockerReadError("Docker command exceeded the fixed timeout")
                continue
            if chunk == b"":
                break
            if len(chunk) > output_cap_bytes - len(payload):
                raise DockerReadError("Docker command exceeded the fixed output cap")
            payload.extend(chunk)
        remaining = deadline - time.monotonic()
        if remaining <= 0:
            raise DockerReadError("Docker command exceeded the fixed timeout")
        if process.wait(timeout=remaining) != 0:
            raise DockerReadError("Docker command exited unsuccessfully")
        return bytes(payload)
    except DockerReadError:
        raise
    except Exception as error:
        raise DockerReadError("Docker command output could not be read") from error
    finally:
        if process.poll() is None:
            _terminate_and_wait(process)


def _default_runner(
    argv: tuple[str, ...],
    *,
    shell: bool,
    timeout_seconds: int,
    output_cap_bytes: int,
    stream: bool,
) -> object:
    if shell is not False:
        raise DockerReadError("Docker commands must use shell=False")
    if stream:
        try:
            process = subprocess.Popen(
                argv,
                shell=False,
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                bufsize=0,
            )
        except Exception as error:
            raise DockerReadError("Docker event stream could not be started") from error
        return _SubprocessEventProcess(process)
    return _bounded_subprocess_output(
        argv,
        timeout_seconds=timeout_seconds,
        output_cap_bytes=output_cap_bytes,
    )


class BoundEventSession:
    """Incremental local event evidence for the adapter's already bound ID."""

    def __init__(
        self,
        adapter: DockerReadAdapter,
        stream: _EventProcess,
        since_nanos: int,
        *,
        _token: object,
        _capability: object,
    ) -> None:
        if (
            _token is not _SESSION_FACTORY_TOKEN
            or not isinstance(adapter, DockerReadAdapter)
            or not adapter._accept_session_issuance(_capability, stream, since_nanos)
        ):
            raise DockerReadError("Bound event sessions may only be issued by their adapter")
        self._adapter = adapter
        self._stream = stream
        self._since_nanos = since_nanos
        self._records = 0
        self._bytes = 0
        self._die_deadline: float | None = None
        self._closed = False

    def poll(self) -> dict[str, object] | None:
        """Poll exactly once for a line, using the frozen five-second timeout."""

        if self._closed:
            raise DockerReadError("Bound Docker event session is closed")
        if self._die_deadline is not None and time.monotonic() >= self._die_deadline:
            self.close()
            return None
        try:
            line = self._stream.poll_line(_EVENT_POLL_SECONDS)
        except Exception as error:
            self._fail("Docker event stream could not be polled", error)
        if line is None:
            if self._die_deadline is not None and time.monotonic() >= self._die_deadline:
                self.close()
            return None
        if type(line) is not bytes or not line:
            self._fail("Docker event stream ended or returned a malformed record")
        self._bytes += len(line)
        if self._bytes > _EVENT_BYTES_MAX or self._records >= _EVENT_RECORDS_MAX:
            self._fail("Docker event stream exceeded its fixed evidence cap")
        try:
            record = _decode_json_mapping(line)
            self._adapter._validate_bound_event(record, self._since_nanos)
        except Exception as error:
            self._fail("Docker event stream contains foreign or malformed evidence", error)
        self._records += 1
        if record["status"] == "die":
            self._die_deadline = time.monotonic() + _DIE_GRACE_SECONDS
        elif record["status"] == "destroy":
            self.close()
        return record

    def close(self) -> None:
        if self._closed:
            return
        self._closed = True
        try:
            self._adapter._close_event_stream(self._stream)
        except Exception as error:
            self._adapter._invalidate()
            raise InvalidEvidence("INVALID_EVIDENCE: event stream reap is uncertain") from error

    def _fail(self, message: str, cause: Exception | None = None) -> None:
        self._adapter._invalidate()
        try:
            self.close()
        except InvalidEvidence as cleanup_error:
            raise cleanup_error from cause
        if cause is None:
            raise InvalidEvidence(f"INVALID_EVIDENCE: {message}")
        raise InvalidEvidence(f"INVALID_EVIDENCE: {message}") from cause


class DockerReadAdapter:
    """Frozen Docker read surface; only one complete candidate may be bound."""

    def __init__(self, project: str, *, runner: Runner | None = None) -> None:
        if type(project) is not str or _PROJECT_RE.fullmatch(project) is None:
            raise ValueError("compose project is not canonical")
        if runner is not None and not callable(runner):
            raise TypeError("runner must be callable")
        self._project = project
        self._runner: Runner = _default_runner if runner is None else runner
        self._bound_container_id: str | None = None
        self._bound_backend_image_id: str | None = None
        self._invalid = False
        self._session_issued = False
        self._session_capability: object | None = None
        self._stop_used = False

    def bind_unique_candidate_since(
        self, since: str, fixed_backend_image_id: str
    ) -> str:
        """Poll exact ``docker ps`` and bind the one inspected complete candidate."""

        since_nanos = _rfc3339_nanos(since)
        if type(fixed_backend_image_id) is not str or _IMAGE_ID_RE.fullmatch(
            fixed_backend_image_id
        ) is None:
            raise DockerReadError("fixed backend image ID is not canonical")
        if self._invalid or self._bound_container_id is not None:
            raise DockerReadError("candidate binding is no longer available")
        deadline = time.monotonic() + _BIND_WINDOW_SECONDS
        while True:
            self._require_bind_time(deadline)
            try:
                candidates = self._local_ps_candidates(fixed_backend_image_id)
            except Exception as error:
                self._candidate_fail("docker ps evidence is malformed", error)
            self._require_bind_time(deadline)
            if len(candidates) > 1:
                self._candidate_fail("multiple local backend-unit candidates exist")
            if len(candidates) == 1:
                container_id = candidates[0]
                self._require_bind_time(deadline)
                try:
                    document = self._inspect_provisional(container_id)
                    self._require_bind_time(deadline)
                    self._validate_candidate_inspection(
                        document,
                        container_id,
                        since_nanos,
                        fixed_backend_image_id,
                    )
                except Exception as error:
                    self._candidate_fail("candidate is foreign or incomplete", error)
                self._require_bind_time(deadline)
                self._bound_container_id = container_id
                self._bound_backend_image_id = fixed_backend_image_id
                self._require_bind_time(deadline)
                return container_id
            remaining = deadline - time.monotonic()
            if remaining <= 0:
                self._candidate_fail("unique backend-unit candidate did not appear")
            try:
                time.sleep(min(_BIND_POLL_SECONDS, remaining))
            except Exception as error:
                self._candidate_fail("candidate polling could not continue", error)

    def open_bound_events(self, since: str) -> BoundEventSession:
        """Open the frozen event stream only after a successful unique bind."""

        since_nanos = _rfc3339_nanos(since)
        if self._invalid or self._bound_container_id is None or self._session_issued:
            raise DockerReadError("bound Docker event stream is unavailable")
        try:
            stream = self._invoke(
                (
                    "docker",
                    "events",
                    "--since",
                    since,
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
                ),
                stream=True,
            )
        except Exception as error:
            self._invalidate()
            raise InvalidEvidence("INVALID_EVIDENCE: Docker event stream could not open") from error
        if not isinstance(stream, _EventProcess):
            self._invalidate()
            raise InvalidEvidence("INVALID_EVIDENCE: Docker event stream is malformed")
        capability = object()
        self._session_capability = capability
        try:
            session = BoundEventSession(
                self,
                stream,
                since_nanos,
                _token=_SESSION_FACTORY_TOKEN,
                _capability=capability,
            )
        except Exception as error:
            self._invalidate()
            raise InvalidEvidence("INVALID_EVIDENCE: bound event session could not issue") from error
        finally:
            self._session_capability = None
        self._session_issued = True
        return session

    def inspect(self, container_id: str) -> dict[str, object]:
        canonical_id = self._require_bound_container_id(container_id)
        document = self._inspect_provisional(canonical_id)
        expected_image = self._bound_backend_image_id
        if document.get("Image") != expected_image:
            self._invalidate()
            raise InvalidEvidence("INVALID_EVIDENCE: bound container image changed")
        self._validate_inspected_container_locality(document, canonical_id)
        return document

    def ps(self) -> tuple[dict[str, object], ...]:
        return self._ps_records()

    def stats(self, container_id: str) -> dict[str, object]:
        canonical_id = self._require_bound_container_id(container_id)
        return _decode_json_mapping(
            self._one_shot(
                (
                    "docker",
                    "stats",
                    "--no-stream",
                    "--format",
                    "{{json .}}",
                    canonical_id,
                )
            )
        )

    def inspect_fixed_images(self) -> dict[str, str]:
        image_ids: dict[str, str] = {}
        for name, image in _FIXED_IMAGES:
            document = _decode_json_mapping(
                self._one_shot(
                    ("docker", "image", "inspect", "--format", "{{json .}}", image)
                )
            )
            image_id = document.get("Id")
            if type(image_id) is not str or _IMAGE_ID_RE.fullmatch(image_id) is None:
                raise DockerReadError("fixed Docker image has no canonical image ID")
            image_ids[name] = image_id
        return image_ids

    def stop_bound_once(self, container_id: str) -> dict[str, str]:
        canonical_id = self._require_bound_container_id(container_id)
        if self._stop_used:
            raise DockerReadError("the guarded Docker stop was already used")
        self._stop_used = True
        output = self._one_shot(("docker", "stop", "--timeout", "5", canonical_id))
        if not _decode_utf8(output).strip():
            raise DockerReadError("Docker stop returned no confirmation")
        return {"result": "stopped"}

    def _candidate_fail(self, message: str, cause: Exception | None = None) -> None:
        self._invalidate()
        if cause is None:
            raise InvalidEvidence(f"INVALID_EVIDENCE: {message}")
        raise InvalidEvidence(f"INVALID_EVIDENCE: {message}") from cause

    def _invalidate(self) -> None:
        self._invalid = True
        self._bound_container_id = None
        self._bound_backend_image_id = None
        self._session_capability = None

    def _require_bind_time(self, deadline: float) -> None:
        if time.monotonic() >= deadline:
            self._candidate_fail("unique backend-unit candidate exceeded the fixed bind deadline")

    def _accept_session_issuance(
        self, capability: object, stream: object, since_nanos: object
    ) -> bool:
        return (
            capability is self._session_capability
            and self._session_issued is False
            and self._invalid is False
            and self._bound_container_id is not None
            and isinstance(stream, _EventProcess)
            and type(since_nanos) is int
            and since_nanos >= 0
        )

    def _invoke(self, argv: tuple[str, ...], *, stream: bool) -> object:
        try:
            return self._runner(
                argv,
                shell=False,
                timeout_seconds=_ONE_SHOT_TIMEOUT_SECONDS,
                output_cap_bytes=_ONE_SHOT_OUTPUT_CAP_BYTES,
                stream=stream,
            )
        except DockerReadError:
            raise
        except Exception as error:
            raise DockerReadError("Docker runner failed") from error

    def _one_shot(self, argv: tuple[str, ...]) -> bytes:
        output = self._invoke(argv, stream=False)
        if type(output) is not bytes:
            raise DockerReadError("Docker runner returned a malformed one-shot result")
        if len(output) > _ONE_SHOT_OUTPUT_CAP_BYTES:
            raise DockerReadError("Docker one-shot output exceeds its fixed cap")
        return output

    def _ps_records(self) -> tuple[dict[str, object], ...]:
        payload = self._one_shot(_PS_ARGV)
        text = _decode_utf8(payload)
        stripped = text.strip()
        if not stripped:
            return ()
        try:
            whole = _strict_json_loads(stripped)
        except DockerReadError:
            whole = None
        if whole == []:
            return ()
        if isinstance(whole, Mapping):
            return (_require_json_mapping(whole),)
        if whole is not None:
            raise DockerReadError("Docker ps did not return line-delimited objects")
        records: list[dict[str, object]] = []
        for line in text.splitlines():
            if not line.strip():
                raise DockerReadError("Docker ps has a blank record")
            records.append(_decode_json_mapping(line.encode("utf-8")))
        return tuple(records)

    def _local_ps_candidates(self, fixed_backend_image_id: str) -> tuple[str, ...]:
        candidates: list[str] = []
        for record in self._ps_records():
            labels = _ps_labels(record.get("Labels"))
            project = labels.get("com.docker.compose.project")
            service = labels.get("com.docker.compose.service")
            if project != self._project or service != _SERVICE:
                continue
            container_id = record.get("ID")
            image = record.get("Image")
            if (
                type(container_id) is not str
                or _CONTAINER_ID_RE.fullmatch(container_id) is None
                or image != fixed_backend_image_id
            ):
                raise DockerReadError("local docker ps candidate is malformed or foreign")
            candidates.append(container_id)
        if len(candidates) != len(set(candidates)):
            raise DockerReadError("docker ps repeats a local candidate")
        return tuple(candidates)

    def _inspect_provisional(self, container_id: str) -> dict[str, object]:
        document = _decode_json_mapping(
            self._one_shot(
                (
                    "docker",
                    "inspect",
                    "--type",
                    "container",
                    "--format",
                    "{{json .}}",
                    container_id,
                )
            )
        )
        if document.get("Id") != container_id:
            raise DockerReadError("Docker inspect returned a foreign container")
        return document

    def _validate_candidate_inspection(
        self,
        document: Mapping[str, object],
        container_id: str,
        since_nanos: int,
        fixed_backend_image_id: str,
    ) -> None:
        if document.get("Image") != fixed_backend_image_id:
            raise DockerReadError("candidate image does not equal the fixed image")
        self._validate_inspected_container_locality(document, container_id)
        state = document.get("State")
        if not isinstance(state, Mapping):
            raise DockerReadError("candidate has no state")
        started_at = state.get("StartedAt")
        started_nanos = _rfc3339_nanos(started_at)
        if not since_nanos <= started_nanos < since_nanos + _BIND_WINDOW_SECONDS * 1_000_000_000:
            raise DockerReadError("candidate start is outside the fixed bind window")

    def _validate_bound_event(
        self, record: Mapping[str, object], since_nanos: int
    ) -> None:
        bound = self._bound_container_id
        if self._invalid or bound is None:
            raise DockerReadError("bound event authority is unavailable")
        container_id = record.get("id")
        if container_id != bound or not isinstance(container_id, str) or _CONTAINER_ID_RE.fullmatch(container_id) is None:
            raise DockerReadError("Docker event belongs to a foreign container")
        if record.get("status") not in _EVENT_ACTIONS:
            raise DockerReadError("Docker event has an unexpected action")
        time_nano = record.get("timeNano")
        if type(time_nano) is not int or time_nano < since_nanos:
            raise DockerReadError("Docker event has an invalid timestamp")
        actor = record.get("Actor")
        if not isinstance(actor, Mapping):
            raise DockerReadError("Docker event has no actor")
        attributes = actor.get("Attributes")
        if not isinstance(attributes, Mapping):
            raise DockerReadError("Docker event has no label attributes")
        if (
            attributes.get("com.docker.compose.project") != self._project
            or attributes.get("com.docker.compose.service") != _SERVICE
        ):
            raise DockerReadError("Docker event is not local backend-unit evidence")

    def _validate_inspected_container_locality(
        self, document: Mapping[str, object], container_id: str
    ) -> None:
        config = document.get("Config")
        if not isinstance(config, Mapping):
            raise DockerReadError("Docker inspect has no container config")
        hostname = config.get("Hostname")
        if type(hostname) is not str or _HOSTNAME_RE.fullmatch(hostname) is None:
            raise DockerReadError("Docker inspect has no canonical hostname marker")
        if not container_id.startswith(hostname):
            raise DockerReadError("Docker inspect hostname is not the container prefix")
        labels: list[Mapping[str, object]] = []
        config_labels = document.get("ConfigLabels")
        if config_labels is not None:
            if not isinstance(config_labels, Mapping):
                raise DockerReadError("Docker inspect ConfigLabels are malformed")
            labels.append(config_labels)
        nested_labels = config.get("Labels")
        if nested_labels is not None:
            if not isinstance(nested_labels, Mapping):
                raise DockerReadError("Docker inspect labels are malformed")
            labels.append(nested_labels)
        if not labels:
            raise DockerReadError("Docker inspect has no compose labels")
        for mapping in labels:
            if (
                mapping.get("com.docker.compose.project") != self._project
                or mapping.get("com.docker.compose.service") != _SERVICE
            ):
                raise DockerReadError("Docker inspect is not local backend-unit evidence")

    def _close_event_stream(self, stream: _EventProcess) -> None:
        failures: list[Exception] = []
        for operation_name in ("terminate", "reap"):
            try:
                operation = getattr(stream, operation_name)
                if not callable(operation):
                    raise TypeError(f"event stream has no {operation_name} method")
                operation()
            except Exception as error:
                failures.append(error)
        if failures:
            raise DockerReadError("Docker event stream could not be terminated and reaped") from failures[0]

    def _require_bound_container_id(self, container_id: str) -> str:
        if self._invalid:
            raise DockerReadError("invalid evidence burned bound container authority")
        canonical_id = _require_container_id(container_id)
        if self._bound_container_id != canonical_id:
            raise DockerReadError("only the uniquely bound container may be read")
        return canonical_id


def _rfc3339_nanos(value: object) -> int:
    if type(value) is not str or _RFC3339_NANO_RE.fullmatch(value) is None:
        raise DockerReadError("timestamp must be canonical RFC3339Nano")
    try:
        parsed = _datetime.datetime.strptime(value[:19], "%Y-%m-%dT%H:%M:%S")
        seconds = calendar.timegm(parsed.timetuple())
        return seconds * 1_000_000_000 + int(value[20:29])
    except ValueError as error:
        raise DockerReadError("timestamp has an invalid calendar value") from error


def _require_container_id(value: object) -> str:
    if type(value) is not str or _CONTAINER_ID_RE.fullmatch(value) is None:
        raise DockerReadError("container ID must be a full lowercase engine ID")
    return value


def _ps_labels(value: object) -> dict[str, str]:
    if isinstance(value, Mapping):
        if not all(type(key) is str and type(item) is str for key, item in value.items()):
            raise DockerReadError("Docker ps labels are malformed")
        return dict(value)
    if type(value) is not str:
        raise DockerReadError("Docker ps labels are malformed")
    labels: dict[str, str] = {}
    if not value:
        return labels
    for entry in value.split(","):
        if entry.count("=") != 1:
            raise DockerReadError("Docker ps labels are malformed")
        key, item = entry.split("=", 1)
        if not key or not item or key in labels:
            raise DockerReadError("Docker ps labels are malformed")
        labels[key] = item
    return labels


def _decode_utf8(payload: bytes) -> str:
    try:
        return payload.decode("utf-8")
    except UnicodeDecodeError as error:
        raise DockerReadError("Docker output is not valid UTF-8") from error


def _decode_json_mapping(payload: bytes) -> dict[str, object]:
    return _require_json_mapping(_strict_json_loads(_decode_utf8(payload)))


def _strict_json_loads(text: str) -> object:
    try:
        return json.loads(
            text,
            object_pairs_hook=_strict_json_object,
            parse_constant=_reject_json_constant,
        )
    except DockerReadError:
        raise
    except Exception as error:
        raise DockerReadError("Docker output is not canonical JSON") from error


def _strict_json_object(pairs: list[tuple[object, object]]) -> dict[str, object]:
    document: dict[str, object] = {}
    for key, value in pairs:
        if type(key) is not str or key in document:
            raise _StrictJsonError("JSON object has a duplicate or non-string key")
        document[key] = value
    return document


def _reject_json_constant(value: str) -> object:
    raise _StrictJsonError(f"non-finite JSON constant is forbidden: {value}")


def _require_json_mapping(value: object) -> dict[str, object]:
    if not isinstance(value, Mapping):
        raise DockerReadError("Docker JSON record is not an object")
    try:
        return {str(key): item for key, item in value.items()}
    except Exception as error:
        raise DockerReadError("Docker JSON mapping could not be materialized") from error
