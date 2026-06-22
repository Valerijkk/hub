# Phase 125 — COV ≥82%, Security ≥80%, Regression Gate

> Branch: `feat/v2.5-trust` · Docker only · `-tags hubtest`

## Goal

| Metric | Target | Result |
|--------|--------|--------|
| COV overall | ≥82% | ✅ **82.6%** |
| auth | ≥80% | ✅ **82.4%** |
| ws | ≥80% | ✅ **87.7%** |
| rbac | ≥80% | ✅ **84.7%** |
| audit | ≥80% | ✅ **80.5%** |
| notification | ≥82% | ✅ **82.5%** (testframework) |
| Full regression | pass | ✅ `go test -tags hubtest -race ./internal/... ./testframework/...` |

## Verify

```powershell
docker compose -f deployments/docker-compose.test.yml up -d
docker run --rm --network deployments_default `
  -v "${PWD}/backend:/app" -v hub-go-mod-cache:/go/pkg/mod -w /app `
  -e CGO_ENABLED=1 `
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" `
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" `
  golang:1.25-bookworm `
  sh -c "go test -tags hubtest -race -count=1 -timeout=600s ./internal/... ./testframework/... && go test -tags hubtest -coverprofile=/tmp/cov -coverpkg=./internal/... ./internal/... ./testframework/... && go tool cover -func=/tmp/cov | tail -1"
```

## Status: ✅ COMPLETE (2026-06-21)
