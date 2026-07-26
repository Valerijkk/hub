# W2 — storage / apigateway (repo-scan 2026-07-23)

> **Provenance class:** PRE_EXISTING · **Confidence:** MIXED · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

**Роль:** be-ver + security-auditor (readonly)  
**Scope:** `backend/internal/storage/`, `backend/internal/apigateway/`  
**Фокус:** SeaweedFS `aws-sdk-go-v2`, presigned URLs, webhook signature no-skip (SEC-02), BodyLimit  
**Код не менялся.**  
**Примечание:** отчёт восстановлен оркестратором — субагент был в Ask mode и не смог записать файл.

## Verdict

| Метрика | Значение |
|--------|----------|
| CRITICAL | 0 |
| HIGH | 0 |
| MEDIUM | 5 |
| LOW / accepted | 3 |
| **verified** | **false** — есть открытые MEDIUM |
| Дата | 2026-07-23 00:58 UTC+3 |

**Кратко:** инварианты TZ по стеку (aws-sdk-go-v2, не minio-go), исходящая HMAC `X-Hub-Signature-256` без skip, SSRF (`ValidateOutboundURL` + `SafeHTTPClient`), Fiber `BodyLimit` 25 MB и rate-limit `/files` 10/мин — **соблюдены**. Блокеров релиза (CRITICAL/HIGH) нет. MEDIUM — целостность confirm/size, lifecycle S3 после soft-delete, секрет вебхука at-rest / empty update / non-return.

---

## Checklist (Hub §3)

| Контроль | Статус | Evidence |
|----------|--------|----------|
| SeaweedFS через `aws-sdk-go-v2` | PASS | `storage/service.go:8-12,50-54` — `s3.NewFromConfig`, path-style; minio-go в prod-коде нет |
| Presigned PUT/GET | PASS | `PresignPutObject` / `PresignGetObject`, TTL 5m / 1h (`service.go:84-88,126-129`) |
| RBAC до выдачи URL | PASS | `handler.go:118-121` upload; `205-212` download через `FileProjectID` (CR-04) |
| Filename sanitize (NF-09 path) | PASS | `filename.go:9-34` |
| BodyLimit 25 MB | PASS | `wiring_fiber.go:33` + `FileUploadRateLimit` `wiring_fiber.go:57` |
| Presigned PUT size limit | ACCEPTED (NF-09) | Нет `ContentLength` в `PresignPutObject` — TZ: известный accepted risk |
| Webhook HMAC always set (SEC-02 outbound) | PASS | `deliverWithRetry` → `signPayload` всегда; header `X-Hub-Signature-256` (`webhook.go:242,296,317-321`); тест `webhook_test.go:166-203` |
| SSRF webhook URL | PASS | `ValidateOutboundURL` create/update/send + `NewSafeHTTPClient` (`webhook.go:39,286-287,351,391`) |
| Response body LimitReader | PASS | `io.LimitReader` 4096 (`webhook.go:306-312`) |
| API key hash at rest | PASS | SHA-256, raw key только при create (`apikey.go:87-112,206-209`) |
| Deactivated user + API key | PASS | APIGW-04 (`apikey.go:178-181`) |
| Webhook IDOR `/webhooks/:id` | PASS | `authorizeEndpoint` → project из БД (`handler.go:81-94`) |
| `http.DefaultClient` | PASS | 0 в обоих пакетах |
| `io.ReadAll` без Limit | PASS | 0 в prod-путях модулей |
| Timing-safe inbound compare | N/A | Модуль **исходящий** signer; inbound verify — у получателя. API key — lookup по hash |

---

## Findings

| Sev | Файл:строка | Уязвимость / gap | Рекомендация | OWASP / ASVS | Статус |
|-----|-------------|------------------|--------------|--------------|--------|
| MEDIUM | `storage/repository.go:62-67` + `service.go:102-105` | `ConfirmUpload` не делает `HeadObject`/проверку объекта в S3; `size` остаётся `0` (`CreatePending` → `SetSize(0)`). Confirmed без реального объекта / без учёта размера. | Перед confirm: `HeadObject`, проверить существование, записать `ContentLength` в `size`, reject если нет объекта или size > политики (поверх NF-09). | A04 Insecure Design · ASVS V12.1 | open |
| MEDIUM | `storage/service.go:154-157` · `orphan_cleanup.go:14-16` | Soft-delete только метаданные; S3-объект confirmed-файла **не** удаляется. Orphan cleanup чистит лишь `pending` >24h. Нарушение lifecycle корзины 30 дней / GDPR residual. | Job purge: soft-deleted >30d → `DeleteObject` + hard delete meta (или отдельный trash-worker). | A01/A04 · ASVS V8.3 | open |
| MEDIUM | `ent/schema/webhookendpoint.go:25-27` · `apigateway/repository.go:136-142` | `secret` вебхука хранится **plaintext** в Postgres (в отличие от API key hash). Утечка БД = подделка исходящих подписей / импersonation receiver verify. | AES-GCM at rest (как integration tokens) или envelope; в API по-прежнему не отдавать. | A02 Cryptographic Failures · ASVS V6.2 | open |
| MEDIUM | `apigateway/repository.go:156-157` · `webhook.go:382-388` | `UpdateEndpoint`: `Secret != nil` → `SetSecret(*req.Secret)` допускает **пустую** строку → HMAC с пустым ключом. | Reject `len(secret) == 0` / min entropy (напр. ≥16); на create тоже валидировать user-supplied secret. | A08 Software Integrity · ASVS V2.1 | open |
| MEDIUM | `apigateway/webhook.go:355-365` · `dto.go:46-53` | Autogen `uuid.New().String()` при пустом secret **не возвращается** в `WebhookResponse` — клиент не может проверить `X-Hub-Signature-256`. | One-time return secret на create (как API key) или требовать client-supplied secret. | A04 / usability→integrity | open |
| LOW | `storage/handler.go:170-177` | `entityType`/`entityID` на confirm без проверки существования/принадлежности сущности проекту. | Allowlist entity types + ownership check в домене. | A01 Broken Access Control | open |
| LOW | `storage/service.go:84-88` | Нет allowlist `ContentType` на presign. | Allowlist MIME; опционально сверять с magic bytes post-upload. | A04 | open |
| LOW (accepted) | `storage/service.go:84-88` · TZ NF-09 | Presigned PUT без `ContentLength`/квоты — обход Fiber BodyLimit 25 MB. | Accepted NF-09; mitigation: bucket policy / Seaweed quota / ContentLength в presign + post-Head. | A04 · TZ NF-09 | **accepted** |

