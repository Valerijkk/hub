from __future__ import annotations

from pathlib import Path
import subprocess


_RUNTIME_PATH = "/opt/hub/testframework/docker/backend_unit/runtime.sh"
_GOTMPDIR = "/var/tmp/hub-go"
_SCRATCH = "/tmp/.backend-unit-stage-build.A1B2C3D4"
_SCRATCH_IDENTITY = "11:12:10001:10001:700:directory"
_SHELL_TIMEOUT_SECONDS = 5.0

_MONITOR_DRIVER = """\
. "$0"
go_child_pid=4242
go_child_pgid=4242
go_child_session=4242
go_child_starttime=17
go_child_identity="$1"
capture_result="$2"
leader_result="$3"
group_result="$4"
probe_log="$5"
GOTMPDIR="$6"
stage="$7"
stage_scratch="$8"
stage_scratch_identity="$9"
package_ordinal="${10}"
budget_stop_marker="${11}"

_execute_scratch_is_current() {
  [ "$#" -eq 2 ] || return 64
  [ "$1" = "$stage_scratch" ] || return 75
  [ "$2" = "$stage_scratch_identity" ] || return 75
}

# Explicit test-only identity/process-boundary stubs.  They do not assert
# anything about a real Go process group.
_capture_owned_go_child_identity() {
  [ "$#" -eq 0 ] || return 64
  printf '%s\n' capture >> "$probe_log"
  [ "$capture_result" = success ] || return 75
  go_child_identity='4242:4242:4242:17:go'
  return 0
}

_owned_go_leader_state() {
  [ "$#" -eq 0 ] || return 64
  printf '%s\n' leader >> "$probe_log"
  printf '%s\n' "$leader_result"
}

_owned_go_group_state() {
  [ "$#" -eq 0 ] || return 64
  printf '%s\n' group >> "$probe_log"
  printf '%s\n' "$group_result"
}

monitor_owned_budget_stop "$go_child_pid" "$budget_stop_marker"
"""

_REAPER_DRIVER = """\
. "$0"
event_log="$1"
result_file="$2"
leader_result="$3"
group_result="$4"
identity_mode="$5"
capture_result="$6"

/bin/sh -c 'exit 17' &
go_child_pid=$!
go_child_pgid="$go_child_pid"
go_child_session="$go_child_pid"
go_child_starttime=17
case "$identity_mode" in
  captured) go_child_identity="$go_child_pid:$go_child_pid:$go_child_pid:17:go" ;;
  empty) go_child_identity= ;;
  *) exit 64 ;;
esac
go_child_reaped=false

cleanup_owned_test_child() {
  if [ -n "${go_child_pid-}" ]; then
    wait "$go_child_pid" 2>/dev/null || :
  fi
}
trap cleanup_owned_test_child EXIT

attempt=0
while [ "$attempt" -lt 100 ]; do
  [ ! -e "/proc/$go_child_pid/stat" ] && break
  attempt=$((attempt + 1))
  sleep 0.02
done
[ ! -e "/proc/$go_child_pid/stat" ] || exit 75

# Explicit test-only identity/process-boundary stubs.  The real parent shell
# owns the child above; these stubs only model the already-captured Go group.
_capture_owned_go_child_identity() {
  [ "$#" -eq 0 ] || return 64
  printf '%s\n' capture >> "$event_log"
  [ "$capture_result" = success ] || return 75
  go_child_identity="$go_child_pid:$go_child_pid:$go_child_pid:17:go"
  return 0
}

_owned_go_leader_state() {
  [ "$#" -eq 0 ] || return 64
  printf '%s\n' leader >> "$event_log"
  printf '%s\n' "$leader_result"
}

_owned_go_group_state() {
  [ "$#" -eq 0 ] || return 64
  printf '%s\n' group >> "$event_log"
  printf '%s\n' "$group_result"
}

_owned_go_group_reaches_terminal_state() {
  [ "$#" -eq 1 ] || return 64
  [ "$1" = 5 ] || return 75
  printf '%s\n' terminal >> "$event_log"
  return 0
}

_post_reaped_gocache_cleanup() {
  [ "$#" -eq 0 ] || return 64
  printf '%s\n' cleanup >> "$event_log"
  return 0
}

if reap_owned_go_child_group; then
  reap_status=0
else
  reap_status=$?
fi
printf '%s|%s|%s|%s|%s|%s|%s|%s\n' \
  "$reap_status" "${go_child_exit-}" "${go_child_reaped-}" \
  "${go_child_pid-}" "${go_child_pgid-}" "${go_child_session-}" \
  "${go_child_starttime-}" "${go_child_identity-}" > "$result_file"
exit "$reap_status"
"""


def _run_monitor(
    tmp_path: Path,
    name: str,
    *,
    identity: str,
    capture_result: str,
    leader_result: str,
    group_result: str,
) -> tuple[subprocess.CompletedProcess[str], Path]:
    probe_log = tmp_path / f"{name}.probe"
    completed = subprocess.run(
        [
            "/bin/sh",
            "-c",
            _MONITOR_DRIVER,
            _RUNTIME_PATH,
            identity,
            capture_result,
            leader_result,
            group_result,
            str(probe_log),
            _GOTMPDIR,
            "build",
            _SCRATCH,
            _SCRATCH_IDENTITY,
            "1",
            f"{_SCRATCH}/stop.1",
        ],
        check=False,
        capture_output=True,
        text=True,
        timeout=_SHELL_TIMEOUT_SECONDS,
    )
    return completed, probe_log


