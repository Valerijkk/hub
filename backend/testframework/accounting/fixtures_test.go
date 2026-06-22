package accounting_test

import (
	"time"

	"github.com/valerii/hub/backend/internal/accounting"
)

type (
	Service               = accounting.Service
	EntRepository         = accounting.EntRepository
	Repository            = accounting.Repository
	AccountResponse       = accounting.AccountResponse
	JournalEntryResponse  = accounting.JournalEntryResponse
	JournalLineInput      = accounting.JournalLineInput
	JournalLineResponse   = accounting.JournalLineResponse
	TrialBalanceRow       = accounting.TrialBalanceRow
	InvoiceResponse       = accounting.InvoiceResponse
	IncomeStatementResponse = accounting.IncomeStatementResponse
	CreateAccountRequest  = accounting.CreateAccountRequest
	CreateJournalRequest  = accounting.CreateJournalRequest
	CreateInvoiceRequest  = accounting.CreateInvoiceRequest
	AccountType           = accounting.AccountType
	EntryStatus           = accounting.EntryStatus
	InvoiceStatus         = accounting.InvoiceStatus
)

const (
	AccountAsset     = accounting.AccountAsset
	AccountLiability = accounting.AccountLiability
	AccountEquity    = accounting.AccountEquity
	AccountRevenue   = accounting.AccountRevenue
	AccountExpense   = accounting.AccountExpense

	StatusDraft  = accounting.StatusDraft
	StatusPosted = accounting.StatusPosted

	InvoiceDraft     = accounting.InvoiceDraft
	InvoiceIssued    = accounting.InvoiceIssued
	InvoicePaid      = accounting.InvoicePaid
	InvoiceCancelled = accounting.InvoiceCancelled
)

var (
	NewService       = accounting.NewService
	NewEntRepository = accounting.NewEntRepository

	ErrNotFound      = accounting.ErrNotFound
	ErrForbidden     = accounting.ErrForbidden
	ErrInvalidInput  = accounting.ErrInvalidInput
	ErrUnbalanced    = accounting.ErrUnbalanced
	ErrAlreadyPosted = accounting.ErrAlreadyPosted
	ErrInvoiceExists = accounting.ErrInvoiceExists
	ErrInvalidInvoice = accounting.ErrInvalidInvoice
)

func mapServiceError(err error) error {
	return accounting.ExportMapServiceError(err)
}

func parseEntryDate(raw string) (time.Time, error) {
	return accounting.ExportParseEntryDate(raw)
}
