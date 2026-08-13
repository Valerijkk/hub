# Hub — Руководство по безопасности

> Единая база знаний. Обновляется ARCH/BE при изменении политик.  
> Аудит: `.planning/audits/code-review-final/SECURITY.md` · TZ: `docs/tz_constraints.md`

---

## Краткий статус (2026-06-22)

> Quality processes audit: `.planning/audits/2026-06/AUDIT-2026-06-22-quality-processes.md`  
> Security re-audit: `.planning/audits/2026-06/AUDIT-2026-06-23-full.md`

| Категория | Статус |
|-----------|--------|
| CRITICAL findings (v2.0.1) | **12/12 FIXED** ✅ |
| HIGH findings | **15 FIXED**, 1 PARTIAL (BE-H1), 1 DEFERRED (bundle size) |
| Production readiness | ✅ YES с оговорками (PG-03 rotation, FE-H5 audit) |
| Auth model | **Magic Link only** — Telegram notify-only (ADR-AUTH-01) |

## OWASP Top 10 — чек-лист Hub

> **Последний аудит:** 2026-06-17 (comprehensive re-audit all CRITICAL + HIGH findings).
> **Baseline:** `.planning/audits/code-review-final/SUMMARY.md` (2026-05-15).

| # | Угроза | Статус Hub | Открытые ID |
|---|--------|------------|-------------|
| A01 | Broken Access Control | ✅ Fixed | — (NF-02, NF-03, dashboard scope закрыты) |
| A02 | Cryptographic Failures | ⚠️ Partial | NF-10 (accepted с rate-limit) |
| A03 | Injection | ✅ | Ent ORM, parameterized queries, Mongo filter |
| A04 | Insecure Design | ✅ Fixed | — (SEC-06 pg_advisory_lock, NF-09 sanitizeFilename) |
| A05 | Security Misconfiguration | ✅ Fixed | — (NF-11, INF-02, SecurityHeaders) |
| A06 | Vulnerable Components | Monitor | Dependabot, pinned versions |
| A07 | Auth Failures | ✅ Fixed | — (SEC-01, APIKeyAuth ordering, refresh single-flight) |
| A08 | Software Integrity | ✅ Fixed | — (SEC-02 sha256 required; webhooks OK) |
| A09 | Logging Failures | ✅ | slog + request_id, audit middleware |
| A10 | SSRF | ✅ Fixed | — (NEW-02 OpenGraph SafeHTTP; NF-07 WebPush; SEC-04 GitLab) |

## Статус findings — формальный re-audit 2026-06-17

### CRITICAL (12) — блокеры продакшена

| ID | Статус | Файл:строка | Верификация |
|----|--------|-------------|-------------|
| SEC-01 / NF-01 | ✅ **FIXED** | `cmd/server/wiring_auth.go:75-84` | `api.Use(APIKeyAuth)` перед `RegisterRoutes` |
| SEC-02 / NF-04 | ✅ **FIXED** | `calls/handler.go:261-265` | `sha256` claim mandatory + `hmac.Equal` |
| SEC-03 / NF-05 | ✅ **FIXED** | `docker-compose.prod.yml:130-153` | `livekit-prod.yaml` удалён; heredoc из env |
| SEC-04 / C-01 | ✅ **FIXED** | `gitlab/service.go:55,71` | `safehttp.NewSafeHTTPClient` + `ValidateOutboundURL` |
| SEC-05 / C-02 | ✅ **FIXED** | `ws/hub.go:213-216` | `select <-client.done` перед send |
| SEC-06 / C-03 | ✅ **FIXED** | `auth/service_registration_lock.go:47` | `pg_advisory_xact_lock` + ent tx |
| SEC-07 / C-04 | ✅ **FIXED** | `auth/magiclink.go:135` | `subtle.ConstantTimeCompare` |
| INF-01 | ✅ **FIXED** | `docker-compose.prod.yml:150` | `departure_timeout: 120` |
| INF-02 | ✅ **FIXED** | `.github/workflows/deploy.yml:63-65` | `DEV_MODE` удалён из CI |
| INF-03 | ✅ **FIXED** | `docker-compose.local.yml:120-122` | `${RESEND_API_KEY:-}` placeholder |
| INF-04 | ✅ **FIXED** | `cmd/server/wiring_integrations.go:97-108` | `backupS3Client` проверка + warn при пустом S3 |
| FE-01 | ✅ **FIXED** | `messenger/message-item.tsx:1-100` | Hooks до conditional return |

