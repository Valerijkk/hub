from __future__ import annotations

import os
from pathlib import Path
import subprocess


_RUNTIME_PATH = "/opt/hub/testframework/docker/backend_unit/runtime.sh"
_PLAN_PATH = "/opt/hub/testframework/docker/backend_unit/plan.sh"
_EXECUTE_PATH = "/opt/hub/testframework/docker/backend_unit/execute.sh"
_GOTMPDIR = "/var/tmp/hub-go"
_SCRATCH_ROOT = "/tmp"
_SHELL_TIMEOUT_SECONDS = 5.0

_CREATE_SCRATCH_DRIVER = """\
. "$0"
GOTMPDIR="$1"
stage="$2"
created_scratch=
created_scratch_identity=

cleanup_created_scratch() {
  cleanup_status=$?
  trap - EXIT
  if [ -n "$created_scratch" ] && [ -n "$created_scratch_identity" ]; then
    if _execute_scratch_is_current "$created_scratch" "$created_scratch_identity"; then
      rmdir -- "$created_scratch" || cleanup_status=75
    else
      cleanup_status=75
    fi
  fi
  exit "$cleanup_status"
}
trap cleanup_created_scratch EXIT

if _execute_create_scratch "$stage"; then
  created_scratch="$stage_scratch"
  created_scratch_identity="$stage_scratch_identity"
else
  create_status=$?
  if [ -n "${stage_scratch-}" ]; then
    created_scratch="$stage_scratch"
    created_scratch_identity="$(_execute_scratch_identity "$created_scratch")" || exit 75
  fi
  exit "$create_status"
fi
case "$stage_scratch" in
  /tmp/.backend-unit-stage-"$stage".[A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9]) ;;
  *) exit 75 ;;
esac
current_identity="$(_execute_scratch_identity "$stage_scratch")" || exit 75
[ "$current_identity" = "$stage_scratch_identity" ] || exit 75
printf '%s\n%s\n' "$stage_scratch" "$stage_scratch_identity"
exit 0
"""

_QUIESCENCE_DRIVER = """\
. "$0"
. "$1"
. "$2"
run_dir="$3"
GOTMPDIR="$4"
stage_scratch="$5"
stage_scratch_identity="$6"
stage=build

stage_run_identity="$(_owned_run_dir_identity)" || exit 75
stage_plan_path="$run_dir/package-plan.build.json"
stage_results_path="$run_dir/package-results.build.jsonl"
stdout_path="$run_dir/build.stdout.log"
stderr_path="$run_dir/build.stderr.log"
stage_json_path="$run_dir/stage-build.json"
stage_results_identity="$(_owned_regular_identity "$stage_results_path")" || exit 75
stage_stdout_identity="$(_owned_regular_identity "$stdout_path")" || exit 75
stage_stderr_identity="$(_owned_regular_identity "$stderr_path")" || exit 75
stage_json_identity="$(_owned_regular_identity "$stage_json_path")" || exit 75
stage_results_tmp="$run_dir/.backend-unit-publish.A1B2C3D4"
stage_stdout_tmp="$run_dir/.backend-unit-publish.B1C2D3E4"
stage_stderr_tmp="$run_dir/.backend-unit-publish.C1D2E3F4"
stage_json_tmp="$run_dir/.backend-unit-publish.D1E2F3G4"
stage_marker_template="$stage_scratch/marker.A1B2C3D4"
stage_marker_identity='11:23:10001:10001:600:regular file'
budget_stop_marker="$stage_scratch/stop.1"
stdout_fifo="$stage_scratch/stdout.1"
stdout_fifo_identity='11:24:10001:10001:600:fifo'
stderr_fifo="$stage_scratch/stderr.1"
stderr_fifo_identity='11:25:10001:10001:600:fifo'
package_stdout_capture="$stage_scratch/stdout-capture.A1B2C3D4"
package_stdout_identity='11:26:10001:10001:600:regular file'
package_stderr_capture="$stage_scratch/stderr-capture.A1B2C3D4"
package_stderr_identity='11:27:10001:10001:600:regular file'
package_ordinal=1
completed_build_final_identities="$stage_results_identity|$stage_stdout_identity|$stage_stderr_identity|$stage_json_identity"
completed_vet_final_identities=
completed_unit_final_identities=
completed_race_final_identities=
go_child_started=false
go_child_reaped=true
go_child_exit=0
stdout_reader_started=false
stderr_reader_started=false
stdout_reader_pid=
stderr_reader_pid=
stdout_reader_identity=
stderr_reader_identity=
owned_output=
package_scratch_ready=true
package_scratch_cleaned=true
child_group_reaped=true
stdout_reader_reaped=true
stderr_reader_reaped=true
unpublished_stage_rolled_back=false

if prove_quiescent_for_container_stop; then
  printf '%s\n' "$unpublished_stage_rolled_back"
  exit 0
else
  quiescence_status=$?
  exit "$quiescence_status"
fi
"""


