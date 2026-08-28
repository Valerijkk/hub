# Hub — Архитектура

> Единая база знаний. Обновляется ARCH при изменении архитектуры.
> Детали: `docs/TECHNICAL.md`, `docs/wiki/ARCHITECTURE.md`
> Аудит кода: `.planning/audits/code-review-final/ARCHITECTURE.md` (2026-05-15)

---

## Архитектурный стиль

**Модульный монолит** на Go с React SPA. Clean Architecture внутри каждого модуля: Handler → Service → Repository → DB.

| Принцип | Реализация в Hub |
|---------|------------------|
| Clean Architecture | Слои handler/service/repository, DI через интерфейсы (service→repo) |
| Event-Driven | NATS JetStream + Outbox pattern для domain events |
| CQRS (частичный) | Read models для search (PG tsvector + Mongo text index) |
| Repository Pattern | `repository.go` interface + `ent_repository.go` — rollout продолжается (v2.6+) |

**Architecture score (2026-06-21): ~85 / 100** — модульный монолит с v3/v4 domain packages; остаточный долг — layering bypass в tracker/rbac и неполный Repository rollout.

## C4 Level 2 — Диаграмма контейнеров

```
┌─────────────┐     HTTPS/WSS      ┌─────────────┐
│   Browser   │ ◄────────────────► │    Caddy    │
│  React SPA  │                    │   :80/443   │
└─────────────┘                    └──────┬──────┘
                                          │
                    ┌─────────────────────┼─────────────────────┐
                    │                     │                     │
              /api/* │               /ws  │              static │
                    ▼                     ▼                     ▼
            ┌───────────────┐     ┌───────────────┐     ┌───────────────┐
            │ Go Backend    │     │ WS Handler    │     │ Frontend      │
            │ Fiber v3      │     │ (Fiber)       │     │ Vite build    │
            │ :3000         │     │ :3000         │     │               │
            └───────┬───────┘     └───────┬───────┘     └───────────────┘
                    │                     │
    ┌───────────────┼─────────────────────┼───────────────┐
    │               │                     │               │
    ▼               ▼                     ▼               ▼
┌────────┐   ┌──────────┐   ┌────────┐   ┌────────┐   ┌──────────┐
│ PG 17  │   │ MongoDB 8│   │ Redis 7│   │ NATS   │   │ SeaweedFS│
│ Ent    │   │ messages │   │ session│   │ JS     │   │ S3 3.97  │
└────────┘   └──────────┘   └────────┘   └────────┘   └──────────┘
                                              │
                                        ┌─────┴─────┐
                                        │  LiveKit  │
                                        │  v1.8     │
                                        └───────────┘
```

## C4 Level 3 — Модули backend

**31 domain-пакет** + **`platform/`** (инфраструктура) + `testutil/`. Точка композиции: thin `backend/cmd/server/main.go` (~34 LOC) + `wiring_*.go` (`wiring_fiber.go`, `wiring_modules.go`, `wiring_auth.go`, …).

```
backend/internal/
├── accounting/     Ledger, journal entries, invoices (v4)
├── ai/             Multi-provider AI, messenger bot, JQL assist (v3)
├── apigateway/     API Key auth, outbound webhook delivery
├── audit/          Audit log, trash, GDPR, retention cron, alerts
├── auth/           Magic Link, JWT, sessions, org/founder
├── backup/         PG + Mongo + S3 ZIP backups
├── bases/          No-code DB: HubBase/Table/Record (v4)
├── calendar/       Unified calendar events (v4)
├── calls/          LiveKit rooms, recordings, scheduled
├── crm/            Contacts, deals lite (v4)
├── dashboard/      Personal/project/admin dashboards
├── dev/            DEV_MODE seed, dev-login (!release)
├── github/         Webhooks, PR linking, AI review
├── gitlab/         Webhooks, MR linking, pipelines
├── hr/             Leave requests (v4)
├── imex/           Import/export CSV/JSON/PDF/Jira
├── integrations/   onec/ — 1C connector (v4.5)
├── intranet/       News, directory, kudos, random coffee (v4)
├── mail/           hub-mail admin settings
├── marketplace/    Internal marketplace listings (v4)
├── messenger/      Channels, messages, threads, presence
├── notification/   In-app, Telegram bind, Web Push
├── project/        Projects, members, invites, teams, templates
├── rbac/           Custom roles, permission bits
├── search/         Global search, Meilisearch indexer, JQL
├── sprint/         Sprints, burndown, planning poker
├── status/         Public status page health
├── storage/        SeaweedFS presigned URLs, orphan cleanup
├── tms/            Test cases, runs, Allure parser
├── tracker/        Issues, workflow, boards, automation
├── ws/             WebSocket Hub, NATS Bridge, room Authorizer
└── platform/       config, cron, crypto, database, errors, http,
                    middleware, nats, observability, outbox, redis, telegram
```

**Отдельные процессы (cmd/):** `server` (main), `hub-mail`, `meilisearch-indexer`.

**Repository pattern (основные модули с interface):** audit, bases, calls, crm, github, gitlab, hr, intranet, marketplace, messenger, notification, project, search, sprint, tms, tracker и др.

**Ent schemas:** **82** сущности в `backend/ent/schema/` (+ `helpers.go`).

## C4 Level 3 — Фронтенд

```
frontend/src/
├── routes/           TanStack Router file-based (**77** `.tsx` routes, measured 2026-07-28)
│   ├── _authenticated/   projects, messenger, tms, calls, dashboard
│   ├── intranet/, crm/, hr/, accounting/, bases/, marketplace/
│   ├── calendar/, settings/, admin/
│   └── login, onboarding, status, call.$code
├── api/              HTTP clients per backend module
├── modules/          Feature modules (calls, dashboard, …)
├── components/       UI + Shadcn primitives
├── hooks/            useQuery/useMutation wrappers
└── stores/           Zustand (auth, messenger, …)
```

**State:** TanStack Query (server) + Zustand (client). **Editor:** Plate (rich-text). **Realtime:** WebSocket client → `/ws`.

