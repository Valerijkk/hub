from __future__ import annotations

import errno
import os
from pathlib import Path
import signal
import subprocess
import time


_POLL_SECONDS = 0.02
_OBSERVATION_SECONDS = 1.5
_SUBPROCESS_SECONDS = 5.0
_PARENT_SCRIPT = """\
identity_file="$1"
release_fifo="$2"
wait_sentinel="$3"
status_file="$4"

/bin/sh -c 'IFS= read -r _ < "$1"; exit 17' backend-unit-owned-child "$release_fifo" &
owned_pid=$!
printf '%s %s\n' "$$" "$owned_pid" > "$identity_file"

iteration=0
while [ "$iteration" -lt 100 ]; do
  [ -e "$wait_sentinel" ] && break
  iteration=$((iteration + 1))
  sleep 0.02
done
[ -e "$wait_sentinel" ] || exit 75

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


def _wait_for_identity(path: Path, deadline: float) -> tuple[int, int]:
    while time.monotonic() < deadline:
        try:
            fields = path.read_text(encoding="ascii").split()
        except FileNotFoundError:
            fields = []
        if len(fields) == 2:
            parent_pid, child_pid = (int(field) for field in fields)
            if parent_pid > 1 and child_pid > 1:
                return parent_pid, child_pid
        time.sleep(_POLL_SECONDS)
    raise AssertionError("owned child identity was not captured before timeout")


def _write_fifo(path: Path, payload: bytes, deadline: float) -> None:
    while time.monotonic() < deadline:
        try:
            descriptor = os.open(path, os.O_WRONLY | os.O_NONBLOCK)
        except OSError as error:
            if error.errno != errno.ENXIO:
                raise
            time.sleep(_POLL_SECONDS)
            continue
        try:
            written = os.write(descriptor, payload)
            if written != len(payload):
                raise AssertionError("short write to owned lifecycle fifo")
            return
        finally:
            os.close(descriptor)
    raise AssertionError("owned lifecycle fifo did not gain a reader before timeout")


def _observe_prewait_state(
    child_pid: int, deadline: float
) -> tuple[str, str, str]:
    final_zombie: tuple[str, str, str] | None = None
    while time.monotonic() < deadline:
        identity = _read_proc_identity(child_pid)
        if identity is None:
            return "absent", "", ""
        if identity[0] == "Z":
            final_zombie = identity
        time.sleep(_POLL_SECONDS)
    if final_zombie is not None:
        return final_zombie
    raise AssertionError("owned child was neither absent nor zombie before wait")


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
    identity_path = tmp_path / "identity"
    release_fifo = tmp_path / "release"
    wait_sentinel = tmp_path / "wait"
    status_path = tmp_path / "wait-status"
    os.mkfifo(release_fifo, 0o600)

    process = subprocess.Popen(
        [
            "setsid",
            "/bin/sh",
            "-c",
            _PARENT_SCRIPT,
            "backend-unit-owned-parent",
            str(identity_path),
            str(release_fifo),
            str(wait_sentinel),
            str(status_path),
        ],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )
    deadline = time.monotonic() + _SUBPROCESS_SECONDS
    try:
        parent_pid, child_pid = _wait_for_identity(
            identity_path, min(deadline, time.monotonic() + _OBSERVATION_SECONDS)
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

        _write_fifo(
            release_fifo, b"release\n", min(deadline, time.monotonic() + _OBSERVATION_SECONDS)
        )
        prewait_state, observed_session, observed_starttime = _observe_prewait_state(
            child_pid, min(deadline, time.monotonic() + _OBSERVATION_SECONDS)
        )
        print(
            "BACKEND_UNIT_SHELL_CHILD_PREWAIT "
            f"state={prewait_state} child_pid={child_pid} "
            f"child_session={observed_session or child_session} "
            f"child_starttime={observed_starttime or child_starttime} "
            f"parent_session={parent_session}"
        )

        wait_sentinel.write_text("wait\n", encoding="ascii")
        remaining = deadline - time.monotonic()
        assert remaining > 0
        parent_stdout, parent_stderr = process.communicate(timeout=remaining)

        assert process.returncode == 0, parent_stderr
        assert parent_stdout == ""
        assert parent_stderr == ""
        assert status_path.read_text(encoding="ascii") == "17\n"
    finally:
        _reap_owned_session(process)
