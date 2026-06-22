package ai

import (
	"github.com/google/uuid"
	wsmod "github.com/valerii/hub/backend/internal/ws"
)

// StreamPublisher — AI stream events через WebSocket.
type StreamPublisher interface {
	PublishStreamEvent(userID uuid.UUID, streamID, event string, payload map[string]interface{})
}

// WSStreamPublisher — ai.stream.* через ws.Hub.SendToUser.
type WSStreamPublisher struct {
	hub *wsmod.Hub
}

// NewWSStreamPublisher — stream publisher через WebSocket.
func NewWSStreamPublisher(hub *wsmod.Hub) *WSStreamPublisher {
	return &WSStreamPublisher{hub: hub}
}

// PublishStreamEvent — event во все tabs пользователя.
func (p *WSStreamPublisher) PublishStreamEvent(userID uuid.UUID, streamID, event string, payload map[string]interface{}) {
	if p == nil || p.hub == nil {
		return
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	payload["stream_id"] = streamID
	p.hub.SendToUser(userID.String(), wsmod.ServerMessage{
		Type:    wsmod.TypeEvent,
		Event:   "ai.stream." + event,
		Payload: payload,
	})
}
