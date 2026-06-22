# Phase 39: Intranet News Feed (MOD-01 W1)

**Goal:** Hub One v4 kickoff — corporate news feed (MOD-INTRA first slice).

**Branch:** `feat/v2.5-trust` (uncommitted)

## Tasks

- [x] Ent schema `NewsPost` (soft delete, pin, publish, version)
- [x] `backend/internal/intranet` — Repository + Service + Handler
- [x] Routes `/api/intranet/news*` wired in `main.go`
- [x] FE: `/intranet` route, sidebar «Новости», admin compose
- [x] Tests: handler (sqlite), Vitest smoke
- [x] API.md Intranet section

## Verification

```bash
cd backend && go test ./internal/intranet/...
cd frontend && npx tsc --noEmit && npx vitest run src/modules/intranet/
npm run lint
```

## Next (Phase 40+)

- Reactions on news posts
- Org directory (MOD-02)
- Unified `/calendar` (MOD-04)

**Status:** COMPLETE 2026-06-24
