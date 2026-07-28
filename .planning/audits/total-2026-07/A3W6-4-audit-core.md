# A3 W6 агент A4 — audit prod + half tests (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`)  
> Baseline: [A2W6-4-audit-core.md](A2W6-4-audit-core.md)  
> Remediation: TASK-090 (purge fail-closed + admin error contract) · TASK-114 (limits/layering/GDPR tx/tests)  
> Фокус: purge fail-closed · GDPR limits · `{"error","request_id"}` · layering  
> Периметр: 22 файла assignment (+ `gdpr_mongo.go`, `retention_store.go` для сверки фиксов) · Прочитано: 22+2 · Пропущено: 0

## Сверка с A2 (A2W6-A4-*)

| A2 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| A2W6-A4-001 | FIXED-OK | закрыт | `PurgeExpired` → `errors.Join(errs...)` (`service.go:333-358`); тест `service_purge_test.go` PropagatesError; TASK-090 |
| A2W6-A4-002 | FIXED-OK | закрыт | `UpdateResourceLimit` Atoi+`>0`; `CheckLimit` deny на empty/unknown/мусор (`service.go:167-211`); `task114_residual_test.go` |
| A2W6-A4-003 | FIXED-OK / RESIDUAL | Ent закрыт; Mongo — open | `WithTx` + cascade rollback (`repository.go:125-141`, `service.go:485-594`); Ent rollback тест `TestDeleteUserGDPR_CascadeRollback_A2W6A4003`. **Mongo eraser внутри tx без compensation** → A3W6-4-001 |
| A2W6-A4-004 | FIXED-OK | закрыт | `RetentionStore` / `retention_store.go`; `RetentionService.store` без `*ent.Client` в структуре (`retention.go:40-61`) |
| A2W6-A4-005 | FIXED-OK | закрыт | `defer recover` в recorder goroutine (`middleware.go:78-83`); `TestHTTPAuditMiddleware_RecorderPanicRecovered_A2W6A4005` |
| A2W6-A4-006 | FIXED-OK | закрыт | HTTP DELETE GDPR founder → 403 (`handler.go:432-433`); `TestHandler_DeleteUserGDPR_FounderForbidden_A2W6A4006` |
| A2W6-A4-007 | FIXED-OK | закрыт | admin 500 — фиксированные RU-сообщения + slog; без `err.Error()` (`handler.go:139+`); `admin_error_contract_task090_test.go` |
| A2W6-A4-008 | FIXED-OK | закрыт | godoc `ErrCannotDeleteFounder` / `DeleteUserGDPR` синхронизированы с P2-BE-19 (`service.go:20-22`, `:470-471`) |
| W6-A4-001…015 (Stage A) | FIXED-OK | без регрессии | cascade eraser wiring, founder ban, SHA-256 CTC alerts, IsSystemAdmin 500, Upsert retry, unknown subject Warn, ban audit fail-closed |

