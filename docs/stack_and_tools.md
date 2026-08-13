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
| Python | **3.12+ target** (локально bundled 3.12.13) | Phase 129 / `testframework/PACKAGE-CONTRACT.md` | Structure-only scaffold commit `22035c45`; package bootstrap, dependencies, executable suites и runtime ещё `NOT_READY`/`NOT_RUN` |

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
через versioned launcher API из `testframework/runner/**`, чей private source
candidate существует, но имеет только статический review и
`/qa-gate` является единственным outer consumer.

| Команда | Действие |
|---------|----------|
| `npm run dev` | Vite dev :5173 |
| `npm run build` | `tsc -b && vite build` |
| `npm run lint` | `biome check ./src` |
| `npx vitest run` / `npm run test:run` | Legacy unit entrypoint до Phase 129; target: `frontend/tests/**` |
| `npm run test:e2e:smoke` | Legacy TypeScript Playwright entrypoint; delete в TASK-219 |
| `npm run test:e2e` | Legacy TypeScript Playwright entrypoint; delete в TASK-219 |

Эти scripts не запускаются агентами на host: после binding и независимой
acceptance все target checks выполняются единственным launcher API из
`testframework/runner/**`.
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

Файлы compose перечислены только как inventory: `docker-compose.local.yml`
(корень), `deployments/docker-compose.prod.yml`,
`deployments/docker-compose.test.yml`. Target service profiles вызываются
только через launcher API из `testframework/runner/**`, который обязан
применять immutable execution assets из `testframework/docker/**`,
CPU/RAM/PID/time budgets и exact cleanup. Прямой Compose/`docker run` fallback
не разрешён; target launcher пока `NOT_READY`.

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
| GitHub Actions | CI/CD | `.github/workflows/ci.yml`, `deploy.yml` | ✅ |
| GitLab CI | mirror | `.gitlab-ci.yml` | ✅ |
| Monitoring | optional | `deployments/docker-compose.monitoring.yml` | Prometheus, Loki, Grafana (pinned); Tempo via `--profile otel` (A14-009) |

### Caddy routing (prod)

```
/api/*  → backend:3000
/ws     → backend:3000 (HTTP/1.1 upgrade)
/lk/*   → livekit:7880 (strip prefix)
/*      → frontend:80 (static)
```

---

## CI/CD (GitHub Actions — current legacy inventory)

Файл: `.github/workflows/ci.yml` · Ветка: `master` · Concurrency:
cancel-in-progress. Таблица ниже фиксирует существующий до-cutover workflow, но
не является target execution contract и не разрешает host/inline Docker запуск.

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

**Private runner source candidate: `STATIC_SOURCE_REVIEWED`; target launcher:
`NOT_READY`; `PS5.1_PARSE_NOT_RUN`; `PRIVATE_CONTRACT_NOT_RUN`; runtime:
`NOT_RUN`.** `testframework/runner/**` — единственный launcher owner;
`testframework/docker/**` предоставляет ему только immutable execution
assets/contracts. Required `images.lock.v3.json` и dependency bytes отсутствуют.
После приёмки GitHub jobs и `/qa-gate` вызывают только этот launcher:
task-owned names, immutable image provenance, CPU/RAM/swap/PID/network limits,
outer timeout, first-exit preservation, `reports/<task-run>/` owner marker и
exact cleanup в `finally`. Legacy jobs и скрипты выше
удаляются/переключаются по cutover manifest, без dual-run/parity.

---

## Внешние сервисы

| Сервис | Назначение | Ограничения |
|--------|------------|-------------|
| Resend / hub-mail | Magic Link email, transactional | Rate limits провайдера |
| Telegram Bot API | Уведомления (не login) | telego SDK |
| GitHub / GitLab | Webhooks, PR/MR | Tokens encrypted at rest |

---

## Линтеры и quality gates

До binding и независимой parser/private-contract/Docker приёмки
`testframework/runner/**` target launcher имеет статус `NOT_READY`,
`PS5.1_PARSE_NOT_RUN`, `PRIVATE_CONTRACT_NOT_RUN`, runtime — `NOT_RUN`. Immutable
execution assets принадлежат `testframework/docker/**`, но сами по себе не дают
runtime authority. Команды старого CI ниже — только инвентарный снимок, а не
разрешённый fallback. После cutover каждый retained job вызывает единственный
launcher с task-owned report, ограничениями ресурсов и unconditional cleanup.

| Целевой job | Область | Исполнение | Статус |
|-------------|---------|------------|--------|
| `backend-unit` | genuine Go units + backend static gates | `/qa-gate` → sole `testframework/runner/**` boundary | `NOT_READY` |
| `frontend-unit` | genuine Vitest/RTL units + TS/Biome/build gates | `/qa-gate` → sole `testframework/runner/**` boundary | `NOT_READY` |
| `python-contract` / `python-api` / `python-integration` | greenfield pytest suites | `/qa-gate` → sole `testframework/runner/**` boundary | `NOT_READY` |
| `python-e2e-smoke` / `python-e2e-full` | greenfield Playwright Python journeys | `/qa-gate` → sole `testframework/runner/**` boundary | `NOT_READY` |
| `python-load` | bounded Locust profiles | `/qa-gate` → sole `testframework/runner/**` boundary | `NOT_READY` |
| dependabot | weekly go/npm/actions | вне test runner | не настроен |

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
    ├── PACKAGE-CONTRACT.md   # docs-only future package/bootstrap contract
    ├── src/hub_testkit/      # reusable installed framework package
    │   └── py.typed          # future typed-package marker; NOT_READY
    ├── tests/
    │   └── {contract,api,integration,e2e}/ # executable pytest suites
    ├── selftests/            # framework contracts only
    ├── cases/                # TC intent + risk links
    ├── schemas/              # closed case/run contracts
    ├── policies/             # marker/environment/redaction/profile policy
    ├── load/                 # separate Locust lifecycle
    ├── docker/               # immutable execution assets only
    ├── runner/               # sole launcher boundary; private static candidate exists
    ├── reports/<task-run>/   # owner-marked generated evidence
    ├── pyproject.toml        # future package/tool authority; NOT_READY
    └── uv.lock               # future exact resolution; NOT_READY
```

Structure/boundary scaffold в commit `22035c45` уже существует. Private runner
source candidate добавлен отдельно и имеет только статический review; scaffold
не включает future package files или executable suites. Reusable imports и suite
roots физически разделены; direct root `core/api/integration/contract/e2e`
packages не являются target.

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
*Последнее обновление: 2026-08-09 — package/suite roots, structure-only scaffold и runner/docker ownership.*
