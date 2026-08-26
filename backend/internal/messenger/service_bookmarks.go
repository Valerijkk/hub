package messenger

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// Максимальная длина превью текста закладки (руны).
const bookmarkPreviewMaxRunes = 120

// AddBookmark добавляет закладку пользователя на сообщение.
// Требуется membership в канале (P1-BE-16).
// A3W3-2-003: сообщение должно принадлежать channelID (как TogglePin).
func (s *MessengerService) AddBookmark(ctx context.Context, userID uuid.UUID, messageID string, channelID uuid.UUID) error {
	_, scoped, accessErr := s.resolveChannelAction(ctx, userID, channelID, ChannelActionBookmark)
	if accessErr != nil {
		return accessErr
	}
	if !scoped {
		isMember, membershipErr := s.repo.IsMember(ctx, channelID, userID)
		if membershipErr != nil {
			return apperrors.NewInternal("ошибка проверки участия")
		}
		if !isMember {
			return apperrors.NewForbidden("вы не являетесь участником канала")
		}
	}

	msg, err := s.repo.FindMessageByID(ctx, messageID)
	if err != nil {
		return apperrors.NewNotFound("сообщение не найдено")
	}
	if msg.ChannelID != channelID {
		return apperrors.NewNotFound("сообщение не найдено")
	}

	_, err = s.repo.AddBookmark(ctx, userID, messageID, channelID)
	if err != nil {
		return apperrors.NewInternal("ошибка добавления закладки")
	}
	return nil
}

// RemoveBookmark удаляет закладку владельца по ID.
// Чужой bookmarkID → 404 (P1-BE-15).
func (s *MessengerService) RemoveBookmark(ctx context.Context, userID, bookmarkID uuid.UUID) error {
	if err := s.repo.RemoveBookmark(ctx, userID, bookmarkID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("закладка не найдена")
		}
		return apperrors.NewInternal("ошибка удаления закладки")
	}
	return nil
}

// ListBookmarks возвращает все закладки пользователя с channel_name и content_preview (A3W9-2-001).
func (s *MessengerService) ListBookmarks(ctx context.Context, userID uuid.UUID) ([]BookmarkResponse, error) {
	bookmarks, err := s.repo.ListBookmarks(ctx, userID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения закладок")
	}

	// Complete fresh PG access proof for every distinct channel before any bookmark enrichment.
	allowedChannels := make(map[uuid.UUID]bool, len(bookmarks))
	for _, b := range bookmarks {
		if _, checked := allowedChannels[b.ChannelID]; checked {
			continue
		}

		_, scoped, accessErr := s.resolveChannelAction(ctx, userID, b.ChannelID, ChannelActionReadTimeline)
		if accessErr != nil {
			allowedChannels[b.ChannelID] = false
			continue
		}
		if !scoped {
			isMember, membershipErr := s.repo.IsMember(ctx, b.ChannelID, userID)
			if membershipErr != nil || !isMember {
				allowedChannels[b.ChannelID] = false
				continue
			}
		}
		allowedChannels[b.ChannelID] = true
	}

	// Кеш имён каналов — один GetChannel на уникальный channel_id.
	channelNames := make(map[uuid.UUID]string, len(bookmarks))
	result := make([]BookmarkResponse, 0, len(bookmarks))
	for _, b := range bookmarks {
		if !allowedChannels[b.ChannelID] {
			continue
		}

		resp := BookmarkResponse{
			ID:        b.ID.String(),
			MessageID: b.MessageID,
			ChannelID: b.ChannelID.String(),
			CreatedAt: b.CreatedAt,
		}

		if name, ok := channelNames[b.ChannelID]; ok {
			resp.ChannelName = name
		} else {
			ch, chErr := s.repo.GetChannel(ctx, b.ChannelID)
			if chErr == nil && ch != nil && ch.Name != "" {
				resp.ChannelName = ch.Name
				channelNames[b.ChannelID] = ch.Name
			} else {
				channelNames[b.ChannelID] = ""
			}
		}

		if msg, msgErr := s.repo.FindMessageByID(ctx, b.MessageID); msgErr == nil && msg != nil && msg.ChannelID == b.ChannelID {
			resp.ContentPreview = truncateRunes(msg.ContentText, bookmarkPreviewMaxRunes)
		}

		result = append(result, resp)
	}
	return result, nil
}
