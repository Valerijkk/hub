# Phase 93 — Auto-Channels NATS + Handler Gaps

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

Поднять `messenger/` до ≥70% — auto_channels NATS lifecycle, thread replies, revoke invite, poll handlers; search indexer + admin shard.

## Deliverables

### messenger (66.8% → ≥70%)
- `auto_channels_service_test.go` — Start/Stop с embedded NATS, invalid payload paths
- `handler_wave93_test.go` — ListThreadReplies, RevokeInvite, VotePoll/ClosePoll/GetPollResults

### search (66.3% → ~68%)
- `handler_wave93_test.go` — EnsureSearchIndexes error path, admin user shard, getUserID UUID locals, deleteFilter bad ID

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/messenger/` | 66.8% | **70.2%** |
| `internal/search/` | 66.3% | **67.5%** |
| Overall `internal/` | 71.1% | **71.5%** |

## Success criteria

- [x] messenger ≥70%
- [x] auto_channels Start/Stop NATS covered
- [x] RevokeInvite + ListThreadReplies handler covered
- [ ] search ≥70% — deferred (PG tsvector)
- [x] race PASS

## Deferred (Phase 94)

- search PG testcontainer for searchIssues/searchUsers
- docker e2e-full local validation
- `service_bot` SendBotMessage with redis unread
- Commit + PR (on user request)
