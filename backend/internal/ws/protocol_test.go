package ws

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageTypeConstants(t *testing.T) {
	assert.Equal(t, MessageType("subscribe"), TypeSubscribe)
	assert.Equal(t, MessageType("unsubscribe"), TypeUnsubscribe)
	assert.Equal(t, MessageType("event"), TypeEvent)
	assert.Equal(t, MessageType("ping"), TypePing)
	assert.Equal(t, MessageType("pong"), TypePong)
}

func TestClientMessage_JSON(t *testing.T) {
	raw := `{"type":"subscribe","room":"project:abc-123"}`
	var msg ClientMessage
	err := json.Unmarshal([]byte(raw), &msg)

	require.NoError(t, err)
	assert.Equal(t, TypeSubscribe, msg.Type)
	assert.Equal(t, "project:abc-123", msg.Room)
}

func TestServerMessage_JSON(t *testing.T) {
	msg := ServerMessage{
		Type:  TypeEvent,
		Event: "tracker.issue.updated",
		Room:  "project:abc-123",
		Payload: map[string]string{
			"issue_key": "DEMO-1",
		},
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var decoded ServerMessage
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, TypeEvent, decoded.Type)
	assert.Equal(t, "tracker.issue.updated", decoded.Event)
	assert.Equal(t, "project:abc-123", decoded.Room)
}

func TestClientMessage_PingHasNoRoom(t *testing.T) {
	msg := ClientMessage{Type: TypePing}
	data, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.NotContains(t, string(data), `"room"`, "ping message should omit room (omitempty)")
}

func TestNewClient(t *testing.T) {
	c := NewClient("user-1", "session-abc")
	assert.Equal(t, "user-1", c.UserID)
	assert.Equal(t, "session-abc", c.SessionID)
	assert.NotNil(t, c.send)
	assert.NotNil(t, c.done)
	assert.NotNil(t, c.rooms)
}
