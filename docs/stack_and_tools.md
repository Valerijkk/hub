# Hub — Стек и инструменты

> Единая база знаний. Обновляется ARCH при изменении стека.  
> Ограничения TZ: `docs/tz_constraints.md` · Детали деплоя: `docs/wiki/DEPLOY.md`

---

## Языки и рантаймы

| Компонент | Версия (факт) | Источник | Назначение |
|-----------|---------------|----------|------------|
| Go | **1.25.0** | `backend/go.mod` | Backend, CLI, hub-mail |
| TypeScript | **~5.7** | `frontend/package.json` | Frontend |
| Node.js | **22** (QA image), 20+ (local) | `testframework/docker/Dockerfile.qa`, `frontend/package.json` | Frontend build and Vitest inside Docker |
| Python | **3.12+ target** (локально bundled 3.12.13) | `testframework/pyproject.toml`, `testframework/uv.lock`, `testframework/PACKAGE-CONTRACT.md` | Package install и 181/181 framework selftests Docker-verified; API/integration/e2e/load остаются `NOT_RUN` и вне minimal merge gate |

---

## Backend

| Библиотека | Версия (go.mod) | Назначение |
|------------|-----------------|------------|
| Fiber | **v3.0.0-beta.3** | HTTP, WebSocket (fasthttp) |
| Ent | **v0.14.1** | ORM, кодогенерация схем (**87 schemas**, `backend/ent/schema/`, measured 2026-08-07) |
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

Production build выполняется только внутри принятого resource-limited Docker
job. Его внутренние шаги — Ent generation и Go release build; этот раздел не
разрешает host-команду или обход versioned launcher/CI contract.

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
| Zustand | **^5.0.0** | Client state (**12** stores в `frontend/src/stores/`, measured 2026-07-28) |
| Tailwind CSS | **^4.0.0** | Styling (`@tailwindcss/vite`) |
| Shadcn/ui + Radix | pinned | UI primitives |
| Plate (`@udecode/plate`) | **^48.0.5** | Rich-text |
| pragmatic-drag-and-drop | **^1.7.9** | Kanban DnD |
| Recharts | **^3.8.0** | Dashboard charts |
| i18next | **^25.9.0** | Локализация RU/EN |
| motion | **^12.40.0** | UI animations |
| livekit-client | **^2.17.3** | WebRTC UI |
| Vitest | **^3.0.0** | Unit tests |
| @testing-library/react | **^16.0.0** | Component tests |
| @playwright/test | **^1.52.0** | E2E |
| @biomejs/biome | latest | Lint + format |
| vite-plugin-pwa | **^1.3.0** | PWA |

### Скрипты npm

Таблица целиком — current legacy inventory, не инструкция для агента. Host
Vite/build/lint/test процессы запрещены; target checks будут вызываться только
через root Compose QA API: `docker compose -f docker-compose.local.yml run --rm <exact-service>`.

| Команда | Действие |
|---------|----------|
| `npm run dev` | Vite dev :5173 |
| `npm run build` | `tsc -b && vite build` |
| `npm run lint` | `biome check ./src` |
| `npx vitest run` / `npm run test:run` | Legacy unit entrypoint до Phase 129; target: `frontend/tests/**` |
| `npm run test:e2e:smoke` | Legacy TypeScript Playwright entrypoint; delete в TASK-219 |
| `npm run test:e2e` | Legacy TypeScript Playwright entrypoint; delete в TASK-219 |

Эти scripts не запускаются агентами на host: после binding и независимой
все target checks выполняются только root Compose exact services.
TypeScript
Playwright entrypoints не входят в target и удаляются exact-manifest batch в
TASK-219 без parity/dual-run. Greenfield Python E2E получает только launcher
profiles по Phase 137–141, не отдельные Docker entrypoints.

---

## Базы данных и хранилища (Docker)

