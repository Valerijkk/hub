# A2 W6 агент A4 — audit prod + half tests

> TASK-008 · Stage A2 · read-only · 2026-07-28  
> Периметр: `W6-ASSIGNMENTS.md` §A4 (22 файла) + `gdpr_mongo.go` (новый с Stage A, домен audit)

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/internal/audit/consumer.go | 220 | 0 | FIXED-OK (unknown subject Warn) |
| backend/internal/audit/dto.go | 88 | 0 | чисто |
| backend/internal/audit/export_hooks.go | 81 | 0 | чисто |
| backend/internal/audit/handler.go | 435 | 2 | NEW + RESIDUAL |
| backend/internal/audit/handler_alerts.go | 123 | 0 | FIXED-OK (SHA-256 CTC) |
| backend/internal/audit/middleware.go | 85 | 1 | RESIDUAL |
| backend/internal/audit/repository.go | 698 | 0 | FIXED-OK (Upsert retry) |
| backend/internal/audit/retention.go | 265 | 1 | RESIDUAL (*ent.Client) |
| backend/internal/audit/service.go | 559 | 3 | RESIDUAL + FIXED + NEW comment |
| backend/internal/audit/gdpr_mongo.go | 41 | 0 | FIXED-OK (вне Stage A list; fail-closed nil) |
| backend/testframework/audit/consumer_ack_a1008_test.go | 91 | 0 | чисто |
| backend/testframework/audit/consumer_lifecycle_test.go | 54 | 0 | чисто |
| backend/testframework/audit/consumer_misc_test.go | 42 | 0 | чисто |
| backend/testframework/audit/consumer_test.go | 152 | 0 | чисто |
| backend/testframework/audit/dto_mapping_test.go | 76 | 0 | чисто |
| backend/testframework/audit/founder_ban_a1009_test.go | 61 | 0 | чисто |
| backend/testframework/audit/gdpr_founder_a1004_test.go | 52 | 0 | чисто (service) |
| backend/testframework/audit/handler_alerts_test.go | 79 | 0 | чисто |
| backend/testframework/audit/handler_founder_ban_a1009_test.go | 68 | 0 | чисто (ban/deact) |
| backend/testframework/audit/handler_limits_a1013_test.go | 71 | 0 | чисто |
| backend/testframework/audit/handler_test.go | 257 | 0 | чисто |
| backend/testframework/audit/handler_user_mgmt_test.go | 178 | 1 | RESIDUAL (нет GDPR founder HTTP) |
| backend/testframework/audit/middleware_test.go | 63 | 0 | чисто |

Файлов assignment: **22/22**. Доп. доменный: `gdpr_mongo.go` прочитан.

## Перепроверка Stage A (W6-A4)

| Старый ID | Tag | Статус | Комментарий |
|-----------|-----|--------|-------------|
| W6-A4-001 | FIXED-OK | закрыт | cascade: files, memberships, scrub PII, MessageEraser; wiring `SetMessageEraser` |
| W6-A4-002 | RESIDUAL | частичный | abort до HardDelete OK; **нет** rollback уже выполненных cascade-шагов |
| W6-A4-003 | FIXED-OK | закрыт | `deleted_by` / actorID в details |
| W6-A4-004 | RESIDUAL | жив | `PurgeExpired` глотает err, всегда `return nil` |
| W6-A4-005 | RESIDUAL | жив | `CheckLimit` Atoi/empty → fail-open; Update без numeric validate |
| W6-A4-006 | RESIDUAL | жив | `RetentionService` держит `*ent.Client` |
| W6-A4-007 | RESIDUAL | жив | `go recorder` без `recover` |
| W6-A4-008 | RESIDUAL | жив | HTTP GDPR founder 403 тест отсутствует |
| W6-A4-009 | FIXED-OK | закрыт | founder (в т.ч. self) → `ErrCannotDeleteFounder` |
| W6-A4-010 | FIXED-OK | закрыт | SHA-256 digests + CTC fixed-size |
| W6-A4-011 | FIXED-OK | закрыт | IsSystemAdmin DB err → 500 |
| W6-A4-012 | FIXED-OK | закрыт | Upsert constraint → retry Update |
| W6-A4-013 | FIXED-OK | закрыт | unknown subject → `slog.Warn` |
| W6-A4-014 | FIXED-OK | закрыт/низкий | naming/Ack покрыт a1008 — не поднимаем |
| W6-A4-015 | FIXED-OK | закрыт | Ban: CreateAuditLog err → return err |

REGRESS: **0**.

## Находки (открытые)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|----------|---------------|------------|
| A2W6-A4-001 | RESIDUAL | P2 | `service.go:297-317` | `PurgeExpired` всегда `nil` при purge err | Cron «успех» при полном отказе БД; TASK-040 не трогал audit | `errors.Join` / return err |
| A2W6-A4-002 | RESIDUAL | P2 | `service.go:174-187` | CheckLimit empty/Atoi → allow; Update без `>0` | Admin `"abc"` → fail-open limits | Validate Atoi+positive; CheckLimit fail-closed |
| A2W6-A4-003 | RESIDUAL | P2 | `service.go:454-521` | Cascade abort без rollback prior steps | Comments deleted, HardDelete aborted → partial erasure | Ent tx (+ compensations Mongo) |
| A2W6-A4-004 | RESIDUAL | P2 | `retention.go:40-51` | `*ent.Client` в Service | TZ layering | RetentionStore / Repository |
| A2W6-A4-005 | RESIDUAL | P2 | `middleware.go:77-81` | goroutine recorder без recover | Panic → process crash на `/api` | `defer recover` + slog |
| A2W6-A4-006 | RESIDUAL | P2 | `handler_user_mgmt_test.go` | Нет HTTP DELETE GDPR founder → 403 | Регресс маппинга Err→status | `TestHandler_DeleteUserGDPR_FounderForbidden` |
| A2W6-A4-007 | NEW | P2 | `handler.go:139+` | Несколько admin paths: `fiber.NewError(500, err.Error())` | Утечка internals ≠ `{"error","request_id"}` канон | Фиксированные сообщения; детали в slog |
| A2W6-A4-008 | NEW | P3 | `service.go:20-22` vs `:429-437` | Godoc `ErrCannotDeleteFounder`: «self-erase разрешён»; код запрещает всегда | Drift после P2-BE-19 | Синхронизировать комментарий с кодом |

## Counts

| | n |
|--|--:|
| NEW | 2 |
| REGRESS | 0 |
| RESIDUAL | 6 |
| FIXED-OK | 9 |
| P0 / P1 / P2 / P3 (open) | 0 / 0 / 7 / 1 |
| **Σ open** | **8** |
