package messenger

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/valerii/hub/backend/internal/platform/observability"
)

// SendMessage отправляет сообщение в канал.
// Проверяет membership, readonly и инкрементирует reply_count для тредов.
func (s *MessengerService) SendMessage(ctx context.Context, channelID, senderID uuid.UUID, req SendMessageRequest) (*MessageResponse, error) {
	return observability.Timed(ctx, "messenger.SendMessage", func(ctx context.Context) (*MessageResponse, error) {
		// Проверка участия
		isMember, err := s.repo.IsMember(ctx, channelID, senderID)
		if err != nil {
			return nil, apperrors.NewInternal("ошибка проверки участия")
		}
		if !isMember {
			return nil, apperrors.NewForbidden("вы не являетесь участником канала")
		}

		// Проверка режима «только чтение»
		ch, err := s.repo.GetChannel(ctx, channelID)
		if err != nil {
			return nil, apperrors.NewInternal("ошибка получения канала")
		}
		if ch.IsReadonly {
			// В канал «только чтение» могут писать только owner/admin
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
			CreatedAt:       time.Now().UTC(),
		}

		messageID, err := s.repo.InsertMessage(ctx, doc)
		if err != nil {
			return nil, apperrors.NewInternal("ошибка сохранения сообщения")
		}
		doc.ID = messageID

		// Инкремент reply_count, если это ответ в треде
		if req.ThreadID != "" {
			if err := s.repo.IncrementReplyCount(ctx, req.ThreadID); err != nil {
				slog.Error("не удалось инкрементировать reply_count", "error", err)
			}
		}

		// Инкремент непрочитанных для участников канала
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

		// Разрешение групповых упоминаний (@all, @online, @role) в конкретные ID пользователей.
		resolvedMentions := s.resolveGroupMentions(ctx, channelID, ch.ProjectID, req.MentionGroups, req.Mentions, senderID)

		// Публикация NATS-события
		s.nats.PublishJSON(fmt.Sprintf(SubjectMessageSent, channelID.String()), MessageSentEvent{
			MessageID:        messageID,
			ChannelID:        channelID,
			SenderID:         senderID,
			ContentText:      req.ContentText,
			ThreadID:         req.ThreadID,
			MentionedUserIDs: resolvedMentions,
			CreatedAt:        doc.CreatedAt,
		})

		s.emitMessageSent(ctx, MessageHookEvent{
			MessageID:   messageID,
			ChannelID:   channelID,
			SenderID:    senderID,
			ContentText: req.ContentText,
			ThreadID:    req.ThreadID,
			Mentions:    resolvedMentions,
		})

		return toMessageResponse(doc), nil
	})
}

// ListMessages возвращает сообщения канала с cursor-пагинацией.
// HI-01: читать сообщения могут только участники канала.
func (s *MessengerService) ListMessages(ctx context.Context, channelID, userID uuid.UUID, before *time.Time, limit int) ([]MessageResponse, error) {
	isMember, err := s.repo.IsMember(ctx, channelID, userID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка проверки участия")
	}
	if !isMember {
		return nil, apperrors.NewForbidden("вы не являетесь участником канала")
	}

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

// EditMessage редактирует сообщение (только автор).
func (s *MessengerService) EditMessage(ctx context.Context, messageID string, userID uuid.UUID, req EditMessageRequest) (*MessageResponse, error) {
	existing, err := s.repo.FindMessageByID(ctx, messageID)
	if err != nil {
		return nil, apperrors.NewNotFound("сообщение не найдено")
	}

	// Проверка авторства
	if existing.SenderID != userID {
		return nil, apperrors.NewForbidden("можно редактировать только свои сообщения")
	}

	updated, err := s.repo.UpdateMessage(ctx, messageID, req.ContentJSON, req.ContentText)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка редактирования сообщения")
	}

	// Публикация NATS-события
	s.nats.PublishJSON(fmt.Sprintf(SubjectMessageEdited, existing.ChannelID.String()), MessageEditedEvent{
		MessageID:   messageID,
		ChannelID:   existing.ChannelID,
		EditorID:    userID,
		ContentText: req.ContentText,
		EditedAt:    time.Now().UTC(),
	})

	return toMessageResponse(updated), nil
}

