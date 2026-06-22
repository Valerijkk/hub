# Phase 95 — Notification Push/TG + Tracker Automation + TMS Export

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

Поднять `notification/`, `tracker/`, `tms/` до ≥70% через push/telegram bind, automation consumer, issue message links, exporter.

## Deliverables

### notification (68.1% → ≥70%)
- `push_wave95_test.go` — sendToSubscription err path + unsafe endpoint delete
- `telegram_bind_wave95_test.go` — handleMessage (/start bind, /link, /unlink, /test)

### tracker (65.1% → ≥70%)
- `automation_consumer_test.go` — parseIssueEventKind, Start/Stop JetStream, processMessage events
- `handler_wave95_test.go` — issue message links CRUD, ToggleReaction, DeleteCustomField
- `service_issue_message_link_test.go` — Create/List/Delete service paths
- `service_test.go` — mockRepo IssueMessageLink implementations

### tms (65.6% → ≥70%)
- `exporter_wave95_test.go` — Excel/CSV/PDF export, truncate, service export paths
- `handler_import_flow_test.go` — excel + pdf/csv run export handlers
- ImportFromJSON/XML + UpdateTestResult service tests

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/notification/` | 68.1% | **73.3%** |
| `internal/tracker/` | 65.1% | **72.4%** |
| `internal/tms/` | 65.6% | **70.0%** |
| Overall `internal/` | 72.0% | **73.4%** |

## Success criteria

- [x] notification ≥70%
- [x] tracker ≥70%
- [x] tms ≥70%
- [x] automation_consumer 0% → covered
- [x] handler_issue_message_link 0% → covered
- [x] exporter Excel/PDF/CSV covered
- [x] race PASS on touched packages

## Deferred (Phase 96+)

- mongo_repository tms (0%, needs Mongo testcontainer)
- testrail_import paths (0%)
- docker e2e-full local
- Commit + PR (on user request)
