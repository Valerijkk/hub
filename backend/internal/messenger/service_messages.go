package messenger

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// SendMessage sends a message to a channel.
// Checks membership, readonly status, increments reply_count for threads.
func (s *MessengerService) SendMessage(ctx context.Context, channelID, senderID uuid.UUID, req SendMessageRequest) (*MessageResponse, error) {
	// Check membership
	isMember, err := s.repo.IsMember(ctx, channelID, senderID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка проверки участия")
	}
	if !isMember {
		return nil, apperrors.NewForbidden("вы не являетесь участником канала")
	}

	// Check readonly
	ch, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения канала")
	}
	if ch.IsReadonly {
		// Only owner/admin can write in readonly channel
		role, _ := s.repo.GetMemberRole(ctx, channelID, senderID)
		if role != "owner" && role != "admin" {
			return nil, apperrors.NewForbidden("канал доступен только для чтения")
		}
	}

	doc := &MessageDocument{
		ChannelID:       channelID,
		SenderID:        senderID,
		ContentJSON:     req.ContentJSON,
		ContentText:     req.ContentText,
		ThreadID:        req.ThreadID,
		Attachments:     req.Attachments,
		Mentions:        req.Mentions,
		MentionGroups:   req.MentionGroups,
		VoiceURL:        req.VoiceURL,
		VoiceDurationMs: req.VoiceDurationMs,
		CreatedAt:       time.Now(),
	}

	messageID, err := s.repo.InsertMessage(ctx, doc)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка сохранения сообщения")
	}
	doc.ID = messageID

	// Increment reply_count if this is a thread reply
	if req.ThreadID != "" {
		if err := s.repo.IncrementReplyCount(ctx, req.ThreadID); err != nil {
			slog.Error("не удалось инкрементировать reply_count", "error", err)
		}
	}

	// Increment unread for channel members
	if s.unread != nil {
		memberIDs, mErr := s.repo.GetChannelMemberIDs(ctx, channelID)
		if mErr == nil {
			memberStrs := make([]string, len(memberIDs))
			for i, m := range memberIDs {
				memberStrs[i] = m.String()
			}
			if uErr := s.unread.Increment(ctx, channelID.String(), memberStrs, senderID.String()); uErr != nil {
				slog.Error("не удалось инкрементировать непрочитанные", "error", uErr)
			}
		}
	}

	// Resolve group mentions (@all, @online, @role) into concrete user IDs.
	resolvedMentions := s.resolveGroupMentions(ctx, channelID, ch.ProjectID, req.MentionGroups, req.Mentions, senderID)

	// Publish NATS event
	s.publishEvent(fmt.Sprintf(SubjectMessageSent, channelID.String()), MessageSentEvent{
		MessageID:        messageID,
		ChannelID:        channelID,
		SenderID:         senderID,
		ContentText:      req.ContentText,
		ThreadID:         req.ThreadID,
		MentionedUserIDs: resolvedMentions,
		CreatedAt:        doc.CreatedAt,
	})

	return toMessageResponse(doc), nil
}

// ListMessages returns channel messages with cursor pagination.
func (s *MessengerService) ListMessages(ctx context.Context, channelID uuid.UUID, before *time.Time, limit int) ([]MessageResponse, error) {
	docs, err := s.repo.ListMessages(ctx, channelID, before, limit)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения сообщений")
	}
	result := make([]MessageResponse, 0, len(docs))
	for _, doc := range docs {
		result = append(result, *toMessageResponse(doc))
	}
	return result, nil
}

// EditMessage edits a message (author only).
func (s *MessengerService) EditMessage(ctx context.Context, messageID string, userID uuid.UUID, req EditMessageRequest) (*MessageResponse, error) {
	existing, err := s.repo.FindMessageByID(ctx, messageID)
	if err != nil {
		return nil, apperrors.NewNotFound("сообщение не найдено")
	}

	// Check authorship
	if existing.SenderID != userID {
		return nil, apperrors.NewForbidden("можно редактировать только свои сообщения")
	}

	updated, err := s.repo.UpdateMessage(ctx, messageID, req.ContentJSON, req.ContentText)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка редактирования сообщения")
	}

	// Publish NATS event
	s.publishEvent(fmt.Sprintf(SubjectMessageEdited, existing.ChannelID.String()), MessageEditedEvent{
		MessageID:   messageID,
		ChannelID:   existing.ChannelID,
		EditorID:    userID,
		ContentText: req.ContentText,
		EditedAt:    time.Now(),
	})

	return toMessageResponse(updated), nil
}

