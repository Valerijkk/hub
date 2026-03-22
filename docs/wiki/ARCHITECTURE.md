# Архитектура Hub

## Обзор

Hub построен как модульный монолит на Go с React SPA фронтендом. Модули общаются через NATS (pub/sub) и общую PostgreSQL базу (через Ent ORM).

## Модули

```
┌────────────────────────────────────────────────────────┐
│                   Frontend (React SPA)                 │
│  TanStack Router · TanStack Query · Zustand · Plate   │
├────────────────────────────────────────────────────────┤
│              Nginx (reverse proxy, SPA)                │
├────────────────────────────────────────────────────────┤
│                Go Backend (Fiber v3)                   │
│                                                        │
│  ┌──────┐ ┌────────┐ ┌──────────┐ ┌──────┐ ┌───────┐ │
│  │ Auth │ │Tracker │ │Messenger │ │ TMS  │ │ Calls │ │
│  └──┬───┘ └───┬────┘ └────┬─────┘ └──┬───┘ └───┬───┘ │
│     │         │           │          │         │      │
│  ┌──┴─────────┴───────────┴──────────┴─────────┴──┐   │
│  │          Platform Layer (shared)               │   │
│  │  Config · Errors · HTTP · Middleware · Outbox   │   │
│  └────────────────────────────────────────────────┘   │
├────────────────────────────────────────────────────────┤
│                  Infrastructure                        │
│  PostgreSQL · MongoDB · Redis · NATS · SeaweedFS       │
│                    LiveKit                             │
└────────────────────────────────────────────────────────┘
```

## Потоки данных

### HTTP запрос
```
Browser → Nginx → Fiber Middleware → Handler → Service → Repository → DB
                  (RequestID, Logger, CORS, Auth, RBAC, RateLimit)
```

### Real-time (WebSocket)
```
Event (task created) → Outbox (PG) → Worker → NATS → WS Bridge → Client
```

### Уведомления
```
NATS Event → JetStream Consumer → NotificationService → In-App + Telegram
```

## База данных

| БД | Данные | Причина выбора |
|----|--------|---------------|
| **PostgreSQL** | Пользователи, проекты, задачи, TMS, workflow | Реляционные связи, Ent ORM |
| **MongoDB** | Сообщения чата, Allure отчёты, эфемерный чат | Документная модель, гибкая схема |
| **Redis** | Кеш RBAC, сессии, magic link токены, presence | Быстрый доступ, TTL |

## Аутентификация

```
JWT Access Token (15 мин, в памяти)
  + Refresh Token (30 дней, в Redis)
  + Session (в PostgreSQL)
```

При каждом запросе middleware проверяет JWT, извлекает `user_id`, проверяет что пользователь не деактивирован.

## RBAC

Три уровня прав:
1. **Системные роли** — Admin, Member
2. **Проектные роли** — Owner, Developer, Tester, Viewer
3. **Кастомные роли** — создаются Owner, с детальными permission bits

Права кешируются в Redis, инвалидируются при любом изменении membership.
