# Phase 40 — Org Directory (MOD-02) — VERIFICATION

**Date:** 2026-07-27  
**Method:** Route/handler/FE/test/doc inspection (no live test run)  
**Sources:** `PLAN.md`, `backend/internal/intranet/handler.go`, FE routes/modules, `backend/testframework/intranet/`, `docs/wiki/API.md`

## Goal

Справочник сотрудников: группировка по отделам, карточки профиля, интеграция с Ctrl+K.

## Measured evidence

| Criterion | Evidence (2026-07-27) | Result |
|-----------|----------------------|--------|
| `GET /api/intranet/directory` | `handler.go` `intra.Get("/directory", h.ListDirectory)` | PASS |
| `GET .../directory/users/:userId` | `GetDirectoryUser` registered | PASS |
| FE `/intranet/directory` | `frontend/src/routes/_authenticated/intranet/directory.tsx` + `OrgDirectory` | PASS |
| Sidebar «Справочник» | `sidebar-nav.tsx` `nav.directory` / «Справочник» | PASS |
| Ctrl+K action | `command-palette.tsx` path `/intranet/directory` | PASS |
| Profile via `?user=` | Route passes `selectedUserId={user}` from search param | PASS |
| BE tests | `TestHandler_ListDirectory_GroupsByDepartment`, `GetDirectoryUser_NotFoundForPending` in `handler_test.go` | PASS |
| FE Vitest | `testframework/stages/unit/frontend-test/domains/v4/org-directory.test.tsx` | PASS |
| API.md Directory | `docs/wiki/API.md` § Directory (~L769) | PASS |

## Verdict: **PASS**
