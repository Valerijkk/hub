# Быстрый старт

> **Обновлено:** 2026-06-21 · Стек: Go 1.25, React 19, Node 22

## Требования для локальной разработки

| Компонент | Версия | Проверка |
|-----------|--------|----------|
| Go | 1.25+ | `go version` |
| Node.js | 22 (CI) / 20+ local | `node -v` |
| Docker | Compose v2 | `docker compose version` |

Полный стек в Docker (рекомендуется):

```bash
cp .env.local.example .env.local   # опционально: Resend, Telegram notify
docker compose --env-file .env.local -f docker-compose.local.yml up --build
# UI: http://localhost:8080
# Демо-данные (только DEV_MODE):
curl -X POST http://localhost:8080/api/dev/seed
```

Нативный backend + Docker infra — см. [DEPLOY.md](DEPLOY.md) § «Локальная разработка».

---

## Вход в систему

**Единственный способ входа** — Magic Link по email (без паролей, без SSO).

### Magic Link

1. Откройте `/login`
2. Введите корпоративный email
3. Получите **6-значный код** на почту (Resend / hub-mail)
4. Введите код на странице верификации — вы в системе

В dev-режиме (`DEV_MODE=true`) доступен **dev-login** для E2E и локальной отладки без почты.

### Telegram

Telegram **не используется для входа**. Только:

- привязка аккаунта для уведомлений (`Настройки → Уведомления → Telegram`);
- доставка push в бот.

---

## Первые шаги после входа

### 1. Онбординг (3 шага)

| Шаг | Содержание |
|-----|------------|
| Профиль | Имя, должность, отдел, часовой пояс |
| Обзор | Краткий тур по модулям Hub |
| Настройки | Тема (тёмная/светлая), язык RU/EN, первый проект |

Завершение: `POST /api/auth/me/onboarding`.

### 2. Создание проекта

1. Сайдбар → **«Новый проект»**
2. Название + ключ (например `HUB`, 2–10 символов, uppercase)
3. Автоматически создаются: workflow, Kanban-доска, канал `#project-hub`, TMS-структура

### 3. Первая задача

1. Проект → **Доска**
2. `Ctrl+N` или кнопка **«+»**
3. Тип: Bug, Task, Story, Epic и др.
4. Drag & drop между колонками Kanban

### 4. Мессенджер

1. Сайдбар → **Мессенджер**
2. Канал проекта уже создан
3. Сообщения, треды, реакции, @mentions

### 5. Глобальный поиск

`Ctrl+K` / `Cmd+K` — задачи, кейсы, пользователи, проекты, каналы, файлы.

Подробнее: [SHORTCUTS.md](SHORTCUTS.md).

---

## Минимальная регрессия (перед коммитом)

```bash
# Backend
cd backend
go generate ./ent
go build -tags release ./...
go test -tags hubtest -race -count=1 -timeout=600s ./internal/... ./testframework/...

# Frontend
cd frontend
npx tsc --noEmit
npm run lint          # Biome
npx vitest run
```

Полный gate: skill `/qa-gate` (см. `docs/test_plan.md`).

---

## Полезные ссылки

| Документ | Содержание |
|----------|------------|
| [API.md](API.md) | REST API справочник |
| [DEPLOY.md](DEPLOY.md) | Prod/staging, CI/CD |
| [SHORTCUTS.md](SHORTCUTS.md) | Горячие клавиши |
| `docs/vision_and_scope.md` | User Stories, KPI |
| `docs/system_specification.md` | FR/NFR, контракты |

---
*2026-06-21 — AUTH-01: Magic Link only, Telegram notify-only*
