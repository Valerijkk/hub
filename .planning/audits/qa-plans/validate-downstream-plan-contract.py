#!/usr/bin/env python3
"""Fail-closed static contract check for the 128–141 live-plan authority."""
from __future__ import annotations

import ast
import hashlib
import json
import os
import re
import stat
import subprocess
import sys
from datetime import datetime
from pathlib import PurePosixPath
from typing import Any
import xml.etree.ElementTree as ET

MAX_FILE_BYTES = 2 * 1024 * 1024
MAX_RECORD_BYTES = 256 * 1024
MAX_PHASE_ENTRIES = 32
MAX_PHASE129_REQUIREMENTS_BYTES = 8192
SEALED_PLAN_COMMIT = "30333f1bb5bd10a9e86e9a93a5f71fdcdcc13ddf"
MASTER_PATH = "docs/superpowers/plans/2026-08-08-python-qa-framework-rebase.md"
CUTOVER_MANIFEST_PATH = ".planning/audits/test-rebase/cutover-manifest.json"
RECEIPT_PATH = ".planning/audits/qa-plans/downstream-plan-contract-receipt.v1.json"
BUILDER_PATH = ".planning/audits/test-rebase/build_cutover_manifest.py"
VALIDATOR_PATH = ".planning/audits/test-rebase/validate_cutover_manifest.py"
PHASE129_TOOL_SHA256 = {
    BUILDER_PATH: "7603e51cb44fd117cb3a9f3beb860678d6e44cf24a20d49e0de220b4238c9b42",
    VALIDATOR_PATH: "0b3c29dcca0549e567c39d8ef9d4119bef6439dc6886f94d178877efcb5eda46",
}
POST_G129P_STATE_ARGUMENT = "--post-g129p-manifest"
PLAN_PATHS = (
    ".planning/phases/128-tf-python-scaffold/PLAN.md", ".planning/phases/129-tf-e2e-shared/PLAN.md",
    ".planning/phases/129-tf-e2e-shared/129-01-PLAN.md", ".planning/phases/129-tf-e2e-shared/129-02-PLAN.md",
    ".planning/phases/129-tf-e2e-shared/129-03-PLAN.md", ".planning/phases/129-tf-e2e-shared/129-04-PLAN.md",
    ".planning/phases/129-tf-e2e-shared/129-05-PLAN.md", ".planning/phases/129-tf-e2e-shared/129-06-PLAN.md",
    ".planning/phases/129-tf-e2e-shared/129-07-PLAN.md", ".planning/phases/130-tf-traceability-gates/PLAN.md",
    ".planning/phases/131-tf-python-core-auth/PLAN.md", ".planning/phases/132-tf-security-platform/PLAN.md",
    ".planning/phases/133-tf-infrastructure-integration/PLAN.md", ".planning/phases/134-tf-work-realtime/PLAN.md",
    ".planning/phases/135-tf-business-integrations/PLAN.md", ".planning/phases/136-tf-python-contracts/PLAN.md",
    ".planning/phases/137-tf-python-e2e-foundation/PLAN.md", ".planning/phases/138-tf-python-e2e-journeys/PLAN.md",
    ".planning/phases/139-tf-accessibility-security/PLAN.md", ".planning/phases/140-tf-load-resilience/PLAN.md",
    ".planning/phases/141-tf-ci-final-verification/PLAN.md",
)
PHASE129_NUMBERED_PATHS = PLAN_PATHS[2:9]
PHASE129_ACTIVE_PATHS = (MASTER_PATH, *PLAN_PATHS[1:9])
LATER_COUNTS = (4, 4, 4, 3, 8, 18, 3, 3, 5, 3, 3, 3)
RECEIPT_KEYS = ("schema_version", "full_head", "checker_blob_sha256", "master_blob_sha256", "ordered_plan_paths", "plan_blob_sha256", "task_count", "batch_ids_sha256", "batch_count", "operation_count", "maximum_batch", "forbidden_match_count", "field_coverage", "verdict", "reviewer", "reviewed_at")
RECEIPT_SCHEMA_VERSION = "downstream-plan-contract-receipt.v1"
PENDING_RECEIPT_REVIEWER = "PENDING_INDEPENDENT_GSD_PLAN_CHECKER"
ACCEPTED_RECEIPT_REVIEWER = "gsd-plan-checker"
PENDING_RECEIPT_STATE = "PRE_REVIEW_PENDING"
ACCEPTED_RECEIPT_STATE = "POST_REVIEW_ACCEPTED"
ACCEPTED_RECEIPT_VERDICT = "STATIC_REVIEW_PASS"
REVIEWED_AT_RE = re.compile(
    r"^[0-9]{4}-(?:0[1-9]|1[0-2])-(?:0[1-9]|[12][0-9]|3[01])T(?:[01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]Z$"
)
FORBIDDEN = (r"testframework/core/", r"testframework/requirements\.lock(?![\w.-])", r"testframework/pytest\.ini(?![\w.-])", r"testframework/docker/Dockerfile(?![\w.-])", r"testframework/docker/compose\.yaml(?![\w.-])", r"\.gitkeep(?![\w.-])")
PHASE129_SCHEMA = "hub.phase129-task-requirements/v1"
PHASE129_FIELDS = ("plan_id", "task_run_id_pattern", "task_id", "batch_id", "expected_operations", "owner_role", "verifier_roles", "manifest_path", "image_lock_path", "mode", "deadlines_seconds")
PHASE129_AUTOMATED_BODY = "NOT_READY — no supported invocation exists; execution remains NOT_RUN until a separately specified, implemented, independently accepted, and atomically resealed API under testframework/runner/** exists."
PHASE129_FIELD_COVERAGE = {"phase129_task_blocks":176,"valid_blocks":176,"batch_blocks":173,"non_batch_blocks":3,"missing_blocks":0,"duplicate_blocks":0,"unknown_fields":0,"invalid_order":0,"invalid_values":0,"command_spans":0,"positive_invocations":0}
TASK_RUN_RE = re.compile(r"^129-0[1-7]-[a-z0-9]+(?:-[a-z0-9]+)*-<caller-supplied-unique-nonce-attempt>$")
BATCH_ID_RE = re.compile(r"^129-0[2-7]-[a-z0-9]+(?:-[a-z0-9]+)*$")
ROLE_RE = re.compile(r"^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$")
CANONICAL_INTEGER_RE = re.compile(r"^(?:0|[1-9][0-9]*)$")
PHASE129_PLAN_IDS = tuple(f"129-0{index}" for index in range(1, 8))
PHASE129_TASK_IDS = {"TASK-193", "TASK-197", "TASK-198", "TASK-201", "TASK-204", "TASK-219", "TASK-234"}
PHASE129_MODES = {"bootstrap-acceptance", "preflight-readonly", "preflight-verify", "apply-and-verify"}
PHASE129_DEADLINES = {"720/900", "1500/1800", "3300/3600"}
NEGATIVE_PROSE_RE = re.compile(
    r"(?i)\b(?:forbid(?:den)?|reject(?:ed|s)?|no supported invocation|no command|not permitted)\b"
    r"|\u043d\u0435\s+\u0440\u0430\u0437\u0440\u0435\u0448\u0435\u043d"
    r"|\u0437\u0430\u043f\u0440\u0435\u0449"
)
NEGATIVE_DECLARATION_LITERALS = ("docker run", "docker compose")
NEGATIVE_DECLARATION_SEGMENT_RE = re.compile(
    r"(?i)(?:\bdirect\s+fallbacks|\u043f\u0440\u044f\u043c\u044b\u0435)\s+`docker run`\s*(?:,|and)\s*`docker compose`"
    r"[^.;]{0,128}?(?:not permitted|\u043d\u0435\s+\u0440\u0430\u0437\u0440\u0435\u0448\u0435\u043d\u044b|\u0437\u0430\u043f\u0440\u0435\u0449\u0435\u043d\u044b)\b"
)
EXPECTED_ACTIVE_NEGATIVE_DECLARATIONS = {literal: 1 for literal in NEGATIVE_DECLARATION_LITERALS}
CONCRETE_LAUNCHER_VERSION_PATTERNS = (
    r"(?i)(?<![A-Za-z0-9_.-])hub\s*\.\s*qa\s*-\s*runner\s*(?:(?:[/.:=()]|-)\s*)*(?:v\s*|version\s*(?:[/.:=]|-)?\s*)[0-9]+(?:\s*\.\s*[0-9]+)*(?:[-+]?[A-Za-z][A-Za-z0-9.-]*)?(?![A-Za-z0-9])",
    r"(?i)(?<![A-Za-z0-9_.-])testframework\s*/\s*runner\s*/\s*\*{2}(?:\s*(?:[/.:=]|-)?\s*api\b)?\s*(?:(?:[/.:=()]|-)\s*)*(?:v\s*|version\s*(?:[/.:=]|-)?\s*)[0-9]+(?:\s*\.\s*[0-9]+)*(?:[-+]?[A-Za-z][A-Za-z0-9.-]*)?(?![A-Za-z0-9])",
)
POSITIVE_INVOCATION_PATTERNS = tuple(re.compile(pattern) for pattern in (
    *CONCRETE_LAUNCHER_VERSION_PATTERNS,
    r"(?i)\b(?:launcher|runner(?:\s+api)?)\s+v[0-9]+(?:\.[0-9]+)*\b",
    r"(?i)(?<![A-Za-z0-9*])--?[A-Za-z][A-Za-z0-9-]*(?=\s|[=,.;:)/]|$)",
    r"(?i)\b(?:docker\s+(?:run|compose)|docker-compose|python(?:3)?\s+-m|(?:pwsh|powershell)(?:\.exe)?\s+-[A-Za-z]|bash\s+-[a-z]|cmd(?:\.exe)?\s+/[a-z])\b",
    r"(?i)testframework/runner/[^\s<`]*(?:\.py|\.ps1|\.sh|\.cmd|\.bat|\.exe)(?=\s+(?:--?[A-Za-z]|[A-Za-z]))",
    r"(?i)\b[a-z0-9_]*(?:validate|apply|run)[a-z0-9_-]*\.(?:py|ps1|sh|cmd|bat|exe)\s+--?[A-Za-z]",
    r"(?i)\b(?:invoke-[a-z0-9-]*runner|runner-(?:alias|shim))\b",
    r"(?:&&|\|\|)",
))
LEGACY_FORBIDDEN_LAUNCHER_FORMS = (
    "testframework/docker/run-task.ps1",
    "run-task.ps1",
    "hub.qa-runner/v1",
    r"C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe -NoProfile -NonInteractive -File testframework/docker/run-task.ps1",
)
NEGATIVE_INVOCATION_FIXTURES = (
    ("legacy-authority", LEGACY_FORBIDDEN_LAUNCHER_FORMS[0]),
    ("future-runner-entrypoint", "testframework/runner/entrypoint.py --mode apply-and-verify"),
    ("direct-docker", "docker run immutable-image"),
    ("direct-compose", "docker compose up"),
    ("host-interpreter", "python3 -m pytest"),
    ("flag-fragment", "-Mode apply-and-verify"),
    ("alias", "Invoke-HubRunner apply"),
    ("shim", "runner-shim apply"),
    ("fallback", "runner-alias || python3 -m pytest"),
    ("code-span", "`testframework/runner/entrypoint.py --mode apply-and-verify`"),
    ("shell-operator", "candidate && fallback"),
    ("replacement-slash-schema-v2", "hub.qa-runner/v2"),
    ("replacement-slash-schema-v999", "hub.qa-runner/v999"),
    ("replacement-slash-schema-v999-variant", "`HUB . QA - RUNNER / V 999.4`,"),
    ("future-runner-api-v2", "testframework/runner/** API v2"),
    ("future-runner-api-v999-variant", "`TESTFRAMEWORK / RUNNER / ** / API: VERSION=999.4`,"),
    ("replacement-whitespace-version", "hub.qa-runner VERSION 2"),
    ("replacement-spaced-whitespace-version", "hub . qa - runner version: 2"),
    ("replacement-attached-prerelease", "hub.qa-runner/v2beta"),
    ("future-runner-api-attached-prerelease", "testframework/runner/** API v2beta"),
    ("replacement-hyphen-prerelease", "hub.qa-runner/v2-rc.1"),
    ("replacement-build-metadata", "hub.qa-runner/v2+build.7"),
    ("future-runner-api-attached-rc", "testframework/runner/** API V 2RC1"),
    ("future-runner-api-prerelease-variant", "TESTFRAMEWORK / RUNNER / ** / API: VERSION=999.4-beta.1"),
)
VALID_NEGATIVE_DECLARATION_FIXTURE = (
    "Direct fallbacks `docker run` and `docker compose` are not permitted."
)
MIXED_NEGATIVE_INVOCATION_FIXTURES = (
    (
        "legacy-plus-future",
        f"Forbidden: `{LEGACY_FORBIDDEN_LAUNCHER_FORMS[0]}`; use `{NEGATIVE_INVOCATION_FIXTURES[1][1]}`",
    ),
    ("negative-plus-version", f"{VALID_NEGATIVE_DECLARATION_FIXTURE} Use launcher v9."),
    ("negative-plus-slash-schema-v2", f"{VALID_NEGATIVE_DECLARATION_FIXTURE} Use hub.qa-runner/v2."),
    ("negative-plus-slash-schema-v999-variant", f"{VALID_NEGATIVE_DECLARATION_FIXTURE} Use HUB . QA - RUNNER / V 999."),
    ("negative-plus-runner-api-v2", f"{VALID_NEGATIVE_DECLARATION_FIXTURE} Use testframework/runner/** API v2."),
    ("negative-plus-runner-api-v999-variant", f"{VALID_NEGATIVE_DECLARATION_FIXTURE} Use TESTFRAMEWORK / RUNNER / ** / API: VERSION=999.4."),
    ("negative-plus-whitespace-version", f"{VALID_NEGATIVE_DECLARATION_FIXTURE} Use hub.qa-runner VERSION 2."),
    ("negative-plus-spaced-whitespace-version", f"{VALID_NEGATIVE_DECLARATION_FIXTURE} Use hub . qa - runner version: 2."),
    ("negative-plus-attached-prerelease", f"{VALID_NEGATIVE_DECLARATION_FIXTURE} Use hub.qa-runner/v2beta."),
    ("negative-plus-runner-api-attached-prerelease", f"{VALID_NEGATIVE_DECLARATION_FIXTURE} Use testframework/runner/** API v2beta."),
    ("negative-plus-option", f"{VALID_NEGATIVE_DECLARATION_FIXTURE} Use --mode apply-and-verify."),
    ("duplicate-negative-span", f"{VALID_NEGATIVE_DECLARATION_FIXTURE} Duplicate `docker run` is forbidden."),
    ("alternate-unbounded-form", "Direct fallbacks `docker run --privileged image` and `docker compose` are not permitted."),
)
NEUTRAL_NON_INVOCATION_FIXTURES = (
    ("future-runner-api-version-unspecified", "The future testframework/runner/** API remains version-unspecified and NOT_READY."),
    ("neutral-launcher-schema-policy", "No concrete launcher schema version is accepted; runtime remains NOT_RUN."),
    ("hub-version-unspecified", "hub.qa-runner version unspecified"),
    ("future-runner-api-vnext", "testframework/runner/** API vNext"),
    ("unrelated-product-version", "product v2"),
    ("github-prefix-near-miss", "github.qa-runner/v2"),
    ("xhub-prefix-near-miss", "xhub.qa-runner/v2"),
)
NEGATIVE_REQUIREMENT_FIXTURES = (
    "duplicate-field", "missing-field", "unknown-field", "reordered-field",
    "extra-attribute", "nested-node", "mixed-text", "bad-adjacency",
    "invalid-schema", "invalid-enum", "invalid-cross-field", "invalid-deadline",
    "invalid-owner-role", "invalid-verifier-role", "name-count-drift",
    "precondition-count-drift", "missing-precondition-identity",
    "missing-files-identity", "manifest-authority-drift", "frontmatter-drift",
)

