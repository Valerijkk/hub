package messenger

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

// ---------------------------------------------------------------------------
// Focused domain interfaces (ISP — Interface Segregation Principle)
// ---------------------------------------------------------------------------

// ChannelRepository управляет каналами, участниками и правами доступа (Ent/PG).
type ChannelRepository interface {
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
}

// StarMuteRepository управляет закреплением (starred) и мьютом каналов (Ent/PG).
type StarMuteRepository interface {
	// ToggleStar переключает is_starred для пользователя в канале.
	ToggleStar(ctx context.Context, channelID, userID uuid.UUID) (bool, error)
	// MuteChannel устанавливает muted_until для пользователя в канале.
	MuteChannel(ctx context.Context, channelID, userID uuid.UUID, duration string) (time.Time, error)
	// UnmuteChannel сбрасывает muted_until для пользователя в канале.
	UnmuteChannel(ctx context.Context, channelID, userID uuid.UUID) error
	// IsChannelMuted проверяет, замьючен ли канал для пользователя.
	IsChannelMuted(ctx context.Context, channelID, userID uuid.UUID) (bool, error)
	// GetUserChannelMemberships возвращает membership-данные пользователя (starred, muted).
	GetUserChannelMemberships(ctx context.Context, userID uuid.UUID) ([]*ent.ChannelMember, error)
}

// MessageRepository управляет сообщениями и тредами (MongoDB).
type MessageRepository interface {
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
}

// ReactionRepository управляет реакциями на сообщения (MongoDB).
type ReactionRepository interface {
	// AddReaction добавляет реакцию пользователя к сообщению.
	AddReaction(ctx context.Context, messageID, userID, emoji string) error
	// RemoveReaction удаляет реакцию пользователя с сообщения.
	RemoveReaction(ctx context.Context, messageID, userID, emoji string) error
	// FindReaction проверяет наличие реакции пользователя.
	FindReaction(ctx context.Context, messageID, userID, emoji string) (bool, error)
}

// SearchRepository обеспечивает полнотекстовый поиск сообщений (MongoDB).
type SearchRepository interface {
	// SearchMessages выполняет полнотекстовый поиск по сообщениям.
	SearchMessages(ctx context.Context, query string, channelID *uuid.UUID, limit int) ([]*MessageDocument, error)
}

// PinRepository управляет закреплёнными сообщениями (Ent/PG).
type PinRepository interface {
	// AddPin закрепляет сообщение в канале.
	AddPin(ctx context.Context, channelID uuid.UUID, messageID string, pinnedBy uuid.UUID) error
	// RemovePin открепляет сообщение.
	RemovePin(ctx context.Context, channelID uuid.UUID, messageID string) error
	// FindPin ищет закреплённое сообщение.
	FindPin(ctx context.Context, channelID uuid.UUID, messageID string) (*ent.PinnedMessage, error)
	// ListPins возвращает все закреплённые сообщения канала.
	ListPins(ctx context.Context, channelID uuid.UUID) ([]*ent.PinnedMessage, error)
}

// BookmarkRepository управляет закладками (Ent/PG).
type BookmarkRepository interface {
	// AddBookmark добавляет закладку пользователя.
	AddBookmark(ctx context.Context, userID uuid.UUID, messageID string, channelID uuid.UUID) (*ent.Bookmark, error)
	// RemoveBookmark удаляет закладку.
	RemoveBookmark(ctx context.Context, bookmarkID uuid.UUID) error
	// ListBookmarks возвращает все закладки пользователя.
	ListBookmarks(ctx context.Context, userID uuid.UUID) ([]*ent.Bookmark, error)
}

// ---------------------------------------------------------------------------
// Composite interface (backward-compatible)
// ---------------------------------------------------------------------------

// Repository — композитный интерфейс для обратной совместимости.
// Разделён на две реализации: EntRepository (PG) и MongoRepository (MongoDB).
type Repository interface {
	ChannelRepository
	StarMuteRepository
	MessageRepository
	ReactionRepository
	SearchRepository
	PinRepository
	BookmarkRepository

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
