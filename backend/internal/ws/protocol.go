package ws

// MessageType identifies the kind of WebSocket message.
type MessageType string

const (
	TypeSubscribe   MessageType = "subscribe"
	TypeUnsubscribe MessageType = "unsubscribe"
	TypeEvent       MessageType = "event"
	TypePing        MessageType = "ping"
	TypePong        MessageType = "pong"
)

// ClientMessage is a message sent from the client to the server.
type ClientMessage struct {
	Type    MessageType `json:"type"`
	Room    string      `json:"room,omitempty"`    // "project:{id}", "channel:{id}"
	Payload any         `json:"payload,omitempty"`
}

// ServerMessage is a message sent from the server to the client.
type ServerMessage struct {
	Type    MessageType `json:"type"`
	Event   string      `json:"event,omitempty"` // NATS subject
	Room    string      `json:"room,omitempty"`
	Payload any         `json:"payload,omitempty"`
}