def fail(message: str) -> None: raise ValueError(message)
def sha(data: bytes) -> str: return hashlib.sha256(data).hexdigest()


def json_load_exact(data: bytes, context: str) -> Any:
    def unique_object(pairs: list[tuple[str, Any]]) -> dict[str, Any]:
        result: dict[str, Any] = {}
        for key, value in pairs:
            if key in result:
                fail(f"{context}: duplicate JSON key: {key}")
            result[key] = value
        return result

    def reject_constant(value: str) -> None:
        fail(f"{context}: non-finite JSON value: {value}")

    try:
        return json.loads(
            data.decode("utf-8"),
            object_pairs_hook=unique_object,
            parse_constant=reject_constant,
        )
    except UnicodeError as error:
        fail(f"{context}: invalid UTF-8: {error.__class__.__name__}")

def open_read(root_fd: int, relative: str, cap: int = MAX_FILE_BYTES) -> bytes:
    parts = PurePosixPath(relative).parts
    if not parts or any(part in ("", ".", "..") for part in parts): fail(f"unsafe path: {relative}")
    fd = os.dup(root_fd)
    try:
        for part in parts[:-1]:
            next_fd = os.open(part, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW, dir_fd=fd)
            os.close(fd); fd = next_fd
        leaf = os.open(parts[-1], os.O_RDONLY | os.O_NOFOLLOW, dir_fd=fd)
        try:
            info = os.fstat(leaf)
            if not stat.S_ISREG(info.st_mode) or info.st_nlink != 1 or info.st_size > cap: fail(f"unsafe file: {relative}")
            chunks: list[bytes] = []; remaining = cap + 1
            while remaining:
                chunk = os.read(leaf, min(65536, remaining))
                if not chunk: break
                chunks.append(chunk); remaining -= len(chunk)
            data = b"".join(chunks)
            if len(data) > cap: fail(f"oversized file: {relative}")
            return data
        finally: os.close(leaf)
    finally: os.close(fd)


