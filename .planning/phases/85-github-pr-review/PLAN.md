# Phase 85 — GitHub PR Review + Coverage Uplift

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

- `UnifiedCodeReviewProvider.FetchReview` **GitHub** path via mock HTTP
- `FetchPRForReview` httptest coverage
- Quiet-hours cross-midnight branch

## Deliverables

### GitHub HTTP injection
- `service.go` — `NewGitHubServiceWithHTTPClient`, `apiClient()` helper
- `pr_review.go` — uses injectable client + API base URL

### Tests
- `pr_review_test.go` — success, no integration, files error partial metadata
- `code_review_unified_test.go` — GitHub FetchReview + bot `/ai review` E2E
- `service_extended_test.go` — `isQuietHours` cross-midnight

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/github/` | 67.8% | **72.6%** |
| `internal/ai/` | 76.8% | **77.0%** |
| Overall `internal/` | 69.1% | **69.3%** |

## Success criteria

- [x] GitHub FetchPRForReview httptest
- [x] AI Unified FetchReview GitHub routing
- [x] Bot `/ai review` GitHub E2E
- [x] race PASS touched packages
- [x] `/qa-gate` PASS

## Deferred

- `search/` ≥65% (needs tsvector fixtures)
- `notification/` ≥65% (consumer NATS paths)
- `tms/` / `tracker` handler gaps
