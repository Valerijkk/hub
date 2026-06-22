# Phase 65 — UX-P1-05 Rate Limit Timer + Messenger Coverage

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-25

## Goal

1. **UX-P1-05 / FB-03:** Live countdown toast on 429 + actionable `request_id` in error toasts  
2. **Coverage:** `messenger/` **48% → ≥50%** via `ent_repository` integration tests

## Results

| Metric | Before | After |
|--------|--------|-------|
| `messenger/` coverage | 48.0% | **52.3%** |
| Vitest (format-api-error) | 5 tests | **9 tests** |
| Full Vitest | 95/95 | **99/99** |

## Deliverables

### Frontend
- `format-api-error.ts`: live countdown (1s tick), `request_id` description
- i18n `apiErrors.requestIdHint` (en/ru)

### Backend
- `ent_repository_test.go`: channel CRUD, members, DM, star/mute, pins, bookmarks

## Success criteria

- [x] 429 toast counts down each second until dismiss
- [x] Non-429 errors show `request_id` when present
- [x] messenger ≥50%
- [x] Vitest green, `go test ./internal/messenger/...` green