### HTTP-запрос

```
Browser → Caddy → Fiber Middleware → Handler → Service → Repository → DB
                  (RequestID, Recover, SecurityHeaders, Logger, CORS, RateLimit)
                  api group: APIKeyAuth, per-route auth/RBAC limits
```

Фактический порядок middleware (`backend/cmd/server/wiring_fiber.go`):
1. `middleware.RequestID()`
2. `fiberrecover.New()` (не `middleware.Recover`)
3. `observability.HTTPTraceMiddleware()`
4. `middleware.SecurityHeaders()` — NF-08 (nosniff, DENY, CSP)
5. `middleware.Logger()`, `middleware.CORS()`
6. `GET /metrics` + `observability.HTTPMetricsMiddleware()`
7. `middleware.RateLimit()` — 100 req/sec (API)
8. `api` group: `MaintenanceMiddleware` → `HTTPAuditMiddleware`
9. Per-route: `/files` 10/min, `/messenger` 60/min; `/auth` 3/10min (Redis, off in DevMode)
10. `apigatewaymod.APIKeyAuth()` — на `api` group **до** `RegisterRoutes` (SEC-01; wiring auth/modules)

### События в реальном времени
```
Domain Event → Outbox (PG) → Worker → NATS → WS Bridge → Client
```

### Уведомления
```
NATS Event → JetStream Consumer → NotificationService → In-App + Telegram + Push
```

### Загрузка файлов
```
Client → POST presign URL → Backend (RBAC check) → SeaweedFS PUT (direct)
```

## Схема развёртывания

| Окружение | Compose-файл | Proxy | Особенности |
|-----------|--------------|-------|-------------|
| Development (full stack) | `docker-compose.local.yml` (корень репо) | Caddy :8080 | DEV_MODE, hot reload |
| Development (infra only) | `deployments/docker-compose.yml` | direct | PG, Mongo, Redis, NATS, S3, LiveKit |
| Production | `deployments/docker-compose.prod.yml` | Caddy | Replicas, secrets via env |
| QA tests | root `docker-compose.local.yml` | `run --rm <exact-service>` | 8 disposable QA services; integration/e2e infra isolated |
| Monitoring (optional) | `deployments/docker-compose.monitoring.yml` | — | Prometheus, Loki, Grafana |

**K3s:** частичные манифесты в `deployments/k3s/`; полный prod stack — Compose-first.

**Observability:** конфиги в `deployments/monitoring/`; `docker-compose.monitoring.yml` — optional stack.

Компоненты в prod compose: backend, frontend (static), postgres, mongodb, redis, nats, seaweedfs, livekit, caddy.

## Безопасность на уровне архитектуры

| Слой | Меры |
|------|------|
| Edge | Caddy TLS, rate limiting |
| Auth | JWT 15min + Redis refresh (hashed), Magic Link rate limit |
| AuthZ | RBAC 3 уровня + founder; Redis cache с инвалидацией |
| Data | User-scope в search, channel membership в messenger |
| Network | SSRF protection (`safehttp`), no `http.DefaultClient` |
| Secrets | Env vars only, AES-GCM at rest для tokens |
| WS | Origin check + `ws.Authorizer` (channel/project/call membership; `wiring_ws.go` + `internal/ws/`) |

## Ключевые паттерны

| Паттерн | Где | Зачем |
|---------|-----|-------|
| Outbox | `platform/outbox` | At-least-once delivery events |
| Optimistic Locking | Ent `version` field | Concurrent edits |
| Soft Delete | Trash 30 days | Recovery, GDPR path |
| Presigned URLs | `storage/` | Direct upload, no backend buffer |
| Repository | `*/repository.go` | Testability, layering |
| Cron helper | `platform/cron` | ME-10 graceful shutdown для tickers |

## Phase 151 — bounded selected-reply retention (2026-08-26)

Within `MessengerService.PromoteToIssue`, the private bounded reply page remains repository-owned input. If the promoted source is a reply absent from that newest-100 page, the service constructs a fresh effective page: append for a short page, or first 99 rows plus the selected reply for a full page. A separately resolved parent is not counted as a reply; an already-present source remains unchanged. Duplicate reply IDs and pages above 100 are rejected as redacted 500 before GetChannel and all promoter/backlink/linker effects. The contract is unit-only: live Mongo/Ent-PostgreSQL/NATS/WS, delete/revoke TOCTOU, post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create), and canonical backlink/post-create lifecycle remain `NOT_RUN` or `OPEN/UNBOUND` as applicable.
## Top-5 архитектурных нарушений (2026-06-17)

| # | Нарушение | Файл:строка |
|---|-----------|-------------|
| 1 | Handler обходит Service — прямой Ent query через escape hatch | `backend/internal/tracker/handler_issues.go:269` |
| 2 | Repository interface утекает `*ent.Client` (`GetDB()`) | `backend/internal/tracker/repository.go:153` |
| 3 | HTTP middleware читает Ent напрямую (deactivated user check) | `backend/internal/auth/middleware.go:120` |
| 4 | RBAC middleware принимает `*ent.Client` (layering leak) | `backend/internal/rbac/middleware.go:46,114` (+ wiring в `wiring_*.go`) |
| 5 | Dev handler обходит Service (`//go:build !release`) | `backend/internal/dev/handler.go:102` |

## Технический долг (известный)

