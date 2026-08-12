package accounting

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/journalentry"
	"github.com/valerii/hub/backend/ent/journalline"
	"github.com/valerii/hub/backend/ent/ledgeraccount"
	"github.com/valerii/hub/backend/ent/user"
)

// Repository — персистентность бухгалтерской книги (V4-03).
type Repository interface {
	EnsureDefaultChart(ctx context.Context) error
	ListAccounts(ctx context.Context) ([]*ent.LedgerAccount, error)
	CreateAccount(ctx context.Context, req CreateAccountRequest) (*ent.LedgerAccount, error)
	AccountExists(ctx context.Context, id uuid.UUID) (bool, error)

	ListJournalEntries(ctx context.Context, limit int) ([]*ent.JournalEntry, error)
	GetJournalEntry(ctx context.Context, id uuid.UUID) (*ent.JournalEntry, error)
	CreateJournalEntry(ctx context.Context, userID uuid.UUID, entryDate time.Time, req CreateJournalRequest) (*ent.JournalEntry, error)
	PostJournalEntry(ctx context.Context, id uuid.UUID) (*ent.JournalEntry, error)
	// SoftDeleteJournalEntry — soft-delete родителя; lines скрываются join-filter (A2W8-3-004).
	SoftDeleteJournalEntry(ctx context.Context, id uuid.UUID) error

	TrialBalance(ctx context.Context) ([]TrialBalanceRow, error)
	IsSystemAdmin(ctx context.Context, userID uuid.UUID) (bool, error)

	GetAccountByCode(ctx context.Context, code string) (*ent.LedgerAccount, error)
	IncomeStatement(ctx context.Context) (*IncomeStatementResponse, error)

	ListInvoices(ctx context.Context, limit int) ([]*ent.Invoice, error)
	CreateInvoice(ctx context.Context, userID uuid.UUID, req CreateInvoiceRequest, number string) (*ent.Invoice, error)
	GetInvoice(ctx context.Context, id uuid.UUID) (*ent.Invoice, error)
	IssueInvoice(ctx context.Context, id uuid.UUID) (*ent.Invoice, error)
	MarkInvoicePaid(ctx context.Context, id uuid.UUID, entryID uuid.UUID) (*ent.Invoice, error)
	// PayInvoiceAtomic — claim issued invoice + journal create/post + mark paid в одной tx.
	PayInvoiceAtomic(ctx context.Context, userID, invoiceID, bankAccountID, revenueAccountID uuid.UUID) (*ent.Invoice, error)
	InvoiceExistsForDeal(ctx context.Context, dealID uuid.UUID) (bool, error)
}

// EntRepository — реализация Repository.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository создаёт Ent-репозиторий.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

var defaultAccounts = []CreateAccountRequest{
	{Code: "1010", Name: "Расчётный счёт", AccountType: AccountAsset},
	{Code: "2010", Name: "Кредиторская задолженность", AccountType: AccountLiability},
	{Code: "3010", Name: "Уставный капитал", AccountType: AccountEquity},
	{Code: "4010", Name: "Выручка", AccountType: AccountRevenue},
	{Code: "5010", Name: "Операционные расходы", AccountType: AccountExpense},
}

