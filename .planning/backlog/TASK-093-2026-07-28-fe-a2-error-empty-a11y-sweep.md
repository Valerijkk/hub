# TASK-093 — 🧹 FE A2: error≠empty residual + a11y delete focus

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: На проверке  
> Источник: A2W9/W10/W11 · Stage A2 dedup · TASK-008

## [OUTCOME]

Оставшиеся A2 FE residual из TASK-047/048/049: isError≠empty, catch→toast, presence offline≠error; notification delete keyboard-visible.

## [ROLE]

- Role: fe1
- Responsibility: фикс + regression/mutation-ready тест
- Handoff to: be-ver/fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - frontend/src/modules/tracker/sprint-backlog.tsx
  - frontend/src/modules/** (по списку A2W10)
  - frontend/src/components/**/notification-panel.tsx
  - frontend/src/hooks/use-presence.ts
  - testframework unit frontend
  - `.planning/backlog/README.md` + этот файл
- Do not touch:
  - новые фичи UI
- Constraints: Conventional Commits RU description; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A (кроме schema)

## [REFERENCES]

- A2W10 · A2W9-4-001/002 · A2W11-3-001 · TASK-047/048
- Dedup: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - Чеклист ID из A2-DEDUP закрыт или WONTFIX с причиной
  - notification-panel group-focus-within для delete
  - ≥3 regression RTL
  - Мутация: silent catch → тест краснеет
- Regression: затронутый `./testframework/...` или frontend unit
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Diff summary
- Test commands + PASS
- Mutation note
- Статус задачи обновлён

## [OPERATING RULES]

- Не закрывать соседние A2 ID «заодно» без evidence
- Docs sync только если контракт API изменился

---

## Progress (fe1 · 2026-07-28)

### Checklist ID

| ID | Вердикт | Evidence |
|----|---------|----------|
| A2W9-4-001 | **FIXED** | `saved-filters.tsx` — `toastApiError` + `loadError`/`QueryErrorAlert` |
| A2W9-4-002 | **FIXED** | `notification-panel.tsx` — `group-focus-within:flex` на delete |
| A2W11-3-001 | **FIXED** | `use-presence.ts` — `isError`/`error`/`refetch`; channel-list → «Статус недоступен» |
| A2W11-3-003 | **FIXED** | heartbeat `console.warn` + счётчик, не silent `catch {}` |
| A2W10-A1-001 | **FIXED** | `accounting-page.tsx` — `isError` early return |
| A2W10-A1-002 | **FIXED** | `backups.tsx` — catch + `QueryErrorAlert` + toast mutations |
| A2W10-A1-003 | **FIXED** | `crm-page.tsx` — contacts `isError` ≠ «Нет контактов» |
| A2W10-A1-004 | **FIXED** | `recordings-list.tsx` — load/download toast + alert |
| A2W10-A1-005 | **FIXED** | `api-keys.tsx` — loadError + toast create/revoke |
| A2W10-A1-006 | **FIXED** | `ai-privacy.tsx` — fail load ≠ default form `allowExternal` |
| A2W10-A1-007 | **FIXED** | `project-templates.tsx` — loadError + toast |
| A2W10-A2-001 | **FIXED** | `webhooks.tsx` — loadError + toast create |
| A2W10-A2-002 | **FIXED** | `environments.tsx` — loadError + toast |
| A2W10-A2-003 | **FIXED** | `github-links.tsx` — loadError ≠ empty PR list |
| A2W10-A2-004 | **FIXED** | burndown store+chart error/retry |
| A2W10-A2-005 | **FIXED** | `import-export.tsx` — toast preview/commit/export |
| A2W10-A3-002 | **FIXED** | `sprint-backlog.tsx` — fail ≠ «Бэклог пуст» |
| A2W10-A3-003 | **FIXED** | velocity store+chart error/retry |
| A2W10-A3-009 | **FIXED** | checkbox `aria-label` на строках бэклога |
| A2W10-A4-001 | **FIXED** | ChatErrorBoundary remount children via `retryCount` key |
| A2W10-A4-002 | **FIXED** | `$issueKey.tsx` — `isError` ≠ not-found |
| A2W10-A5-002 | **FIXED** | `cases_.$caseKey` — `isError` ≠ eternal shimmer |
| A2W10-A5-003 | **FIXED** | `runs_.$runId` — `isError` ≠ eternal shimmer |
| A2W10-A3-001 | **N/A** | TASK-082 |
| A2W10-A5-001 | **N/A** | TASK-083 |
| A2W11-3-002 | **N/A** | TASK-103 |
| A2W9-4-003 | **FIXED** | `notification-panel.tsx` — `var(--blue)` / `--gB` / semantic TYPE_* colors |
| A2W10-A2-006 | **FIXED** | `issue-description.tsx` — zod `plateValueSchema.safeParse` |
| A2W10-A2-007 | **FIXED** | `coverage-report.tsx` — `role=button` + `tabIndex` + Enter/Space |
| A2W10-A2-008 | **FIXED** | `CaseTreePanel` wired in `cases.lazy.tsx`; duplicate tree removed |
| A2W10-A3-004 | **FIXED** | `issue-sidebar-meta.tsx` — `Label htmlFor` на Статус/Приоритет/SP |
| A2W10-A3-005 | **FIXED** | `sprint-board.tsx` — actions вне `<Link>` (no nested interactive) |
| A2W10-A3-006 | **FIXED** | `planning-poker.tsx` — `radiogroup` + `role=radio` + `aria-checked` |
| A2W10-A3-007 | **FIXED** | `admin/dashboard.tsx` — `TabsList`/`TabsTrigger` |
| A2W10-A3-008 | **FIXED** | assignee read-only без мёртвого `cursor-pointer` |

### Тесты

```
cd frontend
node scripts/link-testframework-node-modules.mjs
npx tsc --noEmit   # PASS
npx vitest run \
  ../testframework/stages/unit/frontend-test/domains/platform/notification-tokens.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/platform/notification-delete-focus.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tms/coverage-report-keyboard.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tms/case-tree-panel.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tracker/planning-poker-radiogroup.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tracker/issue-sidebar-meta-a11y.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tracker/sprint-board-nested.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/platform/api-dto-gaps-p2.test.ts \
  ../testframework/stages/unit/frontend-test/domains/tracker/sprint-backlog-error.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/messenger/use-presence-error.test.tsx
# 10 files · 17 tests · PASS
```

### Mutation note

- `saved-filters-error`: убрать `toastApiError` из `handleCreate` catch → тест «toasts on create failure» краснеет.
- `sprint-backlog-error`: вернуть silent `catch {}` без `loadError` → снова «Бэклог пуст», alert пропадает.
- `notification-delete-focus`: убрать `group-focus-within:flex` → className assertion FAIL.
- `use-presence-error`: не прокидывать `isError` → assertion FAIL.
- `notification-tokens`: вернуть `border-blue-500` / `bg-blue-500` → FAIL.
- `coverage-report-keyboard`: убрать `role=button`/`tabIndex` → FAIL.
- `planning-poker-radiogroup`: убрать `role=radiogroup` → FAIL.
- `sprint-board-nested`: обернуть Button снова в Link → `closest("a")` FAIL.

### Ветка

`fix/frontend/task-093-error-empty-a11y` (worktree)

### Sweep 2 (fe1 · 2026-07-28)

Owner lock: все WONTFIX закрыты (WONTFIX=0). Статус остаётся **Исправлена**.
