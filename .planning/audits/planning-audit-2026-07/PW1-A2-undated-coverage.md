# PW1-A2 — Undated coverage claims → date-stamped

**Date:** 2026-07-27  
**Scope:** documentation only (no production code).  
**Rule:** every measured coverage % must carry a measurement date in the same sentence / paragraph / table cell (or adjacent cell).  
**Stamp used:** **2026-06-22 Docker baseline** from `.planning/COVERAGE.md` (until A1 refreshes the gate run). No new coverage run invented.

---

## Method

Searched `AGENTS.md`, `README.md`, `docs/**`, `.planning/**`, `.cursor/rules/**`, `.cursor/**/*.md` for measured statement-coverage percentages (overall and package).

**Excluded (not coverage measurements):**
- Gate/target thresholds (`≥82%`, `≥80%`, `≥50%`) without a stated current fact
- Inventory counts (499 test files, Vitest 99/99, E2E 20 specs)
- Prompt/skill pedagogy that discusses “80% coverage” as a concept
- Historical phase-exit numbers already dated in an adjacent cell (kept; clarified when they differ from the 83.0% baseline)

---

## Hits fixed (stamp = 2026-06-22 Docker baseline unless noted)

| File | What was undated | Can add date? |
|------|------------------|---------------|
| `AGENTS.md` | COV 83.0%; package rows | Yes → baseline |
| `README.md` | badge 83%; table/list 83%/83.0% | Yes → baseline (badge label+alt) |
| `.cursor/rules/00-project.mdc` | COV 83.0% ×2 | Yes |
| `.cursor/rules/40-testing.mdc` | security package %; wrong “факт на 2026-07” | Yes; corrected to 2026-06-22 |
| `.cursor/playbooks/testing.md` | 83.0% floor + table | Yes |
| `.cursor/playbooks/team-workflow.md` | COV/security row; tracker 82.5% typo→82.4% | Yes |
| `.cursor/playbooks/analysis.md` | example matrix 82.0%/82.4% | Yes |
| `.cursor/agents/product.md` | Coverage 83.0% | Yes |
| `.cursor/agents/docs-writer.md` | example package row | Yes |
| `.cursor/REGULATION.md` | COV cell (column had date; cell clarified) | Yes |
| `.planning/STATE.md` | overall + package table | Yes |
| `.planning/COVERAGE.md` | overall/security sections + verdict | Yes (section already dated; cells reinforced) |
| `.planning/decisions/gray-zones/AREA-quality.md` | COV 83.0%; GZ-QA-03 package % | Yes |
| `.planning/archive/july-2026/INDEX-JULY-2026.md` | baseline 83.0% | Yes |
| `.planning/archive/july-2026/PROJECT-STATUS-2026-07-10.md` | COV rows | Yes |
| `.planning/archive/july-2026/EXPANSION-ANALYSIS-2026-07-02.md` | per-module cov + overall 83.0% | Yes |
| `.planning/audits/code-review-final/SUMMARY.md` | banner “Актуально: COV 83%” | Yes |
| `.planning/ops/production/PR-feat-v2.5-trust.md` | overall + package table | Yes |
| `.planning/archive/tasks/TASKS-PHASES-121-FINAL-AUDIT.md` | COV overall cell (section already dated) | Yes |
| `.planning/audits/2026-06/AUDIT-2026-06-23.md` | 83.0% / ws 87.7% | Yes |
| `.planning/audits/2026-06/AUDIT-2026-06-22-quality-processes.md` | 83.0% / 82.6% | Yes (82.6% labeled phase-125 exit 2026-06-21) |
| `.planning/audits/2026-06/WAVE-C-VERIFICATION.md` | COV 83.0% | Yes |
| `.planning/audits/2026-07/repo-scan/W6-tests-coverage.md` | STALE 83.0% rows | Yes |
| `.planning/audits/grok-2026-07/00-SUMMARY.md` | COV 83% baseline | Yes |
| `.planning/audits/grok-2026-07/15-requirements-vs-code.md` | STATE 83.0% | Yes |
| `.planning/archive/history/STATE-HISTORY-THROUGH-2026-07-23.md` | FA-03 COV 83.0% | Yes |
| `.planning/archive/ab-tests/PLAN-bases-opus.md` | bases 78.6% | Yes |
| `.planning/audits/planning-audit-2026-07/00-PLANNING-AUDIT.md` | cited 83.0%/package % | Yes |
| `docs/TESTING.md` | principles + counters 83.0% | Yes |
| `docs/test_plan.md` | statement + module matrix | Yes |
| `docs/competitive_audit.md` | 83.0% / 82.0% FA-11 | Yes |
| `docs/competitive_analysis_2026.md` | backend 83.0% | Yes |
| `docs/vision_and_scope.md` | KPI 83.0% | Yes |
| `docs/user_needs_analysis_2026.md` | 83.0%; 82.6% retain trigger | Yes (82.6% → 2026-06-21 exit) |
| `docs/testframework_strategy.md` | 83.0% ×2 | Yes |
| `docs/system_specification.md` | NFR-29 + gaps + footer | Yes |
| `docs/BUSINESS.md` | Coverage wave 82.6% | Yes (phase exit + pointer to 83.0% baseline) |

