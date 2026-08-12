# Backend: NATS JetStream, outbox, MongoDB

> Часть плейбука. Канон и оглавление — [`../backend.md`](../backend.md).

Асинхронный обмен, гарантии доставки, документные данные.

---

## 8. NATS JetStream и outbox

### 8.1. NATS — единственный pub/sub

NATS + JetStream — единственная шина событий Hub. **Redis pub/sub использовать запрещено.** Core NATS (`Publish`) — для эфемерных real-time сигналов (typing, presence); JetStream — для durable-событий, которые должны пережить рестарт (уведомления, webhook-диспатч, аудит).

### 8.2. Клиент и стрим

<good-example>

Реальный `platform/nats/client.go`:

```go
const StreamName = "EVENTS" // единый JetStream stream для всех модулей

type Client struct {
    Conn      *nats.Conn
    JetStream jetstream.JetStream
}

func (c *Client) EnsureStreams(ctx context.Context) error {
    _, err := c.JetStream.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
        Name: StreamName,
        Subjects: []string{
            "project.>", "tracker.>", "messenger.channel.>",
            "tms.>", "calls.channel.>", "calls.recording.>",
            "auth.session.>", "sprint.>", "notification.>",
        },
        Retention: jetstream.InterestPolicy,
        MaxAge:    7 * 24 * time.Hour,
    })
    return err
}
```

Core-публикация (fire-and-forget, nil-safe):

```go
func (c *Client) PublishJSON(subject string, payload interface{}) {
    if c == nil {
        slog.Debug("NATS не подключён, пропускаем публикацию", "subject", subject)
        return
    }
    data, err := json.Marshal(payload)
    if err != nil {
        slog.Error("не удалось сериализовать событие", "subject", subject, "error", err)
        return
    }
    if err := c.Conn.Publish(subject, data); err != nil {
        slog.Error("не удалось опубликовать событие", "subject", subject, "error", err)
    }
}
```
</good-example>

### 8.3. Соглашение об subjects

Формат subject: `hub.<domain>.<event>` в общей ментальной модели; в реальном коде используются доменно- specific иерархии, совместимые с фильтрами стрима:

- `project.<projectID>.issue.created` / `.updated` / `.moved` / `.assigned`
- `tracker.<...>.issue.>`
- `messenger.channel.<id>.message_sent`
- `auth.session.created`
- `notification.>`, `sprint.>`, `tms.run.completed.>`

Иерархия с точками позволяет подписчикам фильтровать по wildcard (`project.*.issue.>`). Именование должно быть предсказуемым: `<домен>.<сущность>.<действие в прошедшем времени>`.

<good-example>

Реальные фильтры webhook-консьюмера (`nats/client.go`):

```go
var WebhookFilterSubjects = []string{
    "project.*.issue.>",
    "tracker.*.issue.>",
    "project.created",
    "tms.run.completed.>",
    "sprint.>",
    "messenger.channel.*.message_sent",
}
```

Заметьте: `notification.*`, `auth.session.*`, `calls.*` намеренно исключены из webhook-фильтра — они не доставляются наружу (APIGW-03).
</good-example>

### 8.4. Outbox-паттерн (transactional outbox)

Прямая публикация в NATS внутри бизнес-операции опасна: если БД закоммитилась, а публикация упала (или наоборот) — рассинхрон. Решение — **transactional outbox**: событие пишется в таблицу `outbox_event` в ТОЙ ЖЕ транзакции, что и данные; отдельный worker publish'ит его в NATS с ретраями (at-least-once).

<good-example>

Публикатор (`platform/outbox/publisher.go`):

```go
type Publisher struct{}

// PublishInTx: событие в outbox внутри существующей Ent-транзакции — атомарно с данными.
func (p *Publisher) PublishInTx(ctx context.Context, tx *ent.Tx, eventType string, payload any) error {
    data, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    _, err = tx.OutboxEvent.Create().
        SetEventType(eventType).
        SetPayload(data).
        Save(ctx)
    return err
}

// Publish: без внешней транзакции (автономно).
func (p *Publisher) Publish(ctx context.Context, client *ent.Client, eventType string, payload any) error {
    data, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    _, err = client.OutboxEvent.Create().SetEventType(eventType).SetPayload(data).Save(ctx)
    return err
}
```

Репозиторий использует `PublishInTx` внутри `client.Tx` (см. §5.5 `CreateIssue`). Интерфейс `OutboxPublisher` в `tracker/repository.go`:

```go
type OutboxPublisher interface {
    PublishInTx(ctx context.Context, tx *ent.Tx, eventType string, payload any) error
    Publish(ctx context.Context, client *ent.Client, eventType string, payload any) error
}
```
</good-example>

### 8.5. Outbox worker: at-least-once + ретраи + изоляция сбоев

<good-example>

Worker (`platform/outbox/worker.go`) — polling каждые 100мс, до 50 событий за батч, до 10 попыток, локальный recover на каждое событие:

