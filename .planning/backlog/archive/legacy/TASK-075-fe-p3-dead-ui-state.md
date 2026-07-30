# TASK-075 — FE dead UI / unused state

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Закрыта
> Источник: TASK-008 W12 · Classes: P3-FE-08 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-FRONTEND.md`
> Mutation C (2026-07-30, W13 fe1): **9/11 PROVEN · 2 NOT_PROVEN** → `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-fe1-task075.md`
> Mutation C retry (2026-07-30, W13 fe1 slot 2): **2/2 PROVEN → 11/11** → evidence `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-fe1-task075-retry.md`

## [OUTCOME]

Дефекты P3-FE-08 закрыты.

## [ROLE]

- Role: fe1
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: frontend/src/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P3-FE-08)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-FE-08 с мутацией «да» закрыты по реестру
- **Мутация (C):** dead filter → wire or remove; cursor-pointer without handler gone
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён

## [MUTATION CRITERIA] — verifier-ready

Проверено по коду HEAD (2026-07-30, fe2 W13 slot 5 criteria-only). Слой **frontend** (Class **P3-FE-08**). Typecheck: `cd frontend && npx tsc --noEmit` — не ловит dead UI/state.

| ID | file:line | Vitest path + `it` | kill recipe |
|----|-----------|----------------------|-------------|
| W9-A2-019 | `frontend/src/components/dashboard/widgets/deadlines.tsx:67-93` + `task-row-link.tsx:15-28` — `TaskRowLink` при `project_id`+`key` | `domains/dashboard/a3w9-2-widgets.test.tsx` · describe `Dashboard widgets A3W9-2 (TASK-173)` > `it("deadlines row links to issue route with project_id + key (A3W9-2-002)")` | заменить `TaskRowLink` на `<div>` / убрать `project_id` из row → `getByRole("link")` FAIL |
| W9-A3-023 | `frontend/src/components/messenger/mention-autocomplete.tsx:92` / `:177` / `:208` — Esc → `setSuppressed(true)` + `!suppressed` | `domains/platform/dead-ui-state-p3.test.ts` · describe `TASK-075 P3-FE-08 dead UI / unused state` > `it("mention-autocomplete Esc sets suppressed")` | Esc только `return true` без `setSuppressed` → source-guard FAIL |
| W9-A4-012 | `frontend/src/components/tracker/issue-attachments.tsx:216` — empty «Нажмите «Добавить файл»» (без DnD) | same file > `it("issue-attachments empty copy does not promise DnD")` | вернуть «Перетащите…» → `not.toMatch(/Перетащите/)` FAIL |
| W10-A2-025 | `frontend/src/modules/tms/components/run-execution-panel.tsx:74-77` — нет `stepResults` state | same file > `it("run-execution-panel has no dead stepResults state")` | вернуть `useState` `stepResults` без UI → `not.toMatch(/stepResults/)` FAIL |
| W10-A3-019 | `frontend/src/modules/tracker/issue-sidebar-meta.tsx:128-151` — Исполнитель/Автор без `cursor-pointer` | same file > `it("issue-sidebar-meta assignee/reporter are not cursor-pointer")` | добавить `cursor-pointer` на блоки без handler → regex FAIL |
| W10-A4-009 | `frontend/src/routes/_authenticated/projects/$projectId/issues/index.tsx:52-54` — filters без `statusFilter` | same file > `it("issues list has no dead statusFilter state")` | вернуть `statusFilter`/`setStatusFilter` без Select → source-guard FAIL |
| W10-A4-010 | `frontend/src/routes/_authenticated/projects/$projectId/tms/cases.lazy.tsx:106-110` / `:289-300` — `bulkDeleteMutation` + «Удалить» | same file > `it("tms cases bulk bar has delete action")` | убрать bulk delete (оставить только «Снять») → `bulkDeleteMutation`/`Удалить` FAIL |
| W10-A5-014 | `frontend/src/routes/_authenticated/settings.tsx:90-94` — `NavItem` без `placeholder?` ветки | same file > `it("settings NavItem has no unused placeholder branch")` | вернуть `placeholder?:` + `if (placeholder)` → source-guard FAIL |
| W10-A5-016 | `frontend/src/routes/_authenticated/settings/profile.lazy.tsx:163-164` — avatar `onClick` → `fileInputRef.click()` | same file > `it("profile avatar opens file input; no dead department state")` | убрать `onClick` с аватара (оставить только кнопку) → `fileInputRef.current?.click()` / aria-label assert FAIL |
| W10-A5-017 | `frontend/src/routes/_authenticated/settings/profile.lazy.tsx:81` — `const statusDuration = "1d"`; нет `department` в form | same `it` | вернуть `department:` в state / `useState("1d")` без setter → `not.toMatch(/department:/)` / `const statusDuration = "1d"` FAIL |
| W11-A3-016 | `frontend/src/stores/tms-store.ts:19-33` — filters-only store | same file > `it("tms-store is filters-only (no server-cache fields)")` | вернуть `testCases`/`setTestCases`/`suites`/`testRuns` → `not.toHaveProperty("testCases")` FAIL |

### Gaps (mutation-proof)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| 10 findings в `dead-ui-state-p3.test.ts` | ✅ source-guard (+ runtime store) | Class IDs кроме deadlines |
| W9-A2-019 deadlines Link | ✅ RTL в `a3w9-2-widgets.test.tsx` | не в dead-ui file — ок, тот же слой FE |
| Docs-only | N/A | нет |

**Confidence:** HIGH — **11/11** Class source IDs с runnable criteria.
