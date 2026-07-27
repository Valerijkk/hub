#!/usr/bin/env python3
"""C2: sync ROADMAP Progress table with PHASE-STATUS-MATRIX.md facts."""
from __future__ import annotations

import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
ROADMAP = ROOT / ".planning" / "ROADMAP.md"
MATRIX = ROOT / ".planning" / "planning-audit-2026-07" / "PHASE-STATUS-MATRIX.md"

# Matrix facts (cite, do not invent) — scanned 2026-07-27
MATRIX_PHASE_DIRS = 124  # .planning/phases/*/ including 126–136 + 999
MATRIX_VERIFIED_PASS = 63  # live matrix summary after PW2 retro-PASS
MATRIX_VERIFIED_PARTIAL = 5
MATRIX_CODE_GATE_DEPLOY_BLOCKED = 2
MATRIX_NO_VERIFICATION = 54

# Project convention from 00-PLANNING-AUDIT.md §0
# milestones 01–13 (13) + phases dirs 14–125 (112) = 125
TOTAL_PHASES_CONVENTION = 125
MILESTONES_01_13 = 13
PHASES_14_125 = 112

text = ROADMAP.read_text(encoding="utf-8")

# Parse checklist for titles/dates
pre = text.split("## Phase Details")[0]
checklist: dict[int, tuple[str, str]] = {}
for m in re.finditer(
    r"- \[x\] \*\*Phase (\d+): ([^*]+)\*\* - (.+)$", pre, re.M
):
    num = int(m.group(1))
    title = m.group(2).strip()
    blurb = m.group(3).strip()
    date_m = re.search(r"(20\d{2}-\d{2}-\d{2})", blurb)
    date = date_m.group(1) if date_m else "—"
    checklist[num] = (title, date)

# Parse matrix classification for 14-125 where present
matrix = MATRIX.read_text(encoding="utf-8")
class_by_num: dict[int, str] = {}
for m in re.finditer(
    r"^\| (\d+)-[a-z0-9-]+ \|[^|]+\|[^|]+\|[^|]+\|[^|]+\|[^|]+\| ([A-Z_]+) \|",
    matrix,
    re.M,
):
    class_by_num[int(m.group(1))] = m.group(2)


def milestone_for(n: int) -> str:
    if n <= 7:
        return "v1.0"
    if n <= 14:
        return "v2.0"
    if n == 15:
        return "v2.5"
    if n == 16:
        return "v2.6"
    if n <= 20:
        return "v3.x"
    if n <= 38:
        return "arch/obs"
    if n <= 58:
        return "v4"
    return "cov/gates"


def status_for(n: int) -> str:
    if n in (63, 67):
        return "Code PASS / deploy blocked"
    cls = class_by_num.get(n)
    if cls == "VERIFIED_PASS":
        return "Verified PASS"
    if cls == "VERIFIED_PARTIAL":
        return "Verified PARTIAL"
    if cls == "CODE_GATE_DEPLOY_BLOCKED":
        return "Code PASS / deploy blocked"
    if cls == "NO_VERIFICATION":
        if n <= 17:
            return "Complete (no VERIFICATION.md in phases/)"
        return "Tasks done / no VERIFICATION.md"
    # 01-13 live in milestones — treat Complete
    if n <= 13:
        return "Complete"
    return "Complete"


rows: list[str] = [
    "| Phase | Milestone | Plans Complete | Status | Completed |",
    "|-------|-----------|----------------|--------|-----------|",
]

# Keep richer plan counts for 1-17 from existing table where known
legacy_plans = {
    1: "9/9",
    2: "5/5",
    3: "5/5",
    4: "5/5",
    5: "4/4",
    6: "5/5",
    7: "6/6",
    8: "2/2",
    9: "2/2",
    10: "3/3",
    11: "4/4",
    12: "5/5",
    13: "2/2",
    14: "5/5",
    15: "Complete",
    16: "Complete",
    17: "Complete",
}

for n in range(1, 126):
    title, date = checklist.get(n, (f"Phase {n}", "—"))
    plans = legacy_plans.get(n, "see PLAN")
    # Prefer historical dates for early phases if checklist date missing
    if n <= 7 and date == "—":
        date = "2026-03-21"
    elif 8 <= n <= 13 and date == "—":
        date = "2026-03-28"
    elif n == 14 and date == "—":
        date = "2026-05-15"
    rows.append(
        f"| {n}. {title} | {milestone_for(n)} | {plans} | {status_for(n)} | {date} |"
    )