**CRITICAL Verdict:** 12/12 FIXED ✅

---

### HIGH (16) — безопасность и архитектура

| ID | Статус | Файл:строка | Примечание |
|----|--------|-------------|------------|
| SEC-H1 / NF-02 | ✅ **FIXED** | `ws/handler.go:84-105,218-231` | `CheckOrigin` + `CanSubscribe` authz |
| SEC-H2 / NF-03 | ✅ **FIXED** | `search/service.go:585` | `channel_id: {$in: channelIDs}` |
| SEC-H3 / NF-07 | ✅ **FIXED** | `notification/push.go:157,218` | `ValidateOutboundURL` на endpoint |
| BE-H1 / H-01 | ⚠️ **PARTIAL** | `backup/service.go:173,237` | `LimitReader` ✅; `imex/handler.go:66` unbounded но BodyLimit 25MB |
| BE-H2 / H-02 | ✅ **FIXED** | `apigateway/webhook.go:280` | `io.Copy(io.LimitReader)` замена `ReadFrom` |
| BE-H3 / H-03 | ✅ **FIXED** | `dashboard/service.go:180-210` | Фильтр по `userID` / `projectID` |
| BE-H4 / H-04 | ✅ **FIXED** | `apigateway/service.go` + `wiring_integrations.go` (API gateway worker) | Worker pool + `defer Close()` |
| BE-H5 / H-05 | ✅ **FIXED** | `wiring_infra.go:18-26` + `wiring_messenger.go:36-54` | `rawDB` shared pool в `/health` (`/live`/`/ready`) |
| BE-H6 / H-06 | ✅ **FIXED** | `audit/service.go` DeleteUserGDPR + `gdpr_mongo_retry.go` | Каскад Ent tx; Mongo erase post-Commit; C-W19-002 JetStream retry `audit.gdpr.mongo_erase` |
| BE-H7 / H-07 | ✅ **FIXED** | `audit/retention.go:98-127`, `audit/service.go:252-268` | `RunRetentionCleanup` cron + S3 purge записей; `PurgeExpired` корзины 30d |
| FE-H1 | ✅ **FIXED** | `api/client.ts:97-108` | Single-flight mutex `refreshing` |
| FE-H2 | ✅ **FIXED** | `stores/auth-store.ts:46-88` | `BroadcastChannel` + storage fallback |
| FE-H3 | ✅ **FIXED** | `api/auth.ts:121-123` | `POST /auth/logout` endpoint |
| FE-H4 | ✅ **FIXED** | `ui/error-boundary.tsx` + routes | ErrorBoundary в 7 модулях |
| FE-H5 | 🔄 **DEFERRED** | — | Bundle size требует build analysis |
| ARCH-H1 | ✅ **FIXED** | `auth/handler.go:17-26,206-212` | Handler держит только `*AuthService`; `GetMe` → `h.service.GetMe` (без `*ent.Client`) |

**HIGH Verdict:** 15 FIXED, 0 NOT FOUND, 1 PARTIAL (BE-H1), 0 OPEN

---

### Дополнительные findings (новые фичи 2026-06-17)

