package accounting_test
import (
	"github.com/valerii/hub/backend/internal/accounting"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_CreateAccount_ListJournal_ListInvoices(t *testing.T) {
	app, _ := setupAccountingTest(t)

	body, err := json.Marshal(CreateAccountRequest{
		Code: "6010", Name: "Marketing", AccountType: AccountExpense,
	})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/accounting/accounts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var acct AccountResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&acct))
	assert.Equal(t, "6010", acct.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/accounting/journal", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/accounting/invoices", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_PostJournal_AlreadyPostedConflict(t *testing.T) {
	app, _ := setupAccountingTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/accounting/accounts", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	var accounts struct {
		Accounts []AccountResponse `json:"accounts"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&accounts))
	resp.Body.Close()

	var expense, bank AccountResponse
	for _, a := range accounts.Accounts {
		if a.Code == "5010" {
			expense = a
		}
		if a.Code == "1010" {
			bank = a
		}
	}

	body, err := json.Marshal(CreateJournalRequest{
		EntryDate: "2026-06-26", Description: "Duplicate post test",
		Lines: []JournalLineInput{
			{AccountID: expense.ID, DebitCents: 1000, CreditCents: 0},
			{AccountID: bank.ID, DebitCents: 0, CreditCents: 1000},
		},
	})
	require.NoError(t, err)
	req = httptest.NewRequest(http.MethodPost, "/api/accounting/journal", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	var entry JournalEntryResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&entry))
	resp.Body.Close()

	postURL := "/api/accounting/journal/" + entry.ID.String() + "/post"
	req = httptest.NewRequest(http.MethodPost, postURL, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodPost, postURL, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestHandler_TrialBalance_IncomeStatement(t *testing.T) {
	app, _ := setupAccountingTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/accounting/trial-balance", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/accounting/income-statement?from=2026-01-01&to=2026-12-31", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_CreateJournal_InvalidInput(t *testing.T) {
	app, _ := setupAccountingTest(t)
	req := httptest.NewRequest(http.MethodPost, "/api/accounting/journal",
		bytes.NewReader([]byte(`{"entry_date":"","lines":[]}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_CreateAccount_InvalidInput(t *testing.T) {
	app, _ := setupAccountingTest(t)
	req := httptest.NewRequest(http.MethodPost, "/api/accounting/accounts",
		bytes.NewReader([]byte(`{"code":"","name":""}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestMapServiceError_AllBranches(t *testing.T) {
	assert.Equal(t, fiberStatus(mapServiceError(accounting.ErrNotFound)), 404)
	assert.Equal(t, fiberStatus(mapServiceError(ErrForbidden)), 403)
	assert.Equal(t, fiberStatus(mapServiceError(ErrInvalidInput)), 400)
	assert.Equal(t, fiberStatus(mapServiceError(ErrUnbalanced)), 400)
	assert.Equal(t, fiberStatus(mapServiceError(ErrAlreadyPosted)), 409)
	assert.Equal(t, fiberStatus(mapServiceError(ErrInvoiceExists)), 409)
	assert.Equal(t, fiberStatus(mapServiceError(ErrInvalidInvoice)), 400)
}

func fiberStatus(err error) int {
	if fe, ok := err.(*fiber.Error); ok {
		return fe.Code
	}
	return 0
}
