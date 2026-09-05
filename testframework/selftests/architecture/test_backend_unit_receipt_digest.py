from __future__ import annotations

import hashlib
import os
import subprocess
from pathlib import Path

import pytest


_RECEIPT_PATH = "/opt/hub/testframework/docker/backend_unit/receipt.sh"
_SHELL_DRIVER = """\
if [ "${HUB_RECEIPT_TEST_SHA256SUM_OUTPUT+set}" = set ]; then
  sha256sum() {
    printf '%s' "$HUB_RECEIPT_TEST_SHA256SUM_OUTPUT"
  }
fi
. "$0"
{function} "$@"
"""


def _receipt_digest(
    function: str,
    *arguments: str,
    sha256sum_output: str | None = None,
) -> subprocess.CompletedProcess[str]:
    environment = None
    if sha256sum_output is not None:
        environment = os.environ.copy()
        environment["HUB_RECEIPT_TEST_SHA256SUM_OUTPUT"] = sha256sum_output

    return subprocess.run(
        [
            "/bin/sh",
            "-c",
            _SHELL_DRIVER.replace("{function}", function),
            _RECEIPT_PATH,
            *arguments,
        ],
        check=False,
        capture_output=True,
        env=environment,
        text=True,
        timeout=5,
    )


@pytest.mark.parametrize("text", ["", "abc"])
def test_receipt_sha256_text_matches_actual_sha256(text: str) -> None:
    result = _receipt_digest("_receipt_sha256_text", text)

    assert result.returncode == 0, result.stderr
    assert result.stdout == f"sha256:{hashlib.sha256(text.encode()).hexdigest()}\n"
    assert result.stderr == ""


def test_receipt_sha256_file_matches_actual_sha256(tmp_path: Path) -> None:
    payload = tmp_path / "payload"
    payload.write_bytes(b"receipt payload\n")

    result = _receipt_digest("_receipt_sha256_file", str(payload))

    assert result.returncode == 0, result.stderr
    assert result.stdout == f"sha256:{hashlib.sha256(payload.read_bytes()).hexdigest()}\n"
    assert result.stderr == ""


def test_receipt_sha256_file_rejects_missing_path(tmp_path: Path) -> None:
    result = _receipt_digest("_receipt_sha256_file", str(tmp_path / "missing"))

    assert result.returncode == 75
    assert result.stdout == ""


def test_receipt_sha256_file_rejects_symbolic_link(tmp_path: Path) -> None:
    payload = tmp_path / "payload"
    payload.write_bytes(b"receipt payload\n")
    symbolic_link = tmp_path / "payload-link"
    symbolic_link.symlink_to(payload)

    result = _receipt_digest("_receipt_sha256_file", str(symbolic_link))

    assert result.returncode == 75
    assert result.stdout == ""


@pytest.mark.parametrize(
    ("function", "arguments"),
    [
        ("_receipt_sha256_file", ()),
        ("_receipt_sha256_file", ("unused", "extra")),
        ("_receipt_sha256_text", ()),
        ("_receipt_sha256_text", ("unused", "extra")),
    ],
)
def test_receipt_sha256_helpers_reject_wrong_arity(
    function: str, arguments: tuple[str, ...]
) -> None:
    result = _receipt_digest(function, *arguments)

    assert result.returncode == 64
    assert result.stdout == ""


@pytest.mark.parametrize("function", ["_receipt_sha256_file", "_receipt_sha256_text"])
@pytest.mark.parametrize(
    "sha256sum_output",
    [
        "a" * 32 + "  -\n",
        "a" * 63 + "  -\n",
        "a" * 65 + "  -\n",
        "A" * 64 + "  -\n",
        "a" * 63 + "g  -\n",
        "a" * 32 + "\n" + "a" * 32 + "  -\n",
        "",
    ],
    ids=["32", "63", "65", "uppercase", "nonhex", "split-newline", "empty"],
)
def test_receipt_sha256_helpers_reject_malformed_digest_producer_output(
    function: str, sha256sum_output: str, tmp_path: Path
) -> None:
    arguments = ("abc",)
    if function == "_receipt_sha256_file":
        payload = tmp_path / "payload"
        payload.write_bytes(b"abc")
        arguments = (str(payload),)

    result = _receipt_digest(function, *arguments, sha256sum_output=sha256sum_output)

    assert result.returncode == 75
    assert result.stdout == ""
