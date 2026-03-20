# Pitfalls Research

**Domain:** Team collaboration hub (task tracker + messenger + TMS + voice calls)
**Researched:** 2026-03-20
**Confidence:** MEDIUM-HIGH (most pitfalls verified through official docs and multiple sources)

## Critical Pitfalls

### Pitfall 1: Fiber/fasthttp RequestCtx Reuse Causes Data Races and Panics

**What goes wrong:**
Fiber is built on fasthttp, which reuses RequestCtx objects across requests for performance. Any values obtained from `fiber.Ctx` (query params, body, headers) are invalidated after the handler returns. Storing references to ctx values in goroutines or passing ctx to background workers causes data races, corrupted data, or panics. This is the single most dangerous Fiber-specific pitfall -- it violates Go developers' expectations from net/http where request objects are safe to retain.

**Why it happens:**
fasthttp achieves zero-allocation performance by pooling and reusing request/response objects. Developers coming from net/http or Express.js assume context values are stable after handler return. The bug is silent until load increases and the pool starts recycling contexts rapidly.

**How to avoid:**
- Copy all values from ctx before passing to goroutines: `body := make([]byte, len(c.Body())); copy(body, c.Body())` or use `c.Body()` with `copy()`.
- Use `c.Locals()` only within the handler scope.
- For WebSocket upgrade handlers, extract all needed data before the upgrade.
- Use `utils.CopyString(c.Params("id"))` instead of raw `c.Params("id")`.
- Set `app.Config().Immutable = true` during development to catch violations (disables zero-copy, making all values safe but slower).

**Warning signs:**
- Intermittent wrong data in responses under load.
- Panics with nil pointer dereferences in background goroutines referencing ctx.
- Data race warnings in `go test -race`.

**Phase to address:**
Phase 1 (Foundation) -- establish Fiber coding conventions and middleware patterns before any feature code is written. Add `-race` flag to all test runs from day one.

---

### Pitfall 2: PostgreSQL + MongoDB Cross-Database Consistency Gaps

**What goes wrong:**
With tasks/users in PostgreSQL and chat messages/Allure reports in MongoDB, operations that span both databases lack transactional guarantees. Example: creating a project creates a PG record + a MongoDB channel. If MongoDB write fails, you have an orphaned project with no chat channel. Worse, deleting a user requires cleanup in both databases -- partial failures leave ghost data.

**Why it happens:**
Distributed transactions (2PC) across PG and MongoDB are impractical and not supported by Ent or standard Go drivers. Teams assume "it rarely fails" until it does, leaving data in inconsistent states that are hard to detect and harder to fix.

**How to avoid:**
- Designate PostgreSQL as the source of truth. MongoDB is secondary/derived.
- Use the Outbox Pattern: write events to a PG outbox table in the same transaction as the primary write, then a background worker reads the outbox and applies changes to MongoDB. Failed MongoDB writes are retried from the outbox.
- For deletions: soft-delete in PG first (single transaction), then async cleanup in MongoDB. Never hard-delete across both in one operation.
- Build a reconciliation job that periodically checks PG entities have corresponding MongoDB documents and flags orphans.
- Accept eventual consistency for chat/Allure data -- it is tolerable for these use cases.

**Warning signs:**
- Manual "fix the data" requests from users (orphaned channels, missing chat history).
- Inconsistent counts between PG project list and MongoDB channels.
- Failed background jobs silently dropping events.

**Phase to address:**
Phase 1 (Foundation) -- implement the outbox pattern in the data layer before building any cross-database features. This is architectural, not something to bolt on later.

---

### Pitfall 3: Fiber's fasthttp Does Not Support HTTP/2 or HTTP/3

**What goes wrong:**
fasthttp only supports HTTP/1.1. This means Fiber cannot natively serve HTTP/2 connections. Since Caddy is the reverse proxy, this is manageable (Caddy handles HTTP/2 to clients, proxies HTTP/1.1 to Fiber). But if anyone tries to use gRPC (which requires HTTP/2) between services, or removes the proxy layer, the application breaks silently -- connections downgrade to HTTP/1.1 without error.

**Why it happens:**
Developers don't realize fasthttp's limitation until they try to add gRPC or server push. The limitation is fundamental to fasthttp's design and will not be fixed.

