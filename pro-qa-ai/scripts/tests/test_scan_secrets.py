from __future__ import annotations

import io
import sys
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch


SCRIPT_DIR = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SCRIPT_DIR))

import scan_secrets  # noqa: E402
from safe_filesystem import FileSafetyError, SafeTree, TextPolicy  # noqa: E402


class SecretScannerSafeIOSentinels(unittest.TestCase):
    def setUp(self) -> None:
        self._temporary = tempfile.TemporaryDirectory()
        self.base = Path(self._temporary.name)
        self.root = self.base / "root"
        self.root.mkdir()

    def tearDown(self) -> None:
        self._temporary.cleanup()

    def write_bytes(self, relative: str, content: bytes) -> Path:
        path = self.root / relative
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(content)
        return path

    def test_sensitive_names_are_never_opened_automatically(self) -> None:
        self.write_bytes("notes.md", b"ordinary text\n")
        self.write_bytes(".env", b"\x00must-not-open")
        self.write_bytes(".env.production", b"\x00must-not-open")
        self.write_bytes("credentials.json", b"\x00must-not-open")
        self.write_bytes("secrets.yaml", b"\x00must-not-open")
        self.write_bytes("server.pem", b"\x00must-not-open")
        self.write_bytes("client.key", b"\x00must-not-open")
        self.write_bytes("truststore.jks", b"\x00must-not-open")

        self.assertEqual([], scan_secrets.scan_worktree(self.root))

    def test_invalid_utf8_and_binary_regular_text_fail_closed(self) -> None:
        invalid = self.write_bytes("invalid.md", b"\xff\xfe")
        tree = SafeTree(
            self.root,
            scan_secrets.SCAN_POLICY,
            exclude=scan_secrets._skip_tree_entry,
        )
        with self.assertRaisesRegex(FileSafetyError, "valid UTF-8"):
            scan_secrets.scan_worktree(self.root, tree=tree)

        invalid.write_bytes(b"before\x00after")
        tree = SafeTree(
            self.root,
            scan_secrets.SCAN_POLICY,
            exclude=scan_secrets._skip_tree_entry,
        )
        with self.assertRaisesRegex(FileSafetyError, "binary/control"):
            scan_secrets.scan_worktree(self.root, tree=tree)

    def test_scanner_enforces_count_and_total_limits(self) -> None:
        self.write_bytes("a.md", b"1234")
        self.write_bytes("b.md", b"5678")
        count_policy = TextPolicy((".md",), (), 8, 16, 1)
        total_policy = TextPolicy((".md",), (), 8, 7, 2)

        with self.assertRaisesRegex(FileSafetyError, "max file count"):
            SafeTree(self.root, count_policy)
        with self.assertRaisesRegex(FileSafetyError, "max total bytes"):
            SafeTree(self.root, total_policy)

    def test_swap_to_outside_symlink_is_rejected_without_reading_canary(self) -> None:
        victim = self.write_bytes("victim.md", b"ordinary\n")
        canary = self.base / "outside.md"
        canary.write_bytes(b"outside marker\n")
        tree = SafeTree(
            self.root,
            scan_secrets.SCAN_POLICY,
            exclude=scan_secrets._skip_tree_entry,
        )
        victim.unlink()
        victim.symlink_to(canary)

        with patch("safe_filesystem.os.open") as opened:
            with self.assertRaisesRegex(FileSafetyError, "became a symbolic link"):
                scan_secrets.scan_worktree(self.root, tree=tree)
        opened.assert_not_called()

    def test_findings_and_failure_diagnostics_never_expose_values(self) -> None:
        value = "AK" + "IA1234567890ABCDEF"
        self.write_bytes("notes.md", ("value = " + value + "\n").encode("utf-8"))

        findings = scan_secrets.scan_worktree(self.root)

        self.assertEqual(1, len(findings))
        self.assertNotIn(value, findings[0])

        error = io.StringIO()
        with patch("scan_secrets.SafeTree", side_effect=FileSafetyError("denied")):
            with patch("sys.stderr", error):
                exit_code = scan_secrets.main(["--root", str(self.root)])
        self.assertEqual(2, exit_code)
        self.assertIn("SECRETSCAN001", error.getvalue())
        self.assertNotIn(value, error.getvalue())


if __name__ == "__main__":
    unittest.main()
