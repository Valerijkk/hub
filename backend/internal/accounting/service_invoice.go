package accounting

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
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
	_, err := s.createInvoice(ctx, userID, req)
	// A6-004: concurrent won-hook / HTTP create — duplicate = idempotent success.
	if errors.Is(err, ErrInvoiceExists) {
		return nil
	}
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
	inv, err := s.repo.CreateInvoice(ctx, userID, req, "")
	if err != nil {
		if errors.Is(err, ErrInvoiceExists) {
			return nil, ErrInvoiceExists
		}
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
	bank, err := s.repo.GetAccountByCode(ctx, "1010")
	if err != nil {
		return nil, fmt.Errorf("bank account: %w", err)
	}
	revenue, err := s.repo.GetAccountByCode(ctx, "4010")
	if err != nil {
		return nil, fmt.Errorf("revenue account: %w", err)
	}
	paid, err := s.repo.PayInvoiceAtomic(ctx, userID, invoiceID, bank.ID, revenue.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		if errors.Is(err, ErrInvalidInvoice) {
			return nil, ErrInvalidInvoice
		}
		return nil, fmt.Errorf("pay invoice: %w", err)
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
