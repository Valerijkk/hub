# Backend: наблюдаемость

> Часть плейбука. Канон и оглавление — [`../backend.md`](../backend.md).

Логи, метрики, трассировка, что писать при отказе.

---

## 10. Конкурентность

### 10.1. Общий принцип

Go-конкурентность мощна и опасна. Правило Hub: любой код с горутинами/каналами тестируется с `-race`, а разделяемое состояние защищается либо каналами (передача владения), либо `sync`-примитивами. Регрессия обязательно включает `-race`.

### 10.2. Горутины и жизненный цикл

Каждая запущенная горутина должна иметь понятный способ завершения (обычно `ctx.Done()`). Горутина без выхода — утечка.

<good-example>

Из `platform/middleware/ratelimit.go` — cleanup завершается по контексту:

```go
func (rl *rateLimiter) cleanup(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return // graceful shutdown, без утечки
        case <-ticker.C:
            // очистка
        }
    }
}
```
</good-example>

<bad-example>

```go
go func() {
    for {
        time.Sleep(5 * time.Minute)
        cleanup() // нет способа остановить — горутина живёт вечно
    }
}()
```
</bad-example>

### 10.3. Сериализация состояния через одну горутину

Мощный приём против гонок: все мутации разделяемой структуры проходят через одну горутину, принимающую команды по каналам. Так сделан WS Hub.

<good-example>

Из `ws/hub.go` — `register`/`unregister`/`subscribe`/`unsubscribe` сериализуются в одной `Run`-горутине:

```go
func (h *Hub) Run(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            h.beginShutdown()
            h.shutdownAllClients()
            return
        case client := <-h.register:
            h.registerClient(client)     // мутация карт — только здесь
        case client := <-h.unregister:
            h.removeClient(client)
        case req := <-h.subscribe:
            h.subscribeClient(req.client, req.room)
        case req := <-h.unsubscribe:
            h.unsubscribeClient(req.client, req.room)
        }
    }
}
```

Публичные методы просто шлют в канал (send-and-forget), не мутируя состояние напрямую:

```go
func (h *Hub) Subscribe(client *Client, room string) {
    select {
    case h.subscribe <- subscribeReq{client: client, room: room}:
    case <-h.stopping: // не блокируемся после shutdown
    }
}
```
</good-example>

### 10.4. Закрытие каналов: правило SEC-05

Самая частая паника в Go-WS — «send on closed channel». Правило Hub (SEC-05): **закрывает канал только одна сторона, и это не та сторона, что пишет.** В Hub канал `send` НЕ закрывается никогда; сигнал завершения идёт через отдельный канал `done`, а все продюсеры проверяют `<-done` перед записью.

<good-example>

Из `ws/hub.go` — модель клиента и продюсеры:

```go
type Client struct {
    send      chan []byte
    done      chan struct{}
    closeOnce sync.Once
    // …
}

// signalClientDone закрывает ТОЛЬКО done (идемпотентно через closeOnce).
func (h *Hub) signalClientDone(client *Client) {
    client.closeOnce.Do(func() {
        close(client.done)
    })
}

// Продюсер: перед записью в send проверяет done через select.
func (h *Hub) BroadcastToRoom(room string, msg ServerMessage) {
    data, _ := json.Marshal(msg)
    val, ok := h.rooms.Load(room)
    if !ok {
        return
    }
    members := val.(*sync.Map)
    members.Range(func(_, value any) bool {
        client := value.(*Client)
        select {
        case <-client.done:
            // Клиент завершается — не отправляем (no-op, не паника).
        case client.send <- data:
            // Отправлено.
        default:
            // Буфер переполнен — drop, чтобы не блокировать Hub (NF-13, fail-soft).
            slog.Warn("ws client send buffer full, dropping message",
                "userID", client.UserID, "connID", client.ConnectionID, "room", room)
        }
        return true
    })
}
```

При удалении клиента (`removeClient`) закрывается только `done`, а `send` намеренно не закрывается — его соберёт GC:

```go
// SEC-05: закрываем ТОЛЬКО done. Канал send не закрывается, т.к. параллельные
// producer'ы пишут в него и проверяют done через select — иначе "send on closed channel".
h.signalClientDone(client)
```
</good-example>