def literal_assignment(source: bytes, path: str, name: str) -> Any:
    try:
        tree = ast.parse(source.decode("utf-8"), filename=path)
    except (SyntaxError, UnicodeError) as error:
        fail(f"{path}: source parse failed: {error.__class__.__name__}")
    values: list[ast.expr] = []
    for node in tree.body:
        if isinstance(node, ast.Assign) and len(node.targets) == 1:
            target = node.targets[0]
            if isinstance(target, ast.Name) and target.id == name:
                values.append(node.value)
        elif isinstance(node, ast.AnnAssign) and isinstance(node.target, ast.Name) and node.target.id == name:
            values.append(node.value)
    if len(values) != 1:
        fail(f"{path}: {name} assignment cardinality drift")
    try:
        return ast.literal_eval(values[0])
    except (ValueError, TypeError) as error:
        fail(f"{path}: {name} is not a literal: {error.__class__.__name__}")


def phase129_tool_contract(tool_sources: dict[str, bytes]) -> tuple[str, dict[str, tuple[str, ...]], dict[str, tuple[str, ...]]]:
    contracts: list[tuple[str, dict[str, tuple[str, ...]], dict[str, tuple[str, ...]]]] = []
    for path in (BUILDER_PATH, VALIDATOR_PATH):
        source = tool_sources.get(path)
        if source is None or sha(source) != PHASE129_TOOL_SHA256[path]:
            fail(f"{path}: sealed source digest drift")
        order_sha256 = literal_assignment(source, path, "PHASE129_BATCH_ORDER_SHA256")
        plan_tasks = literal_assignment(source, path, "PHASE129_PLAN_TASKS")
        task_overrides = literal_assignment(source, path, "PHASE129_BATCH_TASK_OVERRIDES")
        if not isinstance(order_sha256, str) or re.fullmatch(r"[0-9a-f]{64}", order_sha256) is None:
            fail(f"{path}: ordered batch digest literal invalid")
        if not isinstance(plan_tasks, dict) or set(plan_tasks) != set(PHASE129_PLAN_IDS[1:]):
            fail(f"{path}: plan TASK projection coverage drift")
        if not isinstance(task_overrides, dict):
            fail(f"{path}: batch TASK override projection invalid")
        for context, mapping in (("plan", plan_tasks), ("override", task_overrides)):
            for key, value in mapping.items():
                if (
                    not isinstance(key, str)
                    or (context == "override" and BATCH_ID_RE.fullmatch(key) is None)
                    or not isinstance(value, tuple)
                    or not value
                    or len(value) != len(set(value))
                    or any(not isinstance(item, str) or item not in PHASE129_TASK_IDS for item in value)
                ):
                    fail(f"{path}: {context} TASK projection entry invalid")
        contracts.append((order_sha256, plan_tasks, task_overrides))
    if contracts[0] != contracts[1]:
        fail("builder/validator declarative batch contract differs")
    return contracts[0]


def phase_authority(root_fd: int) -> set[str]:
    phases_fd = os.open(".planning/phases", os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW, dir_fd=root_fd)
    try:
        names = os.listdir(phases_fd)
        if len(names) > MAX_PHASE_ENTRIES * 16: fail("phase authority enumeration budget exceeded")
        found: set[str] = set()
        for name in names:
            if not re.match(r"^(?:12[89]|1[3-4][0-9])-", name): continue
            child = os.open(name, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW, dir_fd=phases_fd)
            try:
                if "PLAN.md" in os.listdir(child): found.add(f".planning/phases/{name}/PLAN.md")
            finally: os.close(child)
        return found
    finally: os.close(phases_fd)

def blocks(text: str, path: str) -> list[str]:
    result = re.findall(r"<task\b[^>]*>.*?</task>", text, flags=re.S)
    if not result: fail(f"{path}: no task blocks")
    seen: set[str] = set()
    for block in result:
        name = re.search(r"(?:id=\"([^\"]+)\"|<name>\s*([^:<]+))", block)
        if not name or not (name.group(1) or name.group(2)): fail(f"{path}: malformed task identity")
        identity = name.group(1) or name.group(2)
        if identity in seen: fail(f"{path}: duplicate task identity {identity}")
        seen.add(identity)
    return result

def has(block: str, tag: str) -> bool: return bool(re.search(rf"<{tag}(?:\s|>)\s*\S", block, flags=re.S))
def task_contract(block: str, path: str, strict_roles: bool) -> None:
    required = ("read_first", "files", "done") + (("owner", "verifier") if strict_roles else ())
    for tag in required:
        if not has(block, tag): fail(f"{path}: missing/empty {tag}")
    if not (has(block, "requirement") or has(block, "requirements")): fail(f"{path}: missing requirements")
    if strict_roles and not has(block, "automated"): fail(f"{path}: missing automated verification")
    if not (has(block, "consumes") or has(block, "precondition")): fail(f"{path}: missing consumes/precondition")
    if not (has(block, "produces") or has(block, "action")): fail(f"{path}: missing produces/action")
    if not (has(block, "failing_sentinel") or re.search(r"\b(?:fail|forbidden|reject)\b", block, re.I)): fail(f"{path}: missing failing sentinel")
    if not (has(block, "atomic_allowlist") or re.search(r"atomic.{0,80}(?:allowlist|commit)", block, re.I | re.S)): fail(f"{path}: missing atomic allowlist")
    verify = re.findall(r"<verify(?:\s[^>]*)?>(.*?)</verify>", block, re.S)
    if len(verify) != 1 or not verify[0].strip(): fail(f"{path}: verify must be exactly one nonempty block")
    files = re.search(r"<files>(.*?)</files>", block, re.S)
    declared = [value.strip() for value in re.split(r"[;,]", files.group(1)) if value.strip()]
    if not 1 <= len(declared) <= 14: fail(f"{path}: Files cardinality outside 1..14")

def literal_file_tokens(files: str, path: str) -> set[str]:
    tokens: set[str] = set()
    for item in re.split(r"[;,]", files):
        value = item.strip().replace("&lt;", "<").replace("&gt;", ">")
        if not value.startswith("testframework/"):
            continue
        value = value.removesuffix(" only").rstrip("/")
        if not value or ".." in PurePosixPath(value).parts:
            fail(f"{path}: traversal Files token")
        if "*" in value:
            if value.startswith("testframework/docs/"):
                fail(f"{path}: broad docs Files token")
            continue
        if value in tokens:
            fail(f"{path}: duplicate literal Files token")
        tokens.add(value)
    return tokens

def command_evidence(value: str) -> tuple[int, int]:
    positive = sum(len(pattern.findall(value)) for pattern in POSITIVE_INVOCATION_PATTERNS)
    positive += sum(value.count(form) for form in LEGACY_FORBIDDEN_LAUNCHER_FORMS)
    command_spans = 0
    for span in re.findall(r"`([^`\r\n]+)`", value):
        if any(pattern.search(span) for pattern in POSITIVE_INVOCATION_PATTERNS) or any(
            form in span for form in LEGACY_FORBIDDEN_LAUNCHER_FORMS
        ):
            command_spans += 1
    return command_spans, positive


def reject_command_like(value: str, context: str) -> None:
    command_spans, positive = command_evidence(value)
    if command_spans or positive:
        fail(f"{context}: command-like invocation rejected")


def active_line_negative_declarations(line: str, path: str, number: int) -> tuple[str, ...]:
    if any(form in line for form in LEGACY_FORBIDDEN_LAUNCHER_FORMS):
        fail(f"{path}:{number}: active legacy launcher literal")
    sanitized = list(line)
    declarations: list[str] = []
    negative_segments = tuple(
        match.span()
        for match in NEGATIVE_DECLARATION_SEGMENT_RE.finditer(line)
        if NEGATIVE_PROSE_RE.search(match.group(0))
    )
    for match in re.finditer(r"`([^`\r\n]+)`", line):
        literal = match.group(1)
        spans, positive = command_evidence(literal)
        if not spans and not positive:
            continue
        if (
            path == MASTER_PATH
            and literal in NEGATIVE_DECLARATION_LITERALS
            and any(start <= match.start() and match.end() <= end for start, end in negative_segments)
        ):
            if literal in declarations:
                fail(f"{path}:{number}: duplicate negative invocation declaration")
            declarations.append(literal)
            sanitized[match.start():match.end()] = " " * (match.end() - match.start())
    spans, positive = command_evidence("".join(sanitized))
    if spans or positive:
        fail(f"{path}:{number}: active command/version/invocation syntax")
    return tuple(declarations)