**How to avoid:**
- Accept HTTP/1.1 for Fiber services; always front with Caddy for HTTP/2 client-facing.
- If inter-service communication needs gRPC later, run a separate net/http-based gRPC server alongside Fiber (different port).
- Use NATS for inter-service communication instead of gRPC -- avoids the HTTP/2 issue entirely and aligns with the chosen stack.
- Document this constraint in the architecture decision record.

**Warning signs:**
- Discussion of adding gRPC endpoints to Fiber services.
- Performance complaints from mobile clients if Caddy proxy is bypassed.

**Phase to address:**
Phase 1 (Foundation) -- document as an architecture constraint. Not a bug to fix, but a constraint to design around.

---

### Pitfall 4: Ent ORM Code Generation Bloat and Complex Queries

**What goes wrong:**
Ent generates large amounts of code per schema -- every entity produces multiple files with hooks, interceptors, predicates, and mutation types. With the Hub's schema complexity (tasks with deep hierarchies, custom fields, workflow states, TMS entities, RBAC), the generated code volume becomes significant: longer compile times, harder debugging, and complex SQL queries even for simple operations. Developers have reported Ent generating unnecessarily complex SELECT statements for straightforward lookups.

**Why it happens:**
Ent's design philosophy is "generate everything you might need." The graph-based query builder creates join chains that may not map efficiently to the intended SQL. Features like Privacy Layer and custom hooks add generated methods even if unused.

**How to avoid:**
- Use Ent's `entc` code generation features to disable unused features (e.g., skip GraphQL integration, disable unused hooks).
- For performance-critical queries (dashboard aggregations, search, reporting), bypass Ent and use raw SQL via `ent.Driver().Query()` or a lightweight query builder like `squirrel`.
- Set up index definitions in schemas from the start -- missing indexes on foreign keys have been reported to cause 50x slowdowns.
- Profile generated SQL in development with query logging enabled.
- Consider Ent for CRUD-heavy modules (tasks, TMS) but raw SQL for analytics/dashboards.

**Warning signs:**
- Compile times growing noticeably as schemas expand.
- Slow queries in development that "shouldn't be slow."
- Developers avoiding Ent's query builder and writing raw SQL informally (sign the abstraction is leaking).

**Phase to address:**
Phase 1 (Foundation) -- establish the pattern of "Ent for CRUD, raw SQL for analytics" early. Add query logging middleware from day one. Define all indexes in initial schemas.

---

### Pitfall 5: WebSocket Connection Management Without Distributed Pub/Sub

**What goes wrong:**
The messenger module needs real-time message delivery to all connected clients in a channel. On a single server instance this is straightforward (in-memory map of connections). But the moment you add a second instance (K3s with replicas), messages sent to server A are invisible to clients connected to server B. This is the classic "split-brain WebSocket" problem that breaks chat applications.

**Why it happens:**
Developers build the single-instance version first, it works perfectly in development, and the distributed case is deferred as "we'll add it later." But retrofitting distributed pub/sub into an existing WebSocket layer requires restructuring connection management, message routing, and presence tracking.

**How to avoid:**
- Use NATS as the WebSocket message bus from the start. Every WebSocket server subscribes to NATS subjects for channels it has active connections for. Messages are published to NATS, which fans out to all server instances.
- Store online presence in Redis (sorted sets with TTL) rather than in-memory maps, so any server instance can query who is online.
- Use sticky sessions (Caddy `lb_policy cookie`) for WebSocket connections to avoid reconnection churn, but never depend on stickiness for correctness.
- Design the WebSocket layer as a thin transport: receive message -> publish to NATS -> NATS delivers to all subscribers -> each server pushes to its local connections.

**Warning signs:**
- Messages work in dev (single instance) but "disappear" in staging (multiple instances).
- Presence shows users as offline when they are connected to a different instance.
- "It works if I scale down to 1 replica."

**Phase to address:**
Phase 2 (Messenger) -- but the NATS integration pattern must be established in Phase 1. The WebSocket + NATS architecture is foundational, not a messenger-specific concern.

---

### Pitfall 6: RBAC Role Explosion in Multi-Context Permission System

**What goes wrong:**
Hub has system roles (Admin, Member), project roles (Owner, Developer, Tester, Viewer), custom roles per project, and channel-level permissions in the messenger. This creates a combinatorial explosion: a user can be Admin globally, Viewer in Project A, Developer in Project B, and channel admin in #project-b-backend. Permission checks become inconsistent, slow (multiple DB queries per request), and impossible to reason about.

