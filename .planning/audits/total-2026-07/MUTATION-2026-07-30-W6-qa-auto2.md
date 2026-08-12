# MUTATION W6 slot 4 — TASK-179 (qa-auto2) · 2026-07-30

> Checker: **qa-auto2** (!= fe1 fixer) · Env: local Vitest · Criteria: `.planning/backlog/archive/2026-07-29/TASK-179-2026-07-29-c-w19-fe-webhook-sprint-completed.md` `[MUTATION CRITERIA]`

## Verdict

**1/1 PROVEN** · Break→RED→restore→GREEN · статус TASK-179 → **Закрыта**

| ID | Verdict | RED tests | Mutation (BREAK) | Layer |
|----|---------|-----------|------------------|-------|
| C-W19-FE-001 | **PROVEN** | оба `it` в describe `TASK-179 C-W19-FE-001 webhook events allowlist` | добавить `"sprint.completed"` в `AVAILABLE_EVENTS` | frontend |

Production restored (StrReplace reverse; `git diff` webhooks.tsx empty). Mock-gap: нет — оба it бьют по export + RTL create-form.

## Per-ID evidence

### C-W19-FE-001 — FE allowlist без `sprint.completed`

```text
BASELINE:
  npx vitest run …/webhook-events-sprint-completed-c-w19.test.tsx
  → 2/2 PASS

BREAK: frontend/src/modules/settings/webhooks.tsx AVAILABLE_EVENTS
  + "sprint.completed" (вернуть pre-trim опцию)
→ FAIL both:
  - AVAILABLE_EVENTS does not include sprint.completed
    expected […(4)] to not include 'sprint.completed'
  - create form checkboxes omit sprint.completed
    expected <label>sprint.completed</label> to be null
→ 2 failed / 2

RESTORE: убрать "sprint.completed"
→ PASS 2/2
```

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs

# baseline / RED / GREEN
npx vitest run ../testframework/stages/unit/frontend-test/domains/platform/webhook-events-sprint-completed-c-w19.test.tsx

# regression (evidence suite)
npx vitest run \
  ../testframework/stages/unit/frontend-test/domains/platform/trash-webhooks-p1.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/platform/tokens-dead-ui-authz-p2.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/platform/webhook-events-sprint-completed-c-w19.test.tsx
→ 14/14 PASS (~4s)
```

## Assumptions / unresolved

- `npx tsc --noEmit` не ловит string-literal в `as const` массиве — зафиксировано в criteria Gaps; guard = Vitest only.
- BE 400 на неизвестный event — OOS FE (TASK-152); FE-критерий = UI allowlist.
- Коммит не создавался (инструкция слота).

## Handoff

TASK-179 **Закрыта** · evidence этот файл · README синхронизирован · next: qa-auto-ver / code-reviewer по желанию волны (mutation gate закрыт).
