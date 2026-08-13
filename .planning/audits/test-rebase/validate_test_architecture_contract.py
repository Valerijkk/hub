#!/usr/bin/env python3
"""Fail-closed validator for the accepted greenfield partial cutover.

The contract, founder manifest and TASK-202 inventory are untrusted inputs.
All authority-bearing counts, selectors, path digests and cutover invariants
are constants in this module. Only Python's standard library is used.
"""

from __future__ import annotations

import argparse
from collections import Counter
from copy import deepcopy
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import re
import stat
import sys
import tempfile
from typing import Any
import unicodedata


_SEALED_REPO_ROOT = os.environ.get("HUB_PHASE128_REPO_ROOT")
REPO_ROOT = (
    Path(_SEALED_REPO_ROOT)
    if _SEALED_REPO_ROOT is not None
    else Path(__file__).resolve().parents[3]
)
if not REPO_ROOT.is_absolute():
    raise RuntimeError("HUB_PHASE128_REPO_ROOT must be absolute")
DEFAULT_CONTRACT = REPO_ROOT / ".planning/audits/test-rebase/test-architecture-contract.json"
DEFAULT_MANIFEST = REPO_ROOT / ".planning/audits/test-rebase/founder-greenfield-cutover-manifest.json"
DEFAULT_INVENTORY = REPO_ROOT / ".planning/audits/test-rebase/inventory.json"
DEFAULT_ADR = REPO_ROOT / ".planning/decisions/ADR-TEST-ARCHITECTURE-REBASE-2026.md"

MAX_JSON_BYTES = 8 * 1024 * 1024
MAX_ADR_BYTES = 1 * 1024 * 1024
MAX_INVENTORY_ROWS = 2_000
MAX_DELETE_PATHS = 1_000
MAX_PATH_COMPONENTS = 64
MAX_COMPONENT_UTF8_BYTES = 255
MAX_PATH_UTF8_BYTES = 4096

LEDGER_REVISION = "b9f690da5cdcf322fc5b28a8d96c71a422407154"
LEDGER_SHA256 = "83b6a31149807d21441a2ede91dec0b36b4a4396c4c0ceb58308b58878f5ad89"
NON_LEDGER_SHA256 = "9b06611c2e5f9342ae081b6917e4033808c642e0b65216d7e60a0b2f89a62d49"

REQUIRED_DECISION = {
    "id": "ADR-TEST-ARCHITECTURE-REBASE-2026",
    "status": "Accepted",
    "accepted_on": "2026-08-07",
    "authority": "founder",
    "migration_mode": "greenfield-rebuild",
    "row_parity_required": False,
    "dual_run_required": False,
    "partial_cutover_allowed": True,
    "phase_129_completion_requires_blocked_units_moved": True,
    "protected_zone_dependency": "forbidden",
    "docker_only_execution": True,
}
REQUIRED_E2E_BASELINE = {
    "status": "pass",
    "execution": "task-owned-limited-docker",
    "collected": 53,
    "passed": 48,
    "expected_skipped": 5,
    "failed": 0,
    "environment_skipped": 0,
    "expected_skip_breakdown": {
        "provider_dependent_ai": 4,
        "health_negative_sentinel": 1,
    },
}
REQUIRED_ROOTS = {
    "backend_unit": "backend/tests/",
    "frontend_unit": "frontend/tests/",
    "frontend_support": "frontend/tests/support/",
    "python_project": "testframework/",
}
REQUIRED_FOUNDER_MANIFEST = {
    "path": ".planning/audits/test-rebase/founder-greenfield-cutover-manifest.json",
    "required_status": "verified-hook-blocked",
    "ledger_revision": LEDGER_REVISION,
    "ledger_sha256": LEDGER_SHA256,
    "rows": 1049,
    "declared_tests": 4246,
    "move_rows": 507,
    "move_declared_tests": 1673,
    "keep_blocked_rows": 245,
    "keep_blocked_declared_tests": 1428,
    "delete_rows": 297,
    "delete_declared_tests": 1145,
    "non_ledger_delete_paths": 28,
    "non_ledger_delete_paths_sha256": NON_LEDGER_SHA256,
}
REQUIRED_PYTHON_ARCHITECTURE = {
    "installable_package_root": "testframework/src/hub_testkit/",
    "executable_suite_roots": [
        "testframework/tests/api/",
        "testframework/tests/integration/",
        "testframework/tests/contract/",
        "testframework/tests/e2e/",
    ],
    "reusable_module_roots": [
        "testframework/src/hub_testkit/config/",
        "testframework/src/hub_testkit/runtime/",
        "testframework/src/hub_testkit/http/",
        "testframework/src/hub_testkit/contract/",
        "testframework/src/hub_testkit/browser/",
        "testframework/src/hub_testkit/integration/",
        "testframework/src/hub_testkit/observability/",
        "testframework/src/hub_testkit/traceability/",
        "testframework/src/hub_testkit/pytest_plugin/",
        "testframework/src/hub_testkit/domains/<domain>/",
    ],
    "composition": {
        "path": "testframework/tests/conftest.py",
        "responsibility": "explicit-plugin-and-fixture-composition-only",
    },
    "framework_contract_root": "testframework/selftests/",
    "case_root": "testframework/cases/",
    "schema_root": "testframework/schemas/",
    "policy_root": "testframework/policies/",
    "docker_root": "testframework/docker/",
    "runner_root": "testframework/runner/",
    "load_root": "testframework/load/",
    "reports_root": "testframework/reports/<task-run>/",
    "materialization": {
        "creation_policy": "first-accepted-consumer-only",
        "empty_directories_required": False,
        "speculative_layers_required": False,
    },
}
REQUIRED_FORBIDDEN_PYTHON_PATHS = [
    "hub_testkit/",
    "tests/",
    "cases/",
    "selftests/",
    "docker/",
    "load/",
    "testframework/python/",
    "testframework/core/",
    "testframework/api/",
    "testframework/integration/",
    "testframework/contract/",
    "testframework/e2e/",
    "testframework/conftest.py",
    "testframework/tests/e2e/support/",
    "testframework/tests/e2e/pages/",
    "testframework/tests/e2e/steps/",
]

REQUIRED_GROUPS = [
    {
        "id": "MOVE-GO-UNIT-HOOK-FREE",
        "action": "move",
        "selector": {
            "runner": "go-test",
            "target_class": "go-unit",
            "hook_consumers_count": 0,
        },
        "expected_count": 350,
        "expected_declared_tests": 1264,
        "source_paths_sha256": "af256cc0c5ad59fe6671641cea5ac138b2e6a4796c23b64aa038ec216ca40ff7",
        "destination_root": "backend/tests/",
        "eligibility": "service-free-isolated-go-unit",
    },
    {
        "id": "KEEP-GO-UNIT-HOOK-DEPENDENT",
        "action": "keep-until-refactor-then-move",
        "selector": {
            "runner": "go-test",
            "target_class": "go-unit",
            "hook_consumers_count": ">0",
        },
        "expected_count": 245,
        "expected_declared_tests": 1428,
        "expected_hook_consumer_links": 653,
        "expected_distinct_hook_files": 42,
        "expected_distinct_hook_symbols": 434,
        "source_paths_sha256": "86e55748e6c5986ec4eef370b51fb3f02c7fca7ec24fa8eb1e85d1913ad33924",
        "release_condition": "zero-hook-consumer-regenerated-ledger-and-docker-green",
    },
    {
        "id": "MOVE-FRONTEND-UNIT",
        "action": "move",
        "selector": {"runner": "vitest", "target_class": "frontend-unit"},
        "expected_count": 156,
        "expected_declared_tests": 409,
        "source_paths_sha256": "96fd1cacc09c39a06e894600de9168c5d9a3cf1971ec63f5bfe9c01d978b2b24",
        "destination_root": "frontend/tests/",
        "eligibility": "service-free-vitest-rtl-unit",
    },
    {
        "id": "MOVE-FRONTEND-RENDER-SUPPORT",
        "action": "move",
        "selector": {
            "source_path": "testframework/stages/unit/frontend-test/render.tsx",
            "runner": "vitest-support",
            "target_class": "frontend-support",
        },
        "expected_count": 1,
        "expected_declared_tests": 0,
        "source_paths_sha256": "a49f50bbf0b6667a10b0392169fb760311defe06288f9a3775f9d23ebdfe5a23",
        "destination_root": "frontend/tests/support/",
        "eligibility": "retained-unit-reachable-render-support",
    },
    {
        "id": "DELETE-GO-COMPONENT",
        "action": "delete-no-translation",
        "selector": {
            "runner": "go-test",
            "target_class_in": ["python-api", "python-integration"],
        },
        "expected_count": 237,
        "expected_declared_tests": 982,
        "source_paths_sha256": "45bc86bd5be985e48d9b599200aa29641e02eb4947c97cb5ff0eb3f49d941e18",
        "replacement_authority": "requirements-and-risk-greenfield-python",
    },
    {
        "id": "DELETE-FRONTEND-STATIC-CONTRACT",
        "action": "delete-no-translation",
        "selector": {"runner": "vitest", "target_class": "python-static-contract"},
        "expected_count": 24,
        "expected_declared_tests": 110,
        "source_paths_sha256": "4952f7970784c22570a263c31836a2e08a31a12592cccbb78e2b696409ae92b9",
        "replacement_authority": "requirements-and-risk-greenfield-python",
    },
    {
        "id": "DELETE-FRONTEND-COMPILE-FIXTURES",
        "action": "delete-no-translation",
        "selector": {
            "source_path_in": [
                "testframework/stages/unit/frontend-test/domains/platform/updateBase-version-required.fixture.ts",
                "testframework/stages/unit/frontend-test/domains/tracker/projects-sprint-version-required.fixture.ts",
            ],
            "runner": "vitest-support",
        },
        "expected_count": 2,
        "expected_declared_tests": 0,
        "source_paths_sha256": "dd0084eb74d5e65fc574849c051c30bedb717379cb215b6ec8c10674f28ea956",
        "replacement_authority": "requirements-and-risk-greenfield-python",
    },
    {
        "id": "DELETE-PLAYWRIGHT-TS-SPECS",
        "action": "delete-no-translation",
        "selector": {"runner": "playwright-ts", "target_class": "python-e2e"},
        "expected_count": 20,
        "expected_declared_tests": 53,
        "source_paths_sha256": "aa521c92eca8744fe41e22d2b278d12d0f9562e08936938a34a6f54c70c0a2fb",
        "replacement_authority": "requirements-and-risk-greenfield-python",
    },
    {
        "id": "DELETE-PLAYWRIGHT-TS-SUPPORT",
        "action": "delete-no-translation",
        "selector": {
            "runner": "playwright-support-ts",
            "target_class": "python-e2e-support",
        },
        "expected_count": 14,
        "expected_declared_tests": 0,
        "source_paths_sha256": "ec40b12fce301f951d299a1e32a651eab73cd4b3cac621d2a1c65aaa6b4e4eef",
        "replacement_authority": "requirements-and-risk-greenfield-python",
    },
]

