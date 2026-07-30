# MUTATION W13 slot 3 refill — TASK-189 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/TASK-189-2026-07-29-audit-be-v4-calendar-poll.md` `[MUTATION CRITERIA]`  
> Findings: **BE-V4-001**, **BE-V4-002**, **BE-V4-003** (3 IDs)  
> Коммит не создавался. Ent + service kills; leftover MUTATED: **нет**.

## Verdict

**3/3 PROVEN** · TASK-189 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| BE-V4-001 | **PROVEN** | `TestListApprovedLeaveInRange_ExcludesSoftDeleted_Task189` |
| BE-V4-002 | **PROVEN** | `TestListSprintsInRange_ExcludesSoftDeleted_Task189` |
| BE-V4-003 | **PROVEN** | `TestDraftNewsPoll_GetVoteHiddenFromNonOwner_Task189` |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)`

## Per-ID evidence

### BE-V4-001 — ListApprovedLeaveInRange `DeletedAtIsNil`

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/calendar/ \
    -run TestListApprovedLeaveInRange_ExcludesSoftDeleted_Task189
  -> PASS (~0.7s)

BREAK (EntRepository correct layer):
  backend/internal/calendar/repository.go ListApprovedLeaveInRange
  — убрать leaverequest.DeletedAtIsNil() из Where
  — MUTATED BE-V4-001

  -> FAIL TestListApprovedLeaveInRange_ExcludesSoftDeleted_Task189 (~0.9s)
       should have 1 item(s), but has 2
       (soft-deleted leave returned)

RESTORE DeletedAtIsNil
  -> PASS (~0.7s)

git diff repository.go → empty
MUTATED marker → отсутствует
```

Layer OK: Ent `LeaveRequest` + `calendar.NewEntRepository` (sqlite enttest).

### BE-V4-002 — ListSprintsInRange `DeletedAtIsNil`

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/calendar/ \
    -run TestListSprintsInRange_ExcludesSoftDeleted_Task189
  -> PASS (в общем Task189 calendar ~0.7s)

BREAK (EntRepository correct layer):
  backend/internal/calendar/repository.go ListSprintsInRange
  — убрать sprintEnt.DeletedAtIsNil() из Where
  — MUTATED BE-V4-002

  -> FAIL TestListSprintsInRange_ExcludesSoftDeleted_Task189 (~0.7s)
       should have 1 item(s), but has 2
       (soft-deleted sprint returned)

RESTORE DeletedAtIsNil
  -> PASS (~0.8s)

git diff repository.go → empty
MUTATED marker → отсутствует
```

Layer OK: Ent `Sprint` + EntRepository (sqlite enttest).

### BE-V4-003 — draft poll Get/Vote → ErrNotFound for non-author

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/intranet/ \
    -run TestDraftNewsPoll_GetVoteHiddenFromNonOwner_Task189
  -> PASS (~0.8s)

BREAK (service correct layer):
  backend/internal/intranet/service_polls.go requirePublishedNewsOrAuthorOrAdmin
  — всегда return nil (draft IDOR guard bypass)
  — MUTATED BE-V4-003

  -> FAIL TestDraftNewsPoll_GetVoteHiddenFromNonOwner_Task189 (~0.8s)
       Expected ErrNotFound but got nil
       (member must not read/vote draft poll)

RESTORE requirePublishedNewsOrAuthorOrAdmin body
  -> PASS Task189 intranet (~0.7s)

git diff service_polls.go → empty
MUTATED marker → отсутствует
```

Layer OK: Ent `NewsPost`/`NewsPoll` + intranet.Service authz.

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=120s ./testframework/calendar/ -run "Task189"
go test -tags hubtest -count=1 -timeout=120s ./testframework/intranet/ -run "Task189"
# BASELINE PASS · FINAL PASS (calendar ~0.6s · intranet ~0.7s)
```

## Gaps

- **N/A** — все 3 ID runtime Ent/service; layer_mismatch/mock≠Ent/docs-only: none.

## Handoff

- Evidence: этот файл · TASK-189 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored after each BREAK; `git diff` на `calendar/repository.go` + `intranet/service_polls.go` — empty · no leftover MUTATED