**Why it happens:**
RBAC is designed incrementally: "just add another role check here." Over time, the permission logic scatters across handlers, middleware, and Ent privacy rules. Custom roles with granular permissions (50+ possible permissions) create a matrix that nobody can mentally model.

**How to avoid:**
- Define a clear permission hierarchy: System > Project > Resource. Higher scope never grants less access than lower scope (principle of hierarchical inheritance).
- Use a permission bitmap or enum set per role, not string-based checks. Precompute effective permissions when role assignments change (cache in Redis), not on every request.
- Implement Ent's Privacy Layer for entity-level access control -- it integrates with the query builder and prevents accidental data leaks at the ORM level.
- Limit custom role permissions to a predefined set (30-40 max). Don't allow arbitrary permission creation.
- Build an admin UI for roles early -- if you can't visualize it, you can't manage it.
- Write comprehensive permission tests: "User with role X in project Y accessing resource Z should get [allow/deny]."

**Warning signs:**
- Permission bugs where users see data they shouldn't, or can't access data they should.
- Multiple `if user.HasPermission(...)` checks scattered through handler code instead of middleware.
- Developers hardcoding role names in business logic (`if role == "admin"`).

**Phase to address:**
Phase 1 (Foundation) -- RBAC is foundational infrastructure. Define the permission model, implement the middleware, and write the test matrix before building features that depend on it.

---

### Pitfall 7: NATS JetStream Consumer Proliferation

**What goes wrong:**
Creating a JetStream consumer per user or per channel for real-time delivery leads to tens of thousands of consumers. Beyond ~100k consumers, NATS server becomes unstable -- Raft traffic increases, meta-leader overloads, and consumer state calculations slow down. For a 50-person team this is unlikely, but patterns like "one durable consumer per notification preference per user per project" scale multiplicatively.

**Why it happens:**
JetStream consumers are easy to create and feel like the "right" abstraction for per-user message delivery. The Synadia team has explicitly called this out as an anti-pattern because consumer info operations hit the meta-leader and are expensive at scale.

**How to avoid:**
- Use core NATS pub/sub for real-time fan-out (notifications, chat messages, presence). These are ephemeral -- no state, no consumers, no Raft.
- Reserve JetStream for durable message processing: background jobs, Allure result ingestion, audit log processing, webhook delivery. These need at-least-once guarantees and persistence.
- Use JetStream's `republish` feature to route messages to subjects without creating consumers.
- For notification delivery: publish to core NATS, let connected WebSocket servers deliver immediately. For offline users, write to a PostgreSQL notification table. Don't use JetStream consumers as mailboxes.

**Warning signs:**
- JetStream consumer count growing proportionally to user count.
- NATS server memory usage climbing steadily.
- Slow consumer info responses.

**Phase to address:**
Phase 1 (Foundation) -- establish the pattern: "core NATS for real-time, JetStream for durable jobs." This architectural decision affects every module.

---

### Pitfall 8: LiveKit Egress Recording Resource Starvation

**What goes wrong:**
LiveKit Egress (call recording + screen share capture) is extremely resource-intensive. A single RoomComposite recording consumes 2-6 CPUs and needs 4GB+ RAM. On a budget K3s cluster, starting a recording can starve other services of resources. Additionally, since v1.7.6, Chrome sandboxing requires `--cap-add=SYS_ADMIN` in Docker -- without it, all recordings silently fail with "chrome failed to start."

**Why it happens:**
Egress works by running a headless Chrome browser that joins the room and records the composite output. This is inherently heavy. Developers test with a single recording in development and don't discover the resource impact until production. The SYS_ADMIN capability requirement is a security concern that K8s administrators may reject.

**How to avoid:**
- Deploy Egress as a separate service on a dedicated node/VM, not co-located with the main LiveKit server or application services.
- Set Kubernetes resource limits and requests explicitly for Egress pods (4 CPU, 4GB RAM minimum per recording instance).
- Limit concurrent recordings (1-2 max for a small cluster). Queue additional recording requests.
- Add `SYS_ADMIN` capability to the Egress pod's security context in K3s manifests from the start.
- Test recording under realistic conditions (50 participants, screen share) in staging before going live.
- Monitor for "Egress not stopping" bug -- RoomComposite recordings may continue after the room ends, consuming resources indefinitely.

**Warning signs:**
- Other services OOMKilled when a recording starts.
- Recordings fail silently (no error in UI, but no file produced).
- Recording files are empty or truncated.

