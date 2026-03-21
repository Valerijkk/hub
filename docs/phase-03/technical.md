# Мессенджер -- Техническая документация

## Архитектура модуля

Модуль мессенджера реализован по паттерну **Handler -> Service -> Repository** с двойным хранилищем (dual-storage): Ent/PostgreSQL для метаданных каналов и участников, MongoDB для сообщений.

```
HTTP-запрос / WebSocket
     |
     v
+------------+    +-------------------+    +---------------------+
|  Handler   | -> | MessengerService  | -> | Repository          |
|  (Fiber)   |    | (бизнес-логика)   |    | (интерфейс)         |
+------------+    +-------------------+    +---------------------+
                       |    |                    |            |
                       |    |                    v            v
                       |    |           +---------------+  +------------------+
                       |    |           | EntRepository |  | MongoRepository  |
                       |    |           | (Ent ORM/PG)  |  | (MongoDB)        |
                       |    |           +---------------+  +------------------+
                       |    |                    |            |
                       |    |                    v            v
                       |    |           +------------+  +----------+
                       |    |           | PostgreSQL |  | MongoDB  |
                       |    |           +------------+  +----------+
                       |    |
                       v    v
              +---------+  +-------+
              | NATS    |  | Redis |
              | (events)|  | (cache)|
              +---------+  +-------+
                   |
                   v
            +----------+
            | WS Bridge|
            +----------+
                   |
                   v
            +----------+
            | Клиенты  |
            | (WS)     |
            +----------+
```

### Обоснование dual-storage

| Хранилище | Назначение | Обоснование |
|-----------|------------|-------------|
| **PostgreSQL (Ent)** | Каналы, участники, закреплённые сообщения, закладки | Реляционные связи, ACID-транзакции, Ent-кодогенерация |
| **MongoDB** | Сообщения, реакции | Документная модель (вложенные реакции, вложения), text index для поиска, высокая запись |

---

## Структура файлов

```
backend/internal/messenger/
  handler.go              # HTTP-обработчики, RegisterRoutes
  service.go              # Бизнес-логика (MessengerService, 19+ методов)
  repository.go           # Интерфейс Repository (45+ методов)
  ent_repository.go       # Ent/PG реализация (каналы, участники, pins, bookmarks)
  mongo_repository.go     # MongoDB реализация (сообщения, реакции, поиск)
  combined_repository.go  # CombinedRepository (делегирует в Ent + MongoDB)
  dto.go                  # Request/Response DTOs, MessageDocument
  events.go               # NATS subject constants, event payloads
  presence.go             # PresenceService (Redis sorted set)
  unread.go               # UnreadService (Redis hash)
  auto_channels.go        # AutoChannelService (NATS subscription)
  service_test.go         # Unit-тесты бизнес-логики (mock repository)
  auto_channels_test.go   # Unit-тесты авто-каналов
  presence_test.go        # Unit-тесты presence (miniredis)
  unread_test.go          # Unit-тесты unread (miniredis)
```

---

## Ent-схемы (PostgreSQL)

### Channel

Сущность канала мессенджера.

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | UUID | Первичный ключ, автогенерация, immutable |
| `channel_type` | Enum(`dm`, `group`, `public`, `private`) | Тип канала |
| `name` | String(100) | Название канала (пусто для DM), optional |
| `description` | String(500) | Описание канала, optional |
| `is_readonly` | Bool | Флаг read-only (default: false) |
| `project_id` | UUID | FK на проект (nil для DM и глобальных), optional, nillable |
| `created_by` | UUID | FK на создателя |
| `dm_key` | String | Ключ дедупликации DM (unique), optional, nillable |
| `created_at` | Time | Дата создания, immutable |
| `updated_at` | Time | Дата обновления, auto-update |

**Edges:**
- `members` -> ChannelMember (one-to-many)
- `project` <- Project (many-to-one, через `project_id`)
- `creator` <- User (many-to-one, через `created_by`)
- `pinned_messages` -> PinnedMessage (one-to-many)

**Индексы:**
- `channel_type` -- фильтрация по типу
- `project_id` -- каналы проекта

