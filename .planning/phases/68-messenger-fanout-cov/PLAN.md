# Phase 68 — Messenger Fan-out & Coverage (52.3% → 60%)

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-26

## Goal

PLAN-CLOSE-GAPS **2.5**: messenger fan-out integration + coverage **≥60%**.

## Deliverables

### Fan-out (MSG-01)
- `service_fanout_test.go` — SendMessage → unread increment for N-1 members, sender excluded

### Ent / auto-channels
- `ent_repository_invites_test.go` — invites CRUD, project members, polls
- `auto_channels_service_test.go` — handleProjectCreated + ent ListProjectMembers, CombinedRepository

### Handler coverage
- `handler_channels_extra_test.go` — join/leave, star/mute, invites, edit/delete, typing

## Results

| Metric | Before | After |
|--------|--------|-------|
| `messenger/` coverage | 52.3% | **62.1%** |

## Success criteria

- [x] messenger ≥60%
- [x] fan-out unread test (sender excluded)
- [x] all messenger tests pass
