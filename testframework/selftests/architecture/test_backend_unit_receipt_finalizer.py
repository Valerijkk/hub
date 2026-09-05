from __future__ import annotations

import json
import subprocess


_RECEIPT_PATH = "/opt/hub/testframework/docker/backend_unit/receipt.sh"
_BUDGET_JSON = (
    '{"metadata_used_bytes":101,"raw_output_partial":false,'
    '"total_used_bytes":202,"truncated_due_to_budget":false}'
)
_TERMINAL_EXIT = 17
_SHA256 = "sha256:" + "a" * 64
_RUN_ID = "backend-unit-20260905T000000Z-0123456789ab-01234567"
_STAGE_RECEIPT = json.dumps(
    {
        "schema": "hub.qa.backend-unit-stage-result/v1", "target": "backend-unit",
        "run_id": _RUN_ID, "owner_nonce": "0123456789abcdef0123456789abcdef",
        "stage": "build", "package_plan_sha256": _SHA256,
        "package_result_sha256": _SHA256, "package_count": 1,
        "package_result_count": 1, "first_exit": 0, "status": "PASS",
        "stdout_sha256": _SHA256, "stdout_size": 0, "stderr_sha256": _SHA256,
        "stderr_size": 0, "cgroup_baseline": "baseline", "cgroup_final": "final",
        "report_budget_truncated_due_to_budget": False,
    }, separators=(",", ":"),
)
_STAGE_PARSER_SHELL_DRIVER = f"""
. "$0"
run_id='{_RUN_ID}'
owner_nonce='0123456789abcdef0123456789abcdef'
receipt_stage='build'
receipt_stage_plan_sha='{_SHA256}'
receipt_stage_results_sha='{_SHA256}'
receipt_stage_stdout_sha='{_SHA256}'
receipt_stage_stdout_size=0
receipt_stage_stderr_sha='{_SHA256}'
receipt_stage_stderr_size=0
container_cgroup_baseline='baseline'
container_cgroup_final='final'
_receipt_parse_stage_json "$1"
"""
_SHELL_DRIVER = f"""\
. \"$0\"

receipt_inventory_json='{{"path":"owner.json","sha256":"{_SHA256}","size":0}}'
receipt_inventory_sha256='{_SHA256}'
container_cgroup_baseline_json='{{"memory_current":0}}'
cgroup_family='v2'
container_cgroup_final_json='{{"memory_current":0}}'
command_profile_sha256='{_SHA256}'
container_hostname='0123456789ab'
go_environment='{{"runtime_affecting_environment":[]}}'
launcher_sha256='{_SHA256}'
owner_nonce='0123456789abcdef0123456789abcdef'
build_plan_sha256='{_SHA256}'
race_plan_sha256='{_SHA256}'
unit_plan_sha256='{_SHA256}'
vet_plan_sha256='{_SHA256}'
run_id='backend-unit-20260905T000000Z-0123456789ab-01234567'
source_manifest_after_sha256='{_SHA256}'
source_manifest_before_sha256='{_SHA256}'
build_stage_sha256='{_SHA256}'
race_stage_sha256='{_SHA256}'
unit_stage_sha256='{_SHA256}'
vet_stage_sha256='{_SHA256}'
receipt_first_stage_exit={_TERMINAL_EXIT}

_receipt_build_container_json \"$@\"
printf '%s' \"$receipt\"
"""


def _build_receipt(*arguments: str) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["/bin/sh", "-c", _SHELL_DRIVER, _RECEIPT_PATH, *arguments],
        check=False,
        capture_output=True,
        text=True,
    )


def test_complete_finalizer_builder_accepts_only_one_budget_argument() -> None:
    complete = _build_receipt(_BUDGET_JSON)

    assert complete.returncode == 0, complete.stderr
    receipt = json.loads(complete.stdout)
    assert receipt["report_budget"] == json.loads(_BUDGET_JSON)
    assert receipt["terminal"] == {"exit_code": _TERMINAL_EXIT, "signal": None}

    assert _build_receipt().returncode == 64
    assert _build_receipt(_BUDGET_JSON, "unused").returncode == 64


def test_stage_receipt_parser_accepts_canonical_line_and_rejects_missing_field() -> None:
    def parse(stage_receipt: str) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            ["/bin/sh", "-c", _STAGE_PARSER_SHELL_DRIVER, _RECEIPT_PATH, stage_receipt],
            check=False, capture_output=True, text=True,
        )

    assert parse(_STAGE_RECEIPT).returncode == 0
    malformed = _STAGE_RECEIPT.replace(
        '"package_result_count":1', '"package_results_count":1'
    )
    assert malformed != _STAGE_RECEIPT
    assert parse(malformed).returncode == 75
