package audit

import (
	"time"

	"github.com/google/uuid"
)

// AuditLogResponse — ответ с записью журнала аудита.
type AuditLogResponse struct {
	ID         string                 `json:"id"`
	Action     string                 `json:"action"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id,omitempty"`
	UserID     *string                `json:"user_id,omitempty"`
	ProjectID  *string                `json:"project_id,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// AuditLogFilters — фильтры для запроса журнала аудита.
type AuditLogFilters struct {
	Action     string
	EntityType string
	UserID     *uuid.UUID
	ProjectID  *uuid.UUID
	DateFrom   *time.Time
	DateTo     *time.Time
	Page       int
	PerPage    int
}

// AuditLogListResponse — пагинированный список записей аудита.
type AuditLogListResponse struct {
	Items   []AuditLogResponse `json:"items"`
	Total   int                `json:"total"`
	Page    int                `json:"page"`
	PerPage int                `json:"per_page"`
}

// ResourceLimitResponse — текущий лимит ресурса.
type ResourceLimitResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// UpdateLimitRequest — запрос на обновление лимита.
type UpdateLimitRequest struct {
	Value string `json:"value"`
}

// TrashItemResponse — элемент корзины.
type TrashItemResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	EntityType    string    `json:"entity_type"`
	DeletedAt     time.Time `json:"deleted_at"`
	DaysRemaining int       `json:"days_remaining"`
}

// TrashListResponse — список элементов корзины с пагинацией.
type TrashListResponse struct {
	Items   []TrashItemResponse `json:"items"`
	Total   int                 `json:"total"`
	Page    int                 `json:"page"`
	PerPage int                 `json:"per_page"`
}

// ReassignRequest — запрос на переназначение задач пользователя.
type ReassignRequest struct {
	ToUserID string `json:"to_user_id"`
}

// UserManagementResponse — ответ на операцию управления пользователем.
type UserManagementResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
