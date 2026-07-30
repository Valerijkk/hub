# MUTATION W9 slot 4 — TASK-172 (fe2) · 2026-07-30

> Checker: **fe2** (!= fe1 criteria author) · Env: local Vitest  
> Criteria: `.planning/backlog/TASK-172-2026-07-29-a3-9-1-components.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**5/5 PROVEN** · TASK-172 → **Закрыта**

| ID | Verdict | RED test (`it`) | Mutation (BREAK) | Layer |
|----|---------|-----------------|------------------|-------|
| A3W9-1-001 | **PROVEN** | `shows QueryErrorAlert on load fail, not empty copy` | omit `setLoadError(true)` in catch → empty «Нет заявок…» | frontend |
| A3W9-1-002 | **PROVEN** | `shows alert+retry on getLimits fail, not save form` | omit `setLoadError(true)` → save-форма «Сохранить» без alert | frontend |
| A3W9-1-003 | **PROVEN** | `shows alert+retry on getTrash fail, not empty copy` | omit `setLoadError(true)` → «Корзина пуста» | frontend |
| A3W9-1-004 | **PROVEN** | `shows banner+retry on fetchLogs fail, not empty-as-fail (A3W9-1-004)` | omit `setLoadError(true)` → нет banner/alert | frontend |
| A3W9-1-005 | **PROVEN** | `exposes aria-label and aria-pressed for logging toggle` | remove `aria-pressed` + `aria-label` с logging toggle | frontend |

Production restored (`git checkout --` на 5 файлах). Mock-gap: нет (RTL на prod-компонентах + mocked API reject). Leftover MUTATED: нет (`git status` clean на 5 prod-файлах).

## Per-ID evidence

### A3W9-1-001 — pending-users loadError

```text
BASELINE GREEN → PASS shows QueryErrorAlert on load fail, not empty copy

BREAK: pending-users-panel.tsx catch — omit setLoadError(true)
→ FAIL Unable to find role="alert"
  (UI shows «Нет заявок на регистрацию»)

RESTORE: git checkout -- frontend/src/components/admin/pending-users-panel.tsx
→ PASS same it
```

### A3W9-1-002 — resource-limits loadError

```text
BREAK: resource-limits.tsx catch — omit setLoadError(true)
→ FAIL Unable to find role="alert"
  (UI shows save-форма / кнопка «Сохранить» без лимитов)

RESTORE: git checkout -- frontend/src/components/admin/resource-limits.tsx
→ PASS same it
```

### A3W9-1-003 — trash-manager loadError

```text
BREAK: trash-manager.tsx catch — omit setLoadError(true)
→ FAIL Unable to find role="alert"
  (UI shows «Корзина пуста»)

RESTORE: git checkout -- frontend/src/components/admin/trash-manager.tsx
→ PASS same it
```

### A3W9-1-004 — audit-log loadError banner

```text
BREAK: audit-log.tsx catch — omit setLoadError(true)
→ FAIL Unable to find role="alert"
  (banner QueryErrorAlert не рендерится)

RESTORE: git checkout -- frontend/src/components/admin/audit-log.tsx
→ PASS same it
```

### A3W9-1-005 — CallChat logging aria-*

```text
BREAK: CallChat.tsx logging toggle — remove aria-pressed + aria-label
→ FAIL Unable to find role="button" name /выключить логирование чата/i

RESTORE: git checkout -- frontend/src/components/calls/CallChat.tsx
→ PASS same it
```

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs

# baseline + post-restore:
npx vitest run \
  ../testframework/stages/unit/frontend-test/domains/admin/pending-users-panel.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/admin/resource-limits.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/admin/trash-manager.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/admin/audit-log.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/calls/call-chat-logging-a11y.test.tsx
→ Test Files 5 passed · Tests 17 passed (~4.3s) after full restore
```

## Assumptions / unresolved

- Нет: criteria CONFIRMED 5/5; все ID ловятся named `it` на заявленном FE-слое.
- GAP OOS (`bg-green-500/15` tokens) — не мутировали (не ID задачи).

## Handoff

Evidence: этот файл · TASK-172 **Закрыта** · fe-ver → code-reviewer (5/5 FIXED + PROVEN).