| ID | Severity | Статус | Файл:строка | Примечание |
|----|----------|--------|-------------|------------|
| NEW-01 | MEDIUM | ✅ **FIXED** | `middleware/ratelimit.go` | `MessageSendRateLimit` 60/min |
| NEW-02 | MEDIUM | ✅ **FIXED** | `platform/http/opengraph.go:26,30,48` | SafeHTTP + ValidateOutboundURL + LimitReader |
| NEW-03 | LOW | ✅ **FIXED** | `backup/service.go:173,237` | `LimitReader` 512 MiB cap |
| NEW-04 | INFO | ✅ **PASS** | `audit/middleware.go` | Async recorder, skip health/docs |
| NEW-05 | INFO | ✅ **PASS** | `audit/handler.go:284-305` | `BanUser` требует admin + audit log |
| NEW-06 | INFO | ✅ **PASS** | `middleware/security_headers.go` + `Caddyfile.prod` | X-Frame-Options, CSP, nosniff; **HSTS на Caddy** (не Fiber) |
| NEW-07 | INFO | ✅ **PASS** | `observability/timing.go` | Timed wrapper, no secrets logged |
| NEW-08 | INFO | ✅ **PASS** | `storage/filename.go` | `sanitizeFilename` 200-char cap |
| NF-10 | MEDIUM | ⚠️ **ACCEPTED** | `auth/magiclink.go:46-52` | 6-digit (~10⁶ entropy) + rate-limit 3/10m |
| NF-11 | MEDIUM | ✅ **FIXED** | `cmd/server/wiring_fiber.go` | `EnableTrustedProxyCheck` + private CIDRs |
| NF-12 | MEDIUM | ✅ **FIXED** | `auth/service_session.go` | SHA-256 hash refresh tokens |
| NF-13 | LOW | ⚠️ **ACCEPTED** | `ws/hub.go:217-220` | Drop message on full buffer (не блокируем room) |
| NF-14 | INFO | ⚠️ **ACCEPTED** | — | CSRF токены moot при Bearer JWT |

## Политика управления секретами

| Правило | Реализация |
|---------|------------|
| Секреты только в env | `.env` в `.gitignore`, `.env.example` без значений |
| Никаких hardcoded creds | Prod LiveKit — env only; local `livekit.yaml` — dev keys |
| Encryption at rest | AES-GCM для GitHub/GitLab tokens |
| Сравнение секретов | `subtle.ConstantTimeCompare` — не `!=` |
| CI secrets | GitHub Secrets, не в workflow yaml |
| Rotation | Manual via Admin panel для API keys |

## Аутентификация и авторизация

```
Magic Link (email) → JWT access (15 min, memory — не sessionStorage)
                  → Refresh token (Redis, 30 days, revocable; FE memory-only)
                  → WS: POST /auth/ws-ticket → GET /ws Sec-WebSocket-Protocol hub.ws.ticket.*
                    (?token= denied; ADR-WS-JWT-HANDSHAKE-2026)
                  → Session record (PostgreSQL)
Telegram         → Только notify bind/delivery (НЕ login)
```

| Контроль | Значение | Источник |
|----------|----------|----------|
| Login methods | **Magic Link only** | ADR-AUTH-01, `auth/handler.go` |
| Auth rate limit | 3 / 10 min | `ratelimit.go` |
| Magic link TTL | 10 min | `magiclink.go` |
| Code compare | `subtle.ConstantTimeCompare` | SEC-07 ✅ |
| Refresh storage | SHA-256 hash in Redis | NF-12 ✅ |
| Session list | User видит device/IP, может отозвать | `/auth/sessions` |
| Admin Deactivate/Ban → WS kill | Redis invalidate + NATS `auth.session.{uid}.revoked` → Bridge UnregisterBySession (C-W19-001); fail-closed при ошибке kill (C-W19-003) | `audit/service.go` + `auth.SessionService.RevokeAllUserSessions` |
| RBAC | System (Admin/Member) → Project roles → Custom | `rbac/` |

**Запрещено (v1–v4 без ADR):** пароли; OAuth-провайдеры соцсетей.  
**SSO/LDAP:** одобрено в scope (2026-07-03); реализация только после принятого `ADR-AUTH-SSO`. До принятия ADR вход — **Magic Link only**.

## Валидация и санитизация

| Вход | Правило |
|------|---------|
| UUID params | `platform/http.ParseUUIDParam` |
| Request body | Fiber BodyParser + validate tags, 25MB limit |
| User URLs | `safehttp.ValidateOutboundURL` before fetch |
| Webhook bodies | Signature required, body hash mandatory |
| File uploads | Presigned PUT, RBAC on presign generation |
| Mongo search | Must filter by user's channel/project IDs |

### Mailbox + Calendar RSVP (future contract)

Security controls for the planned mailbox engine, HTML rendering, remote-image proxy,
attachments, mailbox-owner isolation, iTIP sender binding, parser bounds, rate limits,
admin RBAC/audit and secret handling are frozen in
`.planning/decisions/ADR-MAIL-01-mail-engine-calendar-sync.md` § Security contract and
`.planning/decisions/specs/MAIL-SPEC.md` §7. They are **not implemented or runtime-
verified** and therefore do not inherit the production-readiness statements in this
historical document. Current Resend also exposes a legacy admin broadcast that conflicts
with ADR-NOTIF-01's no-marketing/whitelist policy; treat it as an open governance/security
gap until removed, disabled, or explicitly re-authorized and audited.

