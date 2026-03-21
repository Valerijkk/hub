# Phase 6: Notifications, Search & Dashboards - Research

**Researched:** 2026-03-21
**Domain:** Cross-cutting aggregation features (notifications, full-text search, dashboards, audit, resource limits, trash/restore)
**Confidence:** HIGH

## Summary

Phase 6 delivers four cross-cutting feature groups that aggregate data from all five existing modules (auth, tracker, messenger, TMS, calls). The notification system consumes NATS events from every module, dispatches to in-app bell and Telegram bot, and respects per-user preferences with quiet hours. Global search runs parallel queries against PostgreSQL tsvector and MongoDB text indexes, merging results on the backend. Dashboards use Tremor Raw chart components with react-grid-layout for drag-and-drop widget positioning. Audit, resource limits, trash/restore, and user deletion complete the administrative control plane.

The architecture follows the established modular monolith pattern: new `notification`, `search`, `dashboard`, and `audit` packages under `backend/internal/`, each with handler/service/repository layers. The notification module is the most architecturally significant -- it introduces JetStream durable consumers for reliable notification delivery, a Telegram bot via telego, and a new `notification_preferences` + `notification` Ent schemas. Search is straightforward -- raw SQL tsvector queries via Ent's `ExecContext` for PG entities, plus MongoDB `$text` for messages, merged and ranked by the search service. Dashboards are frontend-heavy, with backend providing aggregation endpoints that use cached Redis counters or direct queries.

**Primary recommendation:** Build notification infrastructure first (NATS consumers + Ent schemas + preferences), then search, then dashboards, then audit/limits. Each builds on the previous -- notifications test the event consumption pipeline, search adds the Ctrl+K overlay, dashboards use both notification and search infrastructure, and audit/limits are independent administrative features.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| NOTF-01 | Триггеры: назначение, обновление, @упоминание, смена статуса, сообщение в чате, звонок, тест-ран, спринт, дедлайн, блокировщик | NATS consumers for all module events; existing events in tracker/messenger/tms/calls events.go cover most triggers; need to add deadline/blocker checks via cron |
| NOTF-02 | Каналы доставки: в приложении (иконка колокольчика) и Telegram-бот | In-app via WS `notification.user.{userID}` subject + Zustand store; Telegram via telego bot with SendMessage API |
| NOTF-03 | Настройки: конфигурация по типу, по каналу доставки и по проекту | NotificationPreference Ent schema with (user_id, notification_type, channel, project_id) composite; defaults on first access |
| NOTF-04 | Тихие часы с возможностью прорыва для срочных уведомлений | User timezone from profile + quiet_hours_start/end fields; urgent flag on notification bypasses quiet hours |
| SRCH-01 | Фильтры задач: статус, приоритет, исполнитель, тип; сохранённые и общие фильтры | SavedFilter Ent schema with JSON filter_config; share via project membership check |
| SRCH-02 | Ctrl+K глобальный поиск: задачи, тест-кейсы, пользователи, проекты, каналы, файлы | PG tsvector for issues/test_cases/users/projects/files + MongoDB $text for messages; parallel query + merge on backend |
| DASH-01 | Личный дашборд с настраиваемыми виджетами (drag & drop, размеры) | react-grid-layout for positioning; DashboardLayout Ent schema (user_id, layout JSON); widget data from aggregation endpoints |
| DASH-02 | Дашборд проекта с шаблонами (Dev/QA/Manager) | Same layout system with project_id scope; template presets stored as JSON; Tremor charts for burndown/velocity/trends |
| DASH-03 | Дашборд администратора: пользователи, проекты, задачи, активность, хранилище, здоровье | Admin-only endpoints; system-wide aggregation queries; DAU/WAU/MAU from audit_log timestamps |
| AUDT-01 | Полный журнал аудита: вход/выход, задачи, проекты, TMS, мессенджер -- Admin просматривает с фильтрами | AuditLog Ent schema; JetStream durable consumer on all module events; paginated admin API with filters |
| AUDT-02 | Настраиваемые лимиты: размер файла, проекты на пользователя, хранилище на проект | SystemSetting schema already exists; extend with limit-specific keys; enforce in existing upload/create handlers |
| AUDT-03 | 30-дневная корзина для задач, тест-кейсов, проектов, файлов; Admin/Owner может восстановить | Soft-delete pattern (deleted_at field) already on Project; extend to Issue, TestCase, FileMetadata; restore endpoint clears deleted_at |
| AUDT-04 | Удаление пользователя: деактивация, переназначение задач, полное удаление данных (опция GDPR) | User.status field (active/deactivated); reassign issues via bulk update; GDPR: cascade delete + anonymize MongoDB messages |
| UI-05 | Горячие клавиши: Ctrl+K (поиск), Ctrl+N (новая задача) | Global keyboard event listener in app shell; Ctrl+K opens CommandPalette (search dialog); Ctrl+N opens issue create dialog |
</phase_requirements>

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| telego | v0.33+ | Telegram bot notifications | 100% Bot API coverage, code-generated types, Go-native, fasthttp transport, MIT |
| Tremor Raw | copy-paste | Dashboard chart components (AreaChart, BarChart, DonutChart, KPI cards) | Tailwind-native, built on Recharts + Radix UI, copy-paste model like shadcn/ui |
| react-grid-layout | 2.2.2 | Drag-and-drop dashboard widget grid | De facto standard for React dashboard layouts, responsive breakpoints, resize handles |
| recharts | 3.8.0 | Chart rendering engine (dependency of Tremor Raw) | Already used by Tremor internally, composable chart API |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| cmdk | 1.0+ | Command palette UI for Ctrl+K search | Accessible, keyboard-navigable, shadcn-compatible command menu |
| date-fns-tz | latest | Timezone handling for quiet hours | Convert user timezone to UTC for quiet hour window checks |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Tremor Raw | Recharts directly | Tremor provides pre-styled dashboard patterns (KPI cards, stat sections); raw Recharts requires more custom styling |
| react-grid-layout | @dnd-kit/core | react-grid-layout is purpose-built for dashboard grids with resize; dnd-kit is general-purpose and needs more setup |
| cmdk | Custom dialog | cmdk provides keyboard navigation, fuzzy matching, sections out of the box |
| telego | go-telegram/bot | telego chosen in PROJECT.md; both are viable, telego has 100% API codegen |

