# MUTATION W5 slot 3 — TASK-149 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** (!= be1 fixer) · Env: localhost · Evidence: `_mut-batch-w5-task149.json`  
> Criteria: `.planning/backlog/archive/2026-07-29/TASK-149-2026-07-29-a3-4-2-platform-nats-outbox.md` `[MUTATION CRITERIA]`  
> Runner: `_run_mutations_w5_task149.py`

## Verdict

**3/3 runnable PROVEN** · dual-worker SQLite = sequential post-commit (не true concurrent PG SKIP LOCKED), но lease-семантика ловится тестом — **не** fake-PROVEN · статус TASK-149 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A3W4-2-001 | **PROVEN** | `TestProcessEvent_PanicInPublish_NoCrash` | убрать `recordPublishFailure` из `recover` | worker (Ent outbox) |
| A3W4-2-002 | **PROVEN** (dual) | `TestClaimAndRelease_DualWorker_DisjointIDs` | Commit claim без `stampClaimLease` / `SetPublishedAt` | worker claim lease |
| A3W4-2-002 | **PROVEN** (stamp) | `TestStampClaimLease_LostRace_Skips` | `stampClaimLease` без `Where(claimAvailable)` + `SetPublishedAt` | worker claim lease |
| A3W4-2-003 | **PROVEN** | `TestPublishWithTraceContext_FlushAfterPublish_A3W4_2_003` | `PublishWithTraceContext` → `Conn.PublishMsg` без Flush | nats client |

Production restored (`git checkout --` после каждого BREAK). Mock-gap: нет (001/002 — Ent sqlite; 003 — embedded NATS + hook).

## Per-ID evidence

### A3W4-2-001 — panic → dead-letter path

```text
BREAK: worker.go recover — drop w.recordPublishFailure(...)
→ FAIL TestProcessEvent_PanicInPublish_NoCrash  (attempts stay 0 / no panic last_error)
RESTORE
→ PASS same test
```

### A3W4-2-002 — claim lease (published_at)

**Gap note (honest):** dual-worker на SQLite — sequential post-commit, не true concurrent `FOR UPDATE SKIP LOCKED`. Критерий волны: не fake-PROVEN если тест не ловит. Здесь тест **ловит** мутацию «без stamp»: второй sequential claim забирает те же ID → RED.

```text
BREAK dual: claimAndRelease — claimed := events (без stampClaimLease)
→ FAIL TestClaimAndRelease_DualWorker_DisjointIDs  (second not empty)
RESTORE → PASS

BREAK stamp: stampClaimLease — UpdateOneID.Save без Where/SetPublishedAt
→ FAIL TestStampClaimLease_LostRace_Skips  (ok2 stays true)
RESTORE → PASS
```

### A3W4-2-003 — Flush after PublishWithTraceContext

```text
BREAK: client.go PublishWithTraceContext → c.Conn.PublishMsg(msg) (без PublishMsg/Flush)
→ FAIL TestPublishWithTraceContext_FlushAfterPublish_A3W4_2_003
   (hookFired=false и/или err==nil)
RESTORE
→ PASS same test
```

## Commands

```text
# baseline + per-ID Break→RED→GREEN
python .planning/audits/total-2026-07/_run_mutations_w5_task149.py
→ BASELINE OK
→ A3W4-2-001 PROVEN · A3W4-2-002-dual PROVEN · A3W4-2-002-stamp PROVEN · A3W4-2-003 PROVEN
→ wall ~122s

# regression after restore
cd backend
go test -tags hubtest -race -count=1 -timeout=300s \
  ./testframework/platform/outbox/ ./testframework/platform/nats/
→ ok  outbox ~5.5s · nats ~7.8s
```

## Assumptions / unresolved

- Path 002 без enum `processing` — lease на `published_at` + TTL reclaim (как у be1). Schema migrate не требовался.
- True multi-replica concurrent claim на Postgres SKIP LOCKED — OOS для sqlite hubtest; покрыт post-commit lease + lost-race stamp.
- Коммит не создавался (инструкция слота).

## Handoff

TASK-149 **Закрыта** · evidence этот файл · README синхронизирован · next: be-ver / code-reviewer по желанию волны (mutation gate закрыт).
