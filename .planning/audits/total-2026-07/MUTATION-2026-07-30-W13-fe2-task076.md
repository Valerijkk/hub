# MUTATION W13 slot 5 refill — TASK-076 (fe2) · 2026-07-30

> Checker: **fe2** · Env: local Vitest · Criteria: `.planning/backlog/archive/legacy/TASK-076-fe-p3-perf-stale.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**3/3 PROVEN** · TASK-076 → **Закрыта**

| ID | Verdict | RED |
|----|---------|-----|
| W10-A3-017 | **PROVEN** | `pipeline refetchInterval only when queued/in_progress` |
| W10-A3-020 | **PROVEN** | `dashboard clock updates via interval, not frozen useMemo` |
| W10-A5-019 | **PROVEN** | `run add-cases dialog hoists existingIds Set outside map` |

Production guards restored (mutation targets only). Leftover MUTATED: **нет**.  
Note: `pipeline-status.tsx` имеет pre-existing WIP diff FE-TA-005 (hub tokens) вне kill-recipe TASK-076 — не трогали.

## Per-ID evidence

### W10-A3-017 — pipeline refetchInterval conditional — PROVEN

```text
BASELINE GREEN → PASS 3/3 perf-stale-p3
BREAK: refetchInterval: (query) => … → refetchInterval: 30_000
→ FAIL pipeline refetchInterval only when queued/in_progress
     expected … to match /refetchInterval:\s*\(query\)/
RESTORE → PASS
```

Guard: `frontend/src/modules/tracker/pipeline-status.tsx:202-208`  
Vitest: `domains/platform/perf-stale-p3.test.ts` · `it("pipeline refetchInterval only when queued/in_progress")`

### W10-A3-020 — dashboard clock interval — PROVEN

```text
BREAK: useState(formatTime)+setInterval → useMemo(() => formatTime(), [])
→ FAIL dashboard clock updates via interval, not frozen useMemo
     expected … to match /setCurrentTime\(formatTime\(\)\)/
RESTORE → PASS
```

Guard: `frontend/src/routes/_authenticated/dashboard.lazy.tsx:70-73`  
Vitest: same file · `it("dashboard clock updates via interval, not frozen useMemo")`

### W10-A5-019 — existingIds Set hoist — PROVEN

```text
BREAK: new Set(…) moved inside .map callback
→ FAIL run add-cases dialog hoists existingIds Set outside map
     beforeMap: expected … to match /const existingIds = new Set/
RESTORE → PASS
```

Guard: `frontend/src/routes/_authenticated/projects/$projectId/tms/runs_.$runId.tsx:313-317`  
Vitest: same file · `it("run add-cases dialog hoists existingIds Set outside map")`

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs
npx vitest run ../testframework/stages/unit/frontend-test/domains/platform/perf-stale-p3.test.ts
→ Baseline GREEN: Test Files 1 passed · Tests 3 passed (~1.0s)
→ Per-ID RED: 3/3 FAIL as above
→ Final GREEN: Test Files 1 passed · Tests 3 passed (~1.1s)
```

## Gaps

| Guard | Ловит тест? | Примечание |
|-------|-------------|-----------|
| pipeline / clock / Set hoist | ✅ source-guard | **3/3** Class source IDs |
| `npx tsc --noEmit` | N/A | criteria: не primary |

**Confidence:** HIGH — 3/3 runnable PROVEN.

## Handoff

Evidence: этот файл · TASK-076 → **Закрыта** · UI mutation targets restored · fe-ver может принимать.
