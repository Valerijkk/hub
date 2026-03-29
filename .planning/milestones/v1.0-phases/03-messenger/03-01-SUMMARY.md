---
phase: 03-messenger
plan: 01
subsystem: messenger
tags: [go, fiber, ent, mongodb, nats, websocket, channels, messages, reactions, threads]

requires:
  - phase: 01-foundation
    provides: Ent ORM, NATS bridge, WS hub, auth middleware, RBAC, MongoDB client
  - phase: 02-tracker
    provides: Handler/Service/Repository pattern, unit test mock pattern

provides:
  - Ent schemas for Channel, ChannelMember, Bookmark, PinnedMessage
  - MongoDB message storage with indexes (channel+created_at, thread, text)
  - Messenger Repository interface (dual storage: Ent + MongoDB)
  - MessengerService with 19 methods (channels, messages, reactions, threads, pins, bookmarks, DM, search, forward)
  - REST API under /api/messenger with full CRUD
  - NATS bridge extension for messenger.channel.> -> WS room channel:{id}
  - DM deduplication via sorted user IDs dm_key

affects: [03-messenger, 04-unread, 05-notifications]

tech-stack:
  added: []
  patterns: [dual-storage repository (Ent+MongoDB), CombinedRepository delegation, DM dedup via dm_key]

key-files:
  created:
    - backend/ent/schema/channel.go
    - backend/ent/schema/channelmember.go
    - backend/ent/schema/bookmark.go
    - backend/ent/schema/pinnedmessage.go
    - backend/internal/messenger/repository.go
    - backend/internal/messenger/ent_repository.go
    - backend/internal/messenger/mongo_repository.go
    - backend/internal/messenger/combined_repository.go
    - backend/internal/messenger/service.go
    - backend/internal/messenger/handler.go
    - backend/internal/messenger/dto.go
    - backend/internal/messenger/events.go
    - backend/internal/messenger/service_test.go
  modified:
    - backend/ent/schema/project.go
    - backend/ent/schema/user.go
    - backend/internal/ws/bridge.go
    - backend/cmd/server/main.go

key-decisions:
  - "CombinedRepository pattern: delegates to EntRepository (PG) + MongoRepository (MongoDB) behind single Repository interface"
  - "DM deduplication via sorted user IDs concatenated with '-' as dm_key (unique index)"
  - "Nil-safe NATS publishEvent for testability without mock NATS client"
  - "queryInt helper for Fiber v3 (no built-in QueryInt method)"

patterns-established:
  - "Dual-storage Repository: single interface, two implementations combined via CombinedRepository"
  - "DM dedup: sort user IDs, join with dash, store as unique dm_key"
  - "Messenger events: NATS subject format messenger.channel.{id}.{event}, WS room channel:{id}"

requirements-completed: [MSG-01, MSG-05, MSG-06, MSG-07, MSG-08, MSG-10]

duration: 10min
completed: 2026-03-21
---

# Phase 03 Plan 01: Messenger Core Backend Summary

**Messenger backend with dual-storage (Ent/PG channels + MongoDB messages), REST API (channels/messages/reactions/threads/pins/bookmarks/DM/search/forward), NATS bridge for real-time delivery, 11 passing unit tests**

## Performance

- **Duration:** 10 min
- **Started:** 2026-03-21T00:51:54Z
- **Completed:** 2026-03-21T01:01:34Z
- **Tasks:** 2
- **Files modified:** 17

## Accomplishments
- 4 Ent schemas (Channel, ChannelMember, Bookmark, PinnedMessage) with proper edges and indexes
- Full Repository interface with dual-storage implementation (EntRepository for PG, MongoRepository for MongoDB)
- MessengerService with 19 business methods covering channels, messages, reactions, threads, DM dedup, pins, bookmarks, forward, search
- Handler with all REST routes under /api/messenger, module wired in main.go with MongoDB client
- NATS bridge extended for messenger.channel.> -> WS room channel:{id}
- 11 unit tests covering all key requirements

