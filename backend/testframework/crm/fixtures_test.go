package crm_test

import (
	"github.com/valerii/hub/backend/internal/crm"
)

type (
	Service              = crm.Service
	EntRepository        = crm.EntRepository
	Repository           = crm.Repository
	ContactResponse      = crm.ContactResponse
	DealResponse         = crm.DealResponse
	PipelineResponse     = crm.PipelineResponse
	CreateContactRequest = crm.CreateContactRequest
	UpdateContactRequest = crm.UpdateContactRequest
	CreateDealRequest    = crm.CreateDealRequest
	UpdateDealRequest    = crm.UpdateDealRequest
	DealStage            = crm.DealStage
)

const (
	StageLead        = crm.StageLead
	StageQualified   = crm.StageQualified
	StageProposal    = crm.StageProposal
	StageNegotiation = crm.StageNegotiation
	StageWon         = crm.StageWon
	StageLost        = crm.StageLost
)

var (
	NewService       = crm.NewService
	NewEntRepository = crm.NewEntRepository
	PipelineStages   = crm.PipelineStages

	ErrNotFound         = crm.ErrNotFound
	ErrInvalidInput     = crm.ErrInvalidInput
	ErrVersionConflict  = crm.ErrVersionConflict
	ErrProjectForbidden = crm.ErrProjectForbidden
	ErrForbidden        = crm.ErrForbidden
)

func mapServiceError(err error) error {
	return crm.ExportMapServiceError(err)
}
