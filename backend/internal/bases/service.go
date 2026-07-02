package bases

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

var (
	ErrNotFound        = errors.New("base entity not found")
	ErrForbidden       = errors.New("base access denied")
	ErrInvalidInput    = errors.New("invalid base input")
	ErrVersionConflict = errors.New("base version conflict")
)

// Service — логика no-code БД (V4-04).
type Service struct {
	repo Repository
}

// NewService создаёт bases service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListBases(ctx context.Context, userID uuid.UUID) ([]BaseResponse, error) {
	items, err := s.repo.ListBases(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list bases: %w", err)
	}
	out := make([]BaseResponse, 0, len(items))
	for _, b := range items {
		out = append(out, toBaseResponse(b))
	}
	return out, nil
}

// authorizeBase проверяет право доступа к базе: без этой проверки любой
// аутентифицированный пользователь мог читать и править чужие базы (аналог NF-03).
// Возвращает ErrForbidden для чужой базы, ErrNotFound — для несуществующей.
func (s *Service) authorizeBase(ctx context.Context, baseID, userID uuid.UUID) (*ent.HubBase, error) {
	b, err := s.repo.GetBase(ctx, baseID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get base: %w", err)
	}
	if b.Edges.Owner == nil || b.Edges.Owner.ID != userID {
		return nil, ErrForbidden
	}
	return b, nil
}

func (s *Service) GetBase(ctx context.Context, userID, id uuid.UUID) (*BaseDetailResponse, error) {
	b, err := s.authorizeBase(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	resp := toBaseResponse(b)
	tables := make([]TableSummary, 0, len(b.Edges.Tables))
	for _, t := range b.Edges.Tables {
		tables = append(tables, toTableSummary(t))
	}
	return &BaseDetailResponse{BaseResponse: resp, Tables: tables}, nil
}

func (s *Service) CreateBase(ctx context.Context, ownerID uuid.UUID, req CreateBaseRequest) (*BaseDetailResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrInvalidInput
	}
	b, err := s.repo.CreateBase(ctx, ownerID, CreateBaseRequest{
		Name:        name,
		Description: strings.TrimSpace(req.Description),
	})
	if err != nil {
		return nil, fmt.Errorf("create base: %w", err)
	}

	table, err := s.repo.CreateTable(ctx, b.ID, CreateTableRequest{Name: "Main"})
	if err != nil {
		return nil, fmt.Errorf("create default table: %w", err)
	}
	_, err = s.repo.CreateColumn(ctx, table.ID, CreateColumnRequest{
		Name:    "Name",
		ColType: ColumnText,
	})
	if err != nil {
		return nil, fmt.Errorf("create default column: %w", err)
	}

	return s.GetBase(ctx, ownerID, b.ID)
}

func (s *Service) UpdateBase(ctx context.Context, userID, id uuid.UUID, req UpdateBaseRequest) (*BaseResponse, error) {
	current, err := s.authorizeBase(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if req.Version > 0 && current.Version != req.Version {
		return nil, ErrVersionConflict
	}
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return nil, ErrInvalidInput
	}
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		req.Name = &trimmed
	}
	if _, err := s.repo.UpdateBase(ctx, id, req); err != nil {
		return nil, fmt.Errorf("update base: %w", err)
	}
	b, err := s.repo.GetBase(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("reload base: %w", err)
	}
	resp := toBaseResponse(b)
	return &resp, nil
}

func (s *Service) DeleteBase(ctx context.Context, userID, id uuid.UUID) error {
	if _, err := s.authorizeBase(ctx, id, userID); err != nil {
		return err
	}
	return s.repo.SoftDeleteBase(ctx, id)
}

func (s *Service) CreateTable(ctx context.Context, userID, baseID uuid.UUID, req CreateTableRequest) (*TableSummary, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrInvalidInput
	}
	if _, err := s.authorizeBase(ctx, baseID, userID); err != nil {
		return nil, err
	}
	t, err := s.repo.CreateTable(ctx, baseID, CreateTableRequest{Name: name})
	if err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}
	_, err = s.repo.CreateColumn(ctx, t.ID, CreateColumnRequest{
		Name:    "Name",
		ColType: ColumnText,
	})
	if err != nil {
		return nil, fmt.Errorf("create default column: %w", err)
	}
	summary := toTableSummary(t)
	return &summary, nil
}

func (s *Service) GetTableView(ctx context.Context, userID, baseID, tableID uuid.UUID) (*TableViewResponse, error) {
	if _, err := s.authorizeBase(ctx, baseID, userID); err != nil {
		return nil, err
	}
	t, err := s.repo.GetTable(ctx, baseID, tableID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get table: %w", err)
	}
	cols := make([]ColumnResponse, 0, len(t.Edges.Columns))
	for _, c := range t.Edges.Columns {
		cols = append(cols, toColumnResponse(c))
	}
	recs := make([]RecordResponse, 0, len(t.Edges.Records))
	for _, rec := range t.Edges.Records {
		recs = append(recs, toRecordResponse(rec))
	}
	return &TableViewResponse{
		BaseID:  baseID,
		Table:   toTableSummary(t),
		Columns: cols,
		Records: recs,
	}, nil
}

