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

func (s *MessengerService) validateThreadParent(ctx context.Context, channelID uuid.UUID, threadID string) error {
	if threadID == "" {
		return nil
	}

	parent, err := s.repo.FindMessageByID(ctx, threadID)
	if err != nil || parent.ChannelID != channelID {
		return apperrors.NewNotFound("\u0441\u043e\u043e\u0431\u0449\u0435\u043d\u0438\u0435 \u043d\u0435 \u043d\u0430\u0439\u0434\u0435\u043d\u043e")
	}
	if parent.DeletedAt != nil {
		return apperrors.NewNotFound("\u0441\u043e\u043e\u0431\u0449\u0435\u043d\u0438\u0435 \u043d\u0435 \u043d\u0430\u0439\u0434\u0435\u043d\u043e")
	}
	if parent.ThreadID != "" {
		return apperrors.NewBadRequest("\u043e\u0442\u0432\u0435\u0442 \u0434\u043e\u043b\u0436\u0435\u043d \u0441\u0441\u044b\u043b\u0430\u0442\u044c\u0441\u044f \u043d\u0430 \u043a\u043e\u0440\u043d\u0435\u0432\u043e\u0435 \u0441\u043e\u043e\u0431\u0449\u0435\u043d\u0438\u0435 \u0442\u0440\u0435\u0434\u0430")
	}

	return nil
}

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

		requested := &MessageDocument{
			ChannelID:       channelID,
			SenderID:        senderID,
			ContentJSON:     req.ContentJSON,
			ContentText:     req.ContentText,
			ThreadID:        req.ThreadID,
			ClientNonce:     req.ClientNonce,
			Attachments:     req.Attachments,
			Mentions:        req.Mentions,
			MentionGroups:   req.MentionGroups,
			VoiceURL:        req.VoiceURL,
			VoiceDurationMs: req.VoiceDurationMs,
			CreatedAt:       time.Now().UTC(),
		}

		if requested.ClientNonce != "" {
			canonical, err := s.repo.FindMessageByClientNonce(ctx, senderID, requested.ClientNonce)
			if err != nil {
				return nil, apperrors.NewInternal("ошибка сохранения сообщения")
			}
			if canonical != nil {
				if err := validateCanonicalMessage(requested, canonical); err != nil {
					return nil, err
				}
				return toMessageResponse(canonical), nil
			}
		}

		if err := s.validateThreadParent(ctx, channelID, requested.ThreadID); err != nil {
			return nil, err
		}

		result, err := s.repo.CreateMessageIdempotent(ctx, requested)
		if err != nil || result == nil || result.Message == nil {
			return nil, apperrors.NewInternal("ошибка сохранения сообщения")
		}
		canonical := result.Message
		if err := validateCanonicalMessage(requested, canonical); err != nil {
			return nil, err
		}
		if !result.Created {
			return toMessageResponse(canonical), nil
		}

		// Инкремент reply_count, если это ответ в треде.
		if canonical.ThreadID != "" {
			if err := s.repo.IncrementReplyCount(ctx, canonical.ThreadID); err != nil {
				slog.Error("не удалось инкрементировать reply_count", "error", err)
			}
		}

		// Инкремент непрочитанных для участников канала.
		if s.unread != nil {
			memberIDs, mErr := s.repo.GetChannelMemberIDs(ctx, canonical.ChannelID)
			if mErr == nil {
				memberStrs := make([]string, len(memberIDs))
				for i, memberID := range memberIDs {
					memberStrs[i] = memberID.String()
				}
				if uErr := s.unread.Increment(ctx, canonical.ChannelID.String(), memberStrs, canonical.SenderID.String()); uErr != nil {
					slog.Error("не удалось инкрементировать непрочитанные", "error", uErr)
				}
			}
		}

		// Разрешение групповых упоминаний (@all, @online, @role) в конкретные ID пользователей.
		resolvedMentions := s.resolveGroupMentions(ctx, canonical.ChannelID, ch.ProjectID, canonical.MentionGroups, canonical.Mentions, canonical.SenderID)

		// Публикация NATS-события.
		event := newMessageSentEvent(canonical, resolvedMentions)
		s.nats.PublishJSON(fmt.Sprintf(SubjectMessageSent, canonical.ChannelID.String()), event)

		s.emitMessageSent(ctx, MessageHookEvent{
			MessageID:   canonical.ID,
			ChannelID:   canonical.ChannelID,
			SenderID:    canonical.SenderID,
			ContentText: canonical.ContentText,
			ThreadID:    canonical.ThreadID,
			Mentions:    resolvedMentions,
		})

		return &event.Message, nil
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

// EditMessage редактирует сообщение (только автор + член канала; message ∈ channel).
// W3-A1-004 / P2-BE-12: channelID из URL обязателен — нельзя править после leave / чужой канал.
func (s *MessengerService) EditMessage(ctx context.Context, channelID uuid.UUID, messageID string, userID uuid.UUID, req EditMessageRequest) (*MessageResponse, error) {
	isMember, err := s.repo.IsMember(ctx, channelID, userID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка проверки участия")
	}
	if !isMember {
		return nil, apperrors.NewForbidden("вы не являетесь участником канала")
	}

	existing, err := s.repo.FindMessageByID(ctx, messageID)
	if err != nil {
		return nil, apperrors.NewNotFound("сообщение не найдено")
	}
	if existing.ChannelID != channelID {
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
// URL-канал должен совпадать с каналом сообщения (P1-BE-17); роль — по msg.ChannelID.
func (s *MessengerService) DeleteMessage(ctx context.Context, messageID string, userID uuid.UUID, channelID string, isAdmin bool) error {
	existing, err := s.repo.FindMessageByID(ctx, messageID)
	if err != nil {
		return apperrors.NewNotFound("сообщение не найдено")
	}

	chID, err := uuid.Parse(channelID)
	if err != nil {
		return apperrors.NewBadRequest("некорректный channel ID")
	}
	if existing.ChannelID != chID {
		return apperrors.NewNotFound("сообщение не найдено")
	}

	// Проверка авторства или прав администратора (роль по каналу сообщения).
	if existing.SenderID != userID && !isAdmin {
		role, _ := s.repo.GetMemberRole(ctx, existing.ChannelID, userID)
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
	// W3-A2-018 / P3-BE-07: path channelID обязан совпадать с source.ChannelID.
	if source.ChannelID != channelID {
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

	// W3-A2-014 / P2-BE-13: тот же unread fanout, что в SendMessage.
	if s.unread != nil {
		memberIDs, mErr := s.repo.GetChannelMemberIDs(ctx, targetID)
		if mErr == nil {
			memberStrs := make([]string, len(memberIDs))
			for i, m := range memberIDs {
				memberStrs[i] = m.String()
			}
			if uErr := s.unread.Increment(ctx, targetID.String(), memberStrs, userID.String()); uErr != nil {
				slog.Error("не удалось инкрементировать непрочитанные при forward", "error", uErr)
			}
		}
	}

	// NATS-событие в целевой канал для WS-fanout.
	event := newMessageSentEvent(doc, nil)
	s.nats.PublishJSON(fmt.Sprintf(SubjectMessageSent, targetID.String()), event)

	return &event.Message, nil
}