**Installation (frontend):**
```bash
npm install react-grid-layout cmdk recharts
# Tremor Raw: copy components from tremor.so (AreaChart, BarChart, DonutChart, Tracker, etc.)
```

**Installation (backend):**
```bash
go get github.com/mymmrac/telego
```

## Architecture Patterns

### Recommended Project Structure

```
backend/internal/
  notification/
    handler.go          # GET /api/notifications, PATCH mark-read, PUT preferences
    service.go          # Dispatch logic, preference checks, quiet hours, batching
    telegram.go         # telego bot wrapper, message formatting
    consumer.go         # NATS JetStream consumer for all module events
    events.go           # NotificationType constants, event mapping
    repository.go       # Notification + NotificationPreference Ent queries
    dto.go              # Request/response types
  search/
    handler.go          # GET /api/search?q=...&type=...
    service.go          # Parallel PG+Mongo query, merge, rank, permission filter
    indexer.go          # tsvector update triggers (Atlas migration, not Go code)
    dto.go              # SearchResult union type
  dashboard/
    handler.go          # GET /api/dashboard/personal, /project/:id, /admin
    service.go          # Aggregation queries, Redis-cached counters
    dto.go              # Widget data types
  audit/
    handler.go          # GET /api/admin/audit (paginated, filtered)
    service.go          # AuditLog queries with filters
    consumer.go         # JetStream durable consumer for audit events
    repository.go       # AuditLog Ent queries
    dto.go

frontend/src/
  components/
    notifications/
      notification-bell.tsx      # Bell icon with unread count badge
      notification-panel.tsx     # Dropdown panel with notification list
      notification-settings.tsx  # Preferences page
    search/
      command-palette.tsx        # Ctrl+K overlay using cmdk
      search-results.tsx         # Result rendering by entity type
      saved-filters.tsx          # Saved/shared filter management
    dashboard/
      widget-grid.tsx            # react-grid-layout wrapper
      widgets/                   # Individual widget components
        my-tasks.tsx
        unread-messages.tsx
        upcoming-calls.tsx
        task-status-chart.tsx
        burndown-chart.tsx
        velocity-chart.tsx
        test-trends.tsx
        team-activity.tsx
        system-health.tsx
        storage-usage.tsx
    admin/
      audit-log.tsx              # Admin audit log viewer
      resource-limits.tsx        # Admin limit configuration
      user-management.tsx        # Deactivation, reassignment, GDPR delete
  stores/
    notification-store.ts        # Zustand: notifications[], unreadCount
    search-store.ts              # Zustand: recent searches, saved filters
    dashboard-store.ts           # Zustand: layout configs per dashboard type
```

