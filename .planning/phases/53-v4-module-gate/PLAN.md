# Phase 53 — v4 Module Gate + CRM↔Accounting Integration

**Status:** COMPLETE  
**Branch:** `feat/v2.5-trust`  
**Requirement:** PLAN-CLOSE-GAPS Wave 7 — v4 super-app test hardening

## Goal

Gate phases 40–52: integration tests for CRM won→invoice, HR reject flow, E2E accounting v4.5 UI, qa-gate green.

## Delivered

### Backend tests

- `crm/service_invoice_test.go` — won deal triggers `DealInvoiceCreator`; zero amount skipped
- `accounting/service_invoice_test.go` — `CreateInvoiceFromWonDeal` idempotent
- `hr/handler_test.go` — leave reject flow

### E2E

- `v4-superapp.spec.ts` — accounting invoices + P&L smoke (Phase 52 UI)

### Coverage uplift (v4 packages)

| Package | Before | After |
|---------|--------|-------|
| accounting | 59.1% | **65.0%** |
| crm | 46.3% | **48.5%** |
| hr | 44.2% | **53.1%** |

## Verification

```bash
cd backend && go build -tags release ./... && go test ./internal/crm/... ./internal/accounting/... ./internal/hr/...
cd frontend && npx vitest run && npm run lint
```

## Gate

| Check | Result |
|-------|--------|
| Vitest | 80/80 ✅ |
| Biome | 0 errors ✅ |
| Release build | ✅ |
| v4 module tests | ✅ |

## Next

Phase 54 — squash commit v4 block (40–53) when requested · v4 module coverage → 60% (bases/marketplace)
