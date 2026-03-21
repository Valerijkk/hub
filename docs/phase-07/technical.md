# Фаза 7: Доработка и расширенные возможности -- Техническая документация

## Архитектура фазы 7

Фаза 7 добавляет 7 бэкенд-модулей и расширяет фронтенд:

```
+------------------+  +------------------+  +------------------+  +------------------+
|     sprint       |  |     github       |  |  calls (ext.)    |  |   apigateway     |
|                  |  |                  |  |                  |  |                  |
| service.go       |  | service.go       |  | recording.go     |  | apikey.go        |
| handler.go       |  | handler.go       |  | scheduled.go     |  | webhook.go       |
| repository.go    |  | repository.go    |  | handler.go (ext) |  | handler.go       |
| poker.go         |  | preview.go       |  | repository (ext) |  | dto.go           |
| dto.go           |  | converters.go    |  | dto.go (ext)     |  | apikey_test.go   |
| service_test.go  |  | dto.go           |  | recording_test.go|  |                  |
+------------------+  | service_test.go  |  +------------------+  +------------------+
                      +------------------+
+------------------+  +------------------+  +------------------+
|     backup       |  |      imex        |  |   search (ext.)  |
|                  |  |                  |  |                  |
| service.go       |  | service.go       |  | jql.go           |
| handler.go       |  | handler.go       |  | jql_test.go      |
| dto.go           |  | dto.go           |  |                  |
| service_test.go  |  |                  |  |                  |
+------------------+  +------------------+  +------------------+

+------------------+
|     cli (hub-cli)|
|                  |
| main.go          |
| cmd/root.go      |
| cmd/push_results |
| cmd/create_bug   |
| cmd/update_task  |
+------------------+
```

Фронтенд-расширения:
- i18next (7 namespaces x 2 языка)
- next-themes (dark/light/system + 5 accent colors)
- VitePWA (service worker + manifest)
- mobile-nav (bottom navigation)
- 6 UI-страниц (API keys, webhooks, backups, import/export, environments, templates)

---

## Ent-схемы

### Sprint (`backend/ent/schema/sprint.go`)

| Поле | Тип | Описание |
|------|-----|----------|
| id | uuid.UUID | PK, автогенерация |
| name | string | Название спринта, max 200, not empty |
| goal | string | Цель спринта, max 1000, optional |
| status | enum(planning, active, completed) | Статус жизненного цикла, default: planning |
| start_date | time.Time | Дата начала |
| end_date | time.Time | Дата окончания |
| retro_notes | JSON(map[string]interface{}) | JSONB: went_well, to_improve, actions |
| created_at | time.Time | Автоматически при создании |
| updated_at | time.Time | Автоматически при обновлении |

**Связи:** project (M2O -> Project), issues (O2M -> Issue)
**Индексы:** (project_id, status), (project_id, start_date)

---

### GitHubIntegration (`backend/ent/schema/githubintegration.go`)

| Поле | Тип | Описание |
|------|-----|----------|
| id | uuid.UUID | PK |
| repo_owner | string | Владелец репозитория (организация/пользователь) |
| repo_name | string | Имя репозитория |
| access_token | string | GitHub PAT (Sensitive) |
| webhook_secret | string | Секрет для HMAC-SHA256 валидации |
| merge_status_id | uuid.UUID | Целевой статус при мерже PR (optional, nillable) |
| created_at | time.Time | Автоматически |

**Связи:** project (M2O -> Project), github_links (O2M -> GitHubLink)
**Индексы:** UNIQUE(repo_owner, repo_name, project_id)

---

### GitHubLink (`backend/ent/schema/githublink.go`)

| Поле | Тип | Описание |
|------|-----|----------|
| id | uuid.UUID | PK |
| link_type | enum(commit, pull_request) | Тип связи |
| external_id | string | GitHub PR number или commit SHA |
| title | string | Заголовок PR / сообщение коммита, max 500 |
| url | string | Полный URL на GitHub |
| status | enum(open, merged, closed, none) | Статус PR (none для коммитов) |
| author | string | Автор на GitHub (optional) |
| branch | string | Имя ветки для PR (optional) |
| created_at, updated_at | time.Time | Timestamps |

