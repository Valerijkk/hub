package messenger

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

// CombinedRepository объединяет EntRepository и MongoRepository
// в единую реализацию интерфейса Repository.
type CombinedRepository struct {
	ent   *EntRepository
	mongo *MongoRepository
}

// NewCombinedRepository создаёт комбинированный репозиторий из Ent и MongoDB частей.
func NewCombinedRepository(entRepo *EntRepository, mongoRepo *MongoRepository) *CombinedRepository {
	return &CombinedRepository{ent: entRepo, mongo: mongoRepo}
}

// --- Каналы (делегируем в EntRepository) ---

// CreateChannel делегирует создание канала в EntRepository.
func (r *CombinedRepository) CreateChannel(ctx context.Context, input CreateChannelInput) (*ent.Channel, error) {
	return r.ent.CreateChannel(ctx, input)
}

// GetChannel делегирует получение канала в EntRepository.
func (r *CombinedRepository) GetChannel(ctx context.Context, channelID uuid.UUID) (*ent.Channel, error) {
	return r.ent.GetChannel(ctx, channelID)
}

// ListUserChannels делегирует получение каналов пользователя в EntRepository.
func (r *CombinedRepository) ListUserChannels(ctx context.Context, userID uuid.UUID) ([]*ent.Channel, error) {
	return r.ent.ListUserChannels(ctx, userID)
}

// UpdateChannel делегирует обновление канала в EntRepository.
func (r *CombinedRepository) UpdateChannel(ctx context.Context, channelID uuid.UUID, input UpdateChannelInput) (*ent.Channel, error) {
	return r.ent.UpdateChannel(ctx, channelID, input)
}

// DeleteChannel делегирует удаление канала в EntRepository.
func (r *CombinedRepository) DeleteChannel(ctx context.Context, channelID uuid.UUID) error {
	return r.ent.DeleteChannel(ctx, channelID)
}

// IsMember делегирует проверку участия в канале в EntRepository.
func (r *CombinedRepository) IsMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	return r.ent.IsMember(ctx, channelID, userID)
}

// GetMemberRole делегирует получение роли участника в EntRepository.
func (r *CombinedRepository) GetMemberRole(ctx context.Context, channelID, userID uuid.UUID) (string, error) {
	return r.ent.GetMemberRole(ctx, channelID, userID)
}

// GetChannelMemberIDs делегирует получение ID участников канала в EntRepository.
func (r *CombinedRepository) GetChannelMemberIDs(ctx context.Context, channelID uuid.UUID) ([]uuid.UUID, error) {
	return r.ent.GetChannelMemberIDs(ctx, channelID)
}

// AddMembers делегирует добавление участников в канал в EntRepository.
func (r *CombinedRepository) AddMembers(ctx context.Context, channelID uuid.UUID, userIDs []uuid.UUID, role string) error {
	return r.ent.AddMembers(ctx, channelID, userIDs, role)
}

// RemoveMember делегирует удаление участника из канала в EntRepository.
func (r *CombinedRepository) RemoveMember(ctx context.Context, channelID, userID uuid.UUID) error {
	return r.ent.RemoveMember(ctx, channelID, userID)
}

// FindDMChannel делегирует поиск DM-канала в EntRepository.
func (r *CombinedRepository) FindDMChannel(ctx context.Context, dmKey string) (*ent.Channel, error) {
	return r.ent.FindDMChannel(ctx, dmKey)
}

// ListMembers делегирует получение участников канала в EntRepository.
func (r *CombinedRepository) ListMembers(ctx context.Context, channelID uuid.UUID) ([]*ent.ChannelMember, error) {
	return r.ent.ListMembers(ctx, channelID)
}

// --- Star / Mute (делегируем в EntRepository) ---

// ToggleStar делегирует переключение star в EntRepository.
func (r *CombinedRepository) ToggleStar(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	return r.ent.ToggleStar(ctx, channelID, userID)
}

// MuteChannel делегирует мьют канала в EntRepository.
func (r *CombinedRepository) MuteChannel(ctx context.Context, channelID, userID uuid.UUID, duration string) (time.Time, error) {
	return r.ent.MuteChannel(ctx, channelID, userID, duration)
}

// UnmuteChannel делегирует снятие мьюта канала в EntRepository.
func (r *CombinedRepository) UnmuteChannel(ctx context.Context, channelID, userID uuid.UUID) error {
	return r.ent.UnmuteChannel(ctx, channelID, userID)
}

// IsChannelMuted делегирует проверку мьюта канала в EntRepository.
func (r *CombinedRepository) IsChannelMuted(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	return r.ent.IsChannelMuted(ctx, channelID, userID)
}

// GetUserChannelMemberships делегирует получение membership-данных пользователя в EntRepository.
func (r *CombinedRepository) GetUserChannelMemberships(ctx context.Context, userID uuid.UUID) ([]*ent.ChannelMember, error) {
	return r.ent.GetUserChannelMemberships(ctx, userID)
}

// --- Сообщения (делегируем в MongoRepository) ---

// InsertMessage делегирует сохранение сообщения в MongoRepository.
func (r *CombinedRepository) InsertMessage(ctx context.Context, doc *MessageDocument) (string, error) {
	return r.mongo.InsertMessage(ctx, doc)
}