**DM дедупликация:**
Поле `dm_key` формируется из отсортированных UUID двух пользователей, соединённых дефисом: `sort([userA, userB])[0] + "-" + sort([userA, userB])[1]`. Unique constraint на `dm_key` предотвращает дублирование DM-каналов.

### ChannelMember

Участник канала.

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | UUID | Первичный ключ |
| `channel_id` | UUID | FK на канал |
| `user_id` | UUID | FK на пользователя |
| `role` | Enum(`owner`, `admin`, `member`) | Роль в канале (default: member) |
| `joined_at` | Time | Время вступления |

**Индексы:**
- Unique composite: `(channel_id, user_id)` -- один пользователь = одна запись в канале

### Bookmark

Персональная закладка пользователя на сообщение.

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | UUID | Первичный ключ |
| `user_id` | UUID | FK на пользователя |
| `message_id` | String | ID сообщения (MongoDB ObjectID) |
| `channel_id` | UUID | FK на канал |
| `created_at` | Time | Дата создания |

### PinnedMessage

Закреплённое сообщение в канале.

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | UUID | Первичный ключ |
| `channel_id` | UUID | FK на канал |
| `message_id` | String | ID сообщения (MongoDB ObjectID) |
| `pinned_by` | UUID | FK на закрепившего пользователя |
| `pinned_at` | Time | Дата закрепления |

---

## MongoDB

### Коллекция `messages`

Структура документа `MessageDocument`:

```json
{
  "_id": "uuid-string",
  "channel_id": "uuid (Binary)",
  "sender_id": "uuid (Binary)",
  "content_json": { "...Plate JSON..." },
  "content_text": "plain text для поиска",
  "thread_id": "parent-message-id (опционально)",
  "reply_count": 0,
  "forwarded_from_id": "original-message-id (опционально)",
  "voice_url": "https://s3.../voice.webm (опционально)",
  "voice_duration_ms": 15000,
  "attachments": [
    {
      "file_id": "uuid",
      "file_name": "screenshot.png",
      "file_size": 245000,
      "mime_type": "image/png",
      "url": "https://s3.../file"
    }
  ],
  "reactions": [
    {
      "emoji": "thumbs_up",
      "user_ids": ["user-uuid-1", "user-uuid-2"]
    }
  ],
  "mentions": ["user-uuid-1"],
  "mention_groups": ["@all"],
  "edited_at": null,
  "deleted_at": null,
  "created_at": "2026-03-21T01:00:00Z"
}
```

### Индексы

| Индекс | Поля | Назначение |
|--------|------|------------|
| Compound | `(channel_id: 1, created_at: -1)` | Постраничная загрузка сообщений канала |
| Thread | `(channel_id: 1, thread_id: 1, created_at: -1)` | Загрузка ответов в тредах |
| Text | `content_text: "text"` | Полнотекстовый поиск |

### Cursor pagination

Пагинация реализована через параметр `before` (RFC3339 timestamp):

```
GET /api/messenger/channels/:id/messages?limit=50&before=2026-03-21T01:00:00Z
```

MongoDB-фильтр: `{ channel_id: X, created_at: { $lt: before }, thread_id: { $exists: false }, deleted_at: null }`. Сортировка: `created_at DESC`. Лимит: 1-100 (default 50).

---

## NATS

### Subject conventions

Формат: `messenger.channel.{channelID}.{event}`

| Subject | Событие | Payload |
|---------|---------|---------|
| `messenger.channel.{id}.message_sent` | Новое сообщение | `MessageSentEvent` |
| `messenger.channel.{id}.message_edited` | Редактирование | `MessageEditedEvent` |
| `messenger.channel.{id}.message_deleted` | Удаление | `MessageDeletedEvent` |
| `messenger.channel.{id}.reaction_toggled` | Реакция (add/remove) | `ReactionToggledEvent` |
| `messenger.channel.{id}.typing` | Набор текста | Typing payload |
| `messenger.channel.{id}.member_joined` | Участник вступил | `{channel_id, user_id}` |
| `messenger.channel.{id}.member_left` | Участник вышел | `{channel_id, user_id}` |
| `messenger.presence.{id}.changed` | Изменение статуса | Presence payload |

