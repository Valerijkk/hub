# TASK-070 — FE a11y structure residual

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Закрыта
> Источник: TASK-008 W12 · Classes: P3-FE-02 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-FRONTEND.md`
> C-мутация (2026-07-30, fe2 W13 slot 5 retry): **7/7 PROVEN** · 015/017 GAP_NOTED · evidence `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-fe2-task070-retry.md`

## [OUTCOME]

Дефекты P3-FE-02 закрыты.

## [ROLE]

- Role: fe2
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: frontend/src/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P3-FE-02)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-FE-02 с мутацией «да» закрыты по реестру
- **Мутация (C):** tablist/heading/aria-hidden/required
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

Проверено по коду HEAD (2026-07-30, fe1 W13 slot 2). Слой **frontend** (Class **P3-FE-02**). Typecheck: `cd frontend && npx tsc --noEmit` — не ловит roles/aria-*.

| ID | file:line | Vitest path + `it` | kill recipe |
|----|-----------|----------------------|-------------|
| W9-A1-018 | `frontend/src/components/admin/trash-manager.tsx:80-86` — `role="tablist"` / `role="tab"` / `aria-selected` | `domains/platform/a11y-structure-p3.test.tsx` · describe `P3-FE-02 · a11y structure residual (TASK-070)` > `it("TrashManager tabs use tablist/tab/aria-selected")` | plain `<button>` без tab roles → tablist/tab asserts FAIL |
| W9-A4-013 | `frontend/src/components/route-error-boundary.tsx:61` — `role="alert"` | same file > `it("RouteErrorBoundary fallback has role=alert")` | убрать `role="alert"` → alert query FAIL |
| W9-A5-013 | `frontend/src/components/ui/card.tsx:36` — `CardTitle` = `<h3>` | same file > `it("CardTitle renders as heading")` | вернуть `<div>` вместо `<h3>` → heading role FAIL |
| W9-A5-016 | `frontend/src/components/ui/empty-state.tsx:24-25` — `aria-hidden="true"` на icon wrap | same file > `it("EmptyState decorative icon is aria-hidden")` | убрать `aria-hidden` с wrap → wrap `toHaveAttribute("aria-hidden")` FAIL (не lucide SVG default) |
| W9-A5-012 | `frontend/src/components/ui/avatar.tsx:37` — `{isOnline && <AvatarOnlineLabel />}` (`status.online` sr-only) | same file > `it("Avatar announces online status when data-online")` | не рендерить `AvatarOnlineLabel` при `data-online` → «В сети» FAIL |
| W9-A5-014 | `frontend/src/components/ui/badge.tsx:33` — `<span>` (не focusable control) | same file > `it("Badge renders as span without focus ring classes")` | заменить на `<button>` / добавить `focus:ring` → tag/class asserts FAIL |
| W9-A5-019 | `frontend/src/components/ui/form-field.tsx:31` / `:39` — `aria-required: true` при `required` | same file > `it("FormField required sets aria-required on control")` | не прокидывать `aria-required` → attribute assert FAIL |

### Gaps (mutation-proof)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| trash tabs / route alert / card / empty wrap / avatar / badge / form-field | ✅ RTL | 7 runnable PROVEN |
| W9-A5-015, W9-A5-017 | ❌ GAP | Class source IDs без dedicated `it` |
| `npx tsc --noEmit` | N/A | не primary |

**Confidence:** HIGH — 7/7 runnable PROVEN (W13 fe2 retry); 015/017 GAP_NOTED.
