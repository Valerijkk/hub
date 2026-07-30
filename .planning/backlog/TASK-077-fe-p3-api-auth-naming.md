# TASK-077 — FE ApiError/auth broadcast/naming

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Закрыта
> Источник: TASK-008 W12 · Classes: P3-FE-10,P3-FE-11,P3-FE-12 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-FRONTEND.md`
> Mutation C: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-fe2-task077.md` · **6/6 PROVEN** · fe2 · 2026-07-30

## [OUTCOME]

Дефекты P3-FE-10..12 закрыты.

## [ROLE]

- Role: fe1
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: frontend/src/{api,hooks,stores}/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P3-FE-10,P3-FE-11,P3-FE-12)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-FE-10,P3-FE-11,P3-FE-12 с мутацией «да» закрыты по реестру
- **Мутация (C):** ApiError parse; auth broadcast once; naming dup gone
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

> W13 slot 1 refill / be1 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified · Classes **P3-FE-10**, **P3-FE-11**, **P3-FE-12**

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| W11-A1-008 | `frontend/src/api/calendar.ts:38-40` — `downloadCalendarExport` через `api.getBlob` (ApiError path) | api client | `getBlob: maps wire error→message and request_id` (+ describe `downloadCalendarExport via api.getBlob`) · `testframework/stages/unit/frontend-test/domains/auth/client-api.test.ts` | `throw new Error("export failed")` / raw fetch без getBlob → request_id/code toast path FAIL |
| W11-A1-009 | `frontend/src/api/client.ts:173-180` — 2xx `JSON.parse` try/catch → `ApiError` code 502 | api client | `it("2xx invalid JSON throws ApiError code 502")` · `testframework/stages/unit/frontend-test/domains/platform/api-auth-naming-p3.test.ts` | убрать try/catch → сырой SyntaxError / нет code 502 → toMatchObject RED |
| W11-A2-011 | `frontend/src/stores/auth-store.ts:155-167` — `setTokens`/`setUser` broadcast с `userId` | store | `it("auth-store broadcasts userId from setTokens/setUser")` · `…/platform/api-auth-naming-p3.test.ts` | всегда `userId: null` / убрать setUser broadcast → source-guard FAIL |
| W11-A2-012 | `frontend/src/stores/auth-store.ts:67-82` — `broadcastAuth` mirror localStorage `hub-auth-event` | store | `it("broadcastAuth mirrors login to localStorage key")` · same file | убрать setItem/removeItem mirror → setSpy not called RED |
| W11-A2-013 | `frontend/src/api/messenger.ts:484-490` канон; `github.ts` без `/link-preview` | api modules | `it("github.ts no longer owns /link-preview client")`; `it("messenger owns fetchLinkPreview / getLinkPreview alias")` · same file | вернуть дубль в github.ts / убрать messenger export → Match/not Match RED |
| W11-A3-017 | `frontend/src/lib/composer-ai-suggest.ts` — helpers не в `hooks/use-*` | lib naming | `it("composer helpers live in lib/, not hooks/use-*")` · same file | вернуть `hooks/use-composer-ai-suggest.ts` / удалить lib → toThrow/Match RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| W11-A1-008 / W11-A1-009 | api client (fetch stub) | **OK** — FE, не Ent |
| W11-A2-011 / W11-A2-012 | auth-store runtime + source | **OK** |
| W11-A2-013 / W11-A3-017 | source-guard + lib import | **OK** |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)`

### Gaps

- Docs-only: **N/A**.
- Class source IDs = **6/6** с dedicated RED (W11-A1-008 в `client-api.test.ts`, остальное в `api-auth-naming-p3.test.ts`).
