package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

// Client — представляет одно WebSocket-соединение пользователя.
//
// HI-07: ConnectionID — уникальный идентификатор именно соединения
// (а не пользователя). Один пользователь может иметь несколько активных
// клиентов (мульти-таб) — каждый со своим ConnectionID.
//
// SEC-05: канал send НИКОГДА не закрывается со стороны Hub/receiver.
// Сигнал завершения распространяется через done — все producer-горутины
// (BroadcastToRoom, SendToUser, readPump-pong) обязаны проверять <-done
// перед записью в send. Это устраняет антипаттерн "receiver closes channel",
// при котором параллельные writer'ы могли отправить в уже закрытый канал
// и получить панику "send on closed channel".
type Client struct {
	UserID       string
	SessionID    string
	ConnectionID string
	send         chan []byte
	done         chan struct{}
	rooms        map[string]bool
	mu           sync.Mutex
	closeOnce    sync.Once
}

// NewClient создаёт нового клиента для указанного пользователя и сессии.
// ConnectionID генерируется автоматически как UUID v4, обеспечивая
// уникальность даже при множественных одновременных соединениях одного юзера.
func NewClient(userID, sessionID string) *Client {
	return &Client{
		UserID:       userID,
		SessionID:    sessionID,
		ConnectionID: uuid.NewString(),
		send:         make(chan []byte, 256),
		done:         make(chan struct{}),
		rooms:        make(map[string]bool),
	}
}

// Send возвращает канал для отправки сообщений клиенту.
func (c *Client) Send() chan []byte {
	return c.send
}

// subscribeReq — внутренняя структура, протаскиваемая через канал
// h.subscribe / h.unsubscribe. Сериализует все переходы состояния
// (register/unregister/subscribe/unsubscribe) в одной горутине Run.
type subscribeReq struct {
	client *Client
	room   string
}

// Hub — управляет активными WebSocket-клиентами и подписками на комнаты.
//
// HI-07: хранилище клиентов организовано в две карты:
//   - clients: connectionID -> *Client  (полный список соединений для cleanup
//     и быстрого поиска по ID соединения)
//   - userClients: userID -> *sync.Map  (внутренняя карта connectionID ->
//     *Client; используется SendToUser для веера по всем табам пользователя)
//
// Это устраняет проблему "новое соединение того же юзера затирает старое",
// которую вызывал старый clients map с ключом userID.
//
// HI-07 follow-through: Subscribe/Unsubscribe теперь идут через канал
// (subscribe/unsubscribe) и обрабатываются в той же горутине Run, что и
// register/unregister. Это сериализует ВСЕ переходы состояния и устраняет
// гонку быстрый-reconnect + Subscribe-после-Unregister.
type Hub struct {
	clients     sync.Map // connectionID -> *Client
	userClients sync.Map // userID       -> *sync.Map[connectionID -> *Client]

	// rooms хранит соответствие roomKey -> (sync.Map: connectionID -> *Client).
	// HI-07: ключ внутренней карты — connectionID, а не userID, иначе одно
	// соединение пользователя могло вытеснить другое из комнаты.
	rooms sync.Map

	register    chan *Client
	unregister  chan *Client
	subscribe   chan subscribeReq
	unsubscribe chan subscribeReq

	// stopping закрывается один раз при ctx.Done() до drain клиентов.
	// Публичные методы select'ят на него, чтобы не блокироваться на каналах
	// после выхода Run (WS-01 / FA-01 P3).
	stopOnce  sync.Once
	stopping  chan struct{}
}

// NewHub создаёт новый Hub.
func NewHub() *Hub {
	return &Hub{
		register:    make(chan *Client, 64),
		unregister:  make(chan *Client, 64),
		subscribe:   make(chan subscribeReq, 256),
		unsubscribe: make(chan subscribeReq, 256),
		stopping:    make(chan struct{}),
	}
}

// Run запускает цикл обработки событий Hub. Блокирует до отмены контекста.
//
// HI-07 follow-through: register / unregister / subscribe / unsubscribe
// обрабатываются в одной горутине, что серилизует все мутации структур
// clients / userClients / rooms между собой.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			h.beginShutdown()
			h.shutdownAllClients()
			return
		case client := <-h.register:
			h.registerClient(client)
			slog.Debug("ws client registered",
				"userID", client.UserID, "connID", client.ConnectionID)
		case client := <-h.unregister:
			h.removeClient(client)
			slog.Debug("ws client unregistered",
				"userID", client.UserID, "connID", client.ConnectionID)
		case req := <-h.subscribe:
			h.subscribeClient(req.client, req.room)
			slog.Debug("ws client subscribed",
				"userID", req.client.UserID, "connID", req.client.ConnectionID,
				"room", req.room)
		case req := <-h.unsubscribe:
			h.unsubscribeClient(req.client, req.room)
			slog.Debug("ws client unsubscribed",
				"userID", req.client.UserID, "connID", req.client.ConnectionID,
				"room", req.room)
		}
	}
}

