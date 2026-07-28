<div align="center">

# Hub

**Self-hosted корпоративный super-app для команд 10–50 человек**

Заменяет **Jira + Slack + TestRail + Zoom + Grafana + модули Bitrix** одним инстансом:
трекер · мессенджер · TMS · звонки · дашборды · интранет · календарь · CRM · HR · учёт · no-code базы · AI-хаб

<p>
  <!-- Бейджи-плейсхолдеры: заменить на реальные URL CI после публикации -->
  <img src="https://img.shields.io/badge/build-passing-brightgreen" alt="Build" />
  <img src="https://img.shields.io/badge/coverage-83%25_(2026--06--22)-brightgreen" alt="Coverage 83% (2026-06-22 Docker baseline)" />
  <img src="https://img.shields.io/badge/license-proprietary-red" alt="License" />
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=flat&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/React-19-61DAFB?style=flat&logo=react&logoColor=black" alt="React" />
  <img src="https://img.shields.io/badge/TypeScript-5.7-3178C6?style=flat&logo=typescript&logoColor=white" alt="TypeScript" />
  <img src="https://img.shields.io/badge/PostgreSQL-17-4169E1?style=flat&logo=postgresql&logoColor=white" alt="PostgreSQL" />
  <img src="https://img.shields.io/badge/MongoDB-8-47A248?style=flat&logo=mongodb&logoColor=white" alt="MongoDB" />
  <img src="https://img.shields.io/badge/Redis-7-DC382D?style=flat&logo=redis&logoColor=white" alt="Redis" />
</p>

<p>
  <strong>286+ API endpoints</strong> · <strong>82 Ent-схемы</strong> · <strong>72 frontend route files</strong> · <strong>515 backend-тестов</strong> · <strong>Product score 89/100</strong>
  <!-- inventory measured 2026-07-27; was 39+/495 -->
</p>

</div>

---

Hub — self-hosted платформа для команд разработки и небольших компаний: полный dev-suite (трекер, мессенджер, TMS, звонки, дашборды) и корпоративные модули (интранет, календарь, CRM-lite, HR, управленческий учёт, no-code базы, AI-хаб) в одном инстансе с единой авторизацией по Magic Link, RBAC и сквозным поиском Ctrl+K. Один сервер, ваши данные, никаких SaaS-подписок.

## Содержание

