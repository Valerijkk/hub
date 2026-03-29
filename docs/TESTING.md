# Hub — Тестовая документация

## Стратегия тестирования

### Уровни тестирования

| Уровень | Инструменты | Покрытие |
|---------|-------------|----------|
| Unit-тесты (Backend) | Go testing + testify + mockery | Сервисный слой всех модулей |
| Unit-тесты (Frontend) | Vitest + RTL | Компоненты, хуки, stores |
| Интеграционные тесты | testcontainers-go | БД, Redis, NATS |
| API-тесты | Go test + httptest | Все HTTP-handler'ы |
| E2E-тесты | Playwright | Критические user flows |
| Build-тесты | TypeScript compiler + Vite build | Фронтенд: типизация + сборка |

### Принципы

1. **Каждый модуль покрыт unit-тестами** на уровне service layer
2. **Тесты изолированы**: mock'и для внешних зависимостей через mockery
3. **Интеграционные тесты** используют testcontainers для реальных PG/Redis/NATS
4. **Frontend**: TypeScript compilation гарантирует типобезопасность, Vite build проверяет сборку
5. **CI**: GitHub Actions запускает все тесты при push/PR

---

## Backend-тесты (Go)

### Покрытие по модулям

| Модуль | Файл | Что тестируется |
|--------|------|-----------------|
| auth | handler_test.go | HTTP-обработчики: magic link, verify, telegram login, refresh, logout, profile |
| auth | service_test.go | Бизнес-логика: генерация токенов, валидация, сессии, регистрация |
| project | service_test.go | Создание проекта, участники, invite, архивирование, удаление |
| tracker | service_test.go | CRUD задач, workflow, комментарии, связи, time tracking, rank |
| messenger | service_test.go | Каналы, сообщения, участники, DM, реакции |
| messenger | auto_channels_test.go | Автосоздание каналов при создании проекта |
| messenger | presence_test.go | Статусы присутствия, heartbeat, online/offline |
| messenger | unread_test.go | Подсчёт непрочитанных сообщений |
| tms | service_test.go | Тест-кейсы, прогоны, результаты, связи |
| tms | allure_parser_test.go | Парсинг Allure JSON, matching стратегии |
| sprint | service_test.go | Создание/запуск/завершение спринтов, burndown, velocity |
| calls | service_test.go | Создание звонков, join/leave, LiveKit токены |
| calls | chat_test.go | Эфемерный чат звонка через NATS |
| calls | recording_test.go | Запись: start/stop, egress, статусы |
| dashboard | service_test.go | Данные дашбордов, раскладки, шаблоны |
| notification | service_test.go | Создание уведомлений, preferences, quiet hours |
| search | service_test.go | Глобальный поиск, фильтрация, сохранённые фильтры |
| search | jql_test.go | JQL-парсер: синтаксис, операторы, валидация |
| github | service_test.go | Webhook обработка, автопривязка, branch creation |
| audit | service_test.go | Запись аудит-логов, фильтрация, пагинация |
| backup | service_test.go | Создание бэкапов, S3 upload, restore |
| apigateway | apikey_test.go | Генерация/валидация API-ключей, хеширование |

### Запуск тестов

```bash
# Все тесты
cd backend && go test ./...

# Конкретный модуль
go test ./internal/auth/...
go test ./internal/tracker/...

# С покрытием
go test -cover ./...

# С verbose output
go test -v ./internal/tms/...

# Только unit (без testcontainers)
go test -short ./...
```

---

## Frontend-тесты

### TypeScript Compilation

```bash
cd frontend && npx tsc --noEmit
```

Проверяет типобезопасность всего фронтенда без генерации файлов.

### Vite Build

```bash
cd frontend && npm run build
```

Проверяет:
- Корректность импортов
- Tree-shaking
- CSS compilation (Tailwind 4)
- Route generation (TanStack Router)

---

## API-тесты по модулям

### Формат: Ожидаемые HTTP-коды

#### Auth

