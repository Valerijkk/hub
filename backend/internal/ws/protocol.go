package ws

// MessageType — тип WebSocket-сообщения (subscribe, event, ping и т.д.).
type MessageType string

// Типы WebSocket-сообщений и протокол.
const (
	TypeSubscribe   MessageType = "subscribe"
	TypeUnsubscribe MessageType = "unsubscribe"
	TypeEvent       MessageType = "event"
	TypePing        MessageType = "ping"
	TypePong        MessageType = "pong"
)

// ClientMessage — сообщение, отправляемое клиентом на сервер через WebSocket.
type ClientMessage struct {
	Type    MessageType `json:"type"`
	Room    string      `json:"room,omitempty"`    // "project:{id}", "channel:{id}"
	Payload any         `json:"payload,omitempty"`
}

// ServerMessage — сообщение, отправляемое сервером клиенту через WebSocket.
type ServerMessage struct {
	Type    MessageType `json:"type"`
	Event   string      `json:"event,omitempty"` // NATS subject
	Room    string      `json:"room,omitempty"`
	Payload any         `json:"payload,omitempty"`
}
