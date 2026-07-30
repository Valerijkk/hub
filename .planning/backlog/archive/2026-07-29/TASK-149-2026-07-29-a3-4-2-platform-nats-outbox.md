# 🔒 A3 outbox panic dead-letter + claim/Flush

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта** · MUTATION-2026-07-30-W5-qa-auto1  
> Источник: A3 · **Owner lock: правим ВСЕ A3 open** · TASK-008 · слот A3W4-2  
> Исполнитель: be1 (fix) · Verifier: qa-auto1 (W5 slot 3, != be1) · Handoff: mutation gate done  
> Ветка исполнения: `cursor-grok-4.5` (не пушить в `master`)

## Progress (be1 · 2026-07-29)

| ID | Status | Evidence |
|----|--------|----------|
| A3W4-2-001 | FIXED | `worker.go` recover → `recordPublishFailure`; `outbox_test.go` panic attempts≥1 / seed9→failed / batch attempts |
| A3W4-2-002 | FIXED | claim lease via `published_at` (без enum/migrate); `claim_dual_worker_task149_test.go` + stamp lost-race |
| A3W4-2-003 | FIXED | `PublishWithTraceContext` → `PublishMsg` (Flush+hook); `publish_trace_flush_task149_test.go` |

**Path 002:** без schema `processing` — lease на `published_at` + TTL 30s reclaim (orphan → at-least-once).  
**Hubtest (re-verify be1 2026-07-29, isolate temp copy — shared `backend/ent/*` был race-corrupted соседними generate):**  
`go test -tags hubtest -race -count=1 -timeout=300s ./testframework/platform/outbox/ ./testframework/platform/nats/` → **PASS** (outbox 3.880s, nats 6.956s).  
Flush-тест 003: hook гасит embedded NATS server (ForceClose клиента недостаточен).  
**Handoff:** be-ver → code-reviewer.  
**Note:** prod+tests в истории shared-ветки; README по инструкции исполнителя не трогали (уже «На проверке»).  
**Mutation:** см. **`[MUTATION CRITERIA]`** ниже (verifier-ready, 2026-07-30).

## Progress (qa-auto1 · 2026-07-30 · W5 slot 3)