| Endpoint | 200 | 400 | 401 | 404 | 429 |
|----------|-----|-----|-----|-----|-----|
| POST /auth/magic-link | Email отправлен | Невалидный email | - | - | Rate limit |
| POST /auth/magic-link/verify | Токены выданы | Невалидный код | - | - | Rate limit |
| POST /auth/telegram | Токены выданы | Невалидные данные | - | - | Rate limit |
| POST /auth/telegram-code | Токены выданы | Неверный код | - | - | Rate limit |
| POST /auth/refresh | Новый access token | Невалидный refresh | - | - | - |
| GET /auth/me | Профиль | - | Не авторизован | - | - |
| PUT /auth/me | Обновлён | Невалидные данные | Не авторизован | - | - |
| POST /auth/logout | Выход | - | Не авторизован | - | - |

#### Projects

| Endpoint | 200/201 | 400 | 401 | 403 | 404 |
|----------|---------|-----|-----|-----|-----|
| POST /projects | Создан | Невалидные данные | Нет токена | - | - |
| GET /projects | Список | - | Нет токена | - | - |
| GET /projects/:id | Проект | - | Нет токена | Не участник | Не найден |
| PUT /projects/:id | Обновлён | Невалидные данные | Нет токена | Нет прав | Не найден |
| DELETE /projects/:id | Удалён | - | Нет токена | Нет прав (Owner) | Не найден |
| POST /projects/:id/members | Добавлен | Невалидные данные | Нет токена | Нет прав | Не найден |

#### Tracker

| Endpoint | 200/201 | 400 | 401 | 403 | 404 | 409 |
|----------|---------|-----|-----|-----|-----|-----|
| POST /projects/:id/issues | Создана | Невалидные данные | Нет токена | RBAC | Проект не найден | - |
| PUT /projects/:id/issues/:key | Обновлена | Невалидные данные | Нет токена | RBAC | Задача не найдена | Version conflict |
| DELETE /projects/:id/issues/:key | Удалена | - | Нет токена | RBAC | Не найдена | - |

#### Messenger

| Endpoint | 200/201 | 400 | 401 | 403 |
|----------|---------|-----|-----|-----|
| POST /messenger/channels | Создан | Невалидные данные | Нет токена | - |
| POST /messenger/channels/:id/messages | Отправлено | Пустое сообщение | Нет токена | Не участник |
| POST /messenger/channels/:id/join | Присоединён | - | Нет токена | Приватный канал |
| POST /messenger/dm | DM создан/получен | Невалидный userId | Нет токена | - |

#### TMS

| Endpoint | 200/201 | 400 | 401 | 403 | 404 |
|----------|---------|-----|-----|-----|-----|
| POST /projects/:id/tms/cases | Создан | Невалидные данные | Нет токена | RBAC | Проект |
| POST /projects/:id/tms/runs | Создан | Невалидные данные | Нет токена | RBAC | Проект |
| POST /projects/:id/tms/allure/import | Импортировано | Невалидный JSON | Нет токена | RBAC | Проект |
| POST /projects/:id/tms/webhook/allure | Принято | Невалидный формат | API Key | - | Проект |

#### Calls

| Endpoint | 200/201 | 400 | 401 | 404 |
|----------|---------|-----|-----|-----|
| POST /calls | Создан | Нет channelId | Нет токена | Канал не найден |
| POST /calls/instant | Создан | - | Нет токена | - |
| POST /calls/:id/join | LiveKit token | - | Нет токена | Звонок не найден |
| POST /calls/:id/recording/start | Запись начата | - | Нет токена | Звонок не найден |

---

## Тест-кейсы по модулям

### AUTH — Авторизация

