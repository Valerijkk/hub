# TASK-076 — FE perf stale intervals

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Закрыта
> Mutation C (2026-07-30, fe2 W13 slot 5): **3/3 PROVEN** — `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-fe2-task076.md`
> Источник: TASK-008 W12 · Classes: P3-FE-09 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-FRONTEND.md`

## [OUTCOME]

Дефекты P3-FE-09 закрыты.

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

- Реестр W12 (P3-FE-09)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-FE-09 с мутацией «да» закрыты по реестру
- **Мутация (C):** interval cleanup; Set hoist
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

Проверено по коду HEAD (2026-07-30, fe2 W13 slot 5 criteria-only). Слой **frontend** (Class **P3-FE-09**). Typecheck: `cd frontend && npx tsc --noEmit` — не ловит interval/Set hoist.

| ID | file:line | Vitest path + `it` | kill recipe |
|----|-----------|----------------------|-------------|
| W10-A3-017 | `frontend/src/modules/tracker/pipeline-status.tsx:202-208` — `refetchInterval: (query) =>` active only `queued`/`in_progress` | `domains/platform/perf-stale-p3.test.ts` · describe `TASK-076 P3-FE-09 perf / stale clock / Set hoist` > `it("pipeline refetchInterval only when queued/in_progress")` | вернуть `refetchInterval: 30_000` всегда → `not.toMatch(/refetchInterval:\s*30_000/)` + conditional asserts FAIL |
| W10-A3-020 | `frontend/src/routes/_authenticated/dashboard.lazy.tsx:70-73` — `useState(formatTime)` + `setInterval` → `setCurrentTime` | same file > `it("dashboard clock updates via interval, not frozen useMemo")` | вернуть `useMemo(() => formatTime(), [])` без interval → source-guard FAIL |
| W10-A5-019 | `frontend/src/routes/_authenticated/projects/$projectId/tms/runs_.$runId.tsx:313-317` — `const existingIds = new Set(…)` **до** `.map` | same file > `it("run add-cases dialog hoists existingIds Set outside map")` | создать `new Set` внутри `.map` callback → hoist-before-map assert FAIL |

### Gaps (mutation-proof)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| pipeline / clock / Set hoist | ✅ source-guard | **3/3** Class source IDs |
| Docs-only | N/A | нет |
| `npx tsc --noEmit` | N/A | не primary |

**Confidence:** HIGH — **3/3** runnable criteria (нет GAP).
