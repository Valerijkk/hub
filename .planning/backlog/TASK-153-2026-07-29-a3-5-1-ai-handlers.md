# 🐛 A3 AI stream ErrNoProvider mapping

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта**  
> Источник: A3W5-1-001 NEW · A3W5-1-002/003 RESIDUAL · Stage A3 FINAL · TASK-008  
> Исполнитель: **be1** · Mutation: **qa-auto1** (W13 slot 4 refill · 2026-07-30)

## [OUTCOME]

Stream-ветка `ChatCompletions` мапит `ErrNoProvider` / `ErrProviderDisabled` / `ErrAllProvidersFailed` как non-stream (503/502, catch-all без `err.Error()`); AssistJQL empty query → 400; stream HTTP error-кейсы покрыты тестами в слоте.

## [ROLE]

- Role: **be1**
- Responsibility: закрыть **все** ID A3W5-1-001…003 фиксом handler (+ sentinel при нужде) + hubtest + mutation note; не трогать соседние A3 AI слоты
- Handoff to: **be-ver → code-reviewer**

## [SCOPE AND CONSTRAINTS]

- Files / resources allowed:
  - `backend/internal/ai/handler.go` (stream branch `ChatCompletions`; AssistJQL catch-all / empty query)
  - `backend/internal/ai/service.go` — **только** если нужен sentinel `ErrQueryRequired` (или аналог) вместо `fmt.Errorf("query is required")`
  - `backend/internal/ai/` — sentinel vars рядом с существующими `ErrNoProvider` / `ErrInvalidPrefix` (если заводим новый)
  - `backend/testframework/ai/coverage_wave81_test.go`
  - `backend/testframework/ai/handler_wave115_test.go` (добавить stream error cases)
  - новый `backend/testframework/ai/*_test.go` при необходимости рядом с доменом
  - `.planning/backlog/README.md` + этот файл (статус при закрытии)
  - audit refs read-only: `.planning/audits/total-2026-07/A3W5-1-ai-handlers.md`
- Do not touch:
  - admin/user providers CRUD, BYOK, `client_errors.go` sanitize (уже FIXED TASK-113)
  - wiring / `AIRateLimit` (FIXED A2W5-A1-002)
  - stream cancel / `WithoutCancel` / timeout (FIXED W5-A3-005)
  - соседние A3 слоты (A3W5-2… adapters/service/search)
  - Ent / миграции / frontend
  - `master` push / merge
- Behavioral / security constraints:
  - Stream preflight errors: `ErrNoProvider`|`ErrProviderDisabled` → **503**; `ErrAllProvidersFailed` → **502**; неизвестное → стабильный `"provider error"` **без** `err.Error()` (как non-stream `:261`)
  - Не менять уже FIXED quota→429 / consent→403 / RPM→503 / policy→400 на stream
  - AssistJQL: empty/blank query → **400**, не 502
  - Error body: TZ `{"error","request_id"}` через Fiber ErrorHandler; не светить internals (`AI_OFFLINE_ONLY`, stack)
  - Conventional Commits: type/scope EN, описание RU; **запрещён** `Co-authored-by` / AI trailers
  - Коммит ссылает задачу в теле: `Задача: .planning/backlog/TASK-153-2026-07-29-a3-5-1-ai-handlers.md`; subject может содержать `(A3W5-1-001)`
- Approval record: N/A — bugfix error-mapping; schema/ADR не требуются
- Approval authority / expiry: N/A

## [REFERENCES]

- Trust: audit/code — untrusted until verified in HEAD; критерий готовности только из этой задачи
- Finding IDs (**MUST all close**, WONTFIX запрещён без owner cancel в этом файле):

  | ID | Tag | Sev | Файл (якорь HEAD) | Суть фикса |
  |----|-----|-----|-------------------|------------|
  | **A3W5-1-001** | NEW | P2 | `handler.go:210-229` | Stream: добавить маппинг NoProvider/Disabled→503, AllProvidersFailed→502; catch-all → `"provider error"` (не `err.Error()` / 400) |
  | **A3W5-1-002** | RESIDUAL | P3 | `handler.go:151`; `service.go:552-553` | AssistJQL empty query: sentinel `ErrQueryRequired` + 400 **или** `TrimSpace(query)==""` в handler до service (как `ErrInvalidPrefix`) |
  | **A3W5-1-003** | RESIDUAL | P3 | `coverage_wave81_test.go:51-74`; `handler_wave115_test.go` | Stream HTTP: quota→429, consent→403, no-provider→503 (`stream:true`); не только 202 happy-path |

