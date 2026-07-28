# Справочник API

> **Обновлено:** 2026-06-22 (Wave C FA-03) · Базовый путь: `/api` · WebSocket: `GET /ws` (вне `/api`)

Hub предоставляет REST API под `/api/*`. Подсчёт:

| Категория | Количество | Примечание |
|-----------|------------|------------|
| OpenAPI paths (`backend/docs/swagger.json`) | **252** | Measured 2026-07-28 — канон для drift-check |
| Prod handler `RegisterRoutes` (исторический inventory) | **~283** | 21+ модулей; число выше swagger из-за группировки/alias |
| Dev-only (`/dev/*`) | **+3** | Только без `-tags release` |
| Infra probes | `/live`, `/ready`, `/health` | `wiring_messenger.go` |
| Metrics | `GET /metrics` | `wiring_fiber.go` |
| **Итого (ориентир wiki)** | **~252 OpenAPI / ~286 route inventory** | См. `docs/system_specification.md` §Endpoint inventory |

Полная спецификация: `docs/system_specification.md`, `docs/TECHNICAL.md`.  
OpenAPI (dev): `GET /docs`, `GET /docs/doc.json`.

---

## Аутентификация запросов

### JWT Bearer (основной)

```http
Authorization: Bearer <access_token>
```

Access token TTL: **15 минут**. Refresh: `POST /api/auth/refresh` (Redis, 30 дней).

### API Key (webhooks, TMS CI)

```http
X-API-Key: <api_key>
```

Ключи: `POST /api/user/api-keys/` (SHA-256 hash at rest).

### Ошибки

Публичный HTTP JSON (NFR-13 / TZ, `platform/errors` + `middleware.ErrorHandler`):

```json
{"error": "описание", "request_id": "uuid"}
```

На фронте тип `ApiError` (`frontend/src/api/client.ts`) — **клиентское** представление: `code` = HTTP status, `message` = поле `error` из тела, опционально `request_id`. Это не отдельный wire-контракт.

---

## Auth — Magic Link only (v2.5+)

> **ADR-AUTH-01:** единственный login — Magic Link. Telegram — **только** notify bind (`/notifications/telegram/*`), не login.  
> Удалено v2.5: `POST /auth/telegram`, `POST /auth/telegram-code`.

| Метод | Путь | Auth | Описание |
|---|---|---|---|
| POST | `/auth/magic-link` | — | Отправить 6-значный код на email (rate 3/10m) |
| POST | `/auth/magic-link/verify` | — | Верифицировать код → access + refresh JWT |
| POST | `/auth/refresh` | — | Обновить access token |
| POST | `/auth/logout` | JWT | Инвалидация refresh-токена текущей сессии |
| GET | `/auth/me` | JWT | Профиль текущего пользователя |
| PUT | `/auth/me` | JWT | Обновить профиль (имя, timezone, язык, тема) |
| POST | `/auth/me/onboarding` | JWT | Завершить онбординг |
| GET | `/auth/sessions` | JWT | Список активных сессий (device, IP) |
| DELETE | `/auth/sessions/:sessionId` | JWT | Отозвать сессию |
| GET | `/admin/registration-mode` | Admin | Режим регистрации (invite-only) |
| PUT | `/admin/registration-mode` | Admin | Установить режим регистрации |

### Пример: Magic Link

```bash
# 1. Запрос кода (через Caddy :8080 — канон локального UX; backend напрямую — :3000)
curl -sS -X POST http://localhost:8080/api/auth/magic-link \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@company.com"}'

# 2. Верификация
curl -sS -X POST http://localhost:8080/api/auth/magic-link/verify \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@company.com","code":"123456"}'
```

> Локальный UX-путь — Caddy `:8080` (см. QUICKSTART/DEPLOY). Прямой backend без прокси — `:3000`.

---

## Projects (22 endpoints)

Проекты, участники, приглашения, команды.

### Проекты

| Метод | Путь | Описание |
|---|---|---|
| POST | `/projects/` | Создать проект |
| GET | `/projects/` | Список проектов пользователя |
| GET | `/projects/trash` | Список удалённых проектов |
| GET | `/projects/:projectId` | Получить проект по ID |
| PUT | `/projects/:projectId` | Обновить проект |
| POST | `/projects/:projectId/archive` | Архивировать проект |
| POST | `/projects/:projectId/unarchive` | Разархивировать проект |
| DELETE | `/projects/:projectId` | Удалить проект (в корзину) |
| POST | `/projects/:projectId/restore` | Восстановить из корзины |

### Участники

| Метод | Путь | Описание |
|---|---|---|
| GET | `/projects/:projectId/members` | Список участников проекта |
| POST | `/projects/:projectId/members` | Добавить участника |
| PUT | `/projects/:projectId/members/:userId` | Изменить роль участника |
| DELETE | `/projects/:projectId/members/:userId` | Удалить участника |

### Приглашения

| Метод | Путь | Описание |
|---|---|---|
| POST | `/projects/:projectId/invites` | Создать приглашение |
| GET | `/projects/:projectId/invites` | Список приглашений |
| DELETE | `/projects/:projectId/invites/:inviteId` | Отозвать приглашение |
| POST | `/invites/join` | Принять приглашение по коду |

### Команды

| Метод | Путь | Описание |
|---|---|---|
| POST | `/teams/` | Создать команду |
| GET | `/teams/` | Список команд |
| POST | `/teams/:teamId/members` | Добавить участника в команду |
| DELETE | `/teams/:teamId/members/:userId` | Удалить участника из команды |
| POST | `/teams/:teamId/projects/:projectId` | Привязать команду к проекту |

---

## Project Templates (5 endpoints)

Шаблоны проектов: сохранение, импорт/экспорт, удаление.

| Метод | Путь | Описание |
|---|---|---|
| GET | `/project-templates/` | Список шаблонов (системные и пользовательские) |
| POST | `/project-templates/` | Сохранить проект как шаблон |
| POST | `/project-templates/import` | Импорт шаблона из JSON |
| GET | `/project-templates/:id/export` | Экспорт шаблона в JSON |
| DELETE | `/project-templates/:id` | Удалить пользовательский шаблон |

---

## Tracker (38 endpoints)

Задачи, workflow, доски, метки, кастомные поля, время, комментарии.

### Задачи

| Метод | Путь | Описание |
|---|---|---|
| POST | `/projects/:projectId/issues/` | Создать задачу |
| GET | `/projects/:projectId/issues/` | Список задач (cursor pagination) |
| GET | `/projects/:projectId/issues/:issueKey` | Получить задачу по ключу |
| PUT | `/projects/:projectId/issues/:issueKey` | Обновить задачу |
| DELETE | `/projects/:projectId/issues/:issueKey` | Удалить задачу |
| PUT | `/projects/:projectId/issues/:issueKey/position` | Переместить на доске (Kanban drag) |
| POST | `/projects/:projectId/issues/jql` | JQL-поиск задач |

