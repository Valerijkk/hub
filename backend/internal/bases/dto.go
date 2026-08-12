package bases

import (
	"time"

	"github.com/google/uuid"
)

// ColumnType — typed field в Hub table.
type ColumnType string

const (
	ColumnText     ColumnType = "text"
	ColumnNumber   ColumnType = "number"
	ColumnDate     ColumnType = "date"
	ColumnSelect   ColumnType = "select"
	ColumnCheckbox ColumnType = "checkbox"
	ColumnURL      ColumnType = "url"
	ColumnUserRef  ColumnType = "user_ref"
)

// BaseResponse — сводка рабочего пространства.
type BaseResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Version     int       `json:"version"`
	OwnerID     uuid.UUID `json:"owner_id"`
	OwnerName   string    `json:"owner_name,omitempty"`
	TableCount  int       `json:"table_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableSummary — table в списке base.
type TableSummary struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	Version   int       `json:"version"`
}

// ColumnResponse — определение column.
type ColumnResponse struct {
	ID        uuid.UUID              `json:"id"`
	Name      string                 `json:"name"`
	ColType   ColumnType             `json:"col_type"`
	Options   map[string]interface{} `json:"options,omitempty"`
	SortOrder int                    `json:"sort_order"`
	Version   int                    `json:"version"`
}

// RecordResponse — строка table.
type RecordResponse struct {
	ID        uuid.UUID              `json:"id"`
	Cells     map[string]interface{} `json:"cells"`
	Version   int                    `json:"version"`
	CreatedBy *uuid.UUID             `json:"created_by,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// TableViewResponse — table с columns и records.
type TableViewResponse struct {
	BaseID   uuid.UUID        `json:"base_id"`
	Table    TableSummary     `json:"table"`
	Columns  []ColumnResponse `json:"columns"`
	Records  []RecordResponse `json:"records"`
	Total    int              `json:"total"`
	HasMore  bool             `json:"has_more"`
}

// BaseDetailResponse — base со списком tables.
type BaseDetailResponse struct {
	BaseResponse
	Tables []TableSummary `json:"tables"`
}

type CreateBaseRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type UpdateBaseRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Version     int     `json:"version,omitempty"`
}

type CreateTableRequest struct {
	Name string `json:"name"`
}

type UpdateTableRequest struct {
	Name      *string `json:"name,omitempty"`
	SortOrder *int    `json:"sort_order,omitempty"`
	Version   int     `json:"version,omitempty"`
}

type CreateColumnRequest struct {
	Name      string                 `json:"name"`
	ColType   ColumnType             `json:"col_type"`
	Options   map[string]interface{} `json:"options,omitempty"`
	SortOrder *int                   `json:"sort_order,omitempty"`
}

type UpdateColumnRequest struct {
	Name      *string                `json:"name,omitempty"`
	ColType   *ColumnType            `json:"col_type,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"`
	SortOrder *int                   `json:"sort_order,omitempty"`
	Version   int                    `json:"version,omitempty"`
}

type CreateRecordRequest struct {
	Cells map[string]interface{} `json:"cells"`
}

type UpdateRecordRequest struct {
	Cells   map[string]interface{} `json:"cells"`
	Version int                      `json:"version,omitempty"`
}