// DeleteMessage выполняет мягкое удаление сообщения.
// Автор удаляет своё; admin/owner — любое сообщение.
func (s *MessengerService) DeleteMessage(ctx context.Context, messageID string, userID uuid.UUID, channelID string, isAdmin bool) error {
	existing, err := s.repo.FindMessageByID(ctx, messageID)
	if err != nil {
		return apperrors.NewNotFound("сообщение не найдено")
	}

	// Проверка авторства или прав администратора
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

	// Декремент reply_count, если это ответ в треде
	if existing.ThreadID != "" {
		if err := s.repo.DecrementReplyCount(ctx, existing.ThreadID); err != nil {
			slog.Error("не удалось декрементировать reply_count", "error", err)
		}
	}

	// Публикация NATS-события
	s.nats.PublishJSON(fmt.Sprintf(SubjectMessageDeleted, existing.ChannelID.String()), MessageDeletedEvent{
		MessageID: messageID,
		ChannelID: existing.ChannelID,
		DeletedBy: userID,
		DeletedAt: time.Now().UTC(),
	})

	return nil
}

// ListThreadReplies возвращает ответы треда с cursor-пагинацией.
// HI-01: ответы треда читают только участники канала. Родительское сообщение
// определяет channel_id для проверки membership — тред привязан к каналу
// родительского сообщения, а не к произвольному channel_id из запроса.
func (s *MessengerService) ListThreadReplies(ctx context.Context, messageID string, userID uuid.UUID, before *time.Time, limit int) ([]MessageResponse, error) {
	parent, err := s.repo.FindMessageByID(ctx, messageID)
	if err != nil {
		return nil, apperrors.NewNotFound("сообщение не найдено")
	}

	isMember, err := s.repo.IsMember(ctx, parent.ChannelID, userID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка проверки участия")
	}
	if !isMember {
		return nil, apperrors.NewForbidden("вы не являетесь участником канала")
	}

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

// ForwardMessage пересылает сообщение в другой канал.
func (s *MessengerService) ForwardMessage(ctx context.Context, channelID, userID uuid.UUID, sourceMessageID string, targetChannelID string) (*MessageResponse, error) {
	targetID, err := uuid.Parse(targetChannelID)
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный target_channel_id")
	}

	// Проверка membership в целевом канале — без неё forward = spam в чужой канал.
	isMember, err := s.repo.IsMember(ctx, targetID, userID)
	if err != nil || !isMember {
		return nil, apperrors.NewForbidden("вы не являетесь участником целевого канала")
	}

	// Получение исходного сообщения
	source, err := s.repo.FindMessageByID(ctx, sourceMessageID)
	if err != nil {
		return nil, apperrors.NewNotFound("исходное сообщение не найдено")
	}

	// HI-02: проверяем membership в ИСХОДНОМ канале — иначе UUID сообщения
	// позволяет любому auth-пользователю эксфильтровать контент из закрытых каналов.
	sourceIsMember, err := s.repo.IsMember(ctx, source.ChannelID, userID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка проверки участия")
	}
	if !sourceIsMember {
		return nil, apperrors.NewForbidden("вы не являетесь участником исходного канала")
	}

	// Новое сообщение с forwarded_from_id — сохраняем provenance для UI.
	doc := &MessageDocument{
		ChannelID:       targetID,
		SenderID:        userID,
		ContentJSON:     source.ContentJSON,
		ContentText:     source.ContentText,
		ForwardedFromID: sourceMessageID,
		Attachments:     source.Attachments,
		CreatedAt:       time.Now().UTC(),
	}

	newID, err := s.repo.InsertMessage(ctx, doc)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка пересылки сообщения")
	}
	doc.ID = newID

	// NATS-событие в целевой канал для WS-fanout.
	s.nats.PublishJSON(fmt.Sprintf(SubjectMessageSent, targetID.String()), MessageSentEvent{
		MessageID:   newID,
		ChannelID:   targetID,
		SenderID:    userID,
		ContentText: source.ContentText,
		CreatedAt:   doc.CreatedAt,
	})

	return toMessageResponse(doc), nil
}