| ID | Название | Предусловие | Шаги | Ожидаемый результат |
|----|----------|-------------|------|---------------------|
| TC-AUTH-01 | Magic Link: успешная отправка | Валидный email | POST /auth/magic-link {email} | 200, email с кодом отправлен |
| TC-AUTH-02 | Magic Link: невалидный email | - | POST /auth/magic-link {email: "invalid"} | 400 |
| TC-AUTH-03 | Magic Link: верификация кода | Код отправлен | POST /auth/magic-link/verify {email, code} | 200, access + refresh token |
| TC-AUTH-04 | Magic Link: неверный код | Код отправлен | POST /auth/magic-link/verify {email, code: "wrong"} | 400/401 |
| TC-AUTH-05 | Telegram: авторизация по коду | Код получен от бота | POST /auth/telegram-code {code} | 200, access + refresh token |
| TC-AUTH-06 | Refresh: обновление токена | Валидный refresh token | POST /auth/refresh {refresh_token} | 200, новый access token |
| TC-AUTH-07 | Refresh: истёкший токен | Просроченный refresh | POST /auth/refresh {refresh_token} | 401 |
| TC-AUTH-08 | Профиль: получение | Авторизован | GET /auth/me | 200, данные профиля |
| TC-AUTH-09 | Профиль: обновление | Авторизован | PUT /auth/me {name, timezone} | 200, обновлён |
| TC-AUTH-10 | Сессии: список | Авторизован | GET /auth/sessions | 200, список сессий |
| TC-AUTH-11 | Сессии: отзыв | Авторизован, 2+ сессии | DELETE /auth/sessions/:id | 200, сессия удалена |
| TC-AUTH-12 | Rate limit: magic link | - | 4 запроса за 10 мин | 429 на 4-м запросе |
| TC-AUTH-13 | Онбординг: завершение | Авторизован, onboarding_completed=false | POST /auth/me/onboarding | 200, onboarding_completed=true |

### PROJ — Проекты

| ID | Название | Предусловие | Шаги | Ожидаемый результат |
|----|----------|-------------|------|---------------------|
| TC-PROJ-01 | Создание проекта | Авторизован | POST /projects {name, key} | 201, проект + workflow + доска + канал |
| TC-PROJ-02 | Дубликат ключа | Проект HUB существует | POST /projects {key: "HUB"} | 400/409, ключ занят |
| TC-PROJ-03 | Список проектов | Участник 2 проектов | GET /projects | 200, 2 проекта |
| TC-PROJ-04 | Архивирование | Owner проекта | POST /projects/:id/archive | 200, status=archived |
| TC-PROJ-05 | Удаление в корзину | Owner проекта | DELETE /projects/:id | 200, deleted_at set |
| TC-PROJ-06 | Восстановление | Проект в корзине | POST /projects/:id/restore | 200, deleted_at null |
| TC-PROJ-07 | Invite: создание | Owner/Developer | POST /projects/:id/invites | 201, код + ссылка |
| TC-PROJ-08 | Invite: присоединение | Валидный invite-код | POST /invites/join {code} | 200, участник добавлен |
| TC-PROJ-09 | Участник: смена роли | Owner | PUT /projects/:id/members/:userId {role: "tester"} | 200 |
| TC-PROJ-10 | Участник: удаление | Owner | DELETE /projects/:id/members/:userId | 200, удалён |

### TRCK — Трекер задач

| ID | Название | Предусловие | Шаги | Ожидаемый результат |
|----|----------|-------------|------|---------------------|
| TC-TRCK-01 | Создание задачи | Участник проекта | POST issues {title, type, priority} | 201, issue_key = HUB-1 |
| TC-TRCK-02 | Автоинкремент | HUB-1 существует | POST issues {title, type} | 201, issue_key = HUB-2 |
| TC-TRCK-03 | Обновление с version check | version=1 | PUT issues/:key {title, version: 1} | 200, version=2 |
| TC-TRCK-04 | Конфликт версий | version=2 в БД | PUT issues/:key {title, version: 1} | 409, conflict |
| TC-TRCK-05 | Комментарий с @mention | Задача существует | POST comments {body, mentions: [userId]} | 201, уведомление отправлено |
| TC-TRCK-06 | Связь задач | 2 задачи | POST issues/:key/links {target, type: "blocks"} | 201 |
| TC-TRCK-07 | Time tracking: manual | Задача существует | POST issues/:key/time {minutes: 60} | 201, time_spent += 60 |
| TC-TRCK-08 | Time tracking: timer | Задача существует | POST timer/start -> POST timer/stop | 201, лог с длительностью |
| TC-TRCK-09 | JQL поиск | Задачи существуют | POST issues/jql {query: "type = bug AND priority = high"} | 200, отфильтрованные задачи |
| TC-TRCK-10 | Workflow: смена статуса | Переход разрешён | PUT issues/:key {status_id} | 200, статус изменён |
| TC-TRCK-11 | Workflow: запрещённый переход | Переход не задан | PUT issues/:key {status_id} | 400, переход не разрешён |
| TC-TRCK-12 | Мягкое удаление | Задача существует | DELETE issues/:key | 200, deleted_at set |