### Pattern 1: NATS JetStream Durable Consumer for Notifications

**What:** A single JetStream consumer processes all domain events and creates notification records. Unlike ephemeral core NATS subscriptions used by the WS bridge, this uses durable consumers to guarantee no notifications are lost even if the server restarts.

**When to use:** Notification dispatch and audit log recording -- both require at-least-once delivery guarantees.

**Example:**
```go
// notification/consumer.go
func (c *Consumer) Start(ctx context.Context) error {
    // Создаём JetStream stream для уведомлений (если не существует)
    stream, err := c.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
        Name:     "NOTIFICATIONS",
        Subjects: []string{
            "project.>",
            "tracker.>",  // Через outbox -> NATS
            "messenger.channel.>",
            "tms.>",
            "calls.channel.>",
        },
        Retention: jetstream.WorkQueuePolicy,
        MaxAge:    7 * 24 * time.Hour,
    })

    // Durable consumer -- переживает рестарты
    cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
        Durable:       "notification-dispatcher",
        AckPolicy:     jetstream.AckExplicitPolicy,
        MaxDeliver:    5,
        AckWait:       30 * time.Second,
        FilterSubject: ">", // Все subjects в стриме
    })

    // Обработка сообщений
    iter, _ := cons.Messages()
    go func() {
        for {
            msg, err := iter.Next()
            if err != nil { return }
            c.processEvent(ctx, msg)
            msg.Ack()
        }
    }()
    return nil
}

func (c *Consumer) processEvent(ctx context.Context, msg jetstream.Msg) {
    // Маршрутизация по subject -> тип уведомления
    notifType := c.mapSubjectToType(msg.Subject())
    if notifType == "" { return }

    // Определяем получателей
    recipients := c.resolveRecipients(ctx, msg)

    for _, userID := range recipients {
        // Проверяем настройки пользователя
        if !c.shouldNotify(ctx, userID, notifType, projectID) { continue }
        // Проверяем тихие часы
        if c.isQuietHour(ctx, userID) && !isUrgent(notifType) { continue }

        // Сохраняем в БД
        notif := c.repo.CreateNotification(ctx, userID, notifType, payload)

        // Отправляем в приложение через NATS -> WS bridge
        c.nats.Publish("notification.user."+userID, notifData)

        // Отправляем в Telegram (если включён)
        if c.shouldSendTelegram(ctx, userID, notifType) {
            c.telegram.Send(ctx, userID, notif)
        }
    }
}
```

### Pattern 2: Parallel Search with Merge

**What:** Global search sends parallel queries to PostgreSQL (tsvector) and MongoDB ($text), collects results, filters by user permissions, and merges into a ranked list.

**When to use:** Ctrl+K global search across all entity types.

**Example:**
```go
// search/service.go
func (s *SearchService) Search(ctx context.Context, query string, userID uuid.UUID, types []string) ([]SearchResult, error) {
    var wg sync.WaitGroup
    var mu sync.Mutex
    var results []SearchResult

    // PG: задачи, тест-кейсы, пользователи, проекты, файлы
    if shouldSearch(types, "issues", "test_cases", "users", "projects", "files") {
        wg.Add(1)
        go func() {
            defer wg.Done()
            pgResults := s.searchPostgres(ctx, query, userID, types)
            mu.Lock()
            results = append(results, pgResults...)
            mu.Unlock()
        }()
    }

    // MongoDB: сообщения
    if shouldSearch(types, "messages") {
        wg.Add(1)
        go func() {
            defer wg.Done()
            mongoResults := s.searchMongo(ctx, query, userID)
            mu.Lock()
            results = append(results, mongoResults...)
            mu.Unlock()
        }()
    }

    wg.Wait()
    return rankAndLimit(results, 20), nil
}

// searchPostgres использует tsvector для полнотекстового поиска
func (s *SearchService) searchPostgres(ctx context.Context, query string, userID uuid.UUID, types []string) []SearchResult {
    // Сырой SQL через Ent ExecContext -- tsvector не поддерживается в Ent DSL
    // SELECT id, issue_key, title, ts_rank(search_vector, plainto_tsquery('russian', $1)) AS rank
    // FROM issues WHERE search_vector @@ plainto_tsquery('russian', $1)
    // AND project_id IN (SELECT project_id FROM project_members WHERE user_id = $2)
    // ORDER BY rank DESC LIMIT 10
}
```

