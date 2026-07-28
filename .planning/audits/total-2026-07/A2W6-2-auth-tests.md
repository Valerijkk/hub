# A2 W6 агент A2 — auth remaining tests

> TASK-008 · Stage A2 · read-only · 2026-07-28  
> Периметр: `W6-ASSIGNMENTS.md` §A2 (13 файлов)

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/testframework/auth/repository_organization_test.go | 121 | 0 | чисто |
| backend/testframework/auth/repository_test.go | 73 | 0 | чисто |
| backend/testframework/auth/resend_telegram_test.go | 184 | 1 | RESIDUAL |
| backend/testframework/auth/service_login_test.go | 135 | 0 | чисто |
| backend/testframework/auth/service_organization_test.go | 120 | 0 | FIXED-OK |
| backend/testframework/auth/service_profile_test.go | 144 | 0 | чисто |
| backend/testframework/auth/service_registration_race_test.go | 53 | 0 | чисто |
| backend/testframework/auth/service_registration_test.go | 99 | 0 | чисто |
| backend/testframework/auth/service_session_test.go | 448 | 0 | чисто |
| backend/testframework/auth/service_test.go | 214 | 1 | RESIDUAL |
| backend/testframework/auth/telegram_bot_coverage_test.go | 26 | 1 | RESIDUAL |
| backend/testframework/auth/telegram_bot_polling_test.go | 136 | 1 | RESIDUAL |
| backend/testframework/auth/telegram_widget_test.go | 97 | 1 | RESIDUAL |

Файлов: **13/13/0**.

## Перепроверка Stage A (W6-A2)

| Старый ID | Tag | Статус | Комментарий |
|-----------|-----|--------|-------------|
| W6-A2-001 | RESIDUAL | жив | `HandleLogin_StoresCode` всё ещё требует `tglogin:` (P1-BE-U02) |
| W6-A2-002 | RESIDUAL | жив | VerifyCode / GenerateTGCode / RecordVerifyFailure suite |
| W6-A2-003 | RESIDUAL | жив | LoginWidget + WebApp `VerifyTelegramAuth_*` |
| W6-A2-004 | FIXED-OK | закрыт | `PendingUsersCRUD` — Approve→Active, Reject→NotFound |
| W6-A2-005 | FIXED-OK | закрыт | `PurgeExpiredPending` — expired −8d → n==1 |

REGRESS: **0**. assert-free: **не найдено**.

## Находки (открытые)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|----------|---------------|------------|
| A2W6-A2-001 | RESIDUAL | P1 | `telegram_bot_polling_test.go:57-88` | Позитивный `/login`→`tglogin:` | ADR-AUTH-01; нормализует запрещённый login | Удалить тест; denial-only polling |
| A2W6-A2-002 | RESIDUAL | P2 | `resend_telegram_test.go`; `telegram_bot_coverage_test.go`; polling VerifyCode | Suite держит legacy login-бота | Единственные потребители мёртвого auth | Удалить login-suite |
| A2W6-A2-003 | RESIDUAL | P2 | `telegram_widget_test.go`; `service_test.go` VerifyTelegramAuth | Login Widget / WebApp crypto без wiring | Противоречит AUTH-01 | Удалить или перенести в notification-bind |

## Counts

| | n |
|--|--:|
| NEW | 0 |
| REGRESS | 0 |
| RESIDUAL | 3 |
| FIXED-OK | 2 |
| P0 / P1 / P2 / P3 (open) | 0 / 1 / 2 / 0 |
| **Σ open** | **3** |
