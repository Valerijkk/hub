# Project Research Summary

**Project:** Hub
**Domain:** Team collaboration hub (task tracker + messenger + TMS + voice calls)
**Researched:** 2026-03-20
**Confidence:** HIGH

## Executive Summary

Hub is a self-hosted unified platform for development teams (10-50 people) that replaces four separate tools -- Jira (task tracker), Slack (messenger), TestRail/Allure TestOps (TMS), and Google Meet (voice calls) -- with a single integrated product. The core differentiator is cross-module traceability: tasks link to test cases, CI test results flow into the TMS via Allure integration, and all events surface in the messenger and Telegram. The expert approach for this class of product is a Go modular monolith with NATS as the internal event bus, PostgreSQL for relational data, MongoDB for append-heavy chat and Allure JSON, and LiveKit for WebRTC voice calls. The closest competitor is Huly, which combines tracker + messenger + calls but lacks a TMS module entirely.

The recommended architecture is a single Go binary with strict module boundaries (auth, tracker, messenger, TMS, calls) communicating exclusively through NATS events, fronted by a React 19 SPA with a single WebSocket gateway bridging NATS to the browser. This pattern supports the target scale (10-50 users) without operational complexity while remaining extractable to microservices if needed. Key infrastructure choices are well-validated: Fiber v3 (fasthttp performance), Ent ORM (graph-based entity traversal for task hierarchies and RBAC privacy layer), SeaweedFS (S3-compatible storage replacing the now-archived MinIO), and Caddy (auto-HTTPS reverse proxy). The frontend stack is React 19 + Vite 8 + shadcn/ui + Tailwind CSS 4, with Plate for rich text, pragmatic-drag-and-drop for Kanban boards, and TanStack Query/Router for data and routing.

The primary risks are: (1) Fiber's fasthttp RequestCtx reuse causing silent data corruption in goroutines -- must establish copy-before-goroutine conventions from day one; (2) PostgreSQL + MongoDB cross-database consistency gaps requiring an outbox pattern for reliable cross-store operations; (3) RBAC role explosion across system, project, and channel scopes -- the permission model must be designed upfront with cached effective permissions; and (4) the sheer breadth of the product (four integrated modules) creating scope pressure that could delay MVP indefinitely. The mitigation for scope risk is a disciplined MVP that delivers all four modules at table-stakes depth, proving the integration value before adding differentiating features.

## Key Findings

### Recommended Stack

The stack is a Go + React split with a polyglot persistence layer. Go 1.26 with Fiber v3 provides the backend with first-class concurrency for real-time features. The data layer uses PostgreSQL 17 (via Ent ORM) for structured data, MongoDB 8 for chat messages and Allure JSON, Redis 7.4 for sessions/cache, and NATS with JetStream for event-driven communication. SeaweedFS replaces MinIO (archived Feb 2026) for S3-compatible object storage. LiveKit handles WebRTC voice calls and screen sharing.

**Core technologies:**
- **Go 1.26 + Fiber v3**: Backend language and framework -- goroutines for real-time, fasthttp for throughput, single binary deployment
- **Ent ORM + Atlas**: PostgreSQL ORM with code-generated type-safe queries, graph traversal for task hierarchies, privacy layer for RBAC, Atlas for versioned migrations
- **NATS + JetStream**: Event bus replacing both RabbitMQ and Redis pub/sub -- core NATS for real-time fan-out, JetStream for durable job processing
- **React 19 + Vite 8 + TypeScript**: Frontend with Server Components available but SPA mode sufficient; Vite 8 ships Rolldown for 10-30x faster builds
- **shadcn/ui + Tailwind CSS 4 + Plate**: UI components (copy-paste, full control), CSS-first config, shadcn-native rich text editor with free mentions and tables
- **LiveKit**: Go-native WebRTC SFU for voice calls + screen share, self-hosted single container, Egress for recording
- **SeaweedFS**: S3-compatible object storage replacing archived MinIO; presigned URLs for direct browser upload/download
- **Caddy**: Reverse proxy with auto-HTTPS, native WebSocket proxying, minimal config

**Critical version note:** MinIO was archived Feb 2026. All references to MinIO in PROJECT.md constraints should be updated to SeaweedFS. The S3 API is compatible -- no code changes needed beyond endpoint configuration.

