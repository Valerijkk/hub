# A3 W6 агент A2 — auth remaining tests (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`)  
> Baseline: [A2W6-2-auth-tests.md](A2W6-2-auth-tests.md) · Stage A: [A6-2-auth-tests.md](A6-2-auth-tests.md)  
> Remediation: TASK-102 (Telegram login removal) · org mutation asserts (W6-A2-004/005)  
> Периметр: `W6-ASSIGNMENTS.md` §A2 (13) · Фокус: mutation gaps · telegram-as-login absence · rune length  
> Прочитано: 13 (+1 denial replacement) · Пропущено: 0

## Сверка с A2 (W6-A2-*/A2W6-A2-*)

| A2 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| W6-A2-001 / A2W6-A2-001 | FIXED-OK | закрыт | Позитивный `HandleLogin_StoresCode` / `tglogin:` удалён; denial `telegram_login_removed_task102_test.go` |
| W6-A2-002 / A2W6-A2-002 | FIXED-OK | закрыт | Login VerifyCode / bot coverage+polling suites удалены; `resend_telegram_test.go` — только Resend |
| W6-A2-003 / A2W6-A2-003 | FIXED-OK | закрыт | Login Widget / WebApp `VerifyTelegramAuth_*` + legacy `service_test` telegram crypto удалены |
| W6-A2-004 | FIXED-OK | закрыт | `PendingUsersCRUD` — Approve→`StatusActive`, Reject→`IsNotFound` (`service_organization_test.go:71-80`) |
| W6-A2-005 | FIXED-OK | закрыт | `PurgeExpiredPending` — expired −8d → `n==1`, fresh остаётся (`:105-112`) |

**REGRESS:** 0 · assert-free: **не найдено**

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/testframework/auth/repository_organization_test.go` | 121 | 0 | чисто (Ent pending/approve/purge) |
| `backend/testframework/auth/repository_test.go` | 73 | 0 | чисто |
| `backend/testframework/auth/resend_telegram_test.go` | 72 | 1 | Resend only; body assert слаб |
| `backend/testframework/auth/service_login_test.go` | 135 | 1 | pending denial OK; deactivated login gap |
| `backend/testframework/auth/service_organization_test.go` | 120 | 0 | FIXED-OK 004/005 |
| `backend/testframework/auth/service_profile_test.go` | 194 | 0 | чисто (TelegramID = bind field) |
| `backend/testframework/auth/service_registration_race_test.go` | 53 | 1 | upsert race OK; founder race отсутствует |
| `backend/testframework/auth/service_registration_test.go` | 99 | 0 | чисто (fail-closed invite_only) |
| `backend/testframework/auth/service_session_test.go` | 448 | 1 | session suite сильный; Invalidate Redis assert gap |
| `backend/testframework/auth/service_test.go` | 90 | 0 | FIXED-OK (JWT/code only, без Telegram) |
| `backend/testframework/auth/telegram_bot_coverage_test.go` | — | 0 | **deleted** TASK-102 |
| `backend/testframework/auth/telegram_bot_polling_test.go` | — | 0 | **deleted** TASK-102 |
| `backend/testframework/auth/telegram_widget_test.go` | — | 0 | **deleted** TASK-102 |
| `backend/testframework/auth/telegram_login_removed_task102_test.go` | 96 | 0 | FIXED-OK denial (вне списка, замена) |

Файлов в списке: 13. Прочитано: 13 (+ denial replacement). Пропущено: 0.  
Deleted 3/13 — intentional TASK-102; не пропуск аудита.

### Cross-slot (фокус rune — A1)

| Файл | Строк | Статус |
|------|------:|--------|
| `backend/testframework/auth/handler_magiclink_runes_test.go` | 81 | **FIXED-OK** — 3 кейса живы; prod `handler.go:126` = `utf8.RuneCountInString` |
| `backend/internal/auth/handler.go:126-127` | — | нет регрессии к `len()` |

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|----------|-------------------|------------|
| A3W6-2-001 | NEW | P2 | `service_login_test.go:44-66` + `service_login.go:61-63` | Denial только для `StatusPending`; нет `LoginWithMagicLink` с `StatusDeactivated` | Мутация `if u.Status == StatusPending` (вместо `!= Active`) → pending-тест зелёный, deactivated получает JWT. Middleware ловит уже после выдачи токенов | Добавить `TestLoginWithMagicLink_DeactivatedUserDenied`: valid code + deactivated → `ErrAccountNotActive`, sessions==0 |
| A3W6-2-002 | NEW | P2 | `service_registration_race_test.go:16-37` vs `:39-53` | «Race» suite гоняет concurrent только `SetRegistrationMode`; `CreateFirstOrMemberUser` — строго sequential | Мутация удаления `firstUserMu` / tx (`service_registration.go:72-81`) не краснеет: оба concurrent call могут увидеть `count==0` → dual founder (SEC-06) | N concurrent `CreateFirstOrMemberUser` → ровно 1 Founder, остальные Member; assert count(Founder)==1 |
| A3W6-2-003 | NEW | P3 | `service_session_test.go:374-406` | `InvalidateUserSessions_ClearsMultiple` assert только DB `Count==0`, Redis keys не проверяет | Мутация skip Redis pipeline Del (`service_session.go:147-155`) → тест зелёный. Частично закрыто через `CreateSession_SingleSessionPolicy` (`:300-301`) | После invalidate: `assert.False(mr.Exists(ExportRefreshRedisKey(plain)))` для обоих seeded tokens |
| A3W6-2-004 | NEW | P3 | `resend_telegram_test.go:24-44` | Success-мок проверяет Method+Bearer, не тело (`to`/`code`/HTML) | Мутация пустого/чужого payload при 200 upstream → PASS | В handler: decode JSON, `assert.Contains` email + code |

## Фокус-чеклист

| Чек | Статус |
|-----|--------|
| telegram-as-login absence | **OK** — prod `telegram.go`/`telegram_bot.go` отсутствуют; denial markers; `events.go` Method = magic_link only; `GetUserByTelegramID` нет в auth Repository |
| Позитивные Telegram login tests | **OK** — `TestAuth_NoPositiveTelegramLoginTests_Task102` |
| Rune length (multibyte / emoji / byte-trap) | **OK** (A1 slot) — нет REGRESS на `RuneCountInString` |
| Mutation-proof pending/purge org | **OK** — A2-004/005 FIXED-OK |
| Mutation gaps remaining | **NEW** A3W6-2-001…004 |
| assert-free | **Не найдено** |

## Чисто

- Magic Link login path only; TelegramID в profile = bind-поле ответа, не login surface.
- Session: hash≠plaintext Redis key, rotate, logout, ownership revoke, expired/hash/deactivated refresh, single-session, NATS created/nil-safe.
- Registration: default invite_only; fail-closed settings fail → `ErrRegistrationInviteOnly`; sequential second user = Member.
- Org/repo Ent: invite token, approve/reject NotFound, purge −8d, CountOrganizations.
- JWT round-trip / expired / wrong secret; `generateCode` 6 ASCII digits.
- Avatar scheme whitelist + onboarding path (TASK-116) без регрессии.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 2 |
| P3 | 2 |
| **Total open** | **4** |

| Tag | Count |
|-----|------:|
| NEW | 4 |
| REGRESS | 0 |
| RESIDUAL | 0 |
| FIXED-OK (A2 closes) | 5 |

A2 baseline open → 0 · FIXED-OK 5 · REGRESS 0 · NEW 4 (P2=2 · P3=2).
