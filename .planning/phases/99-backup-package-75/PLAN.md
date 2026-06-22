# Phase 99 — Backup + Package 75% Push

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

Довести **backup** (65.7%) и пакеты на грани 75% (**notification**, **tms**, **marketplace**) до целевого порога; укрепить overall COV.

## Deliverables

### backup (65.7% → 75.0%)
- `service_wave99_test.go` — fake pg_dump/mongodump via PATH, PerformBackup success + mongo fail, cleanup/markFailed error paths

### notification (74.6% → 75.0%)
- `consumer_wave99_test.go` — handleBlockerAdded skip self / invalid JSON
- `handler_wave99_test.go` — MarkAllRead/Delete unauthorized

### tms (74.7% → 75.3%)
- `allure_matcher_wave99_test.go` — automation_id label + case_key in name matching
- `handler_webhook_wave99_test.go` — multipart missing file
- `allure_parser_test.go` — mockMatcherRepo automation_id + caseKey maps

### marketplace (73.7% → 75.4%)
- `dto_wave99_test.go` — displayName, formatMinorUnits, isValidCategory

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/backup/` | 65.7% | **75.0%** |
| `internal/notification/` | 74.6% | **75.0%** |
| `internal/tms/` | 74.7% | **75.3%** |
| `internal/marketplace/` | 73.7% | **75.4%** |
| Overall `internal/` | 75.0% | **75.2%** |

## Success criteria

- [x] backup ≥75%
- [x] notification/tms/marketplace ≥75%
- [x] overall COV ≥75%
- [x] race PASS on touched packages

## Deferred (Phase 100+)

- RestoreBackup S3 happy path (mock GetObject + pg_restore)
- tracker/messenger/search → 75%
- Commit + PR (on user request)