### Expected Features

The MVP must deliver all four core modules at table-stakes depth to prove the unified platform value. The feature research identified 60+ table stakes features across modules and 20+ differentiators.

**Must have (table stakes for v1.0):**
- Auth: Magic Link + Telegram login, user profiles, system roles (Admin/Member)
- Projects: creation, membership, auto-channel per project, project-scoped RBAC
- Tracker: issue CRUD with rich text, Kanban board with drag-and-drop, basic customizable workflow, issue types/priorities/assignees/labels, parent-child hierarchy, issue linking, comments with @mentions, activity log, file attachments, project-prefixed IDs (HUB-123)
- Messenger: DMs, channels (public+private), threaded replies, reactions, rich text, file sharing, typing indicators, online/offline status, unread counts, message search, pin/bookmark
- TMS: test case CRUD (steps + expected result), folder/suite hierarchy, manual test runs with evidence, pass/fail/blocked/skipped statuses, bidirectional test case to issue linking, basic reporting
- Calls: multi-participant audio, mute/unmute, screen sharing, join via link, participant list
- Cross-module: in-app notifications, global search (Ctrl+K), file storage for all modules, dark/light theme, i18n (RU/EN)

**Should have (add after core validation, v1.x):**
- Sprint management + burndown/velocity charts
- Allure results ingestion from CI pipelines
- Telegram bot notifications (beyond auth)
- GitHub integration (PR linking, auto-status)
- Call recording with project binding
- Custom workflows (visual editor)
- Project-level custom RBAC roles
- BDD test cases (Given/When/Then)
- Planning Poker, scheduled/recurring calls
- Bulk operations, saved/shared filters, link previews in chat

**Defer (v2+):**
- JQL-like query language, traceability matrix view, project templates
- Admin dashboard (DAU/WAU, storage, health)
- CLI tool, PWA offline mode, import from Jira/TestRail
- Custom fields on issues, environments in TMS, swimlanes, WIP limits

**Anti-features (deliberately not building):**
- Video calls (voice + screen share covers 90% of needs)
- Native mobile apps (PWA is sufficient for internal tools)
- Email notifications (Telegram bot replaces this)
- SSO/LDAP/OAuth, Gantt charts, real-time collaborative editing, AI features, plugin marketplace, wikis

### Architecture Approach

The system is a modular monolith: a single Go binary with five domain modules (auth, tracker, messenger, TMS, calls) plus cross-cutting services (notification, storage, search, WebSocket gateway). Modules communicate through NATS events for cross-cutting concerns and share PostgreSQL via Ent ORM for direct entity relationships. A single WebSocket gateway bridges NATS events to browser clients -- modules never touch WebSocket connections directly. The frontend mirrors backend module boundaries with module-specific folders sharing common components.

**Major components:**
1. **REST API (Fiber v3)** -- HTTP request handling with route groups per module, middleware for auth/RBAC/logging
2. **WebSocket Gateway + NATS Bridge** -- single persistent connection per client, subscribes to NATS subjects, fans out events to relevant browsers; enables horizontal scaling without code changes
3. **Domain Modules (auth, tracker, messenger, TMS, calls)** -- self-contained with handler/service/repository/events layers; communicate only through NATS events, never direct imports
4. **PostgreSQL (Ent)** -- source of truth for users, projects, tasks, test cases, roles, file metadata; graph-based entity relations model the cross-module linking
5. **MongoDB** -- chat messages (append-heavy, flexible schema), Allure results JSON (deeply nested, variable structure)
6. **NATS + JetStream** -- core NATS for ephemeral real-time (typing, presence, signaling); JetStream for durable processing (audit, Allure ingestion, notification delivery)
7. **SeaweedFS** -- all file storage via presigned URLs; backend never proxies file bytes

### Critical Pitfalls

1. **Fiber/fasthttp RequestCtx reuse** -- fasthttp reuses request objects; any ctx values passed to goroutines become invalid. Use `utils.CopyString()`, copy body bytes, set `Immutable = true` in dev. Run all tests with `-race` flag. Address in Phase 1.

