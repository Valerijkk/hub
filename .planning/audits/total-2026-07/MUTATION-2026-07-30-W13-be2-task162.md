# MUTATION W13 slot 3 refill — TASK-162 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/archive/2026-07-29/TASK-162-2026-07-29-a3-7-1-notification-mail.md` `[MUTATION CRITERIA]`  
> Findings: **A3W7-1-001**, **A3W7-1-002** (2 IDs)  
> Коммит не создавался.

## Verdict

**2/2 PROVEN** · TASK-162 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| A3W7-1-001 | **PROVEN** | `TestService_SendAlert_SharedDailyLimit_A3W7_1_001` |
| A3W7-1-002 | **PROVEN** | `TestPushSender_QueueOverflow_A3W7_1_002` |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)`

## Per-ID evidence

### A3W7-1-001 — Redis `DailySlotStore` wiring (`SetDailySlotStore`)

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/mail/sharedlimit/ \
    -run TestService_SendAlert_SharedDailyLimit_A3W7_1_001
  -> PASS (~0.59s)

BREAK:
  mail/service.go SetDailySlotStore — noop (`_ = slots`; `// s.slots = slots`)
  -> FAIL TestService_SendAlert_SharedDailyLimit_A3W7_1_001
       3-й/4-й/restart send: Expected ErrDailyLimitExceeded, got nil
       (process-local MemoryDailySlotStore per Service — shared Redis не wired)

RESTORE s.slots = slots
  -> PASS (~0.45s)

git diff mail/service.go → empty
MUTATED marker → отсутствует
```

### A3W7-1-002 — Push queue cap 200 + drop

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/notification/pushoverflow/ \
    -run TestPushSender_QueueOverflow_A3W7_1_002
  -> PASS (~0.54s)

BREAK:
  notification/push.go NewPushSender — make(chan pushMessage, 1000) (cap >> 200)
  -> FAIL TestPushSender_QueueOverflow_A3W7_1_002
       depth 201 ≰ 200 (assert.LessOrEqual + Equal 200)

RESTORE make(chan pushMessage, 200)
  -> PASS (~0.56s)

git diff notification/push.go → empty
MUTATED marker → отсутствует
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=120s ./testframework/mail/sharedlimit/ \
  -run TestService_SendAlert_SharedDailyLimit_A3W7_1_001
go test -tags hubtest -count=1 -timeout=120s ./testframework/notification/pushoverflow/ \
  -run TestPushSender_QueueOverflow_A3W7_1_002
# BASELINE PASS · BREAK FAIL (per ID) · RESTORE PASS
```

## Gaps

- none — criteria IDs PROVEN  
- Docs-only / N/A: none  
- Wave119 `TestPushSender_QueueOverflow` — соседний; kill-канон `*_A3W7_1_002` (как в criteria)

## Handoff

- Evidence: этот файл · TASK-162 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored; no leftover MUTATED; no commit