def validate_active_authority_prose(texts: dict[str, str]) -> tuple[int, int]:
    if tuple(path for path in PHASE129_ACTIVE_PATHS if path not in texts):
        fail("Phase 129 active authority input is incomplete")
    command_spans = positive_invocations = 0
    negative_declarations = {literal: 0 for literal in NEGATIVE_DECLARATION_LITERALS}
    for path in PHASE129_ACTIVE_PATHS:
        for number, line in enumerate(texts[path].splitlines(), 1):
            for literal in active_line_negative_declarations(line, path, number):
                negative_declarations[literal] += 1
    if negative_declarations != EXPECTED_ACTIVE_NEGATIVE_DECLARATIONS:
        fail("Phase 129 exact negative invocation declarations drift")
    ownership_text = "\n".join((texts[MASTER_PATH], texts[PLAN_PATHS[1]]))
    runner_claims = re.findall(
        r"(?is)(?:belongs|authority|owned|owns).{0,180}?(?:only|solely).{0,180}?testframework/runner/\*\*"
        r"|testframework/runner/\*\*.{0,180}?(?:only|solely).{0,180}?(?:belongs|authority|owned|owns)",
        ownership_text,
    )
    docker_claims = re.findall(
        r"(?is)testframework/docker/\*\*.{0,180}?immutable.{0,180}?only"
        r"|immutable.{0,180}?testframework/docker/\*\*.{0,180}?only",
        ownership_text,
    )
    if not runner_claims or not docker_claims:
        fail("Phase 129 runner/Docker ownership authority is not declaratively closed")
    for path in PHASE129_NUMBERED_PATHS:
        text = texts[path]
        if "testframework/runner/**" not in text or "NOT_READY" not in text or "NOT_RUN" not in text:
            fail(f"{path}: future runner/readiness declaration missing")
    return command_spans, positive_invocations


def parse_requirements_xml(source: str, context: str) -> dict[str, str]:
    if not source or len(source.encode("utf-8")) > MAX_PHASE129_REQUIREMENTS_BYTES:
        fail(f"{context}: requirements block size drift")
    try:
        root = ET.fromstring(source)
    except ET.ParseError as error:
        fail(f"{context}: malformed requirements XML: {error}")
    if root.tag != "non_executable_execution_requirements" or root.attrib != {"schema": PHASE129_SCHEMA}:
        fail(f"{context}: outer schema/attribute drift")
    if root.text and root.text.strip():
        fail(f"{context}: mixed root text")
    children = list(root)
    if tuple(child.tag for child in children) != PHASE129_FIELDS:
        fail(f"{context}: missing/duplicate/unknown/reordered field")
    values: dict[str, str] = {}
    dotted_paths = {CUTOVER_MANIFEST_PATH, "testframework/docker/images.lock.json"}
    for child in children:
        if child.attrib or list(child):
            fail(f"{context}: attributes/nested nodes forbidden")
        if child.tail and child.tail.strip():
            fail(f"{context}: mixed child text")
        value = child.text or ""
        if value != value.strip() or "\r" in value or "\n" in value or not value:
            fail(f"{context}: field must be one non-empty scalar line")
        if re.fullmatch(r"[A-Za-z0-9_./,*<>-]+", value) is None:
            fail(f"{context}: non-declarative scalar")
        if "." in value and value not in dotted_paths:
            fail(f"{context}: dotted scalar is not an exact JSON path")
        if re.search(r"(?i)\.(?:py|ps1|sh|cmd|bat|exe)$", value):
            fail(f"{context}: executable suffix forbidden")
        reject_command_like(value, context)
        values[child.tag] = value
    validate_phase129_field_types(values, context)
    return values


def validate_phase129_field_types(values: dict[str, str], context: str) -> None:
    if values["plan_id"] not in PHASE129_PLAN_IDS:
        fail(f"{context}: plan_id enum invalid")
    if TASK_RUN_RE.fullmatch(values["task_run_id_pattern"]) is None:
        fail(f"{context}: task_run_id_pattern invalid")
    if values["task_id"] not in PHASE129_TASK_IDS:
        fail(f"{context}: task_id enum invalid")
    if CANONICAL_INTEGER_RE.fullmatch(values["expected_operations"]) is None:
        fail(f"{context}: expected_operations is not a canonical integer")
    if ROLE_RE.fullmatch(values["owner_role"]) is None:
        fail(f"{context}: owner_role syntax invalid")
    verifier_roles = values["verifier_roles"].split(",")
    if (
        not verifier_roles
        or len(verifier_roles) != len(set(verifier_roles))
        or any(ROLE_RE.fullmatch(role) is None for role in verifier_roles)
    ):
        fail(f"{context}: verifier_roles syntax invalid")
    if values["manifest_path"] not in {"NONE", CUTOVER_MANIFEST_PATH}:
        fail(f"{context}: manifest_path enum invalid")
    if values["image_lock_path"] != "testframework/docker/images.lock.json":
        fail(f"{context}: image_lock_path invalid")
    if values["mode"] not in PHASE129_MODES or values["deadlines_seconds"] not in PHASE129_DEADLINES:
        fail(f"{context}: mode/deadline enum invalid")
    plan_id = values["plan_id"]
    if plan_id == "129-01":
        if values["batch_id"] != "NONE" or values["expected_operations"] != "0" or values["mode"] == "apply-and-verify":
            fail(f"{context}: non-batch type closure invalid")
    elif (
        BATCH_ID_RE.fullmatch(values["batch_id"]) is None
        or not values["batch_id"].startswith(f"{plan_id}-")
        or values["expected_operations"] == "0"
        or values["manifest_path"] != CUTOVER_MANIFEST_PATH
        or values["mode"] != "apply-and-verify"
    ):
        fail(f"{context}: batch type closure invalid")


def frontmatter_plan_id(text: str, path: str) -> str:
    frontmatter = re.match(r"\A---\r?\n(.*?)\r?\n---(?:\r?\n|$)", text, re.S)
    if frontmatter is None:
        fail(f"{path}: missing bounded frontmatter")
    phase = re.findall(r"^phase:\s*([^\s]+)\s*$", frontmatter.group(1), re.M)
    plan = re.findall(r'^plan:\s*"?([^"\s]+)"?\s*$', frontmatter.group(1), re.M)
    if phase != ["129-tf-e2e-shared"] or len(plan) != 1:
        fail(f"{path}: phase/plan frontmatter cardinality")
    normalized = f"129-{plan[0]}" if re.fullmatch(r"0[1-7]", plan[0]) else plan[0]
    filename = re.search(r"/(129-0[1-7])-PLAN\.md$", path)
    if normalized not in PHASE129_PLAN_IDS or filename is None or normalized != filename.group(1):
        fail(f"{path}: frontmatter/file plan drift")
    return normalized


def one_tag(task: str, name: str, context: str) -> str:
    matches = re.findall(rf"<{name}>(.*?)</{name}>", task, re.S)
    if len(matches) != 1 or not matches[0].strip():
        fail(f"{context}: {name} cardinality")
    return matches[0].strip()


def verifier_csv(value: str, context: str) -> str:
    roles = [item.strip() for item in re.split(r"\s+\+\s+", value)]
    if (
        not roles
        or any(ROLE_RE.fullmatch(item) is None for item in roles)
        or len(roles) != len(set(roles))
    ):
        fail(f"{context}: verifier role list invalid")
    return ",".join(roles)


def requirement_ids(value: str, context: str) -> tuple[str, ...]:
    identifiers = tuple(re.findall(r"TASK-[0-9]+", value))
    if not identifiers or len(identifiers) != len(set(identifiers)) or any(
        identifier not in PHASE129_TASK_IDS for identifier in identifiers
    ):
        fail(f"{context}: task requirement identity invalid")
    return identifiers


def contains_batch_identity(value: str, batch_id: str) -> bool:
    return re.search(rf"(?<![a-z0-9-]){re.escape(batch_id)}(?![a-z0-9-])", value) is not None


def contains_operation_count(value: str, expected: int) -> bool:
    numbers = {
        int(item)
        for item in re.findall(r"(?<![A-Za-z0-9-])[0-9]+(?![A-Za-z0-9-])", value)
    }
    return expected in numbers


