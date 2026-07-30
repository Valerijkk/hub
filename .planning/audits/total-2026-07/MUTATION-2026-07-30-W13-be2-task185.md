# MUTATION W13 slot 3 refill — TASK-185 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/TASK-185-2026-07-29-c-w19-allure-existingrun-partial.md` `[MUTATION CRITERIA]`  
> Findings: **C-W19-TT-002-001**, **C-W19-TT-002-002** (2 IDs)  
> Коммит не создавался. Service-layer kill; mock≠Ent GAP_NOTED (no live Mongo/Ent IT).

## Verdict

**2/2 PROVEN** (+ GAP_NOTED mock≠Ent) · TASK-185 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| C-W19-TT-002-001 | **PROVEN** + **GAP_NOTED** | `TestTask185_ImportAllureExistingRun_MongoStoreFail_FailClosed` |
| C-W19-TT-002-002 | **PROVEN** + **GAP_NOTED** | `TestTask185_ImportAllureExistingRun_MatchDBErr_FailClosed` |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: C-W19-TT-002-001, C-W19-TT-002-002` · `docs-only IDs: (none)`

## Per-ID evidence

### C-W19-TT-002-001 — StoreAllureResults err → NewInternal (not slog-only)

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/tms/ \
    -run TestTask185_ImportAllureExistingRun_MongoStoreFail_FailClosed
  -> PASS (~1.1s)

BREAK (service correct layer):
  backend/internal/tms/service_import.go ImportAllureToExistingRun
  — StoreAllureResults err: убрать return NewInternal; оставить slog-only
  — MUTATED C-W19-TT-002-001

  -> FAIL TestTask185_ImportAllureExistingRun_MongoStoreFail_FailClosed (~1.1s)
       An error is expected but got nil

RESTORE fail-closed NewInternal
  -> PASS (~1.2s)

git diff service_import.go → empty
MUTATED marker → отсутствует
```

Layer OK: guard в service; inject через mockRepo.storeAllureErr (не live Mongo) —
**GAP_NOTED** mock≠Ent; kill на prod service path PROVEN.

### C-W19-TT-002-002 — matchErr (infra ≠ NotFound) → return / NewInternal (not continue)

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/tms/ \
    -run TestTask185_ImportAllureExistingRun_MatchDBErr_FailClosed
  -> PASS (в общем baseline TestTask185_ ~1.1s)

BREAK (service correct layer):
  backend/internal/tms/service_import.go buildAllureResultInputs
  — matchErr: return → continue (silent skip)
  — MUTATED C-W19-TT-002-002

  -> FAIL TestTask185_ImportAllureExistingRun_MatchDBErr_FailClosed (~1.3s)
       An error is expected but got nil

RESTORE return nil, TestRunStats{}, matchErr
  -> PASS TestTask185_ (~1.1s)

git diff service_import.go → empty
MUTATED marker → отсутствует
```

Layer OK: guard в service + buildAllureResultInputs; inject через mockRepo.findMatchErr —
**GAP_NOTED** mock≠Ent; kill на prod service path PROVEN.

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=120s ./testframework/tms/ -run "TestTask185_"
# BASELINE PASS ~1.1s · FINAL PASS ~1.1s
```

## Gaps

- **mock≠Ent** (оба ID): fail-closed доказан на service + mock inject; нет live Mongo Store IT / Ent matcher IT для ExistingRun.
- Create-path Mongo best-effort — не kill target (criteria).
- Path B Partial — не выбран; kill = fail-closed.

## Handoff

- Evidence: этот файл · TASK-185 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored after each BREAK; `git diff` на `service_import.go` — empty · no leftover MUTATED
