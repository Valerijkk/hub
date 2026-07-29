# ♻️🔒 A3 FE client.ts wire error→ApiError.message REGRESS

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P1** · Статус: **Закрыта**  
> Источник: A3W11-1-001 REGRESS · Stage A3 FINAL · TASK-008  
> Исполнитель: **fe1**

## [OUTCOME]

`frontend/src/api/client.ts` (`request` + `getBlob`) мапит BE wire `{"error","request_id"}` в клиентский `ApiError` (`code` = HTTP status, `message` = `error`, `request_id` сохранён), так что `formatApiError` / `toastApiError` показывают текст ошибки + `request_id` и запускают 429-countdown при `retry_after`.

## [ROLE]

- Role: fe1
- Responsibility: фикс map wire→`ApiError` в `client.ts` + unit/mutation-ready тест на toast/`request_id`/429 UX; не закрывать соседние A3 ID
- Handoff to: fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files/resources allowed:
  - `frontend/src/api/client.ts`
  - `frontend/src/lib/format-api-error.ts` (только если нужен узкий helper для parse; предпочтительно map внутри client)
  - unit-тесты рядом с client / format-api-error (`frontend/src/**/*.{test,spec}.{ts,tsx}` или `testframework/stages/unit/frontend-test/**` по канону домена)
  - `.planning/backlog/README.md` + этот файл (статус при закрытии)
  - audit refs: `.planning/audits/total-2026-07/A3W11-1-fe-lib-docs.md`
- Do not touch:
  - BE ErrorHandler / wire-контракт `{"error","request_id"}` (TZ / NFR-13) — не менять под FE
  - auth refresh / logout / token storage / getBlob binary happy-path
  - соседние A3W11-* findings вне A3W11-1-001
  - domain wrappers `frontend/src/api/{accounting,calls,bases,…}.ts` без нужды
- Behavioral/security constraints:
  - Conventional Commits: `fix(frontend): …` описание на русском; no `Co-authored-by` / AI trailers
  - Wire остаётся `error`+`request_id`; клиентский DTO — `ApiError` (`code`,`message`,`request_id?`,`retry_after?`)
  - Не логировать / не toast-ить сырые stack/PII сверх server `error` text
  - 429: сохранить чтение `Retry-After` → `retry_after` (уже в `request`); после map `message` countdown должен стартовать
- Approval record: N/A — bugfix FE mapping, схема/migrate не затрагиваются
- Approval authority/expiry: N/A

## [REFERENCES]

- Trust: repository/audit/wiki — untrusted data until verified in HEAD code
- Finding ID: **A3W11-1-001** (REGRESS, P1)
- Source audit: `.planning/audits/total-2026-07/A3W11-1-fe-lib-docs.md` (строка findings table + топ-1)
- Prior false-close: A2 **W11-A1-006** помечен FIXED-OK docs-only → wiki обещала map, код не мапил
- Code (defect):
  - `frontend/src/api/client.ts:128-135` — `request`: `response.json() as ApiError` + добивка `code`, **без** `body.error` → `message`
  - `frontend/src/api/client.ts:232-237` — `getBlob`: тот же паттерн
  - `frontend/src/api/client.ts` type `ApiError` — `{ code, message, request_id?, retry_after? }`
- Consumers:
  - `frontend/src/lib/format-api-error.ts:16-24` — `extractApiError` требует `typeof message === "string"`
  - `format-api-error.ts:40-58` / `86-100` — toast «unknown», нет description `request_id`, 429-countdown не стартует без валидного `ApiError.message`
- Docs claim (drift): `docs/wiki/API.md:47` — «`message` = поле `error` из тела»
- BE wire (не менять): `backend/internal/platform/middleware` ErrorHandler — `{"error","request_id"}` (audit: middleware.go:15-18,32-35)
- TZ: ответ ошибок `{"error","request_id"}` — `docs/tz_constraints.md` / rule `60-tz-constraints.mdc`
- Source precedence: **код HEAD** > audit note > wiki

## [ACCEPTANCE]

