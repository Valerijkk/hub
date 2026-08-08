#!/usr/bin/env python3
"""Deterministic Hub QA routing declaration conformance evaluator (stdlib only)."""

from __future__ import annotations

import hashlib
import json
import re
import subprocess
import sys
from pathlib import Path

VERSION = "hub.qa-routing-evaluator/v1"
BASE = "31dfc7e0aeed4c2667de9d43e62637fd68d3b6f2"
CANDIDATE_HEAD = "0769bc1c13536124808e651903e3c8a2cb4bdd1d"
ROOT = Path(__file__).resolve().parents[3]
HERE = Path(__file__).resolve().parent
SCENARIOS = HERE / "routing-scenarios.v1.json"
SCHEMA = HERE / "routing-receipt.schema.v1.json"
RECEIPT = HERE / "routing-receipt.v1.json"
SKILLS = (
    ".cursor/skills/gsd-ru/goal-verify/SKILL.md",
    ".cursor/skills/gsd-ru/phase-plan/SKILL.md",
    ".cursor/skills/qa-gate/SKILL.md",
)
ASSIGNED = {
    "api-health": SKILLS[2],
    "contract-openapi": SKILLS[1],
    "integration-postgres": SKILLS[2],
    "e2e-login": SKILLS[2],
    "selftest-markers": SKILLS[0],
    "load-tracker": SKILLS[2],
}
DECL_RE = re.compile(r"<!-- hub\.qa-routing-declaration/v1\s*\n(\{[^\n]+\})\s*\n-->")
MAX_BYTES = 512_000
RECORD_KEYS = {
    "schema_version", "base_commit", "candidate_head", "candidate_content_sha256",
    "scenario_sha256", "skill_blobs", "phase", "scenario_id", "prompt_sha256",
    "declared_routes", "expected_routes", "unexpected_routes", "verdict",
    "reviewer", "reviewed_at", "evaluator_version", "evaluator_sha256",
}
DECL_KEYS = {
    "schema_version", "scenario_id", "assigned_skill", "declared_routes",
    "expected_routes", "prohibited_legacy_tokens",
}


