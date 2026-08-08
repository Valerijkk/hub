from __future__ import annotations

import ast
import copy
import json
import sys
import tempfile
import unittest
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parents[1]
TOOLKIT_ROOT = SCRIPT_DIR.parent
sys.path.insert(0, str(SCRIPT_DIR))

from validate_toolkit import (  # noqa: E402
    ToolkitValidator,
    load_config,
    validate_json_instance,
)


class ContractSchemaTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.config = load_config(SCRIPT_DIR / "validation_manifest.json")
        cls.schema_paths = cls.config["contract_schemas"]["schemas"]

    def schema(self, title: str) -> dict[str, object]:
        path = TOOLKIT_ROOT / self.schema_paths[title]
        return json.loads(path.read_text(encoding="utf-8"))

    def example(self, title: str) -> tuple[dict[str, object], dict[str, object]]:
        schema = self.schema(title)
        return schema, copy.deepcopy(schema["examples"][0])

    def assert_invalid(self, title: str, record: dict[str, object]) -> None:
        schema = self.schema(title)
        self.assertTrue(validate_json_instance(record, schema))

    def test_contract_registry_and_examples_are_valid(self) -> None:
        issues = ToolkitValidator(TOOLKIT_ROOT, self.config).run(["contracts"])

        self.assertEqual([], issues)
        for title in self.schema_paths:
            schema = self.schema(title)
            for example in schema["examples"]:
                self.assertEqual([], validate_json_instance(example, schema), title)

    def test_requested_model_is_not_actual_without_runtime_evidence(self) -> None:
        _, record = self.example("ProvenanceRecord")
        record["actual_model"] = record["requested_model"]
        record["actual_model_evidence"] = None

        self.assert_invalid("ProvenanceRecord", record)

    def test_skill_text_cannot_be_capability_authority(self) -> None:
        _, record = self.example("CapabilityProfile")
        record["authority_source"] = "skill:deploy-approved"

        self.assert_invalid("CapabilityProfile", record)

    def test_unbounded_retry_budget_is_rejected(self) -> None:
        _, record = self.example("TaskContract")
        record["budgets"]["max_retries"] = "unbounded"

        self.assert_invalid("TaskContract", record)

    def test_latest_mutable_executable_revision_is_rejected(self) -> None:
        _, record = self.example("ProvenanceRecord")
        record["skills"] = [
            {
                "name": "example-skill",
                "source_uri": "https://agentskills.io/specification",
                "revision": "latest",
                "digest": "sha256:" + "0" * 64,
            }
        ]

        self.assert_invalid("ProvenanceRecord", record)

    def test_completion_without_evidence_is_rejected(self) -> None:
        _, record = self.example("VerificationContract")
        record["status"] = "PASS"
        record["completion_claim"] = True
        record["evidence"] = []

        self.assert_invalid("VerificationContract", record)

    def test_complete_evaluation_without_observations_is_rejected(self) -> None:
        _, record = self.example("WorkflowEvaluationRecord")
        record["status"] = "COMPLETE"

        self.assert_invalid("WorkflowEvaluationRecord", record)

    def test_contract_validation_uses_stdlib_and_local_modules_only(self) -> None:
        source = (SCRIPT_DIR / "validate_toolkit.py").read_text(encoding="utf-8")
        imported_roots = set()
        for node in ast.walk(ast.parse(source)):
            if isinstance(node, ast.Import):
                imported_roots.update(alias.name.split(".", 1)[0] for alias in node.names)
            elif isinstance(node, ast.ImportFrom) and node.module:
                imported_roots.add(node.module.split(".", 1)[0])

        local_modules = {"safe_filesystem", "secret_rules"}
        self.assertTrue(imported_roots - local_modules <= set(sys.stdlib_module_names))
        self.assertFalse({"jsonschema", "fastjsonschema", "pydantic"} & imported_roots)

    def test_missing_contract_schema_fails_closed(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            config = copy.deepcopy(self.config)
            config["contract_schemas"]["schemas"] = {
                "TaskContract": "standards/contracts/missing.schema.json"
            }

            issues = ToolkitValidator(root, config).run(["contracts"])

        self.assertEqual(["CONTRACT001"], [issue.code for issue in issues])


if __name__ == "__main__":
    unittest.main()