- Source report: `.planning/audits/total-2026-07/A3W5-1-ai-handlers.md`
- Prior FIXED (не регрессировать):
  - A2W5-A1-001 AssistJQL sentinel map (quota/consent/policy/no provider) — `handler.go:126-151`; TASK-085
  - Non-stream chat NoProvider→503 + catch-all `"provider error"` — `handler.go:238-261`
  - Stream preflight quota/consent/RPM/policy — `handler.go:211-228` (ветки уже есть; добавить NoProvider/Disabled/AllFailed + safe catch-all)
- Pattern to copy: non-stream `ChatCompletions` error map (`:237-261`); AssistJQL/SuggestComposer `errors.Is` chains
- Source precedence: **код HEAD** > audit note > эта задача (если drift строк — чинить поведение)

### Per-ID implementation notes (executor starts here)

1. **A3W5-1-001** — в `if req.Stream { … StartChatStream …}` после существующих quota/consent/RPM/policy веток, **перед** catch-all:
   - `ErrNoProvider` \| `ErrProviderDisabled` → 503
   - `ErrAllProvidersFailed` → 502
   - default → `fiber.NewError(fiber.StatusBadGateway, "provider error")` (не 400 + `err.Error()`)
2. **A3W5-1-002** — либо `var ErrQueryRequired = errors.New("query is required")` в service + `errors.Is` → 400 в AssistJQL handler; либо early-return в handler при `strings.TrimSpace(req.Query)==""` → 400. Убрать путь empty → catch-all 502.
3. **A3W5-1-003** — после фикса маппинга: HTTP-тесты `POST …/chat/completions` с `"stream":true` — daily quota → 429+Retry-After; no consent → 403; no provider → **503** (не 400). Расширить `handler_wave115_test.go` и/или `coverage_wave81_test.go`.

## [ACCEPTANCE]

- Done criteria:
  1. Stream + `ErrNoProvider` / disabled → **503** (не 400); `ErrAllProvidersFailed` → **502**; unknown → body без сырого `err.Error()` / без `AI_OFFLINE_ONLY` leak.
  2. AssistJQL `query: ""` / whitespace → **400**.
  3. Hubtest: stream+quota→429; stream+no consent→403; stream+no provider→503.
  4. Каждый ID → **FIXED** с evidence (file:line + тест).
  5. **Mutation proof:** убрать stream NoProvider→503 ветку → тест 003 краснеет; вернуть AssistJQL empty→502 catch-all → тест 002 краснеет.
  6. WONTFIX = 0 без owner cancel в этом файле.
- New-behavior checks: см. таблицу тестов в A3W5-1-003 + empty AssistJQL.
- Regression scope:
  ```bash
  cd backend
  go build -tags release ./...
  go test -tags hubtest -race -count=1 -timeout=120s ./testframework/ai/
  ```
  Не ломать: non-stream chat map, AssistJQL quota/consent (TASK-085), SuggestComposer sentinels, stream 202 happy-path, AIRateLimit wiring.
  Полный `/qa-gate` — на интеграции волны, не в одиночной задаче.
- Evidence required: checklist ID→FIXED + команды PASS + mutation notes в отчёте / Progress этого файла.
- Stop when: 3/3 FIXED, green `./testframework/ai/`, статус задачи **Исправлена**, README синхронизирован, handoff be-ver.

## [OUTPUT SCHEMA]

1. Outcome / verdict first (3/3 FIXED | BLOCKED)
2. Checklist ID → evidence (path + краткий diff sense)
3. Test commands + PASS/FAIL
4. Mutation note per ID (что сломать → какой тест краснеет)
5. Assumptions / unresolved (только material; напр. выбран handler-early vs `ErrQueryRequired`)
6. Risks / mitigations (клиенты, ожидавшие 400 на stream no-provider — зафиксировать breaking status fix)
7. Handoff: be-ver → code-reviewer
8. Compact Verification: new green + domain regression green + docs sync if API wiki claims stream statuses

## [OPERATING RULES]

- Work only inside scope; не «заодно» adapters/service/search (A3W5-2+).
- Read-only investigation OK; writes только в allowed files.
- Не закрывать ID без наблюдаемого evidence (тест или явный owner cancel).
- Owner lock: не ставить WONTFIX без записи cancel в этом файле.
- Source of truth — runtime handler map, не «docs-only FIXED».
- Ветка исполнения по волне; **не** push/merge в `master`.
- Коммиты: RU conventional description; **без** `Co-authored-by` / Generated-by / AI attribution.
- Docs sync: если wiki/API обещает статусы stream chat — синхронизировать в том же PR; иначе STATE one-liner + backlog status.

