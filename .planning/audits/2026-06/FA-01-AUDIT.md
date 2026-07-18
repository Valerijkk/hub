# FA-01 — SOLID / DRY Audit (Post TF Migration)

> **Scope:** Wave C FA-01 · Highest-risk areas after test-framework migration  
> **Date:** 2026-06-22  
> **Method:** Read-only code review (SOLID, DRY, layering, shutdown/lifecycle)  
> **Focus files:** `cmd/server/main.go`, `internal/apigateway/*`, `internal/ws/hub.go`, `internal/auth/service_session.go`  
> **Auditor:** code-reviewer  
> **Branch:** `feat/v2.5-trust`

> **Remediation status (2026-06-22):** FA-01 backlog **fully closed** (P0–P3 including APIGW-05). Wave C code path complete; **FA-06 PG-03** manual ops remains.

---

## Executive Summary

| Metric | Count |
|--------|------:|
| CRITICAL | 0 (APIGW-01 ✅ 2026-06-22) |
| HIGH | 4 |
| MEDIUM | 8 |
| LOW | 5 |
| **Verdict** | **CONDITIONAL PASS** — P0 closed; P1/P2/P3 backlog open |

Post-migration **`main.go` is now a thin orchestrator** (~34 LOC) delegating to `wiring_*.go` — a major SRP win vs the prior ~800 LOC monolith. Remaining risk concentrates in **apigateway authorization gaps**, **webhook dispatcher delivery semantics**, and **session refresh validation**. WS hub and session code show deliberate hardening (SEC-05, ME-02) with manageable residual lifecycle/DRY issues.

---

## Findings by Area

### 1. `cmd/server/main.go` + `wiring_*.go`

#### ✅ LGTM (post-refactor)

- Thin `main()` with explicit phase order: `bootstrapServer → wireInfra → … → runServer` (`main.go:18-33`).
- Centralized lifecycle via `serverDeps.addCleanup` / `runCleanups()` (`wiring_deps.go:119-127`) — improves over scattered `defer` in monolith.
- SEC-01 preserved: `APIKeyAuth` registered in `wireAuth` before any `RegisterRoutes` (`wiring_auth.go:85-89`).
- WS authorizer extracted to `wireWebSocket` with same membership closures as before (`wiring_ws.go:37-65`).

#### HIGH — `serverDeps` service-locator + implicit wire ordering

| | |
|---|---|
| **Severity** | HIGH |
| **SOLID** | DIP / OCP — composition root knows every module field |
| **Files** | `wiring_deps.go:41-117`, `main.go:22-32` |

`serverDeps` holds 30+ service pointers; `wireNotification` depends on `d.trackerRepo` from `wireCore`, `wireWebSocket` depends on `d.callsRepo` from `wireIntegrations` — **order in `main()` is the only contract**. Reordering calls silently breaks startup.

**Fix:** Document wire DAG in `wiring_deps.go` header or split into typed sub-structs (`infraDeps`, `coreDeps`). Long-term: constructor per vertical with explicit inputs.

#### MEDIUM — Duplicate `apigateway` repository instances

| | |
|---|---|
| **Severity** | MEDIUM |
| **DRY** | Same Ent client, three wrapper repos |
| **Files** | `wiring_auth.go:78-79`, `wiring_integrations.go:79-80` |

`wireAuth` creates `apigatewayRepo` for API keys; `wireIntegrations` creates **two more** `NewEntRepository(d.db)` for webhook service and dispatcher. Behavior is equivalent but diverges from monolith pattern (single shared repo) and complicates testing/mocking.

**Fix:** Store `d.apigatewayRepo` once in `wireAuth`; pass to `NewWebhookService` / `NewWebhookDispatcher`.

#### MEDIUM — Notification / push consumers lack `addCleanup(Stop)`

| | |
|---|---|
| **Severity** | MEDIUM |
| **SOLID** | Resource lifecycle (SRP of composition root) |
| **Files** | `wiring_notification.go:20-38` |

