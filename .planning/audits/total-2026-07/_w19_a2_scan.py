#!/usr/bin/env python3
"""W19-A2: scan TASK-009…056 for incomplete evidence triad after W18."""
from __future__ import annotations

import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
BACKLOG = ROOT / "backlog"
AUDIT = Path(__file__).resolve().parent

TASK_RE = re.compile(r"TASK-0(0[9]|[1-4]\d|5[0-6])-")
STATUS_RE = re.compile(r"Статус:\s*\**([^*\n]+)")
GO_TEST_RE = re.compile(r"\bTest[A-Za-z0-9_]{5,}\b")
SHA_ANCHOR_RE = re.compile(
    r"(?i)(?:\bSHA\s*[:=]\s*`?([0-9a-f]{7,40})`?"
    r"|\bcommit\s*[:=]\s*`?([0-9a-f]{7,40})`?"
    r"|`([0-9a-f]{7,40})`\s*(?:—|-)\s*(?:fix|test|chore|docs|feat|refactor))"
)
TEMPLATE_EVIDENCE_RE = re.compile(
    r"Evidence:\s*(?:вывод тестов,\s*)?SHA", re.I
)
MUT_REF_RE = re.compile(r"W18-C\d+b?-MUTATION")
PASS_RE = re.compile(r"(?i)\bPASS\b|\*\*PASS\*\*")


def strip_templates(text: str) -> str:
    return "\n".join(
        ln for ln in text.splitlines() if not TEMPLATE_EVIDENCE_RE.search(ln)
    )


def has_sha(text: str) -> bool:
    return bool(SHA_ANCHOR_RE.search(strip_templates(text)))


def has_go_test(text: str) -> bool:
    return bool(GO_TEST_RE.search(text))


def evidence_sections(text: str) -> str:
    parts = []
    for m in re.finditer(
        r"(?is)(?:##\s*Stage C|##\s*\[W18|##\s*\[?W18-C|Mutation:|EVIDENCE).{0,2500}",
        text,
    ):
        parts.append(m.group(0))
    return "\n".join(parts)


def has_test_in_task_body(text: str) -> bool:
    sec = evidence_sections(text) or text
    if has_go_test(sec):
        # Ignore ACCEPTANCE-only mention of mutation without concrete Test name
        if GO_TEST_RE.search(evidence_sections(text)):
            return True
        # Whole-body Go test outside template — also counts (C1 style)
        if GO_TEST_RE.search(text):
            return True
    # Vitest titles in Stage C / evidence backticks
    for m in re.finditer(r"`([^`]+)`", evidence_sections(text)):
        s = m.group(1)
        if s.startswith("Test"):
            return True
        keys = (
            "subscribes",
            "one-retry",
            "listTrash",
            "ignores late",
            "keeps dialog",
            "locale=",
            "gif-picker",
            "PinnedMessages",
            "joinCall",
            "sendMessage",
            "monitorForElements",
            "hides comparison",
            "shows alert",
            "shows loading",
            "refresh failure",
            "sidebar-user",
            "myTasks",
            "WebhooksPage",
            "collapsed Sidebar",
            "defaults to type",
            "clearSession",
            "dismissLater",
            "depends on rooms",
            "rolls back badge",
            "aria-label",
        )
        if any(k in s for k in keys):
            return True
    return False


def task_covered_by_mutation(tid: str, muts: dict[str, str]) -> tuple[bool, str | None]:
    for name, mt in muts.items():
        if re.search(rf"TASK-{tid}\b", mt):
            return True, name
        if re.search(rf"## TASK-{tid}\b", mt):
            return True, name
        # W18-C1 table uses short IDs: | 009 |
        if name.startswith("W18-C1") and re.search(rf"\|\s*{tid}\s*\|", mt):
            return True, name
        if re.search(rf"\| TASK-{tid} \|", mt):
            return True, name
    return False, None


def main() -> None:
    tasks = sorted(p for p in BACKLOG.glob("TASK-0*.md") if TASK_RE.match(p.name))
    muts = {p.name: p.read_text(encoding="utf-8") for p in AUDIT.glob("W18-*-MUTATION.md")}
    all_mut = "\n".join(muts.values())

    fixed: list[str] = []
    open_ids: list[str] = []
    gaps: list[str] = []
    no_test_body: list[str] = []
    test_in_body: list[str] = []
    thin: list[str] = []
    coverage: dict[str, str] = {}

    print(f"mutation files: {sorted(muts)}")
    print(f"SHA in any mutation file: {has_sha(all_mut)}")

    for path in tasks:
        tid = re.match(r"TASK-(\d+)", path.name).group(1)
        body = path.read_text(encoding="utf-8")
        st_m = STATUS_RE.search(body)
        st = st_m.group(1) if st_m else ""
        if "Исправлена" not in st:
            if "Открыта" in st:
                open_ids.append(tid)
            continue
        fixed.append(tid)

        covered, mut_name = task_covered_by_mutation(tid, muts)
        mut_ref = bool(MUT_REF_RE.search(body)) or covered
        if mut_name:
            coverage[tid] = mut_name

        sha_ok = has_sha(body) or has_sha(all_mut)
        test_body = has_test_in_task_body(body)
        # union test: body or mutation log covering task
        test_union = test_body
        if covered and mut_name:
            test_union = test_union or has_go_test(muts[mut_name]) or ("`" in muts[mut_name])

        thin_note = (not test_body) and bool(PASS_RE.search(body)) and mut_ref

        triad_ok = sha_ok and test_union and mut_ref
        if not triad_ok:
            gaps.append(tid)
        if not test_body:
            no_test_body.append(tid)
        else:
            test_in_body.append(tid)
        if thin_note:
            thin.append(tid)

        print(
            f"TASK-{tid}: sha={sha_ok} test_body={test_body} test_union={test_union} "
            f"mut={mut_name or ('ref-only' if mut_ref else None)} thin={thin_note}"
        )

    print("---")
    print(f"fixed={len(fixed)} open={len(open_ids)}")
    print(f"gaps={len(gaps)}")
    print("gap_ids=" + ",".join(gaps))
    print(f"no_test_body={len(no_test_body)} -> {','.join(no_test_body)}")
    print(f"test_in_body={len(test_in_body)} -> {','.join(test_in_body)}")
    print(f"thin={len(thin)} -> {','.join(thin)}")
    print("open=" + ",".join(open_ids))
    assert len(fixed) == 39, fixed
    assert len(open_ids) == 9, open_ids
    assert len(gaps) == 39, gaps
    assert set(gaps) == set(fixed)
    print("OK: all 39 Исправлена = evidence-gap (no SHA in triad)")


if __name__ == "__main__":
    main()