### Связи между задачами

| Метод | Путь | Описание |
|---|---|---|
| POST | `/projects/:projectId/issues/:issueKey/links` | Создать связь |
| GET | `/projects/:projectId/issues/:issueKey/links` | Список связей |
| DELETE | `/projects/:projectId/issues/:issueKey/links/:linkId` | Удалить связь |
| GET | `/projects/:projectId/issue-links` | Batch связей проекта (Gantt, A13-004; cap 2000) |

### Комментарии и реакции

| Метод | Путь | Описание |
|---|---|---|
| POST | `/projects/:projectId/issues/:issueKey/comments` | Добавить комментарий |
| GET | `/projects/:projectId/issues/:issueKey/comments` | Список комментариев |
| PUT | `/projects/:projectId/issues/:issueKey/comments/:commentId` | Редактировать комментарий |
| DELETE | `/projects/:projectId/issues/:issueKey/comments/:commentId` | Удалить комментарий |
| POST | `/projects/:projectId/issues/:issueKey/comments/:commentId/reactions` | Поставить/убрать реакцию |

### Активность

| Метод | Путь | Описание |
|---|---|---|
| GET | `/projects/:projectId/issues/:issueKey/activity` | Лента активности задачи |

### Время

| Метод | Путь | Описание |
|---|---|---|
| POST | `/projects/:projectId/issues/:issueKey/time` | Записать время |
| POST | `/projects/:projectId/issues/:issueKey/timer/start` | Запустить таймер |
| POST | `/projects/:projectId/issues/:issueKey/timer/stop` | Остановить таймер |
| GET | `/projects/:projectId/issues/:issueKey/time-logs` | Список записей времени |

### Доски

| Метод | Путь | Описание |
|---|---|---|
| GET | `/projects/:projectId/boards/` | Список досок проекта |
| POST | `/projects/:projectId/boards/` | Создать доску |
| GET | `/projects/:projectId/boards/:boardId` | Данные доски (колонки + задачи) |

### Workflow

| Метод | Путь | Описание |
|---|---|---|
| GET | `/projects/:projectId/workflow/statuses` | Список статусов |
| POST | `/projects/:projectId/workflow/statuses` | Создать статус |
| PUT | `/projects/:projectId/workflow/statuses/:statusId` | Обновить статус |
| DELETE | `/projects/:projectId/workflow/statuses/:statusId` | Удалить статус |
| GET | `/projects/:projectId/workflow/transitions` | Список переходов |
| POST | `/projects/:projectId/workflow/transitions` | Создать переход |
| DELETE | `/projects/:projectId/workflow/transitions/:transitionId` | Удалить переход |

### Метки

| Метод | Путь | Описание |
|---|---|---|
| GET | `/projects/:projectId/labels/` | Список меток |
| POST | `/projects/:projectId/labels/` | Создать метку |
| DELETE | `/projects/:projectId/labels/:labelId` | Удалить метку |
| GET | `/projects/:projectId/labels/groups` | Список групп меток |
| POST | `/projects/:projectId/labels/groups` | Создать группу меток |

### Кастомные поля

| Метод | Путь | Описание |
|---|---|---|
| GET | `/projects/:projectId/custom-fields/` | Список кастомных полей |
| POST | `/projects/:projectId/custom-fields/` | Создать кастомное поле |
| DELETE | `/projects/:projectId/custom-fields/:fieldId` | Удалить кастомное поле |

---

## Sprint (14 endpoints)

Спринты, burndown, velocity, Planning Poker.

### Спринты

| Метод | Путь | Описание |
|---|---|---|
| POST | `/projects/:projectId/sprints/` | Создать спринт |
| GET | `/projects/:projectId/sprints/` | Список спринтов проекта |
| GET | `/projects/:projectId/sprints/:sprintId` | Получить спринт |
| PUT | `/projects/:projectId/sprints/:sprintId` | Обновить спринт |
| POST | `/projects/:projectId/sprints/:sprintId/start` | Стартовать спринт |
| POST | `/projects/:projectId/sprints/:sprintId/complete` | Завершить спринт |
| PUT | `/projects/:projectId/sprints/:sprintId/retro` | Обновить ретро-заметки |
| GET | `/projects/:projectId/sprints/:sprintId/burndown` | Данные burndown-графика |
| POST | `/projects/:projectId/sprints/:sprintId/issues` | Добавить задачи в спринт |
| POST | `/projects/:projectId/sprints/:sprintId/poker` | Начать сессию Planning Poker |

### Velocity

| Метод | Путь | Описание |
|---|---|---|
| GET | `/projects/:projectId/velocity/` | Данные velocity (завершённые спринты) |

### Planning Poker

| Метод | Путь | Описание |
|---|---|---|
| POST | `/sprints/poker/:sessionId/vote` | Проголосовать (оценить задачу) |
| POST | `/sprints/poker/:sessionId/reveal` | Раскрыть голоса |
| GET | `/sprints/poker/:sessionId` | Получить состояние сессии |

---

## Messenger (42 endpoints)

Каналы, сообщения, присутствие, опросы, пригласительные ссылки.

### Каналы

| Метод | Путь | Описание |
|---|---|---|
| POST | `/messenger/channels/` | Создать канал |
| GET | `/messenger/channels/` | Мои каналы |
| GET | `/messenger/channels/:channelId` | Получить канал |
| PUT | `/messenger/channels/:channelId` | Обновить канал (название, описание) |
| DELETE | `/messenger/channels/:channelId` | Удалить канал |
| POST | `/messenger/channels/:channelId/join` | Присоединиться к каналу |
| POST | `/messenger/channels/:channelId/leave` | Покинуть канал |
| POST | `/messenger/channels/:channelId/members` | Добавить участников |
| GET | `/messenger/channels/:channelId/members` | Список участников канала |
| POST | `/messenger/channels/:channelId/star` | Добавить/убрать из избранного |
| POST | `/messenger/channels/:channelId/mute` | Замьютить канал |
| POST | `/messenger/channels/:channelId/unmute` | Размьютить канал |
| POST | `/messenger/channels/:channelId/typing` | Отправить индикатор набора |

### Сообщения

| Метод | Путь | Описание |
|---|---|---|
| POST | `/messenger/channels/:channelId/messages/` | Отправить сообщение |
| GET | `/messenger/channels/:channelId/messages/` | Список сообщений (cursor pagination) |
| PUT | `/messenger/channels/:channelId/messages/:messageId` | Редактировать сообщение |
| DELETE | `/messenger/channels/:channelId/messages/:messageId` | Удалить сообщение |
| POST | `/messenger/channels/:channelId/messages/:messageId/reactions` | Поставить/убрать реакцию |
| POST | `/messenger/channels/:channelId/messages/:messageId/pin` | Закрепить/открепить |
| POST | `/messenger/channels/:channelId/messages/:messageId/forward` | Переслать сообщение |
| GET | `/messenger/channels/:channelId/messages/:messageId/replies` | Список ответов в треде |

