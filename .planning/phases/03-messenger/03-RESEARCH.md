# Phase 3: Messenger - Research

**Researched:** 2026-03-21
**Domain:** Real-time messenger (MongoDB messages, NATS pub/sub, Redis presence, Ent channel metadata)
**Confidence:** HIGH

## Summary

Phase 3 implements the messenger module for Hub -- DMs, group chats, public/private channels, auto-project channels, read-only channels, threaded replies, emoji reactions, rich text (Plate JSON), message CRUD with permission-based deletion, forwarding/pinning/bookmarks, voice messages, message search, and user presence/status. The architecture splits responsibility: PostgreSQL (Ent) stores channel metadata (members, permissions, settings), MongoDB stores messages (append-heavy, flexible schema for reactions/threads/attachments), Redis stores presence and unread counts, and NATS core pub/sub delivers real-time events through the existing WebSocket bridge.

The existing infrastructure from Phase 1+2 provides a solid foundation: WebSocket hub with room-based subscriptions and NATS bridge (hub.go, bridge.go, protocol.go), MongoDB client (database/mongodb.go), Redis client, NATS client with JetStream, outbox pattern for PG+Mongo consistency, auth middleware, and RBAC engine with channel-specific permissions already defined (PermChannelCreate, PermChannelManage, PermMessageEdit, PermMessageDeleteAny). The messenger module follows the same handler/service/repository pattern established by the tracker module.

The primary technical challenges are: (1) MongoDB document schema design for messages that efficiently supports threads, reactions, and cursor-based pagination; (2) bridging NATS events to WebSocket rooms for real-time message delivery using the existing bridge infrastructure; (3) Redis-based presence tracking (online/offline/away) with idle detection; (4) unread count tracking per user per channel in Redis without consistency drift; (5) MongoDB text index for message search. None of these are novel problems -- they have well-established patterns in the Go/MongoDB ecosystem.

