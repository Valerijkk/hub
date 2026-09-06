from __future__ import annotations

import os
from pathlib import Path
import select
import signal
import subprocess
import time


_OBSERVATION_SECONDS = 1.5
_SUBPROCESS_SECONDS = 5.0
_PARENT_SCRIPT = """\
release_fifo="$1"
wait_control_fifo="$2"
status_file="$3"

exec 3<> "$release_fifo"
exec 4<> "$wait_control_fifo"

/bin/sh -c 'IFS= read -r _ < "$1"; exit 17' backend-unit-owned-child "$release_fifo" &
owned_pid=$!
printf '%s %s\n' "$$" "$owned_pid"

IFS= read -r _ <&4
wait "$owned_pid"
owned_status=$?
printf '%s\n' "$owned_status" > "$status_file"
"""


def _read_proc_identity(pid: int) -> tuple[str, str, str] | None:
    try:
        stat = Path(f"/proc/{pid}/stat").read_text(encoding="ascii")
    except FileNotFoundError:
        return None

    _prefix, separator, fields_text = stat.rpartition(")")
    if separator != ")":
        raise AssertionError(f"invalid proc stat for owned pid {pid}")
    fields = fields_text.split()
    if len(fields) < 20:
        raise AssertionError(f"incomplete proc stat for owned pid {pid}")
    return fields[0], fields[3], fields[19]


def _read_readiness_record(descriptor: int, deadline: float) -> tuple[int, int]:
    remaining = deadline - time.monotonic()
    if remaining <= 0:
        raise AssertionError("owned child readiness deadline elapsed")
    poller = select.poll()
    poller.register(descriptor, select.POLLIN | select.POLLERR | select.POLLHUP)
    if not any(
        events & select.POLLIN
        for _descriptor, events in poller.poll(int(remaining * 1000))
    ):
        raise AssertionError("owned child readiness was not readable before timeout")
    record = os.read(descriptor, 128)
    fields = record[:-1].split(b" ")
    if (
        not record.endswith(b"\n")
        or record.count(b"\n") != 1
        or len(record) == 128
        or len(fields) != 2
    ):
        raise AssertionError("owned child readiness record was malformed")
    try:
        parent_pid, child_pid = (int(field) for field in fields)
    except ValueError as error:
        raise AssertionError("owned child readiness PIDs were invalid") from error
    if (
        parent_pid <= 1
        or child_pid <= 1
        or parent_pid == child_pid
        or tuple(str(pid).encode("ascii") for pid in (parent_pid, child_pid))
        != tuple(fields)
    ):
        raise AssertionError("owned child readiness PIDs were noncanonical")
    return parent_pid, child_pid


def _write_fifo(path: Path, payload: bytes) -> None:
    descriptor = os.open(path, os.O_WRONLY | os.O_NONBLOCK)
    try:
        if os.write(descriptor, payload) != len(payload):
            raise AssertionError("short write to owned lifecycle fifo")
    finally:
        os.close(descriptor)


def _await_pidfd_exit(descriptor: int, deadline: float) -> None:
    remaining = deadline - time.monotonic()
    if remaining <= 0:
        raise AssertionError("owned child exit deadline elapsed")
    poller = select.poll()
    poller.register(descriptor, select.POLLIN | select.POLLERR | select.POLLHUP)
    if not any(
        events & select.POLLIN
        for _descriptor, events in poller.poll(int(remaining * 1000))
    ):
        raise AssertionError("owned child did not exit before the prewait deadline")


def _owned_session_group(process: subprocess.Popen[str]) -> int | None:
    parent_pid = process.pid
    if parent_pid <= 1:
        raise AssertionError("refusing to signal an unsafe parent pid")
    try:
        process_group = os.getpgid(parent_pid)
    except ProcessLookupError:
        return None
    if process_group != parent_pid:
        raise AssertionError("refusing to signal an unproven process group")
    return process_group


def _reap_owned_session(process: subprocess.Popen[str]) -> None:
    if process.returncode is not None:
        return
    process_group = _owned_session_group(process)
    if process.poll() is None and process_group is not None:
        try:
            os.killpg(process_group, signal.SIGTERM)
        except ProcessLookupError:
            pass
    try:
        process.wait(timeout=1)
    except subprocess.TimeoutExpired:
        process_group = _owned_session_group(process)
        if process_group is not None:
            try:
                os.killpg(process_group, signal.SIGKILL)
            except ProcessLookupError:
                pass
        process.wait(timeout=1)


def test_shell_child_is_absent_or_zombie_before_explicit_wait(tmp_path: Path) -> None:
    release_fifo = tmp_path / "release"
    wait_control_fifo = tmp_path / "wait-control"
    status_path = tmp_path / "wait-status"
    os.mkfifo(release_fifo, 0o600)
    os.mkfifo(wait_control_fifo, 0o600)

    process = subprocess.Popen(
        [
            "setsid",
            "/bin/sh",
            "-c",
            _PARENT_SCRIPT,
            "backend-unit-owned-parent",
            str(release_fifo),
            str(wait_control_fifo),
            str(status_path),
        ],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )
    deadline = time.monotonic() + _SUBPROCESS_SECONDS
    child_pidfd: int | None = None
    try:
        assert process.stdout is not None
        parent_pid, child_pid = _read_readiness_record(
            process.stdout.fileno(),
            min(deadline, time.monotonic() + _OBSERVATION_SECONDS),
        )
        assert parent_pid == process.pid > 1
        assert os.getpgid(parent_pid) == parent_pid

        parent_identity = _read_proc_identity(parent_pid)
        child_identity = _read_proc_identity(child_pid)
        assert parent_identity is not None
        assert child_identity is not None
        child_state, child_session, child_starttime = child_identity
        _parent_state, parent_session, _parent_starttime = parent_identity
        assert child_state != "Z"
        assert child_session == parent_session == str(parent_pid)

        try:
            child_pidfd = os.pidfd_open(child_pid)
        except AttributeError as error:
            raise AssertionError("Linux pidfd support is unavailable") from error
        except OSError as error:
            raise AssertionError("owned child pidfd could not be opened") from error

        _write_fifo(release_fifo, b"release\n")
        _await_pidfd_exit(
            child_pidfd, min(deadline, time.monotonic() + _OBSERVATION_SECONDS)
        )
        prewait_identity = _read_proc_identity(child_pid)
        if prewait_identity is None:
            prewait_state, observed_session, observed_starttime = "absent", "", ""
        else:
            prewait_state, observed_session, observed_starttime = prewait_identity
        assert prewait_state in {"Z", "absent"}
        print(
            "BACKEND_UNIT_SHELL_CHILD_PREWAIT "
            f"state={prewait_state} child_pid={child_pid} "
            f"child_session={observed_session or child_session} "
            f"child_starttime={observed_starttime or child_starttime} "
            f"parent_session={parent_session}"
        )

        _write_fifo(wait_control_fifo, b"wait\n")
        remaining = deadline - time.monotonic()
        assert remaining > 0
        parent_stdout, parent_stderr = process.communicate(timeout=remaining)

        assert process.returncode == 0, parent_stderr
        assert parent_stdout == ""
        assert parent_stderr == ""
        assert status_path.read_text(encoding="ascii") == "17\n"
    finally:
        try:
            if child_pidfd is not None:
                os.close(child_pidfd)
        finally:
            _reap_owned_session(process)
