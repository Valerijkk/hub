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
├── mail/           current Resend alerts/admin; future mailbox domain is contract-only
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

`hub-mail` сейчас — Resend процесс для Magic Link/transactional alerts и существующего
legacy admin broadcast, а не SMTP/IMAP/JMAP mailbox engine. Broadcast нарушает no-marketing
policy `ADR-NOTIF-01` и остаётся governance gap до удаления/отключения либо явного
founder-решения. Предпочтительный Stalwart adapter, first-class mailbox domain и Calendar
RSVP/reliability описаны Mail 0.1.5 (`ADR-MAIL-01`, `MAIL-SPEC`, Phase 147 `CONTRACT.md`):
engine хранит raw mail, Hub — calendar/receipts/projections; availability не раскрывает
детали. L0 Inbox обновляет один context-owned bounded worker: replica-safe lease,
hydrated page and full-envelope/cursor atomic commit, separate expiring command lease with
begin/end mutation-epoch edges; sync capture is forbidden while
it is present. Recovery uses no-clear base-state/ID-anchored enumeration with catch-up.
Non-ready projection stays owner-filtered but visibly stale; disabled engine starts no
worker/probe. Reminders используют
durable in-app dedupe и повторную проверку текущих offset/
start/sequence/eligibility перед записью, а transient policy/store failure остаётся
retriable вместо ложного suppression; CRM создаёт обычный idempotent Mail draft и лишь
вручную связывает definitively owned message, в том числе из reader. Draft/message deep
links повторно проверяют Mail ownership. Для связанной встречи `CalendarEvent` — единственная
business-authority, а `ScheduledCall` — event-before-call locked execution projection;
aggregate/availability не считают пару дважды, due worker использует canonical exceptions и
durable versioned occurrence-keyed outbox. Exact-channel call UI only refetches canonical
occurrences on that realtime signal; it does not create a second reminder. Linked-call
attach is exact-match/fail-closed, and external guests stay ordinary-event-only until secure
join credential delivery is accepted. Implementation `NOT_IMPLEMENTED`, runtime `NOT_RUN`.

**Bounded Mail/Calendar Docker receipt (2026-09-01):** the QA composition exercised
the implementation slice for account materialization, live-thread FTS, invitation
correlation, Calendar reverse linkage, accepted RSVP, and quiescent fixture cleanup
with **7/7 scoped cases passing**. The slice keeps canonical iCalendar `PRODID`,
non-nil empty `exdates`, and storage-tolerant persisted PostgreSQL timestamp/nullable
`bytea` handling; its quiescent cleanup retry is isolated from strict one-shot
fault/guard deletion. This is not a full-app result: the aggregate integration job
is `NOT_GREEN` because five unrelated `project_outbox` setup cases have no available
Redis/NATS harness. Broader Mail 0.1.5 production composition and product acceptance
remain outside this receipt.

**Repository pattern (основные модули с interface):** audit, bases, calls, crm, github, gitlab, hr, intranet, marketplace, messenger, notification, project, search, sprint, tms, tracker и др.

**Ent schemas (измерено 2026-08-13):** **85** entity schema types в **87** `.go` files
(`helpers.go` + `mixin_version.go` не являются entity schemas).

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

Mail/Calendar/linked-call v1 events use the shared pure
`platform/events.EnvelopeV1`. Producers pass the complete envelope through the existing
outbox as `event_type == envelope.subject`; the worker publishes the stored JSON unchanged.
This is additive: legacy outbox payloads are not globally wrapped. Envelope `event_id` is
the application dedupe identity, while the outbox-row `Hub-Event-Id` header is a distinct
transport diagnostic id. New consumers reject broker/JSON subject mismatch or unknown
schema before mutation.

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
| Versioned event value | `platform/events` (planned Mail 0.1.5) | Deterministic full envelope and application-event identity without changing legacy payloads |
| Outbox | `platform/outbox` | At-least-once raw stored-payload delivery; row id remains transport identity |
| Optimistic Locking | Ent `version` field | Concurrent edits |
| Soft Delete | Trash 30 days | Recovery, GDPR path |
| Presigned URLs | `storage/` | Direct upload, no backend buffer |
| Repository | `*/repository.go` | Testability, layering |
| Cron helper | `platform/cron` | ME-10 graceful shutdown для tickers |

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
| v4.0 | `intranet/`, calendar aggregate, `crm/`, `hr/`, `accounting/`, `bases/`, `marketplace/`, `status/`, Resend alert `mail/` | `/intranet`, `/calendar`, `/crm`, `/hr`, `/accounting`, `/bases`, `/marketplace`, `/status`; mailbox client is not implemented |
| v4.5 | `integrations/onec/` | settings integrations |

Event bus: **NATS JetStream** (без Redis pub/sub). Search: PostgreSQL + **Meilisearch** indexer (`cmd/meilisearch-indexer`).

## Связанные документы

- `docs/system_specification.md` — API, модели данных
- `docs/security_guidelines.md` — OWASP, уязвимости
- `docs/stack_and_tools.md` — версии стека
- `.planning/decisions/ADR-MAIL-01-mail-engine-calendar-sync.md` — замороженная граница mail engine и Calendar RSVP
- `.planning/decisions/specs/MAIL-SPEC.md` — ERD, state machines, API/event/security contract
- `docs/mail/QA-HANDOFF.md` — статус и будущие verification gates (`NOT_RUN`)
- `.planning/decisions/specs/PLATFORM-VISION-v4.md` — corporate super-app vision
- `.planning/decisions/specs/AI-PROVIDER-HUB-SPEC.md` — AI gateway spec

---
*Последнее обновление: 2026-06-22 — RU-FIX-01: RU section headers*