REQUIRED_CUTOVER_SNAPSHOT = {
    "ledger_revision": LEDGER_REVISION,
    "ledger_sha256": LEDGER_SHA256,
    "rows": 1049,
    "declared_tests": 4246,
    "move_files": 507,
    "move_declared_tests": 1673,
    "keep_blocked_files": 245,
    "keep_blocked_declared_tests": 1428,
    "delete_files": 297,
    "delete_declared_tests": 1145,
    "groups": REQUIRED_GROUPS,
    "non_ledger_delete": {
        "expected_count": 28,
        "source_paths_sha256": NON_LEDGER_SHA256,
        "source": "founder-greenfield-cutover-manifest.json",
    },
}

REQUIRED_CLASS_RUNNERS = {
    "go-unit": "go-test",
    "frontend-unit": "vitest",
    "frontend-support": "vitest-support",
    "python-api": "go-test",
    "python-integration": "go-test",
    "python-static-contract": "vitest",
    "python-e2e": "playwright-ts",
    "python-e2e-support": "playwright-support-ts",
}
GO_UNIT_TARGET = re.compile(r"^backend/tests/[a-z0-9][a-z0-9_-]*/.+_test\.go$")
FRONTEND_UNIT_TARGET = re.compile(
    r"^frontend/tests/[a-z0-9][a-z0-9_-]*/.+\.(?:test|spec)\.tsx?$"
)
EXTERNAL_GO_DEPENDENCY_FRAGMENTS = (
    "dockertest",
    "testcontainers",
    "/backend/testframework/env",
    "/backend/testframework/mongotest",
)

REQUIRED_UNIT_ELIGIBILITY = {
    "go_forbidden_dependency_fragments": [
        "dockertest",
        "testcontainers",
        "testframework/env",
        "testframework/mongotest",
        "postgres",
        "mongodb",
        "redis",
        "nats",
        "real-listener",
        "application-startup",
        "export-hook",
    ],
    "frontend_forbidden_traits": [
        "real-browser",
        "application-server",
        "service-dependency",
        "source-scan",
        "documentation-drift",
        "filename-or-token-assertion",
    ],
    "frontend_support_policy": "retain-only-when-all-consumers-are-retained-units",
}
REQUIRED_GREENFIELD = {
    "legacy_non_unit_translation": "forbidden-as-acceptance",
    "design_authorities": [
        "current-product-behavior",
        "docs-and-accepted-contracts",
        "tasks-and-backlog",
        "incidents-and-observed-risks",
        "security-constraints",
    ],
    "required_case_metadata": [
        "test_case_id",
        "task_or_backlog_ids",
        "requirement_story_nfr_or_contract_ids",
        "risk_id",
        "priority_or_severity",
        "scenario_kind",
        "domain",
        "preconditions",
        "action",
        "observable_expected_result",
    ],
    "required_design_dimensions": [
        "happy-path",
        "negative-and-error-handling",
        "boundary-and-edge",
        "corner-and-state-transition",
        "authorization-and-security",
        "recovery-idempotency-and-concurrency-when-applicable",
    ],
    "applicability_exclusion_requires_owner_reason": True,
    "definition_of_done": [
        "reviewed requirement-risk matrix with unique IDs",
        "non-zero Docker-only collection matching the matrix",
        "recorded pass-fail-skip totals and runtime",
        "controlled negative or mutation sentinel",
        "non-empty sanitized JUnit and Allure linked to IDs",
        "no live external egress or shared-production state",
        "independent verifier acceptance",
    ],
}
REQUIRED_COVERAGE = {
    "historical_go_overall_percent": 82.0373,
    "historical_status": "diagnostic-baseline-only",
    "go_statement_threshold_is_migration_blocker": False,
    "security_package_threshold_is_migration_blocker": False,
    "python_black_box_claims_go_coverage": False,
    "unit_gate": "compile-race-pass-no-unexpected-skip-service-free",
    "future_threshold_requires_separate_founder_approved_adr": True,
}
REQUIRED_SKIP_POLICY = {
    "unit_retries": 0,
    "e2e_infrastructure_retries_max": 2,
    "zero_collection": "fail",
    "mandatory_dependency_skip": "fail",
    "unexpected_skip": "fail",
    "quarantine_requires": ["issue", "owner", "reason", "expiry_wave"],
}
REQUIRED_CUTOVER = {
    "cutover_phase": 129,
    "exact_allowlist_required": True,
    "python_rebuild_is_deletion_precondition": False,
    "partial_cutover_allowed": True,
    "immediate_move_rows": 507,
    "immediate_delete_ledger_rows": 297,
    "immediate_delete_non_ledger_paths": 28,
    "blocked_go_unit_rows_retained": 245,
    "phase_129_completion_requires_blocked_units_refactored_and_moved": True,
    "moved_unit_docker_green_required": True,
    "unexpected_source_path_policy": "fail",
    "deletion_scope": "sealed-297-ledger-rows-plus-28-non-ledger-paths",
    "hook_removal_requires_exact_zero_consumer_scan": True,
    "cutover_manifest": ".planning/audits/test-rebase/cutover-manifest.json",
    "commit_shape": "ordered-atomic-cutover-commits",
    "rollback_strategy": "reverse-order-git-revert-of-affected-atomic-commits",
    "cleanup_scope": "task-owned-resources-only",
    "finished_container_policy": "remove-unless-named-downstream-owner-recorded",
}
REQUIRED_MARKERS = [
    "ADR_STATUS=ACCEPTED",
    "BASELINE_E2E=DOCKER:53_COLLECTED:48_PASS:5_EXPECTED_SKIP:0_FAIL",
    "MIGRATION_MODE=GREENFIELD_REBUILD",
    f"LEDGER_HEAD={LEDGER_REVISION}",
    f"LEDGER_SHA256={LEDGER_SHA256}",
    "CUTOVER_PARTITION=MOVE:507:KEEP_BLOCKED:245:DELETE:297:NONLEDGER_DELETE:28",
    "PARTIAL_CUTOVER_ALLOWED=true",
    "PHASE129_EXIT=HOOK_DEPENDENT_REFACTORED_AND_MOVED",
    "ROOT_PYTHON_ONLY=testframework/**",
    "PYTHON_PACKAGE_ROOT=testframework/src/hub_testkit/**",
    "PYTHON_SUITE_ROOT=testframework/tests/**",
    "PYTHON_COMPOSITION=testframework/tests/conftest.py:COMPOSITION_ONLY",
    "PYTHON_SELFTEST_ROOT=testframework/selftests/**",
    "GENERIC_CORE=FORBIDDEN",
    "PROGRESSIVE_TREE=FIRST_CONSUMER:NO_EMPTY_DIRS",
    "BACKEND_UNIT_ROOT=backend/tests/<domain>/**",
    "FRONTEND_UNIT_ROOT=frontend/tests/<domain>/**",
    "KEEP_ONLY=ISOLATED_NATIVE_UNITS",
    "NO_ROW_PARITY=true",
    "NO_DUAL_RUN=true",
    "GO_COVERAGE_MIGRATION_BLOCKER=false",
    "DELETE_ALLOWLIST=TASK-202:SEALED_SOURCE_PATH_SHA256_GROUPS",
    "TRACEABILITY=TASK:REQUIREMENT:TEST_CASE:RISK",
    "SCENARIO_DESIGN=HAPPY:NEGATIVE:EDGE:CORNER:SECURITY",
    "DOCKER_ONLY=true",
    "ROLLBACK=ATOMIC_COMMIT_REVERT",
    "PROTECTED_ZONE_DEPENDENCY=FORBIDDEN",
]

