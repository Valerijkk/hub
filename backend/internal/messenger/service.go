package messenger

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// MessengerService реализует бизнес-логику модуля мессенджера.
// Оркестрирует операции с каналами, сообщениями, реакциями,
// публикует NATS-события для real-time доставки.
type MessengerService struct {
	repo     Repository
	nats     *natsclient.Client
	presence *PresenceService
	unread   *UnreadService
	db       *ent.Client
}

// NewMessengerService создаёт новый сервис мессенджера.
func NewMessengerService(repo Repository, nats *natsclient.Client, presence *PresenceService, unread *UnreadService) *MessengerService {
	return &MessengerService{repo: repo, nats: nats, presence: presence, unread: unread}
}

// SetDB устанавливает Ent-клиент для разрешения групповых упоминаний (@role).
func (s *MessengerService) SetDB(db *ent.Client) {
	s.db = db
}

// ---------------------------------------------------------------------------
// Каналы
// ---------------------------------------------------------------------------

// CreateChannel создаёт новый канал и добавляет создателя как owner.
func (s *MessengerService) CreateChannel(ctx context.Context, userID uuid.UUID, req CreateChannelRequest) (*ChannelResponse, error) {
	input := CreateChannelInput{
		ChannelType: req.ChannelType,
		Name:        req.Name,
		Description: req.Description,
		IsReadonly:  req.IsReadonly,
		CreatedBy:   userID,
	}

	if req.ProjectID != "" {
		pid, err := uuid.Parse(req.ProjectID)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный project_id")
		}
		input.ProjectID = &pid
	}

	ch, err := s.repo.CreateChannel(ctx, input)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания канала")
	}

	// Добавляем создателя как owner
	if err := s.repo.AddMembers(ctx, ch.ID, []uuid.UUID{userID}, "owner"); err != nil {
		slog.Error("не удалось добавить создателя в канал", "error", err)
	}

	// Добавляем дополнительных участников
	if len(req.MemberIDs) > 0 {
		memberUUIDs := make([]uuid.UUID, 0, len(req.MemberIDs))
		for _, mid := range req.MemberIDs {
			uid, err := uuid.Parse(mid)
			if err != nil {
				continue
			}
			if uid != userID {
				memberUUIDs = append(memberUUIDs, uid)
			}
		}
		if len(memberUUIDs) > 0 {
			if err := s.repo.AddMembers(ctx, ch.ID, memberUUIDs, "member"); err != nil {
				slog.Error("не удалось добавить участников в канал", "error", err)
			}
		}
	}

	return toChannelResponse(ch), nil
}

// GetChannel возвращает канал по ID.
func (s *MessengerService) GetChannel(ctx context.Context, channelID uuid.UUID) (*ChannelResponse, error) {
	ch, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("канал не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения канала")
	}
	return toChannelResponse(ch), nil
}

// ListMyChannels возвращает все каналы пользователя.
func (s *MessengerService) ListMyChannels(ctx context.Context, userID uuid.UUID) ([]ChannelResponse, error) {
	channels, err := s.repo.ListUserChannels(ctx, userID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения каналов")
	}
	result := make([]ChannelResponse, 0, len(channels))
	for _, ch := range channels {
		result = append(result, *toChannelResponse(ch))
	}
	return result, nil
}

// UpdateChannel обновляет канал с проверкой роли owner/admin.
func (s *MessengerService) UpdateChannel(ctx context.Context, channelID, userID uuid.UUID, req CreateChannelRequest) (*ChannelResponse, error) {
	role, err := s.repo.GetMemberRole(ctx, channelID, userID)
	if err != nil {
		return nil, apperrors.NewForbidden("нет доступа к каналу")
	}
	if role != "owner" && role != "admin" {
		return nil, apperrors.NewForbidden("только owner или admin может редактировать канал")
	}

	input := UpdateChannelInput{}
	if req.Name != "" {
		input.Name = &req.Name
	}
	if req.Description != "" {
		input.Description = &req.Description
	}
	input.IsReadonly = &req.IsReadonly

	ch, err := s.repo.UpdateChannel(ctx, channelID, input)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка обновления канала")
	}
	return toChannelResponse(ch), nil
}

