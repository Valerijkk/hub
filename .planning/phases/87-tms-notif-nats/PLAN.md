# Phase 87 — TMS + Notification + NATS Uplift

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

Поднять покрытие `tms`, `notification`, `platform/nats` — закрыть пробелы в consumer events, ent batch repo и NATS publish.

## Deliverables

### notification (63.6% → ≥63.5%)
- `consumer_extra_test.go` — `TestIsRunFailedRecipient` (NOTIF-04 role filter)
- Consolidated duplicate messenger consumer tests (removed `consumer_messenger_test.go`)

### tms (61.9% → ≥62%)
- `ent_repository_batch_test.go` — `BatchCreateTestResults` all optional fields
- `ListTestRuns` status filter
- `GetCoverageReport` via `NewCombinedRepository` + ent integration

### platform/nats (61.5% → ≥65%)
- `client_publish_test.go` — embedded NATS: `PublishJSON`, `Publish`, `EnsureStreams` idempotent

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/tms/` | 61.9% | **62.4%** |
| `internal/notification/` | 63.6% | **63.7%** |
| `internal/platform/nats/` | 61.5% | **80.8%** |
| Overall `internal/` | 69.7% | **69.8%** |

## Success criteria

- [x] nats ≥65%
- [x] tms ≥62%
- [x] notification stable (+0.1%)
- [x] race PASS all packages
- [x] `/qa-gate` PASS (backend)

## Deferred (Phase 88)

- `tms/` ≥65% — handler webhooks, import flows
- `notification/` ≥65% — push `sendToSubscription` (webpush mock)
- Playwright E2E Wave 3
