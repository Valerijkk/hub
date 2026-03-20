# Architecture Research

**Domain:** Team collaboration hub (task tracker + messenger + TMS + voice calls)
**Researched:** 2026-03-20
**Confidence:** HIGH

## Standard Architecture

### System Overview

```
                              +-----------+
                              |   Caddy   |
                              | (reverse  |
                              |  proxy)   |
                              +-----+-----+
                                    |
                    +---------------+---------------+
                    |               |               |
              +-----+-----+  +-----+-----+  +------+------+
              |  React 19 |  |  REST API  |  |  WebSocket  |
              |    SPA     |  | (Fiber)    |  |  Gateway    |
              +-----+-----+  +-----+-----+  +------+------+
                    |               |               |
                    +-------+-------+-------+-------+
                            |               |
                     +------+------+  +-----+-----+
                     |    NATS     |  |   Redis    |
                     | JetStream  |  |  (cache)   |
                     +------+------+  +-----+-----+
                            |               |
         +------------------+------------------+
         |          |           |          |    |
    +----+----+ +---+----+ +---+----+ +---+--+ +-------+
    | Tracker | | Messgr | |  TMS   | | Calls| | Auth  |
    | Module  | | Module | | Module | | Mod  | | Module|
    +----+----+ +---+----+ +---+----+ +---+--+ +---+---+
         |          |           |          |        |
    +----+----------+-----------+----------+--------+----+
    |                PostgreSQL (Ent ORM)                 |
    +----------------------------------------------------+
    |           MongoDB (chat history, Allure JSON)       |
    +----------------------------------------------------+
    |           MinIO (files, recordings, avatars)        |
    +----------------------------------------------------+
    |           LiveKit (WebRTC SFU, screen share)        |
    +----------------------------------------------------+
```

### Component Responsibilities

| Component | Responsibility | Implementation |
|-----------|----------------|----------------|
| **Caddy** | TLS termination, reverse proxy, WebSocket upgrade | Caddyfile config, auto-HTTPS |
| **REST API (Fiber)** | HTTP request handling, routing, middleware, validation | Go Fiber v2, route groups per module |
| **WebSocket Gateway** | Persistent connections, real-time event fan-out | coder/websocket, hub pattern with NATS bridge |
| **NATS JetStream** | Event bus, async job processing, cross-module pub/sub | Subjects per domain event, durable consumers |
| **Redis** | Session store, cache, magic link tokens, online presence | go-redis, TTL-based expiry |
| **Tracker Module** | Tasks, sprints, kanban boards, workflows, time tracking | Ent schemas, PostgreSQL |
| **Messenger Module** | Channels, DMs, threads, reactions, message search | MongoDB for messages, PostgreSQL for channel metadata |
| **TMS Module** | Test cases, test runs, Allure integration, traceability | PostgreSQL for cases, MongoDB for Allure JSON reports |
| **Calls Module** | Voice rooms, screen share, recording management | LiveKit Go SDK, MinIO for recordings |
| **Auth Module** | Magic link, Telegram auth, sessions, RBAC | Redis for tokens, PostgreSQL for users/roles |
| **MinIO** | File storage for all modules (attachments, avatars, recordings) | S3 API, presigned URLs for direct browser upload/download |
| **LiveKit** | WebRTC SFU, media routing, recording via Egress | Self-hosted, Go-native, horizontal scaling via Redis |

## Recommended Project Structure

### Backend (Go Modular Monolith)

