# A3 W7 агент 1 — notification+mail (FINAL re-audit)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · ветка `cursor-grok-4.5`  
> Mirror: [A2W7-1-notification-mail.md](A2W7-1-notification-mail.md) · Baseline Stage A: [A7-1-notification-mail.md](A7-1-notification-mail.md)  
> Периметр: **66** файлов (11+10 prod + 45 tests) · Прочитано: **66/66** · Пропущено: **0**  
> Plan: [A3-PLAN.md](A3-PLAN.md) · Assignment: [W7-ASSIGNMENTS.md](W7-ASSIGNMENTS.md) § A1  
> Фокус: WebPush `ValidateOutboundURL` NF-07 · Telegram ≠ login · mail secrets · rate limits

## Сверка с A2 (A2W7-1 / W7-A1-*)

| A2 / Stage A ID | Tag | Статус в HEAD | Комментарий |
|-----------------|-----|---------------|-------------|
| W7-A1-001 | FIXED-OK | закрыт | `repoLookupErr` → Nak на transient; not-found → Ack — `consumer.go:23–28` + все handler paths; `consumer_ack_a2010_test.go` (+ TASK-024 вне слота) |
| W7-A1-002 | FIXED-OK | закрыт | `ErrInvalidPushEndpoint` → HTTP 400 — `handler.go:325–328`; `SaveSubscription` wrap — `push.go:227–229`; тест `handler_test.go:311–320` (`javascript:` → 400) |
| W7-A1-003 | FIXED-OK | закрыт | `effectivePolicy` + `reserveDailySlot` в `SendAlert` — `service.go:83–93`, `162–194`; store override — `admin_service_test.go` W7_A1_003 |
| W7-A1-004 | FIXED-OK | закрыт | Broadcast `Failed++` + `slog.Warn` — `admin_broadcast.go:69–73`; `TestAdminHandler_BroadcastFailedCount` |
| W7-A1-005 | FIXED-OK | закрыт | `AdminHandler` без `*ent.Client`; `ActiveUserEmailSource` — `admin_handler.go:11–22`, `103–111` |
| W7-A1-006 | FIXED-OK | закрыт | `telegram_wave97_test.go` — format Contains + NotPanics cancel/send |
| W7-A1-007 | FIXED-OK | закрыт | `telegram_lifecycle_test.go` — TryRecv queue/overflow/missing-user |
| W7-A1-008 | FIXED-OK | закрыт | `push_wave95_test.go` — https-sub остаётся после send fail; `file://` удаляется |
| W7-A1-009 | FIXED-OK | закрыт | `TestIsQuietHours_CrossMidnight` — expected True/False — `service_extended_test.go:316–345` |
| W7-A1-010 | FIXED-OK | закрыт | `rdb.Del` → `slog.Warn` — `telegram_bind.go:127–130`, `:248–251` |

**REGRESS:** 0 · **RESIDUAL (хвост A2):** 0

### Cascade инварианты (фокус слота, без регрессии)