### Pattern 3: Dashboard Widget Data Aggregation

**What:** Each widget type has a dedicated aggregation endpoint. Personal dashboard fetches data from multiple endpoints in parallel on the frontend. Project/admin dashboards use Redis-cached counters updated via NATS events to avoid expensive live queries.

**When to use:** All dashboard views.

**Example:**
```go
// dashboard/service.go

// GetPersonalDashboardData возвращает данные для персонального дашборда.
func (s *DashboardService) GetPersonalDashboardData(ctx context.Context, userID uuid.UUID) (*PersonalData, error) {
    return &PersonalData{
        MyTasks:      s.getMyTasks(ctx, userID),       // PG: issues where assignee = userID
        UnreadCount:  s.getUnreadCount(ctx, userID),    // Redis: уже есть в messenger unread service
        UpcomingCalls: s.getUpcomingCalls(ctx, userID), // PG: calls with scheduled_at > now
        RecentRuns:   s.getRecentRuns(ctx, userID),     // PG: test_runs created by user
        Deadlines:    s.getDeadlines(ctx, userID),      // PG: issues with due_date approaching
        TimeToday:    s.getTimeToday(ctx, userID),      // PG: sum time_logs today
    }, nil
}

// GetProjectStats возвращает агрегированные данные проекта.
// Для дешёвого доступа используем Redis-cached счётчики.
func (s *DashboardService) GetProjectStats(ctx context.Context, projectID uuid.UUID) (*ProjectStats, error) {
    // Пробуем кеш (обновляется при каждом событии через NATS consumer)
    cached, err := s.redis.Get(ctx, fmt.Sprintf("dashboard:project:%s", projectID)).Result()
    if err == nil {
        var stats ProjectStats
        json.Unmarshal([]byte(cached), &stats)
        return &stats, nil
    }
    // Fallback: живой запрос
    return s.computeProjectStats(ctx, projectID)
}
```

### Pattern 4: WS Bridge Extension for Notifications

**What:** Add a new NATS subscription in the WS bridge for `notification.user.>` that delivers notifications directly to the target user's WebSocket connection via `SendToUser`.

**Example:**
```go
// ws/bridge.go -- новая подписка для уведомлений
notifSub, err := b.nats.Subscribe("notification.user.>", func(msg *nats.Msg) {
    parts := strings.SplitN(msg.Subject, ".", 3)
    if len(parts) < 3 { return }
    userID := parts[2]

    var payload any
    if len(msg.Data) > 0 {
        _ = json.Unmarshal(msg.Data, &payload)
    }

    b.hub.SendToUser(userID, ServerMessage{
        Type:    TypeEvent,
        Event:   "notification.new",
        Payload: payload,
    })
})
```

### Anti-Patterns to Avoid

- **Live dashboard queries on every page load:** Dashboard aggregation queries (COUNT, SUM, GROUP BY across thousands of rows) are expensive. Use Redis-cached counters updated incrementally via NATS events, with a periodic full recompute (cron every 5 min) as consistency guarantee.
- **Per-module search endpoints:** Do NOT build separate search endpoints in tracker, messenger, TMS. Centralize all search logic in `search/service.go` with a single `/api/search` endpoint.
- **Storing notification preferences as bitmask:** Use relational rows (user_id, type, channel, project_id, enabled) for flexibility. Bitmasks make per-project overrides extremely complex.
- **Polling for new notifications:** Use WebSocket push via NATS bridge. Frontend should NOT poll `/api/notifications` -- only fetch on initial load, then rely on WS events.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Dashboard grid layout | Custom CSS grid with drag handlers | react-grid-layout | Resize handles, responsive breakpoints, layout persistence, constraint system |
| Command palette (Ctrl+K) | Custom dialog with keyboard nav | cmdk | Accessible, keyboard-navigable, fuzzy matching, sections, shadcn-compatible |
| Chart rendering | Custom SVG charts | Tremor Raw (Recharts) | Area/Bar/Donut/Line charts pre-styled for Tailwind, KPI cards included |
| Telegram bot API | Raw HTTP to api.telegram.org | telego | Code-generated types, middleware, error handling, webhook support |
| Full-text search index | Application-level word splitting | PG tsvector + GIN index | Stemming, ranking, language configs, phrase search built into PG |
| Notification dedup | Custom "sent" tracking | Composite unique index (user_id, type, entity_id, created_at truncated) | DB-level dedup prevents double notifications on event replay |

