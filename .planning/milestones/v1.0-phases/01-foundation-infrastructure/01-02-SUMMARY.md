---
phase: 01-foundation-infrastructure
plan: 02
subsystem: database, rbac, storage, messaging
tags: [ent, postgresql, mongodb, redis, nats, jetstream, rbac, bitmap, s3, seaweedfs, outbox]

# Dependency graph
requires:
  - phase: 01-foundation-infrastructure/01
    provides: "Go+Fiber server, config loader, middleware, Docker Compose"
provides:
  - "12 Ent ORM schemas (User, Session, Project, ProjectMember, Role, Team, TeamMember, Invite, FileMetadata, OutboxEvent, SystemSetting)"
  - "Ent code generation pipeline"
  - "PostgreSQL, MongoDB, Redis, NATS platform clients"
  - "RBAC bitmap permission engine with Redis cache"
  - "S3 presigned URL storage service for SeaweedFS"
  - "Outbox pattern worker for PG-to-NATS event relay"
affects: [auth, projects, messenger, tms, storage, websocket]

# Tech tracking
tech-stack:
  added: [entgo.io/ent, lib/pq, mongo-driver/v2, go-redis/v9, nats.go, jetstream, aws-sdk-go-v2]
  patterns: [ent-schema-with-uuid-pk, bitmap-rbac, redis-permission-cache, s3-presigned-url-with-pathstyle, outbox-pattern-poll-and-publish, fiber-ctx-copy-string]

key-files:
  created:
    - backend/ent/schema/user.go
    - backend/ent/schema/session.go
    - backend/ent/schema/project.go
    - backend/ent/schema/projectmember.go
    - backend/ent/schema/role.go
    - backend/ent/schema/team.go
    - backend/ent/schema/teammember.go
    - backend/ent/schema/invite.go
    - backend/ent/schema/filemetadata.go
    - backend/ent/schema/outboxevent.go
    - backend/ent/schema/systemsetting.go
    - backend/ent/generate.go
    - backend/internal/platform/database/postgres.go
    - backend/internal/platform/database/mongodb.go
    - backend/internal/platform/redis/client.go
    - backend/internal/platform/nats/client.go
    - backend/internal/rbac/permissions.go
    - backend/internal/rbac/service.go
    - backend/internal/rbac/middleware.go
    - backend/internal/rbac/cache.go
    - backend/internal/storage/service.go
    - backend/internal/storage/handler.go
    - backend/internal/platform/outbox/publisher.go
    - backend/internal/platform/outbox/worker.go
  modified:
    - backend/go.mod
    - backend/go.sum

key-decisions:
  - "UUID primary keys on all entities for distributed ID generation"
  - "Ent Privacy Layer deferred to module-level implementation (schema foundation established)"
  - "RBAC uses uint64 bitmap with 26 permissions covering all Phase 1 modules"
  - "Redis cache TTL 5 minutes for RBAC permission lookups"
  - "Outbox worker polls at 100ms intervals with 10-attempt retry limit"
  - "S3 UsePathStyle=true mandatory for SeaweedFS compatibility"

patterns-established:
  - "Ent schema pattern: UUID PK, created_at/updated_at, enum fields for status"
  - "Platform client pattern: NewXClient(ctx, url) returning typed client + error"
  - "RBAC middleware: RequireAuth -> RequirePermission with CopyString for Fiber ctx safety"
  - "Outbox pattern: PublishInTx writes to PG, Worker polls and publishes to NATS"
  - "Storage flow: presigned upload URL -> pending status -> confirm -> confirmed status"

requirements-completed: [AUTH-05, PROJ-03, UI-08]

# Metrics
duration: 5min
completed: 2026-03-20
---

# Phase 01 Plan 02: Ent Schemas, DB Clients, RBAC, Storage, Outbox Summary

**12 Ent ORM schemas with code generation, RBAC bitmap permission engine with Redis cache, S3 presigned URL storage for SeaweedFS, and outbox pattern for PG-to-NATS event relay**

## Performance

- **Duration:** 5 min
- **Started:** 2026-03-20T20:33:04Z
- **Completed:** 2026-03-20T20:38:45Z
- **Tasks:** 2
- **Files modified:** 29

