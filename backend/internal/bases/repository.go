package bases

import (
	"context"
	"fmt"
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
	// CountTableRecords — число живых records (для has_more после cap).
	CountTableRecords(ctx context.Context, tableID uuid.UUID) (int, error)
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

// CreateBase создаёт базу вместе с Main-таблицей и колонкой Name в одной tx (W2-A5-001).
func (r *EntRepository) CreateBase(ctx context.Context, ownerID uuid.UUID, req CreateBaseRequest) (*ent.HubBase, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin create base tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	b, err := tx.HubBase.Create().
		SetName(req.Name).
		SetDescription(req.Description).
		SetOwnerID(ownerID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create base: %w", err)
	}
	table, err := tx.HubTable.Create().
		SetName("Main").
		SetBaseID(b.ID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create default table: %w", err)
	}
	_, err = tx.HubColumn.Create().
		SetName("Name").
		SetColType(hubcolumn.ColTypeText).
		SetTableID(table.ID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create default column: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create base: %w", err)
	}
	return r.GetBase(ctx, b.ID)
}

// MaxOwnerBases — P3-BE-06 / W2-A5-012: hard cap списка баз владельца.
// Mutation: снять Limit → unbounded All при N>cap.
const MaxOwnerBases = 100

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
		Limit(MaxOwnerBases).
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
	// W2-A5-004: VersionEQ в WHERE — иначе TOCTOU app-check → lost update.
	upd := r.client.HubBase.UpdateOneID(id).
		Where(
			hubbase.DeletedAtIsNil(),
			hubbase.VersionEQ(req.Version),
		).
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
	// A2W2-5-001: cascade soft-delete tables/columns/records в одной tx.
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin soft-delete base tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC()

	// A3W2-A5-002: все tableIDs базы (вкл. уже soft-deleted) — иначе orphans
	// у таблицы, удалённой раньше через SoftDeleteTable без cascade детей.
	tableIDs, err := tx.HubTable.Query().
		Where(hubtable.HasBaseWith(hubbase.ID(id))).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("list tables for cascade soft-delete: %w", err)
	}

	if len(tableIDs) > 0 {
		if _, err := tx.HubRecord.Update().
			Where(
				hubrecord.HasTableWith(hubtable.IDIn(tableIDs...)),
				hubrecord.DeletedAtIsNil(),
			).
			SetDeletedAt(now).
			Save(ctx); err != nil {
			return fmt.Errorf("cascade soft-delete records: %w", err)
		}
		if _, err := tx.HubColumn.Update().
			Where(
				hubcolumn.HasTableWith(hubtable.IDIn(tableIDs...)),
				hubcolumn.DeletedAtIsNil(),
			).
			SetDeletedAt(now).
			Save(ctx); err != nil {
			return fmt.Errorf("cascade soft-delete columns: %w", err)
		}
		if _, err := tx.HubTable.Update().
			Where(
				hubtable.IDIn(tableIDs...),
				hubtable.DeletedAtIsNil(),
			).
			SetDeletedAt(now).
			Save(ctx); err != nil {
			return fmt.Errorf("cascade soft-delete tables: %w", err)
		}
	}

	affected, err := tx.HubBase.Update().
		Where(
			hubbase.IDEQ(id),
			hubbase.DeletedAtIsNil(),
		).
		SetDeletedAt(now).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return &ent.NotFoundError{}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit soft-delete base: %w", err)
	}
	return nil
}

func (r *EntRepository) CreateTable(ctx context.Context, baseID uuid.UUID, req CreateTableRequest) (*ent.HubTable, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin create table tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	table, err := tx.HubTable.Create().
		SetName(req.Name).
		SetBaseID(baseID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}
	_, err = tx.HubColumn.Create().
		SetName("Name").
		SetColType(hubcolumn.ColTypeText).
		SetTableID(table.ID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create default column: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create table: %w", err)
	}
	return table, nil
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

func (r *EntRepository) CountTableRecords(ctx context.Context, tableID uuid.UUID) (int, error) {
	return r.client.HubRecord.Query().
		Where(
			hubrecord.DeletedAtIsNil(),
			hubrecord.HasTableWith(hubtable.ID(tableID), hubtable.DeletedAtIsNil()),
		).
		Count(ctx)
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
		Where(
			hubtable.DeletedAtIsNil(),
			hubtable.VersionEQ(req.Version),
		).
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
	// A3W2-A5-002: симметрия SoftDeleteBase — soft-delete columns+records в одной tx.
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin soft-delete table tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC()
	if _, err := tx.HubRecord.Update().
		Where(
			hubrecord.HasTableWith(hubtable.ID(tableID)),
			hubrecord.DeletedAtIsNil(),
		).
		SetDeletedAt(now).
		Save(ctx); err != nil {
		return fmt.Errorf("cascade soft-delete table records: %w", err)
	}
	if _, err := tx.HubColumn.Update().
		Where(
			hubcolumn.HasTableWith(hubtable.ID(tableID)),
			hubcolumn.DeletedAtIsNil(),
		).
		SetDeletedAt(now).
		Save(ctx); err != nil {
		return fmt.Errorf("cascade soft-delete table columns: %w", err)
	}
	affected, err := tx.HubTable.Update().
		Where(
			hubtable.IDEQ(tableID),
			hubtable.DeletedAtIsNil(),
		).
		SetDeletedAt(now).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return &ent.NotFoundError{}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit soft-delete table: %w", err)
	}
	return nil
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
		Where(
			hubcolumn.DeletedAtIsNil(),
			hubcolumn.VersionEQ(req.Version),
		).
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
	// W2-A5-004: VersionEQ в WHERE — иначе concurrent PUT → lost update.
	upd := r.client.HubRecord.UpdateOneID(recordID).
		Where(
			hubrecord.DeletedAtIsNil(),
			hubrecord.VersionEQ(req.Version),
		).
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
