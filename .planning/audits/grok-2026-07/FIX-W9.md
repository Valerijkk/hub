# FIX-W9 — Optimistic lock / timer / invoice / coffee

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-27  
**Агент:** be1  
**Предшественник:** FIX-W8 (`4c68b67`)

Правила волны: 1 commit / OPEN finding; mutation proof; no schema; no push master.  
VERIFY-first для A3-004 / A3-005 (W3/W4).

Docker: `postgres` / `mongo` / `redis` healthy; `POSTGRES_TEST_URL=postgresql://hub:hub@localhost:5432/hub?sslmode=disable`.

---

## Итог

| ID | status | SHA |
|----|--------|-----|
| A3-004 | CLOSED_ALREADY | `cd852a8` / Ent `9d0cec1` (W3/W4) |
| A6-003 | FIXED | `cc43d4b` |
| A3-005 | CLOSED_ALREADY | `d280eaa` (W3) |
| A6-004 | FIXED | `d3e8122` |
| A7-006 | FIXED | `9e2bdda` |

---

## VERIFY (skip fix)

### A3-004 — TMS UpdateTestCase VersionEQ

**Вердикт:** CLOSED_ALREADY (W3 `cd852a8` + W4 Ent `9d0cec1`).

- Prod: `tms/ent_repository.go` — `Where(IDEQ, VersionEQ)` + `affected==0` → 409.
- Тесты: `TestUpdateTestCase_OptimisticLock_A3_004` (mock) + `TestEntRepository_UpdateTestCase_OptimisticLock_A3_004` (Postgres Ent).
- Re-verify 2026-07-27: `go test -tags hubtest ./testframework/tms/ -run A3_004` → **PASS**.

### A3-005 — LogTime + StopTimer nil panic

**Вердикт:** CLOSED_ALREADY (W3 `d280eaa`).

- Prod: `LogTime` пишет `StartedAt`+`EndedAt`; `StopTimer` — `EndedAtIsNil` **и** `StartedAtNotNil`.
- Тест: `TestEntRepository_LogTime_DoesNotMatchStopTimer_A3_005`.
- Re-verify 2026-07-27: `go test -tags hubtest ./testframework/tracker/ -run A3_005` → **PASS**.

---

## FIXED

### A6-003 — CRM / marketplace optimistic lock

**Severity:** P1 · **Коммит:** `cc43d4b`

**Проблема:** app-level `if version > 0` + `UpdateOneID` без `VersionEQ` → `version:0` bypass и lost update.

**Фикс:**
- CRM `UpdateContact` / `UpdateDeal`: `Where(DeletedAtIsNil, VersionEQ)` + require `version >= 1`; NotFound → `ErrVersionConflict`.
- Marketplace `UpdateListing`: всегда `VersionEQ`; approve/withdraw/sold — require `version >= 1`, CAS по client version.

**Тесты:**
- `TestUpdateContact_OptimisticLock_A6_003` / `TestUpdateDeal_OptimisticLock_A6_003`
- `TestUpdateListing_OptimisticLock_A6_003`

**Mutation:** снять `crmcontact.VersionEQ` → stale write `err==nil` → **FAIL**; restore → **PASS**.

### A6-004 — duplicate invoice per deal (no schema)

**Severity:** P1 · **Коммит:** `d3e8122`

**Проблема:** `InvoiceExists` + `Create` TOCTOU; schema unique на deal→invoice запрещён в волне.

**Фикс (app-level):**
- `CreateInvoice` при `CrmDealID`: Ent tx → row-lock deal (`AddVersion(0)`) → Exist → Create → Commit.
- `CreateInvoiceFromWonDeal`: `ErrInvoiceExists` → idempotent `nil`.

**Тест:** `TestCreateInvoice_ConcurrentSinglePerDeal_A6_004` (Postgres, 2 goroutines → 1 invoice).

**Mutation:** убрать Exist в tx → `okCount=2`, invoices=2 → **FAIL**; restore → **PASS**.

### A7-006 — coffee rematch Offset(1) off-by-one

**Severity:** P1 · **Коммит:** `9e2bdda`

**Проблема:** `GetPreviousRoundPairKeys` с `Offset(1)` вызывается **до** `CreateCoffeeRound` → пропускает реальный previous; при 1 раунде avoid-map пуст.

**Фикс:** `First()` по `created_at DESC` (latest), без Offset.

**Тест:** обновлён `TestEntRepository_CoffeeUpsertAndPreviousPairs` — после 1 раунда keys содержат пару; после 2 — latest, не round-before-last.

**Mutation:** вернуть `Offset(1)` → keys empty / wrong round → **FAIL**; restore → **PASS**.

---

## Коммиты W9

```text
cc43d4b fix(crm,marketplace): enforce VersionEQ optimistic lock (A6-003)
d3e8122 fix(accounting): atomic invoice-per-deal create under deal lock (A6-004)
9e2bdda fix(intranet): use latest coffee round for rematch avoid-set (A7-006)
```

*Merge в `master` — только пользователем.*
