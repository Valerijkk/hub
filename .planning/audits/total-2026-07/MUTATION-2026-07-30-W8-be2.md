# MUTATION W8 slot 1 — TASK-087 (be2) · 2026-07-30

> Checker: **be2** (!= be1 fixer) · Env: localhost · Go 1.25 · SQLite in-memory (`newTestEntClient`)  
> Criteria: `.planning/backlog/archive/2026-07-28/TASK-087-2026-07-28-outbox-claim-lease-flush.md` `[MUTATION CRITERIA]`  
> Коммит не создавался (инструкция слота).

## Verdict

**1/1 PROVEN** · статус TASK-087 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A2W4-2-001 | **PROVEN** | `TestProcessBatch_ClaimReleasedBeforePublish_TASK087` (+ concurrent FAIL) | hold-during-flush: publish внутри claim-tx **до** `Commit`; `afterClaimCommitHook` только после Flush | worker (Ent outbox claim) |

Production restored (`Copy-Item` backup → `worker.go`). Leftover MUTATED: **нет** (`git status` clean на `worker.go`).

## Per-ID evidence

### A2W4-2-001 — claim Commit before PublishMsg/Flush

```text
BASELINE (clean worker.go)
→ PASS TestProcessBatch_ClaimReleasedBeforePublish_TASK087 (0.29s)
→ PASS TestProcessBatch_ConcurrentWorkerDuringFlush_TASK087 (0.11s)

BREAK: processBatch — claimPending + stampClaimLease, затем processEvent
       (PublishMsg/Flush) ПОКА tx открыт; Commit + afterClaimCommitHook ПОСЛЕ Flush
→ FAIL TestProcessBatch_ClaimReleasedBeforePublish_TASK087
   PublishMsg/Flush must observe claim already committed (not hold-during-flush)
   (publishSawClaimCommitted=false; pending остался 1 — mark published blocked by SQLite lock)
→ FAIL TestProcessBatch_ConcurrentWorkerDuringFlush_TASK087
   database table is locked (insert B во время Flush при открытом claim-tx)

RESTORE
→ PASS ClaimReleasedBeforePublish (0.16s)
→ PASS ConcurrentWorkerDuringFlush (0.13s)
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"

# baseline + after restore
go test -tags hubtest -count=1 -timeout=120s ./testframework/platform/outbox/ \
  -run "TestProcessBatch_ClaimReleasedBeforePublish_TASK087|TestProcessBatch_ConcurrentWorkerDuringFlush_TASK087" -v
→ ok ~1.9s

# under MUTATED (same -run)
→ FAIL both (exit 1)

# related claim regression after restore
go test -tags hubtest -count=1 -timeout=180s ./testframework/platform/outbox/ \
  -run "TASK087|Claim|claim" -v
→ ok DualWorker + LostRace + TASK087×2 ~2.1s
```

## Assumptions / unresolved

- Concurrent RED проявился как `database table is locked` на insert в Flush-hook (сильнее timeout concurrent batch) — всё равно доказывает hold-during-flush / open claim-tx.
- Lease/`stampClaimLease` (A3W4-2-002 / TASK-149) **не** мутировался (per Gaps в TASK-087).
- Коммит не создавался.

## Handoff

TASK-087 **Закрыта** · evidence этот файл · README синхронизирован · be-ver → code-reviewer.
