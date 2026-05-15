<div align="center">

# Hub

**Единая платформа для команд разработки**

Трекер задач · Мессенджер · TMS · Звонки · Дашборды

<p>
  <img src="https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/React-19-61DAFB?style=flat&logo=react&logoColor=black" alt="React" />
  <img src="https://img.shields.io/badge/TypeScript-5.7-3178C6?style=flat&logo=typescript&logoColor=white" alt="TypeScript" />
  <img src="https://img.shields.io/badge/PostgreSQL-17-4169E1?style=flat&logo=postgresql&logoColor=white" alt="PostgreSQL" />
  <img src="https://img.shields.io/badge/MongoDB-8-47A248?style=flat&logo=mongodb&logoColor=white" alt="MongoDB" />
  <img src="https://img.shields.io/badge/Redis-7-DC382D?style=flat&logo=redis&logoColor=white" alt="Redis" />
  <img src="https://img.shields.io/badge/License-Private-red" alt="License" />
</p>

<p>
  <strong>~230 API endpoints</strong> · <strong>39 frontend routes</strong> · <strong>95 requirements v1.0</strong>
</p>

</div>

---

Hub -- платформа, заменяющая Jira + Slack + TestRail + Zoom + Grafana для средних команд разработки (10-50 человек). Один проект вместо пяти подписок. Полноценный трекер с 9 типами задач и кастомным workflow, мессенджер с тредами и голосовыми, TMS с Allure-интеграцией, видеозвонки через LiveKit, и настраиваемые дашборды -- всё под одной авторизацией.

## Возможности

### Трекер задач
- 9 типов задач (Epic, Story, Task, Bug, Sub-Task, Improvement, Research, Design, DevOps)
- Kanban-доски с WIP-лимитами и swimlanes
- Scrum: спринты, Planning Poker, burndown, velocity
- Кастомный workflow (статусы + переходы)
- Кастомные поля, метки, группы меток
- Связи между задачами (blocks, relates to, duplicates)
- Time tracking с таймером
- JQL-поиск по задачам
- Комментарии с rich-text, реакции, лента активности

### Мессенджер
- Каналы (публичные / приватные), DM
- Треды, реакции, пересылка сообщений
- Закрепление сообщений, закладки
- Голосовые сообщения
- Опросы (polls)
- Индикатор набора текста, присутствие (online/offline)
- Кастомные статусы
- Пригласительные ссылки в каналы
- Глобальный поиск по сообщениям

### TMS (Test Management)
- Тест-кейсы: classic и BDD (Gherkin)
- Тестовые наборы (suites) и компоненты
- Тестовые прогоны с назначением тестировщиков
- Результаты: passed / failed / blocked / skipped
- Версионирование тест-кейсов с rollback
- Связывание тест-кейсов с задачами (traceability)
- Создание бага из результата прогона
- Allure import (JSON / ZIP), JUnit import
- CI/CD webhook (автоматические прогоны)
- Отчёты: тренды, flaky-тесты, покрытие, сравнение прогонов, матрица трассировки, дефект-категории, сравнение окружений
- Тестовые окружения
- Экспорт прогонов в PDF

### Звонки
- Аудио + screen share через LiveKit (WebRTC)
- До 50 участников
- Запись звонков
- Планирование звонков, проверка конфликтов
- Instant-звонки из мессенджера
- Приглашение по ссылке
- "Поднять руку" (raise hand)
- Эфемерный чат в звонке

### Дашборды
- Персональный, проектный, админский
- 12+ виджетов (burndown, velocity, тренды тестов и др.)
- Drag-and-drop layout
- Шаблоны дашбордов

### Дополнительно
- **Авторизация**: Magic Link (email) + Telegram, JWT + refresh-токены, управление сессиями
- **RBAC**: фиксированные роли (Owner, Admin, Member, Viewer) + кастомные роли с детальными правами
- **Уведомления**: в приложении + push (Web Push VAPID) + Telegram-бот, тихие часы, настройки по типам
- **Глобальный поиск**: Ctrl+K, поиск по задачам / тест-кейсам / сообщениям, сохранённые фильтры
- **GitHub-интеграция**: вебхуки (push, PR, workflow), автолинк PR и коммитов, создание веток из задач, CI/CD статусы
- **GitLab-интеграция**: вебхуки, пайплайны, создание веток, триггер пайплайнов
- **Импорт**: Jira JSON, CSV/Excel (задачи), TestRail CSV (тест-кейсы)
- **Экспорт**: CSV, Excel, ZIP (полный бэкап проекта), PDF (прогоны, спринты)
- **Бэкапы**: ручное создание, восстановление, удаление
- **API-ключи и вебхуки**: программный доступ + webhook-подписки с историей доставок
- **Аудит**: лог действий, лимиты ресурсов, корзина, деактивация и GDPR-удаление пользователей
- **i18n**: Русский + Английский
- **Темы**: Тёмная / Светлая / Системная + акцентные цвета
- **PWA**: установка на мобильные устройства

