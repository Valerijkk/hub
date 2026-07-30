# MUTATION W12 slot 2 — TASK-121 (qa-auto2) · 2026-07-30

> Checker: **qa-auto2** (!= be1 fixer) · Env: localhost hubtest  
> Criteria: `.planning/backlog/TASK-121-2026-07-28-marketplace-status-dev-exposure.md` `[MUTATION CRITERIA]`  
> Runner: `_run_mutations_w12_task121.py` · Evidence JSON: `_mut-batch-w12-task121.json`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**4/6 PROVEN** · **W7-A5-001 N/A** (layer_mismatch + source-assert, не fake-PROVEN) · **W7-A5-004 GAP_NOTED** (weak final-state)  
→ TASK-121 → **Открыта** (причина: 004 atomicity не убита мутацией; нужен stronger kill / owner accept GAP)

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| W7-A5-001 | **N/A** | — (stub stays GREEN) | wiring_core: снять `requireAuth`/`RequireAdmin` с `/docs` | wiring ≠ stub Fiber |
| W7-A5-002 | **PROVEN** | `TestService_GetPublicStatus_NoComponents_W7A5002` + `TestHandler_GetStatus_PublicAggregate_W7A5002` | `GetPublicStatus` → `getFull`; `/status/details` без `authMw` | service+handler |
| W7-A5-003 | **PROVEN** | `TestListMine_PaginationCap_W7A5003` | `ListListingsByAuthor`: убрать `Limit(perPage)` | repository |
| W7-A5-004 | **GAP_NOTED** | `TestUpdateApproved_AtomicRemoderation_W7A5004` (остался PASS) | split: `UpdateListing` + `SetListingStatus` | repository (weak) |
| W7-A5-005 | **PROVEN** | `TestAuthorName_NoEmailOnPublicCatalog_W7A5005` | `displayName` public → `return u.Email` | dto |
| W7-A5-006 | **PROVEN** | `TestValidatePriceCurrency_W7A5006` | убрать `priceCents<0` / `isValidCurrency` | service |

Production restored (`git checkout --` после каждого BREAK). Leftover MUTATED: нет.

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

Нужен IT против реального `wireCore` или source-assert (сделан) — PROVEN через stub невозможен.

### W7-A5-002 — public status aggregate (PROVEN)

```text
BASELINE → PASS both

BREAK service GetPublicStatus → return s.getFull(ctx)
→ FAIL service: Len==5 expected 1 (postgres/redis/…)
→ FAIL handler: Len==5 + body contains component ids
RESTORE → PASS

BREAK handler: /status/details без authMw
→ FAIL: expected 401 actual 503 (anon got details)
RESTORE → PASS
```

### W7-A5-003 — ListMine Limit (PROVEN)

```text
BASELINE → PASS

BREAK ListListingsByAuthor: drop Limit(perPage) (Offset остаётся)
→ FAIL: Len==3 expected 2 (page1 per_page=2)
RESTORE → PASS
```

### W7-A5-004 — atomic remoderation (GAP_NOTED)

```text
BASELINE → PASS

BREAK UpdateApprovedListingToPending → UpdateListing + SetListingStatus(l.Version)
→ still PASS (final pending+new title; mid-window не наблюдаем)
RESTORE → PASS

SOURCE-ASSERT: SetStatus(pending) + fields в одном upd.Save ✅
```

Criteria Gaps: final-state assert weak; Ent inject между двумя Save — ideal. Не fake-PROVEN.

### W7-A5-005 — author_name PII (PROVEN)

```text
BREAK displayName: return u.Email when !allowEmail
→ FAIL: expected «Пользователь» actual email; contains "@"
RESTORE → PASS
```

### W7-A5-006 — price/currency (PROVEN)

```text
BREAK validateListingInput: skip price/currency guards
→ FAIL: negative price / XYZ currency → err==nil (expected ErrInvalidInput)
RESTORE → PASS
```

## Commands

```text
$env:POSTGRES_TEST_URL = "postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:PYTHONIOENCODING = "utf-8"
python .planning/audits/total-2026-07/_run_mutations_w12_task121.py
→ wall ~199s · 001 N/A · 002(+b) PROVEN · 003 ERROR(anchor)→reprove PROVEN · 004 GAP_NOTED · 005/006 PROVEN
→ git diff prod empty after restore
```

## Assumptions / unresolved

- **W7-A5-001:** N/A — нужны IT на wireCore или принять source-assert как gate (сделан; не PROVEN).
- **W7-A5-004:** GAP_NOTED — stronger mid-window kill вне scope слота.
- Коммит не создавался.

## Handoff

TASK-121 **Открыта** · evidence этот файл · mutation gate неполный (004) · next: stronger 004 test / owner accept GAP → qa-auto-ver.
