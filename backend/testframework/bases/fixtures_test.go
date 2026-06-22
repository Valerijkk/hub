package bases_test

import (
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/bases"
)

type (
	Service            = bases.Service
	EntRepository      = bases.EntRepository
	Repository         = bases.Repository
	BaseResponse       = bases.BaseResponse
	BaseDetailResponse = bases.BaseDetailResponse
	TableSummary       = bases.TableSummary
	TableViewResponse  = bases.TableViewResponse
	ColumnResponse     = bases.ColumnResponse
	RecordResponse     = bases.RecordResponse
	CreateBaseRequest  = bases.CreateBaseRequest
	UpdateBaseRequest  = bases.UpdateBaseRequest
	CreateTableRequest = bases.CreateTableRequest
	UpdateTableRequest = bases.UpdateTableRequest
	CreateColumnRequest = bases.CreateColumnRequest
	UpdateColumnRequest = bases.UpdateColumnRequest
	CreateRecordRequest = bases.CreateRecordRequest
	UpdateRecordRequest = bases.UpdateRecordRequest
	ColumnType         = bases.ColumnType
)

const (
	ColumnText     = bases.ColumnText
	ColumnNumber   = bases.ColumnNumber
	ColumnDate     = bases.ColumnDate
	ColumnSelect   = bases.ColumnSelect
	ColumnCheckbox = bases.ColumnCheckbox
	ColumnURL      = bases.ColumnURL
	ColumnUserRef  = bases.ColumnUserRef
)

var (
	NewService       = bases.NewService
	NewEntRepository = bases.NewEntRepository

	ErrNotFound        = bases.ErrNotFound
	ErrInvalidInput    = bases.ErrInvalidInput
	ErrVersionConflict = bases.ErrVersionConflict
)

func mapServiceError(err error) error {
	return bases.ExportMapServiceError(err)
}

func normalizeCells(cells map[string]interface{}, columns []*ent.HubColumn) map[string]interface{} {
	return bases.ExportNormalizeCells(cells, columns)
}
