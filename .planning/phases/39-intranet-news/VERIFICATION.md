# Phase 39 — Intranet News Feed (MOD-01) — VERIFICATION

**Date:** 2026-07-27  
**Method:** Schema/module/route/test/doc existence + trash-type grep (no live test run)  
**Sources:** `PLAN.md`, `backend/ent/schema/newspost.go`, `backend/internal/intranet/`, `backend/testframework/intranet/`, FE + `docs/wiki/API.md`

## Goal

Hub One v4 kickoff — corporate news feed (MOD-INTRA first slice).

## Measured evidence

| Criterion | Evidence (2026-07-27) | Result |
|-----------|----------------------|--------|
| Ent `NewsPost` (soft delete, pin, publish, version) | `backend/ent/schema/newspost.go`: `version=true`, `deleted_at=true` | PASS |
| `internal/intranet` Repository + Service + Handler | Module dir present; `handler.go` registers `/news` CRUD (+ admin list) | PASS |
| Routes wired | `news.Get("/")`, `Post("/")`, `Get/Put/Delete /:newsId`, admin list | PASS |
| FE `/intranet` + news UI | `frontend/src/modules/intranet/news-feed.tsx`; sidebar «Новости» in `sidebar-nav.tsx`; route under `_authenticated/intranet` | PASS |
| Backend tests | `backend/testframework/intranet/handler_test.go` (`TestHandler_CreateListNews_AdminOnly`, `DeleteNews`, …) + service tests | PASS |
| FE Vitest | `testframework/stages/unit/frontend-test/domains/v4/news-feed.test.tsx` (migrated from `src/modules/intranet/`) | PASS |
| API.md Intranet News | `docs/wiki/API.md` § Intranet / News (~L754–758) | PASS |

## Residual gaps (out of original PLAN checklist; tracked elsewhere)

| Gap | Evidence | Owner |
|-----|----------|-------|
| FR-INTR-08 trash restore | Soft-delete exists; `audit/service.go` trash `case` arms = `issue` / `test_case` / `file` / `project` only — **no `newspost`** | Phase `126-intranet-newspost-trash` |

## Verdict: **PASS**

Phase-39 news-feed slice is present in code/docs/tests. Trash restore is a confirmed later gap, not a missing Phase-39 deliverable checkbox.
