#!/usr/bin/env python3
"""Bounded, static-only validator for the live downstream plan closure."""
from __future__ import annotations

import hashlib
import json
import os
import re
import stat
import sys
from pathlib import PurePosixPath

ROOT = PurePosixPath(".")
MAX_FILE_BYTES = 2 * 1024 * 1024
PLAN_PATHS = (
    ".planning/phases/128-tf-python-scaffold/PLAN.md",
    ".planning/phases/129-tf-e2e-shared/PLAN.md",
    ".planning/phases/129-tf-e2e-shared/129-01-PLAN.md",
    ".planning/phases/129-tf-e2e-shared/129-02-PLAN.md",
    ".planning/phases/129-tf-e2e-shared/129-03-PLAN.md",
    ".planning/phases/129-tf-e2e-shared/129-04-PLAN.md",
    ".planning/phases/129-tf-e2e-shared/129-05-PLAN.md",
    ".planning/phases/129-tf-e2e-shared/129-06-PLAN.md",
    ".planning/phases/129-tf-e2e-shared/129-07-PLAN.md",
    ".planning/phases/130-tf-traceability-gates/PLAN.md",
    ".planning/phases/131-tf-python-core-auth/PLAN.md",
    ".planning/phases/132-tf-security-platform/PLAN.md",
    ".planning/phases/133-tf-infrastructure-integration/PLAN.md",
    ".planning/phases/134-tf-work-realtime/PLAN.md",
    ".planning/phases/135-tf-business-integrations/PLAN.md",
    ".planning/phases/136-tf-python-contracts/PLAN.md",
    ".planning/phases/137-tf-python-e2e-foundation/PLAN.md",
    ".planning/phases/138-tf-python-e2e-journeys/PLAN.md",
    ".planning/phases/139-tf-accessibility-security/PLAN.md",
    ".planning/phases/140-tf-load-resilience/PLAN.md",
    ".planning/phases/141-tf-ci-final-verification/PLAN.md",
)
LATER_COUNTS = (4, 4, 4, 3, 8, 18, 3, 3, 5, 3, 3, 3)
FORBIDDEN = (
    r"testframework/core/", r"testframework/requirements\.lock(?![A-Za-z0-9_.-])",
    r"testframework/pytest\.ini(?![A-Za-z0-9_.-])", r"testframework/docker/Dockerfile(?![A-Za-z0-9_.-])",
    r"testframework/docker/compose\.yaml(?![A-Za-z0-9_.-])", r"\.gitkeep(?![A-Za-z0-9_.-])",
)
REQUIRED_TASK_FIELDS = ("requirements", "read_first", "consumes", "produces", "files", "failing_sentinel", "verify", "done", "atomic_allowlist")

def fail(message: str) -> None:
    raise ValueError(message)

def read_no_follow(repo_fd: int, relative: str) -> bytes:
    parts = PurePosixPath(relative).parts
    if any(part in ("", ".", "..") for part in parts):
        fail(f"unsafe plan path: {relative}")
    fd = os.dup(repo_fd)
    try:
        for part in parts[:-1]:
            next_fd = os.open(part, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW, dir_fd=fd)
            os.close(fd)
            fd = next_fd
        leaf = os.open(parts[-1], os.O_RDONLY | os.O_NOFOLLOW, dir_fd=fd)
        try:
            info = os.fstat(leaf)
            if not stat.S_ISREG(info.st_mode) or info.st_size > MAX_FILE_BYTES:
                fail(f"unsafe or oversized plan: {relative}")
            chunks: list[bytes] = []
            remaining = MAX_FILE_BYTES + 1
            while remaining:
                chunk = os.read(leaf, min(65536, remaining))
                if not chunk:
                    break
                chunks.append(chunk)
                remaining -= len(chunk)
            data = b"".join(chunks)
            if len(data) > MAX_FILE_BYTES:
                fail(f"oversized plan: {relative}")
            return data
        finally:
            os.close(leaf)
    finally:
        os.close(fd)

def task_blocks(text: str) -> list[str]:
    blocks = re.findall(r"<task\b[^>]*>.*?</task>", text, flags=re.S)
    if not blocks:
        fail("missing task blocks")
    return blocks

def check_task_fields(block: str, path: str) -> None:
    for field in REQUIRED_TASK_FIELDS:
        if not re.search(rf"<{field}(?:\s|>)", block):
            fail(f"{path}: task missing {field}")
    files = re.search(r"<files>(.*?)</files>", block, flags=re.S)
    if not files:
        fail(f"{path}: task missing files")
    expanded = [item for item in re.split(r"[;,]", files.group(1)) if item.strip()]
    if len(expanded) > 14:
        fail(f"{path}: task expands above 14 declared paths")
    if re.search(r"(?<!testframework/)(?:hub_testkit|tests|cases|selftests|docker|load)/", files.group(1)):
        fail(f"{path}: unqualified framework output")