def manifest_batch_authority(payload: Any) -> dict[str, dict[str, Any]]:
    if not isinstance(payload, dict) or not isinstance(payload.get("batches"), list):
        fail("cutover manifest batch authority is malformed")
    batches = payload["batches"]
    if len(batches) != 173:
        fail("cutover manifest batch count drift")
    result: dict[str, dict[str, Any]] = {}
    operation_total = maximum = 0
    for index, batch in enumerate(batches):
        context = f"cutover-manifest/batches/{index}"
        if not isinstance(batch, dict):
            fail(f"{context}: batch must be an object")
        batch_id = batch.get("id")
        plan_id = batch.get("plan")
        operations = batch.get("operations")
        tasks = batch.get("tasks")
        owner = batch.get("owner")
        verifiers = batch.get("verifiers")
        if (
            not isinstance(batch_id, str)
            or BATCH_ID_RE.fullmatch(batch_id) is None
            or batch_id in result
            or plan_id not in PHASE129_PLAN_IDS[1:]
            or not batch_id.startswith(f"{plan_id}-")
            or batch.get("phase") != 129
        ):
            fail(f"{context}: identity/plan closure invalid")
        if not isinstance(operations, list) or not 1 <= len(operations) <= 56:
            fail(f"{context}: operation list invalid")
        if (
            not isinstance(tasks, list)
            or not tasks
            or any(not isinstance(item, str) or item not in PHASE129_TASK_IDS for item in tasks)
            or len(tasks) != len(set(tasks))
        ):
            fail(f"{context}: task authority invalid")
        if not isinstance(owner, str) or ROLE_RE.fullmatch(owner) is None:
            fail(f"{context}: owner authority invalid")
        if (
            not isinstance(verifiers, list)
            or not verifiers
            or any(not isinstance(item, str) or ROLE_RE.fullmatch(item) is None for item in verifiers)
            or len(verifiers) != len(set(verifiers))
        ):
            fail(f"{context}: verifier authority invalid")
        result[batch_id] = {
            "plan": plan_id,
            "operations": len(operations),
            "tasks": tuple(tasks),
            "owner": owner,
            "verifiers": tuple(verifiers),
        }
        operation_total += len(operations)
        maximum = max(maximum, len(operations))
    if operation_total != 1120 or maximum != 56:
        fail("cutover manifest arithmetic authority drift")
    return result


def validate_phase129_master_ownership(phase_master: str, authorities: list[dict[str, Any]]) -> None:
    rows: list[tuple[str, tuple[str, ...], str, str]] = []
    pattern = re.compile(
        r"^\|\s*(129-0[1-7](?:-[ABC])?)(?:\s+[^|]*)?\|\s*([^|]+)\|\s*([^|]+)\|\s*([^|]+)\|\s*$"
    )
    for line in phase_master.splitlines():
        match = pattern.fullmatch(line)
        if match is None:
            continue
        context = f"Phase-129 master ownership/{match.group(1)}"
        owner = match.group(3).strip()
        if ROLE_RE.fullmatch(owner) is None:
            fail(f"{context}: owner role invalid")
        rows.append(
            (
                match.group(1),
                requirement_ids(match.group(2), context),
                owner,
                verifier_csv(match.group(4).strip(), context),
            )
        )
    expected_keys = {"129-01-A", "129-01-B", "129-01-C", *PHASE129_PLAN_IDS[1:]}
    if len(rows) != 10 or {row[0] for row in rows} != expected_keys:
        fail("Phase 129 master ownership row coverage drift")
    by_identity = {item["identity"]: item for item in authorities if not item["is_batch"]}
    for key in ("129-01-A", "129-01-B", "129-01-C"):
        matching = [row for row in rows if row[0] == key]
        authority = by_identity.get(key)
        if len(matching) != 1 or authority is None:
            fail(f"Phase 129 master ownership/{key}: row/task missing")
        row = matching[0]
        if (
            row[1] != authority["requirements"]
            or row[2] != authority["owner"]
            or row[3] != authority["verifiers"]
        ):
            fail(f"Phase 129 master ownership/{key}: task conflict")
    for plan_id in PHASE129_PLAN_IDS[1:]:
        plan_rows = [row for row in rows if row[0] == plan_id]
        tasks = [item for item in authorities if item["plan_id"] == plan_id]
        row_pairs = {(row[2], row[3]) for row in plan_rows}
        task_pairs = {(item["owner"], item["verifiers"]) for item in tasks}
        row_requirements = {identifier for row in plan_rows for identifier in row[1]}
        task_requirements = {identifier for item in tasks for identifier in item["requirements"]}
        if not plan_rows or not tasks or row_pairs != task_pairs or row_requirements != task_requirements:
            fail(f"Phase 129 master ownership/{plan_id}: plan/task conflict")


def validate_phase129_task(
    task: str,
    path: str,
    expected_plan_id: str,
    external_batches: dict[str, dict[str, Any]] | None,
) -> dict[str, Any]:
    identity = re.search(r'(?:id="([^"]+)"|<name>\s*([^:<]+))', task)
    context = f"{path}/{(identity.group(1) or identity.group(2)) if identity else 'unknown'}"
    if identity is None:
        fail(f"{context}: task identity missing")
    task_identity = identity.group(1) or identity.group(2)
    matches = list(re.finditer(r'<non_executable_execution_requirements\b[^>]*>[\s\S]*?</non_executable_execution_requirements>', task))
    if len(re.findall(r"<non_executable_execution_requirements\b", task)) != 1 or len(matches) != 1:
        fail(f"{context}: requirements block missing/duplicate")
    match = matches[0]
    adjacent = re.match(r"\s*<verify(?:\s[^>]*)?>(.*?)</verify>", task[match.end():], re.S)
    if adjacent is None:
        fail(f"{context}: requirements block is not adjacent to verify")
    automated = re.fullmatch(r"\s*<automated>(.*?)</automated>\s*", adjacent.group(1), re.S)
    if automated is None or automated.group(1) != PHASE129_AUTOMATED_BODY or "`" in automated.group(1):
        fail(f"{context}: automated readiness body drift")
    values = parse_requirements_xml(match.group(0), context)
    if values["plan_id"] != expected_plan_id:
        fail(f"{context}: plan/frontmatter/file cross-field drift")
    plan_id = values["plan_id"]
    owner = one_tag(task, "owner", context)
    if ROLE_RE.fullmatch(owner) is None:
        fail(f"{context}: owner tag syntax invalid")
    verifiers = verifier_csv(one_tag(task, "verifier", context), context)
    if values["owner_role"] != owner or values["verifier_roles"] != verifiers:
        fail(f"{context}: owner/verifier cross-field drift")
    deadline = re.fullmatch(r"(0|[1-9][0-9]*)/(0|[1-9][0-9]*)", values["deadlines_seconds"])
    if deadline is None or int(deadline.group(1)) >= int(deadline.group(2)):
        fail(f"{context}: deadline order invalid")
    requirements = re.findall(r"<requirements?>(.*?)</requirements?>", task, re.S)
    if len(requirements) != 1:
        fail(f"{context}: task requirements cardinality")
    declared_requirements = requirement_ids(requirements[0], context)
    non_batch = {
        "129-01-A": ("129-01-bootstrap-acceptance-<caller-supplied-unique-nonce-attempt>", "TASK-201", "NONE", "bootstrap-acceptance", "1500/1800", ("TASK-193", "TASK-201")),
        "129-01-B": ("129-01-preflight-readonly-<caller-supplied-unique-nonce-attempt>", "TASK-204", CUTOVER_MANIFEST_PATH, "preflight-readonly", "720/900", ("TASK-204",)),
        "129-01-C": ("129-01-preflight-verify-<caller-supplied-unique-nonce-attempt>", "TASK-204", CUTOVER_MANIFEST_PATH, "preflight-verify", "720/900", ("TASK-204",)),
    }
    task_attr = re.search(r'<task\b[^>]*\bid="([^"]+)"', task)
    if task_attr:
        expected = non_batch.get(task_attr.group(1))
        exact = {"task_run_id_pattern": expected[0], "task_id": expected[1], "batch_id": "NONE", "expected_operations": "0", "manifest_path": expected[2], "mode": expected[3], "deadlines_seconds": expected[4]} if expected else {}
        if (
            plan_id != "129-01"
            or not expected
            or any(values[k] != v for k, v in exact.items())
            or declared_requirements != expected[5]
        ):
            fail(f"{context}: non-batch cross-field drift")
        return {
            "is_batch": False,
            "batch_id": None,
            "operations": 0,
            "identity": task_attr.group(1),
            "plan_id": plan_id,
            "owner": owner,
            "verifiers": verifiers,
            "requirements": declared_requirements,
        }
    name = one_tag(task, "name", context)
    batch = re.match(r"Batch ([a-z0-9]+(?:-[a-z0-9]+)*):", name)
    if batch is None or BATCH_ID_RE.fullmatch(batch.group(1)) is None or values["batch_id"] != batch.group(1):
        fail(f"{context}: batch ID cross-field drift")
    batch_id = batch.group(1)
    precondition = one_tag(task, "precondition", context)
    files = one_tag(task, "files", context)
    tasks = {"129-02":"TASK-197","129-03":"TASK-198","129-04":"TASK-219","129-05":"TASK-219","129-06":"TASK-234","129-07":"TASK-219"}
    count_labels = {
        "129-02":"expected operation count",
        "129-03":"expected operation count",
        "129-04":"expected delete count",
        "129-05":"expected delete count",
        "129-06":"expected unit count",
        "129-07":"expected hook-file delete count",
    }
    deadline_expected = "3300/3600" if plan_id in {"129-06","129-07"} else "1500/1800"
    if re.fullmatch(r"[1-9][0-9]*", values["expected_operations"]) is None:
        fail(f"{context}: operation count invalid")
    operations = int(values["expected_operations"])
    exact = {"task_run_id_pattern":f"{batch_id}-transaction-<caller-supplied-unique-nonce-attempt>","task_id":tasks.get(plan_id),"manifest_path":CUTOVER_MANIFEST_PATH,"mode":"apply-and-verify","deadlines_seconds":deadline_expected}
    if (
        any(values[k] != v for k,v in exact.items())
        or declared_requirements[0] != exact["task_id"]
        or f"{count_labels[plan_id]} {operations}" not in files
    ):
        fail(f"{context}: batch cross-field drift")
    for field_name, field_value in (("name", name), ("precondition", precondition), ("files", files)):
        if not contains_batch_identity(field_value, batch_id):
            fail(f"{context}: batch identity missing from {field_name}")
        if not contains_operation_count(field_value, operations):
            fail(f"{context}: operation count missing from {field_name}")
    if external_batches is not None:
        external = external_batches.get(batch_id)
        if external is None or (
            external["plan"] != plan_id
            or external["operations"] != operations
            or external["tasks"] != declared_requirements
            or external["owner"] != owner
            or ",".join(external["verifiers"]) != verifiers
        ):
            fail(f"{context}: external batch authority cross-field drift")
    return {
        "is_batch": True,
        "batch_id": batch_id,
        "operations": operations,
        "identity": task_identity,
        "plan_id": plan_id,
        "owner": owner,
        "verifiers": verifiers,
        "requirements": declared_requirements,
    }


