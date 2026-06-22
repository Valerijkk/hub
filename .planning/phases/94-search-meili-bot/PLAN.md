# Phase 94 — Search Meili + Bot + Notification JetStream

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

Поднять `search/` до ≥70%, `service_bot` coverage, notification consumer Start с JetStream.

## Deliverables

### search (67.5% → ≥70%)
- `meilisearch_indexer_test.go` — NewMeiliIndexerFromEnv, IndexAllIssues httptest, upsertDocuments errors

### messenger (70.2% → ~73%)
- `service_bot_test.go` — SendBotMessage, UpdateBotMessage, ListMessagesForBot, emitMessageSent hook

### notification (65.5% → ~68%)
- `consumer_wave94_test.go` — Consumer Start/Stop с embedded JetStream + EnsureStreams

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/search/` | 67.5% | **72.7%** |
| `internal/messenger/` | 70.2% | **72.8%** |
| `internal/notification/` | 65.5% | **68.1%** |
| Overall `internal/` | 71.5% | **72.0%** |

## Success criteria

- [x] search ≥70%
- [x] meilisearch indexer 0% → covered
- [x] service_bot covered
- [x] notification Consumer Start JetStream path
- [x] race PASS

## Deferred (Phase 95)

- notification ≥70% (sendToSubscription webpush mock)
- tracker/tms uplift toward 70%
- docker e2e-full local
- Commit + PR (on user request)
