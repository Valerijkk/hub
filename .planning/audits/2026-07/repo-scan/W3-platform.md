# W3-platform — `backend/internal/platform`

> **Provenance class:** FALLBACK_SUMMARY · **Confidence:** PRELIMINARY · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** be-ver + security-auditor (readonly) · **Дата:** 2026-07-23
> **Scope:** outbox, nats, middleware, safedialer, config · **Код не менялся**

## Summary

Платформенный слой в целом **PASS** по TZ: SSRF-guard (`ValidateOutboundURL`/`SafeDialer`), rate-limit константы (API 100/s, files 10/min, messages 60/min), `ErrorHandler` с `{"error","request_id"}`, outbox+NATS JetStream. `*ent.Client` в outbox Worker/Publisher — **допустимо** (infra, не domain Service/Handler).

**Вердикт:** `verified: false` · **historical_self_check:** PASS (platform-core scope) · CRITICAL/HIGH = 0.

## Findings

| Severity | file:line | Issue | Fix |
|----------|-----------|-------|-----|
| **PASS** | `http/safedialer.go:77+` | `ValidateOutboundURL` — канон SSRF | Держать единственной точкой |
| **PASS** | `http/opengraph.go:26,48` | OG fetch: Validate + `LimitReader` | — |
| **PASS** | `middleware/ratelimit.go:29-31` | API/File/Message лимиты = TZ | — |
| **PASS** | `middleware/middleware.go:13-15` | ErrorHandler → error + request_id | — |
| **PASS** | `outbox/publisher.go` + `nats/` | JetStream via outbox, не Redis pub/sub | — |
| **INFO** | `outbox/worker.go:18-23` | Worker держит `*ent.Client` — infra-допустимо | Не копировать в domain service |
| **INFO** | `outbox/publisher.go:41` | `Publish` принимает `*ent.Client` (tx/client) | Ок для outbox |
| **LOW** | `outbox/worker.go:75-78` | recover+log на panic в worker | Метрика/alert на recover |
| **MEDIUM** | домены вне platform | Messenger и др. публикуют NATS **без** outbox (см. W1-messenger) | Outbox в durable path |

## Checklist

| Инвариант | Статус |
|-----------|--------|
| SSRF ValidateOutboundURL | PASS |
| Rate limits TZ | PASS |
| Error JSON request_id | PASS |
| NATS JetStream канон | PASS |
| Нет DefaultClient/panic в platform handlers | PASS |