// DeleteChannel удаляет канал с проверкой роли owner.
func (s *MessengerService) DeleteChannel(ctx context.Context, channelID, userID uuid.UUID) error {
	role, err := s.repo.GetMemberRole(ctx, channelID, userID)
	if err != nil {
		return apperrors.NewForbidden("нет доступа к каналу")
	}
	if role != "owner" {
		return apperrors.NewForbidden("только owner может удалить канал")
	}
	return s.repo.DeleteChannel(ctx, channelID)
}

// JoinChannel позволяет присоединиться к публичному каналу.
func (s *MessengerService) JoinChannel(ctx context.Context, channelID, userID uuid.UUID) error {
	ch, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("канал не найден")
		}
		return apperrors.NewInternal("ошибка получения канала")
	}
	switch ch.ChannelType.String() {
	case "public":
		// Публичный канал — свободное присоединение
	case "private":
		return apperrors.NewForbidden("канал приватный, требуется приглашение")
	case "dm":
		return apperrors.NewForbidden("нельзя присоединиться к DM, используйте CreateOrGetDM")
	default:
		return apperrors.NewForbidden("можно присоединиться только к публичному каналу")
	}

	// Проверяем, не является ли уже участником
	isMember, _ := s.repo.IsMember(ctx, channelID, userID)
	if isMember {
		return nil
	}

	if err := s.repo.AddMembers(ctx, channelID, []uuid.UUID{userID}, "member"); err != nil {
		return apperrors.NewInternal("ошибка вступления в канал")
	}

	// Публикуем NATS-событие
	s.publishEvent(fmt.Sprintf(SubjectMemberJoined, channelID.String()), map[string]interface{}{
		"channel_id": channelID,
		"user_id":    userID,
	})

	return nil
}

// LeaveChannel позволяет покинуть канал.
func (s *MessengerService) LeaveChannel(ctx context.Context, channelID, userID uuid.UUID) error {
	if err := s.repo.RemoveMember(ctx, channelID, userID); err != nil {
		return apperrors.NewInternal("ошибка выхода из канала")
	}

	s.publishEvent(fmt.Sprintf(SubjectMemberLeft, channelID.String()), map[string]interface{}{
		"channel_id": channelID,
		"user_id":    userID,
	})

	return nil
}

// AddMembers добавляет участников в канал с проверкой роли admin/owner.
func (s *MessengerService) AddMembers(ctx context.Context, channelID, userID uuid.UUID, memberIDs []string) error {
	role, err := s.repo.GetMemberRole(ctx, channelID, userID)
	if err != nil {
		return apperrors.NewForbidden("нет доступа к каналу")
	}
	if role != "owner" && role != "admin" {
		return apperrors.NewForbidden("только owner или admin может добавлять участников")
	}

	uuids := make([]uuid.UUID, 0, len(memberIDs))
	for _, mid := range memberIDs {
		uid, err := uuid.Parse(mid)
		if err != nil {
			return apperrors.NewBadRequest("некорректный member ID: " + mid)
		}
		uuids = append(uuids, uid)
	}

	return s.repo.AddMembers(ctx, channelID, uuids, "member")
}

// AddMembersInternal добавляет участников в канал без проверки прав.
// Используется внутренними сервисами (например, AutoChannelService).
func (s *MessengerService) AddMembersInternal(ctx context.Context, channelID uuid.UUID, memberIDs []uuid.UUID) error {
	return s.repo.AddMembers(ctx, channelID, memberIDs, "member")
}

// ListMembers возвращает всех участников канала.
func (s *MessengerService) ListMembers(ctx context.Context, channelID uuid.UUID) ([]ChannelMemberResponse, error) {
	members, err := s.repo.ListMembers(ctx, channelID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения участников")
	}
	result := make([]ChannelMemberResponse, 0, len(members))
	for _, m := range members {
		resp := ChannelMemberResponse{
			ID:       m.ID.String(),
			UserID:   m.UserID.String(),
			Role:     string(m.Role),
			JoinedAt: m.JoinedAt,
		}
		if m.Edges.User != nil {
			resp.UserName = m.Edges.User.Name
		}
		result = append(result, resp)
	}
	return result, nil
}

