# Phase 76 — Sprint Ent + Handler Routes Coverage

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

`internal/sprint/` **58.0% → ≥60%** — EntRepository + RegisterRoutes (RBAC)

## Deliverables

- `repository_ent_test.go` — CRUD, active sprint, issue moves, burndown, velocity
- `handler_routes_test.go` — RegisterRoutes, UpdateSprint, AddIssuesToSprint, RBAC list
- `backup/repository_test.go` — SetRestoredNote, Delete

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/sprint/` | 58.0% | **79.8%** |
| `internal/backup/` | 52.5% | **53.4%** |

## Success criteria

- [x] sprint ≥60%
- [x] EntRepository fully covered
- [x] RegisterRoutes + UpdateSprint covered
- [x] race PASS