def validate_phase129_requirements(
    paths: tuple[str, ...],
    texts: tuple[str, ...],
    phase_master: str,
    external_batches: dict[str, dict[str, Any]] | None,
    command_spans: int,
    positive_invocations: int,
) -> tuple[dict[str, int], list[str], int, int, dict[str, dict[str, Any]]]:
    if paths != PHASE129_NUMBERED_PATHS:
        fail("Phase 129 numbered path order drift")
    total = batch_total = non_batch_total = operation_total = maximum = 0
    batch_ids: list[str] = []
    batch_projection: dict[str, dict[str, Any]] = {}
    authorities: list[dict[str, Any]] = []
    for path, text in zip(paths, texts, strict=True):
        plan_id = frontmatter_plan_id(text, path)
        for task in blocks(text, path):
            total += 1
            authority = validate_phase129_task(task, path, plan_id, external_batches)
            authorities.append(authority)
            if authority["is_batch"]:
                batch_total += 1
                batch_id = str(authority["batch_id"])
                batch_ids.append(batch_id)
                operation_total += int(authority["operations"])
                maximum = max(maximum, int(authority["operations"]))
                batch_projection[batch_id] = {
                    "plan": authority["plan_id"],
                    "operations": int(authority["operations"]),
                    "tasks": tuple(authority["requirements"]),
                    "owner": authority["owner"],
                    "verifiers": tuple(authority["verifiers"].split(",")),
                }
            else:
                non_batch_total += 1
    if (total,batch_total,non_batch_total,len(set(batch_ids)),operation_total,maximum) != (176,173,3,173,1120,56):
        fail("Phase 129 declarative arithmetic drift")
    if external_batches is not None and tuple(batch_ids) != tuple(external_batches):
        fail("Phase 129 task/external-authority batch order drift")
    validate_phase129_master_ownership(phase_master, authorities)
    coverage = {"phase129_task_blocks":total,"valid_blocks":total,"batch_blocks":batch_total,"non_batch_blocks":non_batch_total,"missing_blocks":0,"duplicate_blocks":0,"unknown_fields":0,"invalid_order":0,"invalid_values":0,"command_spans":command_spans,"positive_invocations":positive_invocations}
    if coverage != PHASE129_FIELD_COVERAGE:
        fail("Phase 129 prospective field coverage drift")
    return coverage, batch_ids, operation_total, maximum, batch_projection


def validate_phase129_tool_projection(
    tool_contract: tuple[str, dict[str, tuple[str, ...]], dict[str, tuple[str, ...]]],
    batch_ids: list[str],
    plan_batches: dict[str, dict[str, Any]],
) -> None:
    order_sha256, plan_tasks, task_overrides = tool_contract
    if order_sha256 != sha("\n".join(batch_ids).encode("utf-8")):
        fail("builder/validator ordered batch digest differs from numbered plans")
    derived_overrides: set[str] = set()
    for batch_id in batch_ids:
        authority = plan_batches[batch_id]
        plan_id = authority["plan"]
        base_tasks = plan_tasks.get(plan_id)
        if base_tasks is None:
            fail(f"builder/validator plan TASK authority missing: {plan_id}")
        if authority["tasks"] != base_tasks:
            derived_overrides.add(batch_id)
        if task_overrides.get(batch_id, base_tasks) != authority["tasks"]:
            fail(f"builder/validator TASK projection differs from numbered plan: {batch_id}")
    if set(task_overrides) != derived_overrides:
        fail("builder/validator TASK override set is not exact")


def validate_materialized_phase129_authority(
    plan_batches: dict[str, dict[str, Any]],
    manifest_batches: dict[str, dict[str, Any]],
) -> None:
    if tuple(plan_batches) != tuple(manifest_batches) or plan_batches != manifest_batches:
        fail("post-G129-P manifest projection differs from accepted numbered plans")


