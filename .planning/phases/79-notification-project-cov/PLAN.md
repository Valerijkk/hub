# Phase 79 — Notification + Project Coverage Wave

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

- `internal/notification/` **55.7% → ≥60%**
- `internal/project/` **59.4% → ≥60%**

## Deliverables

### Project
- `repository_ent_extended_test.go` — FindByID, template CRUD, cross-project links, GetProjectWithWorkflowStatusesAndLabels
- `template_handler_routes_test.go` — TemplateHandler RegisterRoutes (list/delete/export/import)
- `template_export_ent_test.go` — ExportTemplate/ImportTemplate with Ent repo
- `provisioner_test.go` — Set*Provisioner + provisionProject paths
- `template_multipart_test.go` — multipart helper for import route

### Notification
- `telegram_lifecycle_test.go` — TelegramSender nil-safe, Send, sendTelegram service paths
- `coverage_wave79_test.go` — processEvent, CheckDeadlines, Dispatch+publishToWS (embedded NATS)
- `telegram_bind_test.go` — SendTest linked/not-linked

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/notification/` | 55.7% | **63.2%** |
| `internal/project/` | 59.4% | **71.7%** |
| Overall `internal/` | 64.5% | **65.4%** |

## Success criteria

- [x] notification ≥60%
- [x] project ≥60%
- [x] race PASS both packages
- [x] no production code changes (tests only)
