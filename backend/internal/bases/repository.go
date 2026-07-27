package bases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/hubbase"
	"github.com/valerii/hub/backend/ent/hubcolumn"
	"github.com/valerii/hub/backend/ent/hubrecord"
	"github.com/valerii/hub/backend/ent/hubtable"
	"github.com/valerii/hub/backend/ent/user"
)

// MaxTableRecords — A13-002: hard cap records в GetTable / GetTableView.
// Без лимита мутации и view тянут полный table scan.
const MaxTableRecords = 500

// Repository — персистентность no-code баз (V4-04).
type Repository interface {
	CreateBase(ctx context.Context, ownerID uuid.UUID, req CreateBaseRequest) (*ent.HubBase, error)
	ListBases(ctx context.Context, ownerID uuid.UUID) ([]*ent.HubBase, error)
	GetBase(ctx context.Context, id uuid.UUID) (*ent.HubBase, error)
	UpdateBase(ctx context.Context, id uuid.UUID, req UpdateBaseRequest) (*ent.HubBase, error)
	SoftDeleteBase(ctx context.Context, id uuid.UUID) error

	CreateTable(ctx context.Context, baseID uuid.UUID, req CreateTableRequest) (*ent.HubTable, error)
	// GetTable возвращает таблицу с колонками и records (cap MaxTableRecords).
	GetTable(ctx context.Context, baseID, tableID uuid.UUID) (*ent.HubTable, error)
	// GetTableMeta возвращает таблицу только с колонками (без records) — write-path.
	GetTableMeta(ctx context.Context, baseID, tableID uuid.UUID) (*ent.HubTable, error)
	// GetRecord возвращает одну запись таблицы (для update/delete без полного scan).
	GetRecord(ctx context.Context, tableID, recordID uuid.UUID) (*ent.HubRecord, error)
	UpdateTable(ctx context.Context, tableID uuid.UUID, req UpdateTableRequest) (*ent.HubTable, error)
	SoftDeleteTable(ctx context.Context, tableID uuid.UUID) error

	CreateColumn(ctx context.Context, tableID uuid.UUID, req CreateColumnRequest) (*ent.HubColumn, error)
	UpdateColumn(ctx context.Context, columnID uuid.UUID, req UpdateColumnRequest) (*ent.HubColumn, error)
	SoftDeleteColumn(ctx context.Context, columnID uuid.UUID) error

	CreateRecord(ctx context.Context, tableID, userID uuid.UUID, cells map[string]interface{}) (*ent.HubRecord, error)
	UpdateRecord(ctx context.Context, recordID uuid.UUID, req UpdateRecordRequest) (*ent.HubRecord, error)
	SoftDeleteRecord(ctx context.Context, recordID uuid.UUID) error
}

// EntRepository — реализация Repository через Ent.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository создаёт Ent-репозиторий.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

func (r *EntRepository) CreateBase(ctx context.Context, ownerID uuid.UUID, req CreateBaseRequest) (*ent.HubBase, error) {
	return r.client.HubBase.Create().
		SetName(req.Name).
		SetDescription(req.Description).
		SetOwnerID(ownerID).
		Save(ctx)
}

func (r *EntRepository) ListBases(ctx context.Context, ownerID uuid.UUID) ([]*ent.HubBase, error) {
	// Владелец видит только свои базы: без owner-фильтра любой аутентифицированный
	// пользователь получал полный список чужих баз (аналог NF-03).
	return r.client.HubBase.Query().
		Where(
			hubbase.DeletedAtIsNil(),
			hubbase.HasOwnerWith(user.IDEQ(ownerID)),
		).
		WithOwner().
		WithTables(func(q *ent.HubTableQuery) {
			q.Where(hubtable.DeletedAtIsNil())
		}).
		Order(ent.Desc(hubbase.FieldUpdatedAt)).
		All(ctx)
}

func (r *EntRepository) GetBase(ctx context.Context, id uuid.UUID) (*ent.HubBase, error) {
	return r.client.HubBase.Query().
		Where(hubbase.ID(id), hubbase.DeletedAtIsNil()).
		WithOwner().
		WithTables(func(q *ent.HubTableQuery) {
			q.Where(hubtable.DeletedAtIsNil()).
				Order(ent.Asc(hubtable.FieldSortOrder), ent.Asc(hubtable.FieldName))
		}).
		Only(ctx)
}

func (r *EntRepository) UpdateBase(ctx context.Context, id uuid.UUID, req UpdateBaseRequest) (*ent.HubBase, error) {
	upd := r.client.HubBase.UpdateOneID(id).
		Where(hubbase.DeletedAtIsNil()).
		AddVersion(1)
	if req.Name != nil {
		upd.SetName(*req.Name)
	}
	if req.Description != nil {
		upd.SetDescription(*req.Description)
	}
	return upd.Save(ctx)
}

func (r *EntRepository) SoftDeleteBase(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.client.HubBase.UpdateOneID(id).
		Where(hubbase.DeletedAtIsNil()).
		SetDeletedAt(now).
		Save(ctx)
	return err
}

func (r *EntRepository) CreateTable(ctx context.Context, baseID uuid.UUID, req CreateTableRequest) (*ent.HubTable, error) {
	return r.client.HubTable.Create().
		SetName(req.Name).
		SetBaseID(baseID).
		Save(ctx)
}

