# TASK-071 — FE tokens + i18n primitives

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Закрыта
> Источник: TASK-008 W12 · Classes: P3-FE-03,P3-FE-04 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-FRONTEND.md`
> Mutation C: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-fe1-task071.md` · **9/9 PROVEN** (2026-07-30, fe1 W13 slot 2)

## [OUTCOME]

Дефекты P3-FE-03/04 закрыты.

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

- Реестр W12 (P3-FE-03,P3-FE-04)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-FE-03,P3-FE-04 с мутацией «да» закрыты по реестру
- **Мутация (C):** raw color → --hub-*; hardcoded RU → i18n key
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

Проверено по коду HEAD (2026-07-30, fe1 W13 slot 2). Слой **frontend** (Classes **P3-FE-03**, **P3-FE-04**). Typecheck: `cd frontend && npx tsc --noEmit` — не ловит className/i18n strings.

| ID | file:line | Vitest path + `it` | kill recipe |
|----|-----------|----------------------|-------------|
| W9-A5-018 | `frontend/src/components/ui/select.tsx:20` — SelectTrigger `rounded-[var(--r2)]` / `border-b1` | `domains/platform/tokens-i18n-p3.test.tsx` · describe `P3-FE-03/04 · tokens + i18n primitives (TASK-071)` > `it("SelectTrigger uses hub radius/border tokens (not shadcn rounded-md/border-input)")` | вернуть `rounded-md` / `border-input` → className asserts FAIL |
| W9-A3-031 | `frontend/src/components/messenger/voice-recorder.tsx:186` — `bg-[var(--red)]` | same file > `it("messenger residual files have no raw Tailwind palette colors")` | `bg-red-500` (или любой `bg/text/border-(red\|orange\|yellow\|green)-\d`) → palette regex FAIL |
| W9-A3-032 | `frontend/src/components/messenger/poll-message.tsx:175` — `text-[var(--orange)]` | same `it` | raw `text-orange-500` → same regex FAIL |
| W9-A3-033 | `frontend/src/components/messenger/channel-list.tsx:561` — `border-l-[var(--green)]` / `bg-[var(--gG)]` | same `it` | raw `border-green-500` / `bg-green-*` → same regex FAIL |
| W9-A3-034 | `frontend/src/components/messenger/thread-issue-links.tsx:85` — `hover:text-[var(--red)]` | same `it` | raw `hover:text-red-500` → same regex FAIL |
| W9-A3-035 | `frontend/src/components/messenger/invite-link-dialog.tsx:261` — `bg-[var(--orange)]/10` / `text-[var(--orange)]` | same `it` | raw `bg-orange-500` / `text-orange-*` → same regex FAIL |
| W9-A5-009 | `frontend/src/components/ui/dialog.tsx:36` — `t("btn.close")` в sr-only | same file > `it("Dialog Close uses i18n btn.close (RU)")`; `it("Dialog Close switches to EN Close via i18n")` | hardcode `Close` / `Закрыть` без `t(...)` → locale switch assert FAIL |
| W9-A5-010 | `frontend/src/components/ui/command.tsx:38` — `t("search.globalSearch")` | same file > `it("CommandDialog title follows locale")` | hardcode «Глобальный поиск» → EN «Global search» FAIL |
| W9-A5-011 | `frontend/src/components/ui/error-boundary.tsx:26` / `:42` — `t("error.somethingWrong")` / `t("btn.refresh")` | same file > `it("ErrorBoundary fallback uses i18n (EN locale)")` | hardcode RU-only → EN «Something went wrong» / «Refresh» FAIL |

### Gaps (mutation-proof)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| Select hub tokens | ✅ RTL className | W9-A5-018 |
| messenger palette (5 files) | ✅ source-guard `readFileSync` | W9-A3-031…035 |
| Dialog / Command / EB i18n | ✅ RTL + i18n.changeLanguage | W9-A5-009…011 |
| `npx tsc --noEmit` | N/A | не primary |

**Confidence:** CONFIRMED — 9/9 Class source IDs runnable.
