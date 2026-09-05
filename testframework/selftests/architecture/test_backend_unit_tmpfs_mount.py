from __future__ import annotations

from pathlib import Path
import re
import subprocess

import pytest


_REPO_ROOT = Path(__file__).resolve().parents[3]
_RUNTIME_PATH = _REPO_ROOT / "testframework/docker/backend_unit/runtime.sh"
_FUNCTION_PATTERN = re.compile(
    r"^_assert_tmpfs_mount\(\) \{\n(?P<body>.*?)^}\n(?=\nassert_owned_tmpfs\(\) \{)",
    re.MULTILINE | re.DOTALL,
)
_AWK_START = '  tmpfs_mount_record="$(awk -v target="$tmpfs_path" \'\n'
_AWK_END = "\n  ' /proc/self/mountinfo)"
_TARGET = "/var/cache/hub-go"
_CANONICAL_TMPFS_RECORD = (
    "45 28 0:43 / /var/cache/hub-go rw,nosuid,nodev,noexec - tmpfs tmpfs "
    "rw,size=805306368,uid=10001,gid=10001,mode=1770\n"
)
_EXPECTED_TMPFS_RECORD = (
    "tmpfs:rw,nosuid,nodev,noexec,rw,size=805306368,uid=10001,gid=10001,mode=1770\n"
)


def _extract_tmpfs_mount_awk_program() -> str:
    runtime = _RUNTIME_PATH.read_text(encoding="utf-8")
    function_matches = list(_FUNCTION_PATTERN.finditer(runtime))
    if len(function_matches) != 1:
        raise AssertionError("expected one _assert_tmpfs_mount function")

    function_body = function_matches[0]["body"]
    if function_body.count(_AWK_START) != 1:
        raise AssertionError("expected one tmpfs mount awk invocation")
    if function_body.count(_AWK_END) != 1:
        raise AssertionError("expected one tmpfs mount awk terminator")

    return function_body.split(_AWK_START, 1)[1].split(_AWK_END, 1)[0]


def _run_tmpfs_mount_program(mountinfo: str) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["awk", "-v", f"target={_TARGET}", _extract_tmpfs_mount_awk_program()],
        input=mountinfo,
        check=False,
        capture_output=True,
        text=True,
        timeout=5,
    )


@pytest.mark.parametrize(
    ("mountinfo", "expected"),
    [
        (_CANONICAL_TMPFS_RECORD, _EXPECTED_TMPFS_RECORD),
        (
            _CANONICAL_TMPFS_RECORD.replace(
                "noexec -", "noexec shared:13 master:1 -", 1
            ),
            _EXPECTED_TMPFS_RECORD,
        ),
    ],
    ids=("canonical", "optional-mount-fields"),
)
def test_tmpfs_mount_awk_program_accepts_canonical_and_optional_mount_fields(
    mountinfo: str, expected: str
) -> None:
    completed = _run_tmpfs_mount_program(mountinfo)

    assert completed.returncode == 0, completed.stderr
    assert completed.stdout == expected
    assert completed.stderr == ""


@pytest.mark.parametrize(
    "mountinfo",
    [
        _CANONICAL_TMPFS_RECORD.replace(_TARGET, "/var/cache/other", 1),
        _CANONICAL_TMPFS_RECORD
        + _CANONICAL_TMPFS_RECORD.replace("45 28", "46 28", 1),
        _CANONICAL_TMPFS_RECORD.replace("tmpfs tmpfs", "ext4 /dev/sda1", 1),
        _CANONICAL_TMPFS_RECORD.replace(
            "- tmpfs tmpfs rw,size=805306368,uid=10001,gid=10001,mode=1770", "shared:13 master:1", 1
        ),
        _CANONICAL_TMPFS_RECORD.replace(
            "- tmpfs tmpfs rw,size=805306368,uid=10001,gid=10001,mode=1770", "- tmpfs", 1
        ),
    ],
    ids=(
        "missing-target",
        "duplicate-matches",
        "non-tmpfs",
        "missing-separator",
        "incomplete-after-separator",
    ),
)
def test_tmpfs_mount_awk_program_rejects_malformed_records(mountinfo: str) -> None:
    completed = _run_tmpfs_mount_program(mountinfo)

    assert completed.returncode == 75
    assert completed.stdout == ""
