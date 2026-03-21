---
phase: 05-voice-calls
plan: 03
subsystem: calls
tags: [mongodb, ephemeral-chat, nats-bridge, websocket, call-lifecycle]

requires:
  - phase: 05-voice-calls
    provides: Ent schemas Call/CallParticipant, CallService, REST API, webhook handler
  - phase: 03-messenger
    provides: MongoDB patterns (MongoRepository), WS bridge patterns, NATS pub/sub
provides:
  - ChatRepository for ephemeral call chat (MongoDB call_chat_messages collection)
  - Ephemeral chat auto-deletion on room_finished (CALL-06)
  - NATS bridge subscriptions for calls.channel.> and calls.chat.>
  - REST endpoints POST/GET /api/calls/:callId/chat
affects: [05-voice-calls, frontend-calls-integration]

tech-stack:
  added: []
  patterns: [ephemeral-mongodb-collection, nats-ws-bridge-routing]

key-files:
  created:
    - backend/internal/calls/chat.go
    - backend/internal/calls/chat_test.go
  modified:
    - backend/internal/calls/service.go
    - backend/internal/calls/handler.go
    - backend/internal/ws/bridge.go
    - backend/cmd/server/main.go

key-decisions:
  - "ChatRepository uses MongoDB with compound index {call_id: 1, created_at: 1} for efficient message retrieval"
  - "Text limit 2000 chars enforced at repository level"
  - "OnRoomFinished deletes chat via chatRepo.DeleteCallChat for ephemeral behavior"
  - "Call events routed to channel:{channelID} WS rooms (same as messenger) with call.* event prefix"
  - "Chat messages routed to call:{callID} WS rooms with chat.message event type"
  - "Frontend files not modified per execution constraint (backend-only in this plan)"

patterns-established:
  - "Ephemeral MongoDB data pattern: insert during lifecycle, DeleteMany on completion"
  - "WS bridge multi-subject routing: calls.channel.> -> channel rooms, calls.chat.> -> call rooms"

requirements-completed: [CALL-06, CALL-01]

duration: 5min
completed: 2026-03-21
---

# Phase 5 Plan 3: Call Lifecycle Integration Summary

Ephemeral call chat with MongoDB storage and NATS-to-WebSocket bridge for call events and chat messages.

## What Was Built

### Task 1: ChatRepository and REST Endpoints
- `ChatRepository` (MongoDB) with `SendMessage`, `ListMessages`, `DeleteCallChat` operations
- `CallChatMessage` struct with bson/json tags for MongoDB document mapping
- Compound index `{call_id: 1, created_at: 1}` for efficient message retrieval
- `CallService` extended with `chatRepo` field, `SendChatMessage` and `ListChatMessages` methods
- `OnRoomFinished` updated to call `DeleteCallChat` ensuring ephemeral behavior (CALL-06)
- REST endpoints: `POST /api/calls/:callId/chat` (send), `GET /api/calls/:callId/chat?limit=N` (list)
- Text validation: max 2000 characters
- Unit tests for SendMessage, ListMessages, DeleteCallChat, EnsureIndexes (skip when MongoDB unavailable)

### Task 2: NATS Bridge Subscriptions
- `calls.channel.>` subscription routes call lifecycle events (started/ended/joined/left) to `channel:{channelID}` WS rooms
- `calls.chat.>` subscription routes chat messages to `call:{callID}` WS rooms with `chat.message` event type
- Call events use `call.` prefix (call.started, call.ended, call.joined, call.left) to distinguish from messenger events
- Both handlers follow established messenger bridge pattern

## Deviations from Plan

### Scope Adjustment

**Frontend files not modified** per execution constraint ("Do NOT modify frontend files"). The plan included frontend tasks (CallChat.tsx, CallRoom.tsx, calls.ts) which are deferred to a separate frontend-focused plan.

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed Fiber v3 QueryInt API**
- **Found during:** Task 1
- **Issue:** `c.QueryInt()` does not exist in Fiber v3
- **Fix:** Used `c.Query("limit")` with `strconv.Atoi` parsing, matching project patterns
- **Files modified:** backend/internal/calls/handler.go
- **Commit:** 5f34ec9

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| 1 | 5f34ec9 | ChatRepository, REST endpoints, OnRoomFinished cleanup, tests |
| 2 | d07e6c1 | NATS bridge subscriptions for call events and chat messages |

## Self-Check: PASSED
