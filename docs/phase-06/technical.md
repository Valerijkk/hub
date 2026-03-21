# Фаза 6: Уведомления, поиск, дашборды и аудит -- Техническая документация

## Архитектура фазы 6

Фаза 6 добавляет 4 модуля, объединяющих все предыдущие подсистемы:

```
+-------------------+    +-------------------+    +-------------------+    +-------------------+
|   notification    |    |      search       |    |    dashboard      |    |      audit        |
|                   |    |                   |    |                   |    |                   |
| consumer.go       |    | service.go        |    | service.go        |    | consumer.go       |
| service.go        |    | handler.go        |    | handler.go        |    | service.go        |
| handler.go        |    | indexer.go        |    | dto.go            |    | handler.go        |
| telegram.go       |    | dto.go            |    | service_test.go   |    | repository.go     |
| repository.go     |    | service_test.go   |    |                   |    | dto.go            |
| events.go         |    |                   |    |                   |    | service_test.go   |
| dto.go            |    |                   |    |                   |    |                   |
| service_test.go   |    |                   |    |                   |    |                   |
+-------------------+    +-------------------+    +-------------------+    +-------------------+
        |                        |                        |                        |
        v                        v                        v                        v
+-----------------------------------------------------------------------------+
|                          NATS JetStream                                       |
|  Streams: NOTIFICATIONS, AUDIT                                               |
|  Consumers: notification-dispatcher, audit-logger                            |
+-----------------------------------------------------------------------------+
        |                        |                        |
        v                        v                        v
+-------------------+    +-------------------+    +-------------------+
|   PostgreSQL      |    |     MongoDB       |    |      Redis        |
|   (Ent ORM)       |    |  messages (text)  |    |  dashboard cache  |
|   + Raw SQL       |    |                   |    |  (5min TTL)       |
+-------------------+    +-------------------+    +-------------------+
```

---

## Уведомления (notification)

### Архитектура

Модуль уведомлений построен на паттерне JetStream durable consumer. События из всех модулей платформы (трекер, мессенджер, TMS, звонки) публикуются в стрим NOTIFICATIONS. Consumer `notification-dispatcher` обрабатывает каждое событие, определяет получателей, проверяет пользовательские настройки и доставляет через два канала: WebSocket (In-App) и Telegram.

```
Модули платформы              JetStream                 Consumer                  Каналы доставки
+------------------+     +-------------------+     +---------------------+     +------------------+
| tracker          |     | Stream:           |     | notification-       |     | NATS pub          |
|  project.>       |---->| NOTIFICATIONS     |---->| dispatcher          |---->| notification.    |
|                  |     |                   |     |                     |     |   user.{userID}  |
| messenger        |     | WorkQueue         |     | 1. parseSubject()   |     | -> WS bridge     |
|  messenger.      |---->| MaxAge: 7 days    |     | 2. isDuplicate()    |     | -> browser       |
|    channel.>     |     | AckExplicit       |     | 3. checkPrefs()     |     +------------------+
|                  |     | MaxDeliver: 5     |     | 4. checkQuietHrs()  |
| tms              |     | AckWait: 30s      |     | 5. saveNotification |     +------------------+
|  tms.>           |---->|                   |     | 6. publishWS()      |---->| TelegramSender   |
|                  |     |                   |     | 7. sendTelegram()   |     | telego bot       |
| calls            |     |                   |     +---------------------+     | rate: 35ms       |
|  calls.channel.> |---->|                   |                                 | buffer: 100      |
+------------------+     +-------------------+                                 +------------------+
```

### Ent-схемы

**Notification** (`backend/ent/schema/notification.go`):

| Поле | Тип | Описание |
|------|-----|----------|
| id | uuid.UUID | Первичный ключ |
| type | string | Тип уведомления (assigned, mentioned, status_changed и т.д.) |
| title | string | Заголовок уведомления |
| body | string | Тело уведомления (опционально) |
| entity_type | string | Тип сущности (issue, test_case, call и т.д.) |
| entity_id | string | ID сущности |
| project_id | uuid.UUID | FK на проект (опционально) |
| actor_id | uuid.UUID | FK на автора действия (опционально) |
| is_read | bool | Прочитано ли (default: false) |
| is_urgent | bool | Срочное ли (пробивает quiet hours) |
| created_at | time.Time | Время создания (автоматически) |