```
hub-backend/
+-- cmd/
|   +-- server/
|       +-- main.go              # Entry point, wiring
+-- internal/
|   +-- auth/                    # Auth module
|   |   +-- handler.go           # Fiber route handlers
|   |   +-- service.go           # Business logic
|   |   +-- repository.go        # Data access interface
|   |   +-- middleware.go         # Auth middleware
|   |   +-- dto.go               # Request/response types
|   |   +-- events.go            # NATS event definitions
|   +-- tracker/                 # Task tracker module
|   |   +-- handler.go
|   |   +-- service.go
|   |   +-- repository.go
|   |   +-- workflow.go          # State machine for task statuses
|   |   +-- dto.go
|   |   +-- events.go
|   +-- messenger/               # Messenger module
|   |   +-- handler.go
|   |   +-- service.go
|   |   +-- repository.go        # MongoDB operations
|   |   +-- dto.go
|   |   +-- events.go
|   +-- tms/                     # Test management module
|   |   +-- handler.go
|   |   +-- service.go
|   |   +-- repository.go
|   |   +-- allure.go            # Allure results parser
|   |   +-- dto.go
|   |   +-- events.go
|   +-- calls/                   # Voice calls module
|   |   +-- handler.go
|   |   +-- service.go
|   |   +-- livekit.go           # LiveKit SDK wrapper
|   |   +-- dto.go
|   |   +-- events.go
|   +-- project/                 # Project management (cross-cutting)
|   |   +-- handler.go
|   |   +-- service.go
|   |   +-- repository.go
|   |   +-- dto.go
|   +-- notification/            # Notification dispatcher
|   |   +-- service.go
|   |   +-- telegram.go          # Telegram bot integration
|   |   +-- hub.go               # In-app notification delivery
|   |   +-- events.go            # NATS consumers for all modules
|   +-- storage/                 # MinIO file operations
|   |   +-- service.go           # Presigned URL generation, bucket mgmt
|   +-- search/                  # Global search (Ctrl+K)
|   |   +-- service.go
|   |   +-- indexer.go
|   +-- dashboard/               # Dashboard data aggregation
|   |   +-- handler.go
|   |   +-- service.go
|   +-- ws/                      # WebSocket gateway
|   |   +-- hub.go               # Connection registry, fan-out
|   |   +-- handler.go           # Upgrade handler
|   |   +-- bridge.go            # NATS-to-WebSocket bridge
|   +-- platform/                # Shared infrastructure
|       +-- database/            # DB connections (PG + Mongo)
|       +-- nats/                # NATS client setup
|       +-- redis/               # Redis client
|       +-- middleware/          # Common middleware (logging, CORS, etc.)
|       +-- config/              # Configuration loading
|       +-- errors/              # Error types
+-- ent/                         # Ent schema definitions (codegen)
|   +-- schema/
|   |   +-- user.go
|   |   +-- project.go
|   |   +-- task.go
|   |   +-- testcase.go
|   |   +-- ...
|   +-- generate.go
+-- migrations/                  # Atlas migration files
+-- api/                         # OpenAPI specs, Swagger
+-- deployments/
|   +-- docker-compose.yml
|   +-- k3s/
+-- Makefile
+-- go.mod
```

### Frontend (React 19)

```
hub-frontend/
+-- src/
|   +-- app/                     # App shell, routing, providers
|   +-- modules/
|   |   +-- auth/                # Login, magic link, profile
|   |   +-- tracker/             # Kanban, backlog, task detail
|   |   +-- messenger/           # Channel list, chat, threads
|   |   +-- tms/                 # Test cases, runs, reports
|   |   +-- calls/               # Call room, screen share UI
|   |   +-- dashboard/           # Personal, project, admin dashboards
|   |   +-- project/             # Project settings, members, roles
|   +-- shared/
|   |   +-- components/          # Shadcn/ui components
|   |   +-- hooks/               # useWebSocket, useAuth, etc.
|   |   +-- api/                 # API client (fetch wrapper)
|   |   +-- stores/              # Zustand stores (global state)
|   |   +-- lib/                 # Utilities
|   +-- i18n/                    # ru/en translations
+-- public/
+-- vite.config.ts
```

### Structure Rationale

