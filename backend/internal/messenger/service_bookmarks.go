package messenger

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// AddBookmark добавляет закладку пользователя на сообщение.
// Требуется membership в канале (P1-BE-16).
func (s *MessengerService) AddBookmark(ctx context.Context, userID uuid.UUID, messageID string, channelID uuid.UUID) error {
	isMember, err := s.repo.IsMember(ctx, channelID, userID)
	if err != nil {
		return apperrors.NewInternal("ошибка проверки участия")
	}
	if !isMember {
		return apperrors.NewForbidden("вы не являетесь участником канала")
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
