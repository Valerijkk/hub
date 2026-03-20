package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
)

// Client represents a single WebSocket connection.
type Client struct {
	UserID    string
	SessionID string
	send      chan []byte
	rooms     map[string]bool
	mu        sync.Mutex
}

// NewClient creates a new client for the given user and session.
func NewClient(userID, sessionID string) *Client {
	return &Client{
		UserID:    userID,
		SessionID: sessionID,
		send:      make(chan []byte, 256),
		rooms:     make(map[string]bool),
	}
}

// Send returns the channel used to push messages to the client.
func (c *Client) Send() chan []byte {
	return c.send
}

// Hub manages all active WebSocket clients and room subscriptions.
type Hub struct {
	// clients maps userID -> *Client. sync.Map for concurrent access.
	clients sync.Map

	// rooms maps roomKey -> (sync.Map of userID -> *Client).
	rooms sync.Map

	register   chan *Client
	unregister chan *Client
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
	}
}

// Run starts the hub event loop. It blocks until ctx is cancelled.
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

// Register adds a client to the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub and all its rooms.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Subscribe adds a client to a room.
func (h *Hub) Subscribe(client *Client, room string) {
	client.mu.Lock()
	client.rooms[room] = true
	client.mu.Unlock()

	actual, _ := h.rooms.LoadOrStore(room, &sync.Map{})
	members := actual.(*sync.Map)
	members.Store(client.UserID, client)

	slog.Debug("ws client subscribed", "userID", client.UserID, "room", room)
}

// Unsubscribe removes a client from a room.
func (h *Hub) Unsubscribe(client *Client, room string) {
	client.mu.Lock()
	delete(client.rooms, room)
	client.mu.Unlock()

	if val, ok := h.rooms.Load(room); ok {
		members := val.(*sync.Map)
		members.Delete(client.UserID)
	}
}

// BroadcastToRoom sends a message to all clients in the given room.
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

// SendToUser sends a message to a specific user if connected.
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

// removeClient removes a client from the clients map and all rooms.
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

	close(client.send)
}
