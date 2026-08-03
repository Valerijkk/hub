# Phase 133 — Frontend units colocated (`src/`)

> Целевая раскладка и недельный порядок —
> [`active/august-2026/05-AUGUST-DETAILED.md`](../../active/august-2026/05-AUGUST-DETAILED.md).
> Юниты живут рядом с кодом, всё остальное — в корневой `testframework/`
> (решение владельца 2026-07-30).


**Status:** PLANNED  
**Migration stage:** Э4  
**Estimate:** 5–8 person-days  
**Branch (suggested):** `feat/qa/tf-frontend-units-colocated`

## Goal

Перенести Vitest-юниты из `testframework/stages/unit/frontend-test/domains/**`
в `frontend/src/<module>/**/*.test.ts(x)` (рядом с кодом). Сейчас юнитов рядом
с кодом **0** — завести конвенцию и перенос по доменам.

## Depends on

- **128** tf-python-scaffold (scaffold + Vitest tooling conventions) — soft-hard: начинать после каркаса `testframework/` и решения по colocated frontend layout.
- Нет hard-dep на **132** (можно параллелить после 128).
- Soft: стабильный frontend test script / link-testframework пока старые пути живы.

## Blocks (machine-readable)

- **BLOCKS:** Microservices split **Э3** (`hub-shell` + Module Federation / first front module) — см. [`.planning/active/august-2026/00-AUGUST-PLAN.md`](../../active/august-2026/00-AUGUST-PLAN.md) «Жёсткая зависимость» и [`01-MICROSERVICES-SPLIT.md`](../../active/august-2026/01-MICROSERVICES-SPLIT.md).
- Причина: `frontend/src/**/*.test.ts(x)` сейчас **0**; резать SPA без colocated unit-тестов = регрессии без сигнала.
## Deliverables

- [ ] Конвенция `__tests__/` vs `*.test.tsx` рядом с файлом (зафиксировать в PLAN/README)
- [ ] Перенос доменов: auth, admin, ai, tracker, v4, …
- [ ] `npm run test:run` green без зависимости от `testframework/stages/unit`
- [ ] Обновить docs (`test_plan` / TESTING) при cutover

## Out of scope

- E2E
- Удаление root `testframework/` (136)

## Success criteria

- [ ] Vitest 99/99 (или актуальный счёт) из `frontend/src`
- [ ] `testframework/stages/unit/frontend-test` опустошён

## Risks

| Risk | Mitigation |
|------|------------|
| Shared `render.tsx` helpers | Перенести в `frontend/src/test-utils/` |
| Path imports break | Codemod + tsc |
