# W7 агент A1 — notification+mail

> TASK-008 · Стадия A · read-only · 2026-07-28  
> Фокус: WebPush `ValidateOutboundURL` (NF-07) · Telegram bind≠login · consumer Ack/Nak · mail `ConstantTimeCompare`

## Прочитанные файлы
| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/internal/notification/consumer.go | 722 | 1 | замечания |
| backend/internal/notification/dto.go | 71 | 0 | чисто |
| backend/internal/notification/events.go | 99 | 0 | чисто |
| backend/internal/notification/export_hooks.go | 220 | 0 | чисто |
| backend/internal/notification/handler.go | 397 | 1 | замечания |
| backend/internal/notification/handler_telegram_bind.go | 128 | 0 | чисто |
| backend/internal/notification/push.go | 295 | 0 | чисто |
| backend/internal/notification/repository.go | 343 | 0 | чисто |
| backend/internal/notification/service.go | 397 | 0 | чисто |
| backend/internal/notification/telegram.go | 139 | 0 | чисто |
| backend/internal/notification/telegram_bind.go | 335 | 1 | замечания |
| backend/testframework/notification/consumer_ack_a2010_test.go | 82 | 0 | чисто |
| backend/testframework/notification/consumer_extra_test.go | 173 | 0 | чисто |
| backend/testframework/notification/consumer_test.go | 290 | 0 | чисто |
| backend/testframework/notification/consumer_wave102_test.go | 153 | 0 | чисто |
| backend/testframework/notification/consumer_wave118_test.go | 160 | 0 | чисто |
| backend/testframework/notification/consumer_wave119_test.go | 238 | 0 | чисто |
| backend/testframework/notification/consumer_wave121_test.go | 197 | 0 | чисто |
| backend/testframework/notification/consumer_wave94_test.go | 46 | 0 | чисто |
| backend/testframework/notification/consumer_wave99_test.go | 60 | 0 | чисто |
| backend/testframework/notification/coverage_wave79_test.go | 261 | 0 | чисто |
| backend/testframework/notification/fixtures_test.go | 85 | 0 | чисто |
| backend/testframework/notification/handler_list_filter_test.go | 62 | 0 | чисто |
| backend/testframework/notification/handler_telegram_bind_test.go | 149 | 0 | чисто |
| backend/testframework/notification/handler_telegram_bind_wave102_test.go | 122 | 0 | чисто |
| backend/testframework/notification/handler_test.go | 406 | 0 | чисто |
| backend/testframework/notification/handler_wave115_test.go | 35 | 0 | чисто |
| backend/testframework/notification/handler_wave116_test.go | 23 | 0 | чисто |
| backend/testframework/notification/handler_wave117_test.go | 57 | 0 | чисто |
| backend/testframework/notification/handler_wave118_test.go | 68 | 0 | чисто |
| backend/testframework/notification/handler_wave119_test.go | 42 | 0 | чисто |
| backend/testframework/notification/handler_wave99_test.go | 45 | 0 | чисто |
| backend/testframework/notification/push_lifecycle_test.go | 60 | 0 | чисто |
| backend/testframework/notification/push_scope_a2005_test.go | 77 | 0 | чисто |
| backend/testframework/notification/push_wave119_test.go | 57 | 0 | чисто |
| backend/testframework/notification/push_wave95_test.go | 65 | 1 | замечания |
| backend/testframework/notification/repository_extra_test.go | 146 | 0 | чисто |
| backend/testframework/notification/resolve_prefs_a13005_test.go | 68 | 0 | чисто |
| backend/testframework/notification/service_extended_test.go | 352 | 1 | замечания |
| backend/testframework/notification/service_test.go | 134 | 0 | чисто |
| backend/testframework/notification/service_wave111_test.go | 178 | 0 | чисто |
| backend/testframework/notification/service_wave117_test.go | 103 | 0 | чисто |
| backend/testframework/notification/service_wave118_test.go | 57 | 0 | чисто |
| backend/testframework/notification/service_wave119_test.go | 62 | 0 | чисто |
| backend/testframework/notification/stubs_test.go | 34 | 0 | чисто |
| backend/testframework/notification/telegram_bind_test.go | 127 | 0 | чисто |
| backend/testframework/notification/telegram_bind_wave102_test.go | 74 | 0 | чисто |
| backend/testframework/notification/telegram_bind_wave95_test.go | 128 | 0 | чисто |
| backend/testframework/notification/telegram_lifecycle_test.go | 99 | 1 | замечания |
| backend/testframework/notification/telegram_wave97_test.go | 34 | 1 | замечания |
| backend/internal/mail/admin_broadcast.go | 70 | 1 | замечания |
| backend/internal/mail/admin_handler.go | 113 | 1 | замечания |
| backend/internal/mail/admin_service.go | 115 | 0 | чисто |
| backend/internal/mail/config.go | 43 | 0 | чисто |
| backend/internal/mail/dto.go | 41 | 0 | чисто |
| backend/internal/mail/export_hooks.go | 13 | 0 | чисто |
| backend/internal/mail/hub_handler.go | 48 | 0 | чисто |
| backend/internal/mail/internal_auth.go | 31 | 0 | чисто |
| backend/internal/mail/service.go | 111 | 1 | замечания |
| backend/internal/mail/settings_store.go | 55 | 0 | чисто |
| backend/testframework/mail/admin_handler_test.go | 145 | 0 | чисто |
| backend/testframework/mail/admin_service_test.go | 114 | 0 | чисто |
| backend/testframework/mail/config_test.go | 44 | 0 | чисто |
| backend/testframework/mail/hub_handler_test.go | 106 | 0 | чисто |
| backend/testframework/mail/internal_auth_test.go | 63 | 0 | чисто |
| backend/testframework/mail/send_alert_a7005_test.go | 25 | 0 | чисто |