REQUIRED_JOBS = [
    {
        "id": "backend-unit",
        "paths": ["backend/tests/**"],
        "services": [],
        "timeout_seconds": 600,
        "artifacts": ["go-unit-results.json", "go-unit.log", "race-result.json"],
    },
    {
        "id": "frontend-unit",
        "paths": ["frontend/tests/**"],
        "services": [],
        "timeout_seconds": 900,
        "artifacts": ["vitest-results.json", "frontend-unit.log", "typecheck-lint.log"],
    },
    {
        "id": "python-contract",
        "paths": ["testframework/tests/contract/**"],
        "services": [],
        "timeout_seconds": 600,
        "artifacts": ["junit-contract.xml", "allure-results/**"],
    },
    {
        "id": "python-api",
        "paths": ["testframework/tests/api/**"],
        "services": ["isolated-application-stack"],
        "timeout_seconds": 1200,
        "artifacts": ["junit-api.xml", "allure-results/**", "application.log"],
    },
    {
        "id": "python-integration",
        "paths": ["testframework/tests/integration/**"],
        "services": ["declared-isolated-dependencies"],
        "timeout_seconds": 1200,
        "artifacts": ["junit-integration.xml", "allure-results/**", "service-logs/**"],
    },
    {
        "id": "python-e2e-smoke",
        "paths": ["testframework/tests/e2e/**"],
        "services": ["isolated-full-app", "chromium"],
        "timeout_seconds": 900,
        "artifacts": [
            "junit-e2e-smoke.xml",
            "allure-results/**",
            "playwright-traces/**",
            "screenshots/**",
        ],
    },
    {
        "id": "python-e2e-full",
        "paths": ["testframework/tests/e2e/**"],
        "services": ["isolated-full-app", "chromium"],
        "timeout_seconds": 1800,
        "artifacts": [
            "junit-e2e-full.xml",
            "allure-results/**",
            "playwright-traces/**",
            "screenshots/**",
            "application.log",
        ],
    },
    {
        "id": "python-load",
        "paths": ["testframework/load/**"],
        "services": ["isolated-full-stack"],
        "timeout_seconds": 3600,
        "artifacts": ["locust-stats.csv", "locust-report.html", "service-logs/**"],
    },
    {
        "id": "test-architecture-contract",
        "paths": [
            ".planning/decisions/ADR-TEST-ARCHITECTURE-REBASE-2026.md",
            ".planning/audits/test-rebase/test-architecture-contract.json",
            ".planning/audits/test-rebase/founder-greenfield-cutover-manifest.json",
            ".planning/audits/test-rebase/inventory.json",
        ],
        "services": [],
        "timeout_seconds": 300,
        "artifacts": [
            "architecture-contract.log",
            "architecture-negative-self-test.log",
        ],
    },
]

REQUIRED_MANIFEST_DECISION = {
    "adr_status": "accepted",
    "migration_mode": "greenfield-rebuild",
    "row_parity_required": False,
    "dual_run_required": False,
    "go_coverage_migration_blocker": False,
    "active_plan_rebased": True,
}
REQUIRED_RECONCILIATION = {
    "move": {"rows": 507, "declared_tests": 1673},
    "keep_blocked": {"rows": 245, "declared_tests": 1428},
    "delete": {"rows": 297, "declared_tests": 1145},
    "total": {"rows": 1049, "declared_tests": 4246},
}
MANIFEST_ACTIONS = {
    "MOVE-GO-UNIT-HOOK-FREE": "move",
    "KEEP-GO-UNIT-HOOK-DEPENDENT": "keep",
    "MOVE-FRONTEND-UNIT": "move",
    "MOVE-FRONTEND-RENDER-SUPPORT": "move",
    "DELETE-GO-COMPONENT": "delete",
    "DELETE-FRONTEND-STATIC-CONTRACT": "delete",
    "DELETE-FRONTEND-COMPILE-FIXTURES": "delete",
    "DELETE-PLAYWRIGHT-TS-SPECS": "delete",
    "DELETE-PLAYWRIGHT-TS-SUPPORT": "delete",
}

FILE_ATTRIBUTE_REPARSE_POINT = getattr(stat, "FILE_ATTRIBUTE_REPARSE_POINT", 0x400)
WINDOWS_RESERVED_NAMES = {
    "CON",
    "PRN",
    "AUX",
    "NUL",
    *(f"COM{number}" for number in range(1, 10)),
    *(f"LPT{number}" for number in range(1, 10)),
    *(f"COM{number}" for number in ("¹", "²", "³")),
    *(f"LPT{number}" for number in ("¹", "²", "³")),
}


def read_bounded_bytes(path: Path, max_bytes: int) -> bytes:
    try:
        size = path.stat().st_size
    except OSError as error:
        raise ValueError(f"metadata unavailable for {path}: {error}") from error
    if size > max_bytes:
        raise ValueError(f"{path} exceeds {max_bytes} byte input budget")
    with path.open("rb") as stream:
        value = stream.read(max_bytes + 1)
    if len(value) > max_bytes:
        raise ValueError(f"{path} exceeded {max_bytes} byte input budget while reading")
    return value


def reject_duplicate_json_object(pairs: list[tuple[str, Any]]) -> dict[str, Any]:
    result: dict[str, Any] = {}
    for key, value in pairs:
        if key in result:
            raise ValueError(f"duplicate JSON key rejected: {key!r}")
        result[key] = value
    return result


def decode_json_object(payload: bytes, label: str) -> dict[str, Any]:
    value = json.loads(payload, object_pairs_hook=reject_duplicate_json_object)
    if not isinstance(value, dict):
        raise ValueError(f"{label} must contain a JSON object")
    return value


def read_json(path: Path, label: str) -> dict[str, Any]:
    return decode_json_object(read_repo_authority(path, MAX_JSON_BYTES, label), label)


def canonical_path_errors(value: Any, label: str) -> list[str]:
    if not isinstance(value, str):
        return [f"{label} must be a string"]
    if not value:
        return [f"{label} path is empty"]
    if "\\" in value:
        return [f"{label} contains a backslash"]
    if value.startswith("//"):
        return [f"{label} is a UNC path"]
    if value.startswith("/"):
        return [f"{label} is an absolute path"]
    if re.match(r"^[A-Za-z]:", value):
        return [f"{label} is drive-qualified"]
    if unicodedata.normalize("NFC", value) != value:
        return [f"{label} is not Unicode-NFC canonical"]
    if any(ord(character) < 32 or ord(character) == 127 for character in value):
        return [f"{label} contains a control character"]
    parts = value.split("/")
    if len(parts) > MAX_PATH_COMPONENTS:
        return [f"{label} exceeds {MAX_PATH_COMPONENTS} component traversal budget"]
    if len(value.encode("utf-8")) > MAX_PATH_UTF8_BYTES:
        return [f"{label} exceeds {MAX_PATH_UTF8_BYTES} UTF-8 byte budget"]
    if any(len(part.encode("utf-8")) > MAX_COMPONENT_UTF8_BYTES for part in parts):
        return [f"{label} component exceeds {MAX_COMPONENT_UTF8_BYTES} UTF-8 byte budget"]
    if "" in parts:
        return [f"{label} contains an empty path segment"]
    if "." in parts:
        return [f"{label} contains a dot path segment"]
    if ".." in parts:
        return [f"{label} contains dotdot traversal"]
    for part in parts:
        if part.rstrip(" .") != part:
            return [f"{label} contains a Windows-aliased trailing dot/space segment"]
        if ":" in part:
            return [f"{label} contains a colon/alternate-data-stream segment"]
        stem = part.split(".", 1)[0].upper()
        if stem in WINDOWS_RESERVED_NAMES:
            return [f"{label} contains reserved Windows name {part!r}"]
    if PurePosixPath(value).as_posix() != value:
        return [f"{label} is not canonical POSIX-relative"]
    return []


def portable_path_key(value: str) -> str:
    """Return the collision key shared by Windows and portable cutover batches."""

    return "/".join(
        unicodedata.normalize("NFC", part).casefold().rstrip(" .")
        for part in value.split("/")
    )


def portable_alias_errors(paths: list[str], label: str) -> list[str]:
    aliases: dict[str, list[str]] = {}
    for value in paths:
        if canonical_path_errors(value, label):
            continue
        aliases.setdefault(portable_path_key(value), []).append(value)
    collisions = [
        sorted(values)
        for values in aliases.values()
        if len(values) > 1
    ]
    if not collisions:
        return []
    collisions.sort()
    return [f"{label} portable path alias/case collision: {collisions[:5]!r}"]


def is_symlink_or_reparse(metadata: os.stat_result) -> bool:
    attributes = int(getattr(metadata, "st_file_attributes", 0))
    return stat.S_ISLNK(metadata.st_mode) or bool(
        attributes & FILE_ATTRIBUTE_REPARSE_POINT
    )


def metadata_signature(metadata: os.stat_result, *, include_content: bool) -> tuple[int, ...]:
    signature = (
        int(metadata.st_mode),
        int(metadata.st_dev),
        int(metadata.st_ino),
        int(metadata.st_nlink),
        int(getattr(metadata, "st_file_attributes", 0)),
    )
    if include_content:
        signature += (
            int(metadata.st_size),
            int(getattr(metadata, "st_mtime_ns", int(metadata.st_mtime * 1_000_000_000))),
            int(getattr(metadata, "st_ctime_ns", int(metadata.st_ctime * 1_000_000_000))),
        )
    return signature