### Сообщения (top-level)

| Метод | Путь | Описание |
|---|---|---|
| POST | `/messages/:messageId/promote-to-issue` | Преобразовать сообщение в задачу |

### Присутствие

| Метод | Путь | Описание |
|---|---|---|
| GET | `/messenger/presence` | Онлайн-статусы пользователей |
| PUT | `/messenger/presence/status` | Установить кастомный статус |
| POST | `/messenger/presence/heartbeat` | Heartbeat (поддержание статуса online) |

### Непрочитанные

| Метод | Путь | Описание |
|---|---|---|
| GET | `/messenger/unread` | Счётчики непрочитанных по каналам |
| POST | `/messenger/channels/:channelId/read` | Отметить канал прочитанным |

### Закреплённые сообщения

| Метод | Путь | Описание |
|---|---|---|
| GET | `/messenger/channels/:channelId/pins` | Список закреплённых сообщений |

### Поиск

| Метод | Путь | Описание |
|---|---|---|
| GET | `/messenger/search` | Поиск по сообщениям |

### Закладки

| Метод | Путь | Описание |
|---|---|---|
| GET | `/messenger/bookmarks` | Список закладок |
| POST | `/messenger/bookmarks` | Добавить закладку |
| DELETE | `/messenger/bookmarks/:bookmarkId` | Удалить закладку |

### DM (личные сообщения)

| Метод | Путь | Описание |
|---|---|---|
| POST | `/messenger/dm` | Создать или открыть DM |

### Опросы

| Метод | Путь | Описание |
|---|---|---|
| POST | `/messenger/channels/:channelId/polls` | Создать опрос в канале |
| POST | `/messenger/polls/:pollId/vote` | Проголосовать в опросе |
| POST | `/messenger/polls/:pollId/close` | Закрыть опрос |
| GET | `/messenger/polls/:pollId` | Результаты опроса |

### Пригласительные ссылки

| Метод | Путь | Описание |
|---|---|---|
| POST | `/messenger/channels/:channelId/invites` | Создать пригласительную ссылку |
| GET | `/messenger/channels/:channelId/invites` | Список приглашений канала |
| DELETE | `/messenger/channels/:channelId/invites/:inviteId` | Отозвать приглашение |
| GET | `/messenger/invites/:token` | Валидировать приглашение по токену |
| POST | `/messenger/invites/:token/accept` | Принять приглашение |

---

## TMS (50 endpoints)

Тест-кейсы, наборы, компоненты, прогоны, результаты, Allure, отчёты, окружения.

### Тест-кейсы

| Метод | Путь | Описание |
|---|---|---|
| POST | `/projects/:projectId/tms/cases/` | Создать тест-кейс |
| GET | `/projects/:projectId/tms/cases/` | Список тест-кейсов |
| GET | `/projects/:projectId/tms/cases/:caseKey` | Получить тест-кейс по ключу |
| PUT | `/projects/:projectId/tms/cases/:caseKey` | Обновить тест-кейс |
| DELETE | `/projects/:projectId/tms/cases/:caseKey` | Удалить тест-кейс |

### Версии тест-кейсов

| Метод | Путь | Описание |
|---|---|---|
| GET | `/projects/:projectId/tms/cases/:caseKey/versions` | Список версий |
| POST | `/projects/:projectId/tms/cases/:caseKey/rollback/:versionId` | Откатиться к версии |

### Связи тест-кейсов с задачами

| Метод | Путь | Описание |
|---|---|---|
| POST | `/projects/:projectId/tms/cases/:caseKey/links` | Привязать к задаче |
| DELETE | `/projects/:projectId/tms/cases/:caseKey/links/:issueId` | Отвязать от задачи |
| GET | `/projects/:projectId/tms/cases/:caseKey/links` | Список привязанных задач |

### Тестовые наборы (Suites)

| Метод | Путь | Описание |
|---|---|---|
| POST | `/projects/:projectId/tms/suites/` | Создать набор |
| GET | `/projects/:projectId/tms/suites/` | Список наборов |
| PUT | `/projects/:projectId/tms/suites/:suiteId` | Обновить набор |
| DELETE | `/projects/:projectId/tms/suites/:suiteId` | Удалить набор |

### Компоненты

| Метод | Путь | Описание |
|---|---|---|
| POST | `/projects/:projectId/tms/components/` | Создать компонент |
| GET | `/projects/:projectId/tms/components/` | Список компонентов |
| DELETE | `/projects/:projectId/tms/components/:componentId` | Удалить компонент |

### Тестовые прогоны

| Метод | Путь | Описание |
|---|---|---|
| POST | `/projects/:projectId/tms/runs/` | Создать прогон |
| GET | `/projects/:projectId/tms/runs/` | Список прогонов |
| GET | `/projects/:projectId/tms/runs/:runId` | Получить прогон |
| PUT | `/projects/:projectId/tms/runs/:runId/status` | Обновить статус прогона |
| POST | `/projects/:projectId/tms/runs/:runId/reopen` | Переоткрыть прогон |
| POST | `/projects/:projectId/tms/runs/:runId/cases` | Добавить кейсы в прогон |
| POST | `/projects/:projectId/tms/runs/:runId/assignees` | Назначить тестировщиков |
| POST | `/projects/:projectId/tms/runs/:runId/start` | Стартовать прогон |
| POST | `/projects/:projectId/tms/runs/:runId/complete` | Завершить прогон |
| GET | `/projects/:projectId/tms/runs/:runId/detail` | Детали прогона (кейсы + результаты) |

### Результаты

| Метод | Путь | Описание |
|---|---|---|
| POST | `/projects/:projectId/tms/runs/:runId/results` | Создать результат |
| PUT | `/projects/:projectId/tms/runs/:runId/results/:resultId` | Обновить результат (выполнить) |
| GET | `/projects/:projectId/tms/runs/:runId/results` | Список результатов прогона |

### Баги из результатов

| Метод | Путь | Описание |
|---|---|---|
| POST | `/projects/:projectId/tms/runs/:runId/results/:resultId/bug` | Создать баг из упавшего теста |

### Allure

| Метод | Путь | Описание |
|---|---|---|
| POST | `/projects/:projectId/tms/allure/import` | Импорт Allure results (новый прогон) |
| POST | `/projects/:projectId/tms/runs/:runId/allure` | Импорт Allure в существующий прогон |

### CI/CD Webhooks

> **Auth (фактическое поведение):** оба endpoint под `authMw` — требуется `Authorization: Bearer <JWT>` (или `X-API-Key`). Заголовок `X-Hub-Token` из Swagger **не обрабатывается** хендлером. RBAC `RequirePermission` на webhook-роутах пока **не применяется** (см. FA-02-02 — IDOR backlog); используйте `/tms/allure/import` (требует `PermTestRunCreate`) для строгой проверки прав.

