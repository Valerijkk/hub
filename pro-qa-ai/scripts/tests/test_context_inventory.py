from __future__ import annotations

import hashlib
import io
import json
import sys
import tempfile
import unittest
from contextlib import redirect_stderr, redirect_stdout
from pathlib import Path
from unittest.mock import patch


SCRIPT_DIR = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SCRIPT_DIR))

from context_inventory import (  # noqa: E402
    InventoryError,
    build_inventory,
    load_config,
    main,
    write_inventory,
)


class ContextInventoryTests(unittest.TestCase):
    def setUp(self) -> None:
        self._temporary = tempfile.TemporaryDirectory()
        self.base = Path(self._temporary.name)
        self.project = self.base / "project"
        self.tests = self.base / "tests"
        self.project.mkdir()
        self.tests.mkdir()
        self.config = json.loads(
            json.dumps(load_config(SCRIPT_DIR / "context_inventory_config.json"))
        )

    def tearDown(self) -> None:
        self._temporary.cleanup()

    @staticmethod
    def write(root: Path, relative: str, data: bytes) -> Path:
        path = root.joinpath(*relative.split("/"))
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(data)
        return path

    def test_inventory_hashes_only_safe_files_from_explicit_roots(self) -> None:
        project_data = b"# Project\n"
        test_data = b"def test_example():\n    pass\n"
        self.write(self.project, "README.md", project_data)
        self.write(self.tests, "test_example.py", test_data)
        self.write(self.project, "asset.png", b"\x89PNG")
        self.write(self.project, "binary.txt", b"before\x00after")
        outside = self.base / "outside.md"
        outside.write_text("not declared\n", encoding="utf-8")

        inventory = build_inventory(self.project, [self.tests], self.config)

        self.assertEqual(2, inventory["file_count"])
        self.assertEqual(
            {
                ("project_root", "README.md"),
                ("test_root_01", "test_example.py"),
            },
            {(item["root_id"], item["path"]) for item in inventory["files"]},
        )
        self.assertEqual(
            hashlib.sha256(project_data).hexdigest(),
            next(
                item["sha256"]
                for item in inventory["files"]
                if item["path"] == "README.md"
            ),
        )
        reasons = {item["reason"] for item in inventory["exclusions"]}
        self.assertIn("unapproved-file-type", reasons)
        self.assertIn("binary-or-non-utf8", reasons)
        self.assertNotIn("outside.md", json.dumps(inventory))

    def test_inventory_limits_are_enforced(self) -> None:
        self.write(self.project, "a.md", b"123456")
        self.write(self.tests, "b.md", b"abcdef")
        self.config["file_policy"]["max_file_bytes"] = 10
        self.config["file_policy"]["max_total_bytes"] = 10

        with self.assertRaisesRegex(InventoryError, "max total bytes"):
            build_inventory(self.project, [self.tests], self.config)

        self.config["file_policy"]["max_total_bytes"] = 20
        self.config["file_policy"]["max_file_count"] = 1
        with self.assertRaisesRegex(InventoryError, "max file count"):
            build_inventory(self.project, [self.tests], self.config)

    def test_reparse_like_entry_fails_inventory(self) -> None:
        linked = self.write(self.project, "linked.md", b"safe target text\n")
        original = __import__("context_inventory").is_link_like

        def fake_is_link_like(path: Path) -> bool:
            if path == linked:
                return True
            return original(path)

        with patch("context_inventory.is_link_like", side_effect=fake_is_link_like):
            with self.assertRaisesRegex(InventoryError, "reparse point"):
                build_inventory(self.project, [self.tests], self.config)

    def test_output_is_exclusive_and_never_overwritten(self) -> None:
        self.write(self.project, "README.md", b"# Project\n")
        inventory = build_inventory(self.project, [self.tests], self.config)
        output = self.base / "inventory.json"
        output.write_text("preserve\n", encoding="utf-8")

        with self.assertRaisesRegex(InventoryError, "already exists"):
            write_inventory(output, inventory, [self.project, self.tests])

        self.assertEqual("preserve\n", output.read_text(encoding="utf-8"))

    def test_cli_writes_inventory_with_compact_default_output(self) -> None:
        self.write(self.project, "README.md", b"# Project\n")
        output = self.base / "inventory.json"
        config_path = self.base / "config.json"
        config_path.write_text(json.dumps(self.config), encoding="utf-8")
        stdout = io.StringIO()
        stderr = io.StringIO()
        with redirect_stdout(stdout), redirect_stderr(stderr):
            status = main(
                [
                    "--project-root",
                    str(self.project),
                    "--test-root",
                    str(self.tests),
                    "--config",
                    str(config_path),
                    "--output",
                    str(output),
                ]
            )

        self.assertEqual(0, status, stderr.getvalue())
        self.assertIn("Context inventory: PASS", stdout.getvalue())
        self.assertNotIn("README.md", stdout.getvalue())
        self.assertTrue(output.is_file())


if __name__ == "__main__":
    unittest.main()