func (h *Hub) beginShutdown() {
	h.stopOnce.Do(func() {
		close(h.stopping)
	})
}

func (h *Hub) isStopping() bool {
	select {
	case <-h.stopping:
		return true
	default:
		return false
	}
}

// shutdownAllClients снимает все активные соединения и закрывает client.done.
// Вызывается только из Run при отмене контекста.
func (h *Hub) shutdownAllClients() {
	h.clients.Range(func(_, value any) bool {
		h.removeClient(value.(*Client))
		return true
	})
}

func (h *Hub) signalClientDone(client *Client) {
	client.closeOnce.Do(func() {
		close(client.done)
	})
}

// Register добавляет клиента в Hub.
func (h *Hub) Register(client *Client) {
	if h.isStopping() {
		h.signalClientDone(client)
		return
	}
	select {
	case h.register <- client:
	case <-h.stopping:
		h.signalClientDone(client)
	}
}

// UnregisterUser удаляет все WS-соединения пользователя (W3-A4-004 / P2-BE-14).
// Используется при revoke/logout сессии: instant revocation не должна оставлять
// клиента в rooms до pong-timeout.
func (h *Hub) UnregisterUser(userID string) {
	val, ok := h.userClients.Load(userID)
	if !ok {
		return
	}
	inner := val.(*sync.Map)
	var clients []*Client
	inner.Range(func(_, value any) bool {
		clients = append(clients, value.(*Client))
		return true
	})
	for _, c := range clients {
		h.Unregister(c)
	}
}

// UnregisterBySession удаляет соединения с совпадающим SessionID.
// Если sessionID пуст — UnregisterUser(userID).
func (h *Hub) UnregisterBySession(userID, sessionID string) {
	if sessionID == "" {
		h.UnregisterUser(userID)
		return
	}
	val, ok := h.userClients.Load(userID)
	if !ok {
		return
	}
	inner := val.(*sync.Map)
	var clients []*Client
	inner.Range(func(_, value any) bool {
		c := value.(*Client)
		if c.SessionID == sessionID {
			clients = append(clients, c)
		}
		return true
	})
	for _, c := range clients {
		h.Unregister(c)
	}
}

// Unregister удаляет клиента из Hub и всех его комнат.
func (h *Hub) Unregister(client *Client) {
	if h.isStopping() {
		h.signalClientDone(client)
		return
	}
	select {
	case h.unregister <- client:
	case <-h.stopping:
		h.signalClientDone(client)
	}
}

// Subscribe ставит клиента в очередь подписки на комнату. Семантика
// send-and-forget: вызывающий не ждёт ack. Обработка происходит в Run.
//
// HI-07 follow-through: теперь это просто послание в канал — фактическая
// мутация h.rooms делается subscribeClient под управлением Run.
func (h *Hub) Subscribe(client *Client, room string) {
	select {
	case h.subscribe <- subscribeReq{client: client, room: room}:
	case <-h.stopping:
	}
}

// Unsubscribe ставит клиента в очередь отписки от комнаты.
func (h *Hub) Unsubscribe(client *Client, room string) {
	select {
	case h.unsubscribe <- subscribeReq{client: client, room: room}:
	case <-h.stopping:
	}
}

// registerClient вставляет клиента в обе карты (clients + userClients).
// Вызывается только из горутины Run, поэтому не требует внешней синхронизации.
//
// W3-A4-001 / P1-BE-21: если Unregister (removeClient) уже закрыл done до того,
// как select в Run дошёл до register, клиента нельзя класть в maps — иначе
// получается зомби (writePump уже вышел, повторного Unregister нет).
func (h *Hub) registerClient(client *Client) {
	select {
	case <-client.done:
		return
	default:
	}

	h.clients.Store(client.ConnectionID, client)

	inner, _ := h.userClients.LoadOrStore(client.UserID, &sync.Map{})
	inner.(*sync.Map).Store(client.ConnectionID, client)
}

// subscribeClient выполняет фактическую подписку — вызывается только из Run,
// что серилизует её с register/unregister/unsubscribe для того же клиента.
// Ключом во внутренней карте служит ConnectionID, чтобы два соединения
// одного и того же юзера могли независимо состоять в одной комнате.
//
// W3-A4-002: late Subscribe после Unregister (done уже закрыт) не реанимирует
// мёртвый Client в rooms. Не требуем clients.Load: Register и Subscribe могут
// одновременно ждать в select Run — Subscribe раньше Register валиден.
func (h *Hub) subscribeClient(client *Client, room string) {
	select {
	case <-client.done:
		return
	default:
	}

	client.mu.Lock()
	client.rooms[room] = true
	client.mu.Unlock()

	actual, _ := h.rooms.LoadOrStore(room, &sync.Map{})
	members := actual.(*sync.Map)
	members.Store(client.ConnectionID, client)
}

