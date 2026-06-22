# Phase 100 — RestoreBackup + Tracker/Messenger/Search 75%

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

Закрыть **RestoreBackup** S3 happy path; довести **tracker**, **messenger**, **search** до ≥75%; удержать overall COV ≥75%.

## Deliverables

### backup (75.0% → 85.8%)
- `service_wave100_test.go` — fake S3 via `httptest` + aws-sdk `BaseEndpoint`; `PerformBackup` upload; full `RestoreBackup` from zip; GetObject fail path

### tracker (73.2% → 75.0%)
- `repository_wave100_test.go` — outbox on UpdateIssue, automation DTO/handler, board assignee+label filters
- `automation_engine_wave100_test.go` — invalid project/channel, custom template, notify edges, nil deps
- `service_wave100_test.go` — LinkThreadToIssue, issue↔message link CRUD

### messenger (74.1% → 75.1%)
- `presence_wave100_test.go` — ParseStatusDuration, GetOnlineUserIDs, cleanup loop, MessengerService custom status
- `handler_wave100_test.go` — SetCustomStatus, RemoveBookmark

### search (72.7% → 75.1%)
- `service_wave100_test.go` — scoped query fail paths, uuidSliceToPgArray
- `handler_wave100_test.go` — shareFilter errors, getUserID, all-shard Search
- `repository_wave100_test.go` — founder admin, shared filters, handler 500 paths
- `jql_wave100_test.go` — ORDER BY, AND/OR chain, IN lists
- `meilisearch_indexer_test.go` — IndexAllIssues Ent error path

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/backup/` | 75.0% | **85.8%** |
| `internal/tracker/` | 73.2% | **75.0%** |
| `internal/messenger/` | 74.1% | **75.1%** |
| `internal/search/` | 72.7% | **75.1%** |
| Overall `internal/` | 75.54% | **75.8%** |

## Success criteria

- [x] RestoreBackup S3 + zip restore path covered
- [x] tracker/messenger/search ≥75%
- [x] overall COV ≥75%
- [x] race PASS on touched packages

## Deferred (Phase 101+)

- Sub-75% packages: project, github, crm, hr, platform/outbox, platform/errors
- `messenger/mongo_repository` (requires MONGODB_TEST_URL)
- Commit + PR (on user request)
