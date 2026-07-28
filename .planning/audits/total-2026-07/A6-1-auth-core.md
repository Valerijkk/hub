# W6 агент A1 — auth core

> TASK-008 · Стадия A · read-only · 2026-07-28  
> Фокус: Magic Link only · JWT 15m · refresh Redis · rate 3/10min · ConstantTimeCompare · Telegram NOT login · rune-safe code length

## Прочитанные файлы
| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/internal/auth/dto.go | 101 | 1 | замечания |
| backend/internal/auth/events.go | 16 | 1 | замечания |
| backend/internal/auth/export_hooks.go | 111 | 0 | чисто |
| backend/internal/auth/handler.go | 551 | 1 | замечания |
| backend/internal/auth/magiclink.go | 175 | 0 | чисто |
| backend/internal/auth/middleware.go | 150 | 0 | чисто |
| backend/internal/auth/repository.go | 462 | 2 | замечания |
| backend/internal/auth/resend.go | 87 | 0 | чисто |
| backend/internal/auth/service.go | 129 | 0 | чисто |
| backend/internal/auth/service_login.go | 109 | 0 | чисто |
| backend/internal/auth/service_organization.go | 113 | 0 | чисто |
| backend/internal/auth/service_profile.go | 65 | 1 | замечания |
| backend/internal/auth/service_registration.go | 96 | 1 | замечания |
| backend/internal/auth/service_session.go | 294 | 0 | чисто |
| backend/internal/auth/telegram.go | 122 | 1 | замечания |
| backend/internal/auth/telegram_bot.go | 187 | 1 | замечания |
| backend/testframework/auth/handler_admin_extra_test.go | 45 | 0 | чисто |
| backend/testframework/auth/handler_admin_test.go | 162 | 0 | чисто |
| backend/testframework/auth/handler_extra_test.go | 134 | 1 | замечания |
| backend/testframework/auth/handler_integration_test.go | 169 | 0 | чисто |
| backend/testframework/auth/handler_invite_coverage_test.go | 62 | 0 | чисто |
| backend/testframework/auth/handler_invite_test.go | 120 | 0 | чисто |
| backend/testframework/auth/handler_magiclink_runes_test.go | 81 | 0 | чисто |
| backend/testframework/auth/handler_organization_test.go | 125 | 0 | чисто |
| backend/testframework/auth/handler_profile_test.go | 160 | 0 | чисто |
| backend/testframework/auth/handler_test.go | 147 | 0 | чисто |
| backend/testframework/auth/handler_wave115_test.go | 79 | 0 | чисто |
| backend/testframework/auth/magiclink_test.go | 365 | 0 | чисто |
| backend/testframework/auth/middleware_failclosed_test.go | 92 | 0 | чисто |
| backend/testframework/auth/middleware_test.go | 196 | 0 | чисто |

Файлов в списке: 30. Прочитано: 30. Пропущено: 0.

## Перепроверка прошлого аудита (grok-2026-07 / A1 auth)

| Старый ID | Статус в текущем коде | Комментарий |
|-----------|----------------------|-------------|
| A1-001 (pending→JWT) | **исправлено** — не дублировать | `service_login.go:61-63` — `u.Status != active` → `ErrAccountNotActive`; middleware считает non-active deactivated |
| A1-007 (Resend → 200) | **исправлено** — не дублировать | `magiclink.go:105-109` возвращает error; тест `magiclink_test.go` A1_007 |
| A1-010 (Del без err) | **исправлено** — не дублировать | `magiclink.go:146-149` fail-closed; тест A1_010 |
| AUTH-01 Telegram login routes | **закрыто на HTTP** | `handler.go` — только magic-link/verify/refresh; wiring: «Telegram login убран» |
| AUTH-01 dead login code | **живо** | `telegram.go` / `telegram_bot.go` — см. W6-A1-001/002 |

## Фокус TZ (сверка)

