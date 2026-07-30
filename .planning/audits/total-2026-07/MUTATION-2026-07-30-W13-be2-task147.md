# MUTATION W13 slot 3 — TASK-147 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/TASK-147-2026-07-29-a3-3-5-calls.md` `[MUTATION CRITERIA]`  
> Driver: `.planning/audits/total-2026-07/_run_mutations_w13_task147.py`  
> Raw: `_mut-batch-w13-task147.json`  
> Коммит не создавался.

## Verdict

**4/4 PROVEN** · TASK-147 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| A3W3-A5-001 | **PROVEN** | `TestCreateCall_DBFail_DeletesLiveKitRoom_A3W3_A5_001` · `TestCreateInstantCall_DBFail_DeletesLiveKitRoom_A3W3_A5_001` |
| A3W3-A5-002 | **PROVEN** | `TestFindConflictingCalls_SkipsSoftDeletedChannel_A3W3_A5_002` |
| A3W3-A5-003 | **PROVEN** | `TestGetProjectIDByChannel_NilProject_NotFound_A3W3_A5_003` · `TestStartRecording_NilProjectID_ErrInvalidInput_A3W3_A5_003` · `TestHandler_StartRecording_NilProject_400_A3W3_A5_003` |
| A3W3-A5-004 | **PROVEN** | `TestCreateInstantCall_TitleOverMaxRunes_A3W3_A5_004` · `TestHandler_CreateInstantCall_TitleOverMax_400_A3W3_A5_004` · `TestGuestJoin_GuestNameOverMaxRunes_A3W3_A5_004` · `TestHandler_GuestJoin_GuestNameOverMax_400_A3W3_A5_004` |

`layer_mismatch IDs: (none)` — 001/004 service(+handler); 002 EntRepository; 003 Ent + service/handler.

## Per-ID evidence

### A3W3-A5-001 — DeleteRoom after DB-fail post-provisionRoom

```text
BASELINE CreateCall -> PASS (~6.8s)
BREAK: убрать deleteRoomBestEffort в CreateCall DB-fail path
-> FAIL deleteRoom count 0 (expected >=1) (~6.9s)
RESTORE -> PASS (~7.7s)

BASELINE CreateInstantCall -> PASS (~7.6s)
BREAK: убрать deleteRoomBestEffort в CreateInstantCall DB-fail path
-> FAIL deleteRoom count 0 (~7.1s)
RESTORE -> PASS (~6.7s)
```

### A3W3-A5-002 — FindConflictingCalls soft-delete filter

```text
BASELINE -> PASS (~6.8s)
BREAK: убрать HasChannelWith(channel.DeletedAtIsNil()) в FindConflictingCalls
-> FAIL soft-deleted channel снова в conflicts (~11.6s)
RESTORE -> PASS (~12.7s)
```

### A3W3-A5-003 — nil ProjectID → NotFound / ErrInvalidInput / 400

```text
BASELINE GetProjectID -> PASS (~7.2s)
BREAK: return (uuid.Nil, nil) вместо NotFound
-> FAIL expected error got nil (~6.9s)
RESTORE -> PASS (~7.3s)

BASELINE StartRecording svc -> PASS (~6.8s)
BREAK: silent Nil repo + снять NotFound/Nil→ErrInvalidInput map
-> FAIL (~1.1s)
RESTORE -> PASS (~6.4s)

BASELINE handler 400 -> PASS (~7.9s)
BREAK: same dual mutation
-> FAIL (~1.2s)
RESTORE -> PASS (~7.6s)
```

### A3W3-A5-004 — title / guest_name rune caps

```text
BASELINE title svc -> PASS (~7.2s)
BREAK: убрать title rune-cap в CreateInstantCall
-> FAIL ErrorIs ErrInvalidInput (~8.8s)
RESTORE -> PASS (~11.3s)

BASELINE title handler -> PASS (~12.2s)
BREAK: same title cap removal
-> FAIL 400 assert (~8.6s)
RESTORE -> PASS (~9.3s)

BASELINE guest svc -> PASS (~7.1s)
BREAK: убрать guest_name rune-cap в GuestJoin
-> FAIL ErrorIs (~9.1s)
RESTORE -> PASS (~8.7s)

BASELINE guest handler -> PASS (~6.1s)
BREAK: inflate handler threshold ×1000 + remove service cap
-> FAIL 400 assert (~9.4s)
RESTORE -> PASS (~6.3s)
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=300s ./testframework/calls/ -run "A3W3_A5"
# BASELINE PASS ~7.7s · FINAL PASS ~7.4s

python .planning/audits/total-2026-07/_run_mutations_w13_task147.py
# exit 0 · 4/4 PROVEN · leftover no
```

## Gaps

- none runnable — all 4 criteria PROVEN  
- 001 injects DB fail via `ExportCallServiceSetCreateCall` hook (criteria note) — OK  
- Title Ent `MaxLen(255)` = bytes; IT ASCII — residual non-ASCII outside criteria (noted in TASK)

## Handoff

- Evidence: этот файл · TASK-147 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored after each BREAK; `git diff` на calls prod files — empty; no `MUTATED` leftover