```go
func (w *Worker) processBatch(ctx context.Context) {
    events, err := w.client.OutboxEvent.Query().
        Where(
            outboxevent.StatusEQ(outboxevent.StatusPending),
            outboxevent.AttemptsLT(10),
        ).
        Order(ent.Asc(outboxevent.FieldCreatedAt)).
        Limit(50).
        All(ctx)
    if err != nil {
        slog.Error("outbox: failed to query events", "error", err)
        return
    }
    for _, evt := range events {
        w.processEvent(ctx, evt) // recover внутри — сбой одной записи не валит батч
    }
}

func (w *Worker) processEvent(ctx context.Context, evt *ent.OutboxEvent) {
    defer func() {
        if r := recover(); r != nil {
            slog.Error("outbox: panic while processing event",
                "event_id", evt.ID, "event_type", evt.EventType,
                "panic", r, "stack", string(debug.Stack()))
        }
    }()

    if err := w.nats.Publish(evt.EventType, evt.Payload); err != nil {
        // Инкремент попытки + запись last_error, событие попадёт в следующую итерацию.
        w.client.OutboxEvent.UpdateOneID(evt.ID).
            SetAttempts(evt.Attempts + 1).
            SetLastError(err.Error()).
            Save(ctx)
        return
    }
    // Успех — помечаем published.
    w.client.OutboxEvent.UpdateOneID(evt.ID).
        SetStatus(outboxevent.StatusPublished).
        SetPublishedAt(time.Now()).
        Save(ctx)
}
```

Семантика: **at-least-once**. Если publish прошёл, но пометить `published` не удалось — событие пере-опубликуется. Это осознанный выбор: «лучше доставить дважды, чем потерять».
</good-example>

**Масштабирование и трейсинг worker'а (2026):**

- **2+ экземпляра worker'а — только с `SELECT ... FOR UPDATE SKIP LOCKED`.** Текущий поллинг `StatusPending` без блокировок корректен ровно для одного экземпляра; при горизонтальном масштабировании relay-воркеров два инстанса заберут один батч и опубликуют дубли. В Ent — выборка батча внутри транзакции с `query.ForUpdate(sql.WithLockAction(sql.SkipLocked))`.
- **Publish-side дедуп через `Nats-Msg-Id`.** Worker при публикации ставит заголовок `Nats-Msg-Id` = ID outbox-события; JetStream с настроенным `Duplicates`-окном стрима сам отбрасывает повторную публикацию после ретрая worker'а. Практический «exactly-once» = Msg-Id дедуп + идемпотентный консьюмер (§8.6).
- **Trace context через событие.** Span context сериализуется в колонку outbox-таблицы; worker инжектит `traceparent`/`tracestate` в NATS-заголовки (`TextMapPropagator` + header carrier), консьюмер извлекает их при обработке. Без этого трейс рвётся на границе producer → outbox → consumer.

### 8.6. Идемпотентность консьюмеров

Раз доставка at-least-once, консьюмеры обязаны быть идемпотентны: одно и то же событие может прийти дважды. Способы:

- Дедупликация по идентификатору события/сущности (обработали `issue.ID` + `event_type` — пропускаем повтор).
- Upsert вместо insert там, где возможно.
- Естественная идемпотентность (установка статуса в конкретное значение).
- Publish-side дополнение (2026): заголовок `Nats-Msg-Id` + `Duplicates`-окно стрима (§8.5) отсеивает повторные публикации ещё на стороне JetStream — но consumer-side дедуп это не отменяет (redelivery после потерянного ack всё равно возможен).

<good-example>

```go
// Консьюмер уведомлений: дедуп по (event_id) через Redis SET NX с TTL.
ok, err := rdb.SetNX(ctx, "notif:seen:"+eventID, 1, 24*time.Hour).Result()
if err != nil {
    return err // временный сбой — пусть JetStream доставит снова
}
if !ok {
    return nil // уже обрабатывали — no-op, ack
}
// … доставить уведомление …
```
</good-example>

<bad-example>

```go
// Не идемпотентно: при повторной доставке пользователь получит два одинаковых пуша.
sendPush(ctx, userID, msg)
```
</bad-example>

### 8.7. Durable consumers

Каждый модуль-потребитель (audit, webhook-dispatcher, notification) создаёт свой durable consumer на стриме `EVENTS`. Durable = позиция чтения переживает рестарт. Эфемерные подписки (`calls.chat.>`) — для real-time без реплея.

### 8.8. Чек-лист NATS/outbox

- [ ] Используется NATS/JetStream, не Redis pub/sub.
- [ ] Subject именуется предсказуемо: `<домен>.<сущность>.<действие>` (глагол в прошедшем времени).
- [ ] Событие, критичное к консистентности, пишется через `PublishInTx` в той же транзакции, что и данные.
- [ ] Эфемерные real-time сигналы — через core `Publish`/`PublishJSON` (nil-safe).
- [ ] Консьюмер идемпотентен (дедуп/upsert) — доставка at-least-once.
- [ ] Долгоживущие консьюмеры — durable; позиция переживает рестарт.
- [ ] Payload сериализуется в JSON; структура события в `events.go` модуля.
- [ ] (2026) При 2+ экземплярах outbox-worker'а — `SKIP LOCKED`; публикация с `Nats-Msg-Id`; trace context пропагируется через NATS-заголовки.

