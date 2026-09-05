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
docker compose -f docker-compose.local.yml up -d
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

Уже аутентифицированный пользователь на `/login` уходит в `/projects` (или `/onboarding`, если онбординг не завершён). `/onboarding` без сессии редиректит на `/login`; при `onboarding_completed` — на `/projects` (`beforeLoad` в stub-роутах).

В dev-режиме (`DEV_MODE=true`) доступен **dev-login** для E2E и локальной отладки без почты. Для magic-link только этот явный флаг разрешает вернуть одноразовый код в HTTP-ответе; OTP никогда не записывается в логи. Пустой `RESEND_API_KEY` отключает email-доставку, но при `DEV_MODE=false` код в ответ не попадает.

### Telegram

Telegram **не используется для входа**. Только:

- привязка аккаунта для уведомлений (`Настройки → Уведомления → Telegram`);
- доставка push в бот.

---

## Первые шаги после входа

### 1. Онбординг (4 шага)

| Шаг | Содержание |
|-----|------------|
| Профиль | Имя, фамилия, должность, отдел, часовой пояс, язык |
| Тур | Краткий обзор модулей Hub |
| Тема | Светлая / тёмная / system + акцент |
| Проект | Создать первый проект или войти по invite-коду |

Завершение: `POST /api/auth/me/onboarding` (`frontend/src/routes/onboarding.lazy.tsx`).

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

В ветке `calls-feature-codex` в заголовке открытого треда добавлена кнопка
«Подписаться» / «Отписаться», если действие разрешено сервером. Во время
запроса и обновления данных повторное нажатие недоступно; после ошибки можно
повторить действие. Это проверенный Docker unit-тестами фрагмент, не релиз
в `master`: автоматическая отметка прочтения треда и браузерная приёмка ещё
не завершены. [Контракт и проверка](../superpowers/plans/2026-09-05-thread-follow-control.md).

### 5. Глобальный поиск

`Ctrl+K` / `Cmd+K` — задачи, кейсы, пользователи, проекты, каналы, файлы.

Подробнее: [SHORTCUTS.md](SHORTCUTS.md).

---

## Проверка перед коммитом

Единственная QA-команда использует один из восьми fixed services:
`docker compose -f docker-compose.local.yml run --rm <exact-service>`.
Допустимы только `backend-unit`, `frontend-unit`, `python-contract`, `python-api`,
`python-integration`, `python-e2e-smoke`, `python-e2e-full` и `python-load`.
Host test/build/lint и удалённые runners не являются fallback. До фактического
evidence проверка остаётся `NOT_RUN`, без ложного green.

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