## Технологический стек

### Backend

| Технология | Версия | Назначение |
|---|---|---|
| Go + Fiber v3 | 1.24 | HTTP-сервер, WebSocket, middleware, конкуррентная обработка |
| Ent ORM + Atlas | -- | PostgreSQL: кодогенерация схем, типобезопасные запросы, миграции |
| MongoDB | 8 | Сообщения мессенджера, Allure-отчёты |
| Redis | 7 | Кеш, сессии, presence, magic link токены, rate limiting |
| NATS + JetStream | -- | Pub/sub события, очереди уведомлений, аудит-лог |
| LiveKit | -- | WebRTC: звонки, screen share, запись |
| SeaweedFS | -- | S3-совместимое файловое хранилище (аватары, вложения, записи) |

### Frontend

| Технология | Версия | Назначение |
|---|---|---|
| React | 19 | UI-фреймворк |
| TypeScript | 5.7 | Типобезопасность |
| TanStack Router | -- | Type-safe файловый роутинг (39 маршрутов) |
| TanStack Query | -- | Серверное состояние, кеширование, инвалидация |
| Zustand | -- | Клиентское состояние (11 stores) |
| Shadcn/ui + Tailwind CSS 4 | -- | UI-компоненты, дизайн-система |
| Plate | -- | Rich-text редактор (описания задач, комментарии) |
| Recharts | -- | Графики (burndown, velocity, тренды) |
| LiveKit React | -- | UI звонков |
| Vite | -- | Сборка, HMR, code splitting |

### Инфраструктура

| Технология | Назначение |
|---|---|
| Docker + Docker Compose | Контейнеризация, локальная разработка |
| Caddy | Reverse proxy, автоматический HTTPS |
| K3s | Kubernetes-деплой (production) |
| GitHub Actions | CI/CD |

## Архитектура

```
                          ┌──────────────────────┐
                          │    Caddy / Nginx     │
                          │  (reverse proxy, TLS)│
                          └──────┬───────┬───────┘
                                 │       │
                    ┌────────────┘       └────────────┐
                    v                                  v
          ┌─────────────────┐              ┌─────────────────────┐
          │  Frontend (SPA) │              │  Backend (Go+Fiber) │
          │  React 19       │   REST/WS    │  ~230 endpoints     │
          │  39 routes      │ ◄──────────► │  WebSocket gateway  │
          │  Vite           │              │  NATS consumers      │
          └─────────────────┘              └──────────┬──────────┘
                                                      │
                    ┌────────────┬────────────┬───────┴───────┐
                    v            v            v               v
             ┌───────────┐ ┌─────────┐ ┌──────────┐  ┌────────────┐
             │PostgreSQL │ │ MongoDB │ │  Redis   │  │    NATS    │
             │  17       │ │  8      │ │  7       │  │ JetStream  │
             │           │ │         │ │          │  │            │
             │ Задачи    │ │ Чат     │ │ Кеш      │  │ События    │
             │ TMS       │ │ Allure  │ │ Сессии   │  │ Очереди    │
             │ Юзеры     │ │         │ │ Presence │  │ Аудит      │
             │ Спринты   │ │         │ │          │  │            │
             └───────────┘ └─────────┘ └──────────┘  └────────────┘
                                │
                    ┌───────────┴───────────┐
                    v                       v
             ┌───────────┐          ┌────────────┐
             │ SeaweedFS │          │  LiveKit   │
             │ (файлы)   │          │ (WebRTC)   │
             └───────────┘          └────────────┘
```

## Быстрый старт

### Требования

- Docker + Docker Compose v2
- Git

### Локальная разработка

```bash
# Клонировать репозиторий
git clone https://github.com/Valerijkk/hub.git
cd hub

# Запустить все сервисы
docker compose -f docker-compose.local.yml up -d
```

| Сервис | URL |
|---|---|
| Frontend | http://localhost:5173 |
| Backend API | http://localhost:3000/api |
| SeaweedFS | http://localhost:9333 |

### Production

```bash
cd deployments
docker compose -f docker-compose.prod.yml --env-file .env up -d
```

## Структура проекта