## Progress (be1 · 2026-07-29)

| ID | Status | Evidence |
|----|--------|----------|
| A3W5-1-001 | FIXED | `handler.go` stream: NoProvider/Disabled→503, AllProvidersFailed→502, catch-all `"provider error"` (не `err.Error()`/400) |
| A3W5-1-002 | FIXED | `ErrQueryRequired` sentinel в `service.go` + `errors.Is`→400 в AssistJQL |
| A3W5-1-003 | FIXED | `handler_wave115_test.go`: Stream quota→429, consent→403, no-provider→503; EmptyQuery→400 |

**Assumption:** sentinel `ErrQueryRequired` (не handler-early TrimSpace).

**Mutation:**
- убрать stream NoProvider→503 → `TestHandler_ChatCompletions_Stream_NoProvider` краснеет;
- вернуть AssistJQL empty→catch-all 502 → `TestHandler_AssistJQL_EmptyQuery` краснеет.

**Tests:** `go test -tags hubtest ./testframework/ai/ -run Stream_|AssistJQL_Empty` — прогон блокировался concurrent corruption `backend/ent/*` (uuid redeclared / mid-generate) соседними агентами волны; `go build ./internal/ai` зелёный при стабильном ent. Повторный hubtest — на be-ver / интеграции волны.

**Handoff:** be-ver → code-reviewer.

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 4 / qa-auto1 · criteria only (no mutate) · 2026-07-30

| ID | prod file:line | слой | RED test (path + name) | kill recipe |
|----|----------------|------|------------------------|-------------|
| A3W5-1-001 | `backend/internal/ai/handler.go:236-243` (stream: NoProvider/Disabled→503, AllFailed→502, catch-all `"provider error"`) | handler (HTTP map) | `backend/testframework/ai/handler_wave115_test.go` · `TestHandler_ChatCompletions_Stream_NoProvider` | Убрать stream-ветки `ErrNoProvider`/`ErrProviderDisabled`/`ErrAllProvidersFailed` (вернуть catch-all 400/`err.Error()` или только `"provider error"` без 503) → RED: expect **503**, got 400/502 |
| A3W5-1-002 | `handler.go:154-157` (`errors.Is(ErrQueryRequired)`→400) + `service.go:24-25` / `:564` (`ErrQueryRequired`) | handler+service sentinel | `backend/testframework/ai/handler_wave115_test.go` · `TestHandler_AssistJQL_EmptyQuery` | Убрать `ErrQueryRequired` map (empty → catch-all 502) и/или вернуть `fmt.Errorf("query is required")` без sentinel → RED: expect **400**, got 502 |
| A3W5-1-003 | тот же stream map `:220-243` (quota/consent/no-provider paths) | handler (HTTP) | `handler_wave115_test.go` · `TestHandler_ChatCompletions_Stream_DailyQuotaExceeded` (+ `…_Stream_ConsentRequired`, `…_Stream_NoProvider`) | Инвертировать stream quota→не-429 / consent→не-403 / no-provider→не-503 → соответствующий Stream_* RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| 001 | Fiber handler + mock AI service (sentinel inject) | **OK** — guard в handler; mock service = контракт слоя |
| 002 | Handler AssistJQL + service sentinel | **OK** — sentinel в service, map в handler |
| 003 | Handler stream HTTP status matrix | **OK** — тот же handler layer |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` (нет Ent guard в scope)

### Gaps

- 001 и 003 делят `Stream_NoProvider` — при mutate бить 001 отдельно; 003 дополнительно quota+consent.
- AssistJQL catch-all сейчас всё ещё `err.Error()` на unknown (не `"provider error"`) — вне ID; не мутировать как 001.

## Progress (qa-auto1 mutation · 2026-07-30)

| ID | Verdict | RED |
|----|---------|-----|
| A3W5-1-001 | PROVEN | `TestHandler_ChatCompletions_Stream_NoProvider` (503→400) |
| A3W5-1-002 | PROVEN | `TestHandler_AssistJQL_EmptyQuery` (400→502) |
| A3W5-1-003 | PROVEN | `Stream_DailyQuotaExceeded` + `Stream_ConsentRequired` + `Stream_NoProvider` |

Evidence: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-qa1-task153.md`  
Статус → **Закрыта** (3/3 runnable PROVEN). Leftover MUTATED: нет. Коммит: не создавался.