## Accomplishments
- 12 Ent schemas covering all Phase 1 entities with proper fields, edges, and indexes
- Platform clients for PostgreSQL, MongoDB, Redis, and NATS with JetStream
- RBAC engine with 26 bitmap permissions, 4 default roles, and Redis-cached lookups
- S3 presigned URL service configured for SeaweedFS (UsePathStyle) with upload/confirm/download/delete flow
- Outbox worker for transactional event publishing from PG to NATS

## Task Commits

Each task was committed atomically:

1. **Task 1: Ent schemas, code generation, DB clients** - `0706d3f` (feat)
2. **Task 2: RBAC engine, Storage service, Outbox pattern** - `a6a1e76` (feat)

## Files Created/Modified
- `backend/ent/schema/*.go` - 12 Ent entity schemas (User, Session, Project, ProjectMember, Role, Team, TeamMember, Invite, FileMetadata, OutboxEvent, SystemSetting) + helpers
- `backend/ent/generate.go` - go:generate directive for Ent code generation
- `backend/internal/platform/database/postgres.go` - PostgreSQL Ent client with auto-migration
- `backend/internal/platform/database/mongodb.go` - MongoDB client (mongo-driver/v2)
- `backend/internal/platform/redis/client.go` - Redis client (go-redis/v9)
- `backend/internal/platform/nats/client.go` - NATS client with JetStream support
- `backend/internal/rbac/permissions.go` - Permission bitmap type with 26 constants and 4 default role sets
- `backend/internal/rbac/service.go` - RBAC service with Redis-cached effective permission lookups
- `backend/internal/rbac/middleware.go` - Fiber middleware: RequireAuth, RequireSystemRole, RequirePermission
- `backend/internal/rbac/cache.go` - Redis-backed permission cache with per-user and per-project invalidation
- `backend/internal/storage/service.go` - S3 presigned URL service for SeaweedFS
- `backend/internal/storage/handler.go` - HTTP handlers for file upload/confirm/download/delete
- `backend/internal/platform/outbox/publisher.go` - Transactional event publisher (writes within Ent TX)
- `backend/internal/platform/outbox/worker.go` - Background worker polling outbox and publishing to NATS

## Decisions Made
- UUID primary keys on all entities for distributed ID generation without coordination
- RBAC uses uint64 bitmap allowing up to 64 permissions (26 defined for Phase 1 modules)
- Redis cache TTL of 5 minutes balances freshness with performance for permission checks
- Outbox worker uses 100ms polling interval for near-real-time event delivery
- S3 UsePathStyle=true is mandatory for SeaweedFS (not AWS) compatibility
- RequireAuth middleware is a placeholder for JWT validation (to be completed in auth module)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Added User->Invite edge**
- **Found during:** Task 1 (Ent schema creation)
- **Issue:** Invite schema references User via created_by edge, but User schema needed a corresponding outbound edge
- **Fix:** Added `edge.To("invites", Invite.Type)` to User schema
- **Files modified:** backend/ent/schema/user.go
- **Verification:** go generate and go build pass
- **Committed in:** 0706d3f (Task 1 commit)

**2. [Rule 2 - Missing Critical] Created helpers.go for shared regex**
- **Found during:** Task 1 (Project schema)
- **Issue:** Project key regex validation needed a shared compiled regexp to avoid duplication
- **Fix:** Created backend/ent/schema/helpers.go with keyRegexp variable
- **Files modified:** backend/ent/schema/helpers.go
- **Verification:** go generate and go build pass
- **Committed in:** 0706d3f (Task 1 commit)

---

**Total deviations:** 2 auto-fixed (2 missing critical)
**Impact on plan:** Both fixes necessary for Ent code generation to succeed. No scope creep.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All Ent schemas ready for auth module (Magic Link, Telegram, JWT sessions)
- RBAC middleware ready to protect API endpoints
- Storage service ready for file upload integration
- Outbox pattern ready for cross-database consistency
- Platform clients ready for server initialization in main.go

## Self-Check: PASSED

All 24 created files verified present. Both task commits (0706d3f, a6a1e76) verified in git history.
