# MUTATION W13 slot 3 refill — TASK-157 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/archive/2026-07-29/TASK-157-2026-07-29-a3-5-5-search-tests.md` `[MUTATION CRITERIA]`  
> Findings: **A3W5-5-001**, **A3W5-5-002** (2 IDs)  
> Коммит не создавался.

## Verdict

**1/1 code PROVEN · 1/1 hygiene N/A** · TASK-157 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED |
|----|---------|-----|
| A3W5-5-001 | **PROVEN** | `TestPostgresSearch_LimitClamp_A3W55_001` |
| A3W5-5-002 | **N/A** (hygiene dedupe) | — (not fake-PROVEN; canons PASS) |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only / hygiene N/A IDs: A3W5-5-002`

## Per-ID evidence

### A3W5-5-001 — `Search` limit clamp (`service.go:62-64`)

```text
BASELINE
  go test -tags hubtest -count=1 -timeout=180s ./testframework/search/ \
    -run TestPostgresSearch_LimitClamp_A3W55_001
  -> PASS (~1.14s)

BREAK:
  backend/internal/search/service.go — `limit > 50` → `limit > 500`
  (+ MUTATED comment)
  -> FAIL TestPostgresSearch_LimitClamp_A3W55_001
       limit=100 must clamp to default 20
       should have 20 item(s), but has 30

RESTORE original clamp `if limit <= 0 || limit > 50 { limit = 20 }`
  -> PASS (~1.10s)

git diff backend/internal/search/service.go → empty
MUTATED marker → отсутствует
```

### A3W5-5-002 — nil-mongo / closed-DB dedupe · **N/A code mutation**

```text
Guard layer: test hygiene (dedupe), not prod Search clamp

Evidence (hygiene N/A, not PROVEN):
  - Canon nil-mongo: TestSearchMessages_NilMongoReturnsEmpty
      backend/testframework/search/service_scope_test.go:169
  - Canon closed-DB: TestGetUserChannelIDs_ClosedDB
      backend/testframework/search/service_wave106_test.go:88
  - Duplicate bodies removed; leftover comments point to canons
      (service_shard_errors_test.go:42 · service_wave106_test.go:105)

RESTORE canons PASS:
  go test -tags hubtest -count=1 ./testframework/search/ \
    -run "TestSearchMessages_NilMongoReturnsEmpty|TestGetUserChannelIDs_ClosedDB"
  -> PASS (same package ok ~1.1s with clamp restore)

Kill recipe per criteria: N/A — do not mutate prod for dedupe.
No fake-PROVEN.
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=180s ./testframework/search/ \
  -run "TestPostgresSearch_LimitClamp_A3W55_001|TestSearchMessages_NilMongoReturnsEmpty|TestGetUserChannelIDs_ClosedDB"
# BASELINE PASS · BREAK FAIL (001 Len=30≠20) · RESTORE PASS
```

## Gaps

- **A3W5-5-002** = hygiene N/A (dedupe). No prod kill.
- `TestSearch_LimitClampedToDefault` remains Unavailable-path — not 001 target (per criteria Gaps).

## Handoff

- Evidence: этот файл · TASK-157 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored; no leftover MUTATED; no commit