func (r *EntRepository) GetTable(ctx context.Context, baseID, tableID uuid.UUID) (*ent.HubTable, error) {
	return r.client.HubTable.Query().
		Where(
			hubtable.ID(tableID),
			hubtable.DeletedAtIsNil(),
			hubtable.HasBaseWith(hubbase.ID(baseID), hubbase.DeletedAtIsNil()),
		).
		WithColumns(func(q *ent.HubColumnQuery) {
			q.Where(hubcolumn.DeletedAtIsNil()).
				Order(ent.Asc(hubcolumn.FieldSortOrder), ent.Asc(hubcolumn.FieldName))
		}).
		WithRecords(func(q *ent.HubRecordQuery) {
			q.Where(hubrecord.DeletedAtIsNil()).
				WithCreatedBy().
				Order(ent.Desc(hubrecord.FieldUpdatedAt)).
				Limit(MaxTableRecords)
		}).
		Only(ctx)
}

// GetTableMeta — columns only; write-path не должен грузить все records (A13-002).
func (r *EntRepository) GetTableMeta(ctx context.Context, baseID, tableID uuid.UUID) (*ent.HubTable, error) {
	return r.client.HubTable.Query().
		Where(
			hubtable.ID(tableID),
			hubtable.DeletedAtIsNil(),
			hubtable.HasBaseWith(hubbase.ID(baseID), hubbase.DeletedAtIsNil()),
		).
		WithColumns(func(q *ent.HubColumnQuery) {
			q.Where(hubcolumn.DeletedAtIsNil()).
				Order(ent.Asc(hubcolumn.FieldSortOrder), ent.Asc(hubcolumn.FieldName))
		}).
		Only(ctx)
}

// GetRecord загружает одну запись в рамках таблицы.
func (r *EntRepository) GetRecord(ctx context.Context, tableID, recordID uuid.UUID) (*ent.HubRecord, error) {
	return r.client.HubRecord.Query().
		Where(
			hubrecord.ID(recordID),
			hubrecord.DeletedAtIsNil(),
			hubrecord.HasTableWith(hubtable.ID(tableID), hubtable.DeletedAtIsNil()),
		).
		WithCreatedBy().
		Only(ctx)
}

func (r *EntRepository) UpdateTable(ctx context.Context, tableID uuid.UUID, req UpdateTableRequest) (*ent.HubTable, error) {
	upd := r.client.HubTable.UpdateOneID(tableID).
		Where(hubtable.DeletedAtIsNil()).
		AddVersion(1)
	if req.Name != nil {
		upd.SetName(*req.Name)
	}
	if req.SortOrder != nil {
		upd.SetSortOrder(*req.SortOrder)
	}
	return upd.Save(ctx)
}

func (r *EntRepository) SoftDeleteTable(ctx context.Context, tableID uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.client.HubTable.UpdateOneID(tableID).
		Where(hubtable.DeletedAtIsNil()).
		SetDeletedAt(now).
		Save(ctx)
	return err
}

func (r *EntRepository) CreateColumn(ctx context.Context, tableID uuid.UUID, req CreateColumnRequest) (*ent.HubColumn, error) {
	create := r.client.HubColumn.Create().
		SetName(req.Name).
		SetTableID(tableID)
	if req.ColType != "" {
		create.SetColType(hubcolumn.ColType(req.ColType))
	}
	if req.Options != nil {
		create.SetOptions(req.Options)
	}
	if req.SortOrder != nil {
		create.SetSortOrder(*req.SortOrder)
	}
	return create.Save(ctx)
}

func (r *EntRepository) UpdateColumn(ctx context.Context, columnID uuid.UUID, req UpdateColumnRequest) (*ent.HubColumn, error) {
	upd := r.client.HubColumn.UpdateOneID(columnID).
		Where(hubcolumn.DeletedAtIsNil()).
		AddVersion(1)
	if req.Name != nil {
		upd.SetName(*req.Name)
	}
	if req.ColType != nil {
		upd.SetColType(hubcolumn.ColType(*req.ColType))
	}
	if req.Options != nil {
		upd.SetOptions(req.Options)
	}
	if req.SortOrder != nil {
		upd.SetSortOrder(*req.SortOrder)
	}
	return upd.Save(ctx)
}

func (r *EntRepository) SoftDeleteColumn(ctx context.Context, columnID uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.client.HubColumn.UpdateOneID(columnID).
		Where(hubcolumn.DeletedAtIsNil()).
		SetDeletedAt(now).
		Save(ctx)
	return err
}

func (r *EntRepository) CreateRecord(ctx context.Context, tableID, userID uuid.UUID, cells map[string]interface{}) (*ent.HubRecord, error) {
	create := r.client.HubRecord.Create().
		SetTableID(tableID).
		SetCreatedByID(userID)
	if cells != nil {
		create.SetCells(cells)
	}
	return create.Save(ctx)
}

func (r *EntRepository) UpdateRecord(ctx context.Context, recordID uuid.UUID, req UpdateRecordRequest) (*ent.HubRecord, error) {
	upd := r.client.HubRecord.UpdateOneID(recordID).
		Where(hubrecord.DeletedAtIsNil()).
		AddVersion(1)
	if req.Cells != nil {
		upd.SetCells(req.Cells)
	}
	return upd.Save(ctx)
}

func (r *EntRepository) SoftDeleteRecord(ctx context.Context, recordID uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.client.HubRecord.UpdateOneID(recordID).
		Where(hubrecord.DeletedAtIsNil()).
		SetDeletedAt(now).
		Save(ctx)
	return err
}
