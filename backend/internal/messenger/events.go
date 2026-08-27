package messenger

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// EventPublisher emits messenger real-time events. Production uses
// *natsclient.Client, while isolated units can supply an in-memory capture.
type EventPublisher interface {
	PublishJSON(subject string, payload any)
}

// ---------------------------------------------------------------------------
// NATS-субъекты событий мессенджера.
// ---------------------------------------------------------------------------

const (
	// UserChannelCreatedEvent is the user-scoped lifecycle event sent after a
	// channel becomes visible to a member.
	UserChannelCreatedEvent = "channel_created"

	// UserChannelUpdatedEvent is the user-scoped lifecycle event sent when a
	// current member must refresh a channel after an update.
	UserChannelUpdatedEvent = "channel_updated"

	// UserChannelRemovedEvent is the user-scoped lifecycle event sent when a
	// channel is no longer visible to a member.
	UserChannelRemovedEvent = "channel_removed"

	// UserNewMessageNotificationEvent is the user-scoped notification sent to
	// channel members other than the canonical message sender.
	UserNewMessageNotificationEvent = "new_message_notification"

	// UserUnreadChangedEvent is an identifier-only prompt to refetch durable
	// unread state from the authoritative API.
	UserUnreadChangedEvent = "unread_changed"

	// SubjectUserChannelCreated is the per-user NATS subject for a channel that
	// has just become visible. Format: messenger.user.{userID}.channel_created.
	SubjectUserChannelCreated = "messenger.user.%s." + UserChannelCreatedEvent

	// SubjectUserChannelUpdated is the per-user NATS subject for a channel
	// update. Format: messenger.user.{userID}.channel_updated.
	SubjectUserChannelUpdated = "messenger.user.%s." + UserChannelUpdatedEvent

	// SubjectUserChannelRemoved is the per-user NATS subject for a channel that
	// is no longer visible. Format: messenger.user.{userID}.channel_removed.
	SubjectUserChannelRemoved = "messenger.user.%s." + UserChannelRemovedEvent

	// SubjectUserNewMessageNotification is the per-user NATS subject for a
	// channel message notification. Format:
	// messenger.user.{userID}.new_message_notification.
	SubjectUserNewMessageNotification = "messenger.user.%s." + UserNewMessageNotificationEvent

	// SubjectUserUnreadChanged is the per-user durable-unread invalidation.
	// Format: messenger.user.{userID}.unread_changed.
	SubjectUserUnreadChanged = "messenger.user.%s." + UserUnreadChangedEvent
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
// Полезная нагрузка событий
// ---------------------------------------------------------------------------

// ChannelLifecycleEvent intentionally contains only the stable channel ID.
// The recipient reloads the authoritative, recipient-specific channel view.
type ChannelLifecycleEvent struct {
	ChannelID string `json:"channel_id"`
}

// NewMessageNotificationEvent intentionally contains no message content or
// recipient-local state. It only prompts the recipient to reload its channel
// view from the authoritative source.
type NewMessageNotificationEvent struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
}

// UnreadChangedEvent contains only routing identifiers. It never carries
// content, a counter, or recipient-local unread state.
type UnreadChangedEvent struct {
	ChannelID     uuid.UUID  `json:"channel_id"`
	TargetKind    string     `json:"target_kind"`
	Category      string     `json:"category"`
	RootMessageID *uuid.UUID `json:"root_message_id,omitempty"`
}

// publishUserChannelLifecycleEvent best-effort publishes a single
// invalidation payload to each explicit, valid user ID exactly once.
func (s *MessengerService) publishUserChannelLifecycleEvent(subjectFormat string, channelID uuid.UUID, userIDs []uuid.UUID) {
	if channelID == uuid.Nil {
		return
	}

	payload := ChannelLifecycleEvent{ChannelID: channelID.String()}
	seen := make(map[uuid.UUID]struct{}, len(userIDs))
	for _, userID := range userIDs {
		if userID == uuid.Nil {
			continue
		}
		if _, duplicate := seen[userID]; duplicate {
			continue
		}
		seen[userID] = struct{}{}
		s.nats.PublishJSON(fmt.Sprintf(subjectFormat, userID.String()), payload)
	}
}

// publishUserNewMessageNotifications best-effort notifies the channel's
// current members, excluding the canonical durable message sender. A member
// lookup failure is logged and deliberately does not affect the primary write.
func (s *MessengerService) publishUserNewMessageNotifications(ctx context.Context, doc *MessageDocument) {
	if doc == nil || doc.ChannelID == uuid.Nil || doc.ID == "" || s.repo == nil {
		return
	}

	memberIDs, err := s.repo.GetChannelMemberIDs(ctx, doc.ChannelID)
	if err != nil {
		slog.Error("не удалось получить участников канала для уведомления о новом сообщении", "error", err, "channel_id", doc.ChannelID)
		return
	}

	payload := NewMessageNotificationEvent{
		ChannelID: doc.ChannelID.String(),
		MessageID: doc.ID,
	}
	seen := make(map[uuid.UUID]struct{}, len(memberIDs))
	for _, userID := range memberIDs {
		if userID == uuid.Nil || userID == doc.SenderID {
			continue
		}
		if _, duplicate := seen[userID]; duplicate {
			continue
		}
		seen[userID] = struct{}{}
		s.nats.PublishJSON(fmt.Sprintf(SubjectUserNewMessageNotification, userID.String()), payload)
	}
}

// MessageSentEvent — данные события отправки сообщения.
// MentionedUserIDs содержит уже разрешённые ID пользователей
// из @all, @online, @role и индивидуальных упоминаний.
type MessageSentEvent struct {
	Message          MessageResponse `json:"message"`
	MentionedUserIDs []string        `json:"mentioned_user_ids,omitempty"`
}

func newMessageSentEvent(doc *MessageDocument, mentionedUserIDs []string) MessageSentEvent {
	return MessageSentEvent{
		Message:          *toMessageResponse(doc),
		MentionedUserIDs: mentionedUserIDs,
	}
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
