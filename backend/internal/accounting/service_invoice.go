package accounting

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/invoice"
)

var (
	ErrInvoiceExists  = errors.New("invoice already exists for deal")
	ErrInvalidInvoice = errors.New("invalid invoice state")
)

// CreateInvoiceFromWonDeal — CRM hook при won deal (non-blocking).
func (s *Service) CreateInvoiceFromWonDeal(
	ctx context.Context,
	userID, dealID uuid.UUID,
	customerName, title string,
	amountCents int64,
	currency string,
) error {
	if amountCents <= 0 || strings.TrimSpace(title) == "" {
		return nil
	}
	exists, err := s.repo.InvoiceExistsForDeal(ctx, dealID)
	if err != nil {
		return fmt.Errorf("check deal invoice: %w", err)
	}
	if exists {
		return nil
	}
	name := strings.TrimSpace(customerName)
	if name == "" {
		name = "Клиент"
	}
	req := CreateInvoiceRequest{
		CustomerName: name,
		Description:  strings.TrimSpace(title),
		AmountCents:  amountCents,
		Currency:     currency,
		CrmDealID:    &dealID,
	}
	_, err = s.createInvoice(ctx, userID, req)
	return err
}

func (s *Service) ListInvoices(ctx context.Context, userID uuid.UUID) ([]InvoiceResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListInvoices(ctx, 50)
	if err != nil {
		return nil, fmt.Errorf("list invoices: %w", err)
	}
	out := make([]InvoiceResponse, 0, len(items))
	for _, inv := range items {
		out = append(out, toInvoiceResponse(inv))
	}
	return out, nil
}

func (s *Service) CreateInvoice(ctx context.Context, userID uuid.UUID, req CreateInvoiceRequest) (*InvoiceResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return nil, err
	}
	inv, err := s.createInvoice(ctx, userID, req)
	if err != nil {
		return nil, err
	}
	resp := toInvoiceResponse(inv)
	return &resp, nil
}

func (s *Service) createInvoice(ctx context.Context, userID uuid.UUID, req CreateInvoiceRequest) (*ent.Invoice, error) {
	req.CustomerName = strings.TrimSpace(req.CustomerName)
	if req.CustomerName == "" || req.AmountCents <= 0 {
		return nil, ErrInvalidInput
	}
	if req.CrmDealID != nil {
		exists, err := s.repo.InvoiceExistsForDeal(ctx, *req.CrmDealID)
		if err != nil {
			return nil, fmt.Errorf("check deal invoice: %w", err)
		}
		if exists {
			return nil, ErrInvoiceExists
		}
	}
	inv, err := s.repo.CreateInvoice(ctx, userID, req, "")
	if err != nil {
		return nil, fmt.Errorf("create invoice: %w", err)
	}
	return s.repo.GetInvoice(ctx, inv.ID)
}

func (s *Service) IssueInvoice(ctx context.Context, userID, invoiceID uuid.UUID) (*InvoiceResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return nil, err
	}
	inv, err := s.repo.IssueInvoice(ctx, invoiceID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrInvalidInvoice
		}
		return nil, fmt.Errorf("issue invoice: %w", err)
	}
	resp := toInvoiceResponse(inv)
	return &resp, nil
}

func (s *Service) MarkInvoicePaid(ctx context.Context, userID, invoiceID uuid.UUID) (*InvoiceResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return nil, err
	}
	if err := s.repo.EnsureDefaultChart(ctx); err != nil {
		return nil, fmt.Errorf("ensure chart: %w", err)
	}
	inv, err := s.repo.GetInvoice(ctx, invoiceID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get invoice: %w", err)
	}
	if inv.Status != invoice.StatusIssued {
		return nil, ErrInvalidInvoice
	}
	bank, err := s.repo.GetAccountByCode(ctx, "1010")
	if err != nil {
		return nil, fmt.Errorf("bank account: %w", err)
	}
	revenue, err := s.repo.GetAccountByCode(ctx, "4010")
	if err != nil {
		return nil, fmt.Errorf("revenue account: %w", err)
	}
	desc := fmt.Sprintf("Оплата счёта %s", inv.Number)
	entryDate := time.Now().UTC()
	if inv.IssuedAt != nil {
		entryDate = inv.IssuedAt.UTC()
	}
	entryReq := CreateJournalRequest{
		EntryDate:   entryDate.Format("2006-01-02"),
		Description: desc,
		Reference:   inv.Number,
		Lines: []JournalLineInput{
			{AccountID: bank.ID, DebitCents: inv.AmountCents, CreditCents: 0, Memo: inv.CustomerName},
			{AccountID: revenue.ID, DebitCents: 0, CreditCents: inv.AmountCents},
		},
	}
	entry, err := s.repo.CreateJournalEntry(ctx, userID, entryDate, entryReq)
	if err != nil {
		return nil, fmt.Errorf("create payment journal: %w", err)
	}
	posted, err := s.repo.PostJournalEntry(ctx, entry.ID)
	if err != nil {
		return nil, fmt.Errorf("post payment journal: %w", err)
	}
	paid, err := s.repo.MarkInvoicePaid(ctx, invoiceID, posted.ID)
	if err != nil {
		return nil, fmt.Errorf("mark paid: %w", err)
	}
	resp := toInvoiceResponse(paid)
	return &resp, nil
}

func (s *Service) IncomeStatement(ctx context.Context, userID uuid.UUID) (*IncomeStatementResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return nil, err
	}
	if err := s.repo.EnsureDefaultChart(ctx); err != nil {
		return nil, fmt.Errorf("ensure chart: %w", err)
	}
	return s.repo.IncomeStatement(ctx)
}

func toInvoiceResponse(inv *ent.Invoice) InvoiceResponse {
	resp := InvoiceResponse{
		ID:           inv.ID,
		Number:       inv.Number,
		CustomerName: inv.CustomerName,
		Description:  inv.Description,
		AmountCents:  inv.AmountCents,
		Currency:     inv.Currency,
		Status:       InvoiceStatus(inv.Status),
		IssuedAt:     inv.IssuedAt,
		PaidAt:       inv.PaidAt,
		Version:      inv.Version,
		CreatedAt:    inv.CreatedAt,
	}
	if inv.DueDate != nil {
		s := inv.DueDate.UTC().Format("2006-01-02")
		resp.DueDate = &s
	}
	if inv.Edges.CrmDeal != nil {
		resp.CrmDealID = &inv.Edges.CrmDeal.ID
	}
	if inv.Edges.JournalEntry != nil {
		resp.JournalID = &inv.Edges.JournalEntry.ID
	}
	return resp
}
