# A3 W4 агент 2 — platform nats+outbox+redis (FINAL re-audit)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · ветка `cursor-grok-4.5`  
> Mirror: [A2W4-2-platform-nats-outbox.md](A2W4-2-platform-nats-outbox.md) · Assignment: [W4-ASSIGNMENTS.md](W4-ASSIGNMENTS.md) § A2  
> Периметр: 15 файлов (7 prod + 8 tests) · Прочитано: **15/15** · Пропущено: **0**  
> Plan: [A3-PLAN.md](A3-PLAN.md) · Remediation: TASK-087 (A2W4-2-001) · TASK-112 (A2W4-2-002/003)

> **Вне списка assignment (сверка FIXED-OK, не в таблице чтения):**  
> `claim_lease_task087_test.go`, `failed_claim_task043_test.go`, `publish_rollback_task059_test.go` — evidence TASK-087 / W4-A2-002/003/007.

## Сверка с A2 (A2W4-2-*)

| A2 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| A2W4-2-001 | FIXED-OK | закрыт | TASK-087: `claimAndRelease` Commit **до** publish (`worker.go:71–117`); mutation `claim_lease_task087_test.go` (вне слота) |
| A2W4-2-002 | FIXED-OK | закрыт | TASK-112: `PublishMsgAck` до mark published (`worker.go:205–213`, `client.go:149–163`); `TestProcessEvent_PubAckFailure_StaysPending` |
| A2W4-2-003 | FIXED-OK | закрыт | TASK-112: `StreamSubjects` канон (`client.go:19–31`) = `EnsureStreams`; `TestStreamSubjects` |

**REGRESS:** 0

### Cascade FIXED-OK (Stage A W4-A2-* / A2 FIXED-OK)

