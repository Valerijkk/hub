# MUTATION W13 slot 3 refill — TASK-146 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/TASK-146-2026-07-29-a3-3-4-ws.md` `[MUTATION CRITERIA]`  
> Finding: **A3W3-4-001** (1 ID)  
> Коммит не создавался.

## Verdict

**1/1 PROVEN** · TASK-146 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED |
|----|---------|-----|
| A3W3-4-001 | **PROVEN** | `TestWSHandler_QueryToken_WithValidTicket_Denied` |

`layer_mismatch IDs: (none)` — Fiber WS handler guard; Ent N/A.

## Per-ID evidence

### A3W3-4-001 — query-guard `?token=` before ticket parse

```text
BASELINE
  go test -tags hubtest -count=1 -timeout=120s ./testframework/ws/ \
    -run "TestWSHandler_QueryToken_WithValidTicket_Denied"
  -> PASS (~1.2s)

BREAK: закомментировать handler.go query-guard (ticket-required оставлен)
  // if utils.CopyString(c.Query("token")) != "" { return 401 "token query…" }

  go test … -run "TestWSHandler_QueryToken_WithValidTicket_Denied|TestWSHandler_QueryToken_Denied" -v
  -> FAIL TestWSHandler_QueryToken_WithValidTicket_Denied (~1.0s)
     expected 401, actual 500 (upgrade path без WS Connection header)
     body: "websocket: the client is not using the websocket protocol…"
     does not contain "token query parameter is not allowed"
  -> PASS TestWSHandler_QueryToken_Denied (false-green: 401 missing ticket)

RESTORE query-guard
  go test … -run "TestWSHandler_QueryToken"
  -> PASS both (~1.1s)

REGRESSION
  go test -tags hubtest -count=1 -timeout=300s ./testframework/ws/
  -> PASS (~2.4s)
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=120s ./testframework/ws/ -run "TestWSHandler_QueryToken"
# BASELINE/FINAL PASS

go test -tags hubtest -count=1 -timeout=300s ./testframework/ws/
# domain PASS ~2.4s
```

## Gaps

- none runnable — criteria PROVEN  
- Query-only `TestWSHandler_QueryToken_Denied` остаётся A2 regression; при этой мутации зелёный (как предсказано criteria)

## Handoff

- Evidence: этот файл · TASK-146 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored; `git diff` на `internal/ws/handler.go` — empty; no `MUTATED` leftover
