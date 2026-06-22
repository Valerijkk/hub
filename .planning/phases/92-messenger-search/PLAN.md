# Phase 92 — Messenger Polls + Search Repo

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

Закрыть 0% gaps в messenger poll repo + handler DeleteChannel/AddMembers; GetSavedFilter в search.

## Deliverables

### messenger (65.3% → ~67%)
- `ent_repository_poll_test.go` — GetPoll, VotePoll, ClosePoll, ListVotes, FindGlobalChannelByName
- `handler` via poll test file — DeleteChannel + AddMembers HTTP paths

### search (66.1% → ~66%)
- `service_wave92_test.go` — EntRepository GetSavedFilter round-trip

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/messenger/` | 65.3% | **66.8%** |
| `internal/search/` | 66.1% | **66.3%** |
| Overall `internal/` | 71.0% | **71.1%** |

## Success criteria

- [x] Poll repo functions no longer 0%
- [x] DeleteChannel + AddMembers handler covered
- [x] GetSavedFilter ent repo covered
- [ ] messenger/search ≥70% — deferred (needs PG tsvector fixtures)

## Deferred (Phase 93)

- messenger auto_channels Start/Stop lifecycle
- search searchIssues/searchUsers with PG testcontainer
- Commit + PR (on user request)
