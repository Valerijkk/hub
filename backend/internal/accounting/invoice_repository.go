package accounting

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/crmdeal"
	"github.com/valerii/hub/backend/ent/invoice"
	"github.com/valerii/hub/backend/ent/ledgeraccount"
)

func (r *EntRepository) GetAccountByCode(ctx context.Context, code string) (*ent.LedgerAccount, error) {
	return r.client.LedgerAccount.Query().
		Where(ledgeraccount.CodeEQ(code), ledgeraccount.DeletedAtIsNil()).
		Only(ctx)
}

func (r *EntRepository) IncomeStatement(ctx context.Context) (*IncomeStatementResponse, error) {
	rows, err := r.TrialBalance(ctx)
	if err != nil {
		return nil, err
	}
	var revenue, expense int64
	for _, row := range rows {
		net := row.DebitCents - row.CreditCents
		switch row.AccountType {
		case AccountRevenue:
			revenue += row.CreditCents - row.DebitCents
		case AccountExpense:
			expense += net
		}
	}
	return &IncomeStatementResponse{
		RevenueCents: revenue,
		ExpenseCents: expense,
		NetCents:     revenue - expense,
	}, nil
}

func (r *EntRepository) nextInvoiceNumber(ctx context.Context) (string, error) {
	year := time.Now().UTC().Year()
	prefix := fmt.Sprintf("INV-%d-", year)
	n, err := r.client.Invoice.Query().
		Where(
			invoice.DeletedAtIsNil(),
			invoice.NumberHasPrefix(prefix),
		).
		Count(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%03d", prefix, n+1), nil
}

func (r *EntRepository) ListInvoices(ctx context.Context, limit int) ([]*ent.Invoice, error) {
	if limit <= 0 {
		limit = 50
	}
	return r.client.Invoice.Query().
		Where(invoice.DeletedAtIsNil()).
		WithCrmDeal().
		WithJournalEntry().
		Order(ent.Desc(invoice.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
}

func (r *EntRepository) CreateInvoice(ctx context.Context, userID uuid.UUID, req CreateInvoiceRequest, number string) (*ent.Invoice, error) {
	if number == "" {
		var err error
		number, err = r.nextInvoiceNumber(ctx)
		if err != nil {
			return nil, err
		}
	}
	b := r.client.Invoice.Create().
		SetNumber(number).
		SetCustomerName(req.CustomerName).
		SetAmountCents(req.AmountCents).
		SetCreatedByID(userID)
	if req.Description != "" {
		b.SetDescription(req.Description)
	}
	if req.Currency != "" {
		b.SetCurrency(req.Currency)
	}
	if req.DueDate != "" {
		if due, err := time.Parse("2006-01-02", req.DueDate); err == nil {
			b.SetDueDate(due.UTC())
		}
	}
	if req.CrmDealID != nil {
		b.SetCrmDealID(*req.CrmDealID)
	}
	return b.Save(ctx)
}

func (r *EntRepository) GetInvoice(ctx context.Context, id uuid.UUID) (*ent.Invoice, error) {
	return r.client.Invoice.Query().
		Where(invoice.ID(id), invoice.DeletedAtIsNil()).
		WithCrmDeal().
		WithJournalEntry().
		Only(ctx)
}

func (r *EntRepository) IssueInvoice(ctx context.Context, id uuid.UUID) (*ent.Invoice, error) {
	now := time.Now().UTC()
	inv, err := r.client.Invoice.UpdateOneID(id).
		Where(
			invoice.DeletedAtIsNil(),
			invoice.StatusEQ(invoice.StatusDraft),
		).
		SetStatus(invoice.StatusIssued).
		SetIssuedAt(now).
		AddVersion(1).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return r.GetInvoice(ctx, inv.ID)
}

func (r *EntRepository) MarkInvoicePaid(ctx context.Context, id uuid.UUID, entryID uuid.UUID) (*ent.Invoice, error) {
	now := time.Now().UTC()
	inv, err := r.client.Invoice.UpdateOneID(id).
		Where(
			invoice.DeletedAtIsNil(),
			invoice.StatusEQ(invoice.StatusIssued),
		).
		SetStatus(invoice.StatusPaid).
		SetPaidAt(now).
		SetJournalEntryID(entryID).
		AddVersion(1).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return r.GetInvoice(ctx, inv.ID)
}

func (r *EntRepository) InvoiceExistsForDeal(ctx context.Context, dealID uuid.UUID) (bool, error) {
	return r.client.Invoice.Query().
		Where(
			invoice.HasCrmDealWith(crmdeal.ID(dealID)),
			invoice.DeletedAtIsNil(),
			invoice.StatusNEQ(invoice.StatusCancelled),
		).
		Exist(ctx)
}
