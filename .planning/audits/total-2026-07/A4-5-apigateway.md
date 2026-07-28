# W4 агент A5 — apigateway

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: SEC-02 (outbound HMAC never skip) · SSRF outbound · API key SHA-256 · dispatcher retry (APIGW-02/03) · IDOR/RBAC webhooks.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/apigateway/apikey.go` | 210 | 1 | замечания |
| `backend/internal/apigateway/dto.go` | 64 | 1 | замечания |
| `backend/internal/apigateway/errors.go` | 65 | 0 | чисто |
| `backend/internal/apigateway/export_hooks.go` | 106 | 0 | чисто |
| `backend/internal/apigateway/handler.go` | 389 | 1 | замечания |
| `backend/internal/apigateway/repository.go` | 190 | 2 | замечания |
| `backend/internal/apigateway/webhook.go` | 460 | 4 | замечания |
| `backend/testframework/apigateway/apikey_auth_test.go` | 238 | 0 | чисто |
| `backend/testframework/apigateway/apikey_test.go` | 81 | 0 | чисто |
| `backend/testframework/apigateway/handler_extended_test.go` | 142 | 1 | замечания |
| `backend/testframework/apigateway/handler_test.go` | 176 | 0 | чисто |
| `backend/testframework/apigateway/handler_webhook_test.go` | 339 | 1 | замечания |
| `backend/testframework/apigateway/repository_extra_test.go` | 57 | 0 | чисто |
| `backend/testframework/apigateway/repository_test.go` | 72 | 0 | чисто |
| `backend/testframework/apigateway/webhook_dispatcher_apigw02_test.go` | 101 | 0 | чисто |
| `backend/testframework/apigateway/webhook_dispatcher_apigw03_test.go` | 98 | 0 | чисто |
| `backend/testframework/apigateway/webhook_dispatcher_wave97_test.go` | 192 | 1 | замечания |
| `backend/testframework/apigateway/webhook_signing_test.go` | 94 | 0 | чисто |
| `backend/testframework/apigateway/webhook_test.go` | 232 | 1 | замечания |

Файлов в списке: 19. Прочитано: 19. Пропущено: 0.

## Counts

| Severity | Count |
|----------|------:|
| P0 | 2 |
| P1 | 2 |
| P2 | 4 |
| P3 | 4 |
| **Σ** | **12** |

### Топ-3

1. **W4-A5-001 (P0)** — `containsEvent` сравнивает UI-имена (`issue.created`) с полным NATS subject (`project.<uuid>.issue.created`) → исходящие вебхуки из UI **никогда не доставляются**.
2. **W4-A5-002 (P0)** — диспетчер шлёт во все активные эндпоинты оргa без фильтра по `project_id` → при `*` (или после фикса матчинга) утечка payload чужих проектов.
3. **W4-A5-003 (P1)** — автогенерированный webhook `secret` не возвращается в API/DTO → клиент не может проверить HMAC.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W4-A5-001 | P0 | `webhook.go:325-331`; UI `frontend/src/modules/settings/webhooks.tsx:13-18` | `containsEvent` — только exact `==` или `*`. UI/API сохраняют короткие имена (`issue.created`, `test_run.completed`, `sprint.completed`), а JetStream subject = `project.<uuid>.issue.created` / `tms.testrun.completed` / … | Продуктовый путь «создать вебхук в UI» никогда не матчит реальные события; доставки пустые. Тесты CRUD пишут `issue.created` и не проверяют dispatch | Нормализовать: suffix/token match (`strings.HasSuffix` / последний сегмент после `.`) + явная таблица aliases UI→subject; интеграционный тест: create с `issue.created` + publish `project.<id>.issue.created` → job enqueued |
| W4-A5-002 | P0 | `webhook.go:209-234`; `repository.go:116-119` | `ListActiveWebhookEndpoints` без scope по проекту; `dispatchEvent` не сверяет `ep.ProjectID` с project из subject | Любой эндпоинт с `events:["*"]` (или после фикса 001) получает payload **всех** проектов инстанса (cross-project leak) | Парсить projectID из subject (`project.<id>.…`) / edges; фильтровать `ep.Edges.Project` / `ProjectID`; denial-тест: hook проекта A не получает issue проекта B |
| W4-A5-003 | P1 | `webhook.go:355-365`; `dto.go:46-53`; `webhook.go:451-459` | При пустом `req.Secret` секрет = `uuid.New()`, но `WebhookResponse` **не содержит** `secret`; UI (`createWebhook`) secret не шлёт и не показывает | Получатель не узнает HMAC-ключ → не может верифицировать `X-Hub-Signature-256`; секрет только в БД plaintext | One-time `secret` в Create-ответе (как API key `key`); UI показать один раз; list/update — без secret |
| W4-A5-004 | P1 | `handler.go:213-216`; `webhook.go:350-360`; `dto.go:35` | `Events` не валидируются в handler/service (`validate` tags не enforced Bind); допустимы `[]` / неизвестные имена | Мёртвые/бессмысленные подписки; нет 400 на пустой список | Require `len(events)>=1` + allowlist (согласованный с AVAILABLE_EVENTS/aliases); тест empty → 400 |
| W4-A5-005 | P2 | `webhook.go:285-300` | `http.NewRequest` без `ctx`; cancel `deliverWithRetry` не прерывает HTTP Do (только delay) | При Stop/shutdown воркеры дожидают полный SafeHTTPClient timeout (10s×retries) | `http.NewRequestWithContext(ctx, …)` + прокинуть ctx в `sendWebhook` |
| W4-A5-006 | P2 | `webhook.go:136-143`; `repository.go:122-133` | Delivery log пишется на **каждую** попытку retry (до 4 строк на job); при queue-full NAK — возможны дубли доставок (at-least-once) без idempotency key | Раздувание `webhook_delivery`; внешний consumer видит повторные POST без `delivery_id`/`X-Hub-Delivery` | Заголовок уникального delivery id; upsert по attempt; документировать at-least-once |
| W4-A5-007 | P2 | `repository.go:99-100`, `168-169`; схемы `apikey`/`webhookendpoint` | Hard delete API key / webhook endpoint (нет soft-delete/корзины) | TZ: soft delete 30д; случайный revoke/delete необратим; deliveries orphan/cascade | Soft-delete + корзина или явное исключение в ADR для credentials |
| W4-A5-008 | P2 | `webhook.go:317-321` | Webhook `secret` хранится и читается plaintext для HMAC (ожидаемо), без at-rest encryption | Компрометация БД = подделка исходящих подписей от имени Hub | AES-GCM (platform/crypto) для secret at rest; decrypt только в dispatcher |
| W4-A5-009 | P3 | `apikey.go:45-46`, `80-84` | Поле `stopCh` создаётся, нигде не используется/не закрывается | Мёртвый код; путаница с lifecycle Close | Удалить `stopCh` или использовать для drain |
| W4-A5-010 | P3 | `webhook_test.go:135-164` | `TestSendWebhook_5xxRetries` дважды зовёт `sendWebhook`, **не** `deliverWithRetry` | Не ловит регресс delays/`CreateWebhookDelivery` на fail path | Тест через `ExportWebhookDispatcherDeliverWithRetry` + stub repo, assert ≥2 deliveries / delays |
| W4-A5-011 | P3 | `handler_extended_test.go:81-84`; `handler_webhook_test.go:139-148` | CRUD-тесты создают events=`issue.created` без проверки матчинга к NATS subject | Регресс W4-A5-001 зелёный на forever | Добавить dispatch/match regression (см. 001) |
| W4-A5-012 | P3 | rules/`60-tz` vs `webhook.go` | Документы ссылают ConstantTimeCompare в `apigateway/webhook.go`; в модуле только **outbound** `signPayload`, сравнения секретов нет | Drift docs↔код; путает SEC-02 inbound (LiveKit/GitHub) с outbound Hub | Уточнить docs: outbound всегда ставит `X-Hub-Signature-256`; CTC — у inbound-верификаторов |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| `webhook.go:242`, `296` | SEC-02 outbound: `signPayload` + заголовок `X-Hub-Signature-256` **всегда** на каждой попытке; skip path нет |
| `webhook.go:39`, `286-288`, `351-353`, `391-393` | SSRF: `NewSafeHTTPClient` + `ValidateOutboundURL` на create/update/send; response body `LimitReader` 4096 (CR-03) |
| `webhook.go:115-145`, `205-234` | Dispatcher retry: MaxDeliver=3, Ack только после успешного enqueue, NAK+delay при queue full / repo error (APIGW-02); FilterSubjects (APIGW-03) |
| `apikey.go:94-96`, `148`, `206-209` | API key: 32 байта crypto/rand → SHA-256 hex at rest; plaintext только в Create-ответе |
| `apikey.go:178-181` | APIGW-04: deactivated user → ErrAPIKeyUserDeactivated |
| `handler.go:62-94`, `209-356` | APIGW-01: RBAC `permWebhookManage`; project из БД на `/webhooks/:id`; nil RBAC fail-closed 500 |
| `errors.go` | MapToFiber sentinel → 401/403/404/400; без panic |
| `export_hooks.go` | Только hubtest exports; без прод-логики |
| `apikey_auth_test.go` | Valid/invalid/deactivated/pass-through/SEC-01 order/RequireAuth — asserts есть |
| `apikey_test.go` | hash determinism/uniqueness, containsEvent/*, signPayload |
| `handler_test.go` | Create validation, revoke bad id, list, expired key |
| `webhook_signing_test.go` | Deterministic HMAC, secret/body sensitivity, hex64 |
| `webhook_dispatcher_apigw02_test.go` | Queue full → err; poison JSON → nil (Ack); repo error → err |
| `webhook_dispatcher_apigw03_test.go` | FilterSubjects parity; notification subject не сканирует endpoints |
| `webhook_test.go` (кроме 5xx) | SSRF blocklist, bad scheme, signature header+hmac.Equal на receiver, truncate body |
| `repository_*.go` tests | CRUD keys/endpoints/deliveries с asserts |
| `handler_webhook_test.go` IDOR | Stranger create/list/update/delete/deliveries → 403; unauthenticated 401 |

## SEC-02 / фокус — резюме

| Контроль | Статус |
|----------|--------|
| Outbound signature never skip | ✅ всегда `signPayload` + header |
| ConstantTimeCompare в модуле | N/A outbound; inbound CTC — другие модули (calls/github); docs drift = W4-A5-012 |
| SSRF outbound | ✅ ValidateOutboundURL + SafeDialer + body limit |
| API key hashing | ✅ SHA-256, не plaintext в БД |
| Dispatcher retry / no silent drop | ✅ APIGW-02 NAK; APIGW-03 filters |
| Event match / project scope | ❌ W4-A5-001, W4-A5-002 |
| Secret disclosure on create | ❌ W4-A5-003 |

## Handoff

Приоритет стадии B: **001 → 002 → 003 → 004**. Без 001 модуль вебхуков фактически мёртв для UI; без 002 фикс матчинга откроет cross-project leak.