// CreateOrGetDM создаёт или возвращает существующий DM-канал.
// Дедупликация через отсортированные user IDs в dm_key.
func (s *MessengerService) CreateOrGetDM(ctx context.Context, userID, targetUserID uuid.UUID) (*ChannelResponse, error) {
	if userID == targetUserID {
		return nil, apperrors.NewBadRequest("нельзя создать DM с самим собой")
	}

	// Формируем dm_key из отсортированных ID
	ids := []string{userID.String(), targetUserID.String()}
	sort.Strings(ids)
	dmKey := ids[0] + "-" + ids[1]

	// Ищем существующий DM
	existing, err := s.repo.FindDMChannel(ctx, dmKey)
	if err == nil {
		return toChannelResponse(existing), nil
	}

	// Создаём новый DM
	input := CreateChannelInput{
		ChannelType: "dm",
		CreatedBy:   userID,
		DMKey:       &dmKey,
	}
	ch, err := s.repo.CreateChannel(ctx, input)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания DM канала")
	}

	// Добавляем обоих участников
	if err := s.repo.AddMembers(ctx, ch.ID, []uuid.UUID{userID, targetUserID}, "member"); err != nil {
		slog.Error("не удалось добавить участников DM", "error", err)
	}

	return toChannelResponse(ch), nil
}

// ---------------------------------------------------------------------------
// Сообщения
// ---------------------------------------------------------------------------

// SendMessage отправляет сообщение в канал.
// Проверяет membership, readonly-статус, инкрементирует reply_count для тредов.
func (s *MessengerService) SendMessage(ctx context.Context, channelID, senderID uuid.UUID, req SendMessageRequest) (*MessageResponse, error) {
	// Проверка membership
	isMember, err := s.repo.IsMember(ctx, channelID, senderID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка проверки участия")
	}
	if !isMember {
		return nil, apperrors.NewForbidden("вы не являетесь участником канала")
	}

	// Проверка readonly
	ch, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения канала")
	}
	if ch.IsReadonly {
		// Только owner/admin может писать в readonly канал
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

	// Инкрементируем reply_count если это ответ в тред
	if req.ThreadID != "" {
		if err := s.repo.IncrementReplyCount(ctx, req.ThreadID); err != nil {
			slog.Error("не удалось инкрементировать reply_count", "error", err)
		}
	}

	// Инкрементируем непрочитанные для участников канала
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

	// Разрешаем групповые упоминания (@all, @online, @role) в конкретные user ID.
	resolvedMentions := s.resolveGroupMentions(ctx, channelID, ch.ProjectID, req.MentionGroups, req.Mentions, senderID)

	// Публикуем NATS-событие
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

// ListMessages возвращает сообщения канала с курсорной пагинацией.
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

// EditMessage редактирует сообщение с проверкой авторства.
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

	// Публикуем NATS-событие
	s.publishEvent(fmt.Sprintf(SubjectMessageEdited, existing.ChannelID.String()), MessageEditedEvent{
		MessageID:   messageID,
		ChannelID:   existing.ChannelID,
		EditorID:    userID,
		ContentText: req.ContentText,
		EditedAt:    time.Now(),
	})

	return toMessageResponse(updated), nil
}

// DeleteMessage удаляет сообщение (мягкое удаление).
// Автор может удалить своё, admin/owner — любое.
func (s *MessengerService) DeleteMessage(ctx context.Context, messageID string, userID uuid.UUID, channelID string, isAdmin bool) error {
	existing, err := s.repo.FindMessageByID(ctx, messageID)
	if err != nil {
		return apperrors.NewNotFound("сообщение не найдено")
	}

	// Проверка авторства или админских прав
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

	// Декрементируем reply_count если это ответ в треде
	if existing.ThreadID != "" {
		if err := s.repo.DecrementReplyCount(ctx, existing.ThreadID); err != nil {
			slog.Error("не удалось декрементировать reply_count", "error", err)
		}
	}

	// Публикуем NATS-событие
	s.publishEvent(fmt.Sprintf(SubjectMessageDeleted, existing.ChannelID.String()), MessageDeletedEvent{
		MessageID: messageID,
		ChannelID: existing.ChannelID,
		DeletedBy: userID,
		DeletedAt: time.Now(),
	})

	return nil
}

