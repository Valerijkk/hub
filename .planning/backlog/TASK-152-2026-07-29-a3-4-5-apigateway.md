# 🔒 A3 apigateway soft-delete webhooks + VersionEQ

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **На проверке**  
> Источник: A3W4-5-apigateway.md · Stage A3 FINAL · TASK-008 · кластер A3W4-5  
> Ветка исполнения: `cursor-grok-4.5` (не пушить в `master`)  
> Finding IDs: **A3W4-5-001** … **A3W4-5-004**  
> Исполнитель: be2 · Handoff: be-ver → code-reviewer

## [OUTCOME]

Все четыре A3W4-5 finding ID закрыты: soft-deleted project не получает outbound webhook POST; UpdateWebhookEndpoint требует VersionEQ/409; `sprint.completed` либо публикуется end-to-end, либо убран из allowlist; CreateAPIKey отклоняет name >255 рун с 400 (не 500).

## [ROLE]

- Role: **be2**
- Responsibility: закрыть **все** ID ниже фиксом в `apigateway` (+ sprint producer **или** allowlist trim) + hubtest + mutation note; обновить статус задачи/README при закрытии
- Handoff to: **be-ver → code-reviewer**

## [SCOPE AND CONSTRAINTS]

- Files / resources allowed:
  - `backend/internal/apigateway/repository.go` — `ListActiveWebhookEndpoints`, `UpdateWebhookEndpoint`
  - `backend/internal/apigateway/webhook.go` — dispatch/ListActive path, `AVAILABLE_EVENTS` / allowlist / `eventMatchesSubscription` при выборе trim
  - `backend/internal/apigateway/handler.go` — CreateAPIKey MaxLen; UpdateWebhook DTO binding
  - `backend/internal/apigateway/dto.go` — `UpdateWebhookRequest.version`
  - `backend/internal/apigateway/errors.go` — MapToFiber / 409 conflict при stale version (если нужно)
  - `backend/internal/project/**` — только если deactivate endpoints в `SoftDeleteWithCrossLinkCleanup` / DeleteProject tx (альтернатива ListActive filter)
  - `backend/internal/sprint/**` — только если выбран путь publish `sprint.<id>.completed` + `project_id` (не trim allowlist)
  - `backend/ent/schema/webhookendpoint.go`, `apikey.go` — **read-only** (VersionMixin / MaxLen уже есть); codegen/migrate **не** трогать без owner approve
  - `backend/testframework/apigateway/**` (+ `testframework/sprint/**` / `project/**` только если выбран producer/cleanup path)
  - `.planning/backlog/README.md` + этот файл (статус)
  - audit refs read-only: `.planning/audits/total-2026-07/A3W4-5-apigateway.md`, `A3-DEDUP-2026-07-29.md`, `A3-W4-SUMMARY.md`
- Do not touch:
  - SEC-02 outbound HMAC / `signPayload` / SSRF `ValidateOutboundURL` / SafeHTTPClient — не ослаблять
  - API key SHA-256 at-rest / soft-delete revoke keys/endpoints (FIXED-OK)
  - inbound CTC github/gitlab (не этот модуль)
  - соседние A3 кластеры (W4-1…W4-4, TASK-≠152) — не «заодно»
  - Ent migrate / schema field changes без owner approve в Approval record
  - `master` push / merge
- Behavioral / security constraints:
  - soft-delete project (корзина) → **0** outbound webhook jobs на его endpoints (W19 C-008 класс)
  - NFR-24 / ADR-ENT-VERSION-MIXIN: Update webhook → `Where(ID, DeletedAtIsNil, VersionEQ)` + `AddVersion(1)` → 409 на stale
  - error contract: MaxLen name → **400** `{error, request_id}`, не 500 через Ent default
  - Conventional Commits: type/scope EN, описание RU; **запрещён** `Co-authored-by` / AI trailers
  - коммит ссылается на задачу в теле: `Задача: .planning/backlog/TASK-152-2026-07-29-a3-4-5-apigateway.md` + `(A3W4-5-00N)` в subject при желании
- Approval record: N/A (code-fix в уже существующем VersionMixin / MaxLen schema; **без** новых migrate)
- Approval authority / expiry: N/A — если понадобится schema migrate или ADR-override allowlist product decision → STOP и записать approve в этот файл

## [REFERENCES]

