# Phase 117 Verification

**Date:** 2026-06-22  
**Branch:** `feat/v2.5-trust`  
**Goal:** storage/notification/tracker push + COV ≥81%  
**Result:** **81.8%** overall — **ACHIEVED** (+0.2pp from 81.6%)

## Commands

```powershell
docker run --rm --network deployments_default `
  -v "${PWD}/backend:/app" -v hub-go-mod-cache:/go/pkg/mod -w /app `
  -e CGO_ENABLED=1 `
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" `
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" `
  golang:1.25-bookworm `
  sh -c "go test -race -count=1 -timeout=600s ./internal/... && go test -coverprofile=cov117_all ./internal/... && go tool cover -func=cov117_all | tail -1"
```

## Results

| Check | Status |
|-------|--------|
| `go test -race -count=1 ./internal/...` (Docker) | **PASS** |
| Overall COV | **81.8%** (was 81.6%) |
| storage RBAC integration | **PASS** |

## Package deltas

| Package | Phase 116 | Phase 117 |
|---------|-----------|-----------|
| storage | 79.6% | **86.8%** |
| notification | 78.7% | **79.2%** |
| tracker | 79.1% | **79.5%** |
| search | 92.2% | 92.2% |
| messenger | 81.7% | 81.7% |

## Goal-backward

- **Achieved:** storage `authorizeProject` RBAC paths (viewer/stranger); orphan S3 branch; notification quiet-hours dispatch; tracker time-tracking not-found paths.
- **Missed:** notification/tracker still below 82% — Phase 118 target.
- **Recommendation:** Phase 118 — notification consumer edges + tracker workflow handler batch.

## Artifacts

- `.planning/phases/117-storage-notification-tracker/PLAN.md`
- 6 new `*_wave117_test.go` files under `backend/internal/`