def _run_create_scratch(
    gotmpdir: Path, stage: str
) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        [
            "/bin/sh",
            "-c",
            _CREATE_SCRATCH_DRIVER,
            _EXECUTE_PATH,
            str(gotmpdir),
            stage,
        ],
        check=False,
        capture_output=True,
        text=True,
        timeout=_SHELL_TIMEOUT_SECONDS,
    )


def _owned_regular_identity(path: Path) -> str:
    metadata = os.stat(path)
    assert (metadata.st_uid, metadata.st_gid) == (10001, 10001)
    assert metadata.st_mode & 0o777 == 0o600
    return f"{metadata.st_dev}:{metadata.st_ino}:10001:10001:600:regular file"


def _create_completed_build_snapshot(run_dir: Path) -> None:
    run_dir.mkdir()
    run_dir.chmod(0o700)
    for name, payload in (
        ("package-plan.build.json", b"{}\n"),
        ("package-results.build.jsonl", b"{}\n"),
        ("build.stdout.log", b""),
        ("build.stderr.log", b""),
        ("stage-build.json", b"{}\n"),
    ):
        path = run_dir / name
        path.write_bytes(payload)
        path.chmod(0o600)
        _owned_regular_identity(path)


def _run_quiescence(
    tmp_path: Path, name: str, scratch: str, scratch_identity: str
) -> subprocess.CompletedProcess[str]:
    run_dir = tmp_path / name
    _create_completed_build_snapshot(run_dir)
    scratch_path = Path(scratch)
    assert not scratch_path.exists()
    assert not scratch_path.is_symlink()
    return subprocess.run(
        [
            "/bin/sh",
            "-c",
            _QUIESCENCE_DRIVER,
            _RUNTIME_PATH,
            _PLAN_PATH,
            _EXECUTE_PATH,
            str(run_dir),
            _GOTMPDIR,
            scratch,
            scratch_identity,
        ],
        check=False,
        capture_output=True,
        text=True,
        timeout=_SHELL_TIMEOUT_SECONDS,
    )


def test_create_scratch_isolated_from_gotmpdir_for_each_stage(tmp_path: Path) -> None:
    for stage in ("build", "vet", "unit", "race"):
        gotmpdir = tmp_path / f"gotmp-{stage}"
        gotmpdir.mkdir()
        gotmpdir.chmod(0o700)

        completed = _run_create_scratch(gotmpdir, stage)

        assert completed.returncode == 0, completed.stderr
        scratch, identity = completed.stdout.splitlines()
        assert scratch.startswith(f"{_SCRATCH_ROOT}/.backend-unit-stage-{stage}.")
        assert identity.endswith(":10001:10001:700:directory")
        assert completed.stderr == ""
        assert list(gotmpdir.iterdir()) == []


def test_quiescent_completed_build_accepts_removed_tmp_scratch(tmp_path: Path) -> None:
    completed = _run_quiescence(
        tmp_path,
        "accepted",
        "/tmp/.backend-unit-stage-build.A1B2C3D4",
        "11:22:10001:10001:700:directory",
    )

    assert completed.returncode == 0, completed.stderr
    assert completed.stdout == "true\n"
    assert completed.stderr == ""


def test_quiescent_completed_build_rejects_legacy_or_invalid_scratch_snapshot(
    tmp_path: Path,
) -> None:
    cases = (
        (
            "legacy-gotmpdir-root",
            f"{_GOTMPDIR}/.backend-unit-stage-build.A1B2C3D4",
            "11:22:10001:10001:700:directory",
        ),
        (
            "traversal",
            "/tmp/.backend-unit-stage-build.A1B2C3D4/../escape",
            "11:22:10001:10001:700:directory",
        ),
        (
            "mismatched-stage",
            "/tmp/.backend-unit-stage-vet.A1B2C3D4",
            "11:22:10001:10001:700:directory",
        ),
        (
            "malformed-identity",
            "/tmp/.backend-unit-stage-build.A1B2C3D4",
            "not-an-identity",
        ),
    )

    for name, scratch, identity in cases:
        completed = _run_quiescence(tmp_path, name, scratch, identity)

        assert completed.returncode == 75, name
        assert completed.stdout == "", name
        assert completed.stderr == "", name
