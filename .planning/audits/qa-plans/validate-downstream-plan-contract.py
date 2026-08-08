#!/usr/bin/env python3
"""Fail-closed static contract check for the 128–141 live-plan authority."""
from __future__ import annotations

import hashlib
import json
import os
import re
import stat
import subprocess
import sys
from pathlib import PurePosixPath

MAX_FILE_BYTES = 2 * 1024 * 1024
MAX_RECORD_BYTES = 256 * 1024
MAX_PHASE_ENTRIES = 32
SEALED_PLAN_COMMIT = "30333f1bb5bd10a9e86e9a93a5f71fdcdcc13ddf"
MASTER_PATH = ".planning/active/august-2026/06-TEST-ARCHITECTURE-REBASE.md"
RECEIPT_PATH = ".planning/audits/qa-plans/downstream-plan-contract-receipt.v1.json"
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
LATER_COUNTS = (4, 4, 4, 3, 8, 18, 3, 3, 5, 3, 3, 3)
RECEIPT_KEYS = ("schema_version", "full_head", "checker_blob_sha256", "master_blob_sha256", "ordered_plan_paths", "plan_blob_sha256", "task_count", "batch_ids_sha256", "batch_count", "operation_count", "maximum_batch", "forbidden_match_count", "field_coverage", "verdict", "reviewer", "reviewed_at")
FORBIDDEN = (r"testframework/core/", r"testframework/requirements\.lock(?![\w.-])", r"testframework/pytest\.ini(?![\w.-])", r"testframework/docker/Dockerfile(?![\w.-])", r"testframework/docker/compose\.yaml(?![\w.-])", r"\.gitkeep(?![\w.-])")

def fail(message: str) -> None: raise ValueError(message)
def sha(data: bytes) -> str: return hashlib.sha256(data).hexdigest()

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

def closure_hash(texts: dict[str, bytes]) -> str:
    material = b"".join(path.encode("utf-8") + b"\0" + texts[path] for path in PLAN_PATHS)
    return sha(material)

