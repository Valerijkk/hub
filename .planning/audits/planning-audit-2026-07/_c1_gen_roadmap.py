#!/usr/bin/env python3
"""Generate ROADMAP Phase Details for phases 18-125 from PLAN.md files."""
from __future__ import annotations

import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
PHASES = ROOT / ".planning" / "phases"
OUT = Path(__file__).with_name("_c1_phase_details_18_125.md")

# Short titles already present in ROADMAP checklist (override PLAN title when useful)
CHECKLIST_TITLES: dict[int, str] = {
    18: "AI Privacy (v3.5)",
    19: "AI Polish",
    20: "Milestone Gate",
    21: "Architecture Completion",
    22: "Architecture storage",
    23: "Architecture apigateway",
    24: "Architecture imex",
    25: "Architecture dashboard",
    26: "Architecture rbac",
    27: "Architecture auth",
    28: "v2.6 Architecture Gate",
    29: "OBS-01",
    30: "OBS-01 FE",
    31: "OBS-01 Prometheus",
    32: "OBS-01 OTel",
    33: "Monitoring stack",
    34: "Loki + alerts",
    35: "Alertmanager",
    36: "Alert webhook",
    37: "Biome gate",
    38: "v2.6 Milestone Gate",
    39: "Intranet News",
    40: "Org directory",
    41: "Unified calendar",
    42: "CRM-lite",
    43: "HR-lite",
    44: "Kudos",
    45: "Marketplace",
    46: "Status page",
    47: "Random Coffee",
    48: "1C Connector",
    49: "Lark Base",
    50: "v4 Milestone Gate",
    51: "Accounting MVP",
    52: "Accounting v4.5",
    53: "v4 Module Gate",
    54: "v4 Coverage Uplift",
    55: "onec + intranet",
    56: "v4 Final Gate",
    57: "Post-v4 CI Handoff",
    58: "v4 Milestone Archive",
    59: "Coverage + UX Close-out",
    60: "UX-1 Gate",
    61: "Coverage + Context Tooltips",
    62: "Onboarding Profile RHF",
    63: "Production Release Gate",
    64: "Auth Coverage Uplift",
}


def slug_title(slug: str) -> str:
    return slug.replace("-", " ").strip().title()


def first_nonempty_lines(text: str, n: int = 40) -> list[str]:
    return [ln.rstrip() for ln in text.splitlines()]


def extract_h1_title(text: str, num: int, slug: str) -> str:
    m = re.search(r"^#\s+Phase\s+\d+\s*[—:\-–]\s*(.+)$", text, re.M)
    if m:
        return m.group(1).strip()
    if num in CHECKLIST_TITLES:
        return CHECKLIST_TITLES[num]
    return slug_title(slug)


def extract_goal(text: str) -> str:
    # **Goal:** / **Goal**: inline
    m = re.search(r"\*\*Goal:\*\*\s*(.+)", text)
    if m:
        g = m.group(1).strip()
        if g:
            return g
    # ## Goal section
    m = re.search(r"^##\s+Goal\s*$", text, re.M)
    if m:
        rest = text[m.end() :]
        # stop at next ##
        stop = re.search(r"^##\s+", rest, re.M)
        block = rest[: stop.start()] if stop else rest
        block = block.strip()
        # Prefer first paragraph / first list items condensed
        lines = []
        for ln in block.splitlines():
            s = ln.strip()
            if not s:
                if lines:
                    break
                continue
            if s.startswith("|") or s.startswith("```"):
                # table goal — take header narrative from before or summarize rows
                break
            if s.startswith("#"):
                break
            lines.append(re.sub(r"^[-*]\s+\[[ xX]\]\s*", "", s))
            if len(lines) >= 3:
                break
        if lines:
            return " ".join(lines)
        # table under Goal
        rows = re.findall(r"^\|\s*([^|]+?)\s*\|\s*([^|]+?)\s*\|\s*([^|]+?)\s*\|", block, re.M)
        if rows:
            # skip header/sep
            bits = []
            for a, b, c in rows:
                if set(a.strip()) <= set("-:") or a.strip().lower() in {
                    "metric",
                    "пакет",
                    "package",
                    "area",
                }:
                    continue
                bits.append(f"{a.strip()}: {b.strip()} → {c.strip()}")
                if len(bits) >= 4:
                    break
            if bits:
                return "; ".join(bits)
    # ## Exit criteria
    m = re.search(r"^##\s+Exit criteria\s*$", text, re.M | re.I)
    if m:
        rest = text[m.end() :].lstrip()
        first = rest.splitlines()[0].strip() if rest.splitlines() else ""
        if first and not first.startswith("#"):
            return first
    # Status COMPLETE line as fallback
    m = re.search(r"\*\*Status:\*\*\s*(.+)", text)
    if m:
        return f"(see PLAN.md) Status: {m.group(1).strip()}"
    return "(see PLAN.md for goal)"