### Bridge (NATS -> WebSocket)

Расширение WS Bridge в `backend/internal/ws/bridge.go`:

```go
// Подписка: messenger.channel.>
// Routing: messenger.channel.{channelID}.{event} -> WS room "channel:{channelID}"
b.nats.Subscribe("messenger.channel.>", func(msg *nats.Msg) {
    parts := strings.SplitN(msg.Subject, ".", 4)
    channelID := parts[2]
    room := "channel:" + channelID
    event := parts[3]
    b.hub.BroadcastToRoom(room, ServerMessage{Type: TypeEvent, Event: event, Room: room, Payload: payload})
})
```

Используется **Core NATS** (не JetStream), так как события мессенджера эфемерные -- потеря при перезапуске допустима.

### AutoChannelService

Подписка на `project.created` для автоматического создания каналов:

1. Outbox worker публикует `project.created` в Core NATS
2. AutoChannelService получает событие
3. Создаёт публичный канал `#project-{key}` (в нижнем регистре)
4. Загружает всех участников проекта из EntRepository
5. Добавляет участников через `AddMembersInternal` (без проверки прав)

---

## Redis

### Presence (онлайн-статус)

**Структура данных:** Redis sorted set

| Ключ | Тип | Описание |
|------|-----|----------|
| `presence:online` | Sorted Set | member = userID, score = Unix timestamp heartbeat |
| `presence:status:{userID}` | String | Кастомный текст статуса, TTL 24ч |

**Алгоритм определения статуса:**
1. Получить score (timestamp последнего heartbeat) из sorted set
2. Если нет записи -> `offline`
3. Если `elapsed < 5 мин` -> `online`
4. Если `elapsed >= 5 мин` -> `away`

**Операции:**
- `Heartbeat(userID)` -> `ZADD presence:online timestamp userID`
- `GetStatus(userID)` -> `ZSCORE presence:online userID` + вычисление elapsed
- `GetBulkStatus(userIDs)` -> Redis pipeline с множественными ZSCORE
- `CleanupStale()` -> `ZREMRANGEBYSCORE presence:online -inf (now-90s)` -- удаляет записи старше 90 секунд
- `SetCustomStatus(userID, text)` -> `SET presence:status:{userID} text EX 86400` (или DEL если пусто)

**Cleanup loop:** Горутина с ticker 60 секунд, очищает записи старше `presenceStaleThreshold` (90 секунд).

### Unread (непрочитанные)

**Структура данных:** Redis hash

| Ключ | Тип | Описание |
|------|-----|----------|
| `unread:{userID}` | Hash | field = channelID, value = count |

**Операции:**
- `Increment(channelID, memberIDs, senderID)` -> Pipeline `HINCRBY unread:{memberID} channelID 1` для всех участников (кроме отправителя)
- `MarkRead(userID, channelID)` -> `HDEL unread:{userID} channelID`
- `GetAll(userID)` -> `HGETALL unread:{userID}`

**Приблизительные счётчики:** Как Slack, счётчики приблизительные -- нет точного отслеживания "до какого сообщения прочитано". При открытии канала счётчик сбрасывается в 0.

---

## REST API

Все эндпоинты под группой `/api/messenger`, требуют аутентификацию (auth middleware).

### Каналы

| Метод | Путь | Описание | Request Body | Response |
|-------|------|----------|-------------|----------|
| POST | `/channels` | Создать канал | `CreateChannelRequest` | 201 `ChannelResponse` |
| GET | `/channels` | Мои каналы | -- | 200 `[]ChannelResponse` |
| GET | `/channels/:channelId` | Канал по ID | -- | 200 `ChannelResponse` |
| PUT | `/channels/:channelId` | Обновить канал | `CreateChannelRequest` | 200 `ChannelResponse` |
| DELETE | `/channels/:channelId` | Удалить канал | -- | 204 |
| POST | `/channels/:channelId/join` | Присоединиться | -- | 200 |
| POST | `/channels/:channelId/leave` | Покинуть | -- | 200 |
| POST | `/channels/:channelId/members` | Добавить участников | `AddMembersRequest` | 200 |
| GET | `/channels/:channelId/members` | Список участников | -- | 200 `[]ChannelMemberResponse` |

