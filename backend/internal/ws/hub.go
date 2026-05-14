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
}

// NewHub создаёт новый Hub.
func NewHub() *Hub {
	return &Hub{
		register:    make(chan *Client, 64),
		unregister:  make(chan *Client, 64),
		subscribe:   make(chan subscribeReq, 256),
		unsubscribe: make(chan subscribeReq, 256),
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

// Register добавляет клиента в Hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister удаляет клиента из Hub и всех его комнат.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Subscribe ставит клиента в очередь подписки на комнату. Семантика
// send-and-forget: вызывающий не ждёт ack. Обработка происходит в Run.
//
// HI-07 follow-through: теперь это просто послание в канал — фактическая
// мутация h.rooms делается subscribeClient под управлением Run.
func (h *Hub) Subscribe(client *Client, room string) {
	h.subscribe <- subscribeReq{client: client, room: room}
}

// Unsubscribe ставит клиента в очередь отписки от комнаты.
func (h *Hub) Unsubscribe(client *Client, room string) {
	h.unsubscribe <- subscribeReq{client: client, room: room}
}

// registerClient вставляет клиента в обе карты (clients + userClients).
// Вызывается только из горутины Run, поэтому не требует внешней синхронизации.
func (h *Hub) registerClient(client *Client) {
	h.clients.Store(client.ConnectionID, client)

	inner, _ := h.userClients.LoadOrStore(client.UserID, &sync.Map{})
	inner.(*sync.Map).Store(client.ConnectionID, client)
}

// subscribeClient выполняет фактическую подписку — вызывается только из Run,
// что серилизует её с register/unregister/unsubscribe для того же клиента.
// Ключом во внутренней карте служит ConnectionID, чтобы два соединения
// одного и того же юзера могли независимо состоять в одной комнате.
func (h *Hub) subscribeClient(client *Client, room string) {
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

	if val, ok := h.rooms.Load(room); ok {
		members := val.(*sync.Map)
		members.Delete(client.ConnectionID)
	}
}

// BroadcastToRoom рассылает сообщение всем клиентам в указанной комнате.
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
		case client.send <- data:
		default:
			// Client send buffer full, skip message to avoid blocking.
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
		select {
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
		if val, ok := h.rooms.Load(room); ok {
			members := val.(*sync.Map)
			members.Delete(client.ConnectionID)
		}
	}

	// Сигнализируем writePump о завершении ДО закрытия send-канала,
	// чтобы writePump успел выйти и не обращался к закрытому соединению.
	client.closeOnce.Do(func() {
		close(client.done)
		close(client.send)
	})
}