Mail 0.1.4 additionally requires authenticated organizer alignment and immutable organizer
identity for REQUEST/CANCEL, capped lifetime tombstones plus bounded receipt retention,
reconciliation before an
outward REPLY retry, all-or-nothing self/shared-project-or-channel availability authorization
without admin bypass, metadata-only reminder
events, fail-closed retriable preference/repository lookup, stable user-row serialization
for absent/existing preference and quiet-hours changes, last-chance reminder state/config/
eligibility reauthorization, durable `policy_invalid` suppression with no UTC/delivery
fallback for malformed policy, post-commit-only external reminder channels with no send on
rollback, and independent deal-owner/
mailbox-owner checks for CRM draft/message routes and links. Recipient data is allowed only
in the ordinary draft body, never URL/history/logs; a reader's owned-deal filter is not
server authorization. Conditional L1 feed
URLs are encrypted/write-only and still pass safe-dial/redirect/decompression checks on
every fetch; L1 code remains blocked until founder decision 5.

Reaching the per-mailbox unbound calendar-state cap is fail-closed and visible: unseen UIDs
quarantine before mutation, known UIDs continue, and one rate-limited content-free ops
signal marks degraded import health. Automated tombstone eviction/reset is forbidden.

## Защита от CSRF/XSS/Clickjacking

| Угроза | Мера |
|--------|------|
| XSS | React escaping, CSP headers (v2.5) |
| CSRF | SameSite cookies, JWT in header not cookie |
| Clickjacking | `X-Frame-Options: DENY` (NF-08 ✅ `SecurityHeaders`) |
| CSWSH | WS origin check vs `CORS_ORIGINS` (NF-02 ✅) |

## Безопасность WebSocket (NF-02) — FIXED

**Реализовано (SEC-H1):**
1. `CheckOrigin` сверяет `Origin` с `CORS_ORIGINS` (`ws/handler.go:84-105`)
2. `Authorizer.CanSubscribe` проверяет `channel:/call:/project:` (`ws/authorizer.go:61-79`)
3. Неизвестные префиксы — fail-closed; `admin:*` не существует в bridge

**Остаточный риск:** `DEV_MODE=true` разрешает все Origin; пустой Origin пропускается (non-browser).

## Безопасность поиска (NF-03) — FIXED

**Реализовано (SEC-H2):** `searchMessages` фильтрует `channel_id ∈ getUserChannelIDs(userID)`.
Issues/projects/files/channels — по `projectIDs`; users — shared-project scope (+ admin bypass).

## Безопасность webhook

```go
// ПРАВИЛЬНО — reject if hash claim missing
expectedHash, ok := claims["sha256"].(string)
if !ok || expectedHash == "" {
    return fiber.ErrUnauthorized
}

// ЗАПРЕЩЕНО — skip if absent
if expectedHash != "" { ... }
```

Применяется к: LiveKit, GitHub, GitLab webhooks.

## Безопасное логирование

- ✅ `slog.InfoContext(ctx, "...", "request_id", id)`
- ✅ Redact tokens, passwords, magic link codes в query logs
- ❌ `log.Printf("%v", req.Body)` с sensitive data
- ❌ Логирование полного JWT

## Ограничение частоты запросов

| Endpoint type | Limit |
|---------------|-------|
| Auth | 3 / 10 min |
| API general | 100 / sec |
| File upload | 10 / min |
| Messages | 60 / min | `MessageSendRateLimit` в `ratelimit.go` |

## v2.0.1 Blockers — итог re-audit 2026-06-17

**CRITICAL (12/12): ✅ ALL FIXED**

**HIGH (16):**
- ✅ 14 FIXED (SEC-H1-H3, BE-H2-H7, FE-H1-H4)
- ⚠️ 1 PARTIAL (BE-H1 imex unbounded но BodyLimit; ARCH-H1 layering ✅ FIXED)
- ❌ 0 NOT FOUND
- 🔄 1 DEFERRED (FE-H5 bundle size — требует отдельного анализа)

**Production Readiness:** ✅ **YES с оговорками**

### Готовность к продакшену:

✅ **Можно деплоить** при условии:
1. Rotation скомпрометированных Resend/Telegram токенов (INF-03 history)
2. Периодический bundle size audit (FE-H5)
3. Мониторинг cron retention (`RunRetentionCleanup`) и GDPR audit-логов

### Остаточные риски (MEDIUM/LOW/INFO):

| ID | Риск | Mitigation |
|----|------|------------|
| BE-H1 | `imex/handler.go:66` unbounded ReadAll | Fiber BodyLimit 25MB срабатывает до ReadAll |
| BE-H6 | GDPR cascade partial errors | `DeleteUserGDPR` логирует и агрегирует через `errors.Join`; мониторить WARN |
| BE-H7 | Retention S3 fail-soft | При nil s3Client записи удаляются из PG, объекты S3 — WARN (dev) |
| NF-10 | 6-digit magic-link | Accepted: 10⁶ entropy + rate-limit 3/10m + TTL 10m |
| NF-13 | WS drop на full buffer | Accepted: fail-soft лучше блокировки room |
| FE-H5 | Bundle 1.1 MB | Deferred: code-splitting в Phase 17 (v2.5) |

## Чек-лист security review (для верификаторов)

- [ ] Нет hardcoded secrets в diff
- [ ] User-controlled URLs через safehttp
- [ ] Webhook signatures не skip
- [ ] RBAC/membership check на каждом endpoint
- [ ] Search/filter scoped to user
- [ ] WS rooms require AuthZ
- [ ] `io.LimitReader` для body reads
- [ ] `subtle.ConstantTimeCompare` для secrets
- [ ] Нет `panic()` в production paths
- [ ] Нет `*ent.Client` в handler/service

## Grep-паттерны (аудит 2026-06-17)

| Паттерн | Результат 2026-06-17 |
|---------|---------------------|
| `http.DefaultClient` | ✅ 0 usage в prod-коде (только комментарии) |
| `io.ReadAll` без `LimitReader` | ⚠️ `imex/handler.go:66`, `imex/jira_import.go:373` unbounded; смягчено BodyLimit 25MB |
| `!=` для секретов | ✅ Magic-link/GitLab webhook `ConstantTimeCompare`; LiveKit `iss` string compare — LOW |
| WS `CheckOrigin` + authz | ✅ FIXED (`ws/handler.go:84-105,218-231`) |
| Search user scope | ✅ FIXED (`search/service.go:585` channel_id filter) |
| Webhook signature skip | ✅ FIXED (LiveKit sha256 required; GitHub/GitLab reject empty; **A5-008/009**: HMAC/token bound to matched integration repo/path, not any-of instance secrets) |
| `*ent.Client` в handler | ⚠️ `auth/handler.go` service методы (рефакторинг Phase 19) |
| `panic()` в production | ✅ 0 найдено |
| `DEV_MODE=true` в CI | ✅ FIXED (deploy.yml:63-65 удалён) |
| Hardcoded credentials | ✅ FIXED (livekit-prod.yaml удалён, local.yml env placeholders) |
| `pg_advisory_xact_lock` | ✅ Используется (`auth/service_registration_lock.go:47`) |
| `safehttp.NewSafeHTTPClient` | ✅ `gitlab/service.go:55`, `platform/http/opengraph.go:30` |
| `ValidateOutboundURL` | ✅ `gitlab:71`, `notification/push.go:157,218`, `opengraph:26` |
| Offline AI loopback (Ollama) | ✅ ADR-AI-OLLAMA-LOOPBACK-2026: `ValidateOfflineAIURL` + `NewOfflineAIHTTPClient` только для host allowlist; глобальный SafeDial **не** ослаблен (P1-BE-U01 / TASK-101) |
| `subtle.ConstantTimeCompare` | ✅ `auth/magiclink.go:135` |
| ErrorBoundary usage | ✅ 7 модулей (messenger, dashboard, calls, tms, projects) |
| BroadcastChannel | ✅ `stores/auth-store.ts:46-88` multi-tab sync |
| Single-flight refresh | ✅ `api/client.ts:97-108` mutex pattern |

---

## Итоговое резюме — формальный re-audit 2026-06-17

**Scope:** All 12 CRITICAL + 16 HIGH findings from `.planning/audits/code-review-final/SUMMARY.md` (2026-05-15)  
**Method:** File-level verification (Grep + Read actual code) against current codebase  
**Auditor:** security-auditor (readonly analysis)

