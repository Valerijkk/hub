from __future__ import annotations

from pathlib import Path


_REPO_ROOT = Path(__file__).resolve().parents[3]
_EXECUTE_PATH = _REPO_ROOT / "testframework/docker/backend_unit/execute.sh"
_DIAGNOSTICS = (
    (
        "BACKEND_UNIT_ABORT_CHILD_START",
        'start_owned_go_child_group "$stage" "$package_import_path"',
        '_execute_abort_unpublished_stage "" "" || return 75',
    ),
    (
        "BACKEND_UNIT_ABORT_STDOUT_READER_IDENTITY",
        "if stdout_reader_identity=",
        '_execute_abort_unpublished_stage "$stdout_reader_pid" "" || return 75',
    ),
    (
        "BACKEND_UNIT_ABORT_STDERR_READER_IDENTITY",
        "if stderr_reader_identity=",
        '_execute_abort_unpublished_stage "$stdout_reader_pid" "$stderr_reader_pid" || return 75',
    ),
    (
        "BACKEND_UNIT_ABORT_MONITOR",
        'monitor_owned_budget_stop "$go_child_pid" "$budget_stop_marker"',
        '_execute_abort_unpublished_stage "$stdout_reader_pid" "$stderr_reader_pid" || return 75',
    ),
    (
        "BACKEND_UNIT_ABORT_READER_EXIT",
        '[ -e "$budget_stop_marker" ] || [ "$stdout_reader_exit" -ne 0 ]',
        '_execute_abort_unpublished_stage "" "" || return 75',
    ),
    (
        "BACKEND_UNIT_ABORT_REAP",
        "if reap_owned_go_child_group; then",
        '_execute_abort_unpublished_stage "" "" || return 75',
    ),
)


def _run_stage_lines() -> list[str]:
    source = _EXECUTE_PATH.read_text(encoding="utf-8")
    start = source.index("run_stage() {\n")
    end = source.index("\n}\n", start) + 2
    return [line.strip() for line in source[start:end].splitlines()]


def test_run_stage_emits_fixed_abort_diagnostics_before_existing_cleanup() -> None:
    lines = _run_stage_lines()

    for marker, failure_condition, abort_call in _DIAGNOSTICS:
        marker_line = f"printf '%s\\n' '{marker}' >&2"
        assert lines.count(marker_line) == 1

        condition_indexes = [
            index for index, line in enumerate(lines) if failure_condition in line
        ]
        assert len(condition_indexes) == 1
        failure_block = lines[condition_indexes[0] : condition_indexes[0] + 10]
        marker_index = failure_block.index(marker_line)
        assert failure_block[marker_index + 1].strip() == abort_call