def extract_depends(text: str, num: int) -> str:
    m = re.search(r"\*\*Depends on:\*\*\s*(.+)", text, re.I)
    if m:
        return m.group(1).strip().strip("`")
    m = re.search(r"Depends on:\*\*\s*(.+)", text, re.I)
    if m:
        return m.group(1).strip()
    m = re.search(r"Depends on[:\*]*\s*(Phase[^\n·|]+)", text, re.I)
    if m:
        return m.group(1).strip()
    if num > 18:
        return f"Phase {num - 1}"
    return f"Phase {num - 1}"


def extract_requirements(text: str) -> str:
    m = re.search(r"\*\*Requirements:\*\*\s*(.+)", text, re.I)
    if m:
        return m.group(1).strip()
    # Collect REQ-like tokens
    ids = re.findall(
        r"\b(?:M3-\d+|MOD-\d+|V4-\d+|UX-P?\d+(?:-\d+)?|FB-\d+|ARCH-\d+|OBS-\d+|PG-\d+|COV-\d+|SEC-\d+|NOTIF|MAIL-\d+)\b",
        text,
    )
    # unique preserve order
    seen = set()
    out = []
    for i in ids:
        if i not in seen:
            seen.add(i)
            out.append(i)
        if len(out) >= 12:
            break
    if out:
        return ", ".join(out)
    return "see PLAN.md"


