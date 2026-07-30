# MUTATION W9 slot 5 — TASK-186 (fe1) · 2026-07-30

> Checker: **fe1** · Env: local Vitest · Criteria: `.planning/backlog/TASK-186-2026-07-29-audit-fe-calls-error-empty.md` `[MUTATION CRITERIA]` (author: fe2 W7)  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**5/5 PROVEN** (001–004, 006) · **005 = N/A** (GAP: нет отдельного `it`)  
TASK-186 → **Закрыта** (005 mutation explicit N/A; code-fix 005 остаётся FIXED, без fake-PROVEN)

| ID | Verdict | RED test (`it`) | Mutation (BREAK) | Layer |
|----|---------|-----------------|------------------|-------|
| FE-MC-001 | **PROVEN** | `shows alert+retry when getChannels fails, not month grid` | drop `channelsError && <QueryErrorAlert>`; `showCalendarGrid = !channelsLoading` | frontend |
| FE-MC-002 | **PROVEN** | `shows projects alert+retry…`; `shows channels alert+retry…` | drop projects/channels `QueryErrorAlert`; `projectsEmpty`/`channelsEmpty` без `!…Error` | frontend |
| FE-MC-003 | **PROVEN** | `event popover close control has accessible name (FE-MC-003)` | drop `aria-label="Закрыть"` с backdrop + × | frontend |
| FE-MC-004 | **PROVEN** | `channel SelectTrigger has accessible name (FE-MC-004)` | drop `aria-label="Выбор канала"` | frontend |
| FE-MC-005 | **N/A** | — (нет `it`) | criteria GAP: remove-participant aria-label | frontend |
| FE-MC-006 | **PROVEN** | `week slot aria-label includes weekday/date, not only hour (FE-MC-006)` | aria-label только hour (`Создать встречу на HH:00`) | frontend |

Production restored (`git checkout --` на 3 файла). Mock-gap: нет (RTL на prod-коде). Leftover MUTATED: нет.

## Per-ID evidence

### FE-MC-001 — channels error≠empty

```text
BASELINE GREEN → PASS shows alert+retry when getChannels fails, not month grid
BREAK: call-calendar.tsx — remove channelsError QueryErrorAlert;
  showCalendarGrid = !channelsLoading (grid on fail)
→ FAIL Unable to find role="alert" (table month grid present)
RESTORE: git checkout -- frontend/src/modules/calls/call-calendar.tsx
→ PASS same it (2/2 file)
```

### FE-MC-002 — projects/channels error≠empty

```text
BREAK: create-call-dialog.tsx — remove projectsError/channelsError → alert;
  projectsEmpty/channelsEmpty = length===0 (без !…Error)
→ FAIL shows projects alert+retry… (no role=alert)
→ FAIL shows channels alert+retry… (no role=alert)
  (success+empty + FE-MC-004 still PASS)
RESTORE: git checkout -- frontend/src/modules/calls/create-call-dialog.tsx
→ PASS 4/4 file
```

### FE-MC-003 — popover close accessible name

```text
BREAK: call-calendar.tsx — remove aria-label="Закрыть" from backdrop + × buttons
→ FAIL getAllByRole("button", { name: /^закрыть$/i }) — Unable to find
RESTORE: git checkout -- frontend/src/modules/calls/call-calendar.tsx
→ PASS same it (2/2 a11y file)
```

### FE-MC-004 — channel SelectTrigger name

```text
BREAK: create-call-dialog.tsx — remove aria-label="Выбор канала" from SelectTrigger
→ FAIL channel SelectTrigger has accessible name (FE-MC-004)
  (getByRole combobox name=/выбор канала/i)
RESTORE: git checkout -- frontend/src/modules/calls/create-call-dialog.tsx
→ PASS 4/4 file
```

### FE-MC-005 — remove-participant aria-label — **N/A**

```text
GAP per [MUTATION CRITERIA]: нет отдельного it.
Code FIXED: create-call-dialog.tsx aria-label={`Удалить участника ${member.name}`}
Verdict: N/A — не ставить PROVEN без слоя теста. Не блокер закрытия
  (instruction: FE-MC-005 без it → N/A; Закрыта если runnable PROVEN).
```

### FE-MC-006 — week slot weekday/date

```text
BREAK: calendar-week-view.tsx — aria-label={`Создать встречу на ${HH}:00`}
  (без weekday/date)
→ FAIL week slot aria-label… (expected not to match /^Создать встречу на \d{2}:00$/)
RESTORE: git checkout -- frontend/src/modules/calls/calendar-week-view.tsx
→ PASS same it
```

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs
# baseline + post-restore:
npx vitest run \
  ../testframework/stages/unit/frontend-test/domains/calls/call-calendar-error-empty.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/calls/create-call-dialog-error-empty.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/calls/call-calendar-a11y.test.tsx
→ Test Files 3 passed · Tests 8 passed (~4.3s) after full restore
```

## Assumptions / unresolved

- FE-MC-005 остаётся code-FIXED без mutation-proof `it` (known GAP в criteria).
- Leftover MUTATED: нет (`git status` clean на `frontend/src/modules/calls/`).

## Handoff

Evidence: этот файл · TASK-186 **Закрыта** · fe-ver → code-reviewer (error≠empty + a11y already FIXED).
