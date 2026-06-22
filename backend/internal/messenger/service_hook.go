package messenger

import (
	"context"

	"github.com/google/uuid"
)

// MessageSentHook вызывается после успешной отправки сообщения (async-safe).
type MessageSentHook interface {
	OnMessageSent(ctx context.Context, evt MessageHookEvent) error
}

// MessageHookEvent — payload для post-send hooks (@hub-ai bot).
type MessageHookEvent struct {
	MessageID   string
	ChannelID   uuid.UUID
	SenderID    uuid.UUID
	ContentText string
	ThreadID    string
	Mentions    []string
}
