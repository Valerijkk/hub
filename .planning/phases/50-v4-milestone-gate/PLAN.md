# Phase 50: v4 Super-app Milestone Gate

**Goal:** Close Hub One v4 block (Phases 40–49) — qa-gate green, Vitest smoke for new modules, docs sync.

**Branch:** `feat/v2.5-trust` (uncommitted)

## Tasks

- [x] Vitest: `bases-page.test.tsx`, `onec-settings.test.tsx`
- [x] Biome lint 0 errors (v4 FE files formatted)
- [x] qa-gate: `go build -tags release`, `go vet`, `go test -race ./...`
- [x] Frontend: `tsc`, `lint`, `vitest run` (79 tests)
- [x] API docs already synced in Phases 40–49
- [x] `VERIFICATION.md` gate report

## v4 Module Matrix (Phases 40–49)

| Phase | Module | BE tests | FE Vitest |
|-------|--------|----------|-----------|
| 40 | Org directory | intranet | org-directory.test |
| 41 | Calendar | calendar | unified-calendar.test |
| 42 | CRM | crm | crm-page.test |
| 43 | HR | hr | leave-requests.test |
| 44 | Kudos | intranet | news-feed.test |
| 45 | Marketplace | marketplace | marketplace-page.test |
| 46 | Status | status | status-page.test |
| 47 | Coffee | intranet | random-coffee.test |
| 48 | 1C | onec | onec-settings.test |
| 49 | Lark Base | bases | bases-page.test |

## Verification

```bash
cd backend && go build -tags release ./... && go test -race -count=1 ./...
cd frontend && npm run lint && npx tsc --noEmit && npx vitest run
```

**Status:** COMPLETE 2026-06-25

## Next

- Commit `feat/v2.5-trust` phases 40–50 (single squash PR recommended)
- PLAN-CLOSE-GAPS Wave 2–3: coverage ≥50%, Playwright E2E
- Phase 51+ : V4-03 full accounting (large)
