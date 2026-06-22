# Phase 98 — Accounting + Calendar + COV 75%

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

Закрыть gaps **accounting** (65%) и **calendar** (65.6%); довести overall `internal/` COV до **≥75%** (deferred из Phase 97).

## Deliverables

### accounting (65% → ≥75%)
- `handler_wave98_test.go` — CreateAccount, ListJournal, ListInvoices, PostJournal 409, TrialBalance, IncomeStatement, mapServiceError
- `service_wave98_test.go` — parseEntryDate, validateLines unbalanced via CreateJournal

### calendar (65.6% → ≥85%)
- `service_wave98_test.go` — expandCallEvents (one-off + RRULE), parseRRule, leaveEvent, validateRange, handler invalid/missing params, range too wide

### backup / bases (COV push)
- `backup/service_wave98_test.go` — extractZIP pg/mongo/unknown entries
- `bases/service_wave98_test.go` — normalizeCells, mapServiceError, DeleteRecord, CreateBase invalid
- `bases/repository_wave98_test.go` — UpdateTable, SoftDeleteTable, UpdateColumn, SoftDeleteColumn

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/accounting/` | 65.0% | **76.5%** |
| `internal/calendar/` | 65.6% | **86.7%** |
| `internal/bases/` | 71.5% | **78.1%** |
| `internal/backup/` | 65.7% | **65.7%** (extractZIP only) |
| Overall `internal/` | 74.4% | **75.0%** |

## Success criteria

- [x] accounting ≥75%
- [x] calendar ≥75%
- [x] overall COV ≥75%
- [x] race PASS on touched packages

## Deferred (Phase 99+)

- backup ≥75% (PerformBackup S3 path needs mock)
- notification/marketplace/tms → 75% per package
- mongo_repository func coverage locally
- Commit + PR (on user request)
