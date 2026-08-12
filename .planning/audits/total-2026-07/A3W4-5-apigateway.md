# A3 W4 агент A5 — apigateway (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline: [A2W4-5-apigateway.md](A2W4-5-apigateway.md) · A1: [A4-5-apigateway.md](A4-5-apigateway.md)  
> Assignment: [W4-ASSIGNMENTS.md](W4-ASSIGNMENTS.md) § A5 · Периметр: **19** · Прочитано: **19/19** · Пропущено: **0**  
> Remediation: TASK-081 (A2W4-5-001) · TASK-086 (A2W4-5-002/003) · TASK-112 (A2W4-5-004/005)  
> Фокус: API key SHA-256 · SEC-02 outbound HMAC no-skip · SSRF ValidateOutboundURL · soft-delete · optimistic lock · timing

## Сверка с A2 (A2W4-5-*)

| A2 ID | Tag | Статус | Комментарий |
|-------|-----|--------|-------------|
| A2W4-5-001 | FIXED-OK | закрыт | TASK-081: `CompleteTestRun` публикует `"project_id"` (`tms/service_runs.go:358-365`); fail-closed scope в `dispatchEvent` остаётся |
| A2W4-5-002 | FIXED-OK | закрыт | TASK-086: `http.NewRequestWithContext(ctx,…)` (`webhook.go:320`); IT `webhook_delivery_ctx_task086_test.go` cancel |
| A2W4-5-003 | FIXED-OK | закрыт | TASK-086: `X-Hub-Delivery` UUID на попытку (`webhook.go:277-328`); IT unique-per-attempt |
| A2W4-5-004 | FIXED-OK | закрыт | TASK-112: CTC inbound github/gitlab; apigateway = outbound `signPayload` always |
| A2W4-5-005 | FIXED-OK | закрыт | TASK-112: `TestSendWebhook_5xxRetries` → `ExportWebhookDispatcherDeliverWithRetry` |

**REGRESS:** 0 · **A2 open residual:** 0 (все 5 A2 active → FIXED-OK)

### Сверка A1 / FIXED-OK cascade (подтверждение на HEAD)

