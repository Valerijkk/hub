# Hub — Стек и инструменты

> Единая база знаний. Обновляется ARCH при изменении стека.  
> Ограничения TZ: `docs/tz_constraints.md` · Детали деплоя: `docs/wiki/DEPLOY.md`

---

## Языки и рантаймы

| Компонент | Версия (факт) | Источник | Назначение |
|-----------|---------------|----------|------------|
| Go | **1.25.0** | `backend/go.mod` | Backend, CLI, hub-mail |
| TypeScript | **~5.7** | `frontend/package.json` | Frontend |
| Node.js | **22** (CI), 20+ (local) | `.github/workflows/ci.yml` | Сборка, Vitest, Playwright |

---

## Backend

| Библиотека | Версия (go.mod) | Назначение |
|------------|-----------------|------------|
| Fiber | **v3.0.0-beta.3** | HTTP, WebSocket (fasthttp) |
| Ent | **v0.14.1** | ORM, кодогенерация схем (~59 schemas) |
| Atlas | indirect (Ent) | SQL-миграции |
| testify | **v1.11.1** | Unit/integration assertions |
| aws-sdk-go-v2 / s3 | **v1.32.7** / **v1.71.1** | SeaweedFS S3 API |
| nats.go | **v1.48.0** | NATS JetStream client |
| go-redis/v9 | **v9.17.2** | Redis (sessions, refresh, cache) |
| mongo-driver | **v2.0.0** | MongoDB (messages, Allure) |
| livekit/server-sdk-go | **v2.4.0** | WebRTC tokens, rooms, egress |
| telego | **v0.31.0** | Telegram Bot API (notify only) |
| golang-jwt/jwt | **v5.3.0** | JWT HS256 |
| swag | **v1.16.6** | OpenAPI (`/docs` в dev) |
| OTel | **v1.40.0** | Tracing (observability stack) |
| prometheus/client_golang | **v1.23.2** | Metrics |

**Не в CI:** `golangci-lint`, `mockery` — упоминаются в wiki, но `.github/workflows/ci.yml` использует `go vet` + `go test -race`.

### Сборка production

```bash
cd backend
go generate ./ent
go build -tags release -o bin/server ./cmd/server
```

Тег `release` исключает `/dev/*` routes (ME-09).

---

## Frontend

| Библиотека | Версия (package.json) | Назначение |
|------------|----------------------|------------|
| React / react-dom | **^19.1.0** | UI |
| Vite | **^6.0.0** | Dev server, build |
| TypeScript | **~5.7.0** | Типизация |
| TanStack Router | **^1.0.0** | File-based routing |
| TanStack Query | **^5.0.0** | Server state |
| Zustand | **^5.0.0** | Client state (11+ stores) |
| Tailwind CSS | **^4.0.0** | Styling (`@tailwindcss/vite`) |
| Shadcn/ui + Radix | pinned | UI primitives |
| Plate (`@udecode/plate`) | **^48.0.5** | Rich-text |
| pragmatic-drag-and-drop | **^1.7.9** | Kanban DnD |
| Recharts | **^3.8.0** | Dashboard charts |
| i18next | **^25.9.0** | Локализация RU/EN |
| motion | **^12.38.0** | UI animations |
| livekit-client | **^2.17.3** | WebRTC UI |
| Vitest | **^3.0.0** | Unit tests |
| @testing-library/react | **^16.0.0** | Component tests |
| @playwright/test | **^1.52.0** | E2E |
| @biomejs/biome | latest | Lint + format |
| vite-plugin-pwa | **^1.2.0** | PWA |

### Скрипты npm

| Команда | Действие |
|---------|----------|
| `npm run dev` | Vite dev :5173 |
| `npm run build` | `tsc -b && vite build` |
| `npm run lint` | `biome check ./src` |
| `npx vitest run` / `npm run test:run` | Unit tests (`testframework/stages/unit/frontend-test/domains/`) |
| `npm run test:e2e:smoke` | Playwright smoke job (`--project=chromium-smoke`) |
| `npm run test:e2e` | Playwright full job (`--project=chromium-full`) |

---

## Базы данных и хранилища (Docker)

| Сервис | Образ (compose) | Порты | Данные |
|--------|-----------------|-------|--------|
| PostgreSQL | `postgres:17` | 5432 | Users, projects, issues, TMS, workflow |
| MongoDB | `mongo:8.0` | 27017 | Messages, Allure reports, call chat |
| Redis | `redis:7.4-alpine` | 6379 | Refresh tokens, RBAC cache, presence |
| SeaweedFS | `chrislusf/seaweedfs:3.97` | 9333, 8333 | Files (S3), recordings |
| NATS JetStream | `nats:2.11` | 4222, 8222 | Pub/sub, outbox, notifications |

Файлы compose: `docker-compose.local.yml` (корень), `deployments/docker-compose.prod.yml`, `deployments/docker-compose.test.yml`.

---

## Real-time и медиа

| Сервис | Образ | Назначение |
|--------|-------|------------|
| LiveKit | `livekit/livekit-server:v1.8` | Voice, screen share, recording (egress) |
| Fiber WS | built-in | Client events (`GET /ws?token=JWT`) |

Prod LiveKit: конфиг из env heredoc в `docker-compose.prod.yml` (не монтировать `livekit-prod.yaml` с секретами).

---

## Инфраструктура