**Key insight:** The notification system is the most complex new subsystem in this phase. It must consume events from 5 modules, check per-user/per-project preferences, enforce quiet hours with timezone awareness, deliver to two channels (in-app + Telegram), and handle deduplication. Every other feature (search, dashboards, audit) is simpler in comparison.

## Common Pitfalls

### Pitfall 1: tsvector Language Configuration
**What goes wrong:** Default 'english' tsvector config cannot stem Russian words. Search for "задача" fails to match "задачу" or "задач".
**Why it happens:** PG full-text search is language-dependent; stemming only works with the correct language config.
**How to avoid:** Use 'russian' config for text fields that contain Russian content. Since Hub supports RU/EN, use `'simple'` config (no stemming, case-insensitive) as a pragmatic default that works for both languages, or create two tsvector columns (one per language) if quality matters.
**Recommended:** Use `'simple'` config -- it performs case-insensitive prefix matching without stemming, which works for both Russian and English at 50 users.

### Pitfall 2: N+1 Permission Checks in Search Results
**What goes wrong:** Search returns 50 results, then checking `canViewProject` for each result issues 50 separate DB queries.
**Why it happens:** Permission filtering applied after query instead of in the query.
**How to avoid:** Pre-fetch user's project membership list (already cached in RBAC cache), filter search results in the WHERE clause: `project_id IN (user's project IDs)`.

### Pitfall 3: Notification Storm on Bulk Operations
**What goes wrong:** Bulk-assigning 20 tasks sends 20 individual notifications to the same user in rapid succession.
**Why it happens:** Each task update independently publishes an event; notification consumer processes each one.
**How to avoid:** Batch/debounce notifications: accumulate events for the same user within a 2-second window, then send a single "You were assigned to 20 tasks" notification.

### Pitfall 4: Telegram Rate Limits
**What goes wrong:** Sending 50+ Telegram messages in rapid succession triggers Bot API rate limiting (30 messages/second to different chats, 1 message/second to same chat).
**Why it happens:** Telegram enforces rate limits at the bot level.
**How to avoid:** Use a buffered channel + goroutine worker for Telegram sends with per-chat rate limiting. Queue messages, send with 35ms delay between different chats.

### Pitfall 5: Dashboard Data Staleness vs Query Cost
**What goes wrong:** Either dashboards show stale data (Redis cache too long) or every page load triggers expensive COUNT/SUM queries.
**Why it happens:** No middle ground between live queries and aggressive caching.
**How to avoid:** Hybrid approach: Redis counters updated on each NATS event (increment/decrement), with full recompute cron every 5 minutes as consistency check. Frontend shows "updated X seconds ago" indicator.

### Pitfall 6: react-grid-layout TypeScript Types
**What goes wrong:** react-grid-layout's TypeScript definitions are incomplete; `Layout` type requires careful handling.
**Why it happens:** Library was originally JavaScript-only; TS types were community-added.
**How to avoid:** Define strict widget layout types in the codebase, cast react-grid-layout results through them.

## Code Examples

### Ent Schema: Notification

```go
// ent/schema/notification.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "github.com/google/uuid"
    "time"
)

type Notification struct {
    ent.Schema
}

func (Notification) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
        field.String("notification_type").NotEmpty().
            Comment("assigned/mentioned/status_changed/comment/call/run_completed/deadline/blocker"),
        field.String("title").NotEmpty().MaxLen(200),
        field.Text("body").Optional(),
        field.String("entity_type").Optional().Comment("issue/test_case/channel/call/run"),
        field.String("entity_id").Optional(),
        field.UUID("project_id", uuid.UUID{}).Optional().Nillable(),
        field.UUID("actor_id", uuid.UUID{}).Optional().Nillable().
            Comment("Кто вызвал уведомление"),
        field.Bool("is_read").Default(false),
        field.Bool("is_urgent").Default(false),
        field.Time("created_at").Default(time.Now).Immutable(),
        field.Time("read_at").Optional().Nillable(),
    }
}

