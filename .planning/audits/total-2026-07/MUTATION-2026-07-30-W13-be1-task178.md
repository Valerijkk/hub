# MUTATION W13 slot 1 refill — TASK-178 (be1) · 2026-07-30

> Checker: **be1** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/archive/2026-07-29/TASK-178-2026-07-29-c-w19-gdpr-mongo-erase-retry.md` `[MUTATION CRITERIA]`  
> Коммит не создавался (инструкция слота).

## Verdict

**1/1 PROVEN** · TASK-178 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| C-W19-002 | **PROVEN** | `TestDeleteUserGDPR_MongoEraseFail_EnqueuesRetry_CW19002` |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)`

## Per-ID evidence

### C-W19-002 — post-Commit eraser fail → EnqueueMongoErase

```text
BASELINE TestDeleteUserGDPR_MongoEraseFail_EnqueuesRetry_CW19002 -> PASS (~0.9s)

BREAK: audit/service.go DeleteUserGDPR —
  убрать блок if s.eraseRetry != nil { … EnqueueMongoErase … }
  (оставить только slog + return eraser err)
-> FAIL gdpr_mongo_erase_retry_c_w19_test.go:79
   "[]" should have 1 item(s), but has 0
   Messages: compensating retry must enqueue on eraser fail
RESTORE -> PASS (~1.2s) + OK/NoRetry + EnqueueAlsoFails
```

Guard: `backend/internal/audit/service.go:639-646` · contract `gdpr_mongo_retry.go` · subject `SubjectGDPRMongoErase`.

Note: stub `eraseRetry` intentional (service-layer enqueue); kill на реальном prod guard → RED именованного теста. HardDelete остаётся committed (orphan window до consumer).

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=120s ./testframework/audit/ -run `
  "TestDeleteUserGDPR_MongoEraseFail_EnqueuesRetry_CW19002|TestDeleteUserGDPR_MongoEraseOK_NoRetry_CW19002|TestDeleteUserGDPR_MongoEraseFail_EnqueueAlsoFails_CW19002"
# FINAL GREEN PASS ~1.2s
```

## Gaps

- HardDelete already committed before erase — kill не откатывает Ent; asserts orphan + missing enqueue (documented in criteria).
- Docs-only: **N/A**

## Handoff

- Evidence: этот файл · TASK-178 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored after BREAK; `git diff` на `audit/service.go` — empty; no `MUTATED` leftover  
- Prod code unchanged this slot (fix already on HEAD; mutation prove only)
