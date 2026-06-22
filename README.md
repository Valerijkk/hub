<div align="center">

# Hub

**Self-hosted corporate super-app для команд разработки**

Трекер · Мессенджер · TMS · Звонки · Интранет · CRM · HR · Учёт · AI

<p>
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=flat&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/React-19-61DAFB?style=flat&logo=react&logoColor=black" alt="React" />
  <img src="https://img.shields.io/badge/TypeScript-5.7-3178C6?style=flat&logo=typescript&logoColor=white" alt="TypeScript" />
  <img src="https://img.shields.io/badge/PostgreSQL-17-4169E1?style=flat&logo=postgresql&logoColor=white" alt="PostgreSQL" />
  <img src="https://img.shields.io/badge/MongoDB-8-47A248?style=flat&logo=mongodb&logoColor=white" alt="MongoDB" />
  <img src="https://img.shields.io/badge/Redis-7-DC382D?style=flat&logo=redis&logoColor=white" alt="Redis" />
  <img src="https://img.shields.io/badge/Coverage-82.6%25-brightgreen" alt="Coverage" />
  <img src="https://img.shields.io/badge/License-Private-red" alt="License" />
</p>

<p>
  <strong>286+ API endpoints</strong> · <strong>39+ frontend routes</strong> · <strong>22 E2E specs</strong> · <strong>Product score 93/100</strong>
</p>

</div>

---

Hub — self-hosted платформа, заменяющая **Jira + Slack + TestRail + Zoom + Bitrix** для команд 10–50 человек. Dev-suite (трекер, мессенджер, TMS, звонки, дашборды) и корпоративные модули (интранет, календарь, CRM-lite, HR, учёт, no-code базы, AI hub) — в одном инстансе с единой авторизацией Magic Link и RBAC.

## Возможности

### Dev-suite

#### Трекер задач
- 9 типов задач (Epic, Story, Task, Bug, Sub-Task, Improvement, Research, Design, DevOps)
- Kanban с WIP-лимитами и swimlanes · Scrum: спринты, Planning Poker, burndown, velocity
- Кастомный workflow · JQL · automation lite · time tracking
- Связи между задачами · GitHub/GitLab webhooks · Gantt (v2.6)

#### Мессенджер
- Каналы (публичные/приватные), DM, треды, reactions, polls, voice messages
- @hub-ai (AI gateway: Ollama + BYOK) · presence · #hub-ci read-only channel
- Глобальный поиск Ctrl+K · issue↔message deep links

#### TMS (Test Management)
- Classic/BDD кейсы · suites · runs · Allure CI import · hub-cli push-results
- Traceability matrix · run comparison · env compare · flaky trends · PDF export

#### Звонки
- LiveKit: voice + screen share + webcams · до 50 участников
- Запись · планирование · instant-call из мессенджера · ephemeral chat

#### Дашборды
- Personal / project / admin · 12+ виджетов · drag-and-drop layout · шаблоны

### Corporate super-app (v4)

| Модуль | Возможности |
|--------|-------------|
| **Интранет** | News feed, kudos, Random Coffee |
| **Org directory** | MOD-02, поиск людей в Ctrl+K |
| **Календарь** | Unified events, iCal export |
| **CRM-lite** | Contacts, deals kanban, won→invoice |
| **HR-lite** | Leave requests, 1C webhook |
| **Marketplace** | Classifieds, mod queue |
| **Status page** | Public `/status` |
| **Bases** | Lark-style no-code tables |
| **Accounting** | Ledger, journal, invoices, P&L |

### Platform

- **Auth:** Magic Link (email) + Telegram bind · JWT 15m · refresh 30d · org + founder role
- **RBAC:** Owner/Admin/Member/Viewer + custom roles
- **Notifications:** in-app + Web Push + Telegram · quiet hours
- **Search:** Meilisearch + Ctrl+K · user-scope (NF-03)
- **AI Hub:** BYOK + Ollama · composer suggest · external AI consent
- **Observability:** Prometheus · Loki · OTel · Alertmanager → audit log
- **Integrations:** GitHub, GitLab, Jira import, 1C connector, hub-cli
- **Admin:** audit log, trash 30d, backups, ban, GDPR, resource limits
- **i18n:** RU + EN · Dark/Light/System themes · PWA

