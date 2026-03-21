---
gsd_state_version: 1.0
milestone: v1.0
milestone_name: milestone
status: unknown
stopped_at: Completed 06-02-PLAN.md
last_updated: "2026-03-21T04:12:37.914Z"
progress:
  total_phases: 7
  completed_phases: 5
  total_plans: 33
  completed_plans: 31
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-03-20)

**Core value:** Unified environment for dev teams -- tracker, TMS, messenger, and calls connected in one interface
**Current focus:** Phase 06 — Notifications, Search & Dashboards

## Current Position

Phase: 06 (Notifications, Search & Dashboards) — EXECUTING
Plan: 4 of 5

## Performance Metrics

**Velocity:**

- Total plans completed: 0
- Average duration: -
- Total execution time: 0 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| - | - | - | - |

**Recent Trend:**

- Last 5 plans: -
- Trend: -

*Updated after each plan completion*
| Phase 01 P01 | 9min | 2 tasks | 37 files |
| Phase 01 P02 | 5min | 2 tasks | 29 files |
| Phase 01 P03 | 7min | 2 tasks | 11 files |
| Phase 01 P04 | 10min | 2 tasks | 8 files |
| Phase 01 P05 | 7min | 3 tasks | 18 files |
| Phase 01 P07 | 12min | 2 tasks | 32 files |
| Phase 01 P06 | 6min | 2 tasks | 3 files |
| Phase 01 P08 | 2min | 2 tasks | 2 files |
| Phase 01 P09 | 5min | 2 tasks | 5 files |
| Phase 02 P01 | 5min | 2 tasks | 19 files |
| Phase 02 P03 | 4min | 2 tasks | 4 files |
| Phase 02 P02 | 12min | 4 tasks | 7 files |
| Phase 02 P04 | 8min | 4 tasks | 15 files |
| Phase 02 P05 | 7min | 2 tasks | 3 files |
| Phase 03 P01 | 10min | 2 tasks | 17 files |
| Phase 03 P02 | 5min | 2 tasks | 6 files |
| Phase 03 P03 | 7min | 2 tasks | 21 files |
| Phase 03 P04 | 7min | 2 tasks | 20 files |
| Phase 03 P05 | 5min | 2 tasks | 3 files |
| Phase 04 P01 | 10min | 2 tasks | 13 files |
| Phase 04 P02 | 10min | 2 tasks | 14 files |
| Phase 04 P03 | 7min | 2 tasks | 12 files |
| Phase 04 P04 | 15min | 2 tasks | 19 files |
| Phase 04 P05 | 8min | 2 tasks | 3 files |
| Phase 05 P01 | 7min | 2 tasks | 15 files |
| Phase 05 P02 | 4min | 2 tasks | 11 files |
| Phase 05 P03 | 5min | 2 tasks | 6 files |
| Phase 05 P04 | 4min | 2 tasks | 3 files |
| Phase 06 P01 | 12min | 2 tasks | 22 files |
| Phase 06 P02 | 12min | 2 tasks | 17 files |
| Phase 06 P04 | 10min | 2 tasks | 25 files |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- MinIO replaced by SeaweedFS (MinIO archived Feb 2026) -- S3 API compatible, no code changes needed
- NATS for pub/sub + JetStream for durable jobs; Redis only for cache/sessions
- Ent ORM Privacy Layer for RBAC enforcement
- [Phase 01]: Vite 6 instead of Vite 8 due to Node.js 22.11 version constraint
- [Phase 01]: Fiber Immutable=true in dev to catch RequestCtx reuse (Pitfall 1)
- [Phase 01]: In-memory sliding window rate limiter (no Redis startup dependency)
- [Phase 01]: RBAC uses uint64 bitmap with 26 permissions for all Phase 1 modules
- [Phase 01]: Outbox pattern: PG outbox table + 100ms polling worker + NATS publish
- [Phase 01]: S3 UsePathStyle=true required for SeaweedFS compatibility
- [Phase 01]: golang-jwt/jwt/v5 for HS256 JWT signing
- [Phase 01]: Auth middleware in auth package (not RBAC) for clean JWT validation
- [Phase 01]: Telegram users get placeholder email until onboarding
- [Phase 01]: Repository interface pattern for project module enables mock-based unit testing without database
- [Phase 01]: RBAC cache invalidation on every membership mutation via CacheInvalidator interface
- [Phase 01]: fasthttp/websocket directly instead of gofiber/contrib/websocket (contrib only supports Fiber v2)
- [Phase 01]: JWT auth for WebSocket via query param (WS handshake lacks custom header support)
- [Phase 01]: TanStack Router file-based routing with _authenticated pathless layout for auth guard
- [Phase 01]: All Phase 1 documentation written in Russian with TC-{REQ_ID}-{NN} test case naming convention
- [Phase 01]: Auth middleware injected as parameter to RegisterRoutes for clean dependency inversion
- [Phase 01]: AddTeamToProject bulk-adds members with developer role, silently skipping duplicates
- [Phase 02]: User schema extended with tracker edges (comments, reactions, activity_logs, time_logs) for Ent back-reference support
- [Phase 02]: FieldChange struct duplicated in schema and tracker packages to keep packages decoupled
- [Phase 02]: Plate v49 API uses usePlateEditor + editor prop instead of value prop on Plate component
- [Phase 02]: Value type from @udecode/plate (not /react), components from @udecode/plate/react
- [Phase 02]: Mock repository for tracker unit tests instead of testcontainers
- [Phase 02]: WorkflowEngine as separate struct for SRP
- [Phase 02]: Board data via nested eager loading (Board->Project->Statuses->Issues)
- [Phase 02]: Refactored $projectId.tsx to layout route with Outlet for child routes
- [Phase 02]: Native select elements used instead of shadcn Select to minimize deps
- [Phase 02]: Russian documentation with TC-{REQ_ID}-{NN} test case naming convention for Phase 2
- [Phase 03]: CombinedRepository pattern: delegates to EntRepository (PG) + MongoRepository (MongoDB) behind single Repository interface
- [Phase 03]: DM deduplication via sorted user IDs concatenated with dash as dm_key (unique index)
- [Phase 03]: Nil-safe NATS publishEvent for testability without mock NATS client
- [Phase 03]: AutoChannelService uses core NATS subscription for project.created events from outbox worker
- [Phase 03]: AddMembersInternal method bypasses permission checks for system-level operations
- [Phase 03]: Inline Plate JSON renderer for read-only message display instead of PlateStatic
- [Phase 03]: useRef pattern for WS event handlers to avoid stale closures
- [Phase 03]: miniredis for Redis unit tests (fast in-process, no Docker)
- [Phase 03]: Presence: 5min away threshold, 90s stale cleanup, 60s cleanup interval
- [Phase 03]: Bookmarks via dialog button (not separate route) in channel list header
- [Phase 04]: TMS EntRepository without OutboxPublisher; offset pagination for TMS; suite tree built in-memory; snapshot versioning with JSON map
- [Phase 04]: Frontend TMS files in src/api/ and src/stores/ (not src/shared/) following project structure
- [Phase 04]: TMS route layout with tabs under $projectId/tms/; execution panel as inline Card (not modal)
- [Phase 04]: recomputeRunStats queries all results after each update for accuracy
- [Phase 04]: CombinedRepository pattern reused from messenger for TMS dual PG+Mongo storage
- [Phase 04]: NATS client in TMSService constructor (nil-safe for testability)
- [Phase 04]: Flaky detection: status variance across last 10 runs per test case
- [Phase 04]: Shared Ent client for cross-module bug creation (no tracker import)
- [Phase 04]: excelize v2 for Excel, maroto v2 for PDF in TMS import/export
- [Phase 04]: Auto-detect field mapping for import with Russian+English header support
- [Phase 04]: Russian documentation following Phase 1 pattern with TC-{REQ_ID}-{NN} test case naming for Phase 4
- [Phase 05]: Manual webhook JWT verification instead of webhook.ReceiveWebhookEvent (Fiber/fasthttp incompatible)
- [Phase 05]: Nil-safe NATS publishing pattern reused from messenger module for calls
- [Phase 05]: Custom CallControls instead of ControlBar for raise hand button support
- [Phase 05]: Ephemeral call chat: MongoDB with auto-delete on room_finished via ChatRepository.DeleteCallChat
- [Phase 05]: Call events routed to channel:{channelID} WS rooms with call.* prefix; chat to call:{callID} rooms
- [Phase 05]: Russian documentation following established Phase 1-4 pattern with TC-CALL-{ID}-{NN} test case naming for Phase 5
- [Phase 06]: JetStream durable consumer for notification delivery with 2s debounce dedup; nil-safe TelegramSender; quiet hours with midnight crossover; reuse auth bot for notifications
- [Phase 06]: Raw SQL via database.OpenRawDB() for tsvector queries (Ent ORM limitation)
- [Phase 06]: Parallel PG tsvector + MongoDB text search with sort-by-score merge
- [Phase 06]: cmdk for command palette with shouldFilter=false (server-side filtering)

### Pending Todos

None yet.

### Blockers/Concerns

- RBAC permission model (system + project + channel scopes) needs design spike in Phase 1
- Fiber/fasthttp RequestCtx reuse -- must establish copy-before-goroutine conventions immediately
- PG + MongoDB cross-database consistency requires outbox pattern from Phase 1

## Session Continuity

Last session: 2026-03-21T04:12:22.830Z
Stopped at: Completed 06-02-PLAN.md
Resume file: None
