# Phase 55 — onec + intranet Coverage (≥50%)

**Status:** COMPLETE  
**Branch:** `feat/v2.5-trust`  
**Requirement:** PLAN-CLOSE-GAPS Wave 7 — remaining v4 integrations

## Goal

Raise `onec/` (38%) and confirm `intranet/` (54%) ≥50% statement coverage.

## Delivered

### onec (`service_extended_test.go`)

- Config validate/decode/encode
- UpdateSettings (valid, invalid field, enable-without-url)
- TestWebhook (not configured, OK, error status)
- ExportApprovedLeave (disabled skip, bad HTTP status, bearer + phone mapping)
- buildLeavePayload edge cases
- AdminHandler update + test + forbidden

### intranet (`service_test.go`)

- News update version conflict
- Kudo self/invalid input
- Admin draft list vs public list
- Draft hidden from members

### Coverage

| Package | Before | After |
|---------|--------|-------|
| onec | 38.3% | **72.0%** ✅ |
| intranet | 54.2% | **61.2%** ✅ |

## Verification

```bash
go test -cover ./internal/integrations/onec/... ./internal/intranet/...
```

## Next

Phase 56 — v4 final gate (full qa-gate) · squash commit 40–55
