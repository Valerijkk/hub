# Phase 91 — Intranet + Observability Uplift

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

Поднять `intranet/` и `platform/observability/` до ≥80% — coffee/news handler gaps и metrics allowlist.

## Deliverables

### intranet (68.7% → ≥80%)
- `handler_wave91_test.go` — ListNewsAdmin, UpdateNews, coffee GET endpoints, kudo/coffee error mapping
- `coffee_repository_test.go` — Upsert update path, GetPreviousRoundPairKeys, ListCoffeeRounds
- `dto_test.go` — toDirectoryUserResponse birthday, displayUserName fallbacks

### observability (74.5% → ≥80%)
- `metrics_allow_middleware_test.go` — MetricsAllowMiddleware empty/valid/invalid CIDR
- `otel_test.go` — skip /metrics path, traceparent extraction

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/intranet/` | 68.7% | **81.3%** |
| `platform/observability/` | 74.5% | **80.7%** |
| Overall `internal/` | 70.5% | **71.0%** |

## Success criteria

- [x] intranet ≥80%
- [x] observability ≥80%
- [x] race PASS intranet + observability
- [x] `/qa-gate` backend + frontend PASS

## Deferred (Phase 92)

- `messenger/` + `search/` uplift toward 70%
- Full docker e2e-full local run
- `imex/ImportIssuesJSON` field mapping edges
- Commit + PR (on user request)