### Сводка статусов

| Severity | Total | ✅ FIXED | ⚠️ PARTIAL | ❌ NOT FOUND | 🔄 DEFERRED |
|----------|-------|---------|-----------|-------------|------------|
| **CRITICAL** | **12** | **12** | **0** | **0** | **0** |
| **HIGH** | **16** | **12** | **2** | **2** | **0** |
| **TOTAL** | **28** | **24** | **2** | **2** | **0** |

### CRITICAL (12) — все исправлено ✅

1. ✅ SEC-01: APIKeyAuth ordering
2. ✅ SEC-02: LiveKit webhook sha256
3. ✅ SEC-03: livekit-prod.yaml secret
4. ✅ SEC-04: GitLab SSRF
5. ✅ SEC-05: WS send-on-closed panic
6. ✅ SEC-06: First-admin race
7. ✅ SEC-07: Magic-link timing attack
8. ✅ INF-01: LiveKit departure_timeout
9. ✅ INF-02: DEV_MODE=true in CI
10. ✅ INF-03: Plaintext credentials
11. ✅ INF-04: BackupService nil s3Client
12. ✅ FE-01: Rules of Hooks violation

### HIGH (16) — разбивка статусов

**✅ FIXED (12):**
- SEC-H1: WS authz (CheckOrigin + CanSubscribe)
- SEC-H2: Search channel filter
- SEC-H3: WebPush SSRF
- BE-H2: respBuf.ReadFrom
- BE-H3: Dashboard user filter
- BE-H4: Goroutine-per-request
- BE-H5: /health DB connection pool
- FE-H1: Refresh-token race
- FE-H2: Multi-tab auth sync
- FE-H3: POST /auth/logout
- FE-H4: Error Boundaries

**⚠️ PARTIAL (1):**
- BE-H1: Unbounded io.ReadAll (imex handlers — mitigated by Fiber BodyLimit 25MB)

**✅ FIXED (ARCH-H1):**
- ARCH-H1: `auth/handler.go` — только `*AuthService`; без прямого `*ent.Client` (комментарий ARCH-H1 + `GetMe` → service)

**✅ FIXED (дополнительно к списку выше):**
- BE-H6: GDPR delete (`audit/service.go:376`, handler `audit/handler.go:364`)
- BE-H7: Retention cleanup (`audit/retention.go:98`, trash purge `audit/service.go:252`)

**🔄 DEFERRED (1):**
- FE-H5: Bundle size audit (requires build tooling analysis)

### Аудит новых фич (2026-06-17)

**Verified Safe:**
- ✅ `audit/middleware.go`: Async recorder, proper skip logic
- ✅ `audit/handler.go:284` BanUser: Admin check + audit log
- ✅ `middleware/security_headers.go`: X-Frame-Options, CSP, nosniff
- ✅ HSTS (`Strict-Transport-Security`): `deployments/Caddyfile.prod` — TLS edge only (A2W4-1-001); Fiber не дублирует HSTS за reverse-proxy
- ✅ AES-GCM AAD: `crypto.EncryptWithAAD` / `DecryptWithAAD` — привязка ciphertext к entityID (A2W4-4-001); legacy `Encrypt`/`MaybeDecryptString` — migration window без AAD
- ✅ `observability/timing.go`: No secret logging
- ✅ `storage/filename.go`: Path traversal mitigation
- ✅ `platform/http/opengraph.go`: SafeHTTP + ValidateOutboundURL + LimitReader

### Вердикт SEC-PROD

**Production Ready:** ✅ **YES**

**Conditions:**
1. Rotate compromised Resend/Telegram tokens (INF-03 commit history)
2. Monitor GDPR/retention cron logs in production
3. Periodic bundle size audit (FE-H5)

**Risk Profile:**
- 0 OPEN CRITICAL vulnerabilities
- 1 PARTIAL HIGH (BE-H1 mitigated by BodyLimit)
- 0 NOT FOUND HIGH

**Recommendation:** Clear for production deployment. Residual risks are LOW severity or architectural debt scheduled for v2.6 Phase 19-20.

---
*Последнее обновление: 2026-06-22 — RU-FIX-01: RU section headers*