| Метод | Путь | Auth | Описание |
|---|---|---|---|
| POST | `/projects/:projectId/tms/webhook/` | JWT | Webhook для Allure results из CI |
| POST | `/projects/:projectId/tms/webhook/junit` | JWT | Webhook для JUnit XML из CI |

### Отчёты

| Метод | Путь | Описание |
|---|---|---|
| GET | `/projects/:projectId/tms/reports/history` | История прогонов |
| GET | `/projects/:projectId/tms/reports/flaky` | Flaky-тесты |
| GET | `/projects/:projectId/tms/reports/coverage` | Покрытие тестами |
| GET | `/projects/:projectId/tms/reports/trends` | Тренды (pass/fail/skip) |
| GET | `/projects/:projectId/tms/reports/comparison` | Сравнение двух прогонов |
| GET | `/projects/:projectId/tms/reports/traceability` | Матрица трассировки (кейсы-задачи) |
| GET | `/projects/:projectId/tms/reports/defect-categories` | Категории дефектов |
| GET | `/projects/:projectId/tms/reports/env-comparison` | Сравнение окружений |

### Импорт/экспорт TMS

| Метод | Путь | Описание |
|---|---|---|
| POST | `/projects/:projectId/tms/import` | Импорт тест-кейсов (CSV/XML) |
| GET | `/projects/:projectId/tms/export` | Экспорт тест-кейсов |
| GET | `/projects/:projectId/tms/runs/:runId/export` | Экспорт отчёта прогона |

### Тестовые окружения

| Метод | Путь | Описание |
|---|---|---|
| POST | `/projects/:projectId/environments/` | Создать окружение |
| GET | `/projects/:projectId/environments/` | Список окружений |
| PUT | `/projects/:projectId/environments/:envId` | Обновить окружение |
| DELETE | `/projects/:projectId/environments/:envId` | Удалить окружение |

---

## Calls (21 endpoints)

Звонки, запись, планирование, чат.

### Звонки

| Метод | Путь | Описание |
|---|---|---|
| POST | `/calls/guest-join` | Присоединиться к звонку как гость (без аккаунта) |
| POST | `/calls/` | Создать звонок в канале (только член канала; иначе 403) |
| POST | `/calls/instant` | Создать мгновенный звонок |
| GET | `/calls/invite/:code` | Присоединиться по коду приглашения |
| GET | `/calls/channel/:channelId` | Получить активный звонок канала |
| POST | `/calls/:callId/join` | Присоединиться к звонку |
| POST | `/calls/:callId/leave` | Покинуть звонок |
| POST | `/calls/:callId/end` | Завершить звонок |
| POST | `/calls/:callId/raise-hand` | Поднять руку |
| POST | `/calls/:callId/chat` | Отправить сообщение в чат звонка |
| GET | `/calls/:callId/chat` | Список сообщений чата звонка |
| GET | `/calls/conflicts` | Проверить конфликты расписания |

### Запись

| Метод | Путь | Описание |
|---|---|---|
| POST | `/calls/:callId/recording/start` | Начать запись (только член канала звонка) |
| POST | `/calls/:callId/recording/stop` | Остановить запись (член канала; `recording_id` принадлежит `callId`) |
| GET | `/projects/:projectId/recordings` | Список записей проекта |
| GET | `/recordings/:recordingId/download` | URL для скачивания записи |

### Планирование

| Метод | Путь | Описание |
|---|---|---|
| POST | `/projects/:projectId/scheduled-calls` | Запланировать звонок |
| GET | `/channels/:channelId/scheduled-calls` | Список запланированных звонков |
| PUT | `/scheduled-calls/:id` | Обновить запланированный звонок |
| DELETE | `/scheduled-calls/:id` | Удалить запланированный звонок |

### Webhook

| Метод | Путь | Описание |
|---|---|---|
| POST | `/calls/webhook` | LiveKit webhook (события звонков) |

---

## Dashboard (7 endpoints)

Данные дашбордов, сохранение layout, шаблоны.

| Метод | Путь | Описание |
|---|---|---|
| GET | `/dashboard/personal` | Персональный дашборд (мои задачи, активность) |
| GET | `/dashboard/project/:projectId` | Проектный дашборд (статистика; только член проекта; иначе 403) |
| GET | `/dashboard/admin` | Админский дашборд (общая статистика) |
| GET | `/dashboard/layout` | Получить layout дашборда |
| PUT | `/dashboard/layout` | Сохранить layout (позиции виджетов) |
| GET | `/dashboard/templates` | Список шаблонов дашбордов |
| POST | `/dashboard/templates/apply` | Применить шаблон |

---

## Notifications (10 endpoints)

Уведомления в приложении, push-подписки, настройки.

| Метод | Путь | Описание |
|---|---|---|
| GET | `/notifications/` | Список уведомлений (pagination) |
| GET | `/notifications/unread-count` | Счётчик непрочитанных |
| PATCH | `/notifications/:id/read` | Отметить прочитанным |
| PATCH | `/notifications/read-all` | Отметить все прочитанными |
| DELETE | `/notifications/:id` | Удалить уведомление |
| GET | `/notifications/preferences` | Настройки уведомлений |
| PUT | `/notifications/preferences` | Обновить настройки |
| GET | `/notifications/push/vapid-key` | Получить VAPID public key |
| POST | `/notifications/push/subscribe` | Подписаться на push-уведомления |
| DELETE | `/notifications/push/unsubscribe` | Отписаться от push |

---

## Search (5 endpoints)

Глобальный поиск по всем сущностям, сохранённые фильтры.

| Метод | Путь | Описание |
|---|---|---|
| GET | `/search` | Глобальный поиск (задачи, тест-кейсы, сообщения) |
| POST | `/filters/` | Создать сохранённый фильтр (только член проекта; иначе 403) |
| GET | `/filters/` | Список сохранённых фильтров (только член проекта; иначе 403) |
| DELETE | `/filters/:id` | Удалить фильтр |
| PATCH | `/filters/:id/share` | Поделиться фильтром |

---

## GitHub (9 endpoints)

Вебхуки, интеграции, CI/CD статусы, создание веток, превью ссылок.

| Метод | Путь | Описание |
|---|---|---|
| POST | `/github/webhook` | GitHub webhook (push, PR, workflow) |
| POST | `/projects/:projectId/github/integrations` | Создать интеграцию |
| GET | `/projects/:projectId/github/integrations` | Список интеграций |
| DELETE | `/projects/:projectId/github/integrations/:integrationId` | Удалить интеграцию |
| GET | `/projects/:projectId/github/workflows` | CI/CD workflow runs |
| GET | `/projects/:projectId/github/workflows/:runId` | Детали workflow run |
| GET | `/issues/:issueKey/github-links` | GitHub-ссылки задачи (PR, коммиты) |
| POST | `/projects/:projectId/issues/:issueKey/github/create-branch` | Создать ветку из задачи (`task.edit`) |
| GET | `/link-preview` | Превью ссылки (OG-данные) |

