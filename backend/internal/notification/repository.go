package notification

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/notification"
	"github.com/valerii/hub/backend/ent/notificationpreference"
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

// IsEnabled проверяет, включено ли уведомление данного типа для канала и пользователя.
// Сначала проверяет per-project настройку, потом глобальную. По умолчанию — включено.
func (r *Repository) IsEnabled(ctx context.Context, userID uuid.UUID, notifType, channel string, projectID *uuid.UUID) (bool, error) {
	// Проверяем per-project настройку.
	if projectID != nil {
		pref, err := r.db.NotificationPreference.Query().
			Where(
				notificationpreference.UserID(userID),
				notificationpreference.NotificationType(notifType),
				notificationpreference.Channel(channel),
				notificationpreference.ProjectID(*projectID),
			).
			First(ctx)
		if err == nil {
			return pref.Enabled, nil
		}
		if !ent.IsNotFound(err) {
			return false, err
		}
	}

	// Проверяем глобальную настройку.
	pref, err := r.db.NotificationPreference.Query().
		Where(
			notificationpreference.UserID(userID),
			notificationpreference.NotificationType(notifType),
			notificationpreference.Channel(channel),
			notificationpreference.ProjectIDIsNil(),
		).
		First(ctx)
	if err == nil {
		return pref.Enabled, nil
	}
	if ent.IsNotFound(err) {
		return true, nil // По умолчанию включено.
	}
	return false, err
}