### Сообщения

| Метод | Путь | Описание | Request Body | Response |
|-------|------|----------|-------------|----------|
| POST | `/channels/:channelId/messages` | Отправить | `SendMessageRequest` | 201 `MessageResponse` |
| GET | `/channels/:channelId/messages` | Список (cursor) | Query: `before`, `limit` | 200 `[]MessageResponse` |
| PUT | `/channels/:channelId/messages/:messageId` | Редактировать | `EditMessageRequest` | 200 `MessageResponse` |
| DELETE | `/channels/:channelId/messages/:messageId` | Удалить (soft) | -- | 204 |
| POST | `/channels/:channelId/messages/:messageId/reactions` | Toggle реакция | `ToggleReactionRequest` | 200 `{added: bool}` |
| POST | `/channels/:channelId/messages/:messageId/pin` | Toggle pin | -- | 200 `{pinned: bool}` |
| POST | `/channels/:channelId/messages/:messageId/forward` | Переслать | `ForwardMessageRequest` | 201 `MessageResponse` |
| GET | `/channels/:channelId/messages/:messageId/replies` | Ответы треда | Query: `before`, `limit` | 200 `[]MessageResponse` |

### Прочее

| Метод | Путь | Описание | Request Body | Response |
|-------|------|----------|-------------|----------|
| GET | `/presence` | Статусы присутствия | Query: `user_ids` | 200 `map[string]string` |
| PUT | `/presence/status` | Custom статус | `{text: string}` | 200 |
| POST | `/presence/heartbeat` | Heartbeat | -- | 200 |
| GET | `/unread` | Непрочитанные | -- | 200 `map[string]int64` |
| POST | `/channels/:channelId/read` | Отметить прочитанным | -- | 200 |
| GET | `/channels/:channelId/pins` | Закреплённые | -- | 200 `[]PinnedMessageResponse` |
| GET | `/search` | Поиск | Query: `query`, `channel_id`, `limit` | 200 `[]MessageResponse` |
| GET | `/bookmarks` | Закладки | -- | 200 `[]BookmarkResponse` |
| POST | `/bookmarks` | Добавить закладку | `AddBookmarkRequest` | 201 |
| DELETE | `/bookmarks/:bookmarkId` | Удалить закладку | -- | 204 |
| POST | `/dm` | Создать/получить DM | `CreateDMRequest` | 200 `ChannelResponse` |

---

## Frontend

### Компоненты

| Компонент | Файл | Описание |
|-----------|------|----------|
| `ChannelList` | `channel-list.tsx` | Список каналов с группировкой по типу (DM, группы, каналы), unread badges, presence dots для DM, кнопка создания канала |
| `ChatView` | `chat-view.tsx` | Основной чат с виртуализацией (@tanstack/react-virtual), infinite scroll вверх, auto-scroll вниз, markRead при открытии |
| `MessageItem` | `message-item.tsx` | Отображение сообщения: Plate JSON рендер (RenderNode), голосовой плеер, реакции, hover-действия (ответить, переслать, pin, bookmark) |
| `MessageComposer` | `message-composer.tsx` | Plate editor с Enter для отправки, Shift+Enter для новой строки, кнопка голосового сообщения |
| `ThreadPanel` | `thread-panel.tsx` | Боковая панель треда: список ответов, MessageComposer для ответа |
| `VoiceRecorder` | `voice-recorder.tsx` | Запись через MediaRecorder API, формат Opus/WebM, лимит 5 минут, upload в S3 |
| `ReactionPicker` | `reaction-picker.tsx` | Popover с сеткой популярных emoji |
| `PresenceIndicator` | `presence-indicator.tsx` | Цветной кружок: зелёный (online), жёлтый (away), серый (offline) |
| `SearchDialog` | `search-dialog.tsx` | Dialog с полнотекстовым поиском, переключатель scope (канал/глобально) |
| `PinnedMessages` | `pinned-messages.tsx` | Панель закреплённых сообщений канала |
| `BookmarksPanel` | `bookmarks-panel.tsx` | Dialog с персональными закладками |
| `ForwardDialog` | `forward-dialog.tsx` | Dialog с выбором целевого канала для пересылки |

