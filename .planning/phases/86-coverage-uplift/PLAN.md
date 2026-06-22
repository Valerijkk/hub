# Phase 86 — Search + Tracker + TMS + Notification Uplift

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

Поднять пакеты ниже 65%: `search`, `tracker`, `tms`, `notification`.

## Deliverables

### search (60.8% → ≥65%)
- `jql_predicates_extended_test.go` — status/priority/type/assignee/date predicates
- `service_shard_errors_test.go` — all-shard parallel search, channel scope errors

### tracker (63.2% → ≥65%)
- `handler_jql_flow_test.go` — JQL POST success + validation
- `automation_service_extra_test.go` — DeleteRule, validation, UpdateRule not found

### tms (61.0% → ≥62%)
- `ent_repository_list_update_test.go` — ListTestCases filters, UpdateTestCase all fields

### notification (63.2% → ≥63%)
- `handler_list_filter_test.go` — is_read filter + mark read flow

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/search/` | 60.8% | **66.1%** |
| `internal/tracker/` | 63.2% | **65.1%** |
| `internal/tms/` | 61.0% | **61.9%** |
| `internal/notification/` | 63.2% | **63.6%** |
| Overall `internal/` | 69.3% | **69.7%** |

## Success criteria

- [x] search ≥65%
- [x] tracker ≥65%
- [x] race PASS all packages
- [x] `/qa-gate` PASS

## Deferred

- `tms/` ≥65% (handler import flows)
- `notification/` ≥65% (push sendToSubscription live mock)
- Playwright E2E Wave 3