Файлов в списке: 66. Прочитано: 66. Пропущено: 0.

## Фокус TZ (сверка)

| Инвариант | Статус | Доказательство |
|-----------|--------|----------------|
| NF-07 WebPush `ValidateOutboundURL` | **OK** | `push.go:164` перед send; `push.go:225` в `SaveSubscription`; client = `safehttp.NewSafeHTTPClient` (`push.go:32`); тесты `file://`/`javascript:` reject |
| Push ownership (A2-005) | **OK** | Update/Delete с `UserID` в Where; `push_scope_a2005_test.go` |
| Telegram bind ≠ login | **OK** | Routes только `/telegram/{bind,confirm,test,status}` + DELETE; bot: `/start bind_*`, `/link`, `/unlink`, `/test` — нет JWT/session/login |
| Consumer Ack/Nak (A2-010) | **частично** | `processEvent`: success→Ack, err→Nak, unknown→Ack; тест `consumer_ack_a2010_test.go`. Gap: soft `return nil` на repo/DB err → Ack (W7-A1-001) |
| Mail `ConstantTimeCompare` | **OK** | `internal_auth.go:26`; empty expected → 503; тесты header/Bearer/wrong/empty |
| DailyLimit / AlertsEnabled | **дыра** | Админ пишет settings; `SendAlert` не читает лимит/флаг (W7-A1-003) |