def negative_self_sentinels() -> None:
    for name, value in (*((f"legacy-{index}", item) for index, item in enumerate(LEGACY_FORBIDDEN_LAUNCHER_FORMS)), *NEGATIVE_INVOCATION_FIXTURES):
        try:
            reject_command_like(value, f"negative-invocation/{name}")
        except ValueError:
            continue
        fail(f"negative-invocation/{name}: accepted forbidden form")
    for name, value in NEUTRAL_NON_INVOCATION_FIXTURES:
        reject_command_like(value, f"neutral-non-invocation/{name}")
    valid_declarations = active_line_negative_declarations(
        VALID_NEGATIVE_DECLARATION_FIXTURE,
        MASTER_PATH,
        0,
    )
    if valid_declarations != NEGATIVE_DECLARATION_LITERALS:
        fail("valid negative invocation declaration sentinel drift")
    for name, value in MIXED_NEGATIVE_INVOCATION_FIXTURES:
        try:
            active_line_negative_declarations(value, MASTER_PATH, 0)
        except ValueError:
            continue
        fail(f"negative-invocation/{name}: accepted mixed or alternate form")
    try:
        json_load_exact(b'{"duplicate":1,"duplicate":2}', "negative-json/duplicate-key")
    except ValueError:
        pass
    else:
        fail("negative-json/duplicate-key: accepted duplicate key")
    xml = """<non_executable_execution_requirements schema="hub.phase129-task-requirements/v1">
  <plan_id>129-02</plan_id>
  <task_run_id_pattern>129-02-go-hook-free-fixture-transaction-&lt;caller-supplied-unique-nonce-attempt&gt;</task_run_id_pattern>
  <task_id>TASK-197</task_id>
  <batch_id>129-02-go-hook-free-fixture</batch_id>
  <expected_operations>1</expected_operations>
  <owner_role>be1</owner_role>
  <verifier_roles>be-ver</verifier_roles>
  <manifest_path>.planning/audits/test-rebase/cutover-manifest.json</manifest_path>
  <image_lock_path>testframework/docker/images.lock.json</image_lock_path>
  <mode>apply-and-verify</mode>
  <deadlines_seconds>1500/1800</deadlines_seconds>
</non_executable_execution_requirements>"""
    batch_id = "129-02-go-hook-free-fixture"
    fixture_manifest = {
        batch_id: {
            "plan": "129-02",
            "operations": 1,
            "tasks": ("TASK-197",),
            "owner": "be1",
            "verifiers": ("be-ver",),
        }
    }
    valid = f"""<task type="auto"><name>Batch {batch_id}: move 1 fixture</name><requirement>TASK-197</requirement><owner>be1</owner><verifier>be-ver</verifier><precondition>The sealed batch {batch_id} has exactly 1 operation.</precondition><files>Exactly batches[id={batch_id}].operations[]; expected operation count 1.</files>{xml}<verify><automated>{PHASE129_AUTOMATED_BODY}</automated></verify></task>"""
    valid_result = validate_phase129_task(
        valid,
        ".planning/phases/129-tf-e2e-shared/129-02-PLAN.md",
        "129-02",
        fixture_manifest,
    )
    if not valid_result["is_batch"] or valid_result["batch_id"] != batch_id or valid_result["operations"] != 1:
        fail("valid declarative self-sentinel drift")
    for name in NEGATIVE_REQUIREMENT_FIXTURES:
        if name == "frontmatter-drift":
            bad_document = f'---\nphase: 129-tf-e2e-shared\nplan: "03"\n---\n{valid}'
            try:
                frontmatter_plan_id(bad_document, ".planning/phases/129-tf-e2e-shared/129-02-PLAN.md")
            except ValueError:
                continue
            fail("negative-requirement/frontmatter-drift: accepted malformed fixture")
        bad = valid
        manifest_for_case = fixture_manifest
        if name == "duplicate-field": bad = bad.replace("</plan_id>", "</plan_id><plan_id>129-02</plan_id>", 1)
        elif name == "missing-field": bad = bad.replace("  <task_id>TASK-197</task_id>\n", "", 1)
        elif name == "unknown-field": bad = bad.replace("</plan_id>", "</plan_id><unknown>x</unknown>", 1)
        elif name == "reordered-field":
            a="  <plan_id>129-02</plan_id>\n"; b="  <task_run_id_pattern>129-02-go-hook-free-fixture-transaction-&lt;caller-supplied-unique-nonce-attempt&gt;</task_run_id_pattern>\n"; bad=bad.replace(a+b,b+a,1)
        elif name == "extra-attribute": bad=bad.replace('schema="hub.phase129-task-requirements/v1"','schema="hub.phase129-task-requirements/v1" extra="x"',1)
        elif name == "nested-node": bad=bad.replace("<plan_id>129-02</plan_id>","<plan_id><nested/></plan_id>",1)
        elif name == "mixed-text": bad=bad.replace("</plan_id>","</plan_id>MIXED",1)
        elif name == "bad-adjacency": bad=bad.replace("</non_executable_execution_requirements><verify>","</non_executable_execution_requirements><gap/><verify>",1)
        elif name == "invalid-schema": bad=bad.replace("hub.phase129-task-requirements/v1","invalid",1)
        elif name == "invalid-enum": bad=bad.replace("<mode>apply-and-verify</mode>","<mode>execute</mode>",1)
        elif name == "invalid-cross-field": bad=bad.replace("<batch_id>129-02-go-hook-free-fixture</batch_id>","<batch_id>129-02-go-hook-free-other</batch_id>",1)
        elif name == "invalid-deadline": bad=bad.replace("<deadlines_seconds>1500/1800</deadlines_seconds>","<deadlines_seconds>1500/1500</deadlines_seconds>",1)
        elif name == "invalid-owner-role":
            bad=bad.replace("<owner>be1</owner>","<owner>be_1</owner>",1).replace("<owner_role>be1</owner_role>","<owner_role>be_1</owner_role>",1)
        elif name == "invalid-verifier-role":
            bad=bad.replace("<verifier>be-ver</verifier>","<verifier>be_ver</verifier>",1).replace("<verifier_roles>be-ver</verifier_roles>","<verifier_roles>be_ver</verifier_roles>",1)
        elif name == "name-count-drift": bad=bad.replace(f"<name>Batch {batch_id}: move 1 fixture</name>",f"<name>Batch {batch_id}: move 2 fixtures</name>",1)
        elif name == "precondition-count-drift": bad=bad.replace(f"has exactly 1 operation.</precondition>",f"has exactly 2 operations.</precondition>",1)
        elif name == "missing-precondition-identity": bad=bad.replace(f"<precondition>The sealed batch {batch_id}","<precondition>The sealed batch 129-02-go-hook-free-other",1)
        elif name == "missing-files-identity": bad=bad.replace(f"<files>Exactly batches[id={batch_id}]", "<files>Exactly batches[id=129-02-go-hook-free-other]", 1)
        elif name == "manifest-authority-drift":
            manifest_for_case = {batch_id: {**fixture_manifest[batch_id], "operations": 2}}
        try:
            validate_phase129_task(
                bad,
                ".planning/phases/129-tf-e2e-shared/129-02-PLAN.md",
                "129-02",
                manifest_for_case,
            )
        except ValueError:
            continue
        fail(f"negative-requirement/{name}: accepted malformed fixture")


def closure_hash(texts: dict[str, bytes]) -> str:
    material = b"".join(path.encode("utf-8") + b"\0" + texts[path] for path in PLAN_PATHS)
    return sha(material)


def validate_receipt_state(
    receipt: dict[str, Any],
    current_plan_sha256: str,
    current_master_sha256: str,
    current_checker_sha256: str,
    field_coverage: dict[str, int],
) -> str:
    if receipt["schema_version"] != RECEIPT_SCHEMA_VERSION:
        fail("receipt schema_version drift")
    if (
        isinstance(receipt["forbidden_match_count"], bool)
        or not isinstance(receipt["forbidden_match_count"], int)
        or receipt["forbidden_match_count"] != 0
    ):
        fail("receipt forbidden_match_count must be exact integer zero")
    receipt_hashes = (
        receipt["plan_blob_sha256"],
        receipt["master_blob_sha256"],
        receipt["checker_blob_sha256"],
    )
    current_hashes = (current_plan_sha256, current_master_sha256, current_checker_sha256)
    if any(not isinstance(value, str) or re.fullmatch(r"[0-9a-f]{64}", value) is None for value in receipt_hashes):
        fail("receipt content hash field type drift")
    hash_matches = tuple(receipt_value == current_value for receipt_value, current_value in zip(receipt_hashes, current_hashes, strict=True))
    pending = (
        receipt["field_coverage"] == "NOT_RUN"
        and receipt["verdict"] == "NOT_READY"
        and receipt["reviewer"] == PENDING_RECEIPT_REVIEWER
        and receipt["reviewed_at"] is None
    )
    receipt_coverage = receipt["field_coverage"]
    exact_accepted_coverage = (
        isinstance(receipt_coverage, dict)
        and tuple(receipt_coverage) == tuple(PHASE129_FIELD_COVERAGE)
        and all(type(value) is int for value in receipt_coverage.values())
        and receipt_coverage == field_coverage
    )
    accepted = (
        exact_accepted_coverage
        and receipt["verdict"] == ACCEPTED_RECEIPT_VERDICT
        and receipt["reviewer"] == ACCEPTED_RECEIPT_REVIEWER
        and isinstance(receipt["reviewed_at"], str)
        and REVIEWED_AT_RE.fullmatch(receipt["reviewed_at"]) is not None
    )
    if pending:
        if any(hash_matches):
            fail("pending receipt mixes current and stale content hashes")
        return PENDING_RECEIPT_STATE
    if accepted:
        if not all(hash_matches):
            fail("accepted receipt contains stale content hashes")
        try:
            datetime.strptime(receipt["reviewed_at"], "%Y-%m-%dT%H:%M:%SZ")
        except ValueError:
            fail("accepted receipt reviewed_at is not a valid canonical UTC timestamp")
        return ACCEPTED_RECEIPT_STATE
    fail("receipt fields mix pending, accepted, or stale review states")

