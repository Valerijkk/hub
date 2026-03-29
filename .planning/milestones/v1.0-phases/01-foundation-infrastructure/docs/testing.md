# Тестовая документация: Phase 1 -- Фундамент и инфраструктура

## 1. Обзор

Документ содержит тест-кейсы для верификации всех требований Phase 1. Каждый тест-кейс привязан к конкретному requirement ID и содержит предусловия, шаги и ожидаемый результат.

**Формат ID тест-кейса:** `TC-{REQ_ID}-{NN}` (например, TC-AUTH-01-01).

---

## 2. Тест-кейсы по требованиям

### AUTH-01: Вход через Magic Link

**TC-AUTH-01-01: Успешная отправка Magic Link на валидный email**
- Предусловия: пользователь не аутентифицирован
- Шаги: `POST /api/auth/magic-link` с телом `{"email": "test@example.com"}`
- Ожидаемый результат: HTTP 200, тело `{"message": "Magic link sent"}`, код логируется в dev-консоль

**TC-AUTH-01-02: Успешная верификация кода**
- Предусловия: код отправлен на email
- Шаги: `POST /api/auth/magic-link/verify` с телом `{"email": "test@example.com", "code": "123456"}`
- Ожидаемый результат: HTTP 200, тело содержит `access_token`, `refresh_token`, `expires_in: 900`

**TC-AUTH-01-03: Неверный код**
- Предусловия: код отправлен на email
- Шаги: `POST /api/auth/magic-link/verify` с телом `{"email": "test@example.com", "code": "000000"}`
- Ожидаемый результат: HTTP 401, `"invalid or expired code"`

**TC-AUTH-01-04: Rate limit (4-й запрос за 10 минут)**
- Предусловия: отправлены 3 запроса Magic Link для одного email за 10 минут
- Шаги: `POST /api/auth/magic-link` с тем же email
- Ожидаемый результат: HTTP 429, `"too many requests, try again later"`

**TC-AUTH-01-05: Невалидный формат email**
- Шаги: `POST /api/auth/magic-link` с телом `{"email": "not-an-email"}`
- Ожидаемый результат: HTTP 400, `"invalid email format"`

**TC-AUTH-01-06: Код должен быть 6 символов**
- Шаги: `POST /api/auth/magic-link/verify` с телом `{"email": "test@example.com", "code": "12345"}`
- Ожидаемый результат: HTTP 400, `"code must be 6 characters"`

---

### AUTH-02: Вход через Telegram

**TC-AUTH-02-01: Успешный вход через Telegram**
- Предусловия: Telegram bot token настроен
- Шаги: `POST /api/auth/telegram` с телом `{"init_data": "<valid_init_data>"}`
- Ожидаемый результат: HTTP 200, тело содержит `access_token`, `refresh_token`

**TC-AUTH-02-02: Просроченная Telegram-аутентификация**
- Шаги: `POST /api/auth/telegram` с `init_data`, выданным более 24 часов назад
- Ожидаемый результат: HTTP 401, `"telegram auth expired"`

**TC-AUTH-02-03: Невалидная подпись Telegram**
- Шаги: `POST /api/auth/telegram` с модифицированным `init_data`
- Ожидаемый результат: HTTP 401, `"telegram auth failed"`

---

### AUTH-03: Режим регистрации

**TC-AUTH-03-01: Получение текущего режима (Admin)**
- Предусловия: пользователь с ролью Admin
- Шаги: `GET /api/admin/registration-mode`
- Ожидаемый результат: HTTP 200, `{"mode": "open"}` (или `"invite_only"`)

**TC-AUTH-03-02: Переключение на invite_only (Admin)**
- Предусловия: пользователь с ролью Admin
- Шаги: `PUT /api/admin/registration-mode` с телом `{"mode": "invite_only"}`
- Ожидаемый результат: HTTP 200, `{"mode": "invite_only"}`

**TC-AUTH-03-03: Запрет регистрации в режиме invite_only**
- Предусловия: режим `invite_only`, email не зарегистрирован
- Шаги: `POST /api/auth/magic-link` с телом `{"email": "new@example.com"}`
- Ожидаемый результат: HTTP 403, `"registration is by invite only"`

**TC-AUTH-03-04: Запрет доступа для не-Admin**
- Предусловия: пользователь с ролью Member
- Шаги: `PUT /api/admin/registration-mode` с телом `{"mode": "open"}`
- Ожидаемый результат: HTTP 403, `"admin access required"`

---

### AUTH-04: Профиль пользователя

