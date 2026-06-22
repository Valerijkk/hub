# Phase 123 — Messenger ≥82%, Storage ≥82%

> Branch: `feat/v2.5-trust` · Docker + network (Mongo for messenger) · `-tags hubtest`

## Goal

| Package | Baseline | Target |
|---------|----------|--------|
| messenger | 77.6% | ≥82% |
| storage | 86.8% | ✅ already ≥82% |

## Tasks

- [x] W1: `ent_repository_wave123_test.go` — mute durations, invites, display name, project members
- [x] W2: `service_wave123_test.go` — bot, mute, invites
- [x] W3: `handler_wave123_test.go` — HTTP mute/bookmark/invite success paths
- [x] W4: `mongo_repository_wave123_test.go` — edge cases (MONGODB_TEST_URL)
- [x] W5: Docker verify — messenger **82.3%**, storage **86.8%**

## Status: ✅ COMPLETE (2026-06-21)

## Verify

```bash
docker compose -f deployments/docker-compose.test.yml up -d
docker run --rm --network deployments_default \
  -v "${PWD}/backend:/app" -v hub-go-mod-cache:/go/pkg/mod -w /app \
  -e CGO_ENABLED=1 \
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" \
  golang:1.25-bookworm \
  sh -c "go test -tags hubtest -cover ./internal/messenger/... ./internal/storage/..."
```
