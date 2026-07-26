package accounting

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/crmdeal"
	"github.com/valerii/hub/backend/ent/invoice"
	"github.com/valerii/hub/backend/ent/journalentry"
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

// PayInvoiceAtomic атомарно оплачивает issued-счёт: journal+post+paid в одной транзакции (A6-002).
// При ошибке MarkPaid откатывается и проводка — без orphan posted journals.
func (r *EntRepository) PayInvoiceAtomic(ctx context.Context, userID, invoiceID, bankAccountID, revenueAccountID uuid.UUID) (*ent.Invoice, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	inv, err := tx.Invoice.Query().
		Where(
			invoice.ID(invoiceID),
			invoice.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if inv.Status != invoice.StatusIssued {
		return nil, ErrInvalidInvoice
	}

	desc := fmt.Sprintf("Оплата счёта %s", inv.Number)
	entryDate := time.Now().UTC()
	if inv.IssuedAt != nil {
		entryDate = inv.IssuedAt.UTC()
	}

	entry, err := tx.JournalEntry.Create().
		SetEntryDate(entryDate).
		SetDescription(desc).
		SetReference(inv.Number).
		SetCreatedByID(userID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create payment journal: %w", err)
	}
	if _, err := tx.JournalLine.Create().
		SetEntryID(entry.ID).
		SetAccountID(bankAccountID).
		SetDebitCents(inv.AmountCents).
		SetCreditCents(0).
		SetMemo(inv.CustomerName).
		Save(ctx); err != nil {
		return nil, fmt.Errorf("create bank line: %w", err)
	}
	if _, err := tx.JournalLine.Create().
		SetEntryID(entry.ID).
		SetAccountID(revenueAccountID).
		SetDebitCents(0).
		SetCreditCents(inv.AmountCents).
		Save(ctx); err != nil {
		return nil, fmt.Errorf("create revenue line: %w", err)
	}

	now := time.Now().UTC()
	_, err = tx.JournalEntry.UpdateOneID(entry.ID).
		Where(journalentry.StatusEQ(journalentry.StatusDraft)).
		SetStatus(journalentry.StatusPosted).
		SetPostedAt(now).
		AddVersion(1).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("post payment journal: %w", err)
	}

	paid, err := tx.Invoice.UpdateOneID(invoiceID).
		Where(
			invoice.DeletedAtIsNil(),
			invoice.StatusEQ(invoice.StatusIssued),
		).
		SetStatus(invoice.StatusPaid).
		SetPaidAt(now).
		SetJournalEntryID(entry.ID).
		AddVersion(1).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("mark paid: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetInvoice(ctx, paid.ID)
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