**Phase to address:**
Phase 4 (Calls) -- but resource planning for the K3s cluster should account for Egress requirements from Phase 1 infrastructure setup.

---

### Pitfall 9: Allure Results Parser Brittleness

**What goes wrong:**
Building a custom parser for `allure-results` JSON files seems simple -- the format is documented. But the format varies subtly across Allure integrations (JUnit vs pytest vs JS frameworks), field presence is inconsistent (some generators omit `statusDetails`, others use different timestamp formats), and the `{uuid}-result.json` naming makes matching to test cases fragile if `@TestCase` annotations are inconsistently applied by test authors.

**Why it happens:**
The Allure results format is a de facto standard, not a strict specification. Each language framework's Allure adapter has quirks. Test authors forget annotations, CI pipelines produce partial results on timeout, and schema changes between Allure versions break parsers.

**How to avoid:**
- Parse defensively: every field is optional, every value might be wrong. Use Go's `json:",omitempty"` and null-safe accessors.
- Match test cases by multiple strategies in priority order: explicit `@TestCase("HUB-TC-42")` annotation > `fullName` field > test class + method name combination. Never rely on a single matching strategy.
- Validate incoming results before processing: reject files that don't parse, log warnings for missing fields, never panic on malformed input.
- Version the parser API so CI pipeline changes don't break the ingestion endpoint.
- Store raw Allure JSON in MongoDB as-is (document store advantage), then extract structured data into PG for the TMS views.

**Warning signs:**
- Test results appearing as "unknown" or unmapped in the TMS.
- Different CI jobs producing results that parse differently.
- Silent data loss -- results ingested but not linked to test cases.

