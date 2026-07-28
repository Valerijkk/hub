# A2 W4 агент 2 — platform nats+outbox+redis

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Зеркало: `W4-ASSIGNMENTS.md` A2 · Stage A: `A4-2-platform-nats-outbox.md`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/platform/nats/client.go` | 208 | 1 | замечания |
| `backend/internal/platform/nats/export_hooks.go` | 16 | 0 | чисто |
| `backend/internal/platform/nats/trace.go` | 59 | 0 | чисто |
| `backend/internal/platform/outbox/export_hooks.go` | 54 | 0 | чисто |
| `backend/internal/platform/outbox/publisher.go` | 48 | 0 | чисто |
| `backend/internal/platform/outbox/worker.go` | 210 | 2 | замечания |
| `backend/internal/platform/redis/client.go` | 21 | 0 | чисто |
| `backend/testframework/platform/nats/client_integration_test.go` | 96 | 0 | чисто |
| `backend/testframework/platform/nats/client_publish_test.go` | 82 | 0 | чисто |
| `backend/testframework/platform/nats/client_test.go` | 26 | 1 | замечания |
| `backend/testframework/platform/nats/redact_url_test.go` | 13 | 0 | чисто |
| `backend/testframework/platform/outbox/outbox_test.go` | 288 | 0 | чисто |
| `backend/testframework/platform/outbox/publisher_test.go` | 45 | 0 | чисто |
| `backend/testframework/platform/outbox/trace_context_test.go` | 129 | 0 | чисто |
| `backend/testframework/platform/outbox/worker_wave101_test.go` | 31 | 0 | чисто |

Файлов в списке: 15. Прочитано: 15. Пропущено: 0.  
(Доп. вне списка, сверка: `publish_rollback_task059_test.go` — FIXED-OK W4-A2-007.)

## Находки (активные)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A2W4-2-001 | NEW | P2 | `outbox/worker.go:62-101`; `:123-145`; `:189-196` | `claimPending` держит `FOR UPDATE SKIP LOCKED` **на всё время** `PublishMsg`+`FlushTimeout` (до 50×5s) внутри одной tx до Commit | Долгие row-lock’и блокируют параллельных outbox-воркеров/реплики; рост latency БД под всплеском | Claim → `processing`+lease, Commit; publish вне lock; либо `PublishMsgAck` батчем с коротким claim |
| A2W4-2-002 | RESIDUAL | P2 | `nats/client.go:116-127` vs `:135-146`; `worker.go:192` | Outbox идёт через core `PublishMsg`+Flush, хотя есть `PublishMsgAck` (JetStream PubAck). Flush закрывает crash-окно буфера (P1-BE-24), но **не** гарантирует запись в stream EVENTS | InterestPolicy / stream lag: «published» в outbox при отсутствии durable message — soft residual исходного W4-A2-001 | Worker → `PublishMsgAck(ctx, msg)` для subject’ов stream |
| A2W4-2-003 | RESIDUAL | P3 | `client_test.go:15-27` | `TestStreamSubjects` — локальный слайс из **7** префиксов; prod `EnsureStreams` = 9 (`notification.>`, `calls.recording.>`) | Unit-тест не ловит удаление ME-07 subjects (integration покрывает notification, unit — нет) | Assert канон-список = subjects из `EnsureStreams` / export |

## FIXED-OK

| Stage A ID | Sev | Статус | Доказательство |
|------------|-----|--------|----------------|
| W4-A2-001 | P1 | FIXED-OK (core Flush) | `PublishMsg` → FlushTimeout до mark published; hubtest drop-hook |
| W4-A2-002 | P2 | FIXED-OK | `StatusFailed` + `markExhaustedFailed` |
| W4-A2-003 | P2 | FIXED-OK | SKIP LOCKED claim (NEW lock-hold = A2W4-2-001) |
| W4-A2-004 | P2 | FIXED-OK | Redis Ping fail → `Close()` |
| W4-A2-006 | P3 | FIXED-OK | `Close` → Drain + timeout |
| W4-A2-007 | P3 | FIXED-OK | rollback-тест (task059) |
| W4-A2-008 | P3 | FIXED-OK | package const maxAttempts/batch |
| W4-A2-009 | P3 | FIXED-OK | cancel mid-batch |

## Чисто

| Зона | Проверено |
|------|-----------|
| NATS redact | userinfo → `***:***@` |
| Trace | Inject/Extract без Baggage |
| Publish-in-tx | `PublishInTx` + TraceContextMap |
| Redis logs | URL не логируется |

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **15 / 15 / 0** |
| Активных | **3** (NEW 1 · REGRESS 0 · RESIDUAL 2) |
| FIXED-OK | **8** |
| P0 | 0 |
| P1 | 0 |
| P2 | 2 |
| P3 | 1 |
