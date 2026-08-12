# MUTATION W13 slot 3 refill — TASK-148 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/archive/2026-07-29/TASK-148-2026-07-29-a3-4-1-platform-http.md` `[MUTATION CRITERIA]`  
> Finding: **A3W4-1-001** (1 ID)  
> Коммит не создавался.

## Verdict

**1/1 PROVEN** · TASK-148 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| A3W4-1-001 | **PROVEN** | `TestRequestID_RejectsInvalid`; `TestRequestID_RejectsOversized` |

`layer_mismatch IDs: (none)` — Fiber middleware guard; Ent N/A.

## Per-ID evidence

### A3W4-1-001 — `validRequestID` reject invalid / oversized X-Request-ID

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/platform/middleware/ \
    -run "TestRequestID_RejectsInvalid|TestRequestID_RejectsOversized"
  -> PASS (~0.9s)

BREAK: validRequestID → return true
  (kill recipe: ослабить guard empty/maxLen/pattern)
  -> FAIL TestRequestID_RejectsInvalid/{spaces,angle_brackets,slash}
       UUID parse error + NotEqual client (~0.8s)
  -> FAIL TestRequestID_RejectsOversized/{ascii_129,rune_cyrillic_129}
       UUID parse error + NotEqual client
  (adjacent also RED: TestRequestID_RejectsOversizedHeader_W4A1002 — not criteria target)

RESTORE requestid.go → original validRequestID body
  -> PASS TestRequestID_* suite (~0.7s)

git diff backend/internal/platform/middleware/requestid.go → empty
MUTATED marker → отсутствует
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=120s ./testframework/platform/middleware/ \
  -run "TestRequestID_RejectsInvalid|TestRequestID_RejectsOversized"
# BASELINE PASS · BREAK FAIL · RESTORE PASS
```

## Gaps

- none — criteria ID PROVEN  
- CopyString mutation (Pitfall 1) — optional adjacent; вне A3W4-1-001 acceptance (N/A)

## Handoff

- Evidence: этот файл · TASK-148 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored; no leftover MUTATED; no commit