func (Notification) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("user", User.Type).Ref("notifications").Unique().Required(),
    }
}

func (Notification) Indexes() []ent.Index {
    return []ent.Index{
        index.Edges("user").Fields("is_read", "created_at"),
        index.Edges("user").Fields("created_at"),
        index.Fields("notification_type"),
    }
}
```

### Ent Schema: NotificationPreference

```go
// ent/schema/notificationpreference.go
type NotificationPreference struct {
    ent.Schema
}

func (NotificationPreference) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
        field.String("notification_type").NotEmpty(),
        field.Enum("channel").Values("in_app", "telegram"),
        field.UUID("project_id", uuid.UUID{}).Optional().Nillable().
            Comment("nil = глобальная настройка, UUID = настройка для проекта"),
        field.Bool("enabled").Default(true),
    }
}

func (NotificationPreference) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("user", User.Type).Ref("notification_preferences").Unique().Required(),
    }
}

func (NotificationPreference) Indexes() []ent.Index {
    return []ent.Index{
        // Уникальность: один пользователь + тип + канал + проект
        index.Edges("user").Fields("notification_type", "channel", "project_id").Unique(),
    }
}
```

### Ent Schema: AuditLog

```go
// ent/schema/auditlog.go
type AuditLog struct {
    ent.Schema
}

func (AuditLog) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
        field.String("action").NotEmpty().Comment("login/logout/create/update/delete/..."),
        field.String("entity_type").NotEmpty().Comment("user/project/issue/test_case/channel/call"),
        field.String("entity_id").Optional(),
        field.UUID("user_id", uuid.UUID{}).Comment("Кто выполнил действие"),
        field.UUID("project_id", uuid.UUID{}).Optional().Nillable(),
        field.JSON("details", map[string]interface{}{}).Optional().Comment("Дополнительные данные"),
        field.String("ip_address").Optional(),
        field.String("user_agent").Optional(),
        field.Time("created_at").Default(time.Now).Immutable(),
    }
}

func (AuditLog) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("user_id", "created_at"),
        index.Fields("entity_type", "created_at"),
        index.Fields("project_id", "created_at"),
        index.Fields("action", "created_at"),
        index.Fields("created_at"),
    }
}
```

### Ent Schema: SavedFilter

```go
// ent/schema/savedfilter.go
type SavedFilter struct {
    ent.Schema
}

func (SavedFilter) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
        field.String("name").NotEmpty().MaxLen(100),
        field.JSON("filter_config", map[string]interface{}{}).
            Comment("JSON: {status:[], priority:[], assignee:[], type:[], ...}"),
        field.Bool("is_shared").Default(false),
        field.UUID("project_id", uuid.UUID{}).Optional().Nillable(),
        field.Time("created_at").Default(time.Now).Immutable(),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (SavedFilter) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("owner", User.Type).Ref("saved_filters").Unique().Required(),
    }
}
```

### Ent Schema: DashboardLayout

```go
// ent/schema/dashboardlayout.go
type DashboardLayout struct {
    ent.Schema
}

func (DashboardLayout) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
        field.Enum("dashboard_type").Values("personal", "project", "admin"),
        field.UUID("project_id", uuid.UUID{}).Optional().Nillable(),
        field.JSON("layout", []map[string]interface{}{}).
            Comment("react-grid-layout items: [{i, x, y, w, h, widget_type, config}]"),
        field.Time("created_at").Default(time.Now).Immutable(),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (DashboardLayout) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("user", User.Type).Ref("dashboard_layouts").Unique().Required(),
    }
}

func (DashboardLayout) Indexes() []ent.Index {
    return []ent.Index{
        index.Edges("user").Fields("dashboard_type", "project_id").Unique(),
    }
}
```

### Atlas Migration: tsvector + GIN Index

```sql
-- Миграция: добавляем search_vector колонки и GIN индексы
-- Выполняется через Atlas (atlas migrate diff)

-- Issues: поисковый вектор из title + description_text + issue_key
ALTER TABLE issues ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(issue_key, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(description_text, '')), 'B')
    ) STORED;
CREATE INDEX idx_issues_search ON issues USING GIN (search_vector);

-- Test cases: поисковый вектор из case_key + title + description
ALTER TABLE test_cases ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(case_key, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(description, '')), 'B')
    ) STORED;
CREATE INDEX idx_test_cases_search ON test_cases USING GIN (search_vector);