## Task Commits

1. **Task 1: Ent schemas, MongoDB documents, Repository interface, DTOs, events, bridge** - `a8e8a7c` (feat)
2. **Task 2: Service, Handler, main.go wiring, unit tests** - `eaedadc` (feat)

## Files Created/Modified
- `backend/ent/schema/channel.go` - Channel entity with type, name, readonly, project, dm_key
- `backend/ent/schema/channelmember.go` - Channel membership with role (owner/admin/member)
- `backend/ent/schema/bookmark.go` - User bookmarks on messages
- `backend/ent/schema/pinnedmessage.go` - Pinned messages in channels
- `backend/ent/schema/project.go` - Added channels edge
- `backend/ent/schema/user.go` - Added created_channels, channel_memberships, bookmarks edges
- `backend/internal/messenger/dto.go` - Request/Response types, MessageDocument (BSON)
- `backend/internal/messenger/events.go` - NATS subject constants and event payloads
- `backend/internal/messenger/repository.go` - Repository interface (45+ methods)
- `backend/internal/messenger/ent_repository.go` - Ent/PG implementation (channels, members, pins, bookmarks)
- `backend/internal/messenger/mongo_repository.go` - MongoDB implementation (messages, reactions, search) with EnsureIndexes
- `backend/internal/messenger/combined_repository.go` - CombinedRepository delegating to both
- `backend/internal/messenger/service.go` - MessengerService with 19 business methods
- `backend/internal/messenger/handler.go` - HTTP handlers and RegisterRoutes
- `backend/internal/messenger/service_test.go` - 11 unit tests with mock repository
- `backend/internal/ws/bridge.go` - Extended with messenger.channel.> subscription
- `backend/cmd/server/main.go` - Wired messenger module with MongoDB client

## Decisions Made
- CombinedRepository pattern to unite Ent and MongoDB repositories behind single interface
- DM deduplication via sorted user IDs as dm_key with unique index
- Nil-safe NATS publishEvent for testability without mock NATS
- Custom queryInt helper for Fiber v3 (lacks built-in QueryInt)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Nil pointer dereference on NATS publish in tests**
- **Found during:** Task 2 (unit tests)
- **Issue:** NATS client was nil in test context, causing panic in publishEvent
- **Fix:** Added nil-check guard in publishEvent method
- **Files modified:** backend/internal/messenger/service.go
- **Verification:** All 11 tests pass
- **Committed in:** eaedadc (Task 2 commit)

**2. [Rule 3 - Blocking] Fiber v3 lacks QueryInt method**
- **Found during:** Task 2 (handler compilation)
- **Issue:** c.QueryInt() does not exist in Fiber v3 Ctx interface
- **Fix:** Created queryInt helper using c.Query() + strconv.Atoi
- **Files modified:** backend/internal/messenger/handler.go
- **Verification:** Build passes, handlers work
- **Committed in:** eaedadc (Task 2 commit)

**3. [Rule 2 - Missing Critical] CombinedRepository for dual-storage**
- **Found during:** Task 2 (service creation)
- **Issue:** Repository interface spans both Ent and MongoDB, but plan only specified separate repos
- **Fix:** Created CombinedRepository that delegates each method to the appropriate backing store
- **Files modified:** backend/internal/messenger/combined_repository.go
- **Verification:** Build passes, main.go wires correctly
- **Committed in:** eaedadc (Task 2 commit)

---

**Total deviations:** 3 auto-fixed (1 bug, 1 blocking, 1 missing critical)
**Impact on plan:** All auto-fixes necessary for correctness. No scope creep.

## Issues Encountered
None beyond the auto-fixed deviations.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Messenger backend foundation complete
- Ready for: unread counts (plan 04), typing indicators, presence, frontend integration
- MongoDB indexes created for message queries and text search

---
*Phase: 03-messenger*
*Completed: 2026-03-21*