| ID | Проблема | Evidence | План |
|----|----------|----------|------|
| TD-01 | ~~DI-точка раздута~~ | thin `main.go` + `wiring_*.go`; cron в `platform/cron` | **DONE** (v2.6) |
| TD-02 | `*ent.Client` в 7 domain-пакетах | `auth/service*.go`, `dashboard/service.go:34`, `backup/service.go:23`, `imex/service.go:23`, `storage/service.go:36`, `apigateway/*.go:40+`, `rbac/service.go:19` | Repository Phase 19 |
| TD-03 | Escape hatch `GetDB()` | `tracker/repository.go:153`, `handler_issues.go:269` | Убрать, JQL search → service |
| TD-04 | K3s partial | `deployments/k3s/` exists; full prod manifests incomplete | DevOps backlog |
| TD-05 | Observability optional stack | `docker-compose.monitoring.yml` + configs; not default prod | Enable per env |
| TD-06 | ~~Security headers~~ | **FIXED** `wiring_fiber.go` + `security_headers.go` | — |
| TD-07 | Handler→Service без интерфейсов | все `handler.go` — `*XxxService` concrete type | v2.6 (низкий приоритет) |
| TD-08 | `audit/retention.go` вне Repository | `retention.go:41` — отдельный `RetentionService` на `*ent.Client` | Phase 19 |
| TD-09 | `notification/push.go` на Ent | `push.go:45` — `PushSender` держит `*ent.Client` | Phase 19 |
| TD-10 | CRITICAL security (v2.0.1) | **12/12 FIXED**; PG-03 rotation manual | ops runbook |

## Реализованная архитектура v3–v4 (2026-06-21)

> Ранее «planned» — теперь в monolith `cmd/server`. Детали: `.planning/decisions/specs/PLATFORM-VISION-v4.md`, `.planning/archive/roadmaps/MILESTONE-3.md`

| Версия | Модули в коде | Frontend routes |
|--------|---------------|-----------------|
| v3.0 | `ai/` — providers, bot, JQL, PR review | `/settings/ai` |
| v4.0 | `intranet/`, `calendar/`, `crm/`, `hr/`, `accounting/`, `bases/`, `marketplace/`, `status/`, `mail/` | `/intranet`, `/calendar`, `/crm`, `/hr`, `/accounting`, `/bases`, `/marketplace`, `/status` |
| v4.5 | `integrations/onec/` | settings integrations |

Event bus: **NATS JetStream** (без Redis pub/sub). Search: PostgreSQL + **Meilisearch** indexer (`cmd/meilisearch-indexer`).

## Связанные документы

- `docs/system_specification.md` — API, модели данных
- `docs/security_guidelines.md` — OWASP, уязвимости
- `docs/stack_and_tools.md` — версии стека
- `.planning/decisions/specs/PLATFORM-VISION-v4.md` — corporate super-app vision
- `.planning/decisions/specs/AI-PROVIDER-HUB-SPEC.md` — AI gateway spec

---
*Последнее обновление: 2026-06-22 — RU-FIX-01: RU section headers*

## Messenger authority boundary — Phase 151 bounded slice (2026-08-26)

For channel-scoped Messenger behavior, Ent/PostgreSQL grants access and MongoDB stores messages but grants nothing. The backend resolves fresh `ChannelAccess`, projects recipient capabilities into channel DTOs, and authorizes message mutations before any message lookup. The frontend renders and executes actions only from that capability snapshot and also checks the direct `channel_removed` revoke barrier immediately before network work. Recipient projection is applied not only to Get/List, but also to Create/Update channel, new/existing DM, and AcceptInvite return paths so a successful operation cannot emit a capability-less DTO in production.

Channel discovery is a two-stage read boundary: an Ent/PostgreSQL query selects at most 100 live public tenant/project candidates in deterministic descending creation order, then the service re-authorizes each candidate with fresh `ReadTimeline` before response projection. Denied candidates are omitted and unexpected repository/authority errors fail the whole request. This browse seed has no Mongo/event/membership side effects and is not a paginated complete directory.

The frontend consumes discovery through a separate user-keyed lazy query. While fetching or after any error it renders no cached rows; 401/403 disables retry and removes the exact identity payload. The Browse dialog is metadata-only and deliberately has no navigation or membership control until the grant/revoke coordinator is production-bound.

Forwarding uses two independent resources: read/forward authority on the source and post authority on the target. Identity changes reset user-scoped channel/unread caches before the next room subscription. Reconnect and realtime notifications trigger authoritative refetch rather than client unread arithmetic.

Saved-message projection uses the same two-database boundary: PostgreSQL bookmark rows are authorized by their channel before Mongo enrichment, and Mongo content is projected only when the message remains bound to that channel. The frontend stores bookmark results under a user-keyed query, masks all cached data without identity, and fences remove/navigation completion against the current auth-store identity. Task 95 now performs one proof per distinct bookmark `channel_id` and completes the full authority phase before Mongo enrichment; live database execution remains unverified.

Pinned-message projection applies a second channel boundary after the fresh requested-channel proof: only non-nil rows whose `ChannelID` equals the requested channel reach the response. Production Ent already scopes its query, so this is defense in depth and has only unit/race evidence.

Pin toggling uses Ent error classification as a write fence: NotFound means create, a found row means delete, and every other lookup error terminates before either mutation. The repository error itself is not exposed to the client.

The pinned-message panel and the root timeline now share explicit identity boundaries. Pin rows use messenger/pins/channelId/userId; message previews, fetch, reconnect invalidation, WS deltas, generic invalidation, and optimistic edit use messenger/messages/channelId/userId. No-user state disables room/query/event work and masks both flattened messages and raw observer data, including legacy and sentinel cache payloads. Delayed A results stay addressable only by A while B observes B. Paged replies and production thread detail use separate channel/root/user schemas to avoid cache-type collision; no-user and delayed A-to-B data are masked in both. Async identity cleanup is schema-aware and preserves only current-user keys, while revoke cleanup remains channel-index compatible. ThreadPanel now treats the detail query as parent-only and renders replies from the WS-aware paged cache. The thread observer retains bounded per-scope sent-message and mutation journals across initial fetch/refetch, overlays edit/delete/reaction patches on both raw and flattened public data while canonical recovery is unresolved, coalesces overflow/reconcile into exact recovery refetches, and masks potentially stale output on journal overflow. Identity/revoke cleanup clears every journal. Pending sends capture the authenticated principal, and thread issue-link queries/mutations are identity-scoped and revoke-fenced.

