package messenger

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/bookmark"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/ent/pinnedmessage"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
)

// EntRepository реализует Ent/PG-часть интерфейса Repository.
// Отвечает за CRUD каналов, участников, закладок и закреплённых сообщений.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository создаёт новый Ent-репозиторий для мессенджера.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

// CreateChannel создаёт новый канал в PG через Ent.
func (r *EntRepository) CreateChannel(ctx context.Context, input CreateChannelInput) (*ent.Channel, error) {
	q := r.client.Channel.Create().
		SetChannelType(channel.ChannelType(input.ChannelType)).
		SetIsReadonly(input.IsReadonly).
		SetCreatedBy(input.CreatedBy)

	if input.Name != "" {
		q.SetName(input.Name)
	}
	if input.Description != "" {
		q.SetDescription(input.Description)
	}
	if input.ProjectID != nil {
		q.SetProjectID(*input.ProjectID)
	}
	if input.DMKey != nil {
		q.SetDmKey(*input.DMKey)
	}

	return q.Save(ctx)
}

// GetChannel возвращает канал по ID с загрузкой участников.
func (r *EntRepository) GetChannel(ctx context.Context, channelID uuid.UUID) (*ent.Channel, error) {
	return r.client.Channel.Query().
		Where(channel.ID(channelID)).
		WithMembers().
		Only(ctx)
}

// ListUserChannels возвращает все каналы, в которых состоит пользователь.
func (r *EntRepository) ListUserChannels(ctx context.Context, userID uuid.UUID) ([]*ent.Channel, error) {
	return r.client.Channel.Query().
		Where(channel.HasMembersWith(channelmember.UserID(userID))).
		All(ctx)
}

// UpdateChannel обновляет данные канала.
func (r *EntRepository) UpdateChannel(ctx context.Context, channelID uuid.UUID, input UpdateChannelInput) (*ent.Channel, error) {
	q := r.client.Channel.UpdateOneID(channelID)
	if input.Name != nil {
		q.SetName(*input.Name)
	}
	if input.Description != nil {
		q.SetDescription(*input.Description)
	}
	if input.IsReadonly != nil {
		q.SetIsReadonly(*input.IsReadonly)
	}
	return q.Save(ctx)
}

// DeleteChannel удаляет канал и связанные данные.
func (r *EntRepository) DeleteChannel(ctx context.Context, channelID uuid.UUID) error {
	return r.client.Channel.DeleteOneID(channelID).Exec(ctx)
}

// IsMember проверяет, является ли пользователь участником канала.
func (r *EntRepository) IsMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	return r.client.ChannelMember.Query().
		Where(
			channelmember.ChannelID(channelID),
			channelmember.UserID(userID),
		).
		Exist(ctx)
}

// GetMemberRole возвращает роль участника в канале.
func (r *EntRepository) GetMemberRole(ctx context.Context, channelID, userID uuid.UUID) (string, error) {
	m, err := r.client.ChannelMember.Query().
		Where(
			channelmember.ChannelID(channelID),
			channelmember.UserID(userID),
		).
		Only(ctx)
	if err != nil {
		return "", err
	}
	return string(m.Role), nil
}

