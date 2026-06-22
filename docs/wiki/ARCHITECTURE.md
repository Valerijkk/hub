# Архитектура Hub

> Wiki-справочник. Полная версия: [`docs/architecture.md`](../architecture.md) · Спецификация: [`docs/system_specification.md`](../system_specification.md)

*Последнее обновление: 2026-06-21*

---

## Обзор

Hub — **модульный монолит**: Go backend (Fiber v3) + React 19 SPA. Модули изолированы пакетами `backend/internal/<module>/` с паттерном Handler → Service → Repository. Межмодульная асинхронная связь — через **NATS JetStream** и Outbox; realtime — **WebSocket** (`/ws`).

**Один бинарь** `cmd/server` монтирует все domain-модули (v1 core + v3 AI + v4 super-app). Дополнительные процессы: `hub-mail`, `meilisearch-indexer`.

---

## Контейнерная диаграмма

```
┌─────────────┐     HTTPS/WSS      ┌─────────────┐
│   Browser   │ ◄────────────────► │    Caddy    │
│  React SPA  │                    │   :80/443   │
└─────────────┘                    └──────┬──────┘
                                          │
                    ┌─────────────────────┼─────────────────────┐
                    │                     │                     │
              /api/*│               /ws   │              static │
                    ▼                     ▼                     ▼
            ┌───────────────┐     ┌───────────────┐     ┌───────────────┐
            │ Go Backend    │     │ WS Handler    │     │ Frontend      │
            │ Fiber v3      │     │ (Fiber)       │     │ Vite build    │
            │ :3000         │     │ :3000         │     │               │
            └───────┬───────┘     └───────┬───────┘     └───────────────┘
                    │                     │
    ┌───────────────┼─────────────────────┼───────────────┐
    ▼               ▼                     ▼               ▼
┌────────┐   ┌──────────┐   ┌────────┐   ┌────────┐   ┌──────────┐
│ PG 17  │   │ MongoDB 8│   │ Redis 7│   │ NATS   │   │ SeaweedFS│
│ Ent    │   │ messages │   │ session│   │ JS     │   │ S3       │
└────────┘   └──────────┘   └────────┘   └────────┘   └──────────┘
                                              │
                                        ┌─────┴─────┐
                                        │  LiveKit  │
                                        │  WebRTC   │
                                        └───────────┘
```

> **Важно:** reverse proxy — **Caddy**, не Nginx (`deployments/Caddyfile`, `Caddyfile.prod`).

---

## Backend-модули (31 domain + platform)

| Группа | Пакеты | Назначение |
|--------|--------|------------|
| **Core v1** | auth, project, tracker, sprint, messenger, tms, calls, dashboard, notification, search, ws | Трекер, чат, TMS, звонки |
| **Integrations** | github, gitlab, apigateway, imex, backup, storage | VCS, API keys, import/export, файлы |
| **Platform** | audit, rbac, dev | Audit, RBAC, DEV_MODE |
| **v3 AI** | ai | Провайдеры, @hub-ai bot, JQL assist |
| **v4 Super-app** | intranet, calendar, crm, hr, accounting, bases, marketplace, status, mail | Корпоративные модули |
| **v4.5** | integrations/onec | Коннектор 1C |
| **Infra** | platform/ | config, middleware, nats, outbox, observability, redis |

Структура каждого domain-модуля:

```
backend/internal/<module>/
├── handler.go          # HTTP
├── service.go          # Business logic
├── repository.go       # Interface (где есть)
├── ent_repository.go   # Ent implementation
└── dto.go              # Request/Response
```

**Ent schemas:** 82 сущности PostgreSQL (`backend/ent/schema/`).

---

## Frontend

| Слой | Технология |
|------|------------|
| Router | TanStack Router (~60+ страниц, `frontend/src/routes/`) |
| Server state | TanStack Query |
| Client state | Zustand |
| UI | Shadcn/ui + Tailwind CSS 4 |
| Rich text | Plate editor |
| Realtime | WebSocket client |