def open_physical_root(path: Path, flags: int, label: str) -> tuple[int, os.stat_result]:
    """Open every absolute root component descriptor-relative and no-follow."""

    absolute = Path(os.path.abspath(path))
    if (
        len(absolute.parts) - 1 > MAX_PATH_COMPONENTS
        or len(str(absolute).encode("utf-8")) > MAX_PATH_UTF8_BYTES
        or any(len(part.encode("utf-8")) > MAX_COMPONENT_UTF8_BYTES for part in absolute.parts[1:])
    ):
        raise ValueError(f"{label} repository root exceeds component/path budget")
    protected = os.path.normcase(os.path.abspath(REPO_ROOT / "pro-qa-ai"))
    candidate = os.path.normcase(str(absolute))
    try:
        inside_protected = os.path.commonpath([candidate, protected]) == protected
    except ValueError:
        inside_protected = False
    if inside_protected:
        raise ValueError(f"{label} repository root enters protected pro-qa-ai")
    anchor = Path(absolute.anchor)
    descriptor = os.open(anchor, flags)
    try:
        for part in absolute.parts[1:]:
            before = os.stat(part, dir_fd=descriptor, follow_symlinks=False)
            if is_symlink_or_reparse(before) or not stat.S_ISDIR(before.st_mode):
                raise ValueError(f"{label} repository ancestor is not physical")
            next_fd = os.open(part, flags, dir_fd=descriptor)
            opened = os.fstat(next_fd)
            if metadata_signature(before, include_content=False) != metadata_signature(opened, include_content=False):
                os.close(next_fd)
                raise ValueError(f"{label} repository ancestor changed while opening")
            os.close(descriptor)
            descriptor = next_fd
        return descriptor, os.fstat(descriptor)
    except Exception:
        os.close(descriptor)
        raise


def descriptor_traversal_error() -> str | None:
    required_flags = ("O_CLOEXEC", "O_DIRECTORY", "O_NOFOLLOW", "O_PATH")
    missing = [name for name in required_flags if not hasattr(os, name)]
    supports_dir_fd = os.open in getattr(os, "supports_dir_fd", set())
    if os.name == "posix" and not missing and supports_dir_fd:
        return None
    return (
        "descriptor-relative O_NOFOLLOW traversal is unavailable; destructive "
        f"authority fails closed (os={os.name} missing={missing} dir_fd={supports_dir_fd})"
    )


def descriptor_metadata_chain(
    repo_root: Path,
    parts: list[str],
    label: str,
) -> tuple[list[tuple[Path, os.stat_result]], Path | None, list[str]]:
    """Traverse existing components using held directory descriptors."""

    unsupported = descriptor_traversal_error()
    if unsupported:
        return [], None, [f"{label} {unsupported}"]
    root = Path(os.path.abspath(repo_root))
    directory_flags = os.O_PATH | os.O_CLOEXEC | os.O_DIRECTORY | os.O_NOFOLLOW
    entry_flags = os.O_PATH | os.O_CLOEXEC | os.O_NOFOLLOW
    descriptors: list[int] = []
    snapshots: list[tuple[Path, os.stat_result]] = []
    current = root
    try:
        try:
            root_descriptor, root_opened = open_physical_root(root, directory_flags, label)
        except OSError as error:
            return [], None, [f"{label} repository root metadata/open failed: {error}"]
        except ValueError as error:
            return [], None, [str(error)]
        descriptors.append(root_descriptor)
        snapshots.append((root, root_opened))

        for index, part in enumerate(parts):
            is_leaf = index == len(parts) - 1
            current = current / part
            try:
                before = os.stat(part, dir_fd=descriptors[-1], follow_symlinks=False)
            except FileNotFoundError:
                return snapshots, current, []
            except OSError as error:
                return snapshots, None, [f"{label} metadata inaccessible for {current}: {error}"]
            if is_symlink_or_reparse(before):
                return snapshots, None, [
                    f"{label} symlink/reparse component rejected: {current}"
                ]
            flags = entry_flags if is_leaf else directory_flags
            try:
                descriptor = os.open(part, flags, dir_fd=descriptors[-1])
            except OSError as error:
                return snapshots, None, [
                    f"{label} descriptor-relative open rejected {current}: {error}"
                ]
            descriptors.append(descriptor)
            opened = os.fstat(descriptor)
            if metadata_signature(before, include_content=is_leaf) != metadata_signature(
                opened, include_content=is_leaf
            ):
                return snapshots, None, [f"{label} metadata changed while opening {current}"]
            if not is_leaf and not stat.S_ISDIR(opened.st_mode):
                return snapshots, None, [f"{label} ancestor is not a directory: {current}"]
            snapshots.append((current, opened))
        return snapshots, None, []
    finally:
        for descriptor in reversed(descriptors):
            try:
                os.close(descriptor)
            except OSError:
                pass


def authority_relative_path(path: Path, label: str) -> str:
    """Reject external/protected CLI inputs before opening any component."""

    root = Path(os.path.abspath(REPO_ROOT))
    candidate = Path(os.path.abspath(path))
    try:
        relative = candidate.relative_to(root).as_posix()
    except ValueError as error:
        raise ValueError(f"{label} must remain inside repository root") from error
    lexical = canonical_path_errors(relative, label)
    if lexical:
        raise ValueError(lexical[0])
    portable_relative = portable_path_key(relative)
    if portable_relative == "pro-qa-ai" or portable_relative.startswith("pro-qa-ai/"):
        raise ValueError(f"{label} enters protected pro-qa-ai")
    return relative


def _read_authority_once(relative: str, max_bytes: int, label: str) -> tuple[bytes, list[tuple[int, ...]]]:
    unsupported = descriptor_traversal_error()
    if unsupported:
        raise ValueError(f"{label} {unsupported}")
    root = Path(os.path.abspath(REPO_ROOT))
    directory_flags = os.O_PATH | os.O_CLOEXEC | os.O_DIRECTORY | os.O_NOFOLLOW
    file_flags = os.O_RDONLY | os.O_CLOEXEC | os.O_NOFOLLOW | getattr(os, "O_NONBLOCK", 0)
    descriptors: list[int] = []
    signatures: list[tuple[int, ...]] = []
    try:
        root_fd, root_opened = open_physical_root(root, directory_flags, label)
        descriptors.append(root_fd)
        signatures.append(metadata_signature(root_opened, include_content=False))
        parts = relative.split("/")
        for index, part in enumerate(parts):
            leaf = index == len(parts) - 1
            before = os.stat(part, dir_fd=descriptors[-1], follow_symlinks=False)
            if is_symlink_or_reparse(before):
                raise ValueError(f"{label} symlink/reparse component rejected: {relative}")
            descriptor = os.open(part, file_flags if leaf else directory_flags, dir_fd=descriptors[-1])
            descriptors.append(descriptor)
            opened = os.fstat(descriptor)
            include_content = leaf
            if metadata_signature(before, include_content=include_content) != metadata_signature(opened, include_content=include_content):
                raise ValueError(f"{label} identity changed while opening: {relative}")
            if leaf:
                if not stat.S_ISREG(opened.st_mode) or int(opened.st_nlink) != 1:
                    raise ValueError(f"{label} must be a single-link regular file")
                if int(opened.st_size) > max_bytes:
                    raise ValueError(f"{label} exceeds {max_bytes} byte input budget")
            elif not stat.S_ISDIR(opened.st_mode):
                raise ValueError(f"{label} non-directory ancestor rejected: {relative}")
            signatures.append(metadata_signature(opened, include_content=include_content))
        leaf_fd = descriptors[-1]
        chunks: list[bytes] = []
        size = 0
        while True:
            chunk = os.read(leaf_fd, min(1024 * 1024, max_bytes + 1 - size))
            if not chunk:
                break
            chunks.append(chunk)
            size += len(chunk)
            if size > max_bytes:
                raise ValueError(f"{label} exceeded {max_bytes} byte input budget while reading")
        after = os.fstat(leaf_fd)
        if signatures[-1] != metadata_signature(after, include_content=True):
            raise ValueError(f"{label} changed during descriptor read")
        return b"".join(chunks), signatures
    finally:
        for descriptor in reversed(descriptors):
            try:
                os.close(descriptor)
            except OSError:
                pass


def read_repo_authority(path: Path, max_bytes: int, label: str) -> bytes:
    relative = authority_relative_path(path, label)
    payload, before = _read_authority_once(relative, max_bytes, label)
    confirmation, after = _read_authority_once(relative, max_bytes, label)
    if payload != confirmation or before != after:
        raise ValueError(f"{label} identity changed during independent recheck")
    return payload


def inspect_repo_path(
    repo_root: Path,
    relative: str,
    label: str,
    *,
    require_leaf: bool,
    require_regular_file: bool,
    reject_existing_leaf: bool,
    safe_directory_cache: dict[Path, tuple[int, ...]] | None = None,
    missing_component_cache: set[Path] | None = None,
) -> tuple[list[str], tuple[int, int] | None]:
    """Inspect components with openat/no-follow and an independent recheck."""

    lexical_errors = canonical_path_errors(relative, label)
    if lexical_errors:
        return lexical_errors, None

    errors: list[str] = []
    parts = relative.split("/")
    snapshots, first_missing, traversal_errors = descriptor_metadata_chain(
        repo_root, parts, label
    )
    errors.extend(traversal_errors)
    leaf_metadata = snapshots[-1][1] if not first_missing and not traversal_errors else None
    if first_missing is not None:
        if require_leaf:
            errors.append(f"{label} required path/component is missing: {first_missing}")
        if missing_component_cache is not None:
            missing_component_cache.add(first_missing)
    elif leaf_metadata is not None:
        if reject_existing_leaf:
            errors.append(f"{label} existing destination collision: {relative}")
        elif require_regular_file and not stat.S_ISREG(leaf_metadata.st_mode):
            errors.append(f"{label} source is not a regular file: {relative}")
        elif require_regular_file and int(leaf_metadata.st_nlink) != 1:
            errors.append(
                f"{label} destructive regular source link count must be one: "
                f"{relative} (st_nlink={leaf_metadata.st_nlink})"
            )

    snapshots_after, missing_after, recheck_errors = descriptor_metadata_chain(
        repo_root, parts, label
    )
    errors.extend(recheck_errors)
    before_signatures = [
        (
            path,
            metadata_signature(
                metadata,
                include_content=first_missing is None and index == len(snapshots) - 1,
            ),
        )
        for index, (path, metadata) in enumerate(snapshots)
    ]
    after_signatures = [
        (
            path,
            metadata_signature(
                metadata,
                include_content=missing_after is None and index == len(snapshots_after) - 1,
            ),
        )
        for index, (path, metadata) in enumerate(snapshots_after)
    ]
    if first_missing != missing_after or before_signatures != after_signatures:
        errors.append(f"{label} metadata changed during descriptor recheck (TOCTOU): {relative}")

    if safe_directory_cache is not None:
        observed_directories = snapshots[:-1] if first_missing is None else snapshots
        for path, metadata in observed_directories:
            signature = metadata_signature(metadata, include_content=False)
            previous = safe_directory_cache.setdefault(path, signature)
            if previous != signature:
                errors.append(
                    f"{label} directory identity changed during validation (TOCTOU): {path}"
                )

    identity = None
    if leaf_metadata is not None and stat.S_ISREG(leaf_metadata.st_mode):
        identity = (int(leaf_metadata.st_dev), int(leaf_metadata.st_ino))
    return errors, identity


