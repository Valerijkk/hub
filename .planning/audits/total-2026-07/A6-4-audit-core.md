# W6 агент A4 — audit prod + half tests

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: founder ban (A1-009), GDPR founder (A1-004), alert `ConstantTimeCompare`, retention (BE-H7), consumer Ack/Nak (A1-008).  
> Сверка: grok-2026-07 A1-004/008/009 CLOSED; BE-H6/H7; ORG-04.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/audit/consumer.go` | 218 | 1 | замечания |
| `backend/internal/audit/dto.go` | 88 | 0 | чисто |
| `backend/internal/audit/export_hooks.go` | 76 | 0 | чисто |
| `backend/internal/audit/handler.go` | 428 | 2 | замечания |
| `backend/internal/audit/handler_alerts.go` | 116 | 1 | замечания |
| `backend/internal/audit/middleware.go` | 85 | 1 | замечания |
| `backend/internal/audit/repository.go` | 639 | 1 | замечания |
| `backend/internal/audit/retention.go` | 265 | 1 | замечания |
| `backend/internal/audit/service.go` | 518 | 5 | замечания |
| `backend/testframework/audit/consumer_ack_a1008_test.go` | 91 | 0 | чисто |
| `backend/testframework/audit/consumer_lifecycle_test.go` | 54 | 0 | чисто |
| `backend/testframework/audit/consumer_misc_test.go` | 42 | 0 | чисто |
| `backend/testframework/audit/consumer_test.go` | 148 | 1 | замечания |
| `backend/testframework/audit/dto_mapping_test.go` | 76 | 0 | чисто |
| `backend/testframework/audit/founder_ban_a1009_test.go` | 61 | 0 | чисто |
| `backend/testframework/audit/gdpr_founder_a1004_test.go` | 53 | 0 | чисто |
| `backend/testframework/audit/handler_alerts_test.go` | 79 | 0 | чисто |
| `backend/testframework/audit/handler_founder_ban_a1009_test.go` | 68 | 0 | чисто |
| `backend/testframework/audit/handler_limits_a1013_test.go` | 71 | 0 | чисто |
| `backend/testframework/audit/handler_test.go` | 257 | 0 | чисто |
| `backend/testframework/audit/handler_user_mgmt_test.go` | 178 | 1 | замечания |
| `backend/testframework/audit/middleware_test.go` | 63 | 0 | чисто |

Файлов в списке: 22. Прочитано: 22. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W6-A4-001 | P1 | `service.go:408-507`; `repository.go:548-619` | `DeleteUserGDPR` стирает user + comments/reactions/sessions/activity/time_logs, но **не** трогает Mongo `messages`, `FileMetadata` uploader, audit_log с IP/UA/PII, memberships | Эндпоинт заявлен как GDPR / right to erasure; после 200/500 с hard-delete в чатах/файлах/журнале остаются персональные данные удалённого пользователя | Расширить каскад (или anonymize): Mongo messages по `sender_id`, files, scrub audit details; тест «после GDPR нет PII по user_id» на слое репозитория/Mongo |
| W6-A4-002 | P1 | `service.go:447-503` | GDPR-каскад **без транзакции**: HardDelete может пройти при `cascadeErrs != nil` → handler 500, пользователь пользователь уже удалён; либо cascade успел удалить comments, а HardDelete упал | Повтор DELETE неидемпотентен / частичное erasure; админ видит ошибку, но состояние «user gone, PII leftover» или «user alive, comments gone» | Одна Ent-tx (+ Mongo шаги с компенсирующей логикой); при cascadeErr **до** HardDelete — abort; после успешного erase — отдельный repair job |
| W6-A4-003 | P2 | `service.go:483-489` | Audit `gdpr_deletion`: `user_id` = **удаляемый**, в details только `deleted_user_id`; **`actorID` не пишется** | Нельзя установить, какой admin инициировал erase (accountability / SOC2-style); после delete user_id в логе — «мертвая» ссылка | `CreateAuditLog(..., &actorID, ..., {"deleted_user_id","deleted_by": actorID})`; HTTP-тест что details содержат actor |
| W6-A4-004 | P2 | `service.go:285-305` | `PurgeExpired` глотает все ошибки purge_* (`slog.Error`), **всегда `return nil`** | Cron `trash-purge-expired` считает tick успешным при полном отказе БД → корзина растёт бесконечно без алерта | Аккумулировать `errors.Join` / вернуть первую ошибку (как `RunRetentionCleanup`); тест: mock fail → `err != nil` |
| W6-A4-005 | P2 | `service.go:150-178`; `handler.go:170-188` | `UpdateResourceLimit` не валидирует numeric/positive; `CheckLimit` при `Atoi` err / пустом значении — **fail-open** (`return true`) | Admin может записать `"abc"`; потребители лимита перестают ограничивать storage/projects | Валидировать `Atoi` + `>0` в Update; CheckLimit fail-closed или default; тест table-driven |
| W6-A4-006 | P2 | `retention.go:40-57`; `wiring_audit_ai.go:104` | `RetentionService` держит `*ent.Client` напрямую (TZ: нет `*ent.Client` в Service) | Layering drift vs `AuditService`→Repository; тесты retention тянут Ent; сложнее мокать | Вынести Recording/SystemSetting запросы в Repository (или узкий `RetentionStore`) |
| W6-A4-007 | P2 | `middleware.go:77-81` | `go recorder(...)` без `recover`; Background+5s timeout | Panic в recorder (или зависимости) → падение всего процесса на любом `/api` запросе | `defer recover` + slog в goroutine; опционально worker pool / sync на shutdown |
| W6-A4-008 | P2 | `handler_user_mgmt_test.go:151-165` vs `handler_founder_ban_a1009_test.go` | HTTP DELETE `/admin/users/:id/gdpr` для founder **не** покрыт 403 (есть только service `gdpr_founder_a1004` + happy-path AdminOK) | Регресс маппинга `ErrCannotDeleteFounder`→500/200 не ловится (класс WEAK A1-009 до HTTP-теста) | `TestHandler_DeleteUserGDPR_FounderForbidden` по образцу ban/deactivate HTTP |
| W6-A4-009 | P2 | `service.go:423-430` | Founder **self-erase** разрешён; нет передачи `system_role=founder`, нет блокировки «последний founder» | ORG-04: org остаётся без единственного founder после успешного GDPR self-delete | Запретить self-erase founder **или** обязательный reassign founder до erase; продукт/ADR |
| W6-A4-010 | P3 | `handler_alerts.go:53` | `subtle.ConstantTimeCompare([]byte(token), secretBytes)` — early-exit при разной длине (как A5-016 gitlab) | Timing oracle длины `ALERT_WEBHOOK_SECRET` | Сравнивать SHA-256(token) vs SHA-256(secret) fixed-size, или pad |
| W6-A4-011 | P3 | `handler.go:54-57` | `IsSystemAdmin` err → **401** «пользователь не найден», не 500 | Маскирует DB outage под unauthenticated; усложняет ops | 500 на repo err; 401 только на not found |
| W6-A4-012 | P3 | `repository.go:233-251` | `UpsertSystemSetting`: Update→если n==0 Create — гонка двух Create | Unique violation на concurrent PUT limits | `ON CONFLICT` / Ent upsert; или serialize |
| W6-A4-013 | P3 | `consumer.go:74-79` | Unknown subject → Ack **без** slog | Тихая потеря misrouted events; сложно отладить filter/subject drift | `slog.Debug`/`Warn` + metric `audit_unknown_subject` |
| W6-A4-014 | P3 | `consumer_test.go:125-133` | `TestProcessEvent_SkipsInvalidJSON` — только `createCalls` empty; stub Ack/Nak no-op; имя «Skips» устарело после A1-008 | Не усиливает mutation-proof Nak (есть в a1008); вводит в заблуждение | Переименовать / делегировать a1008; удалить дубль или assert через counting msg |
| W6-A4-015 | P3 | `service.go:360-363` | Ban успешен, если `CreateAuditLog` для `"ban"` упал (только slog) | Ban без audit trail при DB blip на audit_log | Вернуть err / Join; или outbox; тест fail CreateAuditLog → err |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| **Founder ban A1-009** | `deactivateUser` / `BanUser` → `ErrCannotModifyFounder`; handler 403; service+HTTP тесты mutation-oriented (`deactivateIDs` empty) |
| **GDPR founder A1-004** | Admin→founder deny; founder self-erase allow (service); handler maps `ErrCannotDeleteFounder`→403 |
| **Alert ConstantTimeCompare** | `handler_alerts.go:53` Bearer; empty secret → route не регистрируется; bad secret 401; empty→404 |
| **Consumer Ack A1-008** | Ack только после успешного `CreateAuditLog`; Nak на JSON/DB err; unknown subject Ack; tests counting acks/naks |
| **Retention BE-H7** | S3 DeleteObject до DB delete по snapshot IDs; nil S3 → WARN metadata-only; indexes sparse `created_at`; defaults 730/365/180 |
| **Limits A1-013** | Allowlist 3 keys; HTTP reject `retention.*` |
| **Trash 30d** | `TrashRetentionDays=30`; list `DeletedAtGT(cutoff)`; purge `DeletedAtLT` |
| `dto.go` | Только DTO, без логики |
| `export_hooks.go` | hubtest re-export processEvent/retention/helpers |
| `consumer_ack_a1008_test.go` | Nak on create err / bad JSON; Ack on success |
| `founder_ban_a1009_test.go` / `handler_founder_ban_*` | Service + HTTP 403 |
| `gdpr_founder_a1004_test.go` | Deny admin / allow self |
| `handler_alerts_test.go` | Record + reject + disabled |
| `handler_limits_a1013_test.go` | Reject unknown / allow 3 keys |
| `dto_mapping_test.go` | UUID optional mapping |
| `middleware_test.go` | Skip health / record API |
| `consumer_lifecycle` / `misc` | Start/Stop NATS; nil Stop; nil mongo purge |

**Фокус-вердикт:** A1-004/008/009 и ConstantTimeCompare на alert **закрыты в коде** с регрессионными тестами. Главный остаточный риск — **неполнота GDPR-каскада и отсутствие транзакции** (P1), плюс ops/layering вокруг purge/retention.

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **22 / 22 / 0** |
| Находок | **15** |
| P0 | 0 |
| P1 | 2 |
| P2 | 7 |
| P3 | 6 |

### Топ-3

1. **W6-A4-001 (P1)** — GDPR erase не чистит Mongo-сообщения / файлы / PII в audit_log → right to erasure неполный.
2. **W6-A4-002 (P1)** — GDPR без транзакции: HardDelete + частичный cascade → неконсистентное состояние при ошибках.
3. **W6-A4-004 (P2)** — `PurgeExpired` всегда `nil` → silent failure cron корзины 30d.
