# A2 W6 агент A2 — auth remaining tests

> TASK-008 · Stage A2 · read-only · 2026-07-28  
> Периметр: `W6-ASSIGNMENTS.md` §A2 (13 файлов)  
> **Update 2026-07-28 TASK-102:** Telegram login suite удалён (ADR-AUTH-01).

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/testframework/auth/repository_organization_test.go | 121 | 0 | чисто |
| backend/testframework/auth/repository_test.go | 73 | 0 | чисто |
| backend/testframework/auth/resend_telegram_test.go | — | 0 | **FIXED** TASK-102 (только Resend) |
| backend/testframework/auth/service_login_test.go | 135 | 0 | чисто |
| backend/testframework/auth/service_organization_test.go | 120 | 0 | FIXED-OK |
| backend/testframework/auth/service_profile_test.go | 144 | 0 | чисто |
| backend/testframework/auth/service_registration_race_test.go | 53 | 0 | чисто |
| backend/testframework/auth/service_registration_test.go | 99 | 0 | чисто |
| backend/testframework/auth/service_session_test.go | 448 | 0 | чисто |
| backend/testframework/auth/service_test.go | — | 0 | **FIXED** TASK-102 (JWT/code only) |
| backend/testframework/auth/telegram_bot_coverage_test.go | — | — | **deleted** TASK-102 |
| backend/testframework/auth/telegram_bot_polling_test.go | — | — | **deleted** TASK-102 |
| backend/testframework/auth/telegram_widget_test.go | — | — | **deleted** TASK-102 |
| backend/testframework/auth/telegram_login_removed_task102_test.go | — | 0 | denial-регрессия TASK-102 |

Файлов: **13/13/0** (A2 snapshot) + denial file TASK-102.

## Перепроверка Stage A (W6-A2)

| Старый ID | Tag | Статус | Комментарий |
|-----------|-----|--------|-------------|
| W6-A2-001 | RESIDUAL | **FIXED** TASK-102 | `HandleLogin_StoresCode` удалён; denial `TestAuth_NoTelegramLoginSurface_Task102` |
| W6-A2-002 | RESIDUAL | **FIXED** TASK-102 | login VerifyCode suite удалён |
| W6-A2-003 | RESIDUAL | **FIXED** TASK-102 | LoginWidget / WebApp VerifyTelegramAuth_* удалены |
| W6-A2-004 | FIXED-OK | закрыт | `PendingUsersCRUD` — Approve→Active, Reject→NotFound |
| W6-A2-005 | FIXED-OK | закрыт | `PurgeExpiredPending` — expired −8d → n==1 |

REGRESS: **0**. assert-free: **не найдено**.

## Находки (открытые)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|----------|---------------|------------|
| A2W6-A2-001 | **FIXED** | P1 | deleted | Позитивный `/login`→`tglogin:` | ADR-AUTH-01 | TASK-102 denial tests |
| A2W6-A2-002 | **FIXED** | P2 | deleted | Suite держал legacy login-бота | — | TASK-102 |
| A2W6-A2-003 | **FIXED** | P2 | deleted | Login Widget / WebApp crypto | — | TASK-102 |

## Counts

| | n |
|--|--:|
| NEW | 0 |
| REGRESS | 0 |
| RESIDUAL | 0 |
| FIXED-OK / FIXED | 5 |
| P0 / P1 / P2 / P3 (open) | 0 / 0 / 0 / 0 |
| **Σ open** | **0** |
