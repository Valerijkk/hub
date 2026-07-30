# MUTATION W13 slot 5 refill — TASK-187 (fe2) · 2026-07-30

> Checker: **fe2** · Env: local Vitest · Criteria: `.planning/backlog/archive/2026-07-29/TASK-187-2026-07-29-audit-fe-pipeline-sanitize.md` `[MUTATION CRITERIA]`  
> Prior: `MUTATION-2026-07-30-W11-fe1.md` (001–004 PROVEN; 005 N/A `test_does_not_catch`)  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**5/5 PROVEN** · FE-TA-005 FIXED (hub tokens + Vitest source-guard) · TASK-187 → **Закрыта**

| ID | Verdict | RED test (`it`) | Mutation (BREAK) | Layer |
|----|---------|-----------------|------------------|-------|
| FE-TA-001 | **PROVEN** | `mutation: javascript: html_url → нет href…` (+ source guard) | drop `isSafeUrl`; `href={run.html_url}` raw | frontend |
| FE-TA-002 | **PROVEN** | `mutation: javascript: att.url → нет <a href>` | `if (false && !isSafeUrl)`; `href={candidate}` | frontend |
| FE-TA-003 | **PROVEN** | `network fail → alert+retry, not «сессия не найдена»` | `loadFailed \|\| !pokerSession` → empty copy | frontend |
| FE-TA-004 | **PROVEN** | `statuses fail → alert…`; `channels fail → alert…` | force `statusesError`/`channelsError = false`; remove secondary alerts | frontend |
| FE-TA-005 | **PROVEN** | `source guard: hub tokens — нет text-green/red-600 / bg-gray-*` | restore `text-green-600`/`text-red-600`/`bg-gray-200` | frontend |

Production restored (только intentional fix в `pipeline-status.tsx` + новый `it`). Leftover MUTATED: **нет**.

## Fix (FE-TA-005 — was GAP)

1. `frontend/src/modules/tracker/pipeline-status.tsx` — raw Tailwind (`text-green/red-600`, `bg-gray-*`, `text-gray-*`, yellow) → hub tokens (`text-[var(--green|red|orange|gray)]`, `bg-[var(--gG|gR)]`, `bg-raised`, `text-t*`, `hover:bg-[var(--hover)]`).
2. Vitest source-guard в `pipeline-status-sanitize.test.tsx`:
   - `it("source guard: hub tokens — нет text-green/red-600 / bg-gray-* (FE-TA-005)")`

## Per-ID evidence

### FE-TA-001 — pipeline sanitizeUrl — PROVEN

```text
BASELINE GREEN → PASS 4/4 pipeline-status-sanitize
BREAK: drop isSafeUrl gate; href={run.html_url}
→ FAIL mutation: javascript: html_url → нет href
→ FAIL source guard: isSafeUrl(run.html_url) missing
RESTORE (from FIXED backup, not git HEAD) → PASS
```

### FE-TA-002 — attachments sanitizeUrl — PROVEN

```text
BREAK: if (false && !isSafeUrl); href={candidate}
→ FAIL mutation: javascript: att.url → нет <a href>
RESTORE → PASS 2/2
```

### FE-TA-003 — planning-poker error≠empty — PROVEN

```text
BREAK: loadFailed || !pokerSession → «Сессия не найдена…»
→ FAIL network fail → alert+retry… (Unable to find role="alert")
RESTORE → PASS 2/2
```

### FE-TA-004 — automation secondary error≠empty — PROVEN

```text
BREAK: statusesError/channelsError = false; remove secondary QueryErrorAlert
→ FAIL statuses fail → alert… (2/2)
→ FAIL channels fail → alert…
RESTORE → PASS 2/2
```

### FE-TA-005 — hub-tokens pipeline-status — PROVEN

```text
BASELINE: rg "text-(green|red)-600|bg-gray-" pipeline-status.tsx → EMPTY
BREAK: restore text-green-600 / text-red-600 / bg-gray-200
→ FAIL source guard: hub tokens — нет text-green/red-600 / bg-gray-*
RESTORE → PASS
```

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs
npx vitest run \
  ../testframework/stages/unit/frontend-test/domains/tracker/pipeline-status-sanitize.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tracker/issue-attachments-sanitize.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tracker/planning-poker-error-empty.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tracker/automation-rules-secondary-error.test.tsx
→ Final GREEN: Test Files 4 passed · Tests 10 passed (~1.9s)
```

## Gaps (non-blocking)

| Guard | Ловит тест? | Примечание |
|-------|-------------|-----------|
| pipeline FE-TA-005 tokens | ✅ source-guard | primary criterion |
| Residual audit cluster (`gitlab-pipeline-status`, `github-links`, `triage-inbox`, `issue-header`, `issue-sidebar-meta`, `#2563EB` poker) | ❌ | hygiene leftover вне kill-recipe `pipeline-status.tsx`; не FE-TA-005 mutation target |
| `npx tsc --noEmit` | N/A | criteria: не ловит 001–005 |

**Confidence:** HIGH — 5/5 runnable PROVEN after 005 fix + Vitest.

## Handoff

Evidence: этот файл · TASK-187 → **Закрыта** · UI FIXED (001–005) · fe-ver может принимать.