```
hub/
├── backend/                  # Go backend
│   ├── cmd/server/           # Точка входа (main.go)
│   ├── ent/                  # Ent ORM: схемы, кодогенерация, миграции
│   │   └── schema/           # ~30 схем (User, Project, Issue, TestCase, ...)
│   └── internal/             # Бизнес-логика по модулям
│       ├── auth/             # Magic Link + Telegram, JWT, сессии
│       ├── project/          # Проекты, участники, приглашения, команды, шаблоны
│       ├── tracker/          # Трекер: задачи, workflow, доски, метки, поля, время
│       ├── sprint/           # Спринты, burndown, velocity, Planning Poker
│       ├── messenger/        # Каналы, сообщения, присутствие, опросы, инвайты
│       ├── tms/              # TMS: кейсы, наборы, прогоны, Allure, отчёты
│       ├── calls/            # Звонки: LiveKit, запись, планирование
│       ├── dashboard/        # Дашборды: данные, layout, шаблоны
│       ├── notification/     # Уведомления: in-app, push, Telegram-бот
│       ├── search/           # Глобальный поиск, фильтры
│       ├── github/           # GitHub: вебхуки, интеграции, CI/CD, превью ссылок
│       ├── gitlab/           # GitLab: вебхуки, пайплайны, интеграции
│       ├── rbac/             # RBAC: роли, права, кастомные роли
│       ├── storage/          # Файлы: upload/download URL, SeaweedFS
│       ├── imex/             # Импорт/экспорт: CSV, Excel, Jira, PDF
│       ├── backup/           # Бэкапы: создание, восстановление
│       ├── audit/            # Аудит: лог, лимиты, корзина, GDPR
│       ├── apigateway/       # API-ключи, webhook-подписки
│       ├── ws/               # WebSocket gateway, NATS bridge
│       └── platform/         # Общее: middleware, утилиты, rate limiting
├── frontend/                 # React frontend
│   └── src/
│       ├── api/              # API-клиент (16 модулей, ~192 вызова)
│       ├── components/       # Переиспользуемые UI-компоненты
│       ├── hooks/            # React-хуки (7 шт.)
│       ├── modules/          # Крупные модули (tracker, tms, calls, messenger)
│       ├── routes/           # TanStack Router (39 маршрутов)
│       └── stores/           # Zustand stores (11 шт.)
├── cli/                      # hub-cli (CI/CD утилита)
├── deployments/              # Docker, K3s конфиги
├── docs/wiki/                # Документация (API, архитектура, деплой, шорткаты)
└── docker-compose.*.yml      # Конфигурации Docker Compose
```

## API

Hub предоставляет **~230 REST endpoints**, сгруппированных по модулям:

| Модуль | Endpoints | Описание |
|---|---|---|
| Auth | 13 | Аутентификация, профиль, сессии, регистрация |
| Projects | 22 | CRUD проектов, участники, приглашения, команды |
| Tracker | 35 | Задачи, workflow, доски, метки, поля, время, комментарии |
| Sprint | 14 | Спринты, burndown, velocity, Planning Poker |
| Messenger | 40 | Каналы, сообщения, присутствие, опросы, инвайты |
| TMS | 46 | Тест-кейсы, прогоны, Allure, отчёты, окружения |
| Calls | 22 | Звонки, запись, планирование, чат |
| Dashboard | 7 | Данные дашбордов, layout, шаблоны |
| Notifications | 10 | Список, статусы, настройки, push-подписки |
| Search | 5 | Глобальный поиск, сохранённые фильтры |
| GitHub | 9 | Вебхуки, интеграции, CI/CD, создание веток, превью |
| GitLab | 7 | Вебхуки, интеграции, пайплайны, создание веток |
| Storage | 4 | Upload/download URL, подтверждение, удаление |
| Import/Export | 6 | CSV, Excel, Jira, PDF, ZIP |
| Backup | 4 | Создание, список, восстановление, удаление |
| API Gateway | 8 | API-ключи, webhook-подписки, доставки |
| RBAC | 6 | Кастомные роли, назначение |
| Admin/Audit | 8 | Аудит, лимиты, корзина, GDPR |

Полная документация: [docs/wiki/API.md](docs/wiki/API.md)

## Статистика кодовой базы

| Метрика | Значение |
|---|---|
| API endpoints | **~230** |
| Frontend routes | **39** |
| Zustand stores | **11** |
| API-клиент модули | **16** (~192 вызова) |
| Go-файлов | **~608** |
| TS/TSX-файлов | **~210** |
| Ent-схемы | **~30** |
| Backend-тесты | Tracker, Auth, Sprint, TMS |
| Bundle size (gzip) | **245 KB** main chunk |

## Дорожная карта

| Версия | Статус | Описание |
|---|---|---|
| **v1.0** | Завершён | 95 требований, 7 фаз. Все основные модули: трекер, мессенджер, TMS, звонки, дашборды, GitHub, RBAC, уведомления, поиск, импорт/экспорт |
| **v2.0** | Спланирован | 137 требований -- 6 фаз. UI-редизайн (Slack-стиль), DevOps (monitoring, K3s), GitHub App, AI-фичи, Productivity (templates, shortcuts, offline) |
| **v3.0** | Концепция | AI Chat Hub -- GigaChat, YandexGPT, DeepSeek (без VPN). AI-помощник в трекере и TMS |

## Контрибуция

См. [CONTRIBUTING.md](CONTRIBUTING.md) для деталей по:
- Git workflow и ветвлению
- Формату коммитов
- Code style (Go, TypeScript)
- Процессу code review

## Лицензия

Приватный проект. Все права защищены.

---

<div align="center">

**Разработано с вниманием к деталям**

Go + React + PostgreSQL + MongoDB + Redis + NATS + LiveKit

</div>
