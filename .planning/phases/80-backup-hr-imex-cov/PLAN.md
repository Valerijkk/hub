# Phase 80 — Backup + HR + IMEX Coverage Wave

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

- `internal/backup/` **53.4% → ≥60%**
- `internal/hr/` **53.1% → ≥60%**
- `internal/imex/` **57.8% → ≥60%**

## Deliverables

### Backup
- `coverage_wave80_test.go` — dump/restore exec paths, markFailed, RestoreBackup no-S3
- `handler_test.go` — DeleteBackup success, RestoreBackup confirm without S3

### HR
- `repository_ent_test.go` — ListLeaveByRequester, CancelLeave, ListApprovedLeaveInRange, ListSystemAdminIDs
- `service_extended_test.go` — ListMyLeave, CancelLeave, notifier/exporter hooks
- `handler_test.go` — ListMyLeave, CancelLeave, ListPending forbidden

### IMEX
- `pdf_export_test.go` — ExportRunPDF, ExportSprintPDF, PDF helpers, handler routes
- `repository_pdf_test.go` — ListTestRunResults, GetSprint, ListSprintIssues

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/backup/` | 53.4% | **65.7%** |
| `internal/hr/` | 53.1% | **73.3%** |
| `internal/imex/` | 57.8% | **77.7%** |
| Overall `internal/` | 65.4% | **66.2%** |

## Success criteria

- [x] backup ≥60%
- [x] hr ≥60%
- [x] imex ≥60%
- [x] race PASS all three packages
