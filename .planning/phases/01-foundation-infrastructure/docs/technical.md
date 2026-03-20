# Техническая документация: Phase 1 -- Фундамент и инфраструктура

## 1. Архитектура

### 1.1. Общая схема

Hub -- монорепозиторий с модульной архитектурой:

```
hub/
  backend/         -- Go бэкенд (Fiber v3)
    cmd/server/    -- точка входа
    ent/schema/    -- Ent ORM схемы
    internal/
      auth/        -- аутентификация (Magic Link, Telegram, JWT, профиль, онбординг)
      project/     -- проекты (CRUD, участники, роли, команды, инвайты, архив, корзина)
      rbac/        -- ролевая модель (bitmap-разрешения, Redis-кеш, middleware)
      storage/     -- файловое хранилище (SeaweedFS presigned URLs, метаданные)
      ws/          -- WebSocket gateway (hub, NATS bridge, протокол)
      platform/    -- инфраструктура (конфигурация, клиенты БД, NATS, Redis, middleware)
  frontend/        -- React 19 + TypeScript + Vite 6
    src/
      routes/      -- TanStack Router file-based маршруты
      components/  -- UI-компоненты (Shadcn/ui)
      lib/         -- API-клиент, утилиты
      stores/      -- Zustand сторы
      hooks/       -- React-хуки
  deployments/     -- Docker Compose, Caddyfile
  .github/workflows/ -- CI пайплайны
```

### 1.2. Стек технологий

| Технология | Версия | Назначение |
|-----------|--------|------------|
| Go | 1.23+ | Язык бэкенда |
| Fiber | v3 | HTTP-фреймворк (fasthttp) |
| Ent | latest | ORM с кодогенерацией |
| Atlas | latest | Миграции БД |
| PostgreSQL | 17 | Основная реляционная БД |
| MongoDB | 8 | Документная БД (чаты, Allure) |
| Redis | 7.4 | Кеш, сессии, rate limiting |
| NATS | 2.11 | Pub/sub + JetStream |
| SeaweedFS | latest | S3-совместимое файловое хранилище |
| Caddy | 2 | Reverse proxy, авто-HTTPS |
| React | 19 | UI-фреймворк |
| TypeScript | 5.x | Типизация фронтенда |
| Vite | 6 | Сборщик фронтенда |
| TanStack Router | latest | File-based type-safe маршрутизация |
| TanStack Query | latest | Серверный стейт |
| Shadcn/ui | latest | UI-компоненты (Radix + Tailwind) |
| Tailwind CSS | 4 | CSS-фреймворк |
| Zustand | latest | Клиентский стейт |
| golang-jwt/jwt | v5 | JWT-токены (HS256) |
| aws-sdk-go-v2 | latest | S3 SDK |
| Resend | API | Email-провайдер (Magic Link) |
| Biome | latest | Линтер/форматтер фронтенда |
| golangci-lint | latest | Линтер бэкенда |

---

## 2. База данных

### 2.1. PostgreSQL (Ent ORM)

Все Ent-схемы расположены в `backend/ent/schema/`. Ниже описание каждой.

#### User

Таблица пользователей системы.

| Поле | Тип | Constraints | Описание |
|------|-----|-------------|----------|
| id | UUID | PK, immutable, default: uuid.New | Идентификатор |
| email | string | unique, not empty | Email (уникальный) |
| name | string | optional | Имя |
| display_name | string | optional | Отображаемое имя |
| avatar_url | string | optional | URL аватара |
| telegram_id | string | optional, unique, nillable | Telegram ID |
| phone | string | optional | Телефон |
| position | string | optional | Должность |
| department | string | optional | Отдел |
| timezone | string | default: "UTC" | Часовой пояс |
| language | string | default: "ru" | Язык |
| birthday | time | optional, nillable | Дата рождения |
| system_role | enum(admin, member) | default: "member" | Системная роль |
| onboarding_completed | bool | default: false | Флаг онбординга |
| created_at | time | immutable, default: now | Дата создания |
| updated_at | time | update_default: now | Дата обновления |

