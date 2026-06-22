# Phase 97 — Mongo CI + Webhook Dispatcher + Marketplace

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

Mongo integration tests в CI, закрыть 0% webhook dispatcher, marketplace handlers, telegram sendMessage.

## Deliverables

### CI
- `.github/workflows/ci.yml` — MongoDB 8 service + `MONGODB_TEST_URL` для backend-test

### apigateway (67% → ≥80%)
- `webhook_dispatcher_wave97_test.go` — Start/Stop JetStream, handleMessage, deliverWithRetry, dispatchEvent skip, ListDeliveries

### marketplace (69% → ≥73%)
- `handler_wave97_test.go` — UpdateListing, DeleteListing, pagination query

### notification (73% → ≥74%)
- `telegram_wave97_test.go` — sendMessage ctx cancel + API err path

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/apigateway/` | 67.0% | **82.2%** |
| `internal/marketplace/` | 69.2% | **73.7%** |
| `internal/notification/` | 73.3% | **74.6%** |
| Overall `internal/` | 74.0% | **74.4%** |

## Success criteria

- [x] CI Mongo service configured
- [x] WebhookDispatcher Start/handleMessage/deliverWithRetry covered
- [x] marketplace Update/Delete handlers covered
- [x] telegram sendMessage 100%
- [x] race PASS on touched packages

## Deferred (Phase 98+)

- mongo_repository func coverage locally (needs running Mongo)
- push COV → 75%
- docker e2e-full local
- Commit + PR (on user request)
