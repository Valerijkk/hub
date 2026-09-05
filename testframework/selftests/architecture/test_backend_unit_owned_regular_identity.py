from __future__ import annotations

import os
from pathlib import Path
import subprocess


_RUNTIME_PATH = "/opt/hub/testframework/docker/backend_unit/runtime.sh"
_SHELL_DRIVER = """\
. "$0"
_owned_regular_identity "$@"
"""


def _run_owned_regular_identity(*arguments: str) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["/bin/sh", "-c", _SHELL_DRIVER, _RUNTIME_PATH, *arguments],
        check=False,
        capture_output=True,
        text=True,
        timeout=5,
    )


def _expected_identity(path: Path) -> str:
    metadata = os.stat(path)
    assert metadata.st_uid == 10001
    assert metadata.st_gid == 10001
    return f"{metadata.st_dev}:{metadata.st_ino}:10001:10001:600:regular file\n"


def test_owned_regular_identity_canonicalizes_empty_file_across_write_and_truncate(
    tmp_path: Path,
) -> None:
    path = tmp_path / "owner-receipt"
    path.write_bytes(b"")
    path.chmod(0o600)
    expected = _expected_identity(path)

    empty = _run_owned_regular_identity(str(path))
    path.write_bytes(b"owner receipt")
    nonempty = _run_owned_regular_identity(str(path))
    path.write_bytes(b"")
    truncated = _run_owned_regular_identity(str(path))

    assert (empty.returncode, nonempty.returncode, truncated.returncode) == (0, 0, 0)
    assert (empty.stdout, nonempty.stdout, truncated.stdout) == (
        expected,
        expected,
        expected,
    )
    assert (empty.stderr, nonempty.stderr, truncated.stderr) == ("", "", "")


def test_owned_regular_identity_accepts_nonempty_owned_file(tmp_path: Path) -> None:
    path = tmp_path / "owner-receipt"
    path.write_bytes(b"owner receipt")
    path.chmod(0o600)

    completed = _run_owned_regular_identity(str(path))

    assert completed.returncode == 0, completed.stderr
    assert completed.stdout == _expected_identity(path)
    assert completed.stderr == ""


def test_owned_regular_identity_rejects_unowned_file_shapes_and_wrong_arity(
    tmp_path: Path,
) -> None:
    wrong_mode = tmp_path / "wrong-mode"
    wrong_mode.write_bytes(b"owner receipt")
    wrong_mode.chmod(0o640)

    regular_file = tmp_path / "regular"
    regular_file.write_bytes(b"owner receipt")
    regular_file.chmod(0o600)
    symlink = tmp_path / "symlink"
    symlink.symlink_to(regular_file)

    directory = tmp_path / "directory"
    directory.mkdir()
    missing = tmp_path / "missing"

    rejected = (
        _run_owned_regular_identity(str(wrong_mode)),
        _run_owned_regular_identity(str(symlink)),
        _run_owned_regular_identity(str(directory)),
        _run_owned_regular_identity(str(missing)),
        _run_owned_regular_identity(),
    )

    assert tuple(result.returncode for result in rejected) == (75, 75, 75, 75, 64)
    assert all(result.stdout == "" and result.stderr == "" for result in rejected)