`automationConsumer` and `auditConsumer` register `addCleanup(Stop)`; `notifConsumer`, `telegramSender`, `pushSender`, `telegramBind` start goroutines but **never Stop on shutdown**. JetStream / polling may leak in-flight work or hold connections during deploy.

**Fix:** Mirror audit pattern — expose `Stop()` and `d.addCleanup(...)`.

#### MEDIUM — Outbox worker no graceful drain

| | |
|---|---|
| **Severity** | MEDIUM |
| **Files** | `wiring_infra.go:53-56` |

`go outboxWorker.ProcessLoop(d.rootCtx)` cancels on signal but no `addCleanup` wait for in-flight publishes. Minor data-loss window on hard shutdown.

#### LOW — Repeated `IsSystemAdmin` closure (DRY)

| | |
|---|---|
| **Severity** | LOW |
| **Files** | `wiring_audit_ai.go:49-51,62-64`, `wiring_modules.go` (onec admin) |

Three identical `func(c fiber.Ctx, userID uuid.UUID) (bool, error) { return d.auditSvc.IsSystemAdmin(...) }` lambdas.

**Fix:** `(d *serverDeps) isSystemAdmin(c fiber.Ctx, userID uuid.UUID) (bool, error)`.

#### LOW — Duplicate health probe logic

| | |
|---|---|
| **Severity** | LOW |
| **DRY** | `wireMessenger` inline `/health` + `statusmod.NewService` ping funcs |
| **Files** | `wiring_messenger.go:33-88` |

Same postgres/redis/mongo ping logic written twice in one function.

---

### 2. `internal/apigateway`

#### CRITICAL — Webhook CRUD missing project authorization (IDOR)

| | |
|---|---|
| **Severity** | CRITICAL |
| **SOLID** | SRP violated at HTTP boundary — authz belongs in service/handler |
| **OWASP** | A01 Broken Access Control |
| **Files** | `handler.go:23-40,144-265`, `webhook.go:307-373` |

Routes use only `authMw` (JWT/API-key identity). **No project membership or RBAC check:**

- `POST /api/projects/:projectId/webhooks` — any authenticated user can register outbound webhooks on **any** `projectId`.
- `PUT/DELETE/GET /api/webhooks/:id` — operate on endpoint by ID with **no ownership check**.

Attacker can point webhooks at attacker-controlled URLs and receive project event payloads (tracker, messenger, etc.).

**Fix:** Inject `RBACService` or project membership checker; verify `projectId` / endpoint's project before CRUD. Align with `storageHandler.RegisterRoutes(api, requireAuth, rbacSvc)` pattern in `wiring_core.go:53`.

#### HIGH — JetStream message Ack before async delivery completes

| | |
|---|---|
| **Severity** | HIGH |
| **Correctness** | At-least-once broken when queue drops |
| **Files** | `webhook.go:112-116,169-191` |

Consumer callback runs `handleMessage` → `dispatchEvent` (non-blocking queue push) then **`msg.Ack()` immediately**. If `d.jobs` is full, delivery is dropped with a warn log but message is already acked — **silent event loss** (no JetStream redelivery).

**Fix:** Ack only after job accepted (or after successful HTTP delivery); use Nak/Delay on queue full.

#### HIGH — Unfiltered EVENTS consumer + per-message DB scan

| | |
|---|---|
| **Severity** | HIGH |
| **Performance / SRP** | Dispatcher does work for every stream message |
| **Files** | `webhook.go:100-106,169-174`, `repository.go:112-116` |

Durable consumer on full `EVENTS` stream with no `FilterSubject`. Every NATS message triggers `ListActiveWebhookEndpoints` (full table scan) even when zero webhooks care about the subject.

**Fix:** Add `FilterSubjects` to consumer config or push-only webhook-specific subjects; cache endpoint list with TTL.

#### MEDIUM — HTTP/Fiber errors in service layer

| | |
|---|---|
| **Severity** | MEDIUM |
| **SOLID** | Layering — service depends on Fiber |
| **Files** | `apikey.go:138-139,151-157,173-174`, `webhook.go:309,349,356,368` |