// DeleteMessage soft-deletes a message.
// Author can delete their own; admin/owner can delete any.
func (s *MessengerService) DeleteMessage(ctx context.Context, messageID string, userID uuid.UUID, channelID string, isAdmin bool) error {
	existing, err := s.repo.FindMessageByID(ctx, messageID)
	if err != nil {
		return apperrors.NewNotFound("сообщение не найдено")
	}

	// Check authorship or admin rights
	if existing.SenderID != userID && !isAdmin {
		chID, err := uuid.Parse(channelID)
		if err != nil {
			return apperrors.NewBadRequest("некорректный channel ID")
		}
		role, _ := s.repo.GetMemberRole(ctx, chID, userID)
		if role != "owner" && role != "admin" {
			return apperrors.NewForbidden("нет прав на удаление этого сообщения")
		}
	}

	if err := s.repo.SoftDeleteMessage(ctx, messageID); err != nil {
		return apperrors.NewInternal("ошибка удаления сообщения")
	}

	// Decrement reply_count if this is a thread reply
	if existing.ThreadID != "" {
		if err := s.repo.DecrementReplyCount(ctx, existing.ThreadID); err != nil {
			slog.Error("не удалось декрементировать reply_count", "error", err)
		}
	}

	// Publish NATS event
	s.publishEvent(fmt.Sprintf(SubjectMessageDeleted, existing.ChannelID.String()), MessageDeletedEvent{
		MessageID: messageID,
		ChannelID: existing.ChannelID,
		DeletedBy: userID,
		DeletedAt: time.Now(),
	})

	return nil
}

// ListThreadReplies returns thread replies with cursor pagination.
func (s *MessengerService) ListThreadReplies(ctx context.Context, messageID string, before *time.Time, limit int) ([]MessageResponse, error) {
	docs, err := s.repo.ListThreadReplies(ctx, messageID, before, limit)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения ответов в треде")
	}
	result := make([]MessageResponse, 0, len(docs))
	for _, doc := range docs {
		result = append(result, *toMessageResponse(doc))
	}
	return result, nil
}

// ForwardMessage forwards a message to another channel.
func (s *MessengerService) ForwardMessage(ctx context.Context, channelID, userID uuid.UUID, sourceMessageID string, targetChannelID string) (*MessageResponse, error) {
	targetID, err := uuid.Parse(targetChannelID)
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный target_channel_id")
	}

	// Check membership in target channel
	isMember, err := s.repo.IsMember(ctx, targetID, userID)
	if err != nil || !isMember {
		return nil, apperrors.NewForbidden("вы не являетесь участником целевого канала")
	}

	// Get source message
	source, err := s.repo.FindMessageByID(ctx, sourceMessageID)
	if err != nil {
		return nil, apperrors.NewNotFound("исходное сообщение не найдено")
	}

	// Create new message with forwarded_from_id
	doc := &MessageDocument{
		ChannelID:       targetID,
		SenderID:        userID,
		ContentJSON:     source.ContentJSON,
		ContentText:     source.ContentText,
		ForwardedFromID: sourceMessageID,
		Attachments:     source.Attachments,
		CreatedAt:       time.Now(),
	}

	newID, err := s.repo.InsertMessage(ctx, doc)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка пересылки сообщения")
	}
	doc.ID = newID

	// NATS event to target channel
	s.publishEvent(fmt.Sprintf(SubjectMessageSent, targetID.String()), MessageSentEvent{
		MessageID:   newID,
		ChannelID:   targetID,
		SenderID:    userID,
		ContentText: source.ContentText,
		CreatedAt:   doc.CreatedAt,
	})

	return toMessageResponse(doc), nil
}