| Сервис | Образ (compose) | Порты | Данные |
|--------|-----------------|-------|--------|
| PostgreSQL | `postgres:17` | 5432 | Users, projects, issues, TMS, workflow |
| MongoDB | `mongo:8.0` | 27017 | Messages, Allure reports, call chat |
| Redis | `redis:7.4-alpine` | 6379 | Refresh tokens, RBAC cache, presence |
| SeaweedFS | `chrislusf/seaweedfs:3.97` | 9333, 8333 | Files (S3), recordings |
| NATS JetStream | `nats:2.11` | 4222, 8222 | Pub/sub, outbox, notifications |

Для QA используется только root `docker-compose.local.yml`: `docker compose -f
docker-compose.local.yml run --rm <exact-service>`. Восемь QA services profile-explicit
и disposable; integration/e2e используют изолированную инфраструктуру.
Исторические legacy QA entrypoints архивны. Единственный интерфейс запуска:
`docker compose -f docker-compose.local.yml run --rm <service>`.

---

## Real-time и медиа

| Сервис | Образ | Назначение |
|--------|-------|------------|
| LiveKit | `livekit/livekit-server:v1.8` | Voice, screen share, recording (egress) |
| Fiber WS | built-in | Client events (`GET /ws` + ticket subprotocol) |

Prod LiveKit: конфиг из env heredoc в `docker-compose.prod.yml` (не монтировать `livekit-prod.yaml` с секретами).

---

## Инфраструктура

| Инструмент | Окружение | Файлы | Статус |
|------------|-----------|-------|--------|
| Docker Compose | dev / prod | `docker-compose.local.yml`, `deployments/*.yml` | ✅ |
| Caddy | prod/dev | `deployments/Caddyfile*` | ✅ reverse proxy, WS, HTTPS |
| K3s | prod (planned) | — | ⚠️ частично: `deployments/k3s/monitoring/networkpolicy-metrics.yaml`; полного deployment-набора нет |
| GitHub Actions | manual only | `.github/workflows/ci.yml`, `deploy.yml` | automatic CI/PR triggers disabled; local Docker is canonical |
| GitLab CI | disabled | `.gitlab-ci.yml` | pipelines disabled |
| Monitoring | optional | `deployments/docker-compose.monitoring.yml` | Prometheus, Loki, Grafana (pinned); Tempo via `--profile otel` (A14-009) |

### Caddy routing (prod)

```
/api/*  → backend:3000
/ws     → backend:3000 (HTTP/1.1 upgrade)
/lk/*   → livekit:7880 (strip prefix)
/*      → frontend:80 (static)
```

---

## QA and deploy (local Docker canonical)

GitHub CI и PR Guard доступны только вручную: автоматические запуски на
`push`/`pull_request` отключены, чтобы не расходовать GitHub Actions minutes.
GitLab pipelines отключены. Каноническая проверка выполняется только через
task-owned root Compose; host test/build/validator команды не являются
execution authority.

**Deploy** (`.github/workflows/deploy.yml`): отдельный manual-only
`workflow_dispatch` на текущем `master` после документированного локального
Docker evidence, `FA-06/PG-03`, CTO sign-off и `confirm_production`. Обычный
merge не деплоит; если `master` продвинулся после dispatch, SSH job завершится
fail-closed до изменения checkout:

1. Build & push GHCR images (SHA tag)
2. SSH deploy: `docker compose up` backend, livekit, frontend, caddy
3. Post-deploy health curl
4. **Без** `DEV_MODE=true` в workflow

**Current QA authority:** root Compose `run --rm <exact-service>`; this document
does not claim a session result. QA services are disposable and integration/e2e
infrastructure is isolated. Legacy jobs and scripts remain archival inventory only.

---

## Внешние сервисы

| Сервис | Назначение | Ограничения |
|--------|------------|-------------|
| Resend / hub-mail | Magic Link email, transactional | Rate limits провайдера |
| Telegram Bot API | Уведомления (не login) | telego SDK |
| GitHub / GitLab | Webhooks, PR/MR | Tokens encrypted at rest |

---

## Линтеры и quality gates

Команды старого CI ниже — только инвентарный снимок, а не разрешённый fallback.
Current authority is root Compose `run --rm <exact-service>`; host checks forbidden.