Thread issue-link listing crosses Tracker/PostgreSQL and Messenger channel authority. Route-level project view is necessary but insufficient because RBAC can be cached and system admins bypass project membership: the Tracker service resolves the exact non-deleted project channel, proves a fresh actor↔ProjectMember row, proves actor↔ChannelMember, and only then reads link rows. Hidden channel variants, project-ineligible actors, and channel non-members collapse to the same 404; proof infrastructure failures share one redacted 500. This closes reliance on route-RBAC freshness but remains a bounded sequential proof: revocation between either membership proof and link read is a documented TOCTOU until an atomic authorized-list repository query exists. Create/unlink and UI capability policy are not inferred from this list contract.

The OpenAPI contract is stored as three checked-in projections: `docs.go`, `swagger.json`, and `swagger.yaml`. Task 128 applies one semantic unit oracle to all three through a read-only `backend/docs` container mount; the test never invokes a generator. This prevents stale-image false-green and projection drift for the scoped promotion route and durable unread schema, but source-annotation alignment remains a separate regeneration residual.

Promotion now has a bounded cross-store source authority adapter. Ent/PostgreSQL first proves `ChannelActionPromote(A)`; the service then requires a private capability implemented by the concrete Mongo repository and performs one live source lookup bound to channel A. The atomic Mongo predicate is `_id + channel_id + deleted_at:nil`, so a foreign/deleted/missing source is never decoded through the promotion path and collapses to one concealed 404. Capability, store, and invariant failures stop with a redacted 500 before target-project, issue, link, backlink, or event work. Parent/reply lookups, post-read deletion TOCTOU, and canonical backlink/post-create processing remain separate `UNBOUND` boundaries; target-project freshness is handled by the following bounded adapter.

The target-project decision is now a second bounded adapter. The Messenger port receives `(targetFound, canCreate, error)` from RBAC; the RBAC implementation bypasses Redis and queries active actor organization, exact active/non-soft-deleted same-organization project, and current membership/custom role. Founder/admin bypass is evaluated only after target proof. This closes stale-cache and target-existence decisions at unit level, but the sequential Ent reads and later issue mutation are not atomic. Live PostgreSQL query behavior, parent/reply channel integrity, post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create), and post-create processing remain `NOT_RUN/UNBOUND`.

Pin-list projection is also a capability boundary: the header surface and query exist only with `can_list_pins`; the unpin control additionally requires `can_pin`. Mutation preflight and post-await completion intersect the latest `ChannelResponse` snapshot with current store capabilities and the direct revoke marker before the exact user/channel cache can be invalidated.

Message-row pin mutation and pin-list projection converge through exact query invalidation rather than optimistic row construction: the toggle response carries only `pinned`, not pin id/actor/time. The row captures user/channel, rechecks latest DTO and store capability plus revoke before and after await, and invalidates only the corresponding identity-scoped pins key.

Production WS wiring uses `WSChannelAuthority` for room joins, `messenger.user.*` channel lifecycle delivery, and channel audience resolution. Each path resolves fresh `ChannelActionReceiveEvent`; a denied candidate never reaches the bridge. A valid `channel_created` event may clear the frontend revoke marker, but only before an authoritative channel refetch; ordinary stale snapshots cannot bypass `channel_removed`.

The frozen membership grant/revoke coordinator and canonical post-create/pending-recovery coordinators remain `UNBOUND` in production. The lifecycle APIs still need an unambiguous initiator/subject contract, trusted internal issuer proof, baseline/final CAS, and channel-scoped reset/fencing. The pure message seam now accepts and deep-copies a full prepared `MessageDocument`, freezes resolved mentions, assigns pending/UTC server fields, rejects server-owned state, and validates JSON before repository access. Safe production binding still requires nonce fingerprint/version semantics, effect idempotency or durable processing claim/fencing, recovery ownership, manual-unfollow precedence, complete effect implementations, and one event publisher.

This is unit-verified architecture and production-wiring evidence only. Real PostgreSQL/Mongo/NATS/WS delivery, browser, cutover, and Phase 149 integration evidence is `NOT_BOUND/NOT_RUN`.

## Promotion parent/reply authority boundary (2026-08-26)

After fresh route/source and target proofs, promotion resolves parent and replies only through private channel-bound live Mongo predicates: parent is channel-scoped; replies require `channel_id + thread_id(rootID) + deleted_at:nil`. The adapter validates every returned row, forbids generic fallback, and selects at most 100 rows in `created_at DESC, _id DESC` order before GetChannel/promoter/backlink/linker. This is a bounded unit-level authority seam, not a lifecycle claim: live stores/events are `NOT_RUN`; cross-store delete/revoke TOCTOU, post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create), and canonical backlink/post-create lifecycle remain `OPEN/UNBOUND`.

## Promotion source-ID defensive boundary (2026-08-26)

The scoped source adapter rejects a `found=true` result unless it is nonnil, live, bound to the route channel, and has a nonempty ID exactly equal to route `messageId`. It returns fixed/redacted 500 before target/thread/channel/effects for malformed found data; `found=false` remains a concealed 404. This is unit-level only: live Mongo/Ent-PostgreSQL/NATS/WS are `NOT_RUN`; cross-store delete/revoke TOCTOU, post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create), and canonical backlink/post-create lifecycle remain `OPEN/UNBOUND`.

## Promotion mutation-adjacent target authority boundary (2026-08-26)

`PromoteToIssue` snapshots one authority-checker instance, uses it for an early fresh target proof, and uses the same instance for a second fresh proof with the identical actor/project after parent/reply/GetChannel/all description display-name reads and immediately before `CreateIssueFromPromotion`. Both proofs preserve fail-closed mapping: nil/error/false-true → redacted 500; false-false → concealed 404; true-false → 403. The second rejection terminates before promoter/backlink/linker effects. This improves the sequential boundary at bounded unit level, with root/reply order and same-instance replacement protection; it does not make authorization plus issue creation atomic. Live Mongo/Ent-PostgreSQL/NATS/WS, parent/reply/channel delete/revoke TOCTOU, post-recheck source deletion before issue creation, canonical backlink/post-create lifecycle, and post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create) remain `NOT_RUN` or `OPEN/UNBOUND`.

## Promotion chronological description boundary (2026-08-26)