| Stage A / A2 | Evidence (файл:строка) |
|--------------|------------------------|
| W4-A2-001 Flush | `PublishMsg` → `FlushTimeout` (`client.go:134–142`); hubtest `afterCorePublishHook` |
| W4-A2-002 dead-letter | `StatusFailed` + `markExhaustedFailed` (`worker.go:120–137`, `228–250`) |
| W4-A2-003 SKIP LOCKED | `skipLockedPredicate` (`worker.go:164–168`); SQLite fallback |
| W4-A2-004 Redis Ping | Ping fail → `Close()` (`redis/client.go:18–20`) |
| W4-A2-006 Drain | `Close` Drain + 5s timeout + force (`client.go:85–103`) |
| W4-A2-007 rollback | evidence вне слота `publish_rollback_task059_test.go` |
| W4-A2-008 consts | `maxPublishAttempts`/`batchLimit`/`pollInterval` + export (`worker.go:19–24`) |
| W4-A2-009 cancel mid-batch | `ctx.Done` в цикле (`worker.go:80–86`); тест в слоте |

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/internal/platform/nats/client.go` | 238 | 1 | NEW — PublishWithTraceContext без Flush; остальное FIXED-OK |
| `backend/internal/platform/nats/export_hooks.go` | 21 | 0 | чисто — hubtest redact/hook/drain |
| `backend/internal/platform/nats/trace.go` | 71 | 0 | чисто — TraceContext без Baggage |
| `backend/internal/platform/outbox/export_hooks.go` | 74 | 0 | чисто — claim/hook exports |
| `backend/internal/platform/outbox/publisher.go` | 54 | 0 | чисто — PublishInTx + TraceContextMap |
| `backend/internal/platform/outbox/worker.go` | 250 | 2 | NEW — panic≠dead-letter; claim без processing |
| `backend/internal/platform/redis/client.go` | 24 | 0 | FIXED-OK Ping→Close; URL не в логе |
| `backend/testframework/platform/nats/client_integration_test.go` | 120 | 0 | чисто — Connect/Drain/EnsureStreams |
| `backend/testframework/platform/nats/client_publish_test.go` | 96 | 0 | чисто — PublishJSON / idempotent streams |
| `backend/testframework/platform/nats/client_test.go` | 39 | 0 | FIXED-OK StreamSubjects=9 |
| `backend/testframework/platform/nats/redact_url_test.go` | 19 | 0 | чисто — userinfo → `***:***@` |
| `backend/testframework/platform/outbox/outbox_test.go` | 338 | 1 | NEW — panic-тест закрепляет отсутствие attempts |
| `backend/testframework/platform/outbox/publisher_test.go` | 56 | 0 | чисто — Publish / PublishInTx |
| `backend/testframework/platform/outbox/trace_context_test.go` | 158 | 0 | чисто — REST→outbox→NATS traceparent |
| `backend/testframework/platform/outbox/worker_wave101_test.go` | 39 | 0 | чисто — attempts++ / closed client |

Файлов в списке: **15**. Прочитано: **15**. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|----------|-------------|---------|---------------|------------|
| A3W4-2-001 | NEW | P2 | `outbox/worker.go:185–193`; `outbox_test.go:163–169` | `recover` логирует panic и **не** вызывает `recordPublishFailure` — attempts остаются 0, status=pending | Poison-событие (стабильный panic) ретраится каждые 100 ms бесконечно; dead-letter (`attempts≥10`→failed) не срабатывает. Тест явно закрепляет «attempts не обновляются» | В `recover`: `recordPublishFailure(ctx, evt, fmt.Errorf("panic: %v", r))`; обновить panic-тесты: attempts≥1 / после 10 → failed |
| A3W4-2-002 | NEW | P3 | `outbox/worker.go:94–117`, `139–161` | `claimAndRelease` Commit’ит claim, **не** меняя status (остаётся `pending`) | После Commit SKIP LOCKED снят; второй `ProcessLoop`/реплика может выбрать те же строки и double-publish. At-least-once допускает дубликаты при crash, но здесь — систематический race при >1 worker. Wiring сейчас 1× loop — latent | Claim → `processing`+lease (или `UPDATE … WHERE status=pending RETURNING`), Commit; publish; published/failed. Тест dual-worker |
| A3W4-2-003 | NEW | P3 | `nats/client.go:168–172` (→ `:220–231`) | `PublishWithTraceContext` / `PublishJSONContext` делают `PublishMsg` **без** `FlushTimeout` | `PublishMsg` (P1-BE-24) ждёт Flush; trace-helper — нет. Crash сразу после Publish → сообщение только в клиентском буфере, caller видит success path у JSONContext только через лог ошибки Publish | После `PublishMsg` — `FlushTimeout(publishFlushTimeout)` (как в `PublishMsg`); либо явно запретить durable subjects и гнать через outbox+PubAck |

## FIXED-OK (сводка)

| ID | Evidence |
|----|----------|
| A2W4-2-001 | TASK-087: короткий claim Commit до publish; concurrent flush не блокирует |
| A2W4-2-002 | TASK-112: JetStream `PublishMsgAck` + empty-ack guard; fail → pending+attempts |
| A2W4-2-003 | `StreamSubjects` ↔ EnsureStreams; unit assert 9 subjects |
| W4-A2-001…009 | Flush, dead-letter, SKIP LOCKED, Redis Close, Drain, rollback, consts, cancel mid-batch — без регрессии |

## Чисто

- **JetStream:** `EVENTS` InterestPolicy, MaxAge 7d; subjects ME-07; `WebhookFilterSubjects` — filter durable consumer (не stream config).
- **Outbox at-least-once:** PubAck до `published`; fail mark → retry; mark-published fail → republish (дубликат OK).
- **Retry/poison (ошибочный путь):** `recordPublishFailure` → failed at 10; `markExhaustedFailed` sweep. Gap только panic → A3W4-2-001.
- **Trace:** Inject/Extract/Map без Baggage; `Hub-Event-Id`; publisher persist + worker forward.
- **Redis:** ParseURL + Ping; fail → Close; slog без URL/секретов.
- **Cancel/leak:** ProcessLoop на `ctx.Done`; mid-batch stop; NATS Close Drain+timeout; Redis Close on ping fail.
- **Тесты слота:** StreamSubjects; redact; PubAck failure; cancel mid-batch; trace headers; constants via export.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 1 |
| P3 | 2 |
| **Total** | **3** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 active → closed) | 3 (`A2W4-2-001`…`003`) |
| FIXED-OK (Stage A cascade) | 8 (`W4-A2-001`…`004`,`006`…`009`) |
| RESIDUAL | 0 |
| NEW | 3 (`A3W4-2-001`…`003`) |
| REGRESS | 0 |

**Топ IDs:** `A3W4-2-001` (panic≠dead-letter) · `A3W4-2-002` (claim без processing) · `A3W4-2-003` (trace-publish без Flush).

---

**Покрытие:** 15/15/0 · **counts:** P0=0 P1=0 P2=1 P3=2 Σopen=3 · **топ-3:** A3W4-2-001, A3W4-2-002, A3W4-2-003
