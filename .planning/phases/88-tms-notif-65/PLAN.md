# Phase 88 — TMS + Notification ≥65%

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

Довести `tms` и `notification` до **≥65%** statement coverage.

## Deliverables

### tms (62.4% → ≥65%)
- `handler_flow_test.go` — routes: trends, rollback, bug, env update/delete
- `handler_wave88_test.go` — RollbackTestCase, CreateBugFromResult, env CRUD, comparison validation
- `service_reports_test.go` — GetRunComparison, GetFlakyTests, CompareEnvironments
- `allure_parser_test.go` — MatchStrategy.String()
- `service_test.go` — mock env Update/Delete, runResults-scoped ListTestResults

### notification (63.7% → ≥65%)
- `coverage_wave79_test.go` — processEvent status/comment/new_login
- `handler_telegram_bind_test.go` — TelegramConfirm success
- `service_extended_test.go` — sendToUser unsafe endpoint removal (SEC-H3)

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/tms/` | 62.4% | **65.6%** |
| `internal/notification/` | 63.7% | **65.7%** |
| Overall `internal/` | 69.8% | **70.2%** |

## Success criteria

- [x] tms ≥65%
- [x] notification ≥65%
- [x] race PASS
- [x] `/qa-gate` PASS

## Deferred

- `notification/` push live webpush mock (sendToSubscription network path)
- `tms/` mongo_repository (requires testcontainers)
- Playwright E2E Wave 3
