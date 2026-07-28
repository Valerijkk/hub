#!/usr/bin/env python3
"""Apply C1 ROADMAP restore: phases 18-125 details + checklist gaps."""
from __future__ import annotations

import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
ROADMAP = ROOT / ".planning" / "ROADMAP.md"
AUDIT = Path(__file__).resolve().parent

details = (AUDIT / "_c1_phase_details_18_125.md").read_text(encoding="utf-8")
checklist_65 = (AUDIT / "_c1_checklist_65_125.md").read_text(encoding="utf-8").rstrip()
checklist_39 = (AUDIT / "_c1_checklist_39.md").read_text(encoding="utf-8").strip()

text = ROADMAP.read_text(encoding="utf-8")

needle = "Decimal phases appear between their surrounding integers in numeric order.\n"
note = (
    "Decimal phases appear between their surrounding integers in numeric order.\n"
    "\n"
    "> **Phase artifact locations:** Phases **01–13** live under `.planning/milestones/` "
    "(`v1.0-phases/` for 01–07, `v2.0-phases/` for 08–13). Phases **14–125** "
    "(and later TF/backlog dirs) live under `.planning/phases/`. Convention "
    "`total_phases: 125` = milestones 01–13 (13) + phases dirs 14–125 (112) — "
    "see `planning-audit-2026-07/00-PLANNING-AUDIT.md` §0 and "
    "`planning-audit-2026-07/PHASE-STATUS-MATRIX.md`.\n"
)
if "Phase artifact locations" not in text:
    if needle not in text:
        raise SystemExit("phase numbering needle missing")
    text = text.replace(needle, note, 1)

phases_section = text.split("## Phase Details")[0]
if "**Phase 39:" not in phases_section:
    m = re.search(r"(- \[x\] \*\*Phase 38:[^\n]+\n)(- \[x\] \*\*Phase 40:)", text)
    if not m:
        raise SystemExit("phase 38/40 checklist boundary missing")
    text = (
        text[: m.start(1)]
        + m.group(1)
        + checklist_39
        + "\n"
        + m.group(2)
        + text[m.end(2) :]
    )

phases_section = text.split("## Phase Details")[0]
if "**Phase 65:" not in phases_section:
    m = re.search(r"(- \[x\] \*\*Phase 64:[^\n]+\n)", text)
    if not m:
        raise SystemExit("phase 64 checklist line missing")
    insert_at = m.end(1)
    text = text[:insert_at] + checklist_65 + "\n" + text[insert_at:]

if "### Phase 18:" not in text:
    marker = "\n## Progress\n"
    if marker not in text:
        raise SystemExit("## Progress marker missing")
    text = text.replace(
        marker, "\n" + details.rstrip() + "\n\n## Progress\n", 1
    )
else:
    print("Phase 18 details already present — skip details insert")

ROADMAP.write_text(text, encoding="utf-8")
print("ROADMAP updated")
print("Phase 18 present:", "### Phase 18:" in text)
print("Phase 125 present:", "### Phase 125:" in text)
print(
    "Phase 65 checklist:",
    "**Phase 65:" in text.split("## Phase Details")[0],
)
print(
    "Phase 39 checklist:",
    "**Phase 39:" in text.split("## Phase Details")[0],
)
print("milestones note:", "Phase artifact locations" in text)
print("lines", len(text.splitlines()))