**Edges:** sessions, project_members, team_members, files (FK: uploader_id), invites.
**Indexes:** telegram_id.

#### Session

Сессии пользователей (JWT refresh tokens).

| Поле | Тип | Constraints | Описание |
|------|-----|-------------|----------|
| id | UUID | PK, immutable | Идентификатор |
| refresh_token | string | unique, sensitive | Refresh token (не возвращается в API) |
| device | string | required | User-Agent устройства |
| ip_address | string | optional | IP-адрес |
| last_active_at | time | default: now | Последняя активность |
| expires_at | time | required | Срок действия (30 дней) |
| created_at | time | immutable | Дата создания |

**Edges:** user (required, unique -- belongs to User).
**Indexes:** expires_at.

#### Project

Проекты.

| Поле | Тип | Constraints | Описание |
|------|-----|-------------|----------|
| id | UUID | PK, immutable | Идентификатор |
| name | string | not empty, min=2, max=100 | Название |
| key | string | unique, not empty, regexp(^[A-Z][A-Z0-9]+$) | Ключ проекта (для нумерации задач) |
| description | string | optional, max=2000 | Описание |
| avatar_url | string | optional | URL аватара |
| status | enum(active, archived, deleted) | default: "active" | Статус |
| task_counter | int | default: 0 | Счётчик задач (для нумерации) |
| deleted_at | time | optional, nillable | Дата удаления (soft delete) |
| created_at | time | immutable | Дата создания |
| updated_at | time | update_default: now | Дата обновления |

**Edges:** members, roles, teams, invites, files.
**Indexes:** status, deleted_at.

#### ProjectMember

Связь участника с проектом (many-to-many через junction table).

| Поле | Тип | Constraints | Описание |
|------|-----|-------------|----------|
| id | UUID | PK, immutable | Идентификатор |
| project_role | enum(owner, developer, tester, viewer) | default: "developer" | Роль в проекте |
| joined_at | time | immutable | Дата вступления |

**Edges:** user (required), project (required), custom_role (optional, -> Role).
**Indexes:** unique(user, project) -- один пользователь = одна роль в проекте.

#### Role

Кастомные роли проекта с bitmap-разрешениями.

| Поле | Тип | Constraints | Описание |
|------|-----|-------------|----------|
| id | UUID | PK, immutable | Идентификатор |
| name | string | not empty | Название роли |
| description | string | optional | Описание |
| permissions | uint64 | default: 0 | Bitmap разрешений |
| is_system | bool | default: false | Системная роль (неизменяемая) |
| created_at | time | immutable | Дата создания |

**Edges:** project (required), members (-> ProjectMember).

#### Team

Команды (группы пользователей).

| Поле | Тип | Constraints | Описание |
|------|-----|-------------|----------|
| id | UUID | PK, immutable | Идентификатор |
| name | string | not empty | Название |
| description | string | optional | Описание |
| created_at | time | immutable | Дата создания |

**Edges:** projects (many-to-many <- Project.teams), members (-> TeamMember).

#### TeamMember

Связь пользователя с командой.

| Поле | Тип | Constraints | Описание |
|------|-----|-------------|----------|
| id | UUID | PK, immutable | Идентификатор |
| joined_at | time | immutable | Дата вступления |

**Edges:** team (required), user (required).
**Indexes:** unique(team, user).

#### Invite

Invite-ссылки для присоединения к проекту.

| Поле | Тип | Constraints | Описание |
|------|-----|-------------|----------|
| id | UUID | PK, immutable | Идентификатор |
| code | string | unique, not empty | Уникальный код (8 символов) |
| max_uses | int | default: 0 (0 = безлимит) | Максимум использований |
| use_count | int | default: 0 | Текущее количество использований |
| project_role | enum(developer, tester, viewer) | default: "developer" | Роль при присоединении |
| expires_at | time | optional, nillable | Срок действия |
| created_at | time | immutable | Дата создания |

**Edges:** project (required), created_by (-> User, required).
**Indexes:** expires_at.