**Phase to address:**
Phase 3 (TMS) -- but define the Allure ingestion API contract early so CI pipeline setup can begin before the full TMS is built.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Skip outbox pattern, write to PG+Mongo directly | Faster initial development | Orphaned data, manual cleanup, user-visible bugs | Never -- cross-DB consistency is non-negotiable |
| Use in-memory WebSocket hub (no NATS) | Simpler first implementation | Complete rewrite needed for multi-instance | Only if single-instance is the permanent plan |
| Hardcode RBAC checks in handlers | Quick permission logic | Inconsistent access control, security holes, impossible to audit | Never for a multi-tenant project tool |
| Use Ent auto-migration in production | No manual migration files | Destructive migrations, data loss, no rollback path | Only in early dev -- switch to Atlas versioned migrations before any real data exists |
| Skip presigned URLs for MinIO, serve files through app | Simpler file serving | App server becomes bottleneck, large files block goroutines | Never -- presigned URLs are trivial to implement with MinIO |
| Store sessions in memory instead of Redis | Faster initial auth | Sessions lost on restart, can't scale horizontally | Only for local development |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| LiveKit + Caddy | Using regular HTTP proxy for WebRTC. WebRTC needs UDP, not HTTP proxy | Caddy proxies only the signaling (WebSocket). LiveKit handles UDP media directly. Configure LiveKit with `LIVEKIT_NODE_IP` set to the public IP |
| MinIO + Presigned URLs | Generating presigned URLs with internal Docker hostname (e.g., `minio:9000`) that clients can't reach | Configure MinIO with a public-facing URL for presigned URL generation, separate from the internal service URL |
| NATS + Fiber WebSocket | Publishing NATS messages inside WebSocket read loop without backpressure | Use buffered channels between the WebSocket read loop and NATS publisher. Drop or queue messages if NATS is slow |
| Ent + Atlas Migrations | Running `ent.Schema.Create()` auto-migration in production alongside Atlas versioned migrations | Choose one: Atlas versioned migrations for production, auto-migrate only in tests. Never mix both |
| Telegram Bot + Magic Link | Sending magic link tokens in plain Telegram messages that persist in chat history | Use inline keyboards with callback queries for auth flow. Delete the message containing the token after use. Set short TTL (5 min) on tokens in Redis |
| GitHub Webhooks + Task Matching | Regex matching `HUB-123` in commit messages without word boundaries, matching `THUB-123` or `HUB-1234` | Use `\bHUB-\d+\b` word boundary regex. Validate matched task ID exists before creating link |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| N+1 queries in Ent edge traversal | Task list page takes 2+ seconds, DB shows hundreds of queries | Always use `.WithAssignee()`, `.WithLabels()` eager loading. Profile with query logging | 50+ tasks with edges (very early) |
| Unindexed MongoDB message queries | Chat history load slows as messages accumulate | Compound index on `(channel_id, created_at)`. Use cursor-based pagination, not offset | 10k+ messages per channel |
| Full-text search on PostgreSQL without proper indexes | Global search (Ctrl+K) becomes unusable | Use PG `tsvector` + GIN index for task/TMS search. Consider separate search index (MeiliSearch) if PG FTS is insufficient | 5k+ tasks |
| WebSocket broadcast storms | UI freezes when many users edit tasks simultaneously | Debounce real-time updates on the server (100ms window). Send diffs, not full objects. Use NATS subject filtering so users only receive updates for their subscribed projects | 20+ concurrent editors |
| Dashboard widget queries hitting production DB | Dashboard load causes query latency spikes for all users | Use read replicas or materialized views for dashboard aggregations. Cache widget data in Redis with 30s TTL | 10+ widgets querying simultaneously |
| Kanban drag-and-drop position recalculation | Moving a card recalculates all positions in the column | Use fractional indexing (e.g., LexoRank algorithm, same as Jira). Position = string between neighbors, no reindexing needed | 100+ cards per column |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Magic link tokens with long TTL or no rate limit | Token interception allows unauthorized access; brute-force possible | 5-minute TTL, single-use, rate limit to 3 requests per email per 10 minutes. Store in Redis with auto-expiry |
| WebSocket connections without auth token validation | Unauthenticated users can subscribe to channels and read messages | Validate JWT on WebSocket upgrade. Re-validate periodically (token expiry). Disconnect on invalid token |
| File upload without type/size validation at API level | Malicious file upload, storage exhaustion, stored XSS via SVG/HTML files | Validate Content-Type, enforce size limits per file type, strip metadata, serve uploads from separate domain/MinIO directly with `Content-Disposition: attachment` |
| RBAC checks only on UI, not API | Direct API calls bypass permission checks | Implement RBAC as Fiber middleware + Ent Privacy Layer. Never trust client-side permission checks |
| Audit log gaps in real-time operations | Actions via WebSocket skip audit logging | Route all state-changing WebSocket messages through the same audit pipeline as REST endpoints |
| Allure results endpoint without authentication | Anyone can push fake test results | Require API key authentication on the ingestion endpoint. Validate API key belongs to an authorized CI service account |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Real-time updates replace content user is editing | User loses typed comment or task description when someone else edits | Show "X updated this task" banner instead of auto-replacing. For rich text (Plate editor), implement OT/CRDT or lock-and-warn |
| Notification overload from auto-channels | Users muted by volume, miss important notifications | Smart defaults: only notify on @mentions and direct assignments. Batch low-priority notifications. Never notify on own actions |
| Chat message search returns results without context | Users can't tell which conversation a search result belongs to | Show channel name, date, and 1-2 surrounding messages in search results. Click navigates to the message in context |
| Kanban board shows all tasks regardless of filter state | Board becomes unusable with 200+ tasks visible | Default to "My tasks" or "Current sprint" filter. Remember last filter per user per board |
| Call recording notification is easy to miss | Users don't realize they are being recorded (legal/trust issue) | Persistent visual indicator (red dot) during recording. Audio announcement when recording starts. Cannot be dismissed |

## "Looks Done But Isn't" Checklist

