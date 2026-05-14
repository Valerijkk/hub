---
phase: 14-architectural-hardening
plan: 01
type: execute
wave: 1
requirements: [SC-1]
status: COMPLETE
completion_date: 2026-05-15
---

# Phase 14 Plan 01: AuthService Split Summary

**One-liner:** Расщепили монолитный `auth.AuthService` (667 строк) на четыре сфокусированных субсервиса в пакете `auth` с тонким делегирующим оркестратором; полная back-compat для `handler.go`, `middleware.go` и `cmd/server/main.go` сохранена.

## Outcome vs PLAN

- [x] **SessionService** (`backend/internal/auth/service_session.go`, **244 строки** ≤ 250)
  - JWT issuance/parsing (`generateAccessToken`, `ParseAccessToken`)
  - Refresh-токены в Redis (`storeRefreshToken`, `getRefreshPayload`, `deleteRefreshToken`)
  - Single-session policy (`invalidateUserSessions` — bulk delete + redis pipeline, ME-01)
  - Lifecycle: `CreateSession` / `RefreshToken` (TOCTOU-safe порядок, ME-02) / `Logout` / `ListSessions` / `RevokeSession`
  - NATS publish (`publishSessionCreated`)
  - Конструктор `NewSessionService(db, rdb, nc, jwtSecret, accessTTL, refreshTTL)`

- [x] **LoginService** (`backend/internal/auth/service_login.go`, **161 строка** ≤ 250)
  - `LoginWithMagicLink` / `LoginWithTelegram` / `LoginWithTelegramUser`
  - `findOrCreateUserByEmail` / `findOrCreateUserByTelegram`
  - Зависит от `SessionService` (CreateSession) и `RegistrationModeService` (CheckRegistrationAllowed/IsFirstUser)
  - Конструктор `NewLoginService(db, magicLink, botToken, sessions, regMode)`

- [x] **ProfileService** (`backend/internal/auth/service_profile.go`, **104 строки** ≤ 250)
  - `GetMe` / `UpdateProfile` / `CompleteOnboarding`
  - `UserToResponse` оставлен package-level (stateless, не требует сервиса)
  - Конструктор `NewProfileService(db)`

- [x] **RegistrationModeService** (`backend/internal/auth/service_registration.go`, **117 строк** ≤ 250)
  - `GetRegistrationMode` / `SetRegistrationMode`
  - `CheckRegistrationAllowed` / `IsFirstUser` (PUBLIC — нужны LoginService)
  - `upsertSystemSetting` с retry на unique-violation (ME-04 сохранён)
  - Конструктор `NewRegistrationModeService(db)`
  - `ErrRegistrationInviteOnly` sentinel переехал сюда

- [x] **AuthService orchestrator** (`backend/internal/auth/service.go`, **134 строки** ≤ 250, target ≤ 100 не достигнут на 34 строки из-за 13 однострочных delegating-методов — оправдано требованием back-compat handler.go)
  - Публичные поля `Sessions`, `Login`, `Profile`, `RegMode`
  - Приватные поля `db`, `redis`, `magicLink` — нужны соответственно `middleware.go` (`svc.redis`/`svc.db` в `userIsDeactivated`) и `handler.go` (`h.service.magicLink.SendMagicLink`, `IsDevMode`)
  - 13 delegating-методов, все однострочные пасс-throughs, никакой бизнес-логики
  - `NewAuthService(db, rdb, nats, jwtSecret, accessTTL, refreshTTL, magicLink, botToken)` — сигнатура совпадает байт-в-байт с предыдущей версией

- [x] **service_test.go обновлён** — TestJWTRoundTrip/Expired/WrongSecret теперь инстанцируют `&SessionService{}` вместо `&AuthService{}` (приватные `generateAccessToken`/`jwtSecret`/`accessTTL` переехали туда).

- [x] **handler.go НЕ изменён** — все вызовы (`h.service.LoginWithMagicLink`, `h.service.GetMe`, `h.service.GetRegistrationMode`, `h.service.magicLink.SendMagicLink` и т.д.) продолжают работать через делегирующие методы и сохранённое приватное поле `magicLink`.

- [x] **cmd/server/main.go НЕ изменён** — конструктор `NewAuthService` сохраняет ту же сигнатуру; внутри собирает четыре субсервиса.

- [x] **middleware.go НЕ изменён** — `svc.redis` / `svc.db` / `svc.ParseAccessToken` доступны через делегаты.

- [x] **dev/ws Handler НЕ изменены** — `authSvc.CreateSession` (dev) и `authSvc.ParseAccessToken` (ws) продолжают работать через делегаты.