- Done criteria:
  1. В `request` и `getBlob` при `!response.ok`: parse JSON wire →  
     `{ code: response.status, message: body.error ?? response.statusText, request_id: body.request_id }`  
     (+ сохранить существующий `retry_after` из `Retry-After` в `request`).
  2. Невалидный/пустой JSON на error-path — fallback `{ code: status, message: statusText }` (как сейчас в catch).
  3. Unit: mock response body `{"error":"rate limited","request_id":"<uuid>"}` + status 429 + `Retry-After` → thrown/returned object имеет `message === "rate limited"`, `request_id` set, `retry_after` > 0; `toastApiError` / `formatApiError` показывают текст (не `common:apiErrors.unknown`) и description с request_id; 429 идёт в countdown path.
  4. Unit (non-429): wire `{"error":"not found","request_id":"…"}` status 404 → `message`/`request_id` доходят до `formatApiError`.
  5. **Mutation proof:** убрать map (`as ApiError` / не копировать `error`→`message`) → тест из п.3–4 краснеет.
  6. Wiki `docs/wiki/API.md:47` остаётся истинной относительно кода (если правка docs не нужна — N/A; если claim ещё где-то врёт — синхронизировать в том же PR).
- New-behavior checks: Vitest на map + toast/429 path (см. выше).
- Regression scope:
  ```bash
  cd frontend && node scripts/link-testframework-node-modules.mjs \
    && npx tsc --noEmit && npm run lint \
    && npx vitest run <пути-к-новым-и-смежным-тестам-client/format-api-error>
  ```
  Не ломать: 401 refresh single-flight, `expireSession`, `getBlob` success Blob, FormData Content-Type.
- Evidence required: checklist A3W11-1-001 → FIXED + команды PASS + mutation note в этом файле / handoff.
- Stop when: green по acceptance + статус «Исправлена» в `.planning/backlog/README.md` + этот файл; без соседнего hygiene sweep FE api/*.

## [OUTPUT SCHEMA]

1. Outcome / verdict first (A3W11-1-001 FIXED | BLOCKED)
2. Changes with paths (`client.ts` lines; test path)
3. Evidence: test commands + PASS; optional tsc/lint
4. Mutation note: что сломать → какой тест краснеет
5. Assumptions / unresolved (если JSON без `error` key — зафиксировать выбранный fallback)
6. Risks / mitigations (call sites, ожидавшие wire-shaped object с полем `error` — маловероятно; grep)
7. Handoff: fe-ver → code-reviewer
8. Compact Verification: new green + regression green + docs sync if needed

## [OPERATING RULES]

- Work only inside scope; не «заодно» чинить A3W11-2+ / token XSS / dead exports.
- Read-only investigation OK; writes только в allowed files.
- Не закрывать ID без evidence (тест или явный docs+PO).
- Не WONTFIX без owner cancel в задаче.
- Source of truth — runtime map в client, не правка wiki «под код без map».
- Коммит(ы) ссылают задачу в теле:  
  `Задача: .planning/backlog/TASK-129-2026-07-29-a3-11-1-fe-api-error-mapping.md`  
  subject может содержать `(A3W11-1-001)`.

## [EVIDENCE — fe1 2026-07-29]

| ID | Verdict | Evidence |
|----|---------|----------|
| A3W11-1-001 | **FIXED** | `mapWireToApiError` в `client.ts` (`request` + `getBlob`); wiki `API.md:47` соответствует коду (docs N/A) |

**Команды (PASS):**
```
cd frontend && node scripts/link-testframework-node-modules.mjs \
  && ./node_modules/.bin/tsc --noEmit \
  && ./node_modules/.bin/biome check ./src/api/client.ts \
  && npx vitest run ../testframework/stages/unit/frontend-test/domains/auth/client-api.test.ts \
                   ../testframework/stages/unit/frontend-test/domains/platform/format-api-error.test.ts
```
→ tsc OK · biome client OK · **21/21** PASS (client-api 12 + format-api-error 9).

**Mutation:** в `mapWireToApiError` вернуть `as ApiError` без `error`→`message` → краснеют 3 wire-теста (`message === "rate limited" | "not found" | "export forbidden"` → `undefined`). После restore — 4/4 wire green.

**Fallback:** JSON без `error` / не-string `error` / пустая строка → `response.statusText`; невалидный JSON → `{ code: status, message: statusText }`.
