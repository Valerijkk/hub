package messenger

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

// ---------------------------------------------------------------------------
// Узкие доменные интерфейсы (ISP — Interface Segregation Principle)
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
	// DeleteChannel мягко удаляет канал (deleted_at).
	DeleteChannel(ctx context.Context, channelID uuid.UUID) error
	// IsMember проверяет участие в активном (не soft-deleted) канале.
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
	// FilterUsersSharingChannel возвращает подмножество candidateIDs,
	// которые состоят хотя бы в одном активном канале вместе с callerID (A2W3-1-001).
	FilterUsersSharingChannel(ctx context.Context, callerID uuid.UUID, candidateIDs []uuid.UUID) ([]uuid.UUID, error)
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
	// SearchMessages выполняет полнотекстовый поиск по сообщениям в заданных каналах.
	// Пустой channelIDs → пустой результат (не «все каналы»).
	SearchMessages(ctx context.Context, query string, channelIDs []uuid.UUID, limit int) ([]*MessageDocument, error)
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
	// RemoveBookmark удаляет закладку владельца (userID + bookmarkID).
	RemoveBookmark(ctx context.Context, userID, bookmarkID uuid.UUID) error
	// ListBookmarks возвращает все закладки пользователя.
	ListBookmarks(ctx context.Context, userID uuid.UUID) ([]*ent.Bookmark, error)
}

// MentionsRepository обеспечивает разрешение групповых упоминаний (@role) и lookup
// отображаемых имён пользователей для типинг-индикатора (Ent/PG).
type MentionsRepository interface {
	// ListProjectMembersByRole возвращает user IDs участников проекта с указанной project_role.
	ListProjectMembersByRole(ctx context.Context, projectID uuid.UUID, role string) ([]uuid.UUID, error)
	// GetUserDisplayName возвращает display_name (с fallback на name) пользователя.
	// Используется PublishTyping для красивого отображения "X is typing...".
	GetUserDisplayName(ctx context.Context, userID uuid.UUID) (string, error)
}

// ChannelInviteRepository управляет пригласительными ссылками каналов (Ent/PG).
type ChannelInviteRepository interface {
	// CreateChannelInvite создаёт новую пригласительную ссылку.
	CreateChannelInvite(ctx context.Context, input CreateChannelInviteInput) (*ent.ChannelInvite, error)
	// FindChannelInviteByToken ищет приглашение по токену.
	FindChannelInviteByToken(ctx context.Context, token string) (*ent.ChannelInvite, error)
	// IncrementInviteUses атомарно увеличивает счётчик использований.
	// maxUses>0 → WHERE use_count < maxUses; при 0 затронутых строк — ErrInviteUsesExhausted.
	IncrementInviteUses(ctx context.Context, inviteID uuid.UUID, maxUses int) error
	// DecrementInviteUses откатывает слот использования (после fail AddMembers).
	DecrementInviteUses(ctx context.Context, inviteID uuid.UUID) error
	// RevokeChannelInvite ставит is_revoked=true для приглашения в рамках канала.
	RevokeChannelInvite(ctx context.Context, channelID, inviteID uuid.UUID) error
	// ListChannelInvites возвращает активные (не отозванные) приглашения канала.
	ListChannelInvites(ctx context.Context, channelID uuid.UUID) ([]*ent.ChannelInvite, error)
}

// PollRepository управляет опросами в каналах: создание, голосование, закрытие (Ent/PG).
// Транзакционные шаблоны (CreatePollWithOptions, VotePoll) инкапсулированы внутри
// реализации — Tx не утекает в сервисный слой.
type PollRepository interface {
	// CreatePollWithOptions создаёт опрос вместе с вариантами ответов в одной транзакции.
	CreatePollWithOptions(ctx context.Context, input CreatePollInput) (*ent.Poll, []*ent.PollOption, error)
	// GetPollWithOptions возвращает опрос вместе с вариантами ответов (упорядоченные по position).
	GetPollWithOptions(ctx context.Context, pollID uuid.UUID) (*ent.Poll, []*ent.PollOption, error)
	// GetPoll возвращает опрос по ID (без вариантов).
	GetPoll(ctx context.Context, pollID uuid.UUID) (*ent.Poll, error)
	// ClosePoll переводит опрос в is_closed=true.
	ClosePoll(ctx context.Context, pollID uuid.UUID) error
	// VotePoll идемпотентно записывает голоса пользователя (удаляет старые, добавляет новые) в Tx.
	VotePoll(ctx context.Context, pollID, userID uuid.UUID, optionIDs []uuid.UUID) error
	// ListVotes возвращает все голоса опроса.
	ListVotes(ctx context.Context, pollID uuid.UUID) ([]*ent.PollVote, error)
}

// ---------------------------------------------------------------------------
// Композитный интерфейс (обратная совместимость)
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
	MentionsRepository
	ChannelInviteRepository
	PollRepository

	// EnsureIndexes создаёт необходимые MongoDB-индексы.
	EnsureIndexes(ctx context.Context) error
}

// ---------------------------------------------------------------------------
// Входные структуры репозитория
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
	Version     int
}

// CreateChannelInviteInput — данные для создания пригласительной ссылки.
type CreateChannelInviteInput struct {
	ChannelID uuid.UUID
	CreatorID uuid.UUID
	Token     string
	MaxUses   int
	ExpiresAt *time.Time
}

// CreatePollInput — данные для создания опроса.
type CreatePollInput struct {
	ChannelID     uuid.UUID
	CreatorID     uuid.UUID
	Question      string
	IsAnonymous   bool
	IsMultiChoice bool
	Options       []string // тексты вариантов в порядке отображения (position = index)
}
