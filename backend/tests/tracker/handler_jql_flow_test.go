package tracker_test

// Тесты HTTP-обработчика трекера задач: интеграционные и unit-сценарии testframework.
import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_SearchIssuesJQL_Success проверяет HTTP-слой: HTTP-обработчик search issues jql успех.
func TestHandler_SearchIssuesJQL_Success(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	base := "/api/projects/" + projectID.String()

	createBody := `{"title":"JQL target","issue_type":"task","priority":"medium"}`
	req := httptest.NewRequest(http.MethodPost, base+"/issues/", bytes.NewReader([]byte(createBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	jqlBody := `{"query":"priority = medium ORDER BY created DESC"}`
	req = httptest.NewRequest(http.MethodPost, base+"/issues/jql", bytes.NewReader([]byte(jqlBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, float64(1), result["total"])
}

// TestHandler_SearchIssuesJQL_InvalidJQL проверяет отклонение при HTTP-обработчик search issues jql некорректный jql.
func TestHandler_SearchIssuesJQL_InvalidJQL(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	base := "/api/projects/" + projectID.String()

	req := httptest.NewRequest(http.MethodPost, base+"/issues/jql", bytes.NewReader([]byte(`{"query":""}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
