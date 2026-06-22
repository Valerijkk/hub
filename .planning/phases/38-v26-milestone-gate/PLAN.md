# Phase 38: v2.6 Milestone Gate

**Goal:** Close v2.6 milestone — docs sync, E2E/Vitest for Phase 36 alerts, qa-gate green, CI Biome strict.

**Branch:** `feat/v2.5-trust` (uncommitted)

## Tasks

- [x] `docs/wiki/API.md` — `/metrics`, `/admin/observability/ops`, `/webhooks/alertmanager`, audit alert filters
- [x] Vitest: alert audit entry rendering (`audit-log.test.tsx`)
- [x] E2E: admin audit alert filter options (`admin-audit.spec.ts`)
- [x] CI: Biome job renamed to gate (no continue-on-error)
- [x] Fix race: `ai/service_test.go` mockStreamPublisher mutex
- [x] qa-gate: build + race + tsc + lint + vitest

## Verification

```bash
cd backend && go build -tags release ./... && go test -race -count=1 ./...
cd frontend && npm run lint && npx tsc --noEmit && npx vitest run
```

**Status:** COMPLETE 2026-06-24
