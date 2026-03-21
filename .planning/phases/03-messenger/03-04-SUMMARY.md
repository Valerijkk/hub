---
phase: 03-messenger
plan: 04
subsystem: messenger
tags: [redis, presence, unread, search, mongodb, react, zustand, tanstack-query]

requires:
  - phase: 03-messenger (plans 01-03)
    provides: "Channel CRUD, messages, reactions, threads, WS bridge, Plate editor"
  - phase: 01-foundation
    provides: "Redis client, NATS, auth middleware, WebSocket hub"
provides:
  - "PresenceService: Redis sorted set for online/offline/away tracking"
  - "UnreadService: Redis hash for per-channel unread counts"
  - "Presence/unread REST endpoints"
  - "Frontend presence indicators, unread badges, search dialog, pinned messages panel, bookmarks panel, forward dialog"
affects: [04-tms, 05-calls, 06-integrations]

tech-stack:
  added: [miniredis]
  patterns: [redis-sorted-set-presence, redis-hash-unread, polling-hook-pattern]

key-files:
  created:
    - backend/internal/messenger/presence.go
    - backend/internal/messenger/unread.go
    - backend/internal/messenger/presence_test.go
    - backend/internal/messenger/unread_test.go
    - frontend/src/hooks/use-presence.ts
    - frontend/src/hooks/use-unread.ts
    - frontend/src/components/messenger/presence-indicator.tsx
    - frontend/src/components/messenger/search-dialog.tsx
    - frontend/src/components/messenger/pinned-messages.tsx
    - frontend/src/components/messenger/bookmarks-panel.tsx
    - frontend/src/components/messenger/forward-dialog.tsx
  modified:
    - backend/internal/messenger/service.go
    - backend/internal/messenger/handler.go
    - backend/cmd/server/main.go
    - frontend/src/api/messenger.ts
    - frontend/src/components/messenger/channel-list.tsx
    - frontend/src/components/messenger/chat-view.tsx
    - frontend/src/components/messenger/message-item.tsx

key-decisions:
  - "miniredis for unit tests instead of testcontainers for Redis (fast, in-process)"
  - "Presence away threshold 5min, stale cleanup 90s, cleanup loop 60s interval"
  - "Unread counts via Redis pipeline HIncrBy for batch performance"
  - "Bookmarks panel as dialog button in channel list header (no separate route)"
  - "Heartbeat via dedicated REST endpoint POST /presence/heartbeat (25s interval)"

patterns-established:
  - "Redis sorted set for presence: score=unix timestamp, ZAdd for heartbeat, ZScore for status check"
  - "Redis hash for unread: key=unread:{userID}, field=channelID, value=count, pipeline for batch increment"
  - "usePresence hook: polling pattern with 30s refetchInterval + heartbeat every 25s"
  - "useMarkRead: optimistic local store reset + API call + query invalidation"

requirements-completed: [MSG-09, MSG-12, MSG-13]

duration: 7min
completed: 2026-03-21
---

# Phase 03 Plan 04: Presence, Unread, Search, Pins, Bookmarks, Forward Summary

**Redis-based presence (online/offline/away) and unread tracking, message search dialog, pinned messages panel, bookmarks panel, and forward dialog -- completing messenger feature set**

## Performance

- **Duration:** 7 min
- **Started:** 2026-03-21T01:14:19Z
- **Completed:** 2026-03-21T01:22:06Z
- **Tasks:** 2
- **Files modified:** 20

## Accomplishments
- Backend PresenceService with Redis sorted set: heartbeat, get/bulk status, custom text, stale cleanup loop
- Backend UnreadService with Redis hash: increment (skip sender), mark read, get all counts
- 11 unit tests with miniredis (presence: 7 tests, unread: 4 tests) all passing
- Frontend: PresenceIndicator, SearchDialog, PinnedMessages, BookmarksPanel, ForwardDialog
- usePresence hook with 30s polling + heartbeat, useUnread/useMarkRead hooks
- Channel list: DM presence indicators, unread badges from store, bookmarks button
- ChatView: markRead on channel open, search + pin header buttons
- MessageItem: forward hover action with channel picker dialog

## Task Commits

Each task was committed atomically:

1. **Task 1: Presence, Unread backend services + tests** - `ae81f71` (feat)
2. **Task 2: Frontend presence, unread, search, forward UI** - `21f2944` (feat)

## Files Created/Modified
- `backend/internal/messenger/presence.go` - Redis sorted set presence tracking (heartbeat, status, cleanup)
- `backend/internal/messenger/unread.go` - Redis hash unread counts (increment, mark read, get all)
- `backend/internal/messenger/presence_test.go` - 7 tests with miniredis
- `backend/internal/messenger/unread_test.go` - 4 tests with miniredis
- `backend/internal/messenger/service.go` - Added presence/unread fields, methods, unread increment in SendMessage
- `backend/internal/messenger/handler.go` - Presence, unread, pins endpoints
- `backend/cmd/server/main.go` - Wired PresenceService, UnreadService, cleanup goroutine
- `frontend/src/api/messenger.ts` - Presence, unread, pins API functions
- `frontend/src/hooks/use-presence.ts` - Polling hook with heartbeat
- `frontend/src/hooks/use-unread.ts` - Unread counts + markRead
- `frontend/src/components/messenger/presence-indicator.tsx` - Green/gray/yellow dot
- `frontend/src/components/messenger/search-dialog.tsx` - Full-text search with scope toggle
- `frontend/src/components/messenger/pinned-messages.tsx` - Pinned messages panel
- `frontend/src/components/messenger/bookmarks-panel.tsx` - Saved messages list
- `frontend/src/components/messenger/forward-dialog.tsx` - Channel picker for forwarding
- `frontend/src/components/messenger/channel-list.tsx` - Added presence, unread badges, bookmarks
- `frontend/src/components/messenger/chat-view.tsx` - Added markRead, search, pins header
- `frontend/src/components/messenger/message-item.tsx` - Added forward action

## Decisions Made
- Used miniredis (in-process Redis) for unit tests -- fast, no Docker dependency
- Presence away threshold set to 5 minutes; stale records cleaned after 90 seconds
- Unread counts use Redis pipeline for batch increment performance
- Bookmarks accessed via dialog button in channel list header (avoids creating new route)
- Heartbeat via dedicated POST endpoint rather than WS ping for simplicity

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Bookmarks route /messenger/bookmarks does not exist**
- **Found during:** Task 2 (channel-list.tsx update)
- **Issue:** TanStack Router type-checked Link `to="/messenger/bookmarks"` failed -- route not defined
- **Fix:** Replaced Link with dialog button that opens BookmarksPanel in a Dialog component
- **Files modified:** frontend/src/components/messenger/channel-list.tsx
- **Verification:** TypeScript compilation passes
- **Committed in:** 21f2944 (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Minor UI approach change, same functionality delivered via dialog instead of route.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Messenger feature set complete (channels, messages, threads, reactions, presence, unread, search, pins, bookmarks, forward)
- Ready for Plan 05 (documentation) or Phase 04 (TMS)

---
*Phase: 03-messenger*
*Completed: 2026-03-21*