| Инструмент | Окружение | Файлы | Статус |
|------------|-----------|-------|--------|
| Docker Compose | dev / prod | `docker-compose.local.yml`, `deployments/*.yml` | ✅ |
| Caddy | prod/dev | `deployments/Caddyfile*` | ✅ reverse proxy, WS, HTTPS |
| K3s | prod (planned) | — | ❌ манифестов нет |
| GitHub Actions | CI/CD | `.github/workflows/ci.yml`, `deploy.yml` | ✅ |
| GitLab CI | mirror | `.gitlab-ci.yml` | ✅ |
| Monitoring | optional | `deployments/docker-compose.monitoring.yml` | Prometheus, Loki, Grafana (pinned v3.x) |

### Caddy routing (prod)

```
/api/*  → backend:3000
/ws     → backend:3000 (HTTP/1.1 upgrade)
/lk/*   → livekit:7880 (strip prefix)
/*      → frontend:80 (static)
```

---

## CI/CD (GitHub Actions)

Файл: `.github/workflows/ci.yml` · Ветка: `master` · Concurrency: cancel-in-progress.

| Job | Что делает | Timeout |
|-----|------------|---------|
| `backend-test` | `go generate ./ent`, `go vet -tags hubtest`, `go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...` | 25 min |
| `frontend-test` | `npm ci`, link testframework deps, `tsc --noEmit`, `vitest run`, `npm run lint` (Biome **gate**) | default |
| `build` | `go build -tags release`, `npm ci && npm run build` | default |
| `e2e-smoke` | Playwright frontend-only: login, a11y, critical-gate | после frontend-test |
| `e2e-full` | Docker infra + native backend (`DEV_MODE`) + 20 spec paths | 30 min |
| `coverage-gate` | `go test -coverpkg=./internal/...` ≥82% | master push only |
| `lighthouse` | Lighthouse login page artifact | non-blocking |

**Deploy** (`.github/workflows/deploy.yml`): только после **успешного CI** на push в `master` (`workflow_run`):

1. Build & push GHCR images (SHA tag)
2. SSH deploy: `docker compose up` backend, livekit, frontend, caddy
3. Post-deploy health curl
4. **Без** `DEV_MODE=true` в workflow

E2E config: `frontend/playwright.config.ts` · specs: `testframework/stages/e2e/domains/` · `frontend/scripts/link-testframework-node-modules.mjs` · Infra: `scripts/ci-e2e-backend.sh`.

---

## Внешние сервисы

| Сервис | Назначение | Ограничения |
|--------|------------|-------------|
| Resend / hub-mail | Magic Link email, transactional | Rate limits провайдера |
| Telegram Bot API | Уведомления (не login) | telego SDK |
| GitHub / GitLab | Webhooks, PR/MR | Tokens encrypted at rest |

---

## Линтеры и quality gates

| Инструмент | Область | Команда | CI |
|------------|---------|---------|-----|
| go generate + vet | Backend Ent | `go generate ./ent && go vet ./...` | ✅ |
| go test -race | Backend | `-tags hubtest`, Mongo service in CI | ✅ |
| go build -tags release | Prod binary | `go build -tags release ./cmd/server` | ✅ |
| tsc | Frontend types | `npx tsc --noEmit` | ✅ |
| Biome | Frontend lint | `npm run lint` | ✅ **gate** (не continue-on-error) |
| Vitest | Unit | `npx vitest run` | ✅ |
| Playwright | E2E | smoke + full jobs | ✅ |
| dependabot | Deps | weekly go/npm/actions | ✅ |

---

## CLI

| Утилита | Путь | Назначение |
|---------|------|------------|
| hub-cli | `cli/` | CI/CD: push-results, create-bug, update-task |
| hub-mail | `backend/cmd/hub-mail` | Transactional email microservice |
| meilisearch-indexer | `backend/cmd/meilisearch-indexer` | Search indexing (optional) |

---

## Структура репозитория

```
hub/
├── backend/                 # Go API, Ent schemas, internal/* modules
├── frontend/              # React SPA, playwright.config.ts
├── testframework/         # Pipeline-aligned tests (PIPELINE.md)
│   ├── pipeline/jobs/     # CI job manifests
│   └── stages/
│       ├── unit/frontend-test/domains/
│       ├── e2e/domains/ + e2e/shared/
│       └── quality/coverage-gate/
├── cli/                   # hub-cli
├── deployments/           # Compose prod/test/monitoring, Caddy, LiveKit
├── docker-compose.local.yml
├── docs/                  # База знаний (этот регламент)
├── docs/wiki/             # API, DEPLOY, QUICKSTART, SHORTCUTS
└── .planning/             # GSD roadmap, COVERAGE, STATE, AUDIT-*
```

---

## Обоснование ключевых решений

| Решение | Альтернативы | Почему Hub |
|---------|--------------|------------|
| Go + Fiber | FastAPI | Горутины, единый WS стек |
| PostgreSQL + MongoDB | Только PG | Документная модель для чатов и Allure |
| NATS JetStream | RabbitMQ, Redis pub/sub | Go-native, outbox |
| Ent | GORM | Кодогенерация, граф, Privacy |
| LiveKit | Jitsi | Go SDK, React components, egress |
| SeaweedFS | MinIO | S3-compatible, pinned 3.97 |
| TanStack Router | React Router | Type-safe + Query ecosystem |
| Caddy | Nginx | Auto HTTPS, WS в 15 строк |
| Biome | ESLint+Prettier | Один tool, быстрый CI gate |

## Запрещённые замены

См. `docs/tz_constraints.md`: Fiber, NATS, SeaweedFS SDK, TanStack Router, Shadcn, Caddy, Compose/K3s — без ADR.

---
*Последнее обновление: 2026-06-22 — quality audit: CI backend-test только testframework/, pipeline repo structure*
