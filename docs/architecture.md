# Hub — Architecture

> Единая база знаний. Обновляется ARCH при изменении архитектуры.
> Детали: `docs/TECHNICAL.md`, `docs/wiki/ARCHITECTURE.md`
> Аудит кода: `.planning/code-review-final/ARCHITECTURE.md` (2026-05-15)

---

## Архитектурный стиль

**Модульный монолит** на Go с React SPA. Clean Architecture внутри каждого модуля: Handler → Service → Repository → DB.

| Принцип | Реализация в Hub |
|---------|------------------|
| Clean Architecture | Слои handler/service/repository, DI через интерфейсы (service→repo) |
| Event-Driven | NATS JetStream + Outbox pattern для domain events |
| CQRS (частичный) | Read models для search (PG tsvector + Mongo text index) |
| Repository Pattern | `repository.go` interface + `ent_repository.go` — **11/20 domain-модулей** (v2.6 rollout) |

**Architecture score (2026-06-17): 75 / 100** — цельный ацикличный монолит; основной долг — незавершённый Repository rollout и несколько layering bypass'ов.

## C4 Level 2 — Container Diagram

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

## C4 Level 3 — Backend Modules

**20 domain-модулей** + **`platform/`** (инфраструктура). Точка композиции: `backend/cmd/server/main.go` (~630 LOC).

```
backend/internal/
├── apigateway/     API Key auth, webhook signatures
├── audit/          Audit log, trash, GDPR, retention cron
├── auth/           Magic Link, Telegram, JWT, sessions
├── backup/         PG + Mongo + S3 ZIP backups
├── calls/          LiveKit rooms, recordings, scheduled
├── dashboard/      Personal/project/admin dashboards
├── dev/            DEV_MODE seed, dev-login, API docs HTML (!release)
├── github/         Webhooks, PR linking
├── gitlab/         Webhooks, MR linking
├── imex/           Import/export CSV/JSON/PDF
├── messenger/      Channels, messages, threads, presence
├── notification/   In-app, Telegram, Web Push
├── project/        Projects, members, invites, teams, templates
├── rbac/           Custom roles, permission bits
├── search/         Global search, JQL, saved filters
├── sprint/         Sprints, burndown, planning poker
├── storage/        SeaweedFS presigned URLs, orphan cleanup
├── tms/            Test cases, runs, Allure parser
├── tracker/        Issues, workflow, boards, labels
├── ws/             WebSocket Hub, NATS Bridge, room Authorizer
└── platform/       config, cron, crypto, database, errors, http,
                    middleware, nats, outbox, redis, telegram
```

**Repository pattern (11 модулей):** audit, calls, github, gitlab, messenger, notification, project, search, sprint, tms, tracker.

**Без Repository (7 модулей, `*ent.Client` в service/apigateway):** auth, dashboard, backup, imex, storage, apigateway, rbac.

**Ent schemas:** 59 сущностей в `backend/ent/schema/` (+ `helpers.go`).

## Потоки данных

### HTTP Request
```
Browser → Caddy → Fiber Middleware → Handler → Service → Repository → DB
                  (RequestID, Recover, SecurityHeaders, Logger, CORS, RateLimit)
                  api group: APIKeyAuth, per-route auth/RBAC limits
```

Фактический порядок middleware (`main.go:185-250`):
1. `middleware.RequestID()`
2. `fiberrecover.New()` (не `middleware.Recover`)
3. `middleware.SecurityHeaders()` — NF-08 (nosniff, DENY, CSP)
4. `middleware.Logger()`, `middleware.CORS()`
5. `middleware.RateLimit()` — 100 req/sec
6. Per-route: `/files` 10/min, `/messenger` 60/min, `/auth` 3/10min (Redis, off in DevMode)
7. `apigatewaymod.APIKeyAuth()` — на `api` group **до** `RegisterRoutes` (SEC-01 fix)

### Real-time Event
```
Domain Event → Outbox (PG) → Worker → NATS → WS Bridge → Client
```

### Notification
```
NATS Event → JetStream Consumer → NotificationService → In-App + Telegram + Push
```

### File Upload
```
Client → POST presign URL → Backend (RBAC check) → SeaweedFS PUT (direct)
```

## Схема развёртывания

