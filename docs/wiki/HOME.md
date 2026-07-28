# Hub Wiki

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

## Статус проекта

| Milestone | Статус |
|-----------|--------|
| v1.0 Foundation | ✅ Shipped 2026-03-21 |
| v2.0 Hub Redesign | ✅ 2026-05-15 |
| v2.6 Architecture + OBS | ✅ 2026-06-24 |
| v3.0 AI Native | ✅ 2026-06-22 |
| v4 Super-app | Archived (phases 39–58) |
| Coverage floor | **82%+** `internal/` (Phase 120+) |

Актуальный статус: `.planning/STATE.md`.

---
*2026-06-21 — wiki index, broken links removed*
