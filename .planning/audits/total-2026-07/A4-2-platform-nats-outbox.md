# W4 агент A2 — platform nats+outbox+redis

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: JetStream ack/nak (producer path), outbox worker leaks, publish-in-tx, redis URL secrets in logs, trace propagation.  
> Сверка с `grok-2026-07`: A1-012 (SKIP LOCKED) — residual; A14-003 (NATS URL redact) — закрыт в `nats/client.go`; consumer Ack/Nak живут вне списка (notification/audit/tracker/apigateway).

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/platform/nats/client.go` | 170 | 2 | замечания |
| `backend/internal/platform/nats/export_hooks.go` | 8 | 0 | чисто |
| `backend/internal/platform/nats/trace.go` | 74 | 0 | чисто |
| `backend/internal/platform/outbox/export_hooks.go` | 30 | 0 | чисто |
| `backend/internal/platform/outbox/publisher.go` | 54 | 0 | чисто |
| `backend/internal/platform/outbox/worker.go` | 119 | 3 | замечания |
| `backend/internal/platform/redis/client.go` | 22 | 1 | замечания |
| `backend/testframework/platform/nats/client_integration_test.go` | 107 | 0 | чисто |
| `backend/testframework/platform/nats/client_publish_test.go` | 96 | 0 | чисто |
| `backend/testframework/platform/nats/client_test.go` | 33 | 1 | замечания |
| `backend/testframework/platform/nats/redact_url_test.go` | 19 | 0 | чисто |
| `backend/testframework/platform/outbox/outbox_test.go` | 287 | 1 | замечания |
| `backend/testframework/platform/outbox/publisher_test.go` | 56 | 1 | замечания |
| `backend/testframework/platform/outbox/trace_context_test.go` | 158 | 0 | чисто |
| `backend/testframework/platform/outbox/worker_wave101_test.go` | 39 | 0 | чисто |

Файлов в списке: 15. Прочитано: 15. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W4-A2-001 | P1 | `worker.go:93-118`; `nats/client.go:81-85` | Outbox после `PublishMsg` сразу ставит `status=published`, без `Flush`/`FlushTimeout` и без JetStream `PubAck` | `nats.Conn.PublishMsg` кладёт в клиентский буфер и может вернуть `nil` до подтверждения сервером/стримом. Crash/сетевой разрыв между publish и flush → событие в БД «published», в NATS/EVENTS его нет (потеря notification/webhook/audit). Durable JetStream path требует `js.Publish` + PubAck | В `PublishMsg` (или только в worker): `FlushTimeout` после publish **до** UpdateOne; лучше `JetStream.PublishMsg`/`Publish` и ждать PubAck; тест: mock/conn drop до flush → status остаётся pending |
| W4-A2-002 | P2 | `worker.go:15-16`; `worker.go:52-59`; `worker.go:93-105` | Godoc: после 10 попыток события «помечаются как failed»; схема имеет enum `failed`, но worker **нигде** не ставит `StatusFailed` — только `AttemptsLT(10)` исключает из poll | Dead-letter остаётся `pending` навсегда. Ops/алерты по `status=failed` пустые; расхождение схема↔код↔док | При `attempts+1 >= 10` (или отдельным sweep): `SetStatus(failed)` + `SetLastError`; метрика/лог dead-letter; тест на переход pending→failed |
| W4-A2-003 | P2 | `worker.go:52-59` | Выборка pending без claim (`FOR UPDATE SKIP LOCKED` / статус `processing`) — residual **A1-012** | ≥2 реплики/горутины: один батч публикуется дважды. At-least-once без дедупа на consumer → дубли WS/webhook/notification | Tx + `SkipLocked` (или атомарный `pending→processing` с lease); либо single-worker lock в Redis |
| W4-A2-004 | P2 | `redis/client.go:16-19` | При ошибке `Ping` клиент создан (`redis.NewClient`), но не `Close()` — return `nil, err` | Утечка соединений/горутин go-redis на каждый неудачный старт/реконнект | `defer`/`Close()` перед return err; тест: Ping fail → нет висящего клиента (или явный Close в defer) |
| W4-A2-005 | P2 | `client_test.go:15-27` vs `nats/client.go:111-120` | `TestStreamSubjects` проверяет **локальный** слайс из 7 префиксов, не `EnsureStreams`. В коде стрима ещё `notification.>`, `calls.recording.>` (9 subjects) | Тест зелёный при удалении ME-07 subjects из prod; drift не ловится. Assert-free относительно продукта | Читать subjects через `EnsureStreams`+`Stream.Info` (как integration) или экспортировать/сверять канон-список с `client.go` |
| W4-A2-006 | P3 | `nats/client.go:70-73` | `Close()` без `Drain`/`Flush` перед закрытием соединения | In-flight core publishes (в т.ч. из outbox без flush) могут обрезаться на shutdown | `Conn.Drain()` с timeout, затем Close; согласовать с wiring cleanup order (A8-018) |
| W4-A2-007 | P3 | `publisher_test.go:33-48` | Есть Commit-path для `PublishInTx`, нет Rollback: событие не должно остаться после `tx.Rollback()` | Регресс «Save вне tx / wrong client» не покраснеет | После `PublishInTx` → `Rollback` → `Count==0` |
| W4-A2-008 | P3 | `outbox_test.go:100-117`; `outbox_test.go:219-239` | `TestWorkerConstants` — константы только в тесте; `TestProcessEvent_SuccessIsTracked` по имени happy-path, по факту снова nil-NATS panic | Мутация `Limit(50)`/`AttemptsLT(10)`/интервала в worker не краснеет; имя теста вводит в заблуждение | Вынести maxAttempts/batch/poll в package const и ассертить их; переименовать/добавить реальный success (уже есть `TestProcessEvent_PublishSuccess_MarksPublished`) |
| W4-A2-009 | P3 | `worker.go:51-67` | `processBatch` не смотрит `ctx.Done()` между событиями батча (до 50) | После cancel loop всё равно дожимает батч (ок для at-least-once), но удлиняет shutdown и бьёт в уже закрывающийся DB (смежно A8-018) | Между итерациями: `select { case <-ctx.Done(): return; default: }`; wiring: WaitGroup до `db.Close` |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| `nats/client.go` redact | `redactNATSURL` на connect-логе; userinfo → `***:***@host` (A14-003). `JetStream.New` fail → `nc.Close()`. `MaxReconnects(-1)` |
| `nats/client.go` JetStream stream | `EnsureStreams`: EVENTS, InterestPolicy, MaxAge 7d, ME-07 subjects включая `notification.>`, `calls.recording.>`; `calls.chat.>` намеренно вне stream |
| JetStream ack/nak (scope) | **Consumer** Ack/Nak в этом списке файлов нет — в `notification/audit/tracker` consumers. Producer outbox идёт через **core** `PublishMsg`, не через JS consumer API |
| `nats/trace.go` | Явный W3C propagator (не global no-op); Inject/Extract headers; `TraceContextMap` / FromMap; `HeaderEventID` |
| `outbox/publisher.go` publish-in-tx | `PublishInTx` пишет в `tx.OutboxEvent` + `TraceContextMap(ctx)`; `Publish` — автономный Create. Marshal error пробрасывается. Атомарность с доменной tx — на вызывающем (паттерн верный) |
| Trace REST→outbox→NATS | Worker: ExtractFromMap → Inject headers + Hub-Event-Id; покрыто `trace_context_test.go` (store / empty / in_tx / header round-trip) |
| Redis URL secrets | `redis/client.go:20` логирует только `"connected to Redis"` **без** URL — утечки credentials в slog нет |
| `export_hooks` (nats/outbox) | hubtest-only re-export; без prod-логики |
| `redact_url_test.go` | Credentials URL → redact; plain URL без изменений |
| `client_integration_test` / `client_publish_test` | Connect/Close, Pub/Sub, EnsureStreams (+idempotent), PublishJSON success/bad/nil — asserts есть; Flush в тестах есть (в prod worker — нет → W4-A2-001) |
| `worker_wave101_test.go` | Publish fail → attempts++ / last_error; closed ent client → processBatch no panic |
| `outbox_test.go` HI-09 | Panic recover на nil nats; batch продолжает; ProcessLoop stops on cancel; live NATS → StatusPublished |

**Publish-in-tx вердикт:** реализация корректна; пробел — отсутствие rollback-теста (W4-A2-007).

**Trace propagation вердикт:** закрыт end-to-end в platform-слое (persist + NATS headers + extract).

**Redis secrets вердикт:** URL в логи не попадает (в отличие от NATS, где нужен был redact).

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **15 / 15 / 0** |
| Находок | **9** |
| P0 | 0 |
| P1 | 1 |
| P2 | 4 |
| P3 | 4 |

### Топ-3

1. **W4-A2-001 (P1)** — outbox помечает `published` без NATS Flush / JetStream PubAck → окно потери durable-событий.
2. **W4-A2-002 (P2)** — `StatusFailed` в схеме/godoc не выставляется; dead-letter навечно `pending`.
3. **W4-A2-003 (P2)** — нет claim/SKIP LOCKED (residual A1-012) → дубли при горизонтальном scale.
