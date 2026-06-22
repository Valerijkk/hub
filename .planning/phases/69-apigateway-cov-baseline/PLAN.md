# Phase 69 — Apigateway Coverage + COV Baseline Refresh

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-26

## Goal

1. **SEC-01 / API-01:** `apigateway/` **54% → ≥60%**
2. Refresh overall COV baseline post Phase 68

## Deliverables

- `handler_extended_test.go` — RegisterRoutes, RevokeAPIKey, UpdateEndpoint, ListDeliveries
- `repository_extra_test.go` — Count/Delete API key, webhook delivery log

## Results

| Metric | Before | After |
|--------|--------|-------|
| `apigateway/` | 54.0% | **67.0%** |
| Overall `internal/` | 57.9% | **58.1%** |

## Success criteria

- [x] apigateway ≥60%
- [x] all apigateway tests pass
- [x] overall COV refreshed