---

## SEC-02 detail (apigateway)

Hub здесь **отправитель**, не приёмник inbound webhook JWT/claim (тот класс — LiveKit/GitHub).

- Подпись **всегда** считается: `sig := signPayload(job.body, job.ep.Secret)` (`webhook.go:242`).
- Заголовок **всегда** ставится: `X-Hub-Signature-256: sha256=` + hex HMAC (`webhook.go:296`).
- Нет ветки `if signature != ""` / skip при отсутствии claim.
- Регрессия: `TestSendWebhook_SignatureHeaderPresent`.

→ SEC-02 **no-skip для исходящих Hub webhooks — PASS**.

---

## BodyLimit / files

| Слой | Значение | Файл |
|------|----------|------|
| Fiber global BodyLimit | 25 × 1024 × 1024 | `cmd/server/wiring_fiber.go:33` |
| Rate `/api/files` | 10 / мин / user | `wiring_fiber.go:57` · `FileUploadRateLimit` |
| Upload path | Presigned PUT → SeaweedFS (мимо BodyLimit) | `storage/service.go:83-88` |
| TZ | Presigned без лимита = NF-09 accepted | `docs/tz_constraints.md` |

API JSON (`upload-url`, `confirm`) укладывается в 25 MB; риск размера — только direct PUT (NF-09).

---

## Positive controls (не findings)

1. Path-style S3 client для SeaweedFS (`service.go:50-53`).
2. Fail-closed RBAC если `rbacSvc == nil` (`storage/handler.go:47-50`, `apigateway/handler.go:63-65`).
3. Download/delete/confirm — projectID из БД файла, не из body (CR-04).
4. Webhook deliveries RBAC через `EndpointProjectID` (APIGW-01).
5. `maxWebhookResponseBodyBytes` + LimitReader (CR-03).
6. Queue-full → NAK, не silent Ack (APIGW-02).
7. Secret не в `WebhookResponse` / list API keys без raw key.

---

## Grep-скан (§6)

| Паттерн | Результат |
|---------|-----------|
| `http.DefaultClient` | 0 |
| `io.ReadAll` (prod) | 0 (LimitReader на webhook response) |
| `minio-go` / `minio.New` | 0 в storage/apigateway |
| `aws-sdk-go-v2` | storage service + orphan_cleanup |
| `ValidateOutboundURL` / `NewSafeHTTPClient` | webhook create/update/send |
| `DEV_MODE` | 0 в модулях |
| `*ent.Client` в handler/service | 0 (только EntRepository) |
| `hmac` / `signPayload` | всегда на delivery path |

---

## Handoff

| Owner | Действие |
|-------|----------|
| **be1/be2** | ConfirmUpload + HeadObject/size; reject empty webhook secret; one-time secret on create; optional entity bind check |
| **be + devops** | Soft-delete → S3 purge после 30d; рассмотреть encrypt webhook secret at rest |
| **security-auditor** | Re-audit после фикса MEDIUM → independent acceptance (post-fix target) |

**Вердикт:** `verified: false` — 2026-07-23 00:58 (UTC+3). CRITICAL/HIGH = 0 (релиз не блокируется security-гейтом), MEDIUM открыты.

---

## Verification

1. Scope: только storage + apigateway (+ wiring BodyLimit, schema webhook/file).
2. SEC-02 / SSRF / aws-sdk / BodyLimit / RBAC CR-04+APIGW-01 проверены по коду.
3. Severity не занижен: NF-09 оставлен accepted; empty secret = MEDIUM (нужен уже `PermProjectManageWorkflow`).
4. Denial/positive controls зафиксированы.
