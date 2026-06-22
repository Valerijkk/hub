package accounting

import (
	"time"

	"github.com/google/uuid"
)

// AccountType — классификация плана счетов.
type AccountType string

const (
	AccountAsset     AccountType = "asset"
	AccountLiability AccountType = "liability"
	AccountEquity      AccountType = "equity"
	AccountRevenue     AccountType = "revenue"
	AccountExpense     AccountType = "expense"
)

// EntryStatus — жизненный цикл журнальной записи.
type EntryStatus string

const (
	StatusDraft  EntryStatus = "draft"
	StatusPosted EntryStatus = "posted"
)

// AccountResponse — счёт главной книги.
type AccountResponse struct {
	ID          uuid.UUID   `json:"id"`
	Code        string      `json:"code"`
	Name        string      `json:"name"`
	AccountType AccountType `json:"account_type"`
	IsActive    bool        `json:"is_active"`
	Version     int         `json:"version"`
}

// JournalLineInput — line в create request.
type JournalLineInput struct {
	AccountID   uuid.UUID `json:"account_id"`
	DebitCents  int64     `json:"debit_cents"`
	CreditCents int64     `json:"credit_cents"`
	Memo        string    `json:"memo,omitempty"`
}

// JournalLineResponse — сохранённая строка.
type JournalLineResponse struct {
	ID          uuid.UUID `json:"id"`
	AccountID   uuid.UUID `json:"account_id"`
	AccountCode string    `json:"account_code,omitempty"`
	AccountName string    `json:"account_name,omitempty"`
	DebitCents  int64     `json:"debit_cents"`
	CreditCents int64     `json:"credit_cents"`
	Memo        string    `json:"memo,omitempty"`
}

// JournalEntryResponse — header journal с lines.
type JournalEntryResponse struct {
	ID          uuid.UUID             `json:"id"`
	EntryDate   string                `json:"entry_date"`
	Description string                `json:"description"`
	Reference   string                `json:"reference,omitempty"`
	Status      EntryStatus           `json:"status"`
	Version     int                   `json:"version"`
	PostedAt    *time.Time            `json:"posted_at,omitempty"`
	CreatedBy   uuid.UUID             `json:"created_by"`
	Lines       []JournalLineResponse `json:"lines"`
	CreatedAt   time.Time             `json:"created_at"`
}

// TrialBalanceRow — posted balance по account.
type TrialBalanceRow struct {
	AccountID   uuid.UUID   `json:"account_id"`
	Code        string      `json:"code"`
	Name        string      `json:"name"`
	AccountType AccountType `json:"account_type"`
	DebitCents  int64       `json:"debit_cents"`
	CreditCents int64       `json:"credit_cents"`
}

type CreateAccountRequest struct {
	Code        string      `json:"code"`
	Name        string      `json:"name"`
	AccountType AccountType `json:"account_type"`
}

type CreateJournalRequest struct {
	EntryDate   string             `json:"entry_date"`
	Description string             `json:"description"`
	Reference   string             `json:"reference,omitempty"`
	Lines       []JournalLineInput `json:"lines"`
}

// InvoiceStatus — жизненный цикл счёта (V4-03.5).
type InvoiceStatus string

const (
	InvoiceDraft     InvoiceStatus = "draft"
	InvoiceIssued    InvoiceStatus = "issued"
	InvoicePaid      InvoiceStatus = "paid"
	InvoiceCancelled InvoiceStatus = "cancelled"
)

// InvoiceResponse — счёт клиента.
type InvoiceResponse struct {
	ID           uuid.UUID     `json:"id"`
	Number       string        `json:"number"`
	CustomerName string        `json:"customer_name"`
	Description  string        `json:"description,omitempty"`
	AmountCents  int64         `json:"amount_cents"`
	Currency     string        `json:"currency"`
	Status       InvoiceStatus `json:"status"`
	DueDate      *string       `json:"due_date,omitempty"`
	IssuedAt     *time.Time    `json:"issued_at,omitempty"`
	PaidAt       *time.Time    `json:"paid_at,omitempty"`
	CrmDealID    *uuid.UUID    `json:"crm_deal_id,omitempty"`
	JournalID    *uuid.UUID    `json:"journal_entry_id,omitempty"`
	Version      int           `json:"version"`
	CreatedAt    time.Time     `json:"created_at"`
}

type CreateInvoiceRequest struct {
	CustomerName string     `json:"customer_name"`
	Description  string     `json:"description,omitempty"`
	AmountCents  int64      `json:"amount_cents"`
	Currency     string     `json:"currency,omitempty"`
	DueDate      string     `json:"due_date,omitempty"`
	CrmDealID    *uuid.UUID `json:"crm_deal_id,omitempty"`
}

// IncomeStatementResponse — P&L из posted entries (V4-03.5).
type IncomeStatementResponse struct {
	RevenueCents int64 `json:"revenue_cents"`
	ExpenseCents int64 `json:"expense_cents"`
	NetCents     int64 `json:"net_cents"`
}
