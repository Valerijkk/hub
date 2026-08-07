from __future__ import annotations

import stat
import sys
import tempfile
import unittest
from pathlib import Path
from types import SimpleNamespace
from unittest.mock import patch


SCRIPT_DIR = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SCRIPT_DIR))

from safe_filesystem import (  # noqa: E402
    FileSafetyError,
    SafeTree,
    TextPolicy,
    is_link_like,
)


def text_policy(
    *, max_file_bytes: int = 64, max_total_bytes: int = 256, max_file_count: int = 8
) -> TextPolicy:
    return TextPolicy(
        allowed_text_extensions=(".md", ".txt"),
        allowed_text_names=("LICENSE",),
        max_file_bytes=max_file_bytes,
        max_total_bytes=max_total_bytes,
        max_file_count=max_file_count,
    )


class WindowsReparseDetectionTests(unittest.TestCase):
    def test_mocked_windows_reparse_attribute_is_rejected_without_privilege(self) -> None:
        marker = 0x0400
        metadata = SimpleNamespace(
            st_mode=stat.S_IFREG | 0o600,
            st_file_attributes=marker,
        )
        with patch("safe_filesystem.os.lstat", return_value=metadata):
            with patch.object(
                stat,
                "FILE_ATTRIBUTE_REPARSE_POINT",
                marker,
                create=True,
            ):
                self.assertTrue(is_link_like(Path("mocked"), windows=True))

    def test_mocked_windows_normal_attribute_is_not_reparse(self) -> None:
        marker = 0x0400
        metadata = SimpleNamespace(
            st_mode=stat.S_IFREG | 0o600,
            st_file_attributes=0,
        )
        with patch("safe_filesystem.os.lstat", return_value=metadata):
            with patch.object(
                stat,
                "FILE_ATTRIBUTE_REPARSE_POINT",
                marker,
                create=True,
            ):
                self.assertFalse(is_link_like(Path("mocked"), windows=True))

    def test_missing_windows_attributes_fail_closed(self) -> None:
        metadata = SimpleNamespace(st_mode=stat.S_IFREG | 0o600)
        with patch("safe_filesystem.os.lstat", return_value=metadata):
            with patch.object(
                stat,
                "FILE_ATTRIBUTE_REPARSE_POINT",
                0x0400,
                create=True,
            ):
                with self.assertRaisesRegex(
                    FileSafetyError, "cannot inspect Windows reparse attributes"
                ):
                    is_link_like(Path("mocked"), windows=True)


class SafeTreeSentinelTests(unittest.TestCase):
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

    def test_outside_symlink_canary_is_rejected_before_any_open(self) -> None:
        canary = self.base / "outside-canary.md"
        canary.write_text("outside marker\n", encoding="utf-8")
        (self.root / "linked.md").symlink_to(canary)

        with patch("safe_filesystem.os.open") as opened:
            with self.assertRaisesRegex(FileSafetyError, "symbolic link or reparse"):
                SafeTree(self.root, text_policy())
        opened.assert_not_called()

    def test_file_swap_after_enumeration_is_rejected_before_read(self) -> None:
        victim = self.write_bytes("victim.md", b"original\n")
        tree = SafeTree(self.root, text_policy())
        replacement = self.base / "replacement.md"
        replacement.write_bytes(b"replacement\n")
        victim.unlink()
        victim.symlink_to(replacement)

        with patch("safe_filesystem.os.open") as opened:
            with self.assertRaisesRegex(FileSafetyError, "became a symbolic link"):
                tree.inspect_text(victim)
        opened.assert_not_called()

    def test_regular_file_identity_swap_is_rejected(self) -> None:
        victim = self.write_bytes("victim.md", b"first\n")
        tree = SafeTree(self.root, text_policy())
        victim.unlink()
        victim.write_bytes(b"second\n")

        with self.assertRaisesRegex(FileSafetyError, "identity changed"):
            tree.inspect_text(victim)

    def test_outside_lookup_is_rejected_without_filesystem_probe(self) -> None:
        self.write_bytes("inside.md", b"inside\n")
        tree = SafeTree(self.root, text_policy())

        with self.assertRaisesRegex(FileSafetyError, "escapes"):
            tree.exists(self.base / "outside.md")

    def test_invalid_utf8_binary_and_oversize_are_rejected(self) -> None:
        invalid = self.write_bytes("invalid.md", b"\xff\xfe")
        binary = self.write_bytes("binary.txt", b"before\x00after")
        large = self.write_bytes("large.md", b"123456789")
        tree = SafeTree(
            self.root,
            text_policy(max_file_bytes=8, max_total_bytes=64),
        )

        with self.assertRaisesRegex(FileSafetyError, "valid UTF-8"):
            tree.inspect_text(invalid)
        with self.assertRaisesRegex(FileSafetyError, "binary/control"):
            tree.inspect_text(binary)
        with self.assertRaisesRegex(FileSafetyError, "byte limit"):
            tree.inspect_text(large)

    def test_file_count_and_total_bytes_are_hard_bounds(self) -> None:
        self.write_bytes("a.md", b"1234")
        self.write_bytes("b.md", b"5678")

        with self.assertRaisesRegex(FileSafetyError, "max file count"):
            SafeTree(self.root, text_policy(max_file_count=1))
        with self.assertRaisesRegex(FileSafetyError, "max total bytes"):
            SafeTree(self.root, text_policy(max_total_bytes=7))

    def test_inaccessible_enumeration_fails_closed(self) -> None:
        with patch("safe_filesystem.os.scandir", side_effect=PermissionError):
            with self.assertRaisesRegex(FileSafetyError, "cannot enumerate"):
                SafeTree(self.root, text_policy())

    def test_inspection_cache_reuses_verified_text(self) -> None:
        path = self.write_bytes("cached.md", b"cached\n")
        tree = SafeTree(self.root, text_policy())

        first = tree.inspect_text(path)
        second = tree.inspect_text(path)

        self.assertIs(first, second)
        self.assertEqual(1, len(tree.inspection_cache))


if __name__ == "__main__":
    unittest.main()
