"""Fail-closed, read-only Docker evidence adapter for the backend-unit observer.

The adapter deliberately exposes a very small command surface.  It never
constructs Compose commands and it accepts dynamic values only after checking
their frozen types.  A caller must bind the one candidate container before a
stop can be issued.
"""

from __future__ import annotations

import datetime as _datetime
import json
import re
import subprocess
import threading
import time
from collections.abc import Callable, Iterable, Mapping
from typing import Final, Protocol, runtime_checkable


class DockerReadError(RuntimeError):
    """Raised when Docker evidence cannot be accepted as authoritative."""


__all__ = ("DockerReadAdapter", "DockerReadError")


_PROJECT_RE: Final = re.compile(r"[a-z0-9][a-z0-9_-]{0,62}\Z")
_CONTAINER_ID_RE: Final = re.compile(r"[0-9a-f]{64}\Z")
_IMAGE_ID_RE: Final = re.compile(r"sha256:[0-9a-f]{64}\Z")
_RFC3339_NANO_RE: Final = re.compile(
    r"[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{9}Z\Z"
)

_TIMEOUT_SECONDS: Final = 10
_ONE_SHOT_OUTPUT_CAP_BYTES: Final = 1024 * 1024
_EVENT_RECORDS_MAX: Final = 16
_EVENT_BYTES_MAX: Final = 64 * 1024
_EVENT_ACTIONS: Final = frozenset({"create", "start", "die", "destroy"})
_SERVICE: Final = "backend-unit"
_FIXED_IMAGES: Final = (
    ("python_contract", "worktree-master-python-contract"),
    ("backend_unit", "worktree-master-backend-unit"),
)


class _StrictJsonError(ValueError):
    """Internal sentinel for a JSON representation that is not canonical."""


@runtime_checkable
class _EventProcess(Protocol):
    def __iter__(self) -> Iterable[bytes]: ...

    def terminate(self) -> None: ...

    def reap(self) -> None: ...


Runner = Callable[..., object]


class _SubprocessEventProcess:
    """Small process wrapper whose explicit cleanup is safe after early refusal."""

    def __init__(self, process: subprocess.Popen[bytes]) -> None:
        self._process = process

    def __iter__(self) -> Iterable[bytes]:
        if self._process.stdout is None:
            raise DockerReadError("Docker event stream has no stdout")
        return iter(self._process.stdout)

    def terminate(self) -> None:
        if self._process.poll() is None:
            self._process.terminate()

    def reap(self) -> None:
        try:
            self._process.wait(timeout=_TIMEOUT_SECONDS)
        except subprocess.TimeoutExpired:
            self.terminate()
            self._process.wait(timeout=_TIMEOUT_SECONDS)


def _terminate_and_wait(process: subprocess.Popen[bytes]) -> None:
    if process.poll() is None:
        process.terminate()
    try:
        process.wait(timeout=_TIMEOUT_SECONDS)
    except subprocess.TimeoutExpired:
        process.kill()
        process.wait(timeout=_TIMEOUT_SECONDS)


def _bounded_subprocess_output(
    argv: tuple[str, ...], *, timeout_seconds: int, output_cap_bytes: int
) -> bytes:
    """Read at most one cap plus one byte and fail before retaining more."""

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
    reader_failure: list[Exception] = []
    overflow = threading.Event()

    def read_stdout() -> None:
        try:
            if process.stdout is None:
                raise DockerReadError("Docker command has no stdout")
            while True:
                remaining_plus_probe = output_cap_bytes - len(payload) + 1
                chunk = process.stdout.read(remaining_plus_probe)
                if not chunk:
                    return
                if len(chunk) > output_cap_bytes - len(payload):
                    overflow.set()
                    return
                payload.extend(chunk)
        except Exception as error:  # normalize at the adapter boundary
            reader_failure.append(error)

    reader = threading.Thread(target=read_stdout, daemon=True)
    deadline = time.monotonic() + timeout_seconds
    reader.start()
    reader.join(timeout_seconds)
    try:
        if reader.is_alive():
            _terminate_and_wait(process)
            reader.join(timeout_seconds)
            raise DockerReadError("Docker command exceeded the fixed timeout")
        if reader_failure:
            raise DockerReadError("Docker command output could not be read") from reader_failure[0]
        if overflow.is_set():
            _terminate_and_wait(process)
            raise DockerReadError("Docker command exceeded the fixed output cap")
        remaining = deadline - time.monotonic()
        if remaining <= 0:
            _terminate_and_wait(process)
            raise DockerReadError("Docker command exceeded the fixed timeout")
        try:
            return_code = process.wait(timeout=remaining)
        except subprocess.TimeoutExpired as error:
            _terminate_and_wait(process)
            raise DockerReadError("Docker command exceeded the fixed timeout") from error
        if return_code != 0:
            raise DockerReadError("Docker command exited unsuccessfully")
        return bytes(payload)
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


