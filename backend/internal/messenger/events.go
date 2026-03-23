package messenger

import (
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// NATS subject constants
// ---------------------------------------------------------------------------

const (
	// SubjectMessageSent — тема NATS для события отправки сообщения.
	// Формат: messenger.channel.{channelID}.message_sent
	SubjectMessageSent = "messenger.channel.%s.message_sent"

	// SubjectMessageEdited — тема NATS для события редактирования сообщения.
	SubjectMessageEdited = "messenger.channel.%s.message_edited"

	// SubjectMessageDeleted — тема NATS для события удаления сообщения.
	SubjectMessageDeleted = "messenger.channel.%s.message_deleted"

	// SubjectReactionToggled — тема NATS для события переключения реакции.
	SubjectReactionToggled = "messenger.channel.%s.reaction_toggled"

	// SubjectTyping — тема NATS для индикатора набора текста.
	SubjectTyping = "messenger.channel.%s.typing"

	// SubjectMemberJoined — тема NATS для события вступления участника.
	SubjectMemberJoined = "messenger.channel.%s.member_joined"

	// SubjectMemberLeft — тема NATS для события выхода участника.
	SubjectMemberLeft = "messenger.channel.%s.member_left"

	// SubjectPresenceChanged — тема NATS для события изменения статуса присутствия.
	SubjectPresenceChanged = "messenger.presence.%s.changed"

	// SubjectPollCreated — тема NATS для события создания опроса.
	SubjectPollCreated = "messenger.channel.%s.poll_created"

	// SubjectPollVoted — тема NATS для события голосования в опросе.
	SubjectPollVoted = "messenger.channel.%s.poll_voted"

	// SubjectPollClosed — тема NATS для события закрытия опроса.
	SubjectPollClosed = "messenger.channel.%s.poll_closed"
)

// ---------------------------------------------------------------------------
// Event payloads
// ---------------------------------------------------------------------------

// MessageSentEvent — данные события отправки сообщения.
// MentionedUserIDs содержит уже разрешённые ID пользователей
// из @all, @online, @role и индивидуальных упоминаний.
type MessageSentEvent struct {
	MessageID        string    `json:"message_id"`
	ChannelID        uuid.UUID `json:"channel_id"`
	SenderID         uuid.UUID `json:"sender_id"`
	MessageType      string    `json:"message_type,omitempty"`
	PollID           string    `json:"poll_id,omitempty"`
	ContentText      string    `json:"content_text"`
	ThreadID         string    `json:"thread_id,omitempty"`
	MentionedUserIDs []string  `json:"mentioned_user_ids,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// MessageEditedEvent — данные события редактирования сообщения.
type MessageEditedEvent struct {
	MessageID   string    `json:"message_id"`
	ChannelID   uuid.UUID `json:"channel_id"`
	EditorID    uuid.UUID `json:"editor_id"`
	ContentText string    `json:"content_text"`
	EditedAt    time.Time `json:"edited_at"`
}

// MessageDeletedEvent — данные события удаления сообщения.
type MessageDeletedEvent struct {
	MessageID string    `json:"message_id"`
	ChannelID uuid.UUID `json:"channel_id"`
	DeletedBy uuid.UUID `json:"deleted_by"`
	DeletedAt time.Time `json:"deleted_at"`
}

// ReactionToggledEvent — данные события переключения реакции.
type ReactionToggledEvent struct {
	MessageID string    `json:"message_id"`
	ChannelID uuid.UUID `json:"channel_id"`
	UserID    uuid.UUID `json:"user_id"`
	Emoji     string    `json:"emoji"`
	Added     bool      `json:"added"`
}

// TypingEvent — данные события индикатора набора текста.
type TypingEvent struct {
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
}

// PollCreatedEvent — данные события создания опроса.
type PollCreatedEvent struct {
	PollID    string    `json:"poll_id"`
	ChannelID uuid.UUID `json:"channel_id"`
	CreatorID uuid.UUID `json:"creator_id"`
	Question  string    `json:"question"`
	CreatedAt time.Time `json:"created_at"`
}

// PollVotedEvent — данные события голосования в опросе.
type PollVotedEvent struct {
	PollID    string    `json:"poll_id"`
	ChannelID uuid.UUID `json:"channel_id"`
	UserID    uuid.UUID `json:"user_id"`
}

// PollClosedEvent — данные события закрытия опроса.
type PollClosedEvent struct {
	PollID    string    `json:"poll_id"`
	ChannelID uuid.UUID `json:"channel_id"`
	ClosedBy  uuid.UUID `json:"closed_by"`
}
