package crm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/crmdeal"
)

var (
	ErrNotFound         = errors.New("crm entity not found")
	ErrInvalidInput     = errors.New("invalid crm input")
	ErrVersionConflict  = errors.New("crm version conflict")
	ErrProjectForbidden = errors.New("project access required")
	ErrForbidden        = errors.New("crm access denied")
)

// Service — бизнес-логика CRM-lite.
type Service struct {
	repo            Repository
	invoiceCreator  DealInvoiceCreator
}

// DealInvoiceCreator — draft invoice при won deal (V4-03.5).
type DealInvoiceCreator interface {
	CreateInvoiceFromWonDeal(ctx context.Context, userID, dealID uuid.UUID, customerName, title string, amountCents int64, currency string) error
}

// NewService создаёт сервис CRM.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// SetDealInvoiceCreator подключает accounting hook после init — избегаем циклической зависимости.
func (s *Service) SetDealInvoiceCreator(c DealInvoiceCreator) {
	s.invoiceCreator = c
}

// ListContacts — RT-003: member видит только свои контакты (created_by);
// system admin/founder — полный org-list (CRM-lite admin view).
func (s *Service) ListContacts(ctx context.Context, userID uuid.UUID) ([]ContactResponse, error) {
	scope, err := s.listOwnerScope(ctx, userID)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.ListContacts(ctx, scope)
	if err != nil {
		return nil, fmt.Errorf("list contacts: %w", err)
	}
	out := make([]ContactResponse, 0, len(items))
	for _, c := range items {
		out = append(out, toContactResponse(c))
	}
	return out, nil
}

func (s *Service) CreateContact(ctx context.Context, userID uuid.UUID, req CreateContactRequest) (*ContactResponse, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, ErrInvalidInput
	}
	if req.LinkedUserID != nil {
		ok, err := s.repo.UserExists(ctx, *req.LinkedUserID)
		if err != nil {
			return nil, fmt.Errorf("check linked user: %w", err)
		}
		if !ok {
			return nil, ErrInvalidInput
		}
	}
	c, err := s.repo.CreateContact(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("create contact: %w", err)
	}
	c, err = s.repo.GetContact(ctx, c.ID)
	if err != nil {
		return nil, fmt.Errorf("reload contact: %w", err)
	}
	resp := toContactResponse(c)
	return &resp, nil
}

func (s *Service) UpdateContact(ctx context.Context, userID, id uuid.UUID, req UpdateContactRequest) (*ContactResponse, error) {
	if _, err := s.authorizeContactWrite(ctx, userID, id); err != nil {
		return nil, err
	}
	if req.Version < 1 {
		return nil, ErrVersionConflict
	}
	if req.LinkedUserID != nil {
		ok, err := s.repo.UserExists(ctx, *req.LinkedUserID)
		if err != nil {
			return nil, fmt.Errorf("check linked user: %w", err)
		}
		if !ok {
			return nil, ErrInvalidInput
		}
	}
	c, err := s.repo.UpdateContact(ctx, id, req)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrVersionConflict
		}
		return nil, fmt.Errorf("update contact: %w", err)
	}
	c, err = s.repo.GetContact(ctx, c.ID)
	if err != nil {
		return nil, fmt.Errorf("reload contact: %w", err)
	}
	resp := toContactResponse(c)
	return &resp, nil
}

func (s *Service) DeleteContact(ctx context.Context, userID, id uuid.UUID) error {
	if _, err := s.authorizeContactWrite(ctx, userID, id); err != nil {
		return err
	}
	if err := s.repo.SoftDeleteContact(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("delete contact: %w", err)
	}
	return nil
}

// GetPipeline — RT-003: member видит только свои deals (owner_id);
// system admin/founder — полный org pipeline.
func (s *Service) GetPipeline(ctx context.Context, userID uuid.UUID) (*PipelineResponse, error) {
	scope, err := s.listOwnerScope(ctx, userID)
	if err != nil {
		return nil, err
	}
	deals, err := s.repo.ListDeals(ctx, scope)
	if err != nil {
		return nil, fmt.Errorf("list deals: %w", err)
	}

	byStage := make(map[DealStage][]DealResponse)
	for _, stage := range PipelineStages {
		byStage[stage] = []DealResponse{}
	}
	for _, d := range deals {
		stage := DealStage(d.Stage)
		byStage[stage] = append(byStage[stage], toDealResponse(d))
	}

	stages := make([]PipelineStageResponse, 0, len(PipelineStages))
	for _, stage := range PipelineStages {
		stages = append(stages, PipelineStageResponse{
			Stage: stage,
			Deals: byStage[stage],
		})
	}
	return &PipelineResponse{Stages: stages}, nil
}