The Mongo adapter continues to select the bounded newest-100 reply page in `created_at DESC, _id DESC`. `MessengerService.PromoteToIssue` performs selected-reply retention first, then fails closed unless every effective row has identity/time and the whole page is strictly `CreatedAt DESC, ID DESC`. This check runs before GetChannel, display assembly, or effects. Accepted rows are copied and only that display copy is sorted by `CreatedAt ASC, ID ASC`; the separately resolved parent stays first and outside the reply budget, while repository-owned ordering remains untouched. This is unit-level presentation evidence, not a live-store or transaction claim: live Mongo/Ent-PostgreSQL/NATS/WS are `NOT_RUN`; parent/reply/channel delete/revoke TOCTOU, post-recheck source deletion before issue creation, canonical backlink/post-create lifecycle, and target authorization-plus-create atomicity remain `OPEN/UNBOUND`.

## Promotion dependency snapshot boundary (2026-08-26)

`MessengerService` owns one promotion-configuration `RWMutex` for `IssuePromoter`, `PromotePermissionChecker`, and `IssueThreadLinker`. A promotion captures all three once at request entry and releases the lock before any repository or external dependency call, so current-request behaviour cannot mix old/new objects and callback reconfiguration cannot deadlock. Hubtest field exports acquire the same snapshot. The three public setters remain separate and are not an atomic hot-reload transaction; production wiring occurs before route registration. This race-safety atom does not close live-store, cross-store delete/revoke, target authorization-plus-create, or canonical post-create lifecycle boundaries.

## Promotion mutation-adjacent live-source boundary (2026-08-26)

`PromoteToIssue` resolves the route-channel source through one shared exact live-source mapper at request processing start and again after description assembly. The second source proof precedes the already-final target proof, so tracker mutation cannot proceed from a source known missing/deleted at that point. Both reads preserve the same concealed 404 versus redacted 500 contract, and no promoter/backlink/linker effect runs on second-proof rejection. This remains sequential rather than transactional: source deletion after recheck, parent/reply/channel changes, target changes after proof, and canonical post-create recovery are separate boundaries; live stores/events remain unexecuted.

## Phase 151 promotion live-parent recheck boundary (2026-08-26)

For reply-to-issue promotion, parent resolution is an exact `(channel_id, root_message_id, deleted_at:nil)` live-message capability. The same mapper runs once before reply loading and again after description construction plus the live-source recheck. The final target-project authority proof remains the last check before tracker mutation. Root promotion skips both parent reads. This narrows a sequential TOCTOU window but is not a cross-store transaction: source/parent deletion after recheck, reply/channel revoke-delete, target change after proof, and backlink/post-create recovery remain separate boundaries. No live Mongo, Ent/PostgreSQL, NATS, or WebSocket evidence is inferred from the unit gate.

## Promotion mutation-adjacent source-channel authority boundary (2026-08-27)

The promotion service owns one mapper for configured `ChannelActionPromote` authority and the legacy repository-membership fallback. It invokes that mapper at request entry and again after the route source and optional root have been revalidated, before the final target-authority proof and tracker mutation. Both calls preserve identical public error semantics and the second failure stops before target recheck or post-create effects. The boundary intentionally remains sequential: it narrows authorization drift but does not provide an atomic Mongo/channel-membership/Ent/tracker transaction or durable post-create lifecycle.

## Phase 151 canonical create-result boundary (2026-08-27)

The pure post-create coordinator now treats repository `CreateMessageResult` data as an untrusted boundary. One common validator runs after canonical nonce lookup and after pending persistence, before every return/effect branch. It binds message ID, channel, sender, nonce, creation provenance for lookup hits, and durable pending/complete state agreement; malformed data fails closed with no derived effects. Legacy documents without an effect state are read as complete for compatibility, but the result state must still be explicit. This is defense in depth over the concrete Mongo adapter, not production binding: the following strict v1 boundary now covers pure-seam nonce payload/thread fingerprinting; live writer cutover, live invocation of the frozen typed nonce mapper, effect/recovery fencing, outbox ownership, and cross-store lifecycle remain separate `UNBOUND` architecture decisions.

## Phase 151 durable client-nonce fingerprint v1 boundary (2026-08-27)

Original create intent is now a versioned internal durability boundary rather than a comparison against mutable current message content. The pure preparer canonicalizes the full producer/create payload and stores a SHA-256 v1 digest; the common result validator checks that metadata after lookup and persist before all effects. Set-like mention fields are normalized, attachment order stays semantic, and server/mutable fields are excluded. This prevents retries after edits from redefining intent while still rejecting cross-producer or changed payload reuse.

The seam intentionally rejects missing or unknown versions. It does not infer version 0 from current content because the trusted producer and pre-edit payload are not reliably reconstructable. The later legacy atom therefore reserves fully unversioned nonces without replay or backfill. Live mapper invocation, effect/recovery fencing, publisher ownership, and runtime stores remain separate boundaries. The digest is an equality token, not an authorization credential or secret.

## Phase 151 typed client-nonce outcome boundary (2026-08-27)

The common pure-seam canonical validator classifies nonce ownership only after exact identity and durable state validation. Lookup hits and `Created=false` persist races are existing records: deletion reserves the nonce before fingerprint metadata inspection, while a canonical v1 fingerprint that differs from the request becomes a typed conflict. `Created=true` data stays repository-owned, so deletion or fingerprint mismatch is an invariant failure. Strict v1 form is lowercase 64-hex; missing, unknown, uppercase, or malformed metadata fails closed.

The two exported errors are domain-level control signals, not a transport or security authority. This outcome atom added no HTTP adapter, production writer, migration policy, store transaction, recovery claim, or publisher; the following atom adds only the pure redacted adapter. The sequencing protects the no-effect pure coordinator but does not make Mongo persistence, derived effects, or event delivery atomic.

## Phase 151 post-create service error mapper boundary (2026-08-27)

