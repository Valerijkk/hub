package notification

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/notification"
	"github.com/valerii/hub/backend/ent/notificationpreference"
	"github.com/valerii/hub/backend/ent/user"
)

// Repository предоставляет доступ к данным уведомлений через Ent.
type Repository struct {
	db *ent.Client
}

// NewRepository создаёт новый репозиторий уведомлений.
func NewRepository(db *ent.Client) *Repository {
	return &Repository{db: db}
}

// CreateNotification создаёт новое уведомление. Дедупликация: если уведомление
// с таким же (user_id, type, entity_id) было создано менее 2 секунд назад — пропускает.
func (r *Repository) CreateNotification(ctx context.Context, userID uuid.UUID, notifType, title, body, entityType, entityID string, projectID, actorID *uuid.UUID, isUrgent bool) (*ent.Notification, error) {
	// Дедупликация: проверяем наличие аналогичного уведомления за последние 2 секунды.
	if entityID != "" {
		cutoff := time.Now().Add(-2 * time.Second)
		exists, err := r.db.Notification.Query().
			Where(
				notification.UserID(userID),
				notification.Type(notifType),
				notification.EntityID(entityID),
				notification.CreatedAtGTE(cutoff),
			).
			Exist(ctx)
		if err == nil && exists {
			return nil, nil
		}
	}

	builder := r.db.Notification.Create().
		SetUserID(userID).
		SetType(notifType).
		SetTitle(title).
		SetIsUrgent(isUrgent)

	if body != "" {
		builder.SetBody(body)
	}
	if entityType != "" {
		builder.SetEntityType(entityType)
	}
	if entityID != "" {
		builder.SetEntityID(entityID)
	}
	if projectID != nil {
		builder.SetProjectID(*projectID)
	}
	if actorID != nil {
		builder.SetActorID(*actorID)
	}

	return builder.Save(ctx)
}

// ListByUser возвращает пагинированный список уведомлений пользователя.
// isRead = nil — все, isRead = &true — только прочитанные, &false — только непрочитанные.
func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID, page, perPage int, isRead *bool) ([]*ent.Notification, int, error) {
	query := r.db.Notification.Query().
		Where(notification.UserID(userID))

	if isRead != nil {
		query = query.Where(notification.IsRead(*isRead))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	items, err := query.
		Order(ent.Desc(notification.FieldCreatedAt)).
		Offset(offset).
		Limit(perPage).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// MarkRead помечает уведомление как прочитанное.
func (r *Repository) MarkRead(ctx context.Context, notifID, userID uuid.UUID) error {
	return r.db.Notification.Update().
		Where(
			notification.ID(notifID),
			notification.UserID(userID),
		).
		SetIsRead(true).
		Exec(ctx)
}

// MarkAllRead помечает все уведомления пользователя как прочитанные.
func (r *Repository) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return r.db.Notification.Update().
		Where(
			notification.UserID(userID),
			notification.IsRead(false),
		).
		SetIsRead(true).
		Exec(ctx)
}

// DeleteNotification удаляет уведомление пользователя.
func (r *Repository) DeleteNotification(ctx context.Context, notifID, userID uuid.UUID) error {
	_, err := r.db.Notification.Delete().
		Where(
			notification.ID(notifID),
			notification.UserID(userID),
		).
		Exec(ctx)
	return err
}

// CountUnread возвращает количество непрочитанных уведомлений пользователя.
func (r *Repository) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.db.Notification.Query().
		Where(
			notification.UserID(userID),
			notification.IsRead(false),
		).
		Count(ctx)
}

// GetPreferences возвращает все настройки уведомлений пользователя.
func (r *Repository) GetPreferences(ctx context.Context, userID uuid.UUID) ([]*ent.NotificationPreference, error) {
	return r.db.NotificationPreference.Query().
		Where(notificationpreference.UserID(userID)).
		All(ctx)
}

// UpsertPreference создаёт или обновляет настройку уведомлений.
func (r *Repository) UpsertPreference(ctx context.Context, userID uuid.UUID, notifType, channel string, enabled bool, projectID *uuid.UUID) (*ent.NotificationPreference, error) {
	// Ищем существующую настройку.
	query := r.db.NotificationPreference.Query().
		Where(
			notificationpreference.UserID(userID),
			notificationpreference.NotificationType(notifType),
			notificationpreference.Channel(channel),
		)

	if projectID != nil {
		query = query.Where(notificationpreference.ProjectID(*projectID))
	} else {
		query = query.Where(notificationpreference.ProjectIDIsNil())
	}

	existing, err := query.First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}

	if existing != nil {
		return existing.Update().
			SetEnabled(enabled).
			Save(ctx)
	}

	builder := r.db.NotificationPreference.Create().
		SetUserID(userID).
		SetNotificationType(notifType).
		SetChannel(channel).
		SetEnabled(enabled)

	if projectID != nil {
		builder.SetProjectID(*projectID)
	}

	return builder.Save(ctx)
}

// --- Методы для consumer (без прямого доступа consumer к *ent.Client) ---

// GetIssueAssignee возвращает назначенного пользователя задачи по ID задачи.
// Возвращает nil, nil если задача не найдена или назначенный не задан.
func (r *Repository) GetIssueAssignee(ctx context.Context, issueID uuid.UUID) (*ent.User, error) {
	issue, err := r.db.Issue.Get(ctx, issueID)
	if err != nil {
		return nil, err
	}
	assignee, err := issue.QueryAssignee().First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return assignee, nil
}

