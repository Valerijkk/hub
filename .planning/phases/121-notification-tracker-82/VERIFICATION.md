# Phase 121 Verification

> Date: 2026-06-22 · Docker · `-tags hubtest`

## Goals

| Target | Goal | Actual | Status |
|--------|------|--------|--------|
| notification | ≥82% | **81.9%** (coverpkg via testframework) | ⚠️ -0.1% |
| tracker | ≥82% | **81.3%** | ⚠️ -0.7% |
| TF-M02 notification migration | 100% | **36 files**, 0 in internal | ✅ |
| Docker race full suite | PASS | PASS | ✅ |

## Commands

```powershell
docker compose -f deployments/docker-compose.test.yml up -d
docker run --rm --network deployments_default `
  -v "${PWD}/backend:/app" -v hub-go-mod-cache:/go/pkg/mod -w /app `
  -e CGO_ENABLED=1 `
  golang:1.25-bookworm `
  sh -c "go test -tags hubtest -race -count=1 -timeout=600s ./internal/... ./testframework/..."
```

Coverage (notification via external tests):

```bash
go test -tags hubtest -coverprofile=cov -coverpkg=./internal/notification/... ./testframework/notification/...
go tool cover -func=cov | tail -1
```

## Artifacts

- `backend/internal/notification/export_hooks.go` — `//go:build hubtest` export API
- `backend/testframework/notification/` — 36 migrated test files
- `scripts/migrate-go-tests.ps1` — batch migration script
- `.github/workflows/ci.yml` — `go test -tags hubtest`

## Remaining for Phase 121 close

1. +0.1% notification (Consumer.Start with NATS mock or telegram bind edge)
2. +0.7% tracker (board column routes, automation consumer)
3. Begin TF-M03 auth/ws/rbac/audit migration