A stateless mapper is now the only permitted translation from pure post-create domain outcomes to AppError. It recognizes wrapped sentinels with `errors.Is`, preserves the legacy split between two redacted 409 nonce messages, and collapses every malformed, unauthorized-internal, wrapped, or repository failure to one fixed 500. Nil remains nil. Returning a fresh AppError prevents pipeline cause/detail metadata from entering the centralized public error body.

The adapter is compiled and unit/race verified but remains unreferenced by production writers. It does not authorize a producer, execute the coordinator, select a legacy-record policy, persist data, publish events, or make cross-store effects atomic. Live invocation is therefore still an explicit cutover boundary.

## Phase 151 legacy client-nonce reservation boundary (2026-08-27)

Legacy compatibility is deliberately non-migratory. Once exact canonical identity and durable state are valid, a nondeleted existing document whose fingerprint and version are both zero-state becomes a typed legacy reservation. The coordinator returns no canonical success and runs no effects. Deleted reservation remains earlier in precedence; partial metadata and all `Created=true` zero-state results remain internal invariant failures.

This prevents mutable current content from being treated as historical create intent and avoids a bulk/backfill requirement. The mapper deliberately collapses the legacy sentinel into the generic conflict 409, so clients cannot infer record age. This closes pure classification only; live writer invocation, effect ports, recovery fencing, stores, and events remain unbound.

## Phase 151 Mongo reply-count reconciliation boundary (2026-08-27)

The first concrete canonical post-create effect adapter is `MongoRepository.ReconcileReplyCount`. A pure exported plan freezes the persistence boundary: count `channel_id + thread_id + deleted_at:nil`, then update the exact `_id + channel_id + deleted_at:nil` root only when its `thread_id` is missing or empty. The adapter also revalidates trusted producer shape, actor/channel identity, live pending state, and reply/root distinction before I/O.

The write uses `$max` with the observed live direct-reply count. This makes crash replay a no-op and prevents a lower concurrent create observation from lowering a higher count; `MatchedCount==1` is success even when Mongo reports no modification. Zero observations and missing/deleted/foreign/nested roots fail closed.

This is not an exact cross-lifecycle counter transaction. It cannot lower a historically inflated counter, and a concurrent legacy delete/decrement can race between count and raise. No handler or writer calls the coordinator, no live Mongo test ran, and automatic follow, unread invalidation, event/outbox ownership, recovery fencing, and cutover remain separate boundaries.

## Phase 151 automatic thread-follow effect boundary (2026-08-27)

Automatic follow is deliberately a cross-store adapter on `CombinedRepository`, not a Mongo-only method. `EnsureAutoThreadFollows` constructs a fresh `PGAuthorizedUnreadScopeResolver` from the embedded Ent client and delegates to a pure coordinator plus the existing Mongo transition writer. This preserves PostgreSQL as active tenant/account/channel/project-membership authority immediately before every follow mutation.

The immutable plan orders the canonical author/replier first, then first-occurrence unique `resolved_mention_user_ids`; a root follows its own ID and a reply follows its `thread_id`. Every candidate uses `ThreadFollowModeAuto` and the triggering UTC `(created_at,message_id)` position. Typed current denial or absence of the exact channel target omits that candidate. Mismatched tenant/user facts and infrastructure errors fail pending; successful earlier candidates remain safe because Mongo auto transition is insert-only and never overwrites `manual_follow`, `manual_unfollow`, or existing `auto`.

The adapter uses real production dependency types but is not wired into any live canonical writer. The per-user PG-read/Mongo-write sequence is not atomic, and authorization may change after proof; later reads still require fresh PostgreSQL authority. No live Ent/Mongo evidence or recipient event ownership follows from this unit atom.

## Phase 151 post-create unread invalidation adapter (2026-08-27)

The bounded post-create architecture now has a concrete recipient-only invalidation effect. `Phase151UnreadInvalidationEffect` consumes the existing `WSChannelAuthority` audience projection, which starts from PostgreSQL channel membership and re-authorizes every candidate, then publishes `messenger.user.{userID}.unread_changed` through the current best-effort `EventPublisher`. The sender, nil IDs, and duplicate recipients are filtered. Delivery-time WS handling independently re-authorizes the exact recipient and current audience.

The event carries identifiers only. A root carries `channel_id` with `target_kind=channel/category=channel`; a direct reply additionally carries its UUID `root_message_id` with `target_kind=thread/category=followed_thread`. It never contains content, unread counts, or recipient-local state.

The coordinator calls this effect only after a successful complete CAS. NATS publication is not transactional and has no acknowledgement through the current publisher interface, so CAS-to-publish failure can lose an invalidation. Reconnect/subscription followed by authoritative refetch is the convergence design; no outbox or exactly-once claim exists. This adapter remains standalone until eligible-mention resolution and full repository composition are complete, and production message writers remain on legacy paths.

## Phase 151 eligible mention resolver adapter (2026-08-27)

Mention resolution now has a standalone pre-persist adapter. The actor is rechecked through the PostgreSQL `ChannelAccessResolver` for exact tenant/channel/project access. Candidate discovery comes from the same current `WSChannelAuthority` audience used for recipient events, never Mongo message data. Project-role results from Ent and online IDs from Presence/Redis are only filters feeding that authorized audience.

Individual UUIDs, `@all`, `@online`, and `owner/developer/tester/viewer` groups form a union. The sender, invalid IDs, users outside the current audience, and duplicates are removed, and UUID sorting freezes a deterministic persisted `resolved_mention_user_ids` snapshot. Empty input avoids dependency reads; infrastructure failures prevent canonical Mongo persistence.

The adapter is not yet composed with the Mongo/CombinedRepository effects and no production writer calls it. `WSChannelAuthority` treats one candidate authorization error as an omission, so the boundary is fail-closed for disclosure but live completeness can undercount a non-sender until a future strict-audience policy is chosen. PostgreSQL/Redis integration and recovery behavior remain unverified.

## Phase 151 post-create repository composition (2026-08-27)

