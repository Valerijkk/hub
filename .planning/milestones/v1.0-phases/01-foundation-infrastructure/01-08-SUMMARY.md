---
phase: 01-foundation-infrastructure
plan: 08
subsystem: storage
tags: [s3, seaweedfs, file-upload, fiber, auth-middleware, rbac]

# Dependency graph
requires:
  - phase: 01-foundation-infrastructure
    provides: "auth middleware (RequireAuth), RBAC service, storage module code, Ent schema"
provides:
  - "Storage endpoints wired at /api/files/* with JWT auth and RBAC permissions"
  - "StorageService instantiated in main.go with S3 config"
  - "RBACService instantiated in main.go (reusable by future modules)"
affects: [ui-file-upload, tms-attachments, messenger-attachments]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Auth middleware injection pattern: pass authMw fiber.Handler to RegisterRoutes instead of calling rbac.RequireAuth()"
    - "RBACService instantiation in main.go for shared permission checking"

key-files:
  created: []
  modified:
    - "backend/internal/storage/handler.go"
    - "backend/cmd/server/main.go"

key-decisions:
  - "Auth middleware injected as parameter to RegisterRoutes for clean dependency inversion"
  - "RBACService instantiated in main.go and shared with storage handler for permission checks"

patterns-established:
  - "Handler auth wiring: pass authMw fiber.Handler + rbacSvc *rbac.RBACService to RegisterRoutes"
  - "Direct c.Locals(\"userID\") extraction instead of rbac.UserIDFromCtx helper"

requirements-completed: [UI-08]

# Metrics
duration: 2min
completed: 2026-03-20
---

# Phase 01 Plan 08: Storage Module Gap Closure Summary

**Storage module wired into server with JWT auth middleware injection replacing broken rbac.RequireAuth() placeholder**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-20T22:16:30Z
- **Completed:** 2026-03-20T22:18:19Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Fixed storage handler to use proper auth middleware injection (authMw parameter) instead of broken rbac.RequireAuth() placeholder
- Wired StorageService, RBACService, and storage.Handler in main.go with route registration on /api/files/*
- All file endpoints now reachable with JWT authentication and RBAC permission checks

## Task Commits

Each task was committed atomically:

1. **Task 1: Fix storage handler auth middleware and signature** - `488ed25` (fix)
2. **Task 2: Wire StorageService and storage.Handler in main.go** - `ab35b17` (feat)

## Files Created/Modified
- `backend/internal/storage/handler.go` - Fixed RegisterRoutes signature to accept authMw fiber.Handler, replaced rbac.RequireAuth() with authMw, replaced rbac.UserIDFromCtx with direct c.Locals extraction, added GoDoc comments
- `backend/cmd/server/main.go` - Added storagemod import, instantiated RBACService, StorageService, and storage Handler with route registration

## Decisions Made
- Auth middleware injected as fiber.Handler parameter to RegisterRoutes -- clean dependency inversion, consistent with how auth.RequireAuth(svc) is used elsewhere in the codebase
- RBACService instantiated in main.go between project routes and storage routes -- reusable for future modules that need permission checks
- Used direct c.Locals("userID") extraction pattern instead of rbac.UserIDFromCtx helper -- matches auth middleware contract (sets userID as string)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Storage endpoints are registered and the server compiles successfully
- File upload/download/delete flows require running infrastructure (PostgreSQL, S3/SeaweedFS) for runtime testing
- Gap "Storage module not wired" from VERIFICATION.md is now closed

---
*Phase: 01-foundation-infrastructure*
*Completed: 2026-03-20*