-- Users: поисковый вектор из name + email
ALTER TABLE users ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (
        to_tsvector('simple', coalesce(name, '') || ' ' || coalesce(email, '') || ' ' || coalesce(display_name, ''))
    ) STORED;
CREATE INDEX idx_users_search ON users USING GIN (search_vector);

-- Projects: поисковый вектор из name + key + description
ALTER TABLE projects ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(key, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(name, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(description, '')), 'B')
    ) STORED;
CREATE INDEX idx_projects_search ON projects USING GIN (search_vector);

-- Files: поисковый вектор из filename
ALTER TABLE file_metadata ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (
        to_tsvector('simple', coalesce(filename, ''))
    ) STORED;
CREATE INDEX idx_files_search ON file_metadata USING GIN (search_vector);
```

### MongoDB Text Index for Messages

```go
// search/service.go -- поиск по сообщениям MongoDB
func (s *SearchService) searchMessages(ctx context.Context, query string, userChannelIDs []string) []SearchResult {
    // MongoDB text index уже создан в messenger EnsureIndexes
    // Фильтруем по каналам, в которых пользователь состоит
    filter := bson.M{
        "$text":      bson.M{"$search": query},
        "channel_id": bson.M{"$in": userChannelIDs},
    }
    opts := options.Find().
        SetProjection(bson.M{"score": bson.M{"$meta": "textScore"}}).
        SetSort(bson.M{"score": bson.M{"$meta": "textScore"}}).
        SetLimit(10)

    cursor, err := s.msgCollection.Find(ctx, filter, opts)
    // ... обработка результатов
}
```

### Telegram Notification via telego

```go
// notification/telegram.go
package notification

import (
    "context"
    "fmt"
    "github.com/mymmrac/telego"
    tu "github.com/mymmrac/telego/telegoutil"
)

type TelegramSender struct {
    bot *telego.Bot
}

func NewTelegramSender(botToken string) (*TelegramSender, error) {
    bot, err := telego.NewBot(botToken)
    if err != nil {
        return nil, fmt.Errorf("не удалось создать Telegram бот: %w", err)
    }
    return &TelegramSender{bot: bot}, nil
}

// Send отправляет уведомление пользователю в Telegram.
// telegramID берётся из user.telegram_id в БД.
func (t *TelegramSender) Send(ctx context.Context, telegramChatID int64, notif *NotificationDTO) error {
    text := fmt.Sprintf("*%s*\n%s", escapeMarkdown(notif.Title), notif.Body)

    msg := tu.Message(
        tu.ID(telegramChatID),
        text,
    ).WithParseMode(telego.ModeMarkdownV2)

    _, err := t.bot.SendMessage(msg)
    return err
}
```

### Frontend: Command Palette (Ctrl+K)

```tsx
// components/search/command-palette.tsx
import { Command } from "cmdk";
import { useEffect, useState, useCallback } from "react";
import { Dialog, DialogContent } from "@/components/ui/dialog";

export function CommandPalette() {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");

  // Глобальный хоткей Ctrl+K
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((o) => !o);
      }
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, []);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="p-0 max-w-lg">
        <Command shouldFilter={false}>
          <Command.Input
            value={query}
            onValueChange={setQuery}
            placeholder="Поиск задач, кейсов, пользователей..."
          />
          <Command.List>
            {/* Группы результатов по типу */}
            <Command.Group heading="Задачи">
              {/* SearchResult items */}
            </Command.Group>
            <Command.Group heading="Тест-кейсы">
              {/* SearchResult items */}
            </Command.Group>
          </Command.List>
        </Command>
      </DialogContent>
    </Dialog>
  );
}
```

### Frontend: Notification Store

```typescript
// stores/notification-store.ts
import { create } from "zustand";

interface Notification {
  id: string;
  notification_type: string;
  title: string;
  body?: string;
  entity_type?: string;
  entity_id?: string;
  project_id?: string;
  is_read: boolean;
  is_urgent: boolean;
  created_at: string;
}

interface NotificationState {
  notifications: Notification[];
  unreadCount: number;
  setNotifications: (notifs: Notification[]) => void;
  addNotification: (notif: Notification) => void;
  markRead: (id: string) => void;
  markAllRead: () => void;
}