def extract_success(text: str) -> list[str]:
    items: list[str] = []
    # ## Success criteria / Success Criteria
    m = re.search(r"^##\s+Success [Cc]riteria\s*$", text, re.M)
    if m:
        rest = text[m.end() :]
        stop = re.search(r"^##\s+", rest, re.M)
        block = rest[: stop.start()] if stop else rest
        for ln in block.splitlines():
            s = ln.strip()
            mm = re.match(r"^(?:[-*]\s+\[[ xX]\]\s*|\d+\.\s+)(.+)$", s)
            if mm:
                items.append(mm.group(1).strip())
            elif s.startswith("- "):
                items.append(s[2:].strip())
        if items:
            return items[:6]

    # ## Tasks checkboxes
    m = re.search(r"^##\s+Tasks\s*$", text, re.M)
    if m:
        rest = text[m.end() :]
        stop = re.search(r"^##\s+", rest, re.M)
        block = rest[: stop.start()] if stop else rest
        for ln in block.splitlines():
            mm = re.match(r"^[-*]\s+\[[ xX]\]\s*(.+)$", ln.strip())
            if mm:
                items.append(mm.group(1).strip())
        if items:
            return items[:6]

    # Wave tables Done column
    waves = re.findall(
        r"^##\s+Wave\s+\d+[^\n]*\n+(?:\|[^\n]+\n)+",
        text,
        re.M,
    )
    for w in waves:
        for row in w.splitlines():
            if not row.startswith("|"):
                continue
            cols = [c.strip() for c in row.strip("|").split("|")]
            if len(cols) >= 1 and cols[0].startswith("W"):
                items.append(cols[0] if len(cols) == 1 else cols[0])
                # Prefer task description column
                if len(cols) >= 2 and not cols[0].lower().startswith("task"):
                    # first col may be task id+desc combined
                    pass
                if len(cols) >= 2:
                    desc = cols[0]
                    if re.match(r"^W\d", desc):
                        # table: Task | Agent | ...
                        items[-1] = desc
        if len(items) >= 6:
            break
    if items:
        # clean duplicates
        uniq = []
        seen = set()
        for i in items:
            if i.lower() in {"task", "---", "wave"}:
                continue
            if i not in seen:
                seen.add(i)
                uniq.append(i)
        return uniq[:6]

    # ## Results table
    m = re.search(r"^##\s+Results\s*$", text, re.M)
    if m:
        rest = text[m.end() :]
        stop = re.search(r"^##\s+", rest, re.M)
        block = rest[: stop.start()] if stop else rest
        for row in block.splitlines():
            if not row.startswith("|"):
                continue
            cols = [c.strip() for c in row.strip("|").split("|")]
            if len(cols) >= 3 and set(cols[0]) <= set("-:"):
                continue
            if len(cols) >= 3 and cols[0].lower() not in {"metric", "area", "пакет"}:
                items.append(f"{cols[0]}: {cols[1]} → {cols[2]}")
        if items:
            return items[:6]

    # Status COMPLETE
    m = re.search(r"\*\*Status:\*\*\s*([^\n]+)", text)
    if m:
        return [m.group(1).strip()]
    m = re.search(r"^##\s+Status:\s*(.+)$", text, re.M)
    if m:
        return [m.group(1).strip()]
    return ["See PLAN.md exit criteria / tasks"]


def extract_plans(text: str, num: int, slug: str) -> list[str]:
    plans: list[str] = []
    # numbered PLAN files referenced
    refs = re.findall(rf"{num:02d}-\d{{2}}(?:-PLAN\.md)?", text)
    for r in refs:
        if r not in plans:
            plans.append(r if r.endswith(".md") else f"{r}-PLAN.md")

    # Wave headers as plan units
    for m in re.finditer(r"^##\s+(Wave\s+\d+[^\n]*)", text, re.M):
        plans.append(m.group(1).strip())

    # checkbox tasks under Tasks
    if not plans:
        for ln in text.splitlines():
            mm = re.match(r"^[-*]\s+\[[ xX]\]\s*(.+)$", ln.strip())
            if mm:
                plans.append(mm.group(1).strip())
            if len(plans) >= 8:
                break

    # Deliverables bullets
    if not plans:
        m = re.search(r"^##\s+Deliverables\s*$", text, re.M)
        if m:
            rest = text[m.end() :]
            stop = re.search(r"^##\s+", rest, re.M)
            block = rest[: stop.start()] if stop else rest
            for ln in block.splitlines():
                s = ln.strip()
                if s.startswith("- "):
                    plans.append(s[2:].strip())
                elif s.startswith("### "):
                    plans.append(s[4:].strip())

    if not plans:
        plans = [f"`.planning/phases/{num:02d}-{slug}/PLAN.md` (single plan)"]
    return plans[:10]


def extract_completed_date(text: str) -> str:
    m = re.search(
        r"(?:COMPLETE|completed|Status:.*?COMPLETE).*?(20\d{2}-\d{2}-\d{2})",
        text,
        re.I | re.S,
    )
    if m:
        return m.group(1)
    m = re.search(r"\*\*Date:\*\*\s*(20\d{2}-\d{2}-\d{2})", text)
    if m:
        return m.group(1)
    m = re.search(r"(20\d{2}-\d{2}-\d{2})", text)
    if m:
        return m.group(1)
    return ""


