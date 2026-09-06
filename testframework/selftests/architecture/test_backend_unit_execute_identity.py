from __future__ import annotations

import subprocess


_EXECUTE_PATH = "/opt/hub/testframework/docker/backend_unit/execute.sh"
_SHELL_DRIVER = """\
. "$0"
_execute_identity_is_exact "$@"
"""
_TIMEOUT_SECONDS = 5.0
_BASE_FIELDS = ("11", "22", "10001", "10001", "600")
_CANONICAL_NUMBERS = (
    "0",
    "1",
    "2",
    "3",
    "4",
    "5",
    "6",
    "7",
    "8",
    "9",
    "10",
    "9876543210",
    "1234567890" * 16,
)
_NONCANONICAL_NUMBERS = (
    "",
    "00",
    "01",
    "12x",
    "1x2",
    "x12",
    "-1",
    "+1",
    "1.0",
    "1e2",
    " 1",
    "1 ",
    "1 2",
    "\t1",
    "1\t",
    "1\t2",
    "\n1",
    "1\n",
    "1\n2",
)


def _identity(fields: tuple[str, str, str, str, str], file_type: str = "regular file") -> str:
    return ":".join((*fields, file_type))


def _run_identity(*arguments: str) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["/bin/sh", "-c", _SHELL_DRIVER, _EXECUTE_PATH, *arguments],
        check=False,
        capture_output=True,
        text=True,
        timeout=_TIMEOUT_SECONDS,
    )


def _assert_result(completed: subprocess.CompletedProcess[str], status: int) -> None:
    assert completed.returncode == status
    assert completed.stdout == ""
    assert completed.stderr == ""


def test_execute_identity_accepts_canonical_unsigned_decimals_in_every_field() -> None:
    for position in range(5):
        for numeric in _CANONICAL_NUMBERS:
            fields = list(_BASE_FIELDS)
            fields[position] = numeric
            exact_fields = tuple(fields)
            completed = _run_identity(
                _identity(exact_fields),
                exact_fields[2],
                exact_fields[3],
                exact_fields[4],
                "regular file",
            )

            _assert_result(completed, 0)


def test_execute_identity_rejects_noncanonical_unsigned_decimals_in_every_field() -> None:
    for position in range(5):
        for numeric in _NONCANONICAL_NUMBERS:
            fields = list(_BASE_FIELDS)
            fields[position] = numeric
            exact_fields = tuple(fields)
            completed = _run_identity(
                _identity(exact_fields),
                exact_fields[2],
                exact_fields[3],
                exact_fields[4],
                "regular file",
            )

            _assert_result(completed, 75)


def test_execute_identity_preserves_arity_shape_type_and_ownership_denials() -> None:
    valid_identity = _identity(_BASE_FIELDS)
    cases = (
        ((), 64),
        ((valid_identity, "10001", "10001", "600"), 64),
        ((valid_identity, "10001", "10001", "600", "regular file", "extra"), 64),
        (("11:22:10001:10001:600", "10001", "10001", "600", "regular file"), 75),
        ((f"{valid_identity}:extra", "10001", "10001", "600", "regular file"), 75),
        (("11:22:10001:10001:600:", "10001", "10001", "600", "regular file"), 75),
        (("11:22:10001:10001:600:regular:file", "10001", "10001", "600", "regular file"), 75),
        ((valid_identity, "10002", "10001", "600", "regular file"), 75),
        ((valid_identity, "10001", "10002", "600", "regular file"), 75),
        ((valid_identity, "10001", "10001", "601", "regular file"), 75),
        ((valid_identity, "10001", "10001", "600", "directory"), 75),
    )

    for arguments, status in cases:
        _assert_result(_run_identity(*arguments), status)