**Индексы:**
- `(user_id, is_read, created_at)` -- для списка непрочитанных
- `(user_id, created_at)` -- для пагинированного списка

**Edges:** User (M2O) -- обратная ссылка на получателя

---

**NotificationPreference** (`backend/ent/schema/notificationpreference.go`):

| Поле | Тип | Описание |
|------|-----|----------|
| id | uuid.UUID | Первичный ключ |
| notification_type | string | Тип уведомления |
| channel | string | Канал доставки (in_app, telegram) |
| enabled | bool | Включено ли |
| project_id | uuid.UUID | FK на проект (optional, для per-project настроек) |

**Unique constraint:** `(user_id, notification_type, channel, project_id)` -- одна настройка на комбинацию

**Edges:** User (M2O)

---

**User schema (расширение):**

| Поле | Тип | Описание |
|------|-----|----------|
| quiet_hours_start | string | Начало тихих часов, формат HH:MM (nullable) |
| quiet_hours_end | string | Конец тихих часов, формат HH:MM (nullable) |
| status | enum(active, deactivated) | Статус пользователя (добавлен в Phase 06-04) |

**Новые edges:** notifications (O2M -> Notification), notification_preferences (O2M -> NotificationPreference), dashboard_layouts (O2M -> DashboardLayout)

### REST API

#### GET /api/notifications

Возвращает пагинированный список уведомлений пользователя.

**Query params:**
- `page` (int, default: 1)
- `per_page` (int, default: 20, max: 100)
- `is_read` (bool, optional) -- фильтр по прочитанности

**Response:**
```json
{
  "notifications": [
    {
      "id": "uuid",
      "type": "assigned",
      "title": "Вам назначена задача HUB-42",
      "body": "Задача назначена пользователем Иванов",
      "entity_type": "issue",
      "entity_id": "uuid",
      "project_id": "uuid",
      "actor_id": "uuid",
      "is_read": false,
      "is_urgent": false,
      "created_at": "2026-03-21T10:00:00Z"
    }
  ],
  "total": 42,
  "page": 1,
  "per_page": 20
}
```

#### GET /api/notifications/unread-count

```json
{ "count": 5 }
```

#### PATCH /api/notifications/:id/read

Помечает уведомление прочитанным. Ответ: 204 No Content.

#### PATCH /api/notifications/read-all

Помечает все уведомления пользователя прочитанными. Ответ: 204 No Content.

#### GET /api/notifications/preferences

```json
[
  {
    "id": "uuid",
    "notification_type": "assigned",
    "channel": "telegram",
    "enabled": true,
    "project_id": null
  }
]
```

#### PUT /api/notifications/preferences

**Request:**
```json
{
  "notification_type": "assigned",
  "channel": "telegram",
  "enabled": false,
  "project_id": "uuid"
}
```

### WebSocket интеграция

WS bridge (`backend/internal/ws/bridge.go`) расширен подпиской на `notification.user.>`:

1. NATS субъект: `notification.user.{userID}` -- уникальный на пользователя
2. Bridge извлекает userID из субъекта и вызывает `SendToUser(userID, data)` с event `notification.new`
3. Payload: `WsNotificationPayload` (id, type, title, body, entity_type, entity_id, project_id, actor_id, is_urgent, created_at)
4. Frontend: Zustand store подписан на WS-событие `notification.new`, обновляет unreadCount и показывает sonner toast

### Telegram интеграция

**TelegramSender** (`backend/internal/notification/telegram.go`):

