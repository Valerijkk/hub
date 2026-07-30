# MUTATION W5 slot 5 — TASK-125 (fe2) · 2026-07-30

> Checker: **fe2** (!= fe1 criteria author) · Env: local Vitest · Criteria: `.planning/backlog/TASK-125-2026-07-29-a3-10-2-fe-routes-modules.md` `[MUTATION CRITERIA]`

## Verdict

**5/5 PROVEN** · статус TASK-125 → **Закрыта**

| ID | Verdict | RED test (`it`) | Mutation (BREAK) | Layer |
|----|---------|-----------------|------------------|-------|
| A3W10-2-001 | **PROVEN** | `blocks javascript: and data: schemes → #` (+ RTL github/gitlab sanitize) | `sanitizeUrl` всегда возвращает входную строку | frontend |
| A3W10-2-002 | **PROVEN** | `hides «+ Создать ветку» for viewer (canMutate=false)`; `returns null for viewer (canMutate=false) — no trigger CTA` | убрать `canMutate &&` у CTA / early `if (!canMutate) return null` | frontend |
| A3W10-2-003 | **PROVEN** | `toast «Открыть» calls window.open with noopener; unsafe URL skips open` | `window.open(data.web_url, "_blank")` без sanitize / noopener / skip `#` | frontend |
| A3W10-2-004 | **PROVEN** | `import-export.tsx module is deleted (dead export cleanup)` | восстановить `import-export.tsx` (`ImportExportToolbar`) | frontend |
| A3W10-2-005 | **PROVEN** | `shows loading UI while fetch is in flight, not empty copy`; `sets loading true then false in store around fetchBurndown` | убрать `set({ loading: true…})`; `if (false && loading…)` в chart | frontend |

Production restored (`git checkout --` / delete resurrected file). Mock-gap: нет (unit/RTL на prod-коде + existsSync).

## Per-ID evidence

### A3W10-2-001 — sanitizeUrl whitelist

```text
BREAK: safe-url.ts — sanitizeUrl returns raw string
→ FAIL blocks javascript: and data: schemes → #
→ FAIL sanitizes unsafe PR/commit href to #
→ FAIL sanitizes unsafe pipeline web_url to #
RESTORE
→ PASS blocks javascript: and data: schemes → #
```

### A3W10-2-002 — canMutate Trigger CI / Create branch

```text
BREAK: github-links.tsx — drop canMutate && around «+ Создать ветку»
BREAK: trigger-ci-dialog.tsx — drop if (!canMutate) return null
→ FAIL hides «+ Создать ветку» for viewer (canMutate=false)
→ FAIL returns null for viewer (canMutate=false) — no trigger CTA
RESTORE
→ PASS both
```

### A3W10-2-003 — window.open noopener + skip unsafe

```text
BREAK: trigger-ci-dialog.tsx — window.open(data.web_url, "_blank") raw
→ FAIL toast «Открыть»… (expected StringContaining "noopener")
RESTORE
→ PASS same test
```

### A3W10-2-004 — ImportExportToolbar deleted

```text
BREAK: create frontend/src/modules/tracker/import-export.tsx
→ FAIL import-export.tsx module is deleted (dead export cleanup)
RESTORE: Remove-Item import-export.tsx
→ PASS same test (existsSync=false)
```

### A3W10-2-005 — burndown loading≠empty

```text
BREAK: sprint-store.ts — drop set({ loading: true, error: null })
BREAK: burndown-chart.tsx — if (false && loading && …)
→ FAIL shows loading UI while fetch is in flight, not empty copy
→ FAIL sets loading true then false in store around fetchBurndown
RESTORE
→ PASS both
```

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs
npx vitest run \
  ../testframework/stages/unit/frontend-test/domains/tracker/safe-url.test.ts \
  ../testframework/stages/unit/frontend-test/domains/tracker/github-links-sanitize-authz.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tracker/gitlab-pipeline-sanitize.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tms/trigger-ci-authz-noopener.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tracker/import-export-deleted.test.ts \
  ../testframework/stages/unit/frontend-test/domains/tracker/burndown-chart-loading.test.tsx
→ Test Files 6 passed · Tests 14 passed (~3.4s) after full restore
```

## Assumptions / unresolved

- Коммит не создавался (инструкция слота).
- Criteria authored by fe1; mutation prove by fe2 (W5 slot 5).

## Handoff

TASK-125 **Закрыта** · evidence этот файл · README синхронизирован.