// unsubscribeClient выполняет фактическую отписку — вызывается только из Run.
func (h *Hub) unsubscribeClient(client *Client, room string) {
	client.mu.Lock()
	delete(client.rooms, room)
	client.mu.Unlock()

	h.removeRoomMember(room, client.ConnectionID)
}

// removeRoomMember убирает conn из комнаты и удаляет пустой ключ из h.rooms
// (W3-A4-007): иначе уникальные room id копят пустые *sync.Map при uptime.
// Мутации rooms сериализованы в Run — concurrent Store здесь не конкурирует.
func (h *Hub) removeRoomMember(room, connID string) {
	val, ok := h.rooms.Load(room)
	if !ok {
		return
	}
	members := val.(*sync.Map)
	members.Delete(connID)
	empty := true
	members.Range(func(_, _ any) bool {
		empty = false
		return false
	})
	if empty {
		h.rooms.Delete(room)
	}
}

// BroadcastToRoom рассылает сообщение всем клиентам в указанной комнате.
//
// SEC-05: перед записью в client.send проверяем <-client.done, чтобы не
// отправлять в "обречённый" клиент. Сам канал send никогда не закрывается,
// поэтому отправка после done — это просто no-op (drop), а не паника.
func (h *Hub) BroadcastToRoom(room string, msg ServerMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal server message", "error", err)
		return
	}

	val, ok := h.rooms.Load(room)
	if !ok {
		return
	}

	members := val.(*sync.Map)
	members.Range(func(_, value any) bool {
		client := value.(*Client)
		select {
		case <-client.done:
			// Client завершается — не отправляем.
		case client.send <- data:
		default:
			// Буфер send переполнен — drop, чтобы не блокировать Hub/Broadcast.
			slog.Warn("ws client send buffer full, dropping message",
				"userID", client.UserID, "connID", client.ConnectionID, "room", room)
		}
		return true
	})
}

// SendToUser отправляет сообщение всем активным соединениям пользователя
// (HI-07 multi-tab fan-out). Если ни одно соединение не открыто, ничего
// не делает.
func (h *Hub) SendToUser(userID string, msg ServerMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal server message", "error", err)
		return
	}

	val, ok := h.userClients.Load(userID)
	if !ok {
		return
	}

	inner := val.(*sync.Map)
	inner.Range(func(_, value any) bool {
		client := value.(*Client)
		// SEC-05: учитываем done, чтобы не писать в send закрываемого клиента.
		select {
		case <-client.done:
			// Client завершается — не отправляем.
		case client.send <- data:
		default:
			slog.Warn("ws client send buffer full, dropping message",
				"userID", userID, "connID", client.ConnectionID)
		}
		return true
	})
}

// removeClient удаляет клиента из всех структур Hub. Вызывается только из
// горутины Run, что серилизует его с registerClient / subscribeClient /
// unsubscribeClient. Поскольку ключом теперь служит ConnectionID, поздний
// Subscribe старого соединения не может затереть запись нового соединения
// того же пользователя (HI-07).
func (h *Hub) removeClient(client *Client) {
	h.clients.Delete(client.ConnectionID)

	if val, ok := h.userClients.Load(client.UserID); ok {
		inner := val.(*sync.Map)
		inner.Delete(client.ConnectionID)
		// Опционально: убрать пустую запись из userClients, чтобы карта
		// не разбухала после массовых дисконнектов.
		empty := true
		inner.Range(func(_, _ any) bool {
			empty = false
			return false
		})
		if empty {
			h.userClients.Delete(client.UserID)
		}
	}

	client.mu.Lock()
	rooms := make([]string, 0, len(client.rooms))
	for room := range client.rooms {
		rooms = append(rooms, room)
	}
	client.rooms = make(map[string]bool)
	client.mu.Unlock()

	for _, room := range rooms {
		h.removeRoomMember(room, client.ConnectionID)
	}

	// SEC-05: закрываем ТОЛЬКО done. Канал send намеренно не закрывается,
	// поскольку параллельные producer'ы (BroadcastToRoom, SendToUser,
	// readPump-pong) пишут в него и проверяют done через select — без этого
	// мы получали "send on closed channel" panic при гонке remove+broadcast.
	// send будет собран GC после того, как все ссылки на *Client исчезнут.
	h.signalClientDone(client)
}