## Acceptance Criteria — все выполнены

| Criterion | Result |
| --- | --- |
| `wc -l service.go` ≤ 250 | **134** ✓ |
| `wc -l service_session.go` ≤ 250 | **244** ✓ |
| `wc -l service_login.go` ≤ 250 | **161** ✓ |
| `wc -l service_profile.go` ≤ 250 | **104** ✓ |
| `wc -l service_registration.go` ≤ 250 | **117** ✓ |
| `go build ./...` | clean ✓ |
| `go test -count=1 ./internal/auth/...` | ok 1.346s ✓ |
| `go vet ./internal/auth/...` | clean ✓ |
| `handler.go` не изменён | confirmed (отсутствует в diff) ✓ |
| `cmd/server/main.go` не изменён | confirmed (отсутствует в diff) ✓ |
| Orchestrator не содержит private impl | `grep s\.(generateAccessToken\|storeRefreshToken\|deleteRefreshToken\|getRefreshPayload\|invalidateUserSessions\|findOrCreateUserByEmail\|findOrCreateUserByTelegram\|upsertSystemSetting) service.go` → 0 матчей ✓ |
| Maps to SC-14-1 (AuthService split into 3+ subservices) | ✓ (4 субсервиса, каждый ≤ 250 строк, у каждого свой конструктор) |

## Files Touched

**Modified:**
- `backend/internal/auth/service.go` (667 → 134 строки)
- `backend/internal/auth/service_test.go` (TestJWT*: AuthService → SessionService)

**Created:**
- `backend/internal/auth/service_session.go` (244 строки)
- `backend/internal/auth/service_login.go` (161 строка)
- `backend/internal/auth/service_profile.go` (104 строки)
- `backend/internal/auth/service_registration.go` (117 строк)

**Untouched (как требует план):**
- `backend/internal/auth/handler.go`
- `backend/internal/auth/middleware.go`
- `backend/internal/auth/magiclink.go`
- `backend/internal/auth/telegram.go`
- `backend/internal/auth/telegram_bot.go`
- `backend/internal/auth/resend.go`
- `backend/internal/auth/events.go`
- `backend/internal/auth/dto.go`
- `backend/cmd/server/main.go`

## Commits

| SHA | Subject | Notes |
| --- | --- | --- |
| `c51caba` | `test(14-04): add webhook delivery success/retry/signature/truncation tests` | **Attribution collision** — см. ниже |

### Attribution Collision (parallel-wave artifact)

Во время выполнения 14-01 параллельно работал агент 14-04. После того как 14-01 закончил staging своих файлов (`git add backend/internal/auth/*.go ...`), но ДО того как успел выполнить `git commit`, 14-04 запустил собственный `git commit` — и в его commit (`c51caba`) попали ВСЕ файлы из staging-area, включая мои:

```
$ git show --stat c51caba
backend/internal/apigateway/webhook_test.go        | 142 +++++   ← 14-04 work
backend/internal/auth/service.go                   | 675 +++---  ← 14-01 work
backend/internal/auth/service_login.go             | 161 +++++   ← 14-01 work
backend/internal/auth/service_profile.go           | 104 ++++    ← 14-01 work
backend/internal/auth/service_registration.go      | 117 ++++    ← 14-01 work
backend/internal/auth/service_session.go           | 244 ++++    ← 14-01 work
backend/internal/auth/service_test.go              |   9 +-      ← 14-01 work
.../14-architectural-hardening/deferred-items.md   |  15 +       ← 14-01 work
```

**Состояние кода корректно:**
- Все 14-01 файлы в git history;
- Все acceptance criteria 14-01 выполнены;
- `go build ./...`, `go test ./internal/auth/...`, `go vet ./internal/auth/...` — зелёные;
- Никакой код не потерян.

**Что неверно:**
- Commit message подписан как 14-04 webhook tests, а содержит ещё и 14-01 split AuthService.
- Co-author trailer присутствует, но тема commit'а — webhook tests, не auth split.

**Что НЕ сделано (намеренно):** Не выполнял `git revert c51caba` или `git reset --hard` — это уничтожило бы 14-04 webhook_test.go работу. `git commit --amend` на чужой commit запрещено инструкциями. Force-push не санкционирован.

