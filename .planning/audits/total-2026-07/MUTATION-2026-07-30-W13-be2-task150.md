# MUTATION W13 slot 3 refill — TASK-150 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/archive/2026-07-29/TASK-150-2026-07-29-a3-4-3-platform-obs.md` `[MUTATION CRITERIA]`  
> Findings: **A3W4-3-001**, **A3W4-3-002** (2 IDs)  
> Коммит не создавался.

## Verdict

**2/2 PROVEN** · TASK-150 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| A3W4-3-001 | **PROVEN** | `TestTimed_CancelNotErrorStatus_A3W4_3_001`; `TestTimed_DeadlineNotErrorStatus_A3W4_3_001`; `TestEndConsumerSpan_CancelNotErrorStatus_A3W4_3_001` |
| A3W4-3-002 | **PROVEN** | `TestSanitizeOperationLabel_DomainUUIDToOther_A3W4_3_002`; `TestRecordOperation_DomainUUIDBucketsToOther_A3W4_3_002` |

`layer_mismatch IDs: (none)` — otel/timing/label guards; Ent N/A.

## Per-ID evidence

### A3W4-3-001 — cancel/deadline ≠ span Error / error_count

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/platform/observability/ \
    -run "TestTimed_CancelNotErrorStatus_A3W4_3_001|TestTimed_DeadlineNotErrorStatus_A3W4_3_001|TestEndConsumerSpan_CancelNotErrorStatus_A3W4_3_001|…"
  -> PASS (~1.0s)

BREAK:
  otel.go finishOperationSpan → if err != nil { recordRedactedSpanError… } (no isBenignContextError)
  timing.go → failed := err != nil (cancel counts as error; slog.Error path)
  -> FAIL TestTimed_CancelNotErrorStatus_A3W4_3_001
       Status=Error + ErrorCount=1 (~1.2s)
  -> FAIL TestTimed_DeadlineNotErrorStatus_A3W4_3_001
       Status=Error + ErrorCount=1
  -> FAIL TestEndConsumerSpan_CancelNotErrorStatus_A3W4_3_001
       Status=Error

RESTORE otel.go + timing.go → original carve-out / failed=!benign
  -> PASS (~1.0s)

git diff otel.go timing.go → empty
MUTATED marker → отсутствует
```

### A3W4-3-002 — sanitize domain+UUID → other

```text
BREAK: isValidOperationSegment — allow digit start + hyphen in rest
  (kill recipe: charset с `-` → tracker.<uuid> проходит)
  -> FAIL TestSanitizeOperationLabel_DomainUUIDToOther_A3W4_3_002
       expected "other", got "tracker.550e8400-…" / "auth.123e4567-…"
  -> FAIL TestRecordOperation_DomainUUIDBucketsToOther_A3W4_3_002
       Snapshot.Operation = "tracker.550e8400-…" ≠ "other"
  (001 suite still PASS after 001 restore)

RESTORE operation_label.go → [A-Za-z][A-Za-z0-9_]* only
  -> PASS UUID suite (~1.0s)

git diff operation_label.go → empty
MUTATED marker → отсутствует
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=120s ./testframework/platform/observability/ \
  -run "TestTimed_CancelNotErrorStatus_A3W4_3_001|TestTimed_DeadlineNotErrorStatus_A3W4_3_001|TestEndConsumerSpan_CancelNotErrorStatus_A3W4_3_001|TestSanitizeOperationLabel_DomainUUIDToOther_A3W4_3_002|TestRecordOperation_DomainUUIDBucketsToOther_A3W4_3_002"
# BASELINE PASS · BREAK FAIL (per ID) · RESTORE PASS
```

## Gaps

- none — criteria IDs PROVEN  
- `TestTimed_RealErrorStillErrorStatus_A3W4_3_001` — regression (не kill target); N/A  
- Prior W4-A3-008 bare-UUID (`operation_label_task065`) — complementary; N/A

## Handoff

- Evidence: этот файл · TASK-150 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored; no leftover MUTATED; no commit