#### FileMetadata

Метаданные загруженных файлов.

| Поле | Тип | Constraints | Описание |
|------|-----|-------------|----------|
| id | UUID | PK, immutable | Идентификатор |
| filename | string | not empty | Оригинальное имя файла |
| content_type | string | not empty | MIME-тип |
| size | int64 | required | Размер в байтах |
| s3_key | string | not empty | Ключ в S3 |
| s3_bucket | string | not empty | Бакет в S3 |
| entity_type | string | optional | Тип привязанной сущности (task, message, avatar...) |
| entity_id | string | optional | ID привязанной сущности |
| status | enum(pending, confirmed, deleted) | default: "pending" | Статус файла |
| created_at | time | immutable | Дата создания |

**Edges:** uploader (-> User), project (-> Project).
**Indexes:** (entity_type, entity_id), status, s3_key.

#### OutboxEvent

Transactional outbox для консистентной доставки событий.

| Поле | Тип | Constraints | Описание |
|------|-----|-------------|----------|
| id | UUID | PK, immutable | Идентификатор |
| event_type | string | not empty | Тип события (NATS subject) |
| payload | bytes | required | JSON-сериализованные данные |
| status | enum(pending, published, failed) | default: "pending" | Статус |
| attempts | int | default: 0 | Количество попыток |
| last_error | string | optional | Последняя ошибка |
| created_at | time | immutable | Дата создания |
| published_at | time | optional, nillable | Дата публикации |

**Indexes:** (status, created_at) -- для выборки pending-событий.

#### SystemSetting

Глобальные настройки системы (key-value).

| Поле | Тип | Constraints | Описание |
|------|-----|-------------|----------|
| id | UUID | PK, immutable | Идентификатор |
| key | string | unique, not empty | Ключ настройки |
| value | string | required | Значение |
| description | string | optional | Описание |
| updated_at | time | update_default: now | Дата обновления |

**Используемые ключи:** `registration_mode` (значения: `open`, `invite_only`).

### 2.2. MongoDB

В Phase 1 MongoDB настроена и подключена, но активно не используется. Применяется начиная с Phase 3 (чаты) и Phase 4 (Allure-отчёты).

- **Коллекции (будущие):** `messages`, `allure_results`
- **Подключение:** `mongo:8`, аутентификация через `MONGO_INITDB_ROOT_USERNAME/PASSWORD`

### 2.3. Redis

Используется для:

- **Magic Link коды:** ключ `magic:{email}`, значение -- хеш кода, TTL 5 минут
- **Rate limiting:** ключ `ratelimit:magic:{email}`, sliding window 10 минут, лимит 3
- **RBAC-кеш:** ключ `rbac:{userID}:{projectID}`, значение -- bitmap разрешений (uint64), TTL 5 минут
- **Сессии:** refresh tokens (косвенно через PG Session table; Redis может использоваться для быстрой проверки)

---

## 3. API Endpoints

Все эндпоинты доступны по префиксу `/api`. Аутентификация через Bearer JWT в заголовке `Authorization`.

### 3.1. Auth

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| POST | /api/auth/magic-link | - | Отправка Magic Link кода |
| POST | /api/auth/magic-link/verify | - | Верификация кода, получение токенов |
| POST | /api/auth/telegram | - | Вход через Telegram |
| POST | /api/auth/refresh | - | Обновление токенов |
| POST | /api/auth/logout | JWT | Завершение сессии |
| GET | /api/auth/me | JWT | Получение профиля |
| PUT | /api/auth/me | JWT | Обновление профиля |
| POST | /api/auth/me/onboarding | JWT | Завершение онбординга |
| GET | /api/auth/sessions | JWT | Список сессий |
| DELETE | /api/auth/sessions/:sessionId | JWT | Завершение сессии |

### 3.2. Admin

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| GET | /api/admin/registration-mode | JWT + Admin | Получение режима регистрации |
| PUT | /api/admin/registration-mode | JWT + Admin | Установка режима регистрации |