def source_digest(paths: list[str]) -> str:
    payload = "\n".join(sorted(paths)) + "\n"
    return hashlib.sha256(payload.encode("utf-8")).hexdigest()


def row_matches(row: dict[str, Any], selector: dict[str, Any]) -> bool:
    for key, expected in selector.items():
        if key.endswith("_in"):
            if row.get(key[:-3]) not in expected:
                return False
        elif key == "hook_consumers_count":
            consumers = row.get("hook_consumers")
            if not isinstance(consumers, list):
                return False
            count = len(consumers)
            if expected == ">0":
                if count <= 0:
                    return False
            elif count != expected:
                return False
        elif row.get(key) != expected:
            return False
    return True


def validate_python_architecture(value: Any) -> list[str]:
    if not isinstance(value, dict):
        return ["python architecture must be an object"]
    errors: list[str] = []
    if value != REQUIRED_PYTHON_ARCHITECTURE:
        errors.append("python architecture differs from approved progressive tree")
    path_values: list[str] = []
    for key in (
        "installable_package_root",
        "framework_contract_root",
        "case_root",
        "schema_root",
        "policy_root",
        "docker_root",
        "runner_root",
        "load_root",
        "reports_root",
    ):
        candidate = value.get(key)
        if isinstance(candidate, str):
            path_values.append(candidate)
    for key in ("executable_suite_roots", "reusable_module_roots"):
        candidates = value.get(key)
        if isinstance(candidates, list):
            path_values.extend(item for item in candidates if isinstance(item, str))
    composition = value.get("composition")
    if isinstance(composition, dict) and isinstance(composition.get("path"), str):
        path_values.append(composition["path"])
    for index, path in enumerate(path_values):
        probe = path.replace("<domain>", "sentinel").replace("<task-run>", "sentinel")
        errors.extend(canonical_path_errors(probe.rstrip("/"), f"python architecture path {index}"))
    forbidden = set(REQUIRED_FORBIDDEN_PYTHON_PATHS)
    for path in path_values:
        if path in forbidden:
            errors.append(f"python architecture uses forbidden physical path: {path}")
    suites = value.get("executable_suite_roots")
    if isinstance(suites, list) and any(
        isinstance(path, str) and not path.startswith("testframework/tests/")
        for path in suites
    ):
        errors.append("executable suite root escapes testframework/tests")
    reusable = value.get("reusable_module_roots")
    if isinstance(reusable, list) and any(
        isinstance(path, str) and not path.startswith("testframework/src/hub_testkit/")
        for path in reusable
    ):
        errors.append("reusable module root escapes testframework/src/hub_testkit")
    if not isinstance(composition, dict) or composition.get("path") != "testframework/tests/conftest.py":
        errors.append("composition path must be testframework/tests/conftest.py")
    elif composition.get("responsibility") != "explicit-plugin-and-fixture-composition-only":
        errors.append("testframework/tests/conftest.py must remain composition-only")
    materialization = value.get("materialization")
    if not isinstance(materialization, dict):
        errors.append("progressive materialization contract is missing")
    elif materialization.get("empty_directories_required") is not False or materialization.get("speculative_layers_required") is not False:
        errors.append("empty or speculative framework layers cannot be required")
    return errors


def validate_contract(contract: dict[str, Any], adr_text: str) -> list[str]:
    errors: list[str] = []
    exact_sections = (
        ("decision", REQUIRED_DECISION),
        ("baseline_e2e", REQUIRED_E2E_BASELINE),
        ("roots", REQUIRED_ROOTS),
        ("founder_manifest", REQUIRED_FOUNDER_MANIFEST),
        ("python_architecture", REQUIRED_PYTHON_ARCHITECTURE),
        ("forbidden_python_physical_paths", REQUIRED_FORBIDDEN_PYTHON_PATHS),
        ("cutover_snapshot", REQUIRED_CUTOVER_SNAPSHOT),
        ("unit_eligibility", REQUIRED_UNIT_ELIGIBILITY),
        ("greenfield_python", REQUIRED_GREENFIELD),
        ("coverage", REQUIRED_COVERAGE),
        ("skip_policy", REQUIRED_SKIP_POLICY),
        ("jobs", REQUIRED_JOBS),
        ("cutover_deletion_rollback", REQUIRED_CUTOVER),
        ("required_adr_markers", REQUIRED_MARKERS),
    )
    expected_top_level_keys = {"schema_version"} | {
        key for key, _ in exact_sections
    }
    if set(contract) != expected_top_level_keys:
        errors.append("contract top-level keys differ from the exact allowed set")
    if contract.get("schema_version") != 4:
        errors.append("contract schema_version must be 4")
    for key, expected in exact_sections:
        if contract.get(key) != expected:
            errors.append(f"exact contract section changed: {key}")
    errors.extend(validate_python_architecture(contract.get("python_architecture")))

    for marker in REQUIRED_MARKERS:
        if marker not in adr_text:
            errors.append(f"ADR is missing required marker: {marker}")
    if "PARITY_ACCEPTED" in adr_text or "GO_COVERAGE_MIGRATION_BLOCKER=true" in adr_text:
        errors.append("ADR reintroduced parity or Go coverage migration authority")
    if "This ADR is intentionally **not Accepted**" in adr_text:
        errors.append("ADR still claims to be blocked/proposed")
    if "Phase 129" not in adr_text or "cannot be marked complete" not in adr_text:
        errors.append("ADR is missing the Phase 129 blocked-unit completion gate")
    return errors


def validate_manifest(
    manifest: dict[str, Any], inventory_sha256: str
) -> list[str]:
    errors: list[str] = []
    if manifest.get("status") != "verified-hook-blocked":
        errors.append("founder manifest status must be verified-hook-blocked")
    if manifest.get("decision_contract") != REQUIRED_MANIFEST_DECISION:
        errors.append("founder manifest decision contract changed")

    ledger = manifest.get("source_ledger")
    if not isinstance(ledger, dict):
        errors.append("founder manifest source_ledger must be an object")
    else:
        required_ledger = {
            "path": ".planning/audits/test-rebase/inventory.json",
            "sha256": LEDGER_SHA256,
            "source_revision": LEDGER_REVISION,
            "observed_head": LEDGER_REVISION,
            "rows": 1049,
            "declared_tests": 4246,
            "live_drift": {
                "status": "green-docker-check-self-test",
                "mismatched_source_count": 0,
                "mismatched_source_paths": [],
            },
        }
        if ledger != required_ledger:
            errors.append("founder manifest source ledger changed")
        if ledger.get("sha256") != inventory_sha256:
            errors.append("founder manifest ledger SHA does not match inventory bytes")

    if manifest.get("row_reconciliation") != REQUIRED_RECONCILIATION:
        errors.append("founder manifest row reconciliation changed")

    partitions = manifest.get("row_partitions")
    if not isinstance(partitions, list):
        errors.append("founder manifest row_partitions must be a list")
    else:
        actual: dict[str, dict[str, Any]] = {}
        for item in partitions:
            if not isinstance(item, dict) or not isinstance(item.get("id"), str):
                errors.append("founder manifest partition must have a string id")
                continue
            part_id = item["id"]
            if part_id in actual:
                errors.append(f"duplicate founder manifest partition {part_id}")
                continue
            actual[part_id] = item
        if set(actual) != {group["id"] for group in REQUIRED_GROUPS}:
            errors.append("founder manifest partition IDs changed")
        for group in REQUIRED_GROUPS:
            item = actual.get(group["id"])
            if item is None:
                continue
            required_view = {
                "action": MANIFEST_ACTIONS[group["id"]],
                "selector": group["selector"],
                "source_rows": group["expected_count"],
                "declared_tests": group["expected_declared_tests"],
            }
            for key, expected in required_view.items():
                if item.get(key) != expected:
                    errors.append(f"founder manifest partition changed: {group['id']}:{key}")

    delete_paths = manifest.get("non_ledger_delete_paths")
    if not isinstance(delete_paths, list):
        errors.append("founder manifest non-ledger delete paths must be a list")
    else:
        for index, path in enumerate(delete_paths):
            errors.extend(canonical_path_errors(path, f"non-ledger delete path {index}"))
        if len(delete_paths) != 28 or len(set(delete_paths)) != 28:
            errors.append("non-ledger delete closure must contain 28 unique paths")
        if source_digest(delete_paths) != NON_LEDGER_SHA256:
            errors.append("non-ledger delete closure SHA-256 changed")

    blocker_ids = [
        item.get("id") for item in manifest.get("blockers", []) if isinstance(item, dict)
    ]
    if blocker_ids != ["B3-HOOK-DEPENDENT-UNITS"]:
        errors.append("founder manifest blockers must contain only B3 hook-dependent units")
    return errors


