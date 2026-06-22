# Phase 77 — Storage Handler Routes (CR-04 / Presign)

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

`internal/storage/` **55.7% → ≥60%** — NewHandler, RegisterRoutes, RBAC authorizeProject, NewStorageService

## Deliverables

- `handler_routes_test.go` — full upload/confirm/download/delete via RegisterRoutes + RBAC
- `TestNewStorageService` — config wiring

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/storage/` | 55.7% | **79.6%** |
| Overall COV | 62.6% | **62.8%** |

## Success criteria

- [x] storage ≥60%
- [x] RegisterRoutes covered
- [x] CR-04 project-scoped auth paths covered
- [x] race PASS
