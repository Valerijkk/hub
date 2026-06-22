# Phase 84 — VCS Review + Bases + Intranet Coffee

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

- `UnifiedCodeReviewProvider.FetchReview` GitLab path via mock HTTP
- `internal/bases/` **60.8% → ≥65%**
- `internal/gitlab/` MR review coverage uplift
- `internal/intranet/` coffee service paths

## Deliverables

### GitLab MR review
- `service.go` — `NewGitLabServiceWithHTTPClient` for test injection
- `mr_review_test.go` — FetchMRForReview success, invalid URL, no integration

### AI unified review
- `code_review_unified_test.go` — FetchReview GitLab + bot `/ai review` E2E with httptest

### Bases handler flow
- `handler_flow_test.go` — GetBase, CreateTable, CreateColumn, UpdateRecord, version conflict, invalid UUID

### Intranet coffee
- `coffee_service_test.go` — enrollment, matching, stats, rounds

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/ai/` | 76.4% | **76.8%** |
| `internal/bases/` | 60.8% | **71.5%** |
| `internal/gitlab/` | 68.5% | **79.1%** |
| `internal/intranet/` | 61.2% | **68.7%** |
| Overall `internal/` | 68.4% | **69.1%** |

## Success criteria

- [x] GitLab FetchMRForReview httptest
- [x] AI Unified FetchReview GitLab routing
- [x] bases ≥65%
- [x] race PASS touched packages

## Deferred

- GitHub `FetchPRForReview` live mock (go-github client)
- nats `PublishJSON` with embedded server