### MSG — Мессенджер

| ID | Название | Предусловие | Шаги | Ожидаемый результат |
|----|----------|-------------|------|---------------------|
| TC-MSG-01 | Создание публичного канала | Авторизован | POST channels {type: "public", name} | 201 |
| TC-MSG-02 | DM дедупликация | DM между A и B существует | POST /dm {userId: B} | 200, тот же канал |
| TC-MSG-03 | Отправка сообщения | Участник канала | POST messages {content} | 201, NATS event |
| TC-MSG-04 | Реакция toggle | Сообщение существует | POST reactions {emoji: "thumbsup"} x2 | Добавлена, затем удалена |
| TC-MSG-05 | Pin сообщения | Участник канала | POST pin | 200, pinned |
| TC-MSG-06 | Тред | Сообщение существует | POST messages {reply_to: msgId} | 201, привязано к треду |
| TC-MSG-07 | Read-only канал | is_readonly=true, не Owner | POST messages | 403 |
| TC-MSG-08 | Mute канала | Участник | POST mute {until: "2026-04-01"} | 200, muted_until set |
| TC-MSG-09 | Unread counts | Новые сообщения | GET unread | 200, count > 0 |
| TC-MSG-10 | Поиск сообщений | Сообщения существуют | GET search?q=keyword | 200, результаты |

### TMS — Управление тестированием

| ID | Название | Предусловие | Шаги | Ожидаемый результат |
|----|----------|-------------|------|---------------------|
| TC-TMS-01 | Создание тест-кейса (classic) | Участник проекта | POST cases {title, case_type: "classic", steps} | 201, case_key = HUB-TC-1 |
| TC-TMS-02 | Создание тест-кейса (BDD) | Участник проекта | POST cases {case_type: "bdd", steps: [{keyword: "Given"}]} | 201 |
| TC-TMS-03 | Версионирование | Кейс существует | PUT cases/:key {title: "new"} -> GET versions | Версия 2 создана |
| TC-TMS-04 | Откат версии | Версия 1 и 2 | POST rollback/1 | 200, данные версии 1 |
| TC-TMS-05 | Создание прогона | Кейсы существуют | POST runs {title} + POST runs/:id/cases | 201, кейсы добавлены |
| TC-TMS-06 | Выполнение результата | Прогон in_progress | PUT results/:id {status: "passed"} | 200, passed_count++ |
| TC-TMS-07 | Allure import | Allure JSON | POST allure/import {multipart} | 200, прогон создан, matched |
| TC-TMS-08 | JUnit webhook | JUnit XML | POST webhook/junit | 200, результаты импортированы |
| TC-TMS-09 | Создание бага из failed | Результат failed | POST results/:id/bug | 201, задача Bug создана |
| TC-TMS-10 | Flaky report | Тесты с разными результатами | GET reports/flaky | 200, список flaky тестов |
| TC-TMS-11 | Матрица трассируемости | Кейсы привязаны к задачам | GET reports/traceability | 200, матрица |
| TC-TMS-12 | Экспорт PDF | Прогон completed | GET runs/:id/export | 200, PDF файл |

### CALL — Созвоны

