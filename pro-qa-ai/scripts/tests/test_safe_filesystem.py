from __future__ import annotations

import stat
import sys
import unittest
from pathlib import Path
from types import SimpleNamespace
from unittest.mock import patch


SCRIPT_DIR = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SCRIPT_DIR))

from safe_filesystem import FileSafetyError, is_link_like  # noqa: E402


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


if __name__ == "__main__":
    unittest.main()
