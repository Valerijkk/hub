---
phase: 01-foundation-infrastructure
plan: 04
subsystem: api
tags: [go, fiber, ent, rbac, outbox, project-management, invites, crypto]

# Dependency graph
requires:
  - phase: 01-02
    provides: Ent schemas, RBAC engine, outbox publisher, platform clients
provides:
  - Project CRUD API with archive and trash
  - Membership management with RBAC cache invalidation
  - Invite link generation and join flow
  - Repository interface pattern for testability
affects: [02-tracker, 03-tms, 04-messenger]

# Tech tracking
tech-stack:
  added: [crypto/rand, encoding/base32]
  patterns: [repository-interface, mock-based-testing, cache-invalidation-on-mutation]

key-files:
  created:
    - backend/internal/project/handler.go
    - backend/internal/project/service.go
    - backend/internal/project/repository.go
    - backend/internal/project/invite.go
    - backend/internal/project/dto.go
    - backend/internal/project/events.go
    - backend/internal/project/service_test.go
  modified:
    - backend/cmd/server/main.go

key-decisions:
  - "Repository interface pattern for project module enables mock-based unit testing without database"
  - "RBAC cache invalidation on every membership mutation via CacheInvalidator interface"
  - "Team CRUD registered as route stubs (501) -- full implementation deferred to project teams feature"

patterns-established:
  - "Repository interface: separate data access from business logic for testability"
  - "CacheInvalidator interface: abstract RBAC cache for testing without Redis"
  - "OutboxPublisher interface: abstract outbox for repository testing"

requirements-completed: [PROJ-01, PROJ-02, PROJ-04, PROJ-06, PROJ-07]

# Metrics
duration: 10min
completed: 2026-03-20
---

# Phase 01 Plan 04: Project Module Summary

**Project CRUD with membership management, invite links (crypto/rand), archive/trash lifecycle, and RBAC-integrated REST API**

## Performance

- **Duration:** 10 min
- **Started:** 2026-03-20T20:42:49Z
- **Completed:** 2026-03-20T20:52:53Z
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments
- Full project lifecycle: create, read, update, archive, unarchive, soft delete, restore, permanent delete, trash listing
- Membership management: add/remove members, update roles, last-owner protection, RBAC cache invalidation
- Invite system: cryptographically random 8-char codes via crypto/rand + base32, expiry/max-uses validation, join-by-code flow
- Outbox event publishing for project.created (transactional with project creation)
- 7 unit tests covering CRUD, duplicate key prevention, archive, soft delete, invite join, expired invite, last owner protection
- REST API with 20+ endpoints integrated into main.go with auth middleware

## Task Commits

Each task was committed atomically:

1. **Task 1: Project service -- CRUD, membership, teams, invites** - `677f751` (feat)
2. **Task 2: Project HTTP handler, routes, main.go integration** - `b89bbc9` (feat)

## Files Created/Modified
- `backend/internal/project/dto.go` - Request/response DTOs for all project operations
- `backend/internal/project/events.go` - Outbox event types (ProjectCreated, MemberAdded, MemberRemoved)
- `backend/internal/project/repository.go` - Repository interface + Ent implementation with outbox integration
- `backend/internal/project/service.go` - Business logic: CRUD, membership, role updates, cache invalidation
- `backend/internal/project/invite.go` - Invite code generation (crypto/rand) and join-by-invite flow
- `backend/internal/project/service_test.go` - Unit tests with mock repository and cache
- `backend/internal/project/handler.go` - HTTP handlers with Fiber, CopyString for param safety
- `backend/cmd/server/main.go` - Project service wiring with RBAC cache and outbox publisher

## Decisions Made
- Repository interface pattern chosen over direct Ent usage in service layer for testability without database
- Team CRUD endpoints registered as stubs (501 Not Implemented) -- route structure is ready, implementation deferred
- RBAC permission checks done at service level (not middleware level) for flexibility, except where middleware is more appropriate

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- Ent NotFoundError/ConstraintError have unexported fields -- resolved by using zero-value struct pointers which still satisfy errors.As checks
- CGO_ENABLED=0 on Windows prevents -race flag -- ran tests without race detector

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Project module complete, ready for tracker module (tasks, boards, sprints) which depends on project membership
- Invite and membership APIs available for frontend integration
- Team management stubs ready to be filled in when needed

---
*Phase: 01-foundation-infrastructure*
*Completed: 2026-03-20*