<bad-example>

Классическая паника: receiver закрывает `send`, а параллельный broadcaster пишет в него:

```go
func (h *Hub) removeClient(c *Client) {
    close(c.send) // НЕЛЬЗЯ: другой горутине, которая сейчас делает c.send <- data, прилетит паника
}
```
</bad-example>

### 10.5. sync-примитивы

- `sync.Mutex` — для короткой защиты небольшого состояния (например `Client.rooms` в `ws`).
- `sync.Map` — для конкурентных карт с высоким чтением (реестр клиентов/комнат в `ws`).
- `sync.Once` — идемпотентное однократное действие (`closeOnce`, `stopOnce`).
- Не защищайте мьютексом то, что уже сериализовано одной горутиной (двойная защита = запутанность).

<good-example>

Из `ws/hub.go`:

```go
type Client struct {
    rooms map[string]bool
    mu    sync.Mutex // защищает rooms при доступе из разных горутин
}

func (h *Hub) subscribeClient(client *Client, room string) {
    client.mu.Lock()
    client.rooms[room] = true
    client.mu.Unlock()
    // реестр комнат Hub — sync.Map, безопасен для конкурентного доступа
    actual, _ := h.rooms.LoadOrStore(room, &sync.Map{})
    actual.(*sync.Map).Store(client.ConnectionID, client)
}
```
</good-example>

### 10.6. Неблокирующая отправка (backpressure)

Отправка в канал с `default` (как в §10.4) — стратегия backpressure: при переполнении буфера сообщение дропается, а не блокирует систему (NF-13, осознанный fail-soft для real-time). Для критичных данных используйте durable-путь (outbox/JetStream), а не WS.

### 10.7. Graceful shutdown

При `ctx.Done()`: перестаём принимать новую работу (`stopping` канал), дренируем/закрываем существующие соединения, завершаем фоновые горутины.

<good-example>

Из `ws/hub.go`:

```go
func (h *Hub) beginShutdown() {
    h.stopOnce.Do(func() { close(h.stopping) })
}

func (h *Hub) Register(client *Client) {
    if h.isStopping() {
        h.signalClientDone(client) // после shutdown — сразу сигналим done
        return
    }
    select {
    case h.register <- client:
    case <-h.stopping:
        h.signalClientDone(client)
    }
}
```
</good-example>

### 10.8. Тестирование гонок

Регрессия по конкурентному коду ВСЕГДА с `-race`:

Запускать task-owned Docker job `backend-unit` с focused scope
`./tests/ws/...`, `-race -count=1 -timeout=600s`; boundary/recovery WS cases —
Python `testframework/tests/integration/ws/**` по verified TC; reusable client only at an exact task-scoped `testframework/src/hub_testkit/integration/**` path.

`-count=1` отключает кеш тестов; `-race` детектирует гонки данных; `-timeout` защищает от дедлоков.

Для детерминированного тестирования кода с таймерами/тикерами/фоновыми горутинами (outbox-worker, rate-limiter cleanup, WS hub) используйте `testing/synctest` (stable в Go 1.25): тест выполняется в изолированном «пузыре» с виртуальными часами, `synctest.Wait()` дожидается стабилизации горутин. Это убирает `time.Sleep` и flaky из конкурентных тестов — `-race` при этом остаётся обязательным. (2026)

### 10.9. Чек-лист конкурентности

- [ ] Каждая горутина завершается по `ctx.Done()` (нет утечек).
- [ ] Разделяемое состояние сериализовано (одна горутина) ИЛИ защищено `sync`.
- [ ] Каналы закрываются одной стороной; продюсеры не пишут в закрытый канал.
- [ ] Паттерн `done` + `select` для «обречённых» получателей (SEC-05).
- [ ] `sync.Once` для идемпотентного закрытия/остановки.
- [ ] Неблокирующая отправка (`default`) там, где нельзя блокировать шину.
- [ ] Graceful shutdown реализован (`stopping` + drain).
- [ ] Тесты прогоняются с `-race -count=1`.

---