| ID | Название | Предусловие | Шаги | Ожидаемый результат |
|----|----------|-------------|------|---------------------|
| TC-CALL-01 | Создание звонка | Канал существует | POST calls {channelId} | 201, room_name, LiveKit token |
| TC-CALL-02 | Instant-звонок | Авторизован | POST calls/instant | 201, invite_code |
| TC-CALL-03 | Join по invite-коду | Звонок active | GET calls/invite/:code | 200, информация + token |
| TC-CALL-04 | Leave | В звонке | POST calls/:id/leave | 200, participant_count-- |
| TC-CALL-05 | End call | В звонке | POST calls/:id/end | 200, status=ended |
| TC-CALL-06 | Start recording | В звонке | POST calls/:id/recording/start | 200, egress started |
| TC-CALL-07 | Stop recording | Запись идёт | POST calls/:id/recording/stop | 200, status=processing |
| TC-CALL-08 | Chat в звонке | В звонке | POST calls/:id/chat {text} | 200, NATS event |
| TC-CALL-09 | Raise hand | В звонке | POST calls/:id/raise-hand | 200, NATS event |
| TC-CALL-10 | Webhook: participant_joined | LiveKit event | POST calls/webhook | 200, participant_count++ |

### SPRINT — Спринты

| ID | Название | Предусловие | Шаги | Ожидаемый результат |
|----|----------|-------------|------|---------------------|
| TC-SPRINT-01 | Создание спринта | Участник проекта | POST sprints {name, start_date, end_date} | 201, status=planning |
| TC-SPRINT-02 | Запуск спринта | Спринт planning | POST sprints/:id/start | 200, status=active |
| TC-SPRINT-03 | Двойной active | Спринт уже active | POST sprints/:id2/start | 400, уже есть active |
| TC-SPRINT-04 | Burndown | Спринт active | GET sprints/:id/burndown | 200, точки по дням |
| TC-SPRINT-05 | Velocity | 3+ спринтов completed | GET velocity | 200, planned vs completed |
| TC-SPRINT-06 | Planning Poker | Спринт planning | POST poker -> vote -> reveal | 200, голоса раскрыты |
| TC-SPRINT-07 | Завершение спринта | Спринт active | POST sprints/:id/complete | 200, status=completed |

### NOTIF — Уведомления

| ID | Название | Предусловие | Шаги | Ожидаемый результат |
|----|----------|-------------|------|---------------------|
| TC-NOTIF-01 | Назначение задачи -> уведомление | Задача существует | PUT issue {assignee: userId} | Уведомление type=assigned |
| TC-NOTIF-02 | @mention -> уведомление | Комментарий с mention | POST comment {mentions: [userId]} | Уведомление type=mentioned |
| TC-NOTIF-03 | Mark as read | Уведомление unread | PATCH notifications/:id/read | 200, is_read=true |
| TC-NOTIF-04 | Mark all read | 5 unread | PATCH notifications/read-all | 200, все is_read=true |
| TC-NOTIF-05 | Quiet hours | quiet_hours set, не urgent | Событие в quiet hours | Уведомление отложено |
| TC-NOTIF-06 | Urgent пробивает quiet hours | quiet_hours set, is_urgent=true | Событие в quiet hours | Уведомление доставлено |
| TC-NOTIF-07 | Web Push subscribe | Авторизован | POST push/subscribe {endpoint, keys} | 200 |

### ADMIN — Администрирование

| ID | Название | Предусловие | Шаги | Ожидаемый результат |
|----|----------|-------------|------|---------------------|
| TC-ADMIN-01 | Аудит-лог | Admin | GET audit?action=login | 200, логи входов |
| TC-ADMIN-02 | Деактивация юзера | Admin | POST users/:id/deactivate | 200, status=deactivated |
| TC-ADMIN-03 | GDPR удаление | Admin | DELETE users/:id/gdpr | 200, все данные удалены |
| TC-ADMIN-04 | Бэкап | Admin | POST /admin/backups | 200, backup started |
| TC-ADMIN-05 | Restore | Admin, backup exists | POST /admin/backups/:id/restore | 200, восстановлено |