### 3.3. Projects

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| POST | /api/projects | JWT | Создание проекта |
| GET | /api/projects | JWT | Список проектов пользователя |
| GET | /api/projects/trash | JWT | Список удалённых проектов |
| GET | /api/projects/:projectId | JWT | Детали проекта |
| PUT | /api/projects/:projectId | JWT + RBAC | Обновление проекта |
| POST | /api/projects/:projectId/archive | JWT + RBAC | Архивирование |
| POST | /api/projects/:projectId/unarchive | JWT + RBAC | Разархивирование |
| DELETE | /api/projects/:projectId | JWT + RBAC | Удаление (soft) |
| POST | /api/projects/:projectId/restore | JWT + RBAC | Восстановление |

### 3.4. Members

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| GET | /api/projects/:projectId/members | JWT | Список участников |
| POST | /api/projects/:projectId/members | JWT + RBAC | Добавление участника |
| PUT | /api/projects/:projectId/members/:userId | JWT + RBAC | Смена роли |
| DELETE | /api/projects/:projectId/members/:userId | JWT + RBAC | Удаление участника |

### 3.5. Invites

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| POST | /api/projects/:projectId/invites | JWT + RBAC | Создание invite |
| GET | /api/projects/:projectId/invites | JWT | Список invites |
| DELETE | /api/projects/:projectId/invites/:inviteId | JWT + RBAC | Отзыв invite |
| POST | /api/invites/join | JWT | Присоединение по коду |

### 3.6. Teams (зарезервированы, HTTP 501)

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| POST | /api/teams | JWT | Создание команды |
| GET | /api/teams | JWT | Список команд |
| POST | /api/teams/:teamId/members | JWT | Добавление участника |
| DELETE | /api/teams/:teamId/members/:userId | JWT | Удаление участника |
| POST | /api/teams/:teamId/projects/:projectId | JWT | Привязка к проекту |

### 3.7. Files

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| POST | /api/files/upload-url | JWT + RBAC (FileUpload) | Получение presigned upload URL |
| POST | /api/files/:fileId/confirm | JWT | Подтверждение загрузки |
| GET | /api/files/:fileId/url | JWT | Получение presigned download URL |
| DELETE | /api/files/:fileId | JWT + RBAC (FileDelete) | Удаление файла |

### 3.8. WebSocket

| Протокол | Путь | Auth | Описание |
|----------|------|------|----------|
| WS | /ws?token={jwt} | JWT (query param) | WebSocket-соединение |

**Протокол клиент -> сервер:**
```json
{"type": "subscribe", "room": "project:{id}"}
{"type": "unsubscribe", "room": "project:{id}"}
{"type": "ping"}
```

**Протокол сервер -> клиент:**
```json
{"type": "event", "event": "project.updated", "room": "project:{id}", "payload": {...}}
{"type": "pong"}
```

---

## 4. RBAC

### 4.1. Система разрешений

Bitmap uint64 с 26 permission-флагами:

```go
PermProjectView          = 1 << 0   // 1
PermProjectEdit          = 1 << 1   // 2
PermProjectDelete        = 1 << 2   // 4
PermProjectManageMembers = 1 << 3   // 8
PermProjectManageRoles   = 1 << 4   // 16
PermProjectManageWorkflow= 1 << 5   // 32
PermTaskCreate           = 1 << 6   // 64
PermTaskEdit             = 1 << 7   // 128
PermTaskEditAny          = 1 << 8   // 256
PermTaskDelete           = 1 << 9   // 512
PermTaskAssign           = 1 << 10  // 1024
PermTaskComment          = 1 << 11  // 2048
PermFileUpload           = 1 << 12  // 4096
PermFileDelete           = 1 << 13  // 8192
PermFileDeleteAny        = 1 << 14  // 16384
PermChannelCreate        = 1 << 15  // 32768
PermChannelManage        = 1 << 16  // 65536
PermMessageEdit          = 1 << 17  // 131072
PermMessageDeleteAny     = 1 << 18  // 262144
PermTestCaseCreate       = 1 << 19  // 524288
PermTestCaseEdit         = 1 << 20  // 1048576
PermTestRunCreate        = 1 << 21  // 2097152
PermTestRunExecute       = 1 << 22  // 4194304
PermInviteCreate         = 1 << 23  // 8388608
PermInviteManage         = 1 << 24  // 16777216
PermTeamManage           = 1 << 25  // 33554432
```