export const useNotificationStore = create<NotificationState>((set) => ({
  notifications: [],
  unreadCount: 0,
  setNotifications: (notifications) =>
    set({
      notifications,
      unreadCount: notifications.filter((n) => !n.is_read).length,
    }),
  addNotification: (notif) =>
    set((s) => ({
      notifications: [notif, ...s.notifications],
      unreadCount: s.unreadCount + (notif.is_read ? 0 : 1),
    })),
  markRead: (id) =>
    set((s) => ({
      notifications: s.notifications.map((n) =>
        n.id === id ? { ...n, is_read: true } : n
      ),
      unreadCount: Math.max(0, s.unreadCount - 1),
    })),
  markAllRead: () =>
    set((s) => ({
      notifications: s.notifications.map((n) => ({ ...n, is_read: true })),
      unreadCount: 0,
    })),
}));
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Tremor NPM package (@tremor/react) | Tremor Raw (copy-paste components) | 2025 | No NPM dependency; components are plain Tailwind + Recharts; more customizable |
| Separate Elasticsearch for search | PG tsvector + GIN (for 50 users) | Always valid | No extra infrastructure; PG handles FTS natively; Elastic overkill at this scale |
| Polling for notifications | WebSocket push via NATS bridge | Project pattern | Already established in the project; consistent with other real-time features |

**Deprecated/outdated:**
- `@tremor/react` NPM package: still works but Tremor team now recommends Raw (copy-paste) approach. Last NPM publish was 1+ year ago. Use Tremor Raw from tremor.so.
- MinIO references: replaced by SeaweedFS. All file operations use aws-sdk-go-v2 against SeaweedFS endpoint.

## Open Questions

1. **Notification batching window duration**
   - What we know: 2-second window prevents notification storms on bulk operations
   - What's unclear: Optimal window for different trigger types (bulk assign vs single mention)
   - Recommendation: Start with 2 seconds for all types; tune per-type if users complain about delay

2. **Dashboard widget presets for project templates**
   - What we know: DASH-02 requires Dev/QA/Manager templates
   - What's unclear: Exact widget composition for each template
   - Recommendation: Define 3 JSON presets in code; iterate based on usage

3. **GDPR full delete scope for MongoDB messages**
   - What we know: AUDT-04 requires full data deletion option
   - What's unclear: Whether to delete messages or anonymize (replace sender with "Deleted User")
   - Recommendation: Anonymize messages (preserve thread context), delete user record and all PG data

4. **Soft-delete field on Issue schema**
   - What we know: AUDT-03 requires 30-day trash for issues; Issue schema currently has no deleted_at field
   - What's unclear: Whether to add deleted_at to Issue or use a separate trash table
   - Recommendation: Add `deleted_at` nullable field to Issue (same pattern as Project), filter in all queries with `Where(issue.DeletedAtIsNil())`

## Sources

### Primary (HIGH confidence)
- PostgreSQL 18 documentation on tsvector and GIN indexes: https://www.postgresql.org/docs/current/textsearch-tables.html
- MongoDB Go driver text search documentation: https://www.mongodb.com/docs/drivers/go/current/crud/query/text/
- NATS JetStream documentation (durable consumers): https://docs.nats.io/nats-concepts/jetstream
- Existing project codebase: backend/internal/ws/bridge.go, events.go files, outbox pattern

### Secondary (MEDIUM confidence)
- telego GitHub repository and documentation: https://github.com/mymmrac/telego
- Tremor Raw (copy-paste components): https://www.tremor.so/
- react-grid-layout documentation: https://github.com/react-grid-layout/react-grid-layout
- cmdk command palette: https://cmdk.paco.me

### Tertiary (LOW confidence)
- Telegram Bot API rate limits (30 msg/sec to different chats): community-reported, official docs less specific
- react-grid-layout TypeScript issues: community reports, may improve in future versions

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - all libraries verified via NPM/GitHub, telego actively maintained (March 2026 release)
- Architecture: HIGH - follows established project patterns (NATS bridge, Ent schemas, handler/service/repository layers)
- Pitfalls: HIGH - tsvector language config, notification storms, Telegram rate limits are well-documented issues
- Search: HIGH - PG tsvector + MongoDB $text for 50 users is a proven pattern; no separate search engine needed
- Dashboards: MEDIUM - react-grid-layout TS types and Tremor Raw integration not personally verified but widely used

**Research date:** 2026-03-21
**Valid until:** 2026-04-21 (stable technologies, 30-day window)