2. **PostgreSQL + MongoDB cross-database consistency** -- no distributed transactions across PG and Mongo. Use the outbox pattern: write events to a PG outbox table in the same transaction, background worker applies changes to MongoDB. Soft-delete in PG first, async cleanup in Mongo. Address in Phase 1.

3. **RBAC role explosion** -- system roles + project roles + channel permissions create combinatorial complexity. Define a clear hierarchy (System > Project > Resource), use permission bitmaps, precompute effective permissions in Redis, implement via Ent Privacy Layer. Address in Phase 1.

4. **WebSocket split-brain across instances** -- messages sent to server A invisible to clients on server B. Use NATS as the WebSocket message bus from day one; store presence in Redis, not in-memory. Address in Phase 1-2.

5. **NATS JetStream consumer proliferation** -- creating per-user JetStream consumers leads to instability at scale. Use core NATS for real-time fan-out, reserve JetStream only for durable job processing (audit, Allure, webhooks). Address in Phase 1.

6. **LiveKit Egress resource starvation** -- recording consumes 2-6 CPUs and 4GB+ RAM per instance (headless Chrome). Deploy Egress on dedicated node, limit concurrent recordings, add SYS_ADMIN capability. Address in Phase 4 but plan K3s resources in Phase 1.

## Implications for Roadmap

Based on combined research across stack, features, architecture, and pitfalls, the following phase structure is recommended. The ordering is driven by three factors: dependency chains (auth before everything, projects before modules), architectural foundations (NATS/WebSocket/storage before features that need them), and risk frontloading (tackle the hardest integration patterns early).

### Phase 1: Foundation and Infrastructure

**Rationale:** Every module depends on authentication, project scoping, RBAC, real-time infrastructure, and file storage. Building these first eliminates the most dangerous pitfalls (ctx reuse conventions, outbox pattern, RBAC model, NATS patterns) before feature code exists.
**Delivers:** Working auth flow, project CRUD with membership, RBAC middleware, WebSocket gateway with NATS bridge, SeaweedFS integration with presigned URLs, database schemas and migrations, CI pipeline, Docker Compose dev environment.
**Features addressed:** Magic Link + Telegram auth, user profiles, system roles, project creation and membership, file storage infrastructure.
**Pitfalls addressed:** Fiber RequestCtx reuse (establish conventions), PG+MongoDB consistency (outbox pattern), RBAC model (permission hierarchy + middleware), NATS patterns (core vs JetStream separation), WebSocket gateway (NATS bridge from day one).

### Phase 2: Task Tracker

**Rationale:** The tracker is the highest-value module and the foundation for cross-module linking. TMS needs tasks to exist for bidirectional linking. Messenger auto-channels need project context. Building the tracker first validates the Ent graph model, workflow engine, and real-time board updates.
**Delivers:** Issue CRUD with rich text (Plate), Kanban board with drag-and-drop (pragmatic-dnd), customizable workflows, issue types/priorities/assignees/labels, parent-child hierarchy, issue linking, comments with @mentions, activity log, file attachments, project-prefixed IDs.
**Key stack elements:** Ent graph traversal for task hierarchies, pragmatic-drag-and-drop for Kanban, Plate for rich text, fractional indexing (LexoRank) for card positioning, TanStack Query for optimistic updates.
**Pitfalls to watch:** Ent N+1 queries (eager loading from day one), Kanban position recalculation (use LexoRank), real-time board update storms (debounce server-side).

### Phase 3: Messenger

**Rationale:** The messenger is the second daily-use module and validates the real-time infrastructure built in Phase 1. It exercises the WebSocket gateway, NATS pub/sub, MongoDB write patterns, and presence system at full intensity. Building it after the tracker enables cross-module entity references (mentioning HUB-123 in chat).
**Delivers:** DMs, channels (public + private), auto-channels per project, threaded replies, reactions, rich text formatting, file sharing with image preview, typing indicators, online/offline status, unread counts and badges, message search, pin/bookmark.
**Key stack elements:** MongoDB for messages, NATS core pub/sub for real-time delivery, Redis for presence/unread counts, TanStack Virtual for chat virtualization.
**Pitfalls to watch:** WebSocket distributed delivery (already mitigated by NATS bridge), unindexed MongoDB queries (compound index on channel_id + created_at), notification deduplication (batch edits within time window).

