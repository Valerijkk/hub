# Phase 108 — 79% COV Milestone

**Branch:** `feat/v2.5-trust`  
**Goal:** Cross **79%** overall statement coverage via accounting + search + tms boundary tests.

## Targets

| Package | Before | Target |
|---------|--------|--------|
| `accounting` | 76.5% | ≥80% |
| `tms` | 76.5% | ≥76.5% |
| `search` | 75.1% | ≥75.1% |
| COV overall | 78.9% | **≥79%** |

## Tasks

1. **Accounting handler errors** — PostJournal not-found, CreateInvoice invalid, Issue/Pay invalid state (`handler_wave108_test.go`)
2. **Accounting service validation** — negative lines, unknown account, skip empty deal title (`service_wave108_test.go`)
3. **Search edge paths** — default limit, closed DB project lookup, empty scope files/projects (`service_wave108_test.go`)
4. **TMS handler errors** — invalid JSON, not-found case, invalid suite/component IDs (`handler_wave108_test.go`)

## Verification

```powershell
cd backend
go test -count=1 ./internal/accounting/... ./internal/tms/... ./internal/search/...
go test -coverprofile=cover108_all ./internal/...
go tool cover -func=cover108_all | Select-String total
```
