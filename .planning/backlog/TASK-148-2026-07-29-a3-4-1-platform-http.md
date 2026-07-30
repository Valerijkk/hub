# 🧹 A3 platform Request-ID mutation test

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P3** · Статус: **Закрыта**  
> Источник: A3W4-1-001 NEW · Stage A3 FINAL · TASK-008  
> Исполнитель: **be1** · Mutation checker: **be2** (W13 slot 3 refill)  
> Mutation: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-be2-task148.md` · **1/1 PROVEN** (2026-07-30)

## [OUTCOME]

В `backend/testframework/platform/middleware/middleware_test.go` есть mutation-ready тесты reject невалидного и oversized `X-Request-ID`, так что удаление `validRequestID` / алфавита / `maxRequestIDLen` в `requestid.go` краснеет в слоте A1 (регресс W4-A1-002 CopyString+sanitize ловится).

## [ROLE]

- Role: be1
- Responsibility: добавить `TestRequestID_RejectsInvalid` / `RejectsOversized` (или table-driven) + убедиться, что mutation proof работает; prod-логику `requestid.go` не менять без нужды (gap — в тестах)
- Handoff to: be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files/resources allowed:
  - `backend/testframework/platform/middleware/middleware_test.go` (основной слот)
  - `backend/internal/platform/middleware/requestid.go` — **только** если тест вскроет реальный баг reject (не «заодно» рефакторить)
  - `backend/internal/platform/middleware/export_hooks.go` — только если нужен test hook для `validRequestID` (предпочтительно тестировать через Fiber handler, без экспорта)
  - `.planning/backlog/README.md` + этот файл (статус при закрытии)
  - audit refs: `.planning/audits/total-2026-07/A3W4-1-platform-http.md`
- Do not touch:
  - CORS / rate limit / ErrorHandler / security_headers / logging / safedialer / opengraph
  - соседние A3W4-* findings вне A3W4-1-001
  - Caddy HSTS / wiring_fiber порядок middleware
  - клиентский FE `request_id` mapping
- Behavioral/security constraints:
  - Conventional Commits: `test(platform): …` (или `fix(platform): …` если найден баг) — описание на русском; no `Co-authored-by` / AI trailers
  - Невалидный/oversized client `X-Request-ID` → сгенерированный UUID ≠ client value; Locals + response header совпадают с сгенерированным
  - Safe client ID (алфавит `[A-Za-z0-9._-]`, ≤128 рун / UUID) по-прежнему сохраняется (`TestRequestID_PreservesExistingHeader` не ломать)
  - Не логировать сырой невалидный header как PII dump сверх существующих middleware логов
- Approval record: N/A — hygiene test gap; схема/migrate не затрагиваются
- Approval authority/expiry: N/A

## [REFERENCES]

- Trust: repository/audit — untrusted data until verified in HEAD code
- Finding ID: **A3W4-1-001** (NEW, P3)
- Source audit: `.planning/audits/total-2026-07/A3W4-1-platform-http.md` (findings table + топ-1)
- Prior FIXED-OK (не трогать / не regress): W4-A1-002 RequestID CopyString+алфавит ≤128; A2W4-1-001…003
- Code (prod — эталон поведения):
  - `backend/internal/platform/middleware/requestid.go:12-16` — `maxRequestIDLen = 128`, `requestIDPattern`
  - `requestid.go:21-32` — `RequestID()`: CopyString → `validRequestID` else `uuid.NewString()`; Locals + `c.Set`
  - `requestid.go:35-47` — `validRequestID`: empty / rune len >128 / UUID parse / pattern
- Code (test gap):
  - `backend/testframework/platform/middleware/middleware_test.go:29-50` — `TestRequestID_GeneratesNewID`
  - `middleware_test.go:52-69` — `TestRequestID_PreservesExistingHeader` только safe `my-custom-request-id-123`
  - отсутствует кейс reject пробелов / `<>` / len>128
- Related: Pitfall fasthttp buffer reuse — CopyString обязателен (уже в prod; mutation не должна убирать CopyString без красного теста — опционально покрыть, если удобно в том же PR)
- Source precedence: **код HEAD** > audit note

## [ACCEPTANCE]

- Done criteria:
  1. Тест(ы) с невалидным header: пробелы и/или `<>` (и при желании другие вне алфавита) → body Locals / `X-Request-ID` response — валидный UUID, **не** равен client value.
  2. Тест oversized: header длиной >128 рун (предпочтительно rune-safe, не только байты) → тот же reject-поведение (UUID ≠ client).
  3. Существующие `TestRequestID_GeneratesNewID`, `PreservesExistingHeader`, `UniquePerRequest` — зелёные.
  4. **Mutation proof:** убрать/ослабить `validRequestID` (всегда `return true`) **или** убрать проверку `maxRequestIDLen` / pattern — новый тест краснеет.
  5. Prod `requestid.go` без косметического рефактора; менять только если тест доказал баг.
- New-behavior checks: новые тесты в `middleware_test.go` (см. выше).
- Regression scope:
  ```bash
  cd backend && go test -tags hubtest -race -count=1 \
    ./testframework/platform/middleware/ \
    -run 'TestRequestID_'
  ```
  При полном доменном прогоне: `./testframework/platform/...` — не ломать CORS/RL/ErrorHandler кейсы в том же пакете.
- Evidence required: checklist A3W4-1-001 → FIXED + команды PASS + mutation note в этом файле / handoff.
- Stop when: green по acceptance + статус «Исправлена» в `.planning/backlog/README.md` + этот файл; без соседнего hygiene sweep platform middleware.

## [OUTPUT SCHEMA]

1. Outcome / verdict first (A3W4-1-001 FIXED | BLOCKED)
2. Changes with paths (`middleware_test.go` lines; `requestid.go` только если touch)
3. Evidence: test commands + PASS
4. Mutation note: что сломать в `validRequestID` / pattern / maxLen → какой тест краснеет
5. Assumptions / unresolved (N/A если только тесты)
6. Risks / mitigations (flaky UUID assert — сравнивать ≠ client + `uuid.Parse`, не фиксированный expected UUID)
7. Handoff: be-ver → code-reviewer
8. Compact Verification: new green + domain RequestID suite green

## [OPERATING RULES]

- Work only inside scope; не «заодно» чинить ratelimit Lua / HSTS / ParseUUID.
- Read-only investigation OK; writes только в allowed files.
- Не закрывать ID без evidence (тест + mutation note).
- Не WONTFIX без owner cancel в задаче.
- Source of truth — runtime `RequestID()` + `validRequestID`, не правка audit без кода/теста.
- Коммит(ы) ссылают задачу в теле:  
  `Задача: .planning/backlog/TASK-148-2026-07-29-a3-4-1-platform-http.md`  
  subject может содержать `(A3W4-1-001)`.

## [EVIDENCE — be1 2026-07-29]

- **Verdict:** A3W4-1-001 FIXED (на проверке be-ver)
- **Changes:** `middleware_test.go` — `TestRequestID_RejectsInvalid` (spaces / `<>` / slash) + `TestRequestID_RejectsOversized` (ascii 129 + rune 129 рун). `requestid.go` не менялся.
- **PASS:**
  ```
  go test -tags hubtest -race -count=1 ./testframework/platform/middleware/ -run 'TestRequestID_'
  go test -tags hubtest -race -count=1 ./testframework/platform/middleware/
  ```
  оба ok.
- **Mutation:** `validRequestID` → `return true` → краснеют `TestRequestID_RejectsInvalid/*` и `TestRequestID_RejectsOversized/*` (UUID parse + NotEqual client). Revert подтверждён re-run green.
- **Handoff:** be-ver → code-reviewer

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 3 / be2 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| A3W4-1-001 | `backend/internal/platform/middleware/requestid.go:36-46` `validRequestID` (empty / `maxRequestIDLen=128` `:12-13` / `requestIDPattern` `:16`); call-site `:24-26` | middleware | `TestRequestID_RejectsInvalid`; `TestRequestID_RejectsOversized` · `backend/testframework/platform/middleware/middleware_test.go` | `validRequestID` → `return true` и/или убрать maxLen/pattern → UUID≠client asserts RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Комментарий |
|----|-----------|-------------|
| 001 | Fiber `RequestID()` middleware | Guard в middleware — Ent N/A |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)`

### Gaps

- `TestRequestID_RejectsOversizedHeader_W4A1002` (task043) — prior W4 coverage; W13 kill target — RejectsInvalid/Oversized из TASK-148.
- CopyString mutation (Pitfall 1) — optional adjacent; не в A3W4-1-001 acceptance.

## [EVIDENCE — be2 mutation W13 refill 2026-07-30]

**1/1 PROVEN** · report: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-be2-task148.md` · leftover MUTATED: нет · TASK → **Закрыта**

| ID | Verdict | RED |
|----|---------|-----|
| A3W4-1-001 | PROVEN | `TestRequestID_RejectsInvalid`; `TestRequestID_RejectsOversized` |

Kill: `validRequestID` → `return true` → UUID≠client asserts RED; restore → GREEN.
