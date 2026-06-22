package tms_test

// Тесты HTTP-обработчика TMS (волна 124): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Фаза 124 — HTTP success path for rollback..

func TestHandler_Wave124_RollbackTestCase_Success(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := fmt.Sprintf("/api/projects/%s/tms/cases", projectID.String())

	createBody := `{"title":"RB HTTP","case_type":"classic","priority":"medium"}`
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(createBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created tms.TestCaseResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	updateBody := `{"title":"RB HTTP v2"}`
	req = httptest.NewRequest(http.MethodPut, base+"/"+created.CaseKey, bytes.NewReader([]byte(updateBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodGet, base+"/"+created.CaseKey+"/versions", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var versions []tms.TestCaseVersionResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&versions))
	require.NotEmpty(t, versions)

	req = httptest.NewRequest(http.MethodPost, base+"/"+created.CaseKey+"/rollback/"+versions[0].ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