def validate_inventory(
    inventory: dict[str, Any],
    manifest: dict[str, Any],
    *,
    forced_movable_sources: set[str] | None = None,
) -> list[str]:
    errors: list[str] = []
    if inventory.get("revision") != LEDGER_REVISION:
        errors.append("inventory revision changed")
    rows = inventory.get("rows")
    if not isinstance(rows, list):
        return ["inventory rows must be a list"]
    if len(rows) > MAX_INVENTORY_ROWS:
        return [f"inventory rows exceed {MAX_INVENTORY_ROWS} row budget: {len(rows)}"]
    if len(rows) != 1049:
        errors.append(f"inventory must contain exactly 1,049 rows; got {len(rows)}")

    source_paths: list[str] = []
    row_by_source: dict[str, dict[str, Any]] = {}
    membership: dict[str, list[dict[str, Any]]] = {group["id"]: [] for group in REQUIRED_GROUPS}
    for index, row in enumerate(rows):
        if not isinstance(row, dict):
            errors.append(f"row {index} must be an object")
            continue
        source = row.get("source_path")
        path_errors = canonical_path_errors(source, f"row {index} source_path")
        errors.extend(path_errors)
        if path_errors:
            continue
        if source.startswith("pro-qa-ai/"):
            errors.append(f"protected source entered inventory: {source}")
        source_paths.append(source)
        row_by_source[source] = row

        target_class = row.get("target_class")
        if target_class not in REQUIRED_CLASS_RUNNERS:
            errors.append(f"{source!r} has unsupported target_class {target_class!r}")
            continue
        if row.get("runner") != REQUIRED_CLASS_RUNNERS[target_class]:
            errors.append(f"{source!r} has invalid runner for {target_class}")

        matched = [group for group in REQUIRED_GROUPS if row_matches(row, group["selector"])]
        if len(matched) != 1:
            errors.append(f"{source!r} must match exactly one cutover group; got {len(matched)}")
        else:
            membership[matched[0]["id"]].append(row)

        if target_class in {"python-api", "python-integration"} and not source.startswith(
            "backend/testframework/"
        ):
            errors.append(f"{source!r} Go component escaped backend/testframework")
        if target_class in {"python-static-contract", "frontend-unit", "frontend-support"} and not source.startswith(
            "testframework/stages/unit/frontend-test/"
        ):
            errors.append(f"{source!r} frontend source escaped the sealed root")
        if target_class in {"python-e2e", "python-e2e-support"} and not source.startswith(
            "testframework/stages/e2e/"
        ):
            errors.append(f"{source!r} Playwright source escaped the sealed root")

    duplicates = sorted(path for path, count in Counter(source_paths).items() if count > 1)
    if duplicates:
        errors.append(f"duplicate source paths: {duplicates[:5]!r}")
    if len(row_by_source) != len(source_paths):
        errors.append("source path map is not one-to-one")
    if sum(int(row.get("test_count", 0)) for row in rows if isinstance(row, dict)) != 4246:
        errors.append("inventory declared-test total must be 4,246")

    assigned = sum(len(value) for value in membership.values())
    if assigned != 1049:
        errors.append(f"cutover groups must partition 1,049 rows; got {assigned}")

    for group in REQUIRED_GROUPS:
        selected = membership[group["id"]]
        count = len(selected)
        declared = sum(int(row.get("test_count", 0)) for row in selected)
        paths = [str(row["source_path"]) for row in selected]
        if count != group["expected_count"]:
            errors.append(f"group {group['id']} count changed: {count}")
        if declared != group["expected_declared_tests"]:
            errors.append(f"group {group['id']} declared-test count changed: {declared}")
        if source_digest(paths) != group["source_paths_sha256"]:
            errors.append(f"group {group['id']} source-path SHA-256 changed")

    movable_go = membership["MOVE-GO-UNIT-HOOK-FREE"]
    forced = forced_movable_sources or set()
    movable_sources = {str(row["source_path"]) for row in movable_go} | forced
    for source in sorted(movable_sources):
        row = row_by_source.get(source)
        if row is None:
            errors.append(f"forced movable source is absent: {source}")
            continue
        hooks = row.get("hook_consumers")
        if hooks:
            errors.append(f"hook-dependent Go row admitted to movable: {source}")
        if row.get("target_class") != "go-unit":
            errors.append(f"non-Go-unit row admitted to Go movable: {source}")
        target = row.get("target_path")
        target_errors = canonical_path_errors(target, f"{source!r} target_path")
        errors.extend(target_errors)
        if not target_errors and GO_UNIT_TARGET.fullmatch(target) is None:
            errors.append(f"{source!r} has invalid Go unit destination {target!r}")
        dependencies = " ".join(str(item).lower() for item in row.get("dependency_hints", []))
        if any(fragment in dependencies for fragment in EXTERNAL_GO_DEPENDENCY_FRAGMENTS):
            errors.append(f"{source!r} movable Go unit has an external harness dependency")

    blocked = membership["KEEP-GO-UNIT-HOOK-DEPENDENT"]
    hook_links = 0
    hook_files: set[str] = set()
    hook_symbols: set[str] = set()
    for row in blocked:
        for consumer in row.get("hook_consumers", []):
            if not isinstance(consumer, dict):
                errors.append(f"{row['source_path']!r} has malformed hook consumer")
                continue
            hook_source = consumer.get("hook_source")
            symbol = consumer.get("symbol")
            if not isinstance(hook_source, str) or not isinstance(symbol, str):
                errors.append(f"{row['source_path']!r} has malformed hook link")
                continue
            hook_links += 1
            hook_files.add(hook_source)
            hook_symbols.add(f"{hook_source}::{symbol}")
    if (hook_links, len(hook_files), len(hook_symbols)) != (653, 42, 434):
        errors.append("hook-dependent keep partition must remain 653 links / 42 files / 434 symbols")

    frontend_move_ids = ("MOVE-FRONTEND-UNIT", "MOVE-FRONTEND-RENDER-SUPPORT")
    frontend_targets: list[str] = []
    for group_id in frontend_move_ids:
        for row in membership[group_id]:
            target = row.get("target_path")
            target_errors = canonical_path_errors(target, f"{row['source_path']!r} target_path")
            errors.extend(target_errors)
            if target_errors:
                continue
            if group_id == "MOVE-FRONTEND-UNIT" and FRONTEND_UNIT_TARGET.fullmatch(target) is None:
                errors.append(f"{row['source_path']!r} has invalid frontend unit destination")
            if group_id == "MOVE-FRONTEND-RENDER-SUPPORT" and target != "frontend/tests/support/render.tsx":
                errors.append("render support destination changed")
            frontend_targets.append(target)

    move_targets = [str(row.get("target_path")) for row in movable_go] + frontend_targets
    if len(move_targets) != 507 or len(set(move_targets)) != 507:
        errors.append("immediate move targets must be 507 unique paths")

    delete_paths = manifest.get("non_ledger_delete_paths", [])
    if isinstance(delete_paths, list) and set(delete_paths).intersection(source_paths):
        errors.append("non-ledger deletion closure overlaps ledger rows")
    return errors