## Технологический стек

### Backend

| Технология | Версия | Назначение |
|---|---|---|
| Go + Fiber v3 | 1.25 | HTTP, WebSocket, middleware |
| Ent ORM + Atlas | — | PostgreSQL: схемы, миграции |
| MongoDB | 8 | Сообщения, Allure reports |
| Redis | 7 | Кеш, сессии, presence, rate limiting |
| NATS + JetStream | — | Pub/sub, уведомления, аудит |
| LiveKit | — | WebRTC: звонки, запись |
| SeaweedFS | — | S3-совместимое хранилище |
| Meilisearch | — | Full-text search |

### Frontend

| Технология | Версия | Назначение |
|---|---|---|
| React | 19 | UI |
| TypeScript | 5.7 | Типобезопасность |
| TanStack Router | — | 39+ маршрутов |
| TanStack Query + Zustand | — | Server/client state |
| Shadcn/ui + Tailwind CSS 4 | — | Дизайн-система |
| Plate | — | Rich-text редактор |
| LiveKit React | — | UI звонков |
| Vite | — | Сборка |

### Инфраструктура

| Технология | Назначение |
|---|---|
| Docker Compose | Dev + test (`deployments/docker-compose.test.yml`) |
| Caddy | Reverse proxy, HTTPS |
| K3s | Production Kubernetes |
| GitHub Actions | CI/CD (release build, race tests, E2E) |
| Prometheus + Grafana + Loki | Observability stack |

## Архитектура

```
                          ┌──────────────────────┐
                          │       Caddy          │
                          │  (reverse proxy, TLS)│
                          └──────┬───────┬───────┘
                                 │       │
                    ┌────────────┘       └────────────┐
                    v                                  v
          ┌─────────────────┐              ┌─────────────────────┐
          │  Frontend (SPA) │              │  Backend (Go+Fiber) │
          │  React 19       │   REST/WS    │  286+ endpoints     │
          │  39+ routes     │ ◄──────────► │  WebSocket + NATS   │
          └─────────────────┘              └──────────┬──────────┘
                                                      │
          ┌────────────┬────────────┬───────┴───────┬──────────┐
          v            v            v               v          v
   ┌───────────┐ ┌─────────┐ ┌──────────┐  ┌────────────┐ ┌─────────┐
   │PostgreSQL │ │ MongoDB │ │  Redis   │  │    NATS    │ │Meilisearch│
   │  17       │ │  8      │ │  7       │  │ JetStream  │ │         │
   └───────────┘ └─────────┘ └──────────┘  └────────────┘ └─────────┘
          │
   ┌──────┴──────┐              ┌────────────┐
   v             v              v            v
SeaweedFS    LiveKit      Prometheus    testframework/
 (файлы)     (WebRTC)     Loki/OTel     (Go + E2E tests)
```

## Быстрый старт

### Требования

- Docker + Docker Compose v2
- Git

### Локальная разработка

```bash
git clone https://github.com/Valerijkk/hub.git
cd hub
docker compose -f docker-compose.local.yml up -d
```

| Сервис | URL |
|---|---|
| Frontend | http://localhost:5173 |
| Backend API | http://localhost:3000/api |
| Swagger | http://localhost:3000/docs |
| SeaweedFS | http://localhost:9333 |

### Production

```bash
cd deployments
docker compose -f docker-compose.prod.yml --env-file .env up -d
```

Подробнее: [docs/wiki/DEPLOY.md](docs/wiki/DEPLOY.md) · [docs/wiki/QUICKSTART.md](docs/wiki/QUICKSTART.md)

## Структура проекта