def git_ok() -> None:
    result = subprocess.run(("git", "merge-base", "--is-ancestor", SEALED_PLAN_COMMIT, "HEAD"), check=False, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    if result.returncode: fail("sealed plan commit is not an ancestor of HEAD")
    for path in PLAN_PATHS:
        result = subprocess.run(("git", "cat-file", "-e", f"{SEALED_PLAN_COMMIT}:{path}"), check=False, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        if result.returncode: fail(f"sealed plan commit lacks {path}")

def main() -> int:
    arguments = sys.argv[1:]
    if arguments not in ([], [POST_G129P_STATE_ARGUMENT]) or os.name != "posix" or not hasattr(os, "O_NOFOLLOW"):
        fail("canonical checker is Linux-only, accepts no input paths, and permits only its explicit post-G129-P state")
    post_materialization = arguments == [POST_G129P_STATE_ARGUMENT]
    root_fd = os.open(".", os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW)
    try:
        if len(set(PLAN_PATHS)) != 21: fail("embedded plan allowlist drift")
        if phase_authority(root_fd) != set(PLAN_PATHS): fail("active Phase 128–141 PLAN.md authority is not exact")
        raw = {path: open_read(root_fd, path) for path in PLAN_PATHS}
        texts = {path: data.decode("utf-8") for path, data in raw.items()}
        receipt = json_load_exact(open_read(root_fd, RECEIPT_PATH, MAX_RECORD_BYTES), RECEIPT_PATH)
        master = open_read(root_fd, MASTER_PATH)
        tool_sources = {path: open_read(root_fd, path) for path in (BUILDER_PATH, VALIDATOR_PATH)}
        cutover_manifest: bytes | None = None
        if post_materialization:
            cutover_manifest = open_read(root_fd, CUTOVER_MANIFEST_PATH, MAX_FILE_BYTES)
        else:
            try:
                open_read(root_fd, CUTOVER_MANIFEST_PATH, MAX_FILE_BYTES)
            except FileNotFoundError:
                pass
            else:
                fail("pre-G129-P checker state requires the future manifest to remain absent")
        checker = open_read(root_fd, ".planning/audits/qa-plans/validate-downstream-plan-contract.py")
    finally: os.close(root_fd)
    if not isinstance(receipt, dict) or tuple(receipt) != RECEIPT_KEYS: fail("receipt keys/order drift")
    if receipt["full_head"] != SEALED_PLAN_COMMIT or receipt["ordered_plan_paths"] != list(PLAN_PATHS): fail("receipt provenance/path drift")
    joined = "\n".join(texts.values())
    if any(re.search(token, joined) for token in FORBIDDEN): fail("active forbidden target token")
    active_texts = {**texts, MASTER_PATH: master.decode("utf-8")}
    command_spans, positive_invocations = validate_active_authority_prose(active_texts)
    tool_contract = phase129_tool_contract(tool_sources)
    all_blocks: list[str] = []
    for path in PLAN_PATHS:
        if path.endswith("/129-tf-e2e-shared/PLAN.md"): continue
        current = blocks(texts[path], path); all_blocks.extend(current)
        for block in current: task_contract(block, path, path in PLAN_PATHS[:9])
    numbered = [texts[path] for path in PLAN_PATHS[2:9]]
    phase129_tasks = sum(len(blocks(text, "129")) for text in numbered)
    negative_self_sentinels()
    field_coverage, ids, operation_count, maximum_batch, plan_batches = validate_phase129_requirements(
        PHASE129_NUMBERED_PATHS,
        tuple(numbered),
        texts[PLAN_PATHS[1]],
        None,
        command_spans,
        positive_invocations,
    )
    validate_phase129_tool_projection(tool_contract, ids, plan_batches)
    if post_materialization:
        if cutover_manifest is None:
            fail("post-G129-P checker state lacks the materialized authority")
        validate_materialized_phase129_authority(
            plan_batches,
            manifest_batch_authority(json_load_exact(cutover_manifest, CUTOVER_MANIFEST_PATH)),
        )
    if (
        phase129_tasks != 176
        or len(ids) != 173
        or len(set(ids)) != 173
        or operation_count != 1120
        or maximum_batch != 56
    ):
        fail("Phase 129 sealed arithmetic drift")
    if receipt["task_count"] != {"phase_129":176,"later_130_141":61} or receipt["batch_count"] != 173 or receipt["operation_count"] != 1120 or receipt["maximum_batch"] != 56 or receipt["batch_ids_sha256"] != sha("\n".join(ids).encode()): fail("receipt arithmetic drift")
    later = PLAN_PATHS[9:]; later_blocks: list[str] = []
    for path, count in zip(later, LATER_COUNTS, strict=True):
        phase = PurePosixPath(path).parts[2][:3]; phase_text = texts[path]; current = blocks(phase_text, path)
        if len(current) != count or "NOT_READY" not in phase_text or (phase != "130" and f"G-ACTIVATE-{phase}" not in phase_text): fail(f"{path}: task/activation drift")
        task_map = {re.search(r'id="([^\"]+)"', block).group(1): block for block in current}
        docs_records = re.findall(r'<docs_mutation task_id="([^\"]+)" paths="([^\"]+)"/>', phase_text)
        activation_records = re.findall(r'<task_contract task_id="([^\"]+)" profile="([^\"]+)" docs="([^\"]+)" feature="([^\"]+)" risk_class="([^\"]+)" files_ref="([^\"]+)" rollback_stop="([^\"]+)" founder_id="([^\"]+)" predecessor="([^\"]+)"/>', phase_text)
        if phase == "130":
            if len(docs_records) != len(current) or len({item[0] for item in docs_records}) != len(current): fail(f"{path}: docs contract coverage drift")
            docs_by_task = dict(docs_records)
        else:
            if len(activation_records) != len(current) or len({item[0] for item in activation_records}) != len(current): fail(f"{path}: activation contract coverage drift")
            docs_by_task = {item[0]: item[2] for item in activation_records}
            for task_id, profile, docs, feature, risk, files_ref, rollback, founder, predecessor in activation_records:
                if task_id not in task_map or not all((feature, risk, rollback)) or files_ref != f"task:{task_id}/files" or founder != f"G-ACTIVATE-{phase}" or predecessor != "G130-H": fail(f"{path}: malformed activation contract")
                match = re.search(r'<verify profile="([^\"]+)">', task_map[task_id])
                if not match or match.group(1) != profile: fail(f"{path}: activation profile mismatch")
                if any(token in " ".join((feature, risk, rollback)) for token in ("*", "range", "universal", "bulk", "all phases")): fail(f"{path}: broad activation contract")
        closure_count = 0
        for task_id, block in task_map.items():
            if "NOT_READY" not in block or (phase != "130" and f"G-ACTIVATE-{phase}" not in block): fail(f"{path}: task activation drift")
            files = re.search(r"<files>(.*?)</files>", block, re.S).group(1)
            file_tokens = literal_file_tokens(files, path)
            declared_docs = docs_by_task.get(task_id)
            if declared_docs is None: fail(f"{path}: missing per-task docs declaration")
            if declared_docs != "none":
                closure_count += 1
                declared_set = set()
                for doc in declared_docs.split(","):
                    doc = doc.strip()
                    if not doc or doc in declared_set or "*" in doc or ".." in PurePosixPath(doc).parts or not doc.startswith("testframework/docs/") or doc.endswith(".bak") or ".bak/" in doc or doc not in file_tokens: fail(f"{path}: docs path outside exact literal Files")
                    declared_set.add(doc)
            if any(token.startswith("testframework/docs/") for token in file_tokens) and declared_docs == "none": fail(f"{path}: docs Files outside closure task")
            if re.search(r"\b(?:write|update|create|modify)\b.{0,80}(?:docs|markdown)\b", block, re.I | re.S) and declared_docs == "none": fail(f"{path}: undeclared docs mutation")
        if closure_count != 1: fail(f"{path}: requires exactly one dependent docs closure")
        later_blocks.extend(current)
    if len(later_blocks) != 61: fail("later count drift")
    raw_profile_count = len(re.findall(r'<profile\b', texts[later[0]]))
    profile_rows = re.findall(r'<profile name="([^\"]+)" image_role="([^\"]+)" limit_class="([^\"]+)" target="([^\"]+)" deadline="([^\"]+)"/>', texts[later[0]])
    if raw_profile_count != len(profile_rows) or len(profile_rows) != 47 or len({row[0] for row in profile_rows}) != 47 or any(not all(row) for row in profile_rows): fail("profile registry drift")
    profiles = {row[0] for row in profile_rows}; used = re.findall(r'<verify profile="([^\"]+)">', "\n".join(later_blocks))
    if not used or set(used) - profiles or len(set(used)) != 47: fail("profile use closure drift")
    if any("testframework/runner/profiles.py" in block for block in later_blocks): fail("later profile writer")
    if re.search(r"(?:G-ACTIVATE-(?:13[1-9]|140|141).{0,120}(?:all|range|universal|bulk)|(?:all|range|universal|bulk).{0,120}G-ACTIVATE)", joined, re.I | re.S): fail("bulk activation language")
    if re.search(r"G130-H\s*(?:->|→)\s*(?:13[1-9]|140|141)", joined): fail("automatic dependency beyond G130-H")
    current_plan_sha256 = closure_hash(raw)
    receipt_state = validate_receipt_state(
        receipt,
        current_plan_sha256,
        sha(master),
        sha(checker),
        field_coverage,
    )
    if post_materialization and receipt_state != ACCEPTED_RECEIPT_STATE:
        fail("post-G129-P state requires the exact independently accepted receipt")
    git_ok()
    checker_state = "POST_G129P_PRESENT" if post_materialization else "PRE_G129P_ABSENT"
    output_verdict = ACCEPTED_RECEIPT_VERDICT if receipt_state == ACCEPTED_RECEIPT_STATE else "NOT_READY"
    print(json.dumps({"verdict":output_verdict,"runtime_status":"NOT_RUN","receipt_state":receipt_state,"materialization_state":checker_state,"sealed_plan_commit":SEALED_PLAN_COMMIT,"plan_blob_sha256":current_plan_sha256,"field_coverage":field_coverage}, sort_keys=True))
    return 0

if __name__ == "__main__":
    try: raise SystemExit(main())
    except (ValueError, json.JSONDecodeError) as error:
        print(json.dumps({"verdict":"FAIL","error":str(error)}, sort_keys=True), file=sys.stderr); raise SystemExit(2)
