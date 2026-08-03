from __future__ import annotations

import hashlib
import io
import json
import sys
import tempfile
import unittest
from contextlib import redirect_stderr, redirect_stdout
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SCRIPT_DIR))

from package_portable import (  # noqa: E402
    PackageError,
    build_package_plan,
    inventory_bytes,
    main,
    package_directory,
    parse_policy,
    render_summary,
)


class PortablePackagerTests(unittest.TestCase):
    def setUp(self) -> None:
        self._temporary = tempfile.TemporaryDirectory()
        self.base = Path(self._temporary.name)
        self.root = self.base / "source"
        self.root.mkdir()
        self.output = self.base / "output"
        self.config = json.loads(
            (SCRIPT_DIR / "package_manifest.json").read_text(encoding="utf-8")
        )
        self.config["required_paths"] = []
        self.policy = parse_policy(self.config)
        self.config_path = self.base / "package_manifest.json"
        self.config_path.write_text(
            json.dumps(self.config, indent=2) + "\n",
            encoding="utf-8",
        )

    def tearDown(self) -> None:
        self._temporary.cleanup()

    def write(self, relative: str, content: str = "portable\n") -> Path:
        path = self.root.joinpath(*relative.split("/"))
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(content.encode("utf-8"))
        return path

    def selected_paths(self) -> set[str]:
        return {
            entry.relative_path
            for entry in build_package_plan(self.root, self.policy).entries
        }

    def test_profiles_and_run_evidence_are_excluded_but_templates_ship(self) -> None:
        self.write("agents/core.md")
        self.write("context/templates/PROJECT.md")
        self.write("context/projects/README.md")
        self.write("context/projects/hub/PROJECT.md", "Hub-only facts\n")
        self.write("evals/run-record.template.json", "{}\n")
        self.write("standards/patterns/README.md")
        self.write("standards/patterns/approved/README.md")
        self.write("standards/patterns/approved/hub-card.md", "project-only\n")
        self.write("standards/patterns/candidates/README.md")
        self.write("standards/patterns/candidates/draft.md", "project-only\n")
        self.write("standards/patterns/project-card.md", "project-only\n")
        self.write("standards/team-review/README.md")
        self.write("standards/team-review/manual-root.md", "private feedback\n")
        self.write("standards/team-review/threads/inbox/README.md")
        self.write("standards/team-review/threads/inbox/manual.md", "private feedback\n")
        self.write("standards/team-review/threads/approved/README.md")
        self.write("agents/local.log", "local run evidence\n")

        selected = self.selected_paths()

        self.assertIn("agents/core.md", selected)
        self.assertIn("context/templates/PROJECT.md", selected)
        self.assertIn("context/projects/README.md", selected)
        self.assertIn("standards/patterns/approved/README.md", selected)
        self.assertIn("standards/patterns/candidates/README.md", selected)
        self.assertIn("standards/team-review/README.md", selected)
        self.assertIn("standards/team-review/threads/inbox/README.md", selected)
        self.assertIn("standards/team-review/threads/approved/README.md", selected)
        self.assertNotIn("context/projects/hub/PROJECT.md", selected)
        self.assertNotIn("standards/patterns/approved/hub-card.md", selected)
        self.assertNotIn("standards/patterns/candidates/draft.md", selected)
        self.assertNotIn("standards/patterns/project-card.md", selected)
        self.assertNotIn("standards/team-review/manual-root.md", selected)
        self.assertNotIn("standards/team-review/threads/inbox/manual.md", selected)
        self.assertNotIn("agents/local.log", selected)

    def test_dry_run_selects_and_hashes_without_writing(self) -> None:
        self.write("agents/core.md", "stable payload\n")

        stdout = io.StringIO()
        stderr = io.StringIO()
        with redirect_stdout(stdout), redirect_stderr(stderr):
            status = main(
                [
                    "--root",
                    str(self.root),
                    "--output",
                    str(self.output),
                    "--dry-run",
                ],
                config_path=self.config_path,
            )

        self.assertEqual(0, status, stderr.getvalue())
        self.assertIn("Mode: dry-run", stdout.getvalue())
        self.assertIn("File count: 1", stdout.getvalue())
        self.assertRegex(stdout.getvalue(), r"Manifest SHA-256: [0-9a-f]{64}")
        self.assertNotIn("agents/core.md", stdout.getvalue())
        self.assertFalse(self.output.exists())

    def test_manifest_traversal_is_rejected(self) -> None:
        unsafe = json.loads(json.dumps(self.config))
        unsafe["include_globs"].append("../outside/**")

        with self.assertRaises(PackageError) as raised:
            parse_policy(unsafe)

        self.assertNotIn("outside payload", str(raised.exception))

    def test_output_inside_source_root_is_rejected(self) -> None:
        self.write("agents/core.md")

        with self.assertRaisesRegex(PackageError, "outside the source root"):
            package_directory(
                self.root,
                self.root / "dist",
                self.policy,
                dry_run=True,
            )

    def test_selected_symlink_is_refused_without_reading_target(self) -> None:
        outside = self.base / "outside.txt"
        secret = "AK" + "IA1234567890ABCDEF"
        outside.write_text(secret, encoding="utf-8")
        link = self.root / "agents" / "linked.md"
        link.parent.mkdir(parents=True)
        try:
            link.symlink_to(outside)
        except (NotImplementedError, OSError):
            self.skipTest("symbolic links are unavailable in this environment")

        with self.assertRaises(PackageError) as raised:
            build_package_plan(self.root, self.policy)

        self.assertNotIn(secret, str(raised.exception))

    def test_existing_nonempty_output_is_never_replaced(self) -> None:
        self.write("agents/core.md", "new payload\n")
        self.output.mkdir()
        secret = "AK" + "IA1234567890ABCDEF"
        (self.output / "stale.txt").write_text(secret, encoding="utf-8")

        stdout = io.StringIO()
        stderr = io.StringIO()
        with redirect_stdout(stdout), redirect_stderr(stderr):
            refused = main(
                ["--root", str(self.root), "--output", str(self.output)],
                config_path=self.config_path,
            )

        self.assertEqual(2, refused)
        self.assertIn("choose a new empty output path", stderr.getvalue())
        self.assertNotIn(secret, stderr.getvalue())
        self.assertTrue((self.output / "stale.txt").exists())

        parser_stderr = io.StringIO()
        with redirect_stderr(parser_stderr):
            with self.assertRaises(SystemExit):
                main(
                    [
                        "--root",
                        str(self.root),
                        "--output",
                        str(self.output),
                        "--replace" + "-output",
                    ],
                    config_path=self.config_path,
                )

        self.assertTrue((self.output / "stale.txt").exists())

    def test_existing_empty_output_is_allowed(self) -> None:
        self.write("agents/core.md", "new payload\n")
        self.output.mkdir()

        package_directory(self.root, self.output, self.policy)

        self.assertEqual(
            "new payload\n",
            (self.output / "agents" / "core.md").read_text(encoding="utf-8"),
        )
        self.assertTrue((self.output / "PACKAGE-INVENTORY.json").is_file())

    def test_inventory_and_hashes_are_deterministic(self) -> None:
        first_payload = "alpha\n"
        second_payload = "zeta\n"
        self.write("agents/z.md", second_payload)
        self.write("agents/a.md", first_payload)

        first_output = self.base / "package-one"
        second_output = self.base / "package-two"
        first_plan = package_directory(self.root, first_output, self.policy)
        second_plan = package_directory(self.root, second_output, self.policy)

        first_inventory = (first_output / "PACKAGE-INVENTORY.json").read_bytes()
        second_inventory = (second_output / "PACKAGE-INVENTORY.json").read_bytes()
        decoded = json.loads(first_inventory)

        self.assertEqual(first_plan.manifest_sha256, second_plan.manifest_sha256)
        self.assertEqual(inventory_bytes(first_plan), inventory_bytes(second_plan))
        self.assertEqual(first_inventory, second_inventory)
        self.assertEqual(2, decoded["file_count"])
        self.assertEqual(
            ["agents/a.md", "agents/z.md"],
            [entry["path"] for entry in decoded["files"]],
        )
        self.assertEqual(
            hashlib.sha256(first_payload.encode("utf-8")).hexdigest(),
            decoded["files"][0]["sha256"],
        )
        self.assertEqual(first_plan.manifest_sha256, decoded["manifest_sha256"])

    def test_rejects_unapproved_binary_and_size_limit_breaches(self) -> None:
        self.write("agents/payload.bin", "text despite binary extension\n")
        with self.assertRaisesRegex(PackageError, "unapproved text file type"):
            build_package_plan(self.root, self.policy)

        (self.root / "agents" / "payload.bin").unlink()
        self.write("agents/nul.md", "before\x00after")
        with self.assertRaisesRegex(PackageError, "binary/NUL"):
            build_package_plan(self.root, self.policy)

        (self.root / "agents" / "nul.md").unlink()
        limited = json.loads(json.dumps(self.config))
        limited["file_policy"]["max_file_bytes"] = 8
        limited["file_policy"]["max_total_bytes"] = 16
        self.write("agents/large.md", "123456789")
        with self.assertRaisesRegex(PackageError, "byte limit"):
            build_package_plan(self.root, parse_policy(limited))

    def test_rejects_total_bytes_and_file_count_breaches(self) -> None:
        self.write("agents/a.md", "123456")
        self.write("agents/b.md", "abcdef")

        total_limited = json.loads(json.dumps(self.config))
        total_limited["file_policy"]["max_file_bytes"] = 10
        total_limited["file_policy"]["max_total_bytes"] = 10
        with self.assertRaisesRegex(PackageError, "max total bytes"):
            build_package_plan(self.root, parse_policy(total_limited))

        count_limited = json.loads(json.dumps(self.config))
        count_limited["file_policy"]["max_file_count"] = 1
        with self.assertRaisesRegex(PackageError, "max file count"):
            build_package_plan(self.root, parse_policy(count_limited))

    def test_secret_in_script_test_or_config_blocks_exact_package(self) -> None:
        secret = "AK" + "IA1234567890ABCDEF"
        for index, relative in enumerate(
            (
                "scripts/tests/test_payload.py",
                "scripts/validation_manifest.json",
                "scripts/package_manifest.json",
            )
        ):
            with self.subTest(relative=relative):
                root = self.base / ("secret-source-%d" % index)
                root.mkdir()
                path = root.joinpath(*relative.split("/"))
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_text("value = %s\n" % secret, encoding="utf-8")

                with self.assertRaises(PackageError) as raised:
                    build_package_plan(root, self.policy)

                diagnostic = str(raised.exception)
                self.assertIn("redacted rule", diagnostic)
                self.assertNotIn(secret, diagnostic)

    def test_secret_diagnostic_from_cli_never_prints_value(self) -> None:
        secret = "gh" + "p_" + ("A" * 32)
        self.write("scripts/tests/test_payload.py", secret + "\n")
        stdout = io.StringIO()
        stderr = io.StringIO()
        with redirect_stdout(stdout), redirect_stderr(stderr):
            status = main(
                ["--root", str(self.root), "--output", str(self.output), "--dry-run"],
                config_path=self.config_path,
            )

        self.assertEqual(2, status)
        self.assertIn("redacted rule", stderr.getvalue())
        self.assertNotIn(secret, stderr.getvalue())

    def test_default_summary_is_compact_and_verbose_lists_files(self) -> None:
        self.write("agents/core.md", "portable\n")
        plan = build_package_plan(self.root, self.policy)

        compact = render_summary(plan, dry_run=True)
        verbose = render_summary(plan, dry_run=True, verbose=True)

        self.assertIn("File count: 1", compact)
        self.assertIn("Manifest SHA-256:", compact)
        self.assertNotIn("agents/core.md", compact)
        self.assertIn("agents/core.md", verbose)


if __name__ == "__main__":
    unittest.main()