- [Модули](#модули)
- [Скриншоты](#скриншоты)
- [Архитектура](#архитектура)
- [Технологический стек](#технологический-стек)
- [Быстрый старт (dev)](#быстрый-старт-dev)
- [Deployment (prod)](#deployment-prod)
- [Тестирование](#тестирование)
- [AI-команда разработки](#ai-команда-разработки)
- [Roadmap](#roadmap)
- [Документация](#документация)
- [Безопасность](#безопасность)
- [Лицензия](#лицензия)

## Модули

| Модуль | Описание | Статус |
|--------|----------|--------|
| **Tracker** | Трекер задач: 9 типов задач, Kanban с WIP-лимитами и swimlanes, кастомные workflow, JQL, automation, time tracking, Gantt, GitHub/GitLab webhooks | ✅ Стабильный |
| **Sprint** | Scrum-контур: спринты, Planning Poker, burndown, velocity, ежедневные стендапы | ✅ Стабильный |
| **Messenger** | Каналы (публичные/приватные), DM, треды, реакции, опросы, голосовые сообщения, presence, deep links задача↔сообщение, @hub-ai в чате | ✅ Стабильный |
| **TMS** | Test Management: Classic/BDD кейсы, suites, runs, **импорт Allure из CI** (hub-cli), traceability matrix, flaky trends, сравнение прогонов, PDF-отчёты | ✅ Стабильный |
| **Calls** | Звонки на LiveKit (WebRTC): голос, screen share, веб-камеры, до 50 участников, запись, планирование, instant-call из мессенджера | ✅ Стабильный |
| **Dashboards** | Personal / project / admin дашборды: 12+ виджетов, drag-and-drop layout, шаблоны | ✅ Стабильный |
| **Search + JQL** | Глобальный поиск Ctrl+K на Meilisearch со строгим user-scope; полный JQL для задач | ✅ Стабильный |
| **Notifications** | In-app + Web Push + Telegram, quiet hours; пилот выноса в отдельный сервис `hub-notification` | ✅ Стабильный |
| **AI-хаб** | «Любая ИИшка» — LiteLLM-паттерн: универсальный OpenAI-совместимый адаптер, BYOK, 5+ провайдеров + Ollama, каталог моделей, virtual keys/бюджеты; скоро — MCP-сервер и AI-агенты | ✅ Стабильный (MCP/агенты — планируется) |
| **Intranet** | Корпоративная лента: новости, kudos, Random Coffee, опросы, onboarding-шаблоны | 🧪 MVP (расширение в работе) |
| **Calendar** | Единый календарь событий, iCal export/подписки, ресурсы, recurring-события, busy/free | 🧪 MVP (расширение в работе) |
| **CRM** | CRM-lite: контакты, компании, deals-kanban, won→invoice, импорт CSV/XLSX | 🧪 MVP (расширение в работе) |
| **HR** | Заявки на отпуск, org chart, грейды, авто-начисление отпусков, 1C webhook | 🧪 MVP (расширение в работе) |
| **Accounting** | Управленческий учёт: ledger, журнал проводок, инвойсы, P&L, баланс, cash flow | 🧪 MVP (расширение в работе) |
| **Bases** | No-code базы данных в стиле Lark/Airtable: таблицы, 4 view, формулы, realtime | 🧪 MVP (эпик BS-0..5 в работе) |
| **Wiki** | База знаний: spaces → страницы → версии, связь с задачами, RBAC на уровне space | 📋 Планируется (спека в Волне 1–2) |
| **SSO** | OIDC + LDAP/AD, org-политика magic-link/sso/both, JIT-provisioning, break-glass | 📋 Планируется (после ADR-AUTH-SSO) |

**Платформенное ядро:** Magic Link auth + JWT · RBAC (system + project + custom roles) · организация с invite-регистрацией · audit log · корзина 30 дней · бэкапы · GDPR-delete · i18n RU/EN · тёмная/светлая темы · PWA · observability (Prometheus/Loki/OTel).

## Скриншоты

<!-- Плейсхолдеры: добавить скриншоты в docs/assets/screenshots/ и заменить ссылки -->

| | |
|---|---|
| ![Kanban-доска трекера](docs/assets/screenshots/tracker-kanban.png) *Kanban с WIP-лимитами* | ![Мессенджер](docs/assets/screenshots/messenger.png) *Каналы и треды* |
| ![TMS: прогон с Allure](docs/assets/screenshots/tms-run.png) *TMS: импорт Allure из CI* | ![Звонок LiveKit](docs/assets/screenshots/calls.png) *Звонки со screen share* |
| ![Дашборд проекта](docs/assets/screenshots/dashboard.png) *Дашборд проекта* | ![AI-хаб](docs/assets/screenshots/ai-hub.png) *AI-хаб: любая модель через BYOK* |

## Архитектура

Hub сегодня — **модульный монолит** на Go/Fiber (31+ внутренних модулей, 82 Ent-схемы) с уже вынесенными спутниками (`hub-mail`, `meilisearch-indexer`, внешний LiveKit). Принят ADR о поэтапном выносе доменов в микросервисы по стратегии **strangler fig** — без big-bang, обратимо (route-flip в Caddy):

```
Модульный монолит (hub-core)
   │  Фаза 0: hardening границ (repository-интерфейсы, запрет cross-domain импортов)
   ├─► hub-notification   — пилот выноса (чистый event-consumer) → GATE Go/No-Go
   ├─► hub-ai             — AI-gateway (LiteLLM-паттерн, BYOK-изоляция, MCP)
   └─► realtime-трио      — hub-ws / hub-messenger / hub-calls
```

Ключевые решения:

- **Данные:** shared-таблицы между сервисами запрещены; старт — **schema-per-service** в одном PostgreSQL 17 (свой DB-user + GRANT только на свою схему), физически отдельные БД — точечно по нагрузке. Кросс-доменное чтение — CQRS-проекции через события, не прямые запросы.
- **Взаимодействие:** async через **NATS JetStream + transactional outbox**, идемпотентные consumers; sync-вызовы минимизированы (правило: >1 sync-хопа на запрос = граница нарезана неверно).
- **Auth между сервисами:** HS256 → асимметричная подпись + JWKS; refresh/revoke централизованно в Redis (инварианты Magic Link / 15 мин / 30 дней сохраняются).
- **Предусловия выноса:** OTel trace-context через NATS-headers, path-filtered CI, contract-тесты, K3s namespaces + Caddy ingress.

Подробно: [docs/architecture.md](docs/architecture.md) · ADR: [.planning/decisions/ADR-ARCH-MICROSERVICES-2026.md](.planning/decisions/ADR-ARCH-MICROSERVICES-2026.md) · roadmap выноса: [.planning/archive/roadmaps/MICROSERVICES-ROADMAP.md](.planning/archive/roadmaps/MICROSERVICES-ROADMAP.md)

```
                          ┌──────────────────────┐
                          │        Caddy         │
                          │ (reverse proxy, TLS) │
                          └──────┬───────┬───────┘
                                 │       │
                    ┌────────────┘       └────────────┐
                    v                                  v
          ┌─────────────────┐              ┌─────────────────────┐
          │  Frontend (SPA) │              │  Backend (Go+Fiber) │
          │  React 19       │   REST/WS    │  286+ endpoints     │
          │  72 route files │ ◄──────────► │  WebSocket + NATS   │
          └─────────────────┘              └──────────┬──────────┘
                                                      │
          ┌────────────┬────────────┬───────┴───────┬──────────┐
          v            v            v               v          v
   ┌───────────┐ ┌─────────┐ ┌──────────┐  ┌────────────┐ ┌───────────┐
   │PostgreSQL │ │ MongoDB │ │  Redis   │  │    NATS    │ │Meilisearch│
   │    17     │ │    8    │ │    7     │  │ JetStream  │ │           │
   └───────────┘ └─────────┘ └──────────┘  └────────────┘ └───────────┘
          │
   ┌──────┴──────┐              ┌────────────┐
   v             v              v            v
SeaweedFS    LiveKit      Prometheus    testframework/
 (файлы)     (WebRTC)     Loki/OTel     (тесты всех слоёв)
```

## Технологический стек

### Backend

| Технология | Версия | Назначение |
|---|---|---|
| Go | 1.25 (`go.mod`) | Backend, CLI, hub-mail |
| Fiber v3 | v3.0.0-beta.3 | HTTP, WebSocket (fasthttp) |
| Ent + Atlas | v0.14.x | ORM, кодогенерация схем, миграции |
| PostgreSQL | 17 | Основная БД (users, projects, issues, TMS, workflow) |
| MongoDB | 8 | Сообщения, Allure-отчёты, call-chat |
| Redis | 7 | Refresh-токены, RBAC-кеш, presence, rate limiting |
| NATS + JetStream | 2.11 | Pub/sub, outbox, уведомления, аудит |
| LiveKit | v1.8 | WebRTC: звонки, screen share, запись (egress) |
| SeaweedFS | 3.97 (pinned) | S3-совместимое файловое хранилище |
| Meilisearch | — | Full-text search (Ctrl+K) |

### Frontend

| Технология | Версия | Назначение |
|---|---|---|
| React | 19 | UI |
| TypeScript | 5.7 | Типобезопасность |
| Vite | 6 | Dev server, сборка, PWA |
| TanStack Router | 1.x | File-based routing (**72** `.tsx` under `frontend/src/routes/`, measured 2026-07-27; was claimed 39+) |
| TanStack Query + Zustand | 5.x | Server / client state (11+ stores) |
| Tailwind CSS 4 + Shadcn/ui | — | Дизайн-система (токены `--hub-*`) |
| Plate | 48.x | Rich-text редактор |
| Biome | latest | Lint + format (CI gate) |

### Инфраструктура

| Технология | Назначение |
|---|---|
| Docker Compose | Dev (`docker-compose.local.yml`) + test (`deployments/docker-compose.test.yml`) |
| K3s | Production Kubernetes (namespaces per-service, NetworkPolicy) |
| Caddy | Reverse proxy, авто-HTTPS, WS, route-flip при выносе сервисов |
| GitHub Actions | CI/CD: release build, race-тесты, E2E, coverage-gate, deploy по green CI |
| Prometheus + Grafana + Loki + OTel | Observability stack |

Полный стек с версиями и обоснованием решений: [docs/stack_and_tools.md](docs/stack_and_tools.md)

## Быстрый старт (dev)

### Требования

- Docker + Docker Compose v2
- Go 1.25+, Node.js 20+ (для нативного запуска backend/frontend)
- Git

### Запуск

```bash
git clone <repo-url> hub && cd hub

# 1. Инфраструктура (PostgreSQL, MongoDB, Redis, NATS, SeaweedFS, LiveKit)
docker compose -f docker-compose.local.yml up -d

# 2. Backend (нативно, с hot-однокомандным запуском)
cd backend
go generate ./ent
go run ./cmd/server

# 3. Frontend
cd frontend
npm ci
npm run dev
```

| Сервис | URL |
|---|---|
| Frontend | http://localhost:5173 |
| Backend API | http://localhost:3000/api |
| Swagger (dev) | http://localhost:3000/docs |
| SeaweedFS | http://localhost:9333 |

### Dev-вход

В dev-режиме (`DEV_MODE=true`) доступен упрощённый вход без реальной отправки email — magic-link код выводится в лог backend. В production-сборке (`go build -tags release`) все `/dev/*` маршруты исключены из бинарника.

Подробнее: [docs/wiki/QUICKSTART.md](docs/wiki/QUICKSTART.md) · горячие клавиши: [docs/wiki/SHORTCUTS.md](docs/wiki/SHORTCUTS.md)

## Deployment (prod)

Production-топология: **K3s** (single-node на старте, local-path PVC) + **Caddy** как ingress/reverse proxy с авто-TLS. Переходный вариант — Docker Compose:

```bash
cd deployments
docker compose -f docker-compose.prod.yml --env-file .env up -d
```

- Деплой запускается **только после успешного CI** на push в `master` (GHCR-образы с SHA-тегом, SSH deploy, post-deploy health check).
- Никаких `DEV_MODE=true` и hardcoded-секретов в prod-конфигурации — секреты только через env / K8s Secrets.
- Caddy-маршрутизация: `/api/*` → backend, `/ws` → backend (upgrade), `/lk/*` → LiveKit, `/*` → frontend static.

Полный гайд: [docs/wiki/DEPLOY.md](docs/wiki/DEPLOY.md) · hardening: [SECURITY.md](SECURITY.md)

## Тестирование

Тестовая пирамида Hub (целевая модель: **юниты пишут разработчики, всё остальное — QA-конвейер**):

| Слой | Стек | Кто пишет | Статус |
|---|---|---|---|
| Unit backend | Go (testify, `-tags hubtest`, `-race`) | be-разработчики | ✅ **515** test files (measured 2026-07-27; was 495), coverage **83%** (2026-06-22 Docker baseline) |
| Unit frontend | TypeScript: Vitest + RTL | fe-разработчики | ✅ 99+ тестов |
| E2E | Playwright (TS), smoke + full | QA | ✅ 20+ спеков (сейчас) |
| Integration / API / contract / regression | **Python: pytest + allure-pytest + httpx + playwright-python** | QA-конвейер | 📋 целевой слой (фаза TF-PY, после выноса микросервисов) |
| Нагрузочные | **Python: Locust** (500+ concurrent WS, p95 < 200ms API) | QA | 📋 baseline-профиль до выноса realtime |

Качество и ворота:

- **Coverage-gate:** backend ≥ 82% (сейчас **83.0%**, 2026-06-22 Docker baseline), security-пакеты (auth/ws/rbac/audit/calls +1) ≥ 80% — **6/6** ✅.
- **`/qa-gate`** перед handoff/merge: release-build → hubtest → race → Biome → e2e-smoke.
- **QA-конвейер:** тест-кейсы (TC-\*) → верификация → автотесты по TC → верификация автотестов; трассируемость TC↔автотест через Allure, результаты заливаются в собственный TMS Hub.

```bash
# Backend regression (Docker, canonical)
docker compose -f deployments/docker-compose.test.yml up -d
docker run --rm --network deployments_default \
  -v "${PWD}/backend:/app" -w /app \
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" \
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" \
  golang:1.25-bookworm \
  sh -c "go test -tags hubtest -race -count=1 -timeout=600s ./testframework/..."

# Frontend regression
cd frontend && node scripts/link-testframework-node-modules.mjs && npx tsc --noEmit && npm run lint && npm run test:run
```

Стратегия тестового фреймворка: [docs/testframework_strategy.md](docs/testframework_strategy.md) · тест-план: [docs/test_plan.md](docs/test_plan.md) · карта pipeline: [testframework/PIPELINE.md](testframework/PIPELINE.md)

## AI-команда разработки

Hub разрабатывается **командой из 30 AI-агентов** с ролями, зеркалящими реальную продуктовую команду:

- **Оркестрация:** `team-lead` (Fable 5, thinking-high) — декомпозиция волн, параллельный запуск, финальная приёмка; эскалация на `team-lead-opus` (Opus 4.8 @ 1M контекст) для самых тяжёлых декомпозиций и приёмки milestone; остальные агенты — Sonnet 5 (max effort).
- **Потоки:** дублированные роли (ba1/ba2, sa1/sa2, be1/be2, fe1/fe2, qa-auto1/qa-auto2) работают параллельно с разделением по модулям и взаимным кросс-ревью.
- **Верификаторы:** у каждого артефакта персональный `*-ver` (ba-ver, sa-ver, arch-ver, designer-ver, be-ver, fe-ver, qa-manual-ver, qa-auto-ver) — `verified: true` только при green-регрессии с приложенным логом.
- **Cross-cutting:** devops-engineer, security-auditor, debugger, code-reviewer, docs-writer.
- **QA-конвейер:** тест-кейсы → верификация TC → автотесты → верификация автотестов; юниты пишут разработчики.
- **Практики:** 10 playbooks (backend, frontend, design, testing, observability, git-workflow, analysis, verification, team-workflow, code-comments) — единый источник ремесленных стандартов для всей команды.

Ростер, модели и RACI: [.cursor/playbooks/README.md](.cursor/playbooks/README.md) · регламент: [.cursor/REGULATION.md](.cursor/REGULATION.md) · воркспейс: [AGENTS.md](AGENTS.md)

## Roadmap

Выпущено:

| Версия | Дата | Содержание |
|---|---|---|
| **v1.0 MVP** | 2026-03-21 | Tracker, messenger, TMS, calls, dashboards |
| **v2.0 Hub Redesign** | 2026-05-15 | Slack-style UI |
| **v2.5 Trust** | 2026-06-22 | Magic Link, организация, E2E, WCAG, security-фиксы |
| **v2.6 Automate** | 2026-06-24 | Automation, OBS stack, Biome gate |
| **v3.0 AI Native** | 2026-06-22 | AI gateway, @hub-ai, полный JQL |
| **v4 Super-app** | 2026-06-25 | Intranet, calendar, CRM, HR, bases, accounting |

Текущий план — **два параллельных трека** (~12–14 волн):

- **Track F (функционал):** quick-wins и security-фиксы → паритет v4-модулей с полноценными продуктами (intranet, calendar, CRM, HR, accounting, bases-эпик) → **новые модули: Wiki, AI-хаб по LiteLLM-паттерну + MCP-сервер и AI-агенты, SSO (OIDC + LDAP)**.
- **Track P (платформа):** hardening модульного монолита → пилот выноса `hub-notification` (strangler fig) → GATE Go/No-Go → `hub-ai` → realtime-трио (`hub-ws` / `hub-messenger` / `hub-calls`); schema-per-service, NATS + outbox, K3s.
- **Ops-лейн:** ротация секретов (FA-06) → тег v2.5.0 → SSH deploy.

Детальный план: [.planning/archive/july-2026/ROADMAP-FINAL-2026-07-03.md](.planning/archive/july-2026/ROADMAP-FINAL-2026-07-03.md) · текущее состояние: [.planning/STATE.md](.planning/STATE.md)

## Документация

Единая база знаний в `docs/` — источник правды для команды (людей и агентов):

| Документ | Назначение |
|----------|------------|
| [docs/README.md](docs/README.md) | Индекс базы знаний |
| [docs/vision_and_scope.md](docs/vision_and_scope.md) | Бизнес-цели, KPI, User Stories, MoSCoW |
| [docs/system_specification.md](docs/system_specification.md) | FR/NFR, API-контракты, ERD |
| [docs/architecture.md](docs/architecture.md) | C4, модули, потоки данных, deployment |
| [docs/stack_and_tools.md](docs/stack_and_tools.md) | Версии стека, CI/CD, обоснование решений |
| [docs/test_plan.md](docs/test_plan.md) | Стратегия тестирования, регрессия |
| [docs/testframework_strategy.md](docs/testframework_strategy.md) | Целевая тестовая пирамида (pytest+allure+Locust) |
| [docs/code_style_guide.md](docs/code_style_guide.md) | Стиль Go / TypeScript |
| [docs/security_guidelines.md](docs/security_guidelines.md) | OWASP-чек-лист, статусы findings, политики |
| [docs/tz_constraints.md](docs/tz_constraints.md) | Инварианты ТЗ: mandatory / forbidden |
| [docs/competitive_analysis_2026.md](docs/competitive_analysis_2026.md) | Конкурентный анализ (Jira, Slack, Bitrix, Lark…) |
| [docs/user_needs_analysis_2026.md](docs/user_needs_analysis_2026.md) | Персоны, JTBD |
| [docs/onboarding/](docs/onboarding/) | Онбординг-чек-листы по ролям (BE/FE/QA/PM/ops) |
| [docs/wiki/](docs/wiki/) | API.md (286+ endpoints), QUICKSTART, SHORTCUTS, DEPLOY |
| [SECURITY.md](SECURITY.md) | Политика уязвимостей, security-инварианты, hardening |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Git workflow, conventional commits, review |

## Безопасность

Политика сообщений об уязвимостях, security-инварианты платформы и hardening-guide для self-hosted инсталляций — в [SECURITY.md](SECURITY.md). Технические детали и статусы аудитов — в [docs/security_guidelines.md](docs/security_guidelines.md).

## Лицензия

**Проприетарное ПО. All Rights Reserved.**

Использование, копирование, распространение, модификация и создание производных работ **запрещены без предварительного письменного разрешения правообладателя**. Доступ к репозиторию не означает предоставления каких-либо прав на ПО.

Полный текст: [LICENSE](LICENSE)

---

## Планы на август 2026

Полный план: [.planning/active/august-2026/00-AUGUST-PLAN.md](.planning/active/august-2026/00-AUGUST-PLAN.md)

Два трека, идущих строго последовательно — второй опасен без первого.

### Трек 1 — тестовый фреймворк на Python

[Документ](.planning/audits/planning-audit-2026-07/02-TEST-FRAMEWORK-MIGRATION.md) · фазы 128–136

Фреймворк выносится в корневую папку и переписывается целиком на **pytest +
Playwright + Locust**. Каталоги `testframework/` в `backend/` и `frontend/`
ликвидируются. Юнит-тесты возвращаются **к коду** и пишутся на языке своего слоя
разработчиками, проверяются тестировщиками.

| Фаза | Что |
|---|---|
| 128 | Каркас Python-проекта `/tests` |
| 129–131 | Перенос E2E на Playwright Python (20 сценариев) |
| 132 | Go-юниты возвращаются в `backend/internal/**` |
| 133 | Фронт-юниты в `frontend/src/**` — сейчас их ноль |
| 134 | Интеграционные на Python |
| 135 | Нагрузочные на Locust — новый слой |
| 136 | CI и удаление старого фреймворка |

Замеры на 2026-07-27: 515 Go-файлов в `backend/testframework/`, 0 юнитов рядом с
кодом, 0 юнитов во фронте, покрытие 81.9%.

### Трек 2 — вынос монолита в микросервисы

[Документ](.planning/active/august-2026/01-MICROSERVICES-SPLIT.md) · основание:
[PLAN-team-lead-opus.md](.planning/decisions/ab-microservices/PLAN-team-lead-opus.md)

У каждого сервиса — **свой бэкенд и свой фронтенд**. 34 домена группируются в
**8 bounded contexts**: notification, ai, realtime, tracker, identity, business,
integrations, platform.

**Swagger остаётся единый**, разделённый на зоны по сервисам — всё расписано,
подписано и с примерами. Требование к детализации жёсткое: у каждой ручки полные
схемы запроса и ответа, единая форма `AppError`, пагинация, требуемая роль, и
ответы проработаны **по корнер- и эдж-кейсам** — пустой результат, чужой ресурс,
конкурентная правка, повторный вызов, границы пагинации, частичный успех,
лимиты, отказ внешней зависимости. Для каждого кода описано, **когда именно** он
возвращается. Ручка без описанной схемы ответа не проходит ревью; расхождение
спеки с реализацией валит CI.

| Этап | Что |
|---|---|
| Э1 | Modular Monolith Hardening — запрет кросс-импортов |
| Э2 | Schema-per-service внутри текущего PostgreSQL |
| Э3 | `hub-shell` + Module Federation, первый фронт-модуль |
| Э4 | Пилот `hub-notification` — вертикальный срез целиком |
| **Э5** | **GATE Go/No-Go по результату пилота** |
| Э6–Э7 | `hub-ai`, затем `hub-realtime` |

Раскладка по контейнерам, схема БД, границы сервисов и транспорт — **решения
архитектора**, 7 открытых вопросов в §8 документа.

### Жёсткая зависимость

Фаза 133 обязательна до этапа Э3. Во фронте сейчас **ноль юнит-тестов** — резать
единый SPA на микрофронтенды без них означает регрессии без способа их заметить.

### Открытые решения, блокирующие август

| Что | Кому |
|---|---|
| 7 вопросов по архитектуре сервисов | архитектору |
| [14 решений серых зон, включая GDPR](.planning/audits/planning-audit-2026-07/DECISIONS-PENDING.md) | владельцу |
| [Судьба модуля Wiki](.planning/audits/planning-audit-2026-07/WIKI-SCOPE-OPTIONS.md) — строить / вынести / удалить | владельцу |
| [NFR-24 и NFR-11](.planning/audits/planning-audit-2026-07/NFR-24-VERSION-CLASSIFICATION.md) — раскатывать или сузить формулировку | владельцу |
| PG-03 ротация секретов — блокирует прод-катку | ops |

---

<div align="center">

**Hub — self-hosted super-app для dev-команд**

Go · React · PostgreSQL · MongoDB · Redis · NATS · LiveKit · Meilisearch

*Обновлено: 2026-07-28 (планы на август)*

</div>