**Связи:** issue (M2O -> Issue), integration (M2O -> GitHubIntegration)
**Индексы:** (issue_id), (external_id, link_type)

---

### Recording (`backend/ent/schema/recording.go`)

| Поле | Тип | Описание |
|------|-----|----------|
| id | uuid.UUID | PK |
| egress_id | string | ID egress-процесса LiveKit |
| status | enum(recording, processing, ready, failed) | Статус жизненного цикла |
| file_path | string | S3 ключ: recordings/{projectID}/{callID}.mp4 |
| file_size | int64 | Размер файла в байтах |
| duration_seconds | int | Длительность записи в секундах |
| started_at | time.Time | Время начала записи |
| ended_at | time.Time | Время окончания (optional, nillable) |
| call_id | uuid.UUID | FK на Call |
| project_id | uuid.UUID | FK на Project |
| started_by | uuid.UUID | FK на User |
| created_at | time.Time | Автоматически |

**Связи:** call (M2O -> Call), project (M2O -> Project), user (M2O -> User)
**Индексы:** (project_id, created_at), (call_id), (egress_id)

---

### ScheduledCall (`backend/ent/schema/scheduledcall.go`)

| Поле | Тип | Описание |
|------|-----|----------|
| id | uuid.UUID | PK |
| title | string | Название, max 255 |
| description | string | Описание, max 2000 (optional) |
| scheduled_at | time.Time | Следующее запланированное время |
| rrule | string | RFC 5545 RRULE (optional) |
| channel_id | uuid.UUID | FK на Channel |
| project_id | uuid.UUID | FK на Project |
| created_by | uuid.UUID | FK на User (создатель) |
| is_active | bool | Активен ли, default: true |
| created_at, updated_at | time.Time | Timestamps |

**Связи:** project (M2O), channel (M2O), creator (M2O -> User)
**Индексы:** (channel_id, scheduled_at), (is_active, scheduled_at)

---

### APIKey (`backend/ent/schema/apikey.go`)

| Поле | Тип | Описание |
|------|-----|----------|
| id | uuid.UUID | PK |
| name | string | Человекочитаемое имя, max 255 |
| key_hash | string | SHA-256 хеш ключа, UNIQUE |
| prefix | string | Первые 8 символов ключа |
| last_used_at | time.Time | Время последнего использования (optional, nillable) |
| expires_at | time.Time | Срок действия (optional, nillable) |
| created_at | time.Time | Автоматически |

**Связи:** user (M2O -> User)
**Индексы:** (key_hash), (expires_at)

---

### WebhookEndpoint (`backend/ent/schema/webhookendpoint.go`)

| Поле | Тип | Описание |
|------|-----|----------|
| id | uuid.UUID | PK |
| url | string | URL для POST-запросов |
| secret | string | Секрет для HMAC-SHA256 подписи |
| events | JSON([]string) | Массив подписанных событий |
| is_active | bool | Активен ли, default: true |
| created_at | time.Time | Автоматически |

**Связи:** project (M2O -> Project), deliveries (O2M -> WebhookDelivery)
**Индексы:** (is_active)

---

### WebhookDelivery (`backend/ent/schema/webhookdelivery.go`)

| Поле | Тип | Описание |
|------|-----|----------|
| id | uuid.UUID | PK |
| event_type | string | Тип события |
| payload | JSON(map[string]interface{}) | Payload события |
| status_code | int | HTTP-статус ответа |
| response_body | string | Тело ответа (max 4096) |
| error | string | Сообщение об ошибке (optional, nillable) |
| delivered_at | time.Time | Время доставки |

**Связи:** endpoint (M2O -> WebhookEndpoint)

---

### BackupRecord (`backend/ent/schema/backuprecord.go`)

| Поле | Тип | Описание |
|------|-----|----------|
| id | uuid.UUID | PK |
| s3_key | string | Ключ в S3: backups/{timestamp}.zip |
| file_size | int64 | Размер архива в байтах |
| status | enum(in_progress, completed, failed) | Статус |
| error_message | text | Сообщение об ошибке (optional, nillable) |
| created_at | time.Time | Автоматически |