---

## Regression-тесты

### Критический путь (после каждого деплоя)

1. **Авторизация**: Magic Link -> verify -> получить токен -> GET /auth/me
2. **Проект**: Создать проект -> проверить workflow + канал создан
3. **Задача**: Создать задачу -> обновить статус -> добавить комментарий
4. **Мессенджер**: Отправить сообщение -> реакция -> тред
5. **TMS**: Создать кейс -> прогон -> результат passed
6. **Звонок**: Создать звонок -> получить LiveKit token
7. **Дашборд**: GET /dashboard/personal -> виджеты с данными
8. **Уведомления**: Назначить задачу -> проверить уведомление

### WebSocket Regression

1. Подключение /ws?token=JWT
2. Subscribe to project:{id}
3. Создать задачу через API
4. Получить event issue_created через WS
5. Unsubscribe

---

## Performance-тесты

### Целевые метрики

| Метрика | Целевое значение |
|---------|-----------------|
| API response time (p95) | < 200ms |
| WebSocket message latency | < 100ms |
| Kanban board load (100 задач) | < 500ms |
| Messenger: загрузка канала (1000 сообщений) | < 300ms |
| Глобальный поиск | < 500ms |
| Allure import (1000 результатов) | < 5s |
| Одновременных WS-подключений | 100+ |
| Участников в звонке | до 50 |

### Rate Limiting

| Тест | Ожидание |
|------|----------|
| 4 запроса auth за 10 мин | 4-й -> 429 |
| 101 API запрос за 1 сек | 101-й -> 429 |
| 11 file uploads за 1 мин | 11-й -> 429 |
| 301 сообщение за 1 мин | 301-е -> 429 |

---

## Accessibility-тесты

### Стандарт

WCAG 2.1 AA (целевой для v2.0 Redesign)

### Проверки

| Проверка | Метод | Статус |
|----------|-------|--------|
| Контраст текста (4.5:1) | Automated (axe-core) | Проверяется при UI review |
| Контраст UI элементов (3:1) | Automated (axe-core) | Проверяется при UI review |
| Keyboard navigation | Manual | Tab, Enter, Escape на всех interactive элементах |
| Screen reader labels | Manual | aria-label на кнопках без текста |
| Focus indicators | Visual | Все focusable элементы имеют видимый focus ring |
| Color не единственный индикатор | Visual | Иконки + текст в дополнение к цвету |

---

## UI Review результаты

### Процесс

После каждой фазы редизайна (v2.0) проводится UI review:
1. Screenshot каждой страницы (light + dark mode)
2. Сравнение с HTML-макетом из `редизайн/`
3. Pixel-perfect проверка: шрифты, цвета, отступы, радиусы, тени
4. Функциональное тестирование: все кнопки, формы, drag & drop работают
5. Responsive: desktop + mobile layout

### Покрытие (v2.0 Redesign)

| Страница | Макет | UI Review | Функционал |
|----------|-------|-----------|------------|
| Login | login.html | Passed | Passed |
| Dashboard | dashboard.html | Passed | Passed |
| Messenger | messenger.html | Passed | Passed |
| Kanban | kanban.html | Passed | Passed |
| Issue Detail | issue-detail.html | Passed | Passed |
| Calls | calls.html | Passed | Passed |
| TMS | tms.html | Passed | Passed |
| Settings | settings.html | Passed | Passed |
| Admin | admin.html | Passed | Passed |
| Onboarding | onboarding.html | Passed | Passed |
| Trash | trash.html | Passed | Passed |
| Projects list | (без макета) | Styled consistently | Passed |
| Sprint management | (без макета) | Styled consistently | Passed |
| JQL search | (без макета) | Styled consistently | Passed |
| Project settings | (без макета) | Styled consistently | Passed |
| Project dashboard | (без макета) | Styled consistently | Passed |

---

*Последнее обновление: 2026-03-28*
