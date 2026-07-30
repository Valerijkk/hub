# MUTATION W13 slot 3 refill — TASK-143 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/TASK-143-2026-07-29-a3-3-1-messenger-handlers.md` `[MUTATION CRITERIA]`  
> Коммит не создавался.

## Verdict

**1/1 PROVEN** · TASK-143 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| A3W3-1-001 | **PROVEN** | `TestHandler_GetPollResults_InvalidPollID` |

`layer_mismatch IDs: (none)` — guard в handler `parsePollID`; Ent N/A.

## Per-ID evidence

### A3W3-1-001 — parsePollID → 400 on bad UUID

```text
BASELINE → PASS TestHandler_GetPollResults_InvalidPollID (~0.9s)

BREAK (A): parsePollID accept-any — skip uuid.Parse, return uuid.New(), nil
→ FAIL expected 400 actual 500 (assert StatusBadRequest RED)
  handler_wave114_test.go:47

RESTORE → PASS (~1.0s)
git diff backend/internal/messenger/handler_reactions.go → empty
Select-String MUTATED → none
```

Kill recipe (A) applied on real prod guard `handler_reactions.go:275-284`.  
Recipe (B) (legacy wrong path) not needed — (A) alone proves the named test hits `parsePollID`.

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=120s ./testframework/messenger/ -run "TestHandler_GetPollResults_InvalidPollID"
# BASELINE PASS · MUTATED FAIL (400≠500) · RESTORE PASS
```

## Gaps

- none — criteria ID PROVEN

## Handoff

- Evidence: этот файл · TASK-143 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored; no leftover MUTATED
