package ws_test

// Тесты модуля WebSocket: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/ws"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMessageTypeConstants проверяет сценарий: message type constants.
func TestMessageTypeConstants(t *testing.T) {
	assert.Equal(t, ws.MessageType("subscribe"), ws.TypeSubscribe)
	assert.Equal(t, ws.MessageType("unsubscribe"), ws.TypeUnsubscribe)
	assert.Equal(t, ws.MessageType("event"), ws.TypeEvent)
	assert.Equal(t, ws.MessageType("ping"), ws.TypePing)
	assert.Equal(t, ws.MessageType("pong"), ws.TypePong)
}

// TestClientMessage_JSON проверяет сценарий: client message json.
func TestClientMessage_JSON(t *testing.T) {
	raw := `{"type":"subscribe","room":"project:abc-123"}`
	var msg ws.ClientMessage
	err := json.Unmarshal([]byte(raw), &msg)

	require.NoError(t, err)
	assert.Equal(t, ws.TypeSubscribe, msg.Type)
	assert.Equal(t, "project:abc-123", msg.Room)
}

// TestServerMessage_JSON проверяет сценарий: server message json.
func TestServerMessage_JSON(t *testing.T) {
	msg := ws.ServerMessage{
		Type:  ws.TypeEvent,
		Event: "tracker.issue.updated",
		Room:  "project:abc-123",
		Payload: map[string]string{
			"issue_key": "DEMO-1",
		},
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var decoded ws.ServerMessage
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, ws.TypeEvent, decoded.Type)
	assert.Equal(t, "tracker.issue.updated", decoded.Event)
	assert.Equal(t, "project:abc-123", decoded.Room)
}

// TestClientMessage_RoundTripWithPayload проверяет сценарий: client message round trip with payload.
func TestClientMessage_RoundTripWithPayload(t *testing.T) {
	original := ws.ClientMessage{
		Type: ws.TypeSubscribe,
		Room: "channel:abc",
		Payload: map[string]any{
			"cursor": "xyz",
		},
	}
	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded ws.ClientMessage
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, ws.TypeSubscribe, decoded.Type)
	assert.Equal(t, "channel:abc", decoded.Room)
}

// TestClientMessage_UnsubscribeRoundTrip проверяет сценарий: client message unsubscribe round trip.
func TestClientMessage_UnsubscribeRoundTrip(t *testing.T) {
	raw := `{"type":"unsubscribe","room":"project:deadbeef"}`
	var msg ws.ClientMessage
	require.NoError(t, json.Unmarshal([]byte(raw), &msg))
	assert.Equal(t, ws.TypeUnsubscribe, msg.Type)
	assert.Equal(t, "project:deadbeef", msg.Room)
}

// TestServerMessage_PongOmitsOptionalFields проверяет сценарий: server message pong omits optional fields.
func TestServerMessage_PongOmitsOptionalFields(t *testing.T) {
	data, err := json.Marshal(ws.ServerMessage{Type: ws.TypePong})
	require.NoError(t, err)
	assert.NotContains(t, string(data), `"event"`)
	assert.NotContains(t, string(data), `"room"`)
}

// TestClientMessage_PingHasNoRoom проверяет сценарий: client message ping has no room.
func TestClientMessage_PingHasNoRoom(t *testing.T) {
	msg := ws.ClientMessage{Type: ws.TypePing}
	data, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.NotContains(t, string(data), `"room"`, "ping message should omit room (omitempty)")
}

// TestNewClient проверяет сценарий: new client.
func TestNewClient(t *testing.T) {
	c := ws.NewClient("user-1", "session-abc")
	assert.Equal(t, "user-1", c.UserID)
	assert.Equal(t, "session-abc", c.SessionID)
	assert.NotNil(t, ws.ExportClientSendCh(c))
	assert.NotNil(t, ws.ExportClientDone(c))
	assert.NotNil(t, ws.ExportClientRooms(c))
	assert.NotEmpty(t, c.ConnectionID, "ConnectionID должен быть присвоен при создании")
}

// TestNewClient_ConnectionIDsAreUnique проверяет, что два клиента одного
// и того же юзера получают разные ConnectionID — основа фикса HI-07.
func TestNewClient_ConnectionIDsAreUnique(t *testing.T) {
	a := ws.NewClient("user-1", "s1")
	b := ws.NewClient("user-1", "s2")
	assert.NotEqual(t, a.ConnectionID, b.ConnectionID)
}

// TestHub_SendToUser_FansOutToAllConnections — мульти-таб должен получать
// сообщения на все открытые табы одного юзера (HI-07 регрессионный тест).
func TestHub_SendToUser_FansOutToAllConnections(t *testing.T) {
	hub := ws.NewHub()

	a := ws.NewClient("user-1", "s1")
	b := ws.NewClient("user-1", "s2")

	ws.ExportRegisterClient(hub, a)
	ws.ExportRegisterClient(hub, b)

	hub.SendToUser("user-1", ws.ServerMessage{Type: ws.TypePong})

	for _, c := range []*ws.Client{a, b} {
		select {
		case <-ws.ExportClientSendCh(c):
			// ок
		default:
			t.Fatalf("client %s did not receive fan-out message", c.ConnectionID)
		}
	}
}

// TestHub_RemoveOneConnectionDoesNotAffectOther — удаление одного соединения
// не должно затрагивать другие табы того же юзера (HI-07 регрессия).
func TestHub_RemoveOneConnectionDoesNotAffectOther(t *testing.T) {
	hub := ws.NewHub()
	a := ws.NewClient("user-1", "s1")
	b := ws.NewClient("user-1", "s2")
	ws.ExportRegisterClient(hub, a)
	ws.ExportRegisterClient(hub, b)

	ws.ExportRemoveClient(hub, a)

	hub.SendToUser("user-1", ws.ServerMessage{Type: ws.TypePong})

	// b всё ещё подключён — должен получить сообщение.
	select {
	case <-ws.ExportClientSendCh(b):
		// ок
	default:
		t.Fatal("surviving connection should still receive messages after sibling close")
	}
}
