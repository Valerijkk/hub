# A2 W6 агент A1 — auth core

> TASK-008 · Stage A2 · read-only · 2026-07-28  
> Периметр: `W6-ASSIGNMENTS.md` §A1 (30 файлов) · теги: NEW | REGRESS | RESIDUAL | FIXED-OK

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/internal/auth/dto.go | 101 | 0 | FIXED-OK (validate) |
| backend/internal/auth/events.go | 16 | 1 | RESIDUAL |
| backend/internal/auth/export_hooks.go | 111 | 0 | чисто (legacy export учтён в 001) |
| backend/internal/auth/handler.go | 551 | 0 | чисто (Magic Link only routes) |
| backend/internal/auth/magiclink.go | 175 | 0 | чисто |
| backend/internal/auth/middleware.go | 150 | 0 | чисто |
| backend/internal/auth/repository.go | 463 | 2 | RESIDUAL |
| backend/internal/auth/resend.go | 87 | 0 | чисто |
| backend/internal/auth/service.go | 129 | 0 | чисто |
| backend/internal/auth/service_login.go | 109 | 0 | чисто |
| backend/internal/auth/service_organization.go | 113 | 0 | чисто |
| backend/internal/auth/service_profile.go | 68 | 0 | FIXED-OK |
| backend/internal/auth/service_registration.go | 100 | 0 | FIXED-OK |
| backend/internal/auth/service_session.go | 294 | 0 | чисто |
| backend/internal/auth/telegram.go | 122 | 1 | RESIDUAL |
| backend/internal/auth/telegram_bot.go | 187 | 1 | RESIDUAL |
| backend/testframework/auth/handler_admin_extra_test.go | 45 | 0 | чисто |
| backend/testframework/auth/handler_admin_test.go | 162 | 0 | чисто |
| backend/testframework/auth/handler_extra_test.go | 134 | 1 | RESIDUAL |
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

Файлов в списке: **30**. Прочитано: **30**. Пропущено: **0**.

## Перепроверка Stage A (W6-A1)

| Старый ID | Tag | Статус в коде | Комментарий |
|-----------|-----|---------------|-------------|
| W6-A1-001 | RESIDUAL | **FIXED** TASK-102 | `telegram_bot.go` удалён — login API gone |
| W6-A1-002 | RESIDUAL | **FIXED** TASK-102 | `telegram.go` / `VerifyTelegramAuth` удалены |
| W6-A1-003 | FIXED-OK | закрыт | `service_registration.go:47-50` — err → `ErrRegistrationInviteOnly` (TASK-040) |
| W6-A1-004 | FIXED-OK | закрыт | `wiring_fiber.go` `StructValidator`; DTO tags исполняются |
| W6-A1-005 | RESIDUAL | жив | `repository.go:111-112` — AvatarURL без whitelist схемы |
| W6-A1-006 | RESIDUAL | жив | `handler_extra_test.go:21-40` — нет decode/`is_current` |
| W6-A1-007 | RESIDUAL | **FIXED** TASK-102 | `events.go` Method — только `magic_link` |
| W6-A1-008 | RESIDUAL | **FIXED** TASK-102 | `GetUserByTelegramID` удалён из auth Repository |
| W6-A1-009 | FIXED-OK | закрыт | `CreateOrganizationWithFounder(ctx, email)` — мёртвый `inviteToken` убран |
| W6-A1-010 | FIXED-OK | закрыт | `CompleteOnboarding` в `WithEntTx` |

REGRESS: **0**.

## Находки (открытые)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|----------|---------------|------------|
| A2W6-A1-001 | **FIXED** | P2 | `telegram_bot.go` (deleted) | Legacy Telegram login-бот удалён | ADR-AUTH-01 | TASK-102: delete + denial `TestAuth_NoTelegramLoginSurface_Task102` |
| A2W6-A1-002 | **FIXED** | P2 | `telegram.go` (deleted) | `VerifyTelegramAuth` удалён | Magic Link only | TASK-102 |
| A2W6-A1-003 | RESIDUAL | P2 | `repository.go:111-112` | `AvatarURL` без проверки схемы | `javascript:`/`data:` → stored XSS risk; TASK-042 out-of-scope | Whitelist `https:` / relative `/…` |
| A2W6-A1-004 | RESIDUAL | P2 | `handler_extra_test.go:21-40` | `ListSessions_MarksCurrent` не assert'ит `is_current` | Сломанный IsCurrent → зелёный тест | Decode body; ровно один current = session JWT |
| A2W6-A1-005 | **FIXED** | P3 | `events.go` | Method comment = magic_link only | Drift AUTH-01 closed | TASK-102 |
| A2W6-A1-006 | **FIXED** | P3 | `repository.go` | `GetUserByTelegramID` удалён из auth | Bind остаётся в notification | TASK-102 |

## Фокус TZ

| Инвариант | Статус |
|-----------|--------|
| Magic Link only (HTTP) | OK — routes без Telegram login |
| JWT 15m / refresh Redis / rate 3/10m | OK |
| ConstantTimeCompare | OK |
| Rune-safe code length | OK |
| Telegram NOT login | **OK** — legacy login API/тесты удалены (TASK-102); bind в notification |

## Counts

| | n |
|--|--:|
| NEW | 0 |
| REGRESS | 0 |
| RESIDUAL | 6 |
| FIXED-OK | 4 |
| P0 / P1 / P2 / P3 (open) | 0 / 0 / 4 / 2 |
| **Σ open** | **6** |

Файлов: **30/30**.