### Hooks

| Hook | Файл | Описание |
|------|------|----------|
| `useChannels` | `use-channels.ts` | TanStack Query для списка каналов + WS подписка на `member_joined`/`member_left` события |
| `useMessages` | `use-messages.ts` | Infinite Query с cursor pagination + WS подписка на `message_sent`/`edited`/`deleted`/`reaction_toggled` |
| `useThreadMessages` | `use-messages.ts` | Infinite Query для ответов в треде |
| `usePresence` | `use-presence.ts` | Polling каждые 30 секунд + heartbeat каждые 25 секунд |
| `useUnread` | `use-unread.ts` | Получение счётчиков непрочитанных |
| `useMarkRead` | `use-unread.ts` | Оптимистичный сброс unread count + API вызов + invalidation |

**Паттерн WS event handlers:**
```typescript
// useRef для stable handler reference (без stale closures)
const handlerRef = useRef<Function>(null);
handlerRef.current = useCallback(() => { ... }, [deps]);

useEffect(() => {
  const handler = (data) => handlerRef.current?.(data);
  wsStore.addListener("event_name", handler);
  return () => wsStore.removeListener("event_name", handler);
}, [channelId]);
```

### Store

**messenger-store** (Zustand):

```typescript
interface MessengerStore {
  activeChannelId: string | null;
  channels: Channel[];
  unreadCounts: Record<string, number>;
  threadMessageId: string | null;

  setActiveChannel(id: string): void;
  setChannels(channels: Channel[]): void;
  setUnreadCounts(counts: Record<string, number>): void;
  resetUnread(channelId: string): void;
  openThread(messageId: string): void;
  closeThread(): void;
}
```

### Виртуализация

ChatView использует `@tanstack/react-virtual` для эффективного рендера больших списков сообщений:

- `useVirtualizer` с `count = messages.length`
- Reverse scrolling: новые сообщения внизу, загрузка старых при скролле вверх
- `getScrollElement` привязан к ScrollArea контейнеру
- Auto-scroll при получении нового сообщения (если пользователь в нижней части)

### Маршруты

| Маршрут | Файл | Описание |
|---------|------|----------|
| `/messenger` | `messenger.tsx` | Layout route: sidebar (ChannelList) + Outlet |
| `/messenger/` | `messenger/index.tsx` | Страница "Выберите канал" |
| `/messenger/$channelId` | `messenger/$channelId.tsx` | Страница канала с ChatView |

---

## Решения и компромиссы

| Решение | Обоснование |
|---------|-------------|
| Approximate unread counts (как Slack) | Точные read receipts требуют per-message tracking, избыточно для 50 пользователей |
| No delivery receipts | Осознанное ограничение -- Slack тоже не показывает "доставлено" |
| Plain text extraction из Plate JSON | content_text для MongoDB text index, без отдельного поискового движка |
| Soft delete сообщений (deleted_at) | Возможность восстановления, audit trail, не ломает thread_id ссылки |
| CombinedRepository | Единый интерфейс для двух хранилищ, тестируемость через mock |
| Core NATS (не JetStream) | Ephemeral events: потеря при перезапуске допустима для typing/presence |
| miniredis для тестов | Быстрые in-process тесты без Docker-зависимости |
| Inline Plate JSON renderer | PlateStatic тянет тяжёлые deps, рекурсивный RenderNode проще |
| useRef для WS handlers | Prevents stale closures при re-render в React 19 |
| MediaRecorder API | Нативное решение браузера для записи аудио, без сторонних библиотек |

---
*Phase: 03-messenger*
*Последнее обновление: 2026-03-21*
