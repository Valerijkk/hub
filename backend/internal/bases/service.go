package bases

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

var (
	ErrNotFound        = errors.New("base entity not found")
	ErrForbidden       = errors.New("base access denied")
	ErrInvalidInput    = errors.New("invalid base input")
	ErrVersionConflict = errors.New("base version conflict")
)

// MaxNameRunes — зеркало Ent MaxLen(200) на service-слое (A3W2-A5-003 → 400, не 500).
const MaxNameRunes = 200

func validateNameLen(name string) error {
	if utf8.RuneCountInString(name) > MaxNameRunes {
		return ErrInvalidInput
	}
	return nil
}

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
	if err := validateNameLen(name); err != nil {
		return nil, err
	}
	// W2-A5-001: base+Main+Name атомарно в repo.CreateBase (одна Ent tx).
	b, err := s.repo.CreateBase(ctx, ownerID, CreateBaseRequest{
		Name:        name,
		Description: strings.TrimSpace(req.Description),
	})
	if err != nil {
		return nil, fmt.Errorf("create base: %w", err)
	}
	return s.GetBase(ctx, ownerID, b.ID)
}

func (s *Service) UpdateBase(ctx context.Context, userID, id uuid.UUID, req UpdateBaseRequest) (*BaseResponse, error) {
	if _, err := s.authorizeBase(ctx, id, userID); err != nil {
		return nil, err
	}
	// W2-A5-004 / P3-BE-09: version обязателен (400); VersionEQ в repo WHERE.
	if req.Version < 1 {
		return nil, ErrInvalidInput
	}
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return nil, ErrInvalidInput
	}
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if err := validateNameLen(trimmed); err != nil {
			return nil, err
		}
		req.Name = &trimmed
	}
	if _, err := s.repo.UpdateBase(ctx, id, req); err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrVersionConflict
		}
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
	if err := validateNameLen(name); err != nil {
		return nil, err
	}
	if _, err := s.authorizeBase(ctx, baseID, userID); err != nil {
		return nil, err
	}
	// W2-A5-001 residual: table+Name column атомарно в repo.CreateTable.
	t, err := s.repo.CreateTable(ctx, baseID, CreateTableRequest{Name: name})
	if err != nil {
		return nil, fmt.Errorf("create table: %w", err)
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
	// W2-A5-005: has_more/total — иначе silent truncate при N>MaxTableRecords.
	total, err := s.repo.CountTableRecords(ctx, tableID)
	if err != nil {
		return nil, fmt.Errorf("count table records: %w", err)
	}
	return &TableViewResponse{
		BaseID:  baseID,
		Table:   toTableSummary(t),
		Columns: cols,
		Records: recs,
		Total:   total,
		HasMore: total > MaxTableRecords,
	}, nil
}

func (s *Service) CreateColumn(ctx context.Context, userID, baseID, tableID uuid.UUID, req CreateColumnRequest) (*ColumnResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrInvalidInput
	}
	if err := validateNameLen(name); err != nil {
		return nil, err
	}
	if _, err := s.authorizeBase(ctx, baseID, userID); err != nil {
		return nil, err
	}
	// A13-002: meta only — column create не нуждается в records.
	if _, err := s.repo.GetTableMeta(ctx, baseID, tableID); err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get table meta: %w", err)
	}
	req.Name = name
	if req.ColType == "" {
		req.ColType = ColumnText
	}
	// P3-BE-09 / W2-A5-008: allowlist ColType → 400, не 500 от Ent.
	if !isValidColumnType(req.ColType) {
		return nil, ErrInvalidInput
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
	// A13-002: write-path использует GetTableMeta, не полный GetTable.
	t, err := s.repo.GetTableMeta(ctx, baseID, tableID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get table meta: %w", err)
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
	t, err := s.repo.GetTableMeta(ctx, baseID, tableID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get table meta: %w", err)
	}
	current, err := s.repo.GetRecord(ctx, tableID, recordID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get record: %w", err)
	}
	// W2-A5-004 / P3-BE-09: version обязателен (400); VersionEQ в repo.
	if req.Version < 1 {
		return nil, ErrInvalidInput
	}
	patch := normalizeCells(req.Cells, t.Edges.Columns)
	// W2-A5-002: merge с current.Cells — partial PUT не затирает остальные ячейки.
	merged := make(map[string]interface{}, len(current.Cells)+len(patch))
	for k, v := range current.Cells {
		merged[k] = v
	}
	for k, v := range patch {
		merged[k] = v
	}
	rec, err := s.repo.UpdateRecord(ctx, recordID, UpdateRecordRequest{Cells: merged, Version: req.Version})
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrVersionConflict
		}
		return nil, fmt.Errorf("update record: %w", err)
	}
	resp := toRecordResponse(rec)
	return &resp, nil
}

func (s *Service) DeleteRecord(ctx context.Context, userID, baseID, tableID, recordID uuid.UUID) error {
	if _, err := s.authorizeBase(ctx, baseID, userID); err != nil {
		return err
	}
	if _, err := s.repo.GetTableMeta(ctx, baseID, tableID); err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("get table meta: %w", err)
	}
	if _, err := s.repo.GetRecord(ctx, tableID, recordID); err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("get record: %w", err)
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

// isValidColumnType — allowlist ColType (P3-BE-09 / W2-A5-008).
func isValidColumnType(t ColumnType) bool {
	switch t {
	case ColumnText, ColumnNumber, ColumnDate, ColumnSelect, ColumnCheckbox, ColumnURL, ColumnUserRef:
		return true
	default:
		return false
	}
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