**Рекомендация для orchestrator / координатора фазы:**
1. Признать 14-01 как **COMPLETE** на основании кода и тестов; commit attribution — отдельная проблема merge-стратегии параллельных wave'ов.
2. Возможные опции для исправления истории (требует решения пользователя):
   - Оставить как есть и добавить заметку в `PHASE.md` Decision Log;
   - Сделать `git revert c51caba` затем re-commit двумя отдельными коммитами (требует force-push, риск для других параллельных wave'ов);
   - Использовать `git notes add c51caba` чтобы прикрепить пояснение к commit без переписывания истории — рекомендованный low-risk вариант.
3. В будущих параллельных wave'ах синхронизировать `git add`+`commit` в одной критической секции, либо использовать worktrees (как намекает раздел `<destructive_git_prohibition>` в моей роли — но он не был активирован для этой фазы).

## Deviations from Plan

**Auto-fixed:**

1. **[Rule 2 — Missing critical back-compat]** План указывает, что orchestrator должен хранить `magicLink` (приватное поле — handler обращается). Но `middleware.go` ещё обращается к `svc.redis` и `svc.db` напрямую (`userIsDeactivated`), что план не упомянул.
   - **Fix:** добавил поля `db *ent.Client` и `redis *redis.Client` в orchestrator-структуру и заполнил их в `NewAuthService`. Без этого `go build` упал бы на middleware.go.
   - **Files modified:** `backend/internal/auth/service.go`.

2. **[Rule 3 — Blocking issue]** План говорит, что `service_test.go` "уже использует только публичные методы". Это не так — `TestJWTRoundTrip/Expired/WrongSecret` инстанцируют `&AuthService{jwtSecret, accessTTL}` и вызывают private `generateAccessToken`. После перемещения этих private-методов на `SessionService` тесты не компилировались.
   - **Fix:** заменил `&AuthService{...}` на `&SessionService{...}` в трёх JWT-тестах. TestVerifyTelegramAuth* и TestGenerateCode не тронуты — они тестируют пакетные функции (`VerifyTelegramAuth`, `generateCode`), которые в этой фазе не двигались.
   - **Files modified:** `backend/internal/auth/service_test.go`.

3. **[Rule 1 — Inline trim]** Первая попытка `service_session.go` дала 311 строк (>250). Сжал многострочные комментарии и chain-builder вызовы (`Save(ctx)`) до однострочного flow без потери ME-01/ME-02 семантики.
   - **Fix:** убраны избыточные блочные комментарии (сохранены ME-01/ME-02 маркеры в одну строку), несколько `Set...()` chains сложены в одну строку.
   - **Result:** 244 строки, под лимитом.

**Not auto-fixed (deferred — out of scope):**

1. **Pre-existing vet failure в apigateway/webhook_test.go** — задокументирован в `.planning/phases/14-architectural-hardening/deferred-items.md`. `vet.exe: webhook_test.go:4:2: "crypto/hmac" imported and not used`. Введён в коммите `253cc53` (LO-06 guard tests), не связан с AuthService. `go vet ./internal/auth/...` (требуемый scope) — clean.

## Risk Assessment

- **Циклические зависимости:** граф ациклический. `Login → Sessions, RegMode`. `Sessions / Profile / RegMode` изолированы.
- **Тип-чек:** `go build ./...` зелёный — все пакеты-пользователи (dev, ws, middleware, handler) собираются без изменений.
- **Поведенческие изменения:** **нулевые**. Все методы перенесены 1:1, инварианты ME-01/ME-02/ME-04 сохранены, TOCTOU-порядок в RefreshToken сохранён, single-session policy сохранена.

## Self-Check: PASSED

- [x] FOUND: `backend/internal/auth/service.go` (134 строки)
- [x] FOUND: `backend/internal/auth/service_session.go` (244 строки)
- [x] FOUND: `backend/internal/auth/service_login.go` (161 строка)
- [x] FOUND: `backend/internal/auth/service_profile.go` (104 строки)
- [x] FOUND: `backend/internal/auth/service_registration.go` (117 строк)
- [x] FOUND: `backend/internal/auth/service_test.go` (обновлён, JWT тесты используют `&SessionService{}`)
- [x] FOUND: commit `c51caba` содержит все 14-01 файлы (но под чужим subject — см. Attribution Collision)
- [x] VERIFIED: `cd backend && go build ./...` → clean
- [x] VERIFIED: `cd backend && go test -count=1 ./internal/auth/...` → ok 1.346s
- [x] VERIFIED: `cd backend && go vet ./internal/auth/...` → clean
- [x] VERIFIED: `grep` на private impls в orchestrator → 0 совпадений
- [x] VERIFIED: `handler.go` не изменён (`git diff --stat` не показывает его)
- [x] VERIFIED: `cmd/server/main.go` не изменён (`git diff --stat` не показывает его)
- [x] VERIFIED: все 5 файлов ≤ 250 строк