**Маршруты v4 (верифицировано):** `/intranet`, `/calendar`, `/crm`, `/hr`, `/accounting`, `/bases`, `/marketplace`, `/status`, `/settings/ai`, `/admin/mail`.

---

## Потоки данных

### HTTP-запрос

```
Browser → Caddy → Fiber Middleware → Handler → Service → Repository → DB
                  RequestID · Recover · SecurityHeaders · Logger · CORS
                  RateLimit · APIKeyAuth (на /api group)
```

Per-route limits: auth 3/10min · API 100/sec · files 10/min · messages 60/min.

### Real-time

```
Domain Event → Outbox (PG) → Worker → NATS JetStream → WS Bridge → Client
```

WebSocket rooms: `project:<uuid>`, `channel:<uuid>`, `user:<uuid>`, `admin:<uuid>`. Подписка только после RBAC/membership check (NF-02).

### Уведомления

```
NATS Event → JetStream Consumer → NotificationService
    → In-App (WS) + Telegram (bind only) + Web Push (VAPID)
```

### Загрузка файлов

```
Client → POST /storage/presign → Backend (RBAC) → Presigned PUT → SeaweedFS
```

Max 25 MB через Fiber BodyLimit; presigned PUT — direct to S3.

### Поиск

```
Indexer (cmd/meilisearch-indexer) ← PG/Mongo changes
User Ctrl+K → GET /api/search → user-scope filter (NF-03)
```

---

## Хранилища данных

| БД | Данные | Зачем |
|----|--------|-------|
| **PostgreSQL 17** | Users, projects, issues, TMS, v4 entities, outbox | Реляционная модель, Ent ORM, ACID |
| **MongoDB 8** | Chat messages, Allure reports, call chat | Документы, text index, high write volume |
| **Redis 7** | Refresh tokens (hash), magic link codes, RBAC cache, presence, rate limits | TTL, fast lookup |
| **SeaweedFS** | Attachments, recordings, voice messages | S3-compatible object storage |
| **Meilisearch** | Search index (optional sidecar) | Full-text, user-scope queries |

---

## Аутентификация и RBAC

### Auth (Magic Link only)

```
POST /auth/magic-link        → код на email (Redis TTL 5 min)
POST /auth/magic-link/verify → JWT access (15 min) + refresh (Redis 30 d)
POST /auth/refresh           → новый access token
```

Telegram — **только** bind для notify: `POST /notifications/telegram/bind`. Login через Telegram **удалён** (ADR-AUTH-01).

### RBAC — три уровня

1. **Системные роли:** founder (1/org), admin, member
2. **Проектные:** Owner, Developer, Tester, Viewer
3. **Кастомные:** bitmask permissions (`uint64`)

Кеш прав в Redis; инвалидация при изменении membership.

---

## Развёртывание

| Окружение | Файл |
|-----------|------|
| Dev full stack | `docker-compose.local.yml` (корень) |
| Dev infra | `deployments/docker-compose.yml` |
| Production | `deployments/docker-compose.prod.yml` |
| Monitoring | `deployments/docker-compose.monitoring.yml` |
| CI tests | `deployments/docker-compose.test.yml` |

Observability configs: `deployments/monitoring/` (Prometheus, Loki, Grafana, Alertmanager). K3s: `deployments/k3s/` (partial).

---

## Связанные документы

- [`docs/architecture.md`](../architecture.md) — C4, tech debt, patterns
- [`docs/wiki/API.md`](API.md) — REST endpoint reference
- [`docs/wiki/DEPLOY.md`](DEPLOY.md) — деплой и env vars
- [`docs/security_guidelines.md`](../security_guidelines.md) — OWASP, blockers
- [`.planning/STATE.md`](../../.planning/STATE.md) — текущий статус проекта