```
hub/
├── backend/
│   ├── cmd/server/           # Точка входа
│   ├── ent/schema/           # ~40 Ent схем
│   ├── internal/             # 25+ модулей (auth, tracker, tms, intranet, crm, …)
│   └── testframework/        # Unified test suite (-tags hubtest)
├── frontend/src/
│   ├── api/                  # API-клиент
│   ├── modules/              # tracker, tms, messenger, intranet, crm, …
│   ├── routes/               # TanStack Router
│   └── stores/               # Zustand
├── testframework/e2e/        # Playwright (22 specs)
├── cli/                      # hub-cli (CI/CD)
├── deployments/              # Docker, K3s, observability
└── docs/wiki/                # API, архитектура, деплой
```

## API

**286+ REST endpoints** по модулям:

| Модуль | ~Endpoints | Описание |
|---|---|---|
| Auth | 13 | Magic Link, sessions, org |
| Projects | 22+ | CRUD, teams, templates |
| Tracker | 38+ | Issues, workflow, boards, JQL |
| Messenger | 42+ | Channels, messages, presence |
| TMS | 52+ | Cases, runs, Allure, reports |
| Calls | 22+ | LiveKit, recording |
| v4 modules | 40+ | Intranet, CRM, HR, calendar, bases, accounting |
| AI | 12+ | Gateway, composer, preferences |
| Admin/OBS | 15+ | Audit, metrics, observability |
| Integrations | 20+ | GitHub, GitLab, 1C, webhooks |

Полная документация: [docs/wiki/API.md](docs/wiki/API.md)

## Качество и тесты

| Метрика | Значение |
|---|---|
| Backend coverage | **82.6%** (`internal/`) |
| Security packages ≥80% | **6/6** ✅ |
| Frontend Vitest | **99+** tests |
| Playwright E2E | **22** specs |
| Product score | **93/100** |

```bash
# Backend (Docker, canonical)
docker compose -f deployments/docker-compose.test.yml up -d
docker run --rm --network deployments_default \
  -v "${PWD}/backend:/app" -w /app \
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" \
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" \
  golang:1.25-bookworm \
  sh -c "go test -tags hubtest -race -count=1 ./internal/... ./testframework/..."

# Frontend
cd frontend && npx tsc --noEmit && npx vitest run
```

## Дорожная карта

| Версия | Статус | Описание |
|---|---|---|
| **v1.0** | ✅ Shipped 2026-03-21 | MVP: tracker, messenger, TMS, calls, dashboards |
| **v2.0** | ✅ 2026-05-15 | Hub Redesign (Slack-style UI) |
| **v2.5 Trust** | ✅ 2026-06-22 | Magic Link, org, E2E, WCAG, coverage ≥50% |
| **v2.6 Automate** | ✅ 2026-06-24 | Automation, OBS stack, ARCH-01, Biome gate |
| **v3.0 AI Native** | ✅ 2026-06-22 | AI gateway, @hub-ai, full JQL |
| **v4 Super-app** | ✅ 2026-06-25 | Intranet, CRM, HR, calendar, bases, accounting |
| **v4.5+** | 📋 Planned | Accounting advanced, org chart, TF migration complete |

Детали: [.planning/ROADMAP.md](.planning/ROADMAP.md) · [.planning/STATE.md](.planning/STATE.md)

## Документация

| Документ | Назначение |
|----------|------------|
| [docs/README.md](docs/README.md) | Индекс базы знаний |
| [docs/competitive_analysis_2026.md](docs/competitive_analysis_2026.md) | Конкурентный анализ |
| [docs/user_needs_analysis_2026.md](docs/user_needs_analysis_2026.md) | Персоны и JTBD |
| [AGENTS.md](AGENTS.md) | Регламент AI-команды |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Git workflow, code style |

## Контрибуция

См. [CONTRIBUTING.md](CONTRIBUTING.md): Git workflow, conventional commits, code review, `/qa-gate` перед PR.

## Лицензия

Приватный проект. Все права защищены.

---

<div align="center">

**Self-hosted super-app для dev-команд**

Go · React · PostgreSQL · MongoDB · Redis · NATS · LiveKit · Meilisearch

*Обновлено: 2026-06-21*

</div>
