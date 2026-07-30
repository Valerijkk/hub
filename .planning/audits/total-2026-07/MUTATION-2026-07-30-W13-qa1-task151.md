# MUTATION W13 slot 4 refill — TASK-151 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** · Env: localhost `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` · Go 1.25  
> Criteria: `.planning/backlog/archive/2026-07-29/TASK-151-2026-07-29-a3-4-4-platform-core.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**2/2 PROVEN** · TASK-151 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A3W4-4-001 | **PROVEN** | `TestNewPostgresClient_UnreachableDSN_FailClosed_A3W44001` | drop `PingContext`+`drv.Close` (Open→`ent.NewClient`→slog) | database connect |
| A3W4-4-002 | **PROVEN** | `TestRunTick_TimeoutCancelsHang_A3W44002` | `WithTimeout` → `fn(ctx)` (`_ = tickTimeout`) | cron `runTick` |

Production restored (`git checkout --`) после каждого BREAK. Leftover MUTATED: **нет**.  
`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)`.

## Per-ID evidence

### A3W4-4-001 — Postgres Ping fail-closed

```text
BASELINE → PASS (~2.0s)

BREAK: postgres.go — remove PingContext / drv.Close / fmt import
→ FAIL: An error is expected but got nil (+ slog "connected to PostgreSQL")
RESTORE → PASS (~1.9s)
```

LiveDSN sibling `TestNewPostgresClient_LiveDSN_PingOK_A3W44001` PASS after restore (compose up).

### A3W4-4-002 — cron tick WithTimeout

```text
BASELINE → PASS (~1.7s)

BREAK: cron.go runTick — WithTimeout → fn(ctx)
→ FAIL: tick should have been cancelled by timeout (~2.0s hang)
RESTORE → PASS (~2.0s)
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -race -count=1 -timeout=120s ./testframework/platform/database/ \
  -run "TestNewPostgresClient_UnreachableDSN_FailClosed_A3W44001|TestNewPostgresClient_LiveDSN_PingOK_A3W44001"
go test -tags hubtest -race -count=1 -timeout=120s ./testframework/platform/cron/ \
  -run "TestRunTick_TimeoutCancelsHang_A3W44002"
```

## Handoff

- Evidence: этот файл · TASK-151 **Закрыта** · README sync  
- be-ver → code-reviewer (fix code already FIXED; mutation PROVEN)  
- Коммит: не создавался