| Инвариант | Evidence |
|-----------|----------|
| NF-07 WebPush | `ValidateOutboundURL` на save (`push.go:227`) + send (`:166`); `safehttp.NewSafeHTTPClient` (`:34`); SafeDialContext defense-in-depth |
| Push ownership A2-005 | Update/Delete `UserID` Where — `push.go:233–281`; `push_scope_a2005_test.go` |
| Telegram ≠ login | Routes bind/confirm/unlink/test/status only — `handler.go:38–44`, `handler_telegram_bind.go`; bot `/start bind_*`, `/link`, `/unlink`, `/test` — нет JWT/session/password |
| Mail secrets | `InternalTokenAuth` + `subtle.ConstantTimeCompare` — `internal_auth.go:14–29`; empty expected → 503; Status без API key — `dto.go:14–20`; Resend key только в Header (`service.go:130`) |
| Rate limits (enqueue) | Push 50 ms ≈20/s (`push.go:91–98`); Telegram 35 ms ≈28/s (`telegram.go:66–72`); mail daily_limit enforced (см. NEW ниже — process-local) |

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/internal/notification/consumer.go` | 743 | 0 | FIXED-OK repoLookupErr → Nak |
| `backend/internal/notification/dto.go` | 71 | 0 | чисто |
| `backend/internal/notification/events.go` | 99 | 0 | чисто (TypeNewLogin = notify, не auth) |
| `backend/internal/notification/export_hooks.go` | 220 | 0 | чисто (hubtest) |
| `backend/internal/notification/handler.go` | 401 | 0 | FIXED-OK ErrInvalidPushEndpoint→400; owner-scoped |
| `backend/internal/notification/handler_telegram_bind.go` | 128 | 0 | чисто — Telegram ≠ login |
| `backend/internal/notification/push.go` | 298 | 0 | FIXED-OK NF-07 + SafeHTTPClient + A2-005 |
| `backend/internal/notification/repository.go` | 355 | 0 | чисто — UserID scope, Limit list |
| `backend/internal/notification/service.go` | 404 | 0 | чисто — prefs/quiet/dispatch |
| `backend/internal/notification/telegram.go` | 139 | 0 | чисто — rate + ctx cancel |
| `backend/internal/notification/telegram_bind.go` | 341 | 0 | FIXED-OK Del warn; bind-only |
| `backend/testframework/notification/consumer_ack_a2010_test.go` | 82 | 0 | FIXED-OK Nak/Ack |
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
| `backend/testframework/notification/handler_test.go` | 407 | 0 | FIXED-OK NF-07→400 |
| `backend/testframework/notification/handler_wave115_test.go` | 35 | 0 | чисто |
| `backend/testframework/notification/handler_wave116_test.go` | 23 | 0 | чисто |
| `backend/testframework/notification/handler_wave117_test.go` | 57 | 0 | чисто |
| `backend/testframework/notification/handler_wave118_test.go` | 68 | 0 | чисто |
| `backend/testframework/notification/handler_wave119_test.go` | 42 | 0 | чисто |
| `backend/testframework/notification/handler_wave99_test.go` | 45 | 0 | чисто |
| `backend/testframework/notification/push_lifecycle_test.go` | 60 | 0 | чисто |
| `backend/testframework/notification/push_scope_a2005_test.go` | 77 | 0 | FIXED-OK ownership |
| `backend/testframework/notification/push_wave119_test.go` | 57 | 1 | NEW — QueueOverflow без assert |
| `backend/testframework/notification/push_wave95_test.go` | 77 | 0 | FIXED-OK https keep / file delete |
| `backend/testframework/notification/repository_extra_test.go` | 146 | 0 | чисто |
| `backend/testframework/notification/resolve_prefs_a13005_test.go` | 68 | 0 | чисто |
| `backend/testframework/notification/service_extended_test.go` | 372 | 0 | FIXED-OK quiet hours |
| `backend/testframework/notification/service_test.go` | 134 | 0 | чисто |
| `backend/testframework/notification/service_wave111_test.go` | 178 | 0 | чисто |
| `backend/testframework/notification/service_wave117_test.go` | 103 | 0 | чисто |
| `backend/testframework/notification/service_wave118_test.go` | 57 | 0 | чисто |
| `backend/testframework/notification/service_wave119_test.go` | 62 | 0 | чисто |
| `backend/testframework/notification/stubs_test.go` | 34 | 0 | чисто (helpers) |
| `backend/testframework/notification/telegram_bind_test.go` | 127 | 0 | чисто |
| `backend/testframework/notification/telegram_bind_wave102_test.go` | 74 | 0 | чисто |
| `backend/testframework/notification/telegram_bind_wave95_test.go` | 128 | 0 | чисто |
| `backend/testframework/notification/telegram_lifecycle_test.go` | 114 | 0 | FIXED-OK queue asserts |
| `backend/testframework/notification/telegram_wave97_test.go` | 44 | 0 | FIXED-OK format/cancel |
| `backend/internal/mail/admin_broadcast.go` | 99 | 0 | FIXED-OK Failed count |
| `backend/internal/mail/admin_handler.go` | 120 | 0 | FIXED-OK no ent.Client |
| `backend/internal/mail/admin_service.go` | 116 | 0 | чисто |
| `backend/internal/mail/config.go` | 51 | 0 | чисто — secrets из env |
| `backend/internal/mail/dto.go` | 45 | 0 | чисто — Status без секретов |
| `backend/internal/mail/export_hooks.go` | 13 | 0 | чисто |
| `backend/internal/mail/hub_handler.go` | 54 | 0 | чисто — CTC gate на `/v1/send` |
| `backend/internal/mail/internal_auth.go` | 31 | 0 | FIXED-OK ConstantTimeCompare |
| `backend/internal/mail/service.go` | 203 | 1 | NEW — daily_limit process-local |
| `backend/internal/mail/settings_store.go` | 55 | 0 | чисто |
| `backend/testframework/mail/admin_handler_test.go` | 178 | 0 | FIXED-OK Failed |
| `backend/testframework/mail/admin_service_test.go` | 167 | 0 | FIXED-OK limit/alerts |
| `backend/testframework/mail/config_test.go` | 45 | 0 | чисто |
| `backend/testframework/mail/hub_handler_test.go` | 109 | 0 | чисто |
| `backend/testframework/mail/internal_auth_test.go` | 63 | 0 | FIXED-OK CTC paths |
| `backend/testframework/mail/send_alert_a7005_test.go` | 25 | 0 | чисто — no stub success |

Файлов в списке: **66**. Прочитано: **66**. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|----------|-------------|----------|---------------|------------|
| A3W7-1-001 | NEW | P3 | `mail/service.go:33–35`, `:182–194` | `DailyLimit` считается в process-local `perEmail map` под mutex | TASK-117 закрыл «декоративный» limit (enforcement есть), но счётчик не в Redis/БД: рестарт процесса обнуляет; N реплик Hub/`hub-mail` умножают лимит. Фокус rate limits — ложный потолок | Redis INCR+TTL до UTC midnight (или строка SystemSetting/таблица counters) shared между инстансами; тест: два `Service` → общий cap |
| A3W7-1-002 | NEW | P3 | `push_wave119_test.go:27–41` | `TestPushSender_QueueOverflow` шлёт 201 `Send` без assert | Очередь cap 200 (`push.go:71`); ветка `default` drop (`:123–125`) не проверяется — тест зелёный при любом cap/без drop | После цикла: `ExportTelegramTryRecv`-аналог для push queue depth ≤200 **или** счётчик dropped; mutation: убрать `default` → FAIL |

## FIXED-OK (сводка)

| ID | Evidence |
|----|----------|
| W7-A1-001 | `repoLookupErr` + Nak path; A2-010 тесты |
| W7-A1-002 | NF-07 → 400 |
| W7-A1-003 | SendAlert policy (+ store) — enforcement OK; distributed → NEW |
| W7-A1-004…010 | broadcast Failed, layering, test asserts, Del warn, quiet hours |

## Чисто

- **NF-07:** save+send ValidateOutboundURL; SafeHTTPClient; legacy `file://` purge на send.
- **Telegram ≠ login:** только notification bind; `TypeNewLogin` — in-app событие `auth.session.created`, не Telegram auth.
- **Mail secrets:** CTC internal token; empty → 503; Resend key не в Status/logs; LimitReader на error body.
- **Owner scope:** notifications/mark/delete/push sub по `userID`; prefs ResolveChannelPrefs.
- **Consumer:** explicit err → Nak; unknown subject → Ack; debounce 2s.
- **Resend URL:** hardcoded `https://api.resend.com/emails` (не user-controlled → SafeDialer N/A).

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 2 |
| **Total** | **2** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 / Stage A cascade) | 10 |
| RESIDUAL | 0 |
| NEW | 2 (`A3W7-1-001`, `A3W7-1-002`) |
| REGRESS | 0 |

**Топ IDs:** `A3W7-1-001` (mail daily_limit process-local) · `A3W7-1-002` (push overflow assert-free).

A2 baseline: 0 open → FIXED-OK 10 · NEW 2 · REGRESS 0 · RESIDUAL 0.
