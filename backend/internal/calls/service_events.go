package calls

import (
	"fmt"

	"github.com/google/uuid"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// EventPublisher emits call real-time events. Production uses
// *natsclient.Client, while isolated units can supply an in-memory capture.
type EventPublisher interface {
	PublishJSON(subject string, payload any)
}

// CallEventPublisher инкапсулирует публикацию NATS-событий звонков.
// Все subject-строки (`calls.channel.<id>.started/joined/left/ended`,
// `calls.chat.<id>`) живут только здесь — оркестратор не знает топологии
// событий (SC-14-2).
type CallEventPublisher struct {
	publisher EventPublisher
}

// NewCallEventPublisher создаёт publisher событий звонков.
// nats может быть nil — публикация в этом случае no-op (см. nats.Client.PublishJSON).
func NewCallEventPublisher(nc *natsclient.Client) *CallEventPublisher {
	return &CallEventPublisher{publisher: nc}
}

// Started публикует событие старта звонка в канале.
func (p *CallEventPublisher) Started(channelID *uuid.UUID, callID uuid.UUID, roomName string, by uuid.UUID) {
	if !validChannelID(channelID) {
		return
	}

	p.publisher.PublishJSON(fmt.Sprintf("calls.channel.%s.started", channelID.String()), map[string]interface{}{
		"call_id":    callID,
		"room_name":  roomName,
		"created_by": by,
		"started_by": by,
	})
}

// Joined публикует событие присоединения пользователя к звонку.
func (p *CallEventPublisher) Joined(channelID *uuid.UUID, callID, userID uuid.UUID, userName string) {
	if !validChannelID(channelID) {
		return
	}

	p.publisher.PublishJSON(fmt.Sprintf("calls.channel.%s.joined", channelID.String()), map[string]interface{}{
		"call_id":   callID,
		"user_id":   userID,
		"user_name": userName,
	})
}

// Left публикует событие ухода пользователя из звонка.
func (p *CallEventPublisher) Left(channelID *uuid.UUID, callID, userID uuid.UUID) {
	if !validChannelID(channelID) {
		return
	}

	p.publisher.PublishJSON(fmt.Sprintf("calls.channel.%s.left", channelID.String()), map[string]interface{}{
		"call_id": callID,
		"user_id": userID,
	})
}

// Ended публикует событие завершения звонка с финальным счётчиком участников.
func (p *CallEventPublisher) Ended(channelID *uuid.UUID, callID uuid.UUID, participantCount int) {
	if !validChannelID(channelID) {
		return
	}

	p.publisher.PublishJSON(fmt.Sprintf("calls.channel.%s.ended", channelID.String()), map[string]interface{}{
		"call_id":           callID,
		"participant_count": participantCount,
	})
}

// ChatMessage публикует сообщение чата звонка для real-time доставки через WS bridge.
func (p *CallEventPublisher) ChatMessage(callID string, msg *CallChatMessage) {
	p.publisher.PublishJSON("calls.chat."+callID, map[string]interface{}{
		"id":         msg.ID.Hex(),
		"call_id":    msg.CallID,
		"user_id":    msg.UserID,
		"user_name":  msg.UserName,
		"text":       msg.Text,
		"created_at": msg.CreatedAt,
	})
}

func validChannelID(id *uuid.UUID) bool {
	return id != nil && *id != uuid.Nil
}
