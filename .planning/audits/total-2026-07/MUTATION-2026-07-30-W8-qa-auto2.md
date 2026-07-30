# MUTATION W8 slot 2 — TASK-089 (qa-auto2) · 2026-07-30

> Checker: **qa-auto2** (!= be2 fixer) · Env: localhost Mongo `:27017` + Postgres `:5432` (deployments-*)  
> Criteria: `.planning/backlog/TASK-089-2026-07-28-messenger-crm-invoice-guards.md` `[MUTATION CRITERIA]`

## Verdict

**3/3 PROVEN** · Break→RED→restore→GREEN · статус TASK-089 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A2W3-3-004 | **PROVEN** | `TestAddReaction_ConcurrentFirstReact_NoDuplicate_A2W3_3_004` | убрать `"reactions.emoji": {$ne: emoji}` из push-фильтра | Mongo repository |
| W7-A2-005 | **PROVEN** | `TestCreateDeal_RejectsStrangerContact_W7A2005` | `validateDealRefs`: `authorizeContactWrite` → `repo.GetContact` | Ent CRM service |
| W7-A2-007 | **PROVEN** | `TestCreateInvoice_ConcurrentUniqueNumbers_W7A2007` | `Count(prefix)+1` + без retry на `ent.IsConstraintError` (+50ms sleep после Count для reproducible TOCTOU) | Ent invoice repository |

Production restored (StrReplace reverse; `git diff` на трёх prod-файлах empty). Mock-gap: нет — Mongo/`mongotest`, Ent CRM sqlite, Ent invoice Postgres.

## Per-ID evidence

### A2W3-3-004 — Mongo reaction `$ne` push filter

```text
BASELINE:
  go test -tags hubtest -race -count=1 ./testframework/messenger/ \
    -run TestAddReaction_ConcurrentFirstReact_NoDuplicate_A2W3_3_004
  → PASS (~3.5s)

BREAK: backend/internal/messenger/mongo_repository.go
  push filter: {_id, reactions.emoji: {$ne: emoji}} → {_id} only
→ FAIL:
  fireEntries expected 1 actual 2
  ElementsMatch userIDs diverge (duplicate ReactionEntry)
→ FAIL package

RESTORE: вернуть $ne
→ PASS
```

### W7-A2-005 — CRM contact owner-scope in validateDealRefs

```text
BASELINE:
  go test -tags hubtest -count=1 ./testframework/crm/ \
    -run TestCreateDeal_RejectsStrangerContact_W7A2005
  → PASS (~1.4s)

BREAK: backend/internal/crm/service.go validateDealRefs
  authorizeContactWrite(ctx, userID, *contactID)
  → repo.GetContact(ctx, *contactID)  (без owner/admin)
→ FAIL:
  Expected error crm.ErrForbidden but got nil
→ FAIL package

RESTORE: вернуть authorizeContactWrite
→ PASS (также admin twin TestCreateDeal_AdminMayLinkForeignContact_W7A2005 green)
```

### W7-A2-007 — invoice serial Count+1 + no constraint retry

```text
BASELINE:
  POSTGRES_TEST_URL=postgresql://hub:hub@localhost:5432/hub?sslmode=disable
  go test -tags hubtest -race -count=1 ./testframework/accounting/ \
    -run TestCreateInvoice_ConcurrentUniqueNumbers_W7A2007
  → PASS (~3.8s)

BREAK: backend/internal/accounting/invoice_repository.go
  nextInvoiceNumber: MAX(suffix)+1 → Count(prefix)+1
  CreateInvoice: retry loop на IsConstraintError → single createInvoiceOnce
  Note: Count+1 без sleep на этом хосте ×20 PASS (узкое TOCTOU-окно —
  второй Count часто видит уже committed INSERT). Для reproducible RED
  добавлен time.Sleep(50ms) после Count (только mutation prove).
→ FAIL ×5/5:
  create invoice: ent: constraint failed: pq: duplicate key value
  violates unique constraint "invoice_number"
→ FAIL package

RESTORE: MAX+1 + retry (8 attempts) + imports/const
→ PASS
```

## Commands

```text
$env:POSTGRES_TEST_URL = "postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL  = "mongodb://hub:hub@localhost:27017/?authSource=admin"
cd backend

# named tests (baseline / RED / GREEN)
go test -tags hubtest -race -count=1 -timeout=120s ./testframework/messenger/ \
  -run TestAddReaction_ConcurrentFirstReact_NoDuplicate_A2W3_3_004
go test -tags hubtest -count=1 -timeout=60s ./testframework/crm/ \
  -run TestCreateDeal_RejectsStrangerContact_W7A2005
go test -tags hubtest -race -count=1 -timeout=120s ./testframework/accounting/ \
  -run TestCreateInvoice_ConcurrentUniqueNumbers_W7A2007

# final restore suite (crm incl. admin twin)
go test -tags hubtest -count=1 ./testframework/crm/ -run "TestCreateDeal_.*_W7A2005"
→ all PASS · git diff prod empty
```

## Assumptions / unresolved

- W7-A2-007: чистый Count+1 без retry на localhost может не RED без расширения TOCTOU (scheduling); unique index всё равно даёт constraint при коллизии — sleep 50ms только для prove, в prod не остаётся.
- Коммит не создавался (инструкция слота).

## Handoff

TASK-089 **Закрыта** · evidence этот файл · README синхронизирован · next: qa-auto-ver / be-ver / code-reviewer по желанию волны (mutation gate закрыт).