**Связи:** нет (системная сущность)
**Индексы:** (status), (created_at)

---

### Environment (`backend/ent/schema/environment.go`)

| Поле | Тип | Описание |
|------|-----|----------|
| id | uuid.UUID | PK |
| name | string | Название, max 100 |
| env_label | enum(dev, staging, prod, custom) | Метка окружения |
| config | JSON(map[string]interface{}) | Конфигурация: browser, os, version |
| created_at | time.Time | Автоматически |

**Связи:** project (M2O -> Project), test_runs (O2M -> TestRun)
**Индексы:** (env_label)

---

### ProjectTemplate (`backend/ent/schema/projecttemplate.go`)

| Поле | Тип | Описание |
|------|-----|----------|
| id | uuid.UUID | PK |
| name | string | Название шаблона, max 255 |
| description | text | Описание (optional) |
| template_type | enum(empty, software_dev, qa, kanban_simple, custom) | Тип шаблона |
| config | JSON(map[string]interface{}) | Снимок: statuses, labels, boards, custom_fields |
| is_system | bool | Системный (нельзя удалить), default: false |
| created_at | time.Time | Автоматически |

**Связи:** created_by (M2O -> User, optional)
**Индексы:** (template_type), (is_system)

---

## API Reference

### Sprint модуль

| Метод | Путь | Описание | Auth | RBAC |
|-------|------|----------|------|------|
| POST | `/api/projects/:projectId/sprints` | Создать спринт | JWT | TaskCreate |
| GET | `/api/projects/:projectId/sprints` | Список спринтов | JWT | ProjectView |
| GET | `/api/projects/:projectId/sprints/:sprintId` | Получить спринт | JWT | ProjectView |
| PUT | `/api/projects/:projectId/sprints/:sprintId` | Обновить спринт | JWT | TaskEdit |
| POST | `/api/projects/:projectId/sprints/:sprintId/start` | Запустить спринт | JWT | TaskEdit |
| POST | `/api/projects/:projectId/sprints/:sprintId/complete` | Завершить спринт | JWT | TaskEdit |
| PUT | `/api/projects/:projectId/sprints/:sprintId/retro` | Обновить ретро-заметки | JWT | TaskEdit |
| GET | `/api/projects/:projectId/sprints/:sprintId/burndown` | Burndown chart данные | JWT | ProjectView |
| POST | `/api/projects/:projectId/sprints/:sprintId/issues` | Добавить задачи в спринт | JWT | TaskEdit |
| GET | `/api/projects/:projectId/velocity` | Velocity chart данные | JWT | ProjectView |
| POST | `/api/projects/:projectId/sprints/:sprintId/poker` | Запустить Poker-сессию | JWT | ProjectView |
| POST | `/api/sprints/poker/:sessionId/vote` | Проголосовать | JWT | - |
| POST | `/api/sprints/poker/:sessionId/reveal` | Раскрыть голоса | JWT | - |
| GET | `/api/sprints/poker/:sessionId` | Получить сессию | JWT | - |

**Примеры запросов:**

Создание спринта:
```json
POST /api/projects/{projectId}/sprints
{
  "name": "Sprint 5",
  "goal": "Release MVP",
  "start_date": "2026-03-18T00:00:00Z",
  "end_date": "2026-03-29T00:00:00Z"
}
```

Ответ:
```json
{
  "id": "uuid",
  "name": "Sprint 5",
  "goal": "Release MVP",
  "status": "planning",
  "start_date": "2026-03-18T00:00:00Z",
  "end_date": "2026-03-29T00:00:00Z"
}
```

Голосование в Poker:
```json
POST /api/sprints/poker/{sessionId}/vote
{
  "value": 8
}
```

---

