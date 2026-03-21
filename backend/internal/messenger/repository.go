package messenger

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

// Repository определяет интерфейс доступа к данным модуля мессенджера.
// Разделён на две реализации: EntRepository (PG) и MongoRepository (MongoDB).
type Repository interface {
	// --- Каналы (Ent/PG) ---

	// CreateChannel создаёт новый канал.
	CreateChannel(ctx context.Context, input CreateChannelInput) (*ent.Channel, error)
	// GetChannel возвращает канал по ID с загрузкой участников.
	GetChannel(ctx context.Context, channelID uuid.UUID) (*ent.Channel, error)
	// ListUserChannels возвращает все каналы пользователя.
	ListUserChannels(ctx context.Context, userID uuid.UUID) ([]*ent.Channel, error)
	// UpdateChannel обновляет канал.
	UpdateChannel(ctx context.Context, channelID uuid.UUID, input UpdateChannelInput) (*ent.Channel, error)
	// DeleteChannel удаляет канал.
	DeleteChannel(ctx context.Context, channelID uuid.UUID) error
	// IsMember проверяет, является ли пользователь участником канала.
	IsMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error)
	// GetMemberRole возвращает роль участника в канале.
	GetMemberRole(ctx context.Context, channelID, userID uuid.UUID) (string, error)
	// GetChannelMemberIDs возвращает ID всех участников канала.
	GetChannelMemberIDs(ctx context.Context, channelID uuid.UUID) ([]uuid.UUID, error)
	// AddMembers добавляет участников в канал.
	AddMembers(ctx context.Context, channelID uuid.UUID, userIDs []uuid.UUID, role string) error
	// RemoveMember удаляет участника из канала.
	RemoveMember(ctx context.Context, channelID, userID uuid.UUID) error
	// FindDMChannel ищет существующий DM-канал по ключу дедупликации.
	FindDMChannel(ctx context.Context, dmKey string) (*ent.Channel, error)
	// ListMembers возвращает всех участников канала.
	ListMembers(ctx context.Context, channelID uuid.UUID) ([]*ent.ChannelMember, error)

	// --- Сообщения (MongoDB) ---

	// InsertMessage сохраняет сообщение в MongoDB.
	InsertMessage(ctx context.Context, doc *MessageDocument) (string, error)
	// FindMessageByID возвращает сообщение по ID.
	FindMessageByID(ctx context.Context, messageID string) (*MessageDocument, error)
	// ListMessages возвращает сообщения канала с курсорной пагинацией.
	ListMessages(ctx context.Context, channelID uuid.UUID, before *time.Time, limit int) ([]*MessageDocument, error)
	// UpdateMessage обновляет контент сообщения.
	UpdateMessage(ctx context.Context, messageID string, contentJSON map[string]interface{}, contentText string) (*MessageDocument, error)
	// SoftDeleteMessage выполняет мягкое удаление сообщения.
	SoftDeleteMessage(ctx context.Context, messageID string) error
	// IncrementReplyCount увеличивает счётчик ответов в треде.
	IncrementReplyCount(ctx context.Context, messageID string) error
	// DecrementReplyCount уменьшает счётчик ответов в треде.
	DecrementReplyCount(ctx context.Context, messageID string) error
	// ListThreadReplies возвращает ответы в треде с курсорной пагинацией.
	ListThreadReplies(ctx context.Context, threadID string, before *time.Time, limit int) ([]*MessageDocument, error)

	// --- Реакции (MongoDB) ---

	// AddReaction добавляет реакцию пользователя к сообщению.
	AddReaction(ctx context.Context, messageID, userID, emoji string) error
	// RemoveReaction удаляет реакцию пользователя с сообщения.
	RemoveReaction(ctx context.Context, messageID, userID, emoji string) error
	// FindReaction проверяет наличие реакции пользователя.
	FindReaction(ctx context.Context, messageID, userID, emoji string) (bool, error)

	// --- Поиск (MongoDB) ---

	// SearchMessages выполняет полнотекстовый поиск по сообщениям.
	SearchMessages(ctx context.Context, query string, channelID *uuid.UUID, limit int) ([]*MessageDocument, error)

	// --- Закреплённые сообщения (Ent/PG) ---

	// AddPin закрепляет сообщение в канале.
	AddPin(ctx context.Context, channelID uuid.UUID, messageID string, pinnedBy uuid.UUID) error
	// RemovePin открепляет сообщение.
	RemovePin(ctx context.Context, channelID uuid.UUID, messageID string) error
	// FindPin ищет закреплённое сообщение.
	FindPin(ctx context.Context, channelID uuid.UUID, messageID string) (*ent.PinnedMessage, error)
	// ListPins возвращает все закреплённые сообщения канала.
	ListPins(ctx context.Context, channelID uuid.UUID) ([]*ent.PinnedMessage, error)

	// --- Закладки (Ent/PG) ---

	// AddBookmark добавляет закладку пользователя.
	AddBookmark(ctx context.Context, userID uuid.UUID, messageID string, channelID uuid.UUID) (*ent.Bookmark, error)
	// RemoveBookmark удаляет закладку.
	RemoveBookmark(ctx context.Context, bookmarkID uuid.UUID) error
	// ListBookmarks возвращает все закладки пользователя.
	ListBookmarks(ctx context.Context, userID uuid.UUID) ([]*ent.Bookmark, error)

	// --- Инициализация ---

	// EnsureIndexes создаёт необходимые MongoDB-индексы.
	EnsureIndexes(ctx context.Context) error
}

// ---------------------------------------------------------------------------
// Repository input structs
// ---------------------------------------------------------------------------

// CreateChannelInput — данные для создания канала.
type CreateChannelInput struct {
	ChannelType string
	Name        string
	Description string
	IsReadonly  bool
	ProjectID   *uuid.UUID
	CreatedBy   uuid.UUID
	DMKey       *string
}

// UpdateChannelInput — данные для обновления канала.
type UpdateChannelInput struct {
	Name        *string
	Description *string
	IsReadonly  *bool
}