**TC-AUTH-04-01: Получение профиля**
- Предусловия: аутентифицированный пользователь
- Шаги: `GET /api/auth/me`
- Ожидаемый результат: HTTP 200, JSON с полями `id`, `email`, `system_role`, `onboarding_completed`

**TC-AUTH-04-02: Обновление профиля**
- Предусловия: аутентифицированный пользователь
- Шаги: `PUT /api/auth/me` с телом `{"name": "Иван", "position": "Developer", "timezone": "Europe/Moscow"}`
- Ожидаемый результат: HTTP 200, JSON с обновлёнными полями

**TC-AUTH-04-03: Частичное обновление (PATCH-семантика)**
- Шаги: `PUT /api/auth/me` с телом `{"phone": "+79991234567"}` (остальные поля не указаны)
- Ожидаемый результат: HTTP 200, обновлён только `phone`, остальные без изменений

---

### AUTH-05: Системные роли

**TC-AUTH-05-01: Admin имеет доступ к admin-эндпоинтам**
- Предусловия: пользователь с `system_role: admin`
- Шаги: `GET /api/admin/registration-mode`
- Ожидаемый результат: HTTP 200

**TC-AUTH-05-02: Member не имеет доступ к admin-эндпоинтам**
- Предусловия: пользователь с `system_role: member`
- Шаги: `GET /api/admin/registration-mode`
- Ожидаемый результат: HTTP 403

---

### AUTH-06: Управление сессиями

**TC-AUTH-06-01: Список сессий**
- Предусловия: пользователь вошёл с 2 устройств
- Шаги: `GET /api/auth/sessions`
- Ожидаемый результат: HTTP 200, массив из 2+ сессий, одна с `is_current: true`

**TC-AUTH-06-02: Завершение другой сессии**
- Предусловия: 2 активных сессии
- Шаги: `DELETE /api/auth/sessions/:otherSessionId`
- Ожидаемый результат: HTTP 200, `{"message": "Session revoked"}`

**TC-AUTH-06-03: Refresh token rotation**
- Шаги: `POST /api/auth/refresh` с текущим refresh_token
- Ожидаемый результат: HTTP 200, новая пара access + refresh token, старый refresh token больше не валиден

**TC-AUTH-06-04: Logout**
- Шаги: `POST /api/auth/logout`
- Ожидаемый результат: HTTP 200, refresh token инвалидирован, повторный refresh возвращает 401

---

### AUTH-07: Уведомление о новой сессии

**TC-AUTH-07-01: Событие new_login при входе**
- Предусловия: пользователь уже подключён через WebSocket
- Шаги: вход с нового устройства
- Ожидаемый результат: через WS получено событие `new_login` с данными устройства и IP

---

### AUTH-08: Онбординг

**TC-AUTH-08-01: Завершение онбординга**
- Предусловия: новый пользователь с `onboarding_completed: false`
- Шаги: `POST /api/auth/me/onboarding` с телом `{"profile": {"name": "Иван"}}`
- Ожидаемый результат: HTTP 200, `onboarding_completed: true`

**TC-AUTH-08-02: Редирект на онбординг**
- Предусловия: пользователь с `onboarding_completed: false`, фронтенд загружен
- Шаги: попытка перейти на `/projects`
- Ожидаемый результат: редирект на `/onboarding`

---

### PROJ-01: Создание проекта

**TC-PROJ-01-01: Успешное создание проекта**
- Предусловия: аутентифицированный пользователь
- Шаги: `POST /api/projects` с телом `{"name": "Test Project", "key": "TEST"}`
- Ожидаемый результат: HTTP 201, JSON с `id`, `name`, `key`, `status: active`, `member_count: 1`, `my_role: owner`

**TC-PROJ-01-02: Дублирование ключа проекта**
- Предусловия: проект с ключом "TEST" уже существует
- Шаги: `POST /api/projects` с телом `{"name": "Another", "key": "TEST"}`
- Ожидаемый результат: HTTP 409 или 400, ошибка уникальности

**TC-PROJ-01-03: Валидация имени (слишком короткое)**
- Шаги: `POST /api/projects` с телом `{"name": "X", "key": "XP"}`
- Ожидаемый результат: HTTP 400

---

### PROJ-02: Управление проектом

**TC-PROJ-02-01: Список проектов**
- Предусловия: пользователь является участником 2 проектов
- Шаги: `GET /api/projects`
- Ожидаемый результат: HTTP 200, массив из 2 проектов

**TC-PROJ-02-02: Обновление проекта**
- Предусловия: пользователь является Owner
- Шаги: `PUT /api/projects/:id` с телом `{"name": "Updated Name"}`
- Ожидаемый результат: HTTP 200, `name: "Updated Name"`