func (r *EntRepository) EnsureDefaultChart(ctx context.Context) error {
	n, err := r.client.LedgerAccount.Query().
		Where(ledgeraccount.DeletedAtIsNil()).
		Count(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	for _, a := range defaultAccounts {
		if _, err := r.client.LedgerAccount.Create().
			SetCode(a.Code).
			SetName(a.Name).
			SetAccountType(ledgeraccount.AccountType(a.AccountType)).
			Save(ctx); err != nil {
			return fmt.Errorf("seed account %s: %w", a.Code, err)
		}
	}
	return nil
}

func (r *EntRepository) ListAccounts(ctx context.Context) ([]*ent.LedgerAccount, error) {
	return r.client.LedgerAccount.Query().
		Where(ledgeraccount.DeletedAtIsNil()).
		Order(ent.Asc(ledgeraccount.FieldCode)).
		All(ctx)
}

func (r *EntRepository) CreateAccount(ctx context.Context, req CreateAccountRequest) (*ent.LedgerAccount, error) {
	return r.client.LedgerAccount.Create().
		SetCode(req.Code).
		SetName(req.Name).
		SetAccountType(ledgeraccount.AccountType(req.AccountType)).
		Save(ctx)
}

func (r *EntRepository) AccountExists(ctx context.Context, id uuid.UUID) (bool, error) {
	return r.client.LedgerAccount.Query().
		Where(ledgeraccount.ID(id), ledgeraccount.DeletedAtIsNil()).
		Exist(ctx)
}

func (r *EntRepository) ListJournalEntries(ctx context.Context, limit int) ([]*ent.JournalEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	return r.client.JournalEntry.Query().
		Where(journalentry.DeletedAtIsNil()).
		WithCreatedBy().
		WithLines(func(q *ent.JournalLineQuery) {
			q.WithAccount()
		}).
		Order(ent.Desc(journalentry.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
}

func (r *EntRepository) GetJournalEntry(ctx context.Context, id uuid.UUID) (*ent.JournalEntry, error) {
	return r.client.JournalEntry.Query().
		Where(journalentry.ID(id), journalentry.DeletedAtIsNil()).
		WithCreatedBy().
		WithLines(func(q *ent.JournalLineQuery) {
			q.WithAccount()
		}).
		Only(ctx)
}

func (r *EntRepository) CreateJournalEntry(ctx context.Context, userID uuid.UUID, entryDate time.Time, req CreateJournalRequest) (*ent.JournalEntry, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	entry, err := tx.JournalEntry.Create().
		SetEntryDate(entryDate).
		SetDescription(req.Description).
		SetReference(req.Reference).
		SetCreatedByID(userID).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	for _, line := range req.Lines {
		if _, err := tx.JournalLine.Create().
			SetEntryID(entry.ID).
			SetAccountID(line.AccountID).
			SetDebitCents(line.DebitCents).
			SetCreditCents(line.CreditCents).
			SetMemo(line.Memo).
			Save(ctx); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetJournalEntry(ctx, entry.ID)
}

func (r *EntRepository) PostJournalEntry(ctx context.Context, id uuid.UUID) (*ent.JournalEntry, error) {
	now := time.Now().UTC()
	_, err := r.client.JournalEntry.UpdateOneID(id).
		Where(journalentry.DeletedAtIsNil(), journalentry.StatusEQ(journalentry.StatusDraft)).
		SetStatus(journalentry.StatusPosted).
		SetPostedAt(now).
		AddVersion(1).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return r.GetJournalEntry(ctx, id)
}

// SoftDeleteJournalEntry — A2W8-3-004 join-filter: lines наследуют deleted_at родителя.
func (r *EntRepository) SoftDeleteJournalEntry(ctx context.Context, id uuid.UUID) error {
	affected, err := r.client.JournalEntry.Update().
		Where(journalentry.IDEQ(id), journalentry.DeletedAtIsNil()).
		SetDeletedAt(time.Now().UTC()).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return &ent.NotFoundError{}
	}
	return nil
}

func (r *EntRepository) TrialBalance(ctx context.Context) ([]TrialBalanceRow, error) {
	accounts, err := r.ListAccounts(ctx)
	if err != nil {
		return nil, err
	}
	lines, err := r.client.JournalLine.Query().
		Where(
			journalline.HasEntryWith(
				journalentry.StatusEQ(journalentry.StatusPosted),
				journalentry.DeletedAtIsNil(),
			),
		).
		WithAccount().
		All(ctx)
	if err != nil {
		return nil, err
	}
	totals := map[uuid.UUID]TrialBalanceRow{}
	for _, a := range accounts {
		totals[a.ID] = TrialBalanceRow{
			AccountID:   a.ID,
			Code:        a.Code,
			Name:        a.Name,
			AccountType: AccountType(a.AccountType),
		}
	}
	for _, l := range lines {
		if l.Edges.Account == nil {
			continue
		}
		row := totals[l.Edges.Account.ID]
		row.DebitCents += l.DebitCents
		row.CreditCents += l.CreditCents
		totals[l.Edges.Account.ID] = row
	}
	out := make([]TrialBalanceRow, 0, len(accounts))
	for _, a := range accounts {
		out = append(out, totals[a.ID])
	}
	return out, nil
}

func (r *EntRepository) IsSystemAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	u, err := r.client.User.Get(ctx, userID)
	if err != nil {
		return false, err
	}
	return u.SystemRole == user.SystemRoleAdmin || u.SystemRole == user.SystemRoleFounder, nil
}
