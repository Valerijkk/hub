package accounting_test
import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_PostJournal_NotFound(t *testing.T) {
	app, _ := setupAccountingTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/accounting/journal/"+uuid.New().String()+"/post", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandler_CreateInvoice_InvalidAmount(t *testing.T) {
	app, _ := setupAccountingTest(t)

	body, err := json.Marshal(CreateInvoiceRequest{
		CustomerName: "Bad Co",
		Description:  "Free work",
		AmountCents:  0,
	})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/accounting/invoices", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_CreateInvoice_InvalidJSON(t *testing.T) {
	app, _ := setupAccountingTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/accounting/invoices", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_IssueInvoice_NotFound(t *testing.T) {
	app, _ := setupAccountingTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/accounting/invoices/"+uuid.New().String()+"/issue", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_MarkInvoicePaid_DraftInvalidState(t *testing.T) {
	app, _ := setupAccountingTest(t)

	body, err := json.Marshal(CreateInvoiceRequest{
		CustomerName: "Draft Co",
		Description:  "Consulting",
		AmountCents:  50000,
	})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/accounting/invoices", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var inv InvoiceResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&inv))
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodPost, "/api/accounting/invoices/"+inv.ID.String()+"/pay", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_MarkInvoicePaid_NotFound(t *testing.T) {
	app, _ := setupAccountingTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/accounting/invoices/"+uuid.New().String()+"/pay", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandler_MarkInvoicePaid_AlreadyPaid(t *testing.T) {
	app, _ := setupAccountingTest(t)

	body, err := json.Marshal(CreateInvoiceRequest{
		CustomerName: "Paid Co",
		Description:  "Subscription",
		AmountCents:  100000,
	})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/accounting/invoices", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var inv InvoiceResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&inv))
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodPost, "/api/accounting/invoices/"+inv.ID.String()+"/issue", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	payURL := "/api/accounting/invoices/" + inv.ID.String() + "/pay"
	req = httptest.NewRequest(http.MethodPost, payURL, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodPost, payURL, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
