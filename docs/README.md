# Hub — Документация

**Hub** — единая платформа для команд разработки и тестирования, объединяющая трекер задач, мессенджер, TMS, созвоны и дашборды в одном интерфейсе.

---

## Навигация

| Документ | Описание |
|----------|----------|
| [BUSINESS.md](./BUSINESS.md) | Бизнес-документация: продукт, модули, роли, сценарии, дорожная карта |
| [TECHNICAL.md](./TECHNICAL.md) | Техническая документация: архитектура, БД, API, WebSocket, деплой |
| [TESTING.md](./TESTING.md) | Тестовая документация: стратегия, тест-кейсы, покрытие, результаты |

## Быстрый старт

```bash
# Запуск инфраструктуры
cd deployments && docker compose up -d

# Запуск бэкенда
cd backend && go run ./cmd/server

# Запуск фронтенда
cd frontend && npm run dev
```

Откройте http://localhost:80 (через Caddy) или http://localhost:5173 (Vite dev server).

## Стек технологий

| Слой | Технологии |
|------|-----------|
| Backend | Go 1.23, Fiber v3, Ent ORM, Atlas миграции |
| Frontend | React 19, TypeScript, Vite, TanStack Router/Query |
| Базы данных | PostgreSQL 17, MongoDB 8, Redis 7.4 |
| Real-time | NATS 2.11 + JetStream, WebSocket (fasthttp) |
| Созвоны | LiveKit 1.8 (WebRTC, голос + screen share) |
| Хранилище | SeaweedFS (S3-compatible) |
| UI | Shadcn/ui, Tailwind CSS 4, Plate (rich-text), Tremor (графики) |
| Прокси | Caddy 2 (auto-HTTPS, WebSocket) |

---

*Последнее обновление: 2026-03-28*