## Находки
| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W7-A1-001 | P1 | `backend/internal/notification/consumer.go:253-254` (+448-450, 488-490, 547-548, 596-597, 331-332, 381-382) | При ошибке repo (`GetIssueAssignee`/`GetChannel`/`GetChannelMembers`/`GetProjectMembersWithUser`) handler возвращает `nil` → `processEvent` делает **Ack** | Transient DB/outage → событие навсегда потеряно; `MaxDeliver`/`Nak` бесполезны. A2-010 закрыл только явный `processErr`, не soft-nil | Различать not-found (`nil,nil` / `ent.IsNotFound`) → Ack; прочие `err` → return err → Nak |
| W7-A1-002 | P2 | `backend/internal/notification/handler.go:318-325` | Ошибка `ValidateOutboundURL` из `SaveSubscription` мапится в **500**, не 400 | Клиент с `javascript:`/`file:` получает Internal Server Error; тест `handler_test.go:317` **закрепляет** 500 | `errors.Is`/`strings.Contains` на validation → 400; обновить тест |
| W7-A1-003 | P2 | `backend/internal/mail/service.go:51-99` (+`Status` 103-110) | `DailyLimit` / `alerts_enabled` из admin settings **не enforcement** в `SendAlert`; `Status().AlertsEnabled` всегда `true` | Admin UI/API обещают лимит и выключатель; фактически unlimited send при наличии ключа | Читать лимит/флаг (Redis/DB counter per-day); при `alerts_enabled=false` — fail; Status из cfg/store |
| W7-A1-004 | P2 | `backend/internal/mail/admin_broadcast.go:57-68` | Ошибка `sender.SendAlert` глотается (`continue`), в ответе только `Sent` без failed | Частичный фейл рассылки выглядит как успех; нет логов/счётчика ошибок | Логировать err; `Failed` в response; опционально abort после N ошибок |
| W7-A1-005 | P2 | `backend/internal/mail/admin_handler.go:12-18` | `AdminHandler` держит `*ent.Client` и передаёт в `Broadcast` | TZ/layering: `*ent.Client` не в Handler | User-list через `SettingsStore`/repo interface; db только в store |
| W7-A1-006 | P2 | `backend/testframework/notification/telegram_wave97_test.go:17-34` | Оба теста без assert (только `require.NotNil` + вызов) | Assert-free: проходят при сломанном cancel/format path | Assert: после cancel нет panic; format через `ExportFormatTelegramMessage` / err branch observable |
| W7-A1-007 | P2 | `backend/testframework/notification/telegram_lifecycle_test.go:32-47,93-99` | `SendAndQueueOverflow`, `StartStopWithoutWorkerSend`, `SendTelegram_MissingUser` — без assert | Зелёные при регрессии drop/queue/nil-user | Assert queue depth / try-recv empty / NotPanics явно |
| W7-A1-008 | P2 | `backend/testframework/notification/push_wave95_test.go:19-38` | `ValidEndpointSendFails` вызывает send без assert | Не ловит регрессию (panic/удаление валидной sub) | Assert sub остаётся в БД; err path не удаляет https-endpoint |
| W7-A1-009 | P3 | `backend/testframework/notification/service_extended_test.go:314-326` | `IsQuietHours_CrossMidnight`: `_ = ExportIsQuietHours(...)` | Имя обещает cross-midnight; значения не проверяются | Зафиксировать TZ/час clock или таблицу start/end → expected bool |
| W7-A1-010 | P3 | `backend/internal/notification/telegram_bind.go:127,245` | `rdb.Del` после успешного bind/confirm без проверки err | Токен/код может остаться в Redis → повторное использование окна TTL | Проверять/логировать Del err; желательно Del до link или pipeline |

## Counts

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 1 |
| P2 | 7 |
| P3 | 2 |
| **Σ** | **10** |

Файлов прочитано: **66**/66. С находками: **10**. Чисто: **56**.

## Топ-3

1. **W7-A1-001** — soft `return nil` на DB/repo error в consumer → Ack → потеря уведомлений при сбоях (ломает смысл A2-010 Nak).
2. **W7-A1-003** — mail `DailyLimit`/`alerts_enabled` — декоративные settings, `SendAlert` не enforcement.
3. **W7-A1-002** — NF-07 reject endpoint отдаётся как 500 (тест закрепляет), а не 400.

## Чисто

Файлы без находок — что проверялось:

- **push.go** — NF-07 на save+send; SafeHTTPClient; 404/410 cleanup; A2-005 ownership Where.
- **handler_telegram_bind.go** — только bind/confirm/unlink/test/status; JWT userID; нет login.
- **telegram.go** — queue+rate; ctx-aware send; MarkdownV2 escape; nil-safe.
- **telegram_bind.go** (кроме Del) — Redis TTL tokens; ID taken check; bot commands notification-only.
- **internal_auth.go** — `subtle.ConstantTimeCompare`; empty token → 503; Bearer fallback.
- **consumer_ack_a2010_test.go** — Nak bad JSON / Dispatch fail; Ack success.
- **push_scope_a2005_test.go** — foreign hijack/unsub denied на Ent-слое.
- **dto/events/export_hooks/repository/service/config/dto mail/hub_handler/settings_store** — owner-scoped list/mark/delete; SubjectToType; ResolveChannelPrefs; LimitReader на Resend body; hardcoded Resend URL (не user-controlled).
- **Остальные *_test.go** — assert на HTTP/коды/типы/debounce/prefs/A7-005/A7-008; stubs для JetStream Msg.

## Verification (агент)

- Режим: **read-only** (код не менялся).
- Артефакт: `.planning/audits/total-2026-07/A7-1-notification-mail.md`
- ID: `W7-A1-001` … `W7-A1-010`