func (s *Service) CreateDeal(ctx context.Context, userID uuid.UUID, req CreateDealRequest) (*DealResponse, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, ErrInvalidInput
	}
	if req.Stage != "" && !isValidStage(req.Stage) {
		return nil, ErrInvalidInput
	}
	if err := s.validateDealRefs(ctx, userID, req.ContactID, req.ProjectID); err != nil {
		return nil, err
	}
	d, err := s.repo.CreateDeal(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("create deal: %w", err)
	}
	d, err = s.repo.GetDeal(ctx, d.ID)
	if err != nil {
		return nil, fmt.Errorf("reload deal: %w", err)
	}
	resp := toDealResponse(d)
	return &resp, nil
}

func (s *Service) UpdateDeal(ctx context.Context, userID, id uuid.UUID, req UpdateDealRequest) (*DealResponse, error) {
	current, err := s.authorizeDealWrite(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if req.Version < 1 {
		return nil, ErrVersionConflict
	}
	if req.Stage != nil && !isValidStage(*req.Stage) {
		return nil, ErrInvalidInput
	}
	if err := s.validateDealRefs(ctx, userID, req.ContactID, req.ProjectID); err != nil {
		return nil, err
	}
	d, err := s.repo.UpdateDeal(ctx, id, req)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrVersionConflict
		}
		return nil, fmt.Errorf("update deal: %w", err)
	}
	d, err = s.repo.GetDeal(ctx, d.ID)
	if err != nil {
		return nil, fmt.Errorf("reload deal: %w", err)
	}
	if req.Stage != nil && *req.Stage == StageWon && current.Stage != crmdeal.StageWon {
		s.tryCreateInvoice(ctx, userID, d)
	}
	resp := toDealResponse(d)
	return &resp, nil
}

func (s *Service) DeleteDeal(ctx context.Context, userID, id uuid.UUID) error {
	if _, err := s.authorizeDealWrite(ctx, userID, id); err != nil {
		return err
	}
	if err := s.repo.SoftDeleteDeal(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("delete deal: %w", err)
	}
	return nil
}

// listOwnerScope — nil = org-wide (admin/founder); иначе filter по caller.
func (s *Service) listOwnerScope(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error) {
	ok, err := s.repo.IsSystemAdmin(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("check admin: %w", err)
	}
	if ok {
		return nil, nil
	}
	return &userID, nil
}

// authorizeContactWrite — owner (created_by) или system admin/founder (A6-001).
func (s *Service) authorizeContactWrite(ctx context.Context, actorID, contactID uuid.UUID) (*ent.CrmContact, error) {
	c, err := s.repo.GetContact(ctx, contactID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get contact: %w", err)
	}
	if c.Edges.CreatedBy != nil && c.Edges.CreatedBy.ID == actorID {
		return c, nil
	}
	ok, err := s.repo.IsSystemAdmin(ctx, actorID)
	if err != nil {
		return nil, fmt.Errorf("check admin: %w", err)
	}
	if ok {
		return c, nil
	}
	return nil, ErrForbidden
}

// authorizeDealWrite — owner сделки или system admin/founder (A6-001).
func (s *Service) authorizeDealWrite(ctx context.Context, actorID, dealID uuid.UUID) (*ent.CrmDeal, error) {
	d, err := s.repo.GetDeal(ctx, dealID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get deal: %w", err)
	}
	if d.Edges.Owner != nil && d.Edges.Owner.ID == actorID {
		return d, nil
	}
	ok, err := s.repo.IsSystemAdmin(ctx, actorID)
	if err != nil {
		return nil, fmt.Errorf("check admin: %w", err)
	}
	if ok {
		return d, nil
	}
	return nil, ErrForbidden
}

func (s *Service) validateDealRefs(ctx context.Context, userID uuid.UUID, contactID, projectID *uuid.UUID) error {
	if contactID != nil {
		// W7-A2-005: contact ref только owner-scope (или admin) — иначе member
		// привязывает чужой UUID к своей сделке и обходит RT-003 list-scope.
		if _, err := s.authorizeContactWrite(ctx, userID, *contactID); err != nil {
			if errors.Is(err, ErrNotFound) {
				return ErrInvalidInput
			}
			return err
		}
	}
	if projectID != nil {
		ok, err := s.repo.IsProjectMember(ctx, *projectID, userID)
		if err != nil {
			return fmt.Errorf("check project member: %w", err)
		}
		if !ok {
			return ErrProjectForbidden
		}
	}
	return nil
}

func isValidStage(stage DealStage) bool {
	for _, s := range PipelineStages {
		if s == stage {
			return true
		}
	}
	return false
}

func (s *Service) tryCreateInvoice(ctx context.Context, userID uuid.UUID, d *ent.CrmDeal) {
	if s.invoiceCreator == nil || d.AmountCents == nil || *d.AmountCents <= 0 {
		return
	}
	customer := d.Title
	if d.Edges.Contact != nil && strings.TrimSpace(d.Edges.Contact.Name) != "" {
		customer = d.Edges.Contact.Name
	}
	if err := s.invoiceCreator.CreateInvoiceFromWonDeal(ctx, userID, d.ID, customer, d.Title, *d.AmountCents, d.Currency); err != nil {
		slog.WarnContext(ctx, "crm deal invoice skipped", "deal_id", d.ID, "err", err)
	}
}
