package accounting_test
import (
	"github.com/valerii/hub/backend/internal/accounting"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
)

func setupAccountingTest(t *testing.T) (*fiber.App, uuid.UUID) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:acct?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("cfo@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	repo := NewEntRepository(client)
	svc := NewService(repo)
	h := accounting.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", admin.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)
	return app, admin.ID
}

func TestHandler_CreateAndPostJournal(t *testing.T) {
	app, _ := setupAccountingTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/accounting/accounts", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var accounts struct {
		Accounts []AccountResponse `json:"accounts"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&accounts))
	require.NotEmpty(t, accounts.Accounts)

	expense := accounts.Accounts[0]
	bank := accounts.Accounts[0]
	for _, a := range accounts.Accounts {
		if a.Code == "5010" {
			expense = a
		}
		if a.Code == "1010" {
			bank = a
		}
	}

	body, err := json.Marshal(CreateJournalRequest{
		EntryDate:   "2026-06-25",
		Description: "Office supplies",
		Lines: []JournalLineInput{
			{AccountID: expense.ID, DebitCents: 150000, CreditCents: 0},
			{AccountID: bank.ID, DebitCents: 0, CreditCents: 150000},
		},
	})
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPost, "/api/accounting/journal", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var entry JournalEntryResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&entry))
	assert.Equal(t, StatusDraft, entry.Status)

	req = httptest.NewRequest(http.MethodPost, "/api/accounting/journal/"+entry.ID.String()+"/post", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/accounting/trial-balance", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_ForbiddenForMember(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:acct2?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })

	member, err := client.User.Create().
		SetEmail("dev@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(context.Background())
	require.NoError(t, err)

	h := accounting.NewHandler(NewService(NewEntRepository(client)))
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error {
		c.Locals("userID", member.ID.String())
		return c.Next()
	})

	req := httptest.NewRequest(http.MethodGet, "/api/accounting/accounts", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestHandler_InvoiceFlow(t *testing.T) {
	app, _ := setupAccountingTest(t)

	body, err := json.Marshal(CreateInvoiceRequest{
		CustomerName: "Acme LLC",
		Description:  "Consulting",
		AmountCents:  250000,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/accounting/invoices", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var inv InvoiceResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&inv))
	assert.Equal(t, InvoiceDraft, inv.Status)

	req = httptest.NewRequest(http.MethodPost, "/api/accounting/invoices/"+inv.ID.String()+"/issue", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, "/api/accounting/invoices/"+inv.ID.String()+"/pay", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/accounting/income-statement", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var pl IncomeStatementResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pl))
	assert.Equal(t, int64(250000), pl.RevenueCents)
}