def extract_blurb_for_checklist(goal: str, text: str) -> str:
    # Prefer one-line goal trimmed
    g = re.sub(r"\s+", " ", goal).strip()
    if len(g) > 120:
        g = g[:117] + "…"
    date = extract_completed_date(text)
    if date:
        return f"{g} ({date})"
    return g


def render_phase(num: int, d: Path) -> tuple[str, str, str]:
    """Returns (details_md, checklist_line, progress_row)."""
    slug = d.name.split("-", 1)[1]
    plan_path = d / "PLAN.md"
    text = plan_path.read_text(encoding="utf-8") if plan_path.exists() else ""
    title = CHECKLIST_TITLES.get(num) or extract_h1_title(text, num, slug)
    # Prefer shorter checklist title if H1 is very long
    h1 = extract_h1_title(text, num, slug)
    if num not in CHECKLIST_TITLES:
        title = h1
    goal = extract_goal(text)
    depends = extract_depends(text, num)
    reqs = extract_requirements(text)
    success = extract_success(text)
    plans = extract_plans(text, num, slug)
    date = extract_completed_date(text)
    rel = f".planning/phases/{d.name}/PLAN.md"

    sc_lines = "\n".join(f"  {i}. {s}" for i, s in enumerate(success, 1))
    plan_lines = "\n".join(f"- [x] {p}" for p in plans)

    details = f"""### Phase {num}: {title}
**Goal**: {goal}
**Depends on**: {depends}
**Requirements**: {reqs}
**Success Criteria** (what must be TRUE):
{sc_lines}
**Plans**: {len(plans)} plan(s) · source `{rel}`

Plans:
{plan_lines}
"""

    blurb = extract_blurb_for_checklist(goal, text)
    checklist = f"- [x] **Phase {num}: {title}** - {blurb}"
    status = "Complete"
    completed = date or "—"
    progress = f"| {num}. {title} | — | see PLAN | {status} | {completed} |"
    return details, checklist, progress


def main() -> None:
    dirs: dict[int, Path] = {}
    for d in PHASES.iterdir():
        if not d.is_dir():
            continue
        m = re.match(r"^(\d+)-(.+)$", d.name)
        if not m:
            continue
        n = int(m.group(1))
        if 18 <= n <= 125:
            dirs[n] = d

    details_parts: list[str] = []
    checklist_65: list[str] = []
    checklist_39: str | None = None
    progress_rows: list[str] = []

    header = """
> **Restored 2026-07-27 (planning audit C1):** Phase Details for **18–125** generated from each phase `PLAN.md` under `.planning/phases/`. Phases **01–13** live under `.planning/milestones/` (`v1.0-phases/` + `v2.0-phases/`) — their details above remain the historical ROADMAP copy. Do not treat this block as a substitute for PLAN/VERIFICATION artifacts.

"""
    details_parts.append(header)

    for n in range(18, 126):
        if n not in dirs:
            details_parts.append(
                f"### Phase {n}: MISSING\n**Goal**: phase directory not found under `.planning/phases/`\n"
            )
            continue
        details, checklist, progress = render_phase(n, dirs[n])
        details_parts.append(details)
        progress_rows.append(progress)
        if n == 39:
            checklist_39 = checklist
        if n >= 65:
            checklist_65.append(checklist)

    OUT.write_text("\n".join(details_parts).rstrip() + "\n", encoding="utf-8")
    (OUT.parent / "_c1_checklist_65_125.md").write_text(
        "\n".join(checklist_65) + "\n", encoding="utf-8"
    )
    (OUT.parent / "_c1_checklist_39.md").write_text(
        (checklist_39 or "") + "\n", encoding="utf-8"
    )
    (OUT.parent / "_c1_progress_18_125.md").write_text(
        "\n".join(progress_rows) + "\n", encoding="utf-8"
    )
    print(f"Wrote {OUT} ({len(dirs)} phases)")
    print(f"checklist 65-125: {len(checklist_65)} lines")


if __name__ == "__main__":
    main()