// FindMessageByID делегирует поиск сообщения по ID в MongoRepository.
func (r *CombinedRepository) FindMessageByID(ctx context.Context, messageID string) (*MessageDocument, error) {
	return r.mongo.FindMessageByID(ctx, messageID)
}

// ListMessages делегирует получение сообщений канала в MongoRepository.
func (r *CombinedRepository) ListMessages(ctx context.Context, channelID uuid.UUID, before *time.Time, limit int) ([]*MessageDocument, error) {
	return r.mongo.ListMessages(ctx, channelID, before, limit)
}

// UpdateMessage делегирует обновление сообщения в MongoRepository.
func (r *CombinedRepository) UpdateMessage(ctx context.Context, messageID string, contentJSON map[string]interface{}, contentText string) (*MessageDocument, error) {
	return r.mongo.UpdateMessage(ctx, messageID, contentJSON, contentText)
}

// SoftDeleteMessage делегирует мягкое удаление сообщения в MongoRepository.
func (r *CombinedRepository) SoftDeleteMessage(ctx context.Context, messageID string) error {
	return r.mongo.SoftDeleteMessage(ctx, messageID)
}

// IncrementReplyCount делегирует увеличение счётчика ответов в MongoRepository.
func (r *CombinedRepository) IncrementReplyCount(ctx context.Context, messageID string) error {
	return r.mongo.IncrementReplyCount(ctx, messageID)
}

// DecrementReplyCount делегирует уменьшение счётчика ответов в MongoRepository.
func (r *CombinedRepository) DecrementReplyCount(ctx context.Context, messageID string) error {
	return r.mongo.DecrementReplyCount(ctx, messageID)
}

// ListThreadReplies делегирует получение ответов в треде в MongoRepository.
func (r *CombinedRepository) ListThreadReplies(ctx context.Context, threadID string, before *time.Time, limit int) ([]*MessageDocument, error) {
	return r.mongo.ListThreadReplies(ctx, threadID, before, limit)
}

// --- Реакции (делегируем в MongoRepository) ---

// AddReaction делегирует добавление реакции в MongoRepository.
func (r *CombinedRepository) AddReaction(ctx context.Context, messageID, userID, emoji string) error {
	return r.mongo.AddReaction(ctx, messageID, userID, emoji)
}

// RemoveReaction делегирует удаление реакции в MongoRepository.
func (r *CombinedRepository) RemoveReaction(ctx context.Context, messageID, userID, emoji string) error {
	return r.mongo.RemoveReaction(ctx, messageID, userID, emoji)
}

// FindReaction делегирует проверку наличия реакции в MongoRepository.
func (r *CombinedRepository) FindReaction(ctx context.Context, messageID, userID, emoji string) (bool, error) {
	return r.mongo.FindReaction(ctx, messageID, userID, emoji)
}

// --- Поиск (делегируем в MongoRepository) ---

// SearchMessages делегирует полнотекстовый поиск сообщений в MongoRepository.
func (r *CombinedRepository) SearchMessages(ctx context.Context, query string, channelID *uuid.UUID, limit int) ([]*MessageDocument, error) {
	return r.mongo.SearchMessages(ctx, query, channelID, limit)
}

// --- Закреплённые сообщения (делегируем в EntRepository) ---

// AddPin делегирует закрепление сообщения в EntRepository.
func (r *CombinedRepository) AddPin(ctx context.Context, channelID uuid.UUID, messageID string, pinnedBy uuid.UUID) error {
	return r.ent.AddPin(ctx, channelID, messageID, pinnedBy)
}

// RemovePin делегирует открепление сообщения в EntRepository.
func (r *CombinedRepository) RemovePin(ctx context.Context, channelID uuid.UUID, messageID string) error {
	return r.ent.RemovePin(ctx, channelID, messageID)
}

// FindPin делегирует поиск закреплённого сообщения в EntRepository.
func (r *CombinedRepository) FindPin(ctx context.Context, channelID uuid.UUID, messageID string) (*ent.PinnedMessage, error) {
	return r.ent.FindPin(ctx, channelID, messageID)
}

// ListPins делегирует получение закреплённых сообщений в EntRepository.
func (r *CombinedRepository) ListPins(ctx context.Context, channelID uuid.UUID) ([]*ent.PinnedMessage, error) {
	return r.ent.ListPins(ctx, channelID)
}

// --- Закладки (делегируем в EntRepository) ---

// AddBookmark делегирует добавление закладки в EntRepository.
func (r *CombinedRepository) AddBookmark(ctx context.Context, userID uuid.UUID, messageID string, channelID uuid.UUID) (*ent.Bookmark, error) {
	return r.ent.AddBookmark(ctx, userID, messageID, channelID)
}

// RemoveBookmark делегирует удаление закладки в EntRepository.
func (r *CombinedRepository) RemoveBookmark(ctx context.Context, bookmarkID uuid.UUID) error {
	return r.ent.RemoveBookmark(ctx, bookmarkID)
}

// ListBookmarks делегирует получение закладок пользователя в EntRepository.
func (r *CombinedRepository) ListBookmarks(ctx context.Context, userID uuid.UUID) ([]*ent.Bookmark, error) {
	return r.ent.ListBookmarks(ctx, userID)
}

// --- Инициализация ---

// EnsureIndexes делегирует создание MongoDB-индексов в MongoRepository.
func (r *CombinedRepository) EnsureIndexes(ctx context.Context) error {
	return r.mongo.EnsureIndexes(ctx)
}
