package messenger

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// SetMessageSentHook подключает post-send hook (например @hub-ai bot).
func (s *MessengerService) SetMessageSentHook(h MessageSentHook) {
	s.messageSentHook = h
}

// SetSystemBotID разрешает системному боту читать/писать без membership check.
func (s *MessengerService) SetSystemBotID(id uuid.UUID) {
	s.systemBotID = id
}

func (s *MessengerService) isSystemBot(userID uuid.UUID) bool {
	return s.systemBotID != uuid.Nil && userID == s.systemBotID
}

func (s *MessengerService) emitMessageSent(ctx context.Context, evt MessageHookEvent) {
	if s.messageSentHook == nil || s.isSystemBot(evt.SenderID) {
		return
	}
	go func() {
		hookCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if err := s.messageSentHook.OnMessageSent(hookCtx, evt); err != nil {
			slog.Error("message sent hook failed", "error", err, "channel_id", evt.ChannelID)
		}
	}()
}

// SendBotMessage отправляет сообщение от системного бота (@hub-ai).
// Не вызывает MessageSentHook — предотвращает рекурсию.
func (s *MessengerService) SendBotMessage(ctx context.Context, channelID uuid.UUID, req SendMessageRequest) (*MessageResponse, error) {
	if s.systemBotID == uuid.Nil {
		return nil, apperrors.NewInternal("system bot not configured")
	}
	doc := &MessageDocument{
		ChannelID:   channelID,
		SenderID:    s.systemBotID,
		ContentJSON: req.ContentJSON,
		ContentText: req.ContentText,
		ThreadID:    req.ThreadID,
		CreatedAt:   time.Now().UTC(),
	}
	messageID, err := s.repo.InsertMessage(ctx, doc)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка сохранения сообщения бота")
	}
	doc.ID = messageID

	if req.ThreadID != "" {
		if err := s.repo.IncrementReplyCount(ctx, req.ThreadID); err != nil {
			slog.Error("bot: increment reply_count", "error", err)
		}
	}

	if s.unread != nil {
		memberIDs, mErr := s.repo.GetChannelMemberIDs(ctx, channelID)
		if mErr == nil {
			memberStrs := make([]string, len(memberIDs))
			for i, m := range memberIDs {
				memberStrs[i] = m.String()
			}
			if uErr := s.unread.Increment(ctx, channelID.String(), memberStrs, s.systemBotID.String()); uErr != nil {
				slog.Error("bot: increment unread", "error", uErr)
			}
		}
	}

	event := newMessageSentEvent(doc, nil)
	s.nats.PublishJSON(fmt.Sprintf(SubjectMessageSent, channelID.String()), event)
	s.publishUserNewMessageNotifications(ctx, doc)

	return &event.Message, nil
}

// UpdateBotMessage обновляет сообщение бота (стриминг M3-010).
func (s *MessengerService) UpdateBotMessage(ctx context.Context, messageID, contentText string, contentJSON map[string]interface{}) error {
	if s.systemBotID == uuid.Nil {
		return apperrors.NewInternal("system bot not configured")
	}
	existing, err := s.repo.FindMessageByID(ctx, messageID)
	if err != nil {
		return apperrors.NewNotFound("сообщение не найдено")
	}
	if existing.SenderID != s.systemBotID {
		return apperrors.NewForbidden("not a bot message")
	}
	if contentJSON == nil {
		contentJSON = map[string]interface{}{
			"type": "doc",
			"children": []map[string]interface{}{
				{
					"type": "p",
					"children": []map[string]interface{}{
						{"text": contentText},
					},
				},
			},
			"hub_ai_markdown": true,
		}
	}
	if _, err := s.repo.UpdateMessage(ctx, messageID, contentJSON, contentText); err != nil {
		return apperrors.NewInternal("ошибка обновления сообщения бота")
	}
	s.nats.PublishJSON(fmt.Sprintf(SubjectMessageEdited, existing.ChannelID.String()), MessageEditedEvent{
		MessageID:   messageID,
		ChannelID:   existing.ChannelID,
		EditorID:    s.systemBotID,
		ContentText: contentText,
		EditedAt:    time.Now().UTC(),
	})
	return nil
}

func (s *MessengerService) FindMessageForBot(ctx context.Context, messageID string) (*MessageResponse, error) {
	if s.systemBotID == uuid.Nil {
		return nil, apperrors.NewInternal("system bot not configured")
	}
	doc, err := s.repo.FindMessageByID(ctx, messageID)
	if err != nil {
		return nil, apperrors.NewNotFound("сообщение не найдено")
	}
	resp := toMessageResponse(doc)
	return resp, nil
}

// GetChannelEntity возвращает сущность канала для внутренних ботов.
func (s *MessengerService) GetChannelEntity(ctx context.Context, channelID uuid.UUID) (*ent.Channel, error) {
	if s.systemBotID == uuid.Nil {
		return nil, apperrors.NewInternal("system bot not configured")
	}
	return s.repo.GetChannel(ctx, channelID)
}

// GetChannelMemberIDsInternal возвращает ID участников без auth (только system bot).
func (s *MessengerService) GetChannelMemberIDsInternal(ctx context.Context, channelID uuid.UUID) ([]uuid.UUID, error) {
	if s.systemBotID == uuid.Nil {
		return nil, apperrors.NewInternal("system bot not configured")
	}
	return s.repo.GetChannelMemberIDs(ctx, channelID)
}

// ListMessagesForBot возвращает историю без membership check (только для system bot).
func (s *MessengerService) ListMessagesForBot(ctx context.Context, channelID uuid.UUID, before *time.Time, limit int) ([]MessageResponse, error) {
	if s.systemBotID == uuid.Nil {
		return nil, apperrors.NewInternal("system bot not configured")
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

// ListThreadRepliesForBot возвращает ответы треда без membership check.
func (s *MessengerService) ListThreadRepliesForBot(ctx context.Context, messageID string, before *time.Time, limit int) ([]MessageResponse, error) {
	if s.systemBotID == uuid.Nil {
		return nil, apperrors.NewInternal("system bot not configured")
	}
	docs, err := s.repo.ListThreadReplies(ctx, messageID, before, limit)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения ответов треда")
	}
	result := make([]MessageResponse, 0, len(docs))
	for _, doc := range docs {
		result = append(result, *toMessageResponse(doc))
	}
	return result, nil
}