### Phase 4: TMS (Test Management)

**Rationale:** The TMS depends on the tracker for bidirectional test-case-to-issue linking, which is the core traceability differentiator. It also depends on the messenger for posting test run results to project channels. Building it third means all integration points exist.
**Delivers:** Test case CRUD (steps + expected result), folder/suite hierarchy, manual test runs with evidence (screenshots/comments per step), pass/fail/blocked/skipped statuses, bidirectional test case to issue linking, basic reporting (pass rate, run history), import (CSV/Excel).
**Key stack elements:** Ent for test case schemas with edges to tasks, MongoDB for Allure JSON storage, NATS events for test run completion notifications.
**Pitfalls to watch:** Allure parser brittleness (defensive parsing, multiple matching strategies), cross-module linking complexity (test coverage calculation).

### Phase 5: Voice Calls

**Rationale:** Calls are the most self-contained module -- LiveKit handles the heavy lifting. It depends on auth and projects but not on tracker or TMS. Placed after TMS because the traceability story (tracker + TMS) is the primary differentiator, while calls are table stakes functionality.
**Delivers:** Multi-participant audio calls, mute/unmute, screen sharing (multiple simultaneous), join via link, participant list, call metadata stored in PostgreSQL.
**Key stack elements:** LiveKit server SDK for room management and token generation, @livekit/components-react for UI, NATS for call start/end events to messenger.
**Pitfalls to watch:** LiveKit + Caddy configuration (Caddy proxies signaling only, LiveKit handles UDP media directly), LiveKit NODE_IP configuration for production.

### Phase 6: Notifications, Search, and Dashboards

**Rationale:** These are cross-cutting features that aggregate data from all modules. Notifications consume events from tracker, messenger, TMS, and calls. Global search indexes entities from all modules. Dashboards query across all data. They must come last because they depend on everything.
**Delivers:** In-app notification center (bell icon, mark read/unread), notification preferences, global search (Ctrl+K) across all entities, personal dashboard, project dashboard with widgets.
**Key stack elements:** NATS consumers for notification dispatch, PostgreSQL full-text search (tsvector + GIN index), Tremor for dashboard charts and KPIs, Zustand for notification state.
**Pitfalls to watch:** Dashboard query performance (use materialized views or Redis-cached aggregations, not live queries), notification deduplication and batching, search index staleness.

### Phase 7: Polish and Extended Features

**Rationale:** After all core modules and cross-cutting features work, add the features that deepen each module: sprint management, call recording, Telegram notifications, GitHub integration, and UX polish (i18n, theme, keyboard shortcuts, PWA).
**Delivers:** Sprint management + burndown/velocity, call recording (LiveKit Egress), Telegram bot notifications, GitHub integration (PR linking, auto-status), dark/light theme with accent colors, i18n (RU/EN), keyboard shortcuts, PWA basics.
**Pitfalls to watch:** LiveKit Egress resource starvation (dedicated node, concurrent recording limits, SYS_ADMIN capability), GitHub webhook task matching (word boundary regex).

### Phase Ordering Rationale

- **Foundation first** because 7 of 9 critical pitfalls must be addressed in Phase 1 -- delaying infrastructure decisions causes cascading technical debt
- **Tracker before Messenger** because task management is the primary differentiator (messenger without tasks is just another Slack clone), and the tracker validates the Ent graph model that TMS also needs
- **Messenger before TMS** because it validates the real-time infrastructure at full intensity and enables cross-module entity references needed by TMS
- **TMS before Calls** because the tracker-to-test-case traceability is the core differentiator against Huly and all competitors; calls are commodity functionality
- **Notifications/Search/Dashboards as a group** because they all share the same dependency: needing data from all modules to be meaningful
- **Polish last** because sprint management, recording, and GitHub integration add depth to working modules rather than creating new ones

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 1 (Foundation):** RBAC permission model design needs detailed specification -- the exact permission set, hierarchy rules, and Ent Privacy Layer integration require careful upfront design
- **Phase 4 (TMS):** Allure results parser needs research into actual JSON format variations across frameworks (JUnit, pytest, JS). Test with real allure-results from multiple CI pipelines before finalizing the parser
- **Phase 5 (Calls):** LiveKit self-hosting configuration, Caddy integration for signaling, and K3s deployment with UDP port exposure need hands-on validation
- **Phase 7 (Polish):** LiveKit Egress deployment on K3s with resource isolation and SYS_ADMIN capability requires K8s-specific research