// ToggleReaction добавляет или удаляет реакцию (toggle).
// Возвращает true если реакция добавлена, false если удалена.
func (s *MessengerService) ToggleReaction(ctx context.Context, messageID string, userID uuid.UUID, emoji string) (bool, error) {
	exists, err := s.repo.FindReaction(ctx, messageID, userID.String(), emoji)
	if err != nil {
		return false, apperrors.NewInternal("ошибка проверки реакции")
	}

	msg, err := s.repo.FindMessageByID(ctx, messageID)
	if err != nil {
		return false, apperrors.NewNotFound("сообщение не найдено")
	}

	if exists {
		if err := s.repo.RemoveReaction(ctx, messageID, userID.String(), emoji); err != nil {
			return false, apperrors.NewInternal("ошибка удаления реакции")
		}
	} else {
		if err := s.repo.AddReaction(ctx, messageID, userID.String(), emoji); err != nil {
			return false, apperrors.NewInternal("ошибка добавления реакции")
		}
	}

	added := !exists
	s.publishEvent(fmt.Sprintf(SubjectReactionToggled, msg.ChannelID.String()), ReactionToggledEvent{
		MessageID: messageID,
		ChannelID: msg.ChannelID,
		UserID:    userID,
		Emoji:     emoji,
		Added:     added,
	})

	return added, nil
}

// ListThreadReplies возвращает ответы в треде с курсорной пагинацией.
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

// TogglePin закрепляет или открепляет сообщение.
// Возвращает true если закреплено, false если откреплено.
func (s *MessengerService) TogglePin(ctx context.Context, channelID uuid.UUID, messageID string, userID uuid.UUID) (bool, error) {
	existing, err := s.repo.FindPin(ctx, channelID, messageID)
	if err == nil && existing != nil {
		// Откреплять
		if err := s.repo.RemovePin(ctx, channelID, messageID); err != nil {
			return false, apperrors.NewInternal("ошибка открепления сообщения")
		}
		return false, nil
	}

	// Закреплять
	if err := s.repo.AddPin(ctx, channelID, messageID, userID); err != nil {
		return false, apperrors.NewInternal("ошибка закрепления сообщения")
	}
	return true, nil
}

// ForwardMessage пересылает сообщение в другой канал.
func (s *MessengerService) ForwardMessage(ctx context.Context, channelID, userID uuid.UUID, sourceMessageID string, targetChannelID string) (*MessageResponse, error) {
	targetID, err := uuid.Parse(targetChannelID)
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный target_channel_id")
	}

	// Проверяем membership в целевом канале
	isMember, err := s.repo.IsMember(ctx, targetID, userID)
	if err != nil || !isMember {
		return nil, apperrors.NewForbidden("вы не являетесь участником целевого канала")
	}

	// Получаем исходное сообщение
	source, err := s.repo.FindMessageByID(ctx, sourceMessageID)
	if err != nil {
		return nil, apperrors.NewNotFound("исходное сообщение не найдено")
	}

	// Создаём новое сообщение с forwarded_from_id
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

	// NATS-событие в целевой канал
	s.publishEvent(fmt.Sprintf(SubjectMessageSent, targetID.String()), MessageSentEvent{
		MessageID:   newID,
		ChannelID:   targetID,
		SenderID:    userID,
		ContentText: source.ContentText,
		CreatedAt:   doc.CreatedAt,
	})

	return toMessageResponse(doc), nil
}

