# MUTATION W8 slot 5 — TASK-182 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** (!= fe2 fixer) · Env: local Vitest  
> Criteria: `.planning/backlog/TASK-182-2026-07-29-c-w19-hr-pending-leave-error.md` `[MUTATION CRITERIA]`  
> Evidence JSON: `_mut-batch-w8-task182.json`

## Verdict

**1/1 PROVEN** · статус TASK-182 → **Закрыта**

| ID | Verdict | RED test (`it`) | Mutation (BREAK) | Layer |
|----|---------|-----------------|------------------|-------|
| C-W19-FE-005 | **PROVEN** | `shows approval queue alert when listPendingLeave fails, not silent empty` | `leave-requests.tsx` — gate → `pendingItems.length > 0` only; remove `pendingError ? <QueryErrorAlert>` | frontend |

Production restored (`git checkout -- frontend/src/modules/hr/leave-requests.tsx`). Mock-gap: нет (RTL на prod-компоненте + mocked `listPendingLeave` reject).

## Per-ID evidence

### C-W19-FE-005 — pending leave error≠empty

```text
BASELINE
→ PASS shows approval queue alert when listPendingLeave fails, not silent empty

BREAK: leave-requests.tsx
  - gate: (pendingError || pendingItems.length > 0 || pendingLoading) → pendingItems.length > 0
  - remove branch pendingError ? <QueryErrorAlert … />
→ FAIL Unable to find role="alert"
  (UI shows only «Заявок пока нет» — silent empty / queue hidden)

RESTORE: git checkout -- frontend/src/modules/hr/leave-requests.tsx
→ PASS same it + twin hides approval queue on success+empty
  Test Files 1 passed · Tests 2 passed (~2.9s)
```

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs

# baseline GREEN
npx vitest run --reporter=verbose \
  ../testframework/stages/unit/frontend-test/domains/hr/leave-pending-error.test.tsx \
  -t "shows approval queue alert when listPendingLeave fails, not silent empty"
→ 1 passed | 1 skipped

# after BREAK → RED (exit 1): Unable to find role="alert"

# after RESTORE → GREEN
npx vitest run --reporter=verbose \
  ../testframework/stages/unit/frontend-test/domains/hr/leave-pending-error.test.tsx
→ Test Files 1 passed · Tests 2 passed (~2.9s)
```

## Assumptions / unresolved

- Коммит не создавался (инструкция слота).
- Criteria authored by fe2 (W7); mutation prove by qa-auto1 (W8 slot 5).
- Optional `pipeline-status.tsx` sanitize — N/A (не ID в criteria).

## Handoff

TASK-182 **Закрыта** · evidence этот файл · README синхронизирован · mutation gate закрыт.