### 4.2. Иерархия ролей

Два уровня:

1. **Системные роли** (User.system_role):
   - `admin` -- полный доступ ко всем проектам и настройкам
   - `member` -- стандартный доступ, может создавать проекты

2. **Проектные роли** (ProjectMember.project_role):
   - `owner` -- все 26 разрешений (`^Permission(0)` = 0xFFFFFFFFFFFFFFFF)
   - `developer` -- `ProjectView | TaskCreate | TaskEdit | TaskAssign | TaskComment | FileUpload | FileDelete | ChannelCreate | MessageEdit | TestCaseCreate | TestCaseEdit | TestRunCreate | TestRunExecute | InviteCreate`
   - `tester` -- `ProjectView | TaskCreate | TaskEdit | TaskComment | FileUpload | FileDelete | MessageEdit | TestCaseCreate | TestCaseEdit | TestRunCreate | TestRunExecute`
   - `viewer` -- `ProjectView`
   - кастомная роль -- произвольная комбинация через таблицу `Role`

### 4.3. Redis-кеш

- **Ключ:** `rbac:{userID}:{projectID}`
- **Значение:** uint64 (bitmap эффективных разрешений)
- **TTL:** 5 минут
- **Инвалидация:** автоматическая при любом изменении членства или ролей через `CacheInvalidator` интерфейс

### 4.4. Middleware

Три уровня middleware:

1. **RequireAuth()** -- проверка JWT токена, установка `userID` в Locals
2. **RequireSystemRole(role, client)** -- проверка системной роли (admin/member)
3. **RequirePermission(perm, rbacService)** -- проверка bitmap-разрешения для проекта из параметра `:projectId`

---

## 5. Безопасность

### 5.1. JWT

- **Алгоритм:** HS256 (HMAC-SHA256)
- **Библиотека:** golang-jwt/jwt/v5
- **Access token TTL:** 15 минут
- **Claims:** `userID` (UUID), `sessionID` (UUID), `exp`, `iat`

### 5.2. Refresh Token

- **Формат:** UUID v4
- **Хранение:** PostgreSQL таблица Session
- **TTL:** 30 дней
- **Ротация:** при каждом refresh старый токен инвалидируется, выдаётся новый
- **Отзыв:** через DELETE /api/auth/sessions/:sessionId

### 5.3. Magic Link

- **Код:** 6 случайных цифр (crypto/rand)
- **Хранение:** Redis, ключ `magic:{email}`, TTL 5 минут
- **Rate limit:** 3 запроса за 10 минут на email (sliding window в Redis)
- **Доставка:** Resend API (dev-режим: логирование в stdout)

### 5.4. Telegram-аутентификация

- **Верификация:** HMAC-SHA256 подпись initData с использованием SHA256(bot_token) как ключа
- **Срок действия:** initData считается просроченным через 24 часа
- **Привязка:** автоматическая по telegram_id

### 5.5. Rate Limiting

| Контекст | Лимит | Окно | Реализация |
|----------|-------|------|------------|
| Magic Link | 3 запроса | 10 минут | Redis sliding window per email |
| API (общий) | 100 запросов | 1 секунда | In-memory sliding window |
| Файлы | 10 загрузок | 1 минута | Per user |
| Сообщения | 60 в минуту | 1 минута | Per user (Phase 3) |

### 5.6. CORS

- **Разрешённый origin:** `http://localhost:5173` (Vite dev server)
- **Разрешённые методы:** GET, POST, PUT, DELETE, OPTIONS
- **Разрешённые заголовки:** Authorization, Content-Type

### 5.7. Fiber-специфика