**TC-PROJ-02-03: Получение деталей проекта**
- Шаги: `GET /api/projects/:id`
- Ожидаемый результат: HTTP 200, JSON с полями `id`, `name`, `key`, `status`, `member_count`, `created_at`, `updated_at`

---

### PROJ-03: Роли в проекте

**TC-PROJ-03-01: Добавление участника с ролью**
- Предусловия: Owner проекта
- Шаги: `POST /api/projects/:id/members` с телом `{"user_id": "...", "role": "developer"}`
- Ожидаемый результат: HTTP 201, участник добавлен

**TC-PROJ-03-02: Смена роли участника**
- Предусловия: Owner проекта
- Шаги: `PUT /api/projects/:id/members/:userId` с телом `{"role": "tester"}`
- Ожидаемый результат: HTTP 200, роль обновлена

**TC-PROJ-03-03: Viewer не может редактировать проект**
- Предусловия: пользователь с ролью Viewer
- Шаги: `PUT /api/projects/:id` с телом `{"name": "Hack"}`
- Ожидаемый результат: HTTP 403

**TC-PROJ-03-04: Удаление участника**
- Предусловия: Owner проекта
- Шаги: `DELETE /api/projects/:id/members/:userId`
- Ожидаемый результат: HTTP 204

---

### PROJ-04: Приглашения и команды

**TC-PROJ-04-01: Создание invite-ссылки**
- Предусловия: участник с правом InviteCreate
- Шаги: `POST /api/projects/:id/invites` с телом `{"role": "developer", "max_uses": 5}`
- Ожидаемый результат: HTTP 201, JSON с `code`, `link`, `role`, `max_uses`

**TC-PROJ-04-02: Присоединение по invite-коду**
- Предусловия: валидный invite-код
- Шаги: `POST /api/invites/join` с телом `{"code": "abc12345"}`
- Ожидаемый результат: HTTP 200, пользователь стал участником проекта

**TC-PROJ-04-03: Просроченная ссылка**
- Предусловия: invite с истёкшим `expires_at`
- Шаги: `POST /api/invites/join` с кодом просроченного invite
- Ожидаемый результат: HTTP 400 или 410

**TC-PROJ-04-04: Превышение лимита использований**
- Предусловия: invite с `max_uses: 1`, уже использован 1 раз
- Шаги: `POST /api/invites/join` с тем же кодом
- Ожидаемый результат: HTTP 400 или 410

---

### PROJ-06: Архивирование

**TC-PROJ-06-01: Архивирование проекта**
- Предусловия: активный проект
- Шаги: `POST /api/projects/:id/archive`
- Ожидаемый результат: HTTP 200, `status: archived`

**TC-PROJ-06-02: Разархивирование проекта**
- Предусловия: архивированный проект
- Шаги: `POST /api/projects/:id/unarchive`
- Ожидаемый результат: HTTP 200, `status: active`

**TC-PROJ-06-03: Архивные проекты не в стандартном списке**
- Предусловия: один активный и один архивный проект
- Шаги: `GET /api/projects` (без параметров)
- Ожидаемый результат: только активный проект в списке

---

### PROJ-07: Корзина

**TC-PROJ-07-01: Удаление проекта (soft delete)**
- Шаги: `DELETE /api/projects/:id`
- Ожидаемый результат: HTTP 200, `"Project moved to trash"`, проект не в списке активных

**TC-PROJ-07-02: Список удалённых проектов**
- Шаги: `GET /api/projects/trash`
- Ожидаемый результат: HTTP 200, массив с удалённым проектом

**TC-PROJ-07-03: Восстановление из корзины**
- Предусловия: проект в корзине
- Шаги: `POST /api/projects/:id/restore`
- Ожидаемый результат: HTTP 200, `status: active`

---

### UI-01: Навигация

**TC-UI-01-01: Sidebar отображает проекты**
- Предусловия: пользователь аутентифицирован, есть 2 проекта
- Шаги: открыть приложение
- Ожидаемый результат: sidebar показывает оба проекта

**TC-UI-01-02: Sidebar сворачивается**
- Шаги: нажать кнопку сворачивания sidebar
- Ожидаемый результат: sidebar сворачивается до иконок

**TC-UI-01-03: Topbar показывает профиль**
- Шаги: посмотреть на topbar
- Ожидаемый результат: аватар/инициалы пользователя, меню при клике

---

### UI-07: Обработка ошибок

