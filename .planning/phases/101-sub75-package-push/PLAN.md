# Phase 101 — Sub-75% Package Push

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

Довести пакеты ниже 75% (**project**, **github**, **crm**, **hr**, **platform/outbox**, **platform/errors**, **integrations/onec**) до порога ≥75%; поднять overall COV.

## Deliverables

### project (71.7% → 75.4%)
- `service_wave101_test.go` — PermanentDelete, cross-project link cleanup + NATS, AddTeamToProject
- `template_handler_wave101_test.go` — SaveAsTemplate handler + bad project_id

### github (72.6% → 75.0%)
- `handler_wave101_test.go` — toWorkflowRunResponse, CreateBranch bad JSON, GetLinkPreview missing url
- `repository_wave101_test.go` — UpdateIssueStatus, GetLinkPreview handler with preview service

### crm (72.8% → 76.7%)
- `handler_wave101_test.go` — ListContacts, mapServiceError, update/delete paths
- `service_wave101_test.go` — DeleteContact, CreateDeal validation

### hr (73.3% → 77.9%)
- `handler_wave101_test.go` — mapServiceError, CancelLeave version-zero/forbidden, approve not-found

### platform/outbox (74.3% → 88.6%)
- `worker_wave101_test.go` — publish fail increments attempts, closed client batch

### platform/errors (72.7% → 100%)
- `errors_wave101_test.go` — NewRateLimitedAfter + minimum RetryAfter

### integrations/onec (72.0% → 83.4%)
- `settings_store_wave101_test.go` — EntSettingsStore Get/Upsert, EntUserLookup

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/project/` | 71.7% | **75.4%** |
| `internal/github/` | 72.6% | **75.0%** |
| `internal/crm/` | 72.8% | **76.7%** |
| `internal/hr/` | 73.3% | **77.9%** |
| `internal/platform/outbox/` | 74.3% | **88.6%** |
| `internal/platform/errors/` | 72.7% | **100%** |
| `internal/integrations/onec/` | 72.0% | **83.4%** |
| Overall `internal/` | 75.8% | **76.3%** |

## Success criteria

- [x] All target packages ≥75%
- [x] overall COV ≥75%
- [x] race PASS on touched packages

## Deferred (Phase 102+)

- `internal/dev/` (0% — dev-only routes)
- `messenger/mongo_repository`, `search/searchMessages` (Mongo)
- Commit + PR (on user request)