- **Immutable: true** в dev-режиме -- все строки из контекста копируются автоматически (Pitfall 1: RequestCtx reuse).
- **CopyString:** в production обязательное использование `utils.CopyString()` для всех строк из Fiber context перед передачей в горутины.

---

## 6. Инфраструктура

### 6.1. Docker Compose

Файл: `deployments/docker-compose.yml`

| Сервис | Образ | Порты | Volumes | Healthcheck |
|--------|-------|-------|---------|-------------|
| postgres | postgres:17 | 5432 | pgdata:/var/lib/postgresql/data | pg_isready -U hub |
| mongo | mongo:8 | 27017 | mongodata:/data/db | - |
| redis | redis:7.4-alpine | 6379 | - | - |
| nats | nats:2.11 | 4222, 8222 | natsdata:/data | - |
| seaweedfs | chrislusf/seaweedfs:latest | 9333, 8333 | seaweeddata:/data | - |
| caddy | caddy:2-alpine | 80, 443 | Caddyfile монтируется | - |

**Переменные окружения:**
- PostgreSQL: `POSTGRES_USER=hub`, `POSTGRES_PASSWORD=hub`, `POSTGRES_DB=hub`
- MongoDB: `MONGO_INITDB_ROOT_USERNAME=hub`, `MONGO_INITDB_ROOT_PASSWORD=hub`
- NATS: JetStream включён флагом `-js`, данные в `/data`
- SeaweedFS: S3 включён флагом `-s3`

### 6.2. NATS

**Subjects:**
- `project.{projectID}.{event}` -- события проектов (создание, обновление, удаление, участники)
- `auth.session.{userID}.{event}` -- события сессий (new_login)

**JetStream:** включён для durable consumers (используется в будущих фазах для гарантированной доставки).

### 6.3. WebSocket Gateway

Реализация в `backend/internal/ws/`:

- **Hub** (`hub.go`): управляет клиентами и комнатами через `sync.Map`. Register/Unregister каналы (буфер 64). BroadcastToRoom и SendToUser для доставки.
- **Bridge** (`bridge.go`): подписывается на NATS subjects и маршрутизирует в WS-комнаты. Subject `project.>` -> room `project:{id}`, subject `auth.session.>` -> прямая доставка пользователю.
- **Protocol** (`protocol.go`): JSON-сообщения с типами `subscribe`, `unsubscribe`, `event`, `ping`, `pong`.
- **Handler** (`handler.go`): WebSocket upgrade через `fasthttp/websocket`, аутентификация через JWT в query parameter.

**Характеристики:**
- Буфер отправки: 256 сообщений на клиента
- При переполнении буфера сообщение дропается (не блокирует)
- Одно WS-соединение на пользователя
- Подписка на несколько комнат одновременно

### 6.4. CI Pipeline

**backend.yml (GitHub Actions):**
1. `golangci-lint run` -- статический анализ
2. `go test -race -count=1 ./...` -- тесты с race detector
3. `go build ./cmd/server` -- сборка

**frontend.yml (GitHub Actions):**
1. `npx biome check` -- линтинг
2. `npx vitest run` -- юнит-тесты
3. `npx vite build` -- сборка

### 6.5. Outbox Pattern

Обеспечение консистентности при записи в PostgreSQL и публикации в NATS:

1. Бизнес-операция и создание OutboxEvent выполняются в одной PG-транзакции.
2. Worker каждые 100ms поллит pending-события из таблицы `OutboxEvent`.
3. Для каждого события: NATS publish по `event_type`, при успехе -- статус `published`.
4. При ошибке: инкремент `attempts`, запись `last_error`. После 10 попыток -- статус `failed`.
5. Индекс `(status, created_at)` для эффективного поллинга.

### 6.6. Логирование

- **Библиотека:** Go stdlib `log/slog`
- **Формат:** структурированный JSON в production, текстовый в development
- **Уровни:** Debug, Info, Warn, Error
- **Мониторинг (будущее):** Loki + Prometheus + Grafana