| Инвариант | Статус | Доказательство |
|-----------|--------|----------------|
| Magic Link only | **OK на routes** | `handler.go:35-37` — нет `/auth/telegram`; login только Magic Link |
| JWT 15m | **OK** | `SessionService.accessTTL` из `cfg.JWT.AccessTokenTTL` (`config.go` = 15m); тесты `15*time.Minute` |
| Refresh в Redis | **OK** | `service_session.go:98-105` — `refresh:<sha256>`, TTL `refreshTTL` (30d) |
| Rate 3/10min | **OK** | `magiclink.go:20-21` `RateMax=3`/`RateTTL=10m`; wiring `/auth` `AuthRateLimit` 3/10m |
| ConstantTimeCompare | **OK** | `magiclink.go:141` код; `service_session.go:216` refresh hash; Telegram hash — `hmac.Equal` |
| Telegram NOT login | **частично** | HTTP login нет; legacy `VerifyTelegramAuth`/`TelegramBotAuth`/`tglogin:` остаются в пакете |
| Rune-safe code length | **OK** | `handler.go:122` `utf8.RuneCountInString`; regression `handler_magiclink_runes_test.go` |

## Находки
| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|----------|-------------------|------------|
| W6-A1-001 | P2 | `backend/internal/auth/telegram_bot.go:28-162` | Legacy login-бот: `handleLogin`, `VerifyCode`, префикс `tglogin:`, текст «код входа»; `StartPolling` no-op, но API живой | AUTH-01 / TZ: Telegram — не login. Мёртвый login-путь + export hooks (`ExportTelegramHandleLogin`) упрощают случайный rewire и раздувают attack surface | Удалить `TelegramBotAuth`/login-хелперы из prod (или `//go:build hubtest`); bind оставить в `notification` |
| W6-A1-002 | P2 | `backend/internal/auth/telegram.go:26-120` | `VerifyTelegramAuth` (WebApp + **Login Widget**) без prod-callers (только testframework) | Login Widget = login-контракт Telegram; противоречит «Telegram NOT login»; мёртвый код рядом с auth | Удалить или перенести в notification bind, если понадобится подпись initData; Login Widget не оставлять |
| W6-A1-003 | P2 | `backend/internal/auth/service_registration.go:46-51` | `CheckRegistrationAllowed`: `mode, _ := GetRegistrationMode` — ошибка БД глотается; `mode==""` ≠ `invite_only` → **fail-open** (разрешает регистрацию) | При сбое settings API вернёт «open», хотя дефолт TZ — invite_only. Сейчас вызывается только из тестов (A2), но публичный опасный контракт | Проверять `err`; при ошибке — fail-closed `invite_only` / return err |
| W6-A1-004 | P2 | `backend/internal/auth/dto.go:8-23` (+ `handler.go` Bind без validator) | Теги `validate:"required,email"` / `len=6` на DTO не исполняются | Пустой `refresh_token`, битый email после Bind могут уйти глубже (email частично ловит `mail.ParseAddress`; code — rune-check вручную; refresh — только пустой Bind) | Fiber StructValidator или `validate.Struct` после Bind |
| W6-A1-005 | P2 | `backend/internal/auth/repository.go:111-112` (+ `dto.go:29`, `service_profile.go:32`) | `AvatarURL` пишется без проверки схемы URL | `javascript:`/`data:` в профиле → stored XSS при рендере аватара на фронте | Whitelist `https:` (и относительные `/…`); иначе 400 |
| W6-A1-006 | P2 | `backend/testframework/auth/handler_extra_test.go:21-40` | `TestHandler_ListSessions_MarksCurrent` не читает JSON и **не assert'ит** `is_current` | Имя обещает проверку current-session; тест зелёный при сломанном `IsCurrent` (`handler.go:297` + сравнение UUID) | Decode `[]SessionResponse`; ровно один `IsCurrent==true` с ID текущей сессии |
| W6-A1-007 | P3 | `backend/internal/auth/events.go:15` | Комментарий `Method // magic_link или telegram`; `UserRegisteredEvent` нигде не публикуется | Drift AUTH-01 + мёртвый event-контракт | Убрать `telegram` из комментария; либо публиковать `method=magic_link`, либо удалить тип |
| W6-A1-008 | P3 | `backend/internal/auth/repository.go:38,95-97` | `GetUserByTelegramID` в Repository / Ent — без login-callers в auth | Остаток Telegram-login data path; путает с AUTH-01 | Оставить только если notification/bind нужен; иначе удалить из auth Repository |
| W6-A1-009 | P3 | `backend/internal/auth/repository.go:263-281` | `CreateOrganizationWithFounder(email, inviteToken)` **игнорирует** `inviteToken`, всегда `generateInviteToken()` | Мёртвый параметр; вызывающий код думает, что передаёт токен | Убрать параметр или использовать переданный токен явно |
| W6-A1-010 | P3 | `backend/internal/auth/service_profile.go:35-39` | `CompleteOnboarding`: UpdateProfile, затем MarkOnboardingComplete — не одна транзакция | Сбой на втором шаге → профиль обновлён, `onboarding_completed=false` | Одна Ent-tx / repo-метод «update+mark» |

