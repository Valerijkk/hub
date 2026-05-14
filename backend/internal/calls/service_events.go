package calls

import (
	"fmt"

	"github.com/google/uuid"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// CallEventPublisher инкапсулирует публикацию NATS-событий звонков.
// Все subject-строки (`calls.channel.<id>.started/joined/left/ended`,
// `calls.chat.<id>`) живут только здесь — оркестратор не знает топологии
// событий (SC-14-2).
type CallEventPublisher struct {
	nats *natsclient.Client
}

// NewCallEventPublisher создаёт publisher событий звонков.
// nats может быть nil — публикация в этом случае no-op (см. nats.Client.PublishJSON).
func NewCallEventPublisher(nc *natsclient.Client) *CallEventPublisher {
	return &CallEventPublisher{nats: nc}
}

// Started публикует событие старта звонка в канале.
// Для instant-звонков (channelID==nil) subject содержит "instant" вместо UUID.
func (p *CallEventPublisher) Started(channelID *uuid.UUID, callID uuid.UUID, roomName string, by uuid.UUID) {
	p.nats.PublishJSON(fmt.Sprintf("calls.channel.%s.started", channelIDStr(channelID)), map[string]interface{}{
		"call_id":    callID,
		"room_name":  roomName,
		"created_by": by,
		"started_by": by,
	})
}

// Joined публикует событие присоединения пользователя к звонку.
func (p *CallEventPublisher) Joined(channelID *uuid.UUID, callID, userID uuid.UUID, userName string) {
	p.nats.PublishJSON(fmt.Sprintf("calls.channel.%s.joined", channelIDStr(channelID)), map[string]interface{}{
		"call_id":   callID,
		"user_id":   userID,
		"user_name": userName,
	})
}

// Left публикует событие ухода пользователя из звонка.
func (p *CallEventPublisher) Left(channelID *uuid.UUID, callID, userID uuid.UUID) {
	p.nats.PublishJSON(fmt.Sprintf("calls.channel.%s.left", channelIDStr(channelID)), map[string]interface{}{
		"call_id": callID,
		"user_id": userID,
	})
}

// Ended публикует событие завершения звонка с финальным счётчиком участников.
func (p *CallEventPublisher) Ended(channelID *uuid.UUID, callID uuid.UUID, participantCount int) {
	p.nats.PublishJSON(fmt.Sprintf("calls.channel.%s.ended", channelIDStr(channelID)), map[string]interface{}{
		"call_id":           callID,
		"participant_count": participantCount,
	})
}

// ChatMessage публикует сообщение чата звонка для real-time доставки через WS bridge.
func (p *CallEventPublisher) ChatMessage(callID string, msg *CallChatMessage) {
	p.nats.PublishJSON("calls.chat."+callID, map[string]interface{}{
		"id":         msg.ID.Hex(),
		"call_id":    msg.CallID,
		"user_id":    msg.UserID,
		"user_name":  msg.UserName,
		"text":       msg.Text,
		"created_at": msg.CreatedAt,
	})
}

// channelIDStr безопасно возвращает строку из *uuid.UUID (nil → "instant").
// Используется во всех subject-строках для инкапсуляции instant-звонков.
func channelIDStr(id *uuid.UUID) string {
	if id == nil {
		return "instant"
	}
	return id.String()
}
