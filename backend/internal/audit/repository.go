package audit

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/auditlog"
)

// Repository предоставляет доступ к данным журнала аудита через Ent ORM.
type Repository struct {
	client *ent.Client
}

// NewRepository создаёт новый репозиторий журнала аудита.
func NewRepository(client *ent.Client) *Repository {
	return &Repository{client: client}
}

// CreateAuditLog записывает новую запись в журнал аудита.
func (r *Repository) CreateAuditLog(ctx context.Context, action, entityType, entityID string, userID, projectID *uuid.UUID, details map[string]interface{}, ipAddress, userAgent string) (*ent.AuditLog, error) {
	create := r.client.AuditLog.Create().
		SetAction(action).
		SetEntityType(entityType)

	if entityID != "" {
		create = create.SetEntityID(entityID)
	}
	if userID != nil {
		create = create.SetUserID(*userID)
	}
	if projectID != nil {
		create = create.SetProjectID(*projectID)
	}
	if details != nil {
		create = create.SetDetails(details)
	}
	if ipAddress != "" {
		create = create.SetIPAddress(ipAddress)
	}
	if userAgent != "" {
		create = create.SetUserAgent(userAgent)
	}

	return create.Save(ctx)
}

// ListAuditLogs возвращает пагинированный список записей аудита с фильтрами.
func (r *Repository) ListAuditLogs(ctx context.Context, filters AuditLogFilters) ([]*ent.AuditLog, int, error) {
	query := r.client.AuditLog.Query()

	if filters.Action != "" {
		query = query.Where(auditlog.ActionEQ(filters.Action))
	}
	if filters.EntityType != "" {
		query = query.Where(auditlog.EntityTypeEQ(filters.EntityType))
	}
	if filters.UserID != nil {
		query = query.Where(auditlog.UserIDEQ(*filters.UserID))
	}
	if filters.ProjectID != nil {
		query = query.Where(auditlog.ProjectIDEQ(*filters.ProjectID))
	}
	if filters.DateFrom != nil {
		query = query.Where(auditlog.CreatedAtGTE(*filters.DateFrom))
	}
	if filters.DateTo != nil {
		query = query.Where(auditlog.CreatedAtLTE(*filters.DateTo))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("подсчёт записей аудита: %w", err)
	}

	page := filters.Page
	if page < 1 {
		page = 1
	}
	perPage := filters.PerPage
	if perPage <= 0 || perPage > 100 {
		perPage = 50
	}
	offset := (page - 1) * perPage

	items, err := query.
		Order(ent.Desc(auditlog.FieldCreatedAt)).
		Offset(offset).
		Limit(perPage).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("список записей аудита: %w", err)
	}

	return items, total, nil
}