| Окружение | Compose-файл | Proxy | Особенности |
|-----------|--------------|-------|-------------|
| Development (full stack) | `docker-compose.local.yml` (корень репо) | Caddy :8080 | DEV_MODE, hot reload |
| Development (infra only) | `deployments/docker-compose.yml` | direct | PG, Mongo, Redis, NATS, S3, LiveKit |
| Production | `deployments/docker-compose.prod.yml` | Caddy | Replicas, secrets via env |
| CI tests | `deployments/docker-compose.test.yml` | — | testcontainers-compatible |

**K3s:** задокументирован в README/roadmap как prod-оркестрация, **манифестов в репо нет** (TD-04, v2.0 DevOps backlog).

**Observability (Loki, Prometheus, Grafana):** в roadmap v2.0, **не включены** в `deployments/docker-compose.prod.yml` (TD-05).

Компоненты в prod compose: backend, frontend (static), postgres, mongodb, redis, nats, seaweedfs, livekit, caddy.

## Безопасность на уровне архитектуры

| Слой | Меры |
|------|------|
| Edge | Caddy TLS, rate limiting |
| Auth | JWT 15min + Redis refresh, Magic Link rate limit (Redis sliding window) |
| AuthZ | RBAC 3 уровня, Redis cache с инвалидацией |
| Data | User-scope в search, channel membership в messenger |
| Network | SSRF protection (`safehttp`), no `http.DefaultClient` |
| Secrets | Env vars only, AES-GCM at rest для tokens |
| WS | Origin check + `ws.Authorizer` (channel/project/call membership, `main.go:575-605`) |

## Ключевые паттерны

| Паттерн | Где | Зачем |
|---------|-----|-------|
| Outbox | `platform/outbox` | At-least-once delivery events |
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
| 4 | RBAC middleware принимает `*ent.Client` (layering leak) | `backend/internal/rbac/middleware.go:46,114` + `main.go:540` |
| 5 | Dev handler обходит Service (`//go:build !release`) | `backend/internal/dev/handler.go:102` |

## Технический долг (известный)

| ID | Проблема | Evidence | План |
|----|----------|----------|------|
| TD-01 | DI-точка раздута | `cmd/server/main.go` — 630 LOC; cron вынесен в `platform/cron` | Wave extraction v2.6 |
| TD-02 | `*ent.Client` в 7 domain-пакетах | `auth/service*.go`, `dashboard/service.go:34`, `backup/service.go:23`, `imex/service.go:23`, `storage/service.go:36`, `apigateway/*.go:40+`, `rbac/service.go:19` | Repository Phase 19 |
| TD-03 | Escape hatch `GetDB()` | `tracker/repository.go:153`, `handler_issues.go:269` | Убрать, JQL search → service |
| TD-04 | K3s манифесты отсутствуют | `deployments/` — только Compose + Caddy + LiveKit | v2.0 DevOps |
| TD-05 | Observability не развёрнута | `docker-compose.prod.yml` — нет Loki/Prometheus/Grafana | v2.0 Phase |
| TD-06 | ~~Security headers~~ | **FIXED** `main.go:190` + `security_headers.go` | — |
| TD-07 | Handler→Service без интерфейсов | все `handler.go` — `*XxxService` concrete type | v2.6 (низкий приоритет) |
| TD-08 | `audit/retention.go` вне Repository | `retention.go:41` — отдельный `RetentionService` на `*ent.Client` | Phase 19 |
| TD-09 | `notification/push.go` на Ent | `push.go:45` — `PushSender` держит `*ent.Client` | Phase 19 |
| TD-10 | CRITICAL security (v2.0.1) | 10/12 FIXED в коде (2026-06-17); SEC-06 multi-instance + secret rotation pending | v2.0.1 hotfix |

### Repository coverage detail

| С Repository | Без Repository |
|--------------|----------------|
| audit, calls, github, gitlab, messenger, notification, project, search, sprint, tms, tracker | auth, dashboard, backup, imex, storage, apigateway, rbac |

**Покрытие:** 11 / 18 service-пакетов = **61%** (ws/dev/platform — не DB-service модули).

## Связанные документы

- `docs/system_specification.md` — API, модели данных
- `docs/security_guidelines.md` — OWASP, уязвимости
- `docs/stack_and_tools.md` — версии стека

---
*Последнее обновление: 2026-06-17 — ARCH audit vs codebase*
