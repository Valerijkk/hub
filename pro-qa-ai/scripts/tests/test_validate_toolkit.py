from __future__ import annotations

import json
import sys
import tempfile
import unittest
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SCRIPT_DIR))

from validate_toolkit import (  # noqa: E402
    ToolkitValidator,
    load_config,
    parse_yaml_like,
    render_text,
)


class ToolkitValidatorTests(unittest.TestCase):
    def setUp(self) -> None:
        self._temporary = tempfile.TemporaryDirectory()
        self.root = Path(self._temporary.name)
        self.config = json.loads(
            json.dumps(load_config(SCRIPT_DIR / "validation_manifest.json"))
        )

    def tearDown(self) -> None:
        self._temporary.cleanup()

    def write(self, relative: str, content: str) -> Path:
        path = self.root / relative
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content, encoding="utf-8")
        return path

    def write_bytes(self, relative: str, content: bytes) -> Path:
        path = self.root / relative
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(content)
        return path

    def codes(self, checks: list[str]) -> set[str]:
        return {
            issue.code
            for issue in ToolkitValidator(self.root, self.config).run(checks)
        }

    def test_valid_link_frontmatter_skill_and_whitespace(self) -> None:
        self.write("skills/example/reference.md", "# Reference\n")
        self.write(
            "skills/example/SKILL.md",
            "---\n"
            "name: example\n"
            "description: Use when validating an example.\n"
            "---\n"
            "# Example\n"
            "[Reference](reference.md)\n",
        )

        issues = ToolkitValidator(self.root, self.config).run(
            ["links", "frontmatter", "skills", "trailing_whitespace"]
        )

        self.assertEqual([], issues)

    def test_broken_relative_link_is_reported(self) -> None:
        self.write("README.md", "[Missing](docs/missing.md)\n")

        self.assertIn("LINK001", self.codes(["links"]))

    def test_links_support_parentheses_and_angle_bracket_spaces(self) -> None:
        self.write("docs/name(v1).md", "# Version\n")
        self.write("docs/name with spaces.md", "# Spaced\n")
        self.write(
            "README.md",
            "[Version](docs/name(v1).md)\n"
            "[Spaced][spaced]\n"
            "[spaced]: <docs/name with spaces.md>\n",
        )

        issues = ToolkitValidator(self.root, self.config).run(["links"])

        self.assertEqual([], issues)

    def test_links_reject_escape_absolute_drive_and_unc_targets(self) -> None:
        outside = self.root.parent / "outside.md"
        outside.write_text("# Outside\n", encoding="utf-8")
        self.write(
            "docs/README.md",
            "[Escape](../../outside.md)\n"
            "[Absolute](/etc/passwd)\n"
            "[Drive](C:/Windows/system.ini)\n"
            "[UNC](\\\\server\\share\\file.md)\n"
            "[External](https://example.com/docs)\n",
        )

        issues = ToolkitValidator(self.root, self.config).run(["links"])

        self.assertEqual(4, sum(issue.code == "LINK003" for issue in issues))

    def test_markdown_fragments_and_duplicate_heading_suffixes(self) -> None:
        self.write(
            "docs/target.md",
            "# Intro\n\n## Details & Notes\n\n# Intro\n",
        )
        self.write(
            "README.md",
            "[First](docs/target.md#intro)\n"
            "[Details](docs/target.md#details--notes)\n"
            "[Duplicate](docs/target.md#intro-1)\n"
            "[Missing](docs/target.md#not-present)\n",
        )

        issues = ToolkitValidator(self.root, self.config).run(["links"])

        self.assertEqual(["LINK004"], [issue.code for issue in issues])

    def test_file_policy_rejects_extension_nul_and_oversize(self) -> None:
        self.write("payload.bin", "printable\n")
        self.write_bytes("binary.txt", b"before\x00after")
        self.config["scan"]["file_policy"]["max_file_bytes"] = 8
        self.config["scan"]["file_policy"]["max_total_bytes"] = 32
        self.write("large.md", "123456789")

        codes = self.codes(["file_policy"])

        self.assertIn("FILE001", codes)
        self.assertIn("FILE002", codes)
        self.assertIn("FILE003", codes)

    def test_file_policy_rejects_total_and_count_limits(self) -> None:
        self.write("a.md", "123456")
        self.write("b.md", "abcdef")
        self.config["scan"]["file_policy"]["max_file_bytes"] = 10
        self.config["scan"]["file_policy"]["max_total_bytes"] = 10
        self.config["scan"]["file_policy"]["max_file_count"] = 1

        codes = self.codes(["file_policy"])

        self.assertIn("FILE004", codes)
        self.assertIn("FILE005", codes)

    def test_frontmatter_and_skill_name_are_validated(self) -> None:
        self.write(
            "skills/folder-name/SKILL.md",
            "---\n"
            "name: other-name\n"
            "name: duplicate\n"
            "description: Example.\n"
            "---\n"
            "# Skill\n",
        )

        codes = self.codes(["frontmatter", "skills"])

        self.assertIn("FM002", codes)
        self.assertIn("SKILL003", codes)

    def test_prompt_sequence_finds_duplicate_and_gap(self) -> None:
        self.write(
            "workflows/prompts/01-example.md",
            "# Prompts\n\n1. First\n3. Third\n3. Duplicate third\n",
        )

        codes = self.codes(["prompt_ids"])

        self.assertIn("PROMPT001", codes)
        self.assertIn("PROMPT003", codes)

    def test_generated_profile_requires_resolution_and_provenance(self) -> None:
        self.write(
            "context/PROJECT-CONTEXT.md",
            "---\n"
            "status: filled\n"
            "filled_by: intake\n"
            "filled_at: 2026-07-26\n"
            "---\n"
            "# Context\n"
            "Command: {{test_command}}\n",
        )

        codes = self.codes(["placeholders"])

        self.assertIn("PLACE001", codes)
        self.assertIn("PLACE002", codes)
        self.assertIn("PLACE005", codes)

    def test_structured_context_profile_uses_manifest_provenance(self) -> None:
        structured = self.config["placeholders"]["structured_profiles"]
        structured["template_directories"] = []
        structured["expected_markdown_files"] = ["PROJECT.md"]
        structured["manifest_required_paths"] = [
            "schema_version",
            "profile.slug",
            "profile.status",
            "sources[].id",
        ]
        self.write(
            "context/projects/demo/CONTEXT-MANIFEST.yaml",
            "schema_version: \"1.0\"\n"
            "profile:\n"
            "  slug: demo\n"
            "  status: filled\n"
            "scan:\n"
            "  status: partial\n"
            "  metadata_commands_executed: false\n"
            "  project_commands_executed: false\n"
            "sources:\n"
            "  - id: SRC-001\n",
        )
        self.write(
            "context/projects/demo/PROJECT.md",
            "---\n"
            "context_schema: \"1.0\"\n"
            "profile_slug: demo\n"
            "status: filled\n"
            "generated_at: 2026-07-26\n"
            "snapshot_commit: revision-1\n"
            "---\n"
            "# Project\n"
            "{{unresolved}}\n",
        )

        issues = ToolkitValidator(self.root, self.config).run(["placeholders"])

        self.assertEqual(["PLACE024"], [issue.code for issue in issues])

    def test_yaml_merge_key_is_accepted(self) -> None:
        parsed = parse_yaml_like(
            "defaults: &defaults\n"
            "  sha256: UNKNOWN\n"
            "sources:\n"
            "  - id: SRC-001\n"
            "    <<: *defaults\n"
            "    sha256: abc123\n"
        )

        self.assertEqual([], parsed.errors)
        self.assertEqual("*defaults", parsed.fields["sources[0].<<"])
        self.assertEqual(["sources[0]"], parsed.list_items["sources"])

    def test_complete_manifest_validates_every_source_and_output_item(self) -> None:
        structured = self.config["placeholders"]["structured_profiles"]
        structured["template_directories"] = []
        structured["expected_markdown_files"] = []
        structured["manifest_required_paths"] = [
            "schema_version",
            "profile.slug",
            "profile.status",
            "scan.status",
            "scan.metadata_commands_executed",
            "scan.project_commands_executed",
            "sources[].id",
            "sources[].root",
            "sources[].path",
            "sources[].kind",
            "sources[].sha256",
            "sources[].size_bytes",
            "sources[].modified_at",
            "sources[].observed_at",
            "outputs[].path",
            "outputs[].status",
            "outputs[].sha256",
            "outputs[].size_bytes",
            "outputs[].modified_at",
        ]
        timestamp = "2026-07-26T12:00:00Z"
        self.write(
            "context/projects/demo/CONTEXT-MANIFEST.yaml",
            "schema_version: \"1.0\"\n"
            "profile:\n"
            "  slug: demo\n"
            "  status: filled\n"
            "scan:\n"
            "  status: complete\n"
            "  metadata_commands_executed: false\n"
            "  project_commands_executed: false\n"
            "sources:\n"
            "  - id: SRC-001\n"
            "    root: project_root\n"
            "    path: README.md\n"
            "    kind: documentation\n"
            "    sha256: \"%s\"\n"
            "    size_bytes: 12\n"
            "    modified_at: \"%s\"\n"
            "    observed_at: \"%s\"\n"
            "  - id: SRC-002\n"
            "    root: project_root\n"
            "    path: broken.md\n"
            "    kind: documentation\n"
            "    sha256: bad\n"
            "    size_bytes: \"12\"\n"
            "    modified_at: UNKNOWN\n"
            "    observed_at: \"%s\"\n"
            "outputs:\n"
            "  - path: PROJECT.md\n"
            "    status: filled\n"
            "    sha256: \"%s\"\n"
            "    size_bytes: 20\n"
            "    modified_at: \"%s\"\n"
            "  - path: SOURCE-MAP.md\n"
            "    status: filled\n"
            "    sha256: bad\n"
            "    size_bytes: \"20\"\n"
            "    modified_at: UNKNOWN\n"
            % (
                "a" * 64,
                timestamp,
                timestamp,
                timestamp,
                "b" * 64,
                timestamp,
            ),
        )

        issues = ToolkitValidator(self.root, self.config).run(["placeholders"])

        self.assertGreaterEqual(
            sum(issue.code == "PLACE029" for issue in issues),
            3,
        )
        self.assertGreaterEqual(
            sum(issue.code == "PLACE031" for issue in issues),
            3,
        )
        self.assertTrue(
            any("sources[1]" in issue.message for issue in issues)
        )
        self.assertTrue(
            any("outputs[1]" in issue.message for issue in issues)
        )

    def test_manifest_rejects_legacy_aggregate_commands_field(self) -> None:
        structured = self.config["placeholders"]["structured_profiles"]
        structured["template_directories"] = []
        structured["expected_markdown_files"] = []
        structured["manifest_required_paths"] = [
            "schema_version",
            "profile.slug",
            "profile.status",
        ]
        self.write(
            "context/projects/demo/CONTEXT-MANIFEST.yaml",
            "schema_version: \"1.0\"\n"
            "profile:\n"
            "  slug: demo\n"
            "  status: partial\n"
            "scan:\n"
            "  status: partial\n"
            "  commands_executed: false\n"
            "  metadata_commands_executed: false\n"
            "  project_commands_executed: false\n",
        )

        self.assertIn("PLACE026", self.codes(["placeholders"]))

    def test_cot_route_and_portability_rules_are_configurable(self) -> None:
        self.write(
            "agents/example.md",
            "# Agent\n"
            "Please show your internal reasoning.\n"
            "Run /gsd-pln-phase next.\n"
            "Use Hub-specific defaults.\n",
        )

        codes = self.codes(["cot_disclosure", "gsd_routes", "portable_leakage"])

        self.assertIn("COT001", codes)
        self.assertIn("ROUTE001", codes)
        self.assertIn("LEAK001", codes)

    def test_cot_prohibition_is_not_a_disclosure_request(self) -> None:
        self.write(
            "agents/safe.md",
            "# Agent\nDo not expose private reasoning traces; return evidence instead.\n",
        )

        issues = ToolkitValidator(self.root, self.config).run(["cot_disclosure"])

        self.assertEqual([], issues)

    def test_stepwise_reasoning_disclosure_is_reported(self) -> None:
        self.write(
            "agents/unsafe.md",
            "# Agent\nExplain your thought process step by step.\n",
        )

        issues = ToolkitValidator(self.root, self.config).run(["cot_disclosure"])

        self.assertEqual(1, len(issues))
        self.assertIn("en-explain-reasoning-stepwise", issues[0].message)

    def test_known_gsd_route_with_underscore_is_accepted(self) -> None:
        self.write("README.md", "Run /gsd-extract_learnings after completion.\n")

        issues = ToolkitValidator(self.root, self.config).run(["gsd_routes"])

        self.assertEqual([], issues)

    def test_secret_value_never_appears_in_diagnostics(self) -> None:
        secret = "AK" + "IA1234567890ABCDEF"
        self.write("notes.md", "credential = %s\n" % secret)
        issues = ToolkitValidator(self.root, self.config).run(["secrets"])

        self.assertEqual(1, len(issues))
        self.assertEqual("SECRET001", issues[0].code)
        self.assertNotIn(secret, issues[0].message)
        self.assertNotIn(
            secret, render_text(self.root, ["secrets"], issues)
        )

    def test_secret_scan_includes_scripts_tests_and_config(self) -> None:
        secret = "AK" + "IA1234567890ABCDEF"
        self.write("scripts/tool.py", "value = %s\n" % secret)
        self.write("scripts/tests/test_tool.py", "value = %s\n" % secret)
        self.write("scripts/config.json", "{\"value\": \"%s\"}\n" % secret)

        issues = ToolkitValidator(self.root, self.config).run(["secrets"])
        rendered = render_text(self.root, ["secrets"], issues)

        self.assertEqual(3, sum(issue.code == "SECRET001" for issue in issues))
        self.assertNotIn(secret, rendered)


    def test_toolkit_manifest_provenance_contract(self) -> None:
        self.write("sources/PRIMARY-SOURCES.md", "# Sources\n")
        self.write(
            "MANIFEST.md",
            "---\n"
            "manifest_version: 1\n"
            "toolkit_version: 0.1.0-alpha.1\n"
            "status: pre-release\n"
            "owner: automation-quality\n"
            "layer: ai-quality\n"
            "source_registry: sources/PRIMARY-SOURCES.md\n"
            "validation_config: scripts/validation_manifest.json\n"
            "validation_status: NOT_RUN\n"
            "provenance: official sources plus repository inspection\n"
            "---\n"
            "# Manifest\n",
        )

        issues = ToolkitValidator(self.root, self.config).run(["provenance"])

        self.assertEqual([], issues)


if __name__ == "__main__":
    unittest.main()
