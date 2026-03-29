---
phase: 03-messenger
plan: 02
subsystem: messenger
tags: [channels, nats, auto-channels, readonly, rbac, go]

requires:
  - phase: 03-01
    provides: "Base messenger module with channels, messages, reactions, threads, MongoDB+Ent repositories"
provides:
  - "Public/private channel join differentiation with specific error messages"
  - "DM channel join protection (requires CreateOrGetDM)"
  - "AutoChannelService subscribing to project.created NATS events"
  - "#project-{key} auto-channel creation with project member seeding"
  - "AddMembersInternal for internal service use without permission checks"
  - "ListProjectMembers and GetProjectByID in EntRepository"
  - "11 new unit tests covering channel management scenarios"
affects: [03-03, 03-04, 03-05]

tech-stack:
  added: []
  patterns: [auto-channel-nats-subscription, internal-service-method-no-auth]

key-files:
  created:
    - backend/internal/messenger/auto_channels.go
    - backend/internal/messenger/auto_channels_test.go
  modified:
    - backend/internal/messenger/service.go
    - backend/internal/messenger/ent_repository.go
    - backend/internal/messenger/service_test.go
    - backend/cmd/server/main.go

key-decisions:
  - "AutoChannelService uses core NATS subscription (not JetStream) for project.created events published by outbox worker"
  - "AddMembersInternal method without permission checks for internal auto-channel use"
  - "Test helper simulateProjectCreated pattern for testing auto-channel logic without real EntRepository"

patterns-established:
  - "Internal service methods (*Internal suffix) bypass permission checks for system-level operations"
  - "NATS subscription in separate service struct with Start/Stop lifecycle"

requirements-completed: [MSG-02, MSG-03, MSG-04]

duration: 5min
completed: 2026-03-21
---

# Phase 03 Plan 02: Channel Management Summary

**Public/private channel join logic, auto-channel creation via NATS project.created events, readonly enforcement, and 11 unit tests**

## Performance

- **Duration:** 5 min
- **Started:** 2026-03-21T01:04:33Z
- **Completed:** 2026-03-21T01:09:32Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- JoinChannel now differentiates public (allowed), private (forbidden with invitation message), and DM (forbidden with CreateOrGetDM guidance)
- AutoChannelService subscribes to project.created NATS events and creates #project-{key} public channels with all project members
- 11 new unit tests covering all channel management scenarios (public/private join, readonly block/allow, admin checks, project binding, auto-channels)

## Task Commits

Each task was committed atomically:

1. **Task 1: Public/private channels, readonly, auto-channels via NATS** - `072d43f` (feat)
2. **Task 2: Unit tests for channel management** - `f309b65` (test)

## Files Created/Modified
- `backend/internal/messenger/auto_channels.go` - AutoChannelService with NATS subscription for project.created
- `backend/internal/messenger/auto_channels_test.go` - Tests for auto-channel creation and member seeding
- `backend/internal/messenger/service.go` - Extended JoinChannel with per-type error messages, added AddMembersInternal
- `backend/internal/messenger/ent_repository.go` - Added ListProjectMembers and GetProjectByID
- `backend/internal/messenger/service_test.go` - 11 new tests + fixed existing test for new error message
- `backend/cmd/server/main.go` - Wired AutoChannelService with Start/Stop lifecycle

## Decisions Made
- AutoChannelService uses core NATS (not JetStream) since outbox worker publishes to core NATS subject "project.created"
- AddMembersInternal bypasses permission checks for internal auto-channel member seeding
- Test helper pattern (simulateProjectCreated) avoids needing real EntRepository for auto-channel tests

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed existing test assertion for new error message**
- **Found during:** Task 2
- **Issue:** TestJoinChannel_PublicOnly expected old error message "публичному" but JoinChannel now returns "приватный" for private channels
- **Fix:** Updated assertion to match new error message
- **Files modified:** backend/internal/messenger/service_test.go
- **Verification:** All 22 tests pass
- **Committed in:** f309b65 (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** Necessary fix for test compatibility with new error messages. No scope creep.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Channel management complete, ready for frontend integration (03-03)
- Auto-channels will fire when project creation events flow through NATS
- All 22 messenger tests pass (11 from 03-01 + 11 new)

---
*Phase: 03-messenger*
*Completed: 2026-03-21*