// GetChannel возвращает канал по ID.
func (r *Repository) GetChannel(ctx context.Context, channelID uuid.UUID) (*ent.Channel, error) {
	return r.db.Channel.Get(ctx, channelID)
}

// GetChannelMembers возвращает список участников канала.
func (r *Repository) GetChannelMembers(ctx context.Context, channelID uuid.UUID) ([]*ent.ChannelMember, error) {
	ch, err := r.db.Channel.Get(ctx, channelID)
	if err != nil {
		return nil, err
	}
	return ch.QueryMembers().All(ctx)
}

// GetProjectMembersWithUser возвращает участников проекта с загруженными данными пользователя.
func (r *Repository) GetProjectMembersWithUser(ctx context.Context, projectID uuid.UUID) ([]*ent.ProjectMember, error) {
	proj, err := r.db.Project.Get(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return proj.QueryMembers().WithUser().All(ctx)
}

// GetUser возвращает пользователя по ID. Используется сервисом для чтения
// telegram_id / quiet hours / timezone — поля прямо лежат на User.
func (r *Repository) GetUser(ctx context.Context, userID uuid.UUID) (*ent.User, error) {
	return r.db.User.Get(ctx, userID)
}

// ListIssuesWithUpcomingDeadline возвращает задачи с дедлайном в окне (now, deadline],
// которые не удалены, с preload'ом assignee и project — используется CheckDeadlines.
func (r *Repository) ListIssuesWithUpcomingDeadline(ctx context.Context, now, deadline time.Time) ([]*ent.Issue, error) {
	return r.db.Issue.Query().
		Where(
			issue.DueDateNotNil(),
			issue.DueDateGTE(now),
			issue.DueDateLTE(deadline),
			issue.DeletedAtIsNil(),
		).
		WithAssignee().
		WithProject().
		All(ctx)
}

// IsEnabled проверяет, включено ли уведомление данного типа для канала и пользователя.
// Сначала проверяет per-project настройку, потом глобальную. По умолчанию — включено.
func (r *Repository) IsEnabled(ctx context.Context, userID uuid.UUID, notifType, channel string, projectID *uuid.UUID) (bool, error) {
	prefs, err := r.ResolveChannelPrefs(ctx, userID, notifType, projectID)
	if err != nil {
		return false, err
	}
	enabled, ok := prefs[channel]
	if !ok {
		return true, nil
	}
	return enabled, nil
}

// ResolveChannelPrefs — A13-005: одним (до двух с project) запросом читает prefs
// для всех каналов user+type вместо 3× IsEnabled.
func (r *Repository) ResolveChannelPrefs(ctx context.Context, userID uuid.UUID, notifType string, projectID *uuid.UUID) (map[string]bool, error) {
	out := map[string]bool{
		ChannelInApp:    true,
		ChannelTelegram: true,
		ChannelPush:     true,
	}

	q := r.db.NotificationPreference.Query().Where(
		notificationpreference.UserID(userID),
		notificationpreference.NotificationType(notifType),
	)
	if projectID != nil {
		q = q.Where(
			notificationpreference.Or(
				notificationpreference.ProjectID(*projectID),
				notificationpreference.ProjectIDIsNil(),
			),
		)
	} else {
		q = q.Where(notificationpreference.ProjectIDIsNil())
	}

	rows, err := q.All(ctx)
	if err != nil {
		return nil, err
	}

	// Глобальные сначала, затем project override.
	for _, pref := range rows {
		if pref.ProjectID == nil {
			out[pref.Channel] = pref.Enabled
		}
	}
	if projectID != nil {
		for _, pref := range rows {
			if pref.ProjectID != nil && *pref.ProjectID == *projectID {
				out[pref.Channel] = pref.Enabled
			}
		}
	}
	return out, nil
}

// GetUserTelegram загружает поля telegram пользователя для статуса bind.
func (r *Repository) GetUserTelegram(ctx context.Context, userID uuid.UUID) (*ent.User, error) {
	return r.db.User.Get(ctx, userID)
}

// FindUserByTelegramID находит пользователя по telegram_id.
func (r *Repository) FindUserByTelegramID(ctx context.Context, telegramID string) (*ent.User, error) {
	return r.db.User.Query().Where(user.TelegramIDEQ(telegramID)).Only(ctx)
}

// IsTelegramIDTaken проверяет, привязан ли telegram_id к другому пользователю.
func (r *Repository) IsTelegramIDTaken(ctx context.Context, telegramID string, excludeUserID uuid.UUID) (bool, error) {
	return r.db.User.Query().
		Where(
			user.TelegramIDEQ(telegramID),
			user.IDNEQ(excludeUserID),
		).
		Exist(ctx)
}

// SetUserTelegram записывает telegram_id на пользователя.
func (r *Repository) SetUserTelegram(ctx context.Context, userID uuid.UUID, telegramID string) error {
	_, err := r.db.User.UpdateOneID(userID).
		SetTelegramID(telegramID).
		Save(ctx)
	return err
}

// ClearUserTelegram снимает telegram_id с профиля пользователя.
func (r *Repository) ClearUserTelegram(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.User.UpdateOneID(userID).
		ClearTelegramID().
		Save(ctx)
	return err
}