// GetChannelMemberIDs возвращает ID всех участников канала.
func (r *EntRepository) GetChannelMemberIDs(ctx context.Context, channelID uuid.UUID) ([]uuid.UUID, error) {
	members, err := r.client.ChannelMember.Query().
		Where(channelmember.ChannelID(channelID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(members))
	for _, m := range members {
		ids = append(ids, m.UserID)
	}
	return ids, nil
}

// AddMembers добавляет участников в канал с указанной ролью.
func (r *EntRepository) AddMembers(ctx context.Context, channelID uuid.UUID, userIDs []uuid.UUID, role string) error {
	bulk := make([]*ent.ChannelMemberCreate, 0, len(userIDs))
	for _, uid := range userIDs {
		bulk = append(bulk, r.client.ChannelMember.Create().
			SetChannelID(channelID).
			SetUserID(uid).
			SetRole(channelmember.Role(role)),
		)
	}
	_, err := r.client.ChannelMember.CreateBulk(bulk...).Save(ctx)
	return err
}

// RemoveMember удаляет участника из канала.
func (r *EntRepository) RemoveMember(ctx context.Context, channelID, userID uuid.UUID) error {
	_, err := r.client.ChannelMember.Delete().
		Where(
			channelmember.ChannelID(channelID),
			channelmember.UserID(userID),
		).
		Exec(ctx)
	return err
}

// FindDMChannel ищет существующий DM-канал по ключу дедупликации.
func (r *EntRepository) FindDMChannel(ctx context.Context, dmKey string) (*ent.Channel, error) {
	return r.client.Channel.Query().
		Where(channel.DmKey(dmKey)).
		WithMembers().
		Only(ctx)
}

// ListMembers возвращает всех участников канала с загрузкой связи user.
func (r *EntRepository) ListMembers(ctx context.Context, channelID uuid.UUID) ([]*ent.ChannelMember, error) {
	return r.client.ChannelMember.Query().
		Where(channelmember.ChannelID(channelID)).
		WithUser().
		All(ctx)
}

// ToggleStar переключает is_starred для участника канала.
// Возвращает новое значение is_starred.
func (r *EntRepository) ToggleStar(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	m, err := r.client.ChannelMember.Query().
		Where(
			channelmember.ChannelID(channelID),
			channelmember.UserID(userID),
		).
		Only(ctx)
	if err != nil {
		return false, err
	}

	newVal := !m.IsStarred
	_, err = r.client.ChannelMember.UpdateOneID(m.ID).
		SetIsStarred(newVal).
		Save(ctx)
	if err != nil {
		return false, err
	}
	return newVal, nil
}

// MuteChannel устанавливает muted_until для участника канала.
// duration: "1h", "2h", "8h", "1d", "forever".
func (r *EntRepository) MuteChannel(ctx context.Context, channelID, userID uuid.UUID, duration string) (time.Time, error) {
	m, err := r.client.ChannelMember.Query().
		Where(
			channelmember.ChannelID(channelID),
			channelmember.UserID(userID),
		).
		Only(ctx)
	if err != nil {
		return time.Time{}, err
	}

	var mutedUntil time.Time
	switch duration {
	case "1h":
		mutedUntil = time.Now().Add(1 * time.Hour)
	case "2h":
		mutedUntil = time.Now().Add(2 * time.Hour)
	case "8h":
		mutedUntil = time.Now().Add(8 * time.Hour)
	case "1d":
		mutedUntil = time.Now().Add(24 * time.Hour)
	case "forever":
		// Далёкая дата в будущем (100 лет).
		mutedUntil = time.Now().Add(100 * 365 * 24 * time.Hour)
	default:
		return time.Time{}, fmt.Errorf("некорректная длительность: %s", duration)
	}

	_, err = r.client.ChannelMember.UpdateOneID(m.ID).
		SetMutedUntil(mutedUntil).
		Save(ctx)
	if err != nil {
		return time.Time{}, err
	}
	return mutedUntil, nil
}

// UnmuteChannel сбрасывает muted_until для участника канала.
func (r *EntRepository) UnmuteChannel(ctx context.Context, channelID, userID uuid.UUID) error {
	m, err := r.client.ChannelMember.Query().
		Where(
			channelmember.ChannelID(channelID),
			channelmember.UserID(userID),
		).
		Only(ctx)
	if err != nil {
		return err
	}

	_, err = r.client.ChannelMember.UpdateOneID(m.ID).
		ClearMutedUntil().
		Save(ctx)
	return err
}

// IsChannelMuted проверяет, замьючен ли канал для пользователя.
// Возвращает true если muted_until не nil и > текущего времени.
func (r *EntRepository) IsChannelMuted(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	m, err := r.client.ChannelMember.Query().
		Where(
			channelmember.ChannelID(channelID),
			channelmember.UserID(userID),
		).
		Only(ctx)
	if err != nil {
		return false, err
	}
	if m.MutedUntil == nil {
		return false, nil
	}
	return m.MutedUntil.After(time.Now()), nil
}

// GetUserChannelMemberships возвращает membership-записи пользователя.
func (r *EntRepository) GetUserChannelMemberships(ctx context.Context, userID uuid.UUID) ([]*ent.ChannelMember, error) {
	return r.client.ChannelMember.Query().
		Where(channelmember.UserID(userID)).
		All(ctx)
}

// AddPin закрепляет сообщение в канале.
func (r *EntRepository) AddPin(ctx context.Context, channelID uuid.UUID, messageID string, pinnedBy uuid.UUID) error {
	_, err := r.client.PinnedMessage.Create().
		SetMessageID(messageID).
		SetChannelID(channelID).
		SetPinnedBy(pinnedBy).
		Save(ctx)
	return err
}

// RemovePin открепляет сообщение из канала.
func (r *EntRepository) RemovePin(ctx context.Context, channelID uuid.UUID, messageID string) error {
	_, err := r.client.PinnedMessage.Delete().
		Where(
			pinnedmessage.ChannelID(channelID),
			pinnedmessage.MessageID(messageID),
		).
		Exec(ctx)
	return err
}

// FindPin ищет закреплённое сообщение.
func (r *EntRepository) FindPin(ctx context.Context, channelID uuid.UUID, messageID string) (*ent.PinnedMessage, error) {
	return r.client.PinnedMessage.Query().
		Where(
			pinnedmessage.ChannelID(channelID),
			pinnedmessage.MessageID(messageID),
		).
		Only(ctx)
}

// ListPins возвращает все закреплённые сообщения канала.
func (r *EntRepository) ListPins(ctx context.Context, channelID uuid.UUID) ([]*ent.PinnedMessage, error) {
	return r.client.PinnedMessage.Query().
		Where(pinnedmessage.ChannelID(channelID)).
		All(ctx)
}

// AddBookmark добавляет закладку пользователя на сообщение.
func (r *EntRepository) AddBookmark(ctx context.Context, userID uuid.UUID, messageID string, channelID uuid.UUID) (*ent.Bookmark, error) {
	return r.client.Bookmark.Create().
		SetUserID(userID).
		SetMessageID(messageID).
		SetChannelID(channelID).
		Save(ctx)
}

// RemoveBookmark удаляет закладку по ID.
func (r *EntRepository) RemoveBookmark(ctx context.Context, bookmarkID uuid.UUID) error {
	return r.client.Bookmark.DeleteOneID(bookmarkID).Exec(ctx)
}

// ListBookmarks возвращает все закладки пользователя.
func (r *EntRepository) ListBookmarks(ctx context.Context, userID uuid.UUID) ([]*ent.Bookmark, error) {
	return r.client.Bookmark.Query().
		Where(bookmark.UserID(userID)).
		All(ctx)
}

// ListProjectMembers возвращает ID всех участников проекта.
// Используется AutoChannelService для добавления участников в авто-канал.
func (r *EntRepository) ListProjectMembers(ctx context.Context, projectID uuid.UUID) ([]uuid.UUID, error) {
	members, err := r.client.ProjectMember.Query().
		Where(projectmember.HasProjectWith(project.IDEQ(projectID))).
		WithUser().
		All(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(members))
	for _, m := range members {
		if m.Edges.User != nil {
			ids = append(ids, m.Edges.User.ID)
		}
	}
	return ids, nil
}

// GetProjectByID возвращает проект по ID для получения ключа проекта.
func (r *EntRepository) GetProjectByID(ctx context.Context, projectID uuid.UUID) (*ent.Project, error) {
	return r.client.Project.Get(ctx, projectID)
}