`Phase151PostCreateRepository` is the standalone composition root for all seven canonical post-create repository ports. It embeds a complete `CombinedRepository`, which promotes the bounded Mongo canonical persistence, reply-count, completion-CAS, and Combined automatic-follow methods, while eligible mentions and recipient unread invalidation remain explicit delegates. The constructor rejects missing Ent/Mongo members and plain or typed-nil delegates, so promoted methods cannot be reached from a partially built composition.

This layer changes no live writer. Its unit contract proves interface satisfaction, explicit delegation/error ownership, defensive mention-result copying, and fail-closed nil behavior for the two explicit ports; promoted storage methods and live stores were not exercised. Recovery claim/lease/fencing and a claimed effect runner now exist as standalone seams, while loop binding, writer cutover, outbox loss, membership lifecycle, cross-store TOCTOU, and create/delete counter races remain separate architecture boundaries.

## Phase 151 post-create recovery claim/completion-fence plan (2026-08-27)

Pending canonical messages now have a bounded ownership model: server-owned worker UUID, monotonic fence, and UTC lease deadline. An atomic claim predicate is scoped to exact message/channel/sender, live `pending` state, no completion/deletion, and either no prior lease or an expired lease. Its update sets owner/deadline and increments the fence. Completion binds the same trusted tenant/channel/sender plus exact owner/fence/lease, requires a deadline later than completion time, transitions to `complete`, records completion, and clears owner/deadline while retaining the fence.

This atom emits BSON plans only. The separate Mongo adapter applies them, the claimed runner validates claim tokens and durable provenance, and the coordinator adapter can compose that runner behind the bounded scan loop. No concrete scanner or scheduler constructs the composition, so production recovery is not yet fenced. Lease renewal is absent. Crash takeover may replay idempotent effects and unread publication remains at-least-once/best-effort without a transactional outbox.

## Phase 151 Mongo post-create recovery claim/completion adapter (2026-08-27)

`Phase151PostCreateRecoveryMongoAdapter` separates recovery orchestration from Mongo mechanics through a narrow BSON store. Production claim uses a single return-after `FindOneAndUpdate`, so the acquisition token comes from the same atomic mutation that set owner/lease and incremented fence. The adapter rejects contradictory/malformed results and returns owned copies of the mention snapshot and lease metadata. Fenced completion runs the exact plan with `UpdateOne`; no match means the claim is stale or lost.

This adapter is not a recovery worker. Durable trusted-producer provenance and a standalone claimed effect runner now consume its acquisition contract, but the existing PG target pager/Mongo pending scanner does not invoke either seam. Live Mongo behavior, loop binding, lease renewal, crash-time effect replay, and best-effort NATS delivery remain separate boundaries.

## Phase 151 durable post-create producer provenance (2026-08-27)

Every new pending canonical message prepared by the bounded pipeline carries a server-owned `post_create_producer_provenance` BSON subdocument. Version `1` stores the closed `TrustedMessageProducer`, exact `BotKind`, and integration UUID. Request data cannot preseed it; canonical readback and the concrete Mongo persist boundary require an exact match with the authorized producer scope before derived effects. Recovery claim and completion predicates also match the entire subdocument, while return-after claim decoding revalidates scope equality and detaches the effect input from the store-owned provenance pointer.

Recovery reconstruction separates two authorities. The current target lookup must freshly authorize tenant/channel access, while immutable message data supplies the original sender and versioned producer proof. The reauthorizing actor is never impersonated as the sender. Reconstructed scopes pass the same closed-producer/bot validation used by normal post-create effects; missing, foreign, deleted, complete, unknown-version, unknown-producer, or widened bot records fail closed.

The claimed runner now connects provenance reconstruction, the claim adapter, and derived-effect ports in memory, but the pager/scanner and live writers remain disconnected. There is still no live cross-store transaction or exactly-once guarantee. Completion-before-unread-publication loss, lease renewal, integration revocation policy, and outbox delivery remain explicit later boundaries.

## Phase 151 claimed post-create recovery effect runner (2026-08-27)

The standalone runner accepts fresh target authority, one scanned pending canonical message, an injected clock, worker identity, and the composed recovery repository. It reconstructs the original producer scope, calculates the expected lease, atomically claims, validates the exact post-update message plus owner/fence/lease and provenance, reconstructs authority again, then executes automatic follows and optional reply-count reconciliation before fenced completion. Best-effort unread invalidation is deliberately last. A clean miss performs nothing; every failure after acquisition is reported with `acquired=true` so an outer loop can distinguish contention from owned work that needs observation/retry.

This is orchestration evidence, not a bound worker. A coordinator adapter now supplies the runner seam behind the existing bounded loop interface, but no concrete scanner or production scheduler constructs it and no writer is changed. Effects are designed to be replay-safe but are not transactional with Mongo completion, no lease renewal protects a long-running effect sequence, and a crash between completion and unread publication can lose that notification until authoritative refetch. Live-store and integration behavior remain unverified.

## Phase 151 recovery coordinator-to-runner adapter (2026-08-27)

`Phase151ClaimedRecoveryCoordinatorAdapter` splits the old bounded coordinator repository into a scan-side port and the claimed effect repository. Interval acquisition, stable target paging, fresh reauthorization, and bounded pending-document listing remain scanner responsibilities. The fifth coordinator call invokes the claimed runner with immutable worker UUID, lease, clock, and effect dependencies. A valid fresh target must retain exact tenant/channel identity, a nonnil actor, and a closed producer proof before Mongo scan.

The adapter treats claim non-acquisition as ordinary contention so the bounded loop can continue to later documents. Runner errors preserve identity and stop the pass; malformed construction, target, authority, context, or scanner input fails closed. This is a composition seam only: there is no concrete production scanner implementation or scheduler construction yet, so no live recovery behavior or writer cutover follows from the interface satisfaction.

## Phase 151 recovery scanner authority gap (2026-08-27)

Concrete composition cannot safely choose fresh recovery authority yet. `Phase151PendingRecoveryTarget` identifies tenant/channel/kind but carries no actor or service principal, while production provides no recovery interval gate, stable PostgreSQL target pager, or reauthorizer. The next architecture decision must introduce an explicit server-owned recovery principal or redesign the target-authority facts contract. Selecting an arbitrary current member or channel owner would change audit and authorization semantics and is forbidden. Until the decision is frozen, scanner construction, scheduling, and live writer cutover remain unbound.

