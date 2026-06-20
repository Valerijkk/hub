# Hub — Stack and Tools

> Единая база знаний. Обновляется ARCH при изменении стека.
> Детали: `README.md`, ограничения: `.cursor/rules/60-tz-constraints.mdc`

---

## Языки и рантаймы

| Компонент | Версия | Назначение |
|-----------|--------|------------|
| Go | **1.25** (`backend/go.mod`) | Backend, CLI |
| TypeScript | **5.7** (`frontend/package.json`) | Frontend |
| Node.js | **22** (CI), 20+ (local) | Frontend build, tooling |

## Backend

| Библиотека | Версия (факт) | Назначение |
|------------|---------------|------------|
| Fiber | **v3.0.0-beta.3** | HTTP framework, WebSocket |
| Ent | **v0.14.1** | ORM, кодогенерация схем |
| Atlas | indirect via Ent | SQL миграции |
| testify | **v1.11.1** | Unit/integration assertions |
| testcontainers | dev/test | Integration tests с Docker |
| aws-sdk-go-v2 | **v1.32.7** (+ s3 v1.71.1) | SeaweedFS S3 API |
| nats.go | **v1.48.0** (client) | NATS JetStream client |
| go-redis/v9 | **v9.17.2** | Redis client |
| mongo-driver | **v2.0.0** | MongoDB client |
| livekit/server-sdk-go | **v2.4.0** | WebRTC tokens, rooms |
| telego | **v0.31.0** | Telegram Bot API |
| slog | stdlib | Structured logging |
| swag | **v1.16.6** | OpenAPI annotations (HTML docs via `dev/`) |

> `mockery`, `golangci-lint` — упоминаются в wiki/rules, **не в CI** (`.github/workflows/ci.yml` — только build).

## Frontend

| Библиотека | Версия (факт) | Назначение |
|------------|---------------|------------|
| React | **19.1** | UI framework |
| Vite | **6.x** | Build tool |
| TanStack Router | **1.x** | File-based routing (**47 routes** в `routeTree.gen.ts`) |
| TanStack Query | **5.x** | Server state |
| Zustand | **5.x** | Client state (**11 stores** в `src/stores/`) |
| Tailwind CSS | **4.x** | Styling |
| Shadcn/ui + Radix | latest pins | UI components |
| Plate (`@udecode/plate`) | **48.x** | Rich-text editor |
| pragmatic-drag-and-drop | **1.7.x** | Kanban DnD |
| Recharts | **3.x** | Dashboard charts (не Tremor) |
| i18next + react-i18next | **25.x / 16.x** | Локализация |
| motion | **12.x** | UI animations |
| Playwright | **1.x** | E2E (7 specs) |
| Vitest + RTL | **3.x / 16.x** | Unit tests |
| Biome | latest | Lint + format (`npm run lint`) |
| vite-plugin-pwa | **1.x** | PWA |

## Базы данных и хранилища

| Сервис | Версия (Docker) | Данные |
|--------|-----------------|--------|
| PostgreSQL | **17** | Users, projects, issues, TMS, workflow |
| MongoDB | **8.0** | Messages, Allure reports, ephemeral call chat |
| Redis | **7.4-alpine** | Sessions, refresh tokens, RBAC cache, presence |
| SeaweedFS | **3.97** (pinned) | Files (S3-compatible), recordings |
| NATS JetStream | **2.11** (server image) | Pub/sub, outbox, notifications |

## Real-time и медиа

| Сервис | Версия | Назначение |
|--------|--------|------------|
| LiveKit server | **v1.8** (`livekit/livekit-server:v1.8`) | Voice calls, screen share, recording |
| Fiber WS | built-in (fasthttp/websocket) | Client real-time events |

## Инфраструктура

