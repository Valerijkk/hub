//go:build hubtest

package calls_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/calls"
)

// TASK-064 / W3-A5-013 — nil RBAC на list/download recordings → 500 fail-closed.

func TestHandler_ListProjectRecordings_NilRBAC_FailClosed_TASK064(t *testing.T) {
	svc, projectID, client := newRecordingListFixture(t)
	ctx := t.Context()
	owner, err := client.User.Query().First(ctx)
	require.NoError(t, err)

	h := calls.ExportNewHandlerWithRecSvc(svc) // rbacSvc == nil
	app := newHandlerValidationApp(h)

	req := httptest.NewRequest(fiber.MethodGet, "/projects/"+projectID.String()+"/recordings", nil)
	req.Header.Set("X-Test-User-ID", owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestHandler_GetRecordingDownloadURL_NilRBAC_FailClosed_TASK064(t *testing.T) {
	svc, _, client := newRecordingListFixture(t)
	ctx := t.Context()
	owner, err := client.User.Query().First(ctx)
	require.NoError(t, err)
	rec, err := client.Recording.Query().First(ctx)
	require.NoError(t, err)

	h := calls.ExportNewHandlerWithRecSvc(svc)
	app := newHandlerValidationApp(h)

	req := httptest.NewRequest(fiber.MethodGet, "/recordings/"+rec.ID.String()+"/download", nil)
	req.Header.Set("X-Test-User-ID", owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}
