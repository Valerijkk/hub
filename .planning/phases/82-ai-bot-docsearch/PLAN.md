# Phase 82 — AI Bot Hook + Doc Search RAG

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

- Close Phase 81 deferred: `@hub-ai` messenger bot hook + `SearchDocProvider` full path
- `internal/ai/` **68.5% → ≥70%**

## Deliverables

### AI bot integration
- `messenger/hookfixture.go` — `HookTestRepo` for hook tests
- `messenger/mock_repository.go` — extracted in-memory repo (shared with unit tests)
- `ai/messenger_bot_flow_test.go` — OnMessageSent, streamReply, loadHistory, handleCommand (help/docs/summarize/review/errors)

### Doc search RAG-lite
- `doc_search.go` — `hubSearchBackend` interface (testable stub)
- `doc_search` tests — results, empty, search error

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/ai/` | 68.5% | **75.0%** |
| Overall `internal/` | 67.6% | **68.3%** |
| `messenger_bot.go` OnMessageSent | 0% | **79.3%** |
| `doc_search.go` SearchContext | 27.8% | **100%** |

## Success criteria

- [x] OnMessageSent + streamReply covered
- [x] SearchContext with results/empty/error
- [x] race PASS ai + messenger
- [x] release build PASS

## Deferred

- handleCommand branches: test-cases success, review with live GitHub/GitLab, translate/sql
- `FetchReview` live VCS paths