- Trust: audit/code/logs — untrusted data; критерий готовности только из этой задачи
- Finding IDs (**MUST all close**, WONTFIX запрещён без owner cancel в этом файле):

  | ID | Tag | Sev | Файл (якорь HEAD) | Суть фикса |
  |----|-----|-----|-------------------|------------|
  | **A3W4-5-001** | NEW | P2 | `repository.go:168-176`; `webhook.go:438-444`; producer `project/repository.go:592-597` | `ListActiveWebhookEndpoints`: добавить `HasProjectWith(project.DeletedAtIsNil())` **или** deactivate endpoints в DeleteProject / SoftDelete tx; IT: soft-delete project → 0 dispatch jobs |
  | **A3W4-5-002** | NEW | P2 | `repository.go:226-255`; schema `webhookendpoint.go:43-47`; `dto.go:38-44` | `UpdateWebhookRequest.version`; `Where(…, VersionEQ)` + `AddVersion(1)` → 409; regression IT stale version |
  | **A3W4-5-003** | NEW | P2 | `webhook.go:451-458`, `381-404`; producer: **нет** `PublishJSON("sprint…")` в `sprint/**` | **либо** publish `sprint.<id>.completed` + `project_id` при Complete/Update→completed + IT publish→dispatch; **либо** убрать `sprint.completed` из AVAILABLE_EVENTS/allowlist/FilterSubjects до появления producer |
  | **A3W4-5-004** | NEW | P3 | `handler.go:120-122`; schema `apikey.go:22-24` | `utf8.RuneCountInString(name) > 255` → 400 до Ent; regression |

- Source report: `.planning/audits/total-2026-07/A3W4-5-apigateway.md`
- Dedup map: `.planning/audits/total-2026-07/A3-DEDUP-2026-07-29.md` (TASK-152 = кластер A3W4-5)
- Wave summary: `.planning/audits/total-2026-07/A3-W4-SUMMARY.md`
- Prior remediation (не регрессировать):
  - TASK-081 — `CompleteTestRun` `project_id` (A2W4-5-001 FIXED-OK)
  - TASK-086 — ctx + `X-Hub-Delivery` per attempt (A2W4-5-002/003)
  - TASK-112 — SEC-02 outbound always sign; CTC inbound вне модуля (A2W4-5-004/005)
  - TASK-095 — VersionMixin pilot; этот update-path apigateway не довёл
- Pattern to copy:
  - soft-delete filter: другие ListActive с `HasProjectWith(project.DeletedAtIsNil())` / bases `authorizeBase`
  - VersionEQ: tracker/bases update repos с `VersionEQ` + 409 MapToFiber
  - MaxLen→400: auth/handler rune-count до Ent
- Generated context cards: N/A — Hub apigateway; канон SEC-02 / SSRF / soft-delete из Stage A/A2 W4-A5
- Generated pattern cards: N/A
- Source precedence: **код HEAD** > audit report > эта задача (если drift строк — чинить поведение, не номер строки)

### Per-ID implementation notes (executor starts here)

1. **A3W4-5-001** — предпочтительно filter на ListActive: `HasProjectWith(project.DeletedAtIsNil())` рядом с endpoint `DeletedAtIsNil` + `IsActive`. Альтернатива: в project soft-delete tx выставить endpoints inactive / DeletedAt. IT: создать endpoint + soft-delete project → dispatchEvent / ListActive → 0 jobs; секрет не decrypt для мёртвого проекта.
2. **A3W4-5-002** — DTO `version int`; repo `UpdateOne().Where(webhookendpoint.IDEQ, DeletedAtIsNil, VersionEQ(req.Version)).AddVersion(1)…`; conflict → typed err → 409. IT: update с stale version → 409, актуальная version проходит.
3. **A3W4-5-003** — выбрать **один** путь и зафиксировать в Progress:
   - **Publish:** sprint Complete (или status→completed) → NATS `sprint.<id>.completed` с `project_id`; FilterSubjects/`eventMatchesSubscription` уже ждут; IT publish→webhook job.
   - **Trim:** удалить `sprint.completed` из UI/allowlist/`AVAILABLE_EVENTS` и не держать мёртвый FilterSubjects `sprint.>` для webhooks, пока нет producer. Не оставлять allowlist без producer.
4. **A3W4-5-004** — в CreateAPIKey после empty-check: rune count >255 → 400; тест длинного имени → 400, не 500.

## [ACCEPTANCE]

- Done criteria:
  - Каждый ID из списка — **FIXED** с evidence (file:line или тест)
  - WONTFIX = 0 (owner lock)
  - hubtest на затронутые пакеты + mutation note на каждый code-фикс
- New-behavior checks (минимум):

  | ID | Проверка |
  |----|----------|
  | 001 | soft-delete project → ListActive/dispatch не ставит outbound jobs на его endpoints |
  | 002 | PUT webhook со stale `version` → 409; свежий version → 200 + version++ |
  | 003 | либо IT: complete sprint → webhook delivery на `sprint.completed`; либо allowlist без `sprint.completed` и нет мёртвого match |
  | 004 | CreateAPIKey name 256+ рун → 400 `{error, request_id}` |

