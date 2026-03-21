package messenger

import (
	"context"

	"github.com/google/uuid"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// AddBookmark adds a user bookmark on a message.
func (s *MessengerService) AddBookmark(ctx context.Context, userID uuid.UUID, messageID string, channelID uuid.UUID) error {
	_, err := s.repo.AddBookmark(ctx, userID, messageID, channelID)
	if err != nil {
		return apperrors.NewInternal("ошибка добавления закладки")
	}
	return nil
}

// RemoveBookmark removes a bookmark by ID.
func (s *MessengerService) RemoveBookmark(ctx context.Context, bookmarkID uuid.UUID) error {
	return s.repo.RemoveBookmark(ctx, bookmarkID)
}

// ListBookmarks returns all bookmarks for a user.
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