- Инициализация: `NewTelegramSender(botToken)` -- если botToken пустой, возвращает nil
- Nil-safe: Send() на nil sender -- no-op (метод на указателе, проверка `if s == nil { return }`)
- Буферизованный канал: `queue chan telegramMessage` с capacity=100
- Worker-горутина: извлекает из очереди, отправляет через `bot.SendMessage()` (telego)
- Rate limiting: `time.Sleep(35 * time.Millisecond)` между отправками (Telegram API limit: ~30 msg/sec)
- Форматирование: MarkdownV2 с эскейпингом спецсимволов (`_`, `*`, `[`, `]`, `(`, `)`, `~`, `` ` ``, `>`, `#`, `+`, `-`, `=`, `|`, `{`, `}`, `.`, `!`)

### Тихие часы

Алгоритм проверки (`NotificationService`):

```go
func isInQuietHours(now time.Time, startStr, endStr string) bool {
    start := parseTimeMinutes(startStr) // "22:00" -> 22*60+0 = 1320
    end := parseTimeMinutes(endStr)     // "08:00" -> 8*60+0 = 480
    nowMin := now.Hour()*60 + now.Minute()

    if start <= end {
        // Обычный диапазон: 08:00 - 18:00
        return nowMin >= start && nowMin < end
    }
    // Перенос через полночь: 22:00 - 08:00
    return nowMin >= start || nowMin < end
}
```

- Timezone: берется из профиля пользователя (поле `timezone`), конвертируется через `time.LoadLocation()`
- Если timezone не задан -- используется UTC
- Срочные (is_urgent=true) обходят проверку

### Debounce

- In-memory map: `map[string]time.Time` с ключом `userID:eventType:entityID`
- Окно: 2 секунды
- Mutex: `sync.Mutex` для потокобезопасности
- При перезагрузке сервиса debounce сбрасывается (допустимо для 2-секундного окна)

---

## Поиск (search)

### Архитектура

Глобальный поиск использует параллельные запросы к двум хранилищам: PostgreSQL (tsvector full-text search) для структурированных данных и MongoDB ($text index) для сообщений мессенджера.

```
GET /api/search?q=логин                     SearchService.Search()
        |                                           |
        v                                           v
+------------------+                    +-----------------------+
| SearchHandler    |                    | Параллельные горутины  |
| - parse q, type  |                    |                       |
| - parse limit    |------------------->| go searchIssues()     |
| - getUserID()    |                    | go searchTestCases()  |
+------------------+                    | go searchUsers()      |
                                        | go searchProjects()   |
                                        | go searchFiles()      |
                                        | go searchChannels()   |
                                        | go searchMessages()   | <- MongoDB
                                        +-----------------------+
                                                    |
                                                    v
                                        +-----------------------+
                                        | sync.Mutex append     |
                                        | sort by score DESC    |
                                        | limit results         |
                                        +-----------------------+
```

### tsvector: GENERATED ALWAYS AS STORED

Для 5 PostgreSQL-таблиц добавлены колонки `search_vector` типа `tsvector` с конфигурацией `'simple'` (language-agnostic, подходит для русского и английского):

| Таблица | Поля для индексации | SQL |
|---------|--------------------|----|
| issues | title, description_text | `search_vector tsvector GENERATED ALWAYS AS (to_tsvector('simple', coalesce(title,'') \|\| ' ' \|\| coalesce(description_text,''))) STORED` |
| test_cases | title, case_key | `search_vector tsvector GENERATED ALWAYS AS (to_tsvector('simple', coalesce(title,'') \|\| ' ' \|\| coalesce(case_key,''))) STORED` |
| users | name, email, display_name | `search_vector tsvector GENERATED ALWAYS AS (to_tsvector('simple', coalesce(name,'') \|\| ' ' \|\| coalesce(email,'') \|\| ' ' \|\| coalesce(display_name,''))) STORED` |
| projects | name, key | `search_vector tsvector GENERATED ALWAYS AS (to_tsvector('simple', coalesce(name,'') \|\| ' ' \|\| coalesce(key,''))) STORED` |
| file_metadata | filename | `search_vector tsvector GENERATED ALWAYS AS (to_tsvector('simple', coalesce(filename,''))) STORED` |

Для каждой таблицы создан GIN индекс: `CREATE INDEX idx_{table}_search ON {table} USING GIN(search_vector)`.

Индексы создаются при старте сервера через `search.EnsureSearchIndexes()` (raw SQL через `database.OpenRawDB()`).

### MongoDB text index

Коллекция `messages`:
```javascript
db.messages.createIndex({ content_text: "text" })
```

Поиск через `$text` оператор с проекцией `$meta: "textScore"` для ранжирования.

### Merge и ranking

1. Каждая горутина возвращает `[]SearchResult` с полем `Score`
2. PostgreSQL score: `ts_rank(search_vector, plainto_tsquery('simple', $1))`
3. MongoDB score: `textScore * 0.8` (коэффициент для нормализации между PG и MongoDB)
4. Каналы: фиксированный score=0.5 (ILIKE, не tsvector)
5. Результаты объединяются, сортируются по score DESC, обрезаются по limit

### Безопасность: фильтрация по проектам

Все запросы (кроме users) фильтруют результаты по project membership:

```sql
-- Получение проектов пользователя
SELECT p.id FROM project_members pm
JOIN projects p ON p.id = pm.project_members
WHERE pm.project_member_user = $1

-- Фильтрация в поисковом запросе
WHERE search_vector @@ plainto_tsquery('simple', $1)
  AND issue_project = ANY($2)  -- $2 = массив ID проектов
```

Users доступны глобально (без фильтра по проектам).

### SavedFilter

**Ent-схема** (`backend/ent/schema/savedfilter.go`):

| Поле | Тип | Описание |
|------|-----|----------|
| id | uuid.UUID | Первичный ключ |
| name | string | Имя фильтра (max 200) |
| filter_config | JSON (map[string]interface{}) | Конфигурация фильтра |
| shared | bool | Расшарен ли (default: false) |

**Edges:** owner (M2O -> User), project (M2O -> Project)

**CRUD API:**
- POST /api/filters -- создать фильтр (name, filter_config, project_id)
- GET /api/filters?project_id= -- список (свои + shared проекта)
- DELETE /api/filters/:id -- удалить (только владелец)
- PATCH /api/filters/:id/share -- переключить shared (только владелец)

---

## Дашборды (dashboard)

### Архитектура

```
Frontend                    Backend                      Хранилища
+-------------------+    +---------------------+    +-----------------+
| /dashboard        |    | DashboardService    |    | PostgreSQL      |
|  PersonalPage     |--->| GetPersonalData()   |--->| issues          |
|  WidgetGrid       |    |                     |    | time_logs       |
|  react-grid-layout|    | GetProjectStats()   |    | test_runs       |
+-------------------+    | (Redis cache 5min)  |    | sessions        |
                         |                     |    | calls           |
+-------------------+    | GetAdminStats()     |    +-----------------+
| /projects/$id/    |    |                     |
|   dashboard       |--->| GetLayout()         |    +-----------------+
|  WidgetGrid       |    | SaveLayout()        |    | Redis           |
+-------------------+    +---------------------+    | project_stats:  |
                                                    |   {projectId}   |
+-------------------+                               | TTL: 5min       |
| /admin/dashboard  |                               +-----------------+
|  KPI Cards        |
|  WidgetGrid       |
+-------------------+
```

### DashboardLayout Ent-схема

(`backend/ent/schema/dashboardlayout.go`):

| Поле | Тип | Описание |
|------|-----|----------|
| id | uuid.UUID | Первичный ключ |
| dashboard_type | string | Тип дашборда (personal, project, admin) |
| project_id | uuid.UUID | FK на проект (optional, для project dashboard) |
| layout | JSON | Массив LayoutItem (widget_type, x, y, w, h) |

**Unique index:** `(user_id, dashboard_type, project_id)` -- один layout на комбинацию

**Edges:** User (M2O)

**LayoutItem структура:**
```json
{
  "widget_type": "my_tasks",
  "x": 0, "y": 0,
  "w": 6, "h": 4
}
```

### Aggregation queries

**Personal dashboard:**
- Мои задачи: `Issue.Query().Where(issue.HasAssigneeWith(user.ID(userID)), issue.DeletedAtIsNil()).Limit(20)`
- Непрочитанные: Redis HGET unread:{userID} (из мессенджера)
- Звонки: `Call.Query().Where(call.StatusEQ("active")).WithChannel()`
- Дедлайны: `Issue.Query().Where(issue.DueDateLTE(now+7days), issue.StatusNEQ("done"))`
- Таймтрекинг: `TimeLog.Query().Where(timelog.UserID(userID), timelog.CreatedAtGTE(today))`

**Project dashboard (Redis-cached, TTL 5min):**
```go
cacheKey := fmt.Sprintf("project_stats:%s", projectID)
cached, err := s.rdb.Get(ctx, cacheKey).Result()
if err == nil {
    // cache hit
    json.Unmarshal([]byte(cached), &stats)
    return stats, nil
}
// cache miss -> aggregate from DB -> cache result
```

Данные:
- Статусы задач: GROUP BY status -> PieChart
- Burndown: задачи по дням текущего спринта
- Velocity: закрытые story points по спринтам -> BarChart
- Тест-тренды: результаты прогонов (pass/fail/skip) -> AreaChart
- Активность команды: коммиты/действия пользователей за неделю

**Admin dashboard:**
- Общие метрики: COUNT users, projects, issues
- DAU/WAU/MAU: подсчет уникальных сессий за период
- Хранилище: SUM file sizes из file_metadata
- Здоровье: проверка подключений к PG, MongoDB, Redis, NATS

### Frontend: react-grid-layout v2

**API v2 отличия от v1:**
```tsx
// v2: useContainerWidth hook (вместо WidthProvider HOC)
const { containerRef, width } = useContainerWidth();

// v2: gridConfig/dragConfig/resizeConfig objects (вместо flat props)
<ReactGridLayout
  gridConfig={{ cols: 12, rowHeight: 80, width }}
  dragConfig={{ isDraggable: true }}
  resizeConfig={{ isResizable: true }}
  layout={layout}
  onLayoutChange={handleLayoutChange}
/>
```

**Widget registry pattern:**
```tsx
const WIDGET_REGISTRY: Record<string, React.ComponentType<WidgetProps>> = {
  my_tasks: MyTasksWidget,
  unread_messages: UnreadMessagesWidget,
  upcoming_calls: UpcomingCallsWidget,
  task_status_chart: TaskStatusChartWidget,
  burndown_chart: BurndownChartWidget,
  velocity_chart: VelocityChartWidget,
  test_trends: TestTrendsWidget,
  team_activity: TeamActivityWidget,
  system_health: SystemHealthWidget,
  storage_usage: StorageUsageWidget,
  time_tracking: TimeTrackingWidget,
  deadlines: DeadlinesWidget,
};
```

### Recharts для графиков

Использованы напрямую (без Tremor обертки):
- `PieChart` -- распределение задач по статусам (TaskStatusChart)
- `AreaChart` -- burndown (BurndownChart), тренды тестов (TestTrends)
- `BarChart` -- velocity по спринтам (VelocityChart)

### Шаблоны

5 JSON-пресетов в `frontend/src/components/dashboard/dashboard-templates.ts`:

| Шаблон | Назначение | Виджеты |
|--------|-----------|---------|
| personalDefault | Персональный (по умолчанию) | my_tasks, unread_messages, upcoming_calls, deadlines, time_tracking |
| projectDev | Разработчик проекта | task_status_chart, burndown_chart, velocity_chart, team_activity |
| projectQA | Тестировщик проекта | test_trends, task_status_chart, team_activity |
| projectManager | Менеджер проекта | task_status_chart, burndown_chart, velocity_chart, test_trends, team_activity |
| adminDefault | Администратор | system_health, storage_usage |

---

## Аудит и администрирование (audit)

### AuditLog Ent-схема

(`backend/ent/schema/auditlog.go`):

| Поле | Тип | Описание |
|------|-----|----------|
| id | uuid.UUID | Первичный ключ |
| action | string | Действие (create, update, delete, login, logout и т.д.) |
| entity_type | string | Тип сущности (issue, project, user, test_case и т.д.) |
| entity_id | string | ID сущности |
| user_id | uuid.UUID | FK на пользователя (опционально) |
| project_id | uuid.UUID | FK на проект (опционально) |
| details | JSON (map[string]interface{}) | Дополнительные данные |
| ip_address | string | IP-адрес |
| user_agent | string | User-Agent |
| created_at | time.Time | Время события |

**Индексы:** action, entity_type, user_id, project_id, created_at

### JetStream consumer

```
Модули -> Outbox -> NATS -> AUDIT stream -> audit-logger consumer -> AuditLog table
```

**Consumer `audit-logger`:**
- Stream: AUDIT
- Subjects: `project.>`, `tracker.>`, `messenger.>`, `tms.>`, `calls.>`, `auth.>`
- Durable: yes (at-least-once delivery)
- Обработка: парсинг NATS-субъекта -> извлечение action и entity_type -> создание записи AuditLog через Repository

### Soft-delete

Паттерн `deleted_at` добавлен на 3 схемы:

| Схема | Файл |
|-------|------|
| Issue | `backend/ent/schema/issue.go` |
| TestCase | `backend/ent/schema/testcase.go` |
| FileMetadata | `backend/ent/schema/filemetadata.go` |

**Поле:** `field.Time("deleted_at").Optional().Nillable()` с индексом

**Модификация запросов:**
```go
// Все обычные запросы фильтруют удаленные
query.Where(issue.DeletedAtIsNil())

// Удаление = установка deleted_at
client.Issue.UpdateOneID(id).SetDeletedAt(time.Now()).Save(ctx)

// Восстановление = сброс deleted_at
client.Issue.UpdateOneID(id).ClearDeletedAt().Save(ctx)
```

**Модифицированные методы в трекере:**
- `FindIssueByKey` -- добавлен `.Where(issue.DeletedAtIsNil())`
- `FindIssueByID` -- добавлен `.Where(issue.DeletedAtIsNil())`
- `ListIssues` -- добавлен `.Where(issue.DeletedAtIsNil())`
- `GetBoardData` -- добавлен `.Where(issue.DeletedAtIsNil())`
- `DeleteIssue` -- изменен с hard delete на soft delete (SetDeletedAt)

**Модифицированные методы в TMS:**
- `GetTestCase` -- добавлен `.Where(testcase.DeletedAtIsNil())`
- `FindTestCaseByKey` -- добавлен `.Where(testcase.DeletedAtIsNil())`
- `ListTestCases` -- добавлен `.Where(testcase.DeletedAtIsNil())`
- `DeleteTestCase` -- изменен с hard delete на soft delete

### 30-дневная retention и purge cron

```go
// main.go -- запуск cron
go func() {
    ticker := time.NewTicker(24 * time.Hour)
    for range ticker.C {
        auditService.PurgeExpired(ctx)
    }
}()

// PurgeExpired удаляет записи старше 30 дней (hard delete)
func (s *AuditService) PurgeExpired(ctx context.Context) {
    cutoff := time.Now().AddDate(0, 0, -TrashRetentionDays) // -30 дней
    s.db.Issue.Delete().Where(
        issue.DeletedAtNotNil(),
        issue.DeletedAtLT(cutoff),
    ).Exec(ctx)
    // Аналогично для TestCase и FileMetadata
}
```

### Resource limits

Ключи лимитов в SystemSetting (key-value таблица):

| Ключ | Значение по умолчанию | Описание |
|------|----------------------|----------|
| `limit.file.max_size_mb` | "50" | Максимальный размер файла (MB) |
| `limit.user.max_projects` | "10" | Максимум проектов на пользователя |
| `limit.project.max_storage_mb` | "1024" | Максимум хранилища на проект (MB) |

Enforcement points:
- Загрузка файлов: `storage.Service.Upload()` проверяет max_size_mb
- Создание проекта: проверяет max_projects для текущего пользователя

### GDPR-удаление

Каскадная процедура удаления данных пользователя:

```go
func (s *AuditService) DeleteUserGDPR(ctx context.Context, userID uuid.UUID) error {
    // 1. Очистить assignee у всех задач пользователя
    s.db.Issue.Update().
        Where(issue.HasAssigneeWith(user.ID(userID))).
        ClearAssignee().
        Save(ctx)

    // 2. Удалить все сессии пользователя
    s.db.Session.Delete().
        Where(session.HasUserWith(user.ID(userID))).
        Exec(ctx)

    // 3. Удалить запись пользователя из PostgreSQL
    s.db.User.DeleteOneID(userID).Exec(ctx)

    // 4. Анонимизация данных в MongoDB (сообщения и т.д.)
    // ... замена sender_id на "deleted_user"
}
```

### Auth middleware: проверка деактивированных

```go
// backend/internal/auth/middleware.go
u, err := dbClient.User.Get(ctx, userID)
if u.Status == user.StatusDeactivated {
    return fiber.NewError(fiber.StatusUnauthorized, "пользователь деактивирован")
}
```

---

## API reference: все новые endpoints фазы 6

### Уведомления (/api/notifications)

```bash
# Список уведомлений с пагинацией
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:3000/api/notifications?page=1&per_page=20&is_read=false"

# Количество непрочитанных
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:3000/api/notifications/unread-count"

# Пометить прочитанным
curl -X PATCH -H "Authorization: Bearer $TOKEN" \
  "http://localhost:3000/api/notifications/{id}/read"

# Пометить все прочитанными
curl -X PATCH -H "Authorization: Bearer $TOKEN" \
  "http://localhost:3000/api/notifications/read-all"

# Получить настройки
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:3000/api/notifications/preferences"

# Обновить настройку
curl -X PUT -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"notification_type":"assigned","channel":"telegram","enabled":false}' \
  "http://localhost:3000/api/notifications/preferences"
```

### Поиск (/api/search, /api/filters)

```bash
# Глобальный поиск
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:3000/api/search?q=логин&type=issue,test_case&limit=10"

# Создать сохраненный фильтр
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Мои баги","filter_config":{"type":"bug"},"project_id":"uuid"}' \
  "http://localhost:3000/api/filters"

# Список фильтров проекта
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:3000/api/filters?project_id=uuid"

# Удалить фильтр
curl -X DELETE -H "Authorization: Bearer $TOKEN" \
  "http://localhost:3000/api/filters/{id}"

# Расшарить фильтр
curl -X PATCH -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"shared":true}' \
  "http://localhost:3000/api/filters/{id}/share"
```

### Дашборды (/api/dashboard)

```bash
# Персональный дашборд
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:3000/api/dashboard/personal"

# Проектный дашборд
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:3000/api/dashboard/project/{projectId}"

# Админ-дашборд (только Admin)
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:3000/api/dashboard/admin"

# Получить раскладку
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:3000/api/dashboard/layout?type=personal"

# Сохранить раскладку
curl -X PUT -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"dashboard_type":"personal","items":[{"widget_type":"my_tasks","x":0,"y":0,"w":6,"h":4}]}' \
  "http://localhost:3000/api/dashboard/layout"
```

### Администрирование (/api/admin) -- только Admin

```bash
# Журнал аудита с фильтрами
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:3000/api/admin/audit?action=delete&entity_type=issue&page=1&per_page=50"

# Получить лимиты
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:3000/api/admin/limits"

# Обновить лимит
curl -X PUT -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"value":"100"}' \
  "http://localhost:3000/api/admin/limits/limit.file.max_size_mb"

# Список корзины
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:3000/api/admin/trash?type=issue&page=1"

# Восстановить из корзины
curl -X POST -H "Authorization: Bearer $TOKEN" \
  "http://localhost:3000/api/admin/trash/issue/{id}/restore"

# Деактивировать пользователя
curl -X POST -H "Authorization: Bearer $TOKEN" \
  "http://localhost:3000/api/admin/users/{id}/deactivate"

# Переназначить задачи
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"to_user_id":"uuid"}' \
  "http://localhost:3000/api/admin/users/{id}/reassign"

# GDPR-удаление
curl -X DELETE -H "Authorization: Bearer $TOKEN" \
  "http://localhost:3000/api/admin/users/{id}/gdpr"
```

---

## Диаграммы потоков данных

### Поиск: поток запроса

```
Browser                   Fiber Handler            SearchService           PostgreSQL / MongoDB
   |                           |                        |                        |
   |  GET /api/search?q=test   |                        |                        |
   |-------------------------->|                        |                        |
   |                           |  Search(ctx,q,user,    |                        |
   |                           |    types,limit)         |                        |
   |                           |----------------------->|                        |
   |                           |                        |  getUserProjectIDs()   |
   |                           |                        |----------------------->| PG
   |                           |                        |<-----------------------|
   |                           |                        |                        |
   |                           |                        |  go searchIssues()     |
   |                           |                        |--+                     |
   |                           |                        |  | go searchTestCases()|
   |                           |                        |  +--+                  |
   |                           |                        |  |  | go searchUsers() |
   |                           |                        |  |  +--+               |
   |                           |                        |  |  |  | ...           |
   |                           |                        |  |  |  |               |
   |                           |                        |  |  |  |  searchMsgs() |
   |                           |                        |  |  |  +-------------->| MongoDB
   |                           |                        |  |  |  |<--------------|
   |                           |                        |  |  |  |               |
   |                           |                        |<-+--+--+ wg.Wait()    |
   |                           |                        |                        |
   |                           |                        |  sort by score DESC    |
   |                           |                        |  limit results         |
   |                           |<-----------------------|                        |
   |  200 { results, total }   |                        |                        |
   |<--------------------------|                        |                        |
```

### Дашборд: поток с кешем

```
Browser                   Fiber Handler            DashboardService         Redis           PostgreSQL
   |                           |                        |                    |                  |
   |  GET /dashboard/project/X |                        |                    |                  |
   |-------------------------->|                        |                    |                  |
   |                           |  GetProjectStats(X)    |                    |                  |
   |                           |----------------------->|                    |                  |
   |                           |                        |  GET project_stats:|                  |
   |                           |                        |    X               |                  |
   |                           |                        |------------------->|                  |
   |                           |                        |  nil (miss)        |                  |
   |                           |                        |<-------------------|                  |
   |                           |                        |                    |                  |
   |                           |                        |  Query issues      |                  |
   |                           |                        |----------------------------------->  |
   |                           |                        |  Query test_runs   |                  |
   |                           |                        |----------------------------------->  |
   |                           |                        |  <results>         |                  |
   |                           |                        |<-----------------------------------|
   |                           |                        |                    |                  |
   |                           |                        |  SET project_stats:|                  |
   |                           |                        |    X (TTL 5min)    |                  |
   |                           |                        |------------------->|                  |
   |                           |                        |  OK                |                  |
   |                           |                        |<-------------------|                  |
   |                           |<-----------------------|                    |                  |
   |  200 { stats }            |                        |                    |                  |
   |<--------------------------|                        |                    |                  |
```

### Аудит: поток событий

```
User Action         Module           Outbox Worker        NATS JetStream       Audit Consumer      PostgreSQL
    |                  |                   |                    |                    |                  |
    | Create Issue     |                   |                    |                    |                  |
    |----------------->|                   |                    |                    |                  |
    |                  | INSERT outbox      |                    |                    |                  |
    |                  |------------------>|                    |                    |                  |
    |                  |                   |  Publish           |                    |                  |
    |                  |                   |  tracker.issue.    |                    |                  |
    |                  |                   |    created         |                    |                  |
    |                  |                   |------------------->|                    |                  |
    |                  |                   |                    |  Deliver to        |                  |
    |                  |                   |                    |  audit-logger      |                  |
    |                  |                   |                    |------------------->|                  |
    |                  |                   |                    |                    | INSERT audit_log |
    |                  |                   |                    |                    |----------------->|
    |                  |                   |                    |                    | Ack              |
    |                  |                   |                    |<-------------------|                  |
```

---

*Фаза: 06-notifications-search-dashboards*
*Дата документации: 2026-03-21*
