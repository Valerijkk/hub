# W1-auth — Security Audit `backend/internal/auth/`

> **Provenance class:** PRE_EXISTING · **Confidence:** MIXED · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** security-auditor + be-ver (readonly)  
> **Дата:** 2026-07-23  
> **Scope:** `backend/internal/auth/*`, wiring `cmd/server/wiring_auth.go`, schema `ent/schema/{organization,invite}.go`, конфиг JWT/rate-limit  
> **Код не менялся**

---

## Summary

Публичный HTTP-контур auth соответствует **Magic Link only**: маршруты `/auth/telegram*` сняты (`handler.RegisterRoutes` + комментарий AUTH-01 в `wiring_auth.go`). JWT access **15m**, refresh **Redis 30d** (SHA-256 at rest), rate-limit auth **3/10min** (Redis, prod), magic-link код сравнивается через **`subtle.ConstantTimeCompare`**.

Главный блокер доступа (ORG-06): **pending-пользователь получает JWT сразу после verify** и проходит `RequireAuth` (блокируется только `deactivated`). Очередь admin approve фактически не гейтит API.

Остаточный риск ADR-AUTH-01: живые методы `LoginWithTelegram` / `LoginWithTelegramUser` / `TelegramBotAuth.VerifyCode` + facade на `AuthService` — HTTP не экспонирует, но поверхность и тесты login-path сохранены.

**Вердикт:** `verified: false` — 0 CRITICAL на публичном bypass Telegram, но **1 HIGH (pending→session)** блокирует «auth TZ PASS»; Telegram login = FAIL по духу ADR (код жив), PASS по HTTP-поверхности.

---

## Findings

| Severity | file:line | Issue | Evidence | Fix |
|----------|-----------|-------|----------|-----|
| **HIGH** | `service_login.go:48-65` + `middleware.go:81-92` + `repository.go:386-397` | ORG-06 bypass: invite_only создаёт `StatusPending`, но `LoginWithMagicLink` сразу `CreateSession`; `RequireAuth` режет только `deactivated` | После verify: `CreatePendingMember` → `CreateSession`. Middleware: `deactivated := u.Status == user.StatusDeactivated` — `pending` = active path. Повторный login существующего pending тоже выдаёт токены | После verify: если `u.Status != active` → 403 без сессии; в `RequireAuth`/`RefreshToken` fail-closed на `pending` (+ кэш статуса); тест denial: pending не получает Bearer |
| **MEDIUM** | `service_login.go:94-125`, `service.go:106-114`, `telegram.go`, `telegram_bot.go:28-162` | ADR-AUTH-01: Telegram **login** API снят с роутера, но бизнес-логика login жива (initData → session, bot code → session) | `handler.go:31-56` — нет `/telegram`; `wiring_auth.go:87-89` AUTH-01. Facade/`LoginService` всё ещё создают сессию. Тесты `testframework/auth/service_login_test.go` гоняют login | Удалить/заглушить login-методы (или `//go:build` только tests); оставить bind/notify в `notification`; regression: grep `LoginWithTelegram` = 0 в prod paths |
| **MEDIUM** | `magiclink.go:16-22`, `magiclink.go:46-52` | NF-10: 6-значный код (~10⁶) — accepted risk, но entropy низкая при компромиссе rate-limit / multi-instance edge | `generateCode` → `1_000_000`; TTL 5m; send 3/10m; verify fails 5; доп. IP/email limiter в wiring | Оставить accepted **или** увеличить алфавит/длину; убедиться, что Redis rate и verify-fails не деградируют при Redis errors (сейчас fails-counter error → warn и продолжение) |
| **MEDIUM** | `magiclink.go:85-88` | Magic-link код хранится в Redis **plaintext** (в отличие от refresh SHA-256) | `Set(codeKey, code, TTL)` vs `hashRefreshToken` в session | Хранить HMAC/SHA-256(code) как refresh; compare constant-time с хешем ввода |
| **MEDIUM** | `wiring_auth.go:55-71` + `middleware/ratelimit.go:28` | Auth RL 3/10min на **весь** `/auth` (включая `/refresh`, `/logout`) — жёсткий IP-bucket; в `DevMode` RL **выключен** | Prod: `Use("/auth", AuthRateLimit)`. Dev: ветка `if !DevMode` пропускает лимитер | Исключить `/refresh` из общего bucket или отдельный лимит; в staging с `DevMode=false` держать RL всегда |
| **LOW** | `magiclink.go:90-97`, `handler.go:98-101` | Dev: код в slog + JSON `code` при `resend==nil`; prod без Resend → IsDevMode=true даже при `DEV_MODE=false` | `IsDevMode() = resend == nil`; validate() **не** требует `RESEND_API_KEY` | В `DEV_MODE=false` запретить nil Resend или не возвращать `code` в HTTP никогда |
| **LOW** | `service_session.go` + `middleware.go:76-97` | Revoke/logout инвалидируют refresh+DB session, но access JWT живёт до 15m (нет session-denylist) | `RequireAuth` только Parse JWT + deactivated | Принято при TTL 15m **или** Redis denylist `sid` / jti на revoke |
| **LOW** | `service_organization.go:52-63`, `repository.go:257-260` | Invite token: DB `InviteTokenEQ` без constant-time (oracle через Exist timing) | Уникальный токен 32 hex; сравнение через SQL | Принять (DB) **или** загружать org.token и `subtle.ConstantTimeCompare` |
| **LOW** | `ent/schema/organization.go:20-42` | Нет поля `version` (TZ optimistic lock на изменяемых схемах) | Fields: id, name, slug, invite_token, timestamps — без `version` | Добавить `version` + update path с conflict |
| **INFO** | `telegram_bot.go:112-161` | Legacy `VerifyCode` без `ConstantTimeCompare` (код = ключ Redis) | Мёртвый HTTP-путь; polling disabled | Удалить вместе с login-path |
| **INFO** | `resend.go:30-33,79` | Исходящий HTTP к Resend: свой `http.Client`+timeout; `LimitReader` на error body — OK (URL не user-controlled) | Не `DefaultClient` | — |
| **PASS** | `config.go:171-174`, `service_session.go:65-78,98-105,216` | JWT 15m / refresh 30d Redis / hash at rest / ConstantTimeCompare на refresh hash | Wiring передаёт `AccessTokenTTL`/`RefreshTokenTTL` из config | — |
| **PASS** | `handler.go:31-56`, `wiring_auth.go:87-89` | HTTP: только magic-link + refresh + protected; Telegram login не зарегистрирован | RegisterRoutes без telegram | Держать; не возвращать роуты |
| **PASS** | `magiclink.go:17-20,62-78,123` | Rate send 3/10m + ConstantTimeCompare verify | Согласовано с TZ + AuthRateLimit | — |
| **PASS** | `dev/handler.go` + `//go:build !release` | Dev login без magic-link вырезан из release-бинарника | `handler_release.go` no-op | — |

