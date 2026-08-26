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

Within `MessengerService.PromoteToIssue`, the private bounded reply page remains repository-owned input. If the promoted source is a reply absent from that newest-100 page, the service constructs a fresh effective page: append for a short page, or first 99 rows plus the selected reply for a full page. A separately resolved parent is not counted as a reply; an already-present source remains unchanged. Duplicate reply IDs and pages above 100 are rejected as redacted 500 before GetChannel and all promoter/backlink/linker effects. The contract is unit-only: live Mongo/Ent-PostgreSQL/NATS/WS, delete/revoke TOCTOU, chronology policy, post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create), and canonical backlink/post-create lifecycle remain `NOT_RUN` or `OPEN/UNBOUND` as applicable.
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

After fresh route/source and target proofs, promotion resolves parent and replies only through private channel-bound live Mongo predicates: parent is channel-scoped; replies require `channel_id + thread_id(rootID) + deleted_at:nil`. The adapter validates every returned row, forbids generic fallback, and selects at most 100 rows in `created_at DESC, _id DESC` order before GetChannel/promoter/backlink/linker. This is a bounded unit-level authority seam, not a lifecycle claim: live stores/events are `NOT_RUN`; cross-store delete/revoke TOCTOU, chronology, post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create), and canonical backlink/post-create lifecycle remain `OPEN/UNBOUND`.

## Promotion source-ID defensive boundary (2026-08-26)

The scoped source adapter rejects a `found=true` result unless it is nonnil, live, bound to the route channel, and has a nonempty ID exactly equal to route `messageId`. It returns fixed/redacted 500 before target/thread/channel/effects for malformed found data; `found=false` remains a concealed 404. This is unit-level only: live Mongo/Ent-PostgreSQL/NATS/WS are `NOT_RUN`; cross-store delete/revoke TOCTOU, chronology, post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create), and canonical backlink/post-create lifecycle remain `OPEN/UNBOUND`.

## Promotion mutation-adjacent target authority boundary (2026-08-26)

`PromoteToIssue` snapshots one authority-checker instance, uses it for an early fresh target proof, and uses the same instance for a second fresh proof with the identical actor/project after parent/reply/GetChannel/all description display-name reads and immediately before `CreateIssueFromPromotion`. Both proofs preserve fail-closed mapping: nil/error/false-true → redacted 500; false-false → concealed 404; true-false → 403. The second rejection terminates before promoter/backlink/linker effects. This improves the sequential boundary at bounded unit level, with root/reply order and same-instance replacement protection; it does not make authorization plus issue creation atomic. Live Mongo/Ent-PostgreSQL/NATS/WS, source/channel/reply delete/revoke TOCTOU, chronology, runtime setter synchronization, canonical backlink/post-create lifecycle, and post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create) remain `NOT_RUN` or `OPEN/UNBOUND`.