def validate_filesystem(
    inventory: dict[str, Any],
    manifest: dict[str, Any],
    *,
    repo_root: Path = REPO_ROOT,
) -> list[str]:
    """Validate the physical cutover boundary, not only its JSON spelling."""

    errors: list[str] = []
    rows = inventory.get("rows")
    if not isinstance(rows, list):
        return ["filesystem guard requires inventory rows"]
    if len(rows) > MAX_INVENTORY_ROWS:
        return [
            f"filesystem guard rows exceed {MAX_INVENTORY_ROWS} row budget: {len(rows)}"
        ]

    source_paths = [
        str(row.get("source_path"))
        for row in rows
        if isinstance(row, dict) and isinstance(row.get("source_path"), str)
    ]
    delete_paths_raw = manifest.get("non_ledger_delete_paths")
    if isinstance(delete_paths_raw, list) and len(delete_paths_raw) > MAX_DELETE_PATHS:
        return [
            f"filesystem guard delete paths exceed {MAX_DELETE_PATHS} path budget: "
            f"{len(delete_paths_raw)}"
        ]
    delete_paths = (
        [str(path) for path in delete_paths_raw if isinstance(path, str)]
        if isinstance(delete_paths_raw, list)
        else []
    )
    move_targets: list[str] = []
    for row in rows:
        if not isinstance(row, dict):
            continue
        matched = [group for group in REQUIRED_GROUPS if row_matches(row, group["selector"])]
        if len(matched) != 1 or MANIFEST_ACTIONS.get(matched[0]["id"]) != "move":
            continue
        target = row.get("target_path")
        if isinstance(target, str):
            move_targets.append(target)

    errors.extend(portable_alias_errors(source_paths + delete_paths, "destructive source"))
    errors.extend(portable_alias_errors(move_targets, "move destination"))

    identities: dict[tuple[int, int], str] = {}
    safe_directories: dict[Path, tuple[int, ...]] = {}
    missing_components: set[Path] = set()
    for index, source in enumerate(source_paths):
        path_errors, identity = inspect_repo_path(
            repo_root,
            source,
            f"ledger source {index}",
            require_leaf=True,
            require_regular_file=True,
            reject_existing_leaf=False,
            safe_directory_cache=safe_directories,
            missing_component_cache=missing_components,
        )
        errors.extend(path_errors)
        if identity is not None:
            previous = identities.setdefault(identity, source)
            if previous != source:
                errors.append(
                    f"destructive source filesystem alias/hardlink collision: "
                    f"{previous!r} and {source!r}"
                )

    for index, source in enumerate(delete_paths):
        path_errors, identity = inspect_repo_path(
            repo_root,
            source,
            f"non-ledger delete source {index}",
            require_leaf=True,
            require_regular_file=True,
            reject_existing_leaf=False,
            safe_directory_cache=safe_directories,
            missing_component_cache=missing_components,
        )
        errors.extend(path_errors)
        if identity is not None:
            previous = identities.setdefault(identity, source)
            if previous != source:
                errors.append(
                    f"destructive source filesystem alias/hardlink collision: "
                    f"{previous!r} and {source!r}"
                )

    for index, target in enumerate(move_targets):
        path_errors, _ = inspect_repo_path(
            repo_root,
            target,
            f"move destination {index}",
            require_leaf=False,
            require_regular_file=False,
            reject_existing_leaf=True,
            safe_directory_cache=safe_directories,
            missing_component_cache=missing_components,
        )
        errors.extend(path_errors)

    # Reopen every observed directory descriptor-relatively after the full scan.
    root = Path(os.path.abspath(repo_root))
    for directory, before in safe_directories.items():
        try:
            relative_parts = list(directory.relative_to(root).parts)
        except ValueError:
            errors.append(
                f"filesystem guard observed directory escaped root during final recheck: "
                f"{directory}"
            )
            continue
        snapshots, missing, recheck_errors = descriptor_metadata_chain(
            root, relative_parts, "filesystem guard final directory recheck"
        )
        errors.extend(recheck_errors)
        if missing is not None or not snapshots:
            errors.append(
                f"filesystem guard directory disappeared during final recheck: {directory}"
            )
            continue
        after = metadata_signature(snapshots[-1][1], include_content=False)
        if before != after:
            errors.append(
                f"filesystem guard directory identity changed during validation (TOCTOU): "
                f"{directory}"
            )
    for missing in missing_components:
        try:
            relative_parts = list(missing.relative_to(root).parts)
        except ValueError:
            errors.append(
                f"filesystem guard missing destination escaped root during final recheck: "
                f"{missing}"
            )
            continue
        _, still_missing, recheck_errors = descriptor_metadata_chain(
            root, relative_parts, "filesystem guard final missing recheck"
        )
        errors.extend(recheck_errors)
        if still_missing != missing:
            errors.append(
                f"filesystem guard destination component appeared during validation "
                f"(TOCTOU): {missing}"
            )
    return errors


def validate_all(
    contract: dict[str, Any],
    manifest: dict[str, Any],
    inventory: dict[str, Any],
    inventory_sha256: str,
    adr_text: str,
    *,
    repo_root: Path = REPO_ROOT,
) -> list[str]:
    return (
        validate_contract(contract, adr_text)
        + validate_manifest(manifest, inventory_sha256)
        + validate_inventory(inventory, manifest)
        + validate_filesystem(inventory, manifest, repo_root=repo_root)
    )


def print_errors(errors: list[str]) -> None:
    for error in errors:
        print(f"RED: {error}", file=sys.stderr)


def require_sentinel(errors: list[str], fragment: str, name: str) -> str | None:
    if any(fragment in error for error in errors):
        return None
    return f"RED: {name} sentinel was not rejected: {errors!r}"