def _run_reaper(
    tmp_path: Path,
    name: str,
    *,
    leader_result: str,
    group_result: str,
    identity_mode: str = "captured",
    capture_result: str = "unexpected",
) -> tuple[subprocess.CompletedProcess[str], Path, Path]:
    event_log = tmp_path / f"{name}.events"
    result_file = tmp_path / f"{name}.result"
    completed = subprocess.run(
        [
            "/bin/sh",
            "-c",
            _REAPER_DRIVER,
            _RUNTIME_PATH,
            str(event_log),
            str(result_file),
            leader_result,
            group_result,
            identity_mode,
            capture_result,
        ],
        check=False,
        capture_output=True,
        text=True,
        timeout=_SHELL_TIMEOUT_SECONDS,
    )
    return completed, event_log, result_file


def _read_if_present(path: Path) -> str:
    return path.read_text(encoding="utf-8") if path.exists() else ""


def test_monitor_accepts_captured_missing_leader_and_empty_group(tmp_path: Path) -> None:
    completed, probe_log = _run_monitor(
        tmp_path,
        "captured-missing-none",
        identity="4242:4242:4242:17:go",
        capture_result="unexpected",
        leader_result="missing",
        group_result="none",
    )

    assert completed.returncode == 0, completed.stderr
    assert completed.stdout == ""
    assert completed.stderr == ""
    assert _read_if_present(probe_log) == "leader\ngroup\n"


def test_monitor_missing_empty_preserves_fixed_ram_marker_branch_source_contract() -> None:
    runtime = Path(_RUNTIME_PATH).read_text(encoding="utf-8")
    monitor = runtime.split("monitor_owned_budget_stop() {", 1)[1].split(
        "\n}\n\nterminate_owned_go_child_group()", 1
    )[0]

    # Python's framework rootfs is read-only, so it cannot create the exact
    # fixed-RAM marker.  The executable absent-marker oracle is above; this is
    # deliberately bounded source evidence for the required present-marker=1 branch.
    assert (
        "      valid-zombie:valid-zombie|missing:none)\n"
        '        [ ! -e "$budget_stop_marker" ] || return 1\n'
        "        return 0\n"
        "        ;;"
    ) in monitor


def test_monitor_rejects_uncaptured_and_nonterminal_or_inconsistent_states(
    tmp_path: Path,
) -> None:
    cases = (
        ("uncaptured", "", "fail", "missing", "none", "capture\n"),
        ("missing-live", "4242:4242:4242:17:go", "unexpected", "missing", "valid-live", "leader\ngroup\n"),
        ("invalid-none", "4242:4242:4242:17:go", "unexpected", "invalid", "none", "leader\ngroup\n"),
        ("live-none", "4242:4242:4242:17:go", "unexpected", "valid-live", "none", "leader\ngroup\n"),
    )

    for name, identity, capture_result, leader_result, group_result, expected_log in cases:
        completed, probe_log = _run_monitor(
            tmp_path,
            name,
            identity=identity,
            capture_result=capture_result,
            leader_result=leader_result,
            group_result=group_result,
        )

        assert completed.returncode == 75, name
        assert completed.stdout == "", name
        assert completed.stderr == "", name
        assert _read_if_present(probe_log) == expected_log, name


def test_reaper_waits_for_missing_owned_shell_child_and_preserves_cached_exit_17(
    tmp_path: Path,
) -> None:
    completed, event_log, result_file = _run_reaper(
        tmp_path, "missing-none", leader_result="missing", group_result="none"
    )

    assert completed.returncode == 0, completed.stderr
    assert completed.stdout == ""
    assert completed.stderr == ""
    assert _read_if_present(event_log) == "leader\ngroup\nterminal\ncleanup\n"
    assert _read_if_present(result_file) == "0|17|true|||||\n"


def test_reaper_keeps_valid_zombie_baseline_before_real_parent_wait(tmp_path: Path) -> None:
    completed, event_log, result_file = _run_reaper(
        tmp_path, "zombie-baseline", leader_result="valid-zombie", group_result="valid-zombie"
    )

    assert completed.returncode == 0, completed.stderr
    assert completed.stdout == ""
    assert completed.stderr == ""
    assert _read_if_present(event_log) == "leader\ngroup\nterminal\ncleanup\n"
    assert _read_if_present(result_file) == "0|17|true|||||\n"


def test_reaper_rejects_uncaptured_and_nonterminal_or_inconsistent_states(
    tmp_path: Path,
) -> None:
    cases = (
        ("missing-live", "captured", "unexpected", "missing", "valid-live", "leader\ngroup\n"),
        ("invalid-none", "captured", "unexpected", "invalid", "none", "leader\ngroup\n"),
        ("uncaptured", "empty", "fail", "missing", "none", "capture\n"),
    )

    for name, identity_mode, capture_result, leader_result, group_result, expected_log in cases:
        completed, event_log, result_file = _run_reaper(
            tmp_path,
            name,
            leader_result=leader_result,
            group_result=group_result,
            identity_mode=identity_mode,
            capture_result=capture_result,
        )

        assert completed.returncode == 75, name
        assert completed.stdout == "", name
        assert completed.stderr == "", name
        assert _read_if_present(event_log) == expected_log, name
        assert _read_if_present(result_file).startswith("75|"), name
