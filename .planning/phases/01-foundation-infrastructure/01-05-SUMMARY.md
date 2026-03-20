---
phase: 01-foundation-infrastructure
plan: 05
subsystem: infra, ui
tags: [websocket, nats, zustand, react-hooks, rbac, api-client, fiber]

# Dependency graph
requires:
  - phase: 01-foundation-infrastructure/01-03
    provides: RBAC engine with permissions bitmap
  - phase: 01-foundation-infrastructure/01-04
    provides: Storage service with presigned URL flow
provides:
  - WebSocket gateway with Hub, Handler, and NATS Bridge
  - Frontend API client with auto-refresh on 401
  - Zustand stores for auth, sidebar, and WebSocket state
  - React hooks for auth, websocket, and file upload
  - Frontend permissions mirror matching backend RBAC bitmap
affects: [02-messenger, 03-tms, 04-calls, 01-06-ui-scaffold]

# Tech tracking
tech-stack:
  added: [fasthttp/websocket]
  patterns: [ws-hub-pattern, nats-bridge-fan-out, api-client-auto-refresh, zustand-persist, bigint-permission-bitmap]

key-files:
  created:
    - backend/internal/ws/protocol.go
    - backend/internal/ws/hub.go
    - backend/internal/ws/handler.go
    - backend/internal/ws/bridge.go
    - frontend/src/api/client.ts
    - frontend/src/api/auth.ts
    - frontend/src/api/projects.ts
    - frontend/src/api/files.ts
    - frontend/src/stores/auth-store.ts
    - frontend/src/stores/sidebar-store.ts
    - frontend/src/stores/ws-store.ts
    - frontend/src/hooks/use-auth.ts
    - frontend/src/hooks/use-websocket.ts
    - frontend/src/hooks/use-file-upload.ts
    - frontend/src/lib/permissions.ts
  modified:
    - backend/cmd/server/main.go
    - backend/go.mod
    - backend/go.sum

key-decisions:
  - "fasthttp/websocket directly instead of gofiber/contrib/websocket (contrib only supports Fiber v2, project uses v3)"
  - "JWT auth via query param for WS handshake (WebSocket does not support custom headers)"
  - "sync.Map for hub clients and rooms (lock-free concurrent access for WS connections)"

patterns-established:
  - "WS Hub pattern: register/unregister channels + sync.Map for clients and rooms"
  - "NATS Bridge fan-out: subject prefix -> room key mapping (project.{id}.* -> project:{id})"
  - "API client auto-refresh: 401 triggers tryRefresh before retry"
  - "Zustand persist middleware for auth and sidebar state"
  - "BigInt permission bitmap on frontend mirroring backend uint64 RBAC"

requirements-completed: [UI-08]

# Metrics
duration: 7min
completed: 2026-03-21
---

# Phase 01 Plan 05: WebSocket Gateway + Frontend Infrastructure Summary

**WebSocket gateway with NATS bridge for real-time fan-out, plus frontend API client with auto-refresh, Zustand stores, React hooks, and RBAC permissions mirror**

## Performance

- **Duration:** 7 min
- **Started:** 2026-03-20T20:57:08Z
- **Completed:** 2026-03-20T21:04:00Z
- **Tasks:** 3
- **Files modified:** 18

## Accomplishments
- WebSocket gateway with Hub (sync.Map), Handler (JWT via query param), and NATS Bridge (project.>/auth.session.> fan-out) integrated in main.go
- Frontend API client with automatic JWT refresh on 401 responses
- Zustand stores: auth (persist to localStorage), sidebar (persist), ws (WebSocket connection state with event listeners)
- React hooks: use-auth (auto-fetch profile), use-websocket (connect/rooms/events), use-file-upload (drag-drop/paste with presigned URL flow)
- Frontend permissions.ts with 26 BigInt constants matching backend RBAC bitmap exactly

## Task Commits

Each task was committed atomically:

1. **Task 1: WebSocket gateway backend (Go)** - `8783b45` (feat)
2. **Task 2: Frontend API client, stores, hooks, permissions** - `754f211` (feat)
3. **Task 3: Permissions sync verification** - no commit (verification-only task, 26=26 confirmed)

## Files Created/Modified
- `backend/internal/ws/protocol.go` - WebSocket message types (subscribe, unsubscribe, event, ping, pong)
- `backend/internal/ws/hub.go` - Connection hub with sync.Map, room subscriptions, broadcast
- `backend/internal/ws/handler.go` - Fiber handler with JWT auth via query param, read/write pumps
- `backend/internal/ws/bridge.go` - NATS-to-WebSocket bridge for project and auth session events
- `backend/cmd/server/main.go` - Integrated hub, bridge, and /ws route
- `frontend/src/api/client.ts` - API client with auto-refresh on 401
- `frontend/src/api/auth.ts` - Auth API functions (magic link, telegram, sessions)
- `frontend/src/api/projects.ts` - Project CRUD, members, invites API
- `frontend/src/api/files.ts` - File upload/download with presigned URL flow
- `frontend/src/stores/auth-store.ts` - Auth state with persist middleware
- `frontend/src/stores/sidebar-store.ts` - Sidebar collapsed state with persist
- `frontend/src/stores/ws-store.ts` - WebSocket connection and event listener state
- `frontend/src/hooks/use-auth.ts` - Auth hook with auto-fetch profile
- `frontend/src/hooks/use-websocket.ts` - WebSocket connect/rooms/events hook
- `frontend/src/hooks/use-file-upload.ts` - Drag-drop and paste upload hook
- `frontend/src/lib/permissions.ts` - 26 BigInt permission constants mirroring backend

## Decisions Made
- Used `fasthttp/websocket` directly instead of `gofiber/contrib/websocket` because the contrib package only supports Fiber v2, while the project uses Fiber v3
- JWT authentication for WebSocket via query parameter `?token=` since WebSocket handshake does not support custom headers
- sync.Map for hub clients and rooms for lock-free concurrent access without traditional mutex locking

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fiber v3 WebSocket library incompatibility**
- **Found during:** Task 1 (WebSocket gateway backend)
- **Issue:** Plan referenced Fiber WebSocket middleware but gofiber/contrib/websocket only supports Fiber v2. Fiber v3 uses different Ctx interface (RequestCtx() method instead of direct fasthttp context).
- **Fix:** Used fasthttp/websocket directly with Fiber v3's c.RequestCtx() for upgrade. Fixed CheckOrigin type signature and import.
- **Files modified:** backend/internal/ws/handler.go
- **Verification:** go build ./... passes
- **Committed in:** 8783b45

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Fix was necessary for Fiber v3 compatibility. No scope creep.

## Issues Encountered
None beyond the Fiber v3 WebSocket library compatibility addressed above.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- WebSocket gateway ready for messenger and real-time features in Phase 2
- Frontend infrastructure layer (API client, stores, hooks) ready for UI scaffold in Plan 06
- NATS bridge pattern established for all future real-time modules

---
*Phase: 01-foundation-infrastructure*
*Completed: 2026-03-21*