## Counts

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 6 |
| P3 | 4 |
| **Σ** | **10** |

Файлов прочитано: **30**/30. С находками: **10**. Чисто: **20**.

## Топ-3

1. **W6-A1-001** — legacy Telegram login-бот (`tglogin:` / `handleLogin` / `VerifyCode`) в prod-пакете auth при отключённом polling (AUTH-01 residual).
2. **W6-A1-002** — `VerifyTelegramAuth` + Login Widget без prod-callers: мёртвый login-контракт рядом с Magic Link only.
3. **W6-A1-003** — `CheckRegistrationAllowed` fail-open при ошибке чтения `registration_mode` (глотает err → mode "").

## Чисто

Файлы без находок — что проверялось:

- **export_hooks.go** — hubtest-only экспорты; нет prod-логики; Telegram exports отражают legacy (учтено в W6-A1-001).
- **magiclink.go** — `ConstantTimeCompare`; rate SET NX + Incr ≤3/10m; verify lockout 5; Del fail-closed; OTP не в slog; Resend error не маскируется.
- **middleware.go** — JWT parse; fail-closed на DB status (503); pending/deactivated → 403; Redis cache TTL 30s; API-key short-circuit (A9-012) без ослабления deactivated для JWT-пути.
- **resend.go** — timeout client; `LimitReader` 10MB на error body; код цифровой в HTML (не user-controlled injection).
- **service.go** — оркестратор; Magic Link login only; нет Telegram login facade.
- **service_login.go** — Magic Link only; pending/non-active → `ErrAccountNotActive`; invite gate `CanSendMagicLink`.
- **service_organization.go** — invite validate; pending purge 7d; `IsSystemAdmin` founder\|\|admin.
- **service_session.go** — refresh Redis + DB hash `ConstantTimeCompare`; rotation ME-02; single-session invalidate; deactivated block на refresh.
- **handler_admin_*.go / handler_test.go / handler_integration / invite* / organization / profile / wave115** — assert'ы на HTTP-коды; invite-only 403; rate 429; founder admin; rune-тесты отдельно.
- **handler_magiclink_runes_test.go** — multibyte 6 рун → не 400; emoji/3-руны-6-байт → 400.
- **magiclink_test.go** — rate 3; lockout; Del fail; Resend err; OTP not logged.
- **middleware_*_test.go** — valid/missing/deactivated/cache/fail-closed 503 / bad uid claim.

## Verification (агент)

- Режим: **read-only** (код не менялся).
- Артефакт: `.planning/audits/total-2026-07/A6-1-auth-core.md`
- ID: `W6-A1-001` … `W6-A1-010`