### GitHub модуль

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| POST | `/api/github/webhook` | GitHub webhook (HMAC) | Подпись |
| POST | `/api/projects/:projectId/github/integrations` | Создать интеграцию | JWT+Owner |
| GET | `/api/projects/:projectId/github/integrations` | Список интеграций | JWT |
| DELETE | `/api/projects/:projectId/github/integrations/:id` | Удалить интеграцию | JWT+Owner |
| GET | `/api/issues/:issueKey/github-links` | GitHub ссылки задачи | JWT |
| POST | `/api/issues/:issueKey/github/create-branch` | Создать ветку | JWT |
| GET | `/api/link-preview?url=...` | Превью ссылки | JWT |

**Примеры запросов:**

Создание интеграции:
```json
POST /api/projects/{projectId}/github/integrations
{
  "repo_owner": "myorg",
  "repo_name": "myrepo",
  "access_token": "ghp_...",
  "webhook_secret": "secret123",
  "merge_status_id": "uuid-of-done-status"
}
```

Создание ветки:
```json
POST /api/issues/HUB-42/github/create-branch
{
  "integration_id": "uuid"
}
```

Ответ:
```json
{
  "branch_name": "hub-42-add-user-search",
  "url": "https://github.com/myorg/myrepo/tree/hub-42-add-user-search"
}
```

---

### Calls расширения (Recording + Scheduled)

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| POST | `/api/calls/:callId/recording/start` | Начать запись | JWT |
| POST | `/api/calls/:callId/recording/stop` | Остановить запись | JWT |
| GET | `/api/projects/:projectId/recordings` | Записи проекта | JWT |
| GET | `/api/recordings/:recordingId/download` | URL для скачивания | JWT |
| POST | `/api/projects/:projectId/scheduled-calls` | Создать запланированный звонок | JWT |
| GET | `/api/channels/:channelId/scheduled-calls` | Список запланированных | JWT |
| PUT | `/api/scheduled-calls/:id` | Обновить | JWT |
| DELETE | `/api/scheduled-calls/:id` | Удалить | JWT |

**Пример: Создание повторяющегося звонка:**
```json
POST /api/projects/{projectId}/scheduled-calls
{
  "title": "Daily Standup",
  "description": "Ежедневный стендап команды",
  "channel_id": "uuid",
  "scheduled_at": "2026-03-22T10:00:00Z",
  "rrule": "FREQ=DAILY;BYDAY=MO,TU,WE,TH,FR"
}
```

---

### API Gateway модуль

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| POST | `/api/api-keys` | Создать API-ключ | JWT |
| GET | `/api/api-keys` | Список ключей | JWT |
| DELETE | `/api/api-keys/:id` | Отозвать ключ | JWT |
| POST | `/api/projects/:projectId/webhooks` | Создать вебхук | JWT+Owner |
| GET | `/api/projects/:projectId/webhooks` | Список вебхуков | JWT |
| PUT | `/api/webhooks/:id` | Обновить вебхук | JWT |
| DELETE | `/api/webhooks/:id` | Удалить вебхук | JWT |
| GET | `/api/webhooks/:id/deliveries` | История доставок | JWT |

**Пример: Создание API-ключа:**
```json
POST /api/api-keys
{
  "name": "CI/CD Key",
  "expires_at": "2027-01-01T00:00:00Z"
}
```

Ответ (ключ показывается ТОЛЬКО при создании):
```json
{
  "id": "uuid",
  "name": "CI/CD Key",
  "prefix": "Ab3kT2xq",
  "key": "Ab3kT2xqLm9pR4sW...",
  "expires_at": "2027-01-01T00:00:00Z",
  "created_at": "2026-03-21T10:00:00Z"
}
```

**Пример: Создание вебхука:**
```json
POST /api/projects/{projectId}/webhooks
{
  "url": "https://hooks.slack.com/...",
  "events": ["task.created", "task.status_changed", "run.completed"],
  "secret": "my-webhook-secret"
}
```

---

### Backup модуль

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| POST | `/api/admin/backups/create` | Создать бэкап | JWT+Admin |
| GET | `/api/admin/backups` | Список бэкапов | JWT+Admin |
| POST | `/api/admin/backups/:id/restore` | Восстановить | JWT+Admin |
| DELETE | `/api/admin/backups/:id` | Удалить бэкап | JWT+Admin |