class DockerReadAdapter:
    """The frozen Docker read surface used by the backend-unit observer."""

    def __init__(self, project: str, *, runner: Runner | None = None) -> None:
        if type(project) is not str or _PROJECT_RE.fullmatch(project) is None:
            raise ValueError("compose project is not canonical")
        if runner is not None and not callable(runner):
            raise TypeError("runner must be callable")
        self._project = project
        self._runner: Runner = _default_runner if runner is None else runner
        self._bound_container_id: str | None = None
        self._stop_used = False

    def events(self, since: str) -> tuple[dict[str, object], ...]:
        _require_rfc3339_nano(since)
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
        try:
            return self._read_events(stream)
        finally:
            self._close_event_stream(stream)

    def inspect(self, container_id: str) -> dict[str, object]:
        canonical_id = self._require_bound_container_id(container_id)
        document = _decode_json_mapping(
            self._one_shot(
                (
                    "docker",
                    "inspect",
                    "--type",
                    "container",
                    "--format",
                    "{{json .}}",
                    canonical_id,
                )
            )
        )
        if document.get("Id") != canonical_id:
            raise DockerReadError("Docker inspect returned a foreign container")
        image_id = document.get("Image")
        if type(image_id) is not str or _IMAGE_ID_RE.fullmatch(image_id) is None:
            raise DockerReadError("Docker inspect has no canonical image ID")
        self._validate_inspected_container_locality(document, canonical_id)
        return document

    def ps(self) -> tuple[dict[str, object], ...]:
        payload = self._one_shot(
            ("docker", "ps", "--no-trunc", "--format", "{{json .}}")
        )
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

    def bind_candidate(self, container_id: str) -> None:
        canonical_id = _require_container_id(container_id)
        if self._bound_container_id is None:
            self._bound_container_id = canonical_id
            return
        if self._bound_container_id != canonical_id:
            raise DockerReadError("a different container is already bound")

    def stop_bound_once(self, container_id: str) -> dict[str, str]:
        canonical_id = _require_container_id(container_id)
        if self._bound_container_id != canonical_id:
            raise DockerReadError("only the bound container may be stopped")
        if self._stop_used:
            raise DockerReadError("the guarded Docker stop was already used")
        # A runner failure may happen after Docker accepted the command.  Burn
        # the single-use capability before invoking it so a retry cannot stop
        # a container twice under an ambiguous outcome.
        self._stop_used = True
        output = self._one_shot(
            ("docker", "stop", "--timeout", "5", canonical_id)
        )
        if not _decode_utf8(output).strip():
            raise DockerReadError("Docker stop returned no confirmation")
        return {"result": "stopped"}

    def _invoke(self, argv: tuple[str, ...], *, stream: bool) -> object:
        try:
            return self._runner(
                argv,
                shell=False,
                timeout_seconds=_TIMEOUT_SECONDS,
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

    def _read_events(self, stream: object) -> tuple[dict[str, object], ...]:
        if not isinstance(stream, _EventProcess):
            raise DockerReadError("Docker runner returned a malformed event stream")
        records: list[dict[str, object]] = []
        bytes_seen = 0
        try:
            iterator = iter(stream)
            for line in iterator:
                if type(line) is not bytes:
                    raise DockerReadError("Docker event stream contains a non-byte record")
                bytes_seen += len(line)
                if bytes_seen > _EVENT_BYTES_MAX:
                    raise DockerReadError("Docker event stream exceeded its byte cap")
                if len(records) >= _EVENT_RECORDS_MAX:
                    raise DockerReadError("Docker event stream exceeded its record cap")
                record = _decode_json_mapping(line)
                self._validate_event_locality(record)
                records.append(record)
        except DockerReadError:
            raise
        except Exception as error:
            raise DockerReadError("Docker event stream could not be read") from error
        return tuple(records)

    def _close_event_stream(self, stream: object) -> None:
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
            raise DockerReadError(
                "Docker event stream could not be terminated and reaped"
            ) from failures[0]

    def _require_bound_container_id(self, container_id: str) -> str:
        canonical_id = _require_container_id(container_id)
        if self._bound_container_id != canonical_id:
            raise DockerReadError("only the uniquely bound container may be read")
        return canonical_id

    def _validate_event_locality(self, record: Mapping[str, object]) -> None:
        container_id = record.get("id")
        if type(container_id) is not str or _CONTAINER_ID_RE.fullmatch(container_id) is None:
            raise DockerReadError("Docker event has a noncanonical container ID")
        if record.get("status") not in _EVENT_ACTIONS:
            raise DockerReadError("Docker event has an unexpected action")
        time_nano = record.get("timeNano")
        if type(time_nano) is not int or time_nano < 0:
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
        if type(hostname) is not str or not re.fullmatch(r"[0-9a-f]{12,64}", hostname):
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


def _require_rfc3339_nano(value: str) -> None:
    if type(value) is not str or _RFC3339_NANO_RE.fullmatch(value) is None:
        raise DockerReadError("Docker event start must be canonical RFC3339Nano")
    try:
        _datetime.datetime.strptime(value[:19], "%Y-%m-%dT%H:%M:%S")
    except ValueError as error:
        raise DockerReadError("Docker event start has an invalid calendar timestamp") from error


def _require_container_id(value: str) -> str:
    if type(value) is not str or _CONTAINER_ID_RE.fullmatch(value) is None:
        raise DockerReadError("container ID must be a full lowercase engine ID")
    return value


def _decode_utf8(payload: bytes) -> str:
    try:
        return payload.decode("utf-8")
    except UnicodeDecodeError as error:
        raise DockerReadError("Docker output is not valid UTF-8") from error


def _decode_json_mapping(payload: bytes) -> dict[str, object]:
    document = _strict_json_loads(_decode_utf8(payload))
    return _require_json_mapping(document)


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
