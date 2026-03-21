---
phase: 05-voice-calls
plan: 01
subsystem: calls
tags: [livekit, webrtc, voice-calls, jwt, webhook, docker]

requires:
  - phase: 03-messenger
    provides: Channel schema, messenger patterns (handler/service/repository), NATS pub/sub
  - phase: 01-foundation
    provides: Auth middleware, Ent ORM, config platform, Fiber app setup
provides:
  - Ent schemas Call and CallParticipant with edges to User and Channel
  - CallService with LiveKit room management and JWT token generation
  - REST API for call lifecycle (create/join/leave/end/raise-hand)
  - Webhook handler for LiveKit lifecycle events
  - LiveKit server in Docker Compose
  - LiveKit config (livekit.yaml) with webhook URL
  - Caddy reverse proxy for LiveKit signaling
affects: [05-voice-calls, 06-dashboards]

tech-stack:
  added: [livekit/server-sdk-go/v2, livekit/protocol]
  patterns: [manual-webhook-jwt-verification, nil-safe-nats-publish]

key-files:
  created:
    - backend/ent/schema/call.go
    - backend/ent/schema/callparticipant.go
    - backend/internal/calls/models.go
    - backend/internal/calls/repository.go
    - backend/internal/calls/service.go
    - backend/internal/calls/handler.go
    - backend/internal/calls/service_test.go
    - deployments/livekit.yaml
  modified:
    - backend/ent/schema/user.go
    - backend/ent/schema/channel.go
    - backend/internal/platform/config/config.go
    - backend/cmd/server/main.go
    - deployments/docker-compose.yml
    - deployments/Caddyfile

key-decisions:
  - "Manual webhook JWT verification instead of webhook.ReceiveWebhookEvent (incompatible with Fiber/fasthttp)"
  - "Nil-safe NATS publishing pattern reused from messenger module"
  - "Metadata field as plain string (not pointer) per LiveKit protocol v1.44 API"

patterns-established:
  - "Calls module follows handler/service/repository pattern consistent with messenger and TMS"
  - "Webhook endpoints registered without auth middleware (external service calls directly)"

requirements-completed: [CALL-01]

duration: 7min
completed: 2026-03-21
---

# Phase 5 Plan 1: LiveKit Integration Summary

**LiveKit backend integration with Ent schemas, REST API for call lifecycle, webhook handling, Docker Compose config, and Caddy reverse proxy**

## Performance

- **Duration:** 7 min
- **Started:** 2026-03-21T03:13:09Z
- **Completed:** 2026-03-21T03:20:08Z
- **Tasks:** 2
- **Files modified:** 15

## Accomplishments
- Ent schemas Call and CallParticipant with proper edges, indexes, and codegen
- CallService generating LiveKit JWT tokens, managing rooms via RoomServiceClient, handling webhooks
- REST API: POST /api/calls, POST /api/calls/:callId/join, /leave, /end, /raise-hand, GET /api/calls/channel/:channelId, POST /api/calls/webhook
- LiveKit server configured in Docker Compose with livekit.yaml
- 5 unit tests passing (token generation, nil-safe NATS, webhook validation)

## Task Commits

Each task was committed atomically:

1. **Task 1: Ent schemas, LiveKit config, Docker Compose, models and repository** - `48d97f9` (feat)
2. **Task 2: CallService, Handler, tests, main.go registration** - `f484eb6` (feat)

## Files Created/Modified
- `backend/ent/schema/call.go` - Call entity schema with room_name, status, channel/creator edges
- `backend/ent/schema/callparticipant.go` - CallParticipant entity with call/user edges, unique index
- `backend/internal/calls/models.go` - DTOs for create/join/raise-hand requests and responses
- `backend/internal/calls/repository.go` - EntRepository with CRUD for calls and participants
- `backend/internal/calls/service.go` - CallService with LiveKit room management, token generation, webhooks
- `backend/internal/calls/handler.go` - HTTP handlers with manual webhook JWT verification
- `backend/internal/calls/service_test.go` - Unit tests for token generation, nil-safe NATS, webhook validation
- `deployments/livekit.yaml` - LiveKit server config with webhook URL and room settings
- `backend/ent/schema/user.go` - Added created_calls and call_participations edges
- `backend/ent/schema/channel.go` - Added calls edge
- `backend/internal/platform/config/config.go` - Added LiveKitConfig struct and env loading
- `backend/cmd/server/main.go` - Registered calls module routes
- `deployments/docker-compose.yml` - Added LiveKit service with ports 7880, 7881, 50000-50100/udp
- `deployments/Caddyfile` - Added /livekit/* reverse proxy to localhost:7880

## Decisions Made
- Manual webhook JWT verification: LiveKit webhook.ReceiveWebhookEvent expects *http.Request which is incompatible with Fiber/fasthttp. Implemented JWT parse + SHA256 body hash verification directly.
- Nil-safe NATS publishing: Reused pattern from messenger -- publishEvent logs and returns when NATS client is nil, enabling testability without mocks.
- Metadata as plain string: LiveKit protocol v1.44 uses string (not *string) for UpdateParticipantRequest.Metadata field.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed Metadata pointer type mismatch**
- **Found during:** Task 2 (CallService implementation)
- **Issue:** Plan specified &meta (pointer) for UpdateParticipantRequest.Metadata, but protocol uses plain string
- **Fix:** Changed from &meta to meta (plain string value)
- **Files modified:** backend/internal/calls/service.go
- **Verification:** go build passes
- **Committed in:** f484eb6 (Task 2 commit)

**2. [Rule 3 - Blocking] Added missing fmt import in handler.go**
- **Found during:** Task 2 (Handler implementation)
- **Issue:** fmt package used in verifyWebhookSignature but not imported
- **Fix:** Added fmt to import block
- **Files modified:** backend/internal/calls/handler.go
- **Verification:** go build passes
- **Committed in:** f484eb6 (Task 2 commit)

---

**Total deviations:** 2 auto-fixed (1 bug, 1 blocking)
**Impact on plan:** Both fixes necessary for compilation. No scope creep.

## Issues Encountered
None beyond the auto-fixed deviations above.

## User Setup Required
None - no external service configuration required. LiveKit uses default dev credentials (devkey/devsecret).

## Next Phase Readiness
- Backend calls module complete with all REST endpoints and webhook handler
- Ready for frontend call UI (Phase 5 Plan 2)
- Ready for ephemeral call chat (Phase 5 Plan 3)
- LiveKit Docker service configured and ready to start

## Self-Check: PASSED

All 8 created files verified. Both task commits (48d97f9, f484eb6) confirmed in git log.

---
*Phase: 05-voice-calls*
*Completed: 2026-03-21*