`APIKeyService` / `WebhookService` return `fiber.NewError` instead of domain errors mapped in handlers. Violates backend module template (`handler → service → repository`).

**Fix:** Return typed errors (`ErrNotFound`, `ErrUnauthorized`); map to HTTP in `handler.go`.

#### MEDIUM — `APIKeyAuth` skips deactivated-user check

| | |
|---|---|
| **Severity** | MEDIUM |
| **Security / consistency** | JWT path checks status; API-key path does not |
| **Files** | `apikey.go:182-196`, `auth/middleware.go:81-87` |

`RequireAuth` calls `userIsDeactivated`; `VerifyAPIKey` only loads user edge, not `status`. Deactivated users with valid API keys retain access until keys revoked.

**Fix:** Reject deactivated users in `VerifyAPIKey` or shared auth middleware after key verification.

#### LOW — Dead field `stopCh` in `APIKeyService`

| | |
|---|---|
| **Severity** | LOW |
| **Files** | `apikey.go:43,54` |

`stopCh` allocated but never used; workers exit via closed `touchJobs` only.

#### LOW — `hashKey` duplicates `auth.hashRefreshToken` pattern

| | |
|---|---|
| **Severity** | LOW |
| **DRY** | Identical SHA-256 hex helper in two packages |
| **Files** | `apikey.go:199-203`, `auth/service_session.go:59-62` |

Consider `platform/crypto/hash` helper (non-blocking).

#### ✅ LGTM

- Outbound webhooks use `safehttp.NewSafeHTTPClient` + `ValidateOutboundURL` (`webhook.go:28,243-245`).
- Response body capped via `io.LimitReader` (`webhook.go:267-268`, CR-03).
- Bounded worker pools for touch + delivery (BE-H4, HI-08).
- Repository interfaces split `APIKeyRepository` / `WebhookRepository` (`repository.go:16-59`).

---

### 3. `internal/ws/hub.go`

#### MEDIUM — Hub `Run` exit does not drain client registry

| | |
|---|---|
| **Severity** | MEDIUM |
| **Lifecycle** | Shutdown may block or leak |
| **Files** | `hub.go:107-111,134-156` |

On `ctx.Done()`, `Run` returns immediately without unregistering active clients or closing `done` on remaining connections. `Register`/`Subscribe` use **blocking channel sends** — if `Run` has exited, callers block forever.

**Fix:** On shutdown signal, drain channels or use `select` with `ctx` in public methods; optionally iterate `clients` and call `removeClient`.

#### MEDIUM — Race window: `BroadcastToRoom` vs `removeClient`

| | |
|---|---|
| **Severity** | MEDIUM (benign drop) |
| **Files** | `hub.go:198-223,262-302` |

`BroadcastToRoom` reads `rooms` sync.Map outside `Run` goroutine while `removeClient` mutates under `Run`. SEC-05 `done` check mitigates panics; stale members may receive until removed — acceptable but worth documenting.

#### LOW — DRY: duplicate marshal/send loop

| | |
|---|---|
| **Severity** | LOW |
| **Files** | `hub.go:198-223`, `229-254` |

Extract `deliverToClient(client, data, roomLabel)`.

#### ✅ LGTM

- Single goroutine owns all registry mutations (HI-07).
- SEC-05: `done` closed, `send` never closed — eliminates send-on-closed-channel panic.
- Multi-tab fan-out via `connectionID` keys (HI-07).
- Authorizer fail-closed for unknown room prefixes (`authorizer.go:77-79`).

---

### 4. `internal/auth/service_session.go`

#### HIGH — `RefreshToken` trusts Redis only; no DB session / expiry check

| | |
|---|---|
| **Severity** | HIGH |
| **Correctness / security** | Revoked/expired DB session may still refresh |
| **Files** | `service_session.go:204-232` |

Flow: read Redis payload → rotate refresh → update hash in DB. **Never verifies** session row exists, `expires_at` not passed, or DB hash matches. Scenarios:

- DB row deleted (manual/admin) while Redis key alive → refresh succeeds until Redis TTL.
- `UpdateSessionRefreshHash` fails after new Redis key stored → orphaned valid refresh in Redis.

