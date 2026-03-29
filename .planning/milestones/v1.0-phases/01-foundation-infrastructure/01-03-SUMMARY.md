---
phase: 01-foundation-infrastructure
plan: 03
subsystem: auth
tags: [jwt, magic-link, telegram, redis, nats, fiber, hmac-sha256]

# Dependency graph
requires:
  - phase: 01-foundation-infrastructure/01-02
    provides: Ent schemas (User, Session, SystemSetting), RBAC middleware, Redis/NATS/outbox clients
provides:
  - Auth service with Magic Link and Telegram login flows
  - JWT access token generation and validation (HS256, 15min)
  - Redis refresh token storage with rotation (30 days)
  - Auth HTTP handler with 12 REST endpoints
  - Auth middleware (RequireAuth) for protected routes
  - Session management (list, revoke)
  - Registration mode toggle (open/invite_only)
  - Profile update and onboarding completion
  - NATS event publishing on session creation
affects: [02-project-management, 03-task-tracker, 04-messenger, 05-tms]

# Tech tracking
tech-stack:
  added: [golang-jwt/jwt/v5]
  patterns: [JWT claims with userID+sessionID, Redis refresh token rotation, Telegram WebApp HMAC verification, dev-mode email logging]

key-files:
  created:
    - backend/internal/auth/dto.go
    - backend/internal/auth/events.go
    - backend/internal/auth/magiclink.go
    - backend/internal/auth/telegram.go
    - backend/internal/auth/service.go
    - backend/internal/auth/handler.go
    - backend/internal/auth/middleware.go
    - backend/internal/auth/service_test.go
    - backend/internal/auth/handler_test.go
  modified:
    - backend/cmd/server/main.go
    - backend/go.mod
    - backend/go.sum

key-decisions:
  - "golang-jwt/jwt/v5 for HS256 JWT signing -- standard Go JWT library, minimal footprint"
  - "Auth middleware in auth package instead of updating RBAC RequireAuth -- cleaner separation, RBAC middleware kept as placeholder"
  - "Telegram users get placeholder email (telegram_<id>@hub.local) until onboarding -- required by unique email constraint"
  - "Dev mode: nil ResendClient logs magic link codes to slog instead of sending email"

patterns-established:
  - "Auth middleware pattern: validate JWT, set userID+sessionID in Fiber Locals"
  - "Login flow: find-or-create user, first user becomes admin, check registration mode"
  - "Refresh token rotation: old token deleted, new token created on each refresh"
  - "Handler pattern: parse request, validate, call service, return DTO response"

requirements-completed: [AUTH-01, AUTH-02, AUTH-03, AUTH-04, AUTH-06, AUTH-07, AUTH-08]

# Metrics
duration: 7min
completed: 2026-03-20
---

# Phase 01 Plan 03: Auth Module Summary

**Magic Link + Telegram auth with JWT/Redis session management, 12 REST endpoints, and NATS event publishing**

## Performance

- **Duration:** 7 min
- **Started:** 2026-03-20T20:41:45Z
- **Completed:** 2026-03-20T20:49:13Z
- **Tasks:** 2
- **Files modified:** 11

## Accomplishments
- Full auth service: Magic Link (6-digit code, Redis, rate limit 3/10min) and Telegram (HMAC-SHA256 WebApp verification)
- JWT access tokens (HS256, 15min) with Redis refresh token rotation (30 days) and Ent session persistence
- 12 HTTP endpoints: magic-link send/verify, telegram login, refresh, logout, me (get/update), onboarding, sessions (list/revoke), admin registration-mode (get/set)
- main.go initializes Postgres, Redis, NATS clients with graceful shutdown and outbox worker

## Task Commits

Each task was committed atomically:

1. **Task 1: Auth service -- Magic Link, Telegram, JWT sessions** - `92bc967` (feat)
2. **Task 2: Auth HTTP handler, routes, main.go integration** - `d756d4f` (feat)

## Files Created/Modified
- `backend/internal/auth/dto.go` - Request/response DTOs for all auth endpoints
- `backend/internal/auth/events.go` - NATS event structs (SessionCreated, UserRegistered)
- `backend/internal/auth/magiclink.go` - Magic link code generation, Redis storage, rate limiting
- `backend/internal/auth/telegram.go` - Telegram WebApp initData HMAC-SHA256 verification
- `backend/internal/auth/service.go` - Core auth business logic (login flows, sessions, profile, registration mode)
- `backend/internal/auth/handler.go` - HTTP handlers for all 12 auth endpoints
- `backend/internal/auth/middleware.go` - JWT auth middleware (RequireAuth)
- `backend/internal/auth/service_test.go` - Unit tests for Telegram HMAC, JWT round-trip, code generation
- `backend/internal/auth/handler_test.go` - Handler unit tests for validation and auth guards
- `backend/cmd/server/main.go` - Infrastructure client initialization, auth route registration, outbox worker
- `backend/go.mod` - Added golang-jwt/jwt/v5 dependency

## Decisions Made
- Used golang-jwt/jwt/v5 for JWT signing (standard Go library, HS256)
- Created dedicated auth middleware instead of modifying RBAC RequireAuth (cleaner separation of concerns)
- Telegram users receive placeholder email until onboarding (required by Ent unique email constraint)
- Dev mode: nil ResendClient causes magic link codes to be logged via slog

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added golang-jwt/jwt/v5 dependency**
- **Found during:** Task 1 (service implementation)
- **Issue:** JWT library not in go.mod, required for token signing/validation
- **Fix:** `go get github.com/golang-jwt/jwt/v5`
- **Files modified:** backend/go.mod, backend/go.sum
- **Verification:** Build succeeds
- **Committed in:** 92bc967

**2. [Rule 2 - Missing Critical] Created auth middleware with JWT validation**
- **Found during:** Task 2 (handler implementation)
- **Issue:** RBAC RequireAuth had TODO for JWT validation; handler needs userID/sessionID in Locals
- **Fix:** Created backend/internal/auth/middleware.go with full JWT parsing and claims extraction
- **Files modified:** backend/internal/auth/middleware.go
- **Verification:** Handler tests pass with auth guard checks
- **Committed in:** d756d4f

---

**Total deviations:** 2 auto-fixed (1 blocking, 1 missing critical)
**Impact on plan:** Both essential for auth to function. No scope creep.

## Issues Encountered
- CGO_ENABLED=0 on Windows prevents -race flag in tests; ran without race detector (non-blocking)

## User Setup Required
None - no external service configuration required. Dev mode logs magic link codes.

## Next Phase Readiness
- Auth module complete, ready for protected endpoint development in Phase 2+
- RBAC RequireAuth placeholder remains for backward compatibility; auth.RequireAuth used for actual JWT validation
- Resend email integration ready (just provide RESEND_API_KEY env var)
- Telegram bot integration ready (just provide TELEGRAM_BOT_TOKEN env var)

## Self-Check: PASSED

- All 9 created files verified present on disk
- Commit 92bc967 verified in git log
- Commit d756d4f verified in git log
- `go build ./...` passes
- `go test ./internal/auth/...` passes (13 tests)

---
*Phase: 01-foundation-infrastructure*
*Completed: 2026-03-20*