---

## Интеграции

### LiveKit Egress API

Используется для записи звонков. `RecordingService` создает `lksdk.EgressClient` и вызывает:

- `StartRoomCompositeEgress(RoomCompositeEgressRequest)` -- начало записи с S3 output
- `StopEgress(StopEgressRequest)` -- остановка записи

Конфигурация Egress request:
```go
&livekit.RoomCompositeEgressRequest{
    RoomName:  callEnt.RoomName,
    AudioOnly: true,
    Output: &livekit.RoomCompositeEgressRequest_File{
        File: &livekit.EncodedFileOutput{
            FileType: livekit.EncodedFileType_MP4,
            Filepath: "recordings/{projectID}/{callID}.mp4",
            Output: &livekit.EncodedFileOutput_S3{
                S3: &livekit.S3Upload{
                    Bucket:         cfg.Bucket,
                    Endpoint:       cfg.Endpoint,
                    ForcePathStyle: true,
                },
            },
        },
    },
}
```

Жизненный цикл записи: `recording` -> (StopEgress) -> `processing` -> (OnEgressEnded webhook) -> `ready`/`failed`.

---

### GitHub Webhook

POST /api/github/webhook принимает события:

- **push** -- извлекает ключи задач из commit messages, создает GitHubLink type=commit
- **pull_request** -- извлекает ключи из title/body, создает GitHubLink type=pull_request, обновляет статус

Валидация подписи: `X-Hub-Signature-256: sha256=<HMAC-SHA256(body, webhook_secret)>`

Regex для ключей задач: `\b([A-Z]+-\d+)\b` (word boundaries для точности).

---

### go-github API

`google/go-github/v68` используется для:

- Создание ветки: `client.Git.CreateRef()` на основе HEAD ref
- Получение информации о репозитории: `client.Repositories.Get()`

Slugify для имени ветки: `hub-{number}-{slugified-title}`, max 50 символов, спецсимволы заменяются дефисами, множественные дефисы схлопываются.

---

### otiai10/opengraph

`otiai10/opengraph/v2` извлекает Open Graph метаданные из внешних URL:

```go
og, err := opengraph.Fetch(rawURL)
// og.Title, og.Description, og.Image[0].URL
```

Результаты кешируются в Redis с ключом `link_preview:{url}` и TTL 1 час.

---

## Frontend: i18n архитектура

### Namespace-per-module

```
frontend/src/i18n/
  index.ts              -- инициализация i18next
  locales/
    ru/
      common.json       -- общие строки (buttons, labels, errors)
      tracker.json      -- трекер задач
      messenger.json    -- мессенджер
      tms.json          -- TMS
      calls.json        -- звонки
      dashboard.json    -- дашборды
      settings.json     -- настройки
    en/
      common.json
      tracker.json
      ...
```

Использование в компонентах:
```tsx
const { t } = useTranslation('tracker');
return <h1>{t('sprint.title')}</h1>;
```

Определение языка: `i18next-browser-languagedetector` -> `localStorage` -> `navigator.language` -> fallback "ru".

---

## Frontend: Theme System

### CSS Variables (oklch)

```css
/* app.css */
[data-accent="purple"] {
  --accent-50:  oklch(0.97 0.02 292);
  --accent-500: oklch(0.55 0.25 292);
  --accent-900: oklch(0.25 0.12 292);
}

.dark [data-accent="purple"] {
  --accent-50:  oklch(0.25 0.08 292);
  --accent-500: oklch(0.65 0.25 292);
  --accent-900: oklch(0.90 0.08 292);
}
```

Механизм: `next-themes` управляет `class="dark"` на `<html>`. `useAccentColor()` хук устанавливает `data-accent` атрибут.

5 акцентных цветов: red, blue (default), green, purple, orange.

---

## Frontend: PWA конфигурация

```ts
// vite.config.ts
VitePWA({
  registerType: 'autoUpdate',
  manifest: false, // внешний manifest.webmanifest
  workbox: {
    maximumFileSizeToCacheInBytes: 4 * 1024 * 1024, // 4MB
    globPatterns: ['**/*.{js,css,html,ico,png,svg,woff2}'],
  },
})
```