---

## GitLab (7 endpoints)

Вебхуки, интеграции, пайплайны, создание веток.

| Метод | Путь | Описание |
|---|---|---|
| POST | `/gitlab/webhook` | GitLab webhook (push, MR, pipeline) |
| POST | `/projects/:projectId/gitlab/integrations` | Создать интеграцию |
| GET | `/projects/:projectId/gitlab/integrations` | Список интеграций |
| DELETE | `/projects/:projectId/gitlab/integrations/:integrationId` | Удалить интеграцию |
| GET | `/projects/:projectId/gitlab/pipelines` | Список пайплайнов |
| POST | `/projects/:projectId/gitlab/pipelines/trigger` | Запустить пайплайн |
| POST | `/projects/:projectId/issues/:issueKey/gitlab/create-branch` | Создать ветку из задачи (`task.edit`) |

---

## Storage (4 endpoints)

Файловое хранилище через SeaweedFS (S3-совместимое).

| Метод | Путь | Описание |
|---|---|---|
| POST | `/files/upload-url` | Получить presigned URL для загрузки |
| POST | `/files/:fileId/confirm` | Подтвердить завершение загрузки |
| GET | `/files/:fileId/url` | Получить URL для скачивания |
| DELETE | `/files/:fileId` | Удалить файл |

---

## Import/Export (6 endpoints)

Импорт задач и экспорт данных.

| Метод | Путь | Описание |
|---|---|---|
| POST | `/projects/:projectId/import/issues` | Импорт задач (CSV/Excel) |
| POST | `/projects/:projectId/import/jira` | Импорт из Jira (JSON) |
| GET | `/projects/:projectId/export/issues` | Экспорт задач (CSV/Excel) |
| GET | `/projects/:projectId/export/full` | Полный экспорт проекта (ZIP) |
| GET | `/projects/:projectId/test-runs/:runId/export/pdf` | Экспорт прогона в PDF |
| GET | `/projects/:projectId/sprints/:sprintId/export/pdf` | Экспорт спринта в PDF |

---

## Backup (4 endpoints)

Управление бэкапами (только system admin).

| Метод | Путь | Описание |
|---|---|---|
| POST | `/admin/backups/` | Создать бэкап |
| GET | `/admin/backups/` | Список бэкапов |
| POST | `/admin/backups/:id/restore` | Восстановить из бэкапа |
| DELETE | `/admin/backups/:id` | Удалить бэкап |

---

## API Gateway (8 endpoints)

API-ключи для программного доступа, webhook-подписки.

### API-ключи

| Метод | Путь | Описание |
|---|---|---|
| POST | `/user/api-keys/` | Создать API-ключ |
| GET | `/user/api-keys/` | Список API-ключей |
| DELETE | `/user/api-keys/:id` | Отозвать API-ключ |

### Webhook-подписки (проект)

| Метод | Путь | Описание |
|---|---|---|
| POST | `/projects/:projectId/webhooks/` | Создать endpoint для проекта |
| GET | `/projects/:projectId/webhooks/` | Список endpoints проекта |

### Webhook-подписки (по ID)

| Метод | Путь | Описание |
|---|---|---|
| PUT | `/webhooks/:id` | Обновить endpoint |
| DELETE | `/webhooks/:id` | Удалить endpoint |
| GET | `/webhooks/:id/deliveries` | История доставок |

---

## RBAC (6 endpoints)

Кастомные роли проекта.

| Метод | Путь | Описание |
|---|---|---|
| POST | `/projects/:projectId/roles/` | Создать кастомную роль |
| GET | `/projects/:projectId/roles/` | Список ролей |
| PUT | `/projects/:projectId/roles/:roleId` | Обновить роль (название, права) |
| DELETE | `/projects/:projectId/roles/:roleId` | Удалить роль |
| POST | `/projects/:projectId/roles/:roleId/assign/:userId` | Назначить роль пользователю |
| DELETE | `/projects/:projectId/roles/:roleId/assign/:userId` | Снять роль с пользователя |

---

## Admin / Audit (12 endpoints)

Аудит, лимиты ресурсов, корзина, управление пользователями, режим регистрации. Все маршруты требуют system admin.

### Журнал аудита

| Метод | Путь | Описание |
|---|---|---|
| GET | `/admin/audit` | Лог аудита (фильтры: `action`, `entity_type`, `entity_id`, `user_id`, `project_id`, `ip_address`, `date_from`, `date_to`, `slow` (duration_ms>500), `page`, `per_page`). Actions: `alert.firing`, `alert.resolved` (Alertmanager webhook, Phase 36) |

### HTTP audit (middleware)

Не отдельный endpoint. `HTTPAuditMiddleware` (`audit/middleware.go`) подключён ко всем `/api/*` в `wiring_fiber.go` и асинхронно пишет каждый HTTP-запрос в журнал аудита (method, path без query, status, duration, request_id, user_id, IP, User-Agent). Пропускает: `OPTIONS`, `/api/health`, `/api/docs*`, `/api/webhooks/alertmanager`.

### Лимиты и корзина

| Метод | Путь | Описание |
|---|---|---|
| GET | `/admin/limits` | Лимиты ресурсов (проекты, участники и др.) |
| PUT | `/admin/limits/:key` | Обновить лимит |
| GET | `/admin/trash` | Корзина (удалённые сущности). Query `type`: `issue` \| `test_case` \| `file` \| `project` \| `newspost` \| `scheduled_call` (retention 30 дней, NFR-11) |
| POST | `/admin/trash/:type/:id/restore` | Восстановить из корзины (`newspost` — FR-INTR-08; `scheduled_call` — FR-CAL-05: вся RRULE-серия) |

### Управление пользователями

| Метод | Путь | Описание |
|---|---|---|
| POST | `/admin/users/:id/deactivate` | Деактивировать пользователя |
| POST | `/admin/users/:id/ban` | Заблокировать пользователя (`{"reason": "..."}`) |
| POST | `/admin/users/:id/reassign` | Переназначить задачи пользователя |
| DELETE | `/admin/users/:id/gdpr` | GDPR-удаление пользователя |

### Режим регистрации

| Метод | Путь | Описание |
|---|---|---|
| GET | `/admin/registration-mode` | Получить режим регистрации (open/invite) |
| PUT | `/admin/registration-mode` | Установить режим регистрации |

### Observability (OBS-01, admin)

| Метод | Путь | Описание |
|---|---|---|
| GET | `/admin/observability/ops` | In-memory метрики `observability.Timed()` (operation, count, avg_ms, error_rate) |

