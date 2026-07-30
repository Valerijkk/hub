# MUTATION W13 slot 3 refill — TASK-121 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost hubtest · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/archive/2026-07-28/TASK-121-2026-07-28-marketplace-status-dev-exposure.md` `[MUTATION CRITERIA]`  
> Prior: W12 qa-auto2 — W7-A5-004 **GAP_NOTED** (weak final-state)  
> Runner: `_run_mutations_w13_task121.py` · Evidence JSON: `_mut-batch-w13-task121.json`  
> Коммит: **не создавался**. leftover MUTATED: **нет**.

## Verdict

**5/6 PROVEN** · **W7-A5-001 N/A** (layer_mismatch + source-assert) · **0 GAP_NOTED / NOT_PROVEN**  
→ TASK-121 → **Закрыта** (004 atomicity mid-window kill PROVEN)

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| W7-A5-001 | **N/A** | — (stub stays GREEN) | wiring_core: снять `requireAuth`/`RequireAdmin` с `/docs` | wiring ≠ stub Fiber |
| W7-A5-002 | **PROVEN** | `TestService_GetPublicStatus_NoComponents_W7A5002` + `TestHandler_GetStatus_PublicAggregate_W7A5002` | `GetPublicStatus` → `getFull`; `/status/details` без `authMw` | service+handler |
| W7-A5-003 | **PROVEN** | `TestListMine_PaginationCap_W7A5003` | `ListListingsByAuthor`: убрать `Limit(perPage)` | repository |
| W7-A5-004 | **PROVEN** | `TestUpdateApproved_AtomicRemoderation_W7A5004` | split: `UpdateListing` + `SetListingStatus` | repository |
| W7-A5-005 | **PROVEN** | `TestAuthorName_NoEmailOnPublicCatalog_W7A5005` | `displayName` public → `return u.Email` | dto |
| W7-A5-006 | **PROVEN** | `TestValidatePriceCurrency_W7A5006` | убрать `priceCents<0` / `isValidCurrency` | service |

Production restored (`git checkout --` после каждого BREAK). Leftover MUTATED: нет.

## Fix for prior GAP (W7-A5-004)

W12: final-state assert (`pending`+new title) оставался GREEN при split — mid-window не наблюдался.

W13 strengthen: Ent `MarketplaceListing.Use(hook.On(UpdateOne))` после каждого Save читает row; если `status==approved && title==newTitle` → fail.

- Atomic one-Save: hook видит только `pending`+newTitle → GREEN  
- Split Update+SetStatus: после первого Save hook видит `approved`+newTitle → RED

## Per-ID evidence

### W7-A5-001 — /api/docs authz (N/A, layer_mismatch)

```text
SOURCE-ASSERT wiring_core.go:
  api.Get("/docs", …, requireAuth, rbac.RequireAdmin(rbacRepo)) ✅
  api.Get("/docs/doc.json", …, requireAuth, rbac.RequireAdmin(rbacRepo)) ✅

MISMATCH DEMO: снять middleware с prod wiring_core
→ TestAPIDocs_RequiresAuthAdmin_W7A5001 остался PASS (stub Fiber, не wireCore)
→ fake-PROVEN запрещён; verdict = N/A
```

### W7-A5-002 — public status aggregate (PROVEN)

```text
BASELINE → PASS both
BREAK GetPublicStatus → return s.getFull(ctx)
→ FAIL Len==5 expected 1
RESTORE → PASS

BREAK /status/details без authMw
→ FAIL expected 401 actual 503
RESTORE → PASS
```

### W7-A5-003 — ListMine Limit (PROVEN)

```text
BREAK ListListingsByAuthor: drop Limit(perPage)
→ FAIL Len==3 expected 2
RESTORE → PASS
```

### W7-A5-004 — atomic remoderation (PROVEN)

```text
BASELINE → PASS (~4.1s)

BREAK UpdateApprovedListingToPending → UpdateListing + SetListingStatus(l.Version)
→ FAIL: mid-window approved+new title must not be visible between Saves
RESTORE → PASS (~5.0s)

SOURCE-ASSERT: SetStatus(pending) + fields в одном upd.Save ✅
```

### W7-A5-005 — author_name PII (PROVEN)

```text
BREAK displayName: return u.Email when !allowEmail
→ FAIL expected «Пользователь» actual email
RESTORE → PASS
```

### W7-A5-006 — price/currency (PROVEN)

```text
BREAK validateListingInput: skip price/currency guards
→ FAIL negative price / XYZ → err==nil
RESTORE → PASS
```

## Commands

```text
$env:POSTGRES_TEST_URL = "postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL = "mongodb://hub:hub@localhost:27017/?authSource=admin"
$env:PYTHONIOENCODING = "utf-8"
python .planning/audits/total-2026-07/_run_mutations_w13_task121.py
→ wall ~76s · 001 N/A · 002(+b)/003/004/005/006 PROVEN
→ git diff prod empty after restore

go test -tags hubtest -count=1 -timeout=180s ./testframework/marketplace/ ./testframework/status/ ./testframework/dev/
→ PASS (domain regression)
```

## Assumptions / unresolved

- **W7-A5-001:** N/A — layer_mismatch stub≠wiring; source-assert OK; не блокирует Закрыта (как W12).
- Коммит не создавался.

## Handoff

TASK-121 **Закрыта** · evidence этот файл · 004 PROVEN mid-window · next: qa-auto-ver / code-reviewer.