**REGRESS:** 0

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/audit/consumer.go` | 220 | 0 | FIXED-OK (Ack/Nak + Warn unknown) |
| `backend/internal/audit/dto.go` | 88 | 0 | чисто |
| `backend/internal/audit/export_hooks.go` | 81 | 0 | чисто (hubtest) |
| `backend/internal/audit/handler.go` | 453 | 1 | NEW P3 (ent import) |
| `backend/internal/audit/handler_alerts.go` | 123 | 0 | FIXED-OK (SHA-256 CTC) |
| `backend/internal/audit/middleware.go` | 91 | 0 | FIXED-OK (recover) |
| `backend/internal/audit/repository.go` | 816 | 0 | FIXED-OK (WithTx, Upsert retry) |
| `backend/internal/audit/retention.go` | 261 | 0 | FIXED-OK (RetentionStore) |
| `backend/internal/audit/service.go` | 608 | 2 | RESIDUAL Mongo; BanUser sentinels P3 |
| `backend/internal/audit/gdpr_mongo.go` | 41 | 0 | FIXED-OK (fail-closed nil coll; вне list) |
| `backend/internal/audit/retention_store.go` | 64 | 0 | FIXED-OK (вне list; layering) |
| `backend/testframework/audit/consumer_ack_a1008_test.go` | 91 | 0 | чисто |
| `backend/testframework/audit/consumer_lifecycle_test.go` | 54 | 0 | чисто |
| `backend/testframework/audit/consumer_misc_test.go` | 42 | 0 | чисто |
| `backend/testframework/audit/consumer_test.go` | 152 | 0 | чисто |
| `backend/testframework/audit/dto_mapping_test.go` | 76 | 0 | чисто |
| `backend/testframework/audit/founder_ban_a1009_test.go` | 61 | 0 | чисто |
| `backend/testframework/audit/gdpr_founder_a1004_test.go` | 56 | 0 | чисто |
| `backend/testframework/audit/handler_alerts_test.go` | 79 | 0 | чисто |
| `backend/testframework/audit/handler_founder_ban_a1009_test.go` | 68 | 0 | чисто |
| `backend/testframework/audit/handler_limits_a1013_test.go` | 71 | 0 | чисто (allowlist) |
| `backend/testframework/audit/handler_test.go` | 257 | 0 | чисто |
| `backend/testframework/audit/handler_user_mgmt_test.go` | 184 | 0 | FIXED-OK (GDPR HTTP в task114) |
| `backend/testframework/audit/middleware_test.go` | 63 | 0 | чисто |

Файлов assignment: **22/22**. Доп.: `gdpr_mongo.go`, `retention_store.go`. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|----------|-------------------|------------|
| A3W6-4-001 | RESIDUAL (A2W6-A4-003) | P2 | `service.go:569-576` | `msgEraser.DeleteMessagesBySender` вызывается **внутри** Ent `WithTx`; при успехе Mongo и последующем fail HardDelete/audit → Ent Rollback, Mongo messages уже стёрты | Cross-store partial erasure: user жив, чат-история потеряна; A2 fix закрыл только Ent cascade | Mongo **после** Commit; или saga/compensating re-insert; или outbox «erase after commit»; тест: mock eraser ok + HardDelete fail → assert messages intact **или** documented compensating path |
| A3W6-4-002 | NEW | P3 | `handler.go:342-344` · `service.go:398-402` | Ban validation через `err.Error()` string match (`"укажите причину…"`, `"нельзя заблокировать…"`) | Хрупко к правке текста; не sentinel/`errors.Is` | `ErrBanReasonRequired` / `ErrCannotBanSelf` + `errors.Is` в handler |
| A3W6-4-003 | NEW | P3 | `handler.go:13`, `:59` | `requireAdmin` импортирует `ent` ради `ent.IsNotFound` | TZ layering: Handler → Ent; NotFound должен маппиться в service/repo sentinel | `repo.IsSystemAdmin` → typed err / `ErrUserNotFound`; handler без `ent` |

## Чисто (фокус-инварианты)

- **Purge fail-closed:** любой `purge_*` err → Join наружу; cron wiring `PurgeExpired` получает err (TASK-090).
- **Limits fail-closed:** allowlist keys; non-positive/`abc` → `ErrInvalidResourceLimitValue`; CheckLimit deny на мусор/unknown empty (TASK-114).
- **Error contract admin 500:** List/GetLimits/Update/Trash/Restore/Deactivate/Ban(infra)/Reassign/GDPR — fixed messages; leak-тест TASK-090.
- **GDPR founder:** service + HTTP 403; self-erase founder запрещён; ban/deactivate founder 403.
- **Layering AuditService:** только `Repository` (+ optional Redis/MessageEraser); Retention через `RetentionStore`; prod `SetMessageEraser(NewMongoMessageEraser)` в `wiring_messenger.go`.
- **Alerts:** SHA-256 digests + CTC; empty secret → webhook не регистрируется.
- **Consumer:** Create/JSON fail → Nak; success → Ack; unknown subject → Warn+Ack.
- **Тесты A4 half:** a1008/a1009/a1004/a1013 + middleware + handler user mgmt — без регрессии; residual-suite TASK-114 вне list, но подтверждает closes.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 1 |
| P3 | 2 |
| **Total** | **3** |

A2 baseline open: 8 → FIXED-OK 7 (+ Ent half of 003) · REGRESS 0 · RESIDUAL 1 (Mongo half A2W6-A4-003) · NEW 2.
