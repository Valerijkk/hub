package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
)

// Client — представляет одно WebSocket-соединение пользователя.
type Client struct {
	UserID    string
	SessionID string
	send      chan []byte
	done      chan struct{}
	rooms     map[string]bool
	mu        sync.Mutex
	closeOnce sync.Once
}

// NewClient создаёт нового клиента для указанного пользователя и сессии.
func NewClient(userID, sessionID string) *Client {
	return &Client{
		UserID:    userID,
		SessionID: sessionID,
		send:      make(chan []byte, 256),
		done:      make(chan struct{}),
		rooms:     make(map[string]bool),
	}
}

// Send возвращает канал для отправки сообщений клиенту.
func (c *Client) Send() chan []byte {
	return c.send
}

// Hub — управляет активными WebSocket-клиентами и подписками на комнаты.
type Hub struct {
	// clients хранит соответствие userID -> *Client. sync.Map для потокобезопасного доступа.
	clients sync.Map

	// rooms хранит соответствие roomKey -> (sync.Map: userID -> *Client).
	rooms sync.Map

	register   chan *Client
	unregister chan *Client
}

// NewHub создаёт новый Hub.
func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
	}
}

// Run запускает цикл обработки событий Hub. Блокирует до отмены контекста.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.clients.Store(client.UserID, client)
			slog.Debug("ws client registered", "userID", client.UserID)
		case client := <-h.unregister:
			h.removeClient(client)
			slog.Debug("ws client unregistered", "userID", client.UserID)
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

// Subscribe подписывает клиента на комнату.
func (h *Hub) Subscribe(client *Client, room string) {
	client.mu.Lock()
	client.rooms[room] = true
	client.mu.Unlock()

	actual, _ := h.rooms.LoadOrStore(room, &sync.Map{})
	members := actual.(*sync.Map)
	members.Store(client.UserID, client)

	slog.Debug("ws client subscribed", "userID", client.UserID, "room", room)
}

// Unsubscribe отписывает клиента от комнаты.
func (h *Hub) Unsubscribe(client *Client, room string) {
	client.mu.Lock()
	delete(client.rooms, room)
	client.mu.Unlock()

	if val, ok := h.rooms.Load(room); ok {
		members := val.(*sync.Map)
		members.Delete(client.UserID)
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
				"userID", client.UserID, "room", room)
		}
		return true
	})
}

// SendToUser отправляет сообщение конкретному пользователю, если он подключён.
func (h *Hub) SendToUser(userID string, msg ServerMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal server message", "error", err)
		return
	}

	val, ok := h.clients.Load(userID)
	if !ok {
		return
	}

	client := val.(*Client)
	select {
	case client.send <- data:
	default:
		slog.Warn("ws client send buffer full, dropping message", "userID", userID)
	}
}

// removeClient удаляет клиента из карты клиентов и всех комнат.
func (h *Hub) removeClient(client *Client) {
	h.clients.Delete(client.UserID)

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
			members.Delete(client.UserID)
		}
	}

	// Сигнализируем writePump о завершении ДО закрытия send-канала,
	// чтобы writePump успел выйти и не обращался к закрытому соединению.
	client.closeOnce.Do(func() {
		close(client.done)
		close(client.send)
	})
}
