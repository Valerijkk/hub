# Phase 83 — AI Bot Commands + Search Query

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

- Complete `@hub-ai` slash commands (test-cases, review, explain, translate, sql)
- `internal/search/` **60.5% → ≥60%** (query paths)
- `internal/ai/` **75.0% → ≥76%**

## Deliverables

### AI bot commands
- `messenger_bot_flow_test.go` — test-cases success/not found, review success/error, explain/translate/sql, docs error, stream AI error
- `service_test.go` — `mockAdapter.ChatCompletionStream` respects `err`
- `code_review_test.go` — nil PR/MR converters, no-VCS FetchReview

### Search query paths
- `service_query_test.go` — empty projects, type filter, limit clamp, message without mongo, project lookup error

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/ai/` | 75.0% | **76.4%** |
| `handleCommand` | 55.6% | **84.4%** |
| `streamReply` | 84.2% | **94.7%** |
| `internal/search/` | 60.5% | **60.8%** |
| Overall `internal/` | 68.3% | **68.4%** |

## Success criteria

- [x] All `/ai` command branches covered (except live VCS fetch)
- [x] Search Search() orchestration paths
- [x] race PASS ai + search

## Deferred

- `UnifiedCodeReviewProvider.FetchReview` live GitHub/GitLab HTTP
- `OnMessageSent` no-model path edge cases
