# Phase 96 — TMS TestRail/Mongo + Excel Import

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

Закрыть deferred Phase 95: `mongo_repository`, `testrail_import`, `ImportFromExcel`; подготовить messenger mongo integration tests.

## Deliverables

### tms (70.0% → ≥74%)
- `mongo_repository_test.go` — Allure results CRUD + indexes (skip без MongoDB)
- `testrail_import_wave96_test.go` — CSV/XML parse, service import, handler multipart
- `importer_wave96_test.go` — ImportFromExcel via excelize

### messenger (prep)
- `mongo_repository_test.go` — message lifecycle, reactions, threads, search (skip без MongoDB)

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/tms/` | 70.0% | **74.7%** |
| Overall `internal/` | 73.4% | **74.0%** |
| `testrail_import.go` | ~0% | **~80%+** |
| `ImportFromExcel` | 0% | **93.1%** |

## Success criteria

- [x] testrail CSV/XML/importTestRailCases covered
- [x] ImportTestRailCases handler covered
- [x] ImportFromExcel covered
- [x] mongo_repository tests (integration, skip-safe)
- [x] messenger mongo tests (integration, skip-safe)
- [x] race PASS on tms + messenger

## Deferred (Phase 97+)

- mongo tests in CI (MONGODB_TEST_URL / docker-compose)
- messenger mongo 0% until CI Mongo available
- docker e2e-full local
- Commit + PR (on user request)