manifest.webmanifest: name="Hub", display="standalone", theme_color="#1e293b", иконки 192x192 и 512x512.

---

## Frontend: Mobile Layout

```tsx
// useIsMobile hook
const useIsMobile = () => {
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);
  // resize listener...
  return isMobile;
};

// MobileNav -- bottom navigation (md:hidden)
// 5 tabs: Dashboard, Tracker, Messenger, Calls, More
```

Sidebar на мобильном: slide-over overlay с backdrop, закрывается по клику вне области.

---

## JQL синтаксис

### Поддерживаемые поля

| Поле | Тип | Описание |
|------|-----|----------|
| status | string | Статус задачи |
| priority | enum | critical, high, medium, low |
| type | enum | bug, task, story, epic, sub_task, improvement, tech_debt, spike, release |
| assignee | uuid/function | ID пользователя или currentUser() |
| created | date | Дата создания |
| updated | date | Дата обновления |

### Операторы

| Оператор | Описание | Пример |
|----------|----------|--------|
| = | Равно | `priority = high` |
| != | Не равно | `status != "Done"` |
| IN | В списке | `type IN (bug, task)` |
| NOT IN | Не в списке | `priority NOT IN (low)` |
| > | Больше | `created > "2026-03-01"` |
| < | Меньше | `updated < "2026-03-20"` |
| >= | Больше или равно | `created >= "2026-01-01"` |
| <= | Меньше или равно | `updated <= "2026-12-31"` |

### Логические операторы

- **AND** -- оба условия должны быть истинны
- **OR** -- хотя бы одно условие истинно

### ORDER BY

`ORDER BY field ASC|DESC` -- сортировка результатов.

### Примеры запросов

```
priority = high
```

```
status = "In Progress" AND assignee = currentUser()
```

```
type IN (bug, task) AND priority != low ORDER BY created DESC
```

```
assignee = currentUser() AND status NOT IN ("Done", "Closed") ORDER BY priority ASC
```

### Парсер

Recursive descent parser (`jqlParser`):
1. Считывает поле (identifier)
2. Считывает оператор (=, !=, IN, NOT IN, >, <, >=, <=)
3. Считывает значение (строка в кавычках, идентификатор, или список в скобках для IN)
4. Проверяет AND/OR для следующего условия
5. Проверяет ORDER BY в конце

---

## Backup/Restore

### Процедура создания бэкапа

```
1. BackupRecord(status=in_progress) -> PostgreSQL
2. pg_dump --format=custom $PG_URL -> pgData []byte
3. mongodump --uri=$MONGO_URL --archive -> mongoData []byte
4. ZIP(postgres.dump, mongo.archive) -> zipData []byte
5. S3.PutObject(backups/{timestamp}.zip, zipData)
6. BackupRecord(status=completed, file_size=len(zipData))
7. cleanupOldBackups() -- удаление если > maxBackups(4)
```

### Процедура восстановления

```
1. Проверка: бэкап status == completed
2. Автоматический бэкап текущего состояния (safety net)
3. S3.GetObject(backup.S3Key) -> ZIP
4. Распаковка: postgres.dump, mongo.archive
5. pg_restore + mongorestore
```

### Ограничения

- pg_dump и mongodump должны быть установлены на сервере
- Восстановление перезаписывает все данные (полная замена)
- Максимум 4 бэкапа по умолчанию (настраивается через maxBackups)
- Еженедельный автоматический бэкап через ticker-горутину в main.go

---

## hub-cli

### Команды

```
hub-cli [global flags] <command> [command flags]

Global flags:
  --api-url    URL сервера Hub API (default: $HUB_API_URL или http://localhost:3000)
  --api-key    API-ключ для авторизации (default: $HUB_API_KEY)
  --json       Вывод в JSON формате
```

### push-results

Загрузка результатов тестов (Allure) в Hub TMS:

```bash
hub-cli push-results --project HUB --file ./allure-results
hub-cli push-results --project HUB --file ./allure-results --run-id <existing-run-id>
```

