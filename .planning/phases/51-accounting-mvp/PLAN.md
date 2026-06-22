# Phase 51 — V4-03 Accounting MVP + v4 E2E

**Status:** COMPLETE  
**Branch:** `feat/v2.5-trust`  
**Requirement:** V4-03 — Hub built-in accounting (no 1C, no payroll)

## Goal

Double-entry ledger for admins: chart of accounts, journal entries (draft → posted), trial balance. Plus Playwright smoke for v4 routes.

## Delivered

### Ent

- `LedgerAccount`, `JournalEntry`, `JournalLine`

### Backend (`internal/accounting/`)

- Default COA seed (1010 Bank, 2010 AP, 3010 Equity, 4010 Revenue, 5010 Expense)
- Admin-only API under `/api/accounting/*`
- Balanced entry validation (debits == credits)
- Trial balance from posted entries

### Frontend

- `/accounting` — admin-only (founder/admin redirect)
- Chart, journal form, post draft, trial balance table

### E2E

- `frontend/e2e/v4-superapp.spec.ts` — v4 route smoke

## Out of scope

- Payroll, НДФЛ, invoices PDF, multi-currency, CRM deal auto-posting

## Verification

```bash
cd backend && go test ./internal/accounting/...
cd frontend && npx vitest run && npm run lint
```

## Next

Phase 52 — commit v4 block (40–51) · invoices/reports accounting v4.5
