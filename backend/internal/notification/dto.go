package notification

import (
	"time"

	"github.com/google/uuid"
)

// NotificationDTO — DTO уведомления для API-ответа.
type NotificationDTO struct {
	ID         uuid.UUID  `json:"id"`
	Type       string     `json:"type"`
	Title      string     `json:"title"`
	Body       string     `json:"body,omitempty"`
	EntityType string     `json:"entity_type,omitempty"`
	EntityID   string     `json:"entity_id,omitempty"`
	ProjectID  *uuid.UUID `json:"project_id,omitempty"`
	ActorID    *uuid.UUID `json:"actor_id,omitempty"`
	IsRead     bool       `json:"is_read"`
	IsUrgent   bool       `json:"is_urgent"`
	CreatedAt  time.Time  `json:"created_at"`
}

// NotificationListResponse — ответ со списком уведомлений с пагинацией.
type NotificationListResponse struct {
	Notifications []NotificationDTO `json:"notifications"`
	Total         int               `json:"total"`
	Page          int               `json:"page"`
	PerPage       int               `json:"per_page"`
}

// UnreadCountResponse — ответ с количеством непрочитанных уведомлений.
type UnreadCountResponse struct {
	Count int `json:"count"`
}

// PreferenceDTO — DTO настройки уведомлений.
type PreferenceDTO struct {
	ID               uuid.UUID  `json:"id"`
	NotificationType string     `json:"notification_type"`
	Channel          string     `json:"channel"`
	Enabled          bool       `json:"enabled"`
	ProjectID        *uuid.UUID `json:"project_id,omitempty"`
}

// PreferenceUpdateRequest — запрос на обновление настройки уведомлений.
type PreferenceUpdateRequest struct {
	NotificationType string     `json:"notification_type"`
	Channel          string     `json:"channel"`
	Enabled          bool       `json:"enabled"`
	ProjectID        *uuid.UUID `json:"project_id,omitempty"`
}

// MarkReadRequest — запрос на отметку уведомления прочитанным.
type MarkReadRequest struct {
	ID uuid.UUID `json:"id"`
}

// WsNotificationPayload — payload для отправки уведомления через WebSocket.
type WsNotificationPayload struct {
	ID         uuid.UUID  `json:"id"`
	Type       string     `json:"type"`
	Title      string     `json:"title"`
	Body       string     `json:"body,omitempty"`
	EntityType string     `json:"entity_type,omitempty"`
	EntityID   string     `json:"entity_id,omitempty"`
	ProjectID  *uuid.UUID `json:"project_id,omitempty"`
	ActorID    *uuid.UUID `json:"actor_id,omitempty"`
	IsUrgent   bool       `json:"is_urgent"`
	CreatedAt  time.Time  `json:"created_at"`
}