// SearchMessages выполняет полнотекстовый поиск по сообщениям.
func (s *MessengerService) SearchMessages(ctx context.Context, query, channelID string, limit int) ([]MessageResponse, error) {
	var chID *uuid.UUID
	if channelID != "" {
		parsed, err := uuid.Parse(channelID)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный channel_id")
		}
		chID = &parsed
	}

	docs, err := s.repo.SearchMessages(ctx, query, chID, limit)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка поиска сообщений")
	}
	result := make([]MessageResponse, 0, len(docs))
	for _, doc := range docs {
		result = append(result, *toMessageResponse(doc))
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Закреплённые сообщения
// ---------------------------------------------------------------------------

// ListPins возвращает все закреплённые сообщения канала.
func (s *MessengerService) ListPins(ctx context.Context, channelID uuid.UUID) ([]PinnedMessageResponse, error) {
	pins, err := s.repo.ListPins(ctx, channelID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения закреплённых сообщений")
	}
	result := make([]PinnedMessageResponse, 0, len(pins))
	for _, p := range pins {
		result = append(result, PinnedMessageResponse{
			ID:        p.ID.String(),
			MessageID: p.MessageID,
			ChannelID: p.ChannelID.String(),
			PinnedBy:  p.PinnedBy.String(),
			PinnedAt:  p.PinnedAt,
		})
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Закладки
// ---------------------------------------------------------------------------

// ListBookmarks возвращает все закладки пользователя.
func (s *MessengerService) ListBookmarks(ctx context.Context, userID uuid.UUID) ([]BookmarkResponse, error) {
	bookmarks, err := s.repo.ListBookmarks(ctx, userID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения закладок")
	}
	result := make([]BookmarkResponse, 0, len(bookmarks))
	for _, b := range bookmarks {
		result = append(result, BookmarkResponse{
			ID:        b.ID.String(),
			MessageID: b.MessageID,
			ChannelID: b.ChannelID.String(),
			CreatedAt: b.CreatedAt,
		})
	}
	return result, nil
}

// AddBookmark добавляет закладку пользователя на сообщение.
func (s *MessengerService) AddBookmark(ctx context.Context, userID uuid.UUID, messageID string, channelID uuid.UUID) error {
	_, err := s.repo.AddBookmark(ctx, userID, messageID, channelID)
	if err != nil {
		return apperrors.NewInternal("ошибка добавления закладки")
	}
	return nil
}

// RemoveBookmark удаляет закладку по ID.
func (s *MessengerService) RemoveBookmark(ctx context.Context, bookmarkID uuid.UUID) error {
	return s.repo.RemoveBookmark(ctx, bookmarkID)
}

// ---------------------------------------------------------------------------
// Присутствие (Presence)
// ---------------------------------------------------------------------------

// Heartbeat обновляет время активности пользователя.
func (s *MessengerService) Heartbeat(ctx context.Context, userID string) error {
	if s.presence == nil {
		return nil
	}
	return s.presence.Heartbeat(ctx, userID)
}

// GetPresence возвращает статусы присутствия для списка пользователей.
func (s *MessengerService) GetPresence(ctx context.Context, userIDs []string) (map[string]string, error) {
	if s.presence == nil {
		return map[string]string{}, nil
	}
	return s.presence.GetBulkStatus(ctx, userIDs)
}

// SetCustomStatus устанавливает кастомный текст статуса пользователя.
func (s *MessengerService) SetCustomStatus(ctx context.Context, userID, text string) error {
	if s.presence == nil {
		return nil
	}
	return s.presence.SetCustomStatus(ctx, userID, text)
}

// GetCustomStatus возвращает кастомный текст статуса пользователя.
func (s *MessengerService) GetCustomStatus(ctx context.Context, userID string) (string, error) {
	if s.presence == nil {
		return "", nil
	}
	return s.presence.GetCustomStatus(ctx, userID)
}

// ---------------------------------------------------------------------------
// Непрочитанные (Unread)
// ---------------------------------------------------------------------------

// GetUnreadCounts возвращает счётчики непрочитанных для пользователя.
func (s *MessengerService) GetUnreadCounts(ctx context.Context, userID string) (map[string]int64, error) {
	if s.unread == nil {
		return map[string]int64{}, nil
	}
	return s.unread.GetAll(ctx, userID)
}

// MarkChannelRead сбрасывает счётчик непрочитанных для канала.
func (s *MessengerService) MarkChannelRead(ctx context.Context, userID, channelID string) error {
	if s.unread == nil {
		return nil
	}
	return s.unread.MarkRead(ctx, userID, channelID)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// resolveGroupMentions разрешает групповые упоминания (@all, @online, @role)
// в конкретные user ID. Объединяет с индивидуальными упоминаниями из поля Mentions.
// Исключает отправителя из результата.
func (s *MessengerService) resolveGroupMentions(ctx context.Context, channelID uuid.UUID, projectID *uuid.UUID, groups []string, individualMentions []string, senderID uuid.UUID) []string {
	seen := make(map[string]struct{})
	senderStr := senderID.String()

	// Добавляем индивидуальные упоминания.
	for _, uid := range individualMentions {
		if uid != senderStr {
			seen[uid] = struct{}{}
		}
	}

	if len(groups) == 0 {
		return mapKeys(seen)
	}

	// Получаем участников канала (нужно для @all и пересечения с @online).
	memberIDs, err := s.repo.GetChannelMemberIDs(ctx, channelID)
	if err != nil {
		slog.Error("не удалось получить участников канала для упоминаний", "error", err)
		return mapKeys(seen)
	}

	memberSet := make(map[string]struct{}, len(memberIDs))
	for _, m := range memberIDs {
		memberSet[m.String()] = struct{}{}
	}

	for _, g := range groups {
		switch g {
		case "all":
			// @all — все участники канала.
			for _, m := range memberIDs {
				uid := m.String()
				if uid != senderStr {
					seen[uid] = struct{}{}
				}
			}

		case "online":
			// @online — онлайн-пользователи, пересечённые с участниками канала.
			if s.presence == nil {
				continue
			}
			onlineIDs, pErr := s.presence.GetOnlineUserIDs(ctx)
			if pErr != nil {
				slog.Error("не удалось получить онлайн-пользователей", "error", pErr)
				continue
			}
			for _, oid := range onlineIDs {
				if _, ok := memberSet[oid]; ok && oid != senderStr {
					seen[oid] = struct{}{}
				}
			}

		default:
			// Предполагаем имя роли проекта (owner, developer, tester, viewer).
			if s.db == nil || projectID == nil {
				continue
			}
			members, qErr := s.db.ProjectMember.Query().
				Where(
					projectmember.HasProjectWith(project.IDEQ(*projectID)),
					projectmember.ProjectRoleEQ(projectmember.ProjectRole(g)),
				).
				WithUser().
				All(ctx)
			if qErr != nil {
				slog.Error("не удалось получить участников по роли", "role", g, "error", qErr)
				continue
			}
			for _, m := range members {
				if m.Edges.User == nil {
					continue
				}
				uid := m.Edges.User.ID.String()
				// Учитываем только тех, кто является участником канала.
				if _, ok := memberSet[uid]; ok && uid != senderStr {
					seen[uid] = struct{}{}
				}
			}
		}
	}

	return mapKeys(seen)
}

// mapKeys возвращает ключи map[string]struct{} как срез строк.
func mapKeys(m map[string]struct{}) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

// publishEvent публикует событие в NATS (fire-and-forget).
// Безопасен при nil NATS-клиенте (пропускает публикацию).
func (s *MessengerService) publishEvent(subject string, payload interface{}) {
	if s.nats == nil {
		return
	}
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("не удалось сериализовать событие", "subject", subject, "error", err)
		return
	}
	if err := s.nats.Publish(subject, data); err != nil {
		slog.Error("не удалось опубликовать событие", "subject", subject, "error", err)
	}
}

// toChannelResponse конвертирует Ent Channel в DTO ответа.
func toChannelResponse(ch *ent.Channel) *ChannelResponse {
	resp := &ChannelResponse{
		ID:          ch.ID.String(),
		ChannelType: ch.ChannelType.String(),
		Name:        ch.Name,
		Description: ch.Description,
		IsReadonly:  ch.IsReadonly,
		CreatedBy:   ch.CreatedBy.String(),
		CreatedAt:   ch.CreatedAt,
		UpdatedAt:   ch.UpdatedAt,
	}
	if ch.ProjectID != nil {
		resp.ProjectID = ch.ProjectID.String()
	}
	return resp
}