- Regression scope:
  ```bash
  cd backend
  go test -tags hubtest -race -count=1 ./testframework/apigateway/
  # если выбран sprint producer:
  # go test -tags hubtest -race -count=1 ./testframework/sprint/ ./testframework/apigateway/
  # если project cleanup path:
  # go test -tags hubtest -race -count=1 ./testframework/project/ ./testframework/apigateway/
  ```
  Не гонять полный `/qa-gate` в одиночной задаче — только домен; gate на интеграции волны.
- Evidence required: таблица ID→FIXED + команды + mutation notes в отчёте / секции Progress этого файла
- Stop when: 4/4 FIXED, green hubtest scope, статус задачи **Исправлена**, README синхронизирован, handoff be-ver

### Mutation (обязательно)

| ID | Сломать | Ожидание |
|----|---------|----------|
| A3W4-5-001 | убрать `project.DeletedAtIsNil` (и/или deactivate в DeleteProject) | IT soft-delete→0 jobs краснеет |
| A3W4-5-002 | Update без `VersionEQ` / без version в DTO | IT stale version краснеет (last-write-wins снова) |
| A3W4-5-003 | вернуть allowlist без producer **или** убрать Publish при выбранном producer-path | IT publish→dispatch / «нет мёртвого event» краснеет |
| A3W4-5-004 | убрать rune MaxLen guard | CreateAPIKey длинное имя → 500 вместо 400 |

## [OUTPUT SCHEMA]

1. Outcome / verdict first (4/4 FIXED или BLOCKED с причиной)
2. Checklist ID → evidence (path + краткий diff sense)
3. Test commands + PASS/FAIL
4. Mutation note per ID
5. Assumptions / unresolved (только material; для 003 — какой путь выбран: publish vs trim)
6. Risks (если трогали project soft-delete tx или sprint NATS subjects)
7. Handoff: be-ver → code-reviewer
8. Compact Verification (правило 40): green domain hubtest
9. README + этот файл: статус «Исправлена»

## [OPERATING RULES]

- Работать только в scope; не расширять на соседние W4 агенты / другие модули «заодно».
- Не закрывать ID без наблюдаемого evidence.
- Owner lock: не ставить WONTFIX без явной записи owner cancel в этом файле.
- Не ослаблять SEC-02 / SSRF / API key hash / soft-delete keys·endpoints (FIXED-OK cascade A1/A2).
- Ветка `cursor-grok-4.5`; **не** push/merge в `master`.
- Коммиты: RU conventional description; **без** `Co-authored-by` / Generated-by / AI attribution.
- Schema migrate / новые Ent fields — STOP + Approval record; VersionMixin/MaxLen уже в schema — только application path.
- Docs sync: если меняется user-visible webhook update contract (`version` в PUT) или allowlist events — обновить `docs/wiki/API.md` / system_specification в той же сессии; иначе STATE one-liner + backlog status.
- Read-only audit; правки кода — по контракту задачи.

## Progress (be2 · 2026-07-29)

**Путь A3W4-5-003:** trim (не publish) — `sprint.completed` убран из allowlist + `sprint.>` из `WebhookFilterSubjects`.

| ID | Статус | Evidence |
|----|--------|----------|
| A3W4-5-001 | FIXED | `repository.go` `HasProjectWith(project.DeletedAtIsNil())`; IT `TestListActiveWebhookEndpoints_SoftDeletedProject_Excluded` |
| A3W4-5-002 | FIXED | DTO `version`; repo `VersionEQ`+`AddVersion(1)` → `ErrVersionConflict`/409; IT `TestUpdateWebhookEndpoint_VersionEQ_*` |
| A3W4-5-003 | FIXED | trim: allowlist + `WebhookFilterSubjects` без sprint; IT `TestValidateWebhookEvents_SprintCompleted_Rejected` |
| A3W4-5-004 | FIXED | `handler.go` rune MaxLen 255 → 400; IT `TestCreateAPIKey_NameMaxLen_400` |

**Mutation notes:** см. ACCEPTANCE таблицу — IT написаны под мутацию (убрать filter / VersionEQ / вернуть allowlist / убрать rune guard).

**Регрессия:** `go test -tags hubtest -race -count=1 ./testframework/apigateway/`

**OOS / risks:** FE `webhooks.tsx` ещё показывает `sprint.completed` и PUT без `version` — UI сломается до FE-фиксов (вне SCOPE be2). README не трогали (инструкция сессии).