---
## 9. MongoDB

### 9.1. Где используется MongoDB

PostgreSQL — основная реляционная БД (через Ent). MongoDB 8 используется для документо-ориентированных данных с гибкой схемой и высоким объёмом:

- **messenger** — сообщения чатов (высокая частота записи, гибкая структура, cursor-пагинация).
- **calls** — чат внутри звонков.
- **TMS** — Allure/JUnit отчёты (вложенные структуры результатов тестов).

Драйвер — официальный `go.mongodb.org/mongo-driver/v2`.

### 9.2. Подключение

<good-example>

Реальный `platform/database/mongodb.go`:

```go
func NewMongoClient(ctx context.Context, url, dbName string) (*mongo.Database, error) {
    client, err := mongo.Connect(options.Client().ApplyURI(url))
    if err != nil {
        return nil, err
    }
    if err := client.Ping(ctx, nil); err != nil {
        return nil, err
    }
    slog.Info("connected to MongoDB", "database", dbName)
    return client.Database(dbName), nil
}
```
</good-example>

### 9.3. Паттерн репозитория для Mongo

Тот же принцип слоёв: репозиторий инкапсулирует `*mongo.Collection`, service работает через интерфейс. Mongo-специфика (`bson.M`, курсоры) не протекает в service.

<good-example>

```go
type MessageRepository interface {
    InsertMessage(ctx context.Context, m *Message) (string, error)
    ListByChannel(ctx context.Context, channelID uuid.UUID, before string, limit int) ([]*Message, error)
}

type MongoMessageRepository struct {
    col *mongo.Collection
}

func NewMongoMessageRepository(db *mongo.Database) *MongoMessageRepository {
    return &MongoMessageRepository{col: db.Collection("messages")}
}

func (r *MongoMessageRepository) ListByChannel(ctx context.Context, channelID uuid.UUID, before string, limit int) ([]*Message, error) {
    filter := bson.M{"channel_id": channelID.String()}
    if before != "" {
        // Cursor-пагинация: сообщения старше курсора.
        filter["_id"] = bson.M{"$lt": before}
    }
    opts := options.Find().
        SetSort(bson.D{{Key: "_id", Value: -1}}). // новые сверху
        SetLimit(int64(limit))
    cur, err := r.col.Find(ctx, filter, opts)
    if err != nil {
        return nil, fmt.Errorf("find messages: %w", err)
    }
    defer cur.Close(ctx)

    var out []*Message
    if err := cur.All(ctx, &out); err != nil {
        return nil, fmt.Errorf("decode messages: %w", err)
    }
    return out, nil
}
```

Всегда `defer cur.Close(ctx)`; всегда `ctx` в каждой операции; ошибки оборачиваются `%w`.
</good-example>

### 9.4. Скоуп по пользователю (security)

Как и в §7.5 — поиск/выборка по Mongo фильтруется данными пользователя. Никогда не отдавайте коллекцию без фильтра принадлежности.

<good-example>

```go
// Поиск сообщений — только каналы, где пользователь состоит (NF-03).
filter := bson.M{
    "channel_id": bson.M{"$in": userChannelIDs},
    "$text":      bson.M{"$search": query},
}
```
</good-example>

### 9.5. Индексы

Индексы создаются программно при инициализации (idempotent через `CreateOne`/`CreateMany`). Индексируйте поля фильтра/сортировки: `channel_id`, `created_at`, текстовые индексы для поиска.

<good-example>

```go
_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{
    {Keys: bson.D{{Key: "channel_id", Value: 1}, {Key: "_id", Value: -1}}}, // листинг по каналу + cursor
    {Keys: bson.D{{Key: "content", Value: "text"}}},                        // полнотекстовый поиск
})
if err != nil {
    return fmt.Errorf("create message indexes: %w", err)
}
```
</good-example>

### 9.6. UUID и время в Mongo

- UUID сущностей Postgres хранятся как строки (`channelID.String()`) для консистентного матчинга.
- Время — UTC (как и везде). Не полагайтесь на локальную зону клиента при записи.

### 9.7. Чек-лист MongoDB

- [ ] Доступ через репозиторий-интерфейс; `*mongo.Collection` не протекает в service.
- [ ] `ctx` в каждой операции; `defer cur.Close(ctx)` для курсоров.
- [ ] Выборка скоупится по данным пользователя (channel/project IDs).
- [ ] Индексы под фильтры/сортировки/поиск созданы.
- [ ] Ошибки оборачиваются `%w`.
- [ ] Cursor-пагинация для лент (чат), не offset.

---