| Целевой job | Область | Исполнение | Статус |
|-------------|---------|------------|--------|
| `backend-unit` | genuine Go units + backend static gates | root Compose exact service | PASS: build/vet/unit/race |
| `frontend-unit` | genuine Vitest/RTL units + TypeScript | root Compose exact service | PASS: 338/338 suites, 419/419 tests |
| `python-contract` | install/package/framework contracts | root Compose exact service | PASS: 181/181 framework selftests; SUT product tests removed |
| `python-api` | future/public API behavior | outside minimal merge gate | `NOT_RUN` |
| `python-integration` | future boundary behavior | outside minimal merge gate | `NOT_RUN` |
| `python-e2e-smoke` / `python-e2e-full` | future journeys | outside minimal merge gate | `NOT_RUN` |
| `python-load` | future load profiles | outside minimal merge gate | `NOT_RUN` |
| dependabot | weekly go/npm/actions | вне test service | не настроен |

---

## CLI

| Утилита | Путь | Назначение |
|---------|------|------------|
| hub-cli | `cli/` | CI/CD: push-results, create-bug, update-task |
| hub-mail | `backend/cmd/hub-mail` | Transactional email microservice |
| meilisearch-indexer | `backend/cmd/meilisearch-indexer` | Search indexing (optional) |

---

## Структура репозитория

Текущая структура до cutover:

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

Цель фаз 128–141:

```text
hub/
├── backend/tests/<domain>/ # только Go unit tests
├── frontend/tests/<domain>/ # только Vitest/RTL unit tests
├── frontend/tests/support/ # frontend setup/render/fixtures
└── testframework/          # один Python QA-проект, без python/ nesting
    ├── PACKAGE-CONTRACT.md   # package/bootstrap contract
    ├── src/hub_testkit/      # reusable installed framework package
    │   └── py.typed          # typed-package marker
    ├── tests/
    │   ├── {contract,api,integration}/ # existing pytest source suites
    │   └── e2e/              # absent; not implemented yet
    ├── selftests/            # framework contracts only
    ├── cases/                # TC intent + risk links
    ├── schemas/              # closed case/run contracts
    ├── policies/             # marker/environment/redaction/profile policy
    ├── load/                 # README-only; executable Locust profile absent
    ├── docker/               # QA image/config source for root Compose
    ├── reports/<task-run>/   # owner-marked generated evidence
    ├── pyproject.toml        # package/tool authority
    └── uv.lock               # exact dependency resolution
```

Structure/boundary scaffold в commit `22035c45` уже существует. Root Compose
service definitions, Python QA image and package files are present.
`backend-unit`, `frontend-unit` and framework-only `python-contract` have fresh
Docker PASS evidence for the minimal slice. API/integration/e2e/load are
`NOT_RUN` and are not merge gates for this slice. This does not claim product
or production readiness.
Reusable imports и suite roots физически разделены; direct root
`core/api/integration/contract/e2e` packages не являются target.

Live ledger на pinned revision `b9f690da`: **1,049** rows / **4,246** declared;
**832** Go / **3,674** tests (legacy root 831/3,672 + server 1/2;
595 genuine unit + 237 component: 31 API + 206 integration), **180** Vitest /
**519** cases (156 unit + 24 static contract) + 3 frontend support,
**20** E2E specs / **53** cases + 14 E2E support, hook sources 43 files /
582 symbols / 892 links и 0 target Python tests. SHA-256: `83b6a311…`.

В `backend/tests/**` не переносятся component/integration tests: текущие
**237** `go-component` rows (31 API + 206 integration) удаляются по exact
manifest без one-to-one port. Новый Python `api/integration/contract` coverage
проектируется от current requirements/risks/verified TC. Из 595 genuine units
350 переносятся сразу; 245 hook-dependent units защищены TASK-234 до
refactor/move. После Phase 141
в `testframework/**` допустим только Python для исполняемого test/load кода;
legacy Go/Vitest/TS E2E и их entrypoints удаляются.

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
*Последнее обновление: 2026-08-09 — package/suite roots, structure-only scaffold и service/docker ownership.*
