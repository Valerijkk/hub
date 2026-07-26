# W1 — Security cross-cut (readonly)

> **Provenance class:** PRE_EXISTING · **Confidence:** MIXED · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> Агент: `security-auditor` (Grok 4.5) · Дата: **2026-07-23** · Scope: SSRF (`safedialer`), outbound webhooks (`apigateway/webhook.go`), grep `http.DefaultClient` / `io.ReadAll` / timing secrets / `DEV_MODE` в deploy.
> Код **не менялся**.

## Чек-лист (результат)

| Контроль | Статус | Evidence |
|----------|--------|----------|
| SSRF: `ValidateOutboundURL` + `SafeDialContext` | PASS | `safedialer.go:80-145`; webhook: `webhook.go:39,286-300` |
| DNS rebinding (резолв перед dial) | PASS | `safedialer.go:113-125` |
| Body DoS: `LimitReader` / не голый `ReadAll` в prod | PASS | все `io.ReadAll` в `backend/internal/**` обёрнуты `LimitReader`; webhook resp: `webhook.go:310-311` |
| Timing: секреты не через `==`/`!=` | PASS | сравнение секретов — `subtle.ConstantTimeCompare` / `hmac.Equal` (auth, gitlab, github, calls, mail, audit) |
| Webhook outbound HMAC + SSRF | PASS | `webhook.go:242,286-296,317-321` |
| `http.DefaultClient` в prod | PASS | 0 вызовов в `backend/internal/**` (только комментарии) |
| `DEV_MODE=true` в prod/deploy | PASS | `docker-compose.prod.yml:205` → `"false"`; `deploy.yml` — нет `DEV_MODE` |

## Findings

| Sev | file:line | Уязвимость / наблюдение | Рекомендация | OWASP/ASVS | Статус |
|-----|-----------|-------------------------|--------------|------------|--------|
| LOW | `backend/internal/platform/http/safedialer.go:135-145` | `NewSafeHTTPClient` не задаёт `CheckRedirect`. Дефолт Go следует до 10 редиректов. Private/metadata блокируются повторным `SafeDialContext` на каждом dial (`:113-125`), классический SSRF в RFC1918 закрыт. Остаточный риск: цепочка редиректов на **публичный** хост (эксфильт outbound webhook body + HMAC, DoS на follow). | Для webhook-клиента: `CheckRedirect: func(...) error { return http.ErrUseLastResponse }` или лимит 0 редиректов; опционально общий hardening в `NewSafeHTTPClient`. | A10 SSRF / ASVS V12.6 | open (hardening) |
| INFO | `.github/workflows/ci.yml:225` | `DEV_MODE: 'true'` в job Full E2E — ожидаемо для тестового контура, **не** prod-путь. | Не переносить в `deploy.yml` / `docker-compose.prod.yml`. Оставить как есть. | A05 Misconfig | accepted |
| — | — | CRITICAL/HIGH по scope **не найдены** | — | — | — |

### Закрытые / подтверждённые инварианты (не findings)

1. **`safedialer.go:31-51,65-75`** — blocked CIDRs включают RFC1918, loopback, link-local (`169.254.0.0/16` metadata), CGNAT, IPv6 ULA/link-local/multicast.
2. **`safedialer.go:80-95`** — только `http`/`https`, reject `userinfo`, empty host.
3. **`safedialer.go:121-125`** — если **любой** A/AAAA в blocked → dial abort (fail-closed до выбора `ips[0]`).
4. **`apigateway/webhook.go:286-311`** — `ValidateOutboundURL` → `webhookHTTPClient.Do` → `io.LimitReader(..., 4096)` (CR-03 / BE-H2).
5. **`apigateway/webhook.go:350-353,390-393`** — URL валидируется на Create/Update endpoint.
6. **`apigateway/webhook.go:451-459`** — `toWebhookResponse` **не** отдаёт `Secret` в API.
7. **Inbound webhook verify (смежный контроль, не outbound-файл):** GitHub `github/handler.go:100-124` — signature mandatory + `hmac.Equal`; GitLab `gitlab/handler.go:90-100` — token mandatory + `ConstantTimeCompare`; LiveKit `calls/handler.go:261-268` — claim `sha256` mandatory (SEC-02).
8. **`http.DefaultClient`:** hits только `notification/push.go:21` (комментарий «вместо DefaultClient») и `testframework/github/service_http_test.go:11` (тест). Prod push использует `NewSafeHTTPClient` (`push.go:25`).
9. **`io.ReadAll` без `LimitReader` в `backend/internal/`:** 0. Неограниченный `ReadAll` есть только в `testframework/**` (чтение тестовых HTTP-ответов) — вне prod attack surface.
10. **`!=`/`==` на секретах:** grep по `secret|token|password|signature|apiKey` — только empty-checks (`token == ""`), не сравнение значений секретов.
11. **`DEV_MODE`:** `deployments/docker-compose.prod.yml:205` = `"false"`; `.github/workflows/deploy.yml` — совпадений нет; prod compose не включает `DEV_MODE=true`.

## Grep-лог (кратко)

```
http.DefaultClient     → 0 prod usage (comments/tests only)
io.ReadAll (internal/) → все с io.LimitReader
DEV_MODE               → ci e2e true (accepted); prod compose false; deploy.yml clean
ConstantTime / hmac    → auth magiclink/session, gitlab, github, calls, mail, audit present
```

## Вердикт

- **CRITICAL:** 0
- **HIGH:** 0
- **MEDIUM:** 0
- **LOW:** 1 (CheckRedirect hardening)
- **INFO/accepted:** 1 (`DEV_MODE` в CI E2E)

**verified: false** · **historical_self_check:** PASS (cross-cut scope) — 2026-07-23 00:56 UTC+3

Блокеров релиза по этому cross-cut нет. LOW — передача `be1`/`be2` как optional hardening (не гейтит merge).
