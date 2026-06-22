package crm

import (
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

// DealStage — этап pipeline CRM.
type DealStage string

const (
	StageLead        DealStage = "lead"
	StageQualified   DealStage = "qualified"
	StageProposal    DealStage = "proposal"
	StageNegotiation DealStage = "negotiation"
	StageWon         DealStage = "won"
	StageLost        DealStage = "lost"
)

// PipelineStages — порядок колонок kanban.
var PipelineStages = []DealStage{
	StageLead,
	StageQualified,
	StageProposal,
	StageNegotiation,
	StageWon,
	StageLost,
}

// CreateContactRequest — создание контакта.
type CreateContactRequest struct {
	Name          string     `json:"name"`
	Email         string     `json:"email,omitempty"`
	Phone         string     `json:"phone,omitempty"`
	Company       string     `json:"company,omitempty"`
	Notes         string     `json:"notes,omitempty"`
	LinkedUserID  *uuid.UUID `json:"linked_user_id,omitempty"`
}

// UpdateContactRequest — обновление контакта.
type UpdateContactRequest struct {
	Name         *string    `json:"name,omitempty"`
	Email        *string    `json:"email,omitempty"`
	Phone        *string    `json:"phone,omitempty"`
	Company      *string    `json:"company,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
	LinkedUserID *uuid.UUID `json:"linked_user_id,omitempty"`
	Version      int        `json:"version"`
}

// ContactResponse — DTO контакта.
type ContactResponse struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Email        string     `json:"email,omitempty"`
	Phone        string     `json:"phone,omitempty"`
	Company      string     `json:"company,omitempty"`
	Notes        string     `json:"notes,omitempty"`
	LinkedUserID *uuid.UUID `json:"linked_user_id,omitempty"`
	Version      int        `json:"version"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CreateDealRequest — создание сделки.
type CreateDealRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Stage       DealStage  `json:"stage,omitempty"`
	AmountCents *int64     `json:"amount_cents,omitempty"`
	Currency    string     `json:"currency,omitempty"`
	ContactID   *uuid.UUID `json:"contact_id,omitempty"`
	ProjectID   *uuid.UUID `json:"project_id,omitempty"`
}

// UpdateDealRequest — обновление сделки.
type UpdateDealRequest struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	Stage       *DealStage `json:"stage,omitempty"`
	AmountCents *int64     `json:"amount_cents,omitempty"`
	Currency    *string    `json:"currency,omitempty"`
	ContactID   *uuid.UUID `json:"contact_id,omitempty"`
	ProjectID   *uuid.UUID `json:"project_id,omitempty"`
	Rank        *string    `json:"rank,omitempty"`
	Version     int        `json:"version"`
}

// DealResponse — DTO сделки.
type DealResponse struct {
	ID            uuid.UUID  `json:"id"`
	Title         string     `json:"title"`
	Description   string     `json:"description,omitempty"`
	Stage         DealStage  `json:"stage"`
	AmountCents   *int64     `json:"amount_cents,omitempty"`
	Currency      string     `json:"currency"`
	Rank          string     `json:"rank,omitempty"`
	ContactID     *uuid.UUID `json:"contact_id,omitempty"`
	ContactName   string     `json:"contact_name,omitempty"`
	ProjectID     *uuid.UUID `json:"project_id,omitempty"`
	ProjectKey    string     `json:"project_key,omitempty"`
	OwnerID       uuid.UUID  `json:"owner_id"`
	OwnerName     string     `json:"owner_name,omitempty"`
	Version       int        `json:"version"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// PipelineResponse — kanban-воронка.
type PipelineResponse struct {
	Stages []PipelineStageResponse `json:"stages"`
}

// PipelineStageResponse — колонка pipeline.
type PipelineStageResponse struct {
	Stage DealStage      `json:"stage"`
	Deals []DealResponse `json:"deals"`
}

func toContactResponse(c *ent.CrmContact) ContactResponse {
	resp := ContactResponse{
		ID:        c.ID,
		Name:      c.Name,
		Email:     c.Email,
		Phone:     c.Phone,
		Company:   c.Company,
		Notes:     c.Notes,
		Version:   c.Version,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
	if c.Edges.LinkedUser != nil {
		resp.LinkedUserID = &c.Edges.LinkedUser.ID
	}
	return resp
}

func toDealResponse(d *ent.CrmDeal) DealResponse {
	resp := DealResponse{
		ID:          d.ID,
		Title:       d.Title,
		Description: d.Description,
		Stage:       DealStage(d.Stage),
		AmountCents: d.AmountCents,
		Currency:    d.Currency,
		Rank:        d.Rank,
		Version:     d.Version,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
	if d.Edges.Owner != nil {
		resp.OwnerID = d.Edges.Owner.ID
		if d.Edges.Owner.DisplayName != "" {
			resp.OwnerName = d.Edges.Owner.DisplayName
		} else if d.Edges.Owner.Name != "" {
			resp.OwnerName = d.Edges.Owner.Name
		}
	}
	if d.Edges.Contact != nil {
		resp.ContactID = &d.Edges.Contact.ID
		resp.ContactName = d.Edges.Contact.Name
	}
	if d.Edges.Project != nil {
		resp.ProjectID = &d.Edges.Project.ID
		resp.ProjectKey = d.Edges.Project.Key
	}
	return resp
}