---

## Intranet (MOD-INTRA, Phase 39–40)

Корпоративная лента новостей и справочник сотрудников. Чтение — все авторизованные; публикация новостей — system admin/founder.

### News (MOD-01)

| Метод | Путь | Описание |
|---|---|---|
| GET | `/intranet/news` | Опубликованные новости (`page`, `per_page`; pinned first) |
| GET | `/intranet/news/admin` | Все новости incl. черновики (admin) |
| GET | `/intranet/news/:newsId` | Одна новость |
| POST | `/intranet/news` | Создать (admin) |
| PUT | `/intranet/news/:newsId` | Обновить (admin, optimistic `version`) |
| DELETE | `/intranet/news/:newsId` | Soft delete (admin) |

### Directory (MOD-02)

| Метод | Путь | Описание |
|---|---|---|
| GET | `/intranet/directory` | Дерево отделов → активные сотрудники |
| GET | `/intranet/directory/users/:userId` | Карточка сотрудника (только `status=active`) |

Ctrl+K user search → `/intranet/directory?user={id}`.

### Kudos (MOD-03)

| Метод | Путь | Описание |
|---|---|---|
| GET | `/intranet/kudos/feed` | Публичная лента благодарностей |
| GET | `/intranet/kudos/received` | Полученные текущим пользователем |
| POST | `/intranet/kudos` | Отправить kudos (`recipient_id`, `message`, `is_public?`) |

Уведомление получателю через NotificationService. Push deep-link → `/intranet?tab=kudos`.

### Random Coffee (V4-05)

| Метод | Путь | Описание |
|---|---|---|
| GET | `/intranet/coffee/enrollment` | Статус участия |
| PUT | `/intranet/coffee/enrollment` | Opt-in/out (`{"active": true}`) |
| GET | `/intranet/coffee/match` | Напарник из последнего раунда |
| GET | `/intranet/coffee/stats` | Число участников (admin) |
| GET | `/intranet/coffee/rounds` | История раундов (admin) |
| POST | `/intranet/coffee/run` | Запуск matching (admin, min 2 участника) |

Anti-repeat: старается не повторять пары из предыдущего раунда. Уведомление обеим сторонам пары.

### Intranet v2 — **planned/spec** (FR-INTR, Волна 2)

> Спецификация: `docs/system_specification.md` § «Intranet v2 (FR-INTR)». Реализации в коде ещё нет.

| Метод | Путь | Описание | Статус |
|---|---|---|---|
| POST | `/intranet/news` | + опциональный `poll` (вопрос, 2–10 вариантов) | planned |
| GET | `/intranet/news/:newsId/poll` | Опрос поста + агрегаты голосов | planned |
| POST | `/intranet/news/:newsId/poll/vote` | Голос (1 вариант; повторный → 422) | planned |
| PUT | `/intranet/news/:newsId/poll` | Правка/закрытие опроса (после первого голоса → 422) | planned |
| GET | `/intranet/onboarding/templates` | Системные шаблоны dev/non-dev (admin) | planned |
| POST | `/intranet/onboarding/assignments` | Назначить шаблон новичку (admin; re-assign → 409 без confirm) | planned |
| GET | `/intranet/onboarding/assignments` | Прогресс по всем новичкам (admin, offset/limit) | planned |
| GET | `/intranet/onboarding/me` | Мой чеклист | planned |
| PATCH | `/intranet/onboarding/me/items/:itemId` | Отметить пункт (`checked`, `version`) | planned |
| GET | `/intranet/kudos/leaderboard` | Топ получателей за `period=month\|quarter\|all` | planned |
| DELETE | `/intranet/news/:newsId` | Расширение: admin **или автор** (постмодерация) | planned |

---

## Calendar (MOD-04, Phase 41)

Единый календарь: созвоны, спринты, дедлайны задач, одобренные отпуска (только проекты пользователя для tracker-событий).

| Метод | Путь | Описание |
|---|---|---|
| GET | `/calendar/events` | События за период (`from`, `to` — RFC3339) |
| GET | `/calendar/export.ics` | Экспорт iCalendar за тот же период |

Типы событий в JSON: `call`, `sprint`, `deadline`, `leave`.

### Calendar v2 — **planned/spec** (FR-CAL, Волна 2)

> Спецификация: `docs/system_specification.md` § «Calendar v2 (FR-CAL)». Реализации в коде ещё нет.

| Метод | Путь | Описание | Статус |
|---|---|---|---|
| GET | `/calendar/resources` | Справочник ресурсов (room/equipment/other) | planned |
| POST | `/calendar/resources` | Создать ресурс (admin) | planned |
| PUT | `/calendar/resources/:id` | Обновить (admin, `version`) | planned |
| DELETE | `/calendar/resources/:id` | Деактивировать (admin, брони сохраняются) | planned |
| POST | `/calendar/events` | Событие + `resource_ids` + `recurrence` (конфликт брони → 409) | planned |
| PUT | `/calendar/events/:eventId` | Правка серии целиком | planned |
| DELETE | `/calendar/events/:eventId` | Серия → корзина 30 д, брони освобождаются | planned |
| GET | `/calendar/availability` | Busy/free по `user_ids` + одобренные отпуска HR | planned |
| GET | `/calendar/subscriptions` | Мои iCal-подписки | planned |
| POST | `/calendar/subscriptions` | Подписка на внешний фид (SSRF-валидация; >10 → 422) | planned |
| DELETE | `/calendar/subscriptions/:id` | Удалить подписку и её события | planned |
| GET | `/calendar/export.ics?token=` | Экспорт по секретному токену (замена текущего) | planned |
| POST | `/calendar/export/rotate` | Отозвать/перевыпустить токен | planned |

---

## HR (MOD-06, Phase 43)

HR-lite: заявки на отсутствие (отпуск, больничный, отгул). Согласование — system admin/founder. Одобренные заявки попадают в unified calendar (`type: leave`). Уведомления через NotificationService (PROD-01).

| Метод | Путь | Описание |
|---|---|---|
| GET | `/hr/leave` | Мои заявки |
| POST | `/hr/leave` | Создать заявку (`leave_type`, `start_date`, `end_date`, `reason?`) |
| GET | `/hr/leave/pending` | Очередь на согласование (admin) |
| POST | `/hr/leave/:leaveId/approve` | Одобрить (`version`, `review_note?`) |
| POST | `/hr/leave/:leaveId/reject` | Отклонить (`version`, `review_note?`) |
| POST | `/hr/leave/:leaveId/cancel` | Отменить свою pending-заявку (`version`) |

`leave_type`: `vacation`, `sick`, `personal`, `other`.  
`status`: `pending`, `approved`, `rejected`, `cancelled`.

---

## Marketplace (MOD-08, Phase 45)

Внутренняя доска объявлений: создание → mod queue (admin) → публикация. При approve — optional post в канал `#marketplace` (если существует глобальный канал с таким именем).

