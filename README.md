# 🚀 Hub — Единая платформа для команд разработки

<div align="center">

**Трекер задач · Мессенджер · TMS · Звонки · Дашборды**

*Всё в одном месте — без переключения между десятком инструментов*

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![React](https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=black)](https://react.dev)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.8-3178C6?logo=typescript&logoColor=white)](https://typescriptlang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-4169E1?logo=postgresql&logoColor=white)](https://postgresql.org)
[![MongoDB](https://img.shields.io/badge/MongoDB-8-47A248?logo=mongodb&logoColor=white)](https://mongodb.com)
[![License](https://img.shields.io/badge/License-Private-red)](#)

</div>

---

## Что это?

Hub — платформа для средних команд разработки (10-50 человек), объединяющая:

| Модуль | Заменяет | Описание |
|--------|----------|----------|
| **Трекер задач** | Jira | 9 типов задач, Kanban + Scrum, custom workflow, WIP-лимиты, swimlanes |
| **Мессенджер** | Slack / Telegram | Каналы, треды, реакции, голосовые сообщения, присутствие |
| **TMS** | TestRail / Allure TestOps | Тест-кейсы (classic + BDD), прогоны, Allure интеграция, отчёты |
| **Звонки** | Google Meet / Zoom | Аудио + screen share, до 50 участников, запись, планирование |
| **Дашборды** | Excel / Grafana | Персональный, проектный, админский + drag-and-drop виджеты |

## Ключевые фичи

- 🔐 **Авторизация без паролей** — Magic Link (email) + Telegram
- 📋 **Полноценный трекер** — 9 типов задач, Kanban, Scrum, спринты, Planning Poker
- 💬 **Мессенджер** — каналы, DM, треды, реакции, голосовые, поиск, закладки
- 🧪 **TMS** — тест-кейсы, прогоны, Allure импорт, матрица трассировки, отчёты
- 📞 **Звонки** — LiveKit, screen share, запись, запланированные, эфемерный чат
- 📊 **Дашборды** — 12+ виджетов, burndown, velocity, настраиваемые шаблоны
- 🔔 **Уведомления** — в приложении + Telegram бот, тихие часы
- 🔍 **Глобальный поиск** — Ctrl+K, JQL запросы, сохранённые фильтры
- 🐙 **GitHub интеграция** — вебхуки, автолинк PR/коммитов, создание веток
- 🌐 **i18n** — Русский + Английский
- 🎨 **Темы** — Тёмная / Светлая / Системная + акцентные цвета
- 📱 **PWA** — установка на телефон
- 🔑 **RBAC** — фиксированные + кастомные роли с детальными правами
- 📦 **Импорт** — Jira JSON, TestRail CSV/XML, CSV/Excel
- 📤 **Экспорт** — CSV, Excel, PDF, ZIP-бэкап

## Технологический стек

### Backend
| Технология | Назначение |
|-----------|------------|
| **Go + Fiber v3** | HTTP сервер, WebSocket, middleware |
| **Ent ORM + Atlas** | PostgreSQL, кодогенерация, типобезопасность |
| **MongoDB** | Сообщения чата, Allure отчёты |
| **Redis** | Кеш, сессии, presence, magic link токены |
| **NATS + JetStream** | Pub/sub, очереди, уведомления, аудит |
| **LiveKit** | WebRTC звонки, screen share, запись |
| **SeaweedFS** | S3-совместимое файловое хранилище |

### Frontend
| Технология | Назначение |
|-----------|------------|
| **React 19 + TypeScript** | UI фреймворк |
| **TanStack Router + Query** | Type-safe роутинг, кеширование данных |
| **Zustand** | State management (11 stores) |
| **Shadcn/ui + Tailwind CSS 4** | UI компоненты, стилизация |
| **Plate** | Rich-text редактор (описания, комменты) |
| **LiveKit React** | UI звонков |
| **Recharts** | Графики (burndown, velocity, тренды) |
| **Vite** | Сборка, code splitting, HMR |

## Быстрый старт

### Требования
- Docker + Docker Compose
- Git

### Запуск (development)
```bash
git clone https://github.com/Valerijkk/hub.git
cd hub
docker compose -f docker-compose.local.yml up -d
```

Frontend: http://localhost:5173
Backend: http://localhost:3000

### Запуск (production)
```bash
docker compose -f docker-compose.demo.yml up -d --build
```

## Архитектура

```
┌─────────────────────────────────────────────────┐
│                    Caddy / Nginx                │
│              (reverse proxy, HTTPS)             │
├────────────────────┬────────────────────────────┤
│                    │                            │
│   Frontend (SPA)   │   Backend (Go + Fiber)     │
│   React + Vite     │   REST API + WebSocket     │
│                    │                            │
├────────────────────┴────────────────────────────┤
│                                                 │
│  ┌──────────┐ ┌───────┐ ┌──────┐ ┌──────────┐  │
│  │PostgreSQL│ │MongoDB│ │Redis │ │   NATS   │  │
│  │ (задачи, │ │(чаты, │ │(кеш, │ │(события, │  │
│  │  TMS,    │ │Allure)│ │сессии│ │ очереди) │  │
│  │  юзеры)  │ │       │ │      │ │          │  │
│  └──────────┘ └───────┘ └──────┘ └──────────┘  │
│                                                 │
│  ┌──────────┐ ┌──────────────────┐              │
│  │SeaweedFS │ │     LiveKit      │              │
│  │ (файлы)  │ │ (звонки, WebRTC) │              │
│  └──────────┘ └──────────────────┘              │
└─────────────────────────────────────────────────┘
```

## Статистика кодовой базы

| Метрика | Значение |
|---------|----------|
| API endpoints | **~230** |
| Frontend routes | **39** |
| Requirements (v1) | **95/95 реализовано** |
| Go файлов | ~608 |
| TS/TSX файлов | ~210 |
| Backend тесты | Tracker, Auth, Sprint, TMS |
| Bundle size (gzip) | **245 KB** main chunk |

## Структура проекта

```
hub/
├── backend/           # Go backend
│   ├── cmd/server/    # Точка входа
│   ├── ent/           # Ent ORM схемы + кодогенерация
│   └── internal/      # Бизнес-логика по модулям
│       ├── auth/      # Аутентификация (Magic Link + Telegram)
│       ├── project/   # Проекты, команды, приглашения
│       ├── tracker/   # Трекер задач, workflow, Kanban
│       ├── sprint/    # Спринты, Planning Poker
│       ├── messenger/ # Мессенджер, каналы, presence
│       ├── tms/       # TMS, тест-кейсы, прогоны, Allure
│       ├── calls/     # Звонки, запись, планирование
│       ├── dashboard/ # Дашборды, виджеты
│       ├── notification/ # Уведомления, Telegram бот
│       ├── search/    # Глобальный поиск, JQL, фильтры
│       ├── github/    # GitHub интеграция, вебхуки
│       ├── ws/        # WebSocket gateway, NATS bridge
│       └── ...
├── frontend/          # React frontend
│   └── src/
│       ├── api/       # API клиент (16 модулей, 192 вызова)
│       ├── components/# UI компоненты
│       ├── hooks/     # React хуки (7 шт.)
│       ├── modules/   # Модули (tracker, tms, calls)
│       ├── routes/    # TanStack Router (39 маршрутов)
│       └── stores/    # Zustand stores (11 шт.)
├── cli/               # hub-cli (CI/CD утилита)
├── deployments/       # Docker, K3s конфиги
└── .planning/         # Roadmap, требования, планы
```

## Milestone Roadmap

| Milestone | Статус | Описание |
|-----------|--------|----------|
| **v1.0** | ✅ Завершён | 95 требований, 7 фаз, все модули |
| **v2.0** | 📋 Спланирован | 137 требований — UI polish, DevOps, GitHub App, AI, Productivity |
| **v3.0** | 💡 Идея | AI Chat Hub — GigaChat, YandexGPT, DeepSeek без VPN |

## Лицензия

Приватный проект. Все права защищены.

---

<div align="center">

**Разработано с ❤️ и 🤖**

*Go + React + много кофе*

</div>