Процесс: считывает все JSON-файлы из директории, отправляет POST на `/api/projects/{project}/tms/allure/import`.

### create-bug

Создание баг-репорта:

```bash
hub-cli create-bug --project HUB --title "Login button broken" --priority high
hub-cli create-bug --project HUB --title "API timeout" --description "Details..." --priority critical
```

### update-task

Обновление задачи:

```bash
hub-cli update-task --key HUB-42 --status done
hub-cli update-task --key HUB-42 --assignee user@example.com
```

### Аутентификация

Все команды используют заголовок `X-API-Key` для авторизации. Ключ передается через:
1. Флаг `--api-key`
2. Переменная окружения `HUB_API_KEY`

Ошибка при отсутствии ключа: "API-ключ не задан (--api-key или HUB_API_KEY)".

### Пример CI/CD (GitHub Actions)

```yaml
- name: Push test results
  run: |
    hub-cli push-results \
      --project ${{ env.HUB_PROJECT }} \
      --file ./allure-results
  env:
    HUB_API_URL: https://hub.example.com
    HUB_API_KEY: ${{ secrets.HUB_API_KEY }}
```

---

## Модифицированные схемы (дополнения фазы 7)

### Issue (расширение)

- Добавлен edge `sprint` (M2O -> Sprint, optional) -- привязка задачи к спринту
- Добавлен edge `github_links` (O2M -> GitHubLink) -- GitHub связи

### WorkflowStatus (расширение)

- Добавлено поле `wip_limit` (int, default 0) -- WIP-лимит для Kanban

### Board (расширение)

- Добавлено поле `swimlane_mode` (enum: none, assignee, priority, type) -- режим swimlanes

### Project (расширение)

- Добавлены edges: `sprints`, `github_integrations`, `recordings`, `scheduled_calls`, `webhook_endpoints`, `environments`

### User (расширение)

- Добавлены edges: `api_keys`, `recordings`, `scheduled_calls`, `project_templates`

### Channel (расширение)

- Добавлен edge: `scheduled_calls`

### Call (расширение)

- Добавлен edge: `recordings`

### TestRun (расширение)

- Добавлен edge: `environment` (M2O -> Environment, optional)

---

## Паттерны и решения

### WIPChecker interface

```go
type WIPChecker interface {
    CheckWIPLimit(ctx context.Context, statusID uuid.UUID) error
}
```

Инжектируется в TrackerService через `SetWIPChecker()`. Позволяет sprint-модулю контролировать перемещение задач без жесткой связи с трекером.

### API Key Auth параллельно с JWT

```go
func APIKeyAuth(svc *APIKeyService) fiber.Handler {
    return func(c fiber.Ctx) error {
        key := c.Get("X-API-Key")
        if key == "" {
            return c.Next() // Fallback на JWT
        }
        userID, err := svc.VerifyAPIKey(c.Context(), key)
        // ...
        c.Locals("userID", userID.String())
        return c.Next()
    }
}
```

Middleware проверяет заголовок `X-API-Key`. Если присутствует -- верифицирует ключ. Если отсутствует -- пропускает к JWT middleware. Один и тот же `c.Locals("userID")` для обоих методов.

### Webhook HMAC-SHA256

```go
func signPayload(payload []byte, secret string) string {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    return hex.EncodeToString(mac.Sum(nil))
}
```

Заголовок: `X-Hub-Signature-256: sha256=<hex>`. Получатель верифицирует подпись тем же алгоритмом.

### Nil-safe NATS publishing

Паттерн, повторяющийся во всех модулях фазы 7:

```go
func (s *Service) publishEvent(subject string, payload interface{}) {
    if s.nats == nil {
        slog.Debug("NATS не подключен, пропускаем", "subject", subject)
        return
    }
    data, _ := json.Marshal(payload)
    s.nats.Publish(subject, data)
}
```

Позволяет тестировать сервисы без mock NATS-клиента.

---

*Фаза: 07-polish-extended*
*Последнее обновление: 2026-03-21*
