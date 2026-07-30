# MUTATION W13 slot 1 refill — TASK-041 (be1) · 2026-07-30

> Checker: **be1** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/TASK-041-be-p2-sprint-poker-start.md` `[MUTATION CRITERIA]`  
> Коммит не создавался (инструкция слота).

## Verdict

**3/4 PROVEN** · **1 PROVEN + GAP_NOTED** (007 mock≠Ent) · leftover MUTATED: **нет**  
TASK-041 → **Закрыта**

| ID | Verdict | RED test |
|----|---------|----------|
| W1-A4-002 | **PROVEN** | `TestPoker_VoteAfterReveal_Rejected` |
| W1-A4-009 | **PROVEN** | `TestStartSprintAtomic_Ent_SecondActiveConflicts` |
| W1-A4-005 | **PROVEN** | `TestPoker_RBAC_StrangerForbidden` |
| W1-A4-007 | **PROVEN** + **GAP_NOTED** | `TestUpdateSprint_EndBeforeStart_BadRequest` |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: W1-A4-007` · `docs-only IDs: (none)`

## Per-ID evidence

### W1-A4-002 — CastVote reject Lua result `0` (revealed)

```text
BASELINE TestPoker_VoteAfterReveal_Rejected -> PASS (~0.9s)

BREAK: sprint/poker.go CastVote —
  case 0: drop NewBadRequest("голоса уже раскрыты")
-> FAIL p2_be11_poker_start_test.go:49
   An error is expected but got nil.
RESTORE -> PASS (final suite ~1.0s)
```

Guard: `backend/internal/sprint/poker.go:141-145` + Lua `revealed=='true'→0` — PokerService (miniredis).

### W1-A4-009 — StartSprintAtomic `ErrActiveSprintConflict`

```text
BASELINE TestStartSprintAtomic_Ent_SecondActiveConflicts -> PASS

BREAK: sprint/repository.go StartSprintAtomic —
  убрать return &ErrActiveSprintConflict{…}
-> FAIL p2_be11_poker_start_test.go:224
   ErrorAs *ErrActiveSprintConflict missing;
   chain: UNIQUE constraint failed: sprints.project_sprints
RESTORE -> PASS
```

Guard: `backend/internal/sprint/repository.go:252-264` — EntRepository (enttest SQLite).  
Note: without app-level conflict, SQLite unique still fires — named test still RED on ErrorAs (canonical kill).

### W1-A4-005 — authorizePokerSession stranger → 403

```text
BASELINE TestPoker_RBAC_StrangerForbidden -> PASS

BREAK: sprint/handler.go CastVote —
  skip authorizePokerSession call
-> FAIL p2_be11_poker_start_test.go:304
   expected: 403 actual: 204 (stranger vote must be 403)
RESTORE -> PASS
```

Guard: `backend/internal/sprint/handler.go:452-466` + CastVote call site — handler + Ent RBAC.

### W1-A4-007 — UpdateSprint `end.Before(start)` → BadRequest

```text
BASELINE TestUpdateSprint_EndBeforeStart_BadRequest -> PASS

BREAK: sprint/service.go UpdateSprint —
  убрать if end.Before(start) → NewBadRequest
-> FAIL p2_be11_poker_start_test.go:87
   An error is expected but got nil.
RESTORE -> PASS
```

Guard: `backend/internal/sprint/service.go:107-109` — SprintService.

**GAP_NOTED** mock≠Ent: kill на prod service path PROVEN через `mockRepository`; Ent `UpdateSprint` IT на date-order отсутствует (criteria flag).

## Commands

```powershell
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"
cd backend
go test -tags hubtest -count=1 -timeout=120s `
  -run "TestPoker_VoteAfterReveal_Rejected$|TestStartSprintAtomic_Ent_SecondActiveConflicts$|TestPoker_RBAC_StrangerForbidden$|TestUpdateSprint_EndBeforeStart_BadRequest$" `
  ./testframework/sprint/
# FINAL: ok … 0.986s
```

## Leftover check

`Select-String MUTATED` on `poker.go` / `repository.go` / `handler.go` / `service.go` → empty.