| Stage A / class | Tag | Evidence |
|-----------------|-----|----------|
| W4-A5-001 event match | FIXED-OK | `eventMatchesSubscription` + aliases; `webhook_event_match_test.go` (вне списка, сверка) |
| W4-A5-002 project scope | FIXED-OK | `extractEventProjectID` + `webhookEndpointProjectID`; denial IT |
| W4-A5-003 secret one-time | FIXED-OK | `CreateWebhookResponse.Secret` |
| W4-A5-004 events allowlist | FIXED-OK | `validateWebhookEvents` |
| W4-A5-005/006 ctx+delivery | FIXED-OK | через TASK-086 / A2W4-5-002/003 |
| W4-A5-007 soft-delete keys/eps | FIXED-OK | `DeleteAPIKey` / `DeleteWebhookEndpoint` → `DeletedAt` |
| W4-A5-008 encrypt secret | FIXED-OK | `encryptSecret` + `ErrEncryptionRequired` |
| W4-A5-009 stopCh | FIXED-OK | удалён; `Close`/`Stop` via `stopOnce` |
| W4-A5-011 match/dispatch IT | FIXED-OK | event_match + scope denial |
| SEC-02 outbound | FIXED-OK | всегда `signPayload` + `X-Hub-Signature-256`; skip path нет |
| SSRF | FIXED-OK | `ValidateOutboundURL` create/update/send + `SafeHTTPClient` + `LimitReader` 4096 |
| API key SHA-256 | FIXED-OK | `hashKey` → lookup by hash (не plaintext compare / не `==` секрета) |

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/apigateway/apikey.go` | 208 | 0 | чисто · SHA-256 · touch pool · deactivated deny |
| `backend/internal/apigateway/dto.go` | 71 | 0 | чисто · CreateWebhookResponse.Secret |
| `backend/internal/apigateway/errors.go` | 65 | 0 | чисто · MapToFiber |
| `backend/internal/apigateway/export_hooks.go` | 144 | 0 | hubtest re-exports · чисто |
| `backend/internal/apigateway/handler.go` | 389 | 1 | RBAC/IDOR OK; API key name MaxLen gap |
| `backend/internal/apigateway/repository.go` | 291 | 2 | soft-delete OK; ListActive без project.DeletedAt; Update без VersionEQ |
| `backend/internal/apigateway/webhook.go` | 609 | 2 | SEC-02/SSRF/ctx/delivery OK; soft-delete project scope; sprint allowlist без producer |
| `backend/testframework/apigateway/apikey_auth_test.go` | 238 | 0 | чисто · A9012/deactivated/SEC-01 |
| `backend/testframework/apigateway/apikey_test.go` | 81 | 0 | чисто · hash/sign/containsEvent |
| `backend/testframework/apigateway/handler_extended_test.go` | 142 | 0 | чисто |
| `backend/testframework/apigateway/handler_test.go` | 176 | 0 | чисто · expired key |
| `backend/testframework/apigateway/handler_webhook_test.go` | 339 | 0 | чисто · IDOR/SSRF URL |
| `backend/testframework/apigateway/repository_extra_test.go` | 57 | 0 | чисто |
| `backend/testframework/apigateway/repository_test.go` | 74 | 0 | чисто · WithProject |
| `backend/testframework/apigateway/webhook_dispatcher_apigw02_test.go` | 109 | 0 | чисто · queue-full NAK |
| `backend/testframework/apigateway/webhook_dispatcher_apigw03_test.go` | 98 | 0 | чисто · FilterSubjects |
| `backend/testframework/apigateway/webhook_dispatcher_wave97_test.go` | 192 | 0 | чисто |
| `backend/testframework/apigateway/webhook_signing_test.go` | 94 | 0 | чисто · HMAC |
| `backend/testframework/apigateway/webhook_test.go` | 275 | 0 | чисто · SSRF/5xx retry/signature/truncate |

Файлов в списке: **19**. Прочитано: **19**. Пропущено: **0**.  
Сверка вне списка (evidence remediation): `backend/internal/tms/service_runs.go:358-365`; `backend/testframework/apigateway/webhook_delivery_ctx_task086_test.go`; `webhook_event_match_test.go`.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A3W4-5-001 | NEW | P2 | `repository.go:168-176`; `webhook.go:438-444`; producer `project/repository.go:592-597` | `ListActiveWebhookEndpoints` фильтрует только `IsActive`+endpoint `DeletedAtIsNil`; `WithProject()` не требует `project.DeletedAtIsNil`. `SoftDeleteWithCrossLinkCleanup` не деактивирует webhook endpoints | Soft-deleted проект (корзина) продолжает получать исходящие POST на внешний URL при событиях с тем же `project_id` (W19 C-008). Секреты decrypt'ятся для «мёртвого» проекта | На ListActive: `HasProjectWith(project.DeletedAtIsNil())` **или** deactivate endpoints в DeleteProject tx; IT: soft-delete project → 0 jobs |
| A3W4-5-002 | NEW | P2 | `repository.go:226-255`; schema `webhookendpoint.go:43-47`; DTO `dto.go:38-44` | `WebhookEndpoint` имеет `VersionMixin`, но `UpdateWebhookEndpoint` без `VersionEQ` / `AddVersion`; `UpdateWebhookRequest` без `version` | NFR-24 / ADR-ENT-VERSION-MIXIN: concurrent PUT last-write-wins; клиент не может сделать optimistic lock; TASK-095 закрыл pilot-домены, apigateway update-path не довёл | Принять `version` в Update DTO; `Where(ID, DeletedAtIsNil, VersionEQ)` + `AddVersion(1)` → 409; regression IT stale version |
| A3W4-5-003 | NEW | P2 | `webhook.go:451-458`, `381-404`; producer: **нет** `PublishJSON("sprint…")` в `backend/internal/sprint/**` | UI/allowlist `sprint.completed` + match `sprint.<id>.completed` готовы, но sprint-сервис **не публикует** NATS subject `sprint.*` (только poker.*) | Продуктовая подписка `sprint.completed` мёртва end-to-end (класс A2W4-5-001, другой event). FilterSubjects `sprint.>` бесполезен для webhooks | Publish `sprint.<id>.completed` + `project_id` при Complete/Update→completed **или** убрать из AVAILABLE_EVENTS/allowlist до появления producer; IT publish→dispatch |
| A3W4-5-004 | NEW | P3 | `handler.go:120-122`; schema `apikey.go:22-24` | `CreateAPIKey` проверяет только `name == ""`, не `MaxLen(255)` / rune-count | Имя >255 → Ent validation → ошибка через `MapToFiber` default → **500**, не 400 (класс MaxLen→500) | `utf8.RuneCountInString(name) > 255` → 400; regression |

## Чисто (подтверждено после remediation)

- **SEC-02 outbound:** каждая попытка — `signPayload` + `X-Hub-Signature-256=sha256=…`; skip/empty-signature path отсутствует.
- **SSRF:** `ValidateOutboundURL` на Create/Update/send; `NewSafeHTTPClient` + dial blocklist; response `io.LimitReader` 4096.
- **API key:** crypto/rand 32B → SHA-256 hex at rest; verify by hash lookup (не timing-`==` plaintext); soft-delete revoke; deactivated user → 403; bounded touch pool.
- **HMAC:** outbound sign only; inbound CTC — github/gitlab (не этот модуль).
- **RBAC/IDOR:** `authorizeProject` / `authorizeEndpoint` (project из БД); nil RBAC → 500.
- **Dispatcher:** Ack after enqueue; NAK+delay on queue full; FilterSubjects; ctx cancel на HTTP; `X-Hub-Delivery` per attempt.
- **Soft-delete endpoints/keys:** `DeletedAt` + list/get фильтры; secret at-rest encrypt.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 3 |
| P3 | 1 |
| **Total** | **4** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 active) | 5 (`A2W4-5-001`…`005`) |
| FIXED-OK (A1 cascade) | 10+ |
| RESIDUAL | 0 |
| NEW | 4 (`A3W4-5-001`…`004`) |
| REGRESS | 0 |

A2 baseline: **5** open → FIXED-OK 5 · RESIDUAL 0 · NEW 4 · REGRESS 0 · open **4** (P2=3 · P3=1).

### Топ

1. **A3W4-5-001 (P2 NEW)** — soft-delete project не гасит webhooks / ListActive без `project.DeletedAtIsNil`.
2. **A3W4-5-002 (P2 NEW)** — VersionMixin без VersionEQ на Update webhook.
3. **A3W4-5-003 (P2 NEW)** — `sprint.completed` allowlist/match без NATS producer.
4. **A3W4-5-004 (P3 NEW)** — API key name MaxLen → 500.