def run_self_test(
    contract: dict[str, Any],
    manifest: dict[str, Any],
    inventory: dict[str, Any],
    inventory_sha256: str,
    adr_text: str,
    docker_bind_sentinel: Path | None,
) -> int:
    # The wrapper runs the canonical positive gate once before this method.
    # Self-test is intentionally negative-only so it cannot double the costly
    # physical scan or misrepresent sentinel success as a positive baseline.
    failures: list[str] = []

    try:
        decode_json_object(
            b'{"outer":{"value":1,"value":2}}', "duplicate-key sentinel"
        )
    except ValueError as problem:
        if "duplicate JSON key" not in str(problem):
            failures.append("RED: nested duplicate-key sentinel returned wrong failure")
    else:
        failures.append("RED: nested duplicate JSON key was accepted")

    for name, candidate, fragment in (
        ("protected authority", REPO_ROOT / "pro-qa-ai" / "forbidden.json", "protected pro-qa-ai"),
        ("external authority", REPO_ROOT.parent / "outside.json", "inside repository root"),
    ):
        try:
            authority_relative_path(candidate, name)
        except ValueError as problem:
            if fragment not in str(problem):
                failures.append(f"RED: {name} failed for wrong reason: {problem}")
        else:
            failures.append(f"RED: {name} path was not rejected before read")

    proposed = deepcopy(contract)
    proposed["decision"]["status"] = "Proposed"
    failure = require_sentinel(
        validate_contract(proposed, adr_text),
        "exact contract section changed: decision",
        "status downgrade",
    )
    if failure:
        failures.append(failure)

    extra_top_level = deepcopy(contract)
    extra_top_level["unexpected_authority"] = True
    failure = require_sentinel(
        validate_contract(extra_top_level, adr_text),
        "top-level keys differ from the exact allowed set",
        "unexpected top-level contract authority",
    )
    if failure:
        failures.append(failure)

    oversized_rows = {"rows": [{} for _ in range(MAX_INVENTORY_ROWS + 1)]}
    failure = require_sentinel(
        validate_inventory(oversized_rows, manifest),
        "row budget",
        "inventory row budget",
    )
    if failure:
        failures.append(failure)

    hidden_dual = deepcopy(contract)
    hidden_dual["decision"]["dual_run_required"] = True
    failure = require_sentinel(
        validate_contract(hidden_dual, adr_text),
        "exact contract section changed: decision",
        "dual-run reintroduction",
    )
    if failure:
        failures.append(failure)

    coverage = deepcopy(contract)
    coverage["coverage"]["go_statement_threshold_is_migration_blocker"] = True
    failure = require_sentinel(
        validate_contract(coverage, adr_text),
        "exact contract section changed: coverage",
        "82-percent coverage blocker",
    )
    if failure:
        failures.append(failure)

    phase_exit = deepcopy(contract)
    phase_exit["cutover_deletion_rollback"][
        "phase_129_completion_requires_blocked_units_refactored_and_moved"
    ] = False
    failure = require_sentinel(
        validate_contract(phase_exit, adr_text),
        "exact contract section changed: cutover_deletion_rollback",
        "Phase 129 early completion",
    )
    if failure:
        failures.append(failure)

    manifest_sha = deepcopy(manifest)
    manifest_sha["source_ledger"]["sha256"] = "0" * 64
    failure = require_sentinel(
        validate_manifest(manifest_sha, inventory_sha256),
        "founder manifest source ledger changed",
        "manifest ledger substitution",
    )
    if failure:
        failures.append(failure)

    manifest_delete = deepcopy(manifest)
    manifest_delete["non_ledger_delete_paths"].pop()
    failure = require_sentinel(
        validate_manifest(manifest_delete, inventory_sha256),
        "non-ledger delete closure must contain 28 unique paths",
        "non-ledger closure removal",
    )
    if failure:
        failures.append(failure)

    traversal = deepcopy(inventory)
    traversal["rows"][0]["source_path"] = "backend/tests/auth/../escape_test.go"
    failure = require_sentinel(
        validate_inventory(traversal, manifest),
        "dotdot traversal",
        "source traversal",
    )
    if failure:
        failures.append(failure)

    duplicate = deepcopy(inventory)
    duplicate["rows"][1]["source_path"] = duplicate["rows"][0]["source_path"]
    failure = require_sentinel(
        validate_inventory(duplicate, manifest),
        "duplicate source paths",
        "duplicate source",
    )
    if failure:
        failures.append(failure)

    hook_source = next(
        str(row["source_path"])
        for row in inventory["rows"]
        if row.get("target_class") == "go-unit" and row.get("hook_consumers")
    )
    failure = require_sentinel(
        validate_inventory(inventory, manifest, forced_movable_sources={hook_source}),
        "hook-dependent Go row admitted to movable",
        "hook-dependent row forced into movable",
    )
    if failure:
        failures.append(failure)

    hook_erasure = deepcopy(inventory)
    blocked = next(
        row
        for row in hook_erasure["rows"]
        if row.get("target_class") == "go-unit" and row.get("hook_consumers")
    )
    blocked["hook_consumers"] = []
    failure = require_sentinel(
        validate_inventory(hook_erasure, manifest),
        "group MOVE-GO-UNIT-HOOK-FREE count changed",
        "silent hook erasure/reclassification",
    )
    if failure:
        failures.append(failure)

    group_digest = deepcopy(contract)
    group_digest["cutover_snapshot"]["groups"][0]["source_paths_sha256"] = "0" * 64
    failure = require_sentinel(
        validate_contract(group_digest, adr_text),
        "exact contract section changed: cutover_snapshot",
        "sealed group digest substitution",
    )
    if failure:
        failures.append(failure)

    legacy_job = deepcopy(contract)
    legacy_job["jobs"].append({"id": "legacy-parity-authority"})
    failure = require_sentinel(
        validate_contract(legacy_job, adr_text),
        "exact contract section changed: jobs",
        "legacy parity job injection",
    )
    if failure:
        failures.append(failure)

    architecture_positive = validate_python_architecture(
        deepcopy(REQUIRED_PYTHON_ARCHITECTURE)
    )
    if architecture_positive:
        failures.append(
            f"RED: approved progressive Python tree failed structural validation: "
            f"{architecture_positive!r}"
        )

    for name, mutate, fragment in (
        (
            "direct executable layer root",
            lambda value: value["executable_suite_roots"].__setitem__(0, "testframework/api/"),
            "forbidden physical path",
        ),
        (
            "repo-root framework output",
            lambda value: value.__setitem__("installable_package_root", "hub_testkit/"),
            "forbidden physical path",
        ),
        (
            "root conftest",
            lambda value: value["composition"].__setitem__("path", "testframework/conftest.py"),
            "forbidden physical path",
        ),
        (
            "generic core module",
            lambda value: value["reusable_module_roots"].append("testframework/src/hub_testkit/core/"),
            "approved progressive tree",
        ),
        (
            "UI pages warehouse",
            lambda value: value["executable_suite_roots"].append("testframework/tests/e2e/pages/"),
            "forbidden physical path",
        ),
        (
            "speculative empty layers",
            lambda value: value["materialization"].__setitem__("speculative_layers_required", True),
            "empty or speculative framework layers",
        ),
    ):
        architecture_negative = deepcopy(REQUIRED_PYTHON_ARCHITECTURE)
        mutate(architecture_negative)
        failure = require_sentinel(
            validate_python_architecture(architecture_negative), fragment, name
        )
        if failure:
            failures.append(failure)

    alias_errors = portable_alias_errors(
        ["backend/tests/Auth/case_test.go", "backend/tests/auth/case_test.go"],
        "sentinel destination",
    )
    failure = require_sentinel(
        alias_errors,
        "portable path alias/case collision",
        "portable case alias",
    )
    if failure:
        failures.append(failure)

    fake_reparse = type(
        "FakeReparseMetadata",
        (),
        {"st_mode": stat.S_IFREG, "st_file_attributes": FILE_ATTRIBUTE_REPARSE_POINT},
    )()
    if not is_symlink_or_reparse(fake_reparse):
        failures.append("RED: Windows reparse metadata sentinel was not recognized")

    for reserved in ("COM¹", "COM².txt", "COM³", "LPT¹", "LPT².log", "LPT³"):
        failure = require_sentinel(
            canonical_path_errors(f"backend/tests/{reserved}", "reserved-name sentinel"),
            "reserved Windows name",
            f"Windows superscript reserved name {reserved}",
        )
        if failure:
            failures.append(failure)

    for reserved in ("COM1.txt", "NUL.json"):
        failure = require_sentinel(
            canonical_path_errors(f"backend/tests/{reserved}", "reserved-name sentinel"),
            "reserved Windows name",
            f"Windows device extension name {reserved}",
        )
        if failure:
            failures.append(failure)

    component_budget_path = "/".join("segment" for _ in range(MAX_PATH_COMPONENTS + 1))
    failure = require_sentinel(
        canonical_path_errors(component_budget_path, "component-budget sentinel"),
        "component traversal budget",
        "path component budget",
    )
    if failure:
        failures.append(failure)

    def production_filesystem_errors(
        root: Path,
        source: str,
        *,
        target: str = "frontend/tests/sentinel/new.test.tsx",
    ) -> list[str]:
        sentinel_inventory = {
            "revision": LEDGER_REVISION,
            "rows": [
                {
                    "source_path": source,
                    "runner": "go-test",
                    "target_class": "go-unit",
                    "hook_consumers": [],
                    "target_path": target,
                    "test_count": 1,
                    "dependency_hints": [],
                }
            ],
        }
        return validate_all(
            contract,
            manifest,
            sentinel_inventory,
            inventory_sha256,
            adr_text,
            repo_root=root,
        )

    try:
        with tempfile.TemporaryDirectory(prefix="hub-cutover-path-sentinel-") as temp_dir:
            sentinel_root = Path(temp_dir)
            safe_parent = sentinel_root / "legacy" / "auth"
            safe_parent.mkdir(parents=True)
            safe_source = safe_parent / "safe_test.go"
            safe_source.write_text("package auth\n", encoding="utf-8")

            byte_budget_file = sentinel_root / "oversized-input.json"
            byte_budget_file.write_bytes(b"12345")
            try:
                read_bounded_bytes(byte_budget_file, 4)
            except ValueError as error:
                if "byte input budget" not in str(error):
                    failures.append(
                        f"RED: JSON/file-size budget sentinel failed for wrong reason: {error}"
                    )
            else:
                failures.append("RED: JSON/file-size input budget was not enforced")

            link_parent = sentinel_root / "linked-auth"
            link_parent.symlink_to(safe_parent, target_is_directory=True)
            symlink_errors = production_filesystem_errors(
                sentinel_root, "linked-auth/safe_test.go"
            )
            failure = require_sentinel(
                symlink_errors,
                "symlink/reparse component rejected",
                "symlink/reparse source ancestor",
            )
            if failure:
                failures.append(failure)

            destination = sentinel_root / "backend" / "tests" / "auth"
            destination.mkdir(parents=True)
            (destination / "safe_test.go").write_text("collision\n", encoding="utf-8")
            collision_errors = production_filesystem_errors(
                sentinel_root,
                "legacy/auth/safe_test.go",
                target="backend/tests/auth/safe_test.go",
            )
            failure = require_sentinel(
                collision_errors,
                "existing destination collision",
                "existing destination through validate_all",
            )
            if failure:
                failures.append(failure)

            hardlink_parent = sentinel_root / "hardlink"
            hardlink_parent.mkdir()
            hardlink_source = hardlink_parent / "source_test.go"
            hardlink_source.write_text("package hardlink\n", encoding="utf-8")
            os.link(hardlink_source, sentinel_root / "unlisted-external-alias.go")
            hardlink_errors = production_filesystem_errors(
                sentinel_root, "hardlink/source_test.go"
            )
            failure = require_sentinel(
                hardlink_errors,
                "destructive regular source link count must be one",
                "unlisted external hardlink alias through validate_all",
            )
            if failure:
                failures.append(failure)

            absent_errors, _ = inspect_repo_path(
                sentinel_root,
                "frontend/tests/auth/new.test.tsx",
                "absent destination helper sentinel",
                require_leaf=False,
                require_regular_file=False,
                reject_existing_leaf=True,
            )
            if absent_errors:
                failures.append(
                    f"RED: absent physical destination sentinel unexpectedly failed: "
                    f"{absent_errors!r}"
                )

            delete_budget_manifest = deepcopy(manifest)
            delete_budget_manifest["non_ledger_delete_paths"] = [
                f"legacy/delete-{index}.go" for index in range(MAX_DELETE_PATHS + 1)
            ]
            budget_errors = validate_all(
                contract,
                delete_budget_manifest,
                {"revision": LEDGER_REVISION, "rows": []},
                inventory_sha256,
                adr_text,
                repo_root=sentinel_root,
            )
            failure = require_sentinel(
                budget_errors,
                "delete paths exceed",
                "filesystem delete-path budget through validate_all",
            )
            if failure:
                failures.append(failure)
    except OSError as error:
        failures.append(f"RED: physical path sentinel setup failed closed: {error}")

    if docker_bind_sentinel is None:
        failures.append("RED: Docker-bind symlink/reparse sentinel path is required")
    else:
        bind_errors = production_filesystem_errors(
            docker_bind_sentinel, "linked/source_test.go"
        )
        failure = require_sentinel(
            bind_errors,
            "symlink/reparse component rejected",
            "Docker-bind-created symlink/reparse through validate_all",
        )
        if failure:
            failures.append(failure)

    if failures:
        for failure in failures:
            print(failure, file=sys.stderr)
        return 1
    print(
        "GREEN: negative-only self-test (no positive baseline) passed; contract/manifest/"
        "traversal/duplicate/hook/portable-alias/Windows-superscript/budget sentinels and "
        "validate_all-wired tmp+Docker-bind symlink, hardlink and destination sentinels "
        "all produced RED"
    )
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--contract", type=Path, default=DEFAULT_CONTRACT)
    parser.add_argument("--manifest", type=Path, default=DEFAULT_MANIFEST)
    parser.add_argument("--inventory", type=Path, default=DEFAULT_INVENTORY)
    parser.add_argument("--adr", type=Path, default=DEFAULT_ADR)
    parser.add_argument("--self-test", action="store_true")
    parser.add_argument(
        "--docker-bind-sentinel",
        type=Path,
        help="task-owned bind path containing real/source_test.go and linked -> real",
    )
    args = parser.parse_args()
    try:
        contract = read_json(args.contract, "architecture contract")
        manifest = read_json(args.manifest, "founder manifest")
        inventory_raw = read_repo_authority(args.inventory, MAX_JSON_BYTES, "inventory authority")
        inventory = decode_json_object(inventory_raw, "inventory authority")
        inventory_sha256 = hashlib.sha256(inventory_raw).hexdigest()
        adr_text = read_repo_authority(args.adr, MAX_ADR_BYTES, "ADR authority").decode("utf-8")
    except (OSError, UnicodeError, ValueError, json.JSONDecodeError) as error:
        print(f"RED: {error}", file=sys.stderr)
        return 1

    if args.self_test:
        return run_self_test(
            contract,
            manifest,
            inventory,
            inventory_sha256,
            adr_text,
            args.docker_bind_sentinel,
        )
    errors = validate_all(contract, manifest, inventory, inventory_sha256, adr_text)
    if errors:
        print_errors(errors)
        return 1
    print(
        "GREEN: LEDGER_HEAD b9f690da ledger 1,049/4,246; move 507/1,673, keep-block "
        "245/1,428, delete-no-port 297/1,145 plus exact non-ledger closure 28"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
