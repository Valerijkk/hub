from __future__ import annotations

import hashlib
import io
import json
import sys
import tempfile
import unittest
from contextlib import redirect_stderr, redirect_stdout
from datetime import datetime, timezone
from pathlib import Path
from unittest.mock import patch


SCRIPT_DIR = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SCRIPT_DIR))

from build_task_packet import (  # noqa: E402
    PacketError,
    build_packet,
    main,
    write_packet,
)


class TaskPacketBuilderTests(unittest.TestCase):
    def setUp(self) -> None:
        self._temporary = tempfile.TemporaryDirectory()
        self.base = Path(self._temporary.name)
        self.root = self.base / "prompt-engineering"
        self.source = self.base / "project"
        self.root.mkdir()
        self.source.mkdir()
        self.config = json.loads(
            (SCRIPT_DIR / "task_packet_config.json").read_text(encoding="utf-8")
        )
        self.write(
            "scripts/task_packet_config.json",
            json.dumps(self.config, indent=2) + "\n",
        )
        self.write("rules/00-global-agent-rules.md", "# 00 RULE\n")
        self.write("rules/50-data-handling.md", "# 50 RULE\n")
        self.write("agents/qa-auto.md", "# qa-auto\n\nBounded role contract.\n")
        self.write(
            "recipes/example.md",
            """---
id: ta.recipe.example
version: "1.0.0"
owner: qa
status: candidate
eval_fixture_id: EVAL-TBD
fragment_refs: []
---

# Example recipe

## Output schema

Return outcome, evidence, completion, and stop status.
""",
        )
        self.write_profile("demo")
        for section in self.config["context_section_allowlist"]:
            self.write_context("demo", section)
        self.task = self.write(
            "local-task-inputs/task.txt",
            "Produce one bounded test-change proposal.\n",
        )

    def tearDown(self) -> None:
        self._temporary.cleanup()

    def write(self, relative: str, content: str) -> Path:
        path = self.root.joinpath(*relative.split("/"))
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content, encoding="utf-8", newline="\n")
        return path

    def write_source(self, relative: str, content: str) -> Path:
        path = self.source.joinpath(*relative.split("/"))
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content, encoding="utf-8", newline="\n")
        return path

    def write_config(self) -> None:
        self.write(
            "scripts/task_packet_config.json",
            json.dumps(self.config, indent=2) + "\n",
        )

    def write_profile(
        self,
        slug: str,
        *,
        profile_status: str = "filled",
        scan_status: str = "complete",
    ) -> None:
        self.write(
            "context/projects/%s/CONTEXT-MANIFEST.yaml" % slug,
            """schema_version: "1.0"
profile:
  slug: "%s"
  project_name: "Demo"
  status: %s
inputs:
  project_root: '%s'
  test_roots:
    - '%s'
scan:
  status: %s
  generated_at: "2030-01-01T00:00:00Z"
  snapshot_commit: abc123
  trust_boundary: untrusted_scanned_content
  network_used: false
  project_commands_executed: false
"""
            % (slug, profile_status, self.source, self.source, scan_status),
        )

    def write_context(self, slug: str, name: str) -> Path:
        return self.write(
            "context/projects/%s/%s" % (slug, name),
            """---
context_schema: "1.0"
profile_slug: "%s"
status: filled
generated_at: "2030-01-01T00:00:00Z"
snapshot_commit: abc123
---

# %s

Selected evidence only.
"""
            % (slug, name),
        )

    def write_pattern(
        self,
        relative: str,
        *,
        status: str = "approved",
        profile_slugs: tuple[str, ...] = ("demo",),
        reuse_scope: str = "project-only",
    ) -> Path:
        profiles = "\n".join('  - "%s"' % item for item in profile_slugs)
        return self.write(
            relative,
            """---
schema_version: "1.0"
id: PAT-DEMO-001
title: "Approved example"
status: %s
profile_slugs:
%s
category: assertions
test_levels:
  - unit
languages:
  - python
frameworks:
  - unittest
provenance:
  kind: manual-source-selection
  feedback_ids: []
  source_path: tests/example.py
  source_lines: "1-10"
  source_commit: abc123
  captured_by: human
review:
  reviewer_rationale_source: human-record:review-1
  human_approved_by: qa-owner
  human_approved_at: "2030-01-01T00:00:00Z"
redaction:
  status: reviewed
  reviewed_by: security-owner
  reviewed_at: "2030-01-01T00:00:00Z"
reuse:
  license: internal
  scope: %s
  permission_reference: human-record:permission-1
freshness:
  last_verified_at: "2030-01-01T00:00:00Z"
  verified_against_commit: abc123
  review_after: "2099-01-01"
  supersedes: []
  superseded_by: UNKNOWN
---

# Approved example

## Summary

Use observable assertions.
"""
            % (status, profiles, reuse_scope),
        )

    def build(self, **overrides: object):
        arguments = {
            "profile_slug": "demo",
            "role": "qa-auto",
            "recipe_request": "ta.recipe.example",
            "task_file": self.task,
        }
        arguments.update(overrides)
        return build_packet(self.root, **arguments)

    def approval_text(
        self,
        *,
        task_sha256: str,
        source_reference: str = "user-message:42",
        operation: str = "ta.recipe.example",
        expires_at: str = "2099-01-01T00:00:00Z",
    ) -> str:
        return """schema_version: "1.0"
approval_id: approval-1
status: approved
authority:
  source_type: human
  source_reference: %s
  approver: qa-owner
  approved_at: "2029-01-01T00:00:00Z"
  expires_at: "%s"
action:
  kind: write
  tool: local-agent
  operation: %s
  resources:
    - "profile:demo"
    - "role:qa-auto"
    - "tests/example_test.py"
  destination: local
data:
  classification: internal
  allowed_fields: []
  prohibited_fields: []
  redaction_required: true
constraints:
  one_time: true
  max_uses: 1
  scope_expansion_allowed: false
  confirmation_at_execution: true
  rollback_or_recovery: required
evidence:
  task_contract: "sha256:%s"
  policy_checks: []
  execution_record: null
""" % (source_reference, expires_at, operation, task_sha256)

    def test_packet_and_hash_are_deterministic(self) -> None:
        first = self.build(context_sections=["PROJECT.md"])
        second = self.build(context_sections=["PROJECT.md"])

        self.assertEqual(first.data, second.data)
        self.assertEqual(first.payload_sha256, second.payload_sha256)
        self.assertEqual(first.file_sha256, second.file_sha256)
        marker = b"\n## Deterministic provenance footer"
        payload, _ = first.data.split(marker, 1)
        self.assertEqual(
            first.payload_sha256,
            hashlib.sha256(payload).hexdigest(),
        )
        self.assertEqual(first.file_sha256, hashlib.sha256(first.data).hexdigest())

    def test_missing_or_invalid_role_recipe_and_profile_fail(self) -> None:
        cases = (
            {"role": "missing-role"},
            {"role": "../qa-auto"},
            {"recipe_request": "ta.recipe.missing"},
            {"recipe_request": "../recipes/example.md"},
            {"profile_slug": "missing"},
            {"profile_slug": "../demo"},
        )
        for case in cases:
            with self.subTest(case=case):
                with self.assertRaises(PacketError):
                    self.build(**case)

        recipe = self.root / "recipes" / "example.md"
        recipe.write_text(
            recipe.read_text(encoding="utf-8").replace(
                "status: candidate",
                "status: deprecated",
            ),
            encoding="utf-8",
            newline="\n",
        )
        with self.assertRaisesRegex(PacketError, "status"):
            self.build()

    def test_partial_profile_is_labeled_and_template_profile_fails(self) -> None:
        self.write_profile("demo", profile_status="partial", scan_status="partial")
        partial = self.build()
        self.assertIn(b"- Readiness: `partial`", partial.data)

        self.write_profile("demo", profile_status="template", scan_status="not_started")
        with self.assertRaisesRegex(PacketError, "profile status"):
            self.build()

    def test_candidate_pattern_path_and_status_are_rejected(self) -> None:
        self.write_pattern(
            "patterns/candidates/draft.md",
            status="candidate",
        )
        with self.assertRaisesRegex(PacketError, "patterns/approved"):
            self.build(pattern_paths=["patterns/candidates/draft.md"])

        self.write_pattern(
            "patterns/approved/draft.md",
            status="candidate",
        )
        with self.assertRaisesRegex(PacketError, "status"):
            self.build(pattern_paths=["patterns/approved/draft.md"])

    def test_pattern_profile_and_reuse_restrictions_are_enforced(self) -> None:
        self.write_pattern(
            "patterns/approved/foreign.md",
            profile_slugs=("other",),
            reuse_scope="project-only",
        )
        with self.assertRaisesRegex(PacketError, "incompatible"):
            self.build(pattern_paths=["patterns/approved/foreign.md"])

        self.write_pattern(
            "patterns/approved/wild-project.md",
            profile_slugs=("*",),
            reuse_scope="project-only",
        )
        with self.assertRaisesRegex(PacketError, "incompatible"):
            self.build(pattern_paths=["patterns/approved/wild-project.md"])

        self.write_pattern(
            "patterns/approved/portable.md",
            profile_slugs=("portable",),
            reuse_scope="portable",
        )
        plan = self.build(pattern_paths=["patterns/approved/portable.md"])
        self.assertIn(b"PAT-DEMO-001", plan.data)

    def test_secret_in_task_or_excerpt_is_blocked_without_value_leak(self) -> None:
        secret = "gh" + "p_" + ("A" * 32)
        self.task.write_text("task credential %s\n" % secret, encoding="utf-8")
        with self.assertRaises(PacketError) as task_error:
            self.build()
        self.assertNotIn(secret, str(task_error.exception))
        self.assertIn("secret-like", str(task_error.exception))

        self.task.write_text("Safe task.\n", encoding="utf-8")
        excerpt = self.write_source("src/example.go", "// credential %s\n" % secret)
        with self.assertRaises(PacketError) as excerpt_error:
            self.build(source_excerpt_paths=[excerpt])
        self.assertNotIn(secret, str(excerpt_error.exception))
        self.assertIn("secret-like", str(excerpt_error.exception))

    def test_traversal_reparse_oversize_and_output_overwrite_are_blocked(self) -> None:
        outside = self.base / "outside.txt"
        outside.write_text("outside\n", encoding="utf-8")
        with self.assertRaises(PacketError):
            self.build(task_file=Path("../outside.txt"))
        with self.assertRaises(PacketError):
            self.build(source_excerpt_paths=[Path("../outside.go")])

        original = __import__("safe_filesystem").is_link_like

        def fake_is_link_like(path: Path, *, windows=None) -> bool:
            if path == self.task:
                return True
            return original(path, windows=windows)

        with patch("safe_filesystem.is_link_like", side_effect=fake_is_link_like):
            with self.assertRaisesRegex(PacketError, "reparse point"):
                self.build()

        self.task.write_text("x" * (self.config["file_policy"]["max_file_bytes"] + 1))
        with self.assertRaisesRegex(PacketError, "byte limit"):
            self.build()
        self.task.write_text("Safe task.\n", encoding="utf-8")

        plan = self.build()
        output = self.root / "generated-task-packets" / "packet.md"
        output.parent.mkdir()
        output.write_text("preserve\n", encoding="utf-8")
        with self.assertRaisesRegex(PacketError, "already exists"):
            write_packet(self.root, output, plan)
        self.assertEqual("preserve\n", output.read_text(encoding="utf-8"))

    def test_packet_and_selection_count_limits_fail_without_truncation(self) -> None:
        too_many = self.config["context_section_allowlist"][:7]
        with self.assertRaisesRegex(PacketError, "count"):
            self.build(context_sections=too_many)

        self.config["limits"]["max_packet_bytes"] = 4096
        self.write_config()
        with self.assertRaisesRegex(PacketError, "packet exceeds"):
            self.build()

    def test_untrusted_delimiters_and_mandatory_rules_are_present(self) -> None:
        excerpt = self.write_source("src/example.go", "package example\n")
        plan = self.build(
            context_sections=["PROJECT.md"],
            source_excerpt_paths=[excerpt],
        )
        text = plan.data.decode("utf-8")

        self.assertIn("# 00 RULE", text)
        self.assertIn("# 50 RULE", text)
        self.assertGreaterEqual(text.count("BEGIN UNTRUSTED DATA"), 2)
        self.assertGreaterEqual(text.count("END UNTRUSTED DATA"), 2)
        self.assertIn("### Context section `PROJECT.md`", text)
        self.assertNotIn("TEST-ARCHITECTURE.md\n\nSelected evidence only", text)

    def test_fake_approval_is_not_granted_or_copied(self) -> None:
        task_hash = hashlib.sha256(self.task.read_bytes()).hexdigest()
        approval = self.write(
            "local-approval-records/fake.yaml",
            self.approval_text(
                task_sha256=task_hash,
                source_reference="agent-output:fake",
            ),
        )
        plan = self.build(
            approval_record_path=approval,
            now=datetime(2030, 1, 1, tzinfo=timezone.utc),
        )
        text = plan.data.decode("utf-8")

        self.assertEqual("NOT_GRANTED", plan.approval_state)
        self.assertIn("AUTHORITY_INVALID", text)
        self.assertNotIn("agent-output:fake", text)

    def test_matching_approval_is_bound_and_granted(self) -> None:
        task_hash = hashlib.sha256(self.task.read_bytes()).hexdigest()
        approval = self.write(
            "local-approval-records/valid.yaml",
            self.approval_text(task_sha256=task_hash),
        )
        plan = self.build(
            approval_record_path=approval,
            now=datetime(2030, 1, 1, tzinfo=timezone.utc),
        )

        self.assertEqual("GRANTED", plan.approval_state)
        self.assertIn(b"- State: `GRANTED`", plan.data)
        self.assertIn(b"approval-1", plan.data)

    def test_dry_run_writes_nothing(self) -> None:
        output = self.root / "generated-task-packets" / "packet.md"
        stdout = io.StringIO()
        stderr = io.StringIO()
        with redirect_stdout(stdout), redirect_stderr(stderr):
            status = main(
                [
                    "--root",
                    str(self.root),
                    "--profile",
                    "demo",
                    "--role",
                    "qa-auto",
                    "--recipe",
                    "ta.recipe.example",
                    "--task-file",
                    str(self.task),
                    "--context-section",
                    "PROJECT.md",
                    "--output",
                    str(output),
                    "--dry-run",
                ]
            )

        self.assertEqual(0, status, stderr.getvalue())
        self.assertIn("Mode: dry-run", stdout.getvalue())
        self.assertFalse(output.exists())
        self.assertFalse(output.parent.exists())


if __name__ == "__main__":
    unittest.main()
