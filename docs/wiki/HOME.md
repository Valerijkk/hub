# Hub Wiki

> **Текущая product-truth (founder, 2026-08-12):** Hub остаётся
> незавершённым MVP. Ни один продуктовый модуль не принят как `complete`,
> `stable` или `production-ready`. Метки `SHIPPED`/`COMPLETE`, milestone/audit
> `PASS` и галочки ниже — только историческое planning/technical evidence, не
> product/module acceptance. PG-03/FA-06 — отдельный manual ops-блокер
> production cutover, а не индикатор готовности MVP.

Добро пожаловать в wiki Hub — единой платформы для команд разработки (10–50 человек): Task Tracker, Messenger, TMS, Voice Calls, Dashboards и корпоративные модули v4.

> **Индекс всей базы знаний:** [docs/README.md](../README.md)

---

## Навигация wiki

### Для пользователей

| Страница | Описание |
|----------|----------|
| [Быстрый старт](QUICKSTART.md) | Первый вход, онбординг, первый проект |
| [Горячие клавиши](SHORTCUTS.md) | `Ctrl+K`, `Ctrl+N`, мессенджер, редактор |

Модули (детали в `docs/vision_and_scope.md` и UI):

- **Трекер** — Kanban, Scrum, JQL, time tracking
- **Мессенджер** — каналы, DM, треды, реакции
- **TMS** — кейсы, прогоны, Allure import
- **Звонки** — LiveKit, screen share, запись
- **Дашборды** — личный, проектный, админский
- **v4 super-app** — intranet, calendar, CRM-lite, HR, accounting, bases
- **Wiki (продукт, spaces/pages)** — **planned** (Волны 5/8); G-07 = Partial — см. `docs/vision_and_scope.md`. Эта страница (`docs/wiki/`) — документация репозитория, не продуктовый wiki-модуль.

Ветка calls-feature-codex,2026-09-06: поиск с отдельными error/empty/retry,
защита удаления от повторов/устаревших ответов и передача действий в треды
приняты на unit-уровне (TypeScript +728 Docker tests PASS).
Это не готовность мессенджера «как Slack». Jitsi остаётся неактивированным,
реальные два пользователя/медиа и backend acceptance ещё не приняты.
[Точный статус и ограничения](../superpowers/plans/2026-09-05-communications-completion-queue.md).

### Для администраторов

| Страница | Описание |
|----------|----------|
| [Установка и деплой](DEPLOY.md) | Docker Compose, CI/CD, переменные окружения |
| [Архитектура](ARCHITECTURE.md) | Модули backend, потоки данных |

Операции (runbook в `deployments/`):

- RBAC и роли — `docs/system_specification.md` § RBAC
- Бэкапы — `POST /api/admin/backups`, `deployments/.env.example`
- Секреты — GitHub Secrets, ротация PG-03

### Для разработчиков

| Страница | Описание |
|----------|----------|
| [API Reference](API.md) | REST API (~252 OpenAPI paths; route inventory ~286 — см. API.md) |
| [Архитектура](ARCHITECTURE.md) | C4, NATS, WebSocket bridge |
| `docs/TECHNICAL.md` | ERD, полные контракты, env vars |
| `docs/test_plan.md` | Стратегия тестирования, CI jobs |
| `docs/code_style_guide.md` | Go/TS стиль |
| `docs/security_guidelines.md` | OWASP, security checklist |
| `docs/tz_constraints.md` | Mandatory/forbidden TZ |

---

## Стек (актуальные версии)

| Слой | Технология | Версия |
|------|------------|--------|
| Backend | Go + Fiber v3 | Go **1.25**, Fiber **3.0.0-beta.3** |
| Frontend | React + Vite | React **19.1**, TS **5.7**, Vite **6** |
| CI runtime | Node | **22** |
| PostgreSQL | relational | **17** |
| MongoDB | messages, Allure | **8.0** |
| Redis | sessions, refresh | **7.4-alpine** |
| NATS JetStream | pub/sub | **2.11** |
| SeaweedFS | S3 files | **3.97** |
| LiveKit | WebRTC | **v1.8** |
| Caddy | reverse proxy | **2-alpine** |

Источник: `backend/go.mod`, `frontend/package.json`, `deployments/docker-compose.prod.yml`.

---

## Текущая product-truth и исторические milestone labels

Hub — незавершённый MVP; таблица ниже не подтверждает готовность ни одного
модуля.

| Milestone | Статус |
|-----------|--------|
| v1.0 Foundation | historical SHIPPED label · 2026-03-21 |
| v2.0 Hub Redesign | historical technical label · 2026-05-15 |
| v2.6 Architecture + OBS | historical technical label · 2026-06-24 |
| v3.0 AI Native | historical technical label · 2026-06-22 |
| v4 Super-app | historical archived phases 39–58; not module acceptance |
| Coverage floor | historical **82%+** `internal/` baseline (Phase 120+); not readiness |

Актуальный статус: `.planning/STATE.md`.

---
*2026-08-12 — founder product-truth: unfinished MVP; zero accepted complete/stable modules; PG-03/FA-06 is a separate ops cutover blocker.*
*2026-07-28 — TASK-105: уточнено, что product Wiki planned (G-07 Partial); `docs/wiki/` ≠ Confluence-замена*
*2026-06-21 — wiki index, broken links removed*
