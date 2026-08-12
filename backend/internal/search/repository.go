package search

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/savedfilter"
	"github.com/valerii/hub/backend/ent/user"
)

// Repository абстрагирует Ent-зависимые операции SearchService.
//
// SC-14-3: SearchService больше не хранит *ent.Client; всё, что требует
// типизированных Ent-запросов (членство в проектах через ent.User,
// SystemRole для админского обхода project-scoping, SavedFilter CRUD)
// уходит за этот интерфейс. Raw SQL (tsvector) и MongoDB ($text) живут
// в самом сервисе — это БД-уровневые оптимизации, не ORM-уровневые,
// и репозиторий-абстракция там избыточна.
type Repository interface {
	// IsSystemAdmin возвращает true, если у пользователя SystemRole=admin.
	// Используется searchUsers для обхода project-scoping (HI-12 follow-through).
	IsSystemAdmin(ctx context.Context, userID uuid.UUID) (bool, error)

	// IsProjectMember возвращает true, если userID — участник projectID.
	// A4-003/A4-004: CreateSavedFilter / ListSavedFilters требуют membership.
	IsProjectMember(ctx context.Context, userID, projectID uuid.UUID) (bool, error)

	// CreateSavedFilter создаёт новый сохранённый фильтр.
	CreateSavedFilter(ctx context.Context, input CreateSavedFilterInput) (*ent.SavedFilter, error)
	// ListSavedFilters возвращает фильтры в проекте: либо принадлежащие userID, либо shared.
	ListSavedFilters(ctx context.Context, userID, projectID uuid.UUID) ([]*ent.SavedFilter, error)
	// GetSavedFilter возвращает фильтр по ID.
	GetSavedFilter(ctx context.Context, filterID uuid.UUID) (*ent.SavedFilter, error)
	// DeleteSavedFilter удаляет фильтр по ID.
	DeleteSavedFilter(ctx context.Context, filterID uuid.UUID) error
	// UpdateSavedFilterShared изменяет флаг расшаривания фильтра.
	UpdateSavedFilterShared(ctx context.Context, filterID uuid.UUID, shared bool) (*ent.SavedFilter, error)
}

// CreateSavedFilterInput — данные для создания SavedFilter.
type CreateSavedFilterInput struct {
	Name         string
	FilterConfig map[string]interface{}
	OwnerID      uuid.UUID
	ProjectID    uuid.UUID
}

// EntRepository — Ent-реализация Repository.
type EntRepository struct {
	db *ent.Client
}

// NewEntRepository создаёт новый Ent-репозиторий для search.
func NewEntRepository(db *ent.Client) *EntRepository {
	return &EntRepository{db: db}
}

// IsSystemAdmin реализует Repository.
func (r *EntRepository) IsSystemAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	u, err := r.db.User.Get(ctx, userID)
	if err != nil {
		return false, err
	}
	return u.SystemRole == user.SystemRoleFounder || u.SystemRole == user.SystemRoleAdmin, nil
}

// IsProjectMember реализует Repository.
func (r *EntRepository) IsProjectMember(ctx context.Context, userID, projectID uuid.UUID) (bool, error) {
	exists, err := r.db.ProjectMember.Query().
		Where(
			projectmember.HasProjectWith(project.IDEQ(projectID)),
			projectmember.HasUserWith(user.IDEQ(userID)),
		).
		Exist(ctx)
	if err != nil {
		return false, fmt.Errorf("project membership: %w", err)
	}
	return exists, nil
}

// CreateSavedFilter реализует Repository.
func (r *EntRepository) CreateSavedFilter(ctx context.Context, input CreateSavedFilterInput) (*ent.SavedFilter, error) {
	return r.db.SavedFilter.Create().
		SetName(input.Name).
		SetFilterConfig(input.FilterConfig).
		SetOwnerID(input.OwnerID).
		SetProjectID(input.ProjectID).
		Save(ctx)
}

// ListSavedFilters реализует Repository.
func (r *EntRepository) ListSavedFilters(ctx context.Context, userID, projectID uuid.UUID) ([]*ent.SavedFilter, error) {
	return r.db.SavedFilter.Query().
		Where(
			savedfilter.ProjectID(projectID),
			savedfilter.Or(
				savedfilter.OwnerID(userID),
				savedfilter.Shared(true),
			),
		).
		All(ctx)
}

// GetSavedFilter реализует Repository.
func (r *EntRepository) GetSavedFilter(ctx context.Context, filterID uuid.UUID) (*ent.SavedFilter, error) {
	return r.db.SavedFilter.Get(ctx, filterID)
}

// DeleteSavedFilter реализует Repository.
func (r *EntRepository) DeleteSavedFilter(ctx context.Context, filterID uuid.UUID) error {
	return r.db.SavedFilter.DeleteOneID(filterID).Exec(ctx)
}

// UpdateSavedFilterShared реализует Repository.
func (r *EntRepository) UpdateSavedFilterShared(ctx context.Context, filterID uuid.UUID, shared bool) (*ent.SavedFilter, error) {
	return r.db.SavedFilter.UpdateOneID(filterID).
		SetShared(shared).
		Save(ctx)
}
