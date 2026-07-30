# MUTATION W13 Stage C — TASK-079 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/TASK-079-gdpr-mongo-eraser-fail-closed.md` `[MUTATION CRITERIA]`  
> Findings: **W18-C5-001**, **W18-C5-002** (2 code-mutable)  
> HEAD baseline: `c10b874d` · Коммит **не** создавался.

## Verdict

**2/2 PROVEN** · TASK-079 → **Закрыта** · leftover MUTATED: **no**

| ID | Verdict | RED test name |
|----|---------|---------------|
| W18-C5-001 | **PROVEN** | `TestMongoMessageEraser_NilColl_FailClosed` |
| W18-C5-002 | **PROVEN** | `TestMongoMessageEraser_DeleteBySender_EntMongo` |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)` · `GAP IDs: (none)`

## Per-ID evidence

### W18-C5-001 — nil eraser/coll fail-closed (`gdpr_mongo.go:34-35`)

```text
BREAK:
  if e == nil || e.coll == nil {
    return 0, nil   // was ErrMongoMessageEraserUnavailable
  }

RED:
  go test -tags hubtest -count=1 -timeout=120s ./testframework/audit/ \
    -run TestMongoMessageEraser_NilColl_FailClosed
  -> FAIL (~0.95s)
       Expected error with "mongo message eraser unavailable" in chain but got nil
       (gdpr_mongo_eraser_task079_test.go:21)

RESTORE:
  return 0, ErrMongoMessageEraserUnavailable
```

### W18-C5-002 — DeleteMany filter `sender_id` (`gdpr_mongo.go:37`)

```text
BREAK:
  bson.M{"user_id": senderID}   // was "sender_id"

RED (001 already restored → PASS; 002 RED):
  go test -tags hubtest -count=1 -timeout=120s ./testframework/audit/ \
    -run "TestMongoMessageEraser_NilColl_FailClosed|TestMongoMessageEraser_DeleteBySender_EntMongo"
  -> FAIL TestMongoMessageEraser_DeleteBySender_EntMongo (~0.90s)
       expected deleted=2 actual=0
       expected remainA=0 actual=2
       (gdpr_mongo_eraser_task079_test.go:50,54)

RESTORE:
  bson.M{"sender_id": senderID}
```

### GREEN after full restore

```text
go test -tags hubtest -count=1 -timeout=120s ./testframework/audit/ \
  -run "TestMongoMessageEraser_NilColl_FailClosed|TestMongoMessageEraser_DeleteBySender_EntMongo"
-> ok  github.com/valerii/hub/backend/testframework/audit  0.937s

git diff -- backend/internal/audit/gdpr_mongo.go → empty
rg MUTATED backend/internal/audit/gdpr_mongo.go → no matches
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=120s ./testframework/audit/ \
  -run "TestMongoMessageEraser_NilColl_FailClosed|TestMongoMessageEraser_DeleteBySender_EntMongo"
```

## Gaps

- (none)

## Handoff

- TASK-079 → **Закрыта** (TASK.md + backlog README)
- Evidence: этот файл
- leftover MUTATED: **no**
- Коммит: не создавался (по инструкции)
