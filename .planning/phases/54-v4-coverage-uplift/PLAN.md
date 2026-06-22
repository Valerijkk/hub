# Phase 54 — v4 Coverage Uplift (bases + marketplace)

**Status:** COMPLETE  
**Branch:** `feat/v2.5-trust`  
**Requirement:** PLAN-CLOSE-GAPS — v4 modules ≥60% stmt

## Goal

Raise low-coverage v4 packages `bases/` and `marketplace/` to ≥60% via service + handler tests.

## Delivered

### bases (`service_test.go` + handler extensions)

- Invalid name, version conflict, delete base
- Table/column/record lifecycle + cell normalization
- HTTP: update base, delete record, delete base

### marketplace (`service_test.go` + handler extensions)

- Create validation, approve/reject/sold/withdraw flows
- Re-moderation on approved edit
- List mine/approved, delete permissions
- `formatPrice` helpers
- HTTP: reject, sold, withdraw, list mine

### Coverage

| Package | Before | After |
|---------|--------|-------|
| bases | 34.2% | **60.8%** ✅ |
| marketplace | 32.9% | **69.2%** ✅ |

## Verification

```bash
cd backend && go test -cover ./internal/bases/... ./internal/marketplace/...
```

## Next

Phase 55 — onec/intranet → 50%+ · squash commit v4 (40–54)