---

## Checklist TZ (auth)

| Инвариант | Статус | Комментарий |
|-----------|--------|-------------|
| Magic Link only (единственный login) | **FAIL\*** | HTTP PASS; service-layer Telegram login ещё исполняем |
| Telegram ≠ login (ADR-AUTH-01) | **FAIL\*** | Notify/bind — вне этого пакета; login-код в auth остаётся |
| JWT access TTL 15 min | **PASS** | `config.JWT.AccessTokenTTL = 15 * time.Minute` |
| Refresh Redis, 30 days, instant revoke | **PASS** | TTL 30d; Del Redis + DeleteSession; hash at rest |
| Auth rate 3 / 10 min | **PASS** | `AuthRateLimit` + per-email magiclink 3/10m (prod) |
| `ConstantTimeCompare` / `hmac.Equal` на секретах | **PASS** | magic code, refresh hash; TG initData hash via `hmac.Equal` |
| Invite + Pending queue + 7d purge | **FAIL** | Pending создаётся, purge cron есть; **сессия до approve** |
| `system_role founder` один на org | **PASS** (практика) | First-user advisory lock SEC-06; profile не меняет role |
| Нет hardcoded credentials в auth | **PASS** | Секреты из env/config |
| Layering: нет `*ent.Client` в handler/service | **PASS** | Через `Repository`; ARCH-H1 в handler |

\*FAIL по канону TZ/ADR при наличии исполняемого login-path; для внешнего атакующего без нового wiring — HTTP-only PASS.

---

## 5 top risks

1. **Pending → полный Bearer (HIGH)** — invite_only ORG-06 не enforced на authz; любой с валидным invite + email ownership сразу в API как `member`.
2. **Telegram login dead-but-callable (MEDIUM)** — случайный/враждебный wiring или внутренний вызов `LoginWithTelegram*` снова открывает второй auth factor вопреки ADR-AUTH-01.
3. **6-digit magic code + plaintext Redis (MEDIUM)** — при утечке Redis или ослаблении RL возможен offline/online перебор; mitigation: rate + TTL + (желательно) hash-at-rest.
4. **Resend/nil → «dev» leak кода (LOW→ops HIGH)** — пустой `RESEND_API_KEY` при `DEV_MODE=false` включает `IsDevMode` и отдаёт код в JSON/логах.
5. **Access JWT после revoke до 15m (LOW)** — украденный access после logout/revoke сессии ещё валиден; риск ограничен TTL.

---

## Verification

| Этап | Результат |
|------|-----------|
| Scope files прочитаны | auth handler/session/magiclink/login/org/registration/middleware/telegram*, wiring_auth, schemas |
| Grep §6 (DefaultClient / ReadAll / ConstantTime / telegram login / DEV) | DefaultClient в auth — нет; ReadAll+LimitReader в resend — OK; TG login routes — 0; LoginWithTelegram symbols — да |
| Severity не занижен | pending-session = HIGH (A01/ORG-06), не MEDIUM |
| Denial-полнота | Не хватает тестов «pending не логинится / 403 на API» |
| verified | **false** — 2026-07-23 |

**Handoff:** `be1`/`be2` — фикс pending gate + вынос/удаление Telegram login; повторный проход security-auditor.