## Calls provider seams and Messenger completion CAS (2026-08-27)

`GenerateJoinCredential` is the provider-neutral Calls service boundary. It preserves the complete typed credential for a future Jitsi-native surface, while legacy `GenerateToken` remains a compatibility projection. The pure frontend Jitsi embed contract accepts only the configured self-hosted origin, `jitsi` provider, canonical room name, and JWT; it performs neither DOM/network work nor token logging. LiveKit remains the active provider: there is no public join DTO, Jitsi route/iframe, deployment, media plane, cutover, recording, or live-runtime claim.

The next Calls decisions are browser-reachable LiveKit public URL semantics, the contradiction between canonical `call.RoomName` and issuer identity, a call-aware issuer, expiry/refresh, and guest policy. Separately, the Messenger completion CAS matches the exact trusted producer scope and pending live document before setting only completion state/time. Its recovery scanner/scheduler/authority and direct-writer cutover remain unbound.
The production Messenger graph now constructs and stores the complete Phase 151 post-create repository behind an additive service constructor. A shared PostgreSQL-backed audience projection feeds eligible mentions and unread invalidation; storage ports remain on CombinedRepository. No writer consumes the dependency yet, so this is composition reachability rather than behavior cutover.

## Calls and Messenger bounded hardening (2026-08-28)

Jitsi issuance now treats server-owned `Call.RoomName` as provider room identity and exposes an internal typed expiry equal to the JWT expiry (`343d1d0d`, `366bb825`). These are provider-adapter seams only: LiveKit remains active and self-hosted Jitsi deployment, public routing/DTO, browser embedding, media runtime, and browser E2E are `UNBOUND/NOT_RUN`.

Messenger pending recovery scans now use a pure exact Mongo plan (`44b3a95e`) for live, incomplete, provenance-v1 candidates in one authorized channel, with stable ordering and bounded limit. Lease eligibility remains the atomic claim boundary. `SendMessage` cutover and production recovery still require an explicit server-owned principal plus concrete pager, reauthorization, and scheduler composition.

## Typed Jitsi join projection seam (2026-08-28)

The internal Jitsi issuer now exposes `IssueJoinCredential` (`0a332760`): one private issuance call is projected into provider-neutral data containing exact provider `jitsi`, canonical `Call.RoomName`, validated self-hosted public URL, and token. Errors return a zero credential and the legacy token-only path remains unchanged. The seam is not connected to a provider factory, HTTP route/DTO, frontend, deployment, or cutover; LiveKit remains active.

## Inert Jitsi host and recovery-page boundary (2026-08-28)

`JitsiEmbedHost` (`ee0387e5`) is an inert validated mount seam: one empty ref target for valid self-hosted credentials and no credential persistence/exposure, network, script, iframe, provider API, or runtime activation. It is not connected to `CallPanel`, routing, deployment, or cutover; LiveKit remains active.

The Messenger recovery adapter (`ce46a0b2`) now owns and validates scanner cursors/pages, enforces strict tenant/channel/kind tuple progression and exact last-row `Next`, and suppresses partial results on scanner or validation failure. This does not provide the missing PG pager, recovery principal, interval gate, scheduler, or writer cutover.

## Messenger server-owned recovery authority seam (2026-08-28)

The frozen authority decision is now represented by closed `Phase151RecoveryServicePrincipal` value `messenger.post_create_recovery` and `Phase151RecoveryTargetAuthority{Principal,TenantID,ChannelID,Kind}`. The target contract intentionally has no actor/producer/bot/integration identity: exact validation accepts only the closed principal plus a structurally valid exact tenant/channel/kind target and otherwise fails with `ErrPostCreatePipelineInvalid`. This prevents arbitrary-member or channel-owner substitution from silently defining recovery authorization semantics.

The seam is not yet composed into recovery execution. Existing scanner/runner contracts still use `AuthorizedMessageProducerScope`; no PostgreSQL pager, reauthorizer, scheduler, production writer, or live recovery runtime is bound. The next safe atom is pure provenance reconstruction/claimed runner/coordinator contract migration to `Phase151RecoveryTargetAuthority`; all live-store/event/browser and integration evidence remains `UNBOUND/NOT_RUN`.

## Calls signed credential-expiry projection (2026-08-28)

Expiry is now provider-neutral data, but the signed credential remains its only source of truth. Jitsi reuses the single captured instant used to construct the JWT; LiveKit verifies its own freshly issued JWT and projects the verified claim. The adapter does not approximate `now + TTL`, log tokens or secrets, make network calls, or persist credentials.

This seam does not activate Jitsi or alter the conference lifecycle. LiveKit remains the active provider; public join/refresh DTOs, call-aware service authorization, client consumption, self-hosted deployment, browser/media runtime, and cutover remain separate `UNBOUND/NOT_RUN` gates.

## Closed recovery provenance and call-aware issuer seams (2026-08-28)

Messenger now has a pure `Phase151RecoveryTargetAuthority` to canonical-producer-scope reconstruction seam. It validates the closed exact target tuple, sources tenant/channel only from that authority, and sources sender and producer metadata only from pending canonical-message provenance. It is not a database authorization or recovery-worker migration: scanner, runner, coordinator, PostgreSQL pager, reauthorizer, scheduler, and writer cutover remain `UNBOUND`.

`CallJoinCredentialIssuer` is an internal structural adapter over existing LiveKit and Jitsi issuance. Its shared validation rejects invalid server-loaded call and identity inputs, then delegates without loading calls, authorizing membership/RBAC/invites/guests, mutating state, or binding a service, factory, route, or client. Current Docker evidence for both seams is `NOT_RUN` because Mail owns Docker. LiveKit remains active; Jitsi self-host deployment, public API, client, cutover, and browser/media runtime remain `UNBOUND/NOT_RUN`.
