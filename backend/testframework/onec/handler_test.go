package onec_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	hrmod "github.com/valerii/hub/backend/internal/hr"
	"github.com/valerii/hub/backend/internal/integrations/onec"
)

func TestService_ExportApprovedLeave(t *testing.T) {
	var received []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		received = body
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	store := &memStore{data: map[string]string{}}
	svc := onec.NewService(store, stubUsers{})
	onec.ExportServiceSetClient(svc, &http.Client{Timeout: 5 * time.Second})
	cfg := onec.Config{
		Enabled:         true,
		WebhookURL:      srv.URL,
		EmployeeIDField: onec.EmployeeIDEmail,
	}
	raw, _ := onec.ExportEncodeConfig(cfg)
	store.data[onec.ExportSettingsKey()] = raw

	reviewer := uuid.New()
	req := hrmod.LeaveRequestResponse{
		ID:            uuid.New(),
		LeaveType:     hrmod.LeaveVacation,
		StartDate:     "2026-07-01",
		EndDate:       "2026-07-14",
		RequesterID:   uuid.New(),
		RequesterName: "Dev",
	}
	err := svc.ExportApprovedLeave(context.Background(), req, reviewer)
	require.NoError(t, err)
	require.NotEmpty(t, received)

	var payload onec.LeaveWebhookPayload
	require.NoError(t, json.Unmarshal(received, &payload))
	assert.Equal(t, "leave.approved", payload.Event)
	assert.Equal(t, "dev@hub.local", payload.Employee.ExternalID)
}

func TestService_PostBlocksLoopbackSSRF(t *testing.T) {
	svc := onec.NewService(&memStore{data: map[string]string{}}, stubUsers{})
	cfg := onec.Config{WebhookURL: "http://127.0.0.1:9999/hook"}
	_, err := onec.ExportServicePost(svc, context.Background(), cfg, []byte(`{}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "blocked private address")
}

func TestAdminHandler_GetSettings(t *testing.T) {
	store := &memStore{data: map[string]string{}}
	svc := onec.NewService(store, stubUsers{})
	h := onec.NewAdminHandler(svc, func(c fiber.Ctx, _ uuid.UUID) (bool, error) { return true, nil })
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error {
		c.Locals("userID", uuid.New().String())
		return c.Next()
	})

	req := httptest.NewRequest(http.MethodGet, "/api/admin/integrations/onec/settings", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