def git_ok() -> None:
    result = subprocess.run(("git", "merge-base", "--is-ancestor", SEALED_PLAN_COMMIT, "HEAD"), check=False, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    if result.returncode: fail("sealed plan commit is not an ancestor of HEAD")
    for path in PLAN_PATHS:
        result = subprocess.run(("git", "cat-file", "-e", f"{SEALED_PLAN_COMMIT}:{path}"), check=False, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        if result.returncode: fail(f"sealed plan commit lacks {path}")

def main() -> int:
    if len(sys.argv) != 1 or os.name != "posix" or not hasattr(os, "O_NOFOLLOW"): fail("canonical checker is Linux-only and accepts no input paths")
    root_fd = os.open(".", os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW)
    try:
        if len(set(PLAN_PATHS)) != 21: fail("embedded plan allowlist drift")
        if phase_authority(root_fd) != set(PLAN_PATHS): fail("active Phase 128–141 PLAN.md authority is not exact")
        raw = {path: open_read(root_fd, path) for path in PLAN_PATHS}
        texts = {path: data.decode("utf-8") for path, data in raw.items()}
        receipt = json.loads(open_read(root_fd, RECEIPT_PATH, MAX_RECORD_BYTES).decode("utf-8"))
        master = open_read(root_fd, MASTER_PATH)
        checker = open_read(root_fd, ".planning/audits/qa-plans/validate-downstream-plan-contract.py")
    finally: os.close(root_fd)
    if tuple(receipt) != RECEIPT_KEYS: fail("receipt keys/order drift")
    if receipt["full_head"] != SEALED_PLAN_COMMIT or receipt["ordered_plan_paths"] != list(PLAN_PATHS): fail("receipt provenance/path drift")
    if receipt["plan_blob_sha256"] != closure_hash(raw) or receipt["master_blob_sha256"] != sha(master) or receipt["checker_blob_sha256"] != sha(checker): fail("receipt hash drift")
    joined = "\n".join(texts.values())
    if any(re.search(token, joined) for token in FORBIDDEN): fail("active forbidden target token")
    all_blocks: list[str] = []
    for path in PLAN_PATHS:
        if path.endswith("/129-tf-e2e-shared/PLAN.md"): continue
        current = blocks(texts[path], path); all_blocks.extend(current)
        for block in current: task_contract(block, path, path in PLAN_PATHS[:9])
    numbered = [texts[path] for path in PLAN_PATHS[2:9]]
    phase129_tasks = sum(len(blocks(text, "129")) for text in numbered)
    ids = re.findall(r"<name>Batch ([^:<]+)", "\n".join(numbered)); ops = [int(value) for value in re.findall(r"ExpectedOperations (\d+)", "\n".join(numbered))]
    if phase129_tasks != 176 or len(ids) != 173 or len(set(ids)) != 173 or len(ops) != 173 or sum(ops) != 1120 or max(ops) != 56: fail("Phase 129 sealed arithmetic drift")
    if receipt["task_count"] != {"phase_129":176,"later_130_141":61} or receipt["batch_count"] != 173 or receipt["operation_count"] != 1120 or receipt["maximum_batch"] != 56 or receipt["batch_ids_sha256"] != sha("\n".join(ids).encode()): fail("receipt arithmetic drift")
    later = PLAN_PATHS[9:]; later_blocks: list[str] = []
    for path, count in zip(later, LATER_COUNTS, strict=True):
        phase = PurePosixPath(path).parts[2][:3]; current = blocks(texts[path], path)
        if len(current) != count or "NOT_READY" not in texts[path] or f"G-ACTIVATE-{phase}" not in texts[path]: fail(f"{path}: task/activation drift")
        for block in current:
            if "NOT_READY" not in block or f"G-ACTIVATE-{phase}" not in block: fail(f"{path}: task activation drift")
            files = re.search(r"<files>(.*?)</files>", block, re.S).group(1)
            if "docs/" in files and not block.startswith(f'<task id="{phase}-') or ("docs/" in files and not re.search(rf'<task id="{phase}-(?:D|H|R|C|E)">', block)): fail(f"{path}: docs ownership outside closure task")
            if "docs/" in files and ("**" in files or "*" in files): fail(f"{path}: broad docs ownership")
        later_blocks.extend(current)
    if len(later_blocks) != 61: fail("later count drift")
    profile_rows = re.findall(r'<profile name="([^\"]+)" image_role="([^\"]+)" limit_class="([^\"]+)" target="([^\"]+)" deadline="([^\"]+)"/>', texts[later[0]])
    if len(profile_rows) != 47 or len({row[0] for row in profile_rows}) != 47 or any(not all(row) for row in profile_rows): fail("profile registry drift")
    profiles = {row[0] for row in profile_rows}; used = re.findall(r'<verify profile="([^\"]+)">', "\n".join(later_blocks))
    if not used or set(used) - profiles or len(set(used)) != 47: fail("profile use closure drift")
    if any("testframework/runner/profiles.py" in block for block in later_blocks): fail("later profile writer")
    if re.search(r"(?:G-ACTIVATE-(?:13[1-9]|140|141).{0,120}(?:all|range|universal|bulk)|(?:all|range|universal|bulk).{0,120}G-ACTIVATE)", joined, re.I | re.S): fail("bulk activation language")
    if re.search(r"G130-H\s*(?:->|→)\s*(?:13[1-9]|140|141)", joined): fail("automatic dependency beyond G130-H")
    git_ok()
    print(json.dumps({"verdict":"NOT_READY","sealed_plan_commit":SEALED_PLAN_COMMIT,"plan_blob_sha256":closure_hash(raw)}, sort_keys=True))
    return 0

if __name__ == "__main__":
    try: raise SystemExit(main())
    except (ValueError, json.JSONDecodeError) as error:
        print(json.dumps({"verdict":"FAIL","error":str(error)}, sort_keys=True), file=sys.stderr); raise SystemExit(2)