| Инструмент | Окружение | Назначение | Статус в репо |
|------------|-----------|------------|---------------|
| Docker Compose | dev | Локальная разработка | ✅ `docker-compose.local.yml`, `deployments/*.yml` |
| K3s | prod (planned) | Kubernetes orchestration | ❌ манифестов нет (TD-04) |
| Caddy | prod/dev | Reverse proxy, auto HTTPS, WebSocket | ✅ `deployments/Caddyfile*` |
| GitHub Actions | CI/CD | Build, test, deploy, pr-guard, dependabot | ✅ `.github/workflows/` |
| GitLab CI | CI mirror | Same gates for GitLab remotes | ✅ `.gitlab-ci.yml` |
| Loki + Prometheus + Grafana | prod (planned) | Logs + metrics | ❌ не в compose (TD-05) |

## Внешние сервисы

| Сервис | Назначение | Ограничения |
|--------|------------|-------------|
| Resend | Magic Link email | 3000 писем/мес бесплатно |
| Telegram Bot API | Auth + notifications | telego SDK |
| GitHub / GitLab | Webhooks, PR/MR integration | API tokens encrypted at rest |

## Линтеры и форматеры

| Инструмент | Область | Команда | CI |
|------------|---------|---------|-----|
| go build, go generate | Backend | `cd backend && go generate ./ent && go build ./...` | ✅ |
| go vet / golangci-lint | Backend | — | ❌ не в CI |
| tsc | Frontend types | `cd frontend && npx tsc --noEmit` | ✅ (via `npm run build`) |
| Biome | Frontend lint | `cd frontend && npm run lint` | ✅ baseline (continue-on-error) |
| go test -race | Backend | `go test -race -count=1 -timeout=600s ./...` | ✅ |
| Playwright | E2E smoke + full | `npx playwright test` | ✅ `e2e-smoke`, `e2e-full` |
| pr-guard | Secret scan, labels | on PR | ✅ |
| dependabot | Weekly deps | go.mod, npm, actions | ✅ |
| Biome | Planning scripts | `.planning/biome.json` | — |

## CLI

| Утилита | Путь | Назначение |
|---------|------|------------|
| hub-cli | `cli/` | CI/CD: push-results, create-bug, update-task |

## Структура репозитория

```
hub/
├── backend/              # Go API, Ent schemas (~59), 20 domain modules + platform
├── frontend/             # React SPA (~47 routes, 12 Zustand stores incl. command-store)
├── cli/                  # hub-cli
├── deployments/          # Docker Compose (dev/prod/test), Caddy, LiveKit, SeaweedFS
├── docker-compose.local.yml  # Full local stack (root)
├── docs/                 # Единая база знаний (этот регламент)
├── docs/wiki/            # Детальная wiki (API, deploy, shortcuts)
└── .planning/            # GSD roadmap, phases, audits
```

## Обоснование ключевых решений

| Решение | Альтернативы | Почему Hub |
|---------|--------------|------------|
| Go + Fiber | Python FastAPI | Горутины для real-time, единый WS стек |
| PostgreSQL + MongoDB | Только PG | Документная модель для чатов и Allure JSON |
| NATS JetStream | RabbitMQ, Redis pub/sub | Легче, Go-native, outbox pattern |
| Ent | GORM | Кодогенерация, граф сущностей, Privacy layer |
| LiveKit | Jitsi | Go-native, React SDK, запись из коробки |
| SeaweedFS | MinIO | MinIO archived Feb 2026, S3-compatible |
| TanStack Router | React Router | Type-safe, единая экосистема с Query |
| Caddy | Nginx | Auto HTTPS, WebSocket, 15 строк конфига |
| Recharts | Tremor | Фактически используется в dashboard-коде |

## Запрещённые замены (без согласования)

- Fiber → net/http, gorilla
- NATS → Redis pub/sub
- SeaweedFS SDK → minio-go
- TanStack Router → React Router
- Shadcn → Ant Design, MUI
- Caddy → nginx (prod)
- Docker Compose (dev) / K3s (prod) → другая оркестрация

---
*Последнее обновление: 2026-06-21 — GitLab CI, pr-guard, dependabot, Playwright e2e-full*