progress_body = f"""## Progress

**Execution Order:**
Phases execute in numeric order (1 → 125). Decimal insertions (if any) slot between surrounding integers.

### Honest counters (synced 2026-07-27 · C2)

Sources: [`PHASE-STATUS-MATRIX.md`](planning-audit-2026-07/PHASE-STATUS-MATRIX.md) (live scan A3), [`00-PLANNING-AUDIT.md`](planning-audit-2026-07/00-PLANNING-AUDIT.md) §0, `STATE.md` (`total_phases: 125` — not yet rewritten by A4/`goals_verified`).

| Metric | Value | Source |
|--------|------:|--------|
| `total_phases` (project convention) | **{TOTAL_PHASES_CONVENTION}** | milestones 01–13 (**{MILESTONES_01_13}**) + `.planning/phases/` dirs **14–125** (**{PHASES_14_125}**) = 125 (`00-PLANNING-AUDIT.md` §0). `STATE.md` still lists `total_phases: 125`. |
| Phase dirs under `.planning/phases/` (live) | **{MATRIX_PHASE_DIRS}** | Matrix: includes 14–125 plus later TF dirs 126–136 and backlog `999-test-framework` (not all counted in the 125 convention). |
| `goals_verified` (`VERIFIED_PASS`) | **{MATRIX_VERIFIED_PASS}** | Matrix classification `VERIFIED_PASS` count (has `VERIFICATION.md` + PASS). **Not** equal to `completed_phases: 125`. |
| `VERIFIED_PARTIAL` | **{MATRIX_VERIFIED_PARTIAL}** | Matrix |
| `CODE_GATE_DEPLOY_BLOCKED` | **{MATRIX_CODE_GATE_DEPLOY_BLOCKED}** | Phases **63**, **67** — code gate PASS, deploy blocked |
| `NO_VERIFICATION` | **{MATRIX_NO_VERIFICATION}** | Matrix (includes many early shipped phases + TF/backlog dirs) |

**Deploy blockers** (do not read Progress 100% as “prod unlocked”):

| Phase | Verdict | Blocker |
|------:|---------|---------|
| 63 | PASS (code gate) | deploy blocked on **user commit/PR** + **PG-03** ops |
| 67 | PASS (code gate) | same — **PG-03 / FA-06** secret rotation still manual (`production/PG-03-SECRET-ROTATION.md`, `FA-06-EXECUTION-LOG.md`) |

> **STATE.md note:** A4/`goals_verified` frontmatter not yet present at C2 time — counters above are matrix-backed, not copied from stale `percent: 100`.

{chr(10).join(rows)}

> **Сноска о сверке (drift fixed 2026-07-03, Волна 1 задача 1.13):** чекбоксы Phase 8/9/12 и строки Progress приведены в соответствие с фактом — v2.0 Hub Redesign завершён 2026-05-15 (подтверждено по `.planning/STATE.md` «v2.0 Hub Redesign ✅ COMPLETE» и `AGENTS.md`; планы 08-01/08-02, 09-01/09-02, 12-01…12-05 отмечены [x] в Phase Details). Дополнительно удалены пять дублирующихся блоков «Plans: 12-0X-PLAN.md», ошибочно вставленных в Phase Details фаз 2, 3, 4 и 6 (v1.0), — они дублировали планы Phase 12 с противоречивыми чекбоксами. Строки Phase 15–17 обновлены по факту milestone-статусов (v2.5/v2.6/v3.0 ✅ COMPLETE, см. раздел Milestones выше).

> **Сноска C2 (2026-07-27):** Progress-таблица расширена до Phase 125; статусы VERIFIED_* / deploy-blocked взяты из `PHASE-STATUS-MATRIX.md`. Чекбокс «Complete» в Phases overview ≠ `goals_verified`.
"""

# Replace from ## Progress through just before ## Backlog
m = re.search(r"\n## Progress\n.*?(?=\n---\n\n## Backlog\n)", text, re.S)
if not m:
    # try without --- separator
    m = re.search(r"\n## Progress\n.*?(?=\n## Backlog\n)", text, re.S)
    if not m:
        raise SystemExit("Progress→Backlog span not found")
    replacement = "\n" + progress_body.rstrip() + "\n\n"
else:
    replacement = "\n" + progress_body.rstrip() + "\n\n---\n\n"

text = text[: m.start()] + replacement + text[m.end() :]
ROADMAP.write_text(text, encoding="utf-8")
print("C2 Progress synced")
print("rows", 125)
print("verified_pass", MATRIX_VERIFIED_PASS)
print("deploy_blocked", MATRIX_CODE_GATE_DEPLOY_BLOCKED)
