# Phase 45 — Internal marketplace (MOD-08) — VERIFICATION

**Date:** 2026-07-27  
**Method:** Schema/handler/FE/test/doc inspection (no live test run)  
**Sources:** `PLAN.md`, `backend/ent/schema/marketplacelisting.go`, `backend/internal/marketplace/`, FE, `backend/testframework/marketplace/`, `docs/wiki/API.md`

## Goal

Internal listings with admin/founder moderation, public feed, notifications (mod queue + author).

## Measured evidence

| Criterion | Evidence (2026-07-27) | Result |
|-----------|----------------------|--------|
| Ent `MarketplaceListing` | `marketplacelisting.go`: `version=true`, `deleted_at=true` | PASS |
| CRUD + approve/reject/sold/withdraw | `handler.go` listings Get/Post/Put + approve/reject/sold/withdraw/Delete | PASS |
| Module + notifier adapters | `backend/internal/marketplace/` present (service/handler; notifier wiring cited in PLAN — files present under module) | PASS |
| FE `/marketplace` | `frontend/src/routes/_authenticated/marketplace.tsx` + `frontend/src/modules/marketplace/` | PASS |
| Sidebar + Ctrl+K | `sidebar-nav.tsx` «Барахолка»; command palette `/marketplace` | PASS |
| BE tests | `backend/testframework/marketplace/` — 7 test files | PASS |
| FE Vitest | `testframework/stages/unit/frontend-test/domains/v4/marketplace-page.test.tsx` | PASS |
| API.md Marketplace | `docs/wiki/API.md` § Marketplace (~L872) | PASS |

## Notes

- Optional `#marketplace` channel broadcaster not line-audited end-to-end this session; PLAN marked it optional.

## Verdict: **PASS**
