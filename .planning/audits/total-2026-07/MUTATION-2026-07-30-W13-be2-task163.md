# MUTATION W13 slot 3 refill — TASK-163 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/TASK-163-2026-07-29-a3-7-2-v4-business.md` `[MUTATION CRITERIA]`  
> Findings: **A3W7-2-001**, **A3W7-2-002**, **A3W7-2-003** (3 IDs)  
> Коммит не создавался.

## Verdict

**3/3 PROVEN** · TASK-163 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| A3W7-2-001 | **PROVEN** | `TestCoffeeMatch_NotifyFail_WarnContext_Task163` |
| A3W7-2-002 | **PROVEN** | `TestUpdateNewsPoll_VersionEQ_Task163`; `TestUpdateNewsPoll_Concurrent_VersionEQ_Task163` |
| A3W7-2-003 | **PROVEN** | `TestGetMyCoffeeMatch_TriplePartners_Task163` |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)`

## Per-ID evidence

### A3W7-2-001 — NotifyCoffeeMatch err → slog.WarnContext

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/intranet/ \
    -run "TestCoffeeMatch_NotifyFail_WarnContext_Task163|…"
  -> PASS (~0.9s)

BREAK:
  coffee_service.go notifyCoffeeMatch → _ = NotifyCoffeeMatch (no WarnContext)
  (+ temporary drop unused log/slog import for compile)
  -> FAIL TestCoffeeMatch_NotifyFail_WarnContext_Task163
       "" does not contain "intranet: notify coffee match failed"

RESTORE notifyCoffeeMatch + slog import → WarnContext path
  -> PASS (~0.7s)

git diff coffee_service.go → empty
MUTATED marker → отсутствует
```

### A3W7-2-002 — UpdateNewsPoll version≥1 + VersionEQ

```text
BREAK:
  service_polls.go — remove req.Version < 1 → ErrInvalidInput
  repo_polls.go — UpdateOneID without Where(newspoll.VersionEQ)
  -> FAIL TestUpdateNewsPoll_VersionEQ_Task163
       version=0 / omit → nil (expected ErrInvalidInput);
       stale version → nil (expected ErrVersionConflict)
  -> FAIL TestUpdateNewsPoll_Concurrent_VersionEQ_Task163
       second writer OK + lost update ("loser" wins)

RESTORE VersionEQ + version < 1 guard
  -> PASS (~0.9s)

git diff service_polls.go repo_polls.go → empty
MUTATED marker → отсутствует
```

### A3W7-2-003 — GetMyCoffeeMatch triple → Partners[] len=2

```text
BREAK:
  coffeePartnersForUser → First/один Partner (no hub-spoke)
  -> FAIL TestGetMyCoffeeMatch_TriplePartners_Task163
       len(Partners)=1, want 2 ("triple participant must see both partners")

RESTORE hub-spoke coffeePartnersForUser
  -> PASS (~1.0s full Task163 suite)

git diff coffee_service.go → empty
MUTATED marker → отсутствует
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=180s ./testframework/intranet/ \
  -run "TestCoffeeMatch_NotifyFail_WarnContext_Task163|TestUpdateNewsPoll_VersionEQ_Task163|TestUpdateNewsPoll_Concurrent_VersionEQ_Task163|TestGetMyCoffeeMatch_TriplePartners_Task163"
# BASELINE PASS · BREAK FAIL (per ID) · RESTORE PASS
```

## Gaps

- none — criteria IDs PROVEN  
- Docs-only / N/A: none  
- DTO BWC `partner` = `partners[0]` — pair N=2 path not killed by 003 First-mutation on N=3 fixture

## Handoff

- Evidence: этот файл · TASK-163 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored; no leftover MUTATED; no commit