**Fix:** `GetSession(sessionID)` before rotate; compare stored hash to `hashRefreshToken(old)`; check `expires_at`.

#### MEDIUM — Concurrent refresh race (token replay)

| | |
|---|---|
| **Severity** | MEDIUM |
| **Files** | `service_session.go:218-230` |

Two parallel requests with same old refresh can both pass `getRefreshPayload` before either deletes old key → two valid refresh tokens (ME-02 ordering helps partial failure, not concurrency).

**Fix:** Redis `GETDEL` or transactional Lua script for atomic rotate.

#### MEDIUM — JWT access token valid after session revoke (inherent)

| | |
|---|---|
| **Severity** | MEDIUM (accepted tradeoff) |
| **Files** | `service_session.go:79-94`, `middleware.go:76-78` |

`ParseAccessToken` / WS auth do not check session still in DB. Revoked session leaves access token valid up to 15 min TTL. Document as known limitation or add session-version claim.

#### LOW — `SessionService` embeds `*redis.Client` directly

| | |
|---|---|
| **Severity** | LOW |
| **SOLID** | Could use `RefreshTokenStore` interface |
| **Files** | `service_session.go:23-26` |

Consistent with other Hub modules using Redis directly; optional refactor.

#### ✅ LGTM

- Refresh tokens stored hashed in Redis keys (`refreshRedisKey`).
- Single-session policy with bulk DB delete + Redis pipeline (`invalidateUserSessions`, ME-01).
- ME-02 rotation order documented and tested (`testframework/auth/service_session_test.go`).
- Error wrapping with context (`fmt.Errorf("...: %w", err)`).

---

## Cross-Cutting DRY Matrix

| Pattern | Locations | Severity |
|---------|-----------|----------|
| SHA-256 hex hash helper | `apigateway/apikey.go`, `auth/service_session.go` | LOW |
| `IsSystemAdmin` lambda | `wiring_audit_ai.go`, `wiring_modules.go` | LOW |
| Health ping funcs | `wiring_messenger.go` `/health` + status service | LOW |
| `apigateway` EntRepository | `wiring_auth.go`, `wiring_integrations.go` ×2 | MEDIUM |
| Admin auth closure | 3× identical in wiring | LOW |

---

## Priority Remediation (Wave C backlog)

| Priority | ID | Action | Status |
|----------|-----|--------|--------|
| P0 | APIGW-01 | Add RBAC/project membership to all webhook CRUD handlers | ✅ DONE 2026-06-22 |
| P1 | APIGW-02 | Fix webhook Ack semantics (ack after enqueue success or delivery) | ✅ DONE 2026-06-22 |
| P1 | AUTH-01 | DB + expiry validation in `RefreshToken` | ✅ DONE 2026-06-22 |
| P1 | APIGW-03 | Filter JetStream consumer or cache endpoint list | ✅ DONE 2026-06-22 |
| P2 | BOOT-01 | `addCleanup(Stop)` for notification consumers | ✅ DONE 2026-06-22 |
| P2 | APIGW-04 | Deactivated-user check in API key path | ✅ DONE 2026-06-22 |
| P2 | BOOT-02 | Single shared `apigatewayRepo` on `serverDeps` | ✅ DONE 2026-06-22 |
| P3 | WS-01 | Hub graceful shutdown / non-blocking register | ✅ DONE 2026-06-22 |
| P3 | APIGW-05 | Remove Fiber types from apigateway services | ✅ DONE 2026-06-22 |

---

## Verdict

**PASS (production-ready code path)** — P0/P1/P2 closed 2026-06-22. Remaining **P3** WS-01, APIGW-05 — non-blocking backlog.

**Positive:** TF migration preserved security fixes (SEC-01, SEC-05, SSRF-safe webhooks); `main.go` decomposition is the right direction and should be completed (shared repos, consumer cleanup, admin helper) rather than reverted.

---

*Generated: 2026-06-22 · FA-01 Wave C · Read-only audit · No git commits*
