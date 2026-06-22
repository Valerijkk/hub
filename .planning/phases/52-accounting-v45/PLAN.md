# Phase 52 — V4-03.5 Accounting Invoices + P&L

**Status:** COMPLETE  
**Branch:** `feat/v2.5-trust`  
**Requirement:** V4-03.5 — invoices, income statement, CRM deal → draft invoice

## Goal

Extend Hub accounting with customer invoices (draft → issued → paid with auto journal), P&L report, and CRM won-deal hook.

## Delivered

### Ent

- `Invoice` — number, customer, amount, status, optional `crm_deal` + `journal_entry` edges

### Backend

- `GET/POST /api/accounting/invoices`
- `POST /api/accounting/invoices/:id/issue`
- `POST /api/accounting/invoices/:id/pay` — posts Dr 1010 / Cr 4010
- `GET /api/accounting/income-statement` — revenue/expense/net from posted entries
- CRM `UpdateDeal` → draft invoice when stage → `won` (if amount > 0)

### Frontend

- `/accounting` — P&L cards, invoices table, issue/pay actions

### Tests

- `handler_test.go` — invoice flow + income statement
- Vitest accounting page mocks invoices/P&L

## Out of scope

- PDF export, НДФЛ, multi-currency FX, AR aging

## Verification

```bash
cd backend && go test ./internal/accounting/...
cd frontend && npx vitest run src/modules/accounting/
```

## Next

Phase 53 — commit v4 block (40–52) · PLAN-CLOSE-GAPS coverage wave