| Метод | Путь | Описание |
|---|---|---|
| GET | `/marketplace/listings` | Опубликованные объявления (`category?`, `page`, `per_page`) |
| GET | `/marketplace/listings/mine` | Мои объявления |
| GET | `/marketplace/listings/pending` | Очередь модерации (admin) |
| POST | `/marketplace/listings` | Создать (`title`, `description`, `category`, `price_cents?`) |
| PUT | `/marketplace/listings/:listingId` | Обновить (re-moderation если было approved) |
| POST | `/marketplace/listings/:listingId/approve` | Одобрить (`version`, `review_note?`) |
| POST | `/marketplace/listings/:listingId/reject` | Отклонить |
| POST | `/marketplace/listings/:listingId/sold` | Отметить проданным (author) |
| POST | `/marketplace/listings/:listingId/withdraw` | Снять с публикации |
| DELETE | `/marketplace/listings/:listingId` | Soft delete (author или admin) |

`category`: `equipment`, `books`, `services`, `furniture`, `other`.  
`status`: `pending`, `approved`, `rejected`, `sold`, `withdrawn`.

---

## CRM (MOD-05, Phase 42)

CRM-lite: контакты (отдельная сущность от User) и kanban pipeline сделок. Связь сделки с проектом — optional `project_id` (требуется членство в проекте).

**Доступ (RT-003):** `GET /crm/contacts` и `GET /crm/pipeline` для member — только свои записи (`created_by` / `owner_id`); system admin/founder — полный org-list (CRM-lite admin view). Update/Delete — owner или admin (A6-001).

| Метод | Путь | Описание |
|---|---|---|
| GET | `/crm/contacts` | Список контактов (owner-scoped; admin = org-wide) |
| POST | `/crm/contacts` | Создать контакт |
| PUT | `/crm/contacts/:contactId` | Обновить (optimistic `version`) |
| DELETE | `/crm/contacts/:contactId` | Soft delete |
| GET | `/crm/pipeline` | Kanban: сделки по stage (owner-scoped; admin = org-wide) |
| POST | `/crm/deals` | Создать сделку |
| PUT | `/crm/deals/:dealId` | Обновить / переместить stage |
| DELETE | `/crm/deals/:dealId` | Soft delete |

Stages: `lead`, `qualified`, `proposal`, `negotiation`, `won`, `lost`.

---

## Hub Base — no-code DB (V4-04, Phase 49)

Lark Base clone MVP: таблицы с typed columns и grid-редактором.

**Доступ (owner-scoped):** `GET /bases` возвращает только базы текущего пользователя; все операции над чужой базой (чтение/мутации базы, таблиц, колонок, записей) → `403 access denied`, несуществующая база → `404`.

| Метод | Путь | Описание |
|---|---|---|
| GET | `/bases` | Список баз текущего пользователя |
| POST | `/bases` | Создать базу (+ таблица Main, колонка Name) |
| GET | `/bases/:baseId` | Детали + список таблиц |
| PUT | `/bases/:baseId` | Обновить (`version`) |
| DELETE | `/bases/:baseId` | Soft delete |
| POST | `/bases/:baseId/tables` | Новая таблица |
| GET | `/bases/:baseId/tables/:tableId` | Колонки + записи (grid) |
| POST | `/bases/:baseId/tables/:tableId/columns` | Добавить колонку (`col_type`) |
| POST | `/bases/:baseId/tables/:tableId/records` | Добавить строку (`cells`) |
| PUT | `/bases/:baseId/tables/:tableId/records/:recordId` | Обновить ячейки |
| DELETE | `/bases/:baseId/tables/:tableId/records/:recordId` | Soft delete строки |

**Column types:** `text`, `number`, `date`, `select`, `checkbox`, `url`, `user_ref`. Cells — JSON `{ "<column_uuid>": value }`.

---

## Wiki — **planned/spec** (FR-WIKI, Волны 5/8)

> Новый модуль, кода ещё нет. Спецификация: `docs/system_specification.md` § «Wiki — новый модуль (FR-WIKI)».
> RBAC: роли space (read/write/admin) + наследование + restrict-флаг страницы (GZ-SEC-07/11).

| Метод | Путь | Описание | Статус |
|---|---|---|---|
| GET | `/wiki/spaces` | Мои space | planned |
| POST | `/wiki/spaces` | Создать space (создатель = admin) | planned |
| GET | `/wiki/spaces/:spaceId` | Space | planned |
| PUT | `/wiki/spaces/:spaceId` | Обновить (space admin, `version`) | planned |
| DELETE | `/wiki/spaces/:spaceId` | Удалить пустой space (иначе 422) | planned |
| GET | `/wiki/spaces/:spaceId/members` | Участники и роли | planned |
| PUT | `/wiki/spaces/:spaceId/members/:userId` | Назначить роль (последний admin → 422) | planned |
| DELETE | `/wiki/spaces/:spaceId/members/:userId` | Убрать участника | planned |
| GET | `/wiki/spaces/:spaceId/tree` | Дерево страниц (restrict-страницы скрыты) | planned |
| POST | `/wiki/spaces/:spaceId/pages` | Создать страницу (markdown, `parent_id`) | planned |
| GET | `/wiki/pages/:pageId` | Страница (read + restrict-check) | planned |
| PUT | `/wiki/pages/:pageId` | Правка (write, `version` → 409 при конфликте) | planned |
| DELETE | `/wiki/pages/:pageId` | Поддерево → корзина 30 д | planned |
| POST | `/wiki/pages/:pageId/move` | Переместить (цикл/глубина → 422) | planned |
| PUT | `/wiki/pages/:pageId/restrict` | Restrict-флаг + allow-list (space admin) | planned |
| GET | `/wiki/pages/:pageId/versions` | История версий (offset/limit) | planned |
| GET | `/wiki/pages/:pageId/versions/:number` | Версия + diff с текущей | planned |
| POST | `/wiki/pages/:pageId/versions/:number/restore` | Rollback → новая версия N+1 | planned |
| GET | `/wiki/pages/:pageId/issue-links` | Связанные задачи | planned |
| POST | `/wiki/pages/:pageId/issue-links` | Связать с задачей (403 без доступа к проекту) | planned |
| DELETE | `/wiki/pages/:pageId/issue-links/:issueId` | Убрать связь | planned |
| GET | `/projects/:projectId/issues/:issueId/wiki-links` | Обратная сторона (tracker) | planned |

Поиск — через существующий `GET /search?types=wiki_page` (user-scope фильтр на сервере, NF-03).

---

## Accounting (V4-03, Phase 51)

Hub-only double-entry ledger (admin/founder). Без НДФЛ и payroll.

