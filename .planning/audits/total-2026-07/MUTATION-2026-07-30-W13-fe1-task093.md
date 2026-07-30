# MUTATION W13 slot 2 — TASK-093 (fe1) · 2026-07-30

> Checker: **fe1** · Env: local Vitest · Criteria: `.planning/backlog/TASK-093-2026-07-28-fe-a2-error-empty-a11y-sweep.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**13 PROVEN** · **1 NOT_PROVEN** (A2W10-A2-008) · **GAP_NOTED / N/A** по criteria для FIXED без `it`.  
Partial OK. TASK-093 → **На проверке** (mutation evidence; fe-ver: 008 + GAP rows).

| ID | Verdict | RED |
|----|---------|-----|
| A2W9-4-001 | **PROVEN** | `shows alert + retry…`; `toasts on create failure…` |
| A2W9-4-002 | **PROVEN** | `delete button uses group-focus-within so keyboard focus reveals control` |
| A2W9-4-003 | **PROVEN** | `unread row uses design tokens, not raw Tailwind blue-500` |
| A2W11-3-001 | **PROVEN** | `exposes isError when presence poll fails — not silent offline-only` |
| A2W11-3-003 | **PROVEN** | `logs heartbeat failures instead of empty catch (A2W11-3-003)` |
| A2W10-A3-002 | **PROVEN** | `shows alert + retry instead of «Бэклог пуст» when listIssues fails` |
| A2W10-A2-007 | **PROVEN** | `expandable issue row is keyboard-reachable` |
| A2W10-A2-008 | **NOT_PROVEN** | — (cases.lazy unwire; Vitest imports `CaseTreePanel` напрямую) |
| A2W10-A2-006 | **PROVEN** | `parsePlateValue accepts plate nodes and rejects garbage` |
| A2W10-A3-004 | **PROVEN** | `binds Status/Priority labels via htmlFor` |
| A2W10-A3-008 | **PROVEN** | `assignee empty state is not a dead cursor-pointer control` |
| A2W10-A3-005 | **PROVEN** | `action button is not nested inside the sprint link` |
| A2W10-A3-006 | **PROVEN** | `vote cards expose radiogroup + aria-checked` |
| A2W10-A3-009 | **GAP_NOTED** | нет `it` на aria-label checkbox |
| A2W10-A3-007 | **GAP_NOTED** | нет Vitest на Tabs semantics |
| A2W10-A1-001…007 | **GAP_NOTED** | нет TASK-093-привязанных `it` |
| A2W10-A2-001…005 | **GAP_NOTED** | нет TASK-093 mutation `it` |
| A2W10-A3-003 | **GAP_NOTED** | velocity ≠ error≠empty mutation `it` |
| A2W10-A4-001 | **GAP_NOTED** | Chat remount ≠ общий ErrorBoundary test |
| A2W10-A4-002 | **GAP_NOTED** | нет dedicated `it` isError≠not-found |
| A2W10-A5-002/003 | **GAP_NOTED** | нет dedicated `it` shimmer |
| A2W10-A3-001 | **N/A** | TASK-082 |
| A2W10-A5-001 | **N/A** | TASK-083 |
| A2W11-3-002 | **N/A** | TASK-103 |

Production restored (`git checkout --`). Leftover MUTATED: нет. Mock-gap: нет (RTL на prod-компонентах/хуках).

## Per-ID evidence

### A2W9-4-001 — saved-filters toast + loadError

```text
BREAK: saved-filters.tsx — silent catch load/create; drop loadError → QueryErrorAlert
→ FAIL shows alert + retry instead of silent empty when listSavedFilters fails
→ FAIL toasts on create failure — silent catch would leave mutation invisible
RESTORE → PASS 2/2
```

### A2W9-4-002 — notification delete focus

```text
BREAK: notification-panel.tsx — drop group-focus-within:flex
→ FAIL delete button uses group-focus-within…
RESTORE → PASS
```

### A2W9-4-003 — notification tokens

```text
BREAK: border-blue-500 / bg-blue-500 вместо var(--blue)/var(--gB)
→ FAIL unread row uses design tokens…
RESTORE → PASS
```

### A2W11-3-001 — usePresence isError

```text
BREAK: isError: false (всегда)
→ FAIL exposes isError when presence poll fails…
RESTORE → PASS
```

### A2W11-3-003 — heartbeat silent catch

```text
BREAK: sendHeartbeat().catch(() => {})
→ FAIL logs heartbeat failures instead of empty catch
RESTORE → PASS
```

### A2W10-A3-002 — sprint-backlog error≠empty

```text
BREAK: silent catch (no setLoadError/toast) + drop loadError → QueryErrorAlert
→ FAIL shows alert + retry… (UI «Бэклог пуст»)
RESTORE → PASS
```

### A2W10-A2-007 — coverage keyboard

```text
BREAK: drop role="button" / tabIndex={0}
→ FAIL expandable issue row is keyboard-reachable
RESTORE → PASS
```

### A2W10-A2-008 — CaseTreePanel wiring — **NOT_PROVEN**

```text
BREAK: cases.lazy.tsx — remove <CaseTreePanel …/>
→ still PASS renders suite tree… (тест импортирует CaseTreePanel из module, не route)
Verdict: NOT_PROVEN — test_does_not_catch wiring guard.
Нужен source-guard / route-level it на cases.lazy, либо mutate export.
```

### A2W10-A2-006 — parsePlateValue zod

```text
BREAK: always-accept raw as Value (bypass safeParse)
→ FAIL parsePlateValue… rejects garbage
RESTORE → PASS
```

### A2W10-A3-004 — Label htmlFor

```text
BREAK: remove htmlFor status/priority/SP
→ FAIL binds Status/Priority labels via htmlFor
RESTORE → PASS
```

### A2W10-A3-008 — assignee cursor-pointer

```text
BREAK: add cursor-pointer на «Не назначен»
→ FAIL assignee empty state is not a dead cursor-pointer control
RESTORE → PASS
```

### A2W10-A3-005 — nested interactive

```text
BREAK: move action Buttons inside <Link>
→ FAIL action button is not nested inside the sprint link (closest("a") ≠ null)
RESTORE → PASS
```

### A2W10-A3-006 — radiogroup

```text
BREAK: drop role="radiogroup" + aria-checked
→ FAIL vote cards expose radiogroup + aria-checked
RESTORE → PASS
```

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs
npx vitest run \
  ../testframework/stages/unit/frontend-test/domains/platform/saved-filters-error.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/platform/notification-delete-focus.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/platform/notification-tokens.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/messenger/use-presence-error.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tracker/sprint-backlog-error.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tms/coverage-report-keyboard.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tms/case-tree-panel.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/platform/api-dto-gaps-p2.test.ts \
  ../testframework/stages/unit/frontend-test/domains/tracker/issue-sidebar-meta-a11y.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tracker/sprint-board-nested.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tracker/planning-poker-radiogroup.test.tsx
→ after full restore: Test Files 11 passed · Tests 19 passed (~2.9s)
```

## Assumptions / unresolved

- A2W10-A2-008: component export покрыт, **wiring в `cases.lazy.tsx` — нет**; не fake-PROVEN.
- GAP rows (A1-\*, A2-001…005, A3-003/007/009, A4, A5): FIXED claimed без mutation-ready `it` — criteria GAP, не ломать UI «вслепую».
- `npx tsc --noEmit` не primary mutation layer (как в criteria).
- Leftover MUTATED: нет.

## Handoff

Evidence: этот файл · TASK-093 → **На проверке** · 13 PROVEN / 1 NOT_PROVEN / GAPs documented · fe-ver.