---

## Left as-is (with rationale)

| Location | Why not stamped as 2026-06-22 baseline |
|----------|----------------------------------------|
| `.planning/phases/**` PLAN exit metrics (57.1%, 82.6% phase-125, etc.) | Historical phase results; not presented as live master baseline |
| `.planning/ops/production/PRODUCTION-HANDOFF.md` “Quality metrics (verified 2026-06-21)” with 64.5% | Older dated snapshot (different run) |
| `.planning/archive/tasks/TASKS-CLOSE-GAPS.md` rows with own dates (e.g. calls 80.2% on 2026-06-20) | Already dated nearby |
| `.planning/audits/grok-2026-07/12-tests.md` | Already says `CI baseline 2026-06-22` adjacent |
| `docs/TECHNICAL.md` / `docs/stack_and_tools.md` `≥82%` | Gate floor, not measured current % |
| `pro-qa-ai/**` | Conceptual “80% coverage”, not Hub metrics |
| File-count / Vitest / E2E counts | Inventory, not statement coverage |

---

## Touched files

### Already on `master` via prior commit `1538470` (A6 inventory refresh, same session)

These received the 2026-06-22 Docker baseline stamp before this commit:

```
AGENTS.md
README.md
.cursor/rules/00-project.mdc
.cursor/rules/40-testing.mdc
.cursor/playbooks/team-workflow.md
.planning/COVERAGE.md
.planning/STATE.md
docs/TESTING.md
docs/test_plan.md
docs/testframework_strategy.md
```

### This commit (remaining date-stamps + notes)

```
.cursor/REGULATION.md
.cursor/agents/docs-writer.md
.cursor/agents/product.md
.cursor/playbooks/analysis.md
.cursor/playbooks/testing.md
.planning/archive/ab-tests/PLAN-bases-opus.md
.planning/audits/2026-06/AUDIT-2026-06-22-quality-processes.md
.planning/audits/2026-06/AUDIT-2026-06-23.md
.planning/audits/2026-06/WAVE-C-VERIFICATION.md
.planning/audits/2026-07/repo-scan/W6-tests-coverage.md
.planning/audits/grok-2026-07/00-SUMMARY.md
.planning/audits/grok-2026-07/15-requirements-vs-code.md
.planning/audits/code-review-final/SUMMARY.md
.planning/decisions/gray-zones/AREA-quality.md
.planning/archive/history/STATE-HISTORY-THROUGH-2026-07-23.md
.planning/archive/july-2026/EXPANSION-ANALYSIS-2026-07-02.md
.planning/archive/july-2026/INDEX-JULY-2026.md
.planning/archive/july-2026/PROJECT-STATUS-2026-07-10.md
.planning/audits/planning-audit-2026-07/PW1-A2-undated-coverage.md
.planning/ops/production/PR-feat-v2.5-trust.md
.planning/archive/tasks/TASKS-PHASES-121-FINAL-AUDIT.md
docs/BUSINESS.md
docs/competitive_analysis_2026.md
docs/competitive_audit.md
docs/system_specification.md
docs/user_needs_analysis_2026.md
docs/vision_and_scope.md
```

Note: small date clarifications were also applied in untracked
`.planning/audits/planning-audit-2026-07/00-PLANNING-AUDIT.md` (not added here — owned by the audit pack).

---

## Follow-up

- **A1 / coverage refresh:** re-run Docker `-tags hubtest` coverage gate; replace “2026-06-22 Docker baseline” labels with the new measurement date everywhere this commit touched.
- Rule of thumb for authors: bare `83.0%` / package `%` without a date = invalid status claim.