| Метод | Путь | Описание |
|---|---|---|
| GET | `/accounting/accounts` | План счетов (auto-seed при первом вызове) |
| POST | `/accounting/accounts` | Добавить счёт |
| GET | `/accounting/journal` | Журнал проводок |
| POST | `/accounting/journal` | Создать черновик (balanced lines) |
| POST | `/accounting/journal/:entryId/post` | Провести |
| GET | `/accounting/trial-balance` | ОСВ по posted проводкам |
| GET | `/accounting/invoices` | Список счетов (draft/issued/paid) |
| POST | `/accounting/invoices` | Создать счёт (черновик) |
| POST | `/accounting/invoices/:invoiceId/issue` | Выставить счёт |
| POST | `/accounting/invoices/:invoiceId/pay` | Оплата → posted проводка 1010/4010 |
| GET | `/accounting/income-statement` | P&L: revenue, expense, net (cents) |

**CRM hook:** при переводе сделки в `won` с `amount_cents > 0` создаётся draft invoice (если ещё нет).

**Account types:** `asset`, `liability`, `equity`, `revenue`, `expense`. Amounts in **cents** (int64).

---

## AI Provider Hub (Phase 17, v3.0)

AI gateway, @hub-ai bot, ModelPicker, JQL assist. User routes требуют JWT; admin routes — system admin.

### User AI

| Метод | Путь | Описание |
|---|---|---|
| GET | `/ai/models` | Список доступных моделей (system + BYOK) |
| GET | `/ai/preferences` | Default model пользователя |
| PATCH | `/ai/preferences` | Сохранить default model + `allow_external_ai` (M3-030) |
| GET | `/ai/usage/today` | Дневная квота (tokens, requests, limit 500) |
| GET | `/ai/bot` | Системный пользователь @hub-ai (user_id для DM) |
| POST | `/ai/jql/assist` | NL → Hub JQL (`{"query": "..."}`) |
| POST | `/ai/composer/suggest` | Tab-дополнение сообщения в мессенджере (`{"prefix": "..."}`) |
| POST | `/ai/v1/chat/completions` | OpenAI-compatible proxy (stream → 202 + WS) |

### BYOK (user-scoped providers)

| Метод | Путь | Описание |
|---|---|---|
| GET | `/ai/my/providers` | Список user-scoped провайдеров |
| POST | `/ai/my/providers` | Добавить BYOK провайдер |
| PATCH | `/ai/my/providers/:id` | Обновить BYOK провайдер |

### Admin AI

| Метод | Путь | Описание |
|---|---|---|
| GET | `/admin/ai/providers` | Реестр провайдеров |
| POST | `/admin/ai/providers` | Создать провайдер |
| PATCH | `/admin/ai/providers/:id` | Обновить провайдер |
| POST | `/admin/ai/providers/:id/test` | Ping провайдера |
| GET | `/admin/ai/usage/summary` | Агрегат usage за сегодня |
| GET | `/admin/ai/usage/history` | История по дням (`?days=7`) |
| GET | `/admin/ai/settings` | Bot system prompt + `offline_only` flag |
| PATCH | `/admin/ai/settings` | Обновить system prompt |

### Admin 1C Connector (MOD-07)

| Метод | Путь | Описание |
|---|---|---|
| GET | `/admin/integrations/onec/settings` | Настройки webhook (token masked) |
| PUT | `/admin/integrations/onec/settings` | Сохранить URL, token, employee_id_field |
| POST | `/admin/integrations/onec/test` | Ping webhook (`event: ping`) |

При одобрении отпуска (`POST /hr/leave/:id/approve`) Hub POSTит `leave.approved` на настроенный URL, если `enabled=true`.

**Env:** нет — конфиг в `SystemSetting` (`integrations.onec.config`).


**Bot:** `@hub-ai` в каналах/DM; slash-команды `/ai summarize|test-cases|review|docs|explain|translate|sql|help`.

---

## Infrastructure (6 endpoints)

Служебные маршруты без JWT (мониторинг и документация).

| Метод | Путь | Описание |
|---|---|---|
| GET | `/live` | Liveness — процесс up → всегда 200 (K8s/Docker liveness; A14-001) |
| GET | `/ready` | Readiness — Postgres/Redis/Mongo/NATS → 200 или 503 (Compose healthcheck) |
| GET | `/health` | Alias `/ready` (обратная совместимость; не использовать как liveness) |
| GET | `/status` | Public status page JSON (Postgres, Redis, Mongo, NATS, API) — V4-06 |
| GET | `/metrics` | Prometheus scrape; prod: `METRICS_ALLOWED_CIDRS` обязателен (fail-closed если пусто; compose default `127.0.0.1/32`; non-prod пусто = без IP-фильтра) |
| GET | `/docs` | Swagger UI |
| GET | `/docs/doc.json` | OpenAPI JSON |

---

## Webhooks (ingress, 1 endpoint)

Входящие webhooks без JWT — только shared secret (Bearer или provider signature).

| Метод | Путь | Auth | Описание |
|---|---|---|---|
| POST | `/webhooks/alertmanager` | Bearer `ALERT_WEBHOOK_SECRET` | Alertmanager v4 webhook → audit log (`alert.firing` / `alert.resolved`). Disabled when secret empty. |

---

## Dev (3 endpoints, non-release only)

Доступны только при `DEV_MODE=true` (локальная разработка, не prod).

| Метод | Путь | Описание |
|---|---|---|
| GET | `/dev/status` | Статус dev-режима |
| POST | `/dev/seed` | Заполнить БД тестовыми данными |
| POST | `/dev/login` | Вход под dev-пользователем (`{"email": "..."}`) |

---

## WebSocket

Подключение к WebSocket для realtime-событий.

```
ws://host/ws?token=<jwt>
```

События: новые сообщения, обновления задач, присутствие, уведомления, звонки, typing-индикаторы.

---

## Формат ответов

### Успешный ответ

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Название",
  "created_at": "2026-03-22T12:00:00Z"
}
```

### Ошибка

```json
{
  "error": "Описание ошибки",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

Формат зафиксирован в `docs/tz_constraints.md` § API.

### Пагинация (cursor)

```json
{
  "items": [],
  "next_cursor": "HUB-10",
  "has_more": true,
  "total": 42
}
```

### Пагинация (offset)

```json
{
  "items": [],
  "page": 1,
  "per_page": 20,
  "total": 42
}
```

### HTTP-коды

| Код | Значение |
|---|---|
| 200 | OK |
| 201 | Created |
| 204 | No Content (удаление) |
| 400 | Bad Request |
| 401 | Unauthorized (нет токена / невалидный) |
| 403 | Forbidden (нет прав) |
| 404 | Not Found |
| 409 | Conflict (дубликат) |
| 429 | Too Many Requests (rate limit) |
| 500 | Internal Server Error |

---
*Последнее обновление: 2026-07-28 — Magic Link auth; error wire `{"error","request_id"}`; OpenAPI **252** paths / route inventory ~286*