Phases with standard patterns (skip deep research):
- **Phase 2 (Tracker):** Well-documented patterns -- Kanban boards, workflow engines, and task CRUD are extensively covered in Ent docs and Fiber examples
- **Phase 3 (Messenger):** WebSocket chat with NATS pub/sub is a well-documented pattern with multiple Go reference implementations
- **Phase 6 (Notifications/Search/Dashboards):** Standard patterns -- NATS consumers, PostgreSQL FTS, and Tremor dashboard components have clear documentation

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All technologies verified via official release notes and docs. Versions confirmed current as of March 2026. MinIO deprecation caught and SeaweedFS alternative validated. |
| Features | HIGH | Feature landscape benchmarked against 10+ competitors (Jira, Linear, Plane, Huly, Slack, Mattermost, Zulip, TestRail, Allure TestOps). Clear MVP definition with prioritization matrix. |
| Architecture | HIGH | Modular monolith with NATS is a proven pattern for this scale. Project structure follows Go community conventions. Data flow diagrams cover all cross-module interactions. |
| Pitfalls | MEDIUM-HIGH | Most pitfalls verified through official docs and multiple community sources. Fiber/fasthttp ctx reuse and NATS consumer anti-patterns confirmed by maintainers. LiveKit Egress resource requirements from official docs. Allure parser brittleness is experiential -- less documented. |

**Overall confidence:** HIGH

### Gaps to Address

- **SeaweedFS migration from MinIO references:** PROJECT.md constraints still reference MinIO. All deployment configs and documentation should use SeaweedFS. The S3 API compatibility means no code-level changes, but Docker Compose and K3s manifests need SeaweedFS containers instead of MinIO.
- **Ent Privacy Layer for RBAC:** The exact integration between Ent's Privacy Layer, the RBAC middleware, and Redis-cached effective permissions needs a design spike during Phase 1 planning. The pattern is documented but Hub's multi-scope permission model (system + project + channel) adds complexity not covered in Ent examples.
- **MongoDB message schema design:** The exact document structure for chat messages (threading model, reaction storage, attachment references) needs finalization. The research confirms MongoDB is the right choice but the schema itself needs design work.
- **Allure results format variations:** The parser needs to handle output from at least 3 framework adapters (JUnit/Java, pytest, JavaScript). Real sample files should be collected before Phase 4 implementation.
- **K3s production deployment:** The research covers what to deploy but not the detailed K3s manifest structure (ingress, persistent volumes, secrets management, LiveKit UDP exposure). This needs infrastructure research during Phase 1.
- **Offline/reconnection UX:** The research identifies WebSocket reconnection with gap recovery as a "looks done but isn't" item. The exact client-side reconnection strategy (last message ID cursor, exponential backoff, gap fetch) needs design.

## Sources

### Primary (HIGH confidence)
- Go 1.26 release blog, Fiber v3 docs and release notes, Ent official site
- React 19.2 blog, Vite 8 announcement, Tailwind CSS v4.2 docs
- NATS JetStream documentation, Synadia blog on JetStream anti-patterns
- LiveKit documentation (self-hosting, Egress, components-react)
- PostgreSQL 17 docs, MongoDB Go driver v2 docs
- SeaweedFS documentation, MinIO archival notice (GitHub)

### Secondary (MEDIUM confidence)
- Competitor feature analysis (Huly, Plane, Linear feature pages)
- Ent ORM community criticism and comparison articles
- WebSocket scaling patterns (community articles, reference implementations)
- RBAC implementation pitfalls (security blogs)

### Tertiary (LOW confidence)
- LiveKit Egress resource requirements (community issues, may vary by hardware)
- Allure results format stability across framework adapters (de facto standard, not strict spec)
- SeaweedFS performance benchmarks vs MinIO (third-party comparison sites)

---
*Research completed: 2026-03-20*
*Ready for roadmap: yes*
