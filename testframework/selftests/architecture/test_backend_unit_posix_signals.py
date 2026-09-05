from __future__ import annotations

import os
import re
import signal
import subprocess
from pathlib import Path

import pytest


_REPO_ROOT = Path(__file__).resolve().parents[3]
_KILL_SOURCES = (
    "testframework/docker/backend_unit/runtime.sh",
    "testframework/docker/backend_unit/execute.sh",
)
_KILL_COMMAND = re.compile(r"^\s*(kill\b.+?)\s+\|\|\s+return 75\s*$")
_SHELL_DRIVER = """\
target_pid=$1
go_child_pgid=$target_pid
execute_stop_reader_pid=$target_pid
signal_owned_child() {
  {command}
}
signal_owned_child
"""


def _literal_kill_commands() -> list[str]:
    commands: list[str] = []
    for relative_path in _KILL_SOURCES:
        for line in (_REPO_ROOT / relative_path).read_text(encoding="utf-8").splitlines():
            match = _KILL_COMMAND.fullmatch(line)
            if match is not None:
                commands.append(match.group(1))

    assert len(commands) == 5
    return commands


def _run_exact_kill_command(command: str, child_pid: int) -> subprocess.CompletedProcess[str]:
    assert re.fullmatch(
        r'kill -(TERM|KILL) (?:-- )?"(?:-\$go_child_pgid|\$execute_stop_reader_pid)"',
        command,
    ), "unexpected signal command; refuse execution"
    assert child_pid > 1
    return subprocess.run(
        ["/bin/sh", "-c", _SHELL_DRIVER.replace("{command}", command), "kill-test", str(child_pid)],
        check=False,
        capture_output=True,
        text=True,
        timeout=5,
    )


def _cleanup_owned_child(child: subprocess.Popen[bytes]) -> None:
    if child.poll() is None:
        assert child.pid > 1
        assert os.getpgid(child.pid) == child.pid
        os.killpg(child.pid, signal.SIGKILL)
    child.wait(timeout=5)


@pytest.mark.parametrize("command", _literal_kill_commands())
def test_backend_unit_posix_kill_command_signals_only_its_owned_child_group(
    command: str,
) -> None:
    child = subprocess.Popen(
        ["/bin/sh", "-c", "exec sleep 30"],
        start_new_session=True,
    )
    try:
        assert child.pid > 1
        assert os.getpgid(child.pid) == child.pid
        result = _run_exact_kill_command(command, child.pid)

        assert result.returncode == 0, result.stderr
        expected_signal = signal.SIGTERM if command.startswith("kill -TERM ") else signal.SIGKILL
        assert child.wait(timeout=5) == -expected_signal
    finally:
        _cleanup_owned_child(child)
