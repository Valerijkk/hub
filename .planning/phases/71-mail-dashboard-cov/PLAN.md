# Phase 71 — Mail + Dashboard Coverage (MAIL-01 / BE-H3)

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

1. **MAIL-01 / Phase 15.2:** `mail/` **44.6% → ≥50%** — hub-mail handlers + admin API
2. **BE-H3:** `dashboard/` project repo + `Project` handler **61.7% → ≥65%**
3. Refresh overall COV baseline

## Deliverables

### Mail (`internal/mail/`)
- `hub_handler_test.go` — RegisterRoutes, Health, Status, Send auth/validation
- `admin_handler_test.go` — RegisterRoutes, settings CRUD, broadcast dry-run, RBAC
- `config_test.go` — LoadConfig, getEnv
- `admin_service_test.go` — Broadcast service (dry-run + send)

### Dashboard (`internal/dashboard/`)
- `repository_extra_test.go` — project-scoped issue/test-run/activity queries
- `handler_test.go` — `Project` handler success path

## Results

| Metric | Before | After |
|--------|--------|-------|
| `mail/` | 44.6% | **86.0%** |
| `dashboard/` | 61.7% | **73.7%** |
| Overall `internal/` | 59.2% | **59.8%** |

## Success criteria

- [x] mail ≥50%
- [x] dashboard ≥65%
- [x] hub-mail RegisterRoutes covered
- [x] admin/mail RegisterRoutes covered
- [x] all mail/dashboard tests pass
