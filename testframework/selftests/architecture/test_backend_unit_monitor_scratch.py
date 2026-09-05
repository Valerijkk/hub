from __future__ import annotations

from pathlib import Path
import subprocess


_RUNTIME_PATH = "/opt/hub/testframework/docker/backend_unit/runtime.sh"
_RAM_TMP = "/var/tmp/hub-go"
_SCRATCH_IDENTITY = "11:12:10001:10001:700:directory"
_SHELL_DRIVER = """\
. "$0"
run_dir="$1"
GOTMPDIR="$2"
stage="$3"
stage_scratch="$4"
stage_scratch_identity="$5"
package_ordinal="$6"
expected_scratch="$7"
expected_scratch_identity="$8"
probe_log="$9"
scratch_log="${10}"
monitored_pid="${11}"
budget_stop_marker="${12}"
go_child_pid=4242
go_child_identity='4242:4242:4242:17:go'

_execute_scratch_is_current() {
  [ "$#" -eq 2 ] || return 64
  [ "$1" = "$expected_scratch" ] || return 75
  [ "$2" = "$expected_scratch_identity" ] || return 75
  printf '%s\n' "$1|$2" >> "$scratch_log"
}

_owned_go_leader_state() {
  [ "$#" -eq 0 ] || return 64
  printf '%s\n' leader >> "$probe_log"
  printf '%s\n' valid-zombie
}

_owned_go_group_state() {
  [ "$#" -eq 0 ] || return 64
  printf '%s\n' group >> "$probe_log"
  printf '%s\n' valid-zombie
}

monitor_owned_budget_stop "$monitored_pid" "$budget_stop_marker"
"""


def _run_monitor(
    tmp_path: Path,
    name: str,
    *,
    gotmpdir: str,
    stage: str,
    stage_scratch: str,
    stage_scratch_identity: str,
    package_ordinal: str,
    budget_stop_marker: str,
    monitored_pid: str,
    expected_scratch: str,
    expected_scratch_identity: str,
) -> tuple[subprocess.CompletedProcess[str], Path, Path]:
    probe_log = tmp_path / f"{name}.probe"
    scratch_log = tmp_path / f"{name}.scratch"
    completed = subprocess.run(
        [
            "/bin/sh",
            "-c",
            _SHELL_DRIVER,
            _RUNTIME_PATH,
            str(tmp_path / "report"),
            gotmpdir,
            stage,
            stage_scratch,
            stage_scratch_identity,
            package_ordinal,
            expected_scratch,
            expected_scratch_identity,
            str(probe_log),
            str(scratch_log),
            monitored_pid,
            budget_stop_marker,
        ],
        check=False,
        capture_output=True,
        text=True,
        timeout=5,
    )
    return completed, probe_log, scratch_log


def _log(path: Path) -> str:
    return path.read_text(encoding="utf-8") if path.exists() else ""


def _owned_report_dir(tmp_path: Path) -> Path:
    report_dir = tmp_path / "report"
    report_dir.mkdir()
    report_dir.chmod(0o700)
    metadata = report_dir.stat()
    assert (metadata.st_uid, metadata.st_gid) == (10001, 10001)
    return report_dir


def test_monitor_accepts_current_ram_scratch_stop_marker(tmp_path: Path) -> None:
    report_dir = _owned_report_dir(tmp_path)
    scratch = f"{_RAM_TMP}/.backend-unit-stage-build.A1B2C3D4"
    completed, probe_log, scratch_log = _run_monitor(
        tmp_path,
        "valid",
        gotmpdir=_RAM_TMP,
        stage="build",
        stage_scratch=scratch,
        stage_scratch_identity=_SCRATCH_IDENTITY,
        package_ordinal="1",
        budget_stop_marker=f"{scratch}/stop.1",
        monitored_pid="4242",
        expected_scratch=scratch,
        expected_scratch_identity=_SCRATCH_IDENTITY,
    )

    assert report_dir.is_dir()
    assert completed.returncode == 0, completed.stderr
    assert completed.stdout == ""
    assert completed.stderr == ""
    assert _log(scratch_log) == f"{scratch}|{_SCRATCH_IDENTITY}\n"
    assert _log(probe_log) == "leader\ngroup\n"


def test_monitor_rejects_noncurrent_or_nonram_markers_before_process_probes(
    tmp_path: Path,
) -> None:
    report_dir = _owned_report_dir(tmp_path)
    scratch = f"{_RAM_TMP}/.backend-unit-stage-build.A1B2C3D4"
    cases = (
        (
            "report-root-marker",
            _RAM_TMP,
            "build",
            scratch,
            _SCRATCH_IDENTITY,
            "1",
            str(report_dir / "stop.1"),
            "4242",
            scratch,
            _SCRATCH_IDENTITY,
        ),
        (
            "traversal",
            _RAM_TMP,
            "build",
            f"{scratch}/../escape",
            _SCRATCH_IDENTITY,
            "1",
            f"{scratch}/../escape/stop.1",
            "4242",
            f"{scratch}/../escape",
            _SCRATCH_IDENTITY,
        ),
        (
            "wrong-ordinal",
            _RAM_TMP,
            "build",
            scratch,
            _SCRATCH_IDENTITY,
            "0",
            f"{scratch}/stop.0",
            "4242",
            scratch,
            _SCRATCH_IDENTITY,
        ),
        (
            "wrong-stage",
            _RAM_TMP,
            "lint",
            f"{_RAM_TMP}/.backend-unit-stage-lint.A1B2C3D4",
            _SCRATCH_IDENTITY,
            "1",
            f"{_RAM_TMP}/.backend-unit-stage-lint.A1B2C3D4/stop.1",
            "4242",
            f"{_RAM_TMP}/.backend-unit-stage-lint.A1B2C3D4",
            _SCRATCH_IDENTITY,
        ),
        (
            "wrong-gotmpdir",
            "/tmp/hub-go",
            "build",
            "/tmp/hub-go/.backend-unit-stage-build.A1B2C3D4",
            _SCRATCH_IDENTITY,
            "1",
            "/tmp/hub-go/.backend-unit-stage-build.A1B2C3D4/stop.1",
            "4242",
            "/tmp/hub-go/.backend-unit-stage-build.A1B2C3D4",
            _SCRATCH_IDENTITY,
        ),
        (
            "stale-scratch",
            _RAM_TMP,
            "build",
            scratch,
            "stale",
            "1",
            f"{scratch}/stop.1",
            "4242",
            scratch,
            _SCRATCH_IDENTITY,
        ),
        (
            "mismatched-pid",
            _RAM_TMP,
            "build",
            scratch,
            _SCRATCH_IDENTITY,
            "1",
            f"{scratch}/stop.1",
            "4243",
            scratch,
            _SCRATCH_IDENTITY,
        ),
    )

    for (
        name,
        gotmpdir,
        stage,
        stage_scratch,
        stage_scratch_identity,
        package_ordinal,
        budget_stop_marker,
        monitored_pid,
        expected_scratch,
        expected_scratch_identity,
    ) in cases:
        completed, probe_log, _scratch_log = _run_monitor(
            tmp_path,
            name,
            gotmpdir=gotmpdir,
            stage=stage,
            stage_scratch=stage_scratch,
            stage_scratch_identity=stage_scratch_identity,
            package_ordinal=package_ordinal,
            budget_stop_marker=budget_stop_marker,
            monitored_pid=monitored_pid,
            expected_scratch=expected_scratch,
            expected_scratch_identity=expected_scratch_identity,
        )

        assert completed.returncode == 75, name
        assert completed.stdout == "", name
        assert completed.stderr == "", name
        assert _log(probe_log) == "", name