**Primary recommendation:** Follow the existing modular pattern (handler.go, service.go, repository.go, dto.go, events.go) under `backend/internal/messenger/`. Channel metadata in Ent (new Channel, ChannelMember schemas), messages in MongoDB with compound index on `(channel_id, created_at)`. Extend the existing NATS bridge to handle `messenger.channel.>` subjects. Use Redis sorted sets for presence and hash maps for unread counts.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| MSG-01 | DMs (1-on-1) and group chats | Channel types: "dm", "group" in Ent Channel schema; DM channels are unique per user pair |
| MSG-02 | Public channels (anyone joins) and private channels (invite-only) | Channel type enum: "public", "private"; join API for public, invite API for private |
| MSG-03 | Auto-created project channels (#project-hub) and topic channels | Triggered on project creation via outbox event; topic channels created via API |
| MSG-04 | Read-only channels (bot posts, users read) | Channel field `is_readonly` + permission check on message send |
| MSG-05 | Threaded replies (reply opens a thread) | MongoDB `thread_id` field on message document; thread root is a regular message |
| MSG-06 | Emoji reactions on messages | MongoDB reactions array `[{emoji, user_id, created_at}]` embedded in message document |
| MSG-07 | Rich text formatting (bold, italic, code, lists, links) | Plate JSON stored as `content_json` + plain text `content_text` for search |
| MSG-08 | Edit/delete own messages; Owner/Admin can delete any | REST API with permission checks; edited messages get `edited_at` timestamp; deleted messages get `deleted_at` (soft delete) |
| MSG-09 | Forwarding, pinning, bookmarks (saved messages) | Forward = new message with `forwarded_from` ref; pin = channel-level array in Ent; bookmarks = user-level collection in Ent |
| MSG-10 | Voice messages (audio recording) | Audio file uploaded to SeaweedFS via existing presigned URL flow; message has `voice_url` + `voice_duration_ms` fields |
| MSG-12 | Global + per-channel message search | MongoDB text index on `content_text`; search API with channel filter and cursor pagination |
| MSG-13 | User status: online/offline/away (auto) + custom text | Redis sorted set for presence with heartbeat TTL; custom status text in Ent User or Redis |
</phase_requirements>

## Standard Stack

### Core (already installed)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| go.mongodb.org/mongo-driver/v2 | 2.5.0 | Message storage, search | Already in go.mod, v2 is current line |
| github.com/redis/go-redis/v9 | 9.x | Presence, unread counts | Already in go.mod |
| github.com/nats-io/nats.go | 1.49.x | Real-time message delivery | Already in go.mod, core NATS for ephemeral events |
| entgo.io/ent | 0.14.x | Channel metadata schemas | Already in go.mod, graph traversal for channel membership |
| github.com/gofiber/fiber/v3 | 3.x | REST API handlers | Already in go.mod |

### Supporting (new for Phase 3)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| @tanstack/react-virtual | 3.x | Chat message virtualization | Already installed from Phase 2; use for infinite scroll in chat |
| @udecode/plate | 49.x | Rich text editor for messages | Already installed from Phase 2; reuse Plate config for message composer |

### No New Dependencies Required

All backend and frontend dependencies for the messenger are already installed from Phase 1 and Phase 2. No new `go get` or `npm install` commands needed.

## Architecture Patterns

### Recommended Project Structure

```
backend/internal/messenger/
  handler.go         # Fiber HTTP handlers for channels, messages, presence
  service.go         # Business logic orchestration
  repository.go      # Repository interface
  ent_repository.go  # Ent implementation for channel metadata (PG)
  mongo_repository.go # MongoDB implementation for messages
  presence.go        # Redis presence service (online/offline/away)
  unread.go          # Redis unread count tracking
  dto.go             # Request/response types
  events.go          # NATS event definitions and subject constants
  search.go          # MongoDB text search queries

backend/ent/schema/
  channel.go         # NEW: Channel entity (PG)
  channelmember.go   # NEW: Channel membership (PG)
  bookmark.go        # NEW: User bookmarks / saved messages (PG)
  pinnedmessage.go   # NEW: Pinned messages per channel (PG)

frontend/src/modules/messenger/
  components/        # ChannelList, ChatView, MessageItem, ThreadPanel, Composer
  hooks/             # useMessages, useChannels, usePresence, useUnread
  stores/            # messengerStore (Zustand)
  api/               # REST API client functions
```

### Pattern 1: Dual-Storage (Ent + MongoDB)

**What:** Channel metadata (members, settings, permissions) lives in PostgreSQL via Ent. Messages live in MongoDB. The two are linked by `channel_id` (UUID string).

**When to use:** All messenger operations. Channel CRUD uses Ent. Message CRUD uses MongoDB.

**Example (Channel schema in Ent):**
```go
// backend/ent/schema/channel.go
type Channel struct {
    ent.Schema
}

func (Channel) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
        field.Enum("channel_type").Values("dm", "group", "public", "private"),
        field.String("name").Optional().MaxLen(100),
        field.String("description").Optional().MaxLen(500),
        field.Bool("is_readonly").Default(false),
        field.UUID("project_id", uuid.UUID{}).Optional().Nillable(),
        field.UUID("created_by", uuid.UUID{}),
        field.Time("created_at").Default(time.Now).Immutable(),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (Channel) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("members", ChannelMember.Type),
        edge.From("project", Project.Type).Ref("channels").Field("project_id").Unique(),
        edge.From("creator", User.Type).Ref("created_channels").Field("created_by").Unique().Required(),
    }
}
```

**Example (Message document in MongoDB):**
```go
// MongoDB collection: messages
type MessageDocument struct {
    ID              string                 `bson:"_id"`           // UUID string
    ChannelID       string                 `bson:"channel_id"`   // matches Ent Channel.ID
    SenderID        string                 `bson:"sender_id"`    // matches Ent User.ID
    ContentJSON     map[string]interface{} `bson:"content_json"` // Plate JSON
    ContentText     string                 `bson:"content_text"` // plain text for search
    ThreadID        *string                `bson:"thread_id,omitempty"`  // parent message ID
    ReplyCount      int                    `bson:"reply_count"`
    ForwardedFromID *string                `bson:"forwarded_from_id,omitempty"`
    VoiceURL        *string                `bson:"voice_url,omitempty"`
    VoiceDurationMs *int                   `bson:"voice_duration_ms,omitempty"`
    Attachments     []AttachmentRef        `bson:"attachments,omitempty"` // file IDs from storage
    Reactions       []ReactionEntry        `bson:"reactions,omitempty"`
    Mentions        []string               `bson:"mentions,omitempty"`    // user IDs
    MentionGroups   []string               `bson:"mention_groups,omitempty"` // "all", "online", role names
    EditedAt        *time.Time             `bson:"edited_at,omitempty"`
    DeletedAt       *time.Time             `bson:"deleted_at,omitempty"` // soft delete
    CreatedAt       time.Time              `bson:"created_at"`
}

type ReactionEntry struct {
    Emoji     string    `bson:"emoji"`
    UserID    string    `bson:"user_id"`
    CreatedAt time.Time `bson:"created_at"`
}

type AttachmentRef struct {
    FileID      string `bson:"file_id"`
    FileName    string `bson:"file_name"`
    ContentType string `bson:"content_type"`
    Size        int64  `bson:"size"`
}
```

### Pattern 2: NATS Bridge Extension for Messenger

**What:** Extend the existing `ws/bridge.go` to subscribe to `messenger.channel.>` NATS subjects and route events to `channel:{channelID}` WebSocket rooms. Messages published by the messenger service to NATS are automatically fanned out to all connected clients in that channel.

**When to use:** All real-time messenger events -- new messages, edits, deletes, reactions, typing indicators.

**Example (extending bridge.go):**
```go
// Добавить в Bridge.Start()
messengerSub, err := b.nats.Subscribe("messenger.channel.>", func(msg *nats.Msg) {
    b.handleMessengerEvent(msg)
})

// messenger.channel.{channelID}.{event}
// -> room "channel:{channelID}"
func (b *Bridge) handleMessengerEvent(msg *nats.Msg) {
    parts := strings.SplitN(msg.Subject, ".", 4)
    if len(parts) < 3 {
        return
    }
    channelID := parts[2]
    room := "channel:" + channelID
    event := parts[len(parts)-1]

    var payload any
    if len(msg.Data) > 0 {
        _ = json.Unmarshal(msg.Data, &payload)
    }

    b.hub.BroadcastToRoom(room, ServerMessage{
        Type:    TypeEvent,
        Event:   event,
        Room:    room,
        Payload: payload,
    })
}
```

**NATS subject conventions for messenger:**
```
messenger.channel.{channelID}.message_sent     # Новое сообщение
messenger.channel.{channelID}.message_edited   # Редактирование
messenger.channel.{channelID}.message_deleted  # Удаление
messenger.channel.{channelID}.reaction_toggled # Реакция
messenger.channel.{channelID}.typing           # Typing indicator (core NATS only)
messenger.channel.{channelID}.member_joined    # Участник присоединился
messenger.channel.{channelID}.member_left      # Участник вышел
messenger.presence.{userID}.changed            # Статус изменился
```

### Pattern 3: Redis Presence with Heartbeat

**What:** Track online/offline/away status using Redis sorted sets. Client sends heartbeat every 30s via WebSocket ping. Server updates Redis TTL. Idle > 5 minutes = "away". No heartbeat > 90s = "offline".

**Example:**
```go
// Ключ: presence:online — sorted set, score = timestamp последнего heartbeat
// Ключ: presence:status:{userID} — hash с custom_text, explicit_status

func (p *PresenceService) Heartbeat(ctx context.Context, userID string) error {
    now := float64(time.Now().Unix())
    return p.rdb.ZAdd(ctx, "presence:online", redis.Z{
        Score:  now,
        Member: userID,
    }).Err()
}

func (p *PresenceService) GetStatus(ctx context.Context, userID string) (string, error) {
    score, err := p.rdb.ZScore(ctx, "presence:online", userID).Result()
    if err == redis.Nil {
        return "offline", nil
    }
    lastSeen := time.Unix(int64(score), 0)
    if time.Since(lastSeen) > 5*time.Minute {
        return "away", nil
    }
    return "online", nil
}

// Cron: каждые 60 секунд удаляем записи старше 90 секунд
func (p *PresenceService) CleanupStale(ctx context.Context) error {
    cutoff := float64(time.Now().Add(-90 * time.Second).Unix())
    return p.rdb.ZRemRangeByScore(ctx, "presence:online", "-inf", fmt.Sprintf("%f", cutoff)).Err()
}
```

### Pattern 4: Redis Unread Counts

**What:** Track unread message count per user per channel using Redis hashes. Increment on new message, reset on channel read.

**Example:**
```go
// Ключ: unread:{userID} — hash, field = channelID, value = count

func (u *UnreadService) Increment(ctx context.Context, channelID string, memberIDs []string, senderID string) error {
    pipe := u.rdb.Pipeline()
    for _, memberID := range memberIDs {
        if memberID == senderID {
            continue // Не считаем собственные сообщения
        }
        pipe.HIncrBy(ctx, "unread:"+memberID, channelID, 1)
    }
    _, err := pipe.Exec(ctx)
    return err
}

func (u *UnreadService) MarkRead(ctx context.Context, userID, channelID string) error {
    return u.rdb.HDel(ctx, "unread:"+userID, channelID).Err()
}

func (u *UnreadService) GetAll(ctx context.Context, userID string) (map[string]int64, error) {
    return u.rdb.HGetAll(ctx, "unread:"+userID).Result()
    // Возвращает map[channelID]count
}
```

### Pattern 5: Cursor-Based Pagination for Chat

**What:** Chat messages use cursor-based pagination with `created_at` as cursor. Client loads latest N messages, scrolls up to load older. MongoDB query uses `$lt` on `created_at` with compound index.

**Example:**
```go
func (r *MongoRepository) ListMessages(ctx context.Context, channelID string, before *time.Time, limit int) ([]MessageDocument, error) {
    filter := bson.M{
        "channel_id": channelID,
        "deleted_at": nil, // Не показываем удалённые
    }
    if before != nil {
        filter["created_at"] = bson.M{"$lt": *before}
    }

    opts := options.Find().
        SetSort(bson.D{{Key: "created_at", Value: -1}}).
        SetLimit(int64(limit))

    cursor, err := r.coll.Find(ctx, filter, opts)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var messages []MessageDocument
    if err := cursor.All(ctx, &messages); err != nil {
        return nil, err
    }
    return messages, nil
}
```

### Anti-Patterns to Avoid

- **Storing messages in PostgreSQL:** Append-heavy write patterns with flexible schema (reactions, attachments) are a poor fit for relational tables. Use MongoDB.
- **Creating JetStream consumers per channel:** Use core NATS pub/sub for chat events. JetStream consumers are for durable background jobs only.
- **Opening a separate WebSocket for messenger:** Use the existing single WebSocket gateway. Subscribe to `channel:{id}` rooms.
- **Querying presence from MongoDB or PostgreSQL:** Presence is ephemeral, high-frequency. Redis sorted sets are the right tool.
- **Storing unread counts in PostgreSQL:** Too many writes per message. Redis hash is O(1) per operation.
- **Fetching all channel members on every message send:** Cache channel membership in Redis or service-level cache. Refresh on join/leave events.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| File uploads in messages | Custom file proxy | Existing SeaweedFS presigned URL flow (storage module) | Already built, battle-tested, direct browser upload |
| Rich text editor | Custom contenteditable | Plate (already configured in Phase 2) | @mentions, formatting, slash commands out of the box |
| Message virtualization | Custom windowing | @tanstack/react-virtual (already installed) | Handles variable-height items, scroll anchoring |
| WebSocket transport | Custom WS per module | Existing WS hub + NATS bridge | Single connection, room subscriptions, horizontal scaling |
| UUID generation | Custom ID scheme | google/uuid (already in go.mod) | Message IDs as UUID v7 (time-ordered) for natural sort |
| Presence tracking | Custom heartbeat protocol | Redis sorted set + WS ping/pong | Industry standard, O(log N) operations |
| Text search | Custom inverted index | MongoDB text index on `content_text` | Built-in, good enough for 10-50 users |

**Key insight:** Phase 3 has the luxury of building on a rich foundation from Phase 1+2. The messenger module itself is "just" a new handler/service/repository following established patterns, with MongoDB for messages and extensions to the existing NATS bridge.

## Common Pitfalls

### Pitfall 1: Unindexed MongoDB Queries on Chat History

**What goes wrong:** Chat history query scans all messages in the collection instead of using an index. Response time grows linearly with message count.
**Why it happens:** MongoDB does not create indexes automatically. Default `_id` index is not useful for channel-scoped time-ordered queries.
**How to avoid:** Create compound index on `(channel_id, created_at)` at application startup. Also create a text index on `content_text` for search.
**Warning signs:** Slow chat load (>200ms) with just a few thousand messages.

```go
// Индексы создаются при инициализации модуля
func (r *MongoRepository) EnsureIndexes(ctx context.Context) error {
    indexes := []mongo.IndexModel{
        {
            Keys: bson.D{
                {Key: "channel_id", Value: 1},
                {Key: "created_at", Value: -1},
            },
        },
        {
            Keys: bson.D{
                {Key: "channel_id", Value: 1},
                {Key: "thread_id", Value: 1},
                {Key: "created_at", Value: -1},
            },
        },
        {
            Keys: bson.D{{Key: "content_text", Value: "text"}},
        },
    }
    _, err := r.coll.Indexes().CreateMany(ctx, indexes)
    return err
}
```

### Pitfall 2: WebSocket Message Gap on Reconnect

**What goes wrong:** User loses connection for 30 seconds. Messages sent during that window are never seen. User has to refresh to see missed messages.
**Why it happens:** Core NATS is fire-and-forget. If the WebSocket is disconnected, events are lost.
**How to avoid:** On reconnect, client sends the timestamp of the last received message. Server fetches messages from MongoDB newer than that timestamp and sends them as a batch before resuming real-time delivery.
**Warning signs:** Users report "missing messages" after brief network interruptions.

### Pitfall 3: Unread Count Drift Between Redis and Reality

**What goes wrong:** Redis unread count shows 5, but user sees only 3 unread messages. Or count never resets to 0.
**Why it happens:** Race conditions between increment (on message send) and reset (on channel read). Deleted messages don't decrement the counter.
**How to avoid:** Accept approximate counts (like Slack). On channel open, reset to 0. Don't decrement on delete -- minor inconsistency is acceptable. Periodically reconcile by counting actual unread from MongoDB (optional, low priority).
**Warning signs:** Users complaining about "phantom" unread badges.

### Pitfall 4: DM Channel Duplication

**What goes wrong:** Two users create a DM with each other simultaneously, resulting in two separate DM channels.
**Why it happens:** No unique constraint on the user pair for DM channels.
**How to avoid:** Use a deterministic channel lookup: sort both user IDs, concatenate, and use as a unique identifier or query filter. Create DM channels with a "find or create" pattern inside a transaction.
**Warning signs:** Users see duplicate conversations with the same person.

### Pitfall 5: Thread Count Consistency

**What goes wrong:** Thread reply count on parent message says 10 but only 8 replies exist (some were deleted).
**Why it happens:** `reply_count` is stored denormalized on the parent message document. Deleting a reply does not decrement the counter.
**How to avoid:** Decrement `reply_count` on reply deletion. For accuracy, count replies from MongoDB rather than trusting the counter (on thread open).
**Warning signs:** Thread badge shows wrong number.

### Pitfall 6: Typing Indicator Spam

**What goes wrong:** Typing indicator fires on every keystroke, flooding NATS and WebSocket with events.
**Why it happens:** Naive implementation publishes a NATS event on every `input` event from the editor.
**How to avoid:** Client-side debounce (300ms). Server-side: typing indicators use core NATS (no persistence). Client shows "X is typing..." for 3 seconds after last indicator, then hides. Don't log typing events.
**Warning signs:** High NATS message rate even with idle chat.

## Code Examples

### REST API Routes (following tracker pattern)

```go
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler, rbacSvc *rbac.RBACService) {
    msg := api.Group("/messenger", authMw)

    // Каналы
    channels := msg.Group("/channels")
    channels.Post("/", h.CreateChannel)
    channels.Get("/", h.ListMyChannels)
    channels.Get("/:channelId", h.GetChannel)
    channels.Put("/:channelId", h.UpdateChannel)
    channels.Delete("/:channelId", h.DeleteChannel)
    channels.Post("/:channelId/join", h.JoinChannel)
    channels.Post("/:channelId/leave", h.LeaveChannel)
    channels.Post("/:channelId/members", h.AddMembers)
    channels.Get("/:channelId/members", h.ListMembers)

    // Сообщения
    messages := channels.Group("/:channelId/messages")
    messages.Post("/", h.SendMessage)
    messages.Get("/", h.ListMessages)
    messages.Put("/:messageId", h.EditMessage)
    messages.Delete("/:messageId", h.DeleteMessage)
    messages.Post("/:messageId/reactions", h.ToggleReaction)
    messages.Post("/:messageId/pin", h.TogglePin)
    messages.Post("/:messageId/forward", h.ForwardMessage)

    // Треды
    messages.Get("/:messageId/replies", h.ListThreadReplies)

    // Поиск
    msg.Get("/search", h.SearchMessages)

    // Закладки
    msg.Get("/bookmarks", h.ListBookmarks)
    msg.Post("/bookmarks", h.AddBookmark)
    msg.Delete("/bookmarks/:bookmarkId", h.RemoveBookmark)

    // DM
    msg.Post("/dm", h.CreateOrGetDM)

    // Присутствие
    msg.Get("/presence", h.GetPresence)
    msg.Put("/presence/status", h.SetCustomStatus)

    // Непрочитанные
    msg.Get("/unread", h.GetUnreadCounts)
    msg.Post("/channels/:channelId/read", h.MarkChannelRead)
}
```

### Message Flow: Send -> NATS -> WebSocket

```go
// service.go
func (s *MessengerService) SendMessage(ctx context.Context, channelID, senderID string, req SendMessageRequest) (*MessageResponse, error) {
    // 1. Валидация: пользователь — участник канала
    isMember, err := s.entRepo.IsMember(ctx, channelID, senderID)
    if err != nil || !isMember {
        return nil, apperrors.NewForbidden("не участник канала")
    }

    // 2. Проверка readonly
    channel, err := s.entRepo.GetChannel(ctx, channelID)
    if err != nil {
        return nil, err
    }
    if channel.IsReadonly {
        return nil, apperrors.NewForbidden("канал только для чтения")
    }

    // 3. Сохранить в MongoDB
    doc := MessageDocument{
        ID:          uuid.New().String(),
        ChannelID:   channelID,
        SenderID:    senderID,
        ContentJSON: req.ContentJSON,
        ContentText: extractPlainText(req.ContentJSON),
        ThreadID:    req.ThreadID,
        Attachments: req.Attachments,
        Mentions:    req.Mentions,
        MentionGroups: req.MentionGroups,
        CreatedAt:   time.Now().UTC(),
    }
    if err := s.mongoRepo.InsertMessage(ctx, &doc); err != nil {
        return nil, apperrors.NewInternal("ошибка сохранения сообщения")
    }

    // 4. Если это ответ в тред — инкрементировать reply_count
    if doc.ThreadID != nil {
        _ = s.mongoRepo.IncrementReplyCount(ctx, *doc.ThreadID)
    }

    // 5. Обновить unread counts для всех участников канала
    memberIDs, _ := s.entRepo.GetChannelMemberIDs(ctx, channelID)
    _ = s.unread.Increment(ctx, channelID, memberIDs, senderID)

    // 6. Опубликовать в NATS -> bridge доставит в WebSocket
    event := MessageSentEvent{Message: toMessageResponse(&doc)}
    data, _ := json.Marshal(event)
    _ = s.nats.Publish("messenger.channel."+channelID+".message_sent", data)

    return toMessageResponse(&doc), nil
}
```

### Frontend: WebSocket Room Subscription

```typescript
// hooks/useMessages.ts
function useChannelMessages(channelId: string) {
  const { subscribe, unsubscribe } = useWebSocket();
  const queryClient = useQueryClient();

  useEffect(() => {
    const room = `channel:${channelId}`;
    subscribe(room);

    // Слушаем события из WebSocket
    const handler = (event: ServerEvent) => {
      if (event.room !== room) return;
      switch (event.event) {
        case 'message_sent':
          queryClient.setQueryData(['messages', channelId], (old: Message[]) =>
            [...(old || []), event.payload]
          );
          break;
        case 'message_edited':
          queryClient.setQueryData(['messages', channelId], (old: Message[]) =>
            (old || []).map(m => m.id === event.payload.id ? event.payload : m)
          );
          break;
        case 'message_deleted':
          queryClient.setQueryData(['messages', channelId], (old: Message[]) =>
            (old || []).filter(m => m.id !== event.payload.id)
          );
          break;
      }
    };

    return () => {
      unsubscribe(room);
    };
  }, [channelId]);

  return useQuery({
    queryKey: ['messages', channelId],
    queryFn: () => api.getMessages(channelId),
  });
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Socket.IO for chat | NATS core + single WS gateway | 2024+ | Simpler, no extra dependency, horizontal scaling built-in |
| Messages in PostgreSQL | MongoDB for messages | Well-established | Better write perf, flexible schema for reactions/threads |
| Polling for presence | Redis sorted set + WS heartbeat | Standard | Sub-second presence updates, O(log N) Redis ops |
| Offset pagination in chat | Cursor-based (created_at) | Standard | Consistent results when new messages arrive |
| Separate read-receipt tables | No read receipts, unread counts only | Project decision | Simpler, matches Slack model |

## Open Questions

1. **Voice message encoding format**
   - What we know: Audio recorded in browser, uploaded via presigned URL to SeaweedFS
   - What's unclear: Exact audio codec (Opus/WebM vs MP3), max duration limit, browser MediaRecorder API compatibility
   - Recommendation: Use Opus/WebM (native browser format via MediaRecorder), convert to playable format client-side if needed. Set max duration 5 minutes.

2. **Channel member cache strategy**
   - What we know: Need member IDs for unread count increment on every message
   - What's unclear: Whether to cache in Redis or in-memory with invalidation
   - Recommendation: Redis set `channel:members:{channelID}` updated on join/leave. Read on every message send is still O(N) but fast in Redis.

3. **@all / @online / @role mention notification scope**
   - What we know: These group mentions need to resolve to actual user IDs for notification dispatch
   - What's unclear: How to efficiently resolve @online (requires presence check for all channel members)
   - Recommendation: @all = all channel members (from Redis cache). @online = intersect channel members with presence:online sorted set. @role = query Ent for project members with matching role.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + testify 1.10.x |
| Config file | None specific -- standard `go test` |
| Quick run command | `go test ./backend/internal/messenger/... -v -count=1` |
| Full suite command | `go test ./backend/... -v -count=1 -race` |

### Phase Requirements to Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| MSG-01 | Create DM and group chat | unit | `go test ./backend/internal/messenger/ -run TestCreateDM -v` | No -- Wave 0 |
| MSG-02 | Create public/private channels, join public | unit | `go test ./backend/internal/messenger/ -run TestCreateChannel -v` | No -- Wave 0 |
| MSG-03 | Auto-channel creation on project create | integration | `go test ./backend/internal/messenger/ -run TestAutoChannel -v` | No -- Wave 0 |
| MSG-04 | Read-only channel blocks message send | unit | `go test ./backend/internal/messenger/ -run TestReadOnlyChannel -v` | No -- Wave 0 |
| MSG-05 | Send reply in thread, verify reply_count | unit | `go test ./backend/internal/messenger/ -run TestThreadReply -v` | No -- Wave 0 |
| MSG-06 | Toggle reaction on/off | unit | `go test ./backend/internal/messenger/ -run TestToggleReaction -v` | No -- Wave 0 |
| MSG-07 | Message with Plate JSON stored correctly | unit | `go test ./backend/internal/messenger/ -run TestRichText -v` | No -- Wave 0 |
| MSG-08 | Edit own message, admin deletes any | unit | `go test ./backend/internal/messenger/ -run TestEditDelete -v` | No -- Wave 0 |
| MSG-09 | Forward, pin, bookmark message | unit | `go test ./backend/internal/messenger/ -run TestForwardPinBookmark -v` | No -- Wave 0 |
| MSG-10 | Voice message with audio URL | unit | `go test ./backend/internal/messenger/ -run TestVoiceMessage -v` | No -- Wave 0 |
| MSG-12 | Search messages by text | integration | `go test ./backend/internal/messenger/ -run TestSearchMessages -v` | No -- Wave 0 |
| MSG-13 | Presence heartbeat and status query | unit | `go test ./backend/internal/messenger/ -run TestPresence -v` | No -- Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./backend/internal/messenger/... -v -count=1`
- **Per wave merge:** `go test ./backend/... -v -count=1 -race`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `backend/internal/messenger/service_test.go` -- unit tests with mock repository
- [ ] `backend/internal/messenger/mongo_repository_test.go` -- integration tests with testcontainers
- [ ] `backend/internal/messenger/presence_test.go` -- Redis presence tests with miniredis
- [ ] `backend/internal/messenger/unread_test.go` -- Redis unread count tests
- [ ] MongoDB test container setup in testcontainers (new for project, PG already used)

## Sources

### Primary (HIGH confidence)
- MongoDB Go Driver v2 docs (go.mongodb.org/mongo-driver/v2) -- BSON struct tags, indexing, Find API
- Redis go-redis/v9 sorted sets and hash commands -- verified from redis.io docs
- NATS core pub/sub documentation (docs.nats.io) -- fire-and-forget pattern for typing/presence
- Existing codebase: ws/hub.go, ws/bridge.go, ws/handler.go, ws/protocol.go -- room subscription and NATS bridge patterns
- Existing codebase: tracker/handler.go, tracker/service.go -- established module pattern

### Secondary (MEDIUM confidence)
- MongoDB text index documentation -- basic text search, sufficient for 10-50 users
- Redis presence pattern with sorted sets -- common pattern, verified across multiple sources

### Tertiary (LOW confidence)
- MediaRecorder API browser support for Opus/WebM -- assumed good coverage in modern browsers, needs validation for Safari

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- all dependencies already installed and verified in Phase 1+2
- Architecture: HIGH -- follows established patterns from tracker module, dual-storage split is a project decision
- MongoDB schema: HIGH -- standard document model for chat, compound indexes are well-documented
- NATS bridge extension: HIGH -- existing bridge code is clear, adding messenger.channel.> subscription is trivial
- Redis presence/unread: HIGH -- standard Redis patterns, O(1)/O(log N) operations
- Voice messages: MEDIUM -- browser MediaRecorder API compatibility needs validation
- Pitfalls: HIGH -- all messenger pitfalls are well-documented in the industry

**Research date:** 2026-03-21
**Valid until:** 2026-04-21 (stable patterns, no fast-moving dependencies)
