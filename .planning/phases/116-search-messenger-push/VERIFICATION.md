# Phase 116 Verification

**Date:** 2026-06-22  
**Branch:** `feat/v2.5-trust`  
**Goal:** COV ≥81%  
**Result:** **81.6%** — goal **ACHIEVED** (+0.9pp from 80.7%)

## Commands

```powershell
docker compose -f deployments/docker-compose.test.yml up -d

docker run --rm --network deployments_default `
  -v "${PWD}/backend:/app" -v hub-go-mod-cache:/go/pkg/mod -w /app `
  -e CGO_ENABLED=1 `
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" `
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" `
  golang:1.25-bookworm `
  sh -c "go test -race -count=1 -timeout=600s ./internal/... && go test -coverprofile=cov116_all ./internal/... && go tool cover -func=cov116_all | tail -1"
```

## Results

| Check | Status |
|-------|--------|
| `go test -race -count=1 ./internal/...` (Docker) | **PASS** |
| Overall COV | **81.6%** (was 80.7%) |
| Search PG integration | **PASS** (5 tests) |

## Package deltas

| Package | Phase 115 | Phase 116 |
|---------|-----------|-----------|
| search | ~76% | **92.2%** |
| messenger | ~77% | **81.7%** |
| tms | ~79% | **80.3%** |
| notification | ~78.5% | **78.7%** |
| ai | ~79% | **80.5%** |
| bases | ~80% | **82.7%** |
| storage | 79.6% | 79.6% |
| tracker | ~79% | 79.1% |

## Goal-backward

- **Achieved:** 81% COV target; PostgreSQL tsvector integration for all search shards; messenger channel RBAC service paths; `searchFiles` UUID scan bugfix.
- **Side effect fix:** `searchFiles` returned empty results in production when PG UUID columns were scanned into `*string`.
- **Recommendation:** Phase 117 — storage/notification/tracker push + frontend Vitest Wave 3.

## Artifacts

- `.planning/phases/116-search-messenger-push/PLAN.md`
- 7 new `*_wave116_test.go` files + `search/service.go` scan fix
