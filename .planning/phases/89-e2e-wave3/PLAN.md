# Phase 89 — E2E Wave 3 + Dashboard Uplift

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

Закрыть **E2E-01** (5 critical Playwright paths) и поднять `dashboard/` coverage.

## Deliverables

### E2E Wave 3 (PLAN-CLOSE-GAPS Stage 3.2)
- `frontend/e2e/critical-gate.spec.ts` — 5 paths: login, dev login, navigation, Ctrl+K, API health
- Playwright Chromium installed locally; smoke **3/3 PASS** (login + a11y)
- CI: `e2e-smoke` + `e2e-full` include `critical-gate.spec.ts`

### dashboard (73.7% → ≥75%)
- `service_extended_test.go` — `GetProjectStats_ComputesFromDB` + Redis cache write
- `handler_test.go` — `Project` stats endpoint, `ApplyTemplate` success

## Results

| Metric | Before | After |
|--------|--------|-------|
| Playwright E2E specs | 19 | **20** (+critical-gate) |
| E2E-01 critical paths | implicit | **explicit 5** |
| `internal/dashboard/` | 73.7% | **77.6%** |
| Overall `internal/` | 70.2% | **70.3%** |

## Success criteria

- [x] E2E-01 five paths documented and runnable
- [x] Playwright smoke PASS (frontend-only)
- [x] dashboard ≥75%
- [x] `/qa-gate` backend PASS

## Deferred (Phase 90)

- Full `e2e-full` with docker backend (local/CI infra)
- `imex/` ImportJira handler flows
- Vitest gaps in hooks/stores
