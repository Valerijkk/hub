# MUTATION W13 slot 2 refill — TASK-131 (fe1) · 2026-07-30

> Checker: **fe1** · Env: local Vitest · Criteria: `.planning/backlog/TASK-131-2026-07-29-a3-11-3-fe-stores-guards.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**2/2 PROVEN** · TASK-131 → **Закрыта**

| ID | Verdict | RED |
|----|---------|-----|
| A3W11-3-001 | **PROVEN** | `sets error and rejects when getLayout fails` |
| A3W11-3-002 | **PROVEN** | `skips update when reactions is undefined or non-array` |

Production restored (`git checkout --`). Leftover MUTATED: **нет**. Mock-gap: нет (store unit + RTL hook на prod `useMessages`).

`npx tsc --noEmit` — N/A (не primary per criteria; не ловит store/WS guards).

## Per-ID evidence

### A3W11-3-001 — `loadLayout` error + rethrow

```text
BREAK: dashboard-store.ts loadLayout catch — silent catch
  (убрать set({ error: "Не удалось загрузить раскладку" }) + throw err)
→ FAIL sets error and rejects when getLayout fails
     AssertionError: promise resolved "undefined" instead of rejecting
RESTORE → PASS 2/2 (layout-load-error.test.ts)
```

Guard: `frontend/src/stores/dashboard-store.ts:107-111`  
Vitest: `domains/dashboard/layout-load-error.test.ts` · describe `A3W11-3-001 · dashboard loadLayout error≠empty`

### A3W11-3-002 — `reaction_toggled` Array.isArray

```text
BREAK: use-messages.ts — писать reactions as-is без Array.isArray
  (if (!msgId) break; вместо !Array.isArray(reactEvt.reactions))
→ FAIL skips update when reactions is undefined or non-array
     expected { emoji: '🔥' } to deeply equal [ Array(1) ]
  (valid-array / empty-array its остались PASS)
RESTORE → PASS 3/3 (reaction-toggled-guard.test.tsx)
```

Guard: `frontend/src/hooks/use-messages.ts:132`  
Vitest: `domains/messenger/reaction-toggled-guard.test.tsx` · describe `A3W11-3-002 · reaction_toggled reactions guard`

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs   # if needed

# RED (после BREAK) / GREEN (после RESTORE)
npx vitest run ../testframework/stages/unit/frontend-test/domains/dashboard/layout-load-error.test.ts
npx vitest run ../testframework/stages/unit/frontend-test/domains/messenger/reaction-toggled-guard.test.tsx
```

GREEN (оба файла после restore): Test Files 2 passed · Tests 5 passed · ~1.5s

## Gaps (criteria)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| loadLayout error+rethrow | ✅ unit store | PROVEN |
| reaction_toggled Array.isArray | ✅ RTL hook + Query cache | PROVEN |
| `npx tsc --noEmit` | N/A | не primary |

**Confidence:** CONFIRMED — 2/2 runnable PROVEN.

## Handoff

- Evidence: этот файл · TASK-131 **Закрыта** · README sync  
- fe-ver → code-reviewer (fix code already FIXED; mutation PROVEN)  
- Коммит: не создавался
