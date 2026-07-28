# A2 W4 агент 5 — apigateway

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Зеркало: `W4-ASSIGNMENTS.md` A5 · Stage A: `A4-5-apigateway.md`  
> Фокус: SEC-02 · SSRF · match/scope после TASK-011/012 · false closures P2-BE-16

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/apigateway/apikey.go` | 181 | 0 | чисто |
| `backend/internal/apigateway/dto.go` | 60 | 0 | чисто |
| `backend/internal/apigateway/errors.go` | 57 | 0 | чисто |
| `backend/internal/apigateway/export_hooks.go` | 103 | 0 | чисто |
| `backend/internal/apigateway/handler.go` | 335 | 0 | чисто |
| `backend/internal/apigateway/repository.go` | 264 | 0 | чисто |
| `backend/internal/apigateway/webhook.go` | 531 | 4 | замечания |
| `backend/testframework/apigateway/apikey_auth_test.go` | 192 | 0 | чисто |
| `backend/testframework/apigateway/apikey_test.go` | 67 | 0 | чисто |
| `backend/testframework/apigateway/handler_extended_test.go` | 119 | 0 | чисто |
| `backend/testframework/apigateway/handler_test.go` | 148 | 0 | чисто |
| `backend/testframework/apigateway/handler_webhook_test.go` | 289 | 0 | чисто |
| `backend/testframework/apigateway/repository_extra_test.go` | 45 | 0 | чисто |
| `backend/testframework/apigateway/repository_test.go` | 59 | 0 | чисто |
| `backend/testframework/apigateway/webhook_dispatcher_apigw02_test.go` | 89 | 0 | чисто |
| `backend/testframework/apigateway/webhook_dispatcher_apigw03_test.go` | 80 | 0 | чисто |
| `backend/testframework/apigateway/webhook_dispatcher_wave97_test.go` | 161 | 0 | чисто |
| `backend/testframework/apigateway/webhook_signing_test.go` | 80 | 0 | чисто |
| `backend/testframework/apigateway/webhook_test.go` | 206 | 1 | замечания |

Файлов в списке: 19. Прочитано: 19. Пропущено: 0.  
Сверка продьюсера (вне списка): `backend/internal/tms/service_runs.go:363-368`.

## Находки (активные)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A2W4-5-001 | REGRESS | P1 | `webhook.go:217-230`; `extractEventProjectID:393-419`; producer `tms/service_runs.go:363-368` | После TASK-012 fail-closed: событие без extractable `project_id` **не доставляется**. `CompleteTestRun` публикует `tms.run.completed.<runID>` **без** `project_id` (в отличие от import-path) | UI-подписка `test_run.completed` снова мёртва на основном Complete-path — регресс относительно «фиксов 011+012». Тест scope проверяет payload-with-project_id, не prod CompleteTestRun | В `CompleteTestRun` добавить `"project_id": projectID`; regression: publish shape → dispatch enqueue |
| A2W4-5-002 | RESIDUAL | P2 | `webhook.go:300-315` | `http.NewRequest` **без** ctx; `sendWebhook` не принимает ctx (W4-A5-005). Заявлено в P2-BE-16 — **в коде не сделано** | Shutdown/Stop не отменяет HTTP Do — до 10s×retries hang | `NewRequestWithContext` + прокинуть ctx из `deliverWithRetry` |
| A2W4-5-003 | RESIDUAL | P2 | `webhook.go:254-284`; headers `:310-313` | Delivery log на каждую retry-попытку; нет `X-Hub-Delivery` / idempotency (W4-A5-006). P2-BE-16 claimed — **не сделано** | Раздувание `webhook_delivery`; consumer видит дубли POST без delivery id | Unique delivery id header; upsert attempt; док at-least-once |
| A2W4-5-004 | FIXED | P3 | `60-tz-constraints.mdc` | CTC → github/gitlab inbound; apigateway = outbound sign always | Evidence: TASK-112 | Уточнить: outbound всегда `X-Hub-Signature-256`; CTC — inbound верификаторы |
| A2W4-5-005 | FIXED | P3 | `webhook_test.go` | `TestSendWebhook_5xxRetries` → `ExportWebhookDispatcherDeliverWithRetry` | Evidence: TASK-112 | Assert через `ExportWebhookDispatcherDeliverWithRetry` |

## FIXED-OK

| Stage A ID | Sev | Статус | Доказательство |
|------------|-----|--------|----------------|
| W4-A5-001 | P0 | FIXED-OK | `eventMatchesSubscription` suffix/alias; `webhook_event_match_test.go` |
| W4-A5-002 | P0 | FIXED-OK | project filter + edge Project; denial-тест (регресс A2W4-5-001 = side-effect) |
| W4-A5-003 | P1 | FIXED-OK | `CreateWebhookResponse.Secret` one-time |
| W4-A5-004 | P1 | FIXED-OK | `validateWebhookEvents` allowlist |
| W4-A5-007 | P2 | FIXED-OK | soft-delete `DeletedAt` keys/endpoints |
| W4-A5-008 | P2 | FIXED-OK | encrypt secret at rest (cipher required) |
| W4-A5-009 | P3 | FIXED-OK | `stopCh` удалён |
| W4-A5-011 | P3 | FIXED-OK | match/dispatch regression tests |

| Контроль | Статус |
|----------|--------|
| Outbound signature never skip | ✅ всегда `signPayload` + header |
| SSRF | ✅ Validate + SafeHTTPClient + LimitReader |
| API key SHA-256 | ✅ |
| RBAC / IDOR webhooks | ✅ authorizeEndpoint |
| Event match UI↔NATS | ✅ |
| Project scope | ✅ (с регрессом TMS payload) |

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **19 / 19 / 0** |
| Активных | **5** (NEW 0 · REGRESS 1 · RESIDUAL 4) |
| FIXED-OK | **8** |
| P0 | 0 |
| P1 | 1 |
| P2 | 2 |
| P3 | 2 |

### Топ

1. **A2W4-5-001 (P1 REGRESS)** — CompleteTestRun без `project_id` → webhook TMS не доставляется после fail-closed scope.  
2. **A2W4-5-002/003 (P2 RESIDUAL)** — false closure P2-BE-16: ctx на HTTP и delivery-id не сделаны.