func (s *Service) CreateColumn(ctx context.Context, userID, baseID, tableID uuid.UUID, req CreateColumnRequest) (*ColumnResponse, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, ErrInvalidInput
	}
	if _, err := s.authorizeBase(ctx, baseID, userID); err != nil {
		return nil, err
	}
	t, err := s.repo.GetTable(ctx, baseID, tableID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get table: %w", err)
	}
	_ = t
	req.Name = strings.TrimSpace(req.Name)
	if req.ColType == "" {
		req.ColType = ColumnText
	}
	c, err := s.repo.CreateColumn(ctx, tableID, req)
	if err != nil {
		return nil, fmt.Errorf("create column: %w", err)
	}
	resp := toColumnResponse(c)
	return &resp, nil
}

func (s *Service) CreateRecord(ctx context.Context, baseID, tableID, userID uuid.UUID, req CreateRecordRequest) (*RecordResponse, error) {
	if _, err := s.authorizeBase(ctx, baseID, userID); err != nil {
		return nil, err
	}
	t, err := s.repo.GetTable(ctx, baseID, tableID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get table: %w", err)
	}
	cells := normalizeCells(req.Cells, t.Edges.Columns)
	rec, err := s.repo.CreateRecord(ctx, tableID, userID, cells)
	if err != nil {
		return nil, fmt.Errorf("create record: %w", err)
	}
	resp := toRecordResponse(rec)
	return &resp, nil
}

func (s *Service) UpdateRecord(ctx context.Context, userID, baseID, tableID, recordID uuid.UUID, req UpdateRecordRequest) (*RecordResponse, error) {
	if _, err := s.authorizeBase(ctx, baseID, userID); err != nil {
		return nil, err
	}
	t, err := s.repo.GetTable(ctx, baseID, tableID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get table: %w", err)
	}
	var current *ent.HubRecord
	for _, r := range t.Edges.Records {
		if r.ID == recordID {
			current = r
			break
		}
	}
	if current == nil {
		return nil, ErrNotFound
	}
	if req.Version > 0 && current.Version != req.Version {
		return nil, ErrVersionConflict
	}
	cells := normalizeCells(req.Cells, t.Edges.Columns)
	rec, err := s.repo.UpdateRecord(ctx, recordID, UpdateRecordRequest{Cells: cells})
	if err != nil {
		return nil, fmt.Errorf("update record: %w", err)
	}
	resp := toRecordResponse(rec)
	return &resp, nil
}

func (s *Service) DeleteRecord(ctx context.Context, userID, baseID, tableID, recordID uuid.UUID) error {
	if _, err := s.authorizeBase(ctx, baseID, userID); err != nil {
		return err
	}
	t, err := s.repo.GetTable(ctx, baseID, tableID)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("get table: %w", err)
	}
	found := false
	for _, r := range t.Edges.Records {
		if r.ID == recordID {
			found = true
			break
		}
	}
	if !found {
		return ErrNotFound
	}
	return s.repo.SoftDeleteRecord(ctx, recordID)
}

func normalizeCells(cells map[string]interface{}, columns []*ent.HubColumn) map[string]interface{} {
	if cells == nil {
		cells = map[string]interface{}{}
	}
	allowed := make(map[string]struct{}, len(columns))
	for _, c := range columns {
		allowed[c.ID.String()] = struct{}{}
	}
	out := make(map[string]interface{}, len(cells))
	for k, v := range cells {
		if _, ok := allowed[k]; ok {
			out[k] = v
		}
	}
	return out
}

func toBaseResponse(b *ent.HubBase) BaseResponse {
	resp := BaseResponse{
		ID:          b.ID,
		Name:        b.Name,
		Description: b.Description,
		Version:     b.Version,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
	if b.Edges.Owner != nil {
		resp.OwnerID = b.Edges.Owner.ID
		resp.OwnerName = b.Edges.Owner.DisplayName
		if resp.OwnerName == "" {
			resp.OwnerName = b.Edges.Owner.Name
		}
	}
	if b.Edges.Tables != nil {
		resp.TableCount = len(b.Edges.Tables)
	}
	return resp
}

func toTableSummary(t *ent.HubTable) TableSummary {
	return TableSummary{
		ID:        t.ID,
		Name:      t.Name,
		SortOrder: t.SortOrder,
		Version:   t.Version,
	}
}

func toColumnResponse(c *ent.HubColumn) ColumnResponse {
	return ColumnResponse{
		ID:        c.ID,
		Name:      c.Name,
		ColType:   ColumnType(c.ColType),
		Options:   c.Options,
		SortOrder: c.SortOrder,
		Version:   c.Version,
	}
}

func toRecordResponse(r *ent.HubRecord) RecordResponse {
	resp := RecordResponse{
		ID:        r.ID,
		Cells:     r.Cells,
		Version:   r.Version,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
	if r.Edges.CreatedBy != nil {
		id := r.Edges.CreatedBy.ID
		resp.CreatedBy = &id
	}
	if resp.Cells == nil {
		resp.Cells = map[string]interface{}{}
	}
	return resp
}