**TC-UI-07-01: Toast при ошибке API**
- Шаги: выполнить действие, вызывающее ошибку API (например, создание проекта с дублирующимся ключом)
- Ожидаемый результат: toast-уведомление с сообщением об ошибке

**TC-UI-07-02: Страница 404**
- Шаги: перейти на несуществующий маршрут `/nonexistent`
- Ожидаемый результат: страница "404 -- Страница не найдена" с кнопкой "На главную"

**TC-UI-07-03: Error Boundary**
- Шаги: спровоцировать ошибку рендеринга компонента
- Ожидаемый результат: страница "Произошла ошибка" с кнопкой "Обновить"

---

### UI-08: Файлы

**TC-UI-08-01: Получение presigned URL для загрузки**
- Предусловия: пользователь с правом FileUpload
- Шаги: `POST /api/files/upload-url` с телом `{"filename": "test.png", "contentType": "image/png", "projectId": "..."}`
- Ожидаемый результат: HTTP 200, JSON с presigned upload URL и `fileId`

**TC-UI-08-02: Подтверждение загрузки**
- Предусловия: файл загружен по presigned URL
- Шаги: `POST /api/files/:fileId/confirm` с телом `{"entityType": "task", "entityId": "..."}`
- Ожидаемый результат: HTTP 200, `{"status": "confirmed"}`

**TC-UI-08-03: Получение URL для скачивания**
- Шаги: `GET /api/files/:fileId/url`
- Ожидаемый результат: HTTP 200, `{"downloadUrl": "..."}`

**TC-UI-08-04: Удаление файла**
- Предусловия: пользователь с правом FileDelete
- Шаги: `DELETE /api/files/:fileId`
- Ожидаемый результат: HTTP 200, `{"status": "deleted"}`

---

## 3. Матрица RBAC тестов

Проверка доступа по ролям к ключевым эндпоинтам:

| Эндпоинт | Метод | Admin | Owner | Developer | Tester | Viewer |
|----------|-------|-------|-------|-----------|--------|--------|
| /api/projects | POST | 200 | 200 | 200 | 200 | 200 |
| /api/projects/:id | GET | 200 | 200 | 200 | 200 | 200 |
| /api/projects/:id | PUT | 200 | 200 | 403 | 403 | 403 |
| /api/projects/:id | DELETE | 200 | 200 | 403 | 403 | 403 |
| /api/projects/:id/archive | POST | 200 | 200 | 403 | 403 | 403 |
| /api/projects/:id/members | POST | 200 | 200 | 403 | 403 | 403 |
| /api/projects/:id/members/:uid | PUT | 200 | 200 | 403 | 403 | 403 |
| /api/projects/:id/members/:uid | DELETE | 200 | 200 | 403 | 403 | 403 |
| /api/projects/:id/invites | POST | 200 | 200 | 200 | 403 | 403 |
| /api/projects/:id/invites/:iid | DELETE | 200 | 200 | 200 | 403 | 403 |
| /api/files/upload-url | POST | 200 | 200 | 200 | 200 | 403 |
| /api/files/:id | DELETE | 200 | 200 | 200 | 200 | 403 |
| /api/admin/registration-mode | GET | 200 | 403 | 403 | 403 | 403 |
| /api/admin/registration-mode | PUT | 200 | 403 | 403 | 403 | 403 |

**Примечания:**
- Admin обходит RBAC-проверки на уровне проекта (системная роль).
- Owner имеет все 26 bitmap-разрешений (`^Permission(0)`).
- Developer имеет `InviteCreate`, но не `InviteManage`.
- Tester не имеет `InviteCreate`.
- Viewer имеет только `ProjectView`.

---

## 4. Тестовое окружение

### 4.1. Инфраструктура

- **Docker Compose:** `deployments/docker-compose.yml` (PostgreSQL 17, MongoDB 8, Redis 7.4, NATS 2.11, SeaweedFS, Caddy 2)
- **Команда запуска:** `docker compose -f deployments/docker-compose.yml up -d`

### 4.2. Бэкенд-тесты

- **Фреймворк:** Go testing + testify + testcontainers + mockery
- **Запуск:** `cd backend && go test -race -count=1 ./...`
- **Покрытие:** `cd backend && go test -race -coverprofile=coverage.out ./...`

### 4.3. Фронтенд-тесты

- **Фреймворк:** Vitest + React Testing Library
- **Запуск:** `cd frontend && npx vitest run`
- **E2E (будущее):** Playwright

### 4.4. CI Pipeline

- **backend.yml:** lint (golangci-lint) -> test (go test -race) -> build
- **frontend.yml:** lint (biome) -> test (vitest) -> build (vite)
