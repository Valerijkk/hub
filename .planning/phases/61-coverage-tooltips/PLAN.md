# Phase 61 — Coverage Uplift (project/search) + UX-P1-04 Tooltips

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-25

## Goal

1. Raise `project` and `search` packages to ≥50% coverage
2. UX-P1-04: contextual tooltips (JQL, WIP, Allure)
3. Fix corrupted UTF-8 in `jql-search-page.tsx`

## Deliverables

### Backend
- `backend/internal/project/repository_ent_test.go` — Ent repository CRUD, invites, teams
- `backend/internal/project/handler_flow_test.go` — Delete/Restore, UpdateMemberRole, RevokeInvite
- `backend/internal/search/handler_flow_test.go` — search + filter CRUD handlers
- `backend/internal/search/jql_test.go` — date field predicates via ToEntPredicates

### Frontend
- `frontend/src/components/help/context-help-tip.tsx` + test
- i18n `help.json` → `contextTips.*` (en/ru)
- Wired: JQL page, Kanban WIP count, TMS Allure webhook
- Fixed encoding: JQL subtitle + filter name placeholder

## Verification

```bash
cd backend && go test -count=1 ./internal/project/... ./internal/search/...
cd frontend && npx vitest run
```

## Success criteria

- [x] project ≥50% coverage
- [x] search ≥50% coverage
- [x] Vitest green (context-help-tip + existing suite)
- [x] UX-P1-04 tooltips visible on target surfaces
