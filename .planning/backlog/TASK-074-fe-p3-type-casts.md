# TASK-074 — FE unsafe type casts

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Закрыта
> Источник: TASK-008 W12 · Classes: P3-FE-07 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-FRONTEND.md`
>
> **C mutation (2026-07-30, fe2 W13 slot 5 retry):** 8/8 PROVEN · W9-A4-014 GAP_NOTED → **Закрыта**.  
> Evidence: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-fe2-task074-retry.md`  
> Prior NOT_PROVEN: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-fe2-task074.md`

## [OUTCOME]

Дефекты P3-FE-07 закрыты.

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

- Реестр W12 (P3-FE-07)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-FE-07 с мутацией «да» закрыты по реестру
- **Мутация (C):** unsafe cast → type guard; omit → TS fail
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

Проверено по коду HEAD (2026-07-30, fe2 W13 slot 5 criteria-only). Слой **frontend** (Class **P3-FE-07**). Typecheck: `cd frontend && npx tsc --noEmit` — ловит omit `PlateDoc` / снятие generics; runtime casts — Vitest.

| ID | file:line | Vitest path + `it` | kill recipe |
|----|-----------|----------------------|-------------|
| W9-A3-020 | `frontend/src/lib/plate-doc.ts:17-23` + `message-composer.tsx:317` / `:452` — `plateDocToContentJson(PlateDoc)` без `as unknown as Record` | `domains/platform/type-casts-p3.test.ts` · describe `P3-FE-07 type casts (TASK-074)` > `it("plateDocToContentJson requires PlateDoc fields (omit type → TS fail)")` + `it("message-composer has no as unknown as Record double-cast")` | вернуть `as unknown as Record<string, unknown>` / убрать `plateDocToContentJson` → source-guard FAIL; omit `type` на `PlateDoc` → `expectTypeOf`/tsc FAIL |
| W9-A3-021 | `frontend/src/components/messenger/sticker-picker.tsx:145` / `:156` — `Array.isArray(children)` guard | same file > `it("sticker/render avoid bare children as unknown[] cast")` | вернуть `children as unknown[]` без `Array.isArray` → regex/guard asserts FAIL |
| W9-A3-022 | `frontend/src/components/messenger/render-content.tsx:148` — `const nodes: unknown[] = contentJson` | same `it` | вернуть `contentJson as unknown[]` после `Array.isArray` → `const nodes: unknown[] = contentJson` FAIL |
| W9-A2-016 | `frontend/src/components/dashboard/widget-data.ts:4-19` — `DashboardWidgetData` + `isPersonalData`/`isProjectStats`/`isAdminStats` | same file > `it("widget data guards narrow DashboardWidgetData")` | убрать guards / вернуть `data as ProjectStats` в widgets → `isPersonalData` asserts FAIL |
| W10-A2-026 | `frontend/src/lib/format-env-config.ts:2-12` — `formatEnvConfigValue` + `JSON.stringify` | same file > `it("formatEnvConfigValue avoids [object Object]")` | `${v}` / `String(obj)` без stringify → `{"browser":"ff"}` FAIL |
| W11-A2-014 | `frontend/src/api/marketplace.ts:62` / `:68` / `:74` — `api.get<{ items: MarketplaceListing… }>` | same file > `it("marketplace list calls use explicit generics")` | снять `<T>` с `api.get` (оставить только return-annotation) → `expectTypeOf`/source `api.get<{ items: MarketplaceListing` FAIL |
| W11-A1-010 | `frontend/src/api/ai-types.ts:20-28` — `AdminAIProvider` extends base; `MyAIProvider = AIProviderBase` | same file > `it("AdminAIProvider extends base; MyAIProvider is distinct shape")` + `readSrc("api/ai-types.ts")` | смержить типы (`My=Admin`) / `priority` на base / `priority` только на My → source-guard FAIL |
| W11-A4-016 | `frontend/src/lib/format-api-error.ts:16-22` — `extractApiError` требует `code` number \|\| `request_id` string | same file > `it("extractApiError does not treat plain Error as ApiError")` | любой object с `message` → `as ApiError` → `formatApiError(new Error(…))` ветка/`typeof candidate.code` FAIL |

### Gaps (mutation-proof)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| PlateDoc / sticker / render / widgets / env / marketplace / ai-types / extractApiError | ✅ Vitest (+ expectTypeOf) | **8/9** Class source IDs |
| W9-A4-014 (`comment-list` `bodyJsonToPlateValue`, `jql-search-page` `jqlFromFilterConfig`, `automation-rules-panel` typed `getChannels`) | ❌ GAP | FIXED в коде; **нет** dedicated `it` в `type-casts-p3.test.ts` |
| Docs-only | N/A | нет |

**Confidence:** HIGH для 8 runnable; W9-A4-014 — **GAP_NOTED** (criteria refill only, no mutate).
