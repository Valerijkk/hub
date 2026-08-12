# A2W7-1 — notification+mail (Stage A2)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Baseline: `A7-1-notification-mail.md` · Assignments: `W7-ASSIGNMENTS.md` §A1  
> Tags: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/notification/consumer.go` | 743 | 1 | замечания |
| `backend/internal/notification/dto.go` | 71 | 0 | чисто |
| `backend/internal/notification/events.go` | 99 | 0 | чисто |
| `backend/internal/notification/export_hooks.go` | 220 | 0 | чисто |
| `backend/internal/notification/handler.go` | 397 | 1 | замечания |
| `backend/internal/notification/handler_telegram_bind.go` | 128 | 0 | чисто |
| `backend/internal/notification/push.go` | 295 | 1 | замечания |
| `backend/internal/notification/repository.go` | 343 | 0 | чисто |
| `backend/internal/notification/service.go` | 397 | 0 | чисто |
| `backend/internal/notification/telegram.go` | 139 | 0 | чисто |
| `backend/internal/notification/telegram_bind.go` | 341 | 1 | замечания |
| `backend/testframework/notification/consumer_ack_a2010_test.go` | 82 | 0 | чисто |
| `backend/testframework/notification/consumer_extra_test.go` | 173 | 0 | чисто |
| `backend/testframework/notification/consumer_test.go` | 290 | 0 | чисто |
| `backend/testframework/notification/consumer_wave102_test.go` | 153 | 0 | чисто |
| `backend/testframework/notification/consumer_wave118_test.go` | 160 | 0 | чисто |
| `backend/testframework/notification/consumer_wave119_test.go` | 238 | 0 | чисто |
| `backend/testframework/notification/consumer_wave121_test.go` | 197 | 0 | чисто |
| `backend/testframework/notification/consumer_wave94_test.go` | 46 | 0 | чисто |
| `backend/testframework/notification/consumer_wave99_test.go` | 60 | 0 | чисто |
| `backend/testframework/notification/coverage_wave79_test.go` | 261 | 0 | чисто |
| `backend/testframework/notification/fixtures_test.go` | 85 | 0 | чисто |
| `backend/testframework/notification/handler_list_filter_test.go` | 62 | 0 | чисто |
| `backend/testframework/notification/handler_telegram_bind_test.go` | 149 | 0 | чисто |
| `backend/testframework/notification/handler_telegram_bind_wave102_test.go` | 122 | 0 | чисто |
| `backend/testframework/notification/handler_test.go` | 406 | 0 | чисто |
| `backend/testframework/notification/handler_wave115_test.go` | 35 | 0 | чисто |
| `backend/testframework/notification/handler_wave116_test.go` | 23 | 0 | чисто |
| `backend/testframework/notification/handler_wave117_test.go` | 57 | 0 | чисто |
| `backend/testframework/notification/handler_wave118_test.go` | 68 | 0 | чисто |
| `backend/testframework/notification/handler_wave119_test.go` | 42 | 0 | чисто |
| `backend/testframework/notification/handler_wave99_test.go` | 45 | 0 | чисто |
| `backend/testframework/notification/push_lifecycle_test.go` | 60 | 0 | чисто |
| `backend/testframework/notification/push_scope_a2005_test.go` | 77 | 0 | чисто |
| `backend/testframework/notification/push_wave119_test.go` | 57 | 0 | чисто |
| `backend/testframework/notification/push_wave95_test.go` | 65 | 1 | замечания |
| `backend/testframework/notification/repository_extra_test.go` | 146 | 0 | чисто |
| `backend/testframework/notification/resolve_prefs_a13005_test.go` | 68 | 0 | чисто |
| `backend/testframework/notification/service_extended_test.go` | 370 | 0 | чисто |
| `backend/testframework/notification/service_test.go` | 134 | 0 | чисто |
| `backend/testframework/notification/service_wave111_test.go` | 178 | 0 | чисто |
| `backend/testframework/notification/service_wave117_test.go` | 103 | 0 | чисто |
| `backend/testframework/notification/service_wave118_test.go` | 57 | 0 | чисто |
| `backend/testframework/notification/service_wave119_test.go` | 62 | 0 | чисто |
| `backend/testframework/notification/stubs_test.go` | 34 | 0 | чисто |
| `backend/testframework/notification/telegram_bind_test.go` | 127 | 0 | чисто |
| `backend/testframework/notification/telegram_bind_wave102_test.go` | 74 | 0 | чисто |
| `backend/testframework/notification/telegram_bind_wave95_test.go` | 128 | 0 | чисто |
| `backend/testframework/notification/telegram_lifecycle_test.go` | 99 | 1 | замечания |
| `backend/testframework/notification/telegram_wave97_test.go` | 34 | 1 | замечания |
| `backend/internal/mail/admin_broadcast.go` | 70 | 1 | замечания |
| `backend/internal/mail/admin_handler.go` | 113 | 1 | замечания |
| `backend/internal/mail/admin_service.go` | 115 | 0 | чисто |
| `backend/internal/mail/config.go` | 43 | 0 | чисто |
| `backend/internal/mail/dto.go` | 41 | 0 | чисто |
| `backend/internal/mail/export_hooks.go` | 13 | 0 | чисто |
| `backend/internal/mail/hub_handler.go` | 48 | 0 | чисто |
| `backend/internal/mail/internal_auth.go` | 31 | 0 | чисто |
| `backend/internal/mail/service.go` | 111 | 1 | замечания |
| `backend/internal/mail/settings_store.go` | 55 | 0 | чисто |
| `backend/testframework/mail/admin_handler_test.go` | 145 | 0 | чисто |
| `backend/testframework/mail/admin_service_test.go` | 114 | 0 | чисто |
| `backend/testframework/mail/config_test.go` | 44 | 0 | чисто |
| `backend/testframework/mail/hub_handler_test.go` | 106 | 0 | чисто |
| `backend/testframework/mail/internal_auth_test.go` | 63 | 0 | чисто |
| `backend/testframework/mail/send_alert_a7005_test.go` | 25 | 0 | чисто |

Файлов в списке: **66**. Прочитано: **66**. Пропущено: **0**.

## Находки

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект / статус |
|----|-----|-----|-------------|---------|------------------------|
| W7-A1-001 | FIXED-OK | P1 | `notification/consumer.go:23-28`, handlers | Soft `return nil` на repo err → Ack | Закрыто TASK-024 / W18-C2: `repoLookupErr` → Nak на transient; not-found → Ack |
| W7-A1-002 | FIXED | P2 | `notification/handler.go` | `ValidateOutboundURL` fail → **400** | TASK-117: ErrInvalidPushEndpoint |
| W7-A1-003 | FIXED | P2 | `mail/service.go` | `DailyLimit` / `alerts_enabled` enforcement в `SendAlert` | TASK-117: store+cfg policy |
| W7-A1-004 | FIXED | P2 | `mail/admin_broadcast.go` | `Failed` count + slog | TASK-117 |
| W7-A1-005 | FIXED | P2 | `mail/admin_handler.go` | без `*ent.Client`; ActiveUserEmailSource | TASK-117 |
| W7-A1-006 | FIXED | P2 | `telegram_wave97_test.go` | meaningful asserts | TASK-117 |
| W7-A1-007 | FIXED | P2 | `telegram_lifecycle_test.go` | queue/empty asserts | TASK-117 |
| W7-A1-008 | FIXED | P2 | `push_wave95_test.go` | https-sub остаётся в БД | TASK-117 |
| W7-A1-009 | FIXED-OK | P3 | `service_extended_test.go:314-342` | Cross-midnight quiet hours без assert | Закрыто TASK-058: таблица expected bool |
| W7-A1-010 | FIXED-OK | P3 | `telegram_bind.go:127-130`, `:245+` | `rdb.Del` без проверки err | Закрыто TASK-057: `slog.Warn` на Del fail |

## Чисто (инварианты держат)

| Зона | Проверено |
|------|-----------|
| NF-07 WebPush | `ValidateOutboundURL` на save+send; `safehttp.NewSafeHTTPClient` |
| Push ownership A2-005 | Update/Delete с UserID Where |
| Telegram ≠ login | Только bind/confirm/unlink/test/status |
| Mail CTC | `internal_auth.go` ConstantTimeCompare; empty → 503 |
| Consumer Ack/Nak | Explicit err → Nak после TASK-024 |

## Counts

| Метрика | n |
|---------|--:|
| Файлов | 66 |
| Open (RESIDUAL) | **0** |
| FIXED-OK | 10 |
| NEW | 0 |
| REGRESS | 0 |
| Open by sev | P0:0 · P1:0 · P2:0 · P3:0 |
