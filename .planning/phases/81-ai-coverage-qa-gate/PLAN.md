# Phase 81 — AI Coverage + QA Gate

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

- `internal/ai/` **50.7% → ≥60%**
- Full `/qa-gate` PASS on branch

## Deliverables

### Wave 81a (handler + service paths)
- `audit_test.go` — `LogAIChat` via ent + audit repository
- `doc_search_test.go` — nil search, empty query guards
- `coverage_wave81_test.go` — UpdateProvider, ChatCompletions (sync + stream), preferences, AssistJQL, SuggestComposer, RPM limiter, buildChatRequest, anthropic helpers

### Wave 81b (adapter HTTP mocks)
- `adapter_http_test.go` — OpenAI SSE stream, Yandex/GigaChat/Anthropic completion + stream, OAuth token cache, `TestProvider` adapter routing

### Repository
- `repository_ent_test.go` — `UpdateProvider` all fields, invalid URL, clear API key

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/ai/` | 50.7% | **68.5%** |
| Overall `internal/` | 66.2% | **67.6%** |
| `/qa-gate` | — | **PASS** |

## Success criteria

- [x] ai ≥60%
- [x] go build -tags release ./...
- [x] go vet ./...
- [x] go test -race ./...
- [x] frontend tsc + lint + vitest (99/99)

## Deferred

- `messenger_bot.go` OnMessageSent/streamReply (needs MessengerService integration)
- `SearchDocProvider.SearchContext` with live SearchService results
- Adapter success paths via SSRF-safe dialer to real providers (mock HTTP used instead)
