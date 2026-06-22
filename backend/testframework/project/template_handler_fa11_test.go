package project_test

// FA-11 — handler error paths for template import coverage.
import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateHandler_FA11_ImportTemplate_MissingFile(t *testing.T) {
	env := newTemplateRoutesEnv(t)

	req := httptest.NewRequest(http.MethodPost, "/api/project-templates/import", nil)
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestTemplateHandler_FA11_ExportTemplate_InvalidID(t *testing.T) {
	env := newTemplateRoutesEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/project-templates/not-a-uuid/export", nil)
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