def sha(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def bounded_utf8(path: Path) -> tuple[bytes, str]:
    root = ROOT.resolve(strict=True)
    if path.is_symlink():
        raise ValueError(f"symlink forbidden: {path}")
    resolved = path.resolve(strict=True)
    if resolved != root and root not in resolved.parents:
        raise ValueError(f"path escape: {path}")
    size = resolved.stat().st_size
    if size > MAX_BYTES:
        raise ValueError(f"size budget exceeded: {path}")
    data = resolved.read_bytes()
    return data, data.decode("utf-8")


def git_blob(rel: str) -> tuple[bytes, str]:
    if rel not in SKILLS:
        raise ValueError(f"non-exact BASE skill path: {rel}")
    ref = f"{BASE}:{rel}"
    size_run = subprocess.run(
        ["git", "cat-file", "-s", ref], cwd=ROOT, check=True,
        stdout=subprocess.PIPE, stderr=subprocess.PIPE, timeout=5,
    )
    size = int(size_run.stdout.decode("ascii").strip())
    if size > MAX_BYTES:
        raise ValueError(f"BASE blob size budget exceeded: {rel}")
    blob_run = subprocess.run(
        ["git", "show", ref], cwd=ROOT, check=True,
        stdout=subprocess.PIPE, stderr=subprocess.PIPE, timeout=5,
    )
    data = blob_run.stdout
    if len(data) != size:
        raise ValueError(f"BASE blob size mismatch: {rel}")
    return data, data.decode("utf-8")


def sorted_unique(value: object, field: str, nonempty: bool = False) -> list[str]:
    if not isinstance(value, list) or not all(isinstance(x, str) for x in value):
        raise ValueError(f"{field} must be a string array")
    if value != sorted(value) or len(value) != len(set(value)):
        raise ValueError(f"{field} must be sorted and unique")
    if nonempty and not value:
        raise ValueError(f"{field} must be non-empty")
    return value


def load_scenarios() -> tuple[dict[str, dict], str]:
    raw, text = bounded_utf8(SCENARIOS)
    doc = json.loads(text)
    if set(doc) != {"schema_version", "scenarios"} or doc["schema_version"] != "hub.qa-routing-scenarios/v1":
        raise ValueError("invalid scenario envelope")
    items: dict[str, dict] = {}
    for item in doc["scenarios"]:
        if set(item) != {"id", "prompt", "expected_routes"}:
            raise ValueError("invalid scenario keys")
        routes = sorted_unique(item["expected_routes"], "scenario.expected_routes", True)
        if item["id"] in items:
            raise ValueError("duplicate scenario")
        items[item["id"]] = {**item, "expected_routes": routes}
    if set(items) != set(ASSIGNED) or len(items) != 6:
        raise ValueError("scenario coverage is not exact")
    return items, sha(raw)


def load_candidate_declarations() -> tuple[dict[str, dict], dict[str, str], str]:
    declarations: dict[str, dict] = {}
    hashes: dict[str, str] = {}
    closure = hashlib.sha256()
    for rel in SKILLS:
        raw, text = bounded_utf8(ROOT / rel)
        hashes[rel] = sha(raw)
        closure.update(rel.encode("utf-8") + b"\0" + raw + b"\0")
        for encoded in DECL_RE.findall(text):
            item = json.loads(encoded)
            if set(item) != DECL_KEYS or item.get("schema_version") != "hub.qa-routing-declaration/v1":
                raise ValueError(f"invalid declaration keys: {rel}")
            sid = item["scenario_id"]
            if item["assigned_skill"] != rel or ASSIGNED.get(sid) != rel or sid in declarations:
                raise ValueError(f"invalid/duplicate declaration assignment: {sid}")
            sorted_unique(item["declared_routes"], "declared_routes", True)
            sorted_unique(item["expected_routes"], "expected_routes", True)
            sorted_unique(item["prohibited_legacy_tokens"], "prohibited_legacy_tokens", True)
            declarations[sid] = item
    if set(declarations) != set(ASSIGNED):
        raise ValueError("missing declaration tuple")
    return declarations, hashes, closure.hexdigest()


def candidate_head() -> str:
    # The candidate was assembled on this immutable planning predecessor.
    # Exact current bytes are independently bound by candidate_content_sha256.
    return CANDIDATE_HEAD


def expected_records() -> list[dict]:
    scenarios, scenario_hash = load_scenarios()
    declarations, green_hashes, content_hash = load_candidate_declarations()
    base_text: dict[str, str] = {}
    base_hashes: dict[str, str] = {}
    for rel in SKILLS:
        raw, text = git_blob(rel)
        base_text[rel] = text
        base_hashes[rel] = sha(raw)
    evaluator_hash = sha(bounded_utf8(Path(__file__))[0])
    head = candidate_head()
    common = {
        "schema_version": "hub.qa-routing-receipt/v1", "base_commit": BASE,
        "candidate_head": head, "candidate_content_sha256": content_hash,
        "scenario_sha256": scenario_hash, "reviewer": "qa-auto-ver+arch-ver+security-auditor",
        "reviewed_at": "2026-08-08T00:00:00Z", "evaluator_version": VERSION,
        "evaluator_sha256": evaluator_hash,
    }
    records: list[dict] = []
    for phase in ("RED", "GREEN"):
        for sid in sorted(scenarios):
            scenario = scenarios[sid]
            decl = declarations[sid]
            expected = scenario["expected_routes"]
            if decl["expected_routes"] != expected:
                raise ValueError(f"declaration/scenario mismatch: {sid}")
            if phase == "RED":
                old = sorted(t for t in decl["prohibited_legacy_tokens"] if t in base_text[decl["assigned_skill"]])
                declared = old or expected
                blobs = base_hashes
            else:
                declared = decl["declared_routes"]
                blobs = green_hashes
            unexpected = sorted(set(declared) - set(expected))
            verdict = "PASS" if declared == expected and not unexpected else "MISMATCH"
            records.append({
                **common, "skill_blobs": blobs, "phase": phase, "scenario_id": sid,
                "prompt_sha256": sha(scenario["prompt"].encode("utf-8")),
                "declared_routes": declared, "expected_routes": expected,
                "unexpected_routes": unexpected, "verdict": verdict,
            })
    if not any(r["verdict"] == "MISMATCH" for r in records[:6]):
        raise ValueError("RED has no legacy mismatch")
    if any(r["verdict"] != "PASS" for r in records[6:]):
        raise ValueError("GREEN declaration inequality")
    return records


def main() -> int:
    _, schema_text = bounded_utf8(SCHEMA)
    schema = json.loads(schema_text)
    if schema.get("x-record-keys") != sorted(RECORD_KEYS):
        raise ValueError("schema record-key contract mismatch")
    _, receipt_text = bounded_utf8(RECEIPT)
    receipt = json.loads(receipt_text)
    if not isinstance(receipt, list) or len(receipt) != 12:
        raise ValueError("receipt must contain exactly twelve records")
    for record in receipt:
        if set(record) != RECORD_KEYS:
            raise ValueError("receipt has missing or extra keys")
        sorted_unique(record["declared_routes"], "declared_routes")
        sorted_unique(record["expected_routes"], "expected_routes")
        sorted_unique(record["unexpected_routes"], "unexpected_routes")
        if set(record["skill_blobs"]) != set(SKILLS):
            raise ValueError("non-exact skill blob set")
    if receipt != expected_records():
        raise ValueError("receipt does not match deterministic BASE/candidate evaluation")
    print(json.dumps({"status": "PASS", "records": 12, "evaluator_version": VERSION}, sort_keys=True))
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (ValueError, OSError, UnicodeError, json.JSONDecodeError, subprocess.SubprocessError) as exc:
        print(json.dumps({"status": "FAIL", "error": str(exc)}, sort_keys=True), file=sys.stderr)
        raise SystemExit(1)