def main() -> int:
    if len(sys.argv) != 1:
        fail("the checker accepts no plan arguments; it uses the embedded allowlist")
    if not hasattr(os, "O_NOFOLLOW") or not hasattr(os, "O_DIRECTORY"):
        fail("descriptor-safe checks require O_NOFOLLOW and O_DIRECTORY")
    repo_fd = os.open(".", os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW)
    try:
        texts: dict[str, str] = {}
        for path in PLAN_PATHS:
            data = read_no_follow(repo_fd, path)
            try:
                texts[path] = data.decode("utf-8")
            except UnicodeDecodeError as exc:
                fail(f"{path}: non-UTF-8 content: {exc}")
    finally:
        os.close(repo_fd)
    if len(set(PLAN_PATHS)) != 21:
        fail("embedded plan allowlist must contain 21 unique paths")
    joined = "\n".join(texts.values())
    for token in FORBIDDEN:
        if re.search(token, joined):
            fail(f"active forbidden target token: {token}")
    phase129 = [path for path in PLAN_PATHS if "/129-" in path and path.endswith("PLAN.md")]
    if len(phase129) != 8:
        fail("Phase 129 must retain master plus seven numbered files")
    numbered = [texts[path] for path in PLAN_PATHS[2:9]]
    task_total = sum(len(task_blocks(text)) for text in numbered)
    batch_ids = re.findall(r"<name>Batch ([^:<]+)", "\n".join(numbered))
    operations = [int(value) for value in re.findall(r"ExpectedOperations (\d+)", "\n".join(numbered))]
    if task_total != 176 or len(batch_ids) != 173 or len(set(batch_ids)) != 173:
        fail("Phase 129 task or batch identity drift")
    if len(operations) != 173 or sum(operations) != 1120 or max(operations) != 56:
        fail("Phase 129 operation arithmetic drift")
    later_paths = PLAN_PATHS[9:]
    later_blocks: list[str] = []
    for path, expected in zip(later_paths, LATER_COUNTS, strict=True):
        if not path.endswith("/PLAN.md"):
            fail(f"later authority must be PLAN.md: {path}")
        text = texts[path]
        phase = PurePosixPath(path).parts[2].split("-", 1)[0]
        activation = f"G-ACTIVATE-{phase}"
        if "NOT_READY" not in text or activation not in text:
            fail(f"{path}: missing NOT_READY activation precondition")
        blocks = task_blocks(text)
        if len(blocks) != expected:
            fail(f"{path}: task-count drift")
        for block in blocks:
            check_task_fields(block, path)
            if "NOT_READY" not in block or activation not in block:
                fail(f"{path}: task lacks explicit activation guard")
        later_blocks.extend(blocks)
    if len(later_blocks) != 61:
        fail("later task total drift")
    registry = re.findall(r"<profile name=\"([^\"]+)\" image_role=\"[^\"]+\" limit_class=\"[^\"]+\" target=\"[^\"]+\" deadline=\"[^\"]+\"/>", texts[later_paths[0]])
    if len(registry) != 47 or len(set(registry)) != 47:
        fail("130-A must declare exactly 47 complete unique profiles")
    requested = re.findall(r"<verify profile=\"([^\"]+)\">", "\n".join(later_blocks))
    if any(profile not in registry for profile in requested):
        fail("unknown later verify profile")
    if any("testframework/runner/profiles.py" in block for block in later_blocks):
        fail("later registry writer is forbidden after 130-A")
    if re.search(r"G130-H\s*(?:->|→)\s*(?:13[1-9]|140|141)", joined):
        fail("automatic dependency beyond G130-H")
    report = {
        "schema_version": "downstream-plan-contract-receipt.v1",
        "ordered_plan_paths": list(PLAN_PATHS),
        "task_count": {"phase_129": 176, "later": 61},
        "batch_count": 173,
        "operation_count": 1120,
        "maximum_batch": 56,
        "forbidden_match_count": 0,
        "field_coverage": "complete",
        "verdict": "STATIC_REVIEW_PASS",
    }
    print(json.dumps(report, sort_keys=True, separators=(",", ":")))
    return 0

if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except ValueError as error:
        print(json.dumps({"verdict": "FAIL", "error": str(error)}, sort_keys=True), file=sys.stderr)
        raise SystemExit(2)