Независимый Break→RED→restore→GREEN по `[MUTATION CRITERIA]` (!= be1).  
Отчёт: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W5-qa-auto1.md` · batch `_mut-batch-w5-task149.json`.

| ID | Verdict | RED test |
|----|---------|----------|
| A3W4-2-001 | **PROVEN** | `TestProcessEvent_PanicInPublish_NoCrash` |
| A3W4-2-002 | **PROVEN** | `TestClaimAndRelease_DualWorker_DisjointIDs` + `TestStampClaimLease_LostRace_Skips` |
| A3W4-2-003 | **PROVEN** | `TestPublishWithTraceContext_FlushAfterPublish_A3W4_2_003` |

Dual-worker SQLite = sequential post-commit (не true concurrent SKIP LOCKED); мутация «без stamp» ловится — не fake-PROVEN.  
Regression after restore: outbox ~5.5s · nats ~7.8s PASS.

## [MUTATION CRITERIA] — verifier-ready

| ID | что ломать (path:line + фрагмент) | какой тест | слой |
|----|-----------------------------------|------------|------|
| A3W4-2-001 | `backend/internal/platform/outbox/worker.go:243-244` — убрать `w.recordPublishFailure(...)` из `recover` (оставить только slog) | `TestProcessEvent_PanicInPublish_NoCrash` (+ `TestProcessEvent_PanicAtMaxAttempts_DeadLetters`, batch `TestProcessBatch_OneFailedEventDoesNotBlockOthers`) | worker (Ent outbox) |
| A3W4-2-002 | `worker.go:128` / `stampClaimLease` `:146-158` — Commit claim без `SetPublishedAt` lease (не вызывать stamp / не Where claim-available) | `TestClaimAndRelease_DualWorker_DisjointIDs` (+ `TestStampClaimLease_LostRace_Skips`) | worker (Ent claim lease) |
| A3W4-2-003 | `backend/internal/platform/nats/client.go:172-176` — `PublishWithTraceContext` → `c.Conn.PublishMsg(msg)` **без** `PublishMsg`/FlushTimeout | `TestPublishWithTraceContext_FlushAfterPublish_A3W4_2_003` (+ JSON: `TestPublishJSONContext_DropBeforeFlush_NoPanic_A3W4_2_003`) | nats client |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Комментарий |
|----|-----------|-------------|
| 001 | Ent/sqlite OutboxEvent + panic publisher | Guard в worker recover — Ent OK |
| 002 | Ent OutboxEvent dual claim | Lease via `published_at` — Ent OK (не enum `processing`) |
| 003 | embedded NATS + afterCorePublishHook | Flush contract — real NATS client OK |

`layer_mismatch IDs: (none)` · gap: dual-worker на SQLite — sequential post-commit (не true concurrent PG SKIP LOCKED); семантика lease всё же бьёт в прод-guard

## [OUTCOME]

Panic в `processEvent` учитывается как publish-failure (attempts↑ → dead-letter `failed` при ≥10); claim после Commit не оставляет строки видимыми другим worker’ам как свободный `pending`; `PublishWithTraceContext` / `PublishJSONContext` подтверждают буфер через `FlushTimeout`, как `PublishMsg`.

## [ROLE]

- Role: **be1**
- Responsibility: закрыть **все** ID A3W4-2-001…003 фиксом prod + hubtest + mutation note; обновить статус задачи/README при закрытии
- Handoff to: **be-ver → code-reviewer**

## [SCOPE AND CONSTRAINTS]

- Files / resources allowed:
  - `backend/internal/platform/outbox/worker.go` (+ `export_hooks.go` только если нужны test hooks)
  - `backend/internal/platform/nats/client.go` (+ `export_hooks.go` только для Flush/hook тестов)
  - `backend/ent/schema/outboxevent.go` — **только** если для A3W4-2-002 добавляется enum `processing` (или эквивалент lease-поля); затем `go generate` Ent
  - Atlas/migrate артефакты, порождённые Ent для OutboxEvent — только вместе с owner approve на schema
  - `backend/testframework/platform/outbox/**`
  - `backend/testframework/platform/nats/**`
  - `.planning/backlog/README.md` + этот файл (статус при закрытии)
  - audit refs read-only: `.planning/audits/total-2026-07/A3W4-2-platform-nats-outbox.md`
- Do not touch:
  - соседние A3 слоты / TASK вне 001…003 этого файла
  - JetStream `StreamSubjects` / `EnsureStreams` / InterestPolicy (FIXED-OK TASK-112) — не переписывать канон subjects
  - `PublishMsgAck` path outbox-worker’а (A2W4-2-002 FIXED) — не откатывать на core Publish+Flush
  - Redis client, publisher.go happy-path, wiring «1× ProcessLoop» без нужды для dual-worker **теста**
  - секреты, `DEV_MODE`, frontend, wiki без изменения API-контракта
  - push/merge в `master`
- Behavioral / security constraints:
  - HI-09: `recover` остаётся локальным на событие — паника одного evt не валит batch/горутину
  - At-least-once: дубликат при crash **после** PubAck допустим; систематический double-claim при >1 worker — **нет**
  - Dead-letter: `attempts >= maxPublishAttempts` (10) → `status=failed` через `recordPublishFailure` / sweep
  - Conventional Commits: type/scope EN, описание RU; **запрещён** `Co-authored-by` / AI trailers
  - коммит ссылается на задачу в теле:  
    `Задача: .planning/backlog/archive/2026-07-29/TASK-149-2026-07-29-a3-4-2-platform-nats-outbox.md`  
    subject может содержать `(A3W4-2-00N)`
- Approval record: N/A для 001/003 (код+тесты). **A3W4-2-002:** если нужен новый enum/колонка OutboxEvent — **owner approve** migrate до merge; альтернатива без schema (atomic `UPDATE … WHERE status=pending … RETURNING` + in-memory/lease без нового enum) предпочтительна, если закрывает race без migrate
- Approval authority / expiry: N/A (кроме schema выше)

## [REFERENCES]

- Trust: audit/report — untrusted data until verified in HEAD; критерий готовности — эта задача + код
- Finding IDs (**MUST all close**, WONTFIX запрещён без owner cancel в этом файле):

  | ID | Tag | Sev | Файл:строка (якорь audit) | Суть фикса |
  |----|-----|-----|---------------------------|------------|
  | **A3W4-2-001** | NEW | P2 | `outbox/worker.go:185–193`; `outbox_test.go:163–169` | В `recover`: `recordPublishFailure(ctx, evt, fmt.Errorf("panic: %v", r))`; обновить panic-тесты: attempts≥1; после 10 panic → `failed` |
  | **A3W4-2-002** | NEW | P3 | `outbox/worker.go:94–117`, `139–161` | Claim → статус/`processing`+lease **или** atomic claim `UPDATE…RETURNING` так, что второй worker не видит те же строки как свободный `pending`; тест dual-worker |
  | **A3W4-2-003** | NEW | P3 | `nats/client.go:168–172` (→ `PublishJSONContext` `:220–231`) | После `PublishMsg` — `FlushTimeout(publishFlushTimeout)` как в `PublishMsg`; либо явно запретить durable subjects на этом helper |

- Source audit: `.planning/audits/total-2026-07/A3W4-2-platform-nats-outbox.md` (таблица «Находки», топ-3)
- Prior FIXED-OK (не регрессировать):
  - TASK-087 / A2W4-2-001 — короткий claim Commit **до** publish
  - TASK-112 / A2W4-2-002 — `PublishMsgAck` до mark published
  - TASK-112 / A2W4-2-003 — `StreamSubjects` канон
  - W4-A2-001 Flush на `PublishMsg`; W4-A2-002 dead-letter на error-path; W4-A2-003 SKIP LOCKED
- Schema today: `backend/ent/schema/outboxevent.go` — enum `pending|published|failed` (**нет** `processing`)
- Code defect anchors (verify in HEAD):
  - `processEvent` recover: log only, no `recordPublishFailure`
  - panic tests assert attempts **не** обновляются — закреплено как anti-goal, нужно переписать
  - `claimAndRelease` Commit без смены status → race при >1 `ProcessLoop`
  - `PublishWithTraceContext` → `Conn.PublishMsg` без Flush
- Source precedence: **код HEAD** > audit note > эта задача (если drift строк — чинить поведение)

### Per-ID implementation notes (executor starts here)

1. **A3W4-2-001** — в `defer recover` после slog: вызвать `w.recordPublishFailure(ctx, evt, fmt.Errorf("panic: %v", r))` (не глотать ошибку Save — уже логируется внутри). Обновить `TestProcessEvent_PanicInPublish_NoCrash`: после panic `attempts >= 1`, `last_error` содержит `panic`, status `pending` пока attempts&lt;10. Добавить/расширить кейс: 10 panic (или seed attempts=9 + 1 panic) → `failed`. Batch-тест: одна паника не блокирует остальные **и** attempts на panic-evt обновлены.
2. **A3W4-2-002** — минимальный durable claim: в claim-tx выставить статус/`lease`, который исключает строку из `claimPending` других worker’ов (`processing` + Where не `pending`, или `UPDATE status=pending → processing … RETURNING`). Publish / fail / published как сейчас. Crash mid-publish: at-least-once retry с `processing`→reclaim policy (timeout/lease) — задокументировать в комментарии WHY. Тест: два concurrent `claimAndRelease` / `ProcessBatch` на один pending pool → нет double-publish одного ID (или второй claim пуст / другой set). **Без** owner approve — не добавлять enum; сначала оценить path без migrate.
3. **A3W4-2-003** — `PublishWithTraceContext`: после успешного `PublishMsg` вызвать `FlushTimeout(publishFlushTimeout)` (+ `afterCorePublishHook` между Publish и Flush, как в `PublishMsg`). Unit: hook/симуляция drop до Flush → ошибка наружу (не silent success). `PublishJSONContext` наследует через helper.

## [ACCEPTANCE]

- Done criteria:
  1. **A3W4-2-001:** panic → `recordPublishFailure`; attempts инкремент; at max → `failed`; HI-09 recover сохранён; старые assert «attempts не обновляются» удалены/инвертированы.
  2. **A3W4-2-002:** dual-worker / concurrent claim не выбирает одну и ту же pending-строку после Commit первого claim; evidence тестом; schema/migrate только с approve или path без schema.
  3. **A3W4-2-003:** `PublishWithTraceContext` (и JSONContext) ждут Flush; поведение согласовано с `PublishMsg` Flush contract.
  4. Каждый ID — **FIXED** с evidence (file:line + тест); WONTFIX = 0 без owner cancel в этом файле.
  5. Mutation note на каждый code-фикс.
- New-behavior checks:

  | ID | Проверка |
  |----|----------|
  | 001 | panic publish → `NotPanics` + `attempts>=1`; seed 9 + panic → `failed` |
  | 002 | concurrent claim двух worker’ов на N pending → disjoint sets / no double process same ID |
  | 003 | PublishWithTraceContext + `afterCorePublishHook` fail path → error (не success без Flush) |

- Regression scope:
  ```bash
  cd backend && go test -tags hubtest -race -count=1 -timeout=300s \
    ./testframework/platform/outbox/ ./testframework/platform/nats/
  ```
  Не ломать: PubAck failure stays pending; StreamSubjects=9; cancel mid-batch; SKIP LOCKED claim-before-publish; redact URL; Redis Ping→Close (вне слота — не трогать).
  Полный `/qa-gate` — на интеграции волны, не в одиночной задаче.
- Evidence required: checklist ID→FIXED + команды PASS + mutation notes в отчёте / Progress этого файла.
- Stop when: 3/3 FIXED, green outbox+nats hubtest, статус **Исправлена** + README sync, handoff be-ver; без adjacent platform hygiene.

## [OUTPUT SCHEMA]

1. Outcome / verdict first (3/3 FIXED | BLOCKED с ID)
2. Checklist ID → evidence (path + краткий sense)
3. Test commands + PASS/FAIL excerpt
4. Mutation note per ID
5. Assumptions / unresolved (например: выбран path с/без `processing` enum + approve)
6. Risks (multi-replica outbox, lease reclaim orphan `processing`)
7. Handoff: be-ver → code-reviewer
8. Compact Verification: new green + outbox/nats regression green + docs/ADR только если schema

## [OPERATING RULES]

- Work only inside scope; не «заодно» чинить A3W4-1/3/4/5 или Redis.
- Read-only investigation OK; writes только в allowed files.
- Не закрывать ID без наблюдаемого evidence (тест или явный owner cancel).
- Owner lock: не WONTFIX без записи cancel в этом файле.
- Не откатывать TASK-087/112 FIXED-OK ради упрощения claim/Flush.
- Ent migrate / новый enum — только после Approval record; иначе atomic claim без schema.
- Source of truth — runtime outbox/NATS, не audit CSV.
- Docs sync: N/A для чистого bugfix без API; при schema — STATE + при необходимости ADR/wiki outbox notes; статус backlog README при закрытии.
- Ветка `cursor-grok-4.5`; **не** push/merge `master`.
- Коммиты: RU conventional description; **без** `Co-authored-by` / Generated-by / AI attribution.