- **internal/[module]/**: Each domain module is self-contained with handler, service, repository, and events. Modules communicate only through NATS events, never by importing each other's internals. This enforces clean boundaries and enables future extraction to microservices if needed.
- **internal/platform/**: Shared infrastructure code (DB connections, NATS client, middleware) that any module can import. This is the only cross-cutting dependency allowed.
- **ent/schema/**: Centralized Ent schemas because PostgreSQL is shared. Ent's graph-based schema definition naturally models the relationships between tasks, test cases, users, and projects.
- **internal/ws/**: Single WebSocket gateway that bridges NATS events to browser clients. Modules never touch WebSocket connections directly -- they publish to NATS and the gateway fans out to relevant subscribers.
- **Frontend modules/**: Mirror backend module boundaries. Each module owns its pages, components, and module-specific state. Shared UI components live in shared/.

## Architectural Patterns

### Pattern 1: Modular Monolith with Event Bus

**What:** Single deployable binary with internal module boundaries enforced via NATS events. Modules share the process but communicate asynchronously for cross-cutting concerns.

**When to use:** Teams of 10-50, single deployment target, need for simplicity without sacrificing modularity.

**Trade-offs:**
- PRO: Single deploy, shared database transactions within a module, simple debugging
- PRO: Modules can be extracted to microservices later by replacing in-process NATS with network NATS
- CON: All modules scale together (acceptable at 10-50 users)
- CON: A bug in one module can crash the whole process

**Example:**
```go
// tracker/service.go -- publishes event when task status changes
func (s *TrackerService) UpdateTaskStatus(ctx context.Context, taskID int, newStatus string) error {
    task, err := s.repo.UpdateStatus(ctx, taskID, newStatus)
    if err != nil {
        return err
    }
    // Publish to NATS -- messenger, notification, TMS modules subscribe
    event := events.TaskStatusChanged{
        TaskID:    task.ID,
        ProjectID: task.ProjectID,
        OldStatus: task.PreviousStatus,
        NewStatus: newStatus,
        ChangedBy: auth.UserFromCtx(ctx).ID,
    }
    return s.nats.Publish("tracker.task.status_changed", event)
}

// notification/events.go -- consumes events from any module
func (n *NotificationService) SubscribeAll() {
    n.nats.Subscribe("tracker.task.*", n.handleTrackerEvent)
    n.nats.Subscribe("tms.run.*", n.handleTMSEvent)
    n.nats.Subscribe("messenger.mention.*", n.handleMentionEvent)
}
```

### Pattern 2: WebSocket Gateway with NATS Bridge

**What:** A dedicated WebSocket hub manages all persistent browser connections. It subscribes to NATS subjects and fans out events to relevant clients. Clients are mapped to their subscriptions (projects, channels, rooms) in memory.

**When to use:** Any real-time feature -- chat messages, task updates, notifications, call signaling, online presence.

**Trade-offs:**
- PRO: Modules are decoupled from WebSocket transport; they just publish NATS events
- PRO: Horizontally scalable -- multiple WS gateway instances, NATS handles cross-instance routing
- CON: Extra hop (module -> NATS -> WS gateway -> browser)

**Example:**
```go
// ws/hub.go
type Hub struct {
    clients    map[string]*Client       // userID -> client
    rooms      map[string]map[string]bool // roomKey -> set of userIDs
    register   chan *Client
    unregister chan *Client
}

// ws/bridge.go -- bridges NATS to WebSocket
func (b *Bridge) Start() {
    // Subscribe to all events that need real-time delivery
    b.nats.Subscribe("messenger.channel.>", func(msg *nats.Msg) {
        channelID := extractChannelID(msg.Subject)
        roomKey := "channel:" + channelID
        b.hub.BroadcastToRoom(roomKey, msg.Data)
    })
    b.nats.Subscribe("tracker.project.>", func(msg *nats.Msg) {
        projectID := extractProjectID(msg.Subject)
        roomKey := "project:" + projectID
        b.hub.BroadcastToRoom(roomKey, msg.Data)
    })
}
```

### Pattern 3: Presigned URL File Handling

**What:** Backend never proxies file bytes. It generates short-lived presigned URLs, and the browser uploads/downloads directly to/from MinIO. Backend only stores metadata (filename, size, bucket, key, uploader, linked entity).

**When to use:** All file operations -- attachments in tasks, messages, test cases, voice recordings, avatars.

**Trade-offs:**
- PRO: Backend is not a bottleneck for large files; MinIO handles transfer
- PRO: S3-compatible -- migrate to AWS S3 by changing endpoint
- CON: Requires CORS config on MinIO; slightly more complex client code

**Example flow:**
```
1. Client requests: POST /api/files/upload-url
   Body: { filename: "screenshot.png", contentType: "image/png", entityType: "task", entityID: 123 }

2. Backend validates permissions, generates presigned PUT URL (5 min TTL)
   Response: { uploadURL: "https://minio.hub.local/...", fileID: "uuid" }

3. Client PUTs file directly to MinIO using presigned URL

4. Client confirms: POST /api/files/{fileID}/confirm
   Backend verifies file exists in MinIO, saves metadata to PostgreSQL

5. For download: GET /api/files/{fileID}/url
   Backend generates presigned GET URL (1 hour TTL)
   Response: { downloadURL: "https://minio.hub.local/..." }
```

### Pattern 4: NATS Subject Hierarchy for Domain Events

**What:** Structured NATS subject naming that maps to domain events. JetStream streams persist critical events; core NATS handles ephemeral ones.

**When to use:** All inter-module communication and async processing.

**Subject structure:**
```
{module}.{entity}.{action}

Examples:
  tracker.task.created           # New task
  tracker.task.status_changed    # Task moved to new status
  tracker.task.assigned          # Task assigned to user
  messenger.message.sent         # New chat message
  messenger.channel.joined       # User joined channel
  tms.run.completed              # Test run finished
  tms.case.linked                # Test case linked to task
  calls.room.started             # Voice call started
  calls.recording.ready          # Recording uploaded to MinIO
  auth.user.logged_in            # User login event
  project.member.added           # New project member
```

**JetStream streams (persisted):**
```
TRACKER   -- tracker.>        # Task changes need audit trail
TMS       -- tms.>            # Test results are permanent records
AUDIT     -- *.*.created, *.*.deleted, *.*.updated  # Audit log
```

**Core NATS (ephemeral):**
```
messenger.typing.*             # Typing indicators (fire-and-forget)
presence.*                     # Online/offline status
calls.signal.*                 # Call signaling
```

## Data Flow

### Cross-Module Integration Flows

```
[Task Status Changed]
    |
    +-> NATS: tracker.task.status_changed
         |
         +-> Notification Module -> In-app bell + Telegram bot
         |
         +-> WS Gateway -> Browser (real-time board update)
         |
         +-> TMS Module -> Update linked test run status
         |
         +-> Audit Module -> Write to audit log (JetStream)

[Test Run Completed (from CI/Allure)]
    |
    +-> POST /api/tms/runs (Allure results JSON)
         |
         +-> TMS Module: parse, match test cases, store in MongoDB
         |
         +-> NATS: tms.run.completed
              |
              +-> Tracker Module -> Update linked task coverage
              |
              +-> Notification Module -> Notify assignees
              |
              +-> Messenger Module -> Post summary to #project-ci channel
              |
              +-> WS Gateway -> Browser (dashboard update)

[Chat Message Sent]
    |
    +-> WS: client sends message
         |
         +-> WS Gateway -> validates, stores in MongoDB via Messenger Module
         |
         +-> NATS: messenger.message.sent
              |
              +-> Notification Module -> @mention notifications
              |
              +-> WS Gateway -> fan out to channel subscribers
              |
              +-> Search Indexer -> update search index

[Voice Call Recording Ready]
    |
    +-> LiveKit Egress -> uploads to MinIO
         |
         +-> Webhook -> Calls Module
              |
              +-> Save recording metadata to PostgreSQL
              |
              +-> NATS: calls.recording.ready
                   |
                   +-> Messenger Module -> Post link in project channel
                   |
                   +-> Notification Module -> Notify participants
```

### Database Ownership

```
PostgreSQL (via Ent ORM):
  +-- auth: users, roles, permissions, sessions
  +-- project: projects, members, teams, invites
  +-- tracker: tasks, sprints, boards, workflows, comments, time_logs
  +-- tms: test_cases, test_suites, test_runs (metadata), environments
  +-- calls: call_rooms, call_participants, recordings (metadata)
  +-- storage: file_metadata (all modules reference this)
  +-- notification: notification_preferences, notification_log
  +-- audit: audit_log

MongoDB:
  +-- messenger: messages, threads, reactions (high write volume, flexible schema)
  +-- tms: allure_results (raw JSON, deeply nested, variable structure)
  +-- dashboard: widget_configs (user-customized, flexible schema)
```

### State Management (Frontend)

```
Zustand Stores:
  +-- authStore: current user, session, permissions
  +-- wsStore: WebSocket connection state, reconnection logic
  +-- trackerStore: active board tasks, filters (per-project)
  +-- messengerStore: active channel messages, unread counts
  +-- notificationStore: bell notifications, unread count
  +-- callStore: active call state, participants, mute status

React Query (TanStack Query):
  +-- Server state: task lists, test cases, project data
  +-- Cache invalidation via WebSocket events
  +-- Optimistic updates for task drag-and-drop, message sending
```

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| 10-50 users (target) | Single instance of everything. Modular monolith. Docker Compose. This is the design target -- no scaling heroics needed. |
| 50-200 users | Multiple Go backend replicas behind Caddy. NATS handles cross-instance WS fan-out. LiveKit scales horizontally via Redis. PostgreSQL read replicas for dashboards. |
| 200-1000 users | Extract messenger into separate service (highest write volume). MongoDB sharding. Dedicated NATS cluster. Consider splitting WS gateway. |

### Scaling Priorities

1. **First bottleneck: WebSocket connections.** At ~100 concurrent users, a single Go process handles this easily (25K+ connections possible). Not a real concern at target scale.
2. **Second bottleneck: MongoDB write throughput.** Chat messages are the highest-volume write operation. At target scale (10-50 users), this is a non-issue. Index on channelID + timestamp covers query patterns.
3. **Third bottleneck: PostgreSQL query complexity.** Dashboard aggregations across tasks, sprints, test runs. Solution: materialized views or pre-computed counters updated via NATS events.

## Anti-Patterns

### Anti-Pattern 1: Direct Module-to-Module Imports

**What people do:** Tracker module directly imports messenger module to post a message when a task changes.
**Why it's wrong:** Creates tight coupling, circular dependencies, makes it impossible to modify or extract modules independently.
**Do this instead:** Tracker publishes `tracker.task.status_changed` to NATS. Messenger subscribes and posts to the channel. Neither knows about the other's internals.

### Anti-Pattern 2: Proxying Files Through the Backend

**What people do:** Client uploads file to backend, backend stores it in MinIO. Downloads go through backend too.
**Why it's wrong:** Backend becomes bottleneck for large files, wastes memory and CPU on byte-shuffling.
**Do this instead:** Use presigned URLs. Backend generates the URL, client talks directly to MinIO. Backend only manages metadata.

### Anti-Pattern 3: Putting Chat Messages in PostgreSQL

**What people do:** Store every chat message as a row in PostgreSQL with full relational model.
**Why it's wrong:** Chat is insert-heavy with append-only access patterns. Relational overhead (transactions, constraints) adds no value. Schema changes (reactions, attachments, formatting) require migrations.
**Do this instead:** MongoDB for messages. Document model naturally handles variable message structures (text, reactions, attachments, threads). PostgreSQL for channel metadata (members, permissions).

### Anti-Pattern 4: WebSocket Per Module

**What people do:** Each module (tracker, messenger, calls) opens its own WebSocket connection to the browser.
**Why it's wrong:** Multiple connections per client waste resources, complicate authentication, and create reconnection nightmares.
**Do this instead:** Single WebSocket gateway. Client connects once, subscribes to "rooms" (channels, project boards, calls). NATS bridge routes events from all modules through the single connection.

### Anti-Pattern 5: Synchronous Cross-Module Calls for Notifications

**What people do:** After updating a task, synchronously call notification service, then Telegram service, then audit service in sequence.
**Why it's wrong:** Slow (total latency = sum of all calls), fragile (one failure blocks the response), couples everything.
**Do this instead:** Publish event to NATS, return response immediately. Notification, Telegram, and audit consumers process independently and asynchronously.

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| **LiveKit** | Go Server SDK for room creation and token generation; React SDK (@livekit/components-react) for UI | LiveKit runs as separate container. Backend creates rooms and issues JWT tokens. Frontend connects directly to LiveKit via WebRTC. Egress service handles recording to MinIO. |
| **Telegram Bot** | telego library, webhook mode behind Caddy | Bot handles: magic link auth, notifications, slash commands. Webhook endpoint in auth/notification modules. |
| **GitHub** | Webhooks (incoming) + GitHub REST API (outgoing) | Webhook receives push/PR events, matches "HUB-123" patterns. Outgoing API for branch creation, PR status. Store GitHub app credentials in config. |
| **Allure CI** | hub-cli pushes allure-results JSON to TMS API | CI pipeline: run tests -> generate allure-results -> hub-cli POST /api/tms/runs/import. No Allure server needed. |
| **MinIO** | minio-go SDK v7, presigned URLs | Buckets: attachments, avatars, recordings, exports. Lifecycle policies for temp uploads. CORS headers for browser direct upload. |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| Tracker <-> TMS | NATS events + shared PostgreSQL (Ent relations) | Task-TestCase link is an Ent edge. Events trigger coverage recalculation. |
| Tracker <-> Messenger | NATS events only | Task changes trigger channel posts. No direct dependency. |
| TMS <-> Messenger | NATS events only | Test run completion posts summary to project channel. |
| Calls <-> Messenger | NATS events only | Call start/end posts to channel. Recording link posted when ready. |
| All Modules <-> Notification | NATS events (consumers) | Notification module subscribes to events from every module. Dispatches to in-app bell and Telegram based on user preferences. |
| All Modules <-> Audit | NATS JetStream (durable) | Every create/update/delete event persisted to audit stream. Audit module consumes and writes to audit_log table. |
| All Modules <-> Storage | Direct function call (same process) | Storage service is shared infrastructure, not a domain module. Any module calls `storage.GenerateUploadURL()`. |
| All Modules <-> WS Gateway | NATS events -> WS bridge | Modules never touch WebSocket directly. They publish to NATS, bridge delivers to browsers. |

## Suggested Build Order

Based on dependency analysis, the recommended build sequence:

```
Phase 1: Foundation
  +-- Auth (users, sessions, magic link) -- everything depends on this
  +-- Project (RBAC, teams) -- scopes all other modules
  +-- Platform (DB connections, NATS, Redis, config)
  +-- Storage (MinIO integration) -- needed by all modules for files
  +-- WebSocket Gateway + NATS bridge -- real-time backbone

Phase 2: Core Modules (can be parallel after Phase 1)
  +-- Tracker (tasks, boards, workflows) -- highest business value
  +-- Messenger (channels, messages, threads) -- daily use feature

Phase 3: Extended Modules
  +-- TMS (test cases, runs, Allure) -- depends on Tracker for linking
  +-- Notification (in-app + Telegram) -- depends on all modules emitting events

Phase 4: Rich Features
  +-- Calls (LiveKit integration) -- independent but lower priority than core
  +-- Dashboards (aggregation across modules) -- needs data from everything
  +-- Search (global Ctrl+K) -- needs indexing from all modules

Phase 5: Polish
  +-- GitHub integration (webhooks, auto-linking)
  +-- Import/Export
  +-- Audit log viewer
  +-- Admin panel
  +-- hub-cli
```

**Build order rationale:**
- Auth and Project are foundational -- every API endpoint needs authentication and project scoping
- WebSocket gateway in Phase 1 because both Tracker (board updates) and Messenger (chat) need it from day one
- Tracker before Messenger because task management is the primary differentiator; messenger without tasks is just another Slack clone
- TMS after Tracker because test case linking requires tasks to exist
- Notifications in Phase 3 because they consume events from modules that must exist first
- Calls in Phase 4 because LiveKit is a self-contained integration, not a dependency for other modules
- Dashboards last because they aggregate data from everything else

## Sources

- [NATS JetStream Documentation](https://docs.nats.io/nats-concepts/jetstream)
- [NATS WebSocket Client (nats.ws)](https://github.com/nats-io/nats.ws)
- [Using NATS to build a scalable WebSocket server](https://medium.com/@14domino/using-nats-to-build-a-very-functional-websocket-server-8f4de44c3f12)
- [Building Event-Driven Systems with Go and NATS](https://dasroot.net/posts/2026/02/building-event-driven-systems-go-nats/)
- [LiveKit Architecture (SFU)](https://docs.livekit.io/reference/internals/livekit-sfu/)
- [LiveKit Documentation](https://docs.livekit.io/intro/overview/)
- [MinIO Presigned Operations](https://deepwiki.com/minio/minio-go/5.2-presigned-operations)
- [MinIO Go Client SDK](https://pkg.go.dev/github.com/minio/minio-go/v7)
- [Go Fiber Modular Monolith](https://medium.com/@hocineelhadj/building-a-simple-backend-with-modular-monolith-architecture-in-go-e2ec7b59bc58)
- [Go Modular Monolith Boilerplate (gomomo)](https://github.com/ArkjuniorK/gomomo)
- [coder/websocket (recommended Go WebSocket library)](https://websocket.org/guides/languages/go/)
- [Scalable Go WebSocket Service](https://leapcell.io/blog/building-a-scalable-go-websocket-service-for-thousands-of-concurrent-connections)
- [Ent Framework](https://pkg.go.dev/entgo.io/ent)
- [Go Project Structure Best Practices](https://www.glukhov.org/post/2025/12/go-project-structure/)

---
*Architecture research for: Team collaboration hub (Hub)*
*Researched: 2026-03-20*