- [ ] **WebSocket messenger:** Often missing reconnection with message gap recovery -- verify that missed messages during disconnect are fetched on reconnect (use last message ID cursor)
- [ ] **Task notifications:** Often missing deduplication -- verify that editing a task 5 times in 1 minute sends 1 notification, not 5
- [ ] **File uploads:** Often missing cleanup of orphaned files -- verify that files uploaded but never attached to a message/task are garbage-collected
- [ ] **Kanban drag-and-drop:** Often missing optimistic update rollback -- verify that if the server rejects a position change, the card snaps back to its original position
- [ ] **Sprint auto-carry-over:** Often missing edge cases -- verify behavior when source sprint has no incomplete tasks, or target sprint doesn't exist yet
- [ ] **RBAC custom roles:** Often missing permission inheritance -- verify that removing a permission from a parent role propagates to derived roles
- [ ] **Allure integration:** Often missing partial result handling -- verify behavior when CI uploads 50 of 100 result files (timeout/crash mid-pipeline)
- [ ] **Call recording:** Often missing recording stop on empty room -- verify that Egress stops when all participants leave (known LiveKit bug)
- [ ] **Magic link auth:** Often missing concurrent session limits -- verify that the 5th login doesn't silently invalidate the 1st session without notification
- [ ] **Audit log:** Often missing bulk operation logging -- verify that "move all tasks to Done" logs each task change, not just "bulk update"

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Cross-DB inconsistency (PG/Mongo orphans) | MEDIUM | Build reconciliation script comparing PG entities to MongoDB documents. Fix orphans automatically where safe (create missing Mongo docs). Manual review for ambiguous cases. Implement outbox pattern to prevent recurrence |
| RBAC permission misconfiguration | HIGH | Audit all role assignments via admin API. Compare effective permissions against intended permissions. Fix roles, invalidate Redis permission cache. Write regression tests for the specific misconfiguration |
| Lost WebSocket messages during deploy | LOW | Client-side reconnection with cursor-based gap fetch fills missing messages. Ensure NATS subjects retain messages briefly (WorkQueue policy with short TTL) for replay |
| Ent auto-migration data loss | HIGH | Restore from backup. Switch to Atlas versioned migrations immediately. Never use auto-migration on databases with real data |
| LiveKit Egress stuck recording | LOW | Monitor Egress status via LiveKit API. Auto-terminate recordings older than max expected duration (2 hours). Alert on stuck recordings |
| NATS consumer leak | MEDIUM | List all consumers via NATS CLI. Delete unused/orphaned consumers. Review code for consumer creation without corresponding cleanup. Add consumer count to monitoring dashboard |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Fiber RequestCtx reuse | Phase 1 (Foundation) | `-race` flag in all tests; code review checklist item for ctx value copying |
| PG+MongoDB consistency | Phase 1 (Foundation) | Outbox table exists; integration test that fails MongoDB write and verifies PG rollback |
| HTTP/2 limitation | Phase 1 (Foundation) | Architecture doc explicitly states constraint; Caddy config handles HTTP/2 termination |
| Ent code bloat / slow queries | Phase 1 (Foundation) | Query logging enabled; benchmark suite for critical queries; raw SQL used for dashboards |
| WebSocket distributed delivery | Phase 1-2 (Foundation/Messenger) | Multi-instance test: send message via instance A, verify receipt on instance B |
| RBAC role explosion | Phase 1 (Foundation) | Permission model documented; middleware-based checks; permission test matrix with 20+ scenarios |
| NATS consumer proliferation | Phase 1 (Foundation) | Consumer count metric in monitoring; documented rule: "core NATS for real-time, JetStream for jobs" |
| LiveKit Egress resources | Phase 4 (Calls) | K3s resource limits configured; recording tested with 10+ participants; Egress on dedicated node |
| Allure parser brittleness | Phase 3 (TMS) | Parser tested against real results from 3+ frameworks; graceful handling of malformed input |

## Sources

- [Fiber/fasthttp zero-copy memory model](https://github.com/valyala/fasthttp) -- RequestCtx reuse documentation
- [Fiber WebSocket middleware limitations](https://docs.gofiber.io/contrib/websocket/) -- recover middleware incompatibility
- [fasthttp HTTP/2 limitations](https://github.com/valyala/fasthttp) -- confirmed no HTTP/2 support
- [Ent ORM criticisms](https://dev.to/shandoncodes/stop-using-entgoplease-5gm5) -- code bloat and complex queries
- [Ent ORM comparison 2026](https://encore.cloud/resources/go-orms) -- Ent vs alternatives
- [JetStream Anti-Patterns (Synadia)](https://www.synadia.com/blog/jetstream-design-patterns-for-scale) -- consumer proliferation, info call overhead
- [LiveKit self-hosting deployment](https://docs.livekit.io/transport/self-hosting/deployment/) -- UDP/SSL requirements
- [LiveKit Egress service](https://docs.livekit.io/transport/self-hosting/egress/) -- resource requirements, SYS_ADMIN capability
- [LiveKit Egress Docker issues](https://github.com/livekit/egress/issues/637) -- recording failures
- [RBAC implementation pitfalls](https://idenhaus.com/rbac-implementation-pitfalls/) -- role explosion, over-permissioning
- [WebSocket scaling challenges](https://www.videosdk.live/developer-hub/websocket/websocket-scale) -- distributed state, connection management
- [Allure test result file format](https://allurereport.org/docs/how-it-works-test-result-file/) -- JSON structure documentation

---
*Pitfalls research for: Team collaboration hub (Go+Fiber, PG+MongoDB, NATS, LiveKit, Ent)*
*Researched: 2026-03-20*
