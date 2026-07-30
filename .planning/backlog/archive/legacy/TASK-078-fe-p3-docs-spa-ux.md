# TASK-078 — Docs/SPA copy/UTF-8 + link-preview URLs

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Закрыта
> Источник: TASK-008 W12 · Classes: P3-FE-13,P3-FE-14,P3-FE-15 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-FRONTEND.md`

## [OUTCOME]

Дефекты P3-FE-13..15 закрыты.

## [ROLE]

- Role: docs-writer
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: docs/**; frontend/src/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P3-FE-13,P3-FE-14,P3-FE-15)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-FE-13,P3-FE-14,P3-FE-15 с мутацией «да» закрыты по реестру
- **Мутация (C):** docs pins sync; mojibake gone; preview URL safe
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

> W13 slot 1 refill / be1 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified · Classes **P3-FE-13**, **P3-FE-14**, **P3-FE-15**

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| W11-A4-017 | `docs/architecture.md:105` — routes pin **7x** `.tsx` (не `~60+`) | docs | `it("architecture.md pins route count (not ~60+)")` · `testframework/stages/unit/frontend-test/domains/platform/docs-spa-ux-p3.test.ts` | вернуть `~60+` / убрать routes pin → expect RED |
| W11-A5-020 | `docs/stack_and_tools.md:60,67,73` — **12** stores; motion `^12.40.0`; pwa `^1.3.0` | docs | `it("stack_and_tools.md pins 12 stores and current motion/pwa")` · same file | `11+` / старые pin'ы → Match/not Match RED |
| W11-A5-021 | `docs/wiki/API.md:74-85` — magic-link curl Caddy `:8080` | docs | `it("API.md magic-link curl uses Caddy :8080")` · same file | curl только `:3000` без Caddy → Match RED |
| W10-A2-021 | `frontend/src/routes/_authenticated/settings/security.tsx:58-63` — `<Link to="/settings/profile">` | FE route | `it("security settings uses TanStack Link, not <a href>")` · same file | `<a href="/settings/profile">` → Link Match / href not Match RED |
| W10-A5-015 | `frontend/src/routes/_authenticated/trash.tsx:195` — колонка «Удалено» | FE route | `it("trash column is Удалено (date), not Удалил")` · same file | «Удалил» → Match/not Match RED |
| W10-A5-018 | `frontend/src/modules/tms/components/run-results-table.tsx:221` — UTF-8 «Исполнитель» | FE module | `it("run-results-table has correct UTF-8 Исполнитель")` · same file | corrupt UTF-8 (replacement chars in label) → `not.toMatch` / `toMatch(/Исполнитель/)` RED |
| W10-A1-022 | `frontend/src/modules/messenger/message-bubble.tsx:118-119` — `extractUrls(displayText)` ∪ `contentText` | FE module | `it("message-bubble extracts URLs from displayText and contentText")` · same file | только `extractUrls(contentText)` → displayText Match RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| W11-A4-017, W11-A5-020, W11-A5-021 | docs source-guard (Vitest readFile) | **docs-only** |
| W10-A2-021, W10-A5-015, W10-A5-018, W10-A1-022 | FE source-guard | **OK** |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: W11-A4-017, W11-A5-020, W11-A5-021`

### Gaps

- Docs-only (P3-FE-13): **3/3** — mutation surface = docs files; Vitest source-guards (не N/A: registry mutation «да»).
- Class source IDs = **7/7** с dedicated RED.
- Runtime RTL для Link/preview URL extraction — нет (source-guard only); acceptable for P3 copy/docs.

## [MUTATION RESULT] — W13 slot 2 refill / fe1 · 2026-07-30

Evidence: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-fe1-task078.md`

| ID | Verdict | RED |
|----|---------|-----|
| W11-A4-017 | N/A | docs-only |
| W11-A5-020 | N/A | docs-only |
| W11-A5-021 | N/A | docs-only |
| W10-A2-021 | PROVEN | `security settings uses TanStack Link, not <a href>` |
| W10-A5-015 | PROVEN | `trash column is Удалено (date), not Удалил` |
| W10-A5-018 | PROVEN | `run-results-table has correct UTF-8 Исполнитель` |
| W10-A1-022 | PROVEN | `message-bubble extracts URLs from displayText and contentText` |

**4/4 FE PROVEN · 3/3 docs N/A** → статус **Закрыта**. UI restored. Коммит не создавался.
